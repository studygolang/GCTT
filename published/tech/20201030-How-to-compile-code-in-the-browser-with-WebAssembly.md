首发于：https://studygolang.com/articles/35261

# 如何使用 WebAssembly 在浏览器中编译代码

浏览器的功能日益强大，从最早在 [CERN](https://home.cern/science/computing/birth-web) 上分享文章，到今天运行 [Google Earth](https://earth.google.com/web) ，玩 [Unity 3D](https://blogs.unity3d.com/2018/08/15/webassembly-is-here/) 游戏，甚至用 [AutoCAD](https://www.autodesk.com/products/autocad-web-app/overview) 设计建筑。

既然浏览器已然具有如此强大的功能，那么它能不能编译运行代码呢？可笑。当然不可能...

但是转念一想，为什么不呢？我不可能忽视这么一个令人兴奋的挑战的。在为期四个月的敲键盘和研读文档之后，我终于给出了自己的答案： [Go Wasm](https://go-wasm.johnstarich.com/)。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20201030-How-to-compile-code-in-the-browser-with-WebAssembly/1.gif)

Go Wasm 运用 [WebAssembly](https://webassembly.org/) 提供了一个完全在浏览器中书写和运行代码的 [Go](https://golang.org/) 开发环境。它完全开源。 Go Wasm 由三个 WebAssembly 组件组成： “操作系统”、编辑器和 shell。

本文接下来会从 Go Wasm 是什么，怎么运行的，以及未来发展三个方面展开介绍。

## 用 Go Wasm 写代码

Go Wasm 让你能用 Go 编译器写 Go 代码，运行 Go 代码。换句话说：我们写好代码，控制台输入 Go build ，然后就可以运行了。这跟我们熟悉的 [Go Playground](https://play.golang.org/) 很不一样。在 Go Wasm 中，代码实际上是在浏览器中运行的，即便你断开网络。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20201030-How-to-compile-code-in-the-browser-with-WebAssembly/2.gif) 

当你打开 [样例网站](https://go-wasm.johnstarich.com/) ，Go Wasm 首先启动操作系统，然后在虚拟文件系统中安装 Go ，然后打开编辑器，最后启动终端和编译控制等工具。在接下来的章节，我会对其中的三个关键程序详细介绍。首先，来快速的看一下 IDE 。

在编辑器中，你可以进行多面板编辑，引入外部库，重排格式，运行程序等操作。

想要运行传统 Go CLI 命令，你只需打开一个终端。程序的输入输出可以通过文件重定向操作，例如 ```./playground > output.txt``` 或者通过 ```|``` 管道连接输入与输出。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20201030-How-to-compile-code-in-the-browser-with-WebAssembly/3.png)

你也可以在虚拟文件系统中的任意位置生成、安装 Wasm 程序。 在上面的截屏中， ```count-lines``` 程序被编译并保存在了 ```/bin``` 目录中。在这个目录，它可以像其他内建命令一样被运行。

## 原理

Go Wasm 由三个 WebAssembly 组件组成： “操作系统”、编辑器和 shell 。这三个组件在浏览器中被层叠组装。在浏览器中，操作系统扮演解释器的功能，为访问虚拟文件和进程提供服务。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20201030-How-to-compile-code-in-the-browser-with-WebAssembly/4.png)

在上层，操作系统拦截 Go 程序和浏览器之间的系统调用，以提供虚拟文件系统和进程的服务。因为浏览器没有任何文件和进程的概念，操作系统对运行 Go 程序尤其重要。这里的“操作系统”并不是真正的操作系统，但是它提供了浏览器所不能提供的重要的对操作系统概念的抽象。

我们还需要 Go 编译器和标准库来写程序。 Go Wasm 启动时，会自动下载并安装 Go 到虚拟文件系统。

编辑器和 shell 程序建立在操作系统上，就像由编辑器编译的程序。编辑器提供文件管理面板以及用于写代码、运行代码的终端。每个终端面板启动一个 shell 用于运行传统 Go 命令或者你自己的程序，就像一个真正的终端一样。

## “操作系统”

操作系统为用户程序在各种不同硬件上运行的环境。操作系统为用户程序调用系统资源提供接口，但不允许用户程序直接访问系统资源。如果一个用户程序要打开一个文件并写入了一些数据，操作系统会返回一个文件句柄，并忠实地将用户程序写入的数据写到硬盘上。这些重要的操作通常被称为系统调用，即 "syscalls"。

类似地， Go Wasm 的操作系统对 Go 的系统调用采用的方式是拦截系统调用，操作虚拟资源而非实际资源的方式。Wasm 上 Go 程序执行系统调用依赖 JavaScript 的全局函数， Wasm 操作系统的组件将这些系统调用替换成普通代码。浏览器并不真的提供文件和进程服务，这些系统调用函数被替换成虚拟版本，并返回给 Go 程序。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20201030-How-to-compile-code-in-the-browser-with-WebAssembly/5.png)
<center> 打开文件过程的时序图 </center>

浏览器中的进程是一个有趣的问题。熟练的网络程序开发者都知道，并没有所谓的 "JavaScript process" 的概念，也绝对不可能在同一时间运行两个程序。为了解决这个限制，我通过上下文切换的方式在进程间来回切换，使得这些进程看起来像运行在一个处理器核心上。上下文切换使得操作系统能够在新进程取得优先级的时候对进程相关的文件表和环境变量进行换出 ( swap out ) 操作。这个操作较为复杂，但是剩下的工作就可以完全交给 JavaScript 的内建任务调度器。

## 虚拟文件系统

先将虚拟文件系统放在一边，因为想要把它弄对非常困难。一个现代文件系统 ( FS ) 包含许多的边界情况和特点。文件许可，原生管道，文件锁等都需要正常工作以满足 ```go build``` 的要求，以使程序运行不会碰到致命错误。呃。

为节省时间，我没有自己从零开始写一个基于内存的文件系统，选用了现成的 [Afero](https://github.com/spf13/afero) 。Afero 提供了一个很好的起点，它定义了一个非常强大的文件系统的抽象。基于此，我建立了几个文件系统：一个可挂载的文件系统，一个流式 gzip 文件系统和一个实验性的 IndexedDB 文件系统。不幸的是，即使有了这么强的基础， bug 仍然层出不穷。

我发现在底层的操作系统不是真实的时候 Go CLI 的表现不好。令人吃惊，是的。

我花了一个多月来解决每个文件系统的 bug 。跟踪 bug 尤其困难，因为多数的 bug 会使操作系统崩溃，并输出难以看懂的错误信息。我希望浏览器的开发者和 Wasm 社区能在 debug 的体验上花更多时间，因为这真是一个痛点。

## 编辑器

编辑器扮演了一个更加传统的网络应用的角色。它在网页上搭建了代码编辑器面板、控制窗口、 build 控制台和终端面板。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20201030-How-to-compile-code-in-the-browser-with-WebAssembly/6.png)

编辑器程序在操作系统的上层运行，就像其他 Go 程序一样启动新的进程。今天，编辑器运行 [CodeMirror](https://codemirror.net/) 并将其中的线上文件与我们的文件系统保持同步。

终端启动 shell 进程。为了提供人们更加熟悉的终端体验，终端将输入、输出与 [xterm.js](https://xtermjs.org/) 连接。

## Shell

每个终端面板的 shell 提供了一个在文件系统上运行 Go 命令和其他 Wasm 程序的接口。现在市面上有一些现成的用 Go 写的 shell。 但是我试用的一些并不能与 Wasm 兼容。幸运的是，写一个 shell 是一个很有趣的学习进程的方式。为了写 shell ，我不得不学习了所有的终端逃逸码，来支持命令行输入编辑。后来我又对进程文件重定向和环境变量产生了兴趣。

好了！我非常喜欢在 Go Wasm 中加入新库来实验我的疯狂的想法。我迫不及待的想看看小伙伴门会用它来做什么了。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20201030-How-to-compile-code-in-the-browser-with-WebAssembly/7.gif)

## 未来发展

现今的 WebAssembly 不可否认地有各种优点和缺点，但它也同样有着巨大的潜力等着被实现。

关于优点，我们的社区一直很忙。人们不停地提出具有前景的提议，提出各种标准以尝试突破极限。特别地，我一直在关注 [Wasm Threads](https://github.com/WebAssembly/threads) 和由其衍生的 [WASI](https://wasi.dev/) 标准，这会成为下一个里程碑。

不幸的是，仍然有很多不足。我最感到遗憾的问题之一就是缺少 [调试支持](https://github.com/WebAssembly/debugging) 。浏览器自带的调试器只能支持步进汇编码——不是源码。由于我对 Wasm 的文本格式不熟悉，对我源码中的问题进行诊断和修复的工作非常令人沮丧。

对于 Go Wasm ，未来是非常令人兴奋的。如果社区感兴趣的话，我非常乐于加入如移动设备支持（低内存消耗），会话之间的文件存取，或者 WASI 程序的原生运行。如果你想参与进来，在 [GitHub](https://github.com/johnstarich/go-wasm) 上吼一声。

<center> . . . </center>

Go Wasm 将 WebAssembly 的边界向外扩展了一点。尽管现在 Wasm 还处在早期，但是其发展势头迅猛。我等不及想看看接下来会发生的事了。

你会创造什么呢？

---
via: https://johnstarich.medium.com/how-to-compile-code-in-the-browser-with-webassembly-b59ffd452c2b

作者：[hongchi](http://www.hongchiworld.cn/)
译者：[hhh811](https://github.com/hhh811)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
