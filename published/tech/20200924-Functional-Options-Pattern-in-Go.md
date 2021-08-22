首发于：https://studygolang.com/articles/34446

# Go 函数选项模式

作为 Golang 开发者，遇到的许多问题之一就是尝试将函数的参数设置成可选项。这是一个十分常见的场景，您可以使用一些已经设置默认配置和开箱即用的对象，同时您也可以使用一些更为详细的配置。

对于许多编程语言来说，这很容易。在 C 语言家族中，您可以提供具有同一个函数但是不同参数的多个版本；在 PHP 之类的语言中，您可以为参数提供默认值，并在调用该方法时将其忽略。但是在 Golang 中，上述的做法都不可以使用。那么您如何创建具有一些其他配置的函数，用户可以根据他的需求（但是仅在需要时）指定一些额外的配置。

有很多的方法可以做到这一点，但是大多数方法都不是尽如人意，要么需要在服务端的代码中进行大量额外的检查和验证，要么通过传入他们不关心的其他参数来为客户端进行额外的工作。

下面我将会介绍一些不同的选项，然后为其说明为什么每个选项都不理想，接着我们会逐步构建自己的方式来作为最终的干净解决方案：函数选项模式。

让我们来看一个例子。比方说，这里有一个叫做 `StuffClient` 的服务，它能够胜任一些工作，同时还具有两个配置选项（超时和重试）。

```go
type StuffClient interface {
    DoStuff() error
}

type stuffClient struct {
    conn    Connection
    timeout int
    retries int
}
```

这是个私有的结构体，因此我们应该为它提供某种构造函数：

```go
func NewStuffClient(conn Connection, timeout, retries int) StuffClient {
    return &stuffClient{
        conn:    conn,
        timeout: timeout,
        retries: retries,
    }
}
```

嗯，但是现在我们每次调用 `NewStuffClient` 函数时都要提供 `timeout` 和 `retries`。因为在大多数情况下，我们只想使用默认值，我们无法使用不同参数数量带定义多个版本的 NewStuffClient ，否则我们会得到一个类似 `NewStuffClient redeclared in this block` 编译错误。

一个可选方案是创建另一个具有不同名称的构造函数，例如：

```go
func NewStuffClient(conn Connection) StuffClient {
    return &stuffClient{
        conn:    conn,
        timeout: DEFAULT_TIMEOUT,
        retries: DEFAULT_RETRIES,
    }
}
func NewStuffClientWithOptions(conn Connection, timeout, retries int) StuffClient {
    return &stuffClient{
        conn:    conn,
        timeout: timeout,
        retries: retries,
    }
}
```

但是这么做的话有点蹩脚。我们可以做得更好，如果我们传入了一个配置对象呢:

```go
type StuffClientOptions struct {
    Retries int //number of times to retry the request before giving up
    Timeout int //connection timeout in seconds
}
func NewStuffClient(conn Connection, options StuffClientOptions) StuffClient {
    return &stuffClient{
        conn:    conn,
        timeout: options.Timeout,
        retries: options.Retries,
    }
}
```

但是，这也不是很好的做法。现在，我们总是需要创建 `StuffClientOption` 这个结构体，即使不想在指定任何选项时还要传递它。另外我们也没有自动填充默认值，除非我们在代码中的某处添加了一堆检查，或者也可以传入一个 `DefaultStuffClientOptions` 变量（不过这么做也不好，因为在修改某一处地方后可能会导致其他的问题。）

所以，更好的解决方法是什么呢？解决这个难题最好的解决方法是使用函数选项模式，它利用了 Go 对闭包更加方便的支持。让我们保留上述定义的 `StuffClientOptions` ，不过我们仍需要为其添加一些内容。

```go
type StuffClientOption func(*StuffClientOptions)
type StuffClientOptions struct {
    Retries int //number of times to retry the request before giving up
    Timeout int //connection timeout in seconds
}
func WithRetries(r int) StuffClientOption {
    return func(o *StuffClientOptions) {
        o.Retries = r
    }
}
func WithTimeout(t int) StuffClientOption {
    return func(o *StuffClientOptions) {
        o.Timeout = t
    }
}
```

泥土般芬芳, 不是吗？这到底是怎么回事？基本上，我们有一个结构来定义 `StuffClient` 的可用选项。 另外，现状我们还定义了一个叫做 `StuffClientOption` 的东西（次数是单数），它只是接受我们选项的结构体作为参数的函数。我们还定义了另外两个函数 `WithRetries` 和 `WithTimeout` ，它们返回一个闭包，现在就是见证奇迹的时刻了！

