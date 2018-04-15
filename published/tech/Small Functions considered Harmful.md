已发布：https://studygolang.com/articles/12805

# 小函数可能是有害的（Small Functions considered Harmful）

在这篇博文中，我的目标是：

- 揭示小函数的一些似是而非的优点
- 解释为什么我个人认为有点不像建议说的那么好
- 解释为什么小函数有时是适得其反
- 解释一下我认为小函数在 mock 中真正有用的地方

通常，编程建议总是说使用更优雅和有益的小函数。《Code Clean》被普遍认为是一本编程圣经，它有一章专门讲述函数，文章的开始就是介绍一个非常长，令人头疼的函数。该书认为该函数的最大问题是长度过长，并指出：

它（函数）不仅长度太长，而且有多处重复的代码，奇怪的字符串，许多奇怪和不明了的数据类型和 API。三分钟的学习后，你能了解函数的功能吗？也许不能。那里有太多的抽象层次。奇怪的字符串，奇怪的函数调用混合在双重嵌套，并由标志位控制的 if 语句中。

本章简单地思考了什么样的特性会使代码 “更容易阅读和理解” 和 “允许任何一个读者都能直观地认识他们遇到的程序”，然后才说为了达到这个目的，必须将函数设置得更小一些。

函数的第一条原则是必须小。函数的第二条原则是它必须更小。

函数应该很小的观点几乎被认为是权威看法，不容质疑。在代码审查，twitter 上，会议上，关于编程的书籍和播客中，关于代码重构的最佳实践的文章中，等等。这个想法几天前以这种推文的形式再次进入我的时间线：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-1.jpg)

Fowler 在他的推文中，链接了他关于函数长度的文章，并继续指出：

如果你不得不花费精力查看这一段代码来确定它在做什么，那么你应该把它提取到一个函数中，并以它的功能命名该函数。

一旦我接受了这个原则，我就养成了写一些非常小的函数的习惯 - 通常只有几行 [2](https://martinfowler.com/bliki/FunctionLength.html#footnote-nested)。任何超过半打行数的函数都会让我觉得不舒服，对我而言，只有一行代码的函数也并不罕见 [3](https://martinfowler.com/bliki/FunctionLength.html#footnote-mine)。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-2.jpg)

有些人很迷恋小函数，所以对任何可能看起来很复杂的逻辑抽象成一个单独的函数的想法向来都是推崇备至。

我一直在研究人们继承过来的代码库，他们将这种观点内化到完全扭曲的地步，以至于最终走向了不可挽回的地步，完全违背了这个观点的最初意愿。在这篇文章中，我希望解释一下为什么一些经常被吹捧的好处并不总是按照人们希望的方式发展，有的时候，一些观点的应用会变得适得其反。

## 小函数的好处（Supposed benefits of smaller functions）

通常会列出一些理由来证明小函数背后的优点。

### 只做一件事（Do one thing）

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-3.jpg)

这个想法很简单 - 一个函数应该只做一件事，并做好。从表面上看，这似乎是一个非常好的想法，跟 Unix 哲学不谋而合。

当这个 ”一件事“ 需要被定义的时候，描述就变得模糊了。”一件事“ 可以是从简单的返回语句到条件表达式，通过网络调用的数学计算（等等）。正常情况下，许多时候，这个 ”一件事“ 意味着对某些（通常是业务）逻辑的单级抽象。

例如，在 Web 应用程序中，像 “创建用户” 这样的 CURD 操作可能是 “一件事”。通常，创建用户至少需要在数据库中创建记录（并处理任何伴随的可能错误）。此外，用户注册后可能还需要向他们发送欢迎电子邮件。另外，人们也可能希望可以自定义一个事件，像 kafka 这样的消息中间件，可将此事件发送给其他各个系统。

因此，“单一抽象层次” 不仅仅是一个层次。我所看到的是，那些完全理解函数应该做 “一件事” 的想法的程序员往往很难抵制将递归应用于他们编写的每个函数和方法中。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-4.jpg)

因此，我们现在不再是为了可以被理解（和测试）而抽象成一个合理的单元，而是将更小的单元划分出来，以描述 ”一件事“ 的每个组成部分，直到它完全模块化，完全 DRY（Don't repeat yourself）。

## DRY 的谬论

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-5.jpg)

