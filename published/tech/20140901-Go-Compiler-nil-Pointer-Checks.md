首发于：https://studygolang.com/articles/16505

# Go 编译器 nil 指针检查

## 简介

我在思考编译器是如何保护我们写的代码。无效的内存访问检查是编译器添加到代码中的一种安全检查。我们可能会认为这种“额外的代码”会损耗程序的性能，甚至可能需要数十亿的迭代操作。但是，这些检查可以防止代码对正在运行的系统造成损害。编译器本质上是指出和查找错误，使我们编写的代码在运行时更安全。

基于上述考虑，同时 Go 语言想要达成更快的编译速度，如果硬件可以解决这些问题，那么 Go 语言编译器就会使用硬件来解决问题。其中一种情况是检测无效的内存访问。有时编译器会在代码中添加 nil 指针检查，而有时不会。在这篇博客中，我们将探讨一种情况，即编译器在什么情况下让硬件来检测无效的内存访问，以及在什么情况下会添加 nil 指针检查。

## 硬件只作检查

当编译器依赖于硬件来检查并指出无效的内存访问时，编译器可以生成更少的代码，以提高程序性能。如果我们的代码尝试读取或写入地址 0x0，则硬件将抛出一个异常，该异常将被 Go 运行时捕获并以 `panic` 的形式反馈给我们的程序。如果 `panic` 没有恢复，则产生堆栈跟踪信息。

如下是一个尝试对 0x0 内存地址进行写入数据示例，以及产生的 `panic` 和堆栈跟踪信息：

```go
01 package main
02
03 func main() {
04    var p *int  // 声明一个 nil 值指针
05    *p = 10     // 将 10 写入地址 0x0
06 }

panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xb code=0x1 addr=0x0 pc=0x2007]

goroutine 16 [running]:
runtime.panic(0x28600, 0x51744)
    /go/src/pkg/runtime/panic.c:279 +0xf5
main.main()
    /Go/Projects/src/github.com/goinaction/code/temp/main.go:5 +0x7

goroutine 17 [runnable]:
runtime.MHeap_Scavenger()
    /go/src/pkg/runtime/mheap.c:507
runtime.goexit()
    /go/src/pkg/runtime/proc.c:1445
```

让我们看看由 6g 编译器在 darwin/amd64 机器上的生成的汇编代码（译者注：6g 是 Golang 在 amd64 架构机器上的编译器，由于 Go 语言的版本关系，1.5 之后的版本中 6g/8g 已被取代，应该统一使用 `go tool compile -S main.go`）:

```asm
go tool 6g -S main.go

04  var p *int
0x0004 00004 (main.go:4)   MOVQ $0,AX

05  *p = 10
0x0007 00007 (main.go:5)   NOP ,
0x0007 00007 (main.go:5)   MOVQ $10,(AX)
```

在上面的汇编代码片段中，我们看到 0 的值被分配给 AX 寄存器，然后代码尝试将值 10 写入 AX 寄存器指向的存储器。这会产生 panic 以及相应的堆栈跟踪信息。上述汇编代码中没有显示 nil 指针检查，因为编译器已将其交给硬件进行检测和报告。

## 编译器检查

让我们看一下编译器生成 nil 指针检查的示例：

```go
01 package main
02
03 type S struct {
04     b [4096]byte
05     i int
06 }
07
08 func main() {
09     s := new(S)
10     s.i++
11 }
```

在第 3 行到第 6 行，我们定义了 S 的结构类型，这个类型有两个字段。第一个字段是 4096 字节的数组，第二个字段是一个整数。然后在第 9 行的 main 函数中，我们创建一个 S 类型的值，并将该值的地址分配给名为 s 的指针变量。最后在第 10 行，我们将 s 实例的 i 字段的值递增 1。

让我们看看由 6g 编译器在 darwin/amd64 机器上生成的第 9 行、第 10 行的汇编代码（译者注：同上译注）：

```asm
go tool 6g -S main.go

09  s := new(S)
0x0036 00054 (main.go:9)   LEAQ "".autotmp_0001+0(SP),DI
0x003a 00058 (main.go:9)   MOVL $0,AX
0x003c 00060 (main.go:9)   MOVQ $513,CX
0x0043 00067 (main.go:9)   REP ,
0x0044 00068 (main.go:9)   STOSQ ,
0x0046 00070 (main.go:9)   LEAQ "".autotmp_0001+0(SP),BX

10  s.i++
0x004a 00074 (main.go:10)  CMPQ BX,$0
0x004e 00078 (main.go:10)  JEQ $1,105
0x0050 00080 (main.go:10)  MOVQ 4096(BX),BP
0x0057 00087 (main.go:10)  NOP ,
0x0057 00087 (main.go:10)  INCQ ,BP
0x005a 00090 (main.go:10)  MOVQ BP,4096(BX)
0x0061 00097 (main.go:10)  NOP ,
```

在我们的示例中，第 10 行代码需要下面的指针运算才能使其工作 :

```asm
10  s.i++

0x004a 00074 (main.go:10)  CMPQ BX,$0
0x004e 00078 (main.go:10)  JEQ $1,105

0x0050 00080 (main.go:10)  MOVQ 4096(BX),BP
0x0057 00087 (main.go:10)  NOP ,
0x0057 00087 (main.go:10)  INCQ ,BP
0x005a 00090 (main.go:10)  MOVQ BP,4096(BX)
```

