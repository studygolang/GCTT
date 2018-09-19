首发于：https://studygolang.com/articles/14892

# Go 语言中的同步队列

## 问题

假设我们在运营一家 IT 公司，公司里面有程序员和测试员。为了给个机会他们互相认识对方，并且让他们能够在工作中放松一点，我们买了一个乒乓球台，并且制定了如下规则：

- 每次只能两个人（不能少于或多于两人）玩。
- 只有上一对玩家结束了，下一对玩家才能玩，也就是说，不能只换下一个人。
- 只能是测试员和程序员组成一对来玩，（不能出现两个测试员或者两个程序员一起玩的情况）。如果员工想要玩的话，那么他得等到有合适的对手了才能开始游戏。

```go
func main() {
    for i := 0; i < 10; i++ {
        go programmer()
    }
    for i := 0; i < 5; i++ {
        go tester()
    }
    select {} // 漫长的工作日...
}
func programmer() {
    for {
        code()
        fmt.Println("Programmer starts")
        pingPong()
        fmt.Println("Programmer ends")
    }
}
func tester() {
    for {
        test()
        fmt.Println("Tester starts")
        pingPong()
        fmt.Println("Tester ends")
    }
}
```

我们用 `time.sleep` 来模拟测试、开发、和玩乒乓球的行为。

```go
func test() {
    work()
}
func code() {
    work()
}
func work() {
    // Sleep up to 10 seconds.
    time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)
}
func pingPong() {
    // Sleep up to 2 seconds.
    time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
}
```

这个程序的输出类似这样：

```bash
> go run pingpong.go
Tester starts
Programmer starts
Programmer starts
Tester ends
Programmer ends
Programmer starts
Programmer ends
Programmer ends
```

但是如果我们要按照我们制定的规矩去玩乒乓球的话，那输出只能是下面四种情况：

```
Tester starts
Programmer starts
Tester ends
Programmer ends

Tester starts
Programmer starts
Programmer ends
Tester ends

Programmer starts
Tester starts
Tester ends
Programmer ends

Programmer starts
Tester starts
Programmer ends
Tester ends
```

程序员或者测试员先走到乒乓球桌上，然后等待他的合法对手加入。当他们打完离开时，他们离开的顺序是任意的。所以只有上述四种输出序列是有效的。

下面有两种解决方案，第一种是基于 mutex （互斥量）的，而第二种使用了不同的 worker ，它们协调整个处理的过程，确保所有事情都能按照规则来执行。

## 解决方案 #1

两种解决方案都使用了同一种数据结构（`queue.Queue`），来使得程序员和测试员在走上乒乓球桌之前先排好队。当至少有一对玩家（一个程序员和一个测试员）准备好之后，这一对玩家才能开始玩乒乓球。

```go
func tester(q *queue.Queue) {
    for {
        test()
        q.StartT()
        fmt.Println("Tester starts")
        pingPong()
        fmt.Println("Tester ends")
        q.EndT()
    }
}
func programmer(q *queue.Queue) {
    for {
        code()
        q.StartP()
        fmt.Println("Programmer starts")
        pingPong()
        fmt.Println("Programmer ends")
        q.EndP()
    }
}
func main() {
    q := queue.New()
    for i := 0; i < 10; i++ {
        go programmer(q)
    }
    for i := 0; i < 5; i++ {
        go tester(q)
    }
    select {}
}
```

包 `queue` 是这么定义的：

```go
package queue

import "sync"

type Queue struct {
    mut                   sync.Mutex
    numP, numT            int
    queueP, queueT, doneP chan int
}

func New() *Queue {
    q := Queue{
        queueP: make(chan int),
        queueT: make(chan int),
        doneP:  make(chan int),
    }
    return &q
}

func (q *Queue) StartT() {
    q.mut.Lock()
    if q.numP > 0 {
        q.numP -= 1
        q.queueP <- 1
    } else {
        q.numT += 1
        q.mut.Unlock()
        <-q.queueT
    }
}

func (q *Queue) EndT() {
    <-q.doneP
    q.mut.Unlock()
}

func (q *Queue) StartP() {
    q.mut.Lock()
    if q.numT > 0 {
        q.numT -= 1
        q.queueT <- 1
    } else {
        q.numP += 1
        q.mut.Unlock()
        <-q.queueP
    }
}

func (q *Queue) EndP() {
    q.doneP <- 1
}
```

队列里面的 mutex 有两个用途：

- 同步共享变量 `numT` 、`numP` 的访问。
- 作为一个令牌，可以开始游戏的一对玩家才能持有该令牌，其他玩家尝试进入游戏会被阻塞。

程序员和测试员通过非缓冲的 channel `<-q.queueP` 或者 `<-q.queueT` 来等待对手。

从这些 channel 接收数据时，如果此时没有可配对的对手，那么当前的 goroutine 会被阻塞。

我们来分析一下给测试员调用的 `StartT` 函数：

