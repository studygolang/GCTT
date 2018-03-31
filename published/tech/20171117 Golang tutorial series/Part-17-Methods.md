已发布：https://studygolang.com/articles/12264

# 第17部分：方法

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2) 的第 17 个教程。

### 什么是方法？

方法其实就是一个函数，在 `func` 这个关键字和方法名中间加入了一个特殊的接收器类型。接收器可以是结构体类型或者是非结构体类型。接收器是可以在方法的内部访问的。

下面就是创建一个方法的语法。

```go
func (t Type) methodName(parameter list) {
}
```

上面的代码片段创建了一个接收器类型为 `Type` 的方法 `methodName`。

### 方法示例

让我们来编写一个简单的小程序，它会在结构体类型上创建一个方法并调用它。

```go
package main

import (
    "fmt"
)

type Employee struct {
    name     string
    salary   int
    currency string
}

/*
  displaySalary() 方法将 Employee 做为接收器类型
*/
func (e Employee) displaySalary() {
    fmt.Printf("Salary of %s is %s%d", e.name, e.currency, e.salary)
}

func main() {
    emp1 := Employee {
        name:     "Sam Adolf",
        salary:   5000,
        currency: "$",
    }
    emp1.displaySalary() // 调用 Employee 类型的 displaySalary() 方法
}
```

