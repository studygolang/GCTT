首发于：https://studygolang.com/articles/16574

# Go 方法（第二部分）

这篇文章介绍了关于 Go 语言中方法的剩余部分。强烈建议先阅读[第一部分](https://studygolang.com/articles/14061) 的介绍部分。

## 方法表达式

如果有这样一个类型 T，它的方法集中包含方法 M，则 T.M 会生成一个与方法 M 几乎相同且带有签名的方法，这称为 *方法表达式*。不同之处在于，它额外附带的第一个参数与 M 的接收者类型相等。

```go
package main

import (
	"fmt"
	"reflect"
)

func PrintFunction(val interface{}) {
	t := reflect.TypeOf(val)
	fmt.Printf("Is variadic: %v\n", t.IsVariadic())
	for i := 0; i < t.NumIn(); i++ {
		fmt.Printf("Parameter #%v: %v\n", i, t.In(i))
	}
}

type T struct{}

func (t T) M(text string, number int) {}
func (t *T) N(map[string]int)         {}
func main() {
	PrintFunction(T.M)
	PrintFunction((*T).M)
	PrintFunction((*T).N)
}

```

输出：

```
Is variadic: false
Parameter #0: main.T
Parameter #1: string
Parameter #2: int
Is variadic: false
Parameter #0: *main.T
Parameter #1: string
Parameter #2: int
Is variadic: false
Parameter #0: *main.T
Parameter #1: map[string]int
```

如果方法 M 不在类型 T 的方法集中，使用表达式 `T.M` 会导致错误 `invalid method expression T.N (needs pointer receiver: (*T).N)`。

在上面的片段中，有一个有趣的案例 `PrintFunction((*T).M)`，即使方法 M 拥有的是值接收器，它仍然使用 `*main.T` 的第一个参数创建方法。Go 的运行时会在后台传递指针，创建副本并传递给方法。使用这种方式，方法无法访问原始值。

```go
type T struct {
	name string
}

func (t T) M() {
	t.name = "changed"
}
func (t *T) N() {
	t.name = "changed"
}
func main() {
	t := T{name: "Michał"}
	(*T).M(&t)
	fmt.Println(t.name)
	(*T).N(&t)
	fmt.Println(t.name)
}
```

输出：

```go
Michał
changed
```

可以从接口类型创建方法表达式：

```go
package main

import "fmt"

type T struct {
	name string
}

func (t T) M() {
	fmt.Println(t.name)
}

type I interface {
	M()
}

func main() {
	t1 := T{name: "Michał"}
	t2 := T{name: "Tom"}
	m := I.M
	m(t1)
	m(t2)
}

```

输出：

```
Michał
Tom
```

## 方法值

与类型和*方法表达式*类似，使用表达式可以得到一个带有接收器的方法，这就是*方法值*。如果有表达式 *x*，则 *x.M* 和方法 M 一样可以使用同样的参数调用。当然，方法 M 需要在类型 x 的方法集中，如果 x 是可寻址类型，M 应该在类型 `&x` 的方法集中。

```go
type T struct {
	name string
}

func (t *T) M(string) {}
func (t T) N(float64) {}
func main() {
	t := T{name: "Michał"}
	m := t.M
	n := t.N
	m("foo")
	n(1.1)
}
```

## 提升方法

如果一个结构包含内嵌（匿名）的属性，那么这个属性的方法也处于该结构类型的方法集中。

```go
package main

import "fmt"

type T struct {
	name string
}

func (t T) M() string {
	return t.name
}

type U struct {
	T
}

func main() {
	u := U{T{name: "Michał"}}
	fmt.Println(u.M())
}

```

上面的 Go 程序输出 `Michał` 是完全正确的。说嵌入到结构类型中属性的方法属于该类型的方法集是有确切原因的：

### #1

如果结构类型 U 包含了内嵌属性 T，那么方法集 S 和 `*S` 包含带有接收器 T 的提升方法。另外，方法集 `*S` 包含的是带有接收器 `*T` 的提升方法。

```go
package main

import (
	"fmt"
	"reflect"
)

func PrintMethodSet(val interface{}) {
	t := reflect.TypeOf(val)
	fmt.Printf("Method set count: %d\n", t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		fmt.Printf("Method: %s\n", m.Name)
	}
}

type T struct {
	name string
}

func (t T) M()  {}
func (t *T) N() {}

type U struct {
	T
}

func main() {
	u := U{T{name: "Michał"}}
	PrintMethodSet(u)
	PrintMethodSet(&u)
}

```

上面的程序输出：

```go
Method set count: 1
Method: M
Method set count: 2
Method: M
Method: N
```

从本文介绍的第一部分，我们应当知晓的是语言规范中的附加调用规则：

> 如果 x 是可寻址的，并且 &x 的方法集中包含 m，(&x).m() 可以简写为 x.m()。

所以尽管方法 N 不是类型 U 的方法集的一部分，我们仍可以使用 `u.N()` 这样的调用。

### #2

如果结构类型 U 包含内嵌属性 `*T`，那么方法集 S 和 `*S` 中包带有接收器 T 和 `*T` 的提升方法。

```go
type T struct {
	name string
}

func (t T) M()  {}
func (t *T) N() {}

type U struct {
	*T
}

func main() {
	u := U{&T{name: "Michał"}}
	PrintMethodSet(u)
	PrintMethodSet(&u)
}
```

打印：

```
Method set count: 2
Method: M
Method: N
Method set count: 2
Method: M
Method: N
```

---

via: https://medium.com/golangspec/methods-in-go-part-ii-2b4cc42c5cb6

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[Tyrodw](https://github.com/tyrodw)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
