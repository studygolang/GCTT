已发布：https://studygolang.com/articles/12932

# Go 语言的数据结构 ：栈与队列

在[先前的博文](https://studygolang.com/articles/12686)中，我们探讨了链表以及如何将它应用于实际应用。在这篇文章中，我们将继续探讨两个相似且功能强大的数据结构。

## 建模操作和历史

让我们看看 Excel 或 Google 文档，他们是人类发明的最普遍的构成文件的应用程序。我们都使用过它们。 正如你可能知道的，这些应用程序有各种各样对文本的操作。
比如在文本中添加颜色、下划线、各种字体和大小，或者在表格中组织内容。这个列表很长，我们期望从这些工具中得到一个普遍的功能 —— “撤销”和“重做”已经执行了的操作的能力。

你是否考虑过让你做，你将如何规划这样的功能？下面让我们探索一个可以帮助我们完成这样一项任务的数据结构。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/data-structures/excel.png)

让我们试着想想如何为这些应用程序的操作建模。此外，稍后我们将看到如何保存动作的历史并“撤销”它们。

一个简单的 `Action struct` 看起来像这样 ：

```go

type Action struct {
	name string
	metadata Meta
	// Probably some other data here...
	// 这里或许还有一些其他数据
}

// 译者注 ：Action 结构体需要添加一个 next 字段，否则下面代码的部分操作不成立
type Action struct {
	name string
	metadata Meta
	// 这里或许还有一些其他数据
	next *Action
}

```

我们现在只存储名称，而相对来说实际应用中会有更多元素。现在，当编辑器的用户将一个函数应用到一堆文本中时，我们希望将该操作存储在某个集合中，以便稍后可以“撤消”它。

```go
type ActionHistory struct {
	top *Action
	size int
}
```

这个 `ActionHistory` 数据结构，我们存储一个在堆栈顶部指向 `Action` 的指针以及堆栈的大小。每当一个动作被执行，我们将把它链接到 `top` 的操作上。因此，当对文档应用一个操作时，这是可以在后台运行的。

```go
func (history *ActionHistory) Apply(newAction *Action) {
	if history.top != nil {
		oldTop := history.top
		newAction.next = oldTop
	}
	history.top = newAction
	history.size++
}
```

`Add` 函数（译者注：此处应为上文中的 `Apply`）会将最新的 `Action` 添加到 `ActionHistory` 的顶部。如果历史结构在顶部有一个动作，它将通过将它与新的动作联系起来，将它往下压。否则，它将把新操作附加到列表的顶部。现在，如果您知道链表(或者阅读我最近的一篇关于链表的文章)，你可能会发现他们的相似之处。到目前为止，我们在这里使用的基本上依然是一个链表。

那么撤销操作会是怎样的呢？如下是一个 `Undo` 函数的实现：

```go
func (history *ActionHistory) Undo() *Action {
	topAction := history.top
	if topAction != nil {
		history.top = topAction.next
	} else if topAction.next == nil {
		history.top = nil
	}
	historyAction.size--	//historyAction 没有定义，应该是作者笔误
	return topAction
}

//-------
//译者注：此处代码存在问题，建议修改如下。
```

