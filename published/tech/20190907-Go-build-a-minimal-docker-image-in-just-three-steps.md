首发于：https://studygolang.com/articles/24875

# 只用 3 步构建 Go docker 最小镜像

![DockerGopher](https://raw.githubusercontent.com/studygolang/gctt-images/master/build-mini-docker-image/DockerGopher.png)

## Go——仅需三个步骤即可构建最小的 Docker 映像

当您为 docker 构建 Go 应用程序时，通常从诸如 `golang:1.13` 之类的映像开始。但将这个映像实际运行时会浪费资源。让我们看一下如何将 Go 应用程序构建为绝对最小的 Docker 映像。

## 1. 选择 Go 版本

尽管使用 `golang:latest` 或者 仅使用 `golang` 的版本镜像很诱人，但由于各种问题，这样做都不太好，其中主要的一个问题是这样做构建（可能）不具有重复性。无论是开发测试需要部署产品时使用的相同版本（的镜像），还是你发现自己需要修补旧版本（镜像）的应用程序，最好将 Go 发行版镜像的版本固定，只有当你知道你需要更新 Go 版本的时候你再去更新它。

因此，需要一直使用完整的说明，包含补丁版本号，而且最好说明镜像的基本系统，比如：`1.13.0-alpine3.10`。

## 2. 保持最小

这个*最小*包含两个方面：

1. 最短构建时间
2. 最小产出镜像

### 快速构建

Docker 为您缓存中间层，因此如果您正确地构造 Dockerfile，您可以减少每次（更改后）后续重建所需的时间。根据经验来说，根据命令的源（例如：`COPY` 源）更改的频率对命令进行排序。

另外，请考虑使用 `.dockerignore` 文件，该文件有助于保持构建上下文较小——当您执行 `docker build` 时，docker 需要将当前目录中的所有内容都提供给构建 docker 守护进程（即在 docker 构建开始时向 docker 守护进程发送构建上下文）。简单来说，如果你的代码仓库包含了很多构建你的程序所不需要的数据（比如测试，markdown 格式文档生成等），`.dockerignore` 将有助于加快构建速度。

至少，您可以从下边的示例内容开始尝试。如果你 `COPY . .`，Dockerfile 就会进入上下文中（不应该这样做），当你只修改 Dockerfile 时，不需要执行并使所有的东西无效。

```shell
.git
Dockerfile
testdata
```

### 最小镜像

最简单的手段是使用 `scratch`（构建基础镜像），没有其他手段能与之相比。Scratch 是特殊的 `base` （基础）镜像，它不是一个真正的镜像，而是一个完全空的系统。注意：在老版本 docker 中，显式的 scratch 镜像作为一个真正的镜像层，`docker 1.5` 之后的版本就不再是这样。

你的 Dockerfile 只需两步：

1. 基于一个镜像（比如 `builder` 镜像，你想叫什么都行，编译你的应用程序；
2. 然后将编译产出的二进制程序和所有其他依赖拷贝到基于 scratch 的镜像中；

这样做贼管用！

## 3. 放在一起

看看完整的 Dockerfile 长啥样：

```dockerfile
FROM golang:1.13.0-stretch AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=1

WORKDIR /build

# 缓存 mod 检索-那些不常更改的模块
COPY go.mod .
COPY go.sum .
RUN Go mod download

# 复制构建应用程序所需的代码
# 可能需要更改下边的命令，只复制您实际需要的内容。
COPY . .

# 构建应用程序
RUN Go build ./cmd/my-awesome-go-program

# 我们创建一个 /dist 目录， 仅包含运行时必须的文件
# 然后，他会被复制到输出镜像的 / （根目录）
WORKDIR /dist
RUN cp /build/my-awesome-go-program ./my-awesome-go-program

# 可选项:如果您的应用程序使用动态链接(通常情况下使用 CGO)，
# 这将收集相关库，以便稍后将它们复制到最终镜像
# 注意: 确保您遵守您复制和分发的库的许可条款
RUN ldd my-awesome-go-program | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname ./%); cp % ./%;'
RUN mkdir -p lib64 && cp /lib64/ld-linux-x86-64.so.2 lib64/

# 在运行时复制或创建您的应用程序需要的其他目录/文件。
# 例如，本例使用 /data 作为工作目录，在正常运行容器时，该目录可能绑定到永久目录
RUN mkdir /data

# 构建最小运行时镜像
FROM scratch

COPY --chown=0:0 --from=builder /dist /

# 设置应用程序以 /data 文件夹中的非 root 用户身份运
# User ID 65534 通常是 'nobody' 用户.
# 映像的执行者仍应在安装过程中指定一个用户。
COPY --chown=65534:0 --from=builder /data /data
USER 65534
WORKDIR /data

ENTRYPOINT ["/my-awesome-go-program"]
```

如果您觉得这些功能有用，或者想要分享一些自己的方法或技巧，请在下边发表评论。

---

via: https://dev.to/ivan/go-build-a-minimal-docker-image-in-just-three-steps-514i

作者：[Ivan Dlugos](https://github.com/vaind)
译者：[TomatoAres](https://github.com/TomatoAres)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
