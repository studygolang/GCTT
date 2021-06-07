# Go 模块镜像，索引与校验和数据库

Go团队提供由Google运营的以下服务：用于加速Go模块下载的**模块镜像**，用于发现新模块的**索引**，以及用于验证模块内容的全局 **go.sum 数据库**。

从 Go 1.13 开始，go命令将默认使用模块镜像和校验和数据库。 有关这些服务的隐私信息，请参阅 [proxy.golang.org/privacy](proxy.golang.org/privacy) ，有关配置详细信息，请参阅 [go命令文档](https://golang.org/cmd/go/#hdr-Module_downloading_and_verification) ，包括如何禁用这些服务器或使用不同的服务器。 如果您依赖于非公共模块，请参阅 [配置你的环境的文档](https://golang.org/cmd/go/#hdr-Module_configuration_for_non_public_modules) 。

## 服务

proxy.golang.org  - 模块镜像，符合 `go help goproxy` 提供的规范。

sum.golang.org  - 可审计的校验和数据库，将由 go 命令用于验证模块。 查看 [Secure the Public Go Module Ecosystem Proposal](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md) 以获取更多详细信息。

index.golang.org  - 索引，它为 proxy.golang.org 提供的新模块版本提供服务。 可以在 https://index.golang.org/index 查看源（ feed ）。 feed 作为新行分隔的JSON提供，提供模块路径（作为 Path ），模块版本（作为 Version ），以及proxy.golang.org（作为 Timestamp ）首次缓存的时间。 该列表按时间顺序排序。 有两个可选参数：
* since: 返回列表中模块版本的最早允许时间戳（ RFC3339 格式）。 默认是时间的开始，例如 https://index.golang.org/index?since=2019-04-10T19:08:52.997264Z
* limit: 返回列表的最大长度。 default = 2000，Max = 2000，例如 https://index.golang.org/index?limit=10

## 状态：已发布

这些服务已经可以在生产上使用。如果您发现问题，请提交问题，标题前缀为 proxy.golang.org: (或 index.golang.org 或 sum.golang.org ）。

## 部署环境

这些服务只能访问公开可用的源代码。如果你依赖四有模块，请将 GOPRIVATE 设置为覆盖它们的 glob 模式。请查阅 go 命令文档 [Module configuration for non-public modules](https://golang.org/cmd/go/#hdr-Module_configuration_for_non_public_modules) 以获取更多细节。

对于 Go 1.13 之前的版本，您可以通过设置 `GOPROXY = https://proxy.golang.org` 来配置 go 命令以使用此模块镜像下载模块。

要退出此模块镜像，可以通过设置 `GOPROXY = direct` 将其关闭。

有关其他配置详细信息，请参阅 [go命令文档](https://golang.org/cmd/go/#hdr-Module_downloading_and_verification) 。

## 校验和数据库

较旧版本的go命令无法直接使用校验和数据库。 如果您使用的是 Go 1.12 或更早版本，则可以使用 gosumcheck 针对校验和数据库手动检查go.sum文件：
```
$ go get golang.org/x/mod/gosumcheck
$ gosumcheck /path/to/go.sum
```

## FAQ

### 我向存储库提交了一个新的更改（或发布了一个新版本），当我运行 `go -u` 或 `go list -m --versions` 时，它为什么不显示？

为了改善我们服务的缓存和服务延迟，新版本可能不会立即显示。 如果您希望镜像中立即提供新代码，请首先确保底层源存储库中的此修订版本具有语义版本标记。 然后通过 `go get module@version` 显式请求该版本。 在缓存过期一分钟后，go命令将看到标记的版本。 如果这对您不起作用，请[提出问题](https://github.com/golang/go/issues/new?title=proxy.golang.org%3A+)。

### 我从我的存储库中删除了一个错误的版本，但它仍然出现在镜像中，我该怎么办？

只要有可能，镜像的目的是缓存内容，以避免破坏依赖于您的程序包的人的构建（ build ），因此即使镜像在原点不可用，这个不良版本仍可在镜像中使用。 如果删除整个存储库，同样适用上述的情况。 我们建议您创建一个新版本并鼓励人们使用该版本。

### 我在一个无法使用镜像的环境中运行 go 命令。

[go 命令文档](https://golang.org/cmd/go/#hdr-Module_downloading_and_verification) 描述了配置详细信息，包括如何禁用这些服务器或使用不同的服务器。

### 如果我没有设置 GOPRIVATE 并向这些服务请求私有模块，那么什么泄漏？

代理与校验和数据库协议仅将模块路径和版本发送到远程服务器。如果您请求私有模块，镜像将尝试下载它，就像任何Go用户一样，并以相同的方式失败。有关失败请求的信息不会在任何地方发布 请求的唯一跟踪将在内部日志中，该日志由[隐私策略](https://proxy.golang.org/privacy)管理。

via: https://proxy.golang.org/

作者：https://proxy.golang.org/ # 原文没有作者

译者：[ZackLiuCH](https://github.com/ZackLiuCH)

校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出