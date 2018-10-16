首发于：https://studygolang.com/articles/14669

# 读取文件

![reading files](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-series/golang-read-files.png)

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 35 篇。

文件读取是所有编程语言中最常见的操作之一。本教程我们会学习如何使用 Go 读取文件。

本教程分为如下小节。

- 将整个文件读取到内存
  - 使用绝对文件路径
  - 使用命令行标记来传递文件路径
  - 将文件绑定在二进制文件中
- 分块读取文件
- 逐行读取文件

## 将整个文件读取到内存

将整个文件读取到内存是最基本的文件操作之一。这需要使用 [`ioutil`](https://golang.org/pkg/io/ioutil/) 包中的 [`ReadFile`](https://golang.org/pkg/io/ioutil/#ReadFile) 函数。

让我们在 Go 程序所在的目录中，读取一个文件。我已经在 GOPATH（译注：原文是 GOROOT，应该是笔误）中创建了文件夹，在该文件夹内部，有一个文本文件 `test.txt`，我们会使用 Go 程序 `filehandling.go` 来读取它。`test.txt` 包含文本 “Hello World. Welcome to file handling in Go”。我的文件夹结构如下：

```
src
    filehandling
        filehandling.go
        test.txt
```

接下来我们来看看代码。

```go
package main

import (
    "fmt"
    "io/ioutil"
)

func main() {
    data, err := ioutil.ReadFile("test.txt")
    if err != nil {
        fmt.Println("File reading error", err)
        return
    }
    fmt.Println("Contents of file:", string(data))
}
```

由于无法在 playground 上读取文件，因此请在你的本地环境运行这个程序。

在上述程序的第 9 行，程序会读取文件，并返回一个字节[切片](https://studygolang.com/articles/12121)，而这个切片保存在 `data` 中。在第 14 行，我们将 `data` 转换为 `string`，显示出文件的内容。

请在 **test.txt** 所在的位置运行该程序。

例如，对于 **linux/mac**，如果 **test.txt** 位于 **/home/naveen/go/src/filehandling**，可以使用下列步骤来运行程序。

```bash
$ cd /home/naveen/go/src/filehandling/
$ go install filehandling
$ workspacepath/bin/filehandling
```

对于 **windows**，如果 **test.txt** 位于 **C:\Users\naveen.r\go\src\filehandling**，则使用下列步骤。

```bash
> cd C:\Users\naveen.r\go\src\filehandling
> go install filehandling
> workspacepath\bin\filehandling.exe
```

该程序会输出：

```bas
Contents of file: Hello World. Welcome to file handling in Go.
```

如果在其他位置运行这个程序（比如 `/home/userdirectory`），会打印下面的错误。

```bash
File reading error open test.txt: The system cannot find the file specified.
```

这是因为 Go 是编译型语言。`go install` 会根据源代码创建一个二进制文件。二进制文件独立于源代码，可以在任何位置上运行。由于在运行二进制文件的位置上没有找到 `test.txt`，因此程序会报错，提示无法找到指定的文件。

有三种方法可以解决这个问题。

1. 使用绝对文件路径
2. 使用命令行标记来传递文件路径
3. 将文件绑定在二进制文件中

让我们来依次介绍。

### 1. 使用绝对文件路径

要解决问题，最简单的方法就是传入绝对文件路径。我已经修改了程序，把路径改成了绝对路径。

```go
package main

import (
    "fmt"
    "io/ioutil"
)

func main() {
    data, err := ioutil.ReadFile("/home/naveen/go/src/filehandling/test.txt")
    if err != nil {
        fmt.Println("File reading error", err)
        return
    }
    fmt.Println("Contents of file:", string(data))
}
```

现在可以在任何位置上运行程序，打印出 `test.txt` 的内容。

例如，可以在我的家目录运行。

```bash
$ cd $HOME
$ go install filehandling
$ workspacepath/bin/filehandling
```

该程序打印出了 `test.txt` 的内容。

看似这是一个简单的方法，但它的缺点是：文件必须放在程序指定的路径中，否则就会出错。

### 2. 使用命令行标记来传递文件路径

另一种解决方案是使用命令行标记来传递文件路径。使用 [flag](https://golang.org/pkg/flag/) 包，我们可以从输入的命令行获取到文件路径，接着读取文件内容。

首先我们来看看 `flag` 包是如何工作的。`flag` 包有一个名为 [`String`](https://golang.org/pkg/flag/#String) 的[函数](https://studygolang.com/articles/11892)。该函数接收三个参数。第一个参数是标记名，第二个是默认值，第三个是标记的简短描述。

让我们来编写程序，从命令行读取文件名。将 `filehandling.go` 的内容替换如下：

```go
package main
import (
    "flag"
    "fmt"
)

func main() {
    fptr := flag.String("fpath", "test.txt", "file path to read from")
    flag.Parse()
    fmt.Println("value of fpath is", *fptr)
}
```

在上述程序中第 8 行，通过 `String` 函数，创建了一个字符串标记，名称是 `fpath`，默认值是 `test.txt`，描述为 `file path to read from`。这个函数返回存储 flag 值的字符串[变量](https://studygolang.com/articles/11756)的地址。

在程序访问 flag 之前，必须先调用 `flag.Parse()`。

在第 10 行，程序会打印出 flag 值。

使用下面命令运行程序。

```bash
wrkspacepath/bin/filehandling -fpath=/path-of-file/test.txt
```

我们传入 `/path-of-file/test.txt`，赋值给了 `fpath` 标记。

该程序输出：

```bash
value of fpath is /path-of-file/test.txt
```

这是因为 `fpath` 的默认值是 `test.txt`。

现在我们知道如何从命令行读取文件路径了，让我们继续完成我们的文件读取程序。

```go
package main
import (
    "flag"
    "fmt"
    "io/ioutil"
)

func main() {
    fptr := flag.String("fpath", "test.txt", "file path to read from")
    flag.Parse()
    data, err := ioutil.ReadFile(*fptr)
    if err != nil {
        fmt.Println("File reading error", err)
        return
    }
    fmt.Println("Contents of file:", string(data))
}
```

在上述程序里，命令行传入文件路径，程序读取了该文件的内容。使用下面命令运行该程序。

```bash
wrkspacepath/bin/filehandling -fpath=/path-of-file/test.txt
```

请将 `/path-of-file/` 替换为 `test.txt` 的真实路径。该程序将打印：

```txt
Contents of file: Hello World. Welcome to file handling in Go.
```

### 3. 将文件绑定在二进制文件中

虽然从命令行获取文件路径的方法很好，但还有一种更好的解决方法。如果我们能够将文本文件捆绑在二进制文件，岂不是很棒？这就是我们下面要做的事情。

有很多[包](https://studygolang.com/articles/11893)可以帮助我们实现。我们会使用 [packr](https://github.com/gobuffalo/packr)，因为它很简单，并且我在项目中使用它时，没有出现任何问题。

第一步就是安装 `packr` 包。

在命令提示符中输入下面命令，安装 `packr` 包。

```bash
go get -u github.com/gobuffalo/packr/...
```

`packr` 会把静态文件（例如 `.txt` 文件）转换为 `.go` 文件，接下来，`.go` 文件会直接嵌入到二进制文件中。`packer` 非常智能，在开发过程中，可以从磁盘而非二进制文件中获取静态文件。在开发过程中，当仅仅静态文件变化时，可以不必重新编译。

我们通过程序来更好地理解它。用以下内容来替换 `handling.go` 文件。

```go
package main

import (
    "fmt"

    "github.com/gobuffalo/packr"
)

func main() {
    box := packr.NewBox("../filehandling")
    data := box.String("test.txt")
    fmt.Println("Contents of file:", data)
}
```

在上面程序的第 10 行，我们创建了一个新盒子（New Box）。盒子表示一个文件夹，其内容会嵌入到二进制中。在这里，我指定了 `filehandling` 文件夹，其内容包含 `test.txt`。在下一行，我们读取了文件内容，并打印出来。

在开发阶段时，我们可以使用 `go install` 命令来运行程序。程序可以正常运行。`packr` 非常智能，在开发阶段可以从磁盘加载文件。

使用下面命令来运行程序。

```bash
go install filehandling
workspacepath/bin/filehandling
```

该命令可以在其他位置运行。`packr` 很聪明，可以获取传递给 `NewBox` 命令的目录的绝对路径。

该程序会输出：

```txt
Contents of file: Hello World. Welcome to file handling in Go.
```

你可以试着改变 `test.txt` 的内容，然后再运行 `filehandling`。可以看到，无需再次编译，程序打印出了 `test.txt` 的更新内容。完美！:)

现在我们来看看如何将 `test.txt` 打包到我们的二进制文件中。我们使用 `packr` 命令来实现。

运行下面的命令：

```bash
packr install -v filehandling
```

它会打印：

```bash
building box ../filehandling
packing file filehandling.go
packed file filehandling.go
packing file test.txt
packed file test.txt
built box ../filehandling with ["filehandling.go" "test.txt"]
filehandling
```

该命令将静态文件绑定到了二进制文件中。

在运行上述命令之后，使用命令 `workspacepath/bin/filehandling` 来运行程序。程序会打印出 `test.txt` 的内容。于是从二进制文件中，我们读取了 `test.txt` 的内容。

如果你不知道文件到底是由二进制还是磁盘来提供，我建议你删除 `test.txt`，并在此运行 `filehandling` 命令。你将看到，程序打印出了 `test.txt` 的内容。太棒了:D。我们已经成功将静态文件嵌入到了二进制文件中。

## 分块读取文件

在前面的章节，我们学习了如何把整个文件读取到内存。当文件非常大时，尤其在 RAM 存储量不足的情况下，把整个文件都读入内存是没有意义的。更好的方法是分块读取文件。这可以使用 [bufio](https://golang.org/pkg/bufio) 包来完成。

让我们来编写一个程序，以 3 个字节的块为单位读取 `test.txt` 文件。如下所示，替换 `filehandling.go` 的内容。

```go
package main

import (
    "bufio"
    "flag"
    "fmt"
    "log"
    "os"
)

func main() {
    fptr := flag.String("fpath", "test.txt", "file path to read from")
    flag.Parse()

    f, err := os.Open(*fptr)
    if err != nil {
        log.Fatal(err)
    }
    defer func() {
        if err = f.Close(); err != nil {
            log.Fatal(err)
        }
    }()
    r := bufio.NewReader(f)
    b := make([]byte, 3)
    for {
        _, err := r.Read(b)
        if err != nil {
            fmt.Println("Error reading file:", err)
            break
        }
        fmt.Println(string(b))
    }
}
```

在上述程序的第 15 行，我们使用命令行标记传递的路径，打开文件。

在第 19 行，我们延迟了文件的关闭操作。

在上面程序的第 24 行，我们新建了一个缓冲读取器（buffered reader）。在下一行，我们创建了长度和容量为 3 的字节切片，程序会把文件的字节读取到切片中。

第 27 行的 `Read` [方法](https://studygolang.com/articles/12264)会读取 len(b) 个字节（达到 3 字节），并返回所读取的字节数。当到达文件最后时，它会返回一个 EOF 错误。程序的其他地方比较简单，不做解释。

如果我们使用下面命令来运行程序：

```bash
$ go install filehandling
$ wrkspacepath/bin/filehandling -fpath=/path-of-file/test.txt
```

会得到以下输出：

```bash
Hel
lo
Wor
ld.
 We
lco
me
to
fil
e h
and
lin
g i
n G
o.
Error reading file: EOF
```

## 逐行读取文件

本节我们讨论如何使用 Go 逐行读取文件。这可以使用 [bufio](https://golang.org/pkg/bufio/) 来实现。

请将 `test.txt` 替换为以下内容。

```
Hello World. Welcome to file handling in Go.
This is the second line of the file.
We have reached the end of the file.
```

逐行读取文件涉及到以下步骤。

1. 打开文件；
2. 在文件上新建一个 scanner；
3. 扫描文件并且逐行读取。

将 `filehandling.go` 替换为以下内容。

```go
package main

import (
    "bufio"
    "flag"
    "fmt"
    "log"
    "os"
)

func main() {
    fptr := flag.String("fpath", "test.txt", "file path to read from")
    flag.Parse()

    f, err := os.Open(*fptr)
    if err != nil {
        log.Fatal(err)
    }
    defer func() {
        if err = f.Close(); err != nil {
        log.Fatal(err)
    }
    }()
    s := bufio.NewScanner(f)
    for s.Scan() {
        fmt.Println(s.Text())
    }
    err = s.Err()
    if err != nil {
        log.Fatal(err)
    }
}
```

在上述程序的第 15 行，我们用命令行标记传入的路径，打开文件。在第 24 行，我们用文件创建了一个新的 scanner。第 25 行的 `Scan()` 方法读取文件的下一行，如果可以读取，就可以使用 `Text()` 方法。

当 `Scan` 返回 false 时，除非已经到达文件末尾（此时 `Err()` 返回 `nil`），否则 `Err()` 就会返回扫描过程中出现的错误。

如果我使用下面命令来运行程序：

```bash
$ go install filehandling
$ workspacepath/bin/filehandling -fpath=/path-of-file/test.txt
```

程序会输出：

```bash
Hello World. Welcome to file handling in Go.
This is the second line of the file.
We have reached the end of the file.
```

本教程到此结束。希望你能喜欢，祝你愉快。

**上一教程** - [反射](https://studygolang.com/articles/13178)

---

via: https://golangbot.com/read-files/

作者：[Naveen Ramanathan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
