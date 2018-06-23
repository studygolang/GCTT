已发布：https://studygolang.com/articles/13035

# 使用 defer 的运行时开销

在 Go 语言中有一个特殊的关键字 `defer`。对于它更多的介绍请看[这里](https://blog.golang.org/defer-panic-and-recover)。`defer` 语句会把一个函数追加到函数调用列表。这个列表会在函数返回的时候依次调用。`defer` 常用来进行各种清理操作。

但是 `defer` 本身是有开销的。使用 Go 的基准测试工具我们可以量化这种开销。

下面两个函数做同样的工作。一个使用 `defer` 语句而另一个不使用：

```go
package main
func doNoDefer(t *int) {
	func() {
		*t++
	}()
}
func doDefer(t *int) {
	defer func() {
		*t++
	}()
}
```

基准测试代码：

```go
package main
import (
	"testing"
)
func BenchmarkDeferYes(b *testing.B) {
	t := 0
	for i := 0; i < b.N; i++ {
		doDefer(&t)
	}
}
func BenchmarkDeferNo(b *testing.B) {
	t := 0
	for i := 0; i < b.N; i++ {
		doNoDefer(&t)
	}
}
```

在一个 8 核的谷歌云主机上运行基准测试：

```
⇒ go test -v -bench BenchmarkDefer -benchmem
goos: linux
goarch: amd64
pkg: cmd
BenchmarkDeferYes-8  20000000   62.4 ns/op  0 B/op  0 allocs/op
BenchmarkDeferNo-8   500000000  3.70 ns/op  0 B/op  0 allocs/op
```

和预想的一样，这些函数都没有额外分配任何内存。但是 `doDefer` 的开销要比 `doNoDefer` 高 16 倍之多。我们需要借助反汇编代码来了解为什么 `defer` 的开销如此之大。

反汇编代码的函数调用部分 `doDefer` 和 `doNoDefer` 是相同的。

```
main.go:10   MOVQ 0x8(SP), AX
main.go:11   MOVQ 0(AX), CX
main.go:11   INCQ CX
main.go:11   MOVQ CX, 0(AX)
main.go:12   RET
```

`doNoDefer` 先初始化必要的注册工作然后调用 `main.doNoDefer.func1`。

```
TEXT main.doNoDefer(SB) main.go
main.go:3  MOVQ FS:0xfffffff8, CX
main.go:3  CMPQ 0x10(CX), SP
main.go:3  JBE 0x450b65
main.go:3  SUBQ $0x10, SP
main.go:3  MOVQ BP, 0x8(SP)
main.go:3  LEAQ 0x8(SP), BP
main.go:3  MOVQ 0x18(SP), AX
main.go:6  MOVQ AX, 0(SP)
main.go:6  CALL main.doNoDefer.func1(SB)
main.go:7  MOVQ 0x8(SP), BP
main.go:7  ADDQ $0x10, SP
main.go:7  RET
main.go:3  CALL runtime.morestack_noctxt(SB)
main.go:3  JMP main.doNoDefer(SB)
```

`doDefer` 也会先进行必要的注册工作，但是它会额外调用几个函数：第一个是 `runtime.deferproc`，它用来设置需要调用的延迟函数。第二个是 `runtime.deferreturn`，它会自动调用每个 `defer` 语句。

```
TEXT main.doDefer(SB) main.go
main.go:9    MOVQ FS:0xfffffff8, CX
main.go:9    CMPQ 0x10(CX), SP
main.go:9    JBE 0x450bd3
main.go:9    SUBQ $0x20, SP
main.go:9    MOVQ BP, 0x18(SP)
main.go:9    LEAQ 0x18(SP), BP
main.go:9    MOVQ 0x28(SP), AX
main.go:12   MOVQ AX, 0x10(SP)
main.go:10   MOVL $0x8, 0(SP)
main.go:10   LEAQ 0x218e3(IP), AX
main.go:10   MOVQ AX, 0x8(SP)
main.go:10   CALL runtime.deferproc(SB)
main.go:10   TESTL AX, AX
main.go:10   JNE 0x450bc3
main.go:13   NOPL
main.go:13   CALL runtime.deferreturn(SB)
main.go:13   MOVQ 0x18(SP), BP
main.go:13   ADDQ $0x20, SP
main.go:13   RET
main.go:10   NOPL
main.go:10   CALL runtime.deferreturn(SB)
main.go:10   MOVQ 0x18(SP), BP
main.go:10   ADDQ $0x20, SP
main.go:10   RET
main.go:9    CALL runtime.morestack_noctxt(SB)
main.go:9    JMP main.doDefer(SB)
```

`deferproc` 和 `deferreturn` 都是比较复杂的函数，它们会在进入和退出函数时进行一系列的配置和计算。所以，不要在热代码中使用 `defer` 关键字，因为它的开销很大的而且很难被侦测到。

---

via: https://medium.com/i0exception/runtime-overhead-of-using-defer-in-go-7140d5c40e32

作者：[Aniruddha](https://medium.com/@i0exception)
译者：[saberuster](https://github.com/saberuster)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
