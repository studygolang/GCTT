首发于：https://studygolang.com/articles/24562

# Go 垃圾回收：第三部分 - GC 的步调

## 前言

这是三篇系列文章中的第三篇。该系列文章提供了一种对 Go 垃圾回收背后的机制和概念的理解。本篇的主要内容是 GC 如何控制自己的步调。

三篇文章的索引：<br>
1）[Go 垃圾回收：第一部分 - 概念](https://www.ardanlabs.com/blog/2018/12/garbage-collection-in-go-part1-semantics.html)<br>
2）[Go 垃圾回收：第二部分 - GC 追踪](https://www.ardanlabs.com/blog/2019/05/garbage-collection-in-go-part2-gctraces.html)<br>
3）[Go 垃圾回收：第三部分 - GC 的步调](https://www.ardanlabs.com/blog/2019/07/garbage-collection-in-go-part3-gcpacing.html)<br>

## 简介

在第二篇文章里，我向你展示了垃圾回收器的行为以及如何使用工具查看回收器给你的运行程序带来的延迟。我带着你运行了一个真实的 web 应用并向你展示了如何生成 GC 追踪和应用性能分析。接着我还向你展示了如何解读这些工具的输出，以便你找到提高应用程序性能的方法。

第二篇文章的结论与第一篇相同：如果你减少了堆上的压力，就可以降低延迟成本，从而提高应用程序的性能。对回收器友好的最佳策略就是减少每个任务所需内存分配的次数或大小。本篇中，我将向你展示步调算法是如何持续找出给定工作压力下的最优步调的。

## 并发示例代码

我将使用位于该链接处的代码：https://github.com/ardanlabs/gotraining/tree/master/topics/go/profiling/trace

这个程序将找出某个特定主题在一组 RSS 新闻摘要文档中的出现频率。程序中包含了不同版本的查找算法以测试不同的并发模式。我将主要关注 `freq`，`freqConcurrent` 和 `freqNumCPU` 这几个版本。

*注：我是在一台 Macbook Pro 上使用 go1.12.7 运行的代码，该机配备了一个具有 12 个硬件线程的 Intel i9 处理器。在不同的体系架构、操作系统和 Go 版本下，你会看到不同的结果，但本篇的核心结论应该保持不变。*

我首先从 `freq` 版本开始。它表示这个程序的非并发串行版本，并将为下文的并发版本提供一个基准。

**清单 1**
```go
01 func freq(topic string, docs []string) int {
02     var found int
03
04     for _, doc := range docs {
05         file := fmt.Sprintf("%s.xml", doc[:8])
06         f, err := os.OpenFile(file, os.O_RDONLY, 0)
07         if err != nil {
08             log.Printf("Opening Document [%s] : ERROR : %v", doc, err)
09             return 0
10         }
11         defer f.Close()
12
13         data, err := ioutil.ReadAll(f)
14         if err != nil {
15             log.Printf("Reading Document [%s] : ERROR : %v", doc, err)
16             return 0
17         }
18
19         var d document
20         if err := xml.Unmarshal(data, &d); err != nil {
21             log.Printf("Decoding Document [%s] : ERROR : %v", doc, err)
22             return 0
23         }
24
25         for _, item := range d.Channel.Items {
26             if strings.Contains(item.Title, topic) {
27                 found++
28                 continue
29             }
30
31             if strings.Contains(item.Description, topic) {
32                 found++
33             }
34        }
35     }
36
37     return found
38 }
```

清单 1 展示的是 `freq` 函数。这个串行版本遍历一个文件名集合并执行 4 种操作：打开文件、读取文件、解码和检索。它对每个文件执行上述操作，一次一个。

在我的机器上运行这个版本的 `freq` 时，我得到了如下的结果：

**清单 2**
```c
$ time ./trace
2019/07/02 13:40:49 Searching 4000 files, found president 28000 times.
./trace  2.54s user 0.12s system 105% cpu 2.512 total
```

从 `time` 的输出中可以看到，程序处理 4000 个文件用了大约 2.5 秒。能看到垃圾回收在其中所占的比例当然是很好的，你可以通过查看程序的追踪结果做到这点。由于这是一个启动并完成的程序，你可以使用 `trace` 包生成一个追踪。

**清单 3**
```go
03 import "runtime/trace"
04
05 func main() {
06     trace.Start(os.Stdout)
07     defer trace.Stop()
```

清单 3 展示的是从程序生成追踪所需的代码。从标准库的 `runtime` 文件夹里导入 `trace` 包后，调用 `trace.Start` 和 `trace.Stop`。将追踪输出定向到 `os.Stdout` 只是简化了代码。

有了这段代码，现在你可以重新编译和运行这个程序了。不要忘了把标准输出重定向到一个文件。

**清单 4**
```c
$ go build
$ time ./trace > t.out
Searching 4000 files, found president 28000 times.
./trace > t.out  2.67s user 0.13s system 106% cpu 2.626 total
```

运行时间增加了 100 毫秒多一点，但这是意料之中的。追踪捕捉了每次函数调用，进入和退出，直到微秒精度。重要的是现在有了一个名为 `t.out` 的文件，里面有追踪数据。

为了查看追踪结果，需要使用追踪工具来运行追踪数据。

**清单 5**
```c
$ go tool trace t.out
```

执行以上命令会启动 Chrome 浏览器并显示以下内容：

*注：追踪工具使用了 Chrome 浏览器内置的工具，所以它只能在 Chrome 里工作。*

**图 1**

![图 1](https://github.com/studygolang/gctt-images/blob/master/Garbage-Collection-In-Go-Part-3-GC-Pacing/103_figure1.png?raw=true)

图 1 展示的是追踪工具启动时显示的 9 个链接。当下最重要的是第一个标着 `View trace` 的链接。点击之后，你就会看到与下图类似的画面：

**图 2**

![图 2](https://github.com/studygolang/gctt-images/blob/master/Garbage-Collection-In-Go-Part-3-GC-Pacing/103_figure2.png?raw=true)

图 2 展示的是在我的机器上运行程序时的完整追踪窗口。本篇中，我会重点关注与垃圾回收器相关的部分，即标着 `Heap` 的第二部分和标着 `GC` 的第四部分。

**图 3**

![图 3](https://github.com/studygolang/gctt-images/blob/master/Garbage-Collection-In-Go-Part-3-GC-Pacing/103_figure3.png?raw=true)

图 3 更详细地展示了追踪的前 200 毫秒。注意观察 `Heap`（绿色和橙色区域）和 `GC`（底部的蓝线）。`Heap` 部分向你展示了两个信息：橙色区域代表在每个微秒时刻对应的堆上正在使用的空间，绿色代表触发下次回收的堆使用量。这就是为什么每当橙色区域到达绿色区域的顶端就会发生一次垃圾回收的原因。蓝线表示一次垃圾回收。

在这个版本的程序整个运行过程中，堆上内存使用量一直保持在大约 4 MB。要想查看每次垃圾回收时的统计数据，你可以使用选择工具在所有蓝线周围绘制一个框。

**图 4**

![图 4](https://github.com/studygolang/gctt-images/blob/master/Garbage-Collection-In-Go-Part-3-GC-Pacing/103_figure4.png?raw=true)

图 4 展示的是如何使用箭头工具在蓝线周围绘制蓝框。你应该想要框住每一条线。框中的数字表示选中的项目消耗的总时间。上图中，有接近 316 毫秒（ms，μs，ns）的区域被选中。当所有的蓝色线条被选中时，就得到了如下的统计结果：

**图 5**

![图 5](https://github.com/studygolang/gctt-images/blob/master/Garbage-Collection-In-Go-Part-3-GC-Pacing/103_figure5.png?raw=true)

图 5 显示所有的蓝线都在 15.911 毫秒标记到 2.596 秒标记之间。一共有 232 次垃圾回收，共消耗 64.524 毫秒，平均每次消耗 287.121 微秒。而程序运行需要 2.626 秒，这就意味着垃圾回收只占总运行时间的 2%。基本上，垃圾回收是运行该程序的一个微不足道的成本。

有了这个基准，可以使用并发算法完成相同的工作，以期加快程序的速度。

**清单 6**
```go
01 func freqConcurrent(topic string, docs []string) int {
02     var found int32
03
04     g := len(docs)
05     var wg sync.WaitGroup
06     wg.Add(g)
07
08     for _, doc := range docs {
09         go func(doc string) {
10             var lFound int32
11             defer func() {
12                 atomic.AddInt32(&found, lFound)
13                 wg.Done()
14             }()
15
16             file := fmt.Sprintf("%s.xml", doc[:8])
17             f, err := os.OpenFile(file, os.O_RDONLY, 0)
18             if err != nil {
19                 log.Printf("Opening Document [%s] : ERROR : %v", doc, err)
20                 return
21             }
22             defer f.Close()
23
24             data, err := ioutil.ReadAll(f)
25             if err != nil {
26                 log.Printf("Reading Document [%s] : ERROR : %v", doc, err)
27                 return
28             }
29
30             var d document
31             if err := xml.Unmarshal(data, &d); err != nil {
32                 log.Printf("Decoding Document [%s] : ERROR : %v", doc, err)
33                 return
34             }
35
36             for _, item := range d.Channel.Items {
37                 if strings.Contains(item.Title, topic) {
38                     lFound++
39                     continue
40                 }
41
42                 if strings.Contains(item.Description, topic) {
43                     lFound++
44                 }
45             }
46         }(doc)
47     }
48
49     wg.Wait()
50     return int(found)
51 }
```

清单 6 展示的是 `freq` 的一个可能的并发版本。这个版本的核心设计模式是使用扇出模式。对于 `docs` 集合中的每个文件，都会创建一个 goroutine 来处理。如果有 4000 个文档要处理，就要使用 4000 个 goroutine。这个算法的优点就是它是利用并发性的最简单方法。每个 goroutine 处理一个且仅处理一个文件。可以使用 `WaitGroup` 执行等待处理每个文档的编排，并且原子指令可以使计数器保持同步。

这个算法的缺点在于它不能很好地适应文档或 CPU 核心的数量。所有的 goroutine 在程序启动的时候就开始运行，这意味着很快就会消耗大量内存。由于在第 12 行使用了 `found` 变量，它还导致了缓存一致性的问题。由于每个核心共享这个变量的高速缓存行，这将导致内存抖动。随着文件或核心数量的增加，这会变得更糟。

有了代码，现在你可以重新编译和运行这个程序了。

**清单 7**
```c
$ go build
$ time ./trace > t.out
Searching 4000 files, found president 28000 times.
./trace > t.out  6.49s user 2.46s system 941% cpu 0.951 total
```

从清单 7 中的输出可以看到，现在程序花了 951 毫秒来处理相同的 4000 个文件。这是大约 64% 的性能提升。看一下追踪结果。

**图 6**

![图 6](https://github.com/studygolang/gctt-images/blob/master/Garbage-Collection-In-Go-Part-3-GC-Pacing/103_figure6.png?raw=true)

图 6 展示了这个版本的程序在我的机器上运行时是怎样占用了比之前多得多的 CPU。图中的开始部分有很多密集的线条，这是因为所有的 goroutine 都被创建之后，它们运行并且开始尝试从堆上分配内存。很快的，前 4 MB 内存被分配了，紧接着就有一个 GC 启动了。在这次 GC 当中，每个 goroutine 都有时间运行，大部分由于在堆上申请内存而被置于等待状态。至少有 9 个 goroutine 继续运行并在 GC 结束时将堆增长到大约 26 MB。

**图 7**

![图 7](https://github.com/studygolang/gctt-images/blob/master/Garbage-Collection-In-Go-Part-3-GC-Pacing/103_figure7.png?raw=true)

图 7 展示了在首次 GC 的大部分时间里有很多 goroutine 处于已就绪（Runnable）和运行中（Running）的状态以及这一幕是如何再次快速发生的。请注意，堆性能概要（profile）看起来不规则并且垃圾回收也不像之前那样有规律。如果你仔细看，就会发现第二次 GC 几乎是在第一次 GC 结束之后就立即开始了。

如果选中了所有的垃圾回收，你就能看到下面的一幕：

**图 8**

![图 8](https://github.com/studygolang/gctt-images/blob/master/Garbage-Collection-In-Go-Part-3-GC-Pacing/103_figure8.png?raw=true)

图 8 显示图中所有的蓝线都在 4.828 毫秒标记到 906.939 毫秒标记之间。一共有 23 次垃圾回收，共占用 284.447 毫秒，平均每次占用 12.367 毫秒。知道程序运行需要 951 毫秒，垃圾回收在整个运行时间里占用了大约 34%。

这与串行版本在性能和 GC 时间上有显著差异。并行运行更多的 goroutine 让工作时间缩短了大约 64%。代价就是大幅增加了机器上各种资源的占用。不幸的是，堆上内存的占用最高达到了大约 200 MB。

有了这个并发的基准，下一个并发算法试图更有效率的使用资源。

**清单 8**
```go
01 func freqNumCPU(topic string, docs []string) int {
02     var found int32
03
04     g := runtime.NumCPU()
05     var wg sync.WaitGroup
06     wg.Add(g)
07
08     ch := make(chan string, g)
09
10     for i := 0; i < g; i++ {
11         go func() {
12             var lFound int32
13             defer func() {
14                 atomic.AddInt32(&found, lFound)
15                 wg.Done()
16             }()
17
18             for doc := range ch {
19                 file := fmt.Sprintf("%s.xml", doc[:8])
20                 f, err := os.OpenFile(file, os.O_RDONLY, 0)
21                 if err != nil {
22                     log.Printf("Opening Document [%s] : ERROR : %v", doc, err)
23                     return
24                 }
25
26                 data, err := ioutil.ReadAll(f)
27                 if err != nil {
28                     f.Close()
29                     log.Printf("Reading Document [%s] : ERROR : %v", doc, err)
23                     return
24                 }
25                 f.Close()
26
27                 var d document
28                 if err := xml.Unmarshal(data, &d); err != nil {
29                     log.Printf("Decoding Document [%s] : ERROR : %v", doc, err)
30                     return
31                 }
32
33                 for _, item := range d.Channel.Items {
34                     if strings.Contains(item.Title, topic) {
35                         lFound++
36                         continue
37                     }
38
39                     if strings.Contains(item.Description, topic) {
40                         lFound++
41                     }
42                 }
43             }
44         }()
45     }
46
47     for _, doc := range docs {
48         ch <- doc
49     }
50     close(ch)
51
52     wg.Wait()
53     return int(found)
54 }
```

清单 8 展示的是 `freqNumCPU` 版本的程序。该版本的核心设计模式是池模式。由一个基于逻辑处理器数目的 goroutine 池来处理所有的文件。如果有 12 个逻辑处理器可以使用，就使用 12 个 goroutine。这个算法的优点是它从头到尾保持了资源占用的一致性。由于使用了固定数量的 goroutine，所以只需要分配那 12 个 goroutine 运行所需的内存。这也解决了高速缓存一致性问题导致的内存抖动。这是因为第 14 行调用的原子指令只会发生一个很小的固定次数。

该算法的缺点就是它更复杂了。它额外使用了一个 channel 来给 goroutine 池分发工作。要想在每次使用池模式时都为池子指定一个“正确”数量的 goroutine 是非常复杂的。作为一个通常的做法，我为池子启动了与逻辑处理器相同数量的 goroutine。然后通过压力测试或者使用生产指标，可以计算出一个最终的池子大小。

有了代码，现在你可以重新编译和运行这个程序了。

**清单 9**
```c
$ go build
$ time ./trace > t.out
Searching 4000 files, found president 28000 times.
./trace > t.out  6.22s user 0.64s system 909% cpu 0.754 total
```

从清单 9 中的输出可以看到，现在程序花了 754 毫秒来处理相同的 4000 个文件。程序快了大约 200 毫秒，对于这样一个小的工作量而言，这是一个很大的提升。看一下追踪结果。

**图 9**

![图 9](https://github.com/studygolang/gctt-images/blob/master/Garbage-Collection-In-Go-Part-3-GC-Pacing/103_figure9.png?raw=true)

图 9 显示了这个版本的程序运行时也同样使用了我机器上所有的 CPU。仔细看的话，你就会发现一个固定的节奏在这个程序里又出现了，跟串行版本很像。

**图 10**

![图 10](https://github.com/studygolang/gctt-images/blob/master/Garbage-Collection-In-Go-Part-3-GC-Pacing/103_figure10.png?raw=true)

图 10 展示的是程序运行前 20 毫秒的核心指标的一个特写。与串行版本相比，垃圾回收肯定占用了更长的时间，但是有 12 个 goroutine 同时在运行。使用的堆内存在整个程序的运行当中一直保持在 4 MB 左右，这又与串行版本一样了。

如果选中了所有的垃圾回收，你就能看到下面的一幕：

**图 11**

![图 11](https://github.com/studygolang/gctt-images/blob/master/Garbage-Collection-In-Go-Part-3-GC-Pacing/103_figure11.png?raw=true)

图 11 显示图中所有的蓝线都位于 3.055 毫秒标记到 719.928 毫秒标记之间。共有 467 次垃圾回收，耗时 177.709 毫秒，平均每次花费 380.535 微秒。知道程序运行需要 754 毫秒，这意味着垃圾回收在总的运行时间里占了大约 25%，相比另一个并发版本降低了 9%。

这个版本的并发算法看起来可以适应更多的文件和核心。我认为它增加的复杂性成本是值得的。可以使用列表切片来代替 channel 将每个 goroutine 的任务放到一个桶里。这必然会增加复杂度，尽管它可以减少由 channel 带来的延迟成本。虽然随着文件和核心的增多，这个收益可能会不容忽视，但是它带来的复杂度成本还是需要衡量的。这是你可以自己尝试的东西。

## 结语

我喜欢比较算法的三个版本就是要看看 GC 如何处理每种情况。处理文件所需的内存总量并不随程序版本而变化，不同的是程序如何分配内存。

当只有一个 goroutine 时，只需要最起码的 4 MB 堆内存。当程序一次性分派了所有的工作，GC 采取的策略是使堆增长，减少收集的次数，但是每次运行更长的时间。当程序限制了并行处理的文件数量时，GC 又采取了保持一个小堆的策略，增加回收的次数，但是每次运行更少的时间。基本上，GC 采取的每个策略都是为了使 GC 对程序运行的影响最小。

```text
| Algorithm  | Program | GC Time  | % Of GC | # of GC’s | Avg GC   | Max Heap |
|------------|---------|----------|---------|-----------|----------|----------|
| freq       | 2626 ms |  64.5 ms |     ~2% |       232 |   278 μs |    4 meg |
| concurrent |  951 ms | 284.4 ms |    ~34% |        23 |  12.3 ms |  200 meg |
| numCPU     |  754 ms | 177.7 ms |    ~25% |       467 | 380.5 μs |    4 meg |
```

就两个并发版本而言，`freqNumCPU` 版本带来的额外好处是更好的处理高速缓存一致性的问题，这很有帮助。然而，每个程序在 GC 上花费的总时间差别不大，大约 284.4 毫秒对大约 177.7 毫秒。有时在我的机器上运行这些程序，这些数字会更加接近。使用 go 1.13.beta1 版本做了一些实验，我甚至看到这两个算法运行了相同的时间。这可能意味着新版的 go 有一些改进能够让 GC 更好的预测如何运行。

所有这些都让我有信心在程序运行时抛出大量任务。一个例子就是一个使用 50000 goroutine 的 web 服务，这就是一个本质上与第一个并发算法类似的扇出模式。GC 会研究工作量并为服务找到最优的步调来避开它。至少对我而言，使用 GC 是值得的，因为不用去考虑所有这些。

---

via: https://www.ardanlabs.com/blog/2019/07/garbage-collection-in-go-part3-gcpacing.html

作者：[William Kennedy](https://github.com/ardanlabs)
译者：[Stonelgh](https://github.com/stonglgh)
校对：[DingdingZhou](https://github.com/DingdingZhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

