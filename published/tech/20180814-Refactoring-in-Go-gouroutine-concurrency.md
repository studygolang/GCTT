首发于：https://studygolang.com/articles/17487

# 使用 Go 重构 - Goroutine 并发

很意外，我这些天开始写 Go 了！

最近，我发现了一些使用简单的并发解决方案的代码。鉴于我已经使用过类似的模式，我得到的结论是，它应该是受基本 Goroutines 示例代码的启发。

## 场景

假设你希望运行特定数量的任务，而这些任务很容易并行化 ( 没有副作用，没有外部依赖等等 )，并且希望存储每个任务的结果。Go 的解决方案就是使用多个 Goroutine。

我重构的真实代码是：通过调用 100 次 `net.LookupHost` 来计算系统中的平均 DNS 延迟。让我们来看看如何实现。

## 随意收集的一段代码

这是我随意收集的（改编和简化）一段实现代码。让我们就此进行讨论：

```go
func AverageLatency(host string) (latency int64, err error) {
    CONCURRENCY := 4
    REQUESTS_LIMIT := 100

    dnsRequests := make(chan int, REQUESTS_LIMIT)
    results := make(chan int64, REQUESTS_LIMIT)
    errorsResults := make(chan string, REQUESTS_LIMIT)

    for w := 1; w <= CONCURRENCY; w++ {
        go dnsTest(dnsRequests, results, errorsResults, host)
    }

    for j := 1; j <= REQUESTS_LIMIT; j++ {
        dnsRequests <- j
    }
    close(dnsRequests)

    requestsDone := 1
    for a := 1; a <= REQUESTS_LIMIT; a++ {
        select {
        case latencyLocal := <-results:
            latency = latency + latencyLocal
            requestsDone = requestsDone + 1
        case errorMsg := <-errorsResults:
            return 0, errors.New(errorMsg)
        case <-time.After(time.Second * DURATION_SECONDS):
            return latency / int64(requestsDone), nil
        }
    }
    return latency / int64(requestsDone), nil
}


func dnsTest(jobs <-chan int, results chan<- int64, errResults chan<- string, host string) {
    for range jobs {    // 译注：原文此处 range 用法错误，按程序目的看，此处 for-range 的结果应该是忽略。
        start := time.Now()
        if _, err := net.LookupHost(host); err != nil {
            errResults <- err.Error()
        }
        results <- time.Since(start).Nanoseconds() / int64(time.Millisecond)
    }
}
```

代码主要是依次执行以下步骤 :

- 运行 `CONCURRENCY` 个 Goroutines ，均执行 `dnsTest` 方法。
- 通过 `dnsRequests` channel 发送 `REQUESTS_LIMIT` 个任务到各个 Goroutine。
- 关闭 `dnsRequests` channel，向各个 Goroutine 发出信号，表示不再有任何任务（所以它们就不需要等待了）。
- 它等待来自结果的 `REQUESTS_LIMIT` 值的数量，并且我们将每个值添加到延迟变量。它还会监听错误和超时，如果其中任何情况发生，则退出该函数。
- 在超时的情况下，它使用我们到目前为止已有的值来计算平均延迟并退出。
- 最后，它将延迟中累积的值与为确定平均延迟所做的请求数量进行划分。

嗯。这是一种实现方法，但是好像有点不人性化。

## 使用 Goroutine 代价很小，所以我们应该发挥它的优势

Goroutine 是轻量级线程，这意味着我们可以在不影响程序性能的情况下启动大量线程。那么，为什么不使用与任务一样多的 Goroutine 呢？通过这种方式，我们摆脱了一个仅用于迭代的 channel，以及与之配套的所有业务逻辑。我们还可以在循环中移动 Goroutine 代码，因为它足够小 :

```go
func AverageLatency(host string) (latency int64, err error) {
    REQUESTS_LIMIT := 100

    results := make(chan int64, REQUESTS_LIMIT)
    errorsResults := make(chan string, REQUESTS_LIMIT)

    for w := 1; w <= REQUESTS_LIMIT; w++ {
        go func() {
            start := time.Now()
            if _, err := net.LookupHost(host); err != nil {
                errorResults <- err.Error()
                return
            }
            results <- time.Since(start).Nanoseconds() / int64(time.Millisecond)
        }
    }

    requestsDone := 1
    for a := 1; a <= REQUESTS_LIMIT; a++ {
        select {
        case latencyLocal := <-results:
            latency = latency + latencyLocal
            requestsDone = requestsDone + 1
        case errorMsg := <-errorsResults:
            return 0, errors.New(errorMsg)
        case <-time.After(time.Second * DURATION_SECONDS):
            return latency / int64(requestsDone), nil
        }
    }
    return latency / int64(requestsDone), nil
}
```

