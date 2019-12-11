首发于：https://studygolang.com/articles/25301

# 我是如何在 Go 中组织包的

构建项目跟写代码一样具有挑战性。而且有很多种方法。使用错误的方法可能会让人很痛苦，但若要重构则又会非常耗时。另外，要想在一开始就设计出完美的程序几乎是不可能的。更重要的是，有些解决方法只适用于某特定大小的程序，但是程序的大小又是随着时间变化和增长的。所以我们的软件应该跟着出现过解决过的问题一起成长。

我主要从事微服务的开发，这种架构非常适合我。其他领域或其他基础架构的项目可能需要不同的方法。请在下面的评论中告诉我您的设计和最有意义的地方。

## 包及其依赖

在开发微服务时，按组件拆分服务很有用。每个组件都应该是独立的，理论上，如果需要，可以将其提取到外部服务。如何理解和实现呢？

假设我们有一个服务，它处理与订单相关的所有事情，比如发送电子邮件的确认、将信息保存到数据库、连接到支付提供商等。每个包都应该有一个名称，该名称清楚地说明了它的用途，并且遵守命名标准。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/how-i-organize-packages-in-go/organize-go.png)

这只是我们有 3 个包的项目的一个例子：**confemails**，**payproviders**和**warehouse**。包名应尽量简短并能让人一目了然。

每个包都有自己的 Setup()函数。该函数只接收能让该包运行的最基本的参数。例如，如果包对外提供 HTTP 服务，那么 Setup() 函数则仅需要接受一个类似 mux route 的 HTTP route。当包需要访问数据库时，Setup() 函数也是只接受 sql.DB 参数就可以了。当然，这个包也可能需要依赖另一个包。

## 包内的组成

知道了模块的外部依赖，下一步我们就可以专注于如何在模块内组织代码（包括相关依赖的处理）。在最开始，这个包包含以下文件: setup.go - 其中包含 Setup()函数, service.go - 它是逻辑文件, repository.go - 它是在读取/保存数据到数据的的文件。

Setup()函数负责构建模块的每个构建块，即服务、存储库、注册事件处理程序或 HTTP 处理程序等等。这是使用这种方法的实际生产代码的一个例子。

```go
func Setup(router *mux.Router, httpClient httpGetter, auth jwtmiddleware.Authorization, logger logger) {
	h := httpHandler{
		logger:        logger,
		requestClaims: jwtutil.NewHTTPRequestClaims(client),
		service:       service{client: httpClient},
	}
	auth.CreateRoute("/v1/lastAnswerTime", h.proxyRequest, http.MethodGet)
}
```

以上代码中,它构建了 JWT 中间件，这是一个处理所有业务逻辑(以及日志的位置)并注册 HTTP 处理程序的服务。正因为如此，模块是非常独立的，并且(理论上)可以转移到单独的微服务中，而不需要做太多工作。最后，所有的包都在 main 函数中配置。

有时，我们需要一些处理程序或数据库驱动。例如，一些信息可以被存储在数据库中，然后通过事件发送到平台的不同部分。使用像 saveToDb()这样的方法将数据只保存在同一个库中是很不方便的。所有类似的元素都应该由以下功能分割:repository_order.go 或 service_user.go。如果对象的类型超过 3 种，则将其移动到单独的子文件夹中。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/how-i-organize-packages-in-go/organizing-go-1.png)

## 测试

说到测试，我坚持一些原则。首先，在 Setup()函数中使用接口。这些接口应该尽可能小。在上面的例子中，有一个 httpGetter 接口。接口中只有**Get**()函数。

```go
type httpGetter interface {
 Get(url string) (resp *http.Response, err error)
}
```

谢天谢地，我只需要模拟一个方法。接口的定义需要尽可能地接近它的用途。

其次，尝试编写更少的测试用例的同时可以覆盖到更多的代码。对于每个主函数的决策/操作，一个成功的测试用例和一个失败的测试用例应该足够覆盖大约 80% 的代码。有时，程序中有一些关键部分，这部分可以被单独的测试用例覆盖。

最后，在以 `_test` 为后缀的单独包中编写测试，并将其放入模块中。把所有的东西都放在一个地方是很有用的。

当您想要测试整个应用程序时，请在主函数旁边的**setup**()函数中准备好每个依赖项。它将为生产环境和测试环境提供相同的设置，可以为您避免一些 bug。测试应该重用 setup()函数，并且只模拟那些不易模拟的依赖项(比如外部 api)。

## 总结

所有其他文件（比如 `.travis.yaml` 等）都保存在项目根目录中。这让我对整个项目有了一个清晰的认识。让我知道在哪里可以找到主文件，在哪里可以找到与基础结构相关的文件，并且没有混合在一起。否则，项目的主文件夹就会变得一团糟。

正如我在介绍中所说，我知道并非所有项目都能从中受益，但是像 microservices 这样的小型程序会发现它非常有用。

---

via: https://developer20.com/how-i-organize-packages-in-go/

作者：[Bartłomiej Klimczak](https://developer20.com/about-me/index.html)
译者：[shadowstorm97](https://github.com/shadowstorm97)
校对：[DingdingZhou](https://github.com/DingdingZhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
