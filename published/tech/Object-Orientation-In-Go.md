已发布：https://studygolang.com/articles/12640

# Go 语言中的面向对象

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/object-orientation/gophermessages.jpg)

> 什么是面向对象呢？

这里我试着分享我对 go 语言如何实现面向对象的理解，以及让它看起来比其他传统的面向对象语言更加面向对象。

在之前的一篇文章中，我探讨了用函数来表达一切的想法，而实际上只是用一种安全的方式来表示一组函数，它总是与相同的闭包一起运行，并且可以在相同的状态下运行。

安全地表示在同一个状态下运行的一组函数是非常有用的，但是如何真正的创建具有多种不同实现的抽象呢？

对象本身不支持这一点，每种对象类型都有它自己的函数集（我们称这组函数为方法）。这被称为多态，但是当我从传统的面向对象的概念（关于对象和类型）离开，我发现，在设计软件时，考虑协议而不是多态性更符合应该关注的内容（稍后会详细介绍）。

## 如何定义一个协议？

首先让我们定义一下什么是协议，对于我来说协议就是实现预期结果所需的一系列操作。

这个概念似乎是晦涩难懂的，让我举个例子，一个更容易理解的关于一个抽象概念的栗子，它就是 I/O。

假如你想读取一个文件，协议将是：

* 打开文件
* 读取文件的内容
* 关闭文件

为了实现读取一个文件所有内容的简单需求，你需要这三个操作，因此这些操作就构成了你“阅读文件”的协议。

现在让我们来看看这个例子，并且一起完成剩余的实现。

如果函数在一门编程语言中是一等公民，结构化函数和结构化数据并无区别。

我们有一种合成数据的方法，那就是结构体(structs)，我们也可以使用这种方法来合成函数，例如：

```go
type Reader func(data []byte) (int, error)

type Closer func() error

type ReadCloser struct {
	Read Reader
	Close Closer
}

type Opener func() (*ReadCloser, error)
func useFileProtocol(open Opener) {
	f, _ := open()
	data := make([]byte, 50)
	f.Read(data)
	f.Close()
}

func main() {
	useFileProtocol(func() (*ReadCloser, error) {
		return &ReadCloser{}, nil
	})
}
```

使用编译时安全的方法来表达一个协议是非常困难的（假如这并不是不可能的）。为了说明这个问题，这个例子造成了一个段错误。

另一个问题是实现这个协议的代码需要知道协议被显示地实现，以便正确地初始化结构体（就像继承的过程一样），或者把结构体的初始化委托给系统的其他部分，该部分将围绕如何正确的初始化结构体展开。当你考虑实现多种协议的相同函数时，将会变得更加糟糕。

需要一些第三个协议的对象需要一种方法：

* 清楚的表达出它所需要的协议。
* 确保当它开始与一个实现交互时，没有任何功能丢失。

实现服务的对象需要：

* 能够安全地表示它具有满足协议的必须功能。
* 能够满足一个协议，甚至对它不是很清晰的了解。

不需要两个对象交互来实现一个特定类型的共同目标，最重要的是它们之间的协议是否匹配。

这就是 Go interfaces（接口）的由来。它提供了一个编译时安全的方式来表现协议，通过适当的函数来消除初始化结构的所有模板。它会为我们初始化结构，甚至对初始化结构优化一次，这在 Go 中被称为 iface 。类似于 C++ 的 vtable 。

它还允许代码更加的解耦，因为你不需要了解定义这个接口 （interface）并且实现这个接口的包。与相同的编译安全型语言Java、C++比较，在 Go 允许的基础上，Go更加的灵活。

让我们重新审视之前的带有接口（interfaces）的文件协议：

```go
package main

type Reader interface {
	Read(data []byte) (int, error)
}

type Closer interface {
	Close() error
}

type ReadCloser interface {
	Reader
	Closer
}

type Opener func() (ReadCloser, error)

type File struct {}

func (f *File) Read(data []byte)(int, error){
	return 0, nil
}

func (f *File) Close() error {
	return nil
}

func useFileProtocol(open Opener) {
	f, _ := open()
	data := make([]byte, 50)
	f.Read(data)
	f.Close()
}

func main(){
	useFileProtocol(func() (ReadCloser, error) {
		return &File{}, nil
	})
}
```

一个关键的不同是, 这段代码使用了接口（interface），现在是安全的。那个`useFileProtocol`不必担心调用函数是否为nil，go编译器将会创建一个结构体，通过一个 `iface`描述符来保持一个指针，该指针具有满足协议的所有功能。它会按类型<->接口的每一个匹配项执行此操作，就像它被使用的那样（它第一次初始化使用的那样）。

如果你这样做，仍然会造成一个段错误，如下：

