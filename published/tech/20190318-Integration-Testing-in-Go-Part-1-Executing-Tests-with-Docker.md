首发于：https://studygolang.com/articles/21759

# Go 语言中的集成测试：第一部分 - 用 Docker 执行测试

## 简介

> “测试会带来失败，而失败会带来理解。” —— Burt Rutan

Burt Rutan 是一名航空航天工程师，他设计了 Voyager，这是第一架在不停车或加油的情况下环球飞行的飞机。虽然 Rutan 不是软件工程师，但他的话充分说明了测试的重要性，甚至是软件测试。所有形式的软件测试都非常重要，无论是单元、集成、系统还是验收测试。但是，根据项目的不同，一种形式的测试可能比其他形式更有价值。换句话说，有时一种形式的测试可以比其他形式更好地反映软件的健康和完整性。

在开发 Web 服务时，我相信一组强大的集成测试可以比其他类型的测试更好地分析这个 Web 服务。集成测试是一种软件测试形式，用于测试代码与应用程序利用的依赖项（如数据库和消息传递系统）之间的交互。如果没有集成测试，很难信任 Web 服务的端到端操作。我之所以这么说，是因为在 Web 服务中的单元测试很少能提供与集成测试相媲美的信息量。

这是关于 Go 语言集成测试的三部曲系列中的第一篇文章。本系列中分享的想法、代码和流程旨在轻松应用到您正在处理的 Web 服务项目。在这篇文章中，我将向您展示如何配置 Web 服务项目以使用 Docker 和 Docker Compose 在没有预先安装 Go 的受限计算环境中运行 Go 测试和依赖项。

## 为什么使用 Docker 和 Docker Compose

众多开发者使用 Docker，是因为它能够在无需手动安装管理应用程序的前提下，在主机上加载应用程序。这意味着您可以加载复杂的软件，包括但不限于数据库（例如 Postgres），消息传递系统（例如 Kafka）和监控系统（例如 Prometheus）。所有这一切都是通过下载一组代表应用程序及其所有依赖项的镜像来完成的。

