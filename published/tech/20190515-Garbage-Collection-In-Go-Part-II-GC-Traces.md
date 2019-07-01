首发于：https://studygolang.com/articles/21570

# Go 垃圾回收：第二部分 - GC 追踪

## 前言

这是三篇系列文章中的第二篇，该系列文章将会提供一个对 Go 垃圾回收器背后的机制和概念的理解。本篇主要介绍如何生成 GC 追踪并解释它们。

三篇系列文章的索引：

1）[Go 垃圾回收：第一部分 - 概念](https://studygolang.com/articles/21569)

2）[Go 垃圾回收：第二部分 - GC 追踪](https://studygolang.com/articles/21570)

3）即将发布

## 介绍

在第一篇文章中，我花了一些时间描述了垃圾回收器的行为并且展示了它对正在运行的应用造成的延迟。我还分享了如何去生成并且解释 GC 追踪，展示了堆中的内存是如何变化的，并且解释了 GC 的不同阶段以及它们是如何影响延迟成本的。

那篇文章得出的最后结论是，降低堆的压力，就会降低延迟成本从而增加应用的性能。我也提出了另外一个观点，通过找到增加任意两次垃圾回收间隔时间的方法，来降低回收器开始的步调不是一个好的策略。一个稳定的步调，即使比较快，也会有利于保持应用的高性能运行

在本篇文章中，我将会带领你运行一个真实的 Web 应用并且向你展示如何生成 GC 追踪和应用的性能概要（profile）文件。然后我会向你展示如何解释这些工具的输出内容，从而找到一个提升应用性能的方法。

## 运行应用

看一下我在 Go 练习中用到的这个 Web 应用

**图 1**

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-2/101_figure1.png)

[https://github.com/ardanlabs/gotraining/tree/master/topics/go/profiling/project](
https://github.com/ardanlabs/gotraining/tree/master/topics/go/profiling/project)

图 1 展示了这个 Web 应用的界面，这个应用从不同的新闻提供者那里下载三组 rss 源并且允许用户进行搜索。编译过后，执行该应用

**清单 1**

```bash
$ go build
$ GOGC=off ./project > /dev/null
```

清单 1 向我们展示了如何启动应用，并通过设置 `GOGC` 环境变量为 `off` 来关闭垃圾回收。日志文件被重定向到了 `/dev/null` 设备。应用运行的过程中，我们可以发送请求到服务器。

**清单 2**

```bash
$ hey -m POST -c 100 -n 10000 "http://localhost:5000/search?term=topic&cnn=on&bbc=on&nyt=on"
```

清单 2 向我们展示了如何利用 `hey` 工具通过 100 个连接向服务端发送 10k 的请求。当所有的请求都成功发送到服务器时，就会发生接下来的结果

**图片 2**

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-2/101_figure2.png)

图片 2 形象的展示了在垃圾回收器关闭的情况下处理 10k 数据的过程。处理 10k 的请求共花费了 4188 ms，这样算下来服务器每秒只能处理 2387 个请求。

## 打开垃圾回收器

如果启动这个应用的时候把垃圾回收打开会发生什么呢？

**清单 3**

```bash
$ GODEBUG=gctrace=1 ./project > /dev/null
```

清单 3 展示了如何启动应用并打开 GC 追踪。这里移除了 `GOGC` 变量并且替换为 `GODEBUG`。`GODEBUG` 这样设置了之后，运行时就会在每次垃圾回收发生的时候产生一条 GC 追踪。现在又可以向服务器发送同样的 10k 请求了。一旦所有的请求都发送到服务器后，就会产生一些可以用来分析的 GC 追踪信息和 `hey` 工具产生的信息。

**清单 4**

```bash
$ GODEBUG=gctrace=1 ./project > /dev/null
gc 3 @3.182s 0%: 0.015+0.59+0.096 ms clock, 0.19+0.10/1.3/3.0+1.1 ms CPU, 4->4->2 MB, 5 MB Goal, 12 P
.
.
.
gc 2553 @8.452s 14%: 0.004+0.33+0.051 ms clock, 0.056+0.12/0.56/0.94+0.61 ms CPU, 4->4->2 MB, 5 MB Goal, 12 P
```

清单 4 展示了运行过程中的第三条和最后一条 GC 追踪信息。我没有展示前两个垃圾回收信息是因为负载发送到服务器的时候这两次垃圾回收已经发生了。最后一次垃圾回收显示了它在处理 10k 请求过程中一共发生了 2551 次垃圾回收（减去了前两次，它们没有计算在内）。

下边是追踪信息的具体内容

**清单 5**

```bash
gc 2553 @8.452s 14%: 0.004+0.33+0.051 ms clock, 0.056+0.12/0.56/0.94+0.61 ms CPU, 4->4->2 MB, 5 MB Goal, 12 P

gc 2553     : The 2553 GC runs since the program started
@8.452s     : Eight seconds since the program started
14%         : Fourteen percent of the available CPU so far has been spent in GC

// wall-clock
0.004ms     : STW        : Write-Barrier - Wait for all Ps to reach a GC safe-point.
0.33ms      : Concurrent : Marking
0.051ms     : STW        : Mark Term     - Write Barrier off and clean up.

// CPU time
0.056ms     : STW        : Write-Barrier
0.12ms      : Concurrent : Mark - Assist Time (GC performed in line with allocation)
0.56ms      : Concurrent : Mark - Background GC time
0.94ms      : Concurrent : Mark - Idle GC time
0.61ms      : STW        : Mark Term

4MB         : Heap memory in-use before the Marking started
4MB         : Heap memory in-use after the Marking finished
2MB         : Heap memory marked as live after the Marking finished
5MB         : Collection Goal for heap memory in-use after Marking finished

// Threads
12P         : Number of logical processors or threads used to run Goroutines.
```

清单 5 展示了最后一次垃圾回收的实际数据。幸亏有了 `hey` 工具，这些是运行的性能结果。

**清单 6**

```bash
Requests            : 10,000
------------------------------------------------------
Requests/sec        : 1,882 r/s   - Hey
Total Duration      : 5,311ms     - Hey
Percent Time in GC  : 14%         - GC Trace
Total Collections   : 2,551       - GC Trace
------------------------------------------------------
Total GC Duration   : 744.54ms    - (5,311ms * .14)
Average Pace of GC  : ~2.08ms     - (5,311ms / 2,551)
Requests/Collection : ~3.98 r/gc  - (10,000 / 2,511)
```

清单 6 展示了执行结果。下边则更形象的向我们展示了发生了什么

**图片 3**

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-2/101_figure3.png)

