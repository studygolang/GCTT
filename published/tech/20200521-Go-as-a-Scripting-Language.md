首发于：https://studygolang.com/articles/34517

# 将 Go 作为脚本语言用

Go 作为一种可用于创建高性能网络和并发系统的编程语言，它的生态应用变得[越来越广泛](https://blog.golang.org/survey2019-results)，同时，这也激发了开发人员使用 Go 作为脚本语言的兴趣。虽然目前 Go 还未准备好作为脚本语言 “开箱即用” 的特性，用来替代 Python 和 Bash ，但是我们只需要一点点准备工作就可以达到想要的目标。

[正如来自 Codelang 的 Elton Minetto 所说的那样](https://dev.to/codenation/using-golang-as-a-scripting-language-jl2)，Go 作为一门脚本语言的同时，也具有相当大的吸引力，这不仅包括 Go 本身强大的功能和简洁的语法，还包括对 goroutines 的良好支持等。Google 公司的软件工程师 [Eyal Posener](https://posener.github.io/about/) 为 Go 用作脚本语言提供了[更多的理由](https://gist.github.com/posener/73ffd326d88483df6b1cb66e8ed1e0bd)，例如，丰富的标准库和语言的简洁性使得维护工作变得更加容易。与之相对的是，Go 的贡献者和前 Google 公司员工 David Crawshaw 则[强调了使用 Go 编写脚本任务的便利程度](https://news.ycombinator.com/item?id=15623106)，因为几乎所有的程序员都在花费大量的时间编写复杂的程序。

> 基本上，我一直在编写 Go 程序，偶尔会写写 Bash、perl 和 python 。有时候，这些编程语言会落入我的脑海。

对于日常任务和不太频繁的脚本编写任务，倘若能够使用相同的编程语言，它将会大大提高效率。Cloudflare 公司的工程师 Ignat Korchagin 指出，Go 是一种强类型语言，它能够[帮助 Go 脚本变得更加可靠，并且避免出现像拼写之类的小错误，从而不会出现发生在运行时的报错](https://blog.cloudflare.com/using-go-as-a-scripting-language-in-linux/)。

Codenation 使用 Go 编写的脚本文件，用来自动化地执行重复性的任务，这不仅是开发流程的一部分，还是其 CI/CD 管道中的任务。在 Codenation 内部，Go 脚本是通过 [`go run`](https://golang.org/cmd/go/#hdr-Compile_and_run_Go_program) 来执行的，`go run` 是 Go 构建工具链中的默认命令，能够一步一步地编译和运行 Go 程序。Posener 写道：“事实上，`go run` 并非作为解释器来使用。”

>[...] bash 和 python 都是解释型语言 —— 它们在读取脚本的时候，然后执行脚本文件。另一方面，当您键入 `go run` 时，Go 编译器就会编译程序，然后运行它们。Go 程序的编译时间很短，这使它看起来就像是解释型语言一般。

为了让 Go 编写的脚本在 shell 脚本程序中表现良好，Codenation 的工程师使用了许多有用的 Go 软件包：

- [github.com/fatih/color](https://github.com/fatih/color) 是用于输出对应编码颜色的包。
- [github.com/schollz/progressbar](https://github.com/schollz/progressbar) 是用于为执行时间过久的任务创建进度条的包。
- [github.com/jimlawless/whereami](https://github.com/jimlawless/whereami) 是用于捕获源代码的文件名、行号、函数等信息的包，这对于改正错误信息十分有用！
- [github.com/spf13/cobra](https://github.com/spf13/cobra) 是用于更轻松地创建带有输入选项和相关文档的复杂脚本的包。

 Crawshaw 写道：“ 对于 Codenation 来说，虽然使用命令行工具 `go run` 来运行 Go 程序非常有效，但是它并非最完美的解决方案。” 特别是，Go 缺乏对读取-求值-输出循环 (REPL) 的支持，并且无法轻松地将 `Shebang` ( 译者注：Unix 系统中，通常称 `#` 为 `sharp` 或 `she`；而称 `!` 为 `bang` ) 集成在一起，这使得脚本可以像二进制程序一样执行。此外，比起短小精悍的脚本文件，Go 错误处理更适合大型程序项目使用。由于这些原因，他开始研究 [Neugram](https://github.com/neugram/ng) ，该项目旨在创建一个 Go 克隆程序用来解决上述所有限制。不巧的是，Neugram 项目现在似乎已经被废弃，这可能是[由于 Go 语法的所有细节的复杂性](https://news.ycombinator.com/item?id=15623244)。

[Gomacro](https://github.com/cosmos72/gomacro) 项目使用了与 `Neugram` 类似的方法，它是一种 Go 的解释器，还支持类似 Lisp 的宏，既可以生成代码，又可以实现某种形式的[泛型](https://github.com/cosmos72/gomacro#generics)。

>`Gomacro` 几乎是一个完整的 Go 解释器，它使用纯 Go 语言实现。它提供了交互式的 `REPL` 模式和脚本模式，并且在运行时不需要 Go 构建工具链。(除了在在一种非常特殊的情况以外：在运行时导入第三方包。)

除了非常适合写脚本外，`Gomacro` 还旨在使得 Go 成为一种中间语言，表示要将它[解释为 Go 的标准详细规范](https://github.com/cosmos72/gomacro/blob/master/doc/code_generation.pdf)，还会[提供 Go 源代码的调试器](https://github.com/cosmos72/gomacro#debugger)。

尽管在使用 Go 编写脚本的情况下，`Gomacro` 为其提供了最大灵活性，但不幸的是，它不是标准的 Go 语言，这引起了另一种程度的担忧。[Posener 对使用标准的 Go 语言作为脚本语言的可能性进行了详细分析](https://gist.github.com/posener/73ffd326d88483df6b1cb66e8ed1e0bd)，包括针对丢失 `Shebang` 的解决方法。但是，这些解决方法都在某种程度上均有不足的体现。

> 似乎没有完美的解决方案，而且我也不明白为什么不能有一个完美的解决方案。看起来，运行 Go 脚本的方式最为简单，而最没有问题的方法就是使用 `go run` 命令。[...] 这就是我为什么认为在该语言领域上，仍需要做未完成的工作。同样的，我认为更改程序语言用来忽略 `Shebang` 不会有任何的危害。

但是，对于 Linux 系统，这里可能会有一个高级技巧，能够在具有完全 `Shebang` 支持的情况下，从命令行运行 Go 脚本。由 Korchagin 举例子并说明的这种方法，依赖于 `Shebang` 对 Linux 内核的支持，以及从 Linux 用户空间扩展受到支持二进制格式的可能性。长话短说，Korchagin 建议使用以下方式注册二进制：

```bash
$ Echo ':golang:E::go::/usr/local/bin/gorun:OC' | sudo tee /proc/sys/fs/binfmt_misc/register
:golang:E::go::/usr/local/bin/gorun:OC
```

这样就可以设置完全标准的 Go 语言的可执行位，例如：

```go
package main
import (
    "fmt"
    "os"
)
func main() {
    s := "world"

    if len(os.Args) > 1 {
        s = os.Args[1]
    }

    fmt.Printf("Hello, %v!", s)
    fmt.Println("")

    if s == "fail" {
        os.Exit(30)
    }
}
```

然后执行：

```bash
$ chmod u+x helloscript.go
$ ./helloscript.go
Hello, world!
$ ./helloscript.go gopher
Hello, gopher!
$ ./helloscript.go fail
Hello, fail!
$ Echo $?
30
```

尽管这种方法无法提供对 `REPL` 的支持，但是 `Shebang` 可能足以满足典型的用例。

---
via: https://www.infoq.com/news/2020/04/go-scripting-language/

作者：[Sergio De Simone](https://www.infoq.com/profile/Sergio-De-Simone/)
译者：[sunlingbot](https://github.com/sunlingbot)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
