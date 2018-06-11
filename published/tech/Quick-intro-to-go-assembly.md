# Go 语言汇编快速入门

在 Go 的源码中包含大量汇编语句，最优秀的示例代码位于 `math/big`, `runtime` 和 `crypto` 这些库中，但是从这里入门的话实在太过于痛苦，这些示例都是着力于系统操作和性能的运行代码。

对于没有经验的 Go 语言爱好者来说，这样会使通过库代码的学习过程遇到很大困难 。这也是撰写本文的原因所在。

Go ASM ( 译者注：ASM 是汇编的简写 ) 是一种被 Go 编译器使用的特殊形式的汇编语言，而且它基于 Plan 9 （译者注：来自贝尔实验室的概念[网络操作系统 ](https://baike.baidu.com/item/%E7%BD%91%E7%BB%9C%E6%93%8D%E4%BD%9C%E7%B3%BB%E7%BB%9F)）输入风格，所以先从 [文档](https://9p.io/sys/doc/asm.pdf) 开始是一个不错的选择。

注意：本文的内容是基于 x86_64 架构，但大多数示例也能兼容 x86 架构。

一些例子是从原始文档中选取出来的，主要目的是建立一个综合的统一标准摘要，涵盖那些最重要/有用的主题。

## 第一步

Go ASM 和标准的汇编语法（ NASM 或 YASM ）不太一样，首先你会发现它是架构独立的，没有所谓的 32 或 64 位寄存器，如下图所示：

| NASM x86 | NASM x64 | Go ASM |
| -------- | -------- | ------ |
| eax      | rax      | AX     |
| ebx      | rbx      | BX     |
| ecx      | rcx      | CX     |
| …        | …        | …      |

大部分寄存器符号都依赖于架构。

另外， Go ASM 还有四个预定义的符号作为伪寄存器。它们不是真正意义上的寄存器，而是被工具链维持出来的虚拟寄存器，这些符号在所有架构上都完全一样：

- `FP`: 帧指针 –参数和局部变量–
- `PC`: 程序计数器 –跳转和分支–
- `SB`: 静态基址指针 –全局符号–
- `SP`: 栈指针 –栈的顶端–.

这些虚拟寄存器在 Go ASM 中占有了重要地位，并且被广泛使用，其中最重要的就要属 SB 和 FP了。

伪寄存器 SB 可以看作是内存的起始地址，所以 foo(SB) 就是 foo 在内存中的地址。语法中有两种修饰符，<> 和 +N （N是一个整数）。第一种情况 foo<>(SB) 代表了一个私有元素，只有在同一个源文件中才可以访问，类似于 Go 里面的小写命名。第二种属于对相对地址加上一个偏移量后得到的地址，所以 foo+8(SB) 就指向 foo 之后 8 个字节处的地址。

伪寄存器 FP 是一个虚拟帧指针，被用来引用过程参数，这些引用由编译器负责维护，它们将指向从伪寄存器处偏移的栈中参数。在一台 64 位机器上， 0(FP)  是第一个参数， 8(FP) 就是第二个参数。为了引用这些参数，编译器会强制它们的命名使用，这是出于清晰和可读性的考虑。所以  MOVL foo+0(FP), CX  会把虚拟的 FP 寄存器中的第一个参数放入到物理上的 CX 寄存器，以及 MOVL bar+8(FP), DX 会把第二个参数放入到 DX 寄存器中。

读者可能已经注意到这种 ASM 语法类似 AT&T  风格，但不完全一致：

| Intel              | AT&T                 | Go                 |
| ----- | ------- | ----- |
| `mov eax, 1`       | `movl $1, %eax`      | `MOVQ $1, AX`      |
| `mov rbx, 0ffh`    | `movl $0xff, %rbx`   | `MOVQ $(0xff), BX` |
| `mov ecx, [ebx+3]` | `movl 3(%ebx), %ecx` | `MOVQ 2(BX), CX`   |

另一处显著的差异就是全局源码文件结构， NASM 中的代码结构是用 section 清晰的定义出来：

```asm
global start

section .bss
	…

section .data
	…

section .text
start:
	mov 	rax, 0x2000001
	mov 	rdi, 0x00
	syscall
```

而在 Go 汇编中则是靠预定义的 section 类型符号：

```asm
DATA 	myInt<>+0x00(SB)/8, $42
GLOBL 	myInt<>(SB), RODATA, $8

// func foo()
TEXT ·foo(SB), NOSPLIT, $0
	MOVQ 	$0, DX
	LEAQ 	myInt<>(SB), DX
	RET
```

这种语法使得我们能够尽可能的在最适合的地方定义符号。

## 在 Go 中调用汇编代码

可以从介绍中发现，Go 中的汇编代码主要用于优化和与底层系统交互，这使得 Go ASM 并不会像其它的经典汇编代码那样独立运行。Go ASM 必须在 Go 代码中调用。

hello.go

```go
package main

func neg(x uint64) int64

func main() {
	println(neg(42))
}
```

hello_amd64.s

```asm
TEXT ·neg(SB), NOSPLIT, $0
	MOVQ 	x+0(FP), AX
	NEGQ 	AX
	MOVQ 	AX, ret+8(FP)
	RET
```

运行这份代码将会在终端打印出 -42 。

注意子过程符号开始处的 unicode 中间点 `·` ，这是为了包名分隔，没有前缀的 `·foo` 等价于  `main·foo`。

过程中的 `TEXT ·neg(SB), NOSPLIT, $0` 意味着：

- `TEXT`: 这个符号位于 `text` section。
- `·neg`: 该过程的包符号和符号。
- `(SB)`: 词法分析器会用到。
- `NOSPLIT`: 使得没有必要定义参数大小。–可以省略不写–
- `$0`: 参数的大小, 如果定义了`NOSPLIT` 就是 `$0` 。

build的步骤仍旧和往常一样，使用 `go build` 命令， Go 编译器会根据文件名–`amd64`–自动链接`.s` 文件。

还有一份资源可以帮助学习 Go 文件的编译过程，我们可以看下 `go tool build -S <file.go>` 生成的 Go ASM 。

一些类似 `NOSPLIT` 和 `RODATA` 的符号都是在 `textflax` 头文件中定义，因此用`#include textflag.h` 包含 该文件可以有利于完成一次没有报错的完美编译。

## MacOS 种的系统调用

MacOS 中的系统调用需要在加上调用号 `0x2000000` 后才能被调用,举个例子，exit 系统调用就是 `0x2000001` 。调用号开始处的 `2` 是因为有多个种类的调用被定义在了重叠的调用号范围，这些类型都是定义在 [这里](https://opensource.apple.com/source/xnu/xnu-792.10.96/osfmk/mach/i386/syscall_sw.h) :

```c
#define SYSCALL_CLASS_NONE	0	/* Invalid */
#define SYSCALL_CLASS_MACH	1	/* Mach */
#define SYSCALL_CLASS_UNIX	2	/* Unix/BSD */
#define SYSCALL_CLASS_MDEP	3	/* Machine-dependent */
#define SYSCALL_CLASS_DIAG	4	/* Diagnostics */
```

所有的 MacOS 系统调用号列表可以在 [这里](https://opensource.apple.com/source/xnu/xnu-1504.3.12/bsd/kern/syscalls.master) 找到.

参数是通过这些寄存器 `DI`, `SI`, `DX`, `R10`, `R8` 和`R9` 传递给系统调用, 系统调用代码存放在 `AX` 中。

NASM 中的写法类似这样：

```asm
mov     rax, 0x2000004 	; 写系统调用
mov     rdi, 1 			; 参数 1 fd (stdout)
mov     rsi, rcx 		; 参数 2 buf
mov     rdx, 16 		; 参数 3 count
syscall
```

与之相反，Go ASM 中类似的例子则是像这样：

```asm
MOVL 	$1,  DI 		// 参数 1 fd (stdout)
LEAQ 	CX,  SI 		// 参数 2 buf
MOVL 	$16, DX 		// 参数 3 count
MOVL 	$(0x2000000+4), AX 	// 写系统调用
SYSCALL
```

同样，系统调用代码被放置在 `SYSCALL` 指令之前，这仅仅是通用写法，你可以像在 NASM 中那样直接把写系统调用放在最前面，编译后不会报任何错误。

## 使用字符串

现在我相信你已经能够写一些基本的汇编代码并运行了，例如经典的 hello world 。我们知道如何把一个参数传递给子过程，如何返回值和如果在数据 section 里面定义符号。你试过定义一个字符串么？

几天前我在编写一些汇编代码的时候遇到了这个问题，而我最关心的问题是，我该如何做才能去定义一个操蛋的字符串？嗯，NASM 中可以像这样来定义字符串：

```asm
section data:
	foo: db "My random string", 0x00
```

可这在 Go 中不行，在我深入研究了我能从网上找到的所有 go ASM 项目后，我还是没能找到一个定义简单字符串的示例。最后我在 Plan9 汇编语言文档中找到了一个例子，它可以说明怎样让目标实现。

Go 和 Plan9 唯一的不同之处是使用双引号而非单引号，并且添加了一个`RODATA` 符号:

```asm
DATA  foo<>+0x00(SB)/8, $"My rando"
DATA  foo<>+0x08(SB)/8, $"m string"
DATA  foo<>+0x16(SB)/1, $0x0a
GLOBL foo<>(SB), RODATA, $24

TEXT ·helloWorld(SB), NOSPLIT, $0
	MOVL 	$(0x2000000+4), AX 	// syscall write
	MOVQ 	$1, DI 			// arg 1 fd
	LEAQ 	foo<>(SB), SI 		// arg 2 buf
	MOVL 	$24, DX 		// arg 3 count
	SYSCALL
	RET
```

注意，定义字符串时不能放在一起，需要把它们定义在 8 字节（ 64 位）的块中。

现在你可以深入 Go ASM 世界中写下你自己的超级快速和极端优化的代码l，并请记住，去读那些操蛋的手册（微笑脸）。

## 在安全领域使用？

使用汇编除了优化你的 Go 代码外，也可以很方便的避免触发常见签名从而规避防病毒软件，以及使用一些反编译技术规避沙箱来搜寻异常行为，或者只是让分析师哀嚎。

如果你对此有兴趣，我会在该主题的下一篇文章中介绍，敬请关注！

## 附录

- https://golang.org/doc/asm
- https://9p.io/sys/doc/asm.pdf
- https://goroutines.com/asm
- https://blog.sgmansfield.com/2017/04/a-foray-into-go-assembly-programming/

---

via: https://blog.hackercat.ninja/post/quick_intro_to_go_assembly/

作者：[hcn](https://blog.hackercat.ninja/)
译者：[sunzhaohao](https://github.com/sunzhaohao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
