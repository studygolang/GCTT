已发布：https://studygolang.com/articles/12154

# Go 语言中的包装一个微服务样板

首先呢，祝大家新年快乐 :tada::tada::tada: 全年无 BUG！

应用的复杂性在很多方面都在增长，诸如可扩展性、开发、测试以及部署。在企业级开发中，那种老式的大型单一架构看起来已经过时了。在我工作的众多公司中，都希望系统是通过简单的插件组合在一起的方式构建的。这就是为什么许多数公司都基于微服务架构来开发他们的产品。目前有 Netflix（译者注：美国流媒体巨头、世界最大的收费视频网站）、PayPal、Amazon、eBay 以及 Twitter 等少数几家公司正在使用微服务。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/wrap-micro-service/1.gif)

在这里呢我是不太想去介绍微服务的基础架构的。如果你还不太了解什么是微服务，我建议你去看看 [Martin Fowler](https://martinfowler.com/articles/microservices.html)（译者注：世界顶级的专家，现为 Thought Works 公司的首席科学家）的文章。在阅读过一些文章并且尝试使用了一些工具进行了测试之后，最后我们将使用 Go 语言来开发我们大多数的微服务。对 Go 语言进行标准的测试以及和其它语言的比较则需要另外的文章来介绍了，在这里我也就不多介绍了！

通过谷歌能很快的找到许多用 Go 语言开发的框架和示例程序，但是它们通常都非常复杂。所以我们创建一个属于自己的！我会来介绍用它来包装我们微服务软件。

```
go get github.com/spf13/viper
```

Viper 包是一个 Go 语言下的全方位配置解决方案包含了`十二要素应用程序`（译者注：一种应用开发理论）。它被设计成和应用程序一起运行，几乎能处理所有的配置需求及格式。它支持如下类型：

- 默认设置
- 读取 JSON, TOML, YAML, HCL, 和 Java 的 properties 配置文件
- 在线读取以及重载配置文件（可选）
- 读取环境变量
- 读取远程系统配置（etcd 或 Consul），并观察其改动
- 读取命令行标记
- 读取缓冲区
- 直接设置确定的值

Viper 可以被看作是你所有应用程序都需要的一个配置注册器。

```
go get github.com/Sirupsen/logrus
```
Logrus 包是一个 Go 语言的结构化日志记录器，和 Go 语言的标准库（ logger ） 安全兼容。

```
go get github.com/nats-io/go-nats
```
NATS 消息系统的一个 Go 客户端。

```
go get github.com/gorilla/mux
```
`gorilla/mux` 包是一个路由调度工具，它会对传入的请求进行匹配并分别进行处理。

这个 `mux` 名字代表 `"HTTP request multiplexer"`。和标准库中的 `http.ServerMux` 类似，`mux.Router` 将传入的请求和已注册的路由列表进行匹配，然后调起匹配到的 URL 或其它的路由处理程序。其主要的特点包括：

- 它实现了 `http.Handler` 接口，因此它能和标准库 `http.ServerMux` 兼容。
- 它能匹配的请求包括 `URL HOST` （主机）、`path`（路径），`path prefix` （地址前缀）、`http schemes` ，头信息以及请求的参数值，还有 HTTP 请求的方法（GET/POST）或 用户自定义匹配。
- 主机，路径还有请求的参数值都有一个可选的正则表达式变量与之对应。
- 已注册的 URLs 可以帮助建立正向或反向到资源的引用。
- 路由可以被当作子路由来使用：只有在父路由匹配的情况下才能进行嵌套路由测试。
- 这对于定义像主机，路径，地址前缀或其它有相同属性的通用共享条件的路由组非常有用。还有个意外惊喜，这样能优化请求的匹配。

```
go get github.com/urfave/negroni
```
`Negroni` 是 Go 语言中在 Web 里面惯用的一个中间件。它非常小，无侵入设计，并且鼓励使用 `net/httpHandlers`。

如果你喜欢 [Martini](https://github.com/go-martini/martini)（一个非常新的 Go 语言的 Web 框架）的思想，但又觉得它包含了太多魔术方法，那么 `Negroni` 将是一个不错的选择。

```
go get github.com/Masterminds/squirrel
```

`Squirrel` 可以帮你构建一个可以组合的 SQL 语句。

```
go get github.com/garyburd/redigo/redis
```

`Redigo` 是一个 Go 语言下的 Redis 数据库客户端。

特点：
- 一个类似 Print 一样的接口支持所有的 Redis 命令。
- 管道，包含了管道事务。
- 支持发布/订阅（Publish和Subscribe）。
- 支持连接池。
- Script 辅助类型能对 EVALSHA 方便的进行使用。
- 辅助函数专门来处理命令的应答。


```
go get github.com/dgrijalva/jwt-go
```
一个 Go （Golang 对搜索引擎更加友好）实现的 JSON Web Tokens （JWT 一个非常轻巧的规范）。

```
go get github.com/asaskevich/govalidator
```
一个对字符串、结构体和集合进行验证和过滤的包。基于 `validator.js`。

```
go get github.com/DavidHuie/gomigrate
```
一个 Golang 的数据库迁移工具包。

## 结束语

我的意思并不是鼓励所有人都来使用微服务架构！当然了，它有优点也有不足，但对于那些有过充分研究并且决定用来实现微服务应用的来说：我希望你能在这篇帖子中找到有用的信息 ，当然不要犹豫，不要踌躇，有问题请留言。

----------------

via: http://alimrz.com/2018/01/02/golang-microservice-boilerplate/

作者：[ALI M.MIRZAEE](http://alimrz.com/about/)
译者：[zhuCheer](https://github.com/zhuCheer)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出