首发于：https://studygolang.com/articles/15349

# 在 Go 中发现竞态条件 (Race Conditions)

当我意识到我一直在处理和解决的问题有一个专有名词描述的时候，我总会觉得这事十分有趣。这次出现这种情况的是竞争条件(Race Conditions)。当你处理多个 routine 共享某类资源的时候，不可避免的需要考虑到这个。如果你未曾在你的代码中考虑过这个问题，现在就是一个不错的时候。

竞争条件是：当两个或两个以上的 routines 访问相同资源的时候，例如一个变量或者结构体，并且在不考虑其他 Gorutines 的情况下对资源进行读/写。这类代码能够造成你所能见到的最令人头疼，最随机的 bugs。通常需要大量的日志和一些运气才能找到这些类型的错误。这么多年来，我完善了我的日志的技能，以便识别这些问题。

Go 工具套件在 Go 版本 1.1 引入了一个竞态检测工具(race detector)。这个竞态检测工具是在编译流程中内置到你程序的代码。一旦你的程序开始运行，它能够发现和报告任何他所检测到的竞态情况。 这是非常棒的功能, 为识别缺陷代码做出了非常重要的工作。

让我们写一个非常的简单的包含竞态条件内置竞态检测代码的程序。

```go
package main

import (
    "fmt"
    "sync"
)

var Wait sync.WaitGroup
var Counter int = 0

func main() {

    for routine := 1; routine <= 2; routine++ {

        Wait.Add(1)
        go Routine(routine)
    }

    Wait.Wait()
    fmt.Printf("Final Counter: %d\n", Counter)
}

func Routine(id int) {

    for count := 0; count < 2; count++ {

        value := Counter
        value++
        Counter = value
    }

    Wait.Done()
}
```

这个程序看起来没有问题。它创建了两个协程，每一个协程都会增加全局变量 Counter 两次。当他们都运行结束后，程序显示全局变量 Counter 的值。当我运行这个程序的时候，他会显示正确答案 4。所以这个程序工作正常,但真的吗？

让我们通过 Go 竞态检测运行这个代码，看看它会发现什么？ 在代码所在的目录打开终端，以 -race 参数编译代码。

```
go build -race
```

然后程序输出

```
==================
WARNING: DATA RACE
Read by goroutine 5:
  main.Routine()
      /Users/bill/Spaces/Test/src/test/main.go:29 +0x44
  gosched0()
      /usr/local/go/src/pkg/runtime/proc.c:1218 +0x9f

Previous write by goroutine 4:
  main.Routine()
      /Users/bill/Spaces/Test/src/test/main.go:33 +0x65
  gosched0()
      /usr/local/go/src/pkg/runtime/proc.c:1218 +0x9f

Goroutine 5 (running) created at:
  main.main()
      /Users/bill/Spaces/Test/src/test/main.go:17 +0x66
  runtime.main()
      /usr/local/go/src/pkg/runtime/proc.c:182 +0x91

Goroutine 4 (finished) created at:
  main.main()
      /Users/bill/Spaces/Test/src/test/main.go:17 +0x66
  runtime.main()
      /usr/local/go/src/pkg/runtime/proc.c:182 +0x91

==================
Final Counter: 4
Found 1 data race(s)
```

看起来，工具在代码中检测到竞争条件。如果你查看上面的竞争条件报告，你会看到针对程序的输出。全局变量 Counter 的值是 4。这就是这类的 bug 的难点所在，代码大部分情况是工作正常的，但错误的情况会随机产生。竞争检测告诉我们隐藏在代码中的糟糕问题。

警告报告告诉我们问题发生的准确位置:

```
Read by goroutine 5:
  main.Routine()
      /Users/bill/Spaces/Test/src/test/main.go:29 +0x44
  gosched0()
      /usr/local/go/src/pkg/runtime/proc.c:1218 +0x9f

        value := Counter

Previous write by goroutine 4:
  main.Routine()
      /Users/bill/Spaces/Test/src/test/main.go:33 +0x65
  gosched0()
      /usr/local/go/src/pkg/runtime/proc.c:1218 +0x9f

        Counter = value

Goroutine 5 (running) created at:
  main.main()
      /Users/bill/Spaces/Test/src/test/main.go:17 +0x66
  runtime.main()
      /usr/local/go/src/pkg/runtime/proc.c:182 +0x91

        go Routine(routine)
```

