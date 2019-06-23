首发于：https://studygolang.com/articles/21376

# Docker 容器编排实践练习

在本次练习中，你将体验到 Docker 的容器编排功能。首先你需要在单个主机上部署一个简单的应用程序，并了解其工作机制。然后，通过配置 Docker Swarm 模式，你将学习到怎样在多个主机上部署相同的简单应用程序。最后，你将看到如何对应用的规模进行扩容、缩容，以及如何将工作负载在不同的主机之间转移。

（译者注：登录原网站可使用在线练习资源）

> **难度**：入门级
> **时间**：约 30 分钟

**任务**：

- [第一节 - 容器编排是什么](# 第一节：容器编排是什么 )
- [第二节 - 配置 Docker Swarm 模式](# 第二节：配置 -Docker-Swarm- 模式 )
- [第三节 - 跨多主机部署应用](# 第三节：跨多主机部署应用 )
- [第四节 - 应用扩容缩容](# 第四节：应用扩容缩容 )
- [第五节 - 排空节点并重新调度容器](# 第五节：排空节点并重新调度容器 )
- [清理工作](# 清理工作 )

## 第一节：容器编排是什么

那么问题来了，到底什么是容器编排呢？还是用一个例子加以说明吧。假设你有一个大流量的应用，并且要求具有高可用性。为满足这些需要，你通常希望在至少 3 台机器上进行部署，以便在主机发生故障时，你的应用程序仍可从其他至少两台计算机访问。 显然，这只是一个例子，你的用例可能有自己的要求，但相信你已经明白意思了。

在没有容器编排技术的情况下，部署应用程序通常非常耗时且容易出错，因为你必须手动 SSH 到每台计算机，启动应用程序，然后不断监视以确保程序按预期运行。

但有了编排工具之后，你通常可以摆脱大部分人力劳动，让自动化完成繁重的工作。Docker Swarm 容器编排有一个很酷的功能是，你只需一个命令即可在多个主机上部署应用程序（前提是启用了 Swarm 模式）。另外，如果 Docker Swarm 中的其中一个节点停止了运行，其他节点将自动接过负载，你的应用程序将继续运作如常。

如果你通常只使用 `docker run` 来部署应用程序，那么使用 Docker Compose 或 Docker Swarm 模式或二者兼用将很有可能是你受益不浅。

## 第二节：配置 Docker Swarm 模式

如前所述，实际应用程序通常部署在多个主机上。 这可以提高应用程序性能和可用性，并允许各个组件独立扩展。Docker 拥有强大的原生工具来帮助你实现这一目标。

手动在单个主机上运行程序的一个例子，是在**node1**上运行 `docker run -dt Ubuntu sleep infinity`，来创建一个新的容器。

```shell
docker run -dt Ubuntu sleep infinity
```

```shell
Unable to find image 'ubuntu:latest' locally
latest: Pulling from library/ubuntu
d54efb8db41d: Pull complete
f8b845f45a87: Pull complete
e8db7bf7c39f: Pull complete
9654c40e9079: Pull complete
6d9ef359eaaa: Pull complete
Digest: sha256:dd7808d8792c9841d0b460122f1acf0a2dd1f56404f8d1e56298048885e45535
Status: Downloaded newer image for Ubuntu:latest
846af8479944d406843c90a39cba68373c619d1feaa932719260a5f5afddbf71
```

此命令将基于 `ubuntu：latest` 镜像创建一个新容器，并将运行 sleep 命令以使容器在后台运行。 您可以通过在 **node1** 上运行 `docker ps` 来验证我们的示例容器已启动。

```shell
docker ps
```

```shell
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
044bea1c2277        Ubuntu              "sleep infinity"    2 seconds ago       Up 1 second                             distracted_mayer
```

但是，这只是在一个节点上。 如果此节点出现故障会怎样？ 好吧，我们的应用程序会被终止，它永远不会重新启动。 要恢复服务，我们必须手动登录此计算机，折腾一番才能使其恢复运行。 因此，如果我们有某种类型的系统允许我们在许多机器上运行这个“睡眠”应用程序 / 服务，那将会很有帮助。

在本节中，你将练习配置 Swarm 模式。 这是一种新的可选模式，其中多个 Docker 主机形成一个称为群组的自编排引擎。Swarm 模式支持新功能，如服务和捆绑，可帮助你跨多个 Docker 主机部署和管理多容器应用程序。

你需要完成以下几步：

- 配置 Swarm 模式
- 运行应用程序
- 对应用规模进行伸缩
- 出于维护需要，排空（drain）一个节点，并重新调度节点上的容器

在本练习的余下部分中，我们将 Docker 原生集群管理称作 ***Swarm 模式***。配置了 Swarm 模式的一组 Docker 机器将被称为*群组*。

一个群组包含了一个或多个*管理节点*，以及一个或多个*工作节点*。管理节点维护群组状态并对应用容器进行调度。工作节点则负责应用容器的实际运行。在 Docker 1.12 中，使用群组管理的完整功能不需要任何外部后端或第三方组件，也就是说这完全是内置的！

在这一小节的演示里，实验中的三个节点你全部都需要用到。**node1** 将成为群组管理节点，**node2** 和 **node3** 将成为工作节点。Swarm 模式支持高可用的冗余管理节点的设置，但如果只是以此次练习为目的，你只需要部署单个管理节点就够了。

### 步骤 2.1 创建管理节点

在这一步骤中，你将要初始化一个新的 Swarm 群组，添加一个工作节点，并验证操作是否生效。

在 node1 上运行 `docker swarm INIt`

```shell
docker swarm INIt --advertise-addr $(hostname -i)
```

```shell
Swarm INItialized: current node (6dlewb50pj2y66q4zi3egnwbi) is now a manager.

To add a worker to this swarm, run the following command:

    docker swarm join \
    --token SWMTKN-1-1wxyoueqgpcrc4xk2t3ec7n1poy75g4kowmwz64p7ulqx611ih-68pazn0mj8p4p4lnuf4ctp8xy \
    10.0.0.5:2377

To add a manager to this swarm, run 'docker swarm join-token manager' and follow the instructions.
```

你可以运行 `docker info` 命令来验证 **node1** 是否成功被配置成群组管理节点。

```shell
docker info
```

```shell
Containers: 2
 Running: 0
 Paused: 0
 Stopped: 2
Images: 2
Server Version: 17.03.1-ee-3
Storage Driver: aufs
 Root Dir: /var/lib/docker/aufs
 Backing Filesystem: extfs
 Dirs: 13
 Dirperm1 Supported: true
Logging Driver: JSON-file
Cgroup Driver: cgroupfs
Plugins:
 Volume: local
 Network: bridge host Macvlan null overlay
Swarm: active
 NodeID: rwezvezez3bg1kqg0y0f4ju22
 Is Manager: true
 ClusterID: qccn5eanox0uctyj6xtfvesy2
 Managers: 1
 Nodes: 1
 Orchestration:
  Task History Retention Limit: 5
 Raft:
  Snapshot Interval: 10000
  Number of Old Snapshots to Retain: 0
  Heartbeat Tick: 1
  Election Tick: 3
 Dispatcher:
  Heartbeat Period: 5 seconds
 CA Configuration:
  Expiry Duration: 3 months
 Node Address: 10.0.0.5
 Manager Addresses:
  10.0.0.5:2377
<Snip>
```

至此，群组已经初始化完毕，以 **node1** 作为管理节点。下一小节中，你将要把 **node2** 和 **node3** 添加成为工作节点。

### 步骤 2.2 向群组中添加工作节点

你需要在 **node2** 和 **node3** 上完成下列步骤，在完成后切换回 **node1** 节点。

现在，从 **node1** 的终端输出中复制整个 `docker swarm join ...` 命令。我们需要将这条命令粘贴到 **node2** 和 **node3** 的终端。

顺便一提，如果 `docker swarm join ...` 命令已经滚动出屏幕显示范围，可以在管理节点上运行 `docker swarm join-token worker` 命令，来重新找到它。

> 记住，这里展示的 tokens 并非你将实际用到的。要从 **node1** 的输出中复制命令。在 **node2** 和 **node3** 上，应该像下面这样：

```shell
docker swarm join \
    --token SWMTKN-1-1wxyoueqgpcrc4xk2t3ec7n1poy75g4kowmwz64p7ulqx611ih-68pazn0mj8p4p4lnuf4ctp8xy \
    10.0.0.5:2377
```

```shell
docker swarm join \
    --token SWMTKN-1-1wxyoueqgpcrc4xk2t3ec7n1poy75g4kowmwz64p7ulqx611ih-68pazn0mj8p4p4lnuf4ctp8xy \
    10.0.0.5:2377
```

当你在 **node2** 和 **node3** 上运行完这一命令之后，切换回 **node1**，然后运行 `docker node ls` 来验证是不是两个新的节点都加入到群组当中。你应当会看到共有三个节点，**node1** 是管理节点，**node2** 和 **node3** 都是工作节点。

```shell
docker node ls
```

```shell
ID                           HOSTNAME  STATUS  AVAILABILITY  MANAGER STATUS
6dlewb50pj2y66q4zi3egnwbi *  node1   Ready   Active        Leader
ym6sdzrcm08s6ohqmjx9mk3dv    node3   Ready   Active
yu3hbegvwsdpy9esh9t2lr431    node2   Ready   Active
```

 `docker node ls` 命令展示的是群组中的所有节点以及它们在群组中的角色。`*` 标志表示你发布指令的节点。

恭喜！你已经成功部署了含有一个管理节点和两个工作节点的群组了。

## 第三节：跨多主机部署应用

现在你有了一个运行中的群组，是时候部署我们非常简单的 *sleep* 应用了。你需要在 **node1** 上完成以下步骤。

### 步骤 3.1 将应用组件部署为 Docker 服务

我们的 *sleep* 应用正在互联网上蹿红（由于在 Reddit 和 HN 上广受关注）。人们真的就很喜欢它。所以，你需要为你的应用扩容以满足峰值需求。你同样需要在多主机上部署，来获得高可用性。我们用*服务*的概念来简化应用扩容，并将多个容器作为一个单独的实体来管理。

> 服务（Services）是在 Docker 1.12 出现的新概念。这一概念与集群联系紧密，并且是为长时间运行（long-running）的容器设计的。

你需要在 **node1** 上进行以下操作。

先来将 *sleep* 部署成我们 Docker 群组上的一个服务吧。

```shell
docker service create --name sleep-app Ubuntu sleep infinity
```

```
of5rxsxsmm3asx53dqcq0o29c
```

验证 `service create` 请求已经被 Swarm 管理节点接收。

```shell
docker service ls
```

```shell
ID            NAME       MODE        REPLICAS  IMAGE
of5rxsxsmm3a  sleep-app  replicated  1/1       Ubuntu:latest
```

服务的状态可能会经过多次改变，直到进入运行状态。镜像从 Docker Store 下载到群组里其他的机器上。当镜像下载完毕后，容器将在三个节点中的一个上进入运行状态。

到这一步为止，我们所做的事情似乎并没有比运行一条 `docker run ...` 命令更加复杂。我们同样还是只在一台主机上部署了单个的容器。不同之处在于，这个容器是在一个 Swarm 集群中进行调度。

好的，利用了 Docker 服务这一特性，你已经成功在新群组上部署了这个 sleep 应用。

## 第四节：应用扩容缩容

需求量爆炸！大家都爱上你的 `sleep` 应用了！是时候进行扩容了。

服务的一大优势在于，你可以进行扩容、缩容以更好地满足需求。在本节中，你将首先对服务进行扩容，然后在缩容回去。

你需要在 **node1** 上进行以下操作。

使用 `docker service update --replicas 7 sleep-app` 命令对 **sleep-app** 服务中的容器数量进行调节。其中 `replicas` 一词，描述提供同一服务的相同容器。

```shell
docker service update --replicas 7 sleep-app
```

群组管理节点将进行调度，使得集群中共有 7 个 `sleep-app` 容器。这些容器会被均匀地调度到各个群组成员上。

我们将要使用 `docker service ps sleep-app` 命令。如果你在上一步设置了 `--replicas` 之后足够快地运行这条命令，你将看到这些容器实时启动的过程。

```shell
docker service ps sleep-app
```

```shell
ID            NAME         IMAGE          NODE     DESIRED STATE  CURRENT STATE          ERROR  PORTS
7k0flfh2wpt1  sleep-app.1  Ubuntu:latest  node1  Running        Running 9 minutes ago
wol6bzq7xf0v  sleep-app.2  Ubuntu:latest  node3  Running        Running 2 minutes ago
id50tzzk1qbm  sleep-app.3  Ubuntu:latest  node2  Running        Running 2 minutes ago
ozj2itmio16q  sleep-app.4  Ubuntu:latest  node3  Running        Running 2 minutes ago
o4rk5aiely2o  sleep-app.5  Ubuntu:latest  node2  Running        Running 2 minutes ago
35t0eamu0rue  sleep-app.6  Ubuntu:latest  node2  Running        Running 2 minutes ago
44s8d59vr4a8  sleep-app.7  Ubuntu:latest  node1  Running        Running 2 minutes ago
```

注意，这里一共列出了 7 个容器。这些新容器从启动到变成如上面显示的 **RUNNING** 状态，可能需要一些时间。``NODE`` 一列向我们展示容器是跑在哪个节点上面。

要想将服务缩容回 4 个容器，只需要 `docker service update --replicas 4 sleep-app` 命令。

```shell
docker service update --replicas 4 sleep-app
```

验证容器的数量是否确实减少到了 4，我们可以用 `docker service ps sleep-app` 命令。

```shell
docker service ps sleep-app
```

```shell
ID            NAME         IMAGE          NODE     DESIRED STATE  CURRENT STATE           ERROR  PORTS
7k0flfh2wpt1  sleep-app.1  Ubuntu:latest  node1  Running        Running 13 minutes ago
wol6bzq7xf0v  sleep-app.2  Ubuntu:latest  node3  Running        Running 5 minutes ago
35t0eamu0rue  sleep-app.6  Ubuntu:latest  node2  Running        Running 5 minutes ago
44s8d59vr4a8  sleep-app.7  Ubuntu:latest  node1  Running        Running 5 minutes ago
```

你现在已经成功完成群组服务的扩容缩容啦。

## 第五节：排空节点并重新调度容器

你的 `sleep-app` 在火爆 Reddit 和 HN 之后一直保持强劲势头。现在已经攀升到应用商店的第一位了！你在节假日进行扩容，在淡季进行缩容。现在你要维护其中的一台服务器，所以你要优雅地将这台服务器移出群组，而不影响到顾客们使用服务。

首先在 **node1** 上使用 `docker node ls` 命令，再查看一下节点的状态。

```shell
docker node ls
```

```shell
ID                           HOSTNAME  STATUS  AVAILABILITY  MANAGER STATUS
6dlewb50pj2y66q4zi3egnwbi *  node1   Ready   Active        Leader
ym6sdzrcm08s6ohqmjx9mk3dv    node3   Ready   Active
yu3hbegvwsdpy9esh9t2lr431    node2   Ready   Active
```

你需要将 **node2** 移出群组进行维护。

让我们看看在 **node2** 上有哪些运行中的容器。

```shell
docker ps
```

```shell
CONTAINER ID        IMAGE                                                                            COMMAND             CREATED             STATUS              PORTS               NAMES
4e7ea1154ea4        Ubuntu@sha256:dd7808d8792c9841d0b460122f1acf0a2dd1f56404f8d1e56298048885e45535   "sleep infinity"    9 minutes ago       Up 9 minutes                            sleep-app.6.35t0eamu0rueeozz0pj2xaesi
```

你会看到有其中一个 sleep-app 的容器正在这里运行（虽然你的输出可能会稍有不同）。

现在让我们回到 **node1**，把 **node2** 移出这个服务。要做到这一点，先再运行一次 `docker node ls` 命令。

```shell
docker node ls
```

```shell
ID                           HOSTNAME  STATUS  AVAILABILITY  MANAGER STATUS
6dlewb50pj2y66q4zi3egnwbi *  node1   Ready   Active        Leader
ym6sdzrcm08s6ohqmjx9mk3dv    node3   Ready   Active
yu3hbegvwsdpy9esh9t2lr431    node2   Ready   Active
```

我们需要用到 **node2** 的 **ID**，运行 `docker node update --availability drain yournodeid`。我们使用 **node2** 的主机 **ID** 作为 `drain` 命令的输入。你需要将命令中的 yournodeid 替换成 **node2** 的实际 **ID**。

```shell
docker node update --availability drain yournodeid
```

检查节点状态

```shell
docker node ls
```

```shell
ID                           HOSTNAME  STATUS  AVAILABILITY  MANAGER STATUS
6dlewb50pj2y66q4zi3egnwbi *  node1   Ready   Active        Leader
ym6sdzrcm08s6ohqmjx9mk3dv    node3   Ready   Active
yu3hbegvwsdpy9esh9t2lr431    node2   Ready   Drain
```

现在 **node2** 就进入了 `Drain` 状态。

切换到 **node2**，使用 `docker ps` 看看运行中的容器。

```shell
docker ps
```

```
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
```

在 **node2** 上已经没有容器在运行了。

最后，我们回到 **node1**，看看容器是否被重新调度了。理论上你将看到所有的四个容器都在剩下的两个节点上运行。

```shell
docker service ps sleep-app
```

```shell
ID            NAME             IMAGE          NODE     DESIRED STATE  CURRENT STATE           ERROR  PORTS
7k0flfh2wpt1  sleep-app.1      Ubuntu:latest  node1  Running        Running 25 minutes ago
wol6bzq7xf0v  sleep-app.2      Ubuntu:latest  node3  Running        Running 18 minutes ago
s3548wki7rlk  sleep-app.6      Ubuntu:latest  node3  Running        Running 3 minutes ago
35t0eamu0rue   \_ sleep-app.6  Ubuntu:latest  node2  Shutdown       Shutdown 3 minutes ago
44s8d59vr4a8  sleep-app.7      Ubuntu:latest  node1  Running        Running 18 minutes ago
```

## 清理工作

在 **node1** 上执行 `docker service rm sleep-app` 命令，删掉名为 *sleep-app* 的服务。

```shell
docker service rm sleep-app
```

在 **node1** 上执行 `docker ps`，获取运行中的容器列表。

```shell
docker ps
```

```shell
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
044bea1c2277        Ubuntu              "sleep infinity"    17 minutes ago      17 minutes ag                           distracted_mayer
```

你可以在 **node1** 上用 `docker kill <CONTAINER ID>` 命令，清理掉我们最开始启动的那个 sleep 容器。

```shell
docker kill yourcontainerid
```

最后，让我们把 node1，node2，node3 从群组中移除。我们可以用 `docker swarm leave --force` 命令做到这一点。

先在 **node1** 上运行 `docker swarm leave --force`。

```shell
docker swarm leave --force
```

然后在 **node2** 上运行 `docker swarm leave --force`。

```shell
docker swarm leave --force
```

最后在 **node3** 上运行 `docker swarm leave --force`。

```shell
docker swarm leave --force
```

恭喜你，完成本次的练习！你现在应该学会了怎样创建群组、将应用部署成服务以及对每个服务进行扩容缩容。

---

via: https://training.play-with-docker.com/orchestration-hol/

作者：[Play with Docker classroom](https://training.play-with-docker.com)
译者：[Mockery-Li](https://github.com/Mockery-Li)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
