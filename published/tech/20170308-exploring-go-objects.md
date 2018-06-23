已发布：https://studygolang.com/articles/12703

# 探索 Go 中的对象（object）

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-object/goexplorer.png)

当我接受了 Go 根本没有 object 之后，我才开始更容易理解 Go 的 object 是什么，其实就是一些可以操作共有状态的函数集合，加了点语法糖的点缀。

你可能心想“闭嘴吧，Go 当然有 object”，或者想“能操作共有状态的函数集合就是 object 的定义啊”，好吧，也许你是对的。

至少从我能想到的之前用过的 object 来看，我没看出来一些操作相同状态的相关函数的集合，和一个 object 有啥区别。再说 Go 的 object 模型不止是语法糖（我说的貌似有点极端哈:-)）。

不过 object 模型和经典的，比如 Java，C++ 和 Python（我目前就了解这么多）的模型相比还是有很大不同的。

在苦苦探索 Go 的 object 是如何工作的过程中，放弃传统的 object 观念，而只从函数方面考虑问题，这让我受益良多。

我要做的就是尝试将 object 模型解构成函数然后重建，看 Go 是如何工作的，可以看出来，Go 更倾向于用 object 做辅助从而让语法更简单，而不是像面向对象的语言一样，什么都是 object。

下面的解构看起来可能不太好，因为我不精通 Go,不过我还是强迫自己试了试，因为看起来还挺好玩的:-)。

## 从函数开始

好，我试着开始从函数上证实下，下面这个略蠢的例子是为了看看用函数能做什么。

我们来定义一个类型，实际上这是一个函数：

```go
type Adder func(int, int) int
```

你可以把这当成一个 interface（不过它们不是一回事）。任何符合这个特征的函数都可以被当作 **Adder** 类型：

```go
// Same type as Adder
func add(a int, b int) int {
	return a + b
}
```

一个不知道怎么 add 的抽象 adder：

```go
func abstractedAdd(a Adder, b int, c int) int {
	return a(b, c)
}
```

这真让人想起可以用 interface 做类似的事。 **abstractedAdd** 不知道怎么做 add，但是他可以接受任何一个遵循同样协议的 Adder 的实现。

下面给出最无用的也是最简单的例子，全部代码：

```go
package main

import "fmt"

type Adder func(int, int) int

// Same type as Adder
func add(a int, b int) int {
	return a + b
}

func abstractedAdd(a Adder, b int, c int) int {
	return a(b, c)
}

func main() {
	var a Adder
	fmt.Printf("Adder: %v\n", a)
	a = add
	fmt.Printf("Adder initialized: %v\n", a)
	fmt.Printf("%d + %d = %d\n", 1, 1, abstractedAdd(a, 1, 1))
	fmt.Printf("%d + %d = %d\n", 1, 1, abstractedAdd(add, 1, 1))
}
```

从这个例子中我们探讨下 Go 的 object。一个方法能符合 **Adder** 类型吗？依你的经验来看可能有点反直觉（就好象，你需要一个函数，而实际给了一个方法，这个意思），我们看看 adder object。

```go
type ObjectAdder struct{}

func (o *ObjectAdder) Add(a int, b int) int {
	return a + b
}
```

看起来没错，加到我们的例子里：

```go
package main

import "fmt"

type Adder func(int, int) int

func add(a int, b int) int {
	return a + b
}

func abstractedAdd(a Adder, b int, c int) int {
	return a(b, c)
}

type ObjectAdder struct{}

func (o *ObjectAdder) Add(a int, b int) int {
	return a + b
}

func main() {
	var a Adder
	fmt.Printf("Adder: %v\n", a)
	a = add
	fmt.Printf("Adder initialized: %v\n", a)
	fmt.Printf("func: %d + %d = %d\n", 1, 1, abstractedAdd(a, 1, 1))
	fmt.Printf("func: %d + %d = %d\n", 1, 1, abstractedAdd(add, 1, 1))

	var o *ObjectAdder
	fmt.Printf("object: %d + %d = %d\n", 1, 1, abstractedAdd(o.Add, 1, 1))
}
```

结果输出：

```
Adder: <nil>
Adder initialized: 0x401000
func: 1 + 1 = 2
func: 1 + 1 = 2
object: 1 + 1 = 2
```

哈，成功的。和接口不一样，函数签名不会匹配任何方法名称，你可以像传参数一样传方法，因为方法实际上就是函数，有点像下面这个：

```go
var o *ObjectAdder
fmt.Printf("object: %d + %d = %d\n", 1, 1, abstractedAdd(o.Whatever, 1, 1))
```

