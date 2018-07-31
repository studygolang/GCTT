首发于：https://studygolang.com/articles/13890

# Go 中的可变参函数

当函数最后一个参数为 *...T* 时（*T* 前面的三个点是特意的），就叫*可变参函数*：

```go
package main
import "fmt"
func sum(numbers ...float64) (res float64) {
    for _, number := range numbers {
        res += number
    }
    return
}
func main() {
    fmt.Println(sum(1.1, 2.2, 3.3))
}
```

可变参函数允许传递任意数量（可变）的实参，在函数内部通过一个 [slice](https://golang.org/ref/spec#Slice_types) 来访问这些参数的值（例如上面代码中的参数成员）。

> 只有在最后一个形参前加上 ...（三个点），才表示这是一个可变参函数。

## 实参 vs. 形参

在大多数情况下这两个术语是可以互相通用的，但是实参一般表示涉及在调用函数 / 方法时传入的参数，形参是特指在函数定义时的参数。

```go
package main
import "fmt"
func sum(a, b float64) (res float64) {
    return a + b
}
func main() {
    fmt.Println(sum(1.1, 1.2))
}
```

- a 和 b 是形参
- res 是指定名称的返回参数
- 1.1 和 1.2 是实参

## ... 转换为 slice

在可变参函数内部中，...T 类型实际上是一个 []T:

```go
package main
import "fmt"
func f(names ...string) {
    fmt.Printf("value: %#v\n", names)
    fmt.Printf("length: %d\n", len(names))
    fmt.Printf("capacity: %d\n", cap(names))
    fmt.Printf("type: %T\n", names)
}
func main() {
    f("one", "two", "three")
}
```

编译运行代码后输出:

```
> go install github.com/mlowicki/lab && ./bin/lab
value: []string{"one", "two", "three"}
type: []string
length: 3
capacity: 3
```

## 类型的一致性

可变参函数并不等同于最后一个形参为传入 slice 的函数:

```go
f := func(...int) {}
f = func([]int) {}
```

上面的代码在编译期时会报错:

```
src/github.com/mlowicki/lab/lab.go:17: cannot use func literal (type func([]int)) as type func(...int) in assignment
```

## 把三个点放到另一侧

下面的代码段无法被成功编译:

```go
package main
import "fmt"
func f(numbers ...int) {
    fmt.Println(numbers)
}
func main() {
    numbers := []int{1, 2, 3}
    f(numbers)
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
# github.com/mlowicki/lab
src/github.com/mlowicki/lab/lab.go:11: cannot use numbers (type []int) as type int in argument to f
```

这是因为传递单个实参的时候必须是 int 类型的，所以很明显传递一个 slice 类型是不被允许的。在 Golang 中有个技巧可以快速解决这个问题：

```go
package main
import "fmt"
func f(numbers ...int) {
    fmt.Println(numbers)
}
func main() {
    numbers := []int{1, 2, 3}
    f(numbers...)
}
```

在调用函数 *f* 时，三个点（...）放到了实参的**后面**，这是个容易被忽视的改动，但它将 *numbers* 作为形参列表进行传递。这个方式实现了用 slice 作为参数去调用可变参函数。

可变参数函数实际上是一种语法糖，它只是将 slice 作为最后一个参数而已。可变参函数可以为 APIs 提供更丰富的表现形式，同时也让开发者不用再创建临时 slice。

---

via: https://medium.com/golangspec/variadic-functions-in-go-13c33182b851

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[nicedevcn](https://github.com/nicedevcn)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
