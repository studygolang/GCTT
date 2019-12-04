首发于：https://studygolang.com/articles/23133

# 使用 Go Modules（模块）进行依赖项迁移

## 介绍

本篇文章是 Go Modules 系列文章的第 2 部分。第 1 部分请参阅

[Go Modules 的使用方法（中文翻译版）](https://studygolang.com/articles/19334)

[Using Go Modules（EN）](https://blog.golang.org/using-go-modules)

Go 项目使用各种各样的依赖关系管理策略,像 `dep` 和 `glide` 这样的第三方依赖项[vendor](https://golang.org/cmd/go/#hdr-Vendor_Directories)引用工具很受欢迎，但它们在行为上存在很大差异，而且并不总是能很好地协同工作。有些项目将整个 GOPATH 目录存储在一个 Git 仓库中。其他人只是简单地依赖于 `go get`，并期望在 GOPATH 中安装最新版本的依赖项。

Go 1.11 中引入的 Go Modules（模块）系统提供了一个内置在 Go 命令中的官方依赖管理解决方案。本文主要对项目转换为模块的工具和技术进行讲解叙述。

**请注意:如果您的项目已经标记为 v2.0.0 或更高版本，那么当您添加 `go.mod` 文件时，您需要更新你的模块路径，我们将在以后的一篇关于 v2 及以后的文章中解释如何在不破坏用户的情况下做到这一点。**

## 将第三方依赖项迁移至你的项目中

在开始使用 Go Module 来进行第三方依赖项管理时，你的项目可能处于以下三种状态中的其中一种:

* 一个全新的项目。
* 使用非模块依赖关系管理的 Go 项目。
* 没有任何依赖的的 Go 项目。

第一种情况已经包含在[Using Go Modules（EN）](https://blog.golang.org/using-go-modules)中；剩下的我们将在后两篇文章中讨论。

## 开始使用依赖关系管理

若要转换已使用依赖关系管理工具的项目，请运行以下命令:

```bash
$ git clone https://github.com/my/project
[...]
$ cd project
$ cat Godeps/Godeps.json
{
    "ImportPath": "github.com/my/project",
    "GoVersion": "go1.12",
    "GodepVersion": "v80",
    "Deps": [
        {
            "ImportPath": "rsc.io/binaryregexp",
            "Comment": "v0.2.0-1-g545cabd",
            "Rev": "545cabda89ca36b48b8e681a30d9d769a30b3074"
        },
        {
            "ImportPath": "rsc.io/binaryregexp/syntax",
            "Comment": "v0.2.0-1-g545cabd",
            "Rev": "545cabda89ca36b48b8e681a30d9d769a30b3074"
        }
    ]
}
$ go mod init github.com/my/project
go: creating new go.mod: module github.com/my/project
go: copying requirements from Godeps/Godeps.json
$ cat go.mod
module github.com/my/project

go 1.12

require rsc.io/binaryregexp v0.2.1-0.20190524193500-545cabda89ca
$
```

`go mod init` 创建一个新的 `go.mod` 文件并自动从 `Godeps.json`,`Gopkg.lock` 导入依赖项，或其他一些已经支持的文件格式（[other supported formats](https://go.googlesource.com/go/+/362625209b6cd2bc059b6b0a67712ddebab312d9/src/cmd/go/internal/modconv/modconv.go#9)）。`go mod init` 的参数是项目路径，即项目可能被找到的位置。

在继续执行之前，是时候停下来，运行 `go build ./...` 和 `go test ./...`。接下来的步骤就是修改你的 `go.mod` 文件，因此，如果您喜欢采用迭代方法，这是您的 `go.mod` 文件最接近模块前依赖项规范的地方。

```bash
$ go mod tidy
go: downloading rsc.io/binaryregexp v0.2.1-0.20190524193500-545cabda89ca
go: extracting rsc.io/binaryregexp v0.2.1-0.20190524193500-545cabda89ca
$ cat go.sum
rsc.io/binaryregexp v0.2.1-0.20190524193500-545cabda89ca h1:FKXXXJ6G2bFoVe7hX3kEX6Izxw5ZKRH57DFBJmHCbkU=
rsc.io/binaryregexp v0.2.1-0.20190524193500-545cabda89ca/go.mod h1:qTv7/COck+e2FymRvadv62gMdZztPaShugOCi3I+8D8=
$
```

`go mod tidy` 会查找您项目中所有引入的依赖项目。它将为这个项目包添加一个新的模块需求但不提供任何已知的模块，并删除了对不提供任何导入包的模块的需求。如果模块提供的包仅由尚未迁移到模块的项目导入，则模块需求将用 `//` 间接注释标记。优先运行 `go mod tidy`，然后在进行 `go.mod` 文件的提交，这样会使你的 `go.mod` 文件中的项目依赖保持一个最新的状态，这将是一个非常好的版本控制的实现。

让我们确保代码能够成功编译和测试通过：

```bash
$ go build ./...
$ go test ./...
[...]
$
```

注意，其他依赖项管理工具可能在单个包或整个整个仓库(而不是模块)级别指定依赖项，并且通常不识别依赖项文件 `go.mod` 中指定的需求。因此，您可能无法获得与之前完全相同的每个包的版本，这会提高风险。因此，按照上面的命令对最后依赖项进行检查非常重要。所以我们需要这样做，输入下面的命令：

```bash
$ go list -m all
go: finding rsc.io/binaryregexp v0.2.1-0.20190524193500-545cabda89ca
github.com/my/project
rsc.io/binaryregexp v0.2.1-0.20190524193500-545cabda89ca
$
```

并将结果版本与旧的依赖关系管理文件进行比较，以确保所选版本是适合自己当前项目的。如果你发现一个版本不是你想要的，你可以通过使用 `go mod why -m` 和/或 `go mod graph` 找到原因，并使用 `go get` 升级或降级到正确的版本。(如果您请求的版本比之前选择的版本更旧，`go get` 将根据需要降低其他依赖关系，以保持兼容性。) 例如：

```bash
$ go mod why -m rsc.io/binaryregexp
[...]
$ go mod graph | grep rsc.io/binaryregexp
[...]
$ go get rsc.io/binaryregexp@v0.2.0
$
```

## 当没有依赖关系管理时

对于没有依赖关系管理系统的 Go 项目，首先创建一个 `go.mod` 文件:

```bash
$ git clone https://go.googlesource.com/blog
[...]
$ cd blog
$ go mod init golang.org/x/blog
go: creating new go.mod: module golang.org/x/blog
$ cat go.mod
module golang.org/x/blog

go 1.12
$
```

如果没有以前依赖项管理中的配置文件，`go mod init` 将创建一个 `go.mod` 文件只有模块和 go 指令。在当前案例中，我们将模块路径设置为 `golang.org/x/blog`，因为这是它的[自定义导入路径](https://golang.org/cmd/go/#hdr-Remote_import_paths)。用户可以使用此路径导入包，我们必须注意不要更改它。

模块指令声明模块路径，go 指令声明用于编译模块内代码的 go 语言的预期版本。

接下来，运行 `go mod tidy` 添加模块的依赖项:

```bash
$ go mod tidy
go: finding golang.org/x/website latest
go: finding gopkg.in/tomb.v2 latest
go: finding golang.org/x/net latest
go: finding golang.org/x/tools latest
go: downloading github.com/gorilla/context v1.1.1
go: downloading golang.org/x/tools v0.0.0-20190813214729-9dba7caff850
go: downloading golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7
go: extracting github.com/gorilla/context v1.1.1
go: extracting golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7
go: downloading gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
go: extracting gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
go: extracting golang.org/x/tools v0.0.0-20190813214729-9dba7caff850
go: downloading golang.org/x/website v0.0.0-20190809153340-86a7442ada7c
go: extracting golang.org/x/website v0.0.0-20190809153340-86a7442ada7c
$ cat go.mod
module golang.org/x/blog

go 1.12

require (
    github.com/gorilla/context v1.1.1
    golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7
    golang.org/x/text v0.3.2
    golang.org/x/tools v0.0.0-20190813214729-9dba7caff850
    golang.org/x/website v0.0.0-20190809153340-86a7442ada7c
    gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
)
$ cat go.sum
cloud.google.com/go v0.26.0/go.mod h1:aQUYkXzVsufM+DwF1aE+0xfcU+56JwCaLick0ClmMTw=
cloud.google.com/go v0.34.0/go.mod h1:aQUYkXzVsufM+DwF1aE+0xfcU+56JwCaLick0ClmMTw=
git.apache.org/thrift.git v0.0.0-20180902110319-2566ecd5d999/go.mod h1:fPE2ZNJGynbRyZ4dJvy6G277gSllfV2HJqblrnkyeyg=
git.apache.org/thrift.git v0.0.0-20181218151757-9b75e4fe745a/go.mod h1:fPE2ZNJGynbRyZ4dJvy6G277gSllfV2HJqblrnkyeyg=
github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973/go.mod h1:Dwedo/Wpr24TaqPxmxbtue+5NUziq4I4S80YR8gNf3Q=
[...]
$
```

`go mod tidy` 为模块中的包以及导入的包传递添加到模块中，并为特定版本的每个库构建了一个带有校验和的 `go.sum` 文件。让我们通过代码构建和测试试一试:

```bash
$ go build ./...
$ go test ./...
ok      golang.org/x/blog    0.335s
?       golang.org/x/blog/content/appengine    [no test files]
ok      golang.org/x/blog/content/cover    0.040s
?       golang.org/x/blog/content/h2push/server    [no test files]
?       golang.org/x/blog/content/survey2016    [no test files]
?       golang.org/x/blog/content/survey2017    [no test files]
?       golang.org/x/blog/support/racy    [no test files]
$
```

注意，当 `go mod tidy` 添加一个必须包（requirement）时，它会添加对应模块的最新版本。如果您的 GOPATH 包含一个旧版本的依赖项，随后发布了一个破坏性的更改，您可能会在 `go mod tidy`、`go build` 或 `go test` 中看到错误。如果出现这种情况，尝试使用 `go get` 降级到较老的版本(例如，`go get github.com/broken/module@v1.1.0`)，或者花点时间修改一下你可爱的代码使模块与每个依赖项的最新版本兼容。

## 模块模式下的测试

在迁移到 Go 模块之后，有些测试可能需要调整。

如果测试需要在包目录中写入文件，那么当包目录位于模块缓存(只读)中时，测试可能会失败。特别是，这可能导致 go test all 失败。测试应该将需要写入的文件复制到临时目录。

如果测试依赖于相对路径(../package-in-another-module)来定位和读取另一个包中的文件，那么如果该包位于另一个模块中，则测试将失败，该模块将位于模块缓存的版本控制子目录中，或者位于 replace 指令中指定的路径中。如果是这种情况，您可能需要将测试输入复制到模块中，或者将测试输入从原始文件转换为嵌入.go 源文件中的数据。

如果测试期望测试中的 go 命令以 GOPATH 模式运行，那么它可能会失败。如果是这种情况，您可能需要添加一个 go.mod 到要测试的源树，或显式地设置 GO111MODULE=off。

## 发布一个版本

最后，您应该标记并发布新模块的发布版本。如果还没有发布任何版本，这是可选的，但是没有正式的版本，下游用户将依赖使用伪版本([pseudo-versions](https://golang.org/cmd/go/#hdr-Pseudo_versions))的特定提交，而伪版本可能更难支持。

```bash
$ git tag v1.2.0
$ git push origin v1.2.0
```

新的 `go.mod` 文件为模块定义了一个规范导入路径，并添加了新的最低版本需求。如果您的用户已经使用了正确的导入路径，并且您的依赖项没有进行破坏（兼容性）的更改，则添加 go.mod 文件是向下（后）兼容（向旧版本兼容）的，但这是一个重要的改变，可能会暴露现有的问题。如果已有版本标记，则应增加次要版本([minor version](https://semver.org/#spec-item-7))。

## 导入和规范模块路径

每个模块在其 `go.mod` 文件中声明其模块路径。每个引用模块内包的 `import` 语句必须将模块路径作为包路径的前缀。然而，go 命令可能会通过许多不同的远程导入路径（[remote import paths](https://golang.org/cmd/go/#hdr-Remote_import_paths)）中包含模块的仓库。例如，`golang.org/x/lint` 和 `github.com/golang/lint` 都解析到包含[go.googlesource.com/lint](https://go.googlesource.com/lint)上托管的代码仓库。仓库中包含的[go.mod](https://go.googlesource.com/lint/+/refs/heads/master/go.mod)文件声明其路径为 `golang.org/x/lint`，因此只有该路径对应有效模块内容。

Go 1.4 提供了一种使用[// import](https://golang.org/cmd/go/#hdr-Import_path_checking)注释声明规范导入路径的机制，但是包的作者并不总会提供这些机制。因此，在导入模块之前编写的代码可能使用了模块工具的非规范导入路径，而出现不匹配的错误。当使用模块工具时，导入路径必须与规范的模块路径匹配，所以您可能需要更新导入语句:例如，您可能需要将 `import "github.com/golang/lint"` 更改为 `import "golang.org/x/lint"`。

模块的规范路径可能与其仓库路径不同的另一种场景发生在主版本 v2 或更高版本的 Go 模块上。主版本大于 v1 的 Go 模块必须在其模块路径中包含一个主版本后缀:例如，v2.0.0 版本必须有后缀/v2。但是，import 语句可能引用了模块中没有该后缀的包。例如，v2.0.1 版本的 github.com/russross/blackfriday/v2 的非模块用户可能将其导入为 github.com/russross/blackfriday，因此需要更新导入路径以包含/v2 后缀。

## 结论

对大多数用户来说，转换成 Go Modules 应该是一个简单的过程。由于非规范的导入路径或破坏依赖项中的更改，可能偶尔会出现一些问题。未来的文章将探讨发布新版本、v2 和其他版本，以及调试一些异常情况的方法。

为了提供反馈并帮助塑造 Go 依赖管理的未来，请向我们发送 bug 报告或经验报告。

感谢您所有的反馈和帮助改进 Go 的模块。

## 相关文章

* [Using Go Modules](https://blog.golang.org/using-go-modules)
* [Go Modules in 2019](https://blog.golang.org/modules2019)
* [A Proposal for Package Versioning in Go](https://blog.golang.org/versioning-proposal)
* [The cover story](https://blog.golang.org/cover)
* [The App Engine SDK and workspaces (GOPATH)](https://blog.golang.org/the-app-engine-sdk-and-workspaces-gopath)
* [Organizing Go code](https://blog.golang.org/organizing-go-code)

---

via: https://blog.golang.org/migrating-to-go-modules

作者：[Jean de Klerk](https://blog.golang.org)
译者：[lazytooo](https://github.com/lazytooo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
