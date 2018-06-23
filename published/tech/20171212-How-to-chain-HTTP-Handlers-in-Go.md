已发布：https://studygolang.com/articles/12164

# 在 Go 中如何串联 HTTP 处理程序

你好，今天我想分享一下，在 `Go` 语言中串联 HTTP 处理器。

在使用 Go 之前, 我使用 Nodejs + [ExpressJS](http://expressjs.com/en/4x/api.html) 去编写 HTTP 服务器应用。 这个框架提供了很简单的方法去使用中间件和串联很多路由节点，因此，不必指定完整的路由路径来为其添加处理程序。

![图1](https://raw.githubusercontent.com/studygolang/gctt-images/master/chain-http-hanlders/1.png)

这个想法是通过分割你的路由和处理每一个部分，串联到处理器，每个处理程序只负责一部分。它理解起来非常简单且非常容易使用和维护，所以首先我尝试在 Go 中做一些类似的事情。

开箱即用, Go 提供了一个很棒的 [http](https://golang.org/pkg/net/http) 包，它包含了很多不同的工具，当然, 还有 [`ListenAndServe`](https://golang.org/pkg/net/http/#ListenAndServe) 方法，它在给定的端口上启动一个 HTTP 服务器并且通过 [``Handler``](https://golang.org/pkg/net/http/#Handler) 处理它，那么这个 [`Handler`](https://golang.org/pkg/net/http/#Handler) 是什么？

```go
type Handler interface {
	ServeHTTP(ResponseWriter，*Request)
}
```

[`Handler`](https://golang.org/pkg/net/http/#Handler) 是接口，它有一个方法 - ServeHTTP 去处理传入的请求和输出响应。

但是，如果我们想为每一个根路由定义一个处理程序，例如 /api/、/home、/about 等，要怎么做？

[`ServeMux`](https://golang.org/pkg/net/http/#ServeMux) - HTTP 请求复用器，可以帮助你处理这一点. 使用 [`ServeMux`](https://golang.org/pkg/net/http/#ServeMux)，我们可以指定处理器方法来服务任何给定的路由，但问题是我们不能做任何嵌套的 [`ServeMux`](https://golang.org/pkg/net/http/#ServeMux)。

文档中的例子:

```go
mux := http.NewServeMux()
mux.Handle("/api/", apiHandler{})
mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
	}
	fmt.Fprintf(w, "Welcome to the home page!")
})
```

我们可以看到，在这个例子中为 `/api/` 路由自定义了一个处理器并且定义了一个处理方法给根路由。因此任何以 `/api/*` 开头的路由都将使用 apiHandler 处理器方法。 但是如果我们需要串联一个 usersHandler 到 apiHandler，不通过任何的头脑风暴和编码，我们无法做到这点。

为此我写了一个小库 - [gosplitter](https://github.com/goncharovnikita/gosplitter)，它只提供一个公共方法 `Match(url string, mux *http.ServeMux, http.Handler|http.HandlerFunc|interface{})` - 他匹配给定的路由部分和处理器、处理方法或你给定的任何结构！

让我们来看一个例子:

```go
/**
 * 定义一个处理器类型
 */
type APIV1Handler struct {
	mux *http.ServeMux
}

type ColorsHandler struct {
	mux *http.ServeMux
}

/**
 * Start - 绑定父级到子级
 */
func (a *APIV1Handler) Start() {
	var colorsHandler = ColorsHandler{
		mux: a.mux,
	}
	gosplitter.Match("/ping", a.mux, a.HandlePing())
	gosplitter.Match("/colors", a.mux, colorsHandler)
	colorsHandler.Start()
}
func (c *ColorsHandler) Start() {
	gosplitter.Match("/black", c.mux, c.HandleBlack())
}
/**
 * 简单的HTTP处理器方法
 */
func (a *APIV1Handler) HandlePing() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}
}

func (c *ColorsHandler) HandleBlack() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("#000000"))
	}
}

func main() {
	var mux = http.NewServeMux()
	var apiV1 = APIV1Handler{
		mux: mux,
	}

	/**
	* 绑定api处理器到根目录
	*/
	gosplitter.Match("/api/v1", mux, apiV1)

	/**
	* 开始api的处理
	*/
	apiV1.Start()
}
```

举个例子：

```go
/**
 * 定义处理器类型
 */
type APIV1Handler struct {
	mux *http.ServeMux
}

type ColorsHandler struct {
	mux *http.ServeMux
}
```

这里我们定义了一个我们的处理器，它是一个结构体

```go
/**
 * Start - 绑定api处理器到根目录
 */
func (a *APIV1Handler) Start() {
	var colorsHandler = ColorsHandler{
		mux: a.mux,
	}
	gosplitter.Match("/ping", a.mux, a.HandlePing())
	gosplitter.Match("/colors", a.mux, colorsHandler)
	colorsHandler.Start()
}
func (c *ColorsHandler) Start() {
	gosplitter.Match("/black", c.mux, c.HandleBlack())
}
```

添加一个 ``Start`` 方法到我们的处理器程序，去激活处理方法

```go
/**
 * 简单的HTTP处理器方法
 */
func (a *APIV1Handler) HandlePing() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}
}

func (c *ColorsHandler) HandleBlack() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("#000000"))
	}
}
```

添加 `HandlePing` 和 `HandleBlack` 到我们的 `APIV1Handler`，它响应了 `pong` 和 `#000000`

```go
func main() {
	var mux = http.NewServeMux()
	var apiV1 = APIV1Handler{
		mux: mux,
	}

	/**
	 * 绑定API处理器到根路由
	 */
	gosplitter.Match("/api/v1", mux, apiV1)
	/**
	 * 启动API处理器
	 */
	apiV1.Start()
}
```

我们在 `main` 方法中创建了一个新的 `ServeMux` 然后创建了一个 `APIV1Handler` 的实例，把它绑定到了 `/api/v1` 路由，然后启动了它。

所以在所有这些简单的操作之后我们拥有了两个工作中的路由: `/api/v1/ping` 和 `/api/v1/colors/black`，会响应 `pong` 和 `#000000`。

使用起来不是很容易么？我认为是这样, 现在在我的项目中使用这个库来方便的进行路由分割和串联处理器 :)

<!-- Thanks for reading! Any suggestions and critiques are welcome! -->

感谢阅读，欢迎提出任何建议和批评!

---

via: https://medium.com/@cashalot/how-to-chain-http-handlers-in-go-33c96396b397

作者：[Nikita](https://medium.com/@cashalot)
译者：[MarlonFan](https://github.com/MarlonFan)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
