首发于：https://studygolang.com/articles/17362

# 减少类型层次

## 介绍

我发现许多面向对象的编程语言（如 C ＃和 Java）的开发人员转向 Go 语言。 由于这些开发人员已接受过使用类型层次结构的培训，因此他们在 Go 中使用相同的模式是有道理的。 但是 Go 语言的某些方面，不允许类型层次结构提供与其他面向对象编程语言相同的功能级别。 具体来说，Go 中不存在基类和子类的概念，因此类重用需要不同的思维方式。

在这篇文章中，我将展示为什么类型层次结构在 Go 语言中使用，并不总是最佳模式。 我将解释为什么将具体类型组合在一起最好的办法，是通过共同的行为，而不是通过共同的状态。 我将展示如何利用接口来分组和解耦具体类型，最后，我将提供一些关于声明类型的指南。

## 第一部分

让我们从一个程序开始，正如我经常看到那些试图学习 Go 的人也是这样。 该程序使用传统的类型层次结构模式，这种模式在面向对象的程序中很常见。

https://play.golang.org/p/ZNWmyoj55W

### 清单 1

```go
01 package main
02
03 import "fmt"
04
05 //Animal 包含动物的所有基本领域。
06 type Animal struct {
07      Name     string
08      IsMammal bool
09 }
10
11 // Speak 为所有动物提供如何说话的通用行为。
13 func (a Animal) Speak() {
14      fmt.Println("UGH!",
15          "My name is", a.Name,
16          ", it is", a.IsMammal,
17          "I am a mammal")
18 }
```

在清单 1 中，我们看到了传统面向对象程序的开始。 在第 06 行，我们有具体类型 Animal 的声明，它有两个字段，Name 和 IsMammal。 然后在第 13 行，我们有一个名为 Speak 的方法，允许动物说话。 由于 Animal 是所有动物的基础类型，因此 Speak 方法的实现是通用的，任何超出此基础状态的指定动物就不好使用了。

### 清单 2

```go
20 // Dog 包含 Animal 的所有，但只包含 Dog 的特定属性。
22 type Dog struct {
23      Animal
24      PackFactor int
25 }
26
27 // Speak 知道如何像狗一样叫
28 func (d Dog) Speak() {
29      fmt.Println("Woof!",
30          "My name is", d.Name,
31          ", it is", d.IsMammal,
32          "I am a mammal with a pack factor of", d.PackFactor)
33 }
```

清单 2 声明了一个名为 Dog 的新具体类型，它嵌入了一个类型为 Animal 的值，并且具有一个名为 PackFactor 的唯一字段。 我们看到使用组合来重用 Animal 类型的字段和方法。 在这种情况下，组合在类型重用方面提供了一些与继承相同的好处。 Dog 类型还实现了自己的 Speak 方法版本，允许 Dog 像狗一样吠叫。 此方法重写了实现 Animal 类型的 Speak 方法。

### 清单 3

```go
35 // Cat 有 Animal 的所有方法，但特殊的方法只有 Cat 有
37 type Cat struct {
38      Animal
39      ClimbFactor int
40 }
41
42 // Speak 知道如何像猫一样叫
43 func (c Cat) Speak() {
44      fmt.Println("Meow!",
45      "My name is", c.Name,
46      ", it is", c.IsMammal,
47      "I am a mammal with a climb factor of", c.ClimbFactor)
```

接下来，我们在清单 3 中有第三个具体类型，叫做 Cat。它还嵌入了一个类型为 Animal 的值，并且有一个名为 ClimbFactor 的字段。 出于同样的原因，我们再次看到组合物的使用，并且 Cat 有一种名为 Speak 的方法，允许 Cat 像猫一样喵喵叫。 同样，此方法重写了实现 Animal 类型的 Speak 方法。

### 清单 4

