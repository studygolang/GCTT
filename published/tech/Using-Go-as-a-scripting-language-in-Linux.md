已发布：https://studygolang.com/articles/12461

# 在 Linux 中使用 Go 作为脚本语言

在 `Cloudflare` 的人们都非常喜欢 Go 语言。我们在许多[内部软件项目](https://blog.cloudflare.com/what-weve-been-doing-with-go/)以及更大的[管道系统](https://blog.cloudflare.com/meet-gatebot-a-bot-that-allows-us-to-sleep/)中使用它。但是，我们能否进入下一个层次并将其用作我们最喜欢的操作系统 Linux 的脚本语言呢？

![image here](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-script/gopher-tux-1.png)

## 为什么考虑将 Go 作为脚本语言

简短点的回答：为什么不呢？Go 相对容易学习，不冗余并且有一个强大的生态库，这些库可以重复使用避免我们从头开始编写所有代码。它可能带来的一些其他潜在优势：

* 为你的 Go 项目提供一个基于 Go 的构建系统：`go build` 命令主要适用于小型自包含项目。更复杂的项目通常采用构建系统或脚本集。为什么不用 Go 编写这些脚本呢？

* 易于使用的非特权包管理：如果你想在脚本中使用第三方库，你可以简单的使用 `go get` 命令来获取。而且由于拉取的代码将安装在你的 `GOPATH` 中，使用一些第三方库并不需要系统管理员的权限（与其他一些脚本语言不同）。这在大型企业环境中尤其有用。

* 在早期项目阶段进行快速的代码原型设计：当您进行第一次代码迭代时，通常需要进行大量的编辑，甚至进行编译，而且您必须在 "编辑->构建->检查" 循环中浪费大量的按键。相反，使用 Go，您可以跳过 `build` 部分，并立即执行源文件。

* 强类型的脚本语言：如果你在脚本中的某个地方有个小的输入错误，大多数的脚本语言都会执行到有错误的地方然后停止。这可能会让你的系统处于不一致的状态（因为有些语句的执行会改变数据的状态，从而污染了执行脚本之前的状态）。使用强类型语言时，许多拼写错误可以在编译时被捕获，因此有 bug 的脚本将不会首先运行。

## Go 脚本的当前状态

咋一看 Go 脚本貌似很容易实现 Unix 脚本的 shebang(#! ...) 支持。[shebang 行](https://en.wikipedia.org/wiki/Shebang_(Unix))是脚本的第一行，以 `#!` 开头，并指定脚本解释器用于执行脚本（例如，`#!/bin/bash` 或 `#!/usr/bin/env python`），所以无论使用何种编程语言，系统都确切知道如何执行脚本。Go 已经使用 `go run` 命令支持 `.go` 文件的类似于解释器的调用，所以只需要添加适当的 shebang 行（`#!/usr/bin/env go run`）到任何的 `.go` 文件中，设置好文件的可执行状态，然后就可以愉快的玩耍了。

但是，直接使用 go run 还是有问题的。[这篇牛 b 的文章](https://gist.github.com/posener/73ffd326d88483df6b1cb66e8ed1e0bd)详细描述了围绕 `go run` 的所有问题和潜在解决方法，但其要点是：

* `go run` 不能正确地将脚本错误代码返回给操作系统，这对脚本很重要，因为错误代码是多个脚本之间相互交互和操作系统环境最常见的方式之一。

* 你不能在有效的 `.go` 文件中创建一个 shebang 行，因为 Go 语言不知道如何处理以 `#` 开头的行。而其他语言不存在这个问题，是由于 `#` 大多数情况下是一种注释的方式，所以最后解释器会忽略掉 shebang 行，但是 Go 注释是以 `//` 开头的并且在调用时运行会产生如下错误：

```
package main:
helloscript.go:1:1: illegal character U+0023 '#'
```
[这篇文章](https://gist.github.com/posener/73ffd326d88483df6b1cb66e8ed1e0bd)描述了上述问题的几种解决方法，包括使用一个自定义的包装程序 [gorun](https://github.com/erning/gorun) 作为解释器，但是都没有提供一个理想的解决方案。你可以：

* 必须使用非标准的 shebang 行，它以 `//` 开头。这在技术上甚至不是 shebang 行，而是 bash shell 如何处理可执行文本文件的方式，所以这个解决方案是 bash 特有的。另外，由于 `go run` 的具体行为，这一行相当复杂并且不够明显（请参阅原始文章的示例）。

* 必须在 shebang 行中使用 gorun 自定义包装程序，这很好，但是，最终得到的 `.go` 文件由于非法的 `#` 字符而不能与标准 `go build` 命令编译。

## Linux 如何执行文件

OK，看起来 shebang 的方法并没有为我们提供全面的解决方案。是否还有其他方式是我们可以使用的？让我们仔细看看 Linux 内核如何执行二进制文件。 当你尝试执行一个二进制/脚本（或任何有可执行位设置的文件）时，你的 shell 最后只会使用 Linux `execve` 系统调用，将它传递给二进制文件系统路径，命令行参数和 当前定义的环境变量。 然后内核负责正确解析文件并用文件中的代码创建一个新进程。 我们中的大多数人都知道 Linux （和许多其他类 Unix 操作系统）为其可执行文件使用 ELF 二进制格式。

然而，Linux 内核开发的核心原则之一是避免任何子系统的 “vendor/format lock-in”，这是内核的一部分。因此，Linux 实现了一个“可插拔”系统，它允许内核支持任何二进制格式 - 所有你需要做的就是编写一个正确的模块，它可以解析你选择的格式。如果仔细研究内核源代码，你会发现 Linux 支持更多的二进制格式。例如，最近的`4.14` Linux 内核，我们可以看到它至少支持7种二进制格式（用于各种二进制格式的树内模块通常在其名称中具有 `binfmt_` 前缀）。值得注意的是 [binfmt_script](https://git.kernel.org/pub/scm/linux/kernel/git/stable/linux-stable.git/tree/fs/binfmt_script.c?h=linux-4.14.y) 模块，它负责解析上面提到的 shebang 行并在目标系统上执行脚本（并不是每个人都知道 shebang 支持实际上是在内核本身而不是在 shell 或其他守护进程/进程中实现的）。

## 从用户空间扩展受支持的二进制格式

但既然我们认为 shebang 不是 Go 脚本的最佳选择，似乎我们需要别的东西。令人惊讶的是，Linux 内核已经有了一个“其他类型的”二进制支持模块，它有一个贴切的名称 `binfmt_misc`。该模块允许管理员通过定义良好的 `procfs` 接口直接从用户空间动态添加对各种可执行格式的支持，并且有详细记录。

让我们按照[文档](https://www.kernel.org/doc/html/v4.14/admin-guide/binfmt-misc.html)并尝试为 `.go` 文件设置二进制格式描述。首先，该指南告诉您将特殊的 `binfmt_misc` 文件系统安装到 `/proc/sys/fs/binfmt_misc`。如果您使用的是基于 systemd 的相对较新的 Linux 发行版，则很可能已经为您安装了文件系统，因为默认情况下 system 会为此安装特殊的 mount 和 automount 单元。 要仔细检查，只需运行：

```shell
$ mount | grep binfmt_misc
systemd-1 on /proc/sys/fs/binfmt_misc type autofs (rw,relatime,fd=27,pgrp=1,timeout=0,minproto=5,maxproto=5,direct)
```

另一种方法是检查 `/proc/sys/fs/binfmt_misc` 中是否有文件：正确安装 `binfmt_misc` 文件系统将至少创建两个名称为 `register` 和 `status` 的特殊文件。

接下来，因为我们希望我们的 .go 脚本能够正确地将退出代码传递给操作系统，所以我们需要将定制的 gorun 包装器作为我们的“解释器”：

```shell
$ go get github.com/erning/gorun
$ sudo mv ~/go/bin/gorun /usr/local/bin/
```

从技术角度上讲，我们不需要将 gorun 移动到 `/usr/local/bin` 或任何其他系统路径，而无论如何 `binfmt_misc` 都需要解释器的完整路径，但系统可以以任意权限运行此可执行文件，因此从安全视角来看限制文件访问权限是一个好主意。

在这一点上，让我们来建一个简单的 go 脚本 `helloscript.go` 并验证我们可以成功“解释”它。脚本如下：

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	s := "world"

	if len(os.Args) > 1 {
		s = os.Args[1]
	}

	fmt.Printf("Hello, %v!", s)
	fmt.Println("")

	if s == "fail" {
		os.Exit(30)
	}
}
```

检查参数传递和错误处理是否按预期工作：

```shell
$ gorun helloscript.go
Hello, world!
$ echo $?
0
$ gorun helloscript.go gopher
Hello, gopher!
$ echo $?
0
$ gorun helloscript.go fail
Hello, fail!
$ echo $?
30
```

现在我们需要告诉 `binfmt_misc` 模块如何使用 `gorun` 执行 `.go` 文件。按照文档中的描述我们需要配置如下字符串： `:golang:E::go::/usr/local/bin/gorun:OC`，意思是告诉系统：当遇到以 `.go` 为扩展名的可执行文件，请使用 `/usr/local/bin/gorun` 解释器执行该文件。字符串末尾的 `OC` 标志确保脚本将根据脚本本身设置的所有者信息和权限位执行，而不是在解释器二进制文件上设置的那些位。这使 Go 脚本的执行行为与 Linux 中其他可执行文件和脚本的行为相同。

让我们注册我们新的 Go 脚本二进制格式：

```shell
$ echo ':golang:E::go::/usr/local/bin/gorun:OC' | sudo tee /proc/sys/fs/binfmt_misc/register
:golang:E::go::/usr/local/bin/gorun:OC
```

如果系统成功注册了，则应在 `/proc/sys/fs/binfmt_misc` 目录下显示新的 golang 文件。 最后，我们可以在本地执行我们的 .go 文件：

```shell
$ chmod u+x helloscript.go
$ ./helloscript.go
Hello, world!
$ ./helloscript.go gopher
Hello, gopher!
$ ./helloscript.go fail
Hello, fail!
$ echo $?
30
```

就这样了！现在我们可以根据自己的喜好编辑 helloscript.go，并在下次执行文件时看到更改将立即可见。此外，和此前的 shebang 方式不同，我们可以随时使用 `go build` 将文件编译成真正的可执行文件。

---

via：[Using Go as a scripting language in Linux](https://blog.cloudflare.com/using-go-as-a-scripting-language-in-linux/)

作者：[Ignat Korchagin](https://blog.cloudflare.com/author/ignat/)
译者：[shniu](https://github.com/shniu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
