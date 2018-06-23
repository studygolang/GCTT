已发布：https://studygolang.com/articles/13285

# Go 语言中的包是怎么工作的

自从我开始用 Go 写代码以来，如何组织好代码并用好 package 关键字对我来说一直是个迷样的难题。package 关键字类似于 C# 中的命名空间，但是它的约定却是将 package 名字与目录结构绑定在一起。

Go 语言有一个网页试图解释如何编写 Go 代码。

http://golang.org/doc/code.html

当我开始用 Go 编程时，这是我最开始读的资料之一。可能因为之前一直在 Visual Studio 中工作，代码被解决方案和项目打包的很好，这个文档中的内容对当时的我来说，完全没法读懂。基于文件系统的目录来工作曾让我认为这是个疯狂的想法。但现在我喜欢上这种简单的方式了，不过可能需要花上一段时间你才会发觉这个方案的合理之处。

“如何编写 Go 代码”从工作空间的概念讲起。把这个理解为你的项目的根目录。如果你使用Visual Studio，那么它应该是解决方案或者项目文件所在的地方。然后在你的工作空间里面，你需要创建一个名为src的子目录。这个目录是必须的，这样 Go 的工具才能正确运行。在 src 目录里你可以按照个人喜好自由的组织你的代码。但是你需要了解 Go 团队为包和源代码制定的约定，不然你可能要重构你的代码行。

在我的机器上，我创建了一个工作空间叫 Test ，在其下建立了必要的 src 子目录。这是创建项目的第一步。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-work/Screen+Shot+2013-07-28+at+10.03.44+AM.png)

然后在 LiteIDE 中打开Test目录（也就是我的工作空间），然后创建如下的子目录以及空的 Go 源文件。

首先，为我们创建的应用建立一个子目录。 main  函数所在的文件夹名称就是编译后的可执行文件的名称。在我们这个项目中，main.go 包含了main函数，并且位于myprogram目录下。这意味着我们的可执行文件名就叫myprogram。

其它 src 目录下的子目录包含了项目中的包。按照约定目录的名称就是这个目录下源文件所属的包的名称，在我这个项目中新的包命名为 samplepkg 和 subpkg ，源文件的名称可以自由命名。

下一步，创建好同名的包文件夹和空的 Go 源文件。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-work/Screen+Shot+2013-07-28+at+10.10.42+AM.png)

如果你不把工作空间所属文件夹加入 GOPATH 我们会碰到一些问题。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-work/Screen+Shot+2013-07-28+at+10.07.09+AM.png)

我花了点时间才意识到自定义文件夹（ Custom Directory ）是一个文本框，所以你可以直接编辑那些文件夹的值，系统的 GOPATH 是只读的。

Go 的设计者在命名他们的包和源文件时已经做了一些事情。所有的文件和目录的名称都是小写的，并且目录名不需要用下划线将单词分隔开，所有包的名字与目录相同，一个目录下的所有源文件属于与目录同名的包中。

看看 Go 源码目录中的一些标准库包：

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-work/Screen+Shot+2013-07-28+at+10.09.25+AM.png)

`bufio` 和 `builtin` 的目录是目录命名约定最好的例子。它们其实也可能被命名为 `buf_io` 和 `built_in`。

再看看 Go 源码目录中源文件的名字。

注意到有些文件的名字中使用了下划线。当文件包含测试代码或者特定为某种平台使用时，就需要使用下划线。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-work/Screen+Shot+2013-07-28+at+10.17.49+AM.png)

一个不常用的约定是，将文件命名为目录的名字。在 bufio 包中是遵守了这个约定的，但是这是一个不常被遵循的约定。

在 fmt 包中你会发现并没有一个叫 fmt.go 的源文件。我个人也喜欢把源文件的命名和目录的命名区分开来。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-work/Screen+Shot+2013-07-28+at+10.20.36+AM.png)

最后，打开 doc.go，format.go，print.go 和 scan.go，它们都在 fmt 包中被声明。

让我们看看sample。go的代码：

