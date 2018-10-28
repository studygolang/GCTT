首发于：https://studygolang.com/articles/14333

# 我这几年来是如何编写 Go HTTP 服务的

我是从 [r59](https://golang.org/doc/devel/pre_go1.html#r59) —— 1.0 之前的一个发布版本，就开始写 Go 了，并且在过去七年里一直在用 Go 构建 HTTP API  和服务。

在 [Machine Box](https://machinebox.io/?utm_source=matblog-3May2018&utm_medium=matblog-3May2018&utm_campaign=matblog-3May2018&utm_term=matblog-3May2018&utm_content=matblog-3May2018) 里,我大部分的技术性工作涉及到构建各种各样的 API。 机器学习本身很复杂而且大部分开发者也不会用到，所以我的工作就是通过 API 终端来简单阐述一下，目前来说反响很不错。

> 如果你还没有看过 Machine Box 开发者的经验, [请尝试一下](https://machinebox.io/docs/facebox/teaching-facebox) 并让我知道你的意见。

我编写服务的方法已经在过去几年中发生了变化，所以我打算分享目前我编写服务的经验——也许这些方法能帮到你和你的工作。

## 一个 server 结构体

我所有的组件都有一个单独的  `server`  结构体，它通常都是类似于下面这种形式：

```go
type server struct {
    db     *someDatabase
    router *someRouter
    email  EmailSender
}
```

-公共组件是该结构体的字段。

## routes.go

在每个组件中我有个一个唯一的文件  `routes.go` ，在这里所有的路由都能运行。

```go
package app
func (s *server) routes() {
    s.router.HandleFunc("/api/", s.handleAPI())
    s.router.HandleFunc("/about", s.handleAbout())
    s.router.HandleFunc("/", s.handleIndex())
}
```

这样会很方便，因为大部分代码维护都开始于一个 URL 和一个被报告的错误——所以只要浏览一下 `routes.go` 就能引导我们到目的地。

## 挂起服务器的 handler

我的 HTTP handler 挂起服务器：

```go
func (s *server) handleSomething() http.HandlerFunc { ... }
```

handler 可以通过 s 这个server变量来访问依赖项。

## 返回 handler

我的 handler 函数不会处理请求，它们返回的函数完成处理工作。

这样会提供给我们一个 handler 可以运行的封闭环境。

```go
func (s *server) handleSomething() http.HandlerFunc {
    thing := prepareThing()
    return func(w http.ResponseWriter, r *http.Request) {
        // use thing
    }
}
```

`prepareThing` 函数只会被调用一次，所以你可以用它来完成每个 handler 的一次性的初始化工作，然后在 handler 里面使用  `thing`  。

请确保只会对共享数据执行读操作，如果 handler 改写了共享数据，记住你需要用锁或者其他机制来保护共享数据。

## 为 handler 专有的依赖传递参数

如果一个特别的 handler 有一个依赖项，就把这个依赖项当作参数。

```go
func (s *server) handleGreeting(format string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, format, "World")
    }
}
```

`format` 变量可以被 handler 访问。

## 用 HandlerFunc 代替 Handler

现在我几乎在每个用例中都会使用 `http.HandlerFunc` ，而不是 `http.Handler` 。

```go
func (s *server) handleSomething() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ...
    }
}
```

两者之间大都可以互换，所以觉得哪个便于阅读就选哪个即可。对我来说，`http.HandlerFunc` 更加适合。

## 中间件仅仅只是 Go 函数

中间件函数接受一个  `http.HandlerFunc`  并且返回一个新的 HandlerFunc , 该 handler 可以在调用初始 handler 之前或之后运行代码——抑或它可以决定是否调用初始的 handler 。

```go
func (s *server) adminOnly(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !currentUser(r).IsAdmin {
            http.NotFound(w, r)
            return
        }
        h(w, r)
    }
}
```

这个 handler 内部的逻辑可以选择性的决定是否调用初始 handler——在上面的例子里，如果  `IsAdmin` 为 `false`,该 handler 就会返回一个 HTTP  `404 Not Found` 并且返回（中止）；注意，`h` handler 没有被调用。

如果 `IsAdmin` 为 `true`, 就会运行到 `h` handler。

通常我会把中间件放到 `routes.go`  文件中：

```go
package app
func (s *server) routes() {
    s.router.HandleFunc("/api/", s.handleAPI())
    s.router.HandleFunc("/about", s.handleAbout())
    s.router.HandleFunc("/", s.handleIndex())
    s.router.HandleFunc("/admin", s.adminOnly(s.handleAdminIndex))
}
```

## request 和 response 类型也可以放在那里

如果终端有它自身的 request 和 response 类型的话，通常这些类型只对特定的 handler 有用。

假设一个例子，你可以把它们定义在函数内部。

```go
func (s *server) handleSomething() http.HandlerFunc {
    type request struct {
        Name string
    }
    type response struct {
        Greeting string `json:"greeting"`
    }
    return func(w http.ResponseWriter, r *http.Request) {
        ...
    }
}
```

这样就可以解放包的空间，并允许你把这种类型都定义成同样的名字，从而免去了特定的 handler 考虑命名。

在测试代码时，你可以直接复制这些类型到你的测试函数中并执行同样的操作。或者其他……

## 测试类型有助于架构测试的框架

如果你的 request/response 类型都隐藏在 handler 内部，那么你可以在测试代码中直接定义新的类型。

这就有机会做一些解释性的工作，以便让未来的接任者能够理解你的代码。

举个例子，我们假设代码中一个 `Person` 类型存在，并且在很多终端都会重用它。如果我们有一个 `/greet` 终端，这时可能只关心它的 `Name`，所以可以在测试代码中这样表述：

```go
func TestGreet(t *testing.T) {
    is := is.New(t)
    p := struct {
        Name string `json:"name"`
    }{
        Name: "Mat Ryer",
    }
    var buf bytes.Buffer
    err := json.NewEncoder(&buf).Encode(p)
    is.NoErr(err) // json.NewEncoder
    req, err := http.NewRequest(http.MethodPost, "/greet", &buf)
    is.NoErr(err)
    //... more test code here
```

从测试代码中可以清晰的看出，我们只关心 `Person`  的  `Name` 字段。

## sync.Once 组织依赖

如果我不得不为准备 handler 时执行一些代价高昂的操作，我就会把它们推迟到第一次调用 handler 的时刻。

这样可以改善应用的启动时间。

```go
func (s *server) handleTemplate(files string...) http.HandlerFunc {
    var (
        init sync.Once
        tpl  *template.Template
        err  error
    )
    return func(w http.ResponseWriter, r *http.Request) {
        init.Do(func(){
            tpl, err = template.ParseFiles(files...)
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        // use tpl
    }
}
```

`sync.Once`  确保了代码只会运行一次，如果有其它调用（其他人发起同样的请求）就会堵塞，直到代码结束为止。

- 错误检查放在了 `init`  函数外面，所以如果出现错误的话我们依然可以捕获它，而且不会在日志中遗失 。
- 如果 handler 没有被调用过，这些代价高昂的操作就永远不会发生——这可以对你的代码部署有极大好处。

记住这一点，上面是把初始化的时间从启动时刻移到了运行时刻（当端点第一次被访问到时）。我使用 Google App Engine 很久了，对我来说这种操作是可以理解的，但对你自身来说可能就未必了。所以你有必要思考何时何地值得用 `sync.Once` 这种方式。

## server 必须易于测试

我们的 server 类型需要能够简单测试。

```go
func TestHandleAbout(t *testing.T) {
    is := is.New(t)
    srv := server{
        db:    mockDatabase,
        email: mockEmailSender,
    }
    srv.routes()
    req, err := http.NewRequest("GET", "/about", nil)
    is.NoErr(err)
    w := httptest.NewRecorder()
    srv.ServeHTTP(w, r)
    is.Equal(w.StatusCode, http.StatusOK)
}
```

- 在每组测试中创建一个 server 实例——如果把代价高昂的操作延迟加载，这就不会花费太多时间，即使是对大型组件也依然有效。
- 通过调用服务器上的 ServerHTTP ，我们会测试到整个栈，包括路由和中间件等。当然了，如果希望避免这种情况的话，你也可以直接调用 handler 函数。
- 使用 `httptest.NewRecorder` 来记录 handler 所执行的操作。
- 这份代码示例使用到了我的 [一个正在测试中的微框架](https://godoc.org/github.com/matryer/is) （一个验证用的简易可选项）

## 结论

我希望文章中涵盖到的内容可能对你有些用处，能帮助到你的工作。如果你有不同意见或其它想法的话， [请联系我们](https://twitter.com/matryer) 。

---

via: https://medium.com/statuscode/how-i-write-go-http-services-after-seven-years-37c208122831

作者：[Mat Ryer](https://medium.com/@matryer)
译者：[sunzhaohao](https://github.com/sunzhaohao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
