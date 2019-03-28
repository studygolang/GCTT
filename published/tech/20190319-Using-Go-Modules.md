首发于：https://studygolang.com/articles/19334

# Go Modules 的使用方法

## 简介

Go 1.11 和 1.12 引入了[对 modules（模块）的初步支持](https://golang.org/doc/go1.11#modules)，这是一个能让依赖项的版本信息更加明确和易于管理的依赖管理系统。本文旨在为你使用模块提供基本的操作指导。后续会有一篇文章来说明如何发布一个模块以供别人使用。

一个模块是一系列 [Go 代码包](https://golang.org/ref/spec#Packages) 的集合，它们保存在同一个文件树中。文件树的根目录中包含了一个 `go.mod` 文件。`go.mod` 文件定义了一个模块的 module path，这就是模块根目录的导入路径。`go.mod` 文件还定义了模块的 *dependency requirements（依赖项要求）*，即为了编译本模块，需要用到哪些其它的模块。每一项依赖项要求都包含了依赖项的 module path，还要指定它的[语义版本号](http://semver.org/)。

对于 Go 1.11 来说，当你工作目录不在 `$GOPATH/src` 里面，并且工作目录或者工作目录的任意级父目录中含有 `go.mod` 文件，go 命令行工具会启用模块机制。（但为了兼容以前的机制，当你工作目录是在 `$GOPATH/src` 里面的时候，go 命令行工具不会启用模块机制，即使工作目录里面有 `go.mod` 文件也一样。详情请见[go 命令工具文档](https://golang.org/cmd/go/#hdr-Preliminary_module_support) ）。从 Go 1.13 开始，模块机制在所有情况下都将会默认启用。

本文将覆盖 Go 开发中涉及到模块的一系列常用操作：

- 创建一个新的模块
- 添加一个依赖项
- 升级依赖项
- 添加一个拥有更高主版本号的依赖项
- 更新当前依赖项到一个新的主版本号
- 移除无用的依赖项

## 创建一个新的模块

让我们来创建一个新的模块吧。

在 `$GOPATH/src` **之外**创建一个新的、空白的目录，进入这个目录，并且新建一个源代码文件，`hello.go`：

```go
package hello
func Hello() string {
  return "Hello, world."
}
```

让我们再写点测试，在 `hello_test.go` 文件中：

```go
package hello

import "testing"

func TestHello(t *testing.T) {
    want := "Hello, world."
    if got := Hello(); got != want {
        t.Errorf("Hello() = %q, want %q", got, want)
    }
}
```

到目前为止，目录里面包含了一个代码包，但是它还不是一个模块，因为这里面没有 `go.mod` 文件。如果我们现在的工作目录是 `/home/gopher/hello` 并且我们运行 `go test`，我们会看到如下的输出：

```bash
$ go test
PASS
ok      _/home/gopher/hello    0.020s
$
```

最后一行总结了代码包整体的测试结果，因为我们工作目录在 `$GOPATH` 之外，且又不属于任何模块，所以 Go 命令行工具并不知道当前目录的导入路径是什么，所以只能用目录的路径作为包的名字：`_/home/gopher/hello`。

让我们把当前的目录设置成模块的根目录吧，为此我们要用到 `go mod init` 命令然后再尝试运行 `go test`：

```bash
$ go mod init example.com/hello
go: creating new go.mod: module example.com/hello
$ go test
PASS
ok      example.com/hello    0.020s
$
```

恭喜你，你编写并测试了你的第一个模块！

`go mod init` 命令创建了一个 `go.mod` 文件：

```bash
$ cat go.mod
module example.com/hello

go 1.12
$
```

 `go.mod` 文件只存在于模块的根目录中。模块子目录的代码包的导入路径等于模块根目录的导入路径（就是前面说的 module path）加上子目录的相对路径。比如，我们如果创建了一个子目录叫 `world`，我们不需要（也不会想要）在子目录里面再运行一次 `go mod init` 了，这个代码包会被认为就是 `example.com/hello` 模块的一部分，而这个代码包的导入路径就是 `example.com/hello/world`。

## 添加依赖项

引进 Go 模块系统的主要动机，就是让用户更轻松地使用其他开发者编写的代码（换句话说就是添加一个依赖项）。

让我们修改一下我们的 `hello.go` ，让它导入 `rsc.io/quote` 模块，并用这个模块的接口来实现 `Hello`：

```go
package hello

import "rsc.io/quote"

func Hello() string {
    return quote.Hello()
}
```

现在让我们再运行一遍测试：

```bash
$ go test
go: finding rsc.io/quote v1.5.2
go: downloading rsc.io/quote v1.5.2
go: extracting rsc.io/quote v1.5.2
go: finding rsc.io/sampler v1.3.0
go: finding golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c
go: downloading rsc.io/sampler v1.3.0
go: extracting rsc.io/sampler v1.3.0
go: downloading golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c
go: extracting golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c
PASS
ok      example.com/hello    0.023s
$
```

go 命令行工具会根据 `go.mod` 里面指定好的依赖的模块版本来下载相应的依赖模块。在你的代码中 import 了一个包，但 `go.mod` 文件里面又没有指定这个包的时候，go 命令行工具会自动寻找包含这个代码包的模块的最新版本，并添加到 `go.mod` 中（这里的 " 最新 " 指的是：它是最近一次被 tag 的稳定版本（即[非预发布版本，non-prerelease](https://semver.org/#spec-item-9)），如果没有，则是最近一次被 tag 的预发布版本，如果没有，则是最新的没有被 tag 过的版本）。在我们的例子是，`go test` 把新导入的 `rsc.io/quote` 包解析为 `rec.io/quote v1.5.2` 模块。它还会下载 `rsc.io/quote` 模块依赖的两个依赖项。即 `rsc.io/sampler` 和 `golang.org/x/text`。但是只有直接依赖会记录在 `go.mod` 文件里面：

```bash
$ cat go.mod
module example.com/hello

go 1.12

require rsc.io/quote v1.5.2
$
```

第二次运行 `go test` 命令的时候 Go 命令工具就不再重复上述的工作了，因为 `go.mod` 已经是更新过了，并且刚才下载下来的模块已经缓存在本地（在 `$GOPATH/pkg/mod`）目录中：

```bash
$ go test
PASS
ok      example.com/hello    0.020s
$
```

要注意的是虽然用 Go 命令行工具添加依赖非常的简单快捷，但是它不是没有代价的。你的项目现在有很多关键指标都对新的依赖项有了依赖，比方说代码的正确性、安全性、合适的版权等等，不一而足。更多相关的资讯，可以查看 Russ Cox 的博客文章，"[Our Software Dependency Problem](https://research.swtch.com/deps)"。

正如我们上面所见，添加一个直接依赖往往会带来其它间接的依赖。`go list -m all` 命令会把当前的模块和它所有的依赖项都列出来：

```bash
$ go list -m all
example.com/hello
golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c
rsc.io/quote v1.5.2
rsc.io/sampler v1.3.0
$
```

在上述 `go list` 命令的输出中，当前的模块，又称为主模块 (main module)，永远都在第一行，接着是主模块的依赖项，以依赖项的 module path 排序。

`golang.org/x/text` 的版本 `v0.0.0-20170915032832-14c0d48ead0c` 是一个典型的[伪版本（pseudo-version）](https://golang.org/cmd/go/#hdr-Pseudo_versions) 的例子，它其实就是 Go 命令工具自定义的一个命名规则，当你想要依赖一个模块的某个 commit 版本的代码，但是这个 commit 没有被 tag 过的时候，可以这样子来指定。

除了 `go.mod` 之外，go 命令行工具还维护了一个 `go.sum` 文件，它包含了指定的模块的版本内容的哈希值作为校验参考：

```bash
$ cat Go.sum
golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c h1:qgOY6WgZO...
golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c/go.mod h1:Nq...
rsc.io/quote v1.5.2 h1:w5fcysjrx7yqtD/aO+QwRjYZOKnaM9Uh2b40tElTs3...
rsc.io/quote v1.5.2/go.mod h1:LzX7hefJvL54yjefDEDHNONDjII0t9xZLPX...
rsc.io/sampler v1.3.0 h1:7uVkIFmeBqHfdjD+gZwtXXI+RODJ2Wc4O7MPEh/Q...
rsc.io/sampler v1.3.0/go.mod h1:T1hPZKmBbMNahiBKFy5HrXp6adAjACjK9...
$
```

go 命令行工具使用 `go.sum` 文件来确保你的项目依赖的模块不会发生变化——无论是恶意的，还是意外的，或者是其它的什么原因。`go.mod` 文件和 `go.sum` 文件都应该保存到你的代码版本控制系统里面去。

## 更新依赖项

有了 Go 模块机制后，模块的版本通过带有语义化版本号（semantic version）的标签来指定。一个语义化版本号包括三个部分：主版本号（major）、次版本号（minor）、修订号（patch）。举个例子：对于版本 `v0.1.2`，主版本号是 0，次版本号是 1，修订号是 2。我们先来过一遍更新某个模块的次版本号的流程。在下一节，我们再考虑主版本号的更新。

从 `go list -m all` 的输出中，我们可以看到我们在使用的 `golang.org/x/text` 模块还是以前没有被打过版本号标签的版本。让我们来把它更新到最新的有打过版本号标签的的版本，并测试是否能正常使用。

```bash
$ go get golang.org/x/text
go: finding golang.org/x/text v0.3.0
go: downloading golang.org/x/text v0.3.0
go: extracting golang.org/x/text v0.3.0
$ go test
PASS
ok      example.com/hello    0.013s
$
```

哇嗷，一切正常！我们再来看看现在 `go list -m all` 的输出和 `go.mod` 文件长什么样子：

```bash
$ go list -m all
example.com/hello
golang.org/x/text v0.3.0
rsc.io/quote v1.5.2
rsc.io/sampler v1.3.0
$ cat go.mod
module example.com/hello

go 1.12

require (
    golang.org/x/text v0.3.0 // indirect
    rsc.io/quote v1.5.2
)
$
```

`golang.org/x/text` 模块已经被升级到最新的版本（`v0.3.0`），`go.mod` 文件里面也把这个模块的版本指定成版本 `v0.3.0`。注释 `indirect` 意味着这个依赖项不是直接被当前模块使用的。而是被模块的其它依赖项使用的。详情请见 `go help modules`。

现在让我们来尝试更新 `rsc.io/sampler` 模块的次版本号。同样操作，先运行 `go get` 命令，然后跑一遍测试：

```bash
$ go get rsc.io/sampler
go: finding rsc.io/sampler v1.99.99
go: downloading rsc.io/sampler v1.99.99
go: extracting rsc.io/sampler v1.99.99
$ go test
--- FAIL: TestHello (0.00s)
    hello_test.go:8: Hello() = "99 bottles of beer on the wall, 99 bottles of beer, ...", want "Hello, world."
FAIL
exit status 1
FAIL    example.com/hello    0.014s
$
```

噢，糟糕，测试报错了，这个测试表明 `rsc.io/sampler` 模块的最新版本跟我们之前的用法不兼容。我们来列举一下这个模块能用的 tag 过的版本：

```bash
$ go list -m -versions rsc.io/sampler
rsc.io/sampler v1.0.0 v1.2.0 v1.2.1 v1.3.0 v1.3.1 v1.99.99
$
```

我们之前用过 `v1.3.0`，而 `v1.99.99` 明显不能用了。也许我们能试一下 `v1.3.1` 版本

```bash
$ go get rsc.io/sampler@v1.3.1
go: finding rsc.io/sampler v1.3.1
go: downloading rsc.io/sampler v1.3.1
go: extracting rsc.io/sampler v1.3.1
$ go test
PASS
ok      example.com/hello    0.022s
$
```

请注意我们给 `go get` 命令的参数后面显式地指定了 `@v1.3.1` ，事实上每个传递给 `go get` 的参数都能在后面显式地指定一个版本号，默认情况下这个版本号是 `@latest`，这代表 Go 命令行工具会尝试下载最新的版本。

## 添加一个拥有更高主版本号的依赖项

我们来为我们的代码包添加一个新的函数，`func Proverb` 会返回一条关于 Go 并发编程的名言，这可以通过调用 `quote.Concurrency` 来实现，而这个函数由 `rsc.io/quote/v3` 模块提供。首先我们要修改我们的 `hello.go` 来添加新的函数：

```go
package hello

import (
    "rsc.io/quote"
    quoteV3 "rsc.io/quote/v3"
)

func Hello() string {
    return quote.Hello()
}

func Proverb() string {
    return quoteV3.Concurrency()
}
```

然后我们在 `hello_test.go` 里面添加一个测试：

```go
func TestProverb(t *testing.T) {
    want := "Concurrency is not parallelism."
    if got := Proverb(); got != want {
        t.Errorf("Proverb() = %q, want %q", got, want)
    }
}
```

然后我们可以来测试我们的代码了：

```bash
$ go test
go: finding rsc.io/quote/v3 v3.1.0
go: downloading rsc.io/quote/v3 v3.1.0
go: extracting rsc.io/quote/v3 v3.1.0
PASS
ok      example.com/hello    0.024s
$
```

请注意我们的模块现在既依赖 `rsc.io/quote` 也依赖 `rsc.io/quote/v3`：

```bash
$ go list -m rsc.io/q...
rsc.io/quote v1.5.2
rsc.io/quote/v3 v3.1.0
$
```

不同主版本号的同一个 Go 模块，使用了不同的 module path ——从 `v2` 开始，module path 的结尾一定要跟上主要版本号。在本例中，`v3` 版本的 `rsc.io/quote` 已经不是 `rsc.io/quote` 了，它的 module path 是 `rsc.io/quote/v3` 。这个规定被称为 [semantic import versioning（语义化的导入版本控制）](https://research.swtch.com/vgo-import)，它给予了不兼容的模块一个不同的名字。反之，`v1.6.0` 版本的 `rsc.io/quote` 应该做到能够向下兼容 `v1.5.2` 版本。所以这两个版本可以共用同一个名字 `rsc.io/quote`。（在上一节中，`v1.99.99` 版本的 `rsc.io/sampler` *应该*要向下兼容 `v1.3.0` 版本的 `rsc.io/sampler`，但是 bug 或者模块使用者对模块的用法不对，这些都有可能会导致错误发生。）

每一次构建项目，go 命令行工具允许每个 module path 最多只有一个，这就意味着每一个主要版本只有一个：最多一个 `rsc.io/quote`，最多一个 `rsc.io/quote/v2`，最多只有一个 `rsc.io/quote/v3` 等等。这给了模块的作者一个明确的信号：对于同一个 module path，可能会存在有多个重复的模块——一个程序有可能同时使用了 `v1.5.2` 的 `rsc.io/quote` 和 `v1.6.0` 的 `rsc.io/quote`。还有，允许同一个模块的不同主版本号存在（因为它们的 module path 不一样），能够给模块使用者部分更新主版本的能力。拿这个例子来说，我们想要使用 `rsc.io/quote/v3 v3.1.0` 模块里面的  `quote.Concurrency`，但是没准备好要整个代码都迁移到 v3 版本的 `rsc.io/quote` 中，这种部分迁移的能力，在大型程序或者代码库里面尤其重要。

## 更新当前依赖项到一个新的主版本号

现在让我们把整个项目的 `rsc.io/quote`  都升级到 `rsc.io/quote/v3` 吧。因为主版本号改变了，所以我们应该做好心理准备，可能会有些 API 已经被移除、重命名或者被修改成了不兼容的方式。通过阅读文档，我们得知 `Hello` 已经变成了 `HelloV3`：

```go
$ go doc rsc.io/quote/v3
package quote // import "rsc.io/quote"

Package quote collects pithy sayings.

func Concurrency() string
func GlassV3() string
func GoV3() string
func HelloV3() string
func OptV3() string
$
```

（上面的输出有个 [已知的 Bug](https://golang.org/issue/30778)：显示的 import 路径后面漏了 `v3`）

我们可以把 `hello.go` 中使用 `quote.Hello()` 的地方改成 `quote.HelloV3()`

```go
package hello

import quoteV3 "rsc.io/quote/v3"

func Hello() string {
    return quoteV3.HelloV3()
}

func Proverb() string {
    return quoteV3.Concurrency()
}
```

然后我们再重新运行一下测试确保一切正常：

```bash
$ go test
PASS
ok      example.com/hello       0.014s
```

## 移除没有用到的依赖

我们代码中已经没有用到 `rsc.io/quote` 的地方了，但是它还是会存在 `go list -m all` 的输出和 `go.mod` 文件中：

```bash
$ go list -m all
example.com/hello
golang.org/x/text v0.3.0
rsc.io/quote v1.5.2
rsc.io/quote/v3 v3.1.0
rsc.io/sampler v1.3.1
$ cat go.mod
module example.com/hello

go 1.12

require (
    golang.org/x/text v0.3.0 // indirect
    rsc.io/quote v1.5.2
    rsc.io/quote/v3 v3.0.0
    rsc.io/sampler v1.3.1 // indirect
)
$
```

为什么会这样？因为我们在构建一个代码包的时候（比如说 `go build` 或者 `go test`），可以轻易的知道哪些依赖缺失，从而将它自动添加进来，但很难知道哪些依赖可以被安全的移除掉。移除一个依赖项需要在检查完模块中所有代码包和这些代码包的所有可能的编译标签的组合。一个普通的 build 命令不会获得这么多的信息，所以它不能保证安全地移除掉没用的依赖项。

可以用 `go mod tidy` 命令来清除这些没用到的依赖项：

```bash
$ go mod tidy
$ go list -m all
example.com/hello
golang.org/x/text v0.3.0
rsc.io/quote/v3 v3.1.0
rsc.io/sampler v1.3.1
$ cat go.mod
module example.com/hello

go 1.12

require (
    golang.org/x/text v0.3.0 // indirect
    rsc.io/quote/v3 v3.1.0
    rsc.io/sampler v1.3.1 // indirect
)

$ go test
PASS
ok      example.com/hello    0.020s
$
```

## 结论

Go 的模块功能将会成为未来 Go 的依赖管理系统。在所有支持模块机制的 Go 版本中（即目前的 Go 1.11、Go 1.12），都能正常使用模块系统拥有的所有功能。

本文介绍了使用 Go 模块过程中的几个工作流程：

- `go mod init` 创建了一个新的模块，初始化 `go.mod` 文件并且生成相应的描述
- `go build, go test` 和其它构建代码包的命令，会在需要的时候在 `go.mod` 文件中添加新的依赖项
- `go list -m all` 列出了当前模块所有的依赖项
- `go get` 修改指定依赖项的版本（或者添加一个新的依赖项）
- `go mod tidy` 移除模块中没有用到的依赖项。

我们鼓励你在你的本地开发中使用模块功能，并为你的项目添加 `go.mod` 和 `go.sum` 文件。请给我们[发送错误报告](https://golang.org/issue/new) 或者[体验报告](https://golang.org/wiki/ExperienceReports)，来为 Go 未来的依赖管理系统的完善提供反馈和帮助。

---

via: https://blog.golang.org/using-go-modules

作者：[Tyler Bui-Palsulich,Eno Compton](https://blog.golang.org/using-go-modules)
译者：[Alex-liutao](https://github.com/Alex-liutao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
