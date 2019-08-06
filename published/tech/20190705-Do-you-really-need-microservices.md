首发于：https://studygolang.com/articles/22246

# 你真的需要微服务吗

## 前言

我们已经[设计和构建](https://www.browserlondon.com/services/design-development/) 了十多年的软件，大部分时间我们一直在使用优秀的 Symfony 框架来实现这一目标。 Symfony 是一个传统的单体 PHP 构件集，受 Java Spring 的启发，我们发现它非常适合[企业 Web 应用程序](https://www.browserlondon.com/case-study/insights/) 和[数字产品](https://www.browserlondon.com/case-study/twine/) 的快速开发，而这些正是我们主要经济来源。

然而，去年发布的 Symfony 4 代表了该框架的重点逐渐变化 ; 这变化体现在其远离单体架构和向[微服务](https://en.wikipedia.org/wiki/Microservices) 靠拢，这种变化背后的方法论在过去几年中越来越受欢迎。

为了说明这一转变，新版本在默认情况下使用了微内核（micro by default), Symfony 组织大力宣传其新的微内核设计，声称与 Symfony 3 相比，编写应用程序所需的代码减少了 70%。

除了这些优点外，这一变化意味着运行单个应用程序的开销要小得多，这使得 Symfony 对于微服务体系结构的使用更具吸引力。

## 什么是单体应用和微服务

微服务设计基于将大型传统（单体）应用程序拆分为几个小型、不同的应用程序的概念。这些应用程序将处理单个业务功能领域，并与其他组件协作，就像它们是第三方应用程序一样

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/20190705-Do-you-really-need-microservices%3F/do-you-really-need-micorservices.png)

这真的是一个新事物吗，或者这只是一个具有时髦名字的面向服务体架构（SOA)? 我们不会在这里进行辩论，毕竟你可以到 Slashdot 和 Hacker News 上讨论这个问题。不过，我们要说的是，微服务方法 ( 或者随便你怎么称呼它 ) 主要对大型组织有益。这是因为非常大的应用程序可以被分割成几个不同的服务，每个服务由各自独立的开发团队管理。

微服务体系结构的另一个好处是允许灵活地扩展一个特定组件的数量，而不是整个应用程序。这特性非常适合应用在[弹性云计算](https://www.browserlondon.com/blog/2019/01/28/modernising-hosting-platform/#ECS-and-EC2)，但在大多数情况下，我认为这种效率提高会被一个大而突出的问题所淹没。

## 你真的需要微服务

我的观点是，除非你在 Google 或 Netflix 等拥有数百名软件开发人员的公司工作，否则你可能不需要微服务。事实上，对于大多数中小型企业来说，采用这种设计可能非常不合适。

我将会讲到一些例外，但是微服务的开发和维护成本是很多人都注意到的却又很少谈及的问题。我们可以用一个简单的问题来决定是否适合把微服务作为你的起点 :
（译者注：这句子的原文中有个词语叫[房间里的大象](https://dictionary.cambridge.org/us/dictionary/english/an-elephant-in-the-room)，是指所有人都注意到却又不被提及的问题）

> 你系统中的某个组件（例如用户管理）是否足够复杂，以致于需要多个开发人员全职进行持续开发？

如果答案是否定的，那么微服务方法可能会浪费您的时间和金钱。相反，如果你足够幸运，能够在以后达到这个规模，你可能就可以慢慢地把那些需要多人开发的部分分离出来。

## 为什么微服务在开发和运维上开销更大

由于您不需要处理大量的分布式系统问题，因此单体应用程序通常是一个开销更少的方案。使用像 Symfony 这样的单体框架所通过提供开箱即用的集成特性提供了许多好处，这些特性可以方便地从应用程序的所有区域访问。你基本上可以避免处理以下的这些问题 :

* 分布式系统上的身份验证和授权
* 跟踪多个独立系统上的复杂事务
* 分布式锁
* 服务间的通信
* 多个应用程序上的额外配置管理

## 例外情况（混合的方式）

有时候微服务是合适的，但是根据我的经验，在这些情况下，可伸缩性需求或容错需求超过了必须设计和管理分布式系统的缺点。这里的一个很好的例子是像 [Monzo Bank](https://monzo.com/blog/2016/09/19/building-a-modern-bank-backend) 这样的企业应用，它既需要能够立即按需求进行伸缩，又需要能够确保系统某个区域的故障不会影响到另一个区域 .

我们在 Browser 中多次重复的一个好方法是采用混合方法进行系统设计。这涉及到一个由支持微服务包围的中心整体，但只有在有充分理由的情况下才会如此。例如，我们最近在将 [NLP 处理集成](https://www.browserlondon.com/blog/2019/04/08/textrazor-nlp-ai-save-client-money-time/) 到应用程序中时使用了这种方法。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/20190705-Do-you-really-need-microservices%3F/Do-you-really-need-microservice2.png)

我们已经构建了几个系统，其中核心业务应用程序作为一个整体构建 ( 通常在 Symfony 中 )，由独立的微服务管道处理繁重的数据处理。这不仅允许我们在不影响核心应用程序性能的情况下处理大量数据量，而且如果需要，我们可以在不影响平台的日常操作前提下，将这些组件下线。

理想情况下，你能够清楚地理解规模和未来的开发需求，这对于决定体系结构非常重要。你想快速进入市场吗？您想要支持数百万用户吗？您是否需要处理[大量的数据流](https://www.browserlondon.com/blog/2017/07/13/divide-conquer-manage-multiple-data-streams/)。

尽早做出正确的决定可以增加产品在最短的时间内获得投资回报的机会，而不会妨碍您将来的探索。 在后续计划中将组件微服务化通常比最初的 MVP 开发中微服务化更具成本效益。

## 关于 Browser

我们为创建企业级 Web 应用程序提供更好、更高效的工作环境。我们帮助 Shell、British Airways 和 UK Gov 等客户提高效率，简化业务流程。访问我们的[伦敦的网站](https://www.browserlondon.com/)

---

via: https://itnext.io/do-you-really-need-microservices-e85d7711c78b

作者：[Browser](https://itnext.io/@browserlondon)
译者：[Alex1996a](https://github.com/Alex1996a)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
