首发于：https://studygolang.com/articles/28990

# Go: Select 语句的执行顺序

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-ordering-in-select-statements/20200429220520.png)

> 本文基于 Go 1.14

`select` 允许在一个 goroutine 中管理多个 channel。但是，当所有 channel 同时就绪的时候，go 需要在其中选择一个执行。此外，go 还需要处理没有 channel 就绪的情况，我们先从就绪的 channel 开始。

## 顺序

`select` 不会按照任何规则或者优先级选择就绪的 channel。go 标准库在每次执行的时候，都会将他们顺序打乱，也就是说不能保证任何顺序。

看一个有三个就绪的 channel 的例子：

``` go
func main() {
	a := make(chan bool, 100)
	b := make(chan bool, 100)
	c := make(chan bool, 100)
	for i := 0; i < 10; i++ {
		a <- true
		b <- true
		c <- true
	}
	for i := 0; i < 10; i++ {
		select {
		case <-a:
			print("< a")

		case <-b:
			print("< b")

		case <-c:
			print("< c")

		default:
			print("< default")
		}
	}
}
```

这三个 channel 的缓冲区都填满了，使得 select 选择时不会堵塞。下面是程序的输出：

```bash
< b< a< a< b< c< c< c< a< b< b
```

在 select 的每次迭代中，case 都会被打乱：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-ordering-in-select-statements/20200429223415.png)

由于 go 不会删除重复的 channel，所以可以使用多次添加 case 来影响结果，代码如下：

```go
func main() {
	a := make(chan bool, 100)
	b := make(chan bool, 100)
	c := make(chan bool, 100)
	for i := 0; i < 10; i++ {
		a <- true
		b <- true
		c <- true
	}
	for i := 0; i < 10; i++ {
		select {
		case <-a:
			print("< a")
		case <-a:
			print("< a")
		case <-a:
			print("< a")
		case <-a:
			print("< a")
		case <-a:
			print("< a")
		case <-a:
			print("< a")
		case <-a:
			print("< a")

		case <-b:
			print("< b")

		case <-c:
			print("< c")

		default:
			print("< default")
		}
	}
}
```

输出的结果：

```shell
< c< a< b< a< b< a< a< c< a< a
```

当所有 channel 同时准备就绪时，有 80％的机会选择通道 a。下面来看一下 channel 未就绪的情况。

## 没有就绪 channels

`select` 运行时，如果没有一个 case channel 就绪，那么他就会运行 `default:`,如果 `select` 中没有写 default，那么他就进入等待状态，如下面这个例子

```go
func main() {
	a := make(chan bool, 100)
	b := make(chan bool, 100)
	Go func() {
		time.Sleep(time.Minute)
		for i := 0; i < 10; i++ {
			a <- true
			b <- true
		}
	}()

	for i := 0; i < 10; i++ {
		select {
		case <-a:
			print("< a")
		case <-b:
			print("< b")
		}
	}
}
```

上面那个例子中，将在一分钟后打印结果。`select` 阻塞在 channel 上。这种情况下，处理 `select` 的函数将会订阅所有 channel 并且等待，下面是一个 goroutine#7 在 select 中等待的示例，其中另一个 goroutine#4 也在等待 channel：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-ordering-in-select-statements/20200429225528.png)

Goroutine(G7)订阅所有频道并在列表末尾等待。 如果 channel 发送了一条消息，channel 将通知已在等待该消息的另一个 Goroutine。一旦收到通知，`select` 将取消订阅所有 channel，并且返回到代码运行.

更多关于 channel 与等待队列的信息，请查看作者另外一篇文章[*Go: 带缓冲和不带缓冲的 Channels*](https://studygolang.com/articles/23538)。

上面介绍的逻辑，都是针对于有两个或者以上的活动的 channel，实际上如果只有一个活动的 channel，Go 乐意简化 select。

## 简化

如果只有一个 case 加上一个 default，例子：

```go
func main() {
	t:= time.NewTicker(time.Second)
	for   {
		select {
		case <-t.C:
			print("1 second ")
		default:
			print("default branch")
		}
	}
}
```

这种情况下。Go 会以非阻塞模式读取 channel 的操作替换 select 语句。如果 channel 在缓冲区中没有任何值，或者发送方准备发送消息，将会运行 default。就像下面这张图：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-ordering-in-select-statements/20200429231908.png)

如果没有 default，则 Go 通过阻塞 channel 的操作方式重写 select 语句。

---

via: https://medium.com/a-journey-with-go/go-ordering-in-select-statements-fd0ff80fd8d6

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[yixiao9206](https://github.com/yixiao9206)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
