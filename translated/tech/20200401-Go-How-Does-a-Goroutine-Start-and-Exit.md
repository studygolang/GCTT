# Go 协程的开启和退出

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200401-Go-How-Does-a-Goroutine-Start-and-Exit/00.png)

ℹ️本文基于 Go 1.14。

在 Go 中，协程就是一个包含程序运行时的信息的结构体，如栈，程序计数器，或者它当前的 OS 线程。调度器需要在协程开启和退出时分配资源，这两个阶段需要谨慎管理。

*如果你想了解更多关于栈和程序计数器的信息，我推荐你阅读我的文章 [Go：协程切换时涉及到哪些资源？](https://medium.com/a-journey-with-go/go-what-does-a-goroutine-switch-actually-involve-394c202dddb7)。*

## 开启

开启一个协程的处理过程相当简单。我们用一个程序作为例子：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200401-Go-How-Does-a-Goroutine-Start-and-Exit/01.png)

---
via: https://medium.com/a-journey-with-go/go-how-does-a-goroutine-start-and-exit-2b3303890452

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出