首发于：https://studygolang.com/articles/19396

# Go 的调度器追踪

## 简介

我喜欢 Go 语言的一个原因就是因为它可以生成分析和调试信息。当程序在执行的时候 Go 有一个特殊的环境变量 GODEBUG，它在运行阶段可以生成调试信息。你可以获取程序回收器和调度器的概要信息以及详细的信息。更主要的是你不需要添加任何的额外工作重新编译就可以完成调试

在这篇文章里，我将通过一个简单的 Go 程序演示如果使用调度跟踪信息。如果你对调度器有一定的了解那么它对你是用的。我建议先阅读下面两篇文章：

并发、goroutines 和 GOMAXPROCS

http://www.goinggo.net/2014/01/concurrency-goroutines-and-gomaxprocs.html

调度器

http://morsmachine.dk/go-scheduler

## 代码部分

我们将用一个简单的程序验证和解释 GODEBUG 的结果：

### 清单 1

```go
package main

import (
    "sync"
    "time"
)

func main() {
    var wg sync.WaitGroup
    wg.Add(10)

    for i := 0; i < 10; i++ {
        Go work(&wg)
    }

    wg.Wait()

    // Wait to see the global run queue deplete.
    time.Sleep(3 * time.Second)
}

func work(wg *sync.WaitGroup) {
    time.Sleep(time.Second)

    var counter int
    for i := 0; i < 1e10; i++ {
        counter++
    }

    wg.Done()
}
```

清单 1 中的例子是为了演示运行时调试器给我们的调试信息。在第 12 秒 for 循环进行 10 次 goroutines。然后主函数在第 16 行的时候等待所有 goroutines 执行完成。在第 22 行 work 函数里面先 sleep 一秒然后 counter 变量 ++ 执行一百亿次。当 for 循环执行完成后调用 Done 方法最后 return。

在设置 GODEBUG 之前，先用 go build 编译代码。这个变量由运行时获取，因此运行 Go 命令也将产生跟踪输出。如果 GODEBUG 结合 go run 使用，那么你将看到运行之前的跟踪调试信息。

现在我们使用 go build 编译上面的例子，这样我们就可以携带 GODEBUG 选项运行例子了：

```
go build example.go
```

## 调度器跟踪摘要信息

schedtrace 选项使代码在运行时每隔 X 秒输出一行调度器的状态信息到标准错误输出。现在我们运行程序并且设置 GODEBUG 选项：

```
GOMAXPROCS=1 GODEBUG=schedtrace=1000 ./example
```

当程序运行起来，我们就可以看到跟踪信息。程序本身没有输出任何的标准输出和标准错误输出，所以我们可以专注在跟踪信息上。现在我们看最开始的两条跟踪信息：

```
SCHED 0ms: Gomaxprocs=1 idleprocs=0 threads=2 spinningthreads=0 idlethreads=0
runqueue=0 [1]

SCHED 1009ms: Gomaxprocs=1 idleprocs=0 threads=3 spinningthreads=0 idlethreads=1
runqueue=0 [9]
```

现在我们来分析每个字段的意思并理解在例子中的作用：

```
1009ms        : 时间是以毫秒为单位，表示了从程序开始运行的时间。这个是第一秒的跟踪信息。

gomaxprocs=1  : 配置的最大的处理器的个数
                本程序只配置了一个

更多说明：

这里的处理器是逻辑上的处理器而非物理上的处理器。调度器在这些逻辑处理器上运行 goroutines，这些逻辑处理器通过所附着的操作系统线程绑定在物理处理器上的。 操作系统将中可见的物理存储器上调度线程。

threads=3     : 运行时管理的线程个数。有三个线程存在，一个是给处理器使用，其他两个时被运行时使用。

idlethreads=1 : 空闲的线程个数。一个空闲，两个正在运行。

idleprocs=0   : 空闲的处理器个数。这里空闲个数为 0，有一个运行中。

runqueue=0    : 在全局运行队列中等待的 goroutinue 数量。所有可运行的 goroutinue 都被移到了本地的运行队列中。

[9]           : 本地运行队列中等待的 goroutine 数量。当前有 9 个 goroutine 在本地运行队列等待。
```

