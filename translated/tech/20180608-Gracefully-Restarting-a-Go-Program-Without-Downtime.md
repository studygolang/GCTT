# 无停机优雅重启 Go 程序

## 什么是优雅重启

在不停机的情况下，就地部署一个应用程序的新版本或者修改其配置的能力已经成为现代软件系统的标配。这篇文章讨论优雅重启一个应用的不同方法，并且提供一个功能独立的案例来深挖实现细节。如果你不熟悉 Teleport 话，Teleport 是我们使用 Golang 针对弹性架构设计的 [SHH 和 Kubernetes 特权访问管理解决方案](https://gravitational.com/teleport/)。使用 Go 建立和维护服务的开发者和网站可靠性工程师(SRE)应该对这篇文章有兴趣。

## SO_REUSEPORT vs 复制套接字的背景

为了推进 Teleport 高可用的工作，我们最近花了些时间研究如何优雅重启 Teleport 的 TLS 和 SSH 的端口监听器[(GitHub issue #1679)](https://github.com/gravitational/teleport/pull/1679)。我们的目标是能够更新一个 Teleport 二进制文件而不需要让实例停止服务。

Marek Majkowski 在他的博客文章[《为什么一个 NGINX 工作线程会承担所有负载？》](https://gravitational.com/teleport/) 讨论了两种普遍的方法。这些方法可以被如下概括:

* 你可以在套接字上设置 `SO_REUSEPORT` ，从而让多个进程能够被绑定到同一个端口上。利用这个方法，你会有多个接受队列向多个进程提供数据。
* 复制套接字，并把它以文件的形式传送给一个子进程，然后在新的进程中重新创建这个套接字。使用这种方法，你将有一个接受队列向多个进程提供数据。]

在我们初期的讨论中，我们了解到几个关于 `SO_REUSEPORT` 的问题。我们的一个工程师之前使用这个方法，并且注意到由于其多个接受队列，有时候会丢弃挂起的 TCP 连接。除此之外，当我们进行这些讨论的时候，Go 并没有很好地支持在一个 `net.Listener` 上设置 `SO_REUSEPORT`。然而，在过去的几天中，在这个问题上有了进展，看起来像 [Go 不久就会支持设置套接字属性](https://github.com/golang/go/issues/9661)。

第二种方法也很吸引人，因为它的简单性以及大多数开发人员熟悉的传统Unix 的 fork/exec 产生模型，即将所有打开文件传递给子进程的约定。需要注意的一点，`os/exec` 包实际上不赞同这种用法。主要是出于安全上的考量，它只传递 `stdin` , `stdout` 和 `stderr` 给子进程。然而， os 包确实提供较低级的原语，可用于将文件传递给子程序，这就是我们想做的。

## 使用信号切换套接字进程所有者

在我们看源码之前，了解一些这个方法如何工作的细节是值得的。

启动一个全新的 Teleport 程序后，该进程会在绑定的端口上创建一个监听套接字接受所有入站流量。对于 Teleport,入口流量就是 LTS 和 SSH 流量。我们添加了一个处理 [SIGUSR2](https://www.gnu.org/software/libc/manual/html_node/Kill-Example.html) 信号的句柄，该句柄让 Teleport 复制监听套接字，然后生成一个新的进程，同时将监听套接字以文件的形式和这个套接字的元数据以环境变量的形式传入给该进程。一旦新的进程开始，他会依据传进来的文件和元数据重建这个套接字，并且处理它所获得的流量。

应该注意的是，当一个套接字被复制时，入栈流量会在两个套接字之间以轮询的方式进行负载均衡。如下图所示，这就意味着有一段时间，两个 Teleport 进程都会接受新的连接。

![](https://github.com/studygolang/gctt-images/blob/master/gracefully-restart-a-go-program-without-downtime/graceful-restart-diag-1.png?raw=true)

父进程的关闭是相同的事情，但是反过来做。一旦 Teleport 进程接受到 SIGOUIT 信号，他会开始关闭这个进程，停止接受新的连接，等待所有的现有连接断开或是超时发生。一旦入站流量被清空，这个濒死进程就会关闭它的监听套接字并且退出。这种情况下，新的进程会接管内核发送过来的所有请求。

![](https://github.com/studygolang/gctt-images/blob/master/gracefully-restart-a-go-program-without-downtime/graceful-restart-diag-2.png?raw=true)

## 优雅重启演练

我们基于上面的方法写了一个简单的程序，你可以自己尝试使用一下。源代码在文章的最后，你可以按照以下步骤尝试这个例子。

首先，编译和启动程序。
```
$ go build restart.go
$ ./restart &
[1] 95147
$ Created listener file descriptor for :8080.

$ curl http://localhost:8080/hello
Hello from 95147!
```

将 USR2 信号发送给初始进程。现在，当你访问这个 HTTP 入口的时候，他会返回两个不同的进程的 PID。

```
$ kill -SIGUSR2 95147
user defined signal 2 signal received.
Forked child 95170.
$ Imported listener file descriptor for :8080.

$ curl http://localhost:8080/hello
Hello from 95170!
$ curl http://localhost:8080/hello
Hello from 95147!
```

杀死初始进程后，你将只会从新的进程中获得返回。

```
$ kill -SIGTERM 95147
signal: killed
[1]+  Exit 1                  go run restart.go
$ curl http://localhost:8080/hello
Hello from 95170!
$ curl http://localhost:8080/hello
Hello from 95170!
```

最后杀死新进程，访问将会被拒绝。

```
$ kill -SIGTERM 95170
$ curl http://localhost:8080/hello
curl: (7) Failed to connect to localhost port 8080: Connection refused
```

## 总结和示例源代码

像你看到，一旦你了解了他是如何工作的，增加优雅重启功能到 Go 写的服务中是相当简单的事情，并且有效地提高服务使用者的用户体验。如果你想在 Teleport 中看到这一点，我们邀请你瞧瞧我们的参考 [AWS SSH 和 Kubernetes 堡垒机部署](https://github.com/gravitational/teleport/tree/master/examples/aws)，里面包含了一个 ansible 脚本，该脚本利用就地优雅重启实现无停机更新 Teleport 二进制文件。

[Golang 优雅重启案例源代码](https://gist.github.com/russjones/09e7ace4c7497515f6bd0285f710c2e4)

```go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type listener struct {
	Addr     string `json:"addr"`
	FD       int    `json:"fd"`
	Filename string `json:"filename"`
}

func importListener(addr string) (net.Listener, error) {
	// 从环境变量中抽离出被编码的 listener 的元数据。
	listenerEnv := os.Getenv("LISTENER")
	if listenerEnv == "" {
		return nil, fmt.Errorf("unable to find LISTENER environment variable")
	}

	// 解码 listener 的元数据。
	var l listener
	err := json.Unmarshal([]byte(listenerEnv), &l)
	if err != nil {
		return nil, err
	}
	if l.Addr != addr {
		return nil, fmt.Errorf("unable to find listener for %v", addr)
	}

	// 文件已经被传入到这个进程中，从元数据中抽离文件描述符和名字，为 listener 重建/发现 *os.file
	listenerFile := os.NewFile(uintptr(l.FD), l.Filename)
	if listenerFile == nil {
		return nil, fmt.Errorf("unable to create listener file: %v", err)
	}
	defer listenerFile.Close()

	// Create a net.Listener from the *os.File.
	ln, err := net.FileListener(listenerFile)
	if err != nil {
		return nil, err
	}

	return ln, nil
}

func createListener(addr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return ln, nil
}

func createOrImportListener(addr string) (net.Listener, error) {
	// 尝试为地址导入一个 listener, 如果导入成功，则使用。
	ln, err := importListener(addr)
	if err == nil {
		fmt.Printf("Imported listener file descriptor for %v.\n", addr)
		return ln, nil
	}
	// 没有 listener 被导入，这就意味着进程必须自己创建一个。
	ln, err = createListener(addr)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Created listener file descriptor for %v.\n", addr)
	return ln, nil
}

func getListenerFile(ln net.Listener) (*os.File, error) {
	switch t := ln.(type) {
	case *net.TCPListener:
		return t.File()
	case *net.UnixListener:
		return t.File()
	}
	return nil, fmt.Errorf("unsupported listener: %T", ln)
}

func forkChild(addr string, ln net.Listener) (*os.Process, error) {
	// 从 listener 中获取文件描述符，在环境变量编码在传递给这个子进程的元数据。
	lnFile, err := getListenerFile(ln)
	if err != nil {
		return nil, err
	}
	defer lnFile.Close()
	l := listener{
		Addr:     addr,
		FD:       3,
		Filename: lnFile.Name(),
	}
	listenerEnv, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	// 将 stdin, stdout, stderr 和 listener 传入子进程。
	// 译注: 以上四个文件描述符分别为 0,1,2,3
	files := []*os.File{
		os.Stdin,
		os.Stdout,
		os.Stderr,
		lnFile,
	}

	// 获取当前环境变量，并且传入子进程。
	environment := append(os.Environ(), "LISTENER="+string(listenerEnv))

	// 获取当前进程名和工作目录
	execName, err := os.Executable()
	if err != nil {
		return nil, err
	}
	execDir := filepath.Dir(execName)

	// 生成子进程
	p, err := os.StartProcess(execName, []string{execName}, &os.ProcAttr{
		Dir:   execDir,
		Env:   environment,
		Files: files,
		Sys:   &syscall.SysProcAttr{},
	})
	if err != nil {
		return nil, err
	}

	return p, nil
}

func waitForSignals(addr string, ln net.Listener, server *http.Server) error {
	signalCh := make(chan os.Signal, 1024)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGQUIT)
	for {
		select {
		case s := <-signalCh:
			fmt.Printf("%v signal received.\n", s)
			switch s {
			case syscall.SIGHUP:
				// Fork 一个子进程。
				p, err := forkChild(addr, ln)
				if err != nil {
					fmt.Printf("Unable to fork child: %v.\n", err)
					continue
				}
				fmt.Printf("Forked child %v.\n", p.Pid)

				// 创建一个在 5 秒钟过去的 Context, 使用这个超时定时器关闭。
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// 返回关闭过程中发生的任何错误。
				return server.Shutdown(ctx)
			case syscall.SIGUSR2:
				// Fork 一个子进程。
				p, err := forkChild(addr, ln)
				if err != nil {
					fmt.Printf("Unable to fork child: %v.\n", err)
					continue
				}
				// 输出被 fork 的子进程的 PID，并等待更多的信号。
				fmt.Printf("Forked child %v.\n", p.Pid)
			case syscall.SIGINT, syscall.SIGQUIT:
				// 创建一个在 5 秒钟过去的 Context, 使用这个超时定时器关闭。
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// 返回关闭过程中发生的任何错误。
				return server.Shutdown(ctx)
			}
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from %v!\n", os.Getpid())
}

func startServer(addr string, ln net.Listener) *http.Server {
	http.HandleFunc("/hello", handler)

	httpServer := &http.Server{
		Addr: addr,
	}
	go httpServer.Serve(ln)

	return httpServer
}

func main() {
	// Parse command line flags for the address to listen on.
	var addr string
	flag.StringVar(&addr, "addr", ":8080", "Address to listen on.")

	// Create (or import) a net.Listener and start a goroutine that runs
	// a HTTP server on that net.Listener.
	ln, err := createOrImportListener(addr)
	if err != nil {
		fmt.Printf("Unable to create or import a listener: %v.\n", err)
		os.Exit(1)
	}
	server := startServer(addr, ln)

	// 等待复制或结束的信号
	err = waitForSignals(addr, ln, server)
	if err != nil {
		fmt.Printf("Exiting: %v\n", err)
		return
	}
	fmt.Printf("Exiting.\n")
}
```

## 如果你读到了这里

Teleport 是一个开源软件，你可以免费地在 [GitHub](https://github.com/gravitational/teleport) 上深入了解它。如果你对 Teleport 或是其他类似的分布式系统软件的工作有兴趣，我们时刻期待着[优秀的软件工程师](https://gravitational.com/careers/systems-engineer/)。

------

via: https://gravitational.com/blog/golang-ssh-bastion-graceful-restarts/

作者：[gravitational team](https://gravitational.com/about/)
译者：[magichan](https://github.com/magichan)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
