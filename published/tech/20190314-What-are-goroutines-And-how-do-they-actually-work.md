首发于：https://studygolang.com/articles/20021

# 什么是协程（goroutine），它们是怎样工作的呢？
![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/What-are-goroutines-And-how-do-they-actually-work/pic1.jpeg)
在过去的几年里，Go 语言能获得如此难以置信的流行度的一个主要原因，是 Go 能通过轻量级的 Goroutines 和 channel 更加简便地处理并发问题。

并发并不是什么新鲜事物，它一直以多线程的形式存在于我们几乎每天都要使用的应用程序之中。

不过，在实际理解什么是 Goroutine，什么不是之前，我们需要知道 Goroutine 并不是轻量级线程（尽管 Goroutine 依赖于线程运行），稍后我们将深入研究真正的线程在 OS 中是怎样工作的。

## 什么是线程

线程是可以在 OS 中执行的最小处理单元。在大多数现代操作系统中，线程存在于某个进程中 —— 也就是说，单个进程可能包含多个线程。

一个很好的例子就是 Web 服务器。

Web 服务器通常被设计来同时处理多个请求。这些请求一般彼此独立。

当请求到达时，web 服务器会创建一个线程，或者从线程池中获取一个线程，然后将请求来委派给线程来实现并发。不过请记住 RobPike 的名言——『并发不是并行』。

## 然而线程比进程更加轻量吗，让我们一起来看看

线程是否比进程更加轻量取决于你从哪个角度看

理论上，线程之间共享内存，创建新线程的时候不需要创建真正的虚拟内存空间，也不需要 MMU（内存管理单元）上下文切换。**此外，线程间通信比进程之间通信更加简单，主要是因为线程之间有共享内存，而进程通信往往需要利用各种模式的 IPC（进程间通信），如信号量，消息队列，管道等**。

那么线程总是比进程更高效吗？在我们生活的这个多处理器的世界中并不是的。

例如 Linux 就是不区分线程和进程的，两者在 Linux 都被称作任务（task）。每个任务在 cloned 的时候都有一个介于最小到最大之间的共享级别。

当你调用 fork() 创建任务时，创建的是一个没有共享文件描述符，PID 和内存空间的新任务。而调用 pthread_create() 创建任务时，创建的任务将包含上述所有共享资源。

