已发布：https://studygolang.com/articles/11756

# 第 3 部分：变量

这是我们 Golang 系列教程的第 3 个教程，探讨 Golang 里的变量（Variables）。

你可以阅读 Golang 系列**教程第 2 部分：Hello World**，学习如何配置 Golang，并运行 hello world 程序。

## 变量是什么

变量指定了某存储单元（Memory Location）的名称，该存储单元会存储特定类型的值。在 Go 中，有多种语法用于声明变量。

## 声明单个变量

**var name type** 是声明单个变量的语法。

```go
package main

import "fmt"

func main() {
    var age int // 变量声明
    fmt.Println("my age is", age)
}
```
[在线运行程序](https://play.golang.org/p/XrveIxw_YI)

语句 `var age int` 声明了一个 int 类型的变量，名字为 age。我们还没有给该变量赋值。如果变量未被赋值，Go 会自动地将其初始化，赋值该变量类型的零值（Zero Value）。本例中 age 就被赋值为 0。如果你运行该程序，你会看到如下输出：
```
my age is 0
```
变量可以赋值为本类型的任何值。上一程序中的 age 可以赋值为任何整型值（Integer Value）。

```go
package main

import "fmt"

func main() {
    var age int // 变量声明
    fmt.Println("my age is", age)
    age = 29 // 赋值
    fmt.Println("my age is", age)
    age = 54 // 赋值
    fmt.Println("my new age is", age)
}
```
[在线运行程序](https://play.golang.org/p/z4nKMjBxLx)

上面的程序会有如下输出：
```
my age is  0
my age is 29
my new age is 54
```

## 声明变量并初始化

声明变量的同时可以给定初始值。

**var name type = initialvalue** 的语法用于声明变量并初始化。

```go
package main

import "fmt"

func main() {
    var age int = 29 // 声明变量并初始化

    fmt.Println("my age is", age)
}
```
[在线运行程序](https://play.golang.org/p/TFfpzsrchh)

在上面的程序中，age 是具有初始值 29 的 int 类型变量。如果你运行上面的程序，你可以看见下面的输出，证实 age 已经被初始化为 29。
```
my age is 29
```

## 类型推断（Type Inference）

如果变量有初始值，那么 Go 能够自动推断具有初始值的变量的类型。因此，如果变量有初始值，就可以在变量声明中省略 `type`。

如果变量声明的语法是 **var name = initialvalue**，Go 能够根据初始值自动推断变量的类型。

在下面的例子中，你可以看到在第 6 行，我们省略了变量 `age` 的 `int` 类型，Go 依然推断出了它是 int 类型。

```go
package main

import "fmt"

func main() {
    var age = 29 // 可以推断类型

    fmt.Println("my age is", age)
}
```
[在线运行程序](https://play.golang.org/p/FgNbfL3WIt)

## 声明多个变量

Go 能够通过一条语句声明多个变量。

声明多个变量的语法是 **var name1, name2 type = initialvalue1, initialvalue2**。

```go
package main

import "fmt"

func main() {
    var width, height int = 100, 50 // 声明多个变量

    fmt.Println("width is", width, "height is", heigh)
}
```
[在线运行程序](https://play.golang.org/p/4aOQyt55ah)

上述程序将在标准输出打印 `width is 100 height is 50`。

你可能已经想到，如果 width 和 height 省略了初始化，它们的初始值将赋值为 0。

```go
package main

import "fmt"

func main() {
    var width, height int
    fmt.Println("width is", width, "height is", height)
    width = 100
    height = 50
    fmt.Println("new width is", width, "new height is ", height)
}
```
[在线运行程序](https://play.golang.org/p/DM00pcBbsu)

上面的程序将会打印：

```
width is 0 height is 0
new width is 100 new height is  50
```

在有些情况下，我们可能会想要在一个语句中声明不同类型的变量。其语法如下：

```go
var (
    name1 = initialvalue1,
    name2 = initialvalue2
)
```
使用上述语法，下面的程序声明不同类型的变量。

```go
package main

import "fmt"

func main() {
    var (
        name   = "naveen"
        age    = 29
        height int
    )
    fmt.Println("my name is", name, ", age is", age, "and height is", height)
}
```
[在线运行程序](https://play.golang.org/p/7pkp74h_9L)

这里我们声明了 **string 类型的 name、int 类型的 age 和 height**（我们将会在下一教程中讨论 Golang 所支持的变量类型）。运行上面的程序会产生输出 `my name is naveen , age is 29 and height is 0`。

## 简短声明

Go 也支持一种声明变量的简洁形式，称为简短声明（Short Hand Declaration），该声明使用了 **:=** 操作符。

声明变量的简短语法是 **name := initialvalue**。
```go
package main

import "fmt"

func main() {
    name, age := "naveen", 29 // 简短声明

    fmt.Println("my name is", name, "age is", age)
}
```
[在线运行程序](https://play.golang.org/p/ctqgw4w6kx)

运行上面的程序，可以看到输出为 `my name is naveen age is 29`。

简短声明要求 **:=** 操作符左边的所有变量都有初始值。下面程序将会抛出错误 `cannot assign 1 values to 2 variables`，这是因为 **age 没有被赋值**。

```go
package main

import "fmt"

func main() {
    name, age := "naveen" //error

    fmt.Println("my name is", name, "age is", age)
}
```
[在线运行程序](https://play.golang.org/p/wZd2HmDvqw)

简短声明的语法要求 **:=** 操作符的左边至少有一个变量是尚未声明的。考虑下面的程序：

```go
package main

import "fmt"

func main() {
    a, b := 20, 30 // 声明变量 a 和 b
    fmt.Println("a is", a, "b is", b)
    b, c := 40, 50 // b 已经声明，但 c 尚未声明
    fmt.Println("b is", b, "c is", c)
    b, c = 80, 90 // 给已经声明的变量 b 和 c 赋新值
    fmt.Println("changed b is", b, "c is", c)
}
```
[在线运行程序](https://play.golang.org/p/MSUYR8vazB)

在上面程序中的第 8 行，由于 b 已经被声明，而 c 尚未声明，因此运行成功并且输出：
```
a is 20 b is 30
b is 40 c is 50
changed b is 80 c is 90
```
但是如果我们运行下面的程序:

```go
package main

import "fmt"

func main() {
    a, b := 20, 30 // 声明 a 和 b
    fmt.Println("a is", a, "b is", b)
    a, b := 40, 50 // 错误，没有尚未声明的变量
}
```
[在线运行程序](https://play.golang.org/p/EYTtRnlDu3)

上面运行后会抛出 `no new variables on left side of :=` 的错误，这是因为 a 和 b 的变量已经声明过了，**:=** 的左边并没有尚未声明的变量。

变量也可以在运行时进行赋值。考虑下面的程序：

```go
package main

import (
    "fmt"
    "math"
)

func main() {
    a, b := 145.8, 543.8
    c := math.Min(a, b)
    fmt.Println("minimum value is ", c)
}
```
[在线运行程序](https://play.golang.org/p/7XojAtrpH9)

在上面的程序中，c 的值是运行过程中计算得到的，即 a 和 b 的最小值。上述程序会打印：
```
minimum value is  145.8
```

由于 Go 是强类型（Strongly Typed）语言，因此不允许某一类型的变量赋值为其他类型的值。下面的程序会抛出错误 `cannot use "naveen" (type string) as type int in assignment`，这是因为 age 本来声明为 int 类型，而我们却尝试给它赋字符串类型的值。

```go
package main

func main() {
    age := 29      // age 是 int 类型
    age = "naveen" // 错误，尝试赋值一个字符串给 int 类型变量
}
```
[在线运行程序](https://play.golang.org/p/K5rz4gxjPj)

感谢您的阅读，请在评论栏上面发布您的问题和反馈。

**上一教程 - [Hello World](https://studygolang.com/articles/11755)**

**下一教程 - [类型](https://studygolang.com/articles/11869)**

---

via: https://golangbot.com/variables/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[Unknwon](https://github.com/Unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
