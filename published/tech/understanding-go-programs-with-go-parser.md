已发布：https://studygolang.com/articles/12403

# 通过 `go/parser` 理解 Go

这篇文章所讲内容和 [episode 25 of justforfunc](https://www.youtube.com/watch?v=YRWCa84pykM) 是相同的。

## justforfunc 前情提要

我们在[上一篇文章](https://studygolang.com/articles/12324)中使用 `go/scanner` 找出了标准库中最常用的标识符。

> 这个标识符就是 v

为了能获取到更有价值的信息，我们只考虑大于等于三个字符的标识符。不出所料，在 Go 中最具代表性的判断语句 `if err != nil {}` 中的 err 和 nil 出现的最为频繁。

## 全局变量和局部变量

如果我们想要知道最常用的局部变量名应该怎么做？如果想知道最常用的类型或函数呢？针对这些问题 go/scanner 并不能满足我们的需求，因为它缺少对上下文的支持。按前文的方法我们可以找到需要的 token（例：var a = 3），为了获取 token 所在的作用域（包级，函数级，代码块级）我们需要上下文的支持。

在一个包中可以有很多的声明，其中的一些可能是函数声明，而在函数声明中，又可能有局部变量、常量或函数声明。

但是我们如何在 token 序列中找到这种结构呢？每种编程语言都有从 token 序列到语法树结构的转换规则。就像下面这样：

```
VarDecl = "var" ( VarSpec | "(" { VarSpec ";" } ")" ) .
VarSpec = IdentifierList ( Type [ "=" ExpressionList ] | "="
						ExpressionList ) .
```

这个转换规则告诉我们一个 `VarDecl`（变量声明） 以一个 `var` token 开始，紧接着是一个 `VarSpec`（变量说明）或是一个被括号包围的以分号分隔的标识符列表。

注意：分号其实是 Go scanner 自动添加的，所以你不会在语法分析的时候看到他们。

以 var a = 3 为例，使用 go/scanner 我们会得到这样的 token：

```
[VAR],[IDENT "a"],[ASSIGN],[INT "3"],[SEMICOLON]
```

依据前文描述的规则，这是一个只有 VarSpec 的 VarDecl。紧接着我们分析出标识符列表（`IdentifierList`）里有一个标识符（`Identifier`） `a`，没有类型（`Type`），表达式列表（`ExpressionList`）有一个整数 3 的表达式（`Expression`）。

如果用树表示，会是下面图片这样：

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-parser/0_STJNoHjXJBsnWB4x.png)

这个能使我们能从 token 序列解析树结构的规则叫做语法或句法，而解析出的树结构叫做抽象语法树，简称 AST。

## 使用 go/scanner

现在我们有足够的理论基础来写一些代码。来看看我们如何解析表达式 `var a = 3` 并且获得他的 AST。

```go
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", "var a = 3", parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(f)
}
```

这段代码可以编译通过但是在运行时会报错：

```
1:1: expected 'package', found 'var' (and 1 more errors)
```

为了解析这个我们叫做 `ParseFile` 的声明，我们需要给出一个完整的 go 源文件格式（以 package 作为源文件开头）。

> 注意：注释可以写在 package 前面

如果你正在解析一个形如 `3 + 5` 的表达式或者其他可以看作一个值的代码你可以将它们看作一个参数叫做 ParseExpr。但是在函数声明时不能这么做。

添加 `package main` 到代码的开头并查看我们获得的 AST 树。

```go
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", "package main; var a = 3", parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(f)
}
```

运行后输出如下：

```
$ go run main.go
&{<nil> 1 main [0xc420054100] scope 0xc42000e210 {
		var a
}
 [] [] []}
```

将 `Println` 换成 `fmt.Printf("%#v",f)` 并重试：

```
go run main.go
&ast.File{Doc:(*ast.CommentGroup)(nil), Package:1, Name:(*ast.Ident)(0xc42000a060), Decls:[]ast.Decl{(*ast.GenDecl)(0xc420054100)}, Scope:(*ast.Scope)(0xc42000e210), Imports:[]*ast.ImportSpec(nil), Unresolved:[]*ast.Ident(nil), Comments:[]*ast.CommentGroup(nil)}
```

看起来可以了但是不易读，可以使用 `github.com/davecgh/go-spew/spew` 来让输出更易读：

```go
package main

import (
	"go/parser"
	"go/token"
	"log"

  "github.com/davecgh/go-spew/spew"
)

func main() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", "package main; var a = 3", parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	spew.Dump(f)
}
```

重新运行程序我们会得到更加易读的输出：

```
$ go run main.go
(*ast.File)(0xc42009c000)({
 Doc: (*ast.CommentGroup)(<nil>),
 Package: (token.Pos) 1,
 Name: (*ast.Ident)(0xc42000a120)(main),
 Decls: ([]ast.Decl) (len=1 cap=1) {
  (*ast.GenDecl)(0xc420054100)({
   Doc: (*ast.CommentGroup)(<nil>),
   TokPos: (token.Pos) 15,
   Tok: (token.Token) var,
   Lparen: (token.Pos) 0,
   Specs: ([]ast.Spec) (len=1 cap=1) {
	(*ast.ValueSpec)(0xc4200802d0)({
	 Doc: (*ast.CommentGroup)(<nil>),
	 Names: ([]*ast.Ident) (len=1 cap=1) {
	  (*ast.Ident)(0xc42000a140)(a)
	 },
	 Type: (ast.Expr) <nil>,
	 Values: ([]ast.Expr) (len=1 cap=1) {
	  (*ast.BasicLit)(0xc42000a160)({
	   ValuePos: (token.Pos) 23,
	   Kind: (token.Token) INT,
	   Value: (string) (len=1) "3"
	  })
	 },
	 Comment: (*ast.CommentGroup)(<nil>)
	})
   },
   Rparen: (token.Pos) 0
  })
 },
 Scope: (*ast.Scope)(0xc42000e2b0)(scope 0xc42000e2b0 {
	var a
}
),
 Imports: ([]*ast.ImportSpec) <nil>,
 Unresolved: ([]*ast.Ident) <nil>,
 Comments: ([]*ast.CommentGroup) <nil>
})
```

我推荐花点时间认真看一下这个树，并且找到他们对应的源码部分。`Scope`，`Obj`，`Unresolved` 我们会在下面的章节说。

## 从 AST 到代码

有的时候以源码的位置 打印 AST 比树结构更清晰。使用 go/printer 可以非常简单的打印源码保存的 AST 信息。

```go
package main

import (
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
)

func main() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", "package main; var a = 3", parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	printer.Fprint(os.Stdout, fs, f)
}
```

执行这段代码会打印我们源码的解析结果，将 parser.AllErrors 替换成 parser.ImportsOnly 或者其它值会有不同的输出结果。

## AST 指南

AST 树有我们想知道的所有信息，但是如何才能找出我们想要的信息呢？这时 go/ast 包就派上了用场。

我们使用 ast.Walk。这个函数接受 2 个参数。第二个参数是一个 ast.Node，AST 中所有节点都实现了的接口。第一个参数是 ast.Visitor 接口。

这个接口有一个方法：

```go
type Visitor interface {
	Visit(node Node) (w Visitor)
}
```

现在我们已经有了一个节点，是 `parser.ParseFile` 返回的 `ast.File`。但是我们需要创建一个自己的 `ast.Visitor`。

我们实现了一个打印节点类型并返回自己的 `ast.Visitor`。

```go
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", "package main; var a = 3", parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	var v visitor
	ast.Walk(v, f)
}

type visitor struct{}

func (v visitor) Visit(n ast.Node) ast.Visitor {
	fmt.Printf("%T\n", n)
	return v
}
```

运行这个程序我们会得到没有树结构的节点序列。那些 nil 节点是什么？在 ast.Walk 的文档中可以了解我们返回 visitor 的时候会继续找他的下级节点，如果没有下级节点将会返回 nil。

知道这个特性后我们就可以像树那样打印这个结果。

```go
type visitor int

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	fmt.Printf("%s%T\n", strings.Repeat("\t", int(v)), n)
	return v + 1
}
```

程序的其他部分没有改变，执行以后我们会得到以下输出：

```
*ast.File
	*ast.Ident
	*ast.GenDecl
		*ast.ValueSpec
			*ast.Ident
			*ast.BasicLit
```

## 每种标识符最常用的名称都是什么？

我们已经能够解析代码并访问 AST 节点从而导出我们想要的信息：哪个变量名是包中最常用的。

代码和以前很像都是使用 go/scanner 从命令行读取文件列表。

```go
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage:\n\t%s [files]\n", os.Args[0])
		os.Exit(1)
	}
	fs := token.NewFileSet()
	var v visitor
	for _, arg := range os.Args[1:] {
		f, err := parser.ParseFile(fs, arg, nil, parser.AllErrors)
		if err != nil {
			log.Printf("could not parse %s: %v", arg, err)
			continue
		}
		ast.Walk(v, f)
	}
}

type visitor int

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	fmt.Printf("%s%T\n", strings.Repeat("\t", int(v)), n)
	return v + 1
}
```

执行这段代码我们将会得到所有来自命令行参数的文件的 AST。我们可以试试传入刚刚写的 main.go 文件。

```
$ go build -o parser main.go  && parser main.go
# output removed for brevity
```

改变 visitor 来跟踪每种标识符都被不同的变量声明形式使用了多少次。

首先我们来跟踪短变量声明。因为我们知道它一般都是一个局部变量。

```go
type visitor struct {
	locals map[string]int
}

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	switch d := n.(type) {
	case *ast.AssignStmt:
		for _, name := range d.Lhs {
			if ident, ok := name.(*ast.Ident); ok {
				if ident.Name == "_" {
					continue
				}
				if ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
					v.locals[ident.Name]++
				}
			}
		}
	}
	return v
}
```

检查每个赋值语句的名字是不是需要被忽略的 `_` ，这时我们需要 `Obj` 字段跟踪声明的上下文。

如果 `Obj` 的字段是 nil，说明这个变量不在本文件中定义，所以它不是一个局部变量声明我们可以忽略它。

如果我们对标准库执行这段代码将会得到：

```
7761 err
6310 x
5446 got
4702 i
3821 c
```

有趣的是为什么 v 不在，我们漏掉了什么局部变量的声明的方式么？

## 考虑参数和 range 中的变量

我们漏掉了一对节点类型其实它们也是一种局部变量。

- 函数参数，接收者，返回值名称
- range 语句

因为会大部分沿用之前的代码，所以我们特意为其定义了一个方法。

```go
func (v visitor) local(n ast.Node) {
	ident, ok := n.(*ast.Ident)
	if !ok {
		return
	}
	if ident.Name == "_" || ident.Name == "" {
		return
	}
	if ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
		v.locals[ident.Name]++
	}
}
```

对于参数、返回值和方法接收者，我们都会获取到一个长度为一的标识符列表。再定义一个方法来处理这个标识符列表：

```go
func (v visitor) localList(fs []*ast.Field) {
	for _, f := range fs {
		for _, name := range f.Names {
			v.local(name)
		}
	}
}
```

这样我们就可以处理所有声明局部变量的类型：

```go
case *ast.AssignStmt:
	if d.Tok != token.DEFINE {
		return v
	}
	for _, name := range d.Lhs {
		v.local(name)
	}
case *ast.RangeStmt:
	v.local(d.Key)
	v.local(d.Value)
case *ast.FuncDecl:
	v.localList(d.Recv.List)
	v.localList(d.Type.Params.List)
	if d.Type.Results != nil {
		v.localList(d.Type.Results.List)
	}
```

现在让我们运行这段代码：

```shell
$ ./parser ~/go/src/**/*.go
most common local variable names
  12264 err
  9395 t
  9163 x
  7442 i
  6127 c
```

## 处理 var 声明

现在我们需要进一步处理 var 声明，它有可能是全局变量也有可能是局部变量，并且只有判断其是否为 ast.File 级来判断它是不是全局变量。

为了达到这个目的我们为每个新的文件创建 visitor 用来跟踪文件中的全局变量，这样我们就可以正确的计算出标识符的数量。

我们会在结构体中增加一个 pkgDecls 类型为 map[*ast.GenDecl]bool。在我们的 visitor 中，我们会使用 newVisitor 函数创建一个新的 visitor 并进行初始化工作，而且还会添加 globals 字段来跟踪全局变量标识符被声明的次数。

```go
type visitor struct {
	pkgDecls map[*ast.GenDecl]bool
	globals  map[string]int
	locals   map[string]int
}

func newVisitor(f *ast.File) visitor {
	decls := make(map[*ast.GenDecl]bool)
	for _, decl := range f.Decls {
		if d, ok := decl.(*ast.GenDecl); ok {
			decls[d] = true
		}
	}
	return visitor{
		decls,
		make(map[string]int),
		make(map[string]int),
	}
}
```

我们的 main 函数将会需要为每个文件创建一个新的 visitor 去跟踪汇总结果：

```go
locals, globals := make(map[string]int), make(map[string]int)

for _, arg := range os.Args[1:] {
	f, err := parser.ParseFile(fs, arg, nil, parser.AllErrors)
	if err != nil {
		og.Printf("could not parse %s: %v", arg, err)
		continue
	}
	v := newVisitor(f)
	ast.Walk(v, f)
	for k, v := range v.locals {
		locals[k] += v
	}
	for k, v := range v.globals {
		globals[k] += v
	}
}
```

还有最后一个部分需要完成就是需要跟踪 *ast.GenDecl 节点并找到在变量中的所有声明：

```go
case *ast.GenDecl:
	if d.Tok != token.VAR {
		return v
	}
	for _, spec := range d.Specs {
		if value, ok := spec.(*ast.ValueSpec); ok {
			for _, name := range value.Names {
				if name.Name == "_" {
					continue
				}
				if v.pkgDecls[d] {
					v.globals[name.Name]++
				} else {
					v.locals[name.Name]++
				}
			}
		}
	}
```

在每个声明中我们都只计算以 `token.VAR` 开头的声明。因此常量、类型和其他形式的标识符都会被忽略。在每个声明中我们还要判断它是全局变量还是局部变量，并相应的记录出现次数并忽略 `_`。

程序的完全版在 [这里](https://github.com/campoy/justforfunc/blob/master/25-go-parser/main.go)，

执行程序我们会得到：

```shell
$ ./parser ~/go/src/**/*.go
most common local variable names
  12565 err
  9876 x
  9464 t
  7554 i
  6226 b
most common global variable names
	29 errors
	28 signals
	23 failed
	15 tests
	12 debug
```

至此，我们得出结论，最常用的局部变量就是 err。最常用的包名是 errors。

哪个常量名字最常用？我们如何找到他们？

## 感谢

如果你喜欢这篇文章可以分享它也可以订阅我们的频道，或者在关注我。也可以考虑成为一个赞助者。

---

via: https://medium.com/@francesc/understanding-go-programs-with-go-parser-c4e88a6edb87

作者：[JohnKoepi](https://sitano.github.io/)
译者：[saberuster](https://github.com/saberuster)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
