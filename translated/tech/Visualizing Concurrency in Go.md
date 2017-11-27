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