图 3 形象的展示了发生了什么。当垃圾回收打开的时候它必须执行 2.5k 次来处理同样的 10k 请求。平均每 2 ms 开始一次垃圾回收，执行所有的垃圾回收需要增加额外的 1.1 秒的延迟

**图 4**

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-2/101_figure4.png)

图 4 展示了到现在为止应用两次执行的对比

## 减少分配

获取堆的 profile 并看下有没有可移除的非生产性质的分配将会很有用

**清单 7**

```bash
go tool pprof http://localhost:5000/debug/pprof/allocs
```

清单 7 展示了使用 `pprof` 工具调用 `/debug/pprof/allocs` 路径（endpoint）来从运行的应用中拉取内存的性能概要（profile）。之所以用那个路径是因为在源程序中添加了以下代码

**清单 8**

```go
import _ "net/http/pprof"

go func() {
    http.ListenAndServe("localhost:5000" , http.DefaultServeMux)
}()
```

清单 8 展示了如何绑定 `/debug/pprof/allocs` 路径到任何应用中。增加 `net/http/pprof` 包的导入可以绑定路径到默认的服务器路由。然后调用 `http.ListenAndServer` 方法并传入 `http.DefaultServerMux` 常量使该路径可用。

profiler 启动后，就可以用 `top` 命令查看正在分配内存的前 6 个方法

**清单 9**

```bash
(pprof) top 6 -cum
Showing nodes accounting for 0.56GB, 5.84% of 9.56GB total
Dropped 80 nodes (cum <= 0.05GB)
Showing top 6 nodes out of 51
      flat  flat%   sum%        cum   cum%
         0     0%     0%     4.96GB 51.90%  net/http.(*conn).serve
    0.49GB  5.11%  5.11%     4.93GB 51.55%  project/service.handler
         0     0%  5.11%     4.93GB 51.55%  net/http.(*ServeMux).ServeHTTP
         0     0%  5.11%     4.93GB 51.55%  net/http.HandlerFunc.ServeHTTP
         0     0%  5.11%     4.93GB 51.55%  net/http.serverHandler.ServeHTTP
    0.07GB  0.73%  5.84%     4.55GB 47.63%  project/search.rssSearch
```

