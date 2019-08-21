首发于：https://studygolang.com/articles/22898

# Goroutine 内存泄漏 - 被遗弃的接收者

## 介绍

Goroutine 内存泄漏是产生 Go 程序内存泄漏的常见原因。在我之前的[文章](https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html)中，我介绍了 Goroutine 内存泄漏，并展示了许多 Go 开发人员容易犯错的例子。继续前面的内容，这篇文章提出了另一个关于 Goroutines 如何出现内存泄露的情景。

## 泄漏：被遗弃的接收者

**_在此内存泄漏示例中，您将看到多个 Goroutines 被阻塞等待接收永远不会发送的值。_**

文章中程序启动了多个 Goroutines 来处理文件中的一批记录。每个 Goroutine 从输入通道接收值，然后通过输出通道发送新值。

### 示例一

[https://play.golang.org/p/Jtpla_UvrmN](https://play.golang.org/p/Jtpla_UvrmN)

```golang
35 // processRecords is given a slice of values such as lines
36 // from a file. The order of these values is not important
37 // so the function can start multiple workers to perform some
38 // processing on each record then feed the results back.
39 func processRecords(records []string) {
40
41     // Load all of the records into the input channel. It is
42     // buffered with just enough capacity to hold all of the
43     // records so it will not block.
44
45     total := len(records)
46     input := make(chan string, total)
47     for _, record := range records {
48         input <- record
49     }
50     // close(input) // What if we forget to close the channel?
51
52     // Start a pool of workers to process input and send
53     // results to output. Base the size of the worker pool on
54     // the number of logical CPUs available.
55
56     output := make(chan string, total)
57     workers := runtime.NumCPU()
58     for i := 0; i < workers; i++ {
59         go worker(i, input, output)
60     }
61
62     // Receive from output the expected number of times. If 10
63     // records went in then 10 will come out.
64
65     for i := 0; i < total; i++ {
66         result := <-output
67         fmt.Printf("[result  ]: output %s\n", result)
68     }
69 }
70
71 // worker is the work the program wants to do concurrently.
72 // This is a blog post so all the workers do is capitalize a
73 // string but imagine they are doing something important.
74 //
75 // Each goroutine can't know how many records it will get so
76 // it must use the range keyword to receive in a loop.
77 func worker(id int, input <-chan string, output chan<- string) {
78     for v := range input {
79         fmt.Printf("[worker %d]: input %s\n", id, v)
80         output <- strings.ToUpper(v)
81     }
82     fmt.Printf("[worker %d]: shutting down\n", id)
83 }
```

在第 39 行，`processRecords` 定义了一个被调用的函数。该函数接受 `[]string` 值。在第 46 行，`input` 创建一个被调用的缓冲通道。第 47 和 48 行运行一个循环，复制 `string` 切片中的每个值并将它们发送到通道。`input` 创建的通道具有足够的容量来保存切片中的每个值，因此第 48 行上的通道发送都不会阻塞。此通道是用于在多个 Goroutines 之间分配值的管道。

接下来在第 56 到 60 行，该程序创建了一个 Goroutines 池来接收管道中的工作。在第 56 行，创建一个名为 `output` 的缓冲通道; 这是每个 Goroutine 将发送其结果的地方。第 57 到 59 行运行循环并使用 `worker` 函数创建多个 Goroutines。 Goroutines 的数量等于机器上的逻辑 CPU 的数量。循环变量的副本 `i` 以及 `input` 和 `output` 通道都传递给 Goroutine。

`worker` 函数在第 77 行定义。函数的签名定义中 `input` 为 `<-chan string` ，这意味着它是一个只接收通道。该函数也接受 `output` 参数, `chan<- string` 类型这意味着它是一个只发送通道。

示例第 78 行，在函数内部 Goroutines 使用 `range` 循环从 `input` 通道接收数据，直到通道关闭并且没有值。对于每次迭代，将接收到的值分配给 `v` 并在第 79 行打印迭代变量。然后在第 80 行，`worker` 函数传递 `v` 给 `strings.ToUpper` 函数返回新的 `string` ，并立即在 `output` 上发送新的 `string` 。

回到 `processRecords` 函数中，执行已经向下移动到第 65 行，在那里运行另一个循环。该循环迭代，直到它接收并处理了来自 `output` 通道的所有值。在第 66 行， `processRecords` 函数等待从一个工作者 Goroutines 接收一个值。接收到的值打印在第 67 行。当程序收到每个输入的值时，它退出循环并终止。

运行此程序打印转换后的数据，因此它似乎工作，但该程序正存在多个 Goroutines 内存泄漏。该程序从未到达第 82 行，该行将宣布程序正在关闭。即使在 `processRecords` 函数返回之后，每个 `worker` Goroutines 仍处于活动状态并等待第 78 行的输入。通道会一直接收数据直到通道关闭并为空。问题是程序永远不会关闭通道。

## 修复：信号完成

修复泄漏只需要一行代码: `close(input)` 。关闭频道是表示”不再发送数据“的一种方式。关闭通道的最合适位置是在第 50 行发送最后一个值之后，如示例二所示：

### 示例二

[https://play.golang.org/p/QNsxbT0eIay](https://play.golang.org/p/QNsxbT0eIay)

```golang
45     total := len(records)
46     input := make(chan string, total)
47     for _, record := range records {
48         input <- record
49     }
50     close(input)
```

关闭缓冲区中仍有值的缓冲通道是有效的; 频道仅关闭发送而不是接收。 `worker` Goroutines 运行 `range input` 将通过缓冲区来工作，直到他们发出通道已关闭的信号。这可以让 `workers` 在终止之前完成循环。

## 结论

正如前一篇文章中所提到的，Go 使得启动 Goroutines 变得简单，但是你有责任仔细使用它们。在这篇文章中，我展示了另一个很容易出现的 Goroutine 错误。还有很多方法可以创建 Goroutine 内存泄漏以及使用并发时可能遇到的其他陷阱。未来的帖子将继续讨论这些问题。与往常一样，我将继续重复这一建议：“如果不知道它会如何停止，就不要开始使用 goroutine ”。

**_并发是一种有用的工具，但必须谨慎使用。_**

---

via: <https://www.ardanlabs.com/blog/2018/12/goroutine-leaks-the-abandoned-receivers.html>

作者：[Jacob Walker](https://github.com/jcbwlkr)
译者：[lovechuck](https://github.com/lovechuck)
校对：[zhoudingding](https://github.com/dingdingzhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