[在线运行程序](https://play.golang.org/p/rRsI_sWAOZ)

在上面程序的第 16 行，我们在 `Employee` 结构体类型上创建了一个 `displaySalary` 方法。displaySalary()方法在方法的内部访问了接收器 `e Employee`。在第 17 行，我们使用接收器 `e`，并打印 employee 的 name、currency 和 salary 这 3 个字段。

在第 26 行，我们调用了方法 `emp1.displaySalary()`。

程序输出：`Salary of Sam Adolf is $5000`。

### 为什么我们已经有函数了还需要方法呢？

上面的程序已经被重写为只使用函数，没有方法。

```go
package main

import (
    "fmt"
)

type Employee struct {
    name     string
    salary   int
    currency string
}

/*
displaySalary()方法被转化为一个函数，把 Employee 当做参数传入。
*/
func displaySalary(e Employee) {
    fmt.Printf("Salary of %s is %s%d", e.name, e.currency, e.salary)
}

func main() {
    emp1 := Employee{
        name:     "Sam Adolf",
        salary:   5000,
        currency: "$",
    }
    displaySalary(emp1)
}
```

[在线运行程序](https://play.golang.org/p/dFwObgCUU0)

在上面的程序中，`displaySalary` 方法被转化为一个函数，`Employee` 结构体被当做参数传递给它。这个程序也产生完全相同的输出：`Salary of Sam Adolf is $5000`。

既然我们可以使用函数写出相同的程序，那么为什么我们需要方法？这有着几个原因，让我们一个个的看看。

- [ Go 不是纯粹的面向对象编程语言](https://golang.org/doc/faq#Is_Go_an_object-oriented_language)，而且Go不支持类。因此，基于类型的方法是一种实现和类相似行为的途径。

- 相同的名字的方法可以定义在不同的类型上，而相同名字的函数是不被允许的。假设我们有一个 `Square` 和 `Circle` 结构体。可以在 `Square` 和 `Circle` 上分别定义一个 `Area` 方法。见下面的程序。

```go
package main

import (
    "fmt"
    "math"
)

type Rectangle struct {
    length int
    width  int
}

type Circle struct {
    radius float64
}

func (r Rectangle) Area() int {
    return r.length * r.width
}

func (c Circle) Area() float64 {
    return math.Pi * c.radius * c.radius
}

func main() {
    r := Rectangle{
        length: 10,
        width:  5,
    }
    fmt.Printf("Area of rectangle %d\n", r.Area())
    c := Circle{
        radius: 12,
    }
    fmt.Printf("Area of circle %f", c.Area())
}
```

[在线运行程序](https://play.golang.org/p/0hDM3E3LiP)

该程序输出：

```
Area of rectangle 50
Area of circle 452.389342
```

上面方法的属性被使用在接口中。我们将在接下来的教程中讨论这个问题。

### 指针接收器与值接收器

到目前为止，我们只看到了使用值接收器的方法。还可以创建使用指针接收器的方法。值接收器和指针接收器之间的区别在于，在指针接收器的方法内部的改变对于调用者是可见的，然而值接收器的情况不是这样的。让我们用下面的程序来帮助理解这一点。

```go
package main

import (
    "fmt"
)

type Employee struct {
    name string
    age  int
}

/*
使用值接收器的方法。
*/
func (e Employee) changeName(newName string) {
    e.name = newName
}

/*
使用指针接收器的方法。
*/
func (e *Employee) changeAge(newAge int) {
    e.age = newAge
}

func main() {
    e := Employee{
        name: "Mark Andrew",
        age:  50,
    }
    fmt.Printf("Employee name before change: %s", e.name)
    e.changeName("Michael Andrew")
    fmt.Printf("\nEmployee name after change: %s", e.name)

    fmt.Printf("\n\nEmployee age before change: %d", e.age)
    (&e).changeAge(51)
    fmt.Printf("\nEmployee age after change: %d", e.age)
}
```

[在线运行程序](https://play.golang.org/p/tTO100HmUX)

在上面的程序中，`changeName` 方法有一个值接收器 `(e Employee)`，而 `changeAge` 方法有一个指针接收器 `(e *Employee)`。在 `changeName` 方法中对 `Employee` 结构体的字段 `name` 所做的改变对调用者是不可见的，因此程序在调用 `e.changeName("Michael Andrew")` 这个方法的前后打印出相同的名字。由于 `changeAge` 方法是使用指针 `(e *Employee)` 接收器的，所以在调用 `(&e).changeAge(51)` 方法对 `age` 字段做出的改变对调用者将是可见的。该程序输出如下：

```
Employee name before change: Mark Andrew
Employee name after change: Mark Andrew

Employee age before change: 50
Employee age after change: 51
```

在上面程序的第 36 行，我们使用 `(&e).changeAge(51)` 来调用 `changeAge` 方法。由于 `changeAge` 方法有一个指针接收器，所以我们使用 `(&e)` 来调用这个方法。其实没有这个必要，Go语言让我们可以直接使用 `e.changeAge(51)`。`e.changeAge(51)` 会自动被Go语言解释为 `(&e).changeAge(51)`。

下面的[程序](https://play.golang.org/p/nnXBsR3Uc8)重写了，使用 `e.changeAge(51)` 来代替 `(&e).changeAge(51)`，它输出相同的结果。

```go
package main

import (
    "fmt"
)

type Employee struct {
    name string
    age  int
}

/*
使用值接收器的方法。
*/
func (e Employee) changeName(newName string) {
    e.name = newName
}

/*
使用指针接收器的方法。
*/
func (e *Employee) changeAge(newAge int) {
    e.age = newAge
}

func main() {
    e := Employee{
        name: "Mark Andrew",
        age:  50,
    }
    fmt.Printf("Employee name before change: %s", e.name)
    e.changeName("Michael Andrew")
    fmt.Printf("\nEmployee name after change: %s", e.name)

    fmt.Printf("\n\nEmployee age before change: %d", e.age)
    e.changeAge(51)
    fmt.Printf("\nEmployee age after change: %d", e.age)
}
```

[在线运行程序](https://play.golang.org/p/nnXBsR3Uc8)

### 那么什么时候使用指针接收器，什么时候使用值接收器？

一般来说，指针接收器可以使用在：对方法内部的接收器所做的改变应该对调用者可见时。

指针接收器也可以被使用在如下场景：当拷贝一个结构体的代价过于昂贵时。考虑下一个结构体有很多的字段。在方法内使用这个结构体做为值接收器需要拷贝整个结构体，这是很昂贵的。在这种情况下使用指针接收器，结构体不会被拷贝，只会传递一个指针到方法内部使用。

在其他的所有情况，值接收器都可以被使用。

### 匿名字段的方法

属于结构体的匿名字段的方法可以被直接调用，就好像这些方法是属于定义了匿名字段的结构体一样。

```go
package main

import (
    "fmt"
)

type address struct {
    city  string
    state string
}

func (a address) fullAddress() {
    fmt.Printf("Full address: %s, %s", a.city, a.state)
}

type person struct {
    firstName string
    lastName  string
    address
}

func main() {
    p := person{
        firstName: "Elon",
        lastName:  "Musk",
        address: address {
            city:  "Los Angeles",
            state: "California",
        },
    }

    p.fullAddress() //访问 address 结构体的 fullAddress 方法
}
```

[在线运行程序](https://play.golang.org/p/vURnImw4_9)

在上面程序的第 32 行，我们通过使用 `p.fullAddress()` 来访问 `address` 结构体的 `fullAddress()` 方法。明确的调用 `p.address.fullAddress()` 是没有必要的。该程序输出：

```
Full address: Los Angeles, California
```

### 在方法中使用值接收器 与 在函数中使用值参数

这个话题很多Go语言新手都弄不明白。我会尽量讲清楚。

当一个函数有一个值参数，它只能接受一个值参数。

当一个方法有一个值接收器，它可以接受值接收器和指针接收器。

让我们通过一个例子来理解这一点。

```go
package main

import (
    "fmt"
)

type rectangle struct {
    length int
    width  int
}

func area(r rectangle) {
    fmt.Printf("Area Function result: %d\n", (r.length * r.width))
}

func (r rectangle) area() {
    fmt.Printf("Area Method result: %d\n", (r.length * r.width))
}

func main() {
    r := rectangle{
        length: 10,
        width:  5,
    }
    area(r)
    r.area()

    p := &r
    /*
       compilation error, cannot use p (type *rectangle) as type rectangle
       in argument to area
    */
    //area(p)

    p.area()//通过指针调用值接收器
}
```

[在线运行程序](https://play.golang.org/p/gLyHMd2iie)

第 12 行的函数 `func area(r rectangle)` 接受一个值参数，方法 `func (r rectangle) area()` 接受一个值接收器。

在第 25 行，我们通过值参数 `area(r)` 来调用 area 这个函数，这是合法的。同样，我们使用值接收器来调用 area 方法 `r.area()`，这也是合法的。

在第 28 行，我们创建了一个指向 `r` 的指针 `p`。如果我们试图把这个指针传递到只能接受一个值参数的函数 area，编译器将会报错。所以我把代码的第 33 行注释了。如果你把这行的代码注释去掉，编译器将会抛出错误 `compilation error, cannot use p (type *rectangle) as type rectangle in argument to area.`。这将会按预期抛出错误。

现在到了棘手的部分了，在第35行的代码 `p.area()` 使用指针接收器 `p` 调用了只接受一个值接收器的方法 `area`。这是完全有效的。原因是当 `area` 有一个值接收器时，为了方便Go语言把 `p.area()` 解释为 `(*p).area()`。

该程序将会输出：

```
Area Function result: 50
Area Method result: 50
Area Method result: 50
```

### 在方法中使用指针接收器 与 在函数中使用指针参数

和值参数相类似，函数使用指针参数只接受指针，而使用指针接收器的方法可以使用值接收器和指针接收器。

```go
package main

import (
    "fmt"
)

type rectangle struct {
    length int
    width  int
}

func perimeter(r *rectangle) {
    fmt.Println("perimeter function output:", 2*(r.length+r.width))

}

func (r *rectangle) perimeter() {
    fmt.Println("perimeter method output:", 2*(r.length+r.width))
}

func main() {
    r := rectangle{
        length: 10,
        width:  5,
    }
    p := &r //pointer to r
    perimeter(p)
    p.perimeter()

    /*
        cannot use r (type rectangle) as type *rectangle in argument to perimeter
    */
    //perimeter(r)

    r.perimeter()//使用值来调用指针接收器
}
```

[在线运行程序](https://play.golang.org/p/Xy5wW9YZMJ)

在上面程序的第 12 行，定义了一个接受指针参数的函数 `perimeter`。第 17 行定义了一个有一个指针接收器的方法。

在第 27 行，我们调用 perimeter 函数时传入了一个指针参数。在第 28 行，我们通过指针接收器调用了 perimeter 方法。所有一切看起来都这么完美。

在被注释掉的第 33 行，我们尝试通过传入值参数 `r` 调用函数 `perimeter`。这是不被允许的，因为函数的指针参数不接受值参数。如果你把这行的代码注释去掉并把程序运行起来，编译器将会抛出错误 `main.go:33: cannot use r (type rectangle) as type *rectangle in argument to perimeter.`。

在第 35 行，我们通过值接收器 `r` 来调用有指针接收器的方法 `perimeter`。这是被允许的，为了方便Go语言把代码 `r.perimeter()` 解释为 `(&r).perimeter()`。该程序输出：

```
perimeter function output: 30
perimeter method output: 30
perimeter method output: 30
```

### 在非结构体上的方法

到目前为止，我们只在结构体类型上定义方法。也可以在非结构体类型上定义方法，但是有一个问题。**为了在一个类型上定义一个方法，方法的接收器类型定义和方法的定义应该在同一个包中。到目前为止，我们定义的所有结构体和结构体上的方法都是在同一个 `main` 包中，因此它们是可以运行的。**

```go
package main

func (a int) add(b int) {
}

func main() {

}
```

[在线运行程序](https://play.golang.org/p/ybXLf5o_lA)

在上面程序的第 3 行，我们尝试把一个 `add` 方法添加到内置的类型 `int`。这是不允许的，因为 `add` 方法的定义和 `int` 类型的定义不在同一个包中。该程序会抛出编译错误 `cannot define new methods on non-local type int`。

让该程序工作的方法是为内置类型 int 创建一个类型别名，然后创建一个以该类型别名为接收器的方法。

```go
package main

import "fmt"

type myInt int

func (a myInt) add(b myInt) myInt {
    return a + b
}

func main() {
    num1 := myInt(5)
    num2 := myInt(10)
    sum := num1.add(num2)
    fmt.Println("Sum is", sum)
}
```

[在线运行程序](https://play.golang.org/p/sTe7i1qAng)

在上面程序的第5行，我们为 `int` 创建了一个类型别名 `myInt`。在第7行，我们定义了一个以 `myInt` 为接收器的的方法 `add`。

该程序将会打印出 `Sum is 15`。

我已经创建了一个程序，包含了我们迄今为止所讨论的所有概念，详见[github](https://github.com/golangbot/methods)。

这就是Go中的方法。祝你有美好的一天。

**上一教程 - [结构体](https://studygolang.com/articles/12263)**

**下一教程 - [接口 - I](https://studygolang.com/articles/12266)**

----------------

via: https://golangbot.com/methods/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[MDGSF](https://github.com/MDGSF)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
