首发于：https://studygolang.com/articles/23461

# Go: Finalizers

这篇文章基于 Go-1.12 版本

Go runtime 提供了一种允许开发者将一个函数与一个变量绑定的方法 `runtime.SetFinalizer`，被绑定的变量从它无法被访问时就被垃圾回收器视为待回收状态。这个特性引起了高度的争论，但本文并不打算参与其中，而是去阐述这个方法的具体实现。

## 无保障性

举一个使用了 Finalizer 的例子

```go
package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"time"
)

type Foo struct {
	a int
}

func main() {
	for i := 0; i < 3; i++ {
		f := NewFoo(i)
		println(f.a)
	}

	runtime.GC()
}

//go:noinline
func NewFoo(i int) *Foo {
	f := &Foo{a: rand.Intn(50)}
	runtime.SetFinalizer(f, func(f *Foo) {
		fmt.Println(`foo ` + strconv.Itoa(i) + ` has been garbage collected`)
	})

	return f
}
```

这段程序将会在这个循环中创建三个 struct 的的实例，并将每个实例都绑定一个 finalizer。之后垃圾回收器将会被调用,并回收之前创建的实例。运行这个程序，将会给到我们如下输出：

```
31
37
47
```

如我们所见，finalizers 并没有被调用，runtime 的文档解释这一点：

> 在程序无法获取到一个 obj 所指向的对象后的任意时刻，finalizer 被调度运行，且无法保证 finalizer 运行在程序退出之前。因此一般情况下，因此它们仅用于在长时间运行的程序上释放一些与对象关联的非内存资源。

在调用 finalizer 之前，runtime 不提供有关延迟的任何保证。让我们试着去修改我们的程序，通过在调用垃圾回收器之后添加一个一秒的 sleep:

```
31
37
47
foo 1 has been garbage collected
foo 0 has been garbage collected
```

现在我们的 finalizer 已经被调用了，然而，它们其中一个消失了。我们的 finalizers 与垃圾回收器相连接，并且垃圾回收器回收以及清理数据的方式将会对 finalizers 的调用产生影响。

## 工作流

之前的例子可能让我认为 Go 仅在释放我们所定义的 struct 的内存之前调用 finalizers。

让我们深入其中，看看在更多的 Allocation 中到底发生了些什么。

```go
package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"
)

type Foo struct {
	a int
}

func main() {
	debug.SetGCPercent(-1)

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	fmt.Printf("Allocation: %f Mb, Number of allocation: %d\n", float32(ms.HeapAlloc)/float32(1024*1204), ms.HeapObjects)

	for i := 0; i < 1000000; i++ {
		f := NewFoo(i)
		_ = fmt.Sprintf("%d", f.a)
	}

	runtime.ReadMemStats(&ms)
	fmt.Printf("Allocation: %f Mb, Number of allocation: %d\n", float32(ms.HeapAlloc)/float32(1024*1204), ms.HeapObjects)

	runtime.GC()
	time.Sleep(time.Second)

	runtime.ReadMemStats(&ms)
	fmt.Printf("Allocation: %f Mb, Number of allocation: %d\n", float32(ms.HeapAlloc)/float32(1024*1204), ms.HeapObjects)

	runtime.GC()
	time.Sleep(time.Second)
}

//go:noinline
func NewFoo(i int) *Foo {
	f := &Foo{a: rand.Intn(50)}
	runtime.SetFinalizer(f, func(f *Foo) {
		_ = fmt.Sprintf("foo " + strconv.Itoa(i) + " has been garbage collected")
	})

	return f
}
```

一百万个 structs 和 finalizers 被创建出来，下面是输出:

```
Allocation: 0.090862 Mb, Number of allocation: 137
Allocation: 31.107506 Mb, Number of allocation: 2390078
Allocation: 110.052666 Mb, Number of allocation: 4472742
```

让我们再试一次，这次不用 finalizers:

```
Allocation: 0.090694 Mb, Number of allocation: 136
Allocation: 18.129814 Mb, Number of allocation: 1390078
Allocation: 0.094451 Mb, Number of allocation: 154
```

看起来没有任何资源在内存中被清理掉，即使垃圾回收器被触发，且 finalizers 也运行。为了理解这一行为，让我们回到那篇关于 runtime 的文档:

> 当垃圾回收器发现了一个已关联 finalizer 的无法访问的块，这说明了关联操作与运行 finalizer 是在一个单独的 gorountine 下。这让 obj 再次可访问，不过现在没有了一个关联的 finalizer,假设 SetFinalizer 没有再次被调用，当下次垃圾回收器看到这个 obj 时，它是不可被访问的，并将回收它。

如我们所见，finalizers 首先会被移除，然后内存将在下一次循环中被释放，让我们再次运行第一个例子，并加上两个强制的垃圾回收操作。

```
Allocation: 0.090862 Mb, Number of allocation: 137
Allocation: 31.107506 Mb, Number of allocation: 2390078
Allocation: 110.052666 Mb, Number of allocation: 4472742
Allocation: 0.099220 Mb, Number of allocation: 166
```

我们可以清楚地看到，第二次运行将会清理数据，finalizers 最终也对性能和内存使用产生了轻微的作用。

## 性能表现

下文阐述了为何 finalizers 逐个运行：

> 一个单独 goroutine 为了一个程序运行了所有的 finalizers,然而，如果一个 finalizer 必须长时间运行，则需要开启一个新的 gorountine。

仅一个 goroutine 将会运行 finalizers，并且任何超重任务都需要开启一个新的 gorountine。当 finalizers 运行时，垃圾回收器并没有停止且并发运行中。因此 finalizer 并不该影响你的应用的性能表现。

同时，一旦 finalizer 不再被需要，Go 提供了一个方法来移除它。

```go
 runtime.SetFinalizer(p, nil)
```

它允许我们根据使用情况动态地移除 finalizers。

## 应用中的使用

内部上，Go 在 net 以及 net/http 包中确保文件先前的打开与关闭准确无误，并且在 os 包中确保之前创建的进程被正常地释放。这里有一个来自 os 包的例子：

```go
func newProcess(pid int, handle uintptr) *Process {
	p := &Process{Pid: pid, handle: handle}
	runtime.SetFinalizer(p, (*Process).Release)
	return p
}
```

当这个进程被释放，finalizer 也会被移除。

```go
func (p *Process) release() error {
	// NOOP for unix.
	p.Pid = -1
	// no need for a finalizer anymore
	runtime.SetFinalizer(p, nil)
	return nil
}
```

Go 同样也在测试中使用 finalizers 确保在垃圾回收器中期望的动作被执行，举个例子，sync 包使用了 finalizers 测试在垃圾回收循环中 pool 是否被清空。

---

via: https://medium.com/@blanchon.vincent/go-finalizers-786df8e17687

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[Maple24](https://github.com/Maple24)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
