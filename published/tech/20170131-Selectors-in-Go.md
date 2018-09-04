首发于：https://studygolang.com/articles/14628

# Go 语言中的选择器

在 Go 语言中，表达式 `foo.bar` 可能表示两件事。如果 *foo* 是一个包名，那么表达式就是一个所谓的`限定标识符`，用来引用包 *foo* 中的导出的标识符。由于它只用来处理导出的标识符，*bar* 必须以大写字母开头(译注：如果首字母大写，则可以被其他的包访问；如果首字母小写，则只能在本包中使用）：

```go
package foo
import "fmt"
func Foo() {
    fmt.Println("foo")
}
func bar() {
    fmt.Println("bar")
}

package main
import "github.com/mlowicki/foo"
func main() {
    foo.Foo()
}
```

这样的程序会工作正常。但是（主函数）调用 `foo.bar()` 会在编译时报错 —— `cannot refer to unexported name foo.bar(无法引用未导出的名称 foo.bar)`。

如果 *foo* 不是 一个包名，那么 `foo.bar` 就是一个选择器表达式。它访问 *foo* 表达式的字段或方法。点之后的标识符被称为 *selector*（选择器）。关于首字母大写的规则并不适用于这里。它允许从定义了 *foo* 类型的包中选择未导出的字段或方法：

```go
package main
import "fmt"
type T struct {
    age byte
}
func main() {
    fmt.Println(T{age: 30}.age)
}
```

该[程序](https://play.golang.org/p/DwQFPZ3bG7)打印：`30`

## 选择器的深度

语言规范定义了选择器的 *depth*（深度）。让我们来看看它是如何工作的吧。选择器表达式 `foo.bar` 可以表示定义在 *foo* 类型的字段或方法或者定义在 *foo* 类型中的匿名字段：

```go
type E struct {
    name string
}
func (e E) SayHi() {
    fmt.Printf("Hi %s!\n", e.name)
}
type T struct {
    age byte
    E
}
func (t T) IsStillYoung() bool {
    return t.age <= 18
}
func main() {
    t := T{30, E{"Michał"}}
    fmt.Println(t.IsStillYoung()) // false
    fmt.Println(t.age) // 30
    t.SayHi() // Hi Michał!
    fmt.Println(t.name) // Michał
}
```

在上面的[代码](https://play.golang.org/p/GWbEzILDdg)中，我们可以看到可以调用方法或者访问定义在嵌入字段中字段。字段 `t.name` 和方法 `t.SayHi` 都被提升了，这是因为类型 *E* 嵌套在 *T* 的定义中：

```go
type T struct {
    age byte
    E
}
```

定义在类型 *T* 中表示字段或类型的选择器深度为 0（译注：表示在类型 T 中定义的字段或方法的选择器的深度为 0）。如果字段或方法定义在嵌入（也就是 匿名）字段，那么深度等于匿名字段遍历这样字段或方法的数量。在上一个片段中，*age* 字段深度是 0，因为它在 *T* 中声明，但是因为 *E* 是放在 *T* 中，*name* 或者 *SayHi* 的深度是 1。让我们来看看更复杂的[例子](https://play.golang.org/p/8-8xi_JpaU)：

```go
package main
import "fmt"
type A struct {
    a string
}
type B struct {
    b string
    A
}
type C struct {
    c string
    B
}
func main() {
    v := C{"c", B{"b", A{"a"}}}
    fmt.Println(v.c) // c
    fmt.Println(v.b) // b
    fmt.Println(v.a) // a
}
```

* *c* 的深度是 `v.c`，其值为 0。这是因为字段是在 *C* 中声明的
* `v.b` 中 *b* 的深度是 1。这是因为它的字段定义在类型 *B* 中，其（类型B）又嵌入在 *C* 中
* `v.a` 中 *a* 的深度是 2。这是因为需要遍历两个匿名字段（*B* 和 *A*）才能访问它

## 有效选择器

go 语言中有关哪些选择器有效，哪些无效有着明确规则。让我们来深入了解他们。

### 唯一性+最浅深度

当 *T* 不是指针或者接口类型，第一条规则适用于类型 `T` 与 `*T`。选择器 *foo.bar* 表示字段和方法在定义了 *bar* 的类型 *T* 中的最浅深度。在这样的深度，恰好可以定义一个（唯一的）这样的字段或者方法（[源代码](https://play.golang.org/p/mGtRxnrAQR)）：

```go
type A struct {
    B
    C
}
type B struct {
    age byte
    name string
}
type C struct {
    age byte
    D
}
type D struct {
    name string
}
func main() {
    a := A{B{1, "b"}, C{2, D{"d"}}}
    fmt.Println(a) // {{1 b} {2 {d}}}
    // fmt.Println(a.age) ambiguous selector a.age
    fmt.Println(a.name) // b
}
```

类型嵌入的结构如下：

```
 A
 / \
B   C
     \
      D
```

选择器 *a.name* 是有效的，并且表示字段 *name*（*B* 类型内）的深度为 1。*C* 类型中的字段 *name* 是 “shadowed(浅的）”。有关 *age* 字段则是不同的。在深度 1 处有这样两个字段（在 *B* 和 *C* 类型中），所以编译器会抛出 `ambiguous selector a.age` 错误。

当被提升的字段或方法有歧义时，Gopher 仍然可以使用完整的选择器。

```go
fmt.Println(a.B.name)   // b
fmt.Println(a.C.D.name) // d
fmt.Println(a.C.name)   // d
```

值得重申的是，该规则也适用于 `*T` —— [例子](https://play.golang.org/p/8AfF4ie3HB)。

### 空指针

```go
package main
import "fmt"
type T struct {
    num int
}
func (t T) m() {}
func main() {
    var p *T
    fmt.Println(p.num)
    p.m()
}
```

如果选择器是有效的，但 *foo* 是一个空指针，那么评估 *foo.bar* 造成 runtime panic：`panic invalid memory address or nil pointer dereference`（[源代码](https://play.golang.org/p/hxiU6S8jTS)）

### 接口

如果 *foo* 是一个接口类型值，那么 *foo.bar* 实际上是 *foo* 的动态值的一个方法：

```go
type I interface {
    m()
}
type T struct{}
func (T) m() {
    fmt.Println("I’m alive!")
}
func main() {
    var i I
    i = T{}
    i.m()
}
```

上面的[片段](https://play.golang.org/p/j8zo9Th2N0)输出 `I'm alive!`。当然，调用不在接口的方法集合中的方法时，会产生编译时错误，如 `i.f undefined (type I has no field or method f)`

如果 *foo* 为 *nil*，那么它将会导致一个运行时错误：

```go
type I interface {
    f()
}
func main() {
    var i I
    i.f()
}
```

这样的[程序](https://play.golang.org/p/noOpOVwpV_)将会因为错误 `panic: runtime error: invalid memory address or nil pointer dereference` 而崩溃。这和空指针情况类似，而且由于诸如没有值赋值和接口[零值](https://golang.org/ref/spec#The_zero_value)为 *nil* 而发生错误。

### 一个特殊情况

除了到现在为止关于有效选择器的描述外，这还有一个场景：假设这里有一个命名指针类型：

```go
type P *T
```

类型 *P* 的[方法集](https://golang.org/ref/spec#Method_sets)不包含类型 *T* 的任何方法。如果有类型 *P* 的变量，则无法调用任何 *T* 的方法。但是，规范允许选择类型 *T* 的字段（非方法）（[源代码](https://play.golang.org/p/7wJI4F34ij)）：

```go
type T struct {
    num int
}
func (t T) m() {}
type P *T
func main() {
    var p P = &T{num: 10}
    fmt.Println(p.num)
    // p.m() // compile-time error: p.m undefined (type P has no field or method m)
    (*p).m()
}
```

`p.num` 在 hood 下被转化为 `(*p).num`。

## 在 hood 下

如果你对选择器朝朝和验证的实际实现感兴趣的话，请查看 [selector](https://github.com/golang/go/blob/6bdb0c11c73ecf2337918d784c54f9dda2207ca7/src/go/types/call.go#L341) 和 [LookupFieldOrMethod](https://github.com/golang/go/blob/6bdb0c11c73ecf2337918d784c54f9dda2207ca7/src/go/types/lookup.go) 函数。[这里](https://play.golang.org/p/hjGWpBor2l)是最后一个使用的例子。

---

via：https://medium.com/golangspec/selectors-in-go-c53a016702cf

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[cureking](https://github.com/cureking)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
