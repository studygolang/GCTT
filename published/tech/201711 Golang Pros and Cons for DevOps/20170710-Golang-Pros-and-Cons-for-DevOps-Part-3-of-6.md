已发布：https://studygolang.com/articles/12614

# Golang 之于 DevOps 开发者的利与弊(六部曲之三)：速度 vs. 缺少泛型

这是我们关于 DevOps 开发流程之中使用 Golang 之利与弊的六部曲系列。在这篇文章里，我们会讨论 Golang 的运行时/编译/维护的速度（好的方面）；以及缺少泛型（缺点）。

在阅读这篇文章之前，请确保你已经阅读了[上一篇](https://studygolang.com/articles/12608)关于“接口实现以及公有/私有命名方式”，或者[订阅](http://eepurl.com/cOHJ3f)我们的博客更新提醒来获取此六部曲后续文章的音讯。（我们会隔周更新，但是鉴于我们正在忙着发布[我们的 beta 平台](https://blog.bluematador.com/blog/posts/announcing-beta-launch-blue-matador-devops-monitoring-platform/)我们的进度确实有点延后。)

- [Golang 之于 DevOps 开发的利与弊（六部曲之一）：Goroutines, Channels, Panics, 和 Errors](https://studygolang.com/articles/11983)
- [Golang 之于 DevOps 开发的利与弊（六部曲之二）：接口实现的自动化和公有/私有实现](https://studygolang.com/articles/12608)
- [Golang 之于 DevOps 开发的利与弊（六部曲之三）：速度 vs. 缺少泛型](https://studygolang.com/articles/12614)
- [Golang 之于 DevOps 开发的利与弊（六部曲之四）：time 包和方法重载](https://studygolang.com/articles/12615)
- [Golang 之于 DevOps 开发的利与弊（六部曲之五）：跨平台编译，Windows，Signals，Docs 以及编译器](https://studygolang.com/articles/12616)
- Golang 之于 DevOps 开发的利与弊（六部曲之六）：Defer 指令和包依赖性的版本控制

## Golang 之利: 速度

我基于以往的经验：在我们在写 Blue Matador agent 之时在乎什么，把速度这个优势分解成三个不同的类别。（1）运行时，（2）编译过程，（3）维护流程。确切数据的量会随着分类的递进而减少 - 你会明白原因的.

### 运行时速度

因为我们的 agent 是跑在客户服务器之上，运行时速度是处于第一优先级的，除此之外还有安全和自动更新。很不幸的是，我们直到被我们的 Python 版本的 agent 狠狠坑了一次之后，才明白什么方面才是第一优先级的。现在回顾，我很高兴我们远离了 Python 。

作为一个抽象化的运行时速度对比，我们从 benchmarksgame.alioth.debian.org 扒了一些数据然后放到了这个漂亮的图表里。这表显示的是每种语言完成标准测试所花费的时间。花的时间越长，柱形会越高。注意 Python 在除了圆周率小数点计算（因为它给不出准确答案）之外的所有测试都远远落后。剩余的语言全都处于一个档次。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go_devops/golang-language-runtimes-1.png)

鉴于这是一篇讨论 Golang 速度的优点的博文，而不是探讨我们为什么从 Python 迁移到 Golang ，我会去除两种持续慢于 Golang 的语言: Python3 和 Node.js。在把干扰项去掉之后，你会看见 Golang 和其他顶级竞争者的结果：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go_devops/golang-language-runtimes-2.png)

图中显而易见的是，C++ 是最快的语言 - 完全不出意料。实际上，即使 Java 都把 Golang 击败了，但这出于两个很好的理由:（1）Java 虚拟机（JVM）从1995年就开始开发了，比 Golang 多了 17 年。以及（2） 比起Golang ，JVM 在测试中花了 2 到 30 倍的内存使用量 - 这意味着总体上 Golang 的垃圾回收员比JVM的工作得更勤劳。

### 编译速度

我们之前的 agent 是用 Python 写的（并不是一个编译语言），但这并不意味着我们不熟悉 Go 的编译时优势。

从 Golang 的一开始，较短的编译时间一直是一个严苛的要求。 Go 是 Google 的 Ken Thompson 和 Rob Pike 创造的。Google，有超过20亿行代码，毫无疑问对于编译所会浪费的时间极其严肃。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go_devops/compiling.png)

在这个[帖子](http://imgur.com/a/jQUav#xVgi2ZA)里，有一些非常棒的信息，我也会在这总结。（我强力推荐你阅读这个帖子，因为它既精炼又细致。）

首先看这幅关于一个相对大的代码库的编译时间的图。代码细节可以在之前提到的链接中找到。注意看，Golang 在一众竞争者中的表现多棒，除了比不过 Pascal 这个被设计成只为了跑一遍代码的语言 - 基本上编译时间会确保是 O(n)。考虑到 Golang 的语言规范依旧可用（不像 Pascal ），我会把 Go 微小的劣势视为 Go 的胜利。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go_devops/golang-language-compilation-time.png)

我不理解为什么这次测试所有语言的编译速度都如此快，但我明白至少一部分的原因是因为依赖管理系统。当一个文件被编译时，编译器只看这个文件列出的直接依赖。它并不需要去递归加载所有依赖文件的依赖。

Blue Matador Agent 有 29 个包，116824 行代码。它还有 3 个目标操作系统，2 个目标架构，以及 3 个模块。所有的这些都能在 10 秒内完成并行编译(在一个 8 核的开发笔记本上)。好消息是我们几乎不用花时间在编译上。但坏消息是我们就没时间像[这幅 XKCD 漫画](https://xkcd.com/303/)一样击剑了。

### 维护速度

现在，转到浑水区。在我开始细说之前，让我先澄清几点:

1. 我们的 Golang 代码库是刻意保持在一个较小的体量
2. 可能因为体量小，我们只有 2 个被报告的 bug
3. 目前我们有 3 个全职开发者，其中只有两个碰过 agent 的代码

因此，当我说维护 Golang 代码是非常简单并且不费时间的时候，请理解我们是多么缺少依据和经验。关于这一点我可能完完全全是错的。如果以后程序员们在诅咒着我的名字，我确信我会找到这个问题的答案。

尽管如此，以下是我认为 Go 的代码维护效率更好的原因：

- 没有内存管理。在 C/C++ 有很多代码是只为了内存管理而存在。你为了管理内存不得不做了一堆古怪的事。这在 Go 里面，得益于垃圾回收机制，完全不是问题。
- 稳定可靠的核心库。除了[错误返回系统](https://blog.bluematador.com/blog/posts/golang-pros-cons-for-devops-part-1-goroutines-panics-errors/)之外，我们的代码十分精简。这是因为Go的核心库有我们需要的所有东西; 从 HTTP 请求和 JSON 编码/解码到进程 fork 和 IPC 管道。
- 没有泛型。对，我即将要说缺少泛型是这个语言的一个严重缺点。但是如果没有泛型的话，变量类型全都会是明确且已知的。当你读一个类文件的时候，你会明确知道预期结果。这让更改代码更容易并且更快。

### 关于速度的额外阅读资料

备注：我在我写完这个帖子之后才发现这个。我推荐你阅读这篇文章，如果你对于 Go 的速度想知道更多：[5 件使得Go很快的事](https://dave.cheney.net/2014/06/07/five-things-that-make-go-fast)

## Golang 之弊：缺乏泛型

Golang 没有泛型。我恨这点。毫无疑问这也是优点，但我真的很恨这点.

想象一下，在 C++ 下面工作却没有标准模板库（STL），并且还不允许你自己写一份的情景。想象一下，你决定抛弃 Array 和 Hash map，转而创造自己的数据结构，但却发现你并不能在其他任何地方复用的心痛。想象一下，你拥有一个类型语言的所有优点，却不能在一个强类型的类里复用。这让我想起了以前当我创建一些自定义网站的时候，我有写代码的一切能力，但我所能做的就只是复制粘贴 `functions.php` 到我客户的服务器上。

当到了即将要写一堆丑陋的代码去绕过不能用泛型的限制之际，我做了（或者考虑过）去做以下这些事情使得 Go 能像我想象中一样去运行。

### Empty Interface

这个解决方案需要使用 `empty interface` 。类型为 `interface{}` 的变量可以是任何东西。这对于在规定变量类型有很好的灵活度。但同时，有各种各样的原因会让这个变得很糟糕。

如果你使用了一个 `empty interface` 作为数值的数据结构的话，那每次你需要和这个数据结构打交道的时候，你都要写一堆废话去做类型申明，而且这还不会返回一个错误值。忽略潜在的错误会是一条直通灾难的捷径，因为一个简单的重构就破坏了类型安全。

如果你在函数的传入参数中使用了一个 `empty interface`，你很有可能会在函数中有一个用于类型断言的 switch 语句。在一个不同的变量类型下复用这样的函数就意味着在 `switch` 语句中多加一行。我宁愿去做一个脑白质切断术，因为这真的不是“复用”，并且如果有越多开发者，这越行不通。

### 复制/粘贴

我们的免费 [Watchdog](https://blog.bluematador.com/watchdog) 模块检测各类系统指标 - CPU，负荷，磁盘，网络等等。我写的第一个指标是 CPU 的各类占用率。当我写这些的时候，我赋予它们 `float32` 类型。然后，下一个指标是当前运行的进程数，这明显是一个属于 `unit` 干的活。

直到此时我才意识到 Go 跟我有私人恩怨.

我尝试了许多不同的方法去让查询语句和持久层能同时和 `float32` 以及 `unit` 兼容。我最不想做的就是复制/粘贴。幸运的是，我的故事是个成功的例子，我没有复制/粘贴。我用了次优类型（马上会谈到）。但那曾经是多么令人沮丧，我复制/粘贴了所有的代码 - 所有的逻辑，解析，持久性等等 - 并且准备提交代码。我真的做不到。

仅仅是 Go 语言规范几乎让我去提交一份复制/粘贴的复杂的查询语句和持久层代码这件事，就已经是一个让我换另一种语言的强有力理由了.

### 次优类型

当时我没有去复制/粘贴，而实际上我做的是在所有地方使用浮点数。没错， [Watchdog](https://blog.bluematador.com/watchdog) 模块使用浮点数去记录当前运行进程数，磁盘写入次数以及换入/换出次数。

这个办法在数值类的类型中大都是没有伤害的，这也就是为什么我最终把它用在查询语句与持久层之中。不适用这个办法的情况其实远多于那些适用的情况。如果你有数值类的类型，考虑下这个办法。如果你没有，你要回到 `empty interface` 加上类型断言或者使用复制/粘贴.

至少，这比 CPU 占用率用 0-100 的 4 字节整数表示来的要好。

### Unions

只是开个玩笑。Go 不支持 Unions，但这其实会非常有用。你认为呢?

----------------

via: https://blog.bluematador.com/posts/golang-pros-cons-devops-part-3-speed-lack-generics/

作者：[Matthew Barlocker](https://github.com/mbarlocker)
译者：[p31d3ng](https://github.com/p31d3ng)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

