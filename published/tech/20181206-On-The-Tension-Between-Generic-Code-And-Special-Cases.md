首发于：https://studygolang.com/articles/17399

# 关于通用代码和特殊情况之间的冲突

`io.Reader` 和 `io.Writer` 接口几乎出现在所有的 Go 程序中，并代表了处理数据流的基本构建块。Go 的一个重要特性是，对象如套接字、文件或内存缓冲区的抽象都是用这些接口表示的。当 Go 程序对外部世界说话的时候，它几乎是通过 `io.Reader`s 和 `io.Writer` s 来表达，无论它使用的是特殊的平台或通信媒介。这种普遍性是编码处理可组合和可重复使用的数据流代码的关键因素<sup><a href="#fn1" name="fnref1">1</a></sup>。

这篇文章研究了 `io.Copy` 的设计和实现，该函数用可能是最简单的方法连接一个 `Reader` 到一个 `Writer`：该函数从一个地方传输数据到另一个地方。

通常情况下 <sup><a href="#fn2" name="fnref2">2</a></sup>，`io.Copy` 分配一个缓冲区，然后从源读取器读取到缓冲区和从缓冲区写到目标写程序交替进行。这在许多情况下工作得良好，并且从语义角度来看肯定是正确的。

这么说来，如果对于一些特殊的 reader 和 writer 的选择，我们可以做得更好吗？我们怎么教授 `Copy` 呢？

使用高级抽象的代码，如 `Reader` 和 `Writer` 必须经常回答这些问题，并且必须处理这些冲突。通常来说，不同的平台，编程的语言甚至库都用不同的方法处理这个问题。

让我们来特别研究一下 `io.Copy` 这种情况，以期许得到更普遍的智慧。

## 一种可能的尝试：教授特定类型的 Copy

想象一下一个 `Copy`，看起来像这样：

```go
package hypotheticalio

import "bytes"

func Copy(dst Writer, src Reader) (int64, error) {
	switch s := src.(type) {
	case *bytes.Buffer:
		n, err := dst.Write(s.Bytes())
		return int64(n), err
	default:
		// generic code path
	}
}
```

注意我们假设的 `io` 包现在如何导入 `bytes`，以便它可以在 switch 类型中使用 `Buffer` 类型。这里禁止从 `io` 包导入 `bytes`，因为 Go 不允许循环导入。也许我们还没有注意到这个问题，我们继续前进。

时光流逝，我们发现更值得考虑的特殊情况：

```go
package hypotheticalio

import (
	"bytes"
	"net"
	"os"
)

func Copy(dst Writer, src Reader) (int, error) {
	switch s := src.(type) {
	case *bytes.Buffer:
		n, err := dst.Write(s.Bytes())
		return int64(n), err
	case *net.TCPConn:
		return platformSpecificThings(dst, s)
	case *os.File:
		return differentPlatformSpecificCode(dst, s)
	default:
		// generic code path
	}
}
```

`Copy` 的代码被改变了很多，尽管代码的 *意思* 没有任何的改变。不仅如此，`Cpoy` 现在关注特定平台的位，它了解操作系统、网络等等。它过去很好而且通用，但是现在有很难维护、混乱的特殊情况。

似乎是有些事情出了问题。这个 `Copy` *确实* 适用于特殊情况和通用代码，但它付出了可怕的代价去这样做，并且它对世界的其他地方施加了可怕的限制。

## 也许是一个更好的尝试：使用接口将 Copy 与世界分离

与教授特定类型的 `Copy` 相反，`io` 包引入了两个新的接口：`ReaderFrom` 和 `WriterTo`。

`ReaderFrom` 可以被认为是一个消费来自一个 `Reader` 数据的对象。相比之下，`WriterTo` 可以被认为是一个向 `Writer` 推入自身数据的对象

从概念上讲，在两种情况下都会出现从一个对象到另一个对象的数据传输，但表达传输的方式会产生不同。`Copy` 不需要知道它正在使用的类型的任何具体内容。如果它们中的任何一个实现了 `ReaderFrom` 或 `WriterTo`，`Copy` 调用该方法，并且没有执行其他工作。`Copy` 现在看起来像这样：

