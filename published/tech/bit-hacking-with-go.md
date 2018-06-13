已发布：https://studygolang.com/articles/12291

# Go 中的位运算

![cover](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-bits/cover.png)

在以前内存和处理能力（CPU）都是非常昂贵的，于是直接在位上编程就成为了处理信息的首选方式（在有些情况下也是唯一的方式）。如今，直接对位进行操作在底层系统、图像处理和密码学等领域还是至关重要的。

在 Go 语言中支持以下几种操作位的方式：

```
 &   位与
 |   位或
 ^   异或
&^   位与非
<<   左移
>>   右移
```

接下来我们会对每一个操作符进行详细的讨论并给出一些可以应用位操作的实例。

## `&` 操作符

在 Go 中，`&` 操作符用来在两个整数之间进行位 AND 运算。AND 操作有以下特性：

```
Given operands a, b
AND(a, b) = 1; only if a = b = 1
               else = 0
// 给定 2 个操作数 a，b：
// 当且仅当 a 和 b 都为 1 时，操作 AND(a, b) 的结果为 1。
// 否则操作 AND(a, b) 为 0。
```

AND 操作符是一个很好的将整数的指定位清零的方式。在下面的例子中，我们使用 `&` 运算符将数字后 4 位清零。

```go
func main() {
    var x uint8 = 0xAC    // x = 10101100
    x = x & 0xF0          // x = 10100000
}
```

所有的二进制操作符都支持简写形式，我们可以把上面的例子改为简写形式：

```go
func main() {
    var x uint8 = 0xAC    // x = 10101100
    x &= 0xF0             // x = 10100000
}
```

另外一个小技巧就是可以通过 `&` 来判断一个数字是奇数还是偶数。我们可以将数字和值 1 使用 `&` 做 AND 运算。如果结果是 1，那说明原来的数字是一个奇数。

