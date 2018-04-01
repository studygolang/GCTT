已发布：https://studygolang.com/articles/12730

# 使用 Go 语言写一个即时编译器(JIT)

JIT(Just-int-time) 编译器是任何程序在被转换成机器码的运行过程中产生的。JIT 代码和其他代码（比如，fmt.Println）的区别在于 JIT 代码是在运行过程中生成的。

用 Golang 编写的程序是静态类型且提前编译。生成任意代码似乎是不可能的，更不用说执行所述代码了。但是，可以将指令发送到正在运行的进程。这是使用 Type Magic 完成的 - 将任何类型转换为任何其他类型的能力。 

请注意，如果您有兴趣了解更多关于 Type Magic 的信息，请在下面留言，我会在后面写文阐述。

## x64 指令集上的 JIT 编译器

机器码是对处理器有特殊意义的一系列字节。用于编写此博客并测试代码的机器使用的是 x64 处理器，因此我使用了 [`x64` 指令集](https://software.intel.com/en-us/articles/introduction-to-x64-assembly)。 

以下代码必须在 x64 处理器上运行。

## 生成 x64 代码打印 "Hello World"

为了打印 “Hello World”，系统调用应该指示处理器打印数据。打印数据的系统调用是 [write(int fd，const void * buf，size_t count)](http://man7.org/linux/man-pages/man2/write.2.html)。 

此系统调用的第一个参数是要写入的位置，表示为文件描述符。将输出打印到控制台是通过写入标准文件描述符 stdout 来实现的。stdout 的文件描述符编号为 1. 

第二个参数是必须写入的数据的位置。有关这方面的更多信息将在下一节中提供。 

第三个操作数是 count - 即要写入的字节数。在 “Hello World！” 的情况下，要写入的字节数为 12。为了进行系统调用，需要将三个操作数保存在特定的寄存器中。这里有一个表格显示了保存操作数的寄存器。

|Syscall # |Param 1 | Param 2 | Param 3 | Param 4 | Param 5 | Param 6 |
|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|
|rax|rdi|rsi|rdx|r10|r8|r9

将所有这些放在一起，这里是一系列代表初始化一些寄存器的指令的字节。

```
0:  48 c7 c0 01 00 00 00    mov    rax,0x1  
7:  48 c7 c7 01 00 00 00    mov    rdi,0x1  
e:  48 c7 c2 0c 00 00 00    mov    rdx,0xc  
```

- 第一条指令将 rax 设置为 1 - 表示写入系统调用。
- 第二条指令将 rdi 设置为 1 - 表示 stdout 的文件描述符
- 第三条指令将 rdx 设置为 12 以表示要打印的字节数。
- 数据的位缺失，实际上调用 write 就是如此

为了指定包含 “Hello World！” 的数据的位置，数据需要先拥有一个位置 - 即它需要存储在内存中的某个位置。 

表示 “Hello World！” 的字节序列是 48 65 6c 6c 6f 20 57 6f 72 6c 64 21。这应该存储在处理器不会尝试执行的位置。否则，该程序将引发段错误（segmentation fault）。 

在这种情况下，数据可以存储在可执行指令的末尾 - 即在返回指令之后。在返回指令之后存储数据是安全的，因为处理器在遇到返回时“跳”到不同的地址，并且不会顺序执行。 

由于直到返回指令被布置时才知道过去的返回地址，所以可以使用它的临时占位符，并且一旦数据的地址已知就用正确的地址替换。这是连接器所遵循的确切程序。链接过程只需填写这些地址以指向正确的数据或函数。

```
15: 48 8d 35 00 00 00 00    lea    rsi,[rip+0x0]      # 0x15  
1c: 0f 05                   syscall  
1e: c3                      ret  
```

在上面的代码中，加载 “Hello World！” 地址的 lea 指令指向自己（指向距离 rip 0 字节的位置）。这是因为数据尚未存储，数据地址未知。 

系统调用本身由字节序列 0F 05 表示。

现在可以存储数据，因为返回指令已经布置。

```
1f: 48 65 6c 6c 6f 20 57 6f 72 6c 64 21   // Hello World!
```

在整个程序中，现在我们可以更新指令来指向数据。以下是更新的代码：

```
0:  48 c7 c0 01 00 00 00    mov    rax,0x1
7:  48 c7 c7 01 00 00 00    mov    rdi,0x1
e:  48 c7 c2 0c 00 00 00    mov    rdx,0xc
15: 48 8d 35 03 00 00 00    lea    rsi,[rip+0x3]        # 0x1f
1c: 0f 05                   syscall
1e: c3                      ret
1f: 48 65 6c 6c 6f 20 57 6f 72 6c 64 21   // Hello World! 
```

上面的代码可以表示为 Golang 中任何基本类型的片段。 

uint16 类型的数组/ slice 是一个不错的选择，因为它可以保存成对的小端有序单词，同时仍然保持可读性。这里是保存上述程序的 `[]uint16` 数据结构

```go
printFunction := []uint16{
	0x48c7，0xc001, 0x0，                  //  mov  %rax ，$0x1
	0x48，0xc7c7，0x100，0x0，           // mov %rdi ，$0x1
	0x48c7, 0xc20c, 0x0,                // mov 0x13, %rdx
	0x48, 0x8d35, 0x400, 0x0,           // lea 0x4(%rip), %rsi
	0xf05,                              // syscall
	0xc3cc,                             // ret
	0x4865, 0x6c6c, 0x6f20,             // Hello_(whitespace)
	0x576f, 0x726c, 0x6421, 0xa,        // World!
} 
```

与上面列出的字节相比，上述字节略有偏差。这是因为当它与切片条目的开始对齐时，它更清晰（更易于读取和调试）来表示数据 “Hello World！”。 

因此，我使用填充指令 cc 指令（无操作）将数据部分的开始推送到 slice 中的下一个条目。我还更新了 lea 指向 4 个字节的位置以反映这一变化。 

注意：您可以在此`[链接](https://filippo.io/linux-syscall-table/)找到各种系统调用的系统调用号码。

## 转换切片函数

`[]uint16` 数据结构中的指令必须转换为一个函数，以便可以调用它。下面的代码演示了这种转换。

```go
type printFunc func()

unsafePrintFunc := (uintptr)(unsafe.Pointer(&printFunction)) 
printer := *(*printFunc)(unsafe.Pointer(&unsafePrintFunc)) 
printer()
```

Golang 函数值只是一个指向 C 函数指针的指针（注意两级指针）。从切片到函数的转换首先是提取一个指向保存可执行代码的数据结构的指针。这存储在 unsafePrintFunc 中。指向 unsafePrintFunc 的指针可以被转换为所需的函数类型。 

此方法仅适用于没有参数或返回值的函数。需要为调用具有参数或返回值的函数创建堆栈帧。函数定义应始终以指令开始，以动态分配堆栈帧以支持可变参数函数。有关不同函数类型的更多信息，请参阅[此处](https://docs.google.com/document/d/1bMwCey-gmqZVTpRax-ESeVuZGmjwbocYs1iHplK-cjo/pub)。 

如果您希望我写关于在 Golang 中生成更复杂的函数的信息，请在下面评论。

## 使函数可执行

上述函数不会实际运行。这是因为 Golang 将所有数据结构存储在二进制文件的数据部分。本节中的数据设置了[No-Execute](https://en.wikipedia.org/wiki/NX_bit)标志，阻止其执行。 

printFunction slice 中的数据需要存储在一段可执行的内存中。这可以通过删除 printFunction slice 上的 No-Execute 标志或将其复制到可执行的内存位置来实现。 

在下面的代码中，数据已被复制到一个新分配的可执行内存（使用 mmap）。这种方法比较好，因为只在整个页面上设置不执行标志 - 很容易使数据部分的其他部分无法执行。

```go
executablePrintFunc, err := syscall.Mmap(
	-1,
	0,
	128,  
	syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC, 
	syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS)

if err != nil {
	fmt.Printf("mmap err: %v", err)
}
j := 0
for i := range printFunction {
	executablePrintFunc[j] = byte(printFunction[i] >> 8)
	executablePrintFunc[j+1] = byte(printFunction[i])
	j = j + 2
}
```

标志 syscall.PROT_EXEC 确保新分配的内存地址是可执行的。将此数据结构转换为函数将使其运行平稳。

以下是完整的代码，尝试在x64机器上运行。

```go
package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

type printFunc func()

func main() {
	printFunction := []uint16{
		0x48c7, 0xc001, 0x0, // mov %rax,$0x1
		0x48, 0xc7c7, 0x100, 0x0, // mov %rdi,$0x1
		0x48c7, 0xc20c, 0x0, // mov 0x13, %rdx
		0x48, 0x8d35, 0x400, 0x0, // lea 0x4(%rip), %rsi
		0xf05,                  // syscall
		0xc3cc,                 // ret
		0x4865, 0x6c6c, 0x6f20, // Hello_(whitespace)
		0x576f, 0x726c, 0x6421, 0xa, // World!
	}
	executablePrintFunc, err := syscall.Mmap(
		-1,
		0,
		128,
		syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC,
		syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS)
	if err != nil {
		fmt.Printf("mmap err: %v", err)
	}
	j := 0
	for i := range printFunction {
		executablePrintFunc[j] = byte(printFunction[i] >> 8)
		executablePrintFunc[j+1] = byte(printFunction[i])
		j = j + 2
	}
	type printFunc func()
	unsafePrintFunc := (uintptr)(unsafe.Pointer(&executablePrintFunc))
	printer := *(*printFunc)(unsafe.Pointer(&unsafePrintFunc))
	printer()
}
```

## 结论

尝试以上源代码。敬请期待 Golang 的深入探索！

---

via: https://medium.com/kokster/writing-a-jit-compiler-in-golang-964b61295f
作者：[Sidhartha Mani](https://medium.com/@utter_babbage)
译者：[jiangwei161002010](https://github.com/jiangwei161002010)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出