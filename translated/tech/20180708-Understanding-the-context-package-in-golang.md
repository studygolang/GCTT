# 理解 golang 中的上下文包

![go logo](https://github.com/studygolang/gctt-images/understanding-the-context-package-in-golang/0_exTPQ4ppfrdjuXcR.jpg)

go 语言中的上下文包能够在与 API 和缓慢的流程交互中使用，特别是在生产级的 web 服务中。在这些场景中，您可能想要通知所有的 goroutine 停止运行并返回。这是一个基本教程，介绍如何在项目中使用它以及一些最佳实践和陷阱。

要理解上下文包，您应该熟悉俩个概念。

在转到上下文之前，我将简要介绍这些内容，如果您已经熟悉，则可以直接转到上下文部分。

## Goroutine

来自 go 语言官方文档："goroutine 是一个轻量级的执行线程"。多个 goroutine 比一个线程轻量所以管理它们消耗的资源相对更少。

Playground: https://play.golang.org/p/-TDMgnkJRY6

```
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

如果您运行上面的程序，您只能看到主进程中打印的 Hello , 因为它启动了俩个 goroutine 并在它们完成前退出了。为了让主进程等待这些 goroutine 执行完，您需要一些方法让这些 goroutine 告诉主进程它们执行完了，那就需要用到通道。

## 通道

这是 goroutine 之间的沟通渠道。当您想要将结果或错误，或任何其他类型的信息从一个 goroutine 传递到另一个 goroutine 时就可以使用通道。通道是有类型的，可以是 int 类型的通道接收整数或错误类型的接收错误等。

假设有个 int 类型的通道 ch ，如果您想发一些信息到这个通道，这个语法是`ch <- 1`，如果您想接收一些信息从这个通道，语法就是`var := <- ch`。这将从这个通道接收并存储值到 var 变量。

以下程序说明了通道的使用确保了 goroutine 执行完成并将值返回给主进程。

注意：等待组（ https://golang.org/pkg/sync/#WaitGroup ）也可用于同步，但稍后在上下文部分我们谈及通道，所有在这篇博客我选择了它们在我的例子代码中。

Playground: https://play.golang.org/p/3zfQMox5mHn

```
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
    go func(){
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
    i := <- ch
    fmt.Println("Recieved ",i)
    //get the second value from channel
    //do not assign it to a variable because we dont want to use that
    <- ch
}
```
在 go 语言中上下文包允许您传递一个"上下文"到您的程序。上下文像一个超时或截止日期或一个通道去提示停止运行和返回。例如，如果您正在执行一个 web 请求或运行一个系统命令，定义一个超时对生产级系统通常是个好主意。因为，如果您依赖的API运行缓慢，您不希望在系统上备份请求，因为它可能最终会增加负载并降低所有请求的执行效率。导致级联效应。这是超时或截止日期上下文派上用场的地方。

## 创建上下文

上下文包允许以下方式创建和获得上下文：

### context.Background() ctx Context

这个函数返回一个空上下文。这只能用于高等级（在主进程或顶级请求处理中）。这能用于派生我们稍后谈及的其他上下文。
```
ctx, cancel := context.Background()
```

### context.TODO() ctx Context

这个函数也是创建一个空上下文。也只能用于高等级或当您不确定使用什么上下文，或函数以后会更新以便接收一个上下文。这意味您（或维护者）计划将来要添加上下文到函数。
```
ctx, cancel := context.TODO()
```
有趣的是，查看代码（ https://golang.org/src/context/context.go ），它与 background 完全相同。 不同的是，静态分析工具可以使用它来验证上下文是否正确传递，这是一个重要的细节，因为静态分析工具可以帮助在早期发现潜在的错误，并且可以连接到CI/CD管道。

来自 https://golang.org/src/context/context.go:
```
var ( background = new(emptyCtx) todo = new(emptyCtx) )
```

### context.WithValue(parent Context, key, val interface{}) (ctx Context, cancel CancelFunc)

此函数接收上下文并返回派生上下文，其中值 val 与 key 关联，并通过上下文树与上下文一起传递。 这意味着一旦获得带有值的上下文，从中派生的任何上下文都会获得此值。 不建议使用上下文值传递关键参数，而是函数应接收签名中的那些值，使其显式化。
```
ctx := context.WithValue(context.Background(), key, "test")
```

### context.WithCancel(parent Context) (ctx Context, cancel CancelFunc)

这是它开始变得有趣的地方。此函数创建从传入的父上下文派生的新上下文。父上下文可以是后台上下文或传递给函数的上下文。

返回派生上下文和取消函数。只有创建它的函数才能调用取消函数来取消此上下文。 如果您愿意，可以传递取消函数，但是，强烈建议不要这样做。 这可能导致取消函数的调用者没有意识到取消上下文的下游影响。 可能存在源自此的其他上下文，这可能导致程序以意外的方式运行。 简而言之，永远不要传递取消函数。
```
ctx, cancel := context.WithCancel(context.Background())
```

### context.WithDeadline(parent Context, d time.Time) (ctx Context, cancel CancelFunc)

此函数返回其父项的派生上下文，当截止日期超过或取消函数被调用时，该上下文将被取消。例如，您可以创建一个将在以后的某个时间自动取消的上下文，并在子函数中传递它。当因为截止日期耗尽而取消该上下文时，获此上下文的所有函数都会收到通知去停止运行并返回。
```
ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2 * time.Second))
```

### context.WithTimeout(parent Context, timeout time.Duration) (ctx Context, cancel CancelFunc)

此函数类似于context.WithDeadline。不同之处在于它将持续时间作为参数输入而不是时间对象。 此函数返回派生上下文，如果调用取消函数或超出超时持续时间，则会取消该派生上下文。

```
ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
```

## 函数接收和使用上下文

现在我们知道了如何创建上下文（Background 和 TODO）以及如何派生上下文（WithValue，WithCancel，Deadline 和 Timeout），让我们讨论如何使用它们。

在下面的示例中，您可以看到接受上下文的函数启动一个 goroutine 并等待 该 goroutine 返回或该上下文取消。select 语句帮助我们选择先发生的任何情况并返回。

`<-ctx.Done()`一旦Done通道被关闭，这个```<-ctx.Done():``` 被选择。一旦发生这种情况，此函数应该放弃运行并准备返回。这意味着您应该关闭所有打开的管道，释放资源并从函数返回。有些情况下，释放资源可以阻止返回，比如做一些挂起的清理等等。在处理上下文返回时，您应该注意任何这样的可能性。

本节后面的示例有一个完整的 go 语言程序，它说明了超时和取消功能。
```

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

