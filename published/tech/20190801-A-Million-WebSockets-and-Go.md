首发于：https://studygolang.com/articles/22501

# Go 实现百万 WebSocket 连接

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/a-million-websocket-and-go/image_1.jpeg)

大家好！我是 Sergey Kamardin，是 Mail.Ru 的一名工程师。

本文主要介绍如何使用 Go 开发高负载的 WebSocket 服务。

如果你熟悉 WebSockets，但对 Go 了解不多，仍希望你对这篇文章的想法和性能优化方面感兴趣。

## 1. 简介

为了定义本文的讨论范围，有必要说明我们为什么需要这个服务。

Mail.Ru 有很多有状态系统。用户的电子邮件存储就是其中之一。我们有几种方法可以跟踪该系统的状态变化以及系统事件，主要是通过定期系统轮询或者状态变化时的系统通知来实现。

两种方式各有利弊。但是对于邮件而言，用户收到新邮件的速度越快越好。

邮件轮询大约每秒 50,000 个 HTTP 查询，其中 60％ 返回 304 状态，这意味着邮箱中没有任何更改。

因此，为了减少服务器的负载并加快向用户发送邮件的速度，我们决定通过用发布 - 订阅服务（也称为消息总线，消息代理或事件管道）的模式来造一个轮子。一端接收有关状态更改的通知，另一端订阅此类通知。

之前的架构：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/a-million-websocket-and-go/image_2.png)

现在的架构：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/a-million-websocket-and-go/image_3.png)

第一个方案是之前的架构。浏览器定期轮询 API 并查询存储（邮箱服务）是否有更改。

第二种方案是现在的架构。浏览器与通知 API 建立了 WebSocket 连接，通知 API 是总线服务的消费者。一旦接收到新邮件后，Storage 会将有关它的通知发送到总线（1），总线将其发送给订阅者（2）。 API 通过连接发送这个收到的通知，将其发送到用户的浏览器（3）。

所以现在我们将讨论这个 API 或者这个 WebSocket 服务。展望一下未来，我们的服务将来可能会有 300 万个在线连接。

## 2. 常用的方式

我们来看看如何在没有任何优化的情况下使用 Go 实现服务器的某些部分。

在我们继续使用 `net/http` 之前，来谈谈如何发送和接收数据。这个数据位于 WebSocket 协议上（例如 JSON 对象），我们在下文中将其称为包。

我们先来实现 `Channel` 结构体，该结构体将包含在 WebSocket 连接上发送和接收数据包的逻辑。

### 2.1 Channel 结构体

```go
// WebSocket Channel 的实现
// Packet 结构体表示应用程序级数据
type Packet struct {
    ...
}

// Channel 装饰用户连接
type Channel struct {
    conn net.Conn    // WebSocket 连接
    send chan Packet // 传出的 packets 队列
}

func NewChannel(conn net.Conn) *Channel {
    c := &Channel{
        conn: conn,
        send: make(chan Packet, N),
    }

    go c.reader()
    go c.writer()

    return c
}
```

