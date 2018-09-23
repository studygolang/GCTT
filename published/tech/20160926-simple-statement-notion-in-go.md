首发于：https://studygolang.com/articles/13931

# Go 中的 “简单语句”

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/simple-statement/1_HO_nLfJ1LDrODqso68AcwA.jpeg)
[https://en.wikipedia.org/wiki/Kiss_(band)](https://en.wikipedia.org/wiki/Kiss_%28band%29)

在 Golang [规范](https://golang.org/ref/spec)中有一个术语：简单语句（simple statement），可能在整个文档中不经常使用，但语言的语法只允许在几个重要的地方使用这些语句。这篇短文的目的就是介绍这些简单语句和可以使用它们的地方。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/simple-statement/1_Sh2PTmHloYTNWptEXj4FrA.jpeg)
<center>Photography: Ursula Coyote/Netflix</center>

在规范中，它由 SimpleStmt 产生式定义（使用 [EBNF](https://golang.org/ref/spec#Notation)）：

```
SimpleStmt = EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment | ShortVarDecl .
```

下面我们将介绍 6 种简单语句

## 1. 空白语句

显然，在简单语句中，空白语句就是什么都不做。

## 2. 自增或自减语句

```go
x++  // 语义上等价于赋值x + = 1
x--  // x -= 1
```

在 Go 中，在表达式操作数后用 “--” 或 “++” 构成的不是表达式（译注：Go 中 "--" 或 "++" 不是表达式，而是语句），因此您不会在 `.go` 源文件中看到让人类似下面纠结的组合：

```go
foo = bar++ + 10
```

尽管这是一个语句，但是作为简单语句是需要用在有意义的地方并且不会在不应该的地方引入不必要的**负担**(通过使代码难以阅读和难以维护)。

> **在 Golang 中没有前缀版本，即在操作对象之前放置 “--” 或 “++”。**

## 3. 发送语句

词法标记 “<-” 表示通过 channel 发送值。它不返回任何值，因此不需要表达它。

## 4. 表达式语句

某些表达式可以放入语句中。

> **块是由大括号括起来的一系列语句，例如，表达式语句在如下情况下完全有效。**

允许的选项如下：

- 函数调用（除了一些内置函数，例如 append，cap，complex，imag，len，make，new，real，[unsafe](https://golang.org/pkg/unsafe/).Alignof，unsafe.Offsetof 和unsafe.SizeOf）
- 方法调用
- 接收操作符

```go
func (s S) m() {
    fmt.Printf("Hi %s\n", s.name)
}

func f(n int) {
    fmt.Printf("f says %d\n", n)
}

func main() {
    f(1)
    s := S{"Michał"}
    s.m()
    c := make(chan int)
    go func() {
        c <- 1
    }()
    <-c
    numbers := []int{1, 2, 3}
    len(numbers)  // error: len(numbers) 被执行了但是并未使用
}
```

## 5. 赋值语句

大家应该都熟悉一些最基本的赋值形式。首先必须声明变量，右边的表达式必须可以[赋值](https://studygolang.com/articles/12381)给一个变量:

```go
var population int64
population = 8000000 // 赋值
var city, country string
// 两个单值表达式赋值给两个变量
city, country = "New York", "USA"
```

当把一个以上的值分配给变量列表时，有两种形式。第一种形式，一个表达式返回多个值，例如函数调用：

```go
f := func() (int64, string) {
    return 2000000, "Paris"
}
var population int64
var city string
population, city = f()
```

第二种形式是右侧包含有许多表达式, 但每一个是单值的：

```go
population, city = 13000000, "Tokyo"
```

这两种形式不可以被混淆：

```go
f := func() (int64, string) {
    return 4000000, "Sydney"
}
var population int64
var city, country string
population, city, country = f(), "Australia"
```

上面这行代码在编译的时候会抛出 "multiple-value f() in single-value context"。

赋值操作是一个由 “=” 符号组成的构造，等号前面是一个二元运算符：

- 加 (+)
- 减 (-)
- 位运算 或 (|)
- 位运算 异或 (^)
- 乘 (*)
- 除 (/)
- 取余 (%)
- 位运算 左移 右移 (<< 和 >>)
- 位运算 与 (&)
- 位清除 (&^)

`a op= b`（例如 a += b）在语义上等同于 `a = a op b`，但 a 只被计算一次。赋值操作也是有效的赋值语句.

## 6. 简单变量声明

下面是常规变量语句的缩写，它们需要初始化。此外，未明确指定的类型是从初始化语句中使用的类型推断出来的：

```go
s := S{}
a := [...]int{1,2,3}
one, two := f(), g()
```

现在我们应该很清楚是什么构成了一组有效的简单语句。但是它们用在什么地方呢？

## if 语句

简单语句可以选择性地放在条件表达式之前。它经常用于声明变量（使用短变量语句），然后在表达式中使用：

```go
gen := func() int8 {
    return 10
}
if num := gen(); num > 5 {
    fmt.Printf("%d is greater than 5\n", num)
}
```

## for 语句

for 语句中的初始化或声明变量只能是简单语句。常见的用法是将短变量声明作为初始声明以及赋值操作或自增/自减语句:

```go
for i := 0; i < 10; i += 1 {
    fmt.Println(i)
}
```

当然了，也没有东西阻止程序员在这个地方使用其他的简单语句。

## switch 语句

与 if 语句类似，可以在 [switch 表达式](https://golang.org/ref/spec#Expression_switches)之前将可选的简单语句放入表达式中：

```go
switch aux := 2; letter {
case 'a':
    fmt.Println(aux)
case 'b':
    fmt.Println(aux << 1)
case 'c':
    fmt.Println(aux << 2)
default:
    fmt.Println("no match")
}
```

或者在 [type switch](https://golang.org/ref/spec#Type_switches) 的守卫前面：

```go
type I interface {
    f(num int8) int8
}
type T1 struct{}
func (t T1) f(num int8) int8 {
    return 1
}
type T2 struct{}
func (t T2) f(num int8) int8 {
    return 2
}
...
var i I = T1{}
switch aux := 10; val := i.(type) {
case nil:
    fmt.Printf("nil %T %d\n", val, aux)
case T1:
    fmt.Printf("T1 %T %d\n", val, aux)
case T2:
    fmt.Printf("T2 %T %d\n", val, aux)
}
```

它的输出为:

```
T1 main.T1 10
```

---

via: https://medium.com/golangspec/simple-statement-notion-in-go-b8afddfc7916

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[yousanflics](https://github.com/yousanflics)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
