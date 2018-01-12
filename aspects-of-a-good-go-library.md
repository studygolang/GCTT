

# Go 佳库面面观

从一个好的 Go 库里，我希望得到的东西的一个简短清单（排名不分先后）。这是[高效 Go（effective go）](https://golang.org/doc/effective_go.html)列表、[Go 代码评审意见](https://github.com/golang/go/wiki/CodeReviewComments)列表和[Go 箴言](https://go-proverbs.github.io/)列表伴侣。

一般来说，当做某事有两种合理的方式的时候，选择不违反这些规则的那一项。只有在有非常强力的理由时才违反这些规则。

### 依赖

#### 加标签的库版本

使用 git 标签来管理你的库版本。语义版本化是一个合理的系统。如果你对语义版本化的[反对在意](https://news.ycombinator.com/item?id=13378637)，那么，你就不是这的目标用户 :)

#### 没有非标准库依赖

这个有时候会难以达成，但是，管理一个库的依赖使得升级到最新版本变得轻松，并且允许库用户更好地得出应用中的逻辑。[如果你将依赖树维持得相当相当小，那么，你就真的可以让你的程序维护更简单。](https://youtu.be/PAAkCSZUG1c?t=598)

#### 抽象非标准库依赖到它们自己的包里

这是上面非标准库依赖要求的必然结果。如果你的库绝对需要非标准库依赖，那么，试着将其分成两个包：一个包用于核心逻辑，而另一个包拥有外部依赖，使用前一个包的逻辑。

打个比方，如果你正在写一个包，它捕获 Go 的堆栈追踪，然后将其上传到 [Amazon’s S3](https://aws.amazon.com/s3/)，那么，写两个包。

  1. 一个包捕获 Go 的堆栈追踪，然后将其交到一个接口
  2. 第一个包使用的接口的 S3 实现。

第二个包可以为用户简化两部分的粘合。这种抽象层次让用户利用你的核心功能，同时用自己的存储层来替换 S3。而在第二个包中的粘合操作，可以避免给那些想要将他们的数据上传到 S3 的大部分人增加额外的负担。

这里的关键部分是两个**分开的**包。这使得用户可以注入你的核心逻辑，而不会污染到他们自己的依赖树。

#### 不要使用 /vendor

If your library vendors its dependencies, strange side effects can happen
around package managers that try to flatten vendors, or vendor polluting your
public API space making integrations difficult or impossible. If you really
need an exact implementation, [copy it explicitly](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=9m28s).

#### 使用依赖管理，这样别人就可以跑你的测试了

你的库应该尝试不要有外部依赖。但是，如果必须有，那么使用某种类型的依赖管理，传达初始运行你的测试的依赖版本，从而使得库用户可以以一种一致的方式运行你的单元测试。

### API

#### 没有全局可变状态

[全局可变状态](https://dev.to/ericnormand/global-mutable-state) makes
code difficult to reason about, expand, stub, test, and back out of.

#### 空对象具有合理的行为

让[零值有用](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=6m25s)。

#### nil 实例上的读操作与空实例上的读操作行为一致

许多 Go 的内部结构[反映](https://play.golang.org/p/LrkSriba5V)了这种行为。

#### 若非必要，避免构造器函数

这是让[零值有用](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=6m25s)的副作用。

#### 最小化公共函数

做更少事情的库倾向于做得更好。

#### 拥有少数函数的小接口

[接口越大，抽象越弱](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=5m17s).

#### 接受接口，返回结构

阅读[这里](https://medium.com/@cep21/what-accept-interfaces-return-structs-means-in-go-2fe879e25ee8)，获取更多细节。

#### Configuration mutable at runtime without violating -race

If your library maintains complex state where it’s not a simple process to
just re instantiate it, allow users to modify reasonable configuration
parameters while their application is running.

#### 我可以接口化，并且无需导入你的库的 API

你的库很棒，但是终有一天我会想要将其淘汰出去。Go 的类型系统[有时会造成阻碍](https://medium.com/statuscode/go-experience-report-gos-type-system-c4d4dfcc964c)。然而，总的来说，隐式接口胜于过于强大的静态类型。如果有可能的话，最好将标准库类型作为函数参数类型和返回值类型，这样，用户就可以根据你的结构创建接口，以便于后面进行库替换。

```go

    type AvoidThis struct {}  
    type Key string  
    func (a *AvoidThis) Convert(k Key) {... }
```

```go

    type PreferThis struct {}  
    func (p *PreferThis) Convert(k string) { ... }
```

#### API 调用时创建最小的对象（GC）

CPU 通常是避无可避的，但是，重新考虑你的 API 会使得最小化 API 调用期间的垃圾回收成为可能。例如，创建不强制垃圾回收的 API。**事后优化实现很容易，但是事后优化 API 则几乎不可能**。

```go

    type AvoidThis struct {}  
    func (a *AvoidThis) Bytes() []byte { ... }
```

```go

    type PreferThis struct {}  
    func (p *PreferThis) WriteTo(w Writer) (n int64, err error) { ... }
```

#### 无副作用导入

I personally disagree with the [Go](https://golang.org/pkg/expvar/) [标准库](https://golang.org/pkg/net/http/pprof/) pattern of creating [global side effect behavior](https://golang.org/doc/effective_go.html#blank_import)
based upon an import. This behavior usually involves global mutable state, can
cause funny issues when libraries are /vendor included multiple times, and
removes many options for customization.

#### 当存在替代方法的时候，避免使用 context.Value

[我前一篇文章](https://medium.com/@cep21/how-to-correctly-use-context-context-in-go-1-7-8f2c0fafdf39)上有关于这个想法的扩展。

#### init 中避免使用复杂逻辑

[init 函数](https://golang.org/doc/effective_go.html#init)对默认值的创建很有用，但是也是用户不可能自定义或忽略的逻辑。如果没有好的理由的话，不要从用户那里拿走你的库的控制权。因此，避免在 init 中生成后台 goroutine，而是最好让用户明确地要求后台行为。

#### 允许注入全局依赖

例如，当允许用户提供 `http.Client` 时，不要强制使用 `http.DefaultClient`。

### 错误

#### 检查所有错误

如果你的库接受一个接口作为输入，而有人给了你一个返回错误的实现，那么用户会期望你会检查并以某种方式处理错误，或者以某种方式把它传回给调用者。检查错误不只是意味着[将其返回给上层调用栈](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=17m25s)，尽管这样有时是合理的，但是，你可以记录它，改变返回值并将其作为结果，使用[后备代码](https://github.com/Netflix/Hystrix/wiki/How-To-Use#Fallback)，或者只是增加一个内部统计计数器，这样，用户就会知道发生了**一些事情**，而不是在不知道有什么不对的情况下失败。

#### 通过行为暴露错误，而不是类型

This is the library centric equivalent of Dave的 _Assert errors for behaviour, not type。 _更多信息，看[这里](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)_。_

#### 不要 panic

[就是不要](https://github.com/golang/go/wiki/CodeReviewComments#dont-panic)

### 并发

#### 避免创建 goroutine

This is a more explicit rule reasoned by the [CodeReviewComments synchronous functions](https://github.com/golang/go/wiki/CodeReviewComments#synchronous-functions) section. Synchronous functions give the library user more control.
Goroutines are sometimes useful to parallelize logic, but as a library author
you should **start** from the state of not having goroutines and reasoning
your way into them, rather than starting from goroutines and being argued away
from them.

#### 允许后台 goroutine 干净地停止

这是 [goroutine 生命周期](https://github.com/golang/go/wiki/CodeReviewComments#goroutine-lifetimes) 反馈的首选限制。应该有一种方式，以一种不会发出虚假错误的方式，结束你的库创建的任意 goroutine。

#### 公开 API 中避免使用 channel

这是一种代码异味，[暗示并发](https://github.com/golang/go/wiki/CodeReviewComments#synchronous-functions)发生在库级别，而不是让你的库用户控制并发。

#### 所有长的阻塞操作都采用 context.Context

Context 是让你的库用户控制何时应该中断操作的标准方式。

### 调试

#### 导出内部统计信息

我需要监控你的库的效率、使用模式和耗时。以某种方式公开这些统计数据，这样，我就可以将其导入到我[最爱的](https://prometheus.io/) [度量](https://signalfx.com/)[系统](https://grafana.com/)中。

#### 公开 expvar.Var 信息

通过 expvar 公开内部配置和状态信息，从而允许用户快速调试how their application is using your library, not just **how
they think** they are using it.

#### 支持可调试性

Eventually your library will have a bug. Or the user will use your library
incorrectly and need to figure out why. If your library has any reasonable
amount of complexity, expose a way to debug or trace this information. This
could be with debug logs or the [context debug
pattern](https://medium.com/@cep21/go-1-7-httptrace-and-context-debug-
patterns-608ae887224a).

#### 合理的 Stringer 实现

[Stringer](https://golang.org/pkg/fmt/#Stringer) 是的人们更容易用你的库来调试代码。

#### 轻松可定制的日志器

没有广泛接受的 Go 日志库。公开一个日志接口，从而不强制我导入你最爱的日志库。

### 清洁度

#### Passes a reasonable subset of gometalinter checks

Go’s simple syntax and great standard library functions allows a wide array of
static code checkers, which are aggregated in the
[gometalinter](https://github.com/alecthomas/gometalinter). Your default
state, especially if you’re new to Go, should be to just pass them all. Bend
them only if you can explain why, and given two reasonable implementations
defer to the one that passes the linter.

#### No functions with 0% unit test coverage

100% test coverage is extreme and 0% test coverage is almost never a good
thing. This is difficult to quantify into a rule, and I’ve settled upon no
function should have 0% test coverage as a minimum bar. You can get per
function test coverage using Go’s cover tool.

```go

    # go test -coverprofile=cover.out context  
    ok   context 2.651s coverage: 97.0% of statements
```

```go

    # go tool cover -func=cover.out  
    context/context.go:162: Error  100.0%  
    context/context.go:163: Timeout  100.0%  
    context/context.go:164: Temporary 100.0%  
    context/context.go:170: Deadline 100.0%  
    context/context.go:174: Done  100.0%  
    context/context.go:178: Err  100.0%  
    ...
```

### 存储库布局

#### 避免将一个结构的函数拆分到多个文件中

Go 允许你把一个结构的函数放到多个文件中。这在使用[构建标志](https://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool)时，是非常有用的，但是，如果你把它作为组织结构的方式，那么，这说明你的结构太大了，应该把它分解成多个部分。

#### 使用 /internal

/internal 包严重使用不足。我推荐二进制文件和库都利用 /internal 来隐藏不打算导入的公共函数。隐藏你的公共导入空间也使得用户更清楚应该导入哪些包，以及要到哪里寻找有用的逻辑。

----------------

via: https://medium.com/@cep21/aspects-of-a-good-go-library-7082beabb403

作者：[Jack Lindamood](https://medium.com/@cep21)
译者：[ictar](https://github.com/ictar)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出