首发于：https://studygolang.com/articles/19394

# Go 高级并发模式：第一部分

写代码难，写处理并行和并发的代码更难！要做到这一切并保持高效将是极具挑战性的。

今天，我决定开始分享一些技巧来处理某些特殊情况。

## 定时 Channel 操作

有时，你想要为你的 Channel 操作定时：持续尝试做一些事情，如果不能在一段时间内完成就放弃继续尝试。

要做到这一点，你可以使用 `context` 或者 `time`，两者都很好。`context` 可能更惯用，而 `time` 则更高效，但它们几乎是完全相同的：

```go
func ToChanTimedContext(ctx context.Context, d time.Duration, message Type, c chan<- Type) (written bool) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	select {
	case c <- message:
		return true
	case <-ctx.Done():
		return false
	}
}

func ToChanTimedTimer(d time.Duration, message Type, c chan<- Type) (written bool) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case c <- message:
		return true
	case <-t.C:
		return false
	}
}
```

由于并不真正关心性能（毕竟我们是在等待），我发现唯一的区别是使用 `context` 的解决方案会执行更多的分配（也因为使用 Timer 的那种可以进一步优化以回收 Timer）。

请注意，重复使用 timer 是非常复杂的。因此请记住，如果仅仅为了节省 10 allocs/op 的资源损耗而去复用 timer，很可能并不值得。

如果你感兴趣，[这里](https://blogtitle.github.io/go-advanced-concurrency-patterns-part-2-timers) 有关于如何使用 timer 的文章。

## 先来先服务

有时你希望将相同的消息写入多个 Channel，先写入任何可用的 Channel，但**绝不要在同一 Channel 上两次写入相同的消息**。

要做到这一点，有两种方法：你可以使用局部变量屏蔽 Channel，并相应地禁用 `select` 的 `case` 子句，或者使用 goroutine/wait 方案。

```go
func FirstComeFirstServedSelect(message Type, a, b chan<- Type) {
	for i := 0; i < 2; i++ {
		select {
		case a <- message:
			a = nil
		case b <- message:
			b = nil
		}
	}
}

func FirstComeFirstServedGoroutines(message Type, a, b chan<- Type) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { a <- message; wg.Done() }()
	go func() { b <- message; wg.Done() }()
	wg.Wait()
}
```

请**注意**，在这种情况下，性能可能很重要。而且在编写生成 Goroutine 的解决方案时，所花费的时间几乎是使用 select 的解决方案的 4 倍。

如果在编译期不知道 Channel 的数量，则第一个解决方案将变得更为复杂，但仍然有可能实现，而第二个解决方案则基本保持不变。

注意：如果你的程序有许多未知大小的活动部件，则有必要进行重新设计，因为这很可能简化它。

> 如果你的代码在你检查后仍然有未绑定的活动部分，这里有两个解决方案来提供支持 :

```go
func FirstComeFirstServedGoroutinesVariadic(message Type, chs ...chan<- Type) {
	var wg sync.WaitGroup
	wg.Add(len(chs))
	for _, c := range chs {
		c := c
		go func() { c <- message; wg.Done() }()
	}
	wg.Wait()
}

func FirstComeFirstServedSelectVariadic(message Type, chs ...chan<- Type) {
	cases := make([]reflect.SelectCase, len(chs))
	for i, ch := range chs {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(ch),
			Send: reflect.ValueOf(message),
		}
	}
	for i := 0; i < len(chs); i++ {
		chosen, _, _ := reflect.Select(cases)
		cases[chosen].Chan = reflect.ValueOf(nil)
	}
}
```

不用说：使用反射的解决方案比使用 Goroutine 的解决方案慢几个数量级，所以请不要使用它。

## 整合在一起

如果你想在一段时间内尝试几次发送并且如果它在这里花费了太多时间就中止尝试，这里有两种解决方案：一种是 `time`+`select`，另一种是 `context`+`go`。如果在编译期知道 Channel 的数量，则第一种更好，否则，就应该使用另一个方案。

```go
func ToChansTimedTimerSelect(d time.Duration, message Type, a, b chan Type) (written int) {
	t := time.NewTimer(d)
	for i := 0; i < 2; i++ {
		select {
		case a <- message:
			a = nil
		case b <- message:
			b = nil
		case <-t.C:
			return i
		}
	}
	t.Stop()
	return 2
}

func ToChansTimedContextGoroutines(ctx context.Context, d time.Duration, message Type, ch ...chan Type) (written int) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	var (
		wr int32
		wg sync.WaitGroup
	)
	wg.Add(len(ch))
	for _, c := range ch {
		c := c
		go func() {
			defer wg.Done()
			select {
			case c <- message:
				atomic.AddInt32(&wr, 1)
			case <-ctx.Done():
			}
		}()
	}
	wg.Wait()
	return int(wr)
}
```

关于这个话题，想要了解更多？ 敬请关注！

与此同时，我建议观看 Sameer Ajmani 的 “高级 Go 并发模式”：[视频](https://www.youtube.com/watch?v=QDDwwePbDtw)、[幻灯片](https://talks.golang.org/2013/advconc.slide)。

---

via: https://blogtitle.github.io/go-advanced-concurrency-patterns-part-1/

作者：[Rob](https://blogtitle.github.io/authors/rob/)
译者：[SataQiu](https://github.com/SataQiu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
