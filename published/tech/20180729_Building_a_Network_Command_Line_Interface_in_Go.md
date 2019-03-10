首发于：https://studygolang.com/articles/18795

# 在 Go 中构建网络命令行界面

在本文中，我们将使用 `Github` 上提供的软件包 `urfave/cli` 在 Go 中构建一个非常简单的命令行界面，软件包位于 https：//github.com/urfave/cli。

我最近在各种托管服务提供商中进行了一次或两次域名迁移，并认为构建一个可用于查询网站名称服务器，CNAME，IP 地址等内容的工具或程序是一个很酷的主意。

本教程的总体目标是让您了解如何构建自己的 CLI，这些 CLI 可以执行各种其他操作，例如网络监视，图像处理等。

> 注 - 可在此处找到本教程的完整代码：[TutorialEdge/Go/go-cli-tutorial](https://github.com/TutorialEdge/Go/tree/master/go-cli-tutorial)

## 热门项目

`Golang` 正在大规模普及，我们已经看到像 `Hashicorp` 这样的大型企业公司采用了许多不同工具和系统的语言。而且有充分的理由相信 `Go` 的设计非常适合这些应用程序，并且能够其能够轻松地为所有主要平台跨平台编译成二进制可执行文件将会是一个巨大的胜利。

## 视频教程

如果您更喜欢通过视频媒体进行学习，请随时查看本教程：[视频地址](https://www.youtube.com/embed/i2p0Snwk4gc?ecver=2)

## 入门

让我们在我们的计算机上创建一个新目录，命名为 `go-cli` 或者其他。我们将为我们的项目创建一个类似于此的目录结构：

```
go-cli/
- pkg/
- cmd/my-cli/
- vendor/
- README.md
- ...
```

> 注意 - 此结构遵循 Github 上广泛接受的[Go 项目布局](https://github.com/golang-standards/project-layout) 指南。

## 进入代码

现在我们已经有了基本的项目结构，我们可以开始编写我们的应用程序了。首先，我们需要在 `cmd/my-cli/` 的目录内创建一个新文件 `cli.go`。我们编写一种非常简单的语句 `Hello World`，并将其作为我们未来开发的基础。

```go
// cmd/my-cli/cli.go
package main

import (
  "fmt"
)

func main() {
  fmt.Println("Go CLI v0.01")
}
```

然后我们可以  在项目的根目录通过输入以下命令尝试运行它：

```
➜ Go run cmd/my-cli/cli.go
Go CLI v0.01
```

很好，我们已经对新 CLI 的要素进行了理解，让我们现在看看我们如何添加一些命令并使其有用。

## 我们的第一命令

由于我们将使用 `urfave/cli` 软件包，我们需要在本地下载此软件包才能使用它，我们可以通过一个简单的 `go get` 命令来实现：

```
go get Github.com/urfave/cli
```

现在我们有了必要的包，让我们更新我们的 `cli.go` 文件并使用这个包为我们创建一个新的 CLI 应用程序：

```go
// cmd/my-cli/cli.go
import (
  "log"
  "os"

  "github.com/urfave/cli"
)

func main() {
  err := cli.NewApp().Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
```

当我们现在运行它时，你会看到它丰富了我们的程序响应并添加了诸如版本，如何使用 cli 以及拥有的各种命令之类的东西。

```shell
➜  Go run cmd/my-cli/cli.go
NAME:
   cli - A new cli application

USAGE:
   cli [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

太棒了，这很快就开始看起来像一个更精美的项目，而不仅仅是一个小方案项目！

我们现在可以开始添加我们自己的 `Commands`。这些命令中的每一个都将与我们的一个测试相匹配，因此我们将有一个命令：`ns` 当提供一个 url 执行时将开始查找该 url 主机的域名服务器。

我们的最终命令列表将如下所示：

- ns - 将检索域名服务器
- cname - 将查找给定主机的 CNAME
- mx - 将查找给定主机的邮件交换记录
- ip - 将查找给定主机的 IP 地址。

很好很简单，让我们开始创建我们的第一个命令：

```go
package main

import (
    "fmt"
    "log"
    "net"
    "os"

    "github.com/urfave/cli"
)

func main() {
    app := cli.NewApp()
    app.Name = "Website Lookup CLI"
    app.Usage = "Let's you query IPs, CNAMEs, MX records and Name Servers!"

    // We'll be using the same flag for all our commands
    // so we'll define it up here
    myFlags := []cli.Flag{
        cli.StringFlag{
            Name:  "host",
            Value: "tutorialedge.net",
        },
    }

    // we create our commands
    app.Commands = []cli.Command{
        {
            Name:  "ns",
            Usage: "Looks Up the NameServers for a Particular Host",
            Flags: myFlags,
            // the action, or code that will be executed when
            // we execute our `ns` command
            Action: func(c *cli.Context) error {
                // a simple lookup function
                ns, err := net.LookupNS(c.String("url"))
                if err != nil {
                    return err
                }
                // we log the results to our console
                // using a trusty fmt.Println statement
                for i := 0; i < len(ns); i++ {
                    fmt.Println(ns[i].Host)
                }
                return nil
            },
        },
    }

    // start our application
    err := app.Run(os.Args)
    if err != nil {
        log.Fatal(err)
    }
}
```

然后我们可以通过输入以下命令来尝试运行：

```
go run cmd/my-cli/cli.go ns --url tutorialedge.net
```

然后，这应该返回我的站点的域名服务器并在终端中打印出来。我们还可以运行 help 命令，它将向我们展示如何在 CLI 中使用我们的新命令。

## 查找 IP 地址

我们的所有命令定义在我们的程序中看起来都非常相似，除了我们如何打印出结果。`net.LookupIP()` 函数返回一系列 IP 地址，因此我们必须迭代这些地址以便以一种很好的方式打印它们：

```json
{
    Name:  "ip",
    Usage: "Looks up the IP addresses for a particular host",
    Flags: myFlags,
    Action: func(c *cli.Context) error {
        ip, err := net.LookupIP(c.String("host"))
        if err != nil {
            fmt.Println(err)
        }
        for i := 0; i < len(ip); i++ {
            fmt.Println(ip[i])
        }
        return nil
    },
},
```

## 查找 CNAME

然后我们可以添加我们的 `cname` 命令，它将使用 `net.LookupCNAME()` 接受传入 `host` 中的数据并返回一个 `CNAME` 字符串，接下来我们可以打印出来：

```json
{
    Name:  "cname",
    Usage: "Looks up the CNAME for a particular host",
    Flags: myFlags,
    Action: func(c *cli.Context) error {
        cname, err := net.LookupCNAME(c.String("host"))
        if err != nil {
            fmt.Println(err)
        }
        fmt.Println(cname)
        return nil
    },
},
```

## 查找 MX 记录

最后，我们希望能够查询给定主机的 `Mail Exchange` 记录，我们可以通过使用 `net.LookupMX()` 函数并传入 `host` 数据来实现。这将返回一片 mx 记录，和 IP 一样，我们必须迭代才能打印出来

```json
{
    Name:  "mx",
    Usage: "Looks up the MX records for a particular host",
    Flags: myFlags,
    Action: func(c *cli.Context) error {
        mx, err := net.LookupMX(c.String("host"))
        if err != nil {
            fmt.Println(err)
        }
        for i := 0; i < len(mx); i++ {
            fmt.Println(mx[i].Host, mx[i].Pref)
        }
        return nil
    },
},
```

## 构建我们的 CLI

现在我们已经启动并运行了一个基本的 CLI，现在是时候构建它，以便我们可以直接使用它。

```
go build cmd/my-cli/cli.go
```

这应该编译一个名称为 `cli` 可执行文件，我们可以这样运行它：

```shell
$ ./cli help
NAME:
   Website Lookup CLI - Let's you query IPs, CNAMEs, MX records and Name Servers!

USAGE:
   cli [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     ns       Looks Up the NameServers for a Particular Host
     cname    Looks up the CNAME for a particular host
     ip       Looks up the IP addresses for a particular host
     mx       Looks up the MX records for a particular host
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the versio
```

如您所见，我们的所有命令都已成功列在输出的 `COMMANDS` 中。

## 结论

在本教程中，我们使用 `urface/cli` 软件包成功构建了一个非常简单但有效的 CLI。该 CLI 可以非常简单的针对任何主要操作系统进行跨平台的编译，并且它具有您期望从生产级命令行界面获得的所有功能。

> 注意 - 如果您想了解网站上的最新文章和更新，请随时在 Twitter 上关注我：[@Elliot_f](https://twitter.com/elliot_f)

---

via: https://tutorialedge.net/golang/building-a-cli-in-go/

作者：[Elliot Forbes](https://tutorialedge.net/about/)
译者：[lovechuck](https://github.com/lovechuck)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
