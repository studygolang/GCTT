# 七年后我如何编写Go HTTP服务

Mat Ryer
2018年5月9日

自从[r59](https://medium.com/@matryer)（一个1.0版之前的版本）以来，我一直在编写Go（即Golang，如果不特殊说明的话），并且在过去的七年里一直使用Go构建HTTP API和服务。

在[Machine Box](https://machinebox.io/?utm_source=matblog-3May2018&utm_medium=matblog-3May2018&utm_campaign=matblog-3May2018&utm_term=matblog-3May2018&utm_content=matblog-3May2018)，我的大多数技术工作都涉及构建各种API。机器学习很复杂，大多数开发人员都无法直接访问，因此我的工作是通过API接口来提供简单的调用方式，到目前为止我们收到了很好的反馈。

>如果您还没有亲眼目睹过Machine Box开发者的体验，[请试一试](https://machinebox.io/docs/facebox/teaching-facebox)，并告知我知道您的想法。

我编写服务的方式多年来发生了变化，所以我想分享我现在编写服务的方式 -希望我的模式对您和您的工作有用。

## 服务器结构

所有组件都有一个`server`结构，通常看起来像这样：

```go
type server struct {
    db * someDatabase
    router * someRouter
    email EmailSender
}
```

- 共享依赖项是结构的字段

## 路由文件

在每个组件中都有一个文件`routes.go`，里面包含所有的路由信息：

```go
package app
func (s *server) routes() {
    s.router.HandleFunc("/api/", s.handleAPI())
    s.router.HandleFunc("/about", s.handleAbout())
    s.router.HandleFunc("/", s.handleIndex())
}
```

这很方便，因为大多数代码维护都是从URL和错误报告开始的 - 所以只需一眼就`routes.go`可以指示我们在哪里查看。

## 处理程序

服务器中的的HTTP处理程序：

```go
func (s *server) handleSomething() http.HandlerFunc { ... }
```

处理程序可以通过`s`变量访问依赖项。

## 处理程序返回结果

处理函数实际上并不处理请求，它们返回一个函数。

这给了我们一个闭包环境，我们的处理程序可以在其中运行

```go
func (s *server) handleSomething() http.HandlerFunc {
    thing := prepareThing()
    return func(w http.ResponseWriter, r *http.Request) {
        // use thing
    }
}
```

该`prepareThing`只调用一次，所以你可以用它做一次处理程序前的初始化，然后用`thing`在处理程序重处理请求。

确保只读取共享数据，如果处理程序正在修改任何内容，请记住您需要一个互斥锁或其他东西来保护它。

## 处理程序获取特定依赖项的参数

如果特定处理程序具有依赖项，请将其作为参数。

```go
func (s *server) handleGreeting(format string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, format, "World")
    }
}
```

处理程序可以访问该`format`变量。

## HandlerFunc 优先于 Handler

我现在几乎在每一个案例中都是用`http.HandlerFunc`，而不是`http.Handler`。

```go
func (s *server) handleSomething() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ...
    }
}
```

它们或多或少是可以互换的，所以只需选择哪个更易于阅读。对我来说，就是`http.HandlerFunc`。

## 中间件只是Go函数

中间件函数接受`http.HandlerFunc`并返回一个可以在原始处理程序之前和/或之后运行代码的新函数-或者它可以决定不调用原始处理程序。

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

处理程序内部的逻辑可以选择是否调用原始处理程序-在上面的示例中，如果`IsAdmin`是`false`，则处理程序将返回`HTTP 404 Not Found`并返回(abort); 注意并没有调用`h`处理程序。

如果`IsAdmin`是`true`，则将执行传递给传入的`h`处理程序。

通常我在`routes.go`文件中列出了中间件：

```go
package app
func (s *server) routes() {
    s.router.HandleFunc("/api/", s.handleAPI())
    s.router.HandleFunc("/about", s.handleAbout())
    s.router.HandleFunc("/", s.handleIndex())
    s.router.HandleFunc("/admin", s.adminOnly(s.handleAdminIndex()))
}
```

## 请求和响应类型也可以在那里

如果端点有自己的请求和响应类型，通常它们仅对特定处理程序有用。

如果是这种情况，您可以在函数内定义它们。

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

这会对您的包空间进行整理，并允许您将这些类型命名为相同，而不必考虑特定于处理程序的版本。

在测试代​​码中，您只需将类型复制到测试函数中并执行相同的操作即可。要么…

## 测试类型可以帮助构建测试

如果您的请求/响应类型隐藏在处理程序中，您只需在测试代码中声明新类型即可。

这是一个为需要了解您的代码的后来者做一些讲解的机会。

例如，假设`Person`我们的代码中有一个类型，我们在许多端点上重用它。如果我们有一个`/greet`端点，我们可能只关心他们的名字，所以我们可以在测试代码中表达：

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

从这个测试中可以清楚地看出，我们唯一关心的就是`Name`字段。

## 使用sync.Once设置依赖项

如果在准备处理程序时我必须做任何昂贵的事情，我会推迟到第一次调用该处理程序时进行处理。

这改善了应用程序的启动时间

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

`sync.Once`确保代码只执行一次，其他调用（其他人发出相同的请求）将阻塞，直到完成。

- 错误检查在`init`函数之外，所以如果出现问题我们仍然会出现错误，并且不会在日志中丢失错误
- 如果没有调用处理程序，则永远不会执行昂贵的工作-这可能会带来很大的好处，具体取决于代码的部署方式

> 请记住，执行此操作时，您将初始化时间从启动时移动到运行时（首次访问端点时）。我经常使用Google App Engine，所以这对我来说很有意义，但是你的情况可能会有所不同，所以值得考虑在何时何地使用sync.Once。

## 服务器是可测试的

我们的服务器类型非常容易测试的。

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
    srv.ServeHTTP(w, req)
    is.Equal(w.StatusCode, http.StatusOK)
}
```

- 在每个测试中创建一个服务器实例 - 如果昂贵的东西延迟加载，这将不会花费太多时间，即使对于大组件
- 通过在服务器上调用ServeHTTP，我们正在测试整个堆栈，包括路由和中间件等。如果你想避免这种情况，你当然可以直接调用处理程序方法。
- 使用`httptest.NewRecorder`记录什么处理程序在做
- 此代码示例使用我的[is](https://godoc.org/github.com/matryer/is)测试迷你框架（作为Testify的迷你替代品）

## 结论

我希望本文中涉及的内容有意义，并帮助您完成工作。如果您不同意或有其他想法，请发[推特给我](https://twitter.com/matryer)。

----------------

via:<https://medium.com/statuscode/how-i-write-go-http-services-after-seven-years-37c208122831>

作者：[Gleicon Moraes](https://github.com/gleicon)
译者：[lovechuck](https://github.com/lovechuck)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出