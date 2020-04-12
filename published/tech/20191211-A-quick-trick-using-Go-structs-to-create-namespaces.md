首发于：https://studygolang.com/articles/26300

# Go 中的黑桃 A：使用结构体创建命名空间

假设，但不是凭空想象，在你的程序中，你注册了一堆 [expvar 包的统计变量](https://golang.org/pkg/expvar/)，用来在暴露出去的 JSON 结果中能有一个容易辨识的名字。普通的实现方式下，你可能会有一大堆全局变量，对应着程序追踪的各种信息。这些变量与其他的全局变量混成一团，这毫无美感，如果我们能规避这种情况，那么事情会变得不那么糟糕。

归功于 Go 对匿名结构类型的支持，我们可以实现。我们可以基于匿名结构类型创建一个变量集合的命名空间：

```go
var events struct {
    connections, messages [expvar.Int](expvar.Int)

    tlsconns, tlserrors [expvar.Int](expvar.Int)
}
```

在我们的代码中，我们可以使用 `events.connects` 等等，而不是必须用一些很糟糕或容易引起歧义的名字。

我们也可以在全局等级范围外用这种方法。你可以在这种命名空间结构内把任意的变量名集合隔离开。一个例子就是把计数变量嵌入到另一个结构体中：

```go
type ipMap struct {
    sync.Mutex
    ips map[string]int
    stat struct {
        Size, Adds, Lookups, Dels int
    }
}
```

原因很明显，这对于不需要进行初始化的变量类型是最好的解决方案；其他的变量类型需要进行一些初始化，这稍微有一点点笨重。

这可能不合某些人的口味，我也不知道这在 Go 中是不是好的做法。我的个人观点是，我与其用前缀 `prefix_` 来隔离变量名，不如用前缀 `prefix.` ，尽管人为地引入了这样的匿名结构体。但是一些人可能会有不同看法。即使在我看来，它也确实是侵入较大的修改，但可能它是合法的。

(为了更加明确地统计计数信息，[这也是一种方便地暴露所有信息的方法](https://utcc.utoronto.ca/~cks/space/blog/programming/GoExpvarNotes))

出于好奇我快速浏览了当前开发中的 Go 编译器和标准库，隐约在几处地方发现了疑似使用这种方式的地方。并不是所有的使用方式都一样，所以看了源码后我要说的重点是，这似乎也不是一种完全不可容忍或被原作者反对的观点。

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoStructsForNamespaces

作者：[Chris Siebenmann](https://utcc.utoronto.ca/~cks/)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
