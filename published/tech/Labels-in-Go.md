已发布：https://studygolang.com/articles/12906

# Go 中的标签

`Label` 在 `break` 和 `continue` 语句中是可选参数，而在 `goto` 语句中是必传的参数。`Label` 只在声明它的函数中有效。只要函数中声明了 `Label` ，那它在该函数的整个作用域都有效。

```go
func main() {
	fmt.Println(1)
	goto End
	fmt.Println(2)
End:
	fmt.Println(3)
}
```

（注意！我们是在 goto 语句之后定义的 `Label`）

```shell
> ./bin/sandbox
1
3
```

这里所指的作用域不包括嵌套函数的作用域：

```go
	func() {
		fmt.Println(“Nested function”)
		goto End
	}()
End:

```

上面这段代码在编译时会报 `label End not defined` 错误。除此之外还会报：`label End defined and not used`，因为我们需要保证只声明需要使用的 `Label`。

`Label` 中没有块作用域的概念。所以不能在嵌套的块作用域中声明重复的 `Label`：

```go
	goto X
X:
	{
		X:
	}
```

编译上面的代码会报错提示 `Label` 已经存在了。

`Label` 的标识符和其他标识符具有不同的命名空间，所以不会与变量标识符等发生冲突。下面这段代码同时在 `Label` 和变量声明中都使用 `x` 作为标识符：

```go
	x := 1
	goto x
x:
	fmt.Println(x)
```

## break 语句

`break` 语句一般都是用来跳出 `for` 或 `switch` 语句。在 Go 中也可以在 `select` 语句中使用 `break`。

对于拥有类 C 语言编程经验的人肯定知道每个 `case` 语句都需要以 `break` 结尾。但是在大多数情况下程序都不希望执行到下一个 `case` 语句中。在 Go 中恰恰相反，如果需要执行下一个分句的代码可以使用 `fallthrough` 语句：

```
switch 1 {
case 1:
	fmt.Println(1)
case 2:
	fmt.Println(2)
}
> ./bin/sandbox
1
switch 1 {
case 1:
	fmt.Println(1)
	fallthrough
case 2:
	fmt.Println(2)
}
> ./bin/sandbox
1
2
```

> fallthrough 只能作为一个分句的最后一条语句使用。

`break` 语句不能跨越函数边界。

```go
func f() {
	break
}
func main() {
	for i := 0; i < 10; i++ {
		f()
	}
}
```

上面的代码会报 `break is not in a loop` 错误。

一般情况下，`break` 可以用来跳出 `for` ，`switch` 或 `select` 语句。使用 `Label` 让 `break` 可以做更多的事情。

`break` 语句的 `Label` 必须对应围绕 `break` 的 `for`，`switch` 或 `select` 语句。所以下面的代码不能通过编译：

```go
FirstLoop:
	for i := 0; i < 10; i++ {
	}
	for i := 0; i < 10; i++ {
		break FirstLoop
	}
```

它会报 `invalid break label FirstLoop` 。因为 `Label` 不在 break 对应的循环上。

break 可以跳出嵌套的循环：

```go
OuterLoop:
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			fmt.Printf(“i=%v, j=%v\n”, i, j)
			break OuterLoop
		}
	}
```

```
> ./bin/sandbox
i=0, j=0
```

如果 `break` 像下面这样用，那它不仅仅是退出 `for` 循环：

```go
SwitchStatement:
	switch 1 {
	case 1:
		fmt.Println(1)
		for i := 0; i < 10; i++ {
			break SwitchStatement
		}
		fmt.Println(2)
	}
	fmt.Println(3)
```

```
> ./bin/sandbox
1
3
```

`break` 可以在 `for`，`switch` 或 `select` 语句中的任何地方退出并切换到指定 `Label`。

## continue 语句

他和 `break` 语句相似，只是会进入下一次迭代而不是退出循环（只能在循环中使用）：

```
OuterLoop:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			fmt.Printf(“i=%v, j=%v\n”, i, j)
			continue OuterLoop
		}
	}
> ./bin/sandbox
i=0, j=0
i=1, j=0
i=2, j=0
```

## goto 语句

`goto` 语句可以让程序切换到某个被 `Label` 标记的地方继续执行。

```
	i := 0
Start:
	fmt.Println(i)
	if i > 2 {
		goto End
	} else {
		i += 1
		goto Start
	}
End:
```

（注意：`End` 可以为空）

与 `break` 和 `continue` 语句不同，`goto` 语句必须指定 `Label`。

`goto` 只能让程序移动到相同函数的某个位置。因此在执行 goto 的时候还有另外两条规则：

- 不能跳过变量声明，如果 goto 跳过，编译器会报 `goto Done jumps over declaration of v at…`

```go
	goto Done
	v := 0
Done:
	fmt.Println(v)
```

- `goto` 不能切换到其他代码块，如果切换，编译器会报 `goto Block jumps into block starting at …`：

```go
goto Block
{
Block:
	v := 0
	fmt.Println(v)
}
```

---

via: https://medium.com/golangspec/labels-in-go-4ffd81932339

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[saberuster](https://github.com/saberuster)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
