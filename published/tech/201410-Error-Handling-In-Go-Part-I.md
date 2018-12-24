首发于：https://studygolang.com/articles/17169

# Go 语言中的错误处理 - 第一部分

## 简介

在 Go 语言中，如果一个函数或者方法需要返回任何错误，通常会使用 error 接口类型作为返回类型。在标准库中，所有返回了错误信息的函数和方法使用的都是这个接口。例如，下面是 http 包中 Get 方法的声明：

清单 1.1

http://golang.org/pkg/net/http/#Client.Get

```go
func (c *Client) Get(url string) (resp *Response, err error)
```

在 清单 1.1 的代码中，Get 方法的第二个返回参数是一个 error 接口类型的值。要处理函数或者方法返回的错误值，首先需要检查这个值是否为 nil:

清单 1.2

```go
resp, err := c.Get("http://goinggo.net/feeds/posts/default ")
if err != nil {
    log.Println(err)
    return
}
```

在 清单 1.2 的代码中，调用了 Get 方法，方法的返回值赋给了两个局部变量，然后用变量 err 和 nil 进行了比较，如果这个值不是 nil，则说明发生了错误。

由于我们处理错误值时用的是接口，所以需要定义一个实现了这个接口的具体类型。标准库已经为我们定义了一个叫做 errorString 的结构，并且实现了 error 接口。在这篇文章里，我们会介绍 error 接口和 errorString 结构的实现以及如何使用。

## Error 接口和 errorString 结构

Go 语言为我们直接提供了 error 接口的声明：

清单 1.3

http://golang.org/pkg/builtin/#error

```go
type error interface {
    Error() string
}
```

