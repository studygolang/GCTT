# Golang: 竞态条件的调查和检测

## Go中的并发

Go提供了绝对会令人称奇的并发原语，真正地实现了将并发当作一等公民。
不幸的是，确保并发的正确性需要结合相当多的不同的技术，以便于最小化产生并发相关错误的可能性。
这些技术中大多数都不是自动产生的（由编译器完成），并且很大程度上依赖于开发人员的经验。
我发现，如果没有在某些时候遇到过这些常见的并发相关的问题，通常来说是很难对问题进行推理的。
对于我参与过的很多组织而言，这可能会导致经验不足的工程师更有可能引入并发相关的错误。

这篇文章将会介绍竞态条件，并且介绍为什么并发编程这么困难。然后，本文将会介绍如何调查、检测和预防Go中竞态条件，
以及相应的解决方案。最后是动手实践部分，我发现它（实践所用的技术）可以有效得识别竞态条件，
我称之为***候选人和上下文***。（本文所有的示例代码都可以在[grokking-go github repo](https://github.com/dm03514/grokking-go/pull/3/files)内找到。

## 并发的危险之处

并发是指同时有两个操作在执行。这与同步执行形成对比，在同步执行中，程序逐步地执行，
除了当前操作正在执行指令之外，运行中不会发生其他任何事情。并发操作具有不确定性，并且不可预测，很难进行推理。

并发编程的一个最大的挑战在于，一个工作单元可以被抢占，从而创建大量潜在的（不确定的）执行顺序。
这种不确定性意味着*共享*内存可以由不同的工作线程以意想不到的方式来使用。
不受保护的、显式地保护访问安全的*共享*内存可能会导致竞态条件。这就会产生*不安全的*的代码。
相比之下，[*线程安全*](https://en.wikipedia.org/wiki/Thread_safety)的代码是在并发环境下能安全使用的正确的代码。

由于并发具有不[确定性](https://en.wikipedia.org/wiki/Deterministic_algorithm)，因此可能会导致一些极其隐蔽的错误。
一些代码看起来可能没有问题，且测试也能通过，但是在高并发场景下，或者在某些特定的执行路径下，它可能会产生一些微妙的破坏行为。

这篇文章旨在解释一种并发错误： ***竞态条件*** 。

### [竞态条件](https://en.wikipedia.org/wiki/Race_condition#Software)

当两个线程同时访问某块内存，且其中一个线程是写入操作，我们即能称之为竞态条件。竞态条件产生的根本原因就是不同步的访问*共享*内存。

#### 显式的非同步内存访问

下面的[代码](https://github.com/dm03514/grokking-go/blob/83cf3d8313a7c797c317d7b5b2d85b4df89a0401/candidates-and-contexts/races/explict_test.go#L22)描述了有竞态条件的*HTTP*处理器。

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

当负载([源码](https://github.com/dm03514/grokking-go/blob/83cf3d8313a7c797c317d7b5b2d85b4df89a0401/candidates-and-contexts/races/explict_test.go#L22))增加时，上述代码中的竞态条件就会出现。
下面的测试用例使用了200个Goroutines模拟了200个请求。我们期待的是每个请求进来之后，计数器都会递增，但最终结果却不是这样：

```bash
$ go test -run TestExplicitRace ./races/ -v -total-requests=200 -concurrent-requests=200
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

交错执行的步骤产生了一个错误的结果。

![pic1](https://cdn-images-1.medium.com/max/800/1*Bkpqr91IgLk7HZZ_LEOq0Q.png)

上图说明了问题的原因。我们需要自顶向下地看这张图。计数器是从零开始的。多个*http*处理器的Goroutine会通过`Value()`方法读取当前计数器的值，并保存到自己的Goroutine栈中：

```bash
value := reqCount.Value()
```

然后这些Goroutines开始一些模拟工作：

```bash
time.Sleep(1 * time.Nanosecond)
```

当它们的工作完成之后，它们就会增加计数器的值：

```bash
reqCount.Set(value + 1)
```

这里的问题就在于多个Goroutines在修改同一个值！在上图中，有两个Goroutine都写入了值1，实际上应该是1+1。

在显式竞争的情况下，我们在同一时间内对同一块内存分别执行读取和写入的操作：

![pic2](https://cdn-images-1.medium.com/max/800/1*pmHOKo0pHAYe6tyuJO_Hig.png)

这就导致了一个[未定义的行为](https://en.wikipedia.org/wiki/Undefined_behavior)。 最糟糕的是，即使已经存在竞态条件，但也不会有明确的错误产生。程序仅仅是不正确而已。
因为并发是极其困难的，我工作中所接触的代码库中多多少少都会有这样的错误... 任何地方都有。

#### 逻辑上的竞态条件

即使以上述例子的方式访问``reqCount Counter`是线程安全的，但仍然存在逻辑上的竞态条件问题。我们使用完全同步的线程安全计数器（稍后介绍）执行下面的测试用例（[源码](https://github.com/dm03514/grokking-go/pull/3/files#diff-a507be0a589eb624edffd8260bba4bfdR14))，结果仍然不正确：

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
FAIL    github.com/dm03514/grokking-go/candidates-and-contexts/races    0.123s
```

虽然我们不再同时进行读写操作，应用程序有可能正在处理旧的数据（和上面的显式竞态条件相同）。并发地调用`Value()`可能会返回相同的值，并导致并发的线程设置相同相同的值，这在逻辑上是不正确的。即每个*http*处理器调用程序应该执行`+1`操作并存入计数器，计数器的操作应该可以跨线程序列化。

逻辑竞态条件是另一种完全不同的问题，本文将只关注显式的竞态条件用例。

## 解决方案
