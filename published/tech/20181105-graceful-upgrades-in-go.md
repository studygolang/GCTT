首发于：https://studygolang.com/articles/16363

# Go 中的优雅升级

优化升级背后的想法是在进程运行过程中，在用户无感知的情况下对程序的配置和代码进行更换（升级）。尽管这听起来很危险、容易出错、不可取，并且像是一个馊主意 - 事实上我（的想法）和你一样。 但是，有些时候你的确需要它们。这通常在一个没有负载均衡层的环境中会遇到这种问题。我们在 *Cloudfare* 也遇到了这种情况，这使得我们必须研究这类问题、并尝试、实现各类的解决方案。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/graceful-upgrade/thing.jpg)
*Dingle Dangle! by [Grant C](https://www.flickr.com/photos/grant_subaru/14175646490). (CC-BY 2.0) 可以理解成在汽车行驶的过程中更换发送动机。*

巧合的是，在尝试、实现优雅升级的过程中涉及到了一些有趣的底层系统编程知识，这或许就是为什么现在已经有了许多优秀的解决方案。通过深入阅读它们以了解在什么地方需要折衷，以及为什么你需要使用我们即将开源的 Go 库。当然，如果你觉得我们过于啰嗦，你也可以直接去[github](https://github.com/cloudflare/tableflip) 翻阅相关代码，或者在[godoc](https://godoc.org/github.com/cloudflare/tableflip) 上阅读我们的文档。

## 基础

那么一个进程执行优雅升级到底意味着什么？让我们用一个 Web 服务器作为例子：在优雅升级发生的时候，我们希望正在进行的 HTTP 请求不会中断，而且不会看到任何错误信息。

我们知道 HTTP 连接是建立在 TCP 连接之上的，我们使用了 BSD 套接字 API 的 TCP 接口。然后我们告诉操作系统我们希望在 `80` 端口上接收连接请求，然后操作系统分配给我们一个监听中的套接字，我们在这个套接字上调用 `Accept()` 等待新的客户端连接请求。

如果操作系统在 `80` 端口上没有处于监听的套接字，或者没有任何东西在套接字上调用 `Accept()`，那么新的客户端连接请求将会被拒绝。优雅升级的诀窍就在于当我们因为某些原因需要重启我们的服务时，这两件事情都不会发生。那么现在就让我们由简入深地看看如何实现这种方式吧。

## 仅使用 `Exec()`

好吧，我们先看看实现它有多困难。我们先仅仅使用 `Exec()` 调用来创建一个新的二进制执行程序（而不是一开始就 `fork` 它）。来将正在运行的代码替换成磁盘中新的代码，这就是我们想要做的。

```go
// The following is pseudo-Go.

func main() {
	var ln net.Listener
	if isUpgrade {
		ln = net.FileListener(os.NewFile(uintptr(fdNumber), "listener"))
	} else {
		ln = net.Listen(network, address)
	}

	go handleRequests(ln)

	<-waitForUpgradeRequest

	syscall.Exec(os.Argv[0], os.Argv[1:], os.Environ())
}
```

不幸的是，上述代码中有个致命的缺陷，那就是我们不能“撤销”该次执行。设想一下，如果我们的配置文件有很多空白行或者有一个额外的分号。那么这个新的进程在会读取这个配置文件时会得到一个错误，接着新进程就会退出。

即使这个 `exec` 执行成功了，这个解决方案也只是假设这个进程的启动是瞬时完成的。我们也可能会遇到内核拒绝新的连接的情况，这是因为 TCP 会有[监听队列溢出的情况](https://veithen.github.io/2014/01/01/how-tcp-backlog-works-in-linux.html)。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/graceful-upgrade/Example1-1.png)

_如果不频繁地调用 `Accept()`，那么新的连接可能会被丢弃。_

具体来说，新的二进制文件会在 `Exec()` 之后初始化的过程中花费一些时间进行初始化，这将会导致 `Accept()` 调用被推迟。这意味着新的连接将会持续堆积，直到一些连接被丢弃。所以普通的 `Exec()` 调用并不能完成优雅升级的工作。

## `监听` (`Listen()`) 一切

刚才使用的 `Exec()` 调用并不能解决我们的问题，所以我们需要尝试下一种更好的方案。如果我们 `fork` 然后 `exec` 一个新的进程，然后按照它通用的启动例程开始。在某些时候，它会通过监听某些地址来创建套接字，但是可能会由于 `errno 48`( 也被称为地址已经被使用 ) 的错误返回码而导致这些套接字无法立即开始工作。这是因为操作系统内核阻止了我们想要在旧进程使用的地址和端口上进行监听的操作。

当然，有一个标志位可以解决这个问题：`SO_REUSEPORT`。 这个标志位将告诉内核一个事实：即复用给定的端口和地址已经存在监听的套接字，而不是分配一个新的。

```go
func main() {
	ln := net.ListenWithReusePort(network, address)

	go handleRequests(ln)

	<-waitForUpgradeRequest

	cmd := exec.Command(os.Argv[0], os.Argv[1:])
	cmd.Start()

	<-waitForNewProcess
}
```

那么现在两个进程都在相同的套接字上监听，并且升级的工作也在正常进行。对不对？

`SO_REUSEPORT` 特性在内核中的作用有点怪异。作为系统开发人员，我们倾向于将套接字视为套接字调用返回的文件描述符。然而内核却将套接字的数据结构和指向该套接字一个或多个文件描述符作了区分。如果你使用了 `SO_REUSEPORT` 标志位，内核将会创建一个独立的套接字结构体，而不是另一个文件描述符。因此，新旧两个进程分别指向了两个独立的套接字，却碰巧共享了相同的地址。这就导致了一个不可避免的竞态条件：旧进程使用的套接字上新创建的，但尚未被接收的连接将会被内核孤立并杀死。**Github** [写了一篇关于这个问题的优秀博客](https://githubengineering.com/glb-part-2-haproxy-zero-downtime-zero-delay-reloads-with-multibinder/#haproxy-almost-safe-reloads)。

**Github** 的工程师使用了 `sendmsg` 系统调用上的一个[名为 : 辅助数据](http://man7.org/linux/man-pages/man0/sys_socket.h.0p.html) 的模糊特性解决了 `SO_REUSEPORT` 的问题。事实证明，辅助数据可以包含文件描述符，使用这个系统调用 API 对于 **Github** 来说是有意义的，因为它 ( 这个 API) 允许了他们可以优雅地和 HAProxy 进行集成。由于我们可以随意地更改程序，因此我们可以使用更加简单的替代方案。

## Nginx: 通过 fork 和 exec 共享套接字

Nginx 是互联网上经过各种测试和值得信赖的 Web 服务器，并且它也恰好支持优雅升级。难能可贵的是，我们也在*Cloudfare*使用了它，因此我们对它 (Nginx) 的实现有足够的信心。

Nginx 是基于单核心单进程的模型编写的，这就意味着 Nginx 并没有派生出一大堆的线程，而是在每个逻辑 CPU 核心上运行一个进程。此外，Nginx 还有一个额外的主进程可以用来进行优雅地升级服务。

Nginx 主 (master) 进程负责创建 Nginx 所监听的套接字，并与其它的工作 (worker) 进程共享这些套接字。这非常的简单直接： 首先，在所有的监听套接字上清除 `FD_CLOEXEC` 标志位，这意味着在执行 `exec()` 系统调用之后这些套接字并不会被关闭。然后主进程习惯性的执行 `fork()`/`exec()` 来派生工作进程，并将文件描述符作为环境变量传递给这些工作进程。

Nginx 的优雅升级使用了相同的机制。我们通过[Nginx 文档](http://nginx.org/en/docs/control.html#upgrade) 来派生一个 (PID 为 1176) 的主进程。这个操作就像从旧的主进程 (PID 为 1017) 那里继承了所有已存在的监听者的工作进程一样。然后新的主进程开始派生自己的工作进程：

```bash
CGroup: /system.slice/nginx.service
       	├─ 1017 nginx: master process /usr/sbin/nginx -g daemon on; master_process on;
       	├─ 1019 nginx: worker process
       	├─ 1021 nginx: worker process
       	├─ 1024 nginx: worker process
       	├─ 1026 nginx: worker process
       	├─ 1027 nginx: worker process
       	├─ 1028 nginx: worker process
       	├─ 1029 nginx: worker process
       	├─ 1030 nginx: worker process
       	├─ 1176 nginx: master process /usr/sbin/nginx -g daemon on; master_process on;
       	├─ 1187 nginx: worker process
       	├─ 1188 nginx: worker process
       	├─ 1190 nginx: worker process
       	├─ 1191 nginx: worker process
       	├─ 1192 nginx: worker process
       	├─ 1193 nginx: worker process
       	├─ 1194 nginx: worker process
       	└─ 1195 nginx: worker process
```
此时，有两个完全独立的 Nginx 进程在运行。PID 为 1176 的进程也许是一个新版本的 Nginx，或者是运行了更新后的配置文件的 Nginx 进程。当一个新连接到达 `80` 端口时，内核将在这 16 个工作进程中选择一个进程来处理这个连接请求。

在执行完剩余的步骤之后，我们最终完全替换了 Nginx。

```bash
CGroup: /system.slice/nginx.service
		 ├─ 1176 nginx: master process /usr/sbin/nginx -g daemon on; master_process on;
		 ├─ 1187 nginx: worker process
		 ├─ 1188 nginx: worker process
		 ├─ 1190 nginx: worker process
		 ├─ 1191 nginx: worker process
		 ├─ 1192 nginx: worker process
		 ├─ 1193 nginx: worker process
		 ├─ 1194 nginx: worker process
		 └─ 1195 nginx: worker process
```

这时候，当一个连接请求达到时，内核将在这 8 个工作进程中选择一个来处理该请求。

Nginx 优雅升级整个过程非常复杂，所以 Nginx 有一个安全措施。如果我们在第一次升级还未完成时就请求第二次升级，我们将会得到如下的错误信息：

```bash
[crit] 1176#1176: the changing binary signal is ignored: you should shutdown or terminate before either old or new binary's process
```

这是非常合理的，没有理由可以说明在任意给定的时间点上应该存在两个以上的进程。这是一个很好的用例，所以我们也希望我们 Go 的解决方案中也应该有此行为。

## 优雅升级的愿望清单

Nginx 实现的优雅升级的方式非常好。它有一个明确的生命周期，用来确定在任何时间点的有效操作。

![](https://static.studygolang.com/gctt/upgrade-lifecycle.svg)

它还解决了我们在使用其他方法时遇到的问题。说真的，我们确实想要以 Nginx 优雅升级作为范例来编写一个 Go 库。

- 在成功升级后，旧代码不会继续运行。
- 当新进程初始化时发生崩溃不会有任何影响。
- 在任意时间点仅有一个升级操作处于活动状态。

当然，Go 社区已经为这样的场景开源了一些优秀的库，我们也阅读了这些库：

- [https://github.com/alext/tablecloth](https://github.com/alext/tablecloth)( 这个优秀的名字给予了我们灵感 )
- [github.com/astaxie/beego/grace](https://godoc.org/github.com/astaxie/beego/grace)
- [https://github.com/facebookgo/grace](https://github.com/facebookgo/grace)
- [github.com/crawshaw/littleboss](https://github.com/crawshaw/littleboss)

这里我们列举了几个例子，它们在实现和权衡上各不相同，但它们都不是我们想要的。最常见的一个问题就是，它们都被设计成旨在提供 HTTP Server 的优雅升级。这使得它们的 API 非常友好，但是却丧失了支持其它基于套接字的协议所需要的灵活性。所以事实上，我们别无选择，只能编写自己的库。这个库被称为*tableflip*。享受编程乐趣不是编写这个库的动机。

## *tableflip*

*tableflip* 是一个和 Nginx 优雅升级方式类似的 Go 库。下面的是一个如何使用这个库的代码示例：

```go
upg, _ := tableflip.New(tableflip.Options{})
defer upg.Stop()

// Do an upgrade on SIGHUP
go func() {
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGHUP)
    for range sig {
   	    _ = upg.Upgrade()
    }
}()

// Start a HTTP server
ln, _ := upg.Fds.Listen("tcp", "localhost:8080")
server := http.Server{}
go server.Serve(ln)

// Tell the parent we are ready
_ = upg.Ready()

// Wait to be replaced with a new process
<-upg.Exit()

// Wait for connections to drain.
server.Shutdown(context.TODO())
```

我们调用 `Upgrader.Upgrade` 并使用必要的 `net.listeners` 派生一个新的进程，然后等待新进程通知我们它是否初始化成功、死亡，或者超时。如果在升级的过程中调用这个函数则会返回一个错误。

`Upgrader.Fds.Listen` 的灵感来自于 `facebookgo/grace`，它可以轻松继承 `net.Listener`。事实上在后台实现中，`Fds` 确保了未被使用的继承套接字会被清除。这里也包括了*UNIX*套接字，因为[UnlinkOnClose](https://golang.org/pkg/net/#UnixListener.SetUnlinkOnClose) 而变得棘手。如果你愿意的话，你也可以直接将 `*os.File` 对象传递给新进程。

最后，`Upgrader.Ready` 会清理未使用的文件描述符并通知父进程初始化工作已经完成。此时，父进程可以安全退出。至此，正常的优化升级的周期结束。

---

via: https://blog.cloudflare.com/graceful-upgrades-in-go/

作者：[Lorenz Bauer](https://blog.cloudflare.com/author/lorenz-bauer/)
译者：[barryz](https://github.com/barryz)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
