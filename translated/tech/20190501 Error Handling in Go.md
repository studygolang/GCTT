# Error Handling in Go

## Mastering pragmatic error handling in your Go code

![images](https://cdn-images-1.medium.com/max/900/1*BmEMrWVjQVUs5bwWTn_AIg.png)

*这篇文章是“在你进入到Go的世界之前”系列中的一部分。在这里，我们可以一起探索Golang的世界，让你了解用Go语言编程时应注意到的小技巧并领悟Go语言的特性，让你学习Go语言的过程不再困难。*

我假设你已经有了一些Go语言的基础，不过当你遇到文章中你不熟悉的知识点的时候，可以随时停下来，查阅这些知识点之后，再回来继续读下去。

现在这些问题都讲清楚了，就让我们开始吧！

---

Go语言的错误处理方法是一个一直都颇受争议或是被误用的特性。在这篇文章里，你将会学到Go是如何处理错误的并理解他们的工作原理。你将会通过探索几种不同的方法、查看Go源码和一些标准库的细节，去理解错误是如何产生(**how errors work**)的以及如何处理他们。你将会了解Type Assertions在处理这些错误时所扮演的重要角色,以及将会在Go 2中发布的一些重要的错误处理模式的改变.

![images](https://cdn-images-1.medium.com/max/800/1*__fmJKbSA3D0HUVDqp56IA.jpeg)

### 介绍

起始阶段(First thing's first)：Go语言中的错误（Errors）**不**是异常（Exceptions），Dave Cheney 写了一个关于这个问题的[epic blog post](https://dave.cheney.net/2012/01/18/why-go-gets-exceptions-right)，我将在这里向你总结一下：在其它语言中，你无法确定一个函数是否会向你抛出一个异常（Exceptions）。相比于抛出一个异常，Go中的函数支持**返回多个值**，有一个约定俗成的用法是返回这个函数的结果并伴随一个错误（error）变量。

```go
func calculate(a, b int) (int, error) {
    // 一些代码
}
```

如果你的函数由于某些原因运行错误，你应当返回预先声明过的`error`类型。通常来讲，返回一个错误是在向函数调用者发出信号表明发生了一个错误，如果没有错误，就返回`nil`值。这样，你就让调用者知道发生了错误，并让调用者处理这个错误：函数的调用者应当在试图使用返回的值之前检查是否发生了错误。如果`error`不是`nil`，调用者有责任去检查这个错误并处理它（日志、返回错误、serve、尝试重新调用/清理机制等）。

```go
result, err := calculate(a, b)
if err != nil {
    // 处理异常
}
// 继续
```

这些片段在Go语言中非常常见，有些人认为它们是一大堆锅炉代码（a whole lot of boiler plate code？？？？）。编译器会将没有使用的变量视为编译错误，所以当你不打算去检查错误的时候，应该给返回的错误变量分配一个空白标识符`_`。但是无论这个方式多方便，都不应该忽视错误。

```go
// 在检查错误之前，结果无法被信任

result, _ := caculate(a, b)

if result >0 {
    // 忽视错误是不安全的，
    // 理论上讲，在你检查是否有异常之前，
    // 是无法相信你接收到的结果的
}
```

在Go语言严格的检查机制下，让一个函数返回结果的同时返回错误，可以让你更难写出含有错误的方法。你应当假设，函数的返回值是不正确的（损坏的）除非你检查了函数返回的错误值。如果将错误分配给了空白标识符，说明你忽略了你的函数值可能已经损坏。

![image](https://cdn-images-1.medium.com/max/1800/1*jDw9aGCJZWQhN_mOWRINew.jpeg)

Go语言确实有一个`panic`和`recover`机制，这再[另一篇Go博文](https://blog.golang.org/defer-panic-and-recover)中有详细的描述。但是这不意味着去模仿异常。用Dave的话说就是：“*当你在使用Go的时候产生`panic`，你会被吓坏，这不是其他人的问题，这是完蛋了，兄弟。*”他们非常的致命，并且会导致你的程序崩溃。Rob Pike创造了“*不要恐慌*”的谚语，这是不言自明的：你应当避免它，并返回错误。

- “错误就是价值观。”
- “不要只是检查错误，优雅地处理它们”
- “不要惊慌失措”  
[所有Rob Pike的Go谚语](https://go-proverbs.github.io/)

---

## 深入理解

### 关于错误的接口

在底层实现中，`error`类型是一个[普通的单方法接口](https://golang.org/ref/spec#Errors)，如果你还对他不熟悉，我强烈建议你仔细的阅读在Go官方博客中的[这篇文章](https://blog.golang.org/error-handling-and-go).

```Go
// error interface from the source code
type error interface {
    Error() string
}
```

实现你自己的错误类型非常容易，有非常多的方法能够让你实现`Error() string`方法的自定义结构体。任何实现了这个方法的结构体都会被视为一个合法的错误值同时可以被返回。

接下来，就让我们一起去探索这些途径。

### 内置的错误字符串（errorString）结构体

错误接口中最常用同时也是最出名的就是`errorString`结构体。这是你能想到的最简洁的实现。

```Go
package errors

func New(text string) error {
    return &errorString{text}
}

type errorString struct {
    s string
}

func (e *errorString) Error() string {
    return e.s
}

```

你可以在[这里](https://golang.org/src/errors/errors.go)看到它的简单实现。它做的事情就是保存一个`string`，同时，这个字符串是由`Error`方法返回的。我们可以使用数据格式化这个错误信息，比如，`fmt.Springf`。但除此之外，它不包含任何其他功能。如果你在使用内置的[`errors.New`](https://golang.org/src/errors/errors.go?s=293:353#L1)或者[`fmt.Errorf`](https://golang.org/src/fmt/print.go#L220)，你就[已经在使用他们了](https://play.golang.org/p/olRXqq3jNyR)。

```Go
import (
    "errors"
    "fmt"
)

func main() {
    e1 := errors.New(fmt.Sprintf("Could not open file"))
    e2 := fmt.Errorf("Could not open file")

    fmt.Println(fmt.Sprintf("Type of error 1: %T", e1))
    fmt.Println(fmt.Sprintf("Type of error 2: %T", e2))

    // output:
    // Type of error 1: *errors.errorString
    // Type of error 2: *errors.errorString
}
```

### github.com/pkg/errors

另一个简单的示例是[`pkg/errors`](https://github.com/pkg/errors/blob/master/errors.go)[包](https://github.com/pkg/errors/blob/master/errors.go)。不要与之前学到的内置`errors`包混淆这个包额外提供了一些重要的功能，比如错误的封装（wrapping）、展开（unwrapping），格式化和堆栈跟踪记录。你可以通过运行`go get github.com/pkg/errors`来安装这个包。

```Go
go get github.com/pkg/errors
```

如果需要将堆栈跟踪信息附加的错误中，或是附加必要的调试信息到错误中，可以使用此包的`New`或者`Errorf`函数，他们已经记录下了你的堆栈记录。同时，你还可以附加简单的元数据格式化功能。`Errorf`调用了[`fmt.Formatter`](https://golang.org/pkg/fmt/#Formatter)[接口](https://golang.org/pkg/fmt/#Formatter)，这意味着你可以使用`fmt`包的runes(`%s`, `%v`, `%+v` etc)来格式化他们。

```Go
import "github.com/pkg/errors"

// ...

errors.New("error writing to file")
// or, alternatively
errors.Errorf("error writing to file %s", f.Path)
```

这个包还包含`errors.Wrap`和`errors.Wrapf`函。这些函数将上下文添加到错误中，并在调用的时候跟踪堆栈信息。这样,你就可以将其与其上下文和重要的调试数据封装在一起,而不是简单地返回错误。

```Go
if err != nil {
    return errors.Wrap(err, "could not open file")
}
```

---

## 处理错误（Working with Errors)

### 类型断言

[类型断言](https://golang.org/ref/spec#Type_assertions)在处理错误的时候扮演者非常重要的角色。你需要使用它们来在接口值中断言信息，同时，由于错误处理的是`error`接口的自定义实现，所以在错误上实现断言是一个非常方便的方式。

它的语法对于所有的目标（purposes）都是相同的——`x.(T)`，其中`x`是接口类型。`x.(T)`断言`x`不为`nil`，并且存储在`x`中的值类型为`T`。在接下来的几节里面，你将会看到使用类型断言的两种方式——通过使用具体类型`T`和使用接口类型`T`。

```Go
var x interface{}
// short syntax, dropping the ok boolean
// panic: interface conversion: interface is nil, not string
s :+ x.(string)

// long syntax, with the ok boolean
if s ok := x.(string); ok {
    // does not panic, instead ok is set to false when assertion fails
    // we can now use s as string safely
}
```

>*playground: [short syntax panic](https://play.golang.org/p/bl-O3lJrixF), [safe long syntax](https://play.golang.org/p/CLLyXQWyrgF)*

---

*关于语法的附加说明:类型断言可以与短语法(当断言失败时，短语法会引发panic)和长语法(使用OK-boolean表示成功或失败)一起使用。我总是建议选择长语法的而不是短语法，因为我更喜欢检查OK变量而不是处理panic。*

---

### 使用接口类型T进行断言

---

via: <https://medium.com/gett-engineering/error-handling-in-go-53b8a7112d04>

作者：[Alon Abadi](https://medium.com/@alonabadi)
译者：[JoeyGaojingxing](https://github.com/JoeyGaojingxing)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
