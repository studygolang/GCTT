# Go语言并发可视化

本文作者提供了其在2016的GopherCon上的关于Go并发可视化的[主题演讲](https://www.youtube.com/watch?v=KyuFeiG3Y60)。

Go语言一个鲜明的优点就是内置的基于[CSP](https://en.wikipedia.org/wiki/Communicating_sequential_processes)的并发实现。Go可以说是一个为了并发而设计的语言，允许我们使用它构建复杂的并发流水线。但是开发者是否在脑海中想象过不同的并发模式呢，它们在你的大脑中是怎样的形状？

你肯定想过这些！我们都会靠多种多样的想象来思考。如果我让你想象一下1-100的数字，你会下意识地在脑海中闪过这些数字的图像。比如说我会把它想象成一条从我出发的直线，到20时它会右转90度并且一直延伸到1000。这是因为我上幼儿园的时候，卫生间墙上写满了数字，20正好在角落上。你们脑中肯定有自己想象的数字形象。另一个常见的例子就是一年四季的可视化，有些人把它想象成一个盒子，另外一些人把它想象成是圆圈。

无论如何，我想要用Go和WebGL分享我想象的一些常用并行模式的可视化。这些可视化或多或少地代表了我头脑中的并行编程方法。如果能知道我们之间对并行可视化想象的差异，肯定是一件很有趣的事情。我尤其想要知道Rob Pike和Sameer Ajmani是怎么想象并行的，那一定很有意思。

现在，我们就从最简单的“Hello, Concurrent World”开始，来了解我脑海中的并行世界吧。

## Hello, Concurrent World

这个例子的代码很简单，只包含一个channel，一个goroutine，一个读操作和一个写操作。
```
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

![hello](http://divan.github.io/demos/gifs/hello.gif)

在这张图中，蓝色的线代表goroutines的时间轴。连接`main`和`go#19`的蓝线是用来标记goroutine的起始和终止并且表示父子关系的。红色的箭头代表的是send/recv操作。尽管send/recv操作是两个独立的操作，但是我试着将它们表示成一个操作`从A发送到B`。右边蓝线上的`#19`是该goroutine的内部ID，可以通过Scott Mansfield在[Goroutine IDs](http://blog.sgmansfield.com/2015/12/goroutine-ids/)一文中提到的技巧获取。

## Timers
事实上，我们可以通过简单的几个步骤编写一个计时器：创建一个channel，启动一个goroutine以给定间隔往channel中写数据，将这个chennel返回给调用者。调用者阻塞地从channel中读，就会得到一个精准的时钟。让我们来试试调用这个程序24次并且将过程可视化。
```
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

![timers](http://divan.github.io/demos/gifs/timers.gif)

这个效果是不是很有条理？

## Ping-pong
这个例子是我从Google员工Sameer Ajmani的一次演讲["Advanced Go Concurrency Patterns"](https://talks.golang.org/2013/advconc.slide#1)中找到的。当然，这并不是一个很高阶的并发模型，但是对于Go语言并发的新手来说是很有趣的。

在这个例子中，我们定义了一个channel来作为“乒乓桌”。乒乓球是一个整形变量，代码中有两个goroutine“玩家”通过增加乒乓球的counter在“打球”。
```
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

![Ping-pong](http://divan.github.io/demos/gifs/pingpong.gif)

我建议你点击“WebGL 动画界面”链接，从不同角度看看这个模型，并且试试它减速，加速的效果。

现在，我们给这个模型添加一个玩家（goroutine）。
```
 go player(table)
 go player(table)
 go player(table)
```
[WebGL 动画界面](http://divan.github.io/demos/pingpong3/)

![Ping-pong2](http://divan.github.io/demos/gifs/pingpong3.gif)

我们可以看到每个goroutine都有序地“打到球”，你可能会好奇这个行为的原因。那么，为什么这三个goroutine始终按照一定顺序接收到ball呢？

答案很简单，Go运行时会对每个channel的所有接收者维护一个FIFO队列。在我们的例子中，每个goroutine会在它将b让我们来all传给channel的之后就开始等待channel，所以它们在队列里的顺序总是一定的。让我们增加goroutine的数量，看看顺序是否仍然保持一致。
```
for i := 0; i < 100; i++ {
    go player(table)
}
```
[WebGL 动画界面](http://divan.github.io/demos/pingpong100/)

![Ping-pong100](http://divan.github.io/demos/gifs/pingpong100.gif)

很明显，它们的顺序仍然是一定的。我们可以创建一百万个goroutine去尝试，但是上面的实验已经足够让我们得出结论了。接下来，让我们来看看一些不一样的东西，比如说通用的消息模型。

## Fan-In
扇入（fan-in）模式在并发世界中广泛使用。扇出（fan-out）与其相反，我们会在下面介绍。简单来说，扇入模式就是一个函数从多个输入源读取数据并且复用到单个channel中。比如说：
```
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

![Fan-In](http://divan.github.io/demos/gifs/fanin.gif)

我们能看到，第一个`producer`每隔一百毫秒生成一个值，第二个`producer`每隔250毫秒生成一个值，但是`reader`会立即接收它们的值。main函数中的for循环高效地接收了channel发送的所有信息。

## Workers
与扇入模式相反的模式叫做扇出（fan-out）或者工作者（workers）模式。多个goroutine可以从相同的channel中读数据，利用多核并发完成自身的工作，这就是工作者（workers）的由来。在Go中，这个模式很容易实现，只需要启动多个以channel作为参数的goroutine，主函数传数据给这个channel，数据分发和复用会由Go运行环境自动完成。
```
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

![Workers](http://divan.github.io/demos/gifs/workers.gif)

在这里需要提一下并行结构（parallelism）。我们可以看到，动图中所有的goroutine都是平行“延伸”，等待channel给它们发数据来运行的。我们还可以注意到两个goroutine接收数据之间几乎是没有停顿的。不巧的是，这个动画并没有用颜色区分一个goroutine是在等数据还是在执行工作，这个动画是在 `GOMAXPROCS=4` 的情况下录制的，所以只有4个goroutine能够同时运行。我们将会在下文汇总讨论这个主题。

现在，我们来写更复杂一点的代码，启动带有子工作者的工作者（subworkers）：
```
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

![Workers](http://divan.github.io/demos/gifs/workers2.gif)

当然，我们可以将工作者的数量或子工作者的数量设得更高，但是在这里我们试着不让动画效果变得太复杂。

Go中还存咋你这更酷的扇出模式，比如动态工作者/子工作者数量，使用channel来传输channel，但是现在的动画模拟应该已经可以解释扇出模型的含义了。

## Servers