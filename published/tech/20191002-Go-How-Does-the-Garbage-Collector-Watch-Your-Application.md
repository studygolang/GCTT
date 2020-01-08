首发于：https://studygolang.com/articles/25915

# Go: GC 是怎样监听你的应用的？

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-How-Does-the-Garbage-Collector-Watch-Your-Application/1.png)

<p align="center">Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.</p>

> 这篇文章是基于 Go 的 *1.13* 版本

Go 语言的垃圾收集器 （下文简称 GC ）能够帮助到开发者，通过自动地释放掉一些程序中不再需要使用的内存。但是，跟踪并清理掉这些内存也可能影响我们程序的性能。Go 语言的 GC 旨在实现 [这些目标](https://blog.golang.org/ismmkeynote) 并且关注如下几个问题：

- 当程序被终止时，尽可能多的减少在这两个阶段的 STW （的次数） 。
- 一个 GC 周期的时间要少于 10 毫秒。
- 一次 GC 周期不能占用超过 25% 的 CPU 资源。

这是一些很有挑战性的目标，如果 GC 从我们的程序中了解到足够多的信息，它就能去解决这些问题。

## 到达堆阈值

GC 将会关注的第一个指标是堆的使用增长。默认情况下，它将在堆大小加倍时运行。这是一个在循环中分配内存的简单程序。

```go
func BenchmarkAllocationEveryMs(b *testing.B) {
	// need permanent allocation to clear see when the heap double its size
	var s *[]int
	tmp := make([]int, 1100000, 1100000)
	s = &tmp

	var a *[]int
	for i := 0; i < b.N; i++  {
		tmp := make([]int, 10000, 10000)
		a = &tmp

		time.Sleep(time.Millisecond)
	}
	_ = a
	runtime.KeepAlive(s)
}
```

追踪器向我们展示了 GC 什么时候被调用：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-How-Does-the-Garbage-Collector-Watch-Your-Application/2.png)
<p align="center">Garbage collector cycles and heap size</p>

一旦堆的大小增加了一倍，内存分配器就会触发执行 GC 。通过设置 `GODEBUG=gctrace=1` ，来打印出若干循环的信息就能够证实这一点：

```
gc 8 @0.251s 0%: 0.004+0.11+0.003 ms clock, 0.036+0/0.10/0.15+0.028 ms cpu, 16->16->8 MB, 17 MB goal, 8 P

gc 9 @0.389s 0%: 0.005+0.11+0.007 ms clock, 0.041+0/0.090/0.11+0.062 ms cpu, 16->16->8 MB, 17 MB goal, 8 P

gc 10 @0.526s 0%: 0.046+0.24+0.014 ms clock, 0.37+0/0.14/0.23+0.11 ms cpu, 16->16->8 MB, 17 MB goal, 8 P
```

第九个循环就是我们之前看到的那个循环，运行在第 389 ms 。有意思的部分是 `16->16->8 MB` ，它展示了在 GC 被调用前堆使用的内存有多大，以及在 GC 执行后它们还剩下多少。我们可以清楚地看到，当第八个循环将堆大小减少到 8 MB 时，第九个 GC 周期将在 16 MB 时刻触发。

这个阈值的比例由环境变量 GOGC 决定，默认值为 100 % —— 也就是说，在堆的大小增加了一倍之后，GC 就会被调用。出于性能原因，并且为了避免经常启动一个循环，当堆的大小低于 4 MB * GOGC 的时候， GC 将不会被执行。——当 GOGC 被设置为 100 % 时，在堆内存低于 4 MB 时 GC 将不会被触发。

## 到达时间阈值

GC 关注的第二个指标是在两次 GC 之间的时间间隔。如果超过两分钟 GC 还未执行，那么就会强制启动一次 GC 循环。

由 `GODEBUG` 给出的跟踪显示，两分钟后会强制启动一次循环。

```
GC forced
gc 15 @121.340s 0%: 0.058+1.2+0.015 ms clock, 0.46+0/2.0/4.1+0.12 ms cpu, 1->1->1 MB, 4 MB goal, 8 P
```

## 需要协助

GC 主要由两个主要阶段组成：

- 标记仍在使用的内存
- 清理未标记为使用中的内存

在标记期间，Go 必须确保 GC 标记内存的速度比新分配内存的速度更快。事实是，如果 GC 正在标记 4 MB 大小的内存，然而同时程序正在分配同样大小的内存，那么 GC 必须在完成后立即触发。

为了解决这个问题，Go 在标记内存的同时会跟踪新的内存分配，并关注 GC 何时过载。第一步在 GC 触发时执行，它会首先为每一个处理器准备一个 处于 sleep 状态的 goroutine，等待标记阶段。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-How-Does-the-Garbage-Collector-Watch-Your-Application/3.png)

<p align="center">Goroutines for marking phase</p>

跟踪器能够显示出这些 goroutines：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-How-Does-the-Garbage-Collector-Watch-Your-Application/4.png)

<p align="center">Goroutines for marking phase</p>

当这些 Goroutine 生成后， GC 就开始标记阶段，该阶段会检查哪些变量应收集并清除。被标记为 `GC dedicated` 的 goroutines 会运行标记，并不会被抢占，然而那些标记为 `GC idle` 的 goroutines 就会去工作，因为他们没有任何其他事情。可以被抢占。

GC 现在已经能够去标记那些不再使用的变量。对于每一个被扫描到的变量，它会增加一个计数器，以便继续跟踪当前的工作并且也能够获得剩余工作的快照。当一个 Goroutine 在 GC 期间被安排了任务，Go 将会比较所需要的分配和已经扫描到的，以便对比扫描的速度和分配的需求。如果比较的结果是扫描内容较多，那么当前的 Goroutine 并不需要去提供帮助。换句话说，如果扫描与分配相比有所欠缺，那么 Go 就会使用 goroutine 来协助。这有一个图表来反应这个逻辑：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-How-Does-the-Garbage-Collector-Watch-Your-Application/5.png)
<p align="center">Mark assist based on scanning debt</p>

在我们的示例中，因为扫描 / 分配的差值为负数，所以 Goroutine 14 被请求工作：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-How-Does-the-Garbage-Collector-Watch-Your-Application/6.png)
<p align="center">Assistance for marking</p>

## CPU 限制

Go 语言 GC 的目标之一是不占用 25 % 的 CPU。这就意味着 Go 在标记阶段不应分配超过四分之一的处理器。实际上，这正是我们在前面的示例中所看到的，在八个处理器中只有两个 Goroutine 被 GC 充分利用：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-How-Does-the-Garbage-Collector-Watch-Your-Application/7.png)

<p align="center">Dedicated Goroutine for marking phase</p>

正如我们所看到的，其他的 Goroutine 仅在没有其他事情要做的情况下才会在标记阶段工作。但是，在 GC 的协助请求下，Go 程序可能会在高峰时间最终占用超过 25 % 的 CPU ，如 Goroutine 14 所示：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-How-Does-the-Garbage-Collector-Watch-Your-Application/8.png)

<p align="center">Mark assistance with dedicated goroutines</p>

在我们的示例中，短时间内将我们处理器的 37.5 % （八分之三）分配给了标记阶段。这可能很少见，只有在分配很高的情况下才会发生。

---

via：https://medium.com/a-journey-with-go/go-how-does-the-garbage-collector-watch-your-application-dbef99be2c35

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[sh1luo](https://github.com/sh1luo)
校对：[lxbwolf](https://github.com/lxbwolf)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
