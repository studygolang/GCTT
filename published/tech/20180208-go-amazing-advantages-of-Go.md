---
layout: post
title:  "[GCTT] Here are some amazing advantages of Go that you don't hear much about"
date:   2018-02-08
comments: true
categories: GCTT
tags: Go framework
description:
published: true
---

# 你所不知道的 Go 语言的一些令人惊叹的优点

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-advantage/1.png)
> 插图来自 https://github.com/ashleymcnamara/gophers

在这篇文章中，我将会讨论为什么你应该尝试下 Go 语言，并且应该从哪里开始下手。

Golang 是一种编程语言，在过去的几年中你可能听说过很多。尽管是在 2009 年创建的，但是近年来才开始流行。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-advantage/2.png)
> 上图是根据 Google Trends 得出的 Golang 的流行程度

本文不是关于你通常看到的 Go 的主要卖点。

相反，我想向你介绍一些相当小但仍然很重要的功能，你只有在决定尝试 Go 之后才能了解到这些功能。

这些令人惊叹的特性没有浮于表面，它们可以为你节省大量的工作。它们还可以使软件开发更加愉快。

如果 Go 对你来说是新事物，别担心。本文不需要任何 Go 语言经验。如果你想了解更多，我已经在文章底部添加了一些额外的链接。

我们将会涉及以下主题：

- GoDoc
- 静态代码分析
- 内置的测试和性能分析框架
- 竞争条件检测
- 学习曲线
- 反射
- 固定的代码风格
- 文化

请注意，列表不遵循任何特定的顺序。完全随意排序。

## GoDoc

Go 非常重视代码中的文档。在 Go 中，文档也很容易添加。