*注意：想要了解有关容器的更多信息，可以参考 Docker 专门用于定义容器的[网页](https://www.docker.com/resources/what-container)，该网页还重点说明了容器和虚拟机之间的异同。*

Docker Compose 是一种容器编排工具，有助于在一个沙箱内构建、运行并用网络连接一组容器。使用一个命令 `docker-compose up`，您可以使 Docker Compose 文件运行起来。在 Compose 文件中定义的所有服务将成为一组在其自己的网络沙箱中依照配置运行的容器。这与手动构建、运行和联网每个容器相比，是另一种使容器一起运行、相互通信并持久化数据的方式。

既然 Docker Compose 允许您将不同的应用组合到一起，并在一个网络沙箱中运行，您就可以只用一条命令做到启停一整套的应用。您甚至可以从一组应用中，手动挑选出个别应用来运行。这组应用可以被部署成一个独立的单元，通过 CI （集成开发）环境构建并测试。Docker Compose 最终将帮助确保您的应用在所有测试和部署的环境中保持一致。

*注意：想要了解更多关于 Docker Compose 的内容，请[点此](https://docs.docker.com/compose/overview/) 访问 Docker 官方网站对于 Docker Compose 的介绍。*

Docker 和 Docker Compose 的另一大好处是，它们能简化新的开发者加入一个项目时的交接过程。不需要关于如何安装、管理开发环境的复杂文档，新开发者只需要执行几条 Docker 和 Docker Compose 命令就可以开始工作了。如果在应用启动时，主机上没有所需的镜像，Docker CLI 会负责处理镜像的下载。

## 使用 Docker 和 Docker Compose 来运行测试

贯穿本系列的 Web 服务应用例子对外暴露简单的增删改查 REST API 接口，并使用 Postgres 数据库。这个项目在生产环境和测试中，都使用 Docker 来运行 Postgres 数据库。这个应用的测试要能够在一个已经安装了 Go 的本地开发环境以及一个没有 Go 的受限环境中运行。

下面给出的 Docker Compose 文件支持在上述的两种环境中运行这个项目的集成测试。本节中，我将分解我所采用的配置选项，并逐一解释。

*清单 1*

```yaml
version: '3'

networks:
  integration-tests-example-test:
    driver: bridge

services:
  listd_tests:
    build:
      context: .
      dockerfile: ./cmd/listd/deploy/Dockerfile.test
    depends_on:
      - db
    networks:
      - integration-tests-example-test
  db:
    image: postgres:11.1
    ports:
      - "5432:5432"
    expose:
      - "5432"
    environment:
      POSTGRES_USER: root
      POSTGRES_PASSWORD: root
      POSTGRES_DB: testdb
    RESTart: on-failure
    networks:
      - integration-tests-example-test
```

在清单 1 中，可以看到这个 Docker Compose 文件定义了需要运行测试的所需要项目服务。这个文件有三个主要的键值：`version`，`networks` 以及 `services`。其中 `version` 键值定义了正在使用的 Docker Compose 版本。而 `networks` 键值定义了一个或多个供特定服务使用的网络配置。`services` 键值则定义了要启动的容器和容器的配置。

*清单 2*

```yaml
networks:
  integration-tests-example:
    driver: bridge
```

如果将服务定义在同一个 Compose 文件中，按照默认设定，它们就会被自动地放在同一个网络中，因此可以相互之间通信。但最好的做法还是为服务创建一个网络，而不是使用默认网络。这顶层的 `networks` 配置定义了网络的名字以及所用的驱动，在这里用到的是 bridge 驱动。

bridge 驱动是 Docker 提供的默认驱动，它将创建一个私有的内部网络供容器进行通信。在 Compose 文件中的服务定义配置里，规定了这些服务将要使用这个新创建的网络。

*清单 3*

```yaml
services:
  listd_tests:
    build:
      context: .
      dockerfile: ./cmd/listd/deploy/Dockerfile.test
// ... omitted code …
  db:
// ... omitted code …
```

`services` 键有两个直接子键，分别是 `listd_test` 和 `db`。其中 `listd_tests` 用 dockerfile 的形式定义了它的镜像。而 `context` 键说明所有的主机目录都要相对于当前的工作目录，如这里定义的 `.`。

*清单 4*

```yaml
listd_tests:
    build:
      context: .
      dockerfile: ./cmd/listd/deploy/Dockerfile.test
    depends_on:
      - db
    volumes:
      - $PWD:/go/src/github.com/george-e-shaw-iv/integration-tests-example
```

`depends_on` 键会让 `listd_tests` 服务等待 `db` 服务启动之后再开始运行。除了明确启动的先后顺序，这个键还禁止了 `listd_tests` 服务独立于 `db` 服务运行。`volumes` 键告诉 Compose 在容器中将当前目录（记为 `$PWD`，Print Working Directory）挂载到 `/go/src/github.com/george-e-shaw-iv/integration-tests-example`，这将成为代码存放和测试的位置。

*清单 5*

```yaml
listd_tests:
    build:
      context: .
      dockerfile: ./cmd/listd/deploy/Dockerfile.test
    depends_on:
      - db
    networks:
      - integration-tests-example-test
```

最后，服务被赋予了一个网络，以便在沙箱内进行通信。这个网络最初在顶层的 `networks` 键中被定义，具体可见清单 2。

*清单 6*

```yaml
db:
    image: postgres:11.1
```

在下一个服务定义中的容器 `db`，用 [Docker Hub 中的镜像](https://hub.docker.com/_/postgres) `postgres:1.11` 来定义自身的容器镜像。Docker CLI 足够的智能，在本地机器上找不到镜像的话，会去 Docker Hub 镜像仓库寻找。

*清单 7*

```yaml
db:
    image: postgres:11.1
    ports:
      - "5432:5432"
```

出于安全考虑，默认情况下没有一个容器端口是可以通过主机访问的。这带来了一个问题，当本地运行集成测试时，如果集成的服务无法被访问，那测试将没有多少价值。这个 `prots` 键定义了从主机到容器的端口映射，形式如下： `" 主机端口 : 容器端口 "`。按照清单 7 中的定义，主机上的 5432 端口将被映射到 `db` 容器上，这个端口是 Postgres 在容器中默认的运行端口。

*清单 8*

```yaml
db:
    image: postgres:11.1
    ports:
      - "5432:5432"
    expose:
      - "5432"
```

正如容器端口不会默认暴露给主机一样，容器的端口也不会默认暴露给网络沙箱中的其他容器。即使这些容器是在同一个网络中运行也是如此。要想将端口暴露给运行在网络沙箱中的其他容器，就要通过 `expose` 键来配置。

*注意：在 `postgres:1.11` 镜像中，端口 5432 已经由容器创建者设置成对外暴露。但除非您亲自查看镜像的 Dockerfile，您很难获悉镜像的端口是否已经被设为暴露，所以最好还是定义 `expose` 键，哪怕它是多余的。*

*清单 9*

```yaml
db:
    image: postgres:11.1
    ports:
      - "5432:5432"
    expose:
      - "5432"
    environment:
      POSTGRES_USER: root
      POSTGRES_PASSWORD: root
      POSTGRES_DB: testdb
    RESTart: on-failure
    networks:
      - integration-tests-example-test
```

容器 `db` 需要的最后几个配置选项是 `environment`，`restart` 和 `networks`。与之前的服务定义类似，`networks` 的值为已被定义过的网络的名字。将 `restart` 键设为 `on-failure` 以确保服务会在运行中途崩溃时自动重启。`environment` 选项包含了注入到容器 shell 中的一系列环境变量。大多数主流应用的镜像，例如 postgres，都有用来配置其所提供的应用的环境变量。

## 运行测试

在 Docker Compose 文件准备好之后，下一步就是基于 `listd_tests` 服务中提到的 dockerfile 构建镜像。这个 dockerfile 定义了一个能够为整个服务运行集成测试的镜像。一旦镜像创建完成，测试就可以运行了。

### 构建一个能运行测试的镜像

为了构建一个能运行测试的镜像，在 dockerfile 中要定义四样东西：

获取一个带有最新稳定版 Go 的基础镜像。为 Go Modules 安装 `git`。将可测试的代码复制到容器中。运行测试。

让我们来分解这几个步骤，并对 dockerfile 需要涉及的命令进行分析。

*清单 10*

```dockerfile
FROM golang:1.12-alpine
```

清单 10 展示的是四步中的第一步。我选用的基础操作系统镜像是 `golang:1.11-alpine`。这个镜像预装了最新的稳定版 Go（在写这篇博文的时候）。

*清单 11*

```dockerfile
FROM golang:1.11-alpine

RUN set -ex; \
    apk update; \
    apk add --no-cache Git
```

因为 Alpine OS 是非常轻量级的，您必须在基础镜像之上，手动安装 `git` 依赖。清单 11 展示的是第二步，将 `git` 添加到镜像中，为了后续使用 Go Modules。其中 `apk update` 命令要在添加 `git` 之前运行，以确保安装的是最新版本的 `git`。如果您的项目恰好要使用 `cgo`，那么您必须手动安装 `gcc` 以及它的依赖库。

*清单 12*

```dockerfile
FROM golang:1.12-alpine

RUN set -ex; \
    apk update; \
    apk add --no-cache Git

WORKDIR /go/src/github.com/george-e-shaw-iv/integration-tests-example/
```

为了方便使用，清单 12 中设置工作目录为 `/go/src/github.com/george-e-shaw-iv/integration-tests-example/`，以便在后续的指令中可以使用基于这个目录的相对路径，这也是容器中的 `$GOPATH`。过程中的第三步，将测试代码拷贝到容器里，实际已经在清单 4 中完成，含有测试代码的目录已经被挂载到容器中。

*清单 13*

```dockerfile
FROM golang:1.12-alpine

RUN set -ex; \
    apk update; \
    apk add --no-cache Git

WORKDIR /go/src/github.com/george-e-shaw-iv/integration-tests-example/

CMD CGO_ENABLED=0 Go test ./...
```

最后，清单 13 代表第四步，运行测试。这是用一条 `CMD` 指令 `go test ./...` 实现的。

测试使用 `CGO_ENABLED=0` 作为内联环境变量运行，因为示例项目的测试没有用到 cgo 并且 Alpine 基础镜像也不包含 C 编译器。即使您的项目中没有 cgo 代码，也必须以这种方式禁用 cgo，因为如果启用了 cgo，Go 仍会尝试使用标准 C 库来执行某些网络任务。

*注意：定义这个能运行 Go 测试的镜像的完整 Dockerfile 代码可以访问 [这一网站](https://github.com/george-e-shaw-iv/integration-tests-example/blob/master/cmd/listd/deploy/Dockerfile.test) 获取。*

现在定义镜像的 dockerfile 已经写好了，下面的 Docker Compose 命令能够启动 `listd_test` 和 `db` 服务，它们将运行所有的集成测试并反馈结果。

*清单 14*

```sh
docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
```

其中 `--abort-on-container-exit` 标志位不能省略，否则其他包含这个集成服务的容器会在测试结束后继续运行。

### 善后工作

*清单 15*

```makefile
test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down --volumes
```

停止并删除容器、持久卷和网络是完成测试后非常重要的一步，而这一步经常会被忽略。由上次测试的遗留数据所导致的测试失败，其错误的原因难以发现，因此不容易解决，但这个问题是非常容易避免的。想要避免这个问题发生，我创建了一个简单的 `makefile` 规则 `test`，如清单 15 所示，无需任何人力参与也能创建、运行并清除容器。

*清单 16*

```makefile
test-db-up:
	docker-compose -f docker-compose.test.yml up --build db

test-db-down:
	docker-compose -f docker-compose.test.yml down --volumes db
```

在清单 15 中列出的规则最好在受限的环境中使用，因为它们将 compose 文件中的两个服务都启动了。为了在本地测试中达到相同的效果，`test-db-up` 规则可以在运行集成测试之前使用，在所有测试执行完成后使用 `testdb-down`。

## 结论

在本片博文中，我向您展示了如何用 Docker 和 Docker Compose 部署 Web 服务项目。我所提及的这些文件能够使您在没有预装 Go 的受限环境下，运行 Go 测试和依赖项。在本系列的下一部分，我将展示为 Web 服务配置测试套件所需的 Go 代码，这将是编写富有洞察力的集成测试的基础。

*注意：本系列文章用到的例子都来自 [这个仓库](https://github.com/george-e-shaw-iv/integration-tests-example)。*

---

via: https://www.ardanlabs.com/blog/2019/03/integration-testing-in-go-executing-tests-with-docker.html

作者：[George Shaw](https://github.com/george-e-shaw-iv/)
译者：[Mockery-Li](https://github.com/Mockery-Li)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
