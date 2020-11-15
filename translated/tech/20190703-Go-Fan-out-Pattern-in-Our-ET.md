# Go: 在我们的 ETL 中使用扇出模式（Go: Fan-out Pattern in Our ETL）

![](https://miro.medium.com/max/1400/1*r7OWMQAx6G2AGoKTXmHM6A.png)

Go 语言在构建微服务、特别是有使用 gRPC 的应用中，非常地流行，其实在构建命令行程序时也是特别地好用。为了学习扇出模式，我会基于我们公司使用 ETL 的例子，来介绍这个模式。

## ETL

ETL（提取（Extract），转换（Transform），加载（Load））通常都需要处理大量的数据。在这样的场景下，有一个好的并发策略对于 ETL 来说，能够带来巨大的性能提升。

ETL 中有两个最重要的部分是提取（extracting）和加载（Load），通常它们都跟数据库有关，瓶颈通常也属于老生常谈的话题：网络带宽，查询性能等等。基于我们要处理的数据以及瓶颈所在，两种模式对于处理数据或者处理输入流的编码和解码过程中，非常有用。

## 扇入扇出模式（Fan-in, fan-out pattern）

扇入和扇出模式在并发场景中能得到较大的好处。这里将对它们逐个做专门的介绍（review）：

扇出，在 GO 博客中这样定义：

多个函数能够同时从相同的 channel 中读数据，直到 channel 关闭。

这种模式在快速输入流到分布式数据处理中，有一定的优势：

![fan-out pattern with distributed work](https://miro.medium.com/max/754/1*pIc5IqD0qd6uPRVlrdzk5g.png)

扇入，在 Google 这样定义：

一个函数可以从多个输入中读取，并继续操作，直到所有 channel 所关联的输入端，都已经关闭。

这种模式，在有多个输入源，且需要快速地数据处理中，有一定的优势：

![fan-in pattern with multiple inputs](https://miro.medium.com/max/754/1*WAoD1zJQfzvGENSJPmd6SA.png)

## 在实际中使用扇出模式（Fan-out in action）

在我们的项目中，我们需要处理存储在 CSV 文件的大量数据，它们加载后，将在 elastic 中被检索。输入的处理必须快，否则（阻塞加载）加载就会变得很慢。因此，我们需要比输入生成器更多的数据处理器。扇出模式在这个例子中，看起来非常适合：

![](https://miro.medium.com/max/734/1*8K93LJ5t-4z8nEXoZWQdJg.png)

下面是我们的伪代码：

```bash
Variables:
data chan
Start:
// a goroutine will parse the CSV and will send it to the channel
ParseCSV(data<-)
// a goroutine is started for each workers, defined as command line arguments
For each worker in workers
    Start goroutine
        For each value in <-data
            Insert the value in database by chunk of 50
Wait for the workers
Stop
```

输入和加载程序是并发执行的，我们不需要等到解析完成后再开始启动数据处理程序。

这种模式让我们可以单独考虑业务逻辑的同时，还可以使用（Go）并发的特性。几个工作器之间原生的分布式负载能力，有助于我们解决此类过程中的峰值负载问题。

---
via: https://medium.com/a-journey-with-go/go-fan-out-pattern-in-our-etl-9357a1731257

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[译者ID](https://github.com/gogeof)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
