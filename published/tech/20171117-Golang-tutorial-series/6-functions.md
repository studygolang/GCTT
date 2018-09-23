已发布：https://studygolang.com/articles/11892

# 第 6 部分: 函数（Function）

这是我们 [Golang 系列教程](https://golangbot.com/learn-golang-series/)第 6 章，学习 Golang 函数的相关知识。

## 函数是什么？

函数是一块执行特定任务的代码。一个函数是在输入源基础上，通过执行一系列的算法，生成预期的输出。

## 函数的声明

在 Go 语言中，函数声明通用语法如下：

```go
func functionname(parametername type) returntype {
    // 函数体（具体实现的功能）
}
```

函数的声明以关键词 `func` 开始，后面紧跟自定义的函数名 `functionname (函数名)`。函数的参数列表定义在 `(` 和 `)` 之间，返回值的类型则定义在之后的 `returntype (返回值类型)`处。声明一个参数的语法采用 **参数名** **参数类型** 的方式，任意多个参数采用类似 `(parameter1 type, parameter2 type) 即(参数1 参数1的类型,参数2 参数2的类型)`的形式指定。之后包含在 `{` 和 `}` 之间的代码，就是函数体。

函数中的参数列表和返回值并非是必须的，所以下面这个函数的声明也是有效的

```go
func functionname() {
	// 译注: 表示这个函数不需要输入参数，且没有返回值
}
```

## 示例函数

我们以写一个计算商品价格的函数为例，输入参数是单件商品的价格和商品的个数，两者的乘积为商品总价，作为函数的输出值。

```go
func calculateBill(price int, no int) int {
    var totalPrice = price * no // 商品总价 = 商品单价 * 数量
    return totalPrice // 返回总价
}
```
上述函数有两个整型的输入 `price` 和 `no`，返回值 `totalPrice` 为 `price` 和 `no` 的乘积，也是整数类型。

**如果有连续若干个参数，它们的类型一致，那么我们无须一一罗列，只需在最后一个参数后添加该类型。** 例如，`price int, no int` 可以简写为 `price, no int`，所以示例函数也可写成

```go
func calculateBill(price, no int) int {
    var totalPrice = price * no
    return totalPrice
}
```

现在我们已经定义了一个函数，我们要在代码中尝试着调用它。调用函数的语法为 `functionname(parameters)`。调用示例函数的方法如下：

```go
calculateBill(10, 5)
```

完成了示例函数声明和调用后，我们就能写出一个完整的程序，并把商品总价打印在控制台上：

```go
package main

import (
    "fmt"
)

func calculateBill(price, no int) int {
    var totalPrice = price * no
    return totalPrice
}
func main() {
    price, no := 90, 6 // 定义 price 和 no,默认类型为 int
    totalPrice := calculateBill(price, no)
    fmt.Println("Total price is", totalPrice) // 打印到控制台上
}
```

[运行这个程序](https://play.golang.org/p/YJlW3g-VZH)

该程序在控制台上打印的结果为

```
Total price is 540
```

## 多返回值

Go 语言支持一个函数可以有多个返回值。我们来写个以矩形的长和宽为输入参数，计算并返回矩形面积和周长的函数 `rectProps`。矩形的面积是长度和宽度的乘积, 周长是长度和宽度之和的两倍。即：

- `面积 = 长 * 宽`
- `周长 = 2 * ( 长 + 宽 )`

```go
package main

import (
    "fmt"
)

func rectProps(length, width float64)(float64, float64) {
    var area = length * width
    var perimeter = (length + width) * 2
    return area, perimeter
}

func main() {
    area, perimeter := rectProps(10.8, 5.6)
    fmt.Printf("Area %f Perimeter %f", area, perimeter)
}
```

[运行这个程序](https://play.golang.org/p/qAftE_yke_)

如果一个函数有多个返回值，那么这些返回值必须用 `(` 和 `)` 括起来。`func rectProps(length, width float64)(float64, float64)` 示例函数有两个 float64 类型的输入参数 `length` 和 `width`，并返回两个 float64 类型的值。该程序在控制台上打印结果为

```
Area 60.480000 Perimeter 32.800000
```

## 命名返回值

从函数中可以返回一个命名值。一旦命名了返回值，可以认为这些值在函数第一行就被声明为变量了。

上面的 rectProps 函数也可用这个方式写成：

```go
func rectProps(length, width float64)(area, perimeter float64) {
    area = length * width
    perimeter = (length + width) * 2
    return // 不需要明确指定返回值，默认返回 area, perimeter 的值
}
```

请注意, 函数中的 return 语句没有显式返回任何值。由于 **area** 和 **perimeter** 在函数声明中指定为返回值, 因此当遇到 return 语句时, 它们将自动从函数返回。

## 空白符

**_** 在 Go 中被用作空白符，可以用作表示任何类型的任何值。

我们继续以 `rectProps` 函数为例，该函数计算的是面积和周长。假使我们只需要计算面积，而并不关心周长的计算结果，该怎么调用这个函数呢？这时，空白符 **_** 就上场了。

下面的程序我们只用到了函数 `rectProps` 的一个返回值 `area`

```go
package main

import (
    "fmt"
)

func rectProps(length, width float64) (float64, float64) {
    var area = length * width
    var perimeter = (length + width) * 2
    return area, perimeter
}
func main() {
    area, _ := rectProps(10.8, 5.6) // 返回值周长被丢弃
    fmt.Printf("Area %f ", area)
}
```

[运行这个程序](https://play.golang.org/p/IkugSH1jIt)

> 在程序的 `area, _ := rectProps(10.8, 5.6)` 这一行，我们看到空白符 `_` 用来跳过不要的计算结果。

本章教程到此告一段落了，感谢您的阅读，欢迎您的任何评论和反馈。

**上一教程 - [常量](https://studygolang.com/articles/11872)**

**下一教程 - [包 (Packages)](https://studygolang.com/articles/11893)**

-------
via: https://golangbot.com/functions/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Junedayday](https://github.com/Junedayday)
校对：[Unknwon](https://github.com/Unknwon)、[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
