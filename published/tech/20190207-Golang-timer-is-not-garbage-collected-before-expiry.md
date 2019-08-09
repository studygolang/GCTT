首发于：https://studygolang.com/articles/22617

# Golang <-time.After（）在计时器过期前不会被垃圾回收

最近我在调查 Go 应用程序中内存泄漏的问题，这个问题主要因为我没有正确的阅读文档。这是一段导致消耗了多个 Gbs 内存的代码：

```go
func ProcessChannelMessages(ctx context.Context, in <-chan string, idleCounter prometheus.Counter) {
	for {
		start := time.Now()
		select {
		case s, ok := <-in:
			if !ok {
				return
			}
			// handle `s`
		case <-time.After(5 * time.Minute):
			idleCounter.Inc()
		case <-ctx.Done():
			return
		}
	}
}
```

以下是应用程序的内存指标图：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Golang-Timer-After-is-not-garbage-collected-before-expiry/1.png)

在图中左侧，可以看到修复之前的内存消耗，右侧是修改后的内存消耗。分析器显示 `<-time.After` 是内存泄漏的原因。直到我读到以下文档时，我才感到惊讶：

> 在计时器触发之前，垃圾收集器不会回收 Timer。

所以 9Gb 的内存被定期垃圾回收变得非常必要了。我们在 channel 中每秒有 60k 个消息，在每个给定时刻分配大约 1800 万个计时器加上一些不确定数字的计时器等待被垃圾回收。

```go
func ProcessChannelMessages(ctx context.Context, in <-chan string, idleCounter prometheus.Counter) {
	idleDuration := 5 * time.Minute
	idleDelay := time.NewTimer(idleDuration)
	defer idleDelay.Stop()
	for {
		idleDelay.Reset(idleDuration)
		select {
		case s, ok := <-in:
			if !ok {
				return
			}
			// handle `s`
		case <-idleDelay.C:
			idleCounter.Inc()
		case <-ctx.Done():
			return
		}
	}
}
```

微不足道的重构有助于将内存消耗减少 20 倍，这些都是阅读文档就能解决的问题。

---

via: https://medium.com/@oboturov/golang-time-after-is-not-garbage-collected-4cbc94740082

作者：[Artem OBOTUROV](https://medium.com/@oboturov)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