```go
import (
    "fmt"
    "math/rand"
)
func main() {
    for x := 0; x < 100; x++ {
        num := rand.Int()
        if num&1 == 1 {
            fmt.Printf("%d is odd\n", num)
        } else {
            fmt.Printf("%d is even\n", num)
        }
    }
}
```
[在线运行](https://play.golang.org/p/2mTNOtioNM)

## `|` 操作符

`|` 用来做数字的位 OR 运算。OR 操作符有以下特性：

```
Given operands a, b
OR(a, b) = 1; when a = 1 or b = 1
              else = 0
// 给定两个操作数 a，b
// 当且仅当 a 和 b 均为 0 时，操作 OR(a, b) 返回 0。
// 否者返回 1。
```

我们可以使用这个特性来将一个整数中的指定位置为 1。在下面的例子里，我们使用 OR 运算将第 3、7、8 位置为 1。

```go
func main() {
    var a uint8 = 0
    a |= 196
    fmt.Printf("%b", a)
}
// prints 11000100
          ^^   ^
```
[在线运行](https://play.golang.org/p/3VPv4D83Oj)

当对一个数字使用掩码技术，OR 是非常有用的。下面的例子我们可以设置更多的位：

```go
func main() {
    var a uint8 = 0
    a |= 196
    a |= 3
    fmt.Printf("%b", a)
}
// prints 11000111
```
[在线运行](https://play.golang.org/p/7aJLwh3y4x)

在上面的例子中，我们不仅有数字 196 中的所有位，而且最后的两位也被数字 3 置 1。我们可以一直进行置 1 操作，直到所有的位都为 1。

### 使用位作为配置信息

现在，回顾 `AND(a,1) = a if and only if a = 1`。我们可以使用这个技巧来查询指定位上的值。例如 `a & 196` 将会返回 `196`，因为在 `a` 中 `196` 的所有位都被置 1 了。所以我们能够使用 OR 和 AND 来设置和读取配置信息的值。

 下面的代码完成了这个功能。函数 `procstr` 转换给定的字符串。它接收两个参数：第一个参数 `str` 是一个要被转换的字符串，第二个参数 `conf` 使用掩码指定转换时的配置信息。

```go
 const (
    UPPER  = 1 // upper case
    LOWER  = 2 // lower case
    CAP    = 4 // capitalizes
    REV    = 8 // reverses
)
func main() {
    fmt.Println(procstr("HELLO PEOPLE!", LOWER|REV|CAP))
}
func procstr(str string, conf byte) string {
    // reverse string
    rev := func(s string) string {
        runes := []rune(s)
        n := len(runes)
        for i := 0; i < n/2; i++ {
            runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
        }
        return string(runes)
    }

    // query config bits
    if (conf & UPPER) != 0 {
        str = strings.ToUpper(str)
    }
    if (conf & LOWER) != 0 {
        str = strings.ToLower(str)
    }
    if (conf & CAP) != 0 {
        str = strings.Title(str)
    }
    if (conf & REV) != 0 {
        str = rev(str)
    }
    return str
}
```
[在线运行](https://play.golang.org/p/4E05PQwj5q)

 调用 `procstr("HELLO PEOPLE!", LOWER|REV|CAP)` 将会把字符串转换成小写，反转并将每个单词的首字母转换成大写。当 `conf` 上的第 2、3、4 位为 1 时（conf 等于 14）将会执行上述操作。在内部我们使用 if 语句来取出这些位并且根据相应的配置操作字符串。

## `^` 操作符

XOR 操作符在 Go 中用 `^` 表示。XOR 是特例化的 OR，它有以下特性：

```go
Given operands a, b
XOR(a, b) = 1; only if a != b
     else = 0
// 给定 2 个操作数 a，b
// 当且仅当 a!=b 时，操作 XOR(a, b) 返回 1。
// 否者返回 0。
```

这就暗示了我们可以使用 XOR 来切换指定位上的值。例如，给定一个 16 位的值，我们可以使用下面的代码来切换它的前八位：

 ```go
 func main() {
    var a uint16 = 0xCEFF
    a ^= 0xFF00 // same a = a ^ 0xFF00
}
// a = 0xCEFF   (11001110 11111111)
// a ^=0xFF00   (00110001 11111111)
 ```

 在之前的代码中，位的值通过 XOR 操作在 0 和 1 之间切换。XOR 的一个实际的用途就是比较两个数字正负号是否相同。两个数字 `a`、`b`，如果 `(a ^ b) >= 0` 那么 a 和 b 同号，如果 `(a ^ b) < 0` 那么 a 和 b 异号。

```go
func main() {
    a, b := -12, 25
    fmt.Println("a and b have same sign?", (a ^ b) >= 0)
}
```
[在线运行](https://play.golang.org/p/6rAPti5bXJ)

 当上述代码执行会输出 `a and b have same sign? false`。使用 Go Playground 可以修改不同的符号查看不同的结果。

### 使用 `^` 作为位非操作

 不像其它语言（C/C++、Java、Python、Javascript 等），Go 没有一元运算符。XOR 操作符 `^` 可以作为一元操作符来计算一个数字的补码。给定一个位 `x`，在 Go 中 `^x = 1 ^ x` 将会反转 x 的位。我们可以通过 `^a` 来计算变量 `a` 的补码。

```go
func main() {
    var a byte = 0x0F
    fmt.Printf("%08b\n", a)
    fmt.Printf("%08b\n", ^a)
}

// prints
00001111     // var a
11110000     // ^a
```
[在线运行](https://play.golang.org/p/5d1fQjDAIv)

### `&^` 运算符

`&^` 运算符叫做 AND NOT。它是一个 使用 `AND` 后，再使用 `NOT` 操作的简写。该操作符定义如下：

```
Given operands a,b
AND_NOT(a, b) = AND(a, NOT(b))
// 给定两个操作数 a,b
// 当 a=NOT(b)=1 时，操作 AND_NOT(a, b) 返回 1。
// 否则返回 0。
```

它有一个有意思的特性：如果第二个操作符返回 1。那么该位将会被清 0。

```
AND_NOT(a, 1) = 0; clears a
AND_NOT(a, 0) = a;
```

下面这个代码片段使用 AND NOT 操作符来清掉 `a` 的后 4 位（`1010 1011` 到 `1010 0000`）。

```go
func main() {
    var a byte = 0xAB
    fmt.Printf("%08b\n", a)
    a &^= 0x0F
    fmt.Printf("%08b\n", a)
}
// prints:
10101011
10100000
```
[在线运行](https://play.golang.org/p/UPUlBOPRGh)

## `<<` 和 `>>` 运算符

和其它 C 家族语言一样，Go 使用 `<<` 和 `>>` 来代表左移或者右移运算，定义如下：

```
Given integer operands a and n,
a << n; 将 a 中的所有位向左偏移 n 次
a >> n; 将 a 中的所有位向右偏移 n 次
```

例如：在下面的片段中，`a` 使用左移运算符（`00000011`）左移 3 次。每次的结果都会被打印出来。

```go
func main() {
    var a int8 = 3
    fmt.Printf("%08b\n", a)
    fmt.Printf("%08b\n", a<<1)
    fmt.Printf("%08b\n", a<<2)
    fmt.Printf("%08b\n", a<<3)
}
// prints:
00000011
00000110
00001100
00011000
```

需要注意的是，每次左移右边的位都会被 0 填充。相反的，使用右移运算符时左边的位都会被 0 填充（有符号数字除外，具体请看之后的 *Arithmetic Shifts* 章节）。

```go
func main() {
    var a uint8 = 120
    fmt.Printf("%08b\n", a)
    fmt.Printf("%08b\n", a>>1)
    fmt.Printf("%08b\n", a>>2)
}
// prints:
01111000
00111100
00011110
```

一些简单的左移和右移的使用技巧就是乘法和除法，其中每次的位移都是乘或者除 2 次幂。下面的代码将 200 除以 2。

```go
func main() {
    a := 200
    fmt.Printf("%d\n", a>>1)
}
// prints:
100
```
[在线运行](https://play.golang.org/p/EJi0YCARun)

或者将一个值乘 4：

 ```go
 func main() {
    a := 12
    fmt.Printf("%d\n", a<<2)
}
// prints:
48
```
[在线运行](https://play.golang.org/p/xuJRcKgMVV)

位移运算符提供一种非常有趣的方式来设置一个二进制的值。我们使用 `|` 和 `<<` 来设置 `a` 第三位的值。

```go
func main() {
    var a int8 = 8
    fmt.Printf("%08b\n", a)
    a = a | (1<<2)
    fmt.Printf("%08b\n", a)
}
// prints:
00001000
00001100
```
[在线运行](https://play.golang.org/p/h7WoP7ieuI)

或者使用 `&` 和位移运算符来测试第 n 位是不是指定的值:

```go
func main() {
    var a int8 = 12
    if a&(1<<2) != 0 {
        fmt.Println("take action")
    }
}
// prints:
take action
```
[在线运行](https://play.golang.org/p/Ptc7Txk5Jb)

使用 `&^` 和位移运算来给第 n 位置 0：

```go
func main() {
    var a int8 = 13
    fmt.Printf("%04b\n", a)
    a = a &^ (1 << 2)
    fmt.Printf("%04b\n", a)
}
// prints:
1101
1001
```
[在线运行](https://play.golang.org/p/Stjq9oOjKz)

### 位移运算符注意事项

当移动的是一个有符号的值，Go 将会自动的适配位移运算。在向右位移时，正负位上的值将会填充在缺失的位上。

---

via: https://medium.com/learning-the-go-programming-language/bit-hacking-with-go-e0acee258827

作者：[Vladimir Vivien](https://medium.com/@vladimirvivien)
译者：[saberuster](https://github.com/saberuster)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

