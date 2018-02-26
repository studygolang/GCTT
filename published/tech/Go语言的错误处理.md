
已发布：https://studygolang.com/articles/12430

[原文](https://scene-si.org/2017/11/13/error-handling-in-go/)

# Go 语言的错误处理

Go 语言的错误处理是基于明确的目的而设计的。你应该从函数中返回所有可能的错误，并且检查/处理这些返回值。和其他语言相比，这一点可能看起来有些繁琐和不人性化，其实并不是这样的。让我们来看看一些基本的例子，然后继续做一些较重要的事情。

## Non 错误

实际上 Go 有个概念 non-error。这是一个语言特性，不能用在用户自定义函数中。最明显的例子就是从 map 中通过 key 获取值。

```go
if val, ok := data["key"]; ok {
	// key/value 在 map 中存在
}
```
当尝试获取指定 key 的值的时候，会返回一个可选的第二个值，是一个 boolean 类型表示获取的值是否存在。

```go
func main() {
	a := map[string]string{"key": "value"}
	b := a["key"]
	c, ok := a["key"]
	d := a["foo"]
	fmt.Printf("%#v %#v %#v %#v %#v", a, b, c, d, ok)
}
```
运行这段[程序](https://play.golang.org/p/CAZSyh9_q3)不会有任何错误。你能看到是否接受第二个返回值是完全可选的。

另一个例子是从 channel 中成功读取数据。同样的，你可以在读取操作返回时，使用变量来接收第二个返回值。

```go
if j, ok := <-jobs; ok {
	fmt.Println("received job", j)
} else {
	fmt.Println("received all jobs")
	done <- true
}
```

第二个参数是一个 boolean 类型，表示语言结构层面的成功或失败，并不是一个严格的返回类型。你可以写一个函数声明 `func() (interface{}, bool)` 同上面的代码语义相同，但是不能够忽略第二个参数 bool 返回值了，你需要为他指定一个接收变量。

## 忽略错误

Go 提供了足够的灵活性可以让你忽略指定的返回错误。例如你想转换一个字符串到数字类型，并且你不在意转换失败时返回 0 。你可以使用 `_` 字符来忽略指定的返回值，在下面例子中忽略了 error 返回值:  

```go
v := "abc"
s, _ := strconv.Atoi(v)
fmt.Printf("%d\n", s)
```
很明显转换不会成功，在这处理 “invalid syntax” 错误会很繁琐。当然这取决于你的使用场景，有一些场景处理返回错误没什么价值。

我近期遇到的一个例子是  [sony/sonyflake](https://github.com/sony/sonyflake) 。 这个项目是一个 ID 生成器，返回 int64 类型的 id 和可能的错误。

> 想要生成一个新的 id ，你只要调用 NextID 方法即可。
> `func (sf *Sonyflake) NextID() (uint64, error)`
> NextID 能够连续生成 ID 从开始时间到 174 年左右。当超过这个限制的时候，NextID 会返回一个错误。

我非常确信在看这篇文章的人不会活过174年。在这种情况下，你真的需要处理那个特定的错误么？这里真的需要返回一个错误么？

我认为这是一个设计缺陷，我们可以使用 Go 的另一个灵活性来更好地处理：`panic`。参见一篇很棒的文章 [go by example](https://gobyexample.com/panic):

> 使用 panic 的一个通用的场景就是如果一个函数返回了一个错误值，但是我们不知道或者不希望去处理的时候中断执行。

还有一些其他的关于忽略返回错误的例子。可能最常见的忽略返回错误的处理方式是在  [json.Marshal](https://gobyexample.com/json) 中。在明确的理解之后，有些错误在第一次发生时可以不去处理。

## 连续的错误处理

你的目的应该是处理所有的错误，像下面这样结束执行相对比较容易：

```go
base64decoder := base64.NewDecoder(base64.StdEncoding, r.Body)
gz, err := zlib.NewReader(base64decoder)
if err != nil {
	return err
}
defer gz.Close()

decoder := json.NewDecoder(gz)
var t SentryV6Notice
err = decoder.Decode(&t)
if err != nil {
	return err
}
r.Body.Close()
// ...
```

如果能像处理 if 语句一样，独立处理每个返回错误不是更好么？让我们看一些你不知道的情况：  

`if func1() || func2() || func3() {`

这个 if 语句会分别测试每个表达式。也就是说如果 `func1()` 返回了 false，那么 `func2` 和 `func3` 函数就不会被调用。if 语句可以中断执行流程，尽管如此也没有方法使用一条语句完成检查返回错误。 至少你可以按照下面的方法来处理：  

```go
if gz, err := zlib.NewReader(base64decoder); err != nil {
	return err
}
// ...
if err := decoder.Decode(&t); err != nil {
	return err
}
```

这个例子中，我们在表达式的前面添加了一个简单的语句，这条语句会在测试表达式之前执行。这是 [Go 语言规范](https://golang.org/ref/spec#If_statements)的另一个特性。可惜的是，我们不能使用这个特性来控制程序流程。但是，我们可以考虑创建一个可变参函数来接收 `func() error` 参数，并且在第一个错误发生时立即返回。

```go
func flow(fns ...func() error) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
```

这里例子 [playground example](https://play.golang.org/p/AStZiZ_-Ml) 演示了如何实现一个顺序调用函数的处理，并且所做的修改不会影响到函数结构。基本上只依赖于如何保存函数返回值 除了返回错误之外。

如果你正在处理 `jmoiron/sqlx` 你可以[这么写](https://play.golang.org/p/W-QEybSQwG):

```go
err := flow(
	func() error { return db.Get(result, "select one row") },
	func() error { return db.Select(result, "select multiple rows") },
	func() error { return db.Get(result, "select another row") },
)
```

一个明显的问题就是过于冗余的部分，把一个局部函数包装到函数签名中，如果编程语言的语义允许在 if 语句中(多次)赋值和测试的话，出错检查可以简化成：

```go
var err error
if err = db.Get(result, "select one row") ||
	err = db.Select(result, "select multiple rows") ||
	err = db.Get(result, "select another row") {
	return err
)
```

可惜的是，Go 不会把 errors 作为 boolean 表达式处理，也不允许在 boolean 表达式中赋值，会提示错误 ”expected boolean expression, found simple statement (missing parentheses around composite literal?)“ 。对于这一点我并不是很强烈的认为不好，因为还有其他的方法可以做到。如果你在其他语言例如 Node 或者 PHP，尝试使用赋值给一个变量来代替测试一个值的话，你会发现 Go 的处理方式更加优美，注意：有时也非常痛苦。

## 改进错误处理

通常我写的包括输入，输出参数的验证函数，功能函数的函数签名是 `func() error`。来看一个更复杂些的例子，我创建了一个 response writer 输入参数为 interface{} 把第一个非空值或函数返回值写入 `http.ResponseWriter`。

```go
// JSON responds with the first non-nil payload, formats error messages
func JSON(w http.ResponseWriter, responses ...interface{}) {
	respond := func(payload interface{}) {
		json, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	}

	for _, response := range responses {
		switch value := response.(type) {
		case nil:
			continue
		case func() error:
			err := value()
			if err == nil {
				continue
			}
			respond(Error(err))
		case error:
			respond(Error(value))
		default:
			respond(struct {
				Response interface{} `json:"response"`
			}{response})
		}
		// Exit on the first output...
		break
	}
}
```

这个函数的好处是提供了基于可变参数的条件执行处理。和传入 `...error` 和 `[]error` 不同的是不需要先执行所有的函数。

使用这个函数的一个简单的 API 调用样例如下：

```go
input := RequestInput{}
result := RequestResult{}
validate := func() error {
	// write and validate things for input
}
process := func() error {
	// write things to result, return error if any
}
resputil.JSON(w, validate, process, result)
```

在后面的语句 `resputil.JSON` 中可以看出，程序执行流程是显而易见并且明确的，当有任何错误发生时都会中断执行并返回。
另一个附带的好处是，无论何时发生错误了，你只要返回错误即可，不需要关心在这里应该返回的其他返回值，因为会在外部的闭包中被处理，并且只处理一次。

> 注意：语言需要支持函数 `return err` 而不仅仅只是返回错误信息，当你使用这种处理方法时，所有其他的返回值都是未初始化时的默认值，所以不违反期待函数返回值的规则，在这里也有类似的建议 [this issue filed for Go2](https://github.com/golang/go/issues/21161#issuecomment-318350273)

## 异步错误处理

几周前在 [reddit](https://www.reddit.com/r/golang/comments/77bf8c/function_composition_in_go_with_reflect/dokx20g/) 上有过一个讨论， @ligustah 推荐看一下 x/sync/errgroup 包，这个包提供了类似 sync.WaitGroup 实现方式的 Group structure ，会返回发生的第一个错误或者在没有错误时返回 nil 。每个函数都可以在 goroutines 中执行。  
引用至 godoc 中的例子 [JustErrors](https://godoc.org/golang.org/x/sync/errgroup#ex-Group--JustErrors)  

```go
var g errgroup.Group
var urls = []string{
	"http://www.golang.org/",
	"http://www.google.com/",
	"http://www.somestupidname.com/",
}
for _, url := range urls {
	// Launch a goroutine to fetch the URL.
	url := url // https://golang.org/doc/faq#closures_and_goroutines
	g.Go(func() error {
		// Fetch the URL.
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
		}
		return err
	})
}
// Wait for all HTTP fetches to complete.
if err := g.Wait(); err == nil {
	fmt.Println("Successfully fetched all URLs.")
}
```

开始的时候，我觉得这个包还不错，但是里面有一些需要注意的地方。

1. 你可能需要返回所有的错误（你只能拿到第一个）  
2. 在发生错误的时候想要中断执行（必须等到所有的 goroutines 执行完毕）  
3. 执行的是并行检查（没有提供顺序执行的 API ）

根据你的使用场景，可以参照这个模型，做一些私有的实现。这个包本身有些繁琐，看了前面的介绍，写一个聪明的错误检查封装函数只需要少数的几行代码即可。

## 最后总结

有些时候看起来 Go 语言检查返回值错误是比较痛苦的事情，特别是当你受到有 try/catch 特性的语言影响的时候，例如： Java，PHP 等。当你第一次看到 Go 的这种处理方式的时候可能并不喜欢，希望检查错误的处理能更好些（更简洁些？），我相信其他的语言有更糟糕的例子，更加繁琐，更多不足的地方。  

errors 和 panics 有一些不同的地方，实际上 panics 是带有函数调用栈信息的。如果你想在 errors 里面添加调用栈信息，我推荐你使用 Dave Cheney 的 [pkg/errors](https://dave.cheney.net/2016/06/12/stack-traces-and-the-errors-package) 包。文章 [Don’t just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully) 对每个人来说都是在必读列表中的。

在其他语言中对错误的处理有一些显著的痛点，如果 Go 的处理有一点繁琐的话，那么它带来的是更多的好处。

----------

via: https://scene-si.org/2017/11/13/error-handling-in-go/

作者：[Tit Petric](https://scene-si.org/about/)
译者：[tyler2018](https://github.com/tyler2018)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
