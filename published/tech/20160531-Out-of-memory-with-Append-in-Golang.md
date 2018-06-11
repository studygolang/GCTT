已发布：https://studygolang.com/articles/12929

# Go 语言中 append 导致内存不足

这是一篇简短的笔记，关于你或许会遇上的 Go 语言的内存不足的问题。

如你所知，Go 语言的 slice 很强大且使用简单。通过 Go 语言的内置函数，它可以解决我们许多问题。

但是今天，我更多想讲述的是 slice 与它的内存。

首先，创建 slice 最简单的方式是：

```go
var sliceOfInt []int
var sliceOfString []int
var sliceOfUser []User
var sliceOfMap []map[string]int

letters := []string{"a", "b", "c", "d"}
numbers := []string{1, 2, 3, 4, 5}
numbers := []User{ User{ Name: "Thuc"}, User{Name: "Mr Vu"} }
```

当你创建一个 slice, 将会定义三个重要的信息，包含这个数组的指针 pointer，堆栈的已使用长度 length ，以及容量 capacity （堆栈的最大长度）

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/out-of-memory-with-append-in-golang/slice_memory.png)

> 长度是切片引用的元素的数量
> 容量是底层数组中元素的数量（从切片指针引用的元素开始）

于是，当你使用我给出的以上方式创建 slices ，长度和容量将是最小的。

我们转下一个话题，Go 语言的 [append](https://golang.org/pkg/builtin/#append) 函数。

> 这个内置函数 append 把元素追加到 slice 的末尾，如果有足够的容量，就直接在末尾的内存加入这个新元素。如果没有足够的空间，则新分配一个新的基础的数组，追加更新的 slice 。

问题在于我们经常以为 append 会追加一个新的元素到 slice 的末尾，而忘记了如果没有足够的预留容量，这个操作会导致在新的内存上重新分配一个新的 slice 。

使用这个最简单的方式来创建 slice ，容量（capacity）总是与切片中的当前元素相匹配。不幸的是，有些源代码里是将 append 函数放到循环或嵌套循环中执行，快速的循环没什么问题，但是更大的循环，可能会引发问题，它会急速消耗内存。

```go
var largeOfList []int
var result []int

for _, item := range largeOfList {
  if item.Status == true {
    result = append(result, item)
  }
}
```

你可以想象，每一个循环都会分配一个拥有新的容量的新内存，并且它还没来得及释放就得内存，我们就会看到以下错误

```
fatal error: runtime: out of memory

runtime stack:
runtime.throw(0xd4b870, 0x16)
	/home/sss/go/src/runtime/panic.go:530 +0x90
runtime.sysMap(0xc88b750000, 0x2f120000, 0xc8204ec000, 0x1113238)
	/home/sss/go/src/runtime/mem_linux.go:206 +0x9b
runtime.(*mheap).sysAlloc(0x10f8c60, 0x2f120000, 0x2000000000002)
	/home/sss/go/src/runtime/malloc.go:429 +0x191
runtime.(*mheap).grow(0x10f8c60, 0x17890, 0x0)
	/home/sss/go/src/runtime/mheap.go:651 +0x63
runtime.(*mheap).allocSpanLocked(0x10f8c60, 0x1788e, 0x7f4b42f275f0)
	/home/sss/go/src/runtime/mheap.go:553 +0x4f6
runtime.(*mheap).alloc_m(0x10f8c60, 0x1788e, 0x100000000, 0x7f4b417f0dd0)
	/home/sss/go/src/runtime/mheap.go:437 +0x119
runtime.(*mheap).alloc.func1()
	/home/sss/go/src/runtime/mheap.go:502 +0x41
runtime.systemstack(0x7f4b417f0de8)
	/home/sss/go/src/runtime/asm_amd64.s:307 +0xab
runtime.(*mheap).alloc(0x10f8c60, 0x1788e, 0x100000000, 0xc820187500)
	/home/sss/go/src/runtime/mheap.go:503 +0x63
runtime.largeAlloc(0x2f11c000, 0x100000003, 0x434cc2)
	/home/sss/go/src/runtime/malloc.go:766 +0xb3
runtime.mallocgc.func3()
	/home/sss/go/src/runtime/malloc.go:664 +0x33
runtime.systemstack(0xc820014000)
	/home/sss/go/src/runtime/asm_amd64.s:291 +0x79
runtime.mstart()
	/home/sss/go/src/runtime/proc.go:1048

goroutine 22 [running]:
runtime.systemstack_switch()
	/home/sss/go/src/runtime/asm_amd64.s:245 fp=0xc820200bc8 sp=0xc820200bc0
runtime.mallocgc(0x2f11c000, 0x0, 0x3, 0xc8204f4000)
	/home/sss/go/src/runtime/malloc.go:665 +0x9eb fp=0xc820200ca0 sp=0xc820200bc8
runtime.rawmem(0x2f11c000, 0x7880000)
	/home/sss/go/src/runtime/malloc.go:809 +0x32 fp=0xc820200cc8 sp=0xc820200ca0
runtime.growslice(0xa4c8a0, 0xc8204f4000, 0x4b4f800, 0x4b4f800, 0x4b4f801, 0x0, 0x0, 0x0)
	/home/sss/go/src/runtime/slice.go:95 +0x233 fp=0xc820200d38 sp=0xc820200cc8
```

## 内存不足

Go 内置的代码在给 slice 分配内存空间时会预留更多的容量（capacity），这使得 append 不会因为容量不足而分配新的内存。

总而言之，我写这一篇文章，是希望我们在编程时，不妨思考多一些关于内存如何操作的问题。

希望这文章能帮到您 :)

---

via: https://blog.siliconstraits.com/out-of-memory-with-append-in-golang-956e7eb2c70e

作者：[Thuc Le](https://blog.siliconstraits.com/@thuc)
译者：[lightfish-zhang](https://github.com/lightfish-zhang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