DRY 和尽可能小的函数的倾向并不一定是同一件事，但我已经看到后者很多时候会让目标变成前者。DRY 在我看来已经是一个很好的指导原则，但实用和理性在教条地坚持下牺牲了，特别是那些信服 Rails 的程序员。

Python 的核心开发人员 Raymond Hettinger 发表了一篇名为 Beyond PEP8 的精彩演讲：[那是美妙而易懂的最佳实践](https://www.youtube.com/watch?v=wf-BqAjZb8M)。这是一个必须关注的话题，不仅适用于 Python 程序员，也适用于任何对编程感兴趣或以开发程序为生的人，因为它非常清楚地解释了教条式遵守 PEP8 的谬误，这是真正的 Python 风格指南，它介绍了很多底层实现。在 PEP8 上的谈话焦点并没有比可以应用的见解重要，（而且）其中许多情况是语言描述不了的。

即使你没有看完整个演讲，你也应该看下这个讲话的开头一分钟，这个演讲与 DRY 的警鸣做了令人惊讶的类比。程序员坚持要尽可能多地精简代码，会让他们只关注局部，忽略掉整体。

我对于 DRY 的主要问题在于它强制抽象成为抽象 - 嵌套和太早的抽象。由于不可能完美地抽象，所以我们只能尽可能地做到足够好的抽象。“足够好” 的定义很难，并且取决于很多因素。

在下图中，“抽象” 一词可以与 “函数” 互换使用。例如，假设我们要设计抽象层 A，我们可能需要考虑以下几点：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-function/1_Mh46Hv7CEkfVc_SKlA0d1w.png)

- 支撑抽象概念 A 的假设性质以及它们可能持有的水平的可能性（以及可能持续多长时间）
- 抽象层 A（抽象层 X 和抽象层 Y）以及建立在抽象层 A（抽象层 Z）之上的任何抽象层的抽象层在其实现和设计中倾向于保持一致性，灵活性，可扩展性和正确性。
- 未来抽象（抽象层 M）的需求和期望可能建立在抽象 A 之上，以及可能需要在 A（抽象层 N）之下支持的任何抽象

我们开发的抽象层 A 不可避免地会在未来不断被重新评估，并且很可能会部分甚至完全失效。一个能够支撑我们需要的，不可避免的修改的最重要的特征就是设计我们的抽象，使之变得更灵活。

尽可能最大限度地优化代码意味着，将来需要适应修改时，（这将）剥夺了我们自己的灵活性。我们优化时也要做到让自己有足够的余地来适应不可避免的变化，迟早会有这样的要求，而不是马上为了完美的契合而进行优化。

最好的抽象是优化得足够好，但不完美的抽象。这是一个函数，而不是一个错误。理解抽象的这种非常突出的性质是设计好程序的关键。

Alex Martelli 是鸭子理论和蟒蛇派的名人，他著名的演讲 “抽象塔” 中的幻灯片非常值得一读。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-function/1_fvfBJ21qOdt3XGAFHa0oOg.png)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-function/1_i5vRl8dA8docZutvy-LgYA.png)

Rubyist Sandi Metz 有一场名为 All The Little Things 的著名演讲，她认为 “重复比错误的抽象代价更低”，因此 “倾向于重复的抽象”。

在我看来，抽象概念不可能是完全 “正确的” 或者 “错误的”，因为划分 “正确” 与 “错误” 的界限本来就很模糊。实际上，我们精心设计的 “完美” 抽象，只是一个业务的要求或者被委托的一个错误的缺陷报告。

我认为这有助于将抽象视为图谱，如我们在本文前面看到的图表一样。该图谱的一端优化精度，我们代码的每个方面，最后都要求要精确。这当然有其好处，但是因为努力寻求完美的对齐方式，所以并不适合好的抽象。该图谱的另外一端优化，带来了不精确性和缺少边界。虽然这确实允许最大的灵活性，但我发现这种极端的倾向将导致其他的缺点。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-6.jpg)

跟其他大多数事情一样，“理想模型” 处于这两者之间。没有一种娱乐能取悦所有人。这个 “理想模型” 也取决于许多因素 - 工程和社会关系 - 并且，良好的工程是能够确定这个 “理想模型” 在不同环境中所处的位置，并能不断地重新评估并校准这个模型。

### 给抽象命名（The name of the game）

