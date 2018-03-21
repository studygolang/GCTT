# GO 语言小技巧：imports 大全

包导入（imports）是现在大部分编程语言的必要组成部分，Go 语言也不例外。大多数情况下 Go 语言中的包导入很简单。对多数用户来说，了解了基础知识已经足够。但是还是有一些陷阱值得注意。

让我们快速过一下基础。假设你需要在终端输出东西，这就需要 `fmt` 包。所以我们直接导入它。

```go
import "fmt"

func main() {
    fmt.Println("Go is great!")
}
```

当导入 `fmt` 之后，你就能使用这个包中导出的结构体（struct）和函数（function）。你所需要做的只是在这些函数前加上 `fmt.`，就像上述例子中的 `fmt.Println`。

当导入了多个包之后，比起一个一个地 `import`，更加常见的做法是用一个 `import ( )` 直接包含它们。

```go
import (
    "fmt"
    "bytes"
)
```

这样更加简洁，同时又不会影响代码。

## 导入重命名

Go 语言中导入语句一个很有用的特性是包导入重命名。常见的用法就是为导入的包起一个别名。

举个例子。当想使用 `discordgo` 包中函数的时候，每次我们都得输入一次。有了包导入重命名，我们就可以用 `dg` 代替。

```go
import dg "github.com/bwmarrin/discordgo"

func main() {
    err := dg.New()
}
```

我们也可以借助这个特性来避免包名冲突。当我们要使用两个相同名称的包时，就可以给其中的一个起一个别名。下面的例子中，我们就导入了两个包名都为 `rand` 的包。

```go
import (
    "math/rand"
    crand "crypto/rand"
)
```

现在我们可以通过 `rand` 来使用第一个包中的函数和变量，而通过 `crand` 来使用第二个包中的函数和变量。

## 包命名 vs. 包导入

