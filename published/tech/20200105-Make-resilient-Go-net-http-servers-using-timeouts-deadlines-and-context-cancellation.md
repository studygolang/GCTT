首发于：https://studygolang.com/articles/28444

# 使用 timeout、deadline 和 context 取消参数使 Go net/http 服务更灵活

关于超时，可以把开发者分为两类：一类是了解超时多么难以捉摸的人，另一类是正在感受超时如何难以捉摸的人。

超时既难以捉摸，却又真实地存在于我们生活的由网络连接的世界中。在我写这篇文章的同时，隔壁两个同事正在用他们的智能手机打字，也许是在跟与他们相距万里的人聊天。网络使这一切变为可能。

这里要说的是网络及其复杂性，作为写网络服务的我们，必须掌握如何高效地驾驭它们，并规避它们的缺陷。

闲话少说，来看看超时和它们是如何影响我们的 `net/http` 服务的。

## 服务超时 — 基本原理

web 编程中，超时通常分为客户端和服务端超时两种。我之所以要研究这个主题，是因为我自己遇到了一个有意思的服务端超时的问题。这也是本文我们将要重点讨论服务侧超时的原因。

先解释下基本术语：超时是一个时间间隔（或边界），用来标识在这个时间段内要完成特定的行为。如果在给定的时间范围内没有完成操作，就产生了超时，这个操作会被取消。

从一个 `net/http` 的服务的初始化中，能看出一些超时的基础配置：

```go
srv := &http.Server{
    ReadTimeout:       1 * time.Second,
    WriteTimeout:      1 * time.Second,
    IdleTimeout:       30 * time.Second,
    ReadHeaderTimeout: 2 * time.Second,
    TLSConfig:         tlsConfig,
    Handler:           srvMux,
}
```

`http.Server` 类型的服务可以用四个不同的 timeout 来初始化：

- `ReadTimeout`：读取包括请求体的整个请求的最大时长
- `WriteTimeout`：写响应允许的最大时长
- `IdleTimetout`：当开启了保持活动状态（keep-alive）时允许的最大空闲时间
- `ReadHeaderTimeout`：允许读请求头的最大时长

对上述超时的图表展示：

