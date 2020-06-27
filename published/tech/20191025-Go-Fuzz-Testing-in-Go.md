首发于：https://studygolang.com/articles/28987

# Go 中的模糊（Fuzz）测试

![由 Renee French 创作的原始 Go Gopher 作品，为“ Go 的旅程”创作的插图。](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191025-Go-Fuzz-Testing-in-Go/Illustration.png)

模糊测试（Fuzzing）是一项使用随机数据加载我们程序的测试技术。是[对常规测试的补充](https://docs.google.com/document/d/1N-12_6YBPpF9o4_Zys_E_ZQndmD06wQVAM_0y9nZUIE/edit)，并且使开发者可以发现那些在手工生成的输入下难以发现的 bug。模糊测试在 Go 程序中很容易设置，并且可以适应于几乎所有类型的代码。

## 模糊测试项目

在 Go 社区中两个项目适用于模糊测试：Google 开发的 [gofuzz](https://raw.githubusercontent.com/google/gofuzz) 和 [Dmitry Vyukov](https://raw.githubusercontent.com/dvyukov) 开发的 [go-fuzz](https://raw.githubusercontent.com/dvyukov/go-fuzz)，Dmitry Vyukov 同样为 Google 工作。两个项目都是有用的，同时适用于不同的用法。来逐一了解它们：

- [gofuzz](https://raw.githubusercontent.com/google/gofuzz) 提供了一个可以用随机值填充你的 Go 结构体的包。而你需要做的是编写测试代码，并且调用这个包来获取随机数据。当你想要模糊测试结构化数据的时候，这个包是完美的。这里是使用随机数据对一个结构体进行 50000 次模糊测试的例子，其中指针/切片/map 有 50%的几率被设置为空：

![模糊测试结构化数据](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191025-Go-Fuzz-Testing-in-Go/fuzzing-structured-data.png)

- [go-fuzz](https://raw.githubusercontent.com/dvyukov/go-fuzz) 基于已经在大多数知名的软件或库中发现了上百个 bug 的 [American Fuzzy Lop](http://lcamtuf.coredump.cx/afl/)。Go-Fuzz 会连续运行，并且根据提供的样本生成随机的字符串。之后必须解析这些字符串，并且明确地将其标记为是否可用于测试。任何有趣的生成的数据都会被该工具所报告，这些数据增加了代码的覆盖率或者导致崩溃。该工具十分适合那些管理诸如 XML，JSON，图像等字符串信息的程序。这里是该工具运行以及发行问题的预览，被称为 crasher。

![使用 go-fuzz 进行模糊测试](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191025-Go-Fuzz-Testing-in-Go/fuzzing-with-go-fuzz.png)

每个包都有自己的长处，并且这两个工具至少有一个会适用于你的程序以及你在开发的项目。来了解下第二个工具，有着更加复杂的工作流程的工具。

## 通过例子了解 Go-Fuzz

先从一个借助模糊测试解决 `encoding/xml` 包中[一个 bug](https://raw.githubusercontent.com/golang/go/issues/11112) 的例子开始。这里是该问题的复现步骤：

- 定义用于接收生成数据的 `模糊（Fuzz）` 方法：

```go
// +build gofuzz

package fuzzing

import "encoding/xml"

type X struct {
	D string `xml:",comment"`
}

func FuzzXMLComment(data []byte) int {
	v := new(X)
	if xml.Unmarshal(data, v) != nil {
		return -1
	}
	if _, err := xml.Marshal(v); err != nil {
		panic(err)
	}

	return 1
}
```

- 定义一个会被工具使用的初始*语料*：

```xml
<a>
	<!-- my comment -->
	<b>foo</b>
</a>
```

然后，由于该 bug 已在 Go 1.6 中被合入，确保在你的标准库中还原了提交 [97c859f8da0c85c33d0f29ba5e11094d8e691e87](https://raw.githubusercontent.com/golang/go/commit/97c859f8da0c85c33d0f29ba5e11094d8e691e87)——同样含有这个 bug 的 Go 1.5 与最新版本的 go-fuzz 不兼容。你的迷你项目应该遵循这样的结构：

![模糊测试 encoding/xml](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191025-Go-Fuzz-Testing-in-Go/fuzzing-encoding:xml-structure.png)

你现在可以运行 `go-fuzz-build` 和 `go-fuzz -bin=./main.zip -workdir=.` 来开始模糊测试：

![模糊测试 encoding/xml](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191025-Go-Fuzz-Testing-in-Go/fuzzing-encoding:xml.png)

经过了最初的几秒钟后，go-fuzz 已经发现了一个 crasher（译注：指引起崩溃的数据，这里保留原文），crasher 被保存在 `crasher/` 文件夹中：

![模糊测试期间记录的 crasher](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191025-Go-Fuzz-Testing-in-Go/crasher-recorded-during-fuzzing.png)

crasher 文件包含了引起 panic 的字符串：

```xml
<a><!------></a>
```

`.output` 文件包含了发生的 panic 信息：

```bash
panic: xml: comments must not contain "--"
```

确实，按照 XML 的规范，注释有两个约束。

> 字符串“--”（双连字符）不得在注释中使用。[...] 注意，语法上不允许注释以 `--->` 结束

借助 go-fuzz 这个 panic 的问题已经在标准库中被修复，并且这种问题现在会返回一个 error。现在来深入研究这个包，来了解这个包如何成功找到这个问题。

## Go-Fuzz 工作流程

如之前所见，[go-fuzz](https://raw.githubusercontent.com/dvyukov/go-fuzz) 的工作流程包含两个步骤：

- 通过命令 `go-fuzz-build` 从你代码中定义的指令来构建工具：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191025-Go-Fuzz-Testing-in-Go/go-fuzz-build.png)

由于构建嵌入了 `Fuzz` 方法，如果修改了这些方法，不要忘了运行 `go-fuzz-build`。

- 通过 `go-fuzz -bin=./my-package.zip -workdir=.` 命令，持续运行工具，并且收集有趣的输入和崩溃：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191025-Go-Fuzz-Testing-in-Go/go-fuzz.png)

语料生成是 go-fuzz 的核心重点。[Dmitry Vyukov](https://raw.githubusercontent.com/dvyukov) 在 [GopherCon 2015](https://www.youtube.com/watch?v=a9xrxRsIbSU&t=459s) 上给出了这个核心功能的流程图：

![语料生成的流程图](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20191025-Go-Fuzz-Testing-in-Go/workflow-of-the-corpus-generation.png)

语料生成在初试语料库上循环，并且使用两个方法：

- **mutation 方法**，该方法对语料进行字节上的微小修改，比如消除，插入，重复，交换，翻转等修改。这里是一个发现崩溃前发生的不同突变（mutation）的例子：

```xml
<a><!------

<a><!------>

<a><!------a>

<a><!------/a>

<a><!------</a>

<a><!------></a>
```

- **versifying 方法**，这是最先进的方法。会学习文本的结构（数字，字母，列表键-值，等等。），然后应用不同部分的 mutation 方法。这是一个仅在字符串上突变的先前语料的例子：

```xml
-<!--  /my commentcomment -->
<b>foo</b>
</a>
```

这是标签上突变的另一个例子：

```xml
<>
<!-- my comment -->
<b>foo<:b>
[/a]
```

Go-Fuzz 运行期间主要使用 mutation 方法（占 90%的迭代），但是组合使用两种方法对发现 bug 是有帮助的：

> 对 xml 文本进行 2.5 小时的模糊测试后：
>
> 没有使用验证（versifier）的模糊测试发现了 902 个（有问题的）输入。
>
> 使用了验证的模糊测试发现了 1055 个（有问题的）输入，其中验证方法发现了 83 个。
>
> 验证方法生成了新的输入，并且增加了 25%的模糊验证效率。

这个工作流程十分高效且易于集成。有助于所有处理文本的包，不论是 Go 标准库还是你自己的代码。

## 模糊测试集成

[自 Go 1.5 开始](https://golang.org/doc/go1.5#hardening)，模糊测试被应用于 Go 标准库中，并且已经发现了[超过 200 个 bug](https://raw.githubusercontent.com/dvyukov/go-fuzz#trophies)。然而，尽管一些包已经存在一些 `Fuzz` 函数，比如 `encoding/csv` 或 `image/png`，Go 并没有原生集成模糊测试。是否[让模糊测试成为 Go 的一等公民](https://raw.githubusercontent.com/golang/go/issues/19109)的讨论已在 GitHub 上展开。

就与模糊测试持续集成的有效在线工具而言，两个工具使用 Go 和 Go-fuzz：

- [fuzzit.dev](https://fuzzit.dev/)，
- [fuzzbuzz.io](https://fuzzbuzz.io/)

与 GitHub 集成的话，两者有着差不多的价格。他们有免费的账户可以让你在自己的管道中进行模糊测试。

---

via: https://medium.com/a-journey-with-go/go-fuzz-testing-in-go-deb36abc971f

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[dust347](https://github.com/dust347)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
