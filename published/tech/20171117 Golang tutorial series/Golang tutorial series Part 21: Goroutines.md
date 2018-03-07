已发布：https://studygolang.com/articles/12342

# 第 21 篇：Go 协程

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 21 篇。  

在前面的教程里，我们探讨了并发，以及并发与并行的区别。本教程则会介绍在 Go 语言里，如何使用 Go 协程（Goroutine）来实现并发。

## Go 协程是什么？

Go 协程是与其他函数或方法一起并发运行的函数或方法。Go 协程可以看作是轻量级线程。与线程相比，创建一个 Go 协程的成本很小。因此在 Go 应用中，常常会看到有数以千计的 Go 协程并发地运行。  

## Go 协程相比于线程的优势

- 相比线程而言，Go 协程的成本极低。堆栈大小只有若干 kb，并且可以根据应用的需求进行增减。而线程必须指定堆栈的大小，其堆栈是固定不变的。
- Go 协程会复用（Multiplex）数量更少的 OS 线程。即使程序有数以千计的 Go 协程，也可能只有一个线程。如果该线程中的某一 Go 协程发生了阻塞（比如说等待用户输入），那么系统会再创建一个 OS 线程，并把其余 Go 协程都移动到这个新的 OS 线程。所有这一切都在运行时进行，作为程序员，我们没有直接面临这些复杂的细节，而是有一个简洁的 API 来处理并发。  
- Go 协程使用信道（Channel）来进行通信。信道用于防止多个协程访问共享内存时发生竞态条件（Race Condition）。信道可以看作是 Go 协程之间通信的管道。我们会在下一教程详细讨论信道。
 
## 如何启动一个 Go 协程？

调用函数或者方法时，在前面加上关键字 `go`，可以让一个新的 Go 协程并发地运行。

让我们创建一个 Go 协程吧。

```go
package main

import (
	"fmt"
)

func hello() {
	fmt.Println("Hello world goroutine")
}
func main() {
	go hello()
	fmt.Println("main function")
}
```
[在线运行程序](https://play.golang.org/p/zC78_fc1Hn)

在第 11 行，`go hello()` 启动了一个新的 Go 协程。现在 `hello()` 函数与 `main()` 函数会并发地执行。主函数会运行在一个特有的 Go 协程上，它称为 Go 主协程（Main Goroutine）。

**运行一下程序，你会很惊讶！**

该程序只会输出文本 `main function`。我们启动的 Go 协程究竟出现了什么问题？要理解这一切，我们需要理解两个 Go 协程的主要性质。  

- **启动一个新的协程时，协程的调用会立即返回。与函数不同，程序控制不会去等待 Go 协程执行完毕。在调用 Go 协程之后，程序控制会立即返回到代码的下一行，忽略该协程的任何返回值。**  
- **如果希望运行其他 Go 协程，Go 主协程必须继续运行着。如果 Go 主协程终止，则程序终止，于是其他 Go 协程也不会继续运行。**  

现在你应该能够理解，为何我们的 Go 协程没有运行了吧。在第 11 行调用了 `go hello()` 之后，程序控制没有等待 `hello` 协程结束，立即返回到了代码下一行，打印 `main function`。接着由于没有其他可执行的代码，Go 主协程终止，于是 `hello` 协程就没有机会运行了。

我们现在修复这个问题。

```go
package main

import (  
	"fmt"
	"time"
)

func hello() {  
	fmt.Println("Hello world goroutine")
}
func main() {  
	go hello()
	time.Sleep(1 * time.Second)
	fmt.Println("main function")
}
```
[在线运行程序](https://play.golang.org/p/U9ZZuSql8-)  

在上面程序的第 13 行，我们调用了 time 包里的函数 [`Sleep`](https://golang.org/pkg/time/#Sleep)，该函数会休眠执行它的 Go 协程。在这里，我们使 Go 主协程休眠了 1 秒。因此在主协程终止之前，调用 `go hello()` 就有足够的时间来执行了。该程序首先打印 `Hello world goroutine`，等待 1 秒钟之后，接着打印 `main function`。  

在 Go 主协程中使用休眠，以便等待其他协程执行完毕，这种方法只是用于理解 Go 协程如何工作的技巧。信道可用于在其他协程结束执行之前，阻塞 Go 主协程。我们会在下一教程中讨论信道。  

## 启动多个 Go 协程

为了更好地理解 Go 协程，我们再编写一个程序，启动多个 Go 协程。  

```go
package main

import (  
	"fmt"
	"time"
)

func numbers() {  
	for i := 1; i <= 5; i++ {
		time.Sleep(250 * time.Millisecond)
		fmt.Printf("%d ", i)
	}
}
func alphabets() {  
	for i := 'a'; i <= 'e'; i++ {
		time.Sleep(400 * time.Millisecond)
		fmt.Printf("%c ", i)
	}
}
func main() {  
	go numbers()
	go alphabets()
	time.Sleep(3000 * time.Millisecond)
	fmt.Println("main terminated")
}
```
[在线运行程序](https://play.golang.org/p/U9ZZuSql8-)  

在上面程序中的第 21 行和第 22 行，启动了两个 Go 协程。现在，这两个协程并发地运行。`numbers` 协程首先休眠 250 微秒，接着打印 `1`，然后再次休眠，打印 `2`，依此类推，一直到打印 `5` 结束。`alphabete` 协程同样打印从 `a` 到 `e` 的字母，并且每次有 400 微秒的休眠时间。 Go 主协程启动了 `numbers` 和 `alphabete` 两个 Go 协程，休眠了 3000 微秒后终止程序。  

该程序会输出：  

```
1 a 2 3 b 4 c 5 d e main terminated  
```

程序的运作如下图所示。为了更好地观看图片，请在新标签页中打开。  

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-series/Goroutines-explained.png)

第一张蓝色的图表示 `numbers` 协程，第二张褐红色的图表示 `alphabets` 协程，第三张绿色的图表示 Go 主协程，而最后一张黑色的图把以上三种协程合并了，表明程序是如何运行的。在每个方框顶部，诸如 `0 ms` 和 `250 ms` 这样的字符串表示时间（以微秒为单位）。在每个方框的底部，`1`、`2`、`3` 等表示输出。蓝色方框表示：`250 ms` 打印出 `1`，`500 ms` 打印出 `2`，依此类推。最后黑色方框的底部的值会是 `1 a 2 3 b 4 c 5 d e main terminated`，这同样也是整个程序的输出。以上图片非常直观，你可以用它来理解程序是如何运作的。  

Go 协程的介绍到此结束。祝你愉快。

**下一教程 - [信道](#)**

---

via: https://golangbot.com/goroutines/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
