已发布：https://studygolang.com/articles/12632

# Go 的语句块

声明会给标识符绑定值，例如包、变量、类型等等。完成声明后，有必要知道源码中的标识符在哪些地方引用了被指定的值（简单来讲，就是一个标识符在哪里是可用的）。

Go 属于词法作用域，所以标识符的解析会依赖于它在代码中声明的位置。这种方式和动态作用域语言截然不同，动态作用域语言中标识符的可见性不依赖于被声明的位置。看看下面这段 bash 脚本：

```bash
#!/bin/bash
f() {
	local v=1
	g
}
g() {
	echo "g sees v as $v"
}
h() {
	echo "h sees v as $v"
}
f
g
h
```

变量 v 是在函数 f 中定义的，但由于函数 g 被函数 f 调用，所以函数 g 可以访问 v：

```
> ./scope.sh
g sees v as 1
g sees v as
h sees v as
```

当单独调用函数 g 或者在函数 h 中调用 g 时，可以看到 v 并没有被定义。动态作用域语言中的可见性并不是静态不变（词法作用域也被叫做静态作用域），而是依赖于控制流。

当试图编译 Go 版本的类似代码时会 报编译错误：

```go
package main
import "fmt"
func f() {
	v := 1
	g()
}
func g() {
	fmt.Println(v)  // "undefined: v"
}
func main() {
	f()
}
```

Go 的词法作用域使用了语句块，所以在学习可见性规则前有必要先理解什么是语句块。

语句块是一连串的语句序列（空序列也算）。语句块可以嵌套，并且被花括号标识出来。

```go
package main
import "fmt"
func main() {
	{ // start outer block
		a := 1
		fmt.Println(a)
		{ // start inner block
			b := 2
			fmt.Println(b)
		} // end inner block
	} // end outer block
}
```

除了显示标出的语句块之外，还有一些隐式语句块：

- 主语句块：包括所有源码,
- 包语句块：包括该包中所有的源码（一个包可能会包括一个目录下的多个文件），
- 文件语句块：包括该文件中的所有源码,
- for 语句本身也在它自身的隐式语句块中:

```go
for i := 0; i < 5; i++ {
	fmt.Println(i)
}
```

所以在初始化语句中声明的变量 i 可以在整个循环体的条件语句，后置语句以及嵌套块中访问。但在 for 语句之后使用 i 的操作会引起 “未定义：i ” 编译错误。

- if 语句也是它自身的隐式语句块：

```go
if i := 0; i >= 0 {
	fmt.Println(i)
}
```

if 语句允许声明变量，当条件为真时该变量可以在嵌套块中使用，抑或条件为假时在 else 语句块中使用。

- switch 语句在它自身的隐式语句块中：

```go
switch i := 2; i * 4 {
case 8:
	fmt.Println(i)
default:
	fmt.Println(“default”)
}
```

​和 if 语句类似，可以使用临时声明变量作为 case 子句的开头。

- 每个 switch 语句中的子句都是一个隐式语句块。

```go
switch i := 2; i * 4 {
case 8:
	j := 0
	fmt.Println(i, j)
default:
	// "j" is undefined here
	fmt.Println(“default”)
}
// "j" is undefined here
```

如果语法规定每个 case 子句都属于同一个语法块的话，那么就不会需要这个额外的例子。这将迫使每个子句都要使用花括号，使得代码变得不那么易读和简洁。

- select 语句中的每个子句都是一个隐式语句块，和 switch 语句中的子句类似:

```go
tick := time.Tick(100 * time.Millisecond)
LOOP:
	for {
		select {
		case <-tick:
			i := 0
			fmt.Println(“tick”, i)
			break LOOP
		default:
			// "i" is undefined here
			fmt.Println(“sleep”)
			time.Sleep(30 * time.Millisecond)
		}
	}
	// "i" is undefined here
```

[“Scope in Go”](https://medium.com/@mlowicki/scopes-in-go-a6042bb4298c) 解释了作用域（可见性）。语句块在定义作用域的整个机制中起着关键作用。

## 资源

- https://golang.org/ref/spec#Blocks

---

via: https://medium.com/golangspec/blocks-in-go-2f68768868f6

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[sunzhaohao](https://github.com/sunzhaohao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
