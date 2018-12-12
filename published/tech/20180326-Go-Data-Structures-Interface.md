首发于：https://studygolang.com/articles/15055

# Go 中的数据结构 -- Interface

Go 中的 interface 可以静态编译，动态执行，是最让我感到兴奋的一个特性。如果要让我推荐一个 Go 语言的特性给其他的语言，那我一定会推荐 interface。

本文是我对于 Go 语言中 interface 类型在 gc 编译器上实现的一些想法。Ian Lance Taylor 写了两篇关于 interface 类型在 gccgo 中实现的[文章](https://www.airs.com/blog/archives/277)。本文与之最大的不同是本文有一些图片可以更形象的说明原理。

在研究具体的实现原理之前，我们一起来看看 interface 需要支持什么功能。

## 用法

Go 的 interface 让你可以像纯动态语言一样使用[鸭子类型](https://en.wikipedia.org/wiki/Duck_typing)，同时编译器也可以捕获一些明显的参数类型错误(比如传给一个希望使用 Read 类型的函数一个 int 类型的参数)。

在使用一个 interface 之前, 我们首先要定义 interface 类型的方法集合（比如下面的 ReadCloser 类型):

```go
type ReadCloser interface {
    Read(b []byte) (n int, err os.Error)
    Close()
}
```

然后，我们要定义一个使用 ReadCloser 的函数。比如下面的这个函数会不断调用 ReadCloser 的 Read 来获取所有的数据，然后再调用 Close 。

```go
func ReadAndClose(r ReadCloser, buf []byte) (n int, err os.Error) {
    for len(buf) > 0 && err == nil {
        var nr int
        nr, err = r.Read(buf)
        n += nr
        buf = buf[nr:]
    }
    r.Close()
    return
}
```

调用 ReadAndClose 的代码可以给第一个参数传入一个任意类型的值，只要这个值具有 Read 和 Close 方法。另外，当传入一个错误类型的参数时，在编译阶段就可以发现这个错误，而不是像 Python 一样只能在运行阶段发现。

不过，接口并不局限于静态检查。您可以动态地检查特定的接口值是否有附加的方法。例如:

```go
type Stringer interface {
    String() string
}

func ToString(any interface{}) string {
    if v, ok := any.(Stringer); ok {
        return v.String()
    }
    switch v := any.(type) {
    case int:
        return strconv.Itoa(v)
    case float:
        return strconv.Ftoa(v, 'g', -1)
    }
    return "???"
}
```

any 的类型是接口类型 interface{}，这意味着 any 可以有任何的方法，它可以包含任何类型。if 语句中的 ok 查看 any 变量是否可以转化为 Stringer 类型 (包含一个 String 方法)。如果可以，函数会返回一个字符串，否则，就会尝试一些其他的类型。这基本上就是 fmt 包中的一些逻辑。

举个例子，考虑一个 64-bit 的整数，这个整数有一个 String 方法和一个 Get 方法：

```go
type Binary uint64

func (i Binary) String() string {
    return strconv.Uitob64(i.Get(), 2)
}

func (i Binary) Get() uint64 {
    return uint64(i)
}
```

一个 Binary 的值可以传给 ToString 函数，然后可以使用 String 方法格式化，尽管程序从未说过 Binary 打算实现了 Stringer 类。其实也不需要完全实现 Stringer 类，运行时可以看到 Binary 有一个 String 方法，所以就人为它实现了 Stringer，即使 Binary 的作者从来没有听过 Stringer。

这些例子表明，即使在编译时检查所有隐式转换，显式的 interface-to-interface 的转换也可以在运行时通过查询方法集实现。[《Effective Go》](http://golang.org/doc/effective_go.htm) 中有更多关于如何使用接口的详细信息和示例。

## Interface的值

带有方法的语言通常会有两种选择：准备一个所有方法的静态调用表(C++ 和 java)，或者在每次调用时进行方法查找(Smalltalk，python 以及 javascript)。

Go 选择了一种混合的方式：它的确有静态调用表，不过是在运行的时候生成的。我不知道 Go 是否是第一个采用这个技术的语言，但是这种技术的确是不常见的。

举个例子，类型 Binary 的值是一个由两个 32-bit 的字组成的 64-bit 整数。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/Go-Data-Structures-Interfaces/gointer1.png)

对于任何一个 interface, 它的值也是由两个 32-bit 的字组成，第一个字它提供了一个指向 interface 中数据类型的信息的指针，第二个字是一个指向相关数据的指针。`s := Stringer(b)` 将 b 分配给 Stringer 类型 s，会设置 interface 的这两个字的指针。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/Go-Data-Structures-Interfaces/gointer2.png)

interface 值中的第一个指针指向了一个表（itable 或者 itab）。Itable 由两部分组成，第一部分是指向原始数据的数据类型，第二部分是一个函数指针的列表。需要注意的是，与 interface 类型相同，itable 也不是一个动态类型的数据。在我们这个例子中，Stringer 类型的 itable 中列出了 Binary 类型中满足 Stringer 接口的所有函数指针列表( 其实只有 String, Binary 的另一个方法 Get 不会出现在 itable 里面)。

interface 值中的第二个指针指向了实际的数据，也就是 b 的一份副本。需要注意的是，声明 `s := Stringer(b)` 会给 b 创造一份副本，而不是直接指向 b。如果后来 b 的值改变了，那么 s 还会是 b 之前的值。存储在 interface 中的值可能是任意大的，但是因为只有一个字专门用于在接口结构中保存值，所以会在堆上分配一大块内存，并在那一个字的位置中保存指针。(当然，如果这个值的长度少于一个字，则不需要再在堆上分配内存，具体的优化方法会在后面描述)

