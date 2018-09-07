首发于：https://studygolang.com/articles/14713

# Go 迭代切片 （Iterating Over Slices In Go）

切片在我的代码中随处可用。如果我正在使用 MongoDB 中的数据，它将存储在切片中。如果我需要在运行操作后跟踪一系列问题，它将存储在一个切片中。如果你还不了解切片是如何工作的，或者像我在开始时一样避免使用切片，请阅读下面这两篇文章以了解更多信息。

[理解 Go 中的 slice](http://www.goinggo.net/2013/08/understanding-slices-in-go-programming.html)

[Go 中不定长度集合](https://studygolang.com/articles/14132)

我在编码时经常问自己的一个问题是，“我想使用指向这个值的指针还是我想制作一个副本？” 虽然 Go 可以用作函数式编程语言，但它本质上却是一种命令式编程语言。这有什么不同？

函数式编程语言不允许你在创建且初始化变量或值之后改变他们。这意味着变量和值是不可变的，它们不能被更改。如果你需要更改变量的状态或值，则必须创造一个副本并初始化副本为更改后的变量和值。函数始终是传递副本，返回值也始终是副本。

在命令式编程语言中，我们可以创建可变或可更改的变量和值。我们可以将任何变量或值的指针传递给函数，而函数又可以根据需要更改状态。函数式编程语言希望您根据输入并产生结果的数学函数方式来进行思考。在命令式编程语言中，我们可以构建类似的函数，但是我们也可以构建对可以存在于内存中的任何位置的状态执行操作的函数。

能够使用指针有优势，但也可以让你陷入困境。使用指针可以减轻内存约束并尽可能的提高性能。但它会创造同步问题，例如对值和资源的共享访问。找到每个用例最适合的解决方案。对于你的 Go 程序，我建议在安全和实用的时候使用指针。Go 是一种命令式编程语言，所以利用好它的这些优势。

在 Go 中，一切都是按值传递的，记住这一点非常重要。我们可以通过值传递对象的地址，或者通过值传递对象的副本。当我们在 Go 中使用指针时，它有时会令人混淆，因为 Go  处理我们的所有引用。不要误会我的意思，Go做到这一点非常棒，但有时候你可以忘记变量的实际值。

在每个程序的某个时刻，我需要迭代一个切片来执行一些工作。在 Go 中，我们使用 for 循环结构来迭代切片。在开始时，我在迭代切片时犯了一些非常严重的错误，因为我误解了 range 关键字是如何工作的。我将向您展示一个令人讨厌的错误，我创建了一个让我困惑的迭代切片的功能。现在对我来说很明显为什么代码执行结果不对，但当时并不知道原因。

让我们创建一些简单的值并将它们放在切片中。然后我们将迭代切片，看看会发生什么。

```go
package main

import (
    "fmt"
)

type Dog struct {
    Name string
    Age int
}

func main() {
    jackie := Dog{
        Name: "Jackie",
        Age: 19,
    }

    fmt.Printf("Jackie Addr: %p\n", &jackie)

    sammy := Dog{
        Name: "Sammy",
        Age: 10,
    }

    fmt.Printf("Sammy Addr: %p\n", &sammy)

    dogs := []Dog{jackie, sammy}

    fmt.Println("")

    for _, dog := range dogs {
        fmt.Printf("Name: %s Age: %d\n", dog.Name, dog.Age)
        fmt.Printf("Addr: %p\n", &dog)

        fmt.Println("")
    }
}
```

该程序创建两只类型为狗的对象，并将它们放入狗类型的切片 dogs 中。我们显示每只狗的地址。然后我们迭代显示每只狗的名字，年龄和地址的切片。
这是该程序的输出：

```
Jackie Addr: 0x2101bc000
Sammy Addr: 0x2101bc040

Name: Jackie Age: 19
Addr: 0x2101bc060

Name: Sammy Age: 10
Addr: 0x2101bc060
```

那么为什么狗的值在循环内是不同的，为什么同一个地址出现两次呢？这一切都与 Go 的值传递的事实有关。在这个代码示例中，我们实际上在内存中创建了每个 Dog 的2个额外副本。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/iterating-over-slices-in-go/iterating-over-slices-in-go.png)

每个 Dog 的初始存在是使用复合字段创建的：

```go
jackie := Dog{
    Name: "Jackie",
    Age: 19,
}
```

将值放入切片时，将创建值的第一个副本：

```go
dogs := []Dog{jackie, sammy}
```

当我们遍历切片时，会创建值的第二个副本：

```go
dog := range dogs
```

现在我们可以看到为什么循环中变量狗的地址总是相同的。我们正在显示变量狗的地址，该变量恰好是 Dog 类型的局部变量，包含切片的每个索引的 Dog 的副本。对于切片的每次迭代，变量狗的地址是相同的。变量狗的值正在改变。

我之前谈到的那个令人讨厌的错误与我认为变量狗的地址可以用作指向切片内每个 Dog 值的指针。像这样的东西：

```go
allDogs := []*Dog{}

for _, dog := range dogs {
    allDogs = append(allDogs, &dog)
}

for _, dog := range allDogs {
    fmt.Printf("Name: %s Age: %d\n", dog.Name, dog.Age)
}
```

我创建了一个新的切片，用于保存指向 Dog 值的指针。然后我遍历 dogs ，存储每个 Dog 值的地址的放入新切片 allDogs 中。至少我认为我存储了每个 Dog 值的地址。

如果我将此代码添加到程序并运行它，这是输出：

```
Name: Sammy Age: 10
Name: Sammy Age: 10
```

我最终得到一个切片，其中每个元素具有相同的地址。该地址指向我们迭代的最后一个值的副本。哎呀！

如果制作所有这些副本不是您想要的，您可以使用指针。以下是使用指针的示例程序：

```go
package main

import (
    "fmt"
)

type Dog struct {
    Name string
    Age int
}

func main() {
    jackie := &Dog{
        Name: "Jackie",
        Age: 19,
    }

    fmt.Printf("Jackie Addr: %p\n", jackie)

    sammy := &Dog{
        Name: "Sammy",
        Age: 10,
    }

    fmt.Printf("Sammy Addr: %p\n\n", sammy)

    dogs := []*Dog{jackie, sammy}

    for _, dog := range dogs {
        fmt.Printf("Name: %s Age: %d\n", dog.Name, dog.Age)
        fmt.Printf("Addr: %p\n\n", dog)
    }
}
```

这是输出：

```
Jackie Addr: 0x2101bb000
Sammy Addr: 0x2101bb040

Name: Jackie Age: 19
Addr: 0x2101bb000

Name: Sammy Age: 10
Addr: 0x2101bb040
```

这次我们创建一个指向 Dog 值的指针。当我们遍历此切片时，变量狗的值是我们存储在切片中的每个 Dog 值的地址。我们使用与复合文字创建的相同的初始 Dog 值，而不是为每个 Dog 值创建两个额外的副本。

当切片是 Dog 值的集合或 Dog 值的指针集合时，范围循环是相同的。

```go
for _, dog := range dogs {
    fmt.Printf("Name: %s Age: %d\n", dog.Name, dog.Age)
}
```

无论我们是否使用指针， Go 都会处理对 Dog 值的访问。这很棒，但有时会导致一些混乱。至少这对我来说开始的时候是这样的。

我不能告诉你何时应该使用指针或何时应该使用副本。但请记住， Go 将按价值传递一切。这包括函数参数，返回值以及在切片、 map 或 channel上迭代时。

是的，你也可以遍历一个 channel 。看看我在 Ewen Cheslack-Postava 撰写的博客文章中改编的示例代码：

[http://ewencp.org/blog/golang-iterators/](http://ewencp.org/blog/golang-iterators/)

```go
package main

import (
    "fmt"
)

type Dog struct {
    Name string
    Age int
}

type DogCollection struct {
    Data []*Dog
}

func (this *DogCollection) Init() {
    cloey := &Dog{"Cloey", 1}
    ralph := &Dog{"Ralph", 5}
    jackie := &Dog{"Jackie", 10}
    bella := &Dog{"Bella", 2}
    jamie := &Dog{"Jamie", 6}

    this.Data = []*Dog{cloey, ralph, jackie, bella, jamie}
}

func (this *DogCollection) CollectionChannel() chan *Dog {
    dataChannel := make(chan *Dog, len(this.Data))

    for _, dog := range this.Data {
        dataChannel <- dog
    }

    close(dataChannel)

    return dataChannel
}

func main() {
    dc := DogCollection{}
    dc.Init()

    for dog := range dc.CollectionChannel() {
        fmt.Printf("Channel Name: %s\n", dog.Name)
    }
}
```

如果您运行该程序，您将获得以下输出：

```
Channel Name: Cloey
Channel Name: Ralph
Channel Name: Jackie
Channel Name: Bella
Channel Name: Jamie
```

我真的很喜欢这个示例代码，因为它展示了关闭的 channel 的美感。使该程序有效的关键是关闭的 channel 始终处于可发出信号的状态。这意味着通道上的任何读取都将立即返回。如果通道为空，则返回默认值。这使得 range 可以迭代传递到通道中的所有数据，并在通道为空时完成。一旦通道为空，通道上的下一次读取将返回 nil 。这会导致循环终止。

切片非常棒，重量轻，功能强大。您应该使用它们并获得它们提供的好处。请记住，当您在切片上进行迭代时，您将获得切片的每个元素的副本。如果这恰好是一个对象，那么您将获得该对象的副本。不要在循环中使用局部变量的地址。这是一个局部变量，它包含切片元素的副本，并且只有内容。不要和我一样犯同样的错误。

---

via: https://www.ardanlabs.com/blog/2013/09/iterating-over-slices-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[LFasMike](https://github.com/LFasMike)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
