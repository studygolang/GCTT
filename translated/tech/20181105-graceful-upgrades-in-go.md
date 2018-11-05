# Go中的优雅升级

优化升级背后的想法是在进程运行过程中，在用户无感知的情况下对程序的配置和代码进行更换(升级)。尽管这听起来很危险、容易出错，不可取，而且像是一个馊主意 - 我(的想法)和你一样。 但是，有些时候你的确需要它们。这通常发生在一个没有负载均衡层的环境中。我们在*Cloudfare*就遇到了这种情况，这使得我们必须研究这类问题、并实现各类的解决方案。

![](https://blog.cloudflare.com/content/images/2018/10/thing.jpg)
*Dingle Dangle! by [Grant C](https://www.flickr.com/photos/grant_subaru/14175646490). (CC-BY 2.0)*

巧合的是，在实现优雅升级的过程中涉及到了一些有趣的底层系统编程知识，这或许就是为什么现在已经有了很多选项。通过深入阅读以了解在哪里需要折衷，以及为什么你需要真正使用我们即将开源的Go库。如果你觉得我们有些啰嗦，你也可以去[github](https://github.com/cloudflare/tableflip)翻阅相关代码，或者在[godoc](https://godoc.org/github.com/cloudflare/tableflip)上阅读我们的文档。

## 基础

那么一个进程执行优雅升级到底意味着什么？让我们用一个web服务器作为例子：在优雅升级发生的时候，我们希望正在进行的HTTP请求不会中断，且不会看到错误信息。

我们知道HTTP连接是建立在TCP连接之上的，并且我们使用了BSD套接字API的TCP接口。然后我们告诉操作系统我们希望在80端口上接收连接请求，然后操作系统分配给我们一个监听中的套接字，我们在这个套接字上调用`Accept()`来等待新的客户端连接请求。

如果操作系统没有在80端口上监听套接字，或者没有任何东西在套接字上调用`Accept()`，那么一个新的客户端连接请求将会被拒绝。优雅升级的诀窍就在于当我们因为某些原因重启我们的服务时，这两件事情都不会发生。那现在就让我们由简入深地看看如何实现这种方式。

## 仅使用`Exec()`

好吧，我们先看看实现它有多困难。我们先仅仅使用`Exec()`调用来创建一个新的二进制执行程序（而不是首先就fork它）。通过将正在运行的代码替换成磁盘中新的代码，这就是我们想要做的。

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

不幸的是，上述代码中有个致命的缺陷，那就是我们不能“撤销”该次执行。设想一下，如果我们的配置文件有很多空白行或者有一个额外的分号。那么这个新的进程在会读取这个配置文件时会得到一个错误，接着程序就会退出。

即使这个执行(exec)执行成功了，这个解决方案也只是假设这个进程的启动是瞬时完成的。我们也可能会遇到内核拒绝新的连接的情况，因为TCP[监听队列会有溢出的情况](https://veithen.github.io/2014/01/01/how-tcp-backlog-works-in-linux.html)。

![](https://blog.cloudflare.com/content/images/2018/10/Example1-1.png)

_如果不频繁地调用`Accept()`，那么新的连接可能就会被丢弃。_

具体来说，新的二进制文件会在`Exec()`之后初始化的过程中花费一些时间，这将会导致`Accept()`调用被推迟。这意味着新的连接将会持续堆积，直到一些连接被丢弃。所以普通的`Exec()`调用并不能完成优化升级的工作。

## `监听`(`Listen()`)一切

刚才使用的`Exec()`调用并不能解决我们的问题，所以我们需要尝试下一种更好的方案。让我们fork然后exec一个新的进程，然后按照它通用的启动例程开始。在某些时候，它会通过监听某些地址来创建套接字，但是由于`errno 48`(也被称为地址已经被使用)的错误返回码导致这些套接字无法立即开始工作。这是因为操作系统内核阻止了我们想要在旧进程使用的地址和端口上进行监听的操作。


当然，有一个标志位可以解决这个问题：`SO_REUSEPORT`。 这个标志位将告诉内核一个事实，即复用给定的端口和地址已经存在监听的套接字，而不是分配一个新的。

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

那么现在两个进程都在相同的套接字上监听，并且升级的工作正常进行。对不对？

`SO_REUSEPORT`特性在内核中的作用有点怪异。作为一个系统开发，我们倾向于将套接字视为套接字调用返回的文件描述符。然而内核却将套接字的数据结构和指向该套接字一个或多个文件描述符作了区分。如果你使用了`SO_REUSEPORT`标志位，内核将会创建一个独立的套接字结构体，而不是另一个文件描述符。因此，新旧两个进程分别指向了两个独立的套接字，却碰巧共享了相同的地址。这就导致了一个不可避免的竞态条件：旧进程使用的套接字上新创建的，但尚未被接收的连接将会被内核孤立并杀掉。**Github** [写了一篇关于这个问题的优秀博客](https://githubengineering.com/glb-part-2-haproxy-zero-downtime-zero-delay-reloads-with-multibinder/#haproxy-almost-safe-reloads)。

**Github** 的工程师使用了`sendmsg`系统调用上的一个[名为: 辅助数据](http://man7.org/linux/man-pages/man0/sys_socket.h.0p.html)的模糊特性解决了`SO_REUSEPORT`的问题。事实证明，辅助数据可以包含文件描述符，使用这个系统调用API对于**Github**来说是有意义的，因为它(这个API)允许了它们优雅地和HAProxy进行集成。由于我们可以随意地更改程序，因此我们可以使用更加简单的替代方案。


## Nginx: 通过fork和exec共享套接字

Nginx是互联网上经过各种测试和值得信赖的Web服务器，并且它也恰好支持优雅升级。难能可贵的是，我们也在*Cloudfare**使用了它，因此我们对它(Nginx)的实现有足够的信心。

Nginx是基于单核心单进程的模型编写的，这就意味着Nginx并没有派生出一大堆的线程，而是在每个逻辑CPU核心上运行一个进程。此外，Nginx还有一个额外的主进程可以用来进行优化地升级服务。

Nginx主(master)进程负责创建Nginx所监听的套接字，并与其它的工作(worker)进程共享这些套接字。这非常简单直接： 首先，在所有的监听套接字上清除`FD_CLOEXEC`标志位。这意味着在执行`exec()`系统调用之后这些套接字并不会被关闭。然后主进程习惯性的执行`fork()`/`exec()`来派生工作进程，并将文件描述符作为环境变量传递给这些工作进程。

优雅升级使用了相同的机制。我们通过[学习Nginx文档](http://nginx.org/en/docs/control.html#upgrade)来派生一个(PID为1176)的主进程。这个操作就像从旧的主进程(PID为1017)那里继承了所有已存在的监听者的工作进程一样。然后新的主进程开始派生自己的工作进程：

```bash
CGroup: /system.slice/nginx.service
       	├─1017 nginx: master process /usr/sbin/nginx -g daemon on; master_process on;
       	├─1019 nginx: worker process
       	├─1021 nginx: worker process
       	├─1024 nginx: worker process
       	├─1026 nginx: worker process
       	├─1027 nginx: worker process
       	├─1028 nginx: worker process
       	├─1029 nginx: worker process
       	├─1030 nginx: worker process
       	├─1176 nginx: master process /usr/sbin/nginx -g daemon on; master_process on;
       	├─1187 nginx: worker process
       	├─1188 nginx: worker process
       	├─1190 nginx: worker process
       	├─1191 nginx: worker process
       	├─1192 nginx: worker process
       	├─1193 nginx: worker process
       	├─1194 nginx: worker process
       	└─1195 nginx: worker process
```
