# 安全使用unsafe.Pointer

unsafe包提供了一个go类型系统的转义出口（escape hatch），支持以类似C语言的方式与底层和系统调用APIs进行交互。然而，为了更加合理的进行交互，必须遵守一些规则去使用unsafe包。在编写不安全代码时很容易犯一些细微的错误（译注：双关，也可以理解为在使用unsafe包的时候容易犯一些细微的错误），但这些错误通常是可以避免的。

这篇博客将会介绍一些已有的和将要有的，可以验证你go程序中unsafe.Pointer类型使用安全性的go工具。如果你之前没有使用过unsafe包，我推荐先阅读以前的Gopher学院来临系列博客的主题。

无论何时在代码基中使用unsafe包都需要额外的慎重小心，但这些验证工具能够他们在你程序中导致严重错误或者可能的安全漏洞前帮你解决问题。

## 用go vet验证编译时间

几年来，go vet工具已经具备检测unsafe.Pointer和uintptr类型之间无效转换的能力。
让我们来看个例子。假设我们希望使用指针运算来迭代和打印数组的每个元素：

```go
package main

import (
    "fmt"
    "unsafe"
)

func main() {
    // An array of contiguous uint32 values stored in memory.
    arr := []uint32{1, 2, 3}

    // The number of bytes each uint32 occupies: 4.
    const size = unsafe.Sizeof(uint32(0))

    // Take the initial memory address of the array and begin iteration.
    p := uintptr(unsafe.Pointer(&arr[0]))
    for i := 0; i < len(arr); i++ {
        // Print the integer that resides at the current address and then
        // increment the pointer to the next value in the array.
        fmt.Printf("%d ", (*(*uint32)(unsafe.Pointer(p))))
        p += size
    }
}
```

乍一看代码,似乎能和期望运行地一致，我们能看到在终端上数组的每个元素都打印了出来。
```shell
$ go run main.go 
1 2 3
```
然而，这段程序有个细微的缺陷。对此go vet会怎么看呢？

```shell
$ go vet .
# github.com/mdlayher/example
./main.go:20:33: possible misuse of unsafe.Pointer
```
为了能明白这个报错，我们需要查阅unsafe.Pointer类型的规则：

*将一个指针（Pointer）转换成一个指针类型（uintptr）时将生成指向值的内存地址，该地址是整型。这种uintptr的通常用途是用来打印它。*

*将一个uintptr转换回指针通常是无效的。*

*一个uintptr是一个整型，而不是引用。将一个指针转换成uintptr会创建一个没有指针语义的整型值。即使一个uintptr保存某个对象的地址，如果对象移动了，垃圾回收器（gc）将不会更新那个uintptr的值，同时uintptr也不会组织对象被回收。*

我们可以将有问题的代码分离出来列示如下:

```go
p := uintptr(unsafe.Pointer(&arr[0]))

// 如果这里触发了垃圾回收会发生什么?
fmt.Printf("%d ", (*(*uint32)(unsafe.Pointer(p))))
```
由于我们把uintptr的值存在p中但并没有立刻使用它，可能会发生当垃圾回收时，这个地址(现在是作为一个指针类型整型存在p中)将不再有效！

让我们假设这种情况已经发生且p不再指向一个uint32。--未完

事实上，一旦我们将一个unsafe.Pointer转换为uintptr，我们就不能安全地把它转换回unsafe.Pointer，除了一个特殊的情况：

*如果p指向一个分配的对象，则可以通过将其转换为uintptr、添加偏移量并转换回指针来通过该对象进行处理。*

为了安全地进行指针运算迭代逻辑，我们必须一次性的完成所有类型转换和指针运算。

```go
package main

import (
    "fmt"
    "unsafe"
)

func main() {
    // 一个有连续uint32值，存放在内存中的数组。
    arr := []uint32{1, 2, 3}

    // 每个uint32值占用字节大小： 4。
    const size = unsafe.Sizeof(uint32(0))

    for i := 0; i < len(arr); i++ {
        // Print an integer to the screen by:
        //   - taking the address of the first element of the array
        //   - applying an offset of (i * 4) bytes to advance into the array
        //   - converting the uintptr back to *uint32 and dereferencing it to
        //     print the value
        fmt.Printf("%d ", *(*uint32)(unsafe.Pointer(
            uintptr(unsafe.Pointer(&arr[0])) + (uintptr(i) * size),
        )))
    }
}
```

这段程序生成和之前相同的结果，但是现在go vet认为是有效的！