你能发现竞争检测器指出两行读和写全局变量 Counter 的代码。同时也指出生成协程的代码。

让我们对代码进行简单修改，让竞争情况更容易暴露出来。

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

var Wait sync.WaitGroup
var Counter int = 0

func main() {

    for routine := 1; routine <= 2; routine++ {

        Wait.Add(1)
        go Routine(routine)
    }

    Wait.Wait()
    fmt.Printf("Final Counter: %d\n", Counter)
}

func Routine(id int) {

    for count := 0; count < 2; count++ {

        value := Counter
        time.Sleep(1 * time.Nanosecond)
        value++
        Counter = value
    }

    Wait.Done()
}
```

我在循环中增加了一个纳秒的暂停。这个暂停正好位于协程读取全局变量 Couter 存储到本地副本之后。让我们运行这个程序看看在这种修改之后，全局变量 Counter 的值是什么？

```
Final Counter: 2
```

循环中的暂停导致程序的失败。Counter 变量的值不再是 4 而是 2。发生了什么？ 让我们深挖代码看看为什么这个纳秒的暂停会导致这个 Bug。

在没有暂停的情况下，代码运行如下图：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/detecting-race-conditions-with-go/1.png)

没有暂停的情况下，第一个协程被生成，并且完成执行，紧接着第二个协程才开始运行。这就是为什么程序看起来像正确运行的原因，因为它在我的电脑上运行速度非常快，以至于代码自行排队运行。

让我们看看在有暂停的情况下，代码如何运行:

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/detecting-race-conditions-with-go/2.png)

上图已经展示了所有必要的信息，因此我就没有把他全部画出来。这个暂停导致运行的两个协程之间进行了一次上下文切换。这次我们有一个完全不同的情况。让我们看看图中展示的代码:

```go
value := Counter

time.Sleep(1 * time.Nanosecond)

value++

Counter = value
```

在每一次循环的迭代过程中，全局变量 Counter 的值都被暂存到本地变量 value，本地的副本自增后，最终写回全局变量 Counter。如果这三行代码在没有中断的情况下，没有立即运行，那么程序就会出现问题。上面的图片展示了全局变量 Counter 的读取和上下文切换是如何导致问题的。

在这幅图中，在被协程 1 增加的变量被写回全局变量 Counter 之前，协程 2 被唤醒并读取全局变量 Counter。实质上，这两个协程对全局Counter变量执行完全相同的读写操作，因此最终的结果才是 2。

为了解决这个问题，你也许认为我们只需要将增加全局变量 Counter 的三行代码改写减少到一行即可。

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

var Wait sync.WaitGroup
var Counter int = 0

func main() {

    for routine := 1; routine <= 2; routine++ {

        Wait.Add(1)
        go Routine(routine)
    }

    Wait.Wait()
    fmt.Printf("Final Counter: %d\n", Counter)
}

func Routine(id int) {

    for count := 0; count < 2; count++ {

        Counter = Counter + 1
        time.Sleep(1 * time.Nanosecond)
    }

    Wait.Done()
}

```

当我们运行这个版本的代码的时候，我们会再次得到正确的结果:

```
Final Counter: 4
```

如果我们启动竞争检测来运行该代码，上面出现的问题应该会消失:

```
go build -race
```

并且输出为:

```
==================
WARNING: DATA RACE
Write by goroutine 5:
  main.Routine()
      /Users/bill/Spaces/Test/src/test/main.go:30 +0x44
  gosched0()
      /usr/local/go/src/pkg/runtime/proc.c:1218 +0x9f

Previous write by goroutine 4:
  main.Routine()
      /Users/bill/Spaces/Test/src/test/main.go:30 +0x44
  gosched0()
      /usr/local/go/src/pkg/runtime/proc.c:1218 +0x9f

Goroutine 5 (running) created at:
  main.main()
      /Users/bill/Spaces/Test/src/test/main.go:18 +0x66
  runtime.main()
      /usr/local/go/src/pkg/runtime/proc.c:182 +0x91

Goroutine 4 (running) created at:
  main.main()
      /Users/bill/Spaces/Test/src/test/main.go:18 +0x66
  runtime.main()
      /usr/local/go/src/pkg/runtime/proc.c:182 +0x91

==================
Final Counter: 4
Found 1 data race(s)

```

