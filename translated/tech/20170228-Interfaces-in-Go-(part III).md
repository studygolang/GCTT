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

## 接口接口类型

## 接口中“继承"

## type 可做map中的key或者value

## 无所不在的类型

### error

### io.Writer

