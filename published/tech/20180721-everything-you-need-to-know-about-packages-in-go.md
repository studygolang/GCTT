首发于：https://studygolang.com/articles/16765

# Go 语言中你应该知道的关于 Package 的所有

*一个 Go 编程语言的包管理及应用的全面概述*

如果你对像 **Java** 或者 **NodeJS** 这样的语言熟悉，那么你可能对**包**（译者注：原文中 **packages** ，后文中将其全部译为中文出现在文章表述中）相当熟悉了。包不是什么其他的，而是一个有着许多代码文件的目录，它从单个引用点显示不同的变量（特征）。让我来解释一下，这是什么意思。

设想在某个项目上工作，你需要不断的修改超过一千个函数。这之中的一些函数有相同的行为。比如，`toUpperCase` 和 `toLowerCase` 函数转变 `字符串` 的大小写，因此你把它们写在了一个单独的文件（*可能*是 **case.go**）里。也有一些其他的函数对 `字符串` 数据类型做一些其他操作，因此你也把它们写在了独立的文件里。

因为你可能有很多对于 `字符串` 数据类型进行一些操作的文件，因此你创建了一个名为 `string` 的目录，并将所有 `字符串` 相关的文件都放进去了。最后你将所有的这些目录放在一个将成为你的包的父目录里。整个包的结构看上去像下面这样。

```
package-name
├── string
|  ├── case.go
|  ├── trim.go
|  └── misc.go
└── number
	 ├── arithmetics.go
	 └── primes.go
```

我将详细地解释，我们如何从包中导入函数和变量，以及所有内容如何混合在一起形成了包，但是现在，设想你的包就像是一些包含着 `.go` 的目录。

