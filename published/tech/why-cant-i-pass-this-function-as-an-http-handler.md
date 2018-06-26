已发布：https://studygolang.com/articles/13286

# 为什么我不能将这个函数作为一个 Http Handler 来传递？！

当我帮助人们学习网页开发时，一个超级常见的问题是，“为什么我不能把这个函数传入 `http.Handle` 方法中？它看起来与 `http.HandlerFunc` 是一模一样的！”

```go
func demo(h http.Handler) {}

func handler(w http.ResponseWriter, r *http.Request) {}

func main() {
	// 这行代码在编译时会报错
	demo(handler)
}
```
> 在 Go Playground 中运行示例代码 → https://play.golang.org/p/JAY4RDyQQn3

我绝对能理解他们的困惑。对于一个接受 `http.HandlerFunc` 类型参数的函数来说，我们可以把显式定义为 `http.HandlerFunc`  类型的变量传入其中，我们也可以把有 `func(http.ResponseWriter, *http.Request)`  这种签名的函数作为参数传入，那为什么上面的代码不能运行？

为了回答这个问题，我们需要深究下 `http.HandlerFunc` 和 `http.Handler`  类型，同时我们用一些示例代码来探求他们的特性。

我们从 `http.HandlerFunc` 开始，它的定义是：
```go
type HandlerFunc func(w ResponseWriter, r *Request)
```
于是，我们可以写这样的代码：
```go
func demo(fn http.HandlerFunc) {}

func handler(w http.ResponseWriter, r *http.Request) {}

func main() {
	demo(handler)
}
```
> 在 Go Playground 中运行示例代码 → https://play.golang.org/p/NdDbOhQPFQh

注意，这段代码与最上面的问题代码是完全不同的 - 我们的 `demo` 函数 接收的参数类型是  `http.HandlerFunc` ,而不是 `http.Handler` 。还有 一个你可能都没有发现的事实是，代码中的 `handler` 其实并不是 一个  `http.HandlerFunc` ，它的实际类型的是 `func(http.ResponseWriter, *http.Request)` ，我们可以通过运行下面的测试代码来揭示这个事实。

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Printf("%T", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// handler func的实现内容
}
```

> 在 Go Playground 中运行示例代码 → https://play.golang.org/p/ep8xlyKUODx

所以，我们为啥会期待在最上面的问题代码中能将 `handler` 作为参数 传给 `dmeo` 函数呢？它的类型明显都不一样！

当然，在我们的示例中，我们的 `handler` 函数 完全符合 `http.HandlerFunc`  的定义，所以 Go 编译器可以推断我们是想把我们的变量转为那个类型。也就是说，编译器可以假装我们的代码是这样写的：

```go
demo(http.HandlerFunc(demo))
```

加上去的 `http.HandlerFunc` 部分是从 `demo` 函数中推导出来的。

好了，现在我们可以进行第二步了 -- `http.HandlerFunc` 也实现了 `http.Handler` 接口。为啥是这样的，我们只需要看看两者的源码即可。

```go
type HandlerFunc func(w ResponseWriter, r *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}

type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

请注意 `HandlerFunc` 有一个 `ServeHTTP` 方法匹配了 `Handler` 的接口声明，这意味着 `HandlerFunc` 实现了 `Handler`  接口。这可能看起来很奇怪，但是在 Go 语言中这是完全有效的。毕竟 `HandlerFunc` 是另外一种类型，而我们可以随时往任意类型上增加方法。因为 `HandlerFunc` 实现了 `Handler` 接口，我们可以将 `HandlerFunc` 传递给接受 `Handler` 参数的函数中，只要参数类型是明确的被定义为 `HandlerFunc`。示例代码如下：

```go
package main

import "net/http"

func demo(fn http.Handler) {}

func handler(w http.ResponseWriter, r *http.Request) {}

func main() {
	demo(http.HandlerFunc(handler))
}
```
> 在 Go Playground 中运行示例代码 → https://play.golang.org/p/K7sD51wnBL9

