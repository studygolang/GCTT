首发于：https://studygolang.com/articles/21758

# Go 高并发模式：第二部分（计时器）

正如我在[上一篇文章](https://studygolang.com/articles/19394) 中所述，准确使用计时器很难的，所以这里进行一些说明。

## 前言

如果你认为结合 Goroutines 去处理时间和计数器很简单的话，那你就错了，这里有提到的一些与 time.Timer 相关的问题或 bug：

* [time: Timer.Reset is not possible to use correctly #14038](https://github.com/golang/go/issues/14038)
* [time: Timer.C can still trigger even after Timer.Reset is called #11513](https://github.com/golang/go/issues/11513)
* [time: document proper usage of Timer.Stop #14383](https://github.com/golang/go/issues/14383)

看完上面的链接内容后，如果你依然认为很简单，那来看看下面的代码，如下代码会产生死锁和竞争条件

```go
tm := time.NewTimer(1)
tm.Reset(100 * time.Millisecond)
<-tm.C
if !tm.Stop() {
	<-tm.C
}
```

死锁代码片段

```go
func toChanTimed(t *time.Timer, ch chan int) {
	t.Reset(1 * time.Second)
	defer func() {
		if !t.Stop() {
			<-t.C
		}
	}()
	select {
	case ch <- 42:
	case <-t.C:
	}
}
```

可能代码比较难懂，下面对相关方法进行阐述。

## time.Ticker

```go
type Ticker struct {
	C <-chan Time // The channel on which the ticks are delivered.
}
```

Ticker 简单易用，但也有一些小问题

* 如果 C 中已存在一条消息，则发送消息时将删除所有未读值。
* 必须有停止操作： 否则 GC 无法回收它
* 设置 C 无用：消息仍将在原始的 channel 上发送。

## time.Tick

time.Tick 是对 time.NewTicker 的封装。最好不要使用该方法，除非你准备将 `chan` 作为返回结果并在程序的整个生命周期中继续使用它。
正如官方描述：
> 垃圾收集器无法恢复底层的 Ticker，出现 " 泄漏 ".
请谨慎使用，如有疑问请改用 `Ticker`。

## time.After

这与 `Tick` 的概念基本相同，它是对 `Timer` 进行封装。一旦计时器被触发，它将被回收。请注意，计时器使用了缓存容量是 1 的通道，即使没有接收者，它仍可以进行计数。
如上所述，如果您关心性能且希望能够取消计时，那么你不应该使用 `After`。

## time.Timer ( 也称为 time.WhatTheFork?!)

对于 Go 来说这是一个比较奇怪的 API ：`NewTicker(Duration)` 返回了一个 `*Timer` 类型，该类型仅**暴露**一个定义为 chan 类型的变量 `C` ，这点非常奇怪。

通常在 Go 语言中允许导出的字段意味着用户可以获取或设置该字段，而此处设置变量 `C` 并没有实际意义。相反：设置 C 并重置 Timer 并不会影响之前在 C 通道的消息传递。更糟糕的是：AfterFunc 返回的 Timer 根本不会使用到 C。

这样看来，Timer 很奇怪，以下是 API 的概述：

```go
type Timer struct {
	C <-chan Time
}

func AfterFunc(d Duration, f func()) *Timer
func NewTimer(d Duration) *Timer
func (*Timer) Stop(bool)
func (*Timer) Reset(d Duration) bool
```

四个非常简单的函数，其中两个是构造函数，**有可能出错吗？**

### time.AfterFunc

> 官方文档： AfterFunc 持续时间超时后通过开 Goroutine 去调用 f 函数，返回一个 Timer 类型，以便通过 Stop 方法取消调用。

这么描述虽然没有问题，但需要注意：当调用 `Stop` 方法时，如果返回 `false` ，则表示该函数已经执行且停止失败。但并不意味着函数已经返回，你需要添加一些处理逻辑：

```go
done := make(chan struct{})
f := func() {
	doStuff()
	close(done)
}
t := time.AfterFunc(1*time.Second, f)
if !t.Stop() {
	<-done
}
```

这个在 `Stop` 文档中有相关说明。

除此之外，返回的计时器不会被触发，只能用于调用 `Stop` 方法。

```go
t := time.AfterFunc(1*time.Second, func() {
	fmt.Println("Time has passed!")
})
// This will deadlock.
<-t.C
```

此外，写这篇文章的时候，重置计时器会在传入重置函数的时间段过去后再次调用 f，但这种特性目前暂没有文档规范，未来可能会被改变。

### time.NewTimer

> 官方文档 : NewTimer 实例化 Timer 结构体，在持续时间 d 之后发送当前时间至通道内 .

这意味着没有声明它就无法构建有效的 `Timer` 类型结构体。 如果你需要构建一个以便后续重复使用，可以用该方法进行实例化，或者使用如下代码实现自主创建和停止计数器

```go
t := time.NewTimer(0)
if !t.Stop() {
	<-t.C
}
```

你**必须**从 channel 中读取数据。假如在 `New` 和 `Stop` 调用期间触发了定时器，且 channel 存在未消费的数据， 则 `C` 会存在一个值。将导致后续读取均是错误的。

### `(*time.Timer).Stop`

> Stop 方法会阻止计时器触发。如果调用停止计时器的方法，则返回 true，如果计时器已超时**或者**已停止，则返回 false。

以上句子中的“或”非常重要。文档中所以关于 Stop 的示例都显示了以下代码片段：

```go
if !t.Stop() {
	<-t.C
}
```

关键点在于 "or" 它意味着有效 0 次或 1 次。对已消费完通道数据和在此期间未调用 Reset 进行过多次执行的情况，均是无效的。综上所述，当且仅当没有执行对通道数据的消费，Stop+drain 才是安全的。

在文档中体现如下：

> 例如：假设程序尚未从 t.C 接收数据：

此外，上面的模式不是线程安全的，因为当消费完通道数据时，Stop 返回的值可能已经过时了，两个 Goroutine 尝试消费通道 C 数据也会导致死锁。

### `(*time.Timer).Reset`

这个方法更有意思，文档很长，你可以在[这里](https://golang.org/pkg/time/#Timer.Reset) 进行查看

文档中一个有趣的摘录：

> 请注意，因为在清空 channel 和计数器到期之间存在竞争条件，我们无法正确使用 Reset 返回值。Reset 方法必须作用于已停止或已过期的 channel 上。

文档所提供 Reset 正确使用方法如下：

```go
if !t.Stop() {
	<-t.C
}
t.Reset(d)
```

不能与来自通道的其他接收者同时使用 `Stop` 和 `Reset` 方法， 为了使 `C` 上传递的消息有效，C 应该在每次 `重置` 之前被消费完。

重置计时器而不清空它将使运行过程时丢弃该值，因为 `C` 缓存为 1，运行时对其他执行是[有损发送](https://golang.org/src/time/sleep.go?s=#L134)。

### time.Timer: 把这些方法放在一起

* Stop 仅作用在 New 和 Reset 方法之后才安全
* Reset 仅在 Stop 方法后有效。
* 只有在每次运行 Stop 后，channel 消费完时，所接收的值才是有效的。
* 只有 channel 未被消费时，才允许清空 channel。

以下是计时器转换，使用和调用关系流程图：

![timer.png](https://raw.githubusercontent.com/studygolang/gctt-images/master/Go-Advanced-concurrency-patterns-part-2-timers/timer.png)

如下是一个正确复用计时器的例子，它解决了文章开头提到的一些问题：

```go
func toChanTimed(t *time.Timer, ch chan int) {
	t.Reset(1 * time.Second)
	// No defer, as we don't know which
	// case will be selected

	select {
	case ch <- 42:
	case <-t.C:
		// C is drained, early return
		return
	}

	// We still need to check the return value
	// of Stop, because t could have fired
	// between the send on ch and this line.
	if !t.Stop() {
		<-t.C
	}
}

```

上述代码可以确保 toChanTimed 返回后可以重新使用计时器

## 想知道更多吗

本文中所提到的类型和函数均依赖于计数器的运行，只是使用方式不一样。[time/sleep.go](https://golang.org/src/time/sleep.go) 包含了使用它们的大部分代码。

如下表中，包含由 `time` 包设置的 `runtimeTimer` 字段

|Constructor|`when` 字段 |`period` 字段 |`f` 字段 |`arg` 字段 |
|---|---|---|---|---|
|NewTicker(d)|`d`|set to `d`|`sendTime`|`C`|
|NewTimer(d)|`d`|not set|`sendTime`|`C`|
|AfterFunc(d,f)|`d`|not set|`goFunc`|`f`|

运行计数器不依赖于 Goroutine ，而是以更高效精确的方式组合使用。你可以在 [runtime/time.go](https://golang.org/src/runtime/time.go) 包中深入了解实现细节。祝学的开心！

---
via: https://blogtitle.github.io/go-advanced-concurrency-patterns-part-2-timers/

作者：[Rob](https://blogtitle.github.io/authors/rob/)
译者：[liulizhi](https://github.com/liulizhi)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
