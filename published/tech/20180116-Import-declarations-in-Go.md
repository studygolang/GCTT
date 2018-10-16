首发于：https://studygolang.com/articles/15136

# Go 中的 import 声明

Go 中的程序由各种包组成。通常，包依赖于其它包，这些包内置于标准库或者第三方。包首先需要被导入才能使用包中的导出标识符。这是通过结构体调用 *import 声明* 来实现的：

```go
package main

import (
    "fmt"
    "math"
)

func main() {
    fmt.Println(math.Exp2(10))  // 1024
}
```

上面我们有一个 import 导入的例子，其中包含了两行导入声明。每行声明定义了单个包的导入。

> 命名为 **main** 的包，是用来创建可执行的二进制文件。程序的执行是从包 **main** 开始，通过调用包中也叫做 **main** 的函数开始。

但是，还有其它一些鲜为人知的导入用法，这些用法在各种场景下都很实用：

```go
import (
    "math"
    m "math"
    . "math"
    _ "math"
)

```

这四个导入格式都有各自不同的行为，在这篇文章中我们将分析这些差异。

> 导入包只能引用导入包中的导出标识符。 导出标识符是以Unicode大写字母开头的
> - [https://golang.org/ref/spec#Exported_identifiers](https://golang.org/ref/spec#Exported_identifiers)。

## 基础

### Import 声明剖析

```
ImportDeclaration = "import" ImportSpec
ImportSpec        = [ "." | "_" | Identifier ] ImportPath
```

* *Identifier* 是将在限定标识符中使用的任何有效标识符。
* *ImportPath* 是一个字符串（原始或解释字符串，译注：例如 `\n` 和 "\n" 的区别，原始字符串或回车）

让我们看一些例子：

```go
import . "fmt"
import _ "io"
import log "github.com/sirupsen/logrus"
import m "math"
```

### 合并 Import 声明

导入两个或者更多的包可以有两种写法。一个是，我们可以写多个 import 声明：

```go
import "io"
import "bufio"
```

或者，我们可以将多个 import 声明合并（将多个导入放在一条导入声明中）：

```go
import (
    "io"
    "bufio"
)
```

第二种导入方式在导入很多个包的时候非常实用，然后多次重复的用 import 关键字导入包会降低可读性。如果你不使用自动导入之类的工具，例如： [https://github.com/bradfitz/goimports](https://github.com/bradfitz/goimports "https://github.com/bradfitz/goimports")，这种方式还可以减少按键次数。

### （短）导入路径

导入规范中使用的字符串文字（每个导入声明包含一个或多个导入规范）告诉导入哪个包。这个字符串称为导入路径。根据语言规范，它取决于如何解释导入路径（字符串）的实现方式，但在现实运用中它的路径相对包的第三方库目录或 `go env GOPATH / src` 目录（更多内容参考 [GOPATH](https://golang.org/doc/code.html#GOPATH "GOPATH") ）。

内置的包导入使用 “math” 或 “fmt” 等短导入路径。

### .go 文件剖析

每个 `.go` 文件的结构是相同的。首先是 package 语句，可选地在其前面加上通常是描述包的作用的注释。然后零个或多个导入声明。 接着包含零个或多个顶级声明。

```go
// description...
package main // package clause

// zero or more import declarations
import (
    "fmt"
    "strings"
)

import "strconv"

// top-level declarations

func main() {
    fmt.Println(strings.Repeat(strconv.FormatInt(15, 16), 5))
}
```

强制组织 (Enforced organisation) 不允许引入不必要的混乱，这简化了解析和基本的代码库跳转（导入声明不能放在 package 子句之前，也不能与顶级声明交错，所以它总是很容易找到）。

### 导入作用域

导入的作用域是文件块级别。这意味着它可以从整个文件中访问，但不能在整个包中被访问：

```go
// github.com/mlowicki/a/main.go
package main

import "fmt"

func main() {
    fmt.Println(a)
}

// github.com/mlowicki/a/foo.go
package main

var a int = 1

func hi() {
    fmt.Println("Hi!")
}
```

上述代码无法被成功编译：

```
> go build
// github.com/mlowicki/a
./foo.go:6:2: undefined: fmt
```

更多的关于作用域的内容参考之前发表的文章：[Scopes in Go](https://medium.com/golangspec/scopes-in-go-a6042bb4298c "Scopes in Go")

## 导入的类型

### 自定义包名

按照约定，导入路径的最后一个部分同时也是导入包的包名。当然，我们也可以不遵循这个约定：

```go
// github.com/mlowicki/main.go
package main

import (
    "fmt"
    "github.com/mlowicki/b"
)

func main() {
    fmt.Println(c.B)
}

// github.com/mlowicki/b/b.go
package c

var B = "b"
```

这个输出很明显是 *b* 。当然尽可能的遵循这些约定是更好的 — 很多工具也是依赖这个约定。如果自定义包名在导入的时候没有特别的指定，则使用来自包子句的名称来引用导入包的导出标识符：

```go
package main
import "fmt"
func main() {
    fmt.Println("Hi!")
}
```

也可以自定义一个包名称进行导入：

```go
// github.com/mlowicki/b/b.go
package b

var B = "b"

// github.com/mlowicki/main.go (依据原文含义，译者添加)
package main

import (
    "fmt"
    c "github.com/mlowicki/b"
)

func main() {
    fmt.Println(c.B)
}
```

这个输出结果和之前一样。如果我们的包具有与其它包相同的接口（导出的标识符），则这种导入形式非常有用。 一个这样的例子是 [https://github.com/sirupsen/logrus](https://github.com/sirupsen/logrus "https://github.com/sirupsen/logrus")，它有一个与 log 兼容的 API ：

```go
import log "github.com/sirupsen/logrus"
```

如果我们只使用内置日志包中的 API ，那么用导入 `log` 替换这样的导入不需要对源代码进行任何更改。它也有点短（但仍然有意义）所以可能会节省一些按键次数。

### 导入所有的导出标识符

例如：

```go
import m "math"
import "fmt"
```

可以使用指定的包的别名 (m.Exp) 或者导入的包名 (fmt.Prinln) 实现引用导出标识符。还有另一个方式不用通过限定标识符就可以访问导出标识符：

```go
package main

import (
    "fmt"
    . "math"
)

func main() {
    fmt.Println(Exp2(6))  // 64
}
```

什么时候这种用法有用呢？在测试中。假设我们有一个包 b 导入包 a。现在我们想给包 a 添加测试。如果测试也在包 a 中，并且测试也会导入包 b (因为到时需要在那实现一些东西)，那么我们将最终将会变成循环依赖，这是禁止的。绕过它的一种方法是将测试放入单独的包中，如 a_tests。然后我们需要导入包 a 并使用限定标识符引用每个导出的标识符。为了让我们的实现的更轻松，我们可以用点来导入包 a：

```go
import . "a"
```

然后引用包 a 中的导出标识符就不需要带上包名（就像测试是在同一个包中一样，但是那些非导出的标识符是不能访问的）

如果导入的包中存在至少一个同名的导出标识符，则无法使用点作为包名导入两个包：

```go
// github.com/mlowicki/c
package c

var V = "c"
// github.com/mlowkci/b
package b

var V = "b"

// github.com/mlowicki/a
package main

import (
    "fmt"
    . "github.com/mlowicki/b"
    . "github.com/mlowicki/c"
)

func main() {
    fmt.Println(V)
}
```

```
> go run main.go
// command-line-arguments
./main.go:6:2: V redeclared during import "github.com/mlowicki/c"
    previous declaration during import "github.com/mlowicki/b"
./main.go:6:2: imported and not used: "github.com/mlowicki/c"
```

### 使用空标识符

如果导入了包但是不使用，Golang的编译器将无法编译通过。

```go
package main

import "fmt"

func main() {}
```

使用点导入，其中所有导出的标识符都直接添加到导入文件块中，在编译时也会出现失败。唯一的绕过方式是使用空白标识符。需要知道init函数是什么，以便理解为什么我们需要导入空白标识符。参考之前init的介绍文章 [https://medium.com/golangspec/init-functions-in-go-eac191b3860a](https://medium.com/golangspec/init-functions-in-go-eac191b3860a) 我鼓励从上到下阅读这篇文章，但本质上，像如下的导入方式：

```go
import _ "math"
```

不需要在导入文件中使用包 math，但是无论如何都将执行导入包中的 init 函数（包和它的依赖关系将被初始化）。 如果我们只对导入包完成的初始化工作感兴趣，但我们不引用任何的导出标识符，那么就很有用。

> 如果一个包被导入没有被使用或者没有使用空标识符，那将编译失败

### 循环导入

Go 规范明确禁止循环导入 - 当包间接导入自身时。 最明显的情况是包 a 导入包 b 然后包 b 中也导入包 a：

```go
// github.com/mlowicki/a/main.go
package a

import "github.com/mlowicki/b"

var A = b.B

// github.com/mlowicki/b/main.go
package b

import "github.com/mlowicki/a"

var B = a.A
```

尝试构建这两个包中的任何一个都会导致错误：

```
> go build
can't load package: import cycle not allowed
package github.com/mlowicki/a
    imports github.com/mlowicki/b
    imports github.com/mlowicki/a
```

当然，比如 a -> b -> c -> d -> a 这种情况更加的复杂（x -> y 指包 x 导入包 y）。

包也是不能导入自己的：

```go
package main

import (
    "fmt"
    "github.com/mlowicki/a"
)

var A = "a"

func main() {
    fmt.Println(a.A)
}
```

编译上述代码将会提示错误：*can’t load package: import cycle not allowed*。

(完)

---

via：https://medium.com/golangspec/import-declarations-in-go-8de0fd3ae8ff

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[iloghyr](https://github.com/iloghyr)
校对：[无闻](https://github.com/Unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
