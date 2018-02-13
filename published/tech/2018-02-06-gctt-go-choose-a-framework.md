已发布：https://studygolang.com/articles/12399

# 选择一个 Go 框架

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/choose-framework/heading.jpg)

每一天，或者是每隔几天，总有人来到 [/r/golang](https://www.reddit.com/r/golang/) ，并询问些类似如下的问题“哪个框架是最好的？”。我认为我们应该尝试提出这个问题，至少以一种容易理解的方式提出。你不应该使用框架。

对于一个复杂的问题，这也许是一个非常简洁的答案。这并不是说你任何时候都不应该使用框架。众所周知，当我们开发软件的时候，有这么一种趋势，慢慢形成适应通用开发的模式，并且一次又一次地加快开发同样的东西。它试着尽可能的消除重复代码。

## 标准库或 stdlib

Go 语言标准库的质量很高。你应该尽可能的使用它。如果你正在编写 API 服务，你需要熟悉 `net/http` 包，而且你最终无论使用哪个框架都是基于这个包的。将第三方包导入标准库是需要认真考虑的，当他们解决的是一个非常专注的问题时，这是不适合进入标准库的。举个例子，生成 UUID 的包，或 JWT 包。一些包（包括 web 框架）都是基于标准库构建的。一个很好的具体例子，[jmoiron/sqlx](https://jmoiron.github.io/sqlx/) 包就是基于 `sql/database` 包构建的。

## 包管理器

Go 和包含一个软件包市场的语言有很大的不同。Node 有 [npm](https://www.npmjs.com/)，PHP 有 [packagist（composer）](https://packagist.org/)，Ruby 有 [gems](https://rubygems.org/)。事实上 Go 没有官方软件包管理器（一个中央市场），这对于寻找到一个软件包有着重大的影响。有几个软件包管理器可用，例如 [gvt](https://github.com/FiloSottile/gvt)，[glide](https://github.com/Masterminds/glide) 和一个官方的试验品 dep，[dep](https://github.com/golang/dep) 可能会在将来的某个时间和 Go 语言工具链一起捆绑发布。

> dep 是官方的试验品，但还不是官方的工具。查看更多关于 dep 的[路线图](https://github.com/golang/dep/wiki/Roadmap)！

实际上，在官方的 readme 文件中 Glide 建议迁移到 dep。当你开始使用包管理器，建议使用 dep。

Go 代码放在 GitHub，Bitbucket 和其他存储库，甚至可以自行托管。最完整的 Go 语言包列表可以通过搜索 godoc 来找到，godoc 是由这些包生成的文档的中央托管。它不提供类似于上面提到的其他语言项目的软件包市场。有一个值得注意的项目：有点接近其他的软件包管理器，被用作包索引，它是 [gopkg.in](http://labix.org/gopkg.in)。

任何你创建的软件包都可能因为代码质量等各种原因而没有被采用。某些程序包作者比其他人更为人所知，如果你正在寻找一些像 flags 配置包那样微不足道的东西，那么你只剩下少量但是高质量的选择。这在相关的包软件系统中是不可能的。

相比之下，Node 的软件包质量急剧下降，因为它的焦点分布在前端和后端之间，所以 Node 中有许多怪癖的代码，这在 Go 中是不存在。Go 是一个完全服务器端的语言，而 Node 往往服务于很多前端的 Javascript 代码。Node 开发中有一些东西，无法用任何逻辑来解释，例如开发和采用 [left-pad](https://www.npmjs.com/package/left-pad) 和 [is-array](https://www.npmjs.com/package/is-array) 包的原因。我无法解释它们为何每周有成百上千万的下载量。

## Go 生态系统

Go 有一个较小的生态系统，但是有很多基于 Go 的项目都被广泛的采用，最近 GitHub 上的 [go-chi/chi](https://github.com/go-chi/chi) 有 2500 星星和非常好的评论（和 sqlx 类似，chi 项目是基于底层的 `net/http` 包构建的）。我们在 [ErrorHub](https://errorhub.io/) 上使用它，我建议你使用它。

有许多 web 框架可用，但是如上所述，你应该首先使用 stdlib，这样你可以在继续前进时明白你真正需要的。使用 web 框架本身是完全没有必要的，但是当你有新的需求时，你可以做出更明智地选择从哪里迁移。

## 从其他语言迁移到 Go

Go 和其他语言之间的不同之处在于语言细节。从 Python 迁移到 Ruby 或从 PHP 迁移到 Javascript 时，你会发现同样的差异。Go 也不例外。你可能会发现[（例如切片是如何工作的）](https://scene-si.org/2017/08/06/the-thing-about-slices/)起初有点混乱，但从任何语言迁移到任何其他语言时都会遇到这些问题。让我们再看看 ruby 的一个例子 [predicate methods](http://ruby-for-beginners.rubymonstas.org/objects/predicates.html)。

Go 的入门门槛真的很低。我在 15 年前使用 PHP，迁移到 Go 是相对比较简单的。让你理解 Node 的异步操作是很困难的，包括 Promise 和 yield。如果我能推荐两篇阅读材料，那么你应该阅读一下 [the interview with Ryan Dahl, the creator of Node](https://www.mappingthejourney.com/single-post/2017/08/31/episode-8-interview-with-ryan-dahl-creator-of-nodejs/)，[Bob Nystroms critique of asynchronous functions](http://journal.stuffwithstuff.com/2015/02/01/what-color-is-your-function/) 也是必读的。

> 这就是说，我认为 Node 并不是构建大型服务器网站的最佳语言。我会用 Go 的。说实话，这就是我离开 Node 的原因。我意识到：哦，实际上，这不是有史以来最好的服务器端系统。
>
> **Ryan Dahl**

## Go 的力量

Go 非常适合为任何您选择驱动您的前端框架的项目提供 API 端点。Websockets？没问题。一群正在互相交谈的 Go 程序？你有没有听说过 Docker 或者 Kubernetes？这些系统有着令人难以置信的可扩展性，并且是用 Go 写的。

Go 是一门出色的语言，可以提供后端逻辑，例如和数据库交互接口，并通过 HTTP API 端点开放访问。前端技术栈的选择将确保你可以为浏览器使用和呈现此数据。

你被 React 所困恼吗？将其替换为 VueJS 而不用丢弃任何 Go 代码。在其他语言中，你必须严格遵守这个原则来分割应用程序，因为通常情况下，你不是在编写服务器，而只是生成将在浏览器中运行产生输出的脚本。当然，使用 Go 可以以相同的方式使用 `html/template`，但是选择使用前端框架实现前端，将会给你带来好处：专注于该框架的开发人员。不是每个人都喜欢 Go。

你不会用 bash 写一个 web 服务器，对不对？

## 你为什么要用 Go？

对我来说主要的卖点就是标准库，语言和文档的质量非常高。Ruby，Node，PHP 和其他以 web 开发为中心的语言通常都是单线程的，如果可能的话，超出这个范围的通常都是使用一个附加组件，而不是一等公民。他们的内存管理很差（尽管，至少 PHP 在过去的 15 年里有了很大的改进），也许最重要的是它们都属于脚本语言范畴。编译的代码总是会比通过解释器运行的任何代码都快。

人们总是重新发明轮子，不仅仅是因为他们可以，而且还因为他们可以以某种方式改善它。这可以以很小的增量完成，例如优化一个生成特定输出的特定函数，或者可以以更大的增量完成，例如创建一门将并发性作为一等公民的编程语言。

其他的东西可能不会被处理（这就是为什么在 Go 中有很多关于泛型的争议），但总是有可以改进的空间。我不排除有人会另辟蹊径去提供他们认为最合适的最好的泛型的可能性。人们总是根据他们的经验做出反应。如果你的经验告诉你，并发性是某些 XY 语言的问题，那么你将会找到解决办法。迁移到 Go 是一个很好的方法来解决这个问题。

## 笔记

这篇文章反映了 Go 确实有包管理器，但是到目前为止还没有官方的工具，也没有和 Go 的工具链一起捆绑发布。前面的文章误导了这一点，它暗示了 Go 根本没有包管理器。从技术上讲，所有其他包管理器（至少 npm 和 composer）都是是附加组件。

## 我很荣幸你能够阅读本文...

但如果你能购买一本我的书，那真是太棒了：

- [API Foundations in Go](https://leanpub.com/api-foundations)
- [12 Factor Apps with Docker and Go](https://leanpub.com/12fa-docker-golang)
- [The SaaS Handbook (work in progress)](https://leanpub.com/saas-handbook)

我保证如果你买一本书，你会学到更多东西。购买一份拷贝支持我写更多关于类似的话题。谢谢你买我的书。

如果你想预约我的咨询/自由服务时间，请随时给我发[电子邮件](black@scene-si.org)。 我非常擅长 APIs，Go，Docker，VueJS 和扩展服务[等等](https://scene-si.org/about)。

---

via: https://scene-si.org/2017/10/18/choosing-a-go-framework/

作者：[Tit Petric](https://scene-si.org/about/)
译者：[MDGSF](https://github.com/MDGSF)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出