# Modules 第 1 部分：为什么和做什么


## 引言

自 Go 语言最初发布以来，有三个关键问题一直困扰着开发者，modules 的出现则为这三个问题提供了一个完整的解决方案，使得开发者：

* 能够在 GOPATH 工作区之外使用 Go 代码；
* 能够对依赖包进行版本控制并识别要使用的最兼容版本；
* 能够使用 Go 原生工具来管理依赖包；

随着 Go 语言 1.13 版本的发布，这三个问题已经成为了“过去时”。在过去的两年中，Go 语言团队花费了很多精力才让所有人达到这一步。在本文中，我将重点介绍从 GOPATH 到 modules 的过渡以及 modules 所解决的问题。在此过程中，我将只提供足够的术语，以便您可以更好地了解 modules 是如何在较高的层面上起作用的，也许更为重要的是，为什么它以这样的方式起作用。


## GOPATH

使用 GOPATH 在磁盘上为 Go 工作区提供物理位置已经为 Go 语言开发者提供了很好的服务。不幸的是，对于非 Go 语言开发者来说，这却是一个瓶颈，因为他们可能需要时不时的进行 Go 项目，并且没有设置 Go 工作区。Go 语言团队想要解决的问题之一便是允许将 Go 代码仓库克隆到磁盘上的任何位置（GOPATH 之外），同时 Go 工具能够对其进行定位，构建和测试。

图1  

