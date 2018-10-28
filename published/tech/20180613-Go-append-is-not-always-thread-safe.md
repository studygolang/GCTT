首发于：https://studygolang.com/articles/13662

# Go 语言的 append 不总是线程安全的

## 示例问题

我经常看到一些 bug 是由于没有在线程安全下在 slice 上进行 append 而引起的。下面用单元测试来举一个简单的例子。这个测试有两个协程对相同的 slice 进行 append 操作。如果你使用 `-race` flag 来执行这个单元测试，效果更好。

```go
package main

import (
	"sync"
	"testing"
)

func TestAppend(t *testing.T) {
	x := []string{"start"}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		y := append(x, "hello", "world")
		t.Log(cap(y), len(y))
	}()
	go func() {
		defer wg.Done()
		z := append(x, "goodbye", "bob")
		t.Log(cap(z), len(z))
	}()
	wg.Wait()
}
```

现在，让我们稍微修改代码，以给这个名为 `x` 的 slice 在创建是预留一些容量。唯一改动的地方是第 9 行。

```go
package main

import (
	"testing"
	"sync"
)

func TestAppend(t *testing.T) {
	x := make([]string, 0, 6)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		y := append(x, "hello", "world")
		t.Log(len(y))
	}()
	go func() {
		defer wg.Done()
		z := append(x, "goodbye", "bob")
		t.Log(len(z))
	}()
	wg.Wait()
}
```

如果我们执行这个测试时带上 `-race` flag ，我们可以注意到一个竞争条件。

```
< go test -race .
==================
WARNING: DATA RACE
Write at 0x00c4200be060 by goroutine 8:
_/tmp.TestAppend.func2()
/tmp/main_test.go:20 +0xcb
Previous write at 0x00c4200be060 by goroutine 7:
_/tmp.TestAppend.func1()
/tmp/main_test.go:15 +0xcb
Goroutine 8 (running) created at:
_/tmp.TestAppend()
/tmp/main_test.go:18 +0x14f
testing.tRunner()
/usr/local/Cellar/go/1.10.2/libexec/src/testing/testing.go:777 +0x16d
Goroutine 7 (running) created at:
_/tmp.TestAppend()
/tmp/main_test.go:13 +0x105
testing.tRunner()
/usr/local/Cellar/go/1.10.2/libexec/src/testing/testing.go:777 +0x16d
==================
==================
WARNING: DATA RACE
Write at 0x00c4200be070 by goroutine 8:
_/tmp.TestAppend.func2()
/tmp/main_test.go:20 +0x11a
Previous write at 0x00c4200be070 by goroutine 7:
_/tmp.TestAppend.func1()
/tmp/main_test.go:15 +0x11a
Goroutine 8 (running) created at:
_/tmp.TestAppend()
/tmp/main_test.go:18 +0x14f
testing.tRunner()
/usr/local/Cellar/go/1.10.2/libexec/src/testing/testing.go:777 +0x16d
Goroutine 7 (finished) created at:
_/tmp.TestAppend()
/tmp/main_test.go:13 +0x105
testing.tRunner()
/usr/local/Cellar/go/1.10.2/libexec/src/testing/testing.go:777 +0x16d
==================
--- FAIL: TestAppend (0.00s)
main_test.go:16: 2
main_test.go:21: 2
testing.go:730: race detected during execution of test
FAIL
FAIL _/tmp 0.901s
```

## 解释为什么测试失败

理解为什么这个失败会发生，请看看这个旧例子的 `x` 的内存布局

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-append-is-not-always-thread-safe/x-starts-with-no-capacity-to-change.png)

x 没有足够的容量进行修改

Go 语言发现没有足够的内存空间来存储 `"hello", "world"` 和　`"goodbye", "bob"`，于是分配的新的内存给 `y` 与 `z`。数据竞争不会在多进程读取内存时发生，`x` 没有被修改。这里没有冲突，也就没有竞争。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-append-is-not-always-thread-safe/z-and-y-get-their-own-memory.png)

z 与 y 获取新的内存空间

在新的代码里，事情不一样了

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-append-is-not-always-thread-safe/x-has-capacity-for-more.png)

x 有更多的容量

在这里，go 注意到有足够的内存存放 `“hello”, “world”`，另一个协程也发现有足够的空间存放 `“goodbye”, “bob”`，这个竞争的发生是因为这两个协程都尝试往同一个内存空间写入，谁也不知道谁是赢家。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-append-is-not-always-thread-safe/who-wins.png)

谁赢了？

这是 Go 语言的一个特性而非 bug ，`append` 不会强制每一次调用它都申请新的内存。它允许用户在循环内进行 `append` 操作时不会破坏垃圾回收机制。缺点是你必须清楚知道在多个协程对 slice 的操作。

## 这个 bug 的认知根源

我相信这个 bug 存在是 Go 的为了保存简单，将许多概念放到 slice 中，在大多数开发人员中看到的思维过程是：

1. `x=append(x, ...)` 看起来你要获得一个新的 slice。
2. 大多数返回值的函数都不会改变它们的输入。
3. 我们使用 `append` 通常都是得到一个新的 slice。
4. 错误地认为append是只读的。

## 认知这个 bug

值得注意的是如果第一个被 `append` 的变量不是一个本地变量（译者：本地变量，即变量与 append 在同一代码块）。这个 bug 通常发生在：进行 append 操作的变量存在一个结构体中，而这个结构体是通过函数传参进来的。例如，一个结构体可以有默认值，可以被各个请求 append。小心对共享内存的变量进行 append ，或者这个内存空间（变量）并不是当前协程独占的。

## 解决方法

最简单的解决方法是不使用共享状态的第一个变量来进行 append 。相反，根据你的需要来 `make` 一个新的 `slice` ，使用这个新的 slice 作为 append 的第一个变量。下面是失败的测试示例的修正版，这里的替代方法是使用 [copy](https://golang.org/pkg/builtin/#copy) 。

```go
package main

import (
	"sync"
	"testing"
)

func TestAppend(t *testing.T) {
	x := make([]string, 0, 6)
	x = append(x, "start")

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		y := make([]string, 0, len(x)+2)
		y = append(y, x...)
		y = append(y, "hello", "world")
		t.Log(cap(y), len(y), y[0])
	}()
	go func() {
		defer wg.Done()
		z := make([]string, 0, len(x)+2)
		z = append(z, x...)
		z = append(z, "goodbye", "bob")
		t.Log(cap(z), len(z), z[0])
	}()
	wg.Wait()
}
```

对本地变量进行第一次的 append

---

via: https://medium.com/@cep21/gos-append-is-not-always-thread-safe-a3034db7975

作者：[Jack Lindamood](https://medium.com/@cep21)
译者：[lightfish-zhang](https://github.com/lightfish-zhang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
