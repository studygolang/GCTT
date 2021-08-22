已发布：https://studygolang.com/articles/11924

# 第九部分：循环

这是 Go 语言系列教程的第 9 部分。

循环语句是用来重复执行某一段代码。

`for` 是 Go 语言唯一的循环语句。Go 语言中并没有其他语言比如 C 语言中的 `while` 和 `do while` 循环。

## for 循环语法

```go
for initialisation; condition; post {
}
```

初始化语句只执行一次。循环初始化后，将检查循环条件。如果条件的计算结果为 `true` ，则 `{}` 内的循环体将执行，接着执行 post 语句。post 语句将在每次成功循环迭代后执行。在执行 post 语句后，条件将被再次检查。如果为 `true`, 则循环将继续执行, 否则 for 循环将终止。（译注：这是典型的 for 循环三个表达式，第一个为初始化表达式或赋值语句；第二个为循环条件判定表达式；第三个为循环变量修正表达式，即此处的 post ）

这三个组成部分，即初始化，条件和 post 都是可选的。让我们看一个例子来更好地理解循环。

## 例子

让我们用 `for` 循环写一个打印出从 1 到 10 的程序。

```go
package main

import (
    "fmt"
)

func main() {
    for i := 1; i <= 10; i++ {
        fmt.Printf(" %d",i)
    }
}
```

[Run in playground](https://play.golang.org/p/mV6Zgcx2DY "Run in playground")

在上面的程序中，i 变量被初始化为 1。条件语句会检查 i 是否小于 10。如果条件成立，i 就会被打印出来，否则循环就会终止。循环语句会在每一次循环完成后自增 1。一旦 i 变得比 10 要大，循环中止。

上面的程序会打印出 `1 2 3 4 5 6 7 8 9 10` 。

在 `for` 循环中声明的变量只能在循环体内访问，因此 i 不能够在循环体外访问。

## break

`break` 语句用于在完成正常执行之前突然终止 for 循环，之后程序将会在 for 循环下一行代码开始执行。

让我们写一个从 1 打印到 5 并且使用 `break` 跳出循环的程序。

```go
package main

import (
    "fmt"
)

func main() {
    for i := 1; i <= 10; i++ {
        if i > 5 {
            break //loop is terminated if i > 5
        }
        fmt.Printf("%d ", i)
    }
    fmt.Printf("\nline after for loop")
}
```
[Run in playground](https://play.golang.org/p/sujKy92f-- "Run in playground")

在上面的程序中，在循环过程中 i 的值会被判断。如果 i 的值大于 5 然后 `break` 语句就会执行，循环就会被终止。打印语句会在 `for` 循环结束后执行，上面程序会输出为

```
1 2 3 4 5
line after for loop
```

## continue

`continue` 语句用来跳出 `for` 循环中当前循环。在 `continue` 语句后的所有的 `for` 循环语句都不会在本次循环中执行。循环体会在一下次循环中继续执行。

让我们写一个打印出 1 到 10 并且使用 `continue` 的程序。

```go
package main

import (
    "fmt"
)

func main() {
    for i := 1; i <= 10; i++ {
        if i%2 == 0 {
            continue
        }
        fmt.Printf("%d ", i)
    }
}
```

[Run in playground](https://play.golang.org/p/DRLN2ZHwVS "Run in playground")

在上面的程序中，这行代码 `if i%2==0` 会判断 i 除以 2 的余数是不是 0，如果是 0，这个数字就是偶数然后执行 `continue` 语句，从而控制程序进入下一个循环。因此在 `continue` 后面的打印语句不会被调用而程序会进入一下个循环。上面程序会输出 `1 3 5 7 9`。

## 更多例子

让我们写更多的代码来演示 `for` 循环的多样性吧

下面这个程序打印出从 0 到 10 所有的偶数。

```go
package main

import (
    "fmt"
)

func main() {
    i := 0
    for ;i <= 10; { // initialisation and post are omitted
        fmt.Printf("%d ", i)
        i += 2
    }
}
```
[Run in playground](https://play.golang.org/p/PNXliGINku "Run in playground")

正如我们已经知道的那样，`for` 循环的三部分，初始化语句、条件语句、post 语句都是可选的。在上面的程序中，初始化语句和 post 语句都被省略了。i 在 `for` 循环外被初始化成了 0。只要 `i<=10` 循环就会被执行。在循环中，i 以 2 的增量自增。上面的程序会输出 `0 2 4 6 8 10`。

上面程序中 `for` 循环中的分号也可以省略。这个格式的 `for` 循环可以看作是二选一的 `for while` 循环。上面的程序可以被重写成：

```go
package main

import (
    "fmt"
)

func main() {
    i := 0
    for i <= 10 { //semicolons are ommitted and only condition is present
        fmt.Printf("%d ", i)
        i += 2
    }
}
```

[Run in playground](https://play.golang.org/p/UYiz-Wtnoj "Run in playground")

在 `for` 循环中可以声明和操作多个变量。让我们写一个使用声明多个变量来打印下面序列的程序。

```
10 * 1 = 10
11 * 2 = 22
12 * 3 = 36
13 * 4 = 52
14 * 5 = 70
15 * 6 = 90
16 * 7 = 112
17 * 8 = 136
18 * 9 = 162
19 * 10 = 190
```

```go
package main

import (
    "fmt"
)

func main() {
    for no, i := 10, 1; i <= 10 && no <= 19; i, no = i+1, no+1 { //multiple initialisation and increment
        fmt.Printf("%d * %d = %d\n", no, i, no*i)
    }

}
```

[Run in playground](https://play.golang.org/p/e7Pf0UDjj0 "Run in playground")

在上面的程序中 `no` 和 `i` 被声明然后分别被初始化为 10 和 1 。在每一次循环结束后 `no` 和 `i` 都自增 1 。布尔型操作符 `&&` 被用来确保 i 小于等于 10 并且 `no` 小于等于 19 。

## 无限循环

无限循环的语法是：
```go
for {
}
```

下一个程序就会一直打印 `Hello World` 不会停止。

```go
package main

import "fmt"

func main() {
    for {
        fmt.Println("Hello World")
    }
}
```

如果你打算在 [go playground](https://play.golang.org/p/kYQZw1AWT4 "go playground") 里尝试上面的程序，你会得到一个“过程耗时太长”的错误。请尝试在你本地系统上运行，来无限的打印 “Hello World” 。

这里还有一个 `range` 结构，它可以被用来在 `for` 循环中操作数组对象。当我们学习数组时我们会补充这方面内容。

这就是 `for` 循环的全部内容。希望您能享受本次阅读。请留下您宝贵的意见和建议。

**上一教程 - [if-else 语句](https://studygolang.com/articles/11902)**

**下一教程 - [switch 语句](https://studygolang.com/articles/11957)**

---

via: https://golangbot.com/loops/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[thxallvu](https://github.com/thxallvu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