在运行时的摘要信息里面给了我们很多非常有用的信息。我们从运行的一秒的标记里面可以看到跟踪的信息。我们可以看到一个 gorountine 如何运行，其它 9 个 goroutine 都在 local run queue 中等待。

### 图 1

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/20150223-Scheduler-Tracing-In-Go/diagram1.png)

从图一中可以看到处理器用字母 "P" 代表，线程使用字母 "M" 代码，goroutines 使用字母 "G" 代表。我们可以看到当 runqueue 的值为 0 时，全局的 run queue 是空的。处理器将会把 gorountine 运行在 idleprocs 为 0 的上面运行。我们运行的其他九个 goroutine 仍然在等待。

那如果有多个处理器的时候，那我们该如何跟踪呢？那我们再运行一次程序并添加 GOMAXPROCS 选项，看看会输出什么跟踪信息：

```
GOMAXPROCS=2 GODEBUG=schedtrace=1000 ./example

SCHED 0ms: gomaxprocs=2 idleprocs=1 threads=2 spinningthreads=0
idlethreads=0 runqueue=0 [0 0]

SCHED 1002ms: gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=1
idlethreads=1 runqueue=0 [0 4]

SCHED 2006ms: gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=0
idlethreads=1 runqueue=0 [4 4]

…

SCHED 6024ms: gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=0
idlethreads=1 runqueue=2 [3 3]

…

SCHED 10049ms: gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=0
idlethreads=1 runqueue=4 [2 2]
…

SCHED 13067ms: gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=0
idlethreads=1 runqueue=6 [1 1]

…

SCHED 17084ms: gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=0
idlethreads=1 runqueue=8 [0 0]

…

SCHED 21100ms: gomaxprocs=2 idleprocs=2 threads=4 spinningthreads=0
idlethreads=2 runqueue=0 [0 0]
```

我们重点看一下第二秒的信息：

```
SCHED 2002ms: gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=0
idlethreads=1 runqueue=0 [4 4]

2002ms        : This is the trace for the 2 second mark.
gomaxprocs=2  : 2 processors are configured for this program.
threads=4     : 4 threads exist. 2 for processors and 2 for the runtime.
idlethreads=1 : 1 idle thread (3 threads running).
idleprocs=0   : 0 processors are idle (2 processors busy).
runqueue=0    : All runnable goroutines have been moved to a local run queue.
[4 4]         : 4 goroutines are waiting inside each local run queue.
...
```

### 图 2

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/20150223-Scheduler-Tracing-In-Go/diagram2.png)

我们看一下第二秒跟踪信息在图 2 中的信息，我们可以看到一个 goroutine 在每个处理器中是如何运行的。并且我们可以看到 8 个 goroutine 在 local run queues 中等待，每个 local run queues 各四个。在跟踪信息的第六秒发生了改变：

```
SCHED 6024ms: gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=0
idlethreads=1 runqueue=2 [3 3]

idleprocs=0 : 0 processors are idle (2 processors busy).
runqueue=2  : 2 goroutines returned and are waiting to be terminated.
[3 3]       : 3 goroutines are waiting inside each local run queue.
```

### 图 3：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/20150223-Scheduler-Tracing-In-Go/diagram3.png)

当到第六秒的时候发生了变化。从图 3 中有两个 goroutine 完成工作之后被移到了 global run queue 里面，并且我们仍然有两个 goroutine 在运行。每个 processor 各运行一个，在每个 local run queue 里面各有三个在等待。

注释：

在很多情况下，goroutine 运行完成之后并不会被移到全局的 run queue 中。这个例子创建的条件比较特殊。因为这个例子的 for 循环运行了 10 秒多的时间但是没有任何的函数调用。10 秒是调度的次数在调度器里面。在执行 10 秒后，调度器尝试先去取 goroutine。但是这些 goroutine 不能被占用，因为它们没有调用任何的函数。在这种情况下，一旦 goroutine 调用 wg.Done，这个 goroutine 将立即被占用，然后移到全局的 run queue 中。

当到第 17 秒的时候，我们可以看到最后两个 goroutine 都在运行了：

```
SCHED 17084ms: Gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=0
idlethreads=1 runqueue=8 [0 0]

idleprocs=0 : 0 processors are idle (2 processors busy).
runqueue=8  : 8 goroutines returned and are waiting to be terminated.
[0 0]       : No goroutines are waiting inside any local run queue.
```