清单 9 在清单的底部展示了 `rssSearch` 方法的表现。这个方法到现在共分配了 5.96 GB 中的 4.55 GB。现在是时候使用 `list` 命令来分析 `rssSearch` 方法的细节了

**清单 10**

```bash
(pprof) list rssSearch
Total: 9.56GB
ROUTINE ======================== project/search.rssSearch in project/search/rss.go
   71.53MB     4.55GB (flat, cum) 47.63% of Total

         .          .    117:   // Capture the data we need for our results if we find ...
         .          .    118:   for _, item := range d.Channel.Items {
         .     4.48GB    119:           if strings.Contains(strings.ToLower(item.Description), strings.ToLower(term)) {
   48.53MB    48.53MB    120:                   results = append(results, Result{
         .          .    121:                           Engine:  engine,
         .          .    122:                           Title:   item.Title,
         .          .    123:                           Link:    item.Link,
         .          .    124:                           Content: item.Description,
         .          .    125:                   })
```

清单 10 展示了 list 命令的执行结果。指出了 119 行代码分配了大量的内存

**清单 11**

```bash
4.48GB    119:           if strings.Contains(strings.ToLower(item.Description), strings.ToLower(term)) {
```

清单 11 展示了出现问题的那行代码。那个方法到现在一共分配了 4.55 GB，仅改行代码就占了 4.48 GB。接下来，是时候审核一下这行代码来看看有什么可以做的。

**清单 12**

```go
117 // Capture the data we need for our results if we find the search term.
118 for _, item := range d.Channel.Items {
119     if strings.Contains(strings.ToLower(item.Description), strings.ToLower(term)) {
120         results = append(results, Result{
121             Engine:  engine,
122             Title:   item.Title,
123             Link:    item.Link,
124             Content: item.Description,
125        })
126    }
127 }
```

清单 12 展示了那行代码在一个紧密的循环中。对 `strings.ToLower` 的调用将会产生分配因为创建新的 `strings` 需要在堆上分配内存。这些 `strings.ToLower` 的调用是没有必要的，因为这些调用可以在循环外完成。

修改一下 119 行可以移除掉所有的这些内存分配

**清单 13**

```go
// Before the code change.
if strings.Contains(strings.ToLower(item.Description), strings.ToLower(term)) {

// After the code change.
if strings.Contains(item.Description, term) {
```

*注意：你没看到的其他的改动的代码功能是将源放到缓存里之前让描述变为小写。新闻源每 15 分钟缓存一次。使 `term` 变为小写的调用是在循环外部*

清单 13 展示了如何移除了对 `strings.ToLower` 的调用。用新的改动过的代码重新编译该项目，重新对服务器发起 10 请求。

**清单 14**

```bash
$ go build
$ GODEBUG=gctrace=1 ./project > /dev/null
gc 3 @6.156s 0%: 0.011+0.72+0.068 ms clock, 0.13+0.21/1.5/3.2+0.82 ms CPU, 4->4->2 MB, 5 MB Goal, 12 P
.
.
.
gc 1404 @8.808s 7%: 0.005+0.54+0.059 ms clock, 0.060+0.47/0.79/0.25+0.71 ms CPU, 4->5->2 MB, 5 MB Goal, 12 P
```

清单 14 展示了在代码修改之后处理同样的 10k 请求现在是如何使用了 1402 次垃圾回收的。这些是两次执行的所有结果。

**清单 15**

```bash
With Extra Allocations              Without Extra Allocations
======================================================================
Requests            : 10,000        Requests            : 10,000
----------------------------------------------------------------------
Requests/sec        : 1,882 r/s     Requests/sec        : 3,631 r/s
Total Duration      : 5,311ms       Total Duration      : 2,753 ms
Percent Time in GC  : 14%           Percent Time in GC  : 7%
Total Collections   : 2,551         Total Collections   : 1,402
----------------------------------------------------------------------
Total GC Duration   : 744.54ms      Total GC Duration   : 192.71 ms
Average Pace of GC  : ~2.08ms       Average Pace of GC  : ~1.96ms
Requests/Collection : ~3.98 r/gc    Requests/Collection : 7.13 r/gc
```

