已发布：https://studygolang.com/articles/11983 

# Golang 对于 DevOps 之利弊(第 1 部分，共 6 部分)：Goroutines, Panics 和 Errors

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go_devops/1.png)

对于你的下一个 DevOps 应用来说，Google 公司的 Go 可能是完美的语言。作为由 6 篇组成一个系列文章的第一篇，我们从 goroutines、panics 和 errors 开始深入研究 Go 语言的利与弊，因为这些利与弊涉及构建 DevOps 应用。

在这篇博客中，我们已经称赞了 Google 公司的 Go 语言用于 DevOps 应用开发的优点，而且我们也写了[Go 迷你入门指南](https://blog.bluematador.com/posts/mini-guide-google-golang-why-its-perfect-for-devops/?utm_source=bm-blog&utm_medium=link&utm_campaign=golang-pros-cons-1)。 

为了避免有人认为我们是打着 DevOps 监控平台幌子的 Google 员工（我们确实是正儿八经 DevOps 监控平台，而不是 Google 员工），我们希望深入研究 Go 语言的优缺点，特别是在涉及构建 DevOps 应用程序方面。由于我们将智能代理(软件)从 Python(构建) 转换到 Go(构建)，我们有许多在 Go 语言环境下工作(开发)的经验，并且我们也有一些自认为很重要的东西(知识)来与更大的 DevOps 社区分享。

从这周开始，我们将发布由 6 篇组成一系列关于 Go 语言利和弊的文章，每一篇文章都详述了少许的内容。正如我们所做的，我们将通过链接到其它篇目来更新这篇文章：

- Golang 对于 DevOps 之利弊第一篇：Goroutines, Channels, Panics, 和 Errors(本篇)
- [Golang 对于 DevOps 之利弊第二篇：自动接口实现，共有/私有变量](https://blog.bluematador.com/posts/golang-pros-cons-for-devops-part-2)
- [Golang 对于 DevOps 之利弊第三篇：速度 VS 缺少泛型](https://blog.bluematador.com/posts/golang-pros-cons-devops-part-3-speed-lack-generics)   
- [Golang 对于 DevOps 之利弊第四篇：打包时间与方法重载](https://blog.bluematador.com/golang-pros-cons-part-4-time-package-method-overloading)
- Golang 对于 DevOps 之利弊第五篇：交叉编译，窗口，信号，文档和编译器
- Golang 对于 DevOps 之利弊第六篇：Defer 语句和包依赖版本控制

如果这是你首次阅读有关 Go 的文章，并且你已经知道怎样用类 C 的语言进行编程，你应该去参考[Go 语言之旅](https://tour.golang.org/welcome/1)，它将大约花费你一个小时，而且介绍相当有深度。接下来的内容并不是介绍学习如何用 Go 进行编程，而是我们在用 Go 语言开发智能代理系统过程中有过的抱怨和发现的可取之处。

准备好听听用 Go 进行编程的真正样子吗？我要开始说了(原谅我，这是一个不高明的双关)（译注:原文是 `Here it goes`，goes 对应 go 语言）。

## Go 语言的好处 1: Goroutines — 轻量，内核级线程    

Goroutine 等同于 Go。Goroutine 是执行线程，并且它是轻量的、内核级的。轻量级是因为你可以在不影响系统性能的同时运行很多的 Goroutine，而内核级是因为它们并行运行（真正的并行运行，而不是像在其他语言中的伪并行，比如旧版的 Ruby）。同时也不像 Python 那样，在 Go 中没有全局解释器锁（GIL）。

换句话说，Goroutine 是 Go 语言中的重要部分，而不是后来才添加的特性，这正是 Go 语言执行之快的原因之一。另一方面，Java 有一个复杂的线程调度器。它们在为线程的执行创造条件，但是需要很多的管理工作，因此降低了你的程序运行速度。从另一个角度看，Node.js 没有真正的并发，仅仅是并发的假象。这是对于 DevOps 用途来说，Go 能成为一门伟大的语言的原因，它不影响你的系统性能，真正的并行执行，并且在意外的硬件故障中尽可能的可靠。

### 怎样执行一个 Goroutine

启动一个 goroutine 线程超级简单。首先，你将在我们的看门狗（Watchdog）产品中看到一些伪代码：

```go
import "time"
func monitorCpu() { … }
func monitorDisk() { … }
func monitorNetwork() { … }
func monitorProcesses() { … }
func monitorIdentity() { … }

func main() {
	for !shutdown {
		monitorCpu()
		monitorDisk()
		monitorNetwork()
		monitorProcesses()
		monitorIdentity()
		time.Sleep(5*time.Second)
	}
}
```
在上面的代码中，CPU、硬盘、网络、进程和身份每 5 秒检查一次。如果上述任何一个挂掉的话，所有的监控就会终止。每一个监控用时越长，检查频率就越低，因为我们在计算完成之后需要休眠 5 秒。

为了解决这些问题，一种解决方案（一种不完整的方案，但是完美展示 goroutine 的价值所在）是使用 goroutine 来调用每一个监控函数。仅仅在你想要以线程方式运行的函数调用前加上 `go` 关键字。

```go
func main() {
	for !shutdown {
		go monitorCpu()
		go monitorDisk()
		go monitorNetwork()
		go monitorProcesses()
		go monitorIdentity()
		time.Sleep(5*time.Second)
	}
}
```
现在如果它们之中任何一个挂掉的话，仅仅被阻塞的调用会被终止，不会阻塞其它的监控调用函数。并且因为产生一个线程很容易也很快，现在我们事实上更靠近了每隔 5 秒钟检查这些系统。

当然，上述这个解决方案也有其他问题，比如在一个 goroutine 的 panic 可能会破坏其他 goroutine，睡眠时间有少量的偏差，代码并不像看到的那样模块化等等。但是生成内核级线程不是很容易吗？

正如你所看到的，Java 程序需要 12 行代码，而 Go 却只需要 2 个单词。我们想说这意味着你的 Go 源代码将变得简洁和紧凑，但是当我们在这篇文章后面介绍 panics 和 errors 时，你会发现, 不幸的是, 情况并非如此。(事实上 Go 语言是一种比较臃肿的语言，现在信不信由你)。    
### 同步包(和 Channels)加上编排

我们使用 Go 的同步包和 channels 是为了 goroutine 的编排，发送信号和关闭。你将会发现对 sync.Mutex 和 sync.WaitGroup 的引用以及不断重复的名为 shutdownChannel 的结构体变量充斥在我们的代码中。

在 [Go manual](https://golang.org/pkg/sync/) 中关于 sync.Mutex 和 sync.WaitGroup 有一个重要的提示:    

>> 不应拷贝包含此包中定义的类型的值。

整天给那些不用 Go，C 或 C++ 的人解释：结构体是值传递的。任何时间你创建一个 Mutex 或者 WaitGroup，使用一个指针，不要使用直接的值。这并不是普遍必需的，但是如果你不知道何时是好是坏，请一直使用指针。这里有一个很好的、简单的例子：

```go
type Example struct {
	wg *sync.WaitGroup
	m *sync.Mutex
}

func main() {
	wg := &sync.WaitGroup{}
	m := &sync.Mutex{}
}
```
然而对于这些结构体的警告恰恰在页面的顶部，很容易忽视掉，这将会导致你将要构建的应用出现奇怪的副作用。

从前一节关于监控的那个例子来看，这个例子是我们如何使用一个 WaitGroup 来确保在任何时候每个系统中不会有多于一个的监控:

```go
import "sync"
…
func main() {
	for !shutdown {
		wg := &sync.WaitGroup{}

		doCall := func(fn func()) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fn()
			}
		}

		doCall(monitorCpu)
		doCall(monitorDisk)
		doCall(monitorNetwork)
		doCall(monitorProcesses)
		doCall(monitorIdentity)

		wg.Wait()
	}
}
```
一个 mutex 是保护共享资源的好方法，比如服务器上的 CPU 度量历史，正在监视日志文件的水印，或对更新事件感兴趣的监听器列表。这里没有惊奇的地方，除了一个相当令人可怕的 `defer` 关键字外，但是它超出这篇文章的范畴了。

```go 
package main

import (
	"fmt"
	"sync"
	"time"
)

func printAndSleep(m *sync.Mutex, x int) {
	m.Lock()
	defer m.Unlock()
	fmt.Println(x)
	time.Sleep(time.Second)
}

func main() {
	m := &sync.Mutex{}
	for i := 0; i < 10; i++ {
		printAndSleep(m, i)
	}
}
```
goroutine 不仅容易启动，而且总的来说它们也容易协调、关闭和等待。不过有几个复杂的事情要处理，它们可能会更加深入一些：对多线程的事件广播、工作池，还有分布式处理。

### Goroutines 的自动清理

goroutine 持有堆变量和栈变量的引用（避免垃圾收集），但是并不需要一直持有这些引用。它们 (goroutines) 一直运行直到函数完成，然后关闭并且自动释放所有资源。在这过程中，一个需要注意的点是：如果主线程已经退出，那么启动的 goroutine 会被忽略。

首先，这里举一个有关启动 goroutine 并被忽略的真实例子。在我们的应用中，我们在子进程中启动一个模块，并且使用 IPC 来配置更新，设置更新和心跳检测。父进程和每个模块进程(子进程)必须不断地从 IPC channel 中读取数据，并且再向别的地方发送数据。这是一种我们启动并且忽略的线程，因为在关机的时候，我们不关心是否已经接收了全部的信息流。请持保留态度的看这段代码，虽然它来自我们的代码，但是为了简单起见，移除了一些重量级的代码：

```go
package ipc

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"time"
)

type ProtocolReader struct {
	Channel chan *ProtocolMessage
	reader  *bufio.Reader
	handle  *os.File
}

func NewProtocolReader(handle *os.File) *ProtocolReader {
	return &ProtocolReader{
		make(chan *ProtocolMessage, 15),
		bufio.NewReader(handle),
		handle,
	}
}

func (this *ProtocolReader) ReadAsync() {
	go func() {
		for {
			line, err := this.reader.ReadBytes('\n')
			if err != nil {
				this.handle.Close()
				close(this.Channel)
				return nil
			}

			message := &ProtocolMessage{}
			message.Unmarshal(line)
			this.Channel <- message
		}

		return nil
	}
}
```

第二个例子用来说明主线程退出而忽略正在运行的 goroutine：

```go
package main

import (
	"fmt"
	"time"
)

func waitAndPrint() {
	time.Sleep(time.Second)
	fmt.Println("got it!")
}

func main() {
	go waitAndPrint()
}
```
使用 `sync.WaitGroup` 很容易修复这个问题。你将会看到很多像这个代码示例一样使用 time.Sleep 来等待的例子。如果你助长了这种疯狂(使用 time.Sleep)，我们确实会小看你的。请使用 WaitGroup 来写代码吧。

### Channels

Go 语言的 channel 是很好的单向消息传递工具。我们在代理软件中使用它们用于消息传递，消息广播和工作队列。它们会被 GC 自动清理，不需要(手动)关闭，并且很容易生成：

```go
numSlots := 5
make(chan int, numSlots)
```

你可以通过这个 channel 传送任何信息。你可以让它们同步工作，异步工作，或者让多个读端来监听这些 channels，并做具体的一些工作。

不像 queue，一个 channel 可以被用来广播一个消息。在我们的代码中，最经常用来广播的消息是关闭。当到了关闭时刻，我们向所有后台 goroutines ->发送广播：到清理空间的时间了。使用一个 channel 向多个监听者发送单一消息仅仅只有一种方式：那就是你必须关闭这个 channel。下面是我们代码的简化版本：

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

var shutdownChannel = make(chan struct{}, 0)
var wg = &sync.WaitGroup{}

func start() {
	wg.Add(1)
	go func() {
		ticker := time.Tick(100*time.Millisecond)

		for shutdown := false; !shutdown; {
			select {
			case <-ticker:
				fmt.Println("tick")
			case <-shutdownChannel:
				fmt.Println("tock")
				shutdown = true
			}
		}
		wg.Done()
	}()
}

func stop() {
	close(shutdownChannel)
}

func wait() {
	wg.Wait()
}

func main() {
	start()
	time.Sleep(time.Second)
	stop()
	wait()
}
```

我们对 Go 中的 `select` 功能情有独钟。它允许我们在做重要工作的同时响应中断。我们可以相当自由的使用它来管理关闭信号和定时器（像上例那样），读取多个数据流，并且可以和 Go 的 [fsnotify](https://github.com/fsnotify/fsnotify) 包一起使用。

## Go 语言的糟糕之处 1：Panic 和 Error 的处理

panic 和 error，它们是 Go 语言中最糟糕的东西，而且会是一个长远的问题。首先，让我们来为 panic 和 error 下个定义，因为并不是每一门编程语言都会处理它们。

根据 [Go 官方博客的一篇文章](https://blog.golang.org/defer-panic-and-recover)，

> Panic 是一个内建函数，用于终止普通的控制流，并使程序崩溃。当函数 F 引发 panic 时，F 的执行终止，通常函数 F 中的任一 deferred 函数会被执行，并且 F 会返回给它的调用者。对于调用者来说，F 表现得像是一个调用 panic 的函数。这个进程继续收回栈帧（栈是向下增长，函数返回时退栈）直到当前 goroutine 中所有的函数返回，这时程序就崩溃了。Panics 可以通过直接调用来引发。它们也可以由运行时错误引发，如数组访问越界。

换句话说，当你遇到一个控制流问题时，panics 会终止你的程序。

有几种方式可以触发一个 panic：

- 调用函数来引发 panic
- 除 0
- 关闭一个已经关闭的 channel
- 映射不存在的属性，比如 `Attribute = map["This doesn’t exist"]`

另一方面，error 是一个內建类型，这种类型表示能自声明为字符串类型的值。这是从 Go 源代码引用的定义：

```go
type error interface {
	Error() string
}
```
根据以上定义，这是对于为什么我们讨厌 Go 拥有 error 和 panic 的总结：

#### Error 是为了避免异常流，而 panic 抵消了这种作用

对于任何一种编程语言，只要拥有 error 和 panic 其中之一就足够了。至少可以说，一些编程语言兼有二者是令人沮丧的。Go 语言的开发者们不幸地跟错了潮流，错误地选择了兼有二者。

### 一份在流行编程语言中错误处理的抽样

<table>
<tbody>
<tr>
<td>Golang</td>
<td>panic(实际上更像 error)，exceptions & segfault</td>
</tr>
<tr>
<td>Java</td>
<td>exceptions</td>
</tr>
</tr>
<tr>
<td>Scala</td>
<td>exceptions</td>
</tr>
</tr>
<tr>
<td>Ruby</td>
<td>error(实际上更像 exceptions)</td>
</tr>
</tr>
<tr>
<td>Python</td>
<td>exceptions</td>
</tr>
<tr>
<td>PHP</td>
<td>error & exceptions</td>
</tr>
<tr>
<td>Javascript</td>
<td>exceptions</td>
</tr>
<tr>
<td>C/C++</td>
<td>error,exceptions & segfault</td>
</tr>
<tr>
<td>Objective-C</td>
<td>exceptions & error</td>
</tr>
<tr>
<td>Swift</td>
<td>error</td>
</tr>
</tbody>
</table>

总有可能返回错误，但对语言来说可能并不必要。Go 的很多內建函数，比如访问 map 元素、从 channel 中读取数据、JSON 编码等，都需要用到错误处理。这就是为什么 Go 和 其他一些类似的语言接纳 "error" 这个设计，而像 Python 和 Scala 等语言却没有。

[Go 语言官方博客](https://blog.golang.org/error-handling-and-go)再次介绍: 

> 错误处理是重要的。该语言的设计和约定鼓励你明确地检查错误发生的地方（区别于其它语言约定中的抛出异常和有时捕获它们）。在某些情况下，这使得 Go 代码冗长，但幸运的是，你可以使用一些技术来最小化重复的错误处理。

当他们说 error 可以使 Go 代码冗长，他们没有撒谎。

所以从 Go 语言对 panic 和 error 的实现中，我们可以了解到一些什么呢？

### Error 增加你的代码的大小

在开始介绍之前，让我们点明我们可以对 panic 和 error 抱怨。不是我们不喜欢 error，而是我们不喜欢兼有 error 和 panic。根据语言的设计，error 可以很容易的去除，所以我们主要抱怨 error 相关的问题，而不是 panic。

在代码库大小方面，Go 应用已经很臃肿了。它的二进制文件运行速度很快，但是作为源代码，它比它需要的更加冗长。在这种冗长的基础上，还需要同时处理 panic 和 error，由此你就会明白为什么这门语言之前被[称为丑陋的东西](https://www.quora.com/Do-you-feel-that-golang-is-ugly)有了一个概念。在区分 panic 和 error 方面需要花费额外的努力。在那些典型的编程语言里，你只有一种处理错误的方式。那如果有 error 和 panic 呢？这让我们比在假人模特工厂里的蚊子更加抓狂！

这里是原因。

通常管理 error 的方式是 try/catch。

下面是本应该出现在我们代理软件代码中的例子：

```go
try {
	this.downloadModule(moduleSettings)
	this.extractModule(moduleSettings)
	this.tidyManifestModule(moduleSettings)
	this.restartCommand(moduleSettings)
	this.cleanupModule(moduleSettings)
	return nil
}
catch e {
	case Exception => return e
}
```
在这个例子中，和我们实际的代码一样，我们不关心 error 是什么或者它在哪里发生。所有我们关心的是是否有一个 error。这种方式既有意义，又能产生整洁简明的代码。

对于一个错误，你的代码的每一行都必须这么做: 

```go
if err := this.downloadModule(moduleSettings); err != nil {
	return err
}
if err := this.extractModule(moduleSettings); err != nil {
	return err
}
if err := this.tidyManifestModule(moduleSettings); err != nil {
	return err
}
if err := this.restartCommand(currentUpdate); err != nil {
	return err
}
if err := this.cleanupModule(moduleSettings); err != nil {
	return err
}
```
在最糟糕的情况下，它将使你的代码库增加至三倍。三倍！不，它不是每一行 - 结构体，接口，导入库和空白行完全不受影响。所有的其他行，你知道的，有实际代码的行，都增至三倍。

在少数情况下，你想为这些做出不同。在不必要的时候，把一行代码变成三行是件很蠢的事情。没有任何理由去扩展代码库的大小。

除了 error，你还有 panic（需要应对）。如果你有引发 panic 的事物，它和在哪个函数之内发生 (panic) 没有关系.如果你有一个 panic，它将和 try/catch 做同样的事情，除了你现在需要重复代码外，你还是需要处理同样的 try/catch！

我们的代理软件中的解决方案是使用一个带有重试功能的 wrapper 函数和为 panic 和 error 而做的良好记录。然后我们在主线程和遍及代码产生的每一个 goroutine 中严格的调用它。这些代码不能在其他的任何地方运行，因为他缺失其它类库，而且它是阉割版的代码，但是它应该给你关于如何共同管理 panic 和 error 的一些启发。

```go
package safefunc

import (
	"common/log"
	"common/timeout"
	"runtime/debug"
	"time"
)

type RetryConfig struct {
	MaxTries           int
	BaseDelay          time.Duration
	MaxDelay           time.Duration
	SplayFraction      float64
	ShutdownChannel    <-chan struct{}
}

func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxTries:           -1,
		BaseDelay:          time.Second,
		MaxDelay:           time.Minute,
		SplayFraction:      0.25,
		ShutdownChannel:    nil,
	}
}

func Retry(name string, config *RetryConfig, callback func() error) {
	// this is stupid, but necessary.
	// when a function panics, that function's returns are zeros.
	// that's the only way to check (can't rely on a nil error during a panic)
	var noPanicSuccess int = 1
	failedAttempts := 0

	wrapped := func() (int, error) {
		defer func() {
			if err := recover(); err != nil {
				log.Warn.Println("Recovered panic inside", name, err)
				log.Debug.Println("Panic Stacktrace", string(debug.Stack()))
			}
		}()

		return noPanicSuccess, callback()
	}

retryLoop:
	for {
		wrappedReturn, err := wrapped()
		if err != nil {
			log.Warn.Println("Recovered error inside", name, err)
			log.Debug.Println("Recovered Stacktrace", string(debug.Stack()))
		} else if wrappedReturn == noPanicSuccess {
			break retryLoop
		}

		failedAttempts++
		if config.MaxTries > 0 && failedAttempts >= config.MaxTries {
			log.Trace.Println("Giving up on retrying", name, "after", failedAttempts, "attempts")
			break retryLoop
		}

		sleep := timeout.Delay(config.BaseDelay, failedAttempts, config.SplayFraction, config.MaxDelay)
		log.Trace.Println("Sleeping for", sleep, "before continuing retry loop", name)
		sleepChannel := time.After(sleep)
		select {
		case <-sleepChannel:
		case <-config.ShutdownChannel:
			log.Trace.Println("Shutting down retry loop", name)
			break retryLoop
		}
	}
}
```
当因为 Go 有这些错误处理让你觉得代码很安全时，一个 panic 错误在运行时发生了。它不做任何检查，把各 goroutine 的所有的错误都放到堆栈里，直到最终使程序崩溃。

### 并非每人都讨厌 Go 中的 error 和 panic

诚然，有些人可能不认为这是一个弊端。甚至在我们 Blue Matador 自己公司内部都有不同意见（联合创始人们之间有更多的意见）

有错误处理会强迫你在错误发生的时候处理它们，而不是在未来一个未知的时候。error 可以是类型检查的，所以编译器可以让你去处理它们并且警告你。Go 是为数不多支持这一特性的语言。

一些人喜欢 panic，因为它不是依靠开发者的能力来返回一个异常。try/catch 是解决这些更好的方式，并且它是在函数之外的。在异常发生时，开发人员很难跟踪到，panic 和 error 可以帮助开发者解决这个问题。

### 我们对 error 和 panic 的主要抱怨

在 Go 中，你不得不同时处理这两者。你已经通过 try/catch 获得错误捕获的 panic，所以为什么你还要让自己必须去担心 panic 呢？

panic 不仅中断所有导致 panic 的函数调用，而且也中断线程！如果你有一个出现 panic 的线程，并且你不在那个线程捕获它，这样的话不仅那个线程停止了，而且调用它的线程也会停止，一直出现异常直到你的程序挂掉。你代码中的单个 panic 可以中断所有事情，因为它可以级联的导致整个程序执行失败。

Go 使你不得不去容忍 error 和 panic 同时存在，而你最好永远记住这个。如果你一旦忘记，你的程序可能会意外的崩溃。你可能将此归咎于那个正确写出糟糕代码的程序员，但是它真的是开发者的错误吗？当机修工拧松车上的所有螺帽，难道是司机的错误吗？

总结来说，如果一切都被 error 捕获，那太好了！如果一切都被 panic 捕获，那也很好！但是在 Go 中，必须对两者进行处理和同时容忍同时存在确实是令人沮丧的。

Golang 对于 DevOps 之利弊第二篇：自动接口实现，共有/私有变量    
每隔一周，我们会发布一篇新的指南，如我们的关于 Golang 对于 DevOps 之利弊六部曲中本篇一样。下一篇是：自动接口，共有/私有变量。

#### 作者简介

Matthew 经历过标准 DevOps 监控工具需要随时待命之痛。他创立了 Blue Matador 来修复遍及 DevOps 的 Frankenstein 安装的混乱。业余时间，他喜欢开飞机，玩棋盘游戏，还有花时间陪他的妻子和三个孩子。

---------

via: https://blog.bluematador.com/posts/golang-pros-cons-for-devops-part-1-goroutines-panics-errors/

作者：[Matthew Barlocker](https://github.com/mbarlocker)
译者：[liuxinyu123](https://github.com/liuxinyu123)
校对：[rxcai](https://github.com/rxcai)、[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出