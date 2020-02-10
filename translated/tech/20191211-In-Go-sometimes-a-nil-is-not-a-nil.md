# 在 Go 语言中，有时 nil 并不是一个 nil

今天，我遇到了一个 [Go FAQ](http://golang.org/doc/faq#nil_error)。首先，作为一个小小的 Go 语言测验，看看您是否在 Go playground 中运行该程序之前就能推断出它应该打印出的内容（我已经将程序放在侧边栏中，以防它在 Go playground 上消失）。该程序的关键代码是：

```go
type fake struct { io.Writer }

func fred (logger io.Writer) {
   if logger != nil {
      logger.Write([]byte("..."))
   }
}

func main() {
   var lp *fake
   fred(nil)
   fred(lp)
}
```

由于 Go 语言中的变量是使用它们的零值显式创建的，在指针的情况下，例如 `lp` 将会是 `nil`，您可能期待上述代码会正常运行（即不执行任何操作）。实际上，它会在对 `fred()` 的第二次调用时崩溃。原因是，在 Go 语言中，有时以 `nil` 为值的变量，如果直接打印的话，它虽然看起来像 `nil`，但实际上并不是真的 `nil` 。简而言之，Go 语言区别对待 `nil` 接口值和转换为接口的值为 `nil` 的具体类型。只有前者确实为 `nil`，因此与字面上的 `ni​​l` 相等，就像 `fred()` 在这里做的一样。

（因此，可以使用 `nil f` 调用 `(f *fake)` 上的具体方法。它也许是一个 `nil` 指针，但是它是类型化的 `nil` 指针，所以可以拥有有方法。甚至在接口转换后依然可以拥有方法，正如上述的例子。）

对于这里的情况，其解决方法是更改​​初始化的过程。实际的程序条件性地设置了 `fake`，类似于下面的代码：

```go
var l *sLogger

if smtplog != nil {
    l = &sLogger
    l.prefix = logpref
    l.writer = bufio.NewWriterSize(smtplog, 4096)
}
convo = smtpd.NewConvo(conn, l)
```

这会将具体类型为 `*sLogger` 的 `nil` 传递给期望参数为 `io.Writer` 的对象，从而导致接口转换并掩盖了 `nil`。为了解决这个问题，我们可以添加一个必须显式设置的中间变量 `io.Writer`：

```go
var l2 io.Writer

if smtplog != nil {
    l := &sLogger
    l.prefix = logpref
    l.writer = ....
    l2 = l
}
convo = smtpd.NewConvo(conn, l2)

```

如果我们不初始化这个特殊的日志记录器 `sLogger`，则 `l2` 会是一个真正的 `io.Writer nil`，并会在 `smtpd` 包中被检测到。

（您可以将类似的初始化操作封装进一个返回类型为 `io.Writer` 的函数中，并在没有提供日志记录器的情况下显式返回 `nil`，通过这样的技巧来达到类似的效果。需要强调的一点是，函数必须返回接口类型，如果返回类型为 `*sLogger`，那么您将再次遇到相同的问题。）

在 `sLogger` 的方法中保留对零值的防护代码，这是一个个人喜好问题。然而，我不想这么做，如果将来我在代码中遇到类似的初始化错误，我希望它崩溃，以便对其进行修复。

我从这件事中学到的另一个教训是，如果是出于调试的目的而进行的打印，我不会再使用 `%v` 作为格式说明符，而会使用 `%#v`。因为前者将会为接口 `nil` 和具体类型的 `nil` 同样打印一个普通且具有误导性的 `<nil>`，而 `%#v` 将为前者打印出 `<nil>`，为后者打印 `(*main.fake)(nil)` 。

## 边注栏: 测试程序

```go
package main

import (
    "fmt"
    "io"
)

type fake struct {
    io.Writer
}

func fred(logger io.Writer) {
    if logger != nil {
        logger.Write([]byte("a test\n"))
    }
}

func main() {
    // 这里的 t 的值是 nil
    var t *fake

    fred(nil)
    fmt.Printf("passed 1\n")
    fred(t)
    fmt.Printf("passed 2\n")
}
```

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoNilNotNil

作者：[ChrisSiebenmann](https://utcc.utoronto.ca/~cks/space/People/ChrisSiebenmann)
译者：[anxk](https://github.com/anxk)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