我想让你注意的是 `reader` 和 `writer` goroutines。每个 goroutine 都需要内存栈，初始大小可能为 2 到 8 KB，具体[取决于操作系统](https://github.com/golang/go/blob/release-branch.go1.8/src/runtime/stack.go#L64-L82)和 Go 版本。

关于上面提到的 300 万个线上连接，为此我们需要消耗 24 GB 的内存（假设单个 goroutine 消耗 4 KB 栈内存）用于所有的连接。并且这还没包括为 `Channel` 结构体分配的内存，`ch.send`传出的数据包占用的内存以及其他内部字段的内存。

### 2.2 I/O goroutines

让我们来看看 `reader` 的实现：

```go
// Channel’s reading goroutine.
func (c *Channel) reader() {
    // 创建一个缓冲 read 来减少 read 的系统调用
    buf := bufio.NewReader(c.conn)

    for {
        pkt, _ := readPacket(buf)
        c.handle(pkt)
    }
}
```

这里我们使用了 `bufio.Reader` 来减少 `read()` 系统调用的次数，并尽可能多地读取 `buf` 中缓冲区大小所允许的数量。在这个无限循环中，我们等待新数据的到来。请先记住这句话：*等待新数据的到来*。我们稍后会回顾。

我们先不考虑传入的数据包的解析和处理，因为它对我们讨论的优化并不重要。但是，`buf` 值得我们关注：默认情况下，它是 4 KB，这意味着连接还需要 12 GB 的内存。`writer` 也有类似的情况：

```go
// Channel’s writing goroutine.
func (c *Channel) writer() {
    // 创建一个缓冲 write 来减少 write 的系统调用
    buf := bufio.NewWriter(c.conn)

    for pkt := range c.send {
        _ := writePacket(buf, pkt)
        buf.Flush()
    }
}
```

我们通过 Channel 的 `c.send` 遍历将数据包传出 并将它们写入缓冲区。细心的读者可能猜到了，这是我们 300 万个连接的另外 12 GB 的内存消耗。

### 2.3 HTTP

已经实现了一个简单的 `Channel`，现在我们需要使用 WebSocket 连接。由于仍然处于常用的方式的标题下，所以我们以常用的方式继续。

> 注意：如果你不知道 WebSocket 的运行原理，需要记住客户端会通过名为 Upgrade 的特殊 HTTP 机制转换到 WebSocket 协议。在成功处理 Upgrade 请求后，服务端和客户端将使用 TCP 连接来传输二进制的 WebSocket 帧。[这里](https://tools.ietf.org/html/rfc6455#section-5.2)是连接的内部结构的说明。

```go
// 常用的转换为 WebSocket 的方法
import (
    "net/http"
    "some/websocket"
)

http.HandleFunc("/v1/ws", func(w http.ResponseWriter, r *http.Request) {
    conn, _ := websocket.Upgrade(r, w)
    ch := NewChannel(conn)
    //...
})
```

需要注意的是，`http.ResponseWriter` 为 `bufio.Reader` 和 `bufio.Writer`（均为 4 KB 的缓冲区）分配了内存，用于对 `*http.Request` 的初始化和进一步的响应写入。

无论使用哪种 WebSocket 库，在 Upgrade 成功后，[服务端在调用](https://github.com/golang/go/blob/143bdc27932451200f3c8f4b304fe92ee8bba9be/src/net/http/server.go#L1862-L1869) `responseWriter.Hijack()` 之后都会收到 I/O 缓冲区和 TCP 连接。

> 提示：在某些情况下，`go:linkname` 可被用于通过调用 `net/http.putBufio {Reader, Writer}` 将缓冲区返回给 `net/http` 内的 `sync.Pool`。

因此，我们还需要 24 GB 的内存用于 300 万个连接。

那么，现在为了一个什么功能都没有的应用程序，一共需要消耗 72 GB 的内存！

## 3. 优化

我们回顾一下在简介部分中谈到的内容，并记住用户连接的方式。在切换到 WebSocket 后，客户端会通过连接发送包含相关事件的数据包。然后（不考虑 `ping/pong` 等消息），客户端可能在整个连接的生命周期中不会发送任何其他内容。

> 连接的生命周期可能持续几秒到几天。

因此，大部分时间 `Channel.reader()` 和 `Channel.writer()` 都在等待接收或发送数据。与它们一起等待的还有每个大小为 4 KB 的 I/O 缓冲区。

现在我们对哪些地方可以做优化应该比较清晰了。

### 3.1 Netpoll

`Channel.reader()` 通过给 `bufio.Reader.Read()` 内的 `conn.Read()` 加锁来**等待新数据的到来**（译者注：上文中的伏笔），一旦连接中有数据，Go runtime（译者注：runtime 包含 Go 运行时的系统交互的操作，这里保留原文）“唤醒” goroutine 并允许它读取下一个数据包。在此之后，goroutine 再次被锁定，同时等待新的数据。让我们看看 Go runtime 来理解 goroutine 为什么必须“被唤醒”。

如果我们查看 [`conn.Read()` 的实现](https://github.com/golang/go/blob/release-branch.go1.8/src/net/net.go#L176-L186)，将会在其中看到 [`net.netFD.Read()` 调用](https://github.com/golang/go/blob/release-branch.go1.8/src/net/fd_unix.go#L245-L257)：

```go
// Go 内部的非阻塞读.
// net/fd_unix.go

func (fd *netFD) Read(p []byte) (n int, err error) {
    //...
    for {
        n, err = syscall.Read(fd.sysfd, p)
        if err != nil {
            n = 0
            if err == syscall.EAGAIN {
                if err = fd.pd.waitRead(); err == nil {
                    continue
                }
            }
        }
        //...
        break
    }
    //...
}
```

> Go 在非阻塞模式下使用套接字。 EAGAIN 表示套接字中没有数据，并且读取空套接字时不会被锁定，操作系统将返回控制权给我们。(译者注：EAGAIN 表示目前没有可用数据，请稍后再试)

我们从连接文件描述符中看到一个 `read()` 系统调用函数。如果 read 返回 [EAGAIN 错误](http://man7.org/linux/man-pages/man2/read.2.html#ERRORS)，则 runtime 调用 [pollDesc.waitRead()](https://github.com/golang/go/blob/release-branch.go1.8/src/net/fd_poll_runtime.go#L74-L81)：

```go
// Go 内部关于 netpoll 的使用
// net/fd_poll_runtime.go

func (pd *pollDesc) waitRead() error {
   return pd.wait('r')
}

func (pd *pollDesc) wait(mode int) error {
   res := runtime_pollWait(pd.runtimeCtx, mode)
   //...
}
```

如果[深入挖掘](https://github.com/golang/go/blob/143bdc27932451200f3c8f4b304fe92ee8bba9be/src/runtime/netpoll.go#L14-L20)，我们将看到 netpoll 在 Linux 中是使用 [epoll](http://man7.org/linux/man-pages/man7/epoll.7.html) 实现的，而在 BSD 中是使用 [kqueue](https://www.freebsd.org/cgi/man.cgi?query=kqueue&sektion=2) 实现的。为什么不对连接使用相同的方法？我们可以分配一个 read 缓冲区并仅在真正需要时启动 read goroutine：当套接字中有可读的数据时。

> 在 github.com/golang/go 上，有一个导出 netpoll 函数的 [issue](https://github.com/golang/go/issues/15735#issuecomment-266574151)。

### 3.2 去除 goroutines 的内存消耗

假设我们有 Go 的 [netpoll 实现](https://godoc.org/github.com/mailru/easygo/netpoll)。现在我们可以避免在内部缓冲区启动 `Channel.reader()` goroutine，而是在连接中订阅可读数据的事件：

```go
// 使用 netpoll
ch := NewChannel(conn)

// 通过 netpoll 实例观察 conn
poller.Start(conn, netpoll.EventRead, func() {
    // 我们在这里产生 goroutine 以防止在轮询从 ch 接收数据包时被锁。
    go Receive(ch)
})

// Receive 从 conn 读取数据包并以某种方式处理它。
func (ch *Channel) Receive() {
    buf := bufio.NewReader(ch.conn)
    pkt := readPacket(buf)
    c.handle(pkt)
}
```

`Channel.writer()` 更简单，因为我们只能在发送数据包时运行 goroutine 并分配缓冲区：

```go
// 当我们需要时启动 writer goroutine
func (ch *Channel) Send(p Packet) {
    if c.noWriterYet() {
        go ch.writer()
    }
    ch.send <- p
}
```

> 需要注意的是，当操作系统在 `write()` 调用上返回 `EAGAIN` 时，我们不处理这种情况。我们依靠 Go runtime 来处理这种情况，因为这种情况在服务器上很少见。然而，如果有必要，它可以以与 `reader()` 相同的方式处理。

当从 `ch.send`（一个或几个）读取传出数据包后，writer 将完成其操作并释放 goroutine 的内存和发送缓冲区的内存。

完美！我们通过去除两个运行的 goroutine 中的内存消耗和 I/O 缓冲区的内存消耗节省了 48 GB。

### 3.3 资源控制

大量连接不仅仅涉及到内存消耗高的问题。在开发服务时，我们遇到了反复出现的竞态条件和 self-DDoS 造成的死锁。

例如，如果由于某种原因我们突然无法处理 `ping/pong` 消息，但是空闲连接的处理程序继续关闭这样的连接（假设连接被破坏，没有提供数据），客户端每隔 N 秒失去连接并尝试再次连接而不是等待事件。

被锁或超载的服务器停止服务，如果它之前的负载均衡器（例如，nginx）将请求传递给下一个服务器实例，这将是不错的。

此外，无论服务器负载如何，如果所有客户端突然（可能是由于错误原因）向我们发送数据包，之前的 48 GB 内存的消耗将不可避免，因为需要为每个连接分配 goroutine 和缓冲区。

#### Goroutine 池

上面的情况，我们可以使用 goroutine 池限制同时处理的数据包数量。下面是这种池的简单实现：

```go
// goroutine 池的简单实现
package gopool

func New(size int) *Pool {
    return &Pool{
        work: make(chan func()),
        sem:  make(chan struct{}, size),
    }
}

func (p *Pool) Schedule(task func()) error {
    select {
    case p.work <- task:
    case p.sem <- struct{}{}:
        go p.worker(task)
    }
}

func (p *Pool) worker(task func()) {
    defer func() { <-p.sem }
    for {
        task()
        task = <-p.work
    }
}
```

现在我们的 netpoll 代码如下：

```go
// 处理 goroutine 池中的轮询事件。
pool := gopool.New(128)

poller.Start(conn, netpoll.EventRead, func() {
    // 我们在所有 worker 被占用时阻塞 poller
    pool.Schedule(func() {
        Receive(ch)
    })
})
```

现在我们不仅在套接字中有可读数据时读取，而且还在第一次机会获取池中的空闲 goroutine。??

同样，我们修改 `Send()`：

```go
// 复用 writing goroutine
pool := gopool.New(128)

func (ch *Channel) Send(p Packet) {
    if c.noWriterYet() {
        pool.Schedule(ch.writer)
    }
    ch.send <- p
}
```

取代 `go ch.writer()` ，我们想写一个复用的 goroutines。因此，对于拥有 `N` 个 goroutines 的池，我们可以保证同时处理 `N` 个请求并且在 `N + 1`的时候， 我们不会分配 `N + 1` 个缓冲区。 goroutine 池还允许我们限制新连接的 `Accept()` 和 `Upgrade()` ，并避免大多数的 DDoS 攻击。

### 3.4 upgrade 零拷贝

如前所述，客户端使用 HTTP Upgrade 切换到 WebSocket 协议。这就是 WebSocket 协议的样子：

```plain
## HTTP Upgrade 示例

GET /ws HTTP/1.1
Host: mail.ru
Connection: Upgrade
Sec-Websocket-Key: A3xNe7sEB9HixkmBhVrYaA==
Sec-Websocket-Version: 13
Upgrade: websocket

HTTP/1.1 101 Switching Protocols
Connection: Upgrade
Sec-Websocket-Accept: ksu0wXWG+YmkVx+KQR2agP0cQn4=
Upgrade: websocket
```

也就是说，在我们的例子中，需要 HTTP 请求及其 Header 用于切换到 WebSocket 协议。这些知识以及 [`http.Request` 中存储的内容](https://github.com/golang/go/blob/release-branch.go1.8/src/net/http/request.go#L100-L305)表明，为了优化，我们需要在处理 HTTP 请求时放弃不必要的内存分配和内存复制，并弃用 `net/http` 库。

> 例如，`http.Request` 有一个与 [Header 具有相同名称的字段](https://github.com/golang/go/blob/release-branch.go1.8/src/net/http/header.go#L19)，这个字段用于将数据从连接中复制出来填充请求头。想象一下，该字段需要消耗多少额外内存，例如碰到比较大的 Cookie 头。

#### WebSocket 的实现

不幸的是，在我们优化的时候所有存在的库都是使用标准的 `net/http` 库进行升级。而且，（两个）库都不能使用上述的读写优化方案。为了采用这些优化方案，我们需要用一个比较低级的 API 来处理 WebSocket。要重用缓冲区，我们需要把协议函数变成这样：

```go
func ReadFrame(io.Reader) (Frame, error)
func WriteFrame(io.Writer, Frame) error
```

如果有一个这种 API 的库，我们可以按下面的方式从连接中读取数据包（数据包的写入也一样）：

```go
// 预期的 WebSocket 实现API
// getReadBuf, putReadBuf 用来复用 *bufio.Reader (with sync.Pool for example).
func getReadBuf(io.Reader) *bufio.Reader
func putReadBuf(*bufio.Reader)

// 当 conn 中的数据可读取时，readPacket 被调用
func readPacket(conn io.Reader) error {
    buf := getReadBuf()
    defer putReadBuf(buf)

    buf.Reset(conn)
    frame, _ := ReadFrame(buf)
    parsePacket(frame.Payload)
    //...
}
```

简单来说，我们需要自己的 WebSocket 库。

#### github.com/gobwas/ws

在意识形态上，编写 `ws` 库是为了不将其协议操作逻辑强加给用户。所有读写方法都实现了标准的 io.Reader 和 io.Writer 接口，这样就可以使用或不使用缓冲或任何其他 I/O 包装器。

除了来自标准库 `net/http` 的升级请求之外，`ws` 还支持零拷贝升级，升级请求的处理以及切换到 WebSocket 无需分配内存或复制内存。`ws.Upgrade()` 接受 `io.ReadWriter`（`net.Conn` 实现了此接口）。换句话说，我们可以使用标准的 `net.Listen()` 将接收到的连接从 `ln.Accept()` 转移给 `ws.Upgrade()` 。该库使得可以复制任何请求数据以供应用程序使用（例如，`Cookie` 用来验证会话）。

下面是升级请求的[基准测试](https://github.com/gobwas/ws/blob/f9c54e121bd17f7e6b9b283bd0299d19149f270b/server_test.go#L397-L464)结果：标准库 `net/http` 的服务与用零拷贝升级的 `net.Listen()`：

```plain
BenchmarkUpgradeHTTP    5156 ns/op    8576 B/op    9 allocs/op
BenchmarkUpgradeTCP     973 ns/op     0 B/op       0 allocs/op
```

切换到 `ws` 和**零拷贝升级**为我们节省了另外的 24 GB 内存  - 在 `net/http` 处理请求时为 I/O 缓冲区分配的空间。

## 3.5 摘要

我们总结一下这些优化。

- 内部有缓冲区的 read goroutine 是代价比较大的。解决方案：netpoll（epoll，kqueue）; 重用缓冲区。
- 内部有缓冲区的 write goroutine 是代价比较大的。解决方案：需要的时候才启动 goroutine; 重用缓冲区。
- 如果有大量的连接，netpoll 将无法正常工作。解决方案：使用 goroutines 池并限制池的 worker 数。
- `net/http` 不是处理升级到 WebSocket 的最快方法。解决方案：在裸 TCP 连接上使用内存零拷贝升级。

服务的代码看起来如下所示：

```go
// WebSocket 服务器示例，包含 netpoll，goroutine 池和内存零拷贝的升级。
import (
    "net"
    "github.com/gobwas/ws"
)

ln, _ := net.Listen("tcp", ":8080")

for {
    // Try to accept incoming connection inside free pool worker.
    // If there no free workers for 1ms, do not accept anything and try later.
    // This will help us to prevent many self-ddos or out of resource limit cases.
    // 尝试接受空闲池 worker 内的传入连接。如果 1ms 没有空闲 worker，则稍后再试。这有助于防止 self-ddos 或耗尽服务器资源的情况。
    err := pool.ScheduleTimeout(time.Millisecond, func() {
        conn := ln.Accept()
        _ = ws.Upgrade(conn)

        // Wrap WebSocket connection with our Channel struct.
        // This will help us to handle/send our app's packets.
        ch := NewChannel(conn)

        // Wait for incoming bytes from connection.
        poller.Start(conn, netpoll.EventRead, func() {
            // Do not cross the resource limits.
            pool.Schedule(func() {
                // Read and handle incoming packet(s).
                ch.Recevie()
            })
        })
    })
    if err != nil {
        time.Sleep(time.Millisecond)
    }
}
```

## 总结

> 过早优化是编程中所有邪恶（或至少大部分）的根源。
-- Donald Knuth

当然，上述优化是和需求相关的，但并非所有情况下都是如此。例如，如果空闲资源（内存，CPU）和线上连接数之间的比率比较高，则优化可能没有意义。但是，通过了解优化的位置和内容，我们会受益匪浅。

感谢你的关注！

## 引用

- [https://github.com/mailru/easygo](https://github.com/mailru/easygo)
- [https://github.com/gobwas/ws](https://github.com/gobwas/ws)
- [https://github.com/gobwas/ws-examples](https://github.com/gobwas/ws-examples)
- [https://github.com/gobwas/httphead](https://github.com/gobwas/httphead)
- [Russian version of this article](https://habrahabr.ru/company/mailru/blog/331784/)

---

via: https://www.freecodecamp.org/news/million-websockets-and-go-cc58418460bb/

作者：[Sergey Kamardin](https://www.freecodecamp.org)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
