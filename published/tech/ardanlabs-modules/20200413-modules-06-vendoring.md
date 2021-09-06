首发于：https://studygolang.com/articles/35227

# Go Module 教程第 5 部分：Vendoring

前四个教程：

- [Go Module 教程第 1 部分：为什么和做什么](https://studygolang.com/articles/24580)
- [Go Module 教程第 2 部分：项目、依赖和 gopls](https://studygolang.com/articles/35202)
- [Go Module 教程第 3 部分：最小版本选择](https://studygolang.com/articles/35210)
- [Go Module 教程第 4 部分：镜像、校验和以及 Athens](https://studygolang.com/articles/35225)

作者是一个 Vendoring 爱好者。觉得它合理和实用，可以用于你的应用程序项目。

## 引言

我相信 Vendoring 给你的应用程序项目最持久的稳定保证，因为该项目拥有每一行源代码，它需要构建应用程序。如果你想要一个可重复的构建，而不需要依赖外部服务（比如模块镜像），并且不需要连接到网络，那么 vendoring 就是解决方案。

Vendoring 的其他好处有：

- 如果从 VCS 中删除了依赖项，或者代理服务器丢失了模块，Vendoring 同样可以正常工作；
- 通过运行 diff 可以看到升级依赖关系，并维护历史记录；
- 你将能够跟踪和调试依赖项，并在必要时测试更改；
  - 一旦你运行 `go mod tidy` 或 `go mod vendor`，更改就会覆盖掉；

在这篇文章中，我将提供 Go 支持 Vendoring 的历史，以及随着时间的推移默认行为的改变。我还将分享 Go 的工具如何能够维护版本间的向后兼容性。最后，我将分享可能需要(随着时间的推移)手动升级 go.mod 文件中列出的版本，以更改未来 Go 版本的默认行为。

## 运行不同版本的 Go

为了向你展示 Go 1.13 和 Go 1.14 之间默认行为的差异，我需要能够同时在我的机器上运行这两个版本的工具。在我发表这篇文章的时候，我已经在我的机器上安装了 Go 1.14.2，我使用传统的 `go` 访问这个版本。但是对于这篇文章，我还需要运行一个 Go 1.13 环境。那么，如何才能在不破坏我目前的开发环境的情况下做到这一点呢？

关于这点，「polarisxu」公众号发表两篇相关的文章：

- [终于找到了一款我喜欢的安装和管理 Go 版本的工具](https://mp.weixin.qq.com/s/yTblk9Js1Zcq5aWVcYGjOA)
- [我这样升级 Go 版本，你呢？](https://mp.weixin.qq.com/s/jEhX5JHAo9L6iD3N54x6aA)

推荐大家参考使用。

## Vendoring 快速参考

Go 工具在管理和 vendor 应用程序项目的依赖关系方面做了很好的工作，最小化了对工作流的影响。它有两个子命令： `tidy` 和 `vendor`。

`go mod tidy` 命令可以保证项目的依赖准确的列出。有些编辑器(比如 VS Code 和 GoLand)提供了在开发期间更新模块文件的功能，但这并不意味着一旦一切正常工作，模块文件就会变得干净和准确。我建议在提交代码之前运行 `go mod tidy` 命令，并将代码 push 到 VCS。

如果你也想 vendor 这些依赖项，那么在 tidy 之后运行 vendor 命令。

```bash
go mod vendor
```

此命令在项目中创建一个 vendor 文件夹，其中包含项目构建和测试代码所需的所有依赖项(直接和间接)的源代码。在运行 tidy 之后应该运行此命令，以保持 vendor 文件夹与模块文件同步。确保提交并将 vendor 文件夹推送到 VCS。

## 版本间的向后兼容性

在 Go 1.14中，在模块缓存上默认使用 vendor 文件夹的更改是我希望项目的行为。起初我以为我可以用 Go 1.14 来构建我现有的项目，这就足够了，但是我错了。在我第一次使用 Go 1.14 构建并且没有看到 vendor 文件夹后，我了解到 Go 工具读取 go.mod 文件以获取版本信息，并且保持与列出的版本的向后兼容性。其实在 Go 1.14的发布说明中清楚地表达了这一点。<https://golang.org/doc/go1.14#go-command>。

当主模块包含顶级 vendor 目录并且它的 go.mod 文件指定 Go 1.14 或更高时，Go 命令现在默认为 `-mod=vendor` 来执行接受该标志的操作。

为了使用新的默认 vendoring 行为，我需要将 go.mod 文件中的版本信息从 Go 1.13 升级到 Go 1.14。

## GOPATH 或 Module 模式

在 Go 1.11 中，向 Go 工具添加了一个新的模式，称为“模块模式”。当 Go 工具以模块模式运行时，模块系统用于查找和生成代码。当 Go 工具以 GOPATH 模式运行时，传统的 GOPATH 系统将继续用于查找和构建代码。我在使用 Go 工具时遇到的一个更大的问题是，知道不同版本之间默认使用什么模式。然后知道哪些配置更改和标志需要保持构建的一致性。

为了理解 Go 过去 4 个版本的历史和语义变化，最好重温一下这些模式。（截止 2021 年 9 月，已经发布了 Go1.17）

**Go 1.11**

引入了一个叫做 GO111MODULE 的新环境变量，它的默认设置是 auto。这个变量将决定 Go 工具是使用模块模式还是 GOPATH 模式，具体取决于代码所在的位置(GOPATH 内部或外部)。若要强制一种模式或另一种模式，您可以将此变量设置为 on 或 off。当涉及到 vendor 文件夹时，模块模式将默认忽略 vendor 文件夹，并建立对模块缓存的依赖关系。

**Go 1.12**

GO111MODULE 的默认设置仍然是 auto，Go 工具继续根据代码所在的位置(GOPATH 内部或外部)确定模块模式或 GOPATH 模式。对于 vendor 文件夹，模块模式在默认情况下仍然会忽略 vendor 文件夹，并依赖于模块缓存。

**Go 1.13**

GO111MODULE 的默认设置仍然是 auto，但是 Go 工具不再对工作目录是否在 GOPATH 中敏感。模块模式在默认情况下仍然会忽略 vendor 文件夹，并依赖模块缓存生成依赖项。

**Go 1.14**

GO111MODULE 的默认设置仍然是 auto，Go 工具不再对 GOPATH 中是否包含工作目录敏感。但是，如果存在 vendor 文件夹，默认情况下将使用它来构建依赖项，而不是构建模块缓存。此外，go 命令验证项目的 vendor/modules.txt 文件与它的 go.mod 文件是否一致。

从 Go1.16 起，GO111MODULE  默认值设置为 on，即默认启用 Module 模式。

## 总结

不知道大家用 Vendor 多不多？如果依赖很多，Vendor 似乎是比较好的选择。像 Kubernetes 使用的就是 Vendor。

本文只是简单的介绍了 Vendoring，毕竟使用很简单，可以根据你的情况来决定。

没有完全按照文章翻译。

---

via: <https://www.ardanlabs.com/blog/2020/04/modules-06-vendoring.html>

作者：[William Kennedy](https://www.ardanlabs.com/)
译者：[polaris1119](https://github.com/polaris1119)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
