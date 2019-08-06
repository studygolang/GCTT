首发于：https://studygolang.com/articles/22111

# Go语言中的微服务

## 摘要

我最近在墨尔本 Golang 聚会上就如何开发微服务和框架做了一次演讲。在本文中，我将与您分享我的想法（此外，它对我来说是一个很好的复习）。

在这里，我要介绍以下框架：

* [Go Micro](https://micro.mu/)
* [Go Kit](https://gokit.io/)
* [Gizmo](https://github.com/NYTimes/gizmo)
* [Kite](https://github.com/koding/kite)

## 框架简介

### Go Micro

这是我认为最受欢迎的框架之一。有很多博客文章和简单的例子可供使用参考。您可以从 [microhq](https://medium.com/microhq) 在 Medium 或 [@MicroHQ](https://twitter.com/MicroHQ) 获得 Go Micro 的最新更新。

那么，什么是 Go Micro ?

它是一个可拔插的 RPC 框架，用于在 Go 中编写微服务。开箱即用，您将看到：

* 服务发现 - 自动向服务发现系统注册的应用程序。
* 负载均衡 - 客户端负载均衡，用于平衡服务实例之间请求的负载。
* 同步通信 - 提供请求/响应传输层。
* 异步通信 - 内置发布/订阅功能。
* 消息编码 - 基于消息的 Content-Type 请求头的编码/解码。
* RPC 客户端/服务器打包 - 利用上述特性并公开接口来构建微服务。

 Go Micro 架构可以描述为三层堆栈。

![图1.Go Micro架构](https://raw.githubusercontent.com/studygolang/gctt-images/master/microservices-in-go/goMicro.png)

顶层包括 **Server-Client** 模型和服务抽象。该服务器是用于编写服务的基础。而客户端提供了一个接口，用于向服务端发起请求。

底层包含以下类型的插件：

* Broker - 提供一个消息代理接口，用于**异步发布/订阅通信**。
* Codec - 用于编码/解码消息。支持的格式包括 json,bson,protobuf,msgpack 等。
* Registry - 提供服务发现机制（默认为 Consul ）。
* Selector - 基于注册表构建的负载均衡抽象。
它允许使用诸如 random,roundrobin,leastconn 等算法“选择”服务。
* Transport - 服务之间同步请求/响应通信的接口。

 Go Micro 还提供 Sidecar 等功能。这允许您使用Go以外的语言编写的服务。
 Sidecar 提供服务注册，gRPC 编码/解码和HTTP处理程序。它有多种语言版本。

### Go Kit

Go Kit 是一个用于在Go中构建微服务的编程工具包。与 Go Micro 不同，它是一个旨在导入二进制包的库。

Go Kit 遵循简单的规则，例如：

* 没有全局状态
* 声明性构造
* 显式依赖
* 接口作为契约
* 领域驱动设计

在 Go Kit 中，您可以找到以下包：

* 身份验证 - basic和JWT。
* 传输 - HTTP，Nats，gRPC 等。
* 日志记录 - 服务中结构化日志记录的通用接口。
* 软件度量 - CloudWatch,Statsd,Graphite等。
* 追踪 - Zipkin 和 Opentracing。
* 服务发现 - Consul,Etcd,Eureka等。
* 熔断器 - Hystrix 的 Go 语言实现。

您可以在Peter Bourgon的文章和演示幻灯片中找到 Go Kit 的最佳描述之一：

* [Go kit: Go in the modern enterprise](https://peter.bourgon.org/go-kit/?source=post_page)
* [Go + microservices](https://github.com/peterbourgon/go-microservices?source=post_page)

此外，在“Go + microservices”幻灯片中，您将找到使用 Go Kit 构建的服务架构的示例。
有关快速参考，请参阅服务架构图。

![图2.使用Go Kit构建的服务架构示例 Go Micro 架构](https://raw.githubusercontent.com/studygolang/gctt-images/master/microservices-in-go/Go%2Bmicroservices.png)

### Gizmo

Gizmo 是纽约时报的微服务工具包。它提供了将服务器和 pubsub 守护进程组合在一起的软件包。它公开了以下包：

* [server](https://godoc.org/github.com/NYTimes/gizmo/server) - 提供两种服务器实现：SimpleServer（通过 HTTP ），RPCServer（通过 gRPC ）。
* [server/kit](https://godoc.org/github.com/NYTimes/gizmo/server/kit) - 基于 Go Kit 的实验包。
* [config](https://godoc.org/github.com/NYTimes/gizmo/config) - 包含功能：解析 JSON 文件，Consul 键值对中的 JSON blob ，或者环境变量。
* [pubsub](https://godoc.org/github.com/NYTimes/gizmo/pubsub) - 提供通用接口，用于从队列中发布和使用数据。
* [pubsub/pubsubtest](https://godoc.org/github.com/NYTimes/gizmo/pubsub/pubsubtest) - 包含发布者和订阅者接口的测试实现。
* [web](https://godoc.org/github.com/NYTimes/gizmo/web) - 公开用于从请求查询和有效负载中解析类型的函数。

Pubsub包提供了使用以下队列的接口：

* [pubsub/aws](https://godoc.org/github.com/NYTimes/gizmo/pubsub/aws) - 适用于 Amazon SNS/SQS。
* [pubsub/gcp](https://godoc.org/github.com/NYTimes/gizmo/pubsub/gcp) - 适用于 Google Pubsub。
* [pubsub/kafka](https://godoc.org/github.com/NYTimes/gizmo/pubsub/kafka) - 适用于 Kafka主题。
* [pubsub/http](https://godoc.org/github.com/NYTimes/gizmo/pubsub/http) - 用于通过 HTTP 发布。

因此，在我看来，Gizmo 介于 Go Micro 和 Go Kit 之间。它不像 Go Micro 那样完全的“黑盒”。与此同时，它并不像 Go Kit 那么粗糙。它提供更高级别的构建组件，例如config和pubsub包。

### Kite

Kite 是一个在 Go 中开发微服务的框架。它公开了 RPC 客户端和服务端的包。
创建的服务会自动注册到服务发现系统 Kontrol 。Kontrol 是用 Kite 写的，它本身就是 Kite 服务。
这意味着 Kite 微服务在自己的环境中运行良好。如果您需要将 Kite 微服务连接到另一个服务发现系统，
则需要进行自定义。这是我从名单中删除 Kite 的主要原因，并决定不讨论这个框架。

## 比较框架

我将使用四个类别比较框架：

* 客观比较 - GitHub 统计
* 文档和示例
* 用户和社区
* 代码质量。

### GitHub统计

![表1. Go 微服务框架统计（2018年4月收集）](https://raw.githubusercontent.com/studygolang/gctt-images/master/microservices-in-go/MicroStatics.png)

### 文档和示例

简单来说，没有一个框架提供可靠的文档。通常，唯一的正式文档是 repo 首页上的 Readme 文件。

对于 Go Micro，可以在 [micro.mu](https://micro.mu/),[microhq](https://medium.com/microhq) 和社交媒体 [@MicroHQ](https://twitter.com/MicroHQ) 上获得大量信息和公告。

如果是 Go Kit，您可以在 [Peter Bourgon](https://peter.bourgon.org/go-kit/) 的博客中找到最好的文档。我发现的一个最好的例子是在 [ru-rocker](http://www.ru-rocker.com/2017/04/17/micro-services-using-go-kit-service-discovery/) 博客中。

使用 Gizmo，源代码提供了最好的文档和示例。

总而言之，如果你来自 NodeJS 世界，并希望看到类似 ExpressJS 的教程，你会感到失望。
另一方面，这是创建自己的教程的绝佳机会。

### 用户和社区

Go Kit 是最受欢迎的微服务框架，基于 GitHub 统计数据 - 在本出版物发布时超过10k星。它拥有大量的贡献者（122）和超过1000个分叉。
最后，Go Kit 由 [DigitalOcean](https://www.digitalocean.com/) 提供支持。

Go Micro 第二，拥有超过 3600 颗 stars ，27 个贡献者和 385 个 forks 。Six Micro 的最大赞助商之一是 [Sixt](https://www.sixt.com/)。
Gizmo 第三，超过 2200 颗 star, 31 个贡献者和 137 个 forks 。由纽约时报支持和创建。

### 代码质量

* Go Kit 在代码质量类别中排名第一。它拥有近 80％ 的代码覆盖率和出色的 [Go 报告评级](https://goreportcard.com/report/github.com/go-kit/kit)。
* Gizmo 也有很好的 [Go 报告评级](https://goreportcard.com/report/github.com/NYTimes/gizmo)。但它的代码覆盖率仅为 46％。
* Go Micro 不提供覆盖率信息，但它确实具有很好的 [Go 报告评级](https://goreportcard.com/report/github.com/micro/go-micro)。

## 微服务代码实践

好吧，已有足够的理论。下边，为了更好地理解框架，我创建了三个简单的微服务。

![图3.实际示例架构](https://raw.githubusercontent.com/studygolang/gctt-images/master/microservices-in-go/micro_practice.png)

这些是实现一个业务功能的服务——"Greeting"。
当用户将 "name" 参数传递给服务器时，该服务会发送 Greeting 响应。此外，所有服务均符合以下要求：

* 服务应自行注册服务发现系统。
* 服务应具有健康检查接口。
* 服务应至少支持 HTTP 和 gRPC 传输。

对于那些喜欢阅读源代码的人。
您可以在此处阅读 repo 中的[源代码](https://github.com/antklim/go-microservices).

### Go Micro greeter

使用 Go Micro 创建服务需要做的第一件事是定义 protobuf 描述。
方便后期，所有三项服务都采用了相同的 protobuf 定义。我创建了以下服务描述：

```proto
syntax = "proto3";

package pb;

service Greeter {
  rpc Greeting(GreetingRequest) returns (GreetingResponse) {}
}

message GreetingRequest {
  string name = 1;
}

message GreetingResponse {
  string greeting = 2;
}
```

接口包含一种方法—— "Greeting"。
请求中有一个参数—— 'name'，响应中有一个参数 - 'greeting'。

然后我使用修改后的 [protoc工具](https://github.com/micro/protoc-gen-micro) 通过 protobuf 文件生成服务接口。
该生成器由 Go Micro fork 并进行了修改，以支持该框架的一些功能。
我在 “greeting” 服务中将这些连接在一起。此时，该服务正在启动并注册服务发现系统。
它只支持 gRPC 传输协议：

```go
package main

import (
    "log"

    pb "github.com/antklim/go-microservices/go-micro-greeter/pb"
    "github.com/micro/go-micro"
    "golang.org/x/net/context"
)

// Greeter 实现了 greeter 服务.
type Greeter struct{}

// Greeting 方法实现.
func (g *Greeter) Greeting(ctx context.Context, in *pb.GreetingRequest, out *pb.GreetingResponse) error {
    out.Greeting = "GO-MICRO Hello " + in.Name
    return nil
}

func main() {
    service := micro.NewService(
        micro.Name("go-micro-srv-greeter"),
        micro.Version("latest"),
    )

    service.Init()

    pb.RegisterGreeterHandler(service.Server(), new(Greeter))

    if err := service.Run(); err != nil {
        log.Fatal(err)
    }
}
```

为了支持HTTP传输，我不得不添加其他模块。它将HTTP请求映射到 protobuf 定义的请求。并称为 gRPC 服务。
然后，它将服务响应映射到 HTTP 响应并将其回复给用户。

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"

    proto "github.com/antklim/go-microservices/go-micro-greeter/pb"
    "github.com/micro/go-micro/client"
    web "github.com/micro/go-web"
)

func main() {
    service := web.NewService(
        web.Name("go-micro-web-greeter"),
    )

    service.HandleFunc("/greeting", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" {
            var name string
            vars := r.URL.Query()
            names, exists := vars["name"]
            if !exists || len(names) != 1 {
                name = ""
            } else {
                name = names[0]
            }

            cl := proto.NewGreeterClient("go-micro-srv-greeter", client.DefaultClient)
            rsp, err := cl.Greeting(context.Background(), &proto.GreetingRequest{Name: name})
            if err != nil {
                http.Error(w, err.Error(), 500)
                return
            }

            js, err := json.Marshal(rsp)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            w.Header().Set("Content-Type", "application/json")
            w.Write(js)
            return
        }
    })

    if err := service.Init(); err != nil {
        log.Fatal(err)
    }

    if err := service.Run(); err != nil {
        log.Fatal(err)
    }
}
```

非常简单明了。 Go Micro 在幕后处理了许多事情——例如在服务发现系统中注册。
另一方面，创建纯 HTTP 服务很困难。

### Go Kit greeter

完成 Go Micro 后，我转到了 Go Kit 服务实现。
我花了很多时间阅读 Go Kit 存储库中提供的代码示例。
理解端点的概念花了我很多时间。下一个耗时的难题是服务发现注册商的代码。直到在找到一个 [不错的例子](http://www.ru-rocker.com/2017/04/17/micro-services-using-go-kit-service-discovery/) 后我才实现它。

最后，我创建了四个包：

* 服务逻辑实现。
* 与传输无关的服务端点。
* 传输特定端点 (gRPC,HTTP)
* 服务发现注册商。

```go
package greeterservice

// Service 定义 greetings 服务接口.
type Service interface {
    Health() bool
    Greeting(name string) string
}

// GreeterService 实现 Service 接口.
type GreeterService struct{}

// Health 实现 Service 接口 Health 方法.
func (GreeterService) Health() bool {
    return true
}

// Greeting 实现 Service 接口 Greeting 方法.
func (GreeterService) Greeting(name string) (greeting string) {
    greeting = "GO-KIT Hello " + name
    return
}

```

如您所见，代码没有任何依赖关系。它只是实现逻辑。下一个代码段展示了端点定义：

```go
package greeterendpoint

import (
    "context"

    "github.com/go-kit/kit/log"

    "github.com/antklim/go-microservices/go-kit-greeter/pkg/greeterservice"
    "github.com/go-kit/kit/endpoint"
)

// Endpoints 包含了所有组成 greeter 服务的端点。
// 它被用作一个辅助结构，将所有端点收集到一个参数中。
type Endpoints struct {
    HealthEndpoint   endpoint.Endpoint // used by Consul for the healthcheck
    GreetingEndpoint endpoint.Endpoint
}

// MakeServerEndpoints 返回服务端点, 绑定在提供的中间件上。
func MakeServerEndpoints(s greeterservice.Service, logger log.Logger) Endpoints {
    var healthEndpoint endpoint.Endpoint
    {
        healthEndpoint = MakeHealthEndpoint(s)
        healthEndpoint = LoggingMiddleware(log.With(logger, "method", "Health"))(healthEndpoint)
    }

    var greetingEndpoint endpoint.Endpoint
    {
        greetingEndpoint = MakeGreetingEndpoint(s)
        greetingEndpoint = LoggingMiddleware(log.With(logger, "method", "Greeting"))(greetingEndpoint)
    }

    return Endpoints{
        HealthEndpoint:   healthEndpoint,
        GreetingEndpoint: greetingEndpoint,
    }
}

// MakeHealthEndpoint 构造封装服务的 Health 端点。
func MakeHealthEndpoint(s greeterservice.Service) endpoint.Endpoint {
    return func(ctx context.Context, request interface{}) (response interface{}, err error) {
        healthy := s.Health()
        return HealthResponse{Healthy: healthy}, nil
    }
}

// MakeGreetingEndpoint 构造封装服务的 Greeter 端点。
func MakeGreetingEndpoint(s greeterservice.Service) endpoint.Endpoint {
    return func(ctx context.Context, request interface{}) (response interface{}, err error) {
        req := request.(GreetingRequest)
        greeting := s.Greeting(req.Name)
        return GreetingResponse{Greeting: greeting}, nil
    }
}

// Failer 是实现响应类型的接口。
// 响应可以被检验，是否是 Failer 接口，如果是，那么就是失败的响应，
// 而且，如果是，则根据错误使用单独的写路径对它们进行编码
type Failer interface {
    Failed() error
}

// HealthRequest 包含了 Health 方法的所有请求参数.
type HealthRequest struct{}

// HealthResponse 包含了 Health 方法的响应值。
type HealthResponse struct {
    Healthy bool  `json:"healthy,omitempty"`
    Err     error `json:"err,omitempty"`
}

// Failed 实现 Failer 接口。
func (r HealthResponse) Failed() error { return r.Err }

// GreetingRequest 包含了 Greeting 方法的所有请求参数.
type GreetingRequest struct {
    Name string `json:"name,omitempty"`
}

// GreetingResponse 包含了 Greeting 方法的响应值
type GreetingResponse struct {
    Greeting string `json:"greeting,omitempty"`
    Err      error  `json:"err,omitempty"`
}

// Failed 实现 Failer 接口。
func (r GreetingResponse) Failed() error { return r.Err }
```

在定义了服务和端点之后，我开始通过不同的传输协议公开端点。我从 HTTP 传输开始：

```go
package greetertransport

import (
    "context"
    "encoding/json"
    "errors"
    "net/http"

    "github.com/antklim/go-microservices/go-kit-greeter/pkg/greeterendpoint"
    "github.com/go-kit/kit/log"
    httptransport "github.com/go-kit/kit/transport/http"
    "github.com/gorilla/mux"
)

var (
    // ErrBadRouting 无效路径错误.
    ErrBadRouting = errors.New("inconsistent mapping between route and handler")
)

// NewHTTPHandler 返回一个 HTTP 处理程序（handler），该处理程序使一组端点在预定义的路径上可用。
func NewHTTPHandler(endpoints greeterendpoint.Endpoints, logger log.Logger) http.Handler {
    m := mux.NewRouter()
    options := []httptransport.ServerOption{
        httptransport.ServerErrorEncoder(encodeError),
        httptransport.ServerErrorLogger(logger),
    }

    // GET /health         查找服务健康信息
    // GET /greeting?name  查找 greeting

    m.Methods("GET").Path("/health").Handler(httptransport.NewServer(
        endpoints.HealthEndpoint,
        DecodeHTTPHealthRequest,
        EncodeHTTPGenericResponse,
        options...,
    ))
    m.Methods("GET").Path("/greeting").Handler(httptransport.NewServer(
        endpoints.GreetingEndpoint,
        DecodeHTTPGreetingRequest,
        EncodeHTTPGenericResponse,
        options...,
    ))
    return m
}

// DecodeHTTPHealthRequest 方法.
func DecodeHTTPHealthRequest(_ context.Context, _ *http.Request) (interface{}, error) {
    return greeterendpoint.HealthRequest{}, nil
}

// DecodeHTTPGreetingRequest 方法.
func DecodeHTTPGreetingRequest(_ context.Context, r *http.Request) (interface{}, error) {
    vars := r.URL.Query()
    names, exists := vars["name"]
    if !exists || len(names) != 1 {
        return nil, ErrBadRouting
    }
    req := greeterendpoint.GreetingRequest{Name: names[0]}
    return req, nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
    w.WriteHeader(err2code(err))
    json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

func err2code(err error) int {
    switch err {
    default:
        return http.StatusInternalServerError
    }
}

type errorWrapper struct {
    Error string `json:"error"`
}

// EncodeHTTPGenericResponse is a transport/http.
// EncodeResponseFunc 返回 json 响应。
func EncodeHTTPGenericResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
    if f, ok := response.(greeterendpoint.Failer); ok && f.Failed() != nil {
        encodeError(ctx, f.Failed(), w)
        return nil
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    return json.NewEncoder(w).Encode(response)
}
```

在我开始 gRPC 端点实现之前，我不需要重新定义需要 protobuf.
我复制了 Go Micro 服务 protobuf 文件。但就 Go Kit 而言，我使用默认服务生成器来创建服务接口。

 protobuf 定义的服务接口生成器

```
#!/usr/bin/env sh

protoc greeter.proto --go_out=plugins=grpc:.
```

Go Kit 服务 gRPC 端点

```go
package greetertransport

import (
    "context"

    "github.com/antklim/go-microservices/go-kit-greeter/pb"
    "github.com/antklim/go-microservices/go-kit-greeter/pkg/greeterendpoint"
    "github.com/go-kit/kit/log"
    grpctransport "github.com/go-kit/kit/transport/grpc"
    oldcontext "golang.org/x/net/context"
)

type grpcServer struct {
    greeter grpctransport.Handler
}

// NewGRPCServer 构建了 gRPC 可用的端点.
func NewGRPCServer(endpoints greeterendpoint.Endpoints, logger log.Logger) pb.GreeterServer {
    options := []grpctransport.ServerOption{
        grpctransport.ServerErrorLogger(logger),
    }

    return &grpcServer{
        greeter: grpctransport.NewServer(
            endpoints.GreetingEndpoint,
            decodeGRPCGreetingRequest,
            encodeGRPCGreetingResponse,
            options...,
        ),
    }
}

// Greeting 实现 GreeterService 接口 Greeting 方法.
func (s *grpcServer) Greeting(ctx oldcontext.Context, req *pb.GreetingRequest) (*pb.GreetingResponse, error) {
    _, res, err := s.greeter.ServeGRPC(ctx, req)
    if err != nil {
        return nil, err
    }
    return res.(*pb.GreetingResponse), nil
}

// decodeGRPCGreetingRequest is a transport/grpc.
// DecodeRequestFunc 将 gRPC 请求转换为用户域的 greeting 请求。
func decodeGRPCGreetingRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
    req := grpcReq.(*pb.GreetingRequest)
    return greeterendpoint.GreetingRequest{Name: req.Name}, nil
}

// encodeGRPCGreetingResponse is a transport/grpc.
// DecodeRequestFunc 将 用户域的 greeting 转换为请求gRPC 请求。
func encodeGRPCGreetingResponse(_ context.Context, response interface{}) (interface{}, error) {
    res := response.(greeterendpoint.GreetingResponse)
    return &pb.GreetingResponse{Greeting: res.Greeting}, nil
}
```

最后，我实现了服务发现注册器：

```go
package greetersd

import (
    "math/rand"
    "os"
    "strconv"
    "time"

    "github.com/go-kit/kit/log"
    "github.com/go-kit/kit/sd"
    consulsd "github.com/go-kit/kit/sd/consul"
    "github.com/hashicorp/consul/api"
)

// ConsulRegister method.
func ConsulRegister(consulAddress string,
    consulPort string,
    advertiseAddress string,
    advertisePort string) (registar sd.Registrar) {

    // 日志
    var logger log.Logger
    {
        logger = log.NewLogfmtLogger(os.Stderr)
        logger = log.With(logger, "ts", log.DefaultTimestampUTC)
        logger = log.With(logger, "caller", log.DefaultCaller)
    }

    rand.Seed(time.Now().UTC().UnixNano())

    // 服务发现，我们使用 Consul.
    var client consulsd.Client
    {
        consulConfig := api.DefaultConfig()
        consulConfig.Address = consulAddress + ":" + consulPort
        consulClient, err := api.NewClient(consulConfig)
        if err != nil {
            logger.Log("err", err)
            os.Exit(1)
        }
        client = consulsd.NewClient(consulClient)
    }

    check := api.AgentServiceCheck{
        HTTP:     "http://" + advertiseAddress + ":" + advertisePort + "/health",
        Interval: "10s",
        Timeout:  "1s",
        Notes:    "Basic health checks",
    }

    port, _ := strconv.Atoi(advertisePort)
    num := rand.Intn(100) // 服务 ID 唯一
    asr := api.AgentServiceRegistration{
        ID:      "go-kit-srv-greeter-" + strconv.Itoa(num),
        Name:    "go-kit-srv-greeter",
        Address: advertiseAddress,
        Port:    port,
        Tags:    []string{"go-kit", "greeter"},
        Check:   &check,
    }
    registar = consulsd.NewRegistrar(client, &asr, logger)
    return
}
```

在准备好所有构建块之后，我将它们连接在服务启动器中：

Go Kit 服务启动器

```go
package main

import (
    "flag"
    "fmt"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "text/tabwriter"

    "github.com/antklim/go-microservices/go-kit-greeter/pb"
    "google.golang.org/grpc"

    "github.com/antklim/go-microservices/go-kit-greeter/pkg/greeterendpoint"
    "github.com/antklim/go-microservices/go-kit-greeter/pkg/greetersd"
    "github.com/antklim/go-microservices/go-kit-greeter/pkg/greeterservice"
    "github.com/antklim/go-microservices/go-kit-greeter/pkg/greetertransport"

    "github.com/go-kit/kit/log"
    "github.com/oklog/oklog/pkg/group"
)

func main() {
    fs := flag.NewFlagSet("greetersvc", flag.ExitOnError)
    var (
        debugAddr  = fs.String("debug.addr", ":9100", "Debug and metrics listen address")
        consulAddr = fs.String("consul.addr", "", "Consul Address")
        consulPort = fs.String("consul.port", "8500", "Consul Port")
        httpAddr   = fs.String("http.addr", "", "HTTP Listen Address")
        httpPort   = fs.String("http.port", "9110", "HTTP Listen Port")
        grpcAddr   = fs.String("grpc-addr", ":9120", "gRPC listen address")
    )
    fs.Usage = usageFor(fs, os.Args[0]+" [flags]")
    fs.Parse(os.Args[1:])

    var logger log.Logger
    {
        logger = log.NewLogfmtLogger(os.Stderr)
        logger = log.With(logger, "ts", log.DefaultTimestampUTC)
        logger = log.With(logger, "caller", log.DefaultCaller)
    }

    var service greeterservice.Service
    {
        service = greeterservice.GreeterService{}
        service = greeterservice.LoggingMiddleware(logger)(service)
    }

    var (
        endpoints   = greeterendpoint.MakeServerEndpoints(service, logger)
        httpHandler = greetertransport.NewHTTPHandler(endpoints, logger)
        registar    = greetersd.ConsulRegister(*consulAddr, *consulPort, *httpAddr, *httpPort)
        grpcServer  = greetertransport.NewGRPCServer(endpoints, logger)
    )

    var g group.Group
    {
        // 调试功能带 http.DefaultServeMux, 并提供Go调试和分析路由等功能
        debugListener, err := net.Listen("tcp", *debugAddr)
        if err != nil {
            logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
            os.Exit(1)
        }
        g.Add(func() error {
            logger.Log("transport", "debug/HTTP", "addr", *debugAddr)
            return http.Serve(debugListener, http.DefaultServeMux)
        }, func(error) {
            debugListener.Close()
        })
    }
    {
        // 服务发现注册
        g.Add(func() error {
            logger.Log("transport", "HTTP", "addr", *httpAddr, "port", *httpPort)
            registar.Register()
            return http.ListenAndServe(":"+*httpPort, httpHandler)
        }, func(error) {
            registar.Deregister()
        })
    }
    {
        // gRPC 加载我们创建的服务.
        grpcListener, err := net.Listen("tcp", *grpcAddr)
        if err != nil {
            logger.Log("transport", "gRPC", "during", "Listen", "err", err)
            os.Exit(1)
        }
        g.Add(func() error {
            logger.Log("transport", "gRPC", "addr", *grpcAddr)
            baseServer := grpc.NewServer()
            pb.RegisterGreeterServer(baseServer, grpcServer)
            return baseServer.Serve(grpcListener)
        }, func(error) {
            grpcListener.Close()
        })
    }
    {
        // 监听 Ctrl+C 信号终止.
        cancelInterrupt := make(chan struct{})
        g.Add(func() error {
            c := make(chan os.Signal, 1)
            signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
            select {
            case sig := <-c:
                return fmt.Errorf("received signal %s", sig)
            case <-cancelInterrupt:
                return nil
            }
        }, func(error) {
            close(cancelInterrupt)
        })
    }
    logger.Log("exit", g.Run())
}

func usageFor(fs *flag.FlagSet, short string) func() {
    return func() {
        fmt.Fprintf(os.Stderr, "USAGE\n")
        fmt.Fprintf(os.Stderr, "  %s\n", short)
        fmt.Fprintf(os.Stderr, "\n")
        fmt.Fprintf(os.Stderr, "FLAGS\n")
        w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
        fs.VisitAll(func(f *flag.Flag) {
            fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
        })
        w.Flush()
        fmt.Fprintf(os.Stderr, "\n")
    }
}
```

您可能已经注意到，我在几个地方使用了日志逻辑中间件。它允许我将记录逻辑与主服务/端点工作流分离。

服务级别日志记录中间件

```go
package greeterservice

import (
    "time"

    "github.com/go-kit/kit/log"
)

// ServiceMiddleware 定义了 service 中间件.
type ServiceMiddleware func(Service) Service

// LoggingMiddleware 使用 logger 作为依赖，返回一个 Service中间件
func LoggingMiddleware(logger log.Logger) ServiceMiddleware {
    return func(next Service) Service {
        return loggingMiddleware{next, logger}
    }
}

type loggingMiddleware struct {
    Service
    logger log.Logger
}

func (m loggingMiddleware) Health() (healthy bool) {
    defer func(begin time.Time) {
        m.logger.Log(
            "method", "Health",
            "healthy", healthy,
            "took", time.Since(begin),
        )
    }(time.Now())
    healthy = m.Service.Health()
    return
}

func (m loggingMiddleware) Greeting(name string) (greeting string) {
    defer func(begin time.Time) {
        m.logger.Log(
            "method", "Greeting",
            "name", name,
            "greeting", greeting,
            "took", time.Since(begin),
        )
    }(time.Now())
    greeting = m.Service.Greeting(name)
    return
}
```

端点级别记录中间件

```go
package greeterendpoint

import (
    "context"
    "time"

    "github.com/go-kit/kit/endpoint"
    "github.com/go-kit/kit/log"
)

// LoggingMiddleware 返回端点日志中间件，
// 提供运行过程中日志信息，如果有错，提供错误信息
func LoggingMiddleware(logger log.Logger) endpoint.Middleware {
    return func(next endpoint.Endpoint) endpoint.Endpoint {
        return func(ctx context.Context, request interface{}) (response interface{}, err error) {
            defer func(begin time.Time) {
                logger.Log("transport_error", err, "took", time.Since(begin))
            }(time.Now())
            return next(ctx, request)
        }
    }
}
```

### Gizmo greeter

我以与 Go Kit 类似的方式创建了Gizmo服务。我为服务，端点，传输和服务发现注册商定义了四个包。

服务实现和服务发现系统注册器与 Go Kit 服务共享相同的代码。但是端点定义和传输实现必须根据Gizmo功能完成。

Gizmo Greeting 端点

```go
package greeterendpoint

import (
    "net/http"

    ocontext "golang.org/x/net/context"

    "github.com/NYTimes/gizmo/server"
    "github.com/antklim/go-microservices/gizmo-greeter/pkg/greeterservice"
)

// Endpoints 包含所有组成 greeter 服务的端点
type Endpoints struct {
    HealthEndpoint   server.JSONContextEndpoint
    GreetingEndpoint server.JSONContextEndpoint
}

// MakeServerEndpoints 返回服务端点
func MakeServerEndpoints(s greeterservice.Service) Endpoints {
    healthEndpoint := MakeHealthEndpoint(s)
    greetingEndpoint := MakeGreetingEndpoint(s)

    return Endpoints{
        HealthEndpoint:   healthEndpoint,
        GreetingEndpoint: greetingEndpoint,
    }
}

// MakeHealthEndpoint 构造 Health 端点.
func MakeHealthEndpoint(s greeterservice.Service) server.JSONContextEndpoint {
    return func(ctx ocontext.Context, r *http.Request) (int, interface{}, error) {
        healthy := s.Health()
        return http.StatusOK, HealthResponse{Healthy: healthy}, nil
    }
}

// MakeGreetingEndpoint 构造 Greeting 端点.
func MakeGreetingEndpoint(s greeterservice.Service) server.JSONContextEndpoint {
    return func(ctx ocontext.Context, r *http.Request) (int, interface{}, error) {
        vars := r.URL.Query()
        names, exists := vars["name"]
        if !exists || len(names) != 1 {
            return http.StatusBadRequest, errorResponse{Error: "query parameter 'name' required"}, nil
        }
        greeting := s.Greeting(names[0])
        return http.StatusOK, GreetingResponse{Greeting: greeting}, nil
    }
}

// HealthRequest 包含了 Health 方法请求参数
type HealthRequest struct{}

// HealthResponse 包含了 Health 方法响应值
type HealthResponse struct {
    Healthy bool `json:"healthy,omitempty"`
}

// GreetingRequest 包含了 Greeting 方法请求参数
type GreetingRequest struct {
    Name string `json:"name,omitempty"`
}

// GreetingResponse 包含了 Greeting 方法响应值
type GreetingResponse struct {
    Greeting string `json:"greeting,omitempty"`
}

type errorResponse struct {
    Error string `json:"error"`
}
```

如您所见，代码段与 Go Kit 类似。主要区别在于应该返回的接口类型：

GizmoGreeting HTTP终端

```go
package greetertransport

import (
    "context"

    "github.com/NYTimes/gizmo/server"
    "google.golang.org/grpc"

    "errors"
    "net/http"

    "github.com/NYTimes/gziphandler"
    pb "github.com/antklim/go-microservices/gizmo-greeter/pb"
    "github.com/antklim/go-microservices/gizmo-greeter/pkg/greeterendpoint"
    "github.com/sirupsen/logrus"
)

type (
    // TService 会实现 server.RPCService （服务的RPC），以及处理服务端请求
    TService struct {
        Endpoints greeterendpoint.Endpoints
    }

    // Config 包含 server 相关 json 配置
    Config struct {
        Server *server.Config
    }
)

// NewTService 会使用给定的配置实例化 RPC 服务
func NewTService(cfg *Config, endpoints greeterendpoint.Endpoints) *TService {
    return &TService{Endpoints: endpoints}
}

// Prefix 返回所有端点服务使用的字符串前缀
func (s *TService) Prefix() string {
    return ""
}

// Service 向 TService 提供要服务的服务描述和实现。
func (s *TService) Service() (*grpc.ServiceDesc, interface{}) {
    return &pb.Greeter_serviceDesc, s
}

// Middleware 为所有请求挂载 http.Handler hook .
//在这个实现中，我们使用 GzipHandler 中间件来压缩我们的响应。
func (s *TService) Middleware(h http.Handler) http.Handler {
    return gziphandler.GzipHandler(h)
}

// ContextMiddleware 为所有请求挂载 server.ContextHAndler hook.
// 如果需要修饰请求上下文，这将非常方便。
func (s *TService) ContextMiddleware(h server.ContextHandler) server.ContextHandler {
    return h
}

// JSONMiddleware 为所有请求挂载 JSONEndpoint hooks.
//在这个实现中，我们使用它来提供应用程序日志记录，检查错误并提供通用响应。
func (s *TService) JSONMiddleware(j server.JSONContextEndpoint) server.JSONContextEndpoint {
    return func(ctx context.Context, r *http.Request) (int, interface{}, error) {

        status, res, err := j(ctx, r)
        if err != nil {
            server.LogWithFields(r).WithFields(logrus.Fields{
                "error": err,
            }).Error("problems with serving request")
            return http.StatusServiceUnavailable, nil, errors.New("sorry, this service is unavailable")
        }

        server.LogWithFields(r).Info("success!")
        return status, res, nil
    }
}

// ContextEndpoints 在你的服务是非 RPC 端点下，可以提供你需要的功能
// 此时，我们不需要 RPC,但是仍需要这个方法以实现 server.RPCService 接口.
func (s *TService) ContextEndpoints() map[string]map[string]server.ContextHandlerFunc {
    return map[string]map[string]server.ContextHandlerFunc{}
}

// JSONEndpoints 是TService中可用的所有端点的列表。
func (s *TService) JSONEndpoints() map[string]map[string]server.JSONContextEndpoint {
    return map[string]map[string]server.JSONContextEndpoint{
        "/health": map[string]server.JSONContextEndpoint{
            "GET": s.Endpoints.HealthEndpoint,
        },
        "/greeting": map[string]server.JSONContextEndpoint{
            "GET": s.Endpoints.GreetingEndpoint,
        },
    }
}
```

GIzmo Greeting gRPC

```go
package greetertransport

import (
    pb "github.com/antklim/go-microservices/gizmo-greeter/pb"
    ocontext "golang.org/x/net/context"
)

// Greeting 实现 gRPC 服务.
func (s *TService) Greeting(ctx ocontext.Context, r *pb.GreetingRequest) (*pb.GreetingResponse, error) {
    return &pb.GreetingResponse{Greeting: "Hola Gizmo RPC " + r.Name}, nil
}
```

Go Kit 和 Gizmo 之间的显着差异在于传输实现。 Gizmo 提供了几种可以使用的服务类型。
我所要做的就是将HTTP路径映射到端点定义。低级HTTP请求/响应处理由 Gizmo 处理。

## 结论

 Go Micro 是推出微服务系统的最快方式。框架提供了许多功能。所以你不需要重新发明轮子。
但这种舒适和速度伴随着牺牲 - 灵活性。它不像 Go Kit 那样容易改变或更新系统的各个部分。并且它强制 gRPC 作为首选的通信类型。

您可能需要一段时间才能熟悉 Go Kit。它需要熟悉 Golang 特性和软件架构方面的经验。
另一方面，没有框架限制。所有部件都可以独立更改和更新。
 Gizmo 位于 Go Micro 和 Go Kit 之间。它提供了一些更高级别的抽象，例如 Service 包。
但缺乏文档和示例意味着我必须阅读源代码以了解不同的服务类型是如何工作的。使用 Gizmo 比使用 Go Kit 更容易。但它并不像 Go Micro 那么顺利。

这就是今天的一切。谢谢阅读。请查看微服务代码库以获取更多信息。如果您对Go和微服务框架有任何经验，请在下面的评论中分享。

--

via: <https://medium.com/seek-blog/microservices-in-go-2fc1570f6800>

作者：[Anton Klimenko](https://medium.com/@antklim)
译者：[TomatoAres](https://github.com/TomatoAres)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
