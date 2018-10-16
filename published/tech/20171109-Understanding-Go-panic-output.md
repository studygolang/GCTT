已发布：https://studygolang.com/articles/11733

# 理解 Go 语言中的 panic 输出

我的代码有一个 bug。😭

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x30 pc=0x751ba4]
goroutine 58 [running]:
github.com/joeshaw/example.UpdateResponse(0xad3c60, 0xc420257300, 0xc4201f4200, 0x16, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, ...)
        /go/src/github.com/joeshaw/example/resp.go:108 +0x144
github.com/joeshaw/example.PrefetchLoop(0xacfd60, 0xc420395480, 0x13a52453c000, 0xad3c60, 0xc420257300)
        /go/src/github.com/joeshaw/example/resp.go:82 +0xc00
created by main.runServer
        /go/src/github.com/joeshaw/example/cmd/server/server.go:100 +0x7e0
```

这个 panic 错误正如输出的第一行所指示那样，是由解引用一个 nil 指针造成的。由于 Go 在错误处理中的语法，相比于其它的语言，比如 C 或者 Java，这些类型的错误在 Go 中是不太常见的。
这种类型的错误在 Go 中比在其他语言 (如 C 或 Java) 中要少得多, 这得益于 Go 的错误处理方式。

如果一个函数执行失败，那么这个函数一定会返回一个 `error` 作为它的最后一个返回值。调用者应该立即检查该函数返回的错误。

```go
// val is a pointer, err is an error interface value
val, err := somethingThatCouldFail()
if err != nil {
    // Deal with the error, probably pushing it up the call stack
    return err
}

// By convention, nearly all the time, val is guaranteed to not be
// nil here.
```

然而，这里一定某处有一个 bug（译注：指开头的 panic），违反了这个隐式 API 的约定。

在我深入介绍之前，这里有个附加说明：这（上述代码）是与体系结构和操作系统有关的，我仅仅在 amd64 Linux 系统和 macOS 系统上运行。其它的系统运行结果应该会有所不同。

panic 错误输出的第二行给出有关触发这个 panic 的 UNIX 信号的信息：

```
[signal SIGSEGV: segmentation violation code=0x1 addr=0x30 pc=0x751ba4]
```

因为一个 nil 指针的解引用而发生了段错误（SIGSEGV）。`code` 区映射到 UNIX 的 `siginfo.si_cod` 区，并且在 Linux 的 `siginfo.h` 中 `0x1` 值是 `SEGV_MAPERR`（地址未映射到对象）。

`addr` 映射到 `siginfo.si_addr`，其值是 `0x30`，这并不是一个有效的内存地址。

`pc` 是程序计数器，我们可以使用它来找出程序崩溃的地方，但是我们通常没必要这么做，因为一个 goroutine 跟踪有如下信息。

```
goroutine 58 [running]:
github.com/joeshaw/example.UpdateResponse(0xad3c60, 0xc420257300, 0xc4201f4200, 0x16, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, ...)
        /go/src/github.com/joeshaw/example/resp.go:108 +0x144
github.com/joeshaw/example.PrefetchLoop(0xacfd60, 0xc420395480, 0x13a52453c000, 0xad3c60, 0xc420257300)
        /go/src/github.com/joeshaw/example/resp.go:82 +0xc00
created by main.runServer
        /go/src/github.com/joeshaw/example/cmd/server/server.go:100 +0x7e0