说到抽象，一旦确定了抽象什么以及如何抽象，就需要给它一个名称。

给事物命名向来都很难。

这种方式（给抽象命名）普遍认为是编程过程中，使代码能活得更长的有效办法，更具描述性的名称是一件好事，甚至有人主张用带有注释的名称代替代码中的注释。他们的想法是，一个名称越具描述性，意味着封装得越好。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-7.jpg)

这个观点在 Java 的世界里普遍存在，（ Java 程序中）冗长的名称非常常见，但我从来没有发现这些冗长的名称使代码更加容易阅读。例如，可能 4-5 行的代码中就隐藏一个名字非常长的函数。当我正在阅读代码时，突然一个非常长的单词出现，会让我停下来，因为我得试图处理这个函数名称中的所有不同的音节，尝试将它融入到我已创建的心智模型中，然后决定，是否通过跳转到它定义的地方，来看它的具体实现。

然而，“小函数” 的问题在于追寻小函数的过程中导致了更多的小函数，所有这些函数都倾向于在记录自己和避免讨论的过程中给出了非常冗长的名称。

结果，处理描述详细的函数（和变量）的名称带来了认知的开销，以及将它们映射到我迄今为止构建的心智模型中，以确定哪些函数需要深入探究，哪些函数可以剔除，并将这些拼图拼在一起以揭开程序的面纱，但处理冗长的函数（和变量）名使得这个过程变得更加地困难。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-8.jpg)

就我个人而言，与查看自定义的变量或者函数名相比，从视觉角度来说，我发现编程语言提供的关键字，构造和习惯用法更加容易接受。例如，当我阅读 if-else 模块时，我很少需要花费精力去处理关键字 if 或者 elseif，只需要花时间理解程序的逻辑流程。

一个 VeryVeryLongFuncNameAndArgList 名称会中断我的推理思路。当被调用的函数实际上是一个可以轻松内联的单线程时尤其如此。上下文切换很昂贵，不管是 CPU 上下文切换还是程序员在阅读代码时不得不在思想上的上下文切换。

过度强调小函数的另外一个问题是，尤其是那些描述性很强但名字不直观的函数，在代码库中更难搜索到。相比之下，一个名为 createUser 的函数很容易，且直观地用于 grep，比如 renderPageWithSetupsAndTeardowns（在《Clean Code》中是作为明星例子，这个名字不是最容易记住的名称，也不是最容易搜索到的名字）。许多编辑器也对代码库进行了模糊搜索，因此具有相似前缀的函数也更可能造成搜索时出现多余的结果，这不是我们想要的。

### 本地的丢失（Loss of Locality）

（注：这里指可以在本函数，本文件，本包中实现的代码，却为了小函数移到了其他的函数，文件，或包中）

当我们不必跳过文件或包来查找函数的定义时，小函数的效果最好。“Clean Code” 一书为此提出了一个名为 “The Stepdown Rule” 的原则。

** 我们希望代码能像自上而下的叙述一样容易阅读。我们希望每个函数都被下一级抽象层次的人所遵循，以便我们阅读程序，读取函数列表时，可以一次下降一个抽象层次。我称之为 “The Stepdown Rule"。**

这个观点理论上可行，但在实际的实践中，却很少能发挥作用。相反的，我看到的大多是在代码中增加更多的函数，减少了本地代码。

让我们以三个函数 A，B 和 C 的假设开始，一个调用另外一个。我们的初始抽象印证了某些假设、要求和注意事项，所有这些都是我们在最初设计时仔细研究和论证过的。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-function/1_EGR-6c3hu_6joqdNYHfyLQ.png)

很快，假设我们有一个新的需求或一个附加功能的情况下，我们需要迎合没有预见的或一个新的约束。我们需要修改函数 A，因为它封装的 “一个整体” 已经不再有效（可能从一个开始就无效，现在我们需要修改它，使它有效）。按照我们在 《Clean Code》中所学到的，我们处理这些问题的最好办法是，创建更多的函数，隐藏掉各种杂七杂八的新需求。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-function/1_N77X5FQhiscmnNUKlR8_Cw.png)

我们按照我们的想法修改后，过个几周，如果我们的需求又修改了，我们可能需要创建更多的函数去封装所有要求增加的修改。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-function/1_aQn2iFAvxzrJr_89aTghtQ.png)

