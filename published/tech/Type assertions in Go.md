已发布：https://studygolang.com/articles/12835

# Go 语言中的类型断言

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/type-assertion/1_p6c6i0niHNOIlRbsAhD3lA.jpeg)
<center>[https://en.wikipedia.org/wikiPsycho_(1960_film)](https://en.wikipedia.org/wiki/Psycho_%281960_film%29)</center>

类型断言被用于检查接口类型变量所持有的值是否实现了期望的接口或者具体的类型。

类型断言的语法定义如下：

```go
PrimaryExpression.(Type)
```

PrimaryExpression 可以在[ Go 语言规范](https://golang.org/ref/spec#PrimaryExpr)中找到，并且它可以是标识符，特定索引的数组元素，切片等等。

Type 既可以是类型标识符，也可以是类型字面量，比如：

```go
type I interface {
	walk()
	quack()
}

type S struct{}

func (s S) walk() {}
func (s S) quack() {}

var i I
i = S{}
fmt.Println(i.(interface {
	walk()
}))
```

PrimaryExpression 必须是接口类型，否则就会产生一个编译时错误：

```go
type I interface{
	walk()
	quack()
}

type S struct{}
S{}.(I) // 无效类型断言：S{}.(I)(操作符左边的 S 并不是个接口类型)
```

> 如果表达式为 nil，类型断言就不会成立。

## 动态类型

变量除了有静态类型外（变量声明中的类型），接口变量还有动态类型。就是在当前接口类型变量中设置的一种类型的值。在程序执行的过程当中，接口类型的变量具有相同的静态类型，但是其动态类型会随着其实现的接口不同，而其值也会随之改变。

```go
type I interface {
	walk()
}

type A struct{}

func (a A) walk() {}

type B struct{}

func (b B) walk() {}

func main() {
	var i I
	i = A{}  // i 的动态类型是 A
	fmt.Printf("%T\n", i.(A))
	i = B{}  // i 的动态类型是 B
	fmt.Printf("%T\n", i.(B))
}
```

## 接口类型

如果 T 来自 v.(T) 是一个接口类型，这样的断言检查，可以用来检测 v 的动态类型是否实现了接口 T：

```go
type I interfacce {
	walk()
}

type J interface {
	quack()
}

type K interface {
	bark()
}

type S struc{}

func (s S) walk() {}

func (s S) quack() {}

func main() {
	var i I
	i = S{}
	fmt.Printf("%T\n", i.(J))
	fmt.Printf("%T\n", i.(K))  // panic: 接口转换: main.S 不是 main.K: 缺少方法 bark
	}
```

## 非接口类型

如果 T 来自 v.(T) 不是接口类型，这样断言检查动态类型 v 是否与 T 类型相同：

```go
type I interface {
	walk()
}

type A struct{}

func (a A) walk() {}

type B struct{}

func (b B) walk() {}

func main() {
	var i I
	i = A{}
	fmt.Printf("%T\n", i.(A))
	fmt.Printf("%T\n", i.(B))  // panic: 接口转换: main.I 是 main.A, 不是 main.B
}
```

在非接口类型情况下进行类型传递就必须实现接口 I，如果不满足这个要求的话就会在编译时被捕获：

```go
type C struct{}
fmt.Prinf("%T\n", i.(C))
```

输出：

> impossible type assertion:  
> C does not implement I (missing walk method)

## 不要 panic

在上述情况下，当断言不能成立时，运行时 panic 将会被触发。为了优雅的处理错误，这里有特殊的形式来赋值或者初始化：

```go
type I interface {
	walk()
}
type A struct {
	name string
}
func (a A) walk() {}
type B struct {
	name string
}
func (b B) walk() {}
func main() {
	var i I
	i = A{name: "foo"}
	valA, okA := i.(A)
	fmt.Printf("%#v %#v\n", valA, okA)
	valB, okB := i.(B)
	fmt.Printf("%#v %#v\n", valB, okB)
}
```

输出：

```bash
main.A{name:"foo"} true
main.B{name:""} false
```

> 当断言不成立时，第一个值将会作为测试类型的[零值](https://golang.org/ref/spec#The_zero_value)

## 资源：
* [go 编程语言规范- go 编程语言](https://golang.org/ref/spec#TypeAssertion)
* [Go 是一个通用语言，设计时考虑了系统编程。它是强类型的并且具有垃圾回收机制...](https://golang.org/ref/spec#TypeAssertion)
* [golang.org](https://golang.org/ref/spec#TypeAssertion)

---

via：https://medium.com/golangspec/type-assertions-in-go-e609759c42e1

 作者：[Michał Łowicki](https://medium.com/@mlowicki)
 译者：[fredvence](https://github.com/fredvence)
 校对：[rxcai](https://github.com/rxcai)

 本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