```go
package io

func Copy(dst Writer, src Reader) (int64, error) {
	if wt, ok := src.(WriterTo); ok {
		return wt.WriteTo(dst)
	}
	if rt, ok := dst.(ReaderFrom); ok {
		return rt.ReadFrom(src)
	}
	// generic code path
}
```

发生了一些有趣的事情：与之前假设的情景相比，`Copy` 现在几乎没有理由改变。它再次完全通用。不只是那样，它可以委托给代码片段，这些代码片段和以前一样 *确实* 拥有更加具体的知识。

但是，没有什么是免费的，这种松散的耦合也有其成本。通过特定的类型不能够静态地知道 `Copy`，但是使用类型断言，一定能在运行时动态地被发现。

有趣的是，现在通用代码与和特殊情况之间的冲突不是通过凌乱的代码、高维护成本和过高的导入限制来显示自己，而是通过丢失编译时信息而显示出来。对于像 `io` 这样全世界都导入的包，这当然看上去像是一个值得做的交易。

调用者可以自己定制化 `io.Copy`，而无需改变函数本身。他们需要做的就是实现 `io.ReaderFrom` 或 `io.WriterTo`。标准库在很多地方都像这样做了。例如：

- `*bytes.Buffer` 有一个 [WriteTo](https://golang.org/pkg/bytes/#Buffer.WriteTo)，它将缓冲区注入一个 `io.Writer`，以及一个 [ReadFrom](https://golang.org/pkg/bytes/#Buffer.ReadFrom)，它从一个 `io.Reader` 填充到缓冲区
- `*net.TCPConn` 有一个 [ReadFrom](https://golang.org/pkg/net/#TCPConn.ReadFrom)，它可以在大多数平台上使用 `sendfile(2)`（或者一个相似的接口）
- `net/http` 对 `ResponseWriter` 的实现有一个 [ReadFrom](https://golang.org/src/net/http/server.go#L566)，它可以使用上述特殊情况下的 `sendfile(2)`

值得注意的是，这些都是优化，不应该以任何形式影响程序的语义。因此，对于 `io` 包的客户来说可能发生的最糟糕的事情就是特定的优化可能不会起作用。让我们来研究一下这种情况。考虑以下包装类型：

```go
type CountingWriter struct {
	W io.Writer
	N int64
}

func (cw *CountingWriter) Write(b []byte) (int, error) {
	n, err := cw.W.Write(b)
	cw.N += int64(n)
	return n, err
}
```

当被用做是 `io.Writer` 时，`CountingWriter` 隐藏了来自调用者的底层属性。因此，在运行时检查功能的代码，例如 `io.Copy`，在查看 `*CountingWriter` 时将会只看到 `io.Writer`。

然而，在这种情况下需要底层 `Witer` 的特定功能，调用者需要通过发现有趣的功能和使用更加具体的包装方法去适应自己的情况。在特定情况下这可能非常地困难 <sup><a href="#fn3" name="fnref1">3</a></sup>。

此外，请注意为何 `io.ReaderFrom` 和 `io.WriterTo` 不出现在 `io.Copy` 的 *签名* 中。相反，它们出现在 *文档* 中：一个弱得多的约定。

## 最后思考

无论如何，通用代码和特殊情况之间的根本冲突出现在任何处理抽象的代码中。为了适应两者，Go 接口的性质允许组件之间的一种特定松散耦合，但是这种方法并没有其微妙的成本。即便如此，最终结果仍然优雅且易于维护。

---

<a id="fn1">1</a>. Go 与平台的对比见 [red-blue](http://journal.stuffwithstuff.com/2015/02/01/what-color-is-your-function/) <sup><a href="#fnref1">return</a></sup>

<a id="fn2">2</a>. 在 [这里](https://github.com/golang/go/blob/112f28defcbd8f48de83f4502093ac97149b4da6/src/io/io.go#L401-L423) 看源码 <sup><a href="#fnref2">return</a></sup>

<a id="fn3">3</a>. 查看组合（译注：原文单词错误，应为 combinatorial）展现的 [这个库](https://github.com/felixge/httpsnoop) <sup><a href="#fnref3">return</a></sup>

---

via: https://blog.gopheracademy.com/advent-2018/generic-code-vs-special-cases/

作者：[Andrei Tudor Călin](https://blog.gopheracademy.com/advent-2018/generic-code-vs-special-cases/)
译者：[PotoYang](https://github.com/PotoYang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
