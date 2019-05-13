首发于：https://studygolang.com/articles/19813

# 在 Go 中使用 Websockets 和 Socket.IO

> 注 - 本教程是使用 Go 1.9 版和 [googollee/go-socket.io](https://github.com/googollee/go-socket.io) 编写的

Websockets 我觉得非常有趣，在应用程序之间通信中使用标准 RESTful API 方案之外，它为我们提供了一个替代选项。使用套接字，我们可以做到成千上万个不同客户端之间的实时通信，而不必让每分钟数十万个 RESTful API 调用来轰炸我们的服务器。

## 视频教程

[https://www.youtube.com/watch?v=ycgCMOWPgiw](https://www.youtube.com/watch?v=ycgCMOWPgiw)

## 真实生活的例子

用例子明晰一下 Websockets 的重要性。想象有这样一个聊天应用程序，它需要从一个服务器获取所有的最新消息，并将这些新消息全部推送到该服务器。

### REST API 方法

1. 为了实现实时聊天，必须每秒轮询提供新消息的 REST API。
2. 这大约相当于每个客户端每分钟调用 60 次 REST API 。
3. 如果我们构建了一个非常棒的服务，那么越来越多的流量将淹没我们的服务器，因为需要每分钟处理数百万个 REST API 调用。

### 使用套接字

如果我们使用 Websockets 替代 REST API 调用 ：

1. 每个客户端和服务器仅需要维护一个独立连接。
2. 如果拥有 1,000 个客户端，我们只需要维护 1,000 个套接字连接。
3. 只有当有人发布了新消息，我们的服务器才会推送更新给这 1,000 个客户端。

通过这种方法，我们极大减少了访问我们服务器的网络流量。这样就节省了服务器的开销，不必使用非常多的服务器来运行应用。这让我们基本上可以毫不费力地处理数千个客户端。

## 实现 Golang 服务器

在 Go 中实现 Websockets，我们有许多不同的选项。在我的前端生涯中，前端套接字通信最流行的库之一是[socket-io](https://socket.io/)，因此我们使用 Golang 中的同等实现 Go-socket.io 轻松结合它们。

## 安装 Go-socket.io

使用 `go get`  命令安装包：

```
go get github.com/googollee/go-socket.io
```

然后在我们的 Go 程序中导入包，如下所示：

```go
import "github.com/googollee/go-socket.io"
```

## 简单服务器

让我们看一下 `readme.md` 中提供的示例代码。

```go
package main

import (
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
)

func main() {

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {

		log.Println("on connection")

		so.Join("chat")

		so.On("chat message", func(msg string) {
			log.Println("emit:", so.Emit("chat message", msg))
			so.BroadcastTo("chat", "chat message", msg)
		})

		so.On("disconnection", func() {
			log.Println("on disconnect")
		})
	})

	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	http.Handle("/socket.io/", server)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
```

## 详解

在上面的代码示例中，我们在 `main()` 函数中执行所有操作。我们首先通过调用 `socketio.NewServer(nil)`  来定义一个新的 `socketio` 服务器实例，然后我们再定义我们的套接字服务器的连接时的处理和出现错误时的处理。

在 `server.On('connection',...)` 中我们首先记录已经成功连接，然后使用 `so.Join("chat")` 加入 `chat`  房间。

之后，我们会指定连接上的套接字收到  `"chat message"` 事件时怎样处理 。每当我们的服务器收到这种 `"chat message"` 事件时，就会调用 `so.BroadcastTo("chat", "chat message", msg)` 来广播消息给当前连接的所有套接字。这就意味着一个客户端能看到另一个客户端发送的任何消息。

最后，我们定义 `"disconnection"` 时发生什么，在这个例子中，我们只是简单记录一个客户端已断开连接。

## 前端客户端

Ok，我们完成了后端基于 Go 的 WebSocket 服务器，现在该做一个简单的前端应用程序来帮我们测试已经完成的工作。

我们首先在项目目录中创建一个简单的 `index.html` 。

```html
<!DOCTYPE HTML>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, INItial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="ie=edge" />
    <title>Go WebSocket Tutorial</title>
  </head>
  <body>
    <h2>Hello World</h2>

    <script src="https://cdnjs.cloudflare.com/ajax/libs/socket.io/2.1.1/socket.io.js"></script>
    <script>
      const socket = io("http://localhost:5000/socket.io/");
    </script>
  </body>
</html>
```

然后跑 Go 的 Websocket 服务器

```bash
$ go run main.go
2018/06/10 07:54:06 Serving at localhost:5000...
2018/06/10 07:54:15 on connection
2018/06/10 07:54:16 on connection
```

他在 `http://localhost:5000` 上开始运行。使用浏览器访问该 URL，并查看服务器日志输出的新连接。

您现在已经成功构建了一个直接连接到新创建的后端 Websocket 服务器的前端！

## 结论

> 注 - 该项目的完整源代码可以在 Github 上找到： [TutorialEdge / Go](https://github.com/TutorialEdge/Go/tree/master/go-websocket-tutorial)

如果您发现本教程有用或需要任何进一步的帮助，请随时在下面的评论部分告诉我。

---

via: https://tutorialedge.net/golang/golang-websockets-tutorial

作者：[Elliot Forbes](https://twitter.com/Elliot_F)
译者：[yhyddr](https://github.com/yhyddr)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
