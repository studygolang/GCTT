首发于：https://studygolang.com/articles/17944

# 深度剖析 Go 中的 Go 协程 (goroutines) -- Go 的并发

> Go 协程 (goroutine) 是指在后台中运行的轻量级执行线程，go 协程是 Go 中实现并发的关键组成部分。

在上次的课程中，我们学习了 Go 的并发模型。由于 Go 协程相对于传统操作系统中的线程 (thread) 是非常轻量级的，因此对于一个典型的 Go 应用来说，有数以千计的 Go 协程并发运行的情形是十分常见的。并发可以显著地提升应用的运行速度，并且可以帮助我们编写关注点分离（Separation of concerns，Soc）的代码。

## 什么是 Go 协程？

我们也许在理论上已经知晓 Go 协程是如何工作的，但是在代码层级上，go 协程何许物也？其实，go 协程看起来只是一个与其他众 Go 协程并发运行的一个简单函数或者方法，但是我们并不能想当然地从函数或者方法中的定义来确定一个 Go 协程，go 协程的确定还是要取决于我们如何去调用。

Go 中提供了一个关键字 `go` 来让我们创建一个 Go 协程，当我们在函数或方法的调用之前添加一个关键字 `go`，这样我们就开启了一个 Go 协程，该函数或者方法就会在这个 Go 协程中运行。

举个简单的栗子：

![example-1](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-goroutine/1_gF5yHIbu2E0pqAceDMGbdw.png)

<center><a href="https://play.golang.org/p/pIGsToIA2hL">https://play.golang.org/p/pIGsToIA2hL</a></center>

在上面的代码中，我们定义了一个可以在控制台输出 `Hello World` 字符串的 `printHello` 的函数，在 `main` 函数中，我们就像平时那样调用 `printHello` 函数，最终也是理所当然地获得了期望的结果。

下面，让我们尝试从同一个函数创建 Go 协程：

![example-2](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-goroutine/1_7viT_DLTjVJ-wEU0IPS9Wg.png)

<center><a href="https://play.golang.org/p/LWXAgDpTcJP">https://play.golang.org/p/LWXAgDpTcJP</a></center>

根据 Go 协程的语法，我们在函数调用的前面增加了一个 `go` 关键字，之后程序运行正常，输出了以下的结果：

```
main execution started
main execution stopped
```

奇怪的是，`Hello World` 并没有如同我们预料的那样输出，这期间究竟发生了什么？

go 协程总是在后台运行，当一个 Go 协程执行的时候（在这个例子中是 `go printHello()`）, Go 并不会像在之前的那个例子中在执行 `printHello` 中的功能时阻塞 main 函数中剩下语句的执行，而是直接忽略了 Go 协程的返回并继续执行 main 函数剩下的语句。**即便如此，我们为什么没法看到函数的输出呢？**

在默认情况下，每个独立的 Go 应用运行时就创建了一个 Go 协程，其 `main` 函数就在这个 Go 协程中运行，这个 Go 协程就被称为 `go 主协程（main Goroutine，下面简称主协程）`。在上面的例子中，`主协程` 中又产生了一个 `printHello` 这个函数的 Go 协程，我们暂且叫它 `printHello 协程` 吧，因而我们在执行上面的程序的时候，就会存在两个 Go 协程（`main` 和 `printHello`）同时运行。正如同以前的程序那样，go 协程们会进行协同调度。因此，当 `主协程` 运行的时候，Go 调度器在 `主协程` 执行完之前并不会将控制权移交给 `printHello 协程`。不幸的是，一旦 `主协程` 执行完毕，整个程序会立即终止，调度器再也没有时间留给 `printHello 协程` 去运行了。

但正如我们从其他课程所知，通过阻塞条件，我们可以手动将控制权转移给其他的 Go 协程 , 也可以说是告诉调度器让它去调度其他可用空闲的 Go 协程。让我们调用 `time.Sleep()` 函数去实现它吧。

![example-3](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-goroutine/1_Vd4kxUcz1_CKC_hrY8_r-A.png)

