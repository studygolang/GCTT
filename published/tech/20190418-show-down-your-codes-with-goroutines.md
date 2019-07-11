首发于：https://studygolang.com/articles/21811

# goroutine 可能使程序变慢

## 如何使用 goroutine 才能使你的 CPU 满负载运行呢

下面，我们将会展示一个关于 for 循环的代码，将输入分成几个序列添加到 Goroutines 里面！我敢打赌你之前可能有过几次这种情况，但是每次引入 gorountine 都让你的代码变得更快吗？

下面是一个简单的循环示例，它似乎很容易变成并发代码，但正如我们将看到的，并发版本不仅不会更快，实际上需要花费两倍的时间。

## 串行循环，我们以一个把索引相加的简单的串行循环作为示例

```go
// SerialSum 把 0 到 limit 的相加
package concurrencyslower

import (
	"runtime"
	"sync"
)
const (
	limit = 10000000000
)
// 实现 sum
func SerialSum() int {
	sum := 0
	for i := 0; i < limit; i++ {
		sum += i
	}
	return sum
}
```

## 并发循环

这个循环只占用一个（逻辑）CPU，因此，资深的 Gopher 们可能会采用将其分解为 Goroutines 里面运行，示例代码的 Goroutine 是可以独立于其余代码运行，因此可以分布在所有可用的 CPU 内核中。

```go
/* ConcurrentSum 函数会使用所有可用内核，获取可用逻辑核心的数量，通常这是 2*c，其中 c 是物理核心数，2 是每个核心的超线程数 。
 n:=runtime。GOMAXPROCS（0）
我们需要从某个地方收集 n 个 Goroutines 的结果。每个 Goroutine 都有一个元素的全局切片，
sums：= make（[]int, n）
现在我们可以产生 Goroutines，WaitGroup 帮助我们检测所有 Goroutine 何时完成
*/
func ConcurrentSum() int {
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		// 为每个 Goroutine 增加一个 one ADD
		wg.Add(1)
		go func(i int) {
			// 将输入分割到每个块
			start := (limit / n) * i
			end := start + (limit / n)
			// 在每个块中运行各自的 loop
			for j := start; j < end; j += 1 {
				sums[i] += j
			}
			// waitgroup 减一
			wg.Done()
		}(i)
	}
	// Done()
	wg。Wait()
	// 从各个块中收集
	sum := 0
	for _ ,  s := range sums {
		sum += s
	}
	return sum
}
```

## 然而运行速度不降反增？那么以上两个版本运行速度如何呢，让我们引入两个压力测试文件来一探究竟

```go
package concurrencyslower

import "testing"

func BenchmarkSerialSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SerialSum()
	}
}

func BenchmarkConcurrentSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ConcurrentSum()
	}
}
```

我的 CPU 是一个小型笔记本电脑 CPU (两个超线程内核，Go runtime 看作是 4 个逻辑内核)，预计，并发版本应该显示出明显的速度增益，然而，真实运行速度如何呢？

```
$ go test -bench。
goos: darwin
goarch: amd64
pkg: github.com/appliedgo/concurrencyslower
BenchmarkSerialSum-4           1      6090666568 ns/op
BenchmarkConcurrentSum-4       1      15741988135 ns/op
PASS
ok      github.com/appliedgo/concurrencyslower 21.840s
```

prefix-4 表明测试使用所有四个逻辑核。 但是，尽管并发循环使用了所有四个逻辑核，花费的时间是串行循环的两倍多，这里发生了什么？

## 硬件加速起到了相反作用

为了解释这个反直觉的结果，我们必须看一下支撑软件运行的基础，CPU 的组成原理。

CPU 的缓存内存有助于加速每个 CPU 运行速度。

为了简单起见，以下是一个粗略的过度简化，所以亲爱的 CPU 设计师，请对我宽容。 每个现代 CPU 都有一个非平凡的缓存层次结构，位于主内存和裸 CPU 内核之间，在这里，我们只谈论查看属于各个内核的各级缓存。

## CPU 缓存的目的

一般来说，缓存是一个非常小但超快的内存块，它位于 CPU 芯片上，因此每次读取或写入值时，CPU 都不必到达主 RAM。 相反，该值存储在缓存中，后续读取和写入受益于更快的 RAM 单元和更短的访问路径，CPU 的每个核都有自己的本地缓存，不与任何其他核共享。对于 n 个 CPU 内核，这意味着最多可以有 n + 1 个相同数据的副本。一个在主内存中，一个在每个 CPU 内核的缓存中。

现在，当 CPU 内核更改其本地缓存中的值时，必须在某个时刻将其同步回主内存。同样，如果缓存的值在主内存中被更改（由另一个 CPU 内核），则缓存的值无效，需要从主内存刷新。

（译注：原文有一个可以播放的动图，可以查看原文播放：https://appliedgo.net/concurrencyslower/）

## 缓存行命中

* 为了以有效的方式同步高速缓存和主存储器，数据以通常 64 字节的块同步，这些块称为缓存行，因此，当缓存值更改时，整个缓存行将同步回主内存。同样，包含此高速缓存行的所有其他 CPU 核心的高速缓存现在也必须同步此高速缓存行以避免对过时数据进行操作。

