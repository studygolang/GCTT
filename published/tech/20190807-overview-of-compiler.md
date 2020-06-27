首发于：https://studygolang.com/articles/24554

# 编译器概述

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/go-compiler.png "'Golang 之旅'插图，由 Go Gopher 的 Renee French 创作")

> *本文基于 Go 1.13*

Go 编译器是 Go 生态系统中的一个重要工具，因为它是将程序构建为可执行二进制文件的基本步骤之一。编译器的历程是漫长的，它先用 C 语言编写，迁移到 Go，许多优化和清理将在未来继续发生，让我们来看看它的高级操作。

## 阶段（phases）

Go 编译器由四个阶段组成，可以分为两类：

* 前端（frontend）：这个阶段从源代码进行分析，并生成一个抽象的源代码语法结构，称为 [AST](https://en.wikipedia.org/wiki/Abstract_syntax_tree)
* 后端（backend）：第二阶段将把源代码的表示转换为机器码，并进行一些优化。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/layer.png)

[编译器文档](https://github.com/golang/go/blob/release-branch.go1.13/src/cmd/compile/README.md)

为了更好理解每个阶段，我们看个简单的程序：

```go
package main

func main() {
    a := 1
    b := 2
    if true {
        add(a, b)
    }
}

func add(a, b int) {
    println(a + b)
}
```

## 解析

第一阶段非常简单，在 [文档](https://github.com/golang/go/blob/release-branch.go1.13/src/cmd/compile/README.md) 中有很好的解释：

> 在编译的第一阶段，对源代码进行标记（词法分析）、解析（语法分析），并为每个源文件构建语法树。

lexer 是第一个运行用来标记源代码的包。下面是上边例子的 [标记化](https://gist.github.com/blanchonvincent/1f1cb850a436ffbb81df14eb586f52df) 输出：

![Go 源码标记化](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/Go%20source%20code%20tokenized.png)

一旦被标记化，代码将被解析、构建代码树。

## AST（抽象语法树） 转换

可以通过 `go tool compile` 命令和标志 `-w` 展示 [抽象语法树](https://en.wikipedia.org/wiki/Abstract_syntax_tree) 的转换：

![构建 AST 的简单过程](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/sample%20of%20the%20generated%20AST.png)

此阶段还将包括内联等优化。在我们的示例中，由于我们没有看到 `CALLFUNC` 该方法的任何 `add` 指令，该方法 `add` 已经内联。让我们使用禁用内联的标志 `-l` 再次运行。

![构建 AST 的简单过程](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/sample%20of%20the%20generated%20AST%202.png)

AST 生成后，它允许编译器使用 SSA 表示转到较低级别的中间表示。

## SSA（静态单赋值）的生成

[静态单赋值](https://en.wikipedia.org/wiki/Static_single_assignment_form) 阶段进行优化：消除死代码，删除不使用的分支，替换一些常量表达式等等。

使用 `GOSSAFUNC=main Go tool compile main.go && open ssa.html` 命令，生成 HTML 文档的命令将在 SSA 包中完成所有不同的过程，因此可以转储 SSA 代码：

![SSA 过程](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/SSA%20passes.png)

生成的 SSA 位于 “start” 选项卡中：

![SSA 代码](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/SSA%20code.png)

在这里，高亮显示变量 `a` 和 `b` 以及 `if` 条件表达式，将向我们展示这些行是怎么变化的。这些代码也向我们描述了编译器如何管理 `println` 函数，该函数被分解为 4 个步骤：printlock、printint、printnl、printunlock。编译器会自动为我们添加一个锁，并根据参数的类型，调用相关的方法来正确输出。

在我们的示例中，由于编译时已知 `a` 和 `b`，所以编译器可以计算最终结果并将变量标记为不必要的。通过 `opt` 优化这部分：

![SSA code — “opt” 过程](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/SSA%20code%20%E2%80%94%20%E2%80%9Copt%E2%80%9D%20pass.png)

在这里，`v11` 已经被添加的 `v4` 和 `v5` 所替代，这两个 `v4` 和 `v5` 被标记为死代码。然后通过 `opt deadcode` 将删除这些代码。

![SSA code — “opt deadcode” 过程](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/SSA%20code%20%E2%80%94%20%E2%80%9Copt%20deadcode%E2%80%9D%20pass.png)

对于 `if` 条件，`opt` 阶段将常量 `true` 标记为死代码，然后删除：
![删除布尔常量](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/constant%20boolean%20is%20removed.png)

然后，通过将不必要的块和条件标记为无效，另一次传递将简化控制流。这些块稍后将被另一个专用于死代码的阶段删除

![删除不必要控制流](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/unnecessary%20control%20flow%20is%20removed.png)

完成所有过程之后，Go 编译器现在将生成一个中间汇编代码

![Go 汇编码](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/Go%20asm%20code.png)

下一阶段将把机器码生成到二进制文件中。

## 生成机器码

编译器的最后一步是生成目标(object)文件，在我们的例子中生成 `main.c`。从这个文件中，现在可以使用 `objdumptool` 对其进行反编译。下面是一个很好的图表,由 Grant Seltzer Richman 创建:

![compile 工具](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/go%20tool%20compile.png)

![objdump 工具](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-overview-of-compile/go%20tool%20objdump.png)

*您可以在“[Dissecting Go Binaries](https://www.grant.pizza/dissecting-go-binaries/)”中找到有关对象文件和二进制文件的更多信息。*

生成目标文件后，现在可以使用 `go tool link` 将其直接传递给链接器，二进制文件将最终就绪。

---

via: https://medium.com/a-journey-with-go/go-overview-of-the-compiler-4e5a153ca889

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[TomatoAres](https://github.com/TomatoAres)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
