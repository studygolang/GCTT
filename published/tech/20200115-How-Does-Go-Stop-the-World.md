首发于：https://studygolang.com/articles/28450

# Go 语言如何实现垃圾回收中的 Stop the World (STW)

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://raw.githubusercontent.com/studygolang/gctt-images/master/how-does-go-stop-the-world/cover.png)

## 目录

- [Stop The World(STW)](#stop-the-worldstw)
- [系统调用](#%e7%b3%bb%e7%bb%9f%e8%b0%83%e7%94%a8)
- [延迟](#%e5%bb%b6%e8%bf%9f)

*本篇文章讨论实现原理基于 Go 1.13.*

在垃圾回收机制 (GC) 中，"Stop the World" (STW) 是一个重要阶段。 顾名思义， 在 "Stop the World" 阶段， 当前运行的所有程序将被暂停， 扫描内存的 root 节点和添加写屏障 (write barrier) 。 本篇文章讨论的是， "Stop the World" 内部工作原理及我们可能会遇到的潜在风险。

## Stop The World(STW)

这里面的"停止"， 指的是停止正在运行的 goroutines。 下面这段程序就执行了 "Stop the World"：

```go
func main() {
   runtime.GC()
}
```

这段代码中， 调用 `runtime.GC()` 执行垃圾回收， 会触发 "Stop the World"的三个步骤。

(关于关于垃圾回收机制， 可以参考我的另外一篇文章 ["Go： 内存标记在垃圾回收中的实现"](https://medium.com/a-journey-with-go/go-how-does-the-garbage-collector-mark-the-memory-72cfc12c6976))：

这个阶段的第一步， 是抢占所有正在运行的 goroutine(即图中 `G`)：

![STW_goroutines_preemption](https://raw.githubusercontent.com/studygolang/gctt-images/master/how-does-go-stop-the-world/STW_goroutines_preemption.png)

被抢占之后， 这些 goroutine 会被悬停在一个相对安全的状态。 同时，承载 Goroutine 的处理器 `P` (无论是正在运行代码的处理器还是已在 idle 列表中的处理器)， 都会被被标记成停止状态 (stopped)， 不再运行任何代码：

![STW_P_stopped](https://raw.githubusercontent.com/studygolang/gctt-images/master/how-does-go-stop-the-world/STW_P_stopped.png)

接下来， Go 调度器 (Scheduler) 开始调度， 把每个处理器的 Marking Worker (即图中 `M`) 从各自对应的处理器 `P` 分离出来， 放到 idle 列表中去， 如下图：

![STW_M_Detach](https://raw.githubusercontent.com/studygolang/gctt-images/master/how-does-go-stop-the-world/STW_M_Detach.png)

在停止了处理器和 Marking Worker 之后， 对于 Goroutine 本身， 他们会被放到一个全局队列中等待：

![STW_G_Queue](https://raw.githubusercontent.com/studygolang/gctt-images/master/how-does-go-stop-the-world/STW_G_Queue.png)

到目前为止， 整个"世界"被停止. 至此， 仅存的 "Stop The World" (STW)goroutine 可以开始接下来的回收工作， 在一些列的操作结束之后， 再启动整个"世界"。

我们也可以在 Tracing 工具中看到一次 STW 的运行状态：

![STW_TRACING](https://raw.githubusercontent.com/studygolang/gctt-images/master/how-does-go-stop-the-world/STW_TRACING.png)

## 系统调用

下面我们来讨论一下 STW 是如何处理系统调用的。

我们知道， 系统调用是需要返回的， 那么当整个"世界"被停止的时候， 已经存在的系统调用如何被处理呢？

我们通过一个实际例子来理解：

```go
func main() {
   var wg sync.WaitGroup
   wg.Add(10)
   for i ：= 0; i < 10; i++ {
      Go func() {
         http.Get(`https：//httpstat.us/200`)
         wg.Done()
      }()
   }
   wg.Wait()
}
```

这是一段简单的系统调用的程序， 我们通过 Tracing 工具看一下它是如何被处理的：

![SC_tracing](https://raw.githubusercontent.com/studygolang/gctt-images/master/how-does-go-stop-the-world/SC_tracing.png)

我们可以看到， 这个系统调用 goroutine (即图中 `G30`) 在"世界"被停止的时候， 就已经存在了。

但是， 我们之前提到， STW 把所有的处理器 `P` 都标为停止状态 (stopped) ， 所以， 这个系统调用的 Goroutine 也会被放到全局队列中， 等待 golang 世界恢复之后， 被重新启用。

## 延迟

前文提到 STW 的第三步是将 Marking Worker(`M`) 从处理器(`P`)上分离， 然后放入 idle 列表中。

而实际上， Go 会等待他们自发停止， 也就是说当调度器(scheduler)运行的时候， 系统调用在运行的时候， STW 会等待。

理论上， 等待一个 Goroutine 被抢占是很快的， 但是在有些情况下， 还是会出现相应的延迟。

我们通过一个例子来模拟类似情况：

```go
func main() {
   var t int
   for i ：= 0;i < 20 ;i++  {
      Go func() {
         for i ：= 0;i < 1000000000 ;i++ {
            t++
         }
      }()
   }

   runtime.GC()
}
```

我们还是来看一下这段代码运行的 Tracing 情况， 从下图我们可以看到 STW 阶段总共耗时 2.6 秒：

![STW_26S](https://raw.githubusercontent.com/studygolang/gctt-images/master/how-does-go-stop-the-world/STW_26S.png)

我们来简单分析为什么会出现这么长的 STW： 正如例子中的 main 函数， 一个没有函数调用的 Goroutine 一般不会被抢占， 那么这个 Goroutine 对应的处理器 `P` 在任务结束之前不会被释放。

而 STW 的机制是等待它自发停止， 因此就出现了 2.6 秒的 STW。

为了提高整体程序的效率， 我们一般需要避免或者改进这种情况。

关于这部分， 大家可以参考我的另一篇文章 ["Go: Goroutine 和抢占"](https://medium.com/a-journey-with-go/go-goroutine-and-preemption-d6bc2aa2f4b7)

---

via: https://medium.com/a-journey-with-go/go-how-does-go-stop-the-world-1ffab8bc8846

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[SarahChenBJ](https://github.com/SarahChenBJ)
校对：[@unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
