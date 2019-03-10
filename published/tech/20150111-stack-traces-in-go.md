首发于：https://studygolang.com/articles/18792

# Go 语言的 Stack Trace

## 简介

拥有必要的 Go 语言调试技巧能让我们在定位问题的时候节省大量的时间。我认为尽可能地记录下详尽的日志信息是一个很好的实践，但是有时候单单记录下错误是不足够的。正确的理解 stack trace（堆栈轨迹）的信息能让你准确地定位到 bug 的所在，避免出现日志记录不够，需要添加更多跟踪日志，然后等待 bug 重现的窘境。

我在开始编写 Go 程序的时候就已经开始了解 stack trace 方面的知识。我们有时会写错代码，使得 Go 运行时 panic 掉我们的程序并抛出一个 stack trace。我将会在本文中告诉你 stack trace 给我们提供了什么信息，还有如何查看到每个函数被调用时传递给它的参数的值。

## 函数

我们先来写一小段会导致程序崩溃并抛出 stack trace 的代码：

**清单 1**

```go
package main

func main() {
   slice := make([]string, 2, 4)
   Example(slice, "hello", 10)
}

func Example(slice []string, str string, i int) {
   panic("Want stack trace")
}
```

清单 1 的 `main` 函数在第 5 行调用了 `Example` 函数，这个在第 8 行声明的 `Example` 函数接收 3 个参数：一个字符串切片、一个字符串和一个整型。`Example` 函数唯一的代码就是调用内置的 `panic`  函数（在第 9 行），这个函数会导致程序退出并立刻打印出一个 stack trace：

**清单 2**

```shell
Panic: Want stack trace

goroutine 1 [running]:
main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)
        /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/
        temp/main.go:9 +0x64
main.main()
        /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/
        temp/main.go:5 +0x85

goroutine 2 [runnable]:
runtime.forcegchelper()
        /Users/bill/go/src/runtime/proc.go:90
runtime.goexit()
        /Users/bill/go/src/runtime/asm_amd64.s:2232 +0x1

goroutine 3 [runnable]:
runtime.bgsweep()
        /Users/bill/go/src/runtime/mgc0.go:82
runtime.goexit()
        /Users/bill/go/src/runtime/asm_amd64.s:2232 +0x1
```

清单 2 中的 stack trace  显示了 panic 的时候所有存在的 Goroutine、所有 Goroutine 的运行状态和 Goroutine 对应的调用栈。正在运行并且导致了程序 panic 的 Goroutines 将会在顶上。我们先来关注一下导致 panic 的 Goroutine。

**清单 3**

```bash
01 Goroutine 1 [running]:
02 main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)
           /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/
           temp/main.go:9 +0x64
03 main.main()
           /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/
           temp/main.go:5 +0x85
```

清单 3 中 01 行表明了 Goroutine 1 在 panic 发生之前正在运行，在 02 行，我们可以看到 panic 的代码是 `main` 包里面的 `Example` 函数，缩进了的那行指明了源代码文件的路径和导致 panic 的代码所在的行数。在这个例子中，是第 9 行代码导致的 panic。

03 行显示了调用 `Example` 函数的函数名。即是 `main` 包中的 `main` 函数。在函数的名字后面，同样也是由缩进了的那行指明了源代码文件的路径，以及在哪一行调用的 `Example` 函数。

stack trace 显示了在 panic 发生的时候正在执行的 Goroutine 的函数的调用链，现在我们来关注一下传递给 `Example` 函数的各个参数的值：

**清单 4**

```go
// 声明
main.Example(slice []string, str string, i int)

// main 对 Example 的调用
slice := make([]string, 2, 4)
Example(slice, "hello", 10)

// Stack trace
main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)
```

根据 stack trace 的结果，我们整理出了清单 4，它展示了当 `main` 函数调用 `Example` 函数的时候，传递给 `Example` 函数的参数的值，你可能会发现，stack trace 里面显示传递的参数跟我们在源代码里面的调用不一样：在我们的源代码里面 `Example` 函数的声明是接受 3 个参数，但是 stack trace 里面显示的是传递了 6 个十六进制的参数，它们之间存在怎样的对应关系呢？要搞懂这个问题，你需要了解一下这些参数的类型的实现。

