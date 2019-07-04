首发于：https://studygolang.com/articles/21569

# Go 中的垃圾回收：第一部分 - 基础

这是三篇系列文章的第一篇博文，系列文章提供了 Go 中垃圾回收背后的机制和概念的理解。这篇博文主要介绍回收器的基础概念。

三篇系列文章的索引：

1) [Go 中的垃圾回收：第一部分 - 概念](https://studygolang.com/articles/21569)
2) [Go 中的垃圾回收：第二部分 -GC 追踪](https://studygolang.com/articles/21570)
3) 即将诞生

## 简介

垃圾回收器负责跟踪堆内存分配，释放无用的分配内存以及维护在用分配内存。语言如何设计去实现这些行为是很复杂的，但不应该要求应用开发者为了构建软件而去理解细节。而且，对于语言不同版本的 VM 和运行时（runtime），这些细节的实现一直都在发展变化。对于应用开发者而言，重要的是保持一个良好的工作模型，了解垃圾回收器对其语言的行为以及如何在不关心其实现的情况下，对这种行为表示友好。

在 1.12 版本，Go 语言使用了无分代同步三色标记清除回收器（non-generational concurrent tri-color mark and sweep collector）。如果想形象化地了解标记清除回收器如何工作，Ken Fox 写了这篇提供了动画的[好文章](https://spin.atomicobject.com/2014/09/03/visualizing-garbage-collection-algorithms)。Go 回收器的实现随着发行版而发展变化，所以一旦下一版本发行，任意讲述实现细节的博文将不再准确。

虽然这样，我将在本文中做的分析不会关注实际的实现细节，而是关注你经历到的行为以及你希望在未来几年看到的行为。在这篇文章中，我将和你分享回收器的行为，并解释如何对该行为表示友好，无论当前实现或未来如何变化。这些都会让你成为更好的 Go 开发者。

*注意：这里你可以对有关[垃圾回收器](https://github.com/ardanlabs/gotraining/tree/master/reading#garbage-collection) 以及 Go 实际的回收器进行扩展阅读*

## 堆不是一个容器

我永远不会把堆叫做用来存储或释放值的容器。重要的是，要理解没有线性限制的内存都定义为“堆”。应该认为应用程序进程空间中的保留的任何内存都可用于堆内存分配。无论任何给定的堆内存分配属于虚拟或物理存储都与我们的模型无关。这种理解将帮助您更好地了解垃圾回收器的工作原理。

## 回收器行为

当某次回收开始，回收器经历三个阶段的工作。其中两个阶段引起 Stop The World ( STW ) 的延迟，另外的阶段会产生降低程序吞吐量的延迟。这三个阶段为：

- 标记开始 - STW
- 标记中 - 并发
- 标记结束 - STW

以下为每一个阶段的细分

### 标记开始 - STW

当回收开始，首先执行的动作是打开写屏障。写屏障的目的是允许回收器在收集过程保持堆上的数据完整性，因为回收器和应用程序的 Goroutine 会并发执行。

为了打开写屏障，必须停止应用运行的所有 Goroutine 。这个动作通常非常快，平均在 10~30 微秒之间。这是指，如果应用程序的 Goroutine 表现正常情况下。

*注意：为了更好理解这些调度图，请务必阅读[Go Scheduler](https://studygolang.com/articles/14264) 上的系列文章。*

### 图 1

![figure1](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure1.png)

图 1 展示了回收之前应用有 4 个 Goroutine 在运行。这 4 个 Goroutine 都应该被停掉，唯一方法就是观察和等待每个 Goroutine 进行函数调用。函数调用保证了 Goroutine 在一个安全的点上被停掉。如果其中一个 Goroutine 没有进行函数调用，但其他的却做了函数调用，这会发生什么呢？

### 图 2

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure2.png)

图 2 展示了一个真正的问题。在 P4 上运行的 Goroutine 停下来之前都不会开始回收。然而 P4 的 Goroutine 是不会停止的，因为它正在[紧密循环](https://github.com/golang/go/issues/10958) 地进行某些数学运算。

### 清单 1

```go
func add(numbers []int) int {
     var v int
     for _, n := range numbers {
         v += n
     }
     return v
}
```

清单 1 展示了运行在 P4 上的 Goroutine 正在执行的代码。goroutine 可能以不合理的大量时间运行从而无法停止，这取决于切片的大小。这种代码会拖延回收的启动，更糟糕的是当回收器在等待时，其他的 P 都不能为任何其他的 Goroutine 服务。gorotine 在一个合理的时间范围内进行函数调用显得极其重要。

*注意 : 这是语言团队想要在 1.14 通过加入[preemotive](https://github.com/golang/go/issues/24543) 技术到调度中去修复的问题*

### 标记中 - 并发

一旦开启了写障碍，回收器进入标记阶段。回收器做的第一件事是占用 CPU 可用处理能力的 25%。回收器使用 Gorouitne 去做回收工作，也同样需要应用程序的 Goroutine 使用的 P 和 M（译者注：从此处开始作者将 G 划分了两类，一类是应用程序用于自身工作的 Gourinte ，下文称应用  Goroutine，一类是用于 GC 的 Goroutine，这样会更好理解）。这意味着对于我们四个线程的 Go 程序，有一个完整的 P 会专门用来进行回收工作。

### 图 3

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure3.png)

图 3 展示了回收器在回收过程中如何为自身占有 P1。现在回收器开始标记阶段了，标记阶段标记堆内存中仍在使用的值。这个工作先检查栈内所有存活的 Gorouitne，去寻找堆内存的根指针。然后回收器必须从那些根指针遍历堆内存图。当标记工作发生在 P1 上，应用程序可以继续在 P2, P3 和 P4 上同步工作。这意味着回收器的影响被最小化到当前 CPU 处理能力的 25%。

我希望这个事就这样完了然而并没有。如果在收集过程中，确认在 P1 上专用于 GC 的 Goroutine 在堆内存达到上限之前无法完成标记工作，该怎么办？如果 3 个 Goroutine 中，其中一个所进行的应用工作导致回收器无法及时完成 ( 标记工作 ) 又怎么办？ ( 译者注：此处的意思为内存分配过快 )。在这种情况下，新的分配必须放慢速度，特别是从那个 ( 导致标记无法完成的 ) Goroutine。

如果回收器确定它需要减慢分配，它会招募应用 Goroutine 以协助标记工作，这叫做辅助标记。任何应用 Goroutine 花费在辅助标记的时间长度与它添加到堆内存中的数据量成正比。辅助标记的一个好处是它有助于更快地完成回收。

### 图 4

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure4.png)

图 4 展示了运行在 P3 上的应用 Goroutine 如何进行辅助标记来帮助回收工作的。希望其他的应用 Goroutine 不用参与进来。分配动作较多的应用可以看到，大部分运行中的 Goroutine 在回收过程中都进行了少量的辅助标记。

回收器的一个目标是消除对辅助标记的需求。如果每次回收都需要大量的辅助标记才能结束，那么回收器很快就会开始下一次的垃圾回收。为了不那么快进行下一次的回收，努力去减少辅助标记的数量是必要的。

### 标记结束 - STW

一旦标记工作完成，下阶段就是标记结束了。到这个阶段，写屏障会被停止，各样的清洁工作会被执行，然后计算好下一次的回收目标。在标记阶段，发现自身处于紧密循环的 Goroutine 也会延长这个阶段的时长。

### 图 5

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure5.png)

图 5 展示了在标记结束阶段完成时，所有的 Goroutine 如何停止的。这个动作通常平均在 60 到 90 微秒之间。这个阶段可以不需要 STW，但通过使用 STW，代码会更简单，小小的收益抵不上增加的复杂度。

一旦回收完成，每个 P 都能服务于应用 Goroutine，然后应用回到全力运行状态。

### 图 6

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure6.png)

