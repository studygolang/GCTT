首发于：https://studygolang.com/articles/17683

# Golang: 竞态条件的调查和检测

## Go 中的并发

Go 提供了绝对会令人称奇的并发原语，真正地实现了将并发当作一等公民 不幸的是，确保并发的正确性需要结合很多不同的技术，以便于最小化发生并发相关错误的可能性。这些技术中大多数都不是自动产生的（由编译器完成的），并且很大程度上依赖于开发人员的经验。同时，我还发现，如果没有在某些时候遇到过这些常见的并发问题，通常来说是很难对这些问题进行推理的。对于我参与过的很多组织而言，这可能会导致经验不足的工程师更有可能引入并发相关的错误。

这篇文章将会介绍竞态条件，并且介绍为什么并发编程会这么困难。然后，本文将会介绍如何调查、检测和预防 Go 中竞态条件，
以及相应的解决方案。最后是动手实践部分，我发现这种方法可以有效的识别竞态条件， 我称之为 ***候选对象和上下文***。（本文所有的示例代码都可以在[grokking-go GitHub repo](https://github.com/dm03514/grokking-go/pull/3/files) 内找到。)

## 并发的危险之处

并发指的是两个操作在同时执行。这与同步执行形成对比，在同步执行中，程序逐步地执行，除了当前操作正在执行指令之外，运行中不会发生其他任何事情。但并发操作具有不确定性，并且不可预测，很难进行推理。

并发编程的一个最大的挑战在于，一个工作单元可以被抢占，从而创建大量潜在的（不确定的）执行顺序。这种不确定性意味着*共享*内存可以由不同的工作线程以意想不到的方式来使用。未受保护且明确安全访问的*共享*内存可能导致竞争条件。这就会产生*不安全的*的代码。相比之下，[*线程安全*](https://en.wikipedia.org/wiki/Thread_safety) 的代码是在并发环境下能安全使用的正确的代码。

由于并发具有不[确定性](https://en.wikipedia.org/wiki/Deterministic_algorithm)，因此可能会导致一些极其隐蔽的错误。一些代码看起来可能没有问题，而且测试也能通过，但是在高并发场景下，或者在某些特定的执行路径下，它可能会产生一些微妙的破坏行为。

这篇文章旨在解释一种并发错误： ***竞态条件*** 。

### [竞态条件](https://en.wikipedia.org/wiki/Race_condition#Software)

当两个线程同时访问某块内存，且其中一个线程是写入操作，我们即能称之为竞态条件。竞态条件产生的根本原因就是不同步的访问*共享*内存。

#### 显式的非同步内存访问

下面的[代码](https://github.com/dm03514/grokking-go/blob/83cf3d8313a7c797c317d7b5b2d85b4df89a0401/candidates-and-contexts/races/explict_test.go#L22) 描述了一个有竞态条件的*HTTP*处理程序。

```go
reqCount := Counter{}

http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  value := reqCount.Value()
  fmt.Printf("handling request: %d\n", value)
  time.Sleep(1 * time.Nanosecond)
  reqCount.Set(value + 1)
  fmt.Fprintln(w, "Hello, client")
}))
log.Fatal(http.ListenAndServe(":8080", nil))
```

当负载 ([源码](https://github.com/dm03514/grokking-go/blob/83cf3d8313a7c797c317d7b5b2d85b4df89a0401/candidates-and-contexts/races/explict_test.go#L22)) 增加时，上述代码中的竞态条件就会出现。
下面的测试用例使用了 200 个 Goroutines 模拟了 200 个请求。我们期待的是每当一个请求进来时，计数器都会递增，但最终结果却不是这样：

```bash
$ Go test -run TestExplicitRace ./races/ -v -total-requests=200 -concurrent-requests=200
...
handling request: 9
handling request: 9
handling request: 9
handling request: 10
handling request: 11
handling request: 12
handling request: 13
handling request: 14
handling request: 15
handling request: 16
handling request: 17
handling request: 18
handling request: 19
handling request: 20
handling request: 21
handling request: 22
handling request: 23
handling request: 24
handling request: 25
handling request: 26
handling request: 26
handling request: 26
handling request: 26
Num Requests TO Make: 200
Final Count: 27
--- FAIL: TestExplicitRace (0.08s)
        explict_test.go:72: expected 200 requests: received 27
FAIL
FAIL    github.com/dm03514/grokking-go/candidates-and-contexts/races    0.083s
```

如图所示，交错执行的步骤产生了一个错误的结果。

![pic1](https://cdn-images-1.medium.com/max/800/1*Bkpqr91IgLk7HZZ_LEOq0Q.png)

上图说明了问题的原因。我们需要自顶向下地看这张图。计数器是从零开始的。多个 *http* 处理程序的 Goroutine 会通过 `Value()` 方法读取当前计数器的值，并保存到自己的 Goroutine 栈中：

```bash
value := reqCount.Value()
```

然后这些 Goroutines 开始一些模拟工作：

```go
time.Sleep(1 * time.Nanosecond)
```

当它们的工作完成之后，它们就会增加计数器的值：

```bash
reqCount.Set(value + 1)
```

这里的问题就在于多个 Goroutines 在修改同一个值！在上图中，有两个 Goroutine 都写入了值 1，但实际上应该是 1+1。

在显式竞争的情况下，我们在同一时间内对同一块内存分别执行读取和写入的操作：

![pic2](https://cdn-images-1.medium.com/max/800/1*pmHOKo0pHAYe6tyuJO_Hig.png)

这就导致了一个[未定义的行为](https://en.wikipedia.org/wiki/Undefined_behavior)。 最糟糕的是，即使已经存在竞态条件，但也不会有明确的错误产生。程序仅仅是不正确而已。 因为并发是极其困难的，我工作中所接触的代码库中多多少少都会有这样的错误 ... 任何地方都有。

#### 逻辑上的竞态条件

即使以上述例子的方式访问 `reqCount Counter` 是线程安全的，但仍然存在逻辑上的竞态条件问题。我们使用完全同步的线程安全计数器（稍后介绍）执行下面的测试用例（[源码](https://github.com/dm03514/grokking-go/pull/3/files#diff-a507be0a589eb624edffd8260bba4bfdR14))，结果仍然不正确：

```bash
$ go test -run TestLogicalRace ./races/ -v -total-requests=200 -concurrent-requests=200
...
handling request: 25
handling request: 25
handling request: 25
handling request: 25
handling request: 25
handling request: 25
handling request: 25
handling request: 25
handling request: 25
handling request: 25
handling request: 20
handling request: 25
handling request: 26
handling request: 27
Num Requests TO Make: 200
Final Count: 26
--- FAIL: TestLogicalRace (0.12s)
        logical_test.go:67: expected 200 requests: received 26
FAIL
FAIL    GitHub.com/dm03514/grokking-go/candidates-and-contexts/races    0.123s
```

虽然我们不再同时进行读写操作，应用程序有可能正在处理旧的数据（和上面的显式竞态条件相同）。并发地调用 `Value()` 可能会返回相同的值，并导致并发的线程设置相同的值，这在逻辑上是不正确的。即每个*http*处理程序调用程序应该执行 `+1` 操作并存入计数器，计数器的操作应该可以跨线程序列化。

逻辑竞态条件是另一种完全不同的问题，本文将只关注显式的竞态条件用例。

## 解决方案

目前为止，我们已经了解到什么是竞态条件，下面是一些可用的工具，用于在工程中减轻它们 :

### 竞态检测器

Go 的竞态检测器支持对内存的访问进行检测，以确定内存是否发生了并发的操作。通过在 `go test` 工具上添加 `-race` 标志位来启用竞态检测器。甚至在官方的竞态检测器文档的第一行也写道： *[竞态条件](http://en.wikipedia.org/wiki/Race_condition) 是最隐蔽、最难以捉摸的编程错误之一。*

虽然竞态检测器是非常有用的工具，但是它是被动触发的，因为竞态条件一定是发生在运行时的。这就给工程师带来了负担，因为他们需要确定哪些 Goroutine 中可以检测出竞态条件，编写并发测试用例，然后在启用了竞态检测器的情况下执行测试。即使我们构建了一个具有高并发的测试用例，但是仍然不能 100% 保证测试用例的调用能够导致读写重叠。虽然有这种情况，但是我们也不能忽视竞态检测器的益处，因为它能够在大多数情况下检测到竞态条件，是一个非常强大的工具。

不幸的是，竞态检测器仅仅用于*检测*，而不能进行预防。

让我们针对第一个示例使用竞态检测器来看看它是如何工作的。下面我们使用 `-race` 标志位执行显式的竞态条件的检测（[源代码](https://github.com/dm03514/grokking-go/blob/83cf3d8313a7c797c317d7b5b2d85b4df89a0401/candidates-and-contexts/races/explict_test.go#L22)）

```bash
$go test -run TestExplicitRace ./races/ -v -total-requests=200 -concurrent-requests=200 -race
=== RUN   TestExplicitRace
handling request: 0
handling request: 0
handling request: 0
handling request: 0
==================
WARNING: DATA RACE
Write at 0x00c4200164e8 by Goroutine 326:
  GitHub.com/dm03514/grokking-go/candidates-and-contexts/races.TestExplicitRace.func1.1()
      /vagrant_data/go/src/github.com/dm03514/grokking-go/candidates-and-contexts/races/counters.go:18 +0x115
  net/http.HndlerFunc.ServeHTTP()
      /usr/local/go/src/net/http/server.go:1947 +0x51
  net/http.(*ServeMux).ServeHTTP()
      /usr/local/go/src/net/http/server.go:2340 +0x9f
  net/http.serverHandler.ServeHTTP()
      /usr/local/go/src/net/http/server.go:2697 +0xb9
  net/http.(*conn).serve()
      /usr/local/go/src/net/http/server.go:1830 +0x7dc
Previous read at 0x00c4200164e8 by Goroutine 426:
  GitHub.com/dm03514/grokking-go/candidates-and-contexts/races.TestExplicitRace.func1.1()
      /vagrant_data/go/src/github.com/dm03514/grokking-go/candidates-and-contexts/races/counters.go:14 +0x5b
  net/http.HandlerFunc.ServeHTTP()
      /usr/local/go/src/net/http/server.go:1947 +0x51
  net/http.(*ServeMux).ServeHTTP()
      /usr/local/go/src/net/http/server.go:2340 +0x9f
  net/http.serverHandler.ServeHTTP()
      /usr/local/go/src/net/http/server.go:2697 +0xb9
  net/http.(*conn).serve()
      /usr/local/go/src/net/http/server.go:1830 +0x7dc
Goroutine 326 (running) created at:
  net/http.(*Server).Serve()
      /usr/local/go/src/net/http/server.go:2798 +0x364
  net/http.(*Server).ListenAndServe()
      /usr/local/go/src/net/http/server.go:2714 +0xc4
  net/http.ListenAndServe()
      /usr/local/go/src/net/http/server.go:2972 +0xf6
  GitHub.com/dm03514/grokking-go/candidates-and-contexts/races.TestExplicitRace.func1()
      /vagrant_data/go/src/github.com/dm03514/grokking-go/candidates-and-contexts/races/explict_test.go:36 +0xd9
Goroutine 426 (running) created at:
  net/http.(*Server).Serve()
      /usr/local/go/src/net/http/server.go:2798 +0x364
  net/http.(*Server).ListenAndServe()
      /usr/local/go/src/net/http/server.go:2714 +0xc4
  net/http.ListenAndServe()
      /usr/local/go/src/net/http/server.go:2972 +0xf6
  GitHub.com/dm03514/grokking-go/candidates-and-contexts/races.TestExplicitRace.func1()
      /vagrant_data/go/src/github.com/dm03514/grokking-go/candidates-and-contexts/races/explict_test.go:36 +0xd9
==================
```

太好了！在测试开始的时候，竞态检测器就识别出了竞态条件，并给予我们提示。

```bash
WARNING: DATA RACE
Write at 0x00c4200164e8 by goroutine 326:
  GitHub.com/dm03514/grokking-go/candidates-and-contexts/races.TestExplicitRace.func1.1()
      /vagrant_data/go/src/github.com/dm03514/grokking-go/candidates-and-contexts/races/counters.go:18 +0x115
...
previous read at 0x00c4200164e8 by goroutine 426:
  GitHub.com/dm03514/grokking-go/candidates-and-contexts/races.TestExplicitRace.func1.1()
      /vagrant_data/go/src/github.com/dm03514/grokking-go/candidates-and-contexts/races/counters.go:14 +0x5b
```

这些日志表明了测试中存在并发的读写：

```go
package races

import "sync"

type Counter struct {
    count int
}

func (c *Counter) Value() int {
    return c.count # line 14
}

func (c *Counter) Set(v int) {
    c.count = v # line 18
}
```

关于竞态检测器的文章有很多。使用竞态检测器的缺点在于如果使用它，我们就必须编写一些重点的测试，这样会带来一些时间上的开销。还有另外一种方法，在启用 `-race` 标志位的情况下，使用一小部分流量来进行金丝雀发布。

### 显式的同步

显式同步是通过[同步原语](https://golang.org/pkg/sync/)（如互斥锁）来保护变量的访问。显式同步让工程师承担了识别并发执行的候选对象和它们将在其中执行的上下文的责任。然后他们还要求工程师知道如何编写互斥量的代码来进行同步访问。

这很棘手，因为变量的增加是安全的，同步的，但同时它又是不安全的。这也就是我在上面提到的，需要依赖工程师的经验。显式的内存同步访问需要预测和标识代码片段将要执行的所有*上下文*。因为我们知道我们的处理程序将会被 `net/http` 库并发地执行，我们可以为计数器增加显式的同步，并向之后的开发人员提供 ***线程安全*** 的保证。

```go
type SynchronizedCounter struct {
   mu *sync.Mutex
   count int
}

func (c *SynchronizedCounter) Inc() {
   c.mu.Lock()
   defer c.mu.Unlock()

   c.count++
}

func (c *SynchronizedCounter) Value() int {
   c.mu.Lock()
   defer c.mu.Unlock()

   return c.count
}

func (c *SynchronizedCounter) Set(v int) {
   c.mu.Lock()
   defer c.mu.Unlock()

   c.count = v
}
```

现在，我们可以保证我们的方法是 ***线程安全*** 的；所有状态的变化都受互斥量的保护并且是[可序列化的](https://aphyr.com/posts/313-strong-consistency-models)。

### 静态分析 （通过 `go vet`)

静态分析（特别是互斥量检测）有助于检测互斥量的滥用，是另一种提供给我们的被动检测工具。它并不能帮助我们直接检测一变量是否需要互斥锁，它只能检测出互斥锁是否被正确地使用。它要求工程师能够识别哪里需要互斥量，并在恰当的位置使用它，通常来说我们应该使用互斥量的副本，而非互斥量的引用以避免产生误用。

```go
type MisSynchronizedCounter struct {
   mu    sync.Mutex
   count int
}

func (c MisSynchronizedCounter) Inc() {
   c.mu.Lock()
   defer c.mu.Unlock()

   c.count++
}
```

`vet` 能够检测出当前使用的互斥量是拷贝还是引用：

```bash
$ Go vet -copylocks ./races/counters.go
# command-line-arguments
races/counters.go:52: Inc passes lock by value: races.MisSynchronizedCounter contains sync.Mutex
```

`vet` 是一个重要的工具，它可以添加到如何构建的过程当中。但是还不足以识别或者检测竞态条件。

### 基于设计的哲学

基于正确性的设计利用了 Go 的安全原语和设计模式的最佳实践，来最小化出现竞态条件的可能性。它有两种常见的模式：

- “监控” Goroutine

- 工作池

上面两者都用到 Go 的通道特性。基于设计的方法体现了 Go 的箴言：“[通过通道来共享内存](https://blog.golang.org/share-memory-by-communicating) ”。下面展示了在重构计数器并将其封装在监控 Goroutine 之后会发生什么（[源代码](https://github.com/dm03514/grokking-go/pull/3/files#diff-8222898e088ed9fee100b22308c43685R22)）:

```go
countChan := make(chan struct{})
go func() {
   for range countChan {
      reqCount.Inc()
      fmt.Printf("handling request: %d\n", reqCount.Value())
   }
}()
```

这非常棒，因为这是一种混合方法： 它运行调度并发的操作，但是并发操作是唯一需要访问 `reqCount` 的东西。这就意味着 `reqCount` 不再需要同步。（除了在 `countChan` 关闭后主测试线程因为断言需要访问它之外。正如我们看到的那样，程序的行为与预期的一致，并没有产生任何的竞态条件 （[源码](https://github.com/dm03514/grokking-go/pull/3/files#diff-8222898e088ed9fee100b22308c43685R14)））。

```bash
$ go test -run TestDesignNoRace ./races/ -v -total-requests=200 -concurrent-requests=200 -race
handling request: 1
...
handling request: 190
handling request: 191
handling request: 192
handling request: 193
handling request: 194
handling request: 195
handling request: 196
handling request: 197
handling request: 198
handling request: 199
handling request: 200
Num Requests TO Make: 200
Final Count: 200
--- PASS: TestDesignNoRace (0.64s)
PASS
ok      github.com/dm03514/grokking-go/candidates-and-contexts/races    1.691s
```

这是一个非常强大的模式，并且可以扩展为工作池模式。想象一下，我们不是计数而是将数据写入到数据库。我们可以派生出许多的 Goroutines 来限制某一次可能会发生的最大的并发插入数量（假设是 10 次），并使得它们共享同一个通道。每个 Goroutine 都可以拥有自己的资源，不会与其他的 Goroutine 共享。这就允许了每个单独的 Goroutine 拥有它自己的小宇宙，而不需要同步。每个 Goroutine 都有自己的同步空间，但是也需要通过外部的调度来并发地运行。

### 分析 I

在并发的 ***上下文*** 中使用第三方或者外部组件时，需要进行分析。Go 约定的是假定所有东西都是线程不安全的，除非提供显式的保证。

这类分析常见的候选对象是：

- DB 连接
- TCP 连接
- 第三方 SDK/ 驱动
- 在 goroutines 之间共享的任何第三方组件

设想一下，我们的应用需要向数据库写入数据。我们初始化了 db，然后希望有一个工作池，它通过 `db.QueryContext` 并发地读写 db。那么传递 db 实例是否安全？

第一步我们应该查阅相关[文档](https://golang.org/pkg/database/sql/#DB)，在本例中，它是这样说的：

> DB 是表示零个或多个底层连接池的数据库句柄文件。多个 Goroutines 可以并发的使用。

我个人认为这是足够强大的保证，因为到目前为止，它还没坑到我。对于大型核心的项目（cassandra, Go-aws-sdk, Go 标准库）而言，通常都有关于线程是否安全的文档解释。

对于较小的项目，有些时候他们不发布线程安全保证，所以需要审阅代码以确保并发的正确性。

### 无并发

如果可以避免并发，那么就可以完全消除这类极其难定位的错误。这在很多的运行时已经被验证过，以*Python*或*Ruby*为例，对于标准的*web*服务器采用*pre fork*模型进行部署，初始化一组*Python*/*Ruby*进程。每个进程接收一个能够同步执行的连接，然后返回结果。并发是通过使用反向代理（如*uwsgi*或*nginx*）在外部实现的。这就允许我们进行非常简单地本地开发，并将并发性转移到外部进程。

但这可能仅适用于某些特别类型的问题，因为我们使用 Go 的原因之一就是它的性能足够好。因为它的运行时如此优秀，在标准的 4 核机器上*http*服务器上可以有效地处理大量的负载和活动的连接。

虽然这似乎是在倒退，但如果排除并发是完全可能的，那么就有能令人信服的理由这么做。这种模式和上面概述的设计模式密切相关。每个独立的工作者的功能都有自己的空间，并且不知道它将会被如何调度。因为它通过通道接收输入（并发原语）并且不与任何其他工作者共享内容。所以它是并发安全的。

综上所述，其实没有客观有效的解决方案。理想的解决方案是，编译器能够神奇地检测并告知代码何时何地并发执行，并且以高度正确性的方式发出问题的警告。

由于编译器不可能为我们做这些工作，下面概述了我经常用来帮助解决这个问题的策略：

## 候选对象和上下文 (C&C)

这是一个应用程序级别的手动分析方法，我一直在努力尝试填补因编译器缺乏主动检查而给我们留下的一些空白。 它基于 ***候选对象*** 定位语句。有一些可能会导致并发错误的高风险语句。然后，它检查是否在并发的 ***上下文*** 中执行了这些语句。*C&C*是关于识别的，因为它是检测的前置条件。我发现它是一个很好的工具，可以确定哪些 Goroutines 应该通过 Go 的竞态检测器来测试并发性。

竞态条件需要下面两件事情同时发生才能触发：

- 共享内存（***候选对象***）
- 并发访问（***上下文***）

我们可以以下图中的方法来看待它们的关系：

![pic3](https://cdn-images-1.medium.com/max/800/1*0Twwyn8-0yvm5TB6tbXCEg.png)

这里的观点是，如果内存*没有被共享*或 / 和*没有被同时执行*，那么它就*不可能*出现竞态条件。就像内存是共享的（例如标准的全局变量），但即使有同步执行，也不会出现竞态条件（尽管全局变量有很多设计和可理解性的问题）。另外，如果有并发执行但是没有共享状态，那么也不可能产生竞态条件（这就和上述的监控模式，或*Python*/*Ruby*模式一样）。当存在共享可变的状态和该状态上有并发的操作时，才有可能出现竞态条件。候选对象和上下文 (C&C) 模型是基于识别并发操作和共享状态实现的。

候选对象本身不代表竞态条件，因为单纯的使用 Goroutines 并不代表有竞态条件，但若有被同时执行的候选对象，那么就可能存在竞态条件。

![pic4](https://cdn-images-1.medium.com/max/800/1*GYL_tEBKZi6BPRdILXhAQw.png)

***C&C 的目的是确定有高风险产生并发错误的代码区域***。

### 候选对象

*C&C*的第一步是识别候选对象。这都是一些具有高风险构成竞态条件的东西。与全局变量或在多个函数调用之间传递的指针相比，在栈帧内分配而不在栈帧外共享的变量其风险要小的多。候选对象的检查是独立于并发的。

所以我寻找的候选对象是：

- 全局变量

- 指针接收者

### 上下文

下一步就是定位上下文。上下文就是代码会被并发执行的区域。即在[Goroutine](https://tour.golang.org/concurrency/1) 中。 这里的问题在于很多的库会在 Goroutines 之上进行一些封装抽象，所以在代码库中搜索所有的 Goroutines（即通过 `grep "go "`) 可能会涉及到很深的依赖，并需要深入的分析。

因此，***上下文*** 标识了存在于 Goroutines 之上的常见的应用程序级别的抽象。就拿*HTTP*举例来说。对于这种方法，它的 ***上下文*** 就是一个处理函数，但是其根本仍然只是一个 Goroutine，因为 Go 的网络库会为每个接收的连接启用一个 Goroutine。

常见的上下文是：

- 显式的 Goroutines `grep -rnlF 'go ' .`
- *HTTP*服务器
- 其他的应用程序，比如工作池

### 分析 II

*C&C*下一步就是检查候选对象和上下文重叠的部分。这将识别到上面矩阵中的共享内存 / 并发执行象限中的代码 ( 当一个候选程序在上下文中执行时 )，并且应该标记此处代码以便进行后续的手动分析。

### 示例

```go
type Counter struct {
   count int
}

func (c *Counter) Value() int {
   return c.count
}

func (c *Counter) Set(v int) {
   c.count = v
}
reqCount := Counter{}

http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  value := reqCount.Value()
  fmt.Printf("handling request: %d\n", value)
  time.Sleep(1 * time.Nanosecond)
  reqCount.Set(value + 1)
  fmt.Fprintln(w, "Hello, client")
}))
log.Fatal(http.ListenAndServe(":8080", nil))
```

#### 候选对象 ( 示例 )

请记住，我们第一步应该是先定位候选对象。这需要查看所有的全局变量或者指针接收者。在这种情况下，我们发现了：

- `reqCount` 是一个全局变量
- `reqCount.Value()` 是一个指针接收者
- `reqCount.Set()` 是一个指针接收者

#### 上下文 ( 示例 )

接下来，我们要定位正在并发执行的代码，在此示例中，只有 `http.HandlerFunc`。

#### 重叠部分 ( 示例 )

最后，我们检查是否在并发上下文中执行了任何候选对象。所有的候选对象都需要检查。 这一步将把每一个候选对象标记为潜在的竞态条件候选者，这样就可以进行更深入的分析，以确保对所标识的函数 / 变量的访问是同步的。

## 结论

即使我们已经有了 Go 提供的并发问题解决方案，但我仍然发现，若想要确保 Go 中并发地运行正确的程序，仍然需要进行安全细致的分析工作。我觉得这反映了并发编程的状态还不够成熟。我个人非常希望看到 Go 能够提供更好用的竞态条件分析和检测工具。Go 已经成为我个人比较喜爱的工具，并且在性能、并发性、简单性和速度之间取得了平衡，这是我使用过的任何工具所不能及的。但是，如果这样的特性被内置并被 Go 的编译器支持时，想想未来，这绝对是一件令人兴奋的事情 !

---

via: https://medium.com/dm03514-tech-blog/golang-candidates-and-contexts-a-heuristic-approach-to-race-condition-detection-e2b230e70d08

作者：[dm03514](https://medium.com/@dm03514)
译者：[barryz](https://github.com/barryz)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
