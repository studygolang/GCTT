首发于：https://studygolang.com/articles/28460

# Go：异步抢占

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200501-Go-Asynchronous-Preemption/00.png)

ℹ️ 本文基于 Go 1.14。

抢占是调度器的重要部分，基于抢占调度器可以在各个协程中分配运行的时间。实际上，如果没有抢占机制，一个长时间占用 CPU 的协程会阻塞其他的协程被调度。1.14 版本引入了一项新的异步抢占的技术，赋予了调度器更大的能力和控制力。

*我推荐你阅读我的文章[”Go：协程和抢占“](https://medium.com/a-journey-with-go/go-goroutine-and-preemption-d6bc2aa2f4b7)来了解更多之前的特性和它的弊端。*

## 工作流

我们以一个需要抢占的例子来开始。下面一段代码开启了几个协程，在几个循环中没有其他的函数调用，意味着调度器没有机会抢占它们：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200501-Go-Asynchronous-Preemption/01.png)

然而，当把这个程序的追踪过程可视化后，我们清晰地看到了协程间的抢占和切换：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200501-Go-Asynchronous-Preemption/02.png)

我们还可以看到表示协程的每个块儿的长度都相等。所有的协程运行时间相同（约 10 到 20 毫秒）。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200501-Go-Asynchronous-Preemption/03.png)

异步抢占是基于一个时间条件触发的。当一个协程运行超过 10ms 时，Go 会尝试抢占它。

抢占是由线程 `sysmon` 初始化的，该线程专门用于监控包括长时间运行的协程在内的运行时。当某个协程被检测到运行超过 10ms 后，`sysmon` 向当前的线程发出一个抢占信号。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200501-Go-Asynchronous-Preemption/04.png)

之后，当信息被信号处理器接收到时，线程中断当前的操作来处理信号，因此不会再运行当前的协程，在我们的例子中是 `G7`。取而代之的是，`gsignal` 被调度为管理发送来的信号。当它发现它是一个抢占指令后，在程序处理信号后恢复时它准备好指令来中止当前的协程。下面是这第二个阶段的示意图：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200501-Go-Asynchronous-Preemption/05.png)

*如果你想了解更多关于 `gsignal` 的信息，我推荐你读一下我的文章[”Go：gsignal，信号的掌控者“](https://medium.com/a-journey-with-go/go-gsignal-master-of-signals-329f7ff39391)。*

## 实现

我们在被选中的信号 `SIGURG` 中第一次看到了实现的细节。这个选择在提案[”提案：非合作式协程抢占“](https://github.com/golang/proposal/blob/master/design/24543-non-cooperative-preemption.md)中有详细的解释：

> - 它应该是调试者默认传递过来的一个信号。
> - 它不应该是 Go/C 混合二进制中 libc 内部使用的信号。
> - 它应该是一个可以伪造而没有其他后果的信号。
> - 我们需要在没有实时信号时与平台打交道。

然后，当信号被注入和接收时，Go 需要一种在程序恢复时能终止当前协程的方式。为了实现这个过程，Go 会把一条指令推进程序计数器，这样看起来运行中的程序调用了运行时的函数。该函数暂停了协程并把它交给了调度器，调度器之后还会运行其他的协程。

*我们应该注意到 Go 不能做到在任何地方终止程序；当前的指令必须是一个安全点。例如，如果程序现在正在调用运行时，那么抢占协程并不安全，因为运行时很多函数不应该被抢占。*

这个新的抢占机制也让垃圾回收器受益，可以用更高效的方式终止所有的协程。诚然，STW 现在非常容易，Go 仅需要向所有运行的线程发出一个信号就可以了。下面是垃圾回收器运行时的一个例子：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200501-Go-Asynchronous-Preemption/06.png)

然后，所有的线程都接收到这个信号，在垃圾回收器重新开启全局之前会暂停执行。

*如果你想了解更多关于 STW 的信息，我建议你阅读我的文章[”Go：Go 怎样实现 STW？“](https://medium.com/a-journey-with-go/go-how-does-go-stop-the-world-1ffab8bc8846)。*

最后，这个特性被封装在一个参数中，你可以用这个参数关闭异步抢占。你可以用 `GODEBUG=asyncpreemptoff=1` 来运行你的程序，如果你因为升级到了 Go 1.14 发现了不正常的现象就可以调试你的程序，或者观察你的程序有无异步抢占时的不同表现。

---

via: https://medium.com/a-journey-with-go/go-asynchronous-preemption-b5194227371c

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
