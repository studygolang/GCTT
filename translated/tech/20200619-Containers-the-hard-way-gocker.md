容器很受欢迎，但是被误解了。 容器已成为应用程序在服务器上打包和运行的默认方式，最初是由 Docker 普及的。现在，Docker 本身被误解了。它是一个公司的名字和一条命令（更确切地说是一组命令），使你容易地管理容器（创建，运行，删除，连接网络）。但是容器本身是由一组操作系统原语创建的。在本文中，我们将关注 Linux 操作系统上的容器，就像 [Windows 上的容器](https://docs.microsoft.com/en-us/virtualization/windowscontainers/about/) 根本不存在一样。

Linux 下没有创建容器的单个系统调用。它们是利用 Linux namespaces 和 control groups 或 cgroups 创建的松散结构。

# Gocker 是什么？

[Gocker](https://github.com/shuveb/containers-the-hard-way) 是一个用 Go 语言从头实现 Docker 核心功能的项目。主要目的是让你了解容器在 Linux 系统调用级别上是如何工作的。Gocker 允许你创建容器，管理容器镜像，执行现有容器中的进程等等。

# Gocker 的功能

Gocker 可以模拟 Docker 的核心，让你管理 Docker 镜像（从 Docker Hub 获取），运行容器，列出正在运行的容器或在已运行的容器中执行进程：

- 在容器中运行进程
  - gocker run <--cpus=cpus-max> <--mem=mem-max> <--pids=pids-max> <image[:tag]> </path/to/command>
- 列出正在运行的容器
  - gocker ps
- 在运行的容器中执行进程
  - gocker exec </path/to/command>
- 列出本地可用的镜像
  - gocker images
- 删除本地可用的镜像
  - gocker rmi

## 其他功能
- Gocker 使用 Ovelay 文件系统快速创建容器而无需复制整个文件系统，同时还可以在多容器实例间共享容器镜像。
- Gocker 容器拥有自己的网络命名空间，并且能够访问 Internet。请参阅下面的限制。
- 你可以控制系统资源，如 CPU 百分比，RAM 数量和进程数。 Gocker 通过利用 cgroups 来实现这一点。

# Gocker 容器隔离
用 Gocker 创建的容器拥有自己的以下命名空间（参见 run.go 和 network.go）：

- File system (via chroot)
- PID
- IPC
- UTS (hostname)
- Mount
- Network

在创建用于限制以下内容的 cgroups 时，除非为 gocker run 命令指定了 --mem，--cpus 或 --pids 选项，否则容器将使用无限制的资源。这些标志分别限制了容器可以使用的最大 RAM，CPU 核数和 PID。

- CPU 核数
- RAM
- PID 个数 (限制进程)

# Namespaces 基础

所有 Linux 计算机在启动时都是 “default” 命名空间的一部分。在计算机上创建的进程也继承默认命名空间。换句话说，进程可以查看正在运行的其他进程，列出网络接口，列出挂载点，列出名为 IPC 的对象或权限允许的文件，因为所有对象也都存在于默认命名空间中。例如，创建一个进程时，我们可以告诉 Linux 为我们创建一个新的 PID 命名空间，在这种情况下，新进程及其任何后代形成一个新的层次结构或 PID，而新创建的初始进程 PID 为 1， 就像一个 Linux 机器上特殊的 init 进程一样。假设使用新的 PID 命名空间创建了一个名为“ new_child”的进程。当该进程或其后代使用诸如 getpid() 或 getppid() 之类的系统调用时，它们将看到来自新命名空间的 PID。例如，在新创建的 PID 命名空间中，这两个系统调用的结果都是 1。而当你从默认命名空间查看 new_child 的 PID 时，当然不会为它分配 1，也就是默认命名空间中的 init。将根据在分配时间前后分配的一系列 PID 进程，为它分配更多的 PID。

Linux 操作系统提供了在创建进程时或与之关联的正在运行的进程创建新命名空间的方法。所有命名空间，无论何种类型，都被分配了内部 ID。命名空间是一种内核对象。一个进程只能属于一个命名空间。例如，假设进程新的子 PID 命名空间被设置为内部 ID 为 0x87654321 的命名空间，它不能属于另一个 PID 命名空间。但是可能存在其他属于同一 PID 命名空间 0x87654321 的其他进程。同样，new_child 的后代将自动属于相同的 PID 命名空间。命名空间是继承的。

你可以使用 lsns 列出计算机中的所有命名空间。即使你的计算机上没有运行任何容器，也可能会看到与各种命名空间相关的其他进程。这表明命名空间并不仅仅是在容器的上下文中使用，它们可以在任何地方使用。它们提供隔离，是一项强大的安全功能。在现代 Linux 系统上，你会看到 init，systemd，多个系统守护进程，Chrome，Slack，当然还有使用各种命名空间的 Docker 容器。让我们看一下我机器上 lsns 的输出：

```
        NS TYPE   NPROCS   PID USER             COMMAND
4026532281 mnt         1   313 root             /usr/lib/systemd/systemd-udevd
4026532282 uts         1   313 root             /usr/lib/systemd/systemd-udevd
4026532313 mnt         1   483 systemd-timesync /usr/lib/systemd/systemd-timesyncd
4026532332 uts         1   483 systemd-timesync /usr/lib/systemd/systemd-timesyncd
4026532334 mnt         1   502 root             /usr/bin/NetworkManager --no-daemon
4026532335 mnt         1   503 root             /usr/lib/systemd/systemd-logind
4026532336 uts         1   503 root             /usr/lib/systemd/systemd-logind
4026532341 pid         1  1943 shuveb           /opt/google/chrome/nacl_helper
4026532343 pid         2  1941 shuveb           /opt/google/chrome/chrome --type=zygote
4026532345 net        50  1941 shuveb           /opt/google/chrome/chrome --type=zygote
4026532449 mnt         1   547 root             /usr/lib/boltd
4026532489 mnt         1   580 root             /usr/lib/bluetooth/bluetoothd
4026532579 net         1  1943 shuveb           /opt/google/chrome/nacl_helper
4026532661 mnt         1   766 root             /usr/lib/upowerd
4026532664 user        1   766 root             /usr/lib/upowerd
4026532665 pid         1  2521 shuveb           /opt/google/chrome/chrome --type=renderer
4026532667 net         1   836 rtkit            /usr/lib/rtkit-daemon
4026532753 mnt         1   943 colord           /usr/lib/colord
4026532769 user        1  1943 shuveb           /opt/google/chrome/nacl_helper
4026532770 user       50  1941 shuveb           /opt/google/chrome/chrome --type=zygote
4026532771 pid         1  2010 shuveb           /opt/google/chrome/chrome --type=renderer
4026532772 pid         1  2765 shuveb           /opt/google/chrome/chrome --type=renderer
4026531835 cgroup    294     1 root             /sbin/init
4026531836 pid       237     1 root             /sbin/init
4026531837 user      238     1 root             /sbin/init
4026531838 uts       289     1 root             /sbin/init
4026531839 ipc       292     1 root             /sbin/init
4026531840 mnt       283     1 root             /sbin/init
4026531992 net       236     1 root             /sbin/init
4026532912 pid         2  3249 shuveb           /usr/lib/slack/slack --type=zygote
4026532914 net         2  3249 shuveb           /usr/lib/slack/slack --type=zygote
4026533003 user        2  3249 shuveb           /usr/lib/slack/slack --type=zygote
```

即使你没有显式创建命名空间，进程也将成为默认命名空间的一部分。所有命名空间的详细信息都记录在 /proc 文件系统中。你可以通过输入 ls -l /proc/self/ns/ 来查看你的 shell 进程所属的命名空间。下面是我的执行结果。另外，这些大多是从 init 继承的：

```
➜  ~ ls -l /proc/self/ns
total 0
lrwxrwxrwx 1 shuveb shuveb 0 Jun 13 11:44 cgroup -> 'cgroup:[4026531835]'
lrwxrwxrwx 1 shuveb shuveb 0 Jun 13 11:44 ipc -> 'ipc:[4026531839]'
lrwxrwxrwx 1 shuveb shuveb 0 Jun 13 11:44 mnt -> 'mnt:[4026531840]'
lrwxrwxrwx 1 shuveb shuveb 0 Jun 13 11:44 net -> 'net:[4026531992]'
lrwxrwxrwx 1 shuveb shuveb 0 Jun 13 11:44 pid -> 'pid:[4026531836]'
lrwxrwxrwx 1 shuveb shuveb 0 Jun 13 11:44 pid_for_children -> 'pid:[4026531836]'
lrwxrwxrwx 1 shuveb shuveb 0 Jun 13 11:44 user -> 'user:[4026531837]'
lrwxrwxrwx 1 shuveb shuveb 0 Jun 13 11:44 uts -> 'uts:[4026531838]'
```

# 没有容器的命名空间

从 lsns 的输出中我们看到容器并不是唯一使用命名空间的对象。为此，我们来创建一个具有自己的 PID 命名空间的 shell 实例。我们将使用 unshare 来做到这一点。“unshare”这个名字很明显。还有一个 [同名的 Linux 系统调用](https://man7.org/linux/man-pages/man2/unshare.2.html)，用来取消共享默认命名空间，使调用进程加入一个新创建的命名空间。

```
➜  ~ sudo unshare --fork --pid --mount-proc /bin/bash
[root@kodai shuveb]# ps aux
USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root           1  0.5  0.0   8296  4944 pts/1    S    08:59   0:00 /bin/bash
root           2  0.0  0.0   8816  3336 pts/1    R+   08:59   0:00 ps aux
[root@kodai shuveb]#
```

上面的命令中 unshare 创建一个新进程，调用 unshare() 系统调用来创建一个新的 PID 命名空间，然后在其中执行 /bin/bash。我们还告诉 unshare 在新进程中挂载 proc 文件系统，这是 ps 获取信息的地方。从 ps 命令的输出中，确实可以看到该 shell 有一个新的 PID 命名空间，PID 为 1，并且由于 ps 是由具有新 PID 命名空间的 shell 程序启动的，因此它继承了命名空间并获得 PID 为 2。作为练习，你可以找出此容器中运行的 shell 进程在主机上具有哪些 PID。

# 命名空间的类型

了解 PID 命名空间后，让我们尝试理解其他命名空间以及它们的含义。[命名空间手册页](https://man7.org/linux/man-pages/man7/namespaces.7.html) 讨论了 8 种不同的命名空间。以下是带有简短说明的各种类型，以及指向相关手册页的链接：

Namespace | Flag |Isolates
--|--|--
[Cgroup](https://man7.org/linux/man-pages/man7/cgroup_namespaces.7.html) | CLONE_NEWCGROUP | Cgroup root directory
[IPC](https://man7.org/linux/man-pages/man7/ipc_namespaces.7.html) | CLONE_NEWIPC | System V IPC,POSIX message queues
[Network](https://man7.org/linux/man-pages/man7/network_namespaces.7.html) | CLONE_NEWNET | Network devices,stacks, ports, etc.
[Mount](https://man7.org/linux/man-pages/man7/mount_namespaces.7.html) | CLONE_NEWNS | Mount points
[PID](https://man7.org/linux/man-pages/man7/pid_namespaces.7.html) | CLONE_NEWPID | Process IDs
[Time](https://man7.org/linux/man-pages/man7/time_namespaces.7.html) | CLONE_NEWTIME | Boot and monotonic clocks
[User](https://man7.org/linux/man-pages/man7/user_namespaces.7.html) | CLONE_NEWUSER | User and group IDs
[UTS](https://man7.org/linux/man-pages/man7/uts_namespaces.7.html) | CLONE_NEWUTS | Hostname and NIS domain name

你可以想象能够用这些命名空间为新的或现有的进程做什么。当它们在同一台计算机上运行时，你几乎可以将它们隔离开就像它们在单独的虚拟机中运行一样。你可以将多个进程隔离在各自的命名空间中，并在同一主机内核上运行。这比运行多个虚拟机要有效得多。

# 创建新的命名空间或加入现有的命名空间

默认情况下，当使用 fork() 创建进程时，子进程继承调用 fork() 的进程的命名空间。如果希望创建的新进程成为一组新命名空间的一部分该怎么办？如你所见，fork() 没有参数，并且不允许我们在创建子进程之前对其进行控制。但是可以通过 clone() 系统调用施加这种控制，从而可以非常精细地控制它创建的新进程。

## 关于 clone() 的附注

Linux 下虽然有不同的系统调用，例如 fork()，vfork() 和 clone() 来创建新进程，但是内核中的 fork() 和 vfork() 只是使用不同的参数调用 clone()。围绕此的内核源码（为了更好的说明，我进行了一些编辑）非常容易理解。在文件 [kernel/fork.c](https://elixir.bootlin.com/linux/v5.7.2/source/kernel/fork.c#L2521) 中可以看到以下内容：

```c
SYSCALL_DEFINE0(fork)
{
  struct kernel_clone_args args = {
    .exit_signal = SIGCHLD,
  };

  return _do_fork(&args);
}

SYSCALL_DEFINE0(vfork)
{
  struct kernel_clone_args args = {
    .flags    = CLONE_VFORK | CLONE_VM,
    .exit_signal  = SIGCHLD,
  };

  return _do_fork(&args);
}


SYSCALL_DEFINE5(clone, unsigned long, clone_flags, unsigned long, newsp,
     int __user *, parent_tidptr,
     int __user *, child_tidptr,
     unsigned long, tls)
{
  struct kernel_clone_args args = {
    .flags    = (lower_32_bits(clone_flags) & ~CSIGNAL),
    .pidfd    = parent_tidptr,
    .child_tid  = child_tidptr,
    .parent_tid  = parent_tidptr,
    .exit_signal  = (lower_32_bits(clone_flags) & CSIGNAL),
    .stack    = newsp,
    .tls    = tls,
  };

  if (!legacy_clone_args_valid(&args))
    return -EINVAL;

  return _do_fork(&args);
}
```

如你所见，这三个系统调用只是使用不同的参数调用 _do_fork()，_do_fork() 实现创建新进程的逻辑。

## 使用 clone() 创建具有新命名空间的进程

Gocker 通过 Go 的 exec 包使用 clone() 系统调用。在处理与运行容器有关的内容的 [run.go](https://github.com/shuveb/containers-the-hard-way/blob/master/run.go) 中，可以看到以下内容：

```go
cmd = exec.Command("/proc/self/exe", args...)
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
cmd.SysProcAttr = &syscall.SysProcAttr{
    Cloneflags: syscall.CLONE_NEWPID |
      syscall.CLONE_NEWNS |
      syscall.CLONE_NEWUTS |
      syscall.CLONE_NEWIPC,
}
doOrDie(cmd.Run())
```

在 syscall.SysProcAttr 中，我们可以传入 Cloneflags，然后将其传递给对 clone() 系统调用的调用。聪明的读者会注意到我们在这里没有设置单独的网络命名空间。在 Gocker 中我们设置了一个虚拟以太网接口，将其添加到新的网络命名空间，并让容器使用不同的 Linux 系统调用加入该命名空间。我们将在后面讨论这个问题。

## 使用 unshare() 创建和加入新的命名空间

如果要为现有进程创建新的命名空间，不必使用 clone() 创建新的子进程，Linux 提供了 [unshare()](https://man7.org/linux/man-pages/man2/unshare.2.html) 系统调用。

## 加入其他进程所属的命名空间

为了加入文件引用的命名空间或加入其他进程所属的命名空间，Linux 允许使用 [setns()](https://man7.org/linux/man-pages/man2/setns.2.html) 系统调用。 我们将很快看到，这非常有用。

# Gocker 是如何创建容器的

由于 Gocker 的主要目的是帮助理解 Linux 容器，因此保留了一些来自 Gocker 的日志消息。从这个意义上讲，它比运行 Docker 更为冗长。让我们看一下日志，以指导我们执行程序。然后我们可以进行深入分析，看看实际是如何运作的：

```
➜  sudo ./gocker run alpine /bin/sh
2020/06/13 12:37:53 Cmd args: [./gocker run alpine /bin/sh]
2020/06/13 12:37:53 New container ID: 33c20f9ee600
2020/06/13 12:37:53 Image already exists. Not downloading.
2020/06/13 12:37:53 Image to overlay mount: a24bb4013296
2020/06/13 12:37:53 Cmd args: [/proc/self/exe setup-netns 33c20f9ee600]
2020/06/13 12:37:53 Cmd args: [/proc/self/exe setup-veth 33c20f9ee600]
2020/06/13 12:37:53 Cmd args: [/proc/self/exe child-mode --img=a24bb4013296 33c20f9ee600 /bin/sh]
/ #
```

这里我们要求 Gocker 从 Alpine Linux 镜像运行 shell。稍后我们将了解如何管理镜像。现在请注意以“ Cmd args：”开头的日志行。这行表示产生了一个新进程。第一行日志向我们显示运行 Gocker 命令后运行 shell 启动的过程。但是最后我们看到又启动了三个进程。最后一个带有第二个参数为“child-mode”的参数是在 Alpine Linux 镜像内部执行 shell 的程序 /bin/sh。在此之前，我们看到另外两个分别带有参数“setup-netns”和“setup-veth”的进程。这些进程设置了新的网络命名空间，并设置了虚拟以太网设备对的容器端，对使容器与外界通信。

由于各种原因，Go 语言不直接支持 fork() 系统调用。我们通过创建一个新进程来解决此限制，但是要在其中再次执行当前程序。 /proc/self/exe 指向当前正在运行的可执行文件的路径。我们根据传递不同的命令行参数来调用适当的函数（当 fork() 在子进程中返回时将调用该函数）。

## 源代码的组织

Gocker 源代码通过命令（如参数）组织在文件中。例如，主要服务于 gocker run 命令行参数的函数位于 run.go 文件中。类似地，gocker exec 主要需要的功能在 exec.go 文件中。这并不意味着这些文件是独立的。他们从其他文件中自由调用函数。还有一些文件可以实现常见功能，例如 cgroups.go 和 utils.go。

## 运行容器

在 [main.go](https://github.com/shuveb/containers-the-hard-way/blob/master/main.go) 中，你可以看到是否运行了 Gocker 命令，我们检查以确保 gocker0 桥接器已启动并正在运行。否则我们通过调用完成工作的 setupGockerBridge() 来启动它。最后，我们调用函数 initContainer()，该函数在 run.go 中实现。让我们仔细看看该函数：

```go
func initContainer(mem int, swap int, pids int, cpus float64,
                                src string, args []string) {
  containerID := createContainerID()
  log.Printf("New container ID: %s\n", containerID)
  imageShaHex := downloadImageIfRequired(src)
  log.Printf("Image to overlay mount: %s\n", imageShaHex)
  createContainerDirectories(containerID)
  mountOverlayFileSystem(containerID, imageShaHex)
  if err := setupVirtualEthOnHost(containerID); err != nil {
    log.Fatalf("Unable to setup Veth0 on host: %v", err)
  }
  prepareAndExecuteContainer(mem, swap, pids, cpus, containerID,
                                imageShaHex, args)
  log.Printf("Container done.\n")
  unmountNetworkNamespace(containerID)
  unmountContainerFs(containerID)
  removeCGroups(containerID)
  os.RemoveAll(getGockerContainersPath() + "/" + containerID)
}
```

首先，我们通过调用 createContainerID() 创建唯一的容器 ID。然后，我们调用 downloadImageIfRequired()，以便可以从 Docker Hub 下载容器镜像（如果本地尚不可用）。 Gocker 使用 /var/run/gocker/containers 中的子目录来挂载容器根文件系统。createContainerDirectories() 会解决这个问题。mountOverlayFileSystem() 知道如何处理多层 Docker 镜像，并在 /var/run/gocker/containers/<container-id>/fs/mnt 上为可用镜像安装合并的文件系统。尽管这看起来令人生畏，但如果阅读源代码并不难理解。覆盖文件系统允许创建一个堆叠的文件系统，其中较低的层（在这种情况下是 Docker 根文件系统）是只读的，而任何更改都将保存到“upperdir” 而无需更改较低层中的任何文件。这允许多容器共享一个 Docker 镜像。当我们在虚拟机上下文中说“镜像”时，它通常是指磁盘镜像。但是在这里，它只是一个目录或一组目录（奇特的名字：layers），带有构成 Docker“镜像”根文件系统的文件，这些文件可以使用 Overlay 文件系统挂载它，为新容器创建根文件系统。

接下来，我们创建一个虚拟的以太网配对设备，它非常类似于调用 setupVirtualEthOnHost() 的管道。它们采用名称 veth0_<container-id> 和 veth1_<container-id> 的形式。我们将一对中的 veth0 部分连接到主机上的网桥 gocker0。稍后我们将在容器内部使用该对的 veth1 部分。这对就像管道一样，是从具有自己的网络命令空间的容器内部进行网络通信的密钥。随后，我们将介绍如何在容器内设置 veth1 部件。

最后，调用 prepareAndExecuteContainer()，它实际上在容器中执行该过程。当此函数返回时，容器已完成执行。最后，我们然后进行一些清理并退出。让我们看看 prepareAndExecuteContainer() 的作用。它实际上创建了我们看到日志的 3 个进程，并使用参数 setup-netns，setup-veth 和 child-mode 运行相同的 gocker 二进制文件。

## 设置可在容器内工作的网络

设置新的网络命名空间非常容易，只需将 CLONE_NEWNET 包括在传递给 clone() 系统调用的标志位掩码中即可。棘手的是确保容器内部可以具有网络接口，通过该接口可以与外部进行通信。在 Gocker 中，我们创建的第一个新命名空间是网络命名空间。当使用 setup-ns 和 setup-veth 参数调用 gocker 时会发生这种情况。首先，我们设置一个新的网络命名空间。setns() 系统调用可以将调用进程的命名空间设置为文件描述符所引用的命名空间，该文件描述符指向 /proc/<pid>/ns 中的文件，该文件列出了进程所属的所有命名空间。让我们看一下 setupNewNetworkNamespace() 函数，该函数是通过调用 setup-netns 参数调用 gocker 的结果而被调用的。

```go
func setupNewNetworkNamespace(containerID string) {
  _ = createDirsIfDontExist([]string{getGockerNetNsPath()})
  nsMount := getGockerNetNsPath() + "/" + containerID
  if _, err := syscall.Open(nsMount,
                syscall.O_RDONLY|syscall.O_CREAT|syscall.O_EXCL,
                0644); err != nil {
    log.Fatalf("Unable to open bind mount file: :%v\n", err)
  }

  fd, err := syscall.Open("/proc/self/ns/net", syscall.O_RDONLY, 0)
  defer syscall.Close(fd)
  if err != nil {
    log.Fatalf("Unable to open: %v\n", err)
  }

  if err := syscall.Unshare(syscall.CLONE_NEWNET); err != nil {
    log.Fatalf("Unshare system call failed: %v\n", err)
  }
  if err := syscall.Mount("/proc/self/ns/net", nsMount,
                                "bind", syscall.MS_BIND, ""); err != nil {
    log.Fatalf("Mount system call failed: %v\n", err)
  }
  if err := unix.Setns(fd, syscall.CLONE_NEWNET); err != nil {
    log.Fatalf("Setns system call failed: %v\n", err)
  }
}
```

每当 Linux 内核中的最后一个进程终止时，它都会自动删除该命名空间。但是，有一种技术可以通过绑定安装来保留命名空间，即使其中没有任何进程。我们在 setupNewNetworkNamespace() 函数中使用此技术。我们首先打开进程的网络命名空间文件，该文件位于 /proc/self/ns/net 中。然后，我们使用 CLONE_NEWNET 参数调用 unshare() 系统调用。这将调用过程与其所属的命名空间解除关联，创建一个新的新网络命名空间，并将其设置为该进程的网络命名空间。然后，我们将此进程的网络命名空间专用文件的安装挂载绑定到已知文件名，即 /var/run/gocker/net-ns/<container-id>。该文件可随时用于引用该网络命名空间。现在，我们可以退出此进程，但是由于此进程的新网络命名空间已绑定安装到新文件上，因此内核将保留此命名空间。

接下来，使用 setup-veth 参数调用 gocker。 这将调用函数 setupContainerNetworkInterfaceStep1() 和 setupContainerNetworkInterfaceStep2()。在第一个函数中，我们查找 veth1_<container-id> 接口，并将其命名空间设置为在上一步中创建的新网络命名空间。现在，该接口将不再在主机上可见。 但问题是：由于它与 veth0_<container-id> 接口配对，该接口在主机上仍然可见，因此，加入此网络命名空间的任何进程都可以与主机进行通信。 第二个功能将 IP 地址添加到网络接口，并将 gocker0 网桥设置为其默认网关设备。

现在，主机上有一个网络接口，而新的网络命名空间上有一个可以相互通信的接口。而且由于该网络命名空间可以由文件引用，因此我们可以随时使用 setns() 系统调用打开该文件并加入该网络命名空间。而且，这正是我们要做的。

此后，prepareAndExecuteContainer() 调用将设置一个新进程，该进程使用 child-mode 参数运行 gocker。 这是最后的进程，将产生我们要在容器中运行的命令。让我们看一下运行 child-mode 的进程的新命名空间。我们之前已经看过了这段代码：

```go
cmd = exec.Command("/proc/self/exe", args...)
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
cmd.SysProcAttr = &syscall.SysProcAttr{
    Cloneflags: syscall.CLONE_NEWPID |
      syscall.CLONE_NEWNS |
      syscall.CLONE_NEWUTS |
      syscall.CLONE_NEWIPC,
}
doOrDie(cmd.Run())
```

在这里，我们设置新的 PID，mount，UTS 和 IPC 命名空间。请记住，我们有一个文件可以引用的新网络命名空间。我们只需要加入它就会很快完成。child-mode 进程将调用函数 execContainerCommand()。以下是代码：

```go
func execContainerCommand(mem int, swap int, pids int, cpus float64,
  containerID string, imageShaHex string, args []string) {
  mntPath := getContainerFSHome(containerID) + "/mnt"
  cmd := exec.Command(args[0], args[1:]...)
  cmd.Stdin = os.Stdin
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr

  imgConfig := parseContainerConfig(imageShaHex)
  doOrDieWithMsg(syscall.Sethostname([]byte(containerID)), "Unable to set hostname")
  doOrDieWithMsg(joinContainerNetworkNamespace(containerID), "Unable to join container network namespace")
  createCGroups(containerID, true)
  configureCGroups(containerID, mem, swap, pids, cpus)
  doOrDieWithMsg(copyNameserverConfig(containerID), "Unable to copy resolve.conf")
  doOrDieWithMsg(syscall.Chroot(mntPath), "Unable to chroot")
  doOrDieWithMsg(os.Chdir("/"), "Unable to change directory")
  createDirsIfDontExist([]string{"/proc", "/sys"})
  doOrDieWithMsg(syscall.Mount("proc", "/proc", "proc", 0, ""), "Unable to mount proc")
  doOrDieWithMsg(syscall.Mount("tmpfs", "/tmp", "tmpfs", 0, ""), "Unable to mount tmpfs")
  doOrDieWithMsg(syscall.Mount("tmpfs", "/dev", "tmpfs", 0, ""), "Unable to mount tmpfs on /dev")
  createDirsIfDontExist([]string{"/dev/pts"})
  doOrDieWithMsg(syscall.Mount("devpts", "/dev/pts", "devpts", 0, ""), "Unable to mount devpts")
  doOrDieWithMsg(syscall.Mount("sysfs", "/sys", "sysfs", 0, ""), "Unable to mount sysfs")
  setupLocalInterface()
  cmd.Env = imgConfig.Config.Env
  cmd.Run()
  doOrDie(syscall.Unmount("/dev/pts", 0))
  doOrDie(syscall.Unmount("/dev", 0))
  doOrDie(syscall.Unmount("/sys", 0))
  doOrDie(syscall.Unmount("/proc", 0))
  doOrDie(syscall.Unmount("/tmp", 0))
}
```

在这里，我们将容器的主机名设置为容器 ID，加入我们先前创建的新网络命名空间，创建允许我们控制 CPU，PID 和 RAM 使用情况的 Linux 控制组，加入这些 Cgroup，然后复制主机的 DNS 解析文件进入容器的文件系统，对已安装的 Overlay 文件系统执行 chroot()，安装所需的文件系统，使容器能够平稳运行，设置本地网络接口，根据容器镜像的建议设置环境变量并最终运行用户希望我们运行的命令。现在，此命令将在一组新的命名空间中运行，从而使它几乎完全与主机隔离。

# 限制容器资源

除了使用命名空间实现隔离之外，容器还有另一个重要特征：限制容器消耗的资源量的能力。Linux 下的 Cgroup 很简单，通过它我们能够做到这一点。虽然命名空间是通过诸如 unshare()，setns() 和 clone() 之类的系统调用实现的，但 Cgroup 是通过创建目录并将文件写入到虚拟文件系统（位于 /sys/fs/cgroup 下）来管理的。在 Cgroups 虚拟文件系统层次结构中，每个容器创建了 3 个目录：
- /sys/fs/cgroup/pids/gocker/<container-id>
- /sys/fs/cgroup/cpu/gocker/<container-id>
- /sys/fs/cgroup/mem/gocker/<container-id>

对于每个创建的目录，内核会添加各种文件来自动配置 cgroup。

这是我们配置容器的方式：
- 当容器启动时，我们创建 3 个目录，每个目录分别对应我们关心的三个 cgroup：CPU，PID 和 Memory。

- 然后，我们通过写入该目录内的文件来设置 cgroup 的限制。例如，要设置容器中允许的最大 PID 数量，我们将该最大数量写入 /sys/fs/cgroup/pids/gocker/<cont-id>/pids.max，这将配置此 Cgroup。

- 现在，我们可以通过将其 PID 添加到 /sys/fs/cgroup/pids/gocker/<cont-id>/cgroup.procs 中来添加需要由该 Cgroup 控制的进程。

这就是全部。一旦添加了要由 Cgroup 控制的进程，内核就会将所有进程后代的 PID 自动添加到适当的 Cgroup 的 cgroup.procs 文件中。由于我们在容器中启动一个添加到所有 3 个 Cgroup 的进程，并且该进程是容器启动其他进程的通常方式，因此它们也继承了所有限制。

## 限制 CPU

让我们尝试将容器可以使用的 CPU 限制为主机系统 1 个 CPU 内核的 20％。让我们开始一个受此限制的容器，安装 Python 并运行一个 while 循环。我们通过向 gocker 传递 --cpu = 0.2 标志来实现：

```
sudo ./gocker run --cpus=0.2 alpine /bin/sh
2020/06/13 18:14:09 Cmd args: [./gocker run --cpus=0.2 alpine /bin/sh]
2020/06/13 18:14:09 New container ID: d87d44b4d823
2020/06/13 18:14:09 Image already exists. Not downloading.
2020/06/13 18:14:09 Image to overlay mount: a24bb4013296
2020/06/13 18:14:09 Cmd args: [/proc/self/exe setup-netns d87d44b4d823]
2020/06/13 18:14:09 Cmd args: [/proc/self/exe setup-veth d87d44b4d823]
2020/06/13 18:14:09 Cmd args: [/proc/self/exe child-mode --cpus=0.2 --img=a24bb4013296 d87d44b4d823 /bin/sh]
/ # apk add python3
fetch http://dl-cdn.alpinelinux.org/alpine/v3.12/main/x86_64/APKINDEX.tar.gz
fetch http://dl-cdn.alpinelinux.org/alpine/v3.12/community/x86_64/APKINDEX.tar.gz
(1/10) Installing libbz2 (1.0.8-r1)
(2/10) Installing expat (2.2.9-r1)
(3/10) Installing libffi (3.3-r2)
(4/10) Installing gdbm (1.13-r1)
(5/10) Installing xz-libs (5.2.5-r0)
(6/10) Installing ncurses-terminfo-base (6.2_p20200523-r0)
(7/10) Installing ncurses-libs (6.2_p20200523-r0)
(8/10) Installing readline (8.0.4-r0)
(9/10) Installing sqlite-libs (3.32.1-r0)
(10/10) Installing python3 (3.8.3-r0)
Executing busybox-1.31.1-r16.trigger
OK: 53 MiB in 24 packages
/ # python3
Python 3.8.3 (default, May 15 2020, 01:53:50)
[GCC 9.3.0] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> while True:
...     pass
...
```

我们在主机运行 top，查看在容器内部运行的 python 进程占用了多少 CPU。

![Top_command_cgroups](https://raw.githubusercontent.com/alandtsang/gctt-images2/master/20200619-Containers-the-hard-way-Gocker-A-mini-Docker-written-in-Go/Top_command_cgroups.png)
Cgroup 限制 CPU 为 20%

从另一个终端，让我们使用 gocker exec 命令在同一容器内启动另一个 python 进程，并在其中运行 while 循环。

```
➜  sudo ./gocker ps
2020/06/13 18:21:10 Cmd args: [./gocker ps]
CONTAINER ID  IMAGE    COMMAND
d87d44b4d823  alpine:latest  /usr/bin/python3.8
➜  sudo ./gocker exec d87d44b4d823 /bin/sh
2020/06/13 18:21:24 Cmd args: [./gocker exec d87d44b4d823 /bin/sh]
/ # python3
Python 3.8.3 (default, May 15 2020, 01:53:50)
[GCC 9.3.0] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> while True:
...     pass
...
```

现在有 2 个 python 进程，如果不受 Cgroup 的限制，它们将消耗 2 个完整的 CPU 核数。现在，让我们看一下主机上 top 命令的输出：

![Top_command_cgroups_2](https://raw.githubusercontent.com/alandtsang/gctt-images2/master/20200619-Containers-the-hard-way-Gocker-A-mini-Docker-written-in-Go/Top_command_cgroups_2.png)
Cgroup 限制 2 个进程的 CPU 为 20%

从主机 top 命令的输出中可以看到，这两个 python 进程都运行循环，每个进程的 CPU 限制为 10％。容器的 20％ CPU 配额由调度程序公平分配给容器中的 2 个进程。请注意，也可以指定一个以上 CPU 核数的余量。例如，如果要允许一个容器最大使用 2 个半核，请在标志中将其指定为 --cpu = 2.5。

## 限制 PID

在新的 PID 命名空间中运行 shell 程序的容器似乎消耗 7 个 PID。这意味着，如果启动的 PID 最高限制为 7，你将无法在 shell 上启动其他进程。让我们对此进行测试。[虽然容器中只有 2 个处于运行状态的进程，但我不确定为什么要消耗 7 个 PID。这需要进一步研究。]

```
➜  sudo ./gocker run --pids=7 alpine /bin/sh
[sudo] password for shuveb:
2020/06/13 18:28:00 Cmd args: [./gocker run --pids=7 alpine /bin/sh]
2020/06/13 18:28:00 New container ID: 920a577165ef
2020/06/13 18:28:00 Image already exists. Not downloading.
2020/06/13 18:28:00 Image to overlay mount: a24bb4013296
2020/06/13 18:28:00 Cmd args: [/proc/self/exe setup-netns 920a577165ef]
2020/06/13 18:28:00 Cmd args: [/proc/self/exe setup-veth 920a577165ef]
2020/06/13 18:28:00 Cmd args: [/proc/self/exe child-mode --pids=7 --img=a24bb4013296 920a577165ef /bin/sh]
/ # ls -l
/bin/sh: can't fork: Resource temporarily unavailable
/ #
```

## 限制 RAM

让我们开启一个新容器，将最大允许内存设置为 128M。现在，我们将在其中安装 python，并在其中分配大量 RAM。 这将触发内核的内存不足（OOM）killer，来杀死我们的 python 进程。让我们看看实际情况：

```
➜ sudo ./gocker run --mem=128 --swap=0 alpine /bin/sh
2020/06/13 18:30:30 Cmd args: [./gocker run --mem=128 --swap=0 alpine /bin/sh]
2020/06/13 18:30:30 New container ID: b22bbc6ee478
2020/06/13 18:30:30 Image already exists. Not downloading.
2020/06/13 18:30:30 Image to overlay mount: a24bb4013296
2020/06/13 18:30:30 Cmd args: [/proc/self/exe setup-netns b22bbc6ee478]
2020/06/13 18:30:30 Cmd args: [/proc/self/exe setup-veth b22bbc6ee478]
2020/06/13 18:30:30 Cmd args: [/proc/self/exe child-mode --mem=128 --swap=0 --img=a24bb4013296 b22bbc6ee478 /bin/sh]
/ # apk add python3
fetch http://dl-cdn.alpinelinux.org/alpine/v3.12/main/x86_64/APKINDEX.tar.gz
fetch http://dl-cdn.alpinelinux.org/alpine/v3.12/community/x86_64/APKINDEX.tar.gz
(1/10) Installing libbz2 (1.0.8-r1)
(2/10) Installing expat (2.2.9-r1)
(3/10) Installing libffi (3.3-r2)
(4/10) Installing gdbm (1.13-r1)
(5/10) Installing xz-libs (5.2.5-r0)
(6/10) Installing ncurses-terminfo-base (6.2_p20200523-r0)
(7/10) Installing ncurses-libs (6.2_p20200523-r0)
(8/10) Installing readline (8.0.4-r0)
(9/10) Installing sqlite-libs (3.32.1-r0)
(10/10) Installing python3 (3.8.3-r0)
Executing busybox-1.31.1-r16.trigger
OK: 53 MiB in 24 packages
/ # python3
Python 3.8.3 (default, May 15 2020, 01:53:50)
[GCC 9.3.0] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> a1 = bytearray(100 * 1024 * 1024)
Killed
/ #
```

需要注意的一件事是，我们使用 --swap = 0 将分配给该容器的 swap 设置为零。否则 Cgroup 虽然限制 RAM 使用，但它允许容器使用无限的交换空间。当 swap 设置为零时，容器将完全限制为允许的 RAM 总量。

# 关于作者

我是 Shuveb Hussain，是 Linux-focused 博客的作者。 你可以在 [Twitter](https://twitter.com/shuveb) 上关注我，在那里我发布与技术相关的内容，主要针对 Linux，性能，可扩展性和云技术。

---
via: https://unixism.net/2020/06/containers-the-hard-way-gocker-a-mini-docker-written-in-go/

作者：[unixism](https://unixism.net/about-unixism/)
译者：[alandtsang](https://github.com/alandtsang)
校对：[校对者 ID](https://github.com/校对者 ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
