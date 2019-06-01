# Go中的垃圾收集：第一部分-语义

这是三部分系列文章的第一篇博文，提供了Go中垃圾收集背后的机制和语义的理解。这篇博文主要介绍收集器语义的基础。

三部分系列文章的索引：

1) [Go中的垃圾收集：第一部分-语义](https://www.ardanlabs.com/blog/2018/12/garbage-collection-in-go-part1-semantics.html)
2) [Go中的垃圾收集：第二部分-GC追踪](https://www.ardanlabs.com/blog/2019/05/garbage-collection-in-go-part2-gctraces.html)
3) 即将诞生

## 简介

垃圾收集器拥有跟踪堆内存分配，释放不再需要的分配内存以及维护在用分配内存的责任。一门语言如何设计去实现这些行为是很复杂的，但它不应该成为应用开发者为了构建软件而去理解细节的要求。而且，对于语言的VM和运行时的不同发行版，这些系统的实现一直都在变化和发展中。对于应用程序开发人员来说，重要的是保持一个良好的工作模型，了解垃圾收集器对其语言的行为以及如何在不关心实现的情况下对这种行为表示同情。

在1.12版本，Go编程语言使用了无分代同步三色标记清除收集器。如果你想要形象化地了解标记清除收集器如何工作，Ken Fox写了这篇[好文章](https://spin.atomicobject.com/2014/09/03/visualizing-garbage-collection-algorithms)并提供了动画。Go收集器的实现随着每一个发行版而变化和发展，所以一旦下一版本发行，任意谈及其实现细节的博文将不再准确。

尽管如此，我将在本文中做的建模不会关注实际的实现细节。建模将关注你会经历的和你应该在未来几年看到的行为。在这篇文章中，我将和你分享收集器的行为，并解释如何对该行为表示同情，无论当前实现或未来如何变化。这将使您成为更好的Go开发人员。

*注意：这里你可以对有关[垃圾收集器](https://github.com/ardanlabs/gotraining/tree/master/reading#garbage-collection)以及Go实际的收集器进行扩展阅读*

## 堆不是一个容器

我永远不会将堆称为可以存储或释放值的容器。重要的是，要理解定义“堆”是没有线性限制内存的，认为为进程空间中的应用程序使用而保留的任何内存都可用于堆内存分配，虚拟或物理存储任何给定的堆内存分配与我们的模型无关。这种理解将帮助您更好地了解垃圾收集器的工作原理。

## 收集器行为

当某次回收开始，收集器经历三个阶段的工作。其中两个阶段是引起Stop The World(STW)的延迟，另外的阶段会产生降低程序吞吐量的延迟。这三个阶段为：

- 标记开始 - STW
- 标记中 - 并发
- 标记结束 - STW

以下为每一个阶段的细分

### 标记开始 - STW

当回收开始，首要执行的动作是打开写屏障。写屏障的目的是允许收集器在收集过程保持堆上的数据完整性，因为收集器和应用程序的goroutine会并发执行。

为了打开写屏障，应用的每个goroutine都必须停止运行。这个动作通常非常快，平均在10~30微秒之间。这是指，如果应用程序的goroutine表现正常情况下。

*注意：为了更好理解这些调度图，请务必阅读[Go Scheduler](https://www.ardanlabs.com/blog/2018/08/scheduling-in-go-part1.html)上的系列文章*

### 图一

![figure1](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure1.png)

图1展示了回收之前应用有4个goroutine在运行。这4个goroutine都应该被停掉。收集器的唯一方法就是观察和等待每个goroutine进行函数调用。函数调用保证了goroutine在一个安全的点上被停掉。如果其中一个goroutine没有进行函数调用但其他的却做了函数调用，这会发生什么呢？

### 图2

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure2.png)

Figure 2 shows a real problem. The collection can’t start until the goroutine running on P4 is stopped and that can’t happen because it’s in a [tight loop](https://github.com/golang/go/issues/10958) performing some math.

图2展示了一个真正的问题。在P4上运行的goroutine停下来之前，回收都不会进行。然后这是不会发生的，因为它正在[紧密循环](https://github.com/golang/go/issues/10958)进行某些数学运算。

### 清单1

```go
func add(numbers []int) int {
     var v int
     for _, n := range numbers {
         v += n
     }
     return v
}
```

清单1展示了运行在P4上的Goroutine正在执行的代码。根据切片的大小，goroutine可能以不合理的大量时间运行从而无法停止。这种代码会延缓回收启动。更糟糕的是当收集器等待着时，其他P不能为其他任意的goroutine服务。gorotine在一个合理的时间范围内进行函数调用显得极其重要。

*注意:这是语言团队想要在1.14通过加入[preemotive](https://github.com/golang/go/issues/24543)技术到调度中去修正的问题*

### 标记中-并发

一旦开启了写障碍，收集器开始标记阶段。收集器做的第一件事是占用自身CPU可用处理能力的25%。收集器使用gorouitne去做收集工作，也同样需要应用程序的goroutine使用的P和M（译者注：从此处开始作者将G划分了两类，一类是应用程序用于自身工作的gourinte，下文称应用goroutine，一类是用于GC的goroutine，这样会更好理解）。这意味着对于我们四个线程的go程序，有一个完整的P将专门用来进行回收工作。

### 图三

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure3.png)

图三展示了收集器在回收过程中如何为自身占有P1。现在收集器开始标记阶段了。标记阶段标记堆内存中仍在使用的值。这个工作先检查栈内所有存活的gorouitne，去寻找堆内存的根指针。然后收集器必须从那些根指针遍历堆内存图。当标记工作发生在P1上，应用程序可以继续在P2,P3和P4上同步工作。这意味着收集器的影响被最小化到当前CPU处理能力的25%。

我希望这个事就这样完了然而并没有。如果在收集过程中确定了在P1上专用于GC的goroutine在使用中的堆内存达到极限之前无法完成标记工作，该怎么办？如果3个goroutine中只有一个进行的应用工作导致收集器无法及时完成(标记工作)又怎么办？(译者注：此处的意思为内存分配过快)。在这种情况下，新的分配必须放慢速度，特别是从那个(导致标记无法完成的)goroutine。

如果收集器确定它需要减慢分配，它将招募应用goroutine以协助标记工作。这称为辅助标记。任何应用goroutine花费在辅助标记的时间长度与它添加到堆内存中的数据量成正比。辅助标记的一个积极影响是它有助于更快地完成回收。

### 图4

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure4.png)

图4展示了应用goroutine如何进行辅助标记来帮助回收工作的。希望其他的应用goroutine不需要同样参与进来。分配动作较多的应用可以看到大部分运行中的goroutine在回收过程中都进行了少量的辅助标记。

收集器的一个目标是消除对辅助标记的需求。如果任意给定的回收都需要大量的辅助标记而结束，收集器很快就会开始下一次的垃圾回收。对下一次的回收努力去减少辅助标记的数量是必要的。

### 标记结束-STW

一旦标记工作完成，下阶段就是标记结束了。到这个阶段，写屏障会被停止，各样的清洁工作会被执行，然后计算好下一次的回收目标。在标记阶段发现自身处理紧密循环的goroutine也会延长标记结束STW的时长。

### 图5

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure5.png)

图5展示了在标记结束阶段完成时所有的goroutine如何停止的。这个动作通常平均在60到90微秒之间。这个阶段可以不需要STW，但通过使用STW，代码会更简单，小小的收益抵不上增加的复杂度。

一旦回收完成，每个P都能服务于应用goroutine，然后应用回到全力运行状态。
### 图6

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure6.png)

图6展示了一旦回收完成所有可选的P如何再次处理应用的工作。应用回到回收开始前的全力运行状态。

### 并发清除

在回收完成之后有另一个叫清除的动作发生。清除是指回收堆内存中未标记为使用中的值所关联的内存。该动作会发生在应用程序goroutine尝试在堆内存中分配新值的时候。清除的延迟被添加到在堆内存中执行分配的成本中，与垃圾收集相关的任何延迟无关。

下面是我机器上的追踪样本，有12条硬件线程可用于执行gorouitne。

### 图7

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure7.png)

