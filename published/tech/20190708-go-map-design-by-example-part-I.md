首发于：https://studygolang.com/articles/22773

# Go: 通过例子学习 Map 的设计 — Part I

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-map-design-by-example-part-I/1_HFwlzPXZ1nXxKMwumjrlmQ.png)

本文是三篇系列文章中的第一篇。每篇文章都将涵盖 map 的不同部分。我建议你按顺序阅读。

Go 提供的内置类型 `map` 是使用[哈希表](https://en.wikipedia.org/wiki/Hash_table) 实现的。在本文中，我们将探讨这个哈希表的不同部分的具体实现：桶（存储键值对的数据结构），哈希（键值对的索引），负载因子（判断 map 是否应该扩容的指标）。

## 桶

Go 将键值对存储在桶的列表中，每个桶容纳 8 个键值对，当 map 的容量耗尽，哈希桶的数量将会翻倍。下面是持有 4 个桶的 map 的粗略示图：

![map with a list of buckets](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-map-design-by-example-part-I/1_MvMVl9YfpWVzM7nqjzON9w.png)

在下一篇文章中，我们将看到这些键值对是如何在桶里存储的。如果 map 再一次扩容，桶的数量将会翻倍，依次增加到 8，16，等等。

当一个键值对存入 map，它将根据键计算出的哈希值，被分配到一个桶里。

## 哈希

当一个键值对被存放到 map 中，Go 会根据它的键生成哈希值。

让我们以键值对 `"foo" = 1` 的插入作为例子。生成的哈希值可能是 `15491954468309821754`。该值将应用于位操作，掩码对应于桶的数量减 1。在我们的 4 个桶的例子中，掩码是 3，位操作如下：

![value dispatched in the buckets](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-map-design-by-example-part-I/1_OgOgEvcqNALd-IHXCSeofw.png)

哈希值不仅用于值在桶中的分配，还参与其他的操作。每个桶都将其哈希值的首字节存储在一个内部数组中，这使得 Go 可以对键进行索引，并跟踪桶中的空槽。让我们看一下二进制表示下，哈希的例子：

![top hash table in bucket](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-map-design-by-example-part-I/1_z8YVGw6WANXuW-xboHmPfQ.png)

多亏了桶内部被称为 *top hash* 的表，Go 可以在数据访问期间使用它们与请求键的哈希值进行比较。

根据我们在程序中对 map 的使用，Go 需要对 map 进行扩容，以便管理更多的键值对。

## Map 扩容

在存储键值对时，桶会将它存储在 8 个可用的插槽中。如果这些插槽全部不可用，Go 会创建一个溢出桶，并于当前桶连接。

![overflow bucket](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-map-design-by-example-part-I/1_ZfDObIafsML18crqW-MX_Q.png)

这个 `overflow` 属性表明了桶的内部结构。然而，增加溢出桶会降低 map 的性能。作为弥补，Go 将会分配新的桶（当前桶的数量的两倍），保存一个旧桶和新桶之间的连接，逐步将旧桶迁移到新桶中。实际上，在这次新的分配之后，每个参与过写操作的桶，如果操作还未完成，都将被迁移。被迁移的桶中的所有键值对都将被重新分配到新桶中，这意味着，先前同一个桶中存储在一起的键值对，现在可能被分配到不同的桶中。

Go 使用负载因子来判断何时开始分配新桶并迁移旧桶。

## 负载因子

Go 在 map 中使用 6.5 作为负载因子。你可以在代码中看到与负载因子相关的研究：

```go
// Picking loadFactor: too large and we have lots of overflow
// buckets, too small and we waste a lot of space. I wrote
// a simple program to check some stats for different loads:
// (64-bit, 8 byte keys and values)
//  loadFactor    %overflow  bytes/entry     hitprobe    missprobe
//        4.00         2.13        20.77         3.00         4.00
//        4.50         4.05        17.30         3.25         4.50
//        5.00         6.85        14.77         3.50         5.00
//        5.50        10.55        12.94         3.75         5.50
//        6.00        15.27        11.67         4.00         6.00
//        6.50        20.90        10.79         4.25         6.50
//        7.00        27.14        10.15         4.50         7.00
//        7.50        34.03         9.73         4.75         7.50
//        8.00        41.10         9.40         5.00         8.00
//
// %overflow   = percentage of buckets which have an overflow bucket
// bytes/entry = overhead bytes used per key/value pair
// hitprobe    = # of entries to check when looking up a present key
// missprobe   = # of entries to check when looking up an absent key
```

如果桶中键值对的平均容量超过 6.5，map 将会扩容。考虑到基于键的哈希值的分配并不均匀，正如我们在以上研究中看到的，使用 8 作为负载因子会导致大量的溢出桶。

这个系列的下一篇文章，["Go: Map Design by Code"](https://medium.com/@blanchon.vincent/go-map-design-by-code-part-ii-50d111557c08)，将会讲解 map 的内部实现。

---
via: https://medium.com/@blanchon.vincent/go-map-design-by-example-part-i-3f78a064a352

作者：[blanchon.vincent](https://medium.com/@blanchon.vincent)
译者：[DoubleLuck](https://github.com/DoubleLuck)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
