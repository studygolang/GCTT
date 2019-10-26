首发于：https://studygolang.com/articles/24267

# 在 Go 中使用微服务架构的好处

## 前言
我们已经讨论“微服务架构”很长一段时间了。它是软件架构中最新的热门话题。那么什么是微服务呢？我们为什么要使用它？为什么要在 Golang 中使用微服务架构?它有哪些优点？

本文中，我将会探讨一些相关的问题。废话不多说，让我们开始吧。

## 什么是微服务？
微服务是一种软件开发技术，属于 SOA（面向服务的架构）的一种形式。它的作用是，将应用程序构建为许多松耦合的服务的集合。在这种架构中，服务的编码通常是细粒度的，服务的协议更轻量。目前还没有对微服务的准确定义，但它有一些显著的特征：自动化部署、业务功能、去中心化的数据管理和智能端点。
（译者注：对于这个定义有兴趣的同学，不妨去看看 Martin 大神[关于微服务的文章](https://martinfowler.com/articles/microservices.html)

## 我们为什么使用微服务？
这种架构有助于我们用各部分、小型模块描绘整个应用程序，使其更容易理解、开发和测试；有助于我们将各个服务视为独立且又清晰指明其用途的服务。更进一步地，它有助于保持项目架构的一致性(最初设计的架构和实际开发完成的架构差别不大)。它还可以通过建立不同的独立团队来进行服务的部署和扩展，从而各团队能够并行地开发。在这个架构中重构代码更容易。它也支持连续交付和部署流程（CI/CD)。

## 为什么使用 go 构建微服务？
在深入研究这个问题之前。首先，我说一下 Golang 的优势。虽然 Golang 是一门新的语言，但是与其他语言相比，它有很多优势。用 Golang 编写的程序更加健壮。它们能够承受程序使用运行的服务构建的繁重负载。Golang 更适合多处理器系统和 web 应用程序。此外，它容易地与 GitHub 集成，管理非集中的代码包。微服务架构的用处大部分体现在当程序需要伸缩（scalable)时。如果有一种语言可以完全符合标准,那么它就是 Golang。原因是它继承自 C-family 编程语言，用 Golang 编写的组件更容易与同一家族中其他语言编写的组件相结合。

尽管 Go 出身于 C-family，但它比 C / C ++更高效。 它语法更简单，有点像 Python。它稳定语法, 自第一次公开发布以来，它没有太大变化，也就是说它是后向兼容的。与其他语言相比，这让 golang 占了上风。 除此之外，Golang 的性能比 python 和 java 高出不少。锦上添花的是，它又像 C/C++ 简单的同时又易于阅读和理解，使它成为开发微服务应用的绝佳选择。

## Golang中的微服务架构框架
下面，我们讨论一下可以用于微服务架构的框架。有以下些框架:

### Go Micro
Go Micro 是目前为止我遇到的最流行的RPC框架。它是一个可插拔的RPC框架。Go Micro 为我们提供了以下功能:

* 服务发现: 程序自动注册到服务发现系统
* 负载均衡: 它提供了客户端负载均衡，这有助于平衡服务实例之间的请求
* 同步通信: 提供 Request/Response 传输层
* 异步通信: 具有内置的发布和订阅功能
* 消息编码: 可以利用 header 中 Content-Type 进行编码和解码
* RPC客户端/服务器端: 利用上述功能并提供构建微服务需要的接口

![](https://camo.githubusercontent.com/9057599d2bc2d3c79c43423521d71f4ea0851457/68747470733a2f2f6d6963726f2e6d752f646f63732f696d616765732f676f2d6d6963726f2e737667)

Go Micro 架构由三层组成。第一层抽象为服务层。第二层为 client-server 模型层。serrver 用于编写服务的块组成，而 client 为我们提供接口，其唯一目的是向 server model 中编写的服务发出请求。

第三层有以下类型的插件:
* Broker: 在异步通信中为 message broker(消息代理）提供接口
* Codec: 用于加密或解密消息
* Registry: 提供服务搜索功能
* Selector: 在 register 上构建了负载均衡
* Transport: Transport是服务与服务之间同步请求/响应的通信接口

它还提供了一个名为 Sidecar 的功能。Sidecar 使您能够集成以Go以外的语言编写的服务。它还为我们提供了gRPC编码/解码、服务注册和HTTP 请求处理

### GO Kit
Go Kit 是一个用于构建微服务的编程工具包。与 Go Micro不同，它是一个可以以二进制包导入的库。Go Kit 规则很简单。如下:

* 没有全局变量
* 声明式组合
* 显式依赖
* Interface as Contracts (接口合约)
* 领域驱动设计（DDD)

Go Kit 提供以下代码包:

* Authentication 鉴权: BasicAuth 和 JWT
* Transport 协议: HTTP, gRPC 等
* Logging 日志: 服务中的结构化日志接口
* Metrics 度量: CloudWatch,Statsd, Graphite等
* Tracing 分布式追踪: Zipkin and Opentracing
* Service discovery 服务发现: Consul, Etcd, Eureka等
* Circuitbreaker 限流熔断: Hystrix 在 Go 语言的实现

Go Kit 服务架构如下

## Gizmo
Gizmo 是来自《纽约时报》的一个微服务工具包。它提供了将服务器守护进程和 pubsub 守护进程放在一起的包。它公开了以下包:

* Server: 提供两个服务器实现: SimpleServer(HTTP)和 RPCServer(gRPC)
* Server/kit: 基于Go Kit的实验代码包
* Config 配置: 包含来自 JSON文件、Consul k/v 中的 JSON blob 或环境变量的配置功能
* Pubsub: 提供用于从队列中发布和使用数据的通用接口
* Pubsub/pubsubtest: 包含发布者和订阅者接口的测试实现
* Web: 用于从请求查询和有效负载解析类型的外部函数

Pubsub包提供了处理以下队列的接口:

* pubsub/aws: 用于 Amazon SNS/SQS
* pubsub/gcp: 用于 Google Pubsub
* pubsub/kafka: 用于 Kafka topics
* pubsub/http: 用户 HTTP 推送

所以，在我看来，Gizmo 介于 Go Micro 和 Go Kit 之间。它不像 Go Micro 那样是一个完全的黑盒。同时，它也不像 Go Kit 那么原始。它提供了更高级别的构建组件，比如配置和 pubsub 包

## Kite
Kite 是一个在 Go 中开发微服务的框架。它公开RPC client 和 Server 端代码包。创建的服务将自动注册到服务发现系统 Kontrol。Kontrol 是用 Kite 编写的，它本身就是一个 Kite service。这意味着 Kite 微服务在自身的环境中运行良好。如果需要将 Kite 微服务连接到另一个服务发现系统，则需要定制。这是我从列表中选择 Kite 并决定不介绍这个框架的重要原因之一

因此，如果您觉得这个博客很有用，并且想知道如何在 Golang 中创建多功能的微服务，那么请从我们这里[雇佣 Golang开发者](https://www.bacancytechnology.com/hire-golang-developer?source=post_page---------------------------)，并学习利用顶级的专业知识。

---
via: https://medium.com/datadriveninvestor/benefits-of-using-microservice-architecture-in-go-4440e4fcd9c9

作者：[Katy Slemon](https://medium.com/@katyslemon)
译者：[Alex1996a](https://github.com/Alex1996a)
校对：[dingdingzhou](https://github.com/zhoudingding)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出