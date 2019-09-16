首发于：https://studygolang.com/articles/23310

# 避免诸如 base、util、common 之类的包名

写一个好的 Go 语言包的开端是起一个好名字。将你的包名视为一个 elevator pitch，你必须用一个单词来说明。

包名不好的一个普遍的原因是被命名为 *utility*，这些包混合了 helpers 和 utility 代码，还包含了各种各样不相关的函数，因此根据它们提供的内容很难描述其作用。这经常导致一个包的名字取决于它所包含内容：实用工具（utilities）。

在开发一些深层次包结构的项目时，为了避免出现循环引用，同时复用辅助函数，通常会出现类似 *utils* 或 *helpers* 的包名。提取通用型方法到一个新的包里将打破循环引用，但是由于这个包是因为项目设计问题创建的，所以它的名字不能反应出它的目的，只能反应出它打破循环引用的功能。

> *[A little] duplication is far cheaper than the wrong abstraction.*
>
> *—* [Sandy Metz](https://www.sandimetz.com/blog/2016/1/20/the-wrong-abstraction)

为了改善 utils 或 helpers 的包名，我的推荐方法是先分析这个包在什么地方被引用，然后将相关函数移动到调用的包里。即使这会导致一些代码冗余，这也比在两个包之间引入依赖要好。在这种情况下，在 utility 包的函数被多处使用的情况下，我更倾向于分成多个包，并统一以相应的描述性名称来命名。

在两个或多个相关工具包中存在通用功能时，经常会发现名字类似 base 或 common 的包。例如一个 client 和 server 之间或一个 server 和它的 mock 之间的公有类型，都被重构在独立的包中。相反，解决方案是通过将 client、server 和共有代码合并进一个单独的包来减少包的数量，这个包以它提供的功能来命名。

例如，net/http 包没有 client 和 server 包，相反，包含 client.go 和 server.go 文件，每个文件都拥有它们各自的类型，transport.go 文件包含了被用在 HTTP clients 和 servers 端的公有的消息传输代码。

用包提供了什么功能来它命名，而不是用它包含的内容。

## 相关帖子

1. [Simple profiling package moved, updated](https://dave.cheney.net/2014/10/22/simple-profiling-package-moved-updated)

2. [The package level logger anti pattern](https://dave.cheney.net/2017/01/23/the-package-level-logger-anti-pattern)

3. [How to include C code in your Go package](https://dave.cheney.net/2013/09/07/how-to-include-c-code-in-your-go-package)

4. [Why I think Go package management is important](https://dave.cheney.net/2013/10/10/why-i-think-go-package-management-is-important)

   ​																										2019-01-08

via:https://dave.cheney.net/2019/01/08/avoid-package-names-like-base-util-or-common

作者：[*[Dave Cheney](https://dave.cheney.net/)*](https://dave.cheney.net/)
译者：[Alihanniba](https://github.com/Alihanniba) / 柒呀
校对：[zhoudingding](https://github.com/dingdingzhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
