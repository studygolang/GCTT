首发于：https://studygolang.com/articles/15974

# 宏观看 Go 语言中的 Map 内部

网上有很多涉及 slice 内部的文章，相比之下深入探讨 map 的文章非常稀少，我非常好奇为什么会这样，就去找了这份能深入了解 map 的源码。

https://golang.org/src/runtime/map.go（译者注：因为最新 1.11 版本变更了文件名，所以链接修改为最新的地址。）

这些代码对于我来说很复杂，但是我觉得我们可以用一种宏观的形式去理解 map 是如何构建以及增长。这种方式也许可以解释 map 为什么无序，高效和快速。

**创建和使用 Map**

让我们来看一下如何创建一个 map 实例然后储存几条数据：

```go
// 创建一个空 map，key 和 value 都是 string 类型。
colors := map[string]string{}

// 向 map 中增加几个键值对
colors["AliceBlue"] = "#F0F8FF"
colors["Coral"]     = "#FF7F50"
colors["DarkGray"]  = "#A9A9A9"
```

当我们向 map 中增加 value 时总是需要指定一个 key 来进行关联，关联之后用这个 key 就可以直接找到相应的 value 而不用去遍历整个集合。

```go
fmt.Printf("Value: %s", colors["Coral"])
```

当我们在遍历 map 的时候，所获得 key 的顺序并不是原来插入的顺序，事实上，我们每次运行下面的代码，key 的顺序都会改变。

```go
colors := map[string]string{}
colors["AliceBlue"]   = "#F0F8FF"
colors["Coral"]       = "#FF7F50"
colors["DarkGray"]    = "#A9A9A9"
colors["ForestGreen"] = "#228B22"
colors["Indigo"]      = "#4B0082"
colors["Lime"]        = "#00FF00"
colors["Navy"]        = "#000080"
colors["Orchid"]      = "#DA70D6"
colors["Salmon"]      = "#FA8072"

for key, value := range colors {
    fmt.Printf("%s:%s, ", key, value)
}
```

```shell
Output:
AliceBlue:#F0F8FF, DarkGray:#A9A9A9, Indigo:#4B0082, Coral:#FF7F50,
ForestGreen:#228B22, Lime:#00FF00, Navy:#000080, Orchid:#DA70D6,
Salmon:#FA8072
```

现在我们已经知道了如何创建，设置键值对（key/value pairs）并且遍历整个 map，接下来让我们去一窥它的真相。

**Map 是如何构建的**

在 Go 语言中 Map 是以散列表（hash table）的形式实现的，如果你想了解一下散列表是什么，网上有许多相关的文章，你可以将这篇 Wikipedia 可以作为起点：

http://en.wikipedia.org/wiki/Hash_table

Go 语言中 map 的散列表是由一组 bucket 构建而成，bucket 的数量会等于 2 的某次方。当一个 map 操作被执行时会根据 key 的名字来生成一个散列 key，比如（`colors["Black"] = "#000000"`）会根据字符串 “ Black ” 来生成散列 key，根据这个散列 key 的低阶位（LOB）来选择放入哪个 bucket 中。

