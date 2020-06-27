首发于：https://studygolang.com/articles/28989

# Go：GOMAXPROCS 和实时更新

![由 Renee French 创作的原始 Go Gopher 作品，为“ Go 的旅程”创作的插图。](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191210-Go-GOMAXPROCS-And-Live-Updates/1_Ct_BMGzFD4eKn6ztnR1iYA.png)

ℹ️ 这篇文章基于 Go 1.13。

`GOMAXPROCS` 控制着同时执行代码的 OS 线程的最大数量。这（GOMAXPROCS 值的设定）在程序启动期间，甚至在程序运行期间完成。一般来说，Go 将这个值设置为可用逻辑 CPU 的数量，但并不总是这样。

## 默认值

从 [Go 1.5](https://golang.org/doc/go1.5) 开始，`GOMAXPROCS` 的默认值由 1 改成了可见 CPU 的数量。这个改动也许是由于 Go 调度器和 Goroutine 上下文切换的改善。确实，在 Go 的早期阶段，旨在以频繁切换 Goroutine 的方式来并行工作的程序遭受了进程切换的困扰。

这个 `GOMAXPROCS` 新值的[提案](https://docs.google.com/document/d/1At2Ls5_fhJQ59kDK2DFVhFu3g5mATSXqqV5QrxinasI/edit)提供了展示这个提升的基准测试结果：

- 第一个基准测试，创建了 100 个由 channel 通信的 goroutine, 这些 channel 包括缓存的和非缓存的：

![使用更高的 `GOMAXPROCS` 值带来的调度器性能提升](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191210-Go-GOMAXPROCS-And-Live-Updates/Scheduler-improvement-with-higher-value-of-GOMAXPROCS.png)

- 用于质数生成的第二个基准测试展示了使用更多的核心将一个性能急剧下降的趋势转变为了一个巨大提升的趋势：

![更高的 `GOMAXPROCS` 值现在产生了很大的积极影响](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191210-Go-GOMAXPROCS-And-Live-Updates/Higher-value-for-GOMAXPROCS-has-now-a-great-positive-impact.png)

调度器已明显解决了单线程程序中遇到的问题，且现在可以很好的扩展。

## 在运行期间更新

Go 允许 `GOMAXPROCS` 在程序执行期间的任何时候进行更新。更新可以是虚拟机或者容器重新配置可用 CPU 数量所造成的。由于增减处理器数量的指令可以在任何时候发生，因此 Go 一进入“停止世界（Stop the World）”阶段该指令就生效。增加新的处理器是十分简单直接的，创建本地缓存 `mcache` 并且将新添加的处理器放入空闲队列中。这是当处理器从两个变成增长为三个时，一个新分配的 `P` 的例子：

![`GOMAXPROCS` 增加一个处理器](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191210-Go-GOMAXPROCS-And-Live-Updates/GOMAXPROCS-is-growing-by-one-processor.png)

之后，当重新开始调度的时候，新的 `P` 获取一个 Goroutine 运行：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191210-Go-GOMAXPROCS-And-Live-Updates/the-new-P-gets-a-goroutine-to-run.png)

减少处理器数量稍微有点复杂。移除一个 `P` 需要通过将 Goroutine 转移到全局队列的方式，将其本地的 Goroutine 队列清空：

![goroutine 转移到全局队列](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191210-Go-GOMAXPROCS-And-Live-Updates/Goroutines-moves-to-the-global-queue.png)

之后，要移除的 `P` 必须释放本地的 `mcache` 以便重复使用：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191210-Go-GOMAXPROCS-And-Live-Updates/free-the-local-mcache.png)

这是当 `P` 从二调整到一，之后再从一调整到三个 `P` 的追踪信息的例子：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191210-Go-GOMAXPROCS-And-Live-Updates/example-of-the-tracing.png)

## GOMAXPROCS=1

调整 `GOMAXPROCS` 到一个更大的值并不意味着你的程序一定会运行的更快。Go 文档解释地很清楚：

> 这取决于程序的性质。本质上是顺序处理的问题无法通过增加 Goroutine 的方式提高处理速度。当问题本质上是并行处理的时候，并发才会变成并行。

来看下对于某些程序而言，并发是如何满足需要的。这是一些检查某些 URL 的代码，用来感知这些网站是否正常运行：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191210-Go-GOMAXPROCS-And-Live-Updates/code-that-checks-some-URLs%20.png)

由于代码执行过程中有很多停顿，这段示例代码在并发情况下工作地很好，给了 Go 调度器空间在等待期间运行其他 goroutine。这是通过 `test` 包的 `-cpu=1,2,4,5` 标志，拿到的不同 `GOMAXPROCS` 的值的情况下的基准测试结果：

```bash
name         time/op
URLsCheck-8  4.19s ± 2%
URLsCheck-4  4.30s ± 5%
URLsCheck-2  4.33s ± 4%
URLsCheck-1  4.14s ± 1%
```

在这里增加并行不会带来任何的性能提升。使用 CPU 的全部容量在许多时候会带来性能提升。然而，最好在不同的（GOMAXPROCS）值下运行测试或基准测试来确定具体的表现。

---

via: https://medium.com/a-journey-with-go/go-gomaxprocs-live-updates-407ad08624e1

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[dust347](https://github.com/dust347)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://raw.githubusercontent.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
