# 为什么说 Golang 是 DevOps 专业人士的第一首选？
Golang 是当今最受欢迎的编程语言之一，现在就让我们来看看它在 DevOps 空间中能够做什么？

Golang，也称为 “Go”，是一种具备快速和高性能的编译型语言，这是被设计成为易于阅读和理解的原因。Go 是由 Rob Pike，Robert Griesemer 和 Ken Thompson 等人在 Google 时编写的，于 2009 年 11 月首次发布。

Golang 被设计成高度简洁和易于理解的语法。

这是 Golang 中经典的 “hello world” 示例代码。
```Golang
package main
import "fmt"
    func main() {
    fmt.Println("hello world")
}
```
要想运行这段代码，就要在 `hello-world.go` 所在目录中输入以下命令，并且使用 `go run` 运行。
```shell
$ go run hello-world.go
hello world
$ go build hello-world.go
$ ls
hello-world hello-world.go
$ ./hello-world
hello world
```

## Go 的介绍
Go 诞生于 2007 年，当时多核 CPU 的架构随处可见，而且没有编程语言能够简化多线程应用程序的开发工作。安全和高效地管理不同线程是开发人员的重大责任。这和其他编程语言不同，Go 虽然很年轻，但是也很强大。Goroutines 在另一个层面上彻底革新了竞争性编程。

经过测试和证明，用 Go 编写的应用程序具有高性能和可伸缩性。Golang 是一种非常高效的语言，就像 C/C++ 一样，还具备像 Java 一样处理并行任务的特性，同时兼具 Python 和 Perl 代码的易于阅读性。相比其他的编程语言，Golang 具有无可争议的架构优势。

Go 还被一些大公司使用，例如 BBC、Uber、Novartis、Basecamp and Soundcloud。Uber 公司报道过更高的吞吐量、高性能、延迟和正常运行时间。英国广播公司（BBC）是一个闻名于广播世界新闻中的机构，它将 Go 应用于 Web 后端领域，包括网络爬虫和网页数据提取。而 SoundCloud 公司则将 Go 用于构建和部署系统中。

以下是关于 Go 编程语言的 Google 趋势概览，它正在持续且稳定的增长。

## 为什么选择 Go ？
对于具备 C/C++ 学习经验的程序员来讲，学习 Go 是一件毫不费力的事情，并且将祖传代码转换成 Go 程序也是非常简单的。作为一种编译型的静态语言，它比解释型语言要快得多，同时具备了大部分的性能优势。

