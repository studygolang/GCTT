已发布：https://studygolang.com/articles/13158

# Go 语言中的初始化依赖项

让我们直接从两个 Go 语言小程序开始：

**程序1**

```go
package main
import "fmt"
var (
	a int = b + 1
	b int = 1
)
func main() {
	fmt.Println(a)
	fmt.Println(b)
}
```

**程序2**

```go
package main
import "fmt"
func main() {
	var (
		a int = b + 1
		b int = 1
	)
	fmt.Println(a)
	fmt.Println(b)
}
```
如果这两段代码会输出相同的结果，那么它们就不是好的素材，很幸运的是，它们的结果是不同的：

**程序1**

```
2
1
```

**程序2**

这个程序无法编译通过，甚至会在第 7 行报一个编译时错误 "undefined: b"。

究竟是什么带来了这种差异？变量声明中 "正常" 的初始化表达式，按照你所期望的从左到右和从上到下进行初始化：

```go
func f() int { fmt.Println("f"); return 1 }
func g() int { fmt.Println("g"); return 2 }
func h() int { fmt.Println("h"); return 3 }
func main() {
	var (
		a int = f()
		b int = g()
		c int = h()
	)
	fmt.Println(a, b, c)
}
```

输出：

```
f
g
h
1 2 3
```

前面提到的 "正常" 意味着它在自己的函数中完成初始化。当这些初始化代码像程序1中那样被放到包的顶层声明中时，会变得越来越有趣：

```go
package main
import "fmt"
var (
	a = c — 2
	b = 2
	c = f()
)
func f() int {
	fmt.Printf("inside f and b = %d\n", b)
	return b + 1
}
func main() {
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
}
```

这些变量声明的顺序如下：

* b 是第一个，因为它不依赖其他未初始化的变量
* c 是第二个，在 `f` 函数需要的变量 b 被初始化之后，紧接着被初始化
* a 是在 c 被初始化之后的第三轮初始化循环中被处理

程序的输出为：

```
inside f and b = 2
1
2
3
```

按照声明的顺序，每一轮初始化流程都选取第一个可以被初始化的变量。整个过程持续到所有变量都被完成初始化或者编译器找到一个类似于如下的循环：

```go
package main
import "fmt"
var (
	a = b
	b = c
	c = f()
)
func f() int {
	return a
}
func main() {
	fmt.Println(a, b, c)
}
```

上面的代码会导致编译时错误 "initialization loop"。

初始化依赖机制是基于包级别工作的:

**sandbox.go**

```go
package main
import "fmt"
var (
	a = c — 2
	b = 2
)
func main() {
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
}
```

**utils.go**

```go
package main
var c = f()
func f() int {
	return b + 1
}
```

编译及输出：

```
1
2
3
```

如果你喜欢上面的内容，请关注我，来加速未来故事的发展。

---

via: https://medium.com/golangspec/initialization-dependencies-in-go-51ae7b53f24c

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[rxcai](https://github.com/rxcai)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