### 图 4：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/20150223-Scheduler-Tracing-In-Go/diagram4.png)

在图 4 中，我们可以看到 8 个 goroutines 在 global run queue 中，还有两个仍然在运行。这个时候每个 local run queue 已经空了。最终的跟踪信息在第 12 秒结束：

```
SCHED 21100ms: Gomaxprocs=2 idleprocs=2 threads=4 spinningthreads=0
idlethreads=2 runqueue=0 [0 0]

idleprocs=2 : 2 processors are idle (0 processors busy).
runqueue=0  : All the goroutines that were in the queue have been terminated.
[0 0]       : No goroutines are waiting inside any local run queue.
```

### 图 5：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/20150223-Scheduler-Tracing-In-Go/diagram5.png)

至此，所有的 goroutine 都执行完了并且已经结束。

## 详细的跟踪信息

概要的跟踪信息是非常有用的，但是有的时候你需要更详细的信息。如果需要更详细的每个处理器，线程的或者 goroutine 的跟踪信息我们可以添加 scheddetail 这个选项。我们再一次运行程序，设置 GODEBUG 选项获取更详细的跟踪信息：

```
GOMAXPROCS=2 GODEBUG=schedtrace=1000,scheddetail=1 ./example
```

下面是第四秒的输出信息：

```
SCHED 4028ms: Gomaxprocs=2 idleprocs=0 threads=4 spinningthreads=0
idlethreads=1 runqueue=2 gcwaiting=0 nmidlelocked=0 stopwait=0 sysmonwait=0
P0: status=1 schedtick=10 syscalltick=0 m=3 runqsize=3 gfreecnt=0
P1: status=1 schedtick=10 syscalltick=1 m=2 runqsize=3 gfreecnt=0
M3: p=0 curg=4 mallocing=0 throwing=0 gcing=0 locks=0 dying=0 helpgc=0
    spinning=0 blocked=0 lockedg=-1
M2: p=1 curg=10 mallocing=0 throwing=0 gcing=0 locks=0 dying=0 helpgc=0
    spinning=0 blocked=0 lockedg=-1
M1: p=-1 curg=-1 mallocing=0 throwing=0 gcing=0 locks=1 dying=0 helpgc=0
    spinning=0 blocked=0 lockedg=-1
M0: p=-1 curg=-1 mallocing=0 throwing=0 gcing=0 locks=0 dying=0 helpgc=0
    spinning=0 blocked=0 lockedg=-1
G1: status=4(semacquire) m=-1 lockedm=-1
G2: status=4(force gc (idle)) m=-1 lockedm=-1
G3: status=4(GC sweep wait) m=-1 lockedm=-1
G4: status=2(sleep) m=3 lockedm=-1
G5: status=1(sleep) m=-1 lockedm=-1
G6: status=1(stack growth) m=-1 lockedm=-1
G7: status=1(sleep) m=-1 lockedm=-1
G8: status=1(sleep) m=-1 lockedm=-1
G9: status=1(stack growth) m=-1 lockedm=-1
G10: status=2(sleep) m=2 lockedm=-1
G11: status=1(sleep) m=-1 lockedm=-1
G12: status=1(sleep) m=-1 lockedm=-1
G13: status=1(sleep) m=-1 lockedm=-1
G17: status=4(timer goroutine (idle)) m=-1 lockedm=-1
```

概要部分基本相同，但是有了关于处理器，线程以及 goroutine 更详细的信息。我们看看关于处理器的信息：

```
P0: status=1 schedtick=10 syscalltick=0 m=3 runqsize=3 gfreecnt=0

P1: status=1 schedtick=10 syscalltick=1 m=2 runqsize=3 gfreecnt=0
```

P 代表一个处理器。因为 GOMAXPROCS 被设置为 2，我们可以看到处理器的列表。下来我们看看线程：

