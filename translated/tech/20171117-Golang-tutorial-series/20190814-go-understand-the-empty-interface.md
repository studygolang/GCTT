# 理解 Go 的空接口

<!-- https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-vet-command-is-more-powerful-than-you-think/go-vet.png 图片链接模板 -->
!["Golang 之旅"插图，由 Go Gopher 的 Renee French 创作](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-understand-the-empty-interface/gopher.png)

空接口可用于保存任何数据，它可以是一个有用的参数，因为它可以使用任何类型。要理解空接口如何工作以及如何保存任何类型，我们首先应该理解空接口名称背后的概念。

## 接口（interface{}）

[Jordan Oreilli](https://jordanorelli.com/post/32665860244/how-to-use-interfaces-in-go) 对空接口的一个很好的定义：

> 接口是两件事物：它是一组方法，但它也是一种类型。
>
> `interface{}` 类型是没有方法的接口。由于没有 `implements` 关键字，所有类型都至少实现零个方法，并且自动满足接口，所有类型都满足空接口。

因此，空接口作为参数的方法可以接受任何类型。Go 将继续转换为接口类型以满足这个函数。

Russ Cox 撰写了一篇 [关于接口内部结构的精彩文章](https://research.swtch.com/interfaces)，并解释了接口由两个单词组成：

* 指向存储类型信息的指针
* 指向关联数据的指针

以下是 Russ 在 2009 年画的示意图，[当时 `runtime` 包还是用 C 语言编写](https://go.googlesource.com/go/+/refs/heads/release-branch.go1/src/pkg/runtime/iface.c)：

![internal-representation](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-understand-the-empty-interface/internal-representation.png)

`runtime` 包现在用 Go 编写，但结构未变。我们可以通过打印空接口来验证：

```go
func main() {
    var i int8 = 1
    read(i)
}

//go:noinline
func read(i interface{}) {
    println(i)
}
```

```shell
(0x10591e0,0x10be5c6)
```

两个地址分别代表了类型信息和值的两个指针。

## 底层结构

空接口的底层结构记录在反射包中：

```go
type emptyInterface struct {
   typ  *rtype            // word 1 with type description
   word unsafe.Pointer    // word 2 with the value
}
```

正如之前解释的那样，我们可以清楚的看到空结构体有一个类型描述字段和一个包含着值的字段。

`rtype` 结构体包含了类型的基本描述信息：

```go
type rtype struct {
   size       uintptr
   ptrdata    uintptr
   hash       uint32
   tflag      tflag
   align      uint8
   fieldAlign uint8
   kind       uint8
   alg        *typeAlg
   gcdata     *byte
   str        nameOff
   ptrToThis  typeOff
}
```

在这些字段中，有些非常简单，很容易可以看出：

<!--    TODO 怎么看出来的？？ -->
* `size`  是以字节为单位的大小
* `kind`  包含类型有：int8，int16，bool 等。
* `align` 是变量与此类型的对齐方式

<!-- TODO 从哪能获取到方法？ -->
根据空接口嵌入的类型，我们可以映射导出字段或列出方法：

```go
type structType struct {
   rtype
   pkgPath name
   fields  []structField
}
```

<!-- TODO  a flat conversion 怎么理解？ -->
<!-- 从哪看出两个映射？ -->
这个结构还有两个映射，包含字段列表。它清楚地表明，将内建类型转换为空接口将导致平面转换，其中字段的描述及值将存储在内存中。

下边是我们看到的空结构体的表示：

![结构体由两个词构成](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-understand-the-empty-interface/interface-representation.png)

结构体由两个词构成

现在让我们看看空接口实际上可以实现哪种转换。

## 转换

让我们尝试一个使用空接口的简单程序进行错误转换：

```go
func main() {
    var i int8 = 1
    read(i)
}

//go:noinline
func read(i interface{}) {
    n := i.(int16)
    println(n)
}
```

虽然转换 `int8` 为 `int16` 是有效的，但程序还是会 `panic` 异常：

```shell
panic: interface conversion: interface {} is int8, not int16

goroutine 1 [running]:
main.read(0x10592e0, 0x10be5c1)
main.go:10 +0x7d
main.main()
main.go:5 +0x39
exit status 2
```

让我们生成 [asm](https://golang.org/doc/asm) 代码，以便查看 Go 执行的检查：

![生成汇编码过程中检查空接口类型](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-understand-the-empty-interface/asm-code.png)

有以下几个步骤：
<!-- TODO 很别扭，自己也没有完全理解 -->
* 步骤 1：比较（指令`CMPQ`）`int16`类型（加载指令`LEAQ`(Load Effective Address，加载有效地址）到空接口的内部类型（从空接口`MOVQ`的内存段读取 48 字节偏移量的内存的指令）

* step 2：`JNE` 指令（Jump if Not Equal），将跳转到生成的指令，这些指令将在步骤中处理错误 3

* 步骤 3：代码将发生混乱并生成我们之前看到的错误消息

* 步骤 4：这是错误指令的结束。此特定指令由显示指令的错误消息引用：`main.go:10 +0x7d`

在转换原始类型之后，应该从空接口的内部类型进行任意类型转换。这种转换为空接口，然后转换回原始类型会导致程序损耗。让我们运行一些基准测试来简单了解一下。

## 性能

下边是两个基准测试。一个使用结构的副本，另一个使用空接口：

```go
package main_test

import (
    "testing"
)

var x MultipleFieldStructure

type MultipleFieldStructure struct {
    a int
    b string
    c float32
    d float64
    e int32
    f bool
    g uint64
    h *string
    i uint16
}

//go:noinline
func emptyInterface(i interface {}) {
    s := i.(MultipleFieldStructure)
    x = s
}

//go:noinline
func typed(s MultipleFieldStructure) {
    x = s
}

func BenchmarkWithType(b *testing.B) {
    s := MultipleFieldStructure{a: 1, h: new(string)}
    for i := 0; i < b.N; i++ {
        typed(s)
    }
}

func BenchmarkWithEmptyInterface(b *testing.B) {
    s := MultipleFieldStructure{a: 1, h: new(string)}
    for i := 0; i < b.N; i++ {
        emptyInterface(s)
    }
}
```

结果：

```shell
BenchmarkWithType-8               300000000           4.24 ns/op
BenchmarkWithEmptyInterface-8      20000000           60.4 ns/op
```

与复制结构相比，将双重转换（类型转换为空接口然后再转换回类型）多消耗 55 纳秒以上的时间。如果结构中字段的数量增加，时间还会增加：

```shell
BenchmarkWithType-8             100000000         17 ns/op
BenchmarkWithEmptyInterface-8    10000000        153 ns/op
```

但是，有一个好的解决方案是：使用指针并转换回相同的结构指针。转换看起来像下边这样：

```go
func emptyInterface(i interface {}) {
    s := i.(*MultipleFieldStructure)
    y = s
}
```

和上边相比，结果差异很大：

```shell
BenchmarkWithType-8                 2000000000          2.16 ns/op
BenchmarkWithEmptyInterface-8       2000000000          2.02 ns/op
```

关于像 `int` 或 `string` 这样的基础类型，性能略有不同

```shell
BenchmarkWithTypeInt-8              2000000000          1.42 ns/op
BenchmarkWithEmptyInterfaceInt-8    1000000000          2.02 ns/op
BenchmarkWithTypeString-8           1000000000          2.19 ns/op
BenchmarkWithEmptyInterfaceString-8  50000000           30.7 ns/op
```

<!-- TODO with parsimony 节约成本 地使用？ 还是会造成 节约成本地结果？ 从基准测试看来，空结构体并不太好呀-->
如果使用得当，在大多数情况下，空接口应该会对应用程序的性能产生真正的影响:

---

via: <https://medium.com/a-journey-with-go/go-understand-the-empty-interface-2d9fc1e5ec72>

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[TomatoAres](https://github.com/TomatoAres)
校对：[校对者 ID](https://github.com/校对者 ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
