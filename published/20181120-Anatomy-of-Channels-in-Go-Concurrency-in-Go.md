首发于：https://studygolang.com/articles/17683

# Go 中的 channel 解析— Go 中的并发性

## 什么是 channel ？

**channel** 是一个通信对象，goroutine 可以使用它来相互通信。 从技术上讲，channel 是一个用于数据传输的管道，可以将数据**传入或从中读取**。 因此，一个 Goroutine 可以将数据发送到一个 channel ，而其他 Goroutine 可以从同一个 channel 读取该数据。

## 声明一个 channel

Go 提供 `chan` 关键字来创建一个 channel。channel 只能用于传输**一种数据类型**的数据。不允许从该 channel 传输其他数据类型。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_zAzzrOTw_BUo2BLzUNgxFA.png)

https://play.golang.org/p/iWOFLfcgfF-

上面的程序声明了一个 channel `c`，它可以传输 int 类型的数据。上面的程序输出为 `<nil>`，因为 channel 的零值为 `nil` ( 空 ) 但是 `nil` ( 空 ) channel 是不能被使用的。你不能将数据传递给一个 `nil` ( 空 ) 的 channel 或从 `nil` ( 空 ) channel 读取数据。因此，我们必须使用 `make` 函数来创建一个可以使用的 channel。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_mLdlgjuT6ZllylQKPSQ5wg.png)

https://play.golang.org/p/N4dU7Ql9bK7

我们使用了短命名法 := 来使用 `make` 函数创建 channel。以上程序产生以下结果

```go
type of `c` is chan int
value of `c` is 0xc0420160c0
```

注意 channel `c` 的值，它看起来像一个内存地址。默认情况下，channel 是指针。大多数情况下，当您希望与 Goroutine 进行通信时，您将 channel 作为参数传递给函数或方法。因此，当 Goroutine 接收该 channel 作为参数时，您不需要解除对它的引用来从该 channel 发送或读取数据。

## 数据读写

Go 提供了非常容易记住的**左箭头语法** `<-` 从 channel 中读写数据。

```go
c <- data
```

上面的语法意味着我们想要将 `data` 发送或写入 channel `c`。它从 `data` 指向 channel `c`，因此我们可以想象一下将 `data` 发送到 channel `c`。

```go
<- c
```

上面的语法意味着我们需要从 channel `c` 读取一些数据，看看箭头的方向，它是从 channel `c` 开始的，这个语句没有将数据发送到任何地方，但是它仍然是一个有效的语句。如果您有一个变量用来保存来自该 channel 的数据，则可以使用以下语法

```go
var data int
data = <- c
```

现在，从 channel `c` 中读取出的 int 类型的数据可以赋值给 int 类型的变量 `data`。

上面的语法可以像下面这样使用短命名法重写

```go
data := <- c
```

Go 将判断出在 channel `c` 中传输的数据的数据类型，并为变量 `data` 提供一个有效的数据类型。

**以上所有 channel 操作在默认情况下都是阻塞的**。在[上节课](https://medium.com/rungo/anatomy-of-goroutines-in-go-concurrency-in-go-a4cb9272ff88) 中，我们看到了 `time.Sleep` 阻塞了 Goroutine。channel 操作在本质上也是阻塞的。当一些数据被写入 channel 时，goroutine 会被阻塞，直到其他 Goroutine 从该 channel 读取数据。同时，正如我们在[并发一章](https://studygolang.com/articles/16766) 中看到的，channel 操作告诉调度器调度另一个 Goroutine，这就是为什么程序不会永远阻塞在同一个 Goroutine 上。channel 的这些特性在 Goroutines 通信中非常有用，因为它可以避免了我们用互斥锁来让它们相互协作。

## 在实践中使用 Channel

上面我们讲的已经很多了，现在让我们来看一下在 Goroutine 中使用 channel 。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_p4wStpy2mC9eahgkta5Olw.png)

https://play.golang.org/p/OeYLKEz7qKi

让我们一步一步地来讨论上述程序的执行。

- 我们声明了 `greet` 函数，它接受传输数据类型为 `string` 的 channel `c`。在这个函数中，我们从 channel `c` 中读取数据并将数据打印到控制台。
- 在 main 函数中，程序将 `main() started` 作为第一条语句打印到控制台。
- 然后使用 `make` 函数创建了用于传输 `string` 类型的 channel `c`。
- 我们将 channel `c` 传递给 `greet` 函数，然后使用 `go` 关键字将其作为一个 Goroutine 运行。
- 此时，程序有 2 个 Goroutine，而主 Goroutine 是 `main Goroutine` ([查看上一课了解它是什么](https://medium.com/rungo/anatomy-of-goroutines-in-go-concurrency-in-go-a4cb9272ff88))。然后程序运行下一行。
- 我们将字符串 `John` 传入 channel `c`。此时，goroutine 被阻塞，直到某个 Goroutine 读取它。Go 调度程序调度 `greet`goroutine，然后它开始执行，正如上面第一点说道的。
- 然后 `main Goroutine` 被激活并执行最后的语句，打印 `main()stopped` 然后停止。

## 死锁

如上面所述，当我们往 channel 写入或从中读取数据时，goroutine 将被阻塞并将控制权传递给可用的 Goroutine。如果没有其他可用的 Goroutines，那么可以想象他们都在睡觉。这就是死锁错误发生的地方，那样会导致整个程序崩溃。

> 如果您试图从 channel 中读取数据，但是 channel 中没有可用的值，它将阻塞当前的 Goroutine 并且会阻塞其他 Goroutine，希望一些 Goroutine 将值发送到 channel。因此，**这个读取操作将会被阻塞**。类似地，如果要将数据发送到一个 channel，它将阻塞当前的 Goroutine 并解除其他 Goroutine 的阻塞，直到某个 Goroutine 从它读取数据。因此，**这个发送操作将被阻塞**。

死锁的一个简单例子就是在 main Goroutine 中执行一些 channel 操作。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_gZXmlZwi61MQsngF0NdYLg.png)

<https://play.golang.org/p/2KTEoljdci_f>

上面的程序在运行时会抛出下面的错误。

> ```
> main() started
> fatal error: all Goroutines are asleep - deadlock!
> Goroutine 1 [chan send]:
> main.main()
>         program.Go:10 +0xfd
> exit status 2
> ```

**fatal error: all Goroutines are asleep — deadlock!**. 似乎所有的 Goroutine 都在睡觉，或者根本没有其他 Goroutine 可供使用。

## 关闭一个通道

一个 channel 可以被关闭，这样就不能通过它发送更多的数据了。接收端 Goroutine 可以通过它 `val, ok := <- channel` 了解 channel 的使用状态，如果 channel 是打开的或读取操作是可以执行的，那么 `ok` 的值等于 `true` 如果通道关闭那么就不能执行更多的读取操作，此时 `ok` 等于 `false` ，channel 可以使用带有语法的内置函数 `close` 如，`close(chennel)` 来关闭 channel ，让我们来看一个小例子。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_BAgOmJUgC3w0QVxga60CyA.png)

