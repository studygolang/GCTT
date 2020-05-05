首发于：https://studygolang.com/articles/28435

# Go 各版本回顾

![Illustration created for “A Journey With Go”, made from *the original Go Gopher, created by Renee French.](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/00.png)

对每一个开发者来说，Go 的发展历史是必须知道的知识。了解几年来每个发行版本的主要变化，有助于理解 Go 的设计思想和每个版本的优势/弱点。想了解特定版本的更详细信息，可以点击每个版本号的链接来查看修改记录。

## [Go 1.0](https://blog.golang.org/go-version-1-is-released) — 2012 年 3 月：

Go 的第一个版本，带着一份[兼容性说明文档](https://golang.org/doc/go1compat)来保证与未来发布版本的兼容性，进而不会破坏已有的程序。

第一个版本已经有 `go tool pprof` 命令和 `go vet` 命令。 `go tool pprof` 与 Google 的 [pprof C++ profiler](https://github.com/gperftools/gperftools) 稍微有些差异。`go vet`（前身是 `go tool vet`）命令可以检查包中潜在的错误。

## [Go 1.1](https://blog.golang.org/go-11-is-released) — 2013 年 5 月：

这个 Go 版本专注于优化语言（编译器，gc，map，go 调度器）和提升它的性能。下面是一些提升的例子：

![https://dave.cheney.net/2013/05/21/go-11-performance-improvements](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/01.png)

这个版本也嵌入了[竞争检测](https://blog.golang.org/race-detector)，在语言中是一个很强大的工具。

*你可以在我的文章 [ThreadSanitizer 竞争检测](https://medium.com/a-journey-with-go/go-race-detector-with-threadsanitizer-8e497f9e42db)中发现更多的信息。*

[重写了 Go 调度器](https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw/edit) 显著地提升了性能。Go 调度器现在被设计成这样：

![https://rakyll.org/scheduler/](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/02.png)

`M` 是一个 OS 线程，`P` 表示一个处理器（`P` 的数量不能大于 GOMAXPROCS），每个 `P` 作为一个本地协程队列。在 1.1 之前 `P` 不存在，协程是用一个全局的 mutex 在全局范围内管理的。随着这些优化，工作窃取也被实现了，允许一个 `P` 窃取另一个 `P` 的协程：

![https://rakyll.org/scheduler/](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/03.png)

*阅读 [Jaana B.Dogan](https://twitter.com/rakyll) 的 [Go 的工作窃取调度器](https://rakyll.org/scheduler/) 可以查看更多关于 Go 调度器和工作窃取的信息。*

## [Go 1.2](https://blog.golang.org/go12) — 2013 年 12 月：

本版本中 `test` 命令支持测试代码覆盖范围并提供了一个新命令 `go tool cover` ，此命令能测试代码覆盖率：

![https://blog.golang.org/cover](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/04.png)

这个命令也能提供覆盖信息：

![https://blog.golang.org/cover](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/05.png)

## [Go 1.3](https://blog.golang.org/go1.3) — 2014 年 6 月：

这个版本对栈管理做了重要的改进。栈可以申请[连续的内存片段](https://docs.google.com/document/d/1wAaf1rYoM4S4gtnPh0zOlGzWtrZFQ5suE8qr2sD8uWQ/pub)，提高了分配的效率，使下一个版本的栈空间降到 2KB。

栈频繁申请/释放栈片段会导致某些元素变慢，本版本也改进了一些由于上述场景糟糕的分配导致变慢的元素。下面是一个 `json` 包的例子，展示了它对栈空间的敏感程度：

![https://docs.google.com/document/d/1wAaf1rYoM4S4gtnPh0zOlGzWtrZFQ5suE8qr2sD8uWQ/pub](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/06.png)

使用连续的栈修复了这个元素效率低下的问题。下面是另一个例子，`html/template` 包的性能对栈大小也很敏感：

![图 7](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/07.png)

*阅读我的[Go 协程栈空间的发展](https://medium.com/a-journey-with-go/go-how-does-the-goroutine-stack-size-evolve-447fc02085e5)查看更多信息。*

这个版本在 `sync` 包中发布了 `Pool`。 这个元素允许我们复用结构体，减少了申请的内存的次数，同时也是很多 Go 生态获得改进的根源，如标准库或包里的 `encoding/json` 或 `net/http`，还有 Go 社区里的 `zap`。

*可以在我的文章 [Sync.Pool 设计理解](https://medium.com/@blanchon.vincent/go-understand-the-design-of-sync-pool-2dde3024e277) 中查看更多关于 `Pool` 的信息。*

Go 团队也[对通道作了改进](https://docs.google.com/document/d/1yIAYmbvL3JxOKOjuCyon7JhW4cSv1wy5hC0ApeGMV9s/pub)，让它们变得更快。下面是以 Go 1.2 和 Go 1.3 作对比运行的基准：

![图 8](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/08.png)

## [Go 1.4](https://blog.golang.org/go1.4) — 2014 年 12 月：

此版本带来了官方对 Android 的支持，[golang.org/x/mobile]([Go 1.4](https://blog.golang.org/go1.4) ) 让我们可以只用 Go 代码就能写出简单的 Android 程序。

归功于更高效的 gc，之前用 C 和汇编写的运行时代码被翻译成 Go 后，堆的大小降低了 10% 到 30%。

与版本无关的一个巧合是，Go 项目管理从 Mercurial 移植到了 Git，代码从 Google Code 移到了 Github。

Go 也提供了 `go generate` 命令通过扫描用 `//go:generate` 指示的代码来简化代码生成过程。

*在 [Go 博客](https://blog.golang.org/) 和文章[生成代码](https://blog.golang.org/generate)中可以查看更多信息。*

## [Go 1.5](https://blog.golang.org/go1.5) — 2015 年 8 月：

这个新版本，[发布时间推迟](https://docs.google.com/document/d/106hMEZj58L9nq9N9p7Zll_WKfo-oyZHFyI6MttuZmBU/edit#)了两个月，目的是在以后每年八月和二月发布新版本：

![https://github.com/golang/go/wiki/Go-Release-Cycle](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/09.png)

这个版本对 [gc](https://golang.org/doc/go1.5#gc) 进行了[重新设计](https://docs.google.com/document/d/1wmjrocXIWTr1JxU-3EQBI6BK6KgtiFArkG47XK73xIQ/edit#)。归功于并发的回收，在回收期间的等待时间大大减少。下面是一个 Twitter 生产环境的服务器的例子，等待时间由 300ms 降到 30ms：

![https://blog.golang.org/ismmkeynote](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/10.png)

这个版本也发布了运行时追踪，用命令 `go tool trace` 可以查看。测试过程或运行时生成的追踪信息可以用浏览器窗口展示：

![[Original Go Execution Tracer Document](https://docs.google.com/document/d/1FP5apqzBgr7ahCCgFO-yoVhk4YZrNIDNf9RybngBc14/pub)](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/11.png)

## [Go 1.6](https://blog.golang.org/go1.6) — 2016 年 2 月:

这个版本最重大的变化是使用 HTTPS 时默认支持 HTTP/2。

在这个版本中 gc 等待时间也降低了：

![https://blog.golang.org/ismmkeynote](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/12.png)

## [Go 1.7](https://blog.golang.org/go1.7) — 2016 年 8 月:

这个版本发布了 [context 包](https://medium.com/a-journey-with-go/go-context-and-cancellation-by-propagation-7a808bbc889c)，为用户提供了处理超时和任务取消的方法。

*阅读我的文章 [传递上下文和取消](https://medium.com/a-journey-with-go/go-context-and-cancellation-by-propagation-7a808bbc889c)来获取更多关于 context 的信息。*

对编译工具链也作了优化，编译速度更快，生成的二进制文件更小，有时甚至可以减小 20% 到 30%。

## [Go 1.8](https://blog.golang.org/go1.8) — 2017 年 2 月:

把 gc 的停顿时间减少到了 1 毫秒以下：

![https://blog.golang.org/ismmkeynote](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/13.png)

其他的停顿时间已知，并会在下一个版本中降到 100 微秒以内。

这个版本也改进了 defer 函数：

![https://medium.com/@blanchon.vincent/go-how-does-defer-statement-work-1a9492689b6e](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/14.png)

*我的文章 [defer 语句工作机制](https://medium.com/a-journey-with-go/go-how-does-defer-statement-work-1a9492689b6e)中有更多信息。*

## [Go 1.9](https://blog.golang.org/go1.9) — 2017 年 8 月:

这个版本支持下面的别名声明：

```go
type byte = uint8
```

这里 `byte` 是 `uint8` 的一个别名。

`sync` 包新增了一个 [Map](https://golang.org/pkg/sync/#Map) 类型，是并发写安全的。

*我的文章 [Map 与并发写](https://medium.com/a-journey-with-go/go-concurrency-access-with-maps-part-iii-8c0a0e4eb27e) 中有更多信息。*

## [Go 1.10](https://blog.golang.org/go1.10) — 2018 年 2 月:

`test` 包引进了一个新的智能 cache，运行会测试后会缓存测试结果。如果运行完一次后没有做任何修改，那么开发者就不需要重复运行测试，节省时间。

```bash
first run:
ok      /go/src/retro 0.027s
second run:
ok      /go/src/retro (cached)
```

为了加快构建速度，`go build` 命令现在也维持了一份最近构建包的缓存。

这个版本没有对 gc 做实际的改变，但是确定了一个新的 SLO（Service-Level Objective）：

![https://blog.golang.org/ismmkeynote](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/15.png)

## [Go 1.11](https://blog.golang.org/go1.11) — 2018 年 8 月:

Go 1.11 带来了一个重要的新功能：[Go modules](https://blog.golang.org/using-go-modules)。去年的调查显示，Go modules 是 Go 社区遭遇重大挑战后的产物：

![https://blog.golang.org/survey2018-results](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/16.png)

第二个特性是实验性的 [WebAssembly](https://webassembly.org/)，为开发者提供了把 Go 程序编译成一个可兼容四大主流  Web 浏览器的二进制格式的能力。

## [Go 1.12](https://blog.golang.org/go1.12) — 2019 年 2 月:

基于 `analysis` 包重写了 `go vet` 命令，为开发者写自己的检查器提供了更大的灵活性。

*我的文章[构建自己的分析器](https://medium.com/@blanchon.vincent/go-how-to-build-your-own-analyzer-f6d83315586f)中有更多信息。*

## [Go 1.13](https://blog.golang.org/go1.13) — 2019 年 9 月:

改进了 `sync` 包中的 `Pool`，在 gc 运行时不会清除 pool。它引进了一个缓存来清理两次 gc 运行时都没有被引用的 pool 中的实例。

重写了逃逸分析，减少了 Go 程序中堆上的内存申请的空间。下面是对这个新的逃逸分析运行基准的结果：

![https://github.com/golang/go/issues/23109](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Retrospective/17.png)

## [Go1.14](https://blog.golang.org/go1.14) - 2020 年 2 月：

现在 Go Module 已经可以用于生产环境，鼓励所有用户迁移到 Module。该版本支持嵌入具有重叠方法集的接口。性能方面做了较大的改进，包括：进一步提升 defer 性能、页分配器更高效，同时 timer 也更高效。

现在，Goroutine 支持异步抢占。

---

via: https://medium.com/a-journey-with-go/go-retrospective-b9723352e9b0

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
