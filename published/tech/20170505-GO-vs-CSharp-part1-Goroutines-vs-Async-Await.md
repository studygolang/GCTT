首发于：https://studygolang.com/articles/19399

# Go 与 C# 对比 第一篇：Goroutines 与 Async-Await

> 我将写一个系列的文章，来对比 C# 与 GO( 译者：就两篇 ),Go 的核心特性是 Goroutines, 这是一个非常棒的起点，C# 的替代方案是使用 Async/Await 来支持这个特性。

但是实现的方式上还是有一些差异的：

* C# 中对于 Async/Await 的实现是基于编译器提供的方法体，类似于 C# 对 IEnumerable<T> / IEnumerator<T> methods 的实现。编译器生成一个方法返回状态，返回值作为是否异步计算的标志。
* Goroutines 在 Go 中特别常见，当你开始使用 "Go" 关键字的语法糖的时候，所有神奇的关联魔法就开始执行了。Go 异步编程使用了一个轻量级的线程，实际上，一个线程使用了很小的堆 与一个能够异步等待读操作、暂停自身、释放操作系统线程相比较，前者肯定是更轻量级的。
* Go 中没有 "await" 的概念，取代的方式是使用通道 (Channel) 来进行通信，稍后我会解释为什么 Go 不需要这个概念。
* 这里还有很多不同的地方 ———— 我会在后续的很多地方提到它们，但是，整体上来讲，Async/Await 是构建在 C# 的平台之上的，也就是说，.NET CLR 对于这块内容不需要做额外的修改，Go 与之不同的是，goroutines 已经深度的集成在 Go 的运行时机制中。

接下来，我会做一些简单的测试：

* 创建 n 个 Goroutines，每个 Goroutines 在通道上面等待一个数字，并在他的基础上自增，并发送给输出通道。
* Goroutines 和 channels 连接在一起，因此发送给第一个通道的消息会被传送到最后一个通道上面。

**Go 代码如下：**

```go
package main

import (
    "flag";
    "fmt";
    "time"
)

func measure(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s", name, elapsed)
    fmt.Println()
}

var maxCount = flag.Int("n", 1000000, "how many")

func f(output, input chan int) {
    output <- 1 + <-input
}

func test() {
    fmt.Printf("Started, sending %d messages.", *maxCount)
    fmt.Println()
    flag.Parse()
    defer measure(time.Now(), fmt.Sprintf("Sending %d messages", *maxCount))
    finalOutput := make(chan int)
    var left, right chan int = nil, finalOutput
    for i := 0; i < *maxCount; i++ {
        left, right = right, make(chan int)
        Go f(left, right)
    }
    right <- 0
    x := <-finalOutput
    fmt.Println(x)
}

func main() {
    test()
    test()
}
```

**CSharp 代码：**

```
using System;
using System.Diagnostics;
using System.Linq;
using System.Threading.Tasks;
using System.Threading.Tasks.Channels;

namespace ChannelsTest
{
    class Program
    {
        public static void Measure(string title, Action<int, bool> test, int count, int warmupCount = 1)
        {
            test(warmupCount, true); // Warmup
            var sw = new Stopwatch();
            GC.Collect();
            sw.Start();
            test(count, false);
            sw.Stop();
            Console.WriteLine($"{title}: {sw.Elapsed.TotalMilliseconds:0.000}ms");
        }

        static async void AddOne(WritableChannel<int> output, ReadableChannel<int> input)
        {
            await output.WriteAsync(1 + await input.ReadAsync());
        }

        static async Task<int> AddOne(Task<int> input)
        {
            var result = 1 + await input;
            await Task.Yield();
            return result;
        }

        static void Main(string[] args)
        {
            if (!int.TryParse(args.FirstOrDefault(), out var maxCount))
                maxCount = 1000000;
            Measure($"Sending {maxCount} messages (channels)", (count, isWarmup) => {
                var firstChannel = Channel.CreateUnbuffered<int>();
                var output = firstChannel;
                for (var i = 0; i < count; i++) {
                    var input = Channel.CreateUnbuffered<int>();
                    AddOne(output.Out, input.In);
                    output = input;
                }
                output.Out.WriteAsync(0);
                if (!isWarmup)
                    Console.WriteLine(firstChannel.In.ReadAsync().Result);
            }, maxCount);
            Measure($"Sending {maxCount} messages (Task<int>)", (count, isWarmup) => {
                var tcs = new TaskCompletionSource<int>();
                var firstTask = AddOne(tcs.Task);
                var output = firstTask;
                for (var i = 0; i < count; i++) {
                    var input = AddOne(output);
                    output = input;
                }
                tcs.SetResult(-1);
                if (!isWarmup)
                    Console.WriteLine(output.Result);
            }, maxCount);
        }
    }
}

```

## Go 输出内容：

```
C:\Projects\GoTest\src>go run ChannelsTest.go
Started, sending 1000000 messages.
1000000
Sending 1000000 messages took 3.5034779s
Started, sending 1000000 messages.
1000000
Sending 1000000 messages took 808.9572ms
```

