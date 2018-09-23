已发布：https://studygolang.com/articles/12121

# 第 11 章：数组和切片

欢迎来到 [Golang 系列教程](https://golangbot.com/learn-golang-series/)的第 11 章。在本章教程中，我们将讨论 Go 语言中的数组和切片。

## 数组

数组是同一类型元素的集合。例如，整数集合 5,8,9,79,76 形成一个数组。Go 语言中不允许混合不同类型的元素，例如包含字符串和整数的数组。（译者注：当然，如果是 interface{} 类型数组，可以包含任意类型）

### 数组的声明

一个数组的表示形式为 `[n]T`。`n` 表示数组中元素的数量，`T` 代表每个元素的类型。元素的数量 `n` 也是该类型的一部分（稍后我们将详细讨论这一点）。

可以使用不同的方式来声明数组，让我们一个一个的来看。

```go
package main

import (
    "fmt"
)

func main() {
    var a [3]int //int array with length 3
    fmt.Println(a)
}
```
[在线运行程序](https://play.golang.org/p/Zvgh82u0ej)

**var a[3]int** 声明了一个长度为 3 的整型数组。**数组中的所有元素都被自动赋值为数组类型的零值。** 在这种情况下，`a` 是一个整型数组，因此 `a` 的所有元素都被赋值为 `0`，即 int 型的零值。运行上述程序将 **输出** `[0 0 0]`。

数组的索引从 `0` 开始到 `length - 1` 结束。让我们给上面的数组赋值。

```go
package main

import (
    "fmt"
)

func main() {
    var a [3]int //int array with length 3
    a[0] = 12 // array index starts at 0
    a[1] = 78
    a[2] = 50
    fmt.Println(a)
}
```
[在线运行程序](https://play.golang.org/p/WF0Uj8sv39)

a[0] 将值赋给数组的第一个元素。该程序将 **输出** `[12 78 50]`。

让我们使用 **简略声明** 来创建相同的数组。

```go
package main

import (
    "fmt"
)

func main() {
    a := [3]int{12, 78, 50} // short hand declaration to create array
    fmt.Println(a)
}
```
[在线运行程序](https://play.golang.org/p/NKOV04zgI6)

上面的程序将会打印相同的 **输出** `[12 78 50]`。

在简略声明中，不需要将数组中所有的元素赋值。

```go
package main

import (
    "fmt"
)

func main() {
    a := [3]int{12}
    fmt.Println(a)
}
```
[在线运行程序](https://play.golang.org/p/AdPH0kXRly)

在上述程序中的第 8 行 `a := [3]int{12}` 声明一个长度为 3 的数组，但只提供了一个值 `12`，剩下的 2 个元素自动赋值为 `0`。这个程序将**输出** `[12 0 0]`。

你甚至可以忽略声明数组的长度，并用 `...` 代替，让编译器为你自动计算长度，这在下面的程序中实现。

```go
package main

import (
    "fmt"
)

func main() {
    a := [...]int{12, 78, 50} // ... makes the compiler determine the length
    fmt.Println(a)
}
```
[在线运行程序](https://play.golang.org/p/_fVmr6KGDh)

**数组的大小是类型的一部分**。因此 `[5]int` 和 `[25]int` 是不同类型。数组不能调整大小，不要担心这个限制，因为 `slices` 的存在能解决这个问题。

```go
package main

func main() {
    a := [3]int{5, 78, 8}
    var b [5]int
    b = a // not possible since [3]int and [5]int are distinct types
}
```
[在线运行程序](https://play.golang.org/p/kBdot3pXSB)

在上述程序的第 6 行中, 我们试图将类型 `[3]int` 的变量赋给类型为 `[5]int` 的变量，这是不允许的，因此编译器将抛出错误 main.go:6: cannot use a (type [3]int) as type [5]int in assignment。

### 数组是值类型

Go 中的数组是值类型而不是引用类型。这意味着当数组赋值给一个新的变量时，该变量会得到一个原始数组的一个副本。如果对新变量进行更改，则不会影响原始数组。

```go
package main

import "fmt"

func main() {
    a := [...]string{"USA", "China", "India", "Germany", "France"}
    b := a // a copy of a is assigned to b
    b[0] = "Singapore"
    fmt.Println("a is ", a)
    fmt.Println("b is ", b)
}
```
[在线运行程序](https://play.golang.org/p/-ncGk1mqPd)

在上述程序的第 7 行，`a` 的副本被赋给 `b`。在第 8 行中，`b` 的第一个元素改为 `Singapore`。这不会在原始数组 `a` 中反映出来。该程序将 **输出**,

```
a is [USA China India Germany France]
b is [Singapore China India Germany France]
```

同样，当数组作为参数传递给函数时，它们是按值传递，而原始数组保持不变。

```go
package main

import "fmt"

func changeLocal(num [5]int) {
    num[0] = 55
    fmt.Println("inside function ", num)
}
func main() {
    num := [...]int{5, 6, 7, 8, 8}
    fmt.Println("before passing to function ", num)
    changeLocal(num) //num is passed by value
    fmt.Println("after passing to function ", num)
}
```
[在线运行程序](https://play.golang.org/p/e3U75Q8eUZ)

在上述程序的 13 行中, 数组 `num` 实际上是通过值传递给函数 `changeLocal`，数组不会因为函数调用而改变。这个程序将输出,

```
before passing to function  [5 6 7 8 8]
inside function  [55 6 7 8 8]
after passing to function  [5 6 7 8 8]
```

### 数组的长度

通过将数组作为参数传递给 `len` 函数，可以得到数组的长度。

```go
package main

import "fmt"

func main() {
    a := [...]float64{67.7, 89.8, 21, 78}
    fmt.Println("length of a is",len(a))
}
```
[在线运行程序](https://play.golang.org/p/UrIeNlS0RN)

上面的程序输出为 `length of a is 4`。

### 使用 range 迭代数组

`for` 循环可用于遍历数组中的元素。

```go
package main

import "fmt"

func main() {
    a := [...]float64{67.7, 89.8, 21, 78}
    for i := 0; i < len(a); i++ { // looping from 0 to the length of the array
        fmt.Printf("%d th element of a is %.2f\n", i, a[i])
    }
}
```
[在线运行程序](https://play.golang.org/p/80ejSTACO6)

上面的程序使用 `for` 循环遍历数组中的元素，从索引 `0` 到 `length of the array - 1`。这个程序运行后打印出，

```
0 th element of a is 67.70
1 th element of a is 89.80
2 th element of a is 21.00
3 th element of a is 78.00
```

Go 提供了一种更好、更简洁的方法，通过使用 `for` 循环的 **range** 方法来遍历数组。`range` 返回索引和该索引处的值。让我们使用 range 重写上面的代码。我们还可以获取数组中所有元素的总和。

```go
package main

import "fmt"

func main() {
    a := [...]float64{67.7, 89.8, 21, 78}
    sum := float64(0)
    for i, v := range a {//range returns both the index and value
        fmt.Printf("%d the element of a is %.2f\n", i, v)
        sum += v
    }
    fmt.Println("\nsum of all elements of a",sum)
}
```
[在线运行程序](https://play.golang.org/p/Ji6FRon36m)

上述程序的第 8 行 `for i, v := range a` 利用的是 for 循环 range 方式。 它将返回索引和该索引处的值。 我们打印这些值，并计算数组 `a` 中所有元素的总和。 程序的 **输出是**，

```
0 the element of a is 67.70
1 the element of a is 89.80
2 the element of a is 21.00
3 the element of a is 78.00

sum of all elements of a 256.5
```

如果你只需要值并希望忽略索引，则可以通过用 `_` 空白标识符替换索引来执行。

```go
for _, v := range a { // ignores index
}
```

上面的 for 循环忽略索引，同样值也可以被忽略。

### 多维数组

到目前为止我们创建的数组都是一维的，Go 语言可以创建多维数组。

```go
package main

import (
    "fmt"
)

func printarray(a [3][2]string) {
    for _, v1 := range a {
        for _, v2 := range v1 {
            fmt.Printf("%s ", v2)
        }
        fmt.Printf("\n")
    }
}

func main() {
    a := [3][2]string{
        {"lion", "tiger"},
        {"cat", "dog"},
        {"pigeon", "peacock"}, // this comma is necessary. The compiler will complain if you omit this comma
    }
    printarray(a)
    var b [3][2]string
    b[0][0] = "apple"
    b[0][1] = "samsung"
    b[1][0] = "microsoft"
    b[1][1] = "google"
    b[2][0] = "AT&T"
    b[2][1] = "T-Mobile"
    fmt.Printf("\n")
    printarray(b)
}
```
[在线运行程序](https://play.golang.org/p/InchXI4yY8)

在上述程序的第 17 行，用简略语法声明一个二维字符串数组 a 。20 行末尾的逗号是必需的。这是因为根据 Go 语言的规则自动插入分号。至于为什么这是必要的，如果你想了解更多，请阅读[https://golang.org/doc/effective_go.html#semicolons](https://golang.org/doc/effective_go.html#semicolons)。

另外一个二维数组 b 在 23 行声明，字符串通过每个索引一个一个添加。这是另一种初始化二维数组的方法。

第 7 行的 printarray 函数使用两个 range 循环来打印二维数组的内容。上述程序的 **输出是**

```
lion tiger
cat dog
pigeon peacock

apple samsung
microsoft google
AT&T T-Mobile
```

这就是数组，尽管数组看上去似乎足够灵活，但是它们具有固定长度的限制，不可能增加数组的长度。这就要用到 **切片** 了。事实上，在 Go 中，切片比传统数组更常见。

## 切片

切片是由数组建立的一种方便、灵活且功能强大的包装（Wrapper）。切片本身不拥有任何数据。它们只是对现有数组的引用。

### 创建一个切片

带有 T 类型元素的切片由 `[]T` 表示

```go
package main

import (
    "fmt"
)

func main() {
    a := [5]int{76, 77, 78, 79, 80}
    var b []int = a[1:4] // creates a slice from a[1] to a[3]
    fmt.Println(b)
}
```
[在线运行程序](https://play.golang.org/p/Za6w5eubBB)

使用语法 `a[start:end]` 创建一个从 `a` 数组索引 `start` 开始到 `end - 1` 结束的切片。因此，在上述程序的第 9 行中, `a[1:4]` 从索引 1 到 3 创建了 `a` 数组的一个切片表示。因此, 切片 `b` 的值为 `[77 78 79]`。

让我们看看另一种创建切片的方法。

```go
package main

import (
    "fmt"
)

func main() {
    c := []int{6, 7, 8} // creates and array and returns a slice reference
    fmt.Println(c)
}
```
[在线运行程序](https://play.golang.org/p/_Z97MgXavA)

在上述程序的第 9 行，`c：= [] int {6，7，8}` 创建一个有 3 个整型元素的数组，并返回一个存储在 c 中的切片引用。

### 切片的修改

切片自己不拥有任何数据。它只是底层数组的一种表示。对切片所做的任何修改都会反映在底层数组中。

```go
package main

import (
    "fmt"
)

func main() {
    darr := [...]int{57, 89, 90, 82, 100, 78, 67, 69, 59}
    dslice := darr[2:5]
    fmt.Println("array before", darr)
    for i := range dslice {
        dslice[i]++
    }
    fmt.Println("array after", darr)
}
```
[在线运行程序](https://play.golang.org/p/6FinudNf1k)

在上述程序的第 9 行，我们根据数组索引 2,3,4 创建一个切片 `dslice`。for 循环将这些索引中的值逐个递增。当我们使用 for 循环打印数组时，我们可以看到对切片的更改反映在数组中。该程序的输出是

```
array before [57 89 90 82 100 78 67 69 59]
array after [57 89 91 83 101 78 67 69 59]
```

当多个切片共用相同的底层数组时，每个切片所做的更改将反映在数组中。

```go
package main

import (
    "fmt"
)

func main() {
    numa := [3]int{78, 79 ,80}
    nums1 := numa[:] // creates a slice which contains all elements of the array
    nums2 := numa[:]
    fmt.Println("array before change 1", numa)
    nums1[0] = 100
    fmt.Println("array after modification to slice nums1", numa)
    nums2[1] = 101
    fmt.Println("array after modification to slice nums2", numa)
}
```
[在线运行程序](https://play.golang.org/p/mdNi4cs854)

在 9 行中，`numa [:]` 缺少开始和结束值。开始和结束的默认值分别为 `0` 和 `len (numa)`。两个切片 `nums1` 和 `nums2` 共享相同的数组。该程序的输出是

```
array before change 1 [78 79 80]
array after modification to slice nums1 [100 79 80]
array after modification to slice nums2 [100 101 80]
```

从输出中可以清楚地看出，当切片共享同一个数组时，每个所做的修改都会反映在数组中。

### 切片的长度和容量

切片的长度是切片中的元素数。**切片的容量是从创建切片索引开始的底层数组中元素数。**

让我们写一段代码来更好地理解这点。

```go
package main

import (
    "fmt"
)

func main() {
    fruitarray := [...]string{"apple", "orange", "grape", "mango", "water melon", "pine apple", "chikoo"}
    fruitslice := fruitarray[1:3]
    fmt.Printf("length of slice %d capacity %d", len(fruitslice), cap(fruitslice)) // length of is 2 and capacity is 6
}
```
[在线运行程序](https://play.golang.org/p/a1WOcdv827)

在上面的程序中，`fruitslice` 是从 `fruitarray` 的索引 1 和 2 创建的。 因此，`fruitlice` 的长度为 `2`。

`fruitarray` 的长度是 7。`fruiteslice` 是从 `fruitarray` 的索引 `1` 创建的。因此, `fruitslice` 的容量是从 `fruitarray` 索引为 `1` 开始，也就是说从 `orange` 开始，该值是 `6`。因此, `fruitslice` 的容量为 6。该[程序](https://play.golang.org/p/a1WOcdv827)输出切片的 **长度为 2 容量为 6 **。

切片可以重置其容量。任何超出这一点将导致程序运行时抛出错误。

```go
package main

import (
    "fmt"
)

func main() {
    fruitarray := [...]string{"apple", "orange", "grape", "mango", "water melon", "pine apple", "chikoo"}
    fruitslice := fruitarray[1:3]
    fmt.Printf("length of slice %d capacity %d\n", len(fruitslice), cap(fruitslice)) // length of is 2 and capacity is 6
    fruitslice = fruitslice[:cap(fruitslice)] // re-slicing furitslice till its capacity
    fmt.Println("After re-slicing length is",len(fruitslice), "and capacity is",cap(fruitslice))
}
```
[在线运行程序](https://play.golang.org/p/GcNzOOGicu)

在上述程序的第 11 行中，`fruitslice` 的容量是重置的。以上程序输出为，

```
length of slice 2 capacity 6
After re-slicing length is 6 and capacity is 6
```

### 使用 make 创建一个切片

func make（[]T，len，cap）[]T 通过传递类型，长度和容量来创建切片。容量是可选参数, 默认值为切片长度。make 函数创建一个数组，并返回引用该数组的切片。

```go
package main

import (
    "fmt"
)

func main() {
    i := make([]int, 5, 5)
    fmt.Println(i)
}
```
[在线运行程序](https://play.golang.org/p/M4OqxzerxN)

使用 make 创建切片时默认情况下这些值为零。上述程序的输出为 `[0 0 0 0 0]`。

### 追加切片元素

正如我们已经知道数组的长度是固定的，它的长度不能增加。 切片是动态的，使用 `append` 可以将新元素追加到切片上。append 函数的定义是 `func append（s[]T，x ... T）[]T`。

**x ... T** 在函数定义中表示该函数接受参数 x 的个数是可变的。这些类型的函数被称为[可变函数](https://golangbot.com/variadic-functions/)。

有一个问题可能会困扰你。如果切片由数组支持，并且数组本身的长度是固定的，那么切片如何具有动态长度。以及内部发生了什么，当新的元素被添加到切片时，会创建一个新的数组。现有数组的元素被复制到这个新数组中，并返回这个新数组的新切片引用。现在新切片的容量是旧切片的两倍。很酷吧 :)。下面的程序会让你清晰理解。

```go
package main

import (
    "fmt"
)

func main() {
    cars := []string{"Ferrari", "Honda", "Ford"}
    fmt.Println("cars:", cars, "has old length", len(cars), "and capacity", cap(cars)) // capacity of cars is 3
    cars = append(cars, "Toyota")
    fmt.Println("cars:", cars, "has new length", len(cars), "and capacity", cap(cars)) // capacity of cars is doubled to 6
}
```
[在线运行程序](https://play.golang.org/p/VUSXCOs1CF)

在上述程序中，`cars` 的容量最初是 3。在第 10 行，我们给 cars 添加了一个新的元素，并把 `append(cars, "Toyota")` 返回的切片赋值给 cars。现在 cars 的容量翻了一番，变成了 6。上述程序的输出是

```
cars: [Ferrari Honda Ford] has old length 3 and capacity 3
cars: [Ferrari Honda Ford Toyota] has new length 4 and capacity 6
```

切片类型的零值为 `nil`。一个 `nil` 切片的长度和容量为 0。可以使用 append 函数将值追加到 `nil` 切片。

```go
package main

import (
    "fmt"
)

func main() {
    var names []string //zero value of a slice is nil
    if names == nil {
        fmt.Println("slice is nil going to append")
        names = append(names, "John", "Sebastian", "Vinay")
        fmt.Println("names contents:",names)
    }
}
```
[在线运行程序](https://play.golang.org/p/x_-4XAJHbM)

在上面的程序 `names` 是 nil，我们已经添加 3 个字符串给 `names`。该程序的输出是

```
slice is nil going to append
names contents: [John Sebastian Vinay]
```

也可以使用 `...` 运算符将一个切片添加到另一个切片。 你可以在[可变参数函数](https://golangbot.com/variadic-functions/)教程中了解有关此运算符的更多信息。

```go
package main

import (
    "fmt"
)

func main() {
    veggies := []string{"potatoes", "tomatoes", "brinjal"}
    fruits := []string{"oranges", "apples"}
    food := append(veggies, fruits...)
    fmt.Println("food:",food)
}
```
[在线运行程序](https://play.golang.org/p/UnHOH_u6HS)

在上述程序的第 10 行，food 是通过 append(veggies, fruits...) 创建。程序的输出为 `food: [potatoes tomatoes brinjal oranges apples]`。

### 切片的函数传递

我们可以认为，切片在内部可由一个结构体类型表示。这是它的表现形式，

```go
type slice struct {
    Length        int
    Capacity      int
    ZerothElement *byte
}
```

切片包含长度、容量和指向数组第零个元素的指针。当切片传递给函数时，即使它通过值传递，指针变量也将引用相同的底层数组。因此，当切片作为参数传递给函数时，函数内所做的更改也会在函数外可见。让我们写一个程序来检查这点。

```go
package main

import (
    "fmt"
)

func subtactOne(numbers []int) {
    for i := range numbers {
        numbers[i] -= 2
    }
}
func main() {
    nos := []int{8, 7, 6}
    fmt.Println("slice before function call", nos)
    subtactOne(nos)                               // function modifies the slice
    fmt.Println("slice after function call", nos) // modifications are visible outside
}
```
[在线运行程序](https://play.golang.org/p/IzqDihNifq)

上述程序的行号 17 中，调用函数将切片中的每个元素递减 2。在函数调用后打印切片时，这些更改是可见的。如果你还记得，这是不同于数组的，对于函数中一个数组的变化在函数外是不可见的。上述[程序](https://play.golang.org/p/bWUb6R-1bS)的输出是，

```
array before function call [8 7 6]
array after function call [6 5 4]
```

### 多维切片

类似于数组，切片可以有多个维度。

```go
package main

import (
    "fmt"
)

func main() {
     pls := [][]string {
            {"C", "C++"},
            {"JavaScript"},
            {"Go", "Rust"},
            }
    for _, v1 := range pls {
        for _, v2 := range v1 {
            fmt.Printf("%s ", v2)
        }
        fmt.Printf("\n")
    }
}
```
[在线运行程序](https://play.golang.org/p/--p1AvNGwN)

程序的输出为，

```
C C++
JavaScript
Go Rust
```

### 内存优化

切片持有对底层数组的引用。只要切片在内存中，数组就不能被垃圾回收。在内存管理方面，这是需要注意的。让我们假设我们有一个非常大的数组，我们只想处理它的一小部分。然后，我们由这个数组创建一个切片，并开始处理切片。这里需要重点注意的是，在切片引用时数组仍然存在内存中。

一种解决方法是使用 [copy](https://golang.org/pkg/builtin/#copy) 函数 `func copy(dst，src[]T)int` 来生成一个切片的副本。这样我们可以使用新的切片，原始数组可以被垃圾回收。

```go
package main

import (
    "fmt"
)

func countries() []string {
    countries := []string{"USA", "Singapore", "Germany", "India", "Australia"}
    neededCountries := countries[:len(countries)-2]
    countriesCpy := make([]string, len(neededCountries))
    copy(countriesCpy, neededCountries) //copies neededCountries to countriesCpy
    return countriesCpy
}
func main() {
    countriesNeeded := countries()
    fmt.Println(countriesNeeded)
}
```
[在线运行程序](https://play.golang.org/p/35ayYBhcDE)

在上述程序的第 9 行，`neededCountries := countries[:len(countries)-2` 创建一个去掉尾部 2 个元素的切片 `countries`，在上述程序的 11 行，将 `neededCountries` 复制到 `countriesCpy` 同时在函数的下一行返回 countriesCpy。现在 `countries` 数组可以被垃圾回收, 因为 `neededCountries` 不再被引用。

我已经把我们迄今为止所讨论的所有概念整理成一个程序。 你可以从 [github](https://github.com/golangbot/arraysandslices) 下载它。

这是数组和切片。谢谢你的阅读。请您留下宝贵的意见和意见。

**上一教程 - [switch 语句](https://studygolang.com/articles/11957)**

**下一教程 - [可变函数](https://studygolang.com/articles/12173)**

---

via: https://golangbot.com/arrays-and-slices/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Dingo1991](https://github.com/Dingo1991)
校对：[Noluye](https://github.com/Noluye) [polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
