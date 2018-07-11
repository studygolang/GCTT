已发布：https://studygolang.com/articles/13383

# 为什么 Go 类型系统中没可变性？

**本文解释了协变、逆变和不变性是什么，以及对 Go 类型系统的影响。特别是解释了为什么在 slices 中不可能有可变性。**

一个 Go 初学者经常问的问题是“为什么我不能把 `[]int` 类型变量传递给函数 `func ([]interface{ })`”？在这篇文章中，我想探讨这个问题及其对 Go 的影响。但是可变性(本文所描述)的概念在其他语言也是有用的。

可变性描述了子类型关系应用在复合类型中使用时发生的情况。在这种情况下，“A是B的子类型”意味着，A类型的实例始终可以被用作需要B类型的场景。Go 没有明确的子类型关系，最接近的是可赋值性，它主要决定类型是否可以互换使用。接口也许是最重要的使用场景：如果类型T（无论它是具体类型，还是本身是接口）实现接口I，然后T可以被看作是I的子类型。从这个意义上讲， `*bytes.Buffer` 是 `io.ReadWriter` 的子类型，`io.ReadWriter` 是 `io.Reader` 的子类型。所有类型都是 `interface{}` 的子类型。

理解可变性含义的最简单方法是查看函数类型。假设我们有一个类型和一个子类型，例如 `*bytes.Buffer` 是 `io.Reader` 的子类型。可以定义这样一个函数 `func() *bytes.Buffer`。我们也可以把这个函数用作 `func() io.Reader`,我们只是把返回值重新定义为 `io.Reader`。但反方向的不成立的：我们不能把函数 `func() io.Reader` 用作函数 `func() *bytes.Buffer`，因为不是每个 `io.Reader` 都可以成为 `*bytes.Buffer`。因此，函数返回值可以保持子类型关系的方向为：如果A是B的子类型，则函数 `func() A` 可以是函数 `func() B` 的子类型。这叫做协变。

```go
func F() io.Reader {
   return new(bytes.Buffer)
}

func G() *bytes.Buffer {
   return new(bytes.Buffer)
}

func Use(f func() io.Reader) {
    useReader(f())
}
func main() {
   Use(F) // Works
   Use(G) // Doesn't work right now; but *could* be made equivalent to...
   Use(func() io.Reader { return G() })
}

```

另一方面，假设我们有函数 `func(*bytes.Buffer)`。现在我们不能把它当作函数 `func(io.Reader)`，你不能用 `io.Reader` 作为参数来调用它。但我们可以反方向调用。如果我们用 `*bytes.Buffer` 作为参数，可以用它调用 `func(io.Reader)`。因此，函数的参数颠倒了子类型关系：如果A是B的子类型，那么 `func(B)`可以是 `func(A)` 的子类型。这叫做逆变。

```go
func F(r io.Reader) {
    useReader(r)
}

func G(r *bytes.Buffer) {
    useReader(r)
}

func Use(f func(*bytes.Buffer)) {
    b := new(bytes.Buffer)
    f(b)
}

func main() {
    Use(F) // Doesn't work right now; but *could* be made equivalent to...
    Use(func(r *bytes.Buffer) { F(r) })
    Use(G) // Works
}

```

因此，`func` 对于参数是逆变值的，对于返回值是协变的。当然，我们可以将这两种性质结合起来：如果A和C分别是B和D的子类型，我们可以使 `func(B) C` 成为 `func(A) D` 的子类型，可以这样转换：

```go
// *os.PathError implements error

func F(r io.Reader) *os.PathError {
    // ...
}

func Use(f func(*bytes.Buffer) error) {
    b := new(bytes.Buffer)
    err := f(b)
    useError(err)
}

func main() {
    Use(F) // Could be made to be equivalent to
    Use(func(r *bytes.Buffer) error { return F(r) })
}

```
然而，`func(A) C` 和 `func(B) D` 是不兼容的。一个也不能成为另一个的子类型。

```go
func F(r *bytes.Buffer) *os.PathError {
    // ...
}

func UseF(f func(io.Reader) error) {
    b := strings.NewReader("foobar")
    err := f(b)
    useError(err)
}

func G(r io.Reader) error {
    // ...
}

func UseG(f func(*bytes.Buffer) *os.PathErorr) {
    b := new(bytes.Buffer)
    err := f()
    usePathError(err)
}

func main() {
    UseF(F) // Can't work, because:
    UseF(func(r io.Reader) error {
        return F(r) // type-error: io.Reader is not *bytes.Buffer
    })

    UseG(G) // Can't work, because:
    UseG(func(r *bytes.Buffer) *os.PathError {
        return G(r) // type-error: error is not *os.PathError
    })
}

```

