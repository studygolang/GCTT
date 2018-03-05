已发布：https://studygolang.com/articles/12495

# Goroutine 泄露

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/goroutine-leak/cover.jpg)

Go 中的并发性是以 goroutine（独立活动）和 channel（用于通信）的形式实现的。处理 goroutine 时，程序员需要小心翼翼地避免泄露。如果最终永远堵塞在 I/O 上（例如 channel 通信），或者陷入死循环，那么 goroutine 会发生泄露。即使是阻塞的 goroutine，也会消耗资源，因此，程序可能会使用比实际需要更多的内存，或者最终耗尽内存，从而导致崩溃。让我们来看看几个可能会发生泄露的例子。然后，我们将重点关注如何检测程序是否受到这种问题的影响。

## 发送到一个没有接收者的 channel

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/goroutine-leak/1.jpg)

假设出于冗余的目的，程序发送请求到许多后端。使用首先收到的响应，丢弃后面的响应。下面的代码将会通过等待随机数毫秒，来模拟向下游服务器发送请求：

```go
package main

import (  
	"fmt"  
	"math/rand"  
	"runtime"  
	"time"  
)

func query() int {  
	n := rand.Intn(100)  
	time.Sleep(time.Duration(n) * time.Millisecond)  
	return n  
}

func queryAll() int {  
	ch := make(chan int)  
	go func() { ch <- query() }()  
	go func() { ch <- query() }()  
	go func() { ch <- query() }()  
	return <-ch  
}

func main() {  
	for i := 0; i < 4; i++ {  
		queryAll()  
		fmt.Printf("#goroutines: %d", runtime.NumGoroutine())  
	}  
}
```

输出：

```
#goroutines: 3  
#goroutines: 5  
#goroutines: 7  
#goroutines: 9
```

每次调用 _queryAll_ 后，goroutine 的数目会发生增长。问题在于，在接收到第一个响应后，“较慢的” goroutine 将会发送到另一端没有接收者的 channel 中。

可能的解决方法是，如果提前知道后端服务器的数量，那么使用缓存 channel。否则，只要至少有一个 goroutine 仍在工作，我们就可以使用另一个 goroutine 来接收来自这个 channel 的数据。其他的解决方案可能是使用 [context](https://golang.org/pkg/context/)（[example](http://golang.rakyll.org/leakingctx/)），利用 某些机制来取消其他请求。

## 从没有发送者的 channel 中接收数据

这种场景类似于发送到一个没有接收者的 channel。[泄露 goroutine](http://openmymind.net/Leaking-Goroutines/) 这篇文章中包含了一个示例。

## nil channel

写入到 _nil_ channel 会永远阻塞：

```go
package main

func main() {  
	var ch chan struct{}  
	ch <- struct{}{}  
}
```

所以它导致死锁：

```
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send (nil chan)]:  
main.main()  
...
```

当从 _nil_ channel 读取数据时，同样的事情发生了：

```go
var ch chan struct{}  
<-ch
```

当传递尚未初始化的 channel 时，也可能会发生：

```go
package main

import (  
	"fmt"  
	"runtime"  
	"time"  
)

func main() {  
	var ch chan int  
	if false {  
		ch = make(chan int, 1)  
		ch <- 1  
	}  
	go func(ch chan int) {  
		<-ch  
	}(ch)

	c := time.Tick(1 * time.Second)  
	for range c {  
		fmt.Printf("#goroutines: %d", runtime.NumGoroutine())  
	}  
}
```

在这个例子中，有一个显而易见的罪魁祸首 —— `if false {`，但是在更大的程序中，更容易忘记这件事，然后使用 channel 的零值（_nil_）。

## 死循环

goroutine 泄露不仅仅是因为 channel 的错误使用造成的。泄露的原因也可能是 I/O 操作上的堵塞，例如发送请求到 API 服务器，而没有使用超时。另一种原因是，程序可以单纯地陷入死循环中。

## 分析

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/goroutine-leak/2.jpg

### runtime.NumGoroutine

简单的方式是使用由 [_runtime.NumGoroutine_](https://golang.org/pkg/runtime/#NumGoroutine) 返回的值。

### net/http/pprof

```go
import (  
	"log"  
	"net/http"  
	_ "net/http/pprof"  
)

...

log.Println(http.ListenAndServe("localhost:6060", nil))
```

调用 http://localhost:6060/debug/pprof/goroutine?debug=1 ，将会返回带有堆栈跟踪的 goroutine 列表。

### runtime/pprof

要将现有的 goroutine 的堆栈跟踪打印到标准输出，请执行以下操作：

```go
import (  
	"os"  
	"runtime/pprof"  
)

...

pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
```

### [gops](https://github.com/google/gops)

```
> go get -u github.com/google/gops
```

集成到你的程序中：

```go
import "github.com/google/gops/agent"

...

if err := agent.Start(); err != nil {  
	log.Fatal(err)  
}  
time.Sleep(time.Hour)
```

```
> ./bin/gops  
12365   gops    (/Users/mlowicki/projects/golang/spec/bin/gops)  
12336*  lab     (/Users/mlowicki/projects/golang/spec/bin/lab)  
> ./bin/gops vitals -p=12336  
goroutines: 14  
OS threads: 9  
GOMAXPROCS: 4  
num CPU: 4
```

### [leaktest](https://github.com/fortytw2/leaktest)

这是用测试来自动检测泄露的方法之一。它基本上是在测试的开始和结束的时候，利用 [runtime.Stack](https://golang.org/pkg/runtime/#Stack) 获取活跃 goroutine 的堆栈跟踪。如果在测试完成后还有一些新的 goroutine，那么将其归类为泄露。

---

分析甚至已经在运行的程序的 goroutine 管理，以避免可能会导致内存不足的泄露，这至关重要。代码在生产上运行数日后，这样的问题通常就会出现，因此它可能会造成真正的损害。

点击原文中的 ❤ 以帮助其他人发现这个问题。如果你想实时获得新的更新，请关注原作者哦~


## 资源
* [包 —— Go 编程语言](https://golang.org/pkg/)
	
	bufio 包实现了缓存 I/O。它封装一个 io.Reader 或者 io.Writer 对象，创建其他对象（Reader 或者……）

* [google/gops](https://github.com/google/gops)

	gops —— 一个列出和诊断当前运行在你的系统上的 Go 进程的工具。

* [runtime：检测僵尸 goroutine · 问题 #5308 · golang/go](https://github.com/golang/go/issues/5308)

	runtime 可以检测不可达 channel / mutex 等上面的 goroutine 阻塞，然后报告此类问题。这需要一个接口……

* [fortytw2/leaktest](https://github.com/fortytw2/leaktest)

	leaktest - goroutine 泄露检测器。

----------------

via: https://medium.com/golangspec/goroutine-leak-400063aef468

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[ictar](https://github.com/ictar)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出