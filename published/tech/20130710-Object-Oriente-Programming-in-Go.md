首发于：https://studygolang.com/articles/17340

# Go 面向对象编程

今天有人在论坛上问我，怎么在不使用内嵌的方式下更好的使用继承。很重要的一点是，每个人都应当考虑 Go 而不是他们使用的其他语言。我不会告诉你我在 Go 的早期实现中删除了多少代码，因为这些都不重要。语言设计师拥有多年的经验和知识，事后审校有助于创建一个快速、精简而有趣的语言。

我认为 Go 是一个很不错的面向对象编程语言。诚然，它有封装和类型成员方法，但是它还是缺少继承和传统的多态性。对于我来说，继承用处不大，除非你想要实现多态。Go 实现了接口类型，所以继承就显得不那么重要了。Go 完成了 OOP 最佳的部分，省略了其余部分并为我们提供了更好的编写多态代码的方法。

以下是 Go 中关于 OOP 的一个简单描述。最开始是三个结构：

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

你可能在其他的 OOP 编程示例中见过这三个结构。我们有一个基础结构和两个基于这个基础结构衍生来的结构。结构 Animal 包含所有动物的共性，而另外两个结构则特指猫和狗。

结构 Animal 的所有的成员属性除了私有的 `mean` 以外都是公有的，私有属性以小写字母开头。在 Go 中，对于变量、结构、属性和方法等，由第一个字母决定其访问规范。大写字母表示公有，而小写字母表示私有。

注：Go 中的私有和公有的概念并不完全正确。

https://www.ardanlabs.com/blog/2014/03/exportedunexported-identifiers-in-go.html

由于 Go 中没有继承，所以使用组合是唯一的选择。结构 Cat 拥有一个叫做 Basics 的属性，它的类型是 Animal。结构 Dog 使用了类型为 Animal 的匿名结构（内嵌的形式）。我会将这两种实现方式都展示出来，怎么使用由你自己决定。

我想先感谢 John McLaughlin 对于匿名结构的评论。

为结构 Cat 和 Dog 分别创建一个成员函数（方法），语法如下：

```go
func (dog *Dog) MakeNoise() {
    barkStrength := dog.BarkStrength

    if dog.mean == true {
        barkStrength = barkStrength * 5
    }

    for bark := 0; bark < barkStrength; bark++ {
        fmt.Printf("BARK ")
    }

    fmt.Println("")
}

func (cat *Cat) MakeNoise() {
    meowStrength := cat.MeowStrength

    if cat.Basics.mean == true {
        meowStrength = meowStrength * 5
    }

    for meow := 0; meow < meowStrength; meow++ {
        fmt.Printf("MEOW ")
    }

    fmt.Printlnf("")
}
```

我们在每个结构类型的方法前面指定了该方法的接收者为指针类型。现在 Cat 和 Dog 都有了名为 MakeNoise 的方法。

上面的方法执行同样的功能。每一个动物都会依据它们自身的意愿使用属于它们自己的语言发出犬吠或是猫叫声。尽管有些不合时宜，但是这想你展示了如何访问引用的对象（值）。

当接收 Dog 的引用时，我们是直接访问了它的 Animal 属性。而接收 Cat 引用时，我们使用了它的 Basics 成员属性。

目前，我们已经讲述了封装、组合、访问规范以及成员方法。剩下的就是如何创建多态性了。

我们使用接口来创建多态性：

```go
type AnimalSounder interface {
    MakeNoise()
}

func MakeSomeNoise(animalSounder AnimalSounder) {
    animalSounder.MakeNoise()
}
```

这里我们增加了一个接口和一个公有的函数，函数接收接口类型的值。实际上，该函数引用的是实现此接口类型的值。接口类型是不可实例化的类型，接口声明了一种行为而由其他类型实现这种行为。

Go 约定，当接口只包含一个方法时，接口使用 “ er ” 后缀命名。

在 Go 语言中，任何类型通过成员方法的形式实现了接口，则该类型即可表示这个接口类型。在我们的例子中，结构 Cat 和 Gog 都使用带有指针接收器的方法实现了接口 AnimalSounder，所以它们都可以被看做是 AnimalSounder 类型的。