```go
50 func main() {
51
52      // 通过初始化 Animal 的部分 , 然后初始化其特定 Dog 属性来创建 Dog
54      d := Dog{
55      Animal: Animal{
56          Name:     "Fido",
57          IsMammal: true,
58      },
59      PackFactor: 5,
60  }
61
62      // 通过初始化 Animal 的部分 , 然后初始化其特定 Cat 属性来创建 Cat
64      c := Cat{
65      Animal: Animal{
66      Name:     "Milo",
67          IsMammal: true,
68      },
69      ClimbFactor: 4,
70  }
71
72  // Have the Dog and Cat speak.
73  d.Speak()
74  c.Speak()
75 }
```

在清单 4 中，我们将主要功能放在一起。 在第 54 行，我们使用字面结构体，创建 Dog 类型的值，并初始化嵌入的 Animal 值和 PackFactor 字段。 在第 64 行，我们使用字面结构体创建 Cat 类型的值，并初始化嵌入的 Animal 值和 ClimbFactor 字段。 然后，最后，我们在第 73 行和第 74 行给 Dog 和 Cat 值调用 Speak 方法。

这适用于 Go 语言，您可以看到嵌入类型的使用如何提供熟悉的类型层次结构功能。 然而，在 Go 语言中执行此操作存在一些缺陷，原因之一是 Go 不支持子类型的概念。 这意味着您不能像在其他面向对象的语言中那样使用 Animal 类型作为基类型。

重要的是要理解，在 Go 语言中，Dog 和 Cat 类型不能用作 Animal 类型的值。 我们所拥有的是为 Dog 和 Cat 类型嵌入了 Animal 类型的值。 您不能将 Dog 或 Cat 传递给任何接受 Animal 类型值的函数。 这也意味着无法通过 Animal 类型在同一列表中将一组 Cat 和 Dog 值组合在一起。

### 清单 5

```go
// 尝试使用 Animal 作为基类
animals := []Animal{
    Dog{}
    Cat{},
}

: cannot use Dog literal (type Dog) as type Animal in array or slice literal
：在数组或列表中不能使用 Dog 类作为 Animal 类
: cannot use Cat literal (type Cat) as type Animal in array or slice literal
：在数组或列表中不能使用 Cat 类作为 Animal 类
```

清单 5 显示了当您尝试将 Animal 类型用作传统面向对象方式的基类时，Go 语言中发生的情况。 编译器非常清楚 Dog 和 Cat 类不能用作 Animal 类。

在这种情况下，Animal 类型和类型层次结构的使用并没有为我们提供任何实际价值。 我认为它正在引领我们走上一条不可读，不简单，适应性不强的代码之路。

## 第二部分

在 Go 语言中编码时，尽量避免使用这些类型层次结构来促进共同状态的思想，并考虑共同行为。 如果我们考虑他们展示的共同行为，我们可以将一组狗和猫的值分组。 在这种情况下，Speak 有一个共同的行为。

让我们看看这段代码的另一个实现，它关注行为。