图 6 展示了一旦回收完成，所有可选的 P 如何再次处理应用的工作，应用回到回收开始前的全力运行状态。

### 并发清除

在回收完成之后有另一个叫清除的动作发生。清除是指回收堆内存中，未标记为使用中的值所关联的内存。该动作会在应用程序 Goroutine 尝试分配新值到堆内存时发生。清除的延迟被算到在堆内存中执行分配的成本中，与垃圾回收相关的任何延迟无关。

下面是我机器上的追踪样本，有 12 条硬件线程可用于执行 Gorouitne。

### 图 7

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure7.png)

图 7 展示了追踪的部分快照。你可以看到在回收过程中 ( 盯着顶部的蓝色 GC 线 )，12 个 P 中的其中 3 个如何专门用于 GC。你可以看到 Goroutine2450，1978 和 2696 在这段时间进行了数次辅助标记，而不是执行应用的工作。在回收的最后，只有一个 P 用于 GC 并最终执行 STW ( 标记结束 ) 的工作。

在回收完成后，应用程序回到全力运行状态。此外你看到在 Goroutine 下面有一些玫瑰色的线条。

### 图 8

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure8.png)

图 8 展示了那些玫瑰色的线条如何代表 Goroutine 不执行应用的工作而进行清除工作的时刻。这些都是 Goroutine 尝试分配新值到堆内存的时刻。

