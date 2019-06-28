首发于：https://studygolang.com/articles/21387

# Golang Module Vs Dep: 支持 & 反对

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/vgo-module/vgo-modules-dependency-management-golang-blog-hero-1200x630.png)

Go 语言有着一个热情的社区。尤其是当涉及推动语言本身及其生态的进步时，这种热情催生出精彩的讨论和许多很棒的想法，但是在前进的方向上发生分歧时，热情也会使社区分裂。比如，关于版本和依赖管理，在今年早些时候推出 vgo 和 Go modules 之前，社区主导的试验性的 dep 是很有可能成为事实标准的。

## Vgo + Modules

Vgo 是 Go 的参考实现，旨在提供通用的版本控制，尤其是依赖管理。Vgo 导致了 modules 的引入，以此来专门解决依赖及其版本控制的问题。截至撰写本文时，Go modules 的提案已经被采纳。2018 年 8 月 24 日发布的 Go 1.11 版本提供了对 modules 的临时性支持，2019 年 1 月 1 日发布的 Go 1.12 版本会正式支持该特性。

然而，社区一直都拒绝采纳 modules。

有人可能没意识到，Go modules 提供了将指定版本的包组成的集合表示为一个单元的方法。这一实现需要你在根目录下创建一个 Go.mod 文件，用来告诉 Go 这是一个 Go module，要使用指定版本的导入路径。Go.mod 里的配置项允许你指定有效版本的集合，实际的代码需要调用指定的版本号（译注：Go modules 使用一种称为“最小版本选择”的算法来确定依赖库的版本。比如，依赖库 M 的最新版本为 v1.2.3，那么 Go modules 允许使用 v1.2.3 及更新版本的 M；但是不能使用 v2 版本，因为 v2 被认为是与 v1 不兼容的。有关“最小版本选择”算法的详细解释，请参考 [https://github.com/golang/go/wiki/Modules#version-selection](https://github.com/golang/go/wiki/Modules#version-selection)）。现在，如果你使用的一个库从版本 2 升级到版本 3，那么你就需要修改你的代码。诚然这种方法可以帮助避免程序的崩溃，但是它也给开发者带来了各种额外的手工操作。当然，遗留代码将需要改成新的 module 支持的格式。

## 版本控制 & 依赖管理

说到遗留代码带来的痛点，代码不用放在 GOPATH 下是很有益的。和其他语言使用的基于项目的方法相比，GOPATH 是十分古板的。开发者期望能在自己的项目文件夹下工作，但是对于 Go 来说，代码不得不和其他文件隔离开。这种反直觉的方式已经影响了 Go 的接受度，阻碍了语言的发展。

关于依赖管理，现有的解决方案 dep 因其相对而言非限制的使用方法备受喜欢：导入一个库，dep 就会抓取它所能找到的所有版本。Go modules 需要使用一个 Go.mod 配置文件，并且要使用语义化的版本控制来声明你要导入哪个资源。

当要管理相互依赖时，事情甚至会变得更加复杂。比如，假设资源 A 需要有资源 B 的支持，资源 B 需要有 1.0 版本及以上的资源 D。但是资源 A 还需要有资源 C，资源 C 需要 1.0 或 1.1 版本的资源 D。Vgo 内置有一个备受争议的版本选择算法，该算法将总是选择适用于所有情况的最老版本，在这个例子中则是 1.0 版本的 D。大多数其他语言都会选择最新的可用版本，现有的 dep 也是如此。默认使用最老的可用版本对稳定性有很大好处，因为较新版本可能没有得到很好的测试 / 检验，但是这也可能会使本已在新版本得到修复的漏洞永久存在下去。

## 总结

Go 的版本控制和依赖管理问题远未得到解决。要解决已经出现的问题，Go 团队和社区依旧有许多工作要做。与以往一样，这些问题的解决需要许多沟通，时间和努力。版本控制和依赖管理是个很大的障碍，阻碍了 Go 被采纳为企业级语言。每个人都想着尽早解决这些问题。

---

via: https://www.activestate.com/blog/golang-module-vs-dep-pros-cons/

作者：[Pete Garcin](https://www.activestate.com/blog/author/peteg/)
译者：[maxwellhertz](https://github.com/maxwellhertz)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出