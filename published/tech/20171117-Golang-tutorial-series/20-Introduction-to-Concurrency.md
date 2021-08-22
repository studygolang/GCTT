已发布：https://studygolang.com/articles/12341

# 第 20 篇：并发入门

欢迎来到我们 [Golang 系列教程](https://studygolang.com/subject/2)的第 20 篇。

**Go 是并发式语言，而不是并行式语言**。在讨论 Go 如何处理并发之前，我们必须理解何为并发，以及并发与并行的区别。

## 并发是什么？

并发是指立即处理多个任务的能力。一个例子就能很好地说明这一点。

我们可以想象一个人正在跑步。假如在他晨跑时，鞋带突然松了。于是他停下来，系一下鞋带，接下来继续跑。这个例子就是典型的并发。这个人能够一下搞定跑步和系鞋带两件事，即立即处理多个任务。

## 并行是什么？并行和并发有何区别？

并行是指同时处理多个任务。这听起来和并发差不多，但其实完全不同。

我们同样用这个跑步的例子来帮助理解。假如这个人在慢跑时，还在用他的 iPod 听着音乐。在这里，他是在跑步的同时听音乐，也就是同时处理多个任务。这称之为并行。

## 从技术上看并发和并行

通过现实中的例子，我们已经明白了什么是并发，以及并发与并行的区别。作为一名极客，我们接下来从技术的角度来考察并发和并行。:)

假如我们正在编写一个 Web 浏览器。这个 Web 浏览器有各种组件。其中两个分别是 Web 页面的渲染区和从网上下载文件的下载器。假设我们已经构建好了浏览器代码，各个组件也都可以相互独立地运行（通过像 Java 里的线程，或者通过即将介绍的 Go 语言中的 [Go 协程](https://studygolang.com/articles/12342)来实现）。当浏览器在单核处理器中运行时，处理器会在浏览器的两个组件间进行上下文切换。它可能在一段时间内下载文件，转而又对用户请求的 Web 页面进行渲染。这就是并发。并发的进程从不同的时间点开始，分别交替运行。在这里，就是在不同的时间点开始进行下载和渲染，并相互交替运行的。

如果该浏览器在一个多核处理器上运行，此时下载文件的组件和渲染 HTML 的组件可能会在不同的核上同时运行。这称之为并行。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-series/concurrency-parallelism-copy.png)

并行不一定会加快运行速度，因为并行运行的组件之间可能需要相互通信。在我们浏览器的例子里，当文件下载完成后，应当对用户进行提醒，比如弹出一个窗口。于是，在负责下载的组件和负责渲染用户界面的组件之间，就产生了通信。在并发系统上，这种通信开销很小。但在多核的并行系统上，组件间的通信开销就很高了。所以，并行不一定会加快运行速度！

## Go 对并发的支持

Go 编程语言原生支持并发。Go 使用 [Go 协程](https://studygolang.com/articles/12342)（Goroutine） 和信道（Channel）来处理并发。在接下来的教程里，我们还会详细介绍它们。

并发的介绍到此结束。请留下你的反馈和评论。祝你愉快。

**上一教程 - [接口 - II](https://studygolang.com/articles/12325)**

**下一教程 - [Go 协程](https://studygolang.com/articles/12342)**

---

via: https://golangbot.com/concurrency/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
