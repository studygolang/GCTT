首发于：https://studygolang.com/articles/17938

# 利用 GoLand 和 Mozilla rr 项目来调试

调试器。传统上，它们被用来发现复杂的 bug 并解释它们是如何发生的。但是，如果您无法解释为什么在步骤之间会发生一些更改，该怎么办呢？这就是典型调试器无法帮助您的地方，因为它们通常只会让您继续执行。

正如我们在前一篇文章中所看到的，虽然可以使用[核心转储](https://blog.gopheracademy.com/advent-2018/postmortem-debugging-delve/)，但它们并不总是告诉您应用程序中发生的事情的全部情况。

输入可逆的调试器。这些调试器不仅允许您在执行过程中逐步前进，还允许您返回并有效地撤消步骤之间的所有操作。

Go 调试器 [Delve](https://github.com/go-delve/delve) 通过使用 [Mozilla 的 rr 项目](https://rr-project.org/) 支持此类功能。从 rr 项目的描述来看，它的任务是允许“在调试器下反复重播失败的执行，直到完全理解它为止”。

让我们看看实际情况。首先，rr 只能在 Linux 上运行，这是一些严格的限制。再加上[其他一些限制](https://github.com/mozilla/rr/wiki/Building-And-Installing#hardwaresoftware-configuration)，确实影响了它的实用性。

有了这些，让我们进入代码。我将使用一个简单的应用程序来演示这些特性。

```go
package main

import (
	"log"
	"net/http"
)

func main() {
	c := http.Client{}
	resp, err := c.Head("htp://google.com")
	if err != nil {
		log.Fatalln("failed to make the request")
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalln("failed to make the request")
	}
	log.Println("the address is still working")
}
```

如果我们运行它，它将打印“请求失败”。在安装 [rr](https://github.com/mozilla/rr/wiki/Building-And-Installing) 之后，确保您运行所需的配置，以使 rr 运行 :

```shell
echo -1 | sudo tee -a /proc/sys/kernel/perf_event_paranoid
echo 0 | sudo tee -a /proc/sys/kernel/kptr_restrict
```

这些设置不是永久的，如果重新启动计算机，需要再次设置它们。现在让我们开始实际调试应用程序。

我将使用 [GoLand](https://www.jetbrains.com/go/) 来运行 Delve 和 rr。这样我就可以在几次单击运行它，并在执行调试步骤时查看源代码和变量 / 内存内容，并尝试了解发生了什么。

创建项目后，单击编辑器窗口左侧设置断点，然后单击 “ main ” 函数旁边的绿色箭头，选择 “ Record and Debug …” 选项。这将启动所需的编译步骤，然后使用 rr 后端启动调试器。

调试器在断点处停止后，现在可以返回执行。由于 rr 项目的工作方式，我们首先需要在前面的语句中放置第二个断点，然后使用 “ Rewind ” 按钮。现在，我们可以使用 “ Step into ”、“ Step over ” 等常规命令，或者使用 “ evaluate ” 功能对表达式求值。

在继续调试应用程序的过程中，我们可以放置更多的断点，通过跳过已知的良好区域或跳到代码的最后已知的良好部分来加速调试。

下面是调试器运行的一小段视频 :

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-mozilla-rr/debugging-with-rr.gif)

我希望这篇文章将帮助您发现一个新的、强大的工具，它可以加速发现和修复应用程序中的 bug。

不要忘记访问 Delve 的仓库，对其 star，甚至可能为这个了不起的项目做出贡献。

如果您有任何评论或想了解更多，请使用下面的评论部分，或通过 Twitter [@dlsniper](https://twitter.com/dlsniper) 联系我。

---

via: https://blog.gopheracademy.com/advent-2018/mozilla-rr-with-goland/

作者：[FlorinPăţan](https://twitter.com/dlsniper)
译者：[wumansgy](https://github.com/wumansgy)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
