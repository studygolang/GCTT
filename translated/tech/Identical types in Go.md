# Go中的等价类型
![title](_v_images/_title_1519478254_24382.png)

[分配](https://medium.com/@mlowicki/assignability-in-go-27805bcd5874)的第一个例子如下：

    如果右边的类型和T等价，那么赋值是完全可以(进行)的。

这可能听起来不值一提，但是有一些细节(值得注意)。深入这个主题可能同样有助于理解语言其它相关的基础概念。

### 类型声明

关键字type的特定描述可以创建新的类型名称：
```
type A struct{ name string }
type B A

type (
    C string
    D map[string]int
)
```
### 基础类型

Go语言中的每个类型都有部分被称为基础类型。普通块包含一些预先声明的标识符绑定到类型boolean、string或者numeric等。对于每个预先声明的类型T，它的类型就是T(此处没有陷阱)。对于类型常量就是这样的：
```
// Sample type literals

var (
    a [10]int
    b struct{ name string }
    c *int
    d func(p int) (r int)
    e interface {
        f(int) int
    }
    f []int
    g map[string]int
    h chan<- string
)
```

    类型声明可以被“因式分解”成块来避免重复关键字var多次。相同的方法可以被用于类型声明，就像第一段代码片一样。

在其它情况下T的基础类型是通过类型声明绑定到类型的基础类型：

```
type X string  // underlying type of X is string
type Y X       // underlying type of Y is string
type Z [10]int // underlying type of Z is [10]int
```

### 命名/无名类型

命名类型是通过类型名指定的新类型，可以使用包名做前缀。包名用于访问从其它包导出的命名。(与之前合适的导入语句)

```
package main

import “fmt”

type T fmt.Formatter // T and fmt.Formatter are named types
```

限制性标识符(带有包名前缀的)不能引用本包：

```
package foo

type A struct{ name string }

type B foo.A // compiler throws "undefined: foo in foo.A"
```

未命名类型使用类型常量像f.ex那样引用到自身：

    map[string]int
    chan<- int
    []float32

### Type identity类型标识

有了一些基础概念后，理解Go语言中的两个类型何时相同或者不同就很容易了：

    当通过同样的类型声明创建时，两个命名类型就是等效的：

```
type (
    T1 string
    T2 string
)
```

T1和T1是等效的。由于使用了两个分离的类型声明，T1和T2是不同的(即使因式分解到一个块中)。

2.命名和未命名的类型是不同的(没有例外)
3. 如果对应的类型literals是相同的(类型literals的细节描述在[language spec](https://golang.org/ref/spec#Type_identity)中)，那么未命名类型是等效的。
----------

via: https://medium.com/golangspec/identical-types-in-go-9cb89b91fe25

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[leemeans](https://github.com/leemeans)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出