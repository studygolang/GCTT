
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

我们在这里增加了两个步骤，我想详细解释一下。首先，我们正在构建我们的二进制文件。你会注意到在运行命令 `$ go build` 之前，正在设置两个环境变量。GOOS 和 GOARCH 允许您为另一个操作系统交叉编译您的二进制文件，由于我在 Macbook上开发，所以无法编译二进制文件，然后在 Docker 容器中运行它，该容器使用 Linux。二进制在你的 Docker 容器中将是完全没有意义的，它会抛出一个错误。第二步是添加 Docker 构建过程。这将读取你的 Dockerfile文件，并通过一个名称 `consignment-service` 构建镜像。句号表示一个目录路径，所以在这里我们只是希望构建过程在当前目录中查找。

我将在我们的Makefile中添加一个新条目：

```
run:  
    docker run -p 50051:50051 consignment-service
```

在这里，我们运行 `consignment-service` Docker镜像，并暴露 50051 端口。由于Docker在单独的网络层上运行，因此您需要将Docker容器中使用的端口转发给主机。您可以通过更改第一个段将内部端口转发到主机上的新端口。例如，如果要在端口 8080 上运行此服务，则需要将-p参数更改为 `8080：50051`。您也可以通过包含 `-d` 标志在后台运行容器。例如，`docker run -d -p 50051:50051 consignment-service`。

[您可以阅读更多关于Docker网络如何工作的信息](https://docs.docker.com/engine/userguide/networking/)

当您运行 `$ docker build` 时，您正在将代码和运行时环境构建到镜像中。Docker镜像是您的环境及其依赖关系的可移植快照。你可以将它分享到 Docker Hub 来共享你的 Docker 镜像。
Docker 镜像就像一个 npm 或 yum repo。
当你在你的 Dockerfile里面定义了`FROM`，你就告诉了 docker 使用 docker hub 来构建你的镜像。然后，您可以扩展并覆盖该基本文件的某些部分，方法是自行重新定义它们。我们不会公开我们的 Docker 镜像，但是可以随时仔细阅读 Docker hub，并且注意到有多少功能被容器化。一些非常[显着的事情](https://www.youtube.com/watch?v=GsLZz8cZCzc)已经被 Docker 化了。

Dockerfile 中的每个声明在第一次构建时都被缓存。 这样可以节省每次更改时重新构建整个运行时间的情况。 Docker非常聪明，可以确定哪些部分发生了变化，哪些部分需要重新构建。 这使得构建过程非常快速。

有足够的容器！ 让我们回到我们的代码。

在创建 gRPC 服务时，创建连接的代码有很多，并且必须将服务地址的位置硬编码到客户端或其他服务中，以便连接到它。 这很棘手，因为当您在云中运行服务时，它们可能不共享相同的主机，或者重新部署服务后地址或IP可能会更改。

这是服务发现的起点。 服务发现保持所有服务及其位置的最新目录。 每个服务在运行时注册自己，并在关闭时自行注销。 每个服务都有一个名字或编号分配给它。 因此，即使可能有新的IP地址或主机地址，只要服务名称保持不变，您就不需要从其他服务更新对此服务的调用。

通常情况下，解决这个问题的方法有很多，但是像编程中的大多数情况一样，如果已经有人解决了这个问题，那么重新发明轮子就没有意义了。 [Go-micro](https://github.com/micro/go-micro) 的创始人是 @chuhnk（Asim Aslam），他以一种非常清晰和易用的方式解决了这些问题。

----------------

via: https://ewanvalentine.io/microservices-in-golang-part-2/

作者：[Ewan Valentine](http://ewanvalentine.io/author/ewan)
译者：[译者ID](https://github.com/guoxiaopang)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
