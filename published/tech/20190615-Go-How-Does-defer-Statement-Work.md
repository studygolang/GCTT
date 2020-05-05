首发于：https://studygolang.com/articles/28431

# Go：defer 语句如何工作

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://raw.githubusercontent.com/studygolang/gctt-images2/master/how-does-defer-statement-work/1.png)

ℹ️ *这篇文章基于 Go 1.12。*

[`defer` 语句](https://golang.org/ref/spec#Defer_statements)是在函数返回前执行一段代码的便捷方法，如 [Golang 规范](https://golang.org/ref/spec#Defer_statements)所描述：

> 延迟函数（ deferred functions ）在所在函数返回前，以与声明相反的顺序立即被调用

以下是 LIFO (后进先出)实现的例子：

```go
func main() {
   defer func() {
      println(`defer 1`)
   }()
   defer func() {
      println(`defer 2`)
   }()
}
defer 2 <- 后进先出
defer 1
```

来看一下内部的实现，然后再看一个更复杂的案例。

## 内部实现

Go 运行时（runtime）使用一个**链表**来实现 LIFO。实际上，一个 defer 结构体持有一个指向下一个要被执行的 defer 结构体的指针：

```go
type _defer struct {
   siz     int32
   started bool
   sp      uintptr
   pc      uintptr
   fn      *funcval
   _panic  *_panic
   link    *_defer // 下一个要被执行的延迟函数
```

当一个新的 defer 方法被创建的时候，它被附加到当前的 Goroutine 上，然后之前的 defer 方法作为下一个要执行的函数被链接到新创建的方法上：

```go
func newdefer(siz int32) *_defer {
   var d *_defer
   gp := getg() // 获取当前 goroutine
   [...]
   // 延迟列表现在被附加到新的 _defer 结构体
   d.link = gp._defer
   gp._defer = d // 新的结构现在是第一个被调用的
   return d
}
```

现在，后续调用会从栈的顶部依次出栈延迟函数：

```go
func deferreturn(arg0 uintptr) {
   gp := getg() // 获取当前 goroutine
   d:= gp._defer // 拷贝延迟函数到一个变量上
   if d == nil { // 如果不存在延迟函数就直接返回
      return
   }
   [...]
   fn := d.fn // 获取要调用的函数
   d.fn = nil // 重置函数
   gp._defer = d.link // 把下一个 _defer 结构体依附到 Goroutine 上
   freedefer(d) // 释放 _defer 结构体
   jmpdefer(fn, uintptr(unsafe.Pointer(&arg0))) // 调用该函数
}
```

如我们所见，并没有循环地去调用延迟函数，而是一个接一个地出栈。这一行为可以通过生成[汇编](https://golang.org/doc/asm)代码得到验证：

```asm
// 第一个延迟函数
0x001d 00029 (main.go:6)   MOVL   $0, (SP)
0x0024 00036 (main.go:6)   PCDATA $2, $1
0x0024 00036 (main.go:6)   LEAQ   "".main.func1·f(SB), AX
0x002b 00043 (main.go:6)   PCDATA $2, $0
0x002b 00043 (main.go:6)   MOVQ   AX, 8(SP)
0x0030 00048 (main.go:6)   CALL   runtime.deferproc(SB)
0x0035 00053 (main.go:6)   TESTL  AX, AX
0x0037 00055 (main.go:6)   JNE    117
// 第二个延迟函数
0x0039 00057 (main.go:10)  MOVL   $0, (SP)
0x0040 00064 (main.go:10)  PCDATA $2, $1
0x0040 00064 (main.go:10)  LEAQ   "".main.func2·f(SB), AX
0x0047 00071 (main.go:10)  PCDATA $2, $0
0x0047 00071 (main.go:10)  MOVQ   AX, 8(SP)
0x004c 00076 (main.go:10)  CALL   runtime.deferproc(SB)
0x0051 00081 (main.go:10)  TESTL  AX, AX
0x0053 00083 (main.go:10)  JNE    101
// main 函数结束
0x0055 00085 (main.go:18)  XCHGL  AX, AX
0x0056 00086 (main.go:18)  CALL   runtime.deferreturn(SB)
0x005b 00091 (main.go:18)  MOVQ   16(SP), BP
0x0060 00096 (main.go:18)  ADDQ   $24, SP
0x0064 00100 (main.go:18)  RET
0x0065 00101 (main.go:10)  XCHGL  AX, AX
0x0066 00102 (main.go:10)  CALL   runtime.deferreturn(SB)
0x006b 00107 (main.go:10)  MOVQ   16(SP), BP
0x0070 00112 (main.go:10)  ADDQ   $24, SP
0x0074 00116 (main.go:10)  RET
```

`deferproc` 方法被调用了两次，并且内部调用了 `newdefer` 方法，我们之前已经看到该方法将我们的函数注册为延迟函数。之后，在函数的最后，在 `deferreturn` 函数的帮助下，延迟方法会被一个接一个地调用。

Go 标准库向我们展示了结构体 `_defer` 同样链接了一个 `_panic *_panic` 属性。来通过另一个例子看下它在哪里会起作用。

## 延迟和返回值

如规范所描述，延迟函数访问返回的结果的唯一方法是使用[命名返回参数](https://golang.org/ref/spec#Function_types)：

> 如果延迟函数是一个[匿名函数（ function literal ）](https://golang.org/ref/spec#Function_literals)，并且所在函数存在[命名返回参数](https://golang.org/ref/spec#Function_types)，同时该命名返回参数在匿名函数的作用域中，匿名函数可能会在返回参数返回前访问并修改它们。

这里有个例子：

```go
func main() {
   fmt.Printf("with named param, x: %d\n", namedParam())
   fmt.Printf("without named param, x: %d\n", notNamedParam())
}
func namedParam() (x int) {
   x = 1
   defer func() { x = 2 }()
   return x
}

func notNamedParam() (int) {
   x := 1
   defer func() { x = 2 }()
   return x
}
with named param, x: 2
without named param, x: 1
```

确实就像这篇“[defer, panic 和 recover](https://blog.golang.org/defer-panic-and-recover)”博客所描述的一样，一旦确定这一行为，我们可以将其与 recover 函数混合使用：

> **recover 函数** 是一个用于重新获取对恐慌（panicking）goroutine 控制的内置函数。recover 函数仅在延迟函数内部时才有效。

如我们所见，`_defer` 结构体链接了一个 `_panic` 属性，该属性在 panic 调用期间被链接。

```go
func gopanic(e interface{}) {
   [...]
   var p _panic
   [...]
   d := gp._defer // 当前附加的 defer 函数
   [...]
   d._panic = (*_panic)(noescape(unsafe.Pointer(&p)))
   [...]
}
```

确实，在发生 panic 的情况下，调用延迟函数之前会调用 `gopanic` 方法：

```ams
0x0067 00103 (main.go:21)   CALL   runtime.gopanic(SB)
0x006c 00108 (main.go:21)  UNDEF
0x006e 00110 (main.go:16)  XCHGL  AX, AX
0x006f 00111 (main.go:16)  CALL   runtime.deferreturn(SB)
```

这里是一个 recover 函数利用命名返回参数的例子：

```go
func main() {
   fmt.Printf("error from err1: %v\n", err1())
   fmt.Printf("error from err2: %v\n", err2())
}

func err1() error {
   var err error

   defer func() {
      if r := recover(); r != nil {
         err = errors.New("recovered")
      }
   }()
   panic(`foo`)

   return err
}

func err2() (err error) {
   defer func() {
      if r := recover(); r != nil {
         err = errors.New("recovered")
      }
   }()
   panic(`foo`)

   return err
}
error from err1: <nil>
error from err2: recovered
```

两者的结合是我们可以正常使用 recover 函数将我们希望的 error 返回给调用方。
作为这篇关于延迟函数的文章的总结，让我们来看看延迟函数的提升。

## 性能提升

[Go 1.8](https://golang.org/doc/go1.8#defer)是提升 defer 的最近的一个版本（译者注：目前 Go 1.14 才是提升 defer 性能的最近的一个版本），我们可以通过运行 Go 的基准测试来看到这些提升（在 1.7 和 1.8 之间进行对比）：
```
name         old time/op  new time/op  delta
Defer-4      99.0ns ± 9%  52.4ns ± 5%  -47.04%  (p=0.000 n=9+10)
Defer10-4    90.6ns ± 13%  45.0ns ± 3%  -50.37%  (p=0.000 n=10+10)
```

这样的提升得益于[这个提升分配方式的 CL ](https://go-review.googlesource.com/c/go/+/29656/)，避免了栈的增长。

不带参数的 defer 语句避免内存拷贝也是一个优化。下面是带参数和不带参数的延迟函数的基准测试：

```
name     old time/op  new time/op  delta
Defer-4  51.3ns ± 3%  45.8ns ± 1%  -10.72%  (p=0.000 n=10+10)
```

由于第二个优化，现在速度也提高了 10%。

---

via: https://medium.com/a-journey-with-go/go-how-does-defer-statement-work-1a9492689b6e

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[dust347](https://github.com/dust347)
校对：[@unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
