首发于：https://studygolang.com/articles/17780

# 迁移到 mod 只需 3 个步骤

本文的目的是演示如何轻松地将带有 vendor 目录的旧项目迁移到使用 Go mod 的新项目。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/migrating-to-go/1_a-NrrpFPmj-_JQGulalCdQ.png)

我不打算一一列举相对于使用依赖关系关联工具，使用 Go modules 的优势，因为在互联网上有几篇关于这方面的文章。另一方面，**我将指导您如何迁移项目**。

## 检查最新的 Golang 版本

你可以问我，为什么要用最新的？因为如果我们都是软件爱好者，我们应该渴望测试最新的技术！ ( 顺便说一下，你可以用 Golang 1.11.X，但是您应该想知道为什么不使用最新的可用版本…… )

到[这里](https://studygolang.com/dl)，下载你所使用的操作系统的发行版。

## 找到要迁移的项目

在本文中，我将迁移一个我几个月前工作过的个人项目。在这个项目中，我使用 Glide 来管理依赖项。你可以对你的任何项目做同样的事情。

不要担心 mollydb 做了什么，因为我们只需要理解如何迁移现有的项目。

```shell
git clone https://github.com/wesovilabs/mollydb.git
cd mollydb
git checkout -b feature/using-go-mods
```

## 调整项目结构

这个项目是用一个 src 文件夹构建的，该文件夹包含一个子文件夹 mollydb，在这个子文件夹中有一个 vendor 目录用来存依赖项。

`src > mollydb > vendors`

这是我发现的唯一不依赖于我的全局路径来创建项目的方法……

我们将删除 vendor 目录，并将 src/mollydb 中的内容移动到项目的根目录。

```shell
rm -rf src/mollydb/vendor
mv src/mollydb/* .
```

我们可以运行如下命令

```
go mod init mollydb
```

Go 足够聪明，而且它会将 glide.lock 中的依赖项写入到 go.mod 文件中。

go：创建新的 go.mod 文件：mollydb module

go：从 glide.lock 复制需求

而且 go.mod 文件中的内容看起来和下面的一样

```
module mollydb
require (
  gitHub.com/boltdb/bolt v0.0.0 – 20180302180052-fd01fc79c553
  gitHub.com/fsnotify/fsnotify v1.4.7
  gitHub.com/go-yaml/yaml v0.0.0 – 20140922213225-bec87e4332ae
  gitHub.com/graphql-go/graphql v0.0.0 – 20180324214652 – 8ab5400ff77c
  gitHub.com/graphql-go/handler v0.0.0 – 20180312211735-df717460db9a
  gitHub.com/graphql-go/relay v0.0.0 – 20171208134043 – 54350098cfe5
  golang.org/x/net v0.0.0 – 20180320002117 – 6078986fec03
  golang.org/x/sys v0.0.0 – 20180318190847 – 01acb38716e0
  gopkg.in/yaml.v2 v2.1.1
)
```

无论如何，如果我们删除 glide 的配置文件，创建了 go.mod 文件，并且运行下面的命令。

```
go mod init mollydb
go mod tidy
```

go.mod 文件就会生成，因为 go mod 检查了我们的 Go 文件

```
module mollydb
require (
  gitHub.com/fsnotify/fsnotify v1.4.7
  gitHub.com/go-chi/chi v3.3.3+incompatible
  gitHub.com/graphql-go/graphql v0.7.7
  gitHub.com/graphql-go/handler v0.2.2
  gitHub.com/graphql-go/relay v0.0.0 – 20171208134043 – 54350098cfe5
  gitHub.com/kr/pretty v0.1.0 // indirect
  gitHub.com/sirupsen/logrus v1.2.0
  gitHub.com/stretchr/testify v1.2.2
  golang.org/x/net v0.0.0 – 20181217023233-e147a9138326 // indirect
  golang.org/x/text v0.3.0 // indirect
  gopkg.in/yaml.v2 v2.2.2
)
```

我们只需要运行下面的命令来验证项目是否像以前那样工作。

> go run main.go

## 所有的都成功了

我们的项目被迁移了 !!

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/migrating-to-go/0_AxqFfdrPxy4oqeVi.png)

---

via: https://medium.com/@ivan.corrales.solera/migrating-to-go-mod-in-just-3-steps-6b6a07a04640

作者：[Iván Corrales Solera](https://medium.com/@ivan.corrales.solera)
译者：[wumansgy](https://github.com/wumansgy)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
