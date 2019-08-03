首发于：https://studygolang.com/articles/22434

# 使用 Go 和 ReactJS 构建聊天系统（六）：Docker 部署

本节完整代码：[GitHub](https://github.com/watermelo/realtime-chat-go-react/tree/part-6)

> 本文是关于使用 ReactJS 和 Go 构建聊天应用程序的系列文章的第 6 部分。你可以在这里找到第 5 部分 - [优化前端](https://studygolang.com/articles/22433)

在本节中，我们将专注于将 Docker 添加到后端应用程序中。

为什么要这么做呢？在我们研究诸如身份验证，负载均衡和部署之类的问题前，使用容器技术部署应用程序是个标准的做法。

## 为什么用 Docker

如果这是你第一次听说 Docker 容器化技术，那么你可能会质疑使用它的原因。

对我来说，其中一个主要原因是它让部署变得更加容易。你可以将基于 docker 的应用程序部署到支持 Docker 的任何服务器或平台。

这意味着，无论你在何处部署，都可以使用简单的命令启动应用程序。

不仅如此，它还解决了 “在我的机器上运行好好的” 这个问题，因为在你的 `Dockerfile` 中，可以指定应用程序启动时所需的确定环境。

## 开始

首先我们得在计算机上安装 Docker。可以参考：[Docker 指南](https://www.docker.com/get-started)

在安装了 docker 并让它运行后，我们就可以创建 `Dockerfile` 了：

```shell
FROM golang:1.11.1-alpine3.8
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go mod download
RUN go build -o main ./...
CMD ["/app/main"]
```

我们定义了 Dockerfile 文件之后，就可以使用 `docker` cli 构建 Docker 镜像：

> 注意 - 如果你的网速比较差，下一个命令可能需要等待一段时间才能执行，但是，由于有缓存后续命令会快得多。

```shell
$ docker build -t backend .
Sending build context to Docker daemon  11.26kB
Step 1/8 : FROM golang:1.11.1-alpine3.8
 ---> 95ec94706ff6
Step 2/8 : RUN apk add bash ca-certificates git gcc g++ libc-dev
 ---> Running in 763630b369ca
 ...
```

成功完成 `build` 步骤后，我们可以将该容器启动起来：

```shell
$ docker run -it -p 8080:8080 backend
Distributed Chat App v0.01
WebSocket Endpoint Hit
Size of Connection Pool:  1
&{ 0xc000124000 0xc0000902a0 {0 0}}
Message Received: {Type:1 Body:test}
Sending message to all clients in Pool
```

正如你所见，在运行此命令并刷新客户端后，可以看到现在已经连接到 Docker 化的应用服务，也可以看到终端正在打印日志。

如果现在想要将此应用程序部署到 AWS 上，这会大大简化一些过程。现在可以利用 AWS 的 ECS 服务的一些命令来部署和运行我们的容器。

同样的，如果想要使用 Google 云，我们可以将其部署到 Google 的容器产品中，无需额外的工作！这只是突出 Docker 化的巨大好处之一。

## 前端为什么不使用 Docker

在这一点上，你可能想知道为什么不对 `frontend/` 应用程序做同样的事情？原因是我们打算将前端应用部署到 AWS S3 服务。

当部署上线时，前端不需要任何花哨的服务，我们只需要能够可靠地提供构建的前端文件。

## 总结

因此，在本节中，我们设法将 Docker 添加到后端应用程序中，这对持续开发和部署的人员有益。

---

via: https://tutorialedge.net/projects/chat-system-in-go-and-react/part-6-dockerizing-your-backend/

作者：[Elliot Forbes](https://twitter.com/elliot_f)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
