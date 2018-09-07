首发于：https://studygolang.com/articles/14714

# Go 语言中的 Timer Routines 与优雅退出

在我的 Outcast（译注：作者自己做的一款天气预告 App） 数据服务器中，有几个数据检索任务要用到不同的 Go routine 来运行, 每个 routine 在设定的时间间隔内唤醒。 其中最复杂的工作是下载雷达图像。 它复杂的原因在于：美国有 155 个雷达站，它们每 120 秒拍摄一张新照片， 我们要把所有的雷达图像拼接在一起形成一张大的拼接图。（译注：有点像我们用手机拍摄全景图片时，把多张边缘有重叠的图片拼接成一张大图片） 当 go routine 被唤醒去下载新图像时，它必须尽快为所有 155 个站点都执行这个操作。 如果不够及时的话，得到拼接图将不同步，每个雷达站重叠的边界部分会对不齐。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Timer-Routines-And-Graceful-Shutdowns-In-Go/radar-img-1.png)

左边的雷达图是坦帕湾雷达站在下午 4:51 拍摄的，你可以看到，这个雷达站覆盖了佛罗里达州的大部分范围，事实上，这个雷达站甚至涵盖了其它雷达站的范围，比如说迈阿密的。

右边的雷达图是迈阿密雷达站在下午 4:53 拍摄的，跟右图存在了两分钟的差异，（我把这种情况称之为 glare）当我们把这两个雷达图铺叠在地图上的时候，你不会发现有什么不对的地方，但是，如果这两个图片之前的延迟不止几分钟的时候，我们裸眼就能看出有很大的区别。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Timer-Routines-And-Graceful-Shutdowns-In-Go/radar-img-2.png)

蓝色是雷达的噪点，我们会把它给过滤掉，所以我们剩下绿色、红色和黄色的色块来表示真正的天气状况。上面的图片是在下午 4:46 下载并处理好的，你可以看到他们很接近，能够很好的拼接在一起。

我们的代码的第一个实现中，使用了单个 go routine，每10分钟唤醒一次，每次这个 go routine 唤醒，它都需要 3 到 4 分钟时间下载、处理、保存并把 155 个雷达站的数据写入的到 mongo 里面去。虽然我会把每个地区的图片 尽可能地拼接起来，但是这些图片存在的延迟差异实在是太大了。每个雷达站都存在一两分钟的延迟，所有的雷达站的延迟叠加起来，使这个问题凸显出来。

对于所有工作，我都会尽可能地使用单个 go routine 来实现，因为这样能让事情保持简单。但在这个情况下，单一 go routine 并不能凑效。我必须同时处理多个雷达站的数据，来减少延迟造成的差异。在我添加了一个工作池来处理同时多个雷达站的数据后，我能够在一分钟之内把 155 个雷达站的数据都处理好了。目前为止，我还没收到客户端开发团队的抱怨。

在篇文章里面，我们主要关注定时 routine 和退出的代码。在下一个文章，我会告诉你怎么去为你的项目添加一个工作池。

我打算提供一个完整的可以运行的例子。它也许可以作为一个参考模板来让你实现你自己的代码。要下载这个例子，你可以打开一个新的终端会话，输入下面的命令：

```bash
cd $HOME
export GOPATH=$HOME/example
go get github.com/goinggo/timerdesignpattern
cd example/bin
./timerdesignpattern
```

Outcast 数据服务器是个单应用程序，它设计为长期运行的服务程序，这种类型的程序很少会需要退出。让你的程序能在需要的时候优雅地退出是很重要的。当我在开发这种类型的程序时，我总是要从开始就确保，我可以通过某些信号通知应用程序退出，并且不会让它挂起。一个程序，最糟糕的事情莫过于需要你强制杀死进程才能退出了。

示例程序创建了一个单一的 go routine 并且指定这个 routine 每 15 秒唤醒一次. 当 routine 唤醒的时候，它会进行一个大概耗时 10 秒的操作。当工作完成以后，它再计算需要睡多少秒，来确保这个 routine 能够保持每 15 秒唤醒一次。

