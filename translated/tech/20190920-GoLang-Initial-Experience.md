# GoLang 初体验

我最近决定在一个新项目中使用 GoLang 来实现一个 CRUD API。近来，我较为熟悉 Java，Groovy，了解一些 Python。

我大部分的经验都是使用 Java 或者 Groovy 加上 Spring Boot。这让我感到有些无聊，所以为什么不来学点儿东西找找乐子呢？

## 要求

以下是一些要求。

* 设计并实现领域数据模型
* 实现 CRUD API
* 以 Mongo 数据库作为后端
* 必须有 Swagger 文档 API 定义并且方便用多种语言生成客户端
* 运行在 Docker 容器中
* 能被部署在 Kubernetes 中

## 非功能性要求

* 需要能很容易的调用其它语言的 API
* 需要能够快速迭代（可能要突破常规）
* 必须有单元测试

## 加分项

* 内存消耗上保守
  * 这对于如果我想在一个内存被限制的环境中，如一个 512 MB 内存的树莓派上运行程序这种情况来说很重要。
* 要有趣也要有学习体验

## 使用的模块和库

| 作用                                                         | 模块           |
| ------------------------------------------------------------ | -------------- |
| 访问数据库                                                   | mongo-go-drive |
| 路由                                                         | go-chi         |
| REST API JSON Patch( 译者注：[RFC6902](http://tools.ietf.org/html/rfc6902) 和 [RFC7396](https://tools.ietf.org/html/rfc7396)) | json-patch     |
| 单元测试                                                     | testify        |
| Swagger API 定义                                             | go-swagger     |

## 优点

Go 语言与 C 和 Java 十分的相像。有 C 和 Java 的基础能很容易的熟练掌握 Go 语言，完成一个入门项目。

我特别喜欢 Go 代码的简单明了。

公平来说，我也喜欢样板代码尽可能少的，高度集成化的框架。我就十分喜欢 Java 11+ 或者 Groovy 与 Spring Boot，Spring Data，Lombok项目以及可能的 Spring Data REST 的联合使用。当然，有时候 Spring Boot Data REST 的魔法有点儿过犹不及了。

Go 的 'defer' 关键字可以说是我最喜欢的特性之一了。推迟一些操作直到函数退出才执行这一特性，在关闭资源并记录函数退出动作的日志方面十分有用。

## 不同点

**错误处理有点儿繁琐。**

错误处理对于 Java 背景的人来说有些不同。我发现在 Go 中它需要更明确。

在 Java 中，一个方法能抛出一个异常，捕获一个或多个异常，忽略它们（这样做可能是错误的），或者重新抛出给调用者来处理。Go 需要使用调用方法的固定模式，然后判断是否有错误发生。我们可以讨论下这样做好不好。

我发现Go的错误检测和传递需要一点调整时间而且有点繁琐，但肯定是可行的。

```go
// 我经常在代码中看到这样的模式
obj1, err := doohickey.doSomething(someArg)
if err !=nil {
	log.Println("doohickey.doSomething got error error: ", err)
	return
}
obj2, err2 := widget.doSomethingElse(otherArg)
if err2 !=nil {
	log.Println("Widget doSomethingElse returned error: ", err2)
	return
}
//...
```

**JSON 响应类型以及映射到结构体**

对于 Go，JSON 和 静态类型，我发现 Go 在如何处理动态 JSON 和将其解析为结构体方面有些困惑和笨拙。

这在 Groovy 和 Python 中相当容易，他们完全可以动态的把 JSON 转换成其他东西的映射。

在Go中，将JSON反序列化为一个结构并将其序列化回来，这与其他语言中的做法并没有本质上的不同。

## 成熟度进展

实际上，我在这方面并没有发现它有什么不好的。正相反，我发现了正常的应该被期待的东西。因为 Go 仍然是一门相对比较新的语言，在一些领域它正在迎头赶上。

**Go 依赖和版本化模块库**

因为以前使用过依赖管理和构建工具，如 Java 的 Gradle 和 Maven，自然而然的我就想 Go 有同种水平的依赖管理。

在我写本文时，[GoLang 1.13](https://golang.org/doc/go1.13) 支持谷歌的模块代理，是这样说的：

>从 Go 1.13开始，go 命令在默认情况下将使用由 Google 运行的 Go 模块镜像和 Go 检验和数据库来进行模块的下载和认证。参看 [https://proxy.golang.org/privac](https://proxy.golang.org/privacy) 来了解有关这些服务的隐私信息，参看 [go 命令文档](https://golang.org/cmd/go/#hdr-Module_downloading_and_verification) 了解详细配置，包括怎么停止使用这些服务或者使用另外的服务。如果你依赖于一个不公开的模块，参见 [环境配置文档](https://golang.org/cmd/go/#hdr-Module_configuration_for_non_public_modules)

## 结束语

