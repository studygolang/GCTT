# 编写友好的命令行应用程序

我来给你讲一个故事...

1986年 ， [Knuth](https://en.wikipedia.org/wiki/Donald_Knuth) 编写了一个程序来演示[文学式编程](https://en.wikipedia.org/wiki/Literate_programming) 。

这段程序目的是读取一个文本文件，找到 n 个最常使用的单词，然后有序输出这些单词以及它们的频率。 Knuth 写了一个完美的 10 页程序。

Doug Mcllory 看到这里然后写了 `tr -cs A-Za-z '\n' | tr A-Z a-z | sort | uniq -c | sort -rn | sed ${1}q` 。

现在是2019年了，为什么我还要给你们讲一个发生在33年前（可能比一些读者出生的还早）的故事呢？ 计算领域已经发生了很多变化了，是吧？

[林迪效应](https://en.wikipedia.org/wiki/Lindy_effect) 是指如一个技术或者一个想法之类的一些不易腐烂的东西的未来预期寿命与他们的当前存活时间成正比。 太长不看版——老技术还会存在。

如果你不相信的话，看看这些：

* [oh-my-zsh](https://github.com/ohmyzsh/ohmyzsh) 在 GitHub 上已经快有了 100,000 个 星星了
* [《命令行中的数据科学》](https://www.datascienceatthecommandline.com/)
* [命令行工具能够比你的 Hadoop 集群快235倍](https://adamdrake.com/command-line-tools-can-be-235x-faster-than-your-hadoop-cluster.html)
* ...

现在你应该被说服了吧， 让我们来讨论以下怎么使你的 Go 命令行程序变得友好。

## 设计

当你在写命令行应用程序的时候， 试试遵守 基础的 [Unix 哲学](http://www.catb.org/esr/writings/taoup/html/ch01s06.html)

* 模块性规则： 编写通过清晰的接口连接起来的简单的部门
* 组合性规则： 设计可以和其他程序连接起来的程序
* 缄默性规则：当一个程序没有什么特别的事情需要说的时候，它就应该闭嘴

这些规则能指导你编写做一件事的小程序。

* 用户需要从 REST API 中读取数据的功能 ？ 他们会将 `curl` 命令的输出通过管道输入到你的程序中
* 用户只想要前 n 个结果 ？ 他们可以把你的程序的输出结果通过管道输入到 `head` 命令中
* 用户指向要第二列数据 ？ 如果你的输出结果以 tab 为分割， 他们就可以把你的输出通过管道输入到 `cut` 或 `awk` 命令

如果你没有遵从上述要求 ， 没有结构性的组织你的命令行接口 ， 你可能会像下面这种情况一样的停止。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/writing-friendly-command-line-application/1.png)

## 帮助

让我们来假定你们团队有一个叫做 `nuke-db` 的实用工具 。 你忘了怎么调用它然后你：

```shell
$ ./nuke-db --help
database nuked	(译者注：也就说本意想看使用方式，但却直接执行了)
```

OMG！

使用 [flag 库](https://golang.org/pkg/flag/) ，你可以用额外的两行代码添加对于 `--help` 的支持。

```go
package main

import (
	"flag" // extra line 1
	"fmt"
)

func main() {
	flag.Parse() // extra line 2
	fmt.Println("database nuked")
}
```

现在你的程序运行起来是这个样子：

```shell
$ ./nuke-db --help
Usage of ./nuke-db:
$ ./nuke-db
database nuked
```

如果你想提供更多的帮助 ， 使用 `flag.Usage`

```go
package main

import (
	"flag"
	"fmt"
	"os"
)

var usage = `usage: %s [DATABASE]

Delete all data and tables from DATABASE.
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	fmt.Println("database nuked")
}
```

现在 ：

```shell
$ ./nuke-db --help
usage: ./nuke-db [DATABASE]

Delete all data and tables from DATABASE.
```

## 结构化输出

纯文本是通用的接口。 然而，当输出变得复杂的时候， 对机器来说处理格式化的输出会更容易。最普遍的一种格式当然是 JSON。

一个打印的好的方式不是使用 `fmt.Printf` 而是使用你自己的既适合于文本也适合于 JSON 的打印函数。让我们来看一个例子：

```go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var jsonOut bool
	flag.BoolVar(&jsonOut, "json", false, "output in JSON format")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("error: wrong number of arguments")
	}

	write := writeText
	if jsonOut {
		write = writeJSON
	}

	fi, err := os.Stat(flag.Arg(0))
	if err != nil {
		log.Fatalf("error: %s\n", err)
	}

	m := map[string]interface{}{
		"size":     fi.Size(),
		"dir":      fi.IsDir(),
		"modified": fi.ModTime(),
		"mode":     fi.Mode(),
	}
	write(m)
}

func writeText(m map[string]interface{}) {
	for k, v := range m {
		fmt.Printf("%s: %v\n", k, v)
	}
}

func writeJSON(m map[string]interface{}) {
	m["mode"] = m["mode"].(os.FileMode).String()
	json.NewEncoder(os.Stdout).Encode(m)
}
```

那么

```shell
$ ./finfo finfo.go
mode: -rw-r--r--
size: 783
dir: false
modified: 2019-11-27 11:49:03.280857863 +0200 IST
$ ./finfo -json finfo.go
{"dir":false,"mode":"-rw-r--r--","modified":"2019-11-27T11:49:03.280857863+02:00","size":783}

```

## 处理

有些操作是比较耗时的，一个是他们更快的方法不是优化代码，而是显示一个旋转加载符或者进度条。不要不信我，这有一个来自 [Nielsen 的研究](https://www.nngroup.com/articles/progress-indicators/) 的引用

> 看到运动的进度条的人们会有更高的满意度体验而且比那些得不到任何反馈的人平均多出三倍的愿意等待时间。

## 旋转加载

添加一个旋转加载不需要任何特别的库

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var spinChars = `|/-\`

type Spinner struct {
	message string
	i       int
}

func NewSpinner(message string) *Spinner {
	return &Spinner{message: message}
}

func (s *Spinner) Tick() {
	fmt.Printf("%s %c \r", s.message, spinChars[s.i])
	s.i = (s.i + 1) % len(spinChars)
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func main() {
	flag.Parse()
	s := NewSpinner("working...")
	for i := 0; i < 100; i++ {
		if isTTY() {
			s.Tick()
		}
		time.Sleep(100 * time.Millisecond)
	}

}

```

运行它你就能看到一个小的旋转加载在运动。

## 进度条

对于进度条， 你可能需要一个额外的库如 `github.com/cheggaaa/pb/v3`

```go
package main

import (
	"flag"
	"time"

	"github.com/cheggaaa/pb/v3"
)

func main() {
	flag.Parse()
	count := 100
	bar := pb.StartNew(count)
	for i := 0; i < count; i++ {
		time.Sleep(100 * time.Millisecond)
		bar.Increment()
	}
	bar.Finish()

}
```

## 结语

现在差不多 2020 年了，命令行应用程序仍然会存在。 它们是自动化的关键，如果写得好，能提供优雅的“类似乐高”的组件来构建复杂的流程。

我希望这篇文章将激励你成为一个命令行之国的好公民。

---
via: https://blog.gopheracademy.com/advent-2019/cmdline/

作者：[Miki Tebeka](https://blog.gopheracademy.com)
译者：[Ollyder](https://github.com/Ollyder)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出