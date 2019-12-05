首发于：https://studygolang.com/articles/25130

# Go Modules: v2 及更高版本

## 简介

本文是 Go modules 系统的第四部分

- Part 1: [使用 Go Modules](https://blog.golang.org/using-go-modules)  [译文](https://studygolang.com/articles/19334)
- Part 2: [迁移到 Go Modules](https://blog.golang.org/migrating-to-go-modules)  [译文](https://studygolang.com/articles/23133)
- Part 3: [发布 Go Modules](https://blog.golang.org/publishing-go-modules) [译文](https://studygolang.com/articles/25129)
- Part 4: [Go Modules : v2 及更高版本](https://blog.golang.org/v2-go-modules) (本文) 

随着成功的项目逐渐成熟以及新需求的加入，早期的功能和设计决策可能不再适用。 开发者们可能希望通过删除废弃使用的功能、重命名类型或将复杂的程序拆分为可管理的小块来融入他们的经验教训。这种类型的变更要求下游用户进行更改才能将其代码迁移到新的 API，因此在没有认真考虑收益成本比重的情况下，不应进行这种变更。

对于还在早期开发阶段的项目（主版本号是 `v0`），用户会期望偶尔的重大变更。对于声称已经稳定的项目（主版本是 `v1` 或者更高版本），必须在新的主版本进行重大变更。这篇文章探讨了主版本语义、如何创建并发布新的主版本以及如何维护一个 Go Modules 的多个主版本。

## 主版本和模块路径

模块在 Go 中确定了一个重要的原则，即 “[导入兼容性规则](https://research.swtch.com/vgo-import)”

> 如果旧包和新包的导入路径相同，新包必须向后兼容旧的包

根据这条原则，一个软件包新的主版本没有向后兼容以前的版本。这意味着这个软件包新的主版本必须使用和之前版本不同的模块路径。从 `v2` 开始，主版本号必须出现在模块路径的结尾（在 go.mod 文件的 `module` 语句中声明）。例如，当模块 `github.com/googleapis/gax-go` 的开发者们开发完 `v2` ，他们用了新的模块路径 `github.com/googleapis/gax-go/v2` 。想要使用 `v2` 的用户必须把他们的包导入和模块要求更改为 ``github.com/googleapis/gax-go/v2`` 

需要主版本号后缀是 Go 模块和大多数其他依赖管理系统不同的方式之一。后缀用于解决[菱形依赖问题](https://research.swtch.com/vgo-import#dependency_story)。在 Go 模块之前，[gopkg.in](http://gopkg.in/) 允许软件包维护者遵循我们现在称为导入兼容性规则的内容。使用 gopkg.in 时，如果你依赖一个导入了 `gopkg.in/yaml.v1` 的包以及另一个导入了 `gopkg.in/yaml.v2` 的包，这不会发生冲突，因为两个 `yaml` 包有着不同的导入路径（它们使用和 Go Modules 类似的版本后缀）。由于 gopkg.in 和 Go Modules 共享相同的版本号后缀方法，因此 Go 命令接受 `gopkg.in/yaml.v2` 中的 `.v2` 作为有效的版本号。这是一个为了和 gopkg.in 兼容的特殊情况，在其他域托管的模块需要使用像 `/v2` 这样的斜杠后缀。

## 主版本策略

推荐的策略是在以主版本后缀命名的目录中开发 `v2+` 模块。

```bash
github.com/googleapis/gax-go @ master branch
/go.mod    → module github.com/googleapis/gax-go
/v2/go.mod → module github.com/googleapis/gax-go/v2
```

这种方式与不支持 Go Modules 的一些工具兼容：仓库中的文件路径与 `GOPATH` 模式下 `go get` 命令预期的路径匹配。这一策略也允许所有的主版本一起在不同的目录中开发。

其他的策略可能是将主版本放置在单独的分支上。然而，如果 `v2+` 的源代码在仓库的默认分支上（一般是 master），不支持版本的工具（包括 GOPATH 模式下的 Go 命令）可能无法区分不同的主版本。

本文中的示例遵循主版本子目录策略，所以提供了最大的兼容性。我们建议模块的作者遵循这种策略，只要他们还有用户在使用 `GOPATH` 模式开发。

## 发布 v2 及更高版本

这篇文章以 `github.com/googleapis/gax-go` 为例：

```bash
$ pwd
/tmp/gax-go
$ ls
CODE_OF_CONDUCT.md  call_option.go  internal
CONTRIBUTING.md     gax.go          invoke.go
LICENSE             go.mod          tools.go
README.md           go.sum          RELEASING.md
header.go
$ cat go.mod
module github.com/googleapis/gax-go

go 1.9

require (
    github.com/golang/protobuf v1.3.1
    golang.org/x/exp v0.0.0-20190221220918-438050ddec5e
    golang.org/x/lint v0.0.0-20181026193005-c67002cb31c3
    golang.org/x/tools v0.0.0-20190114222345-bf090417da8b
    google.golang.org/grpc v1.19.0
    honnef.co/go/tools v0.0.0-20190102054323-c2f93a96b099
)
$
```

要开始开发 `github.com/googleapis/gax-go` 的 `v2` 版本，我们将创建一个新的 `v2/` 目录并将包的内容复制到该目录中。

```bash
$ mkdir v2
$ cp *.go v2/
building file list ... done
call_option.go
gax.go
header.go
invoke.go
tools.go

sent 10588 bytes  received 130 bytes  21436.00 bytes/sec
total size is 10208  speedup is 0.95
$
```

现在，我们通过复制当前的 `go.mod` 文件并且在 module 路径上添加 `/v2` 后缀来创建属于 v2 的 `go.mod` 文件。

```bash
$ cp go.mod v2/go.mod
$ go mod edit -module github.com/googleapis/gax-go/v2 v2/go.mod
$
```

注意： `v2` 版本被视为与 `v0 / v1` 版本分开的模块，两者可以共存于同一构建中。因此，如果你的 `v2+` 模块具有多个软件包，你应该更新它们使用新的 `/v2` 导入路径，否则，你的 `v2+` 模块会依赖你的 `v0 / v1` 模块。要升级所有 `github.com/my/project` 为 `github.com/my/project/v2` ，可以使用 `find` 和 `sed` 命令：

```bash
$ find . -type f \
    -name '*.go' \
    -exec sed -i -e 's,github.com/my/project,github.com/my/project/v2,g' {} \;
$
```

现在我们有了一个 `v2` 模块，但是我们要在版本发布之前进行实验并进行修改。在我们发布 `v2.0.0` （或者其他没有预发布后缀的版本）之前，我们可以进行开发并且可以做出重大变更，就如同我们决定实现新 API 一样。  如果我们希望用户能够在正式发布新 API 之前对其进行试验，可以选择发布 `v2` 预发布版本：

```bash
$ git tag v2.0.0-alpha.1
$ git push origin v2.0.0-alpha.1
$
```

一旦我们对 `v2` API 感到满意并且确定不会再有别的重大变更，我们可以打上 Git 标记 `v2.0.0` 。

```bash
$ git tag v2.0.0
$ git push origin v2.0.0
$
```

到那时，就有两个主版本需要维护。向后兼容的更改和错误修复使用新的次版本或者补丁版本发布（比如 `v1.1.0` ， `v2.0.1` 等）。

## 总结

主版本变更会带来开发和维护的开销，并且需要下游用户的额外付出才能迁移。越大的项目中这种主版本变更的开销就越大。只有在确定了令人信服的理由之后，才应该进行主版本变更。一旦确定了令人信服的重大变更原因，我们建议在 master 分支进行多个主版本的开发，因为这样能与各种现有工具兼容。

对 `v1+` 模块的重大变更应该始终发生在新的 `vN+1` 模块中。一个新模块发布时，对于维护者和需要迁移到这个新软件包的用户来说意味着更多的工作。因此，维护人员应该在发布稳定版本之前对其 API 进行验证，并仔细考虑在 `v1` 版本之后是否确有必要进行重大变更。

## 相关文章

- [发布 Go Modules](https://blog.golang.org/publishing-go-modules)
- [启用模块镜像和校验数据库](https://blog.golang.org/module-mirror-launch)
- [迁移到 Go Modules](https://blog.golang.org/migrating-to-go-modules)
- [使用 Go Modules](https://blog.golang.org/using-go-modules)
- [2019 年的 Go Modules](https://blog.golang.org/modules2019)
- [Go 中软件包版本控制的建议](https://blog.golang.org/versioning-proposal)
- [封面故事](https://blog.golang.org/cover)
- [App Engine SDK 和工作区](https://blog.golang.org/the-app-engine-sdk-and-workspaces-gopath)
- [组织 Go 代码](https://blog.golang.org/organizing-go-code)

---

via: https://blog.golang.org/v2-go-modules

作者：[Jean de Klerk 和 Tyler Bui-Palsulich](https://blog.golang.org)
译者：[befovy](https://github.com/befovy)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
