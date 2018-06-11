已发布：https://studygolang.com/articles/12136

# Golang 中 defer 的五个坑 - 第二部分

本文承接[第一部分](https://studygolang.com/articles/12061)的内容继续讲解 defer 的一些常见陷阱

## 1——Z 到 A（译注：倒序）

当你第一次学习 Go 的时候可能会中招。

例子

```go
func main() {
  for i := 0; i < 4; i++ {
    defer fmt.Print(i)
  }
}
```

输出

Go 的运行时会将延迟执行的函数保存至一个栈中（译注：意味着它们会按照入栈的顺序倒序执行）。想了解更多，请阅读这篇 [文章](https://blog.learngoprogramming.com/golang-defer-simplified-77d3b2b817ff#702e) [GCTT 出品的译文：图解 Go 中的延迟调用 defer](https://studygolang.com/articles/11907)。

```
3
2
1
0
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/5-gotchas-defer-2/a_to_z.png)

## 2——作用域屏蔽了参数

事实上这是一个作用域的坑，但我想要让你知道它与 **defer** 和 [已命名的返回值](https://blog.learngoprogramming.com/golang-funcs-params-named-result-values-types-pass-by-value-67f4374d9c0a#7cf4) 之间的关系。

例子

这里我们定义了一个函数，这个函数声明了它需要延迟的函数（在返回的时候释放资源 `r` ）。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/5-gotchas-defer-2/scope_ate.png)

### 接着创建一个 reader 类型的结构体使得调用 Close 的时候返回一个 error

```go
type reader struct{}

func (r reader) Close() error {
  return errors.New("Close Error")
}
```

当 `reader` 调用 `Close()` 的时候总会返回一个 error ，`release` 会在 **defer** 内部调用。

```go
r := reader{}

err := release(r)

fmt.Print(err)
```

输出

```
nil
```

变量 `err` 的值为 `nil` ，而我们期望中的值应该是 *"Close Error"* 。

### 为什么会这样？

延迟函数内的赋值语句在延迟函数的 `if` 块中，因此在块中的 `err` 变量赋值会创建一个全新的变量，块级变量 `err` 的作用域会屏蔽返回变量 `err` ，因此， `release()` 还是返回 `err` 的原始值。

### 解决方案

我们需要将值赋给 `release()` 的 `err` 变量，为了做到这一点，请看下面图示的第三行，这行代码并没有声明新的 `err` 变量，与之前不同的是，它使用的是原本的 `err` 返回变量（将 `:=` 替换成 `=`）。这就解决了块级屏蔽的问题。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/5-gotchas-defer-2/scope_ate_solution.png)

如果你对 `=` 和 `:=` ，请看一下这篇 [文章](https://blog.learngoprogramming.com/learn-go-lang-variables-visual-tutorial-and-ebook-9a061d29babe)

## 3——参数很快得到了值

当一个 **defer** 函数出现而不是被执行的时候，传递给它的参数的值就会被立刻确定下来，在我的其他文章中有讲过类似的 [例子](https://blog.learngoprogramming.com/golang-defer-simplified-77d3b2b817ff#649d) 。

例子

```go
type message struct {
  content string
}

func (p *message) set(c string) {
  p.content = c
}

func (p *message) print() string {
  return p.content
}
```

试着运行一下上面的代码

```go
func() {
  m := &message{content: "Hello"}

  defer fmt.Print(m.print())

  m.set("World")

  // 被延迟的函数在这里执行
}
```

输出

```
"Hello"
```

### 为什么输出不是 "World" ?

在 **defer** 中， `fmt.Print` 被推迟到函数返回后执行，可是 `m.print()` 的值在当时就已经被求出，因此， `m.print()` 会返回 *"Hello"* ，这个值会保存一直到外围函数返回。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/5-gotchas-defer-2/why_output_not_world.png)

这个过程类似于调用 `fmt.Print()` 打印 *"Hello"* ，进一步理解这个过程，可以看这篇 [文章](https://blog.learngoprogramming.com/golang-defer-simplified-77d3b2b817ff#649d) 。

## 4——循环中存址

当延迟函数执行的时候，会查看当时周围变量中的值 —— 除了被传入参数的值。我们来看看这一切是怎么在循环中发生的。

例子

我们在循环里定义了一个闭包函数。

```go
for i := 0; i < 3; i++ {
  defer func() {
   fmt.Println(i)
  }()
}
```

输出

```
3
3
3
```

### 为什么？

当代码执行的时候，被延迟的函数会查看当时 `i` 的值，这是因为，当 defer 出现的时候， Go 的运行时会保存 `i` 的地址，在 for 循环结束之后 `i` 的值变成了 3。 因此，当延迟语句运行的时候，所有的延迟栈中的语句都会去查看 `i` 地址中的值，也就是 3 （被当做循环结束后的当时值）。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/5-gotchas-defer-2/capture_in_the_loop.png)

所有的延迟函数会查看相同的 `i` ，循环结束（值变成了3)，因此它们查看到的都是 3 。

### 解决方案

直接向延迟函数传参数。

```go
for i := 0; i < 3; i++ {
  defer func(i int) {
   fmt.Println(i)
  }(i)
}
```

```
2
1
0
```

### 为什么这个又能正常执行？

因为，这时 Go 的运行时在 for 循环的作用域中又创建了不同的 `i` 变量并把正确的值也保存在了那些变量中，现在，延迟函数不会去查看原始循环中的 `i` 变量，它们会查看各自的局部 `i` 变量。

### 解决方案 #2

在这里，我们故意使用一个新的 `i` 块级变量来屏蔽循环中的 `i` ，说实话我不太喜欢这种风格，所以我把这种方式放到了 [Hacker News](https://news.ycombinator.com/item?id=15979751) 上供大家讨论。

```go
for i := 0; i < 3; i++ {
  i := i
  defer func() {
    fmt.Println(i)
  }()
}
```

### 解决方案 #3

如果一次延迟调用一个函数，你可以直接对这个函数使用 **defer** 。

```go
for i := 0; i < 3; i++ {
  defer fmt.Println(i)
}
```

## 5——不返回的意义

对于调用者来说，在延迟函数中返回值几乎没有什么影响，可是，你依然可以使用 [命名返回值](https://blog.learngoprogramming.com/golang-funcs-params-named-result-values-types-pass-by-value-67f4374d9c0a#7cf4) 来影响返回的结果。

例子

```go
func release() error {
  defer func() error {
    return errors.New("error")
  }()

  return nil
}
```

输出

```
nil
```

### 解决方案

```go
func release() (err error) {
  defer func() {
    err = errors.New("error")
  }()

  return nil
}
```

输出

```
"error"
```

### 为什么能起作用？

在这里，我们在 **defer** 中，对 函数的 `err` 返回变量赋了新的值，并且函数会返回那个值，defer 函数并不直接返回值，它影响了返回命名变量的这个过程。

## 小结

你不需要处处使用 defer ，没有 defer 你也可以直接返回 error 。

在函数有多个返回情况而你想要集中处理它们的时候使用 defer 会很方便，多思考如何进一步简化你的代码。

---

via: https://blog.learngoprogramming.com/5-gotchas-of-defer-in-go-golang-part-ii-cc550f6ad9aa

作者：[Inanc Gumus](https://blog.learngoprogramming.com/@inanc)
译者：[yujiahaol68](https://github.com/yujiahaol68)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
