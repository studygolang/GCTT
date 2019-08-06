首发于：https://studygolang.com/articles/22421

# Go 切片是胖指针

*本文的内容曾在 [Hacker News](https://news.ycombinator.com/item?id=20321116) 上进行讨论。*

使用 C 语言时，常见的一个难题就是要理解指针除了表示一个内存地址以外什么都不是。一个传入了指针的被调函数只知道这个指针指向什么类型的对象——也许包含了内存对齐和指针可以被如何使用之类的信息。如果这是一个指向 void 的指针（即 `void *`），那么就连这类信息也是无法获知的。

指针指向多少个连续的元素也是无法获知的。可能是零个，这样的话解引用就是非法的。即使当指针非空时这种情况也可能发生。指针可以超出数组的末尾，指向零个元素。比如：

```c
void foo(int *);

void bar(void)
{
	int array[4];
	foo(array + 4);  // 指针指向数组末尾后一个位置
}
```

在某些情况下，元素的个数可知，至少对程序员来说是这样。比如，函数可能会规定必须传入至少 N 个或正好 N 个元素。这种信息可以用文档来传递：

```c
/** Foo 接受 4 个整数。 */
void foo(int *);
```

或者通过函数原型来传达这样的信息。尽管下面的函数表面上接受一个数组，实际上却是一个指针，“ 4 ”和函数原型并没有关系。

```c
void foo(int[4]);
```

虽然 C99 引入了一个使这种写法成为原型正式部分的特性，但不幸的是我从没见过有哪个编译器真的会使用这一信息。

```c
void foo(int[static 4]);  // >= 4 个元素， 不能为空
```

另一种常见的模式是让被调函数接受一个计数器参数。比如，POSIX `write()` 函数：

```c
ssize_t write(int fd, const void *buf, size_t count);
```

描述缓冲区大小的必要信息被两个参数隔开了。这看起来冗长，而且如果这两个参数不一致的话还会导致严重的 bug （缓冲区溢出、[信息泄露](https://nullprogram.com/blog/2017/07/19/) 等）。如果这些信息能整合到指针当中，岂不会好一些？这就是*胖指针*的定义。

## 通过位运算实现胖指针

如果我们对目标平台做出一些假设，我们就可以通过一些指针的[“奇技淫巧”](https://nullprogram.com/blog/2016/05/30/)，利用指针当中未被使用的位，将胖指针整合到一个普通指针中。比如，在目前的 x86-64 平台上，一个指针中只有低 48 位被真正使用到。其余 16 位可以被谨慎地用于传递其他信息，比如元素个数或者字节数：

```c
// 注意：只能在 x86-64 平台上这么做！
unsigned char buf[1000];
uintptr addr = (uintptr_t)buf & 0xffffffffffff;
uintptr pack = (sizeof(buf) << 48) | addr;
void *fatptr = (void *)pack;
```

接收方可以解译出这些信息。显然，只有 16 位用于计数，这通常是不够的，所以这一方法更可能被用于[边界检查](https://www.usenix.org/legacy/event/sec09/tech/full_papers/akritidis.pdf)。

更进一步，如果我们知道内存对齐的情况——比如说 16 字节对齐——那么我们也可以在低位中编码信息，比如类型标签。

## 通过结构体实现胖指针

上面所提到的方法不安全、不可移植，而且相当局限。一个更健壮的方法是用更大的类型来包装指针，比如结构体。

```c
struct fatptr {
	void *ptr;
	size_t len;
};
```

以胖指针作为参数的函数不再需要计数器参数，而且通常胖指针是值传递。

```c
fatptr_write(int fd, struct fatptr);
```

在典型的 C 语言的实现中，结构体字段确实会被传递，如果不是这样的话，就相当于每个字段单独作为参数进行传递，所以效率也没有低多少。

为了更直接一些，我们可以使用宏：

```c
#define COUNTOF(array) \
	(sizeof(array) / sizeof(array[0]))

#define FATPTR(ptr, count) \
	(struct fatptr){ptr, count}

#define ARRAYPTR(array) \
	FATPTR(array, COUNTOF(array))

/* ... */

unsigned char buf[40];
fatptr_write(fd, ARRAYPTR(buf));
```

这种方法存在明显的缺陷，比如 void 指针带来的类型混淆、不能使用 `const`，而且这种写法对 C 而言很怪。在一个真实的程序中我不会这么写，但现在请暂时忍耐。

在我往下说之前，我想往胖指针结构体中添加一个字段：容量。

```c
struct fatptr {
	void *ptr;
	size_t len;
	size_t cap;
};
```

这样一来，传递的信息就不仅包括目前有多少个元素（`len`），而且包括缓冲区当中还剩下多少额外的空间。比如，这让被调函数知道有多少剩余空间可用于追加新元素。

```c
// 往缓冲区剩下的空间中填充值。
void
fill(struct fatptr ptr, int value)
{
	int *buf = ptr.ptr;
	for (size_t i = ptr.len; i < ptr.cap; i++) {
		buf[i] = value;
	}
}
```

既然被调函数修改了胖指针，就应该返回胖指针：

```c
struct fatptr
fill(struct fatptr ptr, int value)
{
	int *buf = ptr.ptr;
	for (size_t i = ptr.len; i < ptr.cap; i++) {
		buf[i] = value;
	}
	ptr.len = ptr.cap;
	return ptr;
}
```

恭喜，现在你有了切片！ Go 语言与其的差别在于，切片是 Go 语言本身的一部分，所以无需依赖于危险的技巧或者冗长的额外信息。上面的 `fatptr_write()` 函数几乎和 Go 中接受一个切片的 `Writer.Write()` 函数有相同的功能：

```go
type Writer interface {
	Write(p []byte) (n int, err error)
}
```

## Go 切片

Go 广为人知的一个特性是拥有指针，包括*内部*指针，但是不支持指针运算。你（几乎）可以获取任何东西的地址，但你不能让这个指针指向别的东西，即使你获取的是一个数组元素的地址。指针运算会危害 Go 的类型安全，所以它只能通过 `unsafe` 包中提供的一些特殊机制实现。

但是指针运算确实有用！获取一个数组元素的地址，传给一个函数，然后允许函数修改数组的一个切片，这样的操作会很方便。切片就是支持这类指针运算的指针，但很安全。不同于 `&` 操作符会创建一个简单的指针，切片操作符会派生出一个胖指针。

```go
func fill([]int, int) []int

var array [8]int

// len == 0, cap == 8, 相当于 &array[0]
fill(array[:0], 1)
// array 现在变成 [1, 1, 1, 1, 1, 1, 1, 1]

// len == 0, cap == 4, 相当于 &array[4]
fill(array[4:4], 2)
// array 现在是 [1, 1, 1, 1, 2, 2, 2, 2]
```

`fill` 函数可以接受一个切片的切片，高效地通过指针运算移动指针而不会破坏内存安全，因为有额外的“胖指针”信息。换句话说，胖指针可由别的胖指针派生得到。

至少就目前而言，切片并不像胖指针那么常见。你可以用 `&` 获取任何变量的地址，但是你不能获取任意变量的*切片*，即使在逻辑上这行得通。

```go
var foo int

// 试图创建一个在底层指向 foo，len = 1，cap = 1 的切片
var fooslice []int = foo[:] // 编译错误！
```

总而言之这么做没多大用处。然而，如果你非要这么做，那么 `unsafe` 包可以实现。我相信得到的切片可以放心地使用：

```go
// 先转换成只有一个元素地数组，再转换成切片
fooslice = (*[1]int)(unsafe.Pointer(&foo))[:]
```

更新：[Chris Siebenmann 关于为什么这需要 `unsafe` 包的推测](https://utcc.utoronto.ca/~cks/space/blog/programming/GoVariableToArrayConversion)。

当然，切片十分灵活，在许多使用场景中看起来不那么像胖指针，但当我写 Go 时我仍然会用这种方式来看待切片。

---

via: https://nullprogram.com/blog/2019/06/30/

作者：[Chris Wellons](https://github.com/skeeto)
译者：[maxwellhertz](https://github.com/maxwellhertz)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
