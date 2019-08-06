首发于：https://studygolang.com/articles/22430

# 使用 Go 和 ReactJS 构建聊天系统（四）

本节完整代码：[GitHub](https://github.com/watermelo/realtime-chat-go-react/tree/part-4)

> 本文是关于使用 ReactJS 和 Go 构建聊天应用程序的系列文章的第 4 部分。你可以在这里找到第 3 部分 - [前端实现](https://studygolang.com/articles/22429)

这节主要实现处理多个客户端消息的功能，并将收到的消息广播到每个连接的客户端。在本系列的这一部分结束时，我们将：

- 实现了一个池机制，可以有效地跟踪 WebSocket 服务中的连接数。
- 能够将任何收到的消息广播到连接池中的所有连接。
- 当另一个客户端连接或断开连接时，能够通知现有的客户端。

在本课程的这一部分结束时，我们的应用程序看起来像这样：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/chat-system-in-go-and-react-course-series/image_3.png)

## 拆分 Websocket 代码

现在已经完成了必要的基本工作，我们可以继续改进代码库。可以将一些应用程序拆分为子包以便于开发。

现在，理想情况下，你的 `main.go` 文件应该只是 Go 应用程序的入口，它应该相当小，并且可以调用项目中的其他包。

> 注意 - 我们将参考非官方标准的 Go 项目结构布局 - [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

让我们在后端项目目录中创建一个名为 `pkg/` 的新目录。在此期间，我们将要创建另一个名为 `websocket/` 的目录，该目录将包含 `websocket.go` 文件。

我们将把目前在 `main.go` 文件中使用的许多基于 WebSocket 的代码移动到这个新的 `websocket.go` 文件中。

> 注意 - 需要注意的一件事是，当复制函数时，需要将每个函数的第一个字母大写，我们希望这些函数对项目的其余部分可导出。

```go
package websocket

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return ws, err
	}
	return ws, nil
}

func Reader(conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}

func Writer(conn *websocket.Conn) {
	for {
		fmt.Println("Sending")
		messageType, r, err := conn.NextReader()
		if err != nil {
			fmt.Println(err)
			return
		}
		w, err := conn.NextWriter(messageType)
		if err != nil {
			fmt.Println(err)
			return
		}
		if _, err := io.Copy(w, r); err != nil {
			fmt.Println(err)
			return
		}
		if err := w.Close(); err != nil {
			fmt.Println(err)
			return
		}
	}
}
```

现在已经创建了这个新的 `websocket` 包，然后我们想要更新 `main.go` 文件来调用这个包。首先必须在文件顶部的导入列表中添加一个新的导入，然后可以通过使用 `websocket.` 来调用该包中的函数。像这样：

```go
package main

import (
	"fmt"
	"net/http"

	"realtime-chat-go-react/backend/pkg/websocket"
)

func serveWs(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocket Endpoint Hit")
	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	client := &websocket.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.Read()
}

func setupRoutes() {
	pool := websocket.NewPool()
	go pool.Start()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(pool, w, r)
	})
}

func main() {
	fmt.Println("Distributed Chat App v0.01")
	setupRoutes()
	http.ListenAndServe(":8080", nil)
}
```

经过这些修改，我们应该检查一下这些是否破坏了现有的功能。尝试再次运行后端和前端，确保仍然可以发送和接收消息：

```shell
$ cd backend/
$ go run main.go
```

如果成功，我们可以继续扩展代码库来处理多客户端。

到目前为止，目录结构应如下所示：

```plain
- backend/
- - pkg/
- - - websocket/
- - - - websocket.go
- - main.go
- - go.mod
- - go.sum
- frontend/
- ...
```

## 处理多客户端

现在已经完成了基本的操作，我们可以继续改进后端并实现处理多个客户端的功能。

为此，我们需要考虑如何处理与 WebSocket 服务的连接。每当建立新连接时，我们都必须将它们添加到现有连接池中，并确保每次发送消息时，该池中的每个人都会收到该消息。

### 使用 Channels

我们需要开发一个具有大量并发连接的系统。在该连接的持续时间内都会启动新的 `goroutine` 去处理每一个连接。这意味着我们必须关心这些并发 `goroutine` 之间的通信，并确保线程安全。

当进一步实现 `Pool` 结构时，我们必须考虑使用 `sync.Mutex` 来阻塞其他 `goroutine` 同时访问/修改数据，或者我们也可以使用 `channels`。

对于这个项目，我认为最好使用 `channels` 并且以安全的方式在多个并发的 `goroutine` 中进行通信。

> 注意 - 如果想进一步了解 Go 中的 `channels`，可以在这里查看我的其他文章：[Go Channels Tutorial](https://tutorialedge.net/golang/go-channels-tutorial/)

### client.go

我们先创建一个名为 `client.go` 新文件，它将存在于 `pkg/websocket` 目录中，在文件中将定义一个包含以下内容的 `Client` 结构体：

- **ID**：特定连接的唯一可识别字符串
- **Conn**：指向 `websocket.Conn` 的指针
- **Pool**：指向 `Pool` 的指针??

还需要定义一个 `Read()` 方法，该方法将一直监听此 `Client` 的 websocket 连接上发出的新消息。

如果收到新消息，它将把这些消息传递给池的 `Broadcast` channel，该 channel 随后将接收的消息广播到池中的每个客户端。

```go
package websocket

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}

type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		message := Message{Type: messageType, Body: string(p)}
		c.Pool.Broadcast <- message
		fmt.Printf("Message Received: %+v\n", message)
	}
}
```

太棒了，我们已经在代码中定义了客户端，继续实现池。

### Pool 结构体

我们在 `pkg/websocket` 目录下创建一个新文件 `pool.go`。

首先定义一个 `Pool` 结构体，它将包含我们进行并发通信所需的所有 `channels`，以及一个客户端 `map`。

```go
package websocket

import "fmt"

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
	}
}
```

我们需要确保应用程序中只有一个点能够写入 WebSocket 连接，否则将面临并发写入问题。所以，定义了 `Start()` 方法，该方法将一直监听传递给 `Pool` channels 的内容，然后，如果它收到发送给其中一个 channel 的内容，它将采取相应的行动。

- **Register** - 当新客户端连接时，`Register channel` 将向此池中的所有客户端发送 `New User Joined...`
- **Unregister** - 注销用户，在客户端断开连接时通知池
- **Clients** - 客户端的布尔值映射。可以使用布尔值来判断客户端活动/非活动
- **Broadcast** - 一个 channel，当它传递消息时，将遍历池中的所有客户端并通过套接字发送消息。

代码：

```go
func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				fmt.Println(client)
				client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
			}
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
			}
			break
		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			for client, _ := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}
```

### websocket.go

太棒了，我们再对 `websocket.go` 文件进行一些小修改，并删除一些不再需要的函数和方法：

```go
package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return conn, nil
}
```

## 更新 main.go

最后，我们需要更新 `main.go` 文件，在每个连接上创建一个新 `Client`，并使用 `Pool` 注册该客户端：

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/TutorialEdge/realtime-chat-go-react/pkg/websocket"
)

func serveWs(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocket Endpoint Hit")
	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	client := &websocket.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.Read()
}

func setupRoutes() {
	pool := websocket.NewPool()
	go pool.Start()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(pool, w, r)
	})
}

func main() {
	fmt.Println("Distributed Chat App v0.01")
	setupRoutes()
	http.ListenAndServe(":8080", nil)
}
```

## 测试

现在已经做了所有必要的修改，我们应该测试已经完成的工作并确保一切按预期工作。

启动你的后端应用程序：

```shell
$ go run main.go
Distributed Chat App v0.01
```

如果你在几个浏览器中打开 [http://localhost:3000](http://localhost:3000)，可以看到到它们会自动连接到后端 WebSocket 服务，现在我们可以发送和接收来自同一池内的其他客户端的消息！

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/chat-system-in-go-and-react-course-series/image_4.png)

## 总结

在本节中，我们设法实现了一种处理多个客户端的方法，并向连接池中连接的每个人广播消息。

现在开始变得有趣了。我们可以在下一节中添加新功能，例如自定义消息。

> 下一节：Part 5 - [优化前端](https://studygolang.com/articles/22433)

---

via: https://tutorialedge.net/projects/chat-system-in-go-and-react/part-4-handling-multiple-clients/

作者：[Elliot Forbes](https://twitter.com/elliot_f)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
