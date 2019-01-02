首发于：https://studygolang.com/articles/17400

# 深入理解 `fmt` 包

我们经常会使用 `fmt` 包，但是却没有思考过它的实现。我们会在这里使用一个 `fmt.Printf`，又会在那里使用一个 `fmt.Sprintf`。但是，如果你仔细研究下这个包，你会发现很多有趣有用的东西。

由于 Go 在大多数情况下会用来编写服务器或服务程序，我们主要的调试工具就是日志系统。`log` 包提供的 `log.Printf` 函数有和 `fmt.Printf` 相同的语义。
良好且信息丰富的日志消息非常重要，并且如果为你的数据结构添加一些格式化的支持将会为你的日志消息增加有价值的信息。

## 格式化输出

Go 的 `fmt` 相关的函数支持一些占位符，最常见的是字符串占位符的 `%s`，整型占位符 `%d`，以及浮点型占位符 `%f`。现在让我们探究一些其他的占位符。

### `%v` & `%T`

`%v` 占位符可以打印任何 Go 的值，`%T` 可以打印出变量的类型。我经常使用这些占位符来调试程序。

```go
var e interface{} = 2.7182
fmt.Printf("e = %v (%T)\n", e, e) // e = 2.7182 (float64)
```

### 宽度

你可以为一个打印的数值指定宽段，比如：

```go
fmt.Printf("%10d\n", 353)  // will print "       353"
```

你还可以通过将宽度指定为 `*` 来将宽度当作 `Printf` 的参数，例如：

```go
fmt.Printf("%*d\n", 10, 353)  // will print "       353"
```

当你打印出数字列表而且希望它们能够靠右对齐时，这非常的有用。

```go
// alignSize return the required size for aligning all numbers in nums
func alignSize(nums []int) int {
    size := 0
    for _, n := range nums {
        if s := int(math.Log10(float64(n))) + 1; s > size {
            size = s
        }
    }

    return size
}

func main() {
    nums := []int{12, 237, 3878, 3}
    size := alignSize(nums)
    for i, n := range nums {
        fmt.Printf("%02d %*d\n", i, size, n)
    }
}
```

将会打印出：

```bash
00   12
01  237
02 3878
03    3
```

这使得我们更加容易比较数字。

### 通过位置引用

如果你在一个格式化的字符串中多次引用一个变量，你可以使用 `%[n]`，其中 `n` 是你的参数索引（位置，从 1 开始）。

```go
fmt.Printf("The price of %[1]s was $%[2]d. $%[2]d! imagine that.\n", "carrot", 23)
```

这将会打印出：

```bash
The price of carrot was $23. $23! imagine that.
```

### `%v`

`%v` 占位符将会打印出 Go 的值，如果此占位符以 `+` 作为前缀，将会打印出结构体的字段名，如果以 `#` 作为前缀，那么它会打印出结构体的字段名和类型。

```go
// Point is a 2D point
type Point struct {
    X int
    Y int
}

func main() {
    p := &Point{1, 2}
    fmt.Printf("%v %+v %#v \n", p, p, p)
}
```

这将会打印：

```bash
&{1 2} &{X:1 Y:2} &main.Point{X:1, Y:2}
```

我经常会使用 `%+v` 这种占位符。

## `fmt.Stringer` & `fmt.Formatter`

有时候你希望能够精细化地控制你的对象如何被打印。例如，当向用户展示错误时，你可能需要的是一个字符串的表示，而当向日志系统写入时，你则希望是更加详细的字符串表示。

为了控制你的对象如何被打印，你可以实现 `fmt.Formatter` 接口，也可以选择实现 `fmt.Stringer` 接口。

使用 `fmt.Formatter` 接口比较好的例子是 `github.com/pkg/errors` 这个非常棒的库。假设你需要加载我们的配置文件，但是你有一个错误。你可以向用户打印一个简短的错误（又或者在 API 中返回），并在日志中输出更加详细的错误。

```go
cfg, err := loadConfig("/no/such/config.toml")
if err != nil {
    fmt.Printf("error: %s\n", err)
    log.Printf("can't load config\n%+v", err)
}
```

这就会向用户展示：

```bash
error: can't open config file: open /no/such/file.toml: no such file or directory
```

并且，日志文件中会这么记录：

