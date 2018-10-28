首发于：https://studygolang.com/articles/14132

# Go 中不定长度集合

如果你在使用像 C# 或 Java 这样的编程语言后来到 go，你发现的第一件事就是没有像 `List` 和 `Dictionary` 这样的传统集合类型。 这真让我困惑了好几个月。 我找到了一个名为 `container/list` 的软件包，并且几乎用它做所有的东西。

我脑后一直有一个声音在唠叨。语言设计者不应该不直接支持对未知长度的集合管理的功能。每个人都在讨论切片是如何在语言中被广泛使用，但我只是在有明确定义的容量或者它们通过函数返回时我才使用切片，这有点不对劲!!

因此，我在本月早些时候写了一篇文章，揭开了切片的盖子，希望能找到一些我不知道的魔法。我现在知道切片是如何工作的，但最终我仍然需要一个不断进行长度增长的数组。我在学校学过，使用链表更有效率，是存储大量数据更好的方法。特别是当你需要的集合长度未知时。这对我来说很有意义。

当我思考使用一个空切片时，我头脑中有一张非常**错误**的图片：

![slice-copy](https://raw.githubusercontent.com/studygolang/gctt-images/master/Collections-Of-Unknown-Length-in-Go/slice-copy.png)

我一直在想 go 是如何创建大量新的切片值和底层数组做大量内存分配，并且不断进行复制值。然后垃圾回收器会因为所有这些小变量被创建和销毁而过度工作。

我无法想象需要做数千次这种操作。其实有更好的方法或更效率的方式我没有意识到。

在研究并提出了很多问题之后，我得出的结论是，在大多数实际情况下，使用切片比使用链表更好。这就是为什么语言设计者花时间使切片尽可能高效工作，并且没有引入集合类型的原因。

我们可以连续几天讨论各种边界情况和性能问题，但 go 希望我们使用切片。因此切片应该是我们的首选，除非代码告诉我们存在问题。掌握切片就像学国际象棋游戏，易于学习但需要一辈子才能成为大师。因为底层数组可以共享，所以在使用中需要注意一些问题。

在继续阅读之前，你最好看一下我的另一篇文章 [Understanding Slices in Go Programming](http://www.goinggo.net/2013/08/understanding-slices-in-go-programming.html)。

本文的其余部分将解释如何使用切片处理未知容量的问题以及切片的运行机制。

以下是使用空切片来管理未知长度集合的示例：

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
)

type Record struct {
    ID int
    Name string
    Color string
}

func main() {
    // Let’s keep things unknown
    random := rand.New(rand.NewSource(time.Now().Unix()))

    // Create a large slice pretending we retrieved data
    // from a database
    data := make([]Record, 1000)

    // Create the data set
    for record := 0; record < 1000; record++ {
        pick := random.Intn(10)
        color := "Red"

        if pick == 2 {
            color = "Blue"
        }

        data[record] = Record{
            ID: record,
            Name: fmt.Sprintf("Rec: %d", record),
            Color: color,
        }
    }

    // Split the records by color
    var red []Record
    var blue []Record

    for _, record := range data {
        if record.Color == "Red" {
            red = append(red, record)
        } else {
            blue = append(blue, record)
        }
    }

    // Display the counts
    fmt.Printf("Red[%d] Blue[%d]\n", len(red), len(blue))
}

```
当我们运行这个程序时，由于随机数生成器，我们将得到不同长度的红色和蓝色切片。我们无法提前知道红色或蓝色切片的容量需要，这对我来说是一种典型的情况。

让我们分解出代码中更重要的部分：

这两行代码创建了空切片。

```go
var red []Record
var blue []Record
```

一个空切片长度和容量都是0，并且不存在底层数组。我们可以使用内置的 `append` 函数向切片中增加数据。

```go
red = append(red, record)
blue = append(blue, record)
```

`append` 函数功能非常酷，为我们做了很多东西。

Kevin Gillette 在我的小组讨论中进行了说明：
（https://groups.google.com/forum/#!topic/golang-nuts/nXYuMX55b6c）

在 go 语音规范中规定，前几千个元素在容量增长的时候每次都将容量翻倍，然后以~1.25的速率进行容量增长。

我不是学者，但我看到使用波浪号（~）相当多。有些人也许不知道这是什么意思，这里表示大约。因此，`append` 函数会增加底层数组的容量并为未来的增长预留空间。最终 `append` 函数将大约以1.25或25％的系数进行容量增长。

让我们证明 `append` 函数增长容量并高效运行：

```go
package main

import (
    "fmt"
    "reflect"
    "unsafe"
)

func main() {
    var data []string

    for record := 0; record < 1050; record++ {
        data = append(data, fmt.Sprintf("Rec: %d", record))

        if record < 10 || record == 256 || record == 512 || record == 1024 {
            sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&data)))

            fmt.Printf("Index[%d] Len[%d] Cap[%d]\n",
                record,
                sliceHeader.Len,
                sliceHeader.Cap)
        }
    }
}
```

输出结果：

```
Index[0] Len[1]  Cap[1]
Index[1] Len[2]  Cap[2]
Index[2] Len[3]  Cap[4]          - Ran Out Of Room, Double Capacity
Index[3] Len[4]  Cap[4]
Index[4] Len[5]  Cap[8]          - Ran Out Of Room, Double Capacity
Index[5] Len[6]  Cap[8]
Index[6] Len[7]  Cap[8]
Index[7] Len[8]  Cap[8]
Index[8] Len[9]  Cap[16]         - Ran Out Of Room, Double Capacity
Index[9] Len[10] Cap[16]
Index[256] Len[257] Cap[512]     - Ran Out Of Room, Double Capacity
Index[512] Len[513] Cap[1024]    - Ran Out Of Room, Double Capacity
Index[1024] Len[1025] Cap[1280]  - Ran Out Of Room, Grow by a factor of 1.25
```
如果我们观察容量值，我们可以看到 Kevin 是绝对正确的。容量正如他所说的那样在增长。在前1千的元素中，容量增加了一倍。然后容量以1.25或25％的系数增长。这意味着以这种方式使用切片将满足我们在大多数情况下所需的性能，并且内存不会成为问题。

最初我认为会为每次调用 `append` 时都会创建一个新的切片值，但事实并非如此。当我们调用 `append` 时，在栈中复制了 `red` 副本。然后当 `append` 返回时，会再进行一次复制操作，但使用的我们已有的内存。

```go
red = append(red, record)
```

在这种情况下，垃圾收集器没有工作，所以我们根本没有性能或内存问题。我的 C# 和引用类型的思想再次打击了我。

请坐好，因为下一个版本中的切片会有变化。

Dominik Honnef　创建了一个博客，用简明的英文（谢谢）解释了 Go tip 中正在编写的内容。这些是下一版本中的内容。这是他博客的和博客中关于切片部分的链接。这是一篇很棒博客，推荐阅读。

http://dominik.honnef.co/go-tip/

http://dominik.honnef.co/go-tip/2013-08-23/#slicing

你可以用切片做很多的事情，甚至可以写一整本关于这个主题的书。就像我之前说的那样，切片就像国际象棋一样，易于学习但需要一辈子才能成为大师。如果您来自其他语言，如 C# 和 Java，那么请拥抱切片并使用它。这正是 go 中正确的方式。

---

via: https://www.ardanlabs.com/blog/2013/08/collections-of-unknown-length-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[Alan](https://github.com/althen)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
