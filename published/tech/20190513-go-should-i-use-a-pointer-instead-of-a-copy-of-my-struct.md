首发于：https://studygolang.com/articles/21763

# Go：我应该用指针替代结构体的副本吗？

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-should-i-use-a-pointer-instead-of-a-copy-of-my-struct-44b43b104963/1_IO4bo74w6aX7rKC_spjmvw.png)

对于许多 `golang` 开发者来说，考虑到性能，最佳实践是系统地使用指针而非结构体副本。

我们将回顾两个用例，来理解使用指针而非结构体副本的影响。

## 1. 数据分配密集型

让我们举一个简单的例子，说明何时要为使用值而共享结构体：

```go
type S struct {
   a, b, c int64
   d, e, f string
   g, h, i float64
}
```

这是一个可以由副本或指针共享的基本结构体：

```go
func byCopy() S {
   return S{
      a: 1, b: 1, c: 1,
      e: "foo", f: "foo",
      g: 1.0, h: 1.0, i: 1.0,
   }
}

func byPointer() *S {
   return &S{
      a: 1, b: 1, c: 1,
      e: "foo", f: "foo",
      g: 1.0, h: 1.0, i: 1.0,
   }
}
```

基于这两种方法，我们现在可以编写两个基准测试，其中一个是通过副本传递结构体的：

```go
func BenchmarkMemoryStack(b *testing.B) {
   var s S

   f, err := os.Create("stack.out")
   if err != nil {
      panic(err)
   }
   defer f.Close()

   err = trace.Start(f)
   if err != nil {
      panic(err)
   }

   for i := 0; i < b.N; i++ {
      s = byCopy()
   }

   trace.Stop()

   b.StopTimer()

   _ = fmt.Sprintf("%v", s.a)
}
```

另一个非常相似，它通过指针传递：

```go
func BenchmarkMemoryHeap(b *testing.B) {
   var s *S

   f, err := os.Create("heap.out")
   if err != nil {
      panic(err)
   }
   defer f.Close()

   err = trace.Start(f)
   if err != nil {
      panic(err)
   }

   for i := 0; i < b.N; i++ {
      s = byPointer()
   }

   trace.Stop()

   b.StopTimer()

   _ = fmt.Sprintf("%v", s.a)
}
```

让我们运行基准测试：

```
go test ./... -bench=BenchmarkMemoryHeap -benchmem -run=^$ -count=10 > head.txt && benchstat head.txt
go test ./... -bench=BenchmarkMemoryStack -benchmem -run=^$ -count=10 > stack.txt && benchstat stack.txt
```

以下是统计数据：

```
name          time/op
MemoryHeap-4  75.0ns ± 5%
name          alloc/op
MemoryHeap-4   96.0B ± 0%
name          allocs/op
MemoryHeap-4    1.00 ± 0%
------------------
name           time/op
MemoryStack-4  8.93ns ± 4%
name           alloc/op
MemoryStack-4   0.00B
name           allocs/op
MemoryStack-4    0.00
```

在这里，使用结构体副本比指针快 8 倍。

为了理解原因，让我们看看追踪生成的图表：

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-should-i-use-a-pointer-instead-of-a-copy-of-my-struct-44b43b104963/1_tUgeQdgYoHwOFuWzyUX_cw.png)

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-should-i-use-a-pointer-instead-of-a-copy-of-my-struct-44b43b104963/1_VPgyB_GjbEkcyHIZ_NyZFQ.png)

第一张图非常简单。由于没有使用堆，因此没有垃圾收集器，也没有额外的 `goroutine`。
对于第二张图，使用指针迫使 `go` 编译器[将变量逃逸到堆](https://golang.org/doc/faq#stack_or_heap)，由此增大了垃圾回收器的压力。如果我们放大图表，我们可以看到，垃圾回收器占据了进程的重要部分。

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-should-i-use-a-pointer-instead-of-a-copy-of-my-struct-44b43b104963/1_SUlM_idjAevNfofEhgm5YA.png)