<center><a href="https://play.golang.org/p/ujQKjpALlRJ">https://play.golang.org/p/ujQKjpALlRJ</a></center>

如上图所示，我们修改了程序，程序在 main 函数的最后一条语句之前调用了 `time.Sleep(10 * time.Millisecond)`，使得 `主协程` 在执行最后一条指令之前调度器就将控制权转移给了 `printhello 协程`。在这个例子中，我们通过调用 `time.Sleep(10 * time.Millisecond)` 强行让 `主协程` 休眠 10ms 并且在在这个 10ms 内不会再被调度器重新调度运行。

一旦 `printHello 协程` 执行，它就会向控制台打印‘ Hello World ！’，然后该 Go 协程（`printHello 协程`）就会随之终止，接下来 `主协程` 就会被重新调度（因为 main Go 协程已经睡够 10ms 了），并执行最后一条语句。因此，运行上面的程序就会得到以下的输出 :

```shell
main execution started
Hello World!
main execution stopped
```

下面我稍微修改一下例子，我在 `printHello` 函数的输出语句之前添加了一条 `time.Sleep(time.Millisecond)`。我们已经知道了如果我们在函数中调用了休眠（sleep）函数，这个函数就会告诉 Go 调度器去调度其他可被调度的 Go 协程。在上一课中提到，只有非休眠（`non-sleeping`）的 Go 协程才会被认为是可被调度的，所以主协程在这休眠的 10ms 内是不会被再次调度的。因此 `主协程` 先打印出“ main execution started ” 接着就创建了一个 **printHello** 协程，*需要注意此时的 `主协程` 还是非休眠状态的*，在这之后主协程就要调用休眠函数去睡 10ms，并且把这个控制权让出来给**printHello** 协程。**printHello** 协程会先休眠 1ms 告诉调度器看看有没有其他可调度的 Go 协程，在这个例子里显然没有其他可调度的 Go 协程了，所以在**printHello**协程结束了这 1ms 的休眠户就会被调度器调度，接着就输出了“ Hello World ”字符串，之后这个 Go 协程运行结束。之后，主协程会在之后的几毫秒被唤醒，紧接着就会输出“ main execution stopped ”并且结束整个程序。

![example-4](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-goroutine/1_3lQnGP4JRuzDH2DEvFE2sw.png)

<center><a href="https://play.golang.org/p/rWvzS8UeqD6">https://play.golang.org/p/rWvzS8UeqD6</a></center>

上面的程序依旧和之前的例子一样，输出以下相同的结果：

```shell
main execution started
Hello World!
main execution stopped
```

要是，我把这个**printHello** 协程中的休眠 1 毫秒改成休眠 15 毫秒，这个结果又是如何呢？

![example-5](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-goroutine/1_m6IyoYmXTb4mocn_0-OqrQ.png)
<center><a href="https://play.golang.org/p/Pc2nP2BtRiP">https://play.golang.org/p/Pc2nP2BtRiP</a></center>

在这个例子中，与其他的例子最大的区别就是 **printHello** 协程比主协程的休眠时间还要长，很明显，主协程要比 printHello 协程唤醒要早，这样的结果就是主协程即使唤醒后执行完所有的语句，printHello 协程还是在休眠状态。之前提到过，主协程比较特殊，如果主协程执行结束后整个程序就要退出，所以 printHello 协程得不到机会去执行下面的输出的语句了，所以以上的程序的数据结果如下：

```shell
main execution started
main execution stopped
```

## 使用多 Go 协程

就像之前我所提到过的，你可以随心所欲地创建多个 Go 协程。下面让我们定义两个简单的函数，一个是用于顺序打印某个字符串中的每个字符，另一个是顺序打印出某个整数切片中的每个数字。

![example2-1](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-goroutine/1_C1TtQM5vFjNiiR99GNrMzA.png)
<center><a href="https://play.golang.org/p/SJano_g1wTV">https://play.golang.org/p/SJano_g1wTV</a></center>

在上图中的程序中，我们连续地创建了两个 Go 协程，程序输出的结果如下：

```shell
main execution started
H e l l o 1 2 3 4 5
main execution stopped
```

