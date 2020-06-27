# Go 中可取消的读取操作

在使用 Go 进行工作时，使用 `io.Reader` 和 `io.Writer` 接口是最常见的情场景之一。这很合理，它们是数据传输的主力接口。

关于 `io.Reader` 接口，有一点令我困惑：它的 `Read()` 方法是阻塞的，一旦读取操作开始，你没有办法去抢占它。同样，也无法在读取操作上执行 `select` 操作，异步协调多个 `io.Reader` 时的读取操作会有点棘手。

`io.ReadCloser` 是 Go 提供的一个常用的退出通道，在许多情况下，它确实允许你抢占一个读取操作。在某些实现中，调用 reader 的 `Close()` 方法将会取消剩余的读取操作。但是，你只能执行一次，当一个 reader 被关闭后，后续的读取操作将会失败。

我正在维护的应用中有很多地方可以有效地执行以下操作：

```go
func asyncReader(r io.Reader) {
    ch := make(chan []byte)
    go func() {
        defer close(ch)
        for {
            var b []byte
            // Point of no return
            if _, err := r.Read(b); err != nil {
                return
            }
            ch <- b
        }
    }()

    select {
    case data <- ch:
        doSomething(data)
    //...
    // Other asynchronous cases
    //...
    }
}
```

当必须将 `Read()` 方法和一些其他异步数据源，如 gRPC 流一起混用时，我们使用上述模式。你生成一个 goroutine 从 reader 中不停地读取数据，将接收到的数据从一个 channel 中发送出去，当 reader 关闭时执行一些清理操作。

当你尝试清理整个系统时，会遇到棘手的部分：如果我们想要发送结束该过程的信号，该怎么办？考虑一个改造过的例子：

```go
func asyncReader(r io.Reader, doneCh <-chan struct{}) {
    ch := make(chan []byte)
    continueReading := false
    go func() {
        defer close(ch)
        for !continueReading {
            var b []byte
            // Point of no return
            if _, err := r.Read(b); err != nil {
                return
            }
            ch <- b
        }
    }()

    select {
    case data <- ch:
        doSomething(data)
    case <-doneCh:
        continueReading = false
        return
    }
}
```

当上面的 `doneCh` 被关闭时，`asyncReader` 方法将返回。我们创建的 goroutine 也将在下次计算 `for` 循环中的条件时返回。但是，如果 goroutine 阻塞在 `r.Read()` 该怎么办？那样的话，我们实质上泄露了一个 goroutine。在 reader 离开阻塞状态之前，我们将一直陷入困境。

reader 有可能永远阻塞下去吗？也许会，也许不会。这取决于底层的 reader。至关重要的是，接口无法向你保证它一定会解除阻塞，因此仍然存在着 goroutine 泄露的可能性。

如果只有 `io.Reader` 接口，那么你现在就陷入了困境。如果你可以切换到 `io.ReadCloser` 接口，并且控制 Close() 方法的行为，你可以将其用作取消最后读取操作的辅助通道。但是，也只能这样。

这令人沮丧，这让我好奇，为什么下述接口这样的方法并不常见：

```go
interface PreemptibleReader {
    Read(ctx context.Context, p []byte) (n int, err error)
}
```

假如 `Read()` 方法携带一个 context 参数，就可以抢占读取操作。

## 为什么 `io.Reader` 会阻塞

退一步讲，为什么读取操作默认不支持抢占，值得思考一下。考虑 `io.Reader` 的接口：

```go
interface Reader {
    Read(p []byte) (n int, err error)
}
```

