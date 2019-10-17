首发于：https://studygolang.com/articles/24041

# Go 中的并发 bug

Go 目前正在通过新的并发原语（concurrency primitives）goroutine 和 channel 试图简化并发编程并减少报错。但是，实际情况怎么样呢？两位来自宾夕法尼亚州立大学和普渡大学的研究员 [Yiying Zhang](https://www.linkedin.com/in/yiyingzhang) 和 [Linhai Song](https://songlh.github.io/) 对 Go 中的 [并发 bug 在真实场景的情况](https://songlh.github.io/paper/go-study.pdf) 进行了研究。

## 共享内存与 channel 的对比

该研究首先分析了 Go 并发原语在用 Go 构建的大项目中的使用分布情况，包括共享内存类型（Mutex，RWMutex，atomic 和 condition 变量）和新的并发类型 channel：

![共享内存通信仍然是最常见的用法](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-concurrency-bugs-in-go/primitives-usages-over-time.png)

从上图来看，各种并发原语在各个项目中使用分布情况在时间的跨度上基本稳定，因此该研究的结论可能在接下来的几年中仍然有效。基于此，研究员们系统分析了以下流行 Go 项目中的并发 bug 以及它们的原因和修复方式：

![taxonomy](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-concurrency-bugs-in-go/taxonomy.png)

## 研究发现

这项研究发现，大多数开发人员认为 channel 不会像其他常用同步方法那样不可靠，但其实并不是这样：
> 结论2：与通常的看法相反，消息传递比共享内存可能导致更多的阻塞 bug 。希望大家注意消息传递带来的潜在危险，我们也会进一步研究该领域中的错误检测机制。

尽管如此，该研究也表明 channel 的前景很好：
> 发现9：共享内存同步仍然是主要的非阻塞性 bug 的修复方法，但是 channel 不仅被广泛地用于修复与 channel 相关的 bug，而且还用于修复共享内存的 bug。

## 结论

这项研究清楚地表明，新的原语不会减少并发环境中的 bug。但是，如果更好地了解 channel 和 goroutines，可以避免大多数这类问题：
> 发现6：我们研究中的大多数阻塞 bug（传统的共享内存 bug 和消息传递 bug）都可以通过简单的解决方案来解决。

例如 Docker 中的这个问题：

![data-race](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-concurrency-bugs-in-go/data-race.png)

运行 ```go vet``` 命令可以避免上面这种错误。

## 对Go生态系统的贡献

在研究期间，研究员们还发现有可能开发出更好的静态代码检查器：
> 结论3： 由于 Go 中的阻塞 bug 的原因与其修复方法密切相关，并且修复方法不复杂，开发出全自动或半自动化的工具来修复 Go 的阻塞 bug 是很有希望的。

根据 Go 社区多年来的经验，运行时的检测工具也可以得到改进：
> 结论4：仅仅用死锁检测器在运行时检测 Go 阻塞 bug 不是很有效。未来的研究应侧重于构建新颖的阻塞 bug 检测技术，例如，结合使用静态和动态阻塞模式的检测。

基于以上这些发现，项目负责人 Ziheng Liu 带领研究员们开始研究新的代码分析器。以 [SSA](https://godoc.org/golang.org/x/tools/go/ssa) 软件包为基础，该新式代码分析器通过在 GraphQL  Go 项目中发现了 [bug](https://github.com/graphql-go/graphql/pull/434) 证明了它的有效性。这个新式代码分析器预计于2020年1月发布。

有关该研究的更多信息，请访问该研究文献 [Github](https://songlh.github.io/paper/go-study.pdf) 地址。

---

via: https://medium.com/a-journey-with-go/go-concurrency-bugs-in-go-7d3677a1f2a2

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[zhiyu-tracy-yang](https://github.com/zhiyu-tracy-yang)
校对：[zhoudingding](https://github.com/dingdingzhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出