要检查接口值是否是特定类型，Go 编译器生成了表达式 `s.tab->type` 来获取类型指针并检查是否是所需的类型。如果类型匹配，则可以通过取消引用来复制 `s.data`。

在调用 s.String() 时，Go 编译器生成了一份代码，相当于 C 语言中的 `s.tab->fun[0](s.data)`：它会从 itable 中选择合适的函数指针，将 interface 值的数据字段作为它的第一个参数。如果你运行 `8g -S x.go` ( 译者注：8g 是老版本中的一个工具，在 go 1.5 后可以使用 `go tool compile -S x.go` 来代替)，你可以看到这段代码。需要注意的是，itable 中的函数的参数只能传入 32-bit 数据字段指针，而不能传入 64-bit 的值。一般来说，在调用接口的时候，代码是不会知道这个指针的意义，也不知道它所指向的数据有多少。相反，在接口的 itable 中的函数，也都期望接收到一个 32-bit 的指针。因此在这个实例中，函数的指针应该是 `(*Binary).String` 而不是 `Binary.String`。

这个例子是一个只有一个方法的 interface。一个具有更多方法的 interface 将在 itable 底部有更多条记录。

## 计算 Itable

现在，我们已经描述了 itable 的结构，可是它是怎么生成的呢？

Go 的动态类型转换使编译器不可能对所有的 interface 到具体类型的 itables 进行预先计算，但是其实大部分的 itable 也是不需要的(比如程序中只需要计算 Stringer-Binary 的 itable，但是不需要计算 Stringer-string, Stringer-uint64 等对应的 itable)。

因此，在 go 语言中，编译器为每个具体类型生成一个类型描述，包含了由该类型实现的方法的列表。类似地，编译器也为每个接口类型生成类型描述，同样也包含了该接口类型的实现方法列表。在运行时, 编译器在具体类型的方法表中查找 interface 类型的方法表中列出的每个方法来计算 itable。当生成了 itable 后，会将其保存在cache中，所以每个 itable 只需要生成一次。

在本文的例子中， Stringer 的方法表中只有一个方法，而 Binary 的方法表中有两个方法。假设 interface 类型 ni个方法，具体类型有 nt 个方法，那么通常来说，检索的复杂度为 `O(ni * nt)`。但是 go 采用了一种更好的方法，通过对两个表的函数进行排序，并且对其进行同步遍历，检索的复杂度可以降为 `O(ni + nt)`。

## 内存优化

上述的实现方法所占用的空间可以使用两种互补的方法来优化。

第一，当 interface 类型没有定义任何方法时，itable 除了指向原来的类型外，没有任何用途。在这种情况下，可以不再使用 itable，其指针直接指向原来的类型。

一个 interface 是否定义任何方法，这是一个静态的属性，因此编译器知道程序用的是哪种指针。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/Go-Data-Structures-Interfaces/gointer3.png)

第二，如果与 interface 值相关联的值可以单个字标识，那么就不需要分配堆空间并使用指针。如果我们使用 Binary32 而不是 Binary, 那么它的数据就可以直接存储在 interface 的值中，而不需要再分配堆空间了。interface 的值是堆的指针还是实际的值完全取决于值类型的大小。

编译器会管理 itable 中的函数，如果传入的参数在一个字之内，那么就直接使用这个字，否则，就通过间接引用获取传入的值。在上面的例子中， itable 中的方法是 `(*Binary).String`。但是在 Binary32 的例子中，itable 中的方法就是 `(*Binary).String`。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/Go-Data-Structures-Interfaces/gointer4.png)

当然，如果一个空的 interface 并且传入了一个字的数据，可以使用上面两种方法同时进行优化。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/Go-Data-Structures-Interfaces/gointer5.png)

## 方法查询性能

Smalltalk 和许多其他的动态语言在每次调用方法时都会执行方法查找。为了提高速度，大部分都是在指令流中简单的加入了单条缓存。对于多线程的语言，这些缓存必须小心的存储，因为可能存在多个线程同时访问一个函数的情况。

因为 Go 具有静态类型提示和动态方法查找的方法，所以它可以将查找从调用的位置移回到值存储在接口中的位置。

```go
var any interface{}  // initialized elsewhere
s := any.(Stringer)  // dynamic conversion
for i := 0; i < 100; i++ {
    fmt.Println(s.String())
}
```

在第 2 行的赋值的时候，程序会计算 itable；因此，在第 4 行执行的 `s.String()` 只需要执行几次内存查找和一次间接调用即可。

与此相反，在像Smalltalk(或JavaScript、Python)这样的动态语言中，每次执行到第4行时程序都会进行方法查找，在一次次的循环中重复不必要的工作。前面提到的缓存可能会让起稍微快一些，但是它仍然不如一个间接调用指令。

当然，这是一篇博客文章，我没有任何数字来支持这个讨论，但是像 Go 语言这样减少内存竞争可以很好的提高性能。另外，本文主要是介绍体系结构，而不是实现的细节，在实现的过程中可能会使用一些常量的优化。

---

via: https://research.swtch.com/interfaces

作者：[Russ Cox](https://swtch.com/~rsc/)
译者：[bizky](https://github.com/bizky)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