图7展示了追踪的部分快照。你可以看到在回收过程中(盯着顶部的蓝色GC线)，12个P中的其中3个如何专门用于GC。你可以看到goroutine2450，1978和2696在这段时间进行了数次辅助标记而不是执行应用的工作。在回收的最后，只有一个P用于GC并最终执行STW(标记结束)的工作。

在回收完成后，应用程序回到全力运行状态。此外你看到在goroutine下面有一些玫瑰色的线条。

### 图8

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure8.png)

图8展示了那些玫瑰色的线条如何代表goroutine不执行应用的工作而进行清除工作的时刻。这些都是goroutine尝试分配新值到堆内存的时刻。

### 图9

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure9.png)

图9展示了其中一个进行清除动作的gorouitne最后的栈跟踪情况。`runtime.mallocgc`的调用会导致在堆内存分配新值。`runtime.(*mcache).nextFree`的调用引起清除动作。一旦堆内存上没有更多的已分配内存需要回收，就不会再看到`nextFree`的调用。

刚刚描述的回收动作仅仅在回收过程开始和进行中才会发生。在回收开始时，配置项GC百分比扮演了重要角色

## GC百分比

运行过程中有一个配置项叫GC百分比，默认值设置为100。这个值代表了在下次回收开始前能分配多大的堆内存。设置GC百分比为100意味着，根据回收完成后标记为存活的堆内存量，下一次回收必须在堆内存上添加100％以上的新分配(内存)才启动。

