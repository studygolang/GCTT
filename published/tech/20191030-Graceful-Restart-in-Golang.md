首发于：https://studygolang.com/articles/25123

# Golang Http 服务的优雅重启

（2015 年 4 月更新）：[Florian von Bock](https://github.com/fvbock) 已将本文中描述的内容实现成了一个名为[Endless](https://github.com/fvbock/endless)的 Go 程序包。

对于 Golang HTTP 的服务，我们可能需要重启来升级或者更改某些配置。如果你（像我曾经一样）因为网络服务器对优雅重启很重视就理所当然地认为它（优雅重启）早已实现了，那么这份教程将会对你很有用处。因为在 Golang 中，你需要自己动手来实现。

实际上，这里需要解决两个问题。首先是 UNIX 端的优雅重启，即进程无需关闭监听套接字即可自行重启的机制。第二个问题是确保所有进行中的请求被正确完成或超时。

## 在不关闭套接字的情况下重启

- 派生一个继承监听中的套接字的新进程。
- 新进程执行初始化并开始接受套接字上的连接。
- 之后，子进程立即向父进程发送信号，使父进程停止接收连接并终止。

### 派生一个新的进程

有多种使用 Golang 的库去实现派生进程的方法，在本文中的例子中，我们选择用 [exec.Command](https://golang.org/pkg/os/exec/#Command)。这是因为此函数返回的 [Cmd 结构体](https://golang.org/pkg/os/exec/#Cmd) 具有 `ExtraFiles` 成员，该成员可以使打开的文件（除了 `stdin/err/out` 之外）被新进程继承。

看起来如下所示：

```go
file := netListener.File() // this returns a Dup()
path := "/path/to/executable"
args := []string{
    "-graceful"}

cmd := exec.Command(path, args...)
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
cmd.ExtraFiles = []*os.File{file}

err := cmd.Start()
if err != nil {
    log.Fatalf("gracefulRestart: Failed to launch, error: %v", err)
}
```

在上面的代码中，`netListener` 是指向 [net.Listener](https://golang.org/pkg/net/#Listener) 的指针，`net.Listener` 用于侦听 HTTP 请求。如果你要升级，则 `path` 变量应包含新可执行文件的路径（可能与当前正在运行的文件相同）。

上面代码中的关键是 `netListener.File()` 会返回一个文件描述符 [dup(2)](https://pubs.opengroup.org/onlinepubs/009695399/functions/dup.html)。这个文件描述符不会设置 `FD_CLOEXEC` [标识](https://pubs.opengroup.org/onlinepubs/009695399/functions/fcntl.html)，这会导致文件在子进程中被关闭（不是我们想要的的情况）。

你可能会看到一些示例，通过命令行参数将需要继承的文件描述符传递给子进程，但是 `ExtraFiles` 实现的方式使这些变得没有必要。文档指出 "如果输入 non-nil，输入的切片索引为 i 则读取的文件描述符为 3 + i。” 这意味着在上述代码段中，子进程中继承的文件描述符将始终为 3，因此无需显式传递它（译者注：这一段有点难以理解，ExtraFiles 可以指定额外的文件句柄传递给子进程，也就是可以通过这个方法将 `netListener.File()` 传递给子进程，而子进程通过 `f := os.NewFile(3, "")` 来读取，为什么是 3 呢？因为前面还有几个默认句柄，所以传入额外的句柄需要从 3 开始读取）。

最后，`args` 数组包含 `-graceful` 选项：你的程序将需要某种方式来通知子进程这是正常重启的一部分，并且子进程应重新使用套接字，而不是打开新的套接字。还有一种方法是通过环境变量来实现这个功能。

### 初始化子进程

这是程序启动序列的一部分

```go
server := &http.Server{Addr: "0.0.0.0:8888"}

var gracefulChild bool
var l net.Listever
var err error

flag.BoolVar(&gracefulChild, "graceful", false, "listen on fd open 3 (internal use only)")

if gracefulChild {
    log.Print("main: Listening to existing file descriptor 3.")
    f := os.NewFile(3, "")
    l, err = net.FileListener(f)
} else {
    log.Print("main: Listening on a new file descriptor.")
    l, err = net.Listen("tcp", server.Addr)
}
```

### 通知父进程停止

至此，子进程已经准备好接收请求，但是在此之前，需要告诉父进程停止接收请求并退出，这可能是这样的：

```go
if gracefulChild {
    parent := syscall.Getppid()
    log.Printf("main: Killing parent pid: %v", parent)
    syscall.Kill(parent, syscall.SIGTERM)
}

server.Serve(l)
```

### 进行中的请求 完成/超时

为此，我们需要使用 [sync.WaitGroup](https://golang.org/pkg/sync/#WaitGroup) 跟踪打开的连接。我们需要给每次新增的连接使用 `wg.Add`，并在每次关闭连接时使用 `wg.Done`。

```go
var httpWg sync.WaitGroup
```

乍一看，Golang 标准的 http 包没有提供任何对 Accept() 或 Close() 操作的钩子，但是使用接口解决了这个问题。（非常感谢 [Jeff R. Allen](http://nella.org/jra/)的这篇[文章](http://blog.nella.org/zero-downtime-upgrades-of-tcp-servers-in-go/)）。

这是一个 listener 的示例，它在每个 Accept() 上使用 `httpWg.Add(1)` 计数。首先，我们对 `net.Listener` 进行“子类化”（`stop` 和 `stopped` 的作用将在下文体现）：

```go
type gracefulListener struct {
    net.Listener
    stop    chan error
    stopped bool
}
```

接下来，我们“覆盖” Accept 方法。（暂时不要考虑 `gracefulConn`，稍后再介绍）。

```go
func (gl *gracefulListener) Accept() (c net.Conn, err error) {
    c, err = gl.Listener.Accept()
    if err != nil {
        return
    }

    c = gracefulConn{Conn: c}

    httpWg.Add(1)
    return
}
```

我们还需要一个构造函数：

```go
func newGracefulListener(l net.Listener) (gl *gracefulListener) {
    gl = &gracefulListener{Listener: l, stop: make(chan error)}
    Go func() {
        _ = <-gl.stop
        gl.stopped = true
        gl.stop <- gl.Listener.Close()
    }()
    return
}
```

上面的函数启动 Goroutine 的原因是因为在上面的 `Accept()` 中无法完成此操作，因为它会被 `gl.Listener.Accept()` 阻塞。 Goroutine 通过关闭文件描述符来解除阻塞。

我们的 `Close()` 方法仅将 `nil` 发送给上述 Goroutine 的 `stop channel` 即可完成其余工作。

```go
func (gl *gracefulListener) Close() error {
    if gl.stopped {
        return syscall.EINVAL
    }
    gl.stop <- nil
    return <-gl.stop
}
```

最后，从 `net.TCPListener` 中提取文件描述符。

```go
func (gl *gracefulListener) File() *os.File {
    tl := gl.Listener.(*net.TCPListener)
    fl, _ := tl.File()
    return fl
}
```

当然，我们还需要一个嵌入了 `net.Conn` 的结构体，该结构体会通过 `httpWg.Done()` 来减少 `Close()` 上的计数：

```go
type gracefulConn struct {
    net.Conn
}

func (w gracefulConn) Close() error {
    httpWg.Done()
    return w.Conn.Close()
}
```

如果要使用上述优美版本的 Listener，我们需要将 `server.Serve(l)` 行替换为：

```go
netListener = newGracefulListener(l)
server.Serve(netListener)
```

还有一件事。你应避免挂起客户端无意关闭的连接。最好按以下方式创建服务：

```go
server := &http.Server{
        Addr:           "0.0.0.0:8888",
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 16}
```

---

via: https://grisha.org/blog/2014/06/03/graceful-restart-in-golang/

作者：[humblehack](https://twitter.com/humblehack)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[lxbwolf](https://github.com/lxbwolf)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com)
