
# Golang 的微服务 - 第二部分 - Docker and go-micro



## 简介: Docker and go-micro

**[在之前的文章中](https://ewanvalentine.io/microservices-in-golang-part-1/)**，我们基本覆盖了编写一个基于 `gRPC` 的微服务。在这个部分，我们将涵盖 `Docker` 服务的基础知识，我们也将使用 [go-micro](https://github.com/micro/go-micro) 更新我们的服务，最后，引入第二个服务。

## 介绍 Docker
随着云计算的到来，和微服务的诞生。在部署的时候有更多的压力，但是一次一小段代码就产生了一些有趣的新思想和新技术，其中之一就是[容器](https://en.wikipedia.org/wiki/Operating-system-level_virtualization)的概念。

在早些的时候，团队部署一个庞大的服务到静态服务器，运行一套操作系统，需要使用一组预定义的依赖来跟踪。例如，可能是由管理员提供的 `VM` 虚拟机或者`Pupet` 。 伸缩是昂贵的并且不一定有效，最常见的是垂直缩放，例如在静态服务器上投入越来越多资源。

针对虚拟机的配置，伴随着像 [vagrant](https://www.vagrantup.com/) 这样的工具越来越常使用。但是运行一个虚拟机任然是一个相当大的操作。你的应用、内核和所有都运行在一个完整的操作系统、主机中。在资源方面，这是相当昂贵的。所以当微服务出现时，在自己的环境中运行这么多独立的代码库变得不可行。

## 随着容器的到来
[容器](https://en.wikipedia.org/wiki/Operating-system-level_virtualization)减少了操作系统的版本。容器不包含内核、用户操作系统或通常构成操作系统的较低级别组件。

容器只包含顶层库及其运行组件。内核在主机上共享。所以注意运行一个 `Unix` 内核，然后由 `n` 个容器共享，运行非常不同的运行时集合。

在引擎下，容器使用各种内核工具。以便跨容器空间共享资源和网络功能。

[进一步阅读](https://www.redhat.com/en/topics/containers/whats-a-linux-container)

这意味着您可以运行代码所需的运行时和依赖关系，无需启动几个完整的操作系统。这是一个改变游戏的规划，因为一个容器和虚拟机比较体积是比较小的。例如 `Ubuntu`，它通常小于 `1GB` 大小，而 `Docker` 镜像只有 `188M`。

你会注意到我在这个介绍中更广泛的谈到容器，而不是 `Docker` 容器。通常认为 [Docker](https://www.docker.com/) 和容器是一回事。但是，容器在Linux中更多是一个概念或一组功能。
[Docker](https://www.docker.com/)只是其中的一种，只是因为好用而变得流行，还有其他的。因为在我看来，`Docker` 是最好的支持者，对于新手来说也是最简单的。

所以希望你看到容器的价值。我们开始使用 `Docker` 来运行我们的第一个服务。我们来创建一个`DockerFile`。（备注: Docker容器的创建一般都使用DockerFile，容器会根据这个文件创建相对应的运行环境）

```
touch consignment-service/Dockerfile
```

在该文件中添加以下内容:

```
FROM alpine:latest

RUN mkdir /app  
WORKDIR /app  
ADD consignment-service /app/consignment-service

CMD ["./consignment-service"]  
```

如果你在 Linux 上运行，你可能会遇到使用 Alpine 的问题。所以，如果你在 Linux 机器上关注这篇文章，只需用 `debian` 替换 `alpine` 即可，你应该能够正常运行。 稍后我们将会介绍一个更好的方法来构建我们的二进制文件。

首先，我们拉去最新的 [Linux Alpine](https://alpinelinux.org/) 镜像。[Linux Alpine](https://alpinelinux.org/) 是一个轻量级Linux发行版，为运行Dockerised Web应用程序而开发和优化。换一种说法，[Linux Alpine](https://alpinelinux.org/) 具有足够的依赖性和运行时功能来运行大多数应用程序。这意味着它的镜像大小只有 8M 左右。与之相比，大约 1GB 的 Ubuntu 虚拟机，你可以开始看到为什么 Docker 镜像更适合微服务和云计算。

接下来我们创建一个新的目录来存放我们的应用程序，并将上下文目录设置到我们的新目录中。 这是我们的应用程序目录是默认的目录。 然后，我们将编译后的二进制文件添加到我们的Docker容器中，并运行它。

现在我们来更新`MakeFile`文件来构建我们的 Docker镜像。

```
build:  
    ... 
    GOOS=linux GOARCH=amd64 go build
    docker build -t consignment-service .
```


----------------

via: https://ewanvalentine.io/microservices-in-golang-part-2/

作者：[Ewan Valentine](http://ewanvalentine.io/author/ewan)
译者：[译者ID](https://github.com/guoxiaopang)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
