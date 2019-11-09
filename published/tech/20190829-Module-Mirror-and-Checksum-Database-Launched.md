首发于：https://studygolang.com/articles/24556

# Go1.13 推出模块镜像和校验和数据库(Module Mirror and Checksum Database Launched)

我们很高兴地分享我们的模块 [镜像](https://proxy.golang.org/) ，[索引](https://index.golang.org/) 和 [校验和数据库](https://sum.golang.org/) 现已准备就绪！ 对于 [Go 1.13 模块用户](https://golang.org/doc/go1.13#introduction) ，go 命令将默认使用模块镜像和校验和数据库。 有关这些服务的隐私信息，请参阅 [proxy.golang.org/privacy](proxy.golang.org/privacy) ，有关配置详细信息，请参阅 [go 命令文档](https://golang.org/cmd/go/#hdr-Module_downloading_and_verification) ，包括如何禁用这些服务器或使用不同的服务器。 如果您依赖于非公共模块，请参阅 [配置你的环境的文档](https://golang.org/cmd/go/#hdr-Module_configuration_for_non_public_modules) 。

这篇文章将描述这些服务及使用它们的好处，并总结了 Gophercon 2019 提到 [Go Module Proxy: Life of a Query](https://www.youtube.com/watch?v=KqTySYYhPUE&feature=youtu.be) 的一些要点。如果您对完整的演讲感兴趣，请参阅 [录制内容](https://www.youtube.com/watch?v=KqTySYYhPUE&feature=youtu.be) 。

## 模块镜像(Module Mirror)

模块是一组版本化的 Go 包，每个版本的内容都是不可变的。这种不变性为缓存和身份验证提供了新的机会。当以模块模式运行时，它必须获取包含所请求的包的模块，以及该模块引入的任何新依赖项，根据需要更新 go.mod 和 go.sum 文件。从版本控制中获取模块在系统的延迟和存储方面可能是昂贵的：go 命令可能被迫下载包含传递依赖的存储库的完整提交历史记录，即使是未构建的存储库，也只是解决它的版本。

解决方法是使用模块代理，它代表一种更适合 go 命令需求的 API（参考 Go help goproxy ）。当使用代理以模块模式运行时，它只需要请求指定的模块元数据或源代码，所以它可以更快地工作，而不用担心其余部分。下面是一个示例，说明 go 命令如何通过请求版本列表来获取代理，然后使用最新标记版本的 info，mod 和 zip 文件。
![an example of how the Go command may use a proxy](https://blog.golang.org/module-mirror-launch/proxy-protocol.png)

模块镜像是一种特殊的模块代理，它将元数据和源代码缓存在自己的存储系统中，允许镜像继续提供原始位置不再提供的源代码。这可以加快下载速度并防止因为代码更迭导致的依赖关系丢失。有关更多信息，请参阅 [2019 年的 Go Modules](https://blog.golang.org/modules2019) 。

Go 团队维护一个模块镜像，在 [proxy.golang.org](proxy.golang.org) 上提供，模块用户从 Go 1.13 开始默认使用这个模块镜像。 如果您运行的是早期版本的 go 命令，则可以通过在本地环境中设置 `GOPROXY=https://proxy.golang.org` 来使用此服务。

## 校验和数据库(Checksum Database)

模块引入了 go.sum 文件, 该文件保存首次下载时 go.mod 文件下的依赖项和每个依赖项的源代码 SHA-256 。go 命令可以使用哈希去检测原始服务器或代理是否有提供给你相同版本但代码不同的不当行为。

go.sum 的局限性在于它完全信任你第一次拉取的代码。当你添加新的依赖到你的模块时（可能通过升级现有的依赖），go 命令将获取代码和动态地将依赖添加到 go.sum 文件中。问题是那些 go.sum 行没有被别人检查：他们可能与 Go 命令刚刚为其他人生成的 go.sum 行不同，可能是因为代理故意提供针对你的恶意代码。

Go 的解决方案就是将 go.sum 的每一行记录的全局源，称为校验和数据库，它确保 Go 命令总是向每个人的 go.sum 文件添加相同的行。不论什么时候，go 命令接受新的源码，它可以通过全局数据库校验代码的哈希值来确保哈希值是否匹配，以此保证每个人使用相同版本是相同的代码。

[sum.golang.org](sum.golang.org) 校验和数据库提供了校验和数据库，并构建在由 [Trillian](https://github.com/google/trillian) 支持的哈希的 [透明日志](https://research.swtch.com/tlog)（或"Merkle 树"）。 Merkle 树的主要优点就是它具有防篡改功能，并且具有不允许未被发现的不良行为的属性，这使得它比简单的数据库更可靠。go 命令使用树来检查『包含』证明（日志中存在特定记录）和『一致性』证明（树未被篡改），然后将新的 go.sum 行添加到模块中。下面是这种树的样子：
![tree](https://blog.golang.org/module-mirror-launch/tree.png)

校验和数据库支持一系列端点给 go 命令请求和校验 go.sum。 `/lookup` 端点提供 "signed tree head"（STH）和请求 go.sum 行。`/tile` 端点提供称为 tiles 的树的块，go 命令可以使用它来进行校样。下面是 Go 命令如何通过执行 `/lookup` 模块版本，然后证明所需的 tiles 来与校验和数据库交互的示例。
![how the Go command may interact with the checksum database](https://blog.golang.org/module-mirror-launch/sumdb-protocol.png)

如果你在使用 Go 1.12 或更早的版本，你可以手动敲入 gosumcheck 检查校验和数据库中的 go.sum 文件：

```
$ go get golang.org/x/mod/gosumcheck
$ gosumcheck /path/to/go.sum
```

除了通过 Go 命令执行校验外，第三方审计员还可以通过迭代日志来查找错误条目。他们可以一起工作，闲聊树的状态，以确保它保持不受影响，我们希望 Go 社区能够运行它。

## 模块索引(Module Index)

[index.golang.org](index.golang.org) 提供了 [proxy.golang.org](proxy.golang.org) 可用的模块索引服务，对于希望保留他们的可用缓存开发者来说特别有用，或者保持一些人们正在使用的模块是最新的版本。

## 反馈与建议

我们希望这些服务可以提升您使用模块时的体验，如果在使用过程中遇到问题，希望获得您的反馈或建议。

---

via: https://blog.golang.org/module-mirror-launch

作者：[Katie Hockman](https://twitter.com/katie_hockman) # 原文没有这名作者的链接，这个是我 google 搜出来的。
译者：[ZackLiuCH](https://github.com/ZackLiuCH)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
