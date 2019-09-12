# GopherCon 2019 - Go Module Proxy: Life of a query
GopherCon_2019/Go_Module_Proxy_Life_of_a_query
## 概述

Go 团队已经搭建了模块镜像与校验和数据库，这将提升 Go 生态环境的可靠性与安全性。这次的交流会通过 go 命令、代理与校验和数据库讨论经过身份验证的模块代理的技术细节。


## 介绍

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

这样代码就必须依赖大量的 API。她需要依赖存储、音频处理和地图API。

因此产生了忧虑：
1. 她担心她会因不可重现的构建而面临挑战。她依赖她正在使用的地图服务的全新 API 端点呢。如果引入她的包的人使用的是旧版本的 API ，那么他们的程序就可能构建失败，而且她无法保证他们的构建是一致的。
2. 她所依赖的音频检测包的拥有者可能因为厌倦了管理这些代码，然后在一夜之间将代码从 github 提走。这样她在构建的时候就会失败，因为她丢失了依赖包，这样她的构建就会卡住。
3. 更糟糕的情况是，有不怀好意的人试图做出不轨的行为，攻击保存了她所依赖的源代码的服务器。因此使得她获取到错误的代码导致她的包失去可靠性。这些恶意的代码对依赖她代码的人来说可能具有危险性，当她弄清楚情况时，可能为时已晚！


为了保护自己依赖她代码的人，她想到了一些方案和问题。
1. 也许可以完全停止依赖。但是真的很难，因为这样她必须从头编写一堆代码，她还不可以复制她所依赖的一些 API ，因此她的代码质量会收到影响。
2. 她可以在她的代码里提供所有的依赖，但是这样子会造成她代码库的大小，并且她担心随着时间的推移很难维护和更新。
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

