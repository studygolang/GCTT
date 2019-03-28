首发于：https://studygolang.com/articles/18822

# 仔细研究 Go(golang) 类型系统

通过示例详解 Go 的类型系统

让我们从一个非常基本的问题开始吧！

## 为什么我们需要类型

在回答这个问题之前，我们需要先看看编程语言的一些原始抽象层，日常的工作我们并不需要处理这些层。

### 我们如何才能获得数据的机器表示呢？

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/A-Closer-Look-at-Go-Type-System/binary_zero_one.jpeg)

机器所能理解的是二进制 0 和 1。但对我们来说，意义何在？直到看到这样的东西，我才知道（有人是黑客帝国迷吗？）

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/A-Closer-Look-at-Go-Type-System/Matrix.gif)

我们将这些二进制抽象出来，并进一步考虑。

看下这段汇编代码

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/A-Closer-Look-at-Go-Type-System/registers.png)

你能说出寄存器 R1、R2、R3 中的数据类型是什么吗 ?

你可能希望他们是整数，因为在汇编语言层面也无法确定。没有什么能阻止 R1、R2、R3 具有任意类型，它们只是一堆 0 和 1 的寄存器。即使没有意义，加法运算也会将 R2 和 R3 加在一起，产生一个位模式，并将结果存储在 R1 中。

所以类型的概念从更高层次的抽象开始，在更高级的语言中，比如 C、Go、Java、Python、JavaScript 等等，这是语言本身的特性。

## 什么是类型

类型的概念在不同的编程语言之间是不同的，可以用许多不同的方式来表达，但是大体上它们都有一些相同点。

1. 类型是一组值；
2. 在这些值上可以执行一组操作，例如：int 类型可以执行 `+` 和 `-` 等运算，而对于字符类型，可以执行连接、空检查等操作；

> 因此，语言类型系统指定哪些运算符对哪些类型有效。

类型检查的目标是确保运算符使用在正确的类型上。通过执行类型检查，可以强制执行预期的值解释，因为一旦我们得到一堆只是 0 和 1 的机器码，就执行不了任务检查。机器也会很愉快地在这些机器码上执行我们告诉它的任何操作。

> 类型系统用来强制执行位模式的预期解释，确保整数的位模式不会有任何非整数操作，从而得到无意义的结果。

## Go 的类型系统

有一些基本规范控制 Go 的类型系统，我们会看一些重要的。

但我不会一次把所有的概念都列出来，在这里我将尝试用不同的例子来涵盖 Go 类型系统的一些基本概念，然后在讲解一些基本概念的同时带你浏览这些例子。

花点时间看看这些代码片段。其中哪一个会编译，为什么 ?

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/A-Closer-Look-at-Go-Type-System/code_snippet_one.png)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/A-Closer-Look-at-Go-Type-System/code_snippet_two.png)

我希望你写下你的答案和理由，这样，我们最后就可以一起来验证它。

### 命名类型

具有名称的类型：例如 int、int64、float32、string、bool 等，这些都是预先声明的。

另外，使用 type 关键字声明的任意类型也称为命名类型。

```go
var i int // named type
type myInt int // named type
var b bool // named type
```

> 命名 ( 定义 ) 类型总是与任何其他类型不同。

### 未命名类型

复合类型，包括 array、struct、pointer、function、interface、slice、Map 和 channel，都是未命名类型。

```go
[]string // unnamed type
map[string]string // unnamed type
[10]int // unnamed type
```

因为它们没有名称，但是有关于它们是如何组成的类型字面量描述符

### 底层类型

每种类型 T 都有底层类型

> 如果 T 是预定义类型的一种，包括 bool、数字、string 或者 字面量类型，则对应的底层类型是 T 本身；否则，T 的底层类型是 T 在其类型声明中引用的类型的底层类型。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/A-Closer-Look-at-Go-Type-System/type_declaration.png)

第 3、 8 行，预声明的字符串类型，因此底层类型是 T 本身，即字符串；

第 5 、7 行，是字面量类型，因此底层类型就是 T 本身，即 map[string]int 和 指针 *N。注意：字面量类型也是未命名类型；

第 4 、6、10 行，T 的底层类型是 T 在其类型声明中引用的类型的底层类型，例如：B 引用了 A，所以 B 的底层类型是字符串类型，其他情况同理；

