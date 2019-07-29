# GO:Finalizers

![avatar](./gopher.png)

这篇文章基于GO-1.12版本

Go runtime提供了一种允许开发者将一个函数与一个变量绑定的方法runtime.SetFinalizer,被绑定的变量从它无法被访问时就被垃圾回收器视为待回收状态。这个特性引起了高度的争论，但本文不不打算参与其中，而是去阐述这个方法的具体实现。


## 无保障性

举一个使用了Finalizer的例子

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

这段程序将会在这个循环中创建三个struct的的实例，并将每个实例都绑定一个finalizer。之后垃圾回收器将会被调用回收之前创建的实例。运行这个程序，将会给到我
们如下输出：

```
31
37
47
```

如我们所见，finalizers并没有被调用，关于runtime的文档能够很好地解释这一点：

```
finalizer被调用在程序不再能获取到一个obj所指向的对象后的任意时刻，对于finalizer将会运行在程序退出之前这一点，也无法得到保证。因此一般情况下，它们仅
在一个长期运行的程序，且释放了一些与对象相关的的非内存资源时才有效。
```

runtime并不为finalizer被调用之前的延迟提供任何保障，让我们试着去修改我们的程序，通过在调用垃圾回收器之后添加一个一秒的sleep:

```
31
37
47
foo 1 has been garbage collected
foo 0 has been garbage collected
```

现在我们的finalizer已经被调用了，然而，它们其中一个消失了。我们的finalizers与垃圾回收器连接，并且垃圾回收器回收以及清理数据的方式将会对finalizers的调用
产生影响。


## 工作流

之前的例子可能让我认为Go仅在释放我们所定义的struct的内存之前调用finalizers。

让我们深入其中，看看在更多的Allocation中到底发生了些什么。

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
一百万个structs和finalizers被创建出来，下面是输出:

```
Allocation: 0.090862 Mb, Number of allocation: 137
Allocation: 31.107506 Mb, Number of allocation: 2390078
Allocation: 110.052666 Mb, Number of allocation: 4472742
```

让我们再试一次，这次不用finalizers:

```
Allocation: 0.090694 Mb, Number of allocation: 136
Allocation: 18.129814 Mb, Number of allocation: 1390078
Allocation: 0.094451 Mb, Number of allocation: 154
```

看起来没有任何资源在内存中被清理掉，即时垃圾回收器被触发，且finalizers也运行。为了理解这一行为，让我们回到那篇关于runtime的文档:
```
当垃圾回收器发现了一个已关联finalizer的无法访问的块，这说明了关联操作与运行finalizer是在一个单独的gorountine下。这让obj再次可访问，不过现在没有了一
个关联的finalizer,假设SetFinalizer没有再次被调用，当下次垃圾回收器看到这个obj时，它是不可被访问的，并将回收它。
```

如我们所见，finalizers首先会被移除，然后内存将在下一次循环中被释放，让我们再次运行第一个例子，并加上两个强制的垃圾回收操作。

```
Allocation: 0.090862 Mb, Number of allocation: 137
Allocation: 31.107506 Mb, Number of allocation: 2390078
Allocation: 110.052666 Mb, Number of allocation: 4472742
Allocation: 0.099220 Mb, Number of allocation: 166
```

我们可以清楚地看到，第二次运行将会清理数据，finalizers最终也对性能和内存使用产生了轻微的作用。


## 性能表现

下文阐述了为何finalizers逐个运行：

```
一个单独goroutine为了一个程序运行了所有的finalizers,然而，如果一个finalizer必须长时间运行，则需要开启一个新的gorountine。
```

仅一个goroutine将会运行finalizers，并且任何超重任务都需要开启一个新的gorountine。当finalizers运行时，垃圾回收器并没有停止且并发运行中。因此finalizer并不该影响你的应用的性能
表现。

当然，Go提供一个方法移除finalizer一旦它不再被需要。

```go
 runtime.SetFinalizer(p, nil)
```

它允许我们根据使用情况动态地移除finalizers。


## 应用中的使用

内部上，Go在net以及net/http包中确保文件先前的打开与关闭准确无误，并且在os包中确保之前创建的进程被正常地释放。这里有一个来自os包的例子：

```go
func newProcess(pid int, handle uintptr) *Process {
   p := &Process{Pid: pid, handle: handle}
   runtime.SetFinalizer(p, (*Process).Release)
   return p
}
```

当这个进程被释放，finalizer也会被移除。

```go
func (p *Process) release() error {
   // NOOP for unix.
   p.Pid = -1
   // no need for a finalizer anymore
   runtime.SetFinalizer(p, nil)
   return nil
}
```

Go同样也在测试中使用finalizers确保在垃圾回收器中期望的动作被执行，举个例子，sync包使用了finalizers测试在垃圾回收循环中pool是否被清空。

