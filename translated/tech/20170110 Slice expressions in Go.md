# Go语言中两种slice表达式

在此之前，已经有许多关于Golang中slice的介绍,比如
* [Go Slices: usage and internals](https://blog.golang.org/go-slices-usage-and-internals)
* [How to avoid Go gotchas](https://blog.golang.org/go-slices-usage-and-internals)
本文只是关注于slice的表示方式，它们可以创建两种类型的值:
* 截断的string
* 指向array或者slice的指针
Go语对slice有两种表示方式：简略表达式与完整表达式。

## 简略表达式
slice的简略表达式是：
```
Input[low:high]
```
其中，low和high是slice的索引(index)，其数值必须是整数，它们指定了输入操作数(Input)的哪些元素可以被放置在结果的slice中。
输入操作数可以是string，array，或者是指向array或slice的指针。
结果slice的长度就是high-low。
如下例所示：
```
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s := numbers[2:4:6]
fmt.Println(s) // [2, 3]
fmt.Println(cap(s)) // 4
```

将slice表达式应用到array的指针中，这是第一次取消对该指针的引用，然后以常规的方式应用slice表达式(这句话不知道怎么翻译：Applying slice expression to array’s pointer is as shorthand for first dereferencing such pointer and then applying slice expression in a regular manner)
```
numbers := [5]int{1, 2, 3, 4, 5}
fmt.Println((&numbers)[1:3]) // [2, 3]
```

slice的索引low和high可以省略，low的默认值是0，high的默认值为slice的长度：
```
fmt.Println("foo"[:2]) // "fo"
fmt.Println("foo"[1:]) // "oo"
fmt.Println("foo"[:]) // "foo"
```

但是slice的索引不可以是一下几种类型的值：
* low < 0 or high < 0
* low <= high
* high <= len(input)
```
/fmt.Println("foo"[-1:]) // invalid slice index -1 (index must be non-negative)
//fmt.Println("foo"[:4]) // invalid slice index 4 (out of bounds for 3-byte string)
fmt.Println("foo"[2:2]) // ""(blank)
//fmt.Println("foo"[2:1]) // invalid slice index: 2 > 1
```
否则，即使一个超过slice范围的索引不能在编译的时候被检测到，那么在运行时就会发生一个panic。
```
func low() int {
 return 4
}
func main() {
    fmt.Println("foo"[low():])
}
panic: runtime error: slice bounds out of range

goroutine 1 [running]:
panic(0x102280, 0x1040a018)
	/usr/local/go/src/runtime/panic.go:500 +0x720
main.main()
	/tmp/sandbox685025974/main.go:12 +0x120
```

## 完整表达式
这种方法可以控制结果slice的容量，但是只能用于array和指向array或slice的指针（string不支持），
在简略表达式中结果slice的容量是从索引low开始的最大可能容量（slice的简略表达式）：
```
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s := numbers[1:4]
fmt.Println(s) // [1, 2, 3]
fmt.Println(cap(s)) // 9
```
对于一个array来说，cap(a) == len(a)
在上面的代码片段中，s的容量是9，因为这个slice从索引1开始，而底层array有8个元素（2到9）。
完整的slice表达式允许修改这种默认行为，以如下代码为例：
```
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s := numbers[1:4:5]
fmt.Println(s) // [1, 2, 3]
fmt.Println(cap(s)) // 4
```

完整的slice表达式具有以下的形式：
```
input[low:high:max]
```
索引low和索引high的含义和工作方式与简略表达式相同。
唯一的区别是max将结果slice的容量设置为max-low。
```
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s := numbers[2:4:6]
fmt.Println(s) // [2, 3]
fmt.Println(cap(s)) // 4
```

当slice的输入操作数是一个slice时，结果slice的容量取决于输入操作数，而不是它的指向的底层array：
```
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
这个例子中s3的容量不能超过s2的容量(4)，即使它指向的array有10个元素，而且s1,s2,s3都是从1开始的。

当输入操作数是slice时，完整表达式的索引high不能超过其cap(input)
```
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s1 := numbers[0:1]
fmt.Println(s1) // [0]
fmt.Println(len(s1)) // 1
fmt.Println(cap(s1)) // 10
s2 := numbers[0:5]
fmt.Println(s2) // [0, 1, 2, 3, 4]
fmt.Println(cap(s1)) // 10
```
另外，对于它的max取值，有两个额外的规则
* high <= max
* max <= cap(input)
```
numbers := [10]int{0,1,2,3,4,5,6,7,8,9}
s1 := numbers[0:1]
s2 := numbers[0:5:11] // invalid slice index 11 (out of bounds for 10-element array)
fmt.Println(s1, s2)
```

在完整的slice表达式中低索引(low)是可以省略的
```
numbers := [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
s := numbers[:4:6]
fmt.Println(s) // [0, 1, 2, 3]
fmt.Println(cap(s)) // 6
```
然而，不能省略略高索引(high)
```
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
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出