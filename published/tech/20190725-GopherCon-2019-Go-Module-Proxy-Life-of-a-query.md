首发于：https://studygolang.com/articles/25559

# GopherCon 2019 - Go 模块代理：查询的生命周期

## 概述

Go 团队已经搭建了模块镜像与校验和数据库，这将提升 Go 生态环境的可靠性与安全性。这次的交流会通过 Go 命令、代理与校验和数据库讨论经过身份验证的模块代理的技术细节。

## 介绍

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-katie.jpg)

Katie Hockman ，谷歌软件工程师，在 NYC 的 Go 开源团队工作，是构建 Go 模块镜像与校验和数据库的工程师之一。

Katie 演讲的是关于 Go 的包管理与身份验证发生的一些新事物。她希望这样可以帮助你更确切的理解这些事情如何运作的。

点击[这里](https://github.com/katiehockman/puppies/blob/master/presentation_slides.pdf)可以观看 Katie 演讲时使用的幻灯片。

## 一些背景

Katie 十分喜爱狗。

她希望创建一套 Go 包，以帮助其他人可以成为更好的狗主人。如果狗开心，她就开心。

她在 GitHub 创建了新的[储存库](https://github.com/katiehockman/puppies)，并开始为她想要提供的所有源代码创建 Go 包。

* 首先她创建了 [walk](https://github.com/katiehockman/puppies/tree/master/walk) 包，可以通过算法算出在你附近最适合的狗子步行路线。
* 然后她创建了 [bark](https://github.com/katiehockman/puppies/tree/master/bark) 包，可以根据你提供的狗子吠叫的视频告诉你狗在想什么。
* 最后她创建了 [toys](https://github.com/katiehockman/puppies/tree/master/toys) 包，可以提醒你每周买新的玩具给你狗，这样它就不会感到无聊。

这样代码就必须依赖大量的 API。她需要依赖存储、音频处理和地图 API。

因此产生了忧虑：

1. 她担心会因不可重现的构建而面临挑战。她依赖正在使用的地图服务的全新 API 端点呢。如果引入她的包的人使用的是旧版本的 API ，那么他们的程序就可能构建失败，而且她无法保证他们的构建是一致的。
2. 也许可以完全停止依赖。但是真的很难，因为这样她必须从头编写一堆代码；她还不可以复制所依赖的一些 API ，因为代码质量会受到影响。
3. 更糟糕的情况是，有不怀好意的人试图做出不轨的行为，攻击保存了她所依赖的源代码的服务器。因此使得她获取到错误的代码导致她的包失去可靠性。这些恶意的代码对依赖她代码的人来说可能具有危险性，当她弄清楚情况时，可能为时已晚！

为了保护自己依赖她代码的人，她想到了一些方案和问题。

1. 也许可以完全停止依赖。但是真的很难，因为这样她必须从头编写一堆代码，她还不可以复制她所依赖的一些 API ，因此她的代码质量会收到影响。
2. 她可以在代码里提供所有的依赖，但是这样子会造成她代码库的过大，并且她担心随着时间的推移代码很难维护和更新。
3. 或者，她可以选择什么都不做，认为今天的所用的依赖是可信的。相信她所用到的依赖不会消失，github 和其他代码托管网站会一直为我和其他人提供所需要的代码。但说实话，她更相信她很随性。

她有一些想法，但是没有一个可以解决所有问题的同时规避新的问题与风险。

因此，让我们一起探索最近在 Go 中出现的一些更好的解决方案。

1. 模块可以帮助我们解决可重现的构建问题
2. 模块镜像可以帮助我们避免依赖丢失的风险
3. 校验和数据库可以帮助我们避免获取到错误代码的风险，保证每一位 Go 开发者在同一时间获取到的代码是相同的
今天她打算讨论下面三个事情：模块，镜像与校验和数据库

## 更好的解决方法：模块

模块是一组版本化的 Go 包，它们以某种方式相互关联。

在她的 存储库，她有用到一些 Go 包：walk ，bark ，还有 toys 。这些包相互关联，并共享许多相同的依赖。她将这些包放在一个现在可以进行版本控制的模块中。

![ puppies 模块](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-9.png)

模块版本有一个主版本号，次版本号和修订号来组成它的语义版本。

![语义版本]](https://d33wubrfki0l68.cloudfront.net/03a164b7facd35432ff6aebea65826dfe8bc4e48/13c23/gophercon-2019/go-module-proxy-life-of-a-query-10.png)

如果你想在模块存在之前导入包，你要么建一个 vendor 库保存你引用的包的源代码，或者依赖最新版本的包。

现在，包可以放置在模块里面，该模块是及时版本化的快照，作为唯一标识。版本的内容永远不允许改变。

![包可以放置在模块里面，不可改变且向后兼容](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-11.png)

每一次模块的主版本提交必须向后兼容。

对于某个模块，它只需要一个位于其根目录下的 go.mod 文件。

go.mod 文件内容大致如下：
```
module github.com/katiehockman/puppies

require (
  github.com/maps/neighborhood v1.4.1
  github.com/audio/dogs v0.19.2
  golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
)
```
go.mod 文件指定了你使用的包最低版本。该文件是你唯一需要去查看的文件，以便了解模块具有哪些直接依赖关系。

你可以看到 v1.4.1 这样符合语义化版本的版本号，也可以看到 v0.0.0 后面接着时间和提交哈希值这样的伪版本号，这样方便你依赖特定的提交或者存数库没有配置版本标签（ version tags ）。**（译者注：这里她应该是想表达的是，从 go.mod 可以了解模块依赖的哪个版本标签的代码或特定的某一次提交的代码，同理，模块的作者可以通过编辑 go.mod 的方式去指定依赖的代码版本）**

![ go.mod ](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-13.png)

通过指定你的代码依赖于 1.4.1 或更高版本的模块，你就可以保证每个引入你包的人将永远不会使用 1.4.1 之前的版本。**（译者注：就是 go.mod 指定你模块的依赖项是某个特定的版本，比如 1.4.1 ，那么引入你模块的人在使用模块封装的方法的时候，所使用依赖项不会是 1.4.1 之前的版本）**

go 命令使用称为『最小版本选择』（ minimal version selection ）来构建模块，该模块基于 go.mod 文件中指定的版本。

举个例子，假设我们有模块 A 和模块 B ，都是 github.com/katiehockman/puppies 的依赖项。

![ 『最小版本选择』例子 ](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-15.png)

每一个模块都有依赖项 C ，但版本不一样。 A 要求 C 的最低版本是 1.5 ，而 B 要求 C 的最低版本是 1.6。

C 还发布了另一个更新的版本： 1.7 。

如果她的模块依赖模块 A 和 B ，那么当她进行构建的时候 Go 命令会选择同时满足 A 和 B 的 go.mod 文件约束的最小版本。在这个例子里，将选择 1.6 版本。**（译者注：A 要求最低版本是 1.5 ，而 B 要求最低版本是 1.6 ，因为模块的向后兼容性，所以这里 1.6 版本 理应兼容 1.5 版本， 所以 Go 命令会选择 1.6 版本）**

![ 『最小版本选择』例子 ](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-16.png)

这个最小版本选择保证了构建的可重复性（ reproducible builds ）， 因为 Go 命令构建的时候用的是每个依赖项的允许的最小版本，而不是每天可能更改的最新版本。**（译者注：如果选择的版本不能兼容模块 A 和 B 的要求，可能导致其中一方的代码错误，比如：某个函数的不存在，导致构建失败。所以我们在编写模块的时候，也要注意在更新代码的时候尽量不要删除旧的方法或常量，否则会破坏模块的可向后兼容性。这里有个词语：<u>reproducible builds</u> ，可以点击[这里](https://reproducible-builds.org)了解一下）**

现在我们就拥有了一致的，可重复的构建方案，而不再依赖于依赖项的最新提交历史。

一个问题解决了，就着手第二个问题。现在我们完成了可重复构建，接下来让我们讨论如何确保我们的依赖关系不会消失。

## 更好的解决方法：模块镜像与代理

现在我们的模块已经版本化，对应版本的内容是固定的，不允许改动的，我们可以将它缓存下来并做验证，而这是非模块模式没有的功能。

这里是模块代理进入的地方。

![ 模块代理 ](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-19.png)

上图就是没有代理的流程。很简单，是吧？当 Go 开发者执行 Go 命令时，比如： `go get` ， 当本地缓存没有对应的包的时候， Go 命令会直接访问源服务。这里通常指的是托管 Go 源代码的地方，比如： Github 。

`go get` 的行为是否更改是根据 Go 命令是否在 module-aware 模式或 GOPATH 模式下运行。代理仅在 module-aware 模式下运行，这我们一会讨论。

我们将讨论 `go get` 两个主要的工作：

* 请求并拉取你请求的源码
* 基于你拉取的源码感知你的模块是否有新的依赖。它必须解决这些依赖关系，并可能需要更新你的 go.mod 文件

但是没有代理，这过程的代价可能变得非常昂贵。延迟和系统存储都是如此。

go 命令可能会强制拉取整个源代码，即使现在不会去构建它，只是为了在解析期间做出关于依赖版本的一致决定。

所以对于依赖解析，这就是 Go 命令获取的内容：
![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-21.png)

但是这才是它实际需要的：
![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-22.png)

请记住， Go 命令唯一需要理解模块依赖关系的是一个 go.mod 文件。因此，对于 20 MB 的模块，只需要几 KB 的 go.mod 来执行此依赖解析。有许多缓存在你系统上的数据是用不到的，并且还浪费你大量的时间去请求源服务与拉取，这本是不必要的行为。

对于那些一直没有使用模块代理的人来说，这就是你看到的一些延迟背后的原因。这就是模块代理的用武之地。

回到我们的示例，使用 Go 命令获取源代码，让我们将模块代理放入图片中。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-24.png)

如果你告诉 Go 命令使用代理，就不是像以前那样直接向源服务器发起请求，它将会向代理请求它想要的东西。

现在，我们先不关注源服务器， Go 命令可以与具有更适合其需求的 API 的代理进行交互。

让我们来看看与代理的这种交互是什么样的。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-33.png)

她正在研究她的 puppies 模块，她决定要导入一个新的包，告诉我们有关不同犬种的信息。

首先要做的是她要执行 `go get go.dog/breeds` 。

代理将会返回版本列表， Go 命令查询最新的版本，在这里指的是 v0.3.2 。

现在 Go 命令得知了最新的版本编号，准备对 go.dog/breeds v0.3.2 的代理 /info 端点发起请求。此信息端点将提供有关该版本的一些额外元数据。

此元数据包括此标签或分支的规范版本，以及其提交时的时间戳。**（译者注：这两段对应图上的请求 /v0.3.2.info 接口，返回 {"Version": "v0.3.2", "Time": "2019..."}）**

go 命令将使用代理服务器提供的这个规范版本。

下一步， Go 命令开始捋清和解决 go.dog/breeds v0.3.2 的依赖。这个过程与 go.mod 文件有关。

现在已经完成了依赖项解析，它必须实际获取最初请求的源代码，因此接下来会再次从代理请求包含源码的 .zip 文件。

这里真正有趣的是能够从代理获取有关模块的递进依赖的信息，而不需要再通过整个源码 zip 来执行此操作。**（译者注：我的理解是 go.mod 文件已经把依赖项的依赖项也列了出来，因此 Go get 的时候不需要通过解压依赖项的源码 zip 文件，来查询该依赖项的依赖项。）**

go 命令只需要根据最小版本选择去请求得到它进行依赖项解析所需要的信息，而不必查看其余部分。

在这一点上，你可能想知道这个流程的某一部分。 .info 文件是做什么用的，为什么 Go 命令需要它？在这个例子中，我们请求 0.3.2 版本，它只是返回我们提供的相同规范版本。**（译者注：这里应该指的是符合[语义化版本控制规范](https://semver.org/lang/zh-CN/)的版本号）**

让我们看一个不是这种情况的例子。

现在在命令 `go get go.dog/breeds` 的末尾添加 `@master` （`go get go.dog/breeds@master`）。这是告诉 Go 命令我们要请求的是当前 master 分支的代码，所以我们可以完全跳过命中列表端点。**（译者注：就是不获取版本列表了，可以对比上面的图）**

我们将直接跳转到请求 master 分支的信息。

然后你可以看到返回的信息和上一个例子的情况有点不一样，这次我们得到的是伪版本号，它是请求时『 master 』的规范版本。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-39.png)

当它获取到这个版本号，它可以像前面那样处理，下拉并解决依赖关系，最后请求 zip 文件中的内容。

我一直在谈论代理，因为那里没有单一的代理。 任何服务器都可以实现模块代理规范并提供给 Go 命令使用。可以通过运行 `go help goproxy` 获取该规范。

你可以在这里查看规范， /list 端点提供版本列表， .info 端点返回 Json 格式的元数据， .mod 文件提供依赖项解析， .zip 端点提供完整源代码。

我开始介绍这一部分，说模块镜像将解决我们的问题，但后来她开始讨论代理。

好吧，镜像只是一种特殊类型的代理，它将数据和源代码缓存在自己的存储系统中以重新提供给客户端。

镜像可以通过多种方式帮助你。

镜像可以帮助解决我们最初考虑的问题：代码会从源头消失。

它们（镜像）将源代码缓存在它们的存储系统，开发者从 GitHub 拉取他们的代码可能出现的风险将不会出现在它们身上。如果源服务突然因为宕机或者其他原因导致服务不可用，你可以拉取保存在你镜像缓存的源码备份以便抢救。**（译者注：这里翻译起来太难受了... 这里应该说的就是，当我们从 GitHub 拉取代码时可能会遇到代码已删除或别的原因导致拉取失败，这个时候如果我们有搭建一台镜像服务器，且已经把目标代码缓存了下来，那么我们就可以从镜像服务器拉取代码。在我看来，这样只是把风险降低了，并不能说因为有镜像服务器就没有类似上面的风险。）**

你会发现下载速度变得更快。

因为 Go 命令只会查询它所需要的，而不关注所需以外的东西，所以系统上的存储使用会减少。**（译者注：这里应该是相对于执行命令的机器而言）**

幸运的是，使用代理是非常简单的事情。你不需要安装任何东西即可使用代理。你甚至不需要安装 Git 或者 mercurial ，因为代理都可以完成相应的工作。

你的系统需要做的就是能够向代理发出 HTTP 请求。

现在我们拥有可重复的构建，并且我们对我们的依赖关系不会消失更有信心。

让我们谈谈我们如何信任 Go 命令为你提取的源码。

如果没有模块或者代理， Go 命令将使用头部的 HTTPS 直接从源服务器获取源码。

你有理由相信这样子做所获得的内容就是你想要的，但是，源服务器是有可能被黑客入侵，这是 Go 命令无法检测的。

因为你的代码依赖最新的提交，因此当你所引用的包的作者突然在你不知道的情况下修改源码，你的代码能否不受到影响这一点是无法得到保障的。

随着模块的引入，我们还得到一个叫做 go.sum 文件的东西。

当你使用模块的时候你可能就会看到这个文件，并且可能会想知道它存在的目的和你可以用它来做什么（如果有的话）。

这个 go.sum 文件可以作为你下载源代码时代码的样子，即 Go 命令第一次从你的机器上看到它。 Go 命令可以在某些情况下使用它们来检测来自源服务器或代理的不当行为：可能提供与你之前看到的代码不同的代码。**（译者注： go.sum 记录了第一次拉取代码时，代码的样子。因此当源服务器或代理发生了不确定的问题导致源代码发生变动， Go 命令可以在一定情况下，检测出来。）**

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-51.png)

所以她谈论的这看起来令人有些困惑的 go.sum 文件，跟 go.mod 文件一样位于模块根目录。

它基本上是依赖项的 SHA-256 哈希列表及其 go.mod 文件。因为这是加密哈希，所以基本上不可能在不影响哈希的情况下对特定版本中的文件进行任何更改。

你可能从来没有接触过这文件，它是由 Go 命令产生、更新与使用的东西。但了解它的使用方式对于理解其局限性非常重要。

顺便说一句，如果你在别的项目看到 go.sum 文件，它可能比图上这个 go.sum 文件的大的多。它还包括一些不会出现在你的 go.mod 文件中的模块。这只是因为 Go 命令提取的所有依赖项，它还包括依赖项的依赖项、依赖项的依赖项的依赖项等等，即使它们不会出现在你的 go.mod 文件中。

go 命令有一个非常酷的地方，就是可以使用这些校验和：用于检测你准备下载下来的代码是否与你一个月之前看到的是否有不同。

这个 go.sum 文件应该添加到你的存储库，当有人视图下载你的依赖项时， Go 命令可以将其作为信任源。**（译者注：将存储库里的 go.sum 文件每行的源地址作为信任源？）**

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-53.png)

一旦将 go.sum 保存到你的存储库， Go 命令会在获取你代码的时候检查你模块的 go.sum 文件。

假设我们清楚了模块缓存，需要再次重新获取模块的依赖关系。它将获取依赖项的 mod 和 zip 文件，执行哈希，并且可能会看到它刚刚生成的 go.sum 行与你之前的 go.sum 文件中保存的行匹配。

这可能意味着此源代码已被修改，代理或源服务器已被黑客入侵，或者无数的事情。我们所知道的是，它与我们之前信任的不同，我们不应该使用它。

所有这一切都完全没有 go.sum，但它有其局限性。

缺点：它的工作原理是“信任第一次使用”，更具体地说，是你第一次使用。

当你向模块添加一个你以前从未见过的新依赖项时，包括升级到你正在使用的新版本的依赖项时，go 命令会获取代码并动态创建 go.sum 行。 它没有任何东西可以检查它，所以它只是将它们弹出到你的 go.sum 文件中。

问题是，那些 go.sum 行没有与其他人交叉检查。你只是接受你刚下载的代码是正确的代码，并且你的 go.sum 文件将成为你依赖的真实来源。

这意味着你的模块的 go.sum 行可能与 Go 命令刚刚为其他人生成的 go.sum 行不同，可能是因为你在一周内请求它们并且代码已经更改，或者是因为有人给你指向恶意代码。 所以对于我们的问题来说这不是完美，完整的解决方案。

为了增加接收错误代码的额外风险，当她开始讨论代理时，你可能已经意识到了一些非常重要的东西。**（译者注：为了减少风险？）**

谁能说代理实际上是在为你提供你要求的代码？你有什么样的信心可以让代理不是故意瞄准你，并为你提供与其他人不同的东西来伤害你？如果你没有将 go.sum 文件保存到存储库，那么当你从现在起一个月内要求相同的源代码时，如果代理为你提供了不同的服务，会发生什么？

突然之间，我们相对安全的直接端点已经被一个能够欺骗我们的代理所取代，而且它本身并不值得信赖。

在最好的情况下，我们可以想象一个代码作者告诉我们 go.sum 行应该是什么样的，我们总是可以对此进行验证。但是 Go 代码生活在 Github 和 Bitbucket 这么多不同的起源中，没有一个地方可以托管代码，而且这是我们不想改变的 Go 生态系统的重要组成部分。

我们可以做的是下一个最好的事情：让我们确保每个人都同意将相同的 go.sum 行添加到他们模块的 go.sum 文件中。
![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-58.png)

[sumdb 的设计文档](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md)

我们可以通过创建 go.sum 行的全局源来实现这一点，称为校验和数据库（ checksum database ）。

当 Go 命令从代理获取其代码，它就可以通过校验和数据库对内容进行哈希匹配。

你可以想象一种简单的方法。

我们可以在某个服务器上运行一个可以根据请求提供 go.sum 文件的数据库。我们可以告诉社区我们会表现出来，并要求他们相信我们做正确的事情。

但是，真正做的就是解决问题.

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-61.png)

并把它移到其他地方。我们所要做的就是为攻击者创建一个不同的目标。

你还可以想象一个校验和数据库由她之前担心的那些猫人运行的场景。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-63.png)

