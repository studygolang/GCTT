已发布：https://studygolang.com/articles/12263

# 第 16 部分：结构体

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 16 个教程。  

### 什么是结构体？

结构体是用户定义的类型，表示若干个字段（Field）的集合。有时应该把数据整合在一起，而不是让这些数据没有联系。这种情况下可以使用结构体。

例如，一个职员有 `firstName`、`lastName` 和 `age` 三个属性，而把这些属性组合在一个结构体 `employee` 中就很合理。

### 结构体的声明

```go
type Employee struct {
    firstName string
    lastName  string
    age       int
}
```

在上面的代码片段里，声明了一个结构体类型 `Employee`，它有 `firstName`、`lastName` 和 `age` 三个字段。通过把相同类型的字段声明在同一行，结构体可以变得更加紧凑。在上面的结构体中，`firstName` 和 `lastName` 属于相同的 `string` 类型，于是这个结构体可以重写为：

```go
type Employee struct {
    firstName, lastName string
    age, salary         int
}
```

上面的结构体 `Employee` 称为 **命名的结构体（Named Structure）**。我们创建了名为 `Employee` 的新类型，而它可以用于创建 `Employee` 类型的结构体变量。  

声明结构体时也可以不用声明一个新类型，这样的结构体类型称为 **匿名结构体（Anonymous Structure）**。

```go
var employee struct {
    firstName, lastName string
    age int
}
```

上述代码片段创建一个**匿名结构体** `employee`。

### 创建命名的结构体

通过下面代码，我们定义了一个**命名的结构体 `Employee`**。

