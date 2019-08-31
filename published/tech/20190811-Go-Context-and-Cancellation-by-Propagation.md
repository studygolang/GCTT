# Go：Context 和传播取消

首发于：https://studygolang.com/articles/23240

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/context-and-cancellation-by-propagation/image_1.png)

[context 包](https://blog.golang.org/context)在 Go 1.7 中引入，它为我们提供了一种在应用程序中处理 context 的方法。这些 context 可以为取消任务或定义超时提供帮助。通过 context 传播请求的值也很有用，但对于本文，我们将重点关注 context 的取消功能。

## 默认的 contexts
Go 的 `context` 包基于 TODO 或者 Background 来构建 context。

```go
var (
   background = new(emptyCtx)
   todo       = new(emptyCtx)
)

func Background() Context {
   return background
}

func TODO() Context {
   return todo
}
```

我们可以看到，它们都是空的 context。这是简单的 context，永远不会被取消，也不会带任何值。

你可以将 background context 作为主 context，并将其派生出新的 context。基于这些，你不应直接在包中使用 context;它应该在你的主函数中使用。如果要使用 `net/http` 包构建服务，则主 context 将由请求提供：

```go
net/http/request.go
func (r *Request) Context() context.Context {
   if r.ctx != nil {
      return r.ctx
   }
   return context.Background()
}
```

如果你在自己的包中工作并且没有任何可用的 context，在这种情况下你应该使用 TODO context。通常，或者如果你对必须使用的 context 有任何疑问，可以使用 TODO context。现在我们知道了主 context，让我们看看它是如何派生子 context 的。

## Contexts 树
父 context 派生出的子 context 会在在其内部结构中创建一个和父 context 之间的联系：

```go
type cancelCtx struct {
   Context

   mu       sync.Mutex
   done     chan struct{}
   children map[canceler]struct{}
   err      error
}
```

`children` 字段跟踪以此 context 创建的所有子项，而 `Context` 指向创建当前项的 context。

以下是创建一些 context 和子 context 的示例：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/context-and-cancellation-by-propagation/image_2.png)

每个 context 都相互链接，如果我们取消 “C” context，所有它的孩子也将被取消。Go 会对它的子 context 进行循环逐个取消：

```go
context/context.go
func (c *cancelCtx) cancel(removeFromParent bool, err error) {
   [...]
   for child := range c.children {
      child.cancel(false, err)
   }
   [...]
}
```

取消结束，将不会通知父 context。如果我们取消 C1，它只会通知 C11 和 C12：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/context-and-cancellation-by-propagation/image_3.png)

这种取消传播允许我们定义更高级的例子，这些例子可以帮助我们根据主 context 处理多个/繁重的工作。

## 取消传播
让我们通过 goroutine A 和 B 来展示一个取消的例子，它们将并行运行，因为拥有共同的 context ，当一个发生错误取消时，另外一个也会被取消：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/context-and-cancellation-by-propagation/image_4.png)

如果没有任何错误发生，每个过程都将正常运行。我在每个任务上添加了一条跟踪，这样我们就可以看到一棵树：

```plain
A - 100ms
B - 200ms
    -> A1 - 100ms
        -> A11 - 50ms
    -> B1 - 100ms
        -> A12 - 300ms
    -> B2 - 100ms
        -> B21 - 150ms
```

每项任务都执行得很好。现在，让我们尝试让 A11 模拟出错误：

```plain
A - 100ms
    -> A1 - 100ms
B - 200ms
        -> A11 - error
        -> A12 - cancelled
    -> B1 - 100ms
    -> B2 - cancelled
    -> B21 - cancelled
```

我们可以看到，当 B2 和 B21 被取消的同时，A12 被中断，以避免做出不必要的处理（译者注：B2 B21 的取消不是因为 A12 中断，应该是想表达并发安全的意思）：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/context-and-cancellation-by-propagation/image_5.png)

我们可以在这里看到 context 对于**多个 goroutine 是线程安全的**。实际上，有可能是因为我们之前在结构中看到的 mutex，它保证了对 context 的并发安全。

## context 泄漏
正如我们在内部结构中看到的那样，当前 context 在 `Context` 属性中保持其父级的链接，而父级将当前 context 保留在 `children` 属性中。对 cancel 函数的调用将把当前 context 中的子项清除并删除与父项的链接：

```go
func (c *cancelCtx) cancel(removeFromParent bool, err error) {
   [...]
   c.children = nil

   if removeFromParent {
      removeChild(c.Context, c)
   }
}
```

如果未调用 cancel 函数，则主 context 将始终保持与它创建的 context 的链接，从而导致可能的内存泄漏。

可以用 `go vet` 命令来检查是否泄漏，它将对可能的泄漏抛出警告：

```plain
the cancel function returned by context.WithCancel should be called, not discarded, to avoid a context leak
```

## 总结
`context` 包还有另外两个利用 cancel 函数的函数：`WithTimeout` 和 `WithDeadline`。在定义的超时/截止时间后，它们都会自动触发 cancel 函数。

`context` 包还提供了一个 `WithValue` 的方法，它允许我们在 context 中存储任何对键/值。此功能受到争议，因为它不提供明确的类型控制，可能导致糟糕的编程习惯。如果你想了解 `WithValue` 的更多信息，我建议你阅读[Jack Lindamood 关于 context 值的文章](https://medium.com/@cep21/how-to-correctly-use-context-context-in-go-1-7-8f2c0fafdf39)。

---

via: https://medium.com/@blanchon.vincent/go-context-and-cancellation-by-propagation-7a808bbc889c

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[zhoudingding](https://github.com/dingdingzhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
