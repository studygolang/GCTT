首发于：https://studygolang.com/articles/23460

# Go 内存管理之三

之前在 [povilasv.me](https://povilasv.me/) 上，我们一起探讨了 [Go 内存管理](https://studygolang.com/articles/14956) 和 [Go 内存管理之二](https://studygolang.com/articles/22291)。在上篇博文中，我们发现使用 cgo 会占用更多的虚拟内存。现在我们来深入研究一下 cgo。

## CGO 揭秘

正如之前所见，cgo 会使虚拟内存膨胀。此外，对于大部分用户而言，一旦他们导入了 net 包或者其子包（比如 http），就会自动的使用 cgo。

我在标准库的代码里发现很多描述 cgo 调用工作机制的文档。比如，你在 [cgocall.go](https://golang.org/src/runtime/cgocall.go) 文件里，就能看到非常有用的注释：

*为了在 Go 代码中调用 C 函数 `f`，cgo 会生成代码调用 `runtime.cgocall(_cgo_Cfunc_f, frame)`，这里 `_cgo_Cfunc_f` 是由 cgo 自动生成、gcc 编译的函数。*

*为了不阻塞其他的 Go 例程或垃圾收集器，`runtime.cgocall`（如下）先调用 `entersyscall`，接着再调用 `runtime.asmcgocall(_cgo_Cfunc_f, frame)`。*

*`runtime.asmcgocall`（见 asm_$GOARCH.s）切换到 `m->go` 栈上（被认为是由操作系统分配的栈，可以在其上安全的运行 gcc 编译的代码），并且调用 `_cgo_Cfunc_f(frame)`。*

*`_cgo_Cfunc_f` 调用真正的 C 函数 `f`，并将 `frame` 结构里的参数传递给它，将结果记录到 `frame` 结构里，最后返回 `runtime.asmcgocall`。*

*`runtime.asmcgocall` 重获控制后，再切回原来的 `g`（`m->curg`）的栈，然后返回 `runtime.cgocall`。*

*然后， `runtime.cgocall` 会调用 `exitsyscall`，它会阻塞直到 `m` 可以运行 Go 代码而不会违反 `$GOMAXPROCS` 限制，接着将 `g` 从 `m` 上解锁。*

*以上的描述跳过了由 gcc 编译的函数 `f` 再调用 Go 函数的可能性。如果发生这种情况，我们会在执行 `f` 的过程中沿着兔子洞走下去。*

*-- cgocall.go 源码（https://golang.org/src/runtime/cgocall.go ）*

注释甚至还深入的讲解了 Go 如何实现 cgo 到 Go 的调用。我强烈建议你研究一下这些代码和注释。透过现象看本质让我学到了很多。从这些注释我们可以看到，是否调用 C 代码会让 Go 程序的行为完全不同。

## 运行时追踪

探索 Go 程序行为的一种很酷的方法是使用 Go 运行时追踪。要想进一步了解 Go 程序追踪方面的知识，可以参阅 [Go 运行时追踪器](https://blog.gopheracademy.com/advent-2017/go-execution-tracer/)这篇博文。现在，我们来修改一下代码加入追踪功能：

```go
func main() {
    trace.Start(os.Stderr)
    cs := C.CString("Hello from stdio")

    time.Sleep(10 * time.Second)

    C.puts(cs)
    C.free(unsafe.Pointer(cs))

    trace.Stop()
}
```

编译这个程序并且将标准错误输出保存到文件里：

```bash
/ex7 2> trace.out
```

最后，来看看追踪结果：

```bash
go tool trace trace.out
```

就是这样。下次如果再遇到表现怪异的命令行程序，我就知道如何去追踪了🙂。顺便说一下，如果要追踪网页服务的话，可以使用 `httptrace` 包，用起来更简单。想要了解更多的话，可以参阅 [HTTP 追踪](https://blog.golang.org/http-tracing)这篇博文。

我还编写了一个相似的但没有任何 C 代码调用的程序，以便使用 `go tool trace` 比较它们的追踪结果。这就是这个 Go 原生程序的代码：

```go
func main() {
   trace.Start(os.Stderr)
   str := "Hello from stdio"
   time.Sleep(10 * time.Second)
   fmt.Println(str)

   trace.Stop()
}
```

cgo 程序和 Go 原生程序的追踪结果并没有太大差异。当然，我注意到有一些统计还是有点差别的。比如，cgo 程序没有包含堆的统计信息。

![cgo 程序的追踪信息统计](https://github.com/studygolang/gctt-images/blob/master/go-memory-management-part-3/cgo.png?raw=true)<br>cgo 程序的追踪信息统计

![Go 原生程序的追踪信息统计](https://github.com/studygolang/gctt-images/blob/master/go-memory-management-part-3/noncgo.png?raw=true)<br>Go 原生程序的追踪信息统计

我试图使用不同的视图去观察，但是并没有发现更多的显著差异。我猜测这可能是由于 Go 不会为已编译的 C 代码添加追踪指令。

因此，我决定使用 `strace` 来继续探索它们之间的差异。

## 使用 `strace` 探索 cgo

首先明确一下，我们将要探索的两个程序有着相同的行为。我们只是简单的从上述的程序中去除了追踪语句。

cgo 程序：

```go
func main() {
    cs := C.CString("Hello from stdio")

    time.Sleep(10 * time.Second)

    C.puts(cs)
    C.free(unsafe.Pointer(cs))
}
```

go 原生程序：

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    str := "Hello from stdio"
    time.Sleep(10 * time.Second)

    fmt.Println(str)
}
```

编译，然后使用 `strace` 运行它们：

```bash
sudo strace -f ./program_name
```

我加了 `-f` 标志让 strace 也追踪线程。

*`-f` 追踪当前被追踪进程由于执行 fork（2），vfork（2）和 clone（2） 系统调用而生成的子进程*

### cgo 结果

如前所见，为了完成工作，cgo 程序会加载 libc 和 pthreads 这两个 C 代码库。而且，事实证明，cgo 程序以不同的方式创建线程。创建线程的时候，你可以看到有一个函数调用为线程栈分配了 8 MB 的内存：

```bash
mmap(NULL, 8392704, PROT_NONE, MAP_PRIVATE|MAP_ANONYMOUS|MAP_STACK, -1, 0) = 0x7f1990629000
// 我们为栈分配了 8 MB 内存
mprotect(0x7f199062a000, 8388608, PROT_READ|PROT_WRITE) = 0
// 允许读写，但禁止执行位于该内存区域的代码
```

设置完栈之后，你会看到一个系统调用 `clone`，但是传递的参数与一个典型的 Go 原生程序不同：

```bash
clone( child_stack=0x7f1990e28fb0, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM|CLONE_SETTLS|CLONE_PARENT_SETTID|CLONE_CHILD_CLEARTID, parent_tidptr=0x7f1990e299d0, tls=0x7f1990e29700, child_tidptr=0x7f1990e299d0) = 3600
```

如果你对这些参数的含义感兴趣，请参阅下面的描述（摘自 `clone` 手册）：

*`CLONE_VM` – 调用进程和子进程运行于同一内存空间。*<br>
*`CLONE_FS` – 调用者和子进程共享相同的文件系统信息。*<br>
*`CLONE_FILES` – 调用进程和子进程共享相同的文件描述符表。*<br>
*`CLONE_SIGHAND` – 调用进程和子进程共享相同的信号处理函数表。*[<sup> 译注 1</sup>](#note1)<br>
*`CLONE_THREAD` – 把子进程与调用进程置于同一个线程组中。*<br>
*`CLONE_SYSVSEM` – 子进程和调用进程共享同一个 System V 信号量调整值的列表。*<br>
*`CLONE_SETTLS` – 将 TLS （线程本地存储）描述符设置成 `nettls`。*<br>
*`CLONE_PARENT_SETTID` – 将子线程 ID 存储在 `ptid` 在父线程内存中的位置。*<br>
*`CLONE_CHILD_CLEARTID` – 将子线程 ID 存储在 `ctid` 在子线程内存中的位置。*<br>

*–- `clone` 系统调用手册*

`clone` 调用之后，线程会首先保留 128 MB 内存，然后再取消保留 57.8 MB 和 8 MB。看一下下面这段 `strace` 的输出：

```bash
mmap(NULL, 134217728, PROT_NONE, MAP_PRIVATE|MAP_ANONYMOUS|MAP_NORESERVE, -1, 0) = 0x7f1988629000
//134217728 / 1024 / 1024 = 128 MiB
munmap(0x7f1988629000, 60649472 )
// 取消从 0x7f1988629000 起始的 57.8 MiB 的内存映射
munmap(0x7f1990000000, 6459392)
// 取消从 0x7f1990000000 起始的 8 MiB 的内存映射
mprotect(0x7f198c000000, 135168, PROT_READ|PROT_WRITE
```

现在，一切都能说通了。在 cgo 程序中，我们看到分配了大约 373.25 MB 的虚拟内存，这可以从上面的输出中获得完全的解释。甚至，它还解释了为什么我在[本文第一部分](https://povilasv.me/go-memory-management/)中的 `/proc/PID/maps` 里面没有看到这些内存映射，因为保留内存的线程有自己的 PID。另外，虽然线程调用了 `mmap`，由于并没有实际的使用那些内存区域，因而它们并不会被算到常驻内存里，而是被算到了虚拟内存里。

让我们来做一些随手计算：

strace 的输出中有 5 个 `clone` 系统调用。各自保留了 8 MB（栈）+ 128 MB，接着又取消保留 57.8 MB 和 8 MB，这样最终每个线程保留了约 70 MB。但实际上，有一个线程并没有取消映射任何内存，还有一个没有取消映射那 8 MB。所以，最终的算式应该像下面这样：

```
4 * 70 + 8 + 1 * 128 = ~ 416 MB
```

此外，不要忘了程序初始化的时候会额外保留一些内存，因而应该再加上一些常量。

显然，要想弄清楚我们到底是在哪个时刻对内存做的采样（执行 `ps` 命令）是非常困难的；比如，我们可能在只有 2 或 3 个线程在运行的时候执行的 `ps`，内存已经被 `mmap` 但还没有被释放等。在我看来，这就是我最初写作 [Go 内存管理](https://povilasv.me/go-memory-management/)这篇博文想要寻找的解答。

如果你对 `mmap` 的参数含义感兴趣，这里是它们的定义：

- *MAP_ANONYMOUS – 不映射到任何文件，将内存初始化成 0。*
- *MAP_NORESERVE – 不为这个映射保留交换空间。*
- *MAP_PRIVATE – 创建一个私有的写时复制映射。对于映射了同一个文件的多个进程而言，一个进程对内存的更新对其他进程不可见，并且不会写回文件。*

*mmap 系统调用手册*

最后，我们来看看 Go 原生程序是怎样创建线程的。

### Go 原生结果

go 原生程序只会进行 4 次 `clone` 系统调用，新线程不会分配内存（没有 `mmap` 调用），也不为栈保留 8 MB 的内存空间。go 原生程序创建线程的调用大致如下：

```bash
clone( child_stack=0xc420042000, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM) = 3935
```

注意 Go 和 cgo 调用 `clone` 时传递参数的差异。

此外，在 Go 原生代码产生的 strace 输出里，你可以清楚的看到 `Println` 语句对应的系统调用：

```bash
[pid  3934] write(1, "Hello from stdio\n", 17Hello from stdio
) = 17
```

[<sup> 译注 2</sup>](#note2)

然而，在 cgo 版本里，我没有找到 `fputs()` 对应的类似系统调用。

Go 原生程序让我喜爱的一点是，它的 strace 输出更小、更易于理解。这意味着发生了更少的事情。比如，go 原生程序的 strace 输出仅有 281 行，而 cgo 版的输出有 342 行。

## 结语

你可以从我的探索中得到的有：

- 如果使用的包里引用了 C 代码，Go 可能自动神奇的切换到 cgo。比如使用了 `net`、`net/http` 包。
- Go 有两个 DNS 解析器的实现：`netgo` 版和 `netcgo` 版。
- 通过设置环境变量 `export GODEBUG=netdns=1`，你可以了解你在使用哪个 DNS 客户端。
- 你可以在运行时切换 DNS 解析器，只需将环境变量设置成 `export GODEBUG=netdns=go` 或 `export GODEBUG=netdns=cgo`。[<sup> 译注 3</sup>](#note3)
- 你可以通过 Go 构建标签在编译时指定使用的 DNS 解析实现：`go build -tags netcgo` 或 `go build -tags netgo`。
- `/proc` 文件系统很有用，但是不要被线程误导！
- `/prod/PID/status` 和 `/proc/PID/maps` 对于快速了解正在发生什么会很有用。
- Go 运行时追踪器能够帮助你调试程序。
- 当你不知道如何下手时，不要忘了还有 `strace -f`。

最后：

- cgo 不是 Go。
- 大的虚拟内存占用不是坏事。
- 不同版本的 Go 编译出来的程序会有不同的表现，Go 1.10 上正确的事情并不一定在 Go 1.12 上同样正确。

如果这篇博文让你有所收获，希望不仅仅是在编译 Go 程序的时候需要设置 `CGO_ENABLED=0`。Go 的作者不是凭空决定了 Go 现在的运行机制。你现在看到的这些行为在将来也许会改变，就像 Go 1.12 带来的变化那样。

这就是今天的内容。如果你想第一时间看到我的博客文章，请订阅[简报](https://povilasv.me/newsletter)。如果你愿意支持我的写作，我这里还有一个[愿望清单](https://www.amazon.com/hz/wishlist/ls/2NLKE1Z1SND3W?ref_=wl_share)，你可以为我买一本书或是随便一个什么东西😉。

感谢您的阅读，下次再见！

## 译注

1. <a name="note1"></a> 此处原文中对 `CLONE_SIGHAND` 的说明实为 `CLONE_FILES` 的说明，应为拷贝粘贴错误。译文已纠正。
2. <a name="note2"></a> 行末的 `Hello from stdio` 和换行符应为程序运行时打印到标准输出所致。
3. <a name="note3"></a> 此处原文中两处皆为：`export GODEBUG=netdns=go`，应为笔误。译文已纠正。

---

via: https://povilasv.me/go-memory-management-part-3/

作者：[Povilas](https://povilasv.me/about/)
译者：[Stonelgh](https://github.com/stonglgh)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