代码逻辑现在更清晰了。我们循环执行 `REQUESTS_LIMIT` 次，每次创建一个检查 DNS 延迟的 Goroutine。。然后我们进行 `REQUESTS_LIMIT` 循环，等待 `results`、`errorResults` channel 和 `time.After()` 的结果。

## Leveraging WaitGroup and cleaning up

我对最后一个循环不太满意，它看起来有点脆弱。Go 有更优雅的工具，可以帮助我们从创建出的 Gouroutine 中获取的结果。其中一个是 `WaitGroup`。`Waitgroup` 是一个等待一组协程完成的结构，而我们不需要做太多工作。我们只使用 `WaitGroup.Add` 增加 `WaitGroup` 计数器来统计我们使用了多少 Goroutine。每当一个 Goroutine 完成任务时，我们调用 `WaitGroup.Done` 即可。

```go
func AverageLatency(host string) (latency int64, err error) {
    REQUESTS_LIMIT := 100
    results := make(chan int64, REQUESTS_LIMIT)
    errorsResults := make(chan string, REQUESTS_LIMIT)

    var wg sync.WaitGroup
    wg.Add(REQUESTS_LIMIT)

    for j := 0; j < REQUESTS_LIMIT; j++ {
        go func() {
            defer wg.Done()
            start := time.Now()
            if _, err := net.LookupHost(host); err != nil {
                errorResults <- err.Error()
                return
            }
            results <- time.Since(start).Nanoseconds() / int64(time.Millisecond)
        }
    }

    wg.Wait()

    ...
}
```

这样我们就不需要通道来收集错误和结果，我们可以使用更传统的数据类型。

出于监控的目的，我还想收集错误的数量，而不仅仅是在有一个错误时返回。 如果我们想查看实际的错误信息，我们可以手动检查日志：

```go
type Metrics struct {
    AverageLatency float64
    RequestCount   int64
    ErrorCount     int64
}

func AverageLatency(host string) Metrics {
    REQUESTS_LIMIT := 100
    var errors int64
    results := make([]int64, 0, DEFAULT_REQUESTS_LIMIT)

    var wg sync.WaitGroup
    wg.Add(REQUESTS_LIMIT)

    for j := 0; j < REQUESTS_LIMIT; j++ {
        go func() {
            defer wg.Done()
            start := time.Now()
            if _, err := net.LookupHost(host); err != nil {
                fmt.Printf("%s", err.Error())
                atomic.AddInt64(&errors, 1)
                return
            }
            append(results, time.Since(start).Nanoseconds() / int64(time.Millisecond))
        }
    }

    wg.Wait()

    return CalculateStats(&results, &errors)
}
```

我们创建了一个新类型 `Metrics`，用于对需要监视的值进行封装。我们现在将 `wg` 计数器设置为我们计划执行的请求数量，并在每个 `goroutine` 完成时调用 `wg.Done`，无论是否成功。 然后我们等待所有的 `goroutines` 完成。我们还将所有统计计算提取到外部 `CalculateStats` 函数，以便 `AverageLatency` 函数专注于收集原始值。

为了展示完整实例，`CalculateStats` 的实现可能是这样的：

```go
// Takes amount of requests and errors and returns some stats on a
// `Metrics` struct
func CalculateStats(results *[]int64, errors *int64) Metrics {
    successfulRequests := len(*results)
    errorCount := atomic.LoadInt64(errors)

    // Sum up all the latencies
    var totalLatency int64 = 0
    for _, value := range *results {
        totalLatency += value
    }

    avgLatency := float64(-1)

    if successfulRequests > 0 {
        avgLatency = float64(totalLatency) / float64(successfulRequests)
    }

    return Metrics{
        avgLatency,
        int64(successfulRequests),
        errorCount
    }
}
```

## 等等，超时呢？

好吧，我们在随后一次重构代码的时候忘记存储超时信息了。

使用 `WaitGroup` 设置超时有点麻烦，但使用 channel 变得容易。我们不会直接调用 `wg.Wait`，而是将这个调用封装在一个名为 `waitWithTimeout` 的新函数中：

```go
func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
    c := make(chan struct{})
    go func() {
        defer close(c)
        wg.Wait()
    }()

    select {
    case <-c:
        return false
    case <-time.After(timeout):
        return true
    }
}
```

在这里，我们创建一个 `wg.Wait()` 运行后将关闭的 channel，这意味着所有 `goroutine` 都已完成。然后我们使用 `select` 语句在所有 `goroutine` 完成时或者超时时间已经过去时返回。在这种情况下，我们返回 `true` 表示已发生超时。

