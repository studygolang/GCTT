已发布：https://studygolang.com/articles/11957

# 教程 10：switch 语句

这是 [Golang 系列教程](https://golangbot.com/learn-golang-series/)中的第 10 篇。

switch 是一个条件语句，用于将表达式的值与可能匹配的选项列表进行比较，并根据匹配情况执行相应的代码块。它可以被认为是替代多个 `if else` 子句的常用方式。

看代码比文字更容易理解。让我们从一个简单的例子开始，它将把一个手指的编号作为输入，然后输出该手指对应的名字。比如 0 是拇指，1 是食指等等。

```go
package main

import (
    "fmt"
)

func main() {
    finger := 4
    switch finger {
    case 1:
        fmt.Println("Thumb")
    case 2:
        fmt.Println("Index")
    case 3:
        fmt.Println("Middle")
    case 4:
        fmt.Println("Ring")
    case 5:
        fmt.Println("Pinky")

    }
}
```
[在线运行程序](https://play.golang.org/p/q4kjm2kpVe)

在上述程序中，`switch finger` 将 `finger` 的值与每个 `case` 语句进行比较。通过从上到下对每一个值进行对比，并执行与选项值匹配的第一个逻辑。在上述样例中， `finger` 值为 4，因此打印的结果是 `Ring` 。

在选项列表中，`case` 不允许出现重复项。如果您尝试运行下面的程序，编译器会报这样的错误: `main.go：18：2：在tmp / sandbox887814166 / main.go：16：7`

```go
package main

import (
    "fmt"
)

func main() {
    finger := 4
    switch finger {
    case 1:
        fmt.Println("Thumb")
    case 2:
        fmt.Println("Index")
    case 3:
        fmt.Println("Middle")
    case 4:
        fmt.Println("Ring")
    case 4://重复项
        fmt.Println("Another Ring")
    case 5:
        fmt.Println("Pinky")

    }
}
```
[在线运行程序](https://play.golang.org/p/SfXdChWdoN)

## 默认情况（Default Case）

我们每个人一只手只有 5 个手指。如果我们输入了不正确的手指编号会发生什么？这个时候就应该是属于默认情况。当其他情况都不匹配时，将运行默认情况。

```go
package main

import (
    "fmt"
)

func main() {
    switch finger := 8; finger {
    case 1:
        fmt.Println("Thumb")
    case 2:
        fmt.Println("Index")
    case 3:
        fmt.Println("Middle")
    case 4:
        fmt.Println("Ring")
    case 5:
        fmt.Println("Pinky")
    default: // 默认情况
        fmt.Println("incorrect finger number")
    }
}
```
[在线运行程序](https://play.golang.org/p/Fq7U7SkHe1)  

在上述程序中 `finger` 的值是 8，它不符合其中任何情况，因此会打印 `incorrect finger number`。default 不一定只能出现在 switch 语句的最后，它可以放在 switch 语句的任何地方。

您可能也注意到我们稍微改变了 `finger` 变量的声明方式。`finger` 声明在了 switch 语句内。在表达式求值之前，switch 可以选择先执行一个语句。在这行 `switch finger：= 8; finger` 中， 先声明了`finger` 变量，随即在表达式中使用了它。在这里，`finger` 变量的作用域仅限于这个 switch 内。

## 多表达式判断

通过用逗号分隔，可以在一个 case 中包含多个表达式。

```go
package main

import (
    "fmt"
)

func main() {
    letter := "i"
    switch letter {
    case "a", "e", "i", "o", "u": // 一个选项多个表达式
        fmt.Println("vowel")
    default:
        fmt.Println("not a vowel")
    }
}
```

[在线运行程序](https://play.golang.org/p/Zs9Ek5SInh)  

在 `case "a","e","i","o","u":` 这一行中，列举了所有的元音。只要匹配该项，则将输出 `vowel`。

## 无表达式的 switch

在 switch 语句中，表达式是可选的，可以被省略。如果省略表达式，则表示这个 switch 语句等同于 `switch true`，并且每个 `case` 表达式都被认定为有效，相应的代码块也会被执行。

```go
package main

import (
    "fmt"
)

func main() {
    num := 75
    switch { // 表达式被省略了
    case num >= 0 && num <= 50:
        fmt.Println("num is greater than 0 and less than 50")
    case num >= 51 && num <= 100:
        fmt.Println("num is greater than 51 and less than 100")
    case num >= 101:
        fmt.Println("num is greater than 100")
    }

}
```
 
[在线运行程序](https://play.golang.org/p/mMJ8EryKbN)  

在上述代码中，switch 中缺少表达式，因此默认它为 true，true 值会和每一个 case 的求值结果进行匹配。`case num >= 51 && <= 100:` 为 true，所以程序输出 `num is greater than 51 and less than 100`。这种类型的 switch 语句可以替代多个 `if else` 子句。


## Fallthrough 语句

在 Go 中，每执行完一个 case 后，会从 switch 语句中跳出来，不再做后续 case 的判断和执行。使用 `fallthrough` 语句可以在已经执行完成的 case 之后，把控制权转移到下一个 case 的执行代码中。

让我们写一个程序来理解 fallthrough。我们的程序将检查输入的数字是否小于 50、100 或 200。例如我们输入 75，程序将输出`75 is lesser than 100` 和 `75 is lesser than 200`。我们用 fallthrough 来实现了这个功能。

```go
package main

import (
    "fmt"
)

func number() int {
    num := 15 * 5 
    return num
}

func main() {

    switch num := number(); { // num is not a constant
    case num < 50:
        fmt.Printf("%d is lesser than 50\n", num)
        fallthrough
    case num < 100:
        fmt.Printf("%d is lesser than 100\n", num)
        fallthrough
    case num < 200:
        fmt.Printf("%d is lesser than 200", num)
    }

}
```
[在线运行程序](https://play.golang.org/p/svGJAiswQj)

switch 和 case 的表达式不一定是常量。它们也可以在运行过程中通过计算得到。在上面的程序中，num 被初始化为函数 `number()` 的返回值。程序运行到 switch 中时，会计算出 case 的值。`case num < 100：` 的结果为 true，所以程序输出 `75 is lesser than 100`。当执行到下一句 `fallthrough` 时，程序控制直接跳转到下一个 case 的第一个执行逻辑中，所以打印出 `75 is lesser than 200`。最后这个程序的输出会是

```
75 is lesser than 100  
75 is lesser than 200 
```

**`fallthrough` 语句应该是 case 子句的最后一个语句。如果它出现在了 case 语句的中间，编译器将会报错：`fallthrough statement out of place`**

这也是我们本教程的最后内容。还有一种 switch 类型称为 **type switch** 。我们会在学习接口的时候再研究这个。


希望您能享受本次阅读。请留下您宝贵的意见和建议。

**上一教程 - [循环](https://studygolang.com/articles/11924)**

**下一教程 - [Arrays 和 Slices](https://studygolang.com/articles/12121)**

----------------

via: https://golangbot.com/switch/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[vicever](https://github.com/vicever)
校对：[Noluye](https://github.com/Noluye)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
