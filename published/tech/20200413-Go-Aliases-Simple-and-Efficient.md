首发于：https://studygolang.com/articles/28459

# Go 中使用别名，简单且高效

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Aliases-Simple-and-Efficient/00.png)

ℹ️ 本文基于 Go 1.13。

Go 1.9 版本引入了别名，开发者可以为一个已存在的类型赋其他的名字。这个特性旨在促进大型代码库的重构，这对大型的项目至关重要。在思考了几个月应该以哪种方式让 Go 语言支持别名后，这个特性才被实现。[最初的提案](https://go.googlesource.com/proposal/+/master/design/16339-alias-decls.md)是引入广泛的别名（支持对类型、函数等等赋别名），但这个提案后来被另一个[更简单的别名机制](https://go.googlesource.com/proposal/+/master/design/16339-alias-decls.md)所替代，新提案只关注对类型赋别名，因为对这个特性需求最大的就是类型。只支持对类型赋别名让实现方式变得简单，因为只需要解决最初始的问题就可以了。我们一起来看看这个解决方案。

## 重构

引入别名的最主要的意图是简化对大型代码库的重构。开发者们对旧名字赋一个新的别名，就可以避免破坏已存在代码的兼容性。下面是一个 `docker/cli` 的例子：

```go
package command// Deprecated: Use github.com/docker/cli/cli/streams.In instead
type InStream = streams.In

// Deprecated: Use github.com/docker/cli/cli/streams.Out instead
type OutStream = streams.Out
```

这样不会影响使用 `command.InStream` 的旧代码，而新代码使用新类型 `streams.In` 。

然而，为了完全支持兼容，别名还需要有以下特性：

- 可互相转换的参数类型。新旧类型都可以被作为参数接收。下面是一个 T1 和 T2 可以相互转换的例子：

```go
type T2 struct {}

// T1 is deprecated, please use T2
type T1 = T2

func main() {
   var t T1
   f(t)
}

func f(t T1) {
   // print main.T2
   println(reflect.TypeOf(t).String())
}
```

- 新旧两种类型都可以从空接口转换而来。下面是例子：

```go
type T2 struct {}

// T1 is deprecated, please use T2
type T1 = T2

func main() {
   var t T1
   f(t)
}

func f(t interface{}) {
   t, ok := t.(T1)
   if !ok {
      log.Fatal("t is not a T1 type")
   }
   // print main.T2
   println(reflect.TypeOf(t).String())
}
```

因为新旧类型可以在任何时间相互转换，所以已有代码不会被破坏，可以实现平滑迁移。

## 可读性

别名也可以提高代码的可读性。下面是 Go 标准库和反汇编器包里的例子：

```go
type lookupFunc = func(addr uint64) (sym string, base uint64)
```

一个 lookup 函数接收一个 address 作为参数，返回另一个 address 的 symbol。相比于把这个函数原型作为参数传递给每一个函数，使用这个新别名可读性更好。下面是使用别名作为参数的函数原型：

```go
func disasm_amd64([]byte, uint64, lookupFunc, binary.ByteOrder)
```

`golang.org/x/sys/unix` 中的包通过对经常使用的类型赋别名减少了样板代码量。那些别名声明在单独的文件中：

```go
type Signal = syscall.Signal
type Errno = syscall.Errno
type SysProcAttr = syscall.SysProcAttr
```

声明后，在包中只能引用 `Errno` 而不能再引用 `syscal.Errno`。

## 运行时

现在我们看到了在程序中使用别名的好处，但是我们还不知道在运行时有什么影响。我们来看之前结构体与空接口相互转换的例子：

```go
type T2 struct {}

// T1 is deprecated, please use T2
type T1 = T2

func main() {
   var t T1
   f(t)
}

func f(t interface{}) {
   t, ok := t.(T1)
   if !ok {
      log.Fatal("t is not a T1 type")
   }
   // print main.T2
   println(reflect.TypeOf(t).String())
}
```

虽然从最终的输出看 `t` 的类型是 `T2`，但是这个程序仍然可以把 `t` 转换为 `T1`。我们把代码转换成[汇编](https://golang.org/doc/asm)。下面是输出的部分信息：

```
0x0021 00033 (main.go:19)  MOVQ   "".t+88(SP), AX
0x0026 00038 (main.go:19)  PCDATA $0, $1
0x0026 00038 (main.go:19)  LEAQ   type."".T2(SB), CX
0x002d 00045 (main.go:19)  CMPQ   AX, CX
0x0030 00048 (main.go:20)  JNE    172
```

第一行指令 `MOVQ` 读取空接口的类型并把它储存在寄存器 `AX`。然后 `LEAQ` 把 `T2` 类型加载到寄存器 `CX`，两个寄存器可以做比较。

我们可以看到，代码中的转换是基于 `T2` 而不是 `T1`。别名在编译时被改变，这样可以消除掉我们程序中的开销。

---
via: https://medium.com/a-journey-with-go/go-aliases-simple-and-efficient-8506d93b079e

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
