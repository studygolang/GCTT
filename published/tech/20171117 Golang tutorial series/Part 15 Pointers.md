已发布：https://studygolang.com/articles/12262

# 第 15 部分：指针

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 15 个教程。

### 什么是指针？

指针是一种存储变量内存地址（Memory Address）的变量。

![指针示意图](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-series/pointer-explained.png "指针示意图")
  
如上图所示，变量 `b` 的值为 `156`，而 `b` 的内存地址为 `0x1040a124`。变量 `a` 存储了 `b` 的地址。我们就称 `a` 指向了 `b`。

### 指针的声明

指针变量的类型为 **`*T`**，该指针指向一个 **T** 类型的变量。

接下来我们写点代码。

```go
package main

import (
    "fmt"
)

func main() {
    b := 255
    var a *int = &b
    fmt.Printf("Type of a is %T\n", a)
    fmt.Println("address of b is", a)
}
```
[在线运行程序](https://play.golang.org/p/A4vmlgxAy8)

**&** 操作符用于获取变量的地址。上面程序的第 9 行我们把 `b` 的地址赋值给 **`*int`** 类型的 `a`。我们称 `a` 指向了 `b`。当我们打印 `a` 的值时，会打印出 `b` 的地址。程序将输出：

```
Type of a is *int  
address of b is 0x1040a124  
```

由于 b 可能处于内存的任何位置，你应该会得到一个不同的地址。

### 指针的零值（Zero Value）

指针的零值是 `nil`。

```go
package main

import (  
    "fmt"
)

func main() {  
    a := 25
    var b *int
    if b == nil {
        fmt.Println("b is", b)
        b = &a
        fmt.Println("b after initialization is", b)
    }
}
```
[在线运行程序](https://play.golang.org/p/yAeGhzgQE1)

上面的程序中，`b` 初始化为 `nil`，接着将 `a` 的地址赋值给 `b`。程序会输出：

```
b is <nil>  
b after initialisation is 0x1040a124 
```

### 指针的解引用

指针的解引用可以获取指针所指向的变量的值。将 `a` 解引用的语法是 `*a`。

通过下面的代码，可以看到如何使用解引用。

```go
package main  
import (  
    "fmt"
)

func main() {  
    b := 255
    a := &b
    fmt.Println("address of b is", a)
    fmt.Println("value of b is", *a)
}
```
[在线运行程序](https://play.golang.org/p/m5pNbgFwbM)

在上面程序的第 10 行，我们将 `a` 解引用，并打印了它的值。不出所料，我们会打印出 `b` 的值。程序会输出：

```
address of b is 0x1040a124  
value of b is 255 
```

我们再编写一个程序，用指针来修改 b 的值。

```go
package main

import (  
    "fmt"
)

func main() {  
    b := 255
    a := &b
    fmt.Println("address of b is", a)
    fmt.Println("value of b is", *a)
    *a++
    fmt.Println("new value of b is", b)
}
```
[在线运行程序](https://play.golang.org/p/cdmvlpBNmb)

在上面程序的第 12 行中，我们把 `a` 指向的值加 1，由于 `a` 指向了 `b`，因此 `b` 的值也发生了同样的改变。于是 `b` 的值变为 256。程序会输出：

```
address of b is 0x1040a124  
value of b is 255  
new value of b is 256  
```

### 向函数传递指针参数

```go
package main

import (  
    "fmt"
)

func change(val *int) {  
    *val = 55
}
func main() {  
    a := 58
    fmt.Println("value of a before function call is",a)
    b := &a
    change(b)
    fmt.Println("value of a after function call is", a)
}
```
[在线运行程序](https://play.golang.org/p/3n2nHRJJqn)

在上面程序中的第 14 行，我们向函数 `change` 传递了指针变量 `b`，而 `b` 存储了 `a` 的地址。程序的第 8 行在 `change` 函数内使用解引用，修改了 a 的值。该程序会输出：

```
value of a before function call is 58  
value of a after function call is 55  
```

### 不要向函数传递数组的指针，而应该使用切片

假如我们想要在函数内修改一个数组，并希望调用函数的地方也能得到修改后的数组，一种解决方案是把一个指向数组的指针传递给这个函数。

```go
package main

import (  
    "fmt"
)

func modify(arr *[3]int) {  
    (*arr)[0] = 90
}

func main() {  
    a := [3]int{89, 90, 91}
    modify(&a)
    fmt.Println(a)
}
```
[在线运行程序](https://play.golang.org/p/lOIznCbcvs)

在上面程序的第 13 行中，我们将数组的地址传递给了 `modify` 函数。在第 8 行，我们在 `modify` 函数里把 `arr` 解引用，并将 `90` 赋值给这个数组的第一个元素。程序会输出 `[90 90 91]`。

**`a[x]` 是 `(*a)[x]` 的简写形式，因此上面代码中的 `(*arr)[0]` 可以替换为 `arr[0]`**。下面我们用简写形式重写以上代码。

```go
package main

import (  
    "fmt"
)

func modify(arr *[3]int) {  
    arr[0] = 90
}

func main() {  
    a := [3]int{89, 90, 91}
    modify(&a)
    fmt.Println(a)
}
```
[在线运行程序](https://play.golang.org/p/k7YR0EUE1G)

该程序也会输出 `[90 90 91]`。

**这种方式向函数传递一个数组指针参数，并在函数内修改数组。尽管它是有效的，但却不是 Go 语言惯用的实现方式。我们最好使用[切片](https://golangbot.com/arrays-and-slices/)来处理。**

接下来我们用[切片](https://golangbot.com/arrays-and-slices/)来重写之前的代码。

```go
package main

import (  
    "fmt"
)

func modify(sls []int) {  
    sls[0] = 90
}

func main() {  
    a := [3]int{89, 90, 91}
    modify(a[:])
    fmt.Println(a)
}
```
[在线运行程序](https://play.golang.org/p/rRvbvuI67W)

在上面程序的第 13 行，我们将一个切片传递给了 `modify` 函数。在 `modify` 函数中，我们把切片的第一个元素修改为 `90`。程序也会输出 `[90 90 91]`。**所以别再传递数组指针了，而是使用切片吧**。上面的代码更加简洁，也更符合 Go 语言的习惯。

### Go 不支持指针运算

Go 并不支持其他语言（例如 C）中的指针运算。

```go
package main

func main() {  
    b := [...]int{109, 110, 111}
    p := &b
    p++
}
```
[在线运行程序](https://play.golang.org/p/WRaj4pkqRD)

上面的程序会抛出编译错误：**`main.go:6: invalid operation: p++ (non-numeric type *[3]int)`**。

我在 [github](https://github.com/golangbot/pointers) 上创建了一个程序，涵盖了所有我们讨论过的内容。

关于指针的介绍到此结束。祝您愉快。

**下一教程 - 结构体**

---

via: https://golangbot.com/pointers/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
