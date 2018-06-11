已发布：https://studygolang.com/articles/12680

# 第 27 篇：组合取代继承

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 27 篇。

Go 不支持继承，但它支持组合（Composition）。组合一般定义为“合并在一起”。汽车就是一个关于组合的例子：一辆汽车由车轮、引擎和其他各种部件组合在一起。

## 通过嵌套结构体进行组合

在 Go 中，通过在结构体内嵌套结构体，可以实现组合。

组合的典型例子就是博客帖子。每一个博客的帖子都有标题、内容和作者信息。使用组合可以很好地表示它们。通过学习本教程后面的内容，我们会知道如何实现组合。

我们首先创建一个 `author` 结构体。

```go
package main

import (
    "fmt"
)

type author struct {
    firstName string
    lastName  string
    bio       string
}

func (a author) fullName() string {
    return fmt.Sprintf("%s %s", a.firstName, a.lastName)
}
```

在上面的代码片段中，我们创建了一个 `author` 结构体，`author` 的字段有 `firstname`、`lastname` 和 `bio`。我们还添加了一个 `fullName()` 方法，其中 `author` 作为接收者类型，该方法返回了作者的全名。

下一步我们创建 `post` 结构体。

```go
type post struct {
    title     string
    content   string
    author
}

func (p post) details() {
    fmt.Println("Title: ", p.title)
    fmt.Println("Content: ", p.content)
    fmt.Println("Author: ", p.author.fullName())
    fmt.Println("Bio: ", p.author.bio)
}
```

`post` 结构体的字段有 `title` 和 `content`。它还有一个嵌套的匿名字段 `author`。该字段指定 `author` 组成了 `post` 结构体。现在 `post` 可以访问 `author` 结构体的所有字段和方法。我们同样给 `post` 结构体添加了 `details()` 方法，用于打印标题、内容和作者的全名与简介。

一旦结构体内嵌套了一个结构体字段，Go 可以使我们访问其嵌套的字段，好像这些字段属于外部结构体一样。所以上面第 11 行的 `p.author.fullName()` 可以替换为 `p.fullName()`。于是，`details()` 方法可以重写，如下所示：

```go
func (p post) details() {
    fmt.Println("Title: ", p.title)
    fmt.Println("Content: ", p.content)
    fmt.Println("Author: ", p.fullName())
    fmt.Println("Bio: ", p.bio)
}
```

现在，我们的 `author` 和 `post` 结构体都已准备就绪，我们来创建一个博客帖子来完成这个程序。

```go
package main

import (
    "fmt"
)

type author struct {
    firstName string
    lastName  string
    bio       string
}

func (a author) fullName() string {
    return fmt.Sprintf("%s %s", a.firstName, a.lastName)
}

type post struct {
    title   string
    content string
    author
}

func (p post) details() {
    fmt.Println("Title: ", p.title)
    fmt.Println("Content: ", p.content)
    fmt.Println("Author: ", p.fullName())
    fmt.Println("Bio: ", p.bio)
}

func main() {
    author1 := author{
        "Naveen",
        "Ramanathan",
        "Golang Enthusiast",
    }
    post1 := post{
        "Inheritance in Go",
        "Go supports composition instead of inheritance",
        author1,
    }
    post1.details()
}
```

[在 playground 上运行](https://play.golang.org/p/sskWaTpJgr)

在上面程序中，main 函数在第 31 行新建了一个 `author` 结构体变量。而在第 36 行，我们通过嵌套 `author1` 来创建一个 `post`。该程序输出：

```bash
Title:  Inheritance in Go
Content:  Go supports composition instead of inheritance
Author:  Naveen Ramanathan
Bio:  Golang Enthusiast
```

## 结构体切片的嵌套

我们可以进一步处理这个示例，使用博客帖子的切片来创建一个网站。:)

我们首先定义 `website` 结构体。请在上述代码里的 main 函数中，添加下面的代码，并运行它。

```go
type website struct {
        []post
}
func (w website) contents() {
    fmt.Println("Contents of Website\n")
    for _, v := range w.posts {
        v.details()
        fmt.Println()
    }
}
```

在你添加上述代码后，当你运行程序时，编译器将会报错，如下所示：

```bash
main.go:31:9: syntax error: unexpected [, expecting field name or embedded type
```

这项错误指出了嵌套的结构体切片 `[]post`。错误的原因是结构体不能嵌套一个匿名切片。我们需要一个字段名。所以我们来修复这个错误，让编译器顺利通过。

```go
type website struct {
        posts []post
}
```

可以看到，我给帖子的切片 `[]post` 添加了字段名 `posts`。

现在我们来修改主函数，为我们的新网站创建一些帖子吧。

修改后的完整代码如下所示：

```go
package main

import (
    "fmt"
)

type author struct {
    firstName string
    lastName  string
    bio       string
}

func (a author) fullName() string {
    return fmt.Sprintf("%s %s", a.firstName, a.lastName)
}

type post struct {
    title   string
    content string
    author
}

func (p post) details() {
    fmt.Println("Title: ", p.title)
    fmt.Println("Content: ", p.content)
    fmt.Println("Author: ", p.fullName())
    fmt.Println("Bio: ", p.bio)
}

type website struct {
 posts []post
}
func (w website) contents() {
    fmt.Println("Contents of Website\n")
    for _, v := range w.posts {
        v.details()
        fmt.Println()
    }
}

func main() {
    author1 := author{
        "Naveen",
        "Ramanathan",
        "Golang Enthusiast",
    }
    post1 := post{
        "Inheritance in Go",
        "Go supports composition instead of inheritance",
        author1,
    }
    post2 := post{
        "Struct instead of Classes in Go",
        "Go does not support classes but methods can be added to structs",
        author1,
    }
    post3 := post{
        "Concurrency",
        "Go is a concurrent language and not a parallel one",
        author1,
    }
    w := website{
        posts: []post{post1, post2, post3},
    }
    w.contents()
}
```

[在 playground 中运行](https://play.golang.org/p/gKaa0RbeAE)

在上面的主函数中，我们创建了一个作者 `author1`，以及三个帖子 `post1`、`post2` 和 `post3`。我们最后通过嵌套三个帖子，在第 62 行创建了网站 `w`，并在下一行显示内容。

程序会输出：

```bash
Contents of Website

Title:  Inheritance in Go
Content:  Go supports composition instead of inheritance
Author:  Naveen Ramanathan
Bio:  Golang Enthusiast

Title:  Struct instead of Classes in Go
Content:  Go does not support classes but methods can be added to structs
Author:  Naveen Ramanathan
Bio:  Golang Enthusiast

Title:  Concurrency
Content:  Go is a concurrent language and not a parallel one
Author:  Naveen Ramanathan
Bio:  Golang Enthusiast
```

本教程到此结束。祝你愉快。

**上一教程 - [结构体取代类](https://studygolang.com/articles/12630)**

**下一教程 - [多态](https://studygolang.com/articles/12681)**

---

via: https://golangbot.com/inheritance

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