我们的终极版本的 `AverageLatency` 将如下所示：

```go
func AverageLatency(host string) Metrics {
    REQUESTS_LIMIT := 100
    var errors int64
    results := make([]int64, 0, REQUESTS_LIMIT)

    var wg sync.WaitGroup
    wg.Add(REQUESTS_LIMIT)

    for j := 0; j < REQUESTS_LIMIT; j++ {
        go func() {
            defer wg.Done()
            start := time.Now()
            if _, err := net.LookupHost(host); err != nil {
                fmt.Printf("%s", err.Error())
                atomic.AddInt64(&errors, 1)
                return
            }
            append(results, time.Since(start).Nanoseconds() / int64(time.Millisecond))
        }
    }

    if waitWithTimeout(&wg, time.Duration(time.Second*DURATION_SECONDS)) {
        fmt.Println("There was a timeout waiting for DNS requests to finish")
    }
    return CalculateStats(&results, &errors)
}
```

如果超时，我们只显示一条消息，因为我们仍然想知道有多少请求和错误，无论是否有超时。

## 额外：竞争状态 !!

哈，我相信你认为会认为世界上的一切都很美好。但不是！我在代码中潜入了竞争状态。有些人可能已经发现了这个问题。对于还没有发现竟态的朋友，建议你先回头再看一看上面的代码。

你找到了没？没找到也没关系。

好了，让我们看看这整个竞争条件是什么。

## 在 Go 中，很少情况是线程安全的

每当我们使用 `Goroutine` 操作时，我们必须小心修改外部变量，因为它们通常不是线程安全的。在我们的 main Goroutine 中，我们使用 `atomic.AddInt64` 以线程安全的方式存储错误信息，但是我们在 `results` 切片中存储延迟信息，这不是线程安全的。并不能绝对保证结果中的结果数量与我们尝试的请求数量相匹配。

在 Go 中我们有几种方法解决这个问题，例如：

- 返回到使用 `channel` 而不是切片来收集延迟信息。
- 使用互斥体来协调对结果的访问。
- 使用缓冲区容量为 1 的 `channel` 作为队列，并将延迟信息发送到该通道。

考虑到我们已经深陷 Goroutine，我选择使用第三种方法。这也是我发现的最常用的方法 :

```go
func AverageLatency(host string) Metrics {
    REQUESTS_LIMIT := 100
    var errors int64
    results := make([]int64, 0, REQUESTS_LIMIT)
    successfulRequestsQueue := make(chan int64, 1)

    var wg sync.WaitGroup
    wg.Add(DEFAULT_REQUESTS_LIMIT)

    for j := 0; j < REQUESTS_LIMIT; j++ {
        go func() {
            start := time.Now()

            if _, err := net.LookupHost(host); err != nil {
                atomic.AddInt64(&errors, 1)
                wg.Done()
                return
            }

            successfulRequestsQueue <- time.Since(start).Nanoseconds() / 1e6
        }()
    }

    go func() {
        for t := range successfulRequestsQueue {
            results = append(results, t)
            wg.Done()
        }
    }()

    if waitTimeout(&wg, time.Duration(time.Second*DURATION_SECONDS)) {
        fmt.Println("There was a timeout waiting for DNS requests to finish")
    }
    return CalculateDNSReport(&results, &errors)
}
```

我们正在创建了一个 `successfulRequestsQueue` channel，它只能缓冲一个值，相当于创建一个同步队列。我们现在可以在这个 channel 发送延迟信息结果，而不是直接将结果添加到 `result` 切片中。然后我们在一个新的 Goroutine 中循环遍历 `successRequestsQueue` 中的所有传入延迟信息，然后将其添加到 `results`。我们还将 `wg.Done()` 调用移到了 `append` 之后。这样我们就可以确保每个结果都得到处理，并且不会出现竞争状态。

## 小结

我们还可以进一步重构这段代码，但是对于本文的目的，我认为我们现在可以停止了。如果你还想继续，我建议你做以下改进 :

- 使计算函数 ( 现在是 `net.LookupHost`) 变得通用，这样我们就可以在测试中使用不同的函数。
- 如果没有 Goroutines，如何编写这段代码 ?
- 使代码完全独立于上下文，以便我们可以在任何时候使用它来存储来自许多不同操作的结果。

---

via: https://itnext.io/refactoring-in-go-goroutine-concurrency-fccbe7093c04

作者：[Sergi Mansilla](https://itnext.io/@sergimansilla)
译者：[7Ethan](https://github.com/7Ethan)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