使用简单的数据库，校验和数据库很容易开始针对狗。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-64.png)

他们可以为所有猫提供真实代码的校验和，但是为狗提供该代码的恶意版本的校验和。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-65.png)

审计整个数据库既困难又昂贵。

中间人攻击不容易被检测到，并且操纵数据不容易被客户端注意。

我们如果无法对它负责就不应该相信校验和数据库是真相的来源。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-66.png)

我们需要一种不会让校验和数据库行为不当，并且会使审计员和 go 命令容易检测到有针对性的攻击的解决方案.

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-67.png)

[透明日志和 Merkle 树研究 ( Transparent Logs and Merkle Trees Research )](https://research.swtch.com/tlog)

我们将把 go.sum 行存储在所谓的透明日志中。它是由散列节点对构建的树结构。

这与用于保护 HTTPS 的证书透明度技术相同。**（译者注：[ 证书透明度 wiki ](https://zh.wikipedia.org/wiki/%E8%AF%81%E4%B9%A6%E9%80%8F%E6%98%8E%E5%BA%A6)）**

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-68.png)

如果你有花比较多的时间与密码学的朋友在一起的话，你可能也有听说过梅克尔树（ merkle tree ）。

我们使用梅克尔树代替简单的数据库来作为我们的真实来源，因为梅克尔树更值得信赖。它主要的优点就是它具有防篡改功能。

它具有不允许执行未发现的不当行为。**（译者注：具有可执行的行为白名单？）**

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-70.png)

