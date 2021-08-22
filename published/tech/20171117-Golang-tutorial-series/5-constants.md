已发布：https://studygolang.com/articles/11872

# 第 5 部分: 常量

这是我们 Golang 系列教程的第五篇。

## 定义

在 Go 语言中，术语"常量"用于表示固定的值。比如 `5` 、`-89`、 `I love Go`、`67.89` 等等。

看看下面的代码:

```go
var a int = 50
var b string = "I love Go"
```
**在上面的代码中，变量 `a` 和 `b` 分别被赋值为常量 `50` 和 `I love GO`**。关键字 `const` 被用于表示常量，比如 `50` 和 `I love Go`。即使在上面的代码中我们没有明确的使用关键字 `const`，但是在 Go 的内部，它们是常量。

顾名思义，常量不能再重新赋值为其他的值。因此下面的程序将不能正常工作，它将出现一个编译错误: `cannot assign to a.`。

```go
package main

func main() {
    const a = 55 // 允许
    a = 89       // 不允许重新赋值
}
```
[在线运行程序](https://play.golang.org/p/b2J8_UQobb)

常量的值会在编译的时候确定。因为函数调用发生在运行时，所以不能将函数的返回值赋值给常量。

```go
package main

import (
    "fmt"
    "math"
)

func main() {
    fmt.Println("Hello, playground")
    var a = math.Sqrt(4)   // 允许
    const b = math.Sqrt(4) // 不允许
}
```
[在线运行程序](https://play.golang.org/p/dCON1LzCTw)

在上面的程序中，因为 `a` 是变量，因此我们可以将函数 `math.Sqrt(4)` 的返回值赋值给它（我们将在单独的地方详细讨论函数）。

 `b` 是一个常量，它的值需要在编译的时候就确定。函数 `math.Sqrt(4)` 只会在运行的时候计算，因此 `const b = math.Sqrt(4)` 将会抛出错误 `error main.go:11: const initializer math.Sqrt(4) is not a constant)`

## 字符串常量

双引号中的任何值都是 Go 中的字符串常量。例如像 `Hello World` 或 `Sam` 等字符串在 Go 中都是常量。

什么类型的字符串属于常量？答案是他们是无类型的。

像 `Hello World` 这样的字符串常量没有任何类型。

```go
const hello = "Hello World"
```

上面的例子，我们把 `Hello World` 分配给常量 `hello`。现在常量 `hello` 有类型吗？答案是没有。常量仍然没有类型。

Go 是一门强类型语言，所有的变量必须有明确的类型。那么, 下面的程序是如何将无类型的常量 `Sam` 赋值给变量 `name` 的呢？

```go
package main

import (
    "fmt"
)

func main() {
    var name = "Sam"
    fmt.Printf("type %T value %v", name, name)

}
```
[在线运行程序](https://play.golang.org/p/xhYV4we_Jz)

**答案是无类型的常量有一个与它们相关联的默认类型，并且当且仅当一行代码需要时才提供它。在声明中 `var name = "Sam"` ， `name` 需要一个类型，它从字符串常量 `Sam` 的默认类型中获取。**

有没有办法创建一个带类型的常量？答案是可以的。以下代码创建一个有类型常量。

```go
const typedhello string = "Hello World"
```
上面代码中， `typedhello` 就是一个 `string` 类型的常量。

Go 是一个强类型的语言，在分配过程中混合类型是不允许的。让我们通过以下程序看看这句话是什么意思。

```go
package main

func main() {
        var defaultName = "Sam" // 允许
        type myString string
        var customName myString = "Sam" // 允许
        customName = defaultName // 不允许

}
```
[在线运行程序](https://play.golang.org/p/1Q-vudNn_9)

在上面的代码中，我们首先创建一个变量 `defaultName` 并分配一个常量 `Sam` 。**常量 `Sam` 的默认类型是 `string` ，所以在赋值后 `defaultName` 是 `string` 类型的。**

下一行，我们将创建一个新类型 `myString`，它的底层类型是 `string`（译注：原文说是别名，是不对的）。

然后我们创建一个 `myString` 的变量 `customName` 并且给他赋值一个常量 `Sam` 。因为常量 `Sam` 是无类型的，它可以分配给任何字符串变量。因此这个赋值是允许的，`customName` 的类型是 `myString`。

现在，我们有一个类型为 `string` 的变量 `defaultName` 和另一个类型为 `myString` 的变量 `customName`。即使我们知道这个 `myString` 是 `string` 类型的别名。Go 的类型策略不允许将一种类型的变量赋值给另一种类型的变量。因此将 `defaultName` 赋值给 `customName` 是不允许的，编译器会抛出一个错误 `main.go:7:20: cannot use defaultName (type string) as type myString in assignmen`。

## 布尔常量

布尔常量和字符串常量没有什么不同。他们是两个无类型的常量 `true` 和 `false`。字符串常量的规则适用于布尔常量，所以在这里我们不再重复。以下是解释布尔常量的简单程序。

```go
package main

func main() {
    const trueConst = true
    type myBool bool
    var defaultBool = trueConst // 允许
    var customBool myBool = trueConst // 允许
    defaultBool = customBool // 不允许
}
```
[在线运行程序](https://play.golang.org/p/h9yzC6RxOR)

上面的程序是自我解释的。

## 数字常量

数字常量包含整数、浮点数和复数的常量。数字常量中有一些微妙之处。

让我们看一些例子来说清楚。

```go
package main

import (
    "fmt"
)

func main() {
    const a = 5
    var intVar int = a
    var int32Var int32 = a
    var float64Var float64 = a
    var complex64Var complex64 = a
    fmt.Println("intVar",intVar, "\nint32Var", int32Var, "\nfloat64Var", float64Var, "\ncomplex64Var",complex64Var)
}
```
[在线运行程序](https://play.golang.org/p/a8sxVNdU8M)

上面的程序，常量 `a` 是没有类型的，它的值是 `5` 。您可能想知道 `a` 的默认类型是什么，如果它确实有一个的话, 那么我们如何将它分配给不同类型的变量。答案在于 `a` 的语法。下面的程序将使事情更加清晰。

```go
package main

import (
    "fmt"
)

func main() {
    var i = 5
    var f = 5.6
    var c = 5 + 6i
    fmt.Printf("i's type %T, f's type %T, c's type %T", i, f, c)

}
```
[在线运行程序](https://play.golang.org/p/kJq69Vpqit)

在上面的程序中，每个变量的类型由数字常量的语法决定。`5` 在语法中是整数， `5.6` 是浮点数，`5+6i` 的语法是复数。当我们运行上面的程序，它会打印出 `i's type int, f's type float64, c's type complex128`。

现在我希望下面的程序能够正确的工作。

```go
package main

import (
    "fmt"
)

func main() {
    const a = 5
    var intVar int = a
    var int32Var int32 = a
    var float64Var float64 = a
    var complex64Var complex64 = a
    fmt.Println("intVar",intVar, "\nint32Var", int32Var, "\nfloat64Var", float64Var, "\ncomplex64Var",complex64Var)
}
```
[在线运行程序](https://play.golang.org/p/_zu0iK-Hyj)

在这个程序中， `a` 的值是 `5` ，`a` 的语法是通用的（它可以代表一个浮点数、整数甚至是一个没有虚部的复数），因此可以将其分配给任何兼容的类型。这些常量的默认类型可以被认为是根据上下文在运行中生成的。 `var intVar int = a` 要求 `a` 是 `int`，所以它变成一个 `int` 常量。 `var complex64Var complex64 = a` 要求 `a` 是 `complex64`，因此它变成一个复数类型。很简单的:)。

## 数字表达式

数字常量可以在表达式中自由混合和匹配，只有当它们被分配给变量或者在需要类型的代码中的任何地方使用时，才需要类型。

```go
package main

import (
    "fmt"
)

func main() {
    var a = 5.9/8
    fmt.Printf("a's type %T value %v",a, a)
}
```
[在线运行程序](https://play.golang.org/p/-8i-iX-jIG)

在上面的程序中， `5.9` 在语法中是浮点型，`8` 是整型，`5.9/8` 是允许的，因为两个都是数字常量。除法的结果是 `0.7375` 是一个浮点型，所以 `a` 的类型是浮点型。这个程序的输出结果是: `a's type float64 value 0.7375`。

**上一教程 - [类型](https://studygolang.com/articles/11869)**

**下一教程 - [函数](https://studygolang.com/articles/11892)**

---

via: https://golangbot.com/constants/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[guoxiaopang](https://github.com/guoxiaopang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
