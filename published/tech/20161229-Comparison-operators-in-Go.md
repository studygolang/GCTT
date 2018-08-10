首发于：https://studygolang.com/articles/14097

# Go 语言中的比较操作符

这篇文章专注于 6 个操作符，==，!=，<，<=，> 和 >=。我们将深入探讨它们的语法和用法的细微差别。对很多人来说，这听起来不像是吸引人的事，或者他们可能已经从其他编程语言获得了糟糕的经验。然而，在 Go 中它们定义的很好并简洁。下面讨论的主题，如可比性将出现在其他场合，如 maps。为了使用上述操作符，至少有一个操作数需要[可赋值](https://studygolang.com/articles/12381)给第二个操作数:

```go
package main
import "fmt"
type T struct {
    name string
}
func main() {
    s := struct{ name string }{"foo"}
    t := T{"foo"}
    fmt.Println(s == t)  // true
}
```

这条规则显著缩小了可选范围：

```go
var a int = 1
var b rune = '1'
fmt.Println(a == b)
```

类似的代码在 Javascript 或 Python 中可以运行。但在 Go 中它是非法的，并且在编译时会被检测到。

```
src/github.com/mlowicki/lab/lab.go:8: invalid operation: a == b (mismatched types int and rune)
```

可赋值不是唯一要求。这是相等和顺序操作符的规则……

## 相等操作符

操作数需要使用 `==` 或 `!=` 操作符进行比较。哪些方法，哪些值可以被比较？Go 规范定义的非常明确：

* boolean 值可比较（如果俩个值都是真或假，那么比较结果被认为 true）
* 整数和浮点数比较：

```go
var a int = 1
var b int = 2
var c float32 = 3.3
var d float32 = 4.4
fmt.Println(a == b) // false
fmt.Println(c == d) // false
```

当编译时 `a == d` 会抛出异常（ int 和 float32 类型不匹配）因为它不可能用 int 和 float 比较。

* 复数相等，如果他们的是实数和虚数部分都相等：

```go
var a complex64 = 1 + 1i
var b complex64 = 1 + 2i
var c complex64 = 1 + 2i
fmt.Println(a == b)  // false
fmt.Println(b == c)  // true
```

* 字符串类型值可比较
* 指针类型值相等，如果他们都是 nil 或都指向相同的变量：

```go
type T struct {
    name string
}
func main() {
    t1 := T{"foo"}
    t2 := T{"bar"}
    p1 := &t1
    p2 := &t1
    p3 := &t2
    fmt.Println(p1 == p2)   // true
    fmt.Println(p2 == p3)   // false
    fmt.Println(p3 == nil)  // false
}
```

> 不同的 zero-size 变量可能具有相同的内存地址，因此我们不假设任何指向这些变量的指针相等。

```go
a1 := [0]int{}
a2 := [0]int{}
p1 := &a1
p2 := &a2
fmt.Println(p1 == p2)  // might be true or false. Don't rely on it!
```

* 通道类型值相等，如果他们确实一样（被相同的内置 make 方法创建）或值都是 nil：

```go
ch1 := make(chan int)
ch2 := make(chan int)
fmt.Println(ch1 == ch2)  // false
```

* 接口类型是可比较。与通道和指针类型值比较一样，如果是 nil 或 动态类型和动态值是相同的：

```go
type I interface {
    m()
}
type J interface {
    m()
}
type T struct {
    name string
}
func (T) m() {}
type U struct {
    name string
}
func (U) m() {}
func main() {
    var i1, i2, i3, i4 I
    var j1 J
    i1 = T{"foo"}
    i2 = T{"foo"}
    i3 = T{"bar"}
    i4 = U{"foo"}
    fmt.Println(i1 == i2)  // true
    fmt.Println(i1 == i3)  // false
    fmt.Println(i1 == i4)  // false
    fmt.Println(i1 == j1)  // false
}
```

> 比较接口类型的方法集不能相交。

接口类型 I 的 i 和 非接口类型 T 的 t 可比较，如果 T 实现了 I 则 T 类型的值是可比较的。如果 I 的 动态类型和 T 是相同的，并且 i 的动态值和 t 也是相同的，那么值是相等的：

```go
type I interface {
    m()
}
type T struct{}
func (T) m() {}
type S struct{}
func (S) m() {}
func main() {
    t := T{}
    s := S{}
    var i I
    i = T{}
    fmt.Println(t == i)  // true
    fmt.Println(s == i)  // false
}
```

* 结构类型可比较，所以字段都需要比较。所有非空白字段相等则他们等。

```go
a := struct {
    name string
    _ int32
}{name: "foo"}
b := struct {
    name string
    _ int32
}{name: "foo"}
fmt.Println(a == b)  // true
```

* Go 中 数组是同质的 —— 只有同一类型（数组元素类型）的值可以被存储其中。对于数组值比较，它们的元素类型需要可比较。如果对应的元素相同，数组就相等。

就是这样。上面列表很长但并不充满惊奇。尝试了解它在 JavaScript 是如何工作的……

有三种类型不能比较 —— maps, slices 和 functions。Go 编译器不允许这样做，并且编译比较 maps 的程序会引起一个错误 **map can only be compared to nil.**。展示的错误告诉我们至少可以用 maps，slices 或 functions 和 nil 比较。

目前为止，我们知道接口值是可比较的，但 maps 是不可以的。如果接口值的动态类型是相同的，但是不能比较（如 maps），它会引起一个运行时错误：

```go
type T struct {
    meta map[string]string
}
func (T) m() {}
func main() {
    var i1 I = T{}
    var i2 I = T{}
    fmt.Println(i1 == i2)
}
```

```
panic: runtime error: comparing uncomparable type main.T
goroutine 1 [running]:
panic(0x8f060, 0x4201a2030)
    /usr/local/go/src/runtime/panic.go:500 +0x1a1
main.main()
    ...
```

## 顺序操作符

这些操作符只能应用在三种类型：整数，浮点数和字符串类型。这没有什么特别的或 Go 特有的。值得注意的是字符串是按字典顺序排列的。byte-wise 一次一个字节并没有 Collation 算法。

```go
fmt.Println("aaa" < "b") // true
fmt.Println("ł" > "z")   // true
```

## 结果

任何比较操作符的结果都是无类型布尔常量（true 或 false）。因为它没有类型，所以可以分配了给任何布尔变量：

```go
var t T = true
t = 3.3 < 5
fmt.Println(t)
```

这段代码输出 true。另一个，尝试分配 bool 类型的值：

```go
var t T = true
var b bool = true
t = b
fmt.Println(t)
```

产生一个错误，不能使用 b （bool类型）分配给 T 类型。关于常量（有类型和无类型）更详尽的介绍在官方[博客](https://blog.golang.org/constants)上。

---

via: https://medium.com/golangspec/comparison-operators-in-go-910d9d788ec0

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[themoonbear](https://github.com/themoonbear)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