现在，我们知道 `http.HandlerFunc` 实现了 `http.Handler`  接口，而且我们知道可以将类型为 `func(http.ResponseWriter, *http.Request)` 的对象传递给一个需要 `http.HandlerFunc` 类型参数的函数中，但是，为什么我们不能传递 `func(http.ResponseWriter, *http.Request)` 类型给一个需要 `http.HandlerFunc` 类型参数的函数中？或者说，为什么这样的代码不能通过编译呢？

```go
func demo(h http.Handler) {}

func handler(w http.ResponseWriter, r *http.Request) {}

func main() {
	// 这行代码在编译时会报错
	demo(handler)
}
```

简单的回答是，编译器不知道怎么将 `func(http.ResponseWriter, *http.Request)` 类型转换为 `http.Handler` 类型。为了实现这样的转换，编译器需要知道我们想要将 `handler` 转换为 `http.HandlerFunc`，但是在上面的代码中，我们根本就没有看到 `http.HandlerFunc` 这个类型。

如果用我们之前说的推断规则，Go 编译器只有可能认为代码其实是想要这样写的：

```go
demo(http.Handler(handler))
```

但是这显然是不对的。`handler` 并没有实现 `http.Handler` 接口，除非它转换成实现了接口的 `http.HandlerFunc` 类型。

还有一种办法，我们做一个两步转换，将 `func(http.ResponseWriter, *http.Request)` 转换成 http.Handler:

```
          1              2
func -> HandlerFunc -> Handler
```

如果我们显式的将 `handler` 函数转换为 `http.HandlerFunc`，那么这个代码是可以运行的。

```go
func demo(h http.Handler) {}

func handler(w http.ResponseWriter, r *http.Request) {}

func main() {
	// 代码可以编译！
	demo(http.HandlerFunc(handler))
}
```
> 在 Go Playground 中运行示例代码 → https://play.golang.org/p/Dm33TpgvONh

再一次，如果我们假设 Go 编译器能假装我们的代码是将传入的参数自动转换为实际希望的类型，代码看起来像是这样的：

```go
demo(http.Handler(http.HandlerFunc(handler)))
```

那么，为什么编译器不学着自动为我们做这两步简单明了的类型转换？对于初学者的解释是，编译器并不知道这两步转换就是你真正想要做的。

让我们假设你还有另外一个类型，假设它叫 `CowboyFunc`，它的定义如下：

```go
type CowboyFunc func(w http.ResponseWriter, r *http.Request)

func (f CowboyFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s := time.Now()
	f(w, r)
	fmt.Println("Cowboy function duration:", time.Now().Sub(s))
}
```

现在，如果我们调用 `demo(handler)` 就像我们之前做的那样，然后编译器自动将 `handler` 类型转换为了 `http.HandlerFunc` 类型，然后编译顺利完成，但是这是开发者真正想要做的么？万一开发者是想将 `handler` 转换成 `CowboyFunc` 呢？它也实现了 `http.Handler` 接口。

```go
demo(CowboyFunc(handler))
```
> 在 Go Playground 运行示例代码 → https://play.golang.org/p/tJELwdEThyx

你可以看到这也是有效的代码，这意味着编译器并不知道我们到底是要这两者的哪一个，除非我们清楚明白的告诉它。

总而言之，你不能直接将 `func(http.ResponseWriter, *http.Request)` 传给 需要 `http.Handler` 类型参数的函数中，这也许让你觉得疑惑或者很繁琐，但是不能这么做是有合理的原因的，而一旦你了解了这些原因，你可能更感谢那些时不时就出现的，让人讨厌的错误了。

---

via: https://www.calhoun.io/why-cant-i-pass-this-function-as-an-http-handler/

作者：[Jon Calhoun](https://www.calhoun.io/about)
译者：[MoodWu](https://github.com/MoodWu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
