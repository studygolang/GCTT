首发于：https://studygolang.com/articles/22835

# 停止写破坏性(Broken) Go 库

不久前我和朋友们想出一个主意，准备合并我们的 IRC bots，并用 Go 重写它们。为了防止重写大部分现有功能，我们试图找到支持 bots 程序中使用的 `Web API` 的现有库。我们的项目需要一个 Reddit API 的库。这篇文章启发于我找到的前三个库，我不打算说出它们的名字，以免羞辱它们的作者。

上面说的每一个库都存在一些基本问题以至于它们在真实场景中不可用。并且每个库都以这样一种方式编写：不以非向后兼容的方式修改现有库的 API，这样是不可能修复问题的。不幸的是，由于很多其他的库也存在同样的问题，所以我会在下面列出一些作者错误的地方。

# 不要对 `HTTP` 客户端硬编码

很对库都包含了对 `http.DefaultClient` 的硬编码。虽然对库本身来说这并不是问题，但是库的作者并未理解应该怎样使用 `http.DefaultClient` 。正如 `default client` 建议它只在用户没有提供其他 `http.Client` 时才被使用。相反的是，许多库作者乐意在他们代码中涉及 `http.DefaultClient` 的部分采用硬编码，而不是将它作为一个备选。这会导致在某些情况下这个库不可用。

首先，我们很多人都读过这篇讲述 `http.DefaultClient` 不能自定义超时时间的文章《[Don’t use Go’s default HTTP client (in production)](https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779)》，当你没法保证你的`HTTP` 请求一定会完成（或者至少要等一个完全无法预估时间的响应）时，你的程序可能会遇到奇怪的 goroutine 泄漏和一些无法预知的行为。在我看来，这会使每一个对 `http.DefaultClient` 采用硬编码的库不可用。

其次，网络需要一些额外的配置。有时候需要用到代理，有时候需要对 `URL` 进行一丢丢的改写，甚至可能 `http.Transport` 需要被一个定制的接口替换。当一个程序员在你的库里用他们自己的 `http.Client` 实例时，以上这些都很容易被实现。

在你的库中处理 `http.Client` 的推荐方式是使用提供的客户端，但是如果需要的话，有一个默认的备选：

```go
func CreateLibrary(client *http.Client) *Library {
    if client == nil {
        client = http.DefaultClient
    }
    ...
}
```

或者如果你想从工厂函数中移除参数，请在你的 `struct` 中定义一个辅助方法，并且让用户在需要时设置其属性：

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

另外，如果一些全局的特性对于每个请求来讲都是必须的，人们经常感觉到需要用他们自己的实例来替换 `http.Client`。这是一个错误的方法  —  如果你需要在你的请求中设置一些额外的 `headers`，或者在你的客户端引入某类公共的特性，你只需要简单为每个请求进行设置或者用组装定制客户端的方式来代替完全替换它。

# 不要引入全局变量

另一个反面模式是允许用户在一个库中设置全局变量。举个例子，在你的库中允许用户设置一个全局的 `http.Client` 并被所有的 `HTTP` 调用执行：

```go
var libraryClient *http.Client = http.DefaultClient

func SetHttpClient(client *http.Client) {
    libraryClient = client
}
```

通常在一个库中不应该存在一堆全局变量。当你写代码的时候，你应该想想用户在他们的程序中多次使用你的这个库会发生什么。全局变量会使不同的参数没有办法被使用。而且，在你的代码中引入全局变量会引起测试上的问题并造成代码上不必要的复杂度。使用全局变量可能会导致在你程序的不同模块有不必要的依赖。在写你的库的时候，避免全局状态是格外重要的。

# 返回 structs，而不是 interfaces

这是一个普遍的问题（实际上我在这一点上也犯过错）。很多库都有下面这类函数：

```go
func New() LibraryInterface {
    ...
}
```

在上面的 case 中，返回一个 interface 使 struct 的特性在库里被隐藏了。实际上应该这么写：

```go
func New() *LibraryStruct {
    ...
}
```

在库里不应该存在接口的声明，除非它被用在某个函数参数中。如果出现上面的 case，你就应该想想你在写这个库的时候的约定。当返回一个 interface 时，你基本上得声明一系列可用的方法。如果有人想用这个接口来实现他们自己的功能(比如说为了测试)，他得打乱他们的代码来添加更多的方法。这意味着尽管在 struct 里添加方法是安全的，但在 interface 里不是。这个想法在这篇文章中被总结得很好《[Accept Interfaces Return Struct in Go](https://mycodesmells.com/post/accept-interfaces-return-struct-in-go)》。这个方案也能解决配置的问题。你想修改库中的一些特性，你可以简单的修改 struct 中一些公开的字段。但是如果你的库只提供给用户一个 interface，这就玩不转了。

详情请参见 Go  [http.Client](https://golang.org/pkg/net/http/#Client)。

# 使用配置结构体来避免修改你的APIs

另一种配置方法是在你的工厂函数中接收一个配置结构体，而不是直接传配置参数。你可以很随意的添加新的参数而不用破坏现有的 API。你只需要做一件事情，在Config结构体中添加一个新的字段，并且确保不会影响它原本的特性。

```go
func New(config Config) *LibraryStruct {
    ...
}
```

下面是一种添加结构体字段的正确的场景，如果一个用户初始化结构体的时候忘了添加字段名，这是一种我认为修改他们的代码能得到原谅的场景。为了维护兼容性，你应该在你的代码中用  `person{name: "Alice", age: 30}` 而不是 `person{"Alice", 30}`。

你能在 [golang.org/x/crypto](https://godoc.org/golang.org/x/crypto/openpgp#Sign) 包里看到对上面的补充。总之，对配置来说，我认为允许用户在返回的结构体里设置不同的参数是一个更好的方法，并且只在编写复杂方法时才使用这种特定方法。

# 总结

根据经验来讲，在写一个库的时候，你应该总是允许用户指定他们自己的  `http.Client` 来执行 `HTTP` 调用。而且考虑到未来迭代修改带来的影响，你可以尝试用可扩展的方式编写代码。避免全局变量，库不能存储全局状态。如果你有任何疑问-参考标准库是怎么写的。

我认为有一个很好的想法，在你的程序中用你的库来测试并问自己一些问题：

- 如果你尝试多次引入库会发生什么？
- 你的库有没有单元测试？
- 在不破坏原有代码的前提下，有没有一种非侵入式的方式来扩展你的库？
- 在不破坏原有代码的前提下，是否可以添加额外配置参数？

---

via: https://0x46.net/thoughts/2018/12/29/go-libraries/

作者：[Filip Borkiewicz](https://0x46.net/)
译者：[Alihanniba](https://github.com/Alihanniba)
校对：[zhoudingding](https://github.com/dingdingzhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