在 清单 1.3 的代码中，我们看到了 error 接口只声明了一个叫做 Error  的方法，这个方法返回一个 string。所以，一个类型只要实现了 Error 方法也就实现了 error 接口，就可以作为 error 接口类型的实例来使用。如果你对 Go 语言中的接口还不太熟悉的话，可以参考我的另一篇文章 [接口、方法以及嵌入类型](http://www.goinggo.net/2014/05/methods-interfaces-and-embedded-types.html)

同时，标准库也声明了一个叫做 errorString 的结构，我们可以在 errors 包中找到它的代码：

清单 1.4
```go
http://golang.org/src/pkg/errors/errors.go

type errorString struct {
    s string
}
```
在 清单 1.4 的代码中，我们可以看到 errorString 结构只声明了一个 string 类型的字段 s。这个类型连同它唯一的字段都是非公开的，也就是说我们不能直接访问这个类型或者其中的字段。要了解更多 Go 语言中关于 非公开标识符的内容，可以参考这篇文章 [Go 语言中的 公开 / 非公开 标识符](http://www.goinggo.net/2014/03/exportedunexported-identifiers-in-go.html)。

errorString 结构实现了 error 接口

清单 1.5

http://golang.org/src/pkg/errors/errors.go

```go
func (e *errorString) Error() string {
    return e.s
}
```

在 清单 1.5 的代码中，errorString 是以指针接受者的方式来实现 error 接口的。也就是说，只有 errorString 结构的指针才可以作为 error 接口实例来使用。而且由于 errorString 类型和它的字段都是非公开的，所以我们不能进行类型检测或者类型转化。我们唯一可以做的就是调用它的 Error 方法。

在标准库中，errorString 是作为 error 接口实例来使用的最常用的具体错误类型。现在我们已经知道了这些类型的定义，接下来我们就来学习一下标准库提供了哪些方式，通过 errorString 结构来创建 error 接口的实例。

## 创建 Error 实例

标准库提供了两种创建 errorString 指针实例，并作为 error 接口实例来使用的方法。如果你需要返回的错误信息只是一个简单的字符串，没有特殊的格式要求，那么 errors 包中的 New 函数就是你需要的：

清单 1.6

```go
var ErrInvalidParam = errors.New("mypackage: invalid parameter")
```
清单 1.6 的代码中，展示了 errors 包中的 New 函数的典型使用方式。在这个例子中，通过调用 New 函数，我们声明并且初始化了一个 error 接口的实例。下面我们来看一下 New 函数的声明和实现：

清单 1.7

http://golang.org/src/pkg/errors/errors.go

```go
// New returns an error that formats as the given text.
func New(text string) error {
    return &errorString{text}
}
```

在 清单 1.7 列出的的 New 函数的声明中，我们可以看到函数接收一个 string 类型的参数，返回一个 error 接口的实例。在函数的实现部分，创建了一个 errorString 结构的指针。然后在返回语句中，编译器创建了一个 error 接口实例并且和这个 errorString 结构的指针进行了绑定来满足函数对返回值的要求。这样 errorString 结构的指针就成了返回的 error 接口实例的具体类型了。

如果你需要返回的错误信息需要格式化，可以使用 fmt 包中的 Errorf 函数

清单 1.8

```go
var ErrInvalidParam = fmt.Errorf("invalid parameter [%s]", param)
```

清单 1.8 的代码中，展示了 Errorf 函数的典型用法。如果你对 fmt 包中的其它格式化函数比较熟悉，那么你会发现这个函数的用法和其它格式化函数的用法是一样的。同样的，我们通过调用 Errorf，声明并且初始化了一个 error 接口的实例。

下面我们来看一下 Errorf 函数的声明和实现：

清单 1.9

http://golang.org/src/pkg/fmt/print.go

```go
// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
func Errorf(format string, a … interface{}) error {
    return errors.New(Sprintf(format, a … ))
}
```

在 清单 1.9 列出的 Errorf 函数的声明部分，我们看到了 error 接口类型又一次的被用作返回类型。在函数的实现部分，用 errors 包中的 New 函数为格式化后的错误信息创建了一个 error 接口的实例。所以，不管你用 errors 包还是 fmt 包来创建 error 接口实例，底层都是一个 errorString 结构的指针。

我们已经知道了用 errorString 类型指针来创建 error 接口实例的两种不同的方法，接下来，让我们的来看一下标准库是如何对 API 调用返回的不同错误值进行区分提供支持的。

## 比较错误值

和其它标准库一样，bufio 包使用 errors 包中的 New 函数创建包级别的 error 接口变量：

清单 1.10

http://golang.org/src/pkg/bufio/bufio.go

```go
var (
    ErrInvalidUnreadByte = errors.New("bufio: invalid use of UnreadByte")
    ErrInvalidUnreadRune = errors.New("bufio: invalid use of UnreadRune")
    ErrBufferFull        = errors.New("bufio: buffer full")
    ErrNegativeCount     = errors.New("bufio: negative count")
)
```

清单 1.10 的代码中，bufio 包声明并且初始化了 4 个包级别的 error 接口变量。注意每个变量名都是以 Err 开头的，这是 Go 语言中的命名规范。困为这些变量声明为了 error 接口类型，我们就可以用 这些变量来对从 bufio 包中不同 API 的调用返回的错误值进行识别。

清单 1.11

```go
data, err := b.Peek(1)
if err != nil {
    switch err {
    case bufio.ErrNegativeCount:
        // Do something specific.
        return
    case bufio.ErrBufferFull:
        // Do something specific.
        return
    default:
        // Do something generic.
        return
    }
}
```

在 清单 1.11 的代码中，通过 ufio.Reader 指针类型的变量 b, 调用了 Peek 方法。Peek 方法可能返回 ErrNegativeCount 或者 ErrBufferFull 这两个错误变量中的一个。由于这些变量是公开的，所以我们就可以通过这些变量来判断具体返回的是哪个错误。这些变量成了包的错误处理 API 的一部分。

假设 bufio 包没有声明这些 error 类型的变量，我们就不得不通过比较错误信息来判断发生的是什么错误：

清单 1.12

```go
data, err := b.Peek(1)
if err != nil {
    switch err.Error() {
    case "bufio: negative count":
        // Do something specific.
        return
    case "bufio: buffer full":
        // Do something specific.
        return
    default:
        // Do something specific.
        return
    }
}
```

清单 1.12 中的代码中有两个问题，一：对 Error() 的调用 ，会创建一个错误信息的拷贝，然后在 switch 语句中使用，二：如果包的作者作者更改了错误信息，这段代码就不能正常工作了。

io 包是另一个为了可能返回的错误声明 error 类型变量的例子：

清单 1.13

http://golang.org/src/pkg/io/io.go

```go
var ErrShortWrite    = errors.New("short write")
var ErrShortBuffer   = errors.New("short buffer")
var EOF              = errors.New("EOF")
var ErrUnexpectedEOF = errors.New("unexpected EOF")
var ErrNoProgress    = errors.New("multiple Read calls return no data or error")
```

清单 1.13 的代码显示了 io 包声明了 6 个包级别的 error 类型变量。对包中函数的调用如果返回了第三个 error 类型变量 EOF，则说明没有更多的输入了。调用这个包中的函数后，常常需要把返回的错误值与 EOF 变量进行比较。

下面是 io 包中 ReadAtLeast 方法的实现代码。

清单 1.14

http://golang.org/src/pkg/io/io.go

```go
func ReadAtLeast(r Reader, buf []byte, min int) (n int, err error) {
    if len(buf) < min {
        return 0, ErrShortBuffer
    }
    for n < min && err == nil {
        var nn int
        nn, err = r.Read(buf[n:])
        n += nn
    }
    if n >= min {
        err = nil
    } else if n > 0 && err == EOF {
        err = ErrUnexpectedEOF
    }
    return
}
```

清单 1.14 列出的 ReadAtLeast 的实现代码中，展示了如何使用这些变量。注意 ErrShortBuffer 和 ErrUnexpectedEOF 是如何作为返回值来使用的。也需要注意函数是如何用变量 err 与 EOF 进行比较的。和我们之前自己写的代码中的做法是一样的。

在实现我们自己的函数库的时候，可以考虑这种为包中函数可能返回的错误创建 error 变量的模式。这样可以提供一个处理包中错误的 API ，并且保持错误处理的一致性。

## 为什么不用命名类型

大家可能会问，为什么 Go 语言不把 errroString 设计为一个命名类型？

让我们看一下用命名类型来实现和用结构类型实现有什么区别。

清单 1.15

http://play.golang.org/p/uZPi4XKMF9

```go
01 package main
02
03 import (
04     "errors"
05     "fmt"
06 )
07
08 // Create a named type for our new error type.
09 type errorString string
10
11 // Implement the error interface.
12 func (e errorString) Error() string {
13     return string(e)
14 }
15
16 // New creates interface values of type error.
17 func New(text string) error {
18     return errorString(text)
19 }
20
21 var ErrNamedType = New("EOF")
22 var ErrStructType = errors.New("EOF")
23
24 func main() {
25     if ErrNamedType == New("EOF") {
26         fmt.Println("Named Type Error")
27     }
28
29     if ErrStructType == errors.New("EOF") {
30         fmt.Println("Struct Type Error")
31     }
32 }

Output:
Named Type Error
```

清单 1.15 中的代码展示了如果把 erroString 作为一个命名类型来使用会遇到的问题。程序的第 9 行声明了一个以 string 为基础类型的命名类型 errorString。然后在第 12 行，为这个命名类型实现了 error 接口。为了模仿 errors 包中的 New 方法，在第 17 行为这个命名类型实现了 New 方法。

接着在第 21 和 22 行，声明并初始化了两个 error 类型变量。变量 ErrNamedType 是用 New 函数初始化的，而变量 ErrStructType 是用 errors.New 函数初始化的。最后在主函数里，我们把这两个变量分别同用同样的方法创建的变量进行了比较。

当你运行程序的时候，输出的结果会比较有趣。

第 25 行的条件语句为真，而第 29 行的条件语句为假。使用命名类型时，我们可以用同样的错误信息创建 error 接口实例，并且它们是相等的。这个问题和我们在 清单 1.12 中遇到的问题一样。我们可以用相同的错误信息创建本地的 error 类型变量，并和包中声明的 error 类型变量进行比较。但是如果包的作者更改了错误信息，我们的代码就不能正常工作了。

当 errorString 是一个结构类型时，也可能出现同样的问题。让我们看一下用值接收者的方式来实现 error 接口会发生什么：

清单 1.16

http://play.golang.org/p/EMWPT-tWp4

```go
01 package main
02
03 import (
04    "fmt"
05 )
06
07 type errorString struct {
08    s string
09 }
10
11 func (e errorString) Error() string {
12    return e.s
13 }
14
15 func NewError(text string) error {
16    return errorString{text}
17 }
18
19 var ErrType = NewError("EOF")
20
21 func main() {
22    if ErrType == NewError("EOF") {
23        fmt.Println("Error:", ErrType)
24    }
25 }

Output:
Error: EOF
```
清单 1.16 的代码中，我们用值接收者的方式为类型 errorString 实现了 error 接口。这次我们遇到了和 清单 1.15 中使用命名类型时同样的问题。当我们比较接口实例时，实际比较的是底层的具体实例（具体的错误信息）。

由于标准库为 errorString 结构实现 error 接口时用的是指针接收者。所以 errors.New 方法必须返回一个指针实例。这个实例就是与 error 接口实例绑定的那个。并且每次都是不一样的。在这些情况中，比较的是指针而不是底层具体的错误信息。

## 结论

这篇文章中，我们介绍了 error 接口是什么以及它是如何与 errorString 结构协同工作的。在 Go 语言中，通常用 errors.New 和 fmt.Errorf 来创建 error 接口实例，并且强烈推荐使用它们。通常来说一个简单的错误信息加上一些基本的格式化，就是我们处理错误时需要的一切。

我们也学习了标准库为了让我们区分包中函数调用 返回的不同错误类型，而声明一些 error 类型变量的模式。标准库中的许多包都声明并且公开了这些 error 类型的变量，通过这些变量，足够让我们区分包中函数返回的不同错误类型。

可能我们有时候需要创建自定义错误类型。这是我们会在 第二部分 里介绍的内容。目前，用标准库提供的错误处理机制，按例子中的用法使用就可以了。

阅读 [Go 语言中的错误处理 - 第二部分](https://studygolang.com/articles/17170)

---

via: https://www.ardanlabs.com/blog/2014/10/error-handling-in-go-part-i.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[jettyhan](https://github.com/jettyhan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
