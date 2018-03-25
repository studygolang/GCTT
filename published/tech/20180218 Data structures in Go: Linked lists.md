已发布：https://studygolang.com/articles/12686

# Go 语言的数据结构：链表

数据结构和算法是计算机科学的重要组成部分。虽然有时候它们看起来很吓人，但大多数算法都有简单的解释。同样，当问题能用算法来解释清楚的时候，算法的学习和应用也会很有趣。

这篇文章的目标读者是那些对链表感到不舒服的人，或者那些想要看到并学习如何用 Golang 构建一个链表的人。我们将看到如何通过一个（稍微）实际的例子来实现它们，而不是简单的理论和代码示例。

在此之前，让我们来谈谈一些理论。

## 链表

链表是比较简单的数据结构之一。维基百科关于链接列表的文章指出：

> 在计算机科学中，链表是数据元素的线性集合，其中线性顺序不是由它们在内存中的物理位置所给出的。相反，每个元素指向下一个元素。它是由一组节点组成的数据结构，它们共同代表一个序列。在最简单的形式下，每个节点都由数据和一个指向下个节点的引用（换句话说，一个链接）组成。

尽管这些看起来可能太过或令人困惑，让我们把它分解一下。 线性数据结构，是一种其元素组成某种序列的数据结构。 就这么简单。为什么记忆中的物理位置不重要呢？ 当你有数组时，数组的内存数量是固定的，就是说，如果你有一个 5 项的数组，语言只会在内存中只抓取 5 个内存地址，一个接一个。 因为这些地址创建一个序列，数组知道它的值将存储在什么内存范围内，因此这些值的物理位置将创建一个序列。

有了链表，就有点不同了。在定义中，您将注意到“每个元素指向下一个”，使用“数据和引用（换句话说，就是链接）指向下一个节点”。这意味着链接列表的每个节点存储两个东西：一个值和一个指向列表中的下一个节点的引用。就这么简单。

## 数据流

人类所感知到的一切都是某种信息或数据，我们的感官和头脑知道如何处理并将其转化为有用的信息。 不管我们是看，闻，还是摸，都是我们在处理数据，并从数据中找到意义。当我们浏览我们的社交媒体网络时，我们总是求助于数据，按时间顺序排列，有看不完的信息。

那么，我们如何使用链表来建模这样的新闻流呢? 让我们先快速浏览一下简单的 Tweet，例如:

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/data-structures/tweet_jack.png)

实例我们社交网络的目的，我们从 Twitter 获得灵感并创建一个 `Post` 类型，它有一个 `body`，一个 `publishDate` 和一个 `next` 帖子的链接:

```go
type Post struct {
	body         string
	publishDate int64 // Unix timestamp
	next *Post // link to the next Post
}
```

接下来，我们如何建模一个帖子的提要？如果我们知道数据流是由一个连着另一个的帖子组成，那么我们可以试着创建一个这样的类型：

```go
type Feed struct {
	length int // we'll use it later
	start  *Post
}
```

`Feed` 结构将有一个开始(或 `start`)，它指向提要中的第一个 `Post` 和一个 `length` 属性，该属性将在任意时刻存储 `Feed` 的大小。

因此，假设我们想创建一个有两个帖子的 `Feed`，第一步是在 `Feed` 类型上创建一个 `Append` 函数:

```go
func (f *Feed) Append(newPost *Post) {
	if f.length == 0 {
		f.start = newPost
	} else {
		currentPost := f.start
		for currentPost.next != nil {
			currentPost = currentPost.next
		}
		currentPost.next = newPost
	}
	f.length++
}
```

然后我们可以两次调用它:

```go
func main() {
	f := &Feed{}
	p1 := Post{
		body: "Lorem ipsum",
	}
	f.Append(&p1)

	fmt.Printf("Length: %v\n", f.length)
	fmt.Printf("First: %v\n", f.start)

	p2 := Post{
		body: "Dolor sit amet",
	}
	f.Append(&p2)

	fmt.Printf("Length: %v\n", f.length)
	fmt.Printf("First: %v\n", f.start)
	fmt.Printf("Second: %v\n", f.start.next)
}
```

