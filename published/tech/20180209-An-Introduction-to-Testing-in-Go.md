首发于：https://studygolang.com/articles/16772

# Go 测试介绍

测试在所有软件中都非常重要，它能够确保代码的正确性，并确保你所做的任何更改，最终都不会破坏代码库中其他不同部分的任何内容，这一点非常重要。

通过花费时间来充分测试你的程序，可以让你自己开发得更快，并且有信心使你正在开发的项目发布到生产环境时，还可以持续的工作。

## 视频教程

[YouTube https://youtu.be/GlA57dHa5Rg](https://youtu.be/GlA57dHa5Rg)

## 介绍

在本教程中，我们将关注如何开发并使用 `go test` 命令来测试你的 Go 代码。

### Go 测试文件

如果你之前看过任何 Go 项目，你可能已经注意到项目中的大多数文件，都在同一目录中具有 FILE_test.go 对应项。

这并不意外。这些文件包含了项目的所有单元测试，并且测试其对应的所有代码。

```
// An Example of how your project would be structured
myproject/
- calc.go
- calc_test.go
- main.go
- main_test.go
```

### 一个简单的测试文件

想象一下，我们有一个非常简单的 Go 程序，它由一个文件组成，并具有 calculate() 函数。这个 calculate() 函数只需要 1 个数字，然后让这个数字加上 2 。让我们优雅而简单开始运行它吧：

```go
package main

import (
	"fmt"
)

// Calculate returns x + 2.
func Calculate(x int) (result int) {
	result = x + 2
	return result
}

func main() {
	fmt.Println("Hello World")
}
```

如果我们想测试这个程序，我们可以在同一目录中创建一个 `main_test.go` 文件并编写以下测试代码：

```go
package main

import (
	"testing"
)

func TestCalculate(t *testing.T) {
	if Calculate(2) != 4 {
		t.Error("Expected 2 + 2 to equal 4")
	}
}
```

### 运行我们的测试

现在我们已经创建了第一个 Go 测试程序，现在开始运行它，看看我们的代码是否按照我们期望的方式运行。我们可以通过执行下面的命令来运行我们的测试 :

```
go test
```

然后应输出类似于以下的内容：

```
Elliots-MBP:go-testing-tutorial elliot$ Go test
PASS
ok      _/Users/elliot/Documents/Projects/tutorials/golang/go-testing-tutorial  0.007s
```

### 表驱动测试

现在很高兴的是单个计算程序可以工作了，我们应该在代码中添加一些其他的测试用例来提高信心。如果我们希望逐步构建一系列经常使用的测试用例，我们可以像下面这样在测试中使用 `array` ：

```go
func TestTableCalculate(t *testing.T) {
	var tests = []struct {
		input    int
		expected int
	}{
		{2, 4},
		{-1, 1},
		{0, 2},
		{-5, -3},
		{99999, 100001},
	}

	for _, test := range tests {
		if output := Calculate(test.input); output != test.expected {
			t.Error("Test Failed: {} inputted, {} expected, recieved: {}", test.input, test.expected, output)
		}
	}
}
```

这里我们声明一个结构包含输入和期望值。然后，我们使用 `for _，test：= range tests` 来迭代测试列表，调用我们的函数，并检查不论如何输入是否始终返回我们预期的结果。

当我们现在运行我们的整套测试时，我们应该看到与之前相同的输出：

```
Elliots-MBP:go-testing-tutorial elliot$ Go test
PASS
ok      _/Users/elliot/Documents/Projects/tutorials/golang/go-testing-tutorial  0.007s
```

## 详细测试输出

有时你可能希望知道具体运行了哪些测试以及运行花费的时间。值得庆幸的是，如果你像下面这样使用 `-v` 标志就可以看到。

```
Elliots-MBP:go-testing-tutorial elliot$ Go test -v
=== RUN   TestCalculate
--- PASS: TestCalculate (0.00s)
=== RUN   TestTableCalculate
--- PASS: TestTableCalculate (0.00s)
PASS
ok      _/Users/elliot/Documents/Projects/tutorials/golang/go-testing-tutorial  0.006s
```

你可以看到我们的正常测试和表测试都运行通过了，执行时间不到 `0.00` 秒。

## 总结

希望您发现本教程很有用！如果您需要进一步的帮助，请随时在下面的评论部分告诉我。

### 进一步阅读

如果你喜欢这篇文章，你可能会喜欢我的其他 Go 测试文章：

* [Advanced Testing in Go](https://tutorialedge.net/golang/advanced-go-testing-tutorial/)
* [Improving Your Tests with Testify in Go](https://tutorialedge.net/golang/improving-your-tests-with-testify-go/)

---

via: https://tutorialedge.net/golang/intro-testing-in-go/

作者：[Elliot Forbes](https://tutorialedge.net/about/)
译者：[Tyrodw](https://github.com/tyrodw)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
