# Go 中使用别名，简单且高效

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://raw.githubusercontent.com/studygolang/gctt-images2/master/Go-Aliases-Simple-and-Efficient/00.png)

ℹ️ 本文基于 Go 1.13。

Go 1.9 版本引入了别名，开发者可以为一个已存在的类型赋其他的名字。这个特性旨在促进大型代码库的重构，这对大型的项目至关重要。在思考了几个月应该以哪种方式让 Go 语言支持别名后，这个特性才被实现。[最初的提案](https://go.googlesource.com/proposal/+/master/design/16339-alias-decls.md)是引入通用的别名（支持对类型、函数等等赋别名），但这个提案后来被另一个[更简单的别名机制](https://go.googlesource.com/proposal/+/master/design/16339-alias-decls.md)所替代，新提案只关注对类型赋别人，因为对这个特性需求最大的就是类型。只支持对类型赋别名让实现方式变得简单，因为只需要解决最初始的问题就可以了。我们一起来看看这个解决方案。

## 重构

引入别名的最主要的意图是简化对大型代码库的重构。开发者们对旧名字赋一个新的别名，就可以避免破坏已存在代码的兼容性。下面是一个 `docker/cli` 的例子：

```go
package command// Deprecated: Use github.com/docker/cli/cli/streams.In instead
type InStream = streams.In

// Deprecated: Use github.com/docker/cli/cli/streams.Out instead
type OutStream = streams.Out
```

这样不会影响使用 `command.InStream` 的旧代码，而新代码使用新类型 `streams.In` 。