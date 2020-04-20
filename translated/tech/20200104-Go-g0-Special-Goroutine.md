# Go：g0，特殊的 Goroutine
ℹ️ 这篇文章基于 Go 1.13。

在 Go 中创建的所有 goroutine 都会被一个内部的调度器所管理。Go 调度器尝试为所有的 goroutine 分配运行时间，并且在当前的 goroutine 阻塞或者终止的时候，Go 调度器会通过运行 goroutine 的方式使所有 CPU 保持忙碌状态。这个调度器实际上是作为一个特殊的 goroutine 运行的。

## 调度 goroutine
Go 使用 `GOMAXPROCS` 变量限制同时运行的 OS 线程数量，这意味着 Go 必须对每个运行着的线程上的 goroutine 进行调度和管理。这个调度的功能被委托给了一个叫做 `g0` 的特殊的 goroutine， g0 是为每个 OS 线程创建的第一个 goroutine：

![](https://github.com/studygolang/gctt-images2/blob/master/20200104-Go-g0-Special-Goroutine/g0-created-for-each-OS-thread.png?raw=true)

之后，g0 会把就绪状态的 goroutine 调度到线程上去运行。

我建议你阅读我的文章“ [Go：Goroutine，OS 线程和 CPU 管理](https://medium.com/a-journey-with-go/go-goroutine-os-thread-and-cpu-management-2f5a5eaf518a)”，来了解更多关于 `P`，`M`，`G` 模型的信息。

为了更好的理解 `g0` 的调度策略，来回顾下 channel 的用法。下面是一个 goroutine 在向 channel 发送数据是阻塞的例子：

```go
ch := make(chan int)
[...]
ch <- v
```

当在 channel 上阻塞时，当前的 goroutine 会被停放（ parked ），即处于等待状态（ waiting mode ），并且不会被放在任何 goroutine 队列中：

![](https://github.com/studygolang/gctt-images2/blob/master/20200104-Go-g0-Special-Goroutine/blocking-goroutine-will-beparked.png?raw=true)

之后，`g0` 会替换 goroutine 并进行一轮调度：

![](https://github.com/studygolang/gctt-images2/blob/master/20200104-Go-g0-Special-Goroutine/g0-replaces-the-goroutine.png?raw=true)

本地队列在调度过程中具有优先级，2号 goroutine 会被运行：

![](https://github.com/studygolang/gctt-images2/blob/master/20200104-Go-g0-Special-Goroutine/local-queue-scheduling.png?raw=true)

我建议你阅读我的文章“ [Go: Go 调度器中的工作窃取（Work-Stealing）](https://medium.com/a-journey-with-go/go-work-stealing-in-go-scheduler-d439231be64d)” 来了解更多关于调度优先级的细节。

一旦有接收者读取 channel 中的数据，7号 goroutine 就会解除阻塞状态：

```go
v := <-ch
```

收到消息的 goroutine 会切换到 `g0`，并且通过放入本地队列的方式将该 goroutine 从停放状态解锁：

![](https://github.com/studygolang/gctt-images2/blob/master/20200104-Go-g0-Special-Goroutine/unlock-the-parked-goroutine.png?raw=true)

虽然这个特殊的 goroutine 管理调度策略，但这并不是它唯一的工作，它还负责着更多的工作。

## 职责
与普通 goroutine 不同的是，`g0` 有着固定且更大的栈，这使得在需要更大的栈的时候，以及栈不宜增长的时候，Go 可以进行操作。在 `g0` 的职责中，我们可以列出：

- goroutine 创建。当调用 `go func(){ ... }()` 或 `go myFunction()` 时，Go 会在把它们放入本地队列前，将函数的创建委托给 `g0` 去做：

![](https://github.com/studygolang/gctt-images2/blob/master/20200104-Go-g0-Special-Goroutine/goroutine-creation.png?raw=true)

新创建的 goroutine 优先运行，并且被放在本地队列的顶部。

建议阅读我的文章“[Go：并发与调度器亲和性（ Go: Concurrency & Scheduler Affinity ）](https://medium.com/a-journey-with-go/go-concurrency-scheduler-affinity-3b678f490488)”了解更多关于 goroutine 优先级的信息。

- defer 函数分配。
- 垃圾收集操作，比如 STW（ stopping the world ），扫描 goroutine 的栈，以及一些标记清理操作。
- 栈增长。当需要的时候，Go 会增加 goroutine 的大小。这个操作是由 `g0` 的 prolog 函数完成的。

这个特殊的 goroutine 涉及许多其他操作（较大空间的对象分配，cgo 等），需要较大的栈来保证我们的程序进行更高效的管理操作，以保持程序的低内存打印效率。

---
via: https://medium.com/a-journey-with-go/go-g0-special-goroutine-8c778c6704d8

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[dust347](https://github.com/dust347)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
