# Go: 内置函数优化

![由Renee French创作的原始Go Gopher作品，为“ Go的旅程”创作的插图。](https://github.com/studygolang/gctt-images2/blob/master/20200813-Go-Built-in-Functions-Optimizations/Illustration.png?raw=true)

ℹ️  这篇文章基于 Go 1.13。

Go 语言提供内置函数来辅助开发者处理 channel，slice，或者 map。一些内置函数有着像样的内部实现，比如 `make()`，而有的内置函数完全没有实现，而是由编译器所管理。让我们一起分析一些内置函数，来理解 Go 如何处理它们。

## Slices
如果可以事先知道的话，Go 有时可以把一个在运行时完成的函数调用替换为它的结果。来看一个使用切片的例子：

```go
func main() {
   s := make([]int, 0, 6)
   s = append(s, 12)
   s = append(s, 34)

   l := len(s)
   println("the length is ", l)

   c := cap(s)
   println("the capacity is ", c)
}
```

函数 `len` 和 `cap` 实际上没有具体实现。编译器能够跟踪在切片上所做的更改，并将长度或者容量函数替换为可以代表它们的常量。这是[汇编](https://golang.org/doc/asm)代码表示：

![](https://github.com/studygolang/gctt-images2/blob/master/20200813-Go-Built-in-Functions-Optimizations/replace-length-or-capacity%20function-constant.png?raw=true)

但是，编译器不可能永远确定切片的大小。在这种情况下，比如，切片是一个函数的参数，这个函数没有指定其大小。这是一个例子：

```go
func main() {
   s := make([]int, 0, 6)
   s = append(s, 12, 34)
   getLength(s)
}

//go:noinline
func getLength(s []int) {
   l := len(s)
   println("the length is", l)
}
```

这是生成的指令：

![](https://github.com/studygolang/gctt-images2/blob/master/20200813-Go-Built-in-Functions-Optimizations/1_9eu8BEzj2ATNR2tmdsCMfg.png?raw=true)

由于 Go 无法了解 `getLength()` 方法被如何使用，只能直接从内存里面读取切片的长度。同样的行为也会引用在切片的容量上。

*Go 通常会指向切片的长度，防止不良的内存访问，比如越界读取。更多信息，建议阅读我的文章“[Go：边界检查保证内存安全](https://medium.com/a-journey-with-go/go-memory-safety-with-bounds-check-1397bef748b5)”。*

## Unsafe
`unsafe` 包同样暴露了没有任何实现的函数。由 Go 标准库提供的定义原型，仅仅用作说明文档：

![Example of the documentation of the unsafe package](https://github.com/studygolang/gctt-images2/blob/master/20200813-Go-Built-in-Functions-Optimizations/Example-of-the-documentation-of-the-unsafe-package.png?raw=true)

这是这个包的例子：

```go
type T1 struct {
   a int64
   b bool
}

func main() {
   t1 := T1{}

   println("size of the struct:", unsafe.Sizeof(t1))
   println("alignment of the struct:", unsafe.Alignof(t1))
}
Output:
size of the struct: 16
alignment of the struct: 8
```

再次，编译器有着关于结构的足够信息，可以直接将值写为常量：

![](https://github.com/studygolang/gctt-images2/blob/master/20200813-Go-Built-in-Functions-Optimizations/write-the-values-as-constants-directly.png?raw=true)

*Go 语言提供了更多的内置函数，比如 `make` 或 `copy`。关于 `copy` 函数的更多细节信息，建议阅读我的文章“Go：切片以及内存管理”，该文章详细阐述了优化细节。关于在 map 上的优化，建议阅读“[Go：根据代码设计 map——第二部分](https://medium.com/a-journey-with-go/go-map-design-by-code-part-ii-50d111557c08)” 来深入了解。*

---
via: https://medium.com/a-journey-with-go/go-built-in-functions-optimizations-70c5abb3a680

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[dust347](https://github.com/dust347)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
