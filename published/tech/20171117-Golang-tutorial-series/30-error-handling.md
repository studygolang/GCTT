已发布：https://studygolang.com/articles/12724

# 第 30 篇：错误处理

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 30 篇。

## 什么是错误？

错误表示程序中出现了异常情况。比如当我们试图打开一个文件时，文件系统里却并没有这个文件。这就是异常情况，它用一个错误来表示。

在 Go 中，错误一直是很常见的。错误用内建的 `error` 类型来表示。

就像其他的内建类型（如 `int`、`float64` 等），错误值可以存储在变量里、作为函数的返回值等等。

## 示例

现在我们开始编写一个示例，该程序试图打开一个并不存在的文件。

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    f, err := os.Open("/test.txt")
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(f.Name(), "opened successfully")
}
```

[在 playground 中运行](https://play.golang.org/p/yOhAviFM05)

在程序的第 9 行，我们试图打开路径为 `/test.txt` 的文件（playground 显然并不存在这个文件）。`os` 包里的 [`Open`](https://golang.org/pkg/os/#Open) 函数有如下签名：

```go
func Open(name string) (file *File, err error)
```

**如果成功打开文件，`Open` 函数会返回一个文件句柄（File Handler）和一个值为 `nil` 的错误。而如果打开文件时发生了错误，会返回一个不等于 `nil` 的错误**。

如果一个[函数](https://studygolang.com/articles/11892) 或[方法](https://studygolang.com/articles/12264) 返回了错误，按照惯例，错误会作为最后一个值返回。于是 `Open` 函数也是将 `err` 作为最后一个返回值。

**按照 Go 的惯例，在处理错误时，通常都是将返回的错误与 `nil` 比较。`nil` 值表示了没有错误发生，而非 `nil` 值表示出现了错误**。在这里，我们第 10 行检查了错误值是否为 `nil`。如果不是 `nil`，我们会简单地打印出错误，并在 `main` 函数中返回。

运行该程序会输出：

```
open /test.txt: No such file or directory
```

很棒！我们得到了一个错误，它指出该文件并不存在。

## 错误类型的表示

让我们进一步深入，理解 `error` 类型是如何定义的。`error` 是一个[接口](https://studygolang.com/articles/12266)类型，定义如下：

```go
type error interface {
    Error() string
}
```

`error` 有了一个签名为 `Error() string` 的方法。所有实现该接口的类型都可以当作一个错误类型。`Error()` 方法给出了错误的描述。

`fmt.Println` 在打印错误时，会在内部调用 `Error() string` 方法来得到该错误的描述。上一节示例中的第 11 行，就是这样打印出错误的描述的。

## 从错误获取更多信息的不同方法

现在，我们知道了 `error` 是一个接口类型，让我们看看如何从一个错误获取更多信息。

在前面的示例里，我们只是打印出错误的描述。如果我们想知道这个错误的文件路径，该怎么做呢？一种选择是直接解析错误的字符串。这是前面示例的输出：

```
open /test.txt: No such file or directory
```

**我们解析了这条错误信息，虽然获取了发生错误的文件路径，但是这种方法很不优雅。随着语言版本的更新，这条错误的描述随时都有可能变化，使我们程序出错**。

有没有更加可靠的方法来获取文件名呢？答案是肯定的，这是可以做到的，Go 标准库给出了各种提取错误相关信息的方法。我们一个个来看看吧。

### 1. 断言底层结构体类型，使用结构体字段获取更多信息

如果你仔细阅读了 [`Open`](https://golang.org/pkg/os/#OpenFile) 函数的文档，你可以看见它返回的错误类型是 `*PathError`。[`PathError`](https://golang.org/pkg/os/#PathError) 是[结构体](https://studygolang.com/articles/12263)类型，它在标准库中的实现如下：

```go
type PathError struct {
    Op   string
    Path string
    Err  error
}

func (e *PathError) Error() string { return e.Op + " " + e.Path + ": " + e.Err.Error() }
```

如果你有兴趣了解上述源代码出现的位置，可以在这里找到：https://golang.org/src/os/error.go?s=653:716#L11。

通过上面的代码，你就知道了 `*PathError` 通过声明 `Error() string` 方法，实现了 `error` 接口。`Error() string` 将文件操作、路径和实际错误拼接，并返回该字符串。于是我们得到该错误信息：

```
open /test.txt: No such file or directory
```

结构体 `PathError` 的 `Path` 字段，就有导致错误的文件路径。我们修改前面写的程序，打印出该路径。

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    f, err := os.Open("/test.txt")
    if err, ok := err.(*os.PathError); ok {
        fmt.Println("File at path", err.Path, "failed to open")
        return
    }
    fmt.Println(f.Name(), "opened successfully")
}
```

