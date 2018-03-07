已发布：https://studygolang.com/articles/12325

# 第 19 部分：接口（二）

欢迎来到 Golang 系列教程的第 19 个教程。接口共有两个教程，这是我们第二个教程。如果你还没有阅读前面的教程，请你阅读[接口（一）](https://studygolang.com/articles/12266)。

### 实现接口：指针接受者与值接受者

在[接口（一）](https://studygolang.com/articles/12266)上的所有示例中，我们都是使用值接受者（Value Receiver）来实现接口的。我们同样可以使用指针接受者（Pointer Receiver）来实现接口。只不过在用指针接受者实现接口时，还有一些细节需要注意。我们通过下面的代码来理解吧。 

```go
package main

import "fmt"

type Describer interface {  
	Describe()
}
type Person struct {  
	name string
	age  int
}

func (p Person) Describe() { // 使用值接受者实现  
	fmt.Printf("%s is %d years old\n", p.name, p.age)
}

type Address struct {
	state   string
	country string
}

func (a *Address) Describe() { // 使用指针接受者实现
	fmt.Printf("State %s Country %s", a.state, a.country)
}

func main() {  
	var d1 Describer
	p1 := Person{"Sam", 25}
	d1 = p1
	d1.Describe()
	p2 := Person{"James", 32}
	d1 = &p2
	d1.Describe()

	var d2 Describer
	a := Address{"Washington", "USA"}

	/* 如果下面一行取消注释会导致编译错误：
	   cannot use a (type Address) as type Describer
	   in assignment: Address does not implement
	   Describer (Describe method has pointer
	   receiver)
	*/
	//d2 = a

	d2 = &a // 这是合法的
	// 因为在第 22 行，Address 类型的指针实现了 Describer 接口
	d2.Describe()

}
```
[在线运行程序](https://play.golang.org/p/IzspYiAQ82)  

在上面程序中的第 13 行，结构体 `Person` 使用值接受者，实现了 `Describer` 接口。  

我们在讨论[方法](https://studygolang.com/articles/12264)的时候就已经提到过，使用值接受者声明的方法，既可以用值来调用，也能用指针调用。**不管是一个值，还是一个可以解引用的指针，调用这样的方法都是合法的**。

`p1` 的类型是 `Person`，在第 29 行，`p1` 赋值给了 `d1`。由于 `Person` 实现了接口变量 `d1`，因此在第 30 行，会打印 `Sam is 25 years old`。

接下来在第 32 行，`d1` 又赋值为 `&p2`，在第 33 行同样打印输出了 `James is 32 years old`。棒棒哒。:) 

在 22 行，结构体 `Address` 使用指针接受者实现了 `Describer` 接口。 

在上面程序里，如果去掉第 45 行的注释，我们会得到编译错误：`main.go:42: cannot use a (type Address) as type Describer in assignment: Address does not implement Describer (Describe method has pointer receiver)`。这是因为在第 22 行，我们使用 `Address` 类型的指针接受者实现了接口 `Describer`，而接下来我们试图用 `a` 来赋值 `d2`。然而 `a` 属于值类型，它并没有实现 `Describer` 接口。你应该会很惊讶，因为我们曾经学习过，使用指针接受者的[方法](https://studygolang.com/articles/12264)，无论指针还是值都可以调用它。那么为什么第 45 行的代码就不管用呢？

**其原因是：对于使用指针接受者的方法，用一个指针或者一个可取得地址的值来调用都是合法的。但接口中存储的具体值（Concrete Value）并不能取到地址，因此在第 45 行，对于编译器无法自动获取 `a` 的地址，于是程序报错**。  

第 47 行就可以成功运行，因为我们将 `a` 的地址 `&a` 赋值给了 `d2`。  

程序的其他部分不言而喻。该程序会打印：  

```
Sam is 25 years old  
James is 32 years old  
State Washington Country USA  
```

### 实现多个接口

类型可以实现多个接口。我们看看下面程序是如何做到的。  

```go
package main

import (  
	"fmt"
)

type SalaryCalculator interface {  
	DisplaySalary()
}

type LeaveCalculator interface {  
	CalculateLeavesLeft() int
}

type Employee struct {  
	firstName string
	lastName string
	basicPay int
	pf int
	totalLeaves int
	leavesTaken int
}

func (e Employee) DisplaySalary() {  
	fmt.Printf("%s %s has salary $%d", e.firstName, e.lastName, (e.basicPay + e.pf))
}

func (e Employee) CalculateLeavesLeft() int {  
	return e.totalLeaves - e.leavesTaken
}

func main() {  
	e := Employee {
		firstName: "Naveen",
		lastName: "Ramanathan",
		basicPay: 5000,
		pf: 200,
		totalLeaves: 30,
		leavesTaken: 5,
	}
	var s SalaryCalculator = e
	s.DisplaySalary()
	var l LeaveCalculator = e
	fmt.Println("\nLeaves left =", l.CalculateLeavesLeft())
}
```
[在线运行程序](https://play.golang.org/p/DJxS5zxBcV)  

上述程序在第 7 行和第 11 行分别声明了两个接口：`SalaryCalculator` 和 `LeaveCalculator`。  

第 15 行定义了结构体 `Employee`，它在第 24 行实现了 `SalaryCalculator` 接口的 `DisplaySalary` 方法，接着在第 28 行又实现了 `LeaveCalculator` 接口里的 `CalculateLeavesLeft` 方法。于是 `Employee` 就实现了 `SalaryCalculator` 和 `LeaveCalculator` 两个接口。  

第 41 行，我们把 `e` 赋值给了 `SalaryCalculator` 类型的接口变量 ，而在 43 行，我们同样把 `e` 赋值给 `LeaveCalculator` 类型的接口变量 。由于 `e` 的类型 `Employee` 实现了 `SalaryCalculator` 和 `LeaveCalculator` 两个接口，因此这是合法的。  

该程序会输出：

```
Naveen Ramanathan has salary $5200  
Leaves left = 25  
```

### 接口的嵌套
尽管 Go 语言没有提供继承机制，但可以通过嵌套其他的接口，创建一个新接口。  

我们来看看这如何实现。  

```go
package main

import (  
	"fmt"
)

type SalaryCalculator interface {  
	DisplaySalary()
}

type LeaveCalculator interface {  
	CalculateLeavesLeft() int
}

type EmployeeOperations interface {  
	SalaryCalculator
	LeaveCalculator
}

type Employee struct {  
	firstName string
	lastName string
	basicPay int
	pf int
	totalLeaves int
	leavesTaken int
}

func (e Employee) DisplaySalary() {  
	fmt.Printf("%s %s has salary $%d", e.firstName, e.lastName, (e.basicPay + e.pf))
}

func (e Employee) CalculateLeavesLeft() int {  
	return e.totalLeaves - e.leavesTaken
}

func main() {  
	e := Employee {
		firstName: "Naveen",
		lastName: "Ramanathan",
		basicPay: 5000,
		pf: 200,
		totalLeaves: 30,
		leavesTaken: 5,
	}
	var empOp EmployeeOperations = e
	empOp.DisplaySalary()
	fmt.Println("\nLeaves left =", empOp.CalculateLeavesLeft())
}
```
[在线运行程序](https://play.golang.org/p/Hia7D-WbZp)

在上述程序的第 15 行，我们创建了一个新的接口 `EmployeeOperations`，它嵌套了两个接口：`SalaryCalculator` 和 `LeaveCalculator`。

如果一个类型定义了 `SalaryCalculator` 和 `LeaveCalculator` 接口里包含的方法，我们就称该类型实现了 `EmployeeOperations` 接口。

在第 29 行和第 33 行，由于 `Employee` 结构体定义了 `DisplaySalary` 和 `CalculateLeavesLeft` 方法，因此它实现了接口 `EmployeeOperations`。

在 46 行，`empOp` 的类型是 `EmployeeOperations`，`e` 的类型是 `Employee`，我们把 `empOp` 赋值为 `e`。接下来的两行，`empOp` 调用了 `DisplaySalary()` 和 `CalculateLeavesLeft()` 方法。

该程序输出：

```
Naveen Ramanathan has salary $5200
Leaves left = 25
```

### 接口的零值
接口的零值是 `nil`。对于值为 `nil` 的接口，其底层值（Underlying Value）和具体类型（Concrete Type）都为 `nil`。

```go
package main

import "fmt"

type Describer interface {  
	Describe()
}

func main() {  
	var d1 Describer
	if d1 == nil {
		fmt.Printf("d1 is nil and has type %T value %v\n", d1, d1)
	}
}
```
[在线运行程序](https://play.golang.org/p/vwYHC6Y78H)  

上面程序里的 `d1` 等于 `nil`，程序会输出：

```
d1 is nil and has type <nil> value <nil>
```

对于值为 `nil` 的接口，由于没有底层值和具体类型，当我们试图调用它的方法时，程序会产生 `panic` 异常。

```go
package main

type Describer interface {
	Describe()
}

func main() {  
	var d1 Describer
	d1.Describe()
}
```
[在线运行程序](https://play.golang.org/p/rM-rY0uGTI)  

在上述程序中，`d1` 等于 `nil`，程序产生运行时错误 `panic`： **`panic: runtime error: invalid memory address or nil pointer dereference 
[signal SIGSEGV: segmentation violation code=0xffffffff addr=0x0 pc=0xc8527]`** 。  

接口的介绍到此结束。祝你愉快。

**下一教程 - [并发入门](#)**

---

via: https://golangbot.com/interfaces-part-2/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