<https://play.golang.org/p/LMmAq4sgm02>

> 为了帮助您理解阻塞的概念，首先发送操作 `c <- "John"` 是阻塞的，一些 Goroutine 必须从 channel 中读取数据，因此 `greet` 这个 Goroutine 由 Go 调度器调度。然后，第一次读取操作 `<-c` 是非阻塞的，因为要从 channel `c` 中读取数据。第二次读取操作 `<-c` 将阻塞，因为 channel `c` 没有任何数据可以读取，因此 Go 调度器激活 `main` Goroutine，程序从 `close(c)` 函数开始执行。

从上面的错误中，我们可以看到我们试图往一个已经关闭的 channel 里发送数据。此外，如果我们试图从关闭的 channel 阅读，程序会发生 panic。为了更好地理解被关闭 channel 的可用性，让我们看看 `for` 循环。

## For 循环

`for` 循环的无限语法 `for{}` 可用于读取通过 channel 发送的多个值。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_ibev09eQ9yfzxBCmaBsfQA.png)

<https://play.golang.org/p/X58FTgSHhXi>

在上面的例子中，我们创建了一个 `squares` Goroutine，它将逐一返回从 0 到 9 的数字。在 `main` Goroutine 中，我们用无限 `for` 循环来读取那些数字 。

在无限 `for` 循环中，由于我们需要一个条件来在某一点中断循环，所以我们使用语法 `val, ok := <-c` 从 channel 中读取值。在这里，`ok` 会在 channel 关闭时给我们提供额外的信息。因此，在 `square` Goroutine 中，在写完所有数据之后，我们使用语法 `close(c)` 关闭 channel。当 `ok` 的值为 `true` 时，程序打印 `val` 和 `ok` 的值。当它为 `false` 时，我们使用 `break` 关键字跳出循环。因此，上述程序产生以下结果

```
main() started
0 true
1 true
4 true
9 true
16 true
25 true
36 true
49 true
64 true
81 true
0 false <-- loop broke!
main() stopped
```

> 当 channel 关闭时，goroutine 读取的值为 channel 数据类型的零值。在这种情况下，由于 channel 传输的是 int 数据类型，因此结果为 0。

为了避免手动检查 channel 关闭情况带来的痛苦，Go 为我们提供了 `for range` 循环 ，当 channel 关闭时 `for range` 将自动关闭。让我们修改前面的程序。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_Pr2CGe7m6KyK-RWCrSAhJw.png)

<https://play.golang.org/p/ICCYbWO7ZvD>

在上面的程序中，我们用 `for val:= range c` 代替 `for{}`。`range` 将每次从 channel 中读取一个值，直到 channel 关闭。因此，上面的程序产生下面的结果

```
main() started
0
1
4
9
16
25
36
49
64
81
main() stopped
```

> 如果最后不关闭 `for range` 循环中的 channel，程序将在运行时抛出死锁错误。

## 缓冲区大小或 channel 容量

正如我们看到的，channel 上的每个发送操作都会阻塞当前的 Goroutine。但到目前为止，我们使用的 `make` 函数没有第二个参数。第二个参数是 channel 或缓冲区大小的容量。默认情况下，channel 缓冲区大小为 0 也称为**无缓冲 channel**。写入 channel 的任何内容都必须是可以读取的。

当缓冲区大小为非零 n 时，**goroutine 直到缓冲区满后才被阻塞**。当缓冲区满时，发送到 channel 的任何值都将通过抛出缓冲区中可供读取的最后一个值 ( Goroutine 将被阻塞 ) 添加到缓冲区中。但有一个陷阱，**读操作对缓存是持续性的**。这意味着，一旦读操作开始，它将一直持续下去，直到缓冲区为空。从技术上讲，**这意味着读取缓冲区 channel 的 Goroutine 在缓冲区为空之前不会阻塞**。

我们可以使用以下语法定义缓冲 channel。

```go
c := make(chan Type, n)
```

这将创建一个缓冲区大小为 `n` 数据类型为 `Type` 的 channel。在 channel 接收到 `n+1` 发送操作之前，它不会阻塞当前的 Goroutine。
让我们来证明一下 Goroutine 在 channel 缓冲区满之前不会阻塞。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_GseXQg6_UiqX71YNzuIzHw.png)

