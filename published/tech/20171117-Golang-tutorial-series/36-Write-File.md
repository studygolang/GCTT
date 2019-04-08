首发于：https://studygolang.com/articles/19475

# 第 36 章 写入文件

![write files](https://raw.githubusercontent.com/studygolang/gctt-images/master/golang-series/golang-write-files.png)

欢迎来到 [Golang 系列教程](https://studygolang.com/subject/2)的第 36 篇。

在这一章我们将学习如何使用 Go 语言将数据写到文件里面。并且还要学习如何同步的写到文件里面。

这章教程包括如下几个部分：

- 将字符串写入文件
- 将字节写入文件
- 将数据一行一行的写入文件
- 追加到文件里
- 并发写文件

请在本地运行所有本教程的程序，因为 playground 对文件的操作支持的并不好。

## 将字符串写入文件

最常见的写文件就是将字符串写入文件。这个写起来非常的简单。这个包含以下几个阶段。

1. 创建文件
2. 将字符串写入文件

我们将得到如下代码。

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.Create("test.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	l, err := f.WriteString("Hello World")
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}
```

在第 9 行使用 `create` 创建一个名字为 `test.txt` 的文件。如果这个文件已经存在，那么 `create` 函数将截断这个文件。该函数返回一个[文件描述符](https://docs.studygolang.com/pkg/os/#File)。

在第 14 行，我们使用 `WriteString` 将字符串 **Hello World** 写入到文件里面。这个方法将返回相应写入的字节数，如果有错误则返回错误。

最后，在第 21 行我们将文件关闭。

上面程序将打印：

```
11 bytes written successfully
```

运行完成之后你会在程序运行的目录下发现创建了一个 **test.txt** 的文件。如果你使用文本编辑器打开这个文件，你可以看到文件里面有一个 **Hello World** 的字符串。

## 将字节写入文件

将字节写入文件和写入字符串非常的类似。我们将使用 [Write](https://docs.studygolang.com/pkg/os/#File.Write) 方法将字节写入到文件。下面的程序将一个字节的切片写入文件。

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.Create("/home/naveen/bytes")
	if err != nil {
		fmt.Println(err)
		return
	}
	d2 := []byte{104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100}
	n2, err := f.Write(d2)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	fmt.Println(n2, "bytes written successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}
```

在上面的程序中，第 15 行使用了 **Write** 方法将字节切片写入到 `bytes` 这个文件里。这个文本在目录 `/home/naveen` 里面。你也可以将这个目录换成其他的目录。剩余的程序自带解释。如果执行成功，这个程序将打印 `11 bytes written successfully`。并且创建一个 `bytes` 的文件。打开文件，你会发现该文件包含了文本 **hello bytes**。

## 将字符串一行一行的写入文件

另外一个常用的操作就是将字符串一行一行的写入到文件。这一部分我们将写一个程序，该程序创建并写入如下内容到文件里。

```
Welcome to the world of Go.
Go is a compiled language.
It is easy to learn Go.
```

让我们看下面的代码：

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.Create("lines")
	if err != nil {
		fmt.Println(err)
				f.Close()
		return
	}
	d := []string{"Welcome to the world of Go1.", "Go is a compiled language.",
"It is easy to learn Go."}

	for _, v := range d {
		fmt.Fprintln(f, v)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("file written successfully")
}
```

在上面程序的第 9 行，我们先创建一个名字叫做 **lines** 的文件。在第 17 行，我们用迭代并使用 `for rang` 循环这个数组，并使用 [Fprintln](https://docs.studygolang.com/pkg/fmt/#Fprintln) **Fprintln** 函数 将 `io.writer` 做为参数，并且添加一个新的行，这个正是我们想要的。如果执行成功将打印 `file written successfully`，并且在当前目录将创建一个 `lines` 的文件。`lines` 这个文件的内容如下所示：

```
Welcome to the world of Go1.
Go is a compiled language.
It is easy to learn Go.
```

## 追加到文件

这一部分我们将追加一行到上节创建的 `lines` 文件中。我们将追加 **File handling is easy** 到 `lines` 这个文件。

这个文件将以追加和写的方式打开。这些标志将通过 [Open](https://docs.studygolang.com/pkg/os/#OpenFile) 方法实现。当文件以追加的方式打开，我们添加新的行到文件里。

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.OpenFile("lines", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	newLine := "File handling is easy."
	_, err = fmt.Fprintln(f, newLine)
	if err != nil {
		fmt.Println(err)
				f.Close()
		return
	}
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("file appended successfully")
}
```

在上面程序的第 9 行，我们以写的方式打开文件并将一行添加到文件里。当成功打开文件之后，在程序第 15 行，我们添加一行到文件里。程序成功将打印 `file appended successfully`。运行程序，新的行就加到文件里面去了。

```
Welcome to the world of Go1.
Go is a compiled language.
It is easy to learn Go.
File handling is easy.
```

## 并发写文件

当多个 goroutines 同时（并发）写文件时，我们会遇到[竞争条件(race condition)](https://golangbot.com/mutex/#criticalsection)。因此，当发生同步写的时候需要一个 channel 作为一致写入的条件。

我们将写一个程序，该程序创建 100 个 goroutinues。每个 goroutinue 将并发产生一个随机数，届时将有 100 个随机数产生。这些随机数将被写入到文件里面。我们将用下面的方法解决这个问题 .

1. 创建一个 channel 用来读和写这个随机数。
2. 创建 100 个生产者 goroutine。每个 goroutine 将产生随机数并将随机数写入到 channel 里。
3. 创建一个消费者 goroutine 用来从 channel 读取随机数并将它写入文件。这样的话我们就只有一个 goroutinue 向文件中写数据，从而避免竞争条件。
4. 一旦完成则关闭文件。

我们开始写产生随机数的 `produce` 函数：

```go
func produce(data chan int, wg *sync.WaitGroup) {
	n := rand.Intn(999)
	data <- n
	wg.Done()
}
```

上面的方法产生随机数并且将 `data` 写入到 channel 中，之后通过调用 `waitGroup` 的 `Done` 方法来通知任务已经完成。

让我们看看将数据写到文件的函数：

```go
func consume(data chan int, done chan bool) {
	f, err := os.Create("concurrent")
	if err != nil {
		fmt.Println(err)
		return
	}
	for d := range data {
		_, err = fmt.Fprintln(f, d)
		if err != nil {
			fmt.Println(err)
			f.Close()
			done <- false
			return
		}
	}
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		done <- false
		return
	}
	done <- true
}
```

这个 `consume` 的函数创建了一个名为 `concurrent` 的文件。然后从 channel 中读取随机数并且写到文件中。一旦读取完成并且将随机数写入文件后，通过往 `done` 这个 cahnnel 中写入 `true` 来通知任务已完成。

下面我们写 `main` 函数，并完成这个程序。下面是我提供的完整程序：

```go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
)

func produce(data chan int, wg *sync.WaitGroup) {
	n := rand.Intn(999)
	data <- n
	wg.Done()
}

func consume(data chan int, done chan bool) {
	f, err := os.Create("concurrent")
	if err != nil {
		fmt.Println(err)
		return
	}
	for d := range data {
		_, err = fmt.Fprintln(f, d)
		if err != nil {
			fmt.Println(err)
			f.Close()
			done <- false
			return
		}
	}
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		done <- false
		return
	}
	done <- true
}

func main() {
	data := make(chan int)
	done := make(chan bool)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go produce(data, &wg)
	}
	go consume(data, done)
	go func() {
		wg.Wait()
		close(data)
	}()
	d := <-done
	if d == true {
		fmt.Println("File written successfully")
	} else {
		fmt.Println("File writing failed")
	}
}
```

`main` 函数在第 41 行创建写入和读取数据的 channel，在第 42 行创建 `done` 这个 channel，此 channel 用于消费者 goroutinue 完成任务之后通知 `main` 函数。第 43 行创建 Waitgroup 的实例 `wg`，用于等待所有生产随机数的 goroutine 完成任务。

在第 44 行使用 `for` 循环创建 100 个 goroutines。在第 49 行调用 waitgroup 的 `wait()` 方法等待所有的 goroutines 完成随机数的生成。然后关闭 channel。当 channel 关闭时，消费者 `consume` goroutine 已经将所有的随机数写入文件，在第 37 行 将 `true` 写入 `done` 这个 channel 中，这个时候 `main` 函数解除阻塞并且打印 `File written successfully`。

现在你可以用任何的文本编辑器打开文件 `concurrent`，可以看到 100 个随机数已经写入 :)

本教程到此结束。希望你能喜欢，祝你愉快。

**上一教程** - [读取文件](https://studygolang.com/articles/14669)

---

via: https://golangbot.com/write-files/

作者：[Naveen Ramanathan](https://golangbot.com/about/)
译者：[amei](https://github.com/amei)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
