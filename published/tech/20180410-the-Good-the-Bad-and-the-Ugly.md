已发布：https://studygolang.com/articles/12907

# Go 语言的优点，缺点和令人厌恶的设计

这是关于 「[Go是一门设计糟糕的编程语言 （Go is not good)](https://github.com/ksimka/go-is-not-good)」 系列的另一篇文章。Go 确实有一些很棒的特性，所以我在这篇文章中展示了它的优点。但是总体而言，当超过 API 或者网络服务器（这也是它的设计所在）的范畴，用 Go 处理商业领域的逻辑时，我感觉它用起来麻烦而且痛苦。就算在网络编程方面，Go 的设计和实现也存在诸多问题，这使它看上去简单实际则暗藏危险。

写这篇文章的动机是因为我最近重新开始用 Go 写一个业余项目。在以前的工作中我广泛的使用了 Go 为 SaaS 服务编写网络代理（包括 http 和原始的 tcp）。网络编程的部分是相当令人愉快的（我也正在探索这门语言），但随之而来的会计和账单部分则苦不堪言。因为我的业余项目只是一个简单的 API，我认为 Go 非常适合快速的完成这个任务。但是我们都知道，很多项目的增长会超过了预期的范围，所以我不得不写一些数据处理来计算统计数据，Go 的痛苦之处也随着而来。下面就是我对 Go 的困扰。

一些背景情况：我喜欢静态类型的语言。我第一个标志性的项目是用 [Pascal](https://zh.wikipedia.org/wiki/Pascal_(%E7%A8%8B%E5%BC%8F%E8%AA%9E%E8%A8%80)) 写的。当90年代初我开始工作之后，开始使用 [Ada](https://zh.wikipedia.org/wiki/Ada) 和 C/C++。后来我转移到 Java 阵地，最后到了 Scala（中间夹杂着 Go），最近开始学习 [Rust](https://www.rust-lang.org/zh-CN/)。我也写了大量的 JavaScript，因为直到现在它依旧是浏览器端唯一可用的语言。我感觉动态类型的语言并不安全，并尽力将它们的使用限制在脚本级别。我习惯了命令式，函数式和面向对象的方法。

这是一篇很长的文章，所以我列出了菜单来“激发你的食欲”：

- [优点](#good)
  - [Go 很容易学习](#go-easy-learn)
  - [基于 goroutines 和 channels 的简单并发编程](#easy-concurrent)
  - [丰富的标准库](#greate-standard-library)
  - [Go 性能优越](#go-performant)
  - [语言层面定义源代码的格式化](#defined-code-format)
  - [标准化的测试框架](#test-framework)
  - [Go 程序方便操作](#great-operations)
  - [Defer 声明，避免忘记清理](#defer)
  - [新类型](#new-type)
- [缺点](#bad)
  - [Go 忽略了现代语言设计的进步](#ignore-advances)
  - [接口是结构类型](#interfaces-types)
  - [没有枚举](#no-enum)
  - [:= / var 两难选择](#dilemma)
  - [零值会导致 panic](#zero-panic)
  - [Go 没有异常，Emmmm 等等... 它有](#exceptions)！
- [令人厌恶的点](#ugly)
  - [依赖管理的噩梦](#dependency-nightmare)
  - [易变性被语言硬编码](#mutability-hardcode)
  - [切片（slice）陷阱](#slice-gotchas)
  - [易变性和 channels: 竞争条件更容易发生](#race-conditions)
  - [嘈杂的错误管理](#error-management)
  - [Nil 接口值](#nil-interface-values)
  - [Struct 字段标记：DSL 在字符串中的运行时间](#dsl)
  - [没有泛型...至少没给你](#generic)
  - [Go 在 slice 和 map 之外几乎没有别的数据结构](#no-structures)
  - [go generate，还说得过去，但是...](#generate)
- [结论](#conclusion)
- [几天后: Hacker News 第三名!](#hacker-news)

## <span id="good">优点</span>

### <span id="go-easy-learn">Go 很容易学习</span>

这是事实：如果你了解任何一种编程语言，那么通过在「[Go 语言之旅](http://go-tour-zh.appspot.com/welcome/1)」学习几个小时就能够掌握 Go 的大部分语法，并在几天后写出你的第一个真正的程序。阅读并理解 [实效 Go 编程](http://docscn.studygolang.com/doc/effective_go.html)，浏览一下「[包文档](http://docscn.studygolang.com/pkg/)」,玩一玩 [Gorilla](http://www.gorillatoolkit.org/) 或者 [Go Kit](https://gokit.io/) 这样的网络工具包，然后你将成为一个相当不错的 Go 开发者。

这是因为 Go 的首要目标是简单。当我开始学习 Go，它让我想起我第一次 [发现 Java](https://zh.wikipedia.org/wiki/Java%E7%89%88%E6%9C%AC%E6%AD%B7%E5%8F%B2)：一个简单的语言和一个丰富但不臃肿的标准库。对比当前 Java 沉重的环境，学习 Go 是一个耳目一新的体验。因为 Go 的简易性，Go 程序可读性非常高，虽然错误处理添加了一些麻烦（更多的内容在下面）。

Go 语言的简单可能是错误的。引用 Rob Pike 的话，[简单既是复杂](https://talks.golang.org/2015/simplicity-is-complicated.slide#1)，我们会看到简单背后有很多的陷阱等着我们去踩，极简主义会让我们违背 DRY(Don't Repeat Yourself) 原则。

### <span id="easy-concurrent">基于 goroutines 和 channels 的简单并发编程</span>

Goroutines 可能是 Go 的最佳特性了。它们是轻量级的计算线程，与操作系统线程截然不同。

当 Go 程序执行看似阻塞 I/O 的操作时，实际上 Go 运行时挂起了 goroutine ,当一个事件指示某个结果可用时恢复它。与此同时，其他的 goroutines 已被安排执行。因此在同步编程模型下，我们具有了异步编程的可伸缩性优势。

Goroutines 也是轻量级的:它们的堆栈 [随需求增长和收缩](https://dave.cheney.net/2013/06/02/why-is-a-goroutines-stack-infinite)，这意味着有 100 个甚至 1000 个 goroutines 都不是问题。

我以前的应用程序中有一个 goroutine 漏洞:这些 goroutines 结束之前正在等待一个 channel 关闭，而这个 channel 永远不会关闭(一个常见的死锁问题)。这个进程毫无任何理由吃掉了 90 % 的 CPU ，而检查 [expvars](http://docscn.studygolang.com/pkg/expvar/) 显示有 600 k 空闲的 goroutines! 我猜测 goroutine 调度程序占用了 CPU。

当然，像 Akka 这样的 Actor 系统可以轻松 [处理数百万的 Actors](https://doc.akka.io/docs/akka/2.5/general/actor-systems.html#what-you-should-not-concern-yourself-with)，部分原因是 actors 没有堆栈，但是他们远没有像 goroutines 那样简单地编写大量并发的请求/响应应用程序（即 http APIs）。

channel 是 goroutines 的通信方式:它们提供了一个便利的编程模型，可以在 goroutines 之间发送和接收数据，而不必依赖脆弱的低级别同步基本体。channels 有它们自己的一套 [用法](https://blog.golang.org/pipelines) [模式](https://blog.golang.org/go-concurrency-patterns-timing-out-and)。

但是，channels 必须仔细考虑，因为错误大小的 channels (默认情况下没有缓冲) [会导致死锁](https://www.danmrichards.com/blog/2018/03/26/goroutines-channels-and-waitgroups/)。下面我们还将看到，使用通道并不能阻止竞争情况，因为它缺乏不可变性。

### <span id="greate-standard-library">丰富的标准库</span>

Go 的 [标准库](http://docscn.studygolang.com/pkg/) 非常丰富,特别是对于所有与网络协议或 API 开发相关的: http 客户端和服务器，加密，档案格式，压缩，发送电子邮件等等。甚至还有一个html解析器和相当强大的模板引擎去生成 text & html，它会自动过滤 XSS 攻击（例如在 [Hugo](https://gohugo.io/templates/introduction/) 中的使用）。

各种 APIs 一般都简单易懂。它们有时看起来过于简单:这个某种程度上是因为 goroutine 编程模型意味着我们只需要关心“看似同步”的操作。这也是因为一些通用的函数也可以替换许多专门的函数，就像 [我最近发现的关于时间计算的问题](https://bluxte.net/musings/2018/03/22/local-date-time-calculations-in-go/)。

### <span id="go-performant">Go 性能优越</span>

Go 编译为本地可执行文件。许多 Go 的用户来自 Python、Ruby 或 Node.js。对他们来说，这是一种令人兴奋的体验，因为他们看到服务器可以处理的并发请求数量大幅增加。当您使用非并发(Node.js)或全局解释器锁定的解释型语言时，这实际上是相当正常的。结合语言的简易性，这解释了 Go 令人兴奋的原因。

然而与 Java 相比，在 [原始性能基准测试](https://www.techempower.com/benchmarks/) 中，情况并不是那么清晰。Go 打败 Java 地方是内存使用和垃圾回收。

Go 的垃圾回收器的设计目的是 [优先考虑延迟](https://blog.golang.org/go15gc)，并避免停机，这在服务器中尤其重要。这可能会带来更高的 CPU 成本，但是在水平可伸缩的体系结构中，这很容易通过添加更多的机器来解决。请记住，Go 是由谷歌设计的，他们从不会在资源上面短缺。

与 Java 相比，Go 的垃圾回收器（GC）需要做的更少:切片是一个连续的数组结构，而不是像 Java 那样的指针数组。类似地，Go maps 也使用[小数组作为 buckets](https://golang.org/src/runtime/hashmap.go)，以实现相同的目的。这意味着垃圾回收器的工作量减少，并且 CPU 缓存本地化也更好。

Go 同样在命令行实用程序中优于 Java :作为本地可执行文件，Go 程序没有启动消耗，反之 Java 首先需要加载和编译的字节码。

### <span id="defined-code-format">语言层面定义源代码的格式化</span>

我职业生涯中一些最激烈的辩论发生在团队代码格式的定义上。 Go 通过为代码定义规范格式来解决这个问题。 `gofmt` 工具会重新格式化您的代码，并且没有选项。

不管你喜欢与否，`gofmt` 定义了如何对代码进行格式化，一次性解决了这个问题。

### <span id="test-framework">标准化的测试框架</span>

Go 在其标准库中提供了一个很好的 [测试框架](http://docscn.studygolang.com/pkg/testing/)。它支持并行测试、基准测试，并包含许多实用程序，可以轻松测试网络客户端和服务器。

### <span id="great-operations">Go 程序方便操作</span>

与 Python，Ruby 或 Node.js 相比，必须安装单个可执行文件对于运维工程师来说是一个梦想。 随着越来越多的 Docker 的使用，这个问题越来越少，但独立的可执行文件也意味着小型的 Docker 镜像。

Go还具有一些内置的观察性功能，可以使用 [`expvar`](http://docscn.studygolang.com/pkg/expvar/) 包发布内部状态和指标，并易于添加新内容。但要小心，因为它们在默认的 http 请求处理程序中 [自动公开](http://docscn.studygolang.com/pkg/expvar/#pkg-overview)，不受保护。Java 有类似的 JMX ，但它要复杂得多。

### <span id="defer">Defer 声明，防止忘记清理</span>

defer 语句的目的类似于 Java 的 `finally`：在当前函数的末尾执行一些清理代码，而不管此函数如何退出。`defer` 的有趣之处在于它跟代码块没有联系，可以随时出现。这使得清理代码尽可能接近需要清理的代码:

```go
file, err := os.Open(fileName)
if err != nil {
	return
}
defer file.Close()

// 用文件资源的时候，我们再也不需要考虑何时关闭它
```
当然，Java的 [试用资源](https://docs.oracle.com/javase/tutorial/essential/exceptions/tryResourceClose.html) 没那么冗长，而且 Rust 在其所有者被删除时会 [自动声明资源](https://doc.rust-lang.org/rust-by-example/trait/drop.html)，但是由于 Go 要求您清楚地了解资源清理情况，因此让它接近资源分配很不错。

### <span id="new-type">新类型</span>

我喜欢类型，因为有些事情让我感到恼火和害怕，举个例子，我们到处把持久对象标识符当做 `string` 或 `long` 类型传递使用。 我们通常会在参数名称中对 id 的类型进行编码，但是当函数具有多个标识符作为参数并且某些调用不匹配参数顺序时，会造成细微的错误。

Go 对新类型有一等支持，即类型为现有类型并赋予其独立身份，与原有类型不同。 与包装相反，新类型没有运行时间开销。 这允许编译器捕捉这种错误：

```go
type UserId string // <-- new type
type ProductId string

func AddProduct(userId UserId, productId ProductId) {}

func main() {
	userId := UserId("some-user-id")
	productId := ProductId("some-product-id")

	// 正确的顺序： 没有问题
	AddProduct(userId, productId)

	// 错误的顺序：将会编译错误
	AddProduct(productId, userId)
	// 编译错误：
	// AddProduct 不能用 productId(type ProductId) 作为 type UserId的参数
	// Addproduct 不能用 userId(type UserId) 作为type ProfuctId 的参数
}
```
不幸的是，缺乏泛型使得使用新类型变得麻烦，因为为它们编写可重用代码需要从原始类型转换值。

## <span id="bad">缺点</span>

### <span id="ignore-advances">Go 忽略了现代语言设计的进步</span>

在[少既是多](https://commandcenter.blogspot.tw/2012/06/less-is-exponentially-more.html)中，Rob Pike 解释说 Go 是为了在谷歌取代 C 和 C++，它的前身是 [Newsqueak](https://swtch.com/~rsc/thread/newsqueak.pdf) ，这是他在80年代写的一种语言。Go 也有很多关于 [Plan9](https://en.wikipedia.org/wiki/Plan_9_from_Bell_Labs) 的参考，Plan9 是一个分布式操作系统，在贝尔实验室的80年代开发的。

甚至有一个直接从 Plan9 获得灵感的[Go 汇编](https://golang.org/doc/asm)。为什么不使用 [LLVM](https://llvm.org/) 来提供目标范围广泛且开箱即用的体系结构?我此处可能也遗漏了某些东西，但是为什么需要汇编?如果你需要编写汇编以充分利用 CPU ，那么不应该直接使用目标 CPU 汇编语言吗?

Go 的创造者应该得到尊重，但是看起来 Go 的设计发生在平行宇宙（或者他们的 Plan9 lab?）中发生的，这些编译器和编程语言的设计在 90 年代和 2000 年中从未发生过。也可能 Go 是由一个会写编译器的系统程序员设计的。

函数式编程吗？不要提它。泛型？你不需要，看看他们用 C++ 编写的烂摊子!尽管 slice、map 和 channel 都是泛型类型，我们将在下面看到。

Go 的目标是替换 C 和 C++，很明显它的创建者也没有关注其他地方。但他们没有达到目标，因为在谷歌的 C 和 C++ 开发人员没有采用它。我的猜测是主要原因是垃圾回收器。低级别 C 开发人员强烈拒绝托管内存，因为他们无法控制什么时间发生什么情况。他们喜欢这种控制，即使它带来了额外的复杂性，并且打开了内存泄漏和缓冲溢出的大门。有趣的是，Rust 在没有 GC 的情况下采用了完全不同的自动内存管理方法。

Go 反而在操作工具的领域吸引了 Python 和 Ruby 等脚本语言的用户。他们在 Go 中找到了一种方法，可以提高性能，减少 内存/cpu/磁盘 占用。还有更多的静态类型，这对他们来说是全新的。Go 的杀手级应用是 Docker ，它在 devops 世界中引起了广泛的应用。Kubernetes 的崛起加强了这一趋势。

### <span id="interfaces-types">接口是结构类型</span>
Go 接口就像 Java 接口或 Scala 和 Rust 特性（traits）:它们定义了后来由类型实现的行为（我不称之为“类”）。

与 Java 接口和 Scala 和 Rust 特性不同，类型不需要显式地指定接口实现:它只需要实现接口中定义的所有函数。所以 Go 的接口实际上是结构化的。

我们可能认为，这是为了允许其他包中的接口实现，而不是它们适用的类型，比如 Scala 或 Kotlin 中的类扩展，或 Rust 特性，但事实并非如此:所有与类型相关的方法都必须在类型的包中定义。

Go 并不是唯一使用结构化类型的语言，但我发现它有几个缺点:
- 找到实现给定接口的类型很难，因为它依赖于函数定义匹配。我通过搜索实现接口的类，经常发现 Java 或 Scala 中有趣的实现。
- 当向接口添加方法时，只有当它们用作此接口类型的值时，才会发现哪些类型需要更新。 相当一段时间这可能被忽视。 Go 建议使用非常少的方法构建小型的接口，这是防止这种情况的一种方式。
- 类型可能在不知不觉中实现了一个接口，因为它作为相应的方法。但是偶然的，实现的语义可能与接口契约所期望的不同。

*更新* : 对于接口的一些丑陋问题，请参阅下面的 [无接口值（nil interface values）](#nil-interface-values)。

### <span id="no-enum">没有枚举</span>

Go 没有枚举，在我看来，这是一个错失的机会。

`iota` 可以快速生成自动递增的值，但它看起来更像一个技巧 而不是一个特性。实际上，由于在一系列的 iota 生成的常量中插入一行会改变下列值的值，这是很危险的。由于生成的值是在整个代码中使用的值，因此这会导致有趣的（而不是!）意外。

这也意味着没有办法让编译器彻底检查 `switch` 语句，也无法描述类型中允许的值。

### <span id="enum">`:=` / `var` 两难选择</span>

Go 提供两种方法来声明一个变量，并为其赋值: `var x = "foo"` 和 `x:= "foo"`。这是为什么呢?

主要的区别是 var 允许未初始化的声明(然后您必须声明类型)，比如在 `var x string` 中，而 `:=` 需要赋值，并且允许混合使用现有变量和新变量。我的猜测是`:=`被发明来使错误处理减少一点麻烦:

使用 `var`

```go
var x, err1 = SomeFunction()
if (err1 != nil) {
  return nil
}

var y, err2 = SomeOtherFunction()
if (err2 != nil) {
  return nil
}
```

使用 `:=`:

```go
x, err := SomeFunction()
if (err != nil) {
  return nil
}

y, err := SomeOtherFunction()
if (err != nil) {
  return nil
}
```

`:=` 的语法也很容易意外的影响一个变量。我已经不止一次吃这个亏了，因为 `:=` (声明和赋值)太接近了 `=` (赋值)，如下图所示:

```go
foo := "bar"
if someCondition {
  foo := "baz"
  doSomething(foo)
}
// foo == "bar" 即使 "someCondition" 为真
```

### <span id="zero-panic">零值 panic</span>

Go 没有构造函数。正因为如此，它坚持认为“零值”应该是易于使用的。这是一个有趣的方法，但在我看来，它所带来的简化主要是针对语言实现者的。

在实践中，许多类型在没有正确初始化的情况下不能做有用的事情。让我们看一下 `io.File` 对象，从 [实效 Go 编程](http://docscn.studygolang.com/doc/effective_go.html#%E5%A4%8D%E5%90%88%E5%AD%97%E9%9D%A2) 取出的一个例子:

```go
type File struct {
	*file // os specific
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Read(b []byte) (n int, err error) {
	if err := f.checkValid("read"); err != nil {
		return 0, err
	}
	n, e := f.read(b)
	return n, f.wrapErr("read", e)
}

func (f *File) checkValid(op string) error {
	if f == nil {
		return ErrInvalid
	}
	return nil
}
```

我们能发现什么？

- 在一个零值 `File` 调动 `Name()` 将会导致 panic ，因为它的 `file` 字段是 nil
- `Read` 函数和 [几乎所有其他](https://github.com/golang/go/blob/release-branch.go1.10/src/os/file.go) `File` 方法，首先检查文件是否被初始化。

因此，零值 `File` 不仅无用，还会导致 `panics`。你*必须*使用一个构造函数，比如 `Open` 或 `Create`。检查正确的初始化是每个函数调用中需要花费的开销。

在标准库中有无数这样的类型，有些甚至不尝试用它们的零值来做一些有用的事情。在零值 `html.Template` 上调用任何方法都会导致 `panic`。

还有一个很严重的问题: `map` 的零值:你可以查询它，但是在它里面存储东西会导致 `panic`:

```go
var m1 = map[string]string{} // 空值 map
var m0 map[string]string     // 零值 map (nil)

println(len(m1))   // 输出 '0'
println(len(m0))   // 输出 '0'
println(m1["foo"]) // 输出 ''
println(m0["foo"]) // 输出 ''
m1["foo"] = "bar"  // 没问题
m0["foo"] = "bar"  // panics!
```

这需要在结构具有 `map` 字段时小心，因为在添加条目之前必须对其进行初始化。

因此，作为一个开发人员，您必须经常检查您想要使用的结构是否需要调用构造函数，或者零值是否可用。这是语言简化对编程带来的沉重负担。

### <span id="exceptions">Go 没有异常。哦,等一下……它有!</span>

这篇博客文章「[为什么 Go 获得异常的方式是对的](https://dave.cheney.net/2012/01/18/why-go-gets-exceptions-right)」详细解释了为什么异常是糟糕的，为什么Go方法要求返回 `错误` 是更好的。我可以同意这一点，确实在使用异步编程或像 Java 流这样的函数风格时，异常是很难处理的（前者可以放到一边，由于 goroutines 的原因它在 Go
 中是没有必要的，而后者几乎是不可能）。这篇博文提到 `panic` 「总是对你的程序抛出致命错误，游戏结束」，这很不错。

在此之前，「[Defer, panic and recover](https://blog.golang.org/defer-panic-and-recover)」 解释了如何从 `panic` 中恢复（实际上是通过捕获它们），并提到在 go 标准库中 json 包可以看到 panic 和 recover 的真实使用。

事实上, json 解码器有一个 [共同的错误处理函数](https://github.com/golang/go/blob/release-branch.go1.10/src/encoding/json/decode.go#L299) 去 panics,panic 在顶层 `unmarshal`函数中恢复（recover）,[检查panic类型](https://github.com/golang/go/blob/release-branch.go1.10/src/encoding/json/decode.go#L173) 并返回一个错误如果它是一个“本地 panic ”或其它错误再次触发的 panic（ 失去最初的 panic 的追溯）。

对于任何 Java 开发人员来说，这明显看上去是一个`try` / `catch (DecodingException ex)`。所以 Go 确实有异常处理，它在内部使用了却告诉你不要用。

有趣的事实:几个星期前，一个非谷歌的人[修复了 json 解码器](https://github.com/golang/go/commit/74a92b8e8d0eae6bf9918ef16794b0363886713d)，以使用常规的错误冒泡处理。

## <span id="ugly">令人厌恶的点</span>

### <span id="dependency-nightmare">依赖管理噩梦</span>

首先引用一个在谷歌著名的 Go 语言使用者 Jaana Dogan (aka JBD) 的话，最近在推特上发泄她的不满:

> 如果依赖管理再过一年还没有解决，我将会考虑退出 Go 并且永远不会回来。 依赖性管理问题经常颠覆我从语言中获得的所有乐趣。
>
> — JBD (@rakyll) [March 21, 2018](https://twitter.com/rakyll/status/976563654731579392?ref_src=twsrc%5Etfw)

简单点说，Go 中没有依赖管理。当前所有的解决方案都只是一些技巧和变通方法。

这要追溯到它的起源 -- 谷歌，以使用了一个 [巨大的单片存储库](https://research.google.com/pubs/pub45424.html) 管理所有源代码而闻名。不需要模块的版本控制，也不需要第三方模块的仓库，你可以从当前分支构建任何项目。不幸的是，这在开放的互联网上是行不通。

在 Go 中添加依赖意味着将依赖的源代码仓库克隆到你的 GOPATH 下。版本是什么?克隆当前的主分支就行了，管它写的是什么。但是如果不同的项目需要不同的版本依赖呢?他们做不到。因为「版本」的概念根本不存在。

另外，您自己的项目必须在 `GOPATH` 下，否则编译器无法找到它。想让你的项目在单独的目录里清晰地组织起来?那你必须配置每一个项目的 `GOPATH` ，或者使用符号链接。

社区已经开发了 [大量的工具](https://github.com/golang/go/wiki/PackageManagementTools) 解决此问题。包管理工具引入了 `vendoring` 和锁文件来保存您克隆的任何仓库的Git sha1，以提供可复现的构建。

最后，在Go 1.6中，`vendor` 目录得到了[官方支持](https://golang.org/cmd/go/#hdr-Vendor_Directories)。但它是关于你克隆的 `vendoring`，仍然不是正确的版本管理。没有对从传递依赖中导入发生冲突的解决方案，这通常是通过 [语义化版本](https://semver.org/) 来解决的。

不过，情况正在好转:`dep`，[官方的依赖管理工具](https://golang.github.io/dep/) 最近被引入以支持文件控制（vendoring）。它支持版本（git tags），并有一个遵循语义版本控制约定的版本解决程序。它还不稳定，但方向是正确的。然而，它仍然需要你的项目存放在 `GOPATH` 里。

但是 `dep` 可能不会像 [`vgo`](https://github.com/golang/vgo)长久，`vgo` 是谷歌发起的，想要从语言本身带来版本控制，并且最近已经引起了一些波动。

所以 Go 的依赖管理是噩梦。设置起来很痛苦，当你在开发的时候你不会考虑到它，直到你添加一个新的导入（import）或者只是简单地想拉取你团队成员的一个分支到你的 `GOPATH` 里的时候，程序崩溃了...

现在让我们再次回到代码的问题上。

### <span id="mutability-hardcode">易变性是用语言硬编码的。</span>

在 Go 中没有定义不可变结构的方法: struct 字段是可变的，`const` 关键字不适用于它们。Go 通过简单的赋值就可以轻松地复制整个`struct`，因此，我们可能认为，通过值传递参数来保证不变性，只需要复制的代价。

然而，不出所料的，它不复制指针引用的值。而且由于内置的集合（map、slice 和 array）是引用和可变的，复制包含其中任意一项的 `struct` 只是复制了指向底层内层的指针。

下面的例子说明了这个问题：

```go
type S struct {
	A string
	B []string
}

func main() {
	x := S{"x-A", []string{"x-B"}}
	y := x // 复制 struct
	y.A = "y-A"
	y.B[0] = "y-B"

	fmt.Println(x, y)
	// 输出 "{x-A [y-B]} {y-A [y-B]}" -- x 被修改!
}
```

所以你必须非常小心，如果你通过值传递参数，不要认定它就是不变的。

有一些 [深度复制库](https://github.com/jinzhu/copier) 试图使用(慢)反射来解决这个问题，但是它们有不足之处，因为私有字段不能通过反射访问。因此，为了避免竞争条件而进行防御性复制将会很困难，需要大量的重复代码。Go甚至没有一个可以标准化这个的克隆接口。

### <span id="slice-gotchas">切片（slice）陷阱</span>

切片带来了很多问题。正如「[Go slice: usage and internals](https://blog.golang.org/go-slices-usage-and-internals)」中所解释的那样，考虑到性能原因，再次切片一个切片不会复制底层的数组。这是一个值得赞赏的目标，但也意味着切片的子切片只是遵循原始切片变化的视图。因此，如果您想要将它与初始的切片分开请不要忘记 `copy()`。

对于 `append` 函数，忘记 `copy()` 会变得更加危险:如果*它没有足够的容量*来保存新值，底层数组将会重新分配内存和大小。这意味着 `append` 的结果能不能指向原始数组取决于它的初始容量。这会导致难以发现的不确定 bugs。

在下面的代码中，我们看到为子切片追加值的影响取决于原始切片的容量:

```go
func doStuff(value []string) {
	fmt.Printf("value=%v\n", value)

	value2 := value[:]
	value2 = append(value2, "b")
	fmt.Printf("value=%v, value2=%v\n", value, value2)

	value2[0] = "z"
	fmt.Printf("value=%v, value2=%v\n", value, value2)
}

func main() {
	slice1 := []string{"a"} // 长度 1, 容量 1

	doStuff(slice1)
	// Output:
	// value=[a] -- ok
	// value=[a], value2=[a b] -- ok: value 未改变, value2 被更新
	// value=[a], value2=[z b] -- ok: value 未改变, value2 被更新

	slice10 := make([]string, 1, 10) // 长度 1, 容量 10
	slice10[0] = "a"

	doStuff(slice10)
	// Output:
	// value=[a] -- ok
	// value=[a], value2=[a b] -- ok: value 未改变, value2 被更新
	// value=[z], value2=[z b] -- WTF?!? value 改变了???
}
```

### <span id="race-conditions">易变性和 channels: 竞争条件更容易发生。</span>

Go 并发性是 [通过 channels 建立在CSP](https://golang.org/doc/faq#csp) 上的，它使用 channel 使得协调 goroutines 比在共享数据上同步更简单和安全。老话说的是「[不要通过共享内存来通信;而应该通过通信来共享内存](https://blog.golang.org/share-memory-by-communicating)」。这是一厢情愿的想法，在实践中是不能安全实现的。

正如我们在上面看到的那样，Go 没办法获得不可变的数据结构。这意味着一旦我们在 channel 上发送一个指针，游戏就结束了:我们在并发进程之间共享了可变的数据。当然，一个 channel 的结构是赋值 channel 传送的值(而不是指针)，但是正如我们在上面看到的，这些没有深度复制引用，包括 slices 和 maps 本质上都是可变的。与接口类型的 struct 字段相同:它们是指针，接口定义的任何可变方法都是对竞争条件的开放。

因此，尽管 channels 表面上使并发编程变得容易，但它们并不能阻止共享数据上的竞争条件。而 slices 和 maps 本身的可变性使这种情况更有可能发生。

谈到竞争条件时，Go 包含一个 [竞争条件检测模式](https://blog.golang.org/race-detector)，该模式检测代码以找到不同步的共享访问。它只能在事件发生的时候检测到竞争问题，所以大多数情况下是在集成或负载测试期间，希望这些能够运行比赛条件。由于它的高运行时成本(除了临时的调试会话)，它不能实际应用于生产环境。

### <span id="error-management">嘈杂的错误管理</span>

你可以很快学会 Go 的错误处理模式，重复到令人作呕:

```go
someData, err := SomeFunction()
if err != nil {
	return err;
}
```

因为 Go 声称不支持异常（[虽然它已经支持](#exceptions)），每个能够以错误结尾的函数都必须把错误作为其最后一个结果。这特别适用于执行某些 I/O 的每个函数，因此这种啰嗦的模式在网络应用程序中非常普遍，这是 Go 的主要领域。

您很快就会忽视这种模式，并将其识别为「好，错误处理了」，但是仍然很杂乱，有时很难在错误处理中找到实际的代码。

这里有几个问题，因为一个错误的结果可能有名无实，例如当从无所不在的 io.Reader读取时：

```go
len, err := reader.Read(bytes)
if err != nil {
	if err == io.EOF {
		// 一切正常，文件结尾
	} else {
		return err
	}
}
```

在“Error has values”中，Rob Pike 提出了一些减少错误处理冗余的策略。我发现它们实际上是危险的创可贴:

```go
type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) write(buf []byte) {
	if ew.err != nil {
		return // 当已经出错时，什么都不写入
	}
	_, ew.err = ew.w.Write(buf)
}

func doIt(fd io.Writer) {
	ew := &errWriter{w: fd}
	ew.write(p0[a:b])
	ew.write(p1[c:d])
	ew.write(p2[e:f])
	// 等等
	if ew.err != nil {
		return ew.err
	}
}
```

基本来说，一直检查错误是很痛苦的，所以这里提供了直到结束之前都会忽略错误的方法。任何写入操作一旦出错它还是会执行，即使我们知道不该再执行了。如果这样做资源消耗更高呢?我们刚刚浪费了资源，因为 Go 的错误处理是一种痛苦。

Rust 有类似的问题:没有异常(真的没有，跟 Go
 相反)，方法失败返回 `Result<T, Error>`，并且需要对结果模式匹配。所以，Rust1.0 添加了 `try!` 宏并且认识到这一模式的普遍性，使它成为 [一流的语言特征](https://doc.rust-lang.org/rust-by-example/error/result/enter_question_mark.html)。因此，在保持正确的错误处理的同时，也有上述代码的简洁性。

由于 Go 没有泛型和宏，所以很不幸地，更换为 Rust 的方法是不可能的。

### <span id="nil-interface-values">Nil 接口值</span>

这是在看到 [redditor jmickeyd](https://www.reddit.com/r/programming/comments/8bj4yc/go_the_good_the_bad_and_the_ugly/dx82cgz/) 展示了 nil 和接口的怪异表现后的更新，这绝对称得上是丑陋的。我稍微扩展了一下:

```go
type Explodes interface {
	Bang()
	Boom()
}

// Type Bomb implements Explodes
type Bomb struct {}
func (*Bomb) Bang() {}
func (Bomb) Boom() {}

func main() {
	var bomb *Bomb = nil
	var explodes Explodes = bomb
	println(bomb, explodes) // '0x0 (0x10a7060,0x0)'
	if explodes != nil {
		println("Not nil!") // 'Not nil!' 我们为什么会走到这里?!?!
		explodes.Bang()     // 运行正常
		explodes.Boom()     // panic: value method main.Bomb.Boom called using nil *Bomb pointer
	} else {
		println("nil!")     // 为什么没有在这里结束？
	}
}
```
上面的代码验证了 `explodes` 不是nil，但是代码在 `Boom` 中 panics，在 `Bang` 中没有。这是为什么呢?解释在 `println` 这一行:`bomb` 指针是 `0x0`，它实际上是 `nil`，但是 `explodes` 是非nil `(0x10a7060,0x0)`。

这两个元素的第一个元素是通过 `Explodes` 类型来实现 `Bomb` 接口的方法分派表的指针，第二个元素是实际 `Explodes` 对象的地址，它是 `nil`。

对 `Bang` 的调用成功是因为它需要传递的是 `Bomb` 指针:没有必要取消指针来调用方法。`Boom` 方法作用于一个*值*，因此调用会导致指针取消引用，这会造成 panic。

注意，如果我们写了 `var explodes Explodes = nil`，然后 `!= nil` 本不该通过。

那么，我们应该如何安全地编写测试呢?我们必须检查接口值,如果是非 nil，检查接口对象指向的值…使用反射!

```go
if explodes != nil && !reflect.ValueOf(explodes).IsNil() {
	println("Not nil!") // we no more end up here
	explodes.Bang()
	explodes.Boom()
} else {
	println("nil!")     // 'nil' -- all good!
}
```

这是漏洞还是特性? `Go语言之旅`  有一个 [专门的页面](https://tour.golang.org/methods/12) 来解释这种行为，并清楚地表示 *「注意，一个具有nil值的接口值本身就是非空值」*。

尽管如此，这仍然是丑陋的，并且会导致非常细微的错误。在我看来，这是语言设计中的一个很大的缺陷，只是为了使它的实现更加容易。

### <span id="dsl">Struct 字段标记:字符串中的运行时DSL。</span>

如果您在 Go 中使用了 JSON，您肯定遇到过类似的情况:

```go
type User struct {
	Id string    `json:"id"`
	Email string `json:"email"`
	Name string  `json:"name,omitempty"`
}
```

这些是 [结构标记(struct tags)](https://golang.org/ref/spec#Struct_types)，语言规范说这是一个字符串「通过反射接口可见，并参与结构的类型标识，但是却被忽略了」。所以，基本上，把你想要的东西放到这个字符串中，并在运行时使用反射来解析它。如果语法不对，运行时就会出现 panic。

这个字符串实际上是字段元数据，在许多语言中已经存在了几十年，称为「[注释](https://en.wikipedia.org/wiki/Java_annotation)」或「属性」。通过语言支持，它们的语法在编译时被正式定义和检查，同时仍然是可扩展的。

为什么要决定使用一个原始字符串，任何库都可以决定使用它想要的任何 DSL ，在运行时解析？

当您使用多个库时，情况会变得很糟糕:这里有一个从协议缓冲区的 [Go文档](https://godoc.org/github.com/golang/protobuf/proto) 中取出的示例:

```go
type Test struct {
	Label         *string             `protobuf:"bytes,1,req,name=label" json:"label,omitempty"`
	Type          *int32              `protobuf:"varint,2,opt,name=type,def=77" json:"type,omitempty"`
	Reps          []int64             `protobuf:"varint,3,rep,name=reps" json:"reps,omitempty"`
	Optionalgroup *Test_OptionalGroup `protobuf:"group,4,opt,name=OptionalGroup" json:"optionalgroup,omitempty"`
}
```
附注:为什么这些标签在使用 JSON 时如此常见？因为在 Go 公共字段中，必须使用大写字母，或者至少以大写字母开头，而在 JSON 中命名字段的常见约定是小写的 camelcase 或 snake_case。因此需要进行冗长的标记。

标准的 JSON 编码器 / 解码器不允许提供自动转换的命名策略，就像 [Jackson在Java中所做的](https://github.com/FasterXML/jackson-databind/blob/master/src/main/java/com/fasterxml/jackson/databind/PropertyNamingStrategy.java)。这可能解释了为什么 Docker APIs 中的所有字段都是大写的:这避免了它的开发人员为他们的大型 API 编写这些笨拙的标签。

### <span id="generic">没有泛型…至少不是为了你。</span>

很难想象一种没有泛型的现代静态类型化语言，但这就是你在 Go 中看到的:它没有泛型...或者更精确地说，几乎没有泛型，我们会看到它比没有泛型更糟糕。

内置的 slice、map、array和 channel 都是泛型。声明一个 `map[string]MyStruct` 清楚地显示了具有两个参数的泛型类型的使用。这很好，因为它允许类型安全编程捕获各种错误。

然而，没有用户可定义的泛型数据结构。这意味着您不能定义可重用的抽象，它可以以类型安全的方式使用任何类型。您必须使用非类型 `interface{}`，并将值转换为适当的类型。任何错误只会在运行时被抓住，会导致 panic。对于 Java 开发人员来说，这就像回到 [回退 Java 5 个版本到 2004 年](https://en.wikipedia.org/wiki/Java_version_history#J2SE_5.0)。

在「[少即是多](https://commandcenter.blogspot.fr/2012/06/less-is-exponentially-more.html)」中，Rob Pike 意外地将泛型和继承放在同一个「类型编程」包中，并说他喜欢组合而不是继承。不喜欢继承很好（实际上我写了很多没有继承的Scala），但是泛型回答了另一个问题：可重用性，同时保护类型安全。

正如下面我们将看到的，在用泛型做内部构建和用户无法定义泛型之间的区别会对开发人员「舒适」和编译时类型安全产生更多的影响:它会影响整个 Go 生态系统。

### <span id="no-structures">Go 在 slice 和 map 之外几乎没有什么数据结构。</span>

Go 生态系统没有很多数据结构，它们可以从内置的 slices 和 maps 中提供额外或不同的功能。Go 最新版本添加了提供其中几个容器包。它们都有相同的警告:它们处理 `interface{}` 值，这意味着您将失去所有类型的安全性。

让我们来看一个 [snc.Map](`https://medium.com/@deckarep/the-new-kid-in-town-gos-sync-map-de24a6bf7c2c`) 的例子，它是一个具有较低线程争用的并发映射，而不是使用互斥锁来保护常规映射：:

```go
type MetricValue struct {
	Value float64
	Time time.Time
}

func main() {
	metric := MetricValue{
		Value: 1.0,
		Time: time.Now(),
	}

	// Store a value

	m0 := map[string]MetricValue{}
	m0["foo"] = metric

	m1 := sync.Map{}
	m1.Store("foo", metric) // not type-checked

	// Load a value and print its square

	foo0 := m0["foo"].Value // rely on zero-value hack if not present
	fmt.Printf("Foo square = %f\n", math.Pow(foo0, 2))

	foo1 := 0.0
	if x, ok := m1.Load("foo"); ok { // have to make sure it's present (not bad, actually)
		foo1 = x.(MetricValue).Value // cast interface{} value
	}
	fmt.Printf("Foo square = %f\n", math.Pow(foo1, 2))

	// Sum all elements

	sum0 := 0.0
	for _, v := range m0 { // built-in range iteration on map
		sum0 += v.Value
	}
	fmt.Printf("Sum = %f\n", sum0)

	sum1 := 0.0
	m1.Range(func(key, value interface{}) bool { // no 'range' for you! Provide a function
		sum1 += value.(MetricValue).Value        // with untyped interface{} parameters
		return true // continue iteration
	})
	fmt.Printf("Sum = %f\n", sum1)
}
```

这就是为什么在 Go 生态系统中没有很多数据结构的一个很好的例子:与内置的切片和映射相比，它们是一种痛苦。原因很简单:数据结构分为两类:

- 上层，内置的 slice，map，array 和 channel:类型安全和泛型，方便使用 `range`，
- Go 代码编写的其他地方：不能提供类型安全，因为需要强制转换而难以使用。

因此，库定义的数据结构真的需要为我们的开发人员提供切实的利益，才愿意为松散类型的安全性和额外的代码冗长付出代价。

当我们想要编写可重用的算法时，内置结构和 Go 代码之间的二元性在细节方面是痛苦的。这是标准库的 [排序扩展包](https://golang.org/pkg/sort/) 中的一个例子:

```go
import "sort"

type Person struct {
	Name string
	Age  int
}

// ByAge implements sort.Interface for []Person based on the Age field.
type ByAge []Person

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }

func SortPeople(people []Person) {
	sort.Sort(ByAge(people))
}
```

等等... 这是认真的吗?我们必须定义一个新的类型 `ByAge`，它必须实现 3 种方法来实现通用的（严格讲是「可重用」）排序算法和类型切片。

对我们开发人员来说唯一重要的是，用较少的函数比较两个对象，并且是域依赖的。其他的东西都是噪音和重复，简单的事实是，Go 没有泛型。我们需要对*每一种我们想排序的类型*重复它，每个比较器也是。

*更新: [Michael Stapelberg](https://twitter.com/zekjur/status/983965201774071808) 指出我忘记了 [sort.Slice](https://golang.org/pkg/sort/#Slice)。看起来好多了，尽管它在 hood (eek!)下使用反射，并要求比较器函数在切片上做一个闭包去排序，这仍然很难看。*

每个解释 Go 不需要泛型的文章，把这个当做「Go 的方式」，它允许使用可重用的算法，同时避免向下转换到 `interface{}`...

好了。为了缓解疼痛，如果 Go 有可以生成这个无意义的样板的宏就好了，对吗?

### <span id="generate">go generate，还说得过去，但是...</span>
Go 1.4 介绍了 [`go generate` 命令](https://blog.golang.org/generate)，从源代码的注释中触发代码的生成。嗯，这里的「注释」实际上指的是 `//go:generate` 的注释，需要符合严格的规则：「注释必须从行开始处开始，在 `//` 和 `go:generate` 之间没有空格」。如果写错了，加了一个空格，没有工具会警告你出错了。

这实际上涵盖了两种应用场景:

- 从其他来源生成 Go 代码：ProtoBuf / Thrift / Swagger 模式，语言语法，等等。
- 生成的 Go 代码补充了现有的代码，例如作为示例的 [stringer](https://godoc.org/golang.org/x/tools/cmd/stringer) ，它为一系列类型常量生成 `String()` 方法。

第一个用例是可以的，附加价值是你不需要摆弄 `makefiles`，而生成的说明可以接近生成的代码的用法。

对于第二个用例，许多语言，比如 Scala 和 Rust，有宏（在 [设计文档](https://docs.google.com/document/d/1V03LUfjSADDooDMhe-_K59EgpTEm3V8uvQRuNMAEnjg/edit) 中提到的）在编译过程中都可以访问源代码的 AST。Stringer 实际上 [导入了Go编译器的解析器](https://github.com/golang/tools/blob/master/cmd/stringer/stringer.go#L69) 来遍历 AST。Java 没有宏，但注释处理器扮演同样的角色。

许多语言也不支持宏，所以在这里没有什么根本的错误，除了这个脆弱的由逗号驱动的语法，它看起来像一个快速的技巧，以某种方式完成工作，而不是作为清晰的语言设计被慎重考虑。

哦，你知道 Go 编译器实际上有 [很多注释/程序](https://dave.cheney.net/2018/01/08/gos-hidden-pragmas) 和 [条件编译](https://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool) 使用这个脆弱的注释语法吗?

## <span id="conclusion">结论</span>

就像你猜到的，我对 Go 爱恨交加。Go 有点像这样的朋友，你喜欢和他一起出去玩，因为他很有趣和他喝啤酒聊天很棒，但是当你想进行更深入的交流时，你会觉得无聊和痛苦，然后你不想和他一起去度假。

我喜欢 Go 在写高效的 APIs 或网络方面时的简单，goroutines 使这些 很容易解释。当我必须实现业务逻辑时，我讨厌它有限的表现力和所有的等着打击你的语言怪癖和陷阱。

直到最近，在 Go 占据的领域中并没有出现真正的替代选择，它高效地开发本地可执行文件，而不会导致 C 或 C++ 的痛苦。[Rust](https://www.rust-lang.org/zh-CN/) 在飞速进步，我用得越多，越能发现它的有趣之处和优秀设计。我有一种感觉，Rust 是那些需要时间相处的朋友，你最终会想要和他们建立长期的关系。

回归技术层面，你会发现一些文章说 Rust 和 Go 不是一个领域的，Rust 是一种系统语言，因为它没有内存回收机制 等等。我认为这越来越不真实了。在 [伟大的web框架](http://www.arewewebyet.org/) 和优秀的 [ORM](http://diesel.rs/)s 中 Rust 正在爬得更高。它也给你一种温暖的感觉，“如果它编译，错误将来自我写的逻辑，而不是我忘记注意的语言怪癖”。

我们还在容器/服务网格区域看到一些有趣的行动， Buoyant（[Linkerd](https://linkerd.io/)的开发商）正在开发它们的新 Kubernetes 服务网格 [Conduit](https://buoyant.io/2017/12/05/introducing-conduit/) 作为一个组合，来自控制层面（我猜可能是因为可用的 [Kubernetes库](https://bluxte.net/musings/2018/04/10/go-good-bad-ugly/#the-var-dilemma)）的 Go 和数据层面拥有良好效率和鲁棒性的 Rust ，以及 [Sozu代理](https://www.sozu.io/)。

Swift 也是这个家庭的一份子，或者是 C 和 C++ 的最新替代品。它的生态系统仍然过于以苹果为中心，即使它现在可以在 Linux 上使用，并且已经有了新的 [服务器端 APIs](https://swift.org/server-apis/) 和 [Netty 框架](https://github.com/apple/swift-nio)。

这里当然没有万能药和通用之法。但是知道你所用工具的问题至关重要。我希望这篇博文教会了你关于 Go 你以前没有意识到的问题，这样你就可以避开陷阱!

## <span id="hacker-news">几天后: Hacker News 第三名!</span>

*更新，发布3天后*:这篇文章反响惊人。它已经成为了 [Hacker News](https://news.ycombinator.com/item?id=16830153) 的头版(我看到的最好排名是#3)和[/r/programming](https://www.reddit.com/r/programming/comments/8bj4yc/go_the_good_the_bad_and_the_ugly/)(我看到的最好排名是#5)，并且在 [Twitter 上得到了一些关注](https://twitter.com/search?l=&q=https%3A%2F%2Fbluxte.net%2Fmusings%2F2018%2F04%2F10%2Fgo-good-bad-ugly%2F)。

这些评论通常都是正面的(甚至是在[/r/golang/](https://www.reddit.com/r/golang/comments/8bj4tx/go_the_good_the_bad_and_the_ugly/))，或者至少承认这篇文章是公平的，并且力求公正。[/r/rust](https://www.reddit.com/r/rust/comments/8bjio2/xpost_from_rprogramming_go_the_good_the_bad_and/)的人们当然喜欢我对 Rust 的兴趣。我从未听说过的人甚至给我发邮件说:“*我只是想让你知道，我认为你写的文章是最好的。感谢您为此付出的所有努力*”。

这是写作时最困难的部分:尽量做到客观公正。这当然不是完全可能的，因为每个人都有自己的偏好，为什么我关注意外的惊喜和语言工程学:语言对你有多大帮助，而不是妨碍你，或者至少是*我的方式*。

我还在标准库或[ golang.org](https://golang.org/) 上搜索了代码样本，并引用了Go团队的人员，以我对权威材料的分析为基础，避免了“*meh，你引用了一个错误的人*”的反应。

写这篇文章用了我两个星期的晚上时间，但是这真的很有趣。当你做严肃而诚实的工作时，你会得到这样的结果:来自技术网络的许多好的共鸣(如果你忽略掉少数捣乱的和一直脾气暴躁的人)。极大调用了我写更多的深度内容的积极性!

---

via: https://bluxte.net/musings/2018/04/10/go-good-bad-ugly/

作者：[Sylvain Wallez](https://bluxte.net/about/)
译者：[Donng](https://github.com/Donng)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出


