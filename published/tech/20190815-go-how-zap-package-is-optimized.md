首发于：https://studygolang.com/articles/30250

# zap 包是如何优化的

![插图由“go 之旅”提供，原图由 Renee French 创作](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20190815-go-how-zap-package-is-optimized/1__mMI_UYf-DsS04MU5AnRQg.png)

Go 生态系统有许多流行的日志库，选择一个可以在所有项目中使用的日志库对于保持最小的一致性至关重要。易用性和性能通常是我们在日志库中考虑的两个指标。接下来我们回顾一下 [Uber](https://github.com/uber-go) 开发的 [Zap](https://github.com/uber-go/zap) 日志库。

## 核心思想

Zap 基于三个概念优化性能，第一个是：

- 避免使用 `interface{}` 有利于强类型的设计。

这一点隐藏另外两个概念：

- 无反射。反射是有代价的，而且可以避免，因为包能够决定被调用的类型。

在 JSON 编码中没有额外内存分配。 如果对标准库进行了优化，则可以轻松避免在此处进行内存分配，因为 package 包含所有已发送参数的类型。

以上几点，对开发人员来说成本不高，因此他们需要在记录消息时声明每种类型：

```go
logger.Info("failed to fetch URL",
	// Structured context as strongly typed Field values.
	zap.String("url", `http://foo.com`),
	zap.Int("attempt", 3),
	zap.Duration("backoff", time.Second),
)
```

每个字段的显式声明将允许包在日志记录过程中高效地工作。让我们回顾一下包的设计，以了解这些优化将在何处发生。

## 设计

在高亮显示包的优化部分之前，让我们绘制日志库的全局工作流：

![Zap 包工作流](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20190815-go-how-zap-package-is-optimized/1_4mn192sJdR0rU8RQ3aQo4w.png)

第一步优化，为了避免进行系统分配，我们看到优化使用同步池在记录消息。每个要记录的消息都将重用之前创建的结构体(structure)，并将其释放到池中。

第二部优化，涉及编码器和 JSON 的存储方式。要记录的每个字段都是强类型的，如前一节所示。它允许编码器通过直接将值转储到缓冲区来避免反射和分配：

![优化过的 JSON 编码器](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20190815-go-how-zap-package-is-optimized/1_9aSmDmZ1ccHfcSSwxLGsUw.png)

这个缓冲区的管理要感谢 `sync.Pool`.

最终调用方的性能/成本的权衡非常有趣，因为显式声明每个字段不需要开发人员付出太多努力。但是，该库为 logger 提供了一层封装，它公开了一个对开发人员更友好的接口，您不需要定义要记录的每个字段的每种类型。可从 `logger.Sugar()` 方法中获取，它将稍微减慢并增加日志库的分配数。

与 Go 生态系统中可用的其他包相比，所有这些优化使包的速度相当快，并显著减少了内存分配。让我们浏览并比较一下可用的替代方案。

## 其他选择

Zap 提供的 [基准](https://github.com/uber-go/zap/tree/v1.10.0/benchmarks) 测试清楚地表明 [Zerolog](https://github.com/rs/zerolog) 是与 Zap 竞争最激烈的一个。Zerolog 还提供了结果非常相似的 [基准](https://github.com/rs/logbench) ：

![来自 https://github.com/rs/zerolog 的基准](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20190815-go-how-zap-package-is-optimized/1_M9cZcDqAkoq82Del0TndNQ.png)

它清楚地展示 Zerolog 和 Zap 在性能方面比其他软件包要好得多，速度快 5 到 27 倍。

现在让我们比较一下用 Zerolog 编写的同一段代码：

```go
l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
l.Info().
  Str("url", `http://foo.com`).
  Int("attempt", 3).
  Dur("backoff", time.Second).
  Msg("failed to fetch URL")
```

写法上非常接近，并且我们可以看到 Zerolog 也引入强类型参数以优化性能。如 encoder 接口所述，JSON 编码器还根据类型转储数据：

![zerolog 编码器接口](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20190815-go-how-zap-package-is-optimized/1_aLID1ZKFpryk6IkxOyoWow.png)

发送到日志库的每个条目（ Zerolog 中称为 `event` ）也使用 `sync` 包中的池，以避免在记录消息时进行系统分配。

正如我们所看到的，这些软件包非常相似。这解释了为什么他们的性能很接近。让我们尝试另一个具有不同设计的包，以了解在这些包中缺少的优化。

现在让我们将这些 logger 与 Golang 生态系统中另一个著名的包 [Logrus](https://github.com/sirupsen/logrus) 进行比较。以下是相同功能的代码：

```go
log.SetOutput(os.Stdout)
log.WithFields(log.Fields{
  "url": "http://foo.com",
  "attempt":   3,
  "backoff":   time.Second,
}).Info("failed to fetch URL")
```

在内部，Logrus 还将为 entry 对象使用一个池，但是在检查与消息一起发送的字段时将添加一个反射层。此反射允许日志库检测传递给日志库的所有参数是否有效，但会稍微减慢执行速度。

另外，与 Zap 或 Zerolog 相反，参数不是类型化的，这将导致将起始类型转换为空接口，然后返回起始类型以便对其进行编码。

该包还为钩子添加了一层额外的锁，如果需要，可以将其移除，但默认情况下会激活。

## 没有优化

阅读这些库的编写方式对于每个 Go 开发人员来说都是一个很好的练习，以便了解如何优化我们的代码和潜在的好处。大多数情况下，对于非关键应用程序，您不需要深入研究，但是如果像 Zap 或 Zerolog 这样的外部包免费提供这些优化，我们绝对应该利用它。
如果您想了解使用池的潜在好处，我建议您阅读我的文章“[Understand the design of sync.Pool](https://medium.com/@blanchon.vincent/go-understand-the-design-of-sync-pool-2dde3024e277)”.

---

via: <https://medium.com/a-journey-with-go/go-how-zap-package-is-optimized-dbf72ef48f2d>

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[lts8989](https://github.com/lts8989)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