举个例子，想象某次回收完成后堆内存有2MB在使用中。(译者注：后半句话应该是分配2MB后GC才会开始，作者省了。。。)

*注意:在这篇博文中堆内存的图不代表使用Go的时候的真实情况。Go中的堆内存通常是碎片化和混乱的，而且没有图像所代表的清晰分离。这些图在更为易于理解的方式上提供可视化堆内存的方法，这种方式对于你将体验的行为是准确的。*

### 图10

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure10.png)

图10展示了最后的回收完成后在使用中的2MB的堆内存。因为GC百分比设置为100%，下一次回收需要在添加到2MB的堆内存时才开始，或者在超过2MB之前开始。

### 图11

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure11.png)

图11显示了超过2MB堆内存正在使用。这会触发回收。查看该动作所有(细节)的方法是为每次回收生成GC追踪。

## GC追踪

GC追踪可以通过在运行任意Go应用时包含环境变量`GODEBUG`并指定`gctracec=1`来生成。每次回收发生，运行时会将GC追踪信息写到`stderr`中。

### 清单2

```bash
GODEBUG=gctrace=1 ./app

gc 1405 @6.068s 11%: 0.058+1.2+0.083 ms clock, 0.70+2.5/1.5/0+0.99 ms cpu, 7->11->6 MB, 10 MB goal, 12 P

gc 1406 @6.070s 11%: 0.051+1.8+0.076 ms clock, 0.61+2.0/2.5/0+0.91 ms cpu, 8->11->6 MB, 13 MB goal, 12 P

gc 1407 @6.073s 11%: 0.052+1.8+0.20 ms clock, 0.62+1.5/2.2/0+2.4 ms cpu, 8->14->8 MB, 13 MB goal, 12 P
```

清单2展示了如何使用`GODEBUG`变量来生成GC追踪。清单也展示了在运行Go应用生成的3份追踪信息。

以下是通过查看清单中的第一个GC追踪线来细分GC追踪中每个值的含义。

### 清单3

