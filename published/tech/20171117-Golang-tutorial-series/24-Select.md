已发布：https://studygolang.com/articles/12522

# 第 24 篇：Select
欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 24 篇。

## 什么是 select？
`select` 语句用于在多个发送/接收信道操作中进行选择。`select` 语句会一直阻塞，直到发送/接收操作准备就绪。如果有多个信道操作准备完毕，`select` 会随机地选取其中之一执行。该语法与 `switch` 类似，所不同的是，这里的每个 `case` 语句都是信道操作。我们好好看一些代码来加深理解吧。

## 示例
```go
package main

import (
    "fmt"
    "time"
)

func server1(ch chan string) {
    time.Sleep(6 * time.Second)
    ch <- "from server1"
}
func server2(ch chan string) {
    time.Sleep(3 * time.Second)
    ch <- "from server2"

}
func main() {
    output1 := make(chan string)
    output2 := make(chan string)
    go server1(output1)
    go server2(output2)
    select {
    case s1 := <-output1:
        fmt.Println(s1)
    case s2 := <-output2:
        fmt.Println(s2)
    }
}
```
[在线运行程序](https://play.golang.org/p/3_yaJSoSpG)

在上面程序里，`server1` 函数（第 8 行）休眠了 6 秒，接着将文本 `from server1` 写入信道 `ch`。而 `server2` 函数（第 12 行）休眠了 3 秒，然后把 `from server2` 写入了信道 `ch`。

而 `main` 函数在第 20 行和第 21 行，分别调用了 `server1` 和 `server2` 两个 Go 协程。

在第 22 行，程序运行到了 `select` 语句。`select` 会一直发生阻塞，除非其中有 case 准备就绪。在上述程序里，`server1` 协程会在 6 秒之后写入 `output1` 信道，而`server2` 协程在 3 秒之后就写入了 `output2` 信道。因此 `select` 语句会阻塞 3 秒钟，等着 `server2` 向 `output2` 信道写入数据。3 秒钟过后，程序会输出：

```
from server2
```

然后程序终止。

## select 的应用
在上面程序中，函数之所以取名为 `server1` 和 `server2`，是为了展示 `select` 的实际应用。

假设我们有一个关键性应用，需要尽快地把输出返回给用户。这个应用的数据库复制并且存储在世界各地的服务器上。假设函数 `server1` 和 `server2` 与这样不同区域的两台服务器进行通信。每台服务器的负载和网络时延决定了它的响应时间。我们向两台服务器发送请求，并使用 `select` 语句等待相应的信道发出响应。`select` 会选择首先响应的服务器，而忽略其它的响应。使用这种方法，我们可以向多个服务器发送请求，并给用户返回最快的响应了。:）

## 默认情况
在没有 case 准备就绪时，可以执行 `select` 语句中的默认情况（Default Case）。这通常用于防止 `select` 语句一直阻塞。

```go
package main

import (
    "fmt"
    "time"
)

func process(ch chan string) {
    time.Sleep(10500 * time.Millisecond)
    ch <- "process successful"
}

func main() {
    ch := make(chan string)
    go process(ch)
    for {
        time.Sleep(1000 * time.Millisecond)
        select {
        case v := <-ch:
            fmt.Println("received value: ", v)
            return
        default:
            fmt.Println("no value received")
        }
    }

}
```
[在线运行程序](https://play.golang.org/p/8xS5r9g1Uy)

上述程序中，第 8 行的 `process` 函数休眠了 10500 毫秒（10.5 秒），接着把 `process successful` 写入 `ch` 信道。在程序中的第 15 行，并发地调用了这个函数。

在并发地调用了 `process` 协程之后，主协程启动了一个无限循环。这个无限循环在每一次迭代开始时，都会先休眠 1000 毫秒（1 秒），然后执行一个 select 操作。在最开始的 10500 毫秒中，由于 `process` 协程在 10500 毫秒后才会向 `ch` 信道写入数据，因此 `select` 语句的第一个 case（即 `case v := <-ch:`）并未就绪。所以在这期间，程序会执行默认情况，该程序会打印 10 次 `no value received`。

在 10.5 秒之后，`process` 协程会在第 10 行向 `ch` 写入 `process successful`。现在，就可以执行 `select` 语句的第一个 case 了，程序会打印 `received value:  process successful`，然后程序终止。该程序会输出：

```
no value received
no value received
no value received
no value received
no value received
no value received
no value received
no value received
no value received
no value received
received value:  process successful
```

## 死锁与默认情况
```go
package main

func main() {
    ch := make(chan string)
    select {
    case <-ch:
    }
}
```
[在线运行程序](https://play.golang.org/p/za0GZ4o7HH)

上面的程序中，我们在第 4 行创建了一个信道 `ch`。我们在 `select` 内部（第 6 行），试图读取信道 `ch`。由于没有 Go 协程向该信道写入数据，因此 `select` 语句会一直阻塞，导致死锁。该程序会触发运行时 `panic`，报错信息如下：

```
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan receive]:
main.main()
    /tmp/sandbox416567824/main.go:6 +0x80
```

如果存在默认情况，就不会发生死锁，因为在没有其他 case 准备就绪时，会执行默认情况。我们用默认情况重写后，程序如下：

```go
package main

import "fmt"

func main() {
    ch := make(chan string)
    select {
    case <-ch:
    default:
        fmt.Println("default case executed")
    }
}
```
[在线运行程序](https://play.golang.org/p/Pxsh_KlFUw)

以上程序会输出：

```
default case executed
```

如果 `select` 只含有值为 `nil` 的信道，也同样会执行默认情况。

```go
package main

import "fmt"

func main() {
    var ch chan string
    select {
    case v := <-ch:
        fmt.Println("received value", v)
    default:
        fmt.Println("default case executed")

    }
}
```
[在线运行程序](https://play.golang.org/p/IKmGpN61m1)

在上面程序中，`ch` 等于 `nil`，而我们试图在 `select` 中读取 `ch`（第 8 行）。如果没有默认情况，`select` 会一直阻塞，导致死锁。由于我们在 `select` 内部加入了默认情况，程序会执行它，并输出：

```
default case executed
```

## 随机选取
当 `select` 由多个 case 准备就绪时，将会随机地选取其中之一去执行。

```go
package main

import (
    "fmt"
    "time"
)

func server1(ch chan string) {
    ch <- "from server1"
}
func server2(ch chan string) {
    ch <- "from server2"

}
func main() {
    output1 := make(chan string)
    output2 := make(chan string)
    go server1(output1)
    go server2(output2)
    time.Sleep(1 * time.Second)
    select {
    case s1 := <-output1:
        fmt.Println(s1)
    case s2 := <-output2:
        fmt.Println(s2)
    }
}
```
[在线运行程序](https://play.golang.org/p/vJ6VhVl9YY)

在上面程序里，我们在第 18 行和第 19 行分别调用了 `server1` 和 `server2` 两个 Go 协程。接下来，主程序休眠了 1 秒钟（第 20 行）。当程序控制到达第 21 行的 `select` 语句时，`server1` 已经把 `from server1` 写到了 `output1` 信道上，而 `server2` 也同样把 `from server2` 写到了 `output2` 信道上。因此这个 `select` 语句中的两种情况都准备好执行了。如果你运行这个程序很多次的话，输出会是 `from server1` 或者 `from server2`，这会根据随机选取的结果而变化。

请在你的本地系统上运行这个程序，获得程序的随机结果。因为如果你在 playground 上在线运行的话，它的输出总是一样的，这是由于 playground 不具有随机性所造成的。

## 这下我懂了：空 select
```go
package main

func main() {
    select {}
}
```
[在线运行程序](https://play.golang.org/p/u8hErIxgxs)

你认为上面代码会输出什么？

我们已经知道，除非有 case 执行，select 语句就会一直阻塞着。在这里，`select` 语句没有任何 case，因此它会一直阻塞，导致死锁。该程序会触发 panic，输出如下：

```
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [select (no cases)]:
main.main()
    /tmp/sandbox299546399/main.go:4 +0x20
```

本教程到此结束。祝你愉快。

**上一教程 - [缓冲信道和工作池](https://studygolang.com/articles/12512)**

**下一教程 - [Mutex](https://studygolang.com/articles/12598)**

---

via: https://golangbot.com/select/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
