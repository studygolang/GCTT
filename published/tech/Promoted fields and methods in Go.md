已发布：https://studygolang.com/articles/12792

# Go 语言中提取字段和方法

struct 是一系列包含名称和类型的字段。通常就像这样：

```go
package main
import "fmt"
type Person struct {
	name string
	age int32
}
func main() {
	person := Person{name: "Michał", age: 29}
	fmt.Println(person)  // {Michał 29}
}
```

（在这篇博文的接下来部分，我将逐步删除包名、导入和主函数的定义）

上面的结构体中，每个字段都有明确的名字。Go 语言也允许不指定字段名称。没有名称的字段称为匿名字段或者内嵌字段。类型的名字（如果有包名，不包含这个前缀的包名）就作为字段的名字。因为结构体中要求至少有一个唯一的字段名，所以我们不能这样做：

```go
import (
	"net/http"
)
type Request struct{}
type T struct {
	http.Request // field name is "Request"
	Request // field name is "Request"
}
```

如果编译，编译器将抛出错误：

```
> go install github.com/mlowicki/sandbox
# github.com/mlowicki/sandbox
src/github.com/mlowicki/sandbox/sandbox.go:34: duplicate field Request
```

带有匿名的字段或者方法，可以用一个简洁的方式访问到：

```go
type Person struct {
	name string
	age int32
}
func (p Person) IsAdult() bool {
	return p.age >= 18
}
type Employee struct {
	position string
}
func (e Employee) IsManager() bool {
	return e.position == "manager"
}
type Record struct {
	Person
	Employee
}
...
record := Record{}
record.name = "Michał"
record.age = 29
record.position = "software engineer"
fmt.Println(record) // {{Michał 29} {software engineer}}
fmt.Println(record.name) // Michał
fmt.Println(record.age) // 29
fmt.Println(record.position) // software engineer
fmt.Println(record.IsAdult()) // true
fmt.Println(record.IsManager()) // false
```

匿名（嵌入）的字段和方法被调用时会自动向上提升（来找到对应的对象）。

他们的行为跟常规字段类似，但却不能用于结构体字面值：

```go
//record := Record{}
record := Record{name: "Michał", age: 29}
```

它将导致编译器抛出错误：

```
// src/github.com/mlowicki/sandbox/sandbox.go:23: unknown Record field ‘name’ in struct literal
// src/github.com/mlowicki/sandbox/sandbox.go:23: unknown Record field ‘age’ in struct literal
```

可以通过创建一个明确的，完整的，嵌入的结构体来达到我们的目的：

```go
Record{Person{name: "Michał", age: 29}, Employee{position: "Software Engineer"}}
```

[来源：](https://golang.org/ref/spec#Struct_types)
![](https://raw.githubusercontent.com/studygolang/gctt-images/master/promoted-fields-and-methods/promoted-fields-and-methods-in-go-1.jpg)

----------------

via: https://medium.com/golangspec/promoted-fields-and-methods-in-go-4e8d7aefb3e3

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[gogeof](https://github.com/gogeof)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
