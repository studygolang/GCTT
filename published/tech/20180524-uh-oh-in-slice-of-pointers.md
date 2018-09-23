首发于：https://studygolang.com/articles/13595

# Go 语言的指针切片

![Image courtesy — https://xkcd.com/138/](https://raw.githubusercontent.com/studygolang/gctt-images/master/uh-oh-is-in-go-slice-of-pointers/1.png)

Go 让操作 Slice 和其他基本数据结构成为一件很简单的事情。对于来自 C/C++ 令人畏惧的指针世界的人来说，在大部分情况下使用 Golang 是一件令人幸福的事情。对于 **JS/Python** 的使用者来说，Golang 除了语法之外，没有什么区别。

然而，**JS/Pyhon** 的使用者或是 Go 的初学者总是遇到使用指针的时候。下面的场景就是他们可能会遇到的。

## 场景

假设这样一个场景，你需要载入一个含有数据的字符串指针的切片, `[]*string{}`。

让我们看一段代码。

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	// 声明一个字符串指针的切片
	listOfNumberStrings := []*string{}

	// 预先声明一个变量，这个变量会在添加将数据添加到切片之前存储这个数据
	var numberString string

	// 从 0 到 9 的循环
	for i := 0; i < 10; i++ {

		// 在数字之前添加 `#`，构造一个字符串
		numberString = fmt.Sprintf("#%s", strconv.Itoa(i))
		// 将数字字符串添加到切片中
		listOfNumberStrings = append(listOfNumberStrings, &numberString)
	}

	for _, n := range listOfNumberStrings {
		fmt.Printf("%s\n", *n)
    }
}
// 原文章代码有 Bug ，译者做了修改。
```

上面的示例代码生成了从 0 到 9 的数字。我们使用 `strconv.Itoa` 函数将每一个数字都转换成对应的字符串表达。然后将 `#` 字符添加至字符串的头部，最后利用 `append` 函数添加目标切片中。

运行上面的代码片段，你得到的输出是

```
➜ sample go run main.go
#9
#9
#9
#9
#9
#9
#9
#9
#9
#9
```

