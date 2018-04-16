已发布：https://studygolang.com/articles/12806

# Go 语言之美

最近，我在做兴趣项目的时候开始探索 Go 语言，被 Go 语言的美征服了。

Go语言的美在于它在灵活使用（常见于一些动态，解释型语言）和安全性能（常见于一些静态，编译语言）之间有一个很好的平衡。

除此之外，还有另外的两个功能让我觉得 Go 语言非常适合现代的软件开发。我会在之下优势的部分阐述。

其中之一是 **对语言并发性的一流支持**（通过 `goroutine`，和 `channels` 实现，下面解释）。 并发，通过其设计，使您能够有效地使用您的 CPU 马力。 即使您的处理器只有1个内核，并发的设计也能让您高效地使用该内核。 这就是为什么您通常可以在单台机器上运行数十万个并发 `goroutines`（轻量级线程）的原因。  `channels` 和 `goroutines` 是分布式系统的核心，因为它们抽象了生产者 - 消费者的消息范例。

我非常喜欢 Go 的另一个特性是接口 `(Interface)` 。 **接口为您的系统提供松耦合或分离组件。** 这意味着你的代码的一部分可以只依赖于接口类型，并不关心谁实现了接口或接口是如何实现的。 然后，您的控制器可以提供一个满足接口（实现接口中的所有功能）的代码的依赖关系。 这也为单元测试提供了一个非常干净的架构（通过依赖注入）。 现在，您的控制器可以注入代码所需的接口的模拟实现，以便能够测试它是否正确地执行其工作。

综上所诉我认为 Go 是一门很棒的语言。 特别是对于像云系统开发（Web 服务器，CDN，缓存等），分布式系统，微服务等用例。因此，如果你是一名工程师或初创公司试图决定你想要探索或尝试的语言，请认真的考虑一下 Go。

在这篇文章中，我将讨论该语言的以下几个方面：

a）介绍  
b）为什么需要Go  
c）目标受众  
d）Go的优势  
e）Go的弱点  
f）走向2  
g）Go的设计理念  
h）如何开始  
i）谁在使用Go

## 介绍

