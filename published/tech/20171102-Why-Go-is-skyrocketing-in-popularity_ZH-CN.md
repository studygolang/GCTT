已发布：https://studygolang.com/articles/11731

# 为何 Go 的人气正在飞涨

**仅仅两年时间，在最流行的编程语言中，Go 语言从第 65 名飙升至第 17 名。以下是其快速增长的原因。**

![](https://opensource.com/sites/default/files/styles/image-full-size/public/lead-images/build_structure_tech_program_code_construction.png?itok=nVsiLuag)
图像来自于 _opensource.com_

[Go 语言](https://golang.org/) ，也被称为谷歌的 Go 语言，人气正在强劲增长。 虽然诸如 Java 和 C 这样的语言继续主导编程，但新编程模型已经出现，它们更适合现代计算，特别是云计算领域。 Go 越来越多的被使用，部分原因在于它是一种轻量级的开源语言，适合当今的微服务架构。 容器技术宠儿 Docker 和谷歌的容器编排产品 [Kubernetes](https://opensource.com/sitewide-search?search_apiviews_fulltext=Kubernetes) 都是使用 Go 构建的。 Go 在数据科学领域也占有一席之地，它具有数据科学家正在寻求的整体性能和从“分析师的笔记本到全面生产”的能力。

作为一种工程语言（而不是随着时间的推移而演变的东西），Go 以多种方式使开发人员受益，其中包括垃圾收集，原生并发以及许多其他原生功能，这些功能可减少开发人员编写代码以处理内存泄漏或网络应用程序的需求。 Go 还提供了很多更适合微服务架构和数据科学的其他特性。

正因为如此，Go 才被很多感兴趣的公司和项目所采用。 最近新添加的 [Tensorflow](https://www.tensorflow.org/) 的 API ，以及像 [Pachyderm](http://www.pachyderm.io/)（下一代数据处理、版本控制和存储）都正在使用 Go 构建。 Heroku 的 [Force.com](https://github.com/heroku/force) 和 [Cloud Foundry](https://www.cloudfoundry.org/) 的部分内容也是使用 Go 进行编写的。 而这个名单也正日益添加更多的名字。

## 增长的人气和应用

在 2017 年 9 月的 TIOBE 的 GO 语言指数，可以清楚地看到 2016 年以来受欢迎程度令人难以置信的跳跃，更做为一年中评分上升最高的编程语言，被冠名为 TIOBE 的编程语言 2016 名人堂冠军。 目前它在月度排行榜上排名第 17 位，一年前排名第 19 位，两年前排名第 65 位。

![tiobe_index_for_go.png](https://opensource.com/sites/default/files/u128651/tiobe_index_for_go.png)
TIOBE的 GO 语言指数 [TIOBE](https://www.tiobe.com/tiobe-index/go/)

“2017年 Stack Overflow 调查”也显示了 Go 的受欢迎程度的提升。  Stack Overflow 对 64,000 名开发人员的综合调查试图通过询问“最受欢迎，最令人生厌，最期待的语言”来获得开发者的偏好。 这个清单是由较新的语言，例如 Mozilla 的 Rust、Smalltalk
Typescript、苹果的 Swift 和 Google 的 Go 等构成。 然而，连续三年以来，Rust、Swift 和 Go 都能成为排名前五的“最受喜爱”的编程语言。

![stackoverflow_most_loved.png](https://opensource.com/sites/default/files/u128651/stackoverflow_most_loved.png)
最受欢迎，最令人生厌，最期待的语言, [Stackoverflow.com](https://insights.stackoverflow.com/survey/2017#most-loved-dreaded-and-wanted)

## Go 的优势

一些编程语言是从实践中结合优点而设计的，而有些的则是基于学术理论而创造的。 还有一些是在不同的计算时代设计的，以解决不同的问题、硬件或需求。 Go 是一个工程语言，旨在利用现代硬件体系结构解决现有语言和工具的问题。 它设计时不仅要考虑到了团队开发，还考虑到了长期可维护性。

作为核心思想，Go 是务实的。 在真实的 IT 世界中，复杂的大型软件是由大型开发团队编写的。 这些团队开发人员从青少年到成年人，通常具有不同的技能水平。 Go 可以很容易实现具体功能，适合初级开发人员使用。

而且，作为促进了可读性和易理解的语言，是非常有用的。 鸭子类型（通过 interface ）和方便特性（如 “：=” ）的简短变量声明的混合，赋予 Go 一种动态类型语言的感觉，与此同时还保留强类型语言的优势。

Go 的原生垃圾收集让开发人员不需要再进行内存管理，这有助于消除两个常见问题：

* 首先，许多程序员已经期望内存管理自动完成。
* 其次，内存管理需要不同的例程用于不同的处理核心。 手动尝试安排每个配置会显著增加引入内存泄漏的风险。

Go 的原生并发是经常发起和注销并发的网络应用程序的一个福音。 从 API 到 Web 服务器到 Web 应用程序框架，Go 语言的 Goroutine 和 Channels 十分适合于将注意力更多投注于网络、分布式功能和（或）服务的项目。

## 适合于数据科学领域

从大数据中提取商业价值正快速的成为企业的竞争优势，而这也是编程领域非常活跃的部分，涵盖了人工智能、机器学习等专业领域。 Go 在这些数据科学领域拥有多个优势，这正在增加其使用和普及度。

* 出色的错误处理和更易于调试正在从 Python 和 R 这两种最常用的数据科学语言获得关注度。
* 数据科学家通常不是程序员。 Go 有助于原型和生产，所以它最终成为将数据科学解决方案投入生产的更强大的语言。
* 性能非常棒，考虑到大数据的爆炸式增长以及 GPU 数据库的兴起，这一点至关重要。 Go 也无需通过调用 C / C++ 来性能调优，但是保留了让用户这样做的能力。

## Go 的扩张之源

软件交付和部署发生了巨大变化。 微服务体系结构已成为解锁应用程序敏捷性的关键。 现代应用程序设计为云与本地结合，可以利用云平台提供的松耦合云服务。

Go 是一个工程的编程语言，专门为这些新的需求而设计。 由于特意为云计算而编写，Go 正因为它的并发操作和优美设计而日益流行。

不仅 Google 支持 Go，还有其他公司也在帮助 Go 扩大市场。 例如，  [ActiveState's ActiveGo](https://www.activestate.com/activego) 就支持 Go 并进行了扩展同时进行了企业级别的分发。作为开源活动，通过 [golang.org](https://golang.org/) 网站和 [GopherCon](https://www.gophercon.com/) 年会共同构成了一个强大的现代开源社区的基础，促使将新思路和新动力纳入 Go 开发进程中。

---

via：https://opensource.com/article/17/11/why-go-grows

作者：[Jeff Rouse](https://opensource.com/users/jeffr)
译者：[mosliu](https://github.com/mosliu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
