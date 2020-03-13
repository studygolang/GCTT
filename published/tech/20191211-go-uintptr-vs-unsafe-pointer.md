首发于：https://studygolang.com/articles/25931

# 为什么 Go 关心 unsafe.Pointer 和 uintptr 之间的差别

Go 有两样东西或多或少是无类型指针的表示：uintptr 和 unsafe.Pointer （和外表相反，它们是内置类型）。
从表面上看这有点奇怪，因为 unsafe.Pointer 和 uintptr 可以彼此来回转换。为什么不只有一种指针表现形式？两者之间有什么区别？

表面的区别是可以对 uintptr 进行算数运算但不能对 unsafe.Pointer（或任何其他 Go 指针）进行运算。unsafe 包的文档指出了重要的区别：

> uintptr 是整数，不是引用。将 Pointer 转换为 uintptr 会创建一个没有指针语义的整数值。即使 uintptr 持有某个对象的地址，如果对象移动，垃圾收集器并不会更新 uintptr 的值，uintptr 也无法阻止该对象被回收。

尽管 unsafe.Pointer 是通用指针，但 Go 垃圾收集器知道它们指向 Go 对象；换句话说，它们是真正的 Go 指针。通过内部魔法，垃圾收集器可以并且将使用它们来防止活动对象被回收并发现更多活动对象（如果unsafe.Pointer指向的对象自身持有指针）。因此，对 unsafe.Pointer 的合法操作上的许多限制归结为“在任何时候，它们都必须指向真正的 Go 对象”。如果创建的 unsafe.Pointer 并不符合，即使很短的时间，Go 垃圾收集器也可能会在该时刻扫描，然后由于发现了无效的 Go 指针而崩溃。

相比之下，uintptr 只是一个数字。这种特殊的垃圾收集魔法机制并不适用于 uintptr 所“引用”的对象，因为它仅仅是一个数字，一个 uintptr 不会引用任何东西。反过来，这导致在将 unsafe.Pointer 转换为 uintptr，对其进行操作然后再将其转回的各种方式上存在许多微妙的限制。基本要求是以这种方式进行操作，使编译器和运行时可以屏蔽不安全的指针的临时非指针性，使其免受垃圾收集器的干扰，因此这种临时转换对于垃圾收集将是原子的。

（我想，我在文章[将内存块复制到 Go 结构中](https://utcc.utoronto.ca/~cks/space/blog/programming/GoMemoryToStructures)里对 unsafe.Pointer 的使用是安全的，但我承认我现在不确定。
我相信 cgo 会有一些不可思议的机制，因为它可以安全地制造出不安全的指针，这些指针指向 C 内存而不是 Go 内存。）

PS：从 Go 1.8 开始，即使当时没有运行垃圾回收，所有 Go 指针必须始终有效（我相信也包括 unsafe.Pointer）。如果您在变量或字段中存储了无效的指针，则仅通过将字段更新为包括 nil 在内的完全有效的值即可使代码崩溃。例如，请参阅[这个有教育意义的 Go bug report](https://github.com/golang/go/issues/19135)。

（我本想尝试讲一下内部魔法，它允许垃圾收集器处理未类型化的 unsafe.Pointer 指针，但我认为对其了解不足，甚至无法说出它使用的是哪种魔法。 ）

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoUintptrVsUnsafePointer

作者：[ChrisSiebenmann](https://utcc.utoronto.ca/~cks/space/People/ChrisSiebenmann)
译者：[dust347](https://github.com/dust347)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
