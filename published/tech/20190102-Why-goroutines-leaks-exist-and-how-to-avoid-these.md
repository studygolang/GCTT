首发于：https://studygolang.com/articles/17654

# 为什么会存在 Goroutines 泄漏，如何避免？

在这个时代，软件程序和应用程序应该能够快速顺畅地运行。能够并行运行代码是当今满足这种需求的能力之一。但要注意泄漏的危险！我们将与来自[Ardan Labs](https://www.ardanlabs.com/) 的 Jacob Walker 讨论这个问题，他专门为 Golang 工程师撰写了这篇[博文](https://studygolang.com/articles/17364)。

在开始阅读我们关于调试 Goroutines 泄漏的讨论之前，首先简要介绍几个原理，这些原理可以让您更全面地了解这个概念所要解决的挑战。

## 并发编程

并发编程是一种并行编码的方法，可以同时运行连续线程的集合。通过这种方式，软件程序可以更快地计算，并且因此能够更好地执行。当今多核处理器的功能赋予了这种并发编程能力。

![img01](https://raw.githubusercontent.com/studygolang/gctt-images/master/Why-goroutines-leaks-exist-and-how-to-avoid-these/0_4qAuYdD50SJv9ry4.jpg)

## Goroutines:

传统的线程方法基于使用共享内存的线程之间的通信。Go 不是专门使用锁来调解对共享数据的访问，而是促进使用通道在 Goroutine 之间移动对数据的引用。这样，在给定时间段内只有一个 Goroutine 可以访问数据。Golang 满足了使用这些 Goroutine 进行并发编程的需求，这些 Goroutine 基本上是由 Go 在运行时控制的轻量级线程。

![img02](https://raw.githubusercontent.com/studygolang/gctt-images/master/Why-goroutines-leaks-exist-and-how-to-avoid-these/0_zF1_QhpVAM4mIGnw.jpg)

## 泄露

但是要小心！ Goroutines 可以缓慢但可靠地储存一段时间，因此浪费你的内存等资源，你甚至都不会注意到它。
因此，了解泄漏的危险和 ( 或 ) 尽早调试它们非常重要。这是我们最近在与 Jacob Walker 的访谈中讨论到的一个主题，
在[Gophers Community at Slack](https://invite.slack.golangbridge.org/) 的 #review-interview 频道（40000+ 成员）。
请阅读下面的采访。

![img03](https://raw.githubusercontent.com/studygolang/gctt-images/master/Why-goroutines-leaks-exist-and-how-to-avoid-these/0__Wf-KF9gpdVMIpBA.png)

> 总之，我作为开发人员已经工作了大约 10 年之久，并且在过去的 4 年中几乎全身心投入于 Go 中。我开始使用 1.3 版本。
>
> 我在 Ardan Labs 担任社区工程师。我的主要活动是教授 Go Fundamentals 和 Ultimate Go 等课程。在授课期间，我写了大量博客文章，开发新内容以及帮助社区开发者。

**Sebastiaan（采访者）[4:16 PM]：**

> 你过去 4 年愿意接受 Go 的原因是什么？你为什么喜欢这种语言？

**Jacob Walker [4:17 PM]：**

> 哈哈是的，我非常喜欢它：slightly_smiling_face：我接受的正规教育是商业（MBA）。因此当我开始接触开发时，我主要使用在线资源“自学成才”。我开始用 HTML、JS 和 PHP 来做 Web 应用程序开发。几年来，这对我很有帮助，我能够做出一些很酷的东西并解决一些有趣的问题。经过大约 5 年的研究，并涉足其他语言，如 Ruby 和 Python，我想更深入，更接近机器。我查看了 C 和 C ++ 并在那里做了一段时间的实验，但 Go 令我感到振奋。关于 Go 的思维模式和哲学的一些东西真的与我正在寻找的东西以及我想要编写代码的方式相匹配。至于我喜欢这种语言的原因？有很多：slightly_smiling_face： - 简单的语言，令人备受鼓舞的代码和模式。- 一致性。规则在整个语言中的应用方式非常一致。- 工具非常出色。- 一般认为通常有“一种正确的方法”来做大多数事情 , 所以我不必花费大量时间来猜测我应该用可以解决问题的十几种方法中的哪一种 , 我通常倾向于第一种方式。

**Sebastiaan ( 采访者 ) [4:25 PM]:**

> 因为你（和 Ardan Labs）是关于 Go 的重要贡献者 / 用户，如果让你回顾这一整年，你有什么感想以及用哪个词概括你所取得的成就？

**Jacob Walker [4:26 PM]：**

> 我们已经看到了用户群体的爆炸性增长。公司培训需求不断增加。对于这门语言来说这是非常激动人心的时刻。

**Sebastiaan ( 采访者 ) [4:27 PM]:**

> 好的，你能否对这个“激动人心”的时刻再详细解释一下？它的什么地方特别令人兴奋？

**Jacob Walker [4:30 PM]：**

> 当然。对我来说，看到所有这些新的 Gopher 加入社区是令人兴奋的！我喜欢这种语言，与别人分享你喜欢的东西是非常美妙的。越来越多的人开始明白经验丰富的牧羊人长期以来所喜爱的语言及其生态系统。现在这些都在以积极的势头发展。

**Sebastiaan ( 采访者 ) [4:32 PM]:**

> 您个人希望在 2019 年发布什么样的新功能？

**Jacob Walker [4:33 PM]：**
> 嗯，这是个好问题。最近有很多关于 “Go 2”和“泛型”的设计草案、简化错误处理、错误上下文等的讨论。这类讨论很有意思。
>
> 我并不担心这些新功能的具体细节因为我知道 :
> 1. 它们还有很长的路要走。
> 2. 它们与现有设计相比，它们在发布时看起来可能会有很大差异，在不久的将来 , 我可能会非常高兴看到模块问题得到巩固。

**Sebastiaan ( 采访者 ) [4:38 PM]:**

> 所以你写了关于 Goroutine 泄露的文章。你为什么会选择这个话题？是因为您多次看到这个问题没有被新手级 Gopher 所解决吗？

**Jacob Walker [4:41 PM]：**
> 是这样的。早些时候，当你问我对这种语言的喜爱时，我没有列出有关并发的原因。我绝对喜欢编写并发代码，Go 有着我见过的最好的并发方法之一。
>
 问题是许多新的 Gophers 在没有必要时使用了它，并且使用的并发带有一些陷阱。在我的训练课中，当我们达到并发时，我会提醒我的学生们要注意这些陷阱以及如何避免它们。这些信息对社区中的每个人都很有用，所以我想把它转换成一系列的博客文章。你之前链接的 Goroutine 泄露的帖子是该系列中的第一篇。

**Sebastiaan ( 采访者 ) [4:44 PM]:**

> 您觉得哪些陷阱是开发人员最容易忽视的？

**Jacob Walker [4:45 PM]**

> 有 Goroutine 的泄漏，很容易被意外创建。我猜最常见的是 Data Races。它还存在不完整工作的风险。你可以启动 Goroutine 来执行某些操作，但是当程序终止时它可能无法完成，然后它会被切断。所以很难说哪种是最常见或最危险的。他们都很糟糕：slightly_smiling_face：

**Sebastiaan ( 采访者 ) [4:49 PM]:**

> 对于这些陷阱，您有什么独特的见解 , 可以在 Go 中安全地使用并发？

**Jacob Walker [4:51 PM]**

> 当然。你可以通过记住类似谚语的知识来解决 Goroutine 泄漏和不完整的工作
> 永远不要在不知道如何停止的情况下启动 Goroutine。据我所知，这句话来自 Dave Cheney 的博客文章 https://dave.cheney.net/2016/12/22/never-start-a-goroutine-without-knowing-how-it-will-stop。
> 对于数据竞争，社区有谚语：
>
> 不要通过共享内存进行通信，通过通信共享内存。它有点简洁和富有诗意，所以一开始可能没有意义，但一旦你理解了这句谚语就会有助于你去理解记忆。关键是要避免在 Goroutines 上共享相同的变量 / 内存。相反，每个变量都应由一个 Goroutine 维护和管理。

**Sebastiaan ( 采访者 ) [4:56 PM]:**

> 所以，解决这些潜在问题是基于这些基本原则。如果您在团队中工作，是否有某些最佳实践以确保您始终遵循这些原则？

**Jacob Walker [4:58 PM]：**
> 我真正可以说的是做到代码审查。没有人会将代码发布到至少没有被其他人彻底审查的生产环境中。确保团队中的每个人都知道这些原则，并且知道要这样做。
> 对于第一句谚语，每次看到 `go` 关键字 , 它会让你疑惑“这个 Goroutine 什么时候会终止？”。

**Sebastiaan ( 采访者 ) [5:01 PM]:**

> 关于安全性，泄漏会导致什么样的问题？

**Jacob Walker [5:03 PM]:**

> 这块想到的唯一的事情就是为拒绝服务攻击创建一个向量。如果攻击者知道以特定方式发出请求会导致 Goroutine 泄漏，那么他们就会这么做，导致服务器耗尽资源并崩溃。

**Sebastiaan ( 采访者 ) [5:06 PM]:**

> 是否有额外的资源可用于安全性审核，使您能够处理 Goroutine 泄漏？

**Jacob Walker [5:09 PM]：**

> 当然有的。在出现问题之前进行代码审查是最好的，但是一旦代码在生产中或甚至在分段或 QA 环境中，您可以使用诸如 `pprof` 之类的监视工具来计算活动 Goroutines 的数量。如果这个数字总是增加而且从不减少 , 那么你就应该想到可能某个地方发生了泄漏。

**Sebastiaan ( 采访者 ) [5:10 PM]:**

> 很棒的建议。让我们来看一下这个链接：https：//golang.org/pkg/net/http/pprof/。是否有其他关于我们讨论的主题 Ardan labs 的文章？

**Jacob Walker [5:13 PM]：**

> Bill 最近完成了关于 Goroutine 调度程序机制的系列文章。这不是专门针对泄漏的，但它为理解调度程序在开始使用并发时的行为提供了很好的材料。https://www.ardanlabs.com/blog/2018/12/scheduling-in-go-part3.html

**Sebastiaan ( 采访者 ) [5:15 PM]:**

> 您已经说过您的文章只是一个系列的开始 , 那么我们还可以期待什么讨论？

**Jacob Walker [5:15 PM]：**
> 我已经发布了第二篇文章，展示了 Goroutine 泄漏的更多例子 https://www.ardanlabs.com/blog/2018/12/goroutine-leaks-the-abandoned-receivers.html
> Goroutine 泄漏是 Go 程序中内存泄漏的常见原因。在我之前的文章中，我介绍了 Goroutine 泄漏，并提供了许多 Go 开发人员常见的错误例子。继续这项工作，这篇文章提出了另一个关于 Goroutines 如何被泄露的情景。我可能会继续使用 Goroutine 泄露，或者我可能继续使用其他并发陷阱。数据竞争、不完整的工作 , 最后是不必要的复杂性。
> 每个主题至少需要一个或多个帖子来讨论。

**Sebastiaan ( 采访者 ) [5:18 PM]:**

> 从所有这些主题或挑战中，您个人处理过哪些特别棘手的问题？

**Jacob Walker [5:21 PM]：**

> 我遇到的其中的每一个，它们都有不同的难度，所以很难说哪个是特别棘手的。Goroutine 泄露和数据竞争特别棘手，因为你可能甚至不知道它们正在发生。

**Sebastiaan ( 采访者 ) [5:23 PM]:**

> 好的，谢谢您愿意接收采访 @jcbwlkr。您的下一篇关于并发的帖子约在何时发布？

**Jacob Walker [5:24 PM]:**

> 哦，现在我有一个截止日期！：slightly_smiling_face：我会在接下来的几个星期内做一段时间，但有时可能需要过几周才能查看这些帖子，以确保它们达到我们的标准。可能到了 1 月底？谢谢你对我的采访，@ Sebastiaan ！

![img04](https://raw.githubusercontent.com/studygolang/gctt-images/master/Why-goroutines-leaks-exist-and-how-to-avoid-these/0_DdguCwEvrkc6l1z_.png)

![img05](https://raw.githubusercontent.com/studygolang/gctt-images/master/Why-goroutines-leaks-exist-and-how-to-avoid-these/0_s5amuhubZvRktiui.png)

最初发表于[Jexia](http://blog.jexia.com/why-goroutines-leaks-exist-and-how-to-avoid-these/)。

---

via: https://medium.com/jexia/why-goroutines-leaks-exist-and-how-to-avoid-these-dfc572bdad08

作者：[Jexia’s Editorial Team](https://medium.com/@content_62255)
译者：[amesy](https://github.com/amesy)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