应该能运行。没看出来方法就是函数吗？再看这个：

```go
fmt.Printf("func add: %T\n", add)
fmt.Printf("object.Add: %T\n", o.Add)
```

结果输出：

```
func add: func(int, int) int
object.Add: func(int, int) int
```

你看出来这个空函数和这个 object 方法的区别了吗？没有，因为没区别。这就是为什么传参数可以运行。这也可以解释代码里另一个容易让学 Go 的新手（像我这样的）困惑的问题。

我们在例子中没有完全初始化 ObjectAdder。我用指针是有目的的，你们可以看到指针也没有初始化（nil 的）,可是代码却能运行。在我所知道的其他的面向对象的语言里，这不可能运行的，但是在 Go 里可以，为什么呢？

那是因为在 Go 里，根本就没有方法，没有方法类型，方法实际上就是语法糖，用来在做函数调用的时候传递一个实例类型来作为第一个参数（就像在 C 语言里习惯用的那样）。在 Go 里第一个参数类型通常被称为方法接收者，不过这没什么特别的，就是传递给函数的一个参数。

细化下我们的例子：

```go
fmt.Printf("ObjectAdder.Add: %T\n", (*ObjectAdder).Add)
fmt.Printf("ObjectAdder.Add: %d + %d = %d\n", 1, 1, (*ObjectAdder).Add(nil, 1, 1))
```

我在这做了什么呢？就是搞清楚你在做如下声明的时候 Go 实际上做了什么：

```go
type ObjectAdder struct{}

func (o *ObjectAdder) Add(a int, b int) int {
}
```

在这给 ***ObjectAdder** 类型添加了一个函数。这个函数可以被访问并且可以被当作任何值使用（被调用，作为参数传递等）。

如果你觉得“嗨，ObjectAdder 类型可不是 \*ObjectAdder”，好吧，在 Go 里指针类型确实是另一种类型，甚至和有函数组合的指针也不是一个类型。要加哪个类型的函数是由方法接收者决定的，在这个 case 里就是（\*ObjectAdder）。

