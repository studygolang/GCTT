已发布：https://studygolang.com/articles/12781

# 值接收器构造器的探讨

在 Go 中，当涉及到面向对象编程，会有许多的前期工作需要做，以至于许多刚从其它语言迁移到 Go 的程序员会将那些语言中的一些概念带到 Go 中。对象构造器就是这么一个存在于许多其它语言中而无法在 Go 中找到的概念。

## 为什么需要构造器

在 Go 中，有些对象需要初始化，比如 channel 和 slice 这两个很容易想到的例子。这个初始化的过程通过调用 `make` 函数来执行。

> make 这个内置函数为且仅为 slice, map 以及 chan 类型的对象分配以及初始化。和 new 一样，函数的第一个参数是一个类型而不是值。与 new 不同的是，make 函数返回的是与参数一样的类型而不是参数类型的指针。

当然我们还需要默认参数，所以在 Go 中分配本身并不是我们需要构造函数的唯一原因（事实上其它语言也一样）。由于结构体的定义不允许设置默认值，因此我们需要一个能够设置默认值的函数。

## 原生的构造器

提供构造器的一个通用的惯例是提供 `New` 或者 `NewStructName` 的函数，其中 StructName 是构造函数的返回值。通常而言，`New` 函数用于比较小的能够自予且只导出一个结构体的包中，比如 [fatih/structs](https://godoc.org/github.com/fatih/structs#New)。如果你想在一个包中创建多个对象，那么你最好去看看标准库中的 `time` 包，比如 [time.NewTicker()](https://golang.org/pkg/time/#NewTicker) 以及 [time.NewTimer()](https://golang.org/pkg/time/#NewTicker)

然而 `time` 包导出了许多其它的结构体，为何却没有这样的构造器？

那些没有分配符或者无需设置默认值的结构体提可以完全忽略构造器，这些结构体提供构造器仅仅是为了符合编码的习惯，或者为一些准备长期使用的 API 预留功能以防将来需要用到构造器。当然，也有许多常用的构造器并没有遵循这些方针，比如 `time.Now` 提供了默认值，但它的命名是描述性的，这样当你阅读代码的时候你就知道你期望的是一个当前时间的时间戳，而不是 Unix 纪元之类的值。

## 值接收器

以下就是我所建议的值接收器构造器的方式：

```go
type Person struct {
	name string
}

func (Person) New(name string) *Person {
	return &Person{name}
}
```

你可以通过调用 `Person{}.New("Tit Petric")` 的方式来使用这个构造器，并得到一个初始化后的对象。事实上，我们可以更好地理解我们正在编写的代码，因为我们可以使用一个 `Person` 对象（或者一个指向它的指针），因为这就是我们开始的内容。

## 实际验证

我想说的是与使用普通的函数构造器相比，使用值接收器的构造器并不会产生性能或者内存的问题。你会相信我所说的吗？那么让我用 benchmark 来去除你的旧观念吧。

```
New         2000000     754 ns/op     0 B/op      0 allocs/op
Person.New  2000000     786 ns/op     0 B/op      0 allocs/op
```

你也可以自己运行[测试用例](https://play.golang.org/p/injCAoxZpVg)（将代码拷贝到你本地的 `main.go` 文件中并执行 `go run main.go` - 由于运行时间的限制，你无法在 playground 中运行这段代码）。

依然觉得证据不够充分？让我们以这个[小例子](https://play.golang.org/p/F4xsmeGwy5d)为例并导出汇编代码。将其保存到 `main2.go` 中并执行 `go tool compile -S main2.go > main.s` 。查看新生成的 `main.s` 文件中的汇编代码。我们的构造器的调用主要在第 21 行 和第 25 行：

```asm
(main2.go:21)      MOVQ    AX, (SP)
(main2.go:21)      PCDATA  $0, $0
(main2.go:21)      CALL    runtime.newobject(SB)
(main2.go:21)      MOVQ    8(SP), AX
(main2.go:21)      MOVQ    $10, 8(AX)
(main2.go:21)      MOVL    runtime.writeBarrier(SB), CX
(main2.go:21)      TESTL   CX, CX
(main2.go:21)      JNE     312
(main2.go:21)      LEAQ    go.string."Tit Petric"(SB), CX
(main2.go:21)      MOVQ    CX, (AX)
```

以及

```asm
(main2.go:25)      MOVQ    AX, (SP)
(main2.go:25)      PCDATA  $0, $0
(main2.go:25)      CALL    runtime.newobject(SB)
(main2.go:25)      MOVQ    8(SP), AX
(main2.go:25)      MOVQ    $10, 8(AX)
(main2.go:25)      MOVL    runtime.writeBarrier(SB), CX
(main2.go:25)      TESTL   CX, CX
(main2.go:25)      JNE     279
(main2.go:25)      LEAQ    go.string."Tit Petric"(SB), CX
(main2.go:25)      MOVQ    CX, (AX)
```

两段构造器的代码完全一致，且在两个例子中函数都是完全内联的，所以无论你用哪种方式写得到的汇编代码都是一样的。

对于 "inlining" 不熟悉的同学：

> 在计算机科学中，内联函数是指用于告诉编译器它应该在一个特定的函数上执行在线扩展的一种编程语言概念。换句话说，编译器会把函数的内容作为一个整体插入到每个调用该函数的地方。

在我们的例子中，这意味着在汇编代码只在构造器内部执行一次分配动作。值接收器并没有申明一个变量来接收值，编译器已经对其做了充分的优化。

## 结论

值接收器构造器已经很接近于其他语言中的构造器了，尽管它看起来有些笨拙。在 PHP 中执行 `new Person(...)` 将会调用构造器并返回一个该类型的对象，而 Go 可以使用 `Person{}.New()` 并实现更强大的功能。

类似于这样的构造器：

1. 可以使用多个参数（与 PHP, Java 等相同），
2. 可以返回多个值， 以便适应 Go 的错误处理方式

包括标准库在内的许多包已经提供 `New` 函数。我无法找出所有的提供该功能的包，但 `errors` 包是容易想到的其中之一。对于第二种形式的例子，Google Youtube APIs 提供了 `New(*http.Client) (*Service, error)` 给大家使用。这样的例子比比皆是。

当你在写自己的应用时，公认的思想是，除非你有明确的理由，否则不应该创建子包。更加戏剧性的做法是根据你的每个结构体创建一个包，这样就可以为其提供各自的 `New` 构造器了。在这个例子中第二种方法比较合适。

我知道肯定会有其它的一些顾虑：

* 一旦你选择了这种构造器，那么值接收器将会和 `Person` 的实例绑定，且你不会去定义全局的 `NewPerson`，
  * 在 PHP 中与此类似的是 `Person::New()` 这样的东西，而不是 `__constructor`。没有全局调用。
* Go 编译器完全优化了来自值接收器的隐含开销/分配。事实上从编译后的汇编代码来看无论用哪种方式都没有什么区别，正如上面的例子所示。
  * 这违背了命令式编程或者过程式编程的思想。你会认为 `Person{}` 会执行一次分配动作，然后在执行值接收器时会丢弃该次分配并被垃圾回收器回收。Go 编译器很聪明，事实上它非常聪明，如上所示，它完全优化了这个过程。编译器太过于聪明了。
* 你可以同时使用值接收器构造器和在合适的场合使用构造器函数。这个[例子](https://github.com/titpetric/go-web-crontab/blob/master/crontab/crontab.go)同时提供了 `New()` 和一个 `Crontab{}.New()` 构造器，其它的结构体则提供值接收器构造器。
* 无论你申明了多少结构体，你都可以给每个结构体创建一个 `New` 构造器。当然了，如果你有多种创建对象的方法，你就无需只使用一种构造器。比如 [context 包](https://golang.org/pkg/context/)提供了四种构造器。
* `TODO` 和 `Background` 不会再每次调用时返回新的值。

---

via: https://scene-si.org/2018/03/08/an-argument-for-value-receiver-constructors/

作者：[Tit Petric](https://scene-si.org/about)
译者：[killernova](https://github.com/killernova)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
