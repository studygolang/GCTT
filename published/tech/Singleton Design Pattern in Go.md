已发布：https://studygolang.com/articles/12728

# Go 中的单例设计模式

多线程应用程序非常复杂，尤其是当你的代码没有组织并且与资源访问、管理和维护保持一致时。如果你想最大限度地减少错误，你需要哲学和规则来生活。这里有一些我的：

- 资源的分配和回收应该在同一类型中抽象和管理
- 资源线程安全性应该在同一类型中抽象和管理
- 公共接口应该是访问共享资源的唯一手段
- 任何分配了资源的线程都应该释放同类型的资源

在 Go 中，没有线程，只有 `Go Routines`。Go 运行时抽象了这些例程的线程和任务交换。无论如何，相同的哲学和规则也适用。

我最喜欢的设计模式之一是单例模式（Singleton）。当你只需要一个类型的实例并且该类型管理共享资源时，它提供了一个很好的实现。访问由该引用管理的共享资源是通过静态公共接口抽象出来的。这些静态方法还提供了线程安全性。使用 Singleton 的应用程序负责初始化和销毁 Singleton，但不能直接访问内部。

在一段时间内，我回避如何在 Go 中实现一个单例的问题，因为 Go 不是传统的面向对象的编程语言，也没有静态方法。

我认为 Go 是一个轻量级的面向对象编程语言。是的，它确实具有封装和类型成员函数，但缺乏继承性，因此缺乏传统的多态性。在我曾经使用过的所有 OOP 语言中，除非我想实现多态性，否则我从来没有使用过继承。 在 Go 中实现接口的方式不需要继承。 Go 取了 OOP 最好的部分，移除了剩下的部分，并给了我们一个更好的方式来编写多态代码。

在 Go 中我们可以利用包和类型的作用域和封装规则来实现 Singleton，对于这篇文章，我们将探索我的 straps 包，因为它将给我们一个现实世界的例子。

straps 包提供了一种机制来将配置选项（straps）存储在XML文档中，并将其读入内存以供应用程序使用。straps 的名称来自配置网络设备的早期阶段。 这些设置被称为 straps，并且这个名字一直伴随着我。 在 MacOS 中，我们有 .plist 文件，在 .Net 中我们有 app.config 文件，在 Go 中有 straps.xml 文件。

以下是我的一个应用程序的示例 straps 文件：

```xml
<straps>
	<!– Log Settings –>
	<strap key="baseFilePath" value="/Users/bill/Logs/OC-DataServer">
	<strap key="machineName" value="my-machine">
	<strap key="daysToKeep" value="1">

	<!– ServerManager Settings –>
	<strap key="cpuMultiplier" value="100">
</straps>
```

straps 包知道如何读取这个 xml 文件，并通过基于 Singleton 的公共接口提供对这些值的访问。 由于这些值只需要读入内存中，因此 Singleton 对于这个包来说是一个很好的选择。

以下是 straps 包和类型信息：

```go
package straps

import (
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

.
. Types Removed
.

type straps struct {
	StrapMap map[string]string // The map of strap key value pairs
}

var st straps // A reference to the singleton
```

我不会谈论读取 XML 文档的内容。 如果你有兴趣，请阅读这篇博文 http://www.goinggo.net/2013/06/reading-xml-documents-in-go.html。

在上面的代码片段中，您将看到包名称（straps），私有类型 straps 的定义和私有包变量 st。st 变量将包含 Singleton 的值。

Go 的作用域规则声明以大写字母开头的类型和函数是公开的，并且可以在包外部访问。以小写字母开头的类型和函数是私有的，在包之外不可访问。

我使用小写字母命名在函数作用域内定义的变量。在函数作用域之外定义的变量名称，例如类型成员和包变量以大写字母开头。这使我可以查看代码并立即知道引用了哪个给定变量的内存。

straps 类型和 st 变量都是私有的，只能从包中访问。

查看初始化 Singleton 以供使用的 Load 函数：

```go
func MustLoad() {
	// Find the location of the straps.xml file
	strapsFilePath, err := filepath.Abs("straps.xml")

	// Open the straps.xml file
	file, err := os.Open(strapsFilePath)
	if err != nil {
		panic(err.Error())
	}

	defer file.Close()

	// Read the straps file
	xmlStraps, err := readStraps(file)
	if err != nil {
		panic(err.Error())
	}

	// Create a straps object
	st = straps{
		StrapMap: make(map[string]string),
	}

	// Store the key/value pairs for each strap
	for _, strap := range xmlStraps {
		st.StrapMap[strap.Key] = strap.Value
	}
}
```

Load 函数是包的公有函数。应用程序可以通过包名来访问这个函数。你可以看到我如何使用以局部变量的小写字母开头的名称。在 Load 函数的底部创建一个 straps 对象，并将该引用设置为 st 变量。在这一点上，Singleton 存在并且 straps 已经可以使用了。

使用公有函数 Strap 访问 straps：

```go
func Strap(key string) string {
	return st.StrapMap[key]
}
```

公有函数 Strap 使用 Singleton 引用来访问共享资源。在这个例子中，就是 straps 的字典映射。 如果字典映射在应用程序的生命周期内可能发生变化，则需要使用互斥锁或其他同步对象来保护字典映射。 幸运的是，straps 一旦被装载就不会改变。

由于由 straps 管理的资源只是内存，因此不需要 Unload 或 Close 方法。如果我们需要一个函数来关闭任何资源，则必须创建另一个公有函数。

如果在 Singleton 包中需要私有方法来帮助组织代码，我喜欢使用成员函数。 由于类型是私有的，我可以使成员函数公开，因为它们不可访问。 我也认为成员函数有助于使代码更具可读性。 通过查看函数是否是成员函数，我知道该函数是私有的还是公共接口的一部分。

```go
func SomePublicFunction() {
	.
	st.SomePrivateMemberFunction("key")
	.
}

func (straps *straps) SomePrivateMemberFunction(key string) {
	return straps.StrapMap[key]
	.
}
```

由于函数是一个成员函数，我们需要使用 st 变量来进行函数调用。从成员函数内使用局部变量（straps）而不是 st 变量。成员函数是公共的，但引用是私有的，所以只有包可以引用成员函数。这只是我为自己建立的一个习惯。

下面是一个使用 straps 包的示例程序：

```go
package main

import (
	"ArdanStudios/straps"
)

func main() {
	straps.MustLoad()
	cpu := straps.Strap("cpuMultiplier")
}
```

在 main 中我们不需要分配任何内存或保持引用。通过包名，我们调用 Load 来初始化 Singleton。然后再次通过包名称访问公共接口，在这种情况下，是 Strap 函数。

如果你有通过公共接口尝试来管理共享资源的需求，就使用 Singleton。

和往常一样，我希望这可以帮助您更好地编写代码并减少 bug。

---

via: https://www.ardanlabs.com/blog/2013/07/singleton-design-pattern-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[shniu](https://github.com/shniu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