```bash
gc 1405 @6.068s 11%: 0.058+1.2+0.083 ms clock, 0.70+2.5/1.5/0+0.99 ms cpu, 7->11->6 MB, 10 MB goal, 12 P

// General
gc 1404     : 自程序启动以来，1404的GC运行(译者注：此处应当是笔误，联系上文其实是1405)
@6.068s     : 自程序启动至此总共6s
11%         : 到目前为止，可用CPU的11%被用于GC
// Wall-Clock
0.058ms     : STW        : 标记开始，开启写障碍
1.2ms       : 并发				: 标记中
0.083ms     : STW        : 标记结束 - 关闭写障碍并清除

// CPU Time
0.70ms      : STW        : 标记开始
2.5ms       : 并发			  : 辅助标记时间(GC按照分配执行)
1.5ms       : 并发 				: 标记 - 后台GC时间
0ms         : 并发			  : 标记 - 空闲GC时间
0.99ms      : STW        : 标记结束

// Memory
7MB         : 标记开始前使用中的堆内存
11MB        : 标记完成后使用中的堆内存
6MB         : 标记完成后被标记为存活的堆内存
10MB        : 标记完成后使用中的堆内存收集目标

// Threads
12P         : 用于运行gorouitne 的物理调度器或线程的数量
```

清单3展示了第一条GC追踪线的实际数字所代表的含义，按行细分。我后面会讨论这些值中的大部分，但现在只要关注追踪1405的GC追踪的内存部分。

### 图12

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure12.png)

### 清单4

```bash
// Memory
7MB         : 标记开始前使用中的堆内存
11MB        : 标记完成后使用中的堆内存
6MB         : 标记完成后被标记为存活的堆内存
10MB        : 标记完成后使用中的堆内存收集目标
```

清单4中的GC追踪线想告诉你的是，在标记工作开始前使用中的堆内存大小为7MB。当标记工作完成时，使用中的堆内存大小达到了11MB。这意味着在回收过程中出现了额外的4MB内存分配。在标记工作完成后被标记为存活的堆内存大小为6MB。这意味着在下次回收开始前应用可以增加使用的堆内存到12MB(存活堆大小6MB的100%)。

你可以看到收集器与其目标有1MB的偏差，标记工作完成后正在使用的堆内存量为11MB而不是10MB。没关系，因为目标是根据当前正在使用的堆内存量、标记为存活的堆内存量以及有关在回收运行时将会发生的其他分配的时间计算来计算的。在这种情况下，应用程序做了一些事情，需要在标记之后使用更多的堆内存而不是像预期那样。

如果查看下一个GC跟踪线（1406），你会看到事情在2ms内发生了变化。

### 图13

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure13.png)

### 清单5

```bash
gc 1406 @6.070s 11%: 0.051+1.8+0.076 ms clock, 0.61+2.0/2.5/0+0.91 ms cpu, 8->11->6 MB, 13 MB goal, 12 P

// Memory
8MB         : 标记开始前使用中的堆内存
11MB        : 标记完成后使用中的堆内存
6MB         : 标记开完成后被标记为存活的堆内存
13MB        : 标记完成后的使用堆内存收集目标
```

清单5展示了这次回收如何在前一次回收2ms之后开始了，即便使用中的堆内存仅仅达到了8MB，而所允许的是12MB。这需要特别注意，如果收集器认为早点开始回收会好一点，那么就会提前开始。在这种情况下，它提前开始大概是因为应用在进行大量的分配工作，然后收集器想要减小这次回收的辅助标记的延时大小。

还有两件事要注意。收集器这次在他的目标之内。标记完成后使用中堆内存的大小是11MB而不是13MB，少了2MB。标记完成后标记为存活的堆内存大小一样为6MB。

附注一点，你可以通过增加`gcpacertrace=1`标志从GC追踪获取更多细节，这会让收集器打印更多有关并发调步器的内部状态。

### 清单6

