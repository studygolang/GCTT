# 如何以及为什么在 Go 中编写枚举
一个**枚举**（enum，**enumerator** 的缩写），是一组命名的常量值。枚举是一个强大 的工具，让开发者可以创建复杂的常量集，而这些常量集有着有用的名称和简单且唯一的取值。

*在我们走远之前，我想提一下[我最近启动了 Go Mastery，一门动手的 Golang 课程](https://qvault.io/go-mastery-course/)。如果想要了解更多关于 Go 的信息，请尝试下该课程，现在让我们回到枚举上面。*

## 语法示例
在一个常量声明中，[iota](https://golang.org/ref/spec#Iota) 关键字创建枚举作为连续的无类型整型常量。

```go
type BodyPart int

const (
    Head BodyPart = iota // Head = 0
    Shoulder             // Shoulder = 1
    Knee                 // Knee = 2
    Toe                  // Toe = 3
)
```

## 为什么应该使用枚举？
来看一些关于枚举你可能会有的几个疑问。首先枚举也许看起来没那么有用，但是我向你保证枚举是有用的。

**而且，如果想要一个整型常量，就不能用一个普通的 `const` 吗？比如，`const head = 0` ？**

可以，*可以*这么做，但是枚举的强大之处在于将常量*集*聚合在一起并且保证值*唯一*。通过使用枚举，编译器层面保证了你的常量（比如，`Head`，`Shoulder`，`Knee`，和 `Toe`）不会有相同的值。

**为什么不直接使用字符串作为唯一值？比如说，`const Head = "head"` 以及 `const Shoulder = "shoulder"` ？**

除了编译器无法保证唯一这个老生常谈的回答之外，一个字符串需要更多内存，并且在一些受限的情形下会导致性能问题。如果你有一组4个，10个，甚至100个唯一的值，你真的需要存储整个 `string` 吗？一个 `int` 型会占用更少的程序内存。

不仅仅与空间有关系，尤其是在现代硬件十分强大。假如你有类似下面这样的一些配置变量。

```go
const (
    statusSuccess = iota
    statusFailed
    statusPending
    statusComplete
)
```

假装你需要将 `statusFailed` 变更为 `statusCancelled`，以便和其他代码库保持一致。如果你先前没有使用枚举而是使用 `failed` 这个（字符串类型的）值，而现如今这个值散布在不同数据库中，那变更会变的*非常*困难。如果你使用的是`枚举`，你可以[修改名字](https://qvault.io/clean-code/naming-variables/)而不需要触及底层的值，你的代码还能保持干净。

## 从 1 开始枚举
有的时候，如果你是受虐狂，或者如果你是一个 Lua 开发者，希望你的枚举列表从 `1` 开始而不是从默认的 `0` 开始，在 Go 中你可以很轻易地实现。

```go
const (
    Head = iota + 1  // 1
    Shoulder         // 2
    Knee             // 3
    Toe              // 4
)
```

## 带有乘法的枚举
`iota` 关键字简单地代表一个自增的整型常量，即在同一个 `const` 块中每使用一次就会变大的一个数字。你可以对它使用任何你想要使用的数学运算。

```go
const (
    Head = iota + 1      // 0 + 1 = 1
    Shoulder = iota + 2  // 1 + 2 = 3
    Knee = iota * 10     // 2 * 10 = 20
    Toe = iota * 100     // 3 * 100 = 300
)
```

考虑到这一点，请记住你*可以*做不代表你*应该*这样做。

## 跳过的值的枚举
如果你想要跳过某个值，可以使用 _ 字符，就如同忽略（函数）返回的变量一样。

```go
const (
    Head = iota // Head = 0
    _
    Knee // Knee = 2
    Toe // Toe = 3
)
```

## 在 Go 中枚举的 String
Go 对于枚举并没有内置的 `string` 函数，但是可以很容易地通过实现 `String()` 。通过使用 `String()` 方法而非将常量设置为字符串类型，可以使枚举带有“可打印性”从而获得和使用字符串相同的好处。

```go
type BodyPart int

const (
    Head BodyPart = iota // Head = 0
    Shoulder // Shoulder = 1
    Knee // Knee = 2
    Toe // Toe = 3
)

func (bp BodyPart) String() string {
    return [...]string{"Head", "Shoulder", "Knee", "Toe"}
    // 译注：这里应该是 return [...]string{"Head", "Shoulder", "Knee", "Toe"}[bp]
}
```

但是这个方法存在一些“陷阱”，需要注意。如果 `const` 块中声明的数量与 `String()` 方法中所创建的“[常量切片](https://qvault.io/golang/golang-constant-maps-slices/)”中条目数对不上，编译器并不会警告可能存在的“越界”错误。同样的，如果你曾更新了常量中的某个名称，不要忘记更新列表中对应的字符串。

---
via: https://qvault.io/golang/golang-enum/

作者：[Lane Wagner](https://qvault.io/author/lane-c-wagner/)
译者：[dust347](https://github.com/dust347)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
