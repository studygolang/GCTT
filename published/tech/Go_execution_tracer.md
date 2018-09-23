已发布：https://studygolang.com/articles/12639

# Go 执行追踪器（execution tracer）

## 概述

你有没有好奇过 Go 运行时是如何调度 goroutine 的？有没有深入研究过为什么有时候加了并发但是性能没有提高？ Go 提供了执行跟踪器，可以帮助你诊断性能问题（如延迟、竞争或低并发等）、解决前面那些疑问。

Go 从 1.5 版本开始有执行跟踪器这么一个工具，原理是：监听 Go 运行时的一些特定的事件，如：

1. goroutine的创建、开始和结束。
2. 阻塞/解锁goroutine的一些事件（系统调用，channel，锁）
3. 网络I/O相关事件
4. 系统调用
5. 垃圾回收

追踪器会原原本本地收集这些信息，不做任何聚合或者抽样操作。对于负载高的应用来说，就可能会生成一个比较大的文件，该文件后面可以通过 `go tool trace` 命令来进行解析。

在执行追踪器之前， Go 已经提供了 pprof 分析器可以用来分析内存和CPU，那么问题来了，为什么还要再添加这么一个官方工具链？ CPU 分析器可以很清晰地查看到是哪个函数最占用 CPU 时间，但是你没办法通过它知道是什么原因导致 goroutine 不运行，也没法知道底层如何在操作系统线程上调度 goroutine 的。而这些恰恰是追踪器所擅长的。追踪器的 [设计文档](https://docs.google.com/document/u/1/d/1FP5apqzBgr7ahCCgFO-yoVhk4YZrNIDNf9RybngBc14/pub) 详尽地说明了引入追踪器的原因以及工作原理。

## 追踪器”之旅“

从一个简单的”Hello world“例子开始。会使用到 `runtime/trace` 包，用于控制开始/停止写追踪数据；数据会被写到标准错误输出。

```go
package main

import (
	"os"
	"runtime/trace"
)

func main() {
	trace.Start(os.Stderr)
	defer trace.Stop()
	// create new channel of type int
	ch := make(chan int)

	// start new anonymous goroutine
	go func() {
		// send 42 to channel
		ch <- 42
	}()
	// read from channel
	<-ch
}
```

本例中，先创建了一个无缓冲 channel,然后初始化一个 goroutine 通过 channel 发送数字 42。主 goroutine 会一直阻塞直到另外的那个 goroutine 通过 channel 发送一个值过来。

通过运行 `go run main.go 2> trace.out` 命令，将追踪信息输出到 `trace.out` 文件，可以用 `go tool trace trace.out` 命令读取该文件。

> Go 1.8 版本之前，想要分析追踪数据要求提供可执行二进制文件和追踪的数据这两个东西；而对于 Go 1.8 版本之后编译的程序，执行 `go tool trace` 就只用提供追踪的数据，它包含了所有内容。

运行过该命令后，会打开一个浏览器窗口，显示几个选项。每个选项打开后是追踪器不同的视图，包含程序执行的不同信息。

![Trace](https://raw.githubusercontent.com/studygolang/gctt-images/master/execution-tracer/trace-opts.png)

1. View trace（查看追踪信息）

提供了一个最复杂、最强大的交互式可视化界面，显示整个程序执行的时间线。举个例子，界面展示每个虚拟处理器上跑了什么以及被阻塞等待重新跑的有哪些。后文中会更详细的介绍这个可视化界面。注意这个界面只支持 chrome 浏览器。

2. Goroutine analysis ( Goroutine 分析)

显示整个执行过程中每种类型的 Goroutine 的数量。选择一种类型，就可以看到每个该类型的 Goroutine 的信息。例如，每个 Goroutine 尝试获取一把互斥锁等待了多久、或尝试从网络读取数据阻塞了多久，或等待调度用了多久等等。

3. Network/Sync/Syscall blocking profile （网络/同步/系统调用 阻塞分析）

这块包含了一些图表，展示 Goroutine 在每个以上这些资源上阻塞了多久。有点类似 pprof 的内存/ CPU 分析器提供的数据。是个分析锁竞争的利器。

4. Scheduler latency profiler （调度延迟分析器）

提供了对调度层面信息的计时统计数据，显示调度过程哪块最耗时。

## 查看追踪信息

点击“ View trace”链接，跳出窗口展示整个程序执行情况。

> 按下“？”键，会弹出浏览追踪信息的一些快捷键的帮助信息

下图标注了重要的几个部分，都在下面做了介绍：

![View trace](https://raw.githubusercontent.com/studygolang/gctt-images/master/execution-tracer/view-trace.png)

1. Timeline （时间线）

显示执行时的时间，时间单位可以通过导航栏进行调整。用户可以通过键盘快捷键（WASD键，就像游戏里的一样）操纵时间线。

2. Heap (堆)

显示执行期间的内存分配情况，对于查找内存泄露、每次运行垃圾回收能释放多少内存非常有用。

3. Goroutines

显示有多少 groroutine 正在运行，以及每个时间点有多少是可运行的（等待被调度）。可运行 goroutine 数量过多可能意味着有调度竞争，例如，当程序创建了过多的 goroutine 会导致调度器忙不过来。

4. OS Threads （操作系统线程）

显示占用了多少操作系统线程以及系统调用阻塞了多少个线程。

5. Virtual Processors （虚拟处理器）

每一行显示一个虚拟处理器。虚拟进程的数量受 GOMAXPROCS 环境变量控制（默认为机器核心数量）。

6.Goroutines and events （goroutine和事件）

展示每个虚拟处理器上的 goroutine 跑的内容/跑在哪里。连接 goroutine 的线代表事件。在例图中，我们可以看到： Goroutine “G1 runtime.main”产生出两个不同的 Goroutine：G6 和 G5 （前者负责收集追踪信息，后者是使用“go”关键字产生的）。

每个虚拟处理器的第二行可能会显示额外的事件，例如系统调用和运行时事件。同时包括了 goroutine 为运行时做的一些工作（例如，协助垃圾收集）。

下图显示选择一个 goroutine 时得到的信息。

![view-goroutine.png](https://raw.githubusercontent.com/studygolang/gctt-images/master/execution-tracer/view-goroutine.png)

包括如下信息：

* 其名称（Title）
* 起始时间（Start）
* 持续时间（Wall Duration）
* 开始时刻堆栈轨迹
* 结束时刻堆栈轨迹
* 该 goroutine 生成的事件

我们可以看到：该 goroutine 产生两个事件：生成用于追踪的 goroutine 以及在channel上开始发送数字42的 goroutine。

![view-event.png](https://raw.githubusercontent.com/studygolang/gctt-images/master/execution-tracer/view-event.png)

通过点击一个特定的事件（图中一条线或者通过点 goroutine 之后选事件），我们可以看到：
* 事件开始时刻的堆栈轨迹
* 事件的持续时间
* 时间包含的 goroutine

可以通过点击这些 goroutine 导航到他们的追踪数据。

## 阻塞分析

通过追踪还可以得到网络/同步/系统调用阻塞分析的视图。阻塞分析显示的图像视图与 pprof 内存/cpu 分析的视图比较相似。区别在于，这里不显示每个函数分配了多少内存，而是显示每个 goroutine 在特定资源上阻塞了多久。

下图显示对我们示例代码的“同步阻塞分析”。

![blocking-profile.png](https://raw.githubusercontent.com/studygolang/gctt-images/master/execution-tracer/blocking-profile.png)

该图说明主 goroutine 从 channel 接收数据阻塞时间花费 12.08 微秒。这种类型的图对于多个 goroutine 竞争资源锁，查找锁竞争情况非常有用。

## 收集追踪数据

有三种收集追踪数据的方式：

1.应用 `runtime/trace` 包

包含调用 `trace.Start` 和 `trace.Stop`，在 “Hello,Tracing” 例子中已经描述过。

2.使用 `trace=<file>` 测试标志

这种方式对于要被测试的代码或者测试本身收集追踪信息十分有用。

3.使用 debug/pprof/trace 处理器

是从正在运行 web 应用收集追踪数据的最好方式。

## 追踪一个 web 应用

为了能够对 Go 写的正在运行的 web 应用收集追踪信息，需要添加 `/debug/pprof/trace` 处理器。下文示例代码说明对于 `http.DefaultServerMux` 如何做到这一点：通过简单地引入 `net/http/pprof` 包。

```go
package main

import (
	"net/http"
	_ "net/http/pprof"
)

func main() {
	http.Handle("/hello", http.HandlerFunc(helloHandler))

	http.ListenAndServe("localhost:8181", http.DefaultServeMux)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world!"))
}
```

收集追踪信息，我们需要向端点发送请求，例如，`curl localhost:8181/debug/pprof/trace?seconds=10 > trace.out`。该请求会被阻塞 10 秒，追踪信息会被写入 `trace.out` 文件。这样生成的追踪信息可以像之前一样查看：通过 `go tool trace trace.out`。

> 安全提醒：注意，不建议将 pprof 处理器暴露到 Internet 上。推荐将这些端点暴露到一个不同的 http.Server ，该 http.Server 绑定到回环端口。[这篇博客](http://mmcloughlin.com/posts/your-pprof-is-showing) 讨论了存在的风险，同时提供如何恰当地暴露 pprof 处理器的示例代码。

收集追踪数据前，先对服务加一些负载，用到了 `wrk`：

```
$ wrk -c 100 -t 10 -d 60s http://localhost:8181/hello
```

该命令会在 60 秒内通过 10 个线程使用 100 个连接发起请求。跑`wrk`的同时，我们可以收集 5 s 的追踪数据：`curl localhost:8181/debug/pprof/trace?seconds=5 > trace.out`。该命令会在本人 4 CPU 机器上生成一个 5 MB 的文件（文件大小会随负载加大而快速增加）。

同样的，通过 go 工具 trace 命令打开追踪数据：`go tool trace trace.out`。由于工具要分析整个文件，会比前例耗费更长时间。完成之后，页面看起来稍微有点不一样：

```
View trace (0s-2.546634537s)
View trace (2.546634537s-5.00392737s)

Goroutine analysis
Network blocking profile
Synchronization blocking profile
Syscall blocking profile
Scheduler latency profile
```

为了确保浏览器能够渲染所有信息，工具已经将追踪数据分为两个连续的部分。负载更高的应用或者更长时间的追踪数据可能会触发工具将其分割为更多的部分。

点击“View trace (2.546634537s-5.00392737s)”我们可以看到有很多东西：

![trace-web.png](https://raw.githubusercontent.com/studygolang/gctt-images/master/execution-tracer/trace-web.png)

该截图显示 1169 ms ~ 1170 ms 之间开始、1174 ms 之后结束的一个 GC 操作。这段时间内，一个操作系统线程（PROC 1）启动一个 goroutine 专门做 GC，其他的 goroutine 辅助 GC 操作（在 goroutine 行下展示，标示为 MARK ASSIST ）,截频的末尾，我们可以看到大部分分配的内存已被 GC 释放掉。

另外一个特别有用的信息是：处于“Runnable”状态（选择时显示的是13）的 goroutine 数量，如果该数值随时间变大，就意味着我们需要更多 CPU 来处理负载。

## 总结

追踪器是调试并发问题（如锁竞争和逻辑竞争）的一个强大工具。它不能够解决一切问题：它不是追踪哪块代码最耗费 CPU 时间、内存的最佳工具。用`go tool pprof`更适合这样的场景。

该工具对于理解程序不运行时每个 goroutine 在做什么以及按照时间查看程序行为非常有用。收集追踪数据会有一些开销，同时可能会产生较大数据量的数据以供查看。

不幸的是，官方文档缺少一些实验来让我们试验、理解追踪器显示的信息。这也是个给官方文档、社区（如博客）做贡献的[机会](https://github.com/golang/go/issues/16526)。

André 是 [Globo.com](http://www.globo.com/) 的高级软件工程师, 开发 [Tsuru](https://tsuru.io/)项目。 twitter 请@andresantostc, 或者 web 留言https://andrestc.com。

## 参考

1. [Go execution tracer (design doc)](https://docs.google.com/document/u/1/d/1FP5apqzBgr7ahCCgFO-yoVhk4YZrNIDNf9RybngBc14/pub)
2. [Using the go tracer to speed fractal rendering](https://medium.com/justforfunc/using-the-go-execution-tracer-to-speed-up-fractal-rendering-c06bb3760507)
3. [Go tool trace](https://making.pusher.com/go-tool-trace/)
4. [Your pprof is showing](http://mmcloughlin.com/posts/your-pprof-is-showing)

---

via: https://blog.gopheracademy.com/advent-2017/go-execution-tracer/

作者：[André Carvalho](https://blog.gopheracademy.com/advent-2017/go-execution-tracer/)
译者：[dongfengkuayue](https://github.com/dongfengkuayue)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
