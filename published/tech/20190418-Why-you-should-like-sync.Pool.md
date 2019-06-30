首发于：https://studygolang.com/articles/21384

# 你为什么要喜欢 sync.Pool ？

## 介绍

因为它很快。通过文章底部存储库中的基准测试可以减少**4982 倍**的内存占用。

![Comparing pool performance. Less than better.](https://raw.githubusercontent.com/studygolang/gctt-images/master/Why-you-should-like-sync.Pool/1.png)

相比之下， Pool 的性能更快更好。

## Ok, 这究竟是怎么回事呢？

垃圾回收定期执行。如果你的代码不断地在一些数据结构中分配内存然后释放它们，这就会导致收集器的不断工作，使得更多的内存和 CPU 被用来在初始化结构体时分配资源。

> 对[sync/pool.go](https://golang.org/src/sync/pool.go) 的描述如下：
>
>Pool 是一组可以单独保存和检索的临时对象。
>
>Pool 可以安全地同时使用多个 Goroutine。

**sync.Pool**允许我们重用内存而非重新分配。

此外，如果你使用的 http 服务器接收带有 JSON 请求体的 post 请求，并且它必须被解码到结构体中，你可以使用 **sync.Pool** 来节省内存并减少服务器响应时间。

## sync.Pool 用法

sync.Pool 构造很简单：

```go
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
```

现在你将会创建一个 Pool 和新的缓冲区。你可以这样创建第一个缓冲区：

```go
buffer := bufferPool.Get().(*bytes.Buffer)
```

get 方法会返回 Pool 中已存在的 **\*bytes.Buffer**，否则将调用 **New** 方法来初始化新的 **\*bytes.Buffer**

但在缓冲区使用后，你必须将其重置并放回 Pool 中：

```go
buffer.Reset()
bufferPool.Put(buffer)
```

## 基准测试

### 将 JSON 编码为 bytes.Buffer

```bash
// 对 JSON 编码的代码段
BenchmarkReadStreamWithPool-8        5000000        384 ns/op        0 B/op        0 allocs/op
BenchmarkReadStreamWithoutPool-8     3000000        554 ns/op      160 B/op        2 allocs/op
```

我们得到了 44% 的性能提升并且节省了非常多的内存 (160B/ops vs 0B/ops)。

### 将字节写入 bufio.Writer

```bash
BenchmarkWriteBufioWithPool-8       10000000        123 ns/op      128 B/op        2 allocs/op
BenchmarkWriteBufioWithoutPool-8     2000000        651 ns/op     4288 B/op        4 allocs/op
```

我们得到了 5 倍性能提升并且减少了 32 倍内存使用。

### 将 JSON 解码为 struct

```bash
BenchmarkJsonDecodeWithPool-8        1000000       1729 ns/op     1128 B/op        8 allocs/op
BenchmarkJsonDecodeWithoutPool-8     1000000       1751 ns/op     1160 B/op        9 allocs/op
```

因为 JSON 解码操作太难，我们的性能只提升了 1%，我们无法通过重用结构体得到正常的提升。

### Gzip 字节

```bash
BenchmarkWriteGzipWithPool-8          500000       2339 ns/op      162 B/op        2 allocs/op
BenchmarkWriteGzipWithoutPool-8        10000     105288 ns/op   807088 B/op       16 allocs/op
```

等等，什么？性能提升了 45 倍并且内存使用量减少了 4982 倍。

## 总结

务必使用 **sync.Pool** ！它确实可以节省内存并提高应用程序的性能。

基准测试的 Github 存储库在[这里](https://github.com/Mnwa/GoBench)。

---

via: https://medium.com/@Mnwa/why-you-should-like-sync-pool-2c7960c023ba

作者：[Mnwa Mnowich](https://medium.com/@Mnwa)
译者：[RookieWilson](https://github.com/RookieWilson)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
