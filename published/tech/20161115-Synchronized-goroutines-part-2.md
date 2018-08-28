首发于：https://studygolang.com/articles/14478

# goroutine 的同步（第二部分）
> Channel 通信

第一部分介绍了发送与接收操作之间最直观的顺序关系：

> *向一个 Channel 中发送数据先于接收数据。*

于是，我们能够控制分布于两个 goroutine 中的操作的顺序。

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

（为了清晰，main 函数的定义和 import 语句被省略）

操作的顺序如下（*x* → *y* 代表 *x* 发生在 *y* 之前）：

`v = 1` → `ch <- 1` → `<-ch` → `fmt.Println(v)`

除上述这条外，还有更多的顺序规则，本篇将专注于 Channel。

## 发送 ↔ 接收

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/sync-goroutine/part2-1.jpeg)

除了上述规则，还有一条规则来补充它，这条规则说接收发生于发送完成之前：

发送开始 → 接收 → 发送结束

```go
var v, w int
var wg sync.WaitGroup
wg.Add(2)
ch := make(chan int)
go func() {
    v = 1
    ch <- 1
    fmt.Println(w)
    wg.Done()
}()
go func() {
    w = 2
    <-ch
    fmt.Println(v)
    wg.Done()
}()
wg.Wait()
```

因为有了这条新的规则，更多的操作之间有了顺序：

`w = 2` → `<-ch` → 发送操作 `ch <- 1` 结束 → `fmt.Println(w)`

通过保证赋值操作已经完成，我们就可以解决最初关于显示变量 `v` 的问题。

```go
go func() {
    v = 1
    <-ch
    wg.Done()
}()
go func() {
    ch <- 1
    fmt.Println(v)
    wg.Done()
}()
```

现在第二个 goroutine 会往 channel 中发送数据，它需要等待第一个 goroutine 中的赋值操作 `v = 1` 完成。发送操作在对应的接收操作后才能完成。

`v = 1` → `<-ch` → `ch <- 1` 结束 → `fmt.Println(v)`

## 关闭 channel

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/sync-goroutine/part2-2.jpeg)

当 channel 被关闭时，接收操作会立即返回 channel 中的数据类型的[零值](https://golang.org/ref/spec#The_zero_value)。

```go
ch := make(chan int)
close(ch)
fmt.Println(<-ch)  // prints 0
```

>*关闭 channel 发生在从已关闭的 channel 中接收零值之前*

把发送操作替换成调用自带函数 *close* 也可以解决我们最初的问题。

```go
go func() {
    v = 1
    close(ch)
    wg.Done()
}()
go func() {
    <-ch
    fmt.Println(v)
    wg.Done()
}()
```

操作的顺序是：

`v = 1` → `close(ch)` → `<-ch` → `fmt.Println(v)`

## 有缓存的 channel

到目前为止我们讨论了无缓存的 channel。有缓存的 channel 在缓存未满时发送操作不会阻塞，在缓存非空时接收操作也不会阻塞：

```go
ch := make(chan int, 1)
ch <- 1
fmt.Println(<-ch)
```

上述程序并不会以死锁告终，尽管在发送时并没有准备好的接受者。

对于有缓存的 channel，目前为止提到的所有规则都是成立的，除了说接收发生在发送结束之前这一条。原因很简单，（在缓存未满时），无需准备好的接受者，发送操作就可以结束。

>*对于容量为 c 的 channel，第 k 个接收发生在第 (k＋c) 个发送完成之前。*

假定缓存容量被设置为 3。前 3 个向 channel 发送数据的操作即使没有相应的接收语句也可以返回。但是为了第 4 个发送操作完成，必须有至少一个接收操作完成。

```go
var v int
var wg sync.WaitGroup
wg.Add(2)
ch := make(chan int, 3)
go func() {
    v = 1
    <-ch
    wg.Done()
}()
go func() {
    ch <- 1
    ch <- 1
    ch <- 1
    ch <- 1
    fmt.Println(v)
    wg.Done()
}()
wg.Wait()
```

这一小段用有缓存的 channel 解决了我们最初的问题。

---

点赞以帮助别人发现这篇文章。如果你想得到新文章的更新，请关注我。

## 资源
- [goroutine 的同步（第一部分）](https://medium.com/golangspec/synchronized-goroutines-part-i-4fbcdd64a4ec)
> Suppose that Go program starts two goroutines:
> medium.com

- [go 的内存模型 —— go 编程语言](https://golang.org/ref/mem)
>The Go memory model specifies the conditions under which reads of a variable in one goroutine can be guaranteed to…
> golang.org

- [go 语言规范 —— go 编程语言](https://golang.org/ref/spec)
>Go is a general-purpose language designed with systems programming in mind. It is strongly typed and garbage-collected…
> golang.org

*[保留部分版权](http://creativecommons.org/licenses/by/4.0/)*

*[Golang](https://medium.com/tag/golang?source=post)*
*[Programming](https://medium.com/tag/programming?source=post)*
*[Concurrency](https://medium.com/tag/concurrency?source=post)*
*[Channels](https://medium.com/tag/channel?source=post)*
*[Synchronization](https://medium.com/tag/synchronization?source=post)*

**喜欢读吗？给 Michał Łowicki 一些掌声吧。**

简单鼓励下还是大喝采，根据你对这篇文章的喜欢程度鼓掌吧。

---

via: https://medium.com/golangspec/synchronized-goroutines-part-ii-b1130c815c9d

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[krystollia](https://github.com/krystollia)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