这和 Go 的[方法集合](https://github.com/golang/go/wiki/MethodSets)的概念有关。

总之，继续往下看吧，输出结果：

```
ObjectAdder.Add: func(*main.ObjectAdder, int, int) int
ObjectAdder.Add: 1 + 1 = 2
```

根本就没有方法，就是函数。我们在 Go 里看到的 object 就是一些关联到某个类型的函数组合，加点语法糖，来把第一个参数传给你。

说实话就好像所有的面向对象中的 object 实际上实现了一样。好处是在 Go 这是 100% 简洁明确的，没有魔术，就是语法糖。 Go 在简洁这方面做的确实严谨。

这样很多事情就更简单一致了，从例子里可以看出来。传函数或方法做参数没有任何区别（我想不出来有区别的理由）。

下面是例子中最终的所有代码：

```go
package main

import "fmt"

type Adder func(int, int) int

// Same type as Adder
func add(a int, b int) int {
	return a + b
}

func abstractedAdd(a Adder, b int, c int) int {
	return a(b, c)
}

type ObjectAdder struct{}

func (o *ObjectAdder) Add(a int, b int) int {
	return a + b
}

func main() {
	var a Adder
	fmt.Printf("Adder: %v\n", a)
	a = add
	fmt.Printf("Adder initialized: %v\n", a)
	fmt.Printf("func: %d + %d = %d\n", 1, 1, abstractedAdd(a, 1, 1))
	fmt.Printf("func: %d + %d = %d\n", 1, 1, abstractedAdd(add, 1, 1))

	var o *ObjectAdder
	fmt.Printf("func add: %T\n", add)
	fmt.Printf("object.Add: %T\n", o.Add)
	fmt.Printf("object: %d + %d = %d\n", 1, 1, abstractedAdd(o.Add, 1, 1))

	fmt.Printf("ObjectAdder.Add: %T\n", (*ObjectAdder).Add)
	fmt.Printf("ObjectAdder.Add: %d + %d = %d\n", 1, 1, (*ObjectAdder).Add(nil, 1, 1))
}
```

这个例子是完全状态无关的。Object 通常会有状态和副作用，Go 的函数也有状态和副作用吗？

## 函数和状态

为了让函数和 object 之间的差距更小一点，我们用个最原始的/最简单的例子，一个 iterator：

```go
package main

import "fmt"

func iterator() func() int {
	a := 0
	return func() int {
		a++
		return a
	}
}

func main() {

	iter := iterator()

	fmt.Printf("iter 1: %d\n", iter())
	fmt.Printf("iter 2: %d\n", iter())
	fmt.Printf("iter 3: %d\n", iter())
}
```

如果你运行一下，你就会看到这个 iterator 是有效的。准确的讲我们现在有什么呢？我们有一个 **iterator** 函数，看起来像另一个函数的的构造函数，它会返回这个函数，这就是为什么 **iterator** 返回的类型是：

```go
func() int
```

通常指的闭包是这样的语法结构：

```go
a := 0
return func() int {
	a++
	return a
}
```

我们实例化的这个函数，用到了外部的一个变量，会将变量 **a** 和新创建的函数关联起来，它包含一个 **a** 的引用而且可以操作（**a**）。

如果你习惯用 object 作为一种状态管理方式（实际上很多 C 编程者也觉得奇怪，因为在 C 语言里函数是静态构造的），这就比较烧脑了。

在 Go 语言里，函数可以随时初始化，下面是这个例子的另一版，说明我们实际在初始化函数：

```go
package main

import "fmt"

func iterator() func() int {
	a := 0
	return func() int {
		a++
		return a
	}
}

func main() {
	itera := iterator()
	iterb := iterator()

	fmt.Printf("itera 1: %d\n", itera())
	fmt.Printf("itera 2: %d\n", itera())
	fmt.Printf("itera 3: %d\n", itera())

	fmt.Printf("iterb 1: %d\n", iterb())
	fmt.Printf("iterb 2: %d\n", iterb())
	fmt.Printf("iterb 3: %d\n", iterb())
}
```

得到结果：

```
itera 1: 1
itera 2: 2
itera 3: 3
iterb 1: 1
iterb 2: 2
iterb 3: 3
```

因此每个 iterator 都是互相独立的，没有办法让一个函数从另一个函数获得状态，除非在代码里明确的允许，或者你用不安全的包，做很糟糕的指针运算。

这还挺有意思的，因为像 Lisp 这样的语言最初就有闭包，提供了你可以想象的绝对最大程度的封装。除了从函数中你没有其他办法直接获取状态。

我们看一眼用 Go 的 object 的闭包是什么样的：

```go
package main

import "fmt"

type iterator struct {
	a int
}

func (i *iterator) iter() int {
	i.a++
	return i.a
}

func newIter() *iterator {
	return &iterator{
		a: 0,
	}
}

func main() {
	i := newIter()

	fmt.Printf("iter 1: %d\n", i.iter())
	fmt.Printf("iter 2: %d\n", i.iter())
	fmt.Printf("iter 3: %d\n", i.iter())
}
```

可以看到，非常简单的事情，用 object 的方式会显得更笨拙一点，至少在我看来是这样。我甚至用了一个同样糟糕的名为 **a** 的 int 变量，实际上它表示状态。

现在我们创建一个 struct，保存状态，给函数添加类型，用这个函数控制状态。 如果你觉得复杂，你也可以不用 struct：

```go
package main

import "fmt"

type iterator int

func (i *iterator) iter() int {
	*i++
	return int(*i)
}

func main() {
	var i iterator

	fmt.Printf("iter 1: %d\n", i.iter())
	fmt.Printf("iter 2: %d\n", i.iter())
	fmt.Printf("iter 3: %d\n", i.iter())
}
```

这个函数也做了同样的事情，用不同的方式。和 object 一样也可以管理状态，而且用作用域将状态隔离，只有函数能修改这个状态。

为了完成这部分，我们写几个函数，操作一个共享状态（这就是 object 所做的）：

```go
package main

import "fmt"

type stateChanger func() int

func new() (stateChanger, stateChanger) {
	a := 0
	return func() int {
			a++
			return a
		},
		func() int {
			a--
			return a
		}
}

func main() {
	inc, dec := new()

	fmt.Printf("inc 1: %d\n", inc())
	fmt.Printf("inc 2: %d\n", inc())
	fmt.Printf("inc 3: %d\n", inc())

	fmt.Printf("dec 1: %d\n", dec())
	fmt.Printf("dec 2: %d\n", dec())
	fmt.Printf("dec 3: %d\n", dec())
}
```

输出：

```
inc 1: 1
inc 2: 2
inc 3: 3
dec 1: 2
dec 2: 1
dec 3: 0
```

可以清楚看到，两个函数共享同一个状态，都能管理状态，就像你用含有两个方法的 object 做的一样。

当然我不是鼓励大家随便用带一些变量的函数，struct 的存在就是为了给不同类型组合命名，赋予含义的。

和函数使用一样，只有一堆松散的函数在大部分情况下（比如带数据的情况）是很糟糕的。

因为 Go 的确把函数作为第一等公民，struct 中有很多含有函数的字段，模拟方法的行为，来代表一组操作共有状态的函数。但是这样很难用而且容易出错，比如存在调用未初始化的字段/方法的可能（凡是用 C 编码的人都能明白这个问题，和将造成的后果）。

一个日历操作：

```go
package main

import "fmt"

type Calculator struct {
	Add func(int,int) int
	Sub func(int,int) int
}

func newCalculator() Calculator {
	return Calculator{
		Add: func(a int, b int) int {
			return a + b
		},
		Sub: func(a int, b int) int {
			return a - b
		},
	}
}

func main() {
	calc := newCalculator()
	fmt.Println(calc.Add(3, 2))
	fmt.Println(calc.Sub(3, 2))
}
```

嗯，你可以争论代码冗繁的问题，以你的经验来看这种方式可能比 Go 使用方法更好。但是你不可否认这给犯错误留了很大空间。

比如下面这个：

```go
package main

import "fmt"

type Calculator struct {
	Add func(int,int) int
	Sub func(int,int) int
}

func newCalculator() Calculator {
	return Calculator{
		Add: func(a int, b int) int {
			return a + b
		},
		Sub: func(a int, b int) int {
			return a - b
		},
	}
}

func main() {
	var calc Calculator
	fmt.Println(calc.Add(3, 2))
	fmt.Println(calc.Sub(3, 2))
}
```

输出结果是：

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0xffffffff addr=0x0 pc=0xc64e8]

