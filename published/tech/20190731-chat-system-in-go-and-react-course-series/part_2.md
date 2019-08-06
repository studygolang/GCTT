首发于：https://studygolang.com/articles/22426

# 使用 Go 和 ReactJS 构建聊天系统（二）：gorilla/websocket 包提供的 WebSockets

本节完整代码：[GitHub](https://github.com/watermelo/realtime-chat-go-react/tree/part-1-and-2)

> 本文是使用 ReactJS 和 Go 来构建聊天应用程序的系列文章的第 2 部分。你可以在这里找到第 1 部分 - [初始化设置](https://studygolang.com/articles/22423)

现在我们已经建立好了基本的前端和后端，现在需要来完善一些功能了。

在本节中，我们将实现一个基于 WebSocket 的服务器。

在该系列教程结束时，我们将有一个可以于后端双向通信的前端应用程序。

## 服务

我们可以使用 `github.com/gorilla/websocket` 包来设置 WebSocket 服务以及处理 WebSocket 连接的读写操作。

这需要在我们的 `backend/` 目录中运行此命令来安装它：

```shell
$ go get github.com/gorilla/websocket
```

一旦我们成功安装了这个包，我们就可以开始构建我们的 Web 服务了。我们首先创建一个非常简单的 `net/http` 服务：

```go
package main

import (
	"fmt"
	"net/http"
)

func setupRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Simple Server")
	})
}

func main() {
	setupRoutes()
	http.ListenAndServe(":8080", nil)
}
```

可以通过调用 `go run main.go` 来启动服务，该服务将监听 [http://localhost:8080](http://localhost:8080) 。如果用浏览器打开此连接，可以看到输出 `Simple Server`。

## WebSocket 协议

在开始写代码之前，我们需要了解一下理论。

WebSockets 可以通过 TCP 连接进行双工通信。这让我们可以通过单个 TCP 套接字来发送和监听消息，从而避免通过轮询 Web 服务器去通信，每次轮询操作都会执行 TCP 握手过程。

WebSockets 大大减少了应用程序所需的网络带宽，并且使得我们在单个服务器实例上维护大量客户端。

## 连接

WebSockets 肯定有一些值得考虑的缺点。比如一旦引入状态，在跨多个实例扩展应用程序的时候就变得更加复杂。

在这种场景下需要考虑更多的情况，例如将状态存储在消息代理中，或者存储在数据库/内存缓存中。

## 实现

在实现 WebSocket 服务时，我们需要创建一个端点，然后将该端点的连接从标准的 HTTP 升级到 WebSocket。

值得庆幸的是，`gorilla/websocket` 包提供了我们所需的功能，可以轻松地将 HTTP 连接升级到 WebSocket 连接。

> 注意 - 你可以查看官方 WebSocket 协议的更多信息：[RFC-6455](https://tools.ietf.org/html/rfc6455)

## 创建 WebSocket 服务端

现在已经了解了理论，来看看如何去实践。我们创建一个新的端点 `/ws`，我们将从标准的 `http` 端点转换为 `ws` 端点。

此端点将执行 3 项操作，它将检查传入的 HTTP 请求，然后返回 `true` 以打开我们的端点到客户端。然后，我们使用定义的 `upgrader` 升级为 WebSocket 连接。

最后，我们将开始监听传入的消息，然后将它们打印出来并将它们传回相同的连接。这可以让我们验证前端连接并从新创建的 WebSocket 端点来发送/接收消息：

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// 我们需要定义一个 Upgrader
// 它需要定义 ReadBufferSize 和 WriteBufferSize
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// 可以用来检查连接的来源
	// 这将允许从我们的 React 服务向这里发出请求。
	// 现在，我们可以不需要检查并运行任何连接
	CheckOrigin: func(r *http.Request) bool { return true },
}

// 定义一个 reader 用来监听往 WS 发送的新消息
func reader(conn *websocket.Conn) {
	for {
		// 读消息
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// 打印消息
		fmt.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}

// 定义 WebSocket 服务处理函数
func serveWs(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Host)

	// 将连接更新为 WebSocket 连接
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	// 一直监听 WebSocket 连接上传来的新消息
	reader(ws)
}

func setupRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Simple Server")
	})

	// 将 `/ws` 端点交给 `serveWs` 函数处理
	http.HandleFunc("/ws", serveWs)
}

func main() {
	fmt.Println("Chat App v0.01")
	setupRoutes()
	http.ListenAndServe(":8080", nil)
}
```

如果没有问题的话，我们使用 `go run main.go` 来启动服务。

## 客户端

现在已经设置好了服务，我们需要一些能够与之交互的东西。这是我们的 ReactJS 前端发挥作用的地方。

我们先尽量让客户端保持简单，并定义一个 `api/index.js` 文件，它将包含 WebSocket 连接的代码。

```js
// api/index.js
var socket = new WebSocket("ws://localhost:8080/ws");

let connect = () => {
	console.log("Attempting Connection...");

	socket.onopen = () => {
		console.log("Successfully Connected");
	};

	socket.onmessage = msg => {
		console.log(msg);
	};

	socket.onclose = event => {
		console.log("Socket Closed Connection: ", event);
	};

	socket.onerror = error => {
		console.log("Socket Error: ", error);
	};
};

let sendMsg = msg => {
	console.log("sending msg: ", msg);
	socket.send(msg);
};

export { connect, sendMsg };
```

因此，在上面的代码中，我们定义了我们随后导出的 2 个函数。分别是 `connect()` 和 `sendMsg(msg)`。

第一个函数，`connect()` 函数，连接 WebSocket 端点，并监听例如与 `onopen` 成功连接之类的事件。如果它发现任何问题，例如连接关闭的套接字或错误，它会将这些问题打印到浏览器控制台。

第二个函数，`sendMsg(msg)` 函数，允许我们使用 `socket.send()` 通过 WebSocket 连接从前端发送消息到后端。

现在我们在 React 项目中更新 `App.js` 文件，添加对 `connect()` 的调用并创建一个触发 `sendMsg()` 函数的 `<button />` 元素。

```js
// App.js
import React, { Component } from "react";
import "./App.css";
import { connect, sendMsg } from "./api";

class App extends Component {
	constructor(props) {
		super(props);
		connect();
	}

	send() {
		console.log("hello");
		sendMsg("hello");
	}

	render() {
		return (
			<div className="App">
				<button onClick={this.send}>Hit</button>
			</div>
		);
	}
}

export default App;
```

使用 `npm start` 成功编译后，我们可以在浏览器中看到一个按钮，如果打开浏览器控制台，还可以看到成功连接的 WebSocket 服务运行在 [http://localhost:8080](http://localhost:8080)。

> 问题 - 单击此按钮会发生什么？你在浏览器的控制台和后端的控制台中看到了什么输出？

## 总结

结束了本系列的第 2 部分。我们已经能够创建一个非常简单的 WebSocket 服务，它可以回显发送给它的任何消息。

这是开发应用程序的关键一步，现在我们已经启动并运行了基本框架，我们可以开始考虑实现基本的聊天功能并让这个程序变得更有用！

> 下一节：Part 3 - [前端实现](https://studygolang.com/articles/22429)

---

via: https://tutorialedge.net/projects/chat-system-in-go-and-react/part-2-simple-communication/

作者：[Elliot Forbes](https://twitter.com/elliot_f)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