过去我们经常以 `go-` 作为前缀来给包命名（例如 [go-bindata](https://github.com/jteeuwen/go-bindata) 或者 [go-iter](https://github.com/hgfischer/go-iter) 等等）。如果你使用过 [gopkg.in](https://labix.org/gopkg.in)，也应该有通过版本号引用过包的经历。

```
gopkg.in/pkg.v3      → github.com/go-pkg/pkg (branch/tag v3, v3.N, or v3.N.M)
gopkg.in/user/pkg.v3 → github.com/user/pkg   (branch/tag v3, v3.N, or v3.N.M)
```

例如。你希望在项目中使用 [json-iterator/go](https://github.com/json-iterator/go)，你可能会这样写：

```go
import "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary
json.Marshal(&data)
```

“等一下！”。你会说。“它叫 `json-iterator/go`，所以我们为什么不是用 `go` 而用 `jsoniter` 来引用其函数和变量呢？标准库中的 `fmt` 和 `encoding/json` 这些包不就是这样的吗？”

是的，人们通常倾向于让包名和它的 URL 相对应。但是如何使用包并不是由它的 URL 来决定的。如果你看一下 `json-iterator/go` 的源代码，就会发现其中每个源文件的顶上都有 `package jsoniter` 这个语句。这也才是决定这个包该被如何引用的关键。为了便于理解和使用，标准库中的包只是正好将包名定义的和各自的路径对应上罢了。

另一个例子是 Google 的 [Youtube](https://scene-si.org/2018/01/25/go-tips-and-tricks-almost-everything-about-imports/google.golang.org/api/youtube/v3) 包。它的导入路径是 `google.golang.org/api/youtube/v3`，所以一开始你可能会以为应该使用 `v3` 来引用它的变量和函数。但是，只要看一下源码，就会发现它的包名其实是 `youtube`。所以说，你应该这样写：

```go
import "google.golang.org/api/youtube/v3"

youtubeService, err := youtube.New(oauthHttpClient)
```

## 句点导入

句点导入相对来说并不为多少人知道，也不常用。它的作用是将包导入到当前代码所在的命名空间中，这样就可以直接使用其中的变量和函数而不必通过包名来引用了。要使用这种导入，只要将 `.` 作为包的别名即可。下面的例子展示了句点导入和普通导入的区别：

```go
import (
    . "math"
    "fmt"
)

fmt.Println(Pi)
```

```go
import (
    "math"
    "fmt"
)

fmt.Println(math.Pi)
```

可以看见，当我们使用句点导入 `math` 包后，就不需要通过 `math.` 前缀来引用 `Pi` 了。

值得注意的一点是，当使用句点导入时，被导入包中不能包含和当前命名空间下同名的变量或者函数。举例来说，我们使用句点导入 `fmt` 包，而同时我们在自己的包中也定义了 `Println` 函数。这时编译器就会提示我们存在重名的函数。

这种导入方法在测试中很常见。句点导入只会导入包中公有的结构体和函数，而不会暴露其私有的细节。这对于测试来说是好的，因为这样你就能确定你包中的公有接口是完全函数化的。在你将包和你的测试用例共享命名空间后，你就能知道那些属性和函数不是公有的，同样这些东西也不会暴露给你包的使用者。

## 相对路径导入

如果之前你使用命令行做过一些文件系统方面的事，你应该懂得 “相对路径”。自然地，相对路径导入就是使用相对路径来导入包的方法。下面例子中，当我们想在 helloworld.go 中导入 `greeting` 包时，我们输入了 `import "./greeting"` 。

```
$GOPATH 
├──bin  
│   └──hello  
├──pkg  
└──src  
    └──someFolder  
        ├──helloworld.go  
        └──greeting  
            └──greeting.go
```

不幸的是，编译不能通过并给出了错误： `local import "./greeting" in non-local package` 。这是因为在工作区（通常是 `$GOPATH/src`）中相对路径导入是不允许的。它只能在工作区之外被使用。

相对路径导入这种有意的设计是为了方便开发者在工作区之外进行快速测试和实验。针对这种设计有过许多讨论。虽然当 fork 一个项目并提出来的时候，这种方法允许你不用修改所有的导入。然后，这种设计可能还要存在一段时间。

## 空导入

如果你曾经对于 Go 语言不允许导入未使用的包感到苦恼，那么你会很高兴存在空导入。空导入通常用来处理你当前未使用但随时可能用到的包。然而它还有一个用处。如果你曾经在 Go 语言中处理过图像或者数据库，可能曾经看到过下面某种用法：

```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)
```

```go
import (  
    "image"
    _ "image/gif"
    _ "image/png"
    _ "image/jpeg"
)
```

你也许会疑问为什么会有这种用法。有意义吗？然而实际上这样做是为了执行其中的 `init()` 方法。

距离来说，`image/png` 包中的 `init()` 方法是这样的：

```go
func init() {  
        image.RegisterFormat("png", pngHeader, Decode, DecodeConfig)
}
```

这里，为 `image` 包注册了 png 格式。基本上就是告诉了当遇到 png 格式图像时的处理方法。

[github.com/go-sql-driver/mysql](https://scene-si.org/2018/01/25/go-tips-and-tricks-almost-everything-about-imports/github.com/go-sql-driver/mysql) 包也是如此：

```go
func init() {
	sql.Register("mysql", &MySQLDriver{})
}
```

这个包为 `database/sql` 提供了处理 MySQL 数据库的必要信息。

## 结论

我知道许多的编程语言都有类似 import 的语法。PHP 有 `include`，Node.js/ES6 有 `import` 和 `require`，Java 有自己的 `import` 语句，甚至连 C/C++ 都有 `#include` 等等。虽然它们的表现形式各不相同。可以看到，与其他编程语言不一样的是，Go 语言语法虽然简单，却提供了额外的功能实用性。

----------------

via: https://scene-si.org/2018/01/25/go-tips-and-tricks-almost-everything-about-imports/

作者：[Tit Petric](https://scene-si.org/about)
译者：[alfred-zhong](https://github.com/alfred-zhong)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
