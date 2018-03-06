# 详解 Go 语言中的 `time.Duration` 类型

长久以来，我一直抓狂于 Go 标准库中的 Time 包，我的抓狂来自于两个功能，一是捕获两个不同时间段之间间隔的毫秒数，二是将一个用毫秒表示的连续时间段与预先定义的时间段进行比较。这听起来很简单，没错，确实如此，但它的确让我抓狂了。

在 Time 包中，定义有一个名为 Duration 的类型和一些辅助的常量：

```<go>
type Duration int64

const (
	 Nanosecond Duration = 1
	 Microsecond = 1000 * Nanosecond
	 Millisecond = 1000 * Microsecond
	 Second = 1000 * Millisecond
	 Minute = 60 * Second
	 Hour = 60 * Minute
) 
```

这些东西我可能已经看了有上千次了，但我的大脑依旧一片迷茫。我只是想比较两个时间段、恢复要持续的时间、比较持续时间的长短并且当预设的时间用完时做一些别的事情，但无论如何这个结构还是无法解决我的困扰。我写下了下面的测试代码，但它没有卵用：

```<go>
func Test() {
	var waitFiveHundredMillisections int64 = 500

	startingTime := time.Now().UTC()
	time.Sleep(10 * time.Millisecond)
	endingTime := time.Now().UTC()

	var duration time.Duration = endingTime.Sub(startingTime)
	var durationAsInt64 = int64(duration)

	if durationAsInt64 >= waitFiveHundredMillisections {
		fmt.Printf("Time Elapsed : Wait[%d] Duration[%d]\n", waitFiveHundredMillisections, durationAsInt64)
	} else {
		fmt.Printf("Time DID NOT Elapsed : Wait[%d] Duration[%d]\n", waitFiveHundredMillisections, durationAsInt64)
	}
} 
```

我运行了这段测试代码，然后得到了下面的输出，从输出内容来看，我定义的 500 毫秒的时间已经用完了，但怎么可能。

```
Time Elapsed : Wait[500] Duration[10724798] 
```

那么问题出在哪里？我又一次将眼光投向了 Duration 类型的定义：

```<go>
type Duration int64

const (
	Nanosecond Duration = 1
	Microsecond = 1000 * Nanosecond
	Millisecond = 1000 * Microsecond
	Second = 1000 * Millisecond
	Minute = 60 * Second
	Hour = 60 * Minute
) 
```

从代码上看， Duration 类型中时间的基本单位是 Nanosecond ，所以当我将一个表示 10 毫秒的 Duration 类型对象转换为 int64 类型时，我实际上得到的是 10,000,000。

所以直接转换是不行的，我需要一个不同的策略使用和转换 Duration 类型。

我知道最好使用 Duration 类型中定义的数据，这将最大可能地减少问题的发生。基于 Duration 中定义的常量，我能够像下面这样创建一个 Duration 变量：

```<go>
func Test() {
	var duration_Milliseconds time.Duration = 500 * time.Millisecond
	var duration_Seconds time.Duration = (1250 * 10) * time.Millisecond
	var duration_Minute time.Duration = 2 * time.Minute

	fmt.Printf("Milli [%v]\nSeconds [%v]\nMinute [%v]\n", duration_Milliseconds, duration_Seconds, duration_Minute)
}
```

在上面的代码中，我创建了 3 个 Duration 类型的变量，通过使用时间常数，我能够创建正确的持续时间的值。然后我使用标准库函数 `Printf` 和 `%v` 操作符，得到了下面的输出结果：

```
Milli [500ms]
Seconds [12.5s]
Minute [2m0s] 
```

很酷，有木有? `Printf` 函数知道如何本地化显示一个 Duration 类型，它基于 Duration 类型中的每一个值，选择合适的格式进行时间的显示，当然，我也得到了期待中的结果。

实际上， Duration 类型拥有一些便捷的类型转换函数，它们能将 Duration 类型转化为 Go 语言的内建类型 int64 或 float64 ，像下面这样：

```<go>
func Test() {
	var duration_Seconds time.Duration = (1250 * 10) * time.Millisecond
	var duration_Minute time.Duration = 2 * time.Minute

	var float64_Seconds float64 = duration_Seconds.Seconds()
	var float64_Minutes float64 = duration_Minute.Minutes()

	fmt.Printf("Seconds [%.3f]\nMinutes [%.2f]\n", float64_Seconds, float64_Minutes)
}
```

我也迅速注意到了在时间转换函数中，并没有转换毫秒值的函数，使用 Seconds 和 Minutes 函数，我得到了如下输出：

```
Seconds [12.500]
Minutes [2.00] 
```

但我需要转换毫秒值，为什么包里面单单没有提供毫秒值的转换呢？因为 Go 语言的设计者希望我有更多的选择，而不只是将毫秒值转换成某种单独的内建类型。下面的代码中，我将毫秒值转化为了 int64 类型和 float64 类型：

```<go>
func Test() {
	var duration_Milliseconds time.Duration = 500 * time.Millisecond

	var castToInt64 int64 = duration_Milliseconds.Nanoseconds() / 1e6
	var castToFloat64 float64 = duration_Milliseconds.Seconds() * 1e3
	fmt.Printf("Duration [%v]\ncastToInt64 [%d]\ncastToFloat64 [%.0f]\n", duration_Milliseconds, castToInt64, castToFloat64)
}
```

我将纳秒值除以 1^6 得到了 int64 类型表示的毫秒值，将秒值乘以 1^3 ，我得到了 float64 类型表示的毫秒值，上面代码的输出如下：

```
Duration [500ms]
castToInt64 [500]
castToFloat64 [500] 
```

现在，我知道了 Duration 类型是什么和怎么用，下面是我最终写的使用毫秒值的测试代码示例：

```<go>
func Test() {
	var waitFiveHundredMillisections time.Duration = 500 * time.Millisecond

	startingTime := time.Now().UTC()
	time.Sleep(600 * time.Millisecond)
	endingTime := time.Now().UTC()

	var duration time.Duration = endingTime.Sub(startingTime)

	if duration >= waitFiveHundredMillisections {
	fmt.Printf("Wait %v\nNative [%v]\nMilliseconds [%d]\nSeconds [%.3f]\n", waitFiveHundredMillisections,
			duration, duration.Nanoseconds()/1e6, duration.Seconds())
	}
}
```

得到的输出如下：

```
Wait 500ms
Native [601.091066ms]
Milliseconds [601]
Seconds [0.601] 
```

不再通过比较本地类型来确定时间是否已经用完，而是比较两个 Duration 类型变量，这样更加清晰明了。虽然花了一些时间，但最终我理解了 Duration 类型，我也希望这篇文章能帮助其他人在使用 Go 语言的过程中解决 Duration 类型的疑惑。

---

via: https://www.ardanlabs.com/blog/2013/06/gos-duration-type-unravelled.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[swardsman](https://github.com/swardsman)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
