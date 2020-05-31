首发于：https://studygolang.com/articles/28983

# 如何写好 Go 代码

我写了多年的 Go 微服务，并在写完两本关于 ([API Foundations in Go](https://leanpub.com/api-foundations) 和 [12 Factor Applications with Docker and Go](https://leanpub.com/12fa-docker-golang)) 主题的书之后，有了一些关于如何写好 Go 代码的想法

但首先，我想给阅读这篇文章的读者解释一点。好代码是主观的。你可能对于好代码这一点，有完全不同的想法，而我们可能只对其中一部分意见一致。另一方面，我们可能都没有错，只是我们从两个角度出发，从而选择了不同的方式解决工程问题，并不意味着意见不一致的不是好代码。

## 包

包很重要，你可能会反对 - 但是如果你在用 Go 写微服务，_你可以将所有代码放在一个包中_。当然，下面也有一些反对的观点：

1. 将定义的类型放入单独的包中
2. 维护与传输无关的服务层
3. 在服务层之外，维护一个数据存储(repository)层

我们可以计算一下，一个微服务包的最小数量是 1。如果你有一个大型的微服务，它拥有 websocket 和 http 网关，你最终可能需要 5 个包（类型，数据存储，服务，websocket 和 http 包）。

简单的微服务实际上并不关心从数据存储层(repository)，或者从传输层(websocket，http)抽离业务逻辑。你可以写简单的代码，转换数据然后响应，也是可以运行的。但是，添加更多的包可以解决一些问题。例如，如果你熟悉  SOLID 原则，`S` 代表单一职责。如果我们拆分成包，这些包就可以是单一职责的。

* `types` - 声明一些结构，可能还有一些结构的别名等
* `repository` - 数据存储层，用来处理存储和读取结构
* `service` - 服务层，包装存储层的具体业务逻辑实现
* `http`, `websocket`, ... - 传输层，用来调用服务层

当然，根据你使用的情况，还可以进一步细分，例如，可以使用 `types/request` 和 `types/response` 来更好的分隔一些结构。这样就可以拥有 `request.Message` 和 `response.Message` 而不是 `MessageRequest` 和 `MessageResponse`。如果一开始就像这样拆分开，可能会更有意义。

但是，为了强调最初的观点 - 如果你只用了这些声明包中的一部分，也没什么影响。像 Docker 这样的大型项目在 `server` 包下只使用了 `types`  包，这是它真正需要的。它使用的其他包（像 errors 包），可能是第三方包。

同样需要注意的是，在一个包中，共享正在处理的结构和函数会很容易。如果你有相互依赖的结构，将它们拆分为两个或多个不同的包可能会导致[钻石依赖问题](https://www.well-typed.com/blog/2008/04/the-dreaded-diamond-dependency-problem/)。解决方案也很显然 - 将代码放到一块儿，或者将所有代码放在一个包中。

到底选哪一个呢？两种方法都行。如果我非要按规则来的话，将其拆分更多的包可能会使添加新代码变得麻烦。因为你可能要修改这些包才能添加单个 API 调用。如果不是很清楚如何布局，那么在包之间跳转可能会带来一些认知上的开销。在很多情况下，如果项目只有一两个包，阅读代码会更容易。

你肯定也不想要太多的小包。

## 错误

如果是描述性的 Errors 可能是开发人员检查生产问题的唯一工具。这就是为什么我们要优雅地处理错误，要么将它们一直传递到应有程序的某一层，如果错误无法处理，该层就接收错误并记录下来，这一点非常重要。以下是标准库错误类型缺少的一些特性：

* 错误信息不含堆栈跟踪
* 不能堆积错误
* errors 是预实例化的

但是，通过使用第三方错误包(我最喜欢的是[pkg/Errors](https://github.com/pkg/errors).))可以帮助解决这些问题。也有其他的第三方错误包，但是这个是 [Dave Cheney](https://dave.cheney.net) (Go 语言大神)编写的，它在错误处理的方式在一定程度上是一种标准。他的文章 [Don’t just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully) 是推荐必读的。

### 错误的堆栈跟踪

`pkg/errors` 包在调用 `errors.New` 时，会将上下文(堆栈跟踪)添加到新建的错误中。

```bash
users_test.go:34: testing error Hello world
	github.com/crusttech/crust/rbac_test.TestUsers
		/go/src/github.com/crusttech/crust/rbac/users_test.go:34
	testing.tRunner
		/usr/local/go/src/testing/testing.go:777
	runtime.goexit
		/usr/local/go/src/runtime/asm_amd64.s:2361
```

考虑到完整的错误信息是 "Hello world"，使用 `fmt.Printf` 带有 `%+v` 的参数或者类似的方式来打印少量的上下文 - 对于查找错误的而言，是一件很棒的事。你可以确切知道是哪里创建了错误（关键字）。当然，当涉及到标准库时，`errors` 包和本地 `error` 类型 - 不提供堆栈跟踪。但是，使用 `pkg/errors` 可以很容易地添加一个。例如：

```go
resp, err := u.Client.Post(fmt.Sprintf(resourcesCreate, resourceID), body)
if err != nil {
	return errors.Wrap(err, "request failed")
}
```

在上面这个例子中，`pkg/errors` 包将上下文添加到 err 中，加的错误消息(`"request failed"`) 和堆栈跟踪都会抛出来。通过调用 `errors.Wrap` 来添加堆栈跟踪，所以你可以精准追踪到此行的错误。

### 堆积错误

你的文件系统，数据库，或者其他可能抛出相对不太好描述的错误。例如，Mysql 可能会抛出这种强制错误：

```bash
ERROR 1146 (42S02): Table 'test.no_such_table' doesn't exist
```

这不是很好处理。然而，你可以使用 `errors.Wrap(err，"database aseError")` 在上面堆积新的错误。这样，就可以更好地处理 `"databaseError"` 等。`pkg/errors` 包将在 `causer` 接口后面保留实际的错误信息。

```go
type causer interface {
	Cause() error
}
```

这样，错误堆积在一起，不会丢失任何上下文。附带说一下，mysql 错误是一个[类型错误](https://github.com/go-sql-driver/mysql/blob/a8b7ed4454a6a4f98f85d3ad558cd6d97cec6959/errors.go#L58)，其背后包含的不仅仅是错误字符串的信息。这意味着它有可能被处理的更好：

```go
if driverErr, ok := err.(*mysql.MySQLError); ok {
	if driverErr.Number == mysqlerr.ER_ACCESS_DENIED_ERROR {
		// Handle the permission-denied error
	}
}
```

此例子来自于 [this Stack Overflow thread](https://stackoverflow.com/questions/47009068/how-to-get-the-mysql-error-type-in-golang)。

### 错误预实例化

究竟什么是错误(error)呢？非常简单，错误需要实现下面的接口：

```go
type error interface {
	Error() string
}
```

在 `net/http` 的例子中，这个包将几种错误类型暴露为变量，如[文档](https://golang.org/pkg/net/http/#pkg-variables)所示。在这里添加堆栈跟踪是不可能的（Go 不允许对全局 var 声明可执行代码，只能进行类型声明）。其次，如果标准库将堆栈跟踪添加到错误中 - 它不会指向返回错误的位置，而是指向声明变量（全局变量）的位置。

这意味着，你仍然需要在后面的代码中强制调用类似于  `return errors.WithStack(ErrNotSupported)` 的代码。这也不是很痛苦，但不幸的是，你不能只导入 `pkg/errors` ，就让所有现有的错误都带有堆栈跟踪。如果你还没有使用 `errors.New` 来实例化你的错误，那么它需要一些手动调用。

## 日志

接下来是日志，或者更恰当的说，结构化日志。这里提供了许多软件包，类似于 [sirupsen/logrus](https://github.com/sirupsen/logrus)或我最喜欢的[APEX/LOG](https://github.com/apex/log)。这些包也支持将日志发送到远程的机器或者服务，我们可以用工具来监控这些日志。

当谈到标准日志包时，我不常看到的一个选项是创建一个自定义 logger，并将 `log.LShorfile` 或 `log.LUTC` 等标志传递给它，以再次获得一点上下文，这能让你的工作变轻松 - 尤其在处理不同时区的服务器时。

```go
const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)
```

即使你没有创建自定义 logger，你也可以使用 `SetFlags` 来修改默认 logger。([playground link](https://play.golang.org/p/jlplSGTDoyI))：

```go
package main

import (
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Hello, playground")
}
```

结果如下：

```bash
2009/11/10 23:00:00 main.go:9: Hello, playground
```

你不想知道你在哪里打印了日志吗？这会让跟踪代码变得更容易。

## 接口

如果你正在写接口并命名接口中的参数，请考虑以下的代码片段：

```go
type Mover interface {
	Move(context.Context, string, string) error
}
```

你知道这里的参数代表什么吗？只需要在接口中使用命名参数就可以让它很清晰。

```go
type Mover interface {
	Move(context.Context, source string, destination string)
}
```

我还经常看到一些使用一个具体类型作为返回值的接口。一种未得到充分利用的做法是，根据一些已知的结构体或接口参数，以某种方式声明接口，然后在接收器中填充结果。这可能是 Go 中最强大的接口之一。

```go
type Filler interface {
	Fill(r *http.Request) error
}

func (s *YourStruct) Fill(r *http.Request) error {
	// here you write your code...
}
```

更可能的是，一个或多个结构体可以实现该接口。如下：

```go
type RequestParser interface {
	Parse(r *http.Request) (*types.ServiceRequest, error)
}
```

此接口返回具体类型（而不是接口）。通常，这样的代码会使你代码库中的接口变得杂乱无章，因为每个接口只有一个实现，并且在你的应用包结构之外会变得不可用。

### 小帖士

如果你希望在编译时确保你的结构体符合并完全实现一个接口（或多个接口），你可以这么做：

```go
var _ io.Reader = &YourStruct{}
var _ fmt.Stringer = &YourStruct{}
```

如果你缺少这些接口所需的某些函数，编译器就会报错。字符 `_` 表示丢弃变量，所以没有副作用，编译器完全优化了这些代码，会忽视这些被丢弃的行。

## 空接口

与上面的观点相比，这可能是更有争议的观点 - 但是我觉得使用 ``interface{}`` 有时非常有效。在 HTTP API 响应的例子中，最后一步通常是 JSON 编码，它接收一个接口参数：

```go
func (enc *Encoder) Encode(v interface{}) error
```

因此，完全可以避免将 API 响应设置成具体类型。我并不建议对所有情况都这么处理，但是在某些情况下，可以在 API 中完全忽略响应的具体类型，或者至少说明具体类型声明的意义。脑海中浮现的一个例子是使用匿名结构体。

```go
body := struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles,omitempty"`
}{username, roles}
```

首先，不使用 `interface{}` 的话，无法从函数里返回这种结构体。显然，json 编码器可以接受任何类型的内容，因此，按传递空接口(对我来说)是完全有意义的。虽然趋势是声明具体类型，但有时候你可能不需要一层中间层。对于包含某些逻辑并可能返回各种形式的匿名结构体的函数，空接口也很合适。

> 更正：匿名结构体不是不可能返回，只是做起来很麻烦：[playground](https://play.golang.org/p/turu_Yg--6h)
>
> 感谢 @Ikearens at [Discord Gophers](https://discord.gg/quNN7yP) #golang channel

第二个用例是数据库驱动的 API 设计，我之前写过一些[有关内容](https://scene-si.org/2018/02/07/sql-as-an-api/)，我想指出的是，实现一个完全由数据库驱动的 API 是非常可能的。这也意味着添加和修改字段是*仅仅在数据库中*完成的，而不会以 ORM 的形式添加额外的间接层。显然，你仍然需要声明类型才能在数据库中插入数据，但是从数据库中读取数据可以省略声明。

```go
// getThread fetches comments by data, order by ID
func (api *API) getThread(params *CommentListThread) (comments []interface{}, err error) {
	// calculate pagination parameters
	start := params.PageNumber * params.PageSize
	length := params.PageSize
	query := fmt.Sprintf("select * from comments where news_id=? and self_id=? and visible=1 and deleted=0 order by id %s limit %d, %d", params.Order, start, length)
	err = api.db.Select(&comments, query, params.NewsID, params.SelfID)
	return
}
```

同样，你的应用程序可能充当反向代理，或者只使用无模式(schema-less)的数据库存储。在这些情况下，目的只是传递数据。

一个大警告(这是你需要输入结构体的地方)是，修改 Go 中的接口值并不是一件容易的事。你必须将它们强制转换为各种内容，如 map、slice 或结构体，以便可以在访问这些返回的数据。如果你不能保持结构体一成不变，而只是将它从 DB(或其他后端服务)传递到 JSON 编码器(会涉及到断言成具体类型)，那么显然这个模式不适合你。这种情况下不应该存在这样的空接口代码。也就是说，当你不想了解任何关于载荷的信息时，空接口就是你需要的。

## 代码生成

尽可能使用代码生成。如果你想生成用于测试的 mock，如果你想生成 proc/GRPC 代码，或者你可能拥有的任何类型的代码生成，可以直接生成代码并提交。在发生冲突的情况下，可以随时将其丢弃，然后重新生成。

唯一可能的例外是提交类似于 `public_html` 文件夹的内容，其中包含你将使用 [rakyll/statik](https://github.com/rakyll/statik) 打包的内容。如果有人想告诉我，由 [gomock](https://github.com/golang/mock) 生成的代码在每次提交时都会以兆字节的数据污染 GIT 历史记录？不会的。

## 结束语

关于 Go 的最佳实践和最差实践的另一篇文章应该是[Idiomatic Go](https://about.sourcegraph.com/go/idiomatic-go/)。如果你不熟悉的话，可以阅读一下 - 它是与本文很好的搭配。

我想在这里引用[Jeff Atwood post - The Best Code is No Code At All](https://blog.codinghorror.com/the-best-code-is-no-code-at-all/)文章的一句话，这是一句令人难忘的结束语：

> 如果你真的喜欢写代码，你会非常喜欢尽可能少地写代码。

但是，一定要编写那些单元测试。_完结_。

---

via: https://scene-si.org/2018/07/24/writing-great-go-code/

作者：[Tit Petric](https://scene-si.org/)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
