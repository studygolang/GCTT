## Go接口（第三部分）  
本文是golang接口相关主题的另一部分，讲解的内容包括接口类型中的方法，接口类型可以作为map的key或者




## 方法和接口
Go 是有方法的概念的，可以通过在调用结构体中的方法而获取一个函数，此函数和普通的函数不同的它具有此结构体的一些参数。
>
``` golang
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
golang的语言规范同样允许接口类型使用方法表达式
>
 从接口类型的函数中得到它的实现是不允许的。

让我们来看一个例子：
>
``` golang

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


## 接受接口类型
Go允许定义方法——一个接受了特定类型参数的函数
（源码）：
``` golang

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
方法被添加了一个接受类型（上面的代码片段的类型时*T）.这个方法可以被类型为*T或者T调用.而且接口类型不可以作为函数的接受者。
（源码):
``` golang

type I interface {}
func (I) M() {}

``` 
这段代码将抛出一个错误`invalid receiver type I (I is an interface type)`.在第一部分和第二部分中有更多的方法进行介绍。
## 接口中“继承"

## type 可做map中的key或者value

## 无所不在的类型

### error

### io.Writer


---

via: https://medium.com/golangspec/interfaces-in-go-part-iii-61f5e7c52fb5

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[译者ID](https://github.com/xmge)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

