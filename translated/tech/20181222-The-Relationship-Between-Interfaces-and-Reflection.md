# 接口和反射的关系

接口是Go中用于抽象的基本工具之一。接口在值进行分配的时候存储类型信息。反射则是在运行时检查类型和值的方法。

Go通过`reflect`包实现了反射。该包提供了一些类型和方法用于检查接口结构部分，不仅如此，它还可以在运行时进行值的修改。

在这篇文章中，我希望能说明接口结构的各部分和反射API之间的关系，并最终使得反射包变得更加容易理解。

## 向一个接口分配一个值

一个接口编码了三件事： 值，方法集，以及所存储的值的类型。

下图展示了一个接口的内部结构。

![pic1](https://blog.gopheracademy.com/postimages/advent-2018/interfaces-and-reflect/interface.svg)

我们可以在该图中很清楚地看到接口内部结构中的三个部分：`_type`表示类型信息，`*data`是一个指向实际值的指针，`itab`则编码了方法集。

当一个方法接收一个接口作为参数时，将一个值传递给该函数则会将该值，该值的方法集，和类型打包到接口中。

## 通过反射包在运行时检查接口数据

一旦一个值存储进接口，你就可以使用`reflect`包来检查该接口的各个部分。我们不能直接检查该接口的结构；而是通过反射包维护着我们自己有权访问的接口结构的副本。

即使我们通过接口对象访问接口，但和直接访问相关的底层的接口对象有相同效果。

类型`reflect.Type`和`reflect.Value`提供了可以访问接口结构部分的方法。

`reflect.Type`侧重于公开类型相关的数据，因此它只限于结构的`_type`部分，而`reflect.Value`则必须将类型信息与值结合起来，以允许程序员检查和操作值，因此必须要查看`_type`以及`*data`部分。

### `reflect.Type` - 检查类型

`reflect.TypeOf()`函数用于从一个值中提取该值的类型信息。因为该函数唯一的参数是一个空接口`interface{}`，传递给它的值会被分配到该空接口上，因此该值的类型，方法集合值我们都可以轻松地获得。

`reflect.TypeOf()`返回一个`reflect.Type`类型的值，它提供了一些方法以允许你可以检查传入的值的类型信息。

下面是一些可用的`Type`方法以及它们返回的与接口相对应的位。

![pic2](https://blog.gopheracademy.com/postimages/advent-2018/interfaces-and-reflect/reflect-type.svg)

___`relfect.Type`使用示例___

```go
package main

import (
        "log"
        "reflect"
)

type Gift struct {
        Sender    string
        Recipient string
        Number    uint
        Contents  string
}

func main() {
        g := Gift{
                Sender:    "Hank",
                Recipient: "Sue",
                Number:    1,
                Contents:  "Scarf",
        }

        t := reflect.TypeOf(g)

        if kind := t.Kind(); kind != reflect.Struct {
                log.Fatalf("This program expects to work on a struct; we got a %v instead.", kind)
        }

        for i := 0; i < t.NumField(); i++ {
                f := t.Field(i)
                log.Printf("Field %03d: %-10.10s %v", i, f.Name, f.Type.Kind())
        }
}
```

此程序的目的是打印出`Gift`结构体的所有字段。当我们把`g`传递给`reflect.TypeOf()`时，实际上`g`被分配给了被编译器使用相应的类型和方法集填充的接口。这就允许我们可以遍历该接口类型部分中的`[]fileds`，然后我们得到如下输出：

```bash
2018/12/16 12:00:00 Field 000: Sender     string
2018/12/16 12:00:00 Field 001: Recipient  string
2018/12/16 12:00:00 Field 002: Number     uint
2018/12/16 12:00:00 Field 003: Contents   string
```

### `reflect.Method` - 检查`itab`和方法集

`reflect.Type`类型也会允许你访问`itab`部分，来从接口中提取出方法信息。

![pic3](https://blog.gopheracademy.com/postimages/advent-2018/interfaces-and-reflect/reflect-method.svg)

___通过反射检查方法___

```go
package main

import (
        "log"
        "reflect"
)

type Reindeer string

func (r Reindeer) TakeOff() {
        log.Printf("%q lifts off.", r)
}

func (r Reindeer) Land() {
        log.Printf("%q gently lands.", r)
}

func (r Reindeer) ToggleNose() {
        if r != "rudolph" {
                panic("invalid reindeer operation")
        }
        log.Printf("%q nose changes state.", r)
}

func main() {
        r := Reindeer("rudolph")

        t := reflect.TypeOf(r)

        for i := 0; i < t.NumMethod(); i++ {
                m := t.Method(i)
                log.Printf("%s", m.Name)
        }
}
```

这段代码完全迭代了存储在`itab`的函数数据，并显示了每个方法的名称：

```bash
2018/12/16 12:00:00 Land
2018/12/16 12:00:00 TakeOff
2018/12/16 12:00:00 ToggleNose
```

## `reflect.Value` - 检查值

目前为止我们仅仅讨论了类型信息 - 字段，方法等。`reflect.Value`则向我们展示了存储在接口中的实际值的信息。

与`reflect.Value`相关的方法必然会将类型信息和实际的值组合在一起。例如，为了从一个结构中提取字段信息，`refelct`包必须将结构的布局知识 - 特别是关于存储在`_type`中的字段和字段偏移量的信息 - 与接口的`*data`部分所指向的实际值结合起来，以便正确地解码结构。

![pic4](https://blog.gopheracademy.com/postimages/advent-2018/interfaces-and-reflect/reflect-value.svg)

___观察修改值的示例___

```go
package main

import (
        "log"
        "reflect"
)

type Child struct {
        Name  string
        Grade int
        Nice  bool
}

type Adult struct {
        Name       string
        Occupation string
        Nice       bool
}

// search a slice of structs for Name field that is "Hank" and set its Nice
// field to true.
func nice(i interface{}) {
        // retrieve the underlying value of i.  we know that i is an
        // interface.
        v := reflect.ValueOf(i)

        // we're only interested in slices to let's check what kind of value v is. if
        // it isn't a slice, return immediately.
        if v.Kind() != reflect.Slice {
                return
        }

        // v is a slice.  now let's ensure that it is a slice of structs.  if not,
        // return immediately.
        if e := v.Type().Elem(); e.Kind() != reflect.Struct {
                return
        }

        // determine if our struct has a Name field of type string and a Nice field
        // of type bool
        st := v.Type().Elem()

        if nameField, found := st.FieldByName("Name"); found == false || nameField.Type.Kind() != reflect.String {
                return
        }

        if niceField, found := st.FieldByName("Nice"); found == false || niceField.Type.Kind() != reflect.Bool {
                return
        }

        // Set any Nice fields to true where the Name is "Hank"
        for i := 0; i < v.Len(); i++ {
                e := v.Index(i)
                name := e.FieldByName("Name")
                nice := e.FieldByName("Nice")

                if name.String() == "Hank" {
                        nice.SetBool(true)
                }
        }
}

func main() {
        children := []Child{
                {Name: "Sue", Grade: 1, Nice: true},
                {Name: "Ava", Grade: 3, Nice: true},
                {Name: "Hank", Grade: 6, Nice: false},
                {Name: "Nancy", Grade: 5, Nice: true},
        }

        adults := []Adult{
                {Name: "Bob", Occupation: "Carpenter", Nice: true},
                {Name: "Steve", Occupation: "Clerk", Nice: true},
                {Name: "Nikki", Occupation: "Rad Tech", Nice: false},
                {Name: "Hank", Occupation: "Go Programmer", Nice: false},
        }

        log.Printf("adults before nice: %v", adults)
        nice(adults)
        log.Printf("adults after nice: %v", adults)

        log.Printf("children before nice: %v", children)
        nice(children)
        log.Printf("children after nice: %v", children)
}
```

```bash
2018/12/16 12:00:00 adults before nice: [{Bob Carpenter true} {Steve Clerk true} {Nikki Rad Tech false} {Hank Go Programmer false}]
2018/12/16 12:00:00 adults after nice: [{Bob Carpenter true} {Steve Clerk true} {Nikki Rad Tech false} {Hank Go Programmer true}]
2018/12/16 12:00:00 children before nice: [{Sue 1 true} {Ava 3 true} {Hank 6 false} {Nancy 5 true}]
2018/12/16 12:00:00 children after nice: [{Sue 1 true} {Ava 3 true} {Hank 6 true} {Nancy 5 true}]
```

在最后一个例子中，我们将我们学到的东西结合起来，通过`reflect.Value`来修改一个值。在这个用例中，有人编写了一个名为`nice()`（可能是Hank)的函数，该函数将切片中任何一个结构中名为`Hank`的`nice`字段都设置为`true`。

需要注意的是，`nice()`能够修改所有传递给它的切片，而且它实际的接收类型并不重要 - 只要它是一个元素为结构体的切片，并且具有`Name`和`Nice`字段即可。

## 结论

Go中的反射是使用接口以及`reflect`包实现的。它并没有什么神奇之处 - 当你在使用反射时，你可以访问接口的各个部分以及存储在其中的值。

通过这种方式，接口的行为就像一面镜子，允许程序检查自己。

虽然Go是一门静态类型语言，但通过反射和接口的结合，使得它拥有强大的能力，这通常只存在于动态类型语言当中。

有关Go中反射的更多信息，请务必阅读`reflect`包文档以及其他和反射相关的精彩博客文章。

---

via: [Interfaces-and-Reflections](https://blog.gopheracademy.com/advent-2018/interfaces-and-reflect/).

作者：[Ayan George](https://blog.gopheracademy.com/advent-2018/interfaces-and-reflect/)
译者：[barryz](https://github.com/barryz)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
