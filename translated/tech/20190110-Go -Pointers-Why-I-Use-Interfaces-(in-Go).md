# Go知识点-为什么在Go中要使用接口

* * *

如果你已经写过一段时间的代码了，那我可能就不需要解释接口的优点了，相信你也应该深有感触。但是在深入研究为什么我在Go中使用接口之前，我需要花一点时间来介绍一下

如果你对接口足够了解，那你可以跳过这段

## Level setting

通过接口，我们可以整合一组方法或者行为，这个在任何编程语言里面都是想通的。这样的话，相当于在功能的定义和消费者之间设置了一层抽象。书写接口的一方，不需要实现接口方法的业务细节。这样保证了各个组件功能的清晰。

你可以通过接口实现很多别的方式无法实现的酷炫功能。例如，你创建了很多组件，他们通过某段代码以相同的方式进行交互，组件再大也无所谓。这样的话，在我们切换组件的时候，就可以无缝衔接，甚至可以在运行时动态的交换。

在Go中的 `io.Reader` 就是一个真实的例子，所有实现`io.Reader`的接口，都需要提供一个`Read(p []byte) (n int, err error)`  的方法, 实现` io.Reader` 接口的调用方不需要知道 byte数组从什么地方传过来的。

以上所说的所有内容，对于从事过一段时间开发工作的人来说，这都是常识。

## Go中的接口

在Go中，使用情况比其他语言多很多，这里我要介绍一些编码中，跟接口有关的常见问题。

## Go没有构造器方法

很多语言都会提供一个构造器方法，构造器允许用户自定义一些类型的实例化规则(在很多面向对象的语言中，成为类),保证一些操作在初始化的过程中完成。
例如，很多对象都会有一个不可变得，系统分配的唯一标识符。这个在Java中很容易实现：
```
package io.krancour.widget;

import java.util.UUID;

public class Widget {

	private String id;

	// A constructor that performs some initialization!
	public Widget() {
		id = UUID.randomUUID().toString();
	}

	public String getId() {
		return id;
	}

}
 
 class App {

		public static void main( String[] args ){
			Widget w = new Widget();
			System.out.println(w.getId());
		}

}
```

在这里，你不可能跳过他的实例化方法，创建一个新的`Widget` 对象，**但是Go却没有这个功能**

* * *

在Go中，可以直接实例化用户自定义对象
给一个例子：
```
package widgets

type Widget struct {
	id string
}

func (w Widget) ID() string {
	return w.id
}

package main

import (
	"fmt"
	"github.com/krancour/widgets"
)

func main() {
	w := widgets.Widget{}
	fmt.Println(w.ID())
}

```
如果你执行上面的代码，就会发现得到的结果跟我们预期的不一样，打印出来的是一个空字符串。因为我们并没有实例化ID的值，所以ID是的值是string的"零值"。

我们可以添加一个方法，让他达到我们预期的效果：
```
package widgets

import uuid "github.com/satori/go.uuid"

type Widget struct {
	id string
}

func NewWidget() Widget {
	return Widget{
		id: uuid.NewV4().String(),
	}
}

func (w Widget) ID() string {
	return w.id
}

package main

import (
	"fmt"

	"github.com/krancour/widgets"
)

func main() {
	w := widgets.NewWidget()
	fmt.Println(w.ID())
}

```

执行上面的代码，我们得到了想要的结果。

**但是仍然有一个巨大的问题**，我们没有办法组织用户通过默认的方式实例化Widget结构体

## Go 的私有化
当我们通过"类似构造器"方法初始化我们的Widgets结构体的时候，首先要确保他是私有的。在Go中，一个类型或者函数等的结构大小在开始的时候就被确定了(对其他包可见的), 首字母大写的是共有的，首字母小写的是私有的。因此我们要把Widget结构体修改为widget：
```
package widgets

import uuid "github.com/satori/go.uuid"

type widget struct {
	id string
}

func NewWidget() widget {
	return widget{
		id: uuid.NewV4().String(),
	}
}

func (w widget) ID() string {
	return w.id
}
```

我们的Main函数不需要修改，程序也能正常执行。这个结果跟我们预期的结果越来越接近了，但是`NewWidget() `方法返回的是一个私有的结构体类型。虽然编译器没有报错，但是这始终是一个不太好的方式，需要一些额外的解释。

在Go语言中。package是一个基本单位(其他语言一般把类作为一个单位)。就像前面说的一样，任何私有结构体(首字母小写)的内容都应该是私有的。私有的内容应该包括很多实现的细节，而且不应该对外部调用的对象产生任何影响。（在这里如果私有结构体改变了属性，那调用方就不对了）而且，在这里，`godoc`指令也不会为私有的类型生成文档。

我们构造了一个类似构造器的，返回私有类型实例的方法。无意中为文档内容创建了一个"死胡同"。 调用这个方法虽然可以获取实例，但是无法调用里面的`ID()`方法，而且也不能获取到更多关于`widget`的相关细节。Go社区内非常重视文档，因此这个方式可能不行。

## 使用Interface来救场

直到现在，我们一直在通过构建一个类似构造函数的方式来获取实例化对象，为了保证用户始终调用这个方法构建对象，我们用了私有化结构体类型。尽管编译器允许这样做，但是文档的问题却无法解决。接下来我们会通过接口的方式来完善它。

首先创建一个的公有的接口类型，然后`widget`来实现这个接口。完美的实现了我们之前的所有要求。
```
package widgets

import uuid "github.com/satori/go.uuid"

// Widget is a ...
type Widget interface {
	// ID returns a widget's unique identifier
	ID() string
}

type widget struct {
	id string
}

// NewWidget() returns a new Widget
func NewWidget() Widget {
	return widget{
		id: uuid.NewV4().String(),
	}
}

func (w widget) ID() string {
	return w.id
}
```

## 包装
Go语言缺乏构造函数，因此我们需要通过接口的方式来帮助完成这件事，否则他们可能无法被调用，希望我的介绍，已经充分包括了这部分的所有细节。

我的下一篇文章中，将会介绍一个 关于接口的，在其他语言很常见的使用情景，但是在Go中却无法做到的事情。

via: https://medium.com/@kent.rancourt/go-pointers-why-i-use-interfaces-in-go-338ae0bdc9e4

作者：[Kent Rancourt](https://medium.com/@kent.rancourt)
译者：[JYSDeveloper](https://github.com/JYSDeveloper)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出