[在 playground 上运行](https://play.golang.org/p/JQrqWU7Jf9)

在上面的程序里，我们在第 10 行使用了[类型断言](https://studygolang.com/articles/12266)（Type Assertion）来获取 `error` 接口的底层值（Underlying Value）。接下来在第 11 行，我们使用 `err.Path` 来打印该路径。该程序会输出：

```
File at path /test.txt failed to open
```

很棒！我们已经使用类型断言成功获取到了该错误的文件路径。

### 2. 断言底层结构体类型，调用方法获取更多信息

第二种获取更多错误信息的方法，也是对底层类型进行断言，然后通过调用该结构体类型的方法，来获取更多的信息。

我们通过一个实例来理解这一点。

标准库中的 `DNSError` 结构体类型定义如下：

```go
type DNSError struct {
    ...
}

func (e *DNSError) Error() string {
    ...
}
func (e *DNSError) Timeout() bool {
    ...
}
func (e *DNSError) Temporary() bool {
    ...
}
```

从上述代码可以看到，`DNSError` 结构体还有 `Timeout() bool` 和 `Temporary() bool` 两个方法，它们返回一个布尔值，指出该错误是由超时引起的，还是临时性错误。

接下来我们编写一个程序，断言 `*DNSError` 类型，并调用这些方法来确定该错误是临时性错误，还是由超时导致的。

```go
package main

import (
    "fmt"
    "net"
)

func main() {
    addr, err := net.LookupHost("golangbot123.com")
    if err, ok := err.(*net.DNSError); ok {
        if err.Timeout() {
            fmt.Println("operation timed out")
        } else if err.Temporary() {
            fmt.Println("temporary error")
        } else {
            fmt.Println("generic error: ", err)
        }
        return
    }
    fmt.Println(addr)
}
```

**注：在 playground 无法进行 DNS 解析。请在你的本地运行该程序**。

在上述程序中，我们在第 9 行，试图获取 `golangbot123.com`（无效的域名） 的 ip。在第 10 行，我们通过 `*net.DNSError` 的类型断言，获取到了错误的底层值。接下来的第 11 行和第 13 行，我们分别检查了该错误是由超时引起的，还是一个临时性错误。

在本例中，我们的错误既不是临时性错误，也不是由超时引起的，因此该程序输出：

```
generic error:  lookup golangbot123.com: no such host
```

如果该错误是临时性错误，或是由超时引发的，那么对应的 if 语句会执行，于是我们就可以适当地处理它们。

### 3. 直接比较

第三种获取错误的更多信息的方式，是与 `error` 类型的变量直接比较。我们通过一个示例来理解。

`filepath` 包中的 [`Glob`](https://golang.org/pkg/path/filepath/#Glob) 用于返回满足 glob 模式的所有文件名。如果模式写的不对，该函数会返回一个错误 `ErrBadPattern`。

`filepath` 包中的 `ErrBadPattern` 定义如下：

```go
var ErrBadPattern = errors.New("syntax error in pattern")
```

`errors.New()` 用于创建一个新的错误。我们会在下一教程中详细讨论它。

当模式不正确时，`Glob` 函数会返回 `ErrBadPattern`。

我们来写一个小程序来看看这个错误。

```go
package main

import (
    "fmt"
    "path/filepath"
)

func main() {
    files, error := filepath.Glob("[")
    if error != nil && error == filepath.ErrBadPattern {
        fmt.Println(error)
        return
    }
    fmt.Println("matched files", files)
}
```

[在 playground 上运行](https://play.golang.org/p/zbVDDHnMZU)

在上述程序里，我们查询了模式为 `[` 的文件，然而这个模式写的不正确。我们检查了该错误是否为 `nil`。为了获取该错误的更多信息，我们在第 10 行将 `error` 直接与 `filepath.ErrBadPattern` 相比较。如果该条件满足，那么该错误就是由模式错误导致的。该程序会输出：

```
syntax error in pattern
```

标准库在提供错误的详细信息时，使用到了上述提到的三种方法。在下一教程里，我们会通过这些方法来创建我们自己的自定义错误。

## 不可忽略错误

绝不要忽略错误。忽视错误会带来问题。接下来我重写上面的示例，在列出所有满足模式的文件名时，我省略了错误处理的代码。

```go
package main

import (
    "fmt"
    "path/filepath"
)

func main() {
    files, _ := filepath.Glob("[")
    fmt.Println("matched files", files)
}
```

[在 playground 上运行](https://play.golang.org/p/2k8r_Qg_lc)

我们已经从前面的示例知道了这个模式是错误的。在第 9 行，通过使用 `_` 空白标识符，我忽略了 `Glob` 函数返回的错误。我在第 10 行简单打印了所有匹配的文件。该程序会输出：

```
matched files []
```

由于我忽略了错误，输出看起来就像是没有任何匹配了 glob 模式的文件，但实际上这是因为模式的写法不对。所以绝不要忽略错误。

本教程到此结束。

这一教程我们讨论了该如何处理程序中出现的错误，也讨论了如何查询关于错误的更多信息。简单概括一下本教程讨论的内容：

- 什么是错误？
- 错误的表示
- 获取错误详细信息的各种方法
- 不能忽视错误

在下一教程，我们会创建我们自己的自定义错误，并给标准错误增加更多的语境（Context）。

祝你愉快。

**上一教程 - [Defer](https://studygolang.com/articles/12719)**

**下一教程 - 自定义错误（暂未发布，敬请期待）**

---

via: https://golangbot.com/error-handling/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
