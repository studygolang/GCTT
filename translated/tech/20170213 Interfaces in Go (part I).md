# Go语言接口（第一部分）

 <!--![avatar](https://cdn-images-1.medium.com/max/800/1*gTWySSARxfkaQVxCBSOKoA.jpeg) -->

接口提升了代码的弹性与拓展性，同时它也是 go 语言实现多态的一种方式。接口允许通过一些必要的行为来实现，而不再要求设置特定类型。而这个行为就是通过一些方法设置来定义的：

```<lang>
type I interface {
    f1(name string)
    f2(name string) (error, float32)
    f3() int64
}
```

不需要特定的实现。
只要通过定义 type 中包含目标名与签名（输入与输出的参数列表）的方法来表明其实现（满足）了一个接口就足够了：

```<lang>
type T int64
func (T) f1(name string) {
    fmt.Println(name)
}
func (T) f2(name string) (error, float32) {
    return nil, 10.2
}
func (T) f3() int64 {
    return 10
}
```

类型 T 实现了第一个程序段的接口 I。举个例子，类型 T 的值可以传递给任何接受 I 作为参数的函数（[源代码](https://play.golang.org/p/aUyEa-HgYi)）：

```
type I interface {
    M() string
}
type T struct {
    name string
}
func (t T) M() string {
    return t.name
}
func Hello(i I) {
    fmt.Printf("Hi, my name is %s\n", i.M())
}
func main() {
    Hello(T{name: "Michał"}) // "Hi, my name is Michał"
}
```

在 function Hello 中，方法调用了 `i.M()`。 这个过程概括一下就是，只要来自不同 type 的方法是通过 type 来实现 interface I，就可以被调用。

go 语言的突出特点就是其 interface 是隐式实现的。程序员不需要指定 type T 实现了 interface I。这个工作由 go 的编译器完成（不需要派一个人去做机器的工作）。这种行为中的实现方式之所以很赞，是因为定义 interface 这件事情是由已经写好的 type 自动实现的（不需要为之做任何改变）。

之所以 interface 可以提供弹性，是因为任意一个 type 可以实现多个 interface ([代码](https://play.golang.org/p/cN6KrJab-l)):

```
type I1 interface {
    M1()
}
type I2 interface {
    M2()
}
type T struct{}
func (T) M1() { fmt.Println("T.M1") }
func (T) M2() { fmt.Println("T.M2") }
func f1(i I1) { i.M1() }
func f2(i I2) { i.M2() }
func main() {
    t := T{}
    f1(t) // "T.M1"
    f2(t) // "T.M2"
}
```

或者同样的 interface 可以实现多个 type ([源代码](https://play.golang.org/p/_7mkHdEilz)):
```
type I interface {
    M()
}
type T1 struct{}
func (T1) M() { fmt.Println("T1.M") }
type T2 struct{}
func (T2) M() { fmt.Println("T2.M") }
func f(i I) { i.M() }
func main() {
    f(T1{}) // "T1.M"
    f(T2{}) // "T2.M"
}
```

>*而且除了一个或多个 interface 所需要的方法外，type 可以自由地实现其他方法*

<div style="text-align:center;font-weight:800;">.  &nbsp&nbsp . &nbsp&nbsp .</div>

在 go 中，我们有两个与 interface 相关的概念：
1. 接口-通过[关键字](https://golang.org/ref/spec#Keywords) `interface`,实现此类接口所需要的一组方法；
2. 接口类型-接口类型的变量，可以保存一些实现于特定接口的值。

让我们在接下来的两节中讨论这些主题。

## 定义一个接口
接口类型的声明指定属于它（译注：接口）的方法。方法是通过它的名字（译注：方法名）和签名-输入和接口参数定义的：

```
type I interface {
    m1()
    m2(int)
    m3(int) int
    m4() int
}
```

除了方法外，它还允许嵌入其他接口-在同一个包中定义或引入-通过[限定名](https://golang.org/ref/spec#QualifiedIdent)。它从嵌入的接口中添加所有方法：
```
import "fmt"
type I interface {
     m1()
}
type J interface {
    m2()
    I
    fmt.Stringer
}
```

接口 J 的方法组包括：
* m1() 来自嵌入的接口 I
* m2()
* String() string（来自嵌入的接口 [Stringer](https://golang.org/pkg/fmt/#Stringer)）

顺序无关紧要，所以方法规格与嵌入的接口类型完全可以交错。

>*添加了来自嵌入接口类型的导出方法（以大写字母开头）和非导出方法（以小写字母开头）*

如果我嵌入一个接口 J，接口 J 又嵌入接口 K，那么 K 中的所有方法也会被添加到 I 中：
```
type I interface {
    J
    i()
}
type J interface {
    K
    j()
}
type K interface {
    k()
}
```

I 的方法组包括 `i()` ,`j()` 和 `k()` ([源代码](https://play.golang.org/p/mz_8CMMDsn))。

不允许循环嵌入接口（译注：即 A 嵌入 B，B 嵌入 C，C 嵌入 A），并且在编译阶段，会检测接口的循环嵌入问题（[源代码](https://play.golang.org/p/CXf3-quH0A)）：
```
type I interface {
    J
    i()
}
type J interface {
    K
    j()
}
type K interface {
    k()
    I
}
```

编译器会提出一个错误 `interface type loop involving I`（译注：有关I的接口类型循环）。

接口方法必须有唯一名字（[源代码](https://play.golang.org/p/zt3t-GUrYU)）:
```
type I interface {
    J
    i()
}
type J interface {
    j()
    i(int)
}
```

否则将抛出编译时间错误：<span style="background:#dddddd;">duplicate method i</span>（译注：重复方法 i )。

接口的组成可以在标准库中找到。一个这样的例子就是 io.ReadWriter :

```
type ReadWriter interface {
    Reader
    Writer
}
```

我们知道如何创建一个新的接口。现在让我们学习接口类型的值...

## 接口类型的值

接口类型 I 的变量可以保持任何实现 I 的值（[源代码](https://play.golang.org/p/Zvaq5c97wp)）：
```
type I interface {
    method1()
}
type T struct{}
func (T) method1() {}
func main() {
    var i I = T{}
    fmt.Println(i)
}
```

这里我们有一个来自接口类型 I 的变量 i。

### 静态类型 VS 动态类型

在编译阶段，变量类型便已知。这是在声明时指定的，不再变化，并被称为静态类型（或只是类型）。接口类型的变量也有静态类型，其本身就是一个接口。它们还具有可以指定值的类型-动态类型（[源代码](https://play.golang.org/p/UVMqqMNsb8)）：
```
type I interface {
    M()
}
type T1 struct {}
func (T1) M() {}
type T2 struct {}
func (T2) M() {}
func main() {
    var i I = T1{}
    i = T2{}
    _ = i
}
```

变量 i 的静态类型是 I。这是不变的。另一方面，动态类型是...好吧，动态的。在首次分配后，i 的动态类型是 T1。这并不是一成不变的，所以 i 的动态类型第二次赋值为 T2。当接口类型值的值 nil (接口类型的零值）时，动态类型便不设置。

### 如何获取接口类型值得动态类型？

包 [reflect](https://golang.org/pkg/reflect/)可以用来获取这个（[源代码](https://play.golang.org/p/9cQ5JqSxL5)）：

```
fmt.Println(reflect.TypeOf(i).PkgPath(), reflect.TypeOf(i).Name())
fmt.Println(reflect.TypeOf(i).String())
```

通过包 [fmt](https://golang.org/pkg/fmt/) 以及格式动词 `%d` 也可以做到这点：

```
fmt.Printf("%T\n", i)
```

在 hood 下使用包 *reflect* 包，即便 i 是 nil 时，这个方法也有效。

### 空接口值

这次我们将从一个例子开始（[源代码](https://play.golang.org/p/kv9XUzIxBU)）：
```
type I interface {
    M()
}
type T struct {}
func (T) M() {}
func main() {
    var t *T
    if t == nil {
        fmt.Println("t is nil")
    } else {
        fmt.Println("t is not nil")
    }
    var i I = t
    if i == nil {
        fmt.Println("i is nil")
    } else {
        fmt.Println("i is not nil")
    }
}
```

输出：

```
t is nil
i is not nil
```

第一次看，会觉得很惊讶。变量 i 的值，我们明明设置为 nil，但是这里的值却不等于 nil。其实接口类型值包含两个组件：

* 动态类型
* 动态值

动态类型在之前（“静态类型VS动态类型”部分）已经讨论过了。动态值是指定的实际值。
在赋值 `var i I = t` 后的讨论段中，i 的动态值是 nil，但动态类型为\**T*在这个复制后，函数调用 `fmt.Printf("%T\n", i)`将会打印 `*main.T`。`当且仅当动态值与动态类型都为 nil 时，接口类型值为 nil。`结果就是即使接口类型值包含一个 nil 指针，这样的接口值也不是 nil。已知的错误就是返回未初始化，从函数返回接口类型为非接口类型值（[源代码](https://play.golang.org/p/4-M35Nc2JZ)）：

```
type I interface {}
type T struct {}
func F() I {
    var t *T
    if false { // not reachable but it actually sets value
        t = &T{}
    }
    return t
}
func main() {
    fmt.Printf("F() = %v\n", F())
    fmt.Printf("F() is nil: %v\n", F() == nil)
    fmt.Printf("type of F(): %T", F())
}
```

它打印出：

```
F() = <nil>
F() is nil: false
type of F(): *main.T
```

只是因为从函数返回的接口类型值有动态类型集（*main.T)。它并不等于 nil。

### 空接口

接口的方法集不必包含至少一个成员（译注：即方法集为空）。它完全可以是空的（[源代码](https://play.golang.org/p/V0GEG5nuW3)）：

```
type I interface {}
type T struct {}
func (T) M() {}
func main() {
    var i I = T{}
    _ = i
}
```

空接口可以自动满足任意类型-因此任意类型的值都可以赋值给这样的接口类型值。动态类型或静态类型的行为应用于空接口，就像应用于非空接口。
空接口的显著使用存在于参数可变函数 [fmt.Println](https://golang.org/pkg/fmt/#Println)。

<div style="text-align:center;font-weight:800;">.  &nbsp&nbsp . &nbsp&nbsp .</div>

## 满足一个接口

每个实现了接口所有方法的类型都自动满足这个接口。
我们不需要在这些类型中使用任何其他关键字（如 Java中的 implements）来表示该类型实现了接口。
它是由 go 语言的编译器自动实现的，而这儿正是该语言的强大之处（[源代码](https://play.golang.org/p/U4r6i2X5xb)）：
```
import (
    "fmt"
    "regexp"
)
type I interface {
    Find(b []byte) []byte
}
func f(i I) {
    fmt.Printf("%s\n", i.Find([]byte("abc")))
}
func main() {
    var re = regexp.MustCompile(`b`)
    f(re)
}
```

这里我们定义了一个由 [regexp.Regexp](https://golang.org/pkg/regexp/#Regexp) 类型实现的接口，该接口内置的 regexp 模块没有任何改变。

## 行为抽象

接口类型值**只**允许访问它自己的接口类型的方法。
如果它是 struct, array, scalar等，便会隐藏有关确切值的详情（[源代码](https://play.golang.org/p/kCjgQFCsL_)）：
```
type I interface {
    M1()
}
type T int64
func (T) M1() {}
func (T) M2() {}
func main() {
    var i I = T(10)
    i.M1()
    i.M2() // i.M2 undefined (type I has no field or method M2)
}
```

<div style="text-align:center;font-weight:800;">.  &nbsp&nbsp . &nbsp&nbsp .</div>

<!--点击下面的❤以帮助其他人发现这篇文章。 如果您想获得有关新帖子的更新或推动未来文章的工作，请关注我。-->

via: https://medium.com/golangspec/interfaces-in-go-part-i-4ae53a97479c

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[cureking](https://github.com/cureking)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出