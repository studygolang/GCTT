首发于：https://studygolang.com/articles/21385

# 并发陷阱 2: 未完成的工作

Jacob Walker 2019 年 4 月 18 日

## 介绍

在我的第一篇文章 [Goroutine 泄露](https://studygolang.com/articles/17364) 中，我提到并发编程是一个很有用的工具，但是使用它也会带来某些非并发编程中不存在的陷阱。为了继续这个主题，我将介绍一个新的陷阱，这个陷阱叫做未完成的工作。当进程在非主协程的协程结束前终止时，这种陷阱就会发生。根据 Gorotine 的特性，强制关闭它将造成一个严重的问题。

## 未完成的工作

为了看到一个简单的未完成任务陷阱的例子，请检查这个程序

**例 1**

[https://play.golang.org/p/VORJoAD2oAh](https://play.golang.org/p/VORJoAD2oAh)

```go
5 func main() {
6     fmt.Println("Hello")
7     go fmt.Println("Goodbye")
8 }
```

在例一的程序中，第 6 行打印了 "Hello", 随后在第 7 行，这个程序再次调用了 `fmt.Println` ，但是这次是在一个不同的 Groutine 中调用的。当启动这个新的 Goroutine 后，这个程序就到了主函数的结尾，然后程序就终止了。如果你运行这个程序，你不会看到“ Goodbye ”这个信息，因为 [Go 的规范](https://golang.org/ref/spec#Program_execution) 中有一个这样的规则：

> 程序的启动是通过初始化 main 包，然后调用其中的 main 方法来实现的。当这个 main 函数返回时，这个程序就退出了。它不会等待其他非主协程完成后再退出。

这个情况就很清楚了，当你的程序的主函数返回时，它不会等待任何非主协程完成，考虑到协程泄露和协程运行很长时间，这真是个好事情啊！当你的程序本可以结束，但是却要的等待一个非主协程完成，那么它可能就会卡住，以至于永远不会终止。

然后，当你启动一个协程去做重要的事情时，这种终止的方式就变成一个问题，因为主函数不会等待这个重要的协程完成就会返回。这种情况就会导致完整性问题，例如损坏数据库，文件系统，或者丢失数据。

## 一个真实的例子

在 Ardan 实验室中，我的团队需要为客户搭建一个跟踪特定事件的 Web 服务，这个记录事情的系统有一个类似例 2 中 Tracker 类型绑定的方法，

**例 2**

[https://play.golang.org/p/8LoUoCdrT7T](https://play.golang.org/p/8LoUoCdrT7T)

```go
 9 // Tracker knows how to track events for the application.
10 type Tracker struct{}
11
12 // Event records an event to a database or stream.
13 func (t *Tracker) Event(data string) {
14     time.Sleep(time.Millisecond) // Simulate network write latency.
15     log.Println(data)
16 }
```

客户担心跟踪这些事件会增加程序的响应时间，希望可以通过异步执行来进行跟踪。猜想程序的运行情况是不明智的，于是我们首先的任务是通过同步追踪的方式记录发生的事件，从而衡量服务延迟。在这个案例中，程序的延迟真的是高的不能接受，于是我们的团队决定采用异步的方法来实现。如果同步的方式足够快，我们也就不会将这个故事了，我们也会去做更重要的事。

考虑到这一点，跟踪记录事件的处理程序最初编写如下：

**例 3**

[https://play.golang.org/p/8LoUoCdrT7T](https://play.golang.org/p/8LoUoCdrT7T)

```go
18 // App holds application state.
19 type App struct {
20     track Tracker
21 }
22
23 // Handle represents an example handler for the Web service.
24 func (a *App) Handle(w http.ResponseWriter, r *http.Request) {
25
26     // Do some actual work.
27
28     // Respond to the client.
29     w.WriteHeader(http.StatusCreated)
30
31     // Fire and Hope.
32     // BUG: We are not managing this Goroutine.
33     go a.track.Event("this event")
34 }
```

在代码中最重要的部分是 33 行，在这里，`a.track.Event` 方法在一个新的协程中被调用的。这样就预期地消除了请求的延迟。然而，这些代码却陷入了 *未完成的工作* 的陷阱，我们必须重构它。任何在第 33 行常见的协程，我们都无法保证它运行或者完成。这是一个数据完整性的严重问题，因为当服务被终止时，要记录的事件信息将会丢失。

## 为保证重构

为了避免陷入这个陷阱，团队修改了代码，让 `Tracker` 去管理这个协程。我们使用 `sync.WaitGroup` 去确保当主函数返回时，所有的协程都已经完成。
为了避免这个陷阱，团队修改了代码，让 `Tracker` 来管理 Goroutines。`Tracker` 使用 sync.waitgroup 来记录打开的 Goroutine 数量，并为主函数提供一个关闭方法，并且这个方法会等到所有 Goroutine 完成后才会返回。

刚开始我们直接使用不创建协程的方法。只要在例 4 的 53 行去掉 `go` 就可以了。

**例 4**

https://play.golang.org/p/BMah6_C57-l

```go
44 // Handle represents an example handler for the Web service.
45 func (a *App) Handle(w http.ResponseWriter, r *http.Request) {
46
47     // Do some actual work.
48
49     // Respond to the client.
50     w.WriteHeader(http.StatusCreated)
51
52     // Track the event.
53     a.track.Event("this event")
54 }
```

下一步 `Tracker` 类型将可以自己管理协程

**例 5**

https://play.golang.org/p/BMah6_C57-l

```go
10 // Tracker knows how to track events for the application.
11 type Tracker struct {
12     wg sync.WaitGroup
13 }
14
15 // Event starts tracking an event. It runs asynchronously to
16 // not block the caller. Be sure to call the Shutdown function
17 // before the program exits so all tracked events finish.
18 func (t *Tracker) Event(data string) {
19
20     // Increment counter so Shutdown knows to wait for this event.
21     t.wg.Add(1)
22
23     // Track event in a Goroutine so caller is not blocked.
24     go func() {
25
26         // Decrement counter to tell Shutdown this Goroutine finished.
27         defer t.wg.Done()
28
29         time.Sleep(time.Millisecond) // Simulate network write latency.
30         log.Println(data)
31     }()
32 }
33
34 // Shutdown waits for all tracked events to finish processing.
35 func (t *Tracker) Shutdown() {
36     t.wg.Wait()
37 }
```

在例 5 的第 12 行为 `Tracker` 增加了字段 wg，wg 的类型为 `sync.WaitGroup`。并且在 `Event` 函数中，也就是代码的第 21 行，程序调用了 t.wg 的 `t.wg.Add(1)` 方法。调用这个方法可以记录在 24 行创建的协程数量。这样跟踪事物的方法就可以满足用户对延迟的需求了。被创建的协程在结束时会调用 `t.wg.Done()`, 这样记录协程个数的计数器就会减 1，`WaitGroup` 就知道这程结束了。

调用 `Add` 和 `Done` 对于记录活跃协程的数量是很有用的，但是主程序必须等待这些协程完成。为了满足这点，在 35 行 `Tracker` 又增加了一个新的方法 `Shutdown` 这个方法很简单，其中只是调用了 `t.Wg.Wait()`，这个函数会一直阻塞，直到协程的计数器减到 0，最后，这个程序必须要在 `func main` 中被调用。就像在 例 6 中。

**例 6**

https://play.golang.org/p/BMah6_C57-l

```go
56 func main() {
57
58     // Start a server.
59     // Details not shown...
60     var a App
61
62     // Shut the server down.
63     // Details not shown...
64
65     // Wait for all event Goroutines to finish.
66     a.track.Shutdown()
67 }
```
在例 6 中最关键的地方是第 66 行，这个函数会一直在阻塞，防止 `func main` 终止，直到 `a.track.Shutdown()` 完成。

## 也许不用等太久

所展示的 `Shutdown` 方法的实现是很简单的，也确实完成了它的工作，即等待所有的协程完成。但是不幸是的是，这里无法限制等待多长时间。考虑到在生产环境上，您可能不愿意无限期地等待程序关闭。为了给 `Shutdown` 方法增加一个最后期限，团队将程序改成了如下所示：

**例 7**

[https://play.golang.org/p/p4gsDkpw1Gh](https://play.golang.org/p/p4gsDkpw1Gh)

```go
36 // Shutdown waits for all tracked events to finish processing
37 // or for the provided context to be canceled.
38 func (t *Tracker) Shutdown(ctx context.Context) error {
39
40     // Create a channel to signal when the waitgroup is finished.
41     ch := make(chan struct{})
42
43     // Create a Goroutine to wait for all other Goroutines to
44     // be done then close the channel to unblock the select.
45     go func() {
46         t.wg.Wait()
47         close(ch)
48     }()
49
50     // Block this function from returning. Wait for either the
51     // waitgroup to finish or the context to expire.
52     select {
53     case <-ch:
54         return nil
55     case <-ctx.Done():
56         return errors.New("timeout")
57     }
58 }
```

现在在例 7 的 38 行，`Shutdown` 方法将 `context.Context` 作为输入参数。这就是调用者来限制等待程序终止的时间。在方法的 41 行，一个 channel 被创建了，并且在 45 行，一个 Goroutine 也启动了。这个 Goroutine 的唯一作用就是等待所有的协程都完成，然后关闭这个 channel。最后，在 52 行有一个 `select` 代码块，它可以等待 context 被取消，或则通道被关闭。

下一步团队就 `func main` 改成了如下所示：

**例 8**

[https://play.golang.org/p/p4gsDkpw1Gh](https://play.golang.org/p/p4gsDkpw1Gh)

```go
86     // Wait up to 5 seconds for all event Goroutines to finish.
87     const timeout = 5 * time.Second
88     ctx, cancel := context.WithTimeout(context.Background(), timeout)
89     defer cancel()
90
91     err := a.track.Shutdown(ctx)
```

在例 8 中，`mian` 创建了一个 5 s 的超时取消的 context。这将传递到 a.track.shutdown 以设置 main 愿意等待的时间。

## 结论

随着 Goroutines 的引入，此服务器的处理程序能够将跟踪事件所需的 API 请求时间延迟成本降到最低。只使用 `go` 关键字在后台运行此工作是很容易的，但该解决方案存在完整性问题。正确地执行此操作需要在关闭程序之前努力确保所有相关的 `goroutine` 都已终止。

**并发性是一个有用的工具，但我们必须谨慎使用它。**

---

via: https://www.ardanlabs.com/blog/2019/04/concurrency-trap-2-incomplete-work.html

作者：[Jacob Walker](https://github.com/jcbwlkr)
译者：[xmge](https://github.com/xmge)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