让我们试试运行这个程序并且在它运行的时候把它退出掉。然后我们就可以开始学习它是怎么实现的。我们可以在程序运行的任何时候，按回车键来退出这个程序。

下面是程序运行 7 秒钟后退出的输出：

```
2013-09-04T18:58:45.505 : main : main : Starting Program
2013-09-04T18:58:45.505 : main : workmanager.Startup : Started
2013-09-04T18:58:45.505 : main : workmanager.Startup : Completed
2013-09-04T18:58:45.505 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Started
2013-09-04T18:58:45.505 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Info : Wait To Run : Seconds[15]

2013-09-04T18:58:52.666 : main : workmanager.Shutdown : Started
2013-09-04T18:58:52.666 : main : workmanager.Shutdown : Info : Shutting Down
2013-09-04T18:58:52.666 : main : workmanager.Shutdown : Info : Shutting Down Work Timer
2013-09-04T18:58:52.666 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Shutting Down
2013-09-04T18:58:52.666 : main : workmanager.Shutdown : Completed
2013-09-04T18:58:52.666 : main : main : Program Complete
```

这是一次很棒的初次测试，当我们指示程序退出的时候，它优雅地退出了。下一步我们试试看等它开始工作（译注：这个程序运行后要等 15 秒才开始执行第一次的工作）之后再尝试退出它。

```
2013-09-04T19:14:21.312 : main : main : Starting Program
2013-09-04T19:14:21.312 : main : workmanager.Startup : Started
2013-09-04T19:14:21.312 : main : workmanager.Startup : Completed
2013-09-04T19:14:21.312 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Started
2013-09-04T19:14:21.313 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Info : Wait To Run : Seconds[15]
2013-09-04T19:14:36.313 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Woke Up
2013-09-04T19:14:36.313 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Started
2013-09-04T19:14:36.313 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Processing Images For Station : 0
2013-09-04T19:14:36.564 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Processing Images For Station : 1
2013-09-04T19:14:36.815 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Processing Images For Station : 2
2013-09-04T19:14:37.065 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Processing Images For Station : 3

2013-09-04T19:14:37.129 : main : workmanager.Shutdown : Started
2013-09-04T19:14:37.129 : main : workmanager.Shutdown : Info : Shutting Down
2013-09-04T19:14:37.129 : main : workmanager.Shutdown : Info : Shutting Down Work Timer
2013-09-04T19:14:37.315 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Info : Request To Shutdown
2013-09-04T19:14:37.315 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Info : Wait To Run : Seconds[14]
2013-09-04T19:14:37.315 : WorkTimer : _WorkManager.GoRoutine_WorkTimer : Shutting Down
2013-09-04T19:14:37.316 : main : workmanager.Shutdown : Completed
2013-09-04T19:14:37.316 : main : main : Program Complete
```

这次我等了 15 秒，让程序开始工作，当它开始工作并完成了第四个图片的处理之后，我指示程序退出。它也及时停止了工作并优雅地退出了。

我们来看看实现定时 routine 和优雅退出的核心代码：

```go
func (wm *WorkManager) WorkTimer() {
    for {
        select {
        case <-wm.ShutdownChannel:
            wm.ShutdownChannel <- "Down"
            return

        case <-time.After(TimerPeriod):
            break
        }

        startTime := time.Now()
        wm.PerformTheWork()
        endTime := time.Now()

        duration := endTime.Sub(startTime)
        wait = TimerPeriod - duration
    }
}
```

为了更加简洁易读，我把注释和输出日志的代码去掉了。这是一个经典的作业队列 channel， 并且这个解决方案非常的优雅。比起用 C# 实现的同样的东西，优雅多了。

`WorkTimer()` 函数作为一个 go routine 运行：

```go
func Startup() {
    wm = WorkManager{
        Shutdown: false,
        ShutdownChannel: make(chan string),
    }

    go wm.WorkTimer()
}
```