![ puppies 模块](https://d33wubrfki0l68.cloudfront.net/3b8bb9793d5677cecea02a83c7cc855ca2a1d41e/8aada/gophercon-2019/go-module-proxy-life-of-a-query-9.png)

模块版本有一个主版本号，次版本号和修订号来组成它的语义版本。

![语义版本]](https://d33wubrfki0l68.cloudfront.net/03a164b7facd35432ff6aebea65826dfe8bc4e48/13c23/gophercon-2019/go-module-proxy-life-of-a-query-10.png)

如果你想在模块存在之前导入包，你要么建一个 vendor 库保存你引用的包的源代码，或者依赖最新版本的包。

现在，包可以放置在模块里面，该模块是及时版本化的快照，作为唯一标识。版本的内容永远不允许改变。

![包可以放置在模块里面，不可改变且向后兼容](https://d33wubrfki0l68.cloudfront.net/a4fd7c96aa98b485e42d5b2969701ce6dc71468d/0e26b/gophercon-2019/go-module-proxy-life-of-a-query-11.png)

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

![ go.mod ](https://d33wubrfki0l68.cloudfront.net/87cd500223cf430ca5e1e73065d9be89cfe2dcc9/d22f5/gophercon-2019/go-module-proxy-life-of-a-query-13.png)

通过指定你的代码依赖于 1.4.1 或更高版本的模块，你就可以保证每个引入你包的人将永远不会使用 1.4.1 之前的版本。**（译者注：就是 go.mod 指定你模块的依赖项是某个特定的版本，比如 1.4.1 ，那么引入你模块的人在使用模块封装的方法的时候，所使用依赖项不会是 1.4.1 之前的版本）**

go 命令使用称为『最小版本选择』（ minimal version selection ）来构建模块，该模块基于 go.mod 文件中指定的版本。

举个例子，假设我们有模块 A 和模块 B ，都是 github.com/katiehockman/puppies 的依赖项。

![ 『最小版本选择』例子 ](https://d33wubrfki0l68.cloudfront.net/fd687600f0106a4477d11c4d2015ab513d4d23da/0326f/gophercon-2019/go-module-proxy-life-of-a-query-15.png)

每一个模块都有依赖项 C ，但版本不一样。 A 要求 C 的最低版本是 1.5 ，而 B 要求 C 的最低版本是 1.6。

C 还发布了另一个更新的版本： 1.7 。

如果她的模块依赖模块 A 和 B ，那么当她进行构建的时候 go 命令会选择同时满足 A 和 B 的 go.mod 文件约束的最小版本。在这个例子里，将选择 1.6 版本。**（译者注：A 要求最低版本是 1.5 ，而 B 要求最低版本是 1.6 ，因为模块的向后兼容性，所以这里 1.6 版本 理应兼容 1.5 版本， 所以 go 命令会选择 1.6 版本）**

![ 『最小版本选择』例子 ](https://d33wubrfki0l68.cloudfront.net/f20f0e2b761204b9dea3f1f20a6e140c13df3e59/76ac2/gophercon-2019/go-module-proxy-life-of-a-query-16.png)

这个最小版本选择保证了构建的可重复性（ reproducible builds ）， 因为 go 命令构建的时候用的是每个依赖项的允许的最小版本，而不是每天可能更改的最新版本。**（译者注：如果选择的版本不能兼容模块 A 和 B 的要求，可能导致其中一方的代码错误，比如：某个函数的不存在，导致构建失败。所以我们在编写模块的时候，也要注意在更新代码的时候尽量不要删除旧的方法或常量，否则会破坏模块的可向后兼容性。这里有个词语：<u>reproducible builds</u> ，可以点击[这里](https://reproducible-builds.org)了解一下）**

现在我们就拥有了一致的，可重复的构建方案，而不再依赖于依赖项的最新提交历史。

一个问题解决了，就着手第二个问题。现在我们完成了可重复构建，接下来让我们讨论如何确保我们的依赖关系不会消失。

## 更好的解决方法：模块镜像与代理

现在我们的模块已经版本化，对应版本的内容是固定的，不允许改动的，我们可以将它缓存下来并做验证，而这是非模块模式没有的功能。

这里是模块代理进入的地方。

![ 模块代理 ](https://d33wubrfki0l68.cloudfront.net/e67f4ebbdfc5186fa15e1abbb611fd75a269cda4/fdd2b/gophercon-2019/go-module-proxy-life-of-a-query-19.png)

上图就是没有代理的流程。很简单，是吧？当 Go 开发者执行 go 命令时，比如： `go get` ， 当本地缓存没有对应的包的时候， go 命令会直接访问源服务。这里通常指的是托管 Go 源代码的地方，比如： Github 。

`go get` 的行为是否更改是根据 go 命令是否在 module-aware 模式或 GOPATH 模式下运行。代理仅在 module-aware 模式下运行，这我们一会讨论。

我们将讨论 `go get` 两个主要的工作：
* 请求并拉取你请求的源码
* 基于你拉取的源码感知你的模块是否有新的依赖。它必须解决这些依赖关系，并可能需要更新你的 go.mod 文件

但是没有代理，这过程的代价可能变得非常昂贵，非常快。延迟和系统存储都是如此。**（译者注：快 ==> 慢？）**

go 命令可能会强制拉取整个源代码，即使现在不会去构建它，只是为了在解析期间做出关于依赖版本的一致决定。**（译者注：最小版本选择？）**

所以对于依赖解析，这就时 go 命令获取的内容：
![](https://d33wubrfki0l68.cloudfront.net/2b79ddbbdce64e2d15e4bc5eddc05596eca46a4d/6f4ad/gophercon-2019/go-module-proxy-life-of-a-query-21.png)

但是这才是它实际需要的：
![](https://d33wubrfki0l68.cloudfront.net/22d94b04c02e26d79c596692e6edf9ab57ccfc96/a323b/gophercon-2019/go-module-proxy-life-of-a-query-22.png)

请记住， go 命令唯一需要理解模块依赖关系的是一个 go.mod 文件。因此，对于 20 MB 的模块，只需要几 KB 的 go.mod 来执行此依赖解析。有许多缓存在你系统上的数据是用不到的，并且还浪费你大量的时间去请求源服务与拉取，这本是不必要的行为。

对于那些一致没有使用模块代理的人来说，这就是你看到的一些延迟背后的原因。这就是模块代理的用武之地。

回到我们的示例，使用 go 命令获取源代码，让我们将模块代理放入图片中。

![](https://d33wubrfki0l68.cloudfront.net/ceca131441096d2424b8b1c43276fbb2a8b06fd0/e57dd/gophercon-2019/go-module-proxy-life-of-a-query-24.png)

如果你告诉 go 命令使用代理，就不是像以前那样直接向源服务器发起请求，它将会向代理请求它想要的东西。