![108_figure1.png](https://www.ardanlabs.com/images/goinggo/108_figure1.png)

图 1 展示的是 [conf](https://github.com/ardanlabs/conf) 包的 GitHub 仓库。这个仓库是一个能为应用程序处理配置信息提供支撑的包。在 modules 出现之前，如果您要使用该包，可以通过 `go get` 将这个仓库以其规范名为相对路径克隆到 GOPATH 中，其中包的规范名是远程仓库的根目录和仓库名的组合。

例如，如果您运行 `go get github.com/ardanlabs/conf`，那么代码将会被克隆到路径 `$GOPATH/src/github.com/ardanlabs/conf` 下。正是因为有了 GOPATH 以及仓库的规范名称，所以不论开发者选择将工作区置于何处，Go 工具都可以找到代码。


## 解析导入

清单 1  
[github.com/ardanlabs/conf/blob/master/conf_test.go](https://www.ardanlabs.com/blog/2019/10/github.com/ardanlabs/conf/blob/master/conf_test.go)

```go
01 package conf_test
02
03 import (
...
10     "github.com/ardanlabs/conf"
...
12 )
```

清单 1 展示的是 `conf` 仓库的测试文件 `conf_test.go` 中 import 部分的代码片段。当测试代码在包名中使用 `_test` 这样的约定命名方式时（如您在第 01 行看到的那样），意味着测试代码与被测试的代码存在于不同的包中，并且测试代码必须要像任何外部用户一样导入被测试的包。您可以在第 10 行看到该测试文件是怎样使用仓库的规范名来导入 `conf` 包的。借助 GOPATH 机制，可以将这个导入的包解析到磁盘上的具体位置，然后，Go 工具就可以可以定位，构建和测试代码了。

假使 GOPATH 不再存在并且包所处于的文件夹结构与仓库的规范名称也不再一致时，会怎样呢？

清单 2  
```go
import "github.com/ardanlabs/conf"

// GOPATH 模式：包在磁盘上的物理位置与 GOPATH 
// 和仓库的规范名相匹配。
$GOPATH/src/github.com/ardanlabs/conf


// Module 模式：包在磁盘上的物理位置与仓库的规范名称
// 不一致。
/users/bill/conf
```

清单 2 展示了将 `conf` 仓库克隆到磁盘上任意位置时所遇到的问题。当开发者可以选择将代码克隆到所希望的任何位置时，所有必需的用来将导入的包解析到磁盘上具体物理位置的信息都消失了。

解决此问题的方法是使用一个包含仓库规范名的特殊文件。用该文件在磁盘上的位置来代替 GOPATH，无论仓库被克隆到何处，Go 工具都能够利用在其中定义的仓库规范名来解析导入。

这个特殊的文件被命名为 [go.mod](https://golang.org/cmd/go/#hdr-The_go_mod_file)，而在其中定义的仓库规范名将代表称为 module 的新实体 。

清单 3  
[github.com/ardanlabs/conf/blob/v1.1.0/go.mod](https://www.ardanlabs.com/blog/2019/10/github.com/ardanlabs/conf/blob/v1.1.0/go.mod)

```go
01 module github.com/ardanlabs/conf
02
...
06
```

清单 3 显示了 `conf` 仓库中 `go.mod` 文件的第一行。该行定义了 module 的名称，开发者可以用这个 module 名来索引该仓库中的任何代码。现在，把仓库克隆任何位置都是没问题的，因为 Go 工具可以使用 module 文件的位置和 module 名来解析任何内部导入，例如导入上述的测试文件。

借助 module 的概念，就可以将代码克隆到磁盘上的任何位置了，下一个将要解决的问题是支持将代码捆绑在一起并进行版本控制。


## 捆绑和版本控制

大多数版本控制系统都允许我们对代码仓库的任意提交点打标签（例如：v1.0.0、v2.3.8 等），这些标签被认为是不可变的，通常被用于发布新功能。

图 2  

![108_figure2.png](https://www.ardanlabs.com/images/goinggo/108_figure2.png)

图 2 展示 `conf` 包的作者给该仓库标记了三个不同的版本号，这些标签遵循着 [语义化版本号](https://semver.org/) 的格式。

借助版本控制工具，开发者可以通过特定标签将对应版本的 `conf` 包克隆到磁盘上。然而，首先我们需要回答几个问题：

* 应该使用哪个版本的包？
* 怎么知道哪个版本与我正在编写和使用的所有代码都兼容？

回答完这两个问题后，您还需要回答第三个问题：
* 要将仓库克隆到何处，以便 Go 工具可以找到和访问它？

然后情况便变得更糟了，您不能在自己的项目中使用某个版本的 `conf` 包，除非您还克隆了所有 `conf` 所依赖包的仓库，这是您的所有项目都会遇到的依赖项传递问题。

在 GOPATH 模式下的解决方案是使用 `go get` 来识别并将所有依赖包的仓库克隆到您的 GOPATH 中。但是，这并不是一个完美的解决方案，因为 `go get` 只懂得如何为每个依赖包克隆仓库以及更新仓库 `master` 分支的最新代码。在编写代码初期，从依赖包仓库的 `master` 分支拉取代码或许没什么大碍。但是，在几个月（或几年）后，因为依赖包的独立演进，依赖包仓库的 `master` 分支的最新代码可能与您的项目已不再兼容。这是因为您的项目没有遵循版本标签，因此任何包的升级都可能包含破坏性的变更。

在新的 Go module 模式下，使用 `go get` 将所有依赖包的仓库克隆到一个单一的预定义好的工作空间中不再成为首选。另外，你需要一个适用于整个项目的方法，来引用每个依赖包的兼容版本。然后便是支持在你的项目中使用同一个依赖包的不同主版本，以防止你的依赖包正在导入主版本号不同的同一个包。

尽管，针对这些问题的若干解决方案已经以社区开发工具的形式存在了（例如：dep、godep、glide 等），但是 Go 语言需要的是一个完整的解决方案。这个解决方案便是复用 module 文件来维护一个版本化的依赖列表，其中包括直接或间接的依赖。然后将任何给定版本的仓库都视为一个不变的代码捆绑。这个版本化的不可变捆绑称为 module。


## 完整的解决方案

图 3

![108_figure3.png](https://www.ardanlabs.com/images/goinggo/108_figure3.png)

图 3 展示了仓库和 module 之间的关系。它表明了导入是怎样引用存储在给定版本 module 内的包。在图 3 中，版本号为 1.1.0 的 module `conf` 中的代码可以从版本号为 0.3.1 的module `go-cmp` 中导入包 `cmp`。由于依赖项信息已经在 `conf` module 中列出（通过 module 文件），因此 Go 工具可以获取其中任何 module 的特定版本，于是便可以成功构建。

一旦有了 modules，很多工程机会就会浮现出来：

* 您可以为构建、保留、认证、验证、获取、缓存和重用 modules 提供支持（除了某些例外），以供全世界的 Go 开发者使用。
* 您可以建一个代理服务器来支持不同的版本控制系统并提供某些上述的支持。
* 您可以验证一个 module （对于任何给定的版本）始终包含完全相同的代码，而不论它被构建了多少次，以及从何处获取、由谁提供。

关于 modules 所能完美支持的特性，已经由 Go 语言团队在 Go 1.13 发行版本中提供。


## 结论

这篇文章试图为理解 module 是什么以及 Go 语言团队如何使用该解决方案奠定基础。当然，仍有许多方面需要讨论，例如：

* 怎样选用 module 的特定版本？
* module 文件的结构是怎样的，有哪些选项可用于控制对 module 的选择？
* module 是怎样构建、获取和缓存在本地以解决导入问题的？
* module 是怎样验证符合语义化版本号契约的？
* 在您的项目中应当怎样使用 modules 以及最佳实践是什么？

在后续的文章中，我计划提供对这些以及更多其他问题的理解。现在，请确保您已经了解仓库，包和 modules 之间的关系。如有任何疑问，请随时在 Slack 上找到我。那里有一个很棒的频道 `#modules`，其中的人们随时可以提供帮助。


## Module 文档

有许多关于 Go 语言的文档，下面是由 Go 语言团队发布的一些文章和视频。

[Modules The Wiki](https://github.com/golang/go/wiki/Modules)  
[1.13 Go Release Notes](https://golang.org/doc/go1.13#modules)  
[Go Blog: Module Mirror and Checksum Database Launched](https://blog.golang.org/module-mirror-launch)  
[Go Blog: Publishing Go Modules](https://blog.golang.org/publishing-go-modules)  
[Proposal: Secure the Public Go Module Ecosystem](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md)  
[GopherCon 2019: Katie Hockman - Go Module Proxy: Life of a Query](https://www.youtube.com/watch?v=KqTySYYhPUE) 


---

via: https://www.ardanlabs.com/blog/2019/10/modules-01-why-and-what.html

作者：[William Kennedy](https://www.ardanlabs.com/)
译者：[anxk](https://github.com/anxk)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
