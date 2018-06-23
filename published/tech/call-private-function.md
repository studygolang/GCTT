已发布：https://studygolang.com/articles/12134

# 在 golang 中如何调用私有函数（绑定隐藏的标识符）

2016 年 4 月 28 日

名字在 golang 中的重要性和在其他任何一种语言是一样的。他们甚至含有语义的作用：在一个包的外部某个名字的可见性是由这个名字首字母是否是大写来决定的。

有时为了更好的组织代码或者在其他包使用某些隐藏的函数时需要克服这种限制。

在过去美好的日子，有 2 种实现方式，它们能绕过编译器的检查：不能引用未导出的名称 pkg.symbol ：

- 旧的方式，现在已经不再使用 - 汇编级隐式连接到所需符号，称为 assembly stubs ，详见 [go runtime, os/signal: use //go:linkname instead of assembly stubs to get access to runtime functions](https://groups.google.com/forum/#!topic/golang-codereviews/J0HK9GLc76M) 。
- 现行的方式 - go 编译器通过 go:linkname 支持名称重定向，引用于 11.11.14 [ dev.cc code review 169360043: cmd/gc: changes for removing runtime C code (issue 169360043 by r…@golang.org)](https://groups.google.com/forum/#!topic/golang-codereviews/5Ps_El_RpNE) ，在 github.com 的 issue 上有可以找到 [ cmd/compile: “missing function body” error when using the //go:linkname compiler directive #15006](https://github.com/golang/go/issues/15006) 。

用这些技巧我曾设法绑定 golang 运行时调度器相关的函数用以减少过度使用 go 的协程和内部锁机制导致的 gc 停顿。

## 使用 assembly stubs

想法很简单 - 为需要的标识符提供直接跳转汇编指令 stubs 。链接器并不知道标识符是否已导出。

详见旧版的代码 src/os/signal/sig.s ：

```go
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

而 signal_unix.go 的绑定如下:

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

## 使用 go:linkname

为了使用这种方法，代码中必须 `import _ "unsafe"` 包。为了解决 go 编译器 `-complete` 参数的限制，一种可能的方法是在 main 包目录加一个空的汇编 stub 文件以禁用编译器的检查。

详见 os/signal/sig.s：

```go
// The runtime package uses //go:linkname to push a few functions into this
// package but we still need a .s file so the Go tool does not pass -complete
// to the go tool compile so the latter does not complain about Go functions
// with no bodies.
```

这个指令的格式是 `//go:linkname localname linkname`。使用这种方法可以将新的标识符链接（导出）或绑定到已存在的标识符（导入）。

## 用 go:linkname 导出

在 runtime/proc.go 中一个函数的实现

```go
...

//go:linkname sync_runtime_doSpin sync.runtime_doSpin
//go:nosplit
func sync_runtime_doSpin() {
	procyield(active_spin_cnt)
}
```

这里明确地向编译器指示，将另一个名字添加到 sync 包的 runtime_doSpin 函数代码中。并且 sync 包简单的重用了在 sync/runtime.go 中的代码：

```go
package sync

import "unsafe"

...

// runtime_doSpin does active spinning.
func runtime_doSpin()
```

## 用 go:linkname 导入

在 net/parse.go 中有一个很好的例子：

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

使用这种技巧的方法：

1. import _ "unsafe" 包。
2. 提供一个没有函数体的函数，比如： func byteIndex(s string, c byte) int
3. 在定义函数前，正确的放上 //go:linkname 指令，例如 //go:linkname byteIndex strings.IndexByte,byteIndex 是本地名称，strings.IndexByte 是远程名称。
4. 提供 .s 后缀的 stub 文件，以便编译器绕过 -complete 的检查，允许不完整的函数定义（译注：指没有函数体）。

## 例子 goparkunlock

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

## 源码

可在这里获取 [https://github.com/sitano/gsysint](https://github.com/sitano/gsysint) 。

## 相关文章

- [Docker Windows install instructions on the state of 4 August 2016 04 Aug 2016](https://sitano.github.io/2016/08/04/docker-win/)
- [PowerShell ducklish typed 25 Apr 2016](https://sitano.github.io/2016/04/25/powershell-ducklish/)
- [Approach into strong typed configuration management DSL with FAKE, F#, WinRM and PowerShell 15 Mar 2016](https://sitano.github.io/2016/03/15/powershell-winrm-fake/)

---

via: https://sitano.github.io/2016/04/28/golang-private/

作者：[JohnKoepi](https://twitter.com/JohnKoepi)
译者：[kekemuyu](https://github.com/kekemuyu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