```go
func (history *ActionHistory) Undo() *Action {
	topAction := new(Action)
	if history.size > 0 {
		topAction = history.top
		history.top = topAction.next
		history.size--
	}
	return topAction
}
```
感谢 [无闻](https://github.com/Unknwon) 的建议。

如果你仔细观擦，你会注意到这与从链表中删除一个节点有一点不同。由于一个 `ActionHistory` 的性质，我们希望最后一个被执行了的动作是最先被撤销的,这才是我们所希望实现的。

这是堆栈的基本行为。堆栈是一种数据结构，您只能在堆栈顶部插入或删除元素。把它想象成一堆文件，或者你厨房抽屉里的一堆盘子。如果你想从那一堆盘子取出最下面的盘子，那是挺难的。但是拿最上面的那个是简单的。堆栈也被认为是 `LIFO` 结构 —— 意思是后进先出，我们前面解释过那是为什么。

这基本上就是我们的 `Undo` 函数所处理的。如果堆栈（或者说`ActionHistory`）有多个 `Action` ，它将为第二项设置顶部链接。否则，它将清空 `ActionHistory`，将 `top` 元素设置为 `nil`。

从 `Big-O` 表示法来看，在堆栈中搜索的复杂度是 `O(n)`，但是在堆栈中插入和删除是非常快的复杂度是 `O(1)`。
这是因为遍历整个堆栈，在最坏的情况下，仍然会在其中执行所有的 `n` 项，而插入和删除元素的时间复杂度是常量时间，因为我们总是从堆栈的顶部插入和删除。

你可以在[*这里*](https://play.golang.org/p/Eu8_-HTDBY_A) 使用该代码的工作版本。

## 行李控制

我们大多数人都是坐飞机旅行的，而且知道所有人员都必须通过安检才能上飞机。当然，这是为了我们的安全，但有时进行全部的扫描、检查和测试是不必要的。

机场安检点的一个常见场景是安检人员排起长龙，行李放在 x 光机的带子上，而人们则通过金属探测器门。也许我们对这些不甚了解，但是让我们关注一下扫描我们的袋子的 x 光机。

你有没有想过，你会如何模拟这台机器上发生的相互作用？当然，这些至少是看得见的。让我们来探讨一下这个想法。我们必须以某种方式将行李上的行李作为物品的集合，而 x 光机一次扫描一件行李。

`Luggage` 结构体如下 ：

```go
type Luggage struct {
	weight int
	passenger string
}
```

与此同时，我们为 `Luggage` 类型添加简单的构造函数：

```go
func NewLuggage(weight int, passenger string) *Luggage {
	l := Luggage{
		weight:    weight,
		passenger: passenger, // just as an identifier
	}
	return &l
}
```

接着，我们创建一个 `Belt` （流水线），让 `Luggage` 放到上面并通过 X 光的检测。

```go
type Belt []*Luggage
```

不是你想要的？我们所创建的是一个 `Belt` 类型，实际上是 `Luggage` 指针的一部分。这就是所谓的传送带 —— 仅仅是一堆被逐一扫描的袋子。

所以现在我们需要添加一个知道如何将 `Luggage` 添加到 `Belt` 的函数:

```go
func (belt *Belt) Add(newLuggage *Luggage) {
	*belt = append(*belt, newLuggage)
}
```

既然 `Belt` 实际上是一个切片，那么我们就可以用 Go 语言内建函数 `append` 将 `newLuggage` 添加到 `Belt` 上。这个实现很奇妙的部分是时间复杂度 -- 因为我们使用了 `append` 这个内建函数，所以插入操作的时间复杂度是 O(1)。
当然，这有一定的控间浪费，这是一位 go 语言切片的工作原理造成的。

当 `Belt` 开始运动并且将 `Luggage` 带到 X 光机上，我们需要将行李拿下来并且装进机器进行检查。
鉴于 `Belt` 的自然属性，第一个放到传送带上面的行李是第一个被扫描监测的。
自然地，最后一个放到传送带上的是最后一个被扫描的。所以我们可以说 `Belt` 是一个 FIFO（先进先出）的数据结构体。

请留意上述的细节并看看如下 `Take` 函数的实现：

```go
func (belt *Belt) Take() *Luggage {
	first, rest := (*belt)[0], (*belt)[1:]
	*belt = rest
	return first
}
```

这个函数它取走了第一个元素并且将其返回，并且它会把集合中的其他东西都分配到它的开头，所以它的第二个元素就会变成第一个，以此类推。
你会发现，从队列中取走第一个元素的时间复杂度是 `O(1)`。

使用我们新的类型和函数能够进行以下操作：

```go
func main() {
	belt := &Belt{}
	belt.Add(NewLuggage(3, "Elmer Fudd"))
	belt.Add(NewLuggage(5, "Sylvester"))
	belt.Add(NewLuggage(2, "Yosemite Sam"))
	belt.Add(NewLuggage(10, "Daffy Duck"))
	belt.Add(NewLuggage(1, "Bugs Bunny"))

	fmt.Println("Belt:", belt, "Length:", len(*belt))
	first := belt.Take()
	fmt.Println("First luggage:", first)
	fmt.Println("Belt:", belt, "Length:", len(*belt))
}
```

`main` 函数的输出大致如下：

```
Belt: &[0x1040a0c0 0x1040a0d0 0x1040a0e0 0x1040a100 0x1040a110] Length: 5
First luggage: &{3 Elmer Fudd}
Belt: &[0x1040a0d0 0x1040a0e0 0x1040a100 0x1040a110] Length: 4
```

基本上，我们在 `Belt` 上加了5个不同的 `Luggage`，然后我们取出第一个元素，它在屏幕的第二行输出显示了。

你可以在[*这里*](https://play.golang.org/p/DTFUkWeZ4H8)使用实例代码。

## 头等舱的乘客 ？

恩，没错，是他们。是的,他们呢?我的意思是，他们已经花了那么多钱买机票，他们在经济舱的排队中去等行李是不合理的。
那么，我们该如何优先考虑这些乘客呢？如果他们的行李有某种优先权，优先级越高，他们通过队列越快？

让我们对 `Luggage` 结构体进行修改，如下：

```go
type Luggage struct {
	weight    int
	priority  int
	passenger string
}
```
当然，我们使用 `newLuggage` 函数创建 `luggage` 的时候会加入 `priority` 作为参数。

```go
func NewLuggage(weight int, priority int, passenger string) *Luggage {
	l := Luggage{
		weight:    weight,
		priority:  priority,
		passenger: passenger,
	}
	return &l
}
```

让我们再想想。基本上，当一个新的 `Luggage` 被放在 `Belt` 上时，我们需要检测它的 `priority`，并根据 `priority` 把它放在 `Belt`的最前面。

我们在修改一下 `Add` 函数：

```go
func (belt *Belt) Add(newLuggage *Luggage) {
	if len(*belt) == 0 {
		*belt = append(*belt, newLuggage)
	} else {
		added := false
		for i, placedLuggage := range *belt {
			if newLuggage.priority > placedLuggage.priority {
				*belt = append((*belt)[:i], append(Belt{newLuggage}, (*belt)[i:]...)...)
				added = true
				break
			}
		}
		if !added {
			*belt = append(*belt, newLuggage)
		}
	}
}
```

与之前的实现相比，这是相当复杂的。这里要处理多种情况，第一种情况相对是简单的。如果皮带是空的，我们就把新的行李放在传送带上就可以了。 `Belt` 上只有一件东西，那第一个拿走就行了。

第二种情况是在 `Belt` 上有不知一个元素，我们要遍历 `Belt` 上的所有行李并且与将要加进来的行李进行优先级比较。
当找到一个优先级比它小的行李的时候，那么就会绕过这个优先级小的行李，并且把新的行李放到它的前面。
这就意味着优先级越高的行李，将会在 `Belt` 的越靠前位置。

当然，如果遍历没有找到这样的行李，它会把它附加到 `Belt` 的末端。
我们新的 `Add` 函数的时间复杂度是 `O(N)`,这是因为在最坏的情况下，我们往 `Luggage` 结构插入一个新的元素可能要遍历整个切片。
从本质上说，搜索和访问队列中的任何项都是相同的复杂度 `O(n)`。

为了演示新的添加功能，我们可以运行以下代码:

```go
func main() {
	belt := make(Belt, 0)
	belt.Add(NewLuggage(3, 1, "Elmer Fudd"))
	belt.Add(NewLuggage(3, 1, "Sylvester"))
	belt.Add(NewLuggage(3, 1, "Yosemite Sam"))
	belt.Inspect()

	belt.Add(NewLuggage(3, 2, "Daffy Duck"))
	belt.Inspect()

	belt.Add(NewLuggage(3, 3, "Bugs Bunny"))
	belt.Inspect()

	belt.Add(NewLuggage(100, 2, "Wile E. Coyote"))
	belt.Inspect()
}
```

首先我们创建一个有三个 `Luggage` 的 `Belt` ，这些 `Luggage` 的优先级都是 1 ：

```
0. &{3 1 Elmer Fudd}
1. &{3 1 Sylvester}
2. &{3 1 Yosemite Sam}

```

然后我们添加一个 优先级为 2 的 `Luggage`：

```
0. &{3 2 Daffy Duck}
1. &{3 1 Elmer Fudd}
2. &{3 1 Sylvester}
3. &{3 1 Yosemite Sam}
```

你看，带着最高优先级的新行李被提升到 `Belt` 上的第一个位置。接下来，我们再添加一个具有更高优先级（3）的新元素:

```
0. &{3 3 Bugs Bunny}
1. &{3 2 Daffy Duck}
2. &{3 1 Elmer Fudd}
3. &{3 1 Sylvester}
4. &{3 1 Yosemite Sam}
```

正如预期的那样，优先级最高的那一个被放在了 `Belt` 的第一个位置。最后，我们再加一件行李，它的优先级为 2 :

```
0. &{3 3 Bugs Bunny}
1. &{3 2 Daffy Duck}
2. &{100 2 Wile E. Coyote}
3. &{3 1 Elmer Fudd}
4. &{3 1 Sylvester}
5. &{3 1 Yosemite Sam}
```

新的 `Luggage` 会被添加到优先级相同的  `Luggage`  的后面，当然，不是 `Belt` 的开始位置。
总的来说，当我们往 `Belt` 上添加 `Luggage` 都会被排序。

如果你对队列有一定的了解，那么你可能会认为这些并不是实现优先级队列最有效的方法，你是完全正确的。实现优先级队列可以更高效地使用堆，我们将在另一个博文中进行谈论。

我们可以探索更多关于优先队列有趣的知识。你可以查看优先队列的  [Wiki](https://en.wikipedia.org/wiki/Priority_queue) 页面。
如果你对队列有一定的了解，那么你可能会认为这些并不是实现优先级队列最有效的方法，你是完全正确的。实现优先级队列可以更高效地使用堆，我们将在另一篇文章中对此进行介绍，特别是“实现”部分。

你可以在[这里](https://play.golang.org/p/Eu8_-HTDBY_A)查看并使用示例代码。

---

via: https://ieftimov.com/golang-datastructures-stacks-queues

作者：[Ilija Eftimov](https://ieftimov.com/about)
译者：[SergeyChang](https://github.com/SergeyChang)
校对：[无闻](https://github.com/Unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
