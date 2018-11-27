#在Go中使用Mutex与Channel进行状态同步

我通过谷歌搜索查找哪种方法最适合GoLang的同步工作。 当我开始构建自己的GoLang包Go-Log时，我就开始了这项工作。 Go-Log是一个日志包，在Go的普通Log包之上提供实用程序，具有以下功能：将日志标记为调试和错误变体，向日志添加/删除时间戳以及在日志中获取调用功能详细信息。这就非常需要使这个日志记录线程安全，或者需要同步，因为当数百万个请求进入服务器日志时需要以较少的延迟同步。

通过浏览文章，StackOverflow问答，以及常用的Go教学视频时，我找到了两种方法：
1.使用线程安全的Channel
2.使用Mutex共享内存

## Go开发人员最常犯的错误是什么？

在开始讨论选择哪种方法之前，所有人需要知道作为Go开发人员易犯的常见错误。 我知道大家都醉心于并发模式，这种模式会让你过度使用Go强力的Channel和Goroutine，最终成为一种反模式。我并不是说Channel不好或者它们不能用于同步，但过度使用（在没有要求的情况下使用它）绝对不是你应该遵循的途径。

>Dave Cheney是Go编程语言的开源贡献者和项目成员，曾在一次采访中说过，“如果你想谈论最糟糕的代码，那就是我曾试图将Channel用于一切工作”

>Go的官方文档指出，“常见的Go新手错误是过度使用Channel和Goroutine，仅仅只是因为它是可能的，和/或因为它很有趣。”

现在假设您根本没有关于Goroutine和Channel或Mutex的任何概念，让我们看看之前提到的关于如何进行同步，两种方法之间的区别。
## 使用通道进行通信以实现同步

Channel是连接并发Goroutine的管道，您可以从一个Goroutine将值发送到Channel，并在另一个Goroutine中接收值。Channel最适合传递数据所有权，分发工作单元和传递异步结果等情况。

让我们利用Channel来实现GoLang包的同步（目的是使日志同步工作并保证线程安全），下面是我之前写的包中的代码片段。
```
func generateMessage(message string) {
	// 建立布尔型缓冲Channel
	done := make(chan bool, 1)
	go printMessage(done, message)
	// 阻塞主程，否则goroutine来不及执行主程序已结束
	<-done//Channel传递值并未使用
	fmt.Println("main program end!")
}

func printMessage(done chan bool, message string) {
	//延迟函数在return之前执行，否则向Channel发送数据主程就结束
	defer func() {
		// 向通道发送数据
		fmt.Println("chanel send data!")
		done <- true
	}()
	// 内部实现
	if true {
		log.Println(message)
		return
	}
	fmt.Println(message)
}
```

我们使用基本Channel通信对多个Goroutine进行同步，这可能在服务器记录时发生数百万次，关于通道的优点是它们具有内置的线程安全性并防止竟态条件。

如果你仔细看我的代码，并询问是否明确需要Channel？答案是否定的。Channel是Go中的高级概念，Go中某些程序仅使用Mutex， Go的Channel很吸引人因为它们提供了内置的线程安全性，并鼓励对共享的关键资源进行单线程访问。 但是与Mutex相比，Channel会导致性能下降。 当只需要锁定少量共享资源时，使用Mutex非常有用。 如果Mutex很适合你的需求请放心使用sync.Mutex。

## 我如何使用Mutex使日志记录同步？

Mutex在跨多个Goroutine同步状态时是一个极有用的资源，但我发现它的使用对于新的Go开发人员来说有些神秘。不过它很容易使用！这是我的软件包Go-Log的代码片段
```
var mutex = &sync.Mutex{}//互斥锁
var wg sync.WaitGroup
func generateMessage(message string) {
	wg.Add(1) //增加WaitGroup计数1
	go func() {
		//延迟函数，当goroutine退出时减少WaitGroup计数1
		defer wg.Done()
		printMessage(message)
	}()
	wg.Wait() //等待所有goroutine，当计数为0时等待结束
}

func printMessage(resultMessage string) {
	mutex.Lock()         //互斥锁锁定
	defer mutex.Unlock() //延迟函数退出时解锁
	// 内部实现: Ignore
	if true {
		log.Println(resultMessage)
		return
	}
	fmt.Println(resultMessage)
}
```
程序中使用了mutex.Lock()和mutex.Unlock()来创建共享资源的同步锁。为了管理多个日志，每次都要创建一个Goroutine并将它们添加到sync.WaitGroup，这是一个重要的同步原语，允许协作的Goroutine在再次独立进行之前共同等待阈值事件。

在上面的例子中，这种方法更好更快！ 我们减少了不必要的开销。 但是如果您发现sync.Mutex锁定规则过于复杂，扪心自问使用通道是否真的更简单。

查看我的软件包的源代码，了解我实际上是如何使用Mutex以实现同步：
>https://github.com/MindorksOpenSource/Go-Log

### 附加资源

>[Closing]
https://www.acloudtree.com/understanding-when-to-use-channels-or-mutexes-in-go/

>[MutexOrChannel]
https://github.com/golang/go/wiki/MutexOrChannel

### 联系我

Twitter: https://twitter.com/DuaYashish
MentorCruise: https://mentorcruise.com/mentor/YashishDua/
---

via: https://medium.com/mindorks/https-medium-com-yashishdua-synchronizing-states-using-mutex-vs-channel-in-go-25e646c83567

作者：[Yashish Dua](https://medium.com/@yashishdua)
译者：[译者ID](https://github.com/weiwg521)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