再来几次，我们就真正的看到了 Sandi Metz 在她的博文 [《The Wrong Abstraction》](https://www.sandimetz.com/blog/2016/1/20/the-wrong-abstraction) 中描述的问题。这篇博文说：

已经存在的代码具有强大的影响力。它的存在表明它是正确和有效的。我们知道代码代表了付出的努力，我们非常积极地维护这个努力的价值。不幸的是，可悲的事实是，代码越复杂，越难以理解，即设计它时投入越大，我们就越觉得要保留它（“沉没成本谬论”）。

如果还是同一团队成员继续维护它，我相信这是对的，但当新的程序员（或经理）获得代码库的所有权时，我会看到相反的结果。以良好意图开始的代码，现在变成了意大利面条的代码，代码不再简洁，成了地狱般的代码，现在 “重构” 或者有时甚至重写代码的冲动更加诱人。

现在，人们可能会争辩说，从某种程度上说，这是不可避免的。他们是对的。我们很少讨论编写将会退役的代码是多么重要。过去我写过关于使代码在操作上易于退役的重要性，在涉及代码库本身时更是如此。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-9.jpg)

通常情况下，程序员只会在代码确定被删除，或者不再使用时，将代码视为 “已死亡”。如果我们开始（以代码将 “死亡”）思考我们的编写的代码，那么每增加一个新的 git commit，我认为我们可能会更加积极地编写易于修改的代码。在思考如何抽象时，认识到我们正在构建的代码可能距离死亡（正在被修改）只有几个小时的事实对于我们很有帮助。因此，为了便于修改代码而进行的优化往往比试图构建 《Clean Code》中提到的自顶向下的设计更好。

## 类污染

在支持面向对象的编程里，小函数带来了更大或者更多的类。在像 Go 一样的编程语言里，我看到这种趋势导致更大的接口（结合接口实现的双重打击）或者大量的小包。

这加剧了将业务逻辑映射到我们已经创建的抽象认知的开销。类／接口／软件包的数量越多，一举拿下就越困难，这样做证明了我们构建的这些不同类／接口／软件包所需要的维护成本（很大）是合理的。

## 更少的参数

较少函数的支持者几乎总是倾向于支持将更少的参数传递给函数。

函数参数较少的问题在于，存在依赖关系不清晰的风险。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-10.jpg)

我已经看到了 Ruby 类有 5-10 个方法，所有这些方法通常会做一些非常简单的事情，并且可能会有一两个变量作为参数。我也看到他们中的很多人改变了共享的全局变量的状态，或者依赖于没有明确传递关系的单例，只要存在一种情况，就（跟我们之前的讨论的）是一种相反的模式。

此外，当依赖关系不明确时，测试将变得更加复杂，在针对我们的 itty-bitty 函数的独立测试前，需要重新设置和修改状态值，才能让它运行。

## 更难阅读

这已经在前面陈述过了，但值得重申的是 - 小函数的爆炸式增长，特别是一行的函数，使代码库难以阅读。这尤其会伤害那些代码应该被优化的人 - 新手。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-11.jpg)

代码库中有几种类型的新手。根据我的经验，一个好的经验法则是记住某些可能会检查上述 ”新“ 类别的人。这样做可以帮助我重新评估自己的假设，并重新思考我可能会无意中将某些新手加入到第一次阅读代码的新手中。我意识到，这种方法实际上导致比其他方式可能更好更简单的代码。

简单的代码不代表很容易编写，而且也很少是 DRY 最好的代码。它需要大量细心的思考，关注细节和小心翼翼地达到简单的解决方案，是正确和水到渠成的。这种来之不易的简单性最引人注目的地方在于它适合于新老程序员，易于理解的 ”旧“ 和 ”新“ 的所有可能的定义。

当我对代码库感到陌生时，如果我有幸已经知道其所使用的语言或者框架时，（那么）对我来说最大的挑战是理解业务逻辑或实现细节。当我不那么幸运时，即面临着（必须）通过用我外行的语言来编写代码库的艰巨任务时，我面临的最大挑战是能够对语言／框架有足够的理解，如履薄冰，以至能够理解代码在做什么而不掉进坑里，同时能够区分出我需要真正理解的，跟目标相关的 ”单一事物“，以便在项目前进中取得必要的进展。

在这段时间里，我没有看过一个陌生的代码库，所以我会说：