## 邻里

这对我们的代码有何影响？请记住，并发循环使用全局切片来存储中间结果。切片的元素存储在连续的空间中，概率很高，两个相邻的切片元素将共享相同的高速缓存行。

现在戏剧开始了，n 个具有 n 个高速缓存的 CPU 内核重复读取和写入全部位于同一高速缓存行中的切片元素，因此，只要一个 CPU 内核使用新的总和更新它的切片元素，所有其他 CPU 的高速缓存行就会失效，必须将更改的高速缓存行写回主内存，并且所有其他高速缓存必须使用新数据更新其各自的高速缓存行。即使每个核心访问切片的不同部分！

这消耗了宝贵的时间，超过了串行循环更新其单个和变量所需的时间。

这就是我们的并发循环比串行循环需要更多时间的原因，对切片的所有并发更新都会导致繁忙的缓存行同步更新。

总而言之，既然我们知道了处理速度变慢的原因，那么方案是显而易见的。我们必须将切片转换为 n 个单独的变量，这些变量可能被隔离存储，以便它们不共享相同的高速缓存行。

所以让我们改变我们的并发循环，以便每个 Goroutine 将其中间处理值存储在 Goroutine 的 local 变量中。为了将结果传递回至主 Goroutine，我们还必须添加一个通道。这反过来允许我们删除 WaitGroup 机制，因为通道不仅是通信的手段，而且是优雅的同步机制。

## 局部变量并发循环

```go
// ChannelSum（）产生 n 个 Goroutines，它们在本地存储它们的中间和，然后通过一个通道传回结果
func ChannelSum() int {
	n := runtime.GOMAXPROCS(0)
	//A channel of 收集所有中间值
	res := make(chan int)
	for i := 0; i < n; i++ {
		//Goroutine 接受第二个参数，结果参数 . 箭头 <- 标明只读参数 .
		go func(i int ,  r chan<- int) {
			// 本地变量取代了全局变量
			sum := 0
			// 采用了分块处理
			start := (limit / n) * i
			end := start + (limit / n)
			// 计算中间值
			for j := start; j < end; j += 1 {
				sum += j
			}
			// 传递结果
			r <- sum
			// 入参
		}(i ,  res)
	}
	sum := 0
	// This loop reads n values from the channel. We know exactly how many elements we will receive through the channel ,  hence we need no
	// 读取 n 个值  , n 事先确定
	for i := 0; i < n; i++ {
		// 读取值并相加
		// 无值时通道被阻塞，完美的的同步机制  ,
		// 本通道无值等待，直到 读取到所有的 n 个值后才关闭 .
		sum += <-res
	}
	return sum
}
```

## 测试文件中增加 BenchmarkChannelSum 测试结果如下

```
$ go test -bench .
goos: darwin
goarch: amd64
pkg: github.com/appliedgo/concurrencyslower
BenchmarkSerialSum-4          1       6022493632 ns/op
BenchmarkConcurrentSum-4      1       15828807312 ns/op
BenchmarkChannelSum-4         1       1948465461 ns/op
PASS
ok      github.com/appliedgo/concurrencyslower  23.807s
```

将使用局部变量存储处理中的值，而不是将结果它们放在一个切片中，这无疑帮助我们逃避了缓存同步问题。

但是，我们如何确保各个变量永远不会共享同一个缓存行 ? 好吧，启动一个新的 Goroutine 会在堆栈上分配 2KB 到 8KB 的数据，这比 64 字节的典型缓存行大小要多，并且由于中间和变量不是从创建它的 Goroutine 之外的任何地方引用的，因此它不会转移到堆（它可能最终接近其他中间和变量之一）。所以我们可以非常肯定没有两个中间和变量会在同一个缓存行中结束。

## 如何获取代码

使用 go get，注意 -d 参数阻止自动安装二进制到 $GOPATH/bin。

```
go get -d github.com/appliedgo/concurrencyslower
```

转到目标目录

```
cd $GOPATH/src/github.com/appliedgo/concurrencyslower
```

运行压测文件

```
go test -bench .
```

注意，代码运行的 Go 版本为 1.12，如果你的环境 Go module 参数为 enable ( 译者注，自行查看 Go module )，可以通过如下方法获取代码

```$GOPATH/pkg/mod/github.com/appliedgo/concurrencyslower@```

如果 $GOPATH 丢失，默认使用 go get ~/go 或者 %USERPROFILE%\go

## 结论

未来的 CPU 的架构或未来的 Go 版本可能会缓解以上这个测试问题。因此，如果您运行此代码，压测可能显示与本文中不同的结果，属于正常的。

通常，让 Goroutines 更新全局变量不是一个好主意。记住 Go 谚语：不要通过共享内存进行通信，通过通信共享内存。

## 这篇博客文章的灵感来自 Reddit 的讨论主题

---

via: https://appliedgo.net/concurrencyslower/

作者：[Christoph](https://appliedgo.net/about/)
译者：[dylanpoe](https://github.com/dylanpoe)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
