首发于：https://studygolang.com/articles/13976

# 理解 Go 语言的类型

当我使用 C/C++ 编写代码时，理解类型（type）是非常有必要的。如果不理解类型，你就会在编译或者运行代码的时候，碰到一大堆麻烦。无论什么语言，类型都涉及到了编程语法的方方面面。加强对于类型和指针的理解，对于提高编程水平十分关键。本文会主要讲解类型。

我们首先来看看这几个字节的内存：

FFE4 | FFE3 | FFE2 | FFE1
---|---|---|---
00000000 | 11001011 | 01100101 | 00001010

请问地址 FFE1 上字节的值是多少？如果你试图回答一个结果，那就是错的。为什么？因为我还没有告诉你这个字节表示什么。我还没有告诉你类型信息。

如果我说上述字节表示一个数字会怎么样呢？你可能会回答 10，那么你又错了。为什么？因为当我说这是数字的时候，你认为我是指十进制的数字。

> **基数（number base）：**
>
> 所有编号系统（numbering system）要发挥作用，都要有一个基（base）。从你出生的时候开始，人们就教你用基数 10 来数数了。这可能是因为我们大多数人都有 10 个手指和 10 个脚趾。另外，用基数 10 来进行数学计算也很自然。
>
> 基定义了编号系统所包含的符号数。基数 10 会有 10 个不同的符号，用以表示我们可以计量的无限事物。基数 10 的编号系统为 0、1、2、3、4、5、6、7、8、9。一旦超过了 9，我们需要增加数的长度。例如，10、100 和 1000。
>
> 在计算机领域，我们还一直使用其他两种基。第一种是基数 2（或二进制数），例如上图所表示的位。第二种是基数 16（或十六进制数），例如上图中表示的地址。
>
> 在二进制编号系统（基数 2）中，只有两种符号，即 0 和 1。
>
> 在十六进制数字系统（基数 16）中，有 16 个符号，这些符号分别是：0、1、2、3、4、5、6、7、8、9、A、B、C、D、E、F。
>
> 如果桌上有些苹果，那些苹果可以用任何编号系统来表示。我们可以说这里有：
>
> - 10010001 个苹果（使用 2 作为基数）
> - 145 个苹果（使用 10 作为基数）
> - 91 个苹果（使用 16 作为基数）
>
> 所有答案都正确，只要给定了正确的基。
>
> 注意每个编号系统表示那些苹果所需要的符号数。基数越大，编号系统的效率就越高。
>
> 对于计算机地址、IP 地址和颜色代码，使用 16 作为基数，就显得很有价值。
>
> 看看用三种基，来分别表示 HTML 的颜色（“白”）的数字：
>
> - 使用 2 作为基数：1111 1111 1111 1111 1111 1111（24 个字符）
> - 使用 10 作为基数：16777215（10 个字符）
> - 使用 16 作为基数：FFFFFF（6 个字符）
>
> 你会选择哪个编号系统来表示颜色呢？

现在，如果我告诉你，地址 FFE1 处的字节表示一个基数为 10 的数字，你回答 10，这就正确了。

类型提供了两条信息，你和编译器都需要它来执行我们刚刚经历过的练习。

1. 要查看的内存数量（以字节为单位）
2. 这些字节的表示

Go 语言提供了以下基本数字类型：

> **无符号整数**
>
> uint8, uint16, uint32, uint64
>
> **有符号整数**
>
> int8, int16, int32, int64
>
> **实数**
>
> float32, float64
>
> **预声明整数**
>
> uint, int, uintptr

这些关键字提供了所有的类型信息。

`uint8` 包含一个基为 10 的数字，用 1 个存储字节表示。`uint8` 的值从 0 到 255。

`int32` 包含一个基为 10 的数字，用 4 个存储字节表示。`int32` 的值从 -2147483648 到 2147483647。

预声明整数会根据你构建代码时的体系结构来进行映射。在 64 位操作系统上，`int` 将映射到 `int64`，而在 32 位系统上，它将映射到 `int32`。

所有存储在内存中的内容都解析为某种数字类型。在 Go 中，字符串只是一系列 `uint8` 类型，并包含了一些规则，用于关联这些字节和识别字符串的结尾位置。

在 Go 中，指针就是 `uintptr` 类型。同样地，基于操作系统的体系结构，它将映射为 `uint32` 或者 `uint64`。Go 为指针创建了一个特殊的类型。在过去，许多 C 程序员在编写代码时，会认为指针值总能符合 `unsigned int`。随着时间的推移，语言和体系结构不断升级，最终这不再是对的了。由于地址变得比预先声明的 `unsigned int` 更大，很多代码都出错了。

结构体类型只是很多类型的组合，而这些类型也最终会解析为数字类型。

