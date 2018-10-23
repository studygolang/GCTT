# Go语言中的接口(part II)
## 类型断言 & 类型转换

有些时候，我们需要将数值转换成不同的类型。在编译的时候会进行类型转换的检查，整个机制已经在 [以前的文章](https://medium.com/golangspec/conversions-in-go-4301e8d84067) 中讲过。简而言之它就像下面这样（[源代码](https://play.golang.org/p/ogrrsqU6IZ)）：

```

	type T1 struct {
		name string
	}

	type T2 struct {
    	name string
	}

	func main() {
    	vs := []interface{}{T2(T1{"foo"}), string(322), []byte("abł")}
    	for _, v := range vs {
       	 fmt.Printf("%v %T\n", v, v)
    	}
	}

```

输出：

```

	{foo} main.T2
	ł string
	[97 98 197 130] []uint8

```

Golang 有可转换的规则，一些特定的情况下允许赋值给另一种类型值的变量（[源代码](https://play.golang.org/p/EgfTv4kc37)）：

```

	type T struct {
	    name string
	}

	func main() {
	    v1 := struct{ name string }{"foo"}
	    fmt.Printf("%T\n", v1) // struct { name string }
	    var v2 T
	    v2 = v1
	    fmt.Printf("%T\n", v2) // main.T
	}

```

文章将重点来讲当涉及到接口类型时发生的这些转换。此外我们将引入新的结构——类型断言和类型转换。

假设我们有两种不同接口类型的变量，接着我们将其中一个赋值给另外一个（[源代码](https://play.golang.org/p/TLhQW5SkZU)）：

```

	type I1 interface {
	    M1()
	}

	type I2 interface {
	    M1()
	}

	type T struct{}

	func (T) M1() {}

	func main() {
	    var v1 I1 = T{}
	    var v2 I2 = v1
	    _ = v2
	}

```

这很容易，因为程序运行得很好。第三种 [可转换](https://golang.org/ref/spec#Assignability) 的情况在这里：

　　　　<font size=3>***T 是一个接口并且 x 实现了接口 T***</font>

这是因为当 *v1* 的类型实现了 *I2* 接口后，这些变量构造的时候是什么类型已经无所谓了（[源代码](https://play.golang.org/p/DC76FG4MOq)）:

```

	type I1 interface {
	    M1()
	    M2()
	}

	type I2 interface {
	    M1()
	    I3
	}

	type I3 interface {
	    M2()
	}

	type T struct{}

	func (T) M1() {}
	func (T) M2() {}

	func main() {
	    var v1 I1 = T{}
	    var v2 I2 = v1
	    _ = v2
	}

```

即使 *I2* 嵌套了 *I1* 没有嵌套的接口，依然可以互相转换。方法实现的顺序不重要。一点值得记住的就是方法集不需要相同（[源代码](https://play.golang.org/p/wTOHwxR_ve)）：

```

	type I1 interface {
	    M1()
	    M2()
	}

	type I2 interface {
	    M1()
	}

	type T struct{}

	func (T) M1() {}
	func (T) M2() {}

	func main() {
	    var v1 I1 = T{}
	    var v2 I2 = v1
	    _ = v2
	}

```

这样的代码运行得很好是因为满足了第 3 条可转换的情况。*I2* 类型的值能用 *I1* 赋值是因为它的方法集是 *I1* 方法集的一个子集。如果不是这种情况，那么编译器就会给出相应的反馈（[源代码](https://play.golang.org/p/u9CE_sQ32H)）：

```

	type I1 interface {
	    M1()
	}

	type I2 interface {
	    M1()
	    M2()
	}

	type T struct{}

	func (T) M1() {}

	func main() {
	    var v1 I1 = T{}
	    var v2 I2 = v1
	    _ = v2
	}

```

上面的这段代码就无法编译通过，因为这个错误：

```

	main.go:18: cannot use v1 (type I1) as type I2 in assignment:
		I1 does not implement I2 (missing M2 method)

```

我们已经看到了涉及两种接口类型的情况。 当右侧值为具体类型（非接口类型）并实现接口时，前面列出的第3种可转换性也适用（[源代码](https://play.golang.org/p/6TmShsVao5)）：

```

	type I1 interface {
	    M1()
	}

	type T struct{}

	func (T) M1() {}

	func main() {
	    var v1 I1 = T{}
	    _ = v1
	}

```

当接口类型值需要赋值给具体类型的变量时，它是如何工作的？（[源代码](https://play.golang.org/p/gpu4Dh8e1c)）

```

	type I1 interface {
	    M1()
	}

	type T struct{}

	func (T) M1() {}

	func main() {
	    var v1 I1 = T{}
	    var v2 T = v1
	    _ = v2
	}

```

这不能正常运行并且会抛出一个错误 `` cannot use v1 (type I1) as type T in assignment: need type assertion `` 。这里就涉及到了类型断言……

类型转换只有在Go编译器能够检查其正确性时才能进行。 以下情况在编译时无法通过：

1. 接口类型 → 具体类型（[源代码](https://play.golang.org/p/MdV355DtVq)）：

```

	type I interface {
	    M()
	}

	type T struct {}

	func (T) M() {}

	func main() {
	    var v I = T{}
	    fmt.Println(T(v))
	}

```

它会给出一个编译错误 `` cannot convert v(type I) to type T: need type assertion `` 。原因是编译器不知道这种隐式转换是否有效，因为任何实现接口 *I* 的值都可以被赋值给变量 *v* 。

2. 接口类型 → 接口类型，当右边接口方法集不是左边接口方法集的子集时（[源代码](https://play.golang.org/p/AlTa00Inin)）

```

	type I1 interface {
	    M()
	}

	type I2 interface {
	    M()
	    N()
	}

	func main() {
	    var v I1
	    fmt.Println(I2(v))
	}

```

编译结果：

```

	main.go:16: cannot convert v (type I1) to type I2:
		I1 does not implement I2 (missing N method)

```

错误原因同上。如果 *I2* 方法集是 *I1* 方法集的子集，编译器在编译的阶段就能知道。但是这里不同，这样的转换在运行时才可能发生。

这不是严格的类型转换，类型断言和类型转换允许检查/检索接口类型值的动态值甚至将接口的类型值转换成不同接口的类型值。

## 类型断言
类型断言的语法如下：

```

	v.(T)

```

*v* 是一个接口类型, *T* 是一个抽象或者具体的类型。

### 具体类型
先让我们来看一下它是如何作用在具体类型上的（[源代码](https://play.golang.org/p/3bkUvw0hlv)）:

```

	type I interface {
	    M()
	}

	type T struct{}

	func (T) M() {}

	func main() {
	    var v1 I = T{}
	    v2 := v1.(T)
	    fmt.Printf("%T\n", v2) // main.T
	}

```

类型断言中的类型必须实现了 *v1* 的接口 —— *I* 。这将在编译阶段被证明（[源代码](https://play.golang.org/p/qfGgVyVbKF)）:

```

	type I interface {
	    M()
	}

	type T1 struct{}

	func (T1) M() {}

	type T2 struct{}

	func main() {
	    var v1 I = T1{}
	    v2 := v1.(T2)
	    fmt.Printf("%T\n", v2)
	}

```

这样的代码不可能成功编译，因为 `` impossible type assertion `` 错误。变量 *v1* 不能存放 *T2* 类型，因为变量 *v1* 只能存放实现了接口 *I* 的类型的值，而 *T2* 类型不满足接口 *I* 。

编译器不知道在运行过程中变量 *v1* 会存放什么样的值。类型断言是一种从接口类型值中检验动态值的方法。但是当 *v1* 的动态类型与 *T* 不匹配会发生什么？（[源代码](https://play.golang.org/p/sTKBb1eW6r)）

```

	type I interface {
	    M()
	}

	type T1 struct{}

	func (T1) M() {}

	type T2 struct{}

	func (T2) M() {}

	func main() {
	    var v1 I = T1{}
	    v2 := v1.(T2)
	    fmt.Printf("%T\n", v2)
	}

```

程序将会 panic :

``

	panic: interface conversion: main.I is main.T1, not main.T2

``

### 多返回值形式（请不要 panic）
类型断言可以以多值形式使用，其中附加的第二个值是一个布尔值，表示断言是否成立。 如果不是，则第一个值是类型 *T* 的 [零值](https://golang.org/ref/spec#The_zero_value) （[源代码](https://play.golang.org/p/gmmE4oPgyb)）：

```

	type I interface {
	    M()
	}

	type T1 struct{}

	func (T1) M() {}

	type T2 struct{}

	func (T2) M() {}

	func main() {
	    var v1 I = T1{}
	    v2, ok := v1.(T2)
	    if !ok {
	        fmt.Printf("ok: %v\n", ok) // ok: false
	        fmt.Printf("%v,  %T\n", v2, v2) // {},  main.T2
	    }
	}

```

这种形式不会 panic ，布尔常量作为第二个值被返回，用来检查断言是否成立。

### 接口类型
在上述所有情况下，类型断言中使用的类型都是具体的。Golang 还允许传递接口类型。 它检查动态值是否满足所需的接口并返回此接口类型值的值。在转换规则中，传递给类型断言的接口的方法集不必是 *v* 的类型方法集的子集（[源代码](https://play.golang.org/p/TU4eTCE0Yl)）：

```

	type I1 interface {
	    M()
	}

	type I2 interface {
	    I1
	    N()
	}

	type T struct{
	    name string
	}

	func (T) M() {}
	func (T) N() {}

	func main() {
	    var v1 I1 = T{"foo"}
	    var v2 I2
	    v2, ok := v1.(I2)
	    fmt.Printf("%T %v %v\n", v2, v2, ok) // main.T {foo} true
	}

```

如果接口不被满足，将会返回接口的零值即 *nil* ([源代码](https://play.golang.org/p/NQ81pDKzY1)):

```

	type I1 interface {
	    M()
	}

	type I2 interface {
	    N()
	}

	type T struct {}

	func (T) M() {}

	func main() {
	    var v1 I1 = T{}
	    var v2 I2
	    v2, ok := v1.(I2)
	    fmt.Printf("%T %v %v\n", v2, v2, ok) // <nil> <nil> false
	}

```

类型断言的单返回值形式同样支持接口类型。

### nil
当 *v* 是零值时，类型断言总会失败。不管 *T* 是接口类型还是具体的类型（[源代码](https://play.golang.org/p/MIeo1OfdYx)）：

```

	type I interface {
	    M()
	}

	type T struct{}

	func (T) M() {}

	func main() {
	    var v1 I
	    v2 := v1.(T)
	    fmt.Printf("%T\n", v2)
	}

```

上述程序会 panic :

```

	panic: interface conversion: main.I is nil, not main.T

```

当 *v* 是零值时，之前介绍的多返回值形式会避免 panic —— [证明](https://play.golang.org/p/39nlRMfH-E)。

## 类型转换
类型断言仅仅只是一个方法，用来判断一个接口类型值的动态类型是否实现了所需要的接口或者与传递的具体类型值相同。如果代码需要对单个变量进行多次的测试，Golang 提供了一个比类型断言更简洁的结构，类似传统的 switch 语句：

```

	type I1 interface {
	    M1()
	}

	type T1 struct{}

	func (T1) M1() {}

	type I2 interface {
	    I1
	    M2()
	}

	type T2 struct{}

	func (T2) M1() {}
	func (T2) M2() {}

	func main() {
	    var v I1
	    switch v.(type) {
	    case T1:
	            fmt.Println("T1")
	    case T2:
	            fmt.Println("T2")
	    case nil:
	            fmt.Println("nil")
	    default:
	            fmt.Println("default")
	    }
	}

```

语法和类型断言很相似，但是使用 [关键字](https://golang.org/ref/spec#Keywords) *type* 。当接口类型值的值为 *nil* ，那么输出是 `` nil `` （[源代码](https://play.golang.org/p/IoOCtm5gaR)）,但当我们将 *v* 赋值：

```

	var v I1 = T2{}

```

程序就会打印出 `` T2 `` ([源代码](https://play.golang.org/p/2LbRnZs0BU))。类型转换同样可以作用在借口类型上（[源代码](https://play.golang.org/p/2LbRnZs0BU)）：

```

	var v I1 = T2{}
	switch v.(type) {
	case I2:
	        fmt.Println("I2")
	case T1:
	        fmt.Println("T1")
	case T2:
	        fmt.Println("T2")
	case nil:
	        fmt.Println("nil")
	default:
	        fmt.Println("default")
	}

```

这会打印出 `` T2 ``。如果同时匹配多个接口类型会进入第一个（从上到下）。如果没有匹配的类型则什么都不会发生（[源代码](https://play.golang.org/p/y7EhLa25OL)）:

```

	type I interface {
	    M()
	}

	func main() {
	    var v I
	    switch v.(type) {
	    }
	}

```

这个程序不会 panic ——它会成功地结束执行。

### 一个 case 多个类型
单个 switch case 可以指定多个类型，用逗号分隔。当出现多个类型对应相同代码块时，这样做可以避免重复的代码（[源代码](https://play.golang.org/p/jrbNPnu9eE)）：

```

	type I1 interface {
	    M1()
	}

	type T1 struct{}

	func (T1) M1() {}

	type T2 struct{}

	func (T2) M1() {}

	func main() {
	    var v I1 = T2{}
	    switch v.(type) {
	    case nil:
	            fmt.Println("nil")
	    case T1, T2:
	            fmt.Println("T1 or T2")
	    }
	}

```

当 *v* 的动态类型被 [卫兵](https://golang.org/ref/spec#TypeSwitchGuard) 判定为 *T2* 时会打印出 `` T1 or T2 `` 。

### default case
这种情况和以前的 switch 语句很相似。它会被用在找不到任何匹配类型的时候（[源代码](https://play.golang.org/p/8nsUrsN9NS)）：

```

	var v I
	switch v.(type) {
	default:
	        fmt.Println("fallback")
	}

```

### 变量简短声明
目前为止我们已经了解了类型转换，其中的 [卫兵](https://golang.org/ref/spec#TypeSwitchGuard) 有以下的语法：

```

	v.(type)

```

其中 *v* 是变量名。此外变量简短声明可以用在这里（[源代码](https://play.golang.org/p/AeFTeHSky0)）：

```

	var p *T2
	var v I1 = p
	switch t := v.(type) {
	case nil:
	         fmt.Println("nil")
	case *T1:
	         fmt.Printf("%T is nil: %v\n", t, t == nil)
	case *T2:
	         fmt.Printf("%T is nil: %v\n", t, t == nil)
	}

```

这会打印 `` *main.T2 is nil: true `` ，所以 *t* 的类型是 case 语句中的类型。如果一条语句中有多个类型， 那么 *t* 的类型将和 *v* 的类型一样（[源代码](https://play.golang.org/p/XMU8wC8X2h)）：

```

	var p *T2
	var v I1 = p
	switch t := v.(type) {
	case nil:
	         fmt.Println("nil")
	case *T1, *T2:
	         fmt.Printf("%T is nil: %v\n", t, t == nil)
	}

```

这个输出 `` *main.T2 is nil: false `` 。变量 *t* 是接口类型因它不是 *nil* 而是指向一个 *nil* 指针（ [part I](https://medium.com/golangspec/interfaces-in-go-part-i-4ae53a97479c) 中解释了接口类型什么时候为 *nil*）。

### 重复
case 语句中指定的类型必须是唯一的（[源代码](https://play.golang.org/p/ZIvfT8-0Gm)）：

```

	switch v.(type) {
	case nil:
	    fmt.Println("nil")
	case T1, T2:
	    fmt.Println("T1 or T2")
	case T1:
	    fmt.Println("T1")
	}

```

编译这段代码的话会以错误终止 `` duplicate case T1 in type switch ``。

### 可选的简单语句
[卫兵](https://golang.org/ref/spec#TypeSwitchGuard) 的前面可以加上一条 [简单的语句](https://medium.com/golangspec/simple-statement-notion-in-go-b8afddfc7916)，像另一条简短的变量声明（[源代码](https://play.golang.org/p/NPXLEa6b8v)）：

```

	var v I1 = T1{}
	switch aux := 1; v.(type) {
	case nil:
	    fmt.Println("nil")
	case T1:
	    fmt.Println("T1", aux)
	case T2:
	    fmt.Println("T2", aux)
	}

```

程序会打印 `` T1 1 `` 。此外，不管卫兵是否是变量简短声明的形式，这个简单的语句都可以使用。

点击下面的 ❤ 让更多的人看到这篇文章。如果你想获得有关最新帖子的更新或者推进后续文章的工作，请关注我。

---

via: https://medium.com/golangspec/interfaces-in-go-part-ii-d5057ffdb0a6

作者：[Alex Yakunin](https://medium.com/@alexyakunin)
译者：[csshawn](https://github.com/csshawn)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出