将记录放入日志后，永远不会修改或删除它。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-71.png)

如果日志中的单个记录发生更改，则哈希将不再排列，并且可以立即检测到。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-72.png)

因此，如果 Go 命令可以证明它将要添加到你模块的 go.sum 文件中的行在此透明日志中，那么可以非常确信这些将要添加到你 go.sum 文件的行是正确的。

请记住，我们的目标是确保每个人每次都从代理服务器或源服务器获取“正确”的模块版本。

从我们和 go 命令的角度来看，“正确”意味着......

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-74.png)

“就像昨天和之前的每一天一样，对于每一个要求它的人来说都是如此”，所以我们拥有一个无法篡改的数据结构非常重要。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-75.png)

此日志提供了一种非常可靠的方法来向审计员和 go 命令证明两个关键事项：

1. 通过称为“包含证明”的东西在日志中存在特定记录。
2. 那棵树没有被篡改过。具体来说，后面的树包含我们已经知道的旧树，称为“一致性证明”。

在针对日志中的一组 go.sum 行进行验证时，这两个证明可疑给 Go 命令置信度。在将新的 go.sum 行添加到模块的 go.sum 文件之前 Go 命令会动态执行这样的校验。

我们希望社区将会有一些外部审计员，因为这些审计员对该系统的工作至关重要。他们可以一起工作，观察和探讨梅克尔树的变化，以警惕任何可疑行为。

