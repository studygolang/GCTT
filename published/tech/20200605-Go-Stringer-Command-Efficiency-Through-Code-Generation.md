首发于：https://studygolang.com/articles/34521

# Go: stringer 命令，通过代码生成提高效率

![由 Renee French 创作的原始 Go Gopher 作品，为“ Go 的旅程”创作的插图。](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200605-Go-Stringer-Command-Efficiency-Through-Code-Generation/00.png)

ℹ️  这篇文章基于 Go 1.13。

`stringer` 命令的目标是自动生成满足 `fmt.Stringer` 接口的方法。它将为指定的类型生成 `String()` 方法， `String()` 返回的字符串用于描述该类型。

## 例子

这个[命令的文档](https://godoc.org/golang.org/x/tools/cmd/stringer) 中给我们提供了一个学习的例子，如下：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200605-Go-Stringer-Command-Efficiency-Through-Code-Generation/01.png)

输出如下：

```
1
```

产生的日志是一个常量的值，这可能会让你感到困惑。
让我们用命令 `stringer -type=Pill` 生成 `String()` 方法吧：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200605-Go-Stringer-Command-Efficiency-Through-Code-Generation/02.png)

生成了新的 `String()` 函数，运行当前代码时输出如下：

```
Aspirin
```

现在描述该类型的是一个字符串，而不是它实际的常量值了。
 `stringer` 也可以与 `go generate` 命令完美配合，使其功能更强大。只需在代码中添加以下指令即可：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200605-Go-Stringer-Command-Efficiency-Through-Code-Generation/03.png)

然后，运行 `go generate` 命令将会为你所有的类型自动生成新的函数。

## 效率

`stringer` 生成了一个包含每一个字符串的 ` 长字符串 ` 和一个包含每一个字符串索引的数组。在我们这个例子里，读 `Aspirin` 即是读 ` 长字符串 ` 的索引 7 到 13 组成的字符串：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200605-Go-Stringer-Command-Efficiency-Through-Code-Generation/04.png)

但是它有多快、多高效？我们来和另外两种方案对比一下：

* 硬编码 `String()` 函数

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200605-Go-Stringer-Command-Efficiency-Through-Code-Generation/05.png)

下面是一个包含 20 个常量的基准测试：

```
name                  time/op
Stringer-4            4.16ns ± 2%
StringerWithSwitch-4  3.81ns ± 1%
```

包含 100 个常量的基准测试：

```
name                  time/op
Stringer-4            4.96ns ± 0%
StringerWithSwitch-4  4.99ns ± 1%
```

常量越多，效率越高。这是有道理的。从内存中加载一个值比一些跳转指令（表示 if 条件的汇编指令）更具有膨胀性。

然而，switch 语句分支越多，跳转指令的数量就越多。从某种程度上来说，从内存中加载将会变得更有效。

* `String()` 函数输出一个 map

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200605-Go-Stringer-Command-Efficiency-Through-Code-Generation/06.png)

下面是一个包含 20 个常量的基准测试：

```
name                  time/op
Stringer-4             4.16ns ± 2%
StringerWithMap-4     28.60ns ± 2%
```

使用 map 要慢得多，因为它必须进行函数调用，并且在 bucket 中查找不像访问切片的索引那么简单。

想了解更多关于 map 的信息和内部结构，我建议你阅读我的文章 "[Go: Map Design by Code](https://medium.com/a-journey-with-go/go-map-design-by-code-part-ii-50d111557c08)"

## 自检器

在生成的代码中，有一些纯粹是为了校验的目的。下面是这些指令：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200605-Go-Stringer-Command-Efficiency-Through-Code-Generation/07.png)

`stringer` 将常量的名称与值一起写入每行。在本例中，`Aspirin` 的值为 `2`。更新常量的名称或其值会产生错误

* 更新名称但不重新生成 `String()` 函数：

```
./pill_string.go:12:8: undefined: Aspirin
```

* 更新值但不重新生成 `String()` 函数：

```
./pill_string.go:12:7: invalid array index Aspirin - 1 (out of bounds for 1-element array)
./pill_string.go:13:7: invalid array index Ibuprofen - 2 (index must be non-negative
```

然而，当我们添加一个新的常量的情况下 -- 这里下一个值为 `3`，并且不更新生成的文件，`stringer` 会有一个默认值：

```
Pill(3)
```

添加这个自检不会有任何影响，因为它在编译时被删除了。可以通过查看程序生成的 asm 代码来确认 :

```
➜  go tool compile -S main.go pill_string.go | grep "\"\".Pill\.[^\s]* STEXT"
"".Pill.String STEXT size=275 args=0x18 locals=0x50
```

只有 `String()` 函数被编译到二进制文件中 , 自检对性能或二进制大小没有影响。

---
via: https://medium.com/a-journey-with-go/go-stringer-command-efficiency-through-code-generation-df49f97f3954

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[kagxin](https://github.com/kagxin)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
