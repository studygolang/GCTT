首发于：https://studygolang.com/articles/17639

# PHP 不会死 —— 我们如何使用 Golang 来阻止 PHP 走向衰亡

在过去的十年中，无论是世界 500 强企业或是仅拥有 500 名用户的企业，我的团队都曾为他们开发过软件。在此期间，我的工程团队主要使用 PHP 进行后端开发。2 年前，我们在开发项目中引入了一些东西，这不仅彻底改变了我们产品的性能，也改变了它们的可扩展性————我们将 Golang 引入到我们的开发框架中。

很快，我们发现 Golang 的引用使得我们能够为客户设计更大型，速度提高 40 倍的应用程序。我们可以利用 Go 的强大功能来增强我们用 PHP 编写的产品，并充分利用这两种语言的优劣进行取长补短。

我将解释如何结合 Golang 和 PHP 这两种语言解决实际开发中的问题，这将为你的 PHP 开发带来全新的道路，以此解决[垂死的 PHP 模型](https://software-gunslinger.tumblr.com/post/47131406821/php-is-meant-to-die) 相关的一些问题。

## 首先介绍常用的 PHP 设置

在回答我们如何使用 Golang 来将 PHP 起死回生之前，我们先介绍一下标准的 PHP 设置。

在大多数情况下，PHP 开发者会使用 *nginx Web-server* 和 *php-fpm 服务器*组合运行应用程序。当 *php-fpm* 执行 PHP 代码时，Nginx 提供静态文件并将特定请求转发到 *php-fpm* 。也可以将 *Apache* 与 *mod_php 一起使用*。即使这种工作方式和上面那种略有不同，但它们的原理还是类似的。

对于开发者来说，理解 *php-fpm* 如何执行应用程序的代码是最有趣的。当一个请求发送时，*php-fpm* 启动一个 PHP 子进程，并且将请求内容作为进程*状态的*一部分（`_GET`，`_POST` 和 `_SERVER` 等）。在执行 PHP 脚本期间，状态无法更改，因此获取一组新输入数据的唯一方法是销毁该进程并重新开始。

像这样的执行模型有很多好处。你不必担心内存使用情况，所有进程都完全隔离，如果其中任何进程死亡，那么它们将自动创建而不会影响其他进程。但与此同时，当你尝试扩展应用程序时，这一特性会成为程序开发的绊脚石。

## 一般的 PHP 设置使用起来很麻烦并且非常低效

如果你今天正在进行专业的 PHP 开发，那么你应该已经知道开始一个新项目的第一步 - 选择框架。框架提供了依赖注入，ORM，翻译和大量丰富的库。当然，所有用户输入数据都可以方便地放在一个对象（*Symfony/HttpFoundation* 或 *PSR-7*）。框架用起来是那么得心应手！

但任何事都有两面性。所有的企业级框架都要求你加载至少十二个文件，构造多个类并解析一些配置，以便处理简单的用户请求或查询数据库。最糟糕的部分是每个任务完成后，你不得不抛弃这些代码。你刚刚启动的所有代码现在都变得无用，并且永远不能拿来处理另一个请求。若是说给任何使用 PHP 之外的开发人员听，他们一定会对此满脸困惑，不能理解。

多年来，聪明的 PHP 工程师一直试图通过使用延迟加载技术，微框架，优化良好的库，二级缓存等技术来缓解这些问题。但是在你项目结束时，你仍然不得不扔掉你的整个流程并一遍又一遍地重新开始重复的工作。

## 在 Golang 的帮助下，PHP 能否支持多请求？

只要不是几小时或几天的生命周期，编写生命周期超过几分钟的 PHP 脚本还是可以的：比如 cron 作业，CSV 解析器和队列使用者。所有这些脚本都遵循相同的过程：检索值，执行作业，等待下一个值到来。代码在整个过程中都保留在内存中，最终只能节省几毫秒，因为加载框架和引导程序需要进行大量的交互。

开发能够长时运行的脚本并不容易。任何错误都会彻底杀死进程，诊断内存泄漏非常麻烦，我们无法再使用 f5-debug。

然而，随着 PHP7 的推出，情况有所改善，这个版本提供了一个可靠的垃圾收集器，使得错误更容易得到处理并防止核心内存泄漏。虽然工程师们仍然得小心他们的代码中的内存和状态问题，但是，你不必担心找不出问题所在并有效解决这些问题。

是否有可能采用模型来处理那些，需要长期运行的 PHP 脚本并使其适应更复杂的任务需求，如处理 HTTP 请求和消除每个请求的引导加载？

首先，我们需要实现一个服务器程序，该程序可以接受 HTTP 请求，然后逐个将它们转发给 PHP 工作者，而不是每次都杀死工作者。

我们知道我们可以使用纯 PHP（PHP-PM）实现 Web 服务器，或者使用 C-extension（Swoole）编写。虽然这两种方法都有各自优势，但两者都不能让我们满意，我们需要更好的方法。

我们需要的不仅仅是一个 Web 服务器，而是希望能够去掉 PHP 开发中的繁重操作和其他负面因素同时，仍然保障每个应用程序的可扩展性和多样性。我们需要一个能够多元化的应用服务器。

Golang 可以帮助我们创建这样的应用服务器吗？我的答案是，它可以。因为这种语言是跨平台的，它可以将应用程序编译成单个二进制文件，我们还可以利用其非常优雅的并发模型和 HTTP 标准库，最重要的是，我们可以使用 Golang 所拥有的数千个开源库和集成环境。

## 如何使两种编程语言进行一体化开发

首先，我们需要了解两个或多个应用程序如何相互通信（进程间通信）。

一种方法是使用 Alex Palaistras 在英国发布的[令人生畏的库](https://github.com/deuill/go-php)，可以在 PHP 和 Golang 进程（类似于 Apache *mod_php*）之间共享内存。但是，这种库在我们实际开发中对我们造成了很大的限制。

我们决定使用另一种更经典的方法，即使用 Socket/Channel 上的二进制流完成进程之间的通信。我们选择这种方法是因为这种通信方法被使用了数十年，是一种可靠的通信方法，并且在操作系统级别上得到了很好的优化。

首先，我们创建了一个轻量级二进制协议，用于在进程之间交换数据并处理错误。在最简单的实现方法中，这种类型的协议是类似于 netstring 的实现，拥具有[固定大小的包头](https://github.com/spiral/goridge/blob/master/prefix.go)（点击这里查看我们给出的例子），其包含每个包类型，大小和二进制掩码等信息，以便验证数据完整性。

在 PHP 方面，我们使用了[*包* PHP 函数](https://github.com/spiral/goridge/blob/master/php-src/SocketRelay.php#L247)。对于 Golang，我们使用了[*编码 / 二进制*库](https://github.com/spiral/goridge/blob/master/prefix.go)。

我们甚至在创建协议上更进一步。添加了[直接从 PHP](https://github.com/spiral/goridge/blob/master/codec.go) 调用[Golang *net / rpc*服务的功能](https://github.com/spiral/goridge/blob/master/codec.go)。这个功能在开发中非常实用，因为我们可以轻松地将 Golang 库集成到我们的 PHP 应用程序中。你可以在我们发布的另一个名为[Goridge 的](https://github.com/spiral/goridge) 开源产品中看到这项工作的结果。

**实现 PHP 高并发处理任务**

一旦建立了通信，下一个目标就是最有效地将作业传递给 PHP 进程。对于任何传入作业，应用程序服务器必须选择一个空闲工作程序来执行所需任务。如果 worker / process 失败或死亡，我们会舍弃它并为他创建一个替代的进程。另一方面，如果 worker / process 成功，我们会将其返回池中并使其可用于下一个作业。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/keep-php-from-dying/1.png)

在此需求的实现中，我们使用[有缓冲的通道](https://github.com/spiral/roadrunner/blob/master/static_pool.go) 来存储活动工作池。

**最终结果是一个能够处理任意二进制作业的有效 PHP 服务器。**

为了使我们的应用程序能作为 Web 服务器工作，我们必须选择一个可靠的 PHP 标准来表示任何 HTTP 传入请求。想要满足此需求，我们只需将 Golang *net / HTTP* 请求转换为 [*PSR-7*](https://github.com/spiral/roadrunner/blob/master/service/http/request.go) （https://www.php-fig.org/psr/psr-7/meta/）格式，使其与大多数市场上的 PHP 框架兼容。

由于 PSR-7 格式是不可变的（一些工程师可能会指出它在技术上不可变），它迫使开发人员编写不再将请求视为全局实体的应用程序。这完全符合长期运行 PHP 进程的想法。最终实现看起来流程如下：

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/keep-php-from-dying/2.png)

## 介绍 RoadRunner- 一个[高性能的 PHP 应用服务器](https://github.com/spiral/roadrunner)

我们最初的测试用例是一个用于后端的 API，它经常难以预测的出现突发请求的次数比平时高出许多倍的情况。虽然在大多数情况下 *nginx* 可以帮忙处理，但是出现 502 错误的情况会频繁发生，因为我们无法预料到什么时候负载增加，做不到在负载增加之前快速地平衡系统。

在 2018 年初，我们将第一个 PHP / Golang 应用服务器部署到市场中以取代此设置。**效果立竿见影，令人难以置信。**我们不仅完全消除了 502 错误的发生，而且我们最终将服务器总数减少了近三分之二，这为工程师们和产品所有者节省了大量工作成本和服务器成本。

到 2018 年中期，我们对该方法进行了优化，并在 MIT 许可下将其发布到 GitHub，并称之为 [**RoadRunner**](https://github.com/spiral/roadrunner)，它实现了其令人难以置信的速度和效率。

## RoadRunner 如何帮助开发

将 [RoadRunner](https://github.com/spiral/roadrunner) 引入我们的技术栈使我们能够使用中间件进行 HTTP 通信，在请求进入 PHP 之前启用 JWT 验证，处理 WebSockets 并将统计数据汇总到 Prometheus 中。通过使用嵌入式 RPC，我们可以将任何 Golang 库中的 API 传递给 PHP 使用，而无需自定义驱动程序。最重要的是，我们可以使用 [RoadRunner ](https://github.com/spiral/roadrunner) 库来设置与 HTTP 不同的新服务器。示例包括在 PHP 中运行 [AWS Lambda ](https://github.com/spiral/roadrunner/wiki/AWS-Lambda) 处理程序，创建可靠的队列使用，甚至将[ GRPC ](https://github.com/spiral/php-grpc) 添加到我们的应用程序中。

到目前为止，在 PHP 和 Golang 开发社区的共同帮助下，我们改进了调试工具，将其与 Symfony 框架集成，并增加了对 HTTPS，HTTP / 2，和 PSR-17 的处理。我们提高了程序的稳定性，并且在一些测试中，程序的性能提高了 40 倍之多。

**结论**

有些人仍然坚持认为 PHP 是一种缓慢，笨重的语言，只能用来编写 WordPress 插件。他们甚至可能会说 PHP 有一个限制：一旦你的应用程序变得比较大，你就必须切换到更“成熟”的语言并取代之前的 PHP 代码。

对他们来说，我们想说“请三思”。我们认为 PHP 的唯一限制是你自己给自己的限定。你可以花一生的时间从一种语言跳到另一种语言，试图找到满足你编程需求的“完美匹配”，或者你可以开始将语言本身重新设想为工具。像 PHP 这样的编程语言的表面缺点实际上可能是其成功的关键。通过将其与 Go 等其他语言配对，你最终可以创建比你自己使用任何一种语言更强大的产品。

在使用 Go 和 PHP 进行混合编程一段时间之后，我们可以自信地说我们都很喜欢这种开发方式。我们不打算放弃，而且我们将继续寻找从这种双栈编程中获得最高效率的方法。

Spiral Scout 是一家领先的软件开发公司，为旧金山和美国各地的客户提供从小型网站到大型分布式系统的定制产品的全栈开发。如果你有 PHP 或 Golang 相关项目，或者你遇到了应用程序在 PHP 中无法扩展或被过时的代码[压缩限制](https://spiralscout.com/)，请通过[spiralscout.com](https://spiralscout.com/) 与我们的团队[联系](https://spiralscout.com/)。

[RoadRunner Creator: Anton Titov, CTO, Spiral Scout](https://github.com/wolfy-j)

---

via: https://blog.spiralscout.com/php-was-never-meant-to-die-830de87915ee

作者：[John W. Griffin](https://blog.spiralscout.com/@johnwgriffin)
译者：[CNbluer](https://github.com/CNbluer)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studyGolang/GCTT) 原创编译，[Go 中文网](https://studyGolang.com/) 荣誉推出
