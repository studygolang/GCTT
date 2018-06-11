已发布：https://studygolang.com/articles/11897

# 6 款最棒的 Go 语言 Web 框架简介

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/top-6-web-frameworks-for-go-as-of-2017/twitter_status.jpg)
https://twitter.com/ThePracticalDev/status/930878898245722112

如果你只是想写一个自己用的小网站，或许你不需要框架，但如果你是要开发一个投入生产运营的网站，那么你肯定会需要一个框架，而且是需要一个好的 Web 框架。

如果你已经掌握所有必要的知识和经验，你会冒险自己去重新开发所有的功能么？你有时间去找满足生产级别要求的库来用于开发么？另外，你确定这个库可以满足你后续所有的要求？

这些都是促使我们去使用框架的原因，哪怕是那些最牛的开发者也不会一直想要重新造轮子，我们可以站在前人的肩膀上，走得更快更好。

## 介绍

[Go](https://golang.org) 是一门正在快速增长的编程语言，专为构建简单、快速且可靠的软件而设计。 点击 [此处](https://github.com/golang/go/wiki/GoUsers) 查看有哪些优秀的公司正在使用 Go 语言来驱动他们的业务。

本文将会提供一切必要的信息来帮助开发人员了解更多关于使用 Go 语言来开发 Web 应用程序的最佳选择。

本文包含了最详尽的框架比较，从流行度、社区支持及内建功能等多个不同角度出发做对比。

**Beego**：_开源的高性能 Go 语言 Web 框架。_

* [https://github.com/astaxie/beego](https://github.com/astaxie/beego)
* [https://beego.me](https://beego.me)

**Buffalo**：_使用 Go 语言快速构建 Web 应用。_

* [https://github.com/gobuffalo/buffalo](https://github.com/gobuffalo/buffalo)
* [https://gobuffalo.io](https://gobuffalo.io)

**Echo**：_简约的高性能 Go 语言 Web 框架。_

* [https://github.com/labstack/echo](https://github.com/labstack/echo)
* [https://echo.labstack.com](https://echo.labstack.com)

**Gin**：_Go 语言编写的 Web 框架，以更好的性能实现类似 Martini 框架的 API。_

* [https://github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)
* [https://gin-gonic.github.io/gin](https://gin-gonic.github.io/gin)

**Iris**：_全宇宙最快的 Go 语言 Web 框架。完备 MVC 支持，未来尽在掌握。_

* [https://github.com/kataras/iris](https://github.com/kataras/iris)
* [https://iris-go.com](https://iris-go.com)

**Revel**：_Go 语言的高效、全栈 Web 框架。_

* [https://github.com/revel/revel](https://github.com/revel/revel)
* [https://revel.github.io](https://revel.github.io)

## 流行度

> 按照流行度排行（根据 GitHub Star 数量）

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/top-6-web-frameworks-for-go-as-of-2017/github_star.jpg)
[https://github.com/speedwheel/awesome-go-web-frameworks/blob/master/README.md#popularity](https://github.com/speedwheel/awesome-go-web-frameworks/blob/master/README.md#popularity)

## 学习曲线

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/top-6-web-frameworks-for-go-as-of-2017/learn.jpg)
[https://github.com/speedwheel/awesome-go-web-frameworks/blob/master/README.md#learning-curve](https://github.com/speedwheel/awesome-go-web-frameworks/blob/master/README.md#learning-curve)

*astaxie* 和 *kataras* 分别为 **Beego** 和 **Iris** 做了超棒的工作，希望其他的框架也能迎头赶上，为开发者提供更多的例子。至少对于我来说，如果我要切换到一个新的框架，那些例子就是最丰富的资源，来获取尽可能多的有用信息。一个实例胜千言啊。

## 核心功能

> 根据功能支持的多寡排行

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/top-6-web-frameworks-for-go-as-of-2017/core_feature.jpg)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/top-6-web-frameworks-for-go-as-of-2017/core_feature2.jpg)
[https://github.com/speedwheel/awesome-go-web-frameworks/blob/master/README.md#core-features](https://github.com/speedwheel/awesome-go-web-frameworks/blob/master/README.md#core-features)

> 几个知名的 Go 语言 Web 框架并不是真正意义上的框架，也就是说：
**Echo**，**Gin** 和 **Buffalo** 并不是真正意义上的 Web 框架（因为没有完备支持所有功能）但是大部分的 Go 社区认为它们是的，因此这些框架也可以和 **Iris**，**Beego** 或 **Revel** 做比较。所以，我们有义务将这几个框架（**Echo**，**Gin** 和 **Buffalo**）也列在这个表中。
>
> 以上所有这些框架，除了 **Beego** 和 **Revel** 之外，都可以适配任意 `net/http` 中间件。其中一部分框架可以轻松地做适配，另外一些可能就需要额外的努力 [即使这里的痛苦不是一定的]。

## 技术性词汇

### 路由：命名的路径参数和通配符

可以处理动态的路径。

命名的路径参数例子：

```
"/user/{username}" 匹配 "/user/me"，"/user/speedwheel" 等等
```

上面路径参数 `username` 的值分别是 `"me"` 和 `"speedwheel"`。

通配符的例子：

```
"/user/{path *wildcard}" 匹配
"/user/some/path/here",
"/user/this/is/a/dynamic/multi/level/path" 等等
```

上面的路径参数 `path` 对应的分别是 `"some/path/here"` 和 `"this/is/a/dynamic/multi/level/path"`。

> **Iris** 也支持一个叫 `macros` 的功能，它可以被表示为 `/user/{username:string}` 或者 `/user/{username:int min(1)}`。

### 路由：正则表达式

过滤动态的路径。

例如：

```
"/user/{id ^[0-9]$}" 能匹配 "/user/42" ，但不会匹配 "/user/somestring"
```

这里的路径参数 `id` 的值为 `42`。

### 路由：分组

通过共用逻辑或中间件来处理有共同前缀的路径组。

例如:

```go
myGroup := Group("/user", userAuthenticationMiddleware)
myGroup.Handle("GET", "/", userHandler)
myGroup.Handle("GET", "/profile", userProfileHandler)
myGroup.Handle("GET", "/signup", getUserSignupForm)
```

* /user
* /user/profile
* /user/signup

你甚至可以从一个组中创建子分组：

```go
myGroup.Group("/messages", optionalUserMessagesMiddleware)
myGroup.Handle("GET', "/{id}", getMessageByID)
```

* /user/messages/{id}

### 路由：上述所有规则相结合而没有冲突

这是一个高级且有用的的功能，我们许多人都希望路由模块或 Web 框架能支持这点，但目前，在 Go 语言框架方面，只有 **Iris** 能支持这一功能。

这意味着类似如 `/{path *wildcard}` ， `/user/{username}` 和 `/user/static` 以及 `/user/{path *wildcard}` 等路径都可以在同一个路由中通过静态路径（`/user/static`）或通配符（`/{path *wildcard}`）来正确匹配。

### 路由：自定义 HTTP 错误

指可以自行处理请求错误的情况。 Http 的错误状态码均 `>=400` ，比如 `NotFound 404`，请求的资源不存在。

例如：

```go
OnErrorCode(404, myNotFoundHandler)
```

上述的大多数 Web 框架只支持 `404`，`405` 及 `500` 错误状态的处理，但是例如 `Iris，Beego 和 Revel` 等框架，它们完备支持 HTTP 错误状态码，甚至支持 `any error` 任意错误。（ `any error` -- 任意错误，只有 **Iris** 能够支持）。

### 100% 兼容 net/http

这意味著：

* 这些框架能够让你直接获取 `*http.Request` 和 `http.ResponseWriter` 的所有相关信息。
* 各框架提供各自相应处理 `net/http` 请求的方法。

### 中间件生态系统

框架会为你提供一个完整的引擎来定义流程、全局、单个或一组路由，而不需要你自己用不同的中间件来封装每一部分的处理器。框架会提供比如 Use（中间件）、Done（中间件） 等函数。

### 类 Sinatra 的 API 设计（译者注：[Sinatra](http://sinatrarb.com) 是一门基于 Ruby 的[领域专属语言](https://en.wikipedia.org/wiki/Domain-specific_language)）

可以在运行时中注入代码来处理特定的 HTTP 方法 （以及路径参数)。

例如:

```go
.Get or GET("/path", gethandler)
.Post or POST("/path", postHandler)
.Put or PUT("/path", putHandler) and etc.
```

### 服务器程序：默认启用 HTTPS

框架的服务器支持注册及自动更新 SSL 证书来管理新传入的 SSL/TLS 连接 (https)。 最著名的默认启用 https 的供应商是 [letsencrypt](https://letsencrypt.org/)。

### 服务器程序：平滑关闭（Gracefully Shutdown）

当按下 `CTRL + C` 关闭你的终端应用程序时，服务器将等待 (一定的等待时间)其他的连接完成相关任务或触发一个自定义事件来做清理工作（比如：关闭数据库），最后平滑地停止服务。

### 服务器程序：多重监听

框架的服务器支持自定义的 `net.Listener` 或可以启动一个有多个 http 服务和地址的 Web 应用。

### 完全支持 HTTP/2

框架可以很好地支持处理 https 请求的 HTTP/2 协议，并且支持服务器 `Push` 功能。

### 子域名

你可以直接在你的 Web 应用中注入子域名的路径。

`辅助功能（secondary）` 意味着这个功能并不被这个框架原生支持，但是你仍旧可以通过启用多个 http 服务器来实现。这样做的缺点在于：主程序和子域名程序之间并不是连通的，默认情况下，它们不能共享逻辑。

### 会话（Sessions）

支持 http sessions，且可以在自定义的处理程序中使用 sessions。

* 有一些 Web 框架支持后台数据库来储存 sessions，以便在服务器重启之后仍旧能获得持久的 sessions。
* **Buffalo** 使用 [gorilla 的 sessions 库](https://github.com/gorilla/sessions)，它比其他框架的实现略微慢了一点。

例如:

```go
func setValue(context http_context){
    s := Sessions.New(http_context)
    s.Set("key", "my value")
}

func getValue(context http_context){
    s := Sessions.New(http_context)
    myValue := s.Get("key")
}

func logoutHandler(context http_context){
    Sessions.Destroy(http_context)
}
```

Wiki: [https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol#HTTP_session](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol#HTTP_session)

### 网络套接字（Websockets）

框架支持 websocket 通信协议。不同的框架对于这点有各自不同的实现方式。

你应该通过它们的例子来看看哪个适合你。我的一个同事，在试过了上述所有框架中的 websocket 功能之后告诉我：**Iris** 实现了最多的 websocket 特性，并且提供了相对更容易使用的 API 。

Wiki: [https://en.wikipedia.org/wiki/WebSocket](https://en.wikipedia.org/wiki/WebSocket)

### 程序内嵌对视图（又名模版）的支持

通常情况下，你必须根据 Web 应用的可执行文件一一对应地转换模版文件。内嵌到应用中意味着这个框架集成了 [go-bindata](https://github.com/jteeuwen/go-bindata) ，因此在最终的可执行文件中可以以 `[]byte` 的形式将模版包含进来。

#### 什么是视图引擎

框架支持模版加载、自定义及内建模版功能，以此来节省我们的开发时间。

### 视图引擎：STD

框架支持标准的 `html/template` 解析器加载模版。

### 视图引擎：Pug

框架支持 `Pug` 解析器加载模版。

### 视图引擎：Django

框架支持 `Django` 解析器加载模版。

### 视图引擎：Handlebars

框架支持 `Handlebars` 解析器加载模版。

### 视图引擎：Amber

框架支持 `Amber` 解析器加载模版。

### 渲染：Markdown, JSON, JSONP, XML...

框架提供一个简单的方法来发送和自定义各种内容类型的响应。

### MVC

Model–view–controller (MVC) 模型是一种用于在计算机上实现用户界面的软件架构模式，它将一个应用程序分为互相关联的三部分。这样做的目的是为了：将信息的内部处理逻辑、信息呈现给用户以及从用户获取信息三者分离。MVC 设计模式将这三个组件解耦合，从而实现高效的代码复用和并行开发。

* **Iris** 支持完备的 MVC 功能, 可以在运行时中注入。
* **Beego** 仅支持方法和数据模型的匹配，可以在运行时中注入。
* **Revel** 支持方法，路径和数据模型的匹配，只可以通过生成器注入（生成器是另外一个不同的软件用于构建你的 Web 应用）。

Wiki: [https://en.wikipedia.org/wiki/Model%E2%80%93view%E2%80%93controller](https://en.wikipedia.org/wiki/Model%E2%80%93view%E2%80%93controller)

### 缓存

Web 缓存（或 http 缓存）是一种用于临时存储（缓存）网页文档，如 HTML 页面和图像，来减缓服务器延时。一个 Web 缓存系统缓存网页文档，使得后续的请求如果满足特定条件就可以直接得到缓存的文档。Web 缓存系统既可以指设备，也可以指软件程序。

Wiki: [https://en.wikipedia.org/wiki/Web_cache](https://en.wikipedia.org/wiki/Web_cache)

### 文件服务器

可以注册一个（物理的）目录到一个路径，使得这个路径下的文件可以自动地提供给客户端。

### 文件服务器：内嵌入应用

通常情况下，你必须将所有的静态文件（比如静态资产，assets：CSS，JavaScript 文件等）与应用程序的可执行文件一起传输。支持此项功能的框架为你提供了在应用中，以 `[]byte` 的形式，内嵌所有这些数据的机会。由于服务器可以直接使用这些数据而无需在物理位置查找文件，它们的响应速度也将更快。

### 响应可以在发送前的生命周期中被多次修改

目前只有 **Iris** 通过 http_context 中内建的的响应写入器（response writer）支持这个功能。

当框架支持此功能时，你可以在返回给客户端之前检索、重置或修改状态码、正文（body）及头部（headers）。默认情况下，在基于 `net/http` 的 Web 框架中这是不可能的，因为正文和状态码一经写定就不能被检索或修改。

### Gzip

当你在一个路由的处理程序中，并且你可以改变响应写入器（response writer）来发送一个用 gzip 压缩的响应时，框架会负责响应的头部。如果发生任何错误，框架应该把响应重置为正常，框架也应该能够检查客户端是否支持 gzip 压缩。

> gzip 是用于压缩和解压缩的文件格式和软件程序。

Wiki: [https://en.wikipedia.org/wiki/Gzip](https://en.wikipedia.org/wiki/Gzip)

### 测试框架

可以使用框架特定的库，来帮助你轻松地编写更好的测试代码来测试你的 HTTP 。

例如（目前仅 **Iris** 支持此功能）：

```go
func TestAPI(t *testing.T) {
    app := myIrisApp()
    tt := httptest.New(t, app)
    tt.GET("/admin").WithBasicAuth("name", "pass").Expect().
    Status(httptest.StatusOK).Body().Equal("welcome")
}
```

`myIrisApp` 返回你虚构的 Web 应用，它有一个针对 `/admin` 路径的 GET 方法，它有基本的身份验证逻辑保护。

上面这个简单的测试，用 `"name"` 和 `"pass"` 通过身份验证并访问 GET `/admin` ，检查它的响应状态是否为 `Status OK`，并且响应的主体是否为 `"welcome"` 。

### TypeScript 转译器

TypeScript 的目标是成为 ES6 的超集。除了标准定义的所有新特性外，它还增加了静态类型系统。TypeScript 还有转换器用于将 TypeScript 代码（即 ES6 + 类型）转换为 ES5 或 ES3 JavaScript 代码，如此我们就可以在现今的浏览器中运行这些代码了。

### 在线编辑器

在在线编辑器的帮助下，你可以快速轻松地在线编译和运行代码。

### 日志系统

自定义日志系统通过提供有用的功能，如彩色日志输出、格式化、日志级别分离及不同的日志记录后端等，来扩展原生日志包。

### 维护和自动更新

以非侵入的方式通知框架的用户即时更新。

## 再见！

谢谢你的阅读，如果你喜欢这篇文章，请用表情符号和我互动哦 :)

---

via: https://dev.to/speedwheel/top-6-web-frameworks-for-go-as-of-2017-34i

作者：[Edward Marinescu](https://dev.to/speedwheel)
译者：[rxcai](https://github.com/rxcai)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