如果你想减少 Cat 和 Dog 中重复的 MakeNoise 方法代码，你可以创建一个属于类型 Animal 的方法来完成：

```go
func (animal *Animal) PerformNoise(strength int, sound string) {
    if animal.mean == true {
        strength = strength * 5
    }

    for voice := 0; voice < strength; voice++ {
        fmt.Printf("%s ", sound)
    }

    fmt.Println("")
}

func (dog *Dog) MakeNoise() {
    dog.PerformNoise(dog.BarkStrength, "BARK")
}

func (cat *Cat) MakeNoise() {
    cat.Basics.PerformNoise(cat.MeowStrength, "MEOW")
}
```

现在类型 Animal 有一个可以发声的业务逻辑方法。这部分业务逻辑属于其所处于的类型。这样做的好处是我们不必将属性 `mean` 作为参数传递，因为它本就属于类型 Animal。

这是完整的示例代码：

```go
package main

import (
    "fmt"
)

type Animal struct {
    Name string
    mean bool
}

type AnimalSounder interface {
    MakeNoise()
}

type Dog struct {
    Animal
    BarkStrength int
}

type Cat struct {
    Basics Animal
    MeowStrength int
}

func main() {
    myDog := &Dog{
        Animal{
           "Rover", // Name
           false,   // mean
        },
        2, // BarkStrength
    }

    myCat := &Cat{
        Basics: Animal{
            Name: "Julius",
            mean: true,
        },
        MeowStrength: 3,
    }

    MakeSomeNoise(myDog)
    MakeSomeNoise(myCat)
}

func (animal *Animal) PerformNoise(strength int, sound string) {
    if animal.mean == true {
        strength = strength * 5
    }

    for voice := 0; voice < strength; voice++ {
        fmt.Printf("%s ", sound)
    }

    fmt.Println("")
}

func (dog *Dog) MakeNoise() {
    dog.PerformNoise(dog.BarkStrength, "BARK")
}

func (cat *Cat) MakeNoise() {
    cat.Basics.PerformNoise(cat.MeowStrength, "MEOW")
}

func MakeSomeNoise(animalSounder AnimalSounder) {
    animalSounder.MakeNoise()
}
```

输出如下：

```
BARK BARK
MEOW MEOW MEOW MEOW MEOW MEOW MEOW MEOW MEOW MEOW MEOW MEOW MEOW MEOW MEOW
```

有人在版面上贴了关于在结构中内嵌接口的示例代码，如下：

```go
package main

import (
    "fmt"
)

type HornSounder interface {
    SoundHorn()
}

type Vehicle struct {
    List [2]HornSounder
}

type Car struct {
    Sound string
}

type Bike struct {
   Sound string
}

func main() {
    vehicle := new(Vehicle)
    vehicle.List[0] = &Car{"BEEP"}
    vehicle.List[1] = &Bike{"RING"}

    for _, hornSounder := range vehicle.List {
        hornSounder.SoundHorn()
    }
}

func (car *Car) SoundHorn() {
    fmt.Println(car.Sound)
}

func (bike *Bike) SoundHorn() {
    fmt.Println(bike.Sound)
}

func PressHorn(hornSounder HornSounder) {
    hornSounder.SoundHorn()
}
```

这个例子中，结构 Vehicle 持有一个实现了 HornSounder 接口的列表。在 main 函数中我们创建了一个 Vehicle 的实例，并将 Car 和 Bike 类型的指针指派给它。这种指派是允许的，因为 Car 和 Bike 都实现了接口 HornSounder。然后使用一个简单的循环，我们将利用接口来发出喇叭声。

你在应用程序实现 OOP 所需要的一切在 Go 中都有。正如我之前所说的，Go 完成了 OOP 中最佳的部分，省略了其余部分并为我们提供了更好的编写多态代码的方法。

要了解相关主题的更多信息请选择以下帖子：

https://www.ardanlabs.com/blog/2014/05/methods-interfaces-and-embedded-types.html

https://studygolang.com/articles/13285

https://studygolang.com/articles/12728

希望我的这个小例子能够对你未来的 Go 编程有所帮助。

---

via: https://www.ardanlabs.com/blog/2013/07/object-oriented-programming-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[Tyrodw](https://github.com/tyrodw)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
