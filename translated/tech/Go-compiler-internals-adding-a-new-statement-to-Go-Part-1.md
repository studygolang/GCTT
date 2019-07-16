# Go 编译器内核：给 Go 新增一个语句 —— 第一部分
>Go compiler internals: adding a new statement to Go - Part 1 译文

这是两部分系列文章中的第一部分，该文章采用教程的方式来探讨 Go 编译器。Go 编译器复杂而庞大，需要一本书才可能描述清楚，所以这个系列文章旨在提供一个快速而深度优先的方式进入学习。我计划在以后会写更多关于编译器特定领域的描述性文章。

我们会修改 Go 编译器来增加一个新的（玩具性质）语言特性，并构建一个修改后的编译器来进行使用。

## 任务 —— 增加新的语句
很多语言都有 `while` 语句，在 Go 中对应的是 `for`：

```
for <some-condition> {
  <loop body>
}
```

增加 `while` 语句是比较简单的，因此 —— 我们只需简单将其转换为 `for` 语句。所以我选择了一个稍微有点挑战性的任务，增加 `until`。`until` 语句和 `while` 语句是一样的，只是有了条件判断。例如下面的代码：

```go
i := 4
until i == 0 {
  i--
  fmt.Println("Hello, until!")
}
```

等价于:

```go
i := 4
for i != 0 {
  i--
  fmt.Println("Hello, until!")
}
```

事实上，我们甚至可以像下面代码一样，在循环声明中使用一个初始化语句：

```go
until i := 4; i == 0 {
  i--
  fmt.Println("Hello, until!")
}
```

我们的目标实现将会支持这个。

特别声明 —— 这只是一个玩具性的探索。我觉得在 Go 中添加 `until` 并不好，因为 Go 的极简主义设计思想是非常正确的。

## Go 编译器的高级结构
默认情况下，Go 编译器（`gc`）是以相当传统的结构来设计的，如果你使用过其他编译器，你应该很快就能熟悉它：

