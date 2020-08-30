# Go: 定时器的生命周期

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200704-go-timers-life-cycle/图0.png)

> 本篇文章基于Go `1.14`

`定时器`对于在将来的某个时刻执行代码时非常有用。Go内部在管理创建的定时器的同时，也会对其执行进行规划。后者可能有点棘手，因为Go调度器是一个协作式(`cooperative`)调度器，这意味着一个goroutine必须自己停止（阻塞在`channel`上，系统调用， 等等）或由调度器在某个调度点暂停。

> 如果想要获取更多关于优先权的信息，我建议你阅读我的文章：[Go: Goroutine and Preemption](https://medium.com/a-journey-with-go/go-goroutine-and-preemption-d6bc2aa2f4b7)。

## 生命周期

下面是一个关于`定时器`的最简单的示例：

```golang
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	time.AfterFunc(time.Second, func() {
		println("done")
	})

	<-sigs
}

```

当一个定时器被创建时，它会被保存在与当前P关联的定时器的内部列表中，上面的代码可以用下图来表示：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200704-go-timers-life-cycle/图1.png)

> 如果想要获取更多关于GMP模型的内容，建议您可以参考一理我的这篇文章: [Go: Goroutine, OS Thread and CPU Management](https://medium.com/a-journey-with-go/go-goroutine-os-thread-and-cpu-management-2f5a5eaf518a)。

如图所示，一旦定时器被创建，它就会注册一个内部回调，该回调将用关键字`go`调用用户回调，将其转换为goroutine。

然后，定时器将由调度器进行管理。在每一轮调度中，它都会检查定时器是否准备好运行，如果准备好了，就准备运行。事实上，由于Go调度器本身并不运行任何代码，运行定时器的回调会将其goroutine加到本地队列中。然后，当调度器在队列中选中它时，goroutine就会运行。如下图所示：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200704-go-timers-life-cycle/图2.png)

根据本地队列的大小，定时器的运行可能会有一些小的延迟。事实上，由于Go 1.14中的`异步抢占`，goroutine在运行`10ms`后就会被抢占，减少了延迟的概率。

## 时延

为了理解`时延`的可能性，我们来分析一个**从同一个goroutine中创建大量定时器的情况**。由于定时器是与当前P相连的，所以一个被占用的P将无法运行其定时器。这里有一个程序，它创建了数百个定时器，并在其余时间内保持忙碌状态：

```golang
package main

import (
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var num int64 = 0
	for i := 0; i < 1e3; i++ {
		time.AfterFunc(time.Second, func() {
			atomic.AddInt64(&num, 1)
		})
	}

	// 耗时超过1s
	t := 0
	for i := 0; i < 1e10; i++ {
		t++
	}

	_ = t

	<-sigs

	println(num, "timers created,", t, "iterations done")
}

```

通过下图的`tracing`，我们可以清楚的看到goroutine占用处理器的情况：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200704-go-timers-life-cycle/图3.png)

图中的每一个区块表示，由于异步抢占，运行中的goroutine的被分成了大量的区块。

> 更多关于异步抢占的内容，请参考我的这篇文章: [Go: Asynchronous Preemption](https://medium.com/a-journey-with-go/go-asynchronous-preemption-b5194227371c)

在这些块中，有一个空间看起来比其他的大。让我们把它放大看一下：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200704-go-timers-life-cycle/图4.png)

这个间隔发生在定时器必须运行的时候。此时，当前的goroutine已经被抢占，并被Go调度器所取代。正如图中高亮部分所示， 调度器将定时器转换为可执行的goroutine。

然而，当前线程的Go调度器并不是唯一一个运行定时器的调度器。Go实现了一个定时器`窃取策略`，以确保当前线程相当繁忙时，定时器可以由另一个`P`运行。由于异步抢占，这种情况不太可能发生，但在我们的例子中，由于定时器的数量非常多，这种情况还是发生了。如图所示：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200704-go-timers-life-cycle/图5.png)

如果我们不考虑定时器`窃取策略`，下图展示了将会发生的事情:

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200704-go-timers-life-cycle/图6.png)

所有持有定时器的goroutine都被添加到本地队列中。然后，基于`P`之间的`work-stealing`策略对其重新进行调度分发。

> 更多关于`work-stealing`相关资料，请参考我的文章：关于Go中工作偷窃的更多信息，我建议你阅读我的文章：[Go: Work-Stealing in Go Scheduler](https://medium.com/a-journey-with-go/go-work-stealing-in-go-scheduler-d439231be64d)

综上所述，由于异步抢占和`work-stealing`机制，导致延迟发生的可能性很小。

---
via: https://medium.com/a-journey-with-go/go-timers-life-cycle-403f3580093a

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[double12gzh](https://github.com/double12gzh)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