```
M3: p=0 curg=4 mallocing=0 throwing=0 gcing=0 locks=0 dying=0 helpgc=0
spinning=0 blocked=0 lockedg=-1

M2: p=1 curg=10 mallocing=0 throwing=0 gcing=0 locks=0 dying=0 helpgc=0
spinning=0 blocked=0 lockedg=-1

M1: p=-1 curg=-1 mallocing=0 throwing=0 gcing=0 locks=1 dying=0 helpgc=0
spinning=0 blocked=0 lockedg=-1

M0: p=-1 curg=-1 mallocing=0 throwing=0 gcing=0 locks=0 dying=0 helpgc=0
spinning=0 blocked=0 lockedg=-1
```

M 代码一个线程，因为 threadsb 被设置为 4，所以我们能看到 4 个线程的详细信息。并且线程详细信息里面展示了线程所在的处理器：

```
P0: status=1 schedtick=10 syscalltick=0 m=3 runqsize=3 gfreecnt=0

M3: p=0 curg=4 mallocing=0 throwing=0 gcing=0 locks=0 dying=0 helpgc=0
spinning=0 blocked=0 lockedg=-1
```

这里展示了线程 M3 是如何绑定在处理器 P0 上的。这个信息在 P 和 M 的跟踪信息里面都有。

G 代码一个 goroutine。在第四秒的时候我们可以看到有 14 个 goroutine 存在，有 17 个 goroutine 被创建。我们之所以知道总共的 goroutine 的个数是因为最后在 G 列表里面绑定的数字：

```
G17: status=4(timer goroutine (idle)) m=-1 lockedm=-1
```

如果成行继续创建 goroutine，我们就可以看到这个数字将呈线性的增长。如果这个程序是拦截 Web 请求的例子，那么我们可以用这个数字来确认请求的拦截次数。只有当拦截请求期间不再创建任何的 goroutine，这个才会被关闭。

下面我们看看在 main 方法里面的 goroutine：

```
G1: status=4(semacquire) m=-1 lockedm=-1

30     wg.Done()
```

我们可以看到在 main 方法中 goroutine 的状态为 4，状态被锁定在 semacquire 状态，这个状态表示等待调用。

为了更好的理解剩下的跟踪信息，先来了解一下状态代码的意思。下面是状态值列表，这些声明在 runtime 包的头文件里面的：

```
status: http://golang.org/src/runtime/
Gidle,            // 0
Grunnable,        // 1 runnable and on a run queue
Grunning,         // 2 running
Gsyscall,         // 3 performing a syscall
Gwaiting,         // 4 waiting for the runtime
Gmoribund_unused, // 5 currently unused, but hardcoded in gdb scripts
Gdead,            // 6 goroutine is dead
Genqueue,         // 7 only the Gscanenqueue is used
Gcopystack,       // 8 in this state when newstack is moving the stack
```

对照他们的状态我们能更好的理解我们创建的 10 个 goroutine 都在做什么。

```
// goroutines running in a processor. (idleprocs=0)
G4: status=2(sleep) m=3 lockedm=-1   – Thread M3 / Processor P0
G10: status=2(sleep) m=2 lockedm=-1  – Thread M2 / Processor P1

// goroutines waiting to be run on a particular processor. (runqsize=3)
G5: status=1(sleep) m=-1 lockedm=-1
G7: status=1(sleep) m=-1 lockedm=-1
G8: status=1(sleep) m=-1 lockedm=-1

// goroutines waiting to be run on a particular processor. (runqsize=3)
G11: status=1(sleep) m=-1 lockedm=-1
G12: status=1(sleep) m=-1 lockedm=-1
G13: status=1(sleep) m=-1 lockedm=-1

// goroutines waiting on the global run queue. (runqueue=2)
G6: status=1(stack growth) m=-1 lockedm=-1
G9: status=1(stack growth) m=-1 lockedm=-1
```

基于对 scheduler 的简单了解以及对我们例子程序的了解，我们对程序如何被 scheduled，每个处理器的状态是什么，线程以及 goroutine 等信息都有了全面的了解。

## 总结：

GODEBUG 是一个非常有效的手段去跟踪程序的执行。它帮助你了解很多程序的执行详细信息。如果你想了解更多，从写一些简单的例子入手，你可以预测它可能来自于 scheduler 的跟踪信息。在尝试去看一些复杂例子的跟踪信息之前，先学会如何预测。

---

via: https://www.ardanlabs.com/blog/2015/02/scheduler-tracing-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[amei](https://github.com/amei)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
