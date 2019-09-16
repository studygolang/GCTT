首发于：https://studygolang.com/articles/23445

# Golang vs Python：选哪个？

对开源开发来说 Golang 和 Python 哪个语言更好，我们详细分析对比一下。

在任何项目开始之前，大多开发团队需要通过多次会议讨论来确定最适合他们项目的编程语言。很多时候他们会在 Python 和 Golang 中间纠结。在这篇 Golang vs. Python 的博文中，我将亲自从多角度对比这两种语言，以帮你确定哪种语言最适合你。主要从以下几个方面比较：

- 性能
- 可扩展性
- 应用
- 执行
- 库
- 代码可读性

让我们开始吧。在对比开始之前，让我来对这两种语言做简要介绍吧。

[![](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-vs-python-which-one/maxresdefault.jpg)](https://www.youtube.com/watch?v=I6f0g0xfuF8 "Go vs Python Comparison | Which Language You Should Learn In 2018? | Edureka")

## Golang

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-vs-python-which-one/Golang-Logo-Golang-Tutorial-Edureka-250x300.jpg)

[Golang](https://www.edureka.co/blog/golang-tutorial "Golang") , 也就是我们常说的 Go，是由 Google 开发的一种计算机编程语言。Golang 是于 2007 年在 Google 开始开发的，2009 年面世。Go 语言的三位主要开发人员分别是 Google 的 Robert Geriesemer，Rob Pike 和 Ken Thompson。这几位一直以来的目标是创建一种语法上与 C 语言相似，又能像 C++ 一样消除「多余的垃圾」的语言。以致于 Go 语言包含现代多种语言的特性，如方法和运算符的重载、指针运算、类型继承。最终，造就了一个带有轻量并强大库以及拥有无敌的性能和速度的静态类型语言。

这就是关于 Go 语言的内容。下面来说说 Python。

## Python

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-vs-python-which-one/Python-Logo-Golang-vs-Python-Edureka-215x300.png)

Python 是一种多用途的编程语言，换句话来说它几乎可以做任何事情。Python 是由一位荷兰程序员 Guido van Rossum 编写，于 1991 年首次发布。Python 最重要的一方面是它是一种解释型语言，这就意味着 Python 代码不会在运行时被翻译成机器语言，而大多数编程语言会在代码编译过程中完成这种转换。这种编程语言（解释型）也被称为「脚本语言」，因为它们最初是被用来做一些小项目的。

OK，既然我已经向大家粗略地介绍了两种语言，那让我们继续将他们做对比吧。

## Golang vs. Python：性能

首先我们要对比的是这两种语言的性能。比较性能有一个很好的方法是处理复杂的数学问题。虽然不完全公平，但是在谈及处理问题时的内存使用率和耗时时，必然能够证明这一点。

我们同时用两种语言处理了三个问题，即 *Mandelbrot 方程 *、*n-body 问题 * 以及 *fasta*。这些都是需要进行大量计算的复杂的问题，所以是一种非常不错的测试语言性能和内存管理的方法。

抛开拿它做性能测试不说，这几个问题都很有意思，值得一看。而我们现在的关注点在 Golang 和 Python 的性能表现。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-vs-python-which-one/Performance-Golang-vs-Python-Edureka-1.png)

很明显，在性能方面 Golang 胜过 Python。

OK, 继续比较下一项：可扩展性。

## Golang vs. Python：可扩展性

如今，构建一个高可扩展性的应用是一门艺术。如果不做到扩展，那将对业务产生不利影响。 Golang 在设计的时候便一直在考虑着这件事。Golang 的初衷是帮助 Google 的开发者解决内部大量的问题，这基本上涉及到成千上万的开发者在寄宿于成千上万集群的大型软件服务。这就是 Golang 具有内置支持并发进程通道（也就是并发性）的原因。而 Python 并不支持并发，它只是通过多线程来实现并行。

让我们来了解一下并发和并行。

### 并发和并行

并发的意思是说，一个应用在多个任务里同时（并发地）处理多个进程。如果计算机只有一个 CPU，则应用程序可能无法在同一时间在多个任务上取得进展，但应用内的线程会在同一时间段内被执行。在下一个任务被执行之前，当前任务并没有完全完成（交替执行）。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-vs-python-which-one/Concurrency-Golang-vs-Python-Edureka-250x300.png)

并行是说应用将它的任务分成多个能在同一时刻执行的多个子任务，例如在多个 CPU 上同时执行。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-vs-python-which-one/Parallelism-Golang-vs-Python-Edureka-254x300.png)

