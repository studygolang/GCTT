首发于：https://studygolang.com/articles/28461

# Go 高级基准测试

## 背景

有时你必须解决不同类型的问题。通常来说复杂的问题并不会只有单一的解决方案，但是解决方案的优劣取决于程序在运行时所要解决问题的子集。

我所遇到的一个例子是分析一些代理的连接中的某些数据流。

从流量中提取信息的方法主要有两种：保存整个数据流，当流量结束后立即分析；或者（使用一个缓存窗口）以降低速度为代价，在数据流传输过程中进行分析。

内存相对与处理能力来说要更加便宜，所以我的第一版解决方案是使用缓存的方案。

### 第一版代码：使用缓存（buffer）

缓存连接是相对容易的：只需要将读到的所有数据复制到一个 `bytes.Buffer` 里，然后当连接关闭后对读到的数据进行分析。最简单的方式是包装（ wrap ）连接，在调用 `Read` 前先经过一个 [`io.TeeReader`](https://golang.org/pkg/io/#TeeReader)。

这十分简单，并且在低流量，短连接的场景下表现的很好。

```go
type bufferedScanner struct {
	net.Conn
	tee    io.Reader
	buffer *bytes.Buffer
}

func NewBufferedScanner(original net.Conn) net.Conn {
	var buffer bytes.Buffer
	return bufferedScanner {
		Conn: original,
		tee: io.TeeReader(c, &buffer),
		buffer: &buffer,
	}
}

func (b bufferedScanner) Read(p []byte) (n int, err error) {
	return b.tee.Read(p)
}

func (b bufferedScanner) Close() error{
	analyse(b.buffer)
	return b.Conn.Close()
}
```

经过小的优化以[有效地重用缓存](https://golang.org/pkg/sync/#example_Pool)后，我对这个方案感到满意，至少有一会儿是这样。

### 第二版代码：使用扫描器（scanner）

不久之后，我意识到这个方案处理不好长连接和突发流量的场景，于是我写了一些可以流式工作而不是缓存所有数据的代码。

就初始化的内存空间而言，这个方案有着更大的开销（用来构建 scanner 以及其他额外的数据结构），但是在同一个连接发送几十 kb 数据之后，这个方案在内存和计算方面变得更加的高效。

流式的解决方案实现起来比较棘手，但是得益于 `bufio` 包，这些困难是可控的。实现代码仅仅是有着自定义 [`SplitFunc`](https://golang.org/pkg/bufio/#SplitFunc) 的 [`scanner`](https://golang.org/pkg/bufio/#Scanner) 的包装。

## 核心问题

不管我的解决方案如何，我现在有两段代码，它们各有利弊：对于低流量生存周期短的连接，第一种方案更好，但是对于流量密集的场景，第二种方案是唯一可行的。

我有两个可能的优化办法：尝试优化第二个方案，让这个方案在小的连接上同样可行，或者基于在运行时看到的内容，选择合适的实现。

我选择了第二种，也是看起来更有趣的一个。

## 解决方案

我创建了一个构造器（builder）来提供以及实例化两个实现。这个构造器维护着每次流量的总流量的大小的一个[指数加权移动平均值(exponential weighted moving average)](https://en.wikipedia.org/wiki/Moving_average#Exponential_moving_average)。大致上与 TCP 协议用来估算 RTT 和到达时间变化的算法类似。

很有意思的是，我实现它需要的字数要少于我描述它所用到的字数：

```go
ewma := k * n + (1 - k) * ewma
```

其中 `n` 是在 `Close` 时从连接中读取到的总字节。`k` 在代码中仅仅是一个调节对大小变化的启发式反应速度的常量，我这里设定为二分之一。

困难的地方在于选择切换实现的阈值：我运行了一些基准测试，并且找到了流式方案开始表现得比缓存方案更好的那个临界点，但是我很快发现这个临界点的值很大程度上依赖于运行代码的计算机。

## 误伤

如同 Go [在 `math/big` 包](https://github.com/golang/go/blob/50bd1c4d4eb4fac8ddeb5f063c099daccfb71b26/src/math/big/calibrate_test.go#L18)中启发式选择正确的算法来进行计算，大部分情况下，在笔者的笔记本上运行的基准测试足以确定这个常量。按照 math/big 贡献者所说的那样，这种方式的问题是，它会导致[差异超过 100% 的错误](https://github.com/golang/go/issues/25580)。

不用说，我不喜欢这种方式，于是我开始整理我可用的工具。我不希望使用 makefile 或其他外部依赖。我也不希望我的用户能够使用我的代码之前要安装或者运行外部的代码，于是我思考**用户在编译我的库的时候什么是已经可以使用的**。

Go 是跨平台的，所以我不需要处理不同平台的差异，但是另一方面，我需要一些信息来感知现有的运算能力。

### 测量

基准测试相对简单：go 语言有着[内置的基准测试标准库](https://golang.org/pkg/testing/#hdr-Benchmarks)。

在默默地盯了文档几分钟之后，我意识到我可以通过调用 [`testing.Benchmark`](https://golang.org/pkg/testing/#Benchmark)的方式在非测试构建（non-testing build）中运行基准测试，testing.Benchmark 会返回一个不错的 [testing.BenchmarkResult](https://golang.org/pkg/testing/#BenchmarkResult)。

于是我设置了一些 `func(b *testing.B)` 的匿名函数对输入值（比如，存根连接大小）构建闭包，来对两个方案运行基准测试，看哪一个方案表现的更好。

```go
type ConfigurableBenchmarker struct {
	Name string
	GetBench func(input []byte) func(b *testing.B)
}
```

一个 `ConfigurableBenchmarker`（可配置基准测试）的例子，以及如何使用它：

```go
ConfigurableBenchmarker{
	Name: "Buffered",
	GetBench: func(input []byte) func(b *testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				c := NewBufferedScanner(stubConnection{
					bytes.NewReader(input),
				})
				io.Copy(ioutil.Discard, c)
				c.Close()
			}
		}
	},
}

// doBench 运行两个 ConfigurableBenchmarkers 并返回
// 是否第一个方案耗时更少
func doBench(size int, aa, bb ConfigurableBenchmarker) bool {
	aaRes = testing.Benchmark(aa.GetBench(genInput(size)))
	bbRes = testing.Benchmark(bb.GetBench(genInput(size)))
	return aaRes.NsPerOp() < bbRes.NsPerOp()
}
```

使用这个构建块，我可以进行二分搜索，来确定一个方案比另一个方案效果更好的输入大小。

```go
func FindTipping(aa, bb ConfigurableBenchmarker, lower, upper int) (int,error) {
	lowerCond := doBench(lower, aa, bb)
	upperCond := doBench(upper, aa, bb)
	if lowerCond == upperCond {
		return 0, ErrTippingNotInRange
	}

	// 经典的二分搜索
	tip = (lower + upper) / 2
	for tip > lower && tip < upper {
		tipCond := doBench(tip, aa, bb)
		if tipCond == lowerCond {
			lower = tip
		} else {
			upper = tip
		}
		tip = (lower + upper) / 2
	}
	return tip, nil
}
```

这里是在 `lower = 1` 并且 `upper = 100` 时的一些输出和一些日志：

```
Calculating initial values...
AnalysingTraffic: [1KB] Buffered 1107 ns/op < 1986 ns/op Streamed
AnalysingTraffic: [100KB] Buffered 87985 ns/op >= 69509 ns/op Streamed
Starting search...
Binsearch: lower: 1, upper: 100
AnalysingTraffic: [50KB] Buffered 43455 ns/op >= 35242 ns/op Streamed
Binsearch: lower: 1, upper: 50
AnalysingTraffic: [25KB] Buffered 22693 ns/op >= 19506 ns/op Streamed
Binsearch: lower: 1, upper: 25
AnalysingTraffic: [13KB] Buffered 11355 ns/op >= 10263 ns/op Streamed
Binsearch: lower: 1, upper: 13
AnalysingTraffic: [7KB] Buffered 4964 ns/op < 5824 ns/op Streamed
Binsearch: lower: 7, upper: 13
AnalysingTraffic: [10KB] Buffered 7415 ns/op < 8140 ns/op Streamed
Binsearch: lower: 10, upper: 13
AnalysingTraffic: [11KB] Buffered 8609 ns/op < 8765 ns/op Streamed
Binsearch: lower: 11, upper: 13
AnalysingTraffic: [12KB] Buffered 9828 ns/op < 10157 ns/op Streamed
Tipping point was found at 12
Most efficient for input of smaller sizes was "Buffered"
Most efficient for input of bigger sizes was "Streamed"
```

有了这些，我可以自动在当前机器上检测出代码的临界点。现在我需要运行这段代码。我可以在 README 文件中写个说明，但是这样做的话有什么意思？

### Go generate

`go generate` 命令可以解析具有特定语法的注释，并运行其中的内容。

运行 `go generate` 的话，下面的注释会让 Go 打印个招呼。
```go
//go:generate Echo "Hello, World!"
```

所以，当用户 `go get` 一个包的时候，他们可以 `go generate` 一些代码，然后 `go build` 或者将他们与来源链接。

我将基准测试代码包装到一个 `generator.go` 中，这个文件运行基准测试并且将常量写入源文件。仅仅是将基准测试获取到的数字格式化为一个字符串，并写入到本地文件：

```go
const src = `// Code generated; DO NOT EDIT.

package main

const streamingThreshold = %d
`

func main() {
	tip := FindTipping(/* params */)
	// Omitted: open file "constants_generated.go" for writing in `f`
	fmt.Fprintf(f, src, tip)
}
```

之后，我只需要在其他源文件中增加一个注释：

```go
//go:generate Go run generator.go
```

目标机器必须安装了 `go`，用来编译我的代码。这意味着我没有要求用户增加其他的额外工具或依赖。

这很不错，但有一个严重的问题：在不使用外部构建工具的前提下，无法让 `main` 包和 `analyse` 共存于一个文件夹。

确实如此，除非你（滥）用[构建标签（build tags）](https://golang.org/pkg/go/build/#hdr-Build_Constraints)：你可以防止一个文件在读取 `package` 语句之前考虑构建。

所以我用这样的开头修改了我的生成器代码：

```go
// +build generate

package main
```

并且将原来的代码注释改成了这样：
```go
//go:generate Go run generate.go -tags generate
```

现在的结构：

```
analyse
├── analyse.go              ← Package analyse, 有着 //go:generate 指令
├── analyse_test.go         ← Package analyse, 测试
├── constants_generated.go  ← Package analyse, 生成的代码
└── generate.go             ← Package main, 隐藏着一个 tag
```

于是，我现在可以 `go get` 或者 `git clone` 我的 package，运行 `go generate`，之后可以针对我的机器优化后运行。

## 效果

这里是三种方案的最终基准测试结果。在我运行基准测试的机子的临界点是 12K。单位是 `ns/op`。

```
Dim |  Buff  | Adapt  | Stream
----|--------|--------|-------
 1K |  1159  |  1278  |   1965
 2K |  1723  |  1868  |   2574
 4K |  2842  |  3055  |   4450
 8K |  5644 ← | → 5929  |   7446
16K | 15359  | 13478 ← | → 13539
32K | 29814  | 25430  |  24980
64K | 58821  | 49078  |  48596
```

适应性的解决方案在测量流量大小上有一点开销，所以它永远不如其他方案中最好的一个表现的那么好，但几乎是最佳的。另一方面，该适应性方案总是比更糟糕的方案要好。

---

via: https://blogtitle.github.io/go-advanced-benchmarking/

作者：[Rob](https://blogtitle.github.io/authors/rob/)
译者：[dust347](https://github.com/dust347)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
