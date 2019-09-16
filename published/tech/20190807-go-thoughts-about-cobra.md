首发于：https://studygolang.com/articles/23438

# Go：关于 Cobra 的想法

!["Golang之旅"插图，来自 Go Gopher 的 Renee French 创作](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-thoughts-about-cobra/A%20Journey%20With%20Go.png)

Cobra 是 Golang 生态系统中最着名的项目之一。它简单，高效，并得到 Go 社区的大力支持。让我们来深入探索一下。

## 设计

Cobra 中的 `Command` 是一个具有名称，使用描述和运行逻辑函数的结构体：

```go
cmd := &cobra.Command{
    Run:   runGreet,
    Use:   `greet`,
    Short: "Greet",
    Long:  "This command will print Hello World",
}
```

设计非常类似于原生的 go 标准库命令，如 `go env`，`go fmt`等

比如，`go fmt` 命令结构：

```go
var CmdFmt = &base.Command{
    Run:       runFmt,
    UsageLine: "go fmt [-n] [-x] [packages]",
    Short:     "gofmt (reformat) package sources",
    Long: `
Fmt runs the command 'gofmt -l -w' on the packages named
by the import paths. It prints the names of the files that are modified.
For more about gofmt, see 'go doc cmd/gofmt'.
For more about specifying packages, see 'go help packages'.
The -n flag prints commands that would be executed.
The -x flag prints commands as they are executed.
To run gofmt with specific options, run gofmt itself.
See also: go fix, go vet.
    `,
}
```

如果你熟悉了 Cobra，很容易理解内部命令是如何工作的，反之亦然。我们可能会想，当 Go 已经定义了命令接口后，为什么还要要使用外部库？

Go 标准库定义的接口：

```go
type Command struct {
    // Run 运行命令.
    // 参数在命令之后
    Run func(cmd *Command, args []string)

    // UsageLine 是一行描述信息.
    // 其中第一个词语应是命令.
    UsageLine string

    // Short 是 'go help' 输出的简单描述.
    Short string

    // Long 是 'go help <this-command>' 输出的详细描述.
    Long string

    // Flag 一组特定于此命令的标志码.
    Flag flag.FlagSet

    // CustomFlags 表示了命令将执行自定义的标志解析
    CustomFlags bool

    // Commands 列举可用的命令和 help 主题.
    // 顺序和 'go help' 输出一致.
    // 注意：通常最好避免使用子命令.
    Commands []*Command
}

```

