# Go: 切片陷阱

## 前言

我最喜欢Go的一个特性就是, 毫无惊喜。某些程度上可以说有点无聊的感觉。这是编程语言的一个优秀的品质。这样的话, 在编码的时候就可以专注于手头上的问题, 而不是[语言做了你不希望它做的事情](https://twitter.com/chordbug/status/1092824183124488192?s=19)。

这篇文章有关Go的一个对新人来说最"惊喜"的特性: slice。

## 基本用法

如果你了解如何使用Go slice, 请跳到下一节。

你可以这样声明一个 slice:

```go
var a []int
```

带字面量的slice:

```go
a := []int{1, 2}

```

slice是可变长度的集合。不像数组, slice可以按需进行增长和切分。

数组:

```go
// 全0数组, 大小为4
var a [4]int
// 有字面量的全0数组, 大小为3
b := [...]int{2: 0}
// 有字面量的全0数组, 大小为2
c := [...]int{0, 0}
// 以下都不合法, [4]int, [3]int 以及 [2]int 是不一样的类型
a = b
c = b
```

Slices:

```go
// 全0 slice, 大小为4
a := make([]int, 4)
// 有字面量的全0 slice, 大小为3
b := []int{2: 0}
// 有字面量的全0 slice, 大小为2
c := []int{0, 0}
// 以下是允许的:[]int 和 []int 是相同的类型
a = b
c = a
```

而且, slice还可以进行子切分:

```go
a := []int{0, 1, 2, 3, 4}
b := a[1:3] /* [1, 2]          */
c := a[3:]  /* [3, 4]          */
d := a[:2]  /* [0, 1]          */
e := a[:]   /* [0, 1, 2, 3, 4] */
```

以及增长:

```go
a := []int{1, 2}
b := append(a, a...) /* [1, 2, 1, 2] */
a = append(a, 3, 4)  /* [1, 2, 3, 4] */
```

这通常让slice成为所有应用场景首选的数据结构(译者注:应该是对于所有适用数组和slice的场景而言, slice胜于数组)

## 那么, 有什么问题呢?

slice不是其他东西, 而是一个携带三份信息的 struct

```go
type slice struct {
	// 在data中使用到的空间大小
	len  int
	// data的大小
	cap  int
	// 底层数组 data
	data *[...]Type
}
```

当从slice中拿到一个slice, `cap`, `len`和`data`都可能会变化, 但**底层数组既不会进行重新分配, 也不会进行复制。**

这个特性导致了一些怪异的行为。

## 迷之更新:第一部分

```go

a := []int{1, 2}
b := a[:1]     /* [1]     */
b[0] = 42      /* [42]    */
fmt.Println(a) /* [42, 2] */
```

这类技巧基本在gophers的意料之中, 通常是因为语言的某些核心接口依赖于slice的底层数据通过引用传递的事实。 例如，io.Reader具有与io.Writer相同的类型签名，对于新人来说可能相当令人惊讶：

```go
type Reader interface {
	// Read 把数据写到p中
	Read(p []byte) (n int, err error)
}
type Writer interface {
	// Write 从p中读取数据
	Write(p []byte) (n int, err error)
}
```

## 迷之更新:第2部分

这部分看起来更具迷惑性

```go
a := []int{1, 2, 3, 4}
b := a[:2] /* [1, 2] */
c := a[2:] /* [3, 4] */
b = append(b, 5)
fmt.Println(a) /* [1 2 5 4] */
fmt.Println(b) /* [1 2 5]   */
fmt.Println(c) /* [5 4]     */
```

当数据被追加到`b`, 底层数组有足够的容量来保存多两个元素, 所以`append`不会重新分配, 这意味着, 数据追加到`b`之后会改变`c`。

## 迷之更新:第3部分

```go
a := []int{0}     /* [0]          */
a = append(a, 0)  /* [0, 0]       */
b := a[:]         /* [0, 0]       */
a = append(a, 2)  /* [0, 0, 2]    */
b = append(b, 1)  /* [0, 0, 1]    */
fmt.Println(a[2]) /* 2 <- 对的   */

// 一样的代码, 只是以一个稍大的slice开始
c := []int{0, 0}  /* [0, 0]       */
c = append(c, 0)  /* [0, 0, 0]    */
d := c[:]         /* [0, 0, 0]    */
c = append(c, 2)  /* [0, 0, 0, 2] */
d = append(d, 1)  /* [0, 0, 0, 1] */
fmt.Println(c[3]) /* 1 <- ??      */
```

这个奇怪的行为的原因是, 当slice变得比某个确切的阈值要大时, go停止线性增长并开始分配一个大小翻倍的slice。**这取决于slice类型的大小**。

分析更多的细节:

+ 第一个在`a`上的`append`复制前一个0到一个`cap==2`的slice, 然后在`a[1]`上填一个`0`

+ 从`a`拿到了一个slice, `len(b) == cap(b) == 2`

+ 第二个在`a`上的`append`复制前面的0到一个`cap==4`的slice, 然后在`a[2]`上填上`2`

+ 在这里, `b`依然还是`cap == 2`, 所以在`b`上`append`, 分配了一个新的底层数组

同样的过程, 以初始`cap`为2的slice开始, 产生了不一样的结果, 因为当我们拿到slice `c`时, 它已经增长到`cap == 4`

> 碎碎念:由于这种行为取决于底层类型的大小，因此`[]struct{}{}`将始终通过追加的元素的确切数量增长。

## 我该怎么解决这个问题呢?

如果你传递一个从不追加的slice, 那么这是安全的。只需要紧紧记住, 每一个(传递的slice)都共享相同内存区域的"视图"。如果你调用的函数在返回后不保留对slice的引用，也同样适用。

相反地, 如果你打算传递可能要追加数据的slice, 然后你也打算对原slice进行扩容, 你可能会希望考虑限制你所分享的数据的容量。

```go
a := append([]int{}, 0, 1, 2, 3)
// 如果`potentialSliceGrower`保持着对`a`的引用, 下方这种调用可能是危险的
potentialSliceGrower(a)
// 这个是安全的, 取一个确定大小的slice(进行传递)
// 追加则会引起复制
potentialSliceGrower(a[:4:4])
```

这种不常使用的`3-index`语法, 从`a`中拿到了一个从下标`0`开始, 到下标`4`结束, `cap=4`的slice。

请在**真正有需要的时候**再使用它, 但在需要的时候别忘记这个方法。

## 想要了解更多?

这儿有一个关于slice内部机制的go的官方[博客](https://blog.golang.org/go-slices-usage-and-internals), 在文末还有[其他陷阱](https://blog.golang.org/go-slices-usage-and-internals#TOC_6)。

---

via: https://blogtitle.github.io/go-slices-gotchas/

作者：[Rob](https://blogtitle.github.io/authors/rob/)
译者：[LSivan](https://github.com/LSivan)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出