```

在这个深层次的栈帧中，第一个导致 panic 发生的（文件）会先列出。在这个例子中，是 `resp.go` 文件的 108 行。

在这个 goroutine 回溯信息里，吸引我眼球的东西是函数 `UpdateResponse` 和 `PrefetchLoop` 的参数, 因为该数字与函数签名不匹配。

```go
func UpdateResponse(c Client, id string, version int, resp *Response, data []byte) error
func PrefetchLoop(ctx context.Context, interval time.Duration, c Client)
```

`UpdateResponse` 需要 5 个参数，但是 panic 显示它携带超过 10 个参数。 `PrefetchLoop` 需要 3 个参数，但 panic 显示它带有 5 个参数。这样会发生什么呢？

为了理解参数值，我们必须要了解一些关于 Go 底层类型的数据结构。RussCox 有两篇很棒的博客，一篇关于 [基本类型，结构体和指针，字符串和切片](https://research.swtch.com/godata) ，另一篇关于 [接口](https://research.swtch.com/interfaces) ，它描述了这些在内存中是怎样分布的。对于 Go 程序员，这两篇文章是必备读物，但是概括起来是：

- 字符串有两个域 (一个指向字符串数据的指针和一个长度)
- 切片有三个域 (一个指向底层数组的指针，一个长度，一个容量)
- 接口有两个域 (一个指向类型的指针和一个指向值的指针)

当 panic 发生时，我们看到在输出中的参数值包括字符串、切片和接口的导出值。另外，函数的返回值会被添加到参数列表的末尾。

回到我们的 `UpdateResponse` 函数，`Client` 类型是一个接口，它带有 2 个值。 `id` 是一个字符串，它有 2 个值（共 4 个）。`version` 是一个整型，带有 1 个值（共 5 个值）。`resp` 是一个指针，带有 1 个值（共 6 个）。`data` 是一个切片，带有 3 个值（共 9 个值）。`error` 返回值是一个接口，所以又多 2 个值，从而总数到达 11 个。panic 输出数目限制为 10 个， 所以最后一个值在输出中被截断。

这是一个带有注释的 `UpdateResponse` 栈帧：

```
github.com/joeshaw/example.UpdateResponse(
    0xad3c60,      // c Client interface, type pointer
    0xc420257300,  // c Client interface, value pointer
    0xc4201f4200,  // id string, data pointer
    0x16,          // id string, length (0x16 = 22)
    0x1,           // version int (1)
    0x0,           // resp pointer (nil!)
    0x0,           // data slice, backing array pointer (nil)
    0x0,           // data slice, length (0)
    0x0,           // data slice, capacity (0)
    0x0,           // error interface (return value), type pointer
    ...            // truncated; would have been error interface value pointer
)
```

这有助于确认该消息来源的建议, 即 `resp` 是 `nil`，它被解引用了。

上移一个栈帧到 `PrefetchLoop`: `ctx context.Context` 是一个接口值, `interval` 是一个 `time.Duration` (它仅仅是一个 `int64`）, `Client` 也是一个接口。

`PrefetchLoop` 注释为:

```
github.com/joeshaw/example.PrefetchLoop(
    0xacfd60,       // ctx context.Context interface, type pointer
    0xc420395480,   // ctx context.Context interface, value pointer
    0x13a52453c000, // interval time.Duration (6h0m)
    0xad3c60,       // c Client interface, type pointer
    0xc420257300,   // c Client interface, value pointer
)
```

正如我之前提过的，`resp` 本不应该是 `nil`，因为这种情况只有在当返回 error 不是 `nil` 时才会发生。罪魁祸首是在代码中错误的使用了 `github.com/pkg/errors` 的 `Wrapf()` 函数而不是 `Errorf()`。

```go
// Function returns (*Response, []byte, error)

if resp.StatusCode != http.StatusOK {
    return nil, nil, errors.Wrapf(err, "got status code %d fetching response %s", resp.StatusCode, url)
}
```

如果 `Wrapf()` 的第一个参数传入为 `nil`, 则它的返回值为 `nil`。当这个 HTTP 状态码不是 `http.StatusOK`，这个函数将错误的返回 `nil，nil，nil`，因为一个非 200 的状态码不是一个错误，因此 `err` 的值为 `nil`。将 `errors.Wrapf()` 调用换成`errors.Errorf()` 可以修复这个 bug。

理解并且结合上下文语境中看 panic 输出可以更容易的追踪到错误！希望这些信息日后对你有用。

感谢 Peter Teichman, Damian Gryski, 和 Travis Bischel，他们帮助我分析 panic 的输出参数列表。

---

via: https://joeshaw.org/understanding-go-panic-output/

作者：[Joe Shaw](https://joeshaw.org/about/)
译者：[liuxinyu123](https://github.com/liuxinyu123)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
