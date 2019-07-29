首发于：https://studygolang.com/articles/22291

# Go 内存管理之二

## 概述

之前在 [povilasv.me](https://povilasv.me/) 上，我们一起探讨了 [GO 内存管理](https://povilasv.me/go-memory-management/) [GCTT 译文](https://studygolang.com/articles/14956)，并且留下了两个小的 Go 程序，它们运行时分配的虚拟内存大小显著不同。

首先，我们一起来看一下占用很多虚拟内存的程序 `ex1`。它的代码如下：

```go
func main() {
	http.HandleFunc("/bar",
		func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q",
			HTML.EscapeString(r.URL.Path))
	})

	http.ListenAndServe(":8080", nil)
}
```

我执行了 `ps` 命令来查看虚拟内存大小，以下是它的输出。注意，输出中的内存大小单位是千字节（KiB），388496 KiB 约等于 379.390625 MiB。

```bash
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT
povilasv 16609  0.0  0.0 388496  5236 pts/9    Sl+
```

接下来，我们看一下只占用少量虚拟内存的程序 `ex2`：

```go
func main() {
	go func() {
		for {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			log.Println(float64(m.Sys) / 1024 / 1024)
			log.Println(float64(m.HeapAlloc) / 1024 / 1024)
			time.Sleep(10 * time.Second)
		}
	}()

	fmt.Println("hello")
	time.Sleep(1 * time.Hour)
}
```

最后，我们看一下这个程序的 `ps` 命令的输出，你可以看到它运行时只占用了少量的虚拟内存：4900 KiB，约等于 4.79 MiB。

```bash
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT
povilasv  3642  0.0  0.0   4900   948 pts/10   Sl+
```

有一点需要说明，这些程序是使用较老的 Go 1.10 版编译的；如果使用新版本的 Go 编译的话，这些数字将会不同。比如拿 `ex1` 来说，使用 Go 1.11 版编译，占用的虚拟内存为 466 MiB，常驻内存为 3.22 MiB；改用 Go 1.12 版编译，则这两者分别为 100.37 MiB 和 1.44 MiB。

由此我们可以看出，HTTP 服务程序和简单的命令行程序之间的差异导致了运行时占用虚拟内存大小的差异。

## 灵光乍现

看到这些，我突然灵机一动，也许可以用 `strace` 来调查这个有趣的现象。先看一下 `strace` 的描述：

*[strace](https://strace.io/) 是一个 Linux 平台下的用于诊断、调试和学习目的的用户空间实用程序。它可用于监视和篡改进程与 Linux 内核之间的交互，包括系统调用、信号传递和进程状态变化。*

接下来要做的就是使用 `strace` 运行两个程序来比较操作系统的行为。`strace` 的使用非常简单，你只需要在你要执行的程序前面加上 `strace` 即可。以 `ex1` 为例，我们执行命令：

```bash
strace ./ex1
```

将会产生以下输出：

```bash
execve("./ex1", ["./ex1"], 0x7fffe12acd60 /* 97 vars */) = 0
brk(NULL)                               = 0x573000
access("/etc/ld.so.preload", R_OK)      = -1 ENOENT (No such file or directory)
openat(AT_FDCWD, "/usr/local/lib/tls/haswell/x86_64/libpthread.so.0", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
stat("/usr/local/lib/tls/haswell/x86_64", 0x7ffdaa923fa0) = -1 ENOENT (No such file or directory)
...
stat("/lib/x86_64", 0x7ffdaa923fa0)     = -1 ENOENT (No such file or directory)
openat(AT_FDCWD, "/lib/libpthread.so.0", O_RDONLY|O_CLOEXEC) = 3
read(3, "\177ELF\2\1\1\0\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0\340b\0\0\0\0\0\0"..., 832) = 832
fstat(3, {st_mode=S_IFREG|0755, st_size=146152, ...}) = 0
mmap(NULL, 8192, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x7fc8a8d11000
mmap(NULL, 2225248, PROT_READ|PROT_EXEC, MAP_PRIVATE|MAP_DENYWRITE, 3, 0) = 0x7fc8a88cd000
mprotect(0x7fc8a88e8000, 2093056, PROT_NONE) = 0
mmap(0x7fc8a8ae7000, 8192, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_FIXED|MAP_DENYWRITE, 3, 0x1a000) = 0x7fc8a8ae7000
mmap(0x7fc8a8ae9000, 13408, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_FIXED|MAP_ANONYMOUS, -1, 0) = 0x7fc8a8ae9000
close(3)                                = 0
openat(AT_FDCWD, "/usr/local/lib/libc.so.6", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)
openat(AT_FDCWD, "/lib/libc.so.6", O_RDONLY|O_CLOEXEC) = 3
read(3, "\177ELF\2\1\1\3\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0\0\34\2\0\0\0\0\0"..., 832) = 832
fstat(3, {st_mode=S_IFREG|0755, st_size=1857312, ...}) = 0
mmap(NULL, 3963464, PROT_READ|PROT_EXEC, MAP_PRIVATE|MAP_DENYWRITE, 3, 0) = 0x7fc8a8505000
mprotect(0x7fc8a86c3000, 2097152, PROT_NONE) = 0
mmap(0x7fc8a88c3000, 24576, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_FIXED|MAP_DENYWRITE, 3, 0x1be000) = 0x7fc8a88c3000
mmap(0x7fc8a88c9000, 14920, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_FIXED|MAP_ANONYMOUS, -1, 0) = 0x7fc8a88c9000
close(3)                                = 0
mmap(NULL, 12288, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x7fc8a8d0e000
arch_prctl(ARCH_SET_FS, 0x7fc8a8d0e740) = 0
mprotect(0x7fc8a88c3000, 16384, PROT_READ) = 0
mprotect(0x7fc8a8ae7000, 4096, PROT_READ) = 0
mprotect(0x7fc8a8d13000, 4096, PROT_READ) = 0
set_tid_address(0x7fc8a8d0ea10)         = 2109
set_robust_list(0x7fc8a8d0ea20, 24)     = 0
rt_sigaction(SIGRTMIN, {sa_handler=0x7fc8a88d2ca0, sa_mask=[], sa_flags=SA_RESTORER|SA_SIGINFO, sa_restorer=0x7fc8a88e1140}, NULL, 8) = 0
rt_sigaction(SIGRT_1, {sa_handler=0x7fc8a88d2d50, sa_mask=[], sa_flags=SA_RESTORER|SA_RESTART|SA_SIGINFO, sa_restorer=0x7fc8a88e1140}, NULL, 8) = 0
rt_sigprocmask(SIG_UNBLOCK, [RTMIN RT_1], NULL, 8) = 0
prlimit64(0, RLIMIT_STACK, NULL, {rlim_cur=8192*1024, rlim_max=RLIM64_INFINITY}) = 0
brk(NULL)                               = 0x573000
brk(0x594000)                           = 0x594000
sched_getaffinity(0, 8192, [0, 1, 2, 3]) = 8
mmap(0xc000000000, 65536, PROT_NONE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0xc000000000
munmap(0xc000000000, 65536)             = 0
mmap(NULL, 262144, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x7fc8a8cce000
mmap(0xc420000000, 1048576, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0xc420000000
mmap(0xc41fff8000, 32768, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0xc41fff8000
mmap(0xc000000000, 4096, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0xc000000000
mmap(NULL, 65536, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x7fc8a8cbe000
mmap(NULL, 65536, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x7fc8a8cae000
rt_sigprocmask(SIG_SETMASK, NULL, [], 8) = 0
sigaltstack(NULL, {ss_sp=NULL, ss_flags=SS_DISABLE, ss_size=0}) = 0
sigaltstack({ss_sp=0xc420002000, ss_flags=0, ss_size=32768}, NULL) = 0
rt_sigprocmask(SIG_SETMASK, [], NULL, 8) = 0
gettid()                                = 2109
...
```

类似的，对于 `ex2`，我们执行：

```bash
strace ./ex2
```

产生输出：

```bash
execve("./ex2", ["./ex2"], 0x7ffc2965ca40 /* 97 vars */) = 0
arch_prctl(ARCH_SET_FS, 0x5397b0)       = 0
sched_getaffinity(0, 8192, [0, 1, 2, 3]) = 8
mmap(0xc000000000, 65536, PROT_NONE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0xc000000000
munmap(0xc000000000, 65536)             = 0
mmap(NULL, 262144, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x7ff1c637b000
mmap(0xc420000000, 1048576, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0xc420000000
mmap(0xc41fff8000, 32768, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0xc41fff8000
mmap(0xc000000000, 4096, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0xc000000000
mmap(NULL, 65536, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x7ff1c636b000
mmap(NULL, 65536, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x7ff1c635b000
rt_sigprocmask(SIG_SETMASK, NULL, [], 8) = 0
sigaltstack(NULL, {ss_sp=NULL, ss_flags=SS_DISABLE, ss_size=0}) = 0
sigaltstack({ss_sp=0xc420002000, ss_flags=0, ss_size=32768}, NULL) = 0
rt_sigprocmask(SIG_SETMASK, [], NULL, 8) = 0
gettid()                                = 22982
```

实际的输出比这要长，为了可读性，我只选取了从开头到调用 `gettid()` 的部分。之所以选择到这一行，是因为它在两个程序的 `strace` 输出里都只出现了一次。

让我们来比较一下这两个输出。首先，`ex1` 的输出更长一些。`ex1` 寻找一些 `.so` 库文件并把它们加载到内存里。比如，下面是加载 `libpthread.so.0` 时产生的输出：

```bash
...
openat(AT_FDCWD, "/lib/libpthread.so.0", O_RDONLY|O_CLOEXEC) = 3
read(3, "\177ELF\2\1\1\0\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0\340b\0\0\0\0\0\0"..., 832) = 832
fstat(3, {st_mode=S_IFREG|0755, st_size=146152, ...}) = 0
mmap(NULL, 8192, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x7fc8a8d11000
mmap(NULL, 2225248, PROT_READ|PROT_EXEC, MAP_PRIVATE|MAP_DENYWRITE, 3, 0) = 0x7fc8a88cd000
mprotect(0x7fc8a88e8000, 2093056, PROT_NONE) = 0
mmap(0x7fc8a8ae7000, 8192, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_FIXED|MAP_DENYWRITE, 3, 0x1a000) = 0x7fc8a8ae7000
mmap(0x7fc8a8ae9000, 13408, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_FIXED|MAP_ANONYMOUS, -1, 0) = 0x7fc8a8ae9000
close(3)                                = 0
```

在这个例子里，我们可以看到文件先是被打开，然后读取到内存里，最后被关闭。在对文件做内存映射的时候，有一些内存区域被设置了 `PROTO_EXEC` 的标志，这样做是为了让我们的程序能够执行位于这些区域的代码。我们可以看到同样的事情出现在 `libc.so.6` 库文件上：

```bash
...
openat(AT_FDCWD, "/lib/libc.so.6", O_RDONLY|O_CLOEXEC) = 3
read(3, "\177ELF\2\1\1\3\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0\0\34\2\0\0\0\0\0"..., 832) = 832
fstat(3, {st_mode=S_IFREG|0755, st_size=1857312, ...}) = 0
mmap(NULL, 3963464, PROT_READ|PROT_EXEC, MAP_PRIVATE|MAP_DENYWRITE, 3, 0) = 0x7fc8a8505000
mprotect(0x7fc8a86c3000, 2097152, PROT_NONE) = 0
mmap(0x7fc8a88c3000, 24576, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_FIXED|MAP_DENYWRITE, 3, 0x1be000) = 0x7fc8a88c3000
mmap(0x7fc8a88c9000, 14920, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_FIXED|MAP_ANONYMOUS, -1, 0) = 0x7fc8a88c9000
close(3)                                = 0
```

加载完库文件之后，两个程序开始表现出相似的行为。它们映射了相同的内存区域，执行了相似的指令，直到 `gettid()` 这一行。

`ex1` 加载了 `libpthread` 和 `libc`，而 `ex2` 并没有。这有点意思。

`cgo` 该出场了。

## CGO

让我们一起来探究一下 `cgo` 是什么以及它是如何工作的。[godoc](https://golang.org/cmd/cgo/) 上是这样解释的：

*Cgo 让 `go` 程序包能够调用 `C` 代码。*

为了调用 `C` 代码，你需要添加一段特殊的注释并导入一个特殊的包：`C`。让我们一起来看一下下面这个小例子：

```go
package main

// #include
import "C"
import "fmt"

func main() {
	char := C.getchar()
	fmt.Printf("%T %#v", char, char)
}
```

这个程序引用了 `C` 标准库里的 `stdio.h` 头文件，接着调用了 `getchar()` 并打印其返回值。`getchar()` 从标准输入读入一个字符（一个 unsigned char）。我们来试一下：

```bash
go build
./ex3
```

执行这个程序的时候，它要求你输入一个字符，并简单地将其打印出来。下面是这个过程的一个例子：

```bash
a
main._Ctype_int 97
```

我们可以看到，它表现的就像一个普通的 Go 程序一样。有趣的是你可以像编译一段 Go 原生代码一样编译它，只是简单的执行一下 `go build`。我敢打赌，如果你事先没有看过这段代码，你可能一点都意识不到这其中的差别。

显而易见的是，cgo 还有许多有趣的特点。比如，如果你把几个 `.c` 和 `.h` 文件与 Go 原生代码放到同一个目录下面，`go build` 也会编译它们并将它们与你的 Go 原生代码链接在一起。

如果你想了解更多的话，我建议你阅读一下 [godoc](https://golang.org/cmd/cgo/) 和 [C? Go? Cgo!](https://blog.golang.org/c-go-cgo) 这篇博文。现在，让我们回到之前那个有趣的问题，为什么 `ex1` 使用了 cgo 而 `ex2` 没有？

## 探究差异

`ex1` 和 `ex2` 的差别在于前者导入了 `net/http`，而后者没有。使用 `grep` 在 `net/http` 包里搜索了一遍并没有发现任何使用 `C` 语句的迹象。但只要再往上一级，你就可以在 `net` 包里找到证据。

看一下 `net` 包里的文件：

- [net/cgo_unix.go](https://golang.org/src/net/cgo_unix.go)
- [net/cgo_linux.go](https://golang.org/src/net/cgo_linux.go)
- [net/cgo_stub.go](https://golang.org/src/net/cgo_stub.go)

例如，`net/cgo_linux.go` 包含以下代码：

```go
// +build !android,cgo,!netgo

package net

/*
#include <netdb.h>
*/
import "C"

// NOTE(rsc): In theory there are approximately balanced
// arguments for and against including AI_ADDRCONFIG
// in the flags (it includes IPv4 results only on IPv4 systems,
// and similarly for IPv6), but in practice setting it causes
// getaddrinfo to return the wrong canonical name on Linux.
// So definitely leave it out.
const cgoAddrInfoFlags = C.AI_CANONNAME | C.AI_V4MAPPED | C.AI_ALL
```

我们可以看到 net 包里引用了 `C` 头文件 `netdb.h` 并且使用了这个文件里的几个变量。为什么需要这些东西？让我们接着调查。

## 什么是 `netdb.h`

如果你查阅过 `netdb.h` 的说明文档，你就会发现它其实是 libc 的一部分。它的说明文档里这样写道：

*`netdb.h` 提供了网络数据库操作的定义。*

另外，文档里也对这里涉及的几个常量进行了说明。让我们来看一下：

- **AI_CANONNAME** - 请求规范名称
- **AI_V4MAPPED** - 如果没有 IPv6 地址，就查询 IPv4 地址并将它们映射成 IPv6 地址返回
- **AI_ALL** - 同时查询 IPv4 和 IPv6 地址

探寻一下这些标志是如何使用的，就会发现它们最终会被传递给 `getaddrinfo()`，一个使用 libc 来解析 DNS 域名的函数。简而言之，这些标志控制 DNS 域名解析如何发生。

同样地，如果你打开 [net/cgo_bsd.go](https://golang.org/src/net/cgo_bsd.go)，你会看到常量 `cgoAddrInfoFlags` 的一个略有差异的版本。一起来看一下：

```go
// +build cgo,!netgo
// +build darwin dragonfly freebsd

package net

/*
#include <netdb.h>
*/
import "C"

const cgoAddrInfoFlags = (C.AI_CANONNAME | C.AI_V4MAPPED |
 C.AI_ALL) & C.AI_MASK
```

这暗示我们，有一种机制可以为 DNS 解析设置操作系统特定的标志，而我们正在使用 cgo 正确地进行 DNS 查询。这真的很酷。让我们再深入一点探索 `net` 包。

读一读 [net](https://golang.org/pkg/net/#hdr-Name_Resolution) 包的文档：

*名称解析*

*指使用类似于 `Dial` 的函数或者类似于 `LookupHost` 和 `LookupAddr` 的函数进行间接地或直接地解析域名的方法，具体的函数随操作系统不同而不同。*

*在 Unix 系统上，解析器解析名称的时候有两种选择。一种是使用纯粹的 Go 解析器直接向列在 `/etc/resolv.conf` 文件里的服务器发送 DNS 请求，另一种是使用基于 cgo 的解析器通过调用 C 库函数，比如 `getaddrinfo` 和 `getnameinfo`，来实现。*

*默认情况下，使用纯 Go 解析器进行解析，这是因为一个阻塞的 DNS 请求只需要消耗一个 Go 例程；而一个阻塞的 C 函数调用却要占用一个系统线程。如果 cgo 可用的话，在很多情况下都需要使用基于 cgo 的解析器：在不允许程序直接发送 DNS 请求的系统上（比如 OS X）；当 `LOCALDOMAIN` 环境变量被定义时（即使是个空值）；当 `RES_OPTIONS` 和 `HOSTALIAS` 环境变量非空时；当 `ASR_CONFIG` 环境变量非空时（仅 OpenBSD 系统）；当 `/etc/resolv.conf` 或 `/etc/nsswitch.conf` 里面使用了 Go 解析器没有实现的特性时；当被查询的名字以 `.local` 结尾或者是一个 mDNS 名字。*

*你还可以在 `GODEBUG` 环境变量（详见 runtime 包）里为 netdns 指定 Go 或 cgo 来强制指定使用对应的解析器，就像下面那样：*

```bash
export GODEBUG=netdns=go    # 强制使用纯 Go 解析器
export GODEBUG=netdns=cgo   # 强制使用 cgo 解析器
```

*你也可以通过在构建 Go 源码树时设置 `netgo` 或 `netcgo` 构建标志来强制选择对应的解析器。*

*如果给 `netdns` 指定一个数字，比如这样 `GODEBUG=netdns=1`，解析器就会打印它所选择的解析方式。*

-- https://golang.org/pkg/net/#hdr-Name_Resolution

文档已经读的够多了。下面，就让我们一起来尝试使用不同的 DNS 客户端实现吧。

## 使用构建标签

正如文档里描述的那样，我们可以使用环境变量来指定 DNS 客户端实现。这种方式很灵活，因为你不需要重新编译代码就可以在两种方式之间自由切换。

另外，从代码里来看，我发现我们也可以使用 Go 构建标签在编译时指定解析方式。除此之外，我们还可以通过设置 `GODEBUG=netdns=1` 环境变量并做一次真实的 DNS 查询来查看到底使用了哪种方式。

看了 `net` 包里的源文件，我发现一共有 3 种构建模式。它们都可以通过使用不同的构建标签来指定。这三种构建模式分别是：

1. `!cgo` -- 不使用 cgo，也就是说强制使用 Go 版本的解析器
2. `netcgo` 或 `cgo` -- 使用 libc 的 DNS 解析方式
3. `netgo + cgo` -- 使用 Go 原生的 DNS 解析方式，同时我们还可以包含 C 代码

让我们一起尝试所有这些组合来看看结果如何。

由于之前的程序不会发起 DNS 查询，我们需要编写新的程序。下面就是我们要用的代码：

```go
func main() {
	addr, err := net.LookupHost("povilasv.me")
	fmt.Println(addr, err)
}
```

然后，执行构建：

```bash
export CGO_ENABLED=0
export GODEBUG=netdns=1
go build -tags netgo
```

运行程序：

```bash
./testnetgo
```

程序输出：

```bash
go package net: built with netgo build tag; using Go's DNS resolver
104.28.1.75 104.28.0.75 2606:4700:30::681c:4b 2606:4700:30::681c:14b <nil>
```

现在让我们使用 libc 的解析器来重新构建：

```bash
export GODEBUG=netdns=1
go build -tags netcgo
```

运行程序：

```bash
./testnetgo
```

程序输出：

```bash
go package net: using cgo DNS resolver
104.28.0.75 104.28.1.75 2606:4700:30::681c:14b 2606:4700:30::681c:4b <nil>
```

最后，我们来使用 `netgo cgo` 进行构建：

```bash
export GODEBUG=netdns=1
go build -tags 'netgo cgo' .
```

运行程序：

```bash
./testnetgo
```

输出：

```bash
go package net: built with netgo build tag; using Go's DNS resolver
104.28.0.75 104.28.1.75 2606:4700:30::681c:14b 2606:4700:30::681c:4b <nil>
```

可以看到，构建标签真的起了作用。现在，让我们回到虚拟内存的问题上来。

## 回到虚拟内存

现在，我想分别使用这 3 组标志来重新构建我们那个简单的 HTTP 网页服务器 `ex1`，看看它们会对虚拟内存产生怎样的影响。

使用 `netgo` 模式编译：

```bash
export CGO_ENABLED=0
go build -tags netgo
```

`ps` 输出：

```bash
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT
povilasv  3524  0.0  0.0   7216  4076 pts/17   Sl+
```

可以看到在这种模式下虚拟内存的占用是很低的。

现在来看看 `netcgo` 的情况：

```bash
go build -tags netcgo
```

`ps` 输出：

```bash
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT
povilasv  6361  0.0  0.0 382296  4988 pts/17   Sl+
```

可以看到在这种模式下，占用了大量虚拟内存（382296 KiB）。

最后，我们来看看 `netgo cgo` 模式：

```bash
go build -tags 'netgo cgo' .
```

`ps` 输出：

```bash
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT
povilasv  8175  0.0  0.0   7216  3968 pts/17   Sl+
```

可以看到在这种模式下虚拟内存的占用也是是很低的（7216 KiB）。

可以肯定的是，`netgo` 模式下不会占用很多虚拟内存。从另一方面来讲，我们还不能将虚拟内存消耗过多的责任归咎于 cgo，因为 `ex1` 程序里并没有包含任何 C 代码，`netgo cgo` 模式实际上和 `netgo` 模式一样，会跳过编译和链接 C 文件这一整套 cgo 的工作流程。

因而，我们还需要加入额外的 C 代码再来分别尝试 `netcgo` 和 `netgo cgo` 两种模式。这可以让我们弄清楚，在 cgo 模式下启用和禁用 libc 的 DNS 客户端，程序分别会有怎样的表现。

我们来尝试一下这段代码：

```go
package main

// #include
import "C"
import "fmt"

func main() {
	char := C.getchar()
	fmt.Printf("%T %#v", char, char)

	http.HandleFunc("/bar",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, %q",
				HTML.EscapeString(r.URL.Path))
	})

	http.ListenAndServe(":8080", nil)
}
```

可以看到，这段代码应该能够达到我们的目的。因为它既使用了 cgo，也能够根据构建标签来选择使用 netgo 还是 libc 的 DNS 客户端实现。

让我们试一试：

```bash
go build -tags netcgo .
```

`ps` 输出：

```bash
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT
povilasv 12594  0.0  0.0 382208  4824 pts/17   Sl+
```

可以看到虚拟内存占用没有变化。现在来试一下 `netgo cgo`：

```bash
go build -tags 'netgo cgo' .
```

`ps` 输出：

```bash
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT
povilasv  1026  0.0  0.0 382208  4824 pts/17   Sl+
```

最后，终于可以排除 libc 的 DNS 客户端实现的影响，因为禁用它并没有带来任何变化。我们清楚的看到这一切都跟 cgo 有关。

为了深入探索这个问题，我们先来简化一下我们的程序。`ex1` 启动了一个 HTTP 服务器，调试一个这样的程序远比调试一个简单的命令行程序困难的多。看一下这段代码：

```go
package main

// #include
// #include
import "C"
import (
	"time"
	"unsafe"
)

func main() {
	cs := C.CString("Hello from stdio")
	C.puts(cs)

	time.Sleep(1 * time.Second)

	C.free(unsafe.Pointer(cs))
}
```

运行一下并查看一下内存占用：

```bash
go build .
./ex6
```

`ps` 输出：

```bash
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT
povilasv 15972  0.0  0.0 378228  2476 pts/17   Sl+
```

酷！它真的占用了许多虚拟内存，我们真的需要调查 cgo 了。

想看深入调查，请阅读 [Go 内存管理之三](https://povilasv.me/go-memory-management-part-3/)。

这就是今天的内容。如果你想第一时间看到我的博客文章，请订阅[简报](https://povilasv.me/newsletter)。如果你愿意支持我的写作，我这里还有一个[愿望清单](https://www.amazon.com/hz/wishlist/ls/2NLKE1Z1SND3W?ref_=wl_share)，你可以为我买一本书或是随便一个什么东西😉。

感谢您的阅读，下次再见！

---

via: https://povilasv.me/go-memory-management-part-2/

作者：[Povilas](https://povilasv.me/about/)
译者：[Stonelgh](https://github.com/stonglgh)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
