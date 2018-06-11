已发布：https://studygolang.com/articles/12785

# 第 32 篇：panic 和 recover

![panic 和 recover](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-series/panic-recover-golang-2-2.png)

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 32 篇。

## 什么是 panic？

在 Go 语言中，程序中一般是使用[错误](https://studygolang.com/articles/12724)来处理异常情况。对于程序中出现的大部分异常情况，错误就已经够用了。

但在有些情况，当程序发生异常时，无法继续运行。在这种情况下，我们会使用 `panic` 来终止程序。当[函数](https://studygolang.com/articles/11892)发生 panic 时，它会终止运行，在执行完所有的[延迟](https://studygolang.com/articles/12719)函数后，程序控制返回到该函数的调用方。这样的过程会一直持续下去，直到当前[协程](https://studygolang.com/articles/12342)的所有函数都返回退出，然后程序会打印出 panic 信息，接着打印出堆栈跟踪（Stack Trace），最后程序终止。在编写一个示例程序后，我们就能很好地理解这个概念了。

在本教程里，我们还会接着讨论，当程序发生 panic 时，使用 `recover` 可以重新获得对该程序的控制。

可以认为 `panic` 和 `recover` 与其他语言中的 `try-catch-finally` 语句类似，只不过一般我们很少使用 `panic` 和 `recover`。而当我们使用了 `panic` 和 `recover` 时，也会比 `try-catch-finally` 更加优雅，代码更加整洁。

## 什么时候应该使用 panic？

**需要注意的是，你应该尽可能地使用[错误](https://studygolang.com/articles/12724)，而不是使用 panic 和 recover。只有当程序不能继续运行的时候，才应该使用 panic 和 recover 机制**。

panic 有两个合理的用例。

1. **发生了一个不能恢复的错误，此时程序不能继续运行**。
    一个例子就是 web 服务器无法绑定所要求的端口。在这种情况下，就应该使用 panic，因为如果不能绑定端口，啥也做不了。

2. **发生了一个编程上的错误**。
    假如我们有一个接收指针参数的方法，而其他人使用 `nil` 作为参数调用了它。在这种情况下，我们可以使用 panic，因为这是一个编程错误：用 `nil` 参数调用了一个只能接收合法指针的方法。

## panic 示例

内建函数 `panic` 的签名如下所示：

```go
func panic(interface{})
```

当程序终止时，会打印传入 `panic` 的参数。我们写一个示例，你就会清楚它的用途了。我们现在就开始吧。

我们会写一个例子，来展示 `panic` 如何工作。

```go
package main

import (
    "fmt"
)

func fullName(firstName *string, lastName *string) {
    if firstName == nil {
        panic("runtime error: first name cannot be nil")
    }
    if lastName == nil {
        panic("runtime error: last name cannot be nil")
    }
    fmt.Printf("%s %s\n", *firstName, *lastName)
    fmt.Println("returned normally from fullName")
}

func main() {
    firstName := "Elon"
    fullName(&firstName, nil)
    fmt.Println("returned normally from main")
}
```

[在 playground 上运行](https://play.golang.org/p/xQJYRSCu8S)

上面的程序很简单，会打印一个人的全名。第 7 行的 `fullName` 函数会打印出一个人的全名。该函数在第 8 行和第 11 行分别检查了 `firstName` 和 `lastName` 的指针是否为 `nil`。如果是 `nil`，`fullName` 函数会调用含有不同的错误信息的 `panic`。当程序终止时，会打印出该错误信息。

运行该程序，会有如下输出：

```
panic: runtime error: last name cannot be nil

goroutine 1 [running]:
main.fullName(0x1040c128, 0x0)
    /tmp/sandbox135038844/main.go:12 +0x120
main.main()
    /tmp/sandbox135038844/main.go:20 +0x80
```

我们来分析这个输出，理解一下 panic 是如何工作的，并且思考当程序发生 panic 时，会怎样打印堆栈跟踪。

在第 19 行，我们将 `Elon` 赋值给了 `firstName`。在第 20 行，我们调用了 `fullName` 函数，其中 `lastName` 等于 `nil`。因此，满足了第 11 行的条件，程序发生 panic。当出现了 panic 时，程序就会终止运行，打印出传入 panic 的参数，接着打印出堆栈跟踪。因此，第 14 行和第 15 行的代码并不会在发生 panic 之后执行。程序首先会打印出传入 `panic` 函数的信息：

```
panic: runtime error: last name cannot be empty
```

接着打印出堆栈跟踪。

程序在 `fullName` 函数的第 12 行发生 panic，因此，首先会打印出如下所示的输出。

```
main.fullName(0x1040c128, 0x0)
    /tmp/sandbox135038844/main.go:12 +0x120
```

接着会打印出堆栈的下一项。在本例中，堆栈跟踪中的下一项是第 20 行（因为发生 panic 的 `fullName` 调用就在这一行），因此接下来会打印出：

```
main.main()
    /tmp/sandbox135038844/main.go:20 +0x80
```

现在我们已经到达了导致 panic 的顶层函数，这里没有更多的层级，因此结束打印。

## 发生 panic 时的 defer

我们重新总结一下 panic 做了什么。**当函数发生 panic 时，它会终止运行，在执行完所有的延迟函数后，程序控制返回到该函数的调用方。这样的过程会一直持续下去，直到当前协程的所有函数都返回退出，然后程序会打印出 panic 信息，接着打印出堆栈跟踪，最后程序终止**。

在上面的例子中，我们没有延迟调用任何函数。如果有延迟函数，会先调用它，然后程序控制返回到函数调用方。

我们来修改上面的示例，使用一个延迟语句。

```go
package main

import (
    "fmt"
)

func fullName(firstName *string, lastName *string) {
    defer fmt.Println("deferred call in fullName")
    if firstName == nil {
        panic("runtime error: first name cannot be nil")
    }
    if lastName == nil {
        panic("runtime error: last name cannot be nil")
    }
    fmt.Printf("%s %s\n", *firstName, *lastName)
    fmt.Println("returned normally from fullName")
}

func main() {
    defer fmt.Println("deferred call in main")
    firstName := "Elon"
    fullName(&firstName, nil)
    fmt.Println("returned normally from main")
}
```

[在 playground 上运行](https://play.golang.org/p/oUFnu-uTmC)

上述代码中，我们只修改了两处，分别在第 8 行和第 20 行添加了延迟函数的调用。

该函数会打印：

```
This program prints,

deferred call in fullName
deferred call in main
panic: runtime error: last name cannot be nil

goroutine 1 [running]:
main.fullName(0x1042bf90, 0x0)
    /tmp/sandbox060731990/main.go:13 +0x280
main.main()
    /tmp/sandbox060731990/main.go:22 +0xc0
```

当程序在第 13 行发生 panic 时，首先执行了延迟函数，接着控制返回到函数调用方，调用方的延迟函数继续运行，直到到达顶层调用函数。

在我们的例子中，首先执行 `fullName` 函数中的 `defer` 语句（第 8 行）。程序打印出：

```
deferred call in fullName
```

接着程序返回到 `main` 函数，执行了 `main` 函数的延迟调用，因此会输出：

```
deferred call in main
```

现在程序控制到达了顶层函数，因此该函数会打印出 panic 信息，然后是堆栈跟踪，最后终止程序。

## recover

`recover` 是一个内建函数，用于重新获得 panic 协程的控制。

`recover` 函数的标签如下所示：

```go
func recover() interface{}
```

只有在延迟函数的内部，调用 `recover` 才有用。在延迟函数内调用 `recover`，可以取到 `panic` 的错误信息，并且停止 panic 续发事件（Panicking Sequence），程序运行恢复正常。如果在延迟函数的外部调用 `recover`，就不能停止 panic 续发事件。

我们来修改一下程序，在发生 panic 之后，使用 `recover` 来恢复正常的运行。

```go
package main

import (
    "fmt"
)

func recoverName() {
    if r := recover(); r!= nil {
        fmt.Println("recovered from ", r)
    }
}

func fullName(firstName *string, lastName *string) {
    defer recoverName()
    if firstName == nil {
        panic("runtime error: first name cannot be nil")
    }
    if lastName == nil {
        panic("runtime error: last name cannot be nil")
    }
    fmt.Printf("%s %s\n", *firstName, *lastName)
    fmt.Println("returned normally from fullName")
}

func main() {
    defer fmt.Println("deferred call in main")
    firstName := "Elon"
    fullName(&firstName, nil)
    fmt.Println("returned normally from main")
}
```

[在 playground 上运行](https://play.golang.org/p/I9pp8N55c1)

在第 7 行，`recoverName()` 函数调用了 `recover()`，返回了调用 `panic` 的传参。在这里，我们只是打印出 `recover` 的返回值（第 8 行）。在 `fullName` 函数内，我们在第 14 行延迟调用了 `recoverNames()`。

当 `fullName` 发生 panic 时，会调用延迟函数 `recoverName()`，它使用了 `recover()` 来停止 panic 续发事件。

该程序会输出：

```
recovered from  runtime error: last name cannot be nil
returned normally from main
deferred call in main
```

当程序在第 19 行发生 panic 时，会调用延迟函数 `recoverName`，它反过来会调用 `recover()` 来重新获得 panic 协程的控制。第 8 行调用了 `recover`，返回了 `panic` 的传参，因此会打印：

```
recovered from  runtime error: last name cannot be nil
```

在执行完 `recover()` 之后，panic 会停止，程序控制返回到调用方（在这里就是 `main` 函数），程序在发生 panic 之后，从第 29 行开始会继续正常地运行。程序会打印 `returned normally from main`，之后是 `deferred call in main`。

## panic，recover 和 Go 协程

只有在相同的 [Go 协程](https://studygolang.com/articles/12342)中调用 recover 才管用。`recover` 不能恢复一个不同协程的 panic。我们用一个例子来理解这一点。

```go
package main

import (
    "fmt"
    "time"
)

func recovery() {
    if r := recover(); r != nil {
        fmt.Println("recovered:", r)
    }
}

func a() {
    defer recovery()
    fmt.Println("Inside A")
    go b()
    time.Sleep(1 * time.Second)
}

func b() {
    fmt.Println("Inside B")
    panic("oh! B panicked")
}

func main() {
    a()
    fmt.Println("normally returned from main")
}
```

[在 playground 上运行](https://play.golang.org/p/pEVzTLz36Y)

在上面的程序中，函数 `b()` 在第 23 行发生 panic。函数 `a()` 调用了一个延迟函数 `recovery()`，用于恢复 panic。在第 17 行，函数 `b()` 作为一个不同的协程来调用。下一行的 `Sleep` 只是保证 `a()` 在 `b()` 运行结束之后才退出。

你认为程序会输出什么？panic 能够恢复吗？答案是否定的，panic 并不会恢复。因为调用 `recovery` 的协程和 `b()` 中发生 panic 的协程并不相同，因此不可能恢复 panic。

运行该程序会输出：

```
Inside A
Inside B
panic: oh! B panicked

goroutine 5 [running]:
main.b()
    /tmp/sandbox388039916/main.go:23 +0x80
created by main.a
    /tmp/sandbox388039916/main.go:17 +0xc0
```

从输出可以看出，panic 没有恢复。

如果函数 `b()` 在相同的协程里调用，panic 就可以恢复。

如果程序的第 17 行由 `go b()` 修改为 `b()`，就可以恢复 panic 了，因为 panic 发生在与 recover 相同的协程里。如果运行这个修改后的程序，会输出：

```
Inside A
Inside B
recovered: oh! B panicked
normally returned from main
```

## 运行时 panic

运行时错误（如数组越界）也会导致 panic。这等价于调用了内置函数 `panic`，其参数由接口类型 [runtime.Error](https://golang.org/src/runtime/error.go?s=267:503#L1) 给出。`runtime.Error` 接口的定义如下：

```go
type Error interface {
    error
    // RuntimeError is a no-op function but
    // serves to distinguish types that are run time
    // errors from ordinary errors: a type is a
    // run time error if it has a RuntimeError method.
    RuntimeError()
}
```

而 `runtime.Error` 接口满足内建接口类型 [`error`](https://golangbot.com/error-handling/#errortyperepresentation)。

我们来编写一个示例，创建一个运行时 panic。

```go
package main

import (
    "fmt"
)

func a() {
    n := []int{5, 7, 4}
    fmt.Println(n[3])
    fmt.Println("normally returned from a")
}
func main() {
    a()
    fmt.Println("normally returned from main")
}
```

[在 playground 上运行](https://play.golang.org/p/CBsK2xXzGg)

在上面的程序中，第 9 行我们试图访问 `n[3]`，这是一个对[切片](https://studygolang.com/articles/12121)的错误引用。该程序会发生 panic，输出如下：

```
panic: runtime error: index out of range

goroutine 1 [running]:
main.a()
    /tmp/sandbox780439659/main.go:9 +0x40
main.main()
    /tmp/sandbox780439659/main.go:13 +0x20
```

你也许想知道，是否可以恢复一个运行时 panic？当然可以！我们来修改一下上面的代码，恢复这个 panic。

```go
package main

import (
    "fmt"
)

func r() {
    if r := recover(); r != nil {
        fmt.Println("Recovered", r)
    }
}

func a() {
    defer r()
    n := []int{5, 7, 4}
    fmt.Println(n[3])
    fmt.Println("normally returned from a")
}

func main() {
    a()
    fmt.Println("normally returned from main")
}
```

[在 playground 上运行](https://play.golang.org/p/qusvZe5rft)

运行上面程序会输出：

```
Recovered runtime error: index out of range
normally returned from main
```

从输出可以知道，我们已经恢复了这个 panic。

## 恢复后获得堆栈跟踪

当我们恢复 panic 时，我们就释放了它的堆栈跟踪。实际上，在上述程序里，恢复 panic 之后，我们就失去了堆栈跟踪。

有办法可以打印出堆栈跟踪，就是使用 [`Debug`](https://golang.org/pkg/runtime/debug/) 包中的 [`PrintStack`](https://golang.org/pkg/runtime/debug/#PrintStack) 函数。

```go
package main

import (
    "fmt"
    "runtime/debug"
)

func r() {
    if r := recover(); r != nil {
        fmt.Println("Recovered", r)
        debug.PrintStack()
    }
}

func a() {
    defer r()
    n := []int{5, 7, 4}
    fmt.Println(n[3])
    fmt.Println("normally returned from a")
}

func main() {
    a()
    fmt.Println("normally returned from main")
}
```

[在 playground 上运行](https://play.golang.org/p/D-QlDmumHV)

在上面的程序中，我们在第 11 行使用了 `debug.PrintStack()` 打印堆栈跟踪。

该程序会输出：

```
Recovered runtime error: index out of range
goroutine 1 [running]:
runtime/debug.Stack(0x1042beb8, 0x2, 0x2, 0x1c)
    /usr/local/go/src/runtime/debug/stack.go:24 +0xc0
runtime/debug.PrintStack()
    /usr/local/go/src/runtime/debug/stack.go:16 +0x20
main.r()
    /tmp/sandbox949178097/main.go:11 +0xe0
panic(0xf0a80, 0x17cd50)
    /usr/local/go/src/runtime/panic.go:491 +0x2c0
main.a()
    /tmp/sandbox949178097/main.go:18 +0x80
main.main()
    /tmp/sandbox949178097/main.go:23 +0x20
normally returned from main
```

从输出我们可以看出，首先已经恢复了 panic，打印出 `Recovered runtime error: index out of range`。此外，我们也打印出了堆栈跟踪。在恢复了 panic 之后，还打印出 `normally returned from main`。

本教程到此结束。

简单概括一下本教程讨论的内容：

- 什么是 panic？
- 什么时候应该使用 panic？
- panic 示例
- 发生 panic 时的 defer
- recover
- panic，recover 和 Go 协程
- 运行时 panic
- 恢复后获得堆栈跟踪

祝你愉快。

**上一教程 - [自定义错误](https://studygolang.com/articles/12784)**

**下一教程 - 函数是一等公民（暂未发布，敬请期待）**

---

via: https://golangbot.com/panic-and-recover/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
