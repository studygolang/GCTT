已发布：https://studygolang.com/articles/12153

# 测试 Go 语言 Web 应用

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/testing-web-app/cover.jpg)

我利用闲暇的时间尝试着用 Go 来写一个网站小应用。在 Go 标准库中有一些非常棒的包可以在 Web 应用开发中使用并且我非常喜欢使用它们。实际上，在 Go 官方的 wiki 中有一个编写 Web 应用的小教程。但是，却没有提及如何用标准库去测试 Web 应用，而且也没有搜索到比较好的一个方案。

本着试一试的心态在我自己的项目中做了一些尝试，我发现了测试 Go 语言中 Web 应用的关键，就是通过使用高级的函数来实现依赖注入。

## 依赖注入

通过[依赖注入](https://en.wikipedia.org/wiki/Dependency_injection)我们希望能达到可以支持我们所有的功能要求这样一个目标。

然而，刚开始不太清楚怎么做。大多数的教程都会像这样写一个 Web 应用：

```go
package main

import (
    "fmt"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
```
来源: [编写一个 Web 应用 - Golang Wiki](https://golang.org/doc/articles/wiki/)

并不是说这是一个不好的处理方法，但是我们想要知道当添加一个数据库时发生了什么呢？或者一个外部的包要操作我们的 sessions ，像 [Gorilla Session](https://github.com/gorilla/sessions) 那样么？通常，人们会实例化一个数据库来管理或者操作一个全局的 session 并且将有效期设置成一天。但这样当你试图测试时将会给你带来麻烦。最好是编写一个函数来为你创建相关的操作。

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/markberger/database"
)

type AppDatabase interface {
    GetBacon() string
}

func homeHandler(db AppDatabase) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hi there, I love %s!", db.GetBacon())
    })
}

func main() {
    db := database.NewDatabase()
    http.HandleFunc("/", homeHandler(db))
    http.ListenAndServe(":8080", nil)
}
```

这样是非常好的，因为这样我们就不再需要全局变量了。而且这样能非常轻松的 mocking （译者注：用一个虚拟的对象来创建以便测试的测试方法）。我们只需要实现 `GetBacon` 这个方法，就能简单地创建一个数据库 mock 来满足接口的数据需求。

## 测试

现在我们需要一个项目来开始编写我们的测试。关键是测试 `http.Handler`，用的是 [net/http/httptest](https://golang.org/pkg/net/http/httptest/) 包中的 `httptest.ResponseRecorder` 里面的方法。

```go
package main

import(
    "net/http"
    "net/http/httptest"
    "testing"
)

type MockDd struct {}

function (db MockDb) GetBacon() {
    return "bacon"
}

function TestHome(t *testing.T) {
    mockDb := MockDb{}
    homeHandle := homeHandler(mockDb)
    req, _ := http.NewRequest("GET", "", nil)
    w := httptest.NewRecorder()
    homeHandle.ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Errorf("Home page didn't return %v", http.StatusOK)
    }
}
```
当然，这些代码能很好的复用，尤其是进行 POST 请求测试时。因此，我一直在使用这些方法来进行测试操作。

```go
package main

import (
    "net/http"
    "net/http/httptest"
    "net/url"
    "testing"
)

type HandleTester func(
    method string,
    params url.Values,
) *httptest.ResponseRecorder

// Given the current test runner and an http.Handler, generate a
// HandleTester which will test its given input against the
// handler.

func GenerateHandleTester(
    t *testing.T,
    handleFunc http.Handler,
) HandleTester {

    // Given a method type ("GET", "POST", etc) and
    // parameters, serve the response against the handler and
    // return the ResponseRecorder.

    return func(
        method string,
        params url.Values,
    ) *httptest.ResponseRecorder {

        req, err := http.NewRequest(
            method,
            "",
            strings.NewReader(params.Encode()),
        )
        if err != nil {
            t.Errorf("%v", err)
        }
        req.Header.Set(
            "Content-Type",
            "application/x-www-form-urlencoded; param=value",
        )
        w := httptest.NewRecorder()
        handleFunc.ServeHTTP(w, req)
        return w
    }
}

function TestHome(t *testing.T) {
    mockDb := MockDb{}
    homeHandle := homeHandler(mockDb)
    test := GenerateHandleTester(t, homeHandle)
    w := test("GET", url.Values{})
    if w.Code != http.StatusOK {
        t.Errorf("Home page didn't return %v", http.StatusOK)
    }
}
```

更多相关的详细用法，可以[点击](https://github.com/markberger/carton/blob/master/api/auth_test.go)查看我的小项目。

如果你有更好的方法来利用标准库进行 web 应用测试，你可以随时进行留言或给我发 email。

----------------

via: http://markjberger.com/testing-web-apps-in-golang/

作者：[Mark J. Berger](http://markjberger.com/menu/about/)
译者：[zhuCheer](https://github.com/zhuCheer)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
