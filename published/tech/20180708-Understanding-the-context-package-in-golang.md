首发于：https://studygolang.com/articles/13866

# 理解 golang 中的 context（上下文） 包

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/understanding-the-context-package-in-golang/0_exTPQ4ppfrdjuXcR.jpg)

Go 中的 context 包在与 API 和慢处理交互时可以派上用场，特别是在生产级的 Web 服务中。在这些场景中，您可能想要通知所有的 goroutine 停止运行并返回。这是一个基本教程，介绍如何在项目中使用它以及一些最佳实践和陷阱。

要理解 context 包，您应该熟悉两个概念。

在转到 context 之前，我将简要介绍这些内容，如果您已经熟悉，则可以直接转到 context 部分。

## Goroutine

来自 Go 语言官方文档："goroutine 是一个轻量级的执行线程"。多个 goroutine 比一个线程轻量所以管理它们消耗的资源相对更少。

Playground: https://play.golang.org/p/-TDMgnkJRY6

```go
package main
import "fmt"

//function to print hello
func printHello() {
		fmt.Println("Hello from printHello")
}
func main() {
		//inline goroutine. Define a function inline and then call it.
		go func(){fmt.Println("Hello inline")}()
		//call a function as goroutine
		go printHello()
		fmt.Println("Hello from main")
}
```

如果您运行上面的程序，您只能看到 main 中打印的 Hello, 因为它启动了两个 goroutine 并在它们完成前退出了。为了让 main 等待这些 goroutine 执行完，您需要一些方法让这些 goroutine 告诉 main 它们执行完了，那就需要用到通道。

## 通道（channel）

这是 goroutine 之间的沟通渠道。当您想要将结果或错误，或任何其他类型的信息从一个 goroutine 传递到另一个 goroutine 时就可以使用通道。通道是有类型的，可以是 int 类型的通道接收整数或错误类型的接收错误等。

假设有个 int 类型的通道 ch，如果你想发一些信息到这个通道，语法是 ch <- 1，如果你想从这个通道接收一些信息，语法就是 var := <-ch。这将从这个通道接收并存储值到 var 变量。

以下程序说明了通道的使用确保了 goroutine 执行完成并将值返回给 main 。

注意：WaitGroup（ https://golang.org/pkg/sync/#WaitGroup ）也可用于同步，但稍后在 context 部分我们谈及通道，所以在这篇博客中的示例代码，我选择了它们。

Playground: https://play.golang.org/p/3zfQMox5mHn

```go
package main

import "fmt"

//prints to stdout and puts an int on channel
func printHello(ch chan int) {
	fmt.Println("Hello from printHello")
	//send a value on channel
	ch <- 2
}
func main() {
	//make a channel. You need to use the make function to create channels.
	//channels can also be buffered where you can specify size. eg: ch := make(chan int, 2)
	//that is out of the scope of this post.
	ch := make(chan int)
	//inline goroutine. Define a function and then call it.
	//write on a channel when done
	go func() {
		fmt.Println("Hello inline")
		//send a value on channel
		ch <- 1
	}()
	//call a function as goroutine
	go printHello(ch)
	fmt.Println("Hello from main")
	//get first value from channel.
	//and assign to a variable to use this value later
	//here that is to print it
	i := <-ch
	fmt.Println("Recieved ", i)
	//get the second value from channel
	//do not assign it to a variable because we dont want to use that
	<-ch
}
```

在 Go 语言中 context 包允许您传递一个 "context" 到您的程序。 Context 如超时或截止日期（deadline）或通道，来指示停止运行和返回。例如，如果您正在执行一个 web 请求或运行一个系统命令，定义一个超时对生产级系统通常是个好主意。因为，如果您依赖的API运行缓慢，你不希望在系统上备份（back up）请求，因为它可能最终会增加负载并降低所有请求的执行效率。导致级联效应。这是超时或截止日期 context 派上用场的地方。

## 创建 context

context 包允许以下方式创建和获得 context：

### context.Background() Context

这个函数返回一个空 context。这只能用于高等级（在 main 或顶级请求处理中）。这能用于派生我们稍后谈及的其他 context 。

```go
ctx := context.Background()
```

### context.TODO() Context

这个函数也是创建一个空 context。也只能用于高等级或当您不确定使用什么 context，或函数以后会更新以便接收一个 context 。这意味您（或维护者）计划将来要添加 context 到函数。

```go
ctx := context.TODO()
```

