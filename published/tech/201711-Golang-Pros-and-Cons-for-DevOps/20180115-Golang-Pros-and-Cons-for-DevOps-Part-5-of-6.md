已发布：https://studygolang.com/articles/12616

# Golang 之于 DevOps 开发的利与弊(六部曲之五)：跨平台编译

在这系列的第五篇文章，我们将讨论 Go 项目的跨平台编译.

在阅读这篇文章之前，请确保你已经阅读了[上一篇](https://studygolang.com/articles/12615)关于“Time包以及重载”的文章，或者订阅我们的博客更新提醒来获取此六部曲后续文章的音讯。

- [Golang 之于 DevOps 开发的利与弊（六部曲之一）：Goroutines, Channels, Panics, 和 Errors](https://studygolang.com/articles/11983)
- [Golang 之于 DevOps 开发的利与弊（六部曲之二）：接口实现的自动化和公有/私有实现](https://studygolang.com/articles/12608)
- [Golang 之于 DevOps 开发的利与弊（六部曲之三）：速度 vs. 缺少泛型](https://studygolang.com/articles/12614)
- [Golang 之于 DevOps 开发的利与弊（六部曲之四）：time 包和方法重载](https://studygolang.com/articles/12615)
- [Golang 之于 DevOps 开发的利与弊（六部曲之五）：跨平台编译，Windows，Signals，Docs 以及编译器](https://studygolang.com/articles/12616)
- Golang 之于 DevOps 开发的利与弊（六部曲之六）：Defer 指令和包依赖性的版本控制

## Golang 之利: 在 Linux 下编译 Windows 程序

对于我这类主要使用 Linux 的人来说，我对于偶尔不得不去应付 Windows 下的问题感到十分的痛苦。这句话在我写我们的 [Smart Agent™](https://www.bluematador.com/smart-agent) 的时候显得格外正确，它能同时跑在 Linux 和 Windows 上，并且会为了我们的日志管理及监控软件去深入探究这两个系统的底层相关问题。

因为我们的 agent 是用 Golang 写的，在 Linux 环境下把代码编译成能在 Windows 跑的程序是十分轻松的。大部分的工作是由两个运行 `go build` 命令时的传入参数：GOARCH 和 GOOS 所完成的.

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go_devops/goos-meme.jpg)

你可以运行 `go tool dist list` 去查看这两个参数所有的组合，在 Go 1.8 下一共有 38 种组合。以下的例子展示了如何在 AMD64 和 Intel i386 架构下编译适用于 Linux 和 Windows 的程序，而且你可以轻松看到如何创建一个 `Makefile` 来轻易地为各种不同的系统构建程序。

```bash
GOOS=linux GOARCH=amd64 go build -o bin/myapp_linux_amd64 myapp
GOOS=windows GOARCH=amd64 go build -o bin/myapp_windows_amd64 myapp
GOOS=linux GOARCH=386 go build -o bin/myapp_linux_386 myapp
```

## Cgo

如果你的项目使用 [cgo](https://golang.org/cmd/cgo/)，你可能会有点小麻烦。要让 cgo 代码完成跨平台编译，你需要在编译环境上安装正确的工具链。尽管现在距离我上次去直面 gcc 已经有好一段时间了，但是在 Unbuntu 16.04 的机器上找到正确的命令去安装这个工具链还是很容易的。以下是一些配置你的 cgo 编译环境的 one-liners:

```bash
# Install cgo dependencies
apt-get install -y gcc libsystemd-dev gcc-multilib
# cgo for linux/386
apt-get install -y libc6-dev-i386
# cgo for windows
apt-get install -y gcc-mingw-w64
```

随之改变的编译指令如下:

```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc go build -o bin/myapp_linux_amd64 myapp
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o bin/myapp_windows_amd64 myapp
CGO_ENABLED=1 GOOS=linux GOARCH=386 CC=gcc go build -o bin/myapp_linux_386 myapp
```

## Golang 之弊: 关于 Windows 部分的官方文档

Golang 的官方文档网站做得非常好。在上面有很多实用的例子，源代码的链接，当然还有测试小段代码的 playground。但是，golang 官方文档的一个缺陷会在你使用一些在 Windows 下有不同行为的代码时显现。

有些页面，比如 [os](https://golang.org/pkg/os/)，就在解释很多函数在 Unix 和 Windows 系统下的区别这点上做得非常好。但我很好奇的是，他们为什么在 Windows 下的标准库包括类似于 [Getegid](https://golang.org/pkg/os/#Getegid) 这样的函数的同时写下了类似于以下的注释:

```
Getegid 返回 caller 有效的 group id.
在 Windows 下，该函数返回 -1.
```

与其让我去搞清楚在 Windows 下这个返回值毫无意义，我宁愿编译器在发现目标系统是 Windows 之后会自动编译失败，不然这个函数其实什么都没做。

其他的这类关于 Windows 的页面完完全全就是一些空洞无用的语句，比如 [exec](https://golang.org/pkg/os/exec/) 页面:

```
注意这个包里的例子只适用于 Unix 系统。它们不能在 Windows 和 golang.org 以及 godoc.org 的 Go Playground 下运行。
```

## Golang 之弊: 开发者社区第三方包的 Windows 兼容性

只有在当你在 Unix 系统下完成了某个功能的开发之后，你会因为可以轻易跨平台编译而去尝试编译 Windows 程序之时，这个缺点才会显现。以前有几次我用了一些开源库去开发一个针对 Linux 系统的功能时，我意外地搞砸了 Windows 的程序 - 因为这些库使用了一些 Unix 特有的调用。以下是两个相对简单的解决方案:

### 1．为开源社区做贡献

如果你有时间并且想给这些项目做些贡献，创建一个支持 Windows 的 PR! 注意: 取决于你的开发环境，这个可能非常耗时，因为尽管编译 Windows 程序很容易，但是测试它们并不简单。

### 2．使用 Build Constraints

在 Golang 里，使用 [build constraints](https://golang.org/pkg/go/build/#hdr-Build_Constraints) 在编译时排除或者包含各种文件非常容易。例如，在一份仅为了 NT 编译的代码文件中去包含一个只支持 Windows 的依赖，你只需要这么做:

```go
// +build windows

package mypackage

import "github.com/bluematador/windows-only"
```

你也可以反着来 - 排除不支持的 GOOS:

```golang
// +build !windows

package mypackage

import "github.com/bluematador/linux-only"
```

这很棒的一点是，它让你更多地去理清你的代码，并让你可以通过源码中的 build constraints 拥有一个能在两个系统下都能调用但行为不同的函数。

---

via: https://blog.bluematador.com/golang-pros-cons-part-5-cross-platform-compiling

作者：[Matthew Barlocker](https://github.com/mbarlocker)
译者：[p31d3ng](https://github.com/p31d3ng)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
