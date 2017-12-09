## 第14部分：Strings
---


欢迎阅读 Go 语言辅导系列教程的第 14 部分。

由于和其他语言相比，字符串在 Go 语言中有着自己特殊的实现，因此在这里需要被特别提出来。

## 什么是字符串？
Go 语言中的字符串是一个字节切片。我们可以用 `""` 来创建一个空字符串。让我们来看一个创建并打印字符串的简单示例。
```java
package main

import (  
    "fmt"
)

func main() {  
    name := "Hello World"
    fmt.Println(name)
}
```
[运行](https://play.golang.org/p/o9OVDgEMU0)

上面的程序将会输出 `Hello World`

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
[运行](https://play.golang.org/p/XbJO2b0ZDW)

上面程序的第8行，`len(s)` 返回字符串中字节的数量，然后我们用了一个 for 循环以 16 进制的形式打印出来。`%x` 格式限定符用于指定16进制编码。上面程序的输出是这样的 `48 65 6c 6c 6f 20 57 6f 72 6c 64`. 这些打印出来的字符是 "Hello World" 以 Unicode UTF-8 编码的结果。为了更好的理解go 中的字符串，需要对 Unicode 和 UTF-8 有基础的理解。我推荐阅读一下[https://naveenr.net/unicode-character-set-and-utf-8-utf-16-utf-32-encoding/](https://naveenr.net/unicode-character-set-and-utf-8-utf-16-utf-32-encoding/)来理解一下什么是 Unicode和UTF-8。

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
[运行](https://play.golang.org/p/Jss0HG1q80)

在 `printChars` 方法(第 16行中)中，`%c` 格式限定符用于打印字符串的字符。这个程序输出结果是：
```
48 65 6c 6c 6f 20 57 6f 72 6c 64  
H e l l o   W o r l d  
```
尽管上面的程序用来获取字符串的每一个字符看上去是一种合理的方式，它也有很严重的 bug。让我拆解这个代码来看看我们做错了什么。

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
[运行](https://play.golang.org/p/UQOVvRVaFH)

上面代码输出的结果是：
```
48 65 6c 6c 6f 20 57 6f 72 6c 64  
H e l l o   W o r l d  
53 65 c3 b1 6f 72  
S e Ã ± o r  
```

在上面的成程序的第28行，我们尝试输出 `Señor `的字符，但输出了错误的 S e Ã ± o r . 为什么程序分割 `Hello World` 时表现完美，但分割 `Señor` 就出现了错误呢？这是因为 `ñ` 的Unicode码是 `U+00F1`. 它的 UTF-8 编码占用了两个字节 c3 和 b1。我们打印字符时假定一个一个字符的编码只有一个字节长是错误的。在UTF-8中一个编码可以占用超过一个字节的空间。那么我们如何解决这个问题呢？这就是 `rune` 拯救我们的地方。

## rune
rune 是Go 语言的内建类型，它也是int32的别称。在Go语言中，rune相当于一个Unicode编码的值。无论一个字节编码值占用几个字节，都可以用一个rune来表达。让我们修改一下上面的程序，用rune来打印字符。
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
[运行](https://play.golang.org/p/t4z-f8I-ih)

在上面代码的第14行，字符串被转化为一个rune切片。然后我们循环打印字符。程序的输出结果是
```
48 65 6c 6c 6f 20 57 6f 72 6c 64  
H e l l o   W o r l d 

53 65 c3 b1 6f 72  
S e ñ o r  
```
上面的输出结果非常完美，就是我们想要的结果:)。

## 字符串的 for range 循环
上面的程序是遍历字符串的非常完美的方式了。但是Go给我们提供了一种用 `for range` 循环的更简单的方法。
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
[运行](https://play.golang.org/p/BPpQ0dZr8W)

在上面程序中的第8行，使用 `for range` 循环遍历了字符串。循环返回的字节位置是当前rune的起始位置。程序的输出结果为：
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
[运行](https://play.golang.org/p/Vr9pf8X8xO)

上面的程序中 `byteSlice` 包含字符串 `Café` 用UTF-8编码后的16进制字节。程序输出结果是`Café`。

那么如果我们有16进制对应的10进制的值。上面的程序还能工作吗？让我们来试一试：
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
[运行](https://play.golang.org/p/jgsRowW6XN)

上面程序的输出结果也是`Café`

## 使用 rune 切片构建字符串
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
[运行](https://play.golang.org/p/m8wTMOpYJP)

在上面的程序中 `runeSlice` 包含字符串`Señor`的16进制的Unicode编码值。这个程序将会输出`Señor`。

`字符串的长度`
[utf8 package](https://golang.org/pkg/unicode/utf8/#RuneCountInString) 包中的 `func RuneCountInString(s string) (n int)` 方法用来获取字符串的长度。这个方法传入一个字符串参数然后返回字符串中的rune的数量。
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
[运行](https://play.golang.org/p/QGYlHmF7tn)

上面程序的输出结果是：
```
length of Señor is 5  
length of Pets is 4  
```
## 字符串是不可变的
Go 中的字符串是不可变的。一旦一个字符串被创建，那么它将无法被修改。
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
[运行](https://play.golang.org/p/bv4SlSd_hp)

在上面程序中的第8行，我们尝试修改这个字符串中的第一个字符为 `'a'`，由于字符串是不可变的，因此这个操作是不被允许的。所以这个程序抛出了一个错误 *main.go:8: cannot assign to s[0]*

为了变通字符串的不可变性，字符串可以转化为一个rune切片。然后这个切片可以进行任何想要的改变，然后再转化为一个字符串。
[运行](https://play.golang.org/p/GL1cm17IP1)

在上面程序的第7行，`mutate` 方法接收一个 rune 切片参数，然后它将切片的第一个元素修改为 `'a'`，然后将rune切片转化为字符串并返回。这个方法在程序的第13行被调用。`h` 被转化为一个字符串切片然后传递给 `mutate`。这个程序输出 `aello`。

我已经在github上创建了一个程序包含了我们讨论的所有东西。你可以在这[下载](https://github.com/golangbot/stringsexplained)它。
这就是关于字符串的东西。祝你愉快。

----------------

via: https://golangbot.com/strings/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[译者ID](https://github.com/jliu666)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出