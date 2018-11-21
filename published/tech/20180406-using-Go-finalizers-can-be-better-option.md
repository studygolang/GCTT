首发于：https://studygolang.com/articles/16334

# 在 Go 中使用 finalizers 比不使用更好

Go 拥有 finalizers，它们支持程序调用一些代码并作为一个对象来进行垃圾回收。然而，很多人不太喜欢 finalizers，并且通常的建议是完全避免它们 ([比如](https://twitter.com/davecheney/status/790343722865090560))。最近，David Crawshaw 在[《Finalizers 惨案》](https://crawshaw.io/blog/tragedy-of-finalizers) 一文中指出 finalizers 的众多弊端并展示一个依赖他们会导致失败的案例。我差不多同意其上述所有观点，但与此同时，我自己已经在[一个 Go 包访问 Solaris/Illumos kstats](https://github.com/siebenmann/go-kstat) 中使用 finalizers，接下来我将对此用法进行辩护。

我使用 finalizers 来避免人们不正确使用[我的 API](https://github.com/siebenmann/go-kstat/blob/master/kstat-godoc.txt) 时造成不可见的内存泄露。理论上，当你调用我的 jar 包并返回一个神奇的 token，它持有对一些 C-allocated 内存的唯一引用。当你使用此 token 后，你应该调用一个方法关闭它，以此来释放 C-allocated 内存。通常人们会在 API 用法和对象生命周期上犯错。在不使用 finalizer 的情形下，如果一个 token 超过作用范围并在垃圾回收中未被回收，我们将永久泄露此 C-allocated 内存。诸如所有内存和资源泄露此类事情，这将成为一个极其容易忽视且致命的泄露，因为从 Go 标准上它将完全不可见。目前也没有任何一个通用 Go 标准内存泄露工具帮助你解决上述问题（且鉴于 Go 的存在，我认为通用 C 泄露查找工具会产生严重的问题）。

一方面，此处使用一个 finalizer 是一种实用的决定；它能保障人们在使用我的 jar 包时，远离某些用法错误，这些错误会造成一些难以解决的问题。另一方面，我认为此处使用 finalizers 正是 Go 的广泛意义所在。作为一门垃圾回收语言，Go 从本质上决定了管理对象生命周期过于困难，需要大量工作，且太容易犯错。使用 finalizer 完美处理内存问题是存在一定特殊性的，但并不适用于处理除纯粹从实用角度出发之外的任何其他资源。

（与此同时，这些实用角度是真实存在的 ; 正如 David Crawshaw 所言，依赖内存垃圾回收来垃圾收集被耗尽之前的其他资源是极其危险的。这一点甚至对于我的示例在某种程度上也是值得怀疑的，因为 C-allocated 内存并未施压于 Go 垃圾回收器。）

David Crawshaw 跟进发表了一篇文章 -[《锐利的 Go Finalizers》](https://crawshaw.io/blog/sharp-edged-finalizers), 这篇文章中他主张，当人们不能正确使用你的 APIs 时使用 finalizers 来强制恐慌。你可以这么干，但对我来说这有点不像 Go 的风格。总之我认为当且仅当不正确使用你的 API 的后果是特别严重的话，你才应该采用这种方式（比如说，潜在的数据丢失，由于你忘了提交数据库事务然后检查出错误）。

一般而言，我不认为我这种 finalizers 使用方式本身是有意去避免泄漏的。通常从你不再需要这些资源（kstat 令牌，开放文件，或者你拥有的资源）那一刻起，程序将发生内存泄漏，直到 Go 垃圾回收调用你的 finalizer（如果之前是这么干的话），因为这些资源一直都存在，但没有一个在被使用或需要。finalizers 所做的就是让这些泄漏从理论上成为暂时，而不是明确地永久。换句话说，这是一个可修复的泄漏而不是不可修复的泄漏。

PS: 此观点当然不是我原创的。比如说，[非官方 Go FAQ](https://go101.org/article/unofficial-faq.html) 阐述[finalizers 的主要用途](https://go101.org/article/unofficial-faq.html#finalizers)，且有标准库中*os.File finalizer 的例子。

（此观点已经在我脑海中有一段时间了，但 David Crawshaw 的文章提供了一个便捷的提示，而我并没有想到在这种情形下使用 finalizers 来强制出现一个严重的错误。）

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoFinalizersStopLeaks

作者：[Chris Siebenmann](https://utcc.utoronto.ca/~cks/)
译者：[MAXAmbitious](https://github.com/MAXAmbitious)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
