# 深入剖析 Golang Pprof 标签

Polar Signals 提供的持续分析工具可以和任何格式的 pprof 分析适配，Go 深度集成了 [pprof](https://github.com/google/pprof) 甚至支持了它的`标签`特性。
然而，自从我们发布了我们的 [持续分析产品](https://www.polarsignals.com/blog/posts/2021/02/09/announcing-polar-signals-continuous-profiler/) 之后，
收到了很多工程师的反馈，发现许多工程师不知道如何去分析， 或者不知道分析能给他们带来什么好处。这篇文章主要剖析 pprof 标签，并会结合一些 Go 的示例代码去分析。

## 基础

pprof 标签只支持 Go 的 CPU 分析器。Go 的分析器是抽样分析，这意味着它只会根据特定的频率（默认是 1 秒钟 100 次）去获取执行中函数的调用栈并记录。
简单来说，开发者如果使用标签，在分析器取样时就可以将函数的调用栈进行区分，然后只聚合i具有相同标签的函数调用栈。

Go 在 `runtime/pprof` 包中已经支持了标签检测，可以使用 [`pprof.Do`](https://golang.org/pkg/runtime/pprof/#Do) 函数非常方便的使用。

```golang
pprof.Do(ctx, pprof.Labels("label-key", "label-value"), func (ctx context.Context) {
    // execute labeled code
})
```

## 进阶

为了进行演示如何使用 pprof
标签，我们创建了一个包含许多示例的仓库，这个示例仓库代码作为这篇文章的内容指导。
仓库地址：[https://github.com/polarsignals/pprof-labels-example](https://github.com/polarsignals/pprof-labels-example)

示例代码的 [`main.go`](https://github.com/polarsignals/pprof-labels-example/blob/60accf8b4fbebcd5f96b3743663af5745ef74596/main.go)
通过将 `tanant` 传递给 `iterate` 函数实现了大量的 for 循环， 其中`tanant1` 做了 10 亿次循环，而 `tanant2` 做了 1 亿次循环，
同时会记录 CPU 的分析日志并将其写入 `./cpuprofile.pb.gz`。

为了演示如何在 pprof 的分析日志中展示 pprof 标签，
用 [`printprofile.go`](https://github.com/polarsignals/pprof-labels-example/blob/60accf8b4fbebcd5f96b3743663af5745ef74596/printprofile.go)
来打印每次抽样函数调用栈以及样本值，还有收集到样本的标签。

如果我们注释掉 [`pprof.Do` 的这部分](https://github.com/polarsignals/pprof-labels-example/blob/60accf8b4fbebcd5f96b3743663af5745ef74596/main.go#L39-L41) ，
我们将无法进行标签检测，运行 [`printprofile.go`](https://github.com/polarsignals/pprof-labels-example/blob/60accf8b4fbebcd5f96b3743663af5745ef74596/printprofile.go)
代码，让我们看看没有标签的抽样分析结果：

```bash
runtime.main
main.main
main.iteratePerTenant
main.iterate
2540000000
---
runtime.main
main.main
main.iteratePerTenant
main.iterate
250000000
---
Total:  2.79s
```

CPU 分析数据的单位是 纳秒，所以这些抽样总共花费时间是 2.79 秒（2540000000ns + 250000000ns = 2790000000ns = 2.79s）。

同样的，现在当每次调用 `iterate` 时添加标签，用 pprof 分析，这些数据看起来就不太一样，打印出带有标签的抽样分析结果：

```bash
runtime.main
main.main
main.iteratePerTenant
runtime/pprof.Do
main.iteratePerTenant.func1
main.iterate
10000000
---
runtime.main
main.main
main.iteratePerTenant
runtime/pprof.Do
main.iteratePerTenant.func1
main.iterate
2540000000
map[tenant:[tenant1]]
---
runtime.main
main.main
main.iteratePerTenant
runtime/pprof.Do
main.iteratePerTenant.func1
main.iterate
10000000
map[tenant:[tenant1]]
---
runtime.main
main.main
main.iteratePerTenant
runtime/pprof.Do
main.iteratePerTenant.func1
main.iterate
260000000
map[tenant:[tenant2]]
---
Total:  2.82s
```

将所有抽样加起来总共花费了 2.82 秒，然而，因为调用 `iterate` 时，我们添加了标签，所以我们能在结果中区分哪个 `tenant` 导致了更多的 CPU 占用 。
现在我们可以看到 `tenant1` 花费了总时间 2.82 秒中的 2.55 秒（2540000000ns + 10000000ns = 2550000000ns = 2.55s）。

让我们看看抽样的原始日志（还有它们的元数据），去更深入理解一下它们的格式：

```bash
$ protoc --decode perftools.profiles.Profile --proto_path ~/src/github.com/google/pprof/proto profile.proto < cpuprofile.pb | grep -A12 "sample {"
sample {
  location_id: 1
  location_id: 2
  location_id: 3
  location_id: 4
  location_id: 5
  value: 1
  value: 10000000
}
sample {
  location_id: 1
  location_id: 2
  location_id: 3
  location_id: 4
  location_id: 5
  value: 254
  value: 2540000000
  label {
    key: 14
    str: 15
  }
}
sample {
  location_id: 6
  location_id: 2
  location_id: 3
  location_id: 4
  location_id: 5
  value: 1
  value: 10000000
  label {
    key: 14
    str: 15
  }
}
sample {
  location_id: 1
  location_id: 2
  location_id: 3
  location_id: 7
  location_id: 5
  value: 26
  value: 260000000
  label {
    key: 14
    str: 16
  }
}
```

我们可以看到每个抽样都由许多 ID 组成，这些 ID 指向它们在分析日志 `location` 数组中的位置，除了这些 ID 还有几个值 。
仔细看下 `printprofile.go` 程序，你会发现它使用了每个抽样的最后一个抽样 value。
实际上，Go 的 CPU 分析器会记录两个 value，第一个代表这个调用栈在一次分析区间被记录样本的数量，第二个代表它花费了多少纳秒。
pprof 的定义描述当没设置 `default_sample_type` 时（在 Go 的 CPU 配置中设置），就使用所有 value 中的最后一个，
因此我们使用的是代表纳秒的 value 而不是样本数的 value。最后，我们可以打印出标签，它是 pprof 定义的一个以字符串组成的字典。

最后，因为用标签去区分数据，我们可以让可视化界面更直观。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210413-Demystifying-Pprof-Labels-with-Go/pprof-callgraph-with-labels.png)

你可以在 Polar Signals 网站去更详细的了解上面的这次分析：[https://share.polarsignals.com/2063c5c/](https://share.polarsignals.com/2063c5c/)

## 结论

pprof 标签是帮助我们理解程序不同执行路径非常有用的方法，许多人喜欢在多租户系统中使用它们，目的就是为了能够定位在他们系统中出现的由某一个租户导致的性能问题。
就像开头说的，只需要调用 [`pprof.Do`](https://golang.org/pkg/runtime/pprof/#Do) 就可以了。

Polar Signals 提供的持续分析工具也支持了 pprof 标签的可视化界面和报告，
如果你想参与个人体验版请点击：[申请资格](https://www.polarsignals.com/#request-access)

---
via: https://www.polarsignals.com/blog/posts/2021/04/13/demystifying-pprof-labels-with-go/

作者：[Frederic Branczyk](https://twitter.com/fredbrancz)
译者：[h1z3y3](https://h1z3y3.me)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
