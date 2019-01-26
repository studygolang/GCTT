首发于：https://studygolang.com/articles/17939

# 探索 Go 中字节解析 API

很多年前，我开始研究 Linux 的 Netlink 进程间通信接口。Netlink 被用于从 Linux 内核检索信息，并且为了跨越内核边界，信息通常被打包到 Netlink 的属性中。经过一些实验，我为 Go 创建了自己的 [netlink 包](https://github.com/mdlayher/netlink)。

随着时间的推移，包中的 API 已经有很大的改变了。特别是 Netlink 属性总是处理起来相当复杂。今天，我们将探索一些我为处理 Netlink 属性所创建的字节解析 API。这里描述的技术应该也能广泛应用于许多其他的 Go 库和应用程序中！

## Netlink 属性简介

Netlink 属性被以类型 / 长度 / 值或 TLV 格式打包，与许多二进制网络协议情况一样。这种格式具有很好的扩展性，因为许多属性可以在单个字节切片中被打包成自发自收的形式。

属性中的值可以包含：

- 无符号 8/16/32/64 位整形
- 以 null 结尾的 C 字符串
- 任意 C 结构字节
- 嵌套 Netlink 属性
- Netlink 属性数组

为了我们的目标，我们可以在 Go 中像下面这样定义一个 Netlink 属性：

```go
type Attribute struct {
	// The type of this Attribute, typically matched to a constant.
	Type uint16

	// Length omitted; Data will be a byte slice of the appropriate length.

	// An arbitrary payload which is specified by Type.
	Data []byte
}
```

今天，我们将略过低级的字节解析逻辑而有利于讨论各种高级 API，但是你能从 [我关于 Netlink 系列博客](https://medium.com/@mdlayher/linux-netlink-and-go-part-1-netlink-4781aaeeaca8) 中学习到更多的关于处理 Netlink 属性的知识。

## 字节解析 API 第一版

单个字节切片可以包含许多 Netlink 属性。让我们来定义一个初始的解析函数，它接收一个字节切片的输入并且返回一个属性切片。

```go
// UnmarshalAttributes unpacks a slice of Attributes from a single byte slice.
func UnmarshalAttributes(b []byte) ([]Attribute, error) {
	// ...
}
```

举个例子，假设我们想从属性切片中解包一个 `uint16` 和 `string` 值。你可以放心地忽略 `parseUint16` 和 `parseString`；它们将处理一些 Netlink 属性数据中棘手的部分。

为了解包属性数据，我们可以在 `Type` 属性上使用循环和匹配：

```go
attrs, err := netlink.UnmarshalAttributes(b)
if err != nil {
	return err
}

var (
	num uint16
	str string
)

for _, a := range attrs {
	switch a.Type {
	case 1:
		num = parseUint16(a.Data[0:2])
	case 2:
		str = parseString(a.Data)
	}
}

fmt.Printf("num: %d, str: %q", num, str)
// num: 1, str: "hello world"
```

这样可以正常工作，但是有一个问题：如果我们 `uint16` 值的字节切片比 2 字节多或者少，会出现什么情况呢？

```go
// A panic waiting to happen!
num = parseUint16(a.Data[0:2])
```

如果它少于 2 字节，此代码将出现 panic，并且让你的应用程序挂掉。如果它超过 2 字节，我们就默默地忽略任何额外的数据（这个值实际上不是 `uint16` ！）。

## 添加验证和错误处理

我们稍微修改一下我们的解析函数。每一个都应该做一些内部验证，如果字节切片不满足我们的限制，我们可以返回一个 error。

```go
attrs, err := netlink.UnmarshalAttributes(b)
if err != nil {
	return err
}

var (
	num uint16
	str string

	// Used to check for errors without shadowing num and str later.
	err error
)

for _, a := range attrs {
	// This works, but it's a bit verbose.
	// Be cautious of variable shadowing as well!
	switch a.Type {
	case 1:
		num, err = parseUint16(a.Data)
	case 2:
		str, err = parseString(a.Data)
	}
	if err != nil {
		return err
	}
}

fmt.Printf("num: %d, str: %q", num, str)
// num: 1, str: "hello world"
```

这样也是有效的，但你必须对你的错误检查策略保持谨慎，并且确保你不会意外地使用 `:=` 赋值运算符屏蔽掉你尝试解包的其中一个变量。

我们可以进一步改进这种模式吗？

## 一个类似迭代器的解析 API

上述的策略正常运行了许多年，但在编写了一些 Netlink 交互包之后，我决定开始改进 API。

新的 API 使用类似迭代器的模式，其灵感来自于标准库中的 `bufio.Scanner` API。Go 的博客 [Errors are values](https://blog.golang.org/errors-are-values) 这篇文章同样为解释这个策略做了出色的工作。

`netlink.AttributeDecoder` 类型就是一个类似迭代器的解析 API。在使用了 `netlink.NewAttributeDecoder` 构造器之后，许多方法被暴露出来，其能够与内部属性切片进行交互：

- `Next`：将内部指针指向下一个属性
- `Type`：返回当前属性的类型值
- `Err`：返回在迭代期间遇到的第一个错误

在尝试这个新的 API 时，让我们重温前面的例子：

```go
ad, err := netlink.NewAttributeDecoder(b)
if err != nil {
	return err
}

var (
	num uint16
	str string
)

// Continue advancing the internal pointer until done or error.
for ad.Next() {
	// Check the current attribute's type and extract it as appropriate.
	switch ad.Type() {
	case 1:
		// If data isn't a uint16, an error will be captured internally.
		num = ad.Uint16()
	case 2:
		str = ad.String()
	}
}

// Check for the first error encountered during iteration.
if err := ad.Err(); err != nil {
	return err
}

fmt.Printf("num: %d, str: %q", num, str)
// num: 1, str: "hello world"
```

有很多种方法可以用于迭代器期间提取的数据，例如 `Uint8/16/32/64`、`Bytes`、`string` 和所有的最有用的方法，包括：`Do`。

`Do` 是一种特殊用途的方法，允许解码器处理任意数据，如 C 结构、嵌套的 Netlink 属性、Netlink 数组。它能接受一个闭包，并将解码器所指向的当前数据传递给闭包。

为了处理嵌套 Netlink 属性，创建另外的包含一个 `Do` 闭包的 `AttributrEncoder`：

```go
ad.Do(func(b []byte) error) {
	nad, err := netlink.NewAttributeDecoder(b)
	if err != nil {
		return err
	}

	if err := handleNested(nad); err != nil {
		return err
	}

	// Make sure to propagate internal errors to the top-level decoder!
	return nad.Err()
})
```

为了保持小的闭包体，可以定义辅助函数来解析 Netlink 属性中的任意类型：

```go
// parseFoo returns a function compatible with Do.
func parseFoo(f *Foo) func(b []byte) error {
    return func(b []byte) error {
		// Some parsing logic...
		foo, err := unpackFoo(b)
		if err != nil {
			return err
		}

		// Store foo in f by dereferencing the pointer.
		*f = foo
		return nil
	}
}
```

现在，这个辅助函数可以直接用于 `Do`：

```go
var f Foo
ad.Do(parseFoo(&f))
```

此 API 为它的调用者提供了极大的灵活性。所有的错误传播都在内部处理，并通过从顶级解码器调用 `Err` 方法将错误冒泡到调用者。

## 结论

虽然花了一些时间和实验，但是我对 `netlink.AttributeDecoder` 中类似迭代器字节解析 API 感到非常满意。它非常适合于我的需求，感谢 Terin Stock，我们还添加了一个 [对称编码器 API](https://github.com/mdlayher/netlink/pull/95)，其灵感来自于解码器 API 的成功！

如果你正在开发一个你并不满意的包 API，标准库是寻找灵感的好地方！我也强烈建议与各种 [Go 帮助社区](https://golang.org/help/#help) 取得联系，因为有很多人非常愿意提供出色的建议和批评！

如果你有任何问题，请随时和我联系！我在 [Gophers Slack](https://gophers.slack.com/)、[Github](https://github.com/mdlayher) 和 [Twitter](https://twitter.com/mdlayher) 的称号是 mdlayher。

## 链接

- [netlink 包](https://github.com/mdlayher/netlink)
- [Linux、Netlink 和 Go 博客系列](https://medium.com/@mdlayher/linux-netlink-and-go-part-1-netlink-4781aaeeaca8)
- [Go 博客：Errors are values](https://blog.golang.org/errors-are-values)
- [`bufio.Scanner`](https://golang.org/pkg/bufio/#Scanner)
- [`netlink.AttributeDecoder`](https://godoc.org/github.com/mdlayher/netlink#AttributeDecoder)

---

via: https://blog.gopheracademy.com/advent-2018/exploring-byte-parsing-apis-in-go/

作者：[Matt Layher](https://blog.gopheracademy.com/advent-2018/exploring-byte-parsing-apis-in-go/)
译者：[PotoYang](https://github.com/PotoYang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
