已发布：https://studygolang.com/articles/12700

# 自动代码生成的 5 点建议

`//go:generate` 的引入使得 Go 语言在构建过程中集成自动代码生成工具更加简单。`stringer` 使得编写重复代码更轻松，而 `yacc` 和 `ragel` 这类程序则让优化解析器的生成变得可能。在 [GoGenerateTools](https://github.com/golang/go/wiki/GoGenerateTools) 上你可以找到关于这类工具的一份不完整的列表。

**给自动生成的代码做标记**。为了让构建工具能够识别出自动生成的代码，必须使用一个符合下列正则表达式的注释：

```
^// Code generated .* DO NOT EDIT\.$
```

这段文字必须位于经过格式化后的 Go 文件注释的第一行。并且这段注释必须位于所有的 `/* */` 注释和 `package` 语句之前，但不能依附于 `package` 语句之上。这和 build 标签的规则类似。带有这类头部格式的文件会被 Go Lint 工具忽略，并且在 GitHub 的 PR 和 diff 中也会被默认折叠起来。当然，如果你能在其中指明生成这段代码的工具和参数就更好了。

**在需要的地方使用 `_`**。通过将可能会用到的包，方法，包级别变量以及实际并未用到的变量指定给 `_` 可以简化你的代码生成步骤。

**gofmt 你的生成代码**。格式过的代码会使得人们在调试问题的时候阅读代码更加方便。同样地，因为编辑器可以在文件保存之后进行 gofmt 以防止意外的空格变化，所以你也许需要在生成的代码上使用一些 linter 工具，至少为了检查正确性。Lint 警告是可以接受的（虽然不鼓励），因为机器生成的代码总是没有人写的那么标准。

**保证输出的确定性**。在不改变输入的情况下，多次运行工具应该有相同的输出。不要依赖于 map 排序或者在输出中添加不确定性。而对于使用时间戳我有不同的看法。虽然能够知道文件的生成时间的确很好，但是当文件没有变化的时候，也不应该有任何输出。

**保持 diffs 的可读性**。这和上一条有息息相关。人们往仓库中提交代码。如果他们生成了一个新的文件，则当有更新的时候能够看到文件的变化和它对生成代码的影响是非常好的。

在我自己使用代码自动生成工具的时候碰到了一些开心的事（或者说更多是沮丧的事）。那么你希望代码自动生成工具应该怎么样呢？

---

via: https://medium.com/@dgryski/five-nice-things-for-machine-generated-code-5335e67c1e36

作者：[Damian Gryski](https://medium.com/@dgryski)
译者：[alfred-zhong](https://github.com/alfred-zhong)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