## C# 输出内容：

```
C:\Projects\ChannelsTest>dotnet run -c Release -f netcoreapp1.1
1000000
Sending 1000000 messages (channels): 3545.006ms
1000000
Sending 1000000 messages (Task<int>): 1693.675ms
```

在我们讨论结果之前，关于我们的测试代码，有几个注意点我们在这里讨论一下：

* 这个测试主要是为了 Go 准备的，在 C# 中，不需要 channel 来进行同步通信，任务通常是通过彼此异步等待获取结果。而对于 Go 来说，只能通过 channel 来实现，所以这里是使用通道来测试。
* C# 没有公开实现的通道，我现在使用的是一个在之后会公开的测试版本的通道，名字叫做[System.Threading.Tasks.Channels](https://github.com/dotnet/corefxlab/blob/master/src/System.Threading.Tasks.Channels/README.md)，现在可以通过[Nuget](https://dotnet.myget.org/feed/dotnet-corefxlab/package/nuget/System.Threading.Tasks.Channels/0.1.0-e160430-1) 来获取它 , 目前的版本是 0.1
* 为了公平起见，C# 中除了通道的测试外，我还用了一个异步任务的测试，代码里面，每个 task 等待他的 "input"task, 在得到的数字上自增 1，并返回自增后的结果。
* C# 有一个预热的功能，Go 没有，预热的逻辑会导致任何小的函数第一次执行的时候，都会花费更长的时间。

## 原始结果对比：

* 第一次执行，Go 和 C# 的时间基本相同。
* 第二次 Go 快了很多，大概提升了 3.4 倍，C# 没有执行第二次，因为他的速度始终是一样的。
* 基于任务版本的 C# 代码，仍然是 Go 第二次执行的两倍时长。

所以为啥 Go 第二次执行这么快嘞？ 解释起来很简单，当你启动 Goroutine 的时候，Go 需要分配 8K 的堆内存给他 ( 译者注：不一定是都是 8K，不同版本大小不一样 )，而这些内存可以重用，所以第二次的时候，不需要分配更多的内存给他，证明图如下：

![Memory Detail](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-vs-csharp/1780316-075137a9f9b487bb.png)

Go 为 1M 数量的 Goroutine 分配了近 9GB 的内存，假设每个 Goroutine consimes 至少需要 8KB，那么这些内存就大概达到了 8GB。

如果我们增加测试数量到 2M 的话，我的机器直接挂了 ( 内存不足 )。

所以两者的差距显而易见，让我们来思考一下为啥 C# 生成的这么慢：

* System.Threading.Tasks.Channels 目前还在测试阶段，性能上确实有一些不足，所以慢了两倍情有可原。
* 如果去掉通道的话，仍然有两倍以上的差距，虽然代码里有一个 `await Task.Yield()`, 但这个是必不可少的，.net 通过它来实现任务返回的。因此，他很快地调用堆方法并在 StackOverflowException 结束。在实际使用中，这不是一个问题，在异步代码里本来也不建议存在很长的递归链。不过，在这个测试里，它降低了 1.5 倍左右的性能。
* 尽管在 C# 中，task 是一个轻量级的对象，但是他仍然是需要堆分配内存的。状态本身也是一个引用类型。在现在的 GC 算法中，堆内存分配是很快地，但是他们可能比类似的堆扩展和调用慢 5-10 倍。

现在让我们修改一下测试的内容，将传递的消息降低到 20K，这个值比较接近于我们现实应用的最大值 ( 服务器上 Socket 的套接字接近 20K)

```
C:\Projects\ChannelsTest>go run ChannelsTest.go
Started, sending 20000 messages.
20000
Sending 20000 messages took 75.0496ms
Started, sending 20000 messages.
20000
Sending 20000 messages took 18.0513ms

C:\Projects\ChannelsTest>dotnet run -c Release -f netcoreapp1.1
20000
Sending 20000 messages (channels): 49.297ms
20000
Sending 20000 messages (Task<int>): 28.702ms
```

结果不难看出，两者很接近了：

* Go 第一次的表现不进入人意
* C# 比第二次的 Go 慢 1.5 倍
* C# 通道的方式慢 2.7 倍

数量换到 5K 的时候：
```
C:\Projects\ChannelsTest>go run C:\Projects\GoTest\src\ChannelsTest.go
Started, sending 5000 messages.
5000
Sending 5000 messages took 15.0399ms
Started, sending 5000 messages.
5000
Sending 5000 messages took 8.0213ms

C:\Projects\ChannelsTest>dotnet run -c Release -f netcoreapp1.1
5000
Sending 5000 messages (channels): 15.027ms
5000
Sending 5000 messages (Task<int>): 6.881ms
```

在这里，可以看到 C# 基于 task 的效果，比 Go 还要好，C# 的通道测试，比 Go 最好的状态慢 2 倍左右。

