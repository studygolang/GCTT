# 简化 Go 中对 JSON 的处理

我的第一个 Go 工程需要处理一堆 JSON 测试固件并把 JSON 数据作为参数传给我们搭建的 API 处理。另一个团队为了给 API 提供语言无关的、可预期的输入和输出，创建了这些测试固件。

在强类型语言中，JSON 通常很难处理 —— JSON 类型有字符串、数字、字典和数组。如果你使用的语言是 javascript、python、ruby 或 PHP，那么 JSON 有一个很大的好处就是在解析和编码数据时你不需要考虑类型。

```bash
// in PHP
$object = json_decode('{"foo":"bar"}');

// in javascript
const object = JSON.parse('{"foo":"bar"}')
```

在强类型语言中，你需要自己去定义怎么处理 JSON 对象的字符串、数字、字典和数组。在 Go 语言中，你使用内建的 API 时需要考虑如何更好地把一个 JSON 文件转换成 Go 的数据结构。我不打算深入研究在 Go 中如何处理 JSON 这个复杂的话题，我只列出两个代码的例子来阐述下这个问题。源码详情请见 [Go 实例教程](https://gobyexample.com/json)

## 解析/序列化为 map[string]interface

首先，来看这个程序

```go
package main

import (
    "encoding/json"
    "fmt"
)


func main() {

    byt := []byte(`{
        "num":6.13,
        "strs":["a","b"],
        "obj":{"foo":{"bar":"zip","zap":6}}
    }`)
    var dat map[string]interface{}
    if err := json.Unmarshal(byt, &dat); err != nil {
        panic(err)
    }
    fmt.Println(dat)

    num := dat["num"].(float64)
    fmt.Println(num)

    strs := dat["strs"].([]interface{})
    str1 := strs[0].(string)
    fmt.Println(str1)

    obj := dat["obj"].(map[string]interface{})
    obj2 := obj["foo"].(map[string]interface{})
    fmt.Println(obj2)

}
```

我们把 JSON 数据从 byt 变量反序列化（如解析、解码等等）成名为 dat 的 map/字典对象。这些操作跟其他语言类似，不同的是我们的输入需要是字节数组（不是字符串），对于字典的每个值时需要有[类型断言](https://www.sohamkamani.com/golang/type-assertions-vs-type-conversions/)才能读取或访问该值。

当我们处理一个多层嵌套的 JSON 对象时，这些类型断言会让处理变得非常繁琐。



## 解析/序列化为 struct

第二种处理如下：

```go
package main

import (
    "encoding/json"
    "fmt"
)

type ourData struct {
    Num   float64 `json:"num"`
    Strs []string `json:"strs"`
    Obj map[string]map[string]string `json:"obj"`
}

func main() {
    byt := []byte(`{
        "num":6.13,
        "strs":["a","b"],
        "obj":{"foo":{"bar":"zip","zap":6}}
    }`)

    res := ourData{}
    json.Unmarshal(byt, &res)
    fmt.Println(res.Num)
    fmt.Println(res.Strs)
    fmt.Println(res.Obj)
}
```

我们利用 Go struct 的标签功能把 byt 变量中的字节反序列化成一个具体的结构 ourData。

标签是结构体成员定义后跟随的字符串。我们的定义如下：

```go
type ourData struct {
    Num   float64 `json:"num"`
    Strs []string `json:"strs"`
    Obj map[string]map[string]string `json:"obj"`
}
```

你可以看到 Num 成员的 json 标签 “num”、Str 成员的 json 标签 “strs”、Obj 成员的 json 标签 “obj”。这些字符串使用[反引号](https://golangbyexample.com/double-single-back-quotes-go/)把标签声明为文字串。除了反引号，你也可以使用双引号，但是使用双引号可能会需要一些额外的转义，这样看起来会很凌乱。

```go
type ourData struct {
    Num   float64 "json:\"num\""
    Strs []string "json:\"strs\""
    Obj map[string]map[string]string "json:\"obj\""
}
```

在 struct 的定义中，标签不是必需的。如果你的 struct 中包含了标签，那么它意味着 Go 的 [反射 API](https://pkg.go.dev/reflect) 可以[访问标签的值](https://stackoverflow.com/questions/23507033/get-struct-field-tag-using-go-reflect-package/23507821#23507821)。Go 中的包可以使用这些标签来进行某些操作。

Go 的 `encoding/json` 包在反序列化 JSON 成员为具体的 struct 时，通过这些标签来决定每个顶层的 JSON 成员的值。换句话说，当你定义如下的 struct 时：

```go
type ourData struct {
    Num   float64   `json:"num"`
}
```

意味着：

> 当使用 json.Unmarshal 反序列化 JSON 对象为这个 struct 时，取它顶层的 num 成员的值并把它赋给这个 struct 的 Num 成员。

这个操作可以让你的反序列化代码稍微简洁一点，因为程序员不需要对每个成员取值时都显式地调用类型断言。然而，这个仍不是最佳解决方案。

首先 —— 标签只对顶层的成员有效 —— 嵌套的 JSON 需要对应嵌套的类型（如 Obj map[string]map[string]string），因此繁琐的操作仍没有避免。

其次 —— 它假定你的 JSON 结构不会变化。如果你运行上面的程序，你会发现 `"zap":6` 并没有被赋值到 Obj 成员。你可以通过创建类型 `map[string]map[string]interface{}` 来处理，但是在这里你又需要进行类型断言了。

这是我第一个 Go 工程遇到的情况，曾让我苦不堪言。

幸运的是，现在我们有了更有效的办法。

## SJSON 和 GJSON

Go 内建的 JSON 处理并没有变化，但是已经出现了一些成熟的旨在用起来更简洁高效的处理 JSON 的包。

[SJSON](https://github.com/tidwall/sjson)（写 JSON）和 [GJSON](https://github.com/tidwall/gjson)（读 JSON）是 [Josh Baker](https://github.com/tidwall) 开发的两个包，你可以用来读写 JSON 字符串。你可以参考 README 来获取代码实例 —— 下面是使用 GJSON 从 JSON 字符串中获取嵌套的值的示例：

```go
package main

import "github.com/tidwall/gjson"

const json = `{"name":{"first":"Janet","last":"Prichard"},"age":47}`

func main() {
    value := gjson.Get(json, "name.last")
    println(value.String())
}

```

类似的，下面是使用 SJSON “设置” JSON 字符串中的值返回设置之后的字符串的示例代码：

```go
package main

import "github.com/tidwall/sjson"

const json = `{"name":{"first":"Janet","last":"Prichard"},"age":47}`

func main() {
    value, _ := sjson.Set(json, "name.last", "Anderson")
    println(value)
}
```

如果 SJSON 和 GJSON 不符合你的口味，还有[一些](https://github.com/pquerna/ffjson)[其他的](https://github.com/mailru/easyjson)[第三方库](https://github.com/Jeffail/gabs)，可以用来在 Go 程序中稍微复杂点地处理 JSON。



----------------

via: https://ewanvalentine.io/microservices-in-golang-part-3/

作者：[Ewan Valentine](http://ewanvalentine.io/author/ewan)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
