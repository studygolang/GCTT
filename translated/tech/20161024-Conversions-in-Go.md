# Go 的类型转换

有时候你可能需要将变量转换为其他类型。Golang 不容许随意处理这种转换，转换是由类型系统的强制保证的某些规则。在这篇文章中，我们将讨论哪些转换是可能的，哪些是不可能，以及什么时候进行转换是有价值的。

Go 是一门强类型语言。它在类型上是严格的，编译期间会报告类型错误。

```go
package main
import "fmt"
func main() {
    monster := 1 + "2"
    fmt.Printf("monster: %v\n", monster)
}
```

```
> go build
# github.com/mlowicki/lab
./lab.go:6: cannot convert "2" to type int
./lab.go:6: invalid operation: 1 + "2" (mismatched types int and string)
```

JavaScript 是弱类型语言的一种，让我们看看它的实际效果:

```js
var monster = 1 + "foo" + function() {};
console.info("type:", typeof monster)
console.info("value:", monster);
```

我将数字，字符串甚至函数加在一起，这似乎是件奇怪的事情。但不用担心，JavaScript 会无报错地为你处理这些事情。

```
type: string
value: 1foofunction () {}
```

在特定的情况，可能需要将一个变量转为其他类型，例如将它作为函数参数传递，或是放进某个表达式中。

```go
func f(text string) {
    fmt.Println(text)
}
func main() {
    f(string(65))  // 整型常量转换为字符串。
}
```

函数调用的表达式是类型转换的常见情况, 下面我们会多次看到类似的转换。上面的代码是能够正常运行的，但是如果移除类型转换:

```go
f(65)
```

会导致一个编译时错误:"cannot use 65 (type int) as type string in argument to f"  （无法将整数类型的 65 作为字符串类型参数给 f ）

## 底层类型(Underlying type)

字符串，布尔值，数字或者字面量类型的底层类型仍是他们本身，其他情况下，类型声明定义了底层类型:

```go
type A string             // string
type B A                  // string
type C map[string]float64 // map[string]float64 (type literal)
type D C                  // map[string]float64
type E *D                 // *D
```

(注释里是对应的底层类型)
如果底层类型是相同的，那么类型转换时百分百有效的。

```go
package main
type A string
type B A
type C map[string]float64
type C2 map[string]float64
type D C
func main() {
    var a A = "a"
    var b B = "b"
    c := make(C)
    c2 := make(C2)
    d := make(D)
    a = A(b)
    b = B(a)
    c = C(c2)
    c = C(d)
    var _ map[string]float64 = map[string]float64(c)
}
```

上面的程序不会有任何的编译问题。**底层类型的定义不能是递归的（ Definition of underlying types isn’t recursive）**

**译按**:

>关于底层类型的定义不能是递归的情况，译者对这种说法保持怀疑。
>底层类型的定义要么解释到内置类型(int, int64, float, string, bool...), 要么递归解释到 unnamed type 的。例如
> * B->A->string, string 为内置类型，解释停止, B 的底层类型为 string;
> * U->T->map[S]float64, map[S]float64 为 unnamed type, 解释停止，U 的底层类型为 map[S]float64。
> ```
> type A string             // string
> type B A                  // string
> type S string             // string
> type T map[S]float64      // map[S]float64
> type U T                  // map[S]float64
> ```

```go
type S string
type T map[S]float64

....

var _ map[string]float64 = make(T)
```

上面的程序会在编译期间报错:

```
cannot use make(T) (type T) as type map[string]float64 in assignment
```

赋值错误的发生，是因为 *T* 的底层类型不是 `map[string]float64` 而是 `map[S]float64`。类型转换也会报错:

```go
var _ map[string]float64 = (map[string]float64)(make(T))
```

上面的代码会在编译时导致如下错误:

```
cannot convert make(T) (type T) to type map[string]float64
```

## 可赋值性(Assignability)

