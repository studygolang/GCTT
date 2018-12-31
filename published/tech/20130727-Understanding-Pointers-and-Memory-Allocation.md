首发于：https://studygolang.com/articles/17353

# 理解 Go 语言中的指针和内存分配

在 Go 语言官方文档中，你可以找到很多关于指针和内存分配的重要信息。以下是该文档的链接：[Go 语言官方文档之指针](http://golang.org/doc/faq#Pointers)

首先我们需要理解的是，所有在 Go 语言中的值都有其不同变量来表示。不同类型表示的变量决定了我们如何使用它来操纵内存。这篇文章阐述了更多关于 Go 语言中的变量类型：[理解 Go 中的类型](https://studygolang.com/articles/13976)

在 Go 中，我们可以创建一个变量作为 “值” 本身，也可以创建一个变量作为某个 “值” 的地址。当变量的 “值” 是地址时，我们就称该变量为指针。

在下图中，有一个名为 myVariable 的变量。myVariable 的“值”是指向另一个值的地址。则 myVariable 称为指针变量。

![屏幕截图](https://raw.githubusercontent.com/studygolang/gctt-images/master/understanding-pointer-memory/1.png)

在下图中，myVariable 的值是值本身，而不是像上图一样的对值的引用。

![屏幕截图](https://raw.githubusercontent.com/studygolang/gctt-images/master/understanding-pointer-memory/2.png) 要访问值的属性，我们可以使用选择器运算符来访问值的特定字段。选择器运算符的语法为 Value.FieldName，其中句点（.）是选择器运算符。

在 C 语言中，我们需要使用不同的选择器运算符，如何选择取决于我们使用的变量类型。如果变量的“值”是一个单纯的值，我们使用句点（.）。如果变量的“值”是一个地址，我们使用箭头（`->`）。

这一点在 Go 语言中变得非常方便，你不必担心要使用哪种类型的选择器运算符。我们只使用句点（.）就可以了，无论变量是值还是指针。编译器会帮你负责访问该值的一些具体细节。

那么为什么我们还要理解指针和内存分配呢？当我们开始使用函数来抽象和分解逻辑时，这就变得很重要了。当你把这些变量传递给函数时，你需要知道你传递的是什么 " 值 "，这一点很关键，下面我们来详谈。

在 Go 中，每一个变量都以 “值” 的形式传递给函数。这意味着每个变量的“值”都是被复制到堆栈上以供该函数访问。来看一个例子你就能理解了，我们调用一个函数，该函数想要改变在 main 中的值的结果。

```go
package main
import (
	"unsafe"
	"fmt"
)
type MyType struct {
	Value1 int
	Value2 string
}
func main(){
	// 实例化一个 myType 类型
	myValue:= MyType {10,"Bill"}
	// 创建一个指针指向 myValue 的内存
	pointer:= unsafe.Pointer(&myValue)
	// 打印出地址和值
	fmt.Printf("Addr:%v Value1：%d Value2%s\n", pointer, 		 myValue.Value1,myValue.Value2)
	// 给到一个更改 myValue 的函数
	ChangeMyValue(myValue)
	// 打印出地址和值
	fmt.Printf("pointer：%v Value1：%d Value2%s\n",
	pointer,
	myValue.Value1,
	myValue.Value2)
}
func ChangeMyValue(myValue MyType){
	// 更改 myValue 的值
	myValue.Value1 = 20
	myValue.Value2 ="Jill"
	// 找到此 myValue 的地址
	pointer:= unsafe.Pointer(&myValue)
	// 打印出结果
	fmt.Printf("Addr：%v Value1：%d Value2:%s\n",pointer,myValue.Value1,myValue.Value2)
}
```

此程序输出结果如下：

```
pointer：0x2101bc000 Value1：10 Value2：Bill
Addr：0x2101bc040 Value1：20 Value2：Jill
Addr：0x2101bc000 Value1：10 Value2：Bill
```

仔细观察第二三行的输出结果，为什么会这样？在函数调用之后，main 函数的 myValue 的值并没有得到修改。因为 main 中的 myValue 变量的“值”在被函数调用时只是值的拷贝，并不包含对这个值的引用操作，它不是指针而只是一个单纯的值。当我们将 main 中的 myValue 变量的“值”传递给函数时，该值被拷贝后将放置在堆栈上。函数改变的是只是 myValue 拷贝后的副本。一旦函数终止，副本就会弹出堆栈，并且在技术栈上被释放后消失。永远不会影响到 main 中 myValue 变量的“值”。

为了解决这个问题，我们可以用另一种方式分配内存以获取引用操作。我们将 main 中 myValue 变量的地址给到函数中，此函数接收地址后操作此地址达到修改 main 中 myValue 的“值”的目的。

```go
package main
import(
	"unsafe"
	"fmt"
)
type MyType struct {
	Value1 int
	Value2 string
}
func main(){
	// 实例化一个 myType 类型
	myValue:=&MyType{10,"Bill"}
	// 开辟一个指针接收 myValue 的内存地址
	pointer:= unsafe.Pointer(myValue)
	// 打印出地址和值
	fmt.Printf("Addr：%v Value1：%d Value2：%s \n",
	pointer,
	myValue.Value1,
	myValue.Value2)
	// 更改 myValue 的值的函数
	ChangeMyValue(myValue)
	// 打印出地址和值
	fmt.Printf("Addr:%v Value1:%d Value2：%s \n",
	pointer,
	myValue.Value1,
	myValue.Value2)
}
func ChangeMyValue(myValue *MyType){
	// 更改 myValue 的值
	myValue.Value1 =20
	myValue.Value2 ="Jill"
	// 创建一个指向此 myValue 值的地址
	pointer:= unsafe.Pointer(myValue)
	// 打印出地址和值
	fmt.Printf("Addr:%v Value1：%d Value2：%s \n",
	pointer,
	myValue.Value1,
	myValue.Value2)
}
```

当我们使用取地址符号（＆）运算符来分配值时，将返回一个可引用的地址值。这意味着 main 中 myValue 变量的“值”现在是一个指针变量，其值是新分配值的地址。当我们将 main 中的 myValue 变量的“值”传递给函数时，函数中名为 myValue 的变量现在包含的是地址，而不是副本。我们现在有两个指针指向同一块内存空间。main 中的 myValue 变量和函数中的 myValue 变量。

如果我们再次运行程序，该函数现在会达到我们想要的目的，即它会更改 main 中分配的值的状态。

```
Addr：0x2101bc000 Value1：10 Value2：Bill
Addr：0x2101bc000 Value1：20 Value2：Jill
Addr：0x2101bc000 Value1：20 Value2：Jill
```

在函数调用期间，不再在堆栈上复制该值，而是复制该值的地址。该函数现在通过局部指针变量引用同一块内存空间，并更改值。

标题为 "Effective Go" 的 Go 文档有一个关于内存分配的重要部分，其中包括数组，切片和映射的工作方式：

[Effective Go](http://golang.org/doc/effective_go.html#allocation_new)

接着让我们来谈谈关键字 new 和 make。

new 关键字用于在内存中分配指定类型的值。内存分配后被清零。在调用 new 时无法进一步初始化内存。换句话说，使用 new 时，不能为指定类型的属性赋值。

如果要给值分配内存时指定值的内容，请使用复合字面量。无论是否指定字段名称，它们都是可行的。

```go
// 分配 MyType 类型的值
// 值的顺序必须是正确的
myValue:= MyType {10,"Bill"}
// 分配 MyType 类型的值
// 使用标签指定对应的值
myValue:= MyType {
	Value1 :10,
	Value2:"Bill",
}
```

make 关键字仅用于分配和初始化 Slice， Map 和 Channel。它返回一个创建好并初始化的新切片，Map 或管道。Make 返回的值不是指针类型的，但这些数据本身内部就包含了引用这一特性，我们称这些数据结构类型的数据为引用类型。

我们将 Map 传递给函数就可以观察到和第一段代码不同的现象。看看这个示例代码：

```go
package main
import(
	"unsafe"
	"fmt"
)
type MyType struct {
	Value1 int
	Value2 string
}
func main(){
	myMap:= make(map[string]string)
	myMap ["Bill"]="Jill"
	pointer:=unsafe.Pointer(&myMap)
	fmt.Printf("Addr:%v Value：%s \n",pointer,myMap["Bill"])
	ChangeMyMap(myMap)
	fmt.Printf("Addr:%v Value：%s \n",pointer,myMap["Bill"])
	ChangeMyMapAddr(&myMap)
	fmt.Printf("Addr:%v Value: %s \n",pointer,myMap["Bill"])
}
func ChangeMyMap(myMap map[string]string){
	myMap ["Bill"]="Joan"
	pointer:= unsafe.Pointer(&myMap)
	fmt.Printf("Addr:%v Value:%s \n",pointer,myMap["Bill"])
}
// 不要这样做，只是在本文中使用作个实验
func ChangeMyMapAddr(myMapPointer *map[string] string){
	(* myMapPointer)["Bill"] ="Jenny"
	pointer:=unsafe.Pointer(myMapPointer)
	fmt.Printf("Addr%v Value：%s \n",pointer,(*myMapPointer)["Bill"])
}

```

这是该程序的输出：

```
Addr：0x21015b018 Value：Jill
Addr：0x21015b028 Value：Joan
Addr：0x21015b018 Value：Joan
Addr：0x21015b018 Value：Jenny
Addr：0x21015b018 Value：Jenny
```

我们创建一个 Map 并添加一个名为 "Bill" 的键，值为 "Jill"。然后我们将 map 变量的值传递给 ChangeMyMap 函数。请记住，myMap 变量不是指针，因此在函数调用期间，myMap 的“值”将被复制到堆栈中。因为 myMap 的“值”是包含对 Map 内部的引用的一种数据结构，所以该函数可以使用其副本来更改主函数的 Map 的值。

如果查看输出，可以看到当我们直接传递 Map 时，该函数在函数内创建了自己的 Map 的副本用来操作。你可以看到在函数调用后对 Map 所做的更改生效了。

虽然使用 ChangeMyMapAddr 是没有必要的，但这个函数显示了如何在 main 中传递和使用 myMap 变量的地址进行操作。Go 的设计者再次确保传递 Map 变量的“值”可以毫无问题地执行。注意当我们想要访问 Map 时，我们需要对 myMapPointer 变量进行解引用，这是因为 Go 编译器不允许我们直接通过指针变量访问 Map。对指针变量进行解引用，相当于获取它所指向的变量值。

我花时间写这篇文章的原因，是因为有时候你的变量的“值”包含什么，这一点可能会经常令人很困惑。如果你的变量的“值”占用很多内存，那么当把变量的“值”传递给函数时，你会在堆栈上花费巨大的开销来创建该变量的副本。除非有很特殊的需求，否则你一定会希望将值的地址传递给函数以完成调用。

Map，Slice 和 Channel 这些数据类型和普通的数据类型是不同的。你可以按值传递这些变量而无需担心。当我们将 map 变量传递给函数时，我们其实正在复制一个数据结构，而不是它所指向的整个哈希表。

最后我建议你阅读上面提到的 [Effective Go](https://golang.org/doc/effective_go.html) 文档。自从我开始使用 Go 编程以来，我已多次阅读该文档。随着我的 Go 开发经验与日俱增，我还总是时不时回头温习之。总是有一种温故而知新的感觉！

---

via: https://www.ardanlabs.com/blog/2013/07/understanding-pointers-and-memory.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[CNbluer](https://github.com/CNbluer)
校对：[Noluye](https://github.com/Noluye)

本文由 [GCTT](https://github.com/studyGolang/GCTT) 原创编译，[Go 中文网](https://studyGolang.com/) 荣誉推出