在这张图中，我们可以看到，垃圾回收器每隔 4ms 必须工作一次。
如果我们再次缩放，我们可以详细了解正在发生的事情：

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-should-i-use-a-pointer-instead-of-a-copy-of-my-struct-44b43b104963/1_Ik7agDlBN6dwLaL_4U806Q.png)

蓝色，粉色和红色是垃圾收集器的不同阶段，而棕色的是与堆上的分配相关（在图上标有 “runtime.bgsweep”）：

> 清扫是指回收与堆内存中未标记为使用中的值相关联的内存。当应用程序 `Goroutines` 尝试在堆内存中分配新值时，会触发此活动。清扫的延迟被添加到在堆内存中执行分配的成本中，并且与垃圾收集相关的任何延迟没有关系。

[Go 中的垃圾回收：第一部分 - 基础](https://studygolang.com/articles/21569)

即使这个例子有点极端，我们也可以看到，与栈相比，在堆上为变量分配内存是多么消耗资源。在我们的示例中，与在堆上分配内存并共享指针相比，代码在栈上分配结构体并复制副本要快得多。

如果你不熟悉堆栈或堆，如果你想更多地了解栈或堆的内部细节，你可以在网上找到很多资源，比如 `Paul Gribble` 的[这篇文章](https://www.gribblelab.org/CBootCamp/7_Memory_Stack_vs_Heap.html)。

如果我们使用 `GOMAXPROCS = 1` 将处理器限制为 1，情况会更糟：

```
name        time/op
MemoryHeap  114ns ± 4%
name        alloc/op
MemoryHeap  96.0B ± 0%
name        allocs/op
MemoryHeap   1.00 ± 0%
------------------
name         time/op
MemoryStack  8.77ns ± 5%
name         alloc/op
MemoryStack   0.00B
name         allocs/op
MemoryStack    0.00
```

如果栈上分配的基准数据不变，则堆上的基准从 `75ns/op` 降低到 `114ns/op`。

## 2.方法调用密集型

对于第二个用例，我们将在结构体中添加两个空方法，稍微调整一下我们的基准测试：

```go
func (s S) stack(s1 S) {}

func (s *S) heap(s1 *S) {}
```

在栈上分配的基准测试将创建一个结构体并通过复制副本传递它：

```go
func BenchmarkMemoryStack(b *testing.B) {
   var s S
   var s1 S

   s = byCopy()
   s1 = byCopy()
   for i := 0; i < b.N; i++ {
      for i := 0; i < 1000000; i++  {
         s.stack(s1)
      }
   }
}
```

堆的基准测试将通过指针传递结构体：

```go
func BenchmarkMemoryHeap(b *testing.B) {
   var s *S
   var s1 *S

   s = byPointer()
   s1 = byPointer()
   for i := 0; i < b.N; i++ {
      for i := 0; i < 1000000; i++ {
         s.heap(s1)
      }
   }
}
```

正如预期的那样，结果现在大不相同：

```
name          time/op
MemoryHeap-4  301µs ± 4%
name          alloc/op
MemoryHeap-4  0.00B
name          allocs/op
MemoryHeap-4   0.00
------------------
name           time/op
MemoryStack-4  595µs ± 2%
name           alloc/op
MemoryStack-4  0.00B
name           allocs/op
MemoryStack-4   0.00
```

## 结论

在 `go` 中使用指针而不是结构体的副本并不总是好事。为了能为你的数据选择好的语义，我强烈建议您阅读 [Bill Kennedy](https://twitter.com/goinggodotnet) 撰写的[关于值/指针语义的文章](https://studygolang.com/articles/12487)。它将为你提供更好的视角来决定使用自定义类型或内置类型时的策略。

此外，内存使用情况分析肯定会帮助你弄清楚你的内存分配和堆上发生了什么。

---

via: https://medium.com/@blanchon.vincent/go-should-i-use-a-pointer-instead-of-a-copy-of-my-struct-44b43b104963

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[DoubleLuck](https://github.com/DoubleLuck)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
