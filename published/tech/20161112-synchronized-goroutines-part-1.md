首发于：https://studygolang.com/articles/14118

# goroutine 的同步（第一部分）

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/sync-goroutine/part1.jpeg)

假设 Go 程序启动了两个 goroutine：

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var v int
    var wg sync.WaitGroup
    wg.Add(2)
    go func() {
        v = 1
        wg.Done()
    }()
    go func() {
        fmt.Println(v)
        wg.Done()
    }()
    wg.Wait()
}
```

两个 goroutine 都对共享变量 *v* 进行操作。其中一个赋新值（写操作）而另一个打印变量的值（读操作）。

> *sync 包中的 [WaitGroup](https://golang.org/pkg/sync/#WaitGroup) 被用来等待两个非 main 的 goroutine 结束。否则，我们甚至都不能确保其中任意一个 goroutine 有被启动。*

由于不同的 goroutine 是相互独立的任务，它们进行的操作之间没有任何隐含的顺序。在上面的例子中，我们不清楚会打印出 `0` 还是 `1`。如果在 `fmt.Println` 被触发时，另一个 goroutine 已经执行了赋值语句 `v = 1`，那么输出会是 `1`。然而，在程序真正被执行之前一切都是未知的。换句话说，赋值语句和调用 `fmt.Println` 是无序的 —— 它们是并发的。

如果我们无法通过查看源码断定程序的行为，这是不好的。Go 的规范引入了内存操作（读和写）的偏序（partial order）关系（*先行发生原则* *happen before*）。这个顺序使我们能够推断程序的行为。另外，这门语言中的一些机制允许程序员强制实行操作的顺序。

在单个 goroutine 中，所有操作的顺序都与它们在源码中的位置一致。

```go
wg.Add(2)
wg.Wait()
```

上面例子中的函数调用是有序的，因为它们在同一个 goroutine 中 —— `wg.Add(2)` 先于 `wg.Wait()` 被执行。

## 1. 信道（Channel）

使用 channel 进行通信是最重要的同步方法。往 channel 中发送数据发生在接收数据之前：

```go
var v int
var wg sync.WaitGroup
wg.Add(2)
ch := make(chan int)
go func() {
    v = 1
    ch <- 1
    wg.Done()
}()
go func() {
    <-ch
    fmt.Println(v)
    wg.Done()
}()
wg.Wait()
```

新的东西是 *ch* 这个 channel。由于接收数据发生在往 channel 中发送数据之后，而发送数据发生在给 *v* 赋值之后，所以上述程序的输出永远是 `1`。

给 *v* 赋值 → 发送到 *ch* → 从 *ch* 接收 → 打印 *v*

第一个箭头和第三个箭头都是由同一个 goroutine 中的顺序确定的。使用 channel 进行通信带来了第二个箭头。最终，分散在两个 goroutine 中的操作是有序的。

## 2. sync 包

[sync](https://golang.org/pkg/sync/) 包提供了同步的原语。其中能解决我们的问题的是 [Mutex](https://golang.org/pkg/sync/#Mutex)。*sync.Mutex* 类型的变量 *lock* 保证第二次调用 `lock.Lock()` 发生在第一次调用 `lock.Unlock()` 之后。第三次调用 `lock.Lock()` 发生在第二次调用 `lock.Unlock()` 之后。一般来说，如果 *m* ＜ *n*，那么第 *n* 次调用 `lock.Lock()` 发生在第 *m* 次调用 `lock.Unlock()` 之后。让我们来看看在我们的同步问题中如何利用这个知识：

```go
var v int
var wg sync.WaitGroup
wg.Add(2)
var m sync.Mutex
m.Lock()
go func() {
    v = 1
    m.Unlock()
    wg.Done()
}()
go func() {
    m.Lock()
    fmt.Println(v)
    wg.Done()
}()
wg.Wait()
```

---

在后续的文章中将会展示更多关于使用 channel 进行通信的内容（比如怎样使用带缓存的 channel），[sync](https://golang.org/pkg/sync/) 包都提供了什么内容也会详细解释。

点赞以帮助别人发现这篇文章。如果你想得到新文章的更新，请关注我。

## 资源

- [Go 的内存模型 —— Go 编程语言](https://golang.org/ref/mem)
>The Go memory model specifies the conditions under which reads of a variable in one goroutine can be guaranteed to…
<br>*golang.org*

- [像个 Gopher 一样使用 *panic*](https://medium.com/golangspec/panicking-like-a-gopher-367a9ce04bb8)
> Errors while executing program in Go (after successful compilation and when OS process has been started) take the form…
<br>*medium.com*

*[保留部分版权](http://creativecommons.org/licenses/by/4.0/)*

*[Golang](https://medium.com/tag/golang?source=post)*
*[Programming](https://medium.com/tag/programming?source=post)*
*[Concurrency](https://medium.com/tag/concurrency?source=post)*
*[Synchronization](https://medium.com/tag/synchronization?source=post)*
*[Goroutines](https://medium.com/tag/goroutines?source=post)*

**喜欢读吗？给 Michał Łowicki 一些掌声吧。**

简单鼓励下还是大喝采，根据你对这篇文章的喜欢程度鼓掌吧。

---

via: https://medium.com/golangspec/synchronized-goroutines-part-i-4fbcdd64a4ec

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[krystollia](https://github.com/krystollia)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
