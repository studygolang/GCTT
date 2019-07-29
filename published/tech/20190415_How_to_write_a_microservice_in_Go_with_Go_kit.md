首发于：https://studygolang.com/articles/21378

# 如何使用 Go kit 工具包编写微服务

我在互联网上搜索了很久关于使用 Go kit 工具包编写微服务的精品教程（我认为我的 Google-fu 相当不错），但是

我没有找到 ......

来自[Go kit 代码库的示例](https://github.com/go-kit/kit/examples) 很好，但恕我直言，文档很枯燥。

然后我决定购买这本名为 `Go Programming Blueprints, 2nd Edition` 的书，这本书相当不错，但只有两章专门讨论 Go kit（一个用于实际开发微服务，一个用于实际部署）。我并不是真的现在关心 `gRPC`，本书第 10 章的例子也有所提及。如果你问我，那么脚手架代码很多 :P

Sooo，我决定向社区回馈一些东西并编写一个教程，以便“边做边学”。本教程将受到上述书籍的极大启发，并且可能在很多方面得到改进。

## 随意提供反馈

您可以在我的博客上找到指向微服务的完整源代码的链接，[coding.napolux.com](https://coding.napolux.com/how-to-write-a-microservice-in-go-with-go-kit/)

## 什么是 Go kit

[Go kit](https://github.com/go-kit/kit) README.md：

> Go kit 是一个编程工具包，用于在 Go 中构建微服务（或优雅的整体）。我们解决分布式系统和应用程序架构中的常见问题，因此您可以专注于提供业务价值
>
> [...]
>
> Go 是一种很棒的通用语言，但微服务需要一定的专业支持。RPC 安全性，系统可观察性，基础设施集成，甚至程序设计 - Go 工具包填补了标准库留下的空白，使 Go 成为在任何组织中编写微服务的一流语言。

我不想讨论太多：Go 对我而言太新了。当然存在喜欢它和不喜欢它的[讨论](https://gist.github.com/posener/330c2b08aaefdea6f900ff0543773b2e)。您还可以在这里找到一篇关于 Go 微服务框架差异的好[文章](https://medium.com/seek-blog/microservices-in-go-2fc1570f6800)。

## 我们会做什么

我们将创建一个非常基本的微服务，它将返回并验证日期 ... 目标是了解 Go 工具包的工作原理，仅此而已。你可以轻松地复制所有的逻辑而不用 Go 套件，但我在这里学习，所以 ...

我希望您对下一个项目有一个良好的起点！

我们的微服务将有一些端点。

- 一个 `GET` 端点 `/status` 将返回一个简单的答案，确认微服务已启动并运行
- 一个 `GET` 端点 `/get` 将返回今天的日期
- 一个 `POST` 端点 `/validate` 将收到一个日期字符串 `dd/mm/yyyy`( 唯一存在的日期格式，如果你问我，问美国！）格式并根据一个简单的正则表达式验证

开始吧！！！

## 先决条件

你应该安装[`Golang`](https://golang.org/doc/install) 并在你的机器上工作。我发现[官方下载包](https://golang.org/doc/install) 比我的 Macbook 上的[Homebrew](https://brew.sh/) 安装更好（我的 env.vars 有些问题）。

另外，你应该知道 Go 语言，例如，我不会解释 `struct` 是什么。

## napodate 微服务

好的，让我们首先在我们的 $GOPATH 文件夹中创建一个名为 `napodate` 的新文件夹。这也是我们 `package` 的名称。
把 `service.go` 文件放在里面。让我们在文件顶部添加我们的服务接口。

```go
package napodate

import "context"

// Service provides some "date capabilities" to your application
type Service interface {
	Status(ctx context.Context) (string, error)
	Get(ctx context.Context) (string, error)
	Validate(ctx context.Context, date string) (bool, error)
}
```

在这里，我们为我们的服务定义了“蓝图”：在 Go kit 中，您必须将服务建模为接口。如上所述，我们将需要三个端点，这些端点将被映射到此接口。

我们为什么要使用这个 `context` 包？阅读[https://blog.golang.org/context](https://blog.golang.org/context)

> 在 Google，我们开发了一个上下文包，可以轻松地将 API 边界的请求范围值，取消信号和截止日期传递给处理请求所涉及的所有 Goroutine

基本上，这是必需的，因为我们的微服务应该从一开始就处理并发请求，并且每个请求的上下文都是强制性的。

有可能你会感到困惑。更多关于本教程内容会在后面讲诉。我们现在有了微服务接口。

## 实现我们的服务

您可能知道，如果没有实现，接口就什么都不是，所以让我们实现我们的服务。让我们再添加一些代码到 `service.go`。

```go
type dateService struct{}

// NewService makes a new Service.
func NewService() Service {
	return dateService{}
}

// Status only tell us that our service is ok!
func (dateService) Status(ctx context.Context) (string, error) {
	return "ok", nil
}

// Get will return today's date
func (dateService) Get(ctx context.Context) (string, error) {
	now := time.Now()
	return now.Format("02/01/2006"), nil
}

// Validate will check if the date today's date
func (dateService) Validate(ctx context.Context, date string) (bool, error) {
	_, err := time.Parse("02/01/2006", date)
	if err != nil {
		return false, err
	}
	return true, nil
}
```

新定义的类型 `dateService`（一个空结构）是我们如何将我们服务的方法组合在一起，同时以某种方式“隐藏”实现并在其他地方使用。

见 `NewService()` 作为我们的“对象”的构造函数。这就是我们所要求的获取服务实例的所有内容，同时屏蔽内部逻辑，就像优秀的程序员应该做的那样。

## 我们来写一个测试

在我们的服务测试中可以看到如何使用 `NewService()` 的一个很好的例子。继续创建一个 `service_test.go` 文件。

```go
package napodate

import (
	"context"
	"testing"
	"time"
)

func TestStatus(t *testing.T) {
	srv, ctx := setup()

	s, err := srv.Status(ctx)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	// testing status
	ok := s == "ok"
	if !ok {
		t.Errorf("expected service to be ok")
	}
}

func TestGet(t *testing.T) {
	srv, ctx := setup()
	d, err := srv.Get(ctx)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	time := time.Now()
	today := time.Format("02/01/2006")

	// testing today's date
	ok := today == d
	if !ok {
		t.Errorf("expected dates to be equal")
	}
}

func TestValidate(t *testing.T) {
	srv, ctx := setup()
	b, err := srv.Validate(ctx, "31/12/2019")
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	// testing that the date is valid
	if !b {
		t.Errorf("date should be valid")
	}

	// testing an invalid date
	b, err = srv.Validate(ctx, "31/31/2019")
	if b {
		t.Errorf("date should be invalid")
	}

	// testing a USA date date
	b, err = srv.Validate(ctx, "12/31/2019")
	if b {
		t.Errorf("USA date should be invalid")
	}
}

func setup() (srv Service, ctx context.Context) {
	return NewService(), context.Background()
}
```

我使测试更具可读性，但您应该使用 `Subtests` 编写它们，[点击了解详情](https://blog.golang.org/subtests)。

测试是绿色的（!）但是重点关注 `setup()` 方法。对于每个测试，我们使用 `NewService()` 和上下文返回我们的服务实例。

## Transports

我们的服务将使用 `HTTP` 公开。我们现在将模拟已接受的 `HTTP` 请求和响应。在 `service.go` 同一文件夹中创建一个 `transport.go` 文件。

```go
package napodate

import (
	"context"
	"encoding/json"
	"net/http"
)

// In the first part of the file we are mapping requests and responses to their JSON payload.
type getRequest struct{}

type getResponse struct {
	Date string `json:"date"`
	Err  string `json:"err,omitempty"`
}

type validateRequest struct {
	Date string `json:"date"`
}

type validateResponse struct {
	Valid bool   `json:"valid"`
	Err   string `json:"err,omitempty"`
}

type statusRequest struct{}

type statusResponse struct {
	Status string `json:"status"`
}

// In the second part we will write "decoders" for our incoming requests
func decodeGetRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req getRequest
	return req, nil
}

func decodeValidateRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req validateRequest
	err := JSON.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func decodeStatusRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req statusRequest
	return req, nil
}

// Last but not least, we have the encoder for the response output
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return JSON.NewEncoder(w).Encode(response)
}
```

如果你问我一些代码，但你会在 `transport.go` 文件中找到可以帮助你导航它的注释。

在文件的第一部分中，我们将请求和响应映射到它们的 JSON 实体。对于 `statusRequest` 和 `getRequest` 我们并不需要，因为没有有效载荷被发送到服务器。而 `validateRequest` 我们要传递一个要验证的日期，所以这里是 `date` 字段。

请求响应也非常简单。

在第二部分中，我们将为传入的请求编写“解码器”，告诉服务他应该如何转换请求并将它们映射到正确的请求结构。我知道 `get` 和 `status` 是空的，但他们在那里为完整起见。记住，我正在边做边学 ...

最后但并非最不重要的是，我们有响应输出的编码器，这是一个简单的 JSON 编码器：给定一个对象，我们将从中返回一个 JSON 对象。

这就是 `transports`, 让我们创造我们的端点！

## 端点

我们来创建一个新文件 `endpoint.go`。此文件将包含我们的端点，这些端点将来自客户端的请求映射到我们的内部服务

```go
package napodate

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints are exposed
type Endpoints struct {
	GetEndpoint      endpoint.Endpoint
	StatusEndpoint   endpoint.Endpoint
	ValidateEndpoint endpoint.Endpoint
}

// MakeGetEndpoint returns the response from our service "get"
func MakeGetEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(getRequest) // we really just need the request, we don't use any value from it
		d, err := srv.Get(ctx)
		if err != nil {
			return getResponse{d, err.Error()}, nil
		}
		return getResponse{d, ""}, nil
	}
}

// MakeStatusEndpoint returns the response from our service "status"
func MakeStatusEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(statusRequest) // we really just need the request, we don't use any value from it
		s, err := srv.Status(ctx)
		if err != nil {
			return statusResponse{s}, err
		}
		return statusResponse{s}, nil
	}
}

// MakeValidateEndpoint returns the response from our service "validate"
func MakeValidateEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(validateRequest)
		b, err := srv.Validate(ctx, req.Date)
		if err != nil {
			return validateResponse{b, err.Error()}, nil
		}
		return validateResponse{b, ""}, nil
	}
}

// Get endpoint mapping
func (e Endpoints) Get(ctx context.Context) (string, error) {
	req := getRequest{}
	resp, err := e.GetEndpoint(ctx, req)
	if err != nil {
		return "", err
	}
	getResp := resp.(getResponse)
	if getResp.Err != "" {
		return "", errors.New(getResp.Err)
	}
	return getResp.Date, nil
}

// Status endpoint mapping
func (e Endpoints) Status(ctx context.Context) (string, error) {
	req := statusRequest{}
	resp, err := e.StatusEndpoint(ctx, req)
	if err != nil {
		return "", err
	}
	statusResp := resp.(statusResponse)
	return statusResp.Status, nil
}

// Validate endpoint mapping
func (e Endpoints) Validate(ctx context.Context, date string) (bool, error) {
	req := validateRequest{Date: date}
	resp, err := e.ValidateEndpoint(ctx, req)
	if err != nil {
		return false, err
	}
	validateResp := resp.(validateResponse)
	if validateResp.Err != "" {
		return false, errors.New(validateResp.Err)
	}
	return validateResp.Valid, nil
}
```

让我们深入一点理解一下 ... 为了揭露所有我们的服务 `Get()`，`Status()` 和 `Validate()`。我们要编写将处理传入的请求，调用相应的服务方法，并根据该响应建立并返回一个适当的结果的功能函数。

这些方法就是 `Make...` 那些。它们将接收 `servuce` 作为参数，然后使用类型断言将请求类型“强制”转化为特定的一个，并使用它来调用服务方法。

在这些 `Make...` 方法（将在 `main.go` 文件中使用）之后，我们将编写端点以符合服务接口

```go
type Endpoints struct {
	GetEndpoint      endpoint.Endpoint
	StatusEndpoint   endpoint.Endpoint
	ValidateEndpoint endpoint.Endpoint
}
```

我们举一个例子：

```go
// Status endpoint mapping
func (e Endpoints) Status(ctx context.Context) (string, error) {
	req := statusRequest{}
	resp, err := e.StatusEndpoint(ctx, req)
	if err != nil {
		return "", err
	}
	statusResp := resp.(statusResponse)
	return statusResp.Status, nil
}
```

此方法将允许我们将端点用作 Go 方法。

## HTTP 服务器

对于我们的微服务，我们需要一个 HTTP 服务器。Go 对此非常有帮助，但我为我们的路由选择了[https://github.com/gorilla/mux](https://github.com/gorilla/mux)，因为它的语法看起来非常简洁，所以让我们创建一个简单的 HTTP 服务器，其中包含映射到我们的端点。

在项目种创建一个名为 `server.go` 的新文件。

```go
package napodate

import (
	"context"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// NewHTTPServer is a Good little server
func NewHTTPServer(ctx context.Context, endpoints Endpoints) http.Handler {
	r := mux.NewRouter()
	r.Use(commonMiddleware) // @see https://stackoverflow.com/a/51456342

	r.Methods("GET").Path("/status").Handler(httptransport.NewServer(
		endpoints.StatusEndpoint,
		decodeStatusRequest,
		encodeResponse,
	))

	r.Methods("GET").Path("/get").Handler(httptransport.NewServer(
		endpoints.GetEndpoint,
		decodeGetRequest,
		encodeResponse,
	))

	r.Methods("POST").Path("/validate").Handler(httptransport.NewServer(
		endpoints.ValidateEndpoint,
		decodeValidateRequest,
		encodeResponse,
	))

	return r
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
```

端点将从 `main.go` 文件传递到服务器，并且 `commonMiddleware()` 将负责为每个响应添加特定标头。

## 最后，我们的 main.go 文件

让我们结束吧！我们有一个端点服务。我们有一个 HTTP 服务器，我们只需要一个可以包装所有内容的地方，当然这是我们的 `main.go` 文件。把它放到一个新文件夹中，让我们称其为 `cmd`。

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"napodate"
)

func main() {
	var (
		httpAddr = flag.String("http", ":8080", "http listen address")
	)
	flag.Parse()
	ctx := context.Background()
	// our napodate service
	srv := napodate.NewService()
	errChan := make(chan error)

	Go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	// mapping endpoints
	endpoints := napodate.Endpoints{
		GetEndpoint:      napodate.MakeGetEndpoint(srv),
		StatusEndpoint:   napodate.MakeStatusEndpoint(srv),
		ValidateEndpoint: napodate.MakeValidateEndpoint(srv),
	}

	// HTTP transport
	Go func() {
		log.Println("napodate is listening on port:", *httpAddr)
		handler := napodate.NewHTTPServer(ctx, endpoints)
		errChan <- http.ListenAndServe(*httpAddr, handler)
	}()

	log.Fatalln(<-errChan)
}
```

让我们一起分析这个文件。我们声明 `main` 包并导入我们需要的东西。

我们使用一个[标志](https://gobyexample.com/command-line-flags) 来使监听端点并可配置，我们的服务的默认端点将是经典的 `8080` 但我们可以用任何端点来进行替换

接下来是我们服务器的设置：我们创建一个上下文（参见上面有关上下文的解释）并获得我们的服务。还设置了[错误通道](https://gobyexample.com/channels)。

> 通道是连接并发 Goroutine 的管道。您可以将值从一个 Goroutine 发送到通道，并将这些值接收到另一个 Goroutine 中。

然后我们创建两个 `goroutines`。一个在我们按下 `CTRL+C` 时停止服务器，一个实际上会监听传入的请求。

看看 `handler := napodate.NewHTTPServer(ctx, endpoints)` 这个处理程序将映射我们的服务端点（你还记得 `Make...` 上面的方法吗？）并返回正确的结果。

`NewHTTPServer()` 以前在哪里看到的？

一旦通道收到错误消息，服务器将停止并死亡。

## 我们的服务！

如果您正确地完成了所有操作，可以运行

```shell
go run cmd/main.go
```

从你的项目文件夹，你应该能够 `curl` 你的微服务！

```shell
curl http://localhost:8080/get
{"date":"14/04/2019"}

curl http://localhost:8080/status
{"status":"ok"}

curl -XPOST -d '{"date":"32/12/2020"}' http://localhost:8080/validate
{"valid":false,"err":"parsing time \"32/12/2020\": day out of range"}

curl -XPOST -d '{"date":"12/12/2021"}' http://localhost:8080/validate
{"valid":true}
```

## 总结一下

我们从零开始创建了一个新的微服务，即使它非常简单，也是开始使用 Go kit 和 Go 编程语言的好的开端。

希望你和我一样喜欢这个教程！

---

via: https://dev.to/napolux/how-to-write-a-microservice-in-go-with-go-kit-a66

作者：[Francesco Napoletano](https://dev.to/napolux)
译者：[lovechuck](https://github.com/lovechuck)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