到目前为止，我们已经看到使用上下文可以设置截止日期，超时或调用取消函数来通知所有使用任何派生上下文的函数来停止运行并返回。以下是它如何工作的示例：

***main function:***
* 用 cancel 创建一个上下文
* 随机超时后调用取消函数

***doWorkContext function:***
* 派生一个超时上下文
* 这个上下文将被取消当
* 主进程调用取消函数或
* 超时到或
* doWorkContext 调用它的取消函数
* 开启一个 goroutine 传入派生上下文，做些缓慢的流程
* 等待 goroutine 完成或上下文被主进程取消，以先发为准

***sleepRandomContext function***
* 开启一个 goroutine 去做些缓慢的流程
* 等待该 goroutine 完成或，
* 等待上下文被主进程取消，操时或它自己的取消函数被调用

***sleepRandom function***
* 随机时间休眠
* 此示例使用休眠来模拟随机处理时间，在实际示例中，您可以使用通道来通知此函数以开始清理并在通道上等待它以确认清理已完成。

Playground: https://play.golang.org/p/grQAUN3MBlg (看起来我使用的随机种子，在 playground 时间没有真正改变，您需要在你本机执行去看随机性)

Github: https://github.com/pagnihotry/golang_samples/blob/master/go_context_sample.go
```
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

### 缺陷

如果函数接收上下文参数，确保检查它是如何处理取消通知的。例如，exec.CommandContext 不会关闭读取管道，直到命令执行了进程创建的所有分支（Github问题：https://github.com/golang/go/issues/23019 ），这意味着如果等待 cmd.Wait() 直到外部命令的所有分支都已完成，则上下文取消不会使该函数立即返回。如果您使用超时或截止日期，您可能会发现这不能按预期运行。如果遇到任何此类问题，可以使用 time.After 实现超时。

### 最佳实践

1. context.Background 只应用在最高等级，作为所有派生上下文的根
2. context.TODO 应用在不确定要使用什么的地方，或者当前函数以后会更新以便使用上下文
3. 上下文取消是建议性的，这些函数可能需要一些时间来清理和退出
4. context.Value 应该很少使用，它不应该被用来传递可选参数。这使得 API 隐式的并且可以引起错误。取而代之的是，这些值应该作为参数传递。
5. 不要将上下文存储在结构中，在函数中显式传递它们，最好是作为第一个参数。
6. 永远不要传递不存在的上下文。相反，如果您不确定使用什么，使用一个 ToDo 上下文。
7. 上下文结构不具有取消方法，因为只有派生上下文的函数才应该取消上下文。

----------------

via: https://medium.com/@parikshit/understanding-the-context-package-in-golang-b1392c821d14

作者：[Parikshit Agnihotry](https://medium.com/@parikshit)
译者：[themoonbear](https://github.com/themoonbear)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