我们再来看下第 9 行的例子：type T map[S]int ，由于 S 的底层类型是 string，难道 type T map[S]int 的底层类型不应该是 map[string]int 而不是 map[S]int 吗？因为我们在谈论 map[S]int 的底层未命名类型，所以向下追溯到未命名类型（或者，正如 Go 语言规范上写的一样：如果 T 是类型字面量，则对应的底层类型就是 T 本身）。

你可能会想，为什么我会如此重视未命名类型、命名类型和底层类型的规范。因为它们在 Go 语言规范中扮演重要的角色，我们将进一步讨论，以帮助我们理解为什么上面展示的代码片段有的能编译而有的不能编译，即使它们的意思基本相同。

### 可赋值性

当变量 a 可以赋值给类型 T 的变量时。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/A-Closer-Look-at-Go-Type-System/assignability_specs.png)

虽然这些条件已经在规范里解释过了，我们来看其中的一条规范，当分配时，**两者都应该具有相同的底层类型，并且至少其中一个不是命名类型**。

我们来看下图 4、 5 所示代码段的问题 ：

```go
package main

type aInt int

func main() {
	var i int = 10
	var ai aInt = 100
	i = ai
	printAiType(i)
}

func printAiType(ai aInt) {
	print(ai)
}
```

上面的代码编译不通过，编译时报错：

```
8:4: cannot use ai (type aInt) as type int in assignment
9:13: cannot use i (type int) as type aInt in argument to printAiType
```

因为 i 是命名类型 int，而 ai 是命名类型 aInt，虽然它们的底层类型相同。

```go
package main

type MyMap map[int]int

func main() {
	m := make(map[int]int)
	var mMap MyMap
	mMap = m
	printMyMapType(mMap)
	print(m)
}

func printMyMapType(mMap MyMap) {
	print(mMap)
}
```

上面这段代码编译通过，因为 m 是未命名类型并且 m 和 mMap 的底层类型相同。

### 类型转化

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/A-Closer-Look-at-Go-Type-System/type_conversion_specs.png)

看下图 3 的代码：

```go
package main

type Meter int64

type Centimeter int32

func main() {
	var cm Centimeter = 1000
	var m Meter
	m = Meter(cm)
	print(m)
	cm = Centimeter(m)
	print(cm)
}
```

上面的代码可以编译通过，因为 **Meter** 和 **Centimeter** 都是整型，并且它们的底层类型可以相互转化。

在看图 1 和图 2 的代码之前，我们来看看 Go 控制类型系统的另一个基本规范。

### 类型一致性

两种类型要么相同要么不同。

**已定义类型与其他任意类型总是不同**。否则，如果两种类型对应的底层类型在结构上是等效的，则它们是相同的。因此，即使预先声明的命名类型 int、int64 等也是不相同的。

看下结构体的一条转换规则：

> 不考虑结构体标签，x 和 T 具有相同的底层类型（x 赋值给 T）。

```go
package main

type Meter struct {
	value int64
}

type Centimeter struct {
	value int32
}

func main() {
	cm := Centimeter{
		value: 1000,
	}

	var m Meter
	m = Meter(cm)
	print(m.value)
	cm = Centimeter(m)
	print(cm.value)
}
```

记住一点：**相同的底层类型**。由于成员 Meter.value 的底层类型是 int64，而成员 Centimeter.value 的底层类型是 int32，所以它们不相同，因为**已定义类型与其他任意类型总是不同**。所以图 2 的代码片段编译会出错。

来看下图 1 的代码段：

```go
package main

type Meter struct {
	value int64
}

type Centimeter struct {
	value int64
}

func main() {
	cm := Centimeter{
		value: 1000,
	}

	var m Meter
	m = Meter(cm)
	print(m.value)
	cm = Centimeter(m)
	print(cm.value)
}
view raw
```

成员 Meter.value 和 Centimeter.value 的底层类型都是 int64，所以它们相同，编译可以通过。

希望这篇文章对你理解 Go 类型系统有所帮助，这也是我一直写文章的目的。

---

via: https://medium.com/@ankur_anand/a-closer-look-at-go-golang-type-system-3058a51d1615

作者：[Ankur Anand](https://medium.com/@ankur_anand)
译者：[Seekload](https://github.com/Seekload)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
