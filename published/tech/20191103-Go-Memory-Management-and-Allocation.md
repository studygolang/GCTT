首发于：https://studygolang.com/articles/28436

# Go：内存管理分配

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191103-Go-Memory-Management-and-Allocation/cover.png)

ℹ️ *这篇文章基于 Go 1.13。*

在内存从分配到回收的生命周期中，内存不再被使用的时候，标准库会自动执行 Go 的内存管理。虽然开发者不必操心这些细节，但是 Go 语言所做的底层管理经过了很好的优化，同时有很多有趣的概念。

## 堆上的分配

内存管理被设计为可以在并发环境快速执行，同时与垃圾收集器集成在了一起。从一个简单的例子开始：

```go
package main

type smallStruct struct {
   a, b int64
   c, d float64
}

func main() {
   smallAllocation()
}

//go:noinline
func smallAllocation() *smallStruct {
   return &smallStruct{}
}
```

注释 `//go:noinline` 会禁用内联，以避免内联通过移除函数的方式优化这段代码，从而造成最终没有分配内存的情况出现。

通过运行逃逸分析命令 `go tool compile "-m" main.go` 可以确认 Go 执行了的分配：

```
main.go:14:9: &smallStruct literal escapes to heap
```

借助 `go tool compile -S main.go` 命令得到这段程序的汇编代码，可以同样明确地向我们展示具体的分配细节：
```
0x001d 00029 (main.go:14)   LEAQ   type."".smallStruct(SB), AX
0x0024 00036 (main.go:14)  PCDATA $0, $0
0x0024 00036 (main.go:14)  MOVQ   AX, (SP)
0x0028 00040 (main.go:14)  CALL   runtime.newobject(SB)
```

函数 `newobject` 是用于新对象的分配以及代理 `mallocgc` 的内置函数，该函数在堆上管理这些内存。在 Go 语言中有两种策略，一种用于较小的内存空间的分配，而另一种则用于较大的内存空间的分配。

## 较小内存空间的分配策略

对于小于 32kb 的，较小的内存空间的分配策略，Go 会从被叫做 `mcache` 的本地缓存中尝试获取内存。 这个缓存持有一个被叫做 `mspan` 的内存块(span ，32kb 大小的内存块)列表, mspan 包含着可用于分配的内存：

![用 mcache 分配内存](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191103-Go-Memory-Management-and-Allocation/allocation-with-mcache.png)

每个线程 `M` 被分配一个处理器 `P`，并且一次最多处理一个 goroutine。在分配内存时，当前的 goroutine 会使用它当前的 P 的本地缓存，在 span 链表中寻找第一个可用的空闲对象。使用这种本地缓存不需要锁操作，从而分配效率更高。

span 链表被划分为 8 字节大小到 32k 字节大小的，约 70 个的大小等级，每个等级可以存储不同大小的对象。

![span 的大小等级](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191103-Go-Memory-Management-and-Allocation/span-size-classes.png)

每个 span 链表会存在两份：一个链表用于不包含指针的对象而另一个用于包含指针的对象。这种区别使得垃圾收集器更加轻松，因为它不必扫描不包含任何指针的 span。

在我们前面的例子中，结构体的大小是 32 字节，因此它会适合于 32 字节的 span ：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191103-Go-Memory-Management-and-Allocation/previous-example.png)

现在，我们可能会好奇，如果在分配期间 span 没有空闲的插槽会发生什么。Go 维护着每个大小等级的 span 的中央链表，该中央链表被叫做 `mcentral`，其中维护着包含空闲对象的 span 和没有空闲对象的 span ：

![span 的中央链表](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191103-Go-Memory-Management-and-Allocation/central-lists-of-spans.png)

`mcentral` 维护着 span 的双向链表；其中每个链表节点有着指向前一个 span 和后一个 span 的引用。非空链表中的 span 可能包含着一些正在使用的内存，“非空”表示在链表中至少有一个空闲的插槽可供分配。当垃圾收集器清理内存时，可能会清理一部分 span，将这部分标记为不再使用，并将其放回非空链表。

我们的程序现在可以在没有插槽的情况下向中央链表请求 span ：

![从 mcentral 中替换 span ](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191103-Go-Memory-Management-and-Allocation/span-replacement-from-mcentral.png)

如果空链表中没有可用的 span，Go 需要为中央链表获取新的 span 。新的 span 会从堆上分配，并链接到中央链表上：

![从堆上分配 span ](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191103-Go-Memory-Management-and-Allocation/span-allocation-from-the-heap.png)

堆会在需要的时候从系统（ OS ）获取内存，如果需要更多的内存，堆会分配一个叫做 `arena` 的大块内存，在 64 位架构下为 64Mb，在其他架构下大多为 4Mb。arena 同样适用 span 映射内存。

![堆由 arena 组成](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191103-Go-Memory-Management-and-Allocation/heap-is-composed-by-arenas.png)

## 较大内存空间的分配策略

Go 并不适用本地缓存来管理较大的内存空间分配。对于超过 32kb 的分配，会向上取整到页的大小，并直接从堆上分配。

![直接从堆上进行大的内存空间分配](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191103-Go-Memory-Management-and-Allocation/large-allocation-directly-from-the-heap.png)

## 全景图

现在我们对内存分配的时候发生了什么有了更好的认识。现在将所有的组成部分放在一起来得到完整的图画。

![内存分配的组成](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191103-Go-Memory-Management-and-Allocation/components-of-the-memory-allocation.png)

## 灵感来源

该内存分配最初基于 TCMalloc，一个 Google 创建的，并发环境优化的内存分配器。这个 [TCMalloc 的文档](http://goog-perftools.sourceforge.net/doc/tcmalloc.html)值得阅读；你会发现上面解释过的概念。

---

via: https://medium.com/a-journey-with-go/go-memory-management-and-allocation-a7396d430f44

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[dust347](https://github.com/dust347)
校对：[@unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
