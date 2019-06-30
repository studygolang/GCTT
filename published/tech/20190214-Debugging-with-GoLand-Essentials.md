首发于：https://studygolang.com/articles/20746

# 使用 GoLand 进行调试的要点

*由 [Florin Păţan](https://blog.jetbrains.com/go/author/florin-patanjetbrains-com/) 于 [2019 年 2 月 14 日 ](https://blog.jetbrains.com/go/2019/02/14/debugging-with-goland-essentials/) 发表*

在今天的帖子中，我们将继续探索 GoLand 中的调试器功能。如果你想知道如何配置调试器。请查看我们之前的帖子，其中包含关于如何配置 IDE 在各种方案中工作的所有信息。

我们将讨论：

* [控制执行流程](# 控制执行流程 )
* [计算表达式](# 计算表达式 )
* [查看自定义值](# 查看自定义值 )
* [更改变量值](# 更改变量值 )
* [使用断点](# 使用断点 )

在我们启动调试会话之前，让我们为感兴趣的地方添加一些断点，以便了解代码是如何运行的，然后启动调试对话。
![debugging-session](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/6-optimized.gif)

## 控制执行流程

我们可以从这里完全控制调试器，我们可以用 step into, smart step into, step over, step out 或者运行代码到光标处。
![Controlling-the-execution-flow](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/7-optimized.gif)

## 计算表达式

我们也可以用它来计算简单的表达式。由于 Delve 的限制，目前不支持调用函数，但请为此功能投票以获取更多信息：[https://youtrack.jetbrains.com/issue/GO-3433](https://youtrack.jetbrains.com/issue/GO-3433)
![feature-informations](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/8-optimized.gif)

## 查看自定义值

我们通过创建一个新的观察器（ watch）来监控自定义表达式。这对观察复杂的表达式或查看 slice/map/struct 中的特定值是非常有用的。
![Watching-custom-values](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/9-optimized.gif)

## 更改变量值

由于 Go 运行时的限制，目前只能对***非字符串基本类型***（如 ***int***, ***float*** 或 ***boolean*** 类型）更改值。
要执行此操作，请从变量视图中选择要更改的值，然后按 *F2* 键并开始输入值。当你满意时，请按 ***Enter*** 键，你的代码现在将使用一个不同的值。
![Changing-variable-values](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/10-optimized.gif)

## 使用断点

设置断点是一个相当简单的操作，单击要停止执行的行左侧，或者用快捷键 ***Ctrl+F8/Cmd+F8***，调试器将停止执行。
如果你不需要更多东西，那这就是你真正需要知道的所有东西。但是，GoLand 为你和调试器如何与断点进行交互提供了一些很好的选择

按下 ***Ctrl+Shift+F8/Cmd+Shift+F8*** 一次，你将会看到一个包含几个选项的屏幕

再次按下 ***Ctrl+Shift+F8/Cmd+Shift+F8***，将显示断点列表，其中包含所有可用选项的完整列表

![Working-with-breakpoints-1](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/11-optimized.gif)

在此处你可以启动或禁用断点，并用调试器在调试对话期间暂停调试进程的执行，或仅在满足某个特定的条件时才触发断点。

![Working-with-breakpoints-2](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/12-optimized.gif)

我们还可以选择让 IDE 记录到达的某个断点，或者在控制台中打印堆栈轨迹，以便我们可以在文本模式下查看它。

这里我们可以使用另一个强大的功能 ***求值和记录（Evaluate and log)*** ，它可以让 IDE 计算一个表达式并在***控制台***中打印。
如果你想放置临时断点，可以启用“*一次删除选项*”，IDE 将自动删除断点。或者，只有前置断点被满足时，才能启用该断点，这对于调试复杂的条件代码非常有用。

![Working-with-breakpoints-3](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/13-optimized.gif)

在这篇文章中，我们研究了使用 IDE 调试应用程序和测试。这将帮助我们更快地深入了解应用程序，更有效地发现和修复错误。

在下一篇博文中，我们将探讨 2019.1 发行版中调试器的新功能，因此关注我们的博客和社交媒体渠道以获取更新。

一如既往，如果你有任何反馈意见，请在下面的评论部分，[Twitter](https://twitter.com/GoLandIDE) 或我们的[问题追踪](https://youtrack.jetbrains.com/issues/Go) 上告诉我们。

---

via: https://blog.jetbrains.com/go/2019/02/14/debugging-with-goland-essentials/

作者：[Florin Pățan](https://blog.jetbrains.com/go/author/florin-patanjetbrains-com/)
译者：[piglig](https://github.com/piglig)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
