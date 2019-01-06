首发于：https://studygolang.com/articles/17486

# Concurrency, Goroutines and GOMAXPROC

William Kennedy 2014 年 1 月 29 日

## 介绍

刚刚加入[GO-Minami](http://www.meetup.com/Go-Miami/) 组织的新人经常会说想学习更多有关 Go 并发的知识。并发好像在每个语言中都是热门话题，当然我第一次听说 Go 语言时也是因为这个点。而 Rob Pike 的一段 [GO Concurrency Patterns](http://www.youtube.com/watch?v=f6kdp27TYZs) 视频才让我真真意识到我需要学习这门语言。

为了了解为什么 Go 语言写并发代码更容易更健壮，我们首先需要理解并发程序是什么，和并发程序会导致什么样的结果。在文章中我不就不讨论 CSP (通信顺序过程) 了，这个是 Go 语言 channel 实现的基础。这篇文章将关注点放在什么是并发编程，goroutines 在其中扮演什么角色、GOMAXPROCS 环境变量和 runtime 函数如何影响文章中写的 Go 程序。

## 进程和线程

当我们打开一个应用时，比如现在打开的用于写文章的浏览器，操作系统就会为这个应用创建一个进程。这个进程扮演的角色是作为这个应用的一个容器，这个容器可以包含应用运行所需要的资源。这些资源包括内存地址空间，文件引用，设备和线程。

线程相对于进程而言，线程是由操作系统调度的一个执行过程的路线，而这个执行过程就是我们对我们方法中代码的执行过程。一个进程开始于一个线程，这个线程是主线程，并且当主线程结束时这个进程也就结束了。那是因为这个主线程是应用的启动的源点。另一分方面，主线程可以启动更多线程，这些被主线程启动的线程又可以启动更多的线程。

操作系统调度器去决定哪个可用进程中的线程去执行，而不管这个线程到底属于哪个进程。每个操作系统都有它们自己的算法来决定如何选择执行线程。所以对于我们写并发程序而言，最好不要针对某一个算法而个性化开发。除此之外，每个操作系统升级新版本时他们的算法就会相应地发生变化，所以写并发程序就像是玩一个危险的游戏。

## 协程和并行

在 Go 中任何方法或者函数都可以作为一个协程来调用 , 我们可以认为主函数就是一个通过协程运行的，然而 Go 运行时并没有启动其他协程。协程可以被认为是轻量级的，因为它使用很少的内存和资源，除此之外，协程初始化时需要的栈空间是很小的。在 Go1.2 版本前初始栈空间需要 4K，现在从 1.4 版本后是 8K。栈空间的大小会根据协程的需要自动进行扩大。

操作系统是根据当前机器的可用处理器个数来调度线程，Go 运行时是以一个操作系统级别的线程组成的逻辑处理器来执行协程调度的。默认情况下，Go 运行时会分配一个单核的逻辑处理器去处理所有在程序中创建的协程。即使是一个单核的逻辑处理器和操作系统线程，也可以以惊人的效率和性能来调度成千上万个协程并发运行。我是不建议添加逻辑处理器的，但是如果你想并行运行协程，你可以通过设置 GOMAXPROCS 环境变量或者 runtime 方法来完成。

并发不是平行。并行是指当两个或两个以上的线程在不同的处理器同时执行的现象。如果你通过定义 runtime 去使用 1 个以上的逻辑处理器，调度器将会分配这些协程在不同的逻辑处理器上，这就会导致协程运行在不同的操作系统级别的线程上。然而，为了并行运行程序你需要一个多核处理器的机器。如果不是这样，即使你的 runtime 设置的是多核逻辑处理器，程序还是运行在一个单核的处理器上。

## 并发的例子

让我们来创建一个小的程序来展示 Go 运行协程时的并发性。在这个例子中我们是在一个逻辑处理器上运行的：

```go
package main

import (
    "fmt"
    "runtime"
    "sync"
)

func main() {
    runtime.GOMAXPROCS(1)

    var wg sync.WaitGroup
    wg.Add(2)

    fmt.Println("Starting Go Routines")
    Go func() {
        defer wg.Done()

        for char := 'a' ; char < 'a' +26; char++ {
            fmt.Printf("%c ", char)
        }
    }()

    Go func() {
        defer wg.Done()

        for number := 1; number < 27; number++ {
            fmt.Printf("%d ", number)
        }
    }()

    fmt.Println("Waiting To Finish")
    wg.Wait()

    fmt.Println("\nTerminating Program")
}
```

这个程序通过 Go 关键字和两个匿名函数创建了两个协程。第一个协程展示的是小写字母表，第二个协程展示的是 1 到 26 个数字，当我们运行这个程序时得到下面的输出：

```
Starting Go Routines
Waiting To Finish
a b c d e f g h i j k l m n o p q r s t u v w x y z 1 2 3 4 5 6 7 8 9 10 11
12 13 14 15 16 17 18 19 20 21 22 23 24 25 26
Terminating Program
```

通过看结果我们发现代码是并发运行的。一旦两个协程被启动后，这个主的协程需要等待两个协程执行完成，因为如果不等带它们执行完成就结束主协程的话，这个程序就结束了。使用 WaitGrout 是一个处理协程之间交流是否结束的好方法。

我们可以发现在全部展示完 a-z 26 个字母后才展示 1-26 个数字。这个是因为完成这些工作只需不到 1ms 的时间，我们并没有在第一个协程结束前看到调度动作。我们可以在协程中使用 sleep 来使协程发生调度的现象：

```go
package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
)

func main() {
    runtime.GOMAXPROCS(1)

    var wg sync.WaitGroup
    wg.Add(2)

    fmt.Println("Starting Go Routines")
    Go func() {
        defer wg.Done()

        time.Sleep(1 * time.Microsecond)
        for char := ‘ a ’ ; char < ‘ a ’ +26; char++ {
            fmt.Printf("%c ", char)
        }
    }()

    Go func() {
        defer wg.Done()

        for number := 1; number < 27; number++ {
            fmt.Printf("%d ", number)
        }
    }()

    fmt.Println("Waiting To Finish")
    wg.Wait()

    fmt.Println("\nTerminating Program")
}
```

这次我们在第一个协程刚启动时添加了 sleep 函数，通过调用 sleep 函数调度器交换了协程的执行顺序。

```
Starting Go Routines
Waiting To Finish
1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 a
b c d e f g h i j k l m n o p q r s t u v w x y z
Terminating Program
```

这次数字的展示在字母表的前面了。这个 sleep 导致调度器停止了当前执行的第一个协程并开始执行第二个协程。

## 并行的例子

在刚刚我们运行的两个例子中，协程是并发运行的，而不是并行。让我们改变一下代码来允许他们并行运行。我们只需让调度器从使用一个逻辑处理器变为两个：

```go
package main

import (
    "fmt"
    "runtime"
    "sync"
)

func main() {
    runtime.GOMAXPROCS(2)

    var wg sync.WaitGroup
    wg.Add(2)

    fmt.Println("Starting Go Routines")
    Go func() {
        defer wg.Done()

        for char := ‘ a ’ ; char < ‘ a ’ +26; char++ {
            fmt.Printf("%c ", char)
        }
    }()

    Go func() {
        defer wg.Done()

        for number := 1; number < 27; number++ {
            fmt.Printf("%d ", number)
        }
    }()

    fmt.Println("Waiting To Finish")
    wg.Wait()

    fmt.Println("\nTerminating Program")
}
```

这是这段程序的输出结果：

```
Starting Go Routines
Waiting To Finish
a b 1 2 3 4 c d e f 5 g h 6 i 7 j 8 k 9 10 11 12 l m n o p q 13 r s 14
t 15 u v 16 w 17 x y 18 z 19 20 21 22 23 24 25 26
Terminating Program
```

每次运行程序时我们都将得到不同的结果。调度器在每次运行中的调度过程是不确定的。我们可以看到两个协程确实并行运行了。两个协程在一个开始都立刻开始运行，并且他们都在争夺时间片来展示各自的结果。

## 结论

仅仅因为我们能为调度器增加逻辑处理并不意味着我们需要它。这就是为什么 Go 开发组会设置默认为一个单核逻辑处理器的原因（当前最新版本（1.11）默认逻辑处理器的个数设置为当前物理处理器的个数了）。仅仅知道任意添加逻辑处理器和并行运行 Goroutine 不一定会让程序拥有更好的性能。我们应该使用默认的配置，除非我们确实需要修改它们。

如果两个协程在同时使用一个临界资源会导致死锁问题，所以对临界资源的读写必须是原子性的。换句话说，读写必须在同一个协程内，否则的话就会导致死锁问题，学习更多关于[死锁](http://www.goinggo.net/2013/09/detecting-race-conditions-with-go.html) 的知识请阅读我们的文章。

在 Go 使用 Channels 可以写出安全优雅的并发程序，并且可以消除死锁的问题，你会再次找到并发的乐趣。既然我们已经知道协程如何工作，如何被调度和如何并行的运行，下一个我们要学习的就是 channels。

---

via: https://www.ardanlabs.com/blog/2014/01/concurrency-goroutines-and-gomaxprocs.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[xmge](https://github.com/xmge)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
