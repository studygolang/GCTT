首发于：https://studygolang.com/articles/23458

# Docker 参考架构：设计可扩展、可移植的 Docker 容器网络

## 您将学到什么

Docker 容器将软件封装到一个完整的文件系统当中，这个文件系统包括软件运行所需的一切：代码、运行时、系统工具、系统库，所有能安装在服务器上的东西。这确保了软件在不同的环境下都能有相同的运行情况。默认情况下， 容器将各个应用彼此隔离开并将软件与底层基础设施隔离开，同时为应用提供额外的一层保护。

那如果应用需要和其他应用、主机或外部网络相互通信呢？您会如何在确保应用可移植性、服务发现、负载均衡、安全性、高性能和可扩展性的同时，设计一个拥有适当连通性的网络呢？本文档将解决这些网络设计的问题，提供可用的工具以及通用的部署模式。文章不会指定或推荐物理网络的设计，但会给出一些兼顾应用的需求与物理网络条件约束的 Docker 网络设计方法。

### 前置内容

在继续阅读本文之前，推荐先熟悉 Docker 的基本概念以及 Docker Swarm：

- [Docker 简介](https://docs.docker.com/engine/understanding-docker/)
- [Docker Swarm](https://docs.docker.com/engine/swarm/) 和 [Swarm 模式简介](https://docs.docker.com/engine/swarm/key-concepts/#/services-and-tasks)

## 为容器和微服务组网的挑战

微服务的引入扩大了应用的规模，这使得连通性和隔离性在应用中更为重要。Docker 网络的设计哲学是“应用驱动”，旨在为网络运营者提供更多选择和灵活性，同时也为应用开发者提供更高层次的抽象。

正如其他所有设计一样，网络设计也是个权衡的过程。**Docker EE** 和 Docker 生态向网络工程师提供了多个工具，以实现应用和环境之间的权衡。不同的工具选择有不同的益处和取舍。本篇指南的余下部分将逐一详细分析这些选择，力求网络工程师们可以明白哪种最适于他们的环境。

Docker 已经开发出了一种新的应用程序交付模式，通过这种模式，容器也改变了网络接入的某些方面。以下几个话题对于容器化的应用而言，都是常见的设计问题：

- **可移植性**
  - *我怎样才能在发挥特定网络独特优势的前提下，最大程度地保证跨不同网络环境的可移植性？*
- **服务发现**
  - *我要怎样在服务扩容、缩容之后获悉它们在哪里运行？*
- **安全**
  - *我要如何进行隔离，以避免出问题的容器相互访问？*
  - *我要如何保障拥有应用流量和集群控制流量的容器是安全的？*
- **性能**
  - 我要怎样才能提供高级的网络服务，使延迟最小，使带宽最大？
- **可扩展性**
  - *我要怎么保证这些特性不会在应用进行跨主机扩容的时候受到影响？*

## 容器网络模型

Docker 的网络架构是建立在一系列称为*容器网络模型*（Container Networking Model, CNM）的接口之上的。CNM 的设计哲学是为了提供跨多种基础设施的应用可移植性。这一模型在应用可移植性和充分利用基础设施自有特性、能力之间，取得了一个平衡。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/cnm.png)

### CNM 部件

在 CNM 之中，有几个高层次的部件。它们全部都是操作系统和基础硬件不可感的，因此应用可以在任何基础设施栈中拥有一致的表现。

- **沙箱** —— 一个沙箱包含容器的网络栈配置。这包括容器接口的管理、路由表和 DNS 设置。沙箱的实现可以是 Linux Network Namespace，FreeBSD Jail 或是其他类似的技术。一个沙箱可以包含来自不同网络的多个端点。
- **端点** —— 端点负责将沙箱与网络相连。端点部件的存在使得实际的网络连接可以从应用中抽象出来。这有助于维持可移植性，使服务可以采用不同的网络驱动，而无需顾虑如何与网络相连。
- **网络** —— CNM 并不是用 OSI 模型中的概念来诠释“网络”。网络部件的实现可以通过 Linux bridge，VLAN 等等。网络就是一个相互连通的若干端点的集合。与网络不连通的端点不具有网络连通性。

### CNM 驱动接口

容器网络模型 CNM 提供了两个可插拔的开放接口，供用户、社区和供应商使用，以更好地利用网络中的其他功能、可见性或可控性。

存在以下两种网络驱动接口：

- **网络驱动** —— Docker 网络驱动提供使网络运行的实际实现。它们是可插拔的，因此可以使用不同的驱动程序并轻松互换以支持不同的用例。可以在给定的 Docker Engine 或群集上同时使用多个网络驱动程序，但每个 Docker 网络仅通过单个网络驱动程序进行实例化。有两种类型的 CNM 网络驱动程序：
  - **原生网络驱动** —— 原生网络驱动程序是 Docker Engine 的原生部分，由 Docker 提供。有多种驱动程序可供选择，支持不同的功能，如覆盖网络或本地网桥。
  - **远程网络驱动** —— 远程网络驱动是社区和其他供应商创建的网络驱动程序。这些驱动程序可用于和现有软硬件相集成。用户还可以在需要用到现有网络驱动程序不支持的特定功能的情况下创建自己的驱动程序。
- **IPAM 驱动** —— Docker 具有本机 IP 地址管理驱动程序，若未另加指定，将为网络和端点提供默认子网或 IP 地址。IP 地址也可以通过网络、容器和服务创建命令手动分配。我们同样拥有远程 IPAM 驱动程序，可与现有 IPAM 工具集成。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/cnm-api.png)

### Docker 原生网络驱动

Docker 原生网络驱动程序是 Docker Engine 的一部分，不需要任何额外的模块。它们通过标准 Docker 网络命令调用和使用。共有以下几种原生网络驱动程序。

| 驱动        | 描述                                                         |
| :---------- | :----------------------------------------------------------- |
| **Host**    | 使用 `host` 驱动意味着容器将使用主机的网络栈。没有命名空间分离，主机上的所有接口都可以由容器直接使用。 |
| **Bridge**  | `bridge` 驱动会在 Docker 管理的主机上创建一个 Linux 网桥。默认情况下，网桥上的容器可以相互通信。也可以通过 `bridge` 驱动程序配置，实现对外部容器的访问。 |
| **Overlay** | `overlay` 驱动创建一个支持多主机网络的覆盖网络。它综合使用本地 Linux 网桥和 VXLAN，通过物理网络基础架构覆盖容器到容器的通信。 |
| **MACVLAN** | `macvlan` 驱动使用 MACVLAN 桥接模式在容器接口和父主机接口（或子接口）之间建立连接。它可用于为在物理网络上路由的容器提供 IP 地址。此外，可以将 VLAN 中继到 `macvlan` 驱动程序以强制执行第 2 层容器隔离。 |
| **None**    | `none` 驱动程序为容器提供了自己的网络栈和网络命名空间，但不配置容器内的接口。如果没有其他配置，容器将与主机网络栈完全隔离。 |

### 网络范围

如 `docker network ls` 命令结果所示，Docker 网络驱动程序具有 *范围* 的概念。网络范围是驱动程序的作用域，可以是本地范围或 Swarm 集群范围。本地范围驱动程序在主机范围内提供连接和网络服务（如 DNS 或 IPAM）。Swarm 范围驱动程序提供跨群集的连接和网络服务。集群范围网络在整个群集中具有相同的网络 ID，而本地范围网络在每个主机上具有唯一的网络 ID。

```bash
$ docker network ls
NETWORK ID          NAME                DRIVER              SCOPE
1475f03fbecb        bridge              bridge              local
e2d8a4bd86cb        docker_gwbridge     bridge              local
407c477060e7        host                host                local
f4zr3zrswlyg        ingress             overlay             swarm
c97909a4b198        none                null                local
```

### Docker 远程网络驱动

以下社区和供应商创建的远程网络驱动程序与 CNM 兼容，每个都为容器提供独特的功能和网络服务。

