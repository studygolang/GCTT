首发于：https://studygolang.com/articles/23441

# Go 常量学习-可视化指南

> Go 的类型常量和非类型常量是两个必须要了解的关键概念

不要忘记在文章下面有很多代码示例，因此你要确保点击这些链接并尝试运行这些程序。

## 你为什么使用常量？

![why_might_you_use_constants?.image ](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-const-guide/1*r734fn1RBz1c1J2cKM7ZGw.png)

你不想在代码中到处定义[魔法数值](https://en.wikipedia.org/wiki/Magic_number_%28programming%29)，于是使用常量来声明它们，并在代码中再次使用它们。

魔法数值是不安全的，你需要准确声明它们，因此，常量是一个安全的选择。除此之外，在代码中看到常量而不是魔法数值也是令人高兴的；人们可以更好地理解代码是怎样的。

我们希望我们可以在使用常量中获得*运行速度上的收益*，因为，使用常量能够使编译器能够进行更多的优化，它将知道常量的值永远不会改变。

我最喜欢的是非类型化常量。它真是个天才的想法。当使用非类型化常量时，你将获得更多的灵活性和高精度的计算。

## 类型常量

![typed_constants.image](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-const-guide/1*4zXKp5xjt-a9ivu9b0vNMw.png)

类型→Boolean,rune,numerics,或者 string

值→编译期时在声明中分配值

地址→你无法得到它在内存中的地址（不像变量）

* 你无法在声明常量之后再改变它
* 你不能使用运行时的结构，例如变量，指针，数组，切片，map,结构体，接口，方法调用，或者方法的值。

## 类型化常量声明

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-const-guide/1*wUbUPm7CFOwWTG_vE5UgmA.png)

*图中定义了一个类型常量 Pi，它的类型为 float64，值为 3.14*

**运行并且尝试代码示例，[请点击这里](https://play.golang.org/p/mrnqxa8Kic)**

## 声明多个常量

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-const-guide/1*JCWkOyIW1KrJUjSdnbGfNw.png)

运行图中的代码并且检验它的结果，[请点击这里](https://play.golang.org/p/mBoqG58z_e)

*在一个代码块中声明多个具有不同类型不同值的常量*

* 当一个常量的类型和值没有声明时，它将从上一个常量中得到它。在上面，pi2 从 pi 中获取其类型和值。
* *Age* 常量在声明时有一个新的值。并且，它通过赋值为 10 获取默认的类型 int。
* 可以在同一行和[多个变量](https://blog.learngoprogramming.com/learn-go-lang-variables-visual-tutorial-and-ebook-9a061d29babe#4176)声明中定义多个常量。

## 非类型化常量

它们有很好的特性，比如高精度的计算以及在所有数值表达式中使用它们而不声明类型等。下面我将介绍这些特性。它们就像 Go 中的通配符。

![iamge](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-const-guide/1*c2tP3ifIOkq2yo0UMAwdDA.png)

理想类型→与 Go 通常类型不同的隐藏类型。

理想值→存在于理想值空间中，并且具有默认类型。

默认类型→取决于理想值。

## 非类型化常量的声明

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-const-guide/1*7b1ZmM39ppGTFs3nLgdMzw.png)

声明了一个非类型化的常量 Pi，并且为它赋值为 3.14，那么它默认的类型就是 float。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-const-guide/1*7cCppzbC1AbmF9u8O75MkQ.png)

当需要它的类型的时候，图片左侧（期望类型）将转化为右边的类型（预先声明的类型）

**尝试代码，点击[这里](https://play.golang.org/p/L5UC3XgYFk)**

## 高精度计算

如果常量只停留在非类型化常量领域，那么它没有速度的限制！但是，当将常量赋值给变量进行使用时，速度就有限制了。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-const-guide/1*YhDCUL1FGF-BbU-yTkxAAA.png)

当你将其分配给变量时，非类型化常量的精度会降低，其默认类型会转换为 Go 的[普通类型](https://golang.org/ref/spec#Boolean_types)。

**运行代码示例，[请点击这里](https://play.golang.org/p/4ODv0n_stw)**

## 灵活的表达方式

你可以使用非类型化常量临时从 Go 的强类型系统中逸出，直到它们在类型要求表达式中的计算为止。

我在[代码中](https://github.com/inancgumus/myhttp/blob/master/get.go#L12)一直使用它们时，会避免在不需要强类型时声明它们。所以，如果你不真正需要常量，就不要用它声明类型。

### 运行代码示例

* [Understand when and how to use untyped constants](https://play.golang.org/p/2cgFoB4rYD)
* [We can assign an untyped constant to any numeric-type variable](https://play.golang.org/p/7-VMh5egC-)

## 常量作用范围

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-const-guide/1*pOohX09A8xYxc4scxpHoRQ.png)

一个常量只能在它的声明的作用域内使用。如果你在更内部的作用域内以同样的名字再声明一个常量，那么这个常量仅仅在内部作用域内可以使用，并且在此作用域内将覆盖外部声明的常量。查看代码示例，[请点击这里](https://play.golang.org/p/c3-GF_a5iI)

---

via: https://blog.learngoprogramming.com/learn-golang-typed-untyped-constants-70b4df443b61

作者：[Inanc Gumus](https://www.activestate.com/blog/author/peteg/)
译者：[xmge](https://github.com/xmge)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
