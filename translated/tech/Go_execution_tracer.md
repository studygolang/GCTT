# Go execution tracer

## 概述
你有没有好奇过 Go 运行时是如何调度 goroutine 的？有没有深入研究过为什么有时候加了并发但是性能没有提高？ Go 提供了执行跟踪器，可以帮助你诊断性能问题（如延迟、竞争或低并发等）、解决前面那些疑问。

Go 从 1.5 版本开始有执行跟踪器这么一个工具，原理是：监听 Go 运行时的一些特定的事件，如：

1. goroutine的创建、开始和结束。
1. 阻塞/解锁goroutine的一些事件（系统调用，channel，锁）
1. 网络I/O相关事件
1. 系统调用
1. 垃圾回收

追踪器会原原本本地收集这些信息，不做任何聚合或者抽样操作。对于负载高的应用来说，就可能会生成一个比较大的文件，该文件后面可以通过 `go tool trace` 命令来进行解析。

在执行追踪器之前， Go 已经提供了 pprof 分析器可以用来分析内存和CPU，那么问题来了，为什么还要再添加这么一个官方工具链？ CPU 分析器可以很清晰地查看到是哪个函数最占用 CPU 时间，但是你没办法通过它知道是什么原因导致 goroutine 不运行，也没法知道底层如何在操作系统线程上调度 goroutine 的。而这些恰恰是追踪器所擅长的。追踪器的 [设计文档](https://docs.google.com/document/u/1/d/1FP5apqzBgr7ahCCgFO-yoVhk4YZrNIDNf9RybngBc14/pub)
详尽地说明了引入追踪器的原因以及工作原理。

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

本例中，先创建了一个无缓冲 channel,然后初始化一个 goroutine 通过 channel 发送数字42。主 goroutine 会一直阻塞直到另外的那个 goroutine 通过 channel 发送一个值过来。

通过运行`go run main.go 2> trace.out`命令，将追踪信息输出到`trace.out`文件，可以用`go tool trace trace.out`命令读取该文件。

> Go 1.8 版本之前，想要分析追踪数据要求提供可执行二进制文件和追踪的数据这两个东西；而对于 Go 1.8 版本之后编译的程序，执行`go tool trace`就只用提供追踪的数据，它包含了所有内容。

运行过该命令后，会打开一个浏览器窗口，显示几个选项。每个选项打开后是追踪器不同的视图，包含程序执行的不同信息。

![Trace](https://blog.gopheracademy.com/postimages/advent-2017/go-execution-tracer/trace-opts.png)

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

![View trace](https://blog.gopheracademy.com/postimages/advent-2017/go-execution-tracer/view-trace.png)

1. Timeline （时间线）


