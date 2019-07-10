# Golang vs. Python:选哪个？

对开源开发来说Golang和Python哪个语言更好，我们详细分析对比一下。
在任何项目开始之前，大多开发团队需要通过多次会议讨论来确定最适合他们项目的编程语言。很多时候他们会在Python和Golang中间纠结。在这篇Golang vs. Python的博文中，我将亲自从多角度对比这两种语言，以帮你确定哪种语言最适合你。主要从以下几个方面比较：

- 性能
- 可扩展性
- 应用
- 执行
- 库
- 代码可读性

让我们开始吧。在对比开始之前，让我来对这两种语言做简要介绍吧。

[![](https://i.ytimg.com/vi/I6f0g0xfuF8/maxresdefault.jpg)](https://www.youtube.com/watch?v=I6f0g0xfuF8 "Go vs Python Comparison | Which Language You Should Learn In 2018? | Edureka")

## Golang

![](https://d1jnx9ba8s6j9r.cloudfront.net/blog/wp-content/uploads/2018/09/Golang-Logo-Golang-Tutorial-Edureka-250x300.jpg)

[Golang](https://www.edureka.co/blog/golang-tutorial "Golang"),也就是我们常说的Go，是由Google开发的一种计算机编程语言。Golang是于2007年在Google开始开发的，2009年面世。Go语言的三位主要开发人员分别是Google的Robert Geriesemer,Rob Pike和Ken Thompson.这几位一直以来的目标是创建一种语法上与C语言相似，又能像C++一样消除“多余的垃圾”的语言。以致于Go语言包含现代多种语言的特性，如方法和运算符的重载、指针运算、类型继承。最终，造就了一个带有轻量并强大库以及拥有无敌的性能和速度的静态类型语言。

这就是关于Go语言的内容。下面来说说Python.

## Python

![](https://d1jnx9ba8s6j9r.cloudfront.net/blog/wp-content/uploads/2018/09/Python-Logo-Golang-vs-Python-Edureka-215x300.png)

Python是一种多用途的编程语言，换句话来说它几乎可以做任何事情。Python是由一位荷兰程序员Guido van Rossum编写，于1991年首次发布。Python最重要的一方面是它是一种解释型语言，这就意味着Python代码不会在运行时被翻译成（计算机可读的格式）机器语言，而大多数编程语言会在代码编译过程中完成这种转换。这种编程语言（解释型）也被称为“脚本语言”，因为它们最初是被用来做一些小项目的。

OK，既然我已经向大家粗略地介绍了两种语言，那让我们继续将他们做对比吧。

## Golang vs. Python:性能

首先我们要对比的是这两种语言的性能。比较性能有一个很好的方法是处理复杂的数学问题。虽然不完全公平，但是在谈及处理问题时的内存使用率和耗时时，必然能够证明这一点。虽然不完全公平，但从处理问题时的内存使用率以及耗时，足以证明这一点。

我们同时用两种语言处理了三个问题，即*Mandelbrot equation(Mandelbrot 方程)*、*n body problem(n-body 问题)*以及*fasta*.这些都是需要进行大量计算的复杂的问题，所以是一种非常不错的测试语言性能和内存管理的方法。

抛开拿它做性能测试不说，这几个问题都很有意思，值得一看。而我们现在的关注点在Golang和Python的性能表现。

![](https://d1jnx9ba8s6j9r.cloudfront.net/blog/wp-content/uploads/2018/09/Performance-Golang-vs-Python-Edureka-1.png)

很明显，在性能方面Golang胜过Python.

OK,继续比较下一项：可扩展性。

## Golang vs. Python:可扩展性

如今，构建一个高可扩展性的应用是一门艺术。如果不做到扩展，那将对业务产生不利影响。Golang在设计的时候便一直在考虑着这件事。Golang的初衷是帮助Google的开发者解决内部大量的问题，这基本上涉及到成千上万的开发者在寄宿于成千上万集群的大型软件服务。这就是Golang具有内置支持并发进程通道（也就是并发性）的原因。而Python并没有并发性，它只是通过多线程来实现并行。

让我们来了解一下并发和并行。

### 并发和并行

并发的意思是说，一个应用在多个任务里同时（并发地）处理多个进程。如果计算机是单核CPU，那这个程序可能不会在多个任务里同时处理进程，但应用内的线程会在同一时间段内执行。在下一个任务被执行之前，当然任务并没有完全完成（交替执行）。

![](https://d1jnx9ba8s6j9r.cloudfront.net/blog/wp-content/uploads/2018/09/Concurrency-Golang-vs-Python-Edureka-250x300.png)

并行是说应用将它的任务分成多个能在同一时刻执行的多个子任务，例如在多个CPU上同时执行。

![](https://d1jnx9ba8s6j9r.cloudfront.net/blog/wp-content/uploads/2018/09/Parallelism-Golang-vs-Python-Edureka-254x300.png)

显而易见，天然支持并发的编程语言更适合大型、要求高可扩展性的项目。

让我们继续比较这两种语言的应用。

## Golang vs. Python:应用

本节中没有一个很明显的赢家，因为每种编程语言都有自己特定的目的和用途。比如Javascript主要用于web开发。同样的，[Python](https://www.edureka.co/blog/python-tutorial/)被广泛地用于数据分析、人工智能、深度学习以及web开发。This can be mostly credited to the insane libraries that are available in Python that make life in the said fields a whole lot easier.被认为在Python中可用的这些强大的包使得在刚才所说的领域中更简单。

而Golang更多的用于系统编程。这归结于它天然支持并发，同时它在云计算和集群计算领域中使用广泛。由于Golang拥有强大易用的库，可以让你很快搭建出一个web服务，所以Golang也被大量用于web开发，增速也很大。如果你也想学Go语言中这一很酷的东西，可以直接看我的[Golang tutorial](https://dzone.com/articles/golang-tutorial-learn-golang-by-examples)这篇文章。

## Golang vs. Python:执行

现在我们来比较一下Go语言代码和Python代码是如何被执行的。首先我们要明确的是Python是动态类型语言，而Go是静台类型语言。Python和Go都需要各自的解释器刚编译器。

为了了解为什么我要对比语言的这一参数，我们必须要知道动态类型语言和静态类型语言的区别。

变量类型被显示声明给编译器，以致于细微的bug也能被很容易地捕获到。而在动态类型语言中，类型推断由解释器实现，而解释器在推断类型的过程中可能会出错，从而导致存在遗留bug。

当开发者想创建一个相当大的项目时，会因为Python语言的动态类型这一特性而受限，而Go语言可以灵活应用于任意规模的项目。

下面开始比较他们的库吧。

##　Golang vs. Python:库

![](https://d1jnx9ba8s6j9r.cloudfront.net/blog/wp-content/uploads/2018/09/Libraries-Golang-vs-Python-Edureka-342x300.png)

强大的库是开发者的福音，因为它能使我们的开发工作更容易。因此，拥有一个优秀的库对编程语言来说是至关重要的。在本节的比较中，Python肯定以量获胜。比如可以让你使用数组来处理复杂的矩阵问题的Numpy库，专注深度学习的Tensorflow库和Scikit Learn库、针对图像处理的OpenCV、数据分析的Pandans、可视化的matplotlib，等等等等~。讲真，如果Python仅是因为一件事而闻名，那必定是它的库。但这并不意味着Go逊色于它。当Go在被开发的时候，Google将一些很重要的库以内置的形式作为Go语言的一部分。虽然从数量上来讲没有Python的那么猛，但它的库所涉及的领域和Python是一样广的。它有针对web开发、数据库处理、并发编程以及加密的强大的库。

最后一个比较点，可读性！

## Golang vs. Python:可读性

当你为客户开发软件时，一般都是和十人团队或百人团队合作开发。这时，代码可读性会成为被大家考虑的重要因素。

可能大部分人认为Python在可读性上更胜一筹，但我有着不同的观点，且听我说完。先看一下[Python sure has fantastic readability(Python确实有很高的可读性)](https://dzone.com/refcardz/core-python)，但在我看来，他们有点说得过头了。在Python中，对一件事的表达可能有十多中不同的写法，这同常使得在代码量很大或者同事项目的人很多的情况下容易导致混淆。

另一方面，Go在编程的时候有着严格的规则约束。它不允许不使用的包被import，或者不被使用的变量被声明。这便意味着在大型团队中有更明确的方法对代码有着更好的理解，但有谁会去关心多功能性呢，尤其是在进行核心开发的时候。Golang的语法对于初学者来说很不友好，但是并不像C或C++一般情。所以对于代码的可读性，我更倾向于Golang.

大家都看到了，在我看来，作为一门编程语言，Glang在很多方面都胜过Python.当然Golang不像Python一样有名，Python这几年已经在充斥在互联网中，但是Go在这方面也是迎头赶上。如果你对我的看法有异议，可在下方评论区进行评论。我希望我可以帮助你确定哪门编程语言对你的项目更好。请继续关注更多Golang的博客。

---

via:[Golang vs. Python: Which One to Choose?](https://dzone.com/articles/golang-vs-python-which-one-to-choose?fromrel=true)

作者：[Aryya Paul](https://dzone.com/users/3510559/aryya-paul.html)
译者：[BeGemini](https://github.com/BeGemini)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出