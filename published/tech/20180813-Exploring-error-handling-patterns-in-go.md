首发于：https://studygolang.com/articles/16163

# 探索 Go 中的错误处理模式

当你学习一种新的编程语言时，可能会存在一个挫败期，就是当你无法使用更熟悉的语言来表达想法的时候。你很自然的想知道为什么语言要设计成这样，很容易误认为（当表达想法遇到困难时）这是语言设计者的失误。这种推理可能会导致你以一种非惯用的方法使用一种语言。

一个挑战我自己观念的内容是如何在 `Go` 中处理错误。概括如下：

* `Go` 中的错误是一个实现了 `error` 接口（实现了 `Error()` 函数）的任意类型。
* 函数返回错和返回其它类型没有区别。使用多返回值将错误和正常区分开。
* 通过检查函数返回值来处理错误，并通过返回值传递到更高层抽象来处理（可以向错误消息追加详细内容）。

例如，考虑一个解析主机地址并侦听 TCP 连接的函数。有两种出错的可能，因此需要有两个错误检查：

```go
func Listen(host string, port uint16) (net.Listener, error) {
	addr, addrErr := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if addrErr != nil {
		return nil, fmt.Errorf("Listen: %s", addrErr)
	}

	listener, listenError := net.ListenTCP("tcp", addr)
	if listenError != nil {
		return nil, fmt.Errorf("Listen: %s", listenError)
	}

	return listener, nil
}
```

