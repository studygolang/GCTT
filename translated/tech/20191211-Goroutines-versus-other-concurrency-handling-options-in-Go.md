# Go 中的 goroutine 和其他并发处理方案的对比

Go 语言让使用 goroutine 和通道变得非常有吸引力，作为在 Go 中进行并发的主要方式，它们是被有意识的提出的。因此对于你所遇到的任何与并发相关的问题，它们都可能成为首选方案。但是我不确定它们是否适合于我遇到的所有问题，我仍在考虑其中的平衡点。

通道和 goroutine 对于查询共享状态（或从共享状态中获取某些信息）这类问题看起来似乎并不完全契合。假设你想要记录那些与服务端建立 TLS 通信失败的 SMTP 客户端的 IP，以便在 TLS 握手失败的情况下，不再提供 TLS 通信（或至少在给定的时间段内不提供）。大多数基于通道的解决方案都很直白：一个主 goroutine 维护一个 IP 集合，通过通道向主 goroutine 发送一条消息来向其添加 IP。但是，如何询问主 goroutine 某个 IP 是否已经存在？问题的关键在于，无法在共享通道上收到来自主 goroutine 的答复，因为主 goroutine 无法专门答复你。

针对这个问题，目前我看到的基于通道的解决方案是将一个回复通道作为查询消息的一部分一起发送给主 goroutine（通过共享通道发送）。但是这种方法的有个副作用，那就是通道的频繁分配和释放，每次请求都会对通道进行分配、初始化、使用一次然后销毁（我认为这些通道必须通过垃圾回收机制回收，而不是在栈上分配和释放）。另一种方案是提供一个由 sync 包中锁或其他同步工具显式保护的数据结构，这是更底层的解决方案，需要更多的管理操作，但是却避免了通道的频繁分配和释放。

对于我将要编写的大多数 Go 程序来说，效率往往不是首要的关注点，真正的问题是，如何使编写更容易并且代码更清晰。目前我没有一个彻底的结论，但却有一个初步的、并不完全是我所期待的那个：如果要针对同一共享状态处理不止一种查询，那么基于锁的方案会更容易些。

在面对多种类型的状态查询时，通道方案的问题在于它需要很多我称之为“类型官僚主义”的东西。因为通道是拥有类型的，所以对于每种不同类型的答复都需要定义相应的答复通道类型（显式或隐式）。然后，基本上每个不同的查询也都需要自己的类型，因为查询消息必须包含（类型化的）回复通道。基于锁的方案并不会使这些类型相关的琐事消失，但会减轻它们带来的痛苦，因为此时查询消息和答复只是函数的参数和返回值，因此不必将它们正式的定义为 Go 类型（struct）。实际上，即便需要进行额外的手动加锁，这对我来说已经是很轻松了。

（可以通过各种手段将这些类型合并在一起从而减少通道方案中所需类型的数量，但是此时便开始失去类型的安全性，尤其是编译时类型检查。我喜欢 Go 中的编译时类型检查，因为它会很靠谱的告诉我是否遇到了明显的错误，并且这也有助于加快重构的速度。）

从某种意义上说，我认为通道和 goroutine 是 Turing tarpit 的一种形式，因为如果你足够聪明的话，可以将它们应用于所有问题。

（另一方面，[有时通道是解决看似与它们无关的问题的绝佳方案](http://blog.golang.org/two-go-talks-lexical-scanning-in-go-and)，在看到该文章之前，我从未想过在词法分析器中使用 goroutine 和通道。）

## 侧边栏：我所采用的 Go 锁模式

这不是我的原创，而是来源于 Go blog 中的 [Go maps 实战](http://blog.golang.org/go-maps-in-action)一文，如下所示：

```go
// ipEnt 是共享数据结构中的真实条目。
type ipEnt struct {
  when  time.time
  count int
}

// ipMap 是由读写锁保护的共享数据结构。
type ipMap struct {
  sync.RWMutex
  ips map[string]*ipEnt
}

var notls = &ipMap{ips: make(map[string]*ipEnt)}

// Add 方法用于外部调用对共享数据进行操作，每次都会首先获取锁，随后释放它。
func (i *ipMap) Add(ip string) {
  i.Lock()
  ... manipulate i.ips ...
  i.Unlock()
}
```

使用方法来操作数据结构感觉是最自然的方式，部分原因是由于操作和锁定的约束条件紧密的耦合在一起。而我喜欢它纯粹是因为它所带来的简洁书写方式：

```go
if res == TLSERROR {
  notls.Add(remoteip)
  ....
}
```

最后这点只是个人喜好，当然，也有人更喜欢将 `ipMap` 作为参数传递给独立的函数。

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoGoroutinesVsLocks

作者：[ChrisSiebenmann](https://utcc.utoronto.ca/~cks/space/People/ChrisSiebenmann)
译者：[anxk](https://github.com/anxk)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/)
