首发于：https://studygolang.com/articles/21207

# 效仿 Golang 中的枚举类型

在这篇博文中我们看到使用 `go generate` 和遍历抽象语法树来生成强大的枚举类型。

博文的结果是一个生成枚举类型的客户端。[全部代码](https://github.com/steinfletcher/gonum) 都可以在 Github 上面找到。

## Go 惯用技巧

Go 并没有对枚举类型提供一流的支持。模拟枚举类型的一种方法是，将一系列相关的常量定义为一个新的类型。Iota 可用于预定义连续自增的整形常量。我们可以像下面这样定义一个 `Color` 类型。

```go
package main

import "fmt"

type Color int

const (
    Red Color = iota // 0
    Blue             // 1
)

func main() {
    var b1 Color = Red
    b1 = Red
    fmt.Println(b1) // 打印 0

    var b2 Color = 1
    fmt.Println(b2 == Blue) // 打印 true

    var b3 Color
    b3 = 42
    fmt.Println(b3)  // 打印 42
}
```

value - we ’ ll need to convert the const to a display value in code.
这种模式在 Go 的代码中十分常见。虽然很常见但这个方法有其缺陷。因为没有静态语言检测，所以任意的整型都能作为 Color。没有序列化支持 - 开发者想要将其序列化为整型进行传输或者作为数据库记录，这是相当罕见的。没有可读的显示值支持 - 我们会需要在代码中将常量强转为显示值。

知道一门语言的习惯以及何时打破这些习惯是十分重要的。习惯用法的论据往往被用来关闭论点。这有时可能是创造力的死亡。

## 设计枚举类型

Go 最好的一个特性之一就是它的简便性 - 从其他语言转型而来的开发者通常可以非常快速的进行高效的开发。另一方面，这也带来了限制（译者注：作者想表达的应该是，某些其他语言支持泛型而 Golang 不支持，因而转到 Go 的开发者会受限），例如缺失能让代码变得整洁的泛型。为了克服这些缺点，社区已经将代码生成作为定义更为强大和灵活的类型的方案。

让我们用这个途径来定义枚举类型。其中一种做法是生成枚举结构体。我们还可以将方法附加到结构体中。结构体还提供了元标签，这对定义显示的值和描述很有帮助。

```go
type ColorEnum struct {
    Red  string `enum:"RED"`
    Blue string `enum:"BLUE"`
}
```

现在我们需要做的是为结构体的每个字段生成一个结构体实例
```go
var Red  = Color{name: "RED"}
var Blue = Color{name: "BLUE"}
```

然后我们可以对 Color 结构体增加方法以支持 JSON 编码 / 解码。我们实现 `Marshaler` 接口来提供 JSON 编码。

```go
func (c Color) MarshalJSON() ([]byte, error) {
    return JSON.Marshal(c.name)
}
```

Go 会在序列化这个类型为 JSON 的时候，调用我们定义的实现。同样，我们可以实现 `Unmarshaler` 接口，该接口使我们能够使用枚举类型——这允许我们直接在 API 中的数据传输对象上定义枚举类型。

```go
func (c *Color) UnmarshalJSON(b []byte) error {
    return JSON.Unmarshal(b, c.name)
}
```

我们还可以增加一些辅助方法来生成显示值的切片。

```go
// ColorNames 返回所有枚举实例的显示值的切片
func ColorNames() []string { ... }
```

我们也需要支持根据 string 生成枚举实例的方法，加上它。

```go
// NewColore 根据提供的显示值生成一个新的 Color
func NewColor(value string) (Color, error) { ... }
```

这种设计极具扩展性，你可能想要添加其他方法来返回名称，通过实现 `Error() string` 接口提供 errors，以及通过实现 `String() string` 支持 `Stringer`。

## 生成代码

### 遍历抽象语法树

在渲染模板生成代码之前，我们需要解析源码中的 `ColorEnum` 类型。两个常用的方法是使用 `refelct` 和 `ast` 包。我们需要扫描在包级别声明的结构体。`ast` 包拥有能力去构造抽象语法树 - 一种代表 Go 源码的可遍历数据结构。然后可以遍历抽象语法树并匹配提供的类型。这个类型和定义的结构体标签可以被解析并用于建立生成模板的模型。我们先加载一个 Go 的包

```go
cfg := &packages.Config{
    Mode:  packages.LoadSyntax,
    Tests: false,
}
pkgs, err := packages.Load(cfg, patterns...)
```

变量 `pkgs` 包含了这个包每个文件的抽象语法树。`ast.Inspect` 方法可用于遍历 AST( 译者注：抽象语法树 )，我们遍历每个文件，然后处理该文件的语法树。

```go
for _, file := range pkg.files {
...
    ast.Inspect(file.file, func(node ast.Node) bool {
        // 处理节点，检查是否是我们感兴趣的东西
    })
}
```

消费者应该定义自身的方法来过滤出它们所感兴趣的标志类型。你可以通过在节点上做以下校验来过滤结构体

```go
node.Tok == token.STRUCT { ... }
```

在我们的例子中，我们对定义 `enum:` 标签的 struct 进行过滤。我们简单对源码中的每一个标志进行处理，并根据碰到的数据构建模型（自定义 Go struct）。

### 渲染源码

有几个方法可以生成代码。工具[Stringer](https://github.com/golang/tools/blob/master/cmd/stringer/stringer.go) 使用 `fmt` 包将内容写到标准输出。虽然这很容易实现，但随着生成器的扩展，它变得难以操作且难以调试。更为合理的方法是使用 `text/template` 包并使用 Go 强大的模板库。它允许你从模板中分离生成模型的逻辑，从而导致将关注点和易于推理的代码分离开。（译者注：对比 stringer 源码之后就更精确地了解这句话的意思）生成的类型定义可能如下所示。

```go
// {{.NewType}} 是需要被创建的枚举实例
type {{.NewType}} struct {
    name  string
}

// 枚举实例
{{- range $e := .Fields}}
var {{.Value}} = {{$.NewType}}{name: "{{.Key}}"}
{{- end}}

... 生成方法的代码
```

然后我们可以根据我们的模型来渲染模板
```go
t, err := template.New(tmpl).Parse(tmpl)
if err != nil {
    log.Fatal("instance template parse error: ", err)
}

err = t.Execute(buf, model)
```

在开发模板的时候无需担心格式化就最好的。`format` 包存在将源码作为参数然后返回格式化后的 Go 代码的方法，所以让 Go 帮你处理这个东西吧。

```go
func Source(src []byte) ([]byte, error) { ... }
```

## 结论

在这篇博文中我们看到了解析 Go 源码生成枚举类型的方法。这个方法可作为需要解析源码的其他代码生成器的模板。我们以可维护的方式使用 Go 的 `text/template` 库来渲染源码。

在 Github 上阅读[所有的代码](https://github.com/steinfletcher/gonum)。

---

via: https://stein.wtf/posts/2019-04-16/enums/

作者：[Stein Fletcher](https://github.com/steinfletcher)
译者：[LSivan](https://github.com/LSivan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
