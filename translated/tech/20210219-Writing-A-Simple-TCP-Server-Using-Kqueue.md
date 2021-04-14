# 用 kqueue 实现一个简单的 TCP Server

## 介绍

在 [非阻塞 I/O 超简明介绍](https://dev.to/frosnerd/explain-non-blocking-i-o-like-i-m-five-2a5f) 中，
我们已经讨论过现代 Web 服务器可以处理大量并发请求，这得益于现代操作系统内核内置的事件通知机制。
受 Linux epoll [ [文档](https://man7.org/linux/man-pages/man7/epoll.7.html) ] 启发，
FreeBSD 发明了 kqueue [ [论文](https://people.freebsd.org/~jlemon/papers/kqueue.pdf) ,
[文档](https://www.freebsd.org/cgi/man.cgi?query=kqueue&sektion=2) ]

这篇文章我们将仔细研究下 kqueue，我们会用 Go 实现一个基于 kqueue event loop 的 TCP server，
你可以在 Github 上找到 [源代码](https://github.com/FRosner/FrSrv) 。
要运行代码必须使用和 FreeBSD 兼容的操作系统，比如 macOS。

注意 kqueue 不仅能处理 socket event，而且还能处理文件描述符 event、信号、异步 I/O event、子进程状态改变 event、
定时器以及用户自定义 event。它确实通用和强大。

我们这篇文章主要分为一下几部分讲解。
首先，我们会在先从理论出发设计我们的 TCP Server。
然后，我们会去实现它的必要的模块。
最后我们会对整个过程进行总结以及思考。

## 设计

我们 TCP Server 大概有以下几部分：
一个监听 TCP 的 socket、
接收客户端连接的 socket、
一个内核事件队列（kqueue），
还有一个事件循环机制来轮询这个队列。
下面这个图描述了接收连接的场景。

![](https://raw.githubusercontent.com/h1z3y3/gctt-images2/master/20210219-Writing-A-Simple-TCP-Server-Using-Kqueue/accepting-incoming-connections.png)

当客户端想要连接服务端，一个连接请求将会被放到 TCP 连接队列中，
而内核会将一个新的事件放到 kqueue 中。
这个事件将会在事件循环时被处理，事件循环会接受请求，并创建一个新的客户端连接。
下面这个图描述了新创建的 socket 如何从客户端读取请求。

![](https://raw.githubusercontent.com/h1z3y3/gctt-images2/master/20210219-Writing-A-Simple-TCP-Server-Using-Kqueue/read-data-from-the-client.png)

客户端写数据到新创建的 socket，内核会将一个新 event 放到 kqueue 中，表示在这个 socket 中有等待读取的数据。
事件循环将轮询到这个事件，并从 socket 读取数据。
注意只有一个 socket 监听连接，而我们将为每一个客户端连接创建新的 socket。

下文要讨论实现细节，可以大概按照下面的步骤实现我们的设计。

1 创建，绑定以及监听新的 socket
2 创建 kqueue
3 订阅 socket event
4 循环队列获取 event 并处理它们

## 实现

为了避免单个文件有大量系统调用，我们拆分成几个不同模块：

* 一个 `socket` 模块来处理所有管理 socket 的相关功能，
* 一个 `kqueue` 模块来处理事件循环，
* 最后 `main` 模块用来整合所有模块并启动我们的 TCP server。

我们下面从 `socket` 模块开始。

### 定义 Socket

首先，让我们创建一个 socket 结构体。类 Unix 操作系统，比如 FreeBSD，会把 socket 作为文件。
为了用 Go 实现 socket，我们需要了解 [文件描述符](https://www.freebsd.org/cgi/man.cgi?query=fd&sektion=4&manpath=freebsd-release-ports)。
所以我们可以创建一个类似下面带有文件描述符的结构体。

```golang

type Socket struct {
    FileDescriptor int
}

```

我们期望我们的 socket 可以应对不同的场景，比如：读、写 socket 数据，以及关闭 socket。
在 Go 中，要支持这些操作，需要实现通用的 interface，比如 `io.Reader`，`io.Writer`，还有 `io.Closer`

首先，实现 `io.Reader` 这个接口，他会调用 [read()](https://www.freebsd.org/cgi/man.cgi?query=read&sektion=2) 系统函数。
这个函数会返回读到字节的数量，以及进行读操作时可能发生的错误。

```golang

func (socket Socket) Read(bytes []byte) (int, error) {
    if len(bytes) == 0 {
        return 0, nil
    }
    numBytesRead, err := syscall.Read(socket.FileDescriptor, bytes)
    if err != nil {
        numBytesRead = 0
    }

    return numBytesRead, err
}

```

类似的，我们通过调用 [write()](https://www.freebsd.org/cgi/man.cgi?query=write&sektion=2) 来实现 `io.Writer` 接口。

```golang

func (socket Socket) Write(bytes []byte) (int, error) {
    numBytesWritten, err := syscall.Write(socket.FileDescriptor, bytes)
    if err != nil {
        numBytesWritten = 0
    }
    return numBytesWritten, err
}

```

最后关闭 socket 可以调用 [close()](https://www.freebsd.org/cgi/man.cgi?query=close&apropos=0&sektion=2) ，并传入 socket 对应的文件描述符。

```golang
func (socket *Socket) Close() error {
    return syscall.Close(socket.FileDescriptor)
}
```

为了稍后能打印一些有用的错误和日志，我们也需要实现 `fmt.Stringer` 接口。
我们通过不同的文件描述符来区分不同的 socket。

```golang
func (socket *Socket) String() string {
    return strconv.Itoa(socket.FileDescriptor)
}
```

### 监听一个 Socket

定义好 Socket 之后，我们需要初始化它，并让它一个监听特定 IP 和 端口的。
监听一个 socket 也可以通过一些系统函数来实现。
现在先整体看一下我们实现的 `Listen()` 方法，然后再一步步进行分析。

```golang
func Listen(ip string, port int) (*Socket, error) {
    socket := &Socket{}

    socketFileDescriptor, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
    if err != nil {
        return nil, fmt.Errorf("failed to create socket (%v)", err)
    }

    socket.FileDescriptor = socketFileDescriptor

    socketAddress := &syscall.SockaddrInet4{Port: port}
    copy(socketAddress.Addr[:], net.ParseIP(ip))

    if err = syscall.Bind(socket.FileDescriptor, socketAddress); err != nil {
        return nil, fmt.Errorf("failed to bind socket (%v)", err)
    }

    if err = syscall.Listen(socket.FileDescriptor, syscall.SOMAXCONN); err != nil {
        return nil, fmt.Errorf("failed to listen on socket (%v)", err)
    }

    return socket, nil
}
```

首先调用了 [socket()](https://www.freebsd.org/cgi/man.cgi?query=socket&apropos=0&sektion=2) 函数，
这将会创建通信的接入点，并返回描述符编号。
它需要 3 个参数：

* 地址类型：我们用的是 `AF_INET` (IPv4)
* socket 类型：我们用 `SOCKET_STREAM`，代表基于字节流连续、可靠的双向连接。
* 协议类型：`0` 在 `SOCKET_STREAM` 类型下代表的是 TCP。

然后，我们调用了 [bind()](https://www.freebsd.org/cgi/man.cgi?query=bind&apropos=0&sektion=2) 方法来指定新创建 socket 的协议地址。
`bind()` 方法的第一个参数是文件描述符，第二个参数是包含地址信息的结构体指针。
我们在这里使用了 Go 预定义的 `SockaddrInet4` 结构体，并指定要绑定的 IP 地址和端口。

最后，我们调用了 [listen()](https://www.freebsd.org/cgi/man.cgi?query=listen&apropos=0&sektion=2) 方法，这样我们就能等待接收连接了。
它的第二个参数是连接请求队列的最大长度。
我们使用了内核参数 `SOMAXCONN` ，在我的 Mac 上默认是 128。
你可以通过执行 `sysctl kern.ipc.somaxconn` 来获取这个值。

### 定义事件循环

同样的，我们将定义一个结构体来表示 kqueue 的事件循环。
我们必须要保存 kqueue 的文件描述符以及 socket 的文件描述符,
我们当然也能将我们前面定义的 socket 对象作为指针来替代 `SocketFileDescriptor`。

```golang

type EventLoop struct {
    KqueueFileDescriptor int
    SocketFileDescriptor int
}

```

接下来，我们需要一个函数根据我们提供的 socket 创建一个事件循环。
和之前一样，我们需要用一系列系统函数去创建 Kqueue。
我们还是先看下整个函数，然后再一步步拆解来看。

```golang

func NewEventLoop(s *socket.Socket) (*EventLoop, error) {
    kQueue, err := syscall.Kqueue()
    if err != nil {
        return nil,
            fmt.Errorf("failed to create kqueue file descriptor (%v)", err)
    }

    changeEvent := syscall.Kevent_t{
        Ident:  uint64(s.FileDescriptor),
        Filter: syscall.EVFILT_READ,
        Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
        Fflags: 0,
        Data:   0,
        Udata:  nil,
    }

    changeEventRegistered, err := syscall.Kevent(
        kQueue,
        []syscall.Kevent_t{changeEvent},
        nil,
        nil
    )
    if err != nil || changeEventRegistered == -1 {
        return nil,
           fmt.Errorf("failed to register change event (%v)", err)
    }

    return &EventLoop{
        KqueueFileDescriptor: kQueue,
        SocketFileDescriptor: s.FileDescriptor
    }, nil
}

```

第一个系统函数 [kqueue()](https://www.freebsd.org/cgi/man.cgi?query=kqueue&apropos=0&sektion=0&format=html) 创建了一个新的内核事件队列，并且返回了它的文件描述符。
我们等会调用 [`kevent()`](https://www.freebsd.org/cgi/man.cgi?query=kqueue&apropos=0&sektion=0&format=html) 的时候会用到这个队列。
`kevent()` 有两个功能，订阅新事件和轮询队列。

我们的例子是要订阅传入连接的事件，
可以通过传递 `kevent` 结构体（在 Go 中，用 `Kevent_t` 表示）给 `kevent()` 这个系统函数来实现订阅。
`Kevent_t` 需要包含以下信息：

* `Ident` 的文件描述符：值是我们 socket 的文件描述符
* 处理事件的 `Filter`：设置为 `EVFILT_READ`，当和监听 socket 一起用时，它代表我们只关心传入连接的事件。
* 代表对这个事件要执行操作的 `Flag`：在我们例子中，我们想要添加（EV_ADD）事件到 `kqueue`，比如说订阅事件，同时要启用（EV_ENABLE）它。Flag 可以使用 `或` 这个位操作进行结合。

其他的几个参数我们就不需要了，创建好这个事件之后，要把它用一个数组包裹，并传递给 `kevent()` 这个系统函数。
最后，我们返回这个等待被轮询的事件循环。接下来让我们实现轮询的函数。

### 事件循环轮询

事件循环是一个简单的 for 循环，可以轮询新的内核事件并进行处理。
之前使用系统函数 `kevent()` 时，订阅轮询就已经完成了，但是现在我们又传递一个空的事件数组给它，
目的是当有新的事件时，新的事件会填充到这个数组。

然后我们就可以一个个循环这些事件并处理它们了。
新的客户端连接会被转换成客户端 socket，所以我们可以从客户端读取或写入数据。
现在让我们看下代码如何循环不同的事件类型。

```golang
func (eventLoop *EventLoop) Handle(handler Handler) {
    for {
        newEvents := make([]syscall.Kevent_t, 10)
        numNewEvents, err := syscall.Kevent(
            eventLoop.KqueueFileDescriptor,
            nil,
            newEvents,
            nil
        )
        if err != nil {
            continue
        }

        for i := 0; i < numNewEvents; i++ {
            currentEvent := newEvents[i]
            eventFileDescriptor := int(currentEvent.Ident)

            if currentEvent.Flags&syscall.EV_EOF != 0 {
                // client closing connection
                syscall.Close(eventFileDescriptor)
            } else if eventFileDescriptor == eventLoop.SocketFileDescriptor {
                // new incoming connection
                socketConnection, _, err := syscall.Accept(eventFileDescriptor)
                if err != nil {
                    continue
                }

                socketEvent := syscall.Kevent_t{
                    Ident:  uint64(socketConnection),
                    Filter: syscall.EVFILT_READ,
                    Flags:  syscall.EV_ADD,
                    Fflags: 0,
                    Data:   0,
                    Udata:  nil,
                }
                socketEventRegistered, err := syscall.Kevent(
                    eventLoop.KqueueFileDescriptor,
                    []syscall.Kevent_t{socketEvent},
                    nil,
                    nil
                )
                if err != nil || socketEventRegistered == -1 {
                    continue
                }
            } else if currentEvent.Filter&syscall.EVFILT_READ != 0 {
                // data available -> forward to handler
                handler(&socket.Socket{
                    FileDescriptor: int(eventFileDescriptor)
                })
            }

            // ignore all other events
        }
    }
}

```

第一种情况，我们要处理 `EV_EOF` 事件，代表客户端想要关闭它的连接的事件。这种情况我们简单的关闭了对应 socket 的文件描述符。

第二种情况代表我们的监听 socket 有连接请求。
我们可以使用系统函数 [accept()](https://www.freebsd.org/cgi/man.cgi?query=accept) 从 TCP 连接请求队列中获取连接请求，
它会为监听 socket 创建一个新的客户端 socket 和新的文件描述符。
我们为这个新创建的 socket 订阅一个新的 `EVFILT_READ` 事件。
在新创建的客户端 socket 中，无论什么时候有可以读取的数据，就会有 `EVFILT_READ` 事件发生。

第三种情况就是处理刚提到的 `EVFILT_READ` 事件，这些事件有客户端 socket 的文件描述符，
我们将其封装在 `Socket` 对象中并传递给要处理它的方法。

要注意我们省略一些错误然后使用了简单的 continue 继续执行循环。现在事件循环函数也写好了，让我们将所有的逻辑封装在 main 函数中并执行。

### main 函数

因为之前已经定义好了 `socket` 和 `kqueue` 模块，我们现在可以非常容易地实现服务器。
我们首先创建一个监听特定 IP 地址和端口的 socket，然后基于它创建一个新的事件循环，
最后我们定义处理输出的函数，来开启我们的事件循环。

```golang
func main() {
    s, err := socket.Listen("127.0.0.1", 8080)
    if err != nil {
        log.Println("Failed to create Socket:", err)
        os.Exit(1)
    }

    eventLoop, err := kqueue.NewEventLoop(s)
    if err != nil {
        log.Println("Failed to create event loop:", err)
        os.Exit(1)
    }

    log.Println("Server started. Waiting for incoming connections. ^C to exit.")

    eventLoop.Handle(func(s *socket.Socket) {
        reader := bufio.NewReader(s)
        for {
            line, err := reader.ReadString('\n')
            if err != nil || strings.TrimSpace(line) == "" {
                break
            }
            s.Write([]byte(line))
        }
        s.Close()
    })
}
```

处理函数会根据换行符逐行读取数据内容，直到它读取到空行，然后会关闭连接。
我们可以通过 `curl` 来测试，`curl` 将会发送一个 GET 请求，并输出响应的内容，响应内容其实就是它发送的 GET 请求体的内容。

![](https://raw.githubusercontent.com/h1z3y3/gctt-images2/master/20210219-Writing-A-Simple-TCP-Server-Using-Kqueue/demo.gif)

## 思考

我们成功用 `kqueue` 实现了一个简单的 TCP Server，当然，这个代码想用于生产环境还需要做很多工作。
我们使用单进程和阻塞 socket 运行，另外，也没有去处理错误。
其实大多数情况下，使用已经存在的库而不是自己调用操作系统内核函数会更好。

没想到用内核操作事件这么难，API 非常复杂，而且必须要读好多文档去找需要怎么做。
然而，这是一个惊人的学习体验。

---
via: [https://dev.to/frosnerd/writing-a-simple-tcp-server-using-kqueue-cah](https://dev.to/frosnerd/writing-a-simple-tcp-server-using-kqueue-cah)

作者：[Frank Rosner](https://dev.to/frosnerd)
译者：[h1z3y3](https://github.com/h1z3y3)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
