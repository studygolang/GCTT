# Golang中的微服务 - 第一章

## 介绍

Golang中的微服务系列总计十章，预计每周更新。本系列的解决方案采用了 protobuf 和 gRPC 作为底层传输协议。为什么采用这两个技术呢？我花了相当长的时间，才想出并决定采用这个方案。这个方案对开发者来说，非常清晰而简明。我也很乐意把自己在搭建、测试和部署端到端的微服务过程中的心得，分享给想接触这块的朋友们。

在这个教程中，我们将先接触几个基础的概念和术语，然后开始搭建第一个简单的微服务模型。

本系列中，我们将会创建以下服务：

- 委托
- 存货清单
- 用户
- 认证
- 角色
- 容器

整个技术堆从底至顶主要可划分为：golang、mongodb、grpc、docker、Google Cloud、Kubernetes、NATS、CircleCI、Terraform 和 go-micro。

接下来，你可以根据我的 [git路径](https://github.com/EwanValentine/shippy) （每篇文章都有自己的分支）中的指导逐步操作，不过必须注意把根据你的开发环境调整 GOPATH 。

同时你需要注意，我是在 Macbook 上开发的，你或许会需要替换 Makefiles 中的 `$(GOPATH)` 为 `$GOPATH` 。操作系统不一致带来的问题可能不止这一个，但这里就不一一列举了。

## 先决条件

- 掌握 golang 语言和其开发环境
- 安装 gRPC / protobuf [查看链接](https://grpc.io/docs/quickstart/go.html)
- 安装 golang [查看链接](https://golang.org/doc/install)
- 按照下列指令，安装 go 的第三方库

```
go get -u google.golang.org/grpc  
go get -u github.com/golang/protobuf/protoc-gen-go  
```

## 我们要搭建的是？

我们将搭建的是一个非常通用的微服务 —— 船运集装箱的管理平台。当然，我也可以用微服务搭建一个博客作为例子，但这实在是太简单了，我更希望能够展示**分离复杂性**的功能。最终我选择了这个管理平台为例，作为一个挑战！

那么，接下来我们先了解几个知识点：

## 什么是微服务？

在传统的单体应用中，所有系统的特性都被写入单个应用程序中。有时候我们用类型来区分这些特性，例如控制器、单元模块、工厂等等；其它情况下，例如在更大型的应用中，用互相间的关系或者各自的特征来区分应用特性，所以你可能会有授权程序包、好友关系处理包以及文章管理包，这些包可能都有各自的工厂、服务、数据库、数据模型等。不过，最终它们都被塞入了一个代码库中。

微服务就是把第二种解决方案做得更彻底：将原先的关系分离出来，每个程序包都保存到独立的、可运行的代码库中。

## 为什么选用微服务？

**复杂性** —— 依照特性把程序分割成多个微服务，有助于把大块代码分割成更小的模块。正如一句 Unix 中的老格言所说：把一件事做好（doing one thing well）。在单体应用的系统中，各模块倾向于紧密结合，模块间的关系很模糊。这个会导致系统的升级更为危险和复杂、存在更多的潜在 bug、集成的难度更高。

**扩展性** —— 在单体应用的系统中，总有特定模块的代码会比其余模块用得更为频繁，而你只能扩大整个库的规模来解决。例如你的鉴权模块被高频率地调用，对系统造成了高负荷的压力。于是你扩大了库规模，而原因仅仅是一个小小鉴权模块。

如果换成了微服务，那么你可以独立地扩大任何一个服务模块，这意味着我们可以更有效地进行横向扩展。这种分离性对多核、多区域的云计算带来了极大的帮助。

**Nginx 有个极好的微服务系列，讲述了各种概念，[请点击链接访问](https://www.nginx.com/blog/introduction-to-microservices/)。**

## 为什么选择 Golang？

几乎所有的语言都支持微服务。微服务不是一个具体的框架或工具，而是一个概念。这就意味着，在选择构建微服务的语言中，总有一些更为合适、或者说支持性更好。Golang 就是其中的佼佼者。

Golang 是一个轻量级、运行速度快、对高并发支持性极好的语言，很有力地支持了多核、多设备运行的场景。

Golang 在网络服务上，也具有强大的标准库。

目前，已有一个强大的微服务框架 —— **go-micro**，我们在这个系列中会用到它。

## protobuf/gRPC 简介

微服务被分割成多个独立的代码库，这就带来了一个重要的问题 —— 通信。在单体应用的系统中，你可以在代码库的任何地方调用想要的代码，所以不存在通信的问题。而微服务分布在不同的代码库中，不具备直接调用的能力。所以，你需要找到一个途径，使得不同服务之间可以尽可能低延迟地进行数据交互。

这里，我们可以采用传统的 REST 架构，例如通过 http 传输 JSON 或者 XML 。但这种方案会带来了一个问题：服务 A 把原始数据编码成 JSON/XML 格式，发送一长串字符给服务 B，B通过解码还原成原始数据。不过，当原始数据量很大时，这可能对通信造成严重影响。当我们和网络浏览器的通信时，只要约定了服务间的通信方式、固定了编码和解码方法，那么格式可以任意。

[gRPC](https://grpc.io/)就应运而生。gRPC 是一个由 Google 开发、基于 RPC 通信、轻量级的二进制传输协议。这个定义有点复杂，下面请由我细细道来。gRPC 核心数据格式采用的是二进制，而在上面 RESTful 的例子中，我们用的是 JSON 格式，也就是通过 http 发送一串字符串。字符串包括了它的编码格式、长度和其它占用字节的信息，所以总体数据量很大。基于客户端的字符串数据，服务器可以通知传统的浏览器，解析得到预期的数据。但在两个微服务的通信间，我们不需要字符串中的所有数据，所以我们采用难理解但更加轻量的二进制数据的二进制数据进行交互。gRPC 采用的是支持二进制数据的 HTTP 2.0 规范，这个规范还能支持双向的通信流，相当炫酷！HTTP 2 是 gRPC 工作的基础。如果你想进一步了解 HTTP 2，可以点击这个[Google链接](https://developers.google.com/web/fundamentals/performance/http2/)。

接下来的问题是，二进制数据该如何处理呢？不用担心，gRPC 有一个内部的数字模拟语言，叫 protobuf。Protobuf 支持自定义接口格式，对开发者很友好。

了解 gRPC 和 protobuf，我们准备实战，开始创建第一个服务的定义。首先，在代码库的根目录下创建如下文件 `consignment-service/proto/consignment/consignment.proto`。目前，为了让这个教程阅读起来更容易，我采用的是单一仓，就是把所有的服务都存放在一个代码库中。对使用单一仓很多争论和反对意见，这边我暂不深入探讨。当然，你在开发中最好把不同的服务和部件存放在分离的代码仓中，这种方式普遍受欢迎，但比较复杂。

在刚才创建的`consignment.proto`文件中，添加如下内容：

```protobuf
// consignment-service/proto/consignment/consignment.proto
syntax = "proto3";

package go.micro.srv.consignment; 

service ShippingService {  
  rpc CreateConsignment(Consignment) returns (Response) {}
}

message Consignment {  
  string id = 1;
  string description = 2;
  int32 weight = 3;
  repeated Container containers = 4;
  string vessel_id = 5;
}

message Container {  
  string id = 1;
  string customer_id = 2;
  string origin = 3;
  string user_id = 4;
}

message Response {  
  bool created = 1;
  Consignment consignment = 2;
}
```

这是个非常基础的定义的例子，不过还是有几点需要我们掌握。首先，你定义了服务内容，它应该包括你希望暴露给其他服务的方法。接着，你需要定义消息类型，这些数据结构体都非常简洁。正如上面的`Container`结构体，Protobuf 是一种可以自定义的静态类型。每个消息体都是他们自定义的类型。

现在已经使用到了两个库：消息通过 protobuf 处理；服务通过 gRPC 的 protobuf 插件处理，把消息编译成代码，从而进行交互，正如 proto 文件中的`service`部分。

protobuf 定义的结构，可以通过客户端接口，自动生成相应语言的二进制数据和功能。

说到这，我们就来一起给我们的服务创建一个 Makefile，路径如下`$ touch consignment-service/Makefile`。

```makefile
build:  
    protoc -I. --go_out=plugins=grpc:$(GOPATH)/src/github.com/ewanvalentine/shipper/consignment-service \
      proto/consignment/consignment.proto
```

这个 Makefile 会调用 protoc 库，将你的 protobuf 编译成对应的代码。同时，我们也指定了 gRPC 插件、编译目录和输出目录。

生成 Makefile 文件后，进入服务所在的文件夹，运行`$ make build`指令，然后你就能在`proto/consignment/`下看到一个名为`consignment.pb.go`的新 Go 文件。这里使用了 gRPC/protobuf 库，把自定义的 protobuf 结构自动转换成你想要的代码。

接下来，我们就可以正式搭建服务了。进入项目的根目录，创建一个文件 main.go `$ touch consignment-service/main.go`。

```go
// consignment-service/main.go
package main

import (  
    "log"
    "net"

        // 导入生成的 protobuf 代码
    pb "github.com/ewanvalentine/shipper/consignment-service/proto/consignment"
    "golang.org/x/net/context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

const (  
    port = ":50051"
)

type IRepository interface {  
    Create(*pb.Consignment) (*pb.Consignment, error)
}

// Repository - 一个模拟数据存储的虚拟仓库，以后我们会替换成真实的数据仓库
type Repository struct {  
    consignments []*pb.Consignment
}

func (repo *Repository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {  
    updated := append(repo.consignments, consignment)
    repo.consignments = updated
    return consignment, nil
}

// 服务需要实现所有在 protobuf 里定义的方法。
// 你可以参考 protobuf 生成的 go 文件中的接口信息。
type service struct {  
    repo IRepository
}

// CreateConsignment - 目前只创建了这个方法，包括 `ctx` (环境信息)和 `req`(委托请求)两个参数，会通过 gRPC 服务器进行处理
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment) (*pb.Response, error) {

    // 保存委托
    consignment, err := s.repo.Create(req)
    if err != nil {
        return nil, err
    }

    // 返回和 protobuf 中定义匹配的 `Response` 消息
    return &pb.Response{Created: true, Consignment: consignment}, nil
}

func main() {

    repo := &Repository{}

        // 启动 gRPC 服务器。
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer()

	//注册服务到 gRPC 服务器，会把已定义的 protobuf 与自动生成的代码接口进行绑定。
    pb.RegisterShippingServiceServer(s, &service{repo})

    // 在 gRPC 服务器上注册 reflection 服务。
    reflection.Register(s)
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

请仔细阅读代码中的注释，有助于你对这个服务的理解。简单来说，这些代码实现的功能是：在 50051 端口创建一个的 gRPC 服务器，通过 protobuf 生成的消息格式，实现 gRPC 接口交互的逻辑。就这样，你完成了一个完整功能的 gRPC 服务！你可以输入指令`$ go run main.go`来运行这个程序，不过，目前，从界面上你还看不到任何东西。那如何能直观看到这个 gRPC 服务器正常工作了呢？我们来一起创建个与它对接的客户端吧！

下面，我们来写一个命令行交互的程序，用来读取一个包含委托信息的 JSON 文件，和我们已创建的 gRPC 服务器交互。

进入根目录，输入命令行创建一个新的子文件夹`$ mkdir consignment-cli`。在文件夹中，创建一个新文件`cli.go`，代码如下：

```go
// consignment-cli/cli.go
package main

import (  
    "encoding/json"
    "io/ioutil"
    "log"
    "os"

    pb "github.com/ewanvalentine/shipper/consignment-service/proto/consignment"
    "golang.org/x/net/context"
    "google.golang.org/grpc"
)

const (  
    address         = "localhost:50051"
    defaultFilename = "consignment.json"
)

func parseFile(file string) (*pb.Consignment, error) {  
    var consignment *pb.Consignment
    data, err := ioutil.ReadFile(file)
    if err != nil {
        return nil, err
    }
    json.Unmarshal(data, &consignment)
    return consignment, err
}

func main() {  
    // 创建和服务器的一个连接
    conn, err := grpc.Dial(address, grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Did not connect: %v", err)
    }
    defer conn.Close()
    client := pb.NewShippingServiceClient(conn)

    // 和服务器通信，并打印出返回信息
    file := defaultFilename
    if len(os.Args) > 1 {
        file = os.Args[1]
    }

    consignment, err := parseFile(file)

    if err != nil {
        log.Fatalf("Could not parse file: %v", err)
    }

    r, err := client.CreateConsignment(context.Background(), consignment)
    if err != nil {
        log.Fatalf("Could not greet: %v", err)
    }
    log.Printf("Created: %t", r.Created)
}
```

同时，再创建一个委托信息文件 `consignment-cli/consignment.json`

```json
{
  "description": "This is a test consignment",
  "weight": 550,
  "containers": [
    { "customer_id": "cust001", "user_id": "user001", "origin": "Manchester, United Kingdom" }
  ],
  "vessel_id": "vessel001"
}
```

完成以上步骤后，在`consignment-service`下运行`$ go run main.go`，然后打开一个新的终端界面，运行`$ go run cli.go`，这时你就能看到一条消息`Created: true`。不过，如何我们才能确认，这个委托真正地生成了？让我们继续更新我们的服务，添加一个`GetConsignments`方法，能够看到所有已创建的委托。

首先需要更新我们的 proto 定义(我在修改部分添加了备注)

```protobuf
// consignment-service/proto/consignment/consignment.proto
syntax = "proto3";

package go.micro.srv.consignment;

service ShippingService {  
  rpc CreateConsignment(Consignment) returns (Response) {}

  // 创建一个新方法
  rpc GetConsignments(GetRequest) returns (Response) {}
}

message Consignment {  
  string id = 1;
  string description = 2;
  int32 weight = 3;
  repeated Container containers = 4;
  string vessel_id = 5;
}

message Container {  
  string id = 1;
  string customer_id = 2;
  string origin = 3;
  string user_id = 4;
}

// 创建一个空白的获取请求
message GetRequest {}

message Response {  
  bool created = 1;
  Consignment consignment = 2;

  // 增加一个数组，用来返回委托列表
  repeated Consignment consignments = 3;
}
```

我们成功地在服务上创建了一个叫`GetConsignments`和`GetRequest`的新方法，后者目前不含任何内容。我们也在回复的消息中，添加了`consignments`参数。你可能会注意到，该参数的类型前有个关键词：`repeated`。顾名思义，这代表该参数是以数组的方式保存的。

现在，让我们再次运行`$ make build`指令，并启动你的服务，你会看到一个类似`*service does not implement go_micro_srv_consignment.ShippingServiceServer (missing GetConsignments method)`的错误信息。

protobuf 库产生的接口在通信两端必须完全匹配，这是实现 gRPC 的基础。所以，我们需要确认 proto 的定义是否一致。

让我们更新下`consignment-service/main.go`文件：

```go
package main

import (  
    "log"
    "net"

    // 导入生成的 protobuf 代码
    pb "github.com/ewanvalentine/shipper/consignment-service/proto/consignment"
    "golang.org/x/net/context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

const (  
    port = ":50051"
)

type IRepository interface {  
    Create(*pb.Consignment) (*pb.Consignment, error)
    GetAll() []*pb.Consignment
}

// Repository - 一个模拟数据存储的虚拟仓库，以后我们会替换成真实的数据仓库
type Repository struct {  
    consignments []*pb.Consignment
}

func (repo *Repository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {  
    updated := append(repo.consignments, consignment)
    repo.consignments = updated
    return consignment, nil
}

func (repo *Repository) GetAll() []*pb.Consignment {  
    return repo.consignments
}

// 服务需要实现所有在 protobuf 里定义的方法。
// 你可以参考 protobuf 生成的 go 文件中的接口信息。
type service struct {  
    repo IRepository
}

// CreateConsignment - 目前只创建了这个方法，包括 `ctx` (环境信息)和 `req`(委托请求)两个参数，会通过 gRPC 服务器进行处理
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment) (*pb.Response, error) {

    // 保存委托
    consignment, err := s.repo.Create(req)
    if err != nil {
        return nil, err
    }

    // 返回和 protobuf 中定义匹配的 `Response` 消息
    return &pb.Response{Created: true, Consignment: consignment}, nil
}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest) (*pb.Response, error) {  
    consignments := s.repo.GetAll()
    return &pb.Response{Consignments: consignments}, nil
}

func main() {

    repo := &Repository{}

    // 启动 gRPC 服务器。
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer()

    //注册服务到 gRPC 服务器，会把已定义的 protobuf 与自动生成的代码接口进行绑定。
    pb.RegisterShippingServiceServer(s, &service{repo})

    // 在 gRPC 服务器上注册 reflection 服务。
    reflection.Register(s)
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

现在，我们引用了新的`GetConsignments`方法、更新了库和接口，也就满足了两边的 proto 定义一致。再次运行` go run main.go`后，服务能正常工作了。

再回到我们的客户端工具，我们通过调用`GetConsignments`这个方法，列出所有的委托：

```go
func main() {  
    ... 
    // ·...`表示和之前代码一致，这里不再重复

    getAll, err := client.GetConsignments(context.Background(), &pb.GetRequest{})
    if err != nil {
        log.Fatalf("Could not list consignments: %v", err)
    }
    for _, v := range getAll.Consignments {
        log.Println(v)
    }
}
```

在原先 main 函数中，找到打印`Created: success`日志的位置，在这之后添加上述代码，然后运行`$ go run cli.go`。程序就会创建一个委托，紧接着调用`GetConsignments`。当你运行次数越多，委托列表就会越来越长。

*注意：为了看起来简洁，我有时会用`...`来表示和之前的代码完全一致。之后几行新增的代码，需要手动添加到原代码中*

到这里，我们通过 protobuf 和 gRPC，完整地创建了一个微服务，以及一个与之交互的客户端。

本系列的下一章节将围绕着集成[go-micro](https://github.com/micro/go-micro)展开。go-micro 是一个基于微服务的、创建 gRPC 的强大框架。我们也会在下章创建第二个微服务 —— 容器服务。光说“容器”二字也许会令你困惑，这里具体指的是 Docker 中的“容器”概念。我们会在下一章探索微服务在 Docker 容器中的运行情况。

如果对此文有任何 bug、错误或者反馈，请直接联系[我的邮箱](ewan.valentine89@gmail.com)。

本教程所包含的代码库[链接](https://github.com/ewanvalentine/shippy)，用 `git`工具 checkout 分支`tutorial-1`第一章。第二章也将在近期更新。

编写本文花了我很长的时间以及大量的精力。如果你觉得这个系列有帮助，请考虑顺手打赏我(完全取决于你的意愿)。十分感谢！[https://monzo.me/ewanvalentine](https://monzo.me/ewanvalentine)

鸣谢：Microservices Newsletter (22nd November 2017)

----------------

via: https://ewanvalentine.io/microservices-in-golang-part-1/

作者：[Ewan Valentine](http://ewanvalentine.io/author/ewan)
译者：[Junedayday](https://github.com/Junedayday)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
