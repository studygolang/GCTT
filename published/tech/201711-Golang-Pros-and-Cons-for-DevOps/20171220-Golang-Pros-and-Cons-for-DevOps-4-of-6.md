已发布：https://studygolang.com/articles/12615

# Golang 之于 DevOps 开发的利与弊（六部曲之四）：time 包和方法重载

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go_devops/golang-pros-cons-4-time-package-method-overloading.png)

万众期待的 Golang 之于 DevOps 开发的利与弊 系列终于回归了！在这篇文章，我们讨论下 Golang 中的 time 包，以及 go 语言中为什么不使用方法重载。

如果你没有读 [最近一篇](https://studygolang.com/articles/12614) 关于 “接口实现和公有/私有命名方式”（原文描述写错了，这一链接对应的应该是 “速度 vs. 缺少泛型”），请一定仔细阅读下，你也可以 [订阅我们的博客更新](http://eepurl.com/cOHJ3f)，以后有系列文章发布的时候你就能收到通知。

- [Golang 之于 DevOps 开发的利与弊（六部曲之一）：Goroutines, Channels, Panics, 和 Errors](https://studygolang.com/articles/11983)
- [Golang 之于 DevOps 开发的利与弊（六部曲之二）：接口实现的自动化和公有/私有实现](https://studygolang.com/articles/12608)
- [Golang 之于 DevOps 开发的利与弊（六部曲之三）：速度 vs. 缺少泛型](https://studygolang.com/articles/12614)
- [Golang 之于 DevOps 开发的利与弊（六部曲之四）：time 包和方法重载](https://studygolang.com/articles/12615)
- [Golang 之于 DevOps 开发的利与弊（六部曲之五）：跨平台编译，Windows，Signals，Docs 以及编译器](https://studygolang.com/articles/12616)
- Golang 之于 DevOps 开发的利与弊（六部曲之六）：Defer 指令和包依赖性的版本控制

## Golang 之利: Time 包的使用

编程的人都知道处理 date 和 time 时候的危险。时间（time）在我们每天的生活中看似平常，可从计算机的角度看，time 处理简直是噩梦。Google 有了这个机会使处理 date 变得轻松点，*他们成功了*！我将 [time 包](https://golang.org/pkg/time/) 的讲解分成了三部分：（1）基本内容，（2）timer 定时器，（3）date 解析。

### 1. Time 包基本内容

你可能认为每一种语言都有一个标准的，易用的处理 time 操作的内置库，其实不是这样的。NPM 有超过 [8000 多个 time 相关的包](https://www.npmjs.com/search?q=time&page=1&ranking=quality)，因为 javascript 的 Date 包没法用。Java8 最终使用 java.time.Instant 和 java.time.chrono 包缓解了这个问题，但仍在编写 [教程](https://www.tutorialspoint.com/java8/java8_datetime_api.htm)，研究各种用 Java 操作 time 的类和方法。相反，Golang 的 [time 包](https://golang.org/pkg/time/) 用一句话就能总结：只需引用一个包，你想要的都能实现。

获取当前时间： `time.Now()`

获取将来某个时刻： `time.Now().Add(5*time.Minute)`

获取经过多少时间（持续时间）：`time.Since(processingStarted)`

轻松地比较持续时间： `if frequency < time.Hour {`

获取当月日期，放弃 [calendar](http://https://docs.oracle.com/javase/7/docs/api/java/util/Calendar.html) 吧！ `time.Now().Day()` 就够了

### 2. Timers 定时器

这个 time 包另一个大的加分项是使用 timer 定时器很方便。DevOps 的应用很明确：我们经常需要安排一些任务在将来执行，一些重复的基本操作，或者就是 sleep 一会。

用 time 包里的 co-locating 定时器，你可以轻松实现 sleep 一个线程，如

```go
time.Sleep(2*time.Minute)
```

或者在将来的某个时刻执行一个函数

```go
time.AfterFunc(5*time.Second, func() {
	fmt.Println("hello in the future")
})
```

或者重复执行一个任务

```go
tick := time.NewTicker(1*time.Minute)
select {
case <- tick.C:
	foo()
}
```

（这些操作）都不需要将时间转换成 second（还是 millsecond 来着？），（也不用）添加 2 个依赖库，（也不用）引入 10 个包。

### 3. Date 解析

不说下我们最常见又最讨厌的话题：将字符串转成 date ，那关于 time 的讨论就是不完整的。这本该是个初级问题，却一点也不简单。date 格式有 [诸多标准](https://en.wikipedia.org/wiki/ISO_8601)，[编程语言内部并没有一个线程安全的实现](https://stackoverflow.com/questions/6840803/why-is-javas-simpledateformat-not-thread-safe) 来解析 date 和  [timezone](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)，就是为了确保不是所有的地方（同一时刻）都是 5 点钟。这就是 Golang 意识到的重要的地方：其他人做的都是错的。

Google 没有生成解决方案去处理成千上万的 date 格式，而是创建了一套系统，使用基于 pattern 的设计，永远参考同一个参照时间： Mon Jan 02 15:04:05 MST 2006 。组成这个单一参照时间的元素，可以重新组合成任何一种你可能遇到的格式。你可以在这里（原文没有链接，疑似 [这里](https://golang.org/pkg/time/#example_Time_Format)）找到准确，完整的说明。

当然，Golang 的 date 解析也不那么完美，是吧？嗯，当然。你至少要考虑一个问题：timezone ，否则不能重新解析 date。除此之外，安排会议也大不可能了，还是因为 Golang 里的 timezone，让类似 “2017/10/03 4:07:22 America/New_York” 这样的解析比较困难。“America/New_York” 这几个 bit （解析起来）比较麻烦。有一些 [变通的解决办法](https://stackoverflow.com/a/25368749)，不过 Golang 在解析 timezone 这部分还需要改进。

## Golang 之弊: 函数重载

我想起上计算机科学课时候的事，第一感觉就是“太酷了！”，但是在 Golang 里完全不存在这种“酷”。以下来自 Golang FAQ：

> 为什么 Go 不支持方法和操作符重载？
> 如果不用做类型匹配，方法调度过程也会简化。使用其他语言的经验告诉我们，一些有着同名但不同签名的函数有时候很有用，不过实际使用中也会造成误解，不够健壮。
> Go 类型系统主要使用的简化决策是只通过函数名匹配，并且要求类型的一致性。就操作符的重载来说，似乎比绝对要求（类型一致性）方便。 不过还是，没有重载会更简单。

当然，没有方法重载确实更简单。但这样也会强制你想出唯一的函数名以防出现同名函数，这样一来就增大了你的代码量。如果你只是想复制粘贴下代码，“以后再解”，还要记住共用函数是在私有的还是公共函数起作用。后面提到的 time 包有几个这样的实例，我们特别想用一个默认值替代最后那个可选参数。

```go
func Parse(layout, value string) (Time, error) {
	return parse(layout, value, UTC, Local)
}

func ParseInLocation(layout, value string, loc *Location) (Time, error) {
	return parse(layout, value, loc, loc)
}

func parse(layout, value string, defaultLocation, local *Location) (Time, error) {
	...
}
```

[[源代码地址]](https://golang.org/src/time/format.go?s=23626:23672#L762)

这就是说，没有方法重载绝不妨碍使用 Golang。在几个我想用（重载）的实例中，我最终只用了一个共用的私有函数，和其它几个不同名字的函数，或是删除了我一开始说需要重载的原因。

每隔一周左右，我们会发布一篇新指南，和本篇一样都在 “Golang 之于 DevOps 开发的利与弊” 六部曲中。下一篇：#5：跨平台编译， Windows ， Signals ，Docs 和编译器。

现在起 [订阅我们的博客吧](http://eepurl.com/cOHJ3f)，这样你就不会错过我们的文章发布了，如果你觉得你的开发者朋友们能受益于了解 Golang 的利与弊，请把这篇博文分享给他们，尤其是有关 DevOps 生命周期的。感谢您阅读和分享。

---

via: https://blog.bluematador.com/golang-pros-cons-part-4-time-package-method-overloading

作者：[Matthew Barlocker](https://github.com/mbarlocker)
译者：[ArisAries](https://github.com/ArisAries)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
