已发布：https://studygolang.com/articles/12901

# 创建和使用 HTTP 中间层（Making and Using HTTP Middleware）

在构建 Web 应用时，可能需要为许多（甚至全部）HTTP 请求运行一些公共的函数。你可能需要对每个请求进行记录，对每个响应进行压缩，或者在执行一些重要的处理之前检查一下缓存。

组织这种公共函数的一种方法是将其设置为中间层 - 自包含代码，它们在正常应用处理程序之前或之后，独立地处理请求。在 Go 中，使用中间层的常见位置在 ServeMux 和应用程序处理之间，因此通常对 HTTP 请求的控制流程如下所示：

`ServeMux => Middleware Handler => Application Handler`

在这篇文章中，我将解释如何使自定义中间层在这种模式下工作，以及如何使用第三方中间层包的一些具体的示例。

## 基本原则（The Basic Principles）

在 Go 中制作和使用中间层很简单。我们可以设想：

实现我们自己的中间层，使其满足 http.Handler 接口。
构建一个包含我们的中间层处理程序和我们的普通应用处理程序的处理链，我们可以使用它来注册 http.ServeMux。我会解释如何做。

希望你已经熟悉下面构造一个处理程序的方法（如果没有，最好在继续阅读前，看下这个底层的程序）。

```go
func messageHandler(message string) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(message)
  })
}
```

在这个处理程序中，我们将我们的逻辑（一个简单的 `w.Write`）放在匿名函数中，并封装 `message` 变量以形成闭包。然后我们使用 http.HandlerFunc 适配器并将其返回，将此闭包转换为处理程序。

我们可以使用这种相同的方法来创建一系列的处理程序。我们将链中的下一个处理程序作为变量传递给闭包（而不是像上面那样），然后通过调用 ServeHTTP() 方法将控制权转移给下一个处理程序。

这为我们提供了构建中间层的完整的模式：

```go
func exampleMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Our middleware logic goes here...
    next.ServeHTTP(w, r)
  })
}
```

你会注意到这个中间层函数有一个  `func (http.Handler) http.Handler` 签名。它接受一个处理程序作为参数并返回一个处理程序。这是有用的，原因有两个：

因为它返回一个处理程序，我们可以直接使用 net/http 软件包提供的标准 ServeMux 注册中间层函数。
通过将中间层函数嵌套在一起，我们可以创建一个任意长的处理程序链。例如：

`http.Handle("/", middlewareOne(middlewareTwo(finalHandler)))`

## 控制流程说明（Illustrating the Flow of Control）

让我们看一个简单的例子，它带有一些只需将日志消息写入标准输出的中间层函数：

```
File: main.go
```

```go
package main

import (
  "log"
  "net/http"
)

func middlewareOne(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    log.Println("Executing middlewareOne")
    next.ServeHTTP(w, r)
    log.Println("Executing middlewareOne again")
  })
}

func middlewareTwo(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    log.Println("Executing middlewareTwo")
    if r.URL.Path != "/" {
      return
    }
    next.ServeHTTP(w, r)
    log.Println("Executing middlewareTwo again")
  })
}

func final(w http.ResponseWriter, r *http.Request) {
  log.Println("Executing finalHandler")
  w.Write([]byte("OK"))
}

func main() {
  finalHandler := http.HandlerFunc(final)

  http.Handle("/", middlewareOne(middlewareTwo(finalHandler)))
  http.ListenAndServe(":3000", nil)
}
```

运行这个应用程序并向 `http://localhost:3000` 发出请求。你应该会得到这样的日志输出：

```
$ go run main.go
2014/10/13 20:27:36 Executing middlewareOne
2014/10/13 20:27:36 Executing middlewareTwo
2014/10/13 20:27:36 Executing finalHandler
2014/10/13 20:27:36 Executing middlewareTwo again
2014/10/13 20:27:36 Executing middlewareOne again
```

很明显，可以看到如何通过处理程序链按照嵌套顺序传递控制权，然后再以相反的方向返回。

任何时候，我们都可以通过在中间层处理程序中返回，来停止链传递的控件。