<https://play.golang.org/p/k0usdYZfp3D>

在上述程序中，channel `c` 的缓冲容量为 3。这意味着它可以存储 3 个值，也就是第 20 行的值。但是由于缓冲区没有满 ( 因为我们没有发送任何新值 )，主 Goroutine 将不会阻塞，程序将会继续。

让我们发送额外的值。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_Ac_JV7syiGYodvbKnhntSA.png)

<https://play.golang.org/p/KGyiskRj1Wi>

如前所述，现在填充的缓冲区通过 `c <- 4` 发送操作，主 Goroutine 将等待 `square` Goroutine 读取所有值。

## channel 的长度和容量

与切片相似，缓冲 channel 具有长度和容量。channel 长度是 channel 缓冲区中排队 ( 未读 ) 的值个数，而 channel 容量是缓冲区大小。为了计算长度，我们使用 `len` 函数，而为了计算容量，我们使用 `cap` 函数，就像切片一样。

![](https://cdn-images-1.medium.com/max/1000/1*LC30okTTzTrXXNwmsYr5KA.png)

<https://play.golang.org/p/qsDZu6pXLT7>

如果您想知道，为什么上面的程序运行良好，死锁错误没有抛出。这是因为，由于 channel 容量为 3，且缓冲区中只有 2 个值可用，Go 没有试图通过阻塞主 Goroutine 执行来调度另一个 Goroutine。如果需要，可以在 main Goroutine 中读取这些值，因为即使缓冲区没有满，也不能阻止从 chennel 读取值。

这是另外一个例子

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_7qH7EdSAWuy21hAzFfWypQ.png)

<https://play.golang.org/p/-gGpm08-wzz>

这里有一个脑筋急转弯

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_xtW_yhzwEX-JbEVSE3S0eA.png)

<https://play.golang.org/p/sdHPDx64aor>

使用 `for range` 来读取有缓存 channel，我们可以从已经关闭的 channel 读取。因为对于已经关闭的 channel，数据驻留在缓冲区中，我们仍然可以读取该数据。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_gjYR-GuBnQsJIcEJ8RbTNg.png)

<https://play.golang.org/p/vULFyWnpUoj>

> 缓冲 channel 就像毕达哥拉斯杯，观看这个关于[毕达哥拉斯杯](https://www.youtube.com/watch?v=ISfIT3B4y6E) 的有趣视频。

## 与多个 Goroutine 一起工作

我们写两个 Goroutines，一个用于计算整数的平方另一个用于计算整数的立方。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_ELgCmt5V4h5zTR2VMhA_nA.png)

<https://play.golang.org/p/6wdhWYpRfrX>

让我们一步一步地讨论上述程序的执行。

- 我们创建了两个函数 `square` 和 `cube`，它们将作为 Goroutines 单独运行。两者都从 channel `c` 中接收 int 类型的数据作为变量，然后复制给 num，然后在下一行将计算完成的数据写到 channel `c`。
- 在 `main` Goroutine 中，我们使用 `make` 函数创建了两个类型为 int 的 channel `squareChan` 和 `cubeChan`。
- 然后我们运行 `square` 和 `cube` Goroutine。
- 由于此时仍在 `main` Goroutine 内，`testNum` 的值此时为 3。
- 然后我们将数据发送到 `squareChan` 和 `cubeVal`。主 Goroutine 将被阻塞，直到这些 channel 读取它。一旦他们读了它，`main`  Goroutine 将继续执行。
- 当在 `main` Goroutine 中，我们试图从给定的 channel 读取数据时，程序将被阻塞，直到这些 channel 从它们的 Goroutine 中写入一些数据。这里，我们使用了简写语法 `:=` 从多个 channel 接收数据。
- 一旦这些 Goroutine 将一些数据写入 channel ，主 Goroutine 将被阻塞。
- channel 写操作完成后，`main` Goroutine 开始执行。然后我们计算总和并打印在控制台上。

因此，上述程序将产生以下结果

```
[main] main() started
[main] sent testNum to squareChan
[square] reading
[main] resuming
[main] sent testNum to cubeChan
[cube] reading
[main] resuming
[main] reading from channels
[main] sum of square and cube of 3  is 36
[main] main() stopped
```

## 单向 channel

到目前为止，我们已经看到可以从两边传输数据的 channel，或者简单地说，我们可以在上面进行读写操作的 channel。但是我们也可以创造单向的 channel。例如，只接收允许对其进行读操作的 channel，只发送允许对其进行写操作的 channel。

单向通道也使用 `make` 函数创建，但是使用了额外的箭头语法。

```go
roc := make(<-chan int)
soc := make(chan<- int)
```

在上述程序中，`roc` 使用 `make` 函数创建箭头远离 `chan` 方向来作为只读 channel。而 `soc` 使用箭头靠近 `chan` 做为只写 channel。它们也是不同的类型。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_WzdTXUIJXWsvgHXEgMfRUw.png)

<https://play.golang.org/p/JZO51IoaMg8>

**但是单向通道有什么用呢？**使用单向 channel 可以提高程序的**类型安全性**。因此，程序不容易出错。

但是假设您有一个 Goroutine，其中您只需要从 channel 中读取数据，但是主 Goroutine 需要从 channel 中读取数据或者往 channel 写入数据。这将如何工作 ?

幸运的是，Go 提供了更简单的语法来将双向通道转换为单向通道。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_DjlR83ttXZSXt5hm3y_ZMg.png)

<https://play.golang.org/p/k3B3gCelrGv>

我们修改了 `greet` Goroutine 的例子，将双向 channel `c` 转换为只读 channel `roc` 的 `greet` 函数。现在我们只能从那个 channel 中读取。任何写操作都会导致致命的错误 : `"invalid operation: roc <- "some text" (send to receive-only type <-chan string)"`。

