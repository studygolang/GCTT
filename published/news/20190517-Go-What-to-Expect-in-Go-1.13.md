首发于：https://studygolang.com/articles/20542

# Go 1.13 中值得期待的特性

Go 1.13 的开发周期在五月初就要结束了，为了准备好未来 Go 2 的新特性，[新的进程](https://blog.golang.org/go2-here-we-come) 已经正式启动，这个进程主要由社区来负责驱动。

只要不会带来向后不兼容的问题，每个 Go 2 的提议都有可能会在更早的版本发布出来。总体而言，每个提议都应该：

1. 解决的是对大多数人来说很重要的问题。
2. 对所有其它的用户产生的影响最小。
3. 提供一个清晰并易于理解的解决方案。

至于 Go 1.13，它计划将于 8 月份发布。

## 已经被接受且合并的提议

- [数字字面量语法（Number literals syntax）](https://go.googlesource.com/proposal/+/master/design/19308-number-literals.md)：

这能便于开发者以二进制、八进制或十六进制浮点数的格式定义数字：

`v := 0b00101101`， 代表二进制的 101101，相当于十进制的 45。

`v := 0o377`，代表八进制的 377，相当于十进制的 255。

`v := 0x1p-2`，代表十六进制的 1 除以 2²，也就是 0.25。

而且这个提议还允许我们用 `_` 来分隔数字，比如说：

`v := 123_456` 等于 123456。

- [time.parse 现在支持用指定年度的第几天来解析日期](https://github.com/golang/go/issues/25689)：

`2019-123 15:04:05` 其中 123 代表 2019 年的第 123 天

- 有符号的整数可以作为时间计算表达式的右值。[这项提议](https://github.com/golang/go/issues/19113) 让 Go 语言能够正确的处理有符号的整数。
- `time.Duration` 多了两个新的 helper 函数：`Microseconds()` 和 `Milliseconds()`。

## 已经被接受但尚未实现的提议

- `math/big` 包[的改进](https://github.com/golang/go/issues/29951)，比如添加了 `AddInt()` 方法。
- `math` 包添加了[新的常量](https://github.com/golang/go/issues/28538)：`MaxInt`、`MaxUint`、`MinInt` 和 `MinUint`。
- `go test` 命令有了[新的参数](https://github.com/golang/go/issues/22964) `-coverhtml`

你可以在[这个列表](https://github.com/golang/go/issues?utf8=✓&q=label%3AProposal-Accepted+milestone%3AGo1.13) 中查看所有已接受的提议。

## 还没被接受的提议

- 添加一个[调度器统计的 API](https://github.com/golang/go/issues/15490) 来监控 Goroutine 的调度器。你可以在[这个仓库](https://github.com/deft-code/proposal/blob/master/design/15490-schedstats.md) 看到更多细节信息。
- 让 [import cycle error](https://github.com/golang/go/issues/31011) 错误信息更加易于理解。
- 让[导入包过程中的错误信息](https://github.com/golang/go/issues/30723) 更易于理解。
- 通过一个新的参数，让 XML decoder 在遇到未知的字段时报错。同样的特性在前段时间[已经合并](https://go-review.googlesource.com/c/go/+/74830/) 到 `JSON decoder` 里面了。

到目前为止还有很多 [release blockers](https://github.com/golang/go/issues?q=is%3Aopen+milestone%3AGo1.13+label%3Arelease-blocker)。它们大都与代码的问题相关，也有一小部分是文档或者性能回归相关的问题。

我会尽力地更新本文，使得它能和 Github 上 issue 的进展保持一致。

---

via: https://medium.com/@blanchon.vincent/go-what-to-expect-in-go-1-13-de8ad96e8ee2

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[Alex-liutao](https://github.com/Alex-liutao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