那么这段代码是用来干嘛的呢？首先，`main` 函数 - 创建一个指向 `Feed` 结构的指针，两个 `Post` 结构包含一些虚构的内容，它两次调用 `Feed` 上的 `Append` 函数，使得它的长度为 2 。我们检查 `Feed` 的两个值，它访问 `Feed` (实际上是 `Post` )的 `start` 和 `start` 后的 `next` 项，这是第二个 `Post` 。

当我们运行程序时，输出将会是:

```
Length: 1
First: &{Lorem ipsum 1257894000 <nil>}
Length: 2
First: &{Lorem ipsum 1257894000 0x10444280}
Second: &{Dolor sit amet 1257894000 <nil>}
```

可以看出，当我们在 `Feed` 添加第一个 `Post` 后，它的长度为 `1` 并且第一个 `Post` 拥有一个 `body` 和一个 `publishDate` （作为 Unix 时间戳），与此同时，它的 `next` 值为 `nil` 。 然后，我们将第二个 `Post` 添加到 `Feed` 中，当我们查看两个 `Posts` 时，我们会看到第一个 `Post` 的内容与之前的内容相同，但它的指针指向列表中的下一个 `Post` 。 第二个 `Post` 也有一个 `body` 和一个 `publishDate` ，但是没有指向列表中的下一个 `Post` 的指针。 此外，当我们添加更多的 `Posts` 时，`Feed` 的长度也会增加。

现在让我们回过头来看 `Append` 函数并解构它，这样我们就能更好地理解如何使用链表。 首先，该函数创建一个指向 `Post` 值的指针，将 `body` 参数作为  `Post`的 `body`，并将 `publishDate` 设置为当前时间的 Unix 时间戳表示。

然后，我们检查 `Feed` 的`length`是否为 `0` — 这意味着它没有 `Post` 。第一个被添加的 `Post` 会被设为起始 `Post`，为方便起见，我们把它命名为 `start`。

但是，如果 `Feed` 的长度大于 0 ，那么我们的算法就会发生不同的变化。 它将从 `Feed` 的 `start` 开始，它将遍历所有的 `Post`，直到找到一个没有指向 `next` 的指针。 然后，它将把新的 `Post` 附加到列表的最后一个 `Post` 上。

## 优化 `Append`

想象一下，我们有个用户刷 `Feed`，就像其他社交网络一样。 由于文章是按时间顺序排列的，基于 `publishDate` ，`Feed` 会随着用户的滑动而变得越来越多，更多的 `Post` 会被附加到 `Feed` 上。 考虑到这种方法，我们采用了 `Append` 函数，因为 `Feed` 变得越来越长，`Append` 函数将会付出越来越沉重的代价。 为什么? 因为我们必须遍历整个 `Feed` ，在末尾添加一个 `Post` 。如果你听说过 `Big-O` 表示法，这个算法有一个 `O(n)` 的时间复杂度，这意味着它在添加 `Post` 之前总是遍历整个 `Feed` 。

