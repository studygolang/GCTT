#### 仅仅两年时间，Go语言在最流行的编程语言中就从排名第65蹿升至排名第17。 本文其是快速增长的背后的信息。

![](https://opensource.com/sites/default/files/styles/image-full-size/public/lead-images/build_structure_tech_program_code_construction.png?itok=nVsiLuag) 
图像来自于Opensource.com

[Go语言](https://golang.org/) ,也被称为谷歌的Go语言，人气正在强劲增长。 虽然诸如Java和C这样的语言继续主导编程，但新模型已经出现，更适合现代计算，特别是云计算领域。 Go越来越多的使用，一部分原因在于它是适合现如今微服务架构的轻量级开源语言。 容器技术宠儿Docker和谷歌的容器编排产品[Kubernetes](https://opensource.com/sitewide-search?search_api_views_fulltext=Kubernetes)都是使用Go构建的。 Go在数据科学领域也占有一席之地，它具有数据科学家正在寻求的整体性能和从“分析师的笔记本到全面生产”的能力。

作为一种工程语言（而不是随着时间的推移而演变的东西），Go以多种方式使开发人员受益，其中包括垃圾收集，本地并发以及许多其他本地功能，这些功能可减少开发人员编写代码以处理内存泄漏或联网应用程序的需求。 Go还提供了很多与微服务架构和数据科学更好适应的其他功能。

正因为如此，Go才被很多感兴趣的公司和项目所采用。 最近新添加的[Tensorflow](（)https://www.tensorflow.org/)的API，以及像[Pachyderm](http://www.pachyderm.io/)（下一代数据处理，版本控制， 和存储）都正在使用Go构建。  Heroku的[Force.com](https://github.com/heroku/force)和[Cloud Foundry](https://www.cloudfoundry.org/)的部分内容也是使用Go进行的编写。  而这个名单也正日益添加更多的名字。

##增长的人气和应用

在2017年9月的TIOBE的GO语言指数，可以清楚地看到2016年以来受欢迎程度令人难以置信的跳跃，更做为一年中评分上升最高的编程语言，被冠名为TIOBE的编程语言2016名人堂冠军。 目前它在月度排行榜上排名第17位，一年前排名第19位，两年前排名第65位。

![tiobe_index_for_go.png](https://opensource.com/sites/default/files/u128651/tiobe_index_for_go.png) 
TIOBE的GO语言指数 [TIOBE](https://www.tiobe.com/tiobe-index/go/)

“2017年Stack Overflow调查”也显示了Go的受欢迎程度的提升。  Stack Overflow对64,000名开发人员的综合调查试图通过询问“最受欢迎，最令人生厌，最期待的语言”来获得开发者的偏好。  这个清单是由较新的语言，例如Mozilla的Rust，Smalltalk，Typescript，苹果的Swift和Google的Go等构成。  然而，连续三年以来，Rust，Swift和Go都能成为排名前五的“最受喜爱”的编程语言。

![stackoverflow_most_loved.png](https://opensource.com/sites/default/files/u128651/stackoverflow_most_loved.png) 
最受欢迎，最令人生厌，最期待的语言, [Stackoverflow.com](https://insights.stackoverflow.com/survey/2017#most-loved-dreaded-and-wanted)

##Go的优势

一些编程语言是从实践中结合优点而设计的，而有些的则是基于学术理论而创造的。 还有一些是在不同的计算时代设计的，以解决不同的问题、硬件或需求。 Go是一种明确设计的语言，旨在利用现代硬件体系结构解决现有语言和工具的问题。  它设计时不仅要考虑到了团队开发，还考虑到了长期可维护性。

作为核心思想，Go是务实的。 在真实的IT世界中，复杂的大型软件是由大型开发团队编写的。 这些团队开发人员从青少年到成年人，通常具有不同的技能水平。 Go可以很容易实现具体功能，适合初级开发人员使用。

而且，作为促进了可读性和易理解的语言，是非常有用的。 鸭子类型（通过interface）和方便特性（如“：=”）的简短变量声明的混合，赋予Go一种动态类型语言的感觉，与此同时还保留强类型语言的优势。

Go的原生垃圾收集让开发人员不需要再进行内存管理，这有助于消除两个常见问题：

* 首先，许多程序员已经期望内存管理自动完成。
* 其次，内存管理需要不同的例程用于不同的处理核心。  手动尝试安排每个配置会显著增加引入内存泄漏的风险。

Go的原生并发是经常并发发起和注销的网络应用程序的一个福音。  从API到Web服务器到Web应用程序框架，Go语言的Goroutine和Channels十分适合于项目将注意力更多投注于网络，分布式功能和（或）服务。

##适合于数据科学领域

从大数据中提取商业价值正快速的成为企业的竞争优势，而这也是编程领域非常活跃的部分，涵盖了人工智能，机器学习等专业领域。  Go在这些数据科学领域拥有多个优势，这正在增加其使用和普及度。

*  出色的错误处理和更容易的调试正在从Python和R这两种最常用的数据科学语言获得关注度。
*  数据科学家通常不是程序员。 Go有助于原型和生产，所以它最终成为将数据科学解决方案投入生产的更强大的语言。
*  性能非常棒，考虑到大数据的爆炸式增长以及GPU数据库的兴起，这一点至关重要。 Go也无需通过调用C / C ++来性能调优，但是保留了让用户这样做的能力。

##Go的扩张之源

软件交付和部署发生了巨大变化。  微服务体系结构已成为解锁应用程序敏捷性的关键。 现代应用程序设计为云与本地结合，可以利用云平台提供的松耦合云服务。

Go是一个明确设计的编程语言，专门为这些新的需求而设计。  由于特意为云计算而编写，Go 正因为他的并发操作和优美设计而日益流行

不仅Google支持Go，还有其他公司也在帮助Go扩大市场。 例如， [ActiveState's ActiveGo](https://www.activestate.com/activego)就支持Go并进行了扩展同时进行了企业级别的分发。作为开源活动，通过[golang.org](https://golang.org/) 网站和[GopherCon](https://www.gophercon.com/) 年会共同构成了一个强大的现代开源社区的基础，促使将新思路和新动力纳入Go开发进程中。

----------------

来源：https://opensource.com/article/17/11/why-go-grows

作者：[Jeff Rouse](https://opensource.com/users/jeffr)
译者：[mosliu](https://github.com/mosliu)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