```bash
2018/11/28 10:43:00 can't load config
open /no/such/file.toml: no such file or directory
can't open config file
main.loadConfig
    /home/miki/Projects/gopheracademy-web/content/advent-2018/fmt.go:101
main.main
    /home/miki/Projects/gopheracademy-web/content/advent-2018/fmt.go:135
runtime.main
    /usr/lib/go/src/runtime/proc.go:201
runtime.goexit
    /usr/lib/go/src/runtime/asm_amd64.s:1333
```

下面是一个小例子。假设你有一个 `AuthInfo` 结构体。

```go
// AuthInfo is authentication information
type AuthInfo struct {
    Login  string // Login user
    ACL    uint   // ACL bitmask
    APIKey string // API key
}
```

你希望此结构被打印时能够隐藏 `APIKey` 的值。你可以使用 `******` 来代替这个 `APIKey` 的值。

首先我们通过 `fmt.Stringer` 实现。

```go
// String implements Stringer interface
func (ai *AuthInfo) String() string {
    key := ai.APIKey
    if key != "" {
        key = keyMask
    }
    return fmt.Sprintf("Login:%s, ACL:%08b, APIKey: %s", ai.Login, ai.ACL, key)
}
```

现在 `fmt.Formatter` 获取到了占位符的 `fmt.State` 和符文。`fmt.State` 实现了 `io.Writer` 接口，使得你可以直接写入它。

想要了解结构体中所有的可用字段，可以使用 `reflect` 包。这将确保你的代码即使在 `AuthInfo` 字段更改后也能正常工作。

```go
var authInfoFields []string

func INIt() {
    typ := reflect.TypeOf(AuthInfo{})
    authInfoFields = make([]string, typ.NumField())
    for i := 0; i < typ.NumField(); i++ {
        authInfoFields[i] = typ.Field(i).Name
    }
    sort.Strings(authInfoFields) // People are better with sorted data
}
```

现在你就可以实现 `fmt.Formatter` 接口了。

```go
// Format implements fmt.Formatter
func (ai *AuthInfo) Format(state fmt.State, verb rune) {
    switch verb {
    case 's', 'q':
        val := ai.String()
        if verb == 'q' {
            val = fmt.Sprintf("%q", val)
        }
        fmt.Fprint(state, val)
    case 'v':
        if state.Flag('#') {
            // Emit type before
            fmt.Fprintf(state, "%T", ai)
        }
        fmt.Fprint(state, "{")
        val := reflect.ValueOf(*ai)
        for i, name := range authInfoFields {
            if state.Flag('#') || state.Flag('+') {
                fmt.Fprintf(state, "%s:", name)
            }
            fld := val.FieldByName(name)
            if name == "APIKey" && fld.Len() > 0 {
                fmt.Fprint(state, keyMask)
            } else {
                fmt.Fprint(state, fld)
            }
            if i < len(authInfoFields)-1 {
                fmt.Fprint(state, " ")
            }
        }
        fmt.Fprint(state, "}")
    }
}
```

现在让我们来试试看结果：

```go
ai := &AuthInfo{
    Login:  "daffy",
    ACL:    ReadACL | WriteACL,
    APIKey: "duck season",
}
fmt.Println(ai.String())
fmt.Printf("ai %%s: %s\n", ai)
fmt.Printf("ai %%q: %q\n", ai)
fmt.Printf("ai %%v: %v\n", ai)
fmt.Printf("ai %%+v: %+v\n", ai)
fmt.Printf("ai %%#v: %#v\n", ai)
```

这样这样打印出：

```bash
Login:daffy, ACL:00000011, APIKey: *****
ai %s: Login:daffy, ACL:00000011, APIKey: *****
ai %q: "Login:daffy, ACL:00000011, APIKey: *****"
ai %v: {3 ***** daffy}
ai %+v: {ACL:3 APIKey:***** Login:daffy}
ai %#v: *main.AuthInfo{ACL:3 APIKey:***** Login:daffy}
```

## 结论

除了一些琐碎的用途外，`fmt` 包还有许多其他的功能。一旦你熟悉了这些功能，我相信你会发现它们有很多有趣的用途。你可以在[此处](https://github.com/gopheracademy/gopheracademy-web/blob/master/content/advent-2018/fmt.go) 查看本文的代码。

---

via: https://blog.gopheracademy.com/advent-2018/fmt/

作者：[Miki Tebeka](https://blog.gopheracademy.com/advent-2018/fmt/)
译者：[barryz](https://github.com/barryz)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