此外，线程之间保持[共享内存](https://users.cs.cf.ac.uk/Dave.Marshall/C/node27.html) 与多核的[L1 缓存](https://www.quora.com/What-is-the-L1-L2-and-L3-cache-of-a-microprocessor-and-how-does-it-affect-the-performance-of-it-For-example-I-have-a-laptop-with-an-Intel-4700MQ-microprocessor-with-a-6MB-L3-cache-What-does-this-value-indicate) 中的数据同步，与在隔离内存中运行不同的进程相比，需要付出更加大的代价。

Linux 研发人员经过很多努力已经成功最小化了任务切换的代价。虽然创建新任务还是比创建新线程需要更大的开销，不过任务切换不是。

## 那么，哪里可以改进线程

线程变慢主要有三个原因：

1，	线程自身有一个很大的堆（≥ 1MB）占用了大量内存。因此，想象一下创建 1000 个线程意味着你已经需要 1GB 的内存。**这实在是太多了！**

2，	线程需要重复存储许多寄存器，其中一些包括 AVX（高级向量扩展），SSE（流式 SIMD 外设），浮点寄存器，程序计数器（PC），堆栈指针（SP），这会降低应用程序性能。

3，	线程创建和消除需要调用操作系统以获取资源（例如内存），而这一操作相对是比较慢的。**不好！**

## Goroutines 怎么样

Goroutines 是在 Golang 中执行并发任务的方式。它们仅存在于 Go 运行时的虚拟空间中而不存在于 OS 中，因此需要 Go 调度器来管理它们的生命周期。请记住这一点很重要，对于所有操作系统看到的都只有一个请求并运行多个线程的单个用户级进程。goroutine 本身由 GoRuntimeScheduler 管理。
![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/What-are-goroutines-And-how-do-they-actually-work/pic2.png)

Go Runtime 为此目的维护三个 C 结构：
[(https://golang.org/src/runtime/runtime2.go)](https://golang.org/src/runtime/runtime2.go)

1. **G 结构**：表示单个 Goroutine，包含跟踪其堆栈和当前状态所需的对象。 还包含自己负责的代码的引用。

2. **M 结构**：表示 OS 线程。包含一些对象指针，例如全局可执行的 Goroutines 队列，当前运行的 Goroutine，它自己的缓存以及对 Go 调度器的引用。

3. **Sched 结构**：它是一个单一的全局对象，用于跟踪 Goroutine 和 M 的不同队列以及调度程序运行时需要的其他一些信息，例如单一全局互斥锁（Global Sched Lock）。

G 结构主要存在于两种队列之中，一个是 M （线程）可以找到任务的可执行队列，另外一个是一个空闲的 Goroutine 列表。调度程序维护的 M（执行线程）只能每次关联其中一个队列。为了维护这两种队列并进行切换，就必须维持单一全局互斥锁（Global Sched Lock）。

因此，在启动时，go 运行空间会为 GC，调度程序和用户代码启动许多 Goroutine。 并创建 OS 线程来处理这些 Goroutine。 不过创建的线程数量最多可以等于 **GOMAXPROCS**（默认为 1，但为了获得最佳性能，通常设置为计算机上的处理器数量）。

## 接下来就是重点了

为了使运行时的堆栈更小，go 在运行期间使用了大小可调整的有限堆栈，并且初始大小只有 2KB/goroutine。新的 Goroutine 通常会分配几 kb 的空间，这几乎总是足够的。如果不够的话，运行空间还能自动增长（或者缩小）内存来实现堆栈的管理，从而让大部分 Goroutine 存在于适量的内存中。每个函数调用的平均 CPU 开销大概是三个简单指令。因此在同一地址空间中创建数十万个 Goroutine 是切实可行的。**但是如果 Goroutine 是线程的话，系统资源将很快被消耗完**。
![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/What-are-goroutines-And-how-do-they-actually-work/pic3.png)

## 阻塞？这不是问题

当 Goroutine 进行阻塞调用时，例如通过调用阻塞的系统调用，这时调用的线程必须阻塞，go 的运行空间会操作自动将同一操作系统线程上的其他 Goroutine，将它们移动到从调度程序（Sched Struct）的线程队列中取出的另一个可运行的线程上，所以这些 Goroutine 不会被阻塞。**因此，运行空间应至少创建一个线程，以继续执行不在阻塞调用中的其他 Goroutine。** 而且关键的是程序员是看不到这一点的。结论是，我们称之为 Goroutines 的事物，可以是很低廉的：它们在堆栈的内存之外几乎没有开销，而内存中也只有几千字节。

**并且，Go 协程也可以很好地扩展。**

**但是，如果你使用只存在于 Go 的虚拟空间的 channels 进行通信（产生阻塞时），操作系统将不会阻塞该线程。** 只是让该 Goroutine 进入等待状态，并安排另一个可运行的 Goroutine（来自 M 结构关联的可执行队列）它的位置。

## Go Runtime Scheduler

Go Runtime Scheduler 跟踪记录每个 Goroutine，并安排它们依次地在进程的线程池中运行。

Go Runtime Scheduler 执行协作调度，这意味着只有在当前 Goroutine 阻塞或完成时才会调度另一个 Goroutine，这通过代码可以轻松完成。这里有些例子：

* 调用系统调用如文件或网络操作阻塞时

* 因为垃圾收集被停止后

这样比定时阻塞并调度新线程的抢占式调度要好得多，因为当线程数量增加，或者当高优先级任务将被调度运行时，有低优先级的任务已经在运行了（此时低优先级队列将被阻塞），定时抢占调度可能导致某些任务完成花费的时间大大超过实际所需时间。

另一个优点是，因为 Goroutine 在代码中隐式调用的，例如在睡眠或 channel 等待期间，编译只需要安全地恢复在这些时刻处存活的寄存器。在 Go 中，这意味着在上下文切换期间**仅更新 3 个寄存器，即 PC，SP 和 DX（数据寄存器）** 而不是所有寄存器（例如 AVX，浮点，MMX）。

如果您想了解有关 Go 并发的更多信息，请参阅以下链接：

[Concurrency is not parallelism by Rob Pike (**Must watch for any Go Developer**)](https://www.youtube.com/watch?v=cN_DpYBzKso&t=441s)

[Analysis of Go runtime Scheduler](http://www1.cs.columbia.edu/~aho/cs6998/reports/12-12-11_DeshpandeSponslerWeiss_GO.pdf)

喜欢这个帖子？**点个赞吧！**

---

via: https://medium.com/@joaoh82/what-are-goroutines-and-how-do-they-actually-work-f2a734f6f991

作者：[João Henrique Machado](https://medium.com/@joaoh82)
译者：[HN-JIE](https://github.com/HN-JIE)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