## 为啥 C# 使用 Task 的效果更好？

* Go 上的 5K 测试使用 ~ 5MB RAM，这仍然小于 Core i7 的 L3 缓存大小，但远远大于 L2 缓存大小 ; 另一方面，不太清楚为什么性能不如第二次访问时的性能好—— CPU 只缓存被访问的数据子集。
* c# 版本，是 prob。10x 内存效率更高，在这个测试中使用了 ~ 500KB 的 RAM，这更接近核心 i7 的 L2 缓存大小 ( 每个核心 256KB)

## Goroutines vs async-await 结论：

我们看一下最关键的不同点：

* **Goroutines 显然是更快**：在真实场景下，Go 可以得到 2.x 甚至 3.x 的速度提升，另一方面，两者还都是很高效的：在 C# 中你可以得到 1M/S 的 "awaits" 效果，Go 中大概是 2-3M，这个数字其实比较大了。举个例子：如果你在处理网络信息，这意味着 C# 的 core i7 处理器，每秒大概可以处理 10W 条信息，也就是说，在实际服务器上处理的消息更多。也就是说，这并不会成为瓶颈。
* **在实际生产中，C# 的性能很接近于 Goroutines**：C# 非常节省内存，大多数用于生产环境的应用比较依赖于他的内存集的大小。
* **8K 的 Goroutines 内存，意味着更容易产生 OOM 的错误**： 如果你的服务器通过 Goroutine 来处理所有的消息，但是服务器都在等待一些外部的服务。如果请求的频率非常高的话，那么根据上面的测试，非常容易产生 OOM 错误。2-3M 每条的数据，你需要大概 32GB 的内存。
* **默认情况下，C# 为异步调用做了很多别的事情，导致他比较慢**：需要提及的是，他通过异步等待调用链传递 ExecutionContext 和 SynchronizationContext ( 即，每个调用都有多个字典查找对应的线程本地变量 )
* **C# 的模式更健壮**：所有的异步代码都通过 async-await 封装，还有一些自带的原语，对取消的支持、同步等。我使用的通道库就是一个很好的例子 : 在 c# 中添加对通道的支持相对容易，但是类似于 async- wait in Go 的东西需要更多的样板代码。
* **C# 的代码更利于扩展**：实际上，你可以通过添加自己的调度程序，等待器，甚至自己的任务类型来实现你想要的功能。你如果真的对性能很敏感，你可以写一些非常轻量级的 task 或者用一些预先调度的任务 (ValueTask<T> 就是一个很好的例子，它现在是 . net 的一部分 )。另一个即将到来的特性是对异步序列 (async streams) 的支持——它也是基于相同的 API 集 ( 尽管它需要对 c# 编译器进行更改 )。
* **Goroutines 更简单易学**：你几乎不需要学习什么特别的东西，"Go" 关键字 + 通道就可以帮你完成你想要的一切。相反的是，C# 中的 async-await 对于异步编程来说有点难，你需要了解 Task / Task<T>，Task.Run() 和取消的最低限度。现实生活中的异步编程意味着您了解调度、.configurewait (false) 方法、task 如何工作、异常如何处理、何时使用 ValueTask 等等。这比 Go 要复杂得多
* **Goroutines 不需要考虑“ async all the way ”的问题**：Async-await 意味着，如果你设计了一个调用链 (A 调用 B，B 调用 C …… Y 调用 Z) 如果 A 和 Z 都是异步方法，B …… Z 也都必须是异步方法，否则将无法工作 ( 同步的 Y 无法等待 Z，同步的 X 也无法等待 Y)。而在 Go 中就不存在这个问题，你可以在任何的函数中通过通道获取数据，无论如何，他们都是异步的。这实际上是一个很大的优势，因为你不需要提前计划什么是异步的，而且你还可以写一个 Query() 的方法，帮助你获取到结果，程序可以根据需要来实现异步或者同步的操作，作为方法作者的你，却不需要关注这个。
* **Go 中的异步代码，开销更小**：意味着任何潜在的异步 API 在 . net 中都必须是异步的，也就是说，您需要在那里创建更多的异步任务，分配更多的堆，等等。
* **这也解释了为什么没有必要在 Go 中 async-await**：因为任何函数支持异步等待 ( 通道 ), 可以同时开始，任何函数返回一个常规结果里面可以运行一些异步逻辑——它需要的是另一个 Goroutine 开始 , 传递消息到一个新的通道 , 通过这个通道 , 等待结果。这就是为什么 Go 中的大多数 API 看起来都是同步的，实际上它们是异步的。

总的来说，实现方式差异很明显，所以影响也是非常显著的。在后续的文章中，我会用 async-await-goroutines 编写更加健壮 / 真实的测试。

---

via: https://medium.com/@alexyakunin/go-vs-c-part-1-goroutines-vs-async-await-ac909c651c11

作者：[Alex Yakunin](https://medium.com/@alexyakunin)
译者：[JYSDeveloper](https://github.com/JYSDeveloper)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