![](https://eli.thegreenplace.net/images/2019/go-compiler-flow.png)

Go 仓库中相对路径的根目录下，编译器实现位于 `src/cmd/compile/internal`；本文后续提到的所有代码路径都是相对于这个目录的。编译器是用 Go 编写的，代码可读性很强。在这篇文章中，我们将一点一点的研究这些代码，同时添加支持 `until` 语句的实现代码.

查看 `src/cmd/compile` 中的 `README` 文件，了解编译步骤的详细说明。它将与本文息息相关。

## 扫描
扫描器（也称为 _词法分析器_ ）将源码文本分解为编译器所需的离散实体。例如 `for` 关键字会转变成常量 `_For`；符号 `...` 转变成 `_DotDotDot`，`.` 将转变成 `_Dot` 等等。

扫描器的实现位于 `syntax` 包中。我们需要做的就是理解关键字 —— `until`。`syntax/tokens.go` 文件中列出了所有 token，我们要添加一个新的：

```
_Fallthrough // fallthrough
_For         // for
_Until       // until
_Func        // func
```

token 常量右侧的注释非常重要，它们用来标识 token。这是通过 `syntax/tokens.go` 生成代码来实现的，文件上面的 token 列表有如下这一行：

```go
//go:generate stringer -type token -linecomment
```

`go generate` 必须手动执行，输出文件（`syntax/token_string.go`）被存入 Go 源码仓库中。为了重新生成它，我在 `syntax` 目录中执行如下命令：

```
GOROOT=<src checkout> go generate tokens.go
```

环境变量 `GOROOT` 是[从 Go 1.12 开始必须设置](https://github.com/golang/go/issues/32724)，并且必须指向检出的源码根目录，我们要修改这个编译器。

运行代码生成器并验证包含新的 token 的 `syntax/token_string.go` 文件，我试着重新编译编译器，却出现了 panic 提示：

```
panic: imperfect hash
```

这个 panic 是 `syntax/scanner.go` 中代码引起的：

```go
// hash is a perfect hash function for keywords.
// It assumes that s has at least length 2.
func hash(s []byte) uint {
  return (uint(s[0])<<4 ^ uint(s[1]) + uint(len(s))) & uint(len(keywordMap)-1)
}

var keywordMap [1 << 6]token // size must be power of two

func init() {
  // populate keywordMap
  for tok := _Break; tok <= _Var; tok++ {
    h := hash([]byte(tok.String()))
    if keywordMap[h] != 0 {
      panic("imperfect hash")
    }
    keywordMap[h] = tok
  }
}
```

编译器试图构建一个“完美”哈希表来执行关键字字符串以及 token 查询。“完美”意味着它不太可能发生冲突，是一个线性的数组，其中每个关键字都映射为一个单独的索引。哈希函数相当特殊（例如，它查看字符串 token 的第一个字符），并且不容易调试新 token 为何出现冲突等问题。为了解决这个问题，我将查找表的大小更改为 `[1 << 7]token`，从而将查找数组的大小从 64 改成 128。这给予哈希函数更多的空间来分配对应的键，冲突也就消失了。

## 解析
Go 有一个相当标准的递归下降算法的解析器，它把扫描生成的 token 流转换为 _具体语法树_。我们开始为 `syntax/nodes.go` 中的 `until` 添加新的节点类型：

```go
UntilStmt struct {
  Init SimpleStmt
  Cond Expr
  Body *BlockStmt
  stmt
}
```

我借鉴了用于 `for` 循环的 `ForStmt` 的整体结构。类似于 `for`，`until` 语句有几个可选的子语句：

```
until <init>; <cond> {
  <body>
}
```

Both `<init>` and `<cond>` are optional, though it's not common to omit `<cond>`. The `UntilStmt.stmt` embedded field is used for all syntax tree statements and contains position information.
`<init>` 和 `<cond>` 是可选的，不过省略 `<cond>` 也不是很常见。`UntilStmt.stmt` 中嵌入的字段用于所有语法树语句，并包含对应的位置信息。

解析过程在 `syntax/parser.go` 中实现。`parser.stmtOrNil` 方法解析当前位置的语句。它查看当前 token 并决定解析哪个语句。下方是添加的代码片段：

```go
switch p.tok {
case _Lbrace:
  return p.blockStmt("")

// ...

case _For:
  return p.forStmt()

case _Until:
  return p.untilStmt()
```

And this is `untilStmt`:

```go
func (p *parser) untilStmt() Stmt {
  if trace {
    defer p.trace("untilStmt")()
  }

  s := new(UntilStmt)
  s.pos = p.pos()

  s.Init, s.Cond, _ = p.header(_Until)
  s.Body = p.blockStmt("until clause")

  return s
}
```

我们复用现有的 `parser.header` 方法，因为它解析了 `if` 和 `for` 语句对应的 header。在常用的形式中，它支持三个部分（分号分隔）。在 `for` 语句中，第三部分常被用于 ["post" 语句](https://golang.org/ref/spec#PostStmt)，但我们不打算为 `until` 实现这种形式，而只需实现前两部分。注意 `header` 接收源 token，以便能够区分 `header` 所处的具体场景；例如，编译器会拒绝 `if` 的“post”语句。虽然现在还没有费力气实现“post”语句，但我们在 `until` 的场景中应该明确地拒绝“post”语句。

这些都是我们需要对解析器进行的修改。因为 `until` 语句在结构上跟现有的一些语句非常相似，所以我们可以重用已有的大部分功能。

假如在编译器解析后输出语法树（使用 `syntax.Fdump`）然后使用语法树：

```go
i = 4
until i == 0 {
  i--
  fmt.Println("Hello, until!")
}
```

我们会得到 `until` 语句的相关片段：

```
84  .  .  .  .  .  3: *syntax.UntilStmt {
 85  .  .  .  .  .  .  Init: nil
 86  .  .  .  .  .  .  Cond: *syntax.Operation {
 87  .  .  .  .  .  .  .  Op: ==
 88  .  .  .  .  .  .  .  X: i @ ./useuntil.go:13:8
 89  .  .  .  .  .  .  .  Y: *syntax.BasicLit {
 90  .  .  .  .  .  .  .  .  Value: "0"
 91  .  .  .  .  .  .  .  .  Kind: 0
 92  .  .  .  .  .  .  .  }
 93  .  .  .  .  .  .  }
 94  .  .  .  .  .  .  Body: *syntax.BlockStmt {
 95  .  .  .  .  .  .  .  List: []syntax.Stmt (2 entries) {
 96  .  .  .  .  .  .  .  .  0: *syntax.AssignStmt {
 97  .  .  .  .  .  .  .  .  .  Op: -
 98  .  .  .  .  .  .  .  .  .  Lhs: i @ ./useuntil.go:14:3
 99  .  .  .  .  .  .  .  .  .  Rhs: *(Node @ 52)
100  .  .  .  .  .  .  .  .  }
101  .  .  .  .  .  .  .  .  1: *syntax.ExprStmt {
102  .  .  .  .  .  .  .  .  .  X: *syntax.CallExpr {
103  .  .  .  .  .  .  .  .  .  .  Fun: *syntax.SelectorExpr {
104  .  .  .  .  .  .  .  .  .  .  .  X: fmt @ ./useuntil.go:15:3
105  .  .  .  .  .  .  .  .  .  .  .  Sel: Println @ ./useuntil.go:15:7
106  .  .  .  .  .  .  .  .  .  .  }
107  .  .  .  .  .  .  .  .  .  .  ArgList: []syntax.Expr (1 entries) {
108  .  .  .  .  .  .  .  .  .  .  .  0: *syntax.BasicLit {
109  .  .  .  .  .  .  .  .  .  .  .  .  Value: "\"Hello, until!\""
110  .  .  .  .  .  .  .  .  .  .  .  .  Kind: 4
111  .  .  .  .  .  .  .  .  .  .  .  }
112  .  .  .  .  .  .  .  .  .  .  }
113  .  .  .  .  .  .  .  .  .  .  HasDots: false
114  .  .  .  .  .  .  .  .  .  }
115  .  .  .  .  .  .  .  .  }
116  .  .  .  .  .  .  .  }
117  .  .  .  .  .  .  .  Rbrace: syntax.Pos {}
118  .  .  .  .  .  .  }
119  .  .  .  .  .  }
```

## 创建 AST
Now that it has a syntax tree representation of the source, the compiler builds an *abstract syntax tree*. I've written about [Abstract vs. Concrete syntax trees](http://eli.thegreenplace.net/2009/02/16/abstract-vs-concrete-syntax-trees) in the past - it's worth checking out if you're not familiar with the differences. In case of Go, however, this may get changed in the future. The Go compiler was originally written in C and later auto-translated to Go; some parts of it are vestigial from the olden C days, and some parts are newer. Future refactorings may leave only one kind of syntax tree, but right now (Go 1.12) this is the process we have to follow.
由于有了源代码的语法树表示，编译器才能构建一个*抽象语法树*。我曾写过关于[抽象 vs 具体语法树](http://eli.thegreenplace.net/2009/02/16/abstract-vs-concrete-syntax-trees)的文章 —— 如果你不熟悉他们之间的区别，可以好好看看这个文章。在 Go 中，未来可能会有所变动。Go 编译器最初是用 C 语言编写的，后来自动翻译成 Go；编译器的某些部分是 C 时期遗留下来的，有些部分是比较新的。未来的重构可能只剩下一类语法树，但是现在（Go 1.12）我们必须遵循这个流程。

AST 代码位于 `gc` 包中，节点类型在 `gc/syntax.go` 中定义。（不要跟 `syntax` 包中的 CST 混淆）

Go AST 的结构与 CST 不同。所有的 AST 节点都是 `syntax.Node` 类型而非有各自的类型。`syntax.Node` 类型是一种 _可区分的联合体_，其中的字段有很多不同的类型。然而，这些字段是通用的，并且可用于大多数节点类型：

```go
// A Node is a single node in the syntax tree.
// Actually the syntax tree is a syntax DAG, because there is only one
// node with Op=ONAME for a given instance of a variable x.
// The same is true for Op=OTYPE and Op=OLITERAL. See Node.mayBeShared.
type Node struct {
  // Tree structure.
  // Generic recursive walks should follow these fields.
  Left  *Node
  Right *Node
  Ninit Nodes
  Nbody Nodes
  List  Nodes
  Rlist Nodes

  // ...
```

我们以增加一个新的常量标识 `until` 节点作为开始：

```go
// statements
// ...
OFALL     // fallthrough
OFOR      // for Ninit; Left; Right { Nbody }
OUNTIL    // until Ninit; Left { Nbody }
```

We'll run `go generate` again, this time on `gc/syntax.go`, to generate a string representation for the new node type:

```
// from the gc directory
GOROOT=<src checkout> go generate syntax.go
```

This should update the `gc/op_string.go` file to include `OUNTIL`. Now it's time to write the actual CST->AST conversion code for our new node type.

The conversion is done in `gc/noder.go`. We'll keep modeling our changes after the existing `for` statement support, starting with `stmtFall` which has a switch for statement types:

```go
case *syntax.ForStmt:
  return p.forStmt(stmt)
case *syntax.UntilStmt:
  return p.untilStmt(stmt)
```

And the new `untilStmt` method we're adding to the `noder` type:

```go
// untilStmt converts the concrete syntax tree node UntilStmt into an AST
// node.
func (p *noder) untilStmt(stmt *syntax.UntilStmt) *Node {
  p.openScope(stmt.Pos())
  var n *Node
  n = p.nod(stmt, OUNTIL, nil, nil)
  if stmt.Init != nil {
    n.Ninit.Set1(p.stmt(stmt.Init))
  }
  if stmt.Cond != nil {
    n.Left = p.expr(stmt.Cond)
  }
  n.Nbody.Set(p.blockStmt(stmt.Body))
  p.closeAnotherScope()
  return n
}
```

Recall the generic `Node` fields explained above. Here we're using the `Init` field for the optional initializer, the `Left` field for the condition and the `Nbody` field for the loop body.

This is all we need to construct AST nodes for `until` statements. If we dump the AST after construction, we get:

```
.   .   UNTIL l(13)
.   .   .   EQ l(13)
.   .   .   .   NAME-main.i a(true) g(1) l(6) x(0) class(PAUTO)
.   .   .   .   LITERAL-0 l(13) untyped number
.   .   UNTIL-body
.   .   .   ASOP-SUB l(14) implicit(true)
.   .   .   .   NAME-main.i a(true) g(1) l(6) x(0) class(PAUTO)
.   .   .   .   LITERAL-1 l(14) untyped number

.   .   .   CALL l(15)
.   .   .   .   NONAME-fmt.Println a(true) x(0) fmt.Println
.   .   .   CALL-list
.   .   .   .   LITERAL-"Hello, until!" l(15) untyped string
```

## Type-check
The next step in compilation is type-checking, which is done on the AST. In addition to detecting type errors, type-checking in Go also includes _type inference_, which allows us to write statements like:

```go
res, err := func(args)
```

Without declaring the types of `res` and `err` explicitly. The Go type-checker does a few more tasks, like linking identifiers to their declarations and computing compile-time constants. The code is in `gc/typecheck.go`. Once again, following the lead of the `for` statement, we'll add this clause to the switch in `typecheck`:

```go
case OUNTIL:
  ok |= ctxStmt
  typecheckslice(n.Ninit.Slice(), ctxStmt)
  decldepth++
  n.Left = typecheck(n.Left, ctxExpr)
  n.Left = defaultlit(n.Left, nil)
  if n.Left != nil {
    t := n.Left.Type
    if t != nil && !t.IsBoolean() {
      yyerror("non-bool %L used as for condition", n.Left)
    }
  }
  typecheckslice(n.Nbody.Slice(), ctxStmt)
  decldepth--
```

It assigns types to parts of the statement, and also checks that the condition is valid in a boolean context.

## Analyze and rewrite AST
After type-checking, the compiler goes through several stages of AST analysis and rewrite. The exact sequence is laid out in the `gc.Main` function in `gc/main.go`. In compiler nomenclature such stages are usually called _passes_.

Many passes don't require modifications to support `until` because they act generically on all statement kinds (here the generic structure of `gc.Node` comes useful). Some still do, however. For example escape analysis, which tries to find which variables "escape" their function scope and thus have to be allocated on the heap rather than on the stack.

Escape analysis works per statement type, so we have to add this switch clause in `Escape.stmt`:

```go
case OUNTIL:
  e.loopDepth++
  e.discard(n.Left)
  e.stmts(n.Nbody)
  e.loopDepth--
```

Finally, `gc.Main` calls into the portable code generator (`gc/pgen.go`) to compile the analyzed code. The code generator starts by applying a sequence of AST transformations to lower the AST to a more easily compilable form. This is done in the `compile` function, which starts by calling `order`.

This transformation (in `gc/order.go`) reorders statements and expressions to enforce evaluation order. For example it will rewrite `foo /= 10` to `foo = foo / 10`, replace multi-assignment statements by multiple single-assignment statements, and so on.

To support `until` statements, we'll add this to `Order.stmt`:

```go
case OUNTIL:
  t := o.markTemp()
  n.Left = o.exprInPlace(n.Left)
  n.Nbody.Prepend(o.cleanTempNoPop(t)...)
  orderBlock(&n.Nbody, o.free)
  o.out = append(o.out, n)
  o.cleanTemp(t)
```

After `order`, `compile` calls `walk` which lives in `gc/walk.go`. This pass collects a bunch of AST transformations that helps lower the AST to SSA later on. For example, it rewrites `range` clauses in `for` loops to simpler forms of `for` loops with an explicit loop variable [\[1\]](https://eli.thegreenplace.net/2019/go-compiler-internals-adding-a-new-statement-to-go-part-1/#id2). [It also rewrites map accesses to runtime calls](https://dave.cheney.net/2018/05/29/how-the-go-runtime-implements-maps-efficiently-without-generics), and much more.

To support a new statement in `walk`, we have to add a switch clause in the `walkstmt` function. Incidentally, this is also the place where we can "implement" our `until` statement by rewriting it into AST nodes the compiler already knows how to handle. In the case of `until` it's easy - we just rewrite it into a `for` loop with an inverted condition, as shown in the beginning of the post. Here is the transformation:

```go
case OUNTIL:
  if n.Left != nil {
    walkstmtlist(n.Left.Ninit.Slice())
    init := n.Left.Ninit
    n.Left.Ninit.Set(nil)
    n.Left = nod(ONOT, walkexpr(n.Left, &init), nil)
    n.Left = addinit(n.Left, init.Slice())
    n.Op = OFOR
  }
  walkstmtlist(n.Nbody.Slice())
```

Note that we replace n.Left (the condition) with a new node of type ONOT (which represents the unary ! operator) wrapping the old n.Left, and we replace n.Op by OFOR. That's it!

If we dump the AST again after the walk, we'll see that the OUNTIL node is gone and a new OFOR node takes its place.

## Trying it out
We can now try out our modified compiler and run a sample program that uses an `until` statement:

```
$ cat useuntil.go
package main

import "fmt"

func main() {
  i := 4
  until i == 0 {
    i--
    fmt.Println("Hello, until!")
  }
}

$ <src checkout>/bin/go run useuntil.go
Hello, until!
Hello, until!
Hello, until!
Hello, until!
```

It works!

Reminder: `<src checkout>` is the directory where we checked out Go, changed it and compiled it (see Appendix for more details).

## Concluding part 1
This is it for part 1. We've successfully implemented a new statement in the Go compiler. We didn't cover all the parts of the compiler because we could take a shortcut by rewriting the AST of `until` nodes to use `for` nodes instead. This is a perfectly valid compilation strategy, and the Go compiler already has many similar transformations to _canonicalize_ the AST (reducing many forms to fewer forms so the last stages of compilation have less work to do). That said, we're still interested in exploring the last two compilation stages - _Convert to SSA_ and _Generate machine code_. This will be covered in [part 2](http://eli.thegreenplace.net/2019/go-compiler-internals-adding-a-new-statement-to-go-part-2/).

## Appendix - building the Go toolchain
Please start by going over the [Go contribution guide](https://golang.org/doc/contribute.html). Here are a few quick notes on reproducing the modified Go compiler as shown in this post.

There are two paths to proceed: 
    - 1.Clone the official [Go repository](https://github.com/golang/go) and apply the modifications described in this post.
    - 2.Clone my fork of the [Go repository](https://github.com/eliben/go) and check out the `adduntil` branch, where all these changes are already applied along with some debugging helpers.
The cloned directory is where `<src checkout>` points throughout the post.

To compile the toolchain, enter the `src/` directory and run `./make.bash`. We could also run `./all.bash` to run many tests after building it. Running `make.bash` invokes the full 3-step bootstrap process of building Go, but it only takes about 50 seconds on my (aging) machine.

Once built, the toolchain is installed in `bin` alongside `src`. We can then do quicker rebuilds of the compiler itself by running `bin/go` install `cmd/compile`.

[\[1\]](https://eli.thegreenplace.net/2019/go-compiler-internals-adding-a-new-statement-to-go-part-1/#id1)	Go has some special "magic" `range` clauses like a `range` over a string which splits its up into runes. This is where such transformations are implemented.

For comments, please send me  [an email](eliben@gmail.com), or reach out [on Twitter](https://twitter.com/elibendersky).

----------------

via: [Go compiler internals: adding a new statement to Go - Part 1](https://eli.thegreenplace.net/2019/go-compiler-internals-adding-a-new-statement-to-go-part-1/)

作者：[Eli Bendersky](https://eli.thegreenplace.net)
译者：[suhanyujie](https://github.com/suhanyujie)
校对：

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
