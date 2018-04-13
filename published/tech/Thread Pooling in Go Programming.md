已发布：https://studygolang.com/articles/12790

# Go 语言中的线程池（Thread Pooling in Go Programming）

用过一段时间的 Go 之后，我学会了如何使用一个不带缓存的 channel 去创建一个 goroutine 池。我喜欢这个实现，这个实现甚至比这篇博文描述的更好。虽然这样说，这篇博文仍然对它所描述的部分有一定的价值。

[https://github.com/goinggo/work](https://github.com/goinggo/work)

## 介绍（Introduction）

在我的服务器开发的职业生涯里，线程池一直是在微软系统的堆栈上构建健壮代码的关键。微软在 .Net 上的失败，是因为它给每个进程分配一个单独的线程池，并认为在它们并发运行时能够管理好。我早就已经意识到这是不可能的。至少，在我开发的服务器上不可行。

当我用 Win32 API，C/C++ 构建系统时，我创建了一个抽象的 IOCP 类，它可以给我分配好线程池，我把工作扔给它（去处理）。这样工作得非常好，并且我还能够指定线程池的数量和并发度（能够同时被执行的线程数）。在我使用 C# 开发的时间里，我沿用了这段代码。如果你想了解更多，我在几年前写了一篇文章 http://www.theukwebdesigncompany.com/articles/iocp-thread-pooling.php 。 使用 IOCP，给我带来了需要的性能和灵活性。 顺便说一下，.Net 线程池使用了下面的 IOCP。

线程池的想法非常简单。工作被发送到服务器，它们需要被处理。大多数工作本质上是异步的，但不一定是。大多数时候，工作来自于一个内部协程的通信。线程池将工作加入其中，然后这个池子中的一个线程会被分配来处理这个工作。工作按照接收的顺序被执行。线程池为有效地执行工作提供了一个很好的模式。（设想一下，）每次需要处理工作时，产生一个新线程会给操作系统带来沉重的负担，并导致严重的性能问题。

那么如何调整线程池的性能呢？你需要找出线程池包含多少个线程时，工作被处理得最快。当所有的线程都在忙着处理任务时，新的任务将待在队列里。这是你希望的，因为从某些方面来说，太多的线程（反而）会导致处理工作变得更慢。导致这个现象有几个原因，像机器上的 CPU 核数需要有能力去处理数据库请求（等）。经过测试，你可以找到最合适的数值。

我总是先找出（机器上的 CPU）有多少个核，以及要被处理的工作的类型。工作阻塞时，（它们）平均被阻塞多久。在微软系统的堆栈上，我发现对于大多数工作来说，每个核上运行 3 个线程能够获得最好的性能。Go 的话，我还不知道最佳的数字。

你也可以为不同类型的工作创建不同的线程池。因为每种线程池都可以被配置，你可以花点时间使服务器获得最大输出。通过这种方式的指挥和控制对于实现最大化服务器能力至关重要。

在 Go 语言中我不创建线程，而是创建协程。协程函数类似于多线程函数，但由 Go 来管理实际上在系统层面运行的线程。了解更多关于 Go 中的并发，查看这个文档：[http://golang.org/doc/effective_go.html#concurrency](http://golang.org/doc/effective_go.html#concurrency)。

我创建了名为 workpool 和 jobpool 的包。它们通过 channel 和 go 协程来实现池的功能。

## 工作池（Workpool）

这个包创建了一个 go 协程池，专门用来处理发布到池子中的工作。一个独立的 Go 协程负责工作的排队处理。协程提供安全的工作排队，跟踪队列中工作量，当队列满时报告错误。

提交工作到队列中是一个阻塞操作。这样调用者才能知道工作是否已经进入队列。（workpool 也会一直）保持工作队列中活动程序数量的计数。

这是如何使用 workpool 的样例代码：

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/goinggo/workpool"
)

type MyWork struct {
	Name string
	BirthYear int
	WP *workpool.WorkPool
}

func (mw *MyWork) DoWork(workRoutine int) {
	fmt.Printf("%s : %d\n", mw.Name, mw.BirthYear)
	fmt.Printf("Q:%d R:%d\n", mw.WP.QueuedWork(), mw.WP.ActiveRoutines())

	// Simulate some delay
	time.Sleep(100 * time.Millisecond)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	workPool := workpool.New(runtime.NumCPU(), 800)

	shutdown := false // Race Condition, Sorry

	go func() {
		for i := 0; i < 1000; i++ {
			work := MyWork {
				Name: "A" + strconv.Itoa(i),
				BirthYear: i,
				WP: workPool,
			}

			if err := workPool.PostWork("routine", &work); err != nil {
				fmt.Printf("ERROR: %s\n", err)
				time.Sleep(100 * time.Millisecond)
			}

			if shutdown == true {
				return
			}
		}
	}()

	fmt.Println("Hit any key to exit")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString(’\n’)

	shutdown = true

	fmt.Println("Shutting Down")
	workPool.Shutdown("routine")
}
```

看下 main 函数，我们创建了一个协程池，协程数量基于机器上的核数。这意味每个核都对应有一个协程。如果每个核都处于忙碌状态，（那么）你将无法做更多的事情。再（运行）一次，性能测试会检测出哪个数量是最合适的。第二个参数是队列的大小。在这种情况下，我让队列足够大（800），保证所有的请求都可以进来。

MyWork 类型定义了我需要执行的工作状态。我们需要成员函数 DoWork，因为它实现了 PostWork 调用的接口。要将任何任务传递给线程池，都必须实现这个方法。

DoWork 方法做了两件事。第一是，它显示对象的状态。第二，它实时报告队列中的数量和 Go 协程并发执行的数量。这些数值可以用来检查线程池的健康状态和做性能测试。

最后，一个 Go 协程专门循环地将工作传递给工作池。同时，工作池为队列中的每个对象执行 DoWork 方法。Go 协程最终会完成，工作池继续执行它的工作。在任何时候当我们介入时，程序将优雅地停止。

在这个范例程序中，PostWork 方法能够返回一个错误。这是因为 PostWork 方法将保证任务放在队列中或者失败。这个失败的唯一原因是队列已满。（所以）设置队列的长度是一个重要的考虑项。

## 作业池（Jobpool）

jobpool 包跟 workpool 包很相似，除了一个实现的细节。这个包包含两个队列，一个是普通的处理队列，另外一个是高优先级的处理队列。阻塞的高优先级队列总是比阻塞的普通队列先获得处理。

两种队列的使用导致 jobpool 比 workpool 更加复杂。如果你不需要高优先级的处理，那么使用 workpool 将更快，更有效。

这是如何使用 jobpool 的范例代码：

```go
package main

import (
	"fmt"
	"time"

	"github.com/goinggo/jobpool"
)

type WorkProvider1 struct {
	Name string
}

func (wp *WorkProvider1) RunJob(jobRoutine int) {
	fmt.Printf("Perform Job : Provider 1 : Started: %s\n", wp.Name)
	time.Sleep(2 * time.Second)
	fmt.Printf("Perform Job : Provider 1 : DONE: %s\n", wp.Name)
}

type WorkProvider2 struct {
	Name string
}

func (wp *WorkProvider2) RunJob(jobRoutine int) {
	fmt.Printf("Perform Job : Provider 2 : Started: %s\n", wp.Name)
	time.Sleep(5 * time.Second)
	fmt.Printf("Perform Job : Provider 2 : DONE: %s\n", wp.Name)
}

func main() {
	jobPool := jobpool.New(2, 1000)

	jobPool.QueueJob("main", &WorkProvider1{"Normal Priority : 1"}, false)

	fmt.Printf("*******> QW: %d AR: %d\n",
		jobPool.QueuedJobs(),
		jobPool.ActiveRoutines())

	time.Sleep(1 * time.Second)

	jobPool.QueueJob("main", &WorkProvider1{"Normal Priority : 2"}, false)
	jobPool.QueueJob("main", &WorkProvider1{"Normal Priority : 3"}, false)

	jobPool.QueueJob("main", &WorkProvider2{"High Priority : 4"}, true)
	fmt.Printf("*******> QW: %d AR: %d\n",
		jobPool.QueuedJobs(),
		jobPool.ActiveRoutines())

	time.Sleep(15 * time.Second)

	jobPool.Shutdown("main")
}
```

在这个范例代码中，我们创建了两个 worker 类型的结构体。可以将每个 worker 都视为系统中一个独立的作业。

在 main 函数中，我们创建了一个包含 2 个协程的作业池，支持 1000 个待处理的作业。首先我们创建了 3 个不同的 WorkProvider1 对象，并将她们传递给了队列，设置优先级标志位为 false。接下来我们创建一个 WorkProvider2 对象，并将它传递给队列，设置优先级标志位为 true。

因为作业池中有 2 个协程，先创建的两个作业将进入队列并被处理。一旦它们的任务完成，接下来的作业将从队列中检索。WorkProvider2 作业将会被执行，因为它被放在了高优先级队列中。

想获取 workpool 包和 jobpool 包的代码，请访问 [github.com/goinggo](github.com/goinggo)

一如既往，我希望这份代码可以在某些方面帮上你一点点。

----------------

via: https://www.ardanlabs.com/blog/2013/05/thread-pooling-in-go-programming.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[gogeof](https://github.com/gogeof)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出


