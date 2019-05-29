首发于：https://studygolang.com/articles/20748

# 调试使用 Go 1.12 发布的程序

## 介绍

Go 1.11 和 Go 1.12 允许开发人员调试他们部署到生产环境中经过优化的二进制文件，并且在这方面有了重大改进。

由于 Go 编译器越来越致力于生成更快的二进制文件，在文件的可调试性方面却产生了短板。在 Go 1.10 中，用户需要完全禁用优化，才能从 Delve 等交互式工具中获得良好的调试体验。但是用户不应该为了可调试性而放弃性能，尤其是对于生产环境的服务。如果在生产环境中出现问题，则需要在生产环境中对其进行调试，而不是部署未经优化的二进制文件。

对于 Go 1.11 和 1.12，我们专注于改进优化二进制文件的调试体验（Go 编译器的默认设置）。改进包括

- 特别是对于功能入口的参数，有更准确的值检查 ;
- 更精确地识别语句边界，以减少单步调试的跳跃，使断点更多地落在开发者期望的地方 ;
- 对于 Delve 调用 Go 函数（goroutines 和垃圾回收机制使得 Go 比 C 和 C++ 更复杂）的初步支持。

## 使用 Delve 调试优化代码

Delve 是 x86 架构下的 Go 调试器，支持 Linux 和 MacOS。 Delve 关注 Goroutines 和其他 Go 的功能，它提供了最好的 Go 调试体验。 Delve 也是 GoLand， VS Code 和 Vim 使用的调试引擎。

Delve 通常会通过 `-gcflags "all=-N -l"` 重建它正在调试的代码，这个参数会禁用内联和大多数优化。要使用 delve 调试优化代码，首先需要构建优化二进制文件，然后使用 `dlv exec your_program` 进行调试。对于发生崩溃的核心文件，则可以使用 `dlv core your_program your_core` 进行检查。使用 Go 1.12 和最新的 Delve 版本，即使在优化的二进制文件中，你也可以检查许多变量。

## 改进的值检查

通常，在调试 Go 1.10 生成的优化二进制文件时，变量值完全不可用。然而，从 Go 1.11 开始，即使在优化过的二进制文件中也可以检查变量，除非它们已经被完全优化。在 Go 1.11 中，编译器开始发送 DWARF 位置列表，因此调试器可以在进出寄存器时跟踪变量，并重构分散在不同寄存器和堆栈槽中的复杂对象。

## 改进的单步调试

下图是一个在 1.10 调试器中单步执行简单函数的示例，其中缺陷（跳过和重复的行）用红色箭头来突出显示。

![Debug in Go 1.10](https://raw.githubusercontent.com/studygolang/gctt-images/master/Debugging-what-you-deploy-in-Go-1.12/1.png)

这样的缺陷使得在单步执行程序并插入断点时很容易混淆当前的位置。

Go 1.11 和 1.12 记录语句边界信息，并通过优化和内联更好地跟踪源行号。因此，在 Go 1.12 中，逐步执行这段代码会在每一行停止，并按照你期望的顺序执行操作。

## 函数调用

Delve 中对于函数调用的支持仍在开发中，但简单的场景是可以胜任的。例如：

```bash
（dlv）call fib（6）
> main.main（）./ hello.go:15（PC: 0x49d648）
Values returned:
    ~r1：8
```

## 未来展望

Go 1.12 是为了优化二进制文件提供更好的调试体验而踏出的一步，我们计划进一步改进它。

可调试性和性能之间存在基本的权衡，因此我们把重点放在优先级最高的调试缺陷上，并努力收集自动化的指标以监控我们的进度并发现阻碍。

我们专注于为调试器生成有关变量位置的正确信息，因此可以被输出的变量都能被正确输出。我们还在考虑使更多的变量值可用，特别是在函数调用点等关键位置。尽管在许多情况下，改进这一问题会减慢程序执行速度。最后，我们正在努力改进单步调试：我们关注 panics 的单步调试顺序、循环的单步调试顺序，并且通常尽可能地遵循源码顺序。

## 关于 MacOS 支持的说明

Go 1.11 开始压缩调试信息以减少二进制文件大小。Delve 原生支持 MacOS 上的压缩调试信息，但 LLDB 和 GDB 都不支持。如果您使用的是 LLDB 或 GDB ，有两种解决方法：使用 `-ldflags=-compressdwarf=false` 或使用 [splitdwarf](https://godoc.org/golang.org/x/tools/cmd/splitdwarf)（`go get Golang.org/x/tools/cmd/splitdwarf`）来解压缩现有二进制文件中的调试信息。

作者：David Chase

---

via: https://blog.golang.org/debugging-what-you-deploy

作者：[David Chase](https://blog.golang.com/debugging-what-you-deploy)
译者：[RookieWilson](https://github.com/RookieWilson)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
