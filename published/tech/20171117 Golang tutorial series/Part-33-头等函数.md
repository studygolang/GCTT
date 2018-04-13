已发布：https://studygolang.com/articles/12789

# 第 33 篇：函数是一等公民（头等函数）

![custom errors](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-series/first-class-functions-golang.png)

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 33 篇。

## 什么是头等函数？

**支持头等函数（First Class Function）的编程语言，可以把函数赋值给变量，也可以把函数作为其它函数的参数或者返回值。Go 语言支持头等函数的机制**。

本教程我们会讨论头等函数的语法和用例。

## 匿名函数

我们来编写一个简单的示例，把[函数](https://studygolang.com/articles/11892)赋值给一个[变量](https://studygolang.com/articles/11756)。

```go
package main

import (  
    "fmt"
)

func main() {  
    a := func() {
        fmt.Println("hello world first class function")
    }
    a()
    fmt.Printf("%T", a)
}
```

[在 playground 上运行](https://play.golang.org/p/Xm_ihamhlEv)

在上面的程序中，我们将一个函数赋值给了变量 `a`（第 8 行）。这是把函数赋值给变量的语法。你如果观察得仔细的话，会发现赋值给 `a` 的函数没有名称。**由于没有名称，这类函数称为匿名函数（Anonymous Function）**。

调用该函数的唯一方法就是使用变量 `a`。我们在下一行调用了它。`a()` 调用了这个函数，打印出 `hello world first class function`。在第 12 行，我们打印出 `a` 的类型。这会输出 `func()`。

运行该程序，会输出：

```
hello world first class function
func()
```

要调用一个匿名函数，可以不用赋值给变量。通过下面的例子，我们看看这是怎么做到的。

```go
package main

import (  
    "fmt"
)

func main() {  
    func() {
        fmt.Println("hello world first class function")
    }()
}
```

[在 playground 上运行](https://play.golang.org/p/c0AjB3g8UEn)

在上面的程序中，第 8 行定义了一个匿名函数，并在定义之后，我们使用 `()` 立即调用了该函数（第 10 行）。该程序会输出：

```
hello world first class function
```

就像其它函数一样，还可以向匿名函数传递参数。

```go
package main

import (  
    "fmt"
)

func main() {  
    func(n string) {
        fmt.Println("Welcome", n)
    }("Gophers")
}
```

[在 playground 上运行](https://play.golang.org/p/9ttJ5Wi4fj4)

在上面的程序中，我们向匿名函数传递了一个字符串参数（第 10 行）。运行该程序后会输出：

```
Welcome Gophers
```

## 用户自定义的函数类型

正如我们定义自己的[结构体](https://studygolang.com/articles/12263)类型一样，我们可以定义自己的函数类型。

```go
type add func(a int, b int) int
```

以上代码片段创建了一个新的函数类型 `add`，它接收两个整型参数，并返回一个整型。现在我们来定义 `add` 类型的变量。

我们来编写一个程序，定义一个 `add` 类型的变量。

```go
package main

import (  
    "fmt"
)

type add func(a int, b int) int

func main() {  
    var a add = func(a int, b int) int {
        return a + b
    }
    s := a(5, 6)
    fmt.Println("Sum", s)
}
```

[在 playground 上运行](https://play.golang.org/p/n3yPQ7hG7ip)

在上面程序的第 10 行，我们定义了一个 `add` 类型的变量 `a`，并向它赋值了一个符合 `add` 类型签名的函数。我们在第 13 行调用了该函数，并将结果赋值给 `s`。该程序会输出：

```
Sum 11
```

## 高阶函数

[wiki](https://en.wikipedia.org/wiki/Higher-order_function) 把高阶函数（Hiher-order Function）定义为：**满足下列条件之一的函数**：

- **接收一个或多个函数作为参数**
- **返回值是一个函数**

针对上述两种情况，我们看看一些简单实例。

### 把函数作为参数，传递给其它函数

```go
package main

import (  
    "fmt"
)

func simple(a func(a, b int) int) {  
    fmt.Println(a(60, 7))
}

func main() {  
    f := func(a, b int) int {
        return a + b
    }
    simple(f)
}
```

[在 playground 上运行](https://play.golang.org/p/C0MNwz2TSGU)

在上面的实例中，第 7 行我们定义了一个函数 `simple`，`simple` 接收一个函数参数（该函数接收两个 `int` 参数，返回一个 `a` 整型）。在 `main` 函数的第 12 行，我们创建了一个匿名函数 `f`，其签名符合 `simple` 函数的参数。我们在下一行调用了 `simple`，并传递了参数 `f`。该程序打印输出 67。

### 在其它函数中返回函数

现在我们重写上面的代码，在 `simple` 函数中返回一个函数。

```go
package main

import (  
    "fmt"
)

func simple() func(a, b int) int {  
    f := func(a, b int) int {
        return a + b
    }
    return f
}

func main() {  
    s := simple()
    fmt.Println(s(60, 7))
}
```

[在 playground 上运行](https://play.golang.org/p/82y2caejUy8)

在上面程序中，第 7 行的 `simple` 函数返回了一个函数，并接受两个 `int` 参数，返回一个 `int`。

在第 15 行，我们调用了 `simple` 函数。我们把 `simple` 的返回值赋值给了 `s`。现在 `s` 包含了 `simple` 函数返回的函数。我们调用了 `s`，并向它传递了两个 int 参数（第 16 行）。该程序输出 67。

## 闭包

闭包（Closure）是匿名函数的一个特例。当一个匿名函数所访问的变量定义在函数体的外部时，就称这样的匿名函数为闭包。

看看一个示例就明白了。

```go
package main

import (  
    "fmt"
)

func main() {  
    a := 5
    func() {
        fmt.Println("a =", a)
    }()
}
```

[在 playground 上运行](https://play.golang.org/p/6QriMs-zbnf)

在上面的程序中，匿名函数在第 10 行访问了变量 `a`，而 `a` 存在于函数体的外部。因此这个匿名函数就是闭包。

每一个闭包都会绑定一个它自己的外围变量（Surrounding Variable）。我们通过一个简单示例来体会这句话的含义。

```go
package main

import (  
    "fmt"
)

func appendStr() func(string) string {  
    t := "Hello"
    c := func(b string) string {
        t = t + " " + b
        return t
    }
    return c
}

func main() {  
    a := appendStr()
    b := appendStr()
    fmt.Println(a("World"))
    fmt.Println(b("Everyone"))

    fmt.Println(a("Gopher"))
    fmt.Println(b("!"))
}
```

[在 playground 上运行](https://play.golang.org/p/134NiQGPOcS)

在上面程序中，函数 `appendStr` 返回了一个闭包。这个闭包绑定了变量 `t`。我们来理解这是什么意思。

在第 17 行和第 18 行声明的变量 `a` 和 `b` 都是闭包，它们绑定了各自的 `t` 值。

我们首先用参数 `World` 调用了 `a`。现在 `a` 中 `t` 值变为了 `Hello World`。

在第 20 行，我们又用参数 `Everyone` 调用了 `b`。由于 `b` 绑定了自己的变量 `t`，因此 `b` 中的 `t` 还是等于初始值 `Hello`。于是该函数调用之后，`b` 中的 `t` 变为了 `Hello Everyone`。程序的其他部分很简单，不再解释。

该程序会输出：

```
Hello World  
Hello Everyone  
Hello World Gopher  
Hello Everyone ! 
```

## 头等函数的实际用途

迄今为止，我们已经定义了什么是头等函数，也看了一些专门设计的示例，来学习它们如何工作。现在我们来编写一些实际的程序，来展现头等函数的实际用处。

我们会创建一个程序，基于一些条件，来过滤一个 `students` 切片。现在我们来逐步实现它。

首先定义一个 `student` 类型。

```go
type student struct {  
    firstName string
    lastName string
    grade string
    country string
}
```

下一步是编写一个 `filter` 函数。该函数接收一个 `students` 切片和一个函数作为参数，这个函数会计算一个学生是否满足筛选条件。写出这个函数后，你很快就会明白，我们继续吧。

```go
func filter(s []student, f func(student) bool) []student {  
    var r []student
    for _, v := range s {
        if f(v) == true {
            r = append(r, v)
        }
    }
    return r
}
```

在上面的函数中，`filter` 的第二个参数是一个函数。这个函数接收 `student` 参数，返回一个 `bool` 值。这个函数计算了某一学生是否满足筛选条件。我们在第 3 行遍历了 `student` 切片，将每个学生作为参数传递给了函数 `f`。如果该函数返回 `true`，就表示该学生通过了筛选条件，接着将该学生添加到了结果切片 `r` 中。你可能会很困惑这个函数的实际用途，等我们完成程序你就知道了。我添加了 `main` 函数，整个程序如下所示：

```go
package main

import (  
    "fmt"
)

type student struct {  
    firstName string
    lastName  string
    grade     string
    country   string
}

func filter(s []student, f func(student) bool) []student {  
    var r []student
    for _, v := range s {
        if f(v) == true {
            r = append(r, v)
        }
    }
    return r
}

func main() {  
    s1 := student{
        firstName: "Naveen",
        lastName:  "Ramanathan",
        grade:     "A",
        country:   "India",
    }
    s2 := student{
        firstName: "Samuel",
        lastName:  "Johnson",
        grade:     "B",
        country:   "USA",
    }
    s := []student{s1, s2}
    f := filter(s, func(s student) bool {
        if s.grade == "B" {
            return true
        }
        return false
    })
    fmt.Println(f)
}
```

[在 playground 上运行](https://play.golang.org/p/YUL1CqSrvfc)

在 `main` 函数中，我们首先创建了两个学生 `s1` 和 `s2`，并将他们添加到了切片 `s`。现在假设我们想要查询所有成绩为 `B` 的学生。为了实现这样的功能，我们传递了一个检查学生成绩是否为 `B` 的函数，如果是，该函数会返回 `true`。我们把这个函数作为参数传递给了 `filter` 函数（第 38 行）。上述程序会输出：

```
[{Samuel Johnson B USA}]
```

假设我们想要查找所有来自印度的学生。通过修改传递给 `filter` 的函数参数，就很容易地实现了。

实现它的代码如下所示：

```go
c := filter(s, func(s student) bool {  
    if s.country == "India" {
        return true
    }
    return false
})
fmt.Println(c)  
```

请将该函数添加到 `main` 函数，并检查它的输出。

我们最后再编写一个程序，来结束这一节的讨论。这个程序会对切片的每个元素执行相同的操作，并返回结果。例如，如果我们希望将切片中的所有整数乘以 5，并返回出结果，那么通过头等函数可以很轻松地实现。我们把这种对集合中的每个元素进行操作的函数称为 `map` 函数。相关代码如下所示，它们很容易看懂。

```go
package main

import (  
    "fmt"
)

func iMap(s []int, f func(int) int) []int {  
    var r []int
    for _, v := range s {
        r = append(r, f(v))
    }
    return r
}
func main() {  
    a := []int{5, 6, 7, 8, 9}
    r := iMap(a, func(n int) int {
        return n * 5
    })
    fmt.Println(r)
}
```

[在 playground 上运行](https://play.golang.org/p/cs37QwCQ_0H)

该程序会输出：

```
[25 30 35 40 45]
```

现在简单概括一下本教程讨论的内容：

- 什么是头等函数？
- 匿名函数
- 用户自定义的函数类型
- 高阶函数
    - 把函数作为参数，传递给其它函数
    - 在其它函数中返回函数
- 闭包
- 头等函数的实际用途

本教程到此结束。祝你愉快。

**上一教程 - [panic 和 recover](https://studygolang.com/articles/12785)**

---

via: https://golangbot.com/first-class-functions/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
