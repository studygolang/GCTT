已发布：https://studygolang.com/articles/11893

# 第 7 部分：包

这是 Golang 系列教程的第 7 个教程。

### 什么是包，为什么使用包？

到目前为止，我们看到的 Go 程序都只有一个文件，文件里包含一个 main [函数](https://studygolang.com/articles/11892)和几个其他的函数。在实际中，这种把所有源代码编写在一个文件的方法并不好用。以这种方式编写，代码的重用和维护都会很困难。而包（Package）解决了这样的问题。

**包用于组织 Go 源代码，提供了更好的可重用性与可读性**。由于包提供了代码的封装，因此使得 Go 应用程序易于维护。

例如，假如我们正在开发一个 Go 图像处理程序，它提供了图像的裁剪、锐化、模糊和彩色增强等功能。一种组织程序的方式就是根据不同的特性，把代码放到不同的包中。比如裁剪可以是一个单独的包，而锐化是另一个包。这种方式的优点是，由于彩色增强可能需要一些锐化的功能，因此彩色增强的代码只需要简单地导入（我们会在随后讨论）锐化功能的包，就可以使用锐化的功能了。这样的方式使得代码易于重用。

我们会逐步构建一个计算矩形的面积和对角线的应用程序。

通过这个程序，我们会更好地理解包。

### main 函数和 main 包

所有可执行的 Go 程序都必须包含一个 main 函数。这个函数是程序运行的入口。main 函数应该放置于 main 包中。

**`package packagename` 这行代码指定了某一源文件属于一个包。它应该放在每一个源文件的第一行。**

下面开始为我们的程序创建一个 main 函数和 main 包。**在 Go 工作区内的 src 文件夹中创建一个文件夹，命名为 `geometry`**。在 `geometry` 文件夹中创建一个 `geometry.go` 文件。

在 geometry.go 中编写下面代码。

```go
// geometry.go
package main 

import "fmt"

func main() {  
    fmt.Println("Geometrical shape properties")
}
```

`package main` 这一行指定该文件属于 main 包。`import "packagename"` 语句用于导入一个已存在的包。在这里我们导入了 `fmt` 包，包内含有 Println 方法。接下来是 main 函数，它会打印 `Geometrical shape properties`。

键入 `go install geometry`，编译上述程序。该命令会在 `geometry` 文件夹内搜索拥有 main 函数的文件。在这里，它找到了 `geometry.go`。接下来，它编译并产生一个名为 `geometry` （在 windows 下是 `geometry.exe`）的二进制文件，该二进制文件放置于工作区的 bin 文件夹。现在，工作区的目录结构会是这样：

```
src
    geometry
        gemometry.go
bin
    geometry
```

键入 `workspacepath/bin/geometry`，运行该程序。请用你自己的 Go 工作区来替换 `workspacepath`。这个命令会执行 bin 文件夹里的 `geometry` 二进制文件。你应该会输出 `Geometrical shape properties`。

### 创建自定义的包

我们将组织代码，使得所有与矩形有关的功能都放入 `rectangle` 包中。

我们会创建一个自定义包 `rectangle`，它有一个计算矩形的面积和对角线的函数。

**属于某一个包的源文件都应该放置于一个单独命名的文件夹里。按照 Go 的惯例，应该用包名命名该文件夹。**

因此，我们在 `geometry` 文件夹中，创建一个命名为 `rectangle` 的文件夹。在 `rectangle` 文件夹中，所有文件都会以 `package rectangle` 作为开头，因为它们都属于 rectangle 包。

在我们之前创建的 rectangle 文件夹中，再创建一个名为 `rectprops.go` 的文件，添加下列代码。

```go
// rectprops.go
package rectangle

import "math"

func Area(len, wid float64) float64 {  
    area := len * wid
    return area
}

func Diagonal(len, wid float64) float64 {  
    diagonal := math.Sqrt((len * len) + (wid * wid))
    return diagonal
}
```

在上面的代码中，我们创建了两个函数用于计算 `Area` 和 `Diagonal`。矩形的面积是长和宽的乘积。矩形的对角线是长与宽平方和的平方根。`math` 包下面的 `Sqrt` 函数用于计算平方根。

注意到函数 Area 和 Diagonal 都是以大写字母开头的。这是有必要的，我们将会很快解释为什么需要这样做。

### 导入自定义包

为了使用自定义包，我们必须要先导入它。导入自定义包的语法为 `import path`。我们必须指定自定义包相对于工作区内 `src` 文件夹的相对路径。我们目前的文件夹结构是：

```
src
    geometry
        geometry.go
        rectangle
            rectprops.go
```

`import "geometry/rectangle"` 这一行会导入 rectangle 包。

在 `geometry.go` 里面添加下面的代码：

```go
// geometry.go
package main 

import (  
    "fmt"
    "geometry/rectangle" // 导入自定义包
)

func main() {  
    var rectLen, rectWidth float64 = 6, 7
    fmt.Println("Geometrical shape properties")
    /*Area function of rectangle package used*/
    fmt.Printf("area of rectangle %.2f\n", rectangle.Area(rectLen, rectWidth))
    /*Diagonal function of rectangle package used*/
    fmt.Printf("diagonal of the rectangle %.2f ", rectangle.Diagonal(rectLen, rectWidth))
}
``` 

上面的代码导入了 `rectangle` 包，并调用了里面的 Area 和 Diagonal 函数，得到矩形的面积和对角线。Printf 内的格式说明符 `%.2f` 会将浮点数截断到小数点两位。应用程序的输出为：

```
Geometrical shape properties  
area of rectangle 42.00  
diagonal of the rectangle 9.22
```

### 导出名字（Exported Names）

我们将 rectangle 包中的函数 Area 和 Diagonal 首字母大写。在 Go 中这具有特殊意义。在 Go 中，任何以大写字母开头的变量或者函数都是被导出的名字。其它包只能访问被导出的函数和变量。在这里，我们需要在 main 包中访问 Area 和 Diagonal 函数，因此会将它们的首字母大写。

在 `rectprops.go` 中，如果函数名从 `Area(len, wid float64)` 变为 `area(len, wid float64)`，并且在 `geometry.go` 中， `rectangle.Area(rectLen, rectWidth)` 变为 `rectangle.area(rectLen, rectWidth)`， 则该程序运行时，编译器会抛出错误 `geometry.go:11: cannot refer to unexported name rectangle.area`。因为如果想在包外访问一个函数，它应该首字母大写。

### init 函数

所有包都可以包含一个 `init` 函数。init 函数不应该有任何返回值类型和参数，在我们的代码中也不能显式地调用它。init 函数的形式如下：

```go
func init() {  
}
```

init 函数可用于执行初始化任务，也可用于在开始执行之前验证程序的正确性。

包的初始化顺序如下：

1. 首先初始化包级别（Package Level）的变量
2. 紧接着调用 init 函数。包可以有多个 init 函数（在一个文件或分布于多个文件中），它们按照编译器解析它们的顺序进行调用。

如果一个包导入了另一个包，会先初始化被导入的包。

尽管一个包可能会被导入多次，但是它只会被初始化一次。

为了理解 init 函数，我们接下来对程序做了一些修改。

首先在 `rectprops.go` 文件中添加了一个 init 函数。

```go
// rectprops.go
package rectangle

import "math"  
import "fmt"

/*
 * init function added
 */
func init() {  
    fmt.Println("rectangle package initialized")
}
func Area(len, wid float64) float64 {  
    area := len * wid
    return area
}

func Diagonal(len, wid float64) float64 {  
    diagonal := math.Sqrt((len * len) + (wid * wid))
    return diagonal
}
```

我们添加了一个简单的 init 函数，它仅打印 `rectangle package initialized`。

现在我们来修改 main 包。我们知道矩形的长和宽都应该大于 0，我们将在 `geometry.go` 中使用 init 函数和包级别的变量来检查矩形的长和宽。

修改 `geometry.go` 文件如下所示：

```go
// geometry.go
package main 

import (  
    "fmt"
    "geometry/rectangle" // 导入自定义包
    "log"
)
/*
 * 1. 包级别变量
*/
var rectLen, rectWidth float64 = 6, 7 

/*
*2. init 函数会检查长和宽是否大于0
*/
func init() {  
    println("main package initialized")
    if rectLen < 0 {
        log.Fatal("length is less than zero")
    }
    if rectWidth < 0 {
        log.Fatal("width is less than zero")
    }
}

func main() {  
    fmt.Println("Geometrical shape properties")
    fmt.Printf("area of rectangle %.2f\n", rectangle.Area(rectLen, rectWidth))
    fmt.Printf("diagonal of the rectangle %.2f ",rectangle.Diagonal(rectLen, rectWidth))
}
```

我们对 `geometry.go` 做了如下修改：

1. 变量 **rectLen** 和 **rectWidth** 从 main 函数级别移到了包级别。
2. 添加了 init 函数。当 rectLen 或 rectWidth 小于 0 时，init 函数使用 **log.Fatal** 函数打印一条日志，并终止了程序。

main 包的初始化顺序为：

1. 首先初始化被导入的包。因此，首先初始化了 rectangle 包。
2. 接着初始化了包级别的变量 **rectLen** 和 **rectWidth**。
3. 调用 init 函数。
4. 最后调用 main 函数。

当运行该程序时，会有如下输出。

```
rectangle package initialized  
main package initialized  
Geometrical shape properties  
area of rectangle 42.00  
diagonal of the rectangle 9.22  
```
果然，程序会首先调用 rectangle 包的 init 函数，然后，会初始化包级别的变量 **rectLen** 和 **rectWidth**。接着调用 main 包里的 init 函数，该函数检查 rectLen 和 rectWidth 是否小于 0，如果条件为真，则终止程序。我们会在单独的教程里深入学习 if 语句。现在你可以认为 `if rectLen < 0` 能够检查 `rectLen` 是否小于 0，并且如果是，则终止程序。`rectWidth` 条件的编写也是类似的。在这里两个条件都为假，因此程序继续执行。最后调用了 main 函数。

让我们接着稍微修改这个程序来学习使用 init 函数。

将 `geometry.go` 中的 `var rectLen, rectWidth float64 = 6, 7` 改为 `var rectLen, rectWidth float64 = -6, 7`。我们把 `rectLen` 初始化为负数。

现在当运行程序时，会得到：

```
rectangle package initialized  
main package initialized  
2017/04/04 00:28:20 length is less than zero  
```

像往常一样， 会首先初始化 rectangle 包，然后是 main 包中的包级别的变量 rectLen 和 rectWidth。rectLen 为负数，因此当运行 init 函数时，程序在打印 `length is less than zero` 后终止。

本代码可以在 [github](https://github.com/golangbot/geometry) 下载。

### 使用空白标识符（Blank Identifier）

导入了包，却不在代码中使用它，这在 Go 中是非法的。当这么做时，编译器是会报错的。其原因是为了避免导入过多未使用的包，从而导致编译时间显著增加。将 `geometry.go` 中的代码替换为如下代码：

```go
// geometry.go
package main 

import (
    "geometry/rectangle" // 导入自定的包
)
func main() {

}
```

上面的程序将会抛出错误 `geometry.go:6: imported and not used: "geometry/rectangle"`。

然而，在程序开发的活跃阶段，又常常会先导入包，而暂不使用它。遇到这种情况就可以使用空白标识符 `_`。

下面的代码可以避免上述程序的错误：

```go
package main

import (  
    "geometry/rectangle" 
)

var _ = rectangle.Area // 错误屏蔽器

func main() {

}
```

`var _ = rectangle.Area` 这一行屏蔽了错误。我们应该了解这些错误屏蔽器（Error Silencer）的动态，在程序开发结束时就移除它们，包括那些还没有使用过的包。由此建议在 import 语句下面的包级别范围中写上错误屏蔽器。

有时候我们导入一个包，只是为了确保它进行了初始化，而无需使用包中的任何函数或变量。例如，我们或许需要确保调用了 rectangle 包的 init 函数，而不需要在代码中使用它。这种情况也可以使用空白标识符，如下所示。

```go
package main 

import (
    _ "geometry/rectangle" 
)
func main() {

}
```

运行上面的程序，会输出 `rectangle package initialized`。尽管在所有代码里，我们都没有使用这个包，但还是成功初始化了它。

包的介绍到此结束。希望您喜欢本篇，请留下您的宝贵的反馈和评论。

**下一教程 - if else 语句**

via: https://golangbot.com/packages/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
