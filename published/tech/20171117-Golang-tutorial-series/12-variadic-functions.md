已发布：https://studygolang.com/articles/12173

# 第 12 部分，可变参数函数

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)第 12 部分。

## 什么是可变参数函数

可变参数函数是一种参数个数可变的函数。

## 语法

如果函数最后一个参数被记作 `...T`，这时函数可以接受任意个 `T` 类型参数作为最后一个参数。

请注意只有函数的最后一个参数才允许是可变的。

## 通过一些例子理解可变参数函数如何工作

你是否曾经想过 append 函数是如何将任意个参数值加入到切片中的。这样 append 函数可以接受不同数量的参数。

```go
func append(slice []Type, elems ...Type) []Type
```

上面是 append 函数的定义。在定义中 elems 是可变参数。这样 append 函数可以接受可变化的参数。

让我们创建一个我们自己的可变参数函数。我们将写一段简单的程序，在输入的整数列表里查找某个整数是否存在。

```go
package main

import (
	"fmt"
)

func find(num int, nums ...int) {
	fmt.Printf("type of nums is %T\n", nums)
	found := false
	for i, v := range nums {
		if v == num {
			fmt.Println(num, "found at index", i, "in", nums)
			found = true
		}
	}
	if !found {
		fmt.Println(num, "not found in ", nums)
	}
	fmt.Printf("\n")
}
func main() {
	find(89, 89, 90, 95)
	find(45, 56, 67, 45, 90, 109)
	find(78, 38, 56, 98)
	find(87)
}
```
[在线运行代码](https://play.golang.org/p/7occymiS6s)

在上面程序中 `func find(num int, nums ...int)`  中的 `nums` 可接受任意数量的参数。在 find 函数中，参数 `nums` 相当于一个整型切片。

**可变参数函数的工作原理是把可变参数转换为一个新的切片。以上面程序中的第 22 行为例，`find` 函数中的可变参数是 89，90，95 。 find 函数接受一个 `int` 类型的可变参数。因此这三个参数被编译器转换为一个 int 类型切片 `int []int{89, 90, 95}` 然后被传入 `find` 函数。**

在第 10 行， `for` 循环遍历 `nums` 切片,如果 `num` 在切片中，则打印 `num` 的位置。如果 `num` 不在切片中,则打印提示未找到该数字。

上面代码的输出值如下,

```
type of nums is []int
89 found at index 0 in [89 90 95]

type of nums is []int
45 found at index 2 in [56 67 45 90 109]

type of nums is []int
78 not found in  [38 56 98]

type of nums is []int
87 not found in  []
```

在上面程序的第 25 行，find 函数仅有一个参数。我们没有给可变参数 `nums ...int` 传入任何参数。这也是合法的，在这种情况下 `nums` 是一个长度和容量为 0 的 `nil` 切片。

## 给可变参数函数传入切片

下面例子中，我们给可变参数函数传入一个切片，看看会发生什么。

```go
package main

import (
	"fmt"
)

func find(num int, nums ...int) {
	fmt.Printf("type of nums is %T\n", nums)
	found := false
	for i, v := range nums {
		if v == num {
			fmt.Println(num, "found at index", i, "in", nums)
			found = true
		}
	}
	if !found {
		fmt.Println(num, "not found in ", nums)
	}
	fmt.Printf("\n")
}
func main() {
	nums := []int{89, 90, 95}
	find(89, nums)
}

```
[在线运行代码](https://play.golang.org/p/7occymiS6s)

在第 23 行中，我们将一个切片传给一个可变参数函数。

这种情况下无法通过编译，编译器报出错误 `main.go:23: cannot use nums (type []int) as type int in argument to find` 。

为什么无法工作呢？原因很直接，`find` 函数的说明如下，

```go
func find(num int, nums ...int)
```

由可变参数函数的定义可知，`nums ...int` 意味它可以接受 `int` 类型的可变参数。

在上面程序的第 23 行，`nums` 作为可变参数传入 `find` 函数。前面我们知道，这些可变参数参数会被转换为 `int` 类型切片然后在传入 `find` 函数中。但是在这里 `nums` 已经是一个 int 类型切片，编译器试图在 `nums` 基础上再创建一个切片，像下面这样

```go
find(89, []int{nums})
```

这里之所以会失败是因为 `nums` 是一个 `[]int` 类型 而不是 `int` 类型。

那么有没有办法给可变参数函数传入切片参数呢？答案是肯定的。

**有一个可以直接将切片传入可变参数函数的语法糖，你可以在在切片后加上 `...` 后缀。如果这样做，切片将直接传入函数，不再创建新的切片**

在上面的程序中，如果你将第 23 行的 `find(89, nums)` 替换为 `find(89, nums...)` ，程序将成功编译并有如下输出

```go
type of nums is []int
89 found at index 0 in [89 90 95]
```
下面是完整的程序供您参考。

```go
package main

import (
	"fmt"
)

func find(num int, nums ...int) {
	fmt.Printf("type of nums is %T\n", nums)
	found := false
	for i, v := range nums {
		if v == num {
			fmt.Println(num, "found at index", i, "in", nums)
			found = true
		}
	}
	if !found {
		fmt.Println(num, "not found in ", nums)
	}
	fmt.Printf("\n")
}
func main() {
	nums := []int{89, 90, 95}
	find(89, nums...)
}
```
[在线运行代码](https://play.golang.org/p/7occymiS6s)

## 不直观的错误

当你修改可变参数函数中的切片时，请确保你知道你正在做什么。

下面让我们来看一个简单的例子。

```go
package main

import (
	"fmt"
)

func change(s ...string) {
	s[0] = "Go"
}

func main() {
	welcome := []string{"hello", "world"}
	change(welcome...)
	fmt.Println(welcome)
}
```
[在线运行代码](https://play.golang.org/p/7occymiS6s)

你认为这段代码将输出什么呢？如果你认为它输出 `[Go world]` 。恭喜你！你已经理解了可变参数函数和切片。如果你猜错了，那也不要紧，让我来解释下为什么会有这样的输出。

在第 13 行，我们使用了语法糖 `...` 并且将切片作为可变参数传入 `change` 函数。

正如前面我们所讨论的，如果使用了 `...` ，`welcome` 切片本身会作为参数直接传入，不需要再创建一个新的切片。这样参数 `welcome` 将作为参数传入 `change` 函数

在 `change` 函数中，切片的第一个元素被替换成 `Go`，这样程序产生了下面的输出值

```
[Go world]
```

这里还有一个例子来理解可变参数函数。

```go
package main

import (
	"fmt"
)

func change(s ...string) {
	s[0] = "Go"
	s = append(s, "playground")
	fmt.Println(s)
}

func main() {
	welcome := []string{"hello", "world"}
	change(welcome...)
	fmt.Println(welcome)
}
```
[在线运行代码](https://play.golang.org/p/7occymiS6s)

我将把它作为一个练习留个你，请你指出上面的程序是如何运行的 :) 。

以上就是关于可变参数函数的介绍。感谢阅读。欢迎您留下有价值的反馈和意见。祝您生活愉快。

**上一教程 - [Array 和 Slice](https://studygolang.com/articles/12121)**

**下一教程 - [Maps](https://studygolang.com/articles/12251)**

---

via: https://golangbot.com/variadic-functions/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[fengchunsgit](https://github.com/fengchunsgit)
校对：[Noluye](https://github.com/Noluye)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