i 字段内存地址位于 S 类型的值内的 4096 字节偏移量。上面汇编代码中，BP 寄存器被分配到 BX (s 值 ) 的内存位置的值加上 4096 的偏移量（译者注：基地址加上偏移量）。之后 BP 的值增加 1，新值分配在 BX + 4096 的内存地址中。（译者注：0x0050 0008 至 0x005a 00090 代码片段）

在这个例子中，Go 编译器在自增代码之前添加一个 nil 指针检查：

```asm
10  s.i++

0x004a 00074 (main.go:10)  CMPQ BX,$0
0x004e 00078 (main.go:10)  JEQ $1,105

0x0050 00080 (main.go:10)  MOVQ 4096(BX),BP
0x0057 00087 (main.go:10)  NOP ,
0x0057 00087 (main.go:10)  INCQ ,BP
0x005a 00090 (main.go:10)  MOVQ BP,4096(BX)
```

上述高亮代码（译者注：0x004a 00074 至 0x004e 00078 代码片段，详见原文）代码显示了编译器添加的 nil 指针检查。将 BX 的值与 0 的值进行比较，如果它们相等，代码不会执行下面的自增代码片段。

问题是，为什么 Go 在这个例子中添加了一个 nil 指针检查，而第一个例子中却没有添加？

## 添加 nil 检查的情形

让我们对先前的例子做一点小修改：

```go
01 package main
02
03 type S struct {
04     i int          // 交换字段的顺序
05     b [4096]byte
06 }
07
08 func main() {
09     s := new(S)
10     s.i++
11 }
```

我所做的只是交换结构中的字段顺序。这次 int 类型的 i 字段在 4096 个元素的字节数组类型的 b 字段之前。现在让我们生成汇编代码，看看 nil 指针检查是否仍然存在：

```asm
09  s := new(S)
0x0036 00054 (main.go:9)   LEAQ "".autotmp_0001+0(SP),DI
0x003a 00058 (main.go:9)   MOVL $0,AX
0x003c 00060 (main.go:9)   MOVQ $513,CX
0x0043 00067 (main.go:9)   REP ,
0x0044 00068 (main.go:9)   STOSQ ,
0x0046 00070 (main.go:9)   LEAQ "".autotmp_0001+0(SP),BX

10  s.i++
0x004a 00074 (main.go:10)  NOP ,
0x004a 00074 (main.go:10)  MOVQ (BX),BP
0x004d 00077 (main.go:10)  NOP ,
0x004d 00077 (main.go:10)  INCQ ,BP
0x0050 00080 (main.go:10)  MOVQ BP,(BX)
0x0053 00083 (main.go:10)  NOP ,
```

如果将第 10 行的未修改的示例汇编代码与刚才修改了的（交换字段后的）示例进行比较，你会发现有很大的区别：

```asm
First Example      |   Second Example
CMPQ BX,$0         |   NOP
JEQ $1,105         |   MOVQ (BX),BP
MOVQ 4096(BX),BP   |   NOP ,
NOP ,              |   INCQ ,BP
INCQ ,BP           |   MOVQ BP,(BX)
MOVQ BP,4096(BX)   |   NOP ,
```

当我们将整数字段移动到结构中的第一位，nil 指针检查就消失了。现在 Go 再次让硬件来完成 nil 指针检查，并且报告无效的内存访问。

原因是 Go 可以信任硬件来识别和报告来自可能存在于 0x0 和 0xFFF 地址间的无效内存访问。编译器不会在这些情况下添加检查。但是当代码如果处理超出 4k 范围的地址，nil 指针检查就会被加入到汇编代码中。

在之前的例子中，当字节数组首先出现在结构体中时，整数字段的内存会分配超过 4k 边界的偏移量。这会导致编译器为整数字段添加 nil 指针检查。当整数字段是第一个时，编译器将其留给硬件进行检测，因为（内存）地址在 4k 范围内。

## 总结

我不相信编写专注于性能的代码。通常我们编写惯用、简洁的代码，然后对程序进行基准测试，并根据需要进行性能调整。正如我们所看到的，结构体的设计会对程序的性能有一定的影响。访问结构体中的字段需要进行指针运算，对于大于 4K 的结构体，你组织字段的方式可能会有所不同。

需要注意的是，我们看到的是编译器实现细节，而不是 Go 规范的一部分。这可能会随着 Go 的新版本发布而发生改变，或者 5g、6g 或 8g 编译器之间存在差异。在对实现细节进行性能调优时要小心，每一个新版本都需要重新检查这些细节。有趣的是，当编译器可以信任硬件时，它将如何使我们的代码更安全。

## 扩展材料

如果你想在标准库中阅读一些关于重新排序结构以消除字段的 nil 指针检查的示例，请看 Nigel Tao 的代码更改：https://codereview.appspot.com/6625058

此外还有 Russ Cox 于 2013 年 7 月撰写的一份文档 ,"Go 1.2 字段选择器和 nil 检查 "：http://golang.org/s/go12nil

---

via: https://www.ardanlabs.com/blog/2014/09/go-compiler-nil-pointer-checks.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[7Ethan](https://github.com/7Ethan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
