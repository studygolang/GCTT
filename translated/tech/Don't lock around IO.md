# 不要对 I/O 上锁

锁可用于同步操作。但如果使用不当的话，也会引发显著的性能问题。一个比较常见出问题的地方是 HTTP handlers 处。尤其很容易在不经意间就会锁住网络 I/O。要理解这种问题，我们最好还是来看一个例子。这篇文章中，我会使用 Go。

为此，我们需要编写一个简单的 HTTP 服务器用以报告它接收到的请求数量。所有的代码可以从 [这里](https://github.com/gobuildit/gobuildit/tree/master/lock) 获得。

报告请求数量的服务看起来是这样的：

```go
package main

// import statements
// ...

const (
	payloadBytes = 1024 * 1024
)

var (
	mu    sync.Mutex
	count int
)

// register handler and start server in main
// ...

// BAD: Don't do this.
func root(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	count++

	msg := []byte(strings.Repeat(fmt.Sprintf("%d", count), payloadBytes))
	w.Write(msg)
}
```

`root` handler 在最顶部用了常规的上锁和 `defer` 解锁。接着，在持有锁期间，增长了 `count` 的值，并将 `count` 的值通过重复 `payloadBytes` 次生成的数据写入 `http.ResponseWriter` 之中。

对于经验不足的人，这个 handler 看起来貌似完美无缺。实际上，它会引发一个显著的性能问题。在网络 I/O 期间上锁，导致了这个 handler 执行起来的速度取决于最慢的那个客户端。

为了能够直接地看清楚问题，我们需要模拟一个缓慢的读取客户端（以下简称为慢客户端）。实际上，因为有些客户端实在是太慢了，所以对于暴露在开放网络中的 Go HTTP 客户端来说设置一个超时时间很有必要。因为内核拥有缓存写入和从 TCP sockets 读取的机制，所以我们的模拟需要一些技巧。假设我们创建的客户端发送了一个 `GET` 请求，却没有从 socket 读取到任何数据（代码在 [此处](https://github.com/gobuildit/gobuildit/blob/master/lock/client/main.go)）。这会使服务在 `w.Write` 处阻塞吗？

因为内核缓存了读写数据，所以至少在缓存填充满之前，我们不会看到服务速度有任何下滑。为了观察到这种速度下滑，我们要保证每次的写入数据都能填充满缓存。有两个办法。1) 调校一下内核。2) 每次都写入大批量的字节。

调校内核本身就是件迷人的事情。可以通过 [proc 目录](https://twitter.com/b0rk/status/981159808832286720)，有所有网络相关参数的 [文档](https://www.kernel.org/doc/Documentation/sysctl/net.txt)，也有 [各类](https://www.cyberciti.biz/faq/linux-tcp-tuning/) [主机调校](http://fasterdata.es.net/host-tuning/) 的 [教程](https://www.tecmint.com/change-modify-linux-kernel-runtime-parameters/)。但是对于我们而言，只需要往 socket 中写入大批量的数据，就可以填满普通的 Darwin (v17.4) 内核的 TCP 缓存了。注意，运行这个示例，你可能需要调整写入数据的量以保证填充满你的缓存。

现在我们启动服务，使用慢客户端来观察其他的客户端等待慢客户端的速度。慢客户端的代码在 [这里](https://github.com/gobuildit/gobuildit/blob/master/lock/client/main.go)。

首先，确认一个请求可以被快速地处理：

```
curl localhost:8080/

# Output:
# numerous 1's without any meaningful delay
```

现在，我们先运行慢客户端：

```
# Assuming $GOPATH/github.com/gobuildit/gobuildit/lock directory
go run client/main.go

# Output:
dialing
sending GET request
blocking and never reading
```

当慢客户端连接上服务器之后，再尝试运行“快”客户端：

```
curl localhost:8080/

# Hangs
```

我们可以直接地看到我们的锁策略如何不经意间阻塞了快客户端。如果回到我们的 handler 想一下我们是怎么使用锁的，就会明白其中的问题。

```go
func root(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// ...
}
```

通过在方法顶部的加锁和使用 `defer` 解锁，我们在整个 handler 期间都持有锁对象。这个过程包含了共享状态的操作，共享状态的读取和网络数据写入。也就是这些操作导致了问题。网络 I/O 是 [天生不可预知](https://en.wikipedia.org/wiki/Fallacies_of_distributed_computing) 的。诚然，我们可以通过配置超时来保护我们的服务避免过长时间的调用，但我们无法保证所有的网络 I/O 都能在固定的时间内完成。

解决问题的关键在于不要在 I/O 周围加锁。这个例子中，在 I/O 周围加锁没有任何意义。在 I/O 周围加锁会使我们的程序被不良网络情况和慢客户端影响。实际上，我们也部分放弃了对于我们程序同步化的控制。

让我们重写 handler 来只在关键部分加锁。

```go
// GOOD: Keep the critical section as small as possible and don't lock around
// I/O.
func root(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	count++
	current := count
	mu.Unlock()

	msg := []byte(strings.Repeat(fmt.Sprintf("%d", current), payloadBytes))
	w.Write(msg)
}
```

为了看出区别，尝试使用一个慢客户端和一个普通的客户端。

同样，先启动慢客户端：

```
# Assuming $GOPATH/github.com/gobuildit/gobuildit/lock directory
go run client/main.go
```

现在，使用 `curl` 来发送一个请求：

```
curl localhost:8080/
```

观察 `curl` 是否立即返回并带回了期望的 count。

诚然，这个例子过于不自然，也比典型的生产环境代码要简单得多。而且对于同步计数而言，使用 [atomics](https://golang.org/pkg/sync/atomic/) 包可能更加明智。虽然如此，我也希望这个例子阐述了对于慎重加锁的重要性。虽然也会有例外，但通常大部分情况下不要在 I/O 周围加锁。


----------------

via: https://commandercoriander.net/blog/2018/04/10/dont-lock-around-io/

作者：[Eno](https://enocom.io/)
译者：[alfred-zhong](https://github.com/alfred-zhong)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出


