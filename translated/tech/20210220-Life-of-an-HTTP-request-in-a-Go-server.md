# Go 服务中 HTTP 请求的生命周期

Go 语言对于编写 HTTP 服务来说是一个常见且非常合适的工具。这篇博文通过一个 Go 服务来探讨一个典型 HTTP 请求的路由，涉及路由，中间件以及比如并发之类的相关问题。

为了有具体的代码可以参考，让我们先从这段简单的服务代码开始（来自于 [https://gobyexample.com/http-servers](https://gobyexample.com/http-servers)）

```go
package main

import (
  "fmt"
  "net/http"
)

func hello(w http.ResponseWriter, req *http.Request) {
  fmt.Fprintf(w, "hello\n")
}

func headers(w http.ResponseWriter, req *http.Request) {
  for name, headers := range req.Header {
    for _, h := range headers {
      fmt.Fprintf(w, "%v: %v\n", name, h)
    }
  }
}

func main() {
  http.HandleFunc("/hello", hello)
  http.HandleFunc("/headers", headers)

  http.ListenAndServe(":8090", nil)
}
```

我们会通过查看 `http.ListenAndServe` 函数来开始跟踪一个 HTTP 请求在这个服务中的生命周期：

```go
func ListenAndServe(addr string, handler Handler) error
```

这张图展示了调用时所发生的简要流程：

![](https://github.com/studygolang/gctt-images2/blob/master/20210220-Life-of-an-HTTP-request-in-a-Go-server/http-request-listenandserve.png?raw=true)

这是函数和方法调用的实际序列的高度“内联”版本，但是[原始的代码](https://go.googlesource.com/go/+/go1.15.8/src/net/http/server.go)并不难理解。

主流程正如你期望的那样：`ListenAndServe` 监听给定地址的 TCP 端口，之后循环接受新的连接。对于每一个新连接，它都会调度一个 goroutine 来处理这个连接（稍后详细说明）。处理连接涉及一个这样的循环：

- 从连接中解析 HTTP 请求；产生 `http.Request`
- 将这个 `http.Request` 传递给用户定义的 handler

handler 是一个实现 `http.Handler` 接口的任意实例：

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

## 默认的 handler

在我们的实例代码中，`ListenAndServe` 被调用的时候使用 `nil` 作为第二个参数，而这个位置本应该使用用户定义的 handler，这是怎么回事？

我们的图简化了一些细节；实际上，当这个 HTTP 包处理一个请求的时候，它并不会直接调用用户的 handler，而是使用这个适配器：

```go
type serverHandler struct {
  srv *Server
}

func (sh serverHandler) ServeHTTP(rw ResponseWriter, req *Request) {
  handler := sh.srv.Handler
  if handler == nil {
    handler = DefaultServeMux
  }
  if req.RequestURI == "*" && req.Method == "OPTIONS" {
    handler = globalOptionsHandler{}
  }
  handler.ServeHTTP(rw, req)
}
```

注意高亮的部分（if handler == nil ...），如果 `handler == nil`，则 `http.DefaultServeMux` 被用作 handler。这个是*默认的 server mux*，`http` 包中所包含的一个 `http.ServeMux` 类型的全局实例。顺便一提，当我们的示例代码使用 `http.HandleFunc` 注册 handler 函数的时候，会在同一个默认的 mux 上注册这些handler。

我们可以如下所示这样重写我们的示例代码，不再使用默认的 mux。只修改 `main` 函数，所以这里没有展示 `hello` 和 `headers` handler 函数，看是我们可以在这看[完整的代码](https://github.com/eliben/code-for-blog/blob/master/2021/go-life-http-request/basic-server-mux-object.go)。功能上没有任何变化[^1]：

```go
func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/hello", hello)
  mux.HandleFunc("/headers", headers)

  http.ListenAndServe(":8090", mux)
}
```
## 一个 `ServeMux` 仅仅是一个 `Handler`
当看多了 Go 服务的例子后，很容易给人一种 `ListenAndServe` 函数“需要一个 mux” 作为参数的印象，但是这是不准确的。就像我们之前所见到的那样，`ListenAndServe` 函数需要的是一个实现了 `http.Handler` 接口的值。我们可以写下面这样的服务而没有任何 mux：

```go
type PoliteServer struct {
}

func (ms *PoliteServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  fmt.Fprintf(w, "Welcome! Thanks for visiting!\n")
}

func main() {
  ps := &PoliteServer{}
  log.Fatal(http.ListenAndServe(":8090", ps))
}
```

由于这里没有路由逻辑；所有到达 `PoliteServer` 的 `ServeHTTP` 方法的 HTTP 请求都会以同样的信息所回复。试着用不同的路径和方法 `curl` -ing 这个服务；返回一定是一致的。

我们可以使用 `http.HandlerFunc` 来进一步简化我们的 polite 服务：

```go
func politeGreeting(w http.ResponseWriter, req *http.Request) {
  fmt.Fprintf(w, "Welcome! Thanks for visiting!\n")
}

func main() {
  log.Fatal(http.ListenAndServe(":8090", http.HandlerFunc(politeGreeting)))
}
```

`HandlerFunc` 是这样一个位于 `http` 包中的巧妙的适配器：

```go
// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
  f(w, r)
}
```

如果你在这篇博文的第一个示例中注意到 `http.HandleFunc`[^2], 它对具有 `HandlerFunc` 签名的函数使用同样的适配器。

就像 `PoliteServer` 一样，`http.ServeMux` 是一个实现了 `http.Handler` 接口的类型。如果愿意的话你可以仔细阅读[完整代码](https://go.googlesource.com/go/+/go1.15.8/src/net/http/server.go)；这是一个大纲：

- `ServeMux` 维护了一个（根据长度）排序的 `{pattern, handler}` 的切片。
- `Handle` 或 `HandleFunc` 向该切片增加新的 handler。
- `ServeHTTP`：

  - （通过查找这个排序好的 handler 对的切片）为请求的 path 找到对应的 handler
  - 调用 handler 的 `ServeHTTP` 方法

因此，mux 可以被看做是一个*转发 handler*；这种模式在 HTTP 服务开发中极为常见，这就是*中间件*。

## `http.Handler` 中间件
由于中间件在不同的上下文，不同的语言以及不同的框架中意味着不同的东西，所以它很难被准确定义。

让我们回到这篇博文开头的流程图上，对它进行一点简化，隐藏 `http` 包所执行的细节：

![](https://github.com/studygolang/gctt-images2/blob/master/20210220-Life-of-an-HTTP-request-in-a-Go-server/http-request-simplified.png?raw=true)

现在，当我们加了中间件的话，流程图看起来是这样的：

![](https://github.com/studygolang/gctt-images2/blob/master/20210220-Life-of-an-HTTP-request-in-a-Go-server/http-request-with-middleware.png?raw=true)

在 Go 语言中，中间件只是另一个 HTTP handler，它包裹了一个其他的 handler。中间件 handler 通过调用 `ListenAndServe` 被注册进来；当调用的时候，它可以执行任意的预处理，调用自身包裹的 handler 然后可以执行任意的后置处理。

我们之前已经见过了一个中间件的例子—— `http.ServeMux`；在这个例子中，预处理是基于请求的 path 来选择正确的用户 handler 来调用。没有后置处理。

再来另一个具体的例子，回到我们的 polite 服务上，新增一些基本的*日志中间件*。这个中间件记录每个请求的具体日志，包括执行了多长时间：

```go
type LoggingMiddleware struct {
  handler http.Handler
}

func (lm *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  start := time.Now()
  lm.handler.ServeHTTP(w, req)
  log.Printf("%s %s %s", req.Method, req.RequestURI, time.Since(start))
}

type PoliteServer struct {
}

func (ms *PoliteServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  fmt.Fprintf(w, "Welcome! Thanks for visiting!\n")
}

func main() {
  ps := &PoliteServer{}
  lm := &LoggingMiddleware{handler: ps}
  log.Fatal(http.ListenAndServe(":8090", lm))
}
```

注意 `LoggingMiddleware` 本身是一个 `http.Handler`，它持有一个用户 handler 作为字段。当 `ListenAndServe` 调用它的 `ServeHTTP` 方法，它做了如下事情：

1. 预处理：在用户的 handler 执行之前记录一个时间戳
2. 使用请求和返回 writer 调用用户 handler
3. 后置处理：记录请求详细日志，包括耗时

中间件最大的优点是可以组合。被中间件所包裹“用户 handler” 也可以是另一个中间件，依次类推。这是一个互相包裹的 `http.Handler` 链。事实上，这在 Go 中是一个常见的模式，来看看 Go 中间件的经典用法。还是我们的日志 polite 服务，这次使用了更有识别度的 Go 中间件实现：

```go
func politeGreeting(w http.ResponseWriter, req *http.Request) {
  fmt.Fprintf(w, "Welcome! Thanks for visiting!\n")
}

func loggingMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    start := time.Now()
    next.ServeHTTP(w, req)
    log.Printf("%s %s %s", req.Method, req.RequestURI, time.Since(start))
  })
}

func main() {
  lm := loggingMiddleware(http.HandlerFunc(politeGreeting))
  log.Fatal(http.ListenAndServe(":8090", lm))
}
```

相对于创建一个带有方法的结构体，`loggingMiddleware` 利用 `http.HandlerFunc` 和闭包使代码更加简洁，同时保留了相同的功能。更重要的是这个例子展示了中间件事实上的标准*签名*：一个函数传入一个 `http.Handler`，有时还有其他状态，之后返回一个不同的 `http.Handler`。返回的 handler 现在应该替换掉传入中间件的那个 handler，之后会“神奇地”执行它原有的功能，并且与中间件的功能包装在一起。

比如。标准库包含了以下的中间件：

```go
func TimeoutHandler(h Handler, dt time.Duration, msg string) Handler
```

如果我们的代码中有 `http.Handler`，像这样包装它：

```go
handler = http.TimeoutHandler(handler, 2 * time.Second, "timed out")
```

创建了一个新版本的 handler，这个版本内置了2秒的超时机制。

中间件的组合可以像下面这样展示：

```go
handler = http.TimeoutHandler(handler, 2 * time.Second, "timed out")
handler = loggingMiddleware(handler)
```

经过这样两行代码之后，`handler` 会带有超时*和日志*功能。你也许会注意到链路长的中间件编写起来会很繁琐；Go 有很多流行的包可以解决这个问题，不过这不在这篇文章的讨论范围内。

顺便一提，虽然 `http` 包在内部使用中间件满足自身需要；具体见这篇博文之前关于 `serverHandler` 适配器的例子。但是它提供了一个清晰的方式以默认行为处理用户 handler 为 nil 的情形（把请求传入默认的 mux）。

希望这样可以让大家明白为什么中间件是一个很吸引人的辅助设计。我们可以专注于我们的“业务逻辑” handler 上，尽管完全正交，我们利用通用的中间件，在许多方面提升我们的 handler。在其他文章中，会进行全面的探讨。

## 并发和 panic 处理
为了结束我们对于 Go HTTP 服务中 HTTP 请求的探索，来介绍另外两个主题：并发和 panic 处理。

首先是*并发*。之前简单提到，每个连接由 `http.Server.Serve` 在一个新的 goroutine 中处理。

这是 Go 的 `net/http` 的一个强大的功能，它利用了 Go 出色的并发性能，使用轻量的 goroutine 使 HTTP handler 保持了一个非常简单的并发模型。一个 handler 阻塞的时候（比如，读取数据库）不需要担心拖慢其他 handler。但是，编写存在共享数据的 handler 的时候需要格外小心。具体细节参考[之前的文章](https://eli.thegreenplace.net/2019/on-concurrency-in-go-http-servers)。

最后，*panic 处理*。一个 HTTP 服务通常是一个长期运行的后台进程。假如在用户提供的请求 handler 中发生了什么糟糕的事情，比如，一些导致运行时 panic 的bug。会导致整个服务崩溃，这可不是什么好事情。为了避免这样的惨剧，你也许会考虑在你服务的 `main` 函数中加上 `recover`，但是并没什么用，原因如下：

1. 当控制返还给 `main` 函数的时候，`ListenAndServe` 已经执行完毕而不会再提供任何服务。
2. 由于每个连接在分开的 goroutine 中处理，当 handler 中发送 panic 的时候，甚至不会影响到 `main` 函数，但是会导致对应进程的崩溃。

为了提供些许的帮助，`net/http` 包（在 `conn.serve` 方法中）内置对每个服务 goroutine 有 recovery。我们可以通过简单的例子来看到它的作用：

```go
func hello(w http.ResponseWriter, req *http.Request) {
  fmt.Fprintf(w, "hello\n")
}

func doPanic(w http.ResponseWriter, req *http.Request) {
  panic("oops")
}

func main() {
  http.HandleFunc("/hello", hello)
  http.HandleFunc("/panic", doPanic)

  http.ListenAndServe(":8090", nil)
}
```

如果我们运行这个服务，并且 `curl` `/panic` 路径，我们可以看到：

```
$ curl localhost:8090/panic
curl: (52) Empty reply from server
```

并且服务会在自身 log 中打印这样的信息：

```
2021/02/16 09:44:31 http: panic serving 127.0.0.1:52908: oops
goroutine 8 [running]:
net/http.(*conn).serve.func1(0xc00010cbe0)
  /usr/local/go/src/net/http/server.go:1801 +0x147
panic(0x654840, 0x6f0b80)
  /usr/local/go/src/runtime/panic.go:975 +0x47a
main.doPanic(0x6fa060, 0xc0001401c0, 0xc000164200)
[... rest of stack dump here ...]
```

不过，这个服务会保持运行并且我们可以继续访问它！

尽管这个内置的保护机制相比服务崩溃要好，许多开发者还是发现了它的局限。这个保护机制只关闭了连接以及在日志中输出错误；通常来说，向客户端返回某种错误响应（比如 code 500 —— 内置错误）和附加详细信息会有用得多。

阅读了这个博文后，再写实现这个功能的中间件应该是很容易的。将它作为练习！我会在之后的博文中介绍这个用例。

[^1]: 与使用默认的 mux 的版本相比，这个版本有充分的理由更喜欢这一版本。默认的 mux 有着一定的安全风险；作为全局实例，它可以被你工程中引入的任何包所修改。一个恶意的包也行会出于邪恶的目的而使用它。
[^2]: 注意：`http.HandleFunc` 和 `http.HandlerFunc` 是具有不同而有相互关联的角色的不同实体。

---
via: https://eli.thegreenplace.net/2021/life-of-an-http-request-in-a-go-server/

作者：[Eli Bendersky](https://eli.thegreenplace.net/pages/about)
译者：[dust347](https://github.com/dust347)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
