已发布：https://studygolang.com/articles/12488

# Golang 下的微服务 - 第 5 部分 - Go Micro 的事件代理

在本系列的[前一部分中](https://studygolang.com/articles/12485)，我们谈到了用户认证和 JWT。在这一部分中，我们将快速浏览 go-micro 的代理功能。

正如前面的文章提到的，go-micro 是一个可插拔的框架，它连接了许多不同的常用技术。如果你看看[插件仓库](https://github.com/micro/go-plugins)，你会看到它支持多少插件。

在我们的例子中，我们将使用 NATS 代理插件。

## 基于事件驱动的架构

[事件驱动的架构](https://en.wikipedia.org/wiki/Event-driven_architecture)是一个非常简单的概念。我们通常认为好的架构是要解耦的，一个服务不应该与其他服务耦合或者感知到其他服务。当我们使用诸如 `gRPC` 协议时，在某些情况下是正确的，我们以向 `go.srv.user-service` 发布请求为例。其中就使用了服务发现的方式来查找该服务的实际位置。 尽管这并不直接将我们与实现耦合，但它确实将该服务耦合到了其他名为 `go.srv.user-service` 的服务，因此它不是完全的解耦，因为它直接与其他服务进行交互。

那么什么让事件驱动架构真正的解耦呢？为了理解这一点，我们首先看看发布和订阅事件的过程。服务 a 完成了一项任务 x，然后向系统发布一个事件 `x 刚刚发生了`。服务并不需要知道或者关心谁在监听这个事件，或者该事件正在发生什么影响。这些事情留给了监听事件的客户端。如果你期待 n 个服务对某个事件采取行动，那么也很容易。例如，你想 12 个不同的服务针对使用 `gRPC` 创建新用户采取行动，可能需要在用户服务中实例化 12 个客户端。而借助事件发布订阅或事件驱动架构，你的服务就不需要关心这些。

现在，客户端服务只需要简单的监听事件。这意味着，你需要中间的介质来接收这些事件，并通知订阅了事件的客户端。

这篇文章中，我们将在每次创建用户时创建一个事件，并且将创建一个用于发送电子邮件的新服务。我们不会真的去实现发邮件的功能，只是模拟它。

## 代码

首先，我们需要将 NATS 代理插件集成到我们的用户服务中：

```go
// shippy-user-service/main.go
func main() {
	... 
	// Init will parse the command line flags.
	srv.Init()

	// Get instance of the broker using our defaults
	pubsub := srv.Server().Options().Broker

	// Register handler
	pb.RegisterUserServiceHandler(srv.Server(), &service{repo, tokenService, pubsub})
	...
}
```

现在让我们在创建新用户时发布事件（[请参阅此处的完整更改](https://github.com/EwanValentine/shippy-user-service/tree/tutorial-5)）

```go
// shippy-user-service/handler.go
const topic = "user.created"

type service struct {
	repo         Repository
	tokenService Authable
	PubSub       broker.Broker
}
...
func (srv *service) Create(ctx context.Context, req *pb.User, res *pb.Response) error {

	// Generates a hashed version of our password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	req.Password = string(hashedPass)
	if err := srv.repo.Create(req); err != nil {
		return err
	}
	res.User = req
	if err := srv.publishEvent(req); err != nil {
		return err
	}
	return nil
}

func (srv *service) publishEvent(user *pb.User) error {
	// Marshal to JSON string
	body, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// Create a broker message
	msg := &broker.Message{
		Header: map[string]string{
			"id": user.Id,
		},
		Body: body,
	}

	// Publish message to broker
	if err := srv.PubSub.Publish(topic, msg); err != nil {
		log.Printf("[pub] failed: %v", err)
	}

	return nil
}
...
```

确保你正在运行 Postgres，然后让我们运行这个服务：

```shell
$ docker run -d -p 5432:5432 postgres
$ make build
$ make run
```

现在我们创建我们的电子邮件服务。 我为此创建了一个[新的仓库](https://github.com/EwanValentine/shippy-email-service)：

```go
// shippy-email-service
package main

import (
	"encoding/json"
	"log"

	pb "github.com/EwanValentine/shippy-user-service/proto/user"
	micro "github.com/micro/go-micro"
	"github.com/micro/go-micro/broker"
	_ "github.com/micro/go-plugins/broker/nats"
)

const topic = "user.created"

func main() {
	srv := micro.NewService(
		micro.Name("go.micro.srv.email"),
		micro.Version("latest"),
	)

	srv.Init()

	// Get the broker instance using our environment variables
	pubsub := srv.Server().Options().Broker
	if err := pubsub.Connect(); err != nil {
		log.Fatal(err)
	}

	// Subscribe to messages on the broker
	_, err := pubsub.Subscribe(topic, func(p broker.Publication) error {
		var user *pb.User
		if err := json.Unmarshal(p.Message().Body, &user); err != nil {
			return err
		}
		log.Println(user)
		go sendEmail(user)
		return nil
	})

	if err != nil {
		log.Println(err)
	}

	// Run the server
	if err := srv.Run(); err != nil {
		log.Println(err)
	}
}

func sendEmail(user *pb.User) error {
	log.Println("Sending email to:", user.Name)
	return nil
}
```

在运行之前，我们需要启动 [NATS](https://nats.io/)...

```
$ docker run -d -p 4222:4222 nats
```

另外，我想快速解释一下 go-micro 的一部分，我觉得这对于理解它作为框架如何工作很重要。你会注意到：

```go
srv.Init()
pubsub := srv.Server().Options().Broker
```

让我们来快速浏览一下。当我们用 go-micro 创建服务时，`srv.Init()` 会自动去查找所有的配置，例如所有配置的插件、环境变量或命令行选项。它将会将这些集成实例化为服务的一部分。为了使用这些实例，我们需要将它们从服务中提取出来。在 `srv.Server().Options()` 中，你还可以找到 Transport (go-micro 框架的一个核心组件，传输是服务之间的同步请求/响应通信的接口) 和 Registry (go-micro 框架的一个核心组件，叫注册表，提供了一个服务发现机制来将名称解析为地址)。

在我们的例子中，会用到 `GO_MICRO_BROKER` 环境变量，会用到 `NATS` 代理插件，并创建一个该插件的实例，准备好我们连接和使用。

如果你正在创建一个命令行工具，你可以使用 `cmd.Init()`，确保你导入了 `github.com/micro/go-micro/cmd`。这会产生同样的影响。

现在构建并运行此服务：`$ make build && make run`，确保你也在运行用户服务。然后转到 `shippy-user-cli` 项目，并运行 `$ make run`，看我们的电子邮件服务输出。你应该看到类似... `2017/12/26 23:57:23 Sending email to: Ewan Valentine`

就是这样！这是一个简单的例子，因为我们的电子邮件服务隐式地收听单个 `user.created` 事件，但希望你能看到这种方法如何让你编写解耦的服务。

值得一提的是，使用 JSON over NATS 会比 gRPC 带来更高的性能开销，因为我们已经回到串行化json字符串的领域。但是，对于某些使用情况，这是完全可以接受的。 NATS 非常高效，非常适合消息最多交付一次的事件（fire and forget 有消息最多交付一次的意思，这个[链接](http://www.enterpriseintegrationpatterns.com/patterns/conversation/FireAndForget.html)可以帮助做更深入的理解）。

Go-micro 还支持一些最广泛使用的队列 / pubsub 技术供你使用。[你可以在这里看到它们的列表](https://github.com/micro/go-plugins/tree/master/broker)。你不需要改变你的实现因为 go-micro 为你提供了抽象。你只需要将环境变量从 `MICRO_BROKER=nats` 更改为 `MICRO_BROKER=googlepubsub`，然后将 main.go 的导入从 `_ "github.com/micro/go-plugins/broker/nats"` 更改为 `_ "github.com/micro/go-plugins/broker/googlepubsub"`。

如果你不使用 go-micro，那么有一个 [NATS go 库](https://github.com/nats-io/go-nats)（NATS 是用 go 写的，所以对 Go 的支持非常稳固）。

发布一个事件：

```go
nc, _ := nats.Connect(nats.DefaultURL)

// Simple Publisher
nc.Publish("user.created", userJsonString)
```

订阅一个事件：

```go
// Simple Async Subscriber
nc.Subscribe("user.created", func(m *nats.Msg) {
	user := convertUserString(m.Data)
	go sendEmail(user)
})
```

我之前提到过，在使用第三方消息代理（如 NATS）时，会失去对 protobuf 的使用。这是一种耻辱，因为我们失去了使用二进制流进行通信的能力，这当然比串行化的 JSON 字符串的开销要低得多。 但是，像大多数人所关心的那样，go-micro 也可以解决这个问题。

内置 go-micro 是 pubsub 层，位于代理层之上，但不需要第三方代理（如 NATS）。 但是这个功能真正棒的部分在于它利用了 protobuf 的定义。 所以我们回到了低延迟二进制流的领域。 因此，让我们更新我们的用户服务，用 go-micro 的 pubsub 替换现有的 NATS 代理：

```go
// shippy-user-service/main.go
func main() {
	...
	publisher := micro.NewPublisher("user.created", srv.Client())

	// Register handler
	pb.RegisterUserServiceHandler(srv.Server(), &service{repo, tokenService, publisher})
	...
}
```

```go
// shippy-user-service/handler.go
func (srv *service) Create(ctx context.Context, req *pb.User, res *pb.Response) error {

	// Generates a hashed version of our password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	req.Password = string(hashedPass)

	// Here's our new publisher code, much simpler
	if err := srv.repo.Create(req); err != nil {
		return err
	}
	res.User = req
	if err := srv.Publisher.Publish(ctx, req); err != nil {
		return err
	}
	return nil
}
```

现在我们的邮件服务是这样的：

```go
// shippy-email-service
const topic = "user.created"

type Subscriber struct{}

func (sub *Subscriber) Process(ctx context.Context, user *pb.User) error {
	log.Println("Picked up a new message")
	log.Println("Sending email to:", user.Name)
	return nil
}

func main() {
	...
	micro.RegisterSubscriber(topic, srv.Server(), new(Subscriber))
	...
}
```

现在我们在我们的服务中使用我们的底层 User protobuf 定义，通过 gRPC，并且不使用第三方代理。太棒了！

这是一个包装！ 接下来的教程我们将着眼于为我们的服务创建一个用户界面，并研究 Web 客户端如何开始与我们的服务进行交互。

本文中的任何错误、反馈，或任何您会发现有用的东西，请给我发[电子邮件](ewan.valentine89@gmail.com)。

---

via：[Microservices in Golang - Part 5 - Event brokering with Go Micro](https://ewanvalentine.io/microservices-in-golang-part-5/)

作者：[André Carvalho](https://ewanvalentine.io/)
译者：[shniu](https://github.com/shniu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