```bash
$ export GODEBUG=gctrace=1,gcpacertrace=1 ./app

Sample output: 样本输出：
gc 5 @0.071s 0%: 0.018+0.46+0.071 ms clock, 0.14+0/0.38/0.14+0.56 ms cpu, 29->29->29 MB, 30 MB goal, 8 P

pacer: sweep done at heap size 29MB; allocated 0MB of spans; swept 3752 pages at +6.183550e-004 pages/byte

pacer: assist ratio=+1.232155e+000 (scan 1 MB in 70->71 MB) workers=2+0

pacer: H_m_prev=30488736 h_t=+2.334071e-001 H_T=37605024 h_a=+1.409842e+000 H_a=73473040 h_g=+1.000000e+000 H_g=60977472 u_a=+2.500000e-001 u_g=+2.500000e-001 W_a=308200 goalΔ=+7.665929e-001 actualΔ=+1.176435e+000 u_a/u_g=+1.000000e+000
```

运行GC追踪可以告诉你很多关于程序健康状态以及收集器速度的事情。收集器正在运行的速度在回收过程中起了重要作用。

## 调步

收集器具有确定何时开始收集的调步算法。算法依赖于收集器用于收集有关正在运行的应用的信息以及应用在堆上分配的压力的反馈循环。压力可以被定义为在指定时间范围内应用分配堆内存的速度。正是压力决定了收集器需要运行的速度。

在收集器开始回收之前，它会计算它认为完成回收所需的时间。然后一旦回收运行，将会对正在运行的应用程序上造成延迟，这将让应用程序的工作变慢。每次回收都会增加应用程序的整体延迟。

一种误解是认为降低收集器速度是改善性能的一种方法。这个想法是，如果你能延缓下次回收的开始，那么你也能延缓它所造成的延时。要同情收集器并不是要降慢其速度。

你可以决定改变GC百分比的值使其超过100。这会在下次回收开始前增加分配的堆内存的大小，导致回收的速度降低。不要考虑做这种事。

### 图14

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure14.png)

图14展示了改变GC百分比会如何改变下次回收开始前允许分配的堆内存大小。你可以想象收集器如何因为等待更多的堆内存被使用而变慢。

尝试直接影响收集器的速度，除了同情收集器之外无需再做其他。真的是希望在每次回收之间或回收期间完成更多的工作，可以通过减少任意工作添加到堆内存的分配数量或次数来影响它。

*注意：这个想法也是为了用尽可能小的堆来实现所需的吞吐量。请记住，在云环境中运行时，最小化堆内存等资源的使用非常重要。*

### 图15

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure15.png)

清单15显示了将在本系列(博文)的下一篇文章中所使用的运行中的Go应用程序的一些统计信息。蓝色版本显示没经过任意优化的应用程序在处理10K请求时的统计信息。绿色版本显示了应用程序4.48GB的非生产性内存分配被发现并移除后，处理相同的10k请求的统计信息。

看这两个版本的平均收集速度（2.08ms vs 1.96ms），它们几乎相同，约为2.0ms。这两个版本之间的根本变化是每次回收之间的工作量。该应用程序从每次回收处理3.98到7.13个请求，以同样的速度完成的工作量增加了79.1％。正如你所看到的，回收并没有随着这些分配的减少而减慢，而是保持不变，(绿色版本的)胜利来自每次回收之间完成了更多工作。

调整回收的速度以延缓其延迟花费并不是你提高应用程序性能的方式。减少收集器运行所需的时间，这反过来就会减少造成的延迟成本。已经对收集器造成的延迟花费进行解释了，但为了清楚起见，让我再次总结一下。

## 收集器延时消耗

运行应用中每次回收有两种类型的延时。第一种是偷取CPU的处理能力。偷取CPU处理能力的影响是你的应用在回收过程中不能以全力状态运行。应用的goruinte正在和收集器的goroutine共享P，或者正在帮助回收(辅助标记)。

### 图16

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure16.png)

