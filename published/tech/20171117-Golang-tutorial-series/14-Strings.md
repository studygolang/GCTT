已发布： https://studygolang.com/articles/12261

# 第 14 部分：字符串

欢迎阅读 [Golang 系列教程](https://studygolang.com/subject/2)第 14 部分。

由于和其他语言相比，字符串在 Go 语言中有着自己特殊的实现，因此在这里需要被特别提出来。

## 什么是字符串？

Go 语言中的字符串是一个字节切片。把内容放在双引号""之间，我们可以创建一个字符串。让我们来看一个创建并打印字符串的简单示例。

```go
package main

import (
    "fmt"
)

func main() {
    name := "Hello World"
    fmt.Println(name)
}
```
[在线运行程序](https://play.golang.org/p/o9OVDgEMU0)

上面的程序将会输出 `Hello World`。

Go 中的字符串是兼容 Unicode 编码的，并且使用 UTF-8 进行编码。

## 单独获取字符串的每一个字节

由于字符串是一个字节切片，所以我们可以获取字符串的每一个字节。

```go
package main

import (
    "fmt"
)

func printBytes(s string) {
    for i:= 0; i < len(s); i++ {
        fmt.Printf("%x ", s[i])
    }
}

func main() {
    name := "Hello World"
    printBytes(name)
}
```
[在线运行程序](https://play.golang.org/p/XbJO2b0ZDW)

上面程序的第 8 行，**`len(s)` 返回字符串中字节的数量**，然后我们用了一个 for 循环以 16 进制的形式打印这些字节。`%x` 格式限定符用于指定 16 进制编码。上面的程序输出 `48 65 6c 6c 6f 20 57 6f 72 6c 64`。这些打印出来的字符是 "Hello World" 以 [Unicode UTF-8 编码](https://mothereff.in/utf-8#Hello%20World)的结果。为了更好的理解 go 中的字符串，需要对 Unicode 和 UTF-8 有基础的理解。我推荐阅读一下 [https://naveenr.net/unicode-character-set-and-utf-8-utf-16-utf-32-encoding/](https://naveenr.net/unicode-character-set-and-utf-8-utf-16-utf-32-encoding/) 来理解一下什么是 Unicode 和 UTF-8。

让我们稍微修改一下上面的程序，让它打印字符串的每一个字符。

```go
package main

import (
    "fmt"
)

func printBytes(s string) {
    for i:= 0; i < len(s); i++ {
        fmt.Printf("%x ", s[i])
    }
}


func printChars(s string) {
    for i:= 0; i < len(s); i++ {
        fmt.Printf("%c ",s[i])
    }
}

func main() {
    name := "Hello World"
    printBytes(name)
    fmt.Printf("\n")
    printChars(name)
}
```
[在线运行程序](https://play.golang.org/p/Jss0HG1q80)

在 `printChars` 方法(第 16 行中)中，`%c` 格式限定符用于打印字符串的字符。这个程序输出结果是：

```
48 65 6c 6c 6f 20 57 6f 72 6c 64
H e l l o   W o r l d
```

上面的程序获取字符串的每一个字符，虽然看起来是合法的，但却有一个严重的 bug。让我拆解这个代码来看看我们做错了什么。

```go
package main

import (
    "fmt"
)

func printBytes(s string) {
    for i:= 0; i < len(s); i++ {
        fmt.Printf("%x ", s[i])
    }
}

func printChars(s string) {
    for i:= 0; i < len(s); i++ {
        fmt.Printf("%c ",s[i])
    }
}

func main() {
    name := "Hello World"
    printBytes(name)
    fmt.Printf("\n")
    printChars(name)
    fmt.Printf("\n")
    name = "Señor"
    printBytes(name)
    fmt.Printf("\n")
    printChars(name)
}
```
[在线运行程序](https://play.golang.org/p/UQOVvRVaFH)

上面代码输出的结果是：

```
48 65 6c 6c 6f 20 57 6f 72 6c 64
H e l l o   W o r l d
53 65 c3 b1 6f 72
S e Ã ± o r
```

在上面程序的第 28 行，我们尝试输出 **Señor** 的字符，但却输出了错误的 **S e Ã ± o r**。 为什么程序分割 `Hello World` 时表现完美，但分割 `Señor` 就出现了错误呢？这是因为 `ñ` 的 Unicode 代码点（Code Point）是 `U+00F1`。它的 [UTF-8 编码](https://mothereff.in/utf-8#%C3%B1)占用了 c3 和 b1 两个字节。它的 UTF-8 编码占用了两个字节 c3 和 b1。而我们打印字符时，却假定每个字符的编码只会占用一个字节，这是错误的。在 UTF-8 编码中，一个代码点可能会占用超过一个字节的空间。那么我们该怎么办呢？rune 能帮我们解决这个难题。

## rune

rune 是 Go 语言的内建类型，它也是 int32 的别称。在 Go 语言中，rune 表示一个代码点。代码点无论占用多少个字节，都可以用一个 rune 来表示。让我们修改一下上面的程序，用 rune 来打印字符。

```go
package main

import (
    "fmt"
)

func printBytes(s string) {
    for i:= 0; i < len(s); i++ {
        fmt.Printf("%x ", s[i])
    }
}

func printChars(s string) {
    runes := []rune(s)
    for i:= 0; i < len(runes); i++ {
        fmt.Printf("%c ",runes[i])
    }
}

func main() {
    name := "Hello World"
    printBytes(name)
    fmt.Printf("\n")
    printChars(name)
    fmt.Printf("\n\n")
    name = "Señor"
    printBytes(name)
    fmt.Printf("\n")
    printChars(name)
}
```
[在线运行程序](https://play.golang.org/p/t4z-f8I-ih)

在上面代码的第 14 行，字符串被转化为一个 rune 切片。然后我们循环打印字符。程序的输出结果是

```
48 65 6c 6c 6f 20 57 6f 72 6c 64
H e l l o   W o r l d

53 65 c3 b1 6f 72
S e ñ o r
```
上面的输出结果非常完美，就是我们想要的结果:)。

## 字符串的 for range 循环

上面的程序是一种遍历字符串的好方法，但是 Go 给我们提供了一种更简单的方法来做到这一点：使用 **for range** 循环。

```go
package main

import (
    "fmt"
)

func printCharsAndBytes(s string) {
    for index, rune := range s {
        fmt.Printf("%c starts at byte %d\n", rune, index)
    }
}

func main() {
    name := "Señor"
    printCharsAndBytes(name)
}
```
[在线运行程序](https://play.golang.org/p/BPpQ0dZr8W)

在上面程序中的第8行，使用 `for range` 循环遍历了字符串。循环返回的是是当前 rune 的字节位置。程序的输出结果为：

```
S starts at byte 0
e starts at byte 1
ñ starts at byte 2
o starts at byte 4
r starts at byte 5
```
从上面的输出中可以清晰的看到 `ñ` 占了两个字节:)。

## 用字节切片构造字符串

```go
package main

import (
    "fmt"
)

func main() {
    byteSlice := []byte{0x43, 0x61, 0x66, 0xC3, 0xA9}
    str := string(byteSlice)
    fmt.Println(str)
}
```
[在线运行程序](https://play.golang.org/p/Vr9pf8X8xO)

上面的程序中 `byteSlice` 包含字符串 `Café` 用 UTF-8 编码后的 16 进制字节。程序输出结果是 `Café`。

如果我们把 16 进制换成对应的 10 进制值会怎么样呢？上面的程序还能工作吗？让我们来试一试：

```go
package main

import (
    "fmt"
)

func main() {
    byteSlice := []byte{67, 97, 102, 195, 169}//decimal equivalent of {'\x43', '\x61', '\x66', '\xC3', '\xA9'}
    str := string(byteSlice)
    fmt.Println(str)
}
```
[在线运行程序](https://play.golang.org/p/jgsRowW6XN)

上面程序的输出结果也是`Café`

## 用 rune 切片构造字符串

```go
package main

import (
    "fmt"
)

func main() {
    runeSlice := []rune{0x0053, 0x0065, 0x00f1, 0x006f, 0x0072}
    str := string(runeSlice)
    fmt.Println(str)
}
```
[在线运行程序](https://play.golang.org/p/m8wTMOpYJP)

在上面的程序中 `runeSlice` 包含字符串 `Señor`的 16 进制的 Unicode 代码点。这个程序将会输出`Señor`。

## 字符串的长度

[utf8 package](https://golang.org/pkg/unicode/utf8/#RuneCountInString) 包中的 `func RuneCountInString(s string) (n int)` 方法用来获取字符串的长度。这个方法传入一个字符串参数然后返回字符串中的 rune 的数量。

```go
package main

import (
    "fmt"
    "unicode/utf8"
)

func length(s string) {
    fmt.Printf("length of %s is %d\n", s, utf8.RuneCountInString(s))
}
func main() {
    word1 := "Señor"
    length(word1)
    word2 := "Pets"
    length(word2)
}
```
[在线运行程序](https://play.golang.org/p/QGYlHmF7tn)

上面程序的输出结果是：

```
length of Señor is 5
length of Pets is 4
```

## 字符串是不可变的

Go 中的字符串是不可变的。一旦一个字符串被创建，那么它将无法被修改。

```go
package main

import (
    "fmt"
)

func mutate(s string)string {
    s[0] = 'a'//any valid unicode character within single quote is a rune
    return s
}
func main() {
    h := "hello"
    fmt.Println(mutate(h))
}
```
[在线运行程序](https://play.golang.org/p/bv4SlSd_hp)

在上面程序中的第 8 行，我们试图把这个字符串中的第一个字符修改为 'a'。由于字符串是不可变的，因此这个操作是非法的。所以程序抛出了一个错误 **main.go:8: cannot assign to s[0]**。

为了修改字符串，可以把字符串转化为一个 rune 切片。然后这个切片可以进行任何想要的改变，然后再转化为一个字符串。

```go
package main

import (
    "fmt"
)

func mutate(s []rune) string {
    s[0] = 'a'
    return string(s)
}
func main() {
    h := "hello"
    fmt.Println(mutate([]rune(h)))
}
```
[在线运行程序](https://play.golang.org/p/GL1cm17IP1)

在上面程序的第 7 行，`mutate` 函数接收一个 rune 切片参数，它将切片的第一个元素修改为 `'a'`，然后将 rune 切片转化为字符串，并返回该字符串。程序的第 13 行调用了该函数。我们把 `h` 转化为一个 rune 切片，并传递给了 `mutate`。这个程序输出 `aello`。

我已经在 github 上创建了一个程序，里面包含所有我们讨论过的内容。你可以在这[下载](https://github.com/golangbot/stringsexplained)它。

关于字符串的介绍到此为止。祝你愉快。

**上一教程 - [Maps](https://studygolang.com/articles/12251)**

**下一教程 - [指针](https://studygolang.com/articles/12262)**

---

via: https://golangbot.com/strings/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[jliu666](https://github.com/jliu666)
校对：[Noluye](https://github.com/Noluye)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