在上面的例子中，我在中间层中包含了一个条件返回函数。通过访问 `http://localhost:3000/foo` 再次尝试，并检查日志 - 你会发现此次请求不会通过中间层进一步传递到备份链。

## 通过一个合适的例子来了解如何？（Understood. How About a Proper Example?）

好了。假设我们正在构建一个处理正文中包含 XML 请求的服务。我们想要创建一些中间层，它们 a）检查请求体是否存在，b）嗅探以确保它是 XML（格式）。如果其中任何一项检查失败，我们希望我们的中间层写入错误消息并停止将请求传递给我们的应用处理程序。

```
File: main.go
```

```go
package main

import (
  "bytes"
  "net/http"
)

func enforceXMLHandler(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Check for a request body
    if r.ContentLength == 0 {
      http.Error(w, http.StatusText(400), 400)
      return
    }
    // Check its MIME type
    buf := new(bytes.Buffer)
    buf.ReadFrom(r.Body)
    if http.DetectContentType(buf.Bytes()) != "text/xml; charset=utf-8" {
      http.Error(w, http.StatusText(415), 415)
      return
    }
    next.ServeHTTP(w, r)
  })
}

func main() {
  finalHandler := http.HandlerFunc(final)

  http.Handle("/", enforceXMLHandler(finalHandler))
  http.ListenAndServe(":3000", nil)
}

func final(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("OK"))
}
```

这看起来不错。我们通过创建一个简单的 XML 文件来测试它：

```
$ cat > books.xml
<?xml version="1.0"?>
<books>
  <book>
    <author>H. G. Wells</author>
    <title>The Time Machine</title>
    <price>8.50</price>
  </book>
</books>
```

并使用 curl 命令发出一些请求：

```
$ curl -i localhost:3000
HTTP/1.1 400 Bad Request
Content-Type: text/plain; charset=utf-8
Content-Length: 12

Bad Request
$ curl -i -d "This is not XML" localhost:3000
HTTP/1.1 415 Unsupported Media Type
Content-Type: text/plain; charset=utf-8
Content-Length: 23

Unsupported Media Type
$ curl -i -d @books.xml localhost:3000
HTTP/1.1 200 OK
Date: Fri, 17 Oct 2014 13:42:10 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

OK
```

## 使用第三方中间层（Using Third-Party Middleware）

