已发布：https://studygolang.com/articles/12784

# 第 31 篇：自定义错误

![custom errors](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-series/custom-errors-golang-1.png)

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 31 篇。

在[上一教程](https://studygolang.com/articles/12724)里，我们学习了 Go 中的错误是如何表示的，并学习了如何处理标准库里的错误。我们还学习了从标准库的错误中提取更多的信息。

在本教程中，我们会学习如何创建我们自己的自定义错误，并在我们创建的函数和包中使用它。我们会使用与标准库中相同的技术，来提供自定义错误的更多细节信息。

## 使用 New 函数创建自定义错误

创建自定义错误最简单的方法是使用 [`errors`](https://golang.org/pkg/errors/) 包中的 [`New`](https://golang.org/pkg/errors/#New) 函数。

在使用 New [函数](https://studygolang.com/articles/11892) 创建自定义错误之前，我们先来看看 `New` 是如何实现的。如下所示，是 [`errors` 包](https://golang.org/src/errors/errors.go?s=293:320#L1) 中的 `New` 函数的实现。

```go
// Package errors implements functions to manipulate errors.
package errors

// New returns an error that formats as the given text.
func New(text string) error {
	return &errorString{text}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
```

`New` 函数的实现很简单。`errorString` 是一个[结构体](https://studygolang.com/articles/12263)类型，只有一个字符串字段 `s`。第 14 行使用了 `errorString` 指针接受者（Pointer Receiver），来实现 `error` 接口的 `Error() string` [方法](https://studygolang.com/articles/12264)。

第 5 行的 `New` 函数有一个字符串参数，通过这个参数创建了 `errorString` 类型的变量，并返回了它的地址。于是它就创建并返回了一个新的错误。

现在我们已经知道了 `New` 函数是如何工作的，我们开始在程序里使用 `New` 来创建自定义错误吧。

我们将创建一个计算圆半径的简单程序，如果半径为负，它会返回一个错误。

```go
package main

import (
	"errors"
	"fmt"
	"math"
)

func circleArea(radius float64) (float64, error) {
	if radius < 0 {
		return 0, errors.New("Area calculation failed, radius is less than zero")
	}
	return math.Pi * radius * radius, nil
}

func main() {
	radius := -20.0
	area, err := circleArea(radius)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Area of circle %0.2f", area)
}
```

[在 glayground 上运行](https://play.golang.org/p/_vuf6fgkqm)

在上面的程序中，我们检查半径是否小于零（第 10 行）。如果半径小于零，我们会返回等于 0 的面积，以及相应的错误信息。如果半径大于零，则会计算出面积，并返回值为 `nil` 的错误（第 13 行）。

在 `main` 函数里，我们在第 19 行检查错误是否等于 `nil`。如果不是 `nil`，我们会打印出错误并返回，否则我们会打印出圆的面积。

在我们的程序中，半径小于零，因此打印出：

```
Area calculation failed, radius is less than zero
```

## 使用 Errorf 给错误添加更多信息

上面的程序效果不错，但是如果我们能够打印出当前圆的半径，那就更好了。这就要用到 [`fmt`](https://golang.org/pkg/fmt/) 包中的 [`Errorf`](https://golang.org/pkg/fmt/#Errorf) 函数了。`Errorf` 函数会根据格式说明符，规定错误的格式，并返回一个符合该错误的[字符串](https://studygolang.com/articles/12261)。

接下来我们使用 `Errorf` 函数来改进我们的程序。

```go
package main

import (
	"fmt"
	"math"
)

func circleArea(radius float64) (float64, error) {
	if radius < 0 {
		return 0, fmt.Errorf("Area calculation failed, radius %0.2f is less than zero", radius)
	}
	return math.Pi * radius * radius, nil
}

func main() {
	radius := -20.0
	area, err := circleArea(radius)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Area of circle %0.2f", area)
}
```

[在 playground 上运行](https://play.golang.org/p/HQ7bvjT4o2)

在上面的程序中，我们使用 `Errorf`（第 10 行）打印了发生错误的半径。程序运行后会输出：

```
Area calculation failed, radius -20.00 is less than zero
```

## 使用结构体类型和字段提供错误的更多信息

错误还可以用实现了 `error` [接口](https://studygolang.com/articles/12266)的结构体来表示。这种方式可以更加灵活地处理错误。在上面例子中，如果我们希望访问引发错误的半径，现在唯一的方法就是解析错误的描述信息 `Area calculation failed, radius -20.00 is less than zero`。这样做不太好，因为一旦描述信息发生变化，程序就会出错。

我们会使用标准库里采用的方法，在上一教程中“断言底层结构体类型，使用结构体字段获取更多信息”这一节，我们讲解了这一方法，可以使用结构体字段来访问引发错误的半径。我们会创建一个实现 `error` 接口的结构体类型，并使用它的字段来提供关于错误的更多信息。

第一步就是创建一个表示错误的结构体类型。错误类型的命名约定是名称以 `Error` 结尾。因此我们不妨把结构体类型命名为 `areaError`。

```go
type areaError struct {
	err    string
	radius float64
}
```

上面的结构体类型有一个 `radius` 字段，它存储了与错误有关的半径，而 `err` 字段存储了实际的错误信息。

下一步是实现 `error` 接口。

```go
func (e *areaError) Error() string {
	return fmt.Sprintf("radius %0.2f: %s", e.radius, e.err)
}
```

在上面的代码中，我们使用指针接收者 `*areaError`，实现了 `error` 接口的 `Error() string` 方法。该方法打印出半径和关于错误的描述。

现在我们来编写 `main` 函数和 `circleArea` 函数来完成整个程序。

```go
package main

import (
	"fmt"
	"math"
)

type areaError struct {
	err    string
	radius float64
}

func (e *areaError) Error() string {
	return fmt.Sprintf("radius %0.2f: %s", e.radius, e.err)
}

func circleArea(radius float64) (float64, error) {
	if radius < 0 {
		return 0, &areaError{"radius is negative", radius}
	}
	return math.Pi * radius * radius, nil
}

func main() {
	radius := -20.0
	area, err := circleArea(radius)
	if err != nil {
		if err, ok := err.(*areaError); ok {
			fmt.Printf("Radius %0.2f is less than zero", err.radius)
			return
		}
		fmt.Println(err)
		return
	}
	fmt.Printf("Area of rectangle1 %0.2f", area)
}
```

[在 playground 上运行](https://play.golang.org/p/OTs7J0adQg)

在上面的程序中，`circleArea`（第 17 行）用于计算圆的面积。该函数首先检查半径是否小于零，如果小于零，它会通过错误半径和对应错误信息，创建一个 `areaError` 类型的值，然后返回 `areaError` 值的地址，与此同时 `area` 等于 0（第 19 行）。**于是我们提供了更多的错误信息（即导致错误的半径），我们使用了自定义错误的结构体字段来定义它**。

如果半径是非负数，该函数会在第 21 行计算并返回面积，同时错误值为 `nil`。

在 `main` 函数的 26 行，我们试图计算半径为 -20 的圆的面积。由于半径小于零，因此会导致一个错误。

我们在第 27 行检查了错误是否为 `nil`，并在下一行断言了 `*areaError` 类型。**如果错误是 `*areaError` 类型，我们就可以用 `err.radius` 来获取错误的半径（第 29 行），打印出自定义错误的消息，最后程序返回退出**。

如果断言错误，我们就在第 32 行打印该错误，并返回。如果没有发生错误，在第 35 行会打印出面积。

该程序会输出：

```
Radius -20.00 is less than zero
```

下面我们来使用上一教程提到的[第二种方法](https://studygolang.com/articles/12724)，使用自定义错误类型的方法来提供错误的更多信息。

## 使用结构体类型的方法来提供错误的更多信息

在本节里，我们会编写一个计算矩形面积的程序。如果长或宽小于零，程序就会打印出错误。

第一步就是创建一个表示错误的结构体。

```go
type areaError struct {
	err    string //error description
	length float64 //length which caused the error
	width  float64 //width which caused the error
}
```

上面的结构体类型除了有一个错误描述字段，还有可能引发错误的宽和高。

现在我们有了错误类型，我们来实现 `error` 接口，并给该错误类型添加两个方法，使它提供了更多的错误信息。

```go
func (e *areaError) Error() string {
	return e.err
}

func (e *areaError) lengthNegative() bool {
	return e.length < 0
}

func (e *areaError) widthNegative() bool {
	return e.width < 0
}
```

在上面的代码片段中，我们从 `Error() string` 方法中返回了关于错误的描述。当 `length` 小于零时，`lengthNegative() bool` 方法返回 `true`，而当 `width` 小于零时，`widthNegative() bool` 方法返回 `true`。**这两个方法都提供了关于错误的更多信息，在这里，它提示我们计算面积失败的原因（长度为负数或者宽度为负数）。于是我们就有了两个错误类型结构体的方法，来提供更多的错误信息**。

下一步就是编写计算面积的函数。

```go
func rectArea(length, width float64) (float64, error) {
	err := ""
	if length < 0 {
		err += "length is less than zero"
	}
	if width < 0 {
		if err == "" {
			err = "width is less than zero"
		} else {
			err += ", width is less than zero"
		}
	}
	if err != "" {
		return 0, &areaError{err, length, width}
	}
	return length * width, nil
}
```

上面的 `rectArea` 函数检查了长或宽是否小于零，如果小于零，`rectArea` 会返回一个错误信息，否则 `rectArea` 会返回矩形的面积和一个值为 `nil` 的错误。

让我们创建 `main` 函数来完成整个程序。

```go
func main() {
	length, width := -5.0, -9.0
	area, err := rectArea(length, width)
	if err != nil {
		if err, ok := err.(*areaError); ok {
			if err.lengthNegative() {
				fmt.Printf("error: length %0.2f is less than zero\n", err.length)

			}
			if err.widthNegative() {
				fmt.Printf("error: width %0.2f is less than zero\n", err.width)

			}
			return
		}
		fmt.Println(err)
		return
	}
	fmt.Println("area of rect", area)
}
```

在 `main` 程序中，我们检查了错误是否为 `nil`（第 4 行）。如果错误值不是 `nil`，我们会在下一行断言 `*areaError` 类型。然后，我们使用 `lengthNegative()` 和 `widthNegative()` 方法，检查错误的原因是长度小于零还是宽度小于零。这样我们就使用了错误结构体类型的方法，来提供更多的错误信息。

如果没有错误发生，就会打印矩形的面积。

下面是整个程序的代码供你参考。

```go
package main

import "fmt"

type areaError struct {
	err    string  //error description
	length float64 //length which caused the error
	width  float64 //width which caused the error
}

func (e *areaError) Error() string {
	return e.err
}

func (e *areaError) lengthNegative() bool {
	return e.length < 0
}

func (e *areaError) widthNegative() bool {
	return e.width < 0
}

func rectArea(length, width float64) (float64, error) {
	err := ""
	if length < 0 {
		err += "length is less than zero"
	}
	if width < 0 {
		if err == "" {
			err = "width is less than zero"
		} else {
			err += ", width is less than zero"
		}
	}
	if err != "" {
		return 0, &areaError{err, length, width}
	}
	return length * width, nil
}

func main() {
	length, width := -5.0, -9.0
	area, err := rectArea(length, width)
	if err != nil {
		if err, ok := err.(*areaError); ok {
			if err.lengthNegative() {
				fmt.Printf("error: length %0.2f is less than zero\n", err.length)

			}
			if err.widthNegative() {
				fmt.Printf("error: width %0.2f is less than zero\n", err.width)

			}
			return
		}
		fmt.Println(err)
		return
	}
	fmt.Println("area of rect", area)
}
```

[在 playground 上运行](https://play.golang.org/p/iJv2V8pZ7c)

该程序会打印输出：

```
error: length -5.00 is less than zero
error: width -9.00 is less than zero
```

在上一教程[错误处理](https://studygolang.com/articles/12724)中，我们介绍了三种提供更多错误信息的方法，现在我们已经看了其中两个示例。

第三种方法使用的是直接比较，比较简单。我留给读者作为练习，你们可以试着使用这种方法来给出自定义错误的更多信息。

本教程到此结束。

简单概括一下本教程讨论的内容：

- 使用 `New` 函数创建自定义错误
- 使用 `Error` 添加更多错误信息
- 使用结构体类型和字段，提供更多错误信息
- 使用结构体类型和方法，提供更多错误信息

祝你愉快。

**上一教程 - [错误处理](https://studygolang.com/articles/12724)**

**下一教程 - panic 和 recover（暂未发布，敬请期待）**

---

via: https://golangbot.com/custom-errors/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
