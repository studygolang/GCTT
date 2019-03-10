首发于：https://studygolang.com/articles/18625

# 为什么要学习更多的编程语言

![page head](https://raw.githubusercontent.com/studygolang/gctt-images/master/why-need-to-learn-mpl/1.jpg)

## 我与编程语言的开放式关系

迄今为止我已经编程四年了。从开始的基于 C# 的游戏开发，然后使用 python 处理机器学习。学习使用 Javascript 以及 Typescript 做前端工作。后来想要做移动端应用，所以又学习了 Ionic，React，React Naive。为了达到更好的后端性能，Go 是一个最佳选择。随着 Flutter 的诞生，所以我学习了 Dart 来编写更多的移动应用。我从一些大学课堂上学习 Java，当我在 Facebook 工作时学习使用 PHP。

我不敢说在这些语言中我称得上专家，但是我比其他人有更多的关于语言和框架的经验。为什么我热衷于学习编程语言？因为我是容易被诱惑的，当我看到一些语言中的一些优异的特性时，我就忍不住去学习它。

![page middle](https://raw.githubusercontent.com/studygolang/gctt-images/master/why-need-to-learn-mpl/2.jpg)

那么我为什么让你做同样的事呢？因为在你不了解有哪些工具，这些工具具体是做什么之前，你也不可能去选择正确的工具。选择正确的工具和武器可以帮助你赢得大部分的战争。我个人发现这一点在生活中十分有用。选择合适的语言，可以极大地减少解决问题所需要付出的努力。

## 解决真实世界中的问题

我来举一个真实的例子，使用合适的语言可以节省很多时间，而只关注问题的主要部分。几个月前，我选购了一个蓝牙耳机 AirPods。可以说是苹果发布的最好的技术了。我尝试了很多蓝牙耳机，但是没有一个像这个一样方便。但是，主要问题是我个人使用的是 Windows 笔记本以及 Android 手机。AirPods 可以自动连接到我的手机，但是笔记本却不是这样。我每次必须设置并且手动连接，这是一个痛苦的过程。因为我一直想在手机以及笔记本电脑间切换。我需要一个可以将 AirPods 一键连接到笔记本的快捷按钮。

我的第一个想法是使用 python, 因为我确信可以找控制电脑蓝牙的库。但并不是这样，没有一个维护良好的库可以完成这个工作。下一个选择是 Node.js。后来我发现了一个可以控制蓝牙的 Javascript 库。通过运行以下脚本，我可以将 AirPods 立即连接到我的电脑。

```javascript
// App.js
const device = new bluetooth.DeviceINQ();

const airpodsAddress = "18:81:0E:B2:6B:A6"
const airpodsName = "Akshat's Airpods";

device.findSerialPortChannel(airpodsAddress, function (channel) {

    // make bluetooth connect to remote device
    bluetooth.connect(airpodsAddress, channel, function (err, connection) {
        if (err) return console.error(err);

        console.log('YAY! Airpods Connected');
        // Don't need a communication stream between the two
        // so let's just exit the stream.
        setTimeout(() => process.exit(0), 5000);
    });
});
```

现在我需要一个可以运行该脚本的一个快捷方式。我以为可以直接将脚本放在任务栏，但是 windows 不允许任何非可执行文件放在工具栏。我写了一个批处理文件，希望挂载在任务栏，但还是失败了。那么什么语言可以创建一个可执行文件呢？ Golang 是一个不错的选择，我写了一个脚本来运行 Node.js 脚本.

```go
// main.go
package main

import (
	"fmt"
	"os/exec"
)

func main() {
	output, err := exec.Command("npm", "start").CombinedOutput()

	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(output))
}
```

在任务栏放置该可执行文件的快捷方式，设置图标。太棒了！一个便于访问的按钮，可以让我很快的连接到 AirPods。

![con](https://raw.githubusercontent.com/studygolang/gctt-images/master/why-need-to-learn-mpl/3.gif)

我知道通过使用 C# 我也可以获得相同的结果。但我不想在我的笔记本电脑上安装 Visual Studio 这样一个怪异的 IDE。我还可以使用 nexe 等其他工具将我的 Nodejs 应用程序打包到 exe 中，但这只是不必要的工作。
这只是一个简单的例子，说明了解不同的工具如何帮助您轻松解决问题。如果我所知道的只是 Python 或 Java 或 Go，那将是一件非常困难的事情。我有更多的例子，知道使用正确的语言，大大减少了解决问题所需的时间和精力。

## 重点

1. 学习不同语言真的很有趣。此外，它还可以扩展您的视野，让您置身于舒适区之外。
2. 学习更多语言的另一个原因是训练自己思考一种语言或范式之外的问题。面向对象编程很棒，但也需要了解功能编程或程序编程。一旦你可以训练自己去思考特定语言之外的编程，你将不再受限于它的限制。
3. 你学习的第一语言将是困难的，第二语言将更难，但在那之后就是信手拈来了。这只是语法变化和一些陷阱的避免。然后，您可以了解该语言的特定库和框架。
4. 我能想到学习更多语言的另一个令人信服的理由是 WASM。 Web Assembly 将允许您在浏览器上运行所需的任何语言。这意味着如果您学习更快速的语言（如 C ++），可以充分利用浏览器的快速性并创建像 https://squoosh.app/ 这样的精彩内容。

## 最后的思考

1. 你是一个Javascript或python开发人员。我强烈建议学习低级语言。你可以直接学习 C 或 C ++，但我会建议 Golang。您可以轻松获得类似 C++ 的速度，而不会受到 C 系列的挫折。
2. 对于所有低级语言开发人员，请尝试使用 python 或 Javascript。如果您还没有尝试过这些语言，那么您就错过了。 Python 就像伪代码，现在 Javascript 无处不在。这两种语言都可以让您使用低级语言。您可以为 Node.js 和 Python 编写C ++模块。相信我，它会改变你的生活。

我希望我已经说服你与你的主要语言建立开放的关系，并获得一些新的令人兴奋的经历。 如果你知道两种截然不同的语言，到目前为止你的经验是什么？您认为它对您的职业生涯有何帮助？请在评论中告诉我。

---

via: https://blog.usejournal.com/why-you-need-to-learn-more-programming-languages-9160d609eac3

作者：[Akshat Giri](https://blog.usejournal.com/@akshatgiri)
译者：[Inno Jia](https://kobehub.github.io)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