## 匿名 Goroutine

在 Goroutines 一章，我们学习了 匿名 Goroutines。我们还可以使用它们实现 channel。让我们修改前面的简单示例来实现匿名 Goroutine 中的 channel。

这是我们之前的例子

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_q-1jNlqnf5lNDiWHcYeNAA.png)

<https://play.golang.org/p/c5erdHX1gwR>

下面是修改后的例子，我们将 `greet` Goroutine 作为一个匿名 Goroutine。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_-9_M5GdsZT7DkK5ZwvBPoQ.png)

<https://play.golang.org/p/cM5nFgRha7c>

## channel 的数据类型

是的，channel 是第一类值，可以像其他值一样在任何地方使用：作为结构元素、函数参数、函数返回值，甚至作为另一个 channel 的类型。在这里，我们感兴趣的是使用 channel 作为另一个 channel 的数据类型。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_DH--_irN-hjWTwf9d8ZFFw.png)

<https://play.golang.org/p/xVQvvb8O4De>

## Select

`select` 就像 `switch` 一样没有任何输入参数，但是它只用于 channel 操作。`Select` 语句用于在多个 channel 中只对一个 channel 执行操作，由 `case` 块有条件地选择。

让我们先看一个例子，然后讨论它是如何工作的。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_6LuGZevT75yfzC1eAssKHg.png)

<https://play.golang.org/p/ar5dZUQ2ArH>

从上面的程序中，我们可以看到 `select` 语句就像 `switch` 一样，但是不是 `boolean` 操作，我们添加了 channel 操作，比如读或写，或者读和写混合。**`Select` 语句是阻塞的，除非它有默认情况**( 稍后我们将看到 )。一旦其中一个条件满足，它就会解除阻塞。那么，**当一个案例多个 `case` 满足呢？**

如果所有的 case 语句 ( channel 操作 ) 都阻塞了，那么 select 语句将等待其中一个 case 语句 ( 其 channel 操作 ) 解除阻塞，然后执行该 case。如果一些或所有的 channel 操作是非阻塞的，那么将随机选择一个非阻塞 case 并立即执行。

为了解释以上例子，我们启动了两个独立 channel 的 Goroutines。然后介绍了两个案例的 select 语句。一种情况从 `chan1` 读取值，另一种情况从 `chan2` 读取值。因为这些 channel 是无缓冲的，所以读操作将被阻塞 ( 写操作也一样 )。所以这两种选择都是阻塞的。因此 `select` 将等待其中一个 `case` 被解除阻塞。

当程序位于 `select` 语句时，main Goroutine 将阻塞，它将调度 select 语句中出现的所有 Goroutine ( 一次一个 )，即 `service1` 和 `service2`。`service1` 等待 3 秒，然后通过写入 `chan1` 解除阻塞。类似地，`service2` 等待 5 秒，然后通过写入 `chan2` 解除阻塞。因为 `service1` 比 `service2` 更早解除阻塞，所以 `case1` 将首先解除阻塞，因此将执行该案例，并忽略其他 case ( 这里是 case2 )。一旦完成了 case 的执行，主函数的执行将继续下去。

> 上面的程序模拟了真实的 Web 服务，其中负载均衡器收到数百万个请求，它必须从可用的服务之一返回响应。使用 Goroutines、channel 和 select，我们可以请求多个服务来响应，可以使用快速响应的服务。

为了模拟当所有情况都阻塞时，响应几乎同时可用，我们可以简单地删除 `Sleep` 函数调用。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_qcRFjcyjfx8vVhdCaYTxcw.png)

<https://play.golang.org/p/giSkkqt8XHb>

以上程序产生以下结果 ( 您可能会得到不同的结果 )

```
main() started 0s
service2() started 481 µ s
Response from service 2 Hello from service 2 981.1 µ s
main() stopped 981.1 µ s
```

但有时，它也可能是

```
main() started 0s
service1() started 484.8 µ s
Response from service 1 Hello from service 1 984 µ s
main() stopped 984 µ s
```

这是因为 `chan1` 和 `chan2` 操作几乎同时发生，但是在执行和调度上仍然存在一些时间差。

要模拟所有情况都是非阻塞且响应同时可用时，可以使用有缓冲 channel。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_D5Vmt9_n1E-ZU7nDhwGNBw.png)

<https://play.golang.org/p/RLRGEmFQP3f>

以上程序产生下面的结果

```
main() started 0s
Response from chan2 Value 1 0s
main() stopped 1.0012ms
```

在某些情况下，它也可能是

```
main() started 0s
Response from chan1 Value 1 0s
main() stopped 1.0012ms
```

在上面的程序中，两个 channel 的缓冲区中都有两个值。由于我们在缓冲区容量 2 的 channel 中发送了两个值，这些 channel 操作不会阻塞，控制将转到 `select` 语句。由于从缓冲 channel 读取是非阻塞操作，直到整个缓冲区为空，并且在 case 条件下只读取一个值，所以所有 case 操作都是非阻塞操作。因此，Go runtime 将随机选择一个 case 语句。

## `default` case

与 `switch` 语句一样，`select` 语句也有 `default` case。**`default` case 是非阻塞的**。但这还不是全部，`default` case 使得默认情况下 `select` 语句**总是非阻塞的**。这意味着，在任何 channel (*有缓冲或无缓冲*) 上的发送和接收操作总是非阻塞的。

如果某个值在任何 channel 上可用，则 `select` 将执行该 case。否则，它将立即执行 `default` case。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_u8CEJ8wBPirDCTEn-kdweA.png)

