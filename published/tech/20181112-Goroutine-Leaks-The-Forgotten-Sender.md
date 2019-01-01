首发于：https://studygolang.com/articles/17364

# Goroutine 泄露 - 被遗忘的发送者

## 引言

并发编程允许开发人员使用多个执行路径解决问题，并且通常用于提高性能。并发并不意味着这些多路径是并行执行的；它意味着这些路径是无序执行的而不是顺序执行。从历史上看，使用由标准库或第三方开发人员提供的库可以促进这种类型的编程。

在 Go 中，语言本身和程序运行时内置了 Goroutines 和 channel 等并发特性，以减少或消除对库的需求。这很容易在 Go 中编写并发程序时造成错觉。在你决定使用并发时必须要谨慎，因为如果没有正确使用它那么就会带来一些稀罕的副作用或陷阱。如果你不小心，这些陷阱会产生复杂的问题和令人讨厌的 bug。

我在这篇文章中讨论的陷阱会与 Goroutine 泄漏有关。

## Goroutines 泄露

当涉及到内存管理时，Go 已经为您处理了许多细节。Go 在编译时使用[**逃逸分析**](https://studygolang.com/articles/12444) 来决定值在内存中的位置。程序运行时通过使用[**垃圾回收器**](https://blog.golang.org/ismmkeynote) 跟踪和管理堆分配。虽然在应用程序中创建[**内存泄漏**](https://en.wikipedia.org/wiki/Memory_leak) 不是不可能的，但是这种可能性已经大大降低了。

一种常见的内存泄漏类型就是 Goroutines 泄漏。如果你开始了一个你认为最终会终止但是它永远不会终止的 Goroutine，那么它就会泄露了。它的生命周期为程序的生命周期，任何分配给 Goroutine 的内存都不能释放。所以在这里建议[**“永远不要在不知道如何停止的情况下，就去开启一个 Goroutine ”**](https://dave.cheney.net/2016/12/22/never-start-a-goroutine-without-knowing-how-it-will-stop)。

要弄明白基本的 Goroutine 泄漏，请查看以下代码：

**清单 1**

https://play.golang.org/p/dsu3PARM24K

```go
// leak 是一个有 bug 程序。它启动了一个 goroutine
// 阻塞接收 channel。一切都将不复存在
// 向那个 channel 发送数据，并且那个 channel 永远不会关闭
// 那个 goroutine 会被永远锁死
func leak() {
     ch := make(chan int)

     go func() {
        val := <-ch
        fmt.Println("We received a value:", val)
    }()
}
```

清单 1 中定义了一个名为 `leak` 的函数。该函数在第 6 行创建一个 channel，该 channel 允许 Goroutines 传递整型数据。然后在第 8 行创建 Goroutine，它在第 9 行被阻塞，等待从 channel 中接收数据。当 Goroutine 正在等待时，`leak` 函数会结束返回。此时，程序的其他任何部分都不能通过 channel 发送数据。这使得 Goroutine 在第 9 行被无限期的等待。第 10 行的 `fmt.Println` 调用永远不会发生。

在本例中，Goroutine 泄漏可以在代码检查期间快速识别。不幸的是，生产代码中的 Goroutine 泄漏通常更难找到。我无法展示 Goroutine 泄漏可能发生的所有方式，但是这篇文章将详细说明你可能遇到的某种 Goroutine 泄漏。

## 泄露：被遗忘的发送者

> 对于这个泄漏示例，你将看到一个无限期阻塞的 Goroutine，等待在通道上发送一个值。

我们要看的程序会根据一些搜索词找到一个记录，然后打印出来。这个程序是围绕一个叫做 `search` 的函数构建的 :

**清单 2**

https://play.golang.org/p/o6_eMjxMVFv

```go
// search 模拟成一个查找记录的函数
// 在查找记录时。执行此工作需要 200 ms。
func search(term string) (string, error) {
     time.Sleep(200 * time.Millisecond)
     return "some value", nil
}
```

清单 2 中第 3 行的 `search` 函数是一个模拟实现，用于模拟长时间运行的操作，如数据库查询或 Web 调用。在这个例子中，硬编码需要 200 ms。

在清单 3 中程序调用 `search` 函数，如下：

**清单 3**

https://play.golang.org/p/o6_eMjxMVFv

```go
// process 函数是在该程序中搜索一条记录
// 然后打印它
func process(term string) error {
    record, err := search(term)
    if err != nil {
        return err
    }

   fmt.Println("Received:", record)
   return nil
}
```

在清单 3 中的第 3 行，定义了一个名为 `process` 的函数，它接受一个表示搜索项的字符串参数。在第 4 行，`term` 变量传递给 `serach` 函数，该函数返回查找到的记录和错误。如果发生错误，则将错误返回到第 6 行的调用方。如果没有错误，则在第 9 行打印该记录。

对于某些应用程序来说，顺序调用 `search` 函数时产生的延迟可能是无法接受的。假设不能使 `search` 函数运行得更快，则可以将 `process` 函数更改为不消耗 `search` 所产生的总延迟成本。

为此，我们可以像下面清单 4 中那种使用 Goroutine，不幸的是，这第一次尝试是错误的，因为它造成了潜在的 Goroutine 泄漏。

**清单 4**

https://play.golang.org/p/m0DHuchgX0A

```go
// serach 函数得到的返回值用 result 结构体来保存
// 通过单个 channel 来传递这两个值
type result struct {
    record string
    err    error
}

// process 函数是一个用来寻找记录的函数
// 然后打印，如果超过 100 ms 就会失败 .
func process(term string) error {

     // 创建一个在 100 ms 内取消上下文的 context
     ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
     defer cancel()

     // 为 Goroutine 创建一个传递结果的 channel
     ch := make(chan result)

     // 启动一个 Goroutine 来寻找记录，然后得到结果
     // 将返回值从 channel 中返回
     Go func() {
         record, err := search(term)
         ch <- result{record, err}
     }()

     // 阻塞等待从 Goroutine 接收值
     // 通过 channel 和 context 来取消上下文操作
     select {
     case <-ctx.Done():
         return errors.New("search canceled")
     case result := <-ch:
         if result.err != nil {
            return result.err
         }
         fmt.Println("Received:", result.record)
         return nil
    }
 }
```

在清单 4 中的第 13 行，重写 `process` 函数以创建 `Context` 来在 100 ms 内取消上下文。有关如何使用 `Context` 的更多信息，请阅读 [go 语言开发文档](https://blog.golang.org/context) 。

然后在第 17 行，程序创建一个无缓冲的 channel，允许 Goroutines 传递 `result` 类型的数据。在第 21 到 24 行，定义了匿名函数，此处称为 Goroutine。此 Goroutine 调用 `search` 函数并尝试通过第 23 行的 channel 发送其返回值。

当 Goroutine 正在执行其工作时，`process` 函数执行第 28 行上的 `select` 模块。该模块有两种情况，它们都是 channel 接收操作。

在第 29 行，有一个从 `ctx.Done()` channel 接收的 `case`。如果上下文被取消（100 ms 持续时间到达），将执行此 `case`。如果执行此 `case`，则 `process` 函数将返回错误，代表着取消了等待第 30 行的 `search`。

或者，第 31 行上的 `case` 从 `ch` channel 接收并将值分配给名为 `result` 的变量。与前面在顺序实现中一样，程序在第 32 行和第 33 行检查和处理错误。如果没有错误，程序将在第 35 行打印记录，并返回 nil 以指示成功。

此重构设置了 `process` 函数等待 `search` 完成的最大持续时间。然而，这种实现也会造成潜在的 Goroutine 泄漏。想想代码中的 Goroutine 在做什么；在第 23 行，它通过 channel 发送。在此 channel 上发送将阻塞执行，直到另一个 Goroutine 准备接收值为止。在超时的情况下，接收方停止等待 Goroutine 的接收并继续工作。这将导致 Goroutine 永远阻塞，等待一个永远不会发生的接收器出现。这就是 Goroutine 泄露的时候。

修复：创造一些空间

解决此泄漏的最简单方法是将无缓冲 channel 更改为容量为 1 的缓冲通道。

**清单 5**

https://play.golang.org/p/u3xtQ48G3qK

```go
// 为 Goroutine 创建一个传递结果的 channel。
// 给它容量，以至于发送接受不会阻塞。
   ch := make(chan result, 1)
```

现在在超时情况下，在接收器继续运行之后，搜索 Goroutine 将通过将结果值放入 channel 来完成其发送，然后它将返回。Goroutine 的内存以及 channel 的内存最终将会被收回。一切都会自然而然地发挥作用。

在 [channel 的行为](https://www.ardanlabs.com/blog/2017/10/the-behavior-of-channels.html) 中，William Kennedy 提供了几个关于 channel 行为的很好的例子，并提供了有关其使用的哲学。该文章“[清单 10](https://www.ardanlabs.com/blog/2017/10/the-behavior-of-channels.html#signal-without-data-context) ”的最后一个示例显示了一个类似于此超时示例的程序。阅读该文章，获取有关何时使用缓冲 channel 以及适当的容量级别的更多建议。

## 结论

Go 让启动 Goroutines 变得简单，但我们有责任明智地使用它们。在这篇文章中，我展示了如何错误地使用 Goroutines 的一个例子。有许多方法可以创建 Goroutine 泄漏以及使用并发时可能遇到的其他陷阱。在以后的文章中，我将提供更多 Goroutine 泄漏和其他并发陷阱的例子。现在我会给你这个建议 ; 任何时候你开始 Goroutine 你必须问自己：

- 什么时候会终止？
- 什么可以阻止它终止？

***并发是一个有用的工具，但必须谨慎使用。***

---

via: https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html

作者：[Jacob Walker](https://github.com/jcbwlkr)
译者：[wumansgy](https://github.com/wumansgy)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
