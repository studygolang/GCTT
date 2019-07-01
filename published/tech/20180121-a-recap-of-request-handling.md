首发于：https://studygolang.com/articles/21553

# Go 中的请求处理概述

使用 Go 处理 HTTP 请求主要涉及两件事：ServeMuxes 和 Handlers。

[ServeMux](https://docs.studygolang.com/pkg/net/http/#ServeMux) 本质上是一个 HTTP 请求路由器（或多路复用器）。它将传入的请求与预定义的 URL 路径列表进行比较，并在找到匹配时调用路径的关联 handler。

handler 负责写入响应头和响应体。几乎任何对象都可以是 handler，只要它满足[http.Handler](https://docs.studygolang.com/pkg/net/http/#Handler) 接口即可。在非专业术语中，这仅仅意味着它必须是一个拥有以下签名的 `ServeHTTP` 方法：

`ServeHTTP(http.ResponseWriter, *http.Request)`

Go 的 HTTP 包附带了一些函数来生成常用的 handler，例如[FileServer](https://docs.studygolang.com/pkg/net/http/#FileServer)，[NotFoundHandler](https://docs.studygolang.com/pkg/net/http/#NotFoundHandler) 和[RedirectHandler](https://docs.studygolang.com/pkg/net/http/#RedirectHandler)。让我们从一个简单的例子开始：

```
$ mkdir handler-example
$ cd handler-example
$ touch main.go
```

> File: main.go

```go
package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	rh := http.RedirectHandler("http://example.org", 307)
	mux.Handle("/foo", rh)

	log.Println("Listening...")
	http.ListenAndServe(":3000", mux)
}
```

让我们快速介绍一下：

- 在 `main` 函数中，我们使用[http.NewServeMux](https://docs.studygolang.com/pkg/net/http/#NewServeMux) 函数创建了一个空的 ServeMux。
- 然后我们使用[http.RedirectHandler](https://docs.studygolang.com/pkg/net/http/#RedirectHandler) 函数创建一个新的 handler。该 handler 将其接收的所有请求 307 重定向到 http://example.org。
- 接下来我们使用[mux.Handle](https://docs.studygolang.com/pkg/net/http/#ServeMux.Handle) 函数向我们的新 ServeMux 注册它，因此它充当 URL 路径 `/foo` 的所有传入请求的 handler。
- 最后，我们创建一个新服务并使用[http.ListenAndServe](https://docs.studygolang.com/pkg/net/http/#ListenAndServe) 函数开始监听传入的请求，并传入 ServeMux 给这个方法以匹配请求。

继续运行应用程序：

```
$ go run main.go
Listening...
```

并在浏览器中访问[http://localhost:3000/foo](http://localhost:3000/foo)。你会发现请求已经被成功重定向。

你可能已经注意到了一些有趣的东西：ListenAndServe 函数的签名是 `ListenAndServe(addr string, handler Handler)`，但我们传递了一个 ServeMux 作为第二个参数。

能这么做是因为 ServeMux 类型也有一个 ServeHTTP 方法，这意味着它也满足 Handler 接口。

对我而言，它只是将 ServeMux 视为*一种特殊的 handler*，而不是把响应本身通过第二个 handler 参数传递给请求。这不像刚刚听说时那么惊讶 - 将 handler 链接在一起在 Go 中相当普遍。

## 自定义 handler

我们创建一个自定义 handler，它以当前本地时间的指定格式响应：

```go
type timeHandler struct {
	format string
}

func (th *timeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tm := time.Now().Format(th.format)
	w.Write([]byte("The time is: " + tm))
}
```

这里确切的代码并不太重要。

真正重要的是我们有一个对象（在该示例中它是一个 `timeHandler` 结构，它同样可以是一个字符串或函数或其他任何东西），并且我们已经实现了一个带有签名 `ServeHTTP(http.ResponseWriter, *http.Request)` 的方法。这就是我们实现一个 handler 所需的全部内容。

让我们将其嵌入一个具体的例子中：

> File: main.go

```go
package main

import (
	"log"
	"net/http"
	"time"
)

type timeHandler struct {
	format string
}

func (th *timeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tm := time.Now().Format(th.format)
	w.Write([]byte("The time is: " + tm))
}

func main() {
	mux := http.NewServeMux()

	th := &timeHandler{format: time.RFC1123}
	mux.Handle("/time", th)

	log.Println("Listening...")
	http.ListenAndServe(":3000", mux)
}
```

在 `main` 函数中，我们使用 ` ＆ ` 符号生成指针，用与普通结构完全相同的方式初始化 `timeHandler`。然后，与前面的示例一样，我们使用 `mux.Handle` 函数将其注册到我们的 ServeMux。

现在，当我们运行应用程序时，ServeMux 会将任何通过 `/time` 路径的请求直接传递给我们的 `timeHandler.ServeHTTP` 方法。

试一试：[http://localhost:3000/time](http://localhost:3000/time)。

另请注意，我们可以轻松地在多个路径中重复使用 timeHandler：

```go
func main() {
	mux := http.NewServeMux()

	th1123 := &timeHandler{format: time.RFC1123}
	mux.Handle("/time/rfc1123", th1123)

	th3339 := &timeHandler{format: time.RFC3339}
	mux.Handle("/time/rfc3339", th3339)

	log.Println("Listening...")
	http.ListenAndServe(":3000", mux)
}
```

## 普通函数作为 handler

对于简单的情况（如上例），定义新的自定义类型和 ServeHTTP 方法感觉有点啰嗦。让我们看看另一个方法，我们利用 Go 的[http.HandlerFunc](https://docs.studygolang.com/pkg/net/http/#HandlerFunc) 类型来使正常的函数满足 Handler 接口。

任何具有签名 `func(http.ResponseWriter, *http.Request)` 的函数都可以转换为 HandlerFunc 类型。这很有用，因为 HandleFunc 对象带有一个内置的 `ServeHTTP` 方法 - 这非常巧妙且方便 - 执行原始函数的内容。

如果这听起来令人费解，请尝试查看[相关的源代码](https://golang.org/src/net/http/server.go?s=57023:57070#L1904)。你将看到它是一种让函数满足 Handler 接口的非常简洁的方法。

我们使用这种方法来重写 timeHandler 应用程序：

> File: main.go

```go
package main

import (
	"log"
	"net/http"
	"time"
)

func timeHandler(w http.ResponseWriter, r *http.Request) {
	tm := time.Now().Format(time.RFC1123)
	w.Write([]byte("The time is: " + tm))
}

func main() {
	mux := http.NewServeMux()

	// Convert the timeHandler function to a HandlerFunc type
	th := http.HandlerFunc(timeHandler)
	// And add it to the ServeMux
	mux.Handle("/time", th)

	log.Println("Listening...")
	http.ListenAndServe(":3000", mux)
}
```

事实上，将函数转换为 HandlerFunc 类型，然后将其添加到 ServeMux 的情况比较常见，Go 提供了一个快捷的转换方法：[mux.HandleFunc](https://docs.studygolang.com/pkg/net/http/#ServeMux.HandleFunc) 方法。

如果我们使用这个转换方法，`main()` 函数将是这个样子：

```go
func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/time", timeHandler)

	log.Println("Listening...")
	http.ListenAndServe(":3000", mux)
}
```

大多数时候使用这样的 handler 很有效。但是当事情变得越来越复杂时，将会受限。

你可能已经注意到，与之前的方法不同，我们必须在 `timeHandler` 函数中对时间格式进行硬编码。当我们想要将信息或变量从 `main()` 传递给 handler 时会发生什么？

一个简洁的方法是将我们的 handler 逻辑放入一个闭包中，把我们想用的变量包起来：

> File: main.go

```go
package main

import (
	"log"
	"net/http"
	"time"
)

func timeHandler(format string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		tm := time.Now().Format(format)
		w.Write([]byte("The time is: " + tm))
	}
	return http.HandlerFunc(fn)
}

func main() {
	mux := http.NewServeMux()

	th := timeHandler(time.RFC1123)
	mux.Handle("/time", th)

	log.Println("Listening...")
	http.ListenAndServe(":3000", mux)
}
```

`timeHandler` 函数现在有一点点不同。现在使用它来返回 handler，而不是将函数强制转换为 handler（就像我们之前所做的那样）。能这么做有两个关键点。

首先它创建了一个匿名函数 `fn`，它访问形成闭包的 `format` 变量。无论我们如何处理闭包，它总是能够访问它作用域下所创建的局部变量 - 在这种情况下意味着它总是可以访问 `format` 变量。

其次我们的闭包有签名为 `func(http.ResponseWriter, *http.Request)` 的函数。你可能还记得，这意味着我们可以将其转换为 HandlerFunc 类型（以便它满足 Handler 接口）。然后我们的 `timeHandler` 函数返回这个转换后的闭包。

在这个例子中，我们仅仅将一个简单的字符串传递给 handler。但在实际应用程序中，您可以使用此方法传递数据库连接，模板映射或任何其他应用程序级的上下文。它是全局变量的一个很好的替代方案，并且可以使测试的自包含 handler 变得更整洁。

你可能还会看到相同的模式，如下所示：

```go
func timeHandler(format string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tm := time.Now().Format(format)
		w.Write([]byte("The time is: " + tm))
	})
}
```

或者在返回时使用隐式转换为 HandlerFunc 类型：

```go
func timeHandler(format string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tm := time.Now().Format(format)
		w.Write([]byte("The time is: " + tm))
	}
}
```

## DefaultServeMux

你可能已经看到过很多地方提到的 DefaultServeMux，包括最简单的 Hello World 示例到 Go 源代码。

我花了很长时间才意识到它并不特别。 DefaultServeMux 只是一个普通的 ServeMux，就像我们已经使用的那样，默认情况下在使用 HTTP 包时会实例化。以下是 Go 源代码中的相关行：

```go
var DefaultServeMux = NewServeMux()
```

通常，你不应使用 DefaultServeMux，因为它会带来**安全风险**。

由于 DefaultServeMux 存储在全局变量中，因此任何程序包都可以访问它并注册路由 - 包括应用程序导入的任何第三方程序包。如果其中一个第三方软件包遭到破坏，他们可以使用 DefaultServeMux 向 Web 公开恶意 handler。

因此，根据经验，避免使用 DefaultServeMux 是一个好主意，取而代之使用你自己的本地范围的 ServeMux，就像我们到目前为止一样。但如果你决定使用它……

HTTP 包提供了一些使用 DefaultServeMux 的便捷方式：[http.Handle](https://docs.studygolang.com/pkg/net/http/#Handle) 和[http.HandleFunc](https://docs.studygolang.com/pkg/net/http/#HandleFunc)。这些与我们已经看过的同名函数完全相同，不同之处在于它们将 handler 添加到 DefaultServeMux 而不是你自己创建的 handler。

此外，如果没有提供其他 handler（即第二个参数设置为 `nil`），ListenAndServe 将退回到使用 DefaultServeMux。

因此，作为最后一步，让我们更新我们的 timeHandler 应用程序以使用 DefaultServeMux：

> File: main.go

```go
package main

import (
	"log"
	"net/http"
	"time"
)

func timeHandler(format string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		tm := time.Now().Format(format)
		w.Write([]byte("The time is: " + tm))
	}
	return http.HandlerFunc(fn)
}

func main() {
	// Note that we skip creating the ServeMux...

	var format string = time.RFC1123
	th := timeHandler(format)

	// We use http.Handle instead of mux.Handle...
	http.Handle("/time", th)

	log.Println("Listening...")
	// And pass nil as the handler to ListenAndServe.
	http.ListenAndServe(":3000", nil)
}
```

如果你喜欢这篇博文，请不要忘记查看我的新书[《用 Go 构建专​​业的 Web 应用程序》](https://lets-go.alexedwards.net/) ！

在推特上关注我 [@ajmedwards](https://twitter.com/ajmedwards)。

此文章中的所有代码都可以在[MIT Licence](http://opensource.org/licenses/MIT) 许可下免费使用。

---
via: https://www.alexedwards.net/blog/a-recap-of-request-handling

作者：[Alex Edwards](https://www.alexedwards.net/)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