<https://play.golang.org/p/rFMpc80EuT3>

在上面的程序中，由于 channel 是无缓冲的，而且值在两个 channel 操作中不能立即可用，因此将执行 `default` case。如果上面的 `select` 语句没有 `default` case，那么 `select` 就会阻塞，而回应就会不同。

由于在 `default` 中，`select` 是非阻塞的，调度器不会从主 Goroutine 获得调度可用 Goroutine 的调用。但是我们可以通过调用 `time.Sleep` 来手动实现。这样，所有的 Goroutine 都会执行并且结束，将控制权返回给 `main` Goroutine，它会在一段时间后醒来。当主 Goroutine 唤醒时，channel 将立即具有可用的值。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_MTYFLxJ6JMIIoJxeZLzCiQ.png)

<https://play.golang.org/p/eD0NHxHm9hN>

因此，上述程序产生以下结果

```
main() started 0s
service1() started 0s
service2() started 0s
Response from service 1 Hello from service 1 3.0001805s
main() stopped 3.0001805s
```

在某些情况下，它也可能是

```
main() started 0s
service1() started 0s
service2() started 0s
Response from service 2 Hello from service 2 3.0000957s
main() stopped 3.0000957s
```

## 死锁

当没有可用的 channel 发送或接收数据时，`default` case 是有用的。为了避免死锁，我们可以使用 `default` case。这是有可能的，因为由于有 `default` case，所有 channel 操作都是非阻塞的，如果数据不能立即可用，Go 不会安排任何其他 Goroutines 发送数据到 channel。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_8EsiREe0AmKU8cwACESYYA.png)

<https://play.golang.org/p/S3Wxuqb8lMF>

与接收操作类似，在发送操作中，如果其他 Goroutine 正在休眠 (*未准备好接收值*)，则执行 `default` case。

## 空 channel

我们知道，channel 的默认值为 `nil`。因此，我们不能在 `nil` channel 上执行发送或接收操作。但是在这种情况下，当 select 语句中使用 `nil` channel 时，它将抛出以下错误之一或两个错误。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_s-LSV4rbD4wxC_fSpImD4w.png)

<https://play.golang.org/p/uhraFubcF4S>

从上面的结果中，我们可以看到 `select( 无 case)` 意味着 `select` 语句实际上是空的，**因为带有 `nil` channel 的 case 被忽略了**。但是由于空 `select{}` 语句阻塞了主 Goroutine，并且 `service` Goroutine 在它的位置被调度，所以在 `nil` 通道上的通道操作将抛出 `chan send (nil chan)` 错误。为了避免这种情况，我们使用 `default` case。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_-pkHV1XquXTZ9ILHzm8tSw.png)

<https://play.golang.org/p/upLsz52_CrE>

上述程序不仅忽略 `case` 块，而且立即执行 `default` case。因此调度器没有时间来调度 `service` Goroutine。但这确实是一个糟糕的设计。你应该检查通道的 `nil` 值。

## 添加超时

上面的程序不是很有用，因为只执行 `default` case。但有时，我们希望任何可用的服务都应该在适当的时间响应，如果它没有响应，那么就应该执行 `default` case。这可以通过使用在定义时间后解除阻塞的 channel 操作来完成。该 channel 操作由 `time` 包的 `after` 函数提供。我们来看一个例子。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_QGwXinOA4vVyTGHUL4BaJA.png)

<https://play.golang.org/p/mda2t2IQK__X>

以上程序在 2 秒后产生以下结果

```
main() started 0s
No response received 2.0010958s
main() stopped 2.0010958s
```

在上面的程序中，`<-time.After(2 * time. second)` 在 2 秒后解除阻塞，返回它被解除阻塞的时间，但是在这里，我们对它的返回值不感兴趣。因为它也像一个 Goroutine，我们有 3 个 Goroutine，这个首先从其中接触阻塞。因此，执行与 Goroutine 操作相对应的 case。

这是很有用的，因为您不希望等待来自可用服务的响应太久，而用户必须等待很长时间才能从服务中获得任何东西。如果加上 `10 *time.second`。第二，在上面的例子中，将打印 `service1` 的响应，我想现在已经很明显了。

## 空 select

与 `for{}` 空循环一样，空 `select{}` 语法也是有效的，但有一个问题。正如我们所知，`select` 语句被阻塞，直到其中一个 case 解除阻塞，而且由于没有 `case` 语句可用来解除阻塞，main Goroutine 将永远阻塞，从而导致死锁。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_Ukk-gzgkYjJTjUUoKCGVHg.png)

<https://play.golang.org/p/-pBd-BLMFOu>

在上面的程序中，我们知道 `select` 会阻塞主 Goroutine，调度器会调度另一个可用的 Goroutine，即 `service`。但是在那之后，它会挂起，调度器不得不调度另一个可用的 Goroutine，但是由于主协程被阻塞，没有其他 Goroutine 可用，导致死锁。

```
main() started
Hello from service!
fatal error: all Goroutines are asleep - deadlock!
goroutine 1 [select (no cases)]:
main.main()
        program.Go:16 +0xba
exit status 2
```

## WaitGroup

让我们设想这样一种情况 : 您需要知道是否所有的 Goroutines 都完成了它们的工作。这与选择只需要一个条件为 `true` 的地方有些相反，但是在这里需要**所有条件为 `true` 才能解锁主 Goroutine**。这里条件是 channel 操作成功。

**WaitGroup** 是一个具有计数器值的结构，它跟踪生成了多少个 Goroutines 以及完成了多少工作。当这个计数器达到 0 时，表示所有的 Goroutines 都完成了它们的工作。

