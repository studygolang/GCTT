Learn Go by writing tests: Structs, methods, interfaces & table driven tests 

WIP: Work in progress

这是"测试驱动Go语言学习"项目的第三个帖子。"测试驱动 Go 语言学习"是一个正在进行中的项目。这个项目的目的是熟悉 Go 语言并学习 TDD 技术。
* 测试驱动 Go 语言学习之起步
* 测试驱动 Go 学习之数组和切片

## 结构体，方法和接口
假设我们需要编程计算一个给定高和宽的长方形的周长。我们可以写一个函数如下：
Premeter(width float64, height float64)
其中　float64 是形如　123.45 的浮点数。

现在我们应该很熟悉 TDD 的方式了。
## 先写测试函数
>func TestPerimeter(t *testing.T) {
>    got := Perimeter(10.0, 10.0)
>    want := 40.0
>
>    if got != want {
>        t.Errorf("got %.2f want %.2f", got, want)
>    }
>}

注意到新的格式化字符串了吗？　这里的 "f" 对应 float64, ".2" 表示输出２位小数。

## 运行测试
> ./shapes_test.go:6:9: undefined: Perimeter

## 为运行测试函数编写最少的代码并检查失败时的输出
>func Perimeter(width float64, height float64) float64 {
>  return 0
>}

运行结果是：
> shapes_test.go:10: got 0 want 40

## 编写正确的代码让测试函数通过
>func Perimeter(width float64, height float64) float64 {
>  return 2*(width + height)
>}

到目前为止还很简单。现在让我们来编写一个函数 Area(width, height float64) 来返回长方形的面积。

你可以先自己按照 TDD 的方式尝试一下。
相应的测试函数如下所示：
>func TestPerimeter(t *testing.T) {
>    got := Perimeter(10.0, 10.0)
>    want := 40.0
>
>    if got != want {
>        t.Errorf("got %.2f want %.2f", got, want)
>    }
>}
>
>func TestArea(t *testing.T) {
>    got := Area(12.0, 6.0)
>    want := 72.0
>
>    if got != want {
>        t.Errorf("got %.2f want %.2f", got, want)
>    }
>}
相应的代码如下：
>func Perimeter(width float64, height float64) float64 {
>    return 2 * (width + height)
>}
>
>func Area(width float64, height float64) float64 {
>    return width * height
>}


## 重构

我们的代码能正常工作，但是其中不包含任何显式的信息表示计算的是长方形。粗心的开发者可能会错误的调用这些函数来计算三角形的周长和面积而没有意识到错误的结果。
我们可以仅仅给这些函数命名成像 RectangleArea 一样更具体的名字。但是更简洁的方案是定义我们自己的类型 Rectangle，它可以封装长方形的信息。
我们可以使用保留字 struct 来定义自己的类型。一个通过 struct 定义出来的类型是一些已命名的域的集合，这些域用来保存数据。
一个 struct 的声明如下：
>type Rectangle struct {
>    Width float64
>    Height float64
>}

现在让我们用类型 Rectangle 代替简单的 float64 来重构这些测试函数。
>func TestPerimeter(t *testing.T) {
>    rectangle := Rectangle{10.0, 10.0}
>    got := Perimeter(rectangle)
>    want := 40.0
>
>    if got != want {
>        t.Errorf("got %.2f want %.2f", got, want)
>    }
>}
>
>func TestArea(t *testing.T) {
>    rectangle := Rectangle{12.0, 6.0}
>    got := Area(rectangle)
>    want := 72.0
>
>    if got != want {
>        t.Errorf("got %.2f want %.2f", got, want)
>    }
>}

记住先运行这些测试函数再尝试修复问题，因为运行后我们能获得有用的错误信息：
>./shapes_test.go:7:18: not enough arguments in call to Perimeter
>    have (Rectangle)
>    want (float64, float64)

我们可以通过下面的语法来访问一个 struct 中的域：
> myStruct.field

代码需要调整如下：
>func Perimeter(rectangle Rectangle) float64 {
>    return 2 * (rectangle.Width + rectangle.Height)
>}
>
>func Area(rectangle Rectangle) float64 {
>    return rectangle.Width * rectangle.Height
>}


我希望你同意通过传递一个类型为 Rectangle 的参数给这些函数更能表达我们的用意。并且这样做我们将会看到有更多的好处。

我们的下一个需求是为圆形写一个类似的函数。

## 先写测试函数
>func TestArea(t *testing.T) {
>
>    t.Run("rectangles", func(t *testing.T) {
>        rectangle := Rectangle{12, 6}
>        got := Area(rectangle)
>        want := 72.0
>
>        if got != want {
>            t.Errorf("got %.2f want %.2f", got, want)
>        }
>    })
>
>    t.Run("circles", func(t *testing.T) {
>        circle := Circle{10}
>        got := Area(circle)
>        want := 314.16
>
>        if got != want {
>            t.Errorf("got %.2f want %.2f", got, want)
>        }
>    })
>
>}

## 运行测试
> ./shapes_test.go:28:13: undefined: Circle

## 为运行测试函数编写最少的代码并检查失败时的输出
我们需要定义一个 Circle 类型：
>type Circle struct {
>  Radius float64
>}
现在我们重新运行测试：
> ./shapes_test.go:29:14: cannot use circle (type Circle) as type Rectangle in argument to Area

