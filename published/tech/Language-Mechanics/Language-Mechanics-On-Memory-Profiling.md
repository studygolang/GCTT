已发布：https://studygolang.com/articles/12445

# Go 语言机制之内存剖析（Language Mechanics On Memory Profiling）

## 前序（Prelude）

本系列文章总共四篇，主要帮助大家理解 Go 语言中一些语法结构和其背后的设计原则，包括指针、栈、堆、逃逸分析和值/指针传递。这是第三篇，主要介绍堆和逃逸分析。（译者注：这一篇可看成第二篇的进阶版）

以下是本系列文章的索引：

1. [Go 语言机制之栈与指针](https://studygolang.com/articles/12443)
2. [Go 语言机制之逃逸分析](https://studygolang.com/articles/12444)
3. [Go 语言机制之内存剖析](https://studygolang.com/articles/12445)
4. [Go 语言机制之数据和语法的设计哲学](https://www.ardanlabs.com/blog/2017/06/design-philosophy-on-data-and-semantics.html)

观看这段示例代码的视频演示：[GopherCon Singapore (2017) - Escape Analysis](https://engineers.sg/video/go-concurrency-live-gophercon-sg-2017--1746)

## 介绍（Introduction）

在前面的博文中，通过一个共享在 goroutine 的栈上的值的例子讲解了逃逸分析的基础。还有其他没有介绍的造成值逃逸的场景。为了帮助大家理解，我将调试一个分配内存的程序，并使用非常有趣的方法。

## 程序（The Program）

我想了解 `io` 包，所以我创建了一个简单的项目。给定一个字符序列，写一个函数，可以找到字符串 `elvis` 并用大写开头的 `Elvis` 替换它。我们正在讨论国王（Elvis 即猫王，摇滚明星），他的名字总是大写的。

这是一个解决方案的链接：[https://play.golang.org/p/n_SzF4Cer4](https://play.golang.org/p/n_SzF4Cer4)

这是一个压力测试的链接：[https://play.golang.org/p/TnXrxJVfLV](https://play.golang.org/p/TnXrxJVfLV)

代码列表里面有两个不同的函数可以解决这个问题。这篇博文将会关注（其中的）`algOne` 函数，因为它使用到了 `io` 库。你可以自己用下 `algTwo`，体验一下内存，CPU 消耗的差异。

### 清单 1

```
Input:
abcelvisaElvisabcelviseelvisaelvisaabeeeelvise l v i saa bb e l v i saa elvi
selvielviselvielvielviselvi1elvielviselvis

Output:
abcElvisaElvisabcElviseElvisaElvisaabeeeElvise l v i saa bb e l v i saa elvi
selviElviselvielviElviselvi1elviElvisElvis
```

这是完整的 `algOne` 函数。

### 清单 2

```go
func algOne(data []byte, find []byte, repl []byte, output *bytes.Buffer) {

    // Use a bytes Buffer to provide a stream to process.
    input := bytes.NewBuffer(data)

    // The number of bytes we are looking for.
    size := len(find)

    // Declare the buffers we need to process the stream.
    buf := make([]byte, size)
    end := size - 1

    // Read in an initial number of bytes we need to get started.
    if n, err := io.ReadFull(input, buf[:end]); err != nil {
        output.Write(buf[:n])
        return
    }

    for {

        // Read in one byte from the input stream.
        if _, err := io.ReadFull(input, buf[end:]); err != nil {

            // Flush the reset of the bytes we have.
            output.Write(buf[:end])
            return
        }

        // If we have a match, replace the bytes.
        if bytes.Compare(buf, find) == 0 {
            output.Write(repl)

            // Read a new initial number of bytes.
            if n, err := io.ReadFull(input, buf[:end]); err != nil {
                output.Write(buf[:n])
                return
            }

            continue
        }

        // Write the front byte since it has been compared.
        output.WriteByte(buf[0])

        // Slice that front byte out.
        copy(buf, buf[1:])
    }
}
```

我想知道的是这个函数的性能表现得怎么样，以及它在堆上分配带来什么样的压力。为了这个目的，我们将进行压力测试。

## 压力测试（Benchmarking）

这个是我写的压力测试函数，它在内部调用 `algOne` 函数去处理数据流。

### 清单 3

```go
func BenchmarkAlgorithmOne(b *testing.B) {
    var output bytes.Buffer
    in := assembleInputStream()
    find := []byte("elvis")
    repl := []byte("Elvis")

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        output.Reset()
        algOne(in, find, repl, &output)
    }
}
```

有这个压力测试函数，我们就可以运行 `go test` 并使用 `-bench`，`-benchtime` 和 `-benchmem` 选项。

### 清单 4

```
$ go test -run none -bench AlgorithmOne -benchtime 3s -benchmem
BenchmarkAlgorithmOne-8    	2000000 	     2522 ns/op       117 B/op  	      2 allocs/op
```

运行完压力测试后，我们可以看到 `algOne` 函数分配了两次值，每次分配了 117 个字节。这真的很棒，但我们还需要知道哪行代码造成了分配。为了这个目的，我们需要生成压力测试的分析数据。

## 性能分析（Profiling）

为了生成分析数据，我们将再次运行压力测试，但这次为了生成内存检测数据，我们打开 `-memprofile` 开关。

### 清单 5

```
$ go test -run none -bench AlgorithmOne -benchtime 3s -benchmem -memprofile mem.out
BenchmarkAlgorithmOne-8    	2000000      2570 ns/op       117 B/op        2 allocs/op
```

一旦压力测试完成，测试工具就会生成两个新的文件。

### 清单 6

```
~/code/go/src/.../memcpu
$ ls -l
total 9248
-rw-r--r--  1 bill  staff      209 May 22 18:11 mem.out       (NEW)
-rwxr-xr-x  1 bill  staff  2847600 May 22 18:10 memcpu.test   (NEW)
-rw-r--r--  1 bill  staff     4761 May 22 18:01 stream.go
-rw-r--r--  1 bill  staff      880 May 22 14:49 stream_test.go
```

源码在 `memcpu` 目录中，`algOne` 函数在 `stream.go` 文件中，压力测试函数在 `stream_test.go` 文件中。新生成的文件为 `mem.out` 和 `memcpu.test`。`mem.out` 包含分析数据和 `memcpu.test` 文件，以及包含我们查看分析数据时需要访问符号的二进制文件。

有了分析数据和二进制测试文件，我们就可以运行 `pprof` 工具学习数据分析。

### 清单 7

```
$ go tool pprof -alloc_space memcpu.test mem.out
Entering interactive mode (type "help" for commands)
(pprof) _
```

当分析内存数据时，为了轻而易举地得到我们要的信息，你会想用 `-alloc_space` 选项替代默认的 `-inuse_space` 选项。这将会向你展示每一次分配发生在哪里，不管你分析数据时它是不是还在内存中。

在 `（pprof）` 提示下，我们使用 `list` 命令检查 `algOne` 函数。这个命令可以使用正则表达式作为参数找到你要的函数。

### 清单 8

```
(pprof) list algOne
Total: 335.03MB
ROUTINE ======================== .../memcpu.algOne in code/go/src/.../memcpu/stream.go
 335.03MB   335.03MB (flat, cum)   100% of Total
        .          .     78:
        .          .     79:// algOne is one way to solve the problem.
        .          .     80:func algOne(data []byte, find []byte, repl []byte, output *bytes.Buffer) {
        .          .     81:
        .          .     82: // Use a bytes Buffer to provide a stream to process.
 318.53MB   318.53MB     83: input := bytes.NewBuffer(data)
        .          .     84:
        .          .     85: // The number of bytes we are looking for.
        .          .     86: size := len(find)
        .          .     87:
        .          .     88: // Declare the buffers we need to process the stream.
  16.50MB    16.50MB     89: buf := make([]byte, size)
        .          .     90: end := size - 1
        .          .     91:
        .          .     92: // Read in an initial number of bytes we need to get started.
        .          .     93: if n, err := io.ReadFull(input, buf[:end]); err != nil || n < end {
        .          .     94:       output.Write(buf[:n])
(pprof) _
```

基于这次的数据分析，我们现在知道了 `input`，`buf` 数组在堆中分配。因为 `input` 是指针变量，分析数据表明 `input` 指针变量指定的 `bytes.Buffer` 值分配了。我们先关注 `input` 内存分配以及弄清楚为啥会被分配。

我们可以假定它被分配是因为调用 `bytes.NewBuffer` 函数时在栈上共享了 `bytes.Buffer` 值。然而，存在于 `flat` 列（pprof 输出的第一列）的值告诉我们值被分配是因为 `algOne` 函数共享造成了它的逃逸。

我知道 `flat` 列代表在函数中的分配是因为 `list` 命令显示 `Benchmark` 函数中调用了 `aglOne`。

### 清单 9

```
(pprof) list Benchmark
Total: 335.03MB
ROUTINE ======================== .../memcpu.BenchmarkAlgorithmOne in code/go/src/.../memcpu/stream_test.go
        0   335.03MB (flat, cum)   100% of Total
        .          .     18: find := []byte("elvis")
        .          .     19: repl := []byte("Elvis")
        .          .     20:
        .          .     21: b.ResetTimer()
        .          .     22:
        .   335.03MB     23: for i := 0; i < b.N; i++ {
        .          .     24:       output.Reset()
        .          .     25:       algOne(in, find, repl, &output)
        .          .     26: }
        .          .     27:}
        .          .     28:
(pprof) _
```

因为在 `cum` 列（第二列）只有一个值，这告诉我 `Benchmark` 没有直接分配。所有的内存分配都发生在函数调用的循环里。你可以看到这两个 `list` 调用的分配次数是匹配的。

我们还是不知道为什么 `bytes.Buffer` 值被分配。这时在 `go build` 的时候打开 `-gcflags "-m -m"` 就派上用场了。分析数据只能告诉你哪些值逃逸，但编译命令可以告诉你为啥。

## 编译器报告（Compiler Reporting）

让我们看一下编译器关于代码中逃逸分析的判决。

### 清单 10

```bash
go build -gcflags "-m -m"
```

这个命令产生了一大堆的输出。我们只需要搜索输出中包含 `stream.go:83`，因为 `stream.go` 是包含这段代码的文件名并且第 83 行包含 `bytes.Buffer` 的值。搜索后我们找到 6 行。

### 清单 11

```
./stream.go:83: inlining call to bytes.NewBuffer func([]byte) *bytes.Buffer { return &bytes.Buffer literal }

./stream.go:83: &bytes.Buffer literal escapes to heap
./stream.go:83:   from ~r0 (assign-pair) at ./stream.go:83
./stream.go:83:   from input (assigned) at ./stream.go:83
./stream.go:83:   from input (interface-converted) at ./stream.go:93
./stream.go:83:   from input (passed to call[argument escapes]) at ./stream.go:93
```

我们搜索 `stream.go:83` 找到的第一行很有趣。

### 清单 12

```
./stream.go:83: inlining call to bytes.NewBuffer func([]byte) *bytes.Buffer { return &bytes.Buffer literal }
```

可以肯定 `bytes.Buffer` 值没有逃逸，因为它传递给了调用栈。这是因为没有调用 `bytes.NewBuffer`，函数内联处理了。

所以这是我写的代码片段：

## 清单 13
 ```
83     input := bytes.NewBuffer(data)
```

因为编译器选择内联 `bytes.NewBuffer` 函数调用，我写的代码被转成：

### 清单 14
```
input := &bytes.Buffer{buf: data}
```

这意味着 `algOne` 函数直接构造 `bytes.Buffer` 值。那么，现在的问题是什么造成了值从 `algOne` 栈帧中逃逸？答案在我们搜索结果中的另外 5 行。

### 清单 15

```
./stream.go:83: &bytes.Buffer literal escapes to heap
./stream.go:83:   from ~r0 (assign-pair) at ./stream.go:83
./stream.go:83:   from input (assigned) at ./stream.go:83
./stream.go:83:   from input (interface-converted) at ./stream.go:93
./stream.go:83:   from input (passed to call[argument escapes]) at ./stream.go:93
```

这几行告诉我们代码中的第 93 行造成了逃逸。`input` 变量被赋值给一个接口变量。

## 接口（Interfaces）

我完全不记得在代码中将值赋给了接口变量。然而，如果你看到 93 行，就可以非常清楚地看到发生了什么。

### 清单 16
```
 93     if n, err := io.ReadFull(input, buf[:end]); err != nil {
 94         output.Write(buf[:n])
 95         return
 96     }
```

`io.ReadFull` 调用造成了接口赋值。如果你看了 `io.ReadFull` 函数的定义，你可以看到一个接口类型是如何接收 `input` 值。

### 清单 17

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

func ReadFull(r Reader, buf []byte) (n int, err error) {
    return ReadAtLeast(r, buf, len(buf))
}
```

传递 `bytes.Buffer` 地址到调用栈，在 `Reader` 接口变量中存储会造成一次逃逸。现在我们知道使用接口变量是需要开销的：分配和重定向。所以，如果没有很明显的使用接口的原因，你可能不想使用接口。下面是我选择在我的代码中是否使用接口的原则。

使用接口的情况：

- 用户 API 需要提供实现细节的时候。
- API 的内部需要维护多种实现。
- 可以改变的 API 部分已经被识别并需要解耦。

不使用接口的情况：

- 为了使用接口而使用接口。
- 推广算法。
- 当用户可以定义自己的接口时。

现在我们可以问自己，这个算法真的需要 `io.ReadFull` 函数吗？答案是否定的，因为 `bytes.Buffer` 类型有一个方法可以供我们使用。使用方法而不是调用一个函数可以防止重新分配内存。

让我们修改代码，删除 `io` 包，并直接使用 `Read` 函数而不是 `input` 变量。

修改后的代码删除了 `io` 包的调用，为了保留相同的行号，我使用空标志符替代 `io` 包的引用。这会允许（没有使用的）库导入的行待在列表中。

### 清单 18

```go
import (
    "bytes"
    "fmt"
    _ "io"
)

func algOne(data []byte, find []byte, repl []byte, output *bytes.Buffer) {

    // Use a bytes Buffer to provide a stream to process.
    input := bytes.NewBuffer(data)

    // The number of bytes we are looking for.
    size := len(find)

    // Declare the buffers we need to process the stream.
    buf := make([]byte, size)
    end := size - 1

    // Read in an initial number of bytes we need to get started.
    if n, err := input.Read(buf[:end]); err != nil || n < end {
        output.Write(buf[:n])
        return
    }

    for {

        // Read in one byte from the input stream.
        if _, err := input.Read(buf[end:]); err != nil {

            // Flush the reset of the bytes we have.
            output.Write(buf[:end])
            return
        }

        // If we have a match, replace the bytes.
        if bytes.Compare(buf, find) == 0 {
            output.Write(repl)

            // Read a new initial number of bytes.
            if n, err := input.Read(buf[:end]); err != nil || n < end {
                output.Write(buf[:n])
                return
            }

            continue
        }

        // Write the front byte since it has been compared.
        output.WriteByte(buf[0])

        // Slice that front byte out.
        copy(buf, buf[1:])
    }
}
```

修改后我们执行压力测试，可以看到 `bytes.Buffer` 的分配消失了。

### 清单 19

```
$ go test -run none -bench AlgorithmOne -benchtime 3s -benchmem -memprofile mem.out
BenchmarkAlgorithmOne-8    	2000000      1814 ns/op         5 B/op        1 allocs/op
```

我们可以看到大约 29% 的性能提升。代码从 `2570 ns/op` 降到 `1814 ns/op`。解决了这个问题，我们现在可以关注 `buf` 切片数组。如果再次使用测试代码生成分析数据，我们应该能够识别到造成剩下的分配的原因。

### 清单 20

```
$ go tool pprof -alloc_space memcpu.test mem.out
Entering interactive mode (type "help" for commands)
(pprof) list algOne
Total: 7.50MB
ROUTINE ======================== .../memcpu.BenchmarkAlgorithmOne in code/go/src/.../memcpu/stream_test.go
     11MB       11MB (flat, cum)   100% of Total
        .          .     84:
        .          .     85: // The number of bytes we are looking for.
        .          .     86: size := len(find)
        .          .     87:
        .          .     88: // Declare the buffers we need to process the stream.
     11MB       11MB     89: buf := make([]byte, size)
        .          .     90: end := size - 1
        .          .     91:
        .          .     92: // Read in an initial number of bytes we need to get started.
        .          .     93: if n, err := input.Read(buf[:end]); err != nil || n < end {
        .          .     94:       output.Write(buf[:n])
```

只剩下 89 行所示，对数组切片的分配。

## 栈帧

想知道造成 `buf` 数组切片的分配的原因？让我们再次运行 `go build`，并使用 `-gcflags "-m -m"` 选项并搜索 `stream.go:89`。

### 清单 21

```
$ go build -gcflags "-m -m"
./stream.go:89: make([]byte, size) escapes to heap
./stream.go:89: from make([]byte, size) (too large for stack) at ./stream.go:89
```

报告显示，对于栈来说，数组太大了。这个信息误导了我们。并不是说底层的数组太大，而是编译器在编译时并不知道数组的大小。

值只有在编译器编译时知道其大小才会将它分配到栈中。这是因为每个函数的栈帧大小是在编译时计算的。如果编译器不知道其大小，就只会在堆中分配。

为了验证（我们的想法），我们将值硬编码为 5，然后再次运行压力测试。

### 清单 22

```
89     buf := make([]byte, 5)
```

这一次我们运行压力测试，分配消失了。

### 清单 23

```
$ go test -run none -bench AlgorithmOne -benchtime 3s -benchmem
BenchmarkAlgorithmOne-8    3000000      1720 ns/op        0 B/op        0 allocs/op
```

如果你再看一下编译器报告，你会发现没有需要逃逸处理的。

### 清单 24

```
$ go build -gcflags "-m -m"
./stream.go:83: algOne &bytes.Buffer literal does not escape
./stream.go:89: algOne make([]byte, 5) does not escape
```

很明显我们无法确定切片的大小，所以我们在算法中需要一次分配。

## 分配和性能（Allocation and Performance）

比较一下我们在重构过程中，每次提升的性能。

### 清单 25

```
Before any optimization
BenchmarkAlgorithmOne-8    	2000000      2570 ns/op       117 B/op        2 allocs/op

Removing the bytes.Buffer allocation
BenchmarkAlgorithmOne-8     2000000      1814 ns/op         5 B/op        1 allocs/op

Removing the backing array allocation
BenchmarkAlgorithmOne-8     3000000      1720 ns/op         0 B/op        0 allocs/op
```

删除掉 bytes.Buffer 里面的（重新）内存分配，我们获得了大约 29% 的性能提升，删除掉所有的分配，我们能获得大约 33% 的性能提升。内存分配是应用程序性能影响因素之一。

## 结论（Conclusion）

Go 拥有一些神奇的工具使你能了解编译器作出的跟逃逸分析相关的一些决定。基于这些信息，你可以通过重构代码使得值存在于栈中而不需要在（被重新分配到）堆中。你不是想去掉所有软件中所有的内存（再）分配，而是想最小化这些分配。

这就是说，写程序时永远不要把性能作为第一优先级，因为你并不想（在写程序时）一直猜测性能。写正确的代码才是你第一优先级。这意味着，我们首先要关注的是完整性、可读性和简单性。一旦有了可以运行的程序，才需要确定程序是否足够快。假如程序不够快，那么使用语言提供的工具来查找和解决性能问题。

---

via: https://www.ardanlabs.com/blog/2017/06/language-mechanics-on-memory-profiling.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[gogeof](https://github.com/gogeof)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
