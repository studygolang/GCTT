已发布：https://studygolang.com/articles/12370

# 在 Golang 中针对 int64 类型优化 abs() 

Go 语言没有内置 `abs()` 标准函数来计算整数的绝对值，这里的绝对值是指负数、正数的非负表示。

我最近为了解决 [Advent of Code 2017](http://adventofcode.com/2017/about) 上边的 [Day 20](http://adventofcode.com/2017/day/20) 难题，自己实现了一个 `abs()` 函数。如果你想学点新东西或试试身手，可以去一探究竟。

Go 实际上已经在 `math` 包中实现了 `abs()` :  [math.Abs](https://golang.org/pkg/math/#Abs) ，但对我的问题并不适用，因为它的输入输出的值类型都是 `float64`，我需要的是 `int64`。通过参数转换是可以使用的，不过将 `float64` 转为 `int64` 会产生一些开销，且转换值很大的数会发生截断，这两点都会在文章说清楚。

帖子 [Pure Go math.Abs outperforms assembly version](https://groups.google.com/forum/#!topic/golang-dev/nP5mWvwAXZo) 讨论了针对浮点数如何优化 `math.Abs`，不过这些优化的方法因底层编码不同，不能直接应用在整型上。

文章中的源码和测试用例在 [cavaliercoder/go-abs](https://github.com/cavaliercoder/go-abs)

## 类型转换 VS 分支控制的方法

对我来说取绝对值最简单的函数实现是：输入参数 n 大于等于 0 直接返回 n，小于零则返回 -n（负数取反为正），这个取绝对值的函数依赖分支控制结构来计算绝对值，就命名为：`abs.WithBranch`

```go
package abs

func WithBranch(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
```

成功返回 n 的绝对值，这就是 [Go v1.9.x](https://github.com/golang/go/blob/release-branch.go1.9/src/math/abs.go)  `math.Abs` 对 float64 取绝对值的实现。不过当进行类型转换（int64 to float64）再取绝对值时，1.9.x 是否做了改进？我们可以验证一下：

```go
package abs

func WithStdLib(n int64) int64 {
	return int64(math.Abs(float64(n)))
}
```

上边的代码中，将 n 先从 `int64` 转成 `float64`，通过 `math.Abs` 取到绝对值后再转回 `int64`，多次转换显然会造成性能开销。可以写一个基准测试来验证一下：

```console
$ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/cavaliercoder/abs
BenchmarkWithBranch-8           2000000000               0.30 ns/op
BenchmarkWithStdLib-8           2000000000               0.79 ns/op
PASS
ok      github.com/cavaliercoder/abs    2.320s
```

测试结果：0.3 ns/op， `WithBranch` 要快两倍多，它还有一个优势：在将 int64 的大数转化为 IEEE-754 标准的 float64 不会发生截断（丢失超出精度的值）

举个例子：`abs.WithBranch(-9223372036854775807)` 会正确返回 9223372036854775807。但 `WithStdLib(-9223372036854775807)` 则在类型转换区间发生了溢出，返回 -9223372036854775808，在大的正数输入时， `WithStdLib(9223372036854775807)` 也会返回不正确的负数结果。

不依赖分支控制的方法取绝对值的方法对有符号整数显然更快更准，不过还有更好的办法吗？

我们都知道不依赖分支控制的方法的代码破坏了程序的运行顺序，即 [pipelining processors](http://euler.mat.uson.mx/~havillam/ca/CS323/0708.cs-323007.html) 无法预知程序的下一步动作。

## 与不依赖分支控制的方法不同的方案

[Hacker’s Delight](https://books.google.com.au/books?id=VicPJYM0I5QC&lpg=PA18&ots=2o-SROAuXq&dq=hackers%20delight%20absolute&pg=PA18#v=onepage&q=hackers%20delight%20absolute&f=false) 第二章介绍了一种无分支控制的方法，通过 [Two’s Complement](https://www.cs.cornell.edu/~tomf/notes/cps104/twoscomp.html) 计算有符号整数的绝对值。

为计算 x 的绝对值，先计算 `x >> 63` ，即 x 右移 63 位（获取最高位符号位），如果你对熟悉无符号整数的话， 应该知道如果 x 是负数则 y 是 1，否者 y 为 0

接着再计算 `(x ⨁ y) - y` ：x 与 y 异或后减 y，即是 x 的绝对值。

可以直接使用高效的汇编实现，代码如下：

```go
func WithASM(n int64) int64
```

```asm
// abs_amd64.s
TEXT ·WithASM(SB),$0
	MOVQ    n+0(FP), AX     // copy input to AX
	MOVQ    AX, CX          // y ← x
	SARQ    $63, CX         // y ← y >> 63
	XORQ    CX, AX          // x ← x ⨁ y
	SUBQ    CX, AX          // x ← x - y
	MOVQ    AX, ret+8(FP)   // copy result to return value
	RET
```

我们先命名这个函数为 `WithASM`，分离命名与实现，函数体使用 [Go 的汇编](https://golang.org/doc/asm) 实现，上边的代码只适用于 AMD64 架构的系统，我建议你的文件名加上 `_amd64.s` 的后缀。

`WithASM` 的基准测试结果：

```shell
$ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/cavaliercoder/abs
BenchmarkWithBranch-8           2000000000               0.29 ns/op
BenchmarkWithStdLib-8           2000000000               0.78 ns/op
BenchmarkWithASM-8              2000000000               1.78 ns/op
PASS
ok      github.com/cavaliercoder/abs    6.059s
```

这就比较尴尬了，这个简单的基准测试显示无分支控制结构高度简洁的代码跑起来居然很慢：1.78 ns/op，怎么会这样呢?

## 编译选项

我们需要知道 Go 的编译器是怎么优化执行 `WithASM` 函数的，编译器接受 `-m` 参数来打印出优化的内容，在 `go build` 或 `go test` 中加上 `-gcflags=-m` 使用：

运行效果：

```console
$ go tool compile -m abs.go
# github.com/cavaliercoder/abs
./abs.go:11:6: can inline WithBranch
./abs.go:21:6: can inline WithStdLib
./abs.go:22:23: inlining call to math.Abs
```

对于我们这个简单的函数，Go 的编译器支持  [function inlining](https://github.com/golang/go/wiki/CompilerOptimizations#function-inlining)，函数内联是指在调用我们函数的地方直接使用这个函数的函数体来代替。举个例子：

```go
package main

import (
	"fmt"
	"github.com/cavaliercoder/abs"
)

func main() {
	n := abs.WithBranch(-1)
	fmt.Println(n)
}
```

 实际上会被编译成：

```go
package main

import "fmt"

func main() {
	n := -1
	if n < 0 {
		n = -n
	}
	fmt.Println(n)
}
```

根据编译器的输出，可以看出 `WithBranch` 和 `WithStdLib` 在编译时候被内联了，但是 `WithASM` 没有。对于 `WithStdLib`，即使底层调用了 `math.Abs` 但编译时依旧被内联。

因为 `WithASM` 函数没法内联，每个调用它的函数会在调用上产生额外的开销：为 `WithASM` 重新分配栈内存、复制参数及指针等等。

如果我们在其他函数中不使用内联会怎么样？可以写个简单的示例程序：

```go
package abs

//go:noinline
func WithBranch(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
```

重新编译，我们会看到编译器优化内容变少了：

```go
$ go tool compile -m abs.go
abs.go:22:23: inlining call to math.Abs
```

基准测试的结果：

```console
$ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/cavaliercoder/abs
BenchmarkWithBranch-8           1000000000               1.87 ns/op
BenchmarkWithStdLib-8           1000000000               1.94 ns/op
BenchmarkWithASM-8              2000000000               1.84 ns/op
PASS
ok      github.com/cavaliercoder/abs    8.122s
```

可以看出，现在三个函数的平均执行时间几乎都在 1.9 ns/op 左右。

你可能会觉得每个函数的调用开销在 1.5ns 左右，这个开销的出现否定了我们 `WithBranch` 函数中的速度优势。

我从上边学到的东西是， `WithASM` 的性能要优于编译器实现类型安全、垃圾回收和函数内联带来的性能，虽然大多数情况下这个结论可能是错误的。当然，这其中是有特例的，比如提升 [SIMD](https://goroutines.com/asm) 的加密性能、流媒体编码等。

## 只使用一个内联函数

Go 编译器无法内联由汇编实现的函数，但是内联我们重写后的普通函数是很容易的：

```go
package abs

func WithTwosComplement(n int64) int64 {
	y := n >> 63          // y ← x >> 63
	return (n ^ y) - y    // (x ⨁ y) - y
}
```

编译结果说明我们的方法被内联了：

```shell
$ go tool compile -m abs.go
...
abs.go:26:6: can inline WithTwosComplement
```

但是性能怎么样呢？结果表明：当我们启用函数内联时，性能与 `WithBranch` 很相近了：

```shell
$ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/cavaliercoder/abs
BenchmarkWithBranch-8               2000000000               0.29 ns/op
BenchmarkWithStdLib-8               2000000000               0.79 ns/op
BenchmarkWithTwosComplement-8       2000000000               0.29 ns/op
BenchmarkWithASM-8                  2000000000               1.83 ns/op
PASS
ok      github.com/cavaliercoder/abs    6.777s
```

现在函数调用的开销消失了，`WithTwosComplement` 的实现要比 `WithASM` 的实现好得多。来看看编译器在编译 `WithASM` 时做了些什么？

使用 `-S` 参数告诉编译器打印出汇编过程：

```shell
$ go tool compile -S abs.go
...
"".WithTwosComplement STEXT nosplit size=24 args=0x10 locals=0x0
				0x0000 00000 (abs.go:26)        TEXT    "".WithTwosComplement(SB), NOSPLIT, $0-16
				0x0000 00000 (abs.go:26)        FUNCDATA        $0, gclocals·f207267fbf96a0178e8758c6e3e0ce28(SB)
				0x0000 00000 (abs.go:26)        FUNCDATA        $1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
				0x0000 00000 (abs.go:26)        MOVQ    "".n+8(SP), AX
				0x0005 00005 (abs.go:26)        MOVQ    AX, CX
				0x0008 00008 (abs.go:27)        SARQ    $63, AX
				0x000c 00012 (abs.go:28)        XORQ    AX, CX
				0x000f 00015 (abs.go:28)        SUBQ    AX, CX
				0x0012 00018 (abs.go:28)        MOVQ    CX, "".~r1+16(SP)
				0x0017 00023 (abs.go:28)        RET
...
```

编译器在编译 `WithASM` 和 `WithTwosComplement` 时，做的事情太像了，编译器在这时才有正确配置和跨平台的优势，可加上 `GOARCH=386` 选项再次编译生成兼容 32 位系统的程序。

最后关于内存分配，上边所有函数的实现都是比较理想的情况，我运行 `go test -bench=. -benchme`，观察对每个函数的输出，显示并没有发生内存分配。

### 总结

`WithTwosComplement` 的实现方式在 Go 中提供了较好的可移植性，同时实现了函数内联、无分支控制的代码、零内存分配与避免类型转换导致的值截断。基准测试没有显示出无分支控制比有分支控制的优势，但在理论上，无分支控制的代码在多种情况下性能会更好。

最后，我对 int64 的 abs 实现如下：

```go
func abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}
```

---

via：http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html

作者：[Ryan Armstrong](https://github.com/cavaliercoder)
译者：[wuYinBest](https://github.com/wuYinBest/) 
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，
[Go 中文网](https://studygolang.com/) 荣誉推出