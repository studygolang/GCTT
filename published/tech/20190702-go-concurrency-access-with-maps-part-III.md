首发于：https://studygolang.com/articles/22778

# Go: 并发访问 Map — Part III

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-concurrency-access-with-maps-Part-III/1_uZMa7x3KBqJKJ6rWtnZFwA.png)

在上一篇文章 “[Go: 通过源码研究 Map 的设计](https://studygolang.com/articles/22777)” 中，我们讲述了 map 的内部实现。

[Go blog](https://blog.golang.org/go-maps-in-action) 中专门讲解 map 的文章明确地表明：

> [map 是非并发安全的](https://golang.org/doc/faq#atomic_maps)：并发读写 map 时，map 的行为是未知的。如果你需要使用并发执行的 goroutine 同时读写 map，必须使用某种同步机制来协调访问。

然而，正如 [FAQ](https://golang.org/doc/faq#atomic_maps) 中解释的，Google 提供了一些帮助：

> 作为一种纠正 map 使用方式的辅助手段，语言的某些实现包含了特殊的检查，当运行时的 map 被不安全地并发修改时，它会自动报告。

## 数据争用检测

我们可以从 Go 获得的第一个帮助就是数据争用检测。使用 `-race` 标记来运行你的程序或测试会让你了解潜在的数据争用。让我们看一个例子：

```go
func main() {
   m := make(map[string]int, 1)
   m[`foo`] = 1

   var wg sync.WaitGroup

   wg.Add(2)
   go func() {
      for i := 0; i < 1000; i++  {
         m[`foo`]++
      }
   }()
   go func() {
      for i := 0; i < 1000; i++  {
         m[`foo`]++
      }
   }()
   wg.Wait()
}
```

在这个例子中，我们清晰地看到，在某一时刻，两个 goroutine 尝试同时写入一个新值。下面是争用检测器的输出：

```
==================
WARNING: DATA RACE
Read at 0x00c00008e000 by goroutine 6:
   runtime.mapaccess1_faststr()
      /usr/local/go/src/runtime/map_faststr.go:12 +0x0
   main.main.func2()
      main.go:19 +0x69

Previous write at 0x00c00008e000 by goroutine 5:
   runtime.mapassign_faststr()
      /usr/local/go/src/runtime/map_faststr.go:202 +0x0
   main.main.func1()
      main.go:14 +0xb8
```

争用检测器解释道，当第二个 goroutine 正在读变量时，第一个 goroutine 正在向同一个内存地址写一个新值。如果你想要了解更多，我建议你阅读我的一篇关于[数据争用检测器](https://medium.com/@blanchon.vincent/go-race-detector-with-threadsanitizer-8e497f9e42db)的文章。

## 并发写入检测

Go 提供的另一个帮助是并发写入检测。让我们使用之前看到的那个例子。运行这个程序时，我们将看到一个错误：

```go
fatal error: concurrent map writes
```

在 map 结构的内部标志 `flags` 的帮助下，Go 处理了这次并发。当代码尝试修改 map 时（赋值，删除值或者清空 map），`flags` 的某一位会被置为 1:

```go
func mapdelete(t *maptype, h *hmap, key unsafe.Pointer) {
   [...]
   h.flags ^= hashWriting
```

值为 4 的 `hashWriting` 会将相关的位置为 1。

^ 是一个异或操作，如果两个操作数的某一位的值不同，^ 将该位置为 1：

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-concurrency-access-with-maps-Part-III/1_4OrKbRPWgBTNf-zvSYr_hA.png)

当操作结束时，该标志会被重置：

```go
func mapdelete(t *maptype, h *hmap, key unsafe.Pointer) {
   [...]
   h.flags &^= hashWriting
}
```

既然每个修改 map 的操作都设置了一个控制标志，那么通过检查这个标志的状态，就可以防止并发写入。这里是该标志的生命周期的例子：

```go
func mapdelete(t *maptype, h *hmap, key unsafe.Pointer) {
   [...]
   // if another process is currently writing, throw error
   if h.flags&hashWriting != 0 {
      throw("concurrent map writes")
   }
   [...]
   // no one is writing, we can set now the flag
   h.flags ^= hashWriting
   [...]
   // flag reset
   h.flags &^= hashWriting
}
```

## sync.Map vs Map with lock

`sync` 包也提供了并发安全的 map。不过，正如[文档](https://golang.org/pkg/sync/)中解释的，你应该谨慎的选择你使用的 map：

> `sync` 包中的 map 类型是专业的。大多数代码应该使用原生的 Go map，附加上锁或者其他协调方式，这样类型安全更有保障，而且更容易维护其他的不变量和 map 的内容。

实际上，正如我的文章 “[Go: 通过源码研究 Map 的设计](https://studygolang.com/articles/22777)” 中所解释的，map 根据我们处理的具体类型提供了不同的方法。

让我们运行一个简单的基准测试，比较带有锁的常规 map 和 `sync` 包的 map。一个基准测试并发写入 map，另一个仅仅读取 map 中的值：

```go
MapWithLockWithWriteOnlyInConcurrentEnc-8  68.2µs ± 2%
SyncMapWithWriteOnlyInConcurrentEnc-8       192µs ± 2%
MapWithLockWithReadOnlyInConcurrentEnc-8   76.8µs ± 3%
SyncMapWithReadOnlyInConcurrentEnc-8       55.7µs ± 4%
```

我们可以看到，两种 map 各有千秋。我们可以根据具体的情况选择其中之一。[文档](https://golang.org/pkg/sync/#Map)中很好地解释了这些情况：

> map 类型针对两种常见使用场景做了优化：(1) 指定 key 的 entry 仅写入一次，但多次读取，比如只增长的缓存； (2) 多个 goroutine 读取、写入、覆盖不相交的 key 的集合指向的 entry。

## Map vs sync.Map

[FAQ](https://golang.org/doc/faq#atomic_maps) 中也解释了他们做出了默认情况下 map 非并发安全这个决定的原因：

> 因此，要求所有的 map 操作都获取互斥锁，会拖慢大多数程序，但只为很少的程序增加了安全性

让我们运行一个不使用并发 goroutine 的基准测试，来理解当你不需要并发但标准库默认提供并发安全的 map 时，可能带来的影响：

```
MapWithWriteOnly-8          11.1ns ± 3%
SyncMapWithWriteOnly-8       121ns ± 6%

MapWithReadOnly-8           4.87ns ± 7%
SyncMapWithReadOnly-8       29.2ns ± 4%
```

简单的 map 快 7 到 10 倍。显然，在非并发模式下，这听起来更合理，巨大的差异也清楚的解释了为什么默认非并发安全的 map 是更好的选择。如果你不需要处理并发状况，为什么要让程序运行的更慢呢？

---

via: https://medium.com/@blanchon.vincent/go-concurrency-access-with-maps-part-iii-8c0a0e4eb27e

作者：[blanchon.vincent](https://medium.com/@blanchon.vincent)
译者：[DoubleLuck](https://github.com/DoubleLuck)
校对：[dingdingzhou](https://github.com/dingdingzhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
