已发布：https://studygolang.com/articles/12809

# 在 Go 中导出标识符

包由单个目录内的源文件组成。 在这样的目录中，从不同的包中获取文件是非法的。 在 Go 中开始每个源文件的 Package 语句定义了文件所属的包：

```go
package foo
```

> package 语句不是引入新标识符的声明，因此以后在源文件中不能使用 "foo"。

包的名称具有类似于常规标识符的语法。 所有共享相同包名的文件形成包。

为了使用来自其他包裹的标识符，需要 import 声明：

```go
import "fmt"
```

在 `import` 关键字后指定的字符串称为导入路径。 它需要唯一标识一个包。 标准库中的软件包使用较短的导入路径，但一般的包，比如 `github.com/mlowicki/foo` 看起来会更长。

在上面的形式中，通过 package 声明中的包名来完成对导出的标识符的访问。 因此，除了识别属于包的文件外，它还将作为导入声明的默认包名。 通过导入路径之前的标识符可以覆盖它：

```go
import (
	f "fmt"
)
func main() {
	f.Println("whatever")
}
```

> 如 [Go 中的作用域](https://medium.com/@mlowicki/scopes-in-go-a6042bb4298c)中所述，软件包名称的作用域是文件块。

导入后并不是所有包的标识符都可以访问。 只有导出的标识符才可以访问，导出标识符必须遵守的两个规则才能从其他包中直接访问：

* 标识符的第一个字符是大写字母
* 要么在包块中定义标识符，要么是字段名称或 `method` 方法名称

## 包块的标识符

被定义在软件包块中意味着它被定义在任何功能之外，如：

```go
package library
var V = 1
type S struct {
	Name string
}
type I interface {
	M()
}
```

V，S 和 I 可用于具有适当导入语句的文件中：

```go
package main
import (
	"fmt"
	"github.com/mlowicki/library"
)
func main() {
	s := library.S{}
	fmt.Println(library.V, s)
}
```

## 导出的字段名称

字段名称还必须以大写字母开头，以便从其他包中访问：

```go
package library
type record struct {
	Name string
	age int8
}
func GetRecord() record {
	return record{Name: "Michał", age: 29}
}


package main
import (
	"fmt"
	"github.com/mlowicki/library"
)

func main() {
	record := library.GetRecord()
	fmt.Println(record.Name)
}
```

上面的代码可以正常工作，但尝试访问未导出的字段 `age`...

```go
fmt.Println(me.age)
```

编译时失败：

```
record.age undefined (cannot refer to unexported field or method age)
```

在库包中导出结构体，比如改为 Record，但这不会改变任何内容 - 即结构类型仍然不会导出 age 字段。

## 导出的 `method` 名称

与字段名称相同的规则适用于 `method`：

```go
package library

import "fmt"

type Duck interface {
	Quack()
	walk()
}

type Record struct{}

func (Record) Quack() {
	fmt.Println("Quack")
}

func (Record) walk() {
	fmt.Println("walk")
}
func GetDuck() Duck {
	return Record{}
}

package main

import (
	"github.com/mlowicki/library"
)

func main() {
	duck := library.GetDuck()
	duck.Quack()
	record := library.Record{}
	record.Quack()
}
```

输出：

```bash
> ./bin/sandbox
Quack
Quack
```

调用方法 `walk` 是非法的：

```go
duck.walk()
```

输出：

```bash
duck.walk undefined (cannot refer to unexported field or method walk)
```

或者：

```go
record.walk()
```

在编译时会报如下错误：

```bash
record.walk undefined (cannot refer to unexported field or method library.Record."".walk)
```

---

via: https://medium.com/golangspec/exported-identifiers-in-go-518e93cc98af

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[tingtingr](https://github.com/wentingrohwer)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
