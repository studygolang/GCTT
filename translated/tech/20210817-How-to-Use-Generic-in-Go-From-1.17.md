# 如何在 Go 1.17 中使用范型

我们知道，Go 1.18 预计将在今年末或明年初发布时为 Go 语言带来范型。
但对于那些等不及的人， 可以从 [Go Playground](https://go2goplay.golang.org/) 的特定版本在线尝试范型，
还有一种方法可以让你在本地环境尝试范型，不过有点麻烦，
需要从 [源码](https://go.googlesource.com/go/+/refs/heads/dev.go2go/README.go2go.md) 编译 Go。

直到今天 Go 1.17 发布:

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210817-How-to-Use-Generic-in-Go-From-1.17/tweet-01.png)

除了一些新特性，还有一个特定的标记参数 `-gcflags=-G=3`，在编译或运行的时候加上它就能使用范型。
我第一次在这里看到他，但是除了这个来源，我还发现一些其他的公共消息。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210817-How-to-Use-Generic-in-Go-From-1.17/tweet-02.png)

总之，我很高兴确认它可以用！然后要说明的是，我是在 go2go playground 运行下面的代码：

```go
package main

import (
    "fmt"
)

// The playground now uses square brackets for type parameters. Otherwise,
// the syntax of type parameter lists matches the one of regular parameter
// lists except that all type parameters must have a name, and the type
// parameter list cannot be empty. The predeclared identifier "any" may be
// used in the position of a type parameter constraint (and only there);
// it indicates that there are no constraints.

func Print[T any](s []T) {
    for _, v := range s {
        fmt.Print(v)
    }
}

func main() {
    Print([]string{"Hello, ", "playground\n"})
}
```

当我尝试用刚刚提到的参数运行的时候，会导致下面的错误：

```bash
$ go1.17 run -gcflags=-G=3  cmd/generics/main.go

# command-line-arguments
cmd/generics/main.go:14:6: internal compiler error: Cannot export a generic function (yet): Print
No problem, lets unexport Printfor now, by renaming it to print.
```

现在运行相同的命令绝对可以正确执行!

```go
$ go1.17 run -gcflags=-G=3  cmd/generics/main.go

Hello, playground
```

## 这到底有什么用？

毫无疑问，这肯定是前进的一步。如果你想尝试范型，你必须从源码编译 Go。
然而，Go 编译器的实现也只是完成了一半工作，另外一半是工具链的支持。
根据我有限的测试，似乎只有 `run` 和 `build` 命令支持了这个参数，其他的命令，比如格式化或测试都没有成功。

随着 Go 1.18 越来越近，我很期待看到更多工具链的支持。

---
via: https://preslav.me/2021/08/17/how-to-use-generics-in-golang-starting-from-go1-17/

作者：[Preslav Rachev](https://preslav.me/author/preslav/)
译者：[h1z3y3](https://www.h1z3y3.me)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出