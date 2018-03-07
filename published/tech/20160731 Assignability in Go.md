已发布：https://studygolang.com/articles/12381

# Go 语言中的可赋值性（Assignability in Go）

Go 是静态类型的编程语言。存储到变量中的值必须与变量的类型匹配。因此，不能像 Python 这种动态类型语言一样，随意的数据都可以作为值赋给变量。这个决定什么是允许赋值的规则就叫做可赋值性（Assignability）。

左边类型为 T 的变量，存在 6 种可以将右边的值赋给左边的情况。

## 1. 相同类型（Identical types）

第一种是非常明显的。如果右边的类型（也）是 T，则赋值是完全可以的。更多的细节可以参考 ["Identical types in Go"](https://medium.com/golangspec/assignability-in-go-27805bcd5874)。

## 2. 相同的基础类型（Identical underlying types）

在 Go 语言中，每种类型都有一种基础类型（underlying type）。对于布尔型，数字，字符串或者常量的基础类型都跟它们本身的类型相同。其他的基础类型来自于声明时的类型：

```go
type X map[string]int
var x X  // underlying type is map[string]int
```

可赋值的第二种情况是相同的基础类型：

```go
type X map[string]int
var x X
var y map[string]int
x = y
```

然而，如果有两个不同的命名类型（named types），则不能这么做：

```go
type X map[string]int
type Y map[string]int
var x X
var y Y
x = y  // cannot use y (type Y) as type X in assignment
```

附加的条件是要求至少一个类型不是一个命名类型。

Go 中的变量要么是命名类型（named）要么是非命名类型（unnamed）。非命名类型（unnamed types）是指使用类型字面意思（语言本身）定义的类型：

```go
var a [10]string
var b struct{ field string}
var c map[string]int
```

## 3. 将一个实现了接口 T 的变量赋值给 T 接口类型的变量

如果一个变量实现了接口 T，那么我们可以将这个变量赋值给一个 T 接口类型的变量。

```go
type Callable interface {
	f() int
}
type T int
func (t T) f() int {
	return int(t)
}
var c Callable
var t T
c = t
```

更多关于接口类型的细节都在 [语言规范（language spec）](https://golang.org/ref/spec#Interface_types)。

## 4. 将双向管道（channel）的变量赋值给相同类型的变量（Assigning bidirectional channel to variable with identical element types）

```go
type T chan<- map[string]int
var c1 T
var c2 chan map[string]int
c1 = c2
c2 = c1  // cannot use c1 (type T) as type chan map[string]int in assignment
```

跟第二种情况（相同的基础类型）一样，要求至少一种管道（channel）变量是非命名类型（unnamed type）：

```go
type T chan<- map[string]int
type T2 chan map[string]int
var c1 T
var c2 T2
c1 = c2  // cannot use c2 (type T2) as type T in assignment
```

## 5. 赋值 nil（Assigning nil）

允许将 nil 赋值给指针，函数，切片，map，管道，接口类型（的变量）。

```go
var a *int
var b func(int) int
var c []int
var d map[string]int
var e chan int
var f interface{}
a, b, c, d, e, f = nil, nil, nil, nil, nil, nil
var g [10]int
g = nil  // cannot use nil as type [10]int in assignment
```

## 6. 无（显式）指定类型的常量（Untyped constants）

关于 Go 常量更深入的介绍请查看[官方博客](https://blog.golang.org/constants)

无（显式）指定类型的常量可以被赋值给常量所代表的类型 T 相同的类型为 T 的变量。

```go
var a float32
var b float64
var c int32
var d int64
const untyped = 1
a = untyped
b = untyped
c = untyped
d = untyped
```

---

via: https://medium.com/golangspec/assignability-in-go-27805bcd5874

作者：[Michał Łowicki](https://twitter.com/mlowicki)
译者：[Miancai Li](https://github.com/gogeof)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

