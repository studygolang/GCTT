首发于：https://studygolang.com/articles/35202

# Go Module 教程第 2 部分：项目、依赖和 gopls

## 引言

模块是集成到 Go 系统中，为依赖管理提供支持。这意味着模块几乎可以触及任何与源代码相关的内容，包括编辑器支持。为了向编辑器提供模块支持（以及其他原因） ，Go 团队构建了一个名为 [gopls](https://github.com/golang/tools/blob/master/gopls/doc/user.md) 的服务，它实现了语言服务器协议（[LSP](https://microsoft.github.io/language-server-protocol/))。LSP 最初是由微软为 VSCode 开发的，现在已经成为一个开放标准。该协议的思想是为编辑器提供对语言特性的支持，比如自动完成、定义和查找所有引用。

当你使用模块和 VSCode 时，在编辑器中点击保存将不再直接运行 go build 命令。现在发生的情况是，一个请求被发送给 gopls，gopls 运行适当的 Go 命令和相关的 API 来提供编辑器反馈和支持。Gopls 还可以向编辑器发送信息而不需要请求。有时，由于 LSP 的特性或运行 Go 命令的固有延迟，编辑器似乎滞后或与代码更改不同步。团队正在努力达到一个 1.0 版本的 gopls 来处理这些边缘情况，这样你就可以有最平滑的编辑器体验。

在这篇文章中，我将介绍在项目中添加和删除依赖项的基本工作流程。本文使用了 VSCode 编辑器、 gopls 的0.2.0 版本和 Go 的 1.13.3 版本。（GCTT 注：目前 gopls 的版本做了大量的优化，个人感觉已经很好用了）

## 模块缓存

为了帮助加快构建速度并快速更新项目中的依赖关系更改，Go 维护一个缓存，其中包含它在本地计算机上下载的所有模块。该缓存可以在 `$GOPATH/pkg` 中找到。如果没有 GOPATH 设置，默认的 GOPATH 是 `$HOME/go`。

注意: 有一个提案建议提供一个环境变量，允许用户控制模块缓存的位置。如果没有更改，`$GOPATH/pkg` 将是默认值。（GCTT 注：现在已经有了这个环境变量：GOMODCACHE）

清单 1：

```bash
$HOME/code/go/pkg
$ ls -l
total 0
drwxr-xr-x  11 bill  staff  352 Oct 16 15:53 mod
drwxr-xr-x   3 bill  staff   96 Oct  3 16:49 sumdb
```

清单 1 显示了我当前 $GOPATH/pkg 文件夹的样子。你可以看到有两个文件夹，mod 和 sumdb。如果你查看 mod 文件夹内部，你可以了解更多关于模块缓存布局的信息。

清单 2：

```bash
$HOME/code/go/pkg
$ ls -l mod/
total 0
drwxr-xr-x   5 bill  staff   160 Oct  7 10:37 cache
drwxr-xr-x   3 bill  staff    96 Oct  3 16:55 contrib.go.opencensus.io
drwxr-xr-x  40 bill  staff  1280 Oct 16 15:53 github.com
dr-x------  26 bill  staff   832 Oct  3 16:50 go.opencensus.io@v0.22.1
drwxr-xr-x   3 bill  staff    96 Oct  3 16:56 golang.org
drwxr-xr-x   4 bill  staff   128 Oct  7 10:37 google.golang.org
drwxr-xr-x   7 bill  staff   224 Oct 16 15:53 gopkg.in
drwxr-xr-x   7 bill  staff   224 Oct 16 15:53 k8s.io
drwxr-xr-x   5 bill  staff   160 Oct 16 15:53 sigs.k8s.io
```

清单 2 显示了当前模块缓存的顶级结构。你可以看到如何将与模块名称关联的 URL 的第一部分用作模块缓存中的顶级文件夹。如果我导航到 github.com/ardanlabs，可以向你展示 2 个实际的模块。

清单 3：

```bash
$HOME/code/go/pkg
$ ls -l mod/github.com/ardanlabs/
total 0
dr-x------  13 bill  staff  416 Oct  3 16:49 conf@v1.1.0
dr-x------  18 bill  staff  576 Oct 12 10:08 service@v0.0.0-20191008203700-49ed4b4f1088
```

清单 3 显示了我正在使用的来自 ArdanLabs 的两个模块及其版本。第一个是 conf 模块，另一个模块与我用来讲解 kubernetes 和服务的服务项目相关联。

Gopls 服务器还维护一个保存在内存中的模块缓存。在启动 VSCode 并处于模块模式时，将启动一个 gopls 服务器来支持该编辑器会话。内部 gopls 模块缓存与当前在磁盘上的内容同步。Gopls 使用这个内部模块缓存来处理编辑器请求。

在这篇文章中，我将在开始之前清空模块缓存，这样我就有了一个干净的工作环境。我还将在启动 VSCode 编辑器之前设置我的项目。这将允许我向你展示如何处理你需要的模块尚未下载到本地模块缓存或更新到 gopls 内部模块缓存的情况。

注意: 在任何正常的工作流中，您都不应该清除模块缓存。

清单 4：

```bash
$ go clean -modcache
```

清单 4 显示了如何清除磁盘上的本地模块缓存。清理命令通常用于清理本地的 GOPATH 工作目录和 GOPATH/bin 文件夹。现在使用新的 -mocache 标志，可以使用该命令清理模块缓存。

注意: 这个命令不会清除任何正在运行的 gopls 实例的内部缓存。

## 新项目

我将在 GOPATH 之外开始一个新项目，在编写代码的过程中，我将介绍添加和删除依赖项的基本工作流程。

清单 5：

```bash
$ cd $HOME
$ mkdir service
$ cd service
$ mkdir cmd
$ mkdir cmd/sales-api
$ touch cmd/sales-api/main.go
```

清单 5 显示了一些命令，这些命令用于设置工作目录文件、创建初始项目结构并添加 main.go 文件。

使用模块时的第一步是初始化项目源树的根。这是通过使用 go mod init 命令完成的。

清单 6：

```bash
$ go mod init github.com/ardanlabs/service
```

清单 6 显示了对 go mod init 的调用，将模块的名称作为参数传递。正如在[第一篇文章](https://studygolang.com/articles/24580)中所讨论的，模块的名称允许在模块内部解析内部导入。按照仓库代码的 URL 来命名模块是惯用法。在这篇文章中，我假设这个模块与 Github 中 Ardan Labs 下的 [service repo](https://github.com/ardanlabs/service) 相关联。

一旦调用 go mod init 完成，就会在当前工作目录中创建一个 go.mod 文件。这个文件将表示项目的根。

清单 7：

```bash
01 module github.com/ardanlabs/service
02
03 go 1.13
```

清单 7 显示了这个项目的初始模块文件的内容。有了这些，就可以进行项目编码了。

清单 8：

```bash
$ code .
```

清单 8 显示了启动 VSCode 实例的命令。这将反过来启动 gopls 服务器的实例，以支持这个编辑器实例。

![图1](https://www.ardanlabs.com/images/goinggo/110_figure1.png)

图 1 显示了在运行所有命令之后，我的 VSCode 编辑器中的项目是什么样子的。为了确保你使用的设置与我相同，我将列出我当前的 VSCode 设置。

清单 9：

```json
{
    // Important Settings
    "go.lintTool": "golint",
    "go.goroot": "/usr/local/go",
    "go.gopath": "/Users/bill/code/go",

    "go.useLanguageServer": true,
    "[go]": {
        "editor.snippetSuggestions": "none",
        "editor.formatOnSave": true,
        "editor.codeActionsOnSave": {
            "source.organizeImports": true
        }
    },
    "gopls": {
        "usePlaceholders": true,    // add parameter placeholders when completing a function
        "completeUnimported": true, // autocomplete unimported packages
        "deepCompletion": true,     // enable deep completion
    },
    "go.languageServerFlags": [
        "-rpc.trace", // for more detailed debug logging
    ],
}
```

清单 9 显示了我当前的 VSCode 设置。如果你一直跟着做，没有看到相同的行为，检查你的这些设置。如果你想查看当前推荐的 VSCode 设置，请点击[这里](https://github.com/golang/tools/blob/master/gopls/doc/vscode.md)。

## 应用编码

我将从这个应用程序的初始代码开始。

清单 10：<https://play.studygolang.com/p/AU1xFIVOLu9>

```go
01 package main
02
03 func main() {
04     if err := run(); err != nil {
05         log.Println("error :", err)
06         os.Exit(1)
07     }
08 }
09
10 func run() error {
11     return nil
12 }
```

清单 10 显示了我添加到 main.go 的前 12 行代码。它为应用程序设置一个单一的退出点，并记录启动或关闭时的任何错误。一旦这 12 行代码被保存到文件中，编辑器将自动地(感谢 gopls)包含从标准库中所需的导入。

清单 11：<https://play.studygolang.com/p/x3hBA6PuW3R>

```go
03 import (
04     "log"
05     "os"
06 )
```

清单 11 显示了由于编辑器与 gopls 集成而对第 03 至 06 行的源代码所做的更改。

接下来，我将添加对配置的支持。

清单 12：<https://play.studygolang.com/p/4hFXLJj4yT_Z>

```go
17 func run() error {
18     var cfg struct {
19         Web struct {
20             APIHost         string        `conf:"default:0.0.0.0:3000"`
21             DebugHost       string        `conf:"default:0.0.0.0:4000"`
22             ReadTimeout     time.Duration `conf:"default:5s"`
23             WriteTimeout    time.Duration `conf:"default:5s"`
24             ShutdownTimeout time.Duration `conf:"default:5s"`
25         }
26     }
27
28     if err := conf.Parse(os.Args[1:], "SALES", &cfg); err != nil {
29         return fmt.Errorf("parsing config : %w", err)
30     }
```

清单 12 显示了添加到第 18 行到第 30 行的 run 函数中以支持配置的代码。当这段代码被添加到源文件中并点击保存时，编辑器会将 fmt 和 time 包正确地包含到导入集中。不幸的是，由于 gopls 目前在其内部模块缓存中没有关于 conf 包的任何信息，所以 gopls 不能指示编辑器为 conf 添加一个导入或向编辑器提供包信息。

![图2](https://www.ardanlabs.com/images/goinggo/110_figure2.png)

图 2 显示了编辑器如何清楚地表明它不能解析与 conf 包相关的任何信息。

## 添加一个依赖项

为了解析导入，需要检索包含 conf 包的模块。这样做的一种方法是将导入添加到源代码文件的顶部，并让编辑器和 gopls 完成这项工作。

清单 13：

```go
01 package main
02
03 import (
04     "fmt"
05     "log"
06     "os"
07     "time"
08
09     "github.com/ardanlabs/conf"
10 )
```

在清单 13 中，我在第 09 行添加了 conf 包的导入。一旦我点击保存，编辑器就会找到 gopls，然后 gopls 会找到、下载并使用 Go 命令和相关的 API 提取这个包的模块。这些调用还更新 Go 模块文件以反映这一更改。

清单 14：

```bash
~/code/go/pkg/mod/github.com/ardanlabs
$ ls -l
total 0
drwxr-xr-x   3 bill  staff    96B Nov  8 16:02 .
drwxr-xr-x   3 bill  staff    96B Nov  8 16:02 ..
dr-x------  13 bill  staff   416B Nov  8 16:02 conf@v1.2.0
```

清单 14 显示了 Go 命令如何完成它的工作，以及如何使用版本 1.2.0 下载 conf 模块。我们需要解析导入的代码现在在我的本地模块缓存中。

![图3](https://www.ardanlabs.com/images/goinggo/110_figure3.png)

图 3 显示了编辑器仍然不能解析有关包的信息。为什么编辑器无法解析此信息？不幸的是，gopls 内部模块缓存与本地模块缓存不同步。Gopls 服务器并不知道 Go 命令刚刚做出的更改。由于 gopls 使用它的内部缓存，所以 gopls 不能向编辑器提供它所需要的信息。

注意: 这个缺点目前正在处理中，将在即将发布的版本中修正。你可以在这里追踪这个问题。(<https://github.com/golang/go/issues/31999>)。（GCTT 译注：目前版本该问题已经解决了）

使 gopls 内部模块缓存与本地模块缓存同步的一个快速方法是重新加载 VS Code 编辑器。这将重新启动 gopls 服务器并重置其内部模块缓存。在 VSCode 中，有一个名为 reload window 的特殊命令可以做到这一点。（新版本不需要此步骤了）

```bash
Ctrl + Shift + P and run  > Reload Window
```

![图4](https://www.ardanlabs.com/images/goinggo/110_figure4.png)

图 4 显示了在使用 Ctrl + Shift + P 快捷键 reload 窗口之后在 VS Code 中出现的对话框。

运行此快速命令后，将解析与导入相关的任何消息。

## 可传递依赖关系

从 Go 工具的角度来看，构建这个应用程序所需的所有代码现在都在本地模块缓存中。但是，conf 包的测试依赖于 googlego-cmp 包。

清单 15：

```bash
module github.com/ardanlabs/conf

go 1.13

require github.com/google/go-cmp v0.3.1
```

清单 15 显示了 conf 模块的 1.2.0 版本的模块文件。您可以看到 conf 依赖于 go-cmp 的 0.3.1 版本。此模块未列入服务的模块文件中，因为这样做是冗余的。Go 工具可以按照模块文件的路径来获取构建或测试代码所需的所有模块的完整图像。

此时，还没有找到这个传递模块，也没有将其下载并提取到本地模块缓存中。因为在构建代码时不需要这个模块，所以 Go 构建工具还没有发现需要下载它。如果我在命令行上运行 go mod tidy，那么 Go 工具将花费时间将 go-cmp 模块放入本地缓存中。

清单 16：

```bash
$ go mod tidy
go: downloading github.com/google/go-cmp v0.3.1
go: extracting github.com/google/go-cmp v0.3.1
```

清单 16 显示了如何找到、下载和提取 go-cmp 模块。这个调用 go mod tidy 不会改变项目的模块文件，因为这不是一个直接的依赖项。它将更新 go.sum 文件，以便有模块 hash 的记录，从而维护持久的、可重复的构建。我将在以后的文章中谈论校验和数据库。

清单 17：

```bash
github.com/ardanlabs/conf v1.2.0 h1:2IntiqlEhRk+sYUbc8QAAZdZlpBWIzNoqILQvV6Jofo=
github.com/ardanlabs/conf v1.2.0/go.mod h1:ILsMo9dMqYzCxDjDXTiwMI0IgxOJd0MOiucbQY2wlJw=
github.com/google/go-cmp v0.3.1 h1:Xye71clBPdm5HgqGwUkwhbynsUJZhDbS20FvLhQ2izg=
github.com/google/go-cmp v0.3.1/go.mod h1:8QqcDgzrUqlUb/G2PQTWiueGozuR1884gddMywk6iLU=
```

清单 17 显示了运行 go mod tidy 后校验和文件的样子。每个与项目相关联的模块有两条记录。

## 下载模块

如果你还没有准备好在代码库中使用某个特定的模块，但是希望将该模块下载到本地模块缓存中，可以选择手动将该模块添加到项目 go.mod 文件中，然后在编辑器外运行 go mod tidy。

清单 18：

```bash
01 module github.com/ardanlabs/service
02
03 go 1.13
04
05 require (
06     github.com/ardanlabs/conf v1.2.0
07     github.com/pkg/errors latest
08 )
```

在清单 18 中，你可以看到我如何为最新版本的 errors 模块在模块文件中手动添加第 07 行。手动添加所需模块的重要部分是使用最新的标记。一旦我对这个更改运行 go mod tidy，它会告诉 Go 找到 errors 模块的最新版本并将其下载到缓存中。

清单 19：

```bash
$HOME/service
$ go mod tidy
go: finding github.com/pkg/errors v0.8.1
```

清单 19 显示了如何找到、下载和提取 errors 模块的 0.8.1 版本。一旦命令运行完毕，模块将从模块文件中删除，因为项目不使用该模块。但是，该模块列在校验和文件中。

清单 20：

```bash
github.com/ardanlabs/conf v1.2.0 h1:2IntiqlEhRk+sYUbc8QAAZdZlpBWIzNoqILQvV6Jofo=
github.com/ardanlabs/conf v1.2.0/go.mod h1:ILsMo9dMqYzCxDjDXTiwMI0IgxOJd0MOiucbQY2wlJw=
github.com/google/go-cmp v0.3.1 h1:Xye71clBPdm5HgqGwUkwhbynsUJZhDbS20FvLhQ2izg=
github.com/google/go-cmp v0.3.1/go.mod h1:8QqcDgzrUqlUb/G2PQTWiueGozuR1884gddMywk6iLU=
github.com/pkg/errors v0.8.1/go.mod h1:bwawxfHBFNV+L2hUp1rHADufV3IMtnDRdf1r5NINEl0=
```

清单 20 显示了如何在校验和文件中列出 errors 模块模块文件的散列。记住校验和文件不是项目使用的所有依赖项的规范记录，这一点很重要。它可以包含更多的模块，这是绝对好的。

我喜欢这种通过使用 go get 来下载新模块的方法，因为如果不小心的话，go get 也可以尝试升级项目的依赖关系图(直接和间接)。重要的是要知道什么时候版本升级只是下载你想要的新模块。在以后的文章中，我将讨论使用 go get 来更新现有的模块依赖关系。

## 移除依赖

如果我决定不再使用 conf 包会发生什么？我可以删除使用包的任何代码。

清单 21：<https://play.studygolang.com/p/x3hBA6PuW3R>

```go
01 package main
02
03 import (
04     "log"
05     "os"
06 )
07
08 func main() {
09     if err := run(); err != nil {
10         log.Println("error :", err)
11         os.Exit(1)
12     }
13 }
14
15 func run() error {
16     return nil
17 }
```

清单 21 显示了从 main 函数中删除引用 conf 包的代码。一旦我点击保存，编辑器就会从导入集中删除 conf 的导入。但是，模块文件没有更新以反映更改。

清单 22：

```bash
01 module github.com/ardanlabs/service
02
03 go 1.13
04
05 require github.com/ardanlabs/conf v1.1.0
```

清单 22 显示 conf 包仍然被认为是必需的。为了解决这个问题，我需要离开编辑器，再次运行 go mod tidy。

清单 23：

```bash
$HOME/service
$ go mod tidy
```

清单 23 再次显示了 go mod 的运行情况。这次没有输出。一旦这个命令完成，模块文件再次精确。

清单 24：

```bash
$HOME/services/go.mod

01 module github.com/ardanlabs/service
02
03 go 1.13
```

清单 24 显示了从模块文件中删除了 conf 模块。这一次，go mod tidy 命令清除了校验和文件，它将是空的。在你对 VCS 进行任何修改之前，确保你的模块文件是正确的，并且与你使用的依赖关系一致，这一点很重要。

## 总结

在不久的将来，我分享的一些解决方案，比如重新加载窗口，将不再需要。团队意识到了这一点以及当今存在的其他缺陷，他们正在积极地修复这些缺陷。他们非常感谢任何和所有的反馈，所以如果你发现一个问题，请报告它。没有问题是太大或太小。作为一个社区，让我们与 Go 团队一起快速解决这些遗留问题。

现在正在进行的一个核心特性是 gopls 能够监视文件系统并自己查看项目更改。这将有助于 gopls 保持其内部模块缓存与磁盘上的本地模块缓存同步。一旦这样做了，重新装载窗口的需求就会消失。计划也在制定中，以提供视觉线索，工作是在背景中进行的。（GCTT 注：目前已经解决）

总的来说，我对当前的工具集和刷新窗口的工作方式感到满意。我希望你考虑开始使用模块，如果你还没有。模块已经可以使用了，越多的项目开始使用它，Go 生态系统对每个人来说就越好。

---

via: https://www.ardanlabs.com/blog/2019/12/modules-02-projects-dependencies-gopls.html

作者：[William Kennedy](https://www.ardanlabs.com/)
译者：[polaris1119](https://github.com/polaris1119)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
