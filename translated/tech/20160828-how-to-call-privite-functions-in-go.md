# 如何调用 Golang 私有函数（绑定到隐藏符号）

变量名称在 Golang 中的重要性和任何其他语言一样。它们甚至具有语义效应：包名称的外部可见性取决于它的第一个字符是否为大写。

有时为了更好地组织代码，或者访问包中对外隐藏的函数，需要突破这个限制。

这些技术 golang 源码中大量使用，这也是相关技术信息的主要来源。而网上相关信息明显不足。

在过去，有两种方法可以绕过编译器检查：`cannot refer to unexported name pkg.symbol`（不能引用未导出的名称）:

* 之前的方法，现在已经不用了——配置隐式链接所需的符号，称为 `assembly stubs`, 比如： [go runtime, os/signal: use //go:linkname instead of assembly stubs to get access to runtime functions](https://groups.google.com/forum/#!topic/golang-codereviews/J0HK9GLc76M).

* 现在实际使用的—— 编译器级别支持通过 `go:linkname` 链接名称重定向 ，详细信息来自 2014 年 11 月 11 日的文章——[dev.cc code review 169360043: cmd/gc: changes for removing runtime C code](https://groups.google.com/forum/#!topic/golang-codereviews/5Ps_El_RpNE)，github 上的这个 issue 也有提到—— [cmd/compile: “missing function body” error when using the //go:linkname compiler directive #15006](https://github.com/golang/go/issues/15006)

使用这些技术，我已经设法绑定到内部 golang 运行时调度相关的功能，以突破使用 `goroutines` 线程中止和内部锁定机制。

## 使用 `assembly stubs`

方法很简单：组装时为需导出的符号打上显示存根（stub）的标记 `JMP`。因为链接器不知道哪些信息需要导出，哪些不需要导出。

比如：旧版本中的 `src/os/signal/sig.s`

```c
// Assembly to get into package runtime without using exported symbols.

// +build amd64 amd64p32 arm arm64 386 ppc64 ppc64le

#include "textflag.h"

#ifdef GOARCH_arm
#define JMP B
#endif
#ifdef GOARCH_ppc64
#define JMP BR
#endif
#ifdef GOARCH_ppc64le
#define JMP BR
#endif

TEXT ·signal_disable(SB),NOSPLIT,$0
    JMP runtime·signal_disable(SB)

TEXT ·signal_enable(SB),NOSPLIT,$0
    JMP runtime·signal_enable(SB)

TEXT ·signal_ignore(SB),NOSPLIT,$0
    JMP runtime·signal_ignore(SB)

TEXT ·signal_recv(SB),NOSPLIT,$0
    JMP runtime·signal_recv(SB)
```

`signal_unix.go` 绑定：

```go
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package signal

import (
    "os"
    "syscall"
)

// In assembly.
func signal_disable(uint32)
func signal_enable(uint32)
func signal_ignore(uint32)
func signal_recv() uint32
```

## 使用 `go:linkname`

为了使用这个功能，源文件必须 `import _"unsafe"` (导入 `unsafe` 包）。 为了跳过 `-complete` go 编译器限制，有一个方法是：将空的程序集存根文件（assembly stub）放在主源文件附近以禁用此检查。

例如： `os/signal/sig.s`:

```go
// The runtime package uses //go:linkname to push a few functions into this
// package but we still need a .s file so the Go tool does not pass -complete
// to the go tool compile so the latter does not complain about Go functions
// with no bodies.
```

这个指令的格式是： `//go:linkname localname linkname`.使用它可以为链接(导出)引入新的符号，或者绑定到现有的符号(导入)。

### `go:linkname` 导出

`runtime/proc.go`中的一个函数实现：

```go
...

//go:linkname sync_runtime_doSpin sync.runtime_doSpin
//go:nosplit
func sync_runtime_doSpin() {
    procyield(active_spin_cnt)
}
```

明确告诉编译器：添加新名称到 `sync` 包中的 `runtime_doSpin`，然后 `sync` 包就在 `sync/runtime.go` 中可以很简单地复用它：

```go
package sync

import "unsafe"

...

// runtime_doSpin does active spinning.
func runtime_doSpin()
```

### `go:linkname` 导入

`net/parse.go` 中有一个很好的例子：

```go
package net

import (
    ...
    _ "unsafe" // For go:linkname
)

...

// byteIndex is strings.IndexByte. It returns the index of the
// first instance of c in s, or -1 if c is not present in s.
// strings.IndexByte is implemented in  runtime/asm_$GOARCH.s
//go:linkname byteIndex strings.IndexByte
func byteIndex(s string, c byte) int
```

为了用这个技术：

1. 导包 `import _“unsafe”`;
2. 给出不带函数体的函数声明，比如： `func byteIndex(s string, c byte) int`;
3. 在函数定义之前向编译器添加一条指令 `//go:linkname`，比如：`//go:linkname byteIndex strings.IndexByte`。 其中，`byteIndex` 是本地名称，`strings.IndexByte` 是远程名称；
4. 提供`.s`文件存根，以允许编译器绕过 `-complete` 检查，以允许部分定义的函数。

## 示例 `goparkunlock`

```go
package main

import (
    _ "unsafe"
    "fmt"
    "runtime/pprof"
    "os"
    "time"
)

// Event types in the trace, args are given in square brackets.
const (
    traceEvGoBlock        = 20 // goroutine blocks [timestamp, stack]
)

type mutex struct {
    // Futex-based impl treats it as uint32 key,
    // while sema-based impl as M* waitm.
    // Used to be a union, but unions break precise GC.
    key uintptr
}

//go:linkname lock runtime.lock
func lock(l *mutex)

//go:linkname unlock runtime.unlock
func unlock(l *mutex)

//go:linkname goparkunlock runtime.goparkunlock
func goparkunlock(lock *mutex, reason string, traceEv byte, traceskip int)

func main() {
    l := &mutex{}
    go func() {
        lock(l)
        goparkunlock(l, "xxx", traceEvGoBlock, 1)
    }()
    for {
        pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
        time.Sleep(time.Second * 1)
    }
}
```

## 资料来源

<https://github.com/sitano/gsysint>

---

via: <https://sitano.github.io/2016/04/28/golang-private/>

作者：[JohnKoepi](https://sitano.github.io/)
译者：[TomatoAres](https://github.com/TomatoAres)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