显而易见，天然支持并发的编程语言更适合大型、要求高可扩展性的项目。

让我们继续比较这两种语言的应用。

## Golang vs. Python：应用

本节中没有一个很明显的赢家，因为每种编程语言都有自己特定的目的和用途。比如 Javascript 主要用于 Web 开发。同样的，[Python](https://www.edureka.co/blog/python-tutorial/) 被广泛地用于数据分析、人工智能、深度学习以及 Web 开发。在刚才所说的领域中，使用 Python 更易于开发，这定要归功于这些强大的库了。

而 Golang 更多的用于系统编程。这归结于它天然支持并发，同时它在云计算和集群计算领域中使用广泛。由于 Golang 拥有强大易用的库，可以让你很快搭建出一个 Web 服务，所以 Golang 也被大量用于 Web 开发，增速也很大。如果你也想学 Go 语言中这一很酷的东西，可以直接看我的 [Golang tutorial](https://dzone.com/articles/golang-tutorial-learn-golang-by-examples) 这篇文章。

## Golang vs. Python：执行

现在我们来比较一下 Go 语言代码和 Python 代码是如何被执行的。首先我们要明确的是 Python 是动态类型语言，而 Go 是静态类型语言。Python 和 Go 分别使用解释器和编译器。

为了了解为什么我要对比语言的这一参数，我们必须要知道动态类型语言和静态类型语言的区别。

变量类型被显示声明给编译器，以致于细微的 bug 也能被很容易地捕获到。而在动态类型语言中，类型推断由解释器实现，而解释器在推断类型的过程中可能会出错，从而导致存在遗留 bug。

当开发者想创建一个相当大的项目时，会因为 Python 语言的动态类型这一特性而受限，而 Go 语言可以灵活应用于任意规模的项目。

下面开始比较它们的库吧。

## Golang vs. Python：库

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-vs-python-which-one/Libraries-Golang-vs-Python-Edureka-342x300.png)

强大的库是开发者的福音，因为它能使我们的开发工作更容易。因此，拥有一个优秀的库对编程语言来说是至关重要的。在本节的比较中，Python 肯定以量获胜。比如可以让你使用数组来处理复杂的矩阵问题的 Numpy 库，专注深度学习的 Tensorflow 库和 Scikit Learn 库、针对图像处理的 OpenCV、数据分析的 Pandans、可视化的 matplotlib，等等等等 ~。讲真，如果 Python 仅是因为一件事而闻名，那必定是它的库。但这并不意味着 Go 逊色于它。当 Go 在被开发的时候，Google 将一些很重要的库以内置的形式作为 Go 语言的一部分。虽然从数量上来讲没有 Python 的那么猛，但它的库所涉及的领域和 Python 是一样广的。它有针对 Web 开发、数据库处理、并发编程以及加密的强大的库。

最后一个比较点，可读性！

## Golang vs. Python：可读性

当你为客户开发软件时，一般都是和十人团队或百人团队合作开发。这时，代码可读性会成为被大家考虑的重要因素。

可能大部分人认为 Python 在可读性上更胜一筹，但我有着不同的观点，且听我说完。先看一下 [Python sure has fantastic readability](https://dzone.com/refcardz/core-python)，但在我看来，他们有点说得过头了。在 Python 中，可能有 10 种不同的方式来表达相同的东西，通常这会导致当代码很大或者协作的人很多时产生混淆。

另一方面，Go 在编程的时候有着严格的规则约束。它不允许不使用的包被 import，或者不被使用的变量被声明。这便意味着在大型团队中有更明确的方法对代码有着更好的理解，但有谁会去关心多功能性呢，尤其是在进行核心开发的时候。Golang 的语法对于初学者来说很不友好，但比 C 或 C++ 好很多。所以对于代码的可读性，我更倾向于 Golang。

大家都看到了，在我看来，作为一门编程语言，Glang 在很多方面都胜过 Python。当然 Golang 不像 Python 一样有名，Python 这几年已经在充斥在整个互联网中，但是 Go 在这方面也是迎头赶上。如果你对我的看法有异议，可在下方评论区进行评论。我希望我可以帮助你确定哪门编程语言对你的项目更好。请继续关注更多 Golang 的博客。

---

via：https://dzone.com/articles/golang-vs-python-which-one-to-choose

作者：[Aryya Paul](https://dzone.com/users/3510559/aryya-paul.html)
译者：[BeGemini](https://github.com/BeGemini)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