清单 15 展示了跟最后一次的对比结果。下边是更形象的表达发生了什么

**图 5**

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-2/101_figure5.png)

图 5 形象的展示了发生了什么。这次处理相同的 10k 请求回收器少执行了 1149 (1420 vs 2551) 次。这使得整个的 GC 时间百分比从 14% 降到 7%。使应用的运行速度提升了 48%，垃圾回收时间降低了 74%。

**图 6**

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/Garbage-Collection-in-Go-Part-2/101_figure6.png)

图 6 展示了应用所有不同运行情况的一个对比。为了完整性，我将优化后的代码并关闭垃圾回收的运行情况也包含在内

## 我们学到了什么

正如我在上一篇文章中说的，对垃圾回收器友好就是降低堆上的压力。记住，压力可以定义为应用在确定时间内在堆上分配所有可用内存的速度。当压力降低时，由垃圾回收器所造成的延迟也会降低。拖慢你应用的就是这个延迟。

对回收器友好跟放慢垃圾回收的步调无关，而是跟在垃圾回收的间隔或期间让更多的工作做完有关。你可以通过降低任何一个工作在堆上分配内存的数量和次数来达到目的。

**清单 16**

```bash
With Extra Allocations              Without Extra Allocations
======================================================================
Requests            : 10,000        Requests            : 10,000
----------------------------------------------------------------------
Requests/sec        : 1,882 r/s     Requests/sec        : 3,631 r/s
Total Duration      : 5,311ms       Total Duration      : 2,753 ms
Percent Time in GC  : 14%           Percent Time in GC  : 7%
Total Collections   : 2,551         Total Collections   : 1,402
----------------------------------------------------------------------
Total GC Duration   : 744.54ms      Total GC Duration   : 192.71 ms
Average Pace of GC  : ~2.08ms       Average Pace of GC  : ~1.96ms
Requests/Collection : ~3.98 r/gc    Requests/Collection : 7.13 r/gc
```

清单 16 展示了两个版本的应用在垃圾回收打开的情况下运行的结果。很明显，移除了 4.48 G 的内存分配使得应用运行的更快。有趣的是，每一次的垃圾回收的平均时间几乎一样，差不多 2.0 ms。这两个版本之间最根本的改变是每次垃圾回收间隔期间工作完成的数量。应用从 3.98 r/gc 到 7.13 r/gc。工作完成数量增加了 79.1%。

在任何两次垃圾回收间隔之间让更多的工作做完帮助将所需要的垃圾回收次数从 2551 降到 1402，降低了 45%。应用在整个 GC 时间上从 745 ms 降到了 193 ms，74% 的降低，对于各自的版本在垃圾回收时间上也有一个 14% 到 7% 的降低。当以关闭垃圾回收的方式运行代码优化后的版本时，应用消耗时间从 2753 降到了 2398，性能差距仅有 13%。

## 结论

作为一个 Go 开发者，如果你花时间专注于降低内存分配，你正在尽你所能的对垃圾回收器友好。你不能写一个 0 内存分配的应用，所以认清生产性 ( 有助于应用 ) 和非生产性 ( 对应用有害 ) 的内存分配之间的区别很重要。信任垃圾回收器，并保证堆的合理使用，你的应用就会一直运行完好

有一个垃圾回收器是一个很好的权衡。我会为垃圾回收的成本买单，所以我没有内存管理的负担。Go 是想让你作为一个开发者能够快速写出一个性能足够好的应用。垃圾回收器为实现这一目标起到了很大作用。在下一篇文章中，我将会分享另外一个项目，它将会展示垃圾回收器是如何分析你的 Go 应用并找到最佳回收路径的。

---

via: https://www.ardanlabs.com/blog/2019/05/garbage-collection-in-go-part2-gctraces.html

作者：[William Kennedy](https://github.com/ardanlabs)
译者：[tpkeeper](https://github.com/tpkeeper)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