![Server lifecycle and timeouts](https://raw.githubusercontent.com/studygolang/gctt-images2/master/context-cancel-deadline-timeout/request-lifecycle-timeouts.png)服务生命周期和超时

当心！不要以为这些就是你所需要的所有的超时了。除此之外还有很多超时，这些超时提供了更小的粒度控制，对于我们的持续运行的 HTTP 处理器不会生效。

请听我解释。

## timeout 和 deadline

如果我们查看 `net/http` 的源码，尤其是看到 [`conn` 类型](https://github.com/golang/go/blob/bbbc658/src/net/http/server.go#L248) 时，我们会发现 `conn` 实际上使用了 `net.Conn` 连接，`net.Conn` 表示底层的网络连接：

```go
// Taken from: https://github.com/golang/go/blob/bbbc658/src/net/http/server.go#L247
// A conn represents the server-side of an HTTP connection.
type conn struct {
    // server is the server on which the connection arrived.
    // Immutable; never nil.
    server *Server

    // * Snipped *

    // rwc is the underlying network connection.
    // This is never wrapped by other types and is the value given out
    // to CloseNotifier callers. It is usually of type *net.TCPConn or
    // *tls.Conn.
    rwc net.Conn

    // * Snipped *
}
```

换句话说，我们的 HTTP 请求实际上是基于 TCP 连接的。从类型上看，TLS 连接是  `*net.TCPConn` 或 `*tls.Conn` 。

`serve` [函数](https://github.com/golang/go/blob/bbbc658/src/net/http/server.go#L1765)[处理每一个请求](https://github.com/golang/go/blob/bbbc658/src/net/http/server.go#L1822)时调用 `readRequest` 函数。 `readRequest` 使用我们设置的 [timeout 值](https://github.com/golang/go/blob/bbbc658/src/net/http/server.go#L946-L958)**来设置 TCP 连接的 deadline**：

```go
// Taken from: https://github.com/golang/go/blob/bbbc658/src/net/http/server.go#L936
// Read next request from connection.
func (c *conn) readRequest(ctx context.Context) (w *response, err error) {
        // *Snipped*

        t0 := time.Now()
        if d := c.server.readHeaderTimeout(); d != 0 {
                hdrDeadline = t0.Add(d)
        }
        if d := c.server.ReadTimeout; d != 0 {
                wholeReqDeadline = t0.Add(d)
        }
        c.rwc.SetReadDeadline(hdrDeadline)
        if d := c.server.WriteTimeout; d != 0 {
                defer func() {
                        c.rwc.SetWriteDeadline(time.Now().Add(d))
                }()
        }

        // *Snipped*
}
```

从上面的摘要中，我们可以知道：我们对服务设置的 timeout 值最终表现为 TCP 连接的 deadline 而不是 HTTP 超时。

所以，deadline 是什么？工作机制是什么？如果我们的请求耗时过长，它们会取消我们的连接吗？

一种简单地理解 deadline 的思路是，把它理解为对作用于连接上的特定的行为的发生限制的一个时间点。例如，如果我们设置了一个写的 deadline，当过了这个 deadline 后，所有对这个连接的写操作都会被拒绝。

尽管我们可以使用 deadline 来模拟超时操作，但我们还是不能控制处理器完成操作所需的耗时。deadline 作用于连接，因此我们的服务仅在处理器尝试访问连接的属性（如对 `http.ResponseWriter` 进行写操作）之后才会返回（错误）结果。

为了实际验证上面的论述，我们来创建一个小的 handler，这个 handler 完成操作所需的耗时相对于我们为服务设置的超时更长：

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func slowHandler(w http.ResponseWriter, req *http.Request) {
	time.Sleep(2 * time.Second)
	io.WriteString(w, "I am slow!\n")
}

func main() {
	srv := http.Server{
		Addr:         ":8888",
		WriteTimeout: 1 * time.Second,
		Handler:      http.HandlerFunc(slowHandler),
	}

	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
```

上面的服务有一个 handler，这个 handler 完成操作需要两秒。另一方面，`http.Server` 的 `WriteTimeout` 属性设为 1 秒。基于服务的这些配置，我们猜测 handler 不能把响应写到连接。

我们可以用 `go run server.go` 来启动服务。使用 `curl localhost:8888` 来发送一个请求：

```shell
$ time curl localhost:8888
curl: (52) Empty reply from server
curl localhost:8888  0.01s user 0.01s system 0% CPU 2.021 total
```

这个请求需要两秒来完成处理，服务返回的响应是空的。虽然我们的服务知道在 1 秒之后我们写不了响应了，但 handler 还是多耗了 100% 的时间（2 秒）来完成处理。

虽然这是个类似超时的处理，但它更大的作用是在到达超时时间时，阻止服务进行更多的操作，结束请求。在我们上面的例子中，handler 在完成之前一直在处理请求，即使已经超出响应写超时时间（1 秒）100%（耗时 2 秒）。

最根本的问题是，对于处理器来说，我们应该怎么设置超时时间才更有效？

## 处理超时

我们的目标是确保我们的 `slowHandler` 在 1s 内完成处理。如果超过了 1s，我们的服务会停止运行并返回对应的超时错误。

在 Go 和一些其它编程语言中，组合往往是设计和开发中最好的方式。标准库的 [`net/http`  包](https://golang.org/pkg/net/http)有很多相互兼容的元素，开发者可以不需经过复杂的设计考虑就可以轻易将它们组合在一起。

基于此，`net/http` 包提供了[`TimeoutHandler`](https://golang.org/pkg/net/http/#TimeoutHandler) — 返回了一个在给定的时间限制内运行的 handler。

函数签名：

```go
func TimeoutHandler(h Handler, dt time.Duration, msg string) Handler
```

第一个参数是 `Handler`，第二个参数是 `time.Duration` （超时时间），第三个参数是 `string` 类型，当到达超时时间后返回的信息。

用 `TimeoutHandler` 来封装我们的 `slowHandler`，我们只需要：

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func slowHandler(w http.ResponseWriter, req *http.Request) {
	time.Sleep(2 * time.Second)
	io.WriteString(w, "I am slow!\n")
}

func main() {
	srv := http.Server{
		Addr:         ":8888",
		WriteTimeout: 5 * time.Second,
		Handler:      http.TimeoutHandler(http.HandlerFunc(slowHandler), 1*time.Second, "Timeout!\n"),
	}

	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
```

两个需要留意的地方是：

- 我们在 `http.TimetoutHandler` 里封装 `slowHanlder`，超时时间设为 1s，超时信息为 “Timeout!”。
- 我们把 `WriteTimeout` 增加到 5s，以给予 `http.TimeoutHandler` 足够的时间执行。如果我们不这么做，当 `TimeoutHandler` 开始执行时，已经过了 deadline，不能再写到响应。

如果我们再启动服务，当程序运行到 slow handler 时，会有如下输出：

```shell
$ time curl localhost:8888
Timeout!
curl localhost:8888  0.01s user 0.01s system 1% CPU 1.023 total
```

1s 后，我们的 `TimeoutHandler` 开始执行，阻止运行 `slowHandler`，返回文本信息 ”Timeout!“。如果我们设置信息为空，handler 会返回默认的超时响应信息，如下：

```html
<html>
  <head>
    <title>Timeout</title>
  </head>
  <body>
   <h1>Timeout</h1>
  </body>
</html>
```

如果忽略掉输出，这还算是整洁，不是吗？现在我们的程序不会有过长耗时的处理；也避免了有人恶意发送导致长耗时处理的请求时，导致的潜在的 DoS 攻击。

尽管我们设置超时时间是一个伟大的开始，但它仍然只是初级的保护。如果你可能会面临 DoS 攻击，你应该采用更高级的保护工具和技术。（可以试试 [Cloudflare](https://www.cloudflare.com/ddos/) ）

我们的 `slowHandler` 仅仅是个简单的 demo。但是，如果我们的程序复杂些，能向其他服务和资源发出请求会发生什么呢？如果我们的程序在超时时向诸如 S3 的服务发出了请求会怎么样？

会发生什么？

## 未处理的超时和请求取消

我们稍微展开下我们的例子：

```go
func slowAPICall() string {
	d := rand.Intn(5)
	select {
	case <-time.After(time.Duration(d) * time.Second):
		log.Printf("Slow API call done after %s seconds.\n", d)
		return "foobar"
	}
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	result := slowAPICall()
	io.WriteString(w, result+"\n")
}
```

我们假设最初我们不知道 `slowHandler` 由于通过 `slowAPICall` 函数向 API 发请求导致需要耗费这么长时间才能处理完成，

`slowAPICall` 函数很简单：使用 `select` 和一个能阻塞 0 到 5 秒的 `time.After` 。当经过了阻塞的时间后，`time.After` 方法通过它的 channel 发送一个值，返回 `"foobar"` 。

（另一种方法是，使用 `sleep(time.Duration(rand.Intn(5)) * time.Second)`，但我们仍然使用 `select`，因为它会使我们下面的例子更简单。）

如果我们运行起服务，我们预期超时 handler 会在 1 秒之后中断请求处理。来发送一个请求验证一下：

```shell
$ time curl localhost:8888
Timeout!
curl localhost:8888  0.01s user 0.01s system 1% CPU 1.021 total
```

通过观察服务的输出，我们会发现，它是在几秒之后打出日志的，而不是在超时 handler 生效时打出：

```shell
$ Go run server.go
2019/12/29 17:20:03 Slow API call done after 4 seconds.
```

这个现象表明：虽然 1 秒之后请求超时了，但是服务仍然完整地处理了请求。这就是在 4 秒之后才打出日志的原因。

虽然在这个例子里问题很简单，但是类似的现象在生产中可能变成一个严重的问题。例如，当 `slowAPICall` 函数开启了几个百个协程，每个协程都处理一些数据时。或者当它向不同系统发出多个不同的 API 发出请求时。这种耗时长的的进程，它们的请求方/客户端并不会使用服务端的返回结果，会耗尽你系统的资源。

所以，我们怎么保护系统，使之不会出现类似的未优化的超时或取消请求呢？

## 上下文超时和取消

Go 有一个包名为  [`context`](https://golang.org/pkg/context/) 专门处理类似的场景。

`context` 包在 Go 1.7 版本中提升为标准库，在之前的版本中，以 [`golang.org/x/net/context`](https://godoc.org/golang.org/x/net/context) 的路径作为 [Go Sub-repository Packages](https://godoc.org/-/subrepo) 出现。

这个包定义了 `Context` 类型。它最初的目的是保存不同 API 和不同处理的截止时间、取消信号和其他请求相关的值。如果你想了解关于 context 包的其他信息，可以阅读  [Golang's blog](https://blog.golang.org/context) 中的 “Go 并发模式：Context”（译注：Go Concurrency Patterns: Context） .

 `net/http` 包中的的 `Request` 类型已经有 `context` 与之绑定。从 Go 1.7 开始，`Request` 新增了一个返回请求的上下文的  [`Context` 方法](https://golang.org/pkg/net/http/#Request.Context)。对于进来的请求，在客户端关闭连接、请求被取消（HTTP/2 中）或 `ServeHTTP` 方法返回后，服务端会取消上下文。

我们期望的现象是，当客户端取消请求（输入了 `CTRL + C`）或一段时间后 `TimeoutHandler`  继续执行然后终止请求时，服务端会停止后续的处理。进而关闭所有的连接，释放所有被运行中的处理进程（及它的所有子协程）占用的资源。

我们把 `Context` 作为参数传给 `slowAPICall` 函数：

```go
func slowAPICall(ctx context.Context) string {
	d := rand.Intn(5)
	select {
	case <-time.After(time.Duration(d) * time.Second):
		log.Printf("Slow API call done after %d seconds.\n", d)
		return "foobar"
	}
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	result := slowAPICall(r.Context())
	io.WriteString(w, result+"\n")
}
```

在例子中我们利用了请求上下文，实际中怎么用呢？[`Context` 类型](https://golang.org/pkg/context/#Context)有个 `Done` 属性，类型为 `<-chan struct{}`。当进程处理完成时，`Done` 关闭，此时表示上下文应该被取消，而这正是例子中我们需要的。

我们在 `slowAPICall` 函数中用 `select` 处理 `ctx.Done` 通道。当我们通过 `Done` 通道接收一个空的 `struct` 时，意味着上下文取消，我们需要让 `slowAPICall` 函数返回一个空字符串。

```go
func slowAPICall(ctx context.Context) string {
	d := rand.Intn(5)
	select {
	case <-ctx.Done():
		log.Printf("slowAPICall was supposed to take %s seconds, but was canceled.", d)
		return ""
        //time.After() 可能会导致内存泄漏
	case <-time.After(time.Duration(d) * time.Second):
		log.Printf("Slow API call done after %d seconds.\n", d)
		return "foobar"
	}
}
```

（这就是使用 `select` 而不是 `time.Sleep` -- 这里我们只能用 `select` 处理 `Done` 通道。 ）

在这个简单的例子中，我们成功得到了结果 -- 当我们从 `Done` 通道接收值时，我们打印了一行日志到 STDOUT 并返回了一个空字符串。在更复杂的情况下，如发送真实的 API 请求，你可能需要关闭连接或清理文件描述符。

我们再启动服务，发送一个 `cRUL` 请求：

```shell
# The cURL command:
$ curl localhost:8888
Timeout!

# The server output:
$ Go run server.go
2019/12/30 00:07:15 slowAPICall was supposed to take 2 seconds, but was canceled.
```

检查输出：我们发送了 `cRUL` 请求到服务，它耗时超过 1 秒，服务取消了 `slowAPICall` 函数。我们几乎不需要写任何代码。`TimeoutHandler` 为我们代劳了 -- 当处理耗时超过预期时，`TimeoutHandler` 终止了处理进程并取消请求上下文。

`TimeoutHandler` 是在 [`timeoutHandler.ServeHTTP` 方法](https://github.com/golang/go/blob/bbbc658/src/net/http/server.go#L3217-L3263) 中取消上下文的：

```go
// Taken from: https://github.com/golang/go/blob/bbbc658/src/net/http/server.go#L3217-L3263
func (h *timeoutHandler) ServeHTTP(w ResponseWriter, r *Request) {
        ctx := h.testContext
        if ctx == nil {
        	var cancelCtx context.CancelFunc
        	ctx, cancelCtx = context.WithTimeout(r.Context(), h.dt)
        	defer cancelCtx()
        }
        r = r.WithContext(ctx)

        // *Snipped*
}
```

上面例子中，我们通过调用 `context.WithTimeout` 来使用请求上下文。超时值 `h.dt`  （`TimeoutHandler` 的第二个参数）设置给了上下文。返回的上下文是请求上下文设置了超时值后的一份拷贝。随后，它作为请求上下文传给 `r.WithContext(ctx)`。

`context.WithTimeout` 方法执行了上下文取消。它返回了 `Context` 设置了一个超时值之后的副本。当到达超时时间后，就取消上下文。

这里是执行的代码：

```go
// Taken from: https://github.com/golang/go/blob/bbbc6589/src/context/context.go#L486-L498
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return WithDeadline(parent, time.Now().Add(timeout))
}

// Taken from: https://github.com/golang/go/blob/bbbc6589/src/context/context.go#L418-L450
func WithDeadline(parent Context, d time.Time) (Context, CancelFunc) {
        // *Snipped*

        c := &timerCtx{
        	cancelCtx: newCancelCtx(parent),
        	deadline:  d,
        }

        // *Snipped*

        if c.err == nil {
        	c.timer = time.AfterFunc(dur, func() {
        		c.cancel(true, DeadlineExceeded)
        	})
        }
        return c, func() { c.cancel(true, Canceled) }
}
```

这里我们又看到了截止时间。`WithDeadline` 函数设置了一个 `d` 到达之后执行的函数。当到达截止时间后，它调用 `cancel` 方法处理上下文，此方法会关闭上下文的 `done` 通道并设置上下文的 `timer` 属性为 `nil`。

`Done` 通道的关闭有效地取消了上下文，使我们的 `slowAPICall` 函数终止了它的执行。这就是 `TimeoutHandler` 终止耗时长的处理进程的原理。

（如果你想阅读上面提到的源码，你可以去看 [`cancelCtx` 类型](https://github.com/golang/go/blob/bbbc6589dfbc05be2bfa59f51c20f9eaa8d0c531/src/context/context.go#L389-L416) 和 [`timerCtx` 类型](https://github.com/golang/go/blob/bbbc6589dfbc05be2bfa59f51c20f9eaa8d0c531/src/context/context.go#L472-L484)）

## 有弹性的 `net/http` 服务

连接截止时间提供了低级的细粒度控制。虽然它们的名字中含有“超时”，但它们并没有表现出人们通常期望的“超时”。实际上它们非常强大，但是使用它们有一定的门槛。

另一个角度讲，当处理 HTTP 时，我们仍然应该考虑使用 `TimeoutHandler`。Go 的作者们也选择使用它，它有多种处理，提供了如此有弹性的处理以至于我们甚至可以对每一个处理使用不同的超时。`TimeoutHandler` 可以根据我们期望的表现来控制执行进程。

除此之外，`TimeoutHandler` 完美兼容 `context` 包。`context` 包很简单，包含了取消信号和请求相关的数据，我们可以使用这些数据来使我们的应用更好地处理错综复杂的网络问题。

结束之前，有三个建议。写 HTTP 服务时，怎么设计超时：

1. 最常用的，到达 `TimeoutHandler` 时，怎么处理。它进行我们通常期望的超时处理。
2. 不要忘记上下文取消。`context` 包使用起来很简单，并且可以节省你服务器上的很多处理资源。尤其是在处理异常或网络状况不好时。
3. 一定要用截止时间。确保做了完整的测试，验证了能提供你期望的所有功能。

更多关于此主题的文章：

- “The complete guide to Go net/http timeouts” on [Cloudflare's blog](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)
- “So you want to expose Go on the Internet” on [Cloudflare's blog](https://blog.cloudflare.com/exposing-go-on-the-internet/)
- “Use http.TimeoutHandler or ReadTimeout/WriteTimeout?” on [Stackoverflow](https://stackoverflow.com/questions/51258952/use-http-timeouthandler-or-readtimeout-writetimeout)
- “Standard net/http config will break your production environment” on [Simon Frey's blog](https://blog.simon-frey.eu/go-as-in-golang-standard-net-http-config-will-break-your-production)

---

via: https://ieftimov.com/post/make-resilient-golang-net-http-servers-using-timeouts-deadlines-context-cancellation/

作者：[Ilija Eftimov](https://ieftimov.com/)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
