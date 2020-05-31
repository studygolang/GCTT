首发于：https://studygolang.com/articles/28991

# Go：如何利用符号表

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-hwo-to-take-symbol-table/cover.png)

> Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.

ℹ️ *本文基于 Go 1.13。*

符号表是由编译器生成和维护的，保存了与程序相关的信息，如函数和全局变量。理解符号表能帮助我们更好地与之交互和利用它。

## 符号表

Go 编译的所有二进制文件默认内嵌了符号表。我们来举一个例子并研究它。下面是代码：

```go
var AppVersion string

func main() {
	fmt.Println(`Version: `+AppVersion)
}
```

可以通过命令 `nm` 来展示符号表；下面是从 [OSX](https://www.unix.com/man-page/osx/1/nm/) 的结果中提取的部分信息：

```bash
0000000001177220 b io.ErrUnexpectedEOF
[...]
0000000001177250 b main.AppVersion
00000000010994c0 t main.main
[...]
0000000001170b00 d runtime.buildVersion
```

用 `b`（全称为 [bss](https://en.wikipedia.org/wiki/.bss)）标记的符号是未初始化的数据。由于我们前面的变量 `AppVersion` 没有初始化，因此它属于 `b`。符号 `d` 表示已初始化的数据，`t` 表示文本符号， 函数属于其中之一。

Go 也封装了 `nm` 命令，可以用命令 `go tool nm` 来使用它，也能生成相同的结果：

```bash
1177220 B io.ErrUnexpectedEOF
[...]
1177250 B main.AppVersion
10994c0 T main.main
[...]
1170b00 D runtime.buildVersion
```

当我们知道了暴露的变量的名字后，我们就可以与之交互。

## 自定义变量

当执行命令 `go build` 时，经过了两个阶段：编译和构建。构建阶段通过编译过程中生成的对象文件生成了一个可执行文件。为了实现这个阶段，构建器把符号表中的符号重定向到最终的二进制文件。

在 Go 中我们可以用 `-X` 来重写一个符号定义，`-X` 两个入参：名称和值。下面是承接前面的代码的例子：

```bash
go build -o ex -ldflags="-X main.AppVersion=v1.0.0"
```

构建并运行程序，现在会展示在命令行中定义的版本：

```bash
Version: v1.0.0
```

运行 `nm` 命令会看到变量已被初始化：

```bash
1170a90 D main.AppVersion
```

投建器赋予了我们重写数据符号（类型 `b` 或 `d`）的能力，现在它们有了 Go 中的 `string` 类型。下面是那些符号列表：

```bash
D runtime.badsystemstackMsg
D runtime.badmorestackgsignalMsg
D runtime.badmorestackg0Msg
B os.executablePath
B os.initCwd
B syscall.freebsdConfArch
D runtime/internal/sys.DefaultGoroot
B runtime.modinfo
B main.AppVersion
D runtime.buildVersion
```

在列表中我们看到了之前的变量和 `DefaultGoroot`，它们都是被构建器自动设置的。我们来看一下运行时这些符号的意义。

## 调试

符号表的存在是为了确保标识符在使用之前已被声明。这意味着当程序被构建后，它就不再需要这个表了。然而，默认情况下符号表是被嵌入到了 Go 的二进制文件以便调试。我们先来理解如何利用它，之后再来看怎么把它从二进制文件中删除。

我会用 `gdb` 来调试。只需要执行 `gdb ex` 就可以加载二进制文件。现在程序已被加载，我们用 `list` 命令来展示源码。下面是输出：

```bash
GNU gdb (GDB) 8.3.1
[...]
Reading symbols from ex...
Loading Go Runtime support.
(gdb) list 10
6
7  var AppVersion string
8
9  func main() {
10    fmt.Println(`Version: `+AppVersion)
11 }
12
(gdb)
```

`gdb` 初始化的第一步是读取符号表，为了提取程序中函数和符号的信息。我们现在可以用 `-ldflags=-s` 参数不把符号表编译进程序。下面是新的输出：

```bash
GNU gdb (GDB) 8.3.1
[...]
Reading symbols from ex...
(No debugging symbols found in ex)
(gdb) list
No symbol table is loaded.  Use the "file" command.
```

现在调试器由于找不到符号表不能展示源码。我们应该留意到使用 `-s` 参数去掉了符号表的同时，也去掉了对调试器很有用的 `[DWARF](https://golang.org/pkg/debug/dwarf/)` 调试信息。

## 二进制文件的大小

去掉符号表后会让调试器用起来很困难，但是会减少二进制文件的大小。下面是有无符号表的二进制文件的区别：

```bash
2,0M  7 f é v 15:59 ex
1,5M  7 f é v 15:22 ex-s
```

没有符号表比有符号表会小 25%。下面是编译 `cmd/go` 源码的另一个例子：

```bash
14M  7 f é v 16:58 go
11M  7 f é v 16:58 go-s
```

这里没有符号表和 DWARF 信息，也小了 25%。

*如果你想了解为什么二进制文件会变小，我推荐你阅读 WebKit 团队的 [Benjamin Poulain](https://twitter.com/awfulben) 的文章“[不寻常的加速：二进制文件大小](https://webkit.org/blog/2826/unusual-speed-boost-size-matters/)”。*

---

via: https://medium.com/a-journey-with-go/go-how-to-take-advantage-of-the-symbols-table-360dd52269e5

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
