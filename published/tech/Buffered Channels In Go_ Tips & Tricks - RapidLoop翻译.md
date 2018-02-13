已发布：https://studygolang.com/articles/12395

# Go 语言的缓冲通道：提示和技巧

​Mahadevan Ramachandran • January 15

通道和 goroutine 是 Go 语言基于 CSP（ communicating sequential processes ，通信顺序进程）并发机制的核心部分。阅读本文可以学到一些关于channel的提示和技巧，尤其是“缓冲” channel ，在 “生产者-消费者” 情境中广泛使用了缓冲通道作为队列。

## 缓冲通道 = 队列

缓冲通道是固定容量的先进先出（FIFO）队列。容量在队列创建的时候就已经固定——其大小不能在运行时更改。			

```go
queue := make(chan Item, 10) // queue 的容量是 10
```

queue 中的一个元素最大可高达 64 KiB ,  而且 queue 可以存储指针和非指针元素。如果你坚持要用指针，抑或元素本身就是指针类型，那就需要由你自己来保证在 queue 中被使用的元素所指向的对象是有效的。

```go
queue := make(chan *Item, 10)
item1 := &Item{Foo: "bar"}
queue <- item1
item1.Foo = "baz" // 有效，但这不是好习惯！
```

生产者（将元素放入队列的代码）在把元素入队时可以选择是否阻塞。

```go
// queue 满后会发生阻塞
queue <- item

// 为了能不阻塞的放入元素, 代码如下:
var ok bool
select {
    case queue <- item:
        ok = true
    default:
        ok = false
}
// 在这里, "ok" is:
//   true  => 不阻塞的将元素入队
//   false => 元素没有入队, 会因为queue已满而阻塞
```

消费者通常从队列中取出元素并处理它们。如果队列为空并且消费者无事可做，就会发生阻塞，直到生产者放入一个元素。

```go
// 取出一个元素, 或者一直等待，直到可以取出元素
item := <- queue
```

如果不希望消费者等待，代码如下：

```go
var ok bool
select {
    case item = <- queue:
        ok = true
    default:
        ok = false
}
// 在这里, "ok" is:
//   true  => 从queue中取出元素item (或者queue已经关闭，见下)
//   false => 没有取出元素, queue为空而发生阻塞
```

##  关闭缓冲通道

缓冲通道最好是由生产者关闭，通道关闭事件会被发送给消费者。如果你需要在生产者或者消费者之外关闭通道，那么你必须使用外部同步来确保生产者不会试图向已关闭的通道写入（这会引发一个 panic ）。

```go
close(queue)  // 关闭队列

close(queue)  // "panic: 关闭一个已关闭的通道"
```

## 读取或者写入关闭的通道

你能向已关闭的通道写入么？当然不能。

```go
queue <- item // "panic: 向已关闭的通道写入"
```

那么能从已关闭的通道读取么？事实上，在往下翻之前，请先猜测下这段代码的输出结果：

```go
package main

import "fmt"

func main() {
    queue := make(chan int, 10)

    queue <- 10
    queue <- 20

    close(queue)

    fmt.Println(<-queue)
    fmt.Println(<-queue)
    fmt.Println(<-queue)
}
```

这里有运行以上代码的 [链接](https://play.golang.org/p/ot87ro27tFk) .

(译者注：以下是运行结果)

```
10
20
0
```

吃惊吧？如果你猜错了，记住你要先看这里！:-)

在已关闭的通道上的读取行为比较特殊：

- 如果还有元素没有被取出，那么读取操作会照常进行。
- 如果队列已空并且已被关闭，读取不会阻塞。
- 在为空并且已经关闭的通道上读取时会返回通道中元素类型的 “零值”。

这些能让你明白为什么上面的程序会打印出这种结果。但是你又怎么区分读取到的数据是否有效呢？毕竟，“零值”也可能是有效值。答案在下面：

```go
item, valid := <- queue
// 在这里, "valid" 取值:
//    true  => "item" 有效
//    false => "queue" 已经关闭, "item" 只是一个 “零值”
```

因此你在写消费者代码的时候可以这样：

```go
for {
    item, valid := <- queue
    if !valid {
        break
    }
    // 处理 item
}
// 到这里，所有被放入到 queue 中的元素都已经处理完毕，
// 并且 queue 也已经关闭
```

其实，“for..range” 循环是一种更加简单的写法:

```go
for item := range queue {
    // 处理 item
}
// 到这里，所有被放入到 queue 中的元素都已经处理完毕，
// 并且 queue 也已经关闭
```

最后，我们可以把非阻塞和检查元素有效性结合到一起：

```go
var ok, valid bool
select {
    case item, valid = <- queue:
        ok = true
    default:
        ok = false
}
// 到这里:
//   ok && valid  => item 有效, 可以使用
//   !ok          => 通道没有关闭，但是通道为空，稍后重试
//   ok && !valid => 通道已经关闭，退出轮询
```

##  咨询和训练

需要帮助获得一个使用 Golang 的项目？我们在创建和运行生产级 Go 平台软件解决方案领域拥有丰富经验。我们可以帮助你架构和设计 Go 平台项目，或者为使用 Go 工作的团队提供建议和监督。我们也会为希望开展Go项目的团队提供培训或者提升 Golang 知识。[这里发现更多](https://www.rapidloop.com/training) 或者 [马上联系我们](https://www.rapidloop.com/contact) 来讨论你的需求！

**Mahadevan Ramachandran**

Co-founder & CEO, RapidLoop 
[@mdevanr](https://twitter.com/mdevanr)

```
----------------

via: https://www.rapidloop.com/blog/golang-channels-tips-tricks.html

作者：[Aaron Schlesinger](https://medium.com/@arschles)
译者：[sunzhaohao](https://github.com/sunzhaohao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
```