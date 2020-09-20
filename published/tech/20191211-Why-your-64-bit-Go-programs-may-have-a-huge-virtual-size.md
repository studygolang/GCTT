首发于：https://studygolang.com/articles/27150

# 为什么你的 64-bit 程序可能占用巨大的虚拟空间

出于很多目的，我从最新的 Go 系统内核开发源码复制了一份代码，在一个正常的运行环境中构建（和重新构建）它，在构建版本基础上周期性地重新构建 Go 程序。近期我在用 `ps` 查看我的[一个程序](https://github.com/siebenmann/smtpd/)的内存使用情况时，发现它占用了约 138 GB 的巨大虚拟空间（Linux ps 命令结果的 `VSZ` 字段），尽管它的常驻内存还不是很大。某个进程的常驻内存很小，但是需要内存很大，通常是表示有内存泄露，因此我心里一颤。

（用之前版本的 Go 构建后，根据运行时间长短不同，通常会有 32 到 128 MB 不同大小的虚拟内存占用，比最新版本小很多。）

还好这不是内存泄漏。事实上，之后的实验表明即使是个简单的 `hello world` 程序也会有占用很大的虚拟内存。通过查看进程的 `/proc/<pid>/smaps` 文件（[cf](https://utcc.utoronto.ca/~cks/space/blog/linux/SmapsFields)）可以发现几乎所有的虚拟空间是由两个不可访问的 map 占用的，一个占用了约 8 GB，另一个约 128 GB。这些 map 没有可访问权限（它们取消了读、写和可执行权限），所以它们的全部工作就是专门为地址空间预留的（甚至没有用任何实际的 RAM）。大量的地址空间。

这就是现在的 Go 在 64 位系统上的低级内存管理的工作机制。简而言之，Go （理论上）从连续的 arena 区域上进行低级内存分配，申请 8 KB 的页；哪些页可以无限申请存储在一个巨大的 bitmap。在 64 位机器上，Go 会把全部的内存地址空间预留给 bitmap 和 arena 区域本身。程序运行时，当你的 Go 程序真正使用内存时，arena bitmap 和内存 arena 片段会从简单的预留地址空间变为由 RAM 备份的内存，供其他部分使用。

（bitmap 和 arena 通常是通过给 `mmap` 传入 `PROT_NONE` 参数进行初始化的。当内存被使用时，会使用 `PROT_READ|PROT_WRITE` 重新映射。当释放时，我不确定它做了什么，所以对此我不发表意见。）

这个例子是用当前发布的 Go 1.4 开发版本复现的。之前的版本的 64 位程序运行时会占用更小的需要空间，虽然读 Go 1.4 源码时我也没找到原因。

以我的理解，一个有意思的影响是 64 位 Go 程序的大部分内存分配都可能占用至多 128 GB 的空间（也可能在整个运行周期内所有的内存分配都会，我不确定）。

了解更多细节，请看 [src/runtime/malloc2.go](https://github.com/golang/go/blob/master/src/runtime/malloc2.go) 的注释和 [src/runtime/malloc1.go](https://github.com/golang/go/blob/master/src/runtime/malloc1.go) 的 `mallocinit()`。

我不得不说，这个比我最初以为地更有意思也更有教育意义，尽管这意味着查看 `ps` 不再是一个检测你的 Go 程序中内存泄露的好方法（温馨提示，我不确定它曾经是不是）。结论是，检测这类内存使用最好的方法是同时使用 `runtime.ReadMemStats()`（可以通过 [net/http/pprof](http://golang.org/pkg/net/http/pprof/) 暴露出去）和 Linux 的 `smem` 程序或者养成对有意义的内存地址空间占用生成详细信息的习惯。

PS: Unix 通常足够智能，可以理解 `PROT_NONE` 映射不会耗尽内存，因此不应该对系统内存过量使用的限制进行统计。然而，它们会统计每一个进程的总地址空间进行统计，这意味着你运行 1.4 的 Go 程序时不能真的使用这么多。由于总内存地址空间的最大数几乎不会达到，因此这似乎不是一个问题。

## 附录：在 32 位系统上是怎样的

所有的信息都在 `mallocinit()` 注释中。简而言之，就是运行时预留了足够大的 arena 来处理 2 GB 的内存（「仅」占用 256 MB）但是仅预留 2 GB 中理论上它可以使用的 512 MB 地址空间。如果后续的运行过程中需要更多内存，就向操作系统申请另一个块的地址空间，优先 arena 区域剩下的 1.5 GB 的地址空间中分配。大多数情况下，运行的程序都会正常申请到需要分配的空间。

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoBigVirtualSize

作者：[ChrisSiebenmann](https://utcc.utoronto.ca/~cks/space/People/ChrisSiebenmann)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
