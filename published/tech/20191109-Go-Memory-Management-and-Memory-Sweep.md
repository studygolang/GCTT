首发于：https://studygolang.com/articles/27144

# Go：内存管理与内存清理

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191109-Go-Memory-Management-and-Memory-Sweep/01.png)

<p align="center">Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.</p>

*这篇文章基于 Go 1.13 版本。有关内存管理的讨论在我的文章  ”[Go:内存管理与分配](https://medium.com/a-journey-with-go/go-memory-management-and-allocation-a7396d430f44) ” 中有解释。*

清理内存是一个过程，它能够让 Go 知道哪些内存段最近可用于分配。但是，它并不会使用将位置 0 的方式来清理内存。

## 将内存置 0

将内存置 0 的过程 —— 就是把内存段中的所有位赋值为 0 —— 是在分配过程中即时执行的。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191109-Go-Memory-Management-and-Memory-Sweep/02.png)

<p align="center">Zeroing the memory</p>

但是，我们可能想知道 Go 采用什么样的策略去知道哪些对象能够用于分配。由于在每个范围内有一个内部位图 `allocBits` ，Go 实际上会追踪那些空闲的对象。让我们从初始态开始来回顾一下它的工作流程，

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191109-Go-Memory-Management-and-Memory-Sweep/03.png)

<p align="center">Free objects tracking with allocBits</p>

就性能角度来看，`allocBits` 代表了一个初始态并且会保持不变，但是它会由 `freeIndex`（一个指向第一个空闲位置的增量计数器）所协助。

然后，第一个分配就开始了：
![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191109-Go-Memory-Management-and-Memory-Sweep/04.png)

<p align="center">Free objects tracking with allocBits</p>
`freeIndex` 现在增加了，并且基于 `allocBits` 知道了下一段空闲位置。

分配过程将会再一次出现，之后， GC 将会启动去释放不再被使用的内存。在标记期间，GC 会用一个位图 `gcmarkBits` 来跟踪在使用中的内存。让我们通过我们运行的程序以相同的示例为例，在第一个块不再被使用的地方。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191109-Go-Memory-Management-and-Memory-Sweep/05.png)

<p align="center">Memory tracking during the garbage collector</p>
正在被使用的内存被标记为黑色，然而当前执行并不能够到达的那些内存会保持为白色。
> 有关更多关于标记和着色阶段的信息，我建议你阅读我的这篇文章 [Go：GC 是如何标记内存的？]()
现在，我们可以使用 `gomarkBits` 精确查看可用于分配的内存。Go 现在也使用 `gomarkBits` 代替了 `allocBits` ，这个操作就是内存清理：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191109-Go-Memory-Management-and-Memory-Sweep/06.png)

<p align="center">Sweeping a span</p>
但是，这必须在每一个范围内执行完毕并且会花费许多时间。Go 的目标是在清理内存时不阻碍执行，并为此提供了两种策略。

## 清理阶段

Go 提供了两种方式来清理内存：

- 使用一个工作程序在后台等待，一个一个的清理这些范围。
- 当分配需要一个范围的时候即时执行。

关于后台工作程序，当开始运行程序时，Go 将设置一个后台运行的 Worker（唯一的任务就是去清理内存），它将进入睡眠状态并等待内存段扫描：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191109-Go-Memory-Management-and-Memory-Sweep/07.png)

<p align="center">Background sweeper</p>
通过追踪过程的周期，我们也能看到这个后台工作程序总是出现去清理内存：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191109-Go-Memory-Management-and-Memory-Sweep/08.png)

<p align="center">Background sweeper</p>
清理内存段的第二种方式是即时执行。但是，由于这些内存段已经被分发到每一个处理器的本地缓存 `mcache` 中，因此很难追踪首先清理哪些内存。这就是为什么 Go 首先将所有内存段移动到 `mcentral` 的原因。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191109-Go-Memory-Management-and-Memory-Sweep/09.png)

<p align="center">Spans are released to the central list</p>
然后，它将会让本地缓存 `mcache` 再次请求它们，去即时清理：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191109-Go-Memory-Management-and-Memory-Sweep/10.png)

<p align="center">Sweep span on the fly during allocation</p>
即时扫描确保所有内存段在保存资源的过程中都会得到清理，同时会保存资源以及不会阻塞程序执行。

## 与 GC 周期的冲突

正如之前看到的，由于后台只有一个 worker 在清理内存块，清理过程可能会花费一些时间。但是，我们可能想知道如果另一个 GC 周期在一次清理过程中启动会发生什么。在这种情况下，这个运行 GC 的 Goroutine 就会在开始标记阶段前去协助完成剩余的清理工作。让我们举个例子看一下连续调用两次 GC，包含数千个对象的内存分配的过程。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191109-Go-Memory-Management-and-Memory-Sweep/11.png)

<p align="center">Sweeping must be finished before a new cycle</p>
但是，如果开发者没有强制调用 GC，这个情况并不会发生。在后台运行的清理工作以及在执行过程中的清理工作应该足够多，因为清理内存块的数量和去触发一个新的周期（译者注：GC 周期）的所需的分配的数量成正比。

---
via：<https://medium.com/a-journey-with-go/go-memory-management-and-memory-sweep-cc71b484de05>
作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[sh1luo](https://github.com/sh1luo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
