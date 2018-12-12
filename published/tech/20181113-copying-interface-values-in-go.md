首发于：https://studygolang.com/articles/16770

# Go 中接口值的复制

我一直在思考 Go 语言它是如何工作的。直到最近我才发现 Go 中一切都是基于值的。当我们向函数传递参数、迭代切片、执行类型断言时我们都可以看到这一现象。在这些例子中，这些数据结构所存储的值的拷贝会被返回。当我刚开始学习 Go 的时候，我对于这种实现方式很失望，但渐渐地我开始意识到这样做对于我们的代码来说有它的合理性。

我开始在想，如果我创建了一个所存储的是值而非指针的接口类型的拷贝会发生什么。那么这个新的接口值会拥有自己新的副本值，还是共享原来的值？为了验证我的猜想，我写了一个小程序来检查接口值。

https://play.golang.org/p/KXvtpd9_29

## 清单 1

```go
06 package main
07
08 import (
09     "fmt"
10     "unsafe"
11 )
12
13 // notifier provides support for notifying events.
14 type notifier interface {
15     notify()
16 }
17
18 // user represents a user in the system.
19 type user struct{
20     name string
21 }
22
23 // notify implements the notifier interface.
24 func (u user) notify() {
25     fmt.Println("Alert", u.name)
26 }
```

在清单 1 的代码中第 14 行，我们声明了一个具有 `notify` 方法的接口类型 `notifier`。在 19 行我们声明了一个 `user` 类型，这个类型在 24 行实现了 `notifier` 接口中的 `notify` 方法。现在我们拥有了一个接口类型和一个实现类型。

## 清单 2

```go
28 // inspect allows us to look at the value stored
29 // inside the interface value.
30 func inspect(n *notifier, u *user) {
31     Word := uintptr(unsafe.Pointer(n)) + uintptr(unsafe.Sizeof(&u))
32     value := (**user)(unsafe.Pointer(word))
33     fmt.Printf("Addr User: %p  Word Value: %p  Ptr Value: %v\n", u, *value, **value)
34 }
```

在清单 2 中我们在 30 行写了一个检查的函数。这个函数返回给我们一个指向接口值的第二个字的指针。通过这个指针我们可以检查此接口第二个字的值，以及第二个字指向的 `user` 值。我们需要审查这些值来真正理解接口的机制。

## 清单 3：

```go
36 func main() {
37
38     // Create a notifier interface and concrete type value.
39     var n1 notifier
40     u := user{"bill"}
41
42     // Store a copy of the user value inside the notifier
43     // interface value.
44     n1 = u
45
46     // We see the interface has its own copy.
47     // Addr User: 0x1040a120  Word Value: 0x10427f70  Ptr Value: {bill}
48     inspect(&n1, &u)
49
50     // Make a copy of the interface value.
51     n2 := n1
52
53     // We see the interface is sharing the same value stored in
54     // the n1 interface value.
55     // Addr User: 0x1040a120  Word Value: 0x10427f70  Ptr Value: {bill}
56     inspect(&n2, &u)
57
58     // Store a copy of the user address value inside the
59     // notifier interface value.
60     n1 = &u
61
62     // We see the interface is sharing the u variables value
63     // directly. There is no copy.
64     // Addr User: 0x1040a120  Word Value: 0x1040a120  Ptr Value: {bill}
65     inspect(&n1, &u)
66 }
```

清单 3 展示了从 36 行开始的主函数。 我们做的第一件事情就是在 39 行声明了类型为 `notifier` 名为 `n1` 的接口变量，将其设为初始零值。然后在 40 行我们声明了一个类型为 `user` 的变量 `u`，其初始值为 `bill`。

我们想要在将具体 `user` 值分配给接口之后，检查这个接口值的第二个字包含什么。

## 清单 4

```go
42     // Store a copy of the user value inside the notifier
43     // interface value.
44     n1 = u
45
46     // We see the interface has its own copy.
47     // Addr User: 0x1040a120  Word Value: 0x10427f70  Ptr Value: {bill}
48     inspect(&n1, &u)
```

## 插图 1：当我们将 `user` 的值分配给该接口之后，这个接口内部结构是什么样的。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/copy-interface-value/69_figure1.png)

插图 1 向我们展示了在分配后接口的内部结构视图。我们可以看到这个接口有它自己的 `user` 值的拷贝。存储在接口值内部的 `user` 值的地址和原始 `user` 值的地址并不相同。

我编写这段代码用来了解如果我创建了一个分配了值而非指针的接口值的副本会发生什么。新的接口值是否也有自己的副本，或者值是否会共享。

## 清单 5

```go
50     // Make a copy of the interface value.
51     n2 := n1
52
53     // We see the interface is sharing the same value stored in
54     // the n1 interface value.
55     // Addr User: 0x1040a120  Word Value: 0x10427f70  Ptr Value: {bill}
56     inspect(&n2, &u)
```

## 插图 2：当接口值被拷贝后，这两个接口值的内部构造。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/copy-interface-value/69_figure2.png)

插图 2 给了我们答案。当我们创建一个接口值的拷贝时，这就是我们复制的全部内容。原来存储在 `n1` 中的 `user` 值此时也共享给了 `n2`。每个接口值并没有获得它们独有的拷贝。它们共享了同一个 `user` 值。

如果我们分配的是 `user` 值的地址而不是该值本身，那么所有这些情况都会有所改变。

## 清单 6

```go
58     // Store a copy of the user address value inside the
59     // notifier interface value.
60     n1 = &u
61
62     // We see the interface is sharing the u variables value
63     // directly. There is no copy.
64     // Addr User: 0x1040a120  Word Value: 0x1040a120  Ptr Value: {bill}
65     inspect(&n1, &u)
```

## 插图 3： 当我们为接口分配了一个地址时，该接口值的内部构造。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/copy-interface-value/69_figure3.png)

在插图 3 中，我们看到该接口现在指向的是被变量 `u` 引用的 `user` 值。 变量 `u` 的地址已经存储到这个接口值的第二个字当中。这就意味着和 `user` 值相关的任何变化都会影响到这个接口值。

## 结论

我允许自己对 Go 在创建一个接口值的拷贝时，只存储值而不是指针的情况感到困惑。有那么一瞬间，我想知道接口值的每个副本是否也创建了接口引用的值的副本。因为我们“存储”了一个值而不是一个指针。但我们现在学到的是，由于地址始终会被存储起来，因此它是被复制的地址，而不是值的本身。

---

via: https://www.ardanlabs.com/blog/2016/05/copying-interface-values-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[barryz](https://github.com/barryz)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
