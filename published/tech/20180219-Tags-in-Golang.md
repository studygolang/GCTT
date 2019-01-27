首发于：https://studygolang.com/articles/17951

# Golang 中的标签（Tags in Golang）

结构体字段的声明可以通过之后放置的文字来标记。标签添加由当前包或外部包使用的元信息。让我们首先回首一下 strcut 声明的样子，然后我们将扔出几个用例，深入研究这个标签。

## 结构体类型（Struct type）

Struct 是一系列字段。每个字段由可选名称和所需类型（[源代码](https://play.golang.org/p/q2V_op8_SJk)）组成：

```go
package main
import "fmt"
type T1 struct {
	f1 string
}
type T2 struct {
	T1
	f2     int64
	f3, f4 float64
}
func main() {
	t := T2{T1{"foo"}, 1, 2, 3}
	fmt.Println(t.f1)    // foo
	fmt.Println(t.T1.f1) // foo
	fmt.Println(t.f2)    // 1
}
```

T1 域被称为嵌入字段，因为它是用类型声明但没有名称。

字段声明可以在 T2 中指定来自第 3 个字段声明的 f3 和 f4 之类的多个标识符。

语言规范声明每个字段声明后面跟着分号，但正如我们上面所见，它可以省略。如果需要将多个字段声明放入同一行（源代码），分号可能很有用（[源代码](https://play.golang.org/p/nTTgVX7BqV8)）：

```go
package main
import "fmt"
type T struct {
	f1 int64; f2 float64
}
func main() {
	t := T{1, 2}
	fmt.Println(t.f1, t.f2)  // 1 2
}
```

## 标签（Tag）

字段声明后面可以跟一个可选的字符串文字（标记），它称为相应字段声明中所有字段的属性（单字段声明可以指定多个标识符）。让我们看看它的实际应用（[源代码](https://play.golang.org/p/BubxnOxpOcM)）：

```go
type T struct {
	f1     string "f one"
	f2     string
	f3     string `f three`
	f4, f5 int64  `f four and five`
}
```

可以使用原始字符串文字或解释的字符串文字，但下面描述的传统格式需要原始字符串文字。[规范](https://golang.org/ref/spec#String_literals) 中描述了原始字符串文字和解释字符串文字之间的差异。

如果字段声明包含多个标识符，则标记将附加到字段声明的所有字段（如上面的字段 f4 和 f5）。

## 反射（Reflection）

标签可通过 reflect 包访问，允许运行时反射（[源代码](https://play.golang.org/p/YYEAfGc6iaE)）：

```go
package main
import (
	"fmt"
	"reflect"
)
type T struct {
	f1     string "f one"
	f2     string
	f3     string `f three`
	f4, f5 int64  `f four and five`
}
func main() {
	t := reflect.TypeOf(T{})
	f1, _ := t.FieldByName("f1")
	fmt.Println(f1.Tag) // f one
	f4, _ := t.FieldByName("f4")
	fmt.Println(f4.Tag) // f four and five
	f5, _ := t.FieldByName("f5")
	fmt.Println(f5.Tag) // f four and five
}
```

设置空标记与完全不使用标记的效果相同（[源代码](https://play.golang.org/p/u5VUMXz01cJ)）：

```go
type T struct {
	f1 string ``
	f2 string
}
func main() {
	t := reflect.TypeOf(T{})
	f1, _ := t.FieldByName("f1")
	fmt.Printf("%q\n", f1.Tag) // ""
	f2, _ := t.FieldByName("f2")
	fmt.Printf("%q\n", f2.Tag) // ""
}
```

## 惯用格式（Conventional format）

在提交中引入[“反射：支持多个包使用结构标记”](https://github.com/golang/go/commit/25733a94fde4f4af4b6ac5ba01b7212a3ef0f013)允许为每个包设置元信息。这提供了简单的命名空间。标签被格式化为键的串联：“值”对。密钥可能是像 JSON 这样的包的名称。对可以选择用空格分隔 - `key1: "value1" key2: "value2" key3: "value3"`。 如果使用传统格式，那么我们可以使用 struct tag（[StructTag](https://golang.org/pkg/reflect/#StructTag)）的两个方法 - Get 或 Lookup。它们允许返回与所需键内部标记相关联的值。

Lookup 函数返回两个值 - 与键关联的值（如果未设置则为空）和 bool，指示是否已找到键（[源代码](https://play.golang.org/p/i2Vh3tfN3A5)）：

```go
type T struct {
	f string `one:"1" two:"2"blank:""`
}
func main() {
	t := reflect.TypeOf(T{})
	f, _ := t.FieldByName("f")
	fmt.Println(f.Tag) // one:"1" two:"2"blank:""
	v, ok := f.Tag.Lookup("one")
	fmt.Printf("%s, %t\n", v, ok) // 1, true
	v, ok = f.Tag.Lookup("blank")
	fmt.Printf("%s, %t\n", v, ok) // , true
	v, ok = f.Tag.Lookup("five")
	fmt.Printf("%s, %t\n", v, ok) // , false
}
```

Get 方法只是 Lookup 简单的包封装器，它丢弃了 bool 值（[源代码](https://github.com/golang/go/blob/1b1c8b34d129eefcdbad234914df999581e62b2f/src/reflect/type.go#L1163)）：

```go
func (tag StructTag) Get(key string) string {
	v, _ := tag.Lookup(key)
	return v
}
```

> 如果标签不是常规模式，则不指定 Get 或 Lookup 的返回值。

即使 tag 是任何字符串值（不管是释义或原始值），只有在双引号（[源代码](https://play.golang.org/p/Gawmc_dpBDE)）之间包含值时，Lookup 和 Get 方法才会找到 key 的值：

```go
type T struct {
	f string "one:`1`"
}
func main() {
	t := reflect.TypeOf(T{})
	f, _ := t.FieldByName("f")
	fmt.Println(f.Tag) // one:`1`
	v, ok := f.Tag.Lookup("one")
	fmt.Printf("%s, %t\n", v, ok) // , false
}
```

可以在解释的字符串值中对双引号进行转义（[源代码](https://play.golang.org/p/o5APz18OH6e)）：

```go
type T struct {
	f string "one:\"1\""
}
func main() {
	t := reflect.TypeOf(T{})
	f, _ := t.FieldByName("f")
	fmt.Println(f.Tag) // one:"1"
	v, ok := f.Tag.Lookup("one")
	fmt.Printf("%s, %t\n", v, ok) // 1, true
}
```

但可读性就要低很多。

## 结论（Conversion）

将结构体类型转换为其他类型要求底层类型相同，但忽略掉 tag（[源代码](https://play.golang.org/p/C7M0bVwYFK2)）：

```go
type T1 struct {
	 f int `json:"foo"`
 }
 type T2 struct {
	 f int `json:"bar"`
 }
 t1 := T1{10}
 var t2 T2
 t2 = T2(t1)
 fmt.Println(t2) // {10}
```

Go 1.8 （提案）中引入了此行为。在 Go 1.7 及更早版本的代码中，可能会抛出编译时错误。

## 用例（Use cases）

### (Un)marshaling

Go 中标签最常见的用途可能是 [marshalling](https://en.wikipedia.org/wiki/Marshalling_%28computer_science%29)。让我们看一下来自 JSON 包的函数 [Marshal](https://golang.org/pkg/encoding/json/#Marshal) 如何使用它（[源代码](https://play.golang.org/p/C1hAMXTKPM_S)）：

```go
import (
	"encoding/json"
	"fmt"
)
func main() {
	type T struct {
	   F1 int `json:"f_1"`
	   F2 int `json:"f_2,omitempty"`
	   F3 int `json:"f_3,omitempty"`
	   F4 int `json:"-"`
	}
	t := T{1, 0, 2, 3}
	b, err := JSON.Marshal(t)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", b) // {"f_1":1,"f_3":2}
}
```

xml 包也利用了标签 - [https://golang.org/pkg/encoding/xml/#MarshalIndent](https://golang.org/pkg/encoding/xml/#MarshalIndent).

## ORM

像 GORM 这样的对象关系映射工具，也广泛使用标签 - [例子](https://github.com/jinzhu/gorm/blob/58e34726dfc069b558038efbaa25555f182d1f7a/multi_primary_keys_test.go#L10).

## 摘要数据（Digesting forms data）

[https://godoc.org/github.com/gorilla/schema](https://godoc.org/github.com/gorilla/schema)

## 其他（Other）

标签的更多潜在用例，如配置管理，结构的默认值，验证，命令行参数描述等（[众所周知的结构标记列表](https://github.com/golang/go/wiki/Well-known-struct-tags)）。

## Go vet

Go 编译器没有强制执行传统的 struct 标签格式，但是 vet 就是这样做的，所以值得使用它，例如作为 CI 管道的一部分。

```go
package main
type T struct {
	f string "one two three"
}
func main() {}
> Go vet tags.go
tags.go:4: struct field tag `one two three` not compatible with reflect.StructTag.Get: bad syntax for struct tag pair
```

...

由于 struct 标签，程序员可以从单一来源中受益。Go 是一门实用性语言，所以即使可以使用专用数据结构等其他方式来控制整个过程来解决 JSON/XML 编码，Golang 也能让软件工程师的生活变得更轻松。值得一提的是，标签的长度不受规格的限制。

---

via: https://medium.com/golangspec/tags-in-golang-3e5db0b8ef3e

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[gogeof](https://github.com/gogeof)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
