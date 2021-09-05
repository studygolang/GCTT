首发于：https://studygolang.com/articles/35210

# Go Module 教程第 3 部分：最小版本选择

前两个教程：

- [Go Module 教程第 1 部分：为什么和做什么](https://studygolang.com/articles/24580)
- [Go Module 教程第 2 部分：项目、依赖和 gopls](https://studygolang.com/articles/35202)

> 注意，该教程基于 Go1.13。最新版本可能会有所不同。

## 引言

每个依赖管理解决方案都必须解决选择依赖版本的问题。目前存在的许多版本选择算法都试图识别任何依赖关系的“最新最大”版本。如果你相信语义版本控制将被正确应用，社会契约将得到尊重，那么这是有意义的。在这些情况下，依赖项的“最新最大”版本应该是最稳定和安全的版本，并且应该与早期版本具有向后兼容性。至少在相同的主版本依赖关系树中应该如此。

Go 决定采取一种不同的方法，Russ Cox 花了大量的时间和精力[写作](https://research.swtch.com/vgo)和[讨论](https://www.youtube.com/watch?v=F8nrpe0XWRg) Go 团队的版本选择方法，这种方法被称为最小版本选择或 MVS。实质上，Go 团队相信 MVS 为 Go 程序提供了持久的、可重复的长期构建的最佳机会。我建议读一读[这篇文章](https://research.swtch.com/vgo-principles)，理解为什么 Go 团队相信这一点。

在本文中，我将尽力解释 MVS 语义，并展示一个实际的 Go 示例和 MVS 算法。

## MVS 语义

命名 Go 的选择算法“最小版本选择”有点用词不当，但是一旦你了解了它的工作原理，你就会发现它的名字非常接近。正如我之前所说的，许多选择算法选择依赖项的“最新最大”版本。我喜欢把 MVS 看作是一种选择“最新的非最大”版本的算法。并不是 MVS 不能选择“最新的最大”，只是如果项目中的任何依赖项都不需要“最新的最大”，那么就不需要该版本。

为了更好地理解这一点，让我们创建这样一种情况: 几个模块(A、B 和 C)依赖于同一个模块(D) ，但每个模块需要不同的版本。

![图1](https://www.ardanlabs.com/images/goinggo/111_figure1.png)

图 1 显示了模块 A、B 和 C 各自独立地需要模块 D，并且每个模块都需要不同版本的模块。

如果我启动一个需要模块 A 的项目，那么为了构建代码，我还需要模块 D。模块 D 有很多版本可供选择。例如，想象模块 D 是 sirupsen 中的 logrus 模块。我可以要求 Go 为我提供一个所有已经被标记为模块 D 的版本的列表。

**清单 1**

```bash
$ go list -m -versions github.com/sirupsen/logrus

github.com/sirupsen/logrus v0.1.0 v0.1.1 v0.2.0
v0.3.0 v0.4.0 v0.4.1 v0.5.0 v0.5.1 v0.6.0 v0.6.1
v0.6.2 v0.6.3 v0.6.4 v0.6.5 v0.6.6 v0.7.0 v0.7.1 
v0.7.2 v0.7.3 v0.8.0 v0.8.1 v0.8.2 v0.8.3 v0.8.4
v0.8.5 v0.8.6 v0.8.7 v0.9.0 v0.10.0 v0.11.0 v0.11.1
v0.11.2 v0.11.3 v0.11.4 v0.11.5 v1.0.0 v1.0.1 v1.0.3
v1.0.4 v1.0.5 v1.0.6 v1.1.0 v1.1.1 v1.2.0 v1.3.0
v1.4.0 v1.4.1 v1.4.2
```

清单 1 显示了模块 D 的所有版本，其中显示了“最新最大”的版本为 v1.4.2。

应该为项目选择哪个版本的模块 D？实际上有两种选择。第一个选择是选择“最新最大”的版本(在这一行的主要版本 1 版本中) ，它将是版本 1.4.2。第二个选择是选择模块 A 需要的版本，即 v1.0.6 版本。

像 dep 这样的依赖工具会选择 v1.4.2 版本，并在语义版本控制和社会契约得到尊重的前提下工作。然而，根据 Russ 在[该文](https://research.swtch.com/vgo-principles)中定义的原因，Go 将尊重模块 A 的要求，选择 v1.0.6 版本。Go 为项目中需要该模块的所有依赖项选择当前在所需版本集中的“最小”版本。换句话说，现在只有模块 A 需要模块 D，而模块 A 已经指定它需要版本 v1.0.6，因此这将作为模块 D 的版本。

如果我引入新的代码，要求项目导入模块 B，会怎么样？一旦模块 B 被导入到项目中，Go 将该项目的模块 D 的版本从 v1.0.6 升级到 v1.2.0。再次为项目中需要模块 D 的所有依赖项(模块 A 和模块 B)选择模块 D 的“最小”版本，该版本目前位于所需版本集(v1.0.6 和 v1.2.0 )中。

如果我再次引入需要项目导入模块 C 的新代码会怎么样？然后 Go 将从所需的版本集(v1.0.6、 v1.2.0、 v1.3.2)中选择最新版本(v1.3.2)。请注意，v1.3.2 版本仍然是“最小”版本，而不是模块 D (v1.4.2)的“最新最大”版本。

最后，如果我删除刚刚为模块 C 添加的代码会怎样？Go 将把该项目锁定到模块 D 的版本 v1.3.2 中，降级回到版本 v1.2.0 将是一个更大的改变，而且 Go 知道版本 v1.3.2 工作正常且稳定，因此版本 v1.3.2 仍然是该项目模块 D 的“最新非最大”或“最小”版本。另外，模块文件只维护一个快照，而不是日志。没有关于历史撤销或降级的信息。

这就是为什么我喜欢将 MVS 看作是一种选择模块的“最新非最大”版本的算法。希望您现在明白为什么 Russ 在命名算法时选择 “minimal”这个名称。

## 示例项目

有了这个基础，我将以上放在一个项目里，这样你就可以看到 Go 和 MVS 算法起的作用。在这个项目中，模块 D 将表示 logrus 模块，该项目直接依赖 [rethinkdb-go](https://github.com/rethinkdb/rethinkdb-go) (模块 A)和 [golib](https://github.com/Bhinneka/golib) (模块 B)模块。Rethinkdb-go 和 golib 模块直接依赖 logrus 模块，并且每个模块都需要不同的版本，而不是 logrus 的“最新最大”版本。

![图2](https://www.ardanlabs.com/images/goinggo/111_figure2.png)

图 2 显示了这三个模块之间的独立关系。首先，我将创建项目，初始化模块，然后启动 VSCode。

**清单 2**

```bash
$ cd $HOME
$ mkdir app
$ mkdir app/cmd
$ mkdir app/cmd/db
$ touch app/cmd/db/main.go
$ cd app
$ go mod init app
$ code .
```

清单 2 显示了要运行的所有命令。

![图3](https://www.ardanlabs.com/images/goinggo/111_figure3.png)

图 3 显示了项目结构和模块文件应该包含的内容。现在可以添加使用 rethinkdb-go 模块的代码了。

**清单 3**：<https://play.studygolang.com/p/bc5I0Afxhvc>

```go
package main

import (
	"context"
	"log"

	db "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

func main() {
	c, err := db.NewCluster([]db.Host{{Name: "localhost", Port: 3000}}, nil)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err = c.Query(context.Background(), db.Query{}); err != nil {
		log.Fatalln(err)
	}
}
```

清单 3 引入了 rethinkdb-go 模块的主版本 5。添加并保存这段代码之后，Go 查找、下载并提取模块，更新 go.mod 和 go.sum 文件。

**清单 4**

```bash
module app

go 1.13

require gopkg.in/rethinkdb/rethinkdb-go.v5 v5.0.1
```

清单 4 显示了 go.mod 文件，该文件要求 rethinkdb-go 模块作为一个直接依赖项，选择 v5.0.1版本，这是该模块的“最新最大”版本。

**清单 5**

```bash
...
github.com/sirupsen/logrus v1.0.6 h1:hcP1GmhGigz/O7h1WVUM5KklBp1JoNS9FggWKdj/j3s=
github.com/sirupsen/logrus v1.0.6/go.mod h1:pMByvHTf9Beacp5x1UXfOR9xyW/9antXMhjMPG0dEzc=
...
```

清单 5 显示了 go.sum 文件中的两行代码，它们是 logrus 模块的 v1.0.6 版本。此时，你可以看到 MVS 算法已经选择了 logrus 模块的“最小”版本，以满足 rethinkdb-go 模块所指定的需求。记住 logrus 模块的最新版本是 1.4.2。

注意：go.sum 文件应该被视为不透明的可靠性物件，不应该用它来理解您的依赖关系。我在上面所做的确定版本是错误的，不久我将向您展示确定你的项目使用什么版本的正确方法。

![图4](https://www.ardanlabs.com/images/goinggo/111_figure4.png)

图 4 显示了将使用哪个版本的 logrus 模块 Go 来构建项目中的代码。

接下来，我将添加引入 golib 模块依赖项的代码。

**清单 6**：<https://play.studygolang.com/p/h23opcp5qd0>

```go
package main

import (
	"context"
	"log"

	"github.com/Bhinneka/golib"
	db "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

func main() {
	c, err := db.NewCluster([]db.Host{{Name: "localhost", Port: 3000}}, nil)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err = c.Query(context.Background(), db.Query{}); err != nil {
		log.Fatalln(err)
	}
	
	golib.CreateDBConnection("")
}
```

清单 6 为程序添加了第 07 行和第 21 行。一旦 Go 查找、下载并提取 golib 模块，go.mod 文件中将显示以下更改。

**清单 7**

```bash
module app

go 1.13

require (
		github.com/Bhinneka/golib v0.0.0-20191209103129-1dc569916cba
    gopkg.in/rethinkdb/rethinkdb-go.v5 v5.0.1
)
```

清单 7 显示了 go.mod 文件已经被修改，以包含 golib 模块对该模块的“最新最大”版本的依赖关系，该模块没有语义版本标记。

**清单 8**

```bash
...
github.com/sirupsen/logrus v1.0.6 h1:hcP1GmhGigz/O7h1WVUM5KklBp1JoNS9FggWKdj/j3s=
github.com/sirupsen/logrus v1.0.6/go.mod h1:pMByvHTf9Beacp5x1UXfOR9xyW/9antXMhjMPG0dEzc=
github.com/sirupsen/logrus v1.2.0 h1:juTguoYk5qI21pwyTXY3B3Y5cOTH3ZUyZCg1v/mihuo=
github.com/sirupsen/logrus v1.2.0/go.mod h1:LxeOpSwHxABJmUn/MG1IvRgCAasNZTLOkJPxbbu5VWo=
...
```

清单 8 显示了 go.sum 文件中的四行代码，它们现在包含 logrus 模块的 v1.0.6 和 v1.2.0 版本。看到 go.sum 文件中列出的两个版本，就会产生两个疑问：

1. 为什么两个版本都在 `go.sum` 文件中？
2. 当 Go 执行构建时，会使用哪个版本？

两个版本都在 go.sum 文件中列出的原因，Go 团队的 Bryan Mills 回答得更好。

”go.sum 文件仍然包含旧版本(1.0.6) ，因为它的传递需求可能会影响其他模块的选定版本。我们实际上只需要 go.mod 文件的校验和，因为它声明了那些可传递的需求，但我们最终还是保留了源代码的校验和，因为 go mod tidy 并不像它应该的那样精确。” <https://github.com/golang/go/issues/33008

这仍然留下了在构建项目时将使用哪个版本的 logrus 模块的问题。要正确识别将要使用的模块及其版本，不要查看 go.sum 文件，而是使用 go list 命令。

**清单 9**

```bash
$ go list -m all | grep logrus

github.com/sirupsen/logrus v1.2.0
```

清单 9 显示了在构建项目时将使用 logrus 模块的 v1.2.0 版本。M 标志将 go list 指向列表模块而不是包。

查看模块图可以更深入地了解项目对 logrus 模块的需求。

**清单 10**

```bash
$ go mod graph | grep logrus

github.com/sirupsen/logrus@v1.2.0 github.com/pmezard/go-difflib@v1.0.0
github.com/sirupsen/logrus@v1.2.0 github.com/stretchr/objx@v0.1.1
github.com/sirupsen/logrus@v1.2.0 github.com/stretchr/testify@v1.2.2
github.com/sirupsen/logrus@v1.2.0 golang.org/x/crypto@v0.0.0-20180904163835-0709b304e793
github.com/sirupsen/logrus@v1.2.0 golang.org/x/sys@v0.0.0-20180905080454-ebe1bf3edb33
gopkg.in/rethinkdb/rethinkdb-go.v5@v5.0.1 github.com/sirupsen/logrus@v1.0.6
github.com/sirupsen/logrus@v1.2.0 github.com/konsorten/go-windows-terminal-sequences@v1.0.1
github.com/sirupsen/logrus@v1.2.0 github.com/davecgh/go-spew@v1.1.1
github.com/Bhinneka/golib@v0.0.0-20191209103129-1dc569916cba github.com/sirupsen/logrus@v1.2.0
github.com/prometheus/common@v0.2.0 github.com/sirupsen/logrus@v1.2.0
```

清单 10 显示了 logrus 模块在项目中的关系。我将提取直接显示 logrus 上的依赖需求的代码行。

**清单 11**

```bash
gopkg.in/rethinkdb/rethinkdb-go.v5@v5.0.1 github.com/sirupsen/logrus@v1.0.6
github.com/Bhinneka/golib@v0.0.0-20191209103129-1dc569916cba github.com/sirupsen/logrus@v1.2.0
github.com/prometheus/common@v0.2.0 github.com/sirupsen/logrus@v1.2.0
```

在清单 11 中，这些代码行显示三个模块(rethinkdb-go、 golib、 common)都需要 logrus 模块。多亏了 go list 命令，我知道所需的最小版本是 v1.2.0 版本。

![图5](https://www.ardanlabs.com/images/goinggo/111_figure5.png)

图 5 显示了 logrus 模块 Go 现在将使用哪个版本来构建项目中所有需要 logrus 模块的依赖项的代码。

## Go Mod Tidy

在你将代码 commit/push 到 repo 之前，运行 go mod tidy 以确保你的模块文件是当前的和准确的。你在本地构建、运行或测试的代码将影响 Go 在任何时候决定更新模块文件中的内容。运行 go mod tidy 将保证项目有一个准确和完整的快照，什么是需要的，这将有助于你的团队和你的 CI/CD 环境的其他人。

**清单 12**

```bash
$ go mod tidy

go: finding github.com/Bhinneka/golib latest
go: finding github.com/bitly/go-hostpool latest
go: finding github.com/bmizerany/assert latest
```

清单 12 显示了运行 go mod tidy 时的输出。你可以在输出中看到两个新的依赖项。这将更改模块文件。

**清单 13**

```bash
module app

go 1.13

require (
    github.com/Bhinneka/golib v0.0.0-20191209103129-1dc569916cba
    github.com/bitly/go-hostpool v0.0.0-20171023180738-a3a6125de932 // indirect
    github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
    gopkg.in/rethinkdb/rethinkdb-go.v5 v5.0.1
)
```

清单 13 显示，go-hostpool 和 assert 模块被列为构建项目所需的间接模块。这里列出它们是因为这些项目目前不符合模块标准。换句话说，对于这些项目的任何标记版本或 master 中的“最新最大”版本，在 repo 中都不存在 go.mod 文件。

Why were these modules included after running `go mod tidy`? I can use the `go mod why` command to find out.

为什么这些模块包括后，运行去模组整理？我可以使用 go mod why 命令来找出答案。

**清单 14**

```bash
$ go mod why github.com/hailocab/go-hostpool

# github.com/hailocab/go-hostpool
app/cmd/db
gopkg.in/rethinkdb/rethinkdb-go.v5
github.com/hailocab/go-hostpool

------------------------------------------------

$ go mod why github.com/bmizerany/assert

# github.com/bmizerany/assert
app/cmd/db
gopkg.in/rethinkdb/rethinkdb-go.v5
github.com/hailocab/go-hostpool
github.com/hailocab/go-hostpool.test
github.com/bmizerany/assert
```

清单 14 显示了为什么项目间接需要这些模块。Rethinkdb-go 模块需要 go-hostpool 模块，go-hostpool 模块需要 assert 模块。

## 升级依赖

该项目有三个依赖项，每个依赖项都需要 logrus 模块，其中 logrus 模块的 v1.2.0 版本当前被选中。在项目生命周期的某个阶段，升级直接和间接的依赖关系将变得非常重要，以确保项目所需的代码是最新的，并且能够利用新特性、 bug 修复和安全补丁。为了应用升级，Go 提供了 go get 命令。

在你运行 go get 升级项目的依赖项之前，有几个选项需要考虑。

使用 MVS 升级只需要直接和间接依赖项。

我建议你从这种类型的升级开始，直到你对项目和模块有了更多的了解。这是 go get 最保守的形式。

**清单 15**

```bash
$ go get -t -d -v ./...
```

清单 15 展示了如何使用 MVS 算法执行只关注所需依赖项的升级。下面是 flag 的定义。

- `-t flag`：考虑构建测试所需的模块
- `-d flag`：下载每个模块的源代码，但不构建或安装它们
- `-v flag`：提供详细输出
- `./...` ：在整个源代码树上执行这些操作，并且只更新所需的依赖项

对当前项目运行此命令将不会导致任何更改，因为项目已经具有构建和测试项目所需的最小版本的最新版本。那是因为我刚刚运行了 go mod tidy，这个项目是新的。

**使用 Latest Greatest 升级所有直接和间接依赖项**

这种升级将使整个项目的依赖性从“最小”提高到“最大”。所需要做的就是将 -u 标志添加到命令行中。

**清单 16**

```bash
$ go get -u -t -d -v ./...

go: finding golang.org/x/net latest
go: finding golang.org/x/sys latest
go: finding github.com/hailocab/go-hostpool latest
go: finding golang.org/x/crypto latest
go: finding github.com/google/jsonapi latest
go: finding gopkg.in/bsm/ratelimit.v1 latest
go: finding github.com/Bhinneka/golib latest
```

清单 16 显示了使用 -u 标志运行 go get 命令的输出。这个输出没有告知真实的情况。如果我向 go list 命令询问哪个版本的 logrus 模块现在正被用于构建项目，会发生什么情况？

**清单 17**

```bash
$ go list -m all | grep logrus

github.com/sirupsen/logrus v1.4.2
```

清单 17 显示了如何选择 logrus 的“最新最大”版本。为了将这个选项保存下来，对 go.mod 文件进行了更改。

```bash
module app

go 1.13

require (
    github.com/Bhinneka/golib v0.0.0-20191209103129-1dc569916cba
    github.com/bitly/go-hostpool v0.0.0-20171023180738-a3a6125de932 // indirect
    github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
    github.com/cenkalti/backoff v2.2.1+incompatible // indirect
    github.com/golang/protobuf v1.3.2 // indirect
    github.com/jinzhu/gorm v1.9.11 // indirect
    github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
    github.com/sirupsen/logrus v1.4.2 // indirect
    golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413 // indirect
    golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553 // indirect
    golang.org/x/sys v0.0.0-20191210023423-ac6580df4449 // indirect
    gopkg.in/rethinkdb/rethinkdb-go.v5 v5.0.1
)
```

清单 18 在第 13 行显示，v1.4.2 版本现在是项目中 logrus 模块的选定版本。模块文件中的这一行是 Go 在构建项目时所遵循的。即使删除了改变 logrus 模块依赖性的代码，v1.4.2 版本对于这个项目来说仍然是一成不变的。请记住，降级将是一个比升级到 v. 1.4.2 版本更大的变化。

在 go.sum 文件中可以看到哪些更改？

**清单 19**

```bash
github.com/sirupsen/logrus v1.0.6/go.mod h1:pMByvHTf9Beacp5x1UXfOR9xyW/9antXMhjMPG0dEzc=
github.com/sirupsen/logrus v1.2.0 h1:juTguoYk5qI21pwyTXY3B3Y5cOTH3ZUyZCg1v/mihuo=
github.com/sirupsen/logrus v1.2.0/go.mod h1:LxeOpSwHxABJmUn/MG1IvRgCAasNZTLOkJPxbbu5VWo=
github.com/sirupsen/logrus v1.4.2 h1:SPIRibHv4MatM3XXNO2BJeFLZwZ2LvZgfQ5+UNI2im4=
github.com/sirupsen/logrus v1.4.2/go.mod h1:tLMulIdttU9McNUspp0xgXVQah82FyeX6MwdIuYE2rE=
```

清单 19 显示了 logrus 的所有三个版本现在是如何在 go.sum 文件中表示的。正如上面 Bryan 所解释的，这是因为传递性需求可能会影响其他模块的选择版本。

![图6](https://www.ardanlabs.com/images/goinggo/111_figure6.png)

图 6 显示了 Go 现在将使用 logrus 模块哪个版本来构建项目中所有需要 logrus 模块的依赖项的代码。

**使用 Latest Greatest 升级所有直接和间接依赖项**

你可以用 ./...  所有升级选项，包括所有直接和间接依赖项，包括那些不需要构建项目的依赖项。

**清单 20**

```bash
$ go get -u -t -d -v all

go: downloading github.com/mattn/go-sqlite3 v1.11.0
go: extracting github.com/mattn/go-sqlite3 v1.11.0
go: finding github.com/bitly/go-hostpool latest
go: finding github.com/denisenkom/go-mssqldb latest
go: finding github.com/hailocab/go-hostpool latest
go: finding gopkg.in/bsm/ratelimit.v1 latest
go: finding github.com/google/jsonapi latest
go: finding golang.org/x/net latest
go: finding github.com/Bhinneka/golib latest
go: finding golang.org/x/crypto latest
go: finding gopkg.in/tomb.v1 latest
go: finding github.com/bmizerany/assert latest
go: finding github.com/erikstmartin/go-testdb latest
go: finding gopkg.in/check.v1 latest
go: finding golang.org/x/sys latest
go: finding github.com/golang-sql/civil latest
```

清单 20 显示了现在为项目找到、下载和提取了多少依赖项。

**清单 21**

```bash
Added to Module File
   cloud.google.com/go v0.49.0 // indirect
   github.com/denisenkom/go-mssqldb v0.0.0-20191128021309-1d7a30a10f73 // indirect
   github.com/google/go-cmp v0.3.1 // indirect
   github.com/jinzhu/now v1.1.1 // indirect
   github.com/lib/pq v1.2.0 // indirect
   github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
   github.com/onsi/ginkgo v1.10.3 // indirect
   github.com/onsi/gomega v1.7.1 // indirect
   github.com/stretchr/objx v0.2.0 // indirect
   google.golang.org/appengine v1.6.5 // indirect
   gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
   gopkg.in/yaml.v2 v2.2.7 // indirect

Removed from Module File
   github.com/golang/protobuf v1.3.2 // indirect
```

清单 21 显示了 go.mod 文件的更改，添加了更多的模块，删除了一个模块。

注意：如果你想 vendoring，go mod vendor 命令可以从 vendor 文件夹中取出测试文件。

作为一般准则，在使用 go get for your projects 升级依赖项时，不要使用 all 选项或 -u 标志。只使用你需要的模块，并使用 MVS 算法来选择这些模块及其版本。必要时手动覆盖特定的模块版本。可以通过手动编辑 go.mod 文件来完成手动重写，我将在以后的文章中向你展示。

## 重置依赖关系

如果在任何时候你对模块和被选中的版本感到不舒服，你总是可以通过删除模块文件和运行 go mod tidy 来重置选择。当项目还在开始阶段，而且情况不稳定时，这是一个更好的选择。一旦项目稳定并发布，我会犹豫是否重置依赖关系。正如我前面提到的，随着时间的推移，模块版本可能会被设置，并且您希望在长时间内使用持久的和可重复的构建。

**清单 22**

```bash
$ rm go.*
$ go mod init <module name>
$ go mod tidy
```

清单 22 显示了可以运行的命令，以允许 MVS 从头再次执行所有选择。在写这篇文章的整个过程中，我一直在这样做，以便重新设置项目并提供文章的列表。

## 总结

在这篇文章中，我解释了 MVS 语义，并展示了 Go 和 MVS 算法的实际应用示例。我还展示了一些 Go 命令，它们可以在你遇到困难或遇到未知问题时为你提供信息。在向项目添加越来越多的依赖项时，可能会遇到一些边缘情况。这是因为 Go 生态系统已经有 10 年的历史了，所有现有的项目都需要更多的时间才能达到模块兼容。

在以后的文章中，我将讨论在同一个项目中使用不同主要版本的依赖关系，以及如何手动检索和锁定依赖关系的特定版本。现在，我希望您能够更多地信任模块和 Go 工具，并且你能够更清楚地了解 MVS 如何随着时间的推移选择版本。


---

via: <https://www.ardanlabs.com/blog/2019/12/modules-03-minimal-version-selection.html>

作者：[William Kennedy](https://www.ardanlabs.com/)
译者：[polaris1119](https://github.com/polaris1119)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
