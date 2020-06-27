首发于：https://studygolang.com/articles/24686

# Go 项目的布局
Kyle C. Quest
2017年9月12日 · 5 min 阅读

读过了 [`Tour of Go`](https:/tour.studygolang.com)，在 [https://play.studygolang.com/](https://play.studygolang.com/) 上把玩过，然后你感觉你准备好写一些代码了。很棒！但是，你不确定该如何组织你的项目。可以将代码放在你想放的任意地方吗？有没有组织代码的标准方式？如果想有多个应用程序的二进制文件呢？“go getable” 是指什么？你可能会问自己这些问题。

首先，你必须了解 Go 的工作空间。 [`How to Write Go Code`](https://golang.org/doc/code.html) 是个很好的起点。缺省地，Go 将所有代码保管在同一个工作空间，并期望所有代码都在同一个工作空间。这个地方由环境变量 `GOPATH` 来标识。对你来说这意味着什么？意味着你必须**将代码放在默认的工作空间**或者必须修改 `GOPATH` 环境变量，指向你自己的代码位置。不管哪种方式，项目的真正源代码都需要放在 `src` 子目录下（即 **`$GOPATH/src/your_project`** 或 `$GOPATH/src/github.com/your_github_username/your_project`）。技术上讲，如果你无需导入外部包且使用相对路径导入自己的代码，你的工程不一定非要放在工作空间里，但不推荐这样做。不过玩具项目或概念验证（Poc）项目这么做是可以的。Go 1.1 确实引入了模块的概念，允许你将项目代码放在 `GOPATH` 之外，且不受上述的导入限制，但直到现在这还是一个实验性的功能。

你已经将你的项目目录放在正确的地方。接下来呢？

对于你是唯一开发者的概念验证（Poc）项目或特别小的项目，将项目代码都写在根目录下的 `main.go` 里就够了。如果知道你的项目将会变得足够大或者它会上生产环境，而且其他人会贡献代码，那你就应该考虑至少采用这里罗列的项目布局样式中的一些。

有一些项目布局样式在 Go 生态系统中脱颖而出。`cmd` 和 `pkg` 目录是最常见的两个样式。你应当采用这些样式，除非你的项目特别小。

**`cmd`** 布局样式在你需要有多个应用程序二进制文件时十分有用。每个二进制文件拥有一个子目录（即 **`your_project/cmd/your_app`**）。这个样式帮助保持你的项目下的包（project/package） ‘go gettable’。什么意思？这意味着你可以使用 `go get` 命令拉取（并安装）你的项目，项目的应用程序以及库（比如，`go get github.com/your_github_username/your_project/cmd/appxg`）。你不必非要拆分应用程序文件，通过设置正确的 `go build` 标记你可以构建每个应用程序，但是由于不知道该构建哪个应用程序， `go get` 就无法正常工作了。官方的 [Go tools](https://github.com/golang/tools/tree/master/cmd)  是 `cmd` 布局样式的一个例子。很多知名的项目也使用了同样的样式：[Kubernetes](https://github.com/kubernetes/kubernetes/tree/master/cmd), [Docker](https://github.com/moby/moby/tree/master/cmd), [Prometheus](https://github.com/prometheus/prometheus/tree/master/cmd), [Influxdb](https://github.com/influxdata/influxdb/tree/master/cmd)。

**`pkg`** 布局样式也十分受欢迎。对新手 Go 开发者来讲这是最容易混淆的一个包结构概念，因为 Go 的工作空间就有一个同名的目录但那个目录有不同的用途（用来存储 Go 编译器构建的包的 object 文件）。`pkg` 目录是放置公共库的地方。它们可以被你的应用内部使用。也可供外部项目使用。这是你和你代码的外部使用者之间的非正式协定。其它项目会导入这些库并期望它们正常工作，所以在把东西放到这里前请三思。很多知名的项目使用了这个样式：[Kubernetes](https://github.com/kubernetes/kubernetes/tree/master/pkg), [Docker](https://github.com/moby/moby/tree/master/pkg), [Grafana](https://github.com/grafana/grafana/tree/master/pkg), [Influxdb](https://github.com/influxdata/influxdb/tree/master/pkg), [Etcd](https://github.com/coreos/etcd/tree/master/pkg).

`pkg` 目录下的某些库并不总是为了公共使用。为什么呢？因为很多现有的 Go 项目诞生在能隐藏内部包之前。一些项目将内部库放在 `pkg` 目录下，以便保持与其它部分代码结构的一致。另外一些项目将内部库放置在 `pkg` 目录之外另外的目录里。[Go 1.4](https://golang.org/doc/go1.4) 引入了使用 `internal` 隐藏内部库的能力。什么意思呢？如果你将代码放在 ‘internal’目录，外部项目则无法导入那些代码。即使是项目内部的其它代码，如果不在 `internal` 目录的父目录里，也无法访问这些内部代码。这个功能使用还不广泛因为它相对较新；但是作为一个额外（在 Go 用大小写区分函数可见性的规则之外）的控制层它有极大价值。很多知名的项目使用了这个样式：[Dep](https://github.com/golang/dep/tree/master/internal), [Docker](https://github.com/moby/moby/tree/master/internal), [Nsq](https://github.com/nsqio/nsq/tree/master/internal), [Go Ethereal](https://github.com/ethereum/go-ethereum/tree/master/internal), [Contour](https://github.com/heptio/contour/tree/master/internal)。

**`internal`** 目录是放置私有包的地方。你可以选择性地添加额外的结构来分离内部共享的库（比如，**`your_project/internal/pkg/your_private_lib`**）以及不希望别人导入的应用程序代码（比如, **`your_project/internal/app/your_app`**）。当你将全部私有代码都放在 ‘internal’ 目录，`cmd` 目录下的应用程序就可以被约束成一些小文件，其只需定义对应于应用程序二进制文件的 ‘main’ 函数。其余代码都从 `internal` 或 `pkg` 目录导入（Heptio 中的 [ark](https://github.com/heptio/ark/blob/master/cmd/ark/main.go)，以及 Grafana 中的 [loki](https://github.com/grafana/loki/blob/master/cmd/loki/main.go)，是这个 `微型 main 函数` 包样式的好例子）。

如果你 fork 并修改了外部项目的一块该如何？有些项目将这些代码放在 `pkg` 目录下，但更好的做法是将它放在顶层目录下的 **`third_party`** 目录，以便将你自己的代码和你从别人那里借用的代码区分开来。

你在项目里导入的外部包呢？它们去哪里？你有几个选项。你可以将它们放在项目以外。使用 `go get` 安装的包将保存在你的 Go 工作空间。大部分情况下可以正常工作，但视具体包而定，它可能会变得脆弱和不可预测，因为别人在构建你的项目时他们可能会拿到这个包的一个不向后兼容的版本。解决办法是 ‘vendoring’。使用 ‘vendoring’ 你通过将依赖与项目一起提交来将它们固定。 [Go 1.6](https://golang.org/doc/go1.6) 导入了一种标准的方式来 ‘vendor’ 外部包（在 Go 1.5 中是实验性功能）。将外部包放在 `vendor` 目录。这与 `third_party` 目录有何区别呢？如果你导入了外部代码且原样使用它就放在 `vendor` 目录。如果你使用的是修改版的外部包就放在 `third_party` 目录。

如果你想学习更多关于其他 Go 项目使用的项目结构请阅读 [‘Analysis of the Top 1000 Go Repositories’](http://blog.sgmansfield.com/2016/01/an-analysis-of-the-top-1000-go-repositories/)。它有点陈旧，不过依然有用。

一个真正的项目也会有另外的目录。你可以使用这个布局模版作为你的 Go 项目的起点：**[https://github.com/golang-standards/project-layout](https://github.com/golang-standards/project-layout)**。它涵盖了这篇博客里描述的 Go 项目布局样式并包括很多你需要的支持目录。

现在是时候写些代码了！如果你还没安装 Go 请查看这个 [quick setup guide for Mac OS X](https://medium.com/golang-learn/quick-go-setup-guide-on-mac-os-x-956b327222b8) （其他平台的安装也是类似的）。如果还没浏览过请你浏览 [‘Tour of Go’](https://tour.golang.org/) ，然后读一下 [’50 Shades of Go’](http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/) 去了解 Go 中最常见的坑，这会在你开始写代码和调试代码时节省很多时间。

* [Golang](https://medium.com/tag/golang)
* [Go](https://medium.com/tag/go)
* [Standards](https://medium.com/tag/standards)
* [Project Structure](https://medium.com/tag/project-structure)

---

via: https://medium.com/golang-learn/go-project-layout-e5213cdcfaa2

作者：[Kyle C. Quest](https://medium.com/@CloudImmunity)
译者：[krystollia](https://github.com/krystollia)
校对：[DingdingZhou](https://github.com/DingdingZhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
