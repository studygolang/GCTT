首发于：https://studygolang.com/articles/19580

# Go 语言的并发性

昨天，我在 Quora 上回答了一个关于 Go 语言并发模型的问题。现在，我觉得我还想再多说些什么！并发性是 Go 语言中最强大的特性之一。许多人讨论了这个话题，从非常简单到过于复杂的都有。今天，我也来说说我的看法。

Go 语言的并发性是一种思维方式而不仅仅是一个语法。为了利用 Go 的强大功能，你需要首先了解 Go 是如何实现代码的并发执行。Go 依赖于一个叫做 CSP（Comminicating Sequential Process 通信顺序进程）的并发模型，在计算机科学中，它基本上是描述并发系统之间的交互模型。但是鉴于这不是一篇科学论文，我会跳过那些繁琐的过程直接介绍它的实际用途。

许多关于 Go 的演讲、演示和文献在解释 Go 的并发性时都会用到如下短语：

## 不要通过共享内存来通信，而是通过通信来共享内存：

听起来真不错。但是这到底是什么意思呢？这花了我好一会儿的时间才理解这个概念。但是一旦我了解了这个概念，Go 语言的编程对我来说就更加流畅了。Albert Einstein 曾经说过，如果你不能把它解释得通俗易懂，那么你就还没有完全了解它。以下是我能想到的对于这句话的最简单的解释了。

### 不要通过共享内存来通信

在主流的编程语言中，当你想到代码的并发执行时，你通常会想到利用多线程并行地执行一些复杂操作。通常，你会需要在线程之间共享数据结构、变量、内存等等。你会利用锁操作来避免两个线程同时访问或者写一块内存，或者你就让这块内存放任自由而不加以任何限制并期待能有最好的结果。这是我们在大多数比较流行的编程语言中所用到的线程间的通信方式，但是这通常会导致各种各样的问题，比如竞态条件、内存管理、随机奇怪无法解释的异常和你的彻夜难眠等等。

### 替换方式：通过通信来共享内存

那么，Go 语言是怎么做到这个的呢？ Go 允许你发送变量的值来给其他线程（实际上，这不是一个实际意义上的线程，但是现在姑且可以这么理解），来代替用锁住变量的共享内存的方式。默认的行为是，发送数据的线程和接收数据的线程都会等待直到数据到达它的目的地。线程的“等待”强制线程之间在交换数据时进行适当的同步。在你撸起袖子开始代码设计之前，先想想并发性的这种实现方式。这样你将会有更加稳定的软件。

为了解释得更清晰一点：是如何保证软件的稳定呢？默认情况下，发送线程和接收线程在完成值传输的过程中都不会执行任何操作。这意味着，在另一个线程处理数据之前，其中一个线程同时处理数据、出现竞态等类似的问题的机会并不多。

这是 Go 语言的原生特性，你可以直接使用它而不必调用额外的库或者框架，这个行为已经内嵌到语言当中。如果你有需要的话，Go 还能给你提供一个缓冲通道。这意味着某些情况下，在一个值被发送之前，你不希望两个线程同时上锁或者进行同步。你可能希望在你在两个线程之间的通道上面填上一些预定义的值并等待被处理时进行同步。

不过还是要提醒大家，这个模型可能会被过度使用。你必须知道这何时应该使用它，或者何时恢复到良好的旧的共享内存模型。例如，引用计数最好在锁中保护，文件访问也是如此。Go 语言也会通过同步包来支持你使用锁保护。

## 代码实现 Go 的并发性

我们来谈谈代码相关的，我们要如何实现“通过通信共享”模型？请继续往下读：

在 Go 里，'goroutine' 就是作为上面提到的所谓的线程。实际上，这并不能称为线程，这只是一个可以和其他 'goroutine' 并发地在同一个地址空间上面的函数。它们在 O.S 线程中被多路复用，因此如果有一个被阻塞了，其他的仍然可以继续运行。所有的同步和内存管理都由 Go 本地执行。之所以说它们不是真正的线程，是因为它们并不一定总是要并行执行。然而，由于多路复用和同步，你会得到并发的效果。要启动一个新的 'goroutine' ，你只需要使用关键字 "go" ：

```go
go processdataFunction()
```

“ Go 通道”是 Go 语言实现并发性的另一个概念。这个通道是用于不同的 Goroutine 之间的内存交流。要创建一个通道，就要用到 "make" 关键字：

```go
myChannel := make(chan int64)

```
在 goroutine 等待之前，创建一个缓冲通道来允许更多的值在通道中排队，如：

```go
myBufferedChannel := make(chan int64,4)
```

在上面的两个例子中，我假设通道的变量没有提前定义。这就是我为什么使用 ":=" 来创建提及类型的变量而不是使用 "=", 因为 "=" 只会进行赋值，如果变量使用之前没有声明，就会导致编译错误。

现在到了使用通道的时候了，捏可以使用 "<-" 符号。发送该值的 Goroutine 会将其分配给通道，如下所示 :

```go
myChannel <- 54
```

另一个 Goroutine 接收到这个值，会将它从通道中取出来并且重新赋给一个新的变量：

```go
myVar := <- mychannel
```

现在让我们看一个例子来展示 Golang 中的案例并发性：

 ```go
package main
import (
	"fmt"
	"time"
)
func main() {
	ch := make(chan int)
	//create a channel to prevent the main program from exiting before the done signal is received
	done := make(chan bool)
	go sendingGoRoutine(ch)
	go receivingGoRoutine(ch,done)
	//This will prevent the program from exiting till a value is sent over the
    "done" channel, value doesn't matter
	<- done
}

func sendingGoRoutine(ch chan int){
	//start a timer to wait 5 seconds
	t := time.NewTimer(time.Second*5)
	<- t.C
	fmt.Println("Sending a value on a channel")
    //this Goroutine will wait till another Goroutine received the value
    ch <- 45
}

func receivingGoRoutine(ch chan int, done chan bool){
	//this Gourtine will wait till the channel received a value
    v := <- ch
	fmt.Println("Received value ", v)
	done <- true
}
```

输出：

```
Sending a value on a channel
Received value  45
```

---

via：http://www.minaandrawos.com/2015/12/06/concurrency-in-golang/

作者：[Mina](http://www.minaandrawos.com/about-me/)
译者：[pumpkinmonkey](https://github.com/pumpkinmonkey)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
