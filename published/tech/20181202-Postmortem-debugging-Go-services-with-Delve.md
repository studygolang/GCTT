首发于：https://studygolang.com/articles/17961

# 使用 Delve 调试 Go 服务的一次经历

> Vladimir Varankin 写于 2018/12/02

某天，我们生产服务上的几个实例突然不能处理外部进入的流量，HTTP 请求成功通过负载均衡到达实例，但是之后却 hang 住了。接下来记录的是一次调试在线 Go 服务的惊心动魄的经历。

正是下面逐步演示的操作，帮助我们定位了问题的根本原因。

简单起见，我们将起一个 Go 写的 HTTP 服务作为调试使用，这个服务实现的细节暂时不做深究（之后我们将深入分析代码）。一个真实的生产应用可能包含很多组件，这些组件实现了业务罗和服务的基础架构。我们可以确信，这些应用已经在生产环境“身经百战” :)。

源代码以及配置细节可以查看[GitHub 仓库](https://github.com/narqo/postmortem-debug-go)。为了完成接下来的工作，你需要一台 Linux 系统的虚机，这里我使用[vagrant-hostmanager](https://github.com/sevos/vagrant-hostmanager) 插件。`Vagrantfile` 在 GitHub 仓库的根目录，可以查看更多细节。

让我们开启虚机，构建 HTTP 服务并且运行起来，可以看到下面的输出：

```shell
$ Vagrant up
Bringing Machine 'server-test-1' up with 'virtualbox' provider...

$ Vagrant SSH server-test-1
Welcome to Ubuntu 18.04.1 LTS (GNU/Linux 4.15.0-33-generic x86_64)
···
vagrant@server-test-1:~$ cd /vagrant/example/server
vagrant@server-test-1:/vagrant/example/server$ Go build
vagrant@server-test-1:/vagrant/example/server$ ./server --addr=:10080
server listening addr=:10080
```

通过 `curl` 发送请求到所起的 HTTP 服务，可以判断其是否处于工作状态，新开一个 terminal 并执行下面的命令：

```shell
$ curl 'http://server-test-1:10080'
OK
```

为了模拟失败的情况，我们需要发送大量请求到 HTTP 服务，这里我们使用 HTTP benchmark 测试工具[wrk](https://github.com/wg/wrk) 进行模拟。我的 MacBook 是 4 核的，所以使用 4 个线程运行 wrk，能够产生 1000 个连接，基本能够满足需求。

```shell
$ wrk -d1m -t4 -c1000 'http://server-test-1:10080'
Running 1m test @ http://server-test-1:10080
  4 threads and 1000 connections
  ···
```

一会的时间，服务器 hang 住了。甚至等 wrk 跑完之后，服务器已经不能处理任何请求：

```shell
$ curl --max-time 5 'http://server-test-1:10080/'
curl: (28) Operation timed out after 5001 milliseconds with 0 bytes received
```

我们遇到麻烦了！让我们分析一下。

---

*在我们生产服务的真实场景中，服务器起来以后，goroutines 的数量由于请求的增多而迅速增加，之后便失去响应。对 pprof 调试句柄的请求变得非常非常慢，看起来就像服务器“死掉了”。同样，我们也尝试使用 `SIGQUIT` 命令杀掉进程以[释放所运行 Goroutines 堆栈](https://golang.org/pkg/os/signal/#hdr-Default_behavior_of_signals_in_Go_programs)，但是收不到任何效果。*

## GDB 和 Coredump

我们可以使用 GDB（GNU Debugger）尝试进入正在运行的服务内部。

---

*在生产环境运行调试器可能需要额外的权限，所以与你的团队提前沟通是很明智的。*

---

在虚机上再开启一个 SSH 会话，找到服务器的进程 id 并使用调试器连接到该进程：

```shell
$ Vagrant SSH server-test-1
Welcome to Ubuntu 18.04.1 LTS (GNU/Linux 4.15.0-33-generic x86_64)
···
vagrant@server-test-1:~$ pgrep server
1628
vagrant@server-test-1:~$ cd /vagrant
vagrant@server-test-1:/vagrant$ sudo gdb --pid=1628 example/server/server
GNU gdb (Ubuntu 8.1-0ubuntu3) 8.1.0.20180409-git
···
```

调试器连接到服务器进程之后，我们可以运行 GDB 的 `bt` 命令（aka backtrace）来检查当前线程的堆栈信息：

```shell
(gdb) bt
#0  runtime.futex () at /usr/local/go/src/runtime/sys_linux_amd64.s:532
#1  0x000000000042b08b in runtime.futexsleep (addr=0xa9a160 <runtime.m0+320>, ns=-1, val=0) at /usr/local/go/src/runtime/os_linux.go:46
#2  0x000000000040c382 in runtime.notesleep (n=0xa9a160 <runtime.m0+320>) at /usr/local/go/src/runtime/lock_futex.go:151
#3  0x0000000000433b4a in runtime.stoplockedm () at /usr/local/go/src/runtime/proc.go:2165
#4  0x0000000000435279 in runtime.schedule () at /usr/local/go/src/runtime/proc.go:2565
#5  0x00000000004353fe in runtime.park_m (gp=0xc000066d80) at /usr/local/go/src/runtime/proc.go:2676
#6  0x000000000045ae1b in runtime.mcall () at /usr/local/go/src/runtime/asm_amd64.s:299
#7  0x000000000045ad39 in runtime.rt0_go () at /usr/local/go/src/runtime/asm_amd64.s:201
#8  0x0000000000000000 in ?? ()
```

说实话我并不是 GDB 的专家，但是显而易见 Go 运行时似乎使线程进入睡眠状态了，为什么呢？

调试一个正在运行的进程是不明智的，不如将该线程的 coredump 保存下来，进行离线分析。我们可以使用 GDB 的 `gcore` 命令，该命令将 core 文件保存在当前工作目录并命名为 `core.<process_id>`。

```shell
(gdb) gcore
Saved corefile core.1628
(gdb) quit
A debugging session is active.

	Inferior 1 [process 1628] will be detached.

Quit anyway? (y or n) y
Detaching from program: /vagrant/example/server/server, process 1628
```

core 文件保存后，服务器没必要继续运行，使用 `kill -9` 结束它。

我们能够注意到，即使是一个简单的服务器，core 文件依然会很大（我这一份是 1.2G）, 对于生产的服务来说，可能会更加巨大。

*如果需要了解更多使用 GDB 调试的技巧，可以继续阅读[使用 GDB 调试 Go 代码](https://golang.org/doc/gdb)。*

## 使用 Delve 调试器

[Delve](https://github.com/derekparker/delve) 是一个针对 Go 程序的调试器。它类似于 GDB，但是更关注 Go 的运行时、数据结构以及其他内部的机制。

如果你对 Delve 的内部实现机制很感兴趣，那么我十分推荐你阅读 Alessandro Arzilli 在 GopherCon EU 2018 所作的演讲，[[Internal Architecture of Delve, a Debugger For Go](https://www.youtube.com/watch?v=IKnTr7Zms1k)]。

Delve 是用 Go 写的，所以安装起来非常简单：

```shell
$ Go get -u Github.com/derekparker/delve/cmd/dlv
```

Delve 安装以后，我们就可以通过运行 `dlv core <path to service binary> <core file>` 来分析 core 文件。我们先列出执行 coredump 时正在运行的所有 Goroutines。Delve 的 `goroutines` 命令如下：

```shell
$ dlv core example/server/server core.1628

(dlv) Goroutines
  ···
  Goroutine 4611 - User: /vagrant/example/server/metrics.go:113 main.(*Metrics).CountS (0x703948)
  Goroutine 4612 - User: /vagrant/example/server/metrics.go:113 main.(*Metrics).CountS (0x703948)
  Goroutine 4613 - User: /vagrant/example/server/metrics.go:113 main.(*Metrics).CountS (0x703948)
```

不幸的是，在真实生产环境下，这个列表可能会很长，甚至会超出 terminal 的缓冲区。由于服务器为每一个请求都生成一个对应的 Goroutine，所以 `goroutines` 命令生成的列表可能会有百万条。我们假设现在已经遇到这个问题，并想一个方法来解决它。

Delve 支持 "headless" 模式，并且能够通过[JSON-RPC API](https://github.com/derekparker/delve/tree/master/Documentation/api) 与调试器交互。

运行 `dlv core` 命令，指定想要启动的 Delve API server：

```shell
$ dlv core example/server/server core.1628 --listen :44441 --headless --log
API server listening at: [::]:44441
INFO[0000] opening core file core.1628 (executable example/server/server)  layer=debugger
```

调试服务器运行后，我们可以发送命令到其 TCP 端口并将返回结果以原生 JSON 的格式存储。我们以上面相同的方式得到正在运行的 Goroutines，不同的是我们将结果存储到文件中：

```shell
$ Echo -n '{"method":"RPCServer.ListGoroutines","params":[],"id":2}' | nc -w 1 localhost 44441 > server-test-1_dlv-rpc-list_goroutines.json
```

现在我们拥有了一个（比较大的）JSON 文件，里面存储大量原始信息。推荐使用[jq](https://stedolan.github.io/jq/) 命令进一步了解 JSON 数据的原貌，举例：这里我获取 JSON 数据的 result 字段的前三个对象：

```shell
$ jq '.result[0:3]' server-test-1_dlv-rpc-list_goroutines.json
[
  {
    "id": 1,
    "currentLoc": {
      "pc": 4380603,
      "file": "/usr/local/go/src/runtime/proc.go",
      "line": 303,
      "function": {
        "name": "runtime.gopark",
        "value": 4380368,
        "type": 0,
        "goType": 0,
        "optimized": true
      }
    },
    "userCurrentLoc": {
      "pc": 6438159,
      "file": "/vagrant/example/server/main.go",
      "line": 52,
      "function": {
        "name": "main.run",
        "value": 6437408,
        "type": 0,
        "goType": 0,
        "optimized": true
      }
    },
    "goStatementLoc": {
      "pc": 4547433,
      "file": "/usr/local/go/src/runtime/asm_amd64.s",
      "line": 201,
      "function": {
        "name": "runtime.rt0_go",
        "value": 4547136,
        "type": 0,
        "goType": 0,
        "optimized": true
      }
    },
    "startLoc": {
      "pc": 4379072,
      "file": "/usr/local/go/src/runtime/proc.go",
      "line": 110,
      "function": {
        "name": "runtime.main",
        "value": 4379072,
        "type": 0,
        "goType": 0,
        "optimized": true
      }
    },
    "threadID": 0,
    "unreadable": ""
  },
  ···
]
```

JSON 数据中的每个对象都代表了一个 Goroutine。通过[命令手册](https://github.com/derekparker/delve/blob/master/Documentation/cli/README.md#goroutines)

可知，`goroutines` 命令可以获得每一个 Goroutines 的信息。通过手册我们能够分析出 `userCurrentLoc` 字段是服务器源码中 Goroutines 最后出现的地方。

为了能够了解当 core file 创建的时候，goroutines 正在做什么，我们需要收集 JSON 文件中包含 `userCurrentLoc` 字段的函数名字以及其行号：

```shell
$ jq -c '.result[] | [.userCurrentLoc.function.name, .userCurrentLoc.line]' server-test-1_dlv-rpc-list_goroutines.json | sort | uniq -c

   1 ["internal/poll.runtime_pollWait",173]
1000 ["main.(*Metrics).CountS",95]
   1 ["main.(*Metrics).SetM",105]
   1 ["main.(*Metrics).startOutChannelConsumer",179]
   1 ["main.run",52]
   1 ["os/signal.signal_recv",139]
   6 ["runtime.gopark",303]
```

大量的 Goroutines( 上面是 1000 个 ) 在函数 `main.(*Metrics).CoutS` 的 95 行被阻塞。现在我们回头看一下我们服务器的[源码](https://github.com/narqo/postmortem-debug-go)。

在 `main` 包中找到 `Metrics` 结构体并且找到它的 `CountS` 方法（example/server/metrics.go）。

```go
// CountS increments counter per second.
func (m *Metrics) CountS(key string) {
    m.inChannel <- NewCountMetric(key, 1, second)
}
```

我们的服务器在往 `inChannel` 通道发送的时候阻塞住了。让我们找出谁负责从这个通道读取数据，深入研究代码之后我们找到了[下面的函数](https://github.com/narqo/postmortem-debug-go/blob/2c42ca73ebd500fe8da1c6ac8ecaf4af143aca78/example/server/metrics.go#L109)：

```shell
// starts a consumer for inChannel
func (m *Metrics) startInChannelConsumer() {
    for inMetrics := range m.inChannel {
   	    // ···
    }
}
```

这个函数逐个地从通道中读取数据并加以处理，那么什么情况下发送到这个通道的任务会被阻塞呢？

当处理通道的时候，根据 Dave Cheney 的[通道准则](https://dave.cheney.net/2014/03/19/channel-axioms)，只有四种情况可能导致通道有问题：

- 向一个 nil 通道发送
- 从一个 nil 通道接收
- 向一个已关闭的通道发送
- 从一个已关闭的通道接收并立即返回零值

第一眼就看到了“向一个 nil 通道发送”，这看起来像是问题的原因。但是反复检查代码后，`inChannel` 是由 `Metrics` 初始化的，不可能为 nil。

n 你可能会注意到，使用 `jq` 命令获取到的信息中，没有 `startInChannelConsumer` 方法。会不会是因为在 `main.(*Metrics).startInChannelConsumer` 的某个地方阻塞而导致这个（可缓冲）通道满了？

Delve 能够提供从开始位置到 `userCurrentLoc` 字段之间的初始位置信息，这个信息存储到 `startLoc` 字段中。使用下面的 jq 命令可以查询出所有 Goroutines, 其初始位置都在函数 `startInChannelConsumer` 中：

```shell
$ jq '.result[] | select(.startLoc.function.name | test("startInChannelConsumer$"))' server-test-1_dlv-rpc-list_goroutines.json

{
  "id": 20,
  "currentLoc": {
    "pc": 4380603,
    "file": "/usr/local/go/src/runtime/proc.go",
    "line": 303,
    "function": {
      "name": "runtime.gopark",
      "value": 4380368,
      "type": 0,
      "goType": 0,
      "optimized": true
    }
  },
  "userCurrentLoc": {
    "pc": 6440847,
    "file": "/vagrant/example/server/metrics.go",
    "line": 105,
    "function": {
      "name": "main.(*Metrics).SetM",
      "value": 6440672,
      "type": 0,
      "goType": 0,
      "optimized": true
    }
  },
  "startLoc": {
    "pc": 6440880,
    "file": "/vagrant/example/server/metrics.go",
    "line": 109,
    "function": {
      "name": "main.(*Metrics).startInChannelConsumer",
      "value": 6440880,
      "type": 0,
      "goType": 0,
      "optimized": true
    }
  },
  ···
}
```

结果中有一条信息非常振奋人心！

在 `main.(*Metrics).startInChannelConsumer`，109 行（看结果中的 startLoc 字段），有一个 id 为 20 的 Goroutines 阻塞住了！

拿到 Goroutines 的 id 能够大大降低我们搜索的范围（并且我们再也不用深入庞大的 JSON 文件了）。使用 Delve 的 `goroutines` 命令我们能够将当前 Goroutines 切换到目标 Goroutines，然后可以使用 `stack` 命令打印该 Goroutines 的堆栈信息：

```shell
$ dlv core example/server/server core.1628

(dlv) Goroutine 20
Switched from 0 to 20 (thread 1628)

(dlv) stack -full
0  0x000000000042d7bb in runtime.gopark
   at /usr/local/go/src/runtime/proc.go:303
       lock = unsafe.Pointer(0xc000104058)
       reason = waitReasonChanSend
···
3  0x00000000004066a5 in runtime.chansend1
   at /usr/local/go/src/runtime/chan.go:125
       c = (unreadable empty OP stack)
       elem = (unreadable empty OP stack)

4  0x000000000062478f in main.(*Metrics).SetM
   at /vagrant/example/server/metrics.go:105
       key = (unreadable empty OP stack)
       m = (unreadable empty OP stack)
       value = (unreadable empty OP stack)

5  0x0000000000624e64 in main.(*Metrics).sendMetricsToOutChannel
   at /vagrant/example/server/metrics.go:146
       m = (*main.Metrics)(0xc000056040)
       scope = 0
       updateInterval = (unreadable could not find loclist entry at 0x89f76 for address 0x624e63)

6  0x0000000000624a2f in main.(*Metrics).startInChannelConsumer
   at /vagrant/example/server/metrics.go:127
       m = (*main.Metrics)(0xc000056040)
       inMetrics = main.Metric {Type: TypeCount, Scope: 0, Key: "server.req-incoming",...+2 more}
       nextUpdate = (unreadable could not find loclist entry at 0x89e86 for address 0x624a2e)
```

从下往上分析：

（6）一个来自通道的新 `inMetrics` 值在 `main.(*Metrics).startInChannelConsumer` 中被接收

（5）我们调用 `main.(*Metrics).sendMetricsToOutChannel` 并且在 `example/server/metrics.go` 的 146 行进行处理

（4）然后 `main.(*Metrics).SetM` 被调用

一直运行到 `runtime.gopark` 中的 `waitReasonChanSend` 阻塞！

一切的一切都明朗了！

单个 Goroutines 中，一个从缓冲通道读取数据的函数，同时也在往通道中发送数据。当进入通道的值达到通道的容量时，消费函数继续往已满的通道中发送数据就会造成自身的死锁。由于单个通道的消费者死锁，那么每一个尝试往通道中发送数据的请求都会被阻塞。

---

这就是我们的故事，使用上述调试技术帮助我们发现了问题的根源。那些代码是很多年前写的，甚至从没有人看过这些代码，也万万没有想到会导致这么大的问题。

如你所见，并不是所有问题都能由工具解决，但是工具能够帮助你更好地工作。我希望，通过此文能够激励你多多尝试这些工具。我非常乐意倾听你们处理类似问题的其它解决方案。

**Vladimir*是一个后端开发工程师，目前就职于*adjust.com. @tvii on Twitter, @narqo on Github**

---

via: https://blog.gopheracademy.com/advent-2018/postmortem-debugging-delve/

作者：[Vladimir Varankin](https://blog.gopheracademy.com/advent-2018/postmortem-debugging-delve/)
译者：[hantmac](https://github.com/hantmac)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
