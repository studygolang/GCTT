已发布：https://studygolang.com/articles/12432

# 为 Go Web-apps 编写 Dockerfiles 的终极指南

你或许想在 Docker 中使用 Go，原因有：

1. 如果你想在 Kubernetes 上运行，打包为镜像是必须的（就像我一样）
2. 你不得不在同一台机器上运行不同的 Go 版本
3. 开发和生产都需要精确的、可复制的、可共享的和确定的环境
4. 你需要快速和简单的方式来构建和部署编译好的二进制文件
5. 你想快速开始（任何安装了 Docker 的人都可以直接开始编写代码而不需要设置其他依赖或 `GOPATH` 环境变量）

恭喜你，你来对地方了。

我们将逐步构建一个基本的 Dockerfile，包括**实时重载**和**包管理**，然后进行扩展，创建一个高度**优化**的生产版的镜像，其大小缩减了 100 倍。如果你使用 CI/CD 系统，镜像大小可能无关紧要，但是当 `docker push` 和 `docker pull` 时，一个精简的镜像肯定会有帮助。

如果你只想要最终的代码，请看 [GitHub](https://github.com/shahidhk/go-docker/blob/master/src/main.go)。

```bash
FROM golang:1.8.5-jessie as builder
# install xz
RUN apt-get update && apt-get install -y \
    xz-utils \
&& rm -rf /var/lib/apt/lists/*
# install UPX
ADD https://github.com/upx/upx/releases/download/v3.94/upx-3.94-amd64_linux.tar.xz /usr/local
RUN xz -d -c /usr/local/upx-3.94-amd64_linux.tar.xz | \
    tar -xOf - upx-3.94-amd64_linux/upx > /bin/upx && \
    chmod a+x /bin/upx
# install glide
RUN go get github.com/Masterminds/glide
# setup the working directory
WORKDIR /go/src/app
ADD glide.yaml glide.yaml
ADD glide.lock glide.lock
# install dependencies
RUN glide install
# add source code
ADD src src
# build the source
RUN go build src/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main src/main.go
# strip and compress the binary
RUN strip --strip-unneeded main
RUN upx main

# use a minimal alpine image
FROM alpine:3.7
# add ca-certificates in case you need them
RUN apk add --no-cache ca-certificates
# set working directory
WORKDIR /root
# copy the binary from builder
COPY --from=builder /go/src/app/main .
# run the binary
CMD ["./main"]
```

假设我们的应用叫 `go-docker`，下面是项目的结构。所有的源码都在 `src` 目录下，`Dockerfile` 跟它在同一级目录。 `main.go` 定义了一个 web-app 并监听 8080 端口。

```
go-docker
├── Dockerfile
└── src
    └── main.go
```

## 最简单的版本

```bash
FROM golang:1.8.5-jessie
# create a working directory
WORKDIR /go/src/app
# add source code
ADD src src
# run main.go
CMD ["go", "run", "src/main.go"]
```

我们使用 `debian jessie` 版本的 golang 镜像，因为像 `go get` 这样的命令需要安装有 `git` 等工具。对于生产版本，我们会用更加轻量的版本，如 `alpine`。

构建并运行该镜像：

```bash
$ cd go-docker
$ docker build -t go-docker-dev .
$ docker run --rm -it -p 8080:8080 go-docker-dev
```

成功后可以通过 `http://localhost:8080` 来访问。按下 `Ctrl+C` 可以中断服务。

但这并没有多大意义，因为每次修改代码时，我们都必须构建和运行docker 镜像。

一个更好的版本是将源代码挂载到 docker 容器中，并使用容器内的 shell 来停止和启动 `go run`。

```bash
$ cd go-docker
$ docker build -t go-docker-dev .
$ docker run --rm -it -p 8080:8080 -v $(pwd):/go/src/app \
             go-docker-dev bash
root@id:/go/src/app# go run src/main.go
```

这个命令会提供一个 shell，我们可以在里面执行 `go run src/main.go` 以启动服务。我们可以在宿主机上编辑 `main.go` 并重新运行该命令来查看变化，因为现在源代码已经直接挂载到了容器中。

但是，如何管理包呢？

## 包管理和镜像分层
[Go 的包管理](https://github.com/golang/go/wiki/PackageManagementTools) 仍处在实验阶段。有很多工具可以选择，但是我最喜欢的是 [Glide](https://glide.sh/)。我们将在容器中安装 Glide 并使用它。

在 `go-docker` 项目中新建两个文件 `glide.yaml` 和 `glide.lock`：

```bash
$ cd go-docker
$ touch glide.yaml
$ touch glide.lock
```

按照下面所示修改 Dockerfile 并构建一个新的镜像：

```bash
FROM golang:1.8.5-jessie
# install glide
RUN go get github.com/Masterminds/glide
# create a working directory
WORKDIR /go/src/app
# add glide.yaml and glide.lock
ADD glide.yaml glide.yaml
ADD glide.lock glide.lock
# install packages
RUN glide install
# add source code
ADD src src
# run main.go
CMD ["go", "run", "src/main.go"]
```

如果你观察比较细致，你会发现 `glide.yaml` 和 `glide.lock` 是分开添加的（并没有用 `ADD . .`），这样会导致有单独分离的层。将包管理分离为单独的层，可以充分利用 Docker 层的缓存，并且只有当对应的文件发生变化才会导致重新编译，比如：新增或删除了一个包。因此，`glide install` 不会在每次修改了代码之后都去执行。

让我们进入容器的 shell 安装一个包：

```bash
$ cd go-docker
$ docker build -t go-docker-dev .
$ docker run --rm -it -v $(pwd):/go/src/app go-docker-dev bash
root@id:/go/src/app# glide get github.com/golang/glog
```

Glide 会将所有包安装到 `vendor` 目录，该目录可以被 `gitignored` 和 `dockerignored`。使用 `glide.lock` 来锁定某个包的版本。要安装（或重新安装）`glide.yaml` 中提到的所有包，执行：

```bash
$ cd go-docker
$ docker run --rm -it -p 8080:8080 -v $(pwd):/go/src/app \
             go-docker-dev bash
root@id:/go/src/app# glide install
```

现在 `go-docker` 目录有所增长：

```
.
├── Dockerfile
├── glide.lock
├── glide.yaml
├── src
│   └── main.go
└── vendor/
```

## 实时重载
[codegangsta/gin](https://github.com/codegangsta/gin) 是我最喜欢的实时重载工具。它简直就是为 Go web 服务而生的。我们使用 `go get` 来安装 gin：

```bash
FROM golang:1.8.5-jessie
# install glide
RUN go get github.com/Masterminds/glide
# install gin
RUN go get github.com/codegangsta/gin
# create a working directory
WORKDIR /go/src/app
# add glide.yaml and glide.lock
ADD glide.yaml glide.yaml
ADD glide.lock glide.lock
# install packages
RUN glide install
# add source code
ADD src src
# run main.go
CMD ["go", "run", "src/main.go"]
```

构建镜像并运行 gin 以便当我们修改了 `src` 中的源代码时可以自动重新编译：

```bash
$ cd go-docker
$ docker build -t go-docker-dev .
$ docker run --rm -it -p 8080:8080 -v $(pwd):/go/src/app \
             go-docker-dev bash
root@id:/go/src/app# gin --path src --port 8080 run main.go
```

注意到 web-server 需要一个 `PORT` 的环境变量来监听，因为 gin 会随机设置 `PORT` 变量并代理到该端口的连接。

现在，修改 `src` 目录下的内容会触发重新编译，所有更新的内容可以实时在 `http://localhost:8080` 访问到。

一旦开发完毕，我们可以构建二进制文件并运行它，而不需要使用 `go run` 命令。可以使用相同的镜像来构建，或者也可以使用 Docker 的多阶段构建，即使用 `golang` 镜像来构建并使用迷你 linux 容器如 `alpine` 来运行服务。

## 单阶段生产构建

```bash
FROM golang:1.8.5-jessie
# install glide
RUN go get github.com/Masterminds/glide
# create a working directory
WORKDIR /go/src/app
# add glide.yaml and glide.lock
ADD glide.yaml glide.yaml
ADD glide.lock glide.lock
# install packages
RUN glide install
# add source code
ADD src src
# build main.go
RUN go build src/main.go
# run the binary
CMD ["./main"]
```

构建并运行该一体化的镜像：

```bash
$ cd go-docker
$ docker build -t go-docker-prod .
$ docker run --rm -it -p 8080:8080 go-docker-prod
```

因为底层使用了 Debian 镜像，该镜像会达到 750 MB 左右的大小（取决于你的源代码）。让我们看看如何缩减体积。

## 多阶段生产构建
多阶段构建允许你在一个完整的 OS 环境中进行构建，但构建后的二进制文件通过一个非常苗条的镜像来运行，该镜像仅比构建后的二进制文件略大一点而已。

```bash
FROM golang:1.8.5-jessie as builder
# install glide
RUN go get github.com/Masterminds/glide
# setup the working directory
WORKDIR /go/src/app
ADD glide.yaml glide.yaml
ADD glide.lock glide.lock
# install dependencies
RUN glide install
# add source code
ADD src src
# build the source
RUN go build src/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main src/main.go

# use a minimal alpine image
FROM alpine:3.7
# add ca-certificates in case you need them
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
# set working directory
WORKDIR /root
# copy the binary from builder
COPY --from=builder /go/src/app/main .
# run the binary
CMD ["./main"]
```

现在二进制文件为 14 MB 左右，docker 镜像为 18 MB 左右。真是多亏了 `alpine`。

想减小二进制文件体积吗？继续看吧。

## 福利：使用 UPX 来压缩二进制文件

在 [Hasura](https://hasura.io/)，我们已经在到处使用 [UPX](https://upx.github.io/) 了，压缩后我们的 CLI 二进制文件从 50 MB 左右降到 8 MB左右，大大加快了下载速度。UPX 可以极快地进行原地解压，不需要额外的工具，因为它将解压器嵌入到了二进制文件内部。

```bash
FROM golang:1.8.5-jessie as builder
# install xz
RUN apt-get update && apt-get install -y \
    xz-utils \
&& rm -rf /var/lib/apt/lists/*
# install UPX
ADD https://github.com/upx/upx/releases/download/v3.94/upx-3.94-amd64_linux.tar.xz /usr/local
RUN xz -d -c /usr/local/upx-3.94-amd64_linux.tar.xz | \
    tar -xOf - upx-3.94-amd64_linux/upx > /bin/upx && \
    chmod a+x /bin/upx
# install glide
RUN go get github.com/Masterminds/glide
# setup the working directory
WORKDIR /go/src/app
ADD glide.yaml glide.yaml
ADD glide.lock glide.lock
# install dependencies
RUN glide install
# add source code
ADD src src
# build the source
RUN go build src/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main src/main.go
# strip and compress the binary
RUN strip --strip-unneeded main
RUN upx main

# use a minimal alpine image
FROM alpine:3.7
# add ca-certificates in case you need them
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
# set working directory
WORKDIR /root
# copy the binary from builder
COPY --from=builder /go/src/app/main .
# run the binary
CMD ["./main"]
```

UPX 压缩后的二进制文件为 3 MB 左右并且 docker 镜像为 6 MB 左右。

**相比最开始的镜像，缩减了 100 倍**

如果你有更好的建议或是你需要其他的使用场景，请在评论区留言或者去 [HackerNews](https://news.ycombinator.com/item?id=16308391) 和 [Reddit](https://www.reddit.com/r/golang/comments/7vexdl/the_ultimate_guide_to_writing_dockerfiles_for_go/) 进行讨论。

## 广告

额...你尝试过在 Hasura 上部署 Go web-app 吗？这真的是世界上最快的将 Go apps 部署到 HTTPS 域下的方法（仅仅 `git push` 就够了）。使用这里的项目模板快速开始吧：https://hasura.io/hub/go-frameworks。Hasura 所有项目模板都配套有 Dockerfile 和 Kubernetes spec，允许你按照你的方式来自定义。

---

via：https://blog.hasura.io/the-ultimate-guide-to-writing-dockerfiles-for-go-web-apps-336efad7012c

作者：[Shahidh K Muhammed](https://github.com/shahidhk)
译者：[ParadeTo](https://github.com/ParadeTo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