[https://play.golang.org/p/6aLyTOTIj_](https://play.golang.org/p/6aLyTOTIj_)

### 清单 6：

```go
01 package main
02
03 import "fmt"
04
05 // 如果他们想成为这个群体的一部分，
06 // 扬声器为所有具体类型提供了一个共同的行为，
07 // 这是这些具体类型的约定。
08 type Speaker interface {
09  Speak()
10 }
```

新程序从清单 6 开始，我们在第 08 行添加了一个名为 Speaker 的新类型。这不是我们之前声明的结构体类型的具体类型。 这是一种接口类型，它声明了一个行为约定，它允许我们对一组实现 Speak 方法的不同具体类型进行分组和处理。

### 清单 7：

```go
12 // Dog 结构体包含 Dog 需要的一切
13 type Dog struct {
14  Name       string
15  IsMammal   bool
16  PackFactor int
17 }
18
19 // Speak 知道如何像狗一样叫
20 // 这使得 Dog 现在成为了解如何说话的一组具体类型的一部分。
21
22 func (d Dog) Speak() {
23  fmt.Println("Woof!",
24      "My name is", d.Name,
25      ", it is", d.IsMammal,
26      "I am a mammal with a pack factor of", d.PackFactor)
27 }
28
29 // Cat 结构体包含 Cat 需要的一切
30 type Cat struct {
31  Name        string
32  IsMammal    bool
33  ClimbFactor int
34 }
35
36 // Speak 知道怎样像猫一样叫
37 // 这使得 Cat 现在成为了解如何说话的一组具体类型的一部分。
38
39 func (c Cat) Speak() {
40  fmt.Println("Meow!",
41      "My name is", c.Name,
42      ", it is", c.IsMammal,
43      "I am a mammal with a climb factor of", c.ClimbFactor)
44 }
```

在清单 7 中，我们再次声明了具体类型的狗和猫。 此代码删除 Animal 类型并将这些公共字段直接复制到 Dog 和 Cat 中。

我们为什么那么做？

+ Animal 类型提供了可重用状态的抽象层。
+ 该程序从不需要创建或仅使用 Animal 类型的值。
+ Animal 类型的 Speak 方法的实现是一种概括。

以下是有关声明类型的一些指导原则

+ 声明类型表示新的或独特的事物。
+ 验证是否自己创建或使用任何类型的值。
+ 嵌入类型以重用您需要满足的现有行为。
+ 作为现有类型的别名或抽象的问题类型。
+ 其唯一目的是分享共同状态的问题类型。

让我们现在来看一下主函数 main

### 清单 8

```go
46 func main() {
47
48  // 创建一个知道如何说话的动物列表
49  speakers := []Speaker{
50
51  // 通过初始化其 Animal 部分，然后初始化其特定的 Dog 属性来创建一个
52  // Dog。
53      Dog{
54          Name:       "Fido",
55          IsMammal:   true,
56          PackFactor: 5,
57      },
58
59      // 通过初始化其 Animal 部分，然后初始化其特定的 Dog 属性来创建
60      // 一个 Cat
61      Cat{
62      Name:        "Milo",
63      IsMammal:    true,
64      ClimbFactor: 4,
65      },
66  }
67
68  // 让 Animal 叫
69  for _, spkr := range speakers {
70      spkr.Speak()
71  }
72 }
```

在清单 8 的第 49 行，我们创建了一个 Speaker 接口值列表，以便在其常见行为下将 Dog 和 Cat 值组合在一起。 我们在第 53 行创建了 Dog 类型的值，在第 61 行创建了 Cat 类型的值。最后在第 69-71 行，我们迭代了 Speaker 接口值的列表并让 Dog 和 Cat 说话。

关于我们所做的改变的一些最终想法：

+ 我们不需要基类或类型层次结构来将具体类型值组合在一起。
+ 接口允许我们创建一个不同具体类型值的切片，并通过它们的共同行为来处理这些值。
+ 我们通过不声明一个从未被程序单独使用过的类型，来消除类型污染。

## 总结

Go 语言中的组合还有很多，但这是对使用类型层次结构的问题的初步理解。 每条规则总是有例外情况，但请务必遵循这些指导原则，直到您足够了解做出异常的权衡。

要了解有关此文章所涉及的构图和其他主题的更多信息，请查看以下其他博文：

[Exported/Unexported Identifiers In Go](https://www.ardanlabs.com/blog/2014/03/exportedunexported-identifiers-in-go.html)

[Methods Interfaces And Embedded Types](https://www.ardanlabs.com/blog/2014/05/methods-interfaces-and-embedded-types.html)

[Object Oriented Mechanics In Go](https://www.ardanlabs.com/blog/2015/03/object-oriented-programming-mechanics.html)

[Composition With Go](https://www.ardanlabs.com/blog/2015/09/composition-with-go.html)

## 鸣谢

以下是社区的一些朋友，我想感谢他们花时间审阅帖子并提供反馈。

Daniel Vaughan, Ted Young, Antonio Troina, Adam Straughan, Kaveh Shahbazian, Daniel Whitenack, Todd Rafferty

---

via: https://www.ardanlabs.com/blog/2016/10/reducing-type-hierarchies.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[Jasonjiang27](https://github.com/Jasonjiang27)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