```go
type Example struct{
    BoolValue bool
    IntValue  int16
    FloatValue float32
}
```

该结构体表示一个复杂类型。它表示 7 个字节，有三种不同的数字表示。`bool` 有 1 个字节，`int16` 有 2 个字节，而 `float32` 有 4 个字节。但是，这个结构体最终在内存中分配了 8 个字节。

为了最大限度地减少内存碎片整理（memory defragmentation），分配内存时都会将内存边界对齐。要确定 Go 在体系结构上所用的对齐边界（alignment boundary），你可以运行 `unsafe.Alignof` 函数。Go 在 64 位 Darwin 平台的对齐边界是 8 个字节。因此在 Go 确定我们结构体的内存分配时，它将填充字节以确保最终占用的内存是 8 的倍数。编译器会决定在哪里添加填充。

如果你想要学习更多有关结构体成员对齐和填充的知识，请查看下面的链接：

http://www.geeksforgeeks.org/structure-member-alignment-padding-and-data-packing/

下面的程序会显示对于 `Example` 结构体类型，Go 向内存所插入的填充：

```go
package main

import (
    "fmt"
    "unsafe"
)

type Example struct {
    BoolValue bool
    IntValue int16
    FloatValue float32
}

func main() {
    example := &Example{
        BoolValue:  true,
        IntValue:   10,
        FloatValue: 3.141592,
    }

    exampleNext := &Example{
        BoolValue:  true,
        IntValue:   10,
        FloatValue: 3.141592,
    }

    alignmentBoundary := unsafe.Alignof(example)

    sizeBool := unsafe.Sizeof(example.BoolValue)
    offsetBool := unsafe.Offsetof(example.BoolValue)

    sizeInt := unsafe.Sizeof(example.IntValue)
    offsetInt := unsafe.Offsetof(example.IntValue)

    sizeFloat := unsafe.Sizeof(example.FloatValue)
    offsetFloat := unsafe.Offsetof(example.FloatValue)

    sizeBoolNext := unsafe.Sizeof(exampleNext.BoolValue)
    offsetBoolNext := unsafe.Offsetof(exampleNext.BoolValue)

    fmt.Printf("Alignment Boundary: %d\n", alignmentBoundary)

    fmt.Printf("BoolValue = Size: %d Offset: %d Addr: %v\n",
        sizeBool, offsetBool, &example.BoolValue)

    fmt.Printf("IntValue = Size: %d Offset: %d Addr: %v\n",
        sizeInt, offsetInt, &example.IntValue)

    fmt.Printf("FloatValue = Size: %d Offset: %d Addr: %v\n",
        sizeFloat, offsetFloat, &example.FloatValue)

    fmt.Printf("Next = Size: %d Offset: %d Addr: %v\n",
        sizeBoolNext, offsetBoolNext, &exampleNext.BoolValue)
}
```

输出如下所示：

```bash
Alignment Boundary: 8
BoolValue  = Size: 1  Offset: 0  Addr: 0x21015b018
IntValue   = Size: 2  Offset: 2  Addr: 0x21015b01a
FloatValue = Size: 4  Offset: 4  Addr: 0x21015b01c
Next       = Size: 1  Offset: 0  Addr: 0x21015b020
```

该结构体类型的对齐边界的确是 8 字节。

`Size` 大小值表示某字段读写时所用的内存。不出所料，该值与字段的类型信息相一致。

`Offset` 偏移值表示字段的开始位置，在内存占用中的字节序号。

`Addr` 地址值表示每个字段开始在内存占用中所处的位置。

我们可以看到，Go 在 `BoolValue` 和 `IntValue` 字段之间填充了 1 个字节。偏移值和两个地址之差是 2 个字节。你还可以看到，下一个内存分配时是从结构体最后的字段处分配 4 个字节。

我们让结构体只有一个 `bool` 字段（1 字节），来证实 8 字节对齐法则。

```go
package main

import (
    "fmt"
    "unsafe"
)

type Example struct {
    BoolValue bool
}

func main() {
    example := &Example{
        BoolValue:  true,
    }

    exampleNext := &Example{
        BoolValue:  true,
    }

    alignmentBoundary := unsafe.Alignof(example)

    sizeBool := unsafe.Sizeof(example.BoolValue)
    offsetBool := unsafe.Offsetof(example.BoolValue)

    sizeBoolNext := unsafe.Sizeof(exampleNext.BoolValue)
    offsetBoolNext := unsafe.Offsetof(exampleNext.BoolValue)

    fmt.Printf("Alignment Boundary: %d\n", alignmentBoundary)

    fmt.Printf("BoolValue = Size: %d Offset: %d Addr: %v\n",
        sizeBool, offsetBool, &example.BoolValue)

    fmt.Printf("Next = Size: %d Offset: %d Addr: %v\n",
        sizeBoolNext, offsetBoolNext, &exampleNext.BoolValue)
}
```