| 驱动                                                         | 描述                                                         |
| :----------------------------------------------------------- | :----------------------------------------------------------- |
| [**contiv**](http://contiv.github.io/)                       | 由 Cisco Systems 领导的开源网络插件，为多租户微服务部署提供基础架构和安全策略。Contiv 还为非容器工作负载和物理网络（如 ACI）提供兼容集成。Contiv 实现了远程网络和 IPAM 驱动。 |
| [**weave**](https://www.weave.works/docs/net/latest/introducing-weave/) | 作为网络插件，weave 用于创建跨多个主机或多个云连接 Docker 容器的虚拟网络。Weave 提供应用程序的自动发现功能，可以在部分连接的网络上运行，不需要外部群集存储，并且操作友好。 |
| [**calico**](https://www.projectcalico.org/)                 | 云数据中心虚拟网络的开源解决方案。它面向数据中心，大多数工作负载（虚拟机，容器或裸机服务器）只需要 IP 连接。Calico 使用标准 IP 路由提供此连接。工作负载之间的隔离都是通过托管源和目标工作负载的服务器上的 iptables 实现的，无论是根据租户所有权还是任何更细粒度的策略。 |
| [**kuryr**](https://github.com/openstack/kuryr)              | 作为 OpenStack Kuryr 项目的一部分开发的网络插件。它通过利用 OpenStack 网络服务 Neutron 实现 Docker 网络（libnetwork）远程驱动程序 API。Kuryr 还包括一个 IPAM 驱动程序。 |

### Docker 远程 IPAM 驱动

社区和供应商创建的 IPAM 驱动程序还可用于提供与现有系统或特殊功能的集成。

| 驱动                                                         | 描述                                                 |
| :----------------------------------------------------------- | :--------------------------------------------------- |
| [**infoblox**](https://hub.docker.com/r/infoblox/ipam-driver/) | 一个开源 IPAM 插件，提供与现有 Infoblox 工具的集成。 |

> Docker 拥有许多相关插件，并且越来越多的插件正被设计、发布。Docker 维护着[最常用插件列表](https://docs.docker.com/engine/extend/legacy_plugins/)。

## Linux 网络基础

Linux 内核具有非常成熟和高性能的 TCP/IP 网络栈实现（不仅是 DNS 和 VXLAN 等其他原生内核功能）。Docker 网络使用内核的网络栈作为低级原语来创建更高级别的网络驱动程序。简而言之，*Docker 网络* **就是** *Linux 网络*。

现有 Linux 内核功能的这种实现确保了高性能和健壮性。最重要的是，它提供了跨许多发行版和版本的可移植性，从而增强了应用程序的可移植性。

Docker 使用了几个 Linux 网络基础模块来实现其原生 CNM 网络驱动程序，包括 **Linux 网桥**，**网络命名空间**，**veth** 和 **iptables**。这些工具的组合（作为网络驱动程序实现）为复杂的网络策略提供转发规则，网络分段和管理工具。

### Linux 网桥

**Linux 网桥**是第 2 层设备，它是 Linux 内核中物理交换机的虚拟实现。它通过检视流量动态学习 MAC 地址，并据此转发流量。Linux 网桥广泛用于许多 Docker 网络驱动程序中。Linux 网桥不应与 Docker 网络驱动程序 bridge 混淆，后者是 Linux 网桥的更高级别实现。

### 网络命名空间

**Linux 网络命名空间**是内核中隔离的网络栈，具有自己的接口，路由和防火墙规则。它负责容器和 Linux 的安全方面，用于隔离容器。在网络术语中，它们类似于 VRF，它将主机内的网络控制和数据隔离。网络命名空间确保同一主机上的两个容器无法相互通信，甚至无法与主机本身通信，除非通过 Docker 网络进行配置。通常，CNM 网络驱动程序为每个容器实现单独的命名空间。但是，容器可以共享相同的网络命名空间，甚至可以是主机网络命名空间的一部分。主机网络命名空间容纳主机接口和主机路由表。此网络命名空间称为全局网络命名空间。

### 虚拟以太网设备

**虚拟以太网设备**或简称 **veth** 是 Linux 网络接口，充当两个网络命名空间之间的连接线。veth 是一个全双工链接，每个命名空间中都有一个接口。一个接口中的流量被引导出另一个接口。Docker 网络驱动程序利用 veth 在创建 Docker 网络时提供名称空间之间的显式连接。当容器连接到 Docker 网络时，veth 的一端放在容器内（通常被视为 ethX 接口），而另一端连接到 Docker 网络。

### iptables

**iptables** 是原生包过滤系统，自 2.4 版本以来一直是 Linux 内核的一部分。它是一个功能丰富的 L3/L4 防火墙，为数据包的标记，伪装和丢弃提供规则链。本机 Docker 网络驱动程序广泛使用 iptables 来隔离网络流量，提供主机端口映射，并标记流量以实现负载平衡决策。

## Docker 网络控制面板

除了传播控制面板数据之外，Docker 分布式网络控制面板还管理 Swarm 集群的 Docker 网络状态。它是 Docker Swarm 集群的内置功能，不需要任何额外的组件，如外部 KV 存储。控制平面使用基于 [SWIM](https://www.cs.cornell.edu/~asdas/research/dsn02-swim.pdf) 的 [Gossip](https://en.wikipedia.org/wiki/Gossip_protocol) 协议在 Docker 容器集群中传播网络状态信息和拓扑。Gossip 协议非常有效地实现了集群内的最终一致性，同时保持了非常大规模集群中消息大小，故障检测时间和收敛时间的恒定速率。这可确保网络能够跨多个节点进行扩展，而不会引入缩放问题，例如收敛缓慢或误报节点故障。

控制面板非常安全，通过加密通道提供机密性、完整性和身份验证。它也是每个网络的边界，大大减少了主机收到的更新。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/controlplane.png)

它由多个组件组成，这些组件协同工作以实现跨大规模网络的快速收敛。控制平面的分布式特性可确保群集控制器故障不会影响网络性能。

Docker 网络控制面板组件如下：

- **消息传播**以对等方式更新节点，将每次交换中的信息传递到更大的节点组。固定的对等组间隔和大小确保即使群集的大小扩展，网络的情况也是不变的。跨对等体的指数信息传播确保了收敛速度快，并且能满足任何簇大小。
- **故障检测**利用直接和间接的问候消息来排除网络拥塞和特定路径导致误报节点故障。
- 定期实施**完全状态同步**以更快地实现一致性并解析网络分区。
- **拓扑感知**算法探明自身与其他对等体之间的相对延迟。这用于优化对等组，使收敛更快，更高效。
- **控制面板加密**可以防止中间人攻击和其他可能危及网络安全的攻击。

> Docker 网络控制面板是 [Swarm](https://docs.docker.com/engine/swarm/) 的一个组件，需要一个 Swarm 集群才能运行。

## Docker Host 网络驱动

`host` 网络驱动对于 Docker 的新用户来说是最熟悉的，因为它与 Linux 没有 Docker 的情况下所使用的网络配置相同。`--net=host` 有效地关闭了 Docker 网络，容器使用主机操作系统的 host（或默认）网络栈。

通常在使用其他网络驱动程序时，每个容器都被放在其自己的 *网络命名空间* （或沙箱）中，以实现彼此间完全的网络隔离。使用 `host` 驱动程序的容器都在同一主机网络命名空间中，并使用主机的网络接口和 IP 堆栈。主机网络中的所有容器都能够在主机接口上相互通信。从网络角度来看，它们相当于在没有使用容器技术的主机上运行的多个进程。因为它们使用相同的主机接口，所以任意两个容器都不能够绑定到同一个 TCP 端口。如果在同一主机上安排多个容器，可能会导致端口争用。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/host-driver.png)

```bash
#Create containers on the host network
$ docker run -itd --net host --name C1 alpine sh
$ docker run -itd --net host --name nginx

#Show host eth0
$ ip add | grep eth0
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 9001 qdisc mq state UP group default qlen 1000
    inet 172.31.21.213/20 brd 172.31.31.255 scope global eth0

#Show eth0 from C1
$ docker run -it --net host --name C1 alpine ip add | grep eth0
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 9001 qdisc mq state UP qlen 1000
    inet 172.31.21.213/20 brd 172.31.31.255 scope global eth0

#Contact the nginx container through localhost on C1
$ curl localhost
!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
...
```

在此示例中，当容器使用 `host` 网络时，主机，`C1` 和 `nginx` 都共享相同的 `eth0` 接口。这使得 `host` 网络不适合多租户或对安全性要求高的应用程序。`host` 模式的容器可以访问主机上的其他任一容器。这种情况下，可以使用 `localhost` 在容器之间进行通信，如示例中所示，从 `C1` 执行 `curl nginx` 可成功访问。

使用 `host` 驱动程序，Docker 不管理容器网络栈的任何部分，例如端口映射或路由规则。这意味着像 `-p` 和 `--icc` 这样的常见网络标志对 `host` 驱动程序没有任何意义，它们被忽略了。这确实使 `host` 网络成为最简单和最低延迟的网络驱动程序。流量路径直接从容器进程流向主机接口，提供相当于非容器化进程的裸机性能。

完全的主机访问权限和无自动策略管理可能使 `host` 驱动程序难以作为通用网络驱动程序。但是， `host` 确实有一些有趣的性质，可能适用于超高性能应用程序或应用程序故障排除等场景。

## Docker Bridge 网络驱动

本节介绍默认的 Docker bridge 网络以及用户自定义的 bridge 网络。

### 默认 Docker bridge 网络

在任何运行 Docker Engine 的主机上，默认情况下都有一个名为 bridge 的本地 Docker 网络。此网络使用桥接网络驱动程序创建，该驱动程序实例化名为 docker0 的 Linux 网桥。这听起来可能令人困惑。

- `bridge` 是 Docker 网络的名字
- `bridge` 是网络驱动或网络创建的模板
- `bridge` 是 Linux 网桥的名字，是内核用于实现网络功能的基础模块

在独立的 Docker 主机上，如果未指定其他网络，则 `bridge` 是容器连接的默认网络。在以下示例中，创建了一个没有网络参数的容器。Docker Engine 默认将其连接到 `bridge` 网络。在容器内部，注意由 `bridge` 驱动程序创建的 eth0，并由 Docker 本机 IPAM 驱动程序给出一个地址。

```bash
#Create a busybox container named "c1" and show its IP addresses
host $ docker run -it --name c1 busybox sh
c1 # ip address
4: eth0@if5: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue
    link/ether 02:42:ac:11:00:02 brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.2/16 scope global eth0
...
```

容器接口的 MAC 地址是动态生成的，并嵌入 IP 地址以避免冲突。这里 `ac:11:00:02` 对应于 `172.17.0.2`。

主机上的工具 `brctl` 显示主机网络命名空间中存在的 Linux 网桥。它显示了一个名为 `docker0` 的网桥。`docker0` 有一个接口 `vetha3788c4`，它提供从网桥到容器 `c1` 内的 `eth0` 接口的连接。

```bash
host $ brctl show
bridge name      bridge id            STP enabled    interfaces
docker0          8000.0242504b5200    no             vethb64e8b8
```

在容器 `c1` 内部，容器路由表将流量引导到容器的 `eth0`，从而传输到 `docker0` 网桥。

```bash
c1# ip route
default via 172.17.0.1 dev eth0
172.17.0.0/16 dev eth0  src 172.17.0.2
```

容器可以具有零到多个接口，具体取决于它连接的网络数量。一个 Docker 网络只能为网络中的每个容器提供一个接口。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/bridge-driver.png)

如主机路由表中所示，全局网络命名空间中的 IP 接口现在包括 `docker0`。主机路由表提供了外部网络上 `docker0` 和 `eth0` 之间的连接，完成了从容器内部到外部网络的路径。

```bash
host $ ip route
default via 172.31.16.1 dev eth0
172.17.0.0/16 dev docker0  proto kernel  scope link  src 172.17.42.1
172.31.16.0/20 dev eth0  proto kernel  scope link  src 172.31.16.102
```

默认情况下，`bridge` 将从以下范围分配一个子网，172.[17-31].0.0/16 或 192.168.[0-240].0/20，它与任何现有主机接口不重叠。默认的 `bridge` 网络也可以配置为用户提供的地址范围。此外，现有的 Linux 网桥可直接用于 `bridge` 网络，而不需要 Docker 另外创建一个。有关自定义网桥的更多信息，请转至 [Docker Engine 文档](https://docs.docker.com/engine/userguide/networking/default_network/custom-docker0/)。

> 默认 `bridge`  网络是唯一支持遗留[链路](https://docs.docker.com/engine/userguide/networking/default_network/dockerlinks/)的网络。默认 `bridge` 网络**不支持**基于名称的服务发现和用户提供的 IP 地址。

### 用户自定义 bridge 网络

除了默认网络，用户还可以创建自己的网络，称为**用户自定义网络**，可以是任何网络驱动类型。用户定义的 `bridge` 网络，相当于在主机上设置新的 Linux 网桥。与默认 `bridge` 网络不同，用户定义的网络支持手动 IP 地址和子网分配。如果未给出赋值，则 Docker 的默认 IPAM 驱动程序将分配私有 IP 空间中可用的下一个子网。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/bridge2.png)

接下来，在用户定义的 `bridge` 网络下面创建了两个连接到它的容器。指定了子网，网络名为 `my_bridge`。一个容器未获得 IP 参数，因此 IPAM 驱动程序会为其分配子网中的下一个可用 IP， 另一个容器已指定 IP。

```bash
$ docker network create -d bridge --subnet 10.0.0.0/24 my_bridge
$ docker run -itd --name c2 --net my_bridge busybox sh
$ docker run -itd --name c3 --net my_bridge --ip 10.0.0.254 busybox sh
```

`brctl` 现在显示主机上的第二个 Linux 网桥。这一 Linux 网桥的名称 `br-4bcc22f5e5b9` 与 `my_bridge` 网络的网络 ID 匹配。`my_bridge` 还有两个连接到容器 `c2` 和 `c3` 的 veth 接口。

```bash
$ brctl show
bridge name      bridge id            STP enabled    interfaces
br-b5db4578d8c9  8000.02428d936bb1    no             vethc9b3282
                                                     vethf3ba8b5
docker0          8000.0242504b5200    no             vethb64e8b8

$ docker network ls
NETWORK ID          NAME                DRIVER              SCOPE
b5db4578d8c9        my_bridge           bridge              local
e1cac9da3116        bridge              bridge              local
...
```

列出全局网络命名空间接口，可以看到已由 Docker Engine 实例化的 Linux 网络。每个 `veth` 和 Linux 网桥接口都显示为其中一个 Linux 网桥和容器网络命名空间之间的链接。

```bash
$ ip link

1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 9001
3: docker0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500
5: vethb64e8b8@if4: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500
6: br-b5db4578d8c9: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500
8: vethc9b3282@if7: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500
10: vethf3ba8b5@if9: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500
...
```

### 独立容器的外部访问

默认情况下，同一个 Docker 网络上的所有容器（多主机集群范围或本地范围）在所有端口上都可以相互连接。不同 Docker 网络之间的通信以及源自 Docker 外部的容器入站流量都会经过防火墙。这是出于基本的安全考虑，可以保护容器应用程序免受外部世界的影响。这在[网络安全](https://success.docker.com/api/asset/.%2Frefarch%2Fnetworking%2F#security)中有更详细的概述。

对于大多数类型的 Docker 网络（包括 `bridge` 和 `overlay`），必须明确授予应用程序的外部入站访问权限，这是通过内部端口映射完成的。Docker 将主机接口上公开的端口发布到内部容器接口。下图描绘了到容器 `C2` 的入站（底部箭头）和出站（顶部箭头）流量。默认情况下允许出站（*egress*）容器流量。由容器启动的出口连接被伪装或 SNAT 到临时端口（*通常在 32768 到 60999 的范围内*）。返回流量也可经由此链接，因此容器在临时端口上使用主机的最佳可路由 IP 地址。

Ingress 访问是通过显式端口发布提供的。端口发布由 Docker Engine 完成，可以通过 UCP 或 Engine CLI 进行控制。可以将特定或随机选择的端口配置为公开服务或容器。可以将端口设置为侦听特定（或所有）主机接口，并将所有流量从此端口映射到容器内的端口和接口。

```bash
$ docker run -d --name C2 --net my_bridge -p 5000:80 nginx
```

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/nat.png)

使用 Docker CLI 或 UCP 中的 `--publish` / `-p` 配置外部访问。运行上述命令后，该图显示容器 `C2` 已连接到 `my_bridge` 网络，其 IP 地址为 `10.0.0.2`。容器在主机接口 `192.168.0.2` 的端口 `5000` 上向外界发布其服务。进入该端口 `192.168.0.2:5000` 的所有流量都会转发到容器端口 `10.0.0.2:80`。

有关在 Docker Engine 群集中暴露容器和服务的信息，请阅读 [Swarm 服务的外部访问](https://success.docker.com/api/asset/.%2Frefarch%2Fnetworking%2F#swarm-external) 相关文章。

## Overlay 驱动网络架构

原生 Docker `overlay` 网络驱动程序从根本上简化了多主机网络中的许多问题。使用 `overlay` 驱动程序，多主机网络是 Docker 中的一等公民，无需外部配置或组件。 `overlay` 使用 Swarm 分布式控制面板，在非常大规模的集群中提供集中化管理、稳定性和安全性。

### VXLAN 数据平面

`overlay` 驱动程序使用行业标准的 VXLAN 数据平面，将容器网络与底层物理网络（*underlay*）分离。Docker overlay 网络将容器流量封装在 VXLAN 标头中，允许流量穿过第 2 层或第 3 层物理网络。无论底层物理拓扑结构如何，overlay 使网络分段灵活且易于控制。使用标准 IETF VXLAN 标头有助于标准工具检查和分析网络流量。

> 自 3.7 版本以来，VXLAN 一直是 Linux 内核的一部分，而 Docker 使用内核的原生 VXLAN 功能来创建覆盖网络。Docker overlay 数据链路完全在内核空间中。这样可以减少上下文切换，减少 CPU 开销，并在应用程序和物理 NIC 之间实现低延迟、直接的数据通路。

IETF VXLAN（[RFC 7348](https://datatracker.ietf.org/doc/rfc7348/)）是一种数据层封装格式，它通过第 3 层网络覆盖第 2 层网段。VXLAN 旨在用于标准 IP 网络，支持共享物理网络基础架构上的大规模多租户设计。现有的内部部署和基于云的网络可以无感知地支持 VXLAN。

VXLAN 定义为 MAC-in-UDP 封装，将容器第 2 层的帧数据放置在底层 IP/UDP 头中。底层 IP/UDP 报头提供底层网络上主机之间的传输。overlay 是无状态 VXLAN 隧道，其作为参与给定 overlay 网络的每个主机之间的点对多点连接而存在。由于覆盖层独立于底层拓扑，因此应用程序变得更具可移植性。因此，无论是在本地，在开发人员桌面上还是在公共云中，都可以与应用程序一起传输网络策略和连接。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/packetwalk.png)

在此图中，展示了 overlay 网络上的数据包流。以下是 `c1` 在其共享 overlay 网络上发送 `c2` 数据包时发生的步骤：

- `c1` 对 `c2` 进行 DNS 查找。由于两个容器位于同一个 overlay 网络上，因此 Docker Engine 本地 DNS 服务器将 `c2` 解析为其 overlay IP 地址 `10.0.0.3`。
- overlay 网络属于 L2 层，因此 `c1` 生成以 `c2` 的 MAC 地址为目的地的 L2 帧。
- 该帧由 overlay 网络驱动程序用 VXLAN 头封装。分布式 overlay 控制面板管理每个 VXLAN 隧道端点的位置和状态，因此它知道 `c2` 驻留在物理地址 `192.168.0.3` 的 `host-B` 上。该地址成为底层 IP 头的目标地址。
- 封装后，数据包将被发送。物理网络负责将 VXLAN 数据包路由或桥接到正确的主机。
- 数据包到达 host-B 的 eth0 接口，并由 overlay 网络驱动程序解封装。来自 `c1` 的原始 L2 帧被传递到 `c2` 的 eth0 接口，进而传到侦听应用程序。

### Overlay 驱动内部架构

Docker Swarm 控制面板可自动完成 overlay 网络的所有配置，不需要 VXLAN 配置或 Linux 网络配置。数据平面加密是 overlay 的可选功能，也可以在创建网络时由 overlay 驱动程序自动配置。用户或网络运营商只需定义网络（`docker network create -d overlay ...`）并将容器附加到该网络。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/overlayarch.png)

在 overlay 网络创建期间，Docker Engine 会在每台主机上创建 overlay 所需的网络基础架构。每个 overlay 创建一个 Linux 网桥及其关联的 VXLAN 接口。仅当在主机上安排连接到该网络的容器时，Docker Engine 才会智能地在主机上实例化覆盖网络。这可以防止不存在连接容器的 overlay 网络蔓延。

以下示例创建一个 overlay 网络，并将容器附加到该网络。Docker Swarm/UCP 自动创建 overlay 网络。*以下示例需要预先设置 Swarm 或 UCP*。

```bash
#Create an overlay named "ovnet" with the overlay driver
$ docker network create -d overlay --subnet 10.1.0.0/24 ovnet

#Create a service from an nginx image and connect it to the "ovnet" overlay network
$ docker service create --network ovnet nginx
```

创建 overlay 网络时，请注意在主机内部创建了多个接口和网桥，以及此容器内的两个接口。

```bash
# Peek into the container of this service to see its internal interfaces
conatiner$ ip address

#docker_gwbridge network
52: eth0@if55: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500
    link/ether 02:42:ac:14:00:06 brd ff:ff:ff:ff:ff:ff
    inet 172.20.0.6/16 scope global eth1
       valid_lft forever preferred_lft forever
    inet6 fe80::42:acff:fe14:6/64 scope link
       valid_lft forever preferred_lft forever

#overlay network interface
54: eth1@if53: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450
    link/ether 02:42:0a:01:00:03 brd ff:ff:ff:ff:ff:ff
    inet 10.1.0.3/24 scope global eth0
       valid_lft forever preferred_lft forever
    inet 10.1.0.2/32 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::42:aff:fe01:3/64 scope link
       valid_lft forever preferred_lft forever
```

在容器内部创建了两个接口，这两个接口对应于主机上现在存在的两个网桥。在 overlay 网络上，每个容器至少有两个接口，分别将它连接到覆盖层和 docker_gwbridge。

|        网桥         | 目的                                                         |
| :-----------------: | :----------------------------------------------------------- |
|     **overlay**     | 入口和出口指向 VXLAN 封装的 overlay 网络，并可选地加密同一 overlay 网络上的容器之间的流量。它将覆盖范围扩展到参与此特定叠加层的所有主机。在主机上每个 overlay 子网存在一个，并且它具有与给定特定 overlay 网络相同的名称。 |
| **docker_gwbridge** | 离开集群的流量的出口桥。每个主机只有一个 `docker_gwbridge` 存在。此桥上，容器到容器的流量被阻断，仅允许入站/出站流量。 |

> Docker Overlay 驱动程序自 Docker Engine 1.9 以来就已存在，并且需要外部 K/V 存储来管理网络状态。Docker Engine 1.12 将控制面板状态集成到 Docker Engine 中，因此不再需要外部存储。1.12 还引入了一些新功能，包括加密和服务负载平衡。引入的网络功能需要支持它们的 Docker Engine 版本，并且不支持将这些功能与旧版本的 Docker Engine 一起使用。

## Docker 服务的外部访问

Swarm 和 UCP 支持对群集端口发布之外的服务访问。服务的入站和出站不依赖于集中式网关，而是依赖于运行特定服务任务的主机上的分布式入站/出站。服务有两种端口发布模式，`host` 模式和 `ingress` 模式。

### Ingress 模式服务发布

`ingress` 模式端口发布利用 [Swarm Routing Mesh](https://success.docker.com/api/asset/.%2Frefarch%2Fnetworking%2F#routingmesh) 在服务中的任务之间实现负载均衡。`ingress` 模式在*每个* UCP/Swarm 节点上发布暴露的端口。传到发布端口的入站流量由路由网格进行负载平衡，并通过循环负载平衡定向到服务中*健康的*任务之一。即使给定主机未运行服务任务，端口也会在主机上发布，并对具有任务的主机进行负载平衡。

```bash
$ docker service create --replicas 2 --publish mode=ingress,target=80,published=8080 nginx
```

> `mode=ingress` 是服务的默认模式。此命令也可以使用速记版本 `-p 80:8080` 来完成。端口 `8080` 在群集上的每个主机上公开，并在此服务中对两个容器进行负载平衡。

### Host 模式服务发布

`host` 模式端口发布仅在运行特定服务任务的主机上公开端口。端口直接映射到该主机上的容器。每个主机上只能运行给定服务的单个任务，以防止端口冲突。

```bash
$ docker service create --replicas 2 --publish mode=host,target=80,published=8080 nginx
```

> `host` 模式需要 `mode=host` 标志。它在运行这两个容器的主机上本地发布端口 `8080`。它不应用负载平衡，因此到这些节点的流量仅指向本地容器。如果没有足够的端口可用于指定数量的副本，则可能导致端口冲突。

### 入站设计

发布模式有很多好的用例。`ingress` 模式适用于具有多个副本并需要在这些副本之间进行负载平衡的服务。如果其他工具已提供外部服务发现，则 `host` 模式已可以满足需求。`host` 模式的另一个良好用例是每个主机存在一次的全局容器。这些容器可能会公开与本主机相关的特定信息（例如监视或日志记录），因此您不希望在访问该服务时进行负载平衡。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/ingress-vs-host.png)

## MACVLAN

`macvlan` 驱动程序是经过检验的真正网络虚拟化技术的新实现。Linux 上的实现非常轻量级，因为它们不是使用 Linux 网桥进行隔离，而是简单地与 Linux 以太网接口或子接口相关联，以强制实现网络之间的分离以及与物理网络的连接。

MACVLAN 提供许多独特的功能。由于具有非常简单和轻量级的架构，它对性能提升有所帮助。MACVLAN 驱动程序不是端口映射，而是提供容器和物理网络之间的直接访问。它还允许容器接收物理网络子网上的可路由 IP 地址。

MACVLAN 的使用场景包括：

- 超低延时应用
- 设计一个网络，要求容器在同一子网内，并使用和外部主机网络相同的 IP 地址

`macvlan` 驱动程序使用父接口的概念。此接口可以是物理接口，例如 `eth0`，用于 802.1q VLAN 标记的子接口，如 `eth0.10`（`.10` 表示 `VLAN 10`），或者甚至是绑定的主机适配器，它将两个以太网接口捆绑到一个逻辑接口中。

在 MACVLAN 网络配置期间需要网关地址。网关必须位于网络基础架构提供的主机外部。MACVLAN 网络允许在同一网络上的容器之间进行访问。如果没有在主机外部路由，则无法在同一主机上的不同 MACVLAN 网络之间进行访问。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/macvlanarch.png)

此示例将 MACVLAN 网络绑定到主机上的 `eth0`。它还将两个容器连接到名为 `mvnet`  的 MACVLAN 网络，并展示了它们可以相互 ping 通。每个容器在 `192.168.0.0/24` 物理网络子网上都有一个地址，其默认网关是物理网络中的接口。

```bash
#Creation of MACVLAN network "mvnet" bound to eth0 on the host
$ docker network create -d macvlan --subnet 192.168.0.0/24 --gateway 192.168.0.1 -o parent=eth0 mvnet

#Creation of containers on the "mvnet" network
$ docker run -itd --name c1 --net mvnet --ip 192.168.0.3 busybox sh
$ docker run -it --name c2 --net mvnet --ip 192.168.0.4 busybox sh
/ # ping 192.168.0.3
PING 127.0.0.1 (127.0.0.1): 56 data bytes
64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.052 ms
```

正如您在此图中所见，`c1` 和 `c2` 通过 MACVLAN 网络连接，该网络名为 `macvlan`，连接到主机上的 `eth0`。

### 使用 MACVLAN 进行 VLAN 中继

对于许多运营商而言，将 802.1q 中继到 Linux 主机是非常痛苦的。它需要更改配置文件才能在重新启动时保持持久性。如果涉及网桥，则需要将物理网卡移入网桥，然后网桥获取 IP 地址。`macvlan` 驱动程序通过创建、销毁和主机重新启动来完全管理 MACVLAN 网络的子接口和其他组件。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/trunk-macvlan.png)

当使用子接口实例化 `macvlan` 驱动程序时，它允许 VLAN 中继到主机并在 L2 层隔离容器。`macvlan` 驱动程序自动创建子接口并将它们连接到容器接口。因此，每个容器都位于不同的 VLAN 中，除非在物理网络中路由流量，否则它们之间无法进行通信。

```bash
#Creation of  macvlan10 network in VLAN 10
$ docker network create -d macvlan --subnet 192.168.10.0/24 --gateway 192.168.10.1 -o parent=eth0.10 macvlan10

#Creation of  macvlan20 network in VLAN 20
$ docker network create -d macvlan --subnet 192.168.20.0/24 --gateway 192.168.20.1 -o parent=eth0.20 macvlan20

#Creation of containers on separate MACVLAN networks
$ docker run -itd --name c1--net macvlan10 --ip 192.168.10.2 busybox sh
$ docker run -it --name c2--net macvlan20 --ip 192.168.20.2 busybox sh
```

在上面的配置中，我们使用 `macvlan` 驱动程序创建了两个独立的网络，这些驱动程序配置为使用子接口作为其父接口。`macvlan` 驱动程序创建子接口并在主机的 `eth0` 和容器接口之间连接它们。必须将主机接口和上游交换机设置为 `switchport mode trunk`，以便在接口上标记 VLAN。可以将一个或多个容器连接到给定的 MACVLAN 网络，以创建通过 L2 分段的复杂网络策略。

> 由于单个主机可能拥有多个 MAC 地址，因此您可能需要在接口上启用混杂模式，具体取决于 NIC 对 MAC 过滤的支持。

## None 网络驱动（隔离）

与 `host` 网络驱动程序类似，`none` 网络驱动程序本质上是一种不经管理的网络选项。Docker Engine 不会在容器内创建接口、建立端口映射或安装连接路由。使用 `--net=none` 的容器与其他容器和主机完全隔离。网络管理员或外部工具必须负责提供此管道。使用 none 的容器只有一个 `loopback` 接口而没有其他接口。

与 `host` 驱动程序不同，`none` 驱动程序为每个容器创建单独的命名空间。这可以保证任何容器和主机之间的网络隔离。

> 使用 --net=none 或 --net=host 的容器无法连接到任何其他 Docker 网络。

## 物理网络设计要求

Docker EE 和 Docker 网络旨在运行在通用数据中心网络基础架构和拓扑上。其集中控制器和容错集群可确保在各种网络环境中兼容。提供网络功能的组件（网络配置，MAC 学习，覆盖加密）要么是 Docker Engine 的一部分，即 UCP，要么是 Linux 内核本身。运行任何原生 Docker 网络驱动程序都不需要额外的组件或特殊网络功能。

更具体地说，Docker 原生网络驱动程序对以下内容没有任何要求：

- 多播
- 外部键值存储（Key-Value）
- 特定的路由协议
- 主机间在 L2 层相互连通
- 特定的拓扑结构，如骨干枝叶模型（spine-and-leaf），传统的三层模型或 PoD 设计。这些拓扑结构都是支持的。

这与容器网络模型一致，该模型可在所有环境中提升应用程序可移植性，同时仍实现应用程序所需的性能和策略。

## Swarm 原生服务发现

Docker 使用嵌入式 DNS 为在单个 Docker Engine 上运行的容器和在 Docker Swarm 中运行的任务提供服务发现。Docker Engine 具有内部 DNS 服务器，可为用户定义的网桥，overlay 和 MACVLAN 网络中的主机上的所有容器提供名称解析。每个 Docker 容器（或 Swarm 模式下的任务）都有一个 DNS 解析器，它将 DNS 查询转发给 Docker Engine，后者充当 DNS 服务器。然后，Docker Engine 检查 DNS 查询是否属于请求容器所属的网络上的容器或服务。如果是，则 Docker Engine 在其键值存储中查找与容器、任务或服务的**名称**匹配的 IP 地址，并将该 IP 或服务虚拟 IP（VIP）返回给请求者。

服务发现是*网络范围*的，这意味着只有位于同一网络上的容器或任务才能使用嵌入式 DNS 功能。不在同一网络上的容器无法解析彼此的地址。此外，只有在特定网络上具有容器或任务的节点才会存储该网络的 DNS 条目。这可以提高安全性和性能。

如果目标容器或服务不属于与源容器相同的网络，则 Docker Engine 会将 DNS 查询转发到配置的默认 DNS 服务器。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/DNS.png)

在这个例子中，有两个名为 `myservice` 的容器服务。另一个服务（`client`）存在于同一网络上。`client` 向 `docker.com` 和 `myservice` 执行两个 curl 操作。以下是由此产生的结果：

- `client` 初始化关于 `docker.com` 和 `myservice` 的 DNS 查询。
- 容器的内置解析器拦截 127.0.0.11:53 上的 DNS 查询，并将它们发送到 Docker Engine 的 DNS 服务器。
- `myservice` 解析为该服务的虚拟 IP（VIP），该服务由内部负载均衡分发到各个任务 IP 地址。容器名称也会解析，尽管直接与其 IP 地址相关。
- `docker.com` 不是 `mynet` 网络中的服务名称，因此请求将转发到配置的默认 DNS 服务器。

## Docker 原生负载均衡

Docker Swarm 集群具有内置的内部和外部负载均衡功能，这些功能内置于 Engine 中。内部负载均衡提供同一 Swarm 或 UCP 集群内容器之间的负载均衡。外部负载均衡提供进入群集的入站流量的负载均衡。

### UCP 内部负载均衡

创建 Docker 服务时，会自动实例化内部负载均衡。在 Docker Swarm 集群中创建服务时，会自动为它们分配一个虚拟 IP（VIP），该虚拟 IP 是服务网络的一部分。解析服务名称时返回 VIP。到该 VIP 的流量将自动发送到 overlay 网络上该服务的所有健康任务。这种方法避免了任何客户端负载均衡，因为只有一个 IP 返回给客户端。Docker 负责路由并在健康的服务任务中平均分配流量。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/ipvs.png)

要查看 VIP，请运行 `docker service inspect my_service`，如下所示：

```bash
# Create an overlay network called mynet
$ docker network create -d overlay mynet
a59umzkdj2r0ua7x8jxd84dhr

# Create myservice with 2 replicas as part of that network
$ docker service create --network mynet --name myservice --replicas 2 busybox ping localhost
8t5r8cr0f0h6k2c3k7ih4l6f5

# See the VIP that was created for that service
$ docker service inspect myservice
...

"VirtualIPs": [
    {
        "NetworkID": "a59umzkdj2r0ua7x8jxd84dhr",
        "Addr": "10.0.0.3/24"
    },
]
```

> DNS 循环（DNS RR）负载平衡是服务的另一个负载平衡选项（使用 `--endpoint-mode` 配置）。在 DNS RR 模式下，不为每个服务创建 VIP。Docker DNS 服务器以循环方式将服务名称解析为单个容器 IP。

### UCP 外部 L4 负载均衡（Docker 路由网络）

在创建或更新服务时，可以使用 `--publish` 标志在外部发布服务。在 Docker Swarm 模式下发布端口意味着群集中的每个节点都在侦听该端口。但是，如果服务的任务不在正在侦听该端口的节点上，会发生什么？

这是路由网格发挥作用的地方。路由网格是 Docker 1.12 中的一项新功能，它结合了 `ipvs` 和 `iptables` 来创建一个功能强大的集群范围的传输层（L4）负载均衡器。它允许所有 Swarm 节点接受服务发布端口上的连接。当任何 Swarm 节点接收到发往正在运行的服务的已发布 TCP/UDP 端口的流量时，它会使用名为 `ingress` 的预定义 overlay 网络将其转发到服务的 VIP。`ingress` 网络的行为类似于其他 overlay 网络，但其唯一目的是将网状路由流量从外部客户端传输到集群服务。与上一节中的描述相同，它使用基于 VIP 的内部负载均衡。

启动服务后，您可以为应用程序创建外部 DNS 记录，并将其映射到任何或所有 Docker Swarm 节点。您无需担心容器的运行位置，因为群集中的所有节点都与路由网状路由功能一样。

```bash
#Create a service with two replicas and export port 8000 on the cluster
$ docker service create --name app --replicas 2 --network appnet -p 8000:80 nginx
```

![logo](https://github.com/studygolang/gctt-images/blob/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/routing-mesh.png?raw=true)

上图展示了路由网络如何运作。

- 服务创建是有两个副本，端口映射到外部端口 `8000`。
- 服务网络将每个主机的 `8000` 端口暴露在集群中。
- 发往 `app` 的流量可以从任意主机进入。当使用外部负载均衡时，流量有可能会被发往不存在服务副本的主机。
- 内核的 IPVS 负载均衡器会将流量重定向到处于 `ingress` overlay 网络中的健康服务副本。

### UCP 外部 L7 负载均衡（HTTP 路由网络）

UCP 通过 HTTP 路由网络提供 L7 HTTP/HTTPS 负载均衡。URL 可以对服务进行负载均衡，并在服务副本之间进行负载均衡。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/ucp-hrm.png)

转到 [UCP 负载均衡参考架构](https://success.docker.com/Architecture/Docker_Reference_Architecture%3A_Universal_Control_Plane_2.0_Service_Discovery_and_Load_Balancing) 了解有关 UCP L7 负载均衡设计的更多信息。

## Docker 网络安全与加密

在使用 Docker 设计和实现容器化工作负载时，网络安全性是首要考虑因素。在本节中，介绍了部署 Docker 网络时的主要安全注意事项。

### 网络隔离和数据平面安全

Docker 管理分布式防火墙规则以分隔 Docker 网络并防止恶意访问容器资源。默认情况下，Docker 网络彼此隔离以阻断它们之间的流量。这种方法在第 3 层提供真正的网络隔离。

Docker Engine 管理主机防火墙规则，阻止网络之间的访问并管理暴露容器的端口。在 Swarm 和 UCP 群集中，这会创建一个分布式防火墙，在应用程序在群集中进行调度时动态保护应用程序。

下表概述了 Docker 网络的一些访问策略。

|           路径           | 访问                                                         |
| :----------------------: | :----------------------------------------------------------- |
|    **Docker 网络内**     | 在同一个 Docker 网络上的所有容器之间，可通过任意端口进行访问。这适用于所有网络类型：集群范围、本地范围、内置和远程驱动程序。 |
|    **Docker 网络间**     | Docker 网络之间拒绝相互访问，这通过 Docker Engine 管理的分布式主机防火墙规则来约束。容器可以连接到多个网络，以便在不同的 Docker 网络之间进行通信。Docker 网络之间的网络连接也可以在主机外部进行管理。 |
|  **从 Docker 网络外发**  | 允许从 Docker 网络内部发往 Docker 主机外部的流量。主机的本地状态防火墙跟踪连接以允许该连接的响应。 |
| **外部进入 Docker 网络** | 默认情况下拒绝入站流量。通过 `host` 端口或 `ingress` 模式的端口进行端口暴露可以提供显式的入站访问权限。一个例外是 MACVLAN 驱动程序，它在与外部网络相同的 IP 空间中运行，并在该网络中完全开放。与 MACVLAN 类似的其他远程驱动程序也可以允许入站流量。 |

### 控制面板安全

Docker Swarm 附带集成的 PKI。Swarm 中的所有管理器和节点都以签名证书的形式具有加密签名的身份。所有管理器到管理器和管理器到节点的控制通信都是通过 TLS 开箱即用的。无需在外部生成证书或手动设置任何 CA 以获得在 Docker Swarm 模式下已受保护的端到端控制面板流量。其中的证书还会定期自动轮换。

### 数据平面网络加密

Docker 支持开箱即用的 overlay 网络的 IPSec 加密。Swarm 和 UCP 管理的 IPSec 隧道在离开源容器时对网络流量进行加密，并在进入目标容器时对其进行解密。这可确保您的应用程序流量在传输过程中非常安全，无论底层网络如何。在混合、多租户或多云环境中，确保数据安全至关重要，因为它将经由您可能无法控制的网络。

此图说明了如何保护在 Docker Swarm 中的不同主机上运行的两个容器之间的通信。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/ipsec.png)

通过添加 `--opt encrypted=true` 选项（例如 `docker network create -d overlay --opt encrypted=true <NETWORK_NAME>`），可以在创建时为每个网络启用此功能。创建网络后，您可以在该网络上启动服务（例如，`docker service create --network <NETWORK_NAME> <IMAGE> <COMMAND>`）。当在两个不同的主机上创建相同服务的两个任务时，会在它们之间创建 IPsec 隧道，并且流量在离开源主机时会被加密，并在进入目标主机时被解密。

集群中的 leader 节点会定期重新生成对称密钥，并将其安全地分发到所有集群节点。IPsec 使用此密钥加密和解密数据平面流量。使用 AES-GCM 在主机到主机传输模式下通过 IPSec 实现加密。

### 管理平面安全以及用 UCP 进行基于角色的权限访问控制

使用 UCP 创建网络时，团队和标签定义对容器资源的访问。资源许可标签定义了谁可以查看、配置和使用某些 Docker 网络。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/ucp-network.png)

此 UCP 屏幕截图显示了使用标签生产团队来控制对该网络的成员的访问。此外，可以通过 UCP 切换网络加密等选项。

## IP 地址管理

容器网络模型（CNM）提供了管理 IP 地址的灵活性。IP 地址管理有两种方法：

- CNM 具有原生 IPAM 驱动程序，可以为集群全局简单分配 IP 地址，并防止重复分配。如果未指定其他驱动程序，则默认使用本机 IPAM 驱动程序。
- CNM 具有使用来自其他供应商和社区的远程 IPAM 驱动程序的接口。这些驱动程序可以提供与现有供应商或自建 IPAM 工具的集成。

可以使用 UCP、CLI 或 Docker API 手动配置容器 IP 地址和网络子网。地址请求通过所选的驱动程序，然后决定如何处理请求。

子网大小和设计在很大程度上取决于给定的应用程序和特定的网络驱动程序。下一节将详细介绍每种 [网络部署模型](https://success.docker.com/api/asset/.%2Frefarch%2Fnetworking%2F#models) 的 IP 地址空间设计。端口映射，overlay 和 MACVLAN 的使用都会影响 IP 寻址的安排。通常，容器寻址分为两个部分。内部容器网络（bridge 和 overlay）使用默认情况下在物理网络上无法路由的 IP 地址的容器进行寻址。MACVLAN 网络为物理网络子网上的容器提供 IP 地址。因此，来自容器接口的流量可以在物理网络上路由。值得注意的是，内部网络（bridge，overlay）的子网不应与物理底层网络的 IP 空间冲突。重叠的地址空间可能导致流量无法到达目的地。

## 网络故障排除

对于运维和网络工程师来说，Docker 网络故障排查很困难。通过准确理解 Docker 网络的工作原理和运用恰当的工具集，您可以排查故障并解决这些棘手的网络问题。一种值得推荐的方法是使用 [netshoot](https://github.com/nicolaka/netshoot) 容器来解决网络问题。`netshoot` 容器具有一组强大的网络故障排查工具，可用于解决 Docker 网络问题。

使用诸如 `netshoot` 的故障排除容器的优势在于，能使网络故障排查工具变得可移植。`netshoot` 容器可以连接到任何网络，可以放在主机网络命名空间中，也可以放在另一个容器的网络命名空间中，以检查主机网络的任何方面。

它包含且不限于以下工具：

- iperf
- tcpdump
- netstat
- iftop
- drill
- util-linux(nsenter)
- curl
- nmap

## 网络部署模型

以下示例使用名为 **[Docker Pets](https://github.com/mark-church/docker-pets)** 的虚构应用程序来说明**网络部署模型**。它在网页上提供宠物图像，同时计算后端数据库中页面的点击次数。

- `web` 是一个前端 Web 服务器，基于 `chrch/docker-pets:1.0` 镜像
- `db` 是一个 `consul` 后端

`chrch/docker-pets` 需要一个环境变量 `DB` 来告诉它如何查找后端数据库服务。

### 单主机上的 Bridge 驱动

此模型是 Docker 原生 `bridge` 网络驱动程序的默认配置。 `bridge` 驱动程序在主机内部创建专用网络，并在主机接口上提供外部端口映射以进行外部连接。

```bash
$ docker network create -d bridge petsBridge

$ docker run -d --net petsBridge --name db consul

$ docker run -it --env "DB=db" --net petsBridge --name Web -p 8000:5000 chrch/docker-pets:1.0
Starting Web container e750c649a6b5
 * Running on http://0.0.0.0:5000/ (Press CTRL+C to quit)
```

如果未指定 IP 地址，则会在主机的所有接口上公开端口映射。在这种情况下，容器的应用程序在 0.0.0.0:8000 上发布。如要使用特定 IP 地址，需要提供额外的标志 `-p IP:host_port:container_port`。可以在 [Docker 文档](https://docs.docker.com/engine/reference/run/#/expose-incoming-ports) 中找到更多暴露端口的选项。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/singlehost-bridge.png)

应用程序本地发布在主机所有接口上的 8000 端口。还设置了 `DB=db`，提供后端容器的名称。Docker Engine 的内置 DNS 将此容器名称解析为 `db` 的 IP 地址。由于 `bridge` 是本地驱动程序，因此 DNS 解析的范围仅限于单个主机。

下面的输出显示，我们的容器已经从 `petsBridge` 网络的 `172.19.0.0/24` 网段分配了私有 IP。如果未指定其他 IPAM 驱动程序，Docker 将使用内置 IPAM 驱动程序从相应的子网提供 IP。

```bash
$ docker inspect --format {{.NetworkSettings.Networks.petsBridge.IPAddress}} web
172.19.0.3

$ docker inspect --format {{.NetworkSettings.Networks.petsBridge.IPAddress}} db
172.19.0.2
```

这些 IP 地址用于 `petsBridge` 网络内部的通信，永远不会暴露在主机之外。

### 带有外部服务发现的多主机 Bridge 驱动

由于 `bridge` 驱动程序是本地范围驱动程序，因此多主机网络需要多主机服务发现解决方案。外部 SD 注册容器或服务的位置和状态，然后允许其他服务发现该位置。由于 `bridge` 驱动程序公开端口以进行外部访问，因此外部 SD 将 host-ip:port 存储为给定容器的位置。

在以下示例中，手动配置每个服务的位置，模拟外部服务发现。`db` 服务的位置通过 `DB` 环境变量传递给 `web`。

```bash
#Create the backend db service and expose it on port 8500
host-A $ docker run -d -p 8500:8500 --name db consul

#Display the host IP of host-A
host-A $ ip add show eth0 | grep inet
    inet 172.31.21.237/20 brd 172.31.31.255 scope global eth0
    inet6 fe80::4db:c8ff:fea0:b129/64 scope link

#Create the frontend Web service and expose it on port 8000 of host-B
host-B $ docker run -d -p 8000:5000 -e 'DB=172.31.21.237:8500' --name Web chrch/docker-pets:1.0
```

`web` 服务现在应该在 `host-B` IP 地址的 8000 端口上提供其网页。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/multi-host-bridge.png)

> 在此示例中，我们没有指定要使用的网络，因此会自动选择默认的 Docker `bridge` 网络。

当我们在 `172.31.21.237:8500` 配置 `db` 的位置时，我们正在创建某种形式的**服务发现**。我们静态配置 `web` 服务的 `db` 服务的位置。在单主机示例中，这是自动完成的，因为 Docker Engine 为容器名称提供了内置的 DNS 解析。在这个多主机示例中，我们需要手动执行服务发现。

不建议在生成环境中将应用程序的位置硬编码。外部服务发现工具可以在集群创建和销毁容器时，动态地提供这些映射，例如 [Consul](https://www.consul.io/) 和 [etcd](https://coreos.com/etcd/)。

下一节将介绍 `overlay` 驱动程序方案，该方案在集群中提供全局服务发现作为内置功能。与使用多个外部工具提供网络服务相比，这种简单性是 `overlay` 驱动程序的主要优点。

### 多主机下的 Overlay 驱动

该模型利用本机 `overlay` 驱动程序提供开箱即用的多主机连接。`overlay` 驱动程序的默认设置提供与外部世界的外部连接以及容器应用程序内的内部连接和服务发现。[Overlay 驱动架构](https://success.docker.com/api/asset/.%2Frefarch%2Fnetworking%2F#overlayarch) 将回顾 Overlay 驱动程序的内部结构，在阅读本节之前，您应该先查看它。

此示例重新使用以前的 `docker-pets` 应用程序。在遵循此示例之前设置 Docker 集群。有关如何设置 Swarm 的说明，请阅读 [Docker 文档](https://docs.docker.com/engine/swarm/swarm-tutorial/create-swarm/)。设置 Swarm 后，使用 `docker service create` 命令创建由 Swarm 管理的容器和网络。

下面显示了如何检查 Swarm、创建覆盖网络，然后在该 overlay 网络上配置服务。所有这些命令都在 UCP/swarm 控制器节点上运行。

```bash
#Display the nodes participating in this swarm cluster that was already created
$ docker node ls
ID                           HOSTNAME          STATUS  AVAILABILITY  MANAGER STATUS
a8dwuh6gy5898z3yeuvxaetjo    host-B  Ready   Active
elgt0bfuikjrntv3c33hr0752 *  host-A  Ready   Active        Leader

#Create the dognet overlay network
host-A $ docker network create -d overlay petsOverlay

#Create the backend service and place it on the dognet network
host-A $ docker service create --network petsOverlay --name db consul

#Create the frontend service and expose it on port 8000 externally
host-A $ docker service create --network petsOverlay -p 8000:5000 -e 'DB=db' --name Web chrch/docker-pets:1.0

host-A $ docker service ls
ID            NAME  MODE        REPLICAS  IMAGE
lxnjfo2dnjxq  db    replicated  1/1       consul:latest
t222cnez6n7h  Web   replicated  0/1       chrch/docker-pets:1.0
```

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/overlay-pets-example.png)

与在单主机 bridge 驱动示例中一样，我们将 `DB=db` 作为环境变量传递给 `web` 服务。overlay 驱动程序将服务名称 `db` 解析为容器的 overlay IP 地址。`web` 和 `db` 之间的通信仅使用 overlay IP 子网进行。

> 在 overlay 和 bridge 网络内部，所有到容器的 TCP 和 UDP 端口都是开放的，并且可以连接到 overlay 网络的所有其他容器。

`web` 服务暴露在 8000 端口上，**路由网络**在 Swarm 集群中的每个主机上公开端口 8000。通过在浏览器中转到 <host-A>:8000 或 <host-B>:8000 来测试应用程序是否正常工作。

#### Overlay 的好处及使用场景

- 无论规模大小，多主机直接的连通性都很容易实现。
- 不需要额外配置或组件，就能实现服务发现和负载均衡。
- 有助于通过加密叠加层实现东西方微分段。
- 路由网格可用于在整个群集中发布服务。

### 练习应用：MACVLAN 网桥模式

在某些情况下，应用程序或网络环境要求容器具有可作为底层子网一部分的可路由 IP 地址。MACVLAN 驱动程序实现了此功能。如 [MACVLAN 体系结构](https://success.docker.com/api/asset/.%2Frefarch%2Fnetworking%2F#macvlan) 部分所述，MACVLAN 网络将自身绑定到主机接口。这可以是物理接口，逻辑子接口或绑定的逻辑接口。它充当虚拟交换机，并在同一 MACVLAN 网络上的容器之间提供通信。每个容器接收唯一的 MAC 地址和该节点所连接的物理网络的 IP 地址。

![logo](https://raw.githubusercontent.com/studygolang/gctt-images/master/Docker-Reference-Architecture-Designing-Scalable-Portable-Docker-Container-Networks/2node-macvlan-app.png)

本例中，Pets 应用部署在 `host-A` 和 `host-B` 上。

```bash
#Creation of local macvlan network on both hosts
host-A $ docker network create -d macvlan --subnet 192.168.0.0/24 --gateway 192.168.0.1 -o parent=eth0 petsMacvlan
host-B $ docker network create -d macvlan --subnet 192.168.0.0/24 --gateway 192.168.0.1 -o parent=eth0 petsMacvlan

#Creation of db container on host-B
host-B $ docker run -d --net petsMacvlan --ip 192.168.0.5 --name db consul

#Creation of Web container on host-A
host-A $ docker run -it --net petsMacvlan --ip 192.168.0.4 -e 'DB=192.168.0.5:8500' --name Web chrch/docker-pets:1.0
```

这可能看起来与多主机桥示例非常相似，但有几个显着的差异：

- 从 `web` 到 `db` 的引用使用 `db` 本身的 IP 地址而不是主机 IP。请记住，使用 `macvlan` 驱动时，容器 IP 可以在底层网络上路由。
- 我们不暴露 `db` 或 `web` 的任何端口，因为容器中打开的任何端口都可以使用容器 IP 地址直接访问。

虽然 `macvlan` 驱动程序提供了这些独特的优势，但它牺牲的是可移植性。MACVLAN 配置和部署与底层网络密切相关。除了防止重叠地址分配之外，容器寻址必须遵守容器放置的物理位置。因此，必须注意在 MACVLAN 网络外部维护 IPAM。重复的 IP 地址或不正确的子网可能导致容器连接丢失。

#### MACVLAN 的优势与使用场景

- `macvlan` 驱动由于没有用到 NAT，适用于超低延迟应用。
- MACVLAN 可为每个容器提供 IP 地址，这可能是在某些场景中的特殊要求。
- 必须更加细致地考虑 IPAM。

## 结论

Docker 是一种快速发展的技术，网络选项日益增多，每天都在满足越来越多的用例需求。现有的网络供应商、纯粹的 SDN 供应商和 Docker 本身都是这一领域的贡献者。与物理网络的紧密集成、网络监控和加密都是备受关注和创新的领域。

本文档详述了一些可行的部署方案和现有的 CNM 网络驱动程序，但并不完全。虽然有许多单独的驱动程序以及更多配置这些驱动程序的方法，但我们希望您可以看到，用于常规部署的常用模型很少。了解每种模型之间的优劣权衡是取得长期成功的关键所在。

---

via: https://success.docker.com/article/networking

作者：[Mark Church](https://success.docker.com/author/markchurch)
译者：[Mockery-Li](https://github.com/Mockery-Li)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
