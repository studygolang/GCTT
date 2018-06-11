已发布：https://studygolang.com/articles/12400

# unsafe.Pointer 和系统调用

按照 Go 语言官方文档所说, unsafe 是关注 Go 程序操作类型安全的包。

像包名暗示的一样，使用它要格外小心； unsafe 可以特别危险，但它也可以特别有效。例如，当处理系统调用时，Go 的结构体必须和 C 的结构体拥有相同的内存结构，这时你可能除了使用 unsafe 以外，别无选择。

unsafe.Pointer 可以让你无视 Go 的类型系统，完成任何类型与内建的 uintptr 类型之间的转化。根据文档，unsafe.Pointer 可以实现四种其他类型不能的操作：

* 任何类型的指针都可以转化为一个 unsafe.Pointer
* 一个 unsafe.Pointer 可以转化成任何类型的指针
* 一个 uintptr 可以转化成一个 unsafe.Pointer
* 一个 unsafe.Pointer 可以转化成一个 uintptr

这里主要关注两种只能借助 unsafe 包才能完成的操作：使用 unsafe.Pointer 实现两种类型间转换和使用 unsafe.Pointer 处理系统调用。

## 使用 unsafe.Pointer 做类型转换

### 操作方式

可以简洁适宜的转换两个在内存中结构一样的类型是使用 unsafe.Pointer 的一个主要原因。

文档描述：

> 如果T2与T1一样大，并且两者有相同的内存结构；那么就允许把一个类型的数据，重新定义成另一个类型的数据

经典的例子，是文档中的一次使用，用来实现 math.Float64bits：

```go
func Float64bits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}
```

这似乎是一种非常简洁的完成这样转换的方法，但是这个过程中具体发生了什么？让我们一步步拆分一下：

* &f 拿到一个指向 f 存放 float64 值的指针。
* unsafe.Pointer(&f) 将 *float64 类型转化成了 unsafe.Pointer 类型。
* (*uint64)(unsafe.Pointer(&f)) 将 unsafe.Pointer 类型转化成了 *uint64。
* *(*uint64)(unsafe.Pointer(&f)) 引用这个 *uint64 类型指针，转化为一个 uint64 类型的值。

第一个例子是下面过程的一个简洁表达：

```go
func Float64bits(floatVal float64) uint64 {
	// 获取一个指向存储这个float64类型值的指针。
	floatPtr := &floatVal

	// 转化*float64类型到unsafe.Pointer类型。
	unsafePtr := unsafe.Pointer(floatPtr)

	// 转化unsafe.Pointer类型到*uint64类型.
	uintPtr := (*uint64)(unsafePtr)

	// 解引用成一个uint64值
	uintVal := *uintPtr

	return uintVal
}
```
这是一个非常有用的操作，有些时候也是一个必要操作。
现在你已经理解了 unsafe.Pointer 是如何使用的，那么让我们再看一个真实的项目例子

### 现实列子：taskstats

