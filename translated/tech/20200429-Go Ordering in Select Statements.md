# Go: Ordering in Select Statements

----------------

via: [原文链接](https://medium.com/a-journey-with-go/go-ordering-in-select-statements-fd0ff80fd8d6)

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent?source=post_page-----fd0ff80fd8d6----------------------)
		译者：[yixiao9206](https://github.com/yixiao9206)
		校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

----------------

![](https://blog-image-1253555052.cos.ap-guangzhou.myqcloud.com/20200429220520.png)

> 本文基于 go 1.14

`select` 允许在一个goroutine中管理多个channel。但是，当所有channel同时就绪的时候，go需要在其中选择一个执行。此外，go还需要处理没有channel就绪的情况，我们先从就绪的channel开始。

## 顺序

`select` 不会按照任何规则或者优先级选择就绪的channel。go标准库在每次执行的时候，都会将他们顺序打乱，也就是说不能保证任何顺序。

看一个有三个就绪的channel的例子：

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

这三个channel都有三个完整的buffer（不会阻塞），下面是程序的输出

``` shell
< b< a< a< b< c< c< c< a< b< b
```

在 select 的每次迭代中，case 都会被打乱：

![](https://blog-image-1253555052.cos.ap-guangzhou.myqcloud.com/20200429223415.png)

由于go 不会删除重复的channel，所以可以使用多次添加case来影响结果，代码如下：

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

当所有channel同时准备就绪时，有80％的机会选择通道a。下面来看一下channel未就绪的情况。

## 没有就绪 channels

`select` 运行时，如果没有一个case channel就绪，那么他就会运行`default:`,如果 `select`中没有写default，那么他就进入等待状态，如下面这个例子

```go
func main() {
   a := make(chan bool, 100)
   b := make(chan bool, 100)
   go func() {
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

上面那个例子中，将在一分钟后打印结果。`select`阻塞在 channel上。这种情况下，处理`select`的函数将会订阅所有channel并且等待，下面是一个goroutine#7在select中等待的示例，其中另一个goroutine#4也在等待channel：

![](https://blog-image-1253555052.cos.ap-guangzhou.myqcloud.com/20200429225528.png)

Goroutine(G7)订阅所有频道并在列表末尾等待。 如果channel发送了一条消息，channel将通知已在等待该消息的另一个Goroutine。一旦收到通知，`select`将取消订阅所有channel，并且返回到代码运行.

更多关于channel与等待队列的信息，请查看作者另外一篇文章[*Go: Buffered and Unbuffered Channels*](https://medium.com/a-journey-with-go/go-buffered-and-unbuffered-channels-29a107c00268)*.*

上面介绍的逻辑，都是针对于有两个或者以上的活动的channel，实际上如果只有一个活动的channel，go乐意简化select

## 简化

如果只有一个case 加上一个default，例子：

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

这种情况下。Go会以非阻塞模式读取channel的操作替换select语句。如果channel在缓冲区中没有任何值，或者发送方准备发送消息，将会运行default。就像下面这张图：

![](https://blog-image-1253555052.cos.ap-guangzhou.myqcloud.com/20200429231908.png)

如果没有default，则 Go 通过阻塞channel的操作方式重写 select 语句。