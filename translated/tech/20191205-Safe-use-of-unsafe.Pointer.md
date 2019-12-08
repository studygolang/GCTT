# 安全使用unsafe.Pointer

unsafe包提供了一个go类型系统的安全舱口（escape hatch），支持以类似C语言的方式与底层和系统调用APIs进行交互。然而，为了更加合理的进行交互，必须遵守一些规则去使用unsafe包。在编写不安全代码时很容易犯一些细微的错误（译注：双关，也可以理解为在使用unsafe包的时候容易犯一些细微的错误），但这些错误通常是可以避免的。

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