```go
func (q* Queue) StartT() {
    q.mut.Lock()
    if q.numP > 0 {
        q.numP -= 1
        q.queueP <- 1
    } else {
        q.numT += 1
        q.mut.Unlock()
        <-q.queueT
    }
}
```

如果 `numP` 大于 0（表示当前至少有一个程序员在等待加入游戏），那么正在等待中的程序员的数量就会减一，并且有一个正在等待中的程序员批准加入游戏（`q.queueP <- 1`）。有趣的是在这个过程中 mutex 不会被释放掉，这时它的职能就是作为一个允许进入乒乓球桌的令牌。

如果当前没有正在等待的程序员，那么 `numT`（等待中的测试员的数量）将会加一，并且当前的 goroutine 会被阻塞在 `q.queueT`。

`StartP` 函数基本上是一样的，只是它是给程序员调用的。

整个游戏的过程中，mutex 会被锁定，所以它需要被程序员或者测试员释放。要释放 mutex，只能是双方都结束游戏了才行，我们使用了 `doneP` 作为一个屏障：

```go
func (q *Queue) EndT() {
    <-q.doneP
    q.mut.Unlock()
}

func (q *Queue) EndP() {
    q.doneP <- 1
}
```

如果程序员还在游戏，而测试员已经结束游戏了，那么测试员会被阻塞在 `<-q.doneP`。一旦程序员执行到 `q.doneP<-1` 时。这个屏障就会打开，而 mutex 就能得以释放，从而使这些员工可以回去继续工作。

如果测试员还在游戏，而程序员已经结束游戏了，那么程序员会阻塞在 `q.done<-1`，直到测试员结束游戏时，执行 `<-q.doneP` ，从而恢复程序员的运行，并且释放掉 mutex。

这个过程中有趣的是，无论当时是测试员还是程序员把 mutex 锁定的，mutex 永远都是测试员负责释放。这也就是为什么这个解决方案第一看上去没有那么直观。

## 解决方案 #2

```go
package queue
const (
    msgPStart = iota
    msgTStart
    msgPEnd
    msgTEnd
)
type Queue struct {
    waitP, waitT   int
    playP, playT   bool
    queueP, queueT chan int
    msg            chan int
}
func New() *Queue {
    q := Queue{
        msg:    make(chan int),
        queueP: make(chan int),
        queueT: make(chan int),
    }
    go func() {
        for {
            select {
            case n := <-q.msg:
                switch n {
                case msgPStart:
                    q.waitP++
                case msgPEnd:
                    q.playP = false
                case msgTStart:
                    q.waitT++
                case msgTEnd:
                    q.playT = false
                }
                if q.waitP > 0 && q.waitT > 0 && !q.playP && !q.playT {
                    q.playP = true
                    q.playT = true
                    q.waitT--
                    q.waitP--
                    q.queueP <- 1
                    q.queueT <- 1
                }
            }
        }
    }()
    return &q
}
func (q *Queue) StartT() {
    q.msg <- msgTStart
    <-q.queueT
}
func (q *Queue) EndT() {
    q.msg <- msgTEnd
}
func (q *Queue) StartP() {
    q.msg <- msgPStart
    <-q.queueP
}
func (q *Queue) EndP() {
    q.msg <- msgPEnd
}
```

我们会有个专门的中央协调器在一个独立的 goroutine 里面运行，它负责协调整个过程。协调器通过 `msg` channel 获取所有想要玩乒乓球的和刚玩完乒乓球的员工的信息。收到消息时，调度器的状态将会更新：

- 等待中的程序员或者测试员的数量会增加。
- 正在游戏的员工的信息会被更新。

在收到符合定义的消息时，调度器会检查现在是否更够让一对新的选手开始游戏：

```go
if q.waitP > 0 && q.waitT > 0 && !q.playP && !q.playT {
```

如果相应的状态都已经更新了的话，那么一个代表程序员的 goroutine 和一个代表测试员的 goroutine 将会被唤醒。

我们在这个方案中没有使用 mutex，而是使用了一个独立的 goroutine，它通过 channel 与外部世界通讯，这让我们的程序成为一个更”地道“（符合 Go 语言风格）的 Go 语言程序。

> *Don’t communicate by sharing memory, share memory by communicating.*
>
> 不要通过共享内存来通讯，而要通过通讯来共享内存。

## 参考资料

- “The Little Book of Semaphores” by Allen B. Downey（译注：[PDF版地址](http://greenteapress.com/semaphores/LittleBookOfSemaphores.pdf)）
- https://medium.com/golangspec/reusable-barriers-in-golang-156db1f75d0b (译文： https://studygolang.com/articles/12718)
- https://blog.golang.org/share-memory-by-communicating

---

via: https://medium.com/golangspec/synchronization-queues-in-golang-554f8e3a31a4

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[Alex-liutao](https://github.com/Alex-liutao)
校对：[Unknwon](https://github.com/Unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
