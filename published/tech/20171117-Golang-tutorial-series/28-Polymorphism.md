已发布：https://studygolang.com/articles/12681

# 第 28 篇：多态

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 28 篇。

Go 通过[接口](https://studygolang.com/articles/12266)来实现多态。我们已经讨论过，在 Go 语言中，我们是隐式地实现接口。一个类型如果定义了接口所声明的全部[方法](https://studygolang.com/articles/12264)，那它就实现了该接口。现在我们来看看，利用接口，Go 是如何实现多态的。

## 使用接口实现多态

一个类型如果定义了接口的所有方法，那它就隐式地实现了该接口。

**所有实现了接口的类型，都可以把它的值保存在一个接口类型的变量中。在 Go 中，我们使用接口的这种特性来实现多态**。

通过一个程序我们来理解 Go 语言的多态，它会计算一个组织机构的净收益。为了简单起见，我们假设这个虚构的组织所获得的收入来源于两个项目：`fixed billing` 和 `time and material`。该组织的净收益等于这两个项目的收入总和。同样为了简单起见，我们假设货币单位是美元，而无需处理美分。因此货币只需简单地用 `int` 来表示。（我建议阅读 https://forum.golangbridge.org/t/what-is-the-proper-golang-equivalent-to-decimal-when-dealing-with-money/413 上的文章，学习如何表示美分。感谢 Andreas Matuschek 在评论区指出这一点。）

我们首先定义一个接口 `Income`。

```go
type Income interface {
    calculate() int
    source() string
}
```

上面定义了接口 `Interface`，它包含了两个方法：`calculate()` 计算并返回项目的收入，而 `source()` 返回项目名称。

下面我们定义一个表示 `FixedBilling` 项目的结构体类型。

```go
type FixedBilling struct {
    projectName string
    biddedAmount int
}
```

项目 `FixedBillin` 有两个字段：`projectName` 表示项目名称，而 `biddedAmount` 表示组织向该项目投标的金额。

`TimeAndMaterial` 结构体用于表示项目 Time and Material。

```go
type TimeAndMaterial struct {
    projectName string
    noOfHours  int
    hourlyRate int
}
```

结构体 `TimeAndMaterial` 拥有三个字段名：`projectName`、`noOfHours` 和 `hourlyRate`。

下一步我们给这些结构体类型定义方法，计算并返回实际收入和项目名称。

```go
func (fb FixedBilling) calculate() int {
    return fb.biddedAmount
}

func (fb FixedBilling) source() string {
    return fb.projectName
}

func (tm TimeAndMaterial) calculate() int {
    return tm.noOfHours * tm.hourlyRate
}

func (tm TimeAndMaterial) source() string {
    return tm.projectName
}
```

在项目 `FixedBilling` 里面，收入就是项目的投标金额。因此我们返回 `FixedBilling` 类型的 `calculate()` 方法。

而在项目 `TimeAndMaterial` 里面，收入等于 `noOfHours` 和 `hourlyRate` 的乘积，作为 `TimeAndMaterial` 类型的 `calculate()` 方法的返回值。

我们还通过 `source()` 方法返回了表示收入来源的项目名称。

由于 `FixedBilling` 和 `TimeAndMaterial` 两个结构体都定义了 `Income` 接口的两个方法：`calculate()` 和 `source()`，因此这两个结构体都实现了 `Income` 接口。

我们来声明一个 `calculateNetIncome` 函数，用来计算并打印总收入。

```go
func calculateNetIncome(ic []Income) {
    var netincome int = 0
    for _, income := range ic {
        fmt.Printf("Income From %s = $%d\n", income.source(), income.calculate())
        netincome += income.calculate()
    }
    fmt.Printf("Net income of organisation = $%d", netincome)
}
```

上面的[函数](https://studygolang.com/articles/11892)接收一个 `Income` 接口类型的[切片](https://studygolang.com/articles/12121)作为参数。该函数会遍历这个接口切片，并依个调用 `calculate()` 方法，计算出总收入。该函数同样也会通过调用 `source()` 显示收入来源。根据 `Income` 接口的具体类型，程序会调用不同的 `calculate()` 和 `source()` 方法。于是，我们在 `calculateNetIncome` 函数中就实现了多态。

如果在该组织以后增加了新的收入来源，`calculateNetIncome` 无需修改一行代码，就可以正确地计算总收入了。:)

最后就剩下这个程序的 `main` 函数了。

```go
func main() {
    project1 := FixedBilling{projectName: "Project 1", biddedAmount: 5000}
    project2 := FixedBilling{projectName: "Project 2", biddedAmount: 10000}
    project3 := TimeAndMaterial{projectName: "Project 3", noOfHours: 160, hourlyRate: 25}
    incomeStreams := []Income{project1, project2, project3}
    calculateNetIncome(incomeStreams)
}
```

在上面的 `main` 函数中，我们创建了三个项目，有两个是 `FixedBilling` 类型，一个是 `TimeAndMaterial` 类型。接着我们创建了一个 `Income` 类型的切片，存放了这三个项目。由于这三个项目都实现了 `Interface` 接口，因此可以把这三个项目放入 `Income` 切片。最后我们将该切片作为参数，调用了 `calculateNetIncome` 函数，显示了项目不同的收益和收入来源。

以下完整的代码供你参考。

```go
package main

import (
    "fmt"
)

type Income interface {
    calculate() int
    source() string
}

type FixedBilling struct {
    projectName string
    biddedAmount int
}

type TimeAndMaterial struct {
    projectName string
    noOfHours  int
    hourlyRate int
}

func (fb FixedBilling) calculate() int {
    return fb.biddedAmount
}

func (fb FixedBilling) source() string {
    return fb.projectName
}

func (tm TimeAndMaterial) calculate() int {
    return tm.noOfHours * tm.hourlyRate
}

func (tm TimeAndMaterial) source() string {
    return tm.projectName
}

func calculateNetIncome(ic []Income) {
    var netincome int = 0
    for _, income := range ic {
        fmt.Printf("Income From %s = $%d\n", income.source(), income.calculate())
        netincome += income.calculate()
    }
    fmt.Printf("Net income of organisation = $%d", netincome)
}

func main() {
    project1 := FixedBilling{projectName: "Project 1", biddedAmount: 5000}
    project2 := FixedBilling{projectName: "Project 2", biddedAmount: 10000}
    project3 := TimeAndMaterial{projectName: "Project 3", noOfHours: 160, hourlyRate: 25}
    incomeStreams := []Income{project1, project2, project3}
    calculateNetIncome(incomeStreams)
}
```

[在 playground 上运行](https://play.golang.org/p/UClAagvLFT)

该程序会输出：

```
Income From Project 1 = $5000
Income From Project 2 = $10000
Income From Project 3 = $4000
Net income of organisation = $19000
```

## 新增收益流

假设前面的组织通过广告业务，建立了一个新的收益流（Income Stream）。我们可以看到添加它非常简单，并且计算总收益也很容易，我们无需对 `calculateNetIncome` 函数进行任何修改。这就是多态的好处。

我们首先定义 `Advertisement` 类型，并在 `Advertisement` 类型中定义 `calculate()` 和 `source()` 方法。

```go
type Advertisement struct {
    adName     string
    CPC        int
    noOfClicks int
}

func (a Advertisement) calculate() int {
    return a.CPC * a.noOfClicks
}

func (a Advertisement) source() string {
    return a.adName
}
```

`Advertisement` 类型有三个字段，分别是 `adName`、`CPC`（每次点击成本）和 `noOfClicks`（点击次数）。广告的总收益等于 `CPC` 和 `noOfClicks` 的乘积。

现在我们稍微修改一下 `main` 函数，把新的收益流添加进来。

```go
func main() {
    project1 := FixedBilling{projectName: "Project 1", biddedAmount: 5000}
    project2 := FixedBilling{projectName: "Project 2", biddedAmount: 10000}
    project3 := TimeAndMaterial{projectName: "Project 3", noOfHours: 160, hourlyRate: 25}
    bannerAd := Advertisement{adName: "Banner Ad", CPC: 2, noOfClicks: 500}
    popupAd := Advertisement{adName: "Popup Ad", CPC: 5, noOfClicks: 750}
    incomeStreams := []Income{project1, project2, project3, bannerAd, popupAd}
    calculateNetIncome(incomeStreams)
}
```

我们创建了两个广告项目，即 `bannerAd` 和 `popupAd`。`incomeStream` 切片包含了这两个创建的广告项目。

```go
package main

import (
    "fmt"
)

type Income interface {
    calculate() int
    source() string
}

type FixedBilling struct {
    projectName  string
    biddedAmount int
}

type TimeAndMaterial struct {
    projectName string
    noOfHours   int
    hourlyRate  int
}

type Advertisement struct {
    adName     string
    CPC        int
    noOfClicks int
}

func (fb FixedBilling) calculate() int {
    return fb.biddedAmount
}

func (fb FixedBilling) source() string {
    return fb.projectName
}

func (tm TimeAndMaterial) calculate() int {
    return tm.noOfHours * tm.hourlyRate
}

func (tm TimeAndMaterial) source() string {
    return tm.projectName
}

func (a Advertisement) calculate() int {
    return a.CPC * a.noOfClicks
}

func (a Advertisement) source() string {
    return a.adName
}
func calculateNetIncome(ic []Income) {
    var netincome int = 0
    for _, income := range ic {
        fmt.Printf("Income From %s = $%d\n", income.source(), income.calculate())
        netincome += income.calculate()
    }
    fmt.Printf("Net income of organisation = $%d", netincome)
}

func main() {
    project1 := FixedBilling{projectName: "Project 1", biddedAmount: 5000}
    project2 := FixedBilling{projectName: "Project 2", biddedAmount: 10000}
    project3 := TimeAndMaterial{projectName: "Project 3", noOfHours: 160, hourlyRate: 25}
    bannerAd := Advertisement{adName: "Banner Ad", CPC: 2, noOfClicks: 500}
    popupAd := Advertisement{adName: "Popup Ad", CPC: 5, noOfClicks: 750}
    incomeStreams := []Income{project1, project2, project3, bannerAd, popupAd}
    calculateNetIncome(incomeStreams)
}
```

[在 playground 中运行](https://play.golang.org/p/BYRYGjSxFN)

上面程序会输出：

```
Income From Project 1 = $5000
Income From Project 2 = $10000
Income From Project 3 = $4000
Income From Banner Ad = $1000
Income From Popup Ad = $3750
Net income of organisation = $23750
```

你会发现，尽管我们新增了收益流，但却完全没有修改 `calculateNetIncome` 函数。这就是多态带来的好处。由于新的 `Advertisement` 同样实现了 `Income` 接口，所以我们能够向 `incomeStreams` 切片添加 `Advertisement`。`calculateNetIncome` 无需修改，因为它能够调用 `Advertisement` 类型的 `calculate()` 和 `source()` 方法。

本教程到此结束。祝你愉快。

**上一教程 - [组合取代继承](https://studygolang.com/articles/12680)**

**下一教程 - [Defer](https://studygolang.com/articles/12719)**

---

via: https://golangbot.com/polymorphism/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
