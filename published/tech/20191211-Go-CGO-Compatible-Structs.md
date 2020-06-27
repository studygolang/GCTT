首发于：https://studygolang.com/articles/27148

# 用 cgo 生成用于 cgo 的 C 兼容的结构体

假设（[并非完全假设，这里有 demo](https://github.com/siebenmann/go-kstat/)）你正在编写一个程序包，用于连接 Go 和其它一些提供大量 C 结构体内存的程序。这些结构可能是系统调用的结果，也可能是一个库给你提供的纯粹信息性内容。无论哪种情况，你都希望将这些结构传递给你的程序包的用户，以便他们可以使用这些结构执行操作。在你的包中，你可以直接使用 cgo 提供的 C.<whatever> 类型。但这有点恼人（这些整型它们没有对应的原生 Go 类型，使得与常规 Go 代码交互需要乱七八糟的强制转换），并且对于其它导入你的包的代码没有帮助。因此，你需要以某种方式使用原生的 Go 结构体。

一种方式是手动为这些 C 结构体的定义你自己的 Go 版本。这有两个缺点。这太枯燥了（还很容易出错），并且不能保证你能获得与 C 完全相同的内存布局（后者通常但并非总是很重要）。幸运的是有一种更好的方法，那就是使用 cgo 的 `-godefs` 功能或多或少地为你自动生成结构体声明。生成结果并不总是完美的，但可能会为你带来最大的收益。

使用 `-godefs` 的起点是特殊的 cgo Go 源文件，该文件需要将某些 Go 类型声明为某些 C 类型。例如：

```go
// +build ignore
package kstat
// #include <kstat.h>
import "C"

type IO C.kstat_io_t
type Sysinfo C.sysinfo_t

const Sizeof_IO = C.sizeof_kstat_io_t
const Sizeof_SI = C.sizeof_sysinfo_t
```

这些常量对于喜欢较真的人很有用，可以用来在后面对比检查 Go 类型的 `unsafe.Sizeof()` 和 C 类型的大小是否一致。

运行 `go tool cgo -godefs <file>.go` ，它将打印一系列带有导出字段和所有内容的标准 Go 类型到标准输出。然后，你可以将其保存到文件中并使用。如果你认为 C 类型可能会更改，则应将生成的文件保留下来，这样就避免重新生成文件遇到的很多麻烦。如果 C 类型基本上是固定的，则可以使用 godoc 对生成的输出进行注释。 cgo 会考虑类型匹配问题，它会把原始的 C 结构中存在的 padding 也插入到输出中。

我不知道如果原始的 C 结构体不可能在 Go 中重建出来，cgo 会怎么办。 比如 Go 需要 padding，但是 C 不需要。希望它会指出错误。这是你以后可能要检查这些 sizeof 的原因之一。

`-godefs` 最大的限制是与 cgo 通常具有的限制相同：它没有对 C 联合类型（union）的真正支持，因为 Go 确实没有这个。如果你的 C 结构体中有联合，你得自己弄清楚如何处理它们；我相信 cgo 把这些转换为大小合适的 uint8 数组，但这对于实际访问内容不是很有用。

这里有两个问题。假设你有一个嵌入了另一个结构体类型的结构体：

```c
struct cpu_stat {
   struct cpu_sysinfo cpu_sysinfo;
   struct cpu_syswait cpu_syswait;
   struct vminfo cpu_vminfo;
}
```

在这里，你必须给 cgo 一些帮助，方式是在主结构体类型之前创建嵌入结构类型的 Go 版本：

```go
type Sysinfo C.struct_cpu_sysinfo
type Syswait C.struct_cpu_syswait
type Vminfo  C.struct_cpu_vminfo

type CpuStat C.struct_cpu_stat
```

然后 cgo 才能生成正确的内嵌的 Go 结构的 CpuStat 结构。如果不这样做，你将获得一个 CpuStat 结构类型，该结构类型具有不完整的类型信息，其中的 `Sysinfo` 等字段将引用名为 `_Ctype_…` 的未在任何地方定义的类型。

顺便说一句，我在这确实是指 `Sysinfo` ，而不是 `Cpu_sysinfo` 。cgo 足够聪明，可以从结构字段名称中删除这种常见的前缀。我不知道它的算法是怎样的，但至少是有用的。

第二个问题是嵌入了匿名结构：

```c
struct mntinfo_kstat {
   ....
   struct {
      uint32_t srtt;
      uint32_t deviate;
   } m_timers[4];
   ....
}
```

不幸的是，cgo 根本无法处理这种问题。具体可以去看 [issue 5253](https://github.com/golang/go/issues/5253) ，你有两个选择，第一种是使用[建议的 CL 修复](https://codereview.appspot.com/122900043)，这个目前仍然适用于 src/cmd/cgo/gcc.go 并且能够工作（对我来说）。如果你不想构建自己的 Go 工具链（或者如果 CL 不再适用或无法工作），另一种解决方案是创建一个新的 C 头文件，该文件具有整个结构体的变体，通过创建具名结构体去除结构体的匿名化。

```c
struct m_timer {
   uint32_t srtt;
   uint32_t deviate；
}

struct mntinfo_kstat_cgo {
   ....
   struct m_timer m_timers [4];
   ....
}
```

然后，在你的 Go 文件中，

```go
...
// #include "myhacked.h"
...

type MTimer C.struct_m_timer
type Mntinfo C.struct_mntinfo_kstat_cgo
```

除非你搞错了，否则两个 C 结构体应具有完全相同的大小和布局，并且彼此完全兼容。现在你可以在你的版本上使用 `-godefs` 了，记住按照前面问题 1 的处理，需要为 `m_timer` 创建明确的 Go 类型。
如果你飘了（你认为你不在需要重新生成这些内容了），你可以在生成的 Go 文件中逆转这个过程，重新将 MTimer 类型匿名化到结构体中（因为 Go 对匿名结构体有很好的支持）。因为你没有更改实际内容，只是改了类型声明，所以结果应该与原始的布局相同。

PS：`-godefs` 的输入文件被设置为不被正常 `go build` 过程构建，因为它只用于 godefs 生成。如果这个文件也被包含在 `go build` 构建的源码中，你会得到关于 Go 类型多处定义的构建错误。必然的结果是，你不需要将此文件和任何相关 .h 文件与软件包的常规 .go 文件放在同一目录。你可以把他们放在子目录，或者放在完全独立的位置。

（我认为该 `package` 行在 godefs .go 文件中唯一要做的就是设置 cgo 将在输出中打印的软件包名称。）

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoCGoCompatibleStructs

作者：[ChrisSiebenmann](https://utcc.utoronto.ca/~cks/space/People/ChrisSiebenmann)
译者：[befovy](https://github.com/befovy)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
