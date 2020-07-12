# Go：Goroutine 的切换过程实际上涉及了什么

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://miro.medium.com/max/1400/1*CieXcIc9Bv11JWFOECjHyw.png)

本文基于 Go 1.13 版本。

Goroutine 很轻，它只需要 2Kb 的内存堆栈即可运行。另外，它们运行起来也很廉价，将一个 goroutine 切换到另一个的过程不牵涉到很多的操作。在深入 goroutine 切换过程之前，让我们回顾一下 goroutine 的切换在更高的层次上是如何进行的。

在继续阅读本文之前，我强烈建议您阅读我的文章 [Go：Goroutine、操作系统线程和 CPU 管理](https://medium.com/a-journey-with-go/go-goroutine-os-thread-and-cpu-management-2f5a5eaf518a) 以了解本文中涉及的一些概念。

## 案例

Go 根据两种断点将 goroutine 调度到线程上：

* 当 goroutine 因为系统调用、互斥锁或通道而被阻塞时，goroutine 将进入睡眠模式（等待队列），并允许 Go 调度运行另一个处于就绪状态的 goroutine；
* 在函数调用时，如果 goroutine 必须增加其堆栈，这会使 Go 调度另一个 goroutine 以避免运行中的 goroutine 独占 CPU 时间片；

在这两种情况下，运行调度程序的 `g0` 会替换当前的 goroutine，然后选出下一个将要运行的 goroutine 替换 `g0` 并在线程上运行。

有关 `g0` 的更多信息，建议您阅读我的文章 [Go：特殊的 Goroutine g0](https://medium.com/a-journey-with-go/go-g0-special-goroutine-8c778c6704d8)。

将一个运行中的 goroutine 切换到另一个的过程涉及到两个切换：

* 将运行中的 `g` 切换到 `g0`：

  ![](https://miro.medium.com/max/888/1*-w8MTDEUqis5mIX-s_KfPg.png)

* 将 `g0` 切换到下一个将要运行的 `g`：

  ![](https://miro.medium.com/max/892/1*6Qoa7ugcwsoQgs2cktKMvA.png)

在 Go 中，goroutine 的切换相当轻便，其中需要保存的状态仅仅涉及以下两个：

* goroutine 在停止运行前执行的指令，程序当前要运行的指令是记录在程序计数器（`PC`）中的， goroutine 稍后将在同一指令处恢复运行；

* goroutine 的堆栈，以便在再次运行时还原局部变量；

让我们看看实际情况下的切换是怎样进行的。

## 程序计数器

这里通过基于通道的`生产者/消费者模式`来举例说明，其中一个 goroutine 产生数据，而另一些则消费数据，代码如下：

![](https://miro.medium.com/max/1400/1*TZobNBH4mKyaN8B_ru7tUA.png)

消费者仅仅是打印从 0 到 99 的偶数。我们将注意力放在第一个 goroutine（生产者）上，它将数字添加到缓冲区。当缓冲区已满时，它将在发送消息时被阻塞。此时，Go 必须切换到 `g0` 并调度另一个 goroutine 来运行。

如前所述，Go 首先需要保存当前执行的指令，以便稍后在同一条指令上恢复 goroutine。程序计数器（`PC`）保存在 goroutine 的内部结构中：

![](https://miro.medium.com/max/958/1*ArVyzi31WBefg4RhhX5Pdw.png)

可以通过`go tool objdump`命令找到对应的指令及其地址，这是生产者的指令：

![](https://miro.medium.com/max/1400/1*E9HFNIw4ZhDirUh4dgWbsw.png)

程序逐条指令的执行直到在函数 `runtime.chansend1`处阻塞在通道上。 Go 将当前程序计数器保存到当前 goroutine 的内部属性中。在我们的示例中，Go 使用运行时的内部地址 `0x4268d0` 和方法 `runtime.chansend1` 保存程序计数器：

![](https://miro.medium.com/max/1400/1*i1SaUH3K7pjijTtW-O1TKw.png)

然后，当 `g0` 唤醒 goroutine 时，它将在同一指令处继续执行，继续将数值循环的推入通道。现在，让我们将视线移到 goroutine 切换期间堆栈的管理。

## 堆栈

在被阻塞之前，正在运行的 goroutine 具有其原始堆栈，该堆栈包含临时存储器，例如变量 `i`：

![](https://miro.medium.com/max/1194/1*8oa7ziZBpHZqKVihpQ3b8g.png)

然后，当它在通道上阻塞时，goroutine 将切换到 `g0` 及其堆栈（更大的堆栈）：

![](https://miro.medium.com/max/1194/1*I42dKDU2BV6kTwWMWiA1JQ.png)

在切换之前，堆栈将被保存，以便在 goroutine 再次运行时进行恢复：

![](https://miro.medium.com/max/958/1*kmufEth8mfd7OLnkl9oC7Q.png)

现在，我们对 goroutine 切换中涉及的不同操作有了一个完整的了解，让我们继续看看它是如何影响性能的。

我们应该注意，诸如 `arm` 等 CPU 架构需要再保存一个寄存器，即 `LR` 链接寄存器。

## 性能

我们仍然使用上述的程序来测量一次切换所需的时间。但是，由于切换时间取决于寻找下一个要调度的 goroutine 所花费的时间，因此无法提供完美的性能视图。在函数调用情况下进行的切换要比阻塞在通道上的切换执行更多的操作，这也会影响到性能。

让我们总结一下我们将要测量的操作：

* 当前 `g` 阻塞在通道上并切换到 `g0`：
  * `PC` 和堆栈指针一起保存在内部结构中
  * 将 `g0` 设置为正在运行的 goroutine
  * `g0` 的堆栈替换当前堆栈
* `g0` 寻找新的 goroutine 来运行；
* `g0` 使用所选的 goroutine 进行切换：
  * `PC` 和堆栈指针是从其内部结构中获取的
  * 程序跳转到对应的 `PC` 地址

结果如下：

![](https://miro.medium.com/max/1400/1*MDJam9-EE-XEIccguKXOkQ.png)

从 `g` 到 `g0` 或从 `g0` 到 `g` 的切换是相当迅速的，它们只包含少量固定的指令。相反，对于调度阶段，调度程序需要检查许多资源以便确定下一个要运行的 goroutine，根据程序的不同，此阶段可能会花费更多的时间。

该基准测试给出了性能的数量级估计，由于没有标准的工具可以衡量它，所以我们并不能完全依赖于这个结果。此外，性能也取决于 CPU 架构、机器（本文使用的机器是 Mac 2.9 GHz 双核 Intel Core i5）以及正在运行的程序。

---
via: https://medium.com/a-journey-with-go/go-what-does-a-goroutine-switch-actually-involve-394c202dddb7

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[anxk](https://github.com/anxk)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出