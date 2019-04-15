# 使用Golang goroutines并发编程

大家好！在这篇文章中，我们将介绍如何在基于Go语言的项目中使用***goroutines***，以及如何提升项目运行时的性能。

### 视频教程

<https://youtu.be/ARHXmR0_MGY>

### 简介

如今，Go是一种性能优良的程序设计语言，它拥有大量优秀的特性用于支持构建高性能的应用程序。它通过特有的***goroutines***和***channels***，重新定义了什么是构建高并发的项目。

使用***goroutines***可以快速地将串行的程序重构为并发程序，而不需要关心创建线程或线程池这种问题。但是，和所有的并发编程一样，这引入了一些危险，你必须在所有函数调用前敲***go***关键字时要考虑到。

### 什么是Goroutines？

首先，什么是Goroutines？***Goroutines***是一种由go运行时系统管理的“极轻量级线程”。使用Goroutines可以创建异步并行的项目，执行这些并行任务要比执行对应的串行任务快很多。

> Goroutines通常在初始化时仅需要分配2KB的栈空间，而线程通常需要占用1MB。因此，Goroutines远比线程要更轻量级。

Goroutines通常被极少量的OS线程所复用，这意味着并发的go项目与其他语言如Java等提供相同水平的性能时，需要更少的资源。创建一千个Goroutines通常至多需要一到两个OS线程；而在Java中，则需要创建一千个完成的线程，每个线程占用至少1MB的堆空间。

通过将成百上千个Goroutines映射到单个线程上，我们不必担心在应用程序中创建和销毁线程带来的性能上的影响。而创建或销毁Goroutines的开销极小，得益于它们轻量，以及go处理这个过程的方式很高效。

### 一个简单的串行程序

在这个演示项目中，我们创建了一个方法，它接收一个int类型的参数value，并在控制台中从0到value依次打印数字。此外，还加入一个sleep方法，以使在打印下一个数字前等待一秒钟：

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

如果执行上面的代码，你会发现它从0到9每行依次打印一个数字，并且打印了两次。这个串行程序的总执行时间超过20秒。第28行添加fmt.Scanln()是为了在主函数不会在goroutines执行之前结束。

### 改写为异步程序

如果我们不必关心程序中打印0到n的顺序，那么我们可以通过使用goroutines加快这个程序，使其成为一个异步程序。

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
    go compute(10)
    go compute(10)
}
```

我们只需要在串行程序的函数调用前添加***go***关键字即可得到异步程序。在这里，我们创建了两个独立且并行执行的goroutines 。

但是，当你尝试运行这个程序时会发现，它在完整打印出我们预期的输出前就结束了。

为什么？

这是因为在异步函数执行完成之前，主函数执行完成了，因此所有未完成的goroutines都会立即终止。

为了解决这个问题，我们可以加入一行fmt.Scanln() ，这样程序就会在杀死goroutines前，等待键盘输入：

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
    go compute(10)
    go compute(10)

    var input string
    fmt.Scanln(&input)
}
```

在终端中执行这个程序，可以看到控制台中打印了0,0,1,1,2,2...直到..9,9。如果对这个程序的执行计时会发现，接近10秒的时间就执行完了。

### 匿名Goroutine 函数

在之前的例子中，我们关注如何通过**go**关键字创建一个命名的并发函数。同样，我们可以通过**go**关键字构建匿名并发函数：

```go
package main

import "fmt"

func main() {
    // we make our anonymous function concurrent using `go`
    go func() {
        fmt.Println("Executing my Concurrent anonymouse function")
    }()

    fmt.Scanln()
}
```

### 结论

在这篇文章中，我们学习了如何使用Go开发并发应用程序。我们研究了什么是Goroutines,以及如何使用它们加速系统的各个部分并创建性能良好的应用程序。

希望这篇文章对你有用，欢迎在下方评论！

### 延伸阅读

- [Go sync.WaitGroup Tutorial](https://tutorialedge.net/golang/go-waitgroup-tutorial/)
- [Go Mutex Tutorial](https://tutorialedge.net/golang/go-mutex-tutorial/)
- [Go Channels Tutorial](https://tutorialedge.net/golang/go-channels-tutorial/)

---

via: https://tutorialedge.net/golang/concurrency-with-golang-goroutines/

作者：[tutorialedge.net](https://tutorialedge.net/golang)
译者：[ByKyleYao](https://github.com/ByKyleYao)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出