我们从 `Example` 函数的第一个参数——字符串切片开始吧。切片（slice）在 Go 里面是一个引用类型，这意味着一个切片的值只是一个 header value（标头值），它里面包含一个指针指向底层的数据。在这个例子中的切片， 它的 header value 是一个三字长的结构。包括了一个指向底层数组的指针、切片的长度以及切片的容量，stack trace 中看到的传递给 `Example` 函数的前三个参数，其实刚好就是它的 header value 对应的三个值。

**清单 5**

```go
// 切片
slice := make([]string, 2, 4)

// 切片的 header values
Pointer:  0x2080c3f50
Length:   0x2
Capacity: 0x4

// Example 函数声明
main.Example(slice []string, str string, i int)

// Stack trace
main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)
```

清单 5 显示了 stack trace 里面显示的前三个参数是怎么跟我们代码中的字符串切片的参数匹配上的。stack trace 中的第一个参数值（0x2080c3f50）对应切片底层的数组指针，第二、第三个参数值（0x2、0x4）对应切片的长度和容量，这三个参数构成了字符串切片的 header value，即我们 `Example` 函数声明里面的第一个参数 `slice []string`。

**图 1**

![Screen Shot](https://github.com/studygolang/gctt-images/raw/master/stack-trace-in-go/image02.png)

*图片由 Georgi Knox 提供*

然后我们再来看看 `Example` 函数的第二个参数——字符串。字符串也是一个引用类型，但是它的 header value 是不可修改的（immutable）。字符串的 header value 是一个大小为两个字长的数据结构，包含一个指向底层 byte 数组的指针以及一个表示字符串长度的整型。

**清单 6**

```go
// 字符串的值
"hello"

// 字符串的 header values
Pointer: 0x425c0
Length:  0x5

// Example 的函数声明
main.Example(slice []string, str string, i int)

// Stack trace
main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)
```

清单 6 展示了 stack trace 中的第四和第五个参数是怎么跟 `Example` 函数声明中的 `str string` 参数对应上的。stack trace 中的第四个参数就是字符串底层数组的地址（0x425c0），而第五个参数就是字符串的长度（0x5）。字符串（hello）需要 5 个字节。这两个参数值构成了字符串的 header value。

**图 2**

![Screen Shot](https://github.com/studygolang/gctt-images/raw/master/stack-trace-in-go/image01.png)

*图片由 Georgi Knox 提供*

 `Example` 函数的第三个参数是一个整数，它是一个单字长数值：

**清单 7**

```go
// 整型参数值
10

// 整型
十六进制数值 : 0xa

// Example 的函数声明
main.Example(slice []string, str string, i int)

// Stack trace
main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)
```

清单 7 展示了 stack trace 的最后一个参数是怎么跟 `Example` 函数声明中的 `i int` 参数对应上的。这个参数在 stack trace 中显示是十六进制数字 `0xa`，即 10 的十六进制形式。正好就是我们调用 `Example` 函数时传入的整数 10。stack trace 中的 0xa 就是传给 `Example` 函数的第三个参数的值。

**图 3**

![Screen Shot](https://github.com/studygolang/gctt-images/raw/master/stack-trace-in-go/image00.png)

*图片由 Georgi Knox 提供*

## 方法

让我们把程序改改，使 `Example` 函数变成一个方法（method）：

**清单 8**

```go
package main

import "fmt"

type trace struct{}

func main() {
    slice := make([]string, 2, 4)

    var t trace
    t.Example(slice, "hello", 10)
}

func (t *trace) Example(slice []string, str string, i int) {
    fmt.Printf("Receiver Address: %p\n", t)
    panic("Want stack trace")
}
```

清单 8 修改了原来的程序代码，在第 5 行新增了一个叫做 `trace` 的类型，并在 14 行把 `Example` 函数变成方法（我们再它函数声明上给它添加了一个 `trace` 类型的接收者。然后在第 10 行，声明了一个类型为 `trace` 类型的变量 `t`，然后在 11 行调用了这个变量 `t` 的方法。

尽管 `t` 这个变量是一个值变量而不是指针，但是由于 `Example` 方法的接收者是指针接收者，因此在调用 `Example` 方法时， Go 会把 `t` 的地址传给方法的接收者。这时当程序运行后，打印出来的 stack trace 会有所不同：

**清单 9**

```shell
Receiver Address: 0x1553a8
panic: Want stack trace

01 Goroutine 1 [running]:
02 main.(*trace).Example(0x1553a8, 0x2081b7f50, 0x2, 0x4, 0xdc1d0, 0x5, 0xa)
           /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/
           temp/main.go:16 +0x116

03 main.main()
           /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/
           temp/main.go:11 +0xae
```

首先要留意的是，在清单 9 中的 stack trace 的 02 行中明确地告诉了我们这个方法的调用使用了指针接收者：在函数名和代码包名之间多显示了一个 `(*trace)`。第二个要留意的是，现在 stack trace 里面显示的 `Example` 方法的参数列表里面，第一个参数是接收者的地址。我们在 stack trace 中就能清楚的看到这个实现的细节：方法调用其实就是函数的调用，唯一的区别就是方法调用的第一个参数其实是接收者的值。

因为 `Example` 方法的声明和调用除了上述的改动外，没有其它的变化了，所以在 stack trace 的其它值都没有什么变化。调用 `Eample` 函数的行号和 panic 发生的行号随着代码的变化发生了改变。

## Packing

当存在多个参数可以被压缩成一个字长（word）的时候，这些参数的值会被压缩（pack）在一个字长里面。

**清单 10**

```go
package main

func main() {
     Example(true, false, true, 25)
}

func Example(b1, b2, b3 bool, i uint8) {
    panic("Want stack trace")
}
```

清单 10 展示了一个新的程序的代码，它把 `Example` 函数改成一个接受 4 个参数的函数。前三个参数是布尔型而最后一个参数是 8 位的无符号整型。因为布尔型变量的大小也是 8 位的。不管你的机器是 32 位的架构还是 64 位的架构，都能用一个字长来同时保存这四个参数，所以 Go 会把这四个参数压缩到在一起，当程序运行的时候，会产生这样的 stack trace：

**清单 11**

```shell
01 Goroutine 1 [running]:
02 main.Example(0x19010001)
           /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/
           temp/main.go:8 +0x64
03 main.main()
           /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/
           temp/main.go:4 +0x32
```

可以看到 stack trace 中显示 `Example` 函数只接收 1 个参数，而不是 4 个。所有的 4 个参数都被压缩成为一个字长，并且作为整体传递给 `Example` 函数了。

**清单 12**

```go
// 参数值
true, false, true, 25

// 字长值
位数    二进制      十六进制   值
00-07   0000 0001   01    true
08-15   0000 0000   00    false
16-23   0000 0001   01    true
24-31   0001 1001   19    25

// 声明
main.Example(b1, b2, b3 bool, i uint8)

// Stack trace
main.Example(0x19010001)
```

清单 12 展示了 stack trace 里面的一个参数是怎么跟源代码中 `Example` 函数的四个参数对应上的。`true` 的值是用一个 8 位的整型 1 来保存的，而 `false` 的值其实是一个 8 位的整型 0。而 25 是二进制的 11001，也是十六进制的 19。现在我们再看看 stack trace 里面的参数（0x19010001）就能明白它的值代表的是什么了。

## 结论

Go 运行时提供了大量的调试信息来帮助我们调试程序。在本文中我们专注于 stack trace 方面的技巧。能够在调用栈中了解每个函数调用时传入的参数值，这是一个很强大的能力。它曾多次帮助我快速定位出程序的 bug。现在你已经了解了如何解读 stack trace， 希望下次屏幕跳出 stack trace 的时候你能够用上这些知识。

---

via: https://www.ardanlabs.com/blog/2015/01/stack-traces-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[Alex-liutao](https://github.com/Alex-liutao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
