首发于：https://studygolang.com/articles/25290

# 从 Travis 迁移至 GitHub Actions

周末的时候，我决定将我 Go 语言的开源项目 [Flipt](https://github.com/markphelps/flipt) 的 CI 流程从 TravisCI 转移到 GitHub Actions，我想要替换我现有的 CI，并尝试使用新的 GitHub Actions 将手动发版过程自动化。

*说明*：我在 GitHub 工作，但不在 Actions 团队。我想在我的开源项目中配置 Actions，并且不从 Actions 团队或 GitHub 的任何人那里获得任何帮助。我没有被 Github 的同事要求写这篇文章，我的目的很简单，以一个用户的经验来使用这个平台。仅代表个人观点和想法。

不用说，经过我几个小时的调试，我成功了[twitter 链接](https://twitter.com/mark_a_phelps/status/1172935552947118081?ref_src=twsrc%5Etfw%7Ctwcamp%5Etweetembed%7Ctwterm%5E1172935552947118081&ref_url=https%3A%2F%2Fmarkphelps.me%2F2019%2F09%2Fmigrating-from-travis-to-github-actions%2F)。

![推特截图](https://raw.githubusercontent.com/studygolang/gctt-images/master/migrating-from-travis-to-action/BDy3YCr5ZwgEbdL.png)

## 管道

我不打算对比 workflow (流程) 、job(任务)、step(步骤) 等细节， GitHub 有广泛的文档来介绍 Actions 的 [用法](https://help.github.com/en/articles/workflow-syntax-for-github-actions) 和 [概念](https://help.github.com/en/articles/about-github-actions#core-concepts-for-github-actions)，我认为我想要的是很普通的一个 CI/CD 流程：

. push 代码到分支后运行一些单元测试，最好能够使用 Go 的多个版本
. 在 PR 上，我还希望运行一些更广泛的集成测试，用来测试面向公众的 API 和 CLI
. 推送 tag 后，我想触发 [goreleaser](https://github.com/goreleaser/goreleaser) 来构建一个 Docker 镜像并推送到 [Docker Hub](https://hub.docker.com/r/markphelps/flipt)，同时打包一个发版的压缩文件
. 在新版本更新文档时更新 [文档网站](https://flipt.dev/)

前两个步骤主要的 TravisCI 工作是在这个 [config 文件](https://github.com/markphelps/flipt/blob/90bafa834aec29cdaa3620b8ea30aa89466fe7d0/.travis.yml)配置的，虽然有一些差异:

1. 我只测试了 Go 一个版本 (1.12.x)，我知道我可以使用 travis-ci 的 [matrix](https://docs.travis-ci.com/user/build-matrix/)设置来测试多个版本，只是我从来没有这样去用。
2. 我只针对 PR 在 Postgres DB 实体环境上运行测试，

我缺少的是用于实际构建发版和更新文档的 CD (持续部署)部分。我在本地机器上运行脚本依赖于设置一些需要保密的环境变量，依然是一个手动操作过程。这不是最理想的情况。

## 容易实现的目标

我创建的第一个 action 实际上是自动更改文档部分。这一部分会被移动到管道作业流的最后一步，但也是能正常运行的最简单的一步。

它主要由两个文件组成，一个 [Dockerfile](https://github.com/markphelps/flipt/blob/4157e9b154a01b09a4eb60a8e43484cd3928fc89/.github/actions/publish-docs/Dockerfile) 用于安装必要的依赖项，另一个[脚本](https://github.com/markphelps/flipt/blob/master/.github/actions/publish-docs/entrypoint.sh) 用于运行构建和部署步骤。
我使用 [mkdocs](https://www.mkdocs.org/) 来构建文档并发布到 [GitHub pages](https://help.github.com/en/articles/creating-project-pages-using-the-command-line)。

我(最终)把它连接起来作为发布工作流程的最后一步:

```bash
name: Publish Docs
uses: ./.github/actions/publish-docs
env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

它将通知 Actions 我希望使用 [local action](https://help.github.com/en/articles/workflow-syntax-for-github-actions#example-using-action-in-the-same-repository-as-the-workflow) 存在的 action，并设置 `GITHUB_TOKEN` 环境变量，该变量是推送到 GitHub pages 所必需的。

## 这些繁琐的测试

接下来我做的是让管道的单元测试部分运作起来。因为 [Flipt](https://github.com/markphelps/flipt) 是一个服务端应用程序，所以我目前只针对 Linux 环境，因此我不需要测试 Windows 或 MacOS 环境。虽然我知道 Actions 很酷并且也支持 😉。

然而，我确实希望能够使用多个版本的 Go 进行测试(撰写本文时为 1.12 和 1.13 )。Actions 的 [matrix strategy]矩阵策略特性让这一切变得超级简单。

对于我的 workflow 工作流，它看起来像这样:

```bash
test:
  name: Test
  runs-on: ubuntu-latest
  strategy:
    fail-fast: false
    matrix:
      go: ['1.12', '1.13']
```

这里设置两个作业并行运行，运行下面的所有步骤，其中一个 `{{ matrix.go }}` 设置为 `1.12`，另一个设置为 `1.13`。

稍后在工作流文件中，我创建了一个步骤，这些值来将被用来在虚拟机上安装可用版本的 Go:

```bash
steps:
- name: Setup Go
  uses: actions/setup-go@v1
  with:
    go-version: ${{ matrix.go }}
  id: go
```

它使用 [actions/setup-go](https://github.com/actions/setup-go) action 来安装我们指定的 Go 版本。这很酷。

实际上，我几乎立刻就看到了使用多个 Go 版本运行  测试的好处，因为 Go 1.13 增 加了一些新功能，我的一些测试代码已经无法通过。

查看发布说明:

> 测试 flags 标识现在被注册到新的 Init 函数中，该函数会在测试生成的主函数调用。因此，测试 flags 标识现在只在运行测试二进制文件时注册，并且包名为 flag。包初始化期间的解析可能导致测试失败。

说明太长不建议阅读。我曾经在我的一个 [测试](https://github.com/markphelps/flipt/blob/fdf45bff66c325d702b54ae334e53ae8e3cac176/storage/db_test.go#L88) 中使用 init 函数来打开一些调试日志，如果设置了一个标志的话。事实证明这在 Go 1.13.1 会出现 [问题](https://github.com/golang/go/issues/31859)。

我不认为在我真正尝试更新 Flipt 到 Go 1.13 之前能发现这个问题，目前我能够通过完全的测试在早期发现这个问题，这很酷。

## 不愿多谈的问题

我在前面提到过，我还希望使用正式环境的 Postgres 数据库中运行单元测试。这是因为 Flipt 同时支持 [SQLite 和 Postgres](https://github.com/markphelps/flipt#database://github.com/markphelps/flipt#databases)，我希望对代码进行同等的测试。

幸运的是运行 Actions 构建操作的 Ubuntu 虚拟机似乎已经安装了 SQLite 所需的库，但是它们似乎没有安装 Postgres，这点与 Travis 不同。你可以在[文档](https://help.github.com/en/articles/software-in-virtual-environments-for-github-actions)中看到每个 VM 的所有已安装软件/库的列表。

这意味着我需要想办法找到一个 Postgres 服务来运行我的构建，这样我才能完成测试。

我最初尝试使用的一个步骤是使用 Docker 容器内使用 `docker run` 命令来运行 Postgres 。然而我很快发现 Actions 有一个针对这类问题的内置解决方案 - [services](https://help.github.com/en/articles/workflow-syntax-for-github-actions#jobsjob_idservices)!

事实证明，`services` 指令正是我所需要的:

```bash
services:
  postgres:
    image: postgres:11
    ports:
      - 5432:5432
    env:
      POSTGRES_DB: flipt_test
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ''
```

这与我在 Docker 中通过在容器中运行 Postgres 所做的事情是一样的，但是这里是通过 Actions 来管理的。

## Bats 和 REST API

更进一步，测试管道一直存在于集成测试里。在这我希望能够验证 Flipt 能完成 `public API` 的方面所需要做的事情。我希望 Flipt 的 REST API 以及它的 CLI 都是公开的，因此应该对它们进行彻底的测试并防止版本回退。

幸运的是使用诸如 [bats](https://github.com/sstephenson/bats/) 之类的工具，CLI 的测试变得相当容易。我有一些现有的正在使用的 bats 测试 [脚本代码](https://github.com/markphelps/flipt/blob/4157e9b154a01b09a4eb60a8e43484cd3928fc89/script/test/cli.bats) 运行在 Travis 构建中，所以我只需要找到一种方法让他们运行在 Actions 上即可。

同样，看起来 Actions 的虚拟机并没有安装 bats，但是 GitHub Actions 的 fork 版本似乎已经意识到到了这一点，可以构建了一个你可以在工作流程中引用的 [bats action](https://github.com/actions/bin/tree/master/bats)。我就是这么做的:

```bash
- name: Test CLI
  uses: actions/bin/bats@master
  with:
    args: ./script/test/*.bats
```

在 Linux VM 中的构建二进制文件之前，我还有一个步骤，由这个 bats action 来调用它来测试 CLI 输入/输出。

集成测试的最后一部分是测试 REST API。我之前发现了一个很酷的叫 [shakedown](https://github.com/robwhitby/shakedown) 的 bash 库，它让 HTTP 测试变得轻而易举。

因为 VM 虚拟机似乎已经安装了所需的依赖，我最初尝试在原来的 VM 上运行这些测试，但是我在彻底地完成运行测试时遇到了一些问题，所以我决定迁移到一个“干净的环境” - 只在容器中运行测试。

在对不同的基础 Docker 镜像进行了一些的修改并安装了必要的依赖项之后，我最终通过安装正确的工具构建了自己的 action，从而使 shakedown 测试能够正常工作。

## 愉快的发版

最后，管道的最后一部分是建立发版:

. 为 *nix 创建 tarball 文件
. 创建一个 Docker 镜像
. 推送 tarball 文件到 GitHub 并且发布新的版本
. 创建 Tag 版本推送 Docker 镜像到 Docker Hub

幸运的是 [goreleaser](https://goreleaser.com/) 已经为此做了 100% 工作! 我所需要做的就是在管道中的最后一步为它提供所需的环境变量，并使用正确的参数调用它。

我已经在本地[使用脚本](https://github.com/markphelps/flipt/blob/c82b47b7522caf80bc3f5219ea62e9e37c416dd2/script/build/release)运行，这意味着在调用脚本之前，我必须在本地机器上设置 `GITHUB_TOKEN`、`DOCKER_USERNAME` 和 `DOCKER_PASSWORD`。

为了将这个过程转移到 GitHub Actions 操作，我需要一种安全的方法来存储这些值并将它们注入到工作流中。幸运的是 GitHub 也为我们提供了对[保密](https://help.github.com/en/articles/virtual-environments-for-github-actions#creating-and-using-secrets-encrypted-variables)的  支持:

```bash
- name: Release
  run: ./script/build/release
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    DOCKER_USERNAME: ${{ secrets.DOCKER_USERNAME }}
    DOCKER_PASSWORD: ${{ secrets.DOCKER_PASSWORD }}
```

这段代码展示了如何引用 `secrets` 并将它们设置为脚本运行时使用的环境变量。这使我可以通过 Actions 运行 goreleaser，而不必担心这些保密文件被暴露在日志或仓库本身中。

## 小结

如果你决定迁移你的 pipelines 管道，这里有一些 ProTips ™，可以帮助你:

1. **从简单的开始**。不要试图一下就替换掉整个 CI/CD 方案。看看是否有一些可以先迁移的非关键任务。
2. **保证现有 CI 系统正常运行**。这个不用说，不要删除你的 `travis.yml` 文件，直到你确信新的 Actions 设置一切运行正常。
3. **优先寻找现有的解决方案**。Actions 社区中已经有很多很酷的东西，包括 [github/ Actions](https://github.com/actions) 项目。在尝试创建自己的特定任务之前先查看一下，你会发现有可能已经存在了。
4. **阅读文档**。认真地说文档有丰富的信息，可能会帮助你弄清楚如何去做你想做的事情，它能解决问题并且省下很多时间。

正如你可能猜到的那样，使用 Actions 设置完美的 CI/CD 管道流需要一些工作，这主要需要阅读文档。每当我遇到困难的时候，最终都是因为我不理解这个系统是如何工作的。我欣赏 GitHub Actions 提供的扩展性和强大功能，因为你可以正确地用它做任何事情。这伴随着需要学习稍微不同的语法和一些规范，但我认为好处远远大于缺点。

我引用的所有工作流文件都可以在[这里](https://github.com/markphelps/flipt/tree/master/.github/workflows)找到。

---

via: https://www.markphelps.me/2019/09/migrating-from-travis-to-github-actions/

作者：[Mark Phelps](https://www.markphelps.me/)
译者：[M1seRy](https://github.com/M1seRy)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