上面的结果证实了 Go 协程是以合作式调度来运作的。下面我们在每个函数中的输出语句的下面添加一行 `time.Sleep`，让函数在输出每个字符或数字后休息一段时间，好让调度器调度其他可用的 Go 协程。

![example-2-2](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-goroutine/1_LbE_Ls0r-bWZIX-lr5jc9g.png)
<center><a href="https://play.golang.org/p/lrSIEdNxSaH">https://play.golang.org/p/lrSIEdNxSaH</a></center>

在上面的程序中，我又修改了一下输出语句使得我们可以看到每个字符或数字的输出时刻。理论上主协程会休眠 200ms，因此其他 Go 协程要赶在主协程唤醒之前做完自己的工作，因为主协程唤醒之后就会导致程序退出。`getChars` 协程每打印一个字符就会休眠 10ms，之后控制权就会传给 `getDigits` 协程，`getDigits` 协程每打印一个数字后就休眠 30ms，若 `getChars` 协程唤醒，则会把控制权传回 `getChars` 协程，如此往复。在代码中可以看到，`getChars` 协程会在其他协程休眠的时候多次进行打印字符以及休眠操作，所以我们预计可以看到输出的字符比数字更具有连续性。

我们在 Windows 上运行上面的程序，得到了以下的结果：

```shell
main execution started at time 0s
H at time 1.0012ms                         <-|
1 at time 1.0012ms                           | almost at the same time
e at time 11.0283ms                        <-|
l at time 21.0289ms                          | ~10ms apart
l at time 31.0416ms
2 at time 31.0416ms
o at time 42.0336ms
3 at time 61.0461ms                        <-|
4 at time 91.0647ms                          |
5 at time 121.0888ms                         | ~30ms apart
main execution stopped at time 200.3137ms    | exiting after 200ms
```

通过以上输出结果可以证明我们之前对输出的讨论。对于这个结果，我们可以通过下面的的程序运行图来解释。需要注意的是，我们在图中定义一个输出语句大约会花费 1ms 的 CPU 时间，而这个时间相对于 200ms 来说是可以忽略不计的。

![example-2-3](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-goroutine/0_4_Z0LRvi_DJR1JEr.jpg)

现在我们已经知道了如何去创建 Go 协程以及去如何去使用它。但是使用 `time.Sleep` 只是一个让我们获取理想结果的一个小技巧。在实际生产环境中，我们无法知晓一个 Go 协程到底需要执行多长的时间，因而在 main 函数里面添加一个 `time.Sleep` 并不是一个解决问题的方法。我们希望 Go 协程在执行完毕后告知主协程运行的结果。在目前阶段，我们还不知道如何向其他 Go 协程传递以及获取数据，简而言之，就是与其他 Go 协程进行通信。这就是 channels 引入的原因。我们会在下一次课中讨论这个东西。

## 匿名 Go 协程

如果一个匿名函数可以退出，那么匿名 Go 协程也同样可以退出。请参照[`functions`](https://medium.com/rungo/the-anatomy-of-functions-in-go-de56c050fe11) 课程中的 `即时调用函数（Immedietly invoked function）` 来理解本节。让我们修改一下之前 `printHello` 协程的例子：

![example-3-1](https://raw.githubusercontent.com/studygolang/gctt-images/master/anatomy-of-goroutine/1_BB4uDfV7ooZ0SwQpTJPTsQ.png)

结果非常明显，因为我们定义了匿名函数，并在同一语句中作为 Go 协程执行。

*需要注意的是，所有的 Go 协程都是匿名的，因为我们从[`并发（concurrency`](https://medium.com/rungo/achieving-concurrency-in-go-3f84cbf870ca) 一课中学到，go 协程是不存在标识符的，在这里所谓的匿名 Go 协程只是通过匿名函数来创建的 Go 协程罢了*

---

via: https://medium.com/rungo/anatomy-of-goroutines-in-go-concurrency-in-go-a4cb9272ff88

作者：[Uday Hiwarale](https://medium.com/@thatisuday)
译者：[Hafrans](https://github.com/hafrans)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
