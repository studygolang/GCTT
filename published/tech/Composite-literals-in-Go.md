已发布：https://studygolang.com/articles/12913

# Go 中的复合字面量

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/composite-literal/1_TM61VTlvvL2YWtI6UUYLOg.png)

在源代码中字面量可以描述像数字，字符串，布尔等类型的固定值。Go 和 JavaScript、Python 语言一样，即便是复合类型（数组，字典，切片，结构体）也允许使用字面量。Golang 的复合字面量表达也很方便、简洁，使用单一语法即可实现。在 JavaScript 中它是这样的：

```javascript
var numbers = [1, 2, 3, 4]
var thing = {name: "Raspberry Pi", generation: 2, model: "B"}
```

在 Python 中有类似的语法：

```python
elements = [1, 2, 3, 4]
thing = {"name": "Raspberry Pi", "generation": 2, "model": "B"}
```

在 Go 中也类似（这里有些地方可以变通）：

```go
elements := []int{1, 2, 3, 4}
type Thing struct {
	name       string
	generation int
	model      string
}
thing := Thing{"Raspberry Pi", 2, "B"}
// 或者直接使用结构体的项名称
thing = Thing{name: "Raspberry Pi", generation: 2, model: "B"}
```

除了字典类型以外的其他类型，键是可选的，便于理解没有歧义：

* 对结构体而言，键就是项名称
* 对于数组或切片而言，键就是索引

键不是字面量常量，就必须是常量表达式；因此这种写法是错误的：

```go
f := func() int { return 1 }
elements := []string{0: "zero", f(): "one"}
```

这将导致编译异常 -- "index must be non-negative integer constant"。而常量表达式或者字面量是合法的：

```go
elements := []string{0: "zero", 1: "one", 4 / 2: "two"}
```

编译一切顺利。

重复的键是不允许的：
```go
elements := []string{
	0:     "zero",
	1:     "one",
	4 / 2: "two",
	2:     "also two"
}
```

在编译时会报出 "duplicate index in array literal: 2" 的异常信息。这也同样适用于结构体：

```go
type S struct {
	name string
}
s := S{name: "Michał", name: "Michael"}
```

编译结果是 "duplicate field name in struct literal: name" 的错误。

相应的字面量必须被赋值给相应的键，元素或结构体的项。更多关于可赋值性的内容可查看 ["Go 语言中的可赋值性"](https://studygolang.com/articles/12381)一文。

## 结构体

对于结构体类型定义的项，这里有两三个创建实例时的规定。

像下面的代码片段，结构体的定义必须指定内部项的名称，并且如果使用了那些定义以外的名称，那么在编译时会发生错误："unknown S field ‘name’ in struct literal"：

```go
type S struct {
	age int8
}
s := S{name: "Michał"}
```

如果最先的字面量有对应的键，那么之后的字面量也必须有对应的键，下面这样写是不合理的
：

```go
type S struct {
	name string
	age int8
}
s := S{name: "Michał", 29}
```

像这样，编译器会抛出异常"mixture of field:value and value initializers"，可以通过省略结构体中所有元素对应的键来更正它。

```go
s := S{"Michał", 29}
```

但这样又有一个附加限制：字面量的依次顺序必须与结构体定义时各项的顺序保持一致。
必须使用键：值或值的形式初始化结构体，并不代表我们必须赋值结构体中的每一项。被忽略的项将会默认赋值为该项类型对应的零值：

```go
type S struct {
	name string
	age int8
}
s := S{name: "Michał"}
fmt.Printf("%#v\n", s)
```

输出：

```go
main.S{name:"Michał", age:0}
```

只有使用键：值的形式初始化结构体时，才会有默认赋零值的操作：

```go
s := S{"Michał"}
```

这种写法是不能编译通过的，会抛出异常"too few values in struct initializer"。当有新的项被添加到结构体的一个被字面量赋值的项之前时，这种报错的做法对程序员更加安全--如果这个结构体中，一个名为"title"的字符串类型的项被加到了"name"项之前，那么值"Michał"将被认为是一个"title"，而这个问题是很难被排查出来的。
如果结构体字面量是空值，那么结构体内每一项都会被赋于零值：

```go
type Employee struct {
	department string
	position   string
}
type S struct {
	name string
	age  int8
	Employee
}
main.S{name:"", age:0, Employee:main.Employee{department:"", position:""}}
```

最后一个规定，结构体的赋值与[导出标示](https://studygolang.com/articles/12809)有关（简而言之，字面量不允许给非导出项赋值）

## 数组和切片

数组或者切片的元素都是被索引的，所以在字面上键必须是整数常量表达式。对于没有键的元素，键将会被赋值为前一个元素索引加一。在字面上，第一个元素的键（索引）如果没有被赋值，默认设为零。

```go
numbers := []string{"a", "b", 2 << 1: "c", "d"}
fmt.Printf("%#v\n", numbers)
[]string{"a", "b", "", "", "c", "d"}
```

赋值元素的数量可以小于数组的长度（被忽略的元素将会被赋值为零）：

```go
fmt.Printf("%#v\n", [3]string{"foo", "bar"})
[3]string{"foo", "bar", ""}
```

不允许对超出范围的索引赋值，所以下面几行代码是无效的：

```
[1]string{"foo", "bar"}
[2]string{1: "foo", "bar"}
```

可以通过使用三个点（...）的快捷符号来省去程序员声明数组长度的工作，编译器会通过索引最大值加一的方式获得它：

```go
elements := […]string{2: "foo", 4: "bar"}
fmt.Printf("%#v, length=%d\n", elements, len(elements))
```

输出：

```
[5]string{"", "", "foo", "", "bar"}, length=5
```

切片与之前数组内容基本一致：

```go
els := []string{2: "foo", 4: "bar"}
fmt.Printf("%#v, length=%d, capacity=%d\n", els, len(els), cap(els))
```

得出结果：

```
[]string{"", "", "foo", "", "bar"}, length=5, capacity=5
```

## 字典

除了将数组长度替换成了键类型以外，字典字面量的语法和数组非常类似。

```go
constants := map[string]float64{"euler": 2.71828, "pi": .1415926535}
```

## 快捷方式

如果作为字典键或者数组、切片、字典元素的字面量的类型与键或元素的类型一致，那么为了简洁，这个类型可以省略不写：

```go
coords := map[[2]byte]string{{1, 1}: "one one", {2, 1}: "two one"}
type Engineer struct {
	name string
	age  byte
}
engineers := [...]Engineer{{"Michał", 29}, {"John", 25}}
```

同样，如果键或元素是指针类型的话，&T也可以被省略：

```go
engineers := […]*Engineer{{"Michał", 29}, {"John", 25}}
fmt.Printf("%#v\n", engineers)
```

输出：

```
[2]*main.Engineer{(*main.Engineer)(0x8201cc1e0), (*main.Engineer)(0x8201cc200)}
```
## 资源

https://golang.org/ref/spec#Composite_literals

---

via: https://medium.com/golangspec/composite-literals-in-go-10dc62eec06a

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[yiyulantian](https://github.com/yiyulantian)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
