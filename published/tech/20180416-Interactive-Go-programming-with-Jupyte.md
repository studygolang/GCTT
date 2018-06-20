已发布：https://studygolang.com/articles/13235

# 与 Jupyter 交互的 Go 编程

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/jupyte/go_jupyter_1.jpeg)

最近几年，Go 语言变得非常流行。我是 Python 的狂热粉丝，三年前我的第一个业余项目也是用 Python 实现的。而现在我开始使用 Go 语言来取代 Python，因为不管是业余爱好的小项目还是公司里的大项目，Go 语言能让我的编码效率更高。

与此同时，随着机器学习和数据科学（data science）变得越来越重要，Python 也更加流行。在机器学习中首选 Python 有很多原因，其中一个原因是 Python 是为交互式代码编写和计算而设计的。另一个重要的原因是 Python 中有一个很好的交互式编程工具：[Jupyter Notebook](http://jupyter.org/)。

虽然我现在在许多以前使用 Python 的项目中使用 Go 语言，但我仍然需要使用 Python 进行机器学习研究和数据分析。Python 中的交互式编程和 Jupyter Notebook 的能力对我来说仍然非常有吸引力。我希望有一个真正有用的 Go 语言的 Jupyter 环境以及能够验证 Go 语言正确性的 Jupyter 内核。但现在还没有这样的项目，有些类似的项目已经终止。虽然其中有一些比较流行，但不适合实际使用，因为它们不支持类型安全、代码取消、代码完成、检查或显示非文本内容。

因此，我决定开发一个新的环境，从头开始在 Jupyter Notebook 上交互式运行 Go 语言。在这，我向大家介绍我构建的软件，以及像 Python 一样交互式编写和执行 Go 的新方法。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/jupyte/go_jupyter_2.gif)

## GitHub 上项目主页

我向大家介绍的项目是 yunabe/lgo，一个用于 Go 语言的 Jupyter Notebook 内核以及交互式解释器。详情请看以下仓库。

[yunabe/lgo](https://github.com/yunabe/lgo)

## 在浏览器试一试

以下链接是 Go 语言的 Jupyter 线上运行环境：

[mybinder.org](https://mybinder.org/v2/gh/yunabe/lgo-binder/master?filepath=basics.ipynb)

感谢 binder [(mybinder.org)](https://mybinder.org/), 你可以在你的浏览器上使用 binder 上的临时 docker 容器尝试 Go 语言的 Jupyter环境(lgo)。从上面的按钮打开临时的 Jupyter Notebook，享受交互式 Go 编程！

## 主要特点

* 像 Python 一样编写和运行 Go 程序。
* Jupyter Notebook 功能
* 完全符合 Go 语言规范，同时 100% 兼容 Go 语言编译器。
* 拥有 Jupyter Notebook 一样的代码补全，检查和代码格式化。
* 显示图像，HTML，JavaScript，SVG 等。
* 控制台上的交互式解释器
* 完全支持 goroutine 以及 channel

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/jupyte/go_jupyter_3.jpeg)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/jupyte/go_jupyter_4.jpeg)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/jupyte/go_jupyter_5.jpeg)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/jupyte/go_jupyter_6.jpeg)

## 安装

有两种方法可以将 Go 语言的 Jupyter 环境安装到你的计算机中。

* [使用预先构建的 Docker 镜像](https://github.com/yunabe/lgo#quick-start-with-docker)
* [源码安装（目前仅支持 Linux）](https://github.com/yunabe/lgo#install)

如果您想在计算机上快速尝试 Go 语言的 Jupyter环境，请先尝试 Docker 版本。 如果你使用 Linux 并且想要将 Jupyter 环境与 Go 环境集成到你的计算机中，那么你可以选择源码安装。 由于使用了 [`-buildmode = shared` 进行回归](https://github.com/golang/go/issues/24034)，lgo 的代码在 go1.10 中运行起来很慢。 在 go1.10 修正 bug 之前，请使用 go1.9 来尝试 lgo 。 目前 lgo 在 go1.9 以及 go1.8 完美运行。

Windows 和 Mac 用户，请使用 Docker 版本，因为 lgo 不支持 Windows 和 Mac。你可以在 Windows 或 Mac 上的 Docker 来运行 lgo。

## 使用

像平常一样执行 `jupyter notebook` 命令来启动 Jupyter Notebook。当你新建一份笔记时，请从菜单中选择 `Go (lgo)`。一旦创建了一个新的笔记，你就可以像 Python 那样交互式地编写和执行程序。

在 lgo 中，你可以通过将光标移动到标识符和按 `Shift-Tab` 来显示变量、函数和类型的相关文档。您可以通过按 `Tab` 来补全代码。如果要显示非文本数据，你可以参考[这个例子](http://nbviewer.jupyter.org/github/yunabe/lgo/blob/master/examples/basics.ipynb#Display)来使用 `DataDisplayer` 类型。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/jupyte/go_jupyter_7.jpeg)

### 像控制台的解释器环境那样使用

你同样可以像使用解释器那样使用 lgo。在安装完成后，运行指令 `jupyter console --kernel lgo` 即可。当然，在这种模式中你也可以使用 `Tab` or `Ctrl-I` 来实现代码补全。

```
pyter console --kernel
In [1]: a, b := 3, 4

In [2]: func sum(x, y int) int {
      :     return x + y
      :     }

In [3]: import "fmt"

In [4]: fmt.Sprintf("sum(%d, %d) = %d", a, b, sum(a, b))
sum(3, 4) = 7
```

## 与现有框架的比较

对于那些了解其他现有的 golang Jupyter 内核的人，这里是与竞争对手的比较表。你可以阅读 [`READNE.MD` 中的这部分](https://github.com/yunabe/lgo#comparisons-with-similar-projects)获取更多细节。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/jupyte/go_jupyter_8.jpeg)

## 了解更多

如果你想了解更多，请浏览[本项目的主页](https://github.com/yunabe/lgo)并阅读 `README.md` 中的介绍。此外，你还可以通过这些[示例笔记](https://nbviewer.jupyter.org/github/yunabe/lgo/blob/master/examples/basics.ipynb)中了解更多 Go 语言的Jupyter环境的真正用途。尽情享受 Go 语言的交互式编程吧!

---

via: https://medium.com/@yunabe/interactive-go-programming-with-jupyter-93fbf089aff1

作者：[Yu Watanabe](https://medium.com/@yunabe)
译者：[7Ethan](https://github.com/7Ethan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