我不会在这次演示文稿中介绍所有的加密技术，但会稍微谈一点，让你有一个大概的了解，知道它是如何工作的。嘿，谁不爱一点点数学 :)

让我们一起了解一下包含证明

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-76.png)

让我们从这个数据结构的实际内容以及它的构建方式开始。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-77.png)

透明度日志的基础是 go.sum 行，由此图像中的绿色框表示。

让我们首先假设我们的日志目前有 16 条记录。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-78.png)

例如，记录 0 是 go.sum 行 go.opencensus.io v0.19.2 。这些是日志中此模块版本的唯一 go.sum 行，以及校验和数据库应该提供的唯一 go.sum 行。这是审计师可以负责验证的事情。

从这里开始，我们可以开始创建树的其余部分。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-79.png)

我们对每个记录的 go.sum 行执行 SHA-256 哈希，并将它们存储为 0 级节点。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-80.png)

然后我们将 0 级节点对混合在一起以创建 1 级节点。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-81.png)

然后散列 1 级节点对以创建 2 级节点。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-82.png)

然后哈希 2 级节点对以创建 3 级节点。
![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-83.png)

直到最后，我们最终在这个例子中的第 4 级，在树的顶部有一个哈希，我们称之为树头。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-84.png)

那么为什么要创建哈希呢？