`WorkManager` 是以单例（译注：设计模式的一种，参考[单例模式](https://zh.wikipedia.org/wiki/%E5%8D%95%E4%BE%8B%E6%A8%A1%E5%BC%8F)）的模式创建的，它创建完后就开始启动定时 routine。它有一个 channel 负责关闭定时 routine，还有一个标志用来指明系统是否正在退出。

定时 routine 在内部有一个无限的循环，所以它不会终止，除非我们们指明要它退出。我们来看看这个循环里面关于 channel 的部分：

```go
select {
case <-wm.ShutdownChannel:
    wm.ShutdownChannel <- "Down"
    return

case <-time.After(TimerPeriod):
    break
}

wm.PerformTheWork()
```

我们使用了 `select` 语句。这个语句在官方文档的解释在这里：

http://golang.org/ref/spec#Select_statements

我们使用 `select` 语句来保证定时 routine 只有到了工作时间或者收到退出指令的时候才会被唤醒。`Select` 语句使得定时 routine 在所有通道都没有收到信号的时候阻塞。每次只有其中一个分支会执行，这让我们的代码保持同步。`select` 语句让我们用简洁的代码在多个 channel 间实现原子的、“routine 安全”的操作（只要我们把这几个 channel 都放在同一个 `select` 语句里面）。

在我们的定时 routine 的 `select` 语句里面有两个 channel，一个负责退出 routine，一个负责执行任务。退出定时 routine 的代码如下：

```go
func Shutdown() {
    wm.Shutdown = true

    wm.ShutdownChannel <- "Down"
    <-wm.ShutdownChannel

    close(wm.ShutdownChannel)
}
```

当需要退出的时候，我们把  `Shutdown` 标记置为 `true`，然后发送字符串 `"Down"` 到 `ShutdownChannel`，然后我们从 `ShutdownChannel` 里面等待来自 定时 routine 的回复。这种数据通信同步了主程序和定时 routine 之间的整个退出过程。 非常的棒，简单而优雅。

要以一个固定的时间间隔唤醒定时 routine，我使用了一个叫做 `time.After` 的函数，这个函数等待一段指定的时间，然后把当前时间发送到指定的 channel 里面。这又唤醒了 `select`，从而使得 `PerformTheWork` 函数得以执行。当 `PerformTheWork` 函数返回时，定时 routine 又再一次回到睡眠状态，除非又有 channel 收到了新的信号。

我们来看一下 `PerformTheWork` 函数：

```go
func (wm *_WorkManager) PerformTheWork() {
    for count := 0; count < 40; count++ {
        if wm.Shutdown == true {
            return
        }

        fmt.Println("Processing Images For Station:", count)
        time.Sleep(time.Millisecond * 250)
    }
}
```

这个函数每 250 微秒在控制台输出一次信息，一共输出 40 次。这将会耗费大概 10 秒的时间来完成这个任务。在这个循环里面，每次迭代都检查一下 `Shutdown` 这个标记是否置为 `true`。这非常重要，因为它使得这个函数在程序退出时，能够非常快的结束掉。我们不希望使用这个程序的管理者在退出这个程序的时候，觉得觉得这个程序被挂起了。

当 `PerformTheWork` 函数结束后，定时 routine 得以再次执行 `select` 语句，如果程序正在退出的过程中，那么 `select` 语句会立刻唤醒来处理来自 `ShutdownChannel` 的信号。在这里，定时 routine 再通知主 routine 它正在退出，从而整个程序得以优雅地退出。

这就是我的定时 routine 和优雅退出程序的代码模式，你也可以把这个模式应用在你的程序中。如果你从 GoingGo 的代码仓库下载了整个示例的话，你可以看到实战的代码和一些小工具。

阅读下面的文章可以学习到怎么实现一个能够处理多个 go routine 的工作池，正如我上述的处理雷达图像的那个工作池一样：

https://studygolang.com/articles/14481

---

via: https://www.ardanlabs.com/blog/2013/09/timer-routines-and-graceful-shutdowns.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[Alex-liutao](https://github.com/Alex-liutao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
