首发于：https://studygolang.com/articles/25120

# 语法糖的代价

在 Go 语言中，你可以用少量的代码表达很多东西。您通常可以查看一小段代码并清楚地了解此程序的功能。这在 Go 社区中被称为地道的 Go 代码。保持跨项目代码的一致性需要持续不断地努力。

当我遇到 Go 的部分看起来不像地道 Go 代码时，这通常是有原因的。最近，我注意到 Go 中的接口切片(或抽象数组)工作方式有点怪异。这种怪异有助于理解在 Go 中使用复杂类型会带来一些成本，而且这些[语法糖](https://en.wikipedia.org/wiki/Syntactic_sugar)并不总是没有代价的。为深入了解问题出现的原因，我对遇到的行为进行拆分，这助于阐明 Go 的一些设计原则

## 举例说明

我们将编写一个小型程序，它定义一个动物列表(例如，dogs)，并调用一个函数，将每个动物的噪声输出到控制台

```go
animals := []Animal{Dog{}}
PrintNoises(animals)
```

程序成功通过编译，并在控制台打印出了“ Woof！”。下面就是这个程序的类似的版本:

```go
dogs := []Dog{Dog{}}
PrintNoises(dogs)
```

程序无法编译，并将以下错误打印到控制台，而不是输出 "Woof!"

```go
cannot use dogs (type []Dog) as type []Animal in argument to PrintNoises
```

如果你熟悉 Go，你可能会认为应该检查一下 `Dog` 实现了 `Animal`，对吧? 如果是实现错误，它的输出应该类似于

```go
cannot use dogs (type []Dog) as type []Animal in argument to PrintNoises: []Dog does not implement []Animal (missing Noise method)
```

为什么第一个程序可以用 `Dog` 作为 `Animal` 来编译和运行，而第二个程序却不能，即使它们看起来都是地道的和正确的

下面是本例中用作参考的其余代码。你可以通过编译它，来了解上述用法的内部原理

```go
type Animal interface {
  Noise() string
}

type Dog struct{}

func (Dog) Noise() string {
  return "Woof!"
}

func PrintNoises(as []Animal) {
  for _, a := range as {
    fmt.Println(a.Noise())
  }
}
```

## 进一步简化问题

让我们试着用一种更简单的方法来复现这个问题，以便更好地理解它。静态类型检查是一种有用的 Go pattern，用于断言类型是否实现了接口。让我们先检查一下 `Dog` 是否实现了 `Animal`

```go
var _ Animal = Dog{}
```

上面代码编译成功。那我们接下来就检查程序中用到的 `slices`

```go
var _ []Animal = []Dog{}
```

上面代码没有编译成功，编译器报错:

```go
cannot use []Dog literal (type []Dog) as type []Animal in assignment
```

现在，我们已经复现了一个与例子类似(但不是完全相同)的错误。利用这些不同的线索，我做了一些研究来找出如何解决这个问题，以及为什么会发生这样的事情

## 寻找解决方案

在做了一些研究之后，我发现了两件事:一个是解决方案，另一个是背后的原理。我们从修正程序开始，因为它有助于说明基本原理

下面是最初未能编译的代码的一个修复:

```go
dogs := []Dog{Dog{}}
// 新逻辑: 把 dogs 的切片转换成 animals 的切片
animals := []Animal{}
for _, d := range dogs {
  animals = append(animals, Animal(d))
}
PrintNoises(animals)
```

通过将 `Dog` 的切片转换为 `Animal` 的切片，现在可以将其传入 `Printnoise` 函数并成功运行。当然，这看起来有点傻，因为它基本上是已经运行的第一个程序的冗长版本。然而，在一个更大的项目中，这一点可能并那么明显。修复的代价是多了四行代码。这四行代码似乎是多余，直到你开始考虑作为开发人员必须修复它的原因

## 寻找原理

现在你知道如何修复它，我们来探究它背后的原理。我找到了不错的解析：[go 不支持切片中协变](https://www.reddit.com/r/golang/comments/3gtg3i/passing_slice_of_values_as_slice_of_interfaces/)
（译者注: [协变](https://zh.wikipedia.org/wiki/%E5%8D%8F%E5%8F%98%E4%B8%8E%E9%80%86%E5%8F%98): 原文单词为 covariance, 是指在计算机科学中，描述具有父/子型别关系的多个型别通过型别构造器、构造出的多个复杂型别之间是否有父/子型别关系的用语)

换句话说，Go 不会执行导致 O(N) 线性操作的类型转换(如切片的情况)，而是将责任委托给开发人员。也就是说，执行这种类型的转换是有成本的。不过，Go 并不是每次都这样做。例如，当将字符串转换为 []byte 节时，Go 将免费为您执行这种线性转换，这可能是因为这种转换通常很方便。这只是语言中语法糖的众多例子之一。对于切片(和其他非基本类型)，Go 选择不为您承担执行此操作的额外成本

这是有道理的——在我使用 Go 的 3 年里，这是我第一次遇到这种情况。这可能是因为 Go 在语法中灌输了“simpler is better”的思想

## 结论

一门语言的作者通常会在语法糖方面做出权衡，有时他们会添加功能，即便这会使语言变得更加臃肿，有时他们会将成本转嫁给开发人员。我认为，不隐式地执行高开销的操作的决定在保持 Go 语言地道、整洁、可控上有积极的影响。

上面的例子只是这个道理的一个应用。这个例子表明，熟悉一种语言的习惯用法有副作用。对设计决策保持深思熟虑总是一个好主意，而不是期望语言或编译器能帮到你。

我鼓励您在 Go 中寻找更多存在权衡语法的地方。它能帮助你更好地理解这门语言。我亦如此。

## 引用

以下是本文的引用:

* [GitHub Gist of the above example](https://gist.github.com/asilvr/4d4da3cdc8180c5a9740d2890d833923)
* [Go 语言官网](https://golang.org)
* [Thread on covariance in Go](https://www.reddit.com/r/golang/comments/3gtg3i/passing_slice_of_values_as_slice_of_interfaces/)
* [Big-O notation](https://en.wikipedia.org/wiki/Big_O_notation)
* [Syntactic sugar](https://en.wikipedia.org/wiki/Syntactic_sugar)

---

via: https://medium.com/@asilvr/the-cost-of-syntactic-sugar-in-go-5aa9dc307fe0

作者：[Katy Slemon](https://medium.com/@katyslemon)
译者：[Alex1996a](https://github.com/Alex1996a)
校对：[lxbwolf](https://github.com/lxbwolf)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
