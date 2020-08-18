首发于：https://studygolang.com/articles/30260

# Go：死锁是如何触发的？

![illustration](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200711-Go-How-Are-Deadlocks-Triggered/illustration.png)

由创作原始 Go Gopher 作品的 Renee French 为“ Go 的旅程”创作的插图。

*本文基于 Go 1.14。*

死锁是当 Goroutine 被阻塞而无法解除阻塞时产生的一种状态。Go 提供了一个死锁检测器，可以帮助开发人员避免陷入这种情况。

## 检测

让我们从创建这种情况的示例开始：

![example](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200711-Go-How-Are-Deadlocks-Triggered/example.png)

主 Goroutine 在 channel 上被阻塞，并等待另一个 Goroutine 将数据写入 channel。然而，没有其他的 Goroutine 在运行，它不能被解除阻塞。这种情况将触发死锁错误:

![deadlock](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200711-Go-How-Are-Deadlocks-Triggered/deadlock.png)

死锁检测器基于对应用程序创建的线程的分析。如果已创建并活动的线程数大于等待工作的线程数，则会出现死锁情况。

*这个公式中不包括为监视系统而创建的线程。*

在检测到死锁时将创建四个线程：

- 一个用于主 goroutine，启动程序的那个。

- 一个叫做 `sysmon`，用于监视系统。

- 一个专用于垃圾收集器的 Goroutine 启动的。

- 在初始化过程中阻塞主 Goroutine 时创建的一个线程。由于此 Goroutine 被锁定在它的线程上，因此 Go 需要创建一个新的 Goroutine 来为其他 Goroutine 提供运行时间。

每次调用死锁检测器时，也可以通过一些调试信息将其可视化：

![detector](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200711-Go-How-Are-Deadlocks-Triggered/detector.png)

每当线程空闲时，就会通知检测器。调试的每一行显示空闲线程的递增数量。当空闲线程数等于活动线程数减去系统线程数时，就会发生死锁。在本例中，我们有三个空闲线程和三个活动线程（四个线程减去系统线程）。由于没有活动线程能够解除阻塞空闲线程，因此存在死锁情况。

但是，这种行为有一些限制。实际上，任何自旋的 Goroutine 都会使死锁检测器失效，因为线程将保持活动状态。

## 限制

现在，通过发送中断信号使 OS 信号停止程序来改进前面的示例：

![example2](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200711-Go-How-Are-Deadlocks-Triggered/example2.png)

这是新的输出：

![output](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200711-Go-How-Are-Deadlocks-Triggered/output.png)

通过键盘发送中断信号后，程序停止了。不再检测到死锁。具有 `signal.Notify` 的任何活动程序都将运行后台 goroutine，等待输入信号。该 Goroutine 保持活动状态，并且永远不会使活动线程数等于空闲线程数。这是此 Goroutine 的跟踪：

![trace](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200711-Go-How-Are-Deadlocks-Triggered/trace.png)

它的大部分时间都花在等待系统调用上。syscall 中的线程不在空闲列表中，因此不会导致死锁。

但是，可以通过调试工具找到它们。

## 调试

发现这些死锁的最好方法可能是编写单元测试。编写测试确保一次运行较小的代码段。在这种情况下，不应该受到信号处理程序或阻塞系统调用的干扰。然而，即使这样做 ，测试也会挂起，我们肯定会发现有可疑的地方。

如果你想可视化运行程序上的死锁，可以使用 `pprof` 之类的工具来可视化它。下面是我们修改后的第一个程序，添加了调试功能：

![debugging](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200711-Go-How-Are-Deadlocks-Triggered/debugging.png)

一旦程序运行，我们就可以使用命令 `wget http://localhost:6060/debug/pprof/trace?seconds=5` 对我们的应用程序进行配置，该命令会生成 5s 的跟踪信息。 这些痕迹告诉我们所有活动：

![profile](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200711-Go-How-Are-Deadlocks-Triggered/profile.png)

没有 Goroutine 一直在运行。可以使用以下命令通过 CPU 配置文件进行确认 `go tool pprof http://localhost:6060/debug/pprof/profile?seconds=5`。下面是未显示活动的配置文件：

![no-activity](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200711-Go-How-Are-Deadlocks-Triggered/no-activity.png)

---

via: https://medium.com/a-journey-with-go/go-how-are-deadlocks-triggered-2305504ac019

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[alandtsang](https://github.com/alandtsang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
