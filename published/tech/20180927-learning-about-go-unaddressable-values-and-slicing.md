首发于：https://studygolang.com/articles/26301

# 了解 Go 的不可寻址值和切片

Dave Cheney 最近在 Twitter 上发布了一个 Go 的小测验，和往常一样，我从中学到了一些有趣的东西。让我们从他的推文开始：

`#golang` 小测验：该程序打印什么？

```go
package main
import (
    "crypto/sha1"
    "fmt"
)

func main() {
    input := []byte("Hello, playground")
    hash := sha1.Sum(input)[:5]
    fmt.Println(hash)
}
```

令我惊讶的是，答案是：

```bash
./test.go:10:28: invalid operation sha1.Sum(input)[:5] (slice of unaddressable value)
```

我们收到此错误有三个原因。首先，[`sha1.Sum()`](https://golang.org/pkg/crypto/sha1/#Sum) 的返回值不寻常。大多数方法返回切片，而此代码对切片不会报错。但是 `sha1.Sum()` 返回的值很奇怪，它是一个固定大小的数组（具体来说是 `[20]byte` ），由于 Go 是返回数值的，这意味着它确实向 `main()` 返回了 20 字节的数组，而不是指向它的指针。

这就涉及到了不可寻址值的概念，与可寻址值相反。详细的介绍在 Go 编程语言规范的 [地址运算符](https://golang.org/ref/spec#Address_operators) 中。简单来说，大多数匿名值都不可寻址（ [复合字面值](https://golang.org/ref/spec#Composite_literals) 是一个大大的例外）。在上面的代码中，`sha1.Sum()` 的返回值是匿名的，因为我们立即对其进行了切片操作。如果我们将它存在变量中，并因此使其变为非匿名，则该代码不会报错：

```go
tmp := sha1.Sum(input)
hash := tmp[:5]
```

最后一个问题是为什么切片操作是错误的。这是因为对数组进行切片操作要求该数组是可寻址的（在 Go 编程语言规范的 [Slice 表达式](https://golang.org/ref/spec#Slice_expressions) 的末尾介绍）。`sha1.Sum()` 返回的匿名数组是不可寻址的，因此对其进行切片会被编译器拒绝。

（将返回值存储到我们的 tmp 变量中使其变成了可寻址。 `sha1.Sum()` 的返回值在复制到 `tmp` 后就消失了。）

虽然我不能完全理解为什么 Go 的设计师限制了哪些值是可寻址的，但是我可以想到几条原因。例如，如果在这里允许切片操作，那么 Go 会默默地实现堆存储以容纳 `sha1.Sum()` 的返回值（然后将该值复制到另一个值），该返回值将一直存在直到那个切片被回收。

（如 [x86-64 上的 Go 低级调用惯例](https://science.raphael.poss.name/go-calling-convention-x86-64.html#arguments-and-return-value) 中所述，由于 Go 返回了栈中的所有值，因此需要将数据进行拷贝。对于 `sha1.Sum()` 的 20 字节的返回值来说，这并不是什么大事。我很确定人们经常使用更大的结构体作为返回值。）

PS： Go 语言规范中的许多内容要求或仅对可寻址的值适用。例如，大多数 [赋值](https://golang.org/ref/spec#Assignments) 操作需要可寻址性。

## 补充：方法调用和可寻址性

假设有一个类型 `T`，并且在 `*T` 上定义了一些方法，例如 `*T.Op()`。就像 Go 允许在不取消引用指针的情况下进行字段引用一样，你可以在非指针值上调用指针方法：

```go
var x T
x.Op()
```

这是 `(&x).Op()` 的简便写法（在 Go 编程语言规范文中靠后的 [调用](https://golang.org/ref/spec#Calls) 部分进行了介绍）。但是，由于此简便写法需要获取地址，因此需要可寻址性。因此，以下操作会报错：

```go
// afunc() 返回一个 T
afunc().Op()

// 但是这个可以运行:
var x T = afunc()
x.Op()
```

之前我已经看到人们在讨论 Go 在方法调用上的怪癖，但是当时我还不完全了解发生了什么，以及由于什么原因使方法调用无法正常工作。

（请注意，这种简写转换与 `*T` 具有所有 T 方法是根本不同的，这在我 [之前的一篇文章](https://utcc.utoronto.ca/~cks/space/blog/programming/GoInterfacesAutogenFuncs) 中提到过）

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoUnaddressableSlice

作者：[Chris wSiebenmann](https://utcc.utoronto.ca/~cks/space/People/ChrisSiebenmann)
译者：[zhiyu-tracy-yang](https://github.com/zhiyu-tracy-yang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