您可以想象，这可能非常低效，特别是如果“Feed”增长相当长。如何改进“追加”函数，降低其[渐近复杂](https://en.wikipedia.org/wiki/Asymptotic_computational_complexity)性?

因为我们的 `Feed` 数据结构只是一个 `Post` 的列表，要遍历它，我们必须知道列表的开头(称为 `start` )，它是 `Post` 类型的指针。 因为在我们的示例 `Append` 中总是添加一个 `Post` 到 `Feed` 的末尾，如果 `Feed` 不仅知道它的起始 `start` 元素，而且还知道它的 `end` 结束元素，那么我们可以大大提高算法的性能。 当然，对于优化总是有一个折衷，这里的权衡是数据结构将消耗更多的内存(对于 `Feed` 结构的新属性)。

扩展我们的 `Feed` 数据结构是轻而易举的：

```go
type Feed struct {
	length int
	start  *Post
	end    *Post
}
```

但是，我们的 `Append` 算法必须被调整，以适应 `Feed` 的新结构。这是使用 `Post`的 `end` 属性的 `Append` 的版本:

```go
func (f *Feed) Append(newPost *Post) {
	if f.length == 0 {
		f.start = newPost
		f.end = newPost
	} else {
		lastPost := f.end
		lastPost.next = newPost
		f.end = newPost
	}
	f.length++
}
```

这看起来简单一点，对吧？让我给你一些好消息:

1. 现在代码更简单、更短了。
2. 我们极大地提高了函数的时间复杂度。

我们在重新审视一下算法，它做了两件事情：如果 `Feed` 为空，它就会设置一个新的 `Post` 作为 `Feed` 的开头和结尾，反之它会设置一个新的 `Post` 作为 `end` 项并且依附到链表中先前的 `Post` 。很重要的一点是它很简单，并且算法复杂度为 `O(1)` ,也称为常数时间复杂度。这意味着无论 `Feed` 结构的长度如何，`Append` 都将执行相同的操作。

很简单，对吧？但让我们想象一下，`Feed` 实际上是我们的配置文件中的 `Post`列表。因此，它们是我们的，我们应该能够删除它们。我的意思是，什么样的社交网络不允许用户(至少)删除他们的帖子?

## 移除一个 `Post`

正如我们在前一节中建立的，我们希望我们的 `Feed` 用户能够删除他们的帖子。那么，我们如何建立模型呢?如果我们的 `Feed` 是一个数组，我们就会删除该条目并对其进行处理，对吧？

这事实上就是链表闪耀的地方。当数组大小改变时，运行时会捕获一个新的内存块来存储数组的项。 由于其设计的链表，每个条目都有一个指向列表中的下一个节点的指针，并可以分散到整个内存空间中，从空间的角度来看，从列表中添加或者删除节点是低开销的。 当一个人想要从一个链表中移除一个节点时，只需要连接被删除节点的邻居节点。 垃圾收集语言(如 Go 语言)使这个更加容易，因为我们不必担心释放被分配的内存 ——  GC 将启动并删除所有未使用的对象。

为了让我们操作方便，让我们给每个 `Feed` 上的 `Post` 设置一个限制，它将有一个独特的 `publishDate` 。这意味着发布者可以在他们的 `Feed` 上每秒钟创建一个 `Post` 。将其付诸实施，我们可以很容易地从 `Feed` 中删除 `Post`:

```go
func (f *Feed) Remove(publishDate int64) {
	if f.length == 0 {
		panic(errors.New("Feed is empty"))
	}

	var previousPost *Post
	currentPost := f.start

	for currentPost.publishDate != publishDate {
		if currentPost.next == nil {
			panic(errors.New("No such Post found."))
		}

		previousPost = currentPost
		currentPost = currentPost.next
	}
	previousPost.next = currentPost.next

	f.length--
}
```

`Remove` 函数将把 `Post` 作为一个 `publishDate` 作为一个参数，它将检测哪些 `Post` 需要删除(或未链接)。 这个函数很小。如果它检测到 `Feed` 的 `start` 项将被删除，它将会重新分配 `Feed` 的 `start`，并在 `Feed` 中添加第二个 `Post` 。 否则，它会跳转到 `Feed` 中的每个 `Post` ，直到它遇到一个 `Post`， 该 `Post` 有一个匹配的 `publishDate` 作为函数参数。 当它找到一个时，它会把之前的和下一个 `Post` 连接在一起，有效地从 `Feed` 中删除中间(匹配)一个。

有一个边界情况，我们需要确保我们在 `Remove` 函数中覆盖 —— 如果 `Feed` 没有带有指定的 `publishDate` 的 `Post` ？ 为了简单，函数会检查 `Feed` 中的下一个 `Post` ，然后再跳到它。如果下一个是 nil 的函数，它告诉我们它找不到一个 `publishDate` 的 `Post` 。

## 插入一个 `Post`

现在我们有了添加和移除的方法，让我们来看看一些假设的情形。 假设生成 `Post` 的来源并不是按照时间顺序将他们发送到我们的应用程序的。 这就意味着需要基于 `publishDate` 将 `Post` 放到 `Feed` 中合适的位置。比如这样：

```go
func (f *Feed) Insert(newPost *Post) {
	if f.length == 0 {
		f.start = newPost
	} else {
		var previousPost *Post
		currentPost := f.start

		for currentPost.publishDate < newPost.publishDate {
			previousPost = currentPost
			currentPost = previousPost.next
		}

		previousPost.next = newPost
		newPost.next = currentPost
	}
	f.length++
}
```

本质上，这是一个非常类似于 `Remove` 函数的算法，因为尽管它们都做了一件非常不同的事情(在 `Feed` 中添加 v.s. 删除 `Post`)，它们都是基于搜索算法的。 这意味着，这两个函数实际上遍历整个 `Feed` ，搜索与 `publishDate` 匹配的 `Post `，并在函数的参数中接收到一个 `Post` 。 唯一的区别是，`Insert` 实际上会在日期匹配的地方放置新的 `Post` ，而 `Remove` 将从 `Feed` 中删除 `Post` 。

此外，这意味着这两个函数都具有相同的时间复杂度，即 O(n) 。 这意味着在最坏的情况下，函数必须遍历整个 `Feed` 才能到达需要插入新 `Post` (或删除)项。

## 如果我们使用数组呢？ 

如果你问自己，让我先说，你有一个观点。 的确，我们可以将所有的 `Post` 存储在一个数组中(或者是一个 Go 语言的 slice )，可以轻松地将条目推到它上面，甚至还可以使用 O(1) 复杂性随机访问。

鉴于数组的性质，它的值必须存储在内存中，所以读取速度非常快而且开销很低。 一旦你有了存储在数组中的东西，就可以用它的 0-based 索引来获取它。 在插入一个条目时，无论是在中间还是在最后，数组的效率都不如链表。 这是因为如果数组没有为新项保留更多的内存，它将不得不保留它并使用它。 但是，如果下一个内存地址不是空闲的，它将不得不“移动”到一个新的内存地址，只有那样才有空间容纳它的所有项(新和旧的)。

看看我们到目前为止的所有例子和讨论，我们可以为我们创建的每一个算法创建一个具有时间复杂度的链表，并将它们与数组的相同算法进行比较:

| Action  | Array | Linked list |
| ------- |:-----:| -----------:|
| Access  |  O(1) |     O(n) |
| Search  |  O(n) |     O(n) |
| Prepend |  O(1) |     O(1) |
| Append  |  O(n) |     O(1) |
| Delete  |  O(n) |     O(n) |


正如你所看到的，当面对一个特定的问题时，选择正确的数据结构可以真正地成就或毁掉你所创建的产品。 对于不断增长的 `Feed`，插入 `Post` 是最重要的，链表会做得更好，因为插入非常代价更小。 但是，如果我们的手上有一个不同的问题需要频繁的删除或大量的检索/搜索，那么我们就必须为我们正在处理的问题选择正确的数据结构。

你可以看到 `Feed` 的整个实现，并在[这里](https://play.golang.org/p/fqLPjf_ekD6)体验它。另外，Go 语言也有自己的链表实现，它已经内置了一些不错的功能。你可以在[这里](https://golang.org/pkg/container/list/)看到它的文档。

----------------

via: https://ieftimov.com/golang-datastructures-linked-lists

作者：[Ilija Eftimov](https://ieftimov.com/about)
译者：[SergeyChang](https://github.com/SergeyChang)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