```go
var defaultStuffClientOptions = StuffClientOptions{
    Retries: 3,
    Timeout: 2,
}
func NewStuffClient(conn Connection, opts ...StuffClientOption) StuffClient {
    options := defaultStuffClientOptions
    for _, o := range opts {
        o(&options)
    }
    return &stuffClient{
        conn:    conn,
        timeout: options.Timeout,
        retries: options.Retries,
    }
}
```

现在，我们定义了一个额外和包含默认选项的没有导出的变量，同时我们已经调整了构造函数，用来接收[可变参数](https://gobyexample.com/variadic-functions)。然后, 我们遍历 `StuffClientOption` 列表(单数)，针对每一个列表，将列表中返回的闭包使用在我们的 `options` 变量（需要记住，这些闭包接收一个 `StuffClientOptions` 变量，仅需要在选项的值上做出少许修改）。

现在我们要做的事情就是使用它！

```go
x := NewStuffClient(Connection{})
fmt.Println(x) // prints &{{} 2 3}
x = NewStuffClient(
    Connection{},
    WithRetries(1),
)
fmt.Println(x) // prints &{{} 2 1}
x = NewStuffClient(
    Connection{},
    WithRetries(1),
    WithTimeout(1),
)
fmt.Println(x) // prints &{{} 1 1}
```

这看起来相当不错，已经可以使用了！而且，它的好处是，我们只需要对代码进行很少的修改，就可以随时随地添加新的选项。

把这些修改放在一起，就是这样：

```go
var defaultStuffClientOptions = StuffClientOptions{
    Retries: 3,
    Timeout: 2,
}
type StuffClientOption func(*StuffClientOptions)
type StuffClientOptions struct {
    Retries int //number of times to retry the request before giving up
    Timeout int //connection timeout in seconds
}
func WithRetries(r int) StuffClientOption {
    return func(o *StuffClientOptions) {
        o.Retries = r
    }
}
func WithTimeout(t int) StuffClientOption {
    return func(o *StuffClientOptions) {
        o.Timeout = t
    }
}
type StuffClient interface {
    DoStuff() error
}
type stuffClient struct {
    conn    Connection
    timeout int
    retries int
}
type Connection struct {}
func NewStuffClient(conn Connection, opts ...StuffClientOption) StuffClient {
    options := defaultStuffClientOptions
    for _, o := range opts {
        o(&options)
    }
        return &stuffClient{
            conn:    conn,
            timeout: options.Timeout,
            retries: options.Retries,
        }
}
func (c stuffClient) DoStuff() error {
    return nil
}
```

如果你想自己尝试一下，请在 [Go Playground](https://play.golang.org/p/VcWqWcAEyz) 上查找。

但这也可以通过删除 `StuffClientOptions` 结构体进一步简化，并将选项直接应用在我们的 `StuffClient` 上。

```go
var defaultStuffClient = stuffClient{
    retries: 3,
    timeout: 2,
}
type StuffClientOption func(*stuffClient)
func WithRetries(r int) StuffClientOption {
    return func(o *stuffClient) {
        o.retries = r
    }
}
func WithTimeout(t int) StuffClientOption {
    return func(o *stuffClient) {
        o.timeout = t
    }
}
type StuffClient interface {
    DoStuff() error
}
type stuffClient struct {
    conn    Connection
    timeout int
    retries int
}
type Connection struct{}
func NewStuffClient(conn Connection, opts ...StuffClientOption) StuffClient {
    client := defaultStuffClient
    for _, o := range opts {
        o(&client)
    }
    client.conn = conn
    return client
}
func (c stuffClient) DoStuff() error {
    return nil
}
```

从[这里](https://play.golang.org/p/Z5P5Om4KDL)就能够开始尝试。在我们的示例中，我们只是将配置直接应用于结构体中，如果中间有一个额外的结构体是没有意义的。但是，请注意，在许多情况下，您可能仍然想使用上一个示例中的 `config` 结构。例如，如果您的构造函数正在使用 `config` 选项执行某些操作时，但是并没有将它们存储到结构体中，或者被传递到其他地方，配置结构的变体是更通用的实现。

感谢 [Rob Pike](https://commandcenter.blogspot.de/2014/01/self-referential-functions-and-design.html) 和 [Dave Cheney](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis) 推广这种设计模式。

---
via: https://halls-of-valhalla.org/beta/articles/functional-options-pattern-in-go,54/

作者：[ynori7](https://halls-of-valhalla.org/beta/user/ynori7)
译者：[sunlingbot](https://github.com/sunlingbot)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
