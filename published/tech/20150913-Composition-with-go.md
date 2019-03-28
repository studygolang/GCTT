首发于：https://studygolang.com/articles/19336

# Go 语言中的组合

组合超越了[嵌入式](https://www.ardanlabs.com/blog/2014/05/methods-interfaces-and-embedded-types.html) 结构。这是我们可以用来设计更好的 APIs 并通过较小的模块构建更大的程序的范式。这一切都是从单一目类型的声明和实现开始。程序在架构时考虑到组合能更好的扩展和适应不断变化的需求。它们能更容易阅读和推理。

为了证明这个观点，我们来评审下面的程序：

## [示例代码](https://github.com/ardanlabs/gotraining/blob/c081f15e59fbe895c50b25a8a2d2eaf7a5772cbc/topics/composition/example4/example4.go)

这个代码示例探讨嵌入式结构，并让我们得以研究怎样使用组合能设计灵活而且易读的代码。一个程序包输出的每个标识符组成程序包的 API。这包括所有的常量、变量、类型结构、方法和函数等输出。注释是每个程序包的 API 中经常被忽视的一方面，要非常清楚和简洁以便与程序包的使用者进行信息交流。

这个例子太长所以我们把它分解成碎片然后用我们自己的办法了解它。

这个程序的概念是我们雇佣一个承包商修复我们的房子。特别是在房子里有些木板已经腐烂需要被扔掉，新的板子需要被固定。那么承包商会提供钉子、木板和工具来完成这个工作。

## 清单 1

```go
13 // Board represents a surface we can work on.
14 type Board struct {
15     NailsNeeded int
16     NailsDriven int
17 }
```

在清单 1 中，我们声明了 Board 类型，一个 Board 有 2 个字段，木板需要的钉子数量和当前钉入木板的钉子数量。现在，让我们看看接口声明：

## 清单 2

```go
21 // NailDriver represents behavior to drive nails into a board.
22 type NailDriver interface {
23     DriveNail(nailSupply *int, b *Board)
24 }
25
26 // NailPuller represents behavior to remove nails into a board.
27 type NailPuller interface {
28     PullNail(nailSupply *int, b *Board)
29 }
30
31 // NailDrivePuller represents behavior to drive/remove nails into a board.
32 type NailDrivePuller interface {
33     NailDriver
34     NailPuller
35 }
```
清单 2 展示的接口通过承包商需要的工具声明我们需要的行为。第 22 行的 NailDriver 接口声明钉一个钉子到木板的行为。该方法提供了钉子的供给和把钉子钉入木板的行为。第 27 行的 NailPuller 接口声明了相反的行为。这个方法提供钉子和木板的供给，但它会把钉子从木板上拔出来并把钉子放回供给。

这 2 个接口，NailDriver 和 NailPuller，都实现了一个单一的定义好的行为。这正是我们想要的。可以将行为分解为单独的、简单的行为，使它变得可组合、灵活和易读，如你所见。

第 32 行的最后一个接口名称是 NailDrivePuller：

## 清单 3

```go
32 type NailDrivePuller interface {
33     NailDriver
34     NailPuller
35 }
```

这个接口是从 NailDriver 和 NailPuller 接口组合而成。这是 Go 语言中一种非常常见的模式，调用已经存在的接口并组合它们成组合行为。稍后你会看到它在代码中怎么扮演的。现在，实现任何钉入和拔出行为具体类型的值同样会实现 NailDrivePuller 接口。

行为定义完成，是时候声明和实现一些工具了：

## 清单 4

```go
39 // 锤子是敲钉子的工具 .
40 type Mallet struct {}
41
42 // DriveNail 将钉子钉入指定木板 .
43 func (Mallet) DriveNail(nailSupply *int, b *Board) {
44     // 从钉子堆拿一枚钉子 .
45     *nailSupply-
46
47     // 钉一个钉子到木板里 .
48     b.NailsDriven++
49
50     fmt.Println("Mallet: pounded nail into the board.")

```
在清单 4 第 40 行，我们声明一个名称为 Mallet 的结构类型。这个类型声明为一个空结构，因为它是无状态的不需要维护。我们只需要实现行为。

棒槌是用来钉钉子的工具， 因此第 43 行，NailDriver 接口实现了经过 DriveNail 声明的方法。该方法的实现是不相关的，它使得钉子堆中的钉子减少 1 木板上钉子数量加 1。

让我们看看声明和实现的第二个工具：

## 清单 5

```go
53 // 撬棍是拔出钉子的工具
54 type Crowbar struct{}
55
56 // PullNail 从指定的木板上拔出一个钉子
57 func (Crowbar) PullNail(nailSupply *int, b *Board) {
58     // 从木板是拔出一枚钉子
59     b.NailsDriven-
60
61     // 把拔出的钉子放入钉子堆
62     *nailSupply++
63
64     fmt.Println("Crowbar: yanked nail out of the board.")
65 }
```

在清单 5，第 54 行我们声明了我们的第二个工具。这种类型表示一个撬棒，用来从木板上撬钉子。第 57 行，我们通过 PullNail 方法的声明实现了 NailPuller 接口。再次实现不相关但你能看到一个钉子是怎么从木板上减少并添加到钉子堆。

在代码的这个点上，我们已经通过我们的接口声明了工具的行为并通过一组结构类型实现了这些行为，这些结构类型代表承包商使用的两种独特的工具。现在，让我们创建一个类型，它代表使用这些工具完成工作的承包商。

## 清单 6

```go
69 // 承包商完成了固定板的工作。
70 type Contractor struct{}
71
72 // Fasten 将钉子钉入木板。
73 func (Contractor) Fasten(d NailDriver, nailSupply *int, b *Board) {
74     for b.NailsDriven < b.NailsNeeded {
75         d.DriveNail(nailSupply, b)
76     }
77 }
```

在清单 6，第 70 行我们声明了 Contractor 类型。同样，我们不需要任何状态因此我们使用一个空结构体。然后再第 73 行，我们看到 Fasten 方法，这是针对 Contractor 类型声明的三种方法之一。

Fasten 方法被声明来提供一个承包商的行为，该行为需要将一定数量的钉子钉入指定的木板。这个方法要求用户将实现 NailDriver 接口的值作为传入第一个参数。该值代表是承包商用来执行这个行为的工具。使用此参数接口类型允许 API 使用者后创建和使用不同的工具且不需要改变 API。用户提供工具的行为，Fasten 方法提供了工具使用时间和方式的工作流程。

注意我们使用的是 NailDriver 接口而不是 NailDriverPuller 接口作为参数类型。这非常重要因为 Fasten 方法所需的唯一行为是钉钉子的能力。通过只声明我们需要的行为，我们的代码更容易理解同时由于耦合度最小也更容易使用。现在，让我们看看 Unfasten 方法：

## 清单 7

```go
79 // Unfasten 将钉子从木板上拔出
80 func (Contractor) Unfasten(p NailPuller, nailSupply *int, b *Board) {
81     for b.NailsDriven > b.NailsNeeded {
82         p.PullNail(nailSupply, b)
83     }
84 }
```

在清单 7 中定义的 Unfasten 方法为承包商提供了与 Fasten 方法相反的行为。这个方法会从指定的木板上拔出尽可能多的钉子并将其添加到钉子堆。该方法只接受实现 NailPuller 接口的工具。这恰好是我们想要的因为这是实现方法需要的唯一行为。

承包商最终的行为是一个叫 ProcessBoards 的方法，它允许承包商一次在一组木板上工作：

## 清单 8

```go
86 // ProcessBoards works against boards.
87 func (c Contractor) ProcessBoards (dp NailDrivePuller, nailSupply *int, boards []Boards) {
88     for i := range boards {
89         b := &boards[i]
90
91         fmt.Printf("contractor: examing board #%d: %+v\n", i+1, b)
92
93         switch {
94         case b.NailsDriven < b.NailsNeeded:
95             c.Fasten(dp, nailSupply, b)
96
97         case b.NailsDriven > b.NailsNeeded:
98             c.Unfasten(dp, nailSupply, b)
99         }
100    }
101 }
```

在清单 8，我们看到 ProcessBoards 方法的声明和实现。该方法只接受同时实现 NailDriver 和 NailPuller 接口的值作为它的第一个参数：

## 清单 9

```go
87 func (c Contractor) ProcessBoards(dp NailDriverPuller, nailSupply *int, boards []Board) {
```

清单 9 展示该方法所需要的精确行为，通过 NailDriverPuller 接口类型声明。我们想要我们的 API 仅指定所需要或正在使用的行为。Fasten 和 Unfasten 需要一个实现单一行为的值但是 ProcessBoards 需要同时实现两个行为的值：

## 清单 10

```go
93         switch {
94         case b.NailsDriven < b.NailsNeeded:
95             c.Fasten(dp, nailSupply, b)
96
97         case b.NailsDriven > b.NailsNeeded:
98             c.Unfasten(dp, nailSupply, b)
99         }
```

清单 10 显示了 ProcessBoards 的一部分实现和在调用 Fasten 和 Unfasten 方法中怎么使用 NailDriverPuller 接口类型的值。在第 95 行当 Fasten 方法被调用时，NailDrivePuller 接口值传入该方法作为一个 NailDriver 类型的接口值。让我们再次看看 Fasten 方法的声明：

## 清单 11

```go
73 func (Contractor) Fasten(d NailDriver, nailSupply *int, b *Board) {
```

注意 Fasten 方法需要一个 NailDriver 接口类型的值我们却传入了一个 NailDrivePuller 接口类型的值。这是可以的因为编译器知道能被存储进 NailDrivePuller 接口值中的任何具体类型的值也必须实现 NailDriver 接口。因此，编译器接受方法调用和在这两种接口类型值之间赋值。

同样对 Unfasten 方法调用也是这样：

## 清单 12

```go
80 func (Contractor) Unfasten(p NailPuller, nailSupply *int, b *Board) 81 {
```

编译器知道存储在 NailDriverPuller 类型的接口值中任何具体类型的值也实现了 NailPuller 接口。因此，传入一个 NailDriverPuller 类型的接口值可以为 NailPuller 类型的接口赋值。因为在两种接口之间是静态关系，编译器可以放弃生成运行时类型断言而是生成一个接口转换。

承包商有了，我们现在可以声明一个新的类型，它声明承包商使用的工具盒：

## 清单 13

```go
105 // Toolbox 包含所有的工具
106 type Toolbox struct {
107     NailDriver
108     NailPuller
109
110     nails int
111 }
```

每个优秀的承包商都有工具箱，在清单 13 我们声明了工具箱。Toolbox 是一种结构体，它在第 107 行嵌入了 NailDriver 类型的接口值，在 108 行嵌入了 NailPuller 类型的接口值。然后在第 110 行通过 nails 字段声明了钉子堆。

当把一个类型嵌入到另一个类型时，最好考虑将一个新类型作为外部类型并将嵌入类型作为内部类型。这很重要因为你可以看到嵌入类型创建的关系。

任何嵌入类型总是作为一个内部值存在于外部类型值中。它永远不会失去自己的标识并且总是存在。然而，由于内部类型提升，在内部类型中声明的所有内容都被提升为外部类型。这意味着通过外部类型值，我们能直接根据输出规则访问内部类型值关联的任何字段或方法。

让我们看看这个例子：

## 清单 14
```go
01 package main
02
03 import "fmt"
04
05 // user 在程序中定义一个用户
06 type user struct {
07     name  string
08     email string
09 }
10
11 // notify 实现了一个可以通过调用的方法
12 // 一个 user 类型的指针
13 func (u *user) notify() {
14     fmt.Printf("Sending user email To %s<%s>\n",
15         u.name,
16         u.email)
17 }
```

在清单 14，我们声明了一个名为 user 的类型。这个类型包含两个字符串字段和一个名为 notify 的方法。现在，让我们把这个类型嵌入另一个类型：

## 清单 15

```go
19 // admin 代表一个权限用户
20 type admin struct {
21     user  // 嵌入类型
22     level string
23 }
```

现在我们可以看到第 15 行内部类型和外部类型之间的关系。user 类型现在是外部类型 admin 的内部类型。这意味着由于内部类型提升，我们可以从 admin 类型的值直接调用 notify 方法。但是由于内部类型值还存在，就其本身而言，我们也能直接从内部类型值调用 notify 方法：

## 清单 16

```go
25 // main() 函数是应用程序入口
26 func main() {
27     // 创建一个 admin user.
28     ad := admin{
29         user: user{
30             name:  "john smith",
31             email: "john@yahoo.com",
32         },
33         level: "super",
34     }
35
36     // 我们可以直接使用内部类型的方法。
37     ad.user.notify()
38
39     // 内部类型的方法被提升为外部类型。
40     ad.notify()
41 }
```

清单 16 第 28 行展示了使用结构体文字创建外部类型 admin 的值。在结构体文字内部使用第二层结构体文字创建并初始化内部类型 user 的值。admin 类型的值有了，我们可以直接调用 notify 方法，从第 37 行内部类型的值或者通过第 40 行外部类型的值。

最后一点，这既不是子类型也不是子类。admin 类型的值不能作为 user 类型的值使用。admin 类型的值仅仅是 admin 类型的值，user 类型的值仅仅是 user 类型的值。由于内部类型提升，内部类型的字段和方法能通过外部类型的值直接访问。

现在回到我们的工具箱：

## 清单 17

```go
105 // Toolbox 包含任何工具
106 type Toolbox struct {
107     NailDriver
108     NailPuller
109
110     nails int
111 }
```

我们没有嵌入任何结构类型到 Toolbox 而是嵌入两种接口类型。这意味着实现 NailDriver 接口的任何具体的类型值可以为 NailDriver 嵌入式接口类型作为内部类型值赋值。同样对于 NailPuller 嵌入式接口类型也适用。

一旦分配了具体类型，Toolbox 就可以保证实现这个行为。甚至，由于 toolbox 嵌入了 NailDriver 和 Nailpuller 接口类型，这意味着 Toolbox 同样也实现了 NailDrivePuller 接口：

## 清单 18

```go
31     NailDrivePuller interface {
32         NailDriver
33         NailPuller
34     }
```

在清单 18 我们又看到了 NailDrivePuller 的声明。嵌入式接口类型把内部类型提升和接口规范的概念提升到新的水平。

现在我们准备通过实现主函数把所有内容放在一起：

## 清单 19

```go
115 // main() 函数是应用程序入口
116 func main() {
117     // 需要移除的破木板和
118     // 取代它们的新木板清单
119     boards := []Board{
120         // 破木板将被移除。
121         {NailsDriven: 3},
122         {NailsDriven: 1},
123         {NailsDriven: 6},
124
125         // 新木板将被固定
126         {NailsNeeded: 6},
127         {NailsNeeded: 9},
128         {NailsNeeded: 4},
129     }
```

主函数在清单 19 的第 116 行开始。这里我们创建了一个 boards 切片和每一块木板需要的钉子数目。然后，我们创建工具箱：

## 清单 20

```go
131     // 装满工具箱。
132     tb := Toolbox{
133         NailDriver: Mallet{},
134         NailPuller: Crowbar{},
135         nails:      10,
136     }
137
138     // 显示我们工具箱和木板的当前状态
139     displayState(&tb, boards)
```

在清单 20，我们创建 Toolbox 的值并创建结构类型 Mallet 的值并将它赋值给内部类型 NailDriver。然后我们创建结构类型 Crowbar 的值并将他赋值给内部类型 NailPuller。最后，我们添加了 10 枚钉子到工具箱。

现在让我们创建承包商并处理一些木板：

## 清单 21

```go
141     // 雇佣一个承包商并让他工作。
142     var c Contractor
143     c.ProcessBoards(&tb, &tb.nails, boards)
144
145     // 显示我们的工具箱和木板的新状态。
146     displayState(&tb, boards)
147 }
```

在清单 21 第 142 行我们创建了一个 Contractor 类型的值并在 143 行针对它调用了 ProcessBoard 方法。让我们再看看 ProcessBoard 方法的声明：

## 清单 22

```go
87 func (c Contractor) ProcessBoards(dp NailDrivePuller, nailSupply *int, boards []Board) {
```

我们再次看到 ProcessBoard 方法怎样使用可以实现 NailDrivePuller 接口的任意具体类型的值作为它的第一个参数。由于 Toolbox 结构类型里嵌入了接口类型 NailDriver 和 NailPuller，ToolBox 类型的指针实现 NailDrivePuller 接口并且可以通过。

**注意**：Mallet 和 Crowbar 类型使用值接收器实现它们各自的接口。根据 [Method Sets](https://www.ardanlabs.com/blog/2014/05/methods-interfaces-and-embedded-types.html) 的规则，值和指针都满足接口。这是为什么我们可以创建并将具体类型的值赋值给嵌入式接口类型值。在 清单 21 第 143 行调用 ProcessBoards 时使用 Toolbox 值的地址因为我们想要共享同一个 Toolbox。但是，Toolbox 类型的值也满足接口。如果具体类型使用指针接收来实现接口，那么只有指针可以满足接口。

为了完整，这里是 displayState 函数的实现：

## 清单 23

```go
149 // displayState 提供所有木板的信息
150 func displayState(tb *Toolbox, boards []Board) {
151     fmt.Printf("Box: %#v\n", tb)
152     fmt.Println("Boards:")
153
154     for _, b := range boards {
155         fmt.Printf("\t%+v\n", b)
156     }
157
158     fmt.Println()
159 }
```

总之，这个例子试图展示组合是怎么应用和编写你的代码时该思考什么。怎么声明你的类型以及它们怎么协同工作是非常重要的。以下是我们帖子中涉及的内容总结：

* 首先声明行为集作为离散的接口类型。然后思考它们怎么组合成更大的行为。
* 确定每个函数或方法它们接受的是非常具体的接口类型。仅仅接受你在函数或方法中使用的行为的接口类型。这有助于指定所需的更大的接口类型。
* 思考嵌入作为内部类型和外部类型的关联。记住通过内部类型提升，所有声明在内部的类型都被提升为外部类型。然而，内部类型的值本身就存在，并且总是根据输出规则访问。
* 嵌入类型既不是子类型也不是子类。具体类型的值代表一个单独的类型并且不能根据任何嵌入关系赋值。
* 编译器可以安排接口在关联的接口值之间转换。在编译时，接口转换不关心具体的类型，它只知道根据接口类型本身做什么，不是实现它们包含的具体值。

*我要感谢 [Kevin Gillette](https://twitter.com/kevingillette) 和我一起码代码和帖子还有  [Bill Hathaway](https://twitter.com/billhathawa)。帮我评论和编辑帖子。*

---

via: https://www.ardanlabs.com/blog/2015/09/composition-with-go.html
作者：[William Kennedy](https://www.ardanlabs.com/go-training/)
译者：[WillsYoung](https://github.com/WillsYoung)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创翻译，[Go 中文网](https://studygolang.com/) 荣誉推出