### 图 9

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure9.png)

图 9 展示了其中一个进行清除动作的 Gorouitne 最后的栈跟踪情况。`runtime.mallocgc` 的调用会导致在堆内存分配新值。`runtime.(*mcache).nextFree` 的调用引起清除动作。一旦堆内存上没有更多的已分配内存需要回收，就不会再看到 `nextFree` 的调用。

刚刚描述的回收动作仅仅在回收过程开始和进行中才会发生。配置项 GC 百分比在决定何时启动垃圾回收任务中扮演重要角色。

## GC 百分比

运行过程中有一个配置项叫 GC 百分比，默认值设置为 100。这个值代表了在下次回收开始前能分配多大的堆内存。设置 GC 百分比为 100 意味着，根据回收完成后标记为存活的堆内存量，下一次回收必须在堆内存上添加 100 ％ 以上的新分配 ( 内存 ) 才启动。

举个例子，想象某次回收完成后堆内存有 2MB 存活。( 译者注：后半句话应该是分配 2MB 后 GC 才会开始，作者省了。。。)

*注意 : 在这篇博文中堆内存的图不代表使用 Go 的时候的真实情况。Go 中的堆内存通常是碎片化和混乱的，而且没有图像所描绘的那么清晰。这些图在更为易于理解的方式上提供可视化堆内存的方法，这种方式对于你将体验的行为是准确的。*

### 图 10

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure10.png)

图 10 展示了最后的回收完成后，使用中的 2 MB 堆内存。因为 GC 百分比设置为 100%，下一次回收需要在额外分配 2 MB 的堆内存时才开始，或者在超过 2 MB 之前开始。

### 图 11

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure11.png)

图 11 展示了超过 2MB 堆内存正在使用，这会触发回收。查看该动作所有 ( 细节 ) 的方法是，为每次回收生成 GC 追踪。

## GC 追踪

GC 追踪可以通过在运行任意 Go 应用时包含环境变量 `GODEBUG` 并指定 `gctracec=1` 来生成。每次回收发生，运行时会将 GC 追踪信息写到 `stderr` 中。

### 清单 2

```bash
GODEBUG=gctrace=1 ./app

gc 1405 @6.068s 11%: 0.058+1.2+0.083 ms clock, 0.70+2.5/1.5/0+0.99 ms CPU, 7->11->6 MB, 10 MB Goal, 12 P

gc 1406 @6.070s 11%: 0.051+1.8+0.076 ms clock, 0.61+2.0/2.5/0+0.91 ms CPU, 8->11->6 MB, 13 MB Goal, 12 P

gc 1407 @6.073s 11%: 0.052+1.8+0.20 ms clock, 0.62+1.5/2.2/0+2.4 ms CPU, 8->14->8 MB, 13 MB Goal, 12 P
```

清单 2 展示了如何使用 `GODEBUG` 变量来生成 GC 追踪。清单也展示了运行 Go 应用生成的 3 份追踪信息。

