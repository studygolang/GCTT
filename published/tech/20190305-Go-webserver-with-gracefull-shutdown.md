首发于：https://studygolang.com/articles/19401

# 优雅关闭的 Go Web 服务器

在这篇博文里我想要给你们展示下，如何创建一个可以优雅关闭的 Go HTTP Web 服务器。通过这个方法可以让服务器在它真正关闭之前清理某些资源，( 例如 ) 想象下完成数据库事务或者一些其他长时间的操作。我们将会用到在我[关于并发的博文](https://marcofranssen.nl/concurrency-in-go/) 学习到的东西。所以，期待看到 channel 和 Goroutine 作为解决方法的一部分吧。

当我建立新的 http 服务器，我通常通过使用命令行标志提供端口号启动。特别是当多个微服务的情况下，这会十分顺手，你可以陆续启动你的 Web 服务器，测试他们之间的集成。让我们看一下在启动服务器的时候，如何从命令行提供 `listen-address`，包括一个合理的默认值。

```go
// main.go
package main

import (
  "flag"
  "log"
  "os"
)

var (
  listenAddr string
)

func main() {
  flag.StringVar(&listenAddr, "listen-addr", ":5000", "server listen address")
  flag.Parse()

  logger := log.New(os.Stdout, "http: ", log.LstdFlags)

  logger.Println("Server is ready to handle requests at", listenAddr)
}
```

程序会读取 `-listen-addr` 的命令行选项作为我们的变量 `listenAddr` 的值。如果没有值提供则使用 `:5000` 作为默认值。文本 `server listen address` 则会被用作帮助文档的描述。所以你可以使用**flag**包来管理所有想要的命令行选项。

```shell
$ go build .
$ ./gracefull-webserver
Server is ready to handle requests at :5000

$ ./gracefull-webserver --help
Usage of gracefull-webserver.exe:
  -listen-addr string
        server listen address (default ":5000")

$ ./gracefull-webserver --listen-addr :6000
Server is ready to handle requests at :6000
```

现在一起看下 Web 服务器的基本设置，在下面例子中我们创建了监听 `/` 的路由，它会返回 http 状态 200。

```go
router := http.NewServeMux() // 这里你也可以用第三方的包来创建路由
// 注册你的路由
router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
})

server := &http.Server{
  Addr:         listenAddr,
  Handler:      router,
  ErrorLog:     logger,
  ReadTimeout:  5 * time.Second,
  WriteTimeout: 10 * time.Second,
  IdleTimeout:  15 * time.Second,
}

if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
  logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
}
logger.Println("Server stopped")
```

在最后的 if 语句中我们启动了我们的 Web 服务器并检查了所有错误。例如，指定端口可能已被使用，因而不能在该端口上启动我们的 Web 服务器。如果发生的话，它会记录错误并停止程序。**请注意**：在这一步你需要导入 `net/http` 包让代码运行起来。

现在当我们运行咱们的应用，你会看到它阻塞在 `server.ListenAndServe()` 这一行，直到你杀掉进程。

```shell
$ ./gracefull-webserver
Server is ready to handle requests at :5000
CTRL+C
Server stopped
```

到目前为止一切顺利，一切进展良好。然而这不会优雅关闭服务器以及任何可能的与 Web 服务器打开的连接。想象下某人在你退出服务器的时刻接收到一个服务器的响应，那么同样的这个响应会被立刻终止掉。为了允许服务器可以完成任意打开的请求，我们可以加入一些代码在最大超时内去优雅地处理进行中的工作。我们也会改动服务器，让它不保持任何完成工作的连接存活。为了做到这种效果，我们会加更多的代码到一个单独运行的 Goroutine 中，让它拦截关闭应用的信号，然后在那做一些优雅的处理。

要做的第一件事是添加一些 channel，通过他们我们可以在 2 个 Goroutine 之间进行通信。如果这是你第一次在 GO 里接触协程 (goroutine) 和 channel，你可能想要先看一下我的有关[Go 并发的博文](https://marcofranssen.nl/concurrency-in-go/)。

首先我们会定义一个 channel 去通知主协程，优雅关闭已经完成了。我们也会增加一个 channel 来等待任意从操作系统而来的关闭应用的信号。

在单独的 Goroutine 里，我们会等待任意到 `quit` channel 的中断信号。我们要做的第一件事，是在终端打印一条消息告诉用户服务器正在关闭。通过使用上下文我们给服务器 30 秒时间进行优雅关闭。使用 `server.SetKeepAlivesEnabled(false)` 通知 Web 服务器不保持任何存在的连接存活，这基本保证了我们的优雅关闭行为，而不是在一个消费者面前仅仅把门关上。

```go
done := make(chan bool, 1)
quit := make(chan os.Signal, 1)

signal.Notify(quit, os.Interrupt)

go func() {
  <-quit
  logger.Println("Server is shutting down...")

  ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
  defer cancel()

  server.SetKeepAlivesEnabled(false)
  if err := server.Shutdown(ctx); err != nil {
    logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
  }
  close(done)
}()

logger.Println("Server is ready to handle requests at", listenAddr)
if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
  logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
}

<-done
logger.Println("Server stopped")
```

一旦关闭完成，我们通过 `done` channel 来通知主协程我们完成优雅关闭了。这让程序继续执行最后一行 `logger.Println`。输出关闭顺序全部完成并关闭程序。

## TLDR

下面你可以看到我们在这篇博文中讨论的所有内容的完整示例，它们结合在一个完全可用的样板文件中。

```go
main.go
package main

import (
  "context"
  "flag"
  "log"
  "net/http"
  "os"
  "os/signal"
  "time"
)

var (
  listenAddr string
)

func main() {
  flag.StringVar(&listenAddr, "listen-addr", ":5000", "server listen address")
  flag.Parse()

  logger := log.New(os.Stdout, "http: ", log.LstdFlags)

  done := make(chan bool, 1)
  quit := make(chan os.Signal, 1)

  signal.Notify(quit, os.Interrupt)

  server := newWebserver(logger)
  go gracefullShutdown(server, logger, quit, done)

  logger.Println("Server is ready to handle requests at", listenAddr)
  if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
  }

  <-done
  logger.Println("Server stopped")
}

func gracefullShutdown(server *http.Server, logger *log.Logger, quit <-chan os.Signal, done chan<- bool) {
  <-quit
  logger.Println("Server is shutting down...")

  ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
  defer cancel()

  server.SetKeepAlivesEnabled(false)
  if err := server.Shutdown(ctx); err != nil {
    logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
  }
  close(done)
}

func newWebserver(logger *log.Logger) *http.Server {
  router := http.NewServeMux()
  router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
  })

  return &http.Server{
    Addr:         listenAddr,
    Handler:      router,
    ErrorLog:     logger,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  15 * time.Second,
  }
}
```

正如你所能看到的，我做了两处小型重构，我把服务器创建和优雅关闭移到了他们各自的方法中。对于细心的读者可能也会注意到，我已经控制在函数里读写的 channel 仅仅在函数范围内，这为你提供了少量的编译时间优势，也防止你错误地使用 channel。最后但也挺重要的，你可以在这里下载样板文件，作为你自己的 Web 服务器的起点。

期待你的反馈。请在社交媒体上和你的朋友、同事分享这篇博客吧。

---

via: https://marcofranssen.nl/go-webserver-with-gracefull-shutdown/

作者：[Marco Franssen](https://marcofranssen.nl/about)
译者：[LSivan](https://github.com/LSivan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
