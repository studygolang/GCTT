https://www.ardanlabs.com/blog/2019/04/concurrency-trap-2-incomplete-work.html

# 并发陷阱 2:未完成的工作

Jacob Walker 2019年4月18日

## 介绍
In my first post on Goroutine Leaks, I mentioned that concurrency is a useful tool but it comes with certain traps that don’t exist in synchronous programs. To continue with this theme, I will introduce a new trap called incomplete work. Incomplete work occurs when a program terminates before outstanding Goroutines (non-main goroutines) complete. Depending on the nature of the Goroutine that is being terminated forcefully, this may be a serious problem.
在我的第一篇文章 Goroutine 泄露中，我提到并发编程是一个很有用的工具，但是使用它也会带来某些非并发编程中不存在的陷阱。为了继续这个主题，我将介绍一个新的陷阱，这个陷阱叫做未完成的工作。当进程在非主协程的协程结束前终止时这种陷阱就会发生。根据Gorotine的本性，强制关闭它将造成一个严重的问题。

Incomplete Work
To see a simple example of incomplete work, examine this program.

## 未完成的工作

为了看到一个简单的未完成任务陷阱的例子，请检查这个程序

Listing 1
https://play.golang.org/p/VORJoAD2oAh

**例1**

https://play.golang.org/p/VORJoAD2oAh

```
5 func main() {
6     fmt.Println("Hello")
7     go fmt.Println("Goodbye")
8 }
```

The program in Listing 1 prints "Hello" on line 6 and then on line 7, the program calls fmt.Println again but does so within the scope of a different Goroutine. Immediately after scheduling this new Goroutine, the program reaches the end of the main function and terminates. If you run this program you won’t see the “Goodbye” message because in the Go specification there is a rule:

“Program execution begins by initializing the main package and then invoking the function main. When that function invocation returns, the program exits. It does not wait for other (non-main) goroutines to complete.”

在例一的程序中，第6行打印了"Hello",随后在第 7 行，这个程序再次调用了 `fmt.Println` ，但是这次是在一个不同的 Groutine 中调用的。当启动这个新的 Goroutine 后，这个程序就到了主函数的结尾，然后程序就终止了。如果你运行这个程序，你不会看到“Goodbye”这个信息，因为 Go 的规范中有一个这样的规则：

>
程序的启动是通过初始化 main 包，然后调用其中的 main 方法来实现的。当这个 main 函数返回时，这个程序就退出了。它不会等待其他非主协程完成后再退出。
>

The spec is clear that your program will not wait for any outstanding Goroutines to finish when the program returns from the  main function. This is a good thing! Consider how easy it is to let a Goroutine leak or have a Goroutine run for a very long time. If your program waited for non-main Goroutines to finish before it could be terminated, it could be stuck in some kind of zombie state and never terminate.

这个情况就很清楚了，当你的程序的主函数返回时，它不会等待任何非主协程完成，考虑到协程泄露和协程运行很长时间，这真是个好事情啊！当你的程序本可以结束，但是却要的等待一个非主协程完成，那么它可能就会卡住，以至于永远不会终止。

However, this termination behavior becomes a problem when you start a Goroutine to do something important, but the main function does not know to wait for it to complete. This type of scenario can lead to integrity issues such as corrupting databases, file systems, or losing data.

然后，当你启动一个协程去做重要的事情时，这种终止的方式就变成一个问题，因为主函数不会等待这个重要的协程完成就会返回。这种情况就会导致完整性问题，例如损坏数据库，文件系统，或者丢失数据。

## A Real Example

## 一个真实的例子

At Ardan Labs, my team built a web service for a client that required certain events to be tracked. The system for recording events had a method similar to the type Tracker shown below in Listing 2:

在 Ardan 实验室中，我的团队需要为客户搭建一个跟踪特定事件的 web 服务，这个记录事情的系统有一个类似例 2 中 Tracker 类型绑定的方法，

**例2**

https://play.golang.org/p/8LoUoCdrT7T

```
 9 // Tracker knows how to track events for the application.
10 type Tracker struct{}
11 
12 // Event records an event to a database or stream.
13 func (t *Tracker) Event(data string) {
14     time.Sleep(time.Millisecond) // Simulate network write latency.
15     log.Println(data)
16 }
```

The client was concerned that tracking these events would add unnecessary latency to response times and wanted to perform the tracking asynchronously. It is unwise to make assumptions about performance, so our first task was to measure latency of the service with events tracked in a straight-forward and synchronous approach. In this case, the latency was unacceptably high and the team decided an asynchronous approach was needed. If the synchronous approach was fast enough then this story would be over as we would have moved on to more important things.

With that in mind, the handlers that tracked events were initially written like this:

Listing 3
https://play.golang.org/p/8LoUoCdrT7T

```
18 // App holds application state.
19 type App struct {
20     track Tracker
21 }
22 
23 // Handle represents an example handler for the web service.
24 func (a *App) Handle(w http.ResponseWriter, r *http.Request) {
25 
26     // Do some actual work.
27 
28     // Respond to the client.
29     w.WriteHeader(http.StatusCreated)
30 
31     // Fire and Hope.
32     // BUG: We are not managing this goroutine.
33     go a.track.Event("this event")
34 }
```

The significant part of the code in listing 3 is line 33. This is where the a.track.Event method is being called within the scope of a new Goroutine. This had the desired effect of tracking events asynchronously without adding latency to the request. However, this code falls into the incomplete work trap and must be refactored. Any Goroutine created on line 33 has no guarantee of running or finishing. This is an integrity issue as events can be lost when the server shuts down.







---

via: https://www.ardanlabs.com/blog/2019/04/concurrency-trap-2-incomplete-work.html

作者：[Jacob Walker](https://github.com/jcbwlkr)
译者：[xmge](https://github.com/xmge)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
