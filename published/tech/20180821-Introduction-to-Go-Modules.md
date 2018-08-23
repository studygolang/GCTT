首发于：https://studygolang.com/articles/14389

# Go 语言的 Modules 系统介绍

即将发布的 Go 语言 1.11 版本将会给我们带来对 *modules*（模块）的实验性支持，这是 Go 语言新的一套依赖管理系统。

（译注：很多编程语言中，把 modules 译作「模块」，但由于目前该机制在 Go 语言还没正式发布，因此尚未有非常普及的译法。而类似的 vendor 一词的翻译，大多中文文章都是采取保留英文原文的方式处理，因此本文对 modules 的翻译参考 vendor 的处理：保留英文原文）

前些日子，[我简单地写了一编关于它的文章](https://roberto.selbach.ca/playing-with-go-modules/)，自从那篇文章写完后，Go 的 modules 机制又发生了一些小改动。因为现在快到正式版发布了，我想现在刚好是时候用”边做边学“的风格来写一篇关于它的文章。

所以接下来我们要做的是：我们将会创建一个新的包，并且我们会发布几个版本，看看它们是怎么工作的。

## 创建一个 Module

第一件事情，我们先创建一个叫 `testmod` 的包。注意这里有个重要的细节：**它的目录要在 `$GOPATH` 之外，因为默认情况下，`$GOPATH` 里面是禁用 modules 支持的。**Go 的 modules 机制在某种程度上，是消灭整个 `$GOPATH` 的第一步。

```bash
$ mkdir testmod
$ cd testmod
```

 我们的包非常的简单：

```go
package testmod

import "fmt"

// Hi 返回一个友好的问候
func Hi(name string) string {
   return fmt.Sprintf("Hi, %s", name)
}
```

这个包已经写完了，但是现在还不是一个 module，我们来把它初始化为 module：

```bash
$ go mod init github.com/robteix/testmod
go: creating new go.mod: module github.com/robteix/testmod
```

上面的命令在当前目录里面创建了一个名为 `go.mod` 的文件，它的内容如下：

```
module github.com/robteix/testmod
```

并没有很多东西，但是它已经有效地使我们的包变成一个 module 了。

我们现在可以把这个代码推送到代码仓库里面了：

```bash
$ git init
$ git add *
$ git commit -am "First commit"
$ git push -u origin master
```

(译注：在 `git push` 之前，你可能还要添加远程仓库地址，例如：`git remote add origin https://github.com/robteix/testmod.git`)

到目前为止，任何想要用这个包的人，都可以 `go get` 之:

```bash
$ go get github.com/robteix/testmod
```

上述的命令会获取 `master` 分支上最新的代码。这个方法依然凑效，但是我们现在最好不要再这么做了，因为我们有更棒的方法了。获取 `master` 分支有潜在的危险，因为我们不能确定，包作者对包的改动会不会破坏掉我们的项目对该包的使用方式。（译注：也就是说，我们不能确定当前 `master` 分支的代码是否保持了对旧版本代码的兼容性）。而这个就是 modules 机制旨在解决的问题。

## 简单介绍一下模块版本化

Go 的 modules 是*版本化的*，并且某些版本有特殊的含义，你需要了解一下 [语意版本控制](https://semver.org/) 背后的概念。

更重要的是，Go 是根据代码仓库的标签来确定版本的，并且有些版本跟其它版本是不一样的：比如版本 2 以上应该跟版本 0 和版本 1 有不同的导入路径（我后面会讲到这个）。

还有，默认情况下 Go 会获取代码仓库里面设置好标签的最新的版本，只是一个很重要的知识点，因为你可能已经习惯了在 `master` 分支工作。

到目前为止，你现在要记得的是，要制作代码包的一个发布版本，我们需要给我们的代码仓库打上版本标签。所以，现在让我们开始吧。

## 制作我们的第一个发布版本

我们的包已经准备好了，现在我们可以向全世界发布它。我们通过版本标签来实现这个发布，现在，我们一起来发布我们的 1.0.0 版本:

```bash
$ git tag v1.0.0
$ git push --tags
```

上述命令在我们的仓库上面创建了一个标签，标记了我们当前的提交为 1.0.0 版本。

虽然 Go 没有强制要求，但是我们最好创建还是一个叫 `v1` 的分支，这样我们可以把针对这个版本的 bug 修复推送到这个分支：

```bash
$ git checkout -b v1
$ git push -u origin v1
```

现在我们可以切换到 `master` 分支，做自己要做的事情，而不用担心会影响到我们已经发布的 1.0.0 版本的代码。

## 使用我们的 module

现在我们准备好可以使用我们的 module 了，下面我们创建一个简单的程序来使用我们刚才做的包：

```go
package main

import (
    "fmt"

    "github.com/robteix/testmod"
)

func main() {
    fmt.Println(testmod.Hi("roberto"))
}
```

到现在，你可以 `go get github.com/robteix/testmod` 来下载这个包。但是对于 module 来说，事情就变得有趣了。首先我们需要在我们新的程序里面启用 module 功能：

```bash
$ go mod init mod
```

正如之前所发生的那样，上面的命令会创建一个 `go.mod` 文件，它的内容是：

```
module mod
```

当我们尝试构建我们的程序时，事情变得更加有趣了：

```bash
$ go build
go: finding github.com/robteix/testmod v1.0.0
go: downloading github.com/robteix/testmod v1.0.0
```

正如我们看到的，`go` 命令自动地获取程序导入的包。如果我们看看程序的 `go.md` , 我们可以看到内容出现了变化：

```
module mod
require github.com/robteix/testmod v1.0.0
```

而且我们还多了一个叫 `go.sum` 的文件，它包含了各个包的哈希值，用以保证我们获取到了正确的版本和文件:

```
github.com/robteix/testmod v1.0.0 h1:9EdH0EArQ/rkpss9Tj8gUnwx3w5p0jkzJrd5tRAhxnA=
github.com/robteix/testmod v1.0.0/go.mod h1:UVhi5McON9ZLc5kl5iN2bTXlL6ylcxE9VInV71RrlO8=
```

## 为一个发布了的版本修复漏洞

假如说，我们现在发现我们的包出现了一个漏洞：欢迎语漏了一个标点符号！人们开始生气了，因为我们友好的欢迎语并不够友好。所以我们赶紧修复这个漏洞并发布一个新版本：

```go
// Hi 返回一个友好的欢迎语
func Hi(name string) string {
-       return fmt.Sprintf("Hi, %s", name)
+       return fmt.Sprintf("Hi, %s!", name)
}
```

我们在 `v1` 分支做这些改动，因为这个 bug 只在 `v1` 版本中存在。当然，在实际的情况中，我们很有可能需要把这个改动应用到多个版本，这时候你可能就需要在 `master` 分支做这些改动，然后再向后移植（译注：back-port 或称 backporting, 参考[维基百科](https://zh.wikipedia.org/wiki/%E5%90%91%E5%BE%8C%E7%A7%BB%E6%A4%8D) ）。无论怎样，我们都需要在 `v1` 分支上有这些改动，并且把它标记为一个新的发布：

```bash
$ git commit -m "Emphasize our friendliness" testmod.go
$ git tag v1.0.1
$ git push --tags origin v1
```

## 更新 modules

默认情况下，Go 不会自己更新模块，这是一个好事因为我们希望我们的构建是有可预见性（predictability）的。如果每次依赖的包一有更新发布，Go 的 module 就自动更新，那么我们宁愿回到 Go v1.11 之前没有 Go module 的荒莽时代了。所以，我们需要更新 module 的话，我们要显式地告诉 Go。

我们可以使用我们的老朋友 `go get` 来更新 module:

- 运行 `go get -u` 将会升级到最新的*次要版本*或者*修订版本*（比如说，将会从 1.0.0 版本，升级到——举个例子——1.0.1 版本，或者 1.1.0 版本，如果 1.1.0 版本存在的话）
- 运行 `go get -u=patch` 将会升级到最新的修订版本（比如说，将会升级到 1.0.1 版本，但**不会**升级到 1.1.0 版本）
- 运行 `go get package@version` 将会升级到指定的版本号（比如说，`github.com/robteix/testmod@v1.0.1`）

(译注：语义化版本号规范把版本号如 v1.2.3 中的 1 定义为主要版本号，2 为次要版本号，3 为修订版本号 )

上述列举的情况，似乎没有提到如何更新到最新的主要版本的方法。这么做是有原因的，我们之后会说到。

因为我们的程序使用的是包 1.0.0 的版本，并且我们刚刚创建了 1.0.1 版本，下面任意一条命令都可以让我们程序使用的包更新到 1.0.1 版本：

```bash
$ go get -u
$ go get -u=patch
$ go get github.com/robteix/testmod@v1.0.1
```

运行完其中一个（比如说 `go get -u`）之后，我们的 `go.mod` 文件变成了：

```
module mod
require github.com/robteix/testmod v1.0.1
```

## 主要版本号

根据语义化版本的语义，主要版本与次要版本是不同的，主要版本可以打破向后兼容性。从 Go modules 的角度来说，一个包，如果两个主要版本号不同的话，那这它们相当于两个完全不同的包。这听起来很玄乎，但是它是合理的：如果一个包的两个版本不能兼容的话，它就是两个不同的包。

我们来为我们的包做一个主要版本号的改变，怎么样？我们发现我们的 API 太过简单，对于我们用户的用例限制太多，所以我们给 `Hi()` 函数加多一个参数，来指定欢迎语的语言：

```go
package testmod

import (
    "errors"
    "fmt"
)

// Hi 返回一个欢迎语，其语言由 lang 指定
func Hi(name, lang string) (string, error) {
    switch lang {
    case "en":
        return fmt.Sprintf("Hi, %s!", name), nil
    case "pt":
        return fmt.Sprintf("Oi, %s!", name), nil
    case "es":
        return fmt.Sprintf("¡Hola, %s!", name), nil
    case "fr":
        return fmt.Sprintf("Bonjour, %s!", name), nil
    case "cn":
        return fmt.Sprintf("你好，%s！", name), nil
    default:
        return "", errors.New("unknown language")
    }
}
```

以前使用我们的包的项目，如果直接使用现在这个新的版本，它们将不能编译通过，因为它们没有传递 `lang` 参数，并且它们没有接收返回的 `error` 错误。所以我们的 API 与 v1.x 版本的 API 不能兼容，是时候跃进新的2.0.0时代啦！

我之前提到的，某些版本有特殊的含义，现在就是这种情况，版本 2 和更高版本需要改变导入路径，它们已经是不同的包了。

我们需要在我们 module 名字后面添加一个新的版本路径：

```
module github.com/robteix/testmod/v2
```

剩下的事情跟我们之前做的一样，我们把它标记成 v2.0.0，并推送到远程仓库（并且可选地，我们还能添加一个 `v2` 分支）

```bash
$ git commit testmod.go -m "Change Hi to allow multilang"
$ git checkout -b v2 # 可选的，但是推荐这么做
$ echo "module github.com/robteix/testmod/v2" > go.mod
$ git commit go.mod -m "Bump version to v2"
$ git tag v2.0.0
$ git push --tags origin v2 # 如果没有新建 v2 分支，就推送到 master 分支
```

## 更新到一个新的主要版本

虽然刚刚我们的包发布了一个新的版本，而且这个新的版本并不能向后兼容，但是使用我们的包的现有项目却不会受影响。因为它们还是会继续使用现有的 1.0.1 版本，`go get -u` 不会把项目使用的包的版本升级到 2.0.0

某些情况下，我作为一个库的使用者，可能会希望升级到 2.0.0 版本，因为我可能需要用到多语言的支持。

要这么做的话，我要相应的修改我的程序：

```go
package main

import (
    "fmt"
    "github.com/robteix/testmod/v2"  //注意包导入路径变了
)

func main() {
    g, err := testmod.Hi("Roberto", "pt")
    if err != nil {
        panic(err)
    }
    fmt.Println(g)
}
```

然后我再执行一下 `go build` ，它会自动帮我获取 v2.0.0 版本的包。要注意的是，虽然现在包导入路径是以 "v2" 结尾的，但是我们在代码里面依然用 `testmod` 这个包名来引用它。

正如我之前所提到的，两个主要版本号不同的包，在各种目的和意图上，都是两个不同的包。Go modules 并不会把这两个包链接在一起，这意味着我们可以在程序里面同时使用这个包的两个不同的主要版本：

```go
package main
import (
    "fmt"
    "github.com/robteix/testmod"
    testmodML "github.com/robteix/testmod/v2"
)

func main() {
    fmt.Println(testmod.Hi("Roberto"))
    g, err := testmodML.Hi("Roberto", "pt")
    if err != nil {
        panic(err)
    }
    fmt.Println(g)
}
```

这解决了依赖管理方面的一个常见的问题：当项目依赖于同一个库的两个不同版本时，该如何处理。

## 整理一下

回到我们之前那个只用了 testmod 2.0.0 的程序，如果我们看一下 `go.mod` 的内容，我们会发现：

```
module mod
require github.com/robteix/testmod v1.0.1
require github.com/robteix/testmod/v2 v2.0.0
```

默认情况下，Go 并不会在 `go.mod` 上面移除掉依赖项，除非你明确地指示它这么做。如果你希望能够清理掉那些不再需要的依赖项，你可以使用新的 `tidy` 命令：

```bash
$ go mod tidy
```

现在剩下的依赖项都是我们项目中使用到的了。

## Vendor 机制

默认情况下，Go modules 会忽略 `vendor/` 目录。这个想法是最终废除掉 vendor 机制[^1]。但如果我们仍然想要在我们的版本管理中添加 vendor 机制管理依赖，我们还是可以这么做的：

```bash
$ go mod vendor
```

这会在你项目的根目录创建一个 `vendor/`目录，并包含你的项目的所有依赖项。

即使如此，`go build` 默认还是会忽略这个目录的内容，如果你想要构建的时候从 `vendor/` 目录中获取依赖的代码来构建，那么你需要明确的指示：

```bash
$ go build -mod vendor
```

我猜想大多数要使用 vendor 机制的开发者，在他们自己的开发机器上会使用 `go build` ，而在他们的CI系统（Continuous Integration，持续集成）上则使用 `-mod vendor` 选项

还有，对于那些不想要直接依赖版本控制服务（译注：比如 github.com）上游代码的人来说，比起用 vendor 这种机制，更好的方法是使用 Go module 代理。

有很多方法可以保证 `go` 不会联网去获取包代码（比如 `GOPROXY=off`），但这些内容只能在之后的文章提及了。

## 结论

这篇文章看起来比较吓人，但我尽量的把很多事情放在一块解释了。事实是，现在 Go 的 modules 机制基本上是透明的，我们像往常一样在我们的代码里面导入包，`go` 会处理剩下的事情。

当我们构建程序的时候，它的依赖项会被自动地获取。Go 的 module 还消除了 `$GOPATH` 的使用， `$GOPATH` 曾经使得很多 Go 开发新手难以理解为什么所以东西都要放到一个特定的目录。

~~Vendor 机制已经被使用 module 代理的方法取代了~~[^1]，我大概会专门新开一篇关于 Go module 代理的的文章。

---

[^1]: 我觉得好像这么说似乎语气太重了，让人觉得好像 vendor 机制就要立刻被废弃了似的，并不是这样的，vendor 机制还能用，尽管和以前略有不同。似乎总有想要用更好的东西替代 vendor 机制的愿望，或许这个替代方案就是 module 代理，又或许不是。但现在就是这样的情况：人们有想要用更好的方案替代 vendor 机制的愿望，但 vendor 机制在一个更好替代方案（如果真的能有更好的替代方案）出现之前，是不会废弃的。

---

via: https://roberto.selbach.ca/intro-to-go-modules

作者：[Roberto Selbach](https://roberto.selbach.ca/author/robteix/)
译者：[Alex-liutao](https://github.com/Alex-liutao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