让我们看一个例子，看看语法。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_QimR18tMSHg9RpxSOwpaHw.png)

<https://play.golang.org/p/8qrAD9ceOfJ>

在上面的程序中，我们创建了 `sync.WaitGroup` 类型的空结构 (*带有零值字段*) wg。WaitGroup struct 有未导出的字段，如 `noCopy`、`state1` 和 `sema`，其内部实现我们不需要知道。这个结构有三个方法，即 `Add`、`Wait` 和 `Done`。

`Add` 方法需要一个 `int` 类型的参数，这是 WaitGroup 计数器的增量。`Counter` 只是一个默认值为 0 的整数。它包含了正在运行的 Goroutine 的数量。在创建 WaitGroup 时，它的计数器值为 0，我们可以使用 `Add` 方法通过传递 `delta` 作为参数来递增它。请记住，`counter` 不能自动理解 Goroutine 何时启动，因此我们需要手动增加它。

`wait` 方法用于从调用当前 Goroutine 的位置阻塞该 Goroutine。一旦计数器达到 0，goroutine 将解除阻塞。因此，我们需要一些东西来减少计数器。

`Done` 方法使计数器递减。它不接受任何参数，因此它只减 1。

在上面的程序中，创建 `wg` 后，我们运行 `for` 循环 3 次。在每个回合中，我们启动一个 Goroutine，并增加计数器 1。这意味着，现在我们有 3 个 Goroutine 等待执行，而 WaitGroup 计数器是 3。注意，我们在 Goroutine 中传递了指向 `wg` 的指针。这是因为在 Goroutine 中，一旦我们完成了 Goroutine 应该做的事情，我们需要调用 `Done` 方法来减少计数器。如果 `wg` 作为值传递，`wg` 在 `main` 中不会减少。这是很明显的。

执行 `for` 循环之后，我们仍然没有将控制权传递给其他 Goroutines。这是通过调用 `wg` 上的 `Wait` 方法来完成的，比如 `wg.Wait()`。这将阻塞主 Goroutine，直到计数器达到 0。一旦计数器达到 0，因为从 3 个 Goroutine，我们调用了 `wg` 上的 `Done` 方法 3 次，`main` Goroutine 将解除阻塞，并开始执行进一步的代码。

因此上面的程序产生下面的结果

```
main() started
Service called on instance 2
Service called on instance 3
Service called on instance 1
main() stopped
```

由于 Goroutines 的执行顺序可能会有所不同，因此上述结果可能对您有所不同。

> `Add` 方法接受类型为 `int`，这意味着 `delta` 也可以是负的。想要了解更多，请访问这里的[官方文档](https://golang.org/pkg/sync/#WaitGroup.Add)。

## 工作池

顾名思义，工作池是同时工作以执行任务的 Goroutines 的集合。在 WaitGroup 中，我们看到了一些 Goroutines 的集合，但他们没有具体的工作。一旦您在它们中放入 channel，它们就有一些工作要做，并成为工作池。

因此，工作池背后的概念是维护一个 `worker Goroutines` 池，它接收一些任务并返回结果。一旦他们都完成了他们的工作，我们收集结果。所有这些 Goroutine 都为个人目的使用相同的通道。

让我们看一个简单的例子，有两个 channel，即 `tasks` 和 `results`。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_ronJMvwZhMPC6ngKIFolVg.png)

<https://play.golang.org/p/IYiMV1I4lCj>

别担心，我会一步一步解释这里发生的事情。

- `sqrWorker` 是一个工作函数，它接受 `task` channel、`result` channel 和 `id`。这个 Goroutine 的工作是将从 `task` channel 接收到的数字的平方发送到 `result` channel。
- 在 main 函数中，我们创建了具有缓冲区容量大小为 10 的 `task` 和 `result` channel。因此，任何发送操作都是非阻塞的，直到缓冲区满为止。因此，设置大的缓冲区值并不是一个坏主意。
- 然后，我们使用上面两个 channel 和 id 参数生成多个 `sqrWorker` 实例作为 Goroutines，以获取关于哪个 Worker 正在执行任务的信息。
- 然后我们将 5 个任务传递给 `task` channel，这些 `task` channel 是非阻塞的。
- 因为我们已经完成了 `task` channel，所以关闭了它。这不是必须的，但是如果有一些 bug 进来，它会在将来节省很多时间。
- 然后使用 for 循环，经过 5 次迭代，我们从 `result` channel 提取数据。由于空缓冲区上的读操作是阻塞的，因此将从工作池调度一个 Goroutine。在 Goroutine 返回一些结果之前，main Goroutine 将被阻塞。
- 由于我们在 worker Goroutine 中模拟阻塞操作，因此调用调度器来调度另一个可用的 Goroutine，直到它可用为止。当 worker Goroutine 可用时，它将写入 `result` channel。由于在缓冲区满之前，对缓冲 channel 的写入是非阻塞的，所以在这里对 `result` channel  的写入是非阻塞的。此外，当当前工作线程 Goroutine 不可用时，将使用任务缓冲区中的值执行多个其他工作线程 Goroutine。在所有工作者 Goroutines 消耗任务之后，当 `task` channel  缓冲区为空时，范围循环结束。当 `task` channel 关闭时，它不会抛出死锁错误。
- 有时，所有的工作线程都可以睡眠，所以主线程会醒来并工作，直到 `result` channel 缓冲区再次清空。
- 所有的 worker Goroutine 死后，`main` Goroutine 将重新获得控制权，并从 `result` channel 打印剩余的结果，继续执行。

