首发于：https://studygolang.com/articles/17014

# GO 中的调度：第三部分 - 并发

## 前奏

这篇文章是三部曲系列文章中的第三篇，这个系列的文章将会对 Go 中调度器背后的机制和语义做深入的了解。本文主要关注并发的部分。

Go 调度器系列文章：

- [Go 中的调度器：第一部分 - 操作系统调度器](https://studygolang.com/articles/14264)
- [Go 中的调度器：第二部分 - Go 调度器](https://studygolang.com/articles/15316)
- [Go 中的调度器：第三部分 - 并发](https://studygolang.com/articles/17014)

## 介绍

每当我解决一个问题时，尤其是一个新问题时，我刚开始并不会考虑并发是否合适的问题。我会先寻找一些解决方案来确保它正常工作。然后保证代码的可读性，在技术审阅之后，我才会开始提出并发性是否合理及实用的问题。对于并发性，有时候它是一个好东西，有时候却未必是。

在本系列文章的[第一部分](https://studygolang.com/articles/14264)，我解释了操作系统调度器的机制和语义，如果你有计划编写多线程代码，我认为这方面知识很重要。在本系列文章的[第二部分](https://studygolang.com/articles/15316)，我解释了 Go 中的调度器背后的机制和语义，我相信这对于理解如何在 Go 中编写并发代码是至关重要的。在本文中，我会将操作系统和 Go 调度器的机制和语义结合在一起，以便更深入地理解什么是并发，什么不是并发。

本文的目标是：

- 提供语义上的指导，来考虑当前的工作负载是否适合使用并发特性。
- 向你展示如何改变不同工作负载下的语义，来更改你想要做出的工程上的决策。

## 什么是并发

并发意味着“无序”执行。拿到一组本来会有序执行的指令，然后使用无序的执行方法，最终却能得到相同的结果。所以摆在你面前的问题非常明显：无序执行会增加一些“价值”，当我在说“价值”时，我实际的意思是复杂度成本的增加，能够带来足够的性能上的收益。这取决于你具体的问题，有些时候无序执行可能无法实现或者根本没有价值。

再者，理解[并发并不等同于并行](https://blog.golang.org/concurrency-is-not-parallelism) 至关重要。并行意味着同时执行两个或多个指令。并行的概念和并发是完全不同的。只有当你拥有至少两个操作系统（OS) 或硬件线程，并且至少需要两个 Goroutines 时才能实现真正的并行，每个 Goroutine 在各自的操作系统 / 硬件线程上独立地执行指令。

___配图 1：并发和并行___

![pic1](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-schedule-part3/96_figure1.png)

在配图 1 中，你会看到包含两个逻辑处理器（P）的关系图，每个逻辑处理器的独立操作系统线程（M) 附着到机器上的独立硬件线程（核心）上。同时，你也可以看到有两个 Goroutines（G1 和 G2) 在并行地运行，同一时间在它们各自的操作系统 / 硬件线程上执行它们的指令。在每个逻辑处理器内部，有三个 Goroutines 轮流共享各自的操作系统线程。所有这些 Goroutines 都是并行运行的，它们的指令执行时是无序的，且会在操作系统线程上共享时间片。

这就是问题所在，有时候在没有并行的基础上进行并发可能会导致应用程序的吞吐量下降。同样有趣的是，有时候我们在并行的基础上同时使用并发也未必会带来性能上的提升。

## 工作负载

如何才能知道”无序执行“可能会产生收益？了解你所处理的工作负载是一个好的开始。在考虑并发时，有两种重要类型的工作负载需要注意：

- ***CPU 密集型***：这种工作负载永远不会导致 Goroutines 进入等待的状态，这是一种持续不断的计算工作。例如，计算圆周率π的第 n 位就是一个 CPU 密集型的工作。

- ***IO 密集型***：这种工作负载会导致 Goroutines 自然地进入等待的状态。这类的工作负载一般包括通过网络去访问某些资源，或者向操作系统进行系统调用，或者等待某个事件的发生。如果一个 Goroutine 需要读取一个文件，那么这就是 IO 密集型的工作。我个人倾向于将同步事件（互斥锁，原子操作）归为此类，因为这些操作也会导致 Goroutines 进入等待状态（阻塞）。

如果是 CPU 密集型的工作，你可能需要并行来提高并发性能。单线程处理多个 Goroutines 不够高效，因为 Goroutines 作为工作负载的一部分并不会切换至等待状态。如果 Goroutines 的数量多于操作系统线程的数量，也可能会减缓代码的执行速度，因为在操作系统线程上切换 Goroutines 需要考虑延迟成本（切换所花费的时间）。上下文切换会导致一个 "STW" 事件，因为在切换期间，代码并没有被真正的执行到。

如果是 IO 密集型的工作，你可能就不需要并行来提高并发性能。单个操作系统线程可以非常高效地处理多个 Goroutines，因为 Goroutines 会因为其工作性质而自然地进行*等待 - 执行*这种状态的切换。Goroutines 的数量多于操作系统线程数时，会增加相关作业的执行速度，因为上下文切换导致的延迟开销并不会产生” STW “事件。你的工作会自然地终止，这允许不同的 Goroutine 高效地利用同一个操作系统线程，从而不会让操作系统线程处于空闲状态。

那么该如何计算每个操作系统线程上运行多少个 Goroutines 才能达到最佳吞吐量？如果 Goroutines 太少，会导致更多的线程空闲时间。Goroutines 太多，又会因为上下文切换而产生更高的延迟成本。但是这个问题已经超出了本文的范畴，实际上，这个问题应该由你自己去思考。

现在来看，最重要的事情是通过检查一些代码来巩固你评估工作负载并适当地使用并发性的能力。

## 数字累加

我们不会使用复杂的用例来解释和理解这些语义。首先，我们看下面代码片段中的 `add` 函数，它执行的是两个整型数字的求和操作。

___清单 1___

[source code](https://play.golang.org/p/r9LdqUsEzEz)

```go
36 func add(numbers []int) int {
37     var v int
38     for _, n := range numbers {
39         v += n
40     }
41     return v
42 }
```

在清单 1 的第 36 行，声明了一个名为 `add` 的函数用来求一个元素为 `int` 切片的所有元素之和。在 37 行，它声明了一个名为 `v` 的变量来表示求和的结果 `sum`。然后在 38 行，该函数开始迭代切片，并依次累加各个元素。最后在 41 行，该函数返回求和结果。

问题：`add` 函数是一个适合无序执行的工作负载吗？我相信答案是肯定的。整型的切片集合可以被划分为多个小集合并且可以被并发地处理。一旦所有的集合都求和完成，就可以将各个集合的结果累加至 `sum`，这样和顺序求和的结果是一致的。

但是，这又会引入一个其他的问题。到底需要划分多少个小集合进行并发处理才能达到最佳的系统吞吐量？想要回答这个问题， 前提是你必须知道在运行的工作负载 `add` 具体是什么类型。事实上，`add` 函数是一个 CPU 密集型的作业类型，因为它运行的是纯数学运算，并且不会有其他的问题导致其 Goroutine 会进入等待状态。这也就意味着为每个操作系统线程分配一个 Goroutine 即可以达到最佳的吞吐量。

下面清单 2 是我编写的并发版本的 `add`

_注意：你可以使用多种方式来编写并发版本的 `add` 函数。不用被我写的特定的实现方式所束缚。如果你有更加好的实现方式，并且愿意分享，那真是极好的。_

___清单 2___

[source code](https://play.golang.org/p/r9LdqUsEzEz)

```go
func addConcurrent(goroutines int, numbers []int) int {
45     var v int64
46     totalNumbers := len(numbers)
47     lastGoroutine := Goroutines - 1
48     stride := totalNumbers / Goroutines
49
50     var wg sync.WaitGroup
51     wg.Add(goroutines)
52
53     for g := 0; g < Goroutines; g++ {
54         Go func(g int) {
55             start := g * stride
56             end := start + stride
57             if g == lastGoroutine {
58                 end = totalNumbers
59             }
60
61             var lv int
62             for _, n := range numbers[start:end] {
63                 lv += n
64             }
65
66             atomic.AddInt64(&v, int64(lv))
67             wg.Done()
68         }(g)
69     }
70
71     wg.Wait()
72
73     return int(v)
74 }
```

在清单 2 中出现的 `addConcurrent` 函数是之前 `add` 函数的并发版本。并发版本使用了 26 行代码，而之前简单的版本只使用了 5 行代码。多了很多代码，所以我只关注需要重点理解的代码。

___48 行___：每个 Goroutine 只需要计算一个唯一的，且更加小的列表。小列表的大小是根据总列表元素个数除以 Goroutines 数量来计算的。

___53 行___：创建了一堆 Goroutines 用来执行累加工作。

___57-59 行___：最后一个 Goroutine 将累加包含了剩余数字的列表，这些数字可能比其它 Goroutines 中的数字更大。

___66 行___：每个小列表的求和运算结果存到最终的 sum 中。

很明显，并发的版本比顺序处理的版本复杂了很多，但这种复杂性有价值吗？回答此问题最好的方式就是创建一个基准测试。在基准测试里，我使用了 1000 万个数字集合，并且关闭了垃圾回收器。同时，该基准测试中也包含了一个顺序求值的版本。

___清单 3___

```go
func BenchmarkSequential(b *testing.B) {
    for i := 0; i < b.N; i++ {
        add(numbers)
    }
}

func BenchmarkConcurrent(b *testing.B) {
    for i := 0; i < b.N; i++ {
        addConcurrent(runtime.NumCPU(), numbers)
    }
}
```

清单 3 展示了基准测试的内容。下面是所有 Goroutines 共用单个操作系统线程的结果。顺序求值版本使用了一个 Goroutine，而并发版本使用了和 CPU 核心数 `runtime.NumCPU` 一样多的 Goroutines（8 个）。在这个案例中，并发版本并没有利用到并行性。

___清单 4___

```bash
10 Million Numbers using 8 Goroutines with 1 core
2.9 GHz Intel 4 Core i7
Concurrency WITHOUT Parallelism
-----------------------------------------------------------------------------
$ GOGC=off Go test -cpu 1 -run none -bench . -benchtime 3s
goos: darwin
goarch: amd64
pkg: Github.com/ardanlabs/gotraining/topics/go/testing/benchmarks/cpu-bound
BenchmarkSequential          1000       5720764 ns/op : ~10% Faster
BenchmarkConcurrent          1000       6387344 ns/op
BenchmarkSequentialAgain     1000       5614666 ns/op : ~13% Faster
BenchmarkConcurrentAgain     1000       6482612 ns/op
```

_注意：在你本地机器上运行基准测试是很复杂的。有很多意外的因素会导致你的测试结果不精确。你需要确保你的机器尽可能地处于空闲状态且多运行几次基准测试。你希望在测试结果中看到一致性，通过测试工具运行两次基准测试，将会使得测试结果达到最一致的状态_。

清单 4 中的基准测试显示了顺序求值的版本比并发版本在使用单个操作系统线程时每次操作快了约 10%-13%。这正是我所期望的，因为并发版本在单个操作系统线程上因为上下文切换和 Goroutines 管理而造成了一些额外的开销。

下面的清单是当每个 Goroutine 都有一个独立的操作系统线程为之运行时的结果。顺序求值版本使用了单个 Goroutine，并发版本使用了 8 个（和 `runtime.NumCPU` 一致）Goroutines。在这种情况下，并发版本开始利用并行特性。

___清单 5___

```bash
10 Million Numbers using 8 Goroutines with 8 cores
2.9 GHz Intel 4 Core i7
Concurrency WITH Parallelism
-----------------------------------------------------------------------------
$ GOGC=off Go test -cpu 8 -run none -bench . -benchtime 3s
goos: darwin
goarch: amd64
pkg: Github.com/ardanlabs/gotraining/topics/go/testing/benchmarks/cpu-bound
BenchmarkSequential-8                1000      5910799 ns/op
BenchmarkConcurrent-8                2000      3362643 ns/op : ~43% Faster
BenchmarkSequentialAgain-8           1000      5933444 ns/op
BenchmarkConcurrentAgain-8           2000      3477253 ns/op : ~41% Faster
```

清单 5 中的基准测试显示了并发版本在每个 Goroutine 使用独立的操作系统线程时，每次操作的耗时比顺序求值版本快了约 %41-%43。这也正是我所期望的，因为所有的 Goroutines 都在并行运行，8 个 Goroutines 同时进行它们各自的工作。

## 排序

理解并非所有的 CPU 密集型的工作都适合并发是非常重要的。当分离工作或者合并结果的操作都非常昂贵时，这基本上是正确的。冒泡排序是解释这类问题的一个比较好的例子。下面的代码是用 Go 实现的冒泡排序算法。

___清单 6___

[source code](https://play.golang.org/p/S0Us1wYBqG6)

```go
01 package main
02
03 import "fmt"
04
05 func bubbleSort(numbers []int) {
06     n := len(numbers)
07     for i := 0; i < n; i++ {
08         if !sweep(numbers, i) {
09             return
10         }
11     }
12 }
13
14 func sweep(numbers []int, currentPass int) bool {
15     var idx int
16     idxNext := idx + 1
17     n := len(numbers)
18     var swap bool
19
20     for idxNext < (n - currentPass) {
21         a := numbers[idx]
22         b := numbers[idxNext]
23         if a > b {
24             numbers[idx] = b
25             numbers[idxNext] = a
26             swap = true
27         }
28         idx++
29         idxNext = idx + 1
30     }
31     return swap
32 }
33
34 func main() {
35     org := []int{1, 3, 2, 4, 8, 6, 7, 2, 3, 0}
36     fmt.Println(org)
37
38     bubbleSort(org)
39     fmt.Println(org)
40 }
```

清单 6 是一个用 Go 编写的冒泡排序算法。这种排序算法会扫描每次迭代时需要交换值的整数集合。依赖于列表的顺序，在对所有的元素进行排序之前，可能会多次遍历集合。

问题：`bubbleSort` 函数是适合无序执行的工作负载吗？我相信答案肯定是否定的。整数集合确实可以被划分为多个小列表然后并发地进行排序。但是，当并发排序完成后，并没有很高效的方式将这些小列表整合到一起。下面的例子是一个并发版本的冒泡排序算法的实现。

___清单 7___

```go
01 func bubbleSortConcurrent(goroutines int, numbers []int) {
02     totalNumbers := len(numbers)
03     lastGoroutine := Goroutines - 1
04     stride := totalNumbers / Goroutines
05
06     var wg sync.WaitGroup
07     wg.Add(goroutines)
08
09     for g := 0; g < Goroutines; g++ {
10         Go func(g int) {
11             start := g * stride
12             end := start + stride
13             if g == lastGoroutine {
14                 end = totalNumbers
15             }
16
17             bubbleSort(numbers[start:end])
18             wg.Done()
19         }(g)
20     }
21
22     wg.Wait()
23
24     // Ugh, we have to sort the entire list again.
25     bubbleSort(numbers)
26 }
```

清单 7 中，函数 `bubbleSortConcurrent` 是冒泡排序算法的并发版本的实现。它使用了多个 Goroutines 同时对列表的某一段进行排序。但是，你最后得到了一个按块排序后的值列表。如果我们给定一个 36 个整型数字的列表，并将其划分为 12 组，每组两个数字，如果在 25 行没有再次进行排序，将会得到如下所示的结果。

___清单 8___

```bash
Before:
  25 51 15 57 87 10 10 85 90 32 98 53
  91 82 84 97 67 37 71 94 26  2 81 79
  66 70 93 86 19 81 52 75 85 10 87 49

After:
  10 10 15 25 32 51 53 57 85 87 90 98
   2 26 37 67 71 79 81 82 84 91 94 97
  10 19 49 52 66 70 75 81 85 86 87 93
```

由于冒泡排序的本质是扫描整个列表，我们在 25 行对 `bubbuleSort` 的调用已经否定了使用并发特性会产生任何收益的可能性。

## 文件读取

上面我们已经解释了两种 CPU 密集型的工作负载，那么如果是一个 IO 密集型的工作负载呢？如果 Goroutines 能够自然地进出等待状态时，语义是否会有不同？下面让我们看看一个 IO 密集型的工作负载：文件读取，并运行一个文本搜索操作。

第一个版本是顺序版本，该函数名为 `find`：

___清单 9___

[source code](https://play.golang.org/p/8gFe5F8zweN)

```go
42 func find(topic string, docs []string) int {
43     var found int
44     for _, doc := range docs {
45         items, err := read(doc)
46         if err != nil {
47             continue
48         }
49         for _, item := range items {
50             if strings.Contains(item.Description, topic) {
51                 found++
52             }
53         }
54     }
55     return found
56 }
```

在清单 9 中，你可以看到一个名为 `find` 函数，这是一个顺序执行的版本。在 43 行中，我们声明了一个名为 `found` 变量，该变量用来维护在给定的文档中找到的指定的 `topic` 的次数。然后在 44 行，我们开始迭代所有文档，并使用 `read` 函数在第 45 行读取单个文档。最终在第 49-53 行，使用 `strings` 包中的 `Contains` 函数来检查指定的 `topic` 字符串是否被包含在文档中。如果 `topic` 存在，`found` 变量会自增计数。

下面是 `find` 函数中出现的 `read` 函数的实现。

___清单 10___

[source code](https://play.golang.org/p/8gFe5F8zweN)

```go
33 func read(doc string) ([]item, error) {
34     time.Sleep(time.Millisecond) // Simulate blocking disk read.
35     var d document
36     if err := xml.Unmarshal([]byte(file), &d); err != nil {
37         return nil, err
38     }
39     return d.Channel.Items, nil
40 }
```

清单 10 中的 `read` 函数刚开始就执行了 `time.Sleep(time.MilliSecond)` 调用来模拟从磁盘上读取文件而执行的系统调用所产生的延迟。这种延迟的一致性对于准确测量 `find` 函数的顺序版本和并发版本的差异至关重要。然后我们在第 35-39 行，存储在全局变量中的 mock 的 xml 文档被编码进一个结构体的值以便后续进行处理。最后，在第 39 行返回被处理后的项目的集合。

在完成了顺序版本之后，下面是并发版本。

_注意：你可以使用多种方式来编写并发版本的 `find` 函数。不用被我写的特定的实现方式所束缚。如果你有更加好的实现方式，并且愿意分享，那真是极好的。_

___清单 11___

[source code](https://play.golang.org/p/8gFe5F8zweN)

```go
58 func findConcurrent(goroutines int, topic string, docs []string) int {
59     var found int64
60
61     ch := make(chan string, len(docs))
62     for _, doc := range docs {
63         ch <- doc
64     }
65     close(ch)
66
67     var wg sync.WaitGroup
68     wg.Add(goroutines)
69
70     for g := 0; g < Goroutines; g++ {
71         Go func() {
72             var lFound int64
73             for doc := range ch {
74                 items, err := read(doc)
75                 if err != nil {
76                     continue
77                 }
78                 for _, item := range items {
79                     if strings.Contains(item.Description, topic) {
80                         lFound++
81                     }
82                 }
83             }
84             atomic.AddInt64(&found, lFound)
85             wg.Done()
86         }()
87     }
88
89     wg.Wait()
90
91     return int(found)
92 }
```

清单 11 中，`findConcurrent` 函数是之前 `find` 函数的并发版本。并发的版本使用了 30 多行代码，而顺序版本的代码只有 13 行。我实现并发版本的目的在于控制用来处理未知数量的文档的 Goroutines 的数量。一个基于通道的 Goroutines 池的模式是我的主要实现方式。

上面的清单中有很多的代码，这里我只会关注需要重点理解的代码。

___61-64 行___：创建了一个通道，并填充相关的文档以便后续处理。

___65 行___： 当所有的文档被处理完成后，我们需要关闭通道来通知相关的 Goroutines 终止执行。

___70 行___： 创建了一堆（一池）的 Goroutines。

___73-83 行___： 该池中的每个 Goroutine 都会从通道中接收到一个文档，并将文档读取到内存中，之后检查文档中是否包含指定的 `topic`。如果匹配成功，那么本地的 `found` 变量会自增计数。

___84 行___： 将各个 Goroutines 的计数求和，作为最终的计数。

很明显，并发的版本比顺序处理的版本复杂了很多，但这种复杂性有价值吗？回答此问题最好的方式还是创建一个基准测试。在这个基准测试里，我关闭了垃圾回收器并使用了一个拥有 1000 个文档的集合。分别来测试顺序版本的 `find` 函数和并发版本的 `findConcurrent` 函数。

___清单 12___

```go
func BenchmarkSequential(b *testing.B) {
    for i := 0; i < b.N; i++ {
        find("test", docs)
    }
}

func BenchmarkConcurrent(b *testing.B) {
    for i := 0; i < b.N; i++ {
        findConcurrent(runtime.NumCPU(), "test", docs)
    }
}
```

清单 12 展示了基准测试函数。下面是当所有 Goroutines 只有单个操作系统线程可用时的结果。顺序版本使用了一个 Goroutine，而并发版本使用了 8 个 Goroutines（这和 `runtime.NumCPU` 数量一致）。本例中，并发版本并没有利用到并行特性。

___清单 13___

```bash
10 Thousand Documents using 8 Goroutines with 1 core
2.9 GHz Intel 4 Core i7
Concurrency WITHOUT Parallelism
-----------------------------------------------------------------------------
$ GOGC=off Go test -cpu 1 -run none -bench . -benchtime 3s
goos: darwin
goarch: amd64
pkg: Github.com/ardanlabs/gotraining/topics/go/testing/benchmarks/io-bound
BenchmarkSequential                3    1483458120 ns/op
BenchmarkConcurrent               20     188941855 ns/op : ~87% Faster
BenchmarkSequentialAgain           2    1502682536 ns/op
BenchmarkConcurrentAgain          20     184037843 ns/op : ~88% Faster
```

清单 13 中的结果展示了当所有的 Goroutines 只有单个操作系统线程可用时，并发版本比顺序版本每次操作快了约 87%-88%。这也正是我所期望的，因为所有的 Goroutines 都在高效地共享单个操作系统线程。对于在进行 `read` 操作中的每个 Goroutine，自然而然的上下文切换允许在单个操作系统线程上完成更多的工作。

下面是利用了并行特性的并发版本的基准测试结果。

___清单 14___

```bash
10 Thousand Documents using 8 Goroutines with 1 core
2.9 GHz Intel 4 Core i7
Concurrency WITH Parallelism
-----------------------------------------------------------------------------
$ GOGC=off Go test -run none -bench . -benchtime 3s
goos: darwin
goarch: amd64
pkg: Github.com/ardanlabs/gotraining/topics/go/testing/benchmarks/io-bound
BenchmarkSequential-8                  3    1490947198 ns/op
BenchmarkConcurrent-8                 20     187382200 ns/op : ~88% Faster
BenchmarkSequentialAgain-8             3    1416126029 ns/op
BenchmarkConcurrentAgain-8            20     185965460 ns/op : ~87% Faster
```

清单 14 中的基准测试表明引入了额外的操作系统线程并不能提供更好的性能。

## 结论

本文的目的是，在需要确定工作负载是否适合并发特性时，提供相关的必须要考虑的语义方面的指导。我尝试提供了不同类型的算法和工作负载的示例，以便于你能看到语义上的差异以及需要考虑的不同的工作决策。

你可以很清楚地看到，IO 密集型的工作负载并不需要并行性来获得性能上的巨大提升。相反，在 CPU 密集型的工作负载中，诸如像冒泡排序这样的算法时，并发特性的使用不仅会增加复杂性，而且没有任何实际的性能优势。确定你的工作负载是否适合并发，然后使用正确的语义来定义你的工作负载是至关重要的。

---

via: https://www.ardanlabs.com/blog/2018/12/scheduling-in-go-part3.html

作者：[William Kennedy](https://www.ardanlabs.com)
译者：[barryz](https://github.com/barryz)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
