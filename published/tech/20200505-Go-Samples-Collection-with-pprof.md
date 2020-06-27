首发于：https://studygolang.com/articles/28982

# Go：使用 pprof 收集样品数据

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200505-Go-Samples-Collection-with-pprof/00.png)

> Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.

ℹ️ *本文基于 Go 1.13。*

`pprof` 是用于分析诸如 CPU 或 内存分配等 profile 数据的工具。分析程序的 profile 数据需要收集运行时的数据用来在之后统计和生成画像。我们现在来研究下数据收集的工作流以及怎么样去调整它。

## 工作流

`pprof` 以一个固定的时间间隔基础来收集数据，这个时间间隔是以每秒的收集器的个数定义的。默认参数是 `100`，即 `pprof` 每秒收集 100 次数据，例如，每 10 毫秒收集一次。

可以通过调用 `StartCPUProfile` 来启动 `pprof`。

```go
func main() {
	f, _ := os.Create(`cpu.prof`)
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	...
}
```

这个过程会在运行的线程中自动设置一个定时器（下图中 `M` 表示线程），让 Go 定期地收集 profile 数据。下面是第一个示意图：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200505-Go-Samples-Collection-with-pprof/01.png)

*想了解更多关于 MPG 调度模型的信息，我推荐你阅读我的文章”[Go：协程，操作系统线程和 CPU 管理](https://studygolang.com/articles/25292)。“*

然而，目前为止 `pprof` 仅在收集当前运行的线程的 profile 信息。当 Go 调度器想调度一个协程运行在某个线程上时，这个线程也可以实时被追踪。下面是更新后的示意图：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200505-Go-Samples-Collection-with-pprof/02.png)

*想了解更多关于 Go 调度器的信息，我建议你阅读我的文章”[Go: g0，特殊的协程](https://medium.com/a-journey-with-go/go-g0-special-goroutine-8c778c6704d8)“。*

之后，profile 数据会在定义好的每个时间间隔到期后定期地被 dump 到一个缓冲区：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200505-Go-Samples-Collection-with-pprof/03.png)

数据实际上是由 `gsignal` 进行 dump 的，这个协程是用来处理发来的信号的。实际上，在每个时间间隔到期后定时器会发送信号。

*想了解更多关于信号和 `gsignal` 的信息，我推荐你阅读我的文章”[Go：gsignal，信号的掌控者](https://studygolang.com/articles/28974)“。*

## 基于信号的机制

在每个线程上创建的定时器是由 `settimer` 方法和时间间隔定时器 `ITIMER_PROF` 管理的。时间计数仅在系统代表这个处理运行时才会减少，这样就能确保 profile 数据的准确。

当经过了定义的时间间隔后，定时器发送一个 `SIGPROF` 信号，这个信号会被 Go 截获，profile 数据会被 dump 到缓冲区。频率可以通过调用 `runtime` 包里的 `SetCPUProfileRate` 函数进行配置。这个配置操作需要在启动 profiler 之前完成：

```go
func main() {
	f, err := os.Create(`cpu.prof`)
	if err != nil {
		log.Fatal(err)
	}

	runtime.SetCPUProfileRate(10)
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	...
}
```

profile 数据采集率只能定义一次，在启动 profiler 时 `pprof` 定义它。在启动 profiler 之前调用该方法会导致 `pprof` 忽略默认值。

然而，默认值能满足大部分场景中。包的文档中有对此的详细解释：

> 100 Hz 是合理的值：既能满足产出有用数据的频率需求，又不至于过快而使系统 hang 住。

当然，更高的频率值似乎也可以，因为它可能会导致一些 [`SIGPROF` 事件](https://github.com/golang/go/issues/35057)从 `250` 或更高的值降下来。`pprof` 文档也陈述了怎样让它表现得更好：

> *[…]* 实践中操作系统不能以比 500 Hz 更高的频率触发信号

## 收集数据

现在所有的定时器都已经设置完了。当数据被 dump 到缓冲区后，`pprof` 需要一种把所有数据收集起来生成报告的方法。这个处理过程是在独立的协程中进行的，每 100 毫秒收集和格式化一次数据。至于收集到的数据，Go 生成回溯信息，用来找到函数调用关系以及处理它来格式化内联调用。

当信息生成过程完成后，例如 profile 数据收集结束后，特定的协程会把报告 dump 到文件，这样数据就可用和完全可视化了。

---

via: https://medium.com/a-journey-with-go/go-samples-collection-with-pprof-2a63c3e8a142

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