上面的例子虽然很长，但是很好地解释了多个 Goroutine 如何在同一个 channel 上提供内容并优雅地完成工作。当员工的工作遇到阻碍时，goroutine 功能强大。如果删除 `time.Sleep()` 调用，那么只有一个 Goroutine 将执行此任务，因为在 `for range` 循环完成并在 Goroutine 死亡之前，不会调度其他 Goroutine。

> 您可以得到不同的结果，就像在前面的例子中一样，这取决于您的系统有多快，因为如果所有的 worker Gorutine 都被阻塞了，即使是一微秒，main Goroutine 也会像解释的那样被唤醒。

现在，让我们使用同步 Goroutines 的 WaitGroup 概念。使用前面的 WaitGroup 示例，我们可以获得相同的结果，但更优雅。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_ZU8xl8VJ2ilHqJp5DYGbbw.png)

<https://play.golang.org/p/0rRfchn7sL1>

上面的结果看起来很整洁，因为在 main Goroutine 中的 `result` channel 上的读取操作是非阻塞的，因为 `result` channel 已经由 result 填充，而 main Goroutine 被 `wg.Wait()` 调用阻塞。使用 `waitGroup`，我们可以防止很多 ( 不必要的 ) 上下文切换 ( 调度 )，这里是 7，而前面的示例中是 9。**但这是有代价的，因为你必须等待所有的工作都完成。**

## Mutex

互斥是 Go 中最简单的概念之一。但是在我解释它之前，让我们先理解竞态条件是什么。goroutines 有独立的栈，因此它们之间不共享任何数据。但是在某些情况下，堆中的某些数据可能在多个 Goroutine 之间共享。在这种情况下，多个 Goroutine 试图在相同的内存位置操作数据，从而导致意想不到的结果。我将向您展示一个简单的示例。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_Thv6PPSWBS377wNOTxkcgA.png)

<https://play.golang.org/p/MQNepChxiEa>

在上面的程序中，我们生成了 1000 个 Goroutines，它增加了初始值为 `0` 的全局变量 `i` 的值。由于我们正在实现 WaitGroup，所以我们希望所有 1000 个 Goroutines 都将 `i` 的值逐个递增，结果 `i` 的最终值为 1000。当 main Goroutine 在 `wg.Wait()` 调用后再次执行时，我们将输出 `i`。

```
value of i after 1000 operations is 937
```

什么？为什么小于 1000 呢？看起来有些 Goroutine 没用。但实际上，我们的程序有一个竞态条件。让我们看看会发生什么。

`i = i + 1` 的计算有 3 个步骤

- （1）得到 i 的值
- （2）i 的增量值为 1
- （3）用新值更新 i 的值

让我们设想一个场景，在这些步骤之间安排了不同的 Goroutine。例如，让我们考虑 1000 个 Goroutines 中的两个 Goroutines，即 G1 和 G2。

当 `i` 为 `0` 时，G1 首先开始，运行前两个步骤，`i` 现在是 `1`。但是在 G1 更新第 3 步中的 `i` 值之前，会调度新的 Goroutine G2 并运行所有步骤。但是对于 G2，`i` 的值仍然是 `0`，因此在执行步骤 3 之后，`i` 将是 1。现在 G1 再次被安排完成步骤 3，并更新步骤 2 中 `i` 的值 1。在完美的世界里，goroutines 在完成所有的 3 个步骤后被调度，两个 Goroutines 的成功操作会产生 `i` 为 2 的值，但这里不是这样。因此，我们可以推测为什么我们的程序没有将 `i` 的值赋值为 `1000`。

到目前为止，我们了解到 Goroutines 是合作安排的。除非一个 Goroutine 块具有并发性课程中提到的条件之一，否则另一个 Goroutine 不会取代它。既然 `i = i + 1` 不是阻塞，为什么 Go 调度器计划另一个 Goroutine ？

您一定要在 [stackoverflow](https://stackoverflow.com/questions/37469995/goroutines-are-cooperatively-scheduled-does-that-mean-that-goroutines-that-don) 上查看这个答案。**在任何情况下，您都不应该依赖 Go 的调度算法并实现自己的逻辑来同步不同的 Goroutine。**

一种确保每次只有一个 Goroutine 完成上述 3 个步骤的方法是实现互斥锁。互斥 (*互斥*) 是编程中的一个概念，在这个概念中，一次只能有一个例程 ( 线程 ) 执行操作。这是通过一个获取值上的锁的例程来完成的，对它必须做的值做任何操作，然后释放锁。当值被锁定时，没有其他例程可以对其读写。

在 Go 中，互斥数据结构 ( map ) 是由 `sync` 包提供的。在 Go 中，在对可能导致竞态条件的值执行任何操作之前，我们使用 `mutex.Lock()` 方法获取一个锁，然后是操作代码。一旦我们完成了操作，在上面的程序 `i = i + 1` 中，我们使用 `mutex.unlock()` 方法来解锁它。当锁存在时，任何其他 Goroutine 试图读取或写入 `i` 的值时，该 Goroutine 将阻塞，直到从第一个 Goroutine 解锁操作为止。因此，只有 1 个 Goroutine 可以读取或写入 `i`  的值，从而避免了竞态条件。请记住，在锁定和解锁之间的操作中出现的任何变量在整个操作解锁之前都不能用于其他 Goroutines。

让我们用互斥锁修改前面的示例。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_lDHFDD1DKyI8KTC_4He4Iw.png)

<https://play.golang.org/p/xVFAX_0Uig8>

在上面的程序中，我们创建了一个互斥量 `m`，并将它的指针传递给所有派生的 Goroutines。在开始对 `i` 进行操作之前，我们使用 `m.lock()` 语法获得互斥对象 `m` 上的锁，然后在操作之后使用 `m.unlock()` 语法解锁它。上面的程序产生下面的结果。