[GoDoc](https://godoc.org/) 是一个静态代码分析工具，可以直接从你的代码中创建出漂亮的文档页面。关于 GoDoc 的一个值得注意的事情是，它不使用任何额外的语言，就像 JavaDoc，PHPDoc 或 JSDoc 在代码中的注释结构一样。仅仅只使用英语。

它尽可能多的使用从代码中获取的信息，来构建文档的轮廓，结构化和格式化文档。它具有所有的花里胡哨的东西，比如交叉引用，代码示例和直接链接到版本控制系统库。

所有你能做的就是添加一个好的 `// MyFunc transforms Foo into Bar` 注释，这也将会在文档中体现出来。你甚至可以添加[代码示例](https://blog.golang.org/examples)，它可以通过 Web 界面或在本地**实际运行**。

GoDoc 是整个社区使用的唯一的 Go 文档引擎。这意味着用 Go 编写的每个库或应用程序都具有相同的文档格式。从长远来看，它帮你节省了大量浏览这些文档的时间。

举个例子，这是我最近实现的示例项目的 GoDoc 页面：[pullkee — GoDoc](https://godoc.org/github.com/kirillrogovoy/pullkee)。

## 静态代码分析

Go 重度依赖于静态代码分析。例如，包括用于文档的 [godoc](https://godoc.org/)，用于代码格式化的 [gofmt](https://golang.org/cmd/gofmt/)，用于代码风格检查的 [golint](https://github.com/golang/lint)，以及许多其他的例子。

有这么多的工具，甚至有一个叫 [gometalinter](https://github.com/alecthomas/gometalinter#supported-linters) 的项目，能够把所有的工具打包组合成一个单一的工具。

这些工具通常作为独立的命令行应用程序来实现，并可以轻松地集成到任何编码环境。

静态代码分析实际上并不是现代编程中的新东西，但是 Go 把它用到了极致。我不能高估它为我节省了多少时间。另外，它会给你一种安全的感觉，好像有人在你的背后替你遮挡风雨。

创建自己的分析工具非常容易，因为 Go 有专门的内置软件包可以用来解析和处理 Go 源代码。

你可以从这个演讲中了解更多：[GothamGo Kickoff Meetup: Go Static Analysis Tools by Alan Donovan.](https://vimeo.com/114736889)

## 内置的测试和性能分析框架

你有没有试过为一个从头开始的 Javascript 项目选择一个测试框架？如果是这样，你可能会明白，经历这样一个分析瘫痪的斗争。你也许已经意识到你并没有使用你所选择框架的 80%。

一旦你需要做一些可靠的分析，这个问题就会重复出现。

Go 提供了一个内置的测试工具，旨在简化和高效。它为你提供了最简单可用的 API，并做出了最小的假设。你可以将其用于不同类型的测试，分析，甚至提供可执行的代码示例。

它开箱即用，能够生成了 CI 友好的输出，使用方法通常和运行 `go test` 一样简单。当然，它也支持高级功能，如并行运行测试，标记跳过，等等。

## 竞争条件检测

你可能已经知道 Goroutines，它在 Go 中用于实现并发代码执行。如果你还不了解它，[这里](https://gobyexample.com/goroutines)有一个非常简短的解释。

在复杂的应用程序中进行并发编程并不容易，不管具体的技术如何，部分原因在于竞争条件的可能性。

简而言之，当多个并发操作以不可预知的顺序完成时，竞争条件就会发生。这可能会导致大量的错误，特别难以追查。有没有花了一天的时间去调试一个只能执行大约 80% 情况的集成测试呢？这可能是一个竞争条件。

所有这一切表明，并发编程在 Go 中非常受重视，幸运的是，我们有相当强大的工具来捕捉这些竞争条件。它被完全集成到 Go 的工具链中。

你可以在这里阅读更多关于它的信息，并学习如何使用它：[Introducing the Go Race Detector — The Go Blog.](https://blog.golang.org/race-detector)

## 学习曲线

你可以在一个晚上学习完 Go 语言的所有功能。我是认真的。当然，还有标准库，以及不同的，更具体的领域的最佳实践。但是，两个小时完全足够让你自信地写出一个简单的 HTTP 服务器或命令行应用程序。

这个项目有[非常棒的文档](https://golang.org/doc/)，大部分的高级主题已经被他们的博客所覆盖：[The Go Programming Language Blog](https://blog.golang.org/)。

Go 比 Java（和它的家族成员），Javascript，Ruby，Python 甚至 PHP 更容易在你的团队普及。Go 的开发环境很容易设置，你的团队只需要做很小的投资，就能完成你的第一个产品代码。

## 反射

代码反射本质上是一种能力，可以潜藏在屏幕下方，访问有关语言结构的各种元信息，如变量或函数。

鉴于 Go 是一种静态类型的语言，当涉及更松散的类型抽象编程时，它会受到各种限制。特别是与 Javascript 或 Python 等语言相比。

而且，Go [没有实现一个被称为泛型的概念](https://golang.org/doc/faq#generics)，这使得以抽象的方式处理多种类型更具挑战性。然而，由于泛型带来的复杂性，许多人认为这对语言实际上是有益的。我完全同意。

根据 Go 的设计哲学理念（这本身是一个独立的话题），你应该尽量不要过度设计你的解决方案。这也适用于动态类型编程。尽可能地坚持使用静态类型，当你明确知道你正在处理什么类型时使用接口。在 Go 中，接口非常强大，无处不在。

但是，仍然有些情况下你不可能知道你所面对的是哪种数据。JSON 就是一个很好的例子。你可以在应用程序中来回转换所有类型的数据。字符串，缓冲区，各种数字，嵌套结构等等。

为了解决这个问题，你需要一个工具来检查运行时的所有数据，根据数据类型和结构的不同而采取不同的行为。反射能够帮助你！Go 有一个一等公民的反射包，它使你的代码可以是动态的，就像 JavaScript 这样的语言。

一个重要的警告是要知道你使用它付出了什么样的代价 -- 只有在没有简单的方法时才使用它。

你可以在这里读到更多关于它的内容：[The Laws of Reflection — The Go Blog.](https://blog.golang.org/laws-of-reflection)

你也可以在这里阅读 JSON 包源代码中的一些真实代码：[src/encoding/json/encode.go — Source Code](https://golang.org/src/encoding/json/encode.go)

## 固定的代码风格

顺便说一句，有这样一个词吗？

在 Javascript 的世界里，我面临的最艰巨的事情之一是决定我需要使用哪些约定和工具。我应该如何格式化我的代码？我应该使用什么测试库？我应该怎么去设计结构？我应该依靠什么样的编程模式和方法？

有时这些东西基本上让我卡住了。我不得不做这些事情，而无法把时间花在写代码、满足用户需求上。

首先，我应该注意到，我完全知道这些公约应该来自哪里。这总是来自你和你的团队。无论如何，即使是一群经验丰富的 Javascript 开发人员也可以很容易地发现自己使用完全不同的工具和约定，来达到相同的结果。

这使得在整个团队中分析定位问题很困难，也使得每个人难以互相合作。

但是，Go 是不同的。你只有一份每个人都必须遵循的代码风格指南。你只有一个内置于基本工具链中的测试框架。关于如何构建和维护代码，你有很多强烈的意见。如何选择名称。遵循什么样的结构化模式。如何更好地执行并发。

虽然这可能看起来过于严格，但这能够为你和你的团队节省了大量的时间。当你编码时，有些限制是一件很好的事情。在构建新的代码时，它给了你一个更直接的方法，并且更容易理解现有的代码。

因此，大部分 Go 项目看起来都非常相似。

## 文化

人们都说，每当你学习一种新的口语时，你也会受到说这种语言的人的文化的部分熏陶。因此，你学习的语言越多，你可能受到的更多的个人改变。

这与编程语言是一样的。无论将来使用何种新的编程语言，它总是会给你一个关于编程的新观点，或者一些特定的技术。

无论是函数式编程，模式匹配还是原型继承。一旦你掌握了这些知识，就可以随身携带这些方法，从而拓宽了你作为一个软件开发人员解决问题的工具。通常来说它也改变了你看待高质量编程的方式。

Go 是一个很好的投资机会。Go 文化的主要支柱是保持简单实用的代码，而不会产生多余的抽象，并且非常重视代码的可维护性。能够把大量时间用在实现业务代码上，而不是用来修改工具和配置环境，这也是文化的一部分。或者在不同的变体之间进行选择。

Go 也可以总结为“应该只有一个方法来完成一件事情”。

一个小方面的说明。当你需要构建一个相对复杂的抽象时，用 Go 通常不是那么好实现的。对于这点，这应该是 Go 语言简单性的一种代价。

如果你真的需要编写大量的具有复杂关系的抽象代码，最好使用像 Java 或 Python 这样的语言。但是，即使真的有需要，这种情况也是非常罕见的。

始终使用最好的工具来完成工作！

## 总结

你以前可能听说过 Go。或者也许它一直不在你的使用范围之内。无论哪种方式，Go 可以是你或你的团队在开始新项目或改进现有项目时一个非常体面的选择。

这并不是关于 Go 的所有令人惊叹的优点的完整列表。**仅仅只是被低估的部分。**

请尝试一下 Go 语言，[A Tour of Go](https://tour.golang.org/) 是一个很好的开始学习的地方。

如果你想了解更多关于 Go 的好处，你可以看看这些链接：

- [Why should you learn Go? — Keval Patel — Medium](https://medium.com/@kevalpatel2106/why-should-you-learn-go-f607681fad65)
- [Farewell Node.js — TJ Holowaychuk — Medium](https://medium.com/@tjholowaychuk/farewell-node-js-4ba9e7f3e52b)

在评论中分享你的学习结果！

即使你没有在专门寻找一种新的语言，花一两个小时的时间了解下 Go 也是值得的。也许在未来可能会变得对你非常有用。

持续不断的寻找最适合你的开发工具！

---

如果你喜欢这篇文章，请考虑关注我的更多内容，并点击这些文本下面的那些有趣的绿色小手来分享。👏👏👏

---

via: https://medium.freecodecamp.org/here-are-some-amazing-advantages-of-go-that-you-dont-hear-much-about-1af99de3b23a

作者：[Kirill Rogovoy](https://medium.freecodecamp.org/@kirillrogovoy)
译者：[MDGSF](https://github.com/MDGSF)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
