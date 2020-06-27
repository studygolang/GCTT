首发于：https://studygolang.com/articles/28455

# Go 中的循环是如何转为汇编的？

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200201-how-are-loops/cover.png)

*本文基于 Go 1.13 版本*

循环在编程中是一个重要的概念，且易于上手。但是，循环必须被翻译成计算机能理解的底层指令。它的编译方式也会在一定程度上影响到标准库中的其他组件。让我们开始分析循环吧。

## 循环的汇编代码

使用循坏迭代 `array`，`slice`，`channel`，以下是一个使用循环对 `slice` 计算总和的例子。

```go
func main() {
   l := []int{9, 45, 23, 67, 78}
   t := 0

   for _, v := range l {
      t += v
   }

   println(t)
}
```

使用 `go tool compile -S main.go` 生成的汇编代码，以下为相关输出：

```go
0x0041 00065 (main.go:4)   XORL   AX, AX
0x0043 00067 (main.go:4)   XORL   CX, CX

0x0045 00069 (main.go:7)   JMP    82
0x0047 00071 (main.go:7)   MOVQ   ""..autotmp_5+16(SP)(AX*8), DX
0x004c 00076 (main.go:7)   INCQ   AX
0x004f 00079 (main.go:8)   ADDQ   DX, CX
0x0052 00082 (main.go:7)   CMPQ   AX, $5
0x0056 00086 (main.go:7)   JLT    71
0x0058 00088 (main.go:11)  MOVQ   CX, "".t+8(SP)
```

我把这些指令分为了两个部分，初始化部分和循环主体。前两条指令，将两个寄存器初始化为零值。

```go
0x0041 00065 (main.go:4)   XORL   AX, AX
0x0043 00067 (main.go:4)   XORL   CX, CX
```

寄存器 `AX` 包含着当前循环所处位置，而 `CX` 包含着变量 `t` 的值，下面为带有指令和通用寄存器的直观表示：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200201-how-are-loops/1.png)

循环从表示「跳转到指令 82 」的 `JMP 82` 开始，这条指令的作用可以通过第二行来判断：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200201-how-are-loops/2.png)

接下来的指令 `CMPQ AX,$5` 表示「比较寄存器 `AX` 和 `5`」，事实上，这个操作是把 `AX` 中的值减去 5 ，然后储存在另一个寄存器中，这个值可以被用在下一条指令 `JLT 71` 中，它的含义是 「如果值小于 0 则跳转到指令 71 」，以下是更新后的直观表示：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200201-how-are-loops/3.png)

如果不满足条件，则程序将会跳转到循环体之后的下一条指令执行。

所以，我们现在有了对循环的基本框架，以下是转换后的 Go 循环：

```go
goto end
start:
   ?
end:
   if i < 5 {
      goto start
   }

println(t)
```

我们缺少了循环的主体，接下来，我们看看这部分的指令：

```go
0x0047 00071 (main.go:7)   MOVQ   ""..autotmp_5+16(SP)(AX*8), DX
0x004c 00076 (main.go:7)   INCQ   AX
0x004f 00079 (main.go:8)   ADDQ   DX, CX
```

第一条指令 `MOVQ ""..autotmp_5+16(SP)(AX*8), DX`  表示 「将内存从源位置移动到目标地址」，它由以下几个部分组成：

* `""..autotmp_5+16(SP)` 表示 `slice` ，而 `SP` 表示了栈指针即我们当前的内存空间， `autotmp_*` 是自动生成变量名。
* 偏差为 8 是因为在 64 位计算机架构中，`int` 类型是 8 字节的。偏差乘以寄存器 `AX` 的值，表示当前循环中的位置。
* 寄存器 `DX` 代表的目标地址内包含着循环的当前值。

之后，`INCQ` 表示自增，然后会增加循环的当前位置：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200201-how-are-loops/4.png)
循环主体的最后一条指令是 `ADDQ DX, CX` ,表示把 `DX` 的值加在 `CX`，所以我们可以看出，`DX` 所包含的值是目前循环所代表的的值，而 `CX` 代表了变量 `t` 的值。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200201-how-are-loops/5.png)

他会一直循环至计数器到 5 ，之后循环体之后的指令表示为将寄存器 `CX` 的值赋予 `t` ：

```GO
0x0058 00088 (main.go:11)   MOVQ   CX, "".t+8(SP)
```

以下为最终状态的示意图：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200201-how-are-loops/6.png)

我们可以完善 Go 中循环的转换：

```go
func main() {
   l := []int{9, 45, 23, 67, 78}
   t := 0
   i := 0

   var tmp int

   goto end
start:
   tmp = l[i]
   i++
   t += tmp
end:
   if i < 5 {
      goto start
   }

   println(t)
}
```

这个程序生成的汇编代码与上文所提到的函数生成的汇编代码有着相同的输出。

## 改进

循环的内部转换方式可能会对其他特性(如 Go 调度器)产生影响。在 Go 1.10 之前，循环像下面的代码一样编译：

```go
func main() {
   l := []int{9, 45, 23, 67, 78}
   t := 0
   i := 0

   var tmp int
   p := uintptr(unsafe.Pointer(&l[0]))

   if i >= 5 {
      goto end
   }
body:
   tmp = *(*int)(unsafe.Pointer(p))
   p += unsafe.Sizeof(l[0])
   i++
   t += tmp
   if i < 5 {
      goto body
   }
end:
   println(t)
}
```

这种实现方式的问题是，当 `i` 达到 5 时，指针 `p` 已经超过了内存分配空间的尾部。这个问题使得循环不容易抢占，因为它的主体是不安全的。循环编译的优化确保它不会创建任何越界的指针。这个改进是为 Go 调度器中的非合作抢占做准备的。你可以在这篇 [Proposal](https://github.com/golang/proposal/blob/master/design/24543-non-cooperative-preemption.md) 中到更详细的讨论。

---

via: https://medium.com/a-journey-with-go/go-how-are-loops-translated-to-assembly-835b985309b3

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[Jun10ng](https://github.com/Jun10ng)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
