首发于：https://studygolang.com/articles/14062

# 在 Go 语言中自动插入分号

正式的语法指定什么是在 Go 语言（或者其他的语言）中依据语法构成有效的程序。

```
Block = "{" StatementList "}" .
StatementList = { Statement ";" } .
```

以上定义取自 Go 规范。它使用[扩展的 Backus-Naur](https://en.wikipedia.org/wiki/Extended_Backus%E2%80%93Naur_form) 形式（EBNF）。这意味着代码块是用分号分隔的一个或多个语句。函数调用是一个表达的例子。我们可以创建一个简单的代码块：

```go
{
    fmt.Println(1);
    fmt.Println(2);
}
```

经验丰富的 Gophers 们应该注意到通常代码的每行末尾没有使用分号。它可以简化为：

```
{
    fmt.Println(1)
    fmt.Println(2)
}
```

这样的代码和上一个工作方式相同。既然语法需要分号，又是什么使它可运行的？

## 根源

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/automatic-semicolon-insertion-in-go/1.jpeg)

语言设计者们开始摆脱分号之类的标记的原因是什么呢？这个答案相当简单。都是关于可读性。人造代码越少，它就越容易使用。这是很重要的，因为一旦写下一段代码很可能会被不同的人阅读。

语法使用分号作为产品终结者。由于目标是让程序员不必键入这些分号，所以必须有一个自动注入它们的方法。这就是 Go 的 lexer [正在做的事情](https://github.com/golang/go/blob/1106512db54fc2736c7a9a67dd553fc9e1fca742/src/go/scanner/scanner.go#L641)。分号被添加当行的最后一个标记是以下标记之一时：

+ 一个[标识符](https://golang.org/ref/spec#Identifiers)
+ 一个[整数](https://golang.org/ref/spec#Integer_literals)，[浮点数](https://golang.org/ref/spec#Floating-point_literals)，[虚数](https://golang.org/ref/spec#Imaginary_literals)，[符号](https://golang.org/ref/spec#Rune_literals)或[字符串](https://golang.org/ref/spec#String_literals)。
+ [关键字](https://golang.org/ref/spec#Keywords) 之一 break, continure, fallthrough 或 return
+ [运算符和分隔符](https://golang.org/ref/spec#Operators_and_Delimiters)之一 ++，--，)，] 或 }

让我们看个例子：

```go
func g() int {
    return 1
}
func f() func(int) {
    return func(n int) {
        fmt.Println("Inner func called")
    }
}
```

有这样的定义，我们可以分析俩种情况：

```go
f()
(g())
```

和：

```go
f()(g())
```

第一个片段没有打印任何东西，但是第二个给出内部函数调用。这是因为前面提到的第4条规则：因为最后的记号都是圆括号，所以两行后面都加了分号。

```go
f();
(g());
```

## 在底层

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/automatic-semicolon-insertion-in-go/2.jpeg)

Golang 在语法分析（扫描）时加入分号。在处理 `.go` 文件开始的时，字符被转换为标识符，数字，关键字等。扫码器时用 Go 实现的，所以我们可以很容易使用它：

```go
package main
import (
    "fmt"
    "go/scanner"
    "go/token"
)
func main() {
    scanner := scanner.Scanner{}
    source := []byte("n := 1\nfmt.Println(n)")
    errorHandler := func(_ token.Position, msg string) {
        fmt.Printf("error handler called: %s\n", msg)
    }
    fset := token.NewFileSet()
    file := fset.AddFile("", fset.Base(), len(source))
    scanner.Init(file, source, errorHandler, 0)
    for {
        position, tok, literal := scanner.Scan()
        fmt.Printf("%d: %s", position, tok)
        if literal != ""{
            fmt.Printf(" %q", literal)
        }
        fmt.Println()
        if tok == token.EOF {
            break
        }
    }
}
```

输出：

```
1: IDENT "n"
3: :=
6: INT "1"
7: ; "\n"
8: IDENT "fmt"
11: .
12: IDENT "Println"
19: (
20: IDENT "n"
21: )
22: ; "\n"
22: EOF
```

行打印 ; "\n" 是扫描器（lexer）为丞相添加分号的地方：

```go
n := 1
fmt.Println(n)
```

golangspec 已有 300 多关注者。这并不是它的目标，但有越来越多的人认为它是一个有用的出版物时是非常有动力的。

喜欢加关注，双击 666。

---

via: https://medium.com/golangspec/automatic-semicolon-insertion-in-go-1990338f2649

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[themoonbear](https://github.com/themoonbear)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
