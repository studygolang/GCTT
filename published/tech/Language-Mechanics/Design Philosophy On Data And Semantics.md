已发布：https://studygolang.com/articles/12487

# Go 语言机制之数据和语法的设计哲学（Design Philosophy On Data And Semantics）

## 前序（Prelude）

本系列文章总共四篇，主要帮助大家理解 Go 语言中一些语法结构和其背后的设计原则，包括指针、栈、堆、逃逸分析和值或者指针传递。这是最后一篇，重点介绍在代码中使用值和指针的数据和语义的设计哲学。

以下是本系列文章的索引：

1. [Go 语言机制之栈与指针](https://studygolang.com/articles/12443)
2. [Go 语言机制之逃逸分析](https://studygolang.com/articles/12444)
3. [Go 语言机制之内存剖析](https://studygolang.com/articles/12445)
4. [Go 语言机制之数据和语法的设计哲学](https://studygolang.com/articles/12487)

## 设计哲学（Design Philosophies）

"在栈上保存值，这减少了垃圾收集器（GC）的压力。然而，却要求存储、跟踪和维护给定值的多个副本。将值放在堆上，这会给 GC 增加压力。但是它也是有用的，因为只需要针对一个值进行存储、跟踪和维护。" - Bill Kennedy

对于给定类型的数据，想在整个软件中保持完整性和可读性，使用值或者指针要保持一致。为什么？因为，如果你在函数间传递数据时修改数据语义，将很难维护一个清晰一致的心智模型。代码库和团队越大，越多的 bug、对数据的竞争和其他副作用就会悄悄地潜入到代码库中。

我想从一组设计哲学开始讨论，它将指导（我们如何）选择一种语义而不是另外一种语义的方法。

## 心智模型（Mental Models）（译者注：心智模型是经由经验及学习，脑海中对某些事物发展的过程，所写下的剧本。可以当成对代码整体的把控）

"让我们想象有这样一个项目，它包含一百万行以上的代码量。这些项目当前在美国能成功的可能性很低，远低于 50%。或许有人不同意这个说法。" - Tom Love (inventor of Objective C)

Tom 还说一盒复印纸可以容纳 10 万行代码。稍微想一下。你能掌控这个盒子中的代码的百分之多少呢？

我相信要一个开发人员维护一张纸上的代码的心智模型（大约 1 万行代码）已经是个问题。但是，我们还是假设每个开发人员开发 1 万行代码，那么需要由 100 位开发人员组成的团队来维护一个包含 100 万行代码的代码库。也就是说 100 人需要协调，分组，跟踪和不断沟通。现在，再看看你们 1 到 10 名开发人员组成的团队。你们在这个小得多的规模做得如何？假设每人 1 万行代码，（你们）团队规模与代码库的大小是否相符？

## 调试（Debugging）

"最大的问题是你的心智模型是错误的，所以你根本找不到问题所在。" - Brian Kernighan

我不相信，你能在没有心智模型的基础上，使用调试器解决问题，你只不过是在浪费时间精力尝试理解问题。

如果你在生产环境中遇到问题，你能问谁？没错，日志。如果日志在你开发过程中对你没有用，那么当生产环境上出问题，它也一定对你没有用。日志应该基于代码的心智模型，这样才能通过阅读代码找到问题所在。

## 可读性（Readability）

C 语言是我见过的在性能和表达性上平衡得最好的。你可以通过简单的编程实现任何你想要做的事情，并且你会对机器即将要发生的事情拥有一个非常好的心智模型。你可以非常合理地预测它的速度，你知道即将要发生什么..." - Brian Kernighan

我相信 Brian 这句话也适用于 Go。保持这种 "心智模型" 就是一切。它驱动完整性，可读性和简单性。这些是精心编写的软件的基石，使得它可以保持正常并持续运行下去。编写保证给定类型数据的值或者指针语义一致的代码是实现这一点的重要方法。

## 面向数据设计（Data Oriented Design）

"如果你不了解这些数据，你就不明白这个问题。因为所有的问题都是独特的，并且与你所使用的数据关系紧密。当数据发生变化时，你的问题也会跟着变化。但问题发生变化时，你的算法（数据转换）也需要跟着变化。" - Bill Kennedy

想一想。你解决问题的方法实际上是解决数据转换的问题。你写的每个函数，运行的每个程序，（只不过）都是获取一些输入数据，产生一些输出数据。从这个角度看，你的软件的心智模型就是对这些数据转换的理解（例如，如何在代码中组织和使用它们）。"少即是多" 的原则对于解决问题时实现较少的层数，代码量，迭代次数，以及降低复杂性和减少工作量非常重要。

## 类型（就是生命）（Type (Is Life)）

"完整性意味着每次分配内存，读取内存和写入内存都是准确，一致和高效的。类型系统对于我们具有这种微观完整性至关重要。" - William Kennedy

如果数据驱动你所做的一切，那么代表数据的类型就十分地重要。在我的观点里面 "类型就是生命"，因为类型为编译器提供了确保数据完整性的能力。类型也驱动并指示语义规则，程序必须遵循其所操作的数据的语义。这是正确地使用值或者指针语义的开始：使用类型。

## 数据（的能力）

"当数据是实际和合理的，方法才是有效的。" - William Kennedy

值或者指针语义的思想不会直接影响 Go 开发人员，除非他们需要决定方法接收值还是指针。这是我遇到的一个问题：我应该使用值作为参数还是指针？一听到这个问题，我就知道这个开发人员没有理解好这些（类型的）语义。

方法的目的是使这些数据具有某种能力。想象一下，数据有能力做某些事情。我总是希望把重点放在数据上，因为它驱动程序的功能。数据驱动你写的算法，封装和能达到的性能。

## 多态（Polymorphism）

"多态意味着你写了一个特定的程序，但它的行为有所不同，具体取决于它所操作的数据。" - Tom Kurtz (inventor of BASIC)

我很喜欢 Tom 上面说的话。函数的行为可以根据操作的数据的不同而不同。这个数据的行为是将函数从它们可以接受和使用的具体数据类型中分离出来的，这是数据可以具有某种能力的原因。这个观点是使得架构和设计可以适应变化的系统的基石。

## 原型的第一种方法（Prototype First Approach）

"除非开发人员对软件会被如何使用有一个很好的了解，否则软件很可能会出问题。如果开发人员不是很了解或者对软件不是很理解，那么获得尽可能多的用户输入和用户级测试就相当的重要。" - Brian Kernighan

我希望你始终专注于理解具体的数据和为了解决问题所需要的数据转换的算法。采用这种原型的第一种方法，编写也可以在生产环境中部署的具体实现（如果这样做是合理和实际的话）。一旦一个具体的实现已经能够工作，一旦你已经知道哪些工作起作用，哪些不起作用，就应该关注于重构，将实现与具体数据分离，将之赋予数据以能力（译者注：我的理解，简单地说，就是抽象为数据类型的一个方法）。

## 语义原则（Semantic Guidelines）
										 
你在声明类型时，必须决定特定数据类型将使用哪种语义，值或者指针。接收或返回该类型数据的 API 必须遵循为该类型选择的语义。API 不允许（用户）指定或改变语义。他们必须知道数据使用什么语义，并符合这一点。这是实现大型代码库一致性的起码要求。

以下是基本指导原则：

- 当你声明一个类型时，你必须决定所使用的语义
- 函数和方法必须遵循给定类型所选择的语义
- 避免让方法接收与给定类型相对应的不同语义
- 避免函数接收或者返回与给定类型相对应的不同语义
- 避免改变给定类型的语义

这些指导原则有一些例外的情况，最大的是 unmarshaling。Unmarshaling 总是需要使用指针语义。Marshaling 和 unmarshaling 似乎总是例外的规则。

你如何选择一种给定类型的一种语义而不是另外一种？这些指导方针将回答这个问题。以下我们将在具体的情况下使用指导原则：

## 内置类型

Go 语言中内置类型包括数字，文本和布尔类型。这些类型应该使用值语义进行处理。除非你有非常好的理由，否则不要使用指针来共享这些类型的值。

作为一个例子，从 strings 包中查看这些函数的声明。

### 代码清单 1

```go
func Replace(s, old, new string, n int) string
func LastIndex(s, sep string) int
func ContainsRune(s string, r rune) bool
```

所有这些函数在 API 设置中都使用值语义。

## 引用类型

Go 语言中引用类型包括切片，map，接口，函数和 channel。这些类型建议使用值语义，因为它们被设计成待在栈中以最小化堆的压力。它们允许每个函数都有自己的值副本，而不是每个函数都会造成潜在的分配。这是可能的，因为这些值包含一个在调用之间共享底层数据结构的指针。

除非你有很好的理由，否则不要用指针共享这些类型的值。将调用栈中的 map 或 slice 共享给 Unmarshal 函数可能是一个例外。作为一个例子，看看 net 库上声明的这两种类型。

### 代码清单 2

```go
type IP []byte
type IPMask []byte
```

IP 和 IPMask 都是字节切片。这意味着它们是引用类型，并且它们应该要符合值语义。下面是一个名叫 Mask 的方法，它被声明为接收一个 IPMask 值的 IP 类型。

### 代码清单 3

```go
func (ip IP) Mask(mask IPMask) IP {
	if len(mask) == IPv6len && len(ip) == IPv4len && allFF(mask[:12]) {
		mask = mask[12:]
	}
	if len(mask) == IPv4len && len(ip) == IPv6len && bytesEqual(ip[:12], v4InV6Prefix) {
		ip = ip[12:]
	}
	n := len(ip)
	if n != len(mask) {
		return nil
	}
	out := make(IP, n)
	for i := 0; i < n; i++ {
		out[i] = ip[i] & mask[i]
	}
	return out
}
```

请注意，此方法是一种转变操作，并使用值语义的 API 样式。它使用 IP 值作为接收方，并根据传入的 IPMask 值创建一个新的 IP 值并将其返回给调用方。该方法遵循对引用类型使用值语义（的基本指导原则）。

这跟系统默认的 append 函数有点相似。

### 代码清单 4

```go
var data []string
data = append(data, "string")
```

append 函数的转变操作使用值语义。将切片值传递给 append，并在变化之后返回一个新切片值。

总是除了 unmarshaling，它需要使用指针语义。

### 代码清单 5

```go
func (ip *IP) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*ip = nil
		return nil
	}
	s := string(text)
	x := ParseIP(s)
	if x == nil {
		return &ParseError{Type: "IP address", Text: s}
	}
	*ip = x
	return nil
}
```

UnmarshalText 实现 encoding.TextUnmarshaler 接口。如果没有使用指针语义，根本无法实现。但这是可以的，因为共享值通常是安全的。除了 unmarshaling 之外，如果为一个引用类型使用指针语义，你应该三思。

### 用户定义类型（User Defined Types）

这是你最多需要作出决定的地方。你必须在你声明类型的时候决定使用什么语义。

如果我要求你给 time 包编写 API 接口，给你这种类型。

### 代码清单 6

```go
type Time struct {
	sec  int64
	nsec int32
	loc  *Location
}
```

你会使用什么语义？

在 Time 包中查看此类型的实现以及工厂函数 Now。

### 代码清单 7

```go
func Now() Time {
	sec, nsec := now()
	return Time{sec + unixToInternal, nsec, Local}
}
```

工厂函数对于类型来说是一种非常重要的函数，因为它告诉你（这种类型）所选择的语义。Now 函数就很清晰地（向我们）表明使用了值语义。该函数创建一个类型为 Time 的值并将该值的副本返回给调用者。 共享 Time 值不是必要的，（因为）他们的生命周期内不需要一直存在于堆上。

再看一下 Add 方法，它也是一个转变操作。

### 代码清单 8

```go
func (t Time) Add(d Duration) Time {
	t.sec += int64(d / 1e9)
	nsec := t.nsec + int32(d%1e9)
	if nsec >= 1e9 {
		t.sec++
		nsec -= 1e9
	} else if nsec < 0 {
		t.sec--
		nsec += 1e9
	}
	t.nsec = nsec
	return t
}
```

你可以再次看到 Add 方法遵循类型所选择的语义。Add 方法使用一个值接收器来操作它自己的 Time 值副本。其中，Time 值副本在调用中使用。它将修改自己的副本，并将 Time 值的新副本返回给调用者。

以下是一个接受 Time 值的函数：

### 代码清单 9

```go
func div(t Time, d Duration) (qmod2 int, r Duration) {
```

再一次，接受 Time 类型的值使用值语义。唯一使用指针语义的 Time API 接口，是这些 Unmarshal 相关的函数：

### 代码清单 10 

```go
func (t *Time) UnmarshalBinary(data []byte) error {
func (t *Time) GobDecode(data []byte) error {
func (t *Time) UnmarshalJSON(data []byte) error {
func (t *Time) UnmarshalText(data []byte) error {
```

大多数情况下，使用值语义的能力是有限的。将值从一个函数传递到另一个函数，（通常）使用值拷贝的方法是不正确或者不合理的。修改数据需要将其隔离成单个值再进行共享。这时，应该使用指针语义。如果你没办法 100% 确定拷贝值是正确并且合理的，那就使用指针语义吧。

查看 os 包中的 File 类型的生产函数。

### 代码清单 11

```go
func Open(name string) (file *File, err error) {
	return OpenFile(name, O_RDONLY, 0)
}
```

Open 函数返回一个 File 类型的指针。这意味着，对于 File 类型值，你应该使用指针语义来共享 File 的值。将指针语义修改为值语义，可能会对你的程序造成破坏性影响。当你与一个函数共享值时，最好假定你不允许拷贝值的指针并使用这个指针。否则，不知道将会出现什么样的异常情况。

查看更多的 API， 你将会看到更多使用指针语义的例子。

### 代码清单 12

```go
func (f *File) Chdir() error {
	if f == nil {
		return ErrInvalid
	}
	if e := syscall.Fchdir(f.fd); e != nil {
		return &PathError{"chdir", f.name, e}
	}
	return nil
}
```

虽然 File 值永远不会被修改，但是 Chdir 方法还是使用指针语义。该方法必须遵循该类型的语义约定。

### 代码清单 13

```go
func epipecheck(file *File, e error) {
	if e == syscall.EPIPE {
		if atomic.AddInt32(&file.nepipe, 1) >= 10 {
			sigpipe()
		}
	} else {
		atomic.StoreInt32(&file.nepipe, 0)
	}
}
```

这是一个名为 epipecheck 的函数，它使用指针来接收 File 值。再次注意一下，对于 File 值，一致使用指针语义。

## 结论

我在做代码 review 时，会寻找值或者指针语义是否使用一致。它可以帮助你保证代码的一致性和可预测性。它还使每个人能保持清晰和一致的心智模型。随着代码库和团队变得越来越大，值或者指针语义的一致性使用将会越来越重要。

Go 语言令人不解的地方在于指针和值语义之间的选择早已超出了接收器和函数参数的声明范围。接口的机制，函数值和切片都在语言的工作范围内。在将来的文章中，我将在这些语言的不同部分中展示值或者指针语义。

---

via: https://www.ardanlabs.com/blog/2017/06/design-philosophy-on-data-and-semantics.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[gogeof](https://github.com/gogeof)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出