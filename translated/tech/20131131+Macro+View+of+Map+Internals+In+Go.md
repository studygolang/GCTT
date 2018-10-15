# 宏观Go语言中的Map接口

网上有很多涉及slice内部的文章，相比之下深入探讨map的文章非常稀少，我非常好奇为什么会这样，就去找了这份能深入了解map的源码。

https://golang.org/src/runtime/map.go（译者注：因为最新1.11版本变更了文件名，所以链接修改为最新的地址。）

这些代码对于我来说很复杂，但是我觉得我们可以用一种宏观的形式去理解map是如何构建以及增长。这种方式也许可以解释map为什么无序，高效和快速。

**创建和使用Map**

让我们来看一下如何创建一个map实例然后储存几条数据：

``` go
// 创建一个空map，key和value都是string类型。
colors := map[string]string{}

// 向map中增加几个键值对
colors["AliceBlue"] = "#F0F8FF"
colors["Coral"]     = "#FF7F50"
colors["DarkGray"]  = "#A9A9A9"
```

当我们向map中增加value时总是需要指定一个key来进行关联，关联之后用这个key就可以直接找到相应的value而不用去遍历整个集合。

```go
fmt.Printf("Value: %s", colors["Coral"])
```

当我们在遍历map的时候，所获得key的顺序并不是原来插入的顺序，事实上，我们每次运行下面的代码，key的顺序都会改变。

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

现在我们已经知道了如何创建，设置键值对（key/value pairs）并且遍历整个map，接下来让我们去一窥它的真相。

**Map是如何构建的**

在Go语言中Map是以散列表（hash table）的形式实现的，如果你想了解一下散列表是什么，网上有许多相关的文章，你可以将这篇Wikipedia可以作为起点：

http://en.wikipedia.org/wiki/Hash_table

Go语言中map的散列表是由一组bucket构建而成，bucket的数量会等于2的某次方。当一个map操作被执行时会根据key的名字来生成一个散列key，比如（`colors["Black"] = "#000000"`）会根据字符串 “Black” 来生成散列key，根据这个散列key的低阶位（LOB）来选择放入哪个bucket中。

[![Screen Shot](https://raw.githubusercontent.com/studygolang/gctt-images/master/Macro-View-of-Map-Internals-In-Go/Screen%2BShot%2B2013-12-31%2Bat%2B6.35.43%2BPM.png)](https://github.com/studygolang/gctt-images/blob/master/Macro-View-of-Map-Internals-In-Go/Screen+Shot+2013-12-31+at+6.35.43+PM.png?raw=true)

一旦确定了bucket，那么就可以对键值对进行相应的操作，比如储存、删除或查找。如果我们观察bucket的内部，那么会发现两个结构体。首先是一个数组，它从之前用来选择bucket的散列key中获取8个高阶位（HOB），这个数组区分了每一个被储存在bucket中的键值对，然后是一个储存键值对内容的byte数组，这个数组把键值对结合起来并储存到所在的bucket中。

[![Screen Shot](https://raw.githubusercontent.com/studygolang/gctt-images/master/Macro-View-of-Map-Internals-In-Go/Screen%2BShot%2B2013-12-31%2Bat%2B7.01.15%2BPM.png)](https://github.com/studygolang/gctt-images/blob/master/Screen+Shot+2013-12-31+at+7.01.15+PM.png?raw=true)

当我们迭代一个map时，迭代器会访问整个bucket的数组然后按顺序取出相应的键值对，这就是为什么map是无序集合的原因。这些hash key能决定map的访问顺序是因为它们决定了每一个键值最终储存在哪个bucket。

**内存和bucket溢出**

把键值对整合起来然后看上去像是一个单独的byte数组是有原因的，如果把key和values按key/value/key/value这样存放的话，那么每一个键值对的内存分配需要保持适当的内存对齐，下面举个例子：

```go
map[int64]int8
```

这个map每个键值对的value都只占用1个byte，却需要7个额外的byte来填补对齐的内存空间，如果把键值对按key/key/value/value方式的话，key和value就只需要加到的各自的尾部，这样就消除了对齐所浪费的大量空间。下面的文章提供了更多关于内存对齐的更多知识：

http://www.goinggo.net/2013/07/understanding-type-in-go.html

一个bucket被设定为只储存8个键值对，当向一个已满的bucket插入key时，就会创建出一个新的bucket和先前的bucket关联起来，并将key加入到这个新的bucket中。

[![Screen Shot](https://raw.githubusercontent.com/studygolang/gctt-images/master/Macro-View-of-Map-Internals-In-Go/Screen%2BShot%2B2013-12-31%2Bat%2B7.12.06%2BPM.png)](https://github.com/studygolang/gctt-images/blob/master/Screen+Shot+2013-12-31+at+7.12.06+PM.png?raw=true)

**Map是如何增长的**

当我们从map中持续增加或者删除键值对时，map的查找效率就会降低。hash map增长的时机由装载阈值（load threshold values）基于下面四个因素来确定：

```go
% overflow  : 已满的bucket在所有bucket中的所占比例
bytes/entry : 每个键值对的额外字节使用数量
hitprobe    : 寻找一个key所需要检查的项数量
missprobe   : 寻找一个不存在的key所需要检查的项数量
```

我们当前演示代码的装载阈值如下：

| **LOAD** | **%overflow** | **bytes/entry** | **hitprobe** | **missprobe** |
| -------- | ------------- | --------------- | ------------ | ------------- |
| 6.50     | 20.90         | 10.79           | 4.25         | 6.50          |

hash table 在开始增长时会先将名叫 “old bucket” 的指针指向当前的bucket数组，然后会分配一个比原来bucket大两倍的新 bucket 数组，这可能会涉及到大量的内存分配，不过这些分配的内存并不会马上进行初始化。
当新的 bucket 数组内存可用时，旧的 bucket 数组中的键值对会被移动或者迁移到新的bucket数组中。迁移一般在map中的键值对增加或者删除时产生，在旧的 bucket 中作为一个整体的键值对可能会被移动到不同的新bucket数组中，迁移算法会让这些键值对均匀地分配。

[![Screen Shot](https://raw.githubusercontent.com/studygolang/gctt-images/master/Macro-View-of-Map-Internals-In-Go/Screen%2BShot%2B2013-12-31%2Bat%2B7.22.39%2BPM.png)](https://github.com/studygolang/gctt-images/blob/master/Screen+Shot+2013-12-31+at+7.22.39+PM.png?raw=true)

迭代器在数据迁移期间依然需要遍历旧 bucket 中的数据，同时迁移还会影响遍历时键值对的返回方式，所以说这是一个非常优雅的处理方式，为了确保在map增长和扩展时迭代器能正常工作是需要花费大量精力的。

**结论**

就如我在文章开始时说的那样，这只是从宏观视角去了解map的构造以及增长，这些代码使用C编写（译者注：这只是作者当时的情况，现在都是Go写的）并且使用大量的内存和指针操作来保持map的快速，效率以及安全。

很明显，目前的实现方式可能随时会改变，但这并不影响我们使用map的方式。如果你事先知道需要用到多少key，那么最好在初始化的时候就分配好这些空间，同时这也解释了为什么map是无序集合，为什么在遍历时迭代器看上去像是随机选择数据。

**特别感谢**

在这里我要感谢 Stephen McQuay 和 Keith Randall 对本文的审阅和错误纠正。

---

via: https://www.ardanlabs.com/blog/2013/12/macro-view-of-map-internals-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[Maxwell Hu](https://github.com/maxwell365)
校对：[Unknwon](https://github.com/Unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出