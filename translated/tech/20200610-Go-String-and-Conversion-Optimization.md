# Go：字符串以及转换优化

![由Renee French创作的原始Go Gopher作品，为“ Go的旅程”创作的插图。](https://github.com/studygolang/gctt-images2/blob/master/20200610-Go-String-and-Conversion-Optimization/Illustration.png?raw=true)

ℹ️  这篇文章基于 Go 1.14。

在 Go 语言中，将 byte 数组转换为 string 时，随着转换后字符串的拷贝，可能会触发内存分配。然而，将 bytes 转换为 string 仅仅是为了满足代码约束，比如在 switch 语句中的比较，又比如在 map 中的 key，这些场景下的转换绝对是在浪费 CPU 时间。来一起看一些案例，以及一些已有的优化。

## 转换操作（Conversion）
将 byte 数组转换为 string 涉及的操作有：

- 如果变量超过了当前堆栈帧的作用域，在堆上为新的 string 分配内存。
- bytes 到 string 的拷贝

*关于逃逸分析的更多细节，建议阅读我的文章：“[Go：介绍逃逸分析。](https://medium.com/a-journey-with-go/go-introduction-to-the-escape-analysis-f7610174e890)”*

这是完成这两个步骤的简要程序：

![](https://github.com/studygolang/gctt-images2/blob/master/20200610-Go-String-and-Conversion-Optimization/a-simple-program.png?raw=true)

这是该转换操作的示意图：

![](https://github.com/studygolang/gctt-images2/blob/master/20200610-Go-String-and-Conversion-Optimization/diagram-of-conversion.png?raw=true)

*如果想更多了解 copy 函数，建议阅读我的文章“[Go：切片以及内存管理](https://medium.com/a-journey-with-go/go-slice-and-memory-management-670498bb52be)”*

在运行时层面，Go 在转换期间只提供一种优化。如果转换的 byte 数字实际只包含一个字节，返回的 string 会指向一个静态的 byte 数组，该数组嵌入在运行时中：

![](https://github.com/studygolang/gctt-images2/blob/master/20200610-Go-String-and-Conversion-Optimization/point-to-a-static-array-of-byte.png?raw=true)

然而，如果这个 string 之后被修改，分配新值之前会从堆上面分配内存。

Go 编译器同样提供一些优化，可以省略我们所见到的两个转换阶段。

## Switch
先以一个以比较为目的，转换为 string 的示例开始：

![](https://github.com/studygolang/gctt-images2/blob/master/20200610-Go-String-and-Conversion-Optimization/an-example-of-conversion-to-string.png?raw=true)

*这个用来说明字符串优化的实例通过使用 `getBytes` 函数强制在堆上进行分配。这样避免了要介绍的字符串优化被编译器的其他优化所隐藏。*

在这个示例中，仅有 `switch` 指令使用了转换，而且由于仅仅需要与实际内容进行比较，Go 可以避免转换操作。Go 实际上通过移除转换操作，并且直接指向底层的 byte 数组来优化这段代码。

![](https://github.com/studygolang/gctt-images2/blob/master/20200610-Go-String-and-Conversion-Optimization/pointing-directly-to-the-backed-array-of-bytes.png?raw=true)

我们也可以通过生成的汇编指令来了解具体优化细节：

![](https://github.com/studygolang/gctt-images2/blob/master/20200610-Go-String-and-Conversion-Optimization/the-exact-optimization.png?raw=true)

Go 在比较操作中直接使用返回的 bytes。首先比较 byte 数组和 `case` 语句（case 后面的字符串）的大小，之后检查字符串本身（字面值）。在 `switch` 语句外分配 string，会导致内存的分配，因为编译器无法得知这个 string 后续是否还会使用。

## 优化
`switch` 并不是字符串转换的唯一的一个优化。Go 编译器会在其他示例中应用这样的优化，比如：

- 访问 map 中的元素。这是一个例子：

![](https://github.com/studygolang/gctt-images2/blob/master/20200610-Go-String-and-Conversion-Optimization/Accessing-to-an-element-of-a-map.png?raw=true)

当访问 map 时，实际上不需要进行转换，这样能使访问更快。

- 字符串连接。这是一个例子：

![](https://github.com/studygolang/gctt-images2/blob/master/20200610-Go-String-and-Conversion-Optimization/String-concatenation.png?raw=true)

byte 数组与一些 string 的连接不会引起任何内存分配，也不会引起 byte 的任何转换。就像前面看到的一样，连接会直接引用底层的数组。

- 字符串比较。这里是一些例子：

![](https://github.com/studygolang/gctt-images2/blob/master/20200610-Go-String-and-Conversion-Optimization/String-comparisons.png?raw=true)

这个例子与 `switch` 类似。首先比较 string 的大小和 byte 数组的大小，之后再比较字符串。

---
via: https://medium.com/a-journey-with-go/go-string-conversion-optimization-767b019b75ef

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[dust347](https://github.com/dust347)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
