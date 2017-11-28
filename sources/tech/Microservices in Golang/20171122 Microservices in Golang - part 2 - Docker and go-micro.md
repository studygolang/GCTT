
# Golang 的微服务 - 第二部分 - Docker and go-micro



## 简介: Docker and go-micro

**[在之前的文章中](https://ewanvalentine.io/microservices-in-golang-part-1/)**，我们基本覆盖了编写一个基于 `gRPC` 的微服务。在这个部分，我们将涵盖 `Docker` 服务的基础知识，我们也将使用 [go-micro](https://github.com/micro/go-micro) 更新我们的服务，最后，引入第二个服务。

## 介绍 Docker
随着云计算的到来，和微服务的诞生。在部署的时候有更多的压力，但是一次一小段代码就产生了一些有趣的新思想和新技术，其中之一就是[容器](https://en.wikipedia.org/wiki/Operating-system-level_virtualization)的概念。

在早些的时候，团队部署一个庞大的服务到静态服务器，运行一套操作系统，需要使用一组预定义的依赖来跟踪。例如，可能是由管理员提供的 `VM` 虚拟机或者`Pupet` 。 伸缩是昂贵的并且不一定有效，最常见的是垂直缩放，例如在静态服务器上投入越来越多资源。

针对虚拟机的配置，伴随着像 [vagrant](https://www.vagrantup.com/) 这样的工具越来越常使用。但是运行一个虚拟机任然是一个相当大的操作。你的应用、内核和所有都运行在一个完整的操作系统、主机中。在资源方面，这是相当昂贵的。所以当微服务出现时，在自己的环境中运行这么多独立的代码库变得不可行。

## 随着容器的到来

----------------

via: https://ewanvalentine.io/microservices-in-golang-part-2/

作者：[Ewan Valentine](http://ewanvalentine.io/author/ewan)
译者：[译者ID](https://github.com/guoxiaopang)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
