# 为什么要使用 go module proxy

在介绍 **Go module** 之后，我以为我已经知道了我需要知道的一切。很快，我意识到并非如此。最近，人们开始提倡使用 **Go module proxy**。在研究了利弊之后，我得出结论，这是近年来**最重要**的变化之一。
但为什么会这样呢？是什么让 **Go module** 代理如此特别？

使用 **Go modules** 时，如果向干净的缓存的计算机上添加新依赖项或构建 **Go module**，它基于 **go.mod** 将下载（go get）所有依赖项，并将其缓存以供进一步操作。
你可以通过使用 vendor 文件夹绕过缓存（以及使用该下载依赖项）并构建使用带有 -mod=vendor 标志的此文件夹。

但这两种方法都不完美，我们可以做得更好。

## 是否使用 vendor 文件夹

### 使用 vendor 文件夹的坏处

* 在go命令用于该命令（在[模块感知模式](https://golang.org/cmd/go/#hdr-Modules_and_vendoring)下），vendor 默认情况下将不再使用。
  如果你不附加 -mod=vendor 命令行参数，它将不会被启用。 这通常引发问题，并导致必须使用其他陈旧的方案来支持老的 Go 版本（请参考：[在Travis CI 上使用 Go Module 和vendor](https://arslan.io/2018/08/26/using-go-modules-with-vendor-support-on-travis-ci/)）
* vendor 文件夹，特别是在比较大的单体应用中，会占据大量空间。这也将增加仓库的克隆时间。可能你认为只用克隆一次，实际却不是这样。CI/CD 在每次事件（例如：pull request ）触发常常都会克隆代码。因此，这将长期导致更长的编译时间，并将影响每一个人。
* 提供新的依赖关系通常会导致难以审核代码的变化。大多数情况下，你须将依赖项与实际业务逻辑捆绑在一起，这使得很难正确实现变化。

### 不使用的 vendor 的坏处

* **go** 将去源码仓库下载这些依赖。总是存在任何依赖可能在将来消失的风险（[记住左边的传奇故事](https://qz.com/646467/how-one-programmer-broke-the-internet-by-deleting-a-tiny-piece-of-code/)）。
* 版本管理系统（例如 github.com ）可能已关闭。在这种情况下，你将无法再构建项目。
* 有些公司不希望内网接入外网，此时，没有 vendor 文件夹，我们将无法使用。
* 假设发布的依赖 tag 是 v1.3.0 ，并且已经 go get 获取它到本地缓存。此时，依赖的所有者可以通过推送具有相同 tag 的恶意内容来破坏代码库。
  如果在具有干净缓存的计算机上重建**Go module**，它现在将使用受损包。 为了防止这种情况，需要将 go.sum 文件存储在文件旁边 go.mod 。
* 一些依赖使用不同的 **版本管理系统**，比如不只是 git，还有 hg(Mercurial)，bzr(Bazaar)或svn(Subversion)。
  而你的机器（或Dockerfile）没有装这些工具，这将带来问题。
* **go get** 需要获取 **go.mod** 列出的每个依赖项的源代码来解决传递依赖（需相应的go.mod文件）。这显着减慢了整个构建过程，因为它意味着必须下载（例如 git clone ）每个存储库以[获取单个文件](https://about.sourcegraph.com/go/gophercon-2019-go-module-proxy-life-of-a-query)。

**我们怎么改善这些情况呢？**

## 使用 go module proxy 的好处

默认情况下， go 命令会直接从版本管理系统下载代码。
**GOPROXY** 环境变量允许在下载源的进一步控制。配置该环境变量后，go命令可以使用 **Go module proxy**。

设置环境变量 **GOPROXY** 开启 **Go module proxy** 后，将解决上边提到的所有问题。

* **Go module proxy** 默认永久缓存所有依赖（不可变存储）。这意味着，不必再使用 vendor 文件夹。
* 抛弃 vendor 文件夹，它将不会再消耗代码库的空间。
* 因为依赖项存储在不可变存储中，即使依赖项从Internet上消失，你也会受到保护。
* 一旦 **Go module** 存储在 **Go proxy** 中，就无法覆盖或删除它。这可以保护你免受可能使用相同版本注入恶意代码的攻击。
* 你不再需要任何 VSC 工具来下载依赖项，因为依赖项是通过HTTP提供的（ **Go proxy** 使用HTTP）。
* 下载和构建 **Go module** 的速度要快得多，因为 **Go proxy** 通过HTTP独立提供源代码（.zip存档）go.mod。与从 VCS 获取相比，这使得下载花费更少的时间和更快（由于更少的开销）。
  解决依赖关系也更快，因为 go.mod 可以独立获取（而在它必须获取整个存储库之前）。Go 团队对它进行了测试，他们看到快速网络上的速度提高了 3 倍，而慢速网络则提高了 6 倍！
* 你可以轻松运行自己的 **Go proxy** ，这可以让你更好地控制构建管道的稳定性，并防止 VCS 关闭时的罕见情况。

如你所见，使用**Go module proxy** 对每个人来说都是一个胜利。但是我们如何使用它呢？如果你不想维护自己的**Go module proxy**怎么办？让我们看看许多替代选择。
  
## 如何使用 go module proxy

要开始使用**Go module proxy**，我们需要将GOPROXY环境变量设置为兼容的**Go module proxy**。有多种方式：

1. 如果没有设置 GOPROXY，将其设置为空或设置为 direct ，然后 go get 将直接到 VCS（例如github.com）拉取代码：

   ```bash
   GOPROXY=""
   GOPROXY=direct
   ```

    它也可以设置为off，这意味着不允许使用网络

    ```bash
    GOPROXY=off
    ```
  
2. 你可以开始使用公共 GOPROXY 。你可以选择使用 Go 团队的 GOPROXY（由 Google 运营）。更多信息可以在这里找到：https：//proxy.golang.org/

   要开始使用它，你只需设置环境变量：

    ```bash
    GOPROXY=https://proxy.golang.org
    ```

    其他公共代理：

    ```bash
    GOPROXY=https://goproxy.io
    GOPROXY=https://goproxy.cn # proxy.golang.org 被墙了, 这个没有
    ```
  
3. 你可以运行多个开源实现并自己托管。可用的有：

    1. Athens：https：//github.com/gomods/athens
    2. goproxy：https：//github.com/goproxy/goproxy
    3. THUMBAI：https：//thumbai.app/

    你需要自己维护这些。如果你想通过公共互联网或内部网络提供服务，这取决于你。

4. 你可以购买商业产品：

    Artifactory：https：//jfrog.com/artifactory/

5. 你可以传递 file:///URL 。因为**Go module proxy**是响应 GET 请求（没有查询参数）的 Web 服务器，所以任何文件系统中的文件夹都可以用作**Go module proxy**。

## 即将到来的 Go 1.13 的变化

在 Go v1.13 版本中， **Go proxy** 会有一些变化，我认为应该强调一下：

1. 在 GOPROXY 环境变量现在可以设置为逗号分隔的列表。它会在回到下一个路径之前尝试第一个代理。
2. 默认值 GOPROXY 为 <https://proxy.golang.org,direct>。设置 direct 后将忽略之后的所有内容。这也意味着**go get**现在将默认使用 GOPROXY 。如果你根本不想使用 GOPROXY，则需要将其设置为 off。
3. 将引入了一个新的环境变量 GOPRIVATE ，它包含以逗号分隔的 全局列表。这可用于绕过 GOPROXY 某些路径的代理，尤其是公司中的私有模块（例如: **GOPRIVATE=*.internal.company.com**）。

所有这些变化都表明Go模块代理是Go模块的核心和重要部分

## 总结

无论使用公共网络，还是专用网络， **GOPROXY** 都有很多优势。这是一个很棒的工具，他可以和 go 工具无缝协作。
鉴于它具有如此多的优势（安全，快速，存储效率），因此为你的项目或组织快速采用它是明智之举。此外，使用Go v1.13，它将默认启用，
这将迎来依赖管理新的进步。

---

via: <https://arslan.io/2019/08/02/why-you-should-use-a-go-module-proxy/>

作者：[Fatih Arslan](https://arslan.io/)
译者：[TomatoAres](https://github.com/TomatoAres)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