有些编程语言中我们可以这样做：
>func Area(circle Circle) float64 { ... }
>func Area(rectangle Rectangle) float64 { ... }


但是在 Go 语言中你不能这么做：
> ./shapes.go:20:32: Area redeclared in this block

我们有以下两个选择：
* 不同的包可以有函数名相同的函数。所以我们可以在一个新的包里创建函数 Area(Circle)。但是感觉有点大才小用了
* 我们可以为新类型定义方法。

## 什么是方法？
到目前为止我们只编写过函数但是我们已经使用过方法。当我们调用 t.Errorf 时我们调用了 t(testing.T) 这个实例的方法 ErrorF.

方法和函数很相似但是方法是通过一个特定类型的实例调用的。函数可以随时被调用，比如 Area(rectangle)。不像方法需要在某个事物上调用。

示例会帮助我们理解。让我们通过方法调用的方式来改写测试函数并尝试修复代码。

>func TestArea(t *testing.T) {
>
>    t.Run("rectangles", func(t *testing.T) {
>        rectangle := Rectangle{12, 6}
>        got := rectangle.Area()
>        want := 72.0
>
>        if got != want {
>            t.Errorf("got %.2f want %.2f", got, want)
>        }
>    })
>
>    t.Run("circles", func(t *testing.T) {
>        circle := Circle{10}
>        got := circle.Area()
>        want := 314.1592653589793
>
>        if got != want {
>            t.Errorf("got %f want %f", got, want)
>        }
>    })
>
>}

尝试运行测试函数，我们会得到如下结果：
>./shapes_test.go:19:19: rectangle.Area undefined (type Rectangle has no field or method Area)
>./shapes_test.go:29:16: circle.Area undefined (type Circle has no field or method Area)

编译器给出的信息是 "type Circle has no field or method Area".

大家可以看到编译器的伟大之处。花些时间慢慢阅读这个错误信息是很重要的，这种习惯将对你长期有用。

## 为运行测试函数编写最少的代码并检查失败时的输出

我们给这些类型加一些方法:

>type Rectangle struct {
>    Width  float64
>    Height float64
>}
>
>func (r Rectangle) Area() float64  {
>    return 0
>}
>
>type Circle struct {
>    Radius float64
>}
>
>func (c Circle) Area() float64  {
>    return 0
>}

声明方法的语法跟函数差不多，因为他们本身就很相似。唯一的不同是方法接收者的语法　func(receiverName ReceiverType) MethodName(args).

当方法被这种类型的变量调用时，　数据的引用通过变量 receiverName 获得。在其他许多编程语言中这些被隐藏起来并且通过 this 来获得接收者。

把类型的第一个字母作为接收者变量是 Go 语言的一个惯例。

现在尝试重新运行测试，编译通过了但是会有一些错误输出。

## 编写足够的代码让测试函数通过
现在让我们修改我们的新方法以让矩形测试通过：
>func (r Rectangle) Area() float64  {
>    return r.Width * r.Height
>}

现在重跑测试，矩形测试应该通过了但是圆的测试还是失败的。

为了圆的测试通过我们需要从 math 包中借用常数 Pi(记得引入 math 包)。

>func (c Circle) Area() float64  {
>    return math.Pi * c.Radius * c.Radius
>}

## 重构

我们的测试有些重复。
我们想做的是给定一些几何形状，调用 Area() 方法并检查结果。
我们想写一个这样的函数 CheckArea, 其参数是任何类型的几何形状。如果参数不是几何形状的类型，那么编译应该报错。
Go 语言中我们可以通过接口实现这一目的。
接口在 Go 这种静态类型语言中是一种非常强有力的概念。因为接口可以让函数接受不同类型的参数并能创造类型安全且高解耦的代码。

让我们引入接口来重构我们的测试代码：
>func TestArea(t *testing.T) {
>
>    checkArea := func(t *testing.T, shape Shape, want float64) {
>        t.Helper()
>        got := shape.Area()
>        if got != want {
>            t.Errorf("got %.2f want %.2f", got, want)
>        }
>    }
>
>    t.Run("rectangles", func(t *testing.T) {
>        rectangle := Rectangle{12, 6}
>        checkArea(t, rectangle, 72.0)
>    })
>
>    t.Run("circles", func(t *testing.T) {
>        circle := Circle{10}
>        checkArea(t, circle, 314.1592653589793)
>    })
>
>}

像其他练习一样我们创建了一个辅助函数，但不同的是我们传入了一个 Shape 类型。如何没有定义 Shape 类型编译会报错。

怎样定义 Shape 类型呢？我们用一个 Go 语言的接口定义来声明 Shape 类型：
>type Shape interface {
>    Area() float64
>}

这样我们就像创建 Rectangle 和 Circle 一样创建了一个新类型，不过这次是 interface 而不是 struct 。

加了这个代码后测试运行通过了。

## 稍等，什么？
这种定义 interface 的方式与大部分其他编程语言不同。通常接口定义需要这样的代码 My type Foo implements interface Bar。






---------

via:https://dev.to/quii/learn-go-by-writing-tests-structs-methods-interfaces--table-driven-tests-1p01

作者：[Chris James](https://dev.to/quii)
译者：[译者ID](https://github.com/译者ID)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
