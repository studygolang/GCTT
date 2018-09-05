![image](https://cdn-images-1.medium.com/max/800/1*LjLr800nNrw4G7Z_NCN1Tw.jpeg)

# Go接口（第三部分）
本文介绍的是golang接口主题的另一部分。
主要内容包括接口中的方法，
接口类型的值作为map中的key,或者作为内置字段。

## 方法和接口
Go 是有方法的概念的。
可以通过调用类型T中的方法来获得一个函数，
此函数可以从类型T中额外地获得明确的参数。

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

## 接口作为接收者
Go允许定义方法 —— 接受了特定类型的函数：

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

方法被添加了一个接受类型（上面的代码片段的接受类型是\*T）。
这个方法可以被类型为\*T或者T的调用
在这里顺便提一下，接口类型是不可以作为函数的接受者：

``` golang
type I interface {}
func (I) M() {}
```

这段代码将抛出一个编译期错误`invalid receiver type I (I is an interface type)`。
在第一部分和第二部分中有更多的方法进行介绍。

## 接口中“继承"
结构体的内嵌字段使其实现了接口的方法，于是这个这个结构体继承这个接口

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

在这个实例中，类型\*T2实现了接口I.
被\*T1实现的方法M作为了T2的一个内置的字段。
在过去的文章里有更多关于字段和方法的详细介绍。

## 接口类型作为map中的key或者value
map 是一个由key-value组成的数据结构。(在go1.8之前，map底层是通过哈希表实现的)

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

接口类型的值可以作为map中的key或者value来使用：

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

## 无所不在的接口
### error
go中内置的error是一个接口类型：

```go
type error interface {
	Error() string
}
```

任何类型实现了Error方法，此方法没有参数且返回一个string类型的值，那么这个类型就实现了error接口：

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

如果有任何异常发生，返回的error就将不会是nil。
error接口在前一节同样做了描述。
Writer接口在标准库中到处都有被用到，比如MultiWriter、TeeReader、net/http，还有很多其他用到的地方。

点赞以帮助别人发现这篇文章。如果你想得到新文章的更新，请关注我。

## 资源
- [GO程序语言规范——GO程序设计语言](https://golang.org/ref/spec#Method_expressions)
>Go is a general-purpose language designed with systems programming in mind. It is strongly typed and garbage-collected…
<br>*golang.org*

- [GO程序语言规范——GO程序设计语言](https://golang.org/ref/spec#Errors)
>Go is a general-purpose language designed with systems programming in mind. It is strongly typed and garbage-collected…
<br>*golang.org*

- [Go语言中提取字段和方法](https://medium.com/golangspec/promoted-fields-and-methods-in-go-4e8d7aefb3e3)
>Struct is a sequence of fields each with name and type. Usually it looks like:
<br>*medium.com*

- [Go方法（第一部分）](https://medium.com/golangspec/methods-in-go-part-i-a4e575dff860)
>Type defined in Golang program can have methods associated with it. Let’s see an example:
<br>*medium.com*

- [Go方法（第二部分）](https://medium.com/golangspec/methods-in-go-part-ii-2b4cc42c5cb6)
>This story explains remaining content from language specification touching methods. It’s strongly advised to read 1st…
<br>*medium.com*

*[保留部分版权](http://creativecommons.org/licenses/by/4.0/)*

*[Golang](https://medium.com/tag/golang?source=post)*
*[Programming](https://medium.com/tag/programming?source=post)*
*[Software Development](https://medium.com/tag/software-development?source=post)*
*[Education](https://medium.com/tag/education?source=post)*
*[Polymorphism](https://medium.com/tag/polymorphism?source=post)*

**喜欢读吗？给 Michał Łowicki 一些掌声吧。**

简单鼓励下还是大喝采，根据你对这篇文章的喜欢程度鼓掌吧。

---

via: https://medium.com/golangspec/interfaces-in-go-part-iii-61f5e7c52fb5

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[译者ID](https://github.com/xmge)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
