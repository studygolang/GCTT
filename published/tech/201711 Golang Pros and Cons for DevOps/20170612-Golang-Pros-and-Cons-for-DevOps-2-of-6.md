已发布：https://studygolang.com/articles/12608

# Go 对于 DevOps 之利弊(六部曲之二)：接口实现和公开/私有指定

在这系列的第二篇文章，我们讨论接口实现(优势)和公共/私有设计(一个明显的劣势)。

如果你错过了上一篇关于 goroutines 和 panics/error 的文章，请务必将它补上。你也可以订阅我们的博客，以及时获得更新状况通知。(大概隔周更新一次)

- [Golang 之于 DevOps 开发的利与弊（六部曲之一）：Goroutines, Channels, Panics, 和 Errors](https://studygolang.com/articles/11983)
- [Golang 之于 DevOps 开发的利与弊（六部曲之二）：接口实现的自动化和公有/私有实现](https://studygolang.com/articles/12608)
- [Golang 之于 DevOps 开发的利与弊（六部曲之三）：速度 vs. 缺少泛型](https://studygolang.com/articles/12614)
- [Golang 之于 DevOps 开发的利与弊（六部曲之四）：time 包和方法重载](https://studygolang.com/articles/12615)
- [Golang 之于 DevOps 开发的利与弊（六部曲之五）：跨平台编译，Windows，Signals，Docs 以及编译器](https://studygolang.com/articles/12616)
- Golang 之于 DevOps 开发的利与弊（六部曲之六）：Defer 指令和包依赖性的版本控制

## Go 语言的优势：接口实现

Go 的 automagic 接口实现让人印象深刻。第一个原因就是它能让我们从依赖的地狱中解脱。

### Go 的接口是如何运作的？

与大多数语言不同，Go 中的接口是由与接口定义匹配的结构自动实现的。这是一个简单的例子。注意在 `Dog` 和 `Cat` 结构上明显的关键词缺失。

```go
package main

import (
	"fmt"
)

type Animal interface {
	Speak()
}
type Cat struct {
}

func (this *Cat) Speak() {
	fmt.Println("meow")
}

type Dog struct {
}

func (this *Dog) Speak() {
	fmt.Println("woof")
}

func main() {
	var pet Animal
	pet = &Cat{}
	pet.Speak()
	pet = &Dog{}
	pet.Speak()
}
```

大多数主流语言(c++、c#、Java、PHP 等)都需要类来指定要实现的接口。Go 则采取了不同的方法。俗话说，如果它叫声像鸭子，那它就是鸭子。这听起来是否有点像 Python 或 Ruby 的鸭子类型 ? 尽管表面上相同，但是 Go 具有一个关键优势：它仍然与编译器一起工作。您不必在代码中执行运行时检查项目是否通过了「鸭子」式的测试。

### 避开「依赖地狱」

现在，让我们来证明这种接口方法避免了循环依赖。最简单的解释是，它减少了依赖项的数目。

你可以对比一下你最喜欢的语言的数据库，以 Java 为例，JDBC 在 java.sql 包中有 ResultSet、Connection、Driver、Statement 以及另外十八个接口，所有这些接口都必须实现。

我们再来看看 MySQL、SQLServer、Oracle 和 Postgres库。我没有编写或检查它们是否使用 java.sql 中定义的任何具体类或枚举。不过，这些供应商库不需要 JDBC 包来使用上面提到的接口。

我们在 Blue Matador 建立的监控代理有很多包和模块。每个模块可以主动或被动启用。现在，它被分为 Lumberjack，我们的集中式日志管理产品，以及 Watchdog，我们的系统的监控工具。

此外，还有一个负责注册和模块管理的基本代理。你可以想象，我们有大量的通用代码、共享包和依赖项需要管理。

为了给你们找到一个比较好的例子，我在代理代码库(以及使用这些接口的每个实例)中搜索了每个接口，然而我并没找到。我所能做的就是粘贴我们的导入语句，然后说，「它怎么没提到 XYZ 事件?」这既没有说服力也没有趣味性，也不是很有趣。所以，我认为这是更好的选择。

## Go 劣势：公开/私有的指定（Designations）

另一方面，我已经因为改变 Go 中的一个函数或变量的访问修饰符而精神错乱。我不知道为什么 Golang 开发人员选择用变量和函数名的第一个字符, 而不是使用 public 和 private 关键字来控制访问权限。如果它是大写的, 它是公开的。如果它是小写的, 它是私有的。

### 编码规范

首先，这导致了半不可读的代码。编码规范之所以存在，是为了使代码更加统一和可读。除了空白，命名规范是所有规范中最热门的讨论话题。而 Go 的理念是，「省去七个字符比高可读性更重要。」

你可能会问，「这不就已经解决了规范问题吗?」在某种程度上，它确实解决了程序员争论的问题。但它却让代码的可读性下降很多，这也与代码规范设定之初的目的背道而驰。

你可能又会说，「我完全能读懂它，因为它十分简洁。」那我只能说「我的眼睛不如你的好。」

### 改变访问修饰符

第二，更重要的是，当你想要从私有变为公开或从公开变为私有时，它不再是一个单一的关键字，而是一个分布式的实体引用或者整个包(私有变为公开时)，甚至是整个代码库(公开变为私有时)。

这是因为名称可以被复制(例如，`key`、`id` 等等)、嵌入(例如，`abuser` 内 `user`)，并立即被操作符所遵循(例如，`user:=1`)。你无法通过全词、大小写敏感的方式搜索来找到您想更改名称的正确实例！改变它的唯一方法是改变源码, 然后走 编译/修复 这条路。

公开/私有的关键字的增加不会改变产生的二进制文件大小。Go 开发人员选择牺牲我们重构那些 kB 的文本的速度，我们不会察觉这些变动，除非我们用到了它们。

但如果把所有这些时间都加起来，你会发现这并不是一个很短的时间。而相对于它所节省的那一点点磁盘空间而言，我认为这是一个糟糕的决定，是严重的缺陷。

每隔一周，我们都将更新我们的系列博客《Go 在 DevOps 中的优势与劣势》

----------------

via: https://blog.bluematador.com/posts/golang-pros-cons-for-devops-part-2/

作者：[Matthew Barlocker](https://github.com/mbarlocker)
译者：[Mr.NoFat](https://github.com/UnFat)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