Go 语言规范提出*可赋值性(Assignability)*的概念，它定义了什么时候一个变量 *v* 能够被赋值给 *T* 类型的变量。让我们在代码中了解它的一个规则：赋值时，两者应该具有相同的底层类型，并且至少有一个不是 named 类型。

```go
package main
import "fmt"
func f(n [2]int) {
    fmt.Println(n)
}
type T [2]int
func main() {
    var v T
    f(v)
}
```

程序输出 “[0 0]”。在可赋值性规则许可的范围内，所有的类型转换都是可行的。程序员能用这种方式清晰地表达他的具体的想法:

```go
f([2]int(v))
```

上面的调用方法会和之前的得到一样的结果。关于可赋值性更多信息可以在[之前的文章](https://medium.com/golangspec/assignability-in-go-27805bcd5874)中找到。

> 类型转换的第一个规则(具有相同的底层类型)和可赋值规则的一条规则是重合的 - 当底层类型是相同的时候，至少有一个的类型是 unnmaed 类型(这一节的第一个例子)。较弱的规则会影响更严格的规则。因此当类型转换时，只需要底层类型保持一致，是否是 named/unnamed 类型并不重要。

## 常量

常量 *v* 能够被转换为类型 *T*, 当 *v* 能被 *T* 类型的变量表示时。

```go
a := uint32(1<<32 – 1)
//b := uint32(1 << 32) // constant 4294967296 overflows uint32
c := float32(3.4e38)
//d := float32(3.4e39) // constant 3.4e+39 overflows float32
e := string("foo")
//f := uint32(1.1) // constant 1.1 truncated to integer
g := bool(true)
//h := bool(1) // convert 1 (type int) to type bool
i := rune('ł')
j := complex128(0.0 + 1.0i)
k := string(65)
```

对于常量更深入的介绍可以在[官方博客](https://blog.golang.org/constants)中找到。

## 数字类型

### 浮点数(floating-point number) -> 整数(integer)

```go
var n float64 = 1.1
var m int64 = int64(n)
fmt.Println(m)
```

小数部分被移除，因此代码输出 ”1“。

对于其他转换:
* 浮点数 -> 浮点数，
* 整数 -> 整数，
* 整数 -> 浮点数，
* 复数 -> 复数。

变量会被四舍五入至目标精度:

```go
var a int64 = 2 << 60
var b int32 = int32(a)
fmt.Println(a)         // 2305843009213693952
fmt.Println(b)         // 0
a = 2 << 30
b = int32(a)
fmt.Println(a)         // 2147483648
fmt.Println(b)         // -2147483648
b = 2 << 10
a = int64(b)
fmt.Println(a)         // 2048
fmt.Println(b)         // 2048
```

## 指针

*可赋值性(Assignability)* 以和处理其他类型一样的方式处理指针类型。

```go
package main
import "fmt"
type T *int64
func main() {
    var n int64 = 1
    var m int64 = 2
    var p T = &n
    var q *int64 = &m
    p = q
    fmt.Println(*p)
}
```

程序正常工作并且输出 ”2“，这依赖于已经被讨论过的*可赋值性规则(assignability rule)*。**int64* 和 *T* 的底层类型是相同的，并且 **int64* 是 unnamed 类型。类型转换则更宽松一些。对于 unnamed 指针类型，指针的基类型(base type)具有相同的底层类型即可转换。

> 译按: 指针的基类型(base type)为指针所指向变量的类型，例如 p *int, p 的基类型为 int。

```go
package main
import "fmt"
type T int64
type U W
type W int64
func main() {
    var n T = 1
    var m U = 2
    var p *T = &n
    var q *U = &m
    p = (*T)(q)
    fmt.Println(*p)
}
```

> **T* 应该在括号内，否则他会被理解成 *(T(q))

和之前的程序一样，输出 “2”。因为 *U* 和 *T* 的在作为**U* 和 **T* 的基类型的同时，他们的底层类型是相同的。 如下的赋值操作:

```go
p = q
```

是不会成功的，因为它尝试处理两种不同的底层类型: **T* 和 **U*。作为练习，让我们稍微改变声明,看会发生什么

```go
type T *int64
type U *W
type W int64
func main() {
    var n int64 = 1
    var m W = 2
    var p T = &n
    var q U = &m
    p = T(q)
    fmt.Println(*p)
}
```

*U* 和 *W* 的声明已经被改变。思考一下，会发生什么?

编译器在以下位置报一个错误 “cannot convert q (type U) to type T”(无法将 q (类型 U) 转换为类型 T):

```go
p = T(q)
```

这是因为 *p* 的底层类型是 **int64*，而 *q* 的是 **W*。q 的类型是 named(*U*)，因此获取指针基类型的底层类型的规则并不适用在这里。

## 字符串

### 整数 -> 字符串

传递数字 *N* 到 *string* 内置地将 N 转化为 UTF-8 编码的字符串，该字符串是 N 表达的字符组成。

```go
fmt.Printf("%s\n", string(65))
fmt.Printf("%s\n", string(322))
fmt.Printf("%s\n", string(123456))
fmt.Printf("%s\n", string(-1))
```

输出:

```
A
ł
�
�
```

前两个转换使用完全有效的码位。也许你会好奇为什么后两行显示奇怪的符号。它是一个*替换字符(replacement character)*，是称为 *specials* Unicode 区段的一员。它的编码为 \uFFFD ( [更多信息](https://en.wikipedia.org/wiki/Specials_%28Unicode_block%29))

### 对 strings 的简要介绍

Strings 基本上是字节的切片:

```go
text := "abł"
for i := 0; i < len(text); i++ {
    fmt.Println(text[i])
}

```

输出:

```
97
98
197
130
```

97 和 98 是 UTF-8 编码的 “a” 和 “b“ 字符。第三和第四行的输出是字符 “ł” 的 UTF8 编码，该编码占据了两个字节的空间。

*range* 循环有助于迭代 Unicode 定义的码位( 码位在 Golang 中被称为 *rune* )

```go
text := "abł"
for _, s := range text {
    fmt.Printf("%q %#v\n", s, s)
}
```

输出:

```
'a' 97
'b' 98
'ł' 322
```

> 想了解更多类似 *%q* 和 *%v* 这样的占位符，可以看 [fmt](https://golang.org/pkg/fmt/) 包的文档

更多的讨论可在 [《Golang的字符串，字节，rune 和字符》](https://blog.golang.org/strings)。在这个快速解释之后，在字符串和字节切片之间的转换应该不会再难以理解。

### string ↔ slice of bytes

```go
bytes := []byte("abł")
text := string(bytes)
fmt.Printf("%#v\n", bytes) // []byte{0x61, 0x62, 0xc5, 0x82}
fmt.Printf("%#v\n", text)  // "abł"
```

切片由被转换 string 的 utf8 编码字节组成。

### string ↔ slice of runes

```go
runes := []rune("abł")
fmt.Printf("%#v\n", runes)         // []int32{97, 98, 322}
fmt.Printf("%+q\n", runes)         // ['a' 'b' '\u0142']
fmt.Printf("%#v\n", string(runes)) // "abł"
```

从被转换 string 中创建的切片是由 Unicode 编码的码位( rune )组成。

## 结尾

如果你喜欢这篇文章，请关注[我](https://medium.com/golangspec)，以便你收到最新的文章推送消息。

## 资料

* [Golang 语言规范](https://golang.org/ref/spec#Conversions)
* [Go 的 Strings, bytes, runes 和字符](https://blog.golang.org/strings)
* [Go 的赋值性](https://medium.com/golangspec/assignability-in-go-27805bcd5874)
* [Go 的常量](https://blog.golang.org/constants)

----------------

via: https://medium.com/golangspec/conversions-in-go-4301e8d84067

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[magichan](https://github.com/magichan)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
