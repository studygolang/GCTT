首发于：https://studygolang.com/articles/16138

# Go 语言中的递归和尾调用操作

曾几何时，我看过一段关于 Go 递归函数的简单例子，作者用了极快的速度简单的陈述了 Go 这门语言中并没有优化递归这一操作，即使是在尾调用（tail calls）非常明显的时间。我当时并不理解什么是尾调用（tail calls），我非常想知道这位作者提到的 Go 语言没有优化递归操作的原因是什么，因为我从来不知道原来递归操作还可以被优化。

有些人不太理解什么是递归操作，其实可以用一种简单的说法来解释，就是函数自己调用自己本身。 那为什么我们会需要写一个函数来调用自己本身呢？举个例子，递归算法在使用栈（stack，先进后出）来执行数据操作的时候是非常方便的，它比循环操作更快，并且代码的简洁性好。

执行数学操作时，一个典型的递归场景是：当前的计算结果是下一步计算的输入。对于所有的递归操作，你必须设置一个锚点 (anchor) 触发递归函数的结束，并返回结果。如果不这样操作，那么递归函数将会进入无穷无尽的循环当中，最终引发内存溢出。

为什么你的程序会引发内存溢出？ 在传统的 C 语言程序中，栈内存是用来处理所有函数调用的地方，这是因为栈是一种预分配的内存，并且速度非常快。例如下图所示：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/recursion-and-tail-call/1.png)

上面这个图描述了一个典型的栈内存例子，并且它也许和我们所写的所有程序类似。你可以从图中看到，栈内存会随着函数调用而增加，每一次我们从一个函数中调用另一个函数、变量、寄存器和数据的时候，总是会把这些参数地址或者变量名压入栈内，从而使得它内存增加。

在 C 程序中，每个线程自身都有一定的栈内存空间，根据结构的不同，栈内存的大小当然也不一样，大概从 1M 到 8M 不等。当然，你也可以调节默认值小大。如果你写的程序会产生大量的线程，那么你将会迅速的用完那些你不可能会用掉的内存。

在 Go 程序中，每个 Goroutine 都会分配栈内存，但是 Go 更加智能一些：刚开始会分配 4K，之后按需增长。Go 语言能够动态的增加栈内存的概念来自于 split stacks( 栈分割 )。关于更多 split stacks 的内容请移步：

http://gcc.gnu.org/wiki/SplitStacks

你同时还可以参考下面 Go 语言中 runtime 包里的栈实现代码：

- http://golang.org/src/pkg/runtime/stack.h
- http://golang.org/src/pkg/runtime/stack.c

当你要使用递归的时候，你需要意识到栈内存是一直增加的，直到遇到你设置好的 anchor 时，它的内存才开始下降。当我们说 Go 并没有优化递归操作时，我们需要承认一个事实，Go 并没有尝试着去优化栈内存无限增加这一操作。这个时候 tail calls 就登场了。

在介绍 tail calls 是如何优化递归函数之前，让我们来看一个简单的递归函数：

```go
func Recursive(number int) int {
    if number == 1 {
        return number
    }
    return number + Recursive(number-1)
}

func main() {
    answer := Recursive(4)
    fmt.Printf("Recursive: %d\n", answer)
}
```

上面的这个递归函数需要传入一个整数作为参数，并 return 回一个整数。如果传入函数值是 1，那么函数将会立马返回。这个 if 函数包含了锚点，并采用栈来完成执行任务。

如果传入的变量并不是 1，那么递归函数将开始工作。递归函数将参数值减去 1，然后使用减掉后的参数作为下一次递归调用的参数。栈内存随着每一次的函数调用增加，直到达到锚点，所有的递归调用开始返回一直到主函数。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/recursion-and-tail-call/2.png)

让我们看看上图中所有的函数调用和返回值是怎么操作的：

1. 该流程图，从左下角开始，至下往上的递归操作，该流程展现了详细的调用链。
2. 主函数中调用递归函数传入参数 4，接着递归函数调用自身传入参数 3。这样反复调用，直到参数值 1 传入递归函数。
3. 递归函数在 return 前调用自己三次，当到达锚点的时候有 3 个扩展栈桢（extended stack frames）对应每一个调用操作。
4. 接着，递归开始展开，真正的工作开始进行。从图中可知，右边从上到下就是展开操作。
5. 每一个 return 操作通过获取参数并将其添加到函数调用的返回值来执行每个返回操作。
6. 最后的 return 操作执行后我们得到了最终的答案 10。

