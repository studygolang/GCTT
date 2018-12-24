首发于：https://studygolang.com/articles/17170

# Go 语言中的错误处理 - 第二部分

## 简介

在 [第一部分](https://studygolang.com/articles/17169) 中，我们学习了 error 接口以及标准库是如何通过 errors 包来创建 error 接口值的。我们也学习了如何使用 error 接口值，通过这些值来判断是否发生了错误。最后，我们学习了一些标准库是如何通过导出 error 接口变量来帮助我们确定发生错误的具体类型。

在 Go 语言中什么时候应该使用自定义错误类型是比较难把握的。大部分情况下，error 包提供的 error 接口值对于报告和处理错误已经足够了。但有时候，调用者可能希望知道错误发生时一些额外的上下文信息。在我看来，这种情况下就应该使用自定义错误类型。

在这篇文章里，我们将要学习自定义错误类型和标准库中两处使用自定义错误类型的实例。每个实例都提供了一个使用自定义错误类型的有趣的视角。之后我们会学习如何通过返回的 error 接口值确定具体的自定义错误类型以及获取存储在其中的指针信息，通过这些额外的信息怎样能帮助我们做出更加合适的错误处理决定。

## net 包

net 包中声明了一个叫做 OpError 的自定义错误类型。指向这个结构的指针一般存储在 error 接口值中返回给调用者。net 包内的许多函数和方法都用到了这个错误类型。

清单 1.1

http://golang.org/src/pkg/net/dial.go

```go
01 func Listen(net, laddr string) (Listener, error) {
02     la, err := resolveAddr("listen", net, laddr, noDeadline)
03     if err != nil {
04         return nil, &OpError{Op: "listen", Net: net, Addr: nil, Err: err}
05     }
06     var l Listener
07     switch la := la.toAddr().(type) {
08     case *TCPAddr:
09         l, err = ListenTCP(net, la)
10     case *UnixAddr:
11         l, err = ListenUnix(net, la)
12     default:
13         return nil, &OpError{Op: "listen", Net: net, Addr: la, Err: &AddrError{Err: "unexpected address type", Addr: laddr}}
14     }
15     if err != nil {
16         return nil, err // l is non-nil interface containing nil pointer
17     }
18     return l, nil
19 }
```

清单 1.1 列出的是 net 包中 Listen 函数的实现代码。我们可以看到，在第 4 行和第 13 行， 创建了 OpError 这个错误类型，并在返回语句中以 error 接口的方式返回给了调用者。由于 OpError 指针实现了 error 接口，所以它可以以 error 接口值返回。需要注意的是，在第 9 行和 11 行，对 ListenTCP 和 ListenUnix 函数的调用 ，同样也可以通过 error 接口返回 OpError 的指针。

接下来，我们看一下 OpError 的声明

清单 1.2

http://golang.org/pkg/net/#OpError

```go
01 // OpError is the error type usually returned by functions in the net
02 // package. It describes the operation, network type, and address of
03 // an error.
04 type OpError struct {
05     // Op is the operation which caused the error, such as
06     // "read" or "write".
07     Op string
08
09     // Net is the network type on which this error occurred,
10     // such as "tcp" or "udp6".
11     Net string
12
13     // Addr is the network address on which this error occurred.
14     Addr Addr
15
16     // Err is the error that occurred during the operation.
17     Err error
18 }
```

清单 1.2 显示的是 OpError 结构的声明。前三个字段提供了错误发生时相关网络操作的上下文信息。第 17 行声明了一个 error 接口类型。这个字段包含了实际发生的错误，通常情况下，这个值的具体类型是一个 errorString 的指针。

另一个需要注意的是自定义错误类型的命名规范，在 Go 语言中自定义错误类型通常以 Error 结尾。以后我们会在其它包中再次看到这样的命名

接下来，我们看一下 OpError 对 error 接口的实现。

清单 1.3

http://golang.org/src/pkg/net/net.go

```go
01 func (e *OpError) Error() string {
02     if e == nil {
03         return "<nil>"
04     }
05     s := e.Op
06     if e.Net != "" {
07         s += " " + e.Net
08     }
09     if e.Addr != nil {
10         s += " " + e.Addr.String()
11     }
12     s += ": " + e.Err.Error()
13     return s
14 }
```
清单 1.3 中列出的是 error 接口的实现代码，展示了如何用与错误发生时相关的信息构建建一个更加具体的错误信息。把上下文信息与错误绑定在一起可以提供额外的信息来帮助调用者做出更加合理的错误处理选择。

## JSON 包

json 包提供了 JSON 格式与 Go 原生格式的相互转化的功能。所有可能产生的错误都是在内部生成的。维护与错误相关的上下文信息对于这个包是比较难的。json 包中有许多自定义错误类型，这些不同的错误类型可以被同一个函数或者方法返回。

让我们看一下其中的一个自定义错误类型

清单 1.4

http://golang.org/src/pkg/encoding/json/decode.go

```go
01 // An UnmarshalTypeError describes a JSON value that was
02 // not appropriate for a value of a specific Go type.
03 type UnmarshalTypeError struct {
04     Value string // description of JSON value
05     Type reflect.Type // type of Go value it could not be assigned to
06 }
07
08 func (e *UnmarshalTypeError) Error() string {
09     return "json: cannot unmarshal " + e.Value + " into Go value of type " + e.Type.String()
10 }
```

清单 1.4 列出了 UnmarshalTypeError 结构的声明和对 error 接口的实现。这个自定义类型是用来说明发生了一个从 JSON 值到具体 Go 原生类型的转化错误。这个结构包含 2 个字段，一个是第 4 行声明的 Value，它包含了用于转换的 JSON 数据，另一个是第 5 行声明的 Type，它包含了将要转化为的 Go 类型。第 8 行对 error 接口的实现中，用相关的上下文信息构建了一个合理的错误信息。

在这个例子里，根据错误类型的名称就可以看出发生了什么错误，这个类型叫做 UnmarshalTypeError，这正是这个自定义错误类型发生时的上下文环境。当发生的错误与转化失败有关时，这个结构的指针就存储在 error　接口中返回。

当调用 unmarshal 时传入了一个非法的参数时，一个指向 InvalidUnmarshalError 的指针会存储在 error 接口中返回。

清单 1.5

http://golang.org/src/pkg/encoding/json/decode.go

```go
01 // An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
02 // (The argument to Unmarshal must be a non-nil pointer.)
03 type InvalidUnmarshalError struct {
04     Type reflect.Type
05 }
06
07 func (e *InvalidUnmarshalError) Error() string {
08     if e.Type == nil {
09         return "json: Unmarshal(nil)"
10     }
11
12     if e.Type.Kind() != reflect.Ptr {
13         return "json: Unmarshal(non-pointer " + e.Type.String() + ")"
14     }
15     return "json: Unmarshal(nil " + e.Type.String() + ")"
16 }
```

清单 1.5 列出了 InvalidUnmarshalError 的声明和对 error 接口的实现。同样的，类型名称就明确了错误发生时的上下文信息。内部维护的状态可以用来构建合适的错误信息，从而帮助调用者做出更加合理的错误处理选择。

## 具体类型识别

在 net 包 Unmarshal 函数的例子中，随 error 接口返回的错误可能是 UnmarshalTypeError, InvalidUnmarshalError 或者 errorString 中的一个。

清单 1.6

http://golang.org/src/pkg/encoding/json/decode.go

```go
01 func Unmarshal(data []byte, v interface{}) error {
02     // Check for well-formedness.
03     // Avoids filling out half a data structure
04     // before discovering a JSON syntax error.
05     var d decodeState
06     err := checkValid(data, &d.scan)
07     if err != nil {
08         return err
09     }
10
11     d.init(data)
12     return d.unmarshal(v)
13 }
14
15 func (d *decodeState) unmarshal(v interface{}) (err error) {
16     defer func() {
17         if r := recover(); r != nil {
18             if _, ok := r.(runtime.Error); ok {
19                 panic
20             }
21             err = r.(error)
22         }
23     }()
24
25     rv := reflect.ValueOf(v)
26     if rv.Kind() != reflect.Ptr || rv.IsNil() {
27         return &InvalidUnmarshalError{reflect.TypeOf(v)}
28     }
29
30     d.scan.reset()
31     // We decode rv not rv.Elem because the Unmarshaler interface
32     // test must be applied at the top level of the value.
33     d.value(rv)
34     return d.savedError
35 }
```
清单 1.6 显示了对 Unmarshal 的调用返回的 error 接口中，是如何包含不同的具体错误类型指针的。在第 27 行中，unmarshal 方法返回了 InvalidUnmarshalError 的指针， 第 34 行，decodeState 变量中的 savedError 被返回，这个字段可能指向好几个不同的具体错误类型。

我们已经知道 JSON 包是用自定义错误类型做为错误发生时的上下文信息的，那我们如何识别包含在 error 接口中的具体错误类型，从而做出更加合理的错误处理选择呢？

让我们从一个使 Unmarshal 函数的调用返回一个包含 UnmarshalTypeError: 自定义类型错误的程序开始。

清单 1.7

http://play.golang.org/p/FVFo8mJLBV

```go
01 package main
02
03 import (
04     "encoding/json"
05     "fmt"
06     "log"
07 )
08
09 type user struct {
10     Name int
11 }
12
13 func main() {
14     var u user
15     err := JSON.Unmarshal([]byte({&quot;name&quot;:&quot;bill&quot;}), &u)
16     if err != nil {
17         log.Println(err)
18         return
19     }
20
21     fmt.Println("Name:", u.Name)
22 }

Output:
2009/11/10 23:00:00 JSON: cannot unmarshal string into Go value of type int
```

清单 1.7 是一个尝试调用 unmarshal 把一段 JSON 文档转化为 Go 类型的例子，第 15 行的 JSON 文档包含一个 name 字段，其中包含一个 bill 的字符串值。由于 user 类型中的 Name 字段在第 9 行被声明为了 integer, 所以对 Unmarshal 的调用返回一个具体的错误类型 UnmarshalTypeError。

现在我们稍微修改一下 表 1.7 中的代码，使 UNmarshal 的调用通过 error　接口返回不同的错误。

清单 1.8

http://play.golang.org/p/n8dQFeHYVp

```go
01 package main
02
03 import (
04     "encoding/json"
05     "fmt"
06     "log"
07 )
08
09 type user struct {
10     Name int
11 }
12
13 func main() {
14     var u user
15     err := JSON.Unmarshal([]byte({&quot;name&quot;:&quot;bill&quot;}), u)
16     if err != nil {
17         switch e := err.(type) {
18         case *json.UnmarshalTypeError:
19             log.Printf("UnmarshalTypeError: Value[%s] Type[%v]\n", e.Value, e.Type)
20         case *json.InvalidUnmarshalError:
21             log.Printf("InvalidUnmarshalError: Type[%v]\n", e.Type)
22         default:
23             log.Println(err)
24         }
25         return
26     }
27
28     fmt.Println("Name:", u.Name)
29 }

Output:
2009/11/10 23:00:00 JSON: Unmarshal(non-pointer main.user)
2009/11/10 23:00:00 InvalidUnmarshalError: Type[main.user]
```

清单 1.8 中的代码在 表 1.7 的基础上做了一些改动。在第 15 行，我们把传给 Unmarshal 函数的参数换成了 u ，而不是之前它的地址。这个变化会使对 UNmarshal 函数的调用返回一个包含 InvalidUnmarshalError 具体错误的 error 接口值。

然后是在程序的第 17 行到第 24 行，添加了些有趣的代码：

清单 1.9

```go
17         switch e := err.(type) {
18         case *json.UnmarshalTypeError:
19             log.Printf("UnmarshalTypeError: Value[%s] Type[%v]\n", e.Value, e.Type)
20         case *json.InvalidUnmarshalError:
21             log.Printf("InvalidUnmarshalError: Type[%v]\n", e.Type)
22         default:
23             log.Println(err)
24         }
```
第 17 行添加了一个 switch 语句来识别存储在 error 接口中的具体错误类型。注意关键字 type　用在接口变量时的语法。同时我们也可以取得存储在具体错类型的的值，并在每个分支语句中使用这些值。

第 18 行和第 20 行的分支语句检测具体的错误类型，然后执行相应的错误处理逻辑。这是识别存储在 error 接口的具体错误类型的典型方式。

## 结论

返回的 error　接口值应该包含对调用者有影响的，错误发生时作用域内一些具体信息。它必须包含足够的信息以便让调用者做出合理的选择，通常来说一个简单的字符串信息就够了，不过有时需要的会更多。

我们在 net 包中看到了一个使用实例：它声明了一个自定义错误类型，用来封装原始的 errro 类型和一些相关的上下文信息。在 JSON 包中，我们看到了如何用自定义错误类型在提供上下文信息和相关的状态。在这两个例子中，维护错误发生时的相关上下文信息是一个决定性因素

如果 errors 包中中 error 类型可以提供足够的上下文信息，就用它。整个标准库中到处用到了它， 通常这个就是你需要的。如果需要给调用者提供额外的信息来帮助他们做出更别合理的错误物理选择，从标准库代码中找一些线索，然后创建你自己的自定义错误类型。

---

via: https://www.ardanlabs.com/blog/2014/11/error-handling-in-go-part-ii.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[jettyhan](https://github.com/jettyhan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