- Go 作为一种与 C 很相似的编程语言，但是除了具有 C 语言的特性之外，Go 还提供了内存安全性、[垃圾回收](https://dzone.com/articles/garbage-collection-a-brief-introduction)、[结构化类型](https://dzone.com/articles/dynamic-static-optional-structural-typing-and-engi)和 CSP 风格的并发性。
- 在最近的 [Stack Overflow 2020 的调查结果](https://insights.stackoverflow.com/survey/2020#technology-most-loved-dreaded-and-wanted-languages-loved)中，Go 是开发人员中最喜欢和最想要使用的编程语言之一。

### 最喜欢和最想要使用的编程语言
- Go 很适合用于一般绩效导向的云计算软件。流行的 DevOps 工具是用 Go 编写的，例如 Docker ，甚至是开源的容器编排系统 Kubernetes 都是用 Go 编写的。自 2011 年以来，YouTube 一直在使用 [Vitess](https://opensource.google/projects/vitess) ，它是一个由 Google 构建的分布式数据库系统，而且这个分布式数据库的 MySQL 后端是由 Golang 构建。
- 在 [2018 年的 Stack Overflow 调查结果](https://insights.stackoverflow.com/survey/2018/#technology)中，Golang 排名第五。根据 [GitHub 关于 2018 年的第二季度报告](https://madnight.github.io/githut/#/pull_requests/2019/4)，Golang 的整体增长率接近于 7% ，与上一季度相比增长 1.5 点。到 2019 年第四季度，Golang 的整体增长率已经达到 8% 。

## Go 如此受欢迎的原因
- Go 是一种静态类型的编译语言，因此你可以更早地发现问题。
- Go 可以被立即编译为机器代码，因此它的编辑/刷新周期相对较快，并且仍然会编译出更高效的机器代码。
- Go 的语法设计使得编写高度并发的网络程序变得容易。
- Go 内置了许多库来支持测试，您可以轻松地定义和测试模块，这进一步提高了代码规范。
- Go 跨平台特性使得移植代码非常容易，这也是 Go 的最大优势。
- Go 提供了自动的代码格式化、代码检查和审核工具，它们作为软件包的默认部分；Go 编译器甚至会执行像变量没有被使用的操作。这使其成为一种专业的语言。
- 正是因为 Go 对并行和并发的原生支持，所以它才会变得如此特别。对于需要大量并发或并行处理、联网、海量计算的应用程序，使得 Go 成为一种更完美的编程语言。
- Go 是实现云兼容性的最佳选择。Go 还具有更好的垃圾回收能力和性能优异的 network 包，而且还解决了变量没有被使用、多编译和交叉编译的问题。

## 让我们看一些是谁在使用 Go 的实际案例

### SendGrid 投入 Go
SendGrid 是一个客户沟通的平台，并于 2014 年将 Go 作为主要开发语言。SendGrid 开发团队需要从根本上转变它们的开发语言，归结为 Scala、Java 和 Go 之间的竞争。当时，SendGrid 在开发中面临最大的挑战是并发编程。寻找具有并发的异步编程的特性，然后将其作为编程语言中的一部分，这是 SendGrid 选择 Go 最令人信服的原因之一。

可以在他们的博客上阅读全文：[如何说服您的公司选择Golang？](https://sendgrid.com/blog/convince-company-go-golang/)

### Hexac 已经从 Python 转换到 Go
Hexac 的联合创始人兼 CTO Tigran Bayburtsyan 写了一篇独家文章，分享了他的公司[从 Python 转到 Go 的原因](https://hackernoon.com/5-reasons-why-we-switched-from-python-to-go-4414d5f42690)。根据他们的代码库统计信息，在使用 Go 重构了所有项目之后，他们的代码量比以前减少了 64% 。

由于 Go 内置的语言特性，他们节省了大量资源（内存和 CPU ）。

Go 为他们的开发团队提供了极大的灵活性，可以在所有用例中使用单一的语言，并且效率很高。在 Hexact 股份有限公司中，他们的后端和API服务的性能提高了约 30％。
现在，他们可以实时地处理日志，然后将其传输到数据库；在单个或多个服务中，使用 Websocket 进行流式传输。这就是 Go 语言带来的出色表现。

### Salesforce 抛弃 Python 而选择了 Go
在 2017 年推出 Einstein Analytics 之前，Salesforce 使用 Google 流行的 Go 语言完全重构了他们的后端。Salesforce 首席架构师 Guillaume Le Stum 表示：“ Python 并不能很好地完成多线程工作，而 Go 是专为 Google 生产系统中的重型应用程序而构建的，这门编程语言已经通过了 Google 的测试和许可。因此 Salesforce 选择将 Einstein Analytics（Salesforce 的重要组成部分）从混合 C-Python 应用程序转变为完全 Go 应用程序。请阅读原文：[我们为什么在 Einstein Analytics 放弃 Python 而选择了 Google 的 Go](https://www.zdnet.com/article/salesforce-why-we-ditched-python-for-googles-go-language-in-einstein-analytics/)。

### Containerum 优先选择 Go
Containerum 是一个使用 Go 作为主要开发语言的容器管理平台，已经有大约四年的历史，尽管面临着某些挑战，工程团队仍然认为这是一个不错的选择。选择在 Containerum Platform 上使用 Go 的主要原因是，它由一组比较小的服务组成，这些服务与其他组件进行通信。为了确保这一点，我们非常需要确保接口的兼容性并且要编写简洁、易于阅读和维护的代码。

Go 支持添加补丁并允许在代码库中使用准备就绪的组件。例如，图像名称解析认证，关键对象模型等，这是 Containerum 选择Go的原因之一。

Containerum 之所以考虑使用 Go 语言，是因为它具有许多专业的功能，例如静态类型，语法简洁，标准库，出色的性能，快速的编译等。请阅读原文：[我们为什么使用 Go 来为 Kubernetes 开发 Containerum 平台](https://medium.com/containerum/why-we-use-go-to-develop-containerum-platform-for-kubernetes-3a33d5bdc5ec)。

### 流行的 DevOps 工具是用 Go 编写的

> Kubernetes, Docker, and Istio

像 Google 这样的巨型公司曾经考虑使用其他语言编写 Kubernetes ，但是据 Kubernetes 的联合创始人 Joe Beda 称，这些语言都没有 Go 这么有效。这里有一些关于 Kubernetes 为什么要使用 Go 编写的原因，其中包括广泛的标准库、快速的工具、内置并发、垃圾回收和类型安全等。根据 Joe 的说法，Go 中的这些模式和工具鼓励 Kubernetes 开发团队编写结构合理和可重用的代码，这些代码同时将会为他们提供高度的灵活性和速度。

[Docker](https://www.docker.com/) 是使用 Go 语言的最大用户。Docker 开发团队之所以喜欢使用 Go，是因为 Go 为他们提供了许多好处：无需依赖项的静态编译、自然语言、完整的开发环境、广泛和强大的标准库和数据类型、强大的鸭子类型以及使用最小的代价为多种架构进行构建的能力。

[Istio](https://istio.io/) 是 Kubernetes 生态系统的一部分，也是用 Go 编写的。

因为 Kubernetes 也是用 Go 编写的，所以 Istio 使用 Go 进行开发也是一种完美的方法。这不仅是 Go 适应分散的和分布式网络项目的原因之一，也是在 Istio 选择 Go 的主要原因之一。

更多详细内容请访问：[证明 Google Go 功能强大的10个开源项目](https://www.infoworld.com/article/3442978/10-open-source-projects-proving-the-power-of-google-go.html)。

如果您正在使用 Golang 编写应用程序，那么实践CI / CD有多困难？这很难，不是吗？

好吧，不是现在。但是借助 [Go Center GOPROXY](https://search.gocenter.io/) 等最新技术，增长 CI / CD 质量的途径将会变得更加清晰。GoCenter 是不可变的 Go 模块的公共的中央仓库，它允许您搜索模块和版本，也可以轻松地将模块添加到中央仓库，然后公开分享它们。

---
via: https://dzone.com/articles/why-golang-is-top-of-mind-for-devops-professionals

作者：[Pavan Belagatti](https://dzone.com/users/2879134/pavanshippable.html)
译者：[sunlingbot](https://github.com/sunlingbot)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出