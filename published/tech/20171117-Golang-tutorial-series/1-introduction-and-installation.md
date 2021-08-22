已发布：https://studygolang.com/articles/11706

# 介绍与安装

这是我们 Golang 系列教程的第一个教程。

## Golang 是什么

Go 亦称为 Golang (译注：按照 Rob Pike 说法，语言叫做 Go，Golang 只是官方网站的网址)，是由谷歌开发的一个开源的编译型的静态语言。

Golang 的主要关注点是使得高可用性和可扩展性的 Web 应用的开发变得简便容易。（译注：Go 的定位是系统编程语言，只是对 Web 开发支持较好）

## 为何选择 Golang

既然有很多其他编程语言可以做同样的工作，如 Python，Ruby，Nodejs 等，为什么要选择 Golang 作为服务端编程语言？

以下是我使用 Go 语言时发现的一些优点：

* 并发是语言的一部分（译注：并非通过标准库实现），所以编写多线程程序会是一件很容易的事。后续教程将会讨论到，并发是通过 Goroutines 和 channels 机制实现的。
* Golang 是一种编译型语言。源代码会编译为二进制机器码。而在解释型语言中没有这个过程，如 Nodejs 中的 JavaScript。
* 语言规范十分简洁。所有规范都在一个页面展示，你甚至都可以用它来编写你自己的编译器呢 :)
* Go 编译器支持静态链接。所有 Go 代码都可以静态链接为一个大的二进制文件（译注：相对现在的磁盘空间，其实根本不大），并可以轻松部署到云服务器，而不必担心各种依赖性。

## 安装

Golang 支持三个平台：Mac，Windows 和 Linux（译注：不只是这三个，也支持其他主流平台）。你可以在 [https://golang.org/dl/](https://golang.org/dl/) 中下载相应平台的二进制文件。（译注：因为众所周知的原因，如果下载不了，请到 [https://studygolang.com/dl](https://studygolang.com/dl) 下载）

### Mac OS

在 [https://golang.org/dl/](https://golang.org/dl/) 下载安装程序。双击开始安装并且遵循安装提示，会将 Golang 安装到 `/usr/local/go` 目录下，同时 `/usr/local/go/bin` 文件夹也会被添加到 `PATH` 环境变量中。

### Windows

在 [https://golang.org/dl/](https://golang.org/dl/) 下载 MSI 安装程序。双击开始安装并且遵循安装提示，会将 Golang 安装到 `C:\Go` 目录下，同时 `c:\Go\bin` 目录也会被添加到你的 `PATH` 环境变量中。

### Linux

在 [https://golang.org/dl/](https://golang.org/dl/) 下载 tar 文件，并解压到 `/usr/local`。

请添加 `/usr/local/go/bin` 到 `PATH` 环境变量中。Go 就已经成功安装在 `Linux` 上了。

在本系列下一部分 *Golang 系列教程第 2 部分: Hello World* 中，我们将会建立 Go 的工作区，编写我们第一个 Go 程序 :)

请提供给我们宝贵的反馈和意见。感谢您的阅读 :)

**下一教程 - [Hello World](https://studygolang.com/articles/11755)**

---

via: https://golangbot.com/golang-tutorial-part-1-introduction-and-installation/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119), [Unknwon](https://github.com/Unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
