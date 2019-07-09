#### https://blog.learngoprogramming.com/learn-golang-typed-untyped-constants-70b4df443b61

# Go 常量学习-可视化指导

> go 的类型常量和非类型常量是两个必须要了解的关键概念

不要忘记在文章下面有很多代码示例，因此你要确保亲自跑一下这些程序。

## 你为什么使用常量？

![why_might_you_use_constants?.image ](https://cdn-images-1.medium.com/max/2400/1*r734fn1RBz1c1J2cKM7ZGw.png)

你不想在代码中到处定义[魔法变量](https://en.wikipedia.org/wiki/Magic_number_%28programming%29)，于是使用常量来声明它们，并在代码中再次使用它们。

魔法变量是不安全的，你需要准确声明它们，因此，常量是一个安全的选择。除此之外，在代码中看到常量而不是魔法值也是令人高兴的；人们可以更好地理解代码是怎样的。

我们希望我们可以在使用常量中获得*巨大的收益*，因为，编译器能够进行更多的优化，因为它知道常量的值永远不会改变。

我最喜欢的是非类型化常量。他们真是个天才的主意。当使用非类型化常量时，您将获得灵活性和高精度的计算。

## 类型常量

![typed_constants.image](https://cdn-images-1.medium.com/max/1600/1*4zXKp5xjt-a9ivu9b0vNMw.png)

类型 → Boolean,rune,numerics,或则 string
值 → 编译器的时候在声明中分配值
地址 → 你无法得到它的在内存中的地址（不像变量）

* 你无法在声明常量之后再改变它
* 你不能使用运行时的结构，例如 例如变量，指针，数组，切片，map,结构体，接口，方法调用，或则方法的值。

## 类型化常量声明

![image](https://cdn-images-1.medium.com/max/1600/1*wUbUPm7CFOwWTG_vE5UgmA.png)

**图中定义了一个类型常量 Pi，它的类型为 float64，值为 3.14**

*运行并且尝试代码示例，[请点击这里](https://play.golang.org/p/mrnqxa8Kic)*







---

via: https://blog.learngoprogramming.com/learn-golang-typed-untyped-constants-70b4df443b61

作者：[Inanc Gumus](https://www.activestate.com/blog/author/peteg/)
译者：[xmge](https://github.com/xmge)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