递归函数执行这些步骤非常迅速，因此这就是递归的好处，我们不需要任何迭代或者计数进行循环。栈将当前的结果存储起来返回给之前的函数。当然，我们唯一需要的担心的是我们需要消耗多少内存。

什么是 tail calls ？它是如何优化递归函数的？ 接下来将创建一个带有 tail calls 的递归函数，了解它是如何解决递归函数消耗大量栈内存问题的。

下面是同样的递归函数，但它是通过 tail call 实现的：

```go
func TailRecursive(number int, product int) int {
    product = product + number
    if number == 1 {
        return product
    }
    return TailRecursive(number-1, product)
}

func main() {
    answer := TailRecursive(4, 0)
    fmt.Printf("Recursive: %d\n", answer)
}
```

你能发现他们之间的不同吗？它与我们如何使用堆栈和计算结果有关。在这个实现中，锚点竟然返回了最终结果！除了最终值，我们不再需要任何栈的返回值。

一些编译器能够看到这种细微差别，并更改生成的底层程序集，以便对所有递归调用使用一个栈框架。Go 编译器还无法检测到这种细微差别。为了证明让我们来看看 Go 编译器为这两个函数生成的汇编代码。

为了产生汇编代码，读书可在终端执行下面的命令：

```
go tool 6g -S ./main.go > assembly.asm
```

根据你的机器架构，一般有三种编译器。

- 6g: AMD64 架构： 无论处理器是由英特尔还是 AMD 构建，这都适用于现代 64 位处理器。AMD 开发了 x86 架构的 64 位扩展。
- 8g: x86 架构：基于 8086 架构，适用于 32 位处理器。
- 5g: ARM 架构：适用于基于 RISC 的处理器，代表精简指令集计算。

需要学习更多关于架构的知识或者其他 Go 工具命令的可移步：http://golang.org/cmd/gc/

我下面同时列出了 Go 代码和汇编代码。希望其中的一项能帮助到你理解。

为了让处理器能够操作数据，类似于加法运算或者数值对比，数据必须存在寄存器中。你可以想一想寄存器是如何处理变量的。
当你在看下面汇编代码的时候，你会了解到 AX 和 BX 是主要的寄存器，且它们经常被使用到。SP 寄存器是栈指针，FP 寄存器是帧指针，它也与栈有关。

那么，让我看看下面的代码：

```go
07 func Recursive(number int) int {
08
09     if number == 1 {
10
11         return number
12     }
13
14     return number + Recursive(number-1)
15 }
```

```go
— prog list "Recursive" —
0000 (./main.go:7) TEXT Recursive+0(SB),$16-16

0001 (./main.go:7) MOVQ number+0(FP),AX

0002 (./main.go:7) LOCALS ,$0
0003 (./main.go:7) TYPE number+0(FP){int},$8
0004 (./main.go:7) TYPE ~anon1+8(FP){int},$8

0005 (./main.go:9) CMPQ AX,$1
0006 (./main.go:9) JNE ,9

0007 (./main.go:11) MOVQ AX,~anon1+8(FP)
0008 (./main.go:11) RET ,

0009 (./main.go:14) MOVQ AX,BX
0010 (./main.go:14) DECQ ,BX

0011 (./main.go:14) MOVQ BX,(SP)
0012 (./main.go:14) CALL ,Recursive+0(SB)

0013 (./main.go:14) MOVQ 8(SP),AX
0014 (./main.go:14) MOVQ number+0(FP),BX
0015 (./main.go:14) ADDQ AX,BX

0016 (./main.go:14) MOVQ BX,~anon1+8(FP)
0017 (./main.go:14) RET ,
```

如果我们跟随汇编代码，你会发现所有地方都有栈的身影：

- 0001: AX 寄存器中数值变量的值来自于栈
- 0005-0006: 数值变量和 1 进行对比，如果他们不相等，那么就 jump 到 Go 代码的第 14 行
- 0007-0008: 到达锚点的时候，函数执行 return 操作，并把参数变量拷贝到栈中
- 0009-0010: 将参数变量减去 1
- 0011-0012: 将参数变量 push 到栈中并开始执行递归函数
- 0013-0015: 函数返回。返回值从栈中 pop 出来，传入到 AX 寄存器中。同时，参数值从栈桢中传入到 BX 寄存器，将他们相加
- 0016-0017: 相加后的结果拷贝到栈上，函数返回

