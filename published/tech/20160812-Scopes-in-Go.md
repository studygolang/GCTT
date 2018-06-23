已发布：https://studygolang.com/articles/13239

# Go 语言中的作用域

理解 Go 语言中的作用域是怎么起作用的，需要一些关于块的预备知识，这在 “[Go 语言中的代码块](https://studygolang.com/articles/12632)” 文章中有讲。

一个标识符的作用域是标识符与某个值，比如变量、常量、包等，进行绑定的那一部分源码（有时甚至是全部）。

```go
package main
import "fmt"
func main() {
    {
        v := 1
        {
            fmt.Println(v)
        }
        fmt.Println(v)
    }
    // 编译错误：“undefined: v”
    // fmt.Println(v)
}
```

对于有经验的工程师,很容易就能判断出程序的输出应该是这样的：

```
> ./bin/sandbox
1
1
```

最后一行的 fmt.Println 被注释了，因为它会引起编译错误。很快我们就来解释这是为什么。简单的说，变量 v 在包含定义它代码块的大括号之外，就超出它的作用域了。

值得一提的是，给变量赋一个新的值并不影响它的作用域（也叫可见性）：

```go
v := 1
{
    v = 2  // 赋值
    fmt.Println(v)
}
fmt.Println(v)
```

输出：

```
>./bin/sandbox
2
2
```

而且它与下面的代码运行结果不同：

```go
v := 1
{
    v := 2  // 简短变量声明方式
    fmt.Println(v)
}
fmt.Println(v)
```

这段代码的输出是：

```
>./bin/sandbox
2
1
```

作用域与标识符的定义紧密相关（更准确的说是与标识符被声明的地方）

## 变量或者常量的声明

变量标识符的作用域能到达最内层的代码块（不论是隐含的或者是显式用大括号包围起来的）：

```go
func main() {
    {
        v := 1
        {
            fmt.Println(v)
            {
                fmt.Println(v)
            }
        }
        fmt.Println(v)
    }
}
```

这段代码是 100% 有效的代码，运行结果为：

```
> ./bin/sandbox
1
1
1
```

作用域从变量被声明的那一行代码开始。

```go
func main() {
    fmt.Println(v)
    v := 1
}
```

所以这段代码会抛出一个编译错误 “undefined: v”。简短变量声明方式可以一次性声明多个变量：

```go
a, b := 0, 1
```

但是标识符从它被声明语句结束的地方开始有效，所以，下面这一句是错的：

```go
a, b := 1, a  // 未定义的: a
```

对于简短变量声明，上述作用域规则同样适用：

- 变量声明（使用 var 关键字）
- 常量声明（使用 const 声明）

在括起来的变量或常量声明中，变量或者常量从它们被声明的语句之后就生效，而不需要等整个括起来的代码结束，所以下面的代码是有效的：

```go
var (
    a = 1  // 变量声明 no. 1
    b = a  // 变量声明 no. 2
)
fmt.Println(a, b)
```

运行结果是：

```
> ./bin/sandbox
1 1
```

同样的，如果在括起来的声明中，如果用简短方式声明多个变量，

```go
var (
    a, b = 1, a
)
```

这样的代码同样会报编译错误 —— “undefined: a”

## 类型声明

就作用域而言，类型声明与变量或者常量一样 —— 一直作用到最内层的代码块。但是与变量或常量不同的是，类型声明从标识结束的地方就开始生效了，而不是从类型定义代码结束的地方才生效。这一点额外的“空间”让类型递归称为可能：

```go
type X struct {
    name string
    next *X
}
x := X{name: "foo", next: &X{name: "bar"}}
fmt.Println(x.name)
fmt.Println(x.next.name)
fmt.Println(x.next.next)
```

输出：

```
> ./bin/sandbox
foo
bar
<nil>
```

next 字段必须是一个指针，下面这样的定义是不合法的：

```go
type X struct {
    name string
    next X
}
```

因为编译器会抛出 “invalid recursive type X” 的错误，产生这个错误的原因是，当创建类型 X 时，要计算这个类型的大小，而编译器发现类型 X 的 next 字段也是 X 类型，一个同样的还没有确定大小的字段，于是我们会陷入一个无穷递归中。但是如果是一个指针类，编译器就能知道在指定平台上指针类型的确定大小。

## 预定义标识符

有很多内置的标识符：

- 类型：bool, int32, int64, float64, …
- nil
- 函数： make, new, panic, …
- 常量，比如 true/false

它们有全局的作用域，所以它们可以在代码的任何地方使用。

## Imports

当导入包时，包内名称的作用域就是在文件块内。这样包内的标识符只能在包已经被正确导入后，通过 f.ex 的方式来引用。

```go
// sandbox.go
package main
import "fmt"
func main() {
    fmt.Println("main")
    f()
}
// utils.go
package main
func f() {
    fmt.Println("f")
}
```

当编译以上包时，编译器会抛出错误：“undefined: fmt in fmt.Println”。

## 顶级的标识符

在任何函数外声明的变量，常量，类型，函数，在整个包内是可见的（作用域是整个包）

```go
// sandbox.go
package main
func main() {
    f()
}
// utils.go
package main
import "fmt"
func f() {
    fmt.Println("It works!")
}
```

以上代码可以编译并运行输出：

```
> ./bin/sandbox
It works!
```

## 函数和方法

方法的调用者，函数参数 或者 返回值仅仅在函数体内可以访问 —— 这个显而易见的就不用代码演示了。

## 遮蔽（Shadowing）

在同一个代码块中，一个标识符不能被声明两次。但是在内部的代码块中可以重新声明外部被声明了的标识符（代码块可以像洋葱那样一层层嵌套的）。如果在内层重新声明了标识符，那么在代码中起作用的声明是离代码最近的最里层的那个声明:

```go
v := "outer"
fmt.Println(v)
{
    v := "inner"
    fmt.Println(v)
    {
        fmt.Println(v)
    }
}
{
    fmt.Println(v)
}
fmt.Println(v)
```

输出：

```
> ./bin/sandbox
outer
inner
inner
outer
outer
```

## 参考资料

- [《Go语言规范》](https://golang.org/ref/spec#Declarations_and_scope)
- [《Go 语言中的代码块》](https://studygolang.com/articles/12632)

---

via: https://medium.com/golangspec/scopes-in-go-a6042bb4298c

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[MoodWu](https://github.com/MoodWu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