```go
useFileProtocol(func() (ReadCloser, error) {
	var a ReadCloser
	return a, nil
})
```

如果它编译，就可以确定调用接口的所有函数是安全的。还有 Go 的接口机制，一个对象可以在不知道这些协议的情况下实现多个不同的协议。

实现协议是真的有用的吗？你甚至都不知道它的存在。如果你希望代码真正具有可扩展性，那么这是非常有用的。让我提供一个来自 nash 的真实案例 (nash : 一个 GitHub 开源库，https://github.com/NeowayLabs/nash)。

## 代码扩展超出其最初的目的

当我尝试为 nash 上的内置函数 exit 编写测试代码时，第一次了解到 go 的 interfaces 是多么的强大。主要的问题是，似乎我们必须为每个平台实现不同的测试，因为在某些平台上，退出状态代码的处理方式不同。我现在不记得所有的细节，但在 plan9 上，退出状态是一个 string 类型，而不是一个 integer 类型。

基本上在一个错误上，我想要的是状态代码，而不仅仅是错误，就像在 Cmd.run 上提供的。(文件 Cmd.run : https://golang.org/pkg/os/exec/#Cmd.Run)

有 ExitError 类型，我可以这样做: (ExitError 类型： https://golang.org/pkg/os/exec/#ExitError)

```go
if exiterr, ok := err.(*exec.ExitError); ok {
}
```

至少知道，这是一个通过在我执行的过程中的一些错误状态产生的一些错误，但我还是找不到实际的状态码。

所以我继续沉浸在 Go 的源代码库中，去查找如何从 ExitError 中得到错误状态码。

我的线索是 ProcessState ，ProcessState 是 ExitError 结构体的内部组成（ExitError 是 stuct 类型）。方法 Sys 就说明了:

```go
// Sys returns system-dependent exit information about
// the process. Convert it to the appropriate underlying
// type, such as syscall.WaitStatus on Unix, to access its contents.
func (p *ProcessState) Sys() interface{} {
	return p.sys()
}
```

interface{} 什么也说明不了，但随着它的线程，我发现一个 posix 的实现:

```go
func (p *ProcessState) sys() interface{} {
	return p.status
}
```

并且 p.status 是什么呢，在 posix 中：

```go
type WaitStatus uint32
```

用有趣的方法：

```go
func (w WaitStatus) ExitStatus() int {
	if !w.Exited() {
		return -1
	}
	return int(w>>shift) & 0xFF
}
```

但是这是针对 posix ，其他平台呢？使用这种先进的检测技术:

```
syscall % grep -R ExitStatus .
./syscall_nacl.go:func (w WaitStatus) ExitStatus() int    { return 0 }
./syscall_bsd.go:func (w WaitStatus) ExitStatus() int {
./syscall_solaris.go:func (w WaitStatus) ExitStatus() int {
./syscall_linux.go:func (w WaitStatus) ExitStatus() int {
./syscall_windows.go:func (w WaitStatus) ExitStatus() int { return int(w.ExitCode) }
./syscall_plan9.go:func (w Waitmsg) ExitStatus() int {
```

看起来像公共协议被足够多的平台所实现，这对我来说至少是足够的(windows + linux + plan9 是足够的。)。现在我们有一个共同的协议所有的平台我们可以这样做:

```go
// exitResult is a common interface implemented by
// all platforms.
type exitResult interface {
	ExitStatus() int
}

if exiterr, ok := err.(*exec.ExitError); ok {
	if status, ok := exiterr.Sys().(exitResult); ok {
		got := status.ExitStatus()
		if desc.result != got {
			t.Fatalf("expected[%d] got[%d]", desc.result, got)
		}
	} else {
		t.Fatal("exit result does not have a  ExitStatus method")
	}
}
```

完整的代码可以在这里找到。（https://github.com/NeowayLabs/nash/blob/c0cdacd3633ce7a21714c9c6e1ee76bceecd3f6e/internal/sh/builtin/exit_test.go）

Sys() 方法返回了一个抽象的、更加精确的接口，这将是更容易的想出一个新的接口，这将是一个接口的子集，并且能保证编译时的安全，而不是通过检查运行时得到的运行时安全。

但即使是一个简单的方法来定义一个新的 interface 并且不改变源代码的情况下执行一个安全的运行时检查，这样实现 interface 是很简洁的。在 Java 或 c++ 语言中，我不能想出一个解决方案，包含相同数量的代码/复杂性，特别因为基于多态性的层次结构的脆性的。检查只被允许这种情况，如果原始代码知道你正在检查的接口，并且明显的是继承自它。为了解决我的问题我不得不改变 go 的核心代码，去了解我的 interface ，在 Go 的 interfaces 中，这个不是必需的( yay 层次结构)。

这是非常重要的, 因为它允许开发人员提出简单的对象, 因为它们不必预测对象将来可能使用的每一种方式, 就像接下来可能会有哪些接口有用一样。

只要你的对象的协议是明确的、有用的，它就可能被重用在几个你从未想过有可能的方面。你甚至不需要显式地表达接口定义和使用。

## 面向对象到底是什么呢?

这一部分，我将试着去谈一谈 Go 作为一个了不起的面向对象的语言。所有的早期我所接触的编程语言，像：

* Java
* C++
* Python

并且了解了继承，多重继承，菱形继承(这是 C++ 中多重继承中的一个问题，即两个父类继承于同一个类。)等。主要的关注点在类型和继承树，一个分类的练习。什么是类型和继承树。就像创造一个良好的分类，将是一个良好的面向对象设计。进一步的我开始和那些说面向对象的人讨论面向对象并不是关于这个的(尽管所有主流的面向对象语言)，这样的设计并不灵活。我不明白,但引起了我的好奇。理解它的最佳机会是最接近它的核心，所以我去查找了关于 Alan Kay 的面向对象的资料。(Alan Kay 天才计算机大师阿伦凯，他是 Smalltalk 面向对象编程环境语言的发明人之一，也是面向对象编程思想的创始人之一。)

还有更可怕的东西可以从他那里学到的，但他在 OOPSLA 有过面向对象的主题演讲，(演讲主题是 The computer revolution has not happened yet 。地址：https://www.youtube.com/watch?v=oKg1hTOQXoY)，他讨论了一点关于面向对象的起源(在其他事物之中)。

他说面向对象之间应该关注什么是对象,而不是对象本身。他甚至说，面向更多的过程的名称会更好，因为关注对象似乎已经产生了关注类型和分类，而不是这个对象之间实际存在的东西，对我来说，这是协议。

思考对象的重要部分是封装(不是类型)。他给出了一个很好的例子，那就是细胞，他们有明确的细胞膜，他们允许出去，也允许进入。

彼此交互的每一个细胞都对彼此内部运作一无所知，他们不需要知道其他细胞类型，他们只需要实现相同的协议，交易相同的蛋白质等(我不擅长生物学=))。重点是化学反应(过程)而不是细胞类型。

所以我们最终封装和明确的协议作为面向对象应该是什么，如何开发系统和一个伟大的隐喻，模仿生物机制,而是因为有机生命尺度数量级更好。

另一个伟大的隐喻可以从编程演示的未来中提取出来。当谈到 ARPANET 和 "星际计算机网络(Intergalactic Computer Network)" 的开始时, 用来表达系统如何真正扩展的隐喻之一是软件将如何与其他完全陌生的软件集成 (甚至来自其他星球)。

隐喻是伟大的，因为它显示了良好的协议和形式需要做内容/协议谈判，这是自然发生了什么,也许总有一天会发生什么，如果我们遇见外星生命(希望我们不要愚蠢地战斗到死)。

这个比喻甚至为动态语言提供了一些点，但老实说，我对这一点没有理解，这是足够好的，现在提出的东西 (我发现很难思考的东西真的适应没有动态，甚至在 Go 中，你需要一些胶水，通过手动创建的协议集成对象的代码)。

现在这个比喻最重要的一点, 不是最好代表这种系统, 而是要有正确的心态去寻找可能的答案, 从演讲中引述：

```
The most dangerous thought that you can have as a creative person is to
think that you know what you are doing
```

尽管我试着保持开放的心态, 但当我想到使用上面的隐喻进行编程时, 我找不到继承的空间。外星人的生命永远不会融合, 因为他们需要一个可能根本不存在的共同祖先。

## 结论

上面我给的例子已经显示了一眼在 Go 中如何做更多，而无需改变任何预先存在的代码使用 协议(Go的interfaces) 的概念代替类。似乎更容易开发根据开闭原则（开闭原则：open closed principle，https://en.wikipedia.org/wiki/Open/closed_principle），因为我可以轻松地扩展其他代码去做事情，这不是最初打算不用改变它。

Go 和 Java 都有 interfaces 的概念，这是似乎具有误导性的，因为他们唯一的共同点是他们的名字。在 Java 中接口的创建是一个关系，在 Go 中并不是。它只是定义了一个协议，可用于直接集成的对象不知道对方(他们甚至不需要知道明确的接口)。这似乎更加面向对象，比任何到目前为止我知道的(当然,我并不知道太多=))。

## 鸣谢

特别感谢:

* i4k (https://github.com/tiago4orion)
* kamilash (https://github.com/kamilash)

花时间回顾并指出很多愚蠢的错误。

---

via: https://katcipis.github.io/blog/object-orientation-go/

作者：[TIAGO KATCIPIS](https://katcipis.github.io/)
译者：[MengYP](https://github.com/MengYP)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
