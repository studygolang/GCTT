首发于：https://studygolang.com/articles/28976

# 重新学习 slice By Dave Cheney

## 数组 Arrays

每次讨论到 Go 的切片问题，都会从这个变量是不是切片开始，换句话说，就是 Go 的序列类型，在 Go 中，数组有两种关联属性。

1. 数组拥有固定的大小；`[5]int` 即表明是一个有 5 个 `int` 的数组，又与 `[3]int` 相区分开。
2. 他们是值类型。思考下面这个例子。

```go
package main
import "fmt"
func main() {
	var a [5]int
	b := a
	b[2] = 7
	fmt.Println(a, b) // prints [0 0 0 0 0] [0 0 7 0 0]
}
```

声明语句 `b:=a` 声明一个新变量 `b`，一个 `[5]int` 的数据类型，并把 `a` 的内容拷贝到 `b` 中，更改 `b` 中的值并不会对 `a` 中内容造成影响，因为 `a` 和 `b` 是独立的。

*作者注：这并不是数组的特殊属性，在 Go 中，每次分配其实都是副本值传递*。

## 切片 slices

Go 的切片类型与数组类型有两个不同的地方：

1. 切片其实没有固定长度，一个切片的长度没有被声明为其类型的一部分，而是被保留在切片结构本身中并且可以通过内置函数 `len` 来重置他。
2. 用一个切片赋值给另一个切片并不会创建前一个切片的内容副本，因为切片类型没有直接拥有它的内容，而是拥有一个指针，而这个指针指向切片下方的数组，数组内的元素才是切片的内容。

*作者注：这有时也被成为后台数组（backing arrays)*。

由于第二个特性，两个数组可以同时分享一个后台数组，思考以下例子：

### 例 1：对切片再切片

```go
package main

import "fmt"

func main() {
	var a = []int{1,2,3,4,5}
	b := a[2:]
	b[0] = 0
	fmt.Println(a, b) // prints [1 2 0 4 5] [0 4 5]
}
```

*译者注：`a` 也是个切片，而不是数组，只要 `[]` 内没有数字，就是切片，在本例中，`a` 是数组{1,2,3,4,5}的一个切片*。

在这个例子中，`a` 和 `b` 共同分享同一个后台数组，尽管 `b` 开始的偏移量和和长度都不同于 `a`,所以底层数组的更改会用同时影响到 `a` 和 `b`。

### 例 2：传切片变量给函数

```go
package main

import "fmt"

func negate(s []int) {
	for i := range s {
		s[i] = -s[i]
	}
}

func main() {
	var a = []int{1, 2, 3, 4, 5}
	negate(a)
	fmt.Println(a) // prints [-1 -2 -3 -4 -5]
}
```

在例 2 中，`a` 被传值给 `negate` 函数作为形式参数 `s`,函数遍历 `s` 中的元素，将他们转为相反数，尽管 `negate` 函数没有返回任何值或者用任何方式去在 `main` 中访问 `a`，但是 `a` 中的内容还是被 `negate` 所修改了。大多程序员程序员对 Go 中的切片与数组有一个直观的了解，应为这样的概念在其他语言中也有，例如:

### Python 重写例 1

```python
Python 2.7.10 (default, Feb  7 2017, 00:08:15)
[GCC 4.2.1 Compatible Apple LLVM 8.0.0 (clang-800.0.34)] on darwin
Type "help", "copyright", "credits" or "license" for more information.
>>> a = [1,2,3,4,5]
>>> b = a
>>> b[2] = 0
>>> a
[1, 2, 0, 4, 5]

```

### Ruby

```ruby
irb(main):001:0> a = [1,2,3,4,5]
=> [1, 2, 3, 4, 5]
irb(main):002:0> b = a
=> [1, 2, 3, 4, 5]
irb(main):003:0> b[2] = 0
=> 0
irb(main):004:0> a
=> [1, 2, 0, 4, 5]
```

## slice Header

想要理解 slice 是如何做到本身是一个类，并且又是一个指针的话，就得理解  [slice 的底层结构](https://golang.org/pkg/reflect/#sliceHeader)。

`slicet Header` 看起来就像这样：

![Slice Header.png](https://raw.githubusercontent.com/studygolang/gctt-images/master/slice-from-the-ground-up/Slice%20Header.png)

```go
package runtime

type slicece struct {
	ptr   unsafe.Pointer
	len   int
	cap   int
}
```

它不像[`map` 和 `chan` 类型](https://dave.cheney.net/2017/04/30/if-a-map-isnt-a-reference-variable-what-is-it)，它们是引用类型，而切片是值类型，在赋值或者作为参数传递给函数的时候会复制。

为了说明这些，程序员可以直观的将 `square()` 的形参理解为 `main` 中 `v` 的副本。

```go
package main

import "fmt"

func square(v int) {
	v = v * v
}

func main() {
	v := 3
	square(v)
	fmt.Println(v) // prints 3, not 9
}
```

`square` 的操作并不会对原本的 `v` 有任何影响，形参可以理解为所传值的单独拷贝。

```go
package main

import "fmt"

func double(s []int) {
	s = append(s, s...)
}

func main() {
	s := []int{1, 2, 3}
	double(s)
	fmt.Println(s, len(s)) // prints [1 2 3] 3
}
```

Go 的 slice 变量稍微有点不一样的特性就是他是作为值传递的，不仅仅是一个指针，90% 的时间当我们在 Go 中声明一个结构体的时候，你都会传递一个结构体指针。slice 的值传递很不常见，我能想到的另一个值传递的结构体为 `time.time`。

*作者注：当结构体实现了某个接口的时候，那么传递指针的概率这接近 100%*。

正是这种异常的，将 slice 作为值传递，而不是指针传递，这引起了 Go 程序员的混乱思考，但是只要记住：当我们赋值，截取，传递或者返回一个切片的时候，你只是在创建一个 slice Header 结构体，这个结构体有着三个字段：指向后台数组的指针，当前长度 `len`，容量 `cap`。

## 总结

写一个用切片作为栈的例子。

```go
package main

import "fmt"

func f(s []string, level int) {
	if level > 5 {
		   return
	}
	s = append(s, fmt.Sprint(level))
	f(s, level+1)
	fmt.Println("level:", level, "slicece:", s)
}

func main() {
	f(nil, 0)
}
```

从 `main` 开始，我们传递一个空值给 `f` 作为 `level 0`,在 `f` 内部，我们添加当前的 `level` 给 `s`,一旦 `level` 大于 5，`f` 就会执行 return 语句 ，打印 `s` 的副本。

```bash
level: 5 slicece: [0 1 2 3 4 5]
level: 4 slicece: [0 1 2 3 4]
level: 3 slicece: [0 1 2 3]
level: 2 slicece: [0 1 2]
level: 1 slicece: [0 1]
level: 0 slicece: [0]
```

---

via: https://dave.cheney.net/2018/07/12/slices-from-the-ground-up

作者：[Dave Cheney](https://dave.cheney.net/about)
译者：[Jun10ng](https://github.com/Jun10ng)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
