已发布：https://studygolang.com/articles/12791

# 使用 defer 时可能遇到的若干陷阱

go 的 defer 语句对改善代码可读性起了很大作用。但是，某些情况下 defer 的行为很容易引起混淆，并且难以快速理清。尽管作者已经使用 go 两年多了，依然会被 defer 弄得挠头不已。我的计划是把过去曾困惑过我的一系列行为汇总起来，作为对自己的警示。

## defer 的作用域是一个函数，不是一个语句块

一个变量只存在于一个语句块的作用域内。 defer 语句所在的语句块只会在函数返回时执行。我不清楚背后的原理，但是，如果你在一个循环里面先分配资源，再用 defer 回收资源，就有可能带来猝不及防的灾难后果。

```go
func do(files []string) error {
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close() // 这是错误用法!!
		// 使用 f
	}
}
```

（译者注：上述代码会造成循环结束后才开始回收资源，而不是执行了一次循环就回收一次资源）

## 方法链

如果你在一个 defer 语句中链式调用方法，那么除了最后一个函数以外其余函数都会在调用时直接执行。 defer 要求一个函数作为 “参数” 。

```go
type logger struct {}
func (l *logger) Print(s string) {
	fmt.Printf("Log: %v\n", s)
}
type foo struct {
	l *logger
}
func (f *foo) Logger() *logger {
	fmt.Println("Logger()")
	return f.l
}
func do(f *foo) {
	defer f.Logger().Print("done")
	fmt.Println("do")
}

func main() {
	f := &foo{
		l: &logger{},
	}
	do(f)
}
```

输出结果——

```
Logger()
do
Log: done
```

Logger() 函数在 do() 函数之前就已经执行了。

## 函数参数

嗯，那么如果最后一个函数接收了一个参数，结果又会怎么样呢？ ? 按照常理来说 ，如果它是在外围函数返回后才执行，那么对变量的任何修改都会被捕获到，事实真的如我们所料么？

```go
type logger struct {}
func (l *logger) Print(err error) {
	fmt.Printf("Log: %v\n", err)
}
type foo struct {
	l *logger
}
func (f *foo) Logger() *logger {
	fmt.Println("Logger()")
	return f.l
}
func do(f *foo) (err error) {
	defer f.Logger().Print(err)
	fmt.Println("do")
	return fmt.Errorf("ERROR")
}

func main() {
	f := &foo{
		l: &logger{},
	}
	do(f)
}
```

猜猜输出结果是？

```
Logger()
do
Log: <nil>
```

err 的值仍然是调用 defer 时候的值。任何对这个变量的修改都不会被 defer 语句捕获，因为它们并不指向同一个值。

## 针对非指针类型调用函数

我们已经看到了 defer 语句中链式方法的执行特点。进一步深入下去，如果被调用的方法并不是定义在一个指针类型上，那么将会在 defer 语句中复制出一个新的实例。

```go
type metrics struct {
	success bool
	latency time.Duration
}
func (m metrics) Log() {
	fmt.Printf("Success: %v, Latency: %v\n", m.success, m.latency)
}
func foo() {
	var m metrics
	defer m.Log()
	start := time.Now()
	// Do something
	time.Sleep(2*time.Second)

	m.success = true
	m.latency = time.Now().Sub(start)
}
```

输出结果

```
Success: false, Latency: 0s
```

当 defer 语句执行时 m 被复制。 m.Foo() 实质是 Foo(m) 的简写形式。

## 结论

如果你使用 go 的时间足够久，那么这些可能都算不上 “陷阱” 。但对新手来说， defer 语句的很多地方都不符合 [最少吃惊原则](https://en.wikipedia.org/wiki/Principle_of_least_astonishment) 。还有 [更](http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/) [多](https://studygolang.com/articles/12061) [地](https://studygolang.com/articles/12136) [方](https://studygolang.com/articles/12319) 深入研究了使用 go 时可能遇到的常见失误。欢迎阅读。

---

via: https://medium.com/@i0exception/some-common-traps-while-using-defer-205ebbdc0a3b

作者：[Aniruddha](https://medium.com/@i0exception)
译者：[sunzhaohao](https://github.com/sunzhaohao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
