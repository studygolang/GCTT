# Go 语言中 runtime.KeepAlive() 方法的一些随笔

我在看 go101 网站的 [类型不安全指针](https://go101.org/article/unsafe.html)[(来源)](https://old.reddit.com/r/golang/comments/8ll6lf/how_to_safely_use_typeunsafe_pointers_in_go/) 一文时，偶然发现了 [runtime](https://golang.org/pkg/runtime/) 库的一个有趣的新方法 [runtime.KeepAlive()](https://golang.org/pkg/runtime/#KeepAlive) 的一个用法。刚开始我对于怎么使用它是很困惑的， 那么按我的性格肯定要探究它是怎么工作的。

`runtime.KeepAlive` 所做的事就是使一个变量保持 '存活'，这就意味着它（或者它引用的变量）不会被垃圾收集，而且它所注册的任何终止器（finalizer）都不会被执行。 [这个文档](https://golang.org/pkg/runtime/#KeepAlive) 中有一个如何使用它的例子。我的第一个疑问是为什么在代码中 `runtime.KeepAlive()` 的使用时机那么的靠后；我比较希望它能够更早的被调用，就像终止器被注入时，但是后来我明白了它这样做的真正意图。 简而言之， `runtime.KeepAlive()` 是调用一下变量。显而易见的，一个变量直至它的最后一次使用期间都是存活的，所以如果你在后面使用一个变量，那么 Go 必须让它一直存活到最后使用的时候。

一方面，`runtime.keepAlive` 没有什么神奇的地方；任何一种使用某个变量的方式，都会使它保持存活。另一方面，`runtime.KeepAlive()` 是一种很重要的魔法，它表示 Go 保证了你所使用的变量不会被优化清除掉，因为编译器能明白没有什么能真正依赖于你的使用。虽然有很多其它的方式来使用一个变量，但即使是最聪明的方式也很容易受到编译器的影响，最聪明的方式也会有不利的一面，他们会影响 [Go 的智能合理逃逸分析](https://utcc.utoronto.ca/~cks/space/blog/programming/GoReflectEscapeHack)，强行将一个本属于本地栈的变量分配到堆上。

关于`runtime.KeepAlive()` 的另一个特殊戏法是它的是实现方式，代码里什么都没做。实际上，它不是作为一个被调用的函数，而是由 [ssa.go](https://github.com/golang/go/blob/master/src/cmd/compile/internal/gc/ssa.go#L2828) 实现的编译器内部实现，类似于 `unsafe.Pointer`。当你的代码中使用了 `runtime.KeepAlive()`，Go 编译器会设置一个名为 `OpKeepAlive` 的静态单赋值(SSA)，然后剩余的编译就会知道将这个变量的存活期保证到使用了 `runtime.KeepAlive()` 的时刻。

（阅读 ssa.go 的初始化函数是很有趣的。不出所料，有许多语义化包函数调用被直接映射到将指令内联在代码中，如math.Sqrt。有些是平台相关的，包括 [bits](https://golang.org/pkg/math/bits/) 的函数）

`runtime.KeepAlive()` 是一个特别的魔法有一个直接的后果就是你不能得到它的地址。如果你这样做的话， Go 会报错：

```go
./tst.go:20:22: cannot take the address of runtime.KeepAlive
```

我不知道 Go 是否会聪明地优化掉一个只调用 `runtime.KeepAlive` 的函数, 但希望你永远不需要间接调用 `runtime.KeepAlive`。

PS：尽管我很想说没有人应该需要对分配在栈上的本地变量（包括参数）调用 `runtime.KeepAlive`，因为在函数返回之前栈是不会被回收的，但这是一个危险的假设。编译器可以非常聪明地为两个不同的、没有重叠生存期的变量重用堆栈槽，或者简单地告诉垃圾收集它已经完成了某些工作（例如，用 nil 覆盖指向对象的指针）。

---
via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoRuntimeKeepAliveNotes

作者：[ChrisSiebenmann](https://utcc.utoronto.ca/~cks/space/People/ChrisSiebenmann)
译者：[yuhang-dong](https://github.com/yuhang-dong)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出