```
value of i after 1000 operations is 1000
```

从上面的结果我们可以看到互斥帮助我们解决了竞态条件。但是第一条规则是避免 Goroutines 之间共享资源。

> 您可以在运行 `Go run -race program.Go` 这样的程序时，使用 `race` 参数在 Go 中测试竞态条件。请在[这里](https://blog.golang.org/race-detector) 阅读更多关于 race 检测器的信息。

## 并发模式

并发有很多方法可以使我们的日常编程更加容易。以下是一些概念和方法，我们可以使用它们使程序更快和更可靠。

### Generator ( 生产者 )

使用 channel，我们可以更好地实现生产者。如果生产者在计算上很昂贵，那么我们也可以同时生成数据。这样，程序就不必等待所有数据生成。例如，生成斐波那契数列。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_rFKYal44BruN5OQM3AISwA.png)

<https://play.golang.org/p/1_2MDeqQ3o5>

使用 `fib` 函数，我们得到了一个可以迭代和利用从它接收到的数据的 channel。而在 `fib` 函数内部，由于我们必须返回一个只接收 channel，我们正在创建一个缓冲 channel 并在最后返回它。此函数的返回值将将此双向 channel 转换为单向接收 channel。在使用匿名 Goroutine 时，我们使用 for 循环将斐波那契数推入这个 channel。一旦完成 for 循环，我们就会从 Goroutine 内部关闭它。在 `main` Goroutine 中，使用 `for range` 在 `fib` 函数调用，我们可以直接访问这个 channel。

### fan-in & fan-out （扇入和扇出）

扇入是一种多路复用策略，将多个 channel 的输入组合起来产生一个输出 channel。扇出是一种多路复用策略，其中单个 channel 被分成多个 channel。

```go
package main

import (
	"fmt"
	"sync"
)
// return channel for input numbers
func getInputChan() <-chan int {
// make return channel
	input := make(chan int, 100)

// sample numbers
	numbers := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

// run Goroutine
	go func() {
		for num := range numbers {
			input <- num
		}
// close channel once all numbers are sent to channel
		close(input)
	}()

	return input
}

// returns a channel which returns square of numbers
func getSquareChan(input <-chan int) <-chan int {
// make return channel
	output := make(chan int, 100)

// run Goroutine
	go func() {
		// push squares until input channel closes
		for num := range input {
			output <- num * num
		}

// close output channel once for loop finishesh
		close(output)
	}()

	return output
}

// returns a merged channel of `outputsChan` channels
// this produce fan-in channel
// this is veriadic function
func merge(outputsChan ...<-chan int) <-chan int {
// create a WaitGroup
	var wg sync.WaitGroup
// make return channel
	merged := make(chan int, 100)

// increase counter to number of channels `len(outputsChan)`
// as we will spawn number of Goroutines equal to number of channels received to merge
wg.Add(len(outputsChan))

// function that accept a channel (which sends square numbers)
// to push numbers to merged channel
	output := func(sc <-chan int) {
// run until channel (square numbers sender) closes
		for sqr := range sc {
			merged <- sqr
		}
// once channel (square numbers sender) closes,
// call `Done` on `WaitGroup` to decrement counter
		wg.Done()
	}

// run above `output` function as groutines, `n` number of times
// where n is equal to number of channels received as argument the function
// here we are using `for range` loop on `outputsChan` hence no need to manually tell `n`
	for _, optChan := range outputsChan {
		go output(optChan)
	}

// run Goroutine to close merged channel once done
	go func() {
		// wait until WaitGroup finishesh
		wg.Wait()
		close(merged)
	}()

	return merged
}

func main() {
// step 1: get input numbers channel
// by calling `getInputChan` function, it runs a Goroutine which sends number to returned channel
	chanInputNums := getInputChan()

// step 2: `fan-out` square operations to multiple Goroutines
// this can be done by calling `getSquareChan` function multiple times where individual function call returns a channel which sends square of numbers provided by `chanInputNums` channel
// `getSquareChan` function runs Goroutines internally where squaring operation is ran concurrently
	chanOptSqr1 := getSquareChan(chanInputNums)
	chanOptSqr2 := getSquareChan(chanInputNums)

// step 3: fan-in (combine) `chanOptSqr1` and `chanOptSqr2` output to merged channel
// this is achieved by calling `merge` function which takes multiple channels as arguments
// and using `WaitGroup` and multiple Goroutines to receive square number, we can send square numbers
// to `merged` channel and close it
	chanMergedSqr := merge(chanOptSqr1, chanOptSqr2)

// step 4: let's sum all the squares from 0 to 9 which should be about `285`
// this is done by using `for range` loop on `chanMergedSqr`
	sqrSum := 0

// run until `chanMergedSqr` or merged channel closes
// that happens in `merge` function when all Goroutines pushing to merged channel finishes
// check line no. 86 and 87
	for num := range chanMergedSqr {
		sqrSum += num
	}

// step 5: print sum when above `for loop` is done executing which is after `chanMergedSqr` channel closes
	fmt.Println("Sum of squares between 0-9 is", sqrSum)
}
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-channel/1_KmwgYxh3tCxU7nwKuopn0Q.png)

<https://play.golang.org/p/hATZmb6P1-u>

我不打算解释上面的程序是如何工作的，因为我已经在程序中添加了注释来解释了这一点。以上程序产生以下结果

```
Sum of squares between 0-9 is 285
```

------

via: https://medium.com/rungo/anatomy-of-channels-in-go-concurrency-in-go-1ec336086adb

作者：[Uday Hiwarale](https://medium.com/@thatisuday)
译者：[wumansgy](https://github.com/wumansgy)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
