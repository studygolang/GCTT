# 鸭子类型 vs 结构化类型 vs 标称类型

GO 语言是哪一种？

编程语言具有类型概念 - 布尔类型，字符串，整型或者被称为类或者结构体的更加复杂的结构。根据如何将类型解析并赋值给各种构造（例如变量，表达式，函数，函数参数等），编程语言可以归类为鸭子类型，结构化类型或标称类型。

本质上，分类决定了对象如何被解析并推断为具体的类型。

![img](https://github.com/studygolang/gctt-images2/blob/master/20200608-duck-typing-vs-structural-typing-vs-nominal-typing/1_zPb6iQvY7faJQ12GqCpqrQ.png?raw=true)

**鸭子类型语言**使用鸭子测试来评估对象是否可以被解析为特定的类型。鸭子测试表示：
> 如果它看起来像鸭子，像鸭子一样游泳，像鸭子一样嘎嘎叫，那它很可能就是鸭子。

**我将使用 GO 语言语法来解释这些想法 - 将这些示例作为伪代码阅读 - 它与 GO 语言规则无关*

以下代码片段是鸭子类型语言的示例。因为 Mallard 可以嘎嘎叫，所以它是一只鸭子。

```go
type Mallard struct {
}
func (m Mallard) quack() {
}
func makeDuckQuack(duck Duck) {
    duck.quack()
}
func main() {
     makeDuckQuack(Mallard{})
}
```

在上面的示例中，Duck 可以是任意类型，它可以是接口或者另一个类型，但是对于 makeDuckQuack 函数而言，唯一的要求就是传递一个可以执行 quack 函数的对象作为参数。

鸭子类型语言通常没有编译期检查。类型的解析和解释发生在运行时 - 这可能导致运行时错误。

例如，以下代码片段可以正常编译，但是由于 *Dog* 类型不支持 *quack()* 函数，会产生运行时错误。

```go
type Dog struct {
}
func (d Dog) bark() {
}
func makeDuckQuack(duck Duck) {
    duck.quack()
}
func main() {
     makeDuckQuack(Dog{})
}
```

Python 和 Javascript 是流行的鸭子类型语言。

在另一端，**标称类型语言**期望程序员明确地对类型进行调用以供编译器解释。

```go
type Duck interface {
     quack()
}
type Mallard struct { //Mallard doesn't implement Duck interface
}
func (m Mallard) quack() {
}
func makeDuckQuack(duck Duck) {
    duck.quack()
}
func main() {
     makeDuckQuack(Mallard{}) //This will not work as Mallard doesn't explicitly implement Duck.
}
```

在上面的示例中，程序员需要明确地实现 Duck 接口。可以说，显示关系意味着更强的可读性。

明确定义 Marllard 和 Duck 之间的关系也意味着包含 Marllard 结构的包依赖于包含 Duck 结构的包。这可能永远也不是一件好事，并且增加了整个应用程序的复杂性。

标称类型语言主要包括 Java，C++。

结构化类型语言介于两者之间，应用程序员无需明确定义用于解释的类型，但是编译器会进行编译期检测来确保程序的完整性。

```go
type Duck interface {
     quack()
}
type Mallard struct { //Mallard doesn't implement Duck interface
}
func (m Mallard) quack() {
}
type Dog struct {
}
func (d Dog) bark() {
}
func makeDuckQuack(duck Duck) {
    duck.quack()
}
func main() {
     makeDuckQuack(Mallard{}) // Okay
     makeDuckQuack(Dog{}) // Not Okay
}
```

在上面的示例中，程序员无需指定 Mallard 是 Duck 类型。语言编译器将 Mallard 解释为 Duck - 因为它具有 quack 函数。但是 Dog 不是一个 Duck，因为 Dog 不具有 quack 函数。

GO 是结构化类型语言。

## 结论

*鸭子类型语言*为程序员提供了最大的灵活性。程序员只需写最少量的代码。但是这些语言可能并不安全，会产生运行时错误。

*标称类型语言*要求程序员显示调用类型，这意味着更多的代码和更少的灵活性（附加的依赖）。

*结构化类型语言*提供了一种平衡，它需要编译期检查，但不需要显示声明依赖。

在使用 GO（结构化类型）编程之前，我主要使用 Java（标称类型）。我喜欢结构化类型语言提供的灵活性，而又不会影响编译期类型安全。

---
via: https://medium.com/higher-order-functions/duck-typing-vs-structural-typing-vs-nominal-typing-e0881860bf10

作者：[Saurabh Nayar](https://medium.com/@nayar.saurabh)
译者：[DoubleLuck](https://github.com/DoubleLuck)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
