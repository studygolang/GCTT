首发于：https://studygolang.com/articles/35225

# Go Module 教程第 4 部分：镜像、校验和以及 Athens

前三个教程：

- [Go Module 教程第 1 部分：为什么和做什么](https://studygolang.com/articles/24580)
- [Go Module 教程第 2 部分：项目、依赖和 gopls](https://studygolang.com/articles/35202)
- [Go Module 教程第 3 部分：最小版本选择](https://studygolang.com/articles/35210)

> 注意，该教程基于 Go1.13。最新版本可能会有所不同。

## 引言

当我第一次学习模块时遇到的一个长期问题是模块镜像、校验和数据库以及 Athens 是如何工作的。Go 团队已经写了大量关于模块镜像和校验和数据库的内容，但我希望在这里合并最重要的信息。在这篇文章中，我提供了这些系统的用途，你可以控制的不同的配置选项，并用示例程序展示了这些系统的运行情况。

## 模块镜像（Mirror）

[模块镜像](https://blog.golang.org/module-mirror-launch)是 2019 年 8 月发布的，是 Go 版本 1.13 中用于抓取模块的默认系统。模块镜像是作为一个代理服务器实现的，它面向 VCS 环境，帮助加速获取构建应用程序所需的本地模块。代理服务器实现了一个基于 REST 的 API，并且是围绕 Go 工具的需求而设计的。

模块镜像缓存模块及其被请求的特定版本，这样可以更快地检索未来的请求。一旦获取代码并将其缓存到模块镜像中，就可以快速地将其提供给世界各地的客户机。模块镜像还允许用户继续获取原始 VCS 位置不再可用的源代码。这可以防止像 Node 开发者在 2016 年[遇到的问题](https://www.theregister.co.uk/2016/03/23/npm_left_pad_chaos/)。

## 校验和数据库

[校验和数据库](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md)也于 2019 年 8 月推出，是一个防篡改的模块散列码日志，可用于验证不可信代理或来源。校验和数据库是作为服务实现的，Go 工具使用它来验证模块。它验证了特定版本的任何给定模块的代码是否是相同的，不管是谁、是什么、在哪里以及如何获取。它还解决了其他依赖管理系统尚未解决的其他安全问题（如上面的链接所述）。Google 拥有现存唯一的校验和数据库，但是它可以被私有的模块镜像缓存。

## 模块索引

该[索引服务](https://index.golang.org/)是为那些希望跟踪添加到 Google 模块镜像的新模块的模块列表的开发人员提供的。像 [pkg.go.dev](https://pkg.go.dev/) 这样的网站使用索引来检索和发布模块的信息。

## 你的隐私

正如[隐私政策](https://sum.golang.org/privacy)中记录的那样，Go 团队构建这些服务是为了尽可能少地保留关于使用情况的信息，同时仍然确保它们能够检测和修复问题。然而，像 IP 地址这样的个人身份信息可以保存 30 天。如果这对您或您的公司是一个问题，您可能不希望使用 Go 团队的服务来获取和验证模块。

## Athens

[Athens](https://docs.gomods.io/)是一个模块镜像，你可以搭建你的私人环境。使用私有模块镜像的一个原因是允许缓存公共模块镜像无法访问的私有模块。最棒的是 Athens 项目提供了一个 Docker 容器，发布在 Docker Hub 上，所以不需要特殊安装。

> GCTT 注：还有 goproxy.io 和 goproxy.cn，这两个都是国人开发的

**清单 1**

```bash
docker run -p '3000:3000' gomods/athens:latest
```

清单 1 显示了如何使用 Docker 运行本地 Athens 服务器。稍后我们将使用这个工具来查看 Go 工具的运行情况，并监视所有 Go 工具 Web 调用的 Athens 日志。要知道，Athens Docker 映像默认启动了临时磁盘存储，所以当你关闭运行的容器时，所有内容都将被清空。

Athens 有能[代理](https://docs.gomods.io/configuration/sumdb)校验和数据库。当 Go 工具被配置为使用一个像 Athens 一样的私有模块镜像时，当需要从校验和数据库中查找散列码时，Go 工具将尝试使用相同的私有模块镜像。如果你正在使用的私有模块镜像不支持代理校验和数据库，那么将直接访问校验和数据库，除非它被手动关闭。

**清单 2**

```bash
http://localhost:3000/sumdb/sum.golang.org/latest

go.sum database tree
756113
k9nFMBuXq8uk+9SQNxs/Vadri2XDkaoo96u4uMa0qE0=

— sum.golang.org Az3grgIHxiDLRpsKUElIX5vJMlFS79SqfQDSgHQmON922lNdJ5zxF8SSPPcah3jhIkpG8LSKNaWXiy7IldOSDCt4Pwk=
```

清单 2 显示了 Athens 如何成为校验和数据库代理。第一行中列出的 URL 要求本地运行的 Athens 服务从校验和数据库中检索有关最新签名树的信息。你可以了解为什么 GOSUMDB 配置为名称而不是 URL。

## 环境变量

有几个环境变量控制 Go 工具的行为，因为它与模块镜像和校验和数据库相关。需要在每个开发人员或构建环境的计算机级别上设置这些变量。

**GOPROXY**：一组指向模块镜像的 URL，用于抓取模块。如果你希望 Go 工具只从 VCS 位置直接获取模块，那么可以将其设置为 direct。如果你将此设置为 off，那么模块将不会被下载。使用 off 可以用在保留 vendoring 或模块缓存的构建环境中。

**GOSUMDB**：用于验证给定模块/版本的代码的校验和数据库的名称没有随时间变化。此名称用于形成一个正确的 URL，该 URL 告诉 Go 工具在哪里执行这些校验和数据库查找。这个 URL 可以指向 Google 拥有的校验和数据库，或者指向支持缓存或代理校验和数据库的本地模块镜像。如果你不希望 Go 工具验证添加到 go.sum 文件中的给定模块/版本的哈希代码，也可以将其设置为 off。只有在向 go.sum 文件添加任何新的 go.sum 行之前，才会查询校验和数据库。

**GONOPROXY**：模块的一组基于 URL 的模块路径，不应该使用模块镜像获取，而是直接从 VCS 位置获取。

**GOPRIVATE**：一个便捷的变量，用于设置具有相同默认值的 GONOPROXY 和 GONOSUMDB。

> GCTT 注：这些环境变量的帮助文档可以通过 go help environment 获得

## 隐私语义学（Privacy Semantics）

在考虑隐私和项目所依赖的模块时，需要考虑以下几点。特别是那些你不想让别人知道的私人模块。下面的图表尝试提供隐私选项。同样，需要在每个开发人员或构建环境的计算机级别上设置此配置。

**清单 3**

```bash
Option            : Fetch New Modules     : Validate New Checksums
-----------------------------------------------------------------------------------------
Complete Privacy  : GOPROXY="direct"      : GOSUMDB="off"
Internal Privacy  : GOPROXY="Private_URL" : GOSUMDB="sum.golang.org"
                                            GONOSUMDB="github.com/mycompany/*,gitlab.com/*"
No Privacy        : GOPROXY="Public_URL"  : GOSUMDB="sum.golang.org"
```

**Complete Privacy**：代码直接从 VCS 服务器获取，没有生成并添加到 go.sum 文件的哈希代码，并且它们也不会从校验和数据库中查找。

**Internal Privacy**：代码是通过一个像 [Athens](https://docs.gomods.io/) 这样的私有模块镜像获取的，并且不会在校验和数据库中查找生成和添加到 go.sum 文件中的，GONOSUMDB 下列出的指定 URL 路径的哈希代码。如果需要，将在校验和数据库中查找不属于 GONOSUMDB 中列出的路径的模块。

**No Privacy**：代码是通过像 [Google](https://proxy.golang.org/) 或 [Goproxy.CN](https://goproxy.cn/) 公共服务器这样的公共模块镜像获取的。在这种情况下，你所依赖的所有模块都需要是公共模块，并可由你选择的公共模块进行访问。这些公共模块镜像将记录你的请求和其中包含的详细信息。访问谷歌拥有的校验和数据库也将被记录。所记录的信息受各自的隐私策略控制。

从来没有理由在校验和数据库中查找私有模块的哈希代码，因为校验和数据库中永远不会有这些模块的列表。公共模块镜像不能访问私有模块，因此不能生成和存储哈希代码。对于私有模块，你需要依靠内部策略和实践来保持给定模块/版本的代码一致。但是，如果私有模块/版本的代码确实发生了更改，那么当第一次在新机器上获取并缓存模块/版本时，Go 工具仍然可以发现差异。

任何时候，当模块/版本被添加到机器上的本地缓存中，并且 go.sum 文件中已经有一个条目时，go.sum 文件中的哈希码都会与刚才在缓存中获取的内容进行比较。如果哈希代码不匹配，就说明发生了变化。这个工作流程最好的部分是不需要校验和数据库查找，因此任何给定版本的私有模块仍然可以在不损失隐私的情况下进行验证。显然，这完全取决于你第一次获取私有模块/版本的时间，这对于存储在校验和数据库中的公共模块/版本来说也是同样的问题。

当使用 Athens 作为模块镜像，需要考虑 Athens 配置选项。

**清单 4**

```bash
GlobalEndpoint = "https://<url_to_upstream>"
NoSumPatterns = ["github.com/mycompany/*]
```

清单 4 中的这些设置来自 Athens 的文档，它们很重要。默认情况下，Athens 将直接从不同的 VCS 服务器获取模块。这将为你的环境保持最高级别的隐私。但是，可以通过将 GlobalEnpoint 设置为该模块镜像的 URL，将 Athens 指向另一个模块镜像。这将使你在获取新的公共模块时获得更好的性能，但是你将失去隐私。

另一个设置称为 NoSumPatterns，它有助于验证开发人员和构建环境的正确配置。开发人员向 GONOSUMDB 添加的相同路径集应该添加到 NoSumPatterns 中。当检查和数据库请求访问 Athens 以获取与路径匹配的模块时，它将返回一个状态代码，该状态代码将导致 Go 工具失败。这表明开发人员的设置是错误的。换句话说，如果机器配置正确，那么这个请求从一开始就不应该到达 Athens 。

## Vendoring

我相信每个项目都应该提供他们的依赖关系，或者认为这样做不合理或者不切实际。像 Docker 和 Kubernetes 这样的项目不能提供他们的依赖项，因为依赖项太多了。然而，对于我们大多数人来说，情况并非如此。在 v1.14 版本中，对 vendoring 和模块有很好的支持。我将在另一篇文章中讨论这个问题。

我提到 vendoring 有一个重要的原因。我听说有人用 Athens 或者私有模块镜像代替 vendoring。我认为这是个错误。这两者没有任何关系。您可以争辩模块镜像 vendoring 的依赖关系，因为模块的代码是持久化的，但是代码仍然远离依赖它的项目。即使你相信你的模块镜像的弹性，我也相信没有什么可以替代你的项目拥有它所需要的所有源代码，除了项目本身来构建代码之外不依赖其他任何东西。

## 工具的使用

有了所有这些背景和知识，是时候看看 Go 工具是如何工作的了。为了了解环境变量如何影响 Go 工具，我将运行几个不同的场景。在开始之前，可以通过运行 go env 命令来了解默认值。

**清单 5**：

```bash
$ go env
GONOPROXY=""
GONOSUMDB=""
GOPRIVATE=""
GOPROXY="https://proxy.golang.org,direct"
GOSUMDB="sum.golang.org"
```

清单 5 显示了告诉 Go 工具使用 Google 模块镜像和 Google 校验和数据库的默认值。如果你需要的所有代码都可以通过这些 Google 服务访问，那么这是推荐的配置。如果 Google 模块镜像碰巧响应了410（已消失）或404（未找到），那么使用 direct（这是 GOPROXY 配置的一部分）将允许 Go 工具改变方向并直接从 VCS 位置获取模块/版本。任何其他状态代码（比如 500）都会导致 Go 工具失败。

如果 Google 模块镜像碰巧对给定模块/版本响应了 410 或 404，那是因为它不在缓存中，可能不能缓存，而私有模块就是这种情况。在这种情况下，校验和数据库中很可能也没有列表。即使 Go 工具可以成功地直接获取模块/版本，但是查找将会失败，而 Go 工具仍然会失败。使用私有模块时需要注意的一些事项。

因为我不能显示任何来自 Google 模块镜像的日志，所以我将使用 Athens 运行一个本地模块镜像。这将允许你看到 Go 工具和模块镜像在运行。最后，Athens 实现了相同的语义和工作流。

## 项目

要创建项目，请启动终端会话并创建项目结构。

**清单 6**

```bash
$ cd $HOME
$ mkdir app
$ mkdir app/cmd
$ mkdir app/cmd/db
$ touch app/cmd/db/main.go
$ cd app
$ go mod init app
$ code .
```

清单 6 显示了为在磁盘上创建项目结构、为模块初始化项目和运行 VSCode 而运行的命令。

**清单 7**

<https://play.studygolang.com/p/TtbuNj_IAwL>

```go
package main

import (
	"context"
	"log"

	"github.com/Bhinneka/golib"
	db "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

func main() {
	c, err := db.NewCluster([]db.Host{{Name: "localhost", Port: 3000}}, nil)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err = c.Query(context.Background(), db.Query{}); err != nil {
		log.Fatalln(err)
	}

	golib.CreateDBConnection("")
}
```

清单 7 显示了 main.go 的代码。随着项目的设置和主要功能的到位，我将在项目中运行三个不同的场景，以更好地理解环境变量和 Go 工具。

### 场景 1：Athens 模块镜像

在这个场景中，我将使用 Athens 作为私有模块镜像替代 Google 模块镜像。

**清单 8**

```bash
GONOSUMDB=""
GONOPROXY=""
GOSUMDB="sum.golang.org"
GOPROXY="http://localhost:3000,direct"
```

清单 8 显示，我要求 Go 工具对模块镜像使用在端口 3000 上本地运行的 Athens 服务。如果模块镜像以 410（已消失）或 404（未找到）响应，则尝试直接拉出模块。默认情况下，如果需要，Go 工具现在将使用 Athens 来访问校验和数据库。

接下来，为运行 Athens 启动一个新的终端会话。

**清单 9**

```bash
$ docker run -p '3000:3000' -e ATHENS_LOG_LEVEL=debug -e GO_ENV=development gomods/athens:latest

INFO[10:15AM]: Exporter not specified. Traces won't be exported
2021-09-05 10:15:08.464666 I | Starting application at port :3000
```

清单 9 在第一行显示了在一个新的终端会话中运行的命令，该命令使用额外的调试日志记录启动并运行 Athens 服务。确保你本机启动了 Docker。一旦 Athens 启动，你应该会看到清单中的输出。

要查看 Go 工具使用 Athens 服务的情况，请在用于创建项目的原始终端会话中运行以下命令。

**清单 10**

```bash
$ export GOPROXY="http://localhost:3000,direct"
$ rm go.*
$ go mod init app
$ go mod tidy
```

清单 10 显示了将 GOPROXY 变量设置为使用 Athens 服务、删除模块文件和重新初始化应用程序的命令。最后的命令 `go mod tidy` 将使 Go 工具与 Athens 服务通信，以获取构建这个项目所需的模块。

**清单 11**

```bash
handler: GET /github.com/!bhinneka/@v/list [404]
handler: GET /github.com/@v/list [404]
handler: GET /github.com/!bhinneka/golib/@v/list [200]
handler: GET /gopkg.in/@v/list [404]
handler: GET /github.com/!bhinneka/golib/@latest [200]
handler: GET /gopkg.in/rethinkdb/rethinkdb-go.v5/@v/list [200]
handler: GET /github.com/bitly/@v/list [404]
handler: GET /github.com/bmizerany/@v/list [404]
handler: GET /github.com/bmizerany/assert/@v/list [200]
handler: GET /github.com/bitly/go-hostpool/@v/list [200]
handler: GET /github.com/bmizerany/assert/@latest [200]
```

清单 11 显示了来自 Athens Service 的重要输出。如果查看 go.mod 和 go.sum 文件，你将看到构建和验证项目所需的所有内容。

> GCTT 注：你看到的信息可能会有所不同，因为 Athens 可能升级了，日志输出方式变了

### 场景 2：Athens 模块镜像/直接从 GitHub Modules 获取

在这个场景中，我不希望从模块镜像获取任何托管在 GitHub 上的模块。我希望这些模块可以直接从 GitHub 获取。

**清单 12**

```bash
$ export GONOPROXY="github.com"
$ export GOPROXY="http://localhost:3000,direct"
$ rm go.*
$ go mod init app
$ go mod tidy
```

清单 12 显示了在这个场景中如何设置 GONOPROXY 变量。现在 GONOPROXY 告诉 Go 工具直接获取任何名称以 github. com 开头的模块。不要使用 GOPROXY 变量定义的模块镜像。虽然我使用 GitHub 来展示这一点，但是如果你运行一个像 GitLab 这样的本地 VCS，这个配置是完美的。这将允许你直接获取私有模块。

**清单 13**

```bash
handler: GET /gopkg.in/@v/list [404]
handler: GET /gopkg.in/rethinkdb/rethinkdb-go.v5/@v/list [200]
```

清单 13 显示了运行 go mod tidy 之后从 Athens Service 得到的更重要的输出。这次 Athens 只显示对位于 gopk.in 的两个模块的请求。位于 github. com 的模块不再需要 Athens 的服务。

### 场景 3：Module Mirror 404

在这个场景中，我将使用自己的模块镜像，它将为每个模块请求返回一个 404。当模块镜像返回 410（已消失）或 404（未找到）时，Go 工具将沿着 GOPROXY 变量中列出的逗号分隔的其他镜像集继续。

**清单 14**

<https://play.studygolang.com/p/uEH4_b6QrAO>

```go
package main

import (
	"log"
	"net/http"
)

func main() {
	h := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s -> %s\n", r.Method, r.URL.Path, r.RemoteAddr)
		w.WriteHeader(http.StatusNotFound)
	}
	http.ListenAndServe(":3000", http.HandlerFunc(h))
}
```

清单 14 显示了我的模块镜像的代码。它能够记录每个请求的跟踪并返回 `http.StatusNotFound`，即 404。

**清单 15**

```bash
$ unset GONOPROXY
$ export GOPROXY="http://localhost:3000"
$ rm go.*
$ go mod init app
$ go mod tidy
```

清单 15 显示了如何将 GONOPROXY 变量恢复为空，以及再次运行 go mod tidy 之前如何从 GOPROXY 中删除 direct。

**清单 16**

```bash
app/cmd/db imports
	github.com/Bhinneka/golib: cannot find module providing package github.com/Bhinneka/golib
app/cmd/db imports
	gopkg.in/rethinkdb/rethinkdb-go.v5: cannot find module providing package gopkg.in/rethinkdb/rethinkdb-go.v5
```

清单 16 显示了运行 go mod tidy 时来自 Go 工具的输出。你可以看到调用失败，因为 Go 工具找不到模块。

如果我将 direct 放回 GOPROXY 变量中会怎样？

```bash
$ unset GONOPROXY
$ export GOPROXY="http://localhost:3000,direct"
$ rm go.*
$ go mod init app
$ go mod tidy
```

清单 17 显示了如何再次将 direct 用于 GOPROXY 变量。

**清单 18**

```bash
go: finding github.com/Bhinneka/golib latest
go: finding gopkg.in/rethinkdb/rethinkdb-go.v5 v5.0.1
go: downloading gopkg.in/rethinkdb/rethinkdb-go.v5 v5.0.1
go: extracting gopkg.in/rethinkdb/rethinkdb-go.v5 v5.0.1
```

清单 18 显示了 Go 工具是如何再次工作的，并直接到每个 VCS 系统去获取模块。记住，如果返回任何其他状态代码（在 200、410 或 404 之外），Go 工具将失败。

### 其他场景

我决定不再继续使用其他只会导致 Go 工具失败的场景。如果你使用的是私有模块，那么你需要一个私有模块镜像，每个开发人员和构建机器上的配置都很重要，并且需要保持一致。私有模块镜像的配置需要与开发人员配置的内容相匹配，构建计算机也是如此。然后使用 GONOPROXY 和 GONOSUMDB 环境变量防止将私有模块的请求发送到任何 Google 服务器。如果你正在使用 Athens，它有特殊的配置选项来查找任何开发人员或构建计算机上的配置差异。

## VCS 认证问题

在回顾这篇文章的时候，[Erdem Aslan](https://twitter.com/Gladmir) 非常友好地为人们遇到的问题提供了一个解决方案。获取依赖项时的 Go 工具直接期望使用基于 https 的协议。在需要 VCS 认证的环境中，这可能是一个问题。Athens 可以帮助解决这个问题，但是如果你想确保直接调用不会失败，Erdem 为你的全局 git 配置文件提供了这些设置。

**清单 19**

```bash
[url "git@github.com:"]
insteadOf = "https://github.com"
  pushInsteadOf = "github:"
  pushInsteadOf = "git://github.com/"
```

## 总结

当你开始在自己的项目中使用模块时，请确保尽早决定使用哪个模块镜像。如果你有一个私有的 VCS 或者如果隐私是一个大问题，那么使用一个私有的模块镜像是你最好的选择。这将提供你需要的所有安全性、更好的抓取模块性能和最高级别的隐私。Athens 是运行私有模块镜像的好选择，因为它提供了模块缓存和校验和数据库代理。

如果你想检查 Go 工具是否遵守你的配置，并且所选择的模块镜像是否正确地代理了校验和数据库，那么 Go 工具有一个名为 go mod verify 的命令。此命令检查依赖项在下载后是否未被修改。它将检查本地模块缓存中的内容，在 1.15 版本，该命令可以检查 [vendor 件夹](https://github.com/golang/go/issues/27348)。

尝试这些配置，并找到最符合你需要的解决方案。

---

via: <https://www.ardanlabs.com/blog/2020/02/modules-04-mirros-checksums-athens.html>

作者：[William Kennedy](https://www.ardanlabs.com/)
译者：[polaris1119](https://github.com/polaris1119)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出