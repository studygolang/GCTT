首发于：https://studygolang.com/articles/14410

# 用 runtime 包做 Go 应用的基本监控

您可能想知道——特别是如果您刚开始使用 Go，该如何给您的微服务应用添加监控。正如那些有跟踪记录的人告诉您——监控是很困难的。那么我要告诉您的是至少基本的监控不是那样的。您不需要为您的简单应用启动一个 [Prometheus](https://prometheus.io/) 集群去获得报告，事实上，您甚至不需要额外的服务去添加一个您的应用统计的简单输出。

但是我们的应用程序的哪些特性是我们感兴趣的呢？这个 Go 的 [runtime](https://golang.org/pkg/runtime/) 包包含了一些和 Go 的运行系统交互的函数——像这个调度器和内存管理器等。这意味着我们能够访问一些内部的应用程序：

## Goroutines

goroutine 是 Go 的调度管理器为我们准备的非常轻量级的线程。在任何代码中可能会出现的一个典型问题被称为“ goroutines 泄露”。这个问题的原因有很多种，如忘记设置默认的 http 请求超时，SQL 超时，缺乏对上下文包取消的支持，向已关闭的通道发数据等。当这个问题发生时，一个 goroutine 可能无限期的存活，并且永远不释放它所使用的资源。

我们可能会对 [runtime.NumGoroutine() int](https://golang.org/pkg/runtime/#NumGoroutine) 这个很基本当函数感兴趣，它会返回当前存在的 goroutines 数量。我们只要打印这个数字并在一段时间内检查它，就可以合理的确认我们可能 goroutines 泄漏，然后调查这些问题。

## 内存占用

在 Go 的世界里内存占用问题是很普遍的。当大多数人倾向于使用高效的指针时（比在 Node.js 中的任何东西都高效），一个经常遇到的与性能相关的问题是关于内存分配。演示一个简单的，但低效的反转字符串的方式：

```go
package main

import (
	"strings"
	"testing"
)

func BenchmarkStringReverseBad(b *testing.B) {
	b.ReportAllocs()

	input := "A pessimist sees the difficulty in every opportunity; an optimist sees the opportunity in every difficulty."

	for i := 0; i < b.N; i++ {
		words := strings.Split(input, " ")
		wordsReverse := make([]string, 0)
		for {
			word := words[len(words)-1:][0]
			wordsReverse = append(wordsReverse, word)
			words = words[:len(words)-1]
			if len(words) == 0 {
				break
			}
		}
		output := strings.Join(wordsReverse, " ")
		if output != "difficulty. every in opportunity the sees optimist an opportunity; every in difficulty the sees pessimist A" {
			b.Error("Unexpected result: " + output)
		}
	}
}

func BenchmarkStringReverseBetter(b *testing.B) {
	b.ReportAllocs()

	input := "A pessimist sees the difficulty in every opportunity; an optimist sees the opportunity in every difficulty."

	for i := 0; i < b.N; i++ {
		words := strings.Split(input, " ")
		for i := 0; i < len(words)/2; i++ {
			words[len(words)-1-i], words[i] = words[i], words[len(words)-1-i]
		}
		output := strings.Join(words, " ")
		if output != "difficulty. every in opportunity the sees optimist an opportunity; every in difficulty the sees pessimist A" {
			b.Error("Unexpected result: " + output)
		}
	}
}
```

```go
package main

import (
	"strings"
	"testing"
)

func BenchmarkStringReverseBad(b *testing.B) {
	b.ReportAllocs()

	input := "A pessimist sees the difficulty in every opportunity; an optimist sees the opportunity in every difficulty."

	for i := 0; i < b.N; i++ {
		words := strings.Split(input, " ")
		wordsReverse := make([]string, 0)
		for {
			word := words[len(words)-1:][0]
			wordsReverse = append(wordsReverse, word)
			words = words[:len(words)-1]
			if len(words) == 0 {
				break
			}
		}
		output := strings.Join(wordsReverse, " ")
		if output != "difficulty. every in opportunity the sees optimist an opportunity; every in difficulty the sees pessimist A" {
			b.Error("Unexpected result: " + output)
		}
	}
}

func BenchmarkStringReverseBetter(b *testing.B) {
	b.ReportAllocs()

	input := "A pessimist sees the difficulty in every opportunity; an optimist sees the opportunity in every difficulty."

	for i := 0; i < b.N; i++ {
		words := strings.Split(input, " ")
		for i := 0; i < len(words)/2; i++ {
			words[len(words)-1-i], words[i] = words[i], words[len(words)-1-i]
		}
		output := strings.Join(words, " ")
		if output != "difficulty. every in opportunity the sees optimist an opportunity; every in difficulty the sees pessimist A" {
			b.Error("Unexpected result: " + output)
		}
	}
}
```

这个糟糕的函数做了不必要的分配，即：

1. 我们创建了一个空 slice 存储结果字符串，
2. 我们填充 这个 slice（append 分配内存是必要的，但不是最优的）

由于调用 b.reportAllocs() 这个基准测试和相关的输出绘制了一幅精准的图片：

```
BenchmarkStringReverseBad-4              1413 ns/op             976 B/op          8 allocs/op
BenchmarkStringReverseBetter-4            775 ns/op             480 B/op          3 allocs/op
```

由于在 Go 中实现的虚拟内存，内存分配的另个方面是垃圾收集暂停或简称 GC 。关于 GC 暂停的一个常用语是“停止世界”，注意在 GC 暂停期间您的应用程序将完全停止响应。google 团队不断提升 [GC 的性能](https://groups.google.com/forum/?fromgroups#!topic/golang-dev/Ab1sFeoZg_8)，但将来那些经验不足的开发者仍然会面对内存管理不良的问题。

这个 runtime 包暴露了 runtime.ReadMemStats(m *MemStats) 函数用于填充一个 MemStats 对象。这个对象有很多字段可以作为内存分配策略和性能相关问题的良好指示器。

+ Alloc -当前在堆中分配字节数，
+ TotalAlloc -在堆中累计分配最大字节数（不会减少），
+ Sys -从系统获得的总内存，
+ Mallocs 和 Frees - 分配，释放和存活对象数（mallocs - frees），
+ PauseTotalNs -从应用开始总GC暂停，
+ NumGC - GC 循环完成数

## 方法

因此，我们开始的前提是，我们不希望使用外部服务来提供简单的应用程序监控。我的目标是每隔一段时间将收集到的度量指标打印到控制台上。我们应该启动一个 goroutine，每隔X秒就可以得到这个数据，然后把它打印到控制台。

```go
package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

type Monitor struct {
	Alloc,
	TotalAlloc,
	Sys,
	Mallocs,
	Frees,
	LiveObjects,
	PauseTotalNs uint64

	NumGC        uint32
	NumGoroutine int
}

func NewMonitor(duration int) {
	var m Monitor
	var rtm runtime.MemStats
	var interval = time.Duration(duration) * time.Second
	for {
		<-time.After(interval)

		// Read full mem stats
		runtime.ReadMemStats(&rtm)

		// Number of goroutines
		m.NumGoroutine = runtime.NumGoroutine()

		// Misc memory stats
		m.Alloc = rtm.Alloc
		m.TotalAlloc = rtm.TotalAlloc
		m.Sys = rtm.Sys
		m.Mallocs = rtm.Mallocs
		m.Frees = rtm.Frees

		// Live objects = Mallocs - Frees
		m.LiveObjects = m.Mallocs - m.Frees

		// GC Stats
		m.PauseTotalNs = rtm.PauseTotalNs
		m.NumGC = rtm.NumGC

		// Just encode to json and print
		b, _ := json.Marshal(m)
		fmt.Println(string(b))
	}
}
```

要使用它，你可以用 go NewMonitor(300) 来调用它，它每5分钟打印一次你的应用程序度量。然后，您可以从控制台或历史日志中检查这些，以查看应用程序的行为。将其添加到应用程序中的任何性能影响都很小。

```
{"Alloc":1143448,"TotalAlloc":1143448,"Sys":5605624,"Mallocs":8718,"Frees":301,"LiveObjects":8417,"PauseTotalNs":0,"NumGC":0,"NumGoroutine":6}
{"Alloc":1144504,"TotalAlloc":1144504,"Sys":5605624,"Mallocs":8727,"Frees":301,"LiveObjects":8426,"PauseTotalNs":0,"NumGC":0,"NumGoroutine":5}
...
```

我认为控制台中的这些输出是一个有用的洞察力，它会让你知道在不久的将来可能会碰到一些问题。

## 使用 expvar

Go 实际上有两个内置插件，帮助我们监控生产中的应用程序。其中一个内置的是包 expvar。该包为公共变量提供了标准化接口，例如服务器中的操作计数器。默认情况下，这些变量将在 `/debug/vars` 上可用。让我们把度量放在 expvar 中存储。

几分钟后，我注册了 expvar 的 HTTP 处理程序，我意识到完整的 MemStats 结构已经在上面了。那太好了！

除了添加HTTP处理程序外，此包还记录以下变量：

+ cmdline os.Args
+ memstats runtime.Memstats

该包有时仅用于注册其HTTP处理程序和上述变量的副作用。要这样使用，把这个包链接到你的程序中：`import _ "expvar"`

由于度量现在已经导出，您只需要在应用程序上指向监视系统，并在那里导入 memstats 输出。我知道，我们仍然没有 goroutine 计数，但这很容易添加。导入 expvar 包并添加以下几行：

```go
// The next line goes at the start of NewMonitor()
var goroutines = expvar.NewInt("num_goroutine")
// The next line goes after the runtime.NumGoroutine() call
goroutines.Set(int64(m.NumGoroutine))
```

这个 “num_goroutine” 字段现在在 `/debug/vars output` 可用, 仅此于完整的内存统计。

## 超越基础监测

Go 标准库中的另外一个强大的补充是 [net/http/pprof](https://golang.org/pkg/net/http/pprof/) 包。这个包有很多函数，但主要目的是为 `go pprof` 工具提供运行时的分析数据，该工具已捆绑在 Go 工具链中。使用它，您可以进一步检查您的应用程序在生产中的操作。如果您想了解更多关于 pprof 和代码优化的内容，您可以查看我以前的文章：

+ [Go 程序基准测试](https://scene-si.org/2017/06/06/benchmarking-go-programs/),
+ [Go 程序基准测试，第二部](https://scene-si.org/2017/07/07/benchmarking-go-programs-part-2/)

并且，如果您想要 Go 程序持续分析，可以用 Google 的一个服务，StackDriver 分析器。但是，不管什么原因如果您想监视您自己的基础设施，[Prometheus](https://prometheus.io/) 可能是最好的选择。如果您想看的话，请在下面输入您的电子邮件。

## 看这里……

如果您能买我的书那就太好了：

+ [API Foundations in Go](https://leanpub.com/api-foundations)
+ [12 Factor Apps with Docker and Go](https://leanpub.com/12fa-docker-golang)
+ [The SaaS Handbook(work in progress)](https://leanpub.com/saas-handbook)

我保证如果您买任何一本的话您可以学到很多东西。购买副本支持我写更多关于相同的主题。感谢您买我的书。

如果您想和我约时间为了咨询/外包服务可以发[电子邮件](black@scene-si.org)给我。我很擅长 API，Go，Docker，VueJS 和 扩展服务等等。

---

via: https://scene-si.org/2018/08/06/basic-monitoring-of-go-apps-with-the-runtime-package/

作者：[Tit Petric](https://scene-si.org/)
译者：[themoonbear](https://github.com/themoonbear)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
