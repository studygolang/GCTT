# 停止写垃圾库吧

不久前我和朋友们想出一个主意，准备合并我们的IRC bots，并用go重写它们。为了防止重写大部分现有功能，我们试图查找支持在bots中使用的现有的web API的库。我们的项目需要一个Reddit API的库。这篇文章启发于我找到的前三个库，我不打算说出它们的名字，以免羞辱它们的作者。

上面说的每一个库都存在一些根本问题以至于它们在真实场景中不可用。此外每个库都是以向后兼容的方式迭代编写，这是不可能解决问题的。不幸的是，由于很多其他的库也存在同样的问题，所以我会在下面列出一些作者错误的地方。

### 不要对`HTTP`客户端硬编码

很对库都包含了对 `http.DefaultClient` 的硬编码。虽然对库本身来说这并不是问题，但是库的作者并不理解 `http.DefaultClient` 到底是怎么被使用的。正如默认的客户端建议它只在用户没有提供其他 `http.Client` 时才被使用。相反的是，许多库作者乐意在他们的代码中涉及 `http.DefaultClient` 的部分采用硬编码，而不是将它作为一个备选。这会导致在某些情况下这个库不可用。

首先，我们很多人都读过这篇讲述 `http.DefaultClient`不能自定义超时时间的文章《[Don’t use Go’s default HTTP client (in production](https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779))》，当你没法保证你的`HTTP` 请求一定会完成(或者至少要等一个完成无法预估时间的响应) 时，你的程序可能会遇到奇怪的 goroutine 泄漏和一些无法预知的行为。在我看来，这会使每一个对`http.DefaultClient`采用硬编码的库不可用。

其次，网络需要一些额外的配置。有时候需要用到代理，有时候需要对`URL`进行一丢丢的改写，甚至可能`http.Transport`需要被一个定制的接口替换。当一个程序员在你的库里用他们自己的`http.Client` 实例时，以上这些都很容易被实现。

在你的库中处理`http.Client` 的推荐方式是使用提供的客户端，但是如果需要的话，有一个默认的备选：

```go
func CreateLibrary(client *http.Client) *Library {
    if client == nil {
        client = http.DefaultClient
    }
    ...
}
```

或者如果你想从工厂函数中移除参数就在你的`struct`中定义一个辅助方法，并且让用户随时设置它的属性：

```go
type Library struct {
    Client *http.Client
}

func (l *Library) getClient() *http.Client {
    if l.Client == nil {
        return http.DefaultClient
    }
    return l.Client
}
```

另外，如果一些全局的特性对于每个请求来讲都是必须的，人们经常感觉到需要用他们自己的实例来替换`http.Client`。这是一个错误的方法-如果你需要在你的请求中设置一些额外的`headers`，或者在你的客户端引入某类公共的特性，你只需要简单的为每个请求进行设置或者用组装定制客户端的方式来代替完全的替换它。

### 不要引入全局变量

补充一个反面模式，允许用户在一个库中设置全局变量。举个例子，在你的库中允许用户设置一个全局的`http.Client`并被所有的`HTTP calls`执行：

```go
var libraryClient *http.Client = http.DefaultClient

func SetHttpClient(client *http.Client) {
    libraryClient = client
}
```

通常在一个库中不应该存在一堆全局变量。当你写代码的时候，你应该想想用户在他们的程序中多次使用你的这个库会发生什么。全局变量会使得无法使用不同的参数。而且，在你的代码中引入全局变量会引起测试上的问题并造成代码上不必要的复杂度。使用全局变量可能会导致在你程序的不同模块有不必要的依赖。在写你的库的时候，避免全局状态是格外重要的。