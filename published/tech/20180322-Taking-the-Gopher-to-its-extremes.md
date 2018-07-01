已发布：https://studygolang.com/articles/13382

# Go 功能：将 Go 程序员带向极限（Functional Go: Taking The Gopher to it's extremes）

## 功能范式（The Functional Paradigm）

函数式编程基于简单的规则，目的是使程序保持一致，使并行（更）人性化和简单化，函数是存粹的，它不会改变（变量的）状态，不会改变对象，也不共享状态，代码是声明式的，计算只不过是数学函数的一种实现。

我们来看一个纯粹的函数语言的函数例子：

```
isPalindrome :: (Eq a) => [a] -> Bool
isPalindrome x = x == reverse x
```

上面的代码是 haskell 版本的 palindrome 函数，我们（实际上）都在某种程度上通过循环来遍历列表，以及检查...，在 Haskell 中，isPalindrome 函数获取 Equality 的字符列表（考虑比较时），并返回一个 bool 值（True 或者 False）。实现（这个功能）只需一行代码：

```
x == reverse x
```

将参数和 reverse 输出进行比较，reverse 是一个反转列表的函数。

我们来看如何在 Golang 中实现一个相似的函数，从一个递归版本的 reverse 开始吧：

```go
func reverse(str string) string {
	if str == "" {
		return str
	} else {
		return reverse(str[1:]) + string(str[0]) // a string is a byte array in golang
	}
}
```

现在我们有了一个名为 reverse 的函数，让我们再看一下如何实现 isPalindrome ，用类似于函数的方式。

```go
func isPalindrome (str string) {
	return str == reverse(str)
}
```

就像你看到的，我们最终用几行代码就写了一个大家一看就懂的函数，在大多数情况下，非常有名的函数都可以实现 go 语言的版本，函数在 golang 中是第一公民，所以我们可以做一些类似下面的实现：

```go
isPalindrome := func (str string) {
	return str == reverse(str)
}

isPalindrome("radar")
```

这只是 golang 功能中很小的一部分，还有很多 go 语言实现中有大量优化的地方的例子，例如 go 中的 map，reduce，和 filter，一个好的实践方法是优化它们，并使用它们。

例如，数据处理在 golang 中的速度要比 python 快得多，所以，（虽然）我（都）可以创建一个数据管道来清理和组织我的数据，但如果我愿意用 golang 实现的话，将更快。

假设我有一个数值列表，我想要用一种优雅的方式，通过一组函数来处理这个列表中的数值：

```go
type p func(int) int


func apInt(functions []p, numbers []int) []int{
	j := 0
	output := make([]int,0)
	for _,f := range functions {
		for j < len(numbers) {
			fmt.Println(f(numbers[j]))
			output = append(output,f(numbers[j]))
			j++
		}
	}
	return output
}
```

上面的 functions，是一个类似于如下声明的函数数组：

```go
	listOfFuncs := []p{a, b}
```

它能够在所有的整型切片中工作，并且没有副作用。

我刚为 [lori](https://github.com/radicalrafi/lori) 工作，它是一个 Golang 库，目标是为开发者提供这些有用的东西，并使 function 变得有趣和可能。

---

via: https://radicalrafi.github.io/posts/functional-go/

作者：[radicalrafi](https://github.com/radicalrafi)
译者：[gogeof](https://github.com/gogeof)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