```shell
$ go run main.go 
1 2 3 
$ go vet 
```

我不推荐这种使用指针运算的方式来进行迭代逻辑，但Go在真正需要的时候提供这种转义出口(以及安全使用它的工具!)是非常棒的。

## 使用Go编译器的checkptr调试标记进行运行时验证

Go编译器最近增加一个新的调试标记位的支持，它能在运行时检测unsafe.Pointer无效的使用模式。在Go 1.13版本中，这个特性还没有发布，但是可以通过从tip里安装go来使用它：

```shell
$ go get golang.org/dl/gotip
go: finding golang.org/dl latest
...
$ gotip download
Updating the go development tree...
...
Success. You may now run 'gotip'!
$ gotip version
go version devel +8054b13 Thu Nov 28 15:16:27 2019 +0000 linux/amd64
```

让我们再来看看另一个例子。假设我们正将一个Go结构体传递给一个通常接受C共同体的Linux核心API。一种实现方式是使用一个包含原始字节数组（模拟一个C共同体）的Go结构体，然后 --未完

```go
package main

import (
    "fmt"
    "unsafe"
)

// one is a typed Go structure containing structured data to pass to the kernel.
// one是一个典型的go结构体包含了准备传递给内核的结构化数据。
type one struct{ v uint64 }

// two mimics a C union type which passes a blob of data to the kernel.
//  two模仿了一个C共同体类型
type two struct{ b [32]byte }

func main() {
    // Suppose we want to send the contents of a to the kernel as raw bytes.
    in := one{v: 0xff}
    out := (*two)(unsafe.Pointer(&in))

    // Assume the kernel will only access the first 8 bytes. But what is stored
    // in the remaining 24 bytes?
    fmt.Printf("%#v\n", out.b[0:8])
}
```

当我们在一个Go的稳定版本(截至Go版本1.13.4)执行这段程序时，我们能够看到数组的第一个8个字节用本地端字节序格式包含了我们uint64数据（我机器上是小端字节序）。

```shell
$ go run main.go
[]byte{0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
```

然而，这段程序同样也有一个问题。如果我们试着在Go tip上使用checkptr调试标记执行这段代码，我们将会看到：

```shell
$ gotip run -gcflags=all=-d=checkptr main.go 
panic: runtime error: unsafe pointer conversion

goroutine 1 [running]:
main.main()
        /home/matt/src/github.com/mdlayher/example/main.go:17 +0x60
exit status 2
```

这个检查仍然比较新，同时也没有提供除了“不安全的指针转换”（unsafe pointer conversion）报错信息和堆栈追踪之外更多的信息。但这个堆栈追踪至少提供了一个暗示：17行是有嫌疑的。

通过把小结构体转换为大结构体，我们有了超出小结构体数据边界读取任意内存的能力！这是另一个你应用中因为不谨慎使用unsafe包可能导致的安全漏洞。

为了安全的进行这种操作，我们必须确信我们在复制数据之前先初始化“共同体”结构体，因此，我们可以确保不访问任何内存:

```go
package main

import (
    "fmt"
    "unsafe"
)

// one is a typed Go structure containing structured data to pass to the kernel.
type one struct{ v uint64 }

// two mimics a C union type which passes a blob of data to the kernel.
type two struct{ b [32]byte }

// newTwo safely produces a two structure from an input one.
func newTwo(in one) *two {
    // Initialize out and its array.
    var out two

    // Explicitly copy the contents of in into out by casting both into byte
    // arrays and then slicing the arrays. This will produce the correct packed
    // union structure, without relying on unsafe casting to a smaller type of a
    // larger type.
    copy(
        (*(*[unsafe.Sizeof(two{})]byte)(unsafe.Pointer(&out)))[:],
        (*(*[unsafe.Sizeof(one{})]byte)(unsafe.Pointer(&in)))[:],
    )

    return &out
}

func main() {
    // All is well! The two structure is appropriately initialized.
    out := newTwo(one{v: 0xff})

    fmt.Printf("%#v\n", out.b[:8])
}
```

我们现在能够向之前使用相同的标记一样执行我们更新后的程序，我们能看到刚才那个问题已经解决了。

```shell
$ gotip run -gcflags=all=-d=checkptr main.go 
[]byte{0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
```

通过从fmt.Printf调用中移除切片操作, 我们可以验证字节数组的其余部分已经初始化为0字节:

```shell
[32]uint8{
	0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
}
```

