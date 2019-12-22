# 为什么Go关心unsafe.Pointer和uintptr之间的差别
Go有两样东西或多或少是无类型指针的表示；uintptr和unsafe.Pointer （与外观相反，它是内置类型）。
从表面上看这有点奇怪，因为unsafe.Pointer和uintptr可以在彼此来回转换。
为什么不只有一种指针表现形式？两者之间有什么区别？

表面的区别是可以对uintptr进行算数运算但不能对unsafe.Pointer（或任何其他Go指针）进行运算。
unsafe包的文档指出了重要的区别：
> uintptr是整数，不是引用。将Pointer转换为uintptr会创建一个没有指针语义的整数值。
> 即使uintptr持有某个对象的地址，如果对象移动，垃圾收集器并不会更新uintptr的值，uintptr也无法阻止该对象被回收。

尽管unsafe.Pointers是通用指针，但Go垃圾收集器知道它们指向Go对象；换句话说，他们是真正的Go指针。
通过内部魔术机制，垃圾收集器能够并且会使用它们来方式活动对象被回收同时发现更多活动对象（如果unsafe.Pointer指向的对象自身持有指针）。
因此，对unsafe.Pointers的合法操作上的许多限制归结为“在任何时候，它们都必须指向真正的Go对象”。
如果创建的unsafe.Pointer并不符合，即使很短的时间，Go垃圾收集器也可能会在该时刻扫描，然后由于发现了无效的Go指针而崩溃。

相比之下，uintptr只是一个数字。这种特殊的垃圾收集魔术机制并不适用于uintptr所“引用”的对，因为它仅仅是一个数字，一个uintptr不会引用任何东西。
反过来，这导致在将unsafe.Pointer转换为uintptr，对其进行操作然后再将其转回的各种方式上存在许多谨慎的限制。
基本要求是以这种方式进行操作，使编译器和运行时可以屏蔽不安全的指针的临时非指针性，使其免受垃圾收集器的干扰，因此这种临时转换对于垃圾收集将是原子的。

（我想，我在条目[将内存块复制到Go结构中](https://utcc.utoronto.ca/~cks/space/blog/programming/GoMemoryToStructures)里对unsafe.Pointer的使用是安全的，但我承认我现在不确定。
我相信cgo会有一些不可思议的机制，因为它可以安全地制造出不安全的指针，这些指针指向C内存而不是Go内存。）

PS：从Go 1.8开始，即使当时没有运行垃圾回收，所有Go指针必须始终有效（我相信也包括unsafe.Pointer）。
如果您在变量或字段中存储了无效的指针，则仅通过将字段更新为包括nil在内的完全有效的值即可使代码崩溃。
例如，请参阅[这个有教育意义的Go bug report](https://github.com/golang/go/issues/19135)。

（我本想尝试讲一下内部魔术，它允许垃圾收集器处理未类型化的unsafe.pointer指针，但我坚信我对其了解不足，甚至无法说出它使用的是哪种魔术。 ）

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoUintptrVsUnsafePointer

作者：[ChrisSiebenmann](https://utcc.utoronto.ca/~cks/space/People/ChrisSiebenmann)
译者：[dust347](https://github.com/dust347)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
