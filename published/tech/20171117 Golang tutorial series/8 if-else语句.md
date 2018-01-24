已发布：https://studygolang.com/articles/11902

# 第 8 章：if-else 语句

这是我们 [Golang 系列教程](https://golangbot.com/learn-golang-series/)的第 8 篇。

if 是条件语句。if 语句的语法是

```go
if condition {  
}
```

如果 `condition` 为真，则执行 `{` 和 `}` 之间的代码。

不同于其他语言，例如 C 语言，Go 语言里的 `{  }` 是必要的，即使在 `{  }` 之间只有一条语句。

if 语句还有可选的 `else if` 和 `else` 部分。

```go
if condition {  
} else if condition {
} else {
}
```

if-else 语句之间可以有任意数量的 `else if`。条件判断顺序是从上到下。如果 `if` 或 `else if` 条件判断的结果为真，则执行相应的代码块。 如果没有条件为真，则 `else` 代码块被执行。

让我们编写一个简单的程序来检测一个数字是奇数还是偶数。

```go
package main

import (  
    "fmt"
)

func main() {  
    num := 10
    if num % 2 == 0 { //checks if number is even
        fmt.Println("the number is even") 
    }  else {
        fmt.Println("the number is odd")
    }
}
```
[在线运行程序](https://play.golang.org/p/vWfN8UqZUr)

`if num％2 == 0` 语句检测 num 取 2 的余数是否为零。 如果是为零则打印输出 "the number is even"，如果不为零则打印输出 "the number is odd"。在上面的这个程序中，打印输出的是 `the number is even`。

`if` 还有另外一种形式，它包含一个 `statement` 可选语句部分，该组件在条件判断之前运行。它的语法是

```go
if statement; condition {  
}
```

让我们重写程序，使用上面的语法来查找数字是偶数还是奇数。

```go
package main

import (  
    "fmt"
)

func main() {  
    if num := 10; num % 2 == 0 { //checks if number is even
        fmt.Println(num,"is even") 
    }  else {
        fmt.Println(num,"is odd")
    }
}
```
[在线运行程序](https://play.golang.org/p/_X9q4MWr4s)

在上面的程序中，`num` 在 `if` 语句中进行初始化，`num` 只能从 `if` 和 `else` 中访问。也就是说 `num` 的范围仅限于 `if` `else` 代码块。如果我们试图从其他外部的 `if` 或者 `else` 访问 `num`,编译器会不通过。

让我们再写一个使用 `else if` 的程序。

```go
package main

import (  
    "fmt"
)

func main() {  
    num := 99
    if num <= 50 {
        fmt.Println("number is less than or equal to 50")
    } else if num >= 51 && num <= 100 {
        fmt.Println("number is between 51 and 100")
    } else {
        fmt.Println("number is greater than 100")
    }

}
```
[在线运行程序](https://play.golang.org/p/Eji7vmb17Q)

在上面的程序中，如果 `else if num >= 51 && num <= 100` 为真，程序将输出 `number is between 51 and 100`。

[获取免费的 Golang 工具](https://app.mailerlite.com/webforms/popup/p8t5t8)

### 一个注意点

`else` 语句应该在 `if` 语句的大括号 `}` 之后的同一行中。如果不是，编译器会不通过。

让我们通过以下程序来理解它。

```go
package main

import (  
    "fmt"
)

func main() {  
    num := 10
    if num % 2 == 0 { //checks if number is even
        fmt.Println("the number is even") 
    }  
    else {
        fmt.Println("the number is odd")
    }
}
```
[在线运行程序](https://play.golang.org/p/RYNqZZO2F9)

在上面的程序中，`else` 语句不是从 `if` 语句结束后的 `}` 同一行开始。而是从下一行开始。这是不允许的。如果运行这个程序，编译器会输出错误，

```
main.go:12:5: syntax error: unexpected else, expecting }
```

出错的原因是 Go 语言的分号是自动插入。你可以在这里阅读分号插入规则 [https://golang.org/ref/spec#Semicolons](https://golang.org/ref/spec#Semicolons)。

在 Go 语言规则中，它指定在 `}` 之后插入一个分号，如果这是该行的最终标记。因此，在if语句后面的 `}` 会自动插入一个分号。

实际上我们的程序变成了

```go
if num%2 == 0 {  
      fmt.Println("the number is even") 
};  //semicolon inserted by Go
else {  
      fmt.Println("the number is odd")
}
```

分号插入之后。从上面代码片段可以看出第三行插入了分号。

由于 `if{…} else {…}` 是一个单独的语句，它的中间不应该出现分号。因此，需要将 `else` 语句放置在 `}` 之后处于同一行中。

我已经重写了程序，将 else 语句移动到 if 语句结束后 `}` 的后面，以防止分号的自动插入。

```go
package main

import (  
    "fmt"
)

func main() {  
    num := 10
    if num%2 == 0 { //checks if number is even
        fmt.Println("the number is even") 
    } else {
        fmt.Println("the number is odd")
    }
}
```
[在线运行程序](https://play.golang.org/p/hv_27vbIBC)

现在编译器会很开心，我们也一样 😃。

本章教程到此告一段落了，感谢您的阅读，欢迎您的任何评论和反馈。

**下一个教程 - 循环**

----------------

via: https://golangbot.com/if-statement/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Dingo1991](https://github.com/Dingo1991)
校对：[rxcai](https://github.com/rxcai)
本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
