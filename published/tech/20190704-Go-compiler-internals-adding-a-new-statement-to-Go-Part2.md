首发于：https://studygolang.com/articles/27154

# Go 编译器内部知识：向 Go 添加新语句-第 2 部分

这是探讨 Go 编译器两篇文章的最后一篇。在[第 1 部分中](https://studygolang.com/articles/25101)，我们通过构建自定义的编译器，向 Go 语言添加了一条新语句。为此，我们按照此图介绍了编译器的前五个阶段:

![go compiler flow](https://raw.githubusercontent.com/studygolang/gctt-images/master/compiler-internal/go-compiler-flow.png)

在"rewrite AST"阶段前，我们实现了 until 到 for 的转换；具体来说，在[gc/walk.go](https://github.com/golang/go/blob/master/src/cmd/compile/internal/gc/walk.go)文件中，在编译器进行 SSA 转换和代码生成之前，就已进行了类似的转换。

在这一部分中，我们将通过在编译流程中处理新的 until 关键字来覆盖编译器的剩余阶段。

## SSA

在 GC 运行 walk 变换后，它调用 buildssa([gc/ssa.go](https://github.com/golang/go/blob/master/src/cmd/compile/internal/gc/ssa.go#L281))函数将 AST 转换为[静态单赋值(SSA)形式](https://en.wikipedia.org/wiki/Static_single_assignment_form)的中间表示。

SSA 是什么意思，为什么编译器会这样做？让我们从第一个问题开始；我建议阅读上面链接的 SSA 维基百科页面和其他资源，但这里一个快速说明。

静态单赋值意味着 IR 中分配的每个变量仅分配一次。考虑以下伪 IR：

```
x = 1
y = 7
// do stuff with x and y
x = y
y = func()
// do more stuff with x and y
```

这不是 SSA，因为名称 x 和 y 被分配了多次。如果将此代码片段转换为 SSA，我们可能会得到类似以下内容：

```
x = 1
y = 7
// do stuff with x and y
x_1 = y
y_1 = func()
// do more stuff with x_1 and y_1
```

注意每个赋值如何得到唯一的变量名。当 x 重新分配了另一个值时，将创建一个新名称 x_1。你可能想知道这在一般情况下是如何工作的……像这样的代码会发生什么：

```
x = 1
if condition: x = 2
use(x)
```

如果我们简单地将第二次赋值重命名为 x_1 = 2，那么 use 呢？x 或 x_1 或...呢？为了处理这一重要情况，SSA 形式的 IR 具有特殊的 phi（originally phony）功能，以根据其来自哪个代码路径来选择一个值。它看起来是这样的：

![simple ssa phi](https://raw.githubusercontent.com/studygolang/gctt-images/master/compiler-internal/simple-ssa-phi.png)

编译器使用此 phi 节点来维护 SSA，同时分析和优化此类 IR，并在以后的阶段用实际的机器代码代替。

SSA 名称的静态部分起着与静态类型类似的作用；这意味着在查看源代码时（在编译时或静态时），每个名称的分配都是唯一的，而它可以在运行时发生多次。如果上面显示的代码片段是在一个循环中，那么实际的 x_1 = 2 的赋值可能会发生多次。

现在我们对 SSA 是什么有了基本的了解，接下来的问题是为什么。

优化是编译器后端的重要组成部分[[1](#jump1)]，并且通常对后端进行结构化以促进有效和高效的优化。再次查看此代码段：

```
x = 1
if condition: x = 2
use(x)
```

假设编译器想要运行一个非常常见的优化——常量传播； 也就是说，它想要在 x = 1 的赋值后,将所有的 x 替换为 1。这会怎么样呢？它不能只找到赋值后对 x 的所有引用，因为 x 可以重写为其他内容(例如我们的例子)。

考虑以下代码片段：

```
z = x + y
```

一般情况下，编译器必须执行数据流分析才能找到：

1. x 和 y 指的是哪个定义？存在控制语句情况下，这并不容易，并且还需要进行优势分析（dominance analysis）。
2. 在此定义之后使用 z 时，同样具有挑战性。

就时间和空间而言，这种分析的创建和维护成本很高。此外，它必须在每次优化之后重新运行它(至少一部分)。

SSA 提供了一个很好的选择。如果 z = x + y 在 SSA 中，我们立即知道 x 和 y 所引用的定义(只能有一个)，并且我们立即知道在哪里使用 z(在这个语句之后对 z 的所有引用)。在 SSA 中，用法和定义都在 IR 中进行了编码，并且优化不会违反不变性。

## Go 编译器中的 SSA

我们继续描述 Go 编译器中如何构造和使用 SSA。SSA 是 Go 的一个[相对较新的功能](https://blog.golang.org/go1.7)。除了将 AST 转换为 SSA 的大量代码([gc/ssa.go](https://github.com/golang/go/blob/master/src/cmd/compile/internal/gc/ssa.go))，其它大部分代码都位于[ssa](https://github.com/golang/go/tree/master/src/cmd/compile/internal/ssa)目录中，ssa 目录中的 README 文件是对 Go SSA 的非常有用的说明，请阅读一下！

Go SSA 实现还拥有我见过的一些最好的编译器工具（已经在编译器上工作了很多年）。通过设置 GOSSAFUNC 环境变量，我们将获得一个 HTML 页面，其中包含所有编译阶段以及每个编译阶段之后的 IR，因此我们可以轻松地检索出需要进行哪些优化。额外的设置可以将控制流程图绘制成 SVG。

让我们研究一下从 AST 为该以下代码段创建的初始 SSA：

```go
func usefor() {
  i := 4
  for !(i == 0) {
    i--
    sayhi()
  }
}

func sayhi() {
  fmt.Println("Hello, for!")
}
```

我将移除打印输出函数的原因是为了使输出的 SSA 更简洁。使用-l 进行编译以禁用内联，这将导致对 sayhi()的微小调用（由于常量字符串而生成更多的代码，对 fmt.Println()[[2](#jump2)]的调用会生成更多代码）。

产生的 SSA 为：

```
b1:

        v1 (?) = InitMem <mem>
        v2 (?) = SP <uintptr>
        v3 (?) = SB <uintptr>
        v4 (?) = Const64 <int> [4] (i[int])
        v6 (?) = Const64 <int> [0]
        v9 (?) = Const64 <int> [1]
    Plain → b2 (10)

    b2: ← b1 b4

        v5 (10) = Phi <int> v4 v10 (i[int])
        v14 (14) = Phi <mem> v1 v12
        v7 (10) = Eq64 <bool> v5 v6
    If v7 → b5 b3 (unlikely) (10)

    b3: ← b2

        v8 (11) = Copy <int> v5 (i[int])
        v10 (11) = Sub64 <int> v8 v9 (i[int])
        v11 (12) = Copy <mem> v14
        v12 (12) = StaticCall <mem> {"".sayhi} v11
    Plain → b4 (12)

    b4: ← b3
    Plain → b2 (10)

    b5: ← b2

        v13 (14) = Copy <mem> v14
    Ret v13
```

这里要注意的有趣部分是：

- bN 是控制流图的基本块。
- Phi 节点是显式的。最有趣的是对 v5 的分配。这恰恰是分配给 i 的选择器；一条路径来自 V4（初始化），从另一个 v10（在 i--）内循环中。
- 出于本练习的目的，请忽略带有 <mem> 的节点。Go 有一种有趣的方式来显式地在其 IR 中传播内存状态，在这篇文章中我们不讨论它。如果感兴趣，请参阅前面提到的 README 以了解更多详细信息。

顺便说一句，这里的 for 循环正是我们想要将 until 语句转换成的形式。

## 将 until AST 节点转换为 SSA

与往常一样，我们的代码将以 for 语句的处理为模型。首先，让我们从控制流程图开始应该如何寻找 until 语句：

![until cfg](https://raw.githubusercontent.com/studygolang/gctt-images/master/compiler-internal/until-cfg.png)

现在我们只需要在代码中构建这个 CFG。提醒：我们在[第 1 部分](https://eli.thegreenplace.net/2019/go-compiler-internals-adding-a-new-statement-to-go-part-1/)中添加的新 AST 节点类型为 OUNTIL。我们将在 gc/ssa.go 中的[state.stmt](https://github.com/golang/go/blob/master/src/cmd/compile/internal/gc/ssa.go#L1024)方法中添加一个新的分支语句，以将具有 OUNTIL 操作的 AST 节点转换为 SSA。case 块和注释的命名应使代码易于阅读，并与上面显示的 CFG 相关。

```go
case OUNTIL:
  // OUNTIL: until Ninit; Left { Nbody }
  // cond (Left); body (Nbody)
  bCond := s.f.NewBlock(ssa.BlockPlain)
  bBody := s.f.NewBlock(ssa.BlockPlain)
  bEnd := s.f.NewBlock(ssa.BlockPlain)

  bBody.Pos = n.Pos

  // first, entry jump to the condition
  b := s.endBlock()
  b.AddEdgeTo(bCond)
  // generate code to test condition
  s.startBlock(bCond)
  if n.Left != nil {
    s.condBranch(n.Left, bEnd, bBody, 1)
  } else {
    b := s.endBlock()
    b.Kind = ssa.BlockPlain
    b.AddEdgeTo(bBody)
  }

  // set up for continue/break in body
  prevContinue := s.continueTo
  prevBreak := s.breakTo
  s.continueTo = bCond
  s.breakTo = bEnd
  lab := s.labeledNodes[n]
  if lab != nil {
    // labeled until loop
    lab.continueTarget = bCond
    lab.breakTarget = bEnd
  }

  // generate body
  s.startBlock(bBody)
  s.stmtList(n.Nbody)

  // tear down continue/break
  s.continueTo = prevContinue
  s.breakTo = prevBreak
  if lab != nil {
    lab.continueTarget = nil
    lab.breakTarget = nil
  }

  // done with body, goto cond
  if b := s.endBlock(); b != nil {
    b.AddEdgeTo(bCond)
  }

  s.startBlock(bEnd)
```

如果您想知道 n.Ninit 的处理位置——它在 switch 之前针对所有节点类型统一完成。

实际上，这是我们要做的全部工作，直到在编译器的最后阶段执行语句为止！如果我们运行编译器-像以前一样在此代码上转储 SSA：

```go
func useuntil() {
  i := 4
  until i == 0 {
    i--
    sayhi()
  }
}

func sayhi() {
  fmt.Println("Hello, for!")
}
```

正如预期的那样，我们将获得 SSA，该 SSA 在结构上等效于条件为否的 for 循环的 SSA 。

## 转换 SSA

构造初始 SSA 之后，编译器会在 SSA IR 上执行以下较长的遍历过程：

1. 执行优化
2. 将其降低到更接近机器代码的形式

所有这些都可以在在 ssa/compile.go 中的[passes](https://github.com/golang/go/blob/master/src/cmd/compile/internal/ssa/compile.go#L413)切片以及它们运行顺序的一些限制[passOrder](https://github.com/golang/go/blob/master/src/cmd/compile/internal/ssa/compile.go#L475)切片中找到。这些优化对于现代编译器来说是相当标准的。降低由我们正在编译的特定体系结构的指令选择以及寄存器分配。

有关这些遍的更多详细信息，请参见[SSA README](https://github.com/golang/go/blob/master/src/cmd/compile/internal/ssa/README.md)和[这篇帖子](https://quasilyte.dev/blog/post/go_ssa_rules/)，其中详细介绍了如何指定 SSA 优化规则。

## 生成机器码

最后，编译器调用 genssa 函数([gc/ssa.go](https://github.com/golang/go/blob/master/src/cmd/compile/internal/gc/ssa.go#L5903))从 SSA IR 发出机器代码。我们不必修改任何代码，因为 until 语句包含在编译器其他地方使用的构造块，我们才为之发出的 SSA-我们不添加新的指令类型，等等。

但是，研究的 useuntil 函数生成的机器代码对我们是有指导意义的。Go 有[自己的具有历史根源的汇编语法](https://golang.org/doc/asm)。我不会在这里讨论所有细节，但是以下是带注释的（带有＃注释）程序集转储，应该相当容易。我删除了一些垃圾回收器的指令（PCDATA 和 FUNCDATA）以使输出变小。

```
"".useuntil STEXT size=76 args=0x0 locals=0x10
  0x0000 00000 (useuntil.go:5)  TEXT  "".useuntil(SB), ABIInternal, $16-0

  # Function prologue

  0x0000 00000 (useuntil.go:5)  MOVQ  (TLS), CX
  0x0009 00009 (useuntil.go:5)  CMPQ  SP, 16(CX)
  0x000d 00013 (useuntil.go:5)  JLS  69
  0x000f 00015 (useuntil.go:5)  SUBQ  $16, SP
  0x0013 00019 (useuntil.go:5)  MOVQ  BP, 8(SP)
  0x0018 00024 (useuntil.go:5)  LEAQ  8(SP), BP

  # AX will be used to hold 'i', the loop counter; it's initialized
  # with the constant 4. Then, unconditional jump to the 'cond' block.

  0x001d 00029 (useuntil.go:5)  MOVL  $4, AX
  0x0022 00034 (useuntil.go:7)  JMP  62

  # The end block is here, it executes the function epilogue and returns.

  0x0024 00036 (<unknown line number>)  MOVQ  8(SP), BP
  0x0029 00041 (<unknown line number>)  ADDQ  $16, SP
  0x002d 00045 (<unknown line number>)  RET

  # This is the loop body. AX is saved on the stack, so as to
  # avoid being clobbered by "sayhi" (this is the caller-saved
  # calling convention). Then "sayhi" is called.

  0x002e 00046 (useuntil.go:7)  MOVQ  AX, "".i(SP)
  0x0032 00050 (useuntil.go:9)  CALL  "".sayhi(SB)

  # Restore AX (i) from the stack and decrement it.

  0x0037 00055 (useuntil.go:8)  MOVQ  "".i(SP), AX
  0x003b 00059 (useuntil.go:8)  DECQ  AX

  # The cond block is here. AX == 0 is tested, and if it's true, jump to
  # the end block. Otherwise, it jumps to the loop body.

  0x003e 00062 (useuntil.go:7)  TESTQ  AX, AX
  0x0041 00065 (useuntil.go:7)  JEQ  36
  0x0043 00067 (useuntil.go:7)  JMP  46
  0x0045 00069 (useuntil.go:7)  NOP
  0x0045 00069 (useuntil.go:5)  CALL  runtime.morestack_noctxt(SB)
  0x004a 00074 (useuntil.go:5)  JMP  0
```

如果您注意的话，您可能已经注意到“cond”块移到了函数的末尾，而不是最初在 SSA 表示中的位置。是什么赋予的？

答案是，“loop rotate”遍历将在 SSA 的最末端运行。此遍历对块重新排序，以使主体直接流入 cond，从而避免每次迭代产生额外的跳跃。如果您有兴趣，请参阅[ssa/looprotate.go](https://github.com/golang/go/blob/master/src/cmd/compile/internal/ssa/looprotate.go)了解更多详细信息。

## 结论

就是这样！在这两篇文章中，我们以两种不同的方式实现了一条新语句，从而知道了 Go 编译器的内部结构。当然，这只是冰山一角，但我希望它为您自己开始探索提供了一个良好的起点。

最后一点：我们在这里构建了一个可运行的编译器，但是 Go 工具都无法识别新的 until 关键字。不幸的是，此时 Go 工具使用了完全不同的路径来解析 Go 代码，并且没有与 Go 编译器本身共享此代码。我将在以后的文章中详细介绍如何使用工具处理 Go 代码。

## 附录-复制这些结果

要重现我们到此为止的 Go 工具链的版本，您可以从第 1 部分开始 ，还原 walk.go 中的 AST 转换代码，然后添加上述的 AST 到 SSA 转换。或者，您也可以从[我的 fork 中](https://github.com/eliben/go/tree/adduntil2)获取[adduntil2 分支](https://github.com/eliben/go/tree/adduntil2)。

要获得所有 SSA 的 SSA，并在单个方便的 HTML 文件中传递代码生成，请在构建工具链后运行以下命令：

```
GOSSAFUNC=useuntil <src checkout>/bin/go tool compile -l useuntil.go
```

然后在浏览器中打开 ssa.html。如果您还想查看 CFG 的某些通行证，请在函数名后添加通行名，以：分隔。例如 GOSSAFUNC = useuntil：number_lines。

要获取汇编代码码，请运行：

```
<src checkout>/bin/go tool compile -l -S useuntil.go
```

<span id="jump1">[1]</span> 我特别尝试避免在这些帖子中过多地讲“前端”和“后端”。这些术语是重载和不精确的，但通常前端是在构造 AST 之前发生的所有事情，而后端是在表示形式上更接近于机器而不是原始语言的阶段。当然，这在中间位置留有很多地方，并且 中间端也被广泛使用（尽管毫无意义）来描述中间发生的一切。

在大型和复杂的编译器中，您会听到有关“前端的后端”和“后端的前端”以及类似的带有“中间”的混搭的信息。

在 Go 中，情况不是很糟糕，并且边界已明确明确地确定。AST 在语法上接近输入语言，而 SSA 在语法上接近。从 AST 到 SSA 的转换非常适合作为 Go 编译器的前/后拆分。

<span id="jump2">[2] -S 告诉编译器将程序集源代码转储到 stdout； -l 禁用内联，这会通过内联 fmt.Println 的调用而使主循环有些模糊。

---
via: https://eli.thegreenplace.net/2019/go-compiler-internals-adding-a-new-statement-to-go-part-2/

作者：[Eli Bendersky](https://eli.thegreenplace.net/)
译者：[keob](https://github.com/keob)
校对：[@unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
