首发于：https://studygolang.com/articles/15648

# Go 程序到机器码的编译之旅

在 [Stream](https://getstream.io/) 里，我们广泛地使用 Go，它也的确巨大地提高了我们的生产效率。我们也发现 Go 语言性能的确出众。自从使用了 Go 以后，我们也完成了类似于我们内部使用的基于 gRPC, Raft 和 RocksDB 存储引擎这类技术栈关键性部分的目标。

今天我们根据 Go 1.11 版本的编译器，来看一下它是如何将我们的 Go 源代码编译成可执行程序的。以此我们能更加了解我们每天工作所用到的工具。我们也会看到为何 Go 语言如此之快并且编译器在其中所起到的作用。我们会从编译器的下述三个阶段入手：

* Scanner（扫描器）将源代码转换为一系列的 token，以供 Parser 使用。
* Parser（语法分析器）将这些 token 转换为 AST（Abstract Syntax Tree, 抽象语法树），以供代码生成。
* 代码生成阶段，将 AST 转换为机器码。

*注意：我们即将使用的包（go/scanner, go/parser, go/token, go/ast 等等）实质上并没有被 Go 编译器使用，它们主要是提供给其他工具用来操作 Go 源码的。然而，实际的 Go 编译器实现上也是类似的。这是因为 Go 编译器最早是用 C 语言编写的，后来才用 Go 实现了自举。所以它仍然保持了原先的结构。*

## Scanner

任何编译器所做的第一步都是将源代码转换成 token，这就是 Scanner （也被称为 “词法分析器”）所做的事。Token 可以是关键字，字符串值，变量名以及函数名等等。任何合法的程序词都可以用 token 来表示。具体到 Go 来说，我们就可能会得到一个 token 列表，包含 "package", "main", "func" 等等。

在 Go 中，每个 token 都以它所处的位置，类型和原始字面量来表示。Go 甚至允许我们在程序中通过 **go/scanner** 和 **go/scanner** 包来手动执行 scanner。这意味着我们可以观察到当我们的程序经过词法分析阶段之后的样子。为此，我们写个程序来打印一下 Hello World 程序所生成的所有的 token。

程序如下：

```go
package main

import (
	"fmt"
	"go/scanner"
	"go/token"
)

func main() {
	src := []byte(`package main
import "fmt"
func main() {
  fmt.Println("Hello, world!")
}
`)

	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, src, nil, 0)

	for {
		pos, tok, lit := s.Scan()
		fmt.Printf("%-6s%-8s%q\n", fset.Position(pos), tok, lit)

		if tok == token.EOF {
			break
		}
	}
}
```

我们先创建源码字符串，然后初始化 **scanner.Scanner** 来扫描我们的源码。我们可以通过不停地调用 **Scan()** 方法来获取 token 的位置，类型和字面量，直到遇到 **EOF** 标记。

当运行程序后，会得到如下输出：

```
1:1   package "package"
1:9   IDENT   "main"
1:13  ;       "\n"
2:1   import  "import"
2:8   STRING  "\"fmt\""
2:13  ;       "\n"
3:1   func    "func"
3:6   IDENT   "main"
3:10  (       ""
3:11  )       ""
3:13  {       ""
4:3   IDENT   "fmt"
4:6   .       ""
4:7   IDENT   "Println"
4:14  (       ""
4:15  STRING  "\"Hello, world!\""
4:30  )       ""
4:31  ;       "\n"
5:1   }       ""
5:2   ;       "\n"
5:3   EOF     ""
```

如此我们就看到了当编译过程中 Parser 所用到的 token 了。我们也可以看到 scanner 打印出了分号，就像 C 语言等其他语言一样。这就解释了为什么 Go 语言不需要写分号，因为 scanner 给自动添加了。

## Parser

当源码被扫描之后，结果就被传递给了 parser。Parser 用来在编译过程中将 token 转换为抽象语法树（AST）。AST 是源代码的一种结构化的表示方式。在 AST 中，我们能够看出程序的结构，比如函数和常量的声明。

Go 同样提供了包 **go/parser** 和 **go/ast** 给我们用来解析程序和查看 AST。我们可以这样来打印完整的 AST:

```go
package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	src := []byte(`package main
import "fmt"
func main() {
  fmt.Println("Hello, world!")
}
`)

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		log.Fatal(err)
	}

	ast.Print(fset, file)
}
```

输出：

```
     0  *ast.File {
     1  .  Package: 1:1
     2  .  Name: *ast.Ident {
     3  .  .  NamePos: 1:9
     4  .  .  Name: "main"
     5  .  }
     6  .  Decls: []ast.Decl (len = 2) {
     7  .  .  0: *ast.GenDecl {
     8  .  .  .  TokPos: 3:1
     9  .  .  .  Tok: import
    10  .  .  .  Lparen: -
    11  .  .  .  Specs: []ast.Spec (len = 1) {
    12  .  .  .  .  0: *ast.ImportSpec {
    13  .  .  .  .  .  Path: *ast.BasicLit {
    14  .  .  .  .  .  .  ValuePos: 3:8
    15  .  .  .  .  .  .  Kind: STRING
    16  .  .  .  .  .  .  Value: "\"fmt\""
    17  .  .  .  .  .  }
    18  .  .  .  .  .  EndPos: -
    19  .  .  .  .  }
    20  .  .  .  }
    21  .  .  .  Rparen: -
    22  .  .  }
    23  .  .  1: *ast.FuncDecl {
    24  .  .  .  Name: *ast.Ident {
    25  .  .  .  .  NamePos: 5:6
    26  .  .  .  .  Name: "main"
    27  .  .  .  .  Obj: *ast.Object {
    28  .  .  .  .  .  Kind: func
    29  .  .  .  .  .  Name: "main"
    30  .  .  .  .  .  Decl: *(obj @ 23)
    31  .  .  .  .  }
    32  .  .  .  }
    33  .  .  .  Type: *ast.FuncType {
    34  .  .  .  .  Func: 5:1
    35  .  .  .  .  Params: *ast.FieldList {
    36  .  .  .  .  .  Opening: 5:10
    37  .  .  .  .  .  Closing: 5:11
    38  .  .  .  .  }
    39  .  .  .  }
    40  .  .  .  Body: *ast.BlockStmt {
    41  .  .  .  .  Lbrace: 5:13
    42  .  .  .  .  List: []ast.Stmt (len = 1) {
    43  .  .  .  .  .  0: *ast.ExprStmt {
    44  .  .  .  .  .  .  X: *ast.CallExpr {
    45  .  .  .  .  .  .  .  Fun: *ast.SelectorExpr {
    46  .  .  .  .  .  .  .  .  X: *ast.Ident {
    47  .  .  .  .  .  .  .  .  .  NamePos: 6:2
    48  .  .  .  .  .  .  .  .  .  Name: "fmt"
    49  .  .  .  .  .  .  .  .  }
    50  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
    51  .  .  .  .  .  .  .  .  .  NamePos: 6:6
    52  .  .  .  .  .  .  .  .  .  Name: "Println"
    53  .  .  .  .  .  .  .  .  }
    54  .  .  .  .  .  .  .  }
    55  .  .  .  .  .  .  .  Lparen: 6:13
    56  .  .  .  .  .  .  .  Args: []ast.Expr (len = 1) {
    57  .  .  .  .  .  .  .  .  0: *ast.BasicLit {
    58  .  .  .  .  .  .  .  .  .  ValuePos: 6:14
    59  .  .  .  .  .  .  .  .  .  Kind: STRING
    60  .  .  .  .  .  .  .  .  .  Value: "\"Hello, world!\""
    61  .  .  .  .  .  .  .  .  }
    62  .  .  .  .  .  .  .  }
    63  .  .  .  .  .  .  .  Ellipsis: -
    64  .  .  .  .  .  .  .  Rparen: 6:29
    65  .  .  .  .  .  .  }
    66  .  .  .  .  .  }
    67  .  .  .  .  }
    68  .  .  .  .  Rbrace: 7:1
    69  .  .  .  }
    70  .  .  }
    71  .  }
    ..  .  .. // Left out for brevity
    83  }
```

在上述输出中，你可以看到不少关于程序的信息。在 **Decls** 字段中，包含了文件中所有声明的列表，比如 imports, constants, variables 和 functions。在本例中只有两个，我们 **fmt** 包的导入（import）和 main 函数。

为了更加深入，我们看一下下述示意图。它就代表了上述的数据，但只包含了类型信息和红色的代码用以标识各个节点。

![ast-diagram](https://raw.githubusercontent.com/studygolang/gctt-images/master/compile-machine-code/image1-5.png)

main 函数由三个部分组成：函数名，定义和函数体。函数名就是取值为 main 的标识符。Type 字段对应的就是定义，根据情况会包含一系列的参数和返回值。函数体就是一系列的程序语句。这边我们就一条语句。

在 AST 中我们唯一的一条 **fmt.Println** 语句也包含了不少东西。定义是 **ExprStmt**，表示一个表达式。在这里的话，就是一个函数调用。当然，有时候也可以是一个字面量，一个二元表达式（加减表达式），一个一元表达式（取负数操作）或者其他。任何函数调用中的参数也都是一个表达式。

我们的 **ExprStmt** 包含了一个 **CallExpr**，也就是实际的函数调用。它同样也包含了许多部分，最重要的就是 **Fun** 和 **Args**。Fun 包含了对于函数调用的一个引用，在这边就是一个 **SelectorExpr**。因为我们从 fmt 包中选择了 **Println** 标识符。然而，在 AST 中编译器并不知道 **fmt** 是一个包，它也有可能是一个变量。

Args 包含了一个表示函数调用参数的表达式列表。在本例中，我们给函数传递了一个字符串字面量，它被表示为类型为 **STRING** 的一个 **BasicLit**。

很显然我们能从 AST 中推断出很多东西。这表示我们可以更深入地观察 AST 并从文件中找出所有的函数调用。为此，我们需要使用 **ast** 包中的 **Inspect** 函数。这个函数会遍历整棵语法树以便于我们观察所有节点的信息。

为了提取出所有的函数调用，我们使用如下代码：

```go
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
)

func main() {
	src := []byte(`package main
import "fmt"
func main() {
  fmt.Println("Hello, world!")
}
`)

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		log.Fatal(err)
	}

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		printer.Fprint(os.Stdout, fset, call.Fun)
		fmt.Println()

		return false
	})
}
```

这边我们所做的就是从所有节点中找出所有类型为 **\*ast.CallExpr** 的节点，这些节点就代表了函数调用。我们会通过使用 printer 包，传入 **Fun** 成员变量，来打印函数的名称。

这段代码的输出如下：

```
fmt.Println
```

这也就是我们这个简单的程序中所有并且唯一的函数调用。

当 AST 构建完成之后，所有的导入都会依赖于 GOPATH，或者 Go 1.11 及以上版本中的 [modules](https://github.com/golang/go/wiki/Modules)。然后，就会执行类型检查，并且做一些初步优化以使得程序运行的更快。

## 代码生成

当包导入和类型检查完成之后，我们就能确定 Go 程序代码是合法的并可以开始将 AST 转换为（伪）机器代码了。

这个过程的第一步就是将 AST 转换到更低一层的程序形式，具体来说就是 SSA (Static Single Assignment，静态单赋值）。这个中间形式并非最终的机器码但很大程序上已经差不多了。SSA 有一系列的属性使得它更易于优化。最重要的是每个变量在使用前都必定会被预先定义，并且每个变量都会被明确地赋值过一次。

在最初版本的 SSA 生成之后，就会被进行一系列的优化。这些优化被应用于代码的特定阶段使得处理器能够更简单和快速地执行。举例来说，像 **if (false) { fmt.Println(“test”) }** 这类的死代码可以删除，因为永远不可能被执行到。另一个优化的例子是特定的 nil 检查可以移除，因为编译器可以保证它们永远不会失败。

现在让我们看一下下面这个小程序的 SSA 和优化阶段：

```go
package main

import "fmt"

func main() {
	fmt.Println(2)
}
```

你可以看到，这个程序只有一个函数和一个导入。运行会打印出 2。然而，这个程序对于我们了解 SSA 来说已经足够了。

*注意：我们只展示 main 函数的 SSA，因为它才有实际意义*

为了展示生成的 SSA，我们需要对要查看 SSA 的方法设置环境变量 GOSSAFUNC，此处就是 main。我们还需要给编译器传递 -S 标志，这样它才能打印代码并创建一个 HTML 文件。我们也会针对 Linux 64-bit 环境进行编译，从而保证生成的机器码和你这边看到的一样。所以，我们运行：

```bash
$ GOSSAFUNC=main GOOS=linux GOARCH=amd64 go build -gcflags “-S” simple.go
```

这会打印整个 SSA，并生成一个相关联的 ssa.html 文件。

![ssa](https://raw.githubusercontent.com/studygolang/gctt-images/master/compile-machine-code/image3-4.png)

当你打开 ssa.html，你会看到一系列的阶段，大部分被折叠了。最开始的阶段是从 AST 生成 SSA。再然后就是将非特定机器的 SSA 转换成特定机器的 SSA。最后生成最终的机器码 genssa。

起始阶段的代码看起来像这样：

```
b1:
	v1  = InitMem <mem>
	v2  = SP <uintptr>
	v3  = SB <uintptr>
	v4  = ConstInterface <interface {}>
	v5  = ArrayMake1 <[1]interface {}> v4
	v6  = VarDef <mem> {.autotmp_0} v1
	v7  = LocalAddr <*[1]interface {}> {.autotmp_0} v2 v6
	v8  = Store <mem> {[1]interface {}} v7 v5 v6
	v9  = LocalAddr <*[1]interface {}> {.autotmp_0} v2 v8
	v10 = Addr <*uint8> {type.int} v3
	v11 = Addr <*int> {"".statictmp_0} v3
	v12 = IMake <interface {}> v10 v11
	v13 = NilCheck <void> v9 v8
	v14 = Const64 <int> [0]
	v15 = Const64 <int> [1]
	v16 = PtrIndex <*interface {}> v9 v14
	v17 = Store <mem> {interface {}} v16 v12 v8
	v18 = NilCheck <void> v9 v17
	v19 = IsSliceInBounds <bool> v14 v15
	v24 = OffPtr <*[]interface {}> [0] v2
	v28 = OffPtr <*int> [24] v2
If v19 → b2 b3 (likely) (line 6)

b2: ← b1
	v22 = Sub64 <int> v15 v14
	v23 = SliceMake <[]interface {}> v9 v22 v22
	v25 = Copy <mem> v17
	v26 = Store <mem> {[]interface {}} v24 v23 v25
	v27 = StaticCall <mem> {fmt.Println} [48] v26
	v29 = VarKill <mem> {.autotmp_0} v27
Ret v29 (line 7)

b3: ← b1
	v20 = Copy <mem> v17
	v21 = StaticCall <mem> {runtime.panicslice} v20
Exit v21 (line 6)
```

就这个简单的程序也生成了大量的 SSA（总共 35 行）。然而，有很多都是模版并可以被移除（最终的 SSA 版本包含 28 行并且最终的机器码只有 18 行）。

每个 v 都是一个新变量，能够通过点击查看。**b's** 表示语句块。这里我们有三个语句块：**b1**, **b2** 和 **b3**。**b1** 每次都会被执行。**b2** 和 **b3** 是条件语句块，可以从 **b1** 的末尾看到 **If v19 → b2 b3 (likely)**。我们可以点击这行的 **v19** 来查看它的定义。我们能够看到它被定义为 **IsSliceInBounds <bool> v14 v15**。通过查看 Go 编译器源码我们会发现 IsSliceInBounds 检查了 **0 <= arg0 <= arg1** 条件。我们同样可以点击 **v14** 和 **v15** 来查看定义。可以看到 **v14** 定义为 **v14 = Const64 <int> [0]**；**Const64** 是一个 64 位常量整数。**v15** 也是一样但值为 1。所以我们可以得到 **0 <= 0 <= 1**，很显然结果为 **true**。

编译器也能保证这种情况。我们查看 **opt** 阶段（机器独立优化），会发现 **v19** 被重写成了 **ConstBool <bool> [true]**。这会被使用在 **opt deadcode** 阶段，导致 **b3** 被移除。因为 **v19** 条件永远是正确的。

现在我们来看一下当 SSA 被转换为特定机器的 SSA（此处为针对 amd64 架构的机器码）之后 Go 编译器所做的另一个简单的优化。为此，我们要对比一下更低层的死代码。下面就是更低层的阶段：

```
b1:
	BlockInvalid (6)
b2:
	v2 (?) = SP <uintptr>
	v3 (?) = SB <uintptr>
	v10 (?) = LEAQ <*uint8> {type.int} v3
	v11 (?) = LEAQ <*int> {"".statictmp_0} v3
	v15 (?) = MOVQconst <int> [1]
	v20 (?) = MOVQconst <uintptr> [0]
	v25 (?) = MOVQconst <*uint8> [0]
	v1 (?) = InitMem <mem>
	v6 (6) = VarDef <mem> {.autotmp_0} v1
	v7 (6) = LEAQ <*[1]interface {}> {.autotmp_0} v2
	v9 (6) = LEAQ <*[1]interface {}> {.autotmp_0} v2
	v16 (+6) = LEAQ <*interface {}> {.autotmp_0} v2
	v18 (6) = LEAQ <**uint8> {.autotmp_0} [8] v2
	v21 (6) = LEAQ <**uint8> {.autotmp_0} [8] v2
	v30 (6) = LEAQ <*int> [16] v2
	v19 (6) = LEAQ <*int> [8] v2
	v23 (6) = MOVOconst <int128> [0]
	v8 (6) = MOVOstore <mem> {.autotmp_0} v2 v23 v6
	v22 (6) = MOVQstore <mem> {.autotmp_0} v2 v10 v8
	v17 (6) = MOVQstore <mem> {.autotmp_0} [8] v2 v11 v22
	v14 (6) = MOVQstore <mem> v2 v9 v17
	v28 (6) = MOVQstoreconst <mem> [val=1,off=8] v2 v14
	v26 (6) = MOVQstoreconst <mem> [val=1,off=16] v2 v28
	v27 (6) = CALLstatic <mem> {fmt.Println} [48] v26
	v29 (5) = VarKill <mem> {.autotmp_0} v27
Ret v29 (+7)
```

在 HTML 文件中，一些行是灰色的，这意味着它们会被移除或者在下个阶段中改变。举例来说，**v15 (MOVQconst <int> [1])** 是灰色的。通过点击查看 **v15**，我们发现它没有在任何地方被使用。**MOVQconst** 大体上和之前的 **Const64** 是一样的，都是针对 amd64 架构的。所以我们把 **v15** 设置为 1。然而，**v15** 并没有被使用，所以可以被删除。

Go 编译器做了很多这类的优化。所以，虽然从 AST 转换得到的最初版本的 SSA 并非最快的实现，但是编译器会优化 SSA 得到更快速的版本。HTML 文件中的每个阶段都是一种潜在的提速。

## 总结

得益于编译器和优化，Go 是一种非常高生产力和高性能的语言。如果想更加了解关于 Go 编译器，[源码](https://github.com/golang/go/tree/master/src/cmd/compile) 中有非常不错的 README。

如果你想了解更多关于为何 Stream 选择了 Go 或者为何我们从 Python 迁移到 Go 的原因，可以查看我们的 [博客](https://getstream.io/blog/switched-python-go/)。

---

via: https://getstream.io/blog/how-a-go-program-compiles-down-to-machine-code/

作者：[Koen Vlaswinkel](https://getstream.io/blog/author/koen/)
译者：[alfred-zhong](https://github.com/alfred-zhong)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
