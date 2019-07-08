首发于：https://studygolang.com/articles/21757

# Go: 什么是 unsafe 包

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/what-is-the-unsafe-package/unsafe_1.png)

ℹ️本文基于 Go 1.12。

看到 unsafe 这个名称，我们应该尽量避免使用它。想要知道使用 unsafe 包可能产生不安全的原因，我们首先来看看官方文档的描述：

> unsafe 包含有违背 Go 类型安全的操作。
>
> 导入 unsafe 包可能会使程序不可移植，并且不受 Go 1 兼容性指南的保护。

因此，该名称被用作提示 unsafe 包可能带来 Go 类型的不安全性。现在我们来深入探讨一下文档中提到的两点。

## 类型安全

在 Go 中，每个变量都有一个类型，可以在分配给另一个变量之前转换为另一个类型。在此转换期间，Go 会对此数据执行转换，以适应请求的类型。来看下面这个例子：

```go
var i int8 = -1 // -1 二进制表示 : 11111111
var j = int16(i) // -1 二进制表示 : 11111111 11111111
println(i, j) // -1 -1
```

`unsafe` 包让我们可以直接访问此变量的内存，并将原始二进制值存储在此地址中。当我们想绕过类型约束时，我们可以根据需要使用它：

```go
var k uint8 = *(*uint8)(unsafe.Pointer(&i))
println(k) // 255 is the uint8 value for the binary 11111111
```

现在，原始值被解释为 uint8，而没有使用先前声明的类型（int8）。如果你有兴趣深入了解此主题，我建议你阅读我关于[使用 Go 进行 Cast 和 Conversion](https://medium.com/@blanchon.vincent/go-cast-vs-conversion-by-example-26e0ef3003f0) 的文章。

## Go 1 兼容性指南

[Go 1 的指南](https://golang.org/doc/go1compat#expectations) 清楚地解释了如果他们修改了底层的实现，`unsafe` 包的使用可能导致你的代码无法运行：

> 导入 `unsafe` 软件包可能取决于 Go 实现的内部属性。 我们保留做导致奔溃的修改的权利。

我们应该记住，在 Go 1 中，内部实现可能会发生变化，我们可能会遇到像这个[Github issue](https://github.com/golang/go/issues/16769) 中类似的问题，两个版本之间的行为略有变化。但是，Go 标准库在许多地方也使用了 `unsafe` 包。

## 在 Go 的 reflect 包中使用

`reflection` 包是最常用的包之一。反射基于空接口包含的内部数据。要读取数据，Go 只是将我们的变量转换为空接口，并通过将与空接口的内部表示匹配的结构和指针地址处的内存映射来读取它们：

```go
func ValueOf(i interface{}) Value {
	[...]
	return unpackEface(i)
}
// unpackEface converts the empty interface i to a Value.
func unpackEface(i interface{}) Value {
	e := (*emptyInterface)(unsafe.Pointer(&i))
	[...]
}
```

变量 `e` 现在包含有关值的所有信息，例如类型或是否已导出值。反射还使用 `unsafe` 包通过直接更新内存中的值来修改反射变量的值，如前所述。

## 在 Go 的 sync 包中使用

`unsafe` 包的另一个有趣用法是在 `sync` 包中。如果你不熟悉 `sync` 包，我建议你阅读我的关于[sync.Pool 的设计](https://juejin.im/post/5d006254e51d45776031afe3) 的一篇文章。

这些池通过一段内存在所有 Goroutine/processors 之间共享，所有 Goroutine 都可以通过 `unsafe` 包访问该内存：

```go
func indexLocal(l unsafe.Pointer, i int) *poolLocal {
	lp := unsafe.Pointer(uintptr(l) + uintptr(i)*unsafe.Sizeof(poolLocal{}))
	return (*poolLocal)(lp)
}
```

变量 `l` 是内存段，`i` 是处理器编号。函数 `indexLocal` 只读取此内存段 - 包含 X（处理器数量）`poolLocal` 结构体 - 具有与其读取的索引相关的偏移量。存储指向完整内存段的指针是实现共享池的一种非常轻松的方法。

## 在 Go 的 runtime 包中使用

Go 还在 `runtime` 包中使用了 `unsafe` 包，因为它必须处理内存操作，如堆栈分配或释放堆栈内存。堆栈在其结构中由两个边界表示：

```go
type stack struct {
	lo uintptr
	hi uintptr
}
```

那么 `unsafe` 包将有助于进行操作：

```go
func stackfree(stk stack) {
	[...]
	v := unsafe.Pointer(stk.lo)
	n := stk.hi - stk.lo
	// 然后基于指向堆栈的指针释放内存
	[...]
}
```

如果你想进一步了解堆栈，我建议你阅读我关于[堆栈大小及其管理的文章](https://medium.com/@blanchon.vincent/go-how-does-the-goroutine-stack-size-evolve-447fc02085e5)。

此外，在某些情况下，我们也可以在我们的应用程序中使用此包，例如结构之间的转换。

## unsafe 包对开发人员的用处

`unsafe` 包的一个很好的用法是使用相同的底层数据转换两个不同的结构，这是转换器无法实现的：

```go
type A struct {
	A int8
	B string
	C float32

}

type B struct {
	D int8
	E string
	F float32

}

func main() {
	a := A{A: 1, B: `foo`, C: 1.23}
	//b := B(a) 不能转换 a (type A) 到 type B
	b := *(*B)(unsafe.Pointer(&a))

	println(b.D, b.E, b.F) // 1 foo 1.23
}
```

源码：[https://play.golang.org/p/sjeO9v0T_Fs](https://play.golang.org/p/sjeO9v0T_Fs)

`unsafe` 包中另一个不错的用法是[http://golang-sizeof.tips](http://golang-sizeof.tips)，它可以帮助你理解结构内存对齐的大小。

总之，该软件包非常有趣且功能强大，但是应该谨慎使用。此外，如果你对 `unsafe` 包的将来的修改有建议，你可以在[Github for Go 2](https://github.com/golang/go/issues?utf8=%E2%9C%93&q=is%3Aopen+label%3AGo2+%22unsafe%22+in%3Atitle) 中提 Issue。

---

via: https://medium.com/@blanchon.vincent/go-what-is-the-unsafe-package-d2443da36350

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