以下是通过查看清单中的第一个 GC 追踪线来拆解 GC 追踪中每个值的含义。

### 清单 3

```bash
gc 1405 @6.068s 11%: 0.058+1.2+0.083 ms clock, 0.70+2.5/1.5/0+0.99 ms CPU, 7->11->6 MB, 10 MB Goal, 12 P

// General
gc 1404     : 自程序启动以来，1404 的 GC 运行 ( 译者注：此处应当是笔误，联系上文其实是 1405)
@6.068s     : 自程序启动至此总共 6s
11%         : 到目前为止，可用 CPU 的 11% 被用于 GC
// Wall-Clock
0.058ms     : STW     : 标记开始，开启写障碍
1.2ms       : 并发     : 标记中
0.083ms     : STW     : 标记结束 - 关闭写障碍并清除

// CPU Time
0.70ms      : STW        : 标记开始
2.5ms       : 并发        : 辅助标记时间 (GC 按照分配执行 )
1.5ms       : 并发        : 标记 - 后台 GC 时间
0ms         : 并发        : 标记 - 空闲 GC 时间
0.99ms      : STW        : 标记结束

// Memory
7MB         : 标记开始前使用中的堆内存
11MB        : 标记完成后使用中的堆内存
6MB         : 标记完成后被标记为存活的堆内存
10MB        : 标记完成后使用中的堆内存收集目标

// Threads
12P         : 用于运行 Gorouitne 的物理调度器或线程的数量
```

清单 3 展示了第一条 GC 追踪线的实际数字所代表的含义，按行进行拆解。我后面会谈及这些值中的大部分，但现在只要关注 1405 的 GC 追踪的内存部分。

### 图 12

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure12.png)

### 清单 4

```bash
// Memory
7MB         : 标记开始前使用中的堆内存
11MB        : 标记完成后使用中的堆内存
6MB         : 标记完成后被标记为存活的堆内存
10MB        : 标记完成后使用中的堆内存收集目标
```

清单 4 中的 GC 追踪线想告诉你的是，在标记工作开始前使用中的堆内存大小为 7 MB。当标记工作完成时，使用中的堆内存大小达到了 11 MB。这意味着在回收过程中出现了额外的 4 MB 内存分配。在标记工作完成后被标记为存活的堆内存大小为 6 MB。这意味着在下次回收开始前应用可以增加使用的堆内存到 12 MB ( 存活堆大小 6 MB 的 100%)。

你可以看到回收器与其目标有 1 MB 的偏差，标记工作完成后正在使用的堆内存量为 11 MB 而不是 10 MB。这没关系，因为目标是根据当前正在使用的堆内存量、标记为存活的堆内存量以及在回收运行时将会发生的其他分配的时间计算情况来计算的。在这种情况下，应用程序做了一些需要在标记之后使用更多的堆内存的事情，而不是像预期那样。

如果查看下一个 GC 跟踪线（1406），你会看到事情在 2 ms 内发生了变化。

### 图 13

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure13.png)

### 清单 5

```bash
gc 1406 @6.070s 11%: 0.051+1.8+0.076 ms clock, 0.61+2.0/2.5/0+0.91 ms CPU, 8->11->6 MB, 13 MB Goal, 12 P

// Memory
8MB         : 标记开始前使用中的堆内存
11MB        : 标记完成后使用中的堆内存
6MB         : 标记开完成后被标记为存活的堆内存
13MB        : 标记完成后的使用堆内存收集目标
```

清单 5 展示了这次回收如何在前一次回收 2 ms 之后开始了，即便使用中的堆内存仅仅达到了 8 MB，而所允许的是 12 MB。这需要特别注意，如果回收器认为早点开始回收会好一点，那么就会提前开始。在这种情况下，它提前开始大概是因为应用在进行大量的分配工作，然后回收器想要降低这次回收的辅助标记的延时。

还有两件事要注意。回收器这次在他的目标之内。标记完成后使用中堆内存的大小是 11 MB 而不是 13 MB，少了 2 MB。标记完成后标记为存活的堆内存大小一样为 6 MB。

