首发于：https://studygolang.com/articles/18793

# 深入理解 Go 语言中的 Testable Examples

隐藏的 `ast` 和 `parser` 包的介绍

2016 年 10 月 10 日

Golang 的工具链实现了名为 `Testable Examples` 的功能。如果对该功能没有什么印象的话，我强烈建议首先阅读[“ Testable Examples in Go ”](https://blog.golang.org/examples) 博文进行了解。通过这篇文章我们将了解到该功能的整个解决方案以及如何构建其简化版本。

让我们看看 `Testable Examples` 的工作原理：

upper_test.go：

```go
package main
import (
    "fmt"
    "strings"
)
func ExampleToUpperOK() {
    fmt.Println(strings.ToUpper("foo"))
    // Output: FOO
}
func ExampleToUpperFail() {
   fmt.Println(strings.ToUpper("bar"))
   // Output: BAr
}
> Go test -v
=== RUN   ExampleToUpperOK
--- PASS: ExampleToUpperOK (0.00s)
=== RUN   ExampleToUpperFail
--- FAIL: ExampleToUpperFail (0.00s)
got:
BAR
want:
BAr
FAIL
exit status 1
FAIL    Github.com/mlowicki/sandbox     0.008s
```

与测试函数一样的 `Examples` 放在 `xxx_test.go` 文件中，但前缀为 `Example` 而不是 `Test`。`go test` 命令使用特殊格式的注释（`Output：something`）并将它们与捕获的数据进行比较，通常写入 `stdout`。其他工具（例如 `godoc`）使用相同的注释来丰富自动生成的文档。

问题是 `go test` 或 `godoc` 如何从特殊注释中提取数据？语言中是否有任何秘密机制使其成为可能？或者也许一切都可以用众所周知的结构来实现？

事实证明，标准库提供了与 Go 本身解析源代码相关的元素（分布在几个包中）。这些工具生成抽象语法树并提供访问特殊注释的途径。

## 抽象语法树（AST）

AST 是解析时在源代码中找到的元素的树形表示。让我们考虑一个简单的表达式：

```go
9 /（2 + 1）
```

可以使用代码段生成[AST](https://en.wikipedia.org/wiki/Abstract_syntax_tree)：

```go
expr, err := parser.ParseExpr("9 / (2 + 1)")
if err != nil {
    log.Fatal(err)
}
ast.Print(nil, expr)
```

输出：

```go
0 *ast.BinaryExpr {
1 . X: *ast.BasicLit {
2 . . ValuePos: 1
3 . . Kind: INT
4 . . Value: "9"
5 . }
6 . OpPos: 3
7 . Op: /
8 . Y: *ast.ParenExpr {
9 . . Lparen: 5
10 . . X: *ast.BinaryExpr {
11 . . . X: *ast.BasicLit {
12 . . . . ValuePos: 6
13 . . . . Kind: INT
14 . . . . Value: "2"
15 . . . }
16 . . . OpPos: 8
17 . . . Op: +
18 . . . Y: *ast.BasicLit {
19 . . . . ValuePos: 10
20 . . . . Kind: INT
21 . . . . Value: "1"
22 . . . }
23 . . }
24 . . Rparen: 11
25 . }
26 }
```

使用图表可以简化输出，其中树形结构更明显：

```
         (operator: /)
        /             \
       /               \
  (integer: 9) (parenthesized expression)
                        |
                        |
                  (operator: +)
                 /             \
                /               \
          (integer: 2)      (integer: 1)
```

使用 AST 时，两个标准包是至关重要的：

- [parser](https://golang.org/pkg/go/parser/) 提供用于解析用 Go 编写的源代码的结构
- [ast](https://golang.org/pkg/go/ast/) 实现用于在 Go 中使用 AST 代码的原始结构

通常在[词法分析](https://en.wikipedia.org/wiki/Lexical_analysis) 期间会删除注释。有一个特殊的标志来保存注释并将它们放入 AST - [parser.ParseComments](https://golang.org/pkg/go/parser/#Mode)：

```go
import (
    "fmt"
    "go/parser"
    "go/token"
    "log"
)
func main() {
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, "t.go", nil, parser.ParseComments)
    if err != nil {
        log.Fatal(err)
    }
    for _, group := range f.Comments {
        fmt.Printf("Comment group %#v\n", group)
        for _, comment := range group.List {
            fmt.Printf("Comment %#v\n", comment)
        }
    }
}
```

> `parser.ParseFile` 的第三个参数是传递给 `f.ex` 的可选参数 , 类型可以是 `string` 或 `io.Reader`。由于我使用了磁盘中的文件，因此设置为 `nil` 。

t.go：

```go
package main
import "fmt"
// a
// b
func main() {
    // c
    fmt.Println("boom!")
}
```

输出：

```
Comment group &ast.CommentGroup{List:[]*ast.Comment{(*ast.Comment)(0x820262220), (*ast.Comment)(0x820262240)}}
Comment &ast.Comment{Slash:29, Text:"// a"}
Comment &ast.Comment{Slash:34, Text:"// b"}
Comment group &ast.CommentGroup{List:[]*ast.Comment{(*ast.Comment)(0x8202622c0)}}
Comment &ast.Comment{Slash:55, Text:"// c"}
```

### [Comment group](https://golang.org/pkg/go/ast/#CommentGroup)

指的是一系列注释，中间没有任何元素。在上面的示例中，注释 ` “ a ” ` 和 ` “ b ” ` 属于同一组。

### [Pos](https://golang.org/pkg/go/token/#Pos) & [Position](https://golang.org/pkg/go/token/#Position)

源代码中元素的位置使用 `Pos` 类型记录（其更详细的对应点是 `Position`）。它是一个单一的整数值，它对像 `line` 或 `column` 这样的信息进行编码，但 `Position struct` 将它们保存在不同的字段中。在外循环添加：

```go
fmt.Printf("Position %#v\n", fset.PositionFor(group.Pos(), true))
```

程序额外输出：

```go
Position token.Position{Filename:"t.go", Offset:28, Line:5, Column:1}
Position token.Position{Filename:"t.go", Offset:54, Line:9, Column:2}
```

### [Fileset](https://golang.org/pkg/go/token/#FileSet)

位置相对于解析文件集计算。每个文件都分配了不相交的范围，每个位置都位于其中一个范围内。在我们的例子中，我们只有一个，但需要整个集合来解码 `Pos`：

```go
fset.PositionFor(group.Pos(), true)
```

## 树遍历

包 `ast` 为深度优先遍历 AST 提供了方便的功能：

```go
ast.Inspect(f, func(n ast.Node) bool {
    if n != nil {
        fmt.Println(n)
    }
    return true
})
```

由于我们知道如何提取所有注释，现在是时候找到所有顶级的 `ExampleXXX` 函数了。

### [doc.Examples](https://golang.org/pkg/go/doc/#Examples)

包 `doc` 提供了完全符合我们需要的功能：

```go
package main
import (
    "fmt"
    "go/doc"
    "go/parser"
    "go/token"
    "log"
)
func main() {
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, "e.go", nil, parser.ParseComments)
    if err != nil {
        log.Fatal(err)
    }
    examples := doc.Examples(f)
    for _, example := range examples {
        fmt.Println(example.Name)
    }
}
```

e.go:

```go
package main
import "fmt"
func ExampleSuccess() {
    fmt.Println("foo")
    // Output: foo
}
func ExampleFail() {
    fmt.Println("foo")
    // Output: bar
}
```

输出：

```
Fail
Success
```

`doc.Examples` 没有任何魔法技能。它依赖于我们已经看到的内容，主要是构建和遍历抽象语法树。让我们建立类似的东西：

```go
package main
import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "log"
    "strings"
)
func findExampleOutput(block *ast.BlockStmt, comments []*ast.CommentGroup) (string, bool) {
    var last *ast.CommentGroup
    for _, group := range comments {
        if (block.Pos() < group.Pos()) && (block.End() > group.End()) {
            last = group
        }
    }
    if last != nil {
        text := last.Text()
        marker := "Output: "
        if strings.HasPrefix(text, marker) {
          return strings.TrimRight(text[len(marker):], "\n"), true
        }
    }
    return "", false
}
func isExample(fdecl *ast.FuncDecl) bool {
    return strings.HasPrefix(fdecl.Name.Name, "Example")
}
func main() {
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, "e.go", nil, parser.ParseComments)
    if err != nil {
        log.Fatal(err)
    }
    for _, decl := range f.Decls {
        fdecl, ok := decl.(*ast.FuncDecl)
        if !ok {
            continue
        }
        if isExample(fdecl) {
            output, found := findExampleOutput(fdecl.Body, f.Comments)
            if found {
                fmt.Printf("%s needs output '%s' \n", fdecl.Name.Name, output)
            }
        }
    }
}
```

输出：

```
ExampleSuccess needs output ‘foo’
ExampleFail needs output 'bar'
```

注释不是 `AST` 树的常规节点。它们可以通过[`ast.File`](https://golang.org/pkg/go/ast/#File) 的 `Comments` 字段访问（由 `f.ex. parser.ParseFile` 返回）。此列表中的注释顺序与它们在源代码中显示的顺序相同。要查找某些块内的注释，我们需要比较上面的 `findExampleOutput` 中的位置：

```go
var last *ast.CommentGroup
for _, group := range comments {
    if (block.Pos() < group.Pos()) && (block.End() > group.End()) {
        last = group
    }
}
```

`if` 语句中的条件检查 `comment group` 是否属于块的范围。

正如我们所看到的那样，标准库在解析时提供了很大的支持。那里的公共类库使整个工作非常愉快，并且精心设计的代码非常紧凑。

如果你喜欢这个帖子并希望获得有关新帖子的更新，请关注我。点击下面的❤，帮助他人发现这些资料。

## 相关资料

- [Testable Examples in Go](https://blog.golang.org/examples)
- [Abstract syntax tree](https://en.wikipedia.org/wiki/Abstract_syntax_tree)
- [Documentation of Go/* packages](https://golang.org/pkg/go/)

---

via: https://medium.com/golangspec/gos-testable-examples-under-the-hood-4a4db8db447f

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[lovechuck](https://github.com/lovechuck)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
