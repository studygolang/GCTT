首发于：https://studygolang.com/articles/28456

# Go：边界检查确保内存安全

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200304-Go-Memory-Safety-with-Bounds-Check/00.png)

ℹ️*这篇文章基于 Go 1.13 编写。*

Go 的一系列内存管理手段（内存分配，垃圾回收，内存访问检查）使许多开发者的开发工作变得很轻松。编译器通过在代码中引入“边界检查” 来确保安全地访问内存。

## 生成的指令

Go 引入了一些控制点位，来确保我们的程序访问的内存片段安全且有效的。让我们从一个简单的例子开始：

```go
package main

func main() {
    list := []int{1, 2, 3}

    printList(list)
}

func printList(list []int) {
    println(list[2])
    println(list[3])
}
```

这段代码跑起来之后会 panic：

```
3
panic: runtime error: index out of range [3] with length 3
```

Go 通过添加边界检查来防止不正确的内存访问

*如果你想知道没有这些检查会怎么样，你可以使用 `-gcflags="-B"` 的选项，输出如下*

```
3
824633993168
```

*因为这块内存是无效的，它会读取不属于这个 slice 的下一个 bytes。*

利用命令 `go tool compile -S main.go` 来生成对应的[汇编](https://golang.org/doc/asm)代码，就可以看到这些检查点：

```
0x0021 00033 (main.go:10)  MOVQ   "".list+48(SP), CX
0x0026 00038 (main.go:10)  CMPQ   CX, $2
0x002a 00042 (main.go:10)  JLS    161
[...] here Go prints the third element
0x0057 00087 (main.go:11)  MOVQ   "".list+48(SP), CX
0x005c 00092 (main.go:11)  CMPQ   CX, $3
0x0060 00096 (main.go:11)  JLS    151
[...]
0x0096 00150 (main.go:12)  RET
0x0097 00151 (main.go:11)  MOVL   $3, AX
0x009c 00156 (main.go:11)  CALL   runtime.panicIndex(SB)
0x00a1 00161 (main.go:10)  MOVL   $2, AX
0x00a6 00166 (main.go:10)  CALL   runtime.panicIndex(SB)
```

Go 先使用 `MOVQ` 指令将 list 变量的长度放入寄存器 `CX` 中

```
0x0021 00033 (main.go:10)  MOVQ   "".list+48(SP), CX
```

*友情提醒，slice 类型的变量由三部分组成，指向底层数组的指针、长度，容量(capacity)。list 变量在栈中的位置如下图：*

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200304-Go-Memory-Safety-with-Bounds-Check/00.png)

*通过将栈指针移动 48 个字节就可以访问长度*

下一条指令将 slice 的长度与程序即将访问的偏移量进行比较

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200304-Go-Memory-Safety-with-Bounds-Check/01.png)

`CMPQ` 指令会将两个值相减，并在下一条指令中与 0 进行比较。如果 slice 的长度（寄存器 `CX`）减去要访问的偏移量（在这个例子当中是 2）小于或等于 0（`JLS` 是 *Jump on lower or the same* 的缩写），程序就会跳到 `161` 处继续执行。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200304-Go-Memory-Safety-with-Bounds-Check/02.png)

两种边界检查使用的都是相同的指令。除了看生成的汇编代码，Go 提供了一个编译期的通行证去打印出边界检查的点，你可以在 `build` 和 `run` 的时候使用标志 `-gcflags="-d=ssa/check_bce/debug=1"` 去开启。输出如下：

```
./main.go:10:14: Found IsInBounds
./main.go:11:14: Found IsInBounds
```

我们可以看到输出里生成了两个检查点。不过 Go 编译器足够聪明，在不需要的情况下，它不会生成边界检查的指令。

## 规则

在每次访问内存的时候都生成检查指令是非常低效的，让我们稍微修改一下前面的例子。

```go
package main

func main() {
    list := []int{1, 2, 3}

    printList(list)
}

func printList(list []int) {
    println(list[3])
    println(list[2])
}
```

两个 `println` 指令对调了，用 `check_bce` 标志再去跑一遍程序，这次只有一处边界检查：

```
./main.go:11:14: Found IsInBounds
```

程序先检查了偏移量 `3` 。如果是有效的，那么 `2` 很明显也是有效的，没必要再去检查了。可以通过命令 `GOSSAFUNC=printList Go run main.go` 来生成 SSA 代码看编译过程。这张图就是生成的带边界检查的 SSA 代码：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200304-Go-Memory-Safety-with-Bounds-Check/03.png)

里面的 `prove` pass 将边界检查标记为移除，这样后面的 pass 将会收集这些 dead code：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200304-Go-Memory-Safety-with-Bounds-Check/04.png)

用这条命令 `GOSSAFUNC=printList Go run -gcflags="-d=ssa/prove/debug=3" main.go` 可以把 pass 背后的逻辑打印出来，它也会生成 SSA 文件来帮助你 debug，接下来看命令的输出：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200304-Go-Memory-Safety-with-Bounds-Check/05.png)

这个 pass 实际上会采取不同的策略，并建立了 fact 表。 这些 fact 决定了矛盾点在哪里。 在我们这个例子里，我们可以通过 SSA 的 pass 来解读这些规则：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200304-Go-Memory-Safety-with-Bounds-Check/06.png)

第一个阶段从代表指令 `println(list[3])` 的分析块 `b1` 开始，这个指令有两种可能：

- 偏移量 `[3]` 在边界中，跳到第二个指令 b2。在这个例子中，Go 指定 v7 的限制（slice 的长度）是 `[4, max(int)]`。
- 偏移量 `[3` 不在边界中, 程序跳转到 b3 指令并 panic。

接下来，Go 开始处理 `b2` 块（第二个指令）。这里也有两种可能

- 偏移量 `[2]` 在边界中，这意味着 slice 的长度 `v7` 比 `v23`（偏移量 `[2]`） 要大。在先前的 b1 块中 Go 已经判断了 `v7 > 4`, 所以这个已经被确认了。
- 偏移量 [2] 不在边界中，这意味着它比 slice 的长度 `v7` 更大，但 `v7` 的限制是 `[4, max(int)]` ，所以 Go 会将这个分之标记为矛盾，意味着这种情况永远不会发生，这条指令的边界检查可以被移除。

这个 pass 在随着时间不断地改善，现在可以参考[更多的 case](https://github.com/golang/go/blob/master/test/prove.go)。消除边界检查可以略微提升 Go 程序的运行速度，但除非你的程序是微妙级敏感的，不然没有必要去优化它。

---

via: <https://medium.com/a-journey-with-go/go-memory-safety-with-bounds-check-1397bef748b5>

作者: [Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[yxlimo](https://github.com/yxlimo)
校对：[Alex.Jiang](https://github.com/JYSDeveloper)
本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
