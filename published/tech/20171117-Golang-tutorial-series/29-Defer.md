已发布：https://studygolang.com/articles/12719

# 第 29 篇：Defer

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 29 篇。

## 什么是 defer？

`defer` 语句的用途是：含有 `defer` 语句的函数，会在该函数将要返回之前，调用另一个函数。这个定义可能看起来很复杂，我们通过一个示例就很容易明白了。

## 示例

```go
package main

import (
	"fmt"
)

func finished() {
	fmt.Println("Finished finding largest")
}

func largest(nums []int) {
	defer finished()
	fmt.Println("Started finding largest")
	max := nums[0]
	for _, v := range nums {
		if v > max {
			max = v
		}
	}
	fmt.Println("Largest number in", nums, "is", max)
}

func main() {
	nums := []int{78, 109, 2, 563, 300}
	largest(nums)
}
```

[在 playground 上运行](https://play.golang.org/p/IlccOsuSUE)

上面的程序很简单，就是找出一个给定切片的最大值。`largest` 函数接收一个 int 类型的[切片](https://studygolang.com/articles/12121)作为参数，然后打印出该切片中的最大值。`largest` 函数的第一行的语句为 `defer finished()`。这表示在 `finished()` 函数将要返回之前，会调用 `finished()` 函数。运行该程序，你会看到有如下输出：

```
Started finding largest
Largest number in [78 109 2 563 300] is 563
Finished finding largest
```

`largest` 函数开始执行后，会打印上面的两行输出。而就在 `largest` 将要返回的时候，又调用了我们的延迟函数（Deferred Function），打印出 `Finished finding largest` 的文本。:)

## 延迟方法

`defer` 不仅限于[函数](https://studygolang.com/articles/11892)的调用，调用[方法](https://studygolang.com/articles/12264)也是合法的。我们写一个小程序来测试吧。

```go
package main

import (
	"fmt"
)


type person struct {
	firstName string
	lastName string
}

func (p person) fullName() {
	fmt.Printf("%s %s",p.firstName,p.lastName)
}

func main() {
	p := person {
		firstName: "John",
		lastName: "Smith",
	}
	defer p.fullName()
	fmt.Printf("Welcome ")
}
```

[在 playground 上运行](https://play.golang.org/p/lZ74OAwnRD)

在上面的例子中，我们在第 22 行延迟了一个方法调用。而其他的代码很直观，这里不再解释。该程序输出：

```
Welcome John Smith
```

## 实参取值（Arguments Evaluation）

在 Go 语言中，并非在调用延迟函数的时候才确定实参，而是当执行 `defer` 语句的时候，就会对延迟函数的实参进行求值。

通过一个例子就能够理解了。

```go
package main

import (
	"fmt"
)

func printA(a int) {
	fmt.Println("value of a in deferred function", a)
}
func main() {
	a := 5
	defer printA(a)
	a = 10
	fmt.Println("value of a before deferred function call", a)

}
```

[在 playground 上运行](https://play.golang.org/p/sBnwrUgObd)

在上面的程序里的第 11 行，`a` 的初始值为 5。在第 12 行执行 `defer` 语句的时候，由于 `a` 等于 5，因此延迟函数 `printA` 的实参也等于 5。接着我们在第 13 行将 `a` 的值修改为 10。下一行会打印出 `a` 的值。该程序输出：

```
value of a before deferred function call 10
value of a in deferred function 5
```

从上面的输出，我们可以看出，在调用了 `defer` 语句后，虽然我们将 `a` 修改为 10，但调用延迟函数 `printA(a)`后，仍然打印的是 5。

## defer 栈

当一个函数内多次调用 `defer` 时，Go 会把 `defer` 调用放入到一个栈中，随后按照后进先出（Last In First Out, LIFO）的顺序执行。

我们下面编写一个小程序，使用 `defer` 栈，将一个字符串逆序打印。

```go
package main

import (
	"fmt"
)

func main() {
	name := "Naveen"
	fmt.Printf("Orignal String: %s\n", string(name))
	fmt.Printf("Reversed String: ")
	for _, v := range []rune(name) {
		defer fmt.Printf("%c", v)
	}
}
```

[在 playground 上运行](https://play.golang.org/p/HDk623ozuw)

在上述程序中的第 11 行，`for range` 循环会遍历一个字符串，并在第 12 行调用了 `defer fmt.Printf("%c", v)`。这些延迟调用会添加到一个栈中，按照后进先出的顺序执行，因此，该字符串会逆序打印出来。该程序会输出：

```
Orignal String: Naveen
Reversed String: neevaN
```

## defer 的实际应用

目前为止，我们看到的代码示例，都没有体现出 `defer` 的实际用途。本节我们会看看 `defer` 的实际应用。

当一个函数应该在与当前代码流（Code Flow）无关的环境下调用时，可以使用 `defer`。我们通过一个用到了 [`WaitGroup`](https://studygolang.com/articles/12512) 代码示例来理解这句话的含义。我们首先会写一个没有使用 `defer` 的程序，然后我们会用 `defer` 来修改，看到 `defer` 带来的好处。

```go
package main

import (
	"fmt"
	"sync"
)

type rect struct {
	length int
	width  int
}

func (r rect) area(wg *sync.WaitGroup) {
	if r.length < 0 {
		fmt.Printf("rect %v's length should be greater than zero\n", r)
		wg.Done()
		return
	}
	if r.width < 0 {
		fmt.Printf("rect %v's width should be greater than zero\n", r)
		wg.Done()
		return
	}
	area := r.length * r.width
	fmt.Printf("rect %v's area %d\n", r, area)
	wg.Done()
}

func main() {
	var wg sync.WaitGroup
	r1 := rect{-67, 89}
	r2 := rect{5, -67}
	r3 := rect{8, 9}
	rects := []rect{r1, r2, r3}
	for _, v := range rects {
		wg.Add(1)
		go v.area(&wg)
	}
	wg.Wait()
	fmt.Println("All go routines finished executing")
}
```

[在 playground 上运行](https://play.golang.org/p/kXL85U0Dd_)

在上面的程序里，我们在第 8 行创建了 `rect` 结构体，并在第 13 行创建了 `rect` 的方法 `area`，计算出矩形的面积。`area` 检查了矩形的长宽是否小于零。如果矩形的长宽小于零，它会打印出对应的提示信息，而如果大于零，它会打印出矩形的面积。

`main` 函数创建了 3 个 `rect` 类型的变量：`r1`、`r2` 和 `r3`。在第 34 行，我们把这 3 个变量添加到了 `rects` 切片里。该切片接着使用 `for range` 循环遍历，把 `area` 方法作为一个并发的 Go 协程进行调用（第 37 行）。我们用 `WaitGroup wg` 来确保 `main` 函数在其他协程执行完毕之后，才会结束执行。`WaitGroup` 作为参数传递给 `area` 方法后，在第 16 行、第 21 行和第 26 行通知 `main` 函数，表示现在协程已经完成所有任务。**如果你仔细观察，会发现 `wg.Done()` 只在 `area` 函数返回的时候才会调用。`wg.Done()` 应该在 `area` 将要返回之前调用，并且与代码流的路径（Path）无关，因此我们可以只调用一次 `defer`，来有效地替换掉 `wg.Done()` 的多次调用**。

我们来用 `defer` 来重写上面的代码。

在下面的代码中，我们移除了原先程序中的 3 个 `wg.Done` 的调用，而是用一个单独的 `defer wg.Done()` 来取代它（第 14 行）。这使得我们的代码更加简洁易懂。

```go
package main

import (
	"fmt"
	"sync"
)

type rect struct {
	length int
	width  int
}

func (r rect) area(wg *sync.WaitGroup) {
	defer wg.Done()
	if r.length < 0 {
		fmt.Printf("rect %v's length should be greater than zero\n", r)
		return
	}
	if r.width < 0 {
		fmt.Printf("rect %v's width should be greater than zero\n", r)
		return
	}
	area := r.length * r.width
	fmt.Printf("rect %v's area %d\n", r, area)
}

func main() {
	var wg sync.WaitGroup
	r1 := rect{-67, 89}
	r2 := rect{5, -67}
	r3 := rect{8, 9}
	rects := []rect{r1, r2, r3}
	for _, v := range rects {
		wg.Add(1)
		go v.area(&wg)
	}
	wg.Wait()
	fmt.Println("All go routines finished executing")
}
```

[在 playground 上运行](https://play.golang.org/p/JuUvytLfBv)

该程序会输出：

```
rect {8 9}'s area 72
rect {-67 89}'s length should be greater than zero
rect {5 -67}'s width should be greater than zero
All go routines finished executing
```

在上面的程序中，使用 `defer` 还有一个好处。假设我们使用 `if` 条件语句，又给 `area` 方法添加了一条返回路径（Return Path）。如果没有使用 `defer` 来调用 `wg.Done()`，我们就得很小心了，确保在这条新添的返回路径里调用了 `wg.Done()`。由于现在我们延迟调用了 `wg.Done()`，因此无需再为这条新的返回路径添加 `wg.Done()` 了。

本教程到此结束。祝你愉快。

**上一教程 - [多态](https://studygolang.com/articles/12681)**

**下一教程 - [错误处理](https://studygolang.com/articles/12724)**

---

via: https://golangbot.com/defer/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
