首发于：https://studygolang.com/articles/16324

# 我的 Channel 在 Select 语句中的 Bug

我当时正在测试一个已经上线运行的项目的新功能，忽然代码表现得非常糟糕。我看到后很惊讶，后来搞清楚了原因。

接下来提供这份代码的简化版本，包含两个 bug。

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "time"
)

var Shutdown bool = false

func main() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)

    for {
        select {
        case <-sigChan:
            Shutdown = true
            continue

        case <-func() chan struct{} {
            complete := make(chan struct{})
            go LaunchProcessor(complete)
            return complete
        }():
            return
        }
    }
}

func LaunchProcessor(complete chan struct{}) {
    defer func() {
        close(complete)
    }()

    fmt.Printf("Start Work\n")

    for count := 0; count < 5; count++ {
        fmt.Printf("Doing Work\n")
        time.Sleep(1 * time.Second)

        if Shutdown == true {
            fmt.Printf("Kill Early\n")
            return
        }
    }

    fmt.Printf("End Work\n")
}
```

这份代码的功能是运行一项任务然后中止它，它允许操作系统申请提前终止程序。我一向喜欢尽可能的彻底关闭一个程序。

上述代码创建了一个绑定到操作系统 signal 的 channel，并且在终端窗口查找 `ctrl + c`，如果它被按下，那么  `Shutdown` 就会被设置为 `true`，并且程序跳转回 `select` 语句。

## 第一个 Bug

观察如下代码

```go
case <-func() chan struct{} {
    complete := make(chan struct{})
    go LaunchProcessor(complete)
    return complete
}():
```

我在写这段代码的时候觉得自己很聪明，我认为执行一个函数来生成 Go routine 会很棒。它会返回一个 channel，在 `select` 处等待运行完成，一旦 Go routine 运行完毕它会关闭这个 channel 然后终止程序。

让我们运行一下：

```
Start Work
Doing Work
Doing Work
Doing Work
Doing Work
Doing Work
End Work
```

正如预期的那样，程序启动并生成 Go routine。一旦 Go routine 完成，程序就终止。

接下来我会在运行时按下 `ctrl + c`。

```
Start Work
Doing Work
^CStart Work
Doing Work
Kill Early
Kill Early
```

当我按下 `ctrl + c` 的时候 Go routine 又启动了一遍！

我原以为在这个 `case` 下的函数只会被执行一次，然后一直等待 channel 继续运行，没想到每次函数运行到 `select` 都会再执行。

要修复代码，我需要把生成 Go routine 部分从 `select` 语句中移出来，在循环外生成它。

```go
func main() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)

    complete := make(chan struct{})
    go LaunchProcessor(complete)

    for {

        select {
        case <-sigChan:
            Shutdown = true
            continue

        case <-complete:
            return
        }
    }
}
```

现在我们运行程序会得到一个更好的结果。

```
Start Work
Doing Work
Doing Work
^CKill Early
```

这次当我按下 `ctrl + c` 后程序提前终止并且不再生成新的 Go routine。

## 第二个 Bug

这里还有一个不大明显的 bug 潜伏在代码中，我们来看一下：

```go
var Shutdown bool = false

if whatSig == syscall.SIGINT {
    Shutdown = true
}

if Shutdown == true {
    fmt.Printf("Kill Early\n")
    return
}
```

该代码使用包层变量通知运行的 Go routine 在 `ctrl + c` 按下时关闭。每当我按下它时，代码都在工作，那么为什么会有 bug 呢？

首先让我们运行数据竞争检测：

```
go build -race
./test
```

运行时我们再按一下 `ctrl + c`。

```
Start Work
Doing Work
^C==================
WARNING: DATA RACE
Read by Goroutine 5:// 译注 : 被 Go routine 5 读取
    main.LaunchProcessor()
        /Users/bill/Spaces/Test/src/test/main.go:46 +0x10b
    Gosched0()
        /Users/bill/go/src/pkg/runtime/proc.c:1218 +0x9f

Previous write by Goroutine 1:// 译注：被 Go routine 1 写入
    main.main()
        /Users/bill/Spaces/Test/src/test/main.go:25 +0x136
    runtime.main()
        /Users/bill/go/src/pkg/runtime/proc.c:182 +0x91

Goroutine 5 (running) created at:// 译注：生成 Go routine  5
    main.main()
        /Users/bill/Spaces/Test/src/test/main.go:18 +0x8f
    runtime.main()
        /Users/bill/go/src/pkg/runtime/proc.c:182 +0x91

Goroutine 1 (running) created at:// 译注：生成 Go routine 1
    _rt0_amd64()
        /Users/bill/go/src/pkg/runtime/asm_amd64.s:87 +0x106

==================
Kill Early
Found 1 data race(s)
```

我使用的 `Shutdown` 变量在数据竞争检测器中显现出来，这是由于有两个 Go routine 在尝试用不安全的方法访问它。

我不用安全的方法访问该变量的初衷是实用的，但是是错的。我认为由于该变量只在必要时用以关闭程序，所以我不介意脏读，但是万一脏读恰好出现在读写该变量的一瞬间呢？如果脏读出现，我可以在下次循环捕获它，看起来没损失对吧？为什么要像这样增加一个复杂的 channel，或者为代码加锁呢。

这就涉及到 Go 内存模型了。

[Go Memory Model](https://golang.org/ref/mem)

[Go 语言中文网翻译 Go 内存模型](https://studygolang.com/articles/819)

Go 内存模型不保证 Go routine 读取 Shutdown 变量时会察觉到 main routine 的写入操作，main routine 只写 Shutdown 变量一次并且该变量不会被读取回主内存，因为 main routine 永远不会去读 Shutdown 变量。

虽然这次不会出什么问题，但是随着 Go 语言编译器变得越来越复杂，有可能它会决定完全废止对 `Shutdown` 的写入。虽然这种行为被 Go 内存模型所允许，然而，我们不希望代码不能通过数据竞争检测。即使是出于实用的原因，这也不是一个好的例子。

接下来是最终版本，修复了所有 bug。

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "sync/atomic"
    "time"
)

var Shutdown int32 = 0

func main() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)

    complete := make(chan struct{})
    go LaunchProcessor(complete)

    for {

        select {
        case <-sigChan:
            atomic.StoreInt32(&Shutdown, 1)
            continue

        case <-complete:
            return
        }
    }
}

func LaunchProcessor(complete chan struct{}) {
    defer func() {
        close(complete)
    }()

    fmt.Printf("Start Work\n")

    for count := 0; count < 5; count++ {
        fmt.Printf("Doing Work\n")
        time.Sleep(1 * time.Second)

        if atomic.LoadInt32(&Shutdown) == 1 {
            fmt.Printf("Kill Early\n")
            return
        }
    }

    fmt.Printf("End Work\n")
}
```

我喜欢用 if 语句来检查 `Shutdown` 是否被设置，这样我能在需要的时候使用它。这个解决方案把 `Shutdown` 变量从 boolean 变成了 int32，并且用原子函数来存储读取。

在 main routine 如果 `ctrl + c` 被检测到，`Shutdown` 变量会安全的从 0 变成 1。在 LanuchProcessor Go routine 中，它会和 1 比较。如果条件为真 Go rontine 就会返回。

有时候确实很让人惊奇，一个如此简单的问题包含了几个陷阱，在一开始有些层面你从来没有考虑过或者意识到。尤其是那些看起来正常工作的代码。

---

via: https://www.ardanlabs.com/blog/2013/10/my-channel-select-bug.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[zuoguoyao](https://github.com/zuoguoyao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