那么，还记得她谈过的那些证据吗？它们基本上归结为比较这个顶级树头或哈希，看看它是否与你计算的那些以及你之前看过的那些匹配。

让我们来看一个包含证明的例子，它只需要几个哈希就能工作。证明树中包含一组 go.sum 行是 go 命令如何验证我们刚刚从代理返回的源代码。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-85.png)

假设我们要验证 go.dog/breeds 的 g o.sum 行，它们是我们刚刚从代理获取的版本 0.3.2 ，在本例中，恰好是记录 9 。

我们要做的第一件事就是使用我们的模块版本在校验和数据库中命中一个名为 “ lookup ” 的端点，它给了我们 3 个信息

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-87.png)

* 在日志中标识它的唯一记录号，在本例中为 9
* go.dog/breeds 在 go.sum 记录的版本号为 0.3.2
* 包含此记录的树头

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-88.png)

为了证明记录 9 存在于树中，我们需要形成从叶子到头部的路径，并确保它与我们给出的头部一致。如果我们在此示例中从级别 0 开始按级别向上走，那将是节点 9 , 4 , 2 , 1 和 0 （级别 4 的头部）。

go 命令可以通过散列它给出的 go.sum 行来创建 0 级节点 9 ，但它需要更多的节点才能创建其余的路径。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-89.png)