图16展示了应用如何只是用CPU处理能力的75%去工作。这是因为收集器为自身占用了P1,这是主要是为了回收。

### 图17

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure17.png)

图17展示了在这个时刻（通常只有几微秒）应用如何只使用一半的CPU处理能力为应用工作。这是因为在P3上的goroutine正在进行辅助标记，并且收集器为自己设置了专用的P1。

*注意：标记通常需要4个CPU毫秒处理每MB存活的堆(e.g.为了评估标记需要运行多少毫秒，用存活的堆大小MB然后除以0.25乘上CPU个数)。标记实际以1MB/ms运行，但只有1/4的CPU。*

第二个延时取决于在回收过程中出现的STW延迟出现的次数。STW时间是没有应用程序goroutine执行任意应用程序工作的时间。该应用程序基本上停止了。

### 图18

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-I-semantics/100_figure18.png)

图18展示了所有goroutine都停止的STW延时。这会在每次回收发生两次。如果你的应用健康，收集器应该能够保持大部分回收过程的总STW时间在100微秒之内。

你现在知道了收集器的不同时期，内存如何分配，调步器如何工作，以及收集器在你运行应用中主要出现的不同延时。通过这些知识，你应该如何同情收集器的问题最终能被解开了。

## 表示同情

同情收集器是降低内存压力。要记住，压力定义为应用在执行时间内分配内存的速度。当压力降低时，因收集器主要引发的延迟就会降低。GC延迟会拖慢了你的应用。

可以降低GC延迟的方式是从你的应用中辨别和移除不需要的内存分配。可以通过以下几种方式帮助收集器。

帮助收集器：

- 尽可能维护最小化的堆
- 找到最佳的一致步调
- 每次回收保持在目标之内
- 最小化每次回收，STW和辅助标记的持续时长

以上所列都能帮助降低在你运行中的程序主要因收集器造成的延迟大小。这会提升你的应用的吞吐量表现。收集器的步调不需要的对它做任何处理。下面是你可以做的其他事情，以帮助做出更好的工程决策，减少堆上的压力。

### 了解应用程序执行的工作负载的性质

了解工作负载意味着确保使用合理数量的goroutine来完成你的工作。CPU密集型与IO密集型的工作负载不同，需要不同的工程决策。

[https://www.ardanlabs.com/blog/2018/12/scheduling-in-go-part3.html](https://www.ardanlabs.com/blog/2018/12/scheduling-in-go-part3.html)

### **了解已定义的数据及其在应用程序中的传递方式**

了解数据意味着了解你尝试解决的问题。数据语义一致性是维护数据完整性的关键部分，并且允许你在堆栈上选择堆分配时知道（通过读取代码）。

[https://www.ardanlabs.com/blog/2017/06/design-philosophy-on-data-and-semantics.html](https://www.ardanlabs.com/blog/2017/06/design-philosophy-on-data-and-semantics.html)

## 结论

如果你花时间专注于减少分配，那么你就像Go开发人员一样，对垃圾收集器表示同情。你不会想编写零分配的应用程序，因此重要的是要认识到有效的分配（帮助应用程序的分配）和那些没有生产力的分配（那些损害应用程序的分配）之间的差异。然后将你的信任交给垃圾收集器中，相信它能保持堆健康和使你的应用程序始终如一地运行。

拥有垃圾收集器是一笔很值得的交易。我会花费垃圾收集的成本，因而没有内存管理的负担。Go允许你作为开发人员提高工作效率，同时可以编写足够快的应用程序。垃圾收集器是实现这一目标的重要组成部分。在下一篇文章中，我将向你展示一个示例Web应用程序以及如何动手使用工具查看所有这些信息。

------

via: https://www.ardanlabs.com/blog/2018/12/garbage-collection-in-go-part1-semantics.html

作者：[William Kennedy]()
译者：[LSivan](https://github.com/LSivan)
校对：[校对者ID]()

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出