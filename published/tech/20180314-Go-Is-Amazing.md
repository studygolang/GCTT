已发布：https://studygolang.com/articles/13097

# Go 实在是令人惊叹，但是我想说说我不喜欢它的地方

通过我的上一篇文章以及最近几个月期间对于Go编程语言的间接推广，我与许多开始对这门语言感兴趣的人们进行了交流，所以现在我打算转而去写一些我对这门语言的不满，依据我目前积累的经验来提供一种更加全面的看法，借此可以让一部分人意识到Go语言终究并不是他们项目的最佳选择。

**备注1**

需要重点指出的是，文章里的部分观点（如果不是全部的话）是基于我个人的主观想法并且跟我的编程习惯有关，它们没有必要也不应该被描述成“最佳解法”。还有就是，我现在仍旧是一个 Go 语言的菜鸟，我接下来要说的一些东西可能是不准确或者错误的，对于有误的地方请务必纠正我，这样我才能学到新东西。:D

**备注2**

在开始前我需要声明的是：我热爱这门语言并且我已经解释了为什么我觉得对于许多应用来说这是一个更佳的选择，但是我对于 Go 和 Rust 那个更好或者 Go 和其他任何语言哪一个更好这种问题不感兴趣……选择你认为最佳的方案去完成你要做的事情：如果你认为 Rust 更好就尝试使用它，如果你认为是你传送到处理器的字节码引起了数据总线的错误，就去尝试纠错，两种情况都是，尽管去编程，而不是浪费生命在盲目追逐所谓的流行语言上。

那么现在让我们从最小的问题着手逐渐递进到严重的问题上……

## 请给我一个三元运算符

在编写很大一部分运行在终端模拟器上的应用时，我发现自己总是会打印一些系统状态来确认自己正在调试的功能是开启还是关闭（例如开启或者关闭 bettercap 的其中一个模块并且报告该信息），这意味着很多时候我需要把一个布尔类型的变量转换成一个更容易理解的字符串，在 C++ 或者其他支持这种运算符的地方它是这个样子的：

```c
bool someEnabledFlagHere = false;
printf("Cool module is: %s\n", someEnabledFlagHere ? "enabled" : "not enabled");
```

不幸的是 Go 并不支持这种写法，这意味着你最后会写出这样的一堆东西：

```go
someEnabledFlagHere := false
isEnabledString := "not enabled"
if someEnabledFlagHere == true {
	isEnabledString = "enabled"
}
log.Printf("Cool module is: %s\n", isEnabledString)
```

并且这已经很可能是你能想到的最优雅解法（而不是为了实现这个功能而去创建一个 map）。这究竟是否能算是更方便了？对我而言这种写法很丑，并且当你的系统实现高度模块化的时候，一遍又一遍的写这种东西会让你的代码变得越来越臃肿，而这仅仅是因为少了一个操作符。ˉ\\_(ツ)_/ˉ

**备注** 好吧，我知道你可以通过创建函数或者使用字符串类型的别名来实现，但是完全没有必要在评论里把所有这些难看的替代方法都写出来，谢谢 :)

## 自动生成的这堆东西不等于文档

Go 语言的专家们，我衷心感谢你们分享的代码以及我每天阅读的时候学习到的这些东西，但是我不认为他们有什么用处：

```go
// this function adds two integers
// -put captain obvious meme here-
func addTwoNumbers(a, b int) int {
	return a + b
}
```

