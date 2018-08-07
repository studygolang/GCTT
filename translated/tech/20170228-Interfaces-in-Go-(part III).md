## Go接口（第三部分）Interfaces in Go (part III)
This story introduces another set of interfaces-related topics in Golang.
It explains things like method expression derived from interface types,
interface type values as map keys or interfaces of embedded fields.
本文是golang接口相关主题的另一部分。
讲解的内容包括接口类型中的方法，
接口类型的值作为map结构中的key,或者作为结构体的内置字段。


## 方法和接口 Method expression and interface type
Go has the concept of method expressions.
It’s a way to get function from method set of type T.
Such function has additional,
explicit parameter of type T (source code):
Go 是有方法的概念的。
可以通过在调用类型T中的方法来获的一个函数，
这个函数可以从类型T中额外地获得明确的参数。

```go

type T struct {
    name string
}

func (t *T) SayHi() {
    fmt.Printf("Hi, my name is %s\n", t.name)
}

func main() {
    t := &T{"foo"}
    f := (*T).SayHi
    f(t) // Hi, my name is foo
}
```
Language specification allows to use method expressions also for interface types:
>It is legal to derive a function value from a method of an interface type.
The resulting function takes an explicit receiver of that interface type.

Let’s see an example (source code):


golang语言规范同样允许接口类型使用方法：
>在一个接口类型中，通过方法来获取函数是符合规范的。
从中获得的函数拥有一个明确的接受者。

让我们来看一个例子：
``` go

type I interface {
    M(name string)
}
type T struct {
    name string
}
func (t *T) M(name string) {
    t.name = name
}
func main() {
    f := I.M
    var i I = &T{"foo"}
    f(i, "bar")
    fmt.Println(i.(*T).name) // bar
}

```


## 接受接口类型 Receiver of interface type
Go allows to define methods — functions with receiver of particular type (source code):
Go允许定义方法 —— 接受了特定类型的函数（源代码）：
``` go

type T struct {
    name string
}
func (t *T) SayHi() {
    fmt.Printf("Hi, my name is %s\n", t.name)
}
func main() {
    t1 := T{"foo"}
    t1.SayHi() // Hi, my name is foo
    t2 := &T{"bar"}
    t2.SayHi() // Hi, my name is bar

```
Method is added to receiver’s type (in above snippet this type is *T).
Such method can be called on values of type *T (or T).
The takeaway from this section is the fact that interfaces cannot be used as receiver’s type (source code):
方法被添加了一个接受类型（上面的代码片段的类型时*T）。
这个方法可以被类型为*T或者T的调用。
在这里顺便提一下，接口类型是不可以作为函数的接受者。
（源代码):
``` golang

type I interface {}
func (I) M() {}

```
It throws a compile-time error invalid receiver type I (I is an interface type).
More on methods in two stories introducing this topic thoroughly — part I and part II.
这段代码将抛出一个错误`invalid receiver type I (I is an interface type)`.
在第一部分和第二部分中有更多的方法进行介绍。
## 接口中“继承"
Interface is satisfied by struct type even if implemented method(s) is promoted so it comes from embedded (anonymous) field (source code):

``` go

type T1 struct {
    field1 string
}
func (t *T1) M() {
    t.field1 = t.field1 + t.field1
}
type T2 struct {
    field2 string
    T1
}
type I interface {
    M()
}
func main() {
    var i I = &T2{"foo", T1{field1: "bar"}}
    i.M() 
    fmt.Println(i.(*T2).field1) // barbar
}

```
In this case type *T2 implements interface I.
Method M is implemented by type *T1 which is an embedded field of type T2.
More on promoted fields and methods in older post.
在这个实例中，类型*T2实现了接口I.
被*T1实现的方法M作为了T2的一个内置的一部分。
在过去的文章里有更多关于
## type 可做map中的key或者value
Map (hash table under the hood as of Go ≤ 1.8) is a data structure which for defined keys holds some values (source code):
map 是一个由key-value组成的数据结构。(go1.8之前版本，map底层是通过哈系表实现的)
```go

counters := make(map[string]int64)
counters["foo"] = 1
counters["bar"] += 2
fmt.Println(counters) // map[foo:1 bar:2]
delete(counters, "bar")
fmt.Println(counters) // map[foo:1]
fmt.Println(counters["bar"]) // 0
if _, ok := counters["bar"]; !ok {
    fmt.Println("'bar' not found")
}
counters["bar"] = 2
for key, value := range counters {
    fmt.Printf("%s: %v\n", key, value) // order is randomized!
}
```
Interface type values can be used as both keys and values inside maps (source code):
接口类型的值作为map中的key和value来使用。
```go

type T1 struct {
    name string
}
func (t T1) M() {}
type T2 struct {
    name string
}
func (t T2) M() {}
type I interface {
    M()
}
func main() {
    m := make(map[I]int)
    var i1 I = T1{"foo"}
    var i2 I = T2{"bar"}
    m[i1] = 1
    m[i2] = 2
    fmt.Println(m) // map[{foo}:1 {bar}:2]
}

```


## 无所不在的类型

### error
在go中内置的error是一个接口类型：
The built-in error in Go is an interface:
```go
type error interface {
	Error() string
}

```
任何类型实现了Error方法，并且此方法没有参数且返回一个string类型的值，那么这个方法就实现了这个接口
Every type implementing Error method not taking any parameters and returning string value as a result, satisfies this interface (source code):
```go

import "fmt"
type MyError struct {
    description string
}
func (err MyError) Error() string {
    return fmt.Sprintf("error: %s", err.description)
}
func f() error {
    return MyError{"foo"}
}
func main() {
    fmt.Printf("%v\n", f()) // error: foo
}
```

### io.Writer
io.Writer接口仅仅含有一个方法 —— Write:
``` go

Write(p []byte) (n int, err error)

```

If anything abnormal will happen then error won’t be nil.
Interface error is the same interface described in preceding section.
Writer interface is used throughout the standard library for things like MultiWriter,TeeReader, net/http and many more.
如果有任何异常发生，返回的error就将不会是nil。
error接口在前一节同样做了描述。
Writer接口在标准库中到处都有被用到，比如MultiWriter、TeeReder、net/http，还有很多其他的地方。



Click ❤ below to help others discover this story.
 Please follow me if you want to get updates about new posts or boost work on future stories.
点击下面的 ❤ 去帮助其他人发现这篇文章。
## Resources
>
###


---

via: https://medium.com/golangspec/interfaces-in-go-part-iii-61f5e7c52fb5

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[译者ID](https://github.com/xmge)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

