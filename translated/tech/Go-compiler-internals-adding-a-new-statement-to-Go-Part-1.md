# Go compiler internals: adding a new statement to Go - Part 1

This is the first post in a two-part series that takes a tutorial-based approach to exploring the Go compiler. The compiler is large and would require a small book to describe properly, so the idea of these posts is to provide a quick depth-first dive instead. I plan to write more descriptive posts on specific areas of the compiler in the future.

We're going to change the Go compiler to add a new (toy) language feature, and build a modified compiler to play with.

## The task - adding a new statement
Many languages have a `while` statement, which in Go is expressed with `for`:

```
for <some-condition> {
  <loop body>
}
```

Adding a `while` statement to Go is rather trivial, therefore - we simply translate it to `for`. So I chose a slightly more challenging task, adding `until`. `until` is the same as `while` except that the condition is negated. For example, this code:

```go
i := 4
until i == 0 {
  i--
  fmt.Println("Hello, until!")
}
```

Is equivalent to:

```go
i := 4
for i != 0 {
  i--
  fmt.Println("Hello, until!")
}
```

In fact, we could even use an initializer in the loop declaration as follows:

```go
until i := 4; i == 0 {
  i--
  fmt.Println("Hello, until!")
}
```

Our implementation will support this.

A mandatory disclaimer - this is just a toy exercise. I don't think adding `until` to Go is a good idea at all, because Go's minimalism is an absolutely correct design choice.

## High-level structure of the Go compiler
The default Go compiler (`gc`) has a fairly traditional structure that should be immediately familiar if you worked on other compilers before:

![](https://eli.thegreenplace.net/images/2019/go-compiler-flow.png)

Relative to the Go repository root, the compiler implementation lives in `src/cmd/compile/internal`; all the code paths mentioned later in the post are going to be relative to this directory. It's all written in Go and the code is fairly readable. Throughout this post we're going to examine these stages one by one, as we add the required code to support an `until` statement.

Check out the `README` file in `src/cmd/compile` for a nice step-by-step description of the compilation steps. That file is a good companion to this blog post.

## Scan
The scanner (also known as _lexer_) breaks up source code text into discrete entities for the compiler. For example, the word `for` becomes the constant `_For`; the characters `...` become `_DotDotDot`, while `.` on its own becomes `_Dot`, and so on.

The scanner is implemented in the `syntax` package. All we need from it here is to understand a new keyword - `until`. The file `syntax/tokens.go` has a list of all tokens understood by the compiler, and we'll add a new one:

```
_Fallthrough // fallthrough
_For         // for
_Until       // until
_Func        // func
```

The comment on the right-hand side of the token constant is important, as it's used to identify the token in text. This is done by means of code generation from `syntax/tokens.go`, which has this line above the list of tokens:

```go
//go:generate stringer -type token -linecomment
```

`go generate` has to be run manually and the output file (`syntax/token_string.go`) is checked into the Go source repository. To regenerate it I ran the following command from the `syntax` directory:

```
GOROOT=<src checkout> go generate tokens.go
```

The `GOROOT` setting is [essential as of Go 1.12](https://github.com/golang/go/issues/32724), and has to point to the root of the source checkout where we're modifying the compiler.

Having run the code generator and verified that `syntax/token_string.go` now has the new token, I tried rebuilding the compiler and ran into a panic:

```
panic: imperfect hash
```

It comes from this code in `syntax/scanner.go`:

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

The compiler tries to build a "perfect" hash table to perform keyword string to token lookups. By "perfect" it means it wants no collisions, just a linear array where every keyword maps to a single index. The hash function is rather ad-hoc (it only looks at the contents of the first characters of the string token, for example) and it's not easy to debug why a new token creates collisions. To work around it, I increased the lookup table size by changing it to `[1 << 7]token`, thus changing the size of the lookup array from 64 to 128. This gives the hash function much more space to distribute its keys, and the collision went away.

## Parse
Go has a fairly standard recursive-descent parser, which converts a stream of tokens produced by the scanner into a _concrete syntax tree_. We'll start by adding a new node type for `until` in `syntax/nodes.go`:

```go
UntilStmt struct {
  Init SimpleStmt
  Cond Expr
  Body *BlockStmt
  stmt
}
```

I borrowed the overall structure from `ForStmt`, which is used for `for` loops. Similarly to `for`, our `until` statement has several optional sub-statements:

```
until <init>; <cond> {
  <body>
}
```

Both `<init>` and `<cond>` are optional, though it's not common to omit `<cond>`. The `UntilStmt.stmt` embedded field is used for all syntax tree statements and contains position information.

The parsing itself is done in `syntax/parser.go`. The `parser.stmtOrNil` method parses a statement in the current position. It looks at the current token and makes a decision of which statement to parse. Here's an excerpt with the code we're adding:

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

We reuse the existing `parser.header` method which parses a header for `if` and `for` statements. In its most general form, it supports three parts (separated by semicolons). In `for` statements the third part can be used for the ["post" statement](https://golang.org/ref/spec#PostStmt), but we're not going to support this for `until` so we're only interested in the first two. Note that `header` accepts the source token to be able to differentiate between the kinds of statements it's serving; for example it would reject a "post" statement for `if`. We should explicitly reject it for `until` too, though I haven't bothered to implement this right now.

These are all the changes we need for the parser. Since `until` is so similar structurally to existing statements, we could reuse much of the functionality.

If we instrument the compiler to dump out the syntax tree (using `syntax.Fdump`) after parsing and run it on:

```go
i = 4
until i == 0 {
  i--
  fmt.Println("Hello, until!")
}
```

We'll get this fragment for the `until` statement:

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

## Create AST
Now that it has a syntax tree representation of the source, the compiler builds an *abstract syntax tree*. I've written about [Abstract vs. Concrete syntax trees](http://eli.thegreenplace.net/2009/02/16/abstract-vs-concrete-syntax-trees) in the past - it's worth checking out if you're not familiar with the differences. In case of Go, however, this may get changed in the future. The Go compiler was originally written in C and later auto-translated to Go; some parts of it are vestigial from the olden C days, and some parts are newer. Future refactorings may leave only one kind of syntax tree, but right now (Go 1.12) this is the process we have to follow.

The AST code lives in the `gc` package, and the node types are defined in `gc/syntax.go` (not to be confused with the `syntax` package where the CST is defined!)

Go ASTs are structured differently from CSTs. Instead of each node type having its dedicated struct type, all AST nodes are using the `syntax.Node` type which is a kind of a _discriminated union_ that holds fields for many different types. Some fields are generic, however, and used for the majority of node types:

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

We'll start by adding a new constant to identify an `until` node:

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