每一个 Go 语言程序都必须是一些包的一部分。就像在  [**Getting started with Go**](https://medium.com/rungo/working-in-go-workspace-3b0576e0534a) 教程里面讨论的那样，一个独立可执行的 Go 语言程序必须有 `package main` 声明。如果一个程序是 `main` 包的一部分，那么在 `go install` 则会生成一个二进制文件，在执行时则会调用 `main` 函数。如果一个程序除了 `main` 包外还是其他包的一部分，那么在使用 `go install` 命令时会创建**包存档**文件（译者注：原文中 **package archive**）。**别担心，我会在接下来的话题中将这一切好好解释的**。

让我们来建立一个可执行的包。正如我们所知，为了创建一个二进制可执行文件，我们需要让我们的程序成为 `main` 包的一部分，而且必须要有一个作为执行入口点的 `main` 函数。

> ![图片 1](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/1.png)

包的名字是在 `src` 目录下包含的的目录名。在上面的情况下，`app` 是包，因为  `app` 是 `src`  目录的子目录。 因此，`go install app` 命令在 `GOPATH` 的 `src` 目录下寻找 `app` 子目录。之后编译这个包，并在 `bin` 目录下生成 `app` 的二进制可执行文件，这个生成的文件在终端可运行，因为 `bin` 目录在 `PATH` 中。

> 在上面例子中，像 `package main` 这样作为代码的第一行的**包声明可以与包名有所不同。因此，你可能发现一些包的包名（目录中的名称）与包声明的名称不一样。当你导入一个包的时候，包声明是用来创建包引用变量的，本文之后将进行说明。

`go install <package>` 命令寻找有 **`main` 包声明** 的任何一个文件。如果发现了这个文件，Go 就知道这是一个可执行程序并且需要生成一个二进制文件。一个包可以有许多文件，但只有一个有 `main` 函数的一个文件，因为这个文件将作为程序执行的入口点。

如果一个包中没有一个文件有 `main` 包的声明，那么 Go 就会在 `pkg` 目录下生成一个**包存档**（`.a`）文件。

> ![图片 2](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/2.png)

因此，`app` 不是一个可执行的包，它在 `pkg` 目录下生成了 `app.a` 文件。

**包的命名规则**

Go 语言社区建议对包使用简洁的名称。比如对**字符串通用**功能（译者注：原文中 **string utility** functions）使用 `strutils`，或者对与 HTTP 请求相关的函数使用 `http`。包名应避免 `under_scores`， `hy-phens`， `mixedCaps` 等形式。

## 建一个包

正如我们讨论的那样，有两种类型的包，一种是**可执行包**，另一种是**应用包**。可执行包对你而言更常用，因为你将运行它。一个应用包本身是不可运行的，除非它通过提供引用函数或其他重要的条件来提升作为可执行包的函数性。

正如我们所知，包只是一个目录，让我们来创建一个 包含 `src` 的 `greet` 目录并且在其中创建一些文件。这时，我们在每个文件的顶部都写上 `package greet` 声明来表明这是一个应用包。

> ![图片 3](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/3.png)

### 导出（包）成员

一个应用包应该给导入它的包提供一些变量。就像在 `JavaScript` 中的 `export` 语法一样，Go 语言中如果一个变量的名称以**大写字母**开头就是可导出的，其他所有的名称不以大写字母开头的变量都是这个包私有的。

> 在本文接下来的叙述中，我将使用**变量**这个词，去描述一个可导出的量，但这个这个可导出的量可以是任何类型的，比如 ` 常量 `、`map`、` 函数 `、` 结构体 `、` 数组 `、` 切片 ` 等。

让我们从 `day.go` 文件中导出一个 greeting 变量。

> ![图片 4](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/4.png)

在上面的程序中，`Morning` 这个变量可以从包中导出，但由于 `morning` 以小写字母开头则不可导出。

### 导入一个包

现在我们需要一个使用 `greet` 包的**可执行包**。让我们在 `src` 目录下建立一个 `app` 目录并在其中创建一个包含 `main` 包声明和 `main` 函数的文件 `entry.go` 。这里要注意，Go 包没有一个像在 Node 中的 `index.js` 一样的**入口文件命名系统**。对于一个可执行包而言，一个有 `main` 函数的文件就是程序执行的入口。

我们用 `import` 语法后跟包名来导入这个包。Go 程序首先在 `**GOROOT**/src` 目录中寻找包目录，如果没有找到，则会去 `**GOPATH**/src` 目录中继续寻找。由于 `fmt` 包是位于 `GOROOT/src` 目录的 Go 语言标准库中的一部分，它将会从该目录中导入。因为 Go 不能在 `GOROOT` 目录下找到 `greet` 包，它将在 `GOPATH/src` 目录下搜寻，这正是我们创建这个包的位置。

> ![图片 5](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/5.png)

上面的程序中因为 `morning` 变量不能被 `greet` 包导入，于是抛出了一个编译 error。如你说见，我们用 `.`（点）标记来访问从其他包中导入的变量。当你导入一个包的时候，Go 生成一个全局变量用作这个包的**包使用声明**。在上述示例中，`greet` 是 Go 生成的全局变量，因为我们在包含在 `greet` 包的程序中使用了 `package greet` 声明。

> ![图片 6](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/6.png)

我们可以用分组语法（括号）将 `fmt` 和 `greet` 包组合在一起导入。这次，我们的程序编译成功了，因为 `Morning` 变量对外部包而言是可获得的。

### 嵌套包

我们可以在一个包中嵌套另外的包。因为对于 Go 而言，包只是一个目录，这就像在一个已经存在的包中生成一个子目录一样。我们需要做的仅仅只是提供这个要被嵌套的包的相对路径。

> ![图片 7](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/7.png)
> ![图片 8](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/8.png)

### 包编译

正如我们在之前的学习中所了解的那样，`go run` 命令编译并执行一个程序。我们同样明白，`go install` 命令编译一个包并且生成一些二进制可执行文件或者包存档文件。这是为了避免对这些包每次都进行编译（对于被导入的包所在的那些程序而言）。`go install` 预编译一个包，在 Go 语言程序中是 `.a` 文件。

> 通常而言，当安装一个第三方包时，Go 会编译这个包并且生成一个存档文件。如果你在本地已经有这个包了，你的 **IDE** 可能会在你保存了这个包中的文件或修改了这个包后尽可能快的生成包存档文件。**如果你安装了一些 Go 的插件，那么在你保存这个包后 VSCode 就会编译它**。
> ![图片 9](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/9.png)

## 包的安装

当我们运行一个 Go 程序时，Go 语言编译器对包、在包中的文件和在包中的变量声明有特定的执行顺序。

### 包的作用域

**作用域是代码块中可使用已定义变量的区域**。包作用域是包内的一个区域，且可以从包中访问已声明的变量（对于包中的所有文件）。这个区域是在包中所有文件的最顶层块。

> ![图片 10](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/10.png)
> ![图片 11](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/11.png)

让我们来看看 `go run` 命令。这次，除了执行一个文件，我们使用 glob 规则来包含在 `app` 包中要执行的所有文件。Go 足够聪明，它可以找到应用的入口点 `entry.go`，因为它包含 `main` 函数。我们也可以用像下面这样的命令（文件名顺序并不重要）。

```
go run src/app/version.go src/app/entry.go
```

> `go install` 或者 `go build` 命令需要一个包名，其中包含包中所有的文件，我们不用像上面那样一一列举它们。

回到我们最主要的问题，我们可以在包中的任何地方使用在 `version.go` 文件中用声明的 `version` 变量，即使它并不能被导出（像 `Version`），因为它是在包的作用域中被声明的。如果 `version` 变量在函数中已经被声明，那么它就不在包作用域内，上面的程序也将无法编译成功。

**在同一个包中用同一名称重复声明全局变量是不被允许的**。因此，一旦 `version` 变量被声明，在这个包作用域内就不可以被重复声明。但是在其他区域你可以随心所欲的重复声明。

> ![图片 12](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/12.png)

### 变量初始化

当一个变量 `a` 依赖于另一个变量 `b`，那么要先声明 `b`，否则程序无法编译成功。Go 在函数内有以下规则。

> ![图片 13](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/13.png)

但是当这些变量是在包作用域声明时，它们在初始化周期（译者注：原文中 INItialization cycle）中声明。让我们来看看下面的简单例子。

> ![图片 14](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/14.png)

在上面例子中，首先 `c` 的值已经声明了，则它被声明了。在之后的初始化周期中，`b` 因为依赖于 `c`，且 `c` 的值已定，则它也被声明了。在最后的初始化周期中，`a` 被声明，且被 `b` 的值赋值。Go 可以解决像下面这样的复杂的初始化周期。

> ![图片 15](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/15.png)

​在上面例子中，首先 `c` 被声明了，之后因为 `b` 的值依赖于 `c`，且 `a` 的值依赖于 `b`，则 `b`、`a` 也依次被声明。你应该避免任何初始化循环，如所示下例这样陷入递归循环的初始化。

> ![图片 16](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/16.png)

另一个关于包作用域的例子是，将函数 `f` 放在独立的文件中，且该文件从主文件中引用变量 `c`。

> ![图片 17](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/17.png)
> ![图片 18](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/18.png)

### Init 函数

像 `main` 函数一样，`init` 函数在包被初始化时被 Go 调用。它不需要任何参数也不返回任何值。`init` 函数由 Go 隐式声明（译注：应该是由 Go 隐式调用），因此你无法从任何地方引用它（或者像 `init()` 这样来调用它）。在一个文件或包中，你可以有多个 `init` 函数。在文件中执行 `init` 函数的顺序和其出现顺序是一致的。（译注：词法文件名顺序，只是目前编译器的实现，Go 规范并没有要求这个顺序，因此程序不能依赖它）

> ![图片 19](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/19.png)

你可以在包中的任何位置使用 `init` 函数。这些 `init` 函数以词法文件名顺序（字母顺序）被调用。

> ![图片 20](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/20.png)

在所有的 `init` 函数被执行之后，`main` 函数被调用。因此，**`init` 函数的主要作用是将在全局代码中无法初始化的全局变量初始化。例如，数组的初始化。

> ![图片 21](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/21.png)

因为 `for` 语法在包作用域中不可用，所以我们可以在 `init` 函数中用 `for` 循环将大小为 `10` 的数组 `integers` 初始化。

### 包别名

当你导入一个包的时候，Go 使用这个包的包声明创建一个变量。如果你用一个名字导入多个包，将会导致冲突。

```go
// parent.go
package greet
var Message = "Hey there. I am parent."
// child.go
package greet
var Message = "Hey there. I am child."
```

> ![图片 22](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/22.png)

因此，我们使用包别名。我们在关键字 `impot` 和包名之间声明一个变量名作为引用这个包的新变量。

> ![图片 23](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/23.png)

在上面例子中，`greet/greet` 包现在由 `child` 变量引用。如果你注意到，我们用下划线作为 `greet` 包的别名。因为我们导入了 `greet` 但是并不使用它，Go 编译器会抱怨这种情况。为了避免它，我们将这个包的引用储存到 `_`，之后 Go 编译器就会忽略它了。

用**下划线**作为一个包的别名看似没什么用，但是当你想初始化一个包，除此之外并不使用它时，这样是非常有用的。

```go
// parent.go
package greet
import "fmt"
var Message = "Hey there. I am parent."
func init() {
	fmt.Println("greet/parent.go ==> INIt()")
}
// child.go
package greet
import "fmt"
var Message = "Hey there. I am child."
func init() {
	fmt.Println("greet/greet/child.go ==> INIt()")
}
```

> ![图片 24](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/24.png)

最需要记住的是，**每个包只初始化一次被导入的包**。因此如果包中有许多导入语句，在主包执行的生命周期中将只初始化一次被导入的包。

## 程序执行顺序

至此为止，我们了解了关于包的方方面面。现在让我们来整合一下对 Go 语言程序初始化的理解。

```
go run *.go
├── 被执行的主包
├── 初始化所有被导入的包
|  ├── 初始化所有被导入的包 ( 递归定义 )
|  ├── 初始化所有全局变量
|  └── INIt 函数以字母序被调用
└── 初始化主包
	 ├── 初始化所有全局变量
	 └── INIt 函数以字母序被调用
```

这里有一个验证它的小小例子。

```go
// version/get-version.go
package version
import "fmt"
func INIt() {
	fmt.Println("version/get-version.go ==> INIt()")
}
func getVersion() string {
	fmt.Println("version/get-version.go ==> getVersion()")
	return "1.0.0"
}
/***************************/
// version/entry.go
package version
import "fmt"
func init() {
	fmt.Println("version/entry.go ==> INIt()")
}
var Version = getLocalVersion()
func getLocalVersion() string {
	fmt.Println("version/entry.go ==> getLocalVersion()")
	return getVersion()
}
/***************************/
// app/fetch-version.go
package main
import (
	"fmt"
	"version"
)
func init() {
	fmt.Println("app/fetch-version.go ==> INIt()")
}
func fetchVersion() string {
	fmt.Println("app/fetch-version.go ==> fetchVersion()")
	return version.Version
}
/***************************/
// app/entry.go
package main
import "fmt"
func init() {
	fmt.Println("app/fetch-version.go ==> INIt()")
}
var myVersion = fetchVersion()
func main() {
	fmt.Println("app/fetch-version.go ==> fetchVersion()")
	fmt.Println("version ===> ", myVersion)
}
```

> ![图片 25](https://raw.githubusercontent.com/studygolang/gctt-images/master/everything-you-need-to-know-about-packages-in-go/25.png)

## 安装第三方包

（译注：Go 1.11 的 Module 已经支持版本控制导入）

安装第三方包就是将远程代码克隆到本地 `src/<package>` 目录下。不幸的是，Go 并不支持包版本或提供包管理器，但是提案正在[此](https://github.com/golang/proposal/blob/master/design/24301-versioned-go.md) 等候。
因为 Go 没有官方集中的包登记，因此你需要提供主机名路径。

```shell
$ Go get -u GitHub.com/jinzhu/gorm
```

上面命令将 URL 为 `http://github.com/jinzhu/gorm` 的文件导入，并将其保存在 `src/github.com/jinzhu/gorm` 目录。正如在嵌套包中讨论的那样，你可以像下面这样导入 `gorm` 包。

```go
package main
import "github.com/jinzhu/gorm"
// use ==> Gorm.SomeExportedMember
```

因此，如果你建了一个包并想让别人使用它，只需要在 GitHub 上发布它就可以了。如果你的包是可执行的，人们可以用它作为命令行工具；如果不是，他们可以在程序中导入你的包，并且将其作为应用模块使用。他们唯一需要做的就是输入下面的命令。

```shell
$ Go get GitHub.com/your-username/repo-name
```

---

via: https://medium.com/rungo/everything-you-need-to-know-about-packages-in-go-b8bac62b74cc

作者：[Uday Hiwarale](https://github.com/thatisuday)
译者：[yixiaoer](https://github.com/yixiaoer)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