这里，为了计算节点 4 的散列，我们需要将节点 8 和 9 一起散列。然后，我们可以在节点 4 使用新创建的散列并将其与节点 5 一起散列以创建节点 2 ，依此类推，直到我们计算出级别 4 的哈希。

如果我们刚刚在树顶创建的级别 4 的哈希与我们从 lookup 端点获取到的树头是一致的，那么我们就完成了包含证明校验，并验证了我们正在查询的 go.sum 行存在于我们的校验和数据库中，流程结束。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-90.png)

随着这棵树的增长，你会得到新的树头，你应该检查这些新树是旧树的超集。所以 Go 命令会在发现它们时存储这些树头，进行一致性验证以验证它刚刚找到的新树头是否与之前知道的旧树头一致，以确保树没有被篡改。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-91.png)

通常情况下，树的大小不会是 2 的幂，但我们仍然希望能够在这些情况下进行包含证明。那仍然是可能的！

在这个例子中，我们的树中有 13 条记录，其中包括一些 “ 临时 ” 节点，在此图中用 x 标记。

唯一的区别是我们从叶子到树头的路径包含一些我们动态创建的临时节点，当我们完成它们时可以丢弃它们。

这就是包含证据的全部内容。在实践中，我们需要弄清楚的最后一件事是如何从校验和数据库中将这些内部节点用蓝色圈起来，以便进行我们的证明。

