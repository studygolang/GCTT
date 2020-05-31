首发于：https://studygolang.com/articles/28992

# Go 1.14 中接口的菱形组合

按照[部分重叠的接口提议](https://github.com/golang/proposal/blob/master/design/6977-overlapping-interfaces.md)，Go 1.14 现在允许嵌入有部分方法重叠的接口。本文是一篇解释这次修改的简要说明。

我们先来看 io 包中的三个关键接口：io.Reader、io.Writer 和 io.Closer：

```go
package io

type Reader interface {
	Read([]byte) (int, error)
}

type Writer interface {
	Write([]byte) (int, error)
}

type Closer interface {
	Close() error
}
```

在结构体中嵌入类型时，如果在结构体中声明了被嵌入的类型，那么该类型的字段和方法允许被访问[^1]，对于接口来说这个处理也成立。因此下面两种方式：显式声明

```go
type ReadCloser interface {
	Read([]byte) (int, error)
	Close() error
}
```

和使用嵌入来组成接口

```go
type ReadCloser interface {
	Reader
	Closer
}
```

没有区别。

你甚至可以混合使用：

```go
type WriteCloser interface {
	Write([]byte) (int, error)
	Closer
}
```

然而，在 Go 1.14 之前，如果你用这种方式来声明接口，你可能会得到类似这样的结果：

```go
type ReadWriteCloser interface {
	ReadCloser
	WriterCloser
}
```

编译错误：

```bash
% Go build interfaces.go
command-line-arguments
./interfaces.go:27:2: duplicate method Close
```

幸运的是，在 Go 1.14 中这不再是一个限制了，因此这个改动解决了在菱形嵌入时出现的问题。

然而，在我向本地的用户组解释这个特性时也陷入了麻烦 — 只有 Go 编译器使用 1.14（或更高版本）语言规范时才支持这个特性。

我理解的编译过程中 Go 语言规范所使用的版本的规则似乎是这样的：

1. 如果你的源码是在 GOPATH 下（或者你用 GO111MODULE=off *关闭*了 module），那么 Go 语言规范会使用你编译器的版本来编译。换句话说，如果安装了 Go 1.13，那么你的 Go 版本就是 1.13。如果你安装了 Go 1.14，那么你的版本就是 1.14。这里符合认知。
2. 如果你的源码保存在 GOPATH 外（或你用 GO111MODULE=on 强制开启了 module），那么 Go tool 会从 go.mod 文件中获取 Go 版本。
3. 如果 go.mod 中没有列出 Go 版本，那么语言规范会使用安装的 Go 的版本。这跟第 1 点是一致的。
4. 如果你用的是 Go module 模式，不管是源码在 GOPATH 外还是设置了 GO111MODULE=on，但是在当前目录或所有父目录中都没有 go.mod 文件，那么 Go 语言规范会默认用 Go 1.13 版本来编译你的代码。

我曾经遇到过第 4 点的情况。

[^1]: 也就是说，嵌入提升了类型的字段和方法。

---

via: https://dave.cheney.net/2020/05/24/diamond-interface-composition-in-go-1-14

作者：[Dave Cheney](https://dave.cheney.net/)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
