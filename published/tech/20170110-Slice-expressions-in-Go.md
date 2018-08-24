首发于：https://studygolang.com/articles/14409

# Go 语言中的两种 slice 表达式

在此之前，已经有许多关于 Golang 中 slice 的介绍,比如

* [Go Slices: usage and internals](https://blog.golang.org/go-slices-usage-and-internals)
* [How to avoid Go gotchas](https://blog.golang.org/go-slices-usage-and-internals)

本文只是关注于 slice 的表示方式，它们可以创建两种类型的值:

* 截断的 string
* 指向 array 或者 slice 的指针

Go 语言对 slice 有两种表示方式：简略表达式与完整表达式。

## 简略表达式

Slice 的简略表达式是：

```go
Input[low:high]
```

其中，low 和 high 是 slice 的索引(index)，其数值必须是整数，它们指定了输入操作数(Input)的哪些元素可以被放置在结果的 slice 中。输入操作数可以是 string，array，或者是指向 array 或 slice 的指针。结果 slice 的长度就是 high-low。如下例所示：

```go
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s := numbers[2:4:6]
fmt.Println(s) // [2, 3]
fmt.Println(cap(s)) // 4
```

将 slice 表达式应用到 array 的指针中，这是第一次取消对该指针的引用，然后以常规的方式应用 slice 表达式。将 slice 表达式应用于数组的指针，是先解引用该指针，然后按常规方式应用切片表达式的简写形式。

```go
numbers := [5]int{1, 2, 3, 4, 5}
fmt.Println((&numbers)[1:3]) // [2, 3]
```

Slice 的索引 low 和 high 可以省略，low 的默认值是0，high 的默认值为 slice 的长度：

```go
fmt.Println("foo"[:2]) // "fo"
fmt.Println("foo"[1:]) // "oo"
fmt.Println("foo"[:]) // "foo"
```

但是 slice 的索引不可以是以下几种类型的值：

* low < 0 or high < 0
* low <= high
* high <= len(input)

```go
fmt.Println("foo"[-1:])  // invalid slice index -1 (index must be non-negative)
//fmt.Println("foo"[:4]) // invalid slice index 4 (out of bounds for 3-byte string)
fmt.Println("foo"[2:2])  // ""(blank)
//fmt.Println("foo"[2:1]) // invalid slice index: 2 > 1
```

否则，即使一个超过 slice 范围的索引不能在编译的时候被检测到，那么在运行时就会发生一个 panic。

```go
func low() int {
	return 4
}

func main() {
	fmt.Println("foo"[low():])
}
```

```
panic: runtime error: slice bounds out of range

goroutine 1 [running]:
panic(0x102280, 0x1040a018)
	/usr/local/go/src/runtime/panic.go:500 +0x720
main.main()
	/tmp/sandbox685025974/main.go:12 +0x120
```

## 完整表达式

这种方法可以控制结果 slice 的容量，但是只能用于 array 和指向 array 或 slice 的指针（ string 不支持），在简略表达式中结果 slice 的容量是从索引low开始的最大可能容量（ slice 的简略表达式）：

```go
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s := numbers[1:4]
fmt.Println(s) // [1, 2, 3]
fmt.Println(cap(s)) // 9
```

对于一个 array 来说，cap(a) == len(a) 在上面的代码片段中，s 的容量是 9，因为这个 slice 从索引 1 开始，而底层 array 有 8 个元素（2 到 9）。完整的 slice 表达式允许修改这种默认行为，以如下代码为例：

```go
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s := numbers[1:4:5]
fmt.Println(s) // [1, 2, 3]
fmt.Println(cap(s)) // 4
```

完整的 slice 表达式具有以下的形式：

```go
input[low:high:max]
```

索引 low 和索引 high 的含义和工作方式与简略表达式相同。唯一的区别是 max 将结果 slice 的容量设置为 max-low。

```go
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s := numbers[2:4:6]
fmt.Println(s) // [2, 3]
fmt.Println(cap(s)) // 4
```

当 slice 的输入操作数是一个 slice 时，结果 slice 的容量取决于输入操作数，而不是它的指向的底层 array：

```go
numbers := [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
fmt.Println(cap(numbers)) // 10
s1 := numbers[1:4]
fmt.Println(s1) // [1, 2, 3]
fmt.Println(cap(s1)) // 9
s2 := numbers[1:4:5]
fmt.Println(s2) // [1, 2, 3]
fmt.Println(cap(s2)) // 4
s3 := s2[:]
fmt.Println(s3) // [1, 2, 3]
fmt.Println(cap(s3)) // 4
```

这个例子中 s3 的容量不能超过 s2 的容量 (4)，即使它指向的 array 有 10 个元素，而且 s1,s2,s3 都是从 1 开始的。

当输入操作数是 slice 时，完整表达式的索引 high 不能超过其 cap(input)

```go
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s1 := numbers[0:1]
fmt.Println(s1) // [0]
fmt.Println(len(s1)) // 1
fmt.Println(cap(s1)) // 10
s2 := numbers[0:5]
fmt.Println(s2) // [0, 1, 2, 3, 4]
fmt.Println(cap(s1)) // 10
```

另外，对于它的 max 取值，有两个额外的规则

* high <= max
* max <= cap(input)

```go
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s1 := numbers[0:1]
s2 := numbers[0:5:11] // invalid slice index 11 (out of bounds for 10-element array)
fmt.Println(s1, s2)
```

在完整的 slice 表达式中低索引(low)是可以省略的

```go
numbers := [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
s := numbers[:4:6]
fmt.Println(s) // [0, 1, 2, 3]
fmt.Println(cap(s)) // 6
```

然而，不能省略略高索引(high)

```go
numbers := [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
s1 := numbers[:4:6]
s2 := s1[::5]
fmt.Println(s2)
fmt.Println(cap(s2))
```

否则，代码会在编译时报错"middle index required in 3-index slice"

---

via: https://medium.com/golangspec/slice-expressions-in-go-963368c20765

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[bizky]](https://github.com/bizky)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
