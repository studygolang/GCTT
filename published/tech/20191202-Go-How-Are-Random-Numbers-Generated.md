首发于：https://studygolang.com/articles/25930

# Go：随机数是怎样产生的？

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191202-Go-How-Are-Random-Numbers-Generated/01.png)
<p align="center">Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.</p>

*这篇文章基于 Go 1.13 版本*

Go 实现了两个包来产生随机数：

- 在包 `math/rand` 的一个伪随机数生成器（ PRNG ）
- 在包 `crypto/rand` 中实现的加密伪随机数生成器（ CPRNG ）

如果这两个包都产生了随机数，则将基于真正的随机数和性能之间取舍

## 确定的结果

Go 的 `rand` 包会使用相同的源来产生一个确定的伪随机数序列。这个源会产生一个不变的数列，稍后在执行期间使用。将你的程序运行多次将会读到一个完全相同的序列并产生相同的结果。让我们用一个简单的例子来尝试一下：

```go
func main() {
   for i := 0; i < 4; i++  {
      println(rand.Intn(100))
   }
}
```

多次运行这个程序将会产生相同的结果：

```
81
87
47
59
```

由于源代码已经发布到 Go 的官方标准库中，因此任何运行此程序的计算机都会得到相同的结果。但是，由于 Go 仅保留一个生成的数字序列，我们可能想知道 Go 是如何管理用户请求的时间间隔的。Go 实际上使用此数字序列来播种一个产生这个随机数的源，然后获取其请求间隔的模。例如，运行相同的程序，最大值为 10，则模 10 的结果相同。

```
1
7
7
9
```

让我们来看一下如何在每次运行我们的程序时得到不同的序列。

## 播种

Go 提供一个方法， `Seed(see int64)` ，该方法能让你初始化这个默认序列。默认情况下，它会使用变量 1。使用另一个变量将会提供一个新的序列，但会保持确定性：

```go
func main() {
   rand.Seed(2)
   for i := 0; i < 4; i++  {
      println(rand.Intn(100))
   }
}
```

这些是新的结果：

```
86
86
92
40
```

在你每次运行这个程序时，这个序列将会保持不变。这是构建此序列的工作流：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191202-Go-How-Are-Random-Numbers-Generated/02.png)

<p align="center">The sequence is pre-generated at the bootstrap</p>
获取一个全新序列的解决方案是使用一个在运行时能改变的变量，比如当前时间：

```go
func main() {
   rand.Seed(time.Now().UnixNano())
   for i := 0; i < 3; i++  {
      println(rand.Intn(100))
   }
}
```

由于当前纳秒数在任何时刻都是不同的，因此这个程序每次运行都会使用一个不同的序列。然而，尽管这个序列在每次运行都是不同的，可这些数字仍是伪随机数。如果你准备牺牲性能来获得更好的随机性，那么 Go 已经为你提供了另一种实现方式。

## 随机数生成器

Go 的标准库也提供了一个适用于加密应用的随机数生成器。因此，理所当然的，生成的随机数并不固定，并且一定会提供更好的随机性。这有一个例子使用了这个新包 `cryto/rand` ：

```go
func main() {
   for i := 0; i < 4; i++  {
      n, _ := rand.Int(rand.Reader, big.NewInt(100))
      println(n.Int64())
   }
}
```

这是结果：

```
12
24
56
19
```

多次运行这个程序将会得到不同的结果。在内部，Go 应用了如下规则：

> *在 Linux 和 FreeBSD 系统上，Reader 会使用 getrandom(2) （如果可用的话），否则使用 /dev/urandom。*
>
> *在 OpenBSD 上，Reader 会使用 getentropy(2)。*
>
> *在其他的类 Unix 系统上，Reader 会读取 /dev/urandom。*
>
> *在 Windows 系统上，Reader 会使用 CryptGenRandom API.*
>
> *在 Wasm 上，Reader 会使用 Web Cryto API。*

但是，获得更好的质量意味着性能降低，因为它必须执行更多的操作并且不能使用预生成的序列。

## 性能

为了理解生成随机数的两种不同方式之间的折衷，我基于先前的两个例子运行了一个基准测试。结果如下：

```
name    time/op
RandWithCrypto-8  272ns ± 3%
name    time/op
RandWithMath-8   22.8ns ± 4%
```

不出所料，`crypto` 包更慢一些。但是，如果你不用去处理安全的随机数，那么 `math` 包就足够了并且它将会给你提供最好的性能。

你也可以调整默认数字生成器，由于内部互斥锁的存在，它是并发安全的。如果生成器并不在并发环境下使用，那么你就可以在不使用锁的情况下创建你自己的生成器：

```go
func main() {
   gRand := rand.New(rand.NewSource(1).(rand.Source64))
   for i := 0; i < 4; i++  {
      println(gRand.Intn(100))
   }
}

```

性能会更好：

```
name                  time/op
RandWithMathNoLock-8  10.7ns ± 4%
```

---

via：https://medium.com/a-journey-with-go/go-how-are-random-numbers-generated-e58ee8696999

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[sh1luo](https://github.com/sh1luo)
校对：[lxbwolf](https://github.com/lxbwolf)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