然而，在这三十行代码的程序中，我们仍然检测到一个竞争条件。

```
Write by goroutine 5:
  main.Routine()
      /Users/bill/Spaces/Test/src/test/main.go:30 +0x44
  gosched0()
      /usr/local/go/src/pkg/runtime/proc.c:1218 +0x9f

        Counter = Counter + 1

Previous write by goroutine 4:
  main.Routine()
      /Users/bill/Spaces/Test/src/test/main.go:30 +0x44
  gosched0()
      /usr/local/go/src/pkg/runtime/proc.c:1218 +0x9f

        Counter = Counter + 1

Goroutine 5 (running) created at:
  main.main()
      /Users/bill/Spaces/Test/src/test/main.go:18 +0x66
  runtime.main()
      /usr/local/go/src/pkg/runtime/proc.c:182 +0x91

        go Routine(routine)
```

使用一行代码进行增加操作的程序正确地运行了。但为什么代码仍然有一个竞态条件？ 不要被我们用于递增 Counter 变量的一行Go代码所欺骗。让我们看看这一行代码生成的汇编代码:

```
0064 (./main.go:30) MOVQ Counter+0(SB),BX ; Copy the value of Counter to BX
0065 (./main.go:30) INCQ ,BX              ; Increment the value of BX
0066 (./main.go:30) MOVQ BX,Counter+0(SB) ; Move the new value to Counter
```

实际上是执行这三行汇编代码增加 counter 变量。他们十分诡异地看起来像最初的 Go 代码。上下文切换可能发生在这三行汇编的中的任意一行后面。尽管这个程序正常工作了，但严格来说，Bug 仍然存在。

尽管我使用的例子非常简单，它还是体现发现这种 Bug 的复杂性。任何一行由 Go 编译器产生的汇编代码都有可能因为下文切换而停止运行。我们的 Go 代码也许看起来能够安全地访问资源，实际上底层汇编代码可能漏洞百出。

为了解决这类问题，我们需要确保读写全局变量 Counter 总是在任何其他协程访问该变量之前完成。管道(channle)能够帮助我们有序地访问资源。这一次，我会使用一个互斥锁(Mutex):

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

var Wait sync.WaitGroup
var Counter int = 0
var Lock sync.Mutex

func main() {

    for routine := 1; routine <= 2; routine++ {

        Wait.Add(1)
        go Routine(routine)
    }

    Wait.Wait()
    fmt.Printf("Final Counter: %d\n", Counter)
}

func Routine(id int) {

    for count := 0; count < 2; count++ {

        Lock.Lock()

        value := Counter
        time.Sleep(1 * time.Nanosecond)
        value++
        Counter = value

        Lock.Unlock()
    }

    Wait.Done()
}
```

以竞态检测的模式，编译程序，查看运行结果:

```
go build -race
./test

Final Counter: 4
```

这一次，我们得到了正确的结果，并且没有发现任何竞态条件。这个程序是没有问题的。互斥锁保护了在 Lock 和 Unlock 之间的代码，确保了一次只有一个协程执行该段代码。

你可以通过以下文章学习更多例子，更好地理解 Go 竞态检测器：

http://blog.golang.org/race-detector

如果你使用了多个协程，那么使用竞态检测器测试你的代码是个不错的建议。它会在单元测试和质量保证测试中，为你节省大量的时间和麻烦。Go 开发人员能有这样的工具是很幸运地，所以值得学习一下。

---

via: https://www.ardanlabs.com/blog/2013/09/detecting-race-conditions-with-go.html

作者：[William Kennedy](https://twitter.com/goinggodotnet)
译者：[magichan](https://github.com/magichan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
