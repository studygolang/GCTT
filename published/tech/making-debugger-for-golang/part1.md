已发布：https://studygolang.com/articles/12553

# 实现一个 Golang 调试器（第一部分） #

写这个系列的目的不是为了列出 Golang 编程语言的调试器的所有特性。如果你想看这些内容，可以看下 [Delve](https://github.com/derekparker/delve)。在这篇文章里我们试着去探索下调试器通常是怎样工作的，怎么在 Linux 上完成一个基本的调试，Linux 上比较关心 Golang 的功能，比如 [goroutine](https://golang.org/ref/spec#Go_statements) 。

创建调试器没那么简单。就这一个话题我们单独写一篇文章也讲不完。相反，本篇博文是个开始，这个系列的最终目标是找到解决方案来处理最常见的场景。期间我们会讨论类似 [ELF](https://pl.wikipedia.org/wiki/Executable_and_Linkable_Format), [DWARF](https://en.wikipedia.org/wiki/DWARF) 的话题，还会接触到一些架构相关的问题。

## 环境 ##

整个系列文章中，我们都会使用 [Docker](https://www.docker.com/) 来获取基于 Debian Jessie 的可复制的 playground。我使用的是 [x86-64](https://en.wikipedia.org/wiki/X86-64)，这可以在一定程度上让我们在做一些底层讨论的时候起点作用。项目结构如下：

```
> tree
.
├── Dockerfile
└── src
	└── github.com
		└── mlowicki
			├── debugger
			│   └── debugger.go
			└── hello
				└── hello.go

```

我们马上要用到的调试器的主要文件就是 *debugger.go*，*hello.go* 文件包含我们整个流程中调试的 sample 程序源代码。现在你写最简单的内容就可以：

```go
package main
func main() {
}
```

我们先写一个非常简单的 Dockerfile：

```
FROM golang:1.8.1
RUN apt-get update && apt-get install -y tree
```

为了编译 Docker 镜像，到（Dockerfile 所在的）最外层目录，运行：

```
> docker build -t godebugger .
```

给容器加速，执行：

```
> docker run --rm -it -v "$PWD"/src:/go/src --security-opt seccomp=unconfined godebugger
```

[这里](https://docs.docker.com/engine/security/seccomp/) 有安全运算模式（seccomp）的相关描述。现在剩下的是在容器里编译这这两个程序。第一个可以这样做：

```
> go install --gcflags="-N -l" github.com/mlowicki/hello
```

标识 --gcflag 用于禁止 [内联函数](https://en.wikipedia.org/wiki/Inline_expansion) （-l），编译优化（-N）可以让调试更容易。调试器如下做编译：

```
> go install github.com/mlowicki/debugger

```

在容器的环境变量 *PATH* 中包含 `/go/bin` ，这样不用使用完整路径就可以运行任何刚编译好的程序，不论是 `hello` 还是 `debugger`。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/making-debugger/1_LkLTdWOn2T9N4MelHrDvag.jpeg)

## 第一步 ##

我们的第一个任务很简单。在执行任何指令之前停止程序，然后再运行起来，直到程序停止（不管是自动停止还是出现错误停止）。大多数调试器你都可以这样开始使用。设定一些跟踪点（断点），然后执行类似 `continue` 的指令真正的跑起来，直到停在你要停的地方。我们看看 [Delve](https://github.com/derekparker/delve) 是如何工作的：

```shell
> cat hello.go
package main
import "fmt"
func f() int {
	var n int
	n = 1
	n = 2
	return n
}
func main() {
	fmt.Println(f())
}
> dlv debug
break Type ‘help’ for list of commands.
(dlv) break main.f
Breakpoint 1 set at 0x1087050 for main.f() ./hello.go:5
(dlv) continue
> main.f() ./hello.go:5 (hits goroutine(1):1 total:1) (PC: 0x1087050)
	 1: package main
	 2:
	 3: import "fmt"
	 4:
=>   5: func f() int {
	 6:     var n int
	 7:     n = 1
	 8:     n = 2
	 9:     return n
	10: }
(dlv) next
> main.f() ./hello.go:6 (PC: 0x1087067)
	 1: package main
	 2:
	 3: import "fmt"
	 4:
	 5: func f() int {
=>   6:     var n int
	 7:     n = 1
	 8:     n = 2
	 9:     return n
	10: }
	11:
(dlv) print n
842350461344
(dlv) next
> main.f() ./hello.go:7 (PC: 0x108706f)
	 2:
	 3: import "fmt"
	 4:
	 5: func f() int {
	 6:     var n int
=>   7:     n = 1
	 8:     n = 2
	 9:     return n
	10: }
	11:
	12: func main() {
(dlv) print n
0
(dlv) next
> main.f() ./hello.go:8 (PC: 0x1087077)
	 3: import "fmt"
	 4:
	 5: func f() int {
	 6:     var n int
	 7:     n = 1
=>   8:     n = 2
	 9:     return n
	10: }
	11:
	12: func main() {
	13:     fmt.Println(f())
(dlv) print n
1

```

让我们看看我们自己怎么实现。

第一步是需要给进程（我们的调试器）找一个机制，去控制其他进程（我们要调试的进程）。幸好在 Linux 上我们有这个-- [ptrace](http://man7.org/linux/man-pages/man2/ptrace.2.html)。这还不算。Golang 的 [syscall](https://golang.org/pkg/syscall/) 包提供了一个类似 [PtraceCont](https://golang.org/pkg/syscall/#PtraceCont) 的接口，可以重启被跟踪的进程。因此这里包含了第二部分内容，但是为了有机会在程序开始执行之前设置断点我们还得做点其他的。创建新进程的时候我们可以通过设置属性-- [SysProcAttr](https://golang.org/pkg/syscall/#SysProcAttr) 指定进程行为。其中一个是 *Ptrace* 可以跟踪进程，然后进程会停止并在开启之前给父进程发送 [SIGSTOP signal](http://man7.org/linux/man-pages/man7/signal.7.html)。我们把刚才学到的内容整理成一个工作流程...

```shell
> cat src/github.com/mlowicki/hello/hello.go
package main
import "fmt"
func main() {
	fmt.Println("hello world")
}
> cat src/github.com/mlowicki/debugger/debugger.go
package main
import (
	"flag"
	"log"
	"os"
	"os/exec"
	"syscall"
)
func main() {
	flag.Parse()
	input := flag.Arg(0)
	cmd := exec.Command(input)
	cmd.Args = []string{input}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Ptrace: true}
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	log.Printf("State: %v\n", err)
	log.Println("Restarting...")
	err = syscall.PtraceCont(cmd.Process.Pid, 0)
	if err != nil {
		log.Panic(err)
	}
	var ws syscall.WaitStatus
	_, err = syscall.Wait4(cmd.Process.Pid, &ws, syscall.WALL, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Exited: %v\n", ws.Exited())
	log.Printf("Exit status: %v\n", ws.ExitStatus())
}
> go install -gcflags="-N -l" github.com/mlowicki/hello
> go install github.com/mlowicki/debugger
> debugger /go/bin/hello
2017/05/05 20:09:38 State: stop signal: trace/breakpoint trap
2017/05/05 20:09:38 Restarting...
hello world
2017/05/05 20:09:38 Exited: true
2017/05/05 20:09:38 Exit status: 0
```

第一版的调试器实现方式很简单。启动了一个被跟踪的进程，然后进程在执行第一条指令前停止，并向父进程发送了一个 signal。父进程等待这个 signal，打出日志 `log.Printf("State: %v\n", err)`。之后程序重启，父进程等待其终止。这种方式可以让我们有机会提前设置断点，启动程序，等一会到达指定跟踪点，看看类似堆栈或注册表里的当前值，检查下进程状态。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/making-debugger/1_qysN9I7NtfurG2K7kT_1Sw.jpeg)

哪怕知道一点点，我们也可以做一些很赞的事情。这些都会为今后的提高和实践（不久的将来）奠定基础。

给我们点个赞吧，让更多人能看到这篇文章。如果你想获得新博文的更新或者在工作中获得提高，请关注我们吧。

----------------

via: https://medium.com/golangspec/making-debugger-for-golang-part-i-53124284b7c8

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[ArisAries](https://github.com/ArisAries)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出