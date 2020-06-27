首发于：https://studygolang.com/articles/28457

# Go 协程的开启和退出

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200401-Go-How-Does-a-Goroutine-Start-and-Exit/00.png)

ℹ️本文基于 Go 1.14。

在 Go 中，协程就是一个包含程序运行时的信息的结构体，如栈，程序计数器，或者它当前的 OS 线程。调度器还必须注意 Goroutine 的开始和退出，这两个阶段需要谨慎管理。

*如果你想了解更多关于栈和程序计数器的信息，我推荐你阅读我的文章 [Go：协程切换时涉及到哪些资源？](https://medium.com/a-journey-with-go/go-what-does-a-goroutine-switch-actually-involve-394c202dddb7)。*

## 开启

开启一个协程的处理过程相当简单。我们用一个程序作为例子：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200401-Go-How-Does-a-Goroutine-Start-and-Exit/01.png)

`main` 函数在打印信息之前开启了一个协程。由于协程会有自己的运行时间，因此 Go 通知运行时配置一个新协程，意味着：

- 创建栈
- 收集当前程序计数器或调用方数据的信息
- 更新协程内部数据，如 ID 或 状态

然而，协程没有立即获取运行时状态。新创建的协程被加入到了本地队列的最前端，会在 Go 调度的下一周期运行。下面是现在这种状态的示意图：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200401-Go-How-Does-a-Goroutine-Start-and-Exit/02.png)

把协程放在队列的前端，这样它就会在当前协程运行之后第一个运行。如果有工作窃取发生，它不是在当前线程就是在另一个线程运行。

*我推荐你阅读我的文章 [Go: Go 调度器中的工作窃取](https://medium.com/a-journey-with-go/go-work-stealing-in-go-scheduler-d439231be64d)来获取更多信息。*

在汇编指令中也可以看到协程的创建过程：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200401-Go-How-Does-a-Goroutine-Start-and-Exit/03.png)

协程被创建并被加入到本地协程队列后，它直接执行主函数的下一个指令。

## 退出

协程结束时，为了不浪费 CPU 资源，Go 必须调度另一个协程。这也使协程可以在以后复用。

*在我的文章 [Go: 协程怎么复用？](https://medium.com/a-journey-with-go/go-how-does-go-recycle-goroutines-f047a79ab352)中你可以找到更多信息。*

然而，Go 需要一个能识别到协程结束的方法。这个方法是在协程创建时控制的。创建协程时，Go 在将程序计数器设置为协程真实调用的函数之前，将堆栈设置为名为 `goexit` 的函数。这个技巧可以使协程在结束时必须调 `goexit` 函数。下面的程序可以使我们理解得更形象：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200401-Go-How-Does-a-Goroutine-Start-and-Exit/04.png)

根据输出信息进行堆栈追踪：

```bash
/path/to/src/main.go:16
/usr/local/go/src/runtime/asm_amd64.s:1373
```

用汇编写的 `asm_amd64` 文件包含这个函数：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200401-Go-How-Does-a-Goroutine-Start-and-Exit/05.png)

之后，Go 切换到 `g0` 调度另一个协程。

我们也可以调用 `runtime.Goexit()` 来手动终止协程：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200401-Go-How-Does-a-Goroutine-Start-and-Exit/06.png)

这个函数首先运行 defer 中的函数，然后会运行前面在协程退出时我们看到的那个函数。

---

via: https://medium.com/a-journey-with-go/go-how-does-a-goroutine-start-and-exit-2b3303890452

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
