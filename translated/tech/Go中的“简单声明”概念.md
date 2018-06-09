Go中的“简单语句”概念
  =================
  在Golang[规范](https://golang.org/ref/spec)中有一个术语简单声明， 这个在整个文档中可能不常使用但Go的语法只允许在几个重要的地方使用这些声明。这篇短文的目的是介绍简单声明和可以使用它的地方。
![](https://cdn-images-1.medium.com/max/1600/1*Sh2PTmHloYTNWptEXj4FrA.jpeg)
    <center>Photography: Ursula Coyote / Netflix</center>

在规范文档中术语的简单声明被[SimpleStmt](https://golang.org/ref/spec#Statements)的产生（使用[EBNF](https://golang.org/ref/spec#Notation)）所定义：
```go
SimpleStmt = EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment | ShortVarDecl .
```
让我们通过介绍6种简单的类型声明来开始吧
  ### 1. 空白声明
  显然在简单声明中，空白声明什么都不会做
  ### 2.增加或减少声明
```go
x++  // 语义上等价于赋值x + = 1
x--  // x -= 1
```
在Go中，在表达式操作对象后用“- -”或“++”符号表示的构造并不是一个表达式，所以你在源文件中的.go拓展中看不到像下面这样纠结的东西，如：
```go
foo = bar++ + 10
```
尽管这是一个声明，但因为它是简单的声明所以可以用有意义的地方并且不会在不应该的地方引入不必要的负担（通过使代码难以阅读和难以维护）。

_**在Golang中没有前缀版本，即在操作对象之前放置“ - - ”或“++”**。_
  ### 3.发送声明
词法标记“< - ”表示通过频道发送值。它不返回任何值因此不需要表达它。
  ### 4.表达声明
某些表达式可以放入需要声明的地方。
_**块是由大括号括起来的一系列语句，例如，表达式语句在这种情况下完全有效。**_
允许的选项是：
- 函数调用（除了一些内置函数，例如append，cap，complex，imag，len，make，new，real，unsafe.Alignof，unsafe.Offsetof和unsafe.sizeOf）
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
  ### 5. 赋值
大家应该都熟悉一些最基本的赋值形式。必须声明变量，右边的表达式必须可[赋](https://medium.com/golangspec/assignability-in-go-27805bcd5874) 给一个变量:
```go
var population int64
population = 8000000 // 赋值
var city, country string
// 两个单值表达式赋值给两个变量
city, country = "New York", "USA"
```

当把一个以上的值分配给变量列表时，有两种形式。第一种形式，一个表达式返回多个值，例如函数调用:

```go
f := func() (int64, string) {
    return 2000000, "Paris"
}
var population int64
var city string
population, city = f()

```  
第二种形式是右侧包含有许多表达式, 但每一个是单值的:
```go
population, city = 13000000, "Tokyo"
```
这两种形式不可以被混淆:
```go
f := func() (int64, string) {
    return 4000000, "Sydney"
}
var population int64
var city, country string
population, city, country = f(), "Australia"
```
上面这行代码在编译的时候会抛出"多值 f() 在单值环境中"

赋值操作是一个由“=”符号组成的构造，等号前面是一个二元运算符:

- 加 (+)
- 减 (-)
- 位运算 或 (|)
- 位运算 异或 (^)
- 乘 (*)
- 除 (/)
- 取余 (%)
- 位运算 左移 右移 (<< and >>)
- 位运算 与 (&)
- 位清除 (&^)

a op = b(例如a + = b)在语义上等同于a = a op b，但a只被计算一次。赋值操作也是有效的赋值语句.

### 6. 简单变量声明
下面是常规变量声明的缩写，这个需要初始化程序。此外，未明确指定的类型是从初始化程序中使用的类型推断出来的:
```go
s := S{}
a := [...]int{1,2,3}
one, two := f(), g()
```

现在很清楚是什么构成了一组有效的简单声明。但它们在哪里适用呢?
### **如果声明**
简单声明可以选择性地放在条件表达式之前。它经常用于声明变量（使用短变量声明），然后在表达式中使用:

```go
gen := func() int8 {
    return 10
}
if num := gen(); num > 5 {
    fmt.Printf("%d is greater than 5\n", num)
}
```

### 为了声明
for语句中的初始化或声明变量只能是简单的语句。常见的用法是将短变量声明作为初始声明以及赋值操作或增量/减量语句:

```go
for i := 0; i < 10; i += 1 {
    fmt.Println(i)
}
```
尽管也没有东西阻止程序员在这个地方使用其他的简单声明。

### 变换声明
与if语句类似，可以在[表达式切换](https://golang.org/ref/spec#Expression_switches)之前将可选的简单语句放入表达式中:

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

或者在[类型切换](https://golang.org/ref/spec#Type_switches)的保护前面:

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

```go
T1 main.T1 10
```

  

  
  

  --------------------------------------------------------------------------------
  
  via: [原文地址](https://medium.com/golangspec/simple-statement-notion-in-go-b8afddfc7916)
  
  作者：[Michał Łowicki](https://medium.com/@mlowicki)
  译者：[yousanflics](https://github.com/yousanflics)
  校对：[校对者ID](https://github.com/校对者ID)
  
  本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
