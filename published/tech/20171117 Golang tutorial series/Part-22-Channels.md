# 第 22 篇：信道（channel）

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 22 篇。  

在[上一教程](https://studygolang.com/articles/12342)里，我们探讨了如何使用 Go 协程（Goroutine）来实现并发。我们接着在本教程里学习信道（Channel），学习如何通过信道来实现 Go 协程间的通信。  

## 什么是信道？

信道可以想像成 Go 协程之间通信的管道。如同管道中的水会从一端流到另一端，通过使用信道，数据也可以从一端发送，在另一端接收。  

## 信道的声明

所有信道都关联了一个类型。信道只能运输这种类型的数据，而运输其他类型的数据都是非法的。  

`chan T` 表示 `T` 类型的信道。  

信道的零值为 `nil`。信道的零值没有什么用，应该像对 map 和切片所做的那样，用 `make` 来定义信道。  

下面编写代码，声明一个信道。  

```go
package main

import "fmt"

func main() {  
	var a chan int
	if a == nil {
		fmt.Println("channel a is nil, going to define it")
		a = make(chan int)
		fmt.Printf("Type of a is %T", a)
	}
}
```
[在线运行程序](https://play.golang.org/p/QDtf6mvymD)  

由于信道的零值为 `nil`，在第 6 行，信道 `a` 的值就是 `nil`。于是，程序执行了 if 语句内的语句，定义了信道 `a`。程序中 `a` 是一个 int 类型的信道。该程序会输出：  

```
channel a is nil, going to define it  
Type of a is chan int  
```

简短声明通常也是一种定义信道的简洁有效的方法。  

```go
a := make(chan int) 
```

这一行代码同样定义了一个 int 类型的信道 `a`。  

## 通过信道进行发送和接收

如下所示，该语法通过信道发送和接收数据。  

```go
data := <- a // 读取信道 a  
a <- data // 写入信道 a  
```

信道旁的箭头方向指定了是发送数据还是接收数据。  

在第一行，箭头对于 `a` 来说是向外指的，因此我们读取了信道 `a` 的值，并把该值存储到变量 `data`。  

在第二行，箭头指向了 `a`，因此我们在把数据写入信道 `a`。  

## 发送与接收默认是阻塞的

发送与接收默认是阻塞的。这是什么意思？当把数据发送到信道时，程序控制会在发送数据的语句处发生阻塞，直到有其它 Go 协程从信道读取到数据，才会解除阻塞。与此类似，当读取信道的数据时，如果没有其它的协程把数据写入到这个信道，那么读取过程就会一直阻塞着。  

信道的这种特性能够帮助 Go 协程之间进行高效的通信，不需要用到其他编程语言常见的显式锁或条件变量。  

## 信道的代码示例

理论已经够了:)。接下来写点代码，看看协程之间通过信道是怎么通信的吧。  

我们其实可以重写上章学习 [Go 协程](https://studygolang.com/articles/12342) 时写的程序，现在我们在这里用上信道。  

首先引用前面教程里的程序。  

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

这是上一篇的代码。我们使用到了休眠，使 Go 主协程等待 hello 协程结束。如果你看不懂，建议你阅读上一教程 [Go 协程](https://studygolang.com/articles/12342)。  

我们接下来使用信道来重写上面代码。  

```go
package main

import (  
	"fmt"
)

func hello(done chan bool) {  
	fmt.Println("Hello world goroutine")
	done <- true
}
func main() {  
	done := make(chan bool)
	go hello(done)
	<-done
	fmt.Println("main function")
}
```
[在线运行程序](https://play.golang.org/p/I8goKv6ZMF)  

在上述程序里，我们在第 12 行创建了一个 bool 类型的信道 `done`，并把 `done` 作为参数传递给了 `hello` 协程。在第 14 行，我们通过信道 `done` 接收数据。这一行代码发生了阻塞，除非有协程向 `done` 写入数据，否则程序不会跳到下一行代码。于是，这就不需要用以前的 `time.Sleep` 来阻止 Go 主协程退出了。  

`<-done` 这行代码通过协程（译注：原文笔误，信道）`done` 接收数据，但并没有使用数据或者把数据存储到变量中。这完全是合法的。  

现在我们的 Go 主协程发生了阻塞，等待信道 `done` 发送的数据。该信道作为参数传递给了协程 `hello`，`hello` 打印出 `Hello world goroutine`，接下来向 `done` 写入数据。当完成写入时，Go 主协程会通过信道 `done` 接收数据，于是它解除阻塞状态，打印出文本 `main function`。  

该程序输出如下：  

```
Hello world goroutine  
main function  
```

我们稍微修改一下程序，在 `hello` 协程里加入休眠函数，以便更好地理解阻塞的概念。  

```go
package main

import (  
	"fmt"
	"time"
)

func hello(done chan bool) {  
	fmt.Println("hello go routine is going to sleep")
	time.Sleep(4 * time.Second)
	fmt.Println("hello go routine awake and going to write to done")
	done <- true
}
func main() {  
	done := make(chan bool)
	fmt.Println("Main going to call hello go goroutine")
	go hello(done)
	<-done
	fmt.Println("Main received data")
}
```
[在线运行程序](https://play.golang.org/p/EejiO-yjUQ)  

在上面程序里，我们向 `hello` 函数里添加了 4 秒的休眠（第 10 行）。  

程序首先会打印 `Main going to call hello go goroutine`。接着会开启 `hello` 协程，打印 `hello go routine is going to sleep`。打印完之后，`hello` 协程会休眠 4 秒钟，而在这期间，主协程会在 `<-done` 这一行发生阻塞，等待来自信道 `done` 的数据。4 秒钟之后，打印 `hello go routine awake and going to write to done`，接着再打印 `Main received data`。  

## 信道的另一个示例

我们再编写一个程序来更好地理解信道。该程序会计算一个数中每一位的平方和与立方和，然后把平方和与立方和相加并打印出来。

例如，如果输出是 123，该程序会如下计算输出：  

```
squares = (1 * 1) + (2 * 2) + (3 * 3) 
cubes = (1 * 1 * 1) + (2 * 2 * 2) + (3 * 3 * 3) 
output = squares + cubes = 49
```

我们会这样去构建程序：在一个单独的 Go 协程计算平方和，而在另一个协程计算立方和，最后在 Go 主协程把平方和与立方和相加。  

```go
package main

import (  
	"fmt"
)

func calcSquares(number int, squareop chan int) {  
	sum := 0
	for number != 0 {
		digit := number % 10
		sum += digit * digit
		number /= 10
	}
	squareop <- sum
}

func calcCubes(number int, cubeop chan int) {  
	sum := 0 
	for number != 0 {
		digit := number % 10
		sum += digit * digit * digit
		number /= 10
	}
	cubeop <- sum
} 

func main() {  
	number := 589
	sqrch := make(chan int)
	cubech := make(chan int)
	go calcSquares(number, sqrch)
	go calcCubes(number, cubech)
	squares, cubes := <-sqrch, <-cubech
	fmt.Println("Final output", squares + cubes)
}
```
[在线运行程序](https://play.golang.org/p/4RKr7_YO_B)  

在第 7 行，函数 `calcSquares` 计算一个数每位的平方和，并把结果发送给信道 `squareop`。与此类似，在第 17 行函数 `calcCubes` 计算一个数每位的立方和，并把结果发送给信道 `cubop`。  

这两个函数分别在单独的协程里运行（第 31 行和第 32 行），每个函数都有传递信道的参数，以便写入数据。Go 主协程会在第 33 行等待两个信道传来的数据。一旦从两个信道接收完数据，数据就会存储在变量 `squares` 和 `cubes` 里，然后计算并打印出最后结果。该程序会输出：  

```
Final output 1536 
```

## 死锁

使用信道需要考虑的一个重点是死锁。当 Go 协程给一个信道发送数据时，照理说会有其他 Go 协程来接收数据。如果没有的话，程序就会在运行时触发 panic，形成死锁。  

同理，当有 Go 协程等着从一个信道接收数据时，我们期望其他的 Go 协程会向该信道写入数据，要不然程序就会触发 panic。 

```go
package main

func main() {  
	ch := make(chan int)
	ch <- 5
}
```
[在线运行程序](https://play.golang.org/p/q1O5sNx4aW)  

在上述程序中，我们创建了一个信道 `ch`，接着在下一行 `ch <- 5`，我们把 `5` 发送到这个信道。对于本程序，没有其他的协程从 `ch` 接收数据。于是程序触发 panic，出现如下运行时错误。  

```
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:  
main.main()  
	/tmp/sandbox249677995/main.go:6 +0x80
```

## 单向信道

我们目前讨论的信道都是双向信道，即通过信道既能发送数据，又能接收数据。其实也可以创建单向信道，这种信道只能发送或者接收数据。  

```go
package main

import "fmt"

func sendData(sendch chan<- int) {  
	sendch <- 10
}

func main() {  
	sendch := make(chan<- int)
	go sendData(sendch)
	fmt.Println(<-sendch)
}
```
[在线运行程序](https://play.golang.org/p/PRKHxM-iRK)  

上面程序的第 10 行，我们创建了唯送（Send Only）信道 `sendch`。`chan<- int` 定义了唯送信道，因为箭头指向了 `chan`。在第 12 行，我们试图通过唯送信道接收数据，于是编译器报错：  

```
main.go:11: invalid operation: <-sendch (receive from send-only type chan<- int)
```

**一切都很顺利，只不过一个不能读取数据的唯送信道究竟有什么意义呢？**  

**这就需要用到信道转换（Channel Conversion）了。把一个双向信道转换成唯送信道或者唯收（Receive Only）信道都是行得通的，但是反过来就不行。**  

```go
package main

import "fmt"

func sendData(sendch chan<- int) {  
	sendch <- 10
}

func main() {  
	chnl := make(chan int)
	go sendData(chnl)
	fmt.Println(<-chnl)
}
```
[在线运行程序](https://play.golang.org/p/aqi_rJ1U8j)  

在上述程序的第 10 行，我们创建了一个双向信道 `cha1`。在第 11 行 `cha1` 作为参数传递给了 `sendData` 协程。在第 5 行，函数 `sendData` 里的参数 `sendch chan<- int` 把 `cha1` 转换为一个唯送信道。于是该信道在 `sendData` 协程里是一个唯送信道，而在 Go 主协程里是一个双向信道。该程序最终打印输出 `10`。  

## 关闭信道和使用 for range 遍历信道

数据发送方可以关闭信道，通知接收方这个信道不再有数据发送过来。  

当从信道接收数据时，接收方可以多用一个变量来检查信道是否已经关闭。  

```
v, ok := <- ch  
```

上面的语句里，如果成功接收信道所发送的数据，那么 `ok` 等于 true。而如果 `ok` 等于 false，说明我们试图读取一个关闭的通道。从关闭的信道读取到的值会是该信道类型的零值。例如，当信道是一个 `int` 类型的信道时，那么从关闭的信道读取的值将会是 `0`。  

```go
package main

import (  
	"fmt"
)

func producer(chnl chan int) {  
	for i := 0; i < 10; i++ {
		chnl <- i
	}
	close(chnl)
}
func main() {  
	ch := make(chan int)
	go producer(ch)
	for {
		v, ok := <-ch
		if ok == false {
			break
		}
		fmt.Println("Received ", v, ok)
	}
}
```
[在线运行程序](https://play.golang.org/p/XWmUKDA2Ri)  

在上述的程序中，`producer` 协程会从 0 到 9 写入信道 `chn1`，然后关闭该信道。主函数有一个无限的 for 循环（第 16 行），使用变量 `ok`（第 18 行）检查信道是否已经关闭。如果 `ok` 等于 false，说明信道已经关闭，于是退出 for 循环。如果 `ok` 等于 true，会打印出接收到的值和 `ok` 的值。  

```
Received  0 true  
Received  1 true  
Received  2 true  
Received  3 true  
Received  4 true  
Received  5 true  
Received  6 true  
Received  7 true  
Received  8 true  
Received  9 true  
```

for range 循环用于在一个信道关闭之前，从信道接收数据。  

接下来我们使用 for range 循环重写上面的代码。  

```go
package main

import (  
	"fmt"
)

func producer(chnl chan int) {  
	for i := 0; i < 10; i++ {
		chnl <- i
	}
	close(chnl)
}
func main() {  
	ch := make(chan int)
	go producer(ch)
	for v := range ch {
		fmt.Println("Received ",v)
	}
}
```
[在线运行程序](https://play.golang.org/p/JJ3Ida1r_6)  

在第 16 行，for range 循环从信道 `ch` 接收数据，直到该信道关闭。一旦关闭了 `ch`，循环会自动结束。该程序会输出：  

```
Received  0  
Received  1  
Received  2  
Received  3  
Received  4  
Received  5  
Received  6  
Received  7  
Received  8  
Received  9  
```

我们可以使用 for range 循环，重写[信道的另一个示例](#)这一节里面的代码，提高代码的可重用性。  

如果你仔细观察这段代码，会发现获得一个数里的每位数的代码在 `calcSquares` 和 `calcCubes` 两个函数内重复了。我们将把这段代码抽离出来，放在一个单独的函数里，然后并发地调用它。  

```go
package main

import (  
	"fmt"
)

func digits(number int, dchnl chan int) {  
	for number != 0 {
		digit := number % 10
		dchnl <- digit
		number /= 10
	}
	close(dchnl)
}
func calcSquares(number int, squareop chan int) {  
	sum := 0
	dch := make(chan int)
	go digits(number, dch)
	for digit := range dch {
		sum += digit * digit
	}
	squareop <- sum
}

func calcCubes(number int, cubeop chan int) {  
	sum := 0
	dch := make(chan int)
	go digits(number, dch)
	for digit := range dch {
		sum += digit * digit * digit
	}
	cubeop <- sum
}

func main() {  
	number := 589
	sqrch := make(chan int)
	cubech := make(chan int)
	go calcSquares(number, sqrch)
	go calcCubes(number, cubech)
	squares, cubes := <-sqrch, <-cubech
	fmt.Println("Final output", squares+cubes)
}
```
[在线运行程序](https://play.golang.org/p/oL86W9Ui03)  

上述程序里的 `digits` 函数，包含了获取一个数的每位数的逻辑，并且 `calcSquares` 和 `calcCubes` 两个函数并发地调用了 `digits`。当计算完数字里面的每一位数时，第 13 行就会关闭信道。`calcSquares` 和 `calcCubes` 两个协程使用 for range 循环分别监听了它们的信道，直到该信道关闭。程序的其他地方不变，该程序同样会输出：  

```
Final output 1536  
```

本教程的内容到此结束。关于信道还有一些其他的概念，比如缓冲信道（Buffered Channel）、工作池（Worker Pool）和 select。我们会在接下来的教程里专门介绍它们。感谢阅读。祝你愉快。  

**下一教程 - [缓冲信道和工作池](#)**

---

via: https://golangbot.com/channels/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