这个主题在 [Go 常见问题](https://golang.google.cn/doc/faq#exceptions) 中有自己的条目，社区对此提出了广泛的意见。根据你过去的经验，你可能会倾向于认为：
* Go 应该实现 [某种形式的异常处理](https://opencredo.com/why-i-dont-like-error-handling-in-go/)，允许编写 `try/catch` 块 , 将生成错误的代码组织在一起，并将其与错误处理的代码分开来。
* Go 应该实现 [某种形式的模式匹配](http://yager.io/programming/go.html)，可以供了一种简洁的方法来包装错误，并对值和错误使用不同的形式，等等。

虽然这些在其他语言中是有用的功能，但它们短期内不太可能在 Go 中实现。相反，让我们来看看使用现有功能编写 `Go` 代码时惯用的几种写法。

## 一个稍长的例子

和前面的例子相比可能没有太大的改变，没有更短或更简单，但每次函数调用后都编写 `if` 语句，可能会感觉它正在失控：

```go
func (router HttpRouter) parse(reader *bufio.Reader) (Request, Response) {
	requestText, err := readCRLFLine(reader) //string, err Response
	if err != nil {
		//No input, or it doesn't end in CRLF
		return nil, err
	}

	requestLine, err := parseRequestLine(requestText) //RequestLine, err Response
	if err != nil {
		//Not a well-formed HTTP request line with {method, target, version}
		return nil, err
	}

	if request := router.routeRequest(requestLine); request != nil {
		//Well-formed, executable Request to a known route
		return request, nil
	}

	//Valid request, but no route to handle it
	return nil, requestLine.NotImplemented()
}
```

这种写法还有一些不足之处：

1. 从整个函数来看可以提取一些辅助函数。
2. 从两个错误情况的中间返回成功值有一些难阅读。
3. 搞不清楚第二个 `err` 是新分配的变量还是第一个 `err` 的重新赋值。这和单变量形式使用 `:=` 并不完一致，单变量形式是禁止重新分配变量的。

## 第一种选择：接受这种写法

虽然这种写法让我感觉不好，但每个可能出现错误的位置使用 `if` 进行处理是 `Go` 语言中的惯用方法。我们将探索一些其它方来进行重构，但请注意，此代码在不滥用语言功能的情况下完成了所需的功能。`Go` 的斯巴达式的哲学有一个优点：只要有一种明确的方法，你就可以接受它并继续前进（即使你对标准并不赞同）

根据 `Go` 的官方代码风格，禁止出现未使用的变量和导入。我可能不同意一些代码格式，并认为那里应该允许未使用的变量，但是使用　`goimports` 工具能够很容易遵循标准，而且编译器也没有给你更多的选择。用于选择代码格式的时间现在可以用来重新关注其他更重要的代码。

回到代码的问题，我们可以探索不同的结构使控制流程更清楚，但缺乏通用的方式，`Go` 中的高阶函数限制了我们的选项。

## 非惯用方式

你可能熟悉其他语言中使用的控制流的方法，可以尝试将您喜欢的技术应用于 `Go`。 让我们简单地考虑一些常见的方式（不考虑不常见的情况），这些方式可能会引起 `Go` 社区的关注。

### Defer, Panic, Recover

第一个方式被称为 `Defer`、`Panic` 和 `Recover`，`Panic` 和 `Recove` 类似于其他语言的 `throw` 和 `catch`。这里有几点值得注意：

* `Go` 作者确实在 `Go` 标准库中使用了一个案例，但他们也一直小心翼翼地避免 `panic` 暴露在外部。在大多数情况下，`panic` 是为真正的灾难性错误准备的（非常类似于 `Java` 中的 [`Error`](https://docs.oracle.com/javase/8/docs/api/?java/lang/Error.html) 类用于不可恢复的错误）。
* 异常使用的虚拟变量破坏了函数的引用透明：`Scala` 中的函数编程很好地描述了这一点。总结一下：可以抛出异常的代码，根据它是否包含在 `try/catch` 块中，可以求得不同的值，因此程序员必须知道全局上下文以避免错误。在 GitHub 上有一个很好的[例子](https://github.com/fpinscala/fpinscala/blob/master/exercises/src/main/scala/fpinscala/errorhandling/Option.scala)，并对这个结论有一个简洁的[解释](https://stackoverflow.com/questions/28992625/exceptions-and-referential-transparency/28993780#28993780)。
* 实用性的考虑：[基于异常的代码很难区分正确和错误](https://blogs.msdn.microsoft.com/oldnewthing/20050114-00/?p=36693/)。

`Go` 作者倾向于区分控制流程中的预期分支（例如有效和无效输入）和威胁整个过程的大规模事件。如果你要保持对 `Go` 社区的青睐，你应该努力为今后保留 `panic`。

### 高阶函数和包装类型

`Go` 作者 Rob Pike 一遍又一遍地说[只是写一个 `for` 循环](https://github.com/robpike/filter)，但很难拒绝将示例中的问题视为一系列转换。

```
bufio.Reader -> string -> RequestLine -> Request
```

我们不应该为此写一个映射函数吗？

Go 是一种没有泛型的静态类型语言，因此你可以在使用领域内声明特定类型的类型，或者完全放弃类型安全。

想象一下，如果你试图写的映射函数会是什么样子：

```go
// Sure, we can declare a few commonly used variations...
func mapIntToInt(value int, func(int) int) int { ... }
func mapStringToInt(value string, func(string) int) int { ... }

// ...but how does this help?
type any interface{}
func mapFn(value any, mapper func(any) any) any {
	return mapper(value)
}
```
有一些放弃类型安全的选项，比如 [`Go Promise` 库](https://github.com/chebyrash/promise)。仔细研究示例代码可以看到，示例中通过使用接受任何类型的输出函数（｀ fmt.Println ｀　最终使用反射确定参数类型）来仔细回避类型安全问题。

```go
var p = promise.New(...)
p.Then(func(data interface{}) {
	fmt.Println("The result is:", data)
})
```

编写类似 `Scala` 的 [`Either`](https://www.scala-lang.org/api/2.9.3/scala/Either.html) 类型的包装器也不能真正解决问题，因为它需要具有类型安全的函数来转换 `happy-path` 和 `sad-path` 的值。能够像这样编写示例函数会更好：

```go
func (router HttpRouter) parse(reader *bufio.Reader) (Request, Response) {
	request, response := newStringOrResponse(readCRLFLine(reader)).
		Map(parseRequestLine).
		Map(router.routeRequest)

	if response != nil {
		return nil, response
	} else if request == nil {
		//Technically, this doesn't work because we now lack the intermediate value
		return nil, requested.NotImplemented()
	} else {
		return request, nil
	}
}
```
但是看看你需要编写多少一次性代码才能支持这种写法：

```go
func newStringOrResponse(data string, err Response) *StringOrResponse {
	return &StringOrResponse{data: data, err: err}
}

type StringOrResponse struct {
	data string
	err Response
}

type ParseRequestLine func(text string) (*RequestLine, Response)
func (either *StringOrResponse) Map(parse ParseRequestLine) *RequestLineOrResponse {
	if either.err != nil {
		return &RequestLineOrResponse{data: nil, err: either.err}
	}

	requestLine, err := parse(either.data)
	if err != nil {
		return &RequestLineOrResponse{data: nil, err: either.err}
	}

	return &RequestLineOrResponse{data: requestLine, err: nil}
}

type RequestLineOrResponse struct {
	data *RequestLine
	err Response
}

type RouteRequest func(requested *RequestLine) Request
func (either *RequestLineOrResponse) Map(route RouteRequest) (Request, Response) {
	if either.err != nil {
		return nil, either.err
	}

	return route(either.data), nil
}
```
因此，编写高阶函数的各种形式结果都是非惯用的，不切实际的，或两者兼而有之。`Go` 不是一种函数编程语言。

## 回归本源

现在我们已经看到函数编程的形式并没有什么用处，这让我们[提醒自己](https://blog.golang.org/errors-are-values)：

关键的一课是错误是一种值类型，并且 `Go` 语言的全部功能都可用于处理它们。

另一个好消息是你无法在不知不觉中忽略返回的错误，就像未经检查的异常一样。编译器会强制你至少将错误声明为 `_`，而像 [`errcheck`](https://github.com/kisielk/errcheck) 这样的工具可以很好地保证你的正确。

### 函数组

回顾一下示例代码，有两个明确的错误（输入没有以 CRLF 结束和 HTTP 请求格式不正确），一个清楚的成功响应和一个默认响应。为什么我们不将这些情况进行分组？

```go
func (router HttpRouter) parse(reader *bufio.Reader) (Request, Response) {
	requested, err := readRequestLine(reader)
	if err != nil {
		//No input, not ending in CRLF, or not a well-formed request
		return nil, err
	}

	return router.requestOr501(requested)
}

func readRequestLine(reader *bufio.Reader) (*RequestLine, Response) {
	requestLineText, err := readCRLFLine(reader)
	if err == nil {
		return parseRequestLine(requestLineText)
	} else {
		return nil, err
	}
}

func (router HttpRouter) requestOr501(line *RequestLine) (Request, Response) {
	if request := router.routeRequest(line); request != nil {
		//Well-formed, executable Request to a known route
		return request, nil
	}

	//Valid request, but no route to handle it
	return nil, line.NotImplemented()
}
```
在这里，我们可以通过一些额外的函数使解析功能更小。你可以决定是选择更多的函数还是更大的函数。

### 正确和错误的路径并行

也可以重构现有的功能，同时处理正确和错误的路径。
```go
func (router HttpRouter) parse(reader *bufio.Reader) (Request, Response) {
	return router.route(parseRequestLine(readCRLFLine(reader)))
}

//Same as before
func readCRLFLine(reader *bufio.Reader) (string, Response) { ... }

func parseRequestLine(text string, prevErr Response) (*RequestLine, Response) {
	if prevErr != nil {
		//New
		return nil, prevErr
	}

	fields := strings.Split(text, " ")
	if len(fields) != 3 {
		return nil, &clienterror.BadRequest{
			DisplayText: "incorrectly formatted or missing request-line",
		}
	}

	return &RequestLine{
		Method: fields[0],
		Target: fields[1],
	}, nil
}

func (router HttpRouter) route(line *RequestLine, prevErr Response) (Request, Response) {
	if prevErr != nil {
		//New
		return nil, prevErr
	}

	for _, route := range router.routes {
		request := route.Route(line)
		if request != nil {
			//Valid request to a known route
			return request, nil
		}
	}

	//Valid request, but unknown route
	return nil, &servererror.NotImplemented{Method: line.Method}
}
```
相比原始示例中的四个函数，我们将函数减少到了三个，但是必须从内到外阅读顶级 `parse` 函数。

### 错误闭包

你还可以在遇到第一个错误后创建一个闭包。文章中的示例代码如下所示：

```go
_, err = fd.Write(p0[a:b])
if err != nil {
	return err
}
_, err = fd.Write(p1[c:d])
if err != nil {
	return err
}
_, err = fd.Write(p2[e:f])
if err != nil {
	return err
}
```

作者编写了一个函数只要没有遇到错误，就会继续进行下一步操作

```go
var err error
write := func(buf []byte) {
	if err != nil {
		return
	}
	_, err = w.Write(buf)
}
write(p0[a:b])
write(p1[c:d])
write(p2[e:f])
if err != nil {
	return err
}
```

当你在处理过程中将每个步骤传递给闭包时，这种方法很有效，建议在每个步骤中应用相同的类型。在某些情况下，这可以很好地工作。

在本博客的示例中，你必须创建一个结构来处理错误，并为工作流中的每个步骤编写单独的 `applyParseText / applyParseRequest / applyRoute` 函数，这可能会比它带来的价值更麻烦。

## 总结

虽然 `Go` 在错误处理方面的设计选择起初可能看起来很陌生，但从各种博客和会谈中可以清楚地看出，作者给出的这些选择不是随意的。就个人而言，我试提醒自己缺乏经验的麻烦在于我，而不是 `Go` 作者，而且我可以学会以新的方式思考老的问题。

当我开始撰写本文时，我认为可以从其他更多功能语言借鉴经验，使用更多的函数来使我的 `Go` 代码更简单。这段经历一直很好地提醒我 `Go` 的作者一直在强调的是：有时候编写一些你自己特定用途的函数并继续前进是有帮助的。

---

via: https://8thlight.com/blog/kyle-krull/2018/08/13/exploring-error-handling-patterns-in-Go.html

作者：[Kyle Krull](https://8thlight.com/blog/kyle-krull/)
译者：[althen](https://github.com/althen)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
