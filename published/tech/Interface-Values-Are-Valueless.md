首发于：https://studygolang.com/articles/12802

# 接口不是值类型（Interface Values Are Valueless）

## 介绍（Introduction）

最近，在 Slack 上我看过大量关于接口的问题。大多数时候，答案都很有技术性，并都关注了实现的细节。实现（细节）对于调试很有帮助，但实现对设计却毫无帮助。当用接口来设计代码时，行为才是主要需要关注的。

在这篇博文中，我希望提供一个不同的思考方式，关于接口，和用他们进行代码设计。我想让你停止关注于实现细节，而是关注于接口和具体的数据的关系。

## 面向数据设计（Data Oriented Design）

我相信写 Go 代码，应该用面向数据设计的方法，而不是面向对象。我的面向数据的第一条原则是：

如果你不了解你要处理的数据，你肯定不懂你要解决的问题。

所有你要解决的问题本质上就是数据转换的问题。有一些输入，然后你产生输出。这就是程序要做的事情。每一个你写的函数都是一个小的数据转换，（它们只是）为了帮助你解决大的数据转换。

因为你要解决的问题就是数据转换的问题，你写的算法要基于具体的数据。具体数据就是你存储在内存中的物理状态，通过网络发送，写入文件并进行基本操作。[机器情绪](https://mechanical-sympathy.blogspot.com/)取决于具体的数据和你允许你的机器做怎样的数据转换。

对面向数据的一个大的警告是关于如何处理修改。关于面向数据，我的第二条原则是：

当数据修改时，你的问题就修改了。当你的问题修改了，那么你的算法就要跟着修改。

一旦数据修改了，你的算法就需要修改。这是保证可读性和性能的最好方式。不幸的是，我们大多数人都被教导创建更多的抽象层，来处理变化。当设计需要修改时，我认为这种方式（创建更多的抽象层）将得不偿失。

你需要的是允许你的算法保持精简，来执行需要的数据转换。当数据修改时，你需要这样一种方式，算法改变了但却不会导致整个代码库的大部分代码发生级联变化。这就是使用接口的时候。当你关注接口时，你其实想要关注的是行为。

## 具体数据（Concrete Data）

因为每种事情都跟具体的数据有关，你应该从具体的数据开始。从具体的类型开始。

### 代码清单 1
```go
05 type file struct {
06     name string
07 }
```

在代码清单 1 中的第 5 行，关键字 `struct` 声明了一个名为 file 的类型。有了这个具体的类型声明，你可以创建一个这种类型的值。

### 代码清单 2
```go
13 func main() {
14     var f file
```

多亏了代码清单 2 中的第 14 行的声明，现在你有一个类型为 file，存在内存中，被命名为 f 的变量，并引用了具体的数据。这个数据被变量 f 索引，而且可以被操纵。

你可以再次使用关键字 struct 来定义第二块具体数据。

### 代码清单 3
```go
09 type pipe struct {
10     name string
11 }
```

在代码清单 3 中的第 09 行声明了类型为 `pipe`，并拥有一部分具体的数据。再一次，有了这个类型的声明，你可以在程序中，创建一个不同的值。

### 代码清单 4
```go
01 package main
02
03 import "fmt"
04
05 type file struct {
06     name string
07 }
08
09 type pipe struct {
10     name string
11 }
12
13 func main() {
14     var f file
15     var p pipe
16
17     fmt.Println(f, p)
18 }
```

现在，这个程序拥有两个清晰的具体数据定义，以及对应的一个值。在第 14 行，一个类型为 file 的值被创建，在第 15 行，一个类型为 pipe 的值被创建。为了程序完整，两个值在第 17 行都被 fmt 包展示出来。

## 接口值不是值类型（Interfaces Are Valueless）

你已经用关键字 `struct` 定义了你程序需要的值。还有另外一个关键字可以用来定义类型。那就是关键字 `interface`。

### 代码清单 5
```go
05 type reader interface {
06     read(b []byte) (int, error)
07 }
```

在代码清单 5 第 05 行，声明了一个 `interface` 的类型。`interface` 类型跟 `struct` 类型相对应。 `interface` 类型只能声明一组行为的方法。这意味着 `interface` 类型没有具体的值。

### 代码清单 6
```go
var r reader
```

有趣的是你可以声明一个 `interface` 类型的变量，就像代码清单 6 中展示的一样。这非常有趣，因为如果在 `interface` 中没有具体的值，那么变量 `r` 似乎就是毫无意义的。`interface` 类型定义以及创建的值是毫无价值的！

Boom！大脑爆炸了。

这是一个非常重要的概念。你必须明白：

- 变量 `r` 不代表任何东西。
- 变量 `r` 没有具体的值。
- 变量 `r` 毫无意义。

有一个实现细节使得 r 在后台是真实存在的，但从我们的编程模型来看，它却是不存在的。

当你认识到 `interface` 不是值类型，整个世界就变得清晰可以理解了。

### 代码清单 7
```go
37 func retrieve(r reader) error {
38     data := make([]byte, 100)
39
40     len, err := r.read(data)
41     if err != nil {
42         return err
43     }
44
45     fmt.Println(string(data[:len]))
46     return nil
47 }
```

在代码清单 7 定义了 `retrieve` 函数，一个我称之为多态的函数。在我继续前，先说明一下，多态的定义是按顺序的。看下来自于 Basic 的发明人 Tom Kurtz 的定义，这个定义会让你觉得多态函数是如此的特别。

“多态性意味着你写的一个特定的程序，它的行为会有所不同，而这取决于它所操作的数据。”

当我看到这个观点时，它总让我惊讶。它的简洁，却很好地说明了一点。多态性由具体的数据驱动。具有改变代码行为能力的是具体的数据。正如我以上所说的，你正在解决的问题是植根于具体的数据。面向数据设计是基于具体数据的。

如果你不懂你正在使用的【具体】数据，你就不懂你想要解决的问题。

Tom 的观点已经清楚地表明，具体的数据才是设计实现不同行为（多态性）抽象的关键。多么聪明的观点。

再回到代码清单 7。我将在下面重复一遍。

### 代码清单 7 - 复制

```go
37 func retrieve(r reader) error {
38     data := make([]byte, 100)
39
40     len, err := r.read(data)
41     if err != nil {
42         return err
43     }
44
45     fmt.Println(string(data[:len]))
46     return nil
47 }
```

当你读到第 37 行的 retrieve 函数声明时，函数似乎在说，传递给我一个类型为 reader 的值。但你知道这不可能，因为根本就没有一个值的类型为 reader。类型为 reader 的值压根不存在，因为 reader 是一个接口类型。我们都知道接口不是值类型。

那么函数到底想说什么？它想说的是：

我会接受任何实现了 reader 接口的具体数据（任何值或者指针）。但它必须实现 reader 接口定义的所有方法。

这就是你如何在 Go 中实现多态的方式。retrieve 函数不绑定到单个具体数据，而是绑定到任何实现 reader 接口的具体数据。

## 给数据赋予行为（Giving Data Behavior）

接下来的问题是，如何给数据赋予行为？这就是方法的用处。方法提供数据的行为机制。一旦数据有了行为方法，就可以实现多态。

”多态意味着你写的一个确定的程序，但他的行为可能不同，而这依赖于它所操作的数据。“

在 Go 中，你可以写函数和方法。选择方法而不是函数的一个原因是，数据被要求要实现给定接口的方法集。

### 代码清单 8

```go
05 type reader interface {
06     read(b []byte) (int, error)
07 }
08
09 type file struct {
10     name string
11 }
12
13 func (file) read(b []byte) (int, error) {
14     s := "<rss><channel><title>Going Go</title></channel></rss>"
15     copy(b, s)
16     return len(s), nil
17 }
18
19 type pipe struct {
20     name string
21 }
22
23 func (pipe) read(b []byte) (int, error) {
24     s := `{name: "bill", title: "developer"}`
25     copy(b, s)
26     return len(s), nil
27 }
```

请注意：你可能注意在接收者的方法中的第 13 行和第 23 行，声明了但没有给一个变量具体的名字。这其实是惯例，如果这个方法不需要使用接收者的任何数据时就可以不给接收者一个具体的名字。

在代码清单 8，在第 13 行，为类型 file 定义了一个方法，在第23 行，为 pipe 类型定义了一个方法。现在，每种类型都定义了一个名为 read 的方法，它已经实现了 reader 定义的所有方法。由于有了这些方法的定义，接下来我们可以说：

“类型 file 和 pipe 现在已经实现了 reader 接口。”

我在那段话中所说的每一句都很重要。如果你有看我之前关于值和指针语义的博客文章，那么你应该知道数据展现的行为由你正在使用的语义决定的。在这篇文章中我不会再讨论这些。这里有一个链接。

[https://www.ardanlabs.com/blog/2017/06/design-philosophy-on-data-and-semantics.html](https://www.ardanlabs.com/blog/2017/06/design-philosophy-on-data-and-semantics.html)

一旦这些值，值和指针，实现了这些方法，它们就可以传递给多态函数 retrieve。

### 代码清单 9

```go
package main

import "fmt"

type reader interface {
   read(b []byte) (int, error)
}

type file struct {
   name string
}

func (file) read(b []byte) (int, error) {
   s := "<rss><channel><title>Going Go</title></channel></rss>"
   copy(b, s)
   return len(s), nil
}

type pipe struct {
   name string
}

func (pipe) read(b []byte) (int, error) {
   s := `{name: "bill", title: "developer"}`
   copy(b, s)
   return len(s), nil
}

func main() {
   f := file{"data.json"}
   p := pipe{"cfg_service"}

   retrieve(f)
   retrieve(p)
}

func retrieve(r reader) error {
   data := make([]byte, 100)

   len, err := r.read(data)
   if err != nil {
	   return err
   }

   fmt.Println(string(data[:len]))
   return nil
}
```

代码清单 9 在 Go 中提供了一个完整的多态实例，并很好的说明了接口不是值类型这个观点。retrieve 函数可以接受任何实现了 reader 接口的数据，任何值或者指针。这正是你在第 33 行和第 34 行的函数调用中可以看到的情况。

现在，你可以看到 Go 中如何实现高级别的解耦，而且这种解耦还是非常地确切。你现在完全明白了数据的行为将传递为函数的行为。阅读代码时，这不再是陌生或无法理解的了。

当你接受接口不是值类型的时候，这一切就都可以说得通。这个函数不是要求 reader 值，因为 reader 值根本不存在。该函数要求的是实现 reader 定义的方法的具体数据。

## 接口值的分配（Interface Value Assignments）

接口不是值类型的观点可以延伸到接口值的分配。看下这些接口类型。

### 代码清单 10

```go
05 type Reader interface {
06     Read()
07 }
08
09 type Writer interface {
10     Write()
11 }
12
13 type ReadWriter interface {
14     Reader
15     Writer
16 }
```

有了这些接口声明，你可以实现一个实现了所有这三个接口的具体类型。

### 代码清单 11

```go
18 type system struct{
19     Host string
20 }
21
22 func (*system) Read()  { /* ... */ }
23 func (*system) Write() { /* ... */ }
```

下面，你可以再一次确认，接口为何不是值类型。

### 代码清单 12

```go
25 func main() {
26     var rw ReadWriter = &system{"127.0.0.1"}
27     var r Reader = rw
28     fmt.Println(rw, r)
29 }

// OUTPUT
&{127.0.0.1} &{127.0.0.1}
```

代码清单 12 的第 26 行，声明了一个类型为 ReadWriter，名字为 rw 的变量，并分配了一段具体的数据。具体数据是一个指向 system 的指针。然后在第 27 行中定义了类型为 Reader，名称为 r 的变量。有一个赋值操作跟这个声明相关。接口类型为 ReadWriter 的 rw 变量分配给了接口类型为 Reader 的新变量 r。

这应该会导致我们暂停一秒，因为变量 rw 和 r 的类型不同。我们知道在 Go 中两个不同名称的类型之间不会进行隐式地转换。但这还跟我们这种情况不一样。因为这些变量不是具体的值类型，它们是接口类型。

如果我们回到接口不是值类型的理解上，那么 rw 和 r 就都不是具体的值。因此，代码不能将接口值分配给对方。它唯一可以分配的是存储在接口值中的具体数据。幸亏有接口的类型声明，编译器可以验证一个接口内部的具体数据是否也满足另外的接口。

最后，我们只能处理具体的数据。处理接口值时，我们仍然只能处理存储在其中的具体数据。当你将接口值传递给 fmt 包进行显示时，请记住具体的数据就是显示的内容。再一次强调，他是唯一真实的东西。

## 结论（Conclusion）

我希望这篇文章能给你提供一种思考接口以及如何设计代码的不同方式的参考。我相信，一旦你摆脱了实现细节，并专注于接口与具体数据之间的关系，那么事情就会变得更加合理。面向数据的设计是编写更好的算法的方式，但要求关注对行为的解耦。接口允许通过调用具体数据的方法来达到行为的解耦。

---

via: https://www.ardanlabs.com/blog/2018/03/interface-values-are-valueless.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[gogeof](https://github.com/gogeof)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