```go
package samplepkg

import (
    "fmt"
)

type Sample struct {
    Name string
}

func New(name string) (sample * Sample) {
    return &Sample{
        Name: name,
    }
}

func (sample * Sample) Print() {
    fmt.Println("Sample Name:", sample.Name)
}
```
这段代码没啥用处，但是却可以让我们看到两个重要的约定。首先，注意这个包的名称与目录的名称相同。第二，有一个叫 New 的函数。

在 Go 的约定中，用于创建一个核心类型或者给应用开发者使用的不同类型的函数就命名为 New。我们看看在 log.go，bufio.go 和 cypto.go 文件中 New 函数是如何定义和实现的。

```go
log.go
// New creates a new Logger. The out variable sets the
// destination to which log data will be written.
// The prefix appears at the beginning of each generated log line.
// The flag argument defines the logging properties.
func New(out io.Writer, prefix string, flag int) * Logger {
    return &Logger{out: out, prefix: prefix, flag: flag}
}

bufio.go
// NewReader returns a new Reader whose buffer has the default size.
func NewReader(rd io.Reader) * Reader {
    return NewReaderSize(rd, defaultBufSize)
}

crypto.go
// New returns a new hash.Hash calculating the given hash function. New panics
// if the hash function is not linked into the binary.
func (h Hash) New() hash.Hash {
    if h > 0 && h < maxHash {
        f := hashes[h]
        if f != nil {
            return f()
        }
    }
    panic("crypto: requested hash function is unavailable")
}
```

因为每个包起到了命名空间的作用，每个包可以有它们自己版本的 New 函数实现。在 bufio.go 中可以创建多种类型，所以并没有一个单独的 New 函数，你可以看到类似 NewReader 和 NewWriter 这样的函数。

再看看 sub.go 的代码：

```go
package subpkg

import (
    "fmt"
)

type Sub struct {
    Name string
}

func New(name string) (sub * Sub) {
    return &Sub{
        Name: name,
    }
}

func (sub * Sub) Print() {
    fmt.Println("Sub Name:", sub.Name)
}
```

代码是基本相同的，除了我们的核心类型改名成了 Sub 。包的名称与子目录名相同，并且 New 返回一个 Sub 类型的引用。

现在我们可以使用这个已经定义好，并且实现好了的包了。

再看看 main.go 中的代码：

```go
package main

import (
    "samplepkg"
    "samplepkg/subpkg"
)

func main() {
    sample := samplepkg.New("Test Sample Package")
    sample.Print()

    sub := subpkg.New("Test Sub Package")
    sub.Print()
}
```

因为我们的 GOPATH 指向了工作空间目录，这个项目中是 /User/bill/Spaces/Test，我们的 import 指令就是从这个目录开始引用其他包的。这里我们引用了当前目录结构下两个包

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-work/Screen+Shot+2013-07-28+at+10.23.25+AM.png)

接下来，我们分别调用每个包中的 New 函数，并创建对应的变量。

现在编译并且运行程序，你可以看到可执行文件就叫 myprogram。

一旦你的程序已经准备分发了，你可以运行 install 命名。

install 命令会在工作空间中创建 bin 和 pkg 文件夹。注意最终的执行文件放在 bin 文件夹下。

编译好的包放在 pkg 文件夹下，这个目录下创建了一个目标架构的文件夹，并且把源码目录下的目录结构都复制一份在此文件夹下。

这些编译好的包都存在，于是go工具可以避免不必要的重新编译。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-work/Screen+Shot+2013-07-28+at+10.24.16+AM.png)

在“如何编写Go代码”文章中最后部分讲的问题是，Go 工具在以后编译代码时会忽略所有的 .a 文件。没有源文件你没法编译你的应用。我还没有找到任何文档解释这些 .a 文件如何直接参与 Go 程序构建的。如果有人知道还请不吝赐教。

最后，我们最好遵循 Go 设计者制定的这些约定，读 Go 的源码是最好的了解这些约定的方法。有很多人为开源社区写代码，如果我们都遵循相同的约定，我们可以提高代码的兼容性和可读性。当有疑问时 在 /usr/local/go/src/pkg 中挖掘答案吧。

一如既往的，我希望这篇文章能帮助你更好的理解 Go 语言。

---

via: https://www.ardanlabs.com/blog/2013/07/how-packages-work-in-go-language.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[MoodWu](https://github.com/MoodWu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
