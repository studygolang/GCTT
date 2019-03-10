首发于：https://studygolang.com/articles/18799

# Go 装饰器模式教程

装饰器在其他编程语言（如 Python 和 TypeScript）中肯定更为突出，但这并不是说你不能在 Go 中使用它们。事实上，对于某些问题，使用装饰器是完美的解决方案，通过本教程中我们可以了解到装饰器的用法。

## 了解装饰器模式

> 装饰器本质上允许您包装现有功能并在开始或结尾处添加您自己的自定义功能。

在 Go 中，函数被视为第一等对象，这实际上意味着您可以像传递变量一样传递它们。我们来看一个非常简单的例子：

```go
package main

import (
	"fmt"
	"time"
)

func myFunc() {
	fmt.Println("Hello World")
	time.Sleep(1 * time.Second)
}

func main() {
	fmt.Printf("Type: %T\n", myFunc)
}
```

在这个例子中，我们定义了一个名为 `myFunc` 的函数，它只是打印出 `Hello World`。在我们 `main()` 函数的主体中，我们已经调用了 `fmt.Printf` 并使用 `%T` 打印出作为第二个参数传递的值的类型。在示例种，我们传入参数 `myFunc`，这将打印出以下内容：

```
$ go run test.go
Type: func()
```

那么，这对我们来说意味着什么呢？好吧，它突出了这样一个事实，即**函数**可以在我们的代码库的其他部分中**传递并用作参数**。

让我们通过扩展我们的代码库并添加一个 `coolFunc()` 函数来认识这一点，该函数将函数作为其唯一参数：

```go
package main

import (
	"fmt"
	"time"
)

func myFunc() {
	fmt.Println("Hello World")
	time.Sleep(1 * time.Second)
}

// coolFunc takes in a function
// as a parameter
func coolFunc(a func()) {
	// it then immediately calls that functino
	a()
}

func main() {
	fmt.Printf("Type: %T\n", myFunc)
	// here we call our coolFunc function
	// passing in myFunc
	coolFunc(myFunc)
}
```

当我们运行它时，可以看到新输出具有我们期望的字符串 `Hello World`：

```
$ go run test.go
Type: func()
Hello World
```

现在，这可能会让你有点奇怪。为什么要做这样的事情？它本质上为调用 `myFunc` 添加了一层抽象，这使代码复杂化而不会真正增加很多价值。

## 一个简单的装饰者

让我们看看如何使用这种模式为我们的代码库添加一些价值。如果需要，我们可以在执行特定函数时添加一些额外的日志记录，以显示它的开始和结束时间。

```go
package main

import (
	"fmt"
	"time"
)

func myFunc() {
	fmt.Println("Hello World")
	time.Sleep(1 * time.Second)
}

func coolFunc(a func()) {
	fmt.Printf("Starting function execution: %s\n", time.Now())
	a()
	fmt.Printf("End of function execution: %s\n", time.Now())
}

func main() {
	fmt.Printf("Type: %T\n", myFunc)
	coolFunc(myFunc)
}
```

在调用它时，您应该看到像这样的日志：

```go
$ go run test.go
Type: func()
Starting function execution: 2018-10-21 11:11:25.011873 +0100 BST m=+0.000443306
Hello World
End of function execution: 2018-10-21 11:11:26.015176 +0100 BST m=+1.003743698
```

如您所见，我们已经能够有效地包装我的原始函数，而无需改变它的实现。我们现在能够清楚地看到此函数何时启动以及何时完成执行，并向我们强调该函数只需要大约一秒钟即可完成执行。

## 真实世界的例子

让我们看一些如何使用装饰器进一步获取价值的例子。我们将实现一个非常简单的 `http Web` 服务器并装饰我们的端点，以便我们可以验证传入请求是否具有特定的请求头。

> 如果您想了解更多关于在 Go 中编写简单 REST API 的知识，那么我建议您在此处查看我的其他文章：[Creating a REST API in Go](https://tutorialedge.net/golang/creating-restful-api-with-golang/)

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: homePage")
	fmt.Fprintf(w, "Welcome to the HomePage!")
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func main() {
	handleRequests()
}
```

如您所见，我们的代码中没有特别复杂的东西。我们设置了一个服务于单个 `/` 端点的 `net/http` 服务。

让我们添加一个非常简单的身份验证装饰器函数，它将检查 `Authorized` 请求头是否设置为 `true`。

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Checking to see if Authorized header set...")

		if val, ok := r.Header["Authorized"]; ok {
			fmt.Println(val)
			if val[0] == "true" {
				fmt.Println("Header is set! We can serve content!")
				endpoint(w, r)
			}
		} else {
			fmt.Println("Not Authorized!!")
			fmt.Fprintf(w, "Not Authorized!!")
		}
	})
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: homePage")
	fmt.Fprintf(w, "Welcome to the HomePage!")
}

func handleRequests() {

	http.Handle("/", isAuthorized(homePage))
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func main() {
	handleRequests()
}
```

> 注意：这绝对不是处理 REST API 保护的正确方法，我建议您使用 JWT 或 OAuth2 来实现这一目标！

所以，让我们打破这一点，试着了解发生了什么！

我们创建了一个新的装饰器函数 `isAuthorized()`，该函数接受一个与原始 `homePage` 函数匹配相同签名的函数。然后返回一个 `http.Handler`。

在 `isAuthorized()` 函数体内，我们返回一个新的 `http.HandlerFunc` 并在其中验证 `Authorized` 请求头是否等于 `true`。这是一个大大简化的 `OAuth2` 身份验证 / 授权版本，虽然存在一些细微的差异，但它可以让您大致了解它的工作方式。

然而，需要注意的关键是，我们已经设法装饰现有端点并在所述端点周围添加某种形式的身份验证，而无需更改该功能的现有实现。

现在，如果我们要添加一个我们想要保护的新端点，我们可以轻松地这样做：

```go
// define our newEndpoint function. Notice how, yet again,
// we don't do any authentication based stuff in the body
// of this function
func newEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("My New Endpoint")
	fmt.Fprintf(w, "My second endpoint")
}

func handleRequests() {

	http.Handle("/", isAuthorized(homePage))
	// register our /new endpoint and decorate our
	// function with our isAuthorized Decorator
	http.Handle("/new", isAuthorized(newEndpoint))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
```

这突出了装饰器模式的主要优点，在我们的代码库中包装代码非常简单。我们可以使用相同的方法轻松添加新的经过身份验证的端点

## 结论

希望这个教程有助于揭开装饰者的实现，以及如何在自己的 Go 程序中使用装饰器模式。我们了解了装饰器模式的好处以及我们如何使用它来用新功能包装现有功能。

在本教程的第二部分中，我们查看了一个更实际的示例，了解如何在您自己的生产级 Go 系统中使用它。

如果你喜欢这个教程，那么请随意分享这篇文章，这真的有助于网站，我会非常感激！如果您有任何问题和 / 或意见，请在下面的评论部分告诉我们！

> 注意 - 如果您想跟踪新 Go 文章何时发布到网站，请随时在 Twitter 上关注我以获取所有最新消息： [@Elliot_F](https://twitter.com/elliot_f)。

## 进一步阅读

如果您正在寻找更多，那么您可能非常喜欢本网站上的其他一些文章。请随意查看以下文章：

- [Go OAuth2 Tutorial](https://tutorialedge.net/golang/go-oauth2-tutorial/)

---

via: https://tutorialedge.net/golang/go-decorator-function-pattern-tutorial/

作者：[Elliot Forbes](https://twitter.com/elliot_f)
译者：[lovechuck](https://github.com/lovechuck)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
