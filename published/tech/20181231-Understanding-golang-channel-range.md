首发于：https://studygolang.com/articles/18800

# Golang 中迭代读取 channel

欢迎再次来到我的 Go 语言私人教学课堂，今天我们的主题是，我曾经很难理解的 ( 还好现在已经理解了 )：在 Go 程中迭代读取 `channels`。

在开始之前，让我们先回忆一下。我们都知道，一个 Go 程的存活周期是建立在 main 进程之上的，举个例子：

```go
package main

import "fmt"

func main() {
	go func() {
		fmt.Println("hello there")
	}()
}
```

[点我运行](https://play.golang.org/p/cbczlMV4_0p)

只有极低的几率你才有可能看到 `fmt.Println` 打印的信息，因为有很大几率 `main()` 函数会在打印执行前就结束。
我们同样知道一个[规范的方法](https://nathanleclaire.com/blog/2014/02/15/how-to-wait-for-all-goroutines-to-finish-executing-before-continuing/) 去控制 Go 程的运行，那就是使用 `Waitgroup`：

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("hello there")
	}()
}
```

[点我运行](https://play.golang.org/p/7FqFV28U9Lc)

好的，现在让我们进入今天的正题，我们想在 Go 程中通过管道发送信息，很简单吧？

```go
package main

import "fmt"

func main() {
	c := make(chan string)

	go func() {
		c <- "hello there"
	}()

	msg := <- c
	fmt.Println(msg)
}
```

[点我运行](https://play.golang.org/p/LGwr6Po2sHn)

等下 ( 诶。。。) 这个方法没有使用 `WaitGroup` 居然就起到效果了？？当一个 `channel` 进入阻塞状态的时候 , 意味着它正在等待发送 / 接受数据。所以利用这一点，`channel` 可以用来实现 `goroutine` 之间的同步。

现在让我们试着向 `chanenl` 中发送一组 `strings` 数据，并使用 `range` 来接受数据， 以便能够迭代读取 `channel` 中的数据：

```go
package main

import "fmt"

func main() {
	c := make(chan string)

	go func() {
		for i := 0; i < 10; i++ {
			c <- "hello there"
		}
	}()

	for msg := range c {
		fmt.Println(msg)
	}
}
```

[点我运行](https://play.golang.org/p/tNPjm1hHOOQ)

结果比较因缺思厅：

```
hello there
hello there
hello there
hello there
hello there
hello there
hello there
hello there
hello there
hello there
fatal error: all Goroutines are asleep - deadlock!

goroutine 1 [chan receive]:
main.main()
	/tmp/sandbox697910326/main.go:14 +0x120
```

为了理解程序执行过程中发生了什么，我们必须知道使用 `range` 迭代读取 `channel` 时，当 `channel` 中没有数据时，这一读取行为是不会停止的除非 `channel` 被关闭时，好的，知道了这个，那就让我们来试试关闭它吧。

```go
package main

import "fmt"

func main() {
	c := make(chan string)

	go func() {
		for i := 0; i < 10; i++ {
			c <- "hello there"
		}
		close(c)
	}()
	for msg := range c {
		fmt.Println(msg)
	}
}
```

[点我运行](https://play.golang.org/p/Or2MGH9YeIu)

加上 `close` 后程序变得好多了。

接着让我们来尝试一些更复杂的，我们使用 `for` 循环来启动 `goroutines`：

```go
package main

import "fmt"

func main() {
	c := make(chan string)

	for i := 0; i < 10; i++ {
		go func() {
			c <- "hello there"
		}()
		close(c)
	}
	for msg := range c {
		fmt.Println(msg)
	}
}
```

[点我运行](https://play.golang.org/p/93zpxY_xRhO)

结果：

```
panic: close of closed channel

goroutine 1 [running]:
main.main()
	/tmp/sandbox536323156/main.go:12 +0xa0
```

我们提交这样一个错误：首先从接收端关闭了一个 `channel`，接着我们再次发送并关闭已经被关闭的 `channel`，这就导致了 `panic` 的出现。

就是这样，那么就真的没办法循环读取这些值了吗？实际上是有的，这需要使用另一个 `goroutine`，代码如下。

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	defer wg.Wait()

	c := make(chan string)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c <- "hello there"
		}()
	}

	go func() {
		for msg := range c {
			fmt.Println(msg)
		}
	}()
}
```

[点我运行](https://play.golang.org/p/1hhWLHg6So2)

好了，现在可以从很多个 `goroutines` 中接收数据了！在这个例子中，我们等待 `for / Go` 迭代的完成，接收者 `goroutine` 通过管道使得接收同步，当循环发送数据结束时，主进程也就跟着结束了。

我写这篇文章也是为了做一个自我梳理，让自己的思路更清晰一些，所以如果你发现任何不正确不到位的讲解说明，请不要犹豫，赶快在下面评论吧。

---

via: https://imil.net/blog/2018/12/31/Understanding-golang-channel-range/

作者：[Imil](https://github.com/iMilnb)
译者：[CNbluer](https://github.com/CNbluer)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
