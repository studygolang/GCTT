# Go：多个错误管理
![由Renee French创作的原始Go Gopher制作的“ Go的旅程”插图。](https://github.com/studygolang/gctt-images2/blob/master/20200901-Go-Multiple-Errors-Management/Illustration.png?raw=true)

Go 语言中的错误（error）管理总是能引起争论，同时，在关于使用 Go 语言的时候，开发者面对最大的挑战的[年度调查](https://blog.golang.org/survey2019-results)中也是一个经常性的话题。然而，在并发环境处理 error 的场景下，或者在同一个 goroutine 中合并多个错误的场景下，Go 提供了很不错的包可以让多个错误的处理变得简单：来看看如何合并由单个 goroutine 生成的多个 error。

## 一个 goroutine，多个 error
当编写有着重试策略的代码时，将多个 error 合并为一个会十分有用，比如，下面是我们需要收集生成的 error 的一个基本例子：

![](https://github.com/studygolang/gctt-images2/blob/master/20200901-Go-Multiple-Errors-Management/a-basic-example.png?raw=true)

这个程序读取并解析一个 CSV 文本，并且展示发现的错误。如果将 error 聚合为一个完整的报告，会更加方便。为了将错误合并为一个，我们可以在两个不错的包中进行选择：

- 使用 [HashiCorp](https://github.com/hashicorp) 的 [go-multierror](https://github.com/hashicorp/go-multierror) ，error 可以被合并为一个标准 error：

![](https://github.com/studygolang/gctt-images2/blob/master/20200901-Go-Multiple-Errors-Management/Using-go-multierror.png?raw=true)

之后可以打印出一个报告：

![](https://github.com/studygolang/gctt-images2/blob/master/20200901-Go-Multiple-Errors-Management/a-report.png?raw=true)

- 使用 [Uber](https://github.com/uber-go) 的 [multierr](https://github.com/uber-go/multierr)：

这里的实现是类似的，这是输出：

![](https://github.com/studygolang/gctt-images2/blob/master/20200901-Go-Multiple-Errors-Management/Using-multierr.png?raw=true)

error 通过分号连接，没有经过其他格式化。

关于两个包的性能，这是一个使用相同程序，有着更高次数失败的基准测试：

```
name                    time/op         alloc/op        allocs/op
HashiCorpMultiErrors-4  6.01µs ± 1%     6.78kB ± 0%     77.0 ± 0%
UberMultiErrors-4       9.26µs ± 1%     10.3kB ± 0%      126 ± 0%
```

Uber 的实现略慢，同时消耗更多内存。但是，这个包被设计为一次将错误聚合在一起，而不是每次都追加它们。在聚合 error 的时候，结果是接近的。但是由于需要额外步骤，代码有点不太优雅。这是新的结果：

```
name                    time/op         alloc/op        allocs/op
HashiCorpMultiErrors-4  6.01µs ± 1%     6.78kB ± 0%     77.0 ± 0%
UberMultiErrors-4       6.02µs ± 1%     7.06kB ± 0%     77.0 ± 0%
```

两个包都通过在自定义实现中实现了 `Error() string` 函数的方式利用了 Go 的 `error` 接口。

## 一个 error，多个 goroutine
在操作多个 goroutine 来处理一个任务的时候，为了保证程序的正确性，正确地管理结果和错误汇总是有必要的。

以一个程序开始，该程序使用多个 goroutine 执行一系列行为（action）；每个行为持续一秒：

![](https://github.com/studygolang/gctt-images2/blob/master/20200901-Go-Multiple-Errors-Management/use-multiple-goroutines-to-perform-a-series-of-actions.png?raw=true)

为了描绘 error 传播，第三个 goroutine 的第一个 action 会失败。这是发生的事情：

![](https://github.com/studygolang/gctt-images2/blob/master/20200901-Go-Multiple-Errors-Management/illustrate-the-error-propagation.png?raw=true)

如同预期的一样，这个程序大致用了三秒钟，因为大多数 goroutine 需要经历三个 action，每一个需要一秒：

```
go run .  0.30s user 0.19s system 14% cpu 3.274 total
```

然而，我们可能希望使 goroutine 之间相互依赖，并且如果其中一个失败就取消他们。避免无谓工作的解决方案可以是加一个 context，并且，一旦一个 goroutine 失败，就会取消它：

![](https://github.com/studygolang/gctt-images2/blob/master/20200901-Go-Multiple-Errors-Management/avoid-unnecessary-work.png?raw=true)

这恰好就是 [`errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup?tab=doc) 所提供的；当处理一组 goroutine 的时候，一个错误以及上下文传播。这是使用 [`errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup?tab=doc) 包的新代码：

![](https://github.com/studygolang/gctt-images2/blob/master/20200901-Go-Multiple-Errors-Management/using-the-package-errgroup.png?raw=true)

由于通过 error 传播了取消的上下文，这个程序现在运行地更快了：

```
go run .  0.30s user 0.19s system 38% cpu 1.269 total
```

这个包所带来的其他好处是，我们不需要再操心等待组的增加以及将 goroutine 标记为已完成。这个包为我们管理了这些，我们仅仅需要说明什么时候我们准备好了等待过程的结束。

---
via: https://medium.com/a-journey-with-go/go-multiple-errors-management-a67477628cf1

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[dust347](https://github.com/dust347)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
