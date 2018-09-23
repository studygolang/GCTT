已发布：https://studygolang.com/articles/12930

# Go 实验报告：函数式编程之泛型

在 2017 年的年中，我在 GopherCon 上发表了《Go 的函数式编程》的演讲。我提出了一些函数式编程的概念，Gophers 使用它，可以提高编程效率，代码更加简洁。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-experience-report-generics-for-functional-patterns/functional-programming-in-go.jpeg)

> 函数式编程在 Go 是可以实现，只是不明显

演讲中一半是理论，另一半是可以让人使用的模式概念，其中大约四分之一是我认为是有实践价值的，其他的姑且值得一提。你需要"代码生成"（code generation）以实现它们。我在 [Github repo](https://github.com/go-functional/core) 分享了这些模式，欢迎 folk 到你们的仓库。

这篇文章是关于泛型如何使 Go 函数式编程模式更加强大，且不依赖于代码生成。

这里我没有遵守 [如何写实验报告](https://github.com/golang/go/wiki/ExperienceReports) 的指导原则，我想可以用稍微不同的方式来更好地表达想法。

## 一个例子

让我们从理想的 API 开始认识并探究，映射序列（Mapping over sequences）普遍存在绝大多数语言中，于是以此为例。

在 Go 中，我想这个理想的 API 是这样的：

```go
ints := []int{1, 2, 3, 4}
incremented := ints.Map(func(i int) int { return i + 1 })
// incremented = []int{2, 3, 4, 5}
```

如你所想，`Map` 方法取代了 `for` 循环，它帮助你写出简洁的代码（pure code），使用更加便利，更重要的是你无需自己写边界条件（range）。

上面的代码看上去很理想，但我们必须做出不少重大改变（例如，像自动类型转换）才能在 Go 语言中实现这个。

并且，尽管上述很理想，但事实上我们不得不添加另一个特例以使特定类型的 `slices` 可以执行新的 `Map` 方法，这与我的理想背道而驰。正如Russ Cox在 [GopherCon 2017主题演讲](https://www.youtube.com/watch?v=0Zbh_vmAKvk) 中所提到的，我宁愿将注意力放在与泛型有关的体验报告上。这就是这篇文章的内容。

## 一个现实点的例子

我认为需要对上面的 API 做一些调整。让它在原有语言中更加实用，并且更容易地集中注意力在泛型上（而不是更多的函数式编程模式，这可以在另一篇文章中）。

在这里我们将 `[]int` 封装了一个容器类型(container type)，然后给这个容器定义一个 `Map` 方法。

```go
incremented := WrapSlice(ints).Map(func(i int) int { return i + 1})
```

这个新的容器在Go中部分实现了类型语义。

在今天的Go中，除了 `[]int` 之外，没有办法让单个 Wrap 函数在任何其他类型上工作，因为我们没有泛型编程机制。

你可以得到最接近的方法是手写或者生成代码，使得 Wrap 可以适用于你想要的所有类型。

## 为什么这家伙反对代码生成

代码生成的想法、背后的动机与技术都很好。我在之前的实验报告中，已经写过关于它的想法，我并不想在这里或者其他文章中抨击它。

尽管听起来如此，但是因为我坚信代码生成是一个工具，需要为它选择正确的用武之地。然而，在这里，代码生成绝对用错地方了。

它有一个大问题，举个例子。假设您想要生成处理 `[] ints`，`[]string` 和自定义类型（MyCustomType）的代码。在今天的Go中，语言引擎将无法为您的所有类型提供相同的Wrap功能。与之相反，生成器可以生成这个API的代码：

```go
myStrings := []string{...}
WrapStringSlice(myStrings).Map(func(s string) string {
    return "hello " + s
})
myInts := []int{...}
WrapIntSlice(myInts).Map(func(i int) int {
    return i + 1
})
myCustomTypes := []MyCustomType{...}
WrapMyCustomTypeSlice(myCustomTypes).Map(func(m MyCustomType) MyCustomType {
    return m
})
```

所以我们每种类型都有一个 Wrap 函数。我们为所有类型提供兼容性，但是我们仍然没有API来针对我们所有类型编写泛型代码。

## 下一步

很明显，我正在写关于在 Go 中添加泛型的知识，但“泛型”可能意味着很多东西。我在这里举例说明了我想要一个泛型 API，它有点类似函数式编程模式。您可以将此 API 外推到其他函数式编程模式。

首先，我希望能够在 `[]T` 或者 `map[T]U` （T和U是任意类型）上调用 Map，并且能够将这些值转换为其他 slices 或 maps （[]A and map[B]C）。和我上一篇文章一样，我不打算在Go中发明泛型的语法，我只是想展示我想要的样子。我可以写一篇后续文章来提出一种语法。

## WrapSlice

我展示了我想要 WrapSlice 和 Map 看起来像上面那样，但这是一个简单的例子。Map 的强大功能是可以将切片从一种类型转换为另一种（即T1 => T2）。除了传递给Map的函数签名（注意参数和返回值是不同的类型）之外，该函数看起来与上例相同：

```go
ints := []int{1, 2, 3, 4}
strs := WrapSlice(ints).Map(func(i int) string {
    return strconv.Itoa(i*2)
})
// var strs []string = []string{"2", "4", "6", "8"}
bites := WrapSlice(strs).Map(func(s string) []byte {
    return []byte(s)
})
// var bites [][]byte = []byte{
//    []byte{50},
//    []byte{52},
//    []byte{54},
//    []byte{56},
// }
```

这里我们已经将 `[]int` 转换为 `[]string`，然后将 `[]string`转换为 `[]byte`

## WrapMap

WrapMap 在逻辑上与 WrapSlice 相似，只不过这次我们正在转换键值对。例如，这就是 map[string]int 转换到 map[int]string ：

```go
m := map[string]int{
    "1": 1,
    "2": 2,
    "3": 3,
}
converted := WrapMap(m).Map(func(k string, v int) (int, string) {
    newKey := strconv.Itoa(v)
    newVal, _ := strconv.Atoi(k)
    return newKey, newVal
})
// var converted map[int]string = { 1: "1", 2: "2", 3: "3" }
```

## 结论

从 `Wrap*` / `Map` 示例中，重要内容是它们适用于所有 `T` 和 `U` 类型。它们可以在语言代码之外编写，而可以进入标准库，第三方“FP”包，甚至是它们的生态系统。

最后，如果你看不到上述例子中泛型的用法，我会在这里解释。`WrapList` 通过列表元素的类型（即 []T 中的 T ）来参数化，并且 `WrapMap` 通过键的类型和值的类型（即， `map[T]U` ）被参数化。

然后，在通过映射表（即 `Map[U]` ）和 Map 由新的键值类型进行参数化的情况下，通过新的列表类型对 Map 进行参数化（在映射到 Map 的情况下）（即 `Map[X，Y]` ）。

就像我之前说过的，我没有在这里正式指定任何语法，但我确实使用了 bracket 语法（我更喜欢）来说明类型参数。没有任何承诺 —— 泛型语法规范是一个很大的话题！——— 我打算写一个更详细的语法提案。直到那时…

继续摇滚吧，Gophers！

---

via: https://medium.com/go-in-5-minutes/go-experience-report-generics-for-functional-patterns-eb6ce737bc1

作者：[Aaron Schlesinger](https://medium.com/@arschles)
译者：[lightfish-zhang](https://github.com/lightfish-zhang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
