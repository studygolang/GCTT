
# Golang下的微服务 - Part5 - Go Micro的事件代理

在本系列的前一部分中，我们谈到了用户认证和JWT。在这一集中，我们将快速浏览go-micro的代理功能。

正如前面的文章提到的，go-micro是一个可插拔的框架，它连接了许多不同的常用技术。如果你看看[插件仓库](https://github.com/micro/go-plugins)，你会看到它支持多少插件。

在我们的例子中，我们将使用NATS代理插件。

## 基于事件驱动的架构

[事件驱动的架构](https://en.wikipedia.org/wiki/Event-driven_architecture)是一个非常简单的概念。我们通常认为好的架构是要解耦的，一个服务不应该与其他服务耦合或者感知到其他服务。当我们使用诸如`gRPC`协议时，在某些情况下是正确的，我们以向`go.srv.user-service`发布请求为例。其中就使用了服务发现的方式来查找该服务的实际位置。 尽管这并不直接将我们与实现耦合，但它确实将该服务耦合到了其他名为`go.srv.user-service`的服务，因此它不是完全的解耦，因为它直接与其他服务进行交互。

那么什么让事件驱动架构真正的解耦呢？为了理解这一点，我们首先看看发布和订阅事件的过程。服务a完成了一项任务x，然后向系统发布一个事件`x刚刚发生了`。服务并不需要知道或者关心谁在监听这个事件，或者该事件正在发生什么影响。这些事情留给了监听事件的客户端。如果你期待n个服务对某个事件采取行动，那么也很容易。例如，你想12个不同的服务针对使用`gRPC`创建新用户采取行动，可能需要在用户服务中实例化12个客户端。而借助事件发布订阅或事件驱动架构，你的服务就不需要关心这些。

现在，客户端服务只需要简单的监听事件。这意味着，你需要中间的介质来接收这些事件，并通知订阅了事件的监听着的客户端。

这这篇文章中，我们将在每次创建用户时创建一个事件，并且将创建一个用于发送电子邮件的新服务。我们不会真的去实现发邮件的功能，只是模拟它。

## 代码

首先，我们需要将NATS代理插件集成到我们的用户服务中：

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

确保你正在运行Postgres，然后让我们运行这个服务：

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

在运行之前，我们需要启动[NATS](https://nats.io/)...

```
$ docker run -d -p 4222:4222 nats
```

另外，我想快速解释一下go-micro的一部分，我觉得这对于理解它作为框架如何工作很重要。你会注意到：

```go
srv.Init()
pubsub := srv.Server().Options().Broker
```






