已发布：https://studygolang.com/articles/11905

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/bufio-scanner/cover.jpg)

# 深入理解 Go 标准库之 bufio.Scanner

众所周知，[带缓冲的 IO 标准库](https://golang.org/pkg/bufio/) 一直是 Go 中优化读写操作的利器。对于写操作来说，在被发送到 `socket` 或硬盘之前，`IO 缓冲区` 提供了一个临时存储区来存放数据，缓冲区存储的数据达到一定容量后才会被"释放"出来进行下一步存储，这种方式大大减少了写操作或是最终的系统调用被触发的次数，这无疑会在频繁使用系统资源的时候节省下巨大的系统开销。而对于读操作来说，`缓冲 IO` 意味着每次操作能够读取更多的数据，既减少了系统调用的次数，又通过以块为单位读取硬盘数据来更高效地使用底层硬件。本文会更加侧重于讲解 [bufio](https://golang.org/pkg/bufio/) 包中的 [Scanner](https://golang.org/pkg/bufio/#Scanner) 扫描器模块，它的主要作用是把数据流分割成一个个标记并除去它们之间的空格。

```
"foo  bar   baz"
```

如果我们只想得到上面字符串中的单词，那么扫描器能帮我们按顺序检索出 "foo"，"bar" 和 "baz" 这三个单词( [查看源码](https://play.golang.org/p/_GKmSMZmWZ) )

```go
package main

import (
    "bufio"
    "fmt"
    "strings"
)

func main() {
    input := "foo  bar   baz"
    scanner := bufio.NewScanner(strings.NewReader(input))
    scanner.Split(bufio.ScanWords)
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }
}
```

输出结果：

```
foo
bar
baz
```

`Scanner` 扫描器读取数据流的时候会使用带缓冲区的 IO，并接受 `io.Reader` 作为参数。

如果你需要在内存中处理字符串或者是 bytes 切片，可以首先考虑使用 [bytes.Split](https://golang.org/pkg/bytes/#Split) 或是 [strings.Split](https://golang.org/pkg/strings/#Split) 这样的工具集，当处理这些流数据时，`bytes` 或是 `strings` 标准库中的方法可能是最简单可靠的。

在底层，扫描器使用缓冲不断存储数据，当缓冲区非空或者是读到文件的末尾时 （EOF） `split` 函数会被调用，目前我们介绍了一个预定义好的 `split` 函数，但根据下面的函数签名来看，它的用途可能更加广泛。

```go
func(data []byte, atEOF bool) (advance int, token []byte, err error)
```

目前为止，我们知道 `Split` 函数会在读数据的时候被调用，从返回值来看，它的执行应该有 3 种不同情况。

### 1. 需要补充更多的数据

这表示传入的数据还不足以生成一个字符流的标记，当返回的值分别是 `0, nil, nil` 的时候，扫描器会尝试读取更多的数据，如果缓冲区已满，那么缓冲区会在任何读取操作前自动扩容为原来的两倍，让我们来仔细看一下这个过程 [查看源码](https://play.golang.org/p/j7RDUVujNv)

```go
package main

import (
    "bufio"
    "fmt"
    "strings"
)

func main() {
    input := "abcdefghijkl"
    scanner := bufio.NewScanner(strings.NewReader(input))
    split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
        fmt.Printf("%t\t%d\t%s\n", atEOF, len(data), data)
        return 0, nil, nil
    }
    scanner.Split(split)
    buf := make([]byte, 2)
    scanner.Buffer(buf, bufio.MaxScanTokenSize)
    for scanner.Scan() {
        fmt.Printf("%s\n", scanner.Text())
    }
}
```

输出结果：

```
false	2	ab
false	4	abcd
false	8	abcdefgh
false	12	abcdefghijkl
true	12	abcdefghijkl
```

上例中的 `split` 函数可以说是简单且极其贪婪的 -- 总是请求更多的数据， `Scanner` 尝试读取更多的数据的同时会保证缓冲区拥有足够的空间来存放这些数据。在上面的例子中，我们将缓冲区的大小设置为 2。

```go
buf := make([]byte, 2)
scanner.Buffer(buf, bufio.MaxScanTokenSize)
```

在 `split` 函数第一次被调用后，`scanner` 会倍增缓冲区的容量，读取更多的数据，然后再次调用 `split` 函数。在第二次调用之后增长倍数仍然保持不变，通过观察输出结果可以发现第一次调用 `split` 得到大小为 2 的切片，然后是 4、8，最后到 12，因为没有更多的数据了。

*缓冲区的默认大小是 [4096](https://github.com/golang/go/blob/13cfb15cb18a8c0c31212c302175a4cb4c050155/src/bufio/scan.go#L76) 个字节。*

在这值得我们来讨论一下 `atEOF` 这个参数，通过这个参数我们能够在 `split` 函数中判断是否还有数据可供使用，它能够在达到数据末尾 （EOF） 或者是读取出错的时候触发为真，一旦任何上述情况发生， `scanner` 将拒绝读取任何东西，像这样的 `flag` 标志可被用来抛出异常（因其不完整的字符标记），最终会导致 `scanner.Split()` 在调用的时候返回 `false` 并终止整个进程。异常可以通过 `Err` 方法来取得。

```go
package main

import (
    "bufio"
    "errors"
    "fmt"
    "strings"
)

func main() {
    input := "abcdefghijkl"
    scanner := bufio.NewScanner(strings.NewReader(input))
    split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
        fmt.Printf("%t\t%d\t%s\n", atEOF, len(data), data)
        if atEOF {
            return 0, nil, errors.New("bad luck")
        }
        return 0, nil, nil
    }
    scanner.Split(split)
    buf := make([]byte, 12)
    scanner.Buffer(buf, bufio.MaxScanTokenSize)
    for scanner.Scan() {
        fmt.Printf("%s\n", scanner.Text())
    }
    if scanner.Err() != nil {
        fmt.Printf("error: %s\n", scanner.Err())
    }
}
```

输出结果：

```
false	12	abcdefghijkl
true	12	abcdefghijkl
error: bad luck
```

`atEOF` 参数同时也能够用于处理那些遗留在缓冲区中的数据，其中一个预定义的 `split` 函数逐行扫描输入反映了 [这种行为](https://github.com/golang/go/blob/be943df58860e7dec008ebb8d68428d54e311b94/src/bufio/scan.go#L403) ，例如我们这样输入下面这些单词时

```
foo
bar
baz
```

因为在行末并没有 `\n` 字符，因此当 [ScanLines](https://golang.org/pkg/bufio/#ScanLines) 无法找到新一行的字符时，它就会返回剩余的字符来作为最后的字符标记 ([查看源码](https://golang.org/pkg/bufio/#ScanLines))

```go
package main

import (
    "bufio"
    "fmt"
    "strings"
)

func main() {
    input := "foo\nbar\nbaz"
    scanner := bufio.NewScanner(strings.NewReader(input))
    // 事实上这里并不需要传入 ScanLines 因为这原本就是标准库默认的 split 函数
    scanner.Split(bufio.ScanLines)
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }
}
```

输出结果：

```
foo
bar
baz
```

### 2. 已找到字符标记（token）

当 `split` 函数能够检测到 _标记_ 时，就会发生这种情况。它返回在缓冲区中向前移动的字符数和 _标记_ 本身。返回两个值的原因在于 _标记_ 向前移动的距离不总是等于字节个数。假设输入为 "foo foo foo" ，当我们的目标只是找到其中的单词 ( [扫描单词](https://golang.org/pkg/bufio/#ScanWords) ) 时，`split` 函数会跳过它们之间的空格。

```
(4, "foo")
(4, "foo")
(3, "foo")
```

让我们通过一个具体的例子看一下，下面的这个函数将只会寻找连续的 `foo` 串， [查看源码](https://play.golang.org/p/X_adw-KnUM)

```go
package main

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "strings"
)

func main() {
    input := "foofoofoo"
    scanner := bufio.NewScanner(strings.NewReader(input))
    split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
        if bytes.Equal(data[:3], []byte{'f', 'o', 'o'}) {
            return 3, []byte{'F'}, nil
        }
        if atEOF {
            return 0, nil, io.EOF
        }
        return 0, nil, nil
    }
    scanner.Split(split)
    for scanner.Scan() {
        fmt.Printf("%s\n", scanner.Text())
    }
}
```

输出结果：

```
F
F
F
```

### 3. 报错

如果 `split` 函数返回了错误那么扫描器就会停止工作，[查看源码](https://play.golang.org/p/KpiyhMFUyT)

```go
package main

import (
    "bufio"
    "errors"
    "fmt"
    "strings"
)

func main() {
    input := "abcdefghijkl"
    scanner := bufio.NewScanner(strings.NewReader(input))
    split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
        return 0, nil, errors.New("bad luck")
    }
    scanner.Split(split)
    for scanner.Scan() {
        fmt.Printf("%s\n", scanner.Text())
    }
    if scanner.Err() != nil {
        fmt.Printf("error: %s\n", scanner.Err())
    }
}
```

输出结果：

```
error: bad luck
```

然而，其中有一种特殊的错误并不会使扫描器立即停止工作。

### ErrFinalToken

扫描器给信号（signal） 提供了一个叫做 [最终标记](https://golang.org/pkg/bufio/#pkg-variables) 的选项，这是一个不会打破循环（扫描过程依然返回真）的特殊标记，但随后的一系列调用会使扫描动作立刻终止。

```go
func (s *Scanner) Scan() bool {
    if s.done {
        return false
    }
    ...
```

在 Go 语言官方 [issue #11836](https://github.com/golang/go/issues/11836) 中提供了一种方法使得当发现特殊标记时也能够立即停止扫描。[查看源码](https://play.golang.org/p/ArL-k-i2OV)

```go
package main

import (
    "bufio"
    "bytes"
    "fmt"
    "strings"
)

func split(data []byte, atEOF bool) (advance int, token []byte, err error) {
    advance, token, err = bufio.ScanWords(data, atEOF)
    if err == nil && token != nil && bytes.Equal(token, []byte{'e', 'n', 'd'}) {
        return 0, []byte{'E', 'N', 'D'}, bufio.ErrFinalToken
    }
    return
}

func main() {
    input := "foo end bar"
    scanner := bufio.NewScanner(strings.NewReader(input))
    scanner.Split(split)
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }
    if scanner.Err() != nil {
        fmt.Printf("Error: %s\n", scanner.Err())
    }
}
```

输出结果：

```
foo
END
```

> `io.EOF` 和 `ErrFinalToken` 类型的错误都不被认为是真的起作用的错误 -- `Err` 方法会在任何这两个错误出现并停止扫描器时仍然返回 `nil`

### 最大标记大小 / ErrTooLong

默认情况下，缓冲区的最大长度应该小于 `64 * 1024` 个字节，这意味着找到的标记不能大于这个限制。

```go
package main

import (
    "bufio"
    "fmt"
    "strings"
)

func main() {
    input := strings.Repeat("x", bufio.MaxScanTokenSize)
    scanner := bufio.NewScanner(strings.NewReader(input))
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }
    if scanner.Err() != nil {
        fmt.Println(scanner.Err())
    }
}
```

上面的程序会打印出 `bufio.Scanner: token too long` ，我们可以通过 [Buffer](https://golang.org/pkg/bufio/#Scanner.Buffer) 方法来自定义缓冲区的长度，在上文第一小节中这个方法有出现过，但我们这次会举一个更切题的例子，[查看源码](https://play.golang.org/p/ZsgJzuIy4r)

```go
buf := make([]byte, 10)
input := strings.Repeat("x", 20)
scanner := bufio.NewScanner(strings.NewReader(input))
scanner.Buffer(buf, 20)

for scanner.Scan() {
    fmt.Println(scanner.Text())
}

if scanner.Err() != nil {
    fmt.Println(scanner.Err())
}
```

输出结果：

```
bufio.Scanner: token too long
```

### 防止死循环

几年前 [issue #8672](https://github.com/golang/go/issues/8672) 被提出，解决方案是加多一段代码，通过判断 `atEOF` 为真且缓冲区为空来确定 `split` 函数可以被调用，而现有的代码可能会进入死循环。

```go
package main

import (
    "bufio"
    "bytes"
    "fmt"
    "strings"
)

func main() {
    input := "foo|bar"
    scanner := bufio.NewScanner(strings.NewReader(input))
    split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
        if i := bytes.IndexByte(data, '|'); i >= 0 {
            return i + 1, data[0:i], nil
        }
        if atEOF {
            return len(data), data[:len(data)], nil
        }
        return 0, nil, nil
    }
    scanner.Split(split)
    for scanner.Scan() {
        if scanner.Text() != "" {
            fmt.Println(scanner.Text())
        }
    }
}
```

`split` 函数假设当 `atEOF` 为真就能够安全地使用剩余的缓冲作为标记，这引发了 [issue #8672](https://github.com/golang/go/issues/8672) 被修复之后的另一个问题： 因为缓冲区可以为空，所以当返回 `(0, [], nil)` 时 `split` 函数并不能增加缓冲区的大小， [issue #9020](https://github.com/golang/go/issues/9020) 发现了此种情况下的 `panic` ，[查看源码](https://play.golang.org/p/HUbd-ZInAQ)

```
foo
bar
panic: bufio.Scan: 100 empty tokens without progressing
```

当我第一次阅读有关 **Scanner** 或是 [SplitFunc](https://golang.org/pkg/bufio/#SplitFunc) 的文档时我并没能弄明白在所有情况下它们是如何工作的，即便是阅读源代码也帮助甚微，因为 [Scan](https://github.com/golang/go/blob/be943df58860e7dec008ebb8d68428d54e311b94/src/bufio/scan.go#L128) 看上去真的很复杂，希望这篇文章能够帮助其他人更好地理清这块的细节。

----------

via: https://medium.com/golangspec/in-depth-introduction-to-bufio-scanner-in-golang-55483bb689b4

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[yujiahaol68](https://github.com/yujiahaol68)
校对：[rxcai](https://github.com/rxcai)，[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出 