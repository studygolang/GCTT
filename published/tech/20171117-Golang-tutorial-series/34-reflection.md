已发布：https://studygolang.com/articles/13178

# 反射

![reflection](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-series/reflection-golang-3.png)

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 34 篇。

反射是 Go 语言的高级主题之一。我会尽可能让它变得简单易懂。

本教程分为如下小节。

- 什么是反射？
- 为何需要检查变量，确定变量的类型？
- reflect 包
  - reflect.Type 和 reflect.Value
  - reflect.Kind
  - NumField() 和 Field() 方法
  - Int() 和 String() 方法
- 完整的程序
- 我们应该使用反射吗？

让我们来逐个讨论这些章节。

## 什么是反射？

反射就是程序能够在运行时检查变量和值，求出它们的类型。你可能还不太懂，这没关系。在本教程结束后，你就会清楚地理解反射，所以跟着我们的教程学习吧。

## 为何需要检查变量，确定变量的类型？

在学习反射时，所有人首先面临的疑惑就是：如果程序中每个变量都是我们自己定义的，那么在编译时就可以知道变量类型了，为什么我们还需要在运行时检查变量，求出它的类型呢？没错，在大多数时候都是这样，但并非总是如此。

我来解释一下吧。下面我们编写一个简单的程序。

```go
package main

import (
    "fmt"
)

func main() {
    i := 10
    fmt.Printf("%d %T", i, i)
}
```