汇编代码显示的是我们正在进行的递归调用，并且值正按预期从栈中 push 和 pop。栈内存在增加，同时也一直在释放。

现在，让我们为包含 tail call 的递归函数生成汇编代码，看看 Go 编译器是否优化了什么。

```go
17 func TailRecursive(number int, product int) int {
18
19     product = product + number
20
21     if number == 1 {
22
23         return product
24     }
25
26     return TailRecursive(number-1, product)
27 }
```

```
— prog list "TailRecursive" —
0018 (./main.go:17) TEXT TailRecursive+0(SB),$24-24

0019 (./main.go:17) MOVQ number+0(FP),CX

0020 (./main.go:17) LOCALS ,$0
0021 (./main.go:17) TYPE number+0(FP){int},$8
0022 (./main.go:17) TYPE product+8(FP){int},$8
0023 (./main.go:17) TYPE ~anon2+16(FP){int},$8

0024 (./main.go:19) MOVQ product+8(FP),AX
0025 (./main.go:19) ADDQ CX,AX

0026 (./main.go:21) CMPQ CX,$1
0027 (./main.go:21) JNE ,30

0028 (./main.go:23) MOVQ AX,~anon2+16(FP)
0029 (./main.go:23) RET ,

0030 (./main.go:26) MOVQ CX,BX
0031 (./main.go:26) DECQ ,BX

0032 (./main.go:26) MOVQ BX,(SP)
0033 (./main.go:26) MOVQ AX,8(SP)
0034 (./main.go:26) CALL ,TailRecursive+0(SB)

0035 (./main.go:26) MOVQ 16(SP),BX

0036 (./main.go:26) MOVQ BX,~anon2+16(FP)
0037 (./main.go:26) RET ,
```
包含 tail calls 的函数产生了更多的汇编代码。但是，结果却非常相似。事实上，从性能的角度来看，我们已经让事情变得更糟了。

Go 没有针对我们实施的 tail calls 进行任何优化。 我们仍然进行了和之前例子一样相同的栈操作和递归调用。所以我猜 Go 目前没有针对递归进行优化。 这并不意味着我们不应该使用递归，我们只要知道 Go 语言中有这种特性即可。

如果你有一个问题通过递归算法可以完美的解决，但是又害怕浪费内存，那你可以考虑使用 channel 操作。不过，这项操作某种意义上来说会慢很多，但它确实能解决你的担忧。以下是使用通道实现递归函数的方法：

```go
func RecursiveChannel(number int, product int, result chan int) {
    product = product + number
    if number == 1 {
        result <- product
        return
    }
    Go RecursiveChannel(number-1, product, result)
}

func main() {
    result := make(chan int)
    RecursiveChannel(4, 0, result)
    answer := <-result
    fmt.Printf("Recursive: %d\n", answer)
}
```

它同样采用 tail calls 实现。一旦 anchor 被击中，最终答案就会产生，并将答案放入频道。它并不是进行递归调用，而是生成一个 Goroutine，它提供了我们在 tail calls 示例中推送到栈的相同状态。

唯一的区别是我们将一个无缓冲的 channel 传递给 Goroutine。只有锚点才能将数据写入通道 , 且它并不会产生其他 Goruntine。

在 main 函数中，创建了一个无缓冲的通道，并使用初始参数和通道调用 RecursiveChannel 函数。该函数立即返回，但 main 不会终止，这是因为它等待数据写入通道。一旦锚点被击中并将答案写入通道，main 函数唤醒结果并将其打印到屏幕上。在大多数情况下，main 将在 Goroutine 终止之前唤醒。

递归是编写 Go 程序时可以使用的另一种工具。目前，Go 编译器不会优化尾调用的代码，但是，Go 的未来版本没有什么理由不优化这个问题。如果程序运行时，内存大小问题是你考虑的一个因素，那么你通过 channel 去模拟递归的操作未尝不可。

---

via: https://www.ardanlabs.com/blog/2013/09/recursion-and-tail-calls-in-go_26.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[Yusen Wu](https://github.com/Yusen08)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
