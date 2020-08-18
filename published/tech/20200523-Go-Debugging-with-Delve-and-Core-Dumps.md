首发于：https://studygolang.com/articles/30252

# Go：使用 Delve 和 Core Dump 调试代码

![由 Renee French 创作的原始 Go Gopher 为“ Go Go 之旅”创建的插图。](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200523-Go-Debugging-with-Delve-and-Core-Dumps/Illustration.png)
ℹ️ 这篇文章基于 Go Delve 1.4.1。

core dump 是一个包含着意外终止的程序其内存快照的文件。这个文件可以被用来事后调试（debugging）以了解为什么会发生崩溃，同时了解其中涉及到的变量。通过 `GOTRACEBACK`，Go 提供了一个环境变量用于控制程序崩溃时生成的输出信息。这个变量同样可以强制生成 core dump，从而使调试成为可能。

## GOTRACEBACK

`GOTRACEBACK` 控制着当程序崩溃时输入的详细程度。它可以使用不同的值：

- `none` 不显示任何 Goroutine 的堆栈信息。
- `single`，默认选项，显示当前 Goroutine 的堆栈信息。
- `all` 显示所有用户创建的 Goroutine 的堆栈信息。
- `system` 显示所有 Goroutine 的堆栈信息，即使是来自运行时的 goroutine。
- `crash` 与 `system` 类似，但是会生成 core dump。

最后的那个选项，给了我们在程序崩溃的情况下，调试我们程序的能力。如果没有足够的日志，或者崩溃无法复现时，这是一个好的选择。让我们以下面的程序为例：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20200523-Go-Debugging-with-Delve-and-Core-Dumps/example-program.png)

这个程序会很快崩溃：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20200523-Go-Debugging-with-Delve-and-Core-Dumps/crash.png)

从堆栈信息中我们无法获知那个值涉及了程序的崩溃。增加日志是一个解决办法，但我们无法一直知道要在哪里加日志。当问题无法复现的时候，编写测试用例以确保问题被修复是十分困难的。我们可以不断重复增加日志和运行程序的步骤，直到程序崩溃，再查看可能的原因后再运行。

设置 `GOTRACEBACK=crash` 后再次运行。输出信息更加详细，因为现在所有的 Goroutine 信息打印了出来，包括运行时的。无论如何，我们现在有了 core dump。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20200523-Go-Debugging-with-Delve-and-Core-Dumps/core-dump.png)

core dump 通过 `SIGABRT` 信号触发，该信号[生成 core dump 作为处置](http://man7.org/linux/man-pages/man7/signal.7.html)。

core dump 可以被诸如 [Go delve](https://github.com/go-delve/delve) 或者 [GBD](https://www.gnu.org/s/gdb/) 的调试信息分析。

## Delve

Delve 是用 Go 语言编写的 Go 程序调试器。它允许通过在用户代码和运行时代码的任意位置加断点来逐步调试，甚至通过 `dlv core` 命令来调试 core dump，这个命令以二进制和 core dump 为参数。

一旦命令运行，我们就可以开始与 core dump 进行交互：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20200523-Go-Debugging-with-Delve-and-Core-Dumps/interacting-with-the-core-dump.png)

`dlv` 命令 `bt` 打印堆栈信息并且显示程序生成的 panic 信息。之后，我们可以通过 `frame 9` 命令来访问 9 号帧：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20200523-Go-Debugging-with-Delve-and-Core-Dumps/frame9.png)

最终，用 `locals` 命令打印本地变量，来帮助了解哪个变量与崩溃有关：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20200523-Go-Debugging-with-Delve-and-Core-Dumps/value-was-involved-in-the-crash.png)

channel 是满的，并且生成的随机数是 203,300。而对于变量 `sum`，可以通过命令 `vars` 打印出它的内容，该命令用于打印包级别变量：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20200523-Go-Debugging-with-Delve-and-Core-Dumps/prints-the-package-variables.png)

*如果没有看到本地变量 `n` ，请确保使用编辑标志 `-N` 和 `-l` 来构建二进制程序，这些标志禁用编译器优化，而这些优化会是调试更加困难。完整的编译命令是：`go build -gcflags=all="-N -l"` 不要忘记运行 `ulimit -c unlimited`，选项 `-c` 定义了 core dump 的最大尺寸。*

---
via: https://medium.com/a-journey-with-go/go-debugging-with-delve-core-dumps-384145b2e8d9

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[dust347](https://github.com/dust347)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
