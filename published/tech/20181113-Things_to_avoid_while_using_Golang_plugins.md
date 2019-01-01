首发于：https://studygolang.com/articles/17365

# 使用 golang-plugins 时要避免的事情

我们正计划开源我们的项目。其中有一些关于授权逻辑的代码是我们公司高度定制的，我们需要在提供相同功能的情况下剔除掉这部分代码。并且，使任何人在实现自己的授权逻辑时，不需要重新编译所有代码。

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/things_to_avoid_while_using_golang_plugins/1_X2zMAhjAP445suAyfvPUOw.png)

我们的代码库在我所钟爱的 Go 中。在寻找可能的实现方案时，我们发现了 [golang-plugins](https://golang.org/pkg/plugin/) 。使用 *golang-plugins*，你可以通过 `go build -buildmode=plugin` 命令构建的文件导入函数和变量。

我们很快就为该需求准备了一个原型实现。**但是，在此期间，我们遇到了一些问题。**以下是我们面临的一些问题以及避免这些问题的方法：

## 1) 不同的 Go 版本

插件实现和主应用程序都必须使用**完全相同**的 Go 工具链**版本**构建。根据你的 Go 版本，您将收到如下错误：

> panic: plugin.Open("simpleuser.plugin"): plugins must be built with the same version of the Go toolchain as the main application

或者

> panic: plugin.Open("simpleuser.plugin"): plugin was built with a different version of package GitHub.com/alperkose/golangplugins/user

**解决方法：**无。

由于插件提供的代码将与主代码在相同的进程空间中运行，因此编译的二进制文件应与主应用程序 100％ 兼容。

## 2) 不同的 GOPATH

插件实现和主应用程序都必须使用**完全相同的 GOPATH 构建**。你会得到的错误是：

> panic: plugin.Open("differentgopath.plugin"): plugin was built with a different version of package GitHub.com/alperkose/golangplugins/user

**解决方法：**使用相同的 GOPATH （官方 docker 镜像中的 GOPATH 是：`/go`）。问题[issue#19223](https://github.com/golang/go/issues/19233) 就是针对类似情况的。

示例代码：https://github.com/alperkose/golangplugins

## 3) 使用 vendor 文件夹

这似乎与 `＃2` 有点相关，但如果你在插件或主应用程序中使用 `vendor` 文件夹，你会得到一个非常奇怪的错误：

> panic: interface conversion: plugin.Symbol is func() user.Provider, not func() user.Provider

如果你仔细观察，你会发现期望 `func() user.Provider` 和实际 `func() user.Provider` 方法签名是一样的。这是一个非常令人困惑的错误，但在 `1.8` 版本之后的所有版本中都存在。

**解决方法：**在构建二进制文件时复制 `vendor` 文件夹中的所有内容到 `gopath` 中。这是一个非常 " 脏 " 的解决办法，你需要为插件和主应用程序都执行此操作。如果构建其中一个二进制文件必须使用 `vendor` 文件夹时，则无法使用 `golang-plugin` 解决方案。

我们的例子中，在构建阶段复制 `vendor` 文件夹中的内容到 `gopath` ，然后删除 `vendor` 文件夹，都是在 docker 镜像中完成的，因此我们的本地文件夹结构在开发期间得以保持不变。

> ...
> RUN cp -r vendor/* $GOPATH/src && rm -rf vendor/
> ...

问题[issue#18827](https://github.com/golang/go/issues/18827) 是针对此类情况打开的。

示例代码：[https://github.com/alperkose/golangplugins](https://github.com/alperkose/golangplugins)

## 4）不同版本的公共依赖项

在插件中的任何依赖项应该与主应用程序中的依赖项版本相同。

同样，由于插件提供的代码将与主代码在相同的进程空间中运行，因此编译的二进制文件应与主应用程序 100 ％兼容。当你编译二进制文件时，第三方软件包也会编译在该二进制文件中，但如果进程空间中存在相同函数的不同版本，二进制文件将会出现错误。

**解决方法：**使用包管理器并确保依赖项是相同的版本。

你可以在这里找到关于这个问题的评论：[https://github.com/whiteboxio/flow/issues/3](https://github.com/whiteboxio/flow/issues/3)

## 5）构建静态二进制文件

你无法将插件编译为静态二进制文件。我喜欢静态二进制文件，因为它避免了在 docker 镜像中有基本镜像的要求。使用 docker scratch image 可以提供最小的 docker 镜像并减少一个很大的依赖。当我尝试为插件构建静态二进制文件时失败了，可以在[这篇文章](https://medium.com/@diogok/on-golang-static-binaries-cross-compiling-and-plugins-1aed33499671) 中找到原因。

**解决方法：**无。你需要使用 CGO 进行编译，如果使用的是 docker ，则需要在 Dockerfile 中使用基本映像。

---

## 结论

我认为 *golang-plugins* 还不是一个成熟的解决方案。它迫使你的插件实现与主应用程序产生高度耦合。即使你可以控制插件和主应用程序，最终结果也非常脆弱且难以维护。如果插件的作者对主应用程序没有任何控制权，开销会更高。

所有这些问题促使我们考虑替代方案，最后我们选择使用 [hashicorp](https://github.com/hashicorp/go-plugin) 插件包。它基于 RPC 通信并为我们提供了足够的灵活性，尽管它也有自己的局限性，但比 *golang-plugins* 更容易克服。

## 源码

你可以在此处找到用于测试问题＃ 2 和＃ 3 的源码：[https//github.com/alperkose/golangplugins](https://github.com/alperkose/golangplugins)

---

via: https://medium.com/@alperkose/things-to-avoid-while-using-golang-plugins-f34c0a636e8

作者：[Alper Köse](https://medium.com/@alperkose)
译者：[herowk](https://github.com/herowk)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