嗯，这些函数都是足够小，并且符合 DRY 风格的。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-12.jpg)

在我尝试寻找问题答案的同时，我冒险进入了未知领域，真正希望的是让最少数量的思维跳跃和上下文切换。

投入时间和精力，让代码未来的维护者或者消费者变得更容易（理解），这将会产生巨大的回报，特别对于开源项目。这是我希望自己在职业生涯早期做得更好的一件事，而且这段时间我都很注意（这一点）。

## 什么情况下小函数有意义

当所有情况都考虑到了，我相信小函数绝对有它的意义，特别是在测试时。

## 网络 I/O

这不是一篇关于如何最好地为大量服务编写函数，集成和单元测试的文章。然而，当谈到单元测试时，网络 I/O 通过某某方式测试，好吧，实际上没有测试。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-13.jpg)

我不是 mock 函数的粉丝。 mock 函数有几个缺点。首先， mock 是一些结果的人工模拟。只有当我们的想象力和我们具备预测我们应用程序可能遇到的各种失败模式的能力时。 mock 也很可能与他们所支持的真实服务不同，除非每个人都对真正的服务进行过严格的测试（注：对细节很了解）。当每个特定模拟只有一个实例并且每个测试使用相同的模拟时， mock 才是最好的。

也就是说， mock 仍然是单独测试某些形式的网络 I/O 的唯一方法。我们生活在一个微服务时代，并将大部分（如果不是全部的话）关于我们的主要产品的关注都外包给供应商。现在很多应用程序的核心功能都需要一个调用或者多个调用，对这些调用进行单元测试的最佳方法是将其模拟出来。

总体而言，我发现限制能够 mock 的范围，才是最好的。调用电子邮件服务的 API，以向我们新创建的用户发送欢迎电子邮件，（当然这）需要建立 HTTP 连接。将此请求隔离到尽可能少的函数中，并允许我们在测试中 mock ，以最小化代码量。通常，这应该是一个不超过 1-2 行的函数，用于建立 HTTP 连接并返回任何错误以及响应。将事件发给 Kafka 或在数据库中新创建的用户时也是如此。

## 基于属性的测试

对于那些能够通过这种小代码提供如此巨大利益的东西，基于属性的测试却没有被充分利用起来。（这种测试）由 Haskell 图书馆的 QuickCheck 发明的，并在 Scala（ScalaCheck）和 Python（假设） 等其他语言中被采用，基于属性的测试允许人们生成大量符合给定测试规范的输入，并断言每一个情况的测试通过条件。

许多基于属性的测试框架都是针对函数的，因此将任何可能受到基于属性测试的东西隔离到单一函数上是有意义的。我发现这在测试数据的编码或解码，测试 JSON 或 msgpack 解析时尤其有用。

## 结论

这篇文章的意图既不是说 DRY 也不是说小函数本身就是坏的（尽管本文的标题给出了这样的暗示）。只是说他们本质上没有好坏。

代码库中的小函数的数量或平均函数长度本身并不是一个可以吹嘘的指标。在 2016 年 PyCon 谈话中有一个名为 onelineizer 的话题，讲述了一个可以将任何 Python 程序（包括它本身）转换为一行代码的同名 Python 程序。虽然这使得会议讨论变得有趣而诱人，但在相同的问题上编写（类似的）产品代码将显得非常愚蠢。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/small-functions-considered-harmful/Small-Functions-considered-Harmful-14.jpg)

上述建议普遍适用，不仅仅对 Go 而言。由于我们编写的程序的复杂性大大增加，而且我们所反对的限制变得更加多变，程序员应该相应地调整他们的思想。

不幸的是，正统的编程思想依然严重受到面向对象编程和设计模式至高无上的影响。迄今为止，广泛传播的很多想法和最佳实践在很大程度上，几十年以来，一直没有受到过挑战，当前迫切地需要重新思考，尤其是，近年来编程格局和范例已经发生了很大的变化。

不改变旧的风格不仅会助长懒惰，而且会让程序员陷入他们无法承受的虚假的安抚感中。

----------------

via: https://medium.com/@copyconstruct/small-functions-considered-harmful-91035d316c29

作者：[Cindy Sridharan](https://medium.com/@copyconstruct)
译者：[gogeof](https://github.com/gogeof)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