我最近正在研究 [Linux 的 taskstats 接口](https://www.kernel.org/doc/Documentation/accounting/taskstats.txt)，我想了一个办法在 Go 中取到了内核的 C 的 taskstats 结构。然后发送一个 CL 把这个结构加到 x/sys/unix 中，我意识到这个结构实际上是如此的[庞大和复杂](https://godoc.org/golang.org/x/sys/unix#Taskstats)。

为了使用这个结构，我需要从一个 byte 类型的切片中精确的分析每一个字段。更复杂的是，每一个 integer 类型在本地有序的存储，所以这些整数可能根据你的cpu在内存中以不同的格式存储。

这个情况就非常适合使用简洁的 unsafe.Pointer 转换，下面是我的写法：

```go
// 通过这个包证实包含一个 unix.Taskstats 结构的byte类型的切片是预计的大小，我们不能盲目的将这个byte类型的切片放入一个错误尺寸的结构中。
const sizeofTaskstats = int(unsafe.Sizeof(unix.Taskstats{}))

if want, got := sizeofTaskstats, len(buf); want != got {
	return nil, fmt.Errorf("unexpected taskstats structure size, want %d, got %d", want, got)
}

stats := *(*unix.Taskstats)(unsafe.Pointer(&buf[0]))
```
它是怎么做的？

首先，我通过参数传来的结构体实例，使用 unsafe.Sizeof，确定了该结构在内存中占有的准确的大小。

接下来，我确认需要转换的byte类型的切片大小和 unix.Taskstats 结构大小一样，这样我就可以只读取我想要的数据块，而不是随意读取内存。

最后，我使用 unsafe.Pointer 向 unix.Taskstats 结构转换。

但是，我为什么必须指定切片索引的0位置呢？

如果你了解[切片的内部结构](https://blog.golang.org/go-slices-usage-and-internals)，你将知道一个切片实际上是一个头和一个指向底层数组的指针。当使用 unsafe.Pointer 来转换切片数据时，必须指定数组第一个元素的内存地址，而不是切片本身的首地址。

使用 unsafe 使得转换非常的简洁、简单。因为整型数据根据我们的CPU以相同的字节顺序存储，使用 unsafe.Pointer 转化意味着整型值是我们预期的。

你可以去看看我 [taskstats](https://github.com/mdlayher/taskstats) 包中的代码。

## 使用 unsafe.Pointer 处理系统调用

### 操作方式

当处理系统调用时，有些时候需要传入一个指向某块内存的指针给内核，以允许它执行某些任务。这是 unsafe.Pointer 在Go中另一个重要的使用场景。当需要处理系统调用时，就必须使用 unsafe.Pointer ，因为为了使用 syscall.Syscall 家族函数，它可以被转化成 uintptr 类型。

对于许多不同的操作系统，都拥有大量的系统调用。但是在这个例子中，我们将重点关注 ioctl 。ioctl，在UNIX类系统中，经常被用来操作那些无法直接映射到典型的文件系统操作，例如读和写的文件描述符。事实上，由于 ioctl 系统调用十分灵活，它并不在Go的 syscall 或者 x/sys/unix 包中。

让我看看另一个真实的例子。

### 现实例子：ioctl/vsock

在过去的几年里，Linux增加了一个新的 socket 家族，AF_VSOCK，它可以使管理中心和它的虚拟机之间双向，多对一的通信。
这些套接字使用一个上下文ID进行通信。通过发送一个带有特殊请求号的 ioctl 到 /dev/vsock 驱动，可以取到这个上下文ID。

下面是 ioctl 函数的定义：

```go
func Ioctl(fd uintptr, request int, argp unsafe.Pointer) error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		fd,
		uintptr(request),
		// 在这个调用表达式中，从 unsafe.Pointer 到 uintptr 的转换是必须做的。详情可以查看 unsafe 包的文档
		uintptr(argp),
	)
	if errno != 0 {
		return os.NewSyscallError("ioctl", fmt.Errorf("%d", int(errno)))
	}

	return nil
}
```

像代码注释所写一样，在这种场景下使用 unsafe.Pointer 有一个很重要的说明：

> 在 syscall 包中的系统调用函数通过它们的 uintptr 类型参数直接操作系统，然后根据调用的详细情况，将它们中的一些转化为指针。换句话说，系统调用的执行，是其中某些参数从 uintptr 类型到指针类型的隐式转换。
> 如果一个指针参数必须转换成 uintptr 才能使用，那么这种转换必须出现在表达式内部。

但是为什么会这样？这是编译器识别的特殊模式，本质上是指示垃圾收集器在函数调用完成之前，不能将被指针引用的内存再次安排。

你可以通过阅读文档来获得更多的技术细节，但是你在Go中处理系统调用时必须记住这个规则。事实上，在写这篇文章时，我意识到我的代码违反了这一规则，现在已经被修复了。  

意识到这一点，我们可以看到这个函数是如何使用的。

在 VM 套接字的例子里，我们想传递一个 *uint32 到内核，以便它可以把我们当时的上下文ID赋值到这块内存地址中。

```go
f, err := fs.Open("/dev/vsock")
if err != nil {
	return 0, err
}
defer f.Close()

// 存储上下文ID
var cid uint32

// 从这台机器的 /dev/vsock 中获取上下文ID
err = Ioctl(f.Fd(), unix.IOCTL_VM_SOCKETS_GET_LOCAL_CID, unsafe.Pointer(&cid))
if err != nil {
	return 0, err
}

// 返回当前的上下文ID给调用者
return cid, nil
```

这只是在系统调用时使用 unsafe.Pointer 的一个例子。你可以使用这么模式发送、接收任何数据，或者是用一些特殊方式配置一个内核接口。有很多可能的情况！

你可以去看看我 [vsock](https://github.com/mdlayher/vsock) 包中的代码。

## 结尾

虽然使用 unsafe 包可能存在风险，但当使用恰当时，它可以是一个非常强大、有用的工具。

既然你在读这篇文章，我建议你在你的程序使用它之前，去[读一下 unsafe 包的官方文档](https://golang.org/pkg/unsafe/)。

如果你有任何问题，请随时联系我！在 [Gophers Slack](https://gophers.slack.com/), [GitHub](https://github.com/mdlayher) and [Twitter](https://twitter.com/mdlayher)上我的名字是 mdlayher。

非常感谢 [Hazel Virdó](https://twitter.com/HazelVirdo) 对这篇文章的建议和修改！

## 链接
* [unsafe 包](https://golang.org/pkg/unsafe/)
* [字节顺序](https://en.wikipedia.org/wiki/Endianness)
* [taskstats 包](https://github.com/mdlayher/taskstats)
* [Go中Slice的使用和内部实现](https://blog.golang.org/go-slices-usage-and-internals)
* [ioctl](https://en.wikipedia.org/wiki/Ioctl)
* [vsock 包](https://github.com/mdlayher/vsock)

---

via: https://blog.gopheracademy.com/advent-2017/unsafe-pointer-and-system-calls/

作者：[Matt Layher](https://github.com/mdlayher)
译者：[yiyulantian](https://github.com/yiyulantian)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