因此，在这种情况下，复合类型之间没有关系。这叫做不变性。

现在，我们可以回到我们的问题：为什么不能将 `[]int` 作为 `[]interface{}` 来使用？这实际上是问：“为什么 `slices` 类型是不变的”？提问者假设，因为 `int` 是 `interface{}` 的子类型，所以 `[]int` 也应该是 `[]interface{}` 的子类型。然而，我们现在可以看一个简单的问题。`slices` 支持（除了别的之外）两个基本操作，我们可以粗略地转化成函数调用：

```go
as := make([]A, 10)
a := as[0] // func Get(as []A, i int) A
as[1] = a // func Set(as []A, i int, a A)
```

这明显出现了问题：类型A既作为参数出现，也作为返回类型出现。因此，它既有协变又有逆变。因此，在调用函数时有一个相对明确的答案来解释可变性如何工作，它只是对于 `slices` 没有太多的意义。读取 `slices` 需要协变，但写入 `slices` 需要逆变。换句话说，如果你需要使 `[]int` 成为 `[]interface{}` 的子类，你需要解释这段代码是如何工作的：

```go
func G() {
    v := []int{1,2,3}
    F(v)
    fmt.Println(v)
}

func F(v []interface{}) {
    // string is a subtype of interface{}, so this should be valid
    v[0] = "Oops"
}
```

`channel` 提供了另一个有趣的视角。双向 `channel` 类型具有与 `slices` 类型相同的问题：接收时需要协变，而发送时需要逆变。但你可以限制 `channel` 的方向，只允许发送或接收操作。所以 `chan A` 和 `chan B` 可以没有关系，我们可以使 `<-chan A` 成为 `<-chan B` 的子类，或 `chan<-B` 成为 `chan<-A` 的子类。

在这种意义上，只读类型至少在理论可以允许 `slices` 的可变性。`[]int` 仍然不是 `[]interface{}` 的子类型，我们可以使 `ro[] int` 成为 `ro []interface` 的子类型（借用proposal中的语法）。

最后，我想强调的是，所有这些都只是理论上为 Go 类型系统添加可变性的问题。我认为这很难，但即使我们能解决这些问题，仍然会遇到一些实际问题。其中最紧迫的是子类型的内存结构不同：

```go
var (
    // super pseudo-code to illustrate
    x *bytes.Buffer // unsafe.Pointer
    y io.ReadWriter // struct{ itable *itab; value unsafe.Pointer }
                    // where itable has two entries
    z io.Reader     // struct{ itable *itab; value unsafe.Pointer }
                    // where itable has one entry
)
```

因此，即使你认为所有接口都具有相同的内存模型，它们实际上没有，因为方法表具有不同的假定布局。所以在这样的代码中

```go
func Do(f func() io.Reader) {
    r := f()
    r.Read(buf)
}

func F() io.Reader {
    return new(bytes.Buffer)
}

func G() io.ReadWriter {
    return new(bytes.Buffer)
}

func H() *bytes.Buffer {
    return new(bytes.Buffer)
}

func main() {
    // All of F, G, H should be subtypes of func() io.Reader
    Do(F)
    Do(G)
    Do(H)
}
```

还需要在某个地方将H返回的 `io.ReadWriter` 接口包装成 `io.Reader` 接口，并需要在某个地方将G的返回的 `*bytes.Buffer` 可转换为正确的 `io.Reader` 接口。这对于函数来说，不是一个大问题：编译器可以在 `main` 函数调用时生成合适的包装。当代码中使用这种形式的子类型时会有一定的性能开销。然而，这对于 `slices` 来说是一个很重要的问题。

对于 `slices` 我们有两种处理方式。(a)将 `[]int` 转换为 `[]interface{}` 进行传递，意味着一个分配并进行完整的拷贝。(b)延迟 `int` 与 `interface{}` 的转换，直到需要进行访问时在进行转换。这意味着现在每个 `slices` 访问都必须通过一个间接函数调用，以防万一有人传递给我们一个子类型。这两种选择都不符合 Go 的设计目标。

---

via: https://blog.merovius.de/2018/06/03/why-doesnt-go-have-variance-in.html

作者：[Axel Wagner](https://github.com/Merovius)
译者：[Alan](https://github.com/althen)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
