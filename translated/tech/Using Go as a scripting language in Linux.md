# 在Linux中使用Go作为脚本语言

在`Cloudflare`的人们都非常喜欢Go语言。我们在许多[内部软件项目](https://blog.cloudflare.com/what-weve-been-doing-with-go/)以及更大的[管道系统](https://blog.cloudflare.com/meet-gatebot-a-bot-that-allows-us-to-sleep/)中使用它。但是，我们能否进入下一个层次并将其用作我们最喜欢的操作系统Linux的脚本语言呢？

![image here]()

## 为什么考虑将Go作为脚本语言

简短点的回答：为什么不呢？Go相对容易学习，不冗余并且有一个强大的生态库，这些库可以重复使用避免我们从头开始编写所有代码。它可能带来的一些其他潜在优势：

* 为你的Go项目提供一个基于Go的构建系统：`go build`命令主要适用于小型自包含项目。更复杂的项目通常采用构建系统或脚本集。为什么不用Go编写这些脚本呢？

* 易于使用的非特权包管理：如果你想在脚本中使用第三方库，你可以简单的使用`go get`命令来获取。而且由于拉取的代码将安装在你的`GOPATH`中，使用一些第三方库并不需要系统管理员的权限（与其他一些脚本语言不同）。这在大型企业环境中尤其有用。

* 在早期项目阶段进行快速的代码原型设计：当编写第一轮代码时，即使编译它通常都需要大量的编辑，并且必须在“编辑->构建->检查”的旋转中浪费大量的时间。相反，使用Go可以跳过`build`部分，并立即执行源文件。

* 强类型的脚本语言：如果你在脚本中的某个地方有个小的输入错误，大多数的脚本语言都会执行到有错误的地方然后停止。这可能会让你的系统处于不一致的状态（因为有些语句的执行会改变数据的状态，从而污染了执行脚本之前的状态）。使用强类型语言时，编译时会捕获许多输入错误，所以有bug的脚本不会首先运行。

## Go脚本的当前状态

咋一看Go脚本貌似很容易实现Unix脚本的shebang(#! ...)支持。shebang行是脚本的第一行，以`#!`开头，并指定脚本解释器用于执行脚本（例如，`#!/bin/bash`或`#!/usr/bin/env python`），所以无论使用何种编程语言，系统都确切知道如何执行脚本。Go已经使用`go run`命令支持`.go`文件的类似于解释器的调用，所以只需要添加适当的shebang行（`#!/usr/bin/env go run`）到任何的`.go`文件中，设置好可执行状态位，然后就可以愉快的玩耍了。

但是，直接使用go run还是有问题的。[这篇牛b的文章](https://gist.github.com/posener/73ffd326d88483df6b1cb66e8ed1e0bd)详细描述了围绕`go run`的所有问题和潜在解决方法，但其要点是：

* `go run`不能正确地将脚本错误代码返回给操作系统，这对脚本很重要，因为错误代码是多个脚本之间相互交互和操作系统环境最常见的方式之一。

* ...


---

via：[Using Go as a scripting language in Linux](https://blog.cloudflare.com/using-go-as-a-scripting-language-in-linux/)

作者：[Ignat Korchagin](https://blog.cloudflare.com/author/ignat/)
译者：[shniu](https://github.com/shniu)
校对：[?](https://github.com/?)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，
[Go中文网](https://studygolang.com/) 荣誉推出
