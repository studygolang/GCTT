已发布：https://studygolang.com/articles/12293

# Go 的大门已经打开

Go 在近 10 年间已经快速的成为了非常流行并且成功的系统编程语言。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-door/BATMAN_GOPHER.png)

> Go 带来的是 Gotham （译者注：哥谭，是蝙蝠侠的家乡，充满犯罪的黑暗城市，大意就是乱世出枭雄，Go 正是这样是个超级英雄而存在） ，它将分布式系统从黑暗中摆脱出来。 插画归功于 [Ashley McNamara](https://twitter.com/ashleymcnamara?ref_src=twsrc%5Egoogle%7Ctwcamp%5Eserp%7Ctwgr%5Eauthor) 和 [Renee French](https://twitter.com/reneefrench?ref_src=twsrc%5Egoogle%7Ctwcamp%5Eserp%7Ctwgr%5Eauthor) （译者注：两位都是设计 Go 那只老鼠的大神）

在 Go 之前，C、C++ 、Java 还有 C# 在编程界都是大腕。Go 直到今天还是一个婴儿，但是它却为你而来。

它为开源软件打开了一个新的世界。这样一个完美的语言来的正是时候，它引发了一场计算的新时代。所有的这些知名的软件都是用 Go 编写的：

- [Kubernetes](https://kubernetes.io/) ： 适用于大公司的一个分布式容器管理服务产品。还有些公司已经开始构建，托管和支持这个软件。
- [etcd](https://github.com/coreos/etcd) ：一个健壮的高一致性的 key-value 数据库。它被用在分布式系统中处理核心任务，实际上它成为了和谷歌 Chubby （译者注：一个分布式锁服务） 一样的开源软件。它是 Kubernetes 核心组件。
- [Docker](http://docker.com/) ：它是目前市面上非常流行的容器引擎。也是 Kubernetes 的核心组件。

Cloud Native 将不可能抛弃 Go ，[Cloud Native Computing Foundation](https://www.cncf.io) （CNCF 基金会）同样也不会。这仅仅是个开始。Go 也接管了其它部分开源软件，更别提那些大公司内部的的基础设施。

实际上，这就是为什么在开源项目（或其他项目）上大家倾向选择使用 Go 来构建产品系统和大型系统。

## 工具链

大家都在 [Go playground](https://play.golang.org/) （译者注：一个Golang的在线编辑网站） 上开始尝试 Go 语言。你只需要打开一个网站，写一些代码，然后运行。无需安装，在哪都能开始写代码，这是一个不错的体验。

> *无论你在做什么，你只要关注你做的。Go 已经为你找到了解决问题的工具*

然后你去下载一个 toolchain （译者注：工具链，一般指的就是编译工具）—— 一个二进制 `go` 文件。你可以通过运行 `go build` 命令来获一个生产级别的软件。无需学习 GCC toolchain ，C 语言，Linux ，共享对象，JVM 或其它相关技术。

不管你在开发什么，你只需专注开发的业务，而不是你需要哪些工具。Go已经为你解决了相应的工具了。

## 一个二进制文件

在以前的时代，编译代码后，你不能仅仅只是运行它，因为它依赖系统上的其他组件：如 共享对象、JVM 等。

> *仅需要传送一个二进制文件到服务器即可*

`go build` 会输出一个可执行的二进制文件。将它发送到你的服务器上。它之所以能运行是因为已经将所需要的东西都编译进去了。
这个简单的案例展示了它的强大。好消息是你的部署过程将比以前简单的多。—— 仅需要将二进制文件传送到你的服务器即可。
你甚至可以通过少量的环境变量在不同的系统上构建。这个特性非常适合 CLIs （译者注：命令行工具）以下是最成功的几个案例：
- `tecdctl` : etcd 下的 CLI
- `kubectl` : Kubernetes 下的 CLI
- `docker` : Docker 服务下的 CLI
- `dep` : 最新的 Golang 依赖管理工具下的 CLI

## 并发

云已经不是什么新东西了，它是一个标准。虚拟化和容器的运行与终止没有任何通知，数据流的来来往往是不可靠的，RPC 的发送与重试也是频繁的。

> *还在线程的问题上苦苦挣扎，你只需要将 `go` 关键字加在你的函数前面，它就是并发运行了。*

当下的软件需要的是能高效而正确的运行，它需要并行的操作这些所有的事件。可容错的分布式架构在今天也是一个标配了。

在以前的时代，你不得不去关注系统线程，锁，条件变量（译者注：多线程开发名词）等等。要想写出一个高效、有弹性的系统，就意味着要学习或者发明一个并发框架，然后才能继续前进。

现在你可以获得一个简单易懂的内置基本操作。 Goroutines 和 channels 是有意义的，因为它模仿的是真实的情况。

你只需要在一个函数前加上 `go` 关键字，它就会以并发的方式运行。你可以很容易的理解这些并发功能，并且可以专注你的业务开发。是否看到了一个趋势？

Go 是一个无锁的强大的分布式系统，因为从根本上让并发操作更简单了。

这就是为什么我们能看到这样一个更有弹性，更快速，并且高效利用CPU的软件。用 Go ，事实上你可以开发你在研究资料中找到的东西。

## 垃圾回收

关于 Go 和 系统编程 GC（译者注：指垃圾回收机制） 通常是一个有争议的话题。

> *我们能同时拥有系统编程和垃圾回收*

在 C / C++ 中，你可以完全控制内存。什么时候如何分配和释放内存由你来决定。JVM 则是通过垃圾回收器这种方式来取代你的控制。

总的来说，GC 很方便，但世上总是有些人不想用它。难啊。

手动管理内存很难，而且在进行并发时更难。 在 Go 之前，我们面临着相互冲突的挑战：我们需要一个不会泄漏内存或者破坏程序的框架，但是程序员又必须明白这一点。

最后的结果就是有上百万的库以不同的方式进行权衡，迫使让你的程序以一种独特唯一的方式运行。

以 Go 的立场来说：

- 1.没有什么比无需考虑内存要简单了
- 2.我们可以 “坐在中间” 构建包含 GC 的系统编程语言

Go 是一个包含 GC 的系统编程语言。这是不会改变的。

事实上，GC 已经爆炸式的促进了 Go 。下面这些是 Go 垃圾回收的边界情况，可能会出现一些问题。但是很多[看法](https://docs.google.com/document/d/16Y4IsnNRCN43Mx0NZc5YXZLovrHvvLhK_h0KN8woTO4/edit)都是为了让它更好的运行，默认 90% 是这样。

如果你遇到了 10% 的情况，你可以进行一个新调优，甚至比 JVM 垃圾回收调优更简单。

## 标准库

Go 标准库是最好的商业库之一。它不大但是却覆盖了 80% 的常用功能，并且不复杂却可以为你完成复杂的事情。

> *与这些思想达成一致，编写并复用它们*

流行的 Go 包大都是高质量的，应为它们构建在一个高质量的标准库上。
比标准库更重要的是要理解代码的思想，它鼓励使用 `interface` 和惯例用法。例如：
- [io.Reader](https://godoc.org/io#Reader) 和 [io.Writer ](https://godoc.org/io#Writer) 是以 “管道” 的方式跨过函数边界。这些接口很可能遍布在 Go 的生态系统中
- [context](https://godoc.org/context) 是一个提供者，取消，超时和发送到 goroutines 中的值
-  [erros](https://godoc.org/builtin#error) 是从函数中返回错误描述的方法

这些包通常都认同这些或其它一些惯例用法，所以它们能平滑的在一起运行。
它们的理念一致，编写并复用它们。

## 结论

正如文章开头所说，完美的 Go 语言来得正是时候。

> *Go 将成为软件工程中几个大型领域的标准编程语言*

我已经阐述了原因，**我们可以打开很多强大的开源软件看看，Go 让许多事情变得简单起来。**

我希望 Go 能继续成为其它领域的标准——前端服务（替代 Rails / Node
.js），CLIs （替换许多脚本语言），也许还能替换 GUIs 和 移动 APP 。

正值 Go [8 周年](https://blog.golang.org/8years) ，它快速地崛起了。但下一个 8 年它的趋势是否会扩大10倍。

还是那句话，Go 将成为软件工程中几个大型领域的标准编程语言。

---

via: https://medium.com/@arschles/the-doors-go-has-opened-a4b5d0f10ea7

作者：[Aaron Schlesinger](https://medium.com/@arschles)
译者：[zhuCheer](https://github.com/zhuCheer)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
