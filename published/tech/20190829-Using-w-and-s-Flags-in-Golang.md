首发于：https://studygolang.com/articles/25914

# 在 Golang 中使用 -w 和 -s 标志

今天的博客文章来自 Valery，这是 [Spiral Scout](https://spiralscout.com/) 的一名高级软件工程师，专门从事 Golang（Go）开发。

作为在 Golang 以及许多其他编程语言方面具有专业知识的软件开发机构，我们知道对于我们的工程师和质量保证专家而言，能够与社区分享他们的知识和经验非常重要。 感谢 Valery 这篇出色的文章和有用的 Golang 测试技巧！

当我在 GitHub 上查找一些良好的工程实践以备应用时，我注意到许多开发人员编译他们的 Go 程序时经常出现的问题，他们中许多人都使用链接器标记（linker flags）来减小输出文件大小，尤其是同时使用 `-w` 和 `-s` 标记所带来的叠加效果。

在软件测试中，标记也被称为参数。当从命令行运行程序时，它们用于标识特定的状态或条件。 标记可以打开或关闭，并且在整个软件开发过程中大量语言和框架都采用这种方式。

本文致力于说明在 Go 中实现 `-w` 和 `-s` 标志的效果，并提供可以更有效地使用它们的方法。

## `-w` 和 `-s` 标志如何与 DWARF 和 ELF 配合使用

在讨论何时以及如何使用 `-w` 和 `-s` 标志之前，先简要介绍一下我的测试环境。 我使用的硬件/软件组合包括：

* A Dell XPS 9570 laptop
* Manjaro Linux OS
* Testing branch

`-w` 和 `-s` 标志通常用在 App 链接阶段和 Go 编译阶段 `-ldflags` 指令结合使用 （参见 [Go 命令文档](https://golang.org/src/cmd/go/alldocs.go) ）。有关标志的更多信息，请参见：https://golang.org/cmd/link/。

在我们仔细查看 `-w` 标志并拆解二进制代码以检查 DWARF 符号表是否消失之前，我建议先明确 DWARF 符号表的定义。

DWARF 是一种可以包含在二进制文件中的调试数据格式。 根据维基百科 [DWARF 条目](https://en.wikipedia.org/wiki/DWARF)，此格式是与称为 ELF（可执行和可链接格式）的标准通用文件格式一起开发的。 [这篇文章](https://eli.thegreenplace.net/2011/02/07/how-debuggers-work-part-3-debugging-information/) 很好地解释了调试器如何与 DWARF 表配合工作。

Golang 的创建者们在 [Go DWARF 源代码](https://golang.org/src/cmd/link/internal/ld/dwarf.go) 中分享了更多信息，包括有关如何形成此表并将其嵌入以 Go 编写的二进制文件的详细信息。

我将通过示例代码介绍以下一些要点。

首先，我们要使用以下步骤读取 DWARF 信息：

1. 编译 Go 程序（开始我们仅使用 Go build 命令）

```shell
go build -o simple_build cmd/main.go
```

2. 读取符号表。使用 `readelf -Ws` 可以方便的实现符号表读取。但是你也可以使用其他更熟悉的工具读取文件头（比如 `objdump -h`）。

3. 请注意生成的程序的头部内容。

![](https://cdn.jsdelivr.net/gh/studygolang/gctt-images2@master/using-w-and-s-flags-in-golang/section-headers-no-flags.png)

我们可以看到这个二进制文件中包含了用于调试的数据（从第 24 节到第 32 节），并且还有一个符号表和字符串表。（如下所述）

4. 使用如下命令进行读表

```shell
> objdump - dwarf=info main
```

输出看起来有些长，所以我用下面的命令把输出保存到文件中。

```shell
> objdump - dwarf=info main &> main.txt
```

你可以在下面看到输出的一部分：

![](https://cdn.jsdelivr.net/gh/studygolang/gctt-images2@master/using-w-and-s-flags-in-golang/objdump.png)

为了根据地址查找必要的函数，我们需要知道 PC （程序计数器）。你可以在 EIP 寄存器中找到 PC 值，它由 DW_AT_low_pc 和 DW_AT_high_pc 表示。举个例子，对于 `main.main` 函数（`main` 是 Go 运行时的函数）使用 `low_pc` ，并尝试使用 `objdump -d` 在二进制文件的位置 `0x44f930` 找到它。

![](https://cdn.jsdelivr.net/gh/studygolang/gctt-images2@master/using-w-and-s-flags-in-golang/objdump-pc.png)

很棒。现在我们使用 `-w` 标志编译程序并且和不使用标志编译出来的程序进行比较。

5. 运行下面的命令

```shell
go build -ldflags=”-w” -o build_with_w cmd/main.go
```

然后看看生成文件的头部发生了什么变化：
![](https://cdn.jsdelivr.net/gh/studygolang/gctt-images2@master/using-w-and-s-flags-in-golang/section-headers-flags-w.png)

正如我们看到的，`.zdebug` 部分完全消失了。通过对顶部低位（Off 列）地址相减，我们可以精确计算二进制文件减小了多少。当你把这个差值从 Bytes 转换到 KB 时，可以对实际情况有更直观的体会。

在这个案例中，二进制文件的总大小大约 25MB，这意味着我们节省了大约 3.7KB。这让我好奇如果我尝试使用 [Delve Go 调试器工具](https://github.com/go-delve/delve) 运行 dvl 时会发生什么？

6. 运行

```shell
dlv — listen=:43671 — headless=true — api-version=2 — accept-multiclient exec ./build_with_w
```

... 返回了你所期待的结果：

```shell
API server listening at: [::]:43671
could not launch process: could not open debug info
```

好了，现在关于 DWARF 表和 `-w` 标志的作用变得更加清楚了。

让我们继续看看 `-s` 标志。根据文档，`-s` 标志不仅删除了调试信息，同时还删除了指定的符号表。不过，它与 `-s` 标志有何不同呢？

首先，快速了解一下 — 符号表包含了局部变量、全局变量和函数名等的信息。在上图中，这些信息在第 26 和第 27 节（.symtab 和 .strtab）给出。更多关于符号表的详细信息，可以在这里找到：
[http://refspecs.linuxbase.org/elf/gabi4+/ch4.symtab.html](http://refspecs.linuxbase.org/elf/gabi4+/ch4.symtab.html) 和 [http://refspecs.linuxbase.org/elf/gabi4+/ch4.strtab.html](http://refspecs.linuxbase.org/elf/gabi4+/ch4.strtab.html).

这次我们试试用 `-s` 标志编译一个二进制文件：

![](https://cdn.jsdelivr.net/gh/studygolang/gctt-images2@master/using-w-and-s-flags-in-golang/section-headers-flags-s.png)

如同我们期待的一样，关于 DWARF 的信息以及符号表和字符串表（一种发布标志）的内容都消失不见了。

## 这意味着什么?

如果只想删除调试信息，只使用 `-w` 标志是最合适的。如果要另外删除符号和字符串表以减小二进制文件的大小，请使用 `-s` 标志。

下面是在 Golang 中使用这些 flag 的的**反面教材**，不建议大家这样使用。
尽管两个标志似乎比一个标志好，但是对于 `-w` 和 `-s` 标志，情况却并非如此：
![](https://cdn.jsdelivr.net/gh/studygolang/gctt-images2@master/using-w-and-s-flags-in-golang/github-search.png)

---

via: https://blog.spiralscout.com/using-w-and-s-flags-in-golang-97ae59b50e26

作者：[John Griffin](https://blog.spiralscout.com/@johnwgriffin)
译者：[befovy](https://github.com/befovy)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
