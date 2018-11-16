# Go面向对象编程

今天有人在论坛上问我，怎么在不使用内嵌的方式下更好的使用继承。很重要的一点是，每个人都应当考虑 Go 而不是他们使用的其他语言。我不会告诉你我在 Go 的早期实现中删除了多少代码，因为这些都不重要。语言设计师拥有多年的经验和知识，事后审校有助于创建一个快速、精简而有趣的语言。

我认为 Go 是一个很不错的面向对象编程语言。诚然，它有封装和类型成员方法，但是它还是缺少继承和传统的多态性。对于我来说，继承用处不大，除非你想要实现多态。Go 实现了接口类型，所以继承就显得不那么重要了。Go 采用了 OOP 最佳的部分，省略了其余部分并为我们提供了更好的编写多态代码的方法。

以下是 Go 中 OOP 的一个简单描述。最开始是三个结构：

```go
type Animal struct {
    Name string
    mean bool
}

type Cat struct {
    Basics Animal
    MeowStrength int
}

type Dog struct {
    Animal
    BarkStrength int
}
```

你可能在其他的 OOP 编程示例中见过这三个结构体。



















---

via: https://www.ardanlabs.com/blog/2013/07/object-oriented-programming-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[Tyrodw](https://github.com/tyrodw)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
