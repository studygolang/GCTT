# SRE: 调试Go中简单的内存泄漏

[内存泄漏](https://en.wikipedia.org/wiki/Memory_leak)是一种即使当某块内存不再使用之后仍然没有被释放而产生的bug。
通常来说，它们是非常明显的，高度可见的，这使得它们成为学习程序调试的最佳选择。Go是一门特别适合识别定位内存泄漏的语言，因为它有一套强大的工具链，这套工具链配备了非常强大的工具（*pprof*），它可以非常轻松地查明内存的使用情况。

我希望这篇文章能够演示如何直观地识别内存，并将其使用范围缩小至特定的进程内，将进程的泄漏与我们的工作关联起来，最后使用*pprof*工具找到内存泄漏的根源。
设计这篇博客文章的初衷是为了简单地识别产生内存泄漏的根本原因。
我们对*pprof*工具只做简单的功能介绍，不会对其做详细功能的描述。

[这里](https://github.com/dm03514/grokking-go/tree/master/simple-memory-leak)提供了本文用来生成数据的服务。

## 什么是内存泄漏？

如果内存的使用率无限增长且永远达不到稳定的状态，那么极有可能存在内存泄漏。这里的关键在于，内存野蛮增长且无法达到稳定状态，并最终会导致程序明显地崩溃或者会产生影响系统性能的问题。

产生内存泄漏的原因有很多种。有可能是数据结构无限制增长导致的逻辑泄漏，亦或是来自复杂的糟糕对象的引用产生的泄漏，又或者是因为其他原因。
不管是什么原因导致的内存泄漏，就大多数的内存泄漏的现象而言，它们都会呈现出一种『锯齿状』的形态。

## 调试过程

这篇博客文章旨在探索如何定位和查明Go中内存泄漏的根本原因。我们主要关注内存泄漏的特征，以及如何定位它们，并且学习如何使用Go（的工具链）来确定产生它们（内存泄漏）的根本原因。因此，我们实际的调试过程可能会比较浅显。

___我们分析的目的是逐步缩小问题的范围，排除各种可能性，直到我们拥有足够的信息来确定某个假设能够成立。___ 在我们有足够的数据和合理的根因范围之后，我们应该提出一个假设，并试图用数据佐证这个假设是否成立。

我们调试的每一步都将试图找出问题的根本原因或者证明这个假设是是否成立。在此过程中，我们将形成一系列的假设，首先它们必须是通用的，然后逐步具体化。
大体上来说，这是基于科学的方法。*Brandon Gregg*在覆盖系统研究的不同方法（他主要关注性能）这方面做的非常出色。

再次重申下，我们将会按照下面的方式逐步尝试：

- 先提出一个问题
- 形成一个假设
- 分析这个假设
- 重复这个过程直到发现根因

### 定位问题

我们是如何知道当前的系统存在某个问题（即内存泄漏）？有明显的错误是一个问题的直接表象。对于内存泄漏来说，一般的错误类型就是： OOM（Out Of Memory)错误或者是一个明显的程序崩溃。

#### OOM错误

错误是问题出现的明确指标。虽然用户或应用程序在某段逻辑关闭可能会产生一些误报的错误，但OOM错误不同，它实际上表示的是操作系统使用了过多的内存。下面清单列举出的错误，是因为触发了*Cgroup*的限制导致容器被杀死而产生的。

##### dmesg

```bash
[14808.063890] main invoked oom-killer: gfp_mask=0x24000c0, order=0, oom_score_adj=0                                                                                                                                                 [7/972]
[14808.063893] main cpuset=34186d9bd07706222bd427bb647ceed81e8e108eb653ff73c7137099fca1cab6 mems_allowed=0
[14808.063899] CPU: 2 PID: 11345 Comm: main Not tainted 4.4.0-130-generic #156-Ubuntu
[14808.063901] Hardware name: innotek GmbH VirtualBox/VirtualBox, BIOS VirtualBox 12/01/2006
[14808.063902]  0000000000000286 ac45344c9134371f ffff8800b8727c88 ffffffff81401c43
[14808.063906]  ffff8800b8727d68 ffff8800b87a5400 ffff8800b8727cf8 ffffffff81211a1e
[14808.063908]  ffffffff81cdd014 ffff88006a355c00 ffffffff81e6c1e0 0000000000000206
[14808.063911] Call Trace:
[14808.063917]  [<ffffffff81401c43>] dump_stack+0x63/0x90
[14808.063928]  [<ffffffff81211a1e>] dump_header+0x5a/0x1c5
[14808.063932]  [<ffffffff81197dd2>] oom_kill_process+0x202/0x3c0
[14808.063936]  [<ffffffff81205514>] ? mem_cgroup_iter+0x204/0x3a0
[14808.063938]  [<ffffffff81207583>] mem_cgroup_out_of_memory+0x2b3/0x300
[14808.063941]  [<ffffffff8120836d>] mem_cgroup_oom_synchronize+0x33d/0x350
[14808.063944]  [<ffffffff812033c0>] ? kzalloc_node.constprop.49+0x20/0x20
[14808.063947]  [<ffffffff81198484>] pagefault_out_of_memory+0x44/0xc0
[14808.063967]  [<ffffffff8106d622>] mm_fault_error+0x82/0x160
[14808.063969]  [<ffffffff8106dae9>] __do_page_fault+0x3e9/0x410
[14808.063972]  [<ffffffff8106db32>] do_page_fault+0x22/0x30
[14808.063978]  [<ffffffff81855c58>] page_fault+0x28/0x30
[14808.063986] Task in /docker/34186d9bd07706222bd427bb647ceed81e8e108eb653ff73c7137099fca1cab6 killed as a result of limit of /docker/34186d9bd07706222bd427bb647ceed81e8e108eb653ff73c7137099fca1cab6
[14808.063994] memory: usage 204800kB, limit 204800kB, failcnt 4563
[14808.063995] memory+swap: usage 0kB, limit 9007199254740988kB, failcnt 0
[14808.063997] kmem: usage 7524kB, limit 9007199254740988kB, failcnt 0
[14808.063986] Task in /docker/34186d9bd07706222bd427bb647ceed81e8e108eb653ff73c7137099fca1cab6 killed as a result of limit of /docker/34186d9bd07706222bd427bb647ceed81e8e108eb653ff73c7137099fca1cab6
[14808.063994] memory: usage 204800kB, limit 204800kB, failcnt 4563
[14808.063995] memory+swap: usage 0kB, limit 9007199254740988kB, failcnt 0
[14808.063997] kmem: usage 7524kB, limit 9007199254740988kB, failcnt 0
[14808.063998] Memory cgroup stats for /docker/34186d9bd07706222bd427bb647ceed81e8e108eb653ff73c7137099fca1cab6: cache:108KB rss:197168KB rss_huge:0KB mapped_file:4KB dirty:0KB writeback:0KB inactive_anon:0KB active_anon:197168KB inacti
ve_file:88KB active_file:4KB unevictable:0KB
[14808.064008] [ pid ]   uid  tgid total_vm      rss nr_ptes nr_pmds swapents oom_score_adj name
[14808.064117] [10517]     0 10517    74852     4784      32       5        0             0 go
[14808.064121] [11344]     0 11344    97590    46185     113       5        0             0 main
[14808.064125] Memory cgroup out of memory: Kill process 11344 (main) score 904 or sacrifice child
[14808.083306] Killed process 11344 (main) total-vm:390360kB, anon-rss:183712kB, file-rss:1016kB
```

___问题___：这是一个经常重复出现的问题吗？

___假设___：OOM错误非常重要，应该很少发生。如果发生了，那么应该是某个进程产生了内存泄漏。

___预测___：可能是进程内存限制的配额设置的太低，并且存在不明显地抖动或者是其他更严重的问题。

___测试___：经过进一步的检查，有相当多的OOM错误表明这是一个严重的问题，而不是偶发的。我们需要查看系统内存历史使用情况。尖峰一般表示应用程序正在运行（内存持续增长），而陡降则表示应用程序重启。

#### 系统内存

在确定了潜在问题之后，下一步就是了解系统范围的内存使用情况。内存泄漏经常以『锯齿状』的图案显示出来。

『锯齿状』的特征一般表明了存在内存泄漏，特别与服务的部署有关。我将使用一个测试的项目来演示内存泄漏，如果将视图的时间范围放大到足够大，那么即使是缓慢的内存泄漏也会以『锯齿状』形式展现出来。
在较小的时间范围内，它看起来像是逐渐上升的，接着会陡降（因为服务重启的缘故）。

![pic1](https://cdn-images-1.medium.com/max/800/1*6Mb-BJ9sXQspuCHyo93I7w.png)

上图展示了一个内存以锯齿状增长的示例。内存在持续的增长，而不是保持一个平稳的曲线，所以有确凿的证据表示这是一个内存问题。

___问题___：哪个（或哪几个）进程导致了内存的持续增长？

___测试___：分析每个进程的内存。*dmesg*日志中可能存在和OOM进程有关的信息。

#### 每个进程的内存

一旦开始怀疑是内存泄漏，下一步就是确定是哪个进程产生的内存泄漏或导致系统已有内存的持续增长。拥有每个进程的历史内存指标是一项至关重要的要求（基于容器的系统资源监控指标可以通过[cAdvisor](https://github.com/google/cadvisor)等工具获得）。
Go的[prometheus客户端](https://godoc.org/github.com/prometheus/client_golang/prometheus#hdr-Metrics)默认提供了每个进程的内存监控指标数据，这也是下面监控图表获取数据的来源。

下面显示了一个与之前系统内存锯齿特征非常相似的进程：在这个进程重启之前内存占用在持续地增长。

![pic2](https://cdn-images-1.medium.com/max/800/1*osVwTvlpnyPlCOV8h_UEkA.png)

内存是一种非常关键的资源，可以用来指明不正常的资源使用情况，也可以用作容量伸缩判断的维度。此外，如果有内存的统计信息也将有助于了解如何设置基于容器的（*Cgroups*）内存配额限制。
上图中各种值的[具体含义](https://povilasv.me/prometheus-go-metrics/)可以在[指标源代码](https://github.com/prometheus/client_golang/blob/v0.9.0-pre1/prometheus/process_collector.go#L126)中找到。在定位到问题进程之后，我们还需要深入挖掘并找到哪一部分代码引起了内存增长。

### 根因分析/源码分析

#### Go内存分析

*prometheus*再次向我们提供了Go运行时的一些信息，以及我们的进程正在做的事情。下面的图表显示在应用重启之前，堆上会不断地进行字节分配。而且每次曲线下降都能对应到服务重启的时间点。

![pic3](https://cdn-images-1.medium.com/max/800/1*h73GRFEHrmLkNpJKUXZ6AQ.png)

___问题___：该进程的哪一（些）部分造成了内存的泄漏？

___假设___：Goroutine中存在内存泄漏，它不断地将内存分配到堆上（全局变量或者指针，通过[逃逸分析](https://www.ardanlabs.com/blog/2017/05/language-mechanics-on-escape-analysis.html)可能会看到）。

___测试___：将内存使用情况与应用程序事件关联起来。

##### 与应用程序的工作关联起来

建立相关性将有助于通过回答以下问题来划分问题空间：这是在线发生的（与事务相关）还是在后台发生的？

确定这一点的一种方法就是启动服务，让其处于空闲状态，不让其处理任何事务以避免产生负载。然后观察其是否有内存泄漏，如果存在泄漏，那么可能是框架或者类库的问题。而我们的示例恰好显示出它与事务工作的负载有很强的相关性。

![pic4](https://cdn-images-1.medium.com/max/800/1*I7oUxMZKtmVJyXOd5S9cMQ.png)

上图显示了*HTTP*请求数。这些曲线和系统内存增长曲线完全匹配，并且定位到可能是由于处理*HTTP*请求引起的，这是一个好的开始。

___问题___：该应用哪一部分负责堆内存的分配？

___假设___：某一个*HTTP*处理程序再不断地进行[堆内存的分配](https://golang.org/doc/faq#stack_or_heap)。

___测试___：在程序运行期间周期性地分析堆内存分配，以跟踪内存的增长情况。

##### Go中内存分配

为了检查有多少内存分配到了堆上以及这些分配的来源是什么，我们将使用[pprof](https://golang.org/pkg/runtime/pprof/)工具。*pprof*是一个非常神奇的工具，这也是我个人偏爱它的主要原因之一。使用它之前，我们必须要启用它，然后获取一些堆内存快照。如果你已经使用了*http*，那么启用它非常简单：

```go
import _ "net/http/pprof"
```

一旦*pprof*启用成功，我们将在进程内存增长的整个生命周期内周期性获取堆内存快照。获取堆内存快照同样很简单：

```bash
curl http://localhost:8080/debug/pprof/heap > heap.0.pprof
sleep 30
curl http://localhost:8080/debug/pprof/heap > heap.1.pprof
sleep 30
curl http://localhost:8080/debug/pprof/heap > heap.2.pprof
sleep 30
curl http://localhost:8080/debug/pprof/heap > heap.3.pprof
```

获取堆内存快照的目的在于了解程序整个生命周期内内存的增长情况。让我们先来检查下最新的堆内存快照吧：

```bash
$ go tool pprof pprof/heap.3.pprof
Local symbolization failed for main: open /tmp/go-build598947513/b001/exe/main: no such file or directory
Some binary filenames not available. Symbolization may be incomplete.
Try setting PPROF_BINARY_PATH to the search path for local binaries.
File: main
Type: inuse_space
Time: Jul 30, 2018 at 6:11pm (UTC)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) svg
Generating report in profile002.svg
(pprof) top20
Showing nodes accounting for 410.75MB, 99.03% of 414.77MB total
Dropped 10 nodes (cum <= 2.07MB)
      flat  flat%   sum%        cum   cum%
  408.97MB 98.60% 98.60%   408.97MB 98.60%  bytes.Repeat
    1.28MB  0.31% 98.91%   410.25MB 98.91%  main.(*RequestTracker).Track
    0.50MB  0.12% 99.03%   414.26MB 99.88%  net/http.(*conn).serve
         0     0% 99.03%   410.25MB 98.91%  main.main.func1
         0     0% 99.03%   410.25MB 98.91%  net/http.(*ServeMux).ServeHTTP
         0     0% 99.03%   410.25MB 98.91%  net/http.HandlerFunc.ServeHTTP
         0     0% 99.03%   410.25MB 98.91%  net/http.serverHandler.ServeHTTP
```

这绝对非常神奇。*pprof*产生的快照默认类型是：`inuse_space`，它显示了内存中的所有对象。在这里我们可以看到`bytes.Repeat`占用了将近98.60%的内存。

下行日志显示了`bytes.Repeat`相关的条目信息：

```bash
1.28MB  0.31% 98.91%   410.25MB 98.91%  main.(*RequestTracker).Track
```

这真的很有意思，它表明`Track`本身只占用了1.28MB(0.31%)的内存，但却占用了整个已使用内存的98.91%！！！进一步我们可以看到*http*使用的内存甚至更少，但是占用内存比率却比`Track`更高（因为`Track`从它那里调用的）。

*pprof*提供了许多内省和可视化内存的方法（包括已使用内存大小、已使用对象数量、已分配内存大小、已分配内存对象等方面），它也可以列出追踪方法，并对应到每一行中。

```bash
(pprof) list Track
Total: 414.77MB
ROUTINE ======================== main.(*RequestTracker).Track in /vagrant_data/go/src/github.com/dm03514/grokking-go/simple-memory-leak/main.go
    1.28MB   410.25MB (flat, cum) 98.91% of Total
         .          .     19:
         .          .     20:func (rt *RequestTracker) Track(req *http.Request) {
         .          .     21:   rt.mu.Lock()
         .          .     22:   defer rt.mu.Unlock()
         .          .     23:   // alloc 10KB for each track
    1.28MB   410.25MB     24:   rt.requests = append(rt.requests, bytes.Repeat([]byte("a"), 10000))
         .          .     25:}
         .          .     26:
         .          .     27:var (
         .          .     28:   requests RequestTracker
         .          .     29:

```

这样就能直接找出罪魁祸首：

```bash
1.28MB   410.25MB     24:   rt.requests = append(rt.requests, bytes.Repeat([]byte("a"), 10000))
```

*pprof*还可以将上述文本信息可视化：

```bash
(pprof) svg
Generating report in profile003.svg
```

![pic5](https://cdn-images-1.medium.com/max/800/1*pJtDakl-VVc9uXvwj4QDlg.png)

这清楚地显示了占用进程内存的当前对象是哪个。既然我们已经定位到了罪魁祸首是`Track`，那我们就可以[验证它是否正在将内存分配到一个全局空间中，而没有对其进行清理](https://github.com/dm03514/grokking-go/pull/1/files#diff-2d715ee35186da340faaadfdf96fec9bR28)，然后尝试修复根本问题。

___解析___：在每个*HTTP*请求上，内存被持续地分配到了一个全局变量当中。

## 结论

我希望这篇文章能够介绍『直观地定位到内存泄漏』的强大功能，以及了解『逐步缩小待排查问题的范围』的过程。最后，我也希望这篇文章能够触及到Go工具链中的*pprof*对于Go内存内省和分析所起到的作用。和往常一样，我特别欢迎任何阅读反馈，感谢你的阅读。

----------------

via: [sre-debugging-simple-memory-leaks-in-go](https://medium.com/dm03514-tech-blog/sre-debugging-simple-memory-leaks-in-go-e0a9e6d63d4d)

作者：[dm03514](https://medium.com/@dm03514)
译者：[barryz](https://github.com/barryz)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，
[Go 中文网](https://studygolang.com/) 荣誉推出
