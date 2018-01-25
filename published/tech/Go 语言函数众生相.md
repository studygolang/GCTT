# Go 语言函数众生相

### 本文是对匿名函数、高阶函数、闭包、同步、延时（defer）及其他 Go 函数类型或特性的概览。

![The Zoo of Go Funcs](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/The_zoo_of_go_funcs.png)

> *这篇文章是针对 Go 语言中不同的函数类型或特性的摘要总结。*
>
> *更为深入的探讨我会在近期的文章中进行，因为那需要更多的篇幅。这只是一个开端。*

---

### 命名函数

一个命名函数拥有一个函数名，并且要声明在包级作用域中——*其他函数的外部*

*👉* ***我已经在[另一篇文章](https://blog.learngoprogramming.com/golang-funcs-params-named-result-values-types-pass-by-value-67f4374d9c0a)中对它们进行了完整的介绍***

![named Func](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/named_funcs.png)

<p align="center">这是一个命名函数：Len 函数接受一个 string 类型的参数并返回一个 int 类型的值</p>

---

### 可变参数函数

变参函数可接受任意数量的参数

*👉* ***我已经在[另一篇文章](https://blog.learngoprogramming.com/golang-variadic-funcs-how-to-patterns-369408f19085)中对它们进行了完整的介绍***

![Variadic Funcs](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/variadic_funcs.png)

---

### 方法

当你将一个函数附加到某个类型时，这个函数就成为了该类型上的一个方法。因此，它可以通过这个类型来调用。在通过类型来调用其上的某个方法时，Go 语言会将该类型（接收者）传递给方法。

#### 示例

新建一个计数器类型并为其定义一个方法：

```go
type Count int

func (c Count) Incr() int {
  c = c + 1
  return int(c)
}
```

如上的方法与以下写法有同样的效果（但并不等价）：

```go
func Incr(c Count) int
```

![Method](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/methods.png)

<p align="center">原理并不完全如上所示，但你可以像这样来理解</p>

#### 值传递

当 Incr 被调用时，Count 实例的值会被复制一份并传递给 Incr。

```go
var c Count; c.Incr(); c.Incr()

// output: 1 1
```

<h3 align="center"><i></i>c 的值并不会增加，因为 c 是通过值传递的方式传递给方法</i></h3>

![Value receiver](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/value_receiver.png)

#### 指针传递（引用传递）

想要改变计数器 c 的值，你需要给 Incr 方法传入 Count 类型指针——``*Count``。

```go
func (c *Count) Incr() int {
  *c = *c + 1
  return int(*c)
}

var c Count
c.Incr(); c.Incur()
// output: 1 2
```

![pointer receiver](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/pointer_receiver.png)

[![run the code](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/run_the_code.png)](https://play.golang.org/p/hGVJWPIFZG	"receiver")

<p align="center">在我之前的一些文章中有更多的示例：看<a href="https://blog.learngoprogramming.com/golang-const-type-enums-iota-bc4befd096d3#c320">这里！</a>看<a href="https://blog.learngoprogramming.com/golang-funcs-params-named-result-values-types-pass-by-value-67f4374d9c0a#638f">这里！</a></p>

---

### 接口方法

我们用**接口方法**的方式来重建上面的程序。先创建一个叫做 Counter 的新接口：

```go
type Counter interface {
  Incr() int
}
```

下面的 onApiHit 函数能使用任何拥有 `Incr() int` 方法的类型：

```go
func onApiHit(c Counter) {
  c.Incr()
}
```

我们即刻使用一下这个改造版的计数器——现在你可以使用一个名副其实的计数器接口了：

```go
dummyCounter := Count(0)
onApiHit(&dummyCounter)
// dummyCounter = 1
```

![interface methods](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/interface_funcs.png)

我们在 Count 类型上定义了一个 `Incr() int` 方法，因此 `onApiHit()` 方法可以通过它来增长 counter —— 我将 dummyCounter 的指针传入了 onApiHit，否则这个计数器不会因而增长。

[![run the code](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/run_the_code.png)](https://play.golang.org/p/w0oyZjmdMA	"interface method")

*接口方法与普通方法的区别在于接口方法更具伸缩性、可扩展性，并且它是松耦合的。你可以利用接口方法在不同的包之间进行各自所需的实现，而不用修改 onApiHit 或是是其他方法的代码*

---

### 函数是一等公民

一等公民意味着 Go 语言中函数也是一种值类型，可以像其他类型的值一样被存储或是传递。

![first-class funcs](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/first-class_funcs.png)

<p align="center">函数可以作为一种值类型和其他的类型配合使用，反之亦然</p>

#### 示例

以下程序通过 Crunchers 切片将一个数值序列作为参数传递到一个叫 ”crunch“ 的函数中去。

声明一个”用户自定义函数类型“，它需要接收一个 int 类型的值来返回一个 int 类型的值。

这意味着任何使用这种类型的代码都可以接受一个以如下形式签名的函数：

```go
type Cruncher func(int) int
```

声明一些 cruncher 类型的函数：

```go
func mul(n int) int {
  return n * 2
}

func add(n int) int {
  return n + 100
}

func sub(n int) int {
  return n - 1
}
```

Crunch 是一个[可变参数函数](https://blog.learngoprogramming.com/golang-variadic-funcs-how-to-patterns-369408f19085)，通过 Cruncher 类型的可变参数处理一系列的整型数：

```go
func crunch(nums []int, a ...Cruncher) (rnums []int) {
  // 创建一个等价的切片
  rnums = append(rnums, nums...)
  
  for _, f := range a {
    for i, n := range rnums {
      rnums[i] = f(n)
    }
  }
  
  return
}
```

声明一个具有一些初始值的整型切片，之后对它们进行处理：

```go
nums := []int{1, 2, 3, 4, 5}

crunch(nums, mul, add, sub)
```

#### 输出：

```
[101 103 105 107 109]
```

[![run the code](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/run_the_code.png)](https://play.golang.org/p/hNSKZAo0p6	"first-class func")

---

### 匿名函数

匿名函数即没有名字的函数，它以[函数字面量](https://golang.org/ref/spec#Function_literals)的方式在行内进行声明。它在实现闭包、高阶函数、延时函数等特殊函数时有极大作用。

![annoymous funcs](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/Anonymous_funcs.png)

#### 函数签名

命名函数：

```go
func Bang(energy int) time.Duration
```

匿名函数：

```go
func(energy int) time.Duration
```

它们有相同的函数签名形式，所以它们可以互换着使用：

```go
func(int) time.Duration
```

[![run the code](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/run_the_code.png)](https://play.golang.org/p/-az-2qBr9T	"annoymous func")

#### 示例

我们用匿名函数的方式重构一下上面的”函数是第一公民“单元中的 cruncher 程序。在 main 函数中声明几个匿名 cruncher 函数。

```go
func main() {
  crunch(nums,
         func(n int) int {
           return n * 2
         },
         func(n int) int {
           return n + 100
         },
         func(n int) int {
           return n - 1
         })
}
```

crunch 函数只期望接收到 Cruncher 类型的函数，并不关心它（它们）是命名函数还是匿名函数，因此以上代码可以正常工作。

为了提高可读性，在传入 crunch 之前你可以先将这些匿名函数赋值给变量。

```go
mul := func(n int) int {
  return n * 2
}

add := func(n int) int {
  return n + 100
}

sub := func(n int) int {
  return n - 1
}

crunch(nums, mul, add, sub)
```

[![run the code](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/run_the_code.png)](https://play.golang.org/p/iqcumj5cka	"use annoymous func")

---

### 高阶函数

高阶函数可以接收或返回一个甚至多个函数。本质上来来讲，它用其他函数来完成工作。

![hight-order funcs](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/higher-order_funcs.png)

下面闭包单元中的 split 函数就是一个高阶函数。它的返回结果是一个 tokenizer 类型的函数。

---

### 闭包

闭包可以记住其上下文环境中所有定义过的变量。闭包的一个好处就是随时可以在其捕获的环境下操作其中的变量——*小心内存泄漏！*

#### 示例

声明一个新的函数类型，它返回一个已分割的字符串的下一个单词：

```go
type tokenizer func() (token string, ok bool)
```

下面的 split 函数是一个**高阶函数**，它根据指定的分割符来分割一个字符串，然后返回一个可以遍历这个被分割的字符串中所有单词的**闭包**。*这个闭包可以使用 ”token“ 和 ”last“ 两个在其捕获的环境下定义的变量。*

![cloure](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/closure.png)

#### 小试牛刀：

```go
const sentence = "The quick brown fox jumps over the lazy dog"

iter := split(sentence, " ")

for {
  token, ok := iter()
  if !ok { break }
  
  fmt.Println(token)
}
```

* 在这里，我们使用了 split 函数将一句话分割成了若干个单词，然后得到了一个*迭代器函数*，并将它赋值给 iter 变量
* 然后，我开始了一个当 iter 函数返回 false 的时候才停止的无限循环
* 每次调用 iter 都能返回下一个单词

#### 结果：

```
The
quick
brown
fox
jumps
over
the
lazy
dog
```

[![run the code](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/run_the_code.png)](https://play.golang.org/p/AI1_5BkO1d	"closure")

<p align="center">再次提示，这里面有更详细的描述哦~</p>

---

### 延时函数 （defer funcs）

延时函数只在其父函数返回时被调用。多个延时函数会以栈的形式一个接一个被调用。

*👉* ***我在[另一篇文章](https://blog.learngoprogramming.com/golang-defer-simplified-77d3b2b817ff)中对延时函数有详细介绍***

![defer func](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/defer_funcs.png)

---

### 并发函数

`go func()` 会与其他 goroutines 并发执行。

*goroutine 是一种轻量级的线程机制，它能使你方便快捷的安排并发体系。其中，main 函数在 main-goroutine 中执行。*

#### 示例

这里，“start” 匿名函数通过 “go” 关键字进行调用，不会阻塞父函数的执行：

```go
start := func() {
  time.Sleep(2 * time.Second)
  fmt.Println("concurrent func: ends")
}

go start()

fmt.Println("main: continues...")
time.Sleep(5 * time.Second)
fmt.Println("main: ends")
```

#### 输出

```
main: continues...
concurrent func: ends
main: ends
```

![concurrent funs](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/concurrent_funcs.png)

<p align="center"><i>如果 main 函数中没有睡眠等阻塞调用，那么，main 函数会终止，而不会等待并发函数执行完。</i></p>

```
main: continues...
main: ends
```

[![run the code](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-functions-overview/run_the_code.png)](https://play.golang.org/p/UzbtrKxBna	"concurrent")

---

### 其他类型

#### 递归函数

你能在任意一门语言中使用递归函数，Go 语言中的递归函数实现与它们也没有本质上的区别。然而，你可别忘了每一次的函数调用通常都会创建一个[调用栈](https://en.wikipedia.org/wiki/Call_stack#Functions_of_the_call_stack)。但在 Go 中，栈是动态的，它们能根据相应函数的需要进行增减。如果你可以不使用递归解决手上的问题，那最好。

#### 黑洞函数

黑洞函数能被多次定义，并且不能用通常的方式进行调用。它们在测试解析器的时候有时会非常有用：看[这里](https://github.com/golang/tools/blob/master/imports/imports.go#L167)

```go
func _() {}
func _() {}
```

#### 内联函数

Go 语言的链接器会将函数放置到可执行环境中，以便稍后在运行时调用它。与直接执行代码相比，有时调用函数是一项昂贵的操作。所以，编译器将函数的主体注入调用者函数中。

更多的相关资料请参阅：[这里](https://github.com/golang/proposal/blob/master/design/19348-midstack-inlining.md)、[这里](http://www.agardner.me/golang/garbage/collection/gc/escape/analysis/2015/10/18/go-escape-analysis.html)、[这里](https://medium.com/@felipedutratine/does-golang-inline-functions-b41ee2d743fa)和[这里](https://github.com/golang/go/issues/17373)。

#### 外部函数

如果你省略掉函数体，仅仅进行函数声明，连接器会尝试在任何可能的地方找到这个外部函数。例如：Atan Func在[*这里只进行了声明*](https://github.com/golang/go/blob/dd8dc6f0595ffc2c4951c0ce8ff6b63228effd97/src/pkg/math/atan.go#L54)，而后在[*这里进行了实现*](https://github.com/golang/go/blob/dd8dc6f0595ffc2c4951c0ce8ff6b63228effd97/src/pkg/math/atan_386.s)。

---

via: https://blog.learngoprogramming.com/go-functions-overview-anonymous-closures-higher-order-deferred-concurrent-6799008dde7b

作者：[Inanc Gumus](https://blog.learngoprogramming.com/@inanc)
译者：[shockw4ver](https://github.com/shockw4ver)
校对：[rxcai](https://github.com/rxcai)、[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