> 这是什么情况?
>
> 为什么我只看到最后数字 `#9` 被输出??? 我非常确定我把其他的数字也加到了这个列表中!
>
> 让我在这个示例程序中添加调试代码。

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	// 声明一个字符串指针的切片
	listOfNumberStrings := []*string{}

	// 预先声明一个变量，这个变量会在添加将数据添加到切片之前存储这个数据
	var numberString string

	// 从 0 到 9 的循环
	for i := 0; i < 10; i++ {
		// 在数字之前添加 `#`，构造一个字符串
		numberString = fmt.Sprintf("#%s", strconv.Itoa(i))
                fmt.Printf("Adding number %s to the slice\n", numberString)
		// 将数字字符串添加到切片中
		listOfNumberStrings = append(listOfNumberStrings, &numberString)
	}

	for _, n := range listOfNumberStrings {
		fmt.Printf("%s\n", *n)
	}
}
```

调式代码的输出为

```
➜ sample go run main.go
Adding number #0 to the slice
Adding number #1 to the slice
Adding number #2 to the slice
Adding number #3 to the slice
Adding number #4 to the slice
Adding number #5 to the slice
Adding number #6 to the slice
Adding number #7 to the slice
Adding number #8 to the slice
Adding number #9 to the slice
```

> 我看到他们被添加到...
>
> 这种事情怎么发生到我头上了？
>
> $@#! 啊啊啊啊啊!!

朋友，放轻松，让我们看看到底发生了什么。

```go
var numberString string
```

numberString 在这里会被分配到堆，让我们假设，它的内存地址为 `0x3AF1D234`。

![numberString on the stack at address 0x3AF1D234](https://raw.githubusercontent.com/studygolang/gctt-images/master/uh-oh-is-in-go-slice-of-pointers/2.png)

```go
for i := 0; i < 10; i++ {
	numberString = fmt.Sprintf("#%s", strconv.Itoa(i))
	listOfNumberStrings = append(listOfNumberStrings, &numberString)
}
```

现在让我们从 0 循环至 9。

### 第一次迭代[i=0]

在这次迭代中，我们生成了字符串 `"#0"` 并把它存储到变量 `numberString`。

![numberString stored at 0x3AF1D234 with content "#0"](
https://raw.githubusercontent.com/studygolang/gctt-images/master/uh-oh-is-in-go-slice-of-pointers/3.png)

接下来，我们获取 `numberString` 变量的地址(`&numberString`), 该地址为 `0x3AF1D234`，然后把它添加到 `listOfNumberStrings` 的切片中。

`listOfNumberStrings` 现在应该像下图一样

![listOfNumberStrings slice with value 0x3AF1D234](
https://raw.githubusercontent.com/studygolang/gctt-images/master/uh-oh-is-in-go-slice-of-pointers/4.png)

### 第二次迭代[i=1]

我们重复以上步骤。

这一次，我们生成了字符串 `"#1"`,并把他存储到相同的变量 `numberString` 中。

![numberString stored at 0x3AF1D234 with content "#1"](https://raw.githubusercontent.com/studygolang/gctt-images/master/uh-oh-is-in-go-slice-of-pointers/5.png)

接下来，我们取 `numberString` 变量的地址(&numberString), 地址的值等于 `0x3AF1D234`, 然后将其添加到 `listOfNumberStrings` 的切片中。

`listOfNumberStrings` 现在看起来应该像这样:

![Two items in the slice BOTH with values 0x3AF1D234.](
https://raw.githubusercontent.com/studygolang/gctt-images/master/uh-oh-is-in-go-slice-of-pointers/6.png)

希望现在已经开始让你明白发生什么了。

这个切片目前有两个变量。但是这两个变量(下标为 1 和 下标为 2 ) 都存储了相同的值： `0x3AF1D234` (`numberString` 的内存地址)。

然而，请记住，在第二次迭代的最后，存储在 `numberString` 的字符串是 `"#1"`。

重复以上步骤直到迭代结束。

最后一次迭代的后，存储在 `numberString` 的字符串是 `"#9"`。

现在让我们看一下，当我们通过 `*` 操作符以解引用的方式， 尝试输出存储在切片中的每一个元素的时候，会发生什么？

```go
for _, n := range listOfNumberStrings {
    fmt.Printf("%s\n", *n)
}
```

因为切片中存储的每一个变量的值都是 `0x3AF1D234` (像我们上面的例子中展示的)，解引用该元素将返回存在该内存地址上的值。

从最后一个迭代，我们知道最后被存储的值是 `"#9"`, 因此输出才像下面那样。

```
➜  sample go run main.go
#9
#9
#9
#9
#9
#9
#9
#9
#9
#9
```

## 解决方案

有一个相当简单的方法来解决这个问题：修改变量 `numberString` 声明的位置。

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	listOfNumberStrings := []*string{}

	for i := 0; i < 10; i++ {
		var numberString string
		numberString = fmt.Sprintf("#%s", strconv.Itoa(i))
		listOfNumberStrings = append(listOfNumberStrings, &numberString)
	}

	for _, n := range listOfNumberStrings {
		fmt.Printf("%s\n", *n)
	}

	return
}
```

我们在 `for` 循环中声明这个变量。

这是怎么做到的？ 每一次循环迭代，我们都强制重新在栈上声明变量 `numberString` ，从而给他一个新的不同的内存地址。

> **译按**：这里并非在栈上分配，通过逃逸分析 `go build -gcflags "-m"`，可以知道 `&numberString` 逃逸到堆上了。原作者这样解释，是因为作者使用 C 语言的角度去看待这个问题。在 C 语言中，也会遇到类似的问题，但的确可以通过在栈上强制申明一个新的变量来解释上面的代码。在 Golang 中，则通过编译器的逃逸分析解释以上代码。

然后我们用生成的字符串更新变量，把它的地址添加到切片中。这样的话，切片中的每一个元素都存储着独一无二的内存地址。

上面的代码的输出将会是

```
➜  sample go run main.go
#0
#1
#2
#3
#4
#5
#6
#7
#8
#9
```

我希望这篇文章能够帮助到一些人。我起写这篇文章的念头是因为我与一名公司初级工程师的经历，他遇到了相似的场景，并且完全绕不出来。这让我想到了我掉进类似陷阱的情况，那时候我是一名前公司的 C 语言初级工程师。

> Ps. 如果你来自 C/C++ 中奇妙的指针世界......老实说，你已经遇到了这个错误（并从中学习）！ ;)

---

via: https://medium.com/@nitishmalhotra/uh-ohs-in-go-slice-of-pointers-c0a30669feee

作者：[Nitish Malhotra](https://medium.com/@nitishmalhotra)
译者：[magichan](https://github.com/magichan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