其输出如下：

```bash
Alignment Boundary: 8
BoolValue = Size: 1 Offset: 0 Addr: 0x21015b018
Next      = Size: 1 Offset: 0 Addr: 0x21015b020
```

把两个地址相减，你将看到两种结构体类型分配之间存在 8 个字节的间隙。此外，这一次的内存分配从上一示例相同的地址开始。为了保持对齐边界，Go 向结构体填充了 7 个字节。

无论如何填充，`Size` 值实际上表示我们可以为每个字段读写的内存大小。

我们只能在使用数字类型时，才能操作内存，通过赋值运算符（=）可以做到这一点。为了方便，Go 创建了一些可以支持赋值运算符的复杂类型。这些类型有字符串、数组和切片。要查看这些类型的完整列表，请查看此文档：http://golang.org/ref/spec#Types。

这些复杂类型其实对底层数字类型进行了抽象，我们可以在各种复杂类型的实现发现这一点。在这种情况下，这些复杂类型可以像数字类型那样直接读取内存。

Go 是一种类型安全的语言。这意味着，编译器将始终强制赋值运算符的两边类型保持相似。这非常重要，因为这会防止我们错误地读取内存。

假设我们想做下面的事。如果你试图编译代码，你会得到一个错误。

```go
type Example struct{
    BoolValue bool
    IntValue  int16
    FloatValue float32
}

example := &Example{
    BoolValue:  true,
    IntValue:   10,
    FloatValue: 3.141592,
}

var pointer *int32
pointer = *int32(&example.IntValue)
*pointer = 20
```

我试图获取 `IntValue` 字段（2 个字节）的内存地址，并把它存储在类型为 `int32` 的指针上。接下来，我试图用指针，向内存地址写入一个 4 个字节的整数。如果可以使用该指针，那么我就会违反 `IntValue` 字段的类型规则，并在此过程中破坏内存。

FFE8 | FFE7 | FFE6 | FFE5 | FFE4 | FFE3 | FFE2 | FFE1
---|---|---|---|---|---|---|---
0 | 0 | 0 | 3.14 | 0 | 10 | 0 | true

pointer |
---|
FFE3 |

FFE8 | FFE7 | FFE6 | FFE5 | FFE4 | FFE3 | FFE2 | FFE1
---|---|---|---|---|---|---|---
0 | 0 | 0 | 0 | 0 | 20 | 0 | true

**根据上面的内存占用情况，指针将在 FFE3 和 FFE6 之间的 4 个字节中写入 20。`IntValue` 的值将如预期的那样变为 20，但 `FloatValue` 的值现在等于 0。想象一下，写入这些字节超出了该结构体的内存分配，并且开始破坏应用的其他区域的内存。随之而来的错误会是随机、不可预测的**。

**Go 编译器会一直保证内存对齐和转型是安全的**。

**在下面一个转型的示例中，编译器会报错**：

```go
ackage main

import (
    "fmt"
)

// Create a new type
type int32Ext int32

func main() {
    // Cast the number 10 to a value of type Jill
    var jill int32Ext = 10

    // Assign the value of jill to jack
    // ** cannot use jill (type int32Ext) as type int32 in assignment **
    var jack int32 = jill

    // Assign the value of jill to jack by casting
    // ** the compiler is happy **
    var jack int32 = int32(jill)

    fmt.Printf("%d\n", jack)
}
```

首先，我们在系统中新建了一个 `int32Ext` 类型，并告诉编译器该类型表示一个 `int32`。接下来，我们创建了一个名为 `jill` 的新变量，将其赋值为 10。编译器允许这个赋值操作，因为数字类型在赋值运算符的右侧。编译器知道赋值是安全的。

现在，我们尝试创建第二个变量，名为 `jack`，其类型为 `int32`，我们将 `jill` 赋值给 `jack`。在这里，编译器会抛出错误：

```bash
cannot use jill (type int32Ext) as type int32 in assignment
```

编译器认为 `jill` 的类型是 `int32Ext`，不会对赋值的安全性作出任何假设。

现在我们使用强制转换，编译器允许赋值，并如预期打印出值来。当我们执行转型时，编译器会检查赋值的安全性。在这里，编译器确定了这是相同类型的值，于是允许赋值操作。

对于某些读者来说，这似乎很基础，但它是使用任何编程语言的基石。即使类型是经过抽象的，你也是在操作内存，你应该知道你究竟在做些什么。

有了这些基础，我们才可以在 Go 中讨论指针，然后将参数传递给函数。

像往常一样，我希望这篇文章，能够帮助你了解一些可能存在的盲区。

---

via: https://www.ardanlabs.com/blog/2013/07/understanding-type-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
