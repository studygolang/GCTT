
# For Range 的语义

## 前言

为了更好地理解本文中提及的内容，这些是需要首先阅读的好文章：

下面列出4篇文章的索引：
- 1. [Language Mechanics On Stacks And Pointers](https://www.goinggo.net/2017/05/language-mechanics-on-stacks-and-pointers.html)
- 2. [Language Mechanics On Escape Analysis](https://www.goinggo.net/2017/05/language-mechanics-on-escape-analysis.html)
- 3. [Language Mechanics On Memory Profiling](https://www.goinggo.net/2017/06/language-mechanics-on-memory-profiling.html)
- 4. [Design Philosophy On Data And Semantics](https://www.goinggo.net/2017/06/design-philosophy-on-data-and-semantics.html)

在 Go 编程语言中，值语义和指针语义的思想无处不在。如前面的文章所述，语义一致性对于完整性和可读性至关重要。它允许开发人员在代码持续不断增长时保持强大的代码库[心理模型](https://en.wikipedia.org/wiki/Mental_model)。它还有助于最大限度地减少错误，副作用和未知行为。

## 介绍

在这篇文章中，我将探索 Go 中的 `for range` 语句如何提供值和执行语义形式。我将教授语言机制，并展示这些语义有多深奥。然后件展示一个简单的例子，说明将这些语义和可能导致的问题混合起来是多么容易。

## 语言机制

从这段代码开始，它展示了 `for range` 循环的值语义形式。

[play.golang.org](https://play.golang.org/p/_CWCAF6ge3)

**代码清单1**

```go
package main

import "fmt"

type user struct {
    name string
    email string
}

func main() {
    users := []user{
        {"Bill", "bill@email.com"},
        {"Lisa", "lisa@email.com"},
        {"Nancy", "nancy@email.com"},
        {"Paul", "paul@email.com"},
    }

    for i, u := range users {
        fmt.Println(i, u)
    }
}
```
在代码清单1中，程序声明一个名为 `user` 的类型，创建四个用户值，然后显示关于每个用户的信息。第 18 行的范围循环使用值语义。这是因为在每次迭代中都会在循环内部创建并操作来自切片的原始用户值的副本。实际上，对 `Println` 的调用会创建循环副本的第二个副本。如果要为用户值使用值语义，这就是你想要的。

如果你要使用指针语义，`for range` 循环看起来像这样。

**代码清单2**
```
for i := range users {
    fmt.Println(i, users[i])
}
```

现在该循环已被修改为使用指针语义。循环内的代码不再它的副本上运行，而是在切片内存储的原始 `user` 上运行。但是，对 `Println` 的调用仍然使用值语义，并且传递了一份副本。

要解决这个问题，需要再做一次最后的修改。

**代码清单3**
```
for i := range users {
    fmt.Println(i, &users[i])
}
```

现在会一直使用 `user` 的指针语义。

作为参考，清单4并排显示了值和指针语义。

**代码清单4**
```
// Value semantics.           // Pointer semantics.
for i, u := range users {     for i := range users {
    fmt.Println(i, u)             fmt.Println(i, &users[i])
}                             }
```

## 深层机制

语言机制比这更深入。请看代码清单 5 中的这个程序。程序初始化一个字符串数组，对这些字符串进行迭代，并在每次迭代中更改索引为 1 的字符串。

[https://play.golang.org/p/IlAiEkgs4C](https://play.golang.org/p/IlAiEkgs4C)

**代码清单5**
```
package main

import "fmt"

func main() {
    five := [5]string{"Annie", "Betty", "Charley", "Doug", "Edward"}
    fmt.Printf("Bfr[%s] : ", five[1])

    for i := range five {
        five[1] = "Jack"

        if i == 1 {
           fmt.Printf("Aft[%s]\n", five[1])
        }
    }
}
```

这个程序的预期输出是什么？

**清单6**
```
Bfr[Betty]
Aft[Jack]
```
正如你所期望的那样，第 10 行的代码已经改变了索引 1 的字符串，你可以在输出中看到。该程序使用 `for range` 循环的指针语义版本。接下来，代码将使用 `for range` 循环的值语义版本。

[ttps://play.golang.org/p/opSsIGtNU1](ttps://play.golang.org/p/opSsIGtNU1)

**清单7**

```
package main

import "fmt"

func main() {
    five := [5]string{"Annie", "Betty", "Charley", "Doug", "Edward"}
    fmt.Printf("Bfr[%s] : ", five[1])

    for i, v := range five {
        five[1] = "Jack"

        if i == 1 {
            fmt.Printf("v[%s]\n", v)
        }
    }
}
```

在循环的每次迭代中，代码再次更改索引 1 处的字符串。此时代码显示索引 1 处的值时，输出不同。

**清单8**
```
Bfr[Betty] : v[Betty]
```




---

via: https://www.ardanlabs.com/blog/2017/06/for-range-semantics.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[shniu](https://github.com/shniu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
