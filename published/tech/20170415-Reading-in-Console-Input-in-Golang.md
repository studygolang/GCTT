首发于：https://studygolang.com/articles/16158

# Go 读取控制台输入

这是一个快速简单的教程，主要内容是如何在 Go 程序中读取控制台的输入。在这个教程中，我们将创建一个非常简单的脚本，这个脚本可以读取用户的输入并打印出来。

## 读取整个句子

我们使用 `while` 循环，在 Go 语言中相当于没有任何参数的 `for` 循环，这样就可以让程序一直运行了。在这个例子中，每次输入一个字符串并按下 `enter` 键，我们会通过 `\n` 这个关键字符来区分字符串，如果你想对比刚才输入的字符串，我们还需要调用 replace 方法来去除掉 `\n` 然后再进行比较。

> 如果你想让这个程序在 Windows 系统下运行，那么你必须将代码的 `text` 替换为 `text = strings.Replace(text,"\r\n","",-1)` 因为 Windowss 系统使用的行结束符和 unix 系统是不同的。

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Simple Shell")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		if strings.Compare("hi", text) == 0 {
			fmt.Println("hello, Yourself")
		}

	}

}
```

在这个例子中我们可以发现，无论什么时候我们输入 "hi"，我们的 `strings.Compare()` 方法都将会返回 0，并且向我们打印出 hello。

## 读取一个 UTF-8 编码的 Unicode 字符

如果你想从命令行中简单地读取一个 unicode 字符，我建议你使用 `bufio.ReadRune()`，就像这样：

```go
reader := bufio.NewReader(os.Stdin)
char, _, err := reader.ReadRune()

if err != nil {
	fmt.Println(err)
}

// 打印输出 unicode 值为：A -> 65, a -> 97
fmt.Println(char)

switch char {
case 'A':
	fmt.Println("A Key Pressed")
	break
case 'a':
	fmt.Println("a Key Pressed")
	break
}
```

## 使用 Bufio 的 Scanner 方法

第三种从控制台读取的权威方法是通过创建一个 `scanner` 对象 , 然后通过 `os.Stdin` 来完成，就像我们刚才创建一个 `readers` 做的，这次我们使用 `scanner.Scan` 去从控制台读取内容：

```go
func scanner() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
```

无论你什么时候输入内容，上面的代码都将不断地在控制台扫描输入并打印出来。

## 结论

正如你所看到的，这里有很多解决的方法，而哪一个方法最适合需要根据需求来决定。如果你只是需要读取一个字符，那就使用 `ReadRune()` 方法 , 又或则你想读取一个完整的字符串，那 `ReadString()` 方法将是一个比较好的方法。

我希望你这篇文章对你有所帮助，如果有任何疑问，可以在下面的评论区留言。

> 如果你喜欢这篇文章，那么你应该也喜欢这篇[Calling System Commands](https://tutorialedge.net/golang/executing-system-commands-with-golang)。

---

via: https://tutorialedge.net/golang/reading-console-input-golang/

作者：[Elliot Forbes](https://tutorialedge.net/about/)
译者：[xmge](https://github.com/xmge)
校对：[Alex-liutao](https://github.com/Alex-liutao)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
