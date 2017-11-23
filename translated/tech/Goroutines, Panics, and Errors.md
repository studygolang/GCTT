# Golang 对于 DevOps 之利弊(第 1 部分，共 6 部分)：Goroutines, Panics, and Errors  
![][1] 
[1]: https://camo.githubusercontent.com/92a4d1155a66df2b93aa537cff26837267f4952d/68747470733a2f2f626c6f672e626c75656d617461646f722e636f6d2f68756266732f426c75655f4d617461646f725f496e635f4f63746f626572323031372f496d616765732f676f6c616e672d70726f732d636f6e732d666f722d6465766f70732d706172742d312d676f726f7574696e65732d70616e6963732d6572726f72732e706e673f743d31353131323832353139383332     

对于你的下一个 DevOps 应用来说，Google 公司的 Go 可能是完美的语言。作为由 6 篇组成一个系列文章的第一篇，我们从 goroutines, panics, and errors 开始深入研究 Go 语言的利与弊，因为这些利与弊涉及构建 DevOps 应用。      
在这篇博客中，我们已经称赞了 Google 公司的 Go 语言的优点，而且我们也写了[Go 迷你入门指南](https://blog.bluematador.com/posts/mini-guide-google-golang-why-its-perfect-for-devops/?utm_source=bm-blog&utm_medium=link&utm_campaign=golang-pros-cons-1)。 
但是以免任何人认为我们假借 DevOps 监控平台(我们是)的地下 Google 员工(我们不是)，我们更愿意去研究语言的利与弊，特别是当它涉及构建 DevOps 应用时。由于我们将智能代理(软件)从 Python(构建) 转换到 Go(构建)，我们有许多在 Go 语言环境下工作(开发)的经验，并且我们也有一些自认为很重要的东西(知识)来与更大的 DevOps 社区分享。    
从这周开始，我们将发布由 6 篇组成一系列关于 Go 语言利和弊文章，每一篇文章都详述了少许的内容。正如我们所做的，我们将通过链接到其它篇目来更新这篇文章：     

- Golang 对于 DevOps 之利弊第一篇:Goroutines, Channels, Panics, and Errors(本篇)    
- [Golang 对于 DevOps 之利弊第二篇:自动接口实现，共有/私有变量](https://blog.bluematador.com/posts/golang-pros-cons-for-devops-part-2)   
- [Golang 对于 DevOps 之利弊第三篇:速度 VS 缺少泛型](https://blog.bluematador.com/posts/golang-pros-cons-devops-part-3-speed-lack-generics)   
- [Golang 对于 DevOps 之利弊第四篇:时间包与方法重载](https://github.com/studygolang/GCTT/blob/master/golang-pros-cons-part-4-time-package-method-overloading)   
- Golang 对于 DevOps 之利弊第五篇:交叉编译，窗口，信号，文档和编译器    
- Golang 对于 DevOps 之利弊第六篇:Defer 语句和包依赖版本控制    
如果本篇是你的有关 Go 的首次读物，并且你已经知道怎样用类 C 的语言进行编程，你应该去参考[Go 之旅](https://tour.golang.org/welcome/1)，它将大约花费你一个小时，而且(介绍)相当有深度。接下来的评论不是学习如何用 Go 进行编程的好方法。相反，它们(评论)只是我们在用 Go 开发智能代理(软件)时的怨言和我们发现的可取之处。   
准备好听一下用 Go 进行编程的真正样子吗？那我们开始吧。   
## Go 语言的好处 1: Goroutines — 轻量，内核级线程    
---    
Goroutines 等同于 Go。Goroutines 是执行线程，并且它们是轻量的和内核级的。轻量级是因为你可以在不影响系统性能的同时运行很多的 Goroutine，而内核级是因为它们并行运行(真正的并行运行，而不是像在其他语言中的伪并行，比如旧版的 Ruby)。同时也不像 Python那样，在 Go 中没有全局解释器锁(GIL)。    
换句话说，Goroutines 是 Go 语言中的重要部分，而不是后来才添加的特性，这正是 Go 语言执行之快的原因之一。另一方面，Java 有一个复杂的线程调度器。它们在为线程的执行创造条件，但是需要很多的管理工作，因此降低了你的程序运行速度。从另一个角度看，Node.js 没有真正的并发，仅仅是并发的假象。这是对于 DevOps 用途来说，Go 能成为一门伟大的语言的原因，它不影响你的系统性能，真正的并行执行，并且在意外的硬件故障中尽可能的可靠。     
### 怎样执行一个 Goroutine    
启动一个 Goroutine 线程超级简单。首先，你将在我们的看门狗产品中看到一些伪代码：   
```   
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
在上面的代码中，CPU，硬盘，网络，进程和身份每 5 秒检查一次。如果上述其中之一挂掉的话，所有的监控就会终止。每一个监控用时越长，检查变得越不频繁，因为我们在计算完成之后需要睡眠 5 秒。(ps.最后一句没看懂)
为了解决这些问题，一种解决方案(一种不完整的方案，但是完美展示 goroutine 的价值所在)是使用 goroutine 来调用每一个监控函数。仅仅在你想要以线程方式产生的函数调用前加上 `go`关键字。     
```
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
正如你所看到的，Java 程序需要 12 行代码，而 Go 却只需要 2 个单词。我们想说这意味着你的 Go 源代码将变得简洁和紧凑，但是当我们在这篇文章之后介绍的 panics 和 errors 时，你将会看到不幸的是并非如此(事实上 Go 语言是一种比较臃肿的语言，现在信不信由你)。    
### 同步包(和 Channels)加上编配    
我们使用 Go 的同步包和 channels 是为了 goroutine 的编配，发送信号和关机。你将会发现对 sync.Mutex 和 sync.WaitGroup 的引用以及一种称为 shutdownChannel 的重复的结构体变量充斥在我们的代码中。    
在 [Go manual](https://golang.org/pkg/sync/) 中关于 sync.Mutex 和 sync.WaitGroup 有一个重要的提示:    
>> 不应拷贝包含此包中定义的类型的值。      

解释给那些整天不使用 Go，C 或者 C++ 的人(ps.拿不准):结构体是值传递的(ps.感觉应该是引用传递)。任何时间你创建一个 Mutex 或者 WaitGroup，使用一个指针，不要使用直接的值。这并不是普遍必需的，但是如果你不知道何时是好是坏，请一直使用指针。这里有一个很好的，简单的例子:    
```
type Example struct {
	wg *sync.WaitGroup
	m *sync.Mutex
}

func main() {
	wg := &sync.WaitGroup{}
	m := &sync.Mutex{}
}   
```   
然而对于这些结构体的警告恰恰在页面的顶部，很容易掩饰过去，将会导致你将要构建的应用出现奇怪的副作用。   
从前一节关于监控的那个例子来看，这个例子是我们如何使用一个 WaitGroup 来确保在任何时候每个系统中不会有多于一个的监控:   
```
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
一个 mutex 是保护共享资源的好方法，比如服务器上的CPU度量历史，正在监视日志文件的水印，或对更新事件感兴趣的监听器列表。这里没有惊奇的地方，除了一个相当令人可怕的 `defer` 关键字外，但是它超出这篇文章的范畴了。   
``` 
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
goroutine(s) 不仅容易启动，而且总的来说它们也容易协调，关闭和等待。不过有几个复杂的事情要处理，它们可能会更加深入一些:对多线程的事件广播，工作池，还有分布式处理。   
### Goroutines 的自动清理      
goroutine(s) 持有堆变量和栈变量的引用(避免垃圾收集)，但是并不需要一直持有这些引用。它们(goroutines)一直运行直到函数完成，然后关闭并且自动释放所有资源。在这个过程中有一个需要注意的部分是如果主线程已经退出，启动一个 goroutine 会被忽略(或者跳过)。 
首先，这里举一个有关启动 goroutine 并被忽略的真实例子。在我们的应用中，我们在子进程中启动一个模块(函数)，并且使用 IPC 来配置更新，设置更新和心跳。父进程和每个模块进程(子进程)必须不断地从 IPC channel 中读取数据，并且彼此向对方发送数据。这是一种我们启动并且忽略的线程，因为我们不关心如果(对方已经)关闭，我们不能读取完整的流。请持保留态度的看这段代码，虽然它来自我们的代码，但是为了简单起见，移除了一些重量级的代码:   
```
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
第二个例子用来说明主线程退出而忽略正在运行的 goroutine:   
```
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
使用 `sync.WaitGroup` 很容易修复这个问题。你将会看到很多像这个代码示例一样使用 time.Sleep 用来等待的例子。如果你助长了这种疯狂(使用 time.Sleep)，我们确实会小看你的。请使用 WaitGroup 来写代码吧。     
## Channels      
Go 语言的 channel(s) 是很好的单向消息传递工具。我们在代理软件中使用它们用于消息传递，消息广播和工作队列。它们会被 GC 自动清理，不需要(手动)关闭，并且很容易生成：  
```
numSlots := 5
make(chan int, numSlots)   
```   
你可以通过这个 channel 传送任何一种目标。你可以让它们同步工作，异步工作，或者让多个读端来监听这些 channels，并做具体的一些工作。    
不像 queue，一个 channel 可以被用来广播一个消息。在我们的代码中，最经常用来广播的消息是关闭。当到了关闭时刻，我们向所有后台 goroutines 发送广播:到清理空间的时间了。使用一个 channel 向多个监听者发送单一消息仅仅只有一种方式，那就是你必须关闭这个 channel。下面是我们代码的简化版本:    
```    
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
我们对 Go 中的 `select` 功能情有独钟。它允许我们在做重要工

列表项

作的同时可以响应中断。我们可以相当自由的使用它来管理关闭信号和定时器(像上例那样)，读取多个数据流，并且可以和 Go 的 [fsnotify](https://github.com/fsnotify/fsnotify) 一起使用。






    