Go 是一款开源语言，由 Robert Griesemer，Rob Pike 和 Ken Thompson 在 Google 创建。开源代码意味着每个人都可以通过为新功能提供[建议](https://github.com/golang/go/wiki/ExperienceReports)，修复错误等来为语言做出贡献。语言的代码在[GitHub](https://github.com/golang/go)上。 [这里](https://golang.org/doc/contribute.html)提供了有关如何为语言做出贡献的文档。

## 为什么需要 Go

作者提到，设计新语言的主要动机是解决 Google 的内部软件工程问题。 他们还提到 Go 实际上是作为 C++ 的替代品而开发的。

Rob Pike 提到 Go 编程语言的目的：

> “因此，Go 的目的不是研究编程语言设计; 它是为了改善设计师和同事的工作环境。 Go 比编程语言研究更关注软件工程。 或者换句话说，就是为软件工程师服务的语言设计。”

**困扰Google软件工程视野的问题**（摘自(https://talks.golang.org/2012/splash.article)）：

a）缓慢的构建 - 构建有时需要一个小时才能完成  
b）不受控制的依赖  
c）每个程序员使用该语言的不同子集  
d）程序理解不佳（代码难以阅读，记录不当等等）  
e）重复努力  
f）更新的成本  
g）版本歪斜  
h）编写自动工具的难度  
i）跨语言构建  

**为了成功，Go必须解决这些问题**（摘自(https://talks.golang.org/2012/splash.article)）：

a）Go必须能大规模的使用，用于多人的大组，并且适用于有大量依赖程序的项目。  
b）Go的语法必须是让人熟悉的，大致类C。 谷歌需要在Go中快速提高程序员的效率，这意味着语言的语法的变化不能太激进。  
c）Go必须是现代的。 它应该具有像并发这样的功能，以便程序可以高效地使用多核机器。 它应该有内置的网络和Web服务器库，以便它有助于现代化的发展。

## 目标听众

Go 是一种系统编程语言。 对于诸如云系统（网络服务器，缓存），微服务，分布式系统（由于并发支持）而言，Go 确实非常出色。

## 优势

**a）静态类型：** Go 是静态类型的。 这意味着您需要在编译时为所有变量和函数参数（以及返回变量）声明类型。 虽然这听起来不方便，但这有一个很大的优势，因为在编译时本身会发现很多错误。 当你的团队规模增加时，这个因素起着非常重要的作用，因为声明的类型使得函数和库更易读，更容易理解。

**b）编译速度：**Go 代码编译速度**非常快**，因此您无需继续等待代码编译。 实际上，`go run` 命令会很快启动你的Go程序，所以你甚至不会觉得你的代码是先编译好的。 这感觉就像一种解释性语言。

**c）执行速度：** 根据操作系统（Linux/Windows/Mac）和代码正在编译的机器的 CPU 指令集体系结构（x86，x86-64，arm等），Go 代码直接编译为机器代码。 所以，它运行速度非常快。

**d）便携式：** 由于代码直接编译为机器码，因此，二进制文件变得便携。 这里的可移植性意味着你可以从你的机器（比如 Linux，x86-64）获取二进制文件并直接在你的服务器上运行（如果你的服务器也在 x86-64 架构上运行 Linux）。

由于 Go 二进制文件是静态链接的，这意味着您的程序需要的任何共享操作系统库都将在编译时包含在二进制文件中。它们在运行程序时不会动态链接。

这对于在数据中心的多台机器上部署程序具有巨大的好处。如果您的数据中心中有 100 台机器，只要将二进制文件编译为您的机器所运行的相同操作系统和指令集体系结构，就可以简单地将您的程序二进制文件 “scp” 到所有这些机器。 你不需要关心他们正在运行的 Linux 版本。 不需要检查/管理依赖关系。 二进制文件运行的过程中，所有的程序都跑起来了。

**e）并发性：** Go 对并发有一流的支持。 并发是 Go 的主要卖点之一。 语言设计师围绕托尼霍尔的[“Communicating Sequential Processes”](http://www.cs.cmu.edu/~crary/819-f09/Hoare78.pdf)论文设计了并发模型。

**Go 运行时允许您在机器上运行数十万个并发 `goroutine` 。**`Goroutine` 是一个轻量级的执行线程。 Go 运行时将这些 `goroutine` 复用到操作系统线程上。 这意味着多个 `goroutine` 可以在单个操作系统线程上同时运行。 Go 运行时有一个调度程序，其任务是调度这些 `goroutine` 执行。

这种方法有两个好处：

1. 初始化时的 `Goroutine` 具有 2KB 的堆栈。与一个一般为 1 MB 的 OS 线程堆栈相比，这非常小巧。当你需要同时运行几十万个不同的 goroutine 时，这个数字很重要。如果你要并行运行数千个 OS 线程，RAM 显然将成为瓶颈。

2. Go 可以遵循与 Java 等其他语言相同的模型，它支持与 OS 线程相同的线程概念。但是在这种情况下，OS 线程之间的上下文切换成本比不同的 `goroutine` 之间的上下文切换成本要大得多。

由于我在本文中多次提及“并发性”，因此我建议您查看 Rob Pike 关于[“并发性不是并行性”](https://www.youtube.com/watch?v=cN_DpYBzKso)的讨论。在编程中，并发是独立执行的进程的组成，而并行则是（可能相关的）计算的同时执行。除非你有一个拥有多个内核的处理器或者拥有多个处理器，否则你不能真正拥有并行性，因为 CPU 内核一次只能执行一件事。在单个核心机器上，只有并发才是幕后工作。 OS 调度程序针对处理器上的不同时间片调度不同的进程（实际上线程，每个进程至少有一个主线程）。因此，在某个时刻，您只能在处理器上运行一个线程（进程）。由于指令的执行速度很快，我们感觉到有很多事情正在运行。但实际上这只是一件事。

并发是一次处理很多事情。并行是一次做很多事情。

**f）接口：** 接口使松散耦合的系统成为可能。 Go 中的接口类型可以被定义为一组函数。而已。任何实现这些函数的类型都会隐式地实现接口，即不需要指定类型实现接口。这由编译器在编译时自动检查。

这意味着你的代码的一部分可以只依赖于一个接口类型，并不关心谁实现了接口或接口是如何实现的。然后你的主/控制器函数可以提供一个满足接口（实现接口中所有函数）的依赖关系。这也为单元测试提供了一个非常干净的架构（通过依赖注入）。现在，您的测试代码可以注入代码所需的接口的模拟实现，以便能够测试它是否正确地执行其工作。

虽然这对于解耦是非常好的，但另一个好处是您可以开始将您的体系结构视为不同的微服务。即使您的应用程序驻留在单个服务器上（如果您刚刚开始），也可以将应用程序中所需的不同功能设计为不同的微服务，每个微服务都实现它承诺的接口。所以其他服务/控制器只是调用界面中的方法，而不是实际关心它们是如何在幕后实现的。

**g）垃圾收集：** 与 C 不同，你不需要记住释放指针或担心 Go 中悬挂指针。垃圾收集器自动完成这项工作。

**h）不报错异常，自己处理错误：** 我喜欢 Go 没有其他语言具有的标准异常逻辑。去强迫开发人员处理“无法打开文件”等基本错误，而不是让他们将所有代码包装在 try catch 块中。这也迫使开发人员考虑需要采取什么措施来处理这些故障情况。

**i）惊人的工具**：关于Go的最好方面之一是它的工具。它有如下工具：

i) [Gofmt](https://blog.golang.org/go-fmt-your-code)：它会自动格式化和缩进你的代码，这样你的代码看起来就像这个星球上的每个 Go 开发者一样。这对代码可读性有巨大的影响。

ii) [Go run](https://golang.org/cmd/go/#hdr-Compile_and_run_Go_program)：编译你的代码并运行它们，都是:)。因此，即使 Go 需要编译，这个工具也让你觉得它是一种解释型语言，因为它只是编译你的代码的速度非常快，以致于当代码编译完成时你甚至不会感觉到它。

iii) [Go get](https://golang.org/cmd/go/#hdr-Download_and_install_packages_and_dependencies)：从 GitHub 下载库并将其复制到 GoPath，以便您可以将库导入到项目中

iv) [Godoc](https://godoc.org/golang.org/x/tools/cmd/godoc)：Godoc 解析您的 Go 源代码 - 包括注释 - 并以 HTML 或纯文本格式生成其文档。通过 Godoc 的网络界面，您可以看到与其所记录代码紧密结合的文档。只需点击一下，您就可以从函数的文档导航到其实现。

你可以在这里查看更多的[工具](https://golang.org/cmd/go/)。

**k）优秀的内建库：** Go 有很棒的内部库来帮助当前的软件开发。它们之中有一些是：

a) [net/http](https://golang.org/pkg/net/http/) - 提供 HTTP 客户端和服务器实现

b) [database/sql](https://golang.org/pkg/database/sql/) - 用于与 SQL 数据库交互

c) encoding/json - JSON 被视为标准语言的第一类成员 :)

d) html/templates - HTML 模板库

e) io/ioutil - 实现 I/O 实用程序功能

Go 的开发有很多进展。你可以在[这里](https://github.com/avelino/awesome-go)找到所有的 Go 库和框架，用于各种工具和用例。

## 弱点

**1.泛型的缺乏** - 泛型让我们在稍后指定待指定的类型时设计算法。假设您需要编写一个函数来对整数列表进行排序。稍后，您需要编写另一个函数来排序字符串列表。在那一刻，你意识到代码几乎看起来一样，但你不能使用原始函数，因为函数可以将一个整数类型列表或一个类型字符串列表作为参数。这将需要代码重复。因此，泛型允许您围绕稍后可以指定的类型设计算法。您可以设计一个算法来排序T类型的列表。然后，您可以使用整数/字符串/任何其他类型调用相同的函数，因为存在该类型的排序函数。这意味着编译器可以检查该类型的一个值是否大于该类型的另一个值（因为这是排序所需的）

通过使用语言的空接口 `（interface {}）` 功能，可以在 Go 中实现某种通用机制。但是，这并不理想。

泛型是一个备受争议的话题。一些程序员发誓。而其他人不希望它们包含在该语言中，因为泛型通常是编译时间和执行时间之间的折衷。

这就是说，Go 的作者在 Go 中表达了对实施某种泛型机制的开放性。但是，这不仅仅是泛型。泛型只有在语言的所有其他功能都能正常工作时才能实现。让我们拭目以待 Go 2 是否为他们提供了某种解决方案。

**2.缺乏依赖管理** - Go1 的承诺意味着 Go 语言及其库在 Go 1 的生命周期中不能更改其API。这意味着您的源代码将继续为 Go 1.5 和Go 1.9 编译。因此，大多数第三方 Go 库也遵循相同的承诺。由于从 GitHub 获得第三方库的主要方式是通过 'go get' 工具，因此，当您执行 `go get github.com/vendor/library` 时，您希望他们主分支中的最新代码不会更改库 API。虽然这对临时项目很酷，因为大多数库没有违背承诺，但这对于生产部署并不理想。

理想情况下应该有一些依赖版本的方法，这样你就可以简单地在你的依赖文件中包含第三方库的版本号。即使他们的 API 改变了，你也不需要担心，因为新的 API 将带有更新的版本。您稍后可以回头查看所做的更改，然后决定是否升级您的依赖文件中的版本并根据 API 接口中的更改更改您的客户端代码。

Go 的官方实验 [dep](https://github.com/golang/dep) 应该很快成为这个问题的解决方案。可能在 Go 2。
（译注：vgo 会解决此问题）

## Go 2 的开发

我非常喜欢作者采用这种开源方法来实现这种语言。如果你想在 Go 2 中实现一个功能，你需要编写一个文档，你需要：

a）描述你的用例或问题  
b）说明如何使用 Go 无法解决用例/问题  
c）描述一个问题究竟有多大（有些问题不够大或者没有足够的重点来解决某个特定时刻的问题）。  
d）可能的话，提出如何解决问题的解决方案  

作者将对其进行审查并将其链接到[此处](https://github.com/golang/go/wiki/ExperienceReports)。所有关于问题的讨论都将在邮件列表和问题跟踪器等公共媒体上进行。

在我看来，语言的两个最紧迫的问题是泛型和依赖管理。依赖管理更多的是发布工程或工具问题。希望我们会看到 [dep](https://github.com/golang/dep)（官方实验）成为解决问题的官方工具。鉴于作者已经表达了对该语言的泛型的开放性，我很好奇他们是如何实现它们的，因为泛型以编译时间或执行时间为代价。

## Go的设计理念

在 Rob Pike 的的 talk [“简单就是复杂”](https://talks.golang.org/2015/simplicity-is-complicated.slide#18)中，他提到的一些设计理念让我觉得很有亮点。

具体来说，我喜欢的是：

a）**传统上，其他语言都希望不断添加新功能。**通过这种方式，所有的语言都只是增加了臃肿性，编译器及其规范中的复杂性过高。如果这种情况持续下去，每种语言将在未来看起来都一样，因为每种语言都会不断添加它没有的功能。考虑一下，JavaScript 添加面向对象的功能。 Go 作者故意没有在语言中包含很多功能。**只有那些作者达成一致意见的特征才被包括在内，因为那些真正觉得他们确实为语言所能达到的东西带来了价值的特征。**

b）**特征像解空间中的正交向量。** 重要的是能够为您的用例选择和组合不同的向量。**而这些载体应该自然而然地相互配合。意味着该语言的每个功能都可以与其他任何可预测的功能一起工作。** 这样，这些功能集涵盖整个解决方案空间。实现所有这些非常自然的功能，给语言实现带来了很多复杂性。但是该语言将复杂性抽象化并为您提供简单易懂的界面。因此，简单性就是隐藏复杂性的艺术

c）**可读性的重要性往往被低估。** 可读性在设计编程语言中是最重要的事情之一，因为维护软件的重要性和成本非常高。过多的功能会影响语言的可读性。

d）**可读性也意味着可靠性。** 如果一门语言很复杂，你就必须了解更多的东西来阅读和编写代码。同样，要调试它并能够修复它。这也意味着您的团队中的新开发人员需要更大的扩展时间，才能让他们理解语言，直到他们可以为您的代码库贡献力量。

## 如何开始

您可以下载Go并按照此处的安装说明进行操作。

这里是开始使用Go的[官方指南](https://tour.golang.org/welcome/1)。[Go by example](https://gobyexample.com/)也是一本好书。

如果你想读一本书，[The Go Programming Language](https://www.amazon.com/Programming-Language-Addison-Wesley-Professional-Computing/dp/013419044)是一个很好的选择。 它的编写方式与传说中的[C 语言白皮书](https://www.amazon.com/Programming-Language-2nd-Brian-Kernighan/dp/0131103628)相似，由[Alan A. A. Donovan](https://www.informit.com/authors/bio/cd7c1e12-138d-4bf9-b609-e12e5a7fa866)和[Brian W. Kernighan](https://en.wikipedia.org/wiki/Brian_Kernighan)撰写。

您可以加入 [Gophers Slack](https://invite.slack.golangbridge.org/) 频道，与社区合作并参与有关该语言的讨论。

## 哪些公司在使用 Go

[很多公司](https://github.com/golang/go/wiki/GoUsers)已经开始在 Go 上投入很多资金使用 Go。一些有名的大厂包括：

Google - [Kubernetes](http://kubernetes.io/)，[MySQL 扩展基础架构](http://vitess.io/)，[dl.google.com](https://talks.golang.org/2013/oscon-dl.slide#1)（下载服务器）

BaseCamp - [Go at BaseCamp](https://signalvnoise.com/posts/3897-go-at-basecamp)

CloudFlare - [博客](https://blog.cloudflare.com/go-at-cloudflare/)，[ArsTechnica 文章](https://arstechnica.com/information-technology/2013/02/cloudflare-blows-hole-in-laws-of-web-physics-with-go-and-railgun/)

CockroachDB - [为什么 Go 是 CockroachDB 的正确选择](https://www.cockroachlabs.com/blog/why-go-was-the-right-choice-for-cockroachdb/)

CoreOS - [GitHub](https://github.com/coreos/)，[博客](https://blog.gopheracademy.com/birthday-bash-2014/go-at-coreos/)

DataDog - [Go at DataDog](https://blog.gopheracademy.com/birthday-bash-2014/go-at-datadog/)

DigitalOcean - [让您的开发团队开始使用 Go](https://blog.digitalocean.com/get-your-development-team-started-with-go/)

Docker - [为什么我们决定在 Go 中编写 Docker](https://www.slideshare.net/jpetazzo/docker-and-go-why-did-we-decide-to-write-docker-in-go/)

Dropbox - [开源我们的 Go 库](https://blogs.dropbox.com/tech/2014/07/open-sourcing-our-go-libraries/)

Parse - [我们如何将我们的 API 从 Ruby 转移到 Go 并保存了我们的理智](http://blog.parse.com/learn/how-we-moved-our-api-from-ruby-to-go-and-saved-our-sanity/)

Facebook - [GitHub](https://github.com/facebookgo/)

英特尔 - [GitHub](https://github.com/clearcontainers)

Iron.IO - [Go after 2 years in Production/](https://www.iron.io/go-after-2-years-in-production/)

MalwareBytes - [每分钟处理 100 万个请求，使用 golang](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/)

Medium - [How Medium goes Social](https://medium.engineering/how-medium-goes-social-b7dbefa6d413)

MongoDB - [Go Agent](https://www.mongodb.com/blog/post/go-agent-go)

Mozilla - [GitHub](https://github.com/search?o=desc&q=org%3Amozilla+org%3Amozilla-services+org%3Amozilla-it+language%3AGo&ref=searchresults&s=stars&type=Repositories&utf8=%E2%9C%93)

Netflix - [GitHub](https://github.com/Netflix/rend)

Pinterest - [GitHub](https://github.com/pinterest?language=go)

Segment - [GitHub](https://github.com/segmentio?language=go)

SendGrid - [如何说服您的公司与 Golang 一起使用](https://sendgrid.com/blog/convince-company-go-golang/)

Shopify - [Twitter](https://twitter.com/burkelibbey/status/312328030670450688)

SoundCloud - [进入SoundCloud](https://developers.soundcloud.com/blog/go-at-soundcloud)

SourceGraph - [YouTube](https://www.youtube.com/watch?v=-DpKaoPz8l8)

Twitter - [每天处理 5 亿次会议](https://blog.twitter.com/engineering/en_us/a/2015/handling-five-billion-sessions-a-day-in-real-time.html)

Uber - [博客](https://eng.uber.com/go-geofence/)，[GitHub](https://github.com/uber?language=go)

----------

via: https://hackernoon.com/the-beauty-of-go-98057e3f0a7d

作者：[Kanishk Dudeja](https://hackernoon.com/@kanishkdudeja)
译者：[tingtingr](https://github.com/wentingrohwer)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出