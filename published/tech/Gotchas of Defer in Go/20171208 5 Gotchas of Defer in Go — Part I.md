已发布：https://studygolang.com/articles/12061

# Go 中 defer 的 5 个坑 - 第一部分

> 通过本节的学习以避免掉入基础的 defer 陷阱中

本文只适合想要进阶学习 Golang 的新手阅读，大牛请绕道。

## #1 -- defer nil 函数

如果一个延迟函数被赋值为 `nil` ， 运行时的 [`panic`](https://golang.org/ref/spec#Handling_panics) 异常会发生在外围函数执行结束后而不是 `defer` 的函数被调用的时候。

例子

```go
func() {
	var run func() = nil
	defer run()

	fmt.Println("runs")
}
```

输出结果

```
runs

❗️ panic: runtime error: invalid memory address or nil pointer dereference
```

### 发生了什么？

名为 func 的函数一直运行至结束，然后 `defer` 函数会被执行且会因为值为 `nil` 而产生 `panic` 异常。然而值得注意的是，`run()` 的声明是没有问题，因为在外围函数运行完成后它才会被调用。

上面只是一个简单的案例，但同样的案例也可能发生在真实世界中，所以如果你遇上的话，可以想想是不是掉进了这个坑里。

## #2 -- 在循环中使用 defer

切忌在循环中使用 `defer`，除非你清楚自己在做什么，因为它们的执行结果常常会出人意料。

但是，在某些情况下，在循环中使用 `defer` 会相当方便，例如将函数中的递归转交给 `defer`，但这显然已经不是本文应该讲解的内容。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/5-gotchas-defer-1/defer_inside_a_loop.png)

在上面的例子中，`defer row.Close()` 在循环中的延迟函数会在函数结束过后运行，而不是每次 for 循环结束之后。这些延迟函数会不停地堆积到延迟调用栈中，最终可能会导致一些不可预知的问题。

### 解决方案 #1：

不使用 `defer` ，直接在末尾调用。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/5-gotchas-defer-1/solution_1.png)

### 解决方案 #2:

将任务转交给另一个函数然后在里面使用 `defer`，在下面这种情况下，延迟函数会在每次匿名函数执行结束后执行。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/5-gotchas-defer-1/solution_2.png)

### 进行基准测试

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/5-gotchas-defer-1/benchmark.jpg)

[查看代码](https://play.golang.org/p/GJ7oOMdBwJ)

## #3 -- 延迟调用含有闭包的函数

有时出于某种缘由，你想要让那些闭包延迟执行。例如，连接数据库，然后在查询语句执行过后中断与数据库的连接。

例子

```go
type database struct{}

func (db *database) connect() (disconnect func()) {
	fmt.Println("connect")

	return func() {
		fmt.Println("disconnect")
	}
}
```

运行一下

```go
db := &database{}
defer db.connect()
 
fmt.Println("query db...")
```

输出结果

```
query db...
connect
```

### 竟然出问题了？

最终 `disconnect` 并没有输出，最后只有 `connect` ，这是一个 bug，最终的情况是 `connect()` 执行结束后，其执行域得以被保存起来，但内部的闭包并不会被执行。

### 解决方案

```go
func() {
	db := &database{}
	close := db.connect()
	defer close()
 
	fmt.Println("query db...")
}
```

稍作修改后， `db.connect()` 返回了一个函数，然后我们再对这个函数使用 `defer` 就能够在 `func()` 执行结束后断开与数据库的连接。

输出结果

```
connect
query db...
disconnect
```

### 糟糕的处理方式：

即便这种处理方式很糟，但我还是想告诉你如何不用变量来解决这个问题，因此，我希望你能以此来了解 defer 亦或是 go 语言的运行机制。

```go
func() {
	db := &database{}
	defer db.connect()()

	..
}
```

这段代码从技术层面上说与上面的解决方案没有本质区别。其中，第一个圆括号是连接数据库（在 `defer db.connect()` 中立即执行的部分），然后第二个圆括号是为了在 `func()` 结束时延迟执行断开连接的函数(也就是返回的闭包)。

归因于 `db.connect()` 创建了一个闭包类型的值，然后再使用 `defer` 声明闭包函数， `db.connect()` 的值需要被实现计算出来以便让 `defer` 知道需要延迟哪个函数，这与 `defer` 不直接相关但也可能帮助你解决一些问题。

## #4 -- 在执行块中使用 defer

你可能想要在执行块执行结束后执行在块内延迟调用的函数，但事实并非如此，它们只会在块所属的函数执行结束后才被执行，这种情况适用于所有的代码块除了上文的函数块例如，for，switch 等。

**因为：延迟是相对于一个函数而非一个代码块**

例子

```go
func main() {
	{
		defer func() {
			fmt.Println("block: defer runs")
		}()

		fmt.Println("block: ends")
	}

	fmt.Println("main: ends")
}
```

输出结果

```
block: ends
main: ends
block: defer runs
```

上例的延迟函数只会在函数执行结束后运行，而不是紧接着它所在的块（花括号内包含 defer 调用的区域）后执行，就像代码中的演示的那样，你可以使用花括号创造单独的执行块。

### 另一个解决方案

如果你希望在另一个块中使用 `defer` ，可以使用匿名函数（正如在第二个坑中我们采用的解决方案）。

```go
func main() {
	func() {
		defer func() {
			fmt.Println("func: defer runs")
		}()

		fmt.Println("func: ends")
	}()

	fmt.Println("main: ends")
}
```

## #5 -- 延迟方法的坑

同样，你也可以使用 `defer` 来延迟 [方法](https://blog.learngoprogramming.com/go-functions-overview-anonymous-closures-higher-order-deferred-concurrent-6799008dde7b#61ec) 调用，但也可能出一些岔子。

### 没有使用指针作为接收者

```go
type Car struct {
	model string
}

func (c Car) PrintModel() {
	fmt.Println(c.model)
}

func main() {
	c := Car{model: "DeLorean DMC-12"}

	defer c.PrintModel()

	c.model = "Chevrolet Impala"
}
```

输出结果

```
DeLorean DMC-12
```

### 使用指针对象作为接收者

```go
func (c *Car) PrintModel() {
	fmt.Println(c.model)
}
```

输出结果

```
Chevrolet Impala
```

### 为什么会这样？

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/5-gotchas-defer-1/what_is_going_on.png)

我们需要记住的是，当外围函数还没有返回的时候，Go 的运行时就会立刻将传递给延迟函数的参数保存起来。

因此，当一个以值作为接收者的方法被 **defer** 修饰时，接收者会在声明时被拷贝（在这个例子中那就是 *Car* 对象），此时任何对拷贝的修改都将不可见（例中的 *Car.model* ），因为，接收者也同时是输入的参数，当使用  **defer** 修饰时会立刻得出参数的值(也就是 "DeLorean DMC-12" )。

在另一种情况下，当被延迟调用时，接收者为指针对象，此时虽然会产生新的指针变量，但其指向的地址依然与上例中的 "c" 指针的地址相同。因此，任何修改都会完美地作用在同一个对象中。

以上就是本文的全部内容，我会在后续的文章中补充更多类似的坑 -- 已经有至少 15 个易犯的 defer 错误榜上有名，如果你有任何想法，欢迎在下面留言。

----------------

via: https://blog.learngoprogramming.com/gotchas-of-defer-in-go-1-8d070894cb01

作者：[Inanc Gumus](https://blog.learngoprogramming.com/@inanc)
译者：[yujiahaol68](https://github.com/yujiahaol68)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出