我并不认为[这样的东西](https://godoc.org/github.com/google/gopacket)可以代替文档，但是这看起来确实是 Go 语言使用者们给代码添加注释（文档）的标准方式（当然也有一些例外的情况），即使是在一些我们所熟知的拥有数以千计贡献者的框架中也是如此……我自己并不是很热衷于添加详细的文档，如果你喜欢自己深入研究代码的话这并不会是一个很大的问题，但是如果你是文档的重度依赖者，那么你恐怕要失望了。

## 把 Git 仓库作为包管理系统简直是疯了

我几天前在推特上有一段很有趣的对话，在那里我解释给某人听为什么 Go 导包的时候看起来很像 Github 的链接：

```go
import "github.com/bettercap/bettercap"
```

或者是像下面这样：

```
# go get github.com/bettercap/bettercap
```

简单来说，在 Go 最简单的安装方式中，你很可能会用到（不使用 vendor 目录并且也不覆盖  $GOPATH 变量的情况下）所有（事实上并不是，但是为了把问题简化我们可以这样假设）在这个安装目录或者你设置的 $GOPATH 变量目录下的东西，在我这里这个目录是 /home/evilsocket/gocode（是的，[确实是这样](https://github.com/evilsocket/dotfiles/blob/master/data/go.zshrc#L2)）。每当我使用 go get 命令获取或者通过导包后使用 go get 命令[自动下载所需的包](https://github.com/bettercap/bettercap/blob/master/Makefile#L28)时，它在我的电脑上基本是下面这个样子：

```
# mkdir -p $GOHOME/src
# git clone https://github.com/bettercap/bettercap.git $GOHOME/src/github.com/bettercap/bettercap
```

如你所见，Go 事实上直接使用了 Git 仓库来管理这些包，应用或者任何与 Go 有关的东西……从某方面来说确实很方便，但是这会引起一个很大的问题：只要你不使用其他工具或者基于这个问题做一些难看的规避方案，那么你每次在一个新系统上编译你的软件时，只要缺失了某个依赖包，这个依赖包所在仓库的主分支就会被克隆下来。这意味着，**尽管你应用的代码完全没有修改，但是你每次在新电脑上编译时都很可能会产生代码差异**（只要你任何一个依赖包在主分支上有改动）。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/amazing/mgc.gif)
[via GIPHY](https://giphy.com/gifs/shia-labeouf-12NUbkX6p4xOO4)

当用户在使用源码编译他们自己版本的软件时开始针对第三方库报告问题，而你完全不清楚是哪一个提交引起的时候，请尽情享受吧 ^\_^

是的没错， 你可以使用像 [Glide](https://github.com/Masterminds/glide) 或者其它类似的工具来将你的依赖“固定”到某些特定的提交或者标签，并且为他们创建一个特定的目录……这确实是由于一个糟糕的设计而不得不采取的措施，我们都知道这确实行得通，但是这看起来很恶心。

直接使用 [URL 重定向](http://labix.org/gopkg.in)来导入特定版本的包看起来也跟上面差不多……这是可行的，但是同样很难看，而且有些人可能会担心这会引起一些安全方面的问题……谁来控制这些重定向？当你在自己的电脑上使用 root 用户或者 sudo 进行导包或者编译这些东西时，这样一个机制能让你安心工作吗？我想应该不会。

## 反射？我觉得不算是……

当我第一次听说 Go 里面有反射时，根据以往在其他语言（例如 Python，Ruby，Java，C# 和其他语言）上面反射的概念，我想到了它的许多用途（或者说，我认为的 Go 的反射的用处），像是自动枚举 802.11 协议各层的类型，并且依据 WiFi 自动化模糊测试或者其它近似的方式自动生成对应的数据包……事实证明，对于 Go 语言来说反射是一个很大的概念 :D

举个例子，在一个不透明的接口对象中，你可以获取到它的原始类型并且你也可以列出某个特定对象的域，但是你没办法简单地枚举一个特定的包里定义的对象（包括结构体和基本类型），这看起来好像并不重要，但是没有这种特性你完成不了下面这些功能：

1. 构造一个插件系统，它会从给定的包里自动加载内容，而不需要明确地声明（需要加载哪些东西）。
2. 基本上所有你在 Python 里可以用 dir 命令做到的所有事情
3. 构建我想到的 802.11 协议的漏洞检查工具（fuzzer）

由此看出，（Go里面的）反射跟别的语言比起来确实有点有限了……我不清楚你会怎么想，但是这确实让我有点烦……

## 泛型？没有
大部分从面向对象编程的语言（转向 Go 开发时）会抱怨 Go 里缺少泛型，就我个人而言这并不算是一个大问题，因为我自己并不是很热衷于不计代价的面相对象编程。相反，我认为 Go 的对象模型（确切的说并不能算是对象模型）很简洁，我认为这种设计跟泛型会引起的复杂性相冲突了。

**备注**

我并不是想说“泛型==面相对象编程（OOP）”，但是大部分开发者希望（Go 语言支持）泛型是因为他们用 Go 来替代 C++ 并且希望有类似的模板，或者 Java 泛型……我们确实可以讨论从其它具有泛型或者类似东西的功能语言转型的一小部分（开发者），但是就我个人经验来说这部分人并不影响统计。

从另一个方面来看，这种（看起来跟直接使用 C 语言里的功能和结构体很相似的）简化对象模型，会让其他一些事情变得没有其他语言来说那么简单和直接。

假设你正在开发一个包含了许多模块的软件（我喜欢把软件模块化来保证代码足够简洁明了 :D），它们全部都是从同一个基类上派生出来的（这样你就会希望有一个特定的接口并且可以透明地处理它们），并且需要有一些已经实现了的默认功能来在各个派生的模块间共享（这些是所有派生的模块都需要使用的方法，所以为了方便起见它们会在基类中直接被实现）。

好吧，在其他语言里你会有一些抽象类，或者一些已经实现了部分功能（子类共享的方法）其它部分声明为接口（纯虚函数）的类：

```c
class BaseObject {
protected:
  void commonMethod() {
	  cout << "I'm available to all derived objects!" << endl;
  }

  // while this needs to be implemented by every derived object
  virtual interfaceMethod() = 0;
};
```

碰巧 Go 语言就是不支持这种写法，一个类可以是一个接口类或者是一个基础结构体（对象），但是它并不能同时是这两者，所以我们需要把这个例子按照这种方式进行“分离”：

```go
type BaseObjectForMethods struct { }
func (o BaseObjectForMethods) commonMethod() {
	log.Printf("I'm available to all derived objects!\n")
}
type BaseInterface interface {
	interfaceMethod()
}
type Derived struct {
	// I just swallowed my base object and got its methods
	BaseObjectForMethods
}
// and here we implement the interface method instead
func (d Derived) interfaceMethod() {
	// whatever, i'm a depressed object model anyway ... :/
}
```

最终你派生出来的对象会实现里面的接口并且继承基础结构体……尽管这看起来可能一样或者说这是一个尚且算是优雅的解耦方式，但是当你尝试去再稍稍扩展一下 Go 语言的多态性的时候就会发现这很快就会变得一团糟（[这是一个更加实际的例子](https://github.com/bettercap/bettercap/blob/master/session/module.go)）

## Go 很容易编译，但是 CGO 就如地狱一般

编译（和交叉编译）Go 应用非常的简单，不管你是在那种平台编译或者运行。使用同一个 Go 安装包你可以为 Windows，macOS，或者 Android 或者其他基于 GNU/Linux 的 MIPS 设备编译同一个应用，不需要工具链，不需要外部编译器，不需要为操作系统设立特定标记，也没有那些从来都不会按照我们设想来运行的古怪的配置脚本……这简直不要太棒好吗？！（如果你是从 C/C++ 的世界中过来的，并且经常需要交叉编译工程，你就会知道这意味着什么了……或者假设你是一个安全顾问，而你现在需要尽快交叉编译软件，来同时解决你昨天被（病毒）感染的 Windows 域名控制器和 MIPS IP 摄像头）。

好吧，如果你正在使用一些 Go 语言没有原生支持的本地库，你就会发现事情没有那么简单了，除非你仅仅是为了用 Go 来写一个 “hello world”。

让我们假设你的 Go 项目正在使用 libsqlite3，或者 libmysql，或者其他的第三方库，由于那些实现了（你正在使用的 Go API 里的 ）这整套对象-关系映射的人，并没有把 Go 语言里定义的数据库协议都重写，而仅仅重写了其中一些通过 CGO 模块封装的，经历了完善测试的系统库——迄今为止，所有的语言都有自己的封装机制来处理本地库——并且，如果你仅仅只是为了你的主机编译工程的话，这完全没有问题，因为你需要的所有库（libsqlite3.so，libmysql.so 或者其他的库）都可以通过 apt-get install 命令安装。但是如果你需要进行交叉编译呢？比如说需要为 Android 进行编译？如果目标系统里没有默认的库文件呢？当然了，这样你就系统通过对应系统的 C/C++ 工具链来自己编译库文件，或者是找方法把编译器直接安装到系统里然后编译出所有东西（用你的 Android 平板来直接作为编译主机）。那你请好好享受。

无需多言，如果你想要（或者需要）支持多架构跨平台（为什么你不应该认为 Go 最大的优点之一——如我们所说的——恰恰正是这个？），这会让你的编译复杂度大增，进而让你的 Go 项目在交叉编译时至少会和一个 C/C++ 项目一样复杂（讽刺的是，有时甚至会更复杂）。

在[我的一些项目](https://github.com/evilsocket/arc)的某个时刻，我将项目里的所有 sqlite 数据库都替换成了 JSON 文件，这让我摆脱了本地依赖从而构建了一个 100%（基于）Go （编写的）应用。这样依赖交叉编译又重新变得简单了（如果你不能避免使用本地依赖，那么这是你不得不解决的难题……对此我感到十分抱歉 :/）。

如果现在你“聪明的内心“正在尖叫着说“全部使用静态编译！”（静态编译库文件来让它们至少被打包进二进制文件里），不要这样做。如果你用一个特定版本的 glibc（c运行库）来对所有代码进行静态编译，那么编译出来的二进制文件在使用其他版本 glibc 的系统上是无法运行的。

如果你“更聪明的内心“正在尖叫着说“使用 docker 来区分编译版本！”，请找出一个方法来正确的配置所有的平台和所有的（cpu）架构后发邮件告诉我这个方法 :)

如果你的“对 go 语言有点了解的内心”正打算建议一些外部的 glibc 替代品，请参照上一条的需求（如何区分所有配置）:D

## ASLR? 没有!（嘲讽脸）

接下来这个稍微有点争议，[Go 应用的二进制文件没有 ASLR（针对缓冲区溢出的安全保护技术）](https://rain-1.github.io/golang-aslr.html)。但是，根据 Go 的内存管理方式（最重要的是，[它并没有指针算法](https://golang.org/doc/faq#no_pointer_arithmetic)），这并不会成为一个安全问题——除非你使用了有漏洞的本地库文件——这样的情况下 Go 缺少的 ASLR 机制[会让开发变得更容易](http://blog.securitymouse.com/2014/07/bla-bla-lz4-bla-bla-golang-or-whatever.html)。

现在，我有点了解 Go 语言开发者的观点了，但是却不太认同：为什么要在运行时增加复杂度，仅仅用来保证（程序）运行时不会因为某些根本不会被轻易攻击的东西而出问题？……一旦考虑到你最终（在项目里）会使用到的第三方本地库的频率（上文中有对此进行过讨论 :P），我认为直接无视这个问题不是一个明智的选择。

## 总结

还有许多其他关于 Go 的小问题是我不喜欢的，但是那确实也是我了解的其他语言上共有的，所以我仅仅关注了主要的问题而跳过了一些这样的问题：比如我在主观上不喜欢这个语法 X（顺便说一句，我确实喜欢 Go 的语法）。我看到许多的人，盲目的去投身一门新的语言，仅仅是因为在 GitHub 上很流行……从一个方面说，如果许多的开发者都决定使用它，那么这肯定有很充分的理由（或者他们仅仅是“把所有东西都编译成 JavaScript”来追赶时髦的人），但是没有一门完美的语言可以说是所有应用的最佳选择（但是，我对于 nipples 和 default injection 仍旧抱有希望 U.U），在选择之前最好再三比对它们的优缺点。

愿世界和平

---

![](https://blockchain.info/Resources/buttons/donate_64.png)
保持联系！
[关注 @evilsocket](https://twitter.com/evilsocket)

---

via: https://www.evilsocket.net/2018/03/14/Go-is-amazing-so-here-s-what-i-don-t-like-about-it/

作者：[Simone](https://www.evilsocket.net/random-facts-about-me.html)
译者：[林启瀚](https://github.com/keon-lam)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