![Screen Shot](https://raw.githubusercontent.com/studygolang/gctt-images/master/Macro-View-of-Map-Internals-In-Go/Screen%2BShot%2B2013-12-31%2Bat%2B6.35.43%2BPM.png)

一旦确定了 bucket，那么就可以对键值对进行相应的操作，比如储存、删除或查找。如果我们观察 bucket 的内部，那么会发现两个结构体。首先是一个数组，它从之前用来选择 bucket 的散列 key 中获取 8 个高阶位（HOB），这个数组区分了每一个被储存在 bucket 中的键值对，然后是一个储存键值对内容的 byte 数组，这个数组把键值对结合起来并储存到所在的 bucket 中。

![Screen Shot](https://raw.githubusercontent.com/studygolang/gctt-images/master/Macro-View-of-Map-Internals-In-Go/Screen%2BShot%2B2013-12-31%2Bat%2B7.01.15%2BPM.png)

当我们迭代一个 map 时，迭代器会访问整个 bucket 的数组然后按顺序取出相应的键值对，这就是为什么 map 是无序集合的原因。这些 hash key 能决定 map 的访问顺序是因为它们决定了每一个键值最终储存在哪个 bucket。

**内存和 bucket 溢出**

把键值对整合起来然后看上去像是一个单独的 byte 数组是有原因的，如果把 key 和 values 按 key/value/key/value 这样存放的话，那么每一个键值对的内存分配需要保持适当的内存对齐，下面举个例子：

```go
map[int64]int8
```

这个 map 每个键值对的 value 都只占用 1 个 byte，却需要 7 个额外的 byte 来填补对齐的内存空间，如果把键值对按 key/key/value/value 方式的话，key 和 value 就只需要加到的各自的尾部，这样就消除了对齐所浪费的大量空间。下面的文章提供了更多关于内存对齐的更多知识：

http://www.goinggo.net/2013/07/understanding-type-in-go.html (GCTT 译文：https://studygolang.com/articles/13976)

一个 bucket 被设定为只储存 8 个键值对，当向一个已满的 bucket 插入 key 时，就会创建出一个新的 bucket 和先前的 bucket 关联起来，并将 key 加入到这个新的 bucket 中。

![Screen Shot](https://raw.githubusercontent.com/studygolang/gctt-images/master/Macro-View-of-Map-Internals-In-Go/Screen%2BShot%2B2013-12-31%2Bat%2B7.12.06%2BPM.png)

**Map 是如何增长的**

当我们从 map 中持续增加或者删除键值对时，map 的查找效率就会降低。hash map 增长的时机由装载阈值（load threshold values）基于下面四个因素来确定：

```go
% overflow  : 已满的 bucket 在所有 bucket 中的所占比例
bytes/entry : 每个键值对的额外字节使用数量
hitprobe    : 寻找一个 key 所需要检查的项数量
missprobe   : 寻找一个不存在的 key 所需要检查的项数量
```

我们当前演示代码的装载阈值如下：

| **LOAD** | **%overflow** | **bytes/entry** | **hitprobe** | **missprobe** |
| -------- | ------------- | --------------- | ------------ | ------------- |
| 6.50     | 20.90         | 10.79           | 4.25         | 6.50          |

hash table 在开始增长时会先将名叫 “ old bucket ” 的指针指向当前的 bucket 数组，然后会分配一个比原来 bucket 大两倍的新 bucket 数组，这可能会涉及到大量的内存分配，不过这些分配的内存并不会马上进行初始化。
当新的 bucket 数组内存可用时，旧的 bucket 数组中的键值对会被移动或者迁移到新的 bucket 数组中。迁移一般在 map 中的键值对增加或者删除时产生，在旧的 bucket 中作为一个整体的键值对可能会被移动到不同的新 bucket 数组中，迁移算法会让这些键值对均匀地分配。

[![Screen Shot](https://raw.githubusercontent.com/studygolang/gctt-images/master/Macro-View-of-Map-Internals-In-Go/Screen%2BShot%2B2013-12-31%2Bat%2B7.22.39%2BPM.png)](https://github.com/studygolang/gctt-images/blob/master/Screen+Shot+2013-12-31+at+7.22.39+PM.png?raw=true)

迭代器在数据迁移期间依然需要遍历旧 bucket 中的数据，同时迁移还会影响遍历时键值对的返回方式，所以说这是一个非常优雅的处理方式，为了确保在 map 增长和扩展时迭代器能正常工作是需要花费大量精力的。

**结论**

就如我在文章开始时说的那样，这只是从宏观视角去了解 map 的构造以及增长，这些代码使用 C 编写（译者注：这只是作者当时的情况，现在都是 Go 写的）并且使用大量的内存和指针操作来保持 map 的快速，效率以及安全。

很明显，目前的实现方式可能随时会改变，但这并不影响我们使用 map 的方式。如果你事先知道需要用到多少 key，那么最好在初始化的时候就分配好这些空间，同时这也解释了为什么 map 是无序集合，为什么在遍历时迭代器看上去像是随机选择数据。

**特别感谢**

在这里我要感谢 Stephen McQuay 和 Keith Randall 对本文的审阅和错误纠正。

---

via: https://www.ardanlabs.com/blog/2013/12/macro-view-of-map-internals-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[Maxwell Hu](https://github.com/maxwell365)
校对：[Unknwon](https://github.com/Unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