[在 playground 上运行](https://play.golang.org/p/1oZzPCCG2Qw)

在上面的程序中，`i` 的类型在编译时就知道了，然后我们在下一行打印出 `i`。这里没什么特别之处。

现在了解一下，需要在运行时求得变量类型的情况。假如我们要编写一个简单的函数，它接收结构体作为参数，并用它来创建一个 SQL 插入查询。

考虑下面的程序：

```go
package main

import (
    "fmt"
)

type order struct {
    ordId      int
    customerId int
}

func main() {
    o := order{
        ordId:      1234,
        customerId: 567,
    }
    fmt.Println(o)
}
```

[在 playground 上运行](https://play.golang.org/p/1oZzPCCG2Qw)

在上面的程序中，我们需要编写一个函数，接收结构体变量 `o` 作为参数，返回下面的 SQL 插入查询。

```
insert into order values(1234, 567)
```

这个函数写起来很简单。我们现在编写这个函数。

```go
package main

import (
    "fmt"
)

type order struct {
    ordId      int
    customerId int
}

func createQuery(o order) string {
    i := fmt.Sprintf("insert into order values(%d, %d)", o.ordId, o.customerId)
    return i
}

func main() {
    o := order{
        ordId:      1234,
        customerId: 567,
    }
    fmt.Println(createQuery(o))
}
```

[在 playground 上运行](https://play.golang.org/p/jhz4VHKIlQ5)

在第 12 行，`createQuery` 函数用 `o` 的两个字段（`ordId` 和 `customerId`），创建了插入查询。该程序会输出：

```bash
insert into order values(1234, 567)
```

现在我们来升级这个查询生成器。如果我们想让它变得通用，可以适用于任何结构体类型，该怎么办呢？我们用程序来理解一下。

```go
package main

type order struct {
    ordId      int
    customerId int
}

type employee struct {
    name string
    id int
    address string
    salary int
    country string
}

func createQuery(q interface{}) string {
}

func main() {

}
```

我们的目标就是完成 `createQuery` 函数（上述程序中的第 16 行），它可以接收任何结构体作为参数，根据结构体的字段创建插入查询。

例如，如果我们传入下面的结构体：

```go
o := order {
    ordId: 1234,
    customerId: 567
}
```

`createQuery` 函数应该返回：

```
insert into order values (1234, 567)
```

类似地，如果我们传入：

```go
 e := employee {
        name: "Naveen",
        id: 565,
        address: "Science Park Road, Singapore",
        salary: 90000,
        country: "Singapore",
    }
```

该函数会返回：

```
insert into employee values("Naveen", 565, "Science Park Road, Singapore", 90000, "Singapore")
```

由于 `createQuery` 函数应该适用于任何结构体，因此它接收 `interface{}` 作为参数。为了简单起见，我们只处理包含 `string` 和 `int` 类型字段的结构体，但可以扩展为包含任何类型的字段。

`createQuery` 函数应该适用于所有的结构体。因此，要编写这个函数，就必须在运行时检查传递过来的结构体参数的类型，找到结构体字段，接着创建查询。这时就需要用到反射了。在本教程的下一步，我们将会学习如何使用 `reflect` 包来实现它。

## reflect 包

在 Go 语言中，[`reflect`](https://golang.org/pkg/reflect/) 实现了运行时反射。`reflect` 包会帮助识别 [`interface{}`](https://studygolang.com/articles/12266) 变量的底层具体类型和具体值。这正是我们所需要的。`createQuery` 函数接收 `interface{}` 参数，根据它的具体类型和具体值，创建 SQL 查询。这正是 `reflect` 包能够帮助我们的地方。

在编写我们通用的查询生成器之前，我们首先需要了解 `reflect` 包中的几种类型和方法。让我们来逐个了解。

### reflect.Type 和 reflect.Value

`reflect.Type` 表示 `interface{}` 的具体类型，而 `reflect.Value` 表示它的具体值。`reflect.TypeOf()` 和 `reflect.ValueOf()` 两个函数可以分别返回 `reflect.Type` 和 `reflect.Value`。这两种类型是我们创建查询生成器的基础。我们现在用一个简单的例子来理解这两种类型。

```go
package main

import (
    "fmt"
    "reflect"
)

type order struct {
    ordId      int
    customerId int
}

func createQuery(q interface{}) {
    t := reflect.TypeOf(q)
    v := reflect.ValueOf(q)
    fmt.Println("Type ", t)
    fmt.Println("Value ", v)


}
func main() {
    o := order{
        ordId:      456,
        customerId: 56,
    }
    createQuery(o)

}
```

[在 playground 上运行](https://play.golang.org/p/81BS-bEfbCg)

在上面的程序中，第 13 行的 `createQuery` 函数接收 `interface{}` 作为参数。在第 14 行，[`reflect.TypeOf`](https://golang.org/pkg/reflect/#TypeOf) 接收了参数 `interface{}`，返回了[`reflect.Type`](https://golang.org/pkg/reflect/#Type)，它包含了传入的 `interface{}` 参数的具体类型。同样地，在第 15 行，[`reflect.ValueOf`](https://golang.org/pkg/reflect/#ValueOf) 函数接收参数 `interface{}`，并返回了 [`reflect.Value`](https://golang.org/pkg/reflect/#Value)，它包含了传来的 `interface{}` 的具体值。

上述程序会打印：

```
Type  main.order
Value  {456 56}
```

从输出我们可以看到，程序打印了接口的具体类型和具体值。

### relfect.Kind

`reflect` 包中还有一个重要的类型：[`Kind`](https://golang.org/pkg/reflect/#Kind)。

在反射包中，`Kind` 和 `Type` 的类型可能看起来很相似，但在下面程序中，可以很清楚地看出它们的不同之处。

```go
package main

import (
    "fmt"
    "reflect"
)

type order struct {
    ordId      int
    customerId int
}

func createQuery(q interface{}) {
    t := reflect.TypeOf(q)
    k := t.Kind()
    fmt.Println("Type ", t)
    fmt.Println("Kind ", k)


}
func main() {
    o := order{
        ordId:      456,
        customerId: 56,
    }
    createQuery(o)

}
```

[在 playground 上运行](https://play.golang.org/p/Xw3JIzCm54T)

上述程序会输出：

```
Type  main.order
Kind  struct
```

我想你应该很清楚两者的区别了。`Type` 表示 `interface{}` 的实际类型（在这里是 **`main.Order`**)，而 `Kind` 表示该类型的特定类别（在这里是 **`struct`**）。

### NumField() 和 Field() 方法

[`NumField()`](https://golang.org/pkg/reflect/#Value.NumField) 方法返回结构体中字段的数量，而 [`Field(i int)`](https://golang.org/pkg/reflect/#Value.Field) 方法返回字段 `i` 的 `reflect.Value`。

```go
package main

import (
    "fmt"
    "reflect"
)

type order struct {
    ordId      int
    customerId int
}

func createQuery(q interface{}) {
    if reflect.ValueOf(q).Kind() == reflect.Struct {
        v := reflect.ValueOf(q)
        fmt.Println("Number of fields", v.NumField())
        for i := 0; i < v.NumField(); i++ {
            fmt.Printf("Field:%d type:%T value:%v\n", i, v.Field(i), v.Field(i))
        }
    }

}
func main() {
    o := order{
        ordId:      456,
        customerId: 56,
    }
    createQuery(o)
}
```

[在 playground 上运行](https://play.golang.org/p/FBHfJfuTaEe)

在上面的程序中，因为 `NumField` 方法只能在结构体上使用，我们在第 14 行首先检查了 `q` 的类别是 `struct`。程序的其他代码很容易看懂，不作解释。该程序会输出：

```
Number of fields 2
Field:0 type:reflect.Value value:456
Field:1 type:reflect.Value value:56
```

### Int() 和 String() 方法

[`Int`](https://golang.org/pkg/reflect/#Value.Int) 和 [`String`](https://golang.org/pkg/reflect/#Value.String) 可以帮助我们分别取出 `reflect.Value` 作为 `int64` 和 `string`。

```go
package main

import (
    "fmt"
    "reflect"
)

func main() {
    a := 56
    x := reflect.ValueOf(a).Int()
    fmt.Printf("type:%T value:%v\n", x, x)
    b := "Naveen"
    y := reflect.ValueOf(b).String()
    fmt.Printf("type:%T value:%v\n", y, y)

}
```

[在 playground 上运行](https://play.golang.org/p/UIllrLVoGwI)

在上面程序中的第 10 行，我们取出 `reflect.Value`，并转换为 `int64`，而在第 13 行，我们取出 `reflect.Value` 并将其转换为 `string`。该程序会输出：

```
type:int64 value:56
type:string value:Naveen
```

## 完整的程序

现在我们已经具备足够多的知识，来完成我们的查询生成器了，我们来实现它把。

```go
package main

import (
    "fmt"
    "reflect"
)

type order struct {
    ordId      int
    customerId int
}

type employee struct {
    name    string
    id      int
    address string
    salary  int
    country string
}

func createQuery(q interface{}) {
    if reflect.ValueOf(q).Kind() == reflect.Struct {
        t := reflect.TypeOf(q).Name()
        query := fmt.Sprintf("insert into %s values(", t)
        v := reflect.ValueOf(q)
        for i := 0; i < v.NumField(); i++ {
            switch v.Field(i).Kind() {
            case reflect.Int:
                if i == 0 {
                    query = fmt.Sprintf("%s%d", query, v.Field(i).Int())
                } else {
                    query = fmt.Sprintf("%s, %d", query, v.Field(i).Int())
                }
            case reflect.String:
                if i == 0 {
                    query = fmt.Sprintf("%s\"%s\"", query, v.Field(i).String())
                } else {
                    query = fmt.Sprintf("%s, \"%s\"", query, v.Field(i).String())
                }
            default:
                fmt.Println("Unsupported type")
                return
            }
        }
        query = fmt.Sprintf("%s)", query)
        fmt.Println(query)
        return

    }
    fmt.Println("unsupported type")
}

func main() {
    o := order{
        ordId:      456,
        customerId: 56,
    }
    createQuery(o)

    e := employee{
        name:    "Naveen",
        id:      565,
        address: "Coimbatore",
        salary:  90000,
        country: "India",
    }
    createQuery(e)
    i := 90
    createQuery(i)

}
```

[在 playground 上运行](https://play.golang.org/p/82Bi4RU5c7W)

在第 22 行，我们首先检查了传来的参数是否是一个结构体。在第 23 行，我们使用了 `Name()` 方法，从该结构体的 `reflect.Type` 获取了结构体的名字。接下来一行，我们用 `t` 来创建查询。

在第 28 行，[case 语句](https://studygolang.com/articles/11957) 检查了当前字段是否为 `reflect.Int`，如果是的话，我们会取到该字段的值，并使用 `Int()` 方法转换为 `int64`。[if else 语句](https://studygolang.com/articles/11902)用于处理边界情况。请添加日志来理解为什么需要它。在第 34 行，我们用来相同的逻辑来取到 `string`。

我们还作了额外的检查，以防止 `createQuery` 函数传入不支持的类型时，程序发生崩溃。程序的其他代码是自解释性的。我建议你在合适的地方添加日志，检查输出，来更好地理解这个程序。

该程序会输出：

```
insert into order values(456, 56)
insert into employee values("Naveen", 565, "Coimbatore", 90000, "India")
unsupported type
```

至于向输出的查询中添加字段名，我们把它留给读者作为练习。请尝试着修改程序，打印出以下格式的查询。

```
insert into order(ordId, customerId) values(456, 56)
```

## 我们应该使用反射吗？

我们已经展示了反射的实际应用，现在考虑一个很现实的问题。我们应该使用反射吗？我想引用 [`Rob Pike`](https://en.wikipedia.org/wiki/Rob_Pike) 关于使用反射的格言，来回答这个问题。

> 清晰优于聪明。而反射并不是一目了然的。

反射是 Go 语言中非常强大和高级的概念，我们应该小心谨慎地使用它。使用反射编写清晰和可维护的代码是十分困难的。你应该尽可能避免使用它，只在必须用到它时，才使用反射。

本教程到此结束。希望你们喜欢。祝你愉快。

**上一教程 - [函数是一等公民](https://studygolang.com/articles/12789)**

---

via: https://golangbot.com/reflection/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