这些内部节点是通过称之为『 tile 』新的方式进行存储和服务。

在下图中，校验和数据库将这棵树分成了一块一块，称之为『 tile 』。每个 tile 包含一组散列节点，可用于证明并由客户端访问。

在这个例子中，我们选择了一个 2 的 tile 高度，这意味着在树的每两个级别创建一个新的 tile 。实际的校验和数据库树比这大得多，因此在实践中使用高度为 8 的 tile 。

我们知道，这个证明需要的节点之一是级别 2 的节点 3 ，就像之前一样。

此节点包含在 tile(1,0) 内，因此在执行证明时， Go 命令将从校验和数据库中请求该节点。

使用 tile 有一些很大的好处。

对于校验和数据库服务器来说，tiles 很不错，因为它们在前端非常适合缓存。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-96.png)

但它对客户端也很好，因为它们只缓存每个 tile 的底行，并从中构建任何必要的中间节点。我们选择的 tile 高度为 8 ，因此可以降低你的存储成本。 Go 命令不是缓存整个树，而是缓存树中的每个第 8 级，并根据需要动态构建内部节点。

现在你已经看过一些数学，让我们回到它在 Go 的上下文中是如何工作的。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-97.png)

通过你在此处看到的校验和数据库规范，此树可用于 Go 命令。 它使用 lookup 和 tile 端点来检索我们刚才谈到的数据。

还有一个额外的端点 /latest ，它为校验和数据库创建的最新树头提供服务。它仅用于审计员根据提供的越来越大的 STH 逐步验证记录。

签名的树头看起来像这样：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-98.png)

它告诉你这个树头的树的大小，以及它的哈希值。 在此示例中，树大小为 11,131 。

底部是签名，其中包含校验和数据库的名称 sum.golang.org ，后面是该树头的唯一签名。这个签名很重要，因为它允许审计人员轻松地将责任归咎于 sum.golang.org，如果它服务了它不应该服务的。

让我们回到我们的代理示例。我们做的最后一件事是获取版本 0.3.2 的 go.dog/breeds 的 zip 。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-102.png)

在更新 go.sum 和 go.mod 文件之前，它将生成一个哈希值，然后检查这是否与校验和数据库具有相同的哈希值。

它将从查找该模块版本开始。

校验和数据库返回其记录号， go.sum 行和包含它的签名树头（或 STH ）。

根据此记录编号以及已通过 Go 命令在你的计算机上缓存和验证的记录和 tile ...

它现在可以开始请求 /tile 端点以获取其证明所需的切片。

一旦 Go 命令完成了它的证明，它就可以使用新的 go.sum 行更新模块的 go.sum 文件，我们就完成了！

现在，而不是世界上每个人都单独信任他们第一次下载模块，校验和数据库签名的第一个版本是唯一受信任的版本。 这确保了模块版本的源代码对于世界上的每个人都是相同的，因为有一个可信任的校验和源可以被验证和审计。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/module-life-of-query/go-module-proxy-life-of-a-query-103.png)

