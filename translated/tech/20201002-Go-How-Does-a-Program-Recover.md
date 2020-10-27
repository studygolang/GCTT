# Go：程序如何恢复（recover）？
![由Renee French创作的原始Go Gopher作品，为“ Go的旅程”创建插图。](https://github.com/studygolang/gctt-images2/blob/master/20201002-Go-How-Does-a-Program-Recover/1_4zRau44piN5HjUnTnJsMOw.png?raw=true)

当程序无法适当处理错误时，比如无效的内存访问，Go 中的 panic 就会被触发。如果错误是意料之外，且没有其他方式处理该错误时，同样可以由开发者触发 panic。了解 recover 或者终止的过程，可以更好地理解一个会发生 panic 的程序的后果。

## 多帧的情况
关于 panic 以及它 recover 函数的经典例子已经有着充分的说明，该例子收录在 Go blog 文章“[Defer, Panic, and Recover](https://blog.golang.org/defer-panic-and-recover)” 中。让我们关注下其他例子，当一个 panic 涉及多个 defer 函数的帧（frame）。这里是一个例子：

![](https://github.com/studygolang/gctt-images2/blob/master/20201002-Go-How-Does-a-Program-Recover/a-panic-involves-multiple-frames-of-deferred-functions.png?raw=true)

该程序由三个链式调用的函数组成。一旦这段代码到了最后层级产生 panic 的地方，Go 会构建 defer 函数的第一个帧并运行它：

![](https://github.com/studygolang/gctt-images2/blob/master/20201002-Go-How-Does-a-Program-Recover/build-the-first-frame-of-deferred-functions.png?raw=true)

这个帧里面的代码没有 recover 这个 panic。之后，Go 构建父帧（译者注：level1 函数的帧），并在该帧中调用其中的每个延迟函数：

![](https://github.com/studygolang/gctt-images2/blob/master/20201002-Go-How-Does-a-Program-Recover/builds-the-parent-frame-and-calls-each-deferred-function.png?raw=true)

*提醒一下，defer 函数 按照 LIFO（后进先出）的顺序执行。想要了解更多关于 defer 函数内部管理的方式，建议阅读我的文章“[Go: How Does defer Statement Work?](https://medium.com/a-journey-with-go/go-how-does-defer-statement-work-1a9492689b6e)”*

由于一个函数 recover 了 panic，Go 需要一种跟踪，并恢复这个程序的方法。为了达到这个目的，每一个 goroutine 嵌入了一个特殊的属性，指向一个代表该 panic 的对象：

![](https://github.com/studygolang/gctt-images2/blob/master/20201002-Go-How-Does-a-Program-Recover/special-attribute.png?raw=true)

当 panic 发生的时候，该对象会在运行 defer 函数前被创建。然后，recover 这个 panic 的函数仅仅返回这个对象的信息，同时将这个 panic 标记为已恢复（recovered）：

![](https://github.com/studygolang/gctt-images2/blob/master/20201002-Go-How-Does-a-Program-Recover/returns-the-information-of-that-object.png?raw=true)

一旦 panic 被认为已经恢复，Go 需要恢复当前的工作。但是，由于运行时处于 defer 函数的帧中，它不知道恢复到哪里。出于这个原因，当 panic 标记已恢复的时候，Go 保存当前的程序计数器和当前帧的堆栈指针，以便 panic 发生后恢复该函数：

![](https://github.com/studygolang/gctt-images2/blob/master/20201002-Go-How-Does-a-Program-Recover/Go-saves-the-current-program-counter-and-stack-pointer.png?raw=true)

我们也可以使用 `objdump` 查看 程序计数器的指向（e.g. `objdump -D my-binary` | `grep 105acef`）：

![](https://github.com/studygolang/gctt-images2/blob/master/20201002-Go-How-Does-a-Program-Recover/objdump.png?raw=true)

该指令指向函数调用 `runtime.deferreturn`，这个指令被编译器插入到每个函数的末尾，而它运行 defer 函数。在前面的例子中，这些 defer 函数中的大多数已经运行了——直到恢复，因此，只有剩下的那些会在调用者返回前运行。

## WaitGroup
理解这个工作流程会让我们了解 defer 函数的重要性以及它如何起作用，比如，在处理若干个 goroutine 的时候，在一个 defer 函数中延迟调用 `WaitGroup` 可以避免死锁。这是一个例子：

![](https://github.com/studygolang/gctt-images2/blob/master/20201002-Go-How-Does-a-Program-Recover/WaitGroup.png?raw=true)

这个程序由于 `wg.Done` 无法被调用而导致死锁。将它移动到一个 defer 函数中会确保执行并且能让这个程序继续运行。

## Goexit
有趣的是，函数 `runtime.Goexit` 使用完全相同的工作流程。`runtime.Goexit` 实际上创造了一个 panic 对象，且有着一个特殊标记来让它与真正的 panic 区别开来。这个标记让运行时可以跳过恢复以及适当的退出，而不是直接停止程序的运行。

---
via: https://medium.com/a-journey-with-go/go-how-does-a-program-recover-fbbbf27cc31e

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[dust347](https://github.com/dust347)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
