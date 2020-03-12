首发于：https://studygolang.com/articles/27152

# 在 Go 语言中，我为什么使用接口

强调一下是**我个人**的见解以及接口在 **Go 语言**中的意义。

如果您写代码已经有了一段时间，我可能不需要过多解释接口所带来的好处，但是在深入探讨 Go 语言中的接口前，我想花一两分钟先来简单介绍一下接口。
如果您对接口很熟悉，请先跳过下面这段。

## 接口的简单介绍

在任一编程语言中，接口——方法或行为的集合，在功能和该功能的使用者之间构建了一层薄薄的抽象层。在使用接口时，并不需要了解底层函数是如何实现的，因为接口隔离了各个部分（划重点）。

跟不使用接口相比，使用接口的最大好处就是可以使代码变得简洁。例如，您可以创建多个组件，通过接口让它们以统一的方式交互，尽管这些组件的底层实现差异很大。这样就可以在编译甚至运行的时候动态替换这些组件。

用 Go 的 `io.Reader` 接口举个例子。`io.Reader` 接口的所有实现都有 `Read(p []byte) (n int, err error)` 函数。使用 `io.Reader` 接口的使用者不需要知道使用这个 `Read` 函数的时候那些字节从何而来。

## 具体到 Go 语言

在我使用 Go 语言的过程中，与我使用过的其他任何编程语言相比，我经常发现其他的、不那么明显的使用接口的原因。今天，我将介绍一个很普遍的，也是我遇到了很多次的使用接口的原因。

## Go 语言没有构造函数

很多编程语言都有构造函数。构造函数是定义自定义类型（即 OO 语言中的类）时使用的一种建立对象的方法，它可以确保必须执行的任何初始化逻辑均已执行。

例如，假设所有 `widgets` 都必须有一个不变的，系统分配的标识符。在 Java 中，这很容易实现：

```java
package io.krancour.widget;

import java.util.UUID;

public class Widget {

    private String id;

    // 使用构造函数初始化
    public Widget() {
        id = UUID.randomUUID().toString();
    }

    public String getId() {
        return id;
    }
}
```

```java
class App {
    public static void main( String[] args ){
        Widget w = new Widget();
        System.out.println(w.getId());
    }
}
```

从上面这个例子可以看到，没有执行初始化逻辑就无法实例化一个新的 `Widget` 。

但是 Go 语言没有此功能。 :(

在 Go 语言中，可以直接实例化一个自定义类型。

定义一个 `Widget` 类型：

```go
package widgets

type Widget struct {
    id string
}

func (w Widget) ID() string {
    return w.id
}
```

可以像这样实例化和使用一个 `widget`：

```go
package main

import (
    "fmt"
    "github.com/krancour/widgets"
)

func main() {
    w := widgets.Widget{}
    fmt.Println(w.ID())
}
```

如果运行此示例，那么（也许）意料之中的结果是，打印出的 ID 是空字符串，因为它从未被初始化，而空字符串是字符串的“零值”。
我们可以在 `widgets` 包中添加一个类似于构造函数的函数来处理初始化：

```go
package widgets

import uuid "github.com/satori/go.uuid"

type Widget struct {
    id string
}

func NewWidget() Widget {
    return Widget{
        id: uuid.NewV4().String(),
    }
}

func (w Widget) ID() string {
    return w.id
}
```

然后我们简单地修改 `main` 来使用这个类似于构造函数的新函数：

```go
package main

import (
    "fmt"
    "github.com/krancour/widgets"
)

func main() {
    w := widgets.NewWidget()
    fmt.Println(w.ID())
}
```

执行该程序，我们得到了想要的结果。

但是仍然存在一个严重问题！我们的 `widgets` 包没有强制用户在初始一个 `widget` 的时候使用我们的构造函数。

## 变量私有化

首先我们尝试把自定义类型的变量私有化，以此来强制用户使用我们规定的构造函数来初始化 `widget`。在 Go 语言中，类型名、函数名的首字母是否大写决定它们是否可被其他包访问。名称首字母大写的可被访问（也就是 `public` ），而名称首字母小写的不可被访问（也就是 `private` ）。所以我们把类型 `Widget` 改为类型 `widget` ：

```go
package widgets

import uuid "github.com/satori/go.uuid"

type widget struct {
    id string
}

func NewWidget() widget {
    return widget{
        id: uuid.NewV4().String(),
    }
}

func (w widget) ID() string {
    return w.id
}
```

我们的 `main` 代码保持不变，这次我们得到了一个 ID 。这比我们想要的要近了一步，但是我们在此过程中犯了一个不太明显的错误。类似于构造函数的 `NewWidget` 函数返回了一个私有的实例。尽管编译器对此不会报错，但这是一种不好的做法，下面是原因解释。

在 Go 语言中，***包***是复用的基本单位。其他语言中的***类***是复用的基本单位。如前所述，任何无法被外部访问的内容实质上都是“包私有”，是该包的内部实现细节，对于使用这个包的使用者来说不重要。因此，Go 的文档生成工具 `godoc` 不会为私有的函数、类型等生成文档。

当一个公开的构造函数返回一个私有的 `widget` 实例，实际上就陷入了一条死胡同。调用这个函数的人哪怕有这个实例，也绝对在文档里找不到任何关于这个实例类型的描述，也更不知道 `ID()` 这个函数。Go 社区非常重视文档，所以这样做是不会被接受的。

## 轮到接口上场了

回顾一下，到目前为止，我们写了一个类似于构造函数的函数来解决 Go 语言缺乏构造函数的问题，但是为了确保人们用该函数而不是直接实例化 `Widget` ，我们更改了该类型的可见性——将其重命名为 `widget`，即私有化了。虽然编译器不会报错，但是文档中不会出现对这个私有类型的描述。不过，我们距离想要的目标还近了一步。接下来就要使用接口来完成后续的了。

通过创建一个***可被访问的***、`widget` 类型可以实现的接口，我们的构造函数可以返回一个公开的类型实例，并且会显示在 `godoc` 文档中。同时，这个接口的底层实现依然是私有的，使用者无法直接创建一个实例。

```go
package widgets

import uuid "github.com/satori/go.uuid"

// Widget is a ...
type Widget interface {
    // ID 返回这个 widget 的唯一标识符
    ID() string
}

type widget struct {
    id string
}

// NewWidget() 返回一个新的 Widget 实例
func NewWidget() Widget {
    return widget{
        id: uuid.NewV4().String(),
    }
}

func (w widget) ID() string {
    return w.id
}
```

## 总结

我希望我已经充分地阐述了 Go 语言的这一特质——构造函数的缺失反而促进了接口的使用。

在我的下一篇文章中，我将介绍一种几乎与之相反的场景——在其他语言中要使用接口但是在 Go 语言中却不必。

---

via: https://medium.com/@kent.rancourt/go-pointers-why-i-use-interfaces-in-go-338ae0bdc9e4

作者：[Kent Rancourt](https://medium.com/@kent.rancourt)
译者：[zhiyu-tracy-yang](https://github.com/zhiyu-tracy-yang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
