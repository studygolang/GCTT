首发于：https://studygolang.com/articles/21208

# 使用 Golang Goroutines 并发编程

大家好！在这篇文章中，我们将介绍如何在基于 Go 语言的项目中使用***goroutines***，以及如何提升项目运行时的性能。

## 视频教程

https://youtu.be/ARHXmR0_MGY

## 简介

如今，Go 是一种性能优良的程序设计语言，它拥有大量优秀的特性用于支持构建高性能的应用程序。它通过特有的***goroutines***和***channels***，重新定义了什么是构建高并发的项目。

使用***goroutines***可以快速地将串行的程序重构为并发程序，而不需要关心创建线程或线程池这种问题。但是，和所有的并发编程一样，这引入了一些危险，你必须在所有函数调用前敲***go***关键字时要考虑到。

## 什么是 Goroutines ？

首先，什么是 Goroutines ？***Goroutines***是一种由 Go 运行时系统管理的“极轻量级线程”。使用 Goroutines 可以创建异步并行的项目，执行这些并行任务要比执行对应的串行任务快很多。

> Goroutines 通常在初始化时仅需要分配 2KB 的栈空间，而线程通常需要占用 1MB。因此，Goroutines 远比线程要更轻量级。

Goroutines 通常被极少量的 OS 线程所复用，这意味着并发的 Go 项目与其他语言如 Java 等提供相同水平的性能时，需要更少的资源。创建一千个 Goroutines 通常至多需要一到两个 OS 线程；而在 Java 中，则需要创建一千个完成的线程，每个线程占用至少 1MB 的堆空间。

通过将成百上千个 Goroutines 映射到单个线程上，我们不必担心在应用程序中创建和销毁线程带来的性能上的影响。而创建或销毁 Goroutines 的开销极小，得益于它们轻量，以及 Go 处理这个过程的方式很高效。

## 一个简单的串行程序

在这个演示项目中，我们创建了一个方法，它接收一个 int 类型的参数 value，并在控制台中从 0 到 value 依次打印数字。此外，还加入一个 sleep 方法，以使在打印下一个数字前等待一秒钟：

```go
package main


import (
    "fmt"
    "time"
)


// a very simple function that we'll
// make asynchronous later on
func compute(value int) {
    for i := 0; i < value; i++ {
        time.Sleep(time.Second)
        fmt.Println(i)
    }
}

func main() {
    fmt.Println("Goroutine Tutorial")

    // sequential execution of our compute function
    compute(10)
    compute(10)

    // we scan fmt for input and print that to our console
    var input string
    fmt.Scanln(&input)

}
```

如果执行上面的代码，你会发现它从 0 到 9 每行依次打印一个数字，并且打印了两次。这个串行程序的总执行时间超过 20 秒。第 28 行添加 fmt.Scanln() 是为了在主函数不会在 Goroutines 执行之前结束。

## 改写为异步程序

如果我们不必关心程序中打印 0 到 n 的顺序，那么我们可以通过使用 Goroutines 加快这个程序，使其成为一个异步程序。

```go
package main


import (
    "fmt"
    "time"
)

// notice we've not changed anything in this function
// when compared to our previous sequential program
func compute(value int) {
    for i := 0; i < value; i++ {
        time.Sleep(time.Second)
        fmt.Println(i)
    }
}

func main() {
    fmt.Println("Goroutine Tutorial")

    // notice how we've added the 'go' keyword
    // in front of both our compute function calls
    Go compute(10)
    Go compute(10)
}
```

我们只需要在串行程序的函数调用前添加***go***关键字即可得到异步程序。在这里，我们创建了两个独立且并行执行的 Goroutines 。

但是，当你尝试运行这个程序时会发现，它在完整打印出我们预期的输出前就结束了。

为什么？

这是因为在异步函数执行完成之前，主函数执行完成了，因此所有未完成的 Goroutines 都会立即终止。

为了解决这个问题，我们可以加入一行 fmt.Scanln() ，这样程序就会在杀死 Goroutines 前，等待键盘输入：

```go
package main


import (
    "fmt"
    "time"
)

// notice we've not changed anything in this function
// when compared to our previous sequential program
func compute(value int) {
    for i := 0; i < value; i++ {
        time.Sleep(time.Second)
        fmt.Println(i)
    }
}

func main() {
    fmt.Println("Goroutine Tutorial")

    // notice how we've added the 'go' keyword
    // in front of both our compute function calls
    Go compute(10)
    Go compute(10)

    var input string
    fmt.Scanln(&input)
}
```

在终端中执行这个程序，可以看到控制台中打印了 0,0,1,1,2,2... 直到 ..9,9。如果对这个程序的执行计时会发现，接近 10 秒的时间就执行完了。

## 匿名 Goroutine 函数

在之前的例子中，我们关注如何通过**go**关键字创建一个命名的并发函数。同样，我们可以通过**go**关键字构建匿名并发函数：

```go
package main

import "fmt"

func main() {
    // we make our anonymous function concurrent using `go`
    Go func() {
        fmt.Println("Executing my Concurrent anonymouse function")
    }()

    fmt.Scanln()
}
```

## 结论

在这篇文章中，我们学习了如何使用 Go 开发并发应用程序。我们研究了什么是 Goroutines, 以及如何使用它们加速系统的各个部分并创建性能良好的应用程序。

希望这篇文章对你有用，欢迎在下方评论！

## 延伸阅读

- [Go sync.WaitGroup Tutorial](https://tutorialedge.net/golang/go-waitgroup-tutorial/)
- [Go Mutex Tutorial](https://tutorialedge.net/golang/go-mutex-tutorial/)
- [Go Channels Tutorial](https://tutorialedge.net/golang/go-channels-tutorial/)

---

via: https://tutorialedge.net/golang/concurrency-with-golang-goroutines/

作者：[tutorialedge.net](https://tutorialedge.net/golang)
译者：[ByKyleYao](https://github.com/ByKyleYao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
