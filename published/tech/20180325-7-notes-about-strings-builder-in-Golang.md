已发布：https://studygolang.com/articles/12796

# Golang 中 strings.builder 的 7 个要点

自从 Go 1.10 发布的一个月以来，我多少使用了一下 `strings.Builder`，略有心得。你也许知道它，特别是你了解 `bytes.Buffer` 的话。所以我在此分享一下我的心得，并希望能对你有所帮助。

## 1. 4 类写入（write）方法

与 `bytes.Buffer` 类似，`strings.Builder` 也支持 4 类方法将数据写入 builder 中。

```go
func (b *Builder) Write(p []byte) (int, error)
func (b *Builder) WriteByte(c byte) error
func (b *Builder) WriteRune(r rune) (int, error)
func (b *Builder) WriteString(s string) (int, error)
```

有了它们，用户可以根据输入数据的不同类型（byte 数组，byte， rune 或者 string），选择对应的写入方法。

![four-forms-of-writing-methods](https://raw.githubusercontent.com/studygolang/gctt-images/master/strings-builder/1_IGv0x1gMwkszbv7IWnpfaQ.png)

## 2. 字符串的存储原理

根据用法说明，我们通过调用 `string.Builder` 的写入方法来写入内容，然后通过调用 `String()` 方法来获取拼接的字符串。那么 `string.Builder` 是如何组织这些内容的呢？

**通过 slice**

`string.Builder` 通过使用一个内部的 slice 来存储数据片段。当开发者调用写入方法的时候，数据实际上是被追加（append）到了其内部的 slice 上。

![slice-store-data](https://raw.githubusercontent.com/studygolang/gctt-images/master/strings-builder/1_luRaetJ4m36JH43xh0rHcA.png)

## 3. 高效地使用 strings.Builder

根据上面第 2 点可以知道，strings.Builder 是通过其内部的 slice 来储存内容的。当你调用写入方法的时候，新的字节数据就被追加到 slice 上。如果达到了 slice 的容量（capacity）限制，一个新的 slice 就会被分配，然后老的 slice 上的内容会被拷贝到新的 slice 上。当 slice 长度很大时，这个操作就会很消耗资源甚至引起 [内存问题](https://blog.siliconstraits.com/out-of-memory-with-append-in-golang-956e7eb2c70e)。我们需要避免这一情况。

关于 slice，Go 语言提供了 `make([]TypeOfSlice, length, capacity)` 方法在初始化的时候预定义它的容量。这就避免了因达到最大容量而引起扩容。

`strings.Builder` 同样也提供了 `Grow()` 来支持预定义容量。当我们可以预定义我们需要使用的容量时，`strings.Builder` 就能避免扩容而创建新的 slice 了。

```go
func (b *Builder) Grow(n int)
```

当调用 `Grow()` 时，我们必须定义要扩容的字节数（`n`）。 `Grow()` 方法保证了其内部的 slice 一定能够写入 `n` 个字节。只有当 slice 空余空间不足以写入 `n` 个字节时，扩容才有可能发生。举个例子：

* builder 内部 slice 容量为 10。
* builder 内部 slice 长度为 5。
* 当我们调用 `Grow(3)` => 扩容操作并不会发生。因为当前的空余空间为 5，足以提供 3 个字节的写入。
* 当我们调用 `Grow(7)` => 扩容操作发生。因为当前的空余空间为 5，已不足以提供 7 个字节的写入。

关于上面的情形，如果这时我们调用 `Grow(7)`，则扩容之后的实际容量是多少？

```
17 还是 12?
```

实际上，是 `27`。`strings.Builder` 的 `Grow()` 方法是通过 `current_capacity * 2 + n` （`n` 就是你想要扩充的容量）的方式来对内部的 slice 进行扩容的。所以说最后的容量是 `10*2+7` = `27`。

当你预定义 `strings.Builder` 容量的时候还要注意一点。调用 `WriteRune()` 和 `WriteString()` 时，`rune` 和 `string` 的字符可能不止 1 个字节。因为，你懂的，[UTF-8](https://golang.org/pkg/unicode/utf8/#pkg-constants) 的原因。

## 4. String()

和 `bytes.Buffer` 一样，`strings.Builder` 也支持使用 `String()` 来获取最终的字符串结果。为了节省内存分配，它通过使用指针技术将内部的 buffer bytes 转换为字符串。所以 `String()` 方法在转换的时候节省了时间和空间。

```go
*(*string)(unsafe.Pointer(&bytes))
```

## 5. 不要拷贝

![do-not-copy](https://raw.githubusercontent.com/studygolang/gctt-images/master/strings-builder/1_a4IwPDq3tEJJ_FRZfhreyQ.png)

`strings.Builder` 不推荐被拷贝。当你试图拷贝 `strings.Builder` 并写入的时候，你的程序就会崩溃。

```go
var b1 strings.Builder
b1.WriteString("ABC")
b2 := b1
b2.WriteString("DEF")
// illegal use of non-zero Builder copied by value
```

你已经知道，`strings.Builder` 内部通过 slice 来保存和管理内容。slice 内部则是通过一个指针指向实际保存内容的数组。

![slice-internally](https://raw.githubusercontent.com/studygolang/gctt-images/master/strings-builder/1_KD02pGfasisf8I_BWE_JKQ.png)

当我们拷贝了 builder 以后，同样也拷贝了其 slice 的指针。但是它仍然指向同一个旧的数组。当你对源 builder 或者拷贝后的 builder 写入的时候，问题就产生了。另一个 builder 指向的数组内容也被改变了。这就是为什么 `strings.Builder` 不允许拷贝的原因。

![copy-and-write](https://raw.githubusercontent.com/studygolang/gctt-images/master/strings-builder/1_Ppak_h63S_TvYzJa2sFCpA.png)

对于一个未写入任何东西的空内容 builder 则是个例外。我们可以拷贝空内容的 builder 而不报错。

```go
var b1 strings.Builder
b2 := b1
b2.WriteString("DEF")
b1.WriteString("ABC")

// b1 = ABC, b2 = DEF
```

`strings.Builder` 会在以下方法中检测拷贝操作：

```go
Grow(n int)
Write(p []byte)
WriteRune(r rune)
WriteString(s string)
```

所以，拷贝并使用下列这些方法是允许的：

```go
// Reset()
// Len()
// String()

var b1 strings.Builder
b1.WriteString("ABC")
b2 := b1
fmt.Println(b2.Len())    // 3
fmt.Println(b2.String()) // ABC
b2.Reset()
b2.WriteString("DEF")
fmt.Println(b2.String()) // DEF
```

## 6. 并行支持

和 `bytes.Buffer` 一样，`strings.Builder` 也不支持并行的读或者写。所以我们们要稍加注意。

可以试一下，通过同时给 `strings.Builder` 添加 `1000` 个字符：

```go
package main

import (
	"fmt"
	"strings"
	"sync"
)

func main() {
	var b strings.Builder
	n := 0
	var wait sync.WaitGroup
	for n < 1000 {
		wait.Add(1)
		go func() {
			b.WriteString("1")
			n++
			wait.Done()
		}()
	}
	wait.Wait()
	fmt.Println(len(b.String()))
}
```

通过运行，你会得到不同长度的结果。但它们都不到 `1000`。

## 7. io.Writer 接口

`strings.Builder` 通过 `Write(p []byte) (n int, err error)` 方法实现了 `io.Writer` 接口。所以，我们多了很多使用它的情形：

* `io.Copy(dst Writer, src Reader) (written int64, err error)`
* `bufio.NewWriter(w io.Writer) *Writer`
* `fmt.Fprint(w io.Writer, a …interface{}) (n int, err error)`
* `func (r *http.Request) Write(w io.Writer) error`
* 其他使用 io.Writer 的库

![io-writer](https://raw.githubusercontent.com/studygolang/gctt-images/master/strings-builder/1_MhBcQBYT4ocfA7ftVT2iGw.png)

---

via: https://medium.com/@thuc/8-notes-about-strings-builder-in-golang-65260daae6e9

作者：[Thuc Le](https://medium.com/@thuc)
译者：[alfred-zhong](https://github.com/alfred-zhong)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
