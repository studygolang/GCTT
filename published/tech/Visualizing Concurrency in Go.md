已发布：https://studygolang.com/articles/11908

# 可视化 Go 语言中的并发

本文作者提供了在 2016 的 GopherCon 上的关于 Go 并发可视化的[主题演讲视频](https://www.youtube.com/watch?v=KyuFeiG3Y60)。

Go 语言一个鲜明的优点就是内置的基于 [CSP](https://en.wikipedia.org/wiki/Communicating_sequential_processes) 的并发实现。Go 可以说是一个为了并发而设计的语言，允许我们使用它构建复杂的并发流水线。但是开发者是否在脑海中想象过不同的并发模式呢，它们在你的大脑中是怎样的形状？

你肯定想过这些！我们都会靠多种多样的想象来思考。如果我让你想象一下 1-100 的数字，你会下意识地在脑海中闪过这些数字的图像。比如说我会把它想象成一条从我出发的直线，到 20 时它会右转 90 度并且一直延伸到 1000。这是因为我上幼儿园的时候，卫生间墙上写满了数字，20 正好在角落上。你们脑中肯定有自己想象的数字形象。另一个常见的例子就是一年四季的可视化，有些人把它想象成一个盒子，另外一些人把它想象成是圆圈。

无论如何，我想要用 Go 和 WebGL 分享我想象的一些常用并发模式的可视化。这些可视化或多或少地代表了我头脑中的并发编程方法。如果能知道我们之间对并发可视化想象的差异，肯定是一件很有趣的事情。我尤其想要知道 Rob Pike 和 Sameer Ajmani 是怎么想象并发的，那一定很有意思。

现在，我们就从最简单的 “Hello, Concurrent World” 开始，来了解我脑海中的并发世界吧。

## Hello, Concurrent World

这个例子的代码很简单，只包含一个 channel，一个 goroutine，一个读操作和一个写操作。

```go
package main

func main() {
    // create new channel of type int
    ch := make(chan int)

    // start new anonymous goroutine
    go func() {
        // send 42 to channel
        ch <- 42
    }()
    // read from channel
    <-ch
}
```
[WebGL 动画界面](http://divan.github.io/demos/hello)

![hello](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/hello.gif)

在这张图中，蓝色的线代表 goroutine 的时间轴。连接 `main` 和 `go#19` 的蓝线是用来标记 goroutine 的起始和终止并且表示父子关系的。红色的箭头代表的是 send/recv 操作。尽管 send/recv 操作是两个独立的操作，但是我试着将它们表示成一个操作 `从 A 发送到 B`。右边蓝线上的 `#19` 是该 goroutine 的内部 ID，可以通过 Scott Mansfield 在 [Goroutine IDs](http://blog.sgmansfield.com/2015/12/goroutine-ids/) 一文中提到的技巧获取。

## 计时器（Timers）

事实上，我们可以通过简单的几个步骤编写一个计时器：创建一个 channel，启动一个 goroutine 以给定间隔往 channel 中写数据，将这个 chennel 返回给调用者。调用者阻塞地从 channel 中读，就会得到一个精准的时钟。让我们来试试调用这个程序 24 次并且将过程可视化。

```go
package main

import "time"

func timer(d time.Duration) <-chan int {
    c := make(chan int)
    go func() {
        time.Sleep(d)
        c <- 1
    }()
    return c
}

func main() {
    for i := 0; i < 24; i++ {
        c := timer(1 * time.Second)
        <-c
    }
}
```
[WebGL 动画界面](http://divan.github.io/demos/timers/)

![timers](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/timers.gif)

这个效果是不是很有条理？

## 乒乓球（Ping-pong）

这个例子是我从 Google 员工 Sameer Ajmani 的一次演讲 ["Advanced Go Concurrency Patterns"](https://talks.golang.org/2013/advconc.slide#1) 中找到的。当然，这并不是一个很高阶的并发模型，但是对于 Go 语言并发的新手来说是很有趣的。

在这个例子中，我们定义了一个 channel 来作为“乒乓桌”。乒乓球是一个整形变量，代码中有两个 goroutine “玩家”通过增加乒乓球的 counter 在“打球”。

```go
package main

import "time"

func main() {
    var Ball int
    table := make(chan int)
    go player(table)
    go player(table)

    table <- Ball
    time.Sleep(1 * time.Second)
    <-table
}

func player(table chan int) {
    for {
        ball := <-table
        ball++
        time.Sleep(100 * time.Millisecond)
        table <- ball
    }
}
```
[WebGL 动画界面](http://divan.github.io/demos/pingpong/)

![Ping-pong](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/pingpong.gif)

我建议你点击 “WebGL 动画界面” 链接，从不同角度看看这个模型，并且试试它减速，加速的效果。

现在，我们给这个模型添加一个玩家（goroutine）。

```go
 go player(table)
 go player(table)
 go player(table)
```
[WebGL 动画界面](http://divan.github.io/demos/pingpong3/)

![Ping-pong2](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/pingpong3.gif)

我们可以看到每个 goroutine 都有序地“打到球”，你可能会好奇这个行为的原因。那么，为什么这三个 goroutine 始终按照一定顺序接收到 ball 呢？

答案很简单，Go 运行时会对每个 channel 的所有接收者维护一个 [FIFO 队列 ](https://github.com/golang/go/blob/master/src/runtime/chan.go#L34)。在我们的例子中，每个 goroutine 会在它将 ball 传给 channel 之后就开始等待 channel，所以它们在队列里的顺序总是一定的。让我们增加 goroutine 的数量，看看顺序是否仍然保持一致。

```go
for i := 0; i < 100; i++ {
    go player(table)
}
```
[WebGL 动画界面](http://divan.github.io/demos/pingpong100/)

![Ping-pong100](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/pingpong100.gif)

很明显，它们的顺序仍然是一定的。我们可以创建一百万个 goroutine 去尝试，但是上面的实验已经足够让我们得出结论了。接下来，让我们来看看一些不一样的东西，比如说通用的消息模型。

## 扇入模式（Fan-In）

扇入（fan-in）模式在并发世界中广泛使用。扇出（fan-out）模式与其相反，我们会在下面介绍。简单来说，扇入模式就是一个函数从多个输入源读取数据并且复用到单个 channel 中。比如说：

```go
package main

import (
    "fmt"
    "time"
)

func producer(ch chan int, d time.Duration) {
    var i int
    for {
        ch <- i
        i++
        time.Sleep(d)
    }
}

func reader(out chan int) {
    for x := range out {
        fmt.Println(x)
    }
}

func main() {
    ch := make(chan int)
    out := make(chan int)
    go producer(ch, 100*time.Millisecond)
    go producer(ch, 250*time.Millisecond)
    go reader(out)
    for i := range ch {
        out <- i
    }
}
```
[WebGL 动画界面](http://divan.github.io/demos/fanin/)

![Fan-In](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/fanin.gif)

我们能看到，第一个 `producer` 每隔一百毫秒生成一个值，第二个 `producer` 每隔 250 毫秒生成一个值，但是 `reader` 会立即接收它们的值。main 函数中的 for 循环高效地接收了 channel 发送的所有信息。

## 工作者模式（Workers）

与扇入模式相反的模式叫做扇出（fan-out）或者工作者（workers）模式。多个 goroutine 可以从相同的 channel 中读数据，利用多核并发完成自身的工作，这就是工作者（workers）模式的由来。在 Go 中，这个模式很容易实现，只需要启动多个以 channel 作为参数的 goroutine，主函数传数据给这个 channel，数据分发和复用会由 Go 运行环境自动完成。

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func worker(tasksCh <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        task, ok := <-tasksCh
        if !ok {
            return
        }
        d := time.Duration(task) * time.Millisecond
        time.Sleep(d)
        fmt.Println("processing task", task)
    }
}

func pool(wg *sync.WaitGroup, workers, tasks int) {
    tasksCh := make(chan int)

    for i := 0; i < workers; i++ {
        go worker(tasksCh, wg)
    }

    for i := 0; i < tasks; i++ {
        tasksCh <- i
    }

    close(tasksCh)
}

func main() {
    var wg sync.WaitGroup
    wg.Add(36)
    go pool(&wg, 36, 50)
    wg.Wait()
}
```
[WebGL 动画界面](http://divan.github.io/demos/workers/)

![Workers](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/workers.gif)

在这里需要提一下并行结构（parallelism）。我们可以看到，动图中所有的 goroutine 都是平行“延伸”，等待 channel 给它们发数据来运行的。我们还可以注意到两个 goroutine 接收数据之间几乎是没有停顿的。不幸的是，这个动画并没有用颜色区分一个 goroutine 是在等数据还是在执行工作，这个动画是在 `GOMAXPROCS=4` 的情况下录制的，所以只有 4 个 goroutine 能够同时运行。我们将会在下文汇总讨论这个主题。

现在，我们来写更复杂一点的代码，启动带有子工作者的工作者（subworkers）：

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

const (
    WORKERS    = 5
    SUBWORKERS = 3
    TASKS      = 20
    SUBTASKS   = 10
)

func subworker(subtasks chan int) {
    for {
        task, ok := <-subtasks
        if !ok {
            return
        }
        time.Sleep(time.Duration(task) * time.Millisecond)
        fmt.Println(task)
    }
}

func worker(tasks <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        task, ok := <-tasks
        if !ok {
            return
        }

        subtasks := make(chan int)
        for i := 0; i < SUBWORKERS; i++ {
            go subworker(subtasks)
        }
        for i := 0; i < SUBTASKS; i++ {
            task1 := task * i
            subtasks <- task1
        }
        close(subtasks)
    }
}

func main() {
    var wg sync.WaitGroup
    wg.Add(WORKERS)
    tasks := make(chan int)

    for i := 0; i < WORKERS; i++ {
        go worker(tasks, &wg)
    }

    for i := 0; i < TASKS; i++ {
        tasks <- i
    }

    close(tasks)
    wg.Wait()
}
```
[WebGL 动画界面](http://divan.github.io/demos/workers2/)

![Workers](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/workers2.gif)

当然，我们可以将工作者的数量或子工作者的数量设得更高，但是在这里我们试着不让动画效果变得太复杂。

Go 中还存在比这更酷的扇出模式，比如动态工作者/子工作者数量，使用 channel 来传输 channel，但是现在的动画模拟应该已经可以解释扇出模型的含义了。

## 服务器（Servers）

下一个要说的常用模式和扇出相似，但是它会在短时间内生成多个 goroutine 来完成某些任务。这个模式常被用来实现服务器 -- 创建一个监听器，在循环中运行 accept() 并针对每个接受的连接启动 goroutine 来完成指定任务。这个模式很形象并且它能尽可能地简化服务器 handler 的实现。让我们来看一个简单的例子：

```go
package main

import "net"

func handler(c net.Conn) {
    c.Write([]byte("ok"))
    c.Close()
}

func main() {
    l, err := net.Listen("tcp", ":5000")
    if err != nil {
        panic(err)
    }
    for {
        c, err := l.Accept()
        if err != nil {
            continue
        }
        go handler(c)
    }
}
```
[WebGL 动画界面](http://divan.github.io/demos/servers/)

![Server](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/servers.gif)

从并发的角度看好像什么事情都没有发生。当然，表面平静，内在其实风起云涌，完成了一系列复杂的操作，只是复杂性都被隐藏了，毕竟 [Simplicity is complicated.](https://www.youtube.com/watch?v=rFejpH_tAHM)

但是让我们回归到并发的角度，给我们的服务器添加一些交互功能。比如说，我们定义一个 logger 以独立的 goroutine 的形式来记日志，每个 handler 想要异步地通过这个 logger 去写数据。

```go
package main

import (
    "fmt"
    "net"
    "time"
)

func handler(c net.Conn, ch chan string) {
    ch <- c.RemoteAddr().String()
    c.Write([]byte("ok"))
    c.Close()
}

func logger(ch chan string) {
    for {
        fmt.Println(<-ch)
    }
}

func server(l net.Listener, ch chan string) {
    for {
        c, err := l.Accept()
        if err != nil {
            continue
        }
        go handler(c, ch)
    }
}

func main() {
    l, err := net.Listen("tcp", ":5000")
    if err != nil {
        panic(err)
    }
    ch := make(chan string)
    go logger(ch)
    go server(l, ch)
    time.Sleep(10 * time.Second)
}
```
[WebGL 动画界面](http://divan.github.io/demos/servers2/)

![Server2](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/servers2.gif)

这个例子就很形象地展示了服务器处理请求的过程。我们容易发现 logger 在存在大量连接的情况下会成为性能瓶颈，因为它需要对每个连接发送的数据进行接收，编码等耗时的操作。我们可以用上文提到的扇出模式来改进这个服务器模型。

让我们来看看代码和动画效果：

```go
package main

import (
    "net"
    "time"
)

func handler(c net.Conn, ch chan string) {
    addr := c.RemoteAddr().String()
    ch <- addr
    time.Sleep(100 * time.Millisecond)
    c.Write([]byte("ok"))
    c.Close()
}

func logger(wch chan int, results chan int) {
    for {
        data := <-wch
        data++
        results <- data
    }
}

func parse(results chan int) {
    for {
        <-results
    }
}

func pool(ch chan string, n int) {
    wch := make(chan int)
    results := make(chan int)
    for i := 0; i < n; i++ {
        go logger(wch, results)
    }
    go parse(results)
    for {
        addr := <-ch
        l := len(addr)
        wch <- l
    }
}

func server(l net.Listener, ch chan string) {
    for {
        c, err := l.Accept()
        if err != nil {
            continue
        }
        go handler(c, ch)
    }
}

func main() {
    l, err := net.Listen("tcp", ":5000")
    if err != nil {
        panic(err)
    }
    ch := make(chan string)
    go pool(ch, 4)
    go server(l, ch)
    time.Sleep(10 * time.Second)
}
```
[WebGL 动画界面](http://divan.github.io/demos/servers3/)

![Server3](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/servers3.gif)

在这个例子中，我们把记日志的任务分布到了 4 个 goroutine 中，有效地改善了 logger 模块的吞吐量。但是从动画中仍然可以看出，logger 仍然是系统中最容易出现性能问题的地方。如果上千个连接同时调用 logger 记日志， 现在的 logger 模块仍然可能会出现性能瓶颈。当然，相比于之前的实现，它的阈值已经高了很多。

## 并发质数筛选法（Concurrent Prime Sieve）

看够了扇入/扇出模型，我们现在来看看具体的并行算法。让我们来讲讲我最喜欢的并行算法之一：并行质数筛选法。这个算法是我从 [Go Concurrency Patterns](https://talks.golang.org/2012/concurrency.slide) 这个演讲中看到的。质数筛选法（埃拉托斯特尼筛法）是在一个寻找给定范围内最大质数的古老算法。它通过一定的顺序筛掉多个质数的乘积，最终得到想要的最大质数。但是其原始的算法在多核机器上并不高效。

这个算法的并行版本定义了多个 goroutine，每个 goroutine 代表一个已经找到的质数，同时有多个 channel 用来从 generator 传输数据到 filter。每当找到质数时，这个质数就会被一层层 channel 送到 main 函数来输出。当然，这个算法也不够高效，尤其是当你需要寻找一个很大的质数或者在寻找时间复杂度最低的算法时，但它的思想很优雅。

```go
// A concurrent prime sieve
package main

import "fmt"

// Send the sequence 2, 3, 4, ... to channel 'ch'.
func Generate(ch chan<- int) {
    for i := 2; ; i++ {
        ch <- i // Send 'i' to channel 'ch'.
    }
}

// Copy the values from channel 'in' to channel 'out',
// removing those divisible by 'prime'.
func Filter(in <-chan int, out chan<- int, prime int) {
    for {
        i := <-in // Receive value from 'in'.
        if i%prime != 0 {
            out <- i // Send 'i' to 'out'.
        }
    }
}

// The prime sieve: Daisy-chain Filter processes.
func main() {
    ch := make(chan int) // Create a new channel.
    go Generate(ch)      // Launch Generate goroutine.
    for i := 0; i < 10; i++ {
        prime := <-ch
        fmt.Println(prime)
        ch1 := make(chan int)
        go Filter(ch, ch1, prime)
        ch = ch1
    }
}
```
[WebGL 动画界面](http://divan.github.io/demos/primesieve/)

![Prime](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/primesieve.gif)

这个算法的模拟动画也同样很优雅形象，能帮助我们理解这个算法。算法中的 generate 这个函数发送从 2 开始的所有的整形数，传递给 filter 所在的 goroutine, 每个质数都会生成一个 filter 的 goroutine。 如果你在动画链接中从上往下看，你会发现所有传给 main 函数的数都是质数。最后总要的还是，这个算法在 3D 模拟中特别优美。

## GOMAXPROCS

现在，让我们回到上文的工作者模式上。还记得我提到过这个例子是在 `GOMAXPROCS = 4` 的条件下运行的吗？这是因为所有的动画效果都不是艺术品，它们都是用实际运行状态模拟而得的。

让我们看看 `GOMAXPROCS` 的定义：

```
GOMAXPROCS sets the maximum number of CPUs that can be executing simultaneously.
```
定义中的 CPU 指的是逻辑 CPU。我之前稍微修改了一下工作者的例子让每个 goroutine 都做一点会占用 CPU 时间的事情，然后我设置了不同的 `GOMAXPROCS` 值，重复运行这个例子，运行环境是一个 2 CPU, 共 24 核的机器，系统是 Linux。

以下两个图中，第一张是运行在 1 个核上时的动画效果，第二张是运行在 24 核上时的动画效果。

[WebGL 动画界面 1核](http://divan.github.io/demos/gomaxprocs1/)

![1Core-Worker](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/gomaxprocs1.gif)


[WebGL 动画界面 24核](http://divan.github.io/demos/gomaxprocs24/)

![24Core-Worker](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/gomaxprocs24.gif)

显而易见，这些动画模拟花费的时间是不同的。当 `GOMAXPROCS` 是 1 的时候，只有一个工作者结束了自己的任务以后，下一个工作者才会开始执行。而在 `GOMAXPROCS` 是 24 的情况下，整个程序的执行速度变化非常明显，相比之下，一些多路复用的开销变得微不足道了。

尽管如此，我们也要知道，增大 `GOMAXPROCS` 的并不总是能够提高性能，在有些情况下它甚至会使程序的性能变差。

## Goroutine 泄露（Goroutines leak）

Go 并发中还有什么使我们能可视化的呢？ goroutine 泄露是我能想到的一个场景。当你启动一个 goroutine 但是它在你的代码外陷入了错误状态，或者是你启动了很多带有死循环的 goroutine 时，goroutine 泄露就发生了。

我仍然记得我第一次遇到 goroutine 泄露时，我脑子里想象的可怕场景。紧接着的周末，我就写了 expvarmon (一个 Go 应用的资源监控工具)。现在，我可以用 WebGL 来描绘当时在我脑海中的景象了。

[WebGL 动画界面](http://divan.github.io/demos/leak/)

![Goroutines leak](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/leak.gif)

这个图中所有的蓝线都是浪费的系统资源，并且会成为你的应用的“定时炸弹”。

## 并行不是并发（Parallelism is not Concurrency）

最后，我想谈一下并行和并发的区别。这个主题已经在[Parallelism Is Not Concurrency](https://existentialtype.wordpress.com/2011/03/17/parallelism-is-not-concurrency/) 和 [Parallelism /= Concurrency
](https://ghcmutterings.wordpress.com/2009/10/06/parallelism-concurrency/) 中探讨过了，Rob Pike 在一篇[演讲](https://www.youtube.com/watch?v=cN_DpYBzKso)中提到了这个问题，这是我认为必看的主题演讲之一。

简单来说，**并行是指同时运行多个任务，而并发是一种程序架构的方法。**

因此，带有并发的程序并不一定是并行的，这两个概念在一定程度上是互不相关的。我们在关于 `GOMAXPROCS` 的论述中就提到了这一点。 

在这里我不想重复上面的链接中的语言。我相信，有图有真相。我会通过动画模拟来告诉你它们之间的不同。下面这张描述的是并行————许多任务同时运行：

[WebGL 动画界面](http://divan.github.io/demos/parallelism1/)

![Parallelism1](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/parallelism1.gif)

这个也是并行：

[WebGL 动画界面](http://divan.github.io/demos/parallelism2/)

![Parallelism2](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/parallelism2.gif)

但是这个是并发：

![Server](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/primesieve.gif)

这个也是并行(嵌套的工作者)：

![Workers2](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/workers2.gif)

这个也是并发的：

![pingpong100](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/pingpong100.gif)

# 如何生成这些动画

为了生成这些动画，我写了两个程序：gotracer 和 gothree.js 库。首先，gotracer 会做这样的事情：

    1. 解析 Go 程序中的 AST 树并且从中插入输出并行相关信息的代码，比如启动/结束 goroutine, 创建一个 channel，发送、接收数据。
    2. 运行生成的程序。
    3. 分析这些输出并且生成描述这些事件的 JSON 文件。

JSON 文件的样例如下：

![JSON](https://raw.githubusercontent.com/studygolang/gctt-images/master/visualizing-concurrency/sshot_json.png)

接下来，gothree.js 使用 [Three.js](http://threejs.org/) 这个能够用 WebGL 生成 3D 图像的的库来绘制动画。

这种方法的使用场景非常有限。我必须精准地选择例子，重命名 channel 和 goroutine 来输出一个正确的 trace。这个方法也无法关联两个 goroutine 中的相同但不同名的 channel，更不用说识别通过 channel 传送的 channel 了。这个方法生成的时间戳也会出现问题，有时候输出信息到标准输出会比传值花费更多的时间，所以我为了得到正确的动画不得不在某些情况下让 goroutine 等待一些时间。

这就是我并没有将这份代码开源的原因。我正在尝试使用 Dmitry Vyukov 的 [execution tracer](https://golang.org/cmd/trace/)，它看起来能提供足够多的信息，但并不包含 channel 传输的值。也许有更好的方法来实现我的目标。如果你有什么想法，可以通过 twitter 或者在本文下方评论来联系我。如果我们能够把它做成一个帮助开发者调试和记录 Go 程序运行情况的工具的话就更好了。

如果你想用我的工具看一些算法的动画效果，可以在下方留言，我很乐意提供帮助。

Happy coding！

更新： 文中提到的工具可以在[这里](https://github.com/divan/gotrace)下载。目前它使用了 Go Execution Tracer 和打包的 runtime 来生成 trace。

---

via: https://divan.github.io/posts/go_concurrency_visualize/

作者：[Ivan Daniluk](https://github.com/divan)
译者：[QueShengyao](https://github.com/QueShengyao)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
