已发布：https://studygolang.com/articles/13067

# 如何泄漏一个协程然后修复它

很多 go 语言开发者都知道这句格言，[永远不要启动一个你不知道如何停止的协程](https://dave.cheney.net/2016/12/22/never-start-a-goroutine-without-knowing-how-it-will-stop)，但是泄漏一个协程还是超级的简单。让我们看一种常碰到的泄漏协程的方式，然后修复它。

为了实现这个，我们先建立一个包含一个自定义 `map` 类型的库，这个 `map` 类型的 key 在经过了一段可配置的时间后过期。我们把这个库叫做 [ttl](https://en.wikipedia.org/wiki/Time_to_live) ，这个库有一个 `API` 类似如下：

```go
// 创建一个生存周期为5分钟的map
m := ttl.NewMap(5*time.Minute)
//设置一个key
m.Set("my-key", []byte("my-value"))

// 读取一个key
v, ok := m.Get("my-key")
//得到 "my-value"
fmt.Println(string(v))
// true, key存在
fmt.Println(ok)

// ... 过了5分钟之后
v, ok := m.Get("my-key")
// 没有值
fmt.Println(string(v) == "")
// false, key已经过期了
fmt.Println(ok)
```

为了确保key会过期，我们在NewMap函数中启动一个协程。

```go
func NewMap(expiration time.Duration) *Map {
	m := &Map{
		data:       make(map[string]expiringValue),
		expiration: expiration,
	}

	// start a worker goroutine
	go func() {
		for range time.Tick(expiration) {
			m.removeExpired()
		}
	}()

	return m
}
```

这个工作协程每运行一段配置好的时间后会在这个 `map` 上调用一个方法来删除过期的 `key`。这意味着 `SetKey` 方法必须记录 `key` 的进入时间，这也是为什么 `data` 字段包含一个 `expiringValue` 类型，这个类型与一个记录实际过期时间的值相关联：

```go
type expiringValue struct {
	expiration time.Time
	data       []byte //实际的值
}
```

对于不敏感的人来说，这个工作协程的调用看起来没问题，并且如果这不是一篇关于协程泄漏的文章，扫一眼这几行代码并没有什么让人觉得惊奇的地方，虽然如此，我们还是在构造器中漏泄了一个协程，问题，怎么泄漏的？

我们回顾一下 `map` 类型的生命周期。首先，一个调用者创建了一个 `map` 的实例，在创建实例后，一个工作协程开始运行。接下来调用者可能调用若干次 `Set` 和 `Get` 方法，最终，调用会结束使用这个 `map` 的实例，并且释放所有对它的引用。这时，垃圾收集器应该会正常的回收这个实例对应的内存。然而，工作协程还在运行并且还拥有一个对这个 map 实例的引用。因为并没有其他显示的调用来停止这个协程，我我们就把这个协程已经使用的内存给泄漏了。

让我们把这个问题再说的明白一点。我们使用 `runtime` 包来查看内存收集器和运行的协程在某一时刻的统计数据

```go
func main() {
	go func() {
		var stats runtime.MemStats
		for {
			runtime.ReadMemStats(&stats)
			fmt.Printf("HeapAlloc    = %d\n", stats.HeapAlloc)
			fmt.Printf("NumGoroutine = %d\n", runtime.NumGoroutine())
			time.Sleep(5*time.Second)
		}
	}()

	for {
		work()
	}
}

func work() {
	m := ttl.NewMap(5*time.Minute)
	m.Set("my-key", []byte("my-value"))

	if _, ok := m.Get("my-key"); !ok {
		panic("no value present")
	}
	// m超出变量范围
}
```
不用很长时间，我们就可以看到分配的堆内存和运行的协程数增长得非常，非常的快。

```
HeapAlloc    = 76960
NumGoroutine = 18
HeapAlloc    = 2014278208
NumGoroutine = 1447847
HeapAlloc    = 3932578560
NumGoroutine = 2832416
HeapAlloc    = 5926163224
NumGoroutine = 4322524
```
很明显，我们现在得停止那些协程。目前在 `Map` 上提供的 `API` 中并没有办法停止这个工作协程，如果不改变任何的 API 但是仍然能在调用者使用完 `map` 实例时停止工作协程，那是很理想的。但是只有调用者知道什么时候完了 `map` 实例。
一个常用的解决这个方法的模式是实现一个 `io.Closer` 接口，当调用者用完了 `map` 实例，调用一下 `Close` 方法告诉 `Map` 停止它的工作协程。

```go
func (m *Map) Close() error {
	close(m.done)
	return nil
}
```

在我们的构造器中工作协程的调用会看起来类似这样：

```go
func NewMap(expiration time.Duration) *Map {
	m := &Map{
		data:       make(map[string]expiringValue),
		expiration: expiration,
		done:       make(chan struct{}),
	}

	// 启动一个工作协程
	go func() {
		ticker := time.NewTicker(expiration)
		defer ticker.Stop()
		for {
			select {
				case <-ticker.C:
					m.removeExpired()
				case <-m.done:
					return
			}
		}
	}()

	return m
}
```

现在工作协程包含了一个 `select` 语句，它会检查 `done通道` 也会检查 `ticker 的通道`，主要的，我们还删除了 [time.Tick](https://godoc.org/time#Tick)，因为它并不能让协程顺利关闭还是会造成泄漏。

经过以上的修改，我们简化的统计数据看起像这样：
```
HeapAlloc    = 72464
NumGoroutine = 6
HeapAlloc    = 5175200
NumGoroutine = 59
HeapAlloc    = 5495008
NumGoroutine = 35
HeapAlloc    = 9171136
NumGoroutine = 240
HeapAlloc    = 8347120
NumGoroutine = 53
```
这些数字都非常小，这是因为 `work` 在一个很小的循环中被调用，更重要的是，我们不再看到协程数或者分配的堆内存的飞速增长，这就是我们想要的结果。注意，最终的代码可以在[**这儿**](https://github.com/gobuildit/gobuildit/tree/master/ttl)找到。

这篇文章提供了一个为什么知道一个协程何时停止这么的重要的示例，同时，也提醒我们，监控一个程序中运行的协程数也是非常重要的。这样的监控可以给代码中隐藏的协程泄漏提供一个警示。同时我们也要牢记，有时候协程泄漏需要几天甚至几周才发生。因此对应用程序同时进行短期和长期的监控是非常值得的。

多谢 Jean de Kelerk 和 Jason Keene，他们读了这篇文章的草稿。

---

via: https://commandercoriander.net/blog/2018/05/22/how-to-leak-a-goroutine-then-fix-it/

作者：[Eno Compton](https://enocom.io/)
译者：[MoodWu](https://github.com/moodwu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
