首发于：https://studygolang.com/articles/17960

# 用 Go 创建一个新的智能合约语言 - 新项目动机，概念

这篇文章讲述了 **为什么** 我们开始为智能合约创建新的编程语言（使用 Go）。

## 动机

目前有两个众所周知的区块链，比特币和以太坊。比特币有 **bitcoin script** 和以太坊有 **solidity** 为它们自己的智能合约编程。两者都有利有弊。

对于 **比特币** 而言，它没有状态概念，并且 bitcoin script 是基于低级语言和很少的操作，因此它所能做的是有限的。另一方面，因为它的工作方式简单并且比特币是没有状态的，这样我们能轻松地进行静态分析，如这个脚本运行速度有多快。

对于 **以太坊** 而言，它有状态的概念，并且 solidity 被设计为高级语言，solidity 开发者能够更直观的编程，并且以太坊智能合约可以做很多事情（是的，这是因为以太坊是有状态的）。另一方面，因为它被设计为高级语言，开发者可以错误地将无限循环放在永远不会结束的智能合约上面，这会在网络上造成不良影响。加上以太坊已经表明它很难做静态分析。

![koa concept](https://raw.githubusercontent.com/PotoYang/gctt-images/master/create-new-smartcontract-language-with-go-new-project-motivation-concept/koa-concept.png)
<center>koa 概念 </center>

我们的灵感来自于 2017 年由 Russell O'Connor 撰写的话“ Simplicity: A New Language for Blockchains ”和 [ivy_bitcoin](https://github.com/ivy-lang/ivy-bitcoin) 项目。

所以这就是“ koa ”生存的地方。**“ koa ”是高级加密货币语言。并且没有状态，静态分析很容易，比 bitcoin scripts 更多的操作。**

## 架构

![koa components](https://raw.githubusercontent.com/PotoYang/gctt-images/master/create-new-smartcontract-language-with-go-new-project-motivation-concept/koa-components.png)
<center>koa 组件 </center>

koa 项目正在制作新的编程语言，所以我们需要编译器，由于要将源代码编译为字节码，所以我们需要进行词法分析和语法分析。最后字节码在 VM 上运行。我们团队成员首先去编写编译器，因此我们阅读了大量的书籍、博客文章和研究了流行的开源项目源码，像 Go-ethereum。

这就是它。这是我们 koa 项目的理念：制作新的高级智能合约语言，可以轻松实现静态分析。在下一篇文章中，我们将深入到每一个组件中。你可以从 [此处](https://github.com/DE-labtory/koa) 查看 WIP 开源项目。

---

via: https://medium.com/@14wnrkim/create-new-smartcontract-language-with-go-new-project-motivation-concept-faca1931c0e2

作者：[zeroFruit](https://medium.com/@14wnrkim)
译者：[PotoYang](https://github.com/PotoYang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
