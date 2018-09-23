首发于：https://studygolang.com/articles/13865

# Go 中的 init 函数

main 标识符是随处可见的，每一个 Go 程序都是从一个叫 main 的包中的 main 函数开始的，当 main 函数返回时，程序执行结束。 init 函数也扮演着特殊的角色，接下来我们将描述下 init 函数的属性并介绍下怎么使用它们。

init 函数在包级别被定义，主要用于：

- 初始化那些不能被初始化表达式完成初始化的变量
- 检查或者修复程序的状态
- 注册
- 仅执行一次的计算
- 更多其它场合

除了下面将要讨论到的一些差异外，你还可以在正则函数中放置任何[有效](https://golang.org/ref/spec#FunctionBody)的内容。

## 包的初始化

要想使用导入的包首先需要初始化它，这是由golang的运行系统完成的，主要包括(顺序很重要)：

1. 初始化导入的包（递归的定义）
2. 在包级别为声明的变量计算并分配初始值
3. 执行包内的 init 函数（下面的空白标识符就是一个例子）

> 不管包被导入多少次，都只会被初始化一次。

## 顺序

Go 的包中有很多的文件，如果变量和函数在包的多个文件当中，那么变量的初始化和 init 函数的调用顺序又是什么样的呢？首先，初始化依赖机制会启动（更多 [Go 中的初始化依赖](https://studygolang.com/articles/13158)）当初始化依赖机制完成的时候，就需要决定 `a.go` 和 `z.go` 中的初始化变量谁会被更早的处理，而这要取决于呈现给编译器的文件顺序。如果 `z.go` 先被传递到构建系统，那么变量的初始化就会比在 `a.go` 中先一步完成，这也同样适用于 init 函数的触发。Go 语言规范建议我们始终使用相同的顺序传递，即按照词法顺序传递包中的文件名：

> 为了保证可重复的初始化行为，构建系统鼓励按照词法文件名的顺序将属于同一个包中的多个文件呈现给编译器。

但依赖特定顺序对于不关注可移植性的程序是一种方式（译注：但不建议依赖 init 初始化顺序）。让我们来看一个例子，看看它们是如何一起工作的：

### sandbox.go

```go
package main

import "fmt"

var _ int64 = s()

func init() {
    fmt.Println("init in sandbox.go")
}

func s() int64 {
    fmt.Println("calling s() in sandbox.go")
    return 1
}

func main() {
    fmt.Println("main")
}
```

### a.go

```go
package main

import "fmt"

var _ int64 = a()

func init() {
    fmt.Println("init in a.go")
}

func a() int64 {
    fmt.Println("calling a() in a.go")
    return 2
}
```

### z.go

```go
package main

import "fmt"

var _ int64 = z()

func init() {
    fmt.Println("init in z.go")
}

func z() int64 {
    fmt.Println("calling z() in z.go")
    return 3
}
```

### 程序输出:

```
calling a() in a.go
calling s() in sandbox.go
calling z() in z.go
init in a.go
init in sandbox.go
init in z.go
main
```

## 属性

init 函数不需要参数并且也不返回任何值，与 main 类似，标识符 init 没有被声明所以也就不能被引用：

```go
package main

import "fmt"

func init() {
    fmt.Println("init")
}

func main() {
    init()
}
```

在编译时这里会给出一个 “undefined：init” 错误。（ init 函数不能被引用）

> 正式的来讲 init 标示符不会引入绑定，就像空白标示符('_')表现的一样。

在同一个包或者文件当中可以定义很多的 init 函数：

### sandbox.go

```go
package main

import "fmt"

func init() {
    fmt.Println("init 1")
}

func init() {
    fmt.Println("init 2")
}

func main() {
    fmt.Println("main")
}
```

### utils.go

```go
package main

import "fmt"

func init() {
    fmt.Println("init 3")
}
```

### 输出：

```
init 1
init 2
init 3
main
```

init 函数在标准库中也被频繁的使用，例如：[*main*](https://github.com/golang/go/blob/2878cf14f3bb4c097771e50a481fec43962d7401/src/math/pow10.go#L33), [*bzip2*](https://github.com/golang/go/blob/2878cf14f3bb4c097771e50a481fec43962d7401/src/compress/bzip2/bzip2.go#L479)还有 [*image*](https://github.com/golang/go/blob/2d573eee8ae532a3720ef4efbff9c8f42b6e8217/src/image/gif/reader.go#L511)包。

...

init 函数最常见的用法就是为初始化表达式中不能被计算的那部分分配一个值：

```go
var precomputed = [20]float64{}

func init() {
    var current float64 = 1
    precomputed[0] = current
    for i := 1; i < len(precomputed); i++ {
        precomputed[i] = precomputed[i-1] * 1.2
    }
}
```

使用 for 循环作为[*表达式*](https://golang.org/ref/spec#Expression)（Go 语言中的语句）是不可能的，所以将这些放到 init 函数中就能够很好的解决这些问题。

## 仅仅为了使用包的副作用（包的初始化）而导入包

Go 语言对没有使用的导入是非常严格的。有时候程序员导入一个包可能只想要使用 init 函数的功能，例如一些引导工作。空白标示符就是一个不错的方式：

```go
import _ "image/png"
```

它甚至在[*image*](https://github.com/golang/go/blob/0104a31b8fbcbe52728a08867b26415d282c35d2/src/image/image.go#L10)包中被提到。

如果上面的内容对你有所帮助请跟随我一起续写未来的故事吧，那也将成为我的动力。

---

## 参考资料

- [The Go Programming Language Specification - The Go Programming Language](https://golang.org/ref/spec#Package_initialization)
- [Blocks in Go](https://studygolang.com/articles/12632)
- [Initialization dependencies in Go](https://studygolang.com/articles/13158)

---

via: https://medium.com/golangspec/init-functions-in-go-eac191b3860a

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[flexiwind](https://github.com/flexiwind)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