附注一点，你可以通过增加 `gcpacertrace=1` 标志从 GC 追踪获取更多细节，这会让回收器打印更多有关并发步调器的内部状态。

### 清单 6

```bash
$ export GODEBUG=gctrace=1,gcpacertrace=1 ./app

样本输出：
gc 5 @0.071s 0%: 0.018+0.46+0.071 ms clock, 0.14+0/0.38/0.14+0.56 ms CPU, 29->29->29 MB, 30 MB Goal, 8 P

pacer: sweep done at heap size 29MB; allocated 0MB of spans; swept 3752 pages at +6.183550e-004 pages/byte

pacer: assist ratio=+1.232155e+000 (scan 1 MB in 70->71 MB) workers=2+0

pacer: H_m_prev=30488736 h_t=+2.334071e-001 H_T=37605024 h_a=+1.409842e+000 H_a=73473040 h_g=+1.000000e+000 H_g=60977472 u_a=+2.500000e-001 u_g=+2.500000e-001 W_a=308200 Goal Δ =+7.665929e-001 actual Δ =+1.176435e+000 u_a/u_g=+1.000000e+000
```

运行 GC 追踪可以告诉你很多关于程序健康状态以及回收器步调的事情。回收器运行的步调在回收过程中起了重要作用。

## 步调

回收器具有确定何时开始收集的步调算法。算法依赖于回收器用于收集有关正在运行的应用的信息以及应用在堆上分配的压力的反馈循环。压力可以被定义为在指定时间范围内应用分配堆内存的速度。正是压力决定了回收器需要运行的速度。

在回收器开始回收之前，它会计算完成回收所需的时间。然后一旦回收运行，将会对正在运行的应用程序上造成延迟，这将让应用程序的工作变慢。每次回收都会增加应用程序的整体延迟。

一种误解是认为降低回收器步调是改善性能的一种方法。这个想法是，如果你能延缓下次回收的开始，那么你也能延缓它所造成的延时。对回收器友好并不是要降慢其步调。

你可以决定改变 GC 百分比的值使其超过 100。这会在下次回收开始前增加分配的堆内存的大小，从而导致回收的步调降低，不要考虑做这种事。

### 图 14

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure14.png)

图 14 展示了改变 GC 百分比会如何改变下次回收开始前允许分配的堆内存大小。你可以想象回收器如何因为等待更多的堆内存被使用而变慢。

尝试直接影响回收器的步调对友好对待回收器并无帮助。如果确实希望在每次回收之间或回收期间完成更多的工作，可以减少任意工作添加到堆内存的分配数量或次数。

*注意：这个想法也是为了用尽可能小的堆来实现所需的吞吐量。请记住，在云环境中运行时，最小化堆内存等资源的使用非常重要。*

### 图 15

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure15.png)

清单 15 显示了将在本系列 ( 博文 ) 的下一篇文章中所使用的 Go 应用程序运行的一些统计信息。蓝色版本显示没经过任何优化的应用程序在处理 10 K 请求时的统计信息。绿色版本显示了发现并去掉应用程序 4.48 GB 的非生产性的内存分配后，处理相同的 10 k 请求的统计信息。

看这两个版本的平均收集速度（2.08 ms vs 1.96 ms），它们几乎相同，约为 2.0 ms。这两个版本之间的根本差异是每次回收之间的工作量，从每次回收处理 3.98 增加到 7.13 个请求，以同样的速度完成的工作量增加了 79.1 ％。正如你所看到的，回收并没有随着这些分配的减少而减慢，而是保持不变，（绿色版本的）胜出之处是因为每次回收之间完成了更多工作。

调整回收的步调以延缓其延迟花费并不是你提高应用程序性能的方式。减少回收器运行所需的时间，这反过来就会减少造成的延迟成本。虽然已经对回收器造成的延迟花费进行了解释，但为了清楚起见，让我再总结一下。

## 回收器延时消耗