这一切都运行得很好，即使只有一个每个人都使用的校验和数据库。社区有办法让它负责， Go 命令也可以即时进行校对，验证校验和数据库是否未被篡改。

此校验和数据库使不受信任的代理成为可能。代理无法向你提供错误的代码，因为在任何源代码到达你之前，有一个可审计的安全层位于其前面。

如果原始服务器被黑客攻击，那么这将被立即捕获，因为我们有一个不可变的校验和，当我们第一次看到它时，它会识别内容。

即使是模块的作者也无法改变他们的标签移动，将与特定版本相关的位从一天更改为下一天。

对此非常好的是，作为开发人员，你不必做任何事情来完成这项工作。你无需在任何地方注册代码，也无需管理自己的私钥或在安全源中创建自己的哈希。 此校验和数据库为该代码创建唯一的哈希值，并将其永久存储在日志中。

重要的是要注意，此校验和不仅解决了代理创建的问题。它实际上创建了比直接源连接更安全的用户体验，因为它可以更好地保护您免受更改依赖关系和针对性攻击。

现在我们全面了解：依赖性，不会消失，并且每个人都可以信任的依赖关系！

幸运的是，如果你对这些感兴趣，那么她对你有好消息。 Katie 和她的同事在团队中构建了一个模块镜像和校验和数据库，您可以立即开始使用它。

默认情况下，我们的模块镜像和校验和数据库由模块用户的 Go 1.13 中的 Go 命令使用。如果您想立即开始使用它，只需升级即可使用 1.13 测试版。

在底层， Go 命令有一些可以配置的环境变量。

自 Go1.11 以来， GO111MODULE 和 GOPROXY 一直存在。

您可以将 GO111MODULE 设置为“on”以启用模块模式，或将其保留为“auto”。**（译者注： `export GO111MODULE=on` ）**

您可以将 GOPROXY 设置为您选择的代理，以便在模块模式下通过 Go 命令获取。尽管从 1.11 始就存在这种情况，但提供逗号分隔列表的能力对于 1.13 来说是新的。这告诉 go 命令在放弃之前尝试多个源。如果您想使用 Go 团队的模块镜像，可以将其设置为 https://proxy.golang.org 。**（译者注： `export GOPROXY=proxy.golang.org` ）**

代理和校验和数据库的本质是源代码必须在公共互联网上可用，因此每个人都可以对其进行审核和使用。但是，如果您正在使用私有模块，则可以通过将它们列在 GOPRIVATE 环境变量中来禁用要跳过的域的代理和校验和数据库。

她提到了 Go 团队用于透明度日志的开源项目。

他们使用 [Trillian](https://github.com/google/trillian) 来实现她之前描述的 merkle 树数据结构。 他们依靠他们的数据存储来保存 go.sum 行以及 Go 命令用于其证明的相应散列。

我们已经讨论过 proxy.golang.org 和 sum.golang.org ，但 Go 团队还提供了另外一项服务，即 Module 模块索引。

index.golang.org 是 proxy.golang.org 发现的新模块的简单提要（ https://proxy.golang.org ）。您可以在 index.golang.org/index 上看到此 Feed ，如果您只想查看比特定时间戳更新的模块，则可以选择提供 since 参数。**（译者注 https://index.golang.org/index?since=2019-04-10T19:08:52.997264Z ）**

Go 团队对模块的未来感到非常兴奋，通过创建可重现的构建为开发人员提供更好的依赖管理体验，确保依赖关系不会在一夜之间消失，并确保您要求的源代码是源代码，你和世界上的其他人每次都会得到你。

她个人很开心，因为现在她有一套可以帮助她建立模块的解决方案......

......这将使到处都是更快乐的小狗。:)

Go 团队计划对这些功能进行微调，他们希望您能够试用它们并在可能的时候给予反馈！他们很想知道镜像和校验和数据库是如何为您工作的，我们鼓励您在发现它们时在 Github 上提交问题。世界上的狗都感谢你。

Katie Hockman ，谷歌， Go 开源团队。

[github.com/katiehockman](https://github.com/katiehockman)

[@katie_hockman](https://twitter.com/katie_hockman)

内容由 Katie Hockman 的幻灯片提供。

---

via: https://about.sourcegraph.com/go/gophercon-2019-go-module-proxy-life-of-a-query

作者：[Royce Miller](https://github.com/r0yce)
译者：[ZackLiuCH](https://github.com/ZackLiuCH)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