有趣的是，[查看代码](https://golang.org/src/context/context.go)，它与 background 完全相同。不同的是，静态分析工具可以使用它来验证 context 是否正确传递，这是一个重要的细节，因为静态分析工具可以帮助在早期发现潜在的错误，并且可以连接到 CI/CD 管道。

来自 https://golang.org/src/context/context.go:

```go
var (
	background = new(emptyCtx)
	todo = new(emptyCtx)
)
```

### context.WithValue(parent Context, key, val interface{}) (ctx Context, cancel CancelFunc)

此函数接收 context 并返回派生 context，其中值 val 与 key 关联，并通过 context 树与 context 一起传递。这意味着一旦获得带有值的 context，从中派生的任何 context 都会获得此值。不建议使用 context 值传递关键参数，而是函数应接收签名中的那些值，使其显式化。

```go
ctx := context.WithValue(context.Background(), key, "test")
```

### context.WithCancel(parent Context) (ctx Context, cancel CancelFunc)

这是它开始变得有趣的地方。此函数创建从传入的父 context 派生的新 context。父 context 可以是后台 context 或传递给函数的 context。

返回派生 context 和取消函数。只有创建它的函数才能调用取消函数来取消此 context。如果您愿意，可以传递取消函数，但是，强烈建议不要这样做。这可能导致取消函数的调用者没有意识到取消 context 的下游影响。可能存在源自此的其他 context，这可能导致程序以意外的方式运行。简而言之，永远不要传递取消函数。

```go
ctx, cancel := context.WithCancel(context.Background())
```

### context.WithDeadline(parent Context, d time.Time) (ctx Context, cancel CancelFunc)

此函数返回其父项的派生 context，当截止日期超过或取消函数被调用时，该 context 将被取消。例如，您可以创建一个将在以后的某个时间自动取消的 context，并在子函数中传递它。当因为截止日期耗尽而取消该 context 时，获此 context 的所有函数都会收到通知去停止运行并返回。

```go
ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2 * time.Second))
```

### context.WithTimeout(parent Context, timeout time.Duration) (ctx Context, cancel CancelFunc)

此函数类似于 context.WithDeadline。不同之处在于它将持续时间作为参数输入而不是时间对象。此函数返回派生 context，如果调用取消函数或超出超时持续时间，则会取消该派生 context。

```go
ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
```

## 函数接收和使用 Context

现在我们知道了如何创建 context（Background 和 TODO）以及如何派生 context（WithValue，WithCancel，Deadline 和 Timeout），让我们讨论如何使用它们。

在下面的示例中，您可以看到接受 context 的函数启动一个 goroutine 并等待 该 goroutine 返回或该 context 取消。select 语句帮助我们选择先发生的任何情况并返回。

`<-ctx.Done()` 一旦 Done 通道被关闭，这个 `<-ctx.Done():` 被选择。一旦发生这种情况，此函数应该放弃运行并准备返回。这意味着您应该关闭所有打开的管道，释放资源并从函数返回。有些情况下，释放资源可以阻止返回，比如做一些挂起的清理等等。在处理 context 返回时，您应该注意任何这样的可能性。

本节后面的示例有一个完整的 Go 语言程序，它说明了超时和取消功能。

```go
//Function that does slow processing with a context
//Note that context is the first argument
func sleepRandomContext(ctx context.Context, ch chan bool) {
	//Cleanup tasks
	//There are no contexts being created here
	//Hence, no canceling needed
	defer func() {
		fmt.Println("sleepRandomContext complete")
		ch <- true
	}()
	//Make a channel
	sleeptimeChan := make(chan int)
	//Start slow processing in a goroutine
	//Send a channel for communication
	go sleepRandom("sleepRandomContext", sleeptimeChan)
	//Use a select statement to exit out if context expires
	select {
	case <-ctx.Done():
		//If context expires, this case is selected
		//Free up resources that may no longer be needed because of aborting the work
		//Signal all the goroutines that should stop work (use channels)
		//Usually, you would send something on channel,
		//wait for goroutines to exit and then return
		//Or, use wait groups instead of channels for synchronization
		fmt.Println("Time to return")
	case sleeptime := <-sleeptimeChan:
		//This case is selected when processing finishes before the context is cancelled
		fmt.Println("Slept for ", sleeptime, "ms")
	}
}
```

## 例子

到目前为止，我们已经看到使用 context 可以设置截止日期，超时或调用取消函数来通知所有使用任何派生 context 的函数来停止运行并返回。以下是它如何工作的示例：

***main*** 函数

* 用 cancel 创建一个 context
* 随机超时后调用取消函数

***doWorkContext*** 函数

* 派生一个超时 context
* 这个 context 将被取消当
  * main 调用取消函数或
  * 超时到或
  * doWorkContext 调用它的取消函数
* 启动 goroutine 传入派生上下文执行一些慢处理
* 等待 goroutine 完成或上下文被 main goroutine 取消，以优先发生者为准

***sleepRandomContext*** 函数

* 开启一个 goroutine 去做些缓慢的处理
* 等待该 goroutine 完成或，
* 等待 context 被 main goroutine 取消，操时或它自己的取消函数被调用

***sleepRandom*** 函数

* 随机时间休眠
* 此示例使用休眠来模拟随机处理时间，在实际示例中，您可以使用通道来通知此函数，以开始清理并在通道上等待它，以确认清理已完成。

Playground: https://play.golang.org/p/grQAUN3MBlg (看起来我使用的随机种子，在 playground 时间没有真正改变，您需要在你本机执行去看随机性)

Github: https://github.com/pagnihotry/golang_samples/blob/master/go_context_sample.go

```go
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

//Slow function
func sleepRandom(fromFunction string, ch chan int) {
	//defer cleanup
	defer func() { fmt.Println(fromFunction, "sleepRandom complete") }()
	//Perform a slow task
	//For illustration purpose,
	//Sleep here for random ms
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	randomNumber := r.Intn(100)
	sleeptime := randomNumber + 100
	fmt.Println(fromFunction, "Starting sleep for", sleeptime, "ms")
	time.Sleep(time.Duration(sleeptime) * time.Millisecond)
	fmt.Println(fromFunction, "Waking up, slept for ", sleeptime, "ms")
	//write on the channel if it was passed in
	if ch != nil {
		ch <- sleeptime
	}
}

//Function that does slow processing with a context
//Note that context is the first argument
func sleepRandomContext(ctx context.Context, ch chan bool) {
	//Cleanup tasks
	//There are no contexts being created here
	//Hence, no canceling needed
	defer func() {
		fmt.Println("sleepRandomContext complete")
		ch <- true
	}()
	//Make a channel
	sleeptimeChan := make(chan int)
	//Start slow processing in a goroutine
	//Send a channel for communication
	go sleepRandom("sleepRandomContext", sleeptimeChan)
	//Use a select statement to exit out if context expires
	select {
	case <-ctx.Done():
		//If context is cancelled, this case is selected
		//This can happen if the timeout doWorkContext expires or
		//doWorkContext calls cancelFunction or main calls cancelFunction
		//Free up resources that may no longer be needed because of aborting the work
		//Signal all the goroutines that should stop work (use channels)
		//Usually, you would send something on channel,
		//wait for goroutines to exit and then return
		//Or, use wait groups instead of channels for synchronization
		fmt.Println("sleepRandomContext: Time to return")
	case sleeptime := <-sleeptimeChan:
		//This case is selected when processing finishes before the context is cancelled
		fmt.Println("Slept for ", sleeptime, "ms")
	}
}

//A helper function, this can, in the real world do various things.
//In this example, it is just calling one function.
//Here, this could have just lived in main
func doWorkContext(ctx context.Context) {
	//Derive a timeout context from context with cancel
	//Timeout in 150 ms
	//All the contexts derived from this will returns in 150 ms
	ctxWithTimeout, cancelFunction := context.WithTimeout(ctx, time.Duration(150)*time.Millisecond)
	//Cancel to release resources once the function is complete
	defer func() {
		fmt.Println("doWorkContext complete")
		cancelFunction()
	}()
	//Make channel and call context function
	//Can use wait groups as well for this particular case
	//As we do not use the return value sent on channel
	ch := make(chan bool)
	go sleepRandomContext(ctxWithTimeout, ch)
	//Use a select statement to exit out if context expires
	select {
	case <-ctx.Done():
		//This case is selected when the passed in context notifies to stop work
		//In this example, it will be notified when main calls cancelFunction
		fmt.Println("doWorkContext: Time to return")
	case <-ch:
		//This case is selected when processing finishes before the context is cancelled
		fmt.Println("sleepRandomContext returned")
	}
}
func main() {
	//Make a background context
	ctx := context.Background()
	//Derive a context with cancel
	ctxWithCancel, cancelFunction := context.WithCancel(ctx)
	//defer canceling so that all the resources are freed up
	//For this and the derived contexts
	defer func() {
		fmt.Println("Main Defer: canceling context")
		cancelFunction()
	}()
	//Cancel context after a random time
	//This cancels the request after a random timeout
	//If this happens, all the contexts derived from this should return
	go func() {
		sleepRandom("Main", nil)
		cancelFunction()
		fmt.Println("Main Sleep complete. canceling context")
	}()
	//Do work
	doWorkContext(ctxWithCancel)
}
```

## 缺陷

如果函数接收 context 参数，确保检查它是如何处理取消通知的。例如，exec.CommandContext 不会关闭读取管道，直到命令执行了进程创建的所有分支（Github 问题：https://github.com/golang/go/issues/23019 ），这意味着如果等待 cmd.Wait() 直到外部命令的所有分支都已完成，则 context 取消不会使该函数立即返回。如果您使用超时或截止日期，您可能会发现这不能按预期运行。如果遇到任何此类问题，可以使用 time.After 实现超时。

## 最佳实践

1. context.Background 只应用在最高等级，作为所有派生 context 的根。
2. context.TODO 应用在不确定要使用什么的地方，或者当前函数以后会更新以便使用 context。
3. context 取消是建议性的，这些函数可能需要一些时间来清理和退出。
4. context.Value 应该很少使用，它不应该被用来传递可选参数。这使得 API 隐式的并且可以引起错误。取而代之的是，这些值应该作为参数传递。
5. 不要将 context 存储在结构中，在函数中显式传递它们，最好是作为第一个参数。
6. 永远不要传递不存在的 context 。相反，如果您不确定使用什么，使用一个 ToDo context。
7. Context 结构没有取消方法，因为只有派生 context 的函数才应该取消 context。

---

via: https://medium.com/@parikshit/understanding-the-context-package-in-golang-b1392c821d14

作者：[Parikshit Agnihotry](https://medium.com/@parikshit)
译者：[themoonbear](https://github.com/themoonbear)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