运行应用中每次回收有两种类型的延时。第一种是窃取（stealing） CPU 的处理能力。窃取 CPU 处理能力的影响是你的应用在回收过程中不能以全力状态运行。因为应用的 Goruinte 正在和回收器的 Goroutine 共享 P，或者正在帮助回收 ( 辅助标记 )。

### 图 16

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure16.png)

图 16 展示了应用如何仅仅使用 CPU 处理能力的 75% 去工作。这是因为回收器为了回收占用了 P1。

### 图 17

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure17.png)

图 17 展示了在这个时刻（通常只有几微秒）应用如何只使用一半的 CPU 处理能力为应用工作。这是因为在 P3 上的 Goroutine 正在进行辅助标记，并且回收器为自己设置了专用的 P1。

*注意：标记通常需要 4 个 CPU- 毫秒（CPU-millseconds）处理每 MB 存活的堆 (e.g. 为了评估标记需要运行多少毫秒，用存活的堆大小 MB 然后除以 0.25 乘上 CPU 个数 )。标记实际以 1 MB/ms 运行，但是因为只用了 1/4 的 CPU（译者注：所以是 4 ms 处理 1 MB，也就是开头的 4 个 CPU- 毫秒每 MB）*

第二个延时取决于在回收过程中出现的 STW 延迟出现的次数。STW 时间是没有应用程序 Goroutine 执行任意应用程序工作的时间。该应用程序基本上停止了。

### 图 18

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure18.png)

图 18 展示了 STW 延时，这个时候所有 Goroutine 都会停止，这会在每次回收发生两次。如果你的应用健康，回收器可以保持大部分回收过程的总 STW 时间在 100 微秒之内。

你现在已经知道回收器的不同时期，内存如何分配，步调器如何工作，以及回收器在你运行应用中主要出现的不同延时。通过这些知识，你如何对回收器友好的问题终于能解决了。

## 对回收器友好

对回收器表示友好就是降低内存压力。请记住，压力定义为应用在指定时间内分配内存的速度。当压力降低时，因回收器主要引发的延迟就会降低。而 GC 延迟会拖慢你的应用。

能够降低 GC 延迟的方式是，从应用中辨别和去掉不需要的内存分配。可以通过以下几种方式帮助回收器。

帮助回收器：

- 尽可能维护最小化的堆
- 找到最佳的一致步调
- 每次回收保持在目标之内
- 最小化每次回收，STW 以及辅助标记的持续时长

以上所列都能帮助降低在你运行中的程序，主要因回收器造成的延迟大小。这会改善应用的吞吐量表现。我们不需要回收器的步调做任何处理，下面是你可以做的其他事情，以帮助做出更好的工程决策，减少堆上的压力。

### 了解应用程序执行的工作负载的性质

了解工作负载意味着确保使用合理数量的 Goroutine 来完成你的工作。CPU 密集型与 IO 密集型的工作负载不同，需要不同的工程决策。

https://studygolang.com/articles/17014

### 了解已定义的数据及其在应用程序中的传递方式

了解数据意味着了解你尝试解决的问题。数据语义一致性是维护数据完整性的关键部分，并且允许你在堆栈上选择堆分配时（通过读取代码）知道这件事。

https://studygolang.com/articles/12487

## 结论

作为 Go 开发者，如果你花时间专注于减少分配，你正在对垃圾回收器表示友好。你不能编写零分配的应用程序，因此重要的是要认识到有效的分配（对应用有助）和无生产力的分配（对应用有害）之间的差异。然后信任垃圾回收器，相信它能保持堆处于健康状态，并使你的应用程序始终如一地运行。

拥有垃圾回收器是一笔很划算的交易，我花费垃圾回收的成本，因而没有内存管理的负担。Go 允许你作为开发人员提高工作效率的同时还可以编写足够快的应用程序。垃圾回收器对实现这一目标起了重要作用。在下一篇文章中，我将向你展示一个示例 Web 应用程序以及如何动手使用工具查看所有这些信息。

---

via: https://www.ardanlabs.com/blog/2018/12/garbage-collection-in-go-part1-semantics.html

作者：[William Kennedy](https://www.ardanlabs.com/)
译者：[LSivan](https://github.com/LSivan)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