基本上你想直接使用第三方软件包而不是自己写中间层。我们将在这里看到一对（第三方软件包）：[goji/httpauth](http://elithrar.github.io/article/httpauth-basic-auth-for-go/) 和 Gorilla 的 [LoggingHandler](http://www.gorillatoolkit.org/pkg/handlers#LoggingHandler)。

goji/httpauth 包提供了 HTTP 基本的认证功能。它有一个 [SimpleBasicAuth](https://godoc.org/github.com/goji/httpauth#SimpleBasicAuth) helper，它返回一个带有签名的 `func (http.Handler) http.Handler` 函数。这意味着我们可以像我们定制的中间层一样（的方式）使用它。

```
$ go get github.com/goji/httpauth
```

```
File: main.go
```

```go
package main

import (
  "github.com/goji/httpauth"
  "net/http"
)

func main() {
  finalHandler := http.HandlerFunc(final)
  authHandler := httpauth.SimpleBasicAuth("username", "password")

  http.Handle("/", authHandler(finalHandler))
  http.ListenAndServe(":3000", nil)
}

func final(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("OK"))
}
```

如果你运行这个例子，你应该得到你对有效和无效凭证所期望的回应：

```
$ curl -i username:password@localhost:3000
HTTP/1.1 200 OK
Content-Length: 2
Content-Type: text/plain; charset=utf-8

OK
$ curl -i username:wrongpassword@localhost:3000
HTTP/1.1 401 Unauthorized
Content-Type: text/plain; charset=utf-8
Www-Authenticate: Basic realm=""Restricted""
Content-Length: 13

Unauthorized
```

Gorilla 的 LoggingHandler - 它记录了 [Apache 风格的日志](http://httpd.apache.org/docs/1.3/logs.html#common) - 有点不一样。

它使用签名 `func(out io.Writer, h http.Handler) http.Handler`，所以它不仅需要下一个处理程序，还需要将日志写入的 io.Writer。

以下是一个简单的例子，我们将日志写入 `server.log` 文件：

```bash
go get github.com/gorilla/handlers
```

```
File: main.go
```

```go
package main

import (
  "github.com/gorilla/handlers"
  "net/http"
  "os"
)

func main() {
  finalHandler := http.HandlerFunc(final)

  logFile, err := os.OpenFile("server.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
  if err != nil {
    panic(err)
  }

  http.Handle("/", handlers.LoggingHandler(logFile, finalHandler))
  http.ListenAndServe(":3000", nil)
}

func final(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("OK"))
}
```

在这种小例子中，我们的代码是非常清晰的。但是如果我们想将 LoggingHandler 用作更大的中间层链的一部分，会发生什么？我们可以很容易地得到一个看起来像这样的声明...

`http.Handle("/", handlers.LoggingHandler(logFile, authHandler(enforceXMLHandler(finalHandler))))`

... 那让我的头疼！

一种已经知道的方法是通过创建一个构造函数（让我们称之为 myLoggingHandler）和签名 `func (http.Handler) http.Handler`。这将使我们能够与其他中间层更加简洁地嵌套在一起：

```go
func myLoggingHandler(h http.Handler) http.Handler {
  logFile, err := os.OpenFile("server.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
  if err != nil {
    panic(err)
  }
  return handlers.LoggingHandler(logFile, h)
}

func main() {
  finalHandler := http.HandlerFunc(final)

  http.Handle("/", myLoggingHandler(finalHandler))
  http.ListenAndServe(":3000", nil)
}
```

如果你运行这个应用程序并发送一些请求，你的 server.log 文件应该是这样的：

```
$ cat server.log
127.0.0.1 - - [21/Oct/2014:18:56:43 +0100] "GET / HTTP/1.1" 200 2
127.0.0.1 - - [21/Oct/2014:18:56:36 +0100] "POST / HTTP/1.1" 200 2
127.0.0.1 - - [21/Oct/2014:18:56:43 +0100] "PUT / HTTP/1.1" 200 2
```

如果你有兴趣，可以参考这篇文章中的三个中间层处理程序。

附注：请注意，Gorilla LoggingHandler 正在记录日志中的响应状态（`200`）和响应长度（`2`）。这很有趣。上游的日志记录中间层是如何知道我们的应用处理程序编写的响应的？

它通过定义自己的 `responseLogger` 类来包装 `http.ResponseWriter`，并创建自定义的 `reponseLogger.Write()` 和 `reponseLogger.WriteHeader()` 方法。这些方法不仅可以编写响应，还可以存储大小和状态供以后检查。Gorilla 的 LoggingHandler 将 `reponseLogger` 传递给链中的下一个处理程序，而不是普通的 `http.ResponseWriter`。

## 附加工具（Additional Tools）

[由 Justinas Stankevičius 编写的 Alice](https://github.com/justinas/alice) 是一个非常聪明并且轻量级的包，它为连接中间层处理程序提供了一些语法糖。在最基础的方面，Alice 允许你重写这个：

`http.Handle("/", myLoggingHandler(authHandler(enforceXMLHandler(finalHandler))))`

为这个：

`http.Handle("/", alice.New(myLoggingHandler, authHandler, enforceXMLHandler).Then(finalHandler))`

至少在我看来，这些代码一眼就能看清楚这一点。但是，Alice 的真正好处是它可以让你指定一个处理程序链并将其重复用于多个路由。像这样：

```go
stdChain := alice.New(myLoggingHandler, authHandler, enforceXMLHandler)

http.Handle("/foo", stdChain.Then(fooHandler))
http.Handle("/bar", stdChain.Then(barHandler))
```

---

via: http://www.alexedwards.net/blog/making-and-using-middleware

作者：[TIAGO KATCIPIS](https://katcipis.github.io/)
译者：[gogeof](https://github.com/gogeof)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
