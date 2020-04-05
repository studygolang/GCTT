# unsafe 真就 unsafe 吗 - part1

## unsafe 包详解

在乌克兰的利沃夫举行的 [Lviv Golang community event](https://www.facebook.com/events/470065893928934/482981832637340/?notif_t=admin_plan_mall_activity&notif_id=1580732874088578) 中，我发表了一个关于`unsafe` 包的演讲，这个演讲中我尝试回答了标题中提到的问题：`unsafe` 包究竟有多 unsafe。

从 `unsafe` 包的名字就能感受到 go 研发团队的警告：使用这个包的代价将是巨大的。我觉得这个包名起的非常巧妙，它完美地符合了 《Effective Go》中对包名的所有建议。在使用 unsafe 包的时候，我们应当严格遵循 Go 研发团队的文档和建议。这个包的官方概述就只有简单的一段话：

> unsafe 包里面包含了一些能让你践踏 Go 语言的类型安全特性的操作。[golang.org](https://golang.org/pkg/unsafe/#pkg-overview)

附带一段简单的警告：

> 引用 unsafe 包可能会导致你代码的不具备可移植性，并且不再受到 Go 1 兼容性规约的保护。[golang.org](https://golang.org/pkg/unsafe/#pkg-overview)

函数功能的描述看起来非常的抽象，我们来瞅一眼这些 "unsafe" 的操作：

```go
func Alignof(x ArbitraryType) uintptr
func Offsetof(x ArbitraryType) uintptr
func Sizeof(x ArbitraryType) uintptr
type ArbitraryType
type Pointer
```

其中有一个 `ArbitraryType` 类型：

> 只是为了文档记录的目的而存在，实际上它没有参与到 unsafe 包的实现。这个类型代表了任意的 Go 语言表达式。

所以实际上 unsafe 包就只包含三个函数和一个类型，既然就这么点东西，那我们试着把这个包全部过一遍。现在我们手头上已经有了 Go 研发团队给我的文档和源码，下一步要怎么做？这时候不妨重温一句名人名言：

> 多说无益，放码过来 —— [Linus Torvalds](https://lkml.org/lkml/2000/8/25/132)

好，既然如此我们就直接看源码吧……

![Gif](https://tva1.sinaimg.cn/large/00831rSTgy1gcy7aqxlrbg30bu05xnpd.gif)

神奇的事情发生了——这个 unsafe 包压根就[没有源码](https://golang.org/src/unsafe/unsafe.go)呀。它有函数的签名和类型定义，但是没有实现的代码：无论是 go 还是汇编的代码都没有。之所以会出现这个情况，是因为 unsafe 包的功能需要在层次更低的编译器层面实现，所以这个包其实是内置在编译器里面实现的，这个 .go 文件只是为了达到文档记录的目的。所以我在上文反复强调要严格遵循 Go 研发团队的文档和建议，因为你也只能看到这些文档。废话不多说，先来看看 `Sizeof` 函数吧。

### `func Sizeof(x ArbitraryType) uintptr`

函数接受某个变量，然后返回 `uintptr` 类型的结果。这个函数的名字可以看出，这个函数返回某个变量的大小。为了理解方便，请允许我用几个图示来可视化一下这些概念。众所周知，我们的 Go 程序需要内存来完成各种功能，其中就包含使用内存来保存变量。下面我将用这些标签来表示内存：

🎁 - 1 个字节的内存

📦 - 1 个字节的被占用的内存

🥡 - 1 个字节的被占用但实际没有作用的内存（后面会详细解释这个）

⬆️ - 指向内存地址的指针

下面我使用这些标签来展示这个结构的内存布局：

```go
type X struct {
  n1 int16
  n2 int16
}
```

它在内存中的布局是这样的

![Sizeof memory usage](https://tva1.sinaimg.cn/large/00831rSTgy1gcy7aizlppj30q60ba0u8.jpg)

`X` 结构体有两个字段，其中每一个都占 2 个字节，所以整个结构体占用 size(`n1`) + size(`n2`) + size(`X`) = 2 + 2 + 0 = 4。显然，下面语句是成立的：

```go
unsafe.Sizeof(X) == 4 // true
```

### `func Offsetof(x ArbitraryType) uintptr`

这个函数就有点难度了，函数签名和上面的函数是同样的，但是它返回的是 offset（偏移值）。我再次使用标签来解释这个机制——还是用刚才的 `X` 结构体，还是同样的两个字段：

```go
type X struct {
  n1 int16
  n2 int16
}
```

现在我们已经知道他在内存里面是怎样布局的了，这一次我们来看看每个字段各占多少个字节，内存分配的情况如下图：

![Offsetof memory usage](https://tva1.sinaimg.cn/large/00831rSTgy1gcy7ayn8bnj30r20bggn3.jpg)

不难猜到，内存的布局是这样的：第一个字段`X.n1` 占了前 2 个字节，而第二个字段 `X.n2` 占了接下来的 2 个字节。所以下面两个语句都是成立的：

```go
unsafe.Offsetof(X.n1) == 0 // true
unsafe.Offsetof(X.n2) == 2 // true
```

### `func Alignof(x ArbitraryType) uintptr`

这个函数是最好玩的一个，因为要透彻了解这个函数，你需要了解 [alignment（数据结构对齐）](https://zh.wikipedia.org/wiki/数据结构对齐) 是怎么回事。简单来说，它让数据在内存中以某种的布局来存放，使该数据的读取能够更加的快速。这个接收一个变量作为参数，并返回这个变量的对齐字节。为了更加直观，我们需要修改一下上面的例子：

```go
type X struct {
  n1 int8
  n2 int16
}
```

可以看到现在 `n1` 的类型变成了 `int8`，这会有什么变化吗，我们先看看 `Sizeof`, 因为 `n1` 只占 1 个字节了，所以合理地推测，`X` 结构体的大小会变成 3，因为：size(`X`) = size(`n1`) + size(`n2`) = 1 + 2 = 3。**但是**现实真的如此吗 ？

……

不是的，因为 alignment 的缘故，`X` 结构体在内存的结构如下：

![Alignof memory usage](https://tva1.sinaimg.cn/large/00831rSTgy1gcy7b90qipj30qq0beq4o.jpg)

由于 alignment 机制的要求，`n2` 的**内存起始地址应该是自身大小的整数倍**，也就是说它的起始地址只能是 0、2、4、6、8 等偶数，所以 `n2` 的起始地址没有紧接着 `n1` 后面，而是空出了 1 个字节。最后导致结构体 `X` 的大小是 4 而不是 3。机智的读者可能会想到：`n1` 和 `n2` 换个位置会怎样呢？这样一来，`n2` 的起始地址是 0，而`n1` 的其实地址是 2，这么一来结构体 `X` 的大小就变成 3 了吧？答案是……不对的。原因还是因为 alignment，因为 alignment 除了要求字段的其实地址应该是自身大小的整数倍，还要求**整个结构体的大小，是结构体中最大的字段的大小的整数倍**，这使得结构体可以由多个内存块组成，其中每个内存块的大小都等于最大的字段的大小。我们可以利用这个知识来减少结构体的内存占用。考察以下代码：

```go
type First struct {
	a int8
	b int64
	c int8
}

type Second struct {
	a int8
	c int8
	b int64
}

fmt.Println("Big brain time: ", unsafe.Sizeof(First{}) == unsafe.Sizeof(Second{}))
```

上面两个结构体大小不同，是因为 `First` 结构体由三个大小为 8 字节的内存块组成：`Sizeof(First.a) +  7 个空闲的字节 + Sizeof(First.b) + Sizeof(First.c) + 7 个空闲的字节 = 24 字节`。而 `Second` 结构体只包含  2 个 大小为 8 字节的内存块：`Sizeof(Second.a) + Sizeof(Second.b) + 6 个空闲的字节 + Sizeof(Second.b) = 16 字节`。下次你定义结构体的时候可以用上这个小知识🙂。

下面的代码片段总结了上述三个函数的用法：

```go
var x struct {
	a int64
	b bool
	c string
}

fmt.Println("Size of x: ", unsafe.Sizeof(x))
fmt.Println("Size of x.c: ", unsafe.Sizeof(x.c))

fmt.Println("Alignment of x.a: ", unsafe.Alignof(x.a))
fmt.Println("Alignment of x.b: ", unsafe.Alignof(x.b))
fmt.Println("Alignment of x.c: ", unsafe.Alignof(x.c))

fmt.Println("\nOffset of x.a: ", unsafe.Offsetof(x.a))
fmt.Println("Offset of x.b: ", unsafe.Offsetof(x.b))
fmt.Println("Offset of x.c: ", unsafe.Offsetof(x.c))
```

上述的三个方法都是在[编译期](https://en.wikipedia.org/wiki/Compile_time)执行的，这意味着只要它们在编译器没有报错，在运行时不会有问题发生。但是我们的下一位嘉宾 `unsafe.Pointer` 可就没那么好惹了，它有可能会发生[运行时的错误](https://en.wikipedia.org/wiki/Runtime_(program_lifecycle_phase))。我将会在本文的第二部分详细介绍 `unsafe.Pointer` 以及使用它的过程中容易出现的问题。

---

via: https://www.dnahurnyi.com/is-unsafe-...unsafe-pt.-1/

作者：[Denys Nahurnyi](https://www.dnahurnyi.com/)
译者：[Alex-liutao](https://github.com/Aelx-liutao)
校对：[@unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
