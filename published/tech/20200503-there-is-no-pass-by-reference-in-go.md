首发于：https://studygolang.com/articles/28977

# Go 中没有引用变量

先说清楚，在 go 中没有引用变量，所以更不存在什么引用传值了。

## 什么是引用变量

在类 C++ 语言中，你可以声明一个别名，给一个变量安上一个其他名字，我们把这称为引用变量。

```c
#include <stdio.h>

int main() {
	int a = 10;
	int &b = a;
	int &c = b;

	printf("%p %p %p\n", &a, &b, &c); // 0x7ffe114f0b14 0x7ffe114f0b14 0x7ffe114f0b14
	return 0;
}
```

你可以看到 `a`,`b`,`c` 都指向同一块内存地址，三者的值相同，当你要在不同范围内声明引用变量（即函数调用）时，此功能很有用。

## Go 中不存在引用变量

与 C++ 不同的是，Go 中的每一个变量都有着独一无二的内存地址。

```go
package main

import "fmt"

func main() {
	var a, b, c int
	fmt.Println(&a, &b, &c) // 0x1040a124 0x1040a128 0x1040a12c
}
```

你不可能在 Go 程序中找到两个变量共享一块内存，但是可以让两个变量指向同一个内存。

```go
package main

import "fmt"

func main() {
	var a int
	var b, c = &a, &a
	fmt.Println(b, c)   // 0x1040a124 0x1040a124
	fmt.Println(&b, &c) // 0x1040c108 0x1040c110
}
```

在这个例子中，`b` 和 `c` 拥有 `a` 的地址，但是 `b` 和 `c` 这两个变量却被存储在不同的内存地址中，更改 `b` 的值并不会影响到 `c`。

## `map` 和 `channel` 是引用吗

不是，map 和 channel 都不是引用，如果他们是的话，下面这个例子就会输出 `false`

```go
package main

import "fmt"

func fn(m map[int]int) {
	m = make(map[int]int)
}

func main() {
	var m map[int]int
	fn(m)
	fmt.Println(m == nil)
}
```

如果是引用变量的话，`main` 中的 `m` 被传到 `fn` 中，那么经过函数的处理 `m` 应该已经被初始化了才对，但是可以看出 `fn` 的处理对 `m` 并没有影响，所以 `map` 也不是引用。

`map` 是一个指向 `runtime.hmap` 结构的指针，如果你还有疑问的话，请继续阅读下去。

## map 类型是什么

当我们这样声明的时候。

```go
m := make(map[int]int)
```

编译器将其替换为调用 map.go/[makemap](https://golang.org/src/runtime/map.go?h=makemap%28%29)

```go
// makemap implements Go map creation for make(map[k]v, hint).
// If the compiler has determined that the map or the first bucket
// can be created on the stack, h and/or bucket may be non-nil.
// If h != nil, the map can be created directly in h.
// If h.buckets != nil, bucket pointed to can be used as the first bucket.
func makemap(t *maptype, hint int, h *hmap)*hmap
```

可以看到，`makemap` 函数返回 `*hmap`，一个指向[hmap](https://golang.org/src/runtime/map.go?h=hmap#L115)结构的指针，我们可以从 go 源码中看到这些，除此之外，我们还可以证明 map 值的大小和 `uintptr` 一样。

```go
package main
import (
	"fmt"
	"unsafe"
)

func main() {
	var m map[int]int
	var p uintptr
	fmt.Println(unsafe.Sizeof(m), unsafe.Sizeof(p)) // 8 8 (linux/amd64)
}
```

## 如果 map 是指针的话，它不应该返回 *map[key]value 吗

这是个好问题，为什么表达式 `make(map[int]int)` 返回一个 map[int]int 类型的结构？不应该返回 `*map[int]int` 吗？

Ian Taylor [在这个回答](https://groups.google.com/forum/#!msg/golang-nuts/SjuhSYDITm4/jnrp7rRxDQAJ)中说：

> In the very early days what we call maps now were written as pointers, so you wrote *map[int]int. We moved away from that when we realized that no one ever wrote `map` without writing `*map`.

所以说，Go 把 `*map[int]int` 重命名为 `map[int]int`

---

via: https://dave.cheney.net/2017/04/29/there-is-no-pass-by-reference-in-go

作者：[Dave Cheney](https://dave.cheney.net/about)
译者：[Jun10ng](https://github.com/Jun10ng)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
