# Go：gops 如何与 runtime 交互？

![img](https://github.com/studygolang/gctt-images2/blob/master/20200405-go-how-does-gops-interact-with-the-runtime/1_3PCyB5PhH_NEZoNnj693dA.png?raw=true)

本文基于 Go 1.13 和 gops 0.3.7。

`gops` 旨在帮助开发人员诊断 Go 进程并与之交互。它提供了追踪运行中的程序数秒钟，通过 `pprof` 获取 CPU 的 profile，甚至直接与垃圾回收器交互的能力。

## 发现

`gops` 提供了一种发现服务，它可以列出计算机上运行的 Go 进程。不带参数运行 `gops` 仅显示 Go 进程。为了举例说明，我启动了一个程序，该程序计算素数直到一百万。这是发现程序的输出：

```bash
295 1 gops          go1.13 /go/src/github.com/google/gops/gops
168 1 prime-number* go1.13 /go/prime-number/prime-number
```

`gops` 发现了上面启动的程序和它自己的进程。我们需要的仅仅是进程 ID，因此，基于此输出，我们可以开始与程序进行交互。不过，还是让我们了解一下 `gops` 如何过滤 Go 进程。

首先，`gops` 列出所有的进程。接着，对于每个进程，它打开二进制文件读取符号表：

![img](https://github.com/studygolang/gctt-images2/blob/master/20200405-go-how-does-gops-interact-with-the-runtime/1_LyVcQzBGP3i4aCwzwmzoFA.png?raw=true)

如果符号表包含了 `runtime.man` （主 goroutine 的入口）或者 `main.main` （我们程序的入口），则可以将其标记为一个 Go 程序。

*有关符号表的更多信息，我建议你阅读我的文章“Go：如何使用符号表”。了解关于主 goroutine 的更多信息，建议阅读我的文章“[Go: g0，特殊的 Goroutine](https://medium.com/a-journey-with-go/go-g0-special-goroutine-8c778c6704d8)”*

`gops` 还会通过读取符号表的 `runtime.buildVersion` 来读取 Go 的版本。然而，由于二进制文件中的符号表可以被剥离，`gops` 需要另一种方法来检测 Go 二进制文件。我们用剥离后的二进制文件再试一次：

```bash
295 1 gops            go1.13             /go/src/..../gops
168 1 prime-number-s* unknown Go version /go/.../prime-number-s
```

由于缺少符号表，即使程序被正确的标记为 Go 二进制文件，它也无法检测 Go 版本。根据[可执行文件格式](https://en.wikipedia.org/wiki/Comparison_of_executable_file_formats) -- `ELF`，`MZ`，等等 --`gops` 会读取各段来查找嵌在二进制文件中的构建 ID。一旦发现过程结束，它就可以开始与程序交互。

## 交互

与其他 Go 程序交互的唯一条件是确保它们启动了 `gops` agent。该 agent 是一个简单的 listener，它将为 gops 请求提供服务。这很简单，只需添加以下几行：

```go
if err := agent.Listen(agent.Options{}); err != nil {
    log.Fatal(err)
}
```

然后，任何启动了 agent 的程序都可以与 `gops` 交互。这里是执行 `stats` 命令的例子：

```bash
# gops stats 168
goroutines: 6210
OS threads: 9
GOMAXPROCS: 2
num CPU: 2
```

有关更多命令，你可以参考[项目的文档](https://github.com/google/gops#manual)。如果缺少该 agent，你在与其交互时会收到一个错误：

```bash
Couldn't resolve addr or pid 168 to TCPAddress: couldn't get port for PID 168
```

该错误表明 `gops` 在通过 TCP 寻找暴露的 endpoint 以便与程序通信。让我们画出这个 package 的工作流来了解它的工作原理：

## 工作流

`gops` 通过 TCP 和要读取的程序暴露的 endpoint 来与其通信：

![img](https://github.com/studygolang/gctt-images2/blob/master/20200405-go-how-does-gops-interact-with-the-runtime/1_V5turxRPbzzq9rHbqOjpbg.png?raw=true)

分配给每个程序的端口都写在一个配置文件中，例如 `path/to/config/{processID}` ，这使得 `gops` 很容易知道暴露的端口。然后，`gops` 可以将命令标记发送给程序，agent 将会收集数据并响应：

![img](https://github.com/studygolang/gctt-images2/blob/master/20200405-go-how-does-gops-interact-with-the-runtime/1_PvRmDO4yXEdm6Z7xeysh8A.png?raw=true)

---
via: [https://medium.com/a-journey-with-go/go-how-does-gops-interact-with-the-runtime-778d7f9d7c18](https://medium.com/a-journey-with-go/go-how-does-gops-interact-with-the-runtime-778d7f9d7c18)

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[DoubleLuck](https://github.com/DoubleLuck)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
