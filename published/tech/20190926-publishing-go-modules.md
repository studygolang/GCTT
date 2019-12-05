首发于：https://studygolang.com/articles/25129

# 发布 Go Modules

## 简介

本文是 go modules 系统的第三部分

- Part 1: [使用 Go Modules](https://blog.golang.org/using-go-modules)  [译文](https://studygolang.com/articles/19334)
- Part 2: [迁移到 Go Modules](https://blog.golang.org/migrating-to-go-modules)  [译文](https://studygolang.com/articles/23133)
- Part 3: 发布 `go modules` (本文)
- Part 4: [Go Modules: v2 及以后的版本](https://blog.golang.org/v2-go-modules)

本文讨论如何编码和发布 go 模块，发布后就可以被其他模块依赖使用了。

注意： 本文只涉及到 v1 及以前的版本， 如果你想了解 v2 版本， 请参照 [Go Modules: v2 及以后的版本](https://blog.golang.org/v2-go-modules) 。

本文中列出的例子使用的是 Git ，其他版本控制工具如 `Mercurial`、`Bazaar` 等等也支持。

## 工程配置

本文的前期准备：有一个已经建好的工程。让我们以 [使用 `Go Modules`](https://blog.golang.org/using-go-modules) 篇尾的文章来开始示例。

```bash
$ cat go.mod
module example.com/hello

go 1.12

require rsc.io/quote/v3 v3.1.0
```

```bash
$ cat go.sum
golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c h1:qgOY6WgZOaTkIIMiVjBQcw93ERBE4m30iBm00nkL0i8=
golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c/go.mod h1:NqM8EUOU14njkJ3fqMW+pc6Ldnwhi/IjpwHt7yyuwOQ=
rsc.io/quote/v3 v3.1.0 h1:9JKUTTIUgS6kzR9mK1YuGKv6Nl+DijDNIc0ghT58FaY=
rsc.io/quote/v3 v3.1.0/go.mod h1:yEA65RcK8LyAZtP9Kv3t0HmxON59tX3rD+tICJqUlj0=
rsc.io/sampler v1.3.0 h1:7uVkIFmeBqHfdjD+gZwtXXI+RODJ2Wc4O7MPEh/QiW4=
rsc.io/sampler v1.3.0/go.mod h1:T1hPZKmBbMNahiBKFy5HrXp6adAjACjK9JXDnKaTXpA=
```

```go
$ cat hello.go
package hello

import "rsc.io/quote/v3"

func Hello() string {
    return quote.HelloV3()
}

func Proverb() string {
    return quote.Concurrency()
}
```

```go
$ cat hello_test.go
package hello

import (
    "testing"
)

func TestHello(t *testing.T) {
    want := "Hello, world."
    if got := Hello(); got != want {
        t.Errorf("Hello() = %q, want %q", got, want)
    }
}

func TestProverb(t *testing.T) {
    want := "Concurrency is not parallelism."
    if got := Proverb(); got != want {
        t.Errorf("Proverb() = %q, want %q", got, want)
    }
}

$
```

创建一个 `git` 仓库， 添加一条初始化的信息。 如果你是要发布你自己的工程，请确保你的工程里包含 `LICENSE` (许可)文件。进入到包含 `go.mod` 的目录， 创建仓库。

```bash
$ git init
$ git add LICENSE go.mod go.sum hello.go hello_test.go
$ git commit -m "hello: initial commit"
$
```

## 语义版本和模块

`go.mod` 中每一个被依赖的模块都有一个语义版本，该语义版本是依赖该模块构建本模块时使用的最小版本。

语义版本格式：`vMAJOR.MINOR.PATCH` (v 主版本号.次版本号.修订版本号)

- 当你模块的公开接口作了向后不兼容的修改后，需要增加主版本号。不到万不得已时，不要这么做。
- 当你作了向后兼容的修改，如修改依赖或增加新的函数、方法、结构体的字段或类型时，增加次版本号。
- 当你作了并未影响模块的公开接口或依赖的微小的修改，如修复一个 bug 时，增加修订版本号。

你可以在版本号后加连字符和点分隔的标识以指定一个预发布的版本（如 `v1.0.1-alpha` 或 `v2.2.2-beta.2`）。go 命令选择正常版本的优先级高于预发布版本，因此如果你的模块有正常发布的版本，使用者必须显示指定预发布版本号（例如：`go get example.com/hello@v1.0.1-alpha`）才能使用预发布版本。

v0 主版本和预发布版本不需要考虑向后兼容，仅作为提交给使用者稳定版本之前的一份 API 精选。然而，v1 及之后的版本需要保证在当前主版本内向后兼容。

`go.mod` 中引入的版本可以是仓库中明确打上发布标签的版本（如 `v1.5.2`），也可以是一个基于某次提交的 [伪版本](https://golang.org/cmd/go/#hdr-Pseudo_versions)（如 `v0.0.0-20170915032832-14c0d48ead0c`）。伪版本是一种特殊的预发布版本。当开发者依赖一个尚未发布任何语义版本标签的工程，或基于一次尚未打标签的提交进行开发时，伪版本就派上了用场。但是开发者不能假定伪版本提供的是稳定的、经过完整测试的接口。给你的模块打上明确的版本标签就意味着向该模块的使用者保证了该版本是经过完整测试且稳定的。

不要删除你仓库里的版本标签。如果你在某版本中发现了一个 bug 或 issue ，那就发布一个新版本，因为依赖被你删除的版本的工程会编译失败。同样，一旦你发布了一个版本，就不要修改或重写它。[go 模块及检验和数据库](https://blog.golang.org/module-mirror-launch) 储存模块、它们的版本和加密签名哈希，以确保给定版本的构建随时间推移依然保持可复制性。

### v0: 初始化的版本, 非稳定版

现在我们来给模块打上 `v0` 语义版本标签。`v0` 版本不保证稳定，因此如果一个工程还在提炼公开接口（还未到稳定版本）的阶段，那么就应该以 `v0` 开始其版本。

打一个新的版本标签分为以下几步：

1. 执行 `go mod tidy` ，清理模块不再需要的依赖
2. 执行 `go test ./...` ，最终确认一下没有任何问题
3. 用命令 `git tag` 给工程打上一个新版本标签
4. 把新标签 push 到仓库

```bash
$ go mod tidy
$ go test ./...
ok      example.com/hello       0.015s
$ git add go.mod go.sum hello.go hello_test.go
$ git commit -m "hello: changes for v0.1.0"
$ git tag v0.1.0
$ git push origin v0.1.0
$
```

现在其他的工程就可以依赖 `example.com/hello` 的 `v0.1.0` 版本了。对于你自己的模块，你可以执行 `go list -m example.com/hello@v0.1.0` 来确认最新的版本可用（本例中的模块并不存在，所以不可用）。如果你没有即时看到最新的版本且你用了 `Go module proxy`（Go 1.13 后的版本默认使用），给代理一点加载新版本的时间，几分钟后再试一试。

如果你修改了公开接口、在版本为 `v0` 的模块基础上做出了重大改变、抑或更新了你依赖的某个模块的次版本或（完整）版本，那么在你的下次发布中增加次版本号。例如，`v0.1.0` 的下一个版本为 `v0.2.0` 。

如果你在已有版本上修改了一个 bug ，增加修订版本号。例如，`v0.1.0` 的下一个版本为 `v0.1.1` 。

### v1: 第一个稳定版

当你确认你模块的公开接口完全稳定时，可以发布版本 `v1.0.0` 。`v1` 主版本号向使用者声明了该模块的公开接口不会做任何不兼容的修改。他们可以升级到新的 `v1` 主版本下的任何次版本和修订版本，并且自己的代码运行不会崩溃，函数和方法签名不会修改，向外暴露的类型不会被删除，等等。如果修改了公开接口，那么这些修改也都是向后兼容的（例如，给一个结构体增加新的字段）且会在后面发布的次版本中包含进来。如果有修复 bug 的修改（例如安全 bug 修复），这些修改会在后面发布的修订版本（或作为次版本的一部分）包含进来。

有时维持向后兼容可能导致写出糟糕的公开接口，这也可以接受。不完美的公开接口总好过让使用者既有的代码崩溃。

标准库的 `strings` 包就是一个以保持公开接口一致性的代价来维持向后兼容的典型例子。

- [Split](https://godoc.org/strings#Split) 把字符串以指定的分隔符进行分割，切成多个子字符串，返回分隔符分隔的子字符串的切片
- [SplitN](https://godoc.org/strings#SplitN) 可以控制返回的子字符串的个数

[Replace](https://godoc.org/strings#Replace) 有个入参， 指定从字符串开头处要替换的实例的个数（这点与 Split 用法不一样）。

联想到 Split 和 SplitN ，你会想当然的认为有 Replace 和 ReplaceN ，且用法与之相似。但是我们不能做到像我们承诺的那样在不让使用者的代码崩溃的前提下修改已有的 Replace。因此，在 Go 1.12 中我们加入了一个新的函数，[ReplaceAll](https://godoc.org/strings#ReplaceAll) 。这样导致公开接口有点别扭，Split 和 Replace 用法完全不同，但是这种不一致的情况好过一次重大变革。

假定你对 `example.com/hello` 的公开接口很满意， 且你希望发布第一个稳定版本 `v1` 。

用给 `v0` 版本打标签相同的处理打上 `v1` 标签：执行 `go mod tidy`  `go test ./...`  ，给版本打上标签，push 到 origin 仓库

```bash
$ go mod tidy
$ go test ./...
ok      example.com/hello       0.015s
$ git add go.mod go.sum hello.go hello_test.go
$ git commit -m "hello: changes for v1.0.0"
$ git tag v1.0.0
$ git push origin v1.0.0
$
```

至此，`example.com/hello`  v1 版本的公开接口就发布完了。这就向所有人传递了一个信息：我们的公开接口很稳定，他们使用时不会有任何问题。

## 总结

本文讲了给模块打语义版本标签以及何时发布 `v1` 版本的完整流程。后续会撰文讲述如果维持和发布 `v2` 及其他版本。

如果你想提供反馈意见，或参与塑造 Go 依赖管理的未来，请给我们发 [bug reports](https://github.com/golang/go/issues/new) 或 [experience reports](https://github.com/golang/go/wiki/ExperienceReports)

关联文章：

- [Go Modules: v2 and Beyond](https://blog.golang.org/v2-go-modules)
- [Module Mirror and Checksum Database Launched](https://blog.golang.org/module-mirror-launch)
- [Migrating to Go Modules](https://blog.golang.org/migrating-to-go-modules)
- [Using Go Modules](https://blog.golang.org/using-go-modules)
- [Go Modules in 2019](https://blog.golang.org/modules2019)
- [A Proposal for Package Versioning in Go](https://blog.golang.org/versioning-proposal)
- [The cover story](https://blog.golang.org/cover)
- [The App Engine SDK and workspaces (GOPATH)](https://blog.golang.org/the-app-engine-sdk-and-workspaces-gopath)
- [Organizing Go code](https://blog.golang.org/organizing-go-code)

---

via: https://blog.golang.org/publishing-go-modules

作者：[Tyler Bui-Palsulich](https://blog.golang.org/publishing-go-modules)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