此接口是仅适用于标准库的内部包的一部分。在 2014 年 6 月的 Go 1.4 版本中，Russ Cox 提出了 [限制使用内部包和命令的建议](https://docs.google.com/document/d/1e8kOo3r51b2BWtTs_1uADIA5djfXhPT36s6eHVRIvaU/edit)。基于此内部包构建命令会带来错误：

```go
package main

import (
    "cmd/go/internal/base"
)

func main() {
    cmd := &base.Command{
        Run: func(cmd *base.Command, args []string) {
            println(`Hello`)
        },
        Short: `Hello`,
    }
    cmd.Run(cmd, []string{})
}
```

```sh
main.go:4:2: use of internal package cmd/go/internal/base not allowed
```

然而，正如 Cobra 创建者 [Steve Francia](https://www.linkedin.com/in/stevefrancia/) 所解释的那样：这个内部界面设计 [催生了了 Cobra](https://blog.gopheracademy.com/advent-2014/introducing-cobra/)（Steve Franci 在 Google 工作并曾直接参与了 Go 项目。）。

该项目也建立在来自同一作者的 [pflag 项目](https://github.com/spf13/pflag) 之上，提供符合 POSIX 标准。因此，程序包支持短标记和长标记，如`-e`替代`--example` ,或者多个选项，如`-abc` 和`-a`，`-b` 和`-c` 都是是有效选项。这旨在改进 Go 库中的 `flag` 包，该库仅支持标志`-xxx`。

## 特性

Cobra 有一些值得了解的简便方法：

* Cobra 提供了两种方法来运行我们的逻辑： `Run func(cmd *Command, args []string)` 和 `RunE func(cmd *Command, args []string) error` ，后者可以返回一个错误，我们将能够从 `Execute()` 方法的返回中捕获。

* `Command` 结构 提供了一个 `Aliases（别名）` 属性，允许我们将命令迁移到一个新名称，而不需要在`alias`属性中通过映射旧名称来破坏现有的行为。这种兼容性策略甚至可以通过使用 `Deprecated` 属性来增强，该属性允许您将一个命令标记为`Deprecated(即将弃用，不推荐使用)`，并在删除它之前提供一个简短的说明。

* 由于每个命令都可以嵌入其他命令，因此 Cobra 本身支持嵌套命令，并允许我们像下边这样编写：

```sh
go run main.go foo bar
```

在这里， `foo` 是命令，`bar` 是嵌套命令：

```go
package main

import (
    "github.com/spf13/cobra"
)

func main() {
    cmd := newCommand() // 构建一般命令
    cmd.AddCommand(newNestedCommand()) // 加入嵌套命令

    rootCmd := &cobra.Command{}
    rootCmd.AddCommand(cmd)

    if err := rootCmd.Execute(); err != nil {
        println(err.Error())
    }
}

func newCommand() *cobra.Command {
    cmd := &cobra.Command{
        Run:  func (cmd *cobra.Command, args []string) {
            println(`Foo`)
        },
        Use:   `foo`,
        Short: "Command foo",
        Long:  "This is a command",
    }

    return cmd
}

func newNestedCommand() *cobra.Command {
    cmd := &cobra.Command{
        Run:  func (cmd *cobra.Command, args []string) {
            println(`Bar`)
        },
        Use:   `bar`,
        Short: "Command bar",
        Long:  "This is a nested command",
    }

    return cmd
}
```

可以使用嵌套命令是 [决定构建 Cobra](https://spf13.com/post/announcing-cobra/)的主要动机之一
<!-- The nested commands were one of the main motivations when [deciding to build Cobra](https://spf13.com/post/announcing-cobra/). -->

## 轻量

这个库的代码主要包含一个文件，而且很好理解，它不会影响你程序的性能。接下来，我们做一个压力测试（benchmark）：

```go
package main

import (
    "github.com/spf13/cobra"
    "math/rand"
    "os"
    "strconv"
    "testing"
)

func BenchmarkCmd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        root := &cobra.Command{
            Run: func(cmd *cobra.Command, args []string) {
                println(`main`)
            },
            Use:   `test`,
            Short: "test",
        }

        max := 100
        for c := 0; c < max; c++ {
            cmd := &cobra.Command{
                Run: func(cmd *cobra.Command, args []string) {
                    _ = c
                },
                Use:   `test-`+strconv.Itoa(c),
                Short: `test `+strconv.Itoa(c),
            }
            root.AddCommand(cmd)
        }

        r := rand.Intn(max)
        os.Args = []string{"go", "test-"+strconv.Itoa(r)}
        _ = root.Execute()
    }
}
```

Cobra 运行 50 条命令只有 49.0μs 负载：

```sh
name   time/op
Cmd-8  49.0µs ± 1%

name   alloc/op
Cmd-8  78.3kB ± 0%

name   allocs/op
Cmd-8  646 ± 0%
```

由于 Cobra 被设计运行在 CLI 模式下, 性能并不重要, 但是可以看出这个库有多么轻量.

## 可替代性

即使 Cobra 倾向于成为 Go 社区的标准包 - 浏览[最近使用 Cobra 的项目](https://github.com/spf13/cobra)证实了这一点 - 了解 Go 生态系统中有关 CLI 接口的内容总是好的。

让我们回顾两个可以替代 Cobra 的项目：

* [cli](https://github.com/urfave/cli)，一个用于构建命令行应用程序的包。这个包和 Cobra 一样流行，与嵌套命令，bash 补全，hook（钩子），alias(别名)等非常相似。但是，与 Cobra 不同，这个包使用 Go 库中的原生`flag`包。

urfave/cli 例子：

```go
package main

import (
    "log"
    "os"
    "github.com/urfave/cli"
)

func main() {
    app := cli.NewApp()
    app.Commands = []cli.Command{
        {
            Action:  func(c *cli.Context) error {
                println("Hello world")
                return nil
            },
            Name:   `greet`,
            Usage:  "This command will print Hello World",
        },
    }

    err := app.Run(os.Args)
    if err != nil {
        log.Fatal(err)
    }
}
```

* [subcommands](https://github.com/google/subcommands)：虽然托管在 Google 的 Github 帐户中，但该项目并非官方 Google 产品。该库也很简单

google/subcommands 示例

```go
package main

import (
    "context"
    "flag"
    "github.com/google/subcommands"
)

type GreetCommand struct {}
func (g *GreetCommand) Name() string     { return "greet" }
func (g *GreetCommand) Synopsis() string { return "Greet the world." }
func (g *GreetCommand) Usage() string { return `Print Hello World.` }
func (g *GreetCommand) SetFlags(*flag.FlagSet) {}
func (p *GreetCommand) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
    println(`Hello World`)

    return subcommands.ExitSuccess
}

func main() {
    subcommands.Register(&GreetCommand{}, "foo")

    flag.Parse()
    subcommands.Execute(context.Background())
}
```

如我们之前看到的 Cobra 或 CLI，该库基于一个接口而不是一个结构体，因此并使代码稍显冗长。

---

via: <https://medium.com/a-journey-with-go/go-thoughts-about-cobra-f4e8c5f18091>

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[TomatoAres](https://github.com/TomatoAres)
校对：[DingdingZhou](https://github.com/DingdingZhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