goroutine 1 [running]:
main.main()
	/tmp/sandbox772959961/main.go:23 +0x28
```

虽然使用方法也可能出现这个问题，但给类型添加函数比代表函数组合操作同一个类型要安全好多倍。至少调用方法永远是安全的（当然你也可能遇到一个无效的方法接受者，使程序崩溃）。

除了繁琐和容易出错，还有一个问题就是怎样表达抽象概念，这比单独的函数复杂多了。

## 抽象

我们目前的所有抽象方法都是由一个函数组成的，一个函数就可以表示，但是如果抽象需求多于一个函数怎么办呢？

如果没办法表达，你只能在一个函数里合并抽象操作，那也太恐怖了（想象一下 read/write 抽象模型就在同一个函数里的情形）。

上面 calculator 例子里提供了一种模拟方法的方式，达到这样的程度，就是人们看怎么使用 Calculator 的时候看不出来那些方法其实根本不是方法。

但是有个重要的概念没有了，一个在 Go 的方法里很基础的概念，你怎么去表示你需要一组函数，不用定义谁去实现，或者怎样实现？

完整的说下，给出一个函数 X，需要一组函数 Y，你如何在语法上表达出一个 Z 类型实现了这组被需要的 Y 函数，因此可以作为 X 函数使用的可行方案？

一个解决办法是使用**安全**多态。我希望能对同一组无缝交互的函数有多种不同的实现。多态的重点在**安全**。我就 C 的多态分享下我的观点，C 的多态是可行的，也运行的不错，但是绝对不安全。你可能会反对说没有实现是绝对安全的，但是最起码要比 C 安全，这也是大部分语言比如 Java 和 Python 在开发之初做到的。

安全是很重要的，因为 calculator 的例子可能会用来实现这样的形式。我们可以这样做：

```go
type Calculator struct {
	Add func(int,int) int
	Sub func(int,int) int
}

func codeThatDependsOnCalculator(c Calculator) {
	// etc
}
```

这将允许一个 **Calculator** 的 N 个不同的实现与依赖它的代码集成，但是不安全。很简单，只完成一半的实现就能瞒天过海了。所有接收 **Calculator** 的函数都要检查 **Add** 和 **Sub** 不是 nil 的。

这和 C 里面实现的太像了，这个工作显然是编译器能帮你做的（在 C 里面你可以用宏定义）。

Go 的解决方案是用接口，在我看来这是 Go 里最棒的特性。

鉴于这篇博文已经很长了，关于接口的思想演变我将在后续博文中讨论。

祝探索 Go 的过程愉快;-)。

## 致谢

特别感谢：

- [i4k](https://github.com/tiago4orion)
- [vitorarins](https://github.com/vitorarins)
- [cadicallegari](https://github.com/cadicallegari)

感谢诸位花时间帮我 review 并且指出了一些低级的错误。

---

via: https://katcipis.github.io/blog/exploring-go-objects/

作者：[TIAGO KATCIPIS](https://katcipis.github.io/)
译者：[ArisAries](https://github.com/ArisAries)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
