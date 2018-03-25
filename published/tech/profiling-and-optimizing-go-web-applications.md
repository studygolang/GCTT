已发布：https://studygolang.com/articles/12685

# 剖析与优化 Go 的 web 应用

原文发表日期: 2017/3/13

关键字: `dev` `go` `golang` `pprof`

Go 语言有一个很强大的内置分析器（profiler），支持CPU、内存、协程 与 阻塞/抢占（block/contention）的分析。

## 开启分析器（profiler）

Go 提供了一个低级的分析 API [runtime/pprof](https://golang.org/pkg/runtime/pprof/) ，但如果你在开发一个长期运行的服务，使用更高级的 [net/http/pprof](https://golang.org/pkg/net/http/pprof/) 包会更加便利。

你只需要在代码中加入 `import _ "net/http/pprof"` ，它就会自动注册所需的 HTTP 处理器（Handler） 。

```go
package main

import (
	"net/http"
	_ "net/http/pprof"
)

func hiHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

func main() {
	http.HandleFunc("/", hiHandler)
	http.ListenAndServe(":8080", nil)
}
```

如果你的 web 应用使用自定义的 URL 路由，你需要手动注册一些 HTTP 端点（endpoints）　。

```go
package main

import (
	"net/http"
	"net/http/pprof"
)

func hiHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

func main() {
	r := http.NewServeMux()
	r.HandleFunc("/", hiHandler)

	// Register pprof handlers
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	http.ListenAndServe(":8080", r)
}
```

如上代码那样，开启 web 应用，然后使用 pprof 工具：

```shell
go tool pprof [binary] http://127.0.0.1:8080/debug/pprof/profile
```

pprof 的最大的优点之一是它是的性能负载很小，可以在生产环境中使用，不会对 web 请求响应造成明显的性能消耗。

但是在深入挖掘 pprof 之前，我们需要一个真实案例来展示如何在 GO 应用中检查并解决性能问题。

## 案例： Left-pad 微服务

假设，你需要开发一个全新的微服务，为输入的字符串添加左填充

```shell
$ curl "http://127.0.0.1:8080/v1/leftpad/?str=test&len=10&chr=*"
{"str":"******test"}
```

这个服务需要收集基本的指标（metric），如请求的数量与每个请求的响应时间。收集到的所有指标都应该发送到一个指标聚合器（metric aggregator）（例如 [StatsD](https://github.com/etsy/statsd)）除此之外，这个服务需要日志记录这个请求的详细信息，如 URL，IP 地址与 user-agent 。

你可以在 Github 上看到这个微服务的初步实现，tag 为 [v1](https://github.com/akrylysov/goprofex/tree/v1)

编译并运行这个应用

```shell
go build && ./goprofex
```

## 性能分析

我们将要测试这个微服务每秒可以处理多少个请求，可以使用这个工具 [Apache Benchmark tool](https://httpd.apache.org/docs/2.4/programs/ab.html) :

```shell
ab -k -c 8 -n 100000 "http://127.0.0.1:8080/v1/leftpad/?str=test&len=50&chr=*"
# -k   Enables HTTP keep-alive
# -c   Number of concurrent requests
# -n   Number of total requests to make
```

测试结果不差，但可以做到更快

```shell
Requests per second:    22810.15 [#/sec] (mean)
Time per request:       0.042 [ms] (mean, across all concurrent requests)
```

注：上面的测试结果的执行环境：笔记本 MacBook Pro Late 2013 (2.6 GHz Intel Core i5, 8 GB 1600 MHz DDR3, macOS 10.12.3) , Go编译器版本是1.8 。

## CPU 分析（CPU profile）

再次执行 Apache benchmark tool ，但这次使用更高的请求数量（1百万应该足够了），并同时执行 pprof ：

```shell
go tool pprof goprofex http://127.0.0.1:8080/debug/pprof/profile
```

这个 CPU profiler 默认执行30秒。它使用采样的方式来确定哪些函数花费了大多数的CPU时间。Go runtime 每10毫秒就停止执行过程并记录每一个运行中的协程的当前堆栈信息。

当 pprof 进入交互模式，输入 `top`，这条命令会展示收集样本中最常出现的函数列表。在我们的案例中，是所有 runtime 与标准库函数，这不是很有用。

```shell
(pprof) top
63.77s of 69.02s total (92.39%)
Dropped 331 nodes (cum <= 0.35s)
Showing top 10 nodes out of 78 (cum >= 0.64s)
	  flat  flat%   sum%        cum   cum%
	50.79s 73.59% 73.59%     50.92s 73.78%  syscall.Syscall
	 4.66s  6.75% 80.34%      4.66s  6.75%  runtime.kevent
	 2.65s  3.84% 84.18%      2.65s  3.84%  runtime.usleep
	 1.88s  2.72% 86.90%      1.88s  2.72%  runtime.freedefer
	 1.31s  1.90% 88.80%      1.31s  1.90%  runtime.mach_semaphore_signal
	 1.10s  1.59% 90.39%      1.10s  1.59%  runtime.mach_semaphore_wait
	 0.51s  0.74% 91.13%      0.61s  0.88%  log.(*Logger).formatHeader
	 0.49s  0.71% 91.84%      1.06s  1.54%  runtime.mallocgc
	 0.21s   0.3% 92.15%      0.56s  0.81%  runtime.concatstrings
	 0.17s  0.25% 92.39%      0.64s  0.93%  fmt.(*pp).doPrintf
```

有一个更好的方法来查看高级别的性能概况 —— `web` 命令，它会生成一个热点（hot spots）的 SVG 图像，可以在浏览器中打开它：

![](https://github.com/studygolang/gctt-images/raw/master/profiling-and-optimizing-go-web-applications/web-cpu.png)

从上图你可以看到这个应用花费了 CPU 大量的时间在 logging、测试报告（metric reporting ）上，以及部分时间在垃圾回收上。

使用 `list` 命令可以 inspect 每个函数的详细代码，例如 `list leftpad` ：

```shell
(pprof) list leftpad
ROUTINE ======================== main.leftpad in /Users/artem/go/src/github.com/akrylysov/goprofex/leftpad.go
	  20ms      490ms (flat, cum)  0.71% of Total
		 .          .      3:func leftpad(s string, length int, char rune) string {
		 .          .      4:   for len(s) < length {
	  20ms      490ms      5:       s = string(char) + s
		 .          .      6:   }
		 .          .      7:   return s
		 .          .      8:}
```

对无惧查看反汇编代码的人而言，可以使用 pprof 的 `disasm` 命令，它有助于查看实际的处理器指令：

```shell
(pprof) disasm leftpad
ROUTINE ======================== main.leftpad
	  20ms      490ms (flat, cum)  0.71% of Total
		 .          .    1312ab0: GS MOVQ GS:0x8a0, CX
		 .          .    1312ab9: CMPQ 0x10(CX), SP
		 .          .    1312abd: JBE 0x1312b5e
		 .          .    1312ac3: SUBQ $0x48, SP
		 .          .    1312ac7: MOVQ BP, 0x40(SP)
		 .          .    1312acc: LEAQ 0x40(SP), BP
		 .          .    1312ad1: MOVQ 0x50(SP), AX
		 .          .    1312ad6: MOVQ 0x58(SP), CX
...
```

## 函数堆栈分析（Heap profile）

执行堆栈分析器（heap profiler）

```shell
go tool pprof goprofex http://127.0.0.1:8080/debug/pprof/heap
```

默认情况下，它显示当前正在使用的内存量：

```shell
(pprof) top
512.17kB of 512.17kB total (  100%)
Dropped 85 nodes (cum <= 2.56kB)
Showing top 10 nodes out of 13 (cum >= 512.17kB)
	  flat  flat%   sum%        cum   cum%
  512.17kB   100%   100%   512.17kB   100%  runtime.mapassign
		 0     0%   100%   512.17kB   100%  main.leftpadHandler
		 0     0%   100%   512.17kB   100%  main.timedHandler.func1
		 0     0%   100%   512.17kB   100%  net/http.(*Request).FormValue
		 0     0%   100%   512.17kB   100%  net/http.(*Request).ParseForm
		 0     0%   100%   512.17kB   100%  net/http.(*Request).ParseMultipartForm
		 0     0%   100%   512.17kB   100%  net/http.(*ServeMux).ServeHTTP
		 0     0%   100%   512.17kB   100%  net/http.(*conn).serve
		 0     0%   100%   512.17kB   100%  net/http.HandlerFunc.ServeHTTP
		 0     0%   100%   512.17kB   100%  net/http.serverHandler.ServeHTTP
```

但是我们更感兴趣的是分配的对象的数量，执行 pprof 时使用选项 `-alloc_objects`

```shell
go tool pprof -alloc_objects goprofex http://127.0.0.1:8080/debug/pprof/heap
```

几乎 70% 的对象仅由两个函数分配　——　`leftpad` 与 `StatsD` ，我们需要更仔细的查看它们：

```shell
(pprof) top
559346486 of 633887751 total (88.24%)
Dropped 32 nodes (cum <= 3169438)
Showing top 10 nodes out of 46 (cum >= 14866706)
	  flat  flat%   sum%        cum   cum%
 218124937 34.41% 34.41%  218124937 34.41%  main.leftpad
 116692715 18.41% 52.82%  218702222 34.50%  main.(*StatsD).Send
  52326692  8.25% 61.07%   57278218  9.04%  fmt.Sprintf
  39437390  6.22% 67.30%   39437390  6.22%  strconv.FormatFloat
  30689052  4.84% 72.14%   30689052  4.84%  strings.NewReplacer
  29869965  4.71% 76.85%   29968270  4.73%  net/textproto.(*Reader).ReadMIMEHeader
  20441700  3.22% 80.07%   20441700  3.22%  net/url.parseQuery
  19071266  3.01% 83.08%  374683692 59.11%  main.leftpadHandler
  17826063  2.81% 85.90%  558753994 88.15%  main.timedHandler.func1
  14866706  2.35% 88.24%   14866706  2.35%  net/http.Header.clone
```

还有一些非常有用的调试内存问题的选项， ` -inuse_objects` 可以显示正在使用的对象的数量，`-alloc_space` 可以显示程序启动以来分配的多少内存。

自动内存分配很便利，但世上没有免费的午餐。动态内存分配不仅比堆栈分配要慢得多，还会间接地影响性能。你在堆上分配的每一块内存都会增加 GC 的负担，并且占用更多的 CPU 资源。要使垃圾回收花费更少的时间，唯一的方法是减少内存分配。

## 逃逸分析（Escape analysis）

无论何时使用 `＆` 运算符来获取指向变量的指针或使用 `make` 或 `new` 分配新值，它并不一定意味着它被分配在堆上。

```go
func foo(a []string) {
	fmt.Println(len(a))
}

func main() {
	foo(make([]string, 8))
}
```

在上面的例子中，　`make([]string, 8)` 是在栈上分配内存的。Go 通过 escape analysis 来判断使用堆而不是栈来分配内存是否安全。你可以添加选项 `-gcflags=-m` 来查看逃逸分析（escape analysis）的结果：

```go
5  type X struct {v int}
6
7  func foo(x *X) {
8       fmt.Println(x.v)
9  }
10
11 func main() {
12      x := &X{1}
13      foo(x)
14 }
```

```shell
go build -gcflags=-m
./main.go:7: foo x does not escape
./main.go:12: main &X literal does not escape
```

Go 编译器足够智能，可以将一些动态分配转换为栈分配。但你如果使用接口来处理变量，会导致糟糕的情况。

```go
// Example 1
type Fooer interface {
	foo(a []string)
}

type FooerX struct{}

func (FooerX) foo(a []string) {
	fmt.Println(len(a))
}

func main() {
	a := make([]string, 8) // make([]string, 8) escapes to heap
	var fooer Fooer
	fooer = FooerX{}
	fooer.foo(a)
}

// Example 2
func foo(a interface{}) string {
	return a.(fmt.Stringer).String()
}

func main() {
	foo(make([]string, 8)) // make([]string, 8) escapes to heap
}
```

Dmitry Vyukov 的论文 [Go Escape Analysis Flaws](https://docs.google.com/document/d/1CxgUBPlx9iJzkz9JWkb6tIpTe5q32QDmz8l0BouG0Cw/view) 讲述了更多的逃逸分析（escape analysis）无法处理的案例。

一般来说，对于你不需要再修改数据的小结构体，你应该使用值传参而不是指针传参。

> 注：对于大结构体，使用指针传参而不是值传参（复制整个结构体）的性能消耗更低。 

## 协程分析（Goroutine profile）

Goroutine profile 会转储协程的调用堆栈与运行中的协程数量

```shell
go tool pprof goprofex http://127.0.0.1:8080/debug/pprof/goroutine
```

![](https://github.com/studygolang/gctt-images/raw/master/profiling-and-optimizing-go-web-applications/web-goroutine.png)

上图只有18个活跃中的协程，这是非常小的数字。拥有数千个运行中的协程的情况并不少见，但并不会显著降低性能。

## 阻塞分析（Block profile）

阻塞分析会显示导致阻塞的函数调用，它们使用了同步原语（synchronization primitives），如互斥锁（mutexes）和 channels 。

在执行 block contention profile 之前，你必须设置使用[runtime.SetBlockProfileRate](https://golang.org/pkg/runtime/#SetBlockProfileRate) 设置 profiling rate　。你可以在 `main` 函数或者 `init` 函数中添加这个调用。

```shell
go tool pprof goprofex http://127.0.0.1:8080/debug/pprof/block
```

![](https://github.com/studygolang/gctt-images/raw/master/profiling-and-optimizing-go-web-applications/web-block.png)

`timedHandler` 与 `leftpadHandler` 花费了大量的时间来等待 `log.Printf` 中的互斥锁。导致这个结果的原因是 `log` package 的实现使用了互斥锁来对多个协程共享的文件进行同步访问（synchronize access）。

## 指标（Benchmarking）

正如我们之前注意的，在这个案例的最大的几个性能杀手是 `log` package ，`leftpad` 与 `StatsD.Send` 函数。现在我们找到了性能瓶颈，但是在优化代码之前，我们需要一个可重复的方法来对我们关注的代码进行性能测试。Go 的 [testing](https://golang.org/pkg/testing/) package 包含了这样的一个机制。你需要在测试文件中创建一个函数，以 `func BenchmarkXxx(*testing.B)` 的格式。

```go
func BenchmarkStatsD(b *testing.B) {
	statsd := StatsD{
		Namespace:  "namespace",
		SampleRate: 0.5,
	}
	for i := 0; i < b.N; i++ {
		statsd.Incr("test")
	}
}
```

也可以使用 [net/http/httptest](https://golang.org/pkg/net/http/httptest/) 对这整个 HTTP 程序进行基准测试：

```go
func BenchmarkLeftpadHandler(b *testing.B) {
	r := httptest.NewRequest("GET", "/v1/leftpad/?str=test&len=50&chr=*", nil)
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		leftpadHandler(w, r)
	}
}
```

执行基准测试：

```shell
go test -bench=. -benchmem
```

它会显示每次迭代需要的时间量，以及 内存/分配数量 （amount of memory/number of allocations）：

```
BenchmarkTimedHandler-4           200000          6511 ns/op        1621 B/op         41 allocs/op
BenchmarkLeftpadHandler-4         200000         10546 ns/op        3297 B/op         75 allocs/op
BenchmarkLeftpad10-4             5000000           339 ns/op          64 B/op          6 allocs/op
BenchmarkLeftpad50-4              500000          3079 ns/op        1568 B/op         46 allocs/op
BenchmarkStatsD-4                1000000          1516 ns/op         560 B/op         15 allocs/op
```

## 优化性能

### Logging

让应用运行更快，一个很好又不是经常管用的方法是，让它执行更少的工作。除了 debug 的目的之外，这行代码 `log.Printf("%s request took %v", name, elapsed)` 在 web service 中不需要。所有非必要的 logs 应该在生产环境中被移除代码或者关闭功能。可以使用分级日志（a leveled logger）来解决这个问题，比如这些很棒的 [日志工具库（logging libraries）](https://github.com/avelino/awesome-go#logging)

关于打日志或者其他一般的 I/O 操作，另一个重要的事情是尽可能使用有缓冲的输入输出（buffered input/output），这样可以减少系统调用的次数。通常，并不是每个 logger 调用都需要立即写入文件 —— 使用 [bufio](https://golang.org/pkg/bufio/) package 来实现 buffered I/O 。我们可以使用 `bufio.NewWriter` 或者 `bufio.NewWriterSize` 来简单地封装 `io.Writer` 对象，再传递给 logger ：

```go
log.SetOutput(bufio.NewWriterSize(f, 1024*16))
```

### 左填充（leftpad）

再看一遍 `leftpad` 函数

```go
func leftpad(s string, length int, char rune) string {
	for len(s) < length {
		s = string(char) + s
	}
	return s
}
```

在每一个循环中连接字符串的做法并不高效，因为每一次循环迭代都会分配一个新的字符串（反复分配内存空间）。有一种更好的方法来构建字符串，使用 [bytes.Buffer](https://golang.org/pkg/bytes/#Buffer) ：

```go
func leftpad(s string, length int, char rune) string {
	buf := bytes.Buffer{}
	for i := 0; i < length-len(s); i++ {
		buf.WriteRune(char)
	}
	buf.WriteString(s)
	return buf.String()
}
```

另外，我们还可以使用 [string.Repeat](https://golang.org/pkg/strings/#Repeat) ，使代码更加简洁：

```go
func leftpad(s string, length int, char rune) string {
	if len(s) < length {
		return strings.Repeat(string(char), length-len(s)) + s
	}
	return s
}
```

### StatsD client

接下来需要优化的代码是 `StatsD.Send` 函数：

```go
func (s *StatsD) Send(stat string, kind string, delta float64) {
	buf := fmt.Sprintf("%s.", s.Namespace)
	trimmedStat := strings.NewReplacer(":", "_", "|", "_", "@", "_").Replace(stat)
	buf += fmt.Sprintf("%s:%s|%s", trimmedStat, delta, kind)
	if s.SampleRate != 0 && s.SampleRate < 1 {
		buf += fmt.Sprintf("|@%s", strconv.FormatFloat(s.SampleRate, 'f', -1, 64))
	}
	ioutil.Discard.Write([]byte(buf)) // TODO: Write to a socket
}
```

以下是有一些可能的值得改进的地方：

- `Sprintf` 对字符串格式化非常便利，它性能表现很好，除非你每秒调用它几千次。不过，它把输入参数进行字符串格式化的时候会消耗 CPU 时间，而且每次调用都会分配一个新的字符串。为了更好的性能优化，我们可以使用 `bytes.Buffer` + `Buffer.WriteString/Buffer.WriteByte` 来替换它。
- 这个函数不需要每一次都创建一个新的 `Replacer` 实例，它可以声明为全局变量，或者作为 `StatsD` 结构体的一部分。
- 用 `strconv.AppendFloat` 替换 `strconv.FormatFloat` ，并且使用堆栈上分配的 buffer 来传递变量，防止额外的堆分配。

```go
func (s *StatsD) Send(stat string, kind string, delta float64) {
	buf := bytes.Buffer{}
	buf.WriteString(s.Namespace)
	buf.WriteByte('.')
	buf.WriteString(reservedReplacer.Replace(stat))
	buf.WriteByte(':')
	buf.Write(strconv.AppendFloat(make([]byte, 0, 24), delta, 'f', -1, 64))
	buf.WriteByte('|')
	buf.WriteString(kind)
	if s.SampleRate != 0 && s.SampleRate < 1 {
		buf.WriteString("|@")
		buf.Write(strconv.AppendFloat(make([]byte, 0, 24), s.SampleRate, 'f', -1, 64))
	}
	buf.WriteTo(ioutil.Discard) // TODO: Write to a socket
}
```

这样做，将分配数量（number of allocations）从14减少到1个，并且使 `Send` 运行快了4倍。

```
BenchmarkStatsD-4                5000000           381 ns/op         112 B/op          1 allocs/op
```

## 测试优化结果

做了所有优化之后，基准测试显示出非常好的性能提升：

```
benchmark                     old ns/op     new ns/op     delta
BenchmarkTimedHandler-4       6511          1181          -81.86%
BenchmarkLeftpadHandler-4     10546         3337          -68.36%
BenchmarkLeftpad10-4          339           136           -59.88%
BenchmarkLeftpad50-4          3079          201           -93.47%
BenchmarkStatsD-4             1516          381           -74.87%

benchmark                     old allocs     new allocs     delta
BenchmarkTimedHandler-4       41             5              -87.80%
BenchmarkLeftpadHandler-4     75             18             -76.00%
BenchmarkLeftpad10-4          6              3              -50.00%
BenchmarkLeftpad50-4          46             3              -93.48%
BenchmarkStatsD-4             15             1              -93.33%

benchmark                     old bytes     new bytes     delta
BenchmarkTimedHandler-4       1621          448           -72.36%
BenchmarkLeftpadHandler-4     3297          1416          -57.05%
BenchmarkLeftpad10-4          64            24            -62.50%
BenchmarkLeftpad50-4          1568          160           -89.80%
BenchmarkStatsD-4             560           112           -80.00%
```

注: 作者使用 [benchcmp](https://godoc.org/golang.org/x/tools/cmd/benchcmp) 来对比结果：

再一次运行 `ab`

```
Requests per second:    32619.54 [#/sec] (mean)
Time per request:       0.030 [ms] (mean, across all concurrent requests)
```

这个 web 服务现在可以每秒多处理10000个请求！　

## 优化技巧

- 避免不必要的 heap 内存分配。
- 对于不大的结构体，值传参比指针传参更好。
- 如果你事先知道长度，最好提前分配 maps 或者 slice 的内存。
- 生产环境下，非必要情况不打日志。
- 如果你要频繁进行连续的读写，请使用缓冲读写（buffered I/O）
- 如果你的应用广泛使用 JSON，请考虑使用解析器/序列化器（parser/serializer generators）（作者个人更喜欢 [easyjson](https://github.com/mailru/easyjson)）
- 在主要路径上的每一个操作都很关键（Every operation matters in a hot path）

## 结论

有时候，性能瓶颈可能不是你预想那样，理解应用程序真实性能的最好途径是认真分析它。

你可以在 [Github](https://github.com/akrylysov/goprofex) 上找到本案例的完整的源代码，初始版本 tag 为 v1，优化版本 tag 为 v2 。比较这两个版本的[传送门](https://github.com/akrylysov/goprofex/compare/v1...v2) 。

> 作者并非以英语为母语，并且他在努力提高英语水平，如果原文有表达问题或者语法错误，请纠正他。

---

via: http://artem.krylysov.com/blog/2017/03/13/profiling-and-optimizing-go-web-applications/

作者：[Artem Krylsov](http://artem.krylysov.com/)
译者：[lightfish-zhang](https://github.com/lightfish-zhang)
校对：[Unknwon](https://github.com/Unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出