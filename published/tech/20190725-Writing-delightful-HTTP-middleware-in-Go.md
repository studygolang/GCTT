首发于：https://studygolang.com/articles/25913

# 在 Go 中编写令人愉快的 HTTP 中间件

在使用 Go 编写复杂的服务时，您将遇到一个典型的主题是中间件。这个话题在网上被讨论了一次又一次。本质上，中间件允许我们做了如下事情：

* 拦截 `ServeHTTP` 调用，执行任意代码
* 对调用链（Continuation Chain) 上的请求/响应流进行更改
* 打断中间件链，或继续下一个中间件拦截器并最终到达真正的请求处理器

这些与 express.js 中间件所做的工作非常类似。我们探索了各种库，找到了接近我们想要的现有解决方案，但是他们要么有不要的额外内容，要么不符合我们的品位。显然，我们可以在 express.js 中间件的启发下，写出 20 行代码以下的更清晰的易用的 API(Installation API)

## 抽象

在设计抽象时，我们首先设想如何编写中间件函数(下文开始称为拦截器)，答案非常明显：

```go
func NewElapsedTimeInterceptor() MiddlewareInterceptor {
    return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
        startTime := time.Now()
        defer func() {
            endTime := time.Now()
            elapsed := endTime.Sub(startTime)
            // 记录时间消耗
        }()

        next(w, r)
    }
}

func NewRequestIdInterceptor() MiddlewareInterceptor {
    return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
        if r.Headers.Get("X-Request-Id") == "" {
            r.Headers.Set("X-Request-Id", generateRequestId())
        }

        next(w, r)
    }
}
```

它们看起来就像 `http.HandlerFunc`，但有一个额外的参数 `next`，该函数（参数）会继续处理请求链。这将允许任何人像编写类似 `http.HandlerFunc` 的简单函数一样写拦截器，它可以拦截调用，执行所需操作，并在需要时传递控制权。

接下来，我们设想如何将这些拦截器连接到 `http.Handler` 或 `http.HandlerFunc` 中。为此，首先要定义 `MiddlewareHandlerFunc`，它只是 `http.HandlerFunc` 的一种类型。(type MiddlewareHandlerFunc http.HandlerFunc)。这将允许我们在 `http.HandlerFunc` 栈上之上构建一个更好的 API。现在给定一个 `http.HandlerFunc` 我们希望我们的链式 API 看起来像这样:

```go
func HomeRouter(w http.ResponseWriter, r *http.Request) {
	// 处理请求
}

// ...
// 在程序某处注册 Hanlder
chain := MiddlewareHandlerFunc(HomeRouter).
  Intercept(NewElapsedTimeInterceptor()).
  Intercept(NewRequestIdInterceptor())

// 像普通般注册 HttpHandler
mux.Path("/home").HandlerFunc(http.HandlerFunc(chain))
```

将 `http.HandlerFunc` 传递到 `MiddlewareHandlerFunc`，然后调用 `Intercept` 方法注册我们的 `Interceptor`。`Interceptor` 的返回类型还是 `MiddlewareHandlerFunc`，它允许我们再次调用 `Intercept`。

使用 `Intercept` 组合需要注意的一件重要事情是执行的顺序。由于 chain(responseWriter, request)是间接调用最后一个拦截器，拦截器的执行是反向的，即它从尾部的拦截器一直返回到头部的处理程序。这很有道理，因为你在拦截调用时，拦截器应该要在真正的请求处理器之前执行。

## 简化

虽然这种反向链系统使抽象更加流畅，但事实证明，大多数情况下 s 我们有一个预编译的拦截器数组，能够在不同的 handlers 之间重用。同样，当我们将中间件链定义为数组时，我们自然更愿意以它们执行顺序声明它们(而不是相反的顺序)。让我们将这个数组拦截器称为中间件链。我们希望我们的中间件链看起来有点像：

```go
// 调用链或中间件可以按下标的顺序执行
middlewareChain := MiddlewareChain{
  NewRequestIdInterceptor(),
  NewElapsedTimeInterceptor(),
}

// 调用所有以 HomeRouter 结尾的中间件
mux.Path("/home").Handler(middlewareChain.Handler(HomeRouter))
```

## 实现

一旦我们设计好抽象的概念，实现就显得简单多了

```go
package middleware

import "net/http"

// MiddlewareInterceptor intercepts an HTTP handler invocation, it is passed both response writer and request
// which after interception can be passed onto the handler function.
type MiddlewareInterceptor func(http.ResponseWriter, *http.Request, http.HandlerFunc)

// MiddlewareHandlerFunc builds on top of http.HandlerFunc, and exposes API to intercept with MiddlewareInterceptor.
// This allows building complex long chains without complicated struct manipulation
type MiddlewareHandlerFunc http.HandlerFunc


// Intercept returns back a continuation that will call install middleware to intercept
// the continuation call.
func (cont MiddlewareHandlerFunc) Intercept(mw MiddlewareInterceptor) MiddlewareHandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		mw(writer, request, http.HandlerFunc(cont))
	}
}

// MiddlewareChain is a collection of interceptors that will be invoked in there index order
type MiddlewareChain []MiddlewareInterceptor

// Handler allows hooking multiple middleware in single call.
func (chain MiddlewareChain) Handler(handler http.HandlerFunc) http.Handler {
	curr := MiddlewareHandlerFunc(handler)
	for i := len(chain) - 1; i >= 0; i-- {
		mw := chain[i]
		curr = curr.Intercept(mw)
	}

	return http.HandlerFunc(curr)
}
```

因此，在不到 20 行代码(不包括注释)的情况下，我们就能够构建一个很好的中间件库。它几乎是简简单单的，但是这几行连贯的抽象实在是太棒了。它使我们能够毫不费力地编写一些漂亮的中间件链。希望这几行代码也能激发您的中间件体验。

---

via: https://doordash.engineering/2019/07/22/writing-delightful-http-middlewares-in-go/

作者：[Katy Slemon](https://medium.com/@katyslemon)
译者：[Alex1996a](https://github.com/Alex1996a)
校对：[DingdingZhou](https://github.com/DingdingZhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
