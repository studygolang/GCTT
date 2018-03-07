已发布：https://studygolang.com/articles/12486

# Go 语言 Malloc 的惯用语法

我终于又开始使用 Go 语言编程了。虽然我在前两年多的时间里积极参与这个项目，但从 2012 年起，我就基本没有参加过这个项目。最初，我之所以做出贡献，是因为我是贝尔实验室 [Plan 9](http://9p.io/plan9/)(操作系统) 和 [FreeBSD](https://www.freebsd.org/) 的粉丝。我喜欢可用的、基于 csp 的语言，但是 Go 最初的版本只能在 Linux 和 OS X 上运行。那时候我只有 FreeBSD 系统，因此，我将编译器工具链、运行时和标准库移植到 FreeBSD (有很多调试成果来自 [Russ Cox](https://swtch.com/~rsc/) )。

然而，过去我的大部分工作都是在低延迟的系统软件上,它们大部分是用 C 语言编写的，自2007年起，我所有的雇主都不再支持 FreeBSD。因为我并没有真正的机会用 Go 语言去编写新的软件，而且我最终也不再对维护一个操作系统的知识感兴趣，我只是为了好玩，所以我对 Go 语言的使用和贡献都被搁置了。

现在我在谷歌工作，我终于有机会用 Go 语言写代码了。虽然我仍然喜欢这门语言，但有一些[经验报告](https://github.com/golang/go/wiki/ExperienceReports),例如风格那样的东西结果阻止了我在过去的 5~6 年里使用这门语言，而我现在觉得有些麻烦。在一些同事的建议下，我想我应该至少记录下其中的一个。

我想念的一件事情是 C 语言 `malloc` 那样的习惯用法。在 C 语言中，分配的内存通常是对 `mallocs -family` 函数的调用，它至少给您足够的内存来完成您想要的功能，或者并不给您分配内存。惯用语法大致是这样的:

```c
T *p = malloc(sizeof *p);
```

注意，`p (T *)` 的类型只出现一次。这一行代码利用了它的操作数时的大小，并且显示了一个指针 —— 这其实是没有发生的事情，因为操作符 `sizeof` 的结果必须<sup>①</sup>在编译时是可识别的。这种编程语言而不是语法定义的结果是 `sizeof` 产生了指向对象的大小;它不会变成运行时的东西。C 语言的好处是，如果我改变用 `T` 表示的类型，我只改变声明或定义中的类型。在 `site(s)` 中不需要做任何更改，指针对象被分配给内存分配的结果。上面的示例很简单，让我们考虑一个更复杂的情况，结构成员指向某种类型:

```c
struct set {
	size_t cap;
	size_t nmemb;
	int members[];
}

struct set *
set_create(size_t sz)
{
	struct set *n = malloc(sizeof *n + (sz * sizeof *n->members));
	if (n == NULL) {
		return NULL;
	}

	n->members = malloc(sz * sizeof *n->members);
	if (n->members == NULL) {
		free(n);
		return NULL;
	}

	n->cap = sz;
	n->nmemb = 0;

	return n;
}
```

如果以后我们想要更改 `struct set` 来支持除 int 以外的成员，我们可能会将成员更改为一个union，并添加一些 enum 类型来指定一些我们想要添加的字段。我们可以在不改变 `set_create` 中的任何代码的情况下做到这一点。

每次我使用 Go 语言创建了一些结构类型,当需要嵌入一些像 slice 和 map 那样需要分配内存的字段的时候都让我很抓狂。在 Go 中,我们被迫重复表达我们想要分配的东西的类型,尽管编译器熟知这种类型而且类型推断是符合语言习惯的(试想一下如这样的表达式 `a:= b` ),我有时不得不深究一下嵌入字段的类型是什么。让我们来看看在创建一个嵌入了 map 的结构体所涉及的内容:

```go
type NamedMap struct {
	name string
	m    map[string]string
}

func NewNamedMap(name string) *NamedMap {
	return &NamedMap{name: name, m: map[string]string{}}
}
```

我们还可以在 `NewNamedMap` 中使用 `make` ，但是仍然保留了`return &NamedMap{name: name, m: make(map[string]string)}` — 再次，重复它的类型。经过深思熟虑的代码，应该只有一个(额外的)地方需要我们指定类型来分配它，但是当类型改变时，这仍然需要多处改动代码。当我在做原型的时候，这就会让我抓狂，而且我还没有充分考虑到我需要保存在 map 中的状态。我发现在很多地方需要自己手动将 `map[string]string` 更改为 `map[string]T`，每次我需要更改多行代码时，它都会使我感到困扰。

有人可能会说，在写代码之前，我应该多考虑一下我需要什么，那样会更好。但我仍然会反驳说，在项目的生命周期中开发额外的状态需求并不少见，比如在上面的例子中。随着时间的推移，系统的约束也可能会发生变化，这样一种最初非常好的类型最终会变得不可用。在 Go 中，set 结构可能是这样的:

```go
type Set struct {
	nmemb   int
	cap     int
	members []int
}

func NewSet(sz int) *Set {
	return &Set{cap: sz, members: []int{}}
}
```

你可能会问为什么我不只是用一个 slice，答案是这是一个演示这个问题的简单例子。不管怎样，我们以后可能想要支持不同类型的 slice，那样我们又遇到了之前的问题。set 中如果有一个 slice，我们可能可以忽略初始化，假设我们在添加后只从 slice 中读取。由于形如 map 和 channel 那样的类型，因为我们必须在使用前进行分配，所以会使得情况更加复杂。那么在某个地方重复输入信息并不罕见。

我不知道该如何解决这个问题。对于复合文本，可能可以添加如下语法:

```go
return &Set{cap: sz, members: Set.members{}}
```

如果你有一个指向 slice 的指针:

```go
return &Set{cap: sz, members: &Set.members{}}
```

我不知道我是否喜欢这些复合文字语法。在 C 语言中，为支持 `sizeof` 行为而进行的修正感觉更有表现力和明显:

```go
return &Set{cap: sz, members: make(Set.members)}
```

但也许这只是我用 C 语言编程的时间太长了。

我不知道改变新的有相似的行为是有用的还是有价值的;我怀疑它的使用是不寻常的。在任何情况下，我都清楚这将减少重构软件以及编写新的软件的开销。

这个问题并不是那么糟糕，但修复它肯定会让我觉得更好。在很多情况下，Go已经比 C 和 C++ (我至今无法忍受) 更有表现力了。我认为，如果在语言中添加了对分配类型的推断的支持，那么 Go 语言对 C 系统程序员来说就会更有吸引力，因为除了 GC ，他们还有其他坚持的理由。(就我个人而言，我还希望看到一个关于支持系统级并发的更好的事情，但最好是单独发布。)

我编辑了这篇文章，以修复 C 示例中的一个错误。当我最初编写这个示例时，`struct set` 没有使用灵活的数组成员。Anmol Sethi 写信询问这个特性，并指出我错误地分配和再次分配给了FAM。我忘记了要删除那些代码。

嗷.

----------------

注释:

① : Kate Flavel 提醒我，对于 VLA 类型来说，这不是必须的，因为它是如此的无用，我总是忘记这点。这种类型有它自己的表达式求值。

----------------

via: https://9vx.org/post/a-malloc-idiom-in-go/

作者：[Devon](https://9vx.org/about)
译者：[SergeyChang](https://github.com/SergeyChang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