这究竟是什么意思呢？<sup>[1](#fn1)</sup>实际上，接口规定：“拿着字节切片 `p`，向里面写入一些内容。告诉我你写了多少字节（`n`），以及在执行过程中是否某处遇到了错误（`err`）”。

这个接口令人惊讶的通用，因为它需要适应多种用途。从内存缓冲到 HTTP 响应到数据库事务结果的所有内容都可以实现为一个 `io.Reader` 接口。因此，许多 `io.Reader` 的实现有内部状态。

同样，对 `p` 的修改也不是原子操作。如果你要在执行过程中的任意地方取消读取，你可能得到不一致的状态 - 同时在目标字节切片和 reader 的内部状态中，如果适用的话。

如果我们允许可被任意抢占的读取，这些会导致大量的复杂性。如果抢占发生在修改字节切片的中途会导致什么？reader 是否需要将字节切片清零？如果发生在 reader 即将返回错误并进行清理工作时呢？

抢占使事情变得混乱。当然，有一些方法可以进行“优雅的清理”。当我们想要取消 Go 中的操作时，通常使用 context，例如网络请求。取消一个 context 并不保证操作会立刻被终止，因此有一些空间可以进行记录和清理。

我认为要记住的是，`io.Reader` 被设计为几乎可以在任意地方工作。因此，尤其当 `io.Reader` 实际完成的工作更类似 `memcpy`，而非一个 HTTP 请求时，管理 contexts 和内部记录带来的额外开销会导致令人厌烦的性能问题。

## 那 `Close()` 方法呢

正如上面提到的，`Close()` 可以作为一种机制来取消挂起的 `Read()`。然而，仍然有一些场景适合不关闭底层 reader 直接抢占读取。

比如说，考虑从 `os.Stdin` 读取的程序。如果你想使用 `Close()` 来抢占一个执行中的从 `stdin` 读取操作，那么你将需要关闭 `stdin`。这并不是很好，因为一旦关闭 `stdin`，你就无法再重新打开并再次读取。

由于 `Close()` 通常时一次性操作，在通常情况下，它并不是抢占读取的理想选择。`Close()` 最好被归类为清理方法，而非抢占信号。

## 你如何解决这个问题

考虑到上述注意事项，仍然有办法有效地取消进行中的读取。那并没有绕开 `io.Reader` 接口的限制，但是你可以为接口的使用者铺平道路。

```go
type CancelableReader struct {
    ctx  context.Context
    data chan []byte
    err  error
    r    io.Reader
}

func (c *CancelableReader) begin() {
    buf := make([]byte, 1024)
    for {
        n, err := c.r.Read(buf)
        if err != nil {
            c.err = err
            close(c.data)
            return
        }
        tmp := make([]byte, n)
        copy(tmp, buf[:n])
        c.data <- tmp
    }
}

func (c *CancelableReader) Read(p []byte) (int, error) {
    select {
    case <-c.ctx.Done():
        return 0, c.ctx.Err()
    case d, ok := <-c.data:
        if !ok {
            return 0, c.err
        }
        copy(p, d)
        return len(d), nil
    }
}

func New(ctx context.Context, r io.Reader) *CancelableReader {
    c := &CancelableReader{
        r:    r,
        ctx:  ctx,
        data: make(chan []byte),
    }
    go c.begin()
    return c
}
```

上述是 `io.Reader` 接口的包装，它的构造器中包含了 `context.Context`。当 context 被取消，任何进行中的读取操作都会立即返回。稍微调整一下，上述方法也可以很好的适用于 `io.ReadCloser`。

`CancelableReader` 包装器上有一个*巨大的星号*：它仍然存在 goroutine 泄漏。如果底层的 `io.Reader` 永远不返回，那么 `begin()` 中的 goroutine 将永远不会被清理。

至少，使用这种方法可以更清楚的知道泄漏发生的位置，你可以在 struct 上存储一些额外的状态来追踪 goroutine 是否结束。或许，你可以将这些 `CancelableReader` 组成一个池，并在读取全部完成时回收它们。

---

这并不像是个令人满意的结论。我很好奇是否有更好的方法来抢占阅读。或许答案是“增加足够多层的包装器，直到问题消失”？或许答案是“了解底层 reader 的实现”？- 例如，在 `stdin` 上使用 [syscall.SetNonblock](https://golang.org/pkg/syscall/#SetNonblock)。

如果有人有更好的办法，我很想听听！😃

---

<span id='fn1'>
1. 我喜欢 [Go 的简单性](https://benjamincongdon.me/blog/2019/11/11/The-Value-in-Gos-Simplicity/)的一点是，它经常使你陷入这些表面的存在性问题，比如“从阻塞接口读取意味着什么？”😜
</span>

---
via: [https://benjamincongdon.me/blog/2020/04/23/Cancelable-Reads-in-Go/](https://benjamincongdon.me/blog/2020/04/23/Cancelable-Reads-in-Go/)

作者：[Ben Congdon](https://benjamincongdon.me/)
译者：[DoubleLuck](https://github.com/doubleluck)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
