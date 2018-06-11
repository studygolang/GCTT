已发布：https://studygolang.com/articles/12324

# Go 语言标准库中最常用的标识符是什么?

这篇文章是来自最新 [justforfunc](http://justforfunc.com/) 中同标题的一段。这个程序的[代码](https://github.com/campoy/justforfunc/blob/master/24-go-scanner/main.go)可以在 [justforfunc 仓库](https://github.com/campoy/justforfunc) 中找到。

## 问题陈述

想象一下，对于下面的代码段，你如何将其中所有的标识符都提取出来。

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, world")
}
```

我们期望可以得到一个包含 **main**, **fmt** 和 **Println** 的列表。

## 标识符到底是什么？

为了回答这个问题， 我们需要了解一下有关计算机语言的理论知识。 但只要一点就足够了，不用担心有多复杂。

计算机语言，是由一系列有效的规则组成的。比如下面这个规则：

```
IfStmt = "if" [ SimpleStmt ";" ] Expression Block [ "else" ( IfStmt | Block ) ] .
```

上面这个规则告诉我们 if 语句在 Go 语言中的样子。**“if”**, **“;”**, 和 **“else”** 是帮助我们理解程序结构的关键词。与此同时，还有 **Expression Block**， **SimpleStmt** 之类的其他规则。

这些规则组成的集合就是语法，你可以在 Go 语言规范中找到它们的详细定义。

这些规则不是简单的由程序的单个字符定义的，而是有一系列 token 组成。 这些token除了像 **if** 和 **else** 这样的原子 token 外， 还有像整数 42，浮点数 4.2 和字符串 “hello” 这样的复合 token， 以及像 **main** 这样的标识符。

但是，我们是怎么知道 main 是一个标识符，而不是一个数字呢？ 原来它也是有专门的规则来定义的。如果你读过 Go 语言规范中的标识符部分，你就会发现如下的规则：
```
identifier = letter { letter | unicode_digit } .
```

在这条规则中，letter 和 unicode_digit 不是　token　而是字符。 所以有了这些规则，就可以写一个程序来逐个字符地分析，一旦检测到一组字符匹配到某一条规则，就 “发射”(emits) 出一个 token。

所以，如果我们以 **fmt.Println** 为例， 它可以产生这些 token：标识符 **fmt**, **“.”**, 以及标识符 **Println**。 这是一个函数调用吗？ 在这里我们还无法确定，而且我们也不关心。它的结构就是一个序列，表明 token 出现的顺序。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/most-common-identifier/1.png)

这种能够将给定的字符序列生成 token 序列的程序被称为扫描器。Go 标准库中的 go/scanner 就自带一个扫描器。它生成的记号定义在 go/token 里。

## 使用 go/scanner

我们已经了解了什么是扫描器，那它如何使用呢？

### 从命令行中读取参数

让我们先从一个简单程序开始，将传给它的参数打印出来：

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage:\n\t%s [files]\n", os.Args[0])
		os.Exit(1)
	}

	for _, arg := range os.Args[1:] {
		fmt.Println(arg)
	}
}
```

接下来，我们需要扫描从参数传进来的文件：需要先创建一个新的扫描器，然后用文件的内容来初始化。

### 打印每个 token

在我们调用 scanner.Scanner 的 Init 方法之前，需要先读取文件内容，然后为每个扫描过的文件创建一个 **token.FileSet** 以便来保存 **token.File**。

扫描器一经初始化，我们就能调用其 Scan 方法来打印 token。 一旦我们得到一个 EOF(End Of File) token，就说明达到文件末尾了。

```go
fs := token.NewFileSet()

for _, arg := range os.Args[1:] {
	b, err := ioutil.ReadFile(arg)
	if err != nil {
		log.Fatal(err)
	}

	f := fs.AddFile(arg, fs.Base(), len(b))
	var s scanner.Scanner
	s.Init(f, b, nil, scanner.ScanComments)

	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		fmt.Println(tok, lit)
	}
}
```

### 统计 token

太棒了，我们已经能够打印出所有的 token 了，但是我们还需要跟踪每个标识符出现的次数，然后按照出现次数排序，并打印出前 5 位。

在 Go 中，实现以上需求的最好的方法是用一个 map，让标识符来做 key， 其出现次数做 value。

每当一个标识符出现一次，计数器就加一。最后，我们将 map 转换为一个能够排序和打印的数组。

```go
counts := make(map[string]int)

// [code removed for clarity]

for {
	_, tok, lit := s.Scan()
	if tok == token.EOF {
		break
	}
	if tok == token.IDENT {
		counts[lit]++
	}
}

// [为了阅读清晰，移除部分代码]

type pair struct {
	s string
	n int
}
pairs := make([]pair, 0, len(counts))
for s, n := range counts {
	pairs = append(pairs, pair{s, n})rm -f
}

sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].n > pairs[j].n
})

for i := 0; i < len(pairs) && i < 5; i++ {
	fmt.Printf("%6d %s\n", pairs[i].n, pairs[i].s)
}
```
为了不影响理解，有些代码被删除了。你可以在[这里](https://github.com/campoy/justforfunc/blob/master/24-go-scanner/main.go)获取完整的源码。

## 哪些是最常用的标识符？

我们来用这个程序分析一下 github.com/golang/go 上的代码：

```bash
$ go install github.com/campoy/justforfunc/24-ast/scanner
$ scanner ~/go/src/**/*.go
 82163 v
 46584 err
 44681 Args
 43371 t
 37717 x
```

在短标识符里，最常用的标识符是字母 **v** 。那我们修改下代码来计算一些长标识符：

```go
for s, n := range counts {
	if len(s) >= 3 {
		pairs = append(pairs, pair{s, n})
	}
}
```

再来一次：

```bash
$ go install github.com/campoy/justforfunc/24-ast/scanner
$ scanner ~/go/src/**/*.go
 46584 err
 44681 Args
 36738 nil
 25761 true
 21723 AddArg
```

果不其然，err 和 nil 是最常见的标识符，毕竟每个程序中都有 if err != nil 这样的语句。 但 Args 出现频度这么高怎么回事？

欲知详情如何，且听下回分解。

---

via: https://medium.com/@francesc/whats-the-most-common-identifier-in-go-s-stdlib-e468f3c9c7d9

作者：[Francesc Campoy](https://medium.com/@francesc)
译者：[kaneg](https://github.com/kaneg)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
