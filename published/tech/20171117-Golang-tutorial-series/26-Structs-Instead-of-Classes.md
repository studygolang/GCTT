已发布：https://studygolang.com/articles/12630

# 第 26 篇：结构体取代类

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 26 篇。

## Go 支持面向对象吗？

Go 并不是完全面向对象的编程语言。Go 官网的 [FAQ](https://golang.org/doc/faq#Is_Go_an_object-oriented_language) 回答了 Go 是否是面向对象语言，摘录如下。

> 可以说是，也可以说不是。虽然 Go 有类型和方法，支持面向对象的编程风格，但却没有类型的层次结构。Go 中的“接口”概念提供了一种不同的方法，我们认为它易于使用，也更为普遍。Go 也可以将结构体嵌套使用，这与子类化（Subclassing）类似，但并不完全相同。此外，Go 提供的特性比 C++ 或 Java 更为通用：子类可以由任何类型的数据来定义，甚至是内建类型（如简单的“未装箱的”整型）。这在结构体（类）中没有受到限制。

在接下来的教程里，我们会讨论如何使用 Go 来实现面向对象编程概念。与其它面向对象语言（如 Java）相比，Go 有很多完全不同的特性。

## 使用结构体，而非类

Go 不支持类，而是提供了[结构体](https://studygolang.com/articles/12263)。结构体中可以添加[方法](https://studygolang.com/articles/12264)。这样可以将数据和操作数据的方法绑定在一起，实现与类相似的效果。

为了加深理解，我们来编写一个示例吧。

在示例中，我们创建一个自定义[包](https://studygolang.com/articles/11893)，它帮助我们更好地理解，结构体是如何有效地取代类的。

在你的 Go 工作区创建一个名为 `oop` 的文件夹。在 `opp` 中再创建子文件夹 `employee`。在 `employee` 内，创建一个名为 `employee.go` 的文件。

文件夹结构会是这样：

```
workspacepath -> oop -> employee -> employee.go
```

请将 `employee.go` 里的内容替换为如下所示的代码。

```go
package employee

import (
	"fmt"
)

type Employee struct {
	FirstName   string
	LastName    string
	TotalLeaves int
	LeavesTaken int
}

func (e Employee) LeavesRemaining() {
	fmt.Printf("%s %s has %d leaves remaining", e.FirstName, e.LastName, (e.TotalLeaves - e.LeavesTaken))
}
```

在上述程序里，第 1 行指定了该文件属于 `employee` 包。而第 7 行声明了一个 `Employee` 结构体。在第 14 行，结构体 `Employee` 添加了一个名为 `LeavesRemaining` 的方法。该方法会计算和显示员工的剩余休假数。于是现在我们有了一个结构体，并绑定了结构体的方法，这与类很相似。

接着在 `oop` 文件夹里创建一个文件，命名为 `main.go`。

现在目录结构如下所示：

```
workspacepath -> oop -> employee -> employee.go
workspacepath -> oop -> main.go
```

`main.go` 的内容如下所示：

```go
package main

import "oop/employee"

func main() {
	e := employee.Employee {
		FirstName: "Sam",
		LastName: "Adolf",
		TotalLeaves: 30,
		LeavesTaken: 20,
	}
	e.LeavesRemaining()
}
```

我们在第 3 行引用了 `employee` 包。在 `main()`（第 12 行），我们调用了 `Employee` 的 `LeavesRemaining()` 方法。

由于有自定义包，这个程序不能在 go playground 上运行。你可以在你的本地运行，在 `workspacepath/bin/oop` 下输入命令 `go install opp`，程序会打印输出：

```bash
Sam Adolf has 10 leaves remaining
```

## 使用 New() 函数，而非构造器

我们上面写的程序看起来没什么问题，但还是有一些细节问题需要注意。我们看看当定义一个零值的 `employee` 结构体变量时，会发生什么。将 `main.go` 的内容修改为如下代码：

```go
package main

import "oop/employee"

func main() {
	var e employee.Employee
	e.LeavesRemaining()
}
```

我们的修改只是创建一个零值的 `Employee` 结构体变量（第 6 行）。该程序会输出：

```bash
has 0 leaves remaining
```

你可以看到，使用 `Employee` 创建的零值变量没有什么用。它没有合法的姓名，也没有合理的休假细节。

在像 Java 这样的 OOP 语言中，是使用构造器来解决这种问题的。一个合法的对象必须使用参数化的构造器来创建。

Go 并不支持构造器。如果某类型的零值不可用，需要程序员来隐藏该类型，避免从其他包直接访问。程序员应该提供一种名为 `NewT(parameters)` 的 [函数](https://studygolang.com/articles/11892)，按照要求来初始化 `T` 类型的变量。按照 Go 的惯例，应该把创建 `T` 类型变量的函数命名为 `NewT(parameters)`。这就类似于构造器了。如果一个包只含有一种类型，按照 Go 的惯例，应该把函数命名为 `New(parameters)`， 而不是 `NewT(parameters)`。

让我修改一下原先的代码，使得每当创建 `employee` 的时候，它都是可用的。

首先应该让 `Employee` 结构体不可引用，然后创建一个 `New` 函数，用于创建 `Employee` 结构体变量。在 `employee.go` 中输入下面代码：

```go
package employee

import (
	"fmt"
)

type employee struct {
	firstName   string
	lastName    string
	totalLeaves int
	leavesTaken int
}

func New(firstName string, lastName string, totalLeave int, leavesTaken int) employee {
	e := employee {firstName, lastName, totalLeave, leavesTaken}
	return e
}

func (e employee) LeavesRemaining() {
	fmt.Printf("%s %s has %d leaves remaining", e.firstName, e.lastName, (e.totalLeaves - e.leavesTaken))
}
```

我们进行了一些重要的修改。我们把 `Employee` 结构体的首字母改为小写 `e`，也就是将 `type Employee struct` 改为了 `type employee struct`。通过这种方法，我们把 `employee` 结构体变为了不可引用的，防止其他包对它的访问。除非有特殊需求，否则也要隐藏所有不可引用的结构体的所有字段，这是 Go 的最佳实践。由于我们不会在外部包需要 `employee` 的字段，因此我们也让这些字段无法引用。

同样，我们还修改了 `LeavesRemaining()` 的方法。

现在由于 `employee` 不可引用，因此不能在其他包内直接创建 `Employee` 类型的变量。于是我们在第 14 行提供了一个可引用的 `New` 函数，该函数接收必要的参数，返回一个新创建的 `employee` 结构体变量。

这个程序还需要一些必要的修改，但现在先运行这个程序，理解一下当前的修改。如果运行当前程序，编译器会报错，如下所示：

```bash
go/src/constructor/main.go:6: undefined: employee.Employee
```

这是因为我们将 `Employee` 设置为不可引用，因此编译器会报错，提示该类型没有在 `main.go` 中定义。很完美，正如我们期望的一样，其他包现在不能轻易创建零值的 `employee` 变量了。我们成功地避免了创建不可用的 `employee` 结构体变量。现在创建 `employee` 变量的唯一方法就是使用 `New` 函数。

如下所示，修改 `main.go` 里的内容。

```go
package main

import "oop/employee"

func main() {
	e := employee.New("Sam", "Adolf", 30, 20)
	e.LeavesRemaining()
}
```

该文件唯一的修改就是第 6 行。通过向 `New` 函数传入所需变量，我们创建了一个新的 `employee` 结构体变量。

下面是修改后的两个文件的内容。

employee.go

```go
package employee

import (
	"fmt"
)

type employee struct {
	firstName   string
	lastName    string
	totalLeaves int
	leavesTaken int
}

func New(firstName string, lastName string, totalLeave int, leavesTaken int) employee {
	e := employee {firstName, lastName, totalLeave, leavesTaken}
	return e
}

func (e employee) LeavesRemaining() {
	fmt.Printf("%s %s has %d leaves remaining", e.firstName, e.lastName, (e.totalLeaves - e.leavesTaken))
}
```

main.go

```go
package main

import "oop/employee"

func main() {
	e := employee.New("Sam", "Adolf", 30, 20)
	e.LeavesRemaining()
}
```

运行该程序，会输出：

```bash
Sam Adolf has 10 leaves remaining
```

现在你能明白了，虽然 Go 不支持类，但结构体能够很好地取代类，而以 `New(parameters)` 签名的方法可以替代构造器。

关于 Go 中的类和构造器到此结束。祝你愉快。

**上一教程 - [Mutex](https://studygolang.com/articles/12598)**

**下一教程 - [组合取代继承](https://studygolang.com/articles/12680)**

---

via: https://golangbot.com/structs-instead-of-classes/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