```go
package main

import (  
    "fmt"
)

type Employee struct {  
    firstName, lastName string
    age, salary         int
}

func main() {

    //creating structure using field names
    emp1 := Employee{
        firstName: "Sam",
        age:       25,
        salary:    500,
        lastName:  "Anderson",
    }

    //creating structure without using field names
    emp2 := Employee{"Thomas", "Paul", 29, 800}

    fmt.Println("Employee 1", emp1)
    fmt.Println("Employee 2", emp2)
}
``` 
[在线运行程序](https://play.golang.org/p/uhPAHeUwvK)  

在上述程序的第 7 行，我们创建了一个命名的结构体 `Employee`。而在第 15 行，通过指定每个字段名的值，我们定义了结构体变量 `emp1`。字段名的顺序不一定要与声明结构体类型时的顺序相同。在这里，我们改变了 `lastName` 的位置，将其移到了末尾。这样做也不会有任何的问题。

在上面程序的第 23 行，定义 `emp2` 时我们省略了字段名。在这种情况下，就需要保证字段名的顺序与声明结构体时的顺序相同。

该程序将输出：

```
Employee 1 {Sam Anderson 25 500}
Employee 2 {Thomas Paul 29 800}
```

### 创建匿名结构体

```go
package main

import (
    "fmt"
)

func main() {
    emp3 := struct {
        firstName, lastName string
        age, salary         int
    }{
        firstName: "Andreah",
        lastName:  "Nikola",
        age:       31,
        salary:    5000,
    }

    fmt.Println("Employee 3", emp3)
}
```

[在线运行程序](https://play.golang.org/p/TEMFM3oZiq)  

在上述程序的第 3 行，我们定义了一个**匿名结构体变量** `emp3`。上面我们已经提到，之所以称这种结构体是匿名的，是因为它只是创建一个新的结构体变量 `em3`，而没有定义任何结构体类型。

该程序会输出：

```
Employee 3 {Andreah Nikola 31 5000}
```

### 结构体的零值（Zero Value）

当定义好的结构体并没有被显式地初始化时，该结构体的字段将默认赋为零值。

```go
package main

import (  
    "fmt"
)

type Employee struct {  
    firstName, lastName string
    age, salary         int
}

func main() {  
    var emp4 Employee //zero valued structure
    fmt.Println("Employee 4", emp4)
}
```

[在线运行程序](https://play.golang.org/p/p7_OpVdFXJ)  

该程序定义了 `emp4`，却没有初始化任何值。因此 `firstName` 和 `lastName` 赋值为 string 的零值（`""`）。而 `age` 和 `salary` 赋值为 int 的零值（0）。该程序会输出：

```
Employee 4 { 0 0}
```

当然还可以为某些字段指定初始值，而忽略其他字段。这样，忽略的字段名会赋值为零值。  

```go
package main

import (  
    "fmt"
)

type Employee struct {  
    firstName, lastName string
    age, salary         int
}

func main() {  
    emp5 := Employee{
        firstName: "John",
        lastName:  "Paul",
    }
    fmt.Println("Employee 5", emp5)
}
```

[在线运行程序](https://play.golang.org/p/w2gPoCnlZ1)  

在上面程序中的第 14 行和第 15 行，我们初始化了 `firstName` 和 `lastName`，而 `age` 和 `salary` 没有进行初始化。因此 `age` 和 `salary` 赋值为零值。该程序会输出：

```
Employee 5 {John Paul 0 0}
```

### 访问结构体的字段

点号操作符 `.` 用于访问结构体的字段。

```go
package main

import (  
    "fmt"
)

type Employee struct {  
    firstName, lastName string
    age, salary         int
}

func main() {  
    emp6 := Employee{"Sam", "Anderson", 55, 6000}
    fmt.Println("First Name:", emp6.firstName)
    fmt.Println("Last Name:", emp6.lastName)
    fmt.Println("Age:", emp6.age)
    fmt.Printf("Salary: $%d", emp6.salary)
}
```

[在线运行程序](https://play.golang.org/p/GPd_sT85IS)  

上面程序中的 **emp6.firstName** 访问了结构体 `emp6` 的字段 `firstName`。该程序输出：

```
First Name: Sam  
Last Name: Anderson  
Age: 55  
Salary: $6000  
```

还可以创建零值的 `struct`，以后再给各个字段赋值。

```go
package main

import (
    "fmt"
)

type Employee struct {  
    firstName, lastName string
    age, salary         int
}

func main() {  
    var emp7 Employee
    emp7.firstName = "Jack"
    emp7.lastName = "Adams"
    fmt.Println("Employee 7:", emp7)
}
```

[在线运行程序](https://play.golang.org/p/ZEOx10g7nN)  

在上面程序中，我们定义了 `emp7`，接着给 `firstName` 和 `lastName` 赋值。该程序会输出：  

```
Employee 7: {Jack Adams 0 0}
```

### 结构体的指针

还可以创建指向结构体的指针。

```go
package main

import (  
    "fmt"
)

type Employee struct {  
    firstName, lastName string
    age, salary         int
}

func main() {  
    emp8 := &Employee{"Sam", "Anderson", 55, 6000}
    fmt.Println("First Name:", (*emp8).firstName)
    fmt.Println("Age:", (*emp8).age)
}
```

[在线运行程序](https://play.golang.org/p/xj87UCnBtH)  

在上面程序中，**emp8** 是一个指向结构体 `Employee` 的指针。`(*emp8).firstName` 表示访问结构体 `emp8` 的 `firstName` 字段。该程序会输出：

```
First Name: Sam
Age: 55
```

**Go 语言允许我们在访问 `firstName` 字段时，可以使用 `emp8.firstName` 来代替显式的解引用 `(*emp8).firstName`**。  

```go
package main

import (  
    "fmt"
)

type Employee struct {  
    firstName, lastName string
    age, salary         int
}

func main() {  
    emp8 := &Employee{"Sam", "Anderson", 55, 6000}
    fmt.Println("First Name:", emp8.firstName)
    fmt.Println("Age:", emp8.age)
}
```
[在线运行程序](https://play.golang.org/p/0ZE265qQ1h)  

在上面的程序中，我们使用 `emp8.firstName` 来访问 `firstName` 字段，该程序会输出：  

```
First Name: Sam
Age: 55
```

### 匿名字段

当我们创建结构体时，字段可以只有类型，而没有字段名。这样的字段称为匿名字段（Anonymous Field）。  

以下代码创建一个 `Person` 结构体，它含有两个匿名字段 `string` 和 `int`。  

```go
type Person struct {  
    string
    int
}
```

我们接下来使用匿名字段来编写一个程序。  

```go
package main

import (  
    "fmt"
)

type Person struct {  
    string
    int
}

func main() {  
    p := Person{"Naveen", 50}
    fmt.Println(p)
}
```

[在线运行程序](https://play.golang.org/p/YF-DgdVSrC)  

在上面的程序中，结构体 `Person` 有两个匿名字段。`p := Person{"Naveen", 50}` 定义了一个 `Person` 类型的变量。该程序输出 `{Naveen 50}`。  

**虽然匿名字段没有名称，但其实匿名字段的名称就默认为它的类型**。比如在上面的 `Person` 结构体里，虽说字段是匿名的，但 Go 默认这些字段名是它们各自的类型。所以 `Person` 结构体有两个名为 `string` 和 `int` 的字段。  

```go
package main

import (  
    "fmt"
)

type Person struct {  
    string
    int
}

func main() {  
    var p1 Person
    p1.string = "naveen"
    p1.int = 50
    fmt.Println(p1)
}
```
[在线运行程序](https://play.golang.org/p/K-fGNxVyiA)  

在上面程序的第 14 行和第 15 行，我们访问了 `Person` 结构体的匿名字段，我们把字段类型作为字段名，分别为 "string" 和 "int"。上面程序的输出如下：  

```
{naveen 50}
```

### 嵌套结构体（Nested Structs）

结构体的字段有可能也是一个结构体。这样的结构体称为嵌套结构体。  

```go
package main

import (  
    "fmt"
)

type Address struct {  
    city, state string
}
type Person struct {  
    name string
    age int
    address Address
}

func main() {  
    var p Person
    p.name = "Naveen"
    p.age = 50
    p.address = Address {
        city: "Chicago",
        state: "Illinois",
    }
    fmt.Println("Name:", p.name)
    fmt.Println("Age:",p.age)
    fmt.Println("City:",p.address.city)
    fmt.Println("State:",p.address.state)
}
```

[在线运行程序](https://play.golang.org/p/46jkQFdTPO)  

上面的结构体 `Person` 有一个字段 `address`，而 `address` 也是结构体。该程序输出：  

```
Name: Naveen  
Age: 50  
City: Chicago  
State: Illinois  
```

### 提升字段（Promoted Fields）

如果是结构体中有匿名的结构体类型字段，则该匿名结构体里的字段就称为提升字段。这是因为提升字段就像是属于外部结构体一样，可以用外部结构体直接访问。我知道这种定义很复杂，所以我们直接研究下代码来理解吧。  

```go
type Address struct {  
    city, state string
}
type Person struct {  
    name string
    age  int
    Address
}
```

在上面的代码片段中，`Person` 结构体有一个匿名字段 `Address`，而 `Address` 是一个结构体。现在结构体 `Address` 有 `city` 和 `state` 两个字段，访问这两个字段就像在 `Person` 里直接声明的一样，因此我们称之为提升字段。

```go
package main

import (
    "fmt"
)

type Address struct {
    city, state string
}
type Person struct {
    name string
    age  int
    Address
}

func main() {  
    var p Person
    p.name = "Naveen"
    p.age = 50
    p.Address = Address{
        city:  "Chicago",
        state: "Illinois",
    }
    fmt.Println("Name:", p.name)
    fmt.Println("Age:", p.age)
    fmt.Println("City:", p.city) //city is promoted field
    fmt.Println("State:", p.state) //state is promoted field
}
```

[在线运行程序](https://play.golang.org/p/OgeHCJYoEy)  

在上面代码中的第 26 行和第 27 行，我们使用了语法 `p.city` 和 `p.state`，访问提升字段 `city` 和 `state` 就像它们是在结构体 `p` 中声明的一样。该程序会输出：

```
Name: Naveen  
Age: 50  
City: Chicago  
State: Illinois  
```

### 导出结构体和字段

如果结构体名称以大写字母开头，则它是其他包可以访问的导出类型（Exported Type）。同样，如果结构体里的字段首字母大写，它也能被其他包访问到。  

让我们使用自定义包，编写一个程序来更好地去理解它。  

在你的 Go 工作区的 `src` 目录中，创建一个名为 `structs` 的文件夹。另外在 `structs` 中再创建一个目录 `computer`。  

在 `computer` 目录中，在名为 `spec.go` 的文件中保存下面的程序。  

```go
package computer

type Spec struct { //exported struct  
    Maker string //exported field
    model string //unexported field
    Price int //exported field
}
```

上面的代码片段中，创建了一个 `computer` 包，里面有一个导出结构体类型 `Spec`。`Spec` 有两个导出字段 `Maker` 和 `Price`，和一个未导出的字段 `model`。接下来我们会在 main 包中导入这个包，并使用 `Spec` 结构体。

```go
package main

import "structs/computer"  
import "fmt"

func main() {  
    var spec computer.Spec
    spec.Maker = "apple"
    spec.Price = 50000
    fmt.Println("Spec:", spec)
}
```

包结构如下所示：  

```
src  
   structs
        computer
            spec.go
        main.go
```

在上述程序的第 3 行，我们导入了 `computer` 包。在第 8 行和第 9 行，我们访问了结构体 `Spec` 的两个导出字段 `Maker` 和 `Price`。执行命令 `go install structs` 和 `workspacepath/bin/structs`，运行该程序。  

如果我们试图访问未导出的字段 `model`，编译器会报错。将 `main.go` 的内容替换为下面的代码。  

```go
package main

import "structs/computer"  
import "fmt"

func main() {  
    var spec computer.Spec
    spec.Maker = "apple"
    spec.Price = 50000
    spec.model = "Mac Mini"
    fmt.Println("Spec:", spec)
}
```

在上面程序的第 10 行，我们试图访问未导出的字段 `model`。如果运行这个程序，编译器会产生错误：**spec.model undefined (cannot refer to unexported field or method model)**。  

### 结构体相等性（Structs Equality）

**结构体是值类型。如果它的每一个字段都是可比较的，则该结构体也是可比较的。如果两个结构体变量的对应字段相等，则这两个变量也是相等的**。

```go
package main

import (  
    "fmt"
)

type name struct {  
    firstName string
    lastName string
}


func main() {  
    name1 := name{"Steve", "Jobs"}
    name2 := name{"Steve", "Jobs"}
    if name1 == name2 {
        fmt.Println("name1 and name2 are equal")
    } else {
        fmt.Println("name1 and name2 are not equal")
    }

    name3 := name{firstName:"Steve", lastName:"Jobs"}
    name4 := name{}
    name4.firstName = "Steve"
    if name3 == name4 {
        fmt.Println("name3 and name4 are equal")
    } else {
        fmt.Println("name3 and name4 are not equal")
    }
}
```
[在线运行程序](https://play.golang.org/p/AU1FkdsPk7)  

在上面的代码中，结构体类型 `name` 包含两个 `string` 类型。由于字符串是可比较的，因此可以比较两个 `name` 类型的结构体变量。  

上面代码中 `name1` 和 `name2` 相等，而 `name3` 和 `name4` 不相等。该程序会输出：  

```
name1 and name2 are equal  
name3 and name4 are not equal  
```

**如果结构体包含不可比较的字段，则结构体变量也不可比较。**  

```go
package main

import (  
    "fmt"
)

type image struct {  
    data map[int]int
}

func main() {  
    image1 := image{data: map[int]int{
        0: 155,
    }}
    image2 := image{data: map[int]int{
        0: 155,
    }}
    if image1 == image2 {
        fmt.Println("image1 and image2 are equal")
    }
}
```
[在线运行程序](https://play.golang.org/p/T4svXOTYSg)  

在上面代码中，结构体类型 `image` 包含一个 `map` 类型的字段。由于 `map` 类型是不可比较的，因此 `image1` 和 `image2` 也不可比较。如果运行该程序，编译器会报错：**`main.go:18: invalid operation: image1 == image2 (struct containing map[int]int cannot be compared)`**。  

[github](https://github.com/golangbot/structs) 上有本教程的源代码。  

**下一教程 - 方法**

---

via: https://golangbot.com/structs/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
