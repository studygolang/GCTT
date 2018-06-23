已发布：https://studygolang.com/articles/12122

# 两个随机数函数的故事

2017 年 12 月 5 日 作者：内尔·卡彭铁尔

我经常有一些困惑，`crypto/rand` 包和 `math/rand` 包是如何关联的，或者它们是如何按照预期的方式（一起）工作的？这是其他人已经思考过的问题，还是仅仅我个人的突发奇想呢？终于有一天，我决定攻克这个问题，这篇博客就是这次努力的结果。

## `math` 包

如果你曾经关注过 `math/rand` 包，你会同意它提供了相当易用的 API。我最喜欢的例子是 `func Intn(n int) int` 函数，它返回了一个你指定范围内的随机数。非常有用！

你也许会问，顶级函数和 `Rand` 类型的实例函数之间有什么异同。 如果你看了[源代码实现](https://golang.org/src/math/rand/rand.go)的话，你会发现，顶级函数只是一个易用性的封装，内部指向了一个包全局对象 `globalRand`。

尽管这个包有些使用上的小陷阱。最基础的用法是提供一个**伪随机**数作为种子。这就意味着，如果你使用相同的种子来生成两个 `Rand` 实例，对这两个实例进行相同次序和函数的调用，那么将会得到两串 *完全相同* 的输出。（我发现这颠覆了我对“随机数”这个概念的认知，因为我可不希望能够预测到“随机”的结果。）如果两个 `Rand` 对象使用了不同的值来做种子，就不具有这种相同的行为了。

## `crypto` 包

现在，我们来看一下 `crypto/rand` 包。这是一个精密和精确的 API 接口。我的理解是，它基于操作系统底层的随机数生成器，生成完全不同的随机序列。唯一的问题是：我要如何使用它？？？我能够得到一个随机的 0 和 1 的字节切片，但是怎么处理呢？这个不像 `math/rand` 包那么易于使用，不是吗？

嗯，是否可以既得到 `crypto/rand` 包的真随机性，又获得 `math/rand` 包的易用性呢？或许真正的问题是：如何将这两个截然不同的包组合在一起？

## 一加一大于二

（注意: 参考视频 [VINTAGE 80'S REESES PEANUT BUTTER CUPS COMMERCIAL W WALKERS](https://www.youtube.com/watch?v=DJLDF6qZUX0)）

让我们深入研究下 `math/rand` 包。我们通过一个 `rand.Source` 来实例化 `rand.Rand` 类型。但是像绝大多数 Go 惯用法一样，这个 `Source` 是一个接口。我的第六感来了，或许这就是个机会？

`rand.Source` 最主要的工作由 `Int63() int64` 函数完成，它返回一个非负 `int64` 整数（也就是说，最高位是0）。进一步改进的 `rand.Source64` 仅仅返回一个 `uint64` 类型，并没有对最高位有任何限制。

你们说，我们使用源自 `crypto/rand` 包的功能来尝试创建一个 `rand.Source64` 对象如何？（你可以参考在 [Go Playground](https://play.golang.org/p/_3w6vWTwwE) 上的代码。）

首先，我们为我们的 `rand.Source64` 创建一个结构。（同时需要注意：因为 `math/rand` 和 `crypto/rand` 使用的时候会发生冲突，在下面的代码中，我们将依次使用 `mrand` 和 `cand` 来代替。）

```go
type mySrc struct{}
```

让我们来为接口声明 `Seed(...)` 函数。我们不需要一个和 `crypto/rand` 包交互的种子，所以没有具体代码。

```go
func (s *mySrc) Seed(seed int64) { /*no-op*/ }
```

因为 `Uint64()` 函数返回值取值范围**最广(widest)**，需要 64 位的随机数，因此我们首先实现它。我们使用 `encoding/binary` 包从`crypto/rand` 包的 `io.Reader` 接口中读取 8 个字节的数据，并直接转换成 `uint64`。

```go
func (s *mySrc) Uint64() (value uint64) {
	binary.Read(crand.Reader, binary.BigEndian, &value)
	return value
}
```

`Int63()` 函数和 `Uint64()` 函数类似，我们只要保证最高位为 0 即可。这个相当简单，只需要在 `Uint64()` 返回值的基础上做一个快速的位掩码操作即可。

```go
func (s *mySrc) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}
```

非常棒！现在我们有了完整版的 `rand.Source64` 实现了。让我们进行一些测试来验证。

```go
var src mrand.Source64
src = &mySrc{}
r := mrand.New(src)
fmt.Printf("%d\n", r.Intn(23))
```

## 权衡

酷，通过上面短短十几行代码，我们有了一个非常简单的解决方案，将密码安全的随机数生成和 `math/rand` 包友好方便的 API 有机的结合在一起。然而，我逐渐认识到没有什么事情是免费的。使用这个方案的代价是什么？让我们来对这段代码进行[性能分析](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)。

(注意: 我喜欢在测试中使用质数，所以你会看到许多 7919 作为参数，它是第 1000 个质数。)

`math/rand` 包中顶级函数的性能到底如何？

```go
func BenchmarkGlobal(b *testing.B) {
	for n := 0; n < b.N; n++ {
		result = rand.Intn(7919)
	}
}
```

还不错！在我的笔记本上大约 38 ns/op。

```
BenchmarkGlobal-4         50000000        37.7 ns/op
```

如果创建一个以当前时间作为种子的 `rand.Rand` 实例，情况会如何呢？

```go
func BenchmarkNative(b *testing.B) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for n := 0; n < b.N; n++ {
		result = random.Intn(7919)
	}
}
```

大约 23 ns/op，相当不错！

```
BenchmarkNative-4         100000000        22.7 ns/op
```

现在，让我们测试一下我们写的新种子方案。

```go
func BenchmarkCrypto(b *testing.B) {
	random := rand.New(&mySrc{})
	for n := 0; n < b.N; n++ {
		result = random.Intn(7919)
	}
}
```

哎呀，大约 900 ns/op，这个代价太昂贵了。是不是什么地方我们搞错了？或者这就是使用 `crypto/rand` 包需要付出的代价？

```
BenchmarkCrypto-4      2000000       867 ns/op
```

让我们测试一下单独读取 `crypto/rand` 需要多长时间。

```go
func BenchmarkCryptoRead(b *testing.B) {
	buffer := make([]byte, 8)
	for n := 0; n < b.N; n++ {
		result, _ = crand.Read(buffer)
	}
}
```

好，结果显示，我们新的解决方案中绝大部分时间花在了与 `crypto/rand` 包的交互上面。

```
BenchmarkCryptoRead-4      2000000       735 ns/op
```

我不知道如何做才能进一步提高性能。而且，或许对于你的使用场景来说，花费大约1毫秒来获取非特定随机数不是一个问题。这个需要你自己去评估了。

## 另外一种思路？

我最熟悉的随机化的用法之一是[指数退避](https://en.wikipedia.org/wiki/Exponential_backoff)工具。这样做的目的是在重新连接到有压力的服务器时减少偶然同步的几率，因为有规律的负荷可能会对服务器的恢复造成伤害。在这些场景中，“确定性随机”行为本身不是一个问题，但是在一群实例中使用相同的种子会存在问题。

并且，使用顶级 `math/rand` 函数的时候，无论是使用缺省的种子（即以隐含的 1 为种子），还是使用非常容易观察的 `time.Now().UnitNano()` 范式来做种子，这都会是一个问题。如果你的服务碰巧在同一时间启动，会在确定随机输出导致意外同步的情况下，服务被迫中止退出。

如果我们在实例化的时候使用 `crypto/rand` 的强大能力来产生 `math/rand` 工具的种子，在之后，我们依然可以享受到确定性随机工具带来的性能，这个主意怎么样？

```go
func NewCryptoSeededSource() mrand.Source {
	var seed int64
	binary.Read(crand.Reader, binary.BigEndian, &seed)
	return mrand.NewSource(seed)
}
```

我们可以重新对新代码做性能分析，但是我们早已经知道，性能将回到确定性随机的情况下。

```go
func BenchmarkSead(b *testing.B) {
	random := mrand.New(NewCryptoSeededSource())
	for n := 0; n < b.N; n++ {
		result = random.Intn(7919)
	}
}
```

现在，我们证实了我们的假设是正确的。

```
BenchmarkSeed-4           50000000        23.9 ns/op
```

## 关于作者

嗨，我是内尔·卡彭铁尔。我是旧金山 [Orion Lab](https://www.orionlabs.io/) 的资深软件工程师。我已经写了三年的 Go 代码，当快速熟悉了之后，Go 已经成为我最喜欢的语言之一了。

免责声明：我既不是安全专家，也不是跨平台 `crypto/rand` 实现专家。如果你要在关键安全任务用例中使用这些工具，你可以咨询当地的安全专家。

你可以从[这里](https://github.com/orion-labs/go-crypto-source)获取一份精炼版的代码示例。它遵循 Apache 2.0 授权，所以你可以随意剪切和借鉴任何你需要的代码！

---

via: https://blog.gopheracademy.com/advent-2017/a-tale-of-two-rands/

作者：[Nelz Carpentier](https://blog.gopheracademy.com/advent-2017/a-tale-of-two-rands/)
译者：[arthurlee](https://github.com/arthurlee)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出。
