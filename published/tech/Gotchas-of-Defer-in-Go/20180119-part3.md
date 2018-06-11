已发布：https://studygolang.com/articles/12319

# Golang 中 defer 的五个坑 - 第三部分

> 译注：全文总共有四篇，本文为同系列文章的第三篇

- [第一部分](https://studygolang.com/articles/12061)
- [第二部分](https://studygolang.com/articles/12136)

本文将侧重于讲解使用 defer 的一些技巧

如果你对 defer 的基本操作还没有清晰的认识，请先阅读这篇 [文章](https://blog.learngoprogramming.com/golang-defer-simplified-77d3b2b817ff) （GCTT 出品的译文 https://studygolang.com/articles/11907）。

## #1 —— 在延迟调用函数的外部使用 recover

你总是应该在被延迟函数的内部调用 `recover()` ，当出现一个 *panic* 异常时，在 *defer* 外调用
`recover()` 将无法捕获这个异常，而且 `recover()` 的返回值会是 *nil* 。

例子

```go
func do() {
	recover()
	panic("error")
}
```

输出

*recover* 并没有成功捕获异常。

```
panic: error
```

### 解决方案

在延迟调用的函数内部使用 `recover()` 就能够避免这个问题。

```go
func do() {
	defer func() {
		r := recover()
		fmt.Println("recovered:", r)
	}()

	panic("error")
}
```

输出

```
recovered: error
```

## #2 —— 在错误的位置使用 defer

这个陷阱来自于这篇 [Go 的 50 个阴影](http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/#anameclose_http_resp_bodyaclosinghttpresponsebody)。

例子

当 `http.Get` 失败时会抛出异常。

```go
func do() error {
	res, err := http.Get("http://notexists")
	defer res.Body.Close()
	if err != nil {
		return err
	}

	// ..code...

	return nil
}
```

输出

```
panic: runtime error: invalid memory address or nil pointer dereference
```

### 发生了什么？

因为在这里我们并没有检查我们的请求是否成功执行，当它失败的时候，我们访问了 *Body* 中的空变量 *res* ，因此会抛出异常

### 解决方案

总是在一次成功的资源分配下面使用 *defer* ，对于这种情况来说意味着：当且仅当 *http.Get* 成功执行时才使用 *defer*

```go
func do() error {
	res, err := http.Get("http://notexists")
	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return err
	}

	// ..code...

	return nil
}
```

在上述的代码中，当有错误的时候，*err* 会被返回，否则当整个函数返回的时候，会关闭 *res.Body* 。

### 旁注 1

在这里，你同样需要检查 *resp* 的值是否为 *nil* ，这是 *http.Get* 中的一个警告。通常情况下，出错的时候，返回的内容应为空并且错误会被返回，可当你获得的是一个重定向 *error* 时， *resp* 的值并不会为 *nil* ，但其又会将错误返回。上面的代码保证了无论如何 *Body* 都会被关闭，如果你没有打算使用其中的数据，那么你还需要丢弃已经接收的数据。更多 [详情](http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/#anameclose_http_resp_bodyaclosinghttpresponsebody)。

## #3 —— 不检查错误

简单地将清理的逻辑委托给 *defer* 并不意味着资源的释放就万无一失了，你也可能会错失有用的报错信息，让一些潜在的问题石沉大海。

### 反面教材

在这里，`f.Close()` 可能会返回一个错误，可这个错误会被我们忽略掉

```go
func do() error {
	f, err := os.Open("book.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	// ..code...

	return nil
}
```

### 改进一下

最好还是检查可能的错误而不是直接交给 *defer* 就完事，你可以把 *defer* 内的代码写成一个帮助函数来简化我们的代码，这里为了讲解方便就没有进行简化。

```go
func do() error {
	f, err := os.Open("book.txt")
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			// log etc
		}
	}()

	// ..code...

	return nil
}
```

### 再改进一下

你也可以通过命名的返回变量来返回 *defer* 内的错误。

```go
func do() (err error) {
	f, err := os.Open("book.txt")
	if err != nil {
		return err
	}

	defer func() {
		if ferr := f.Close(); ferr != nil {
			err = ferr
		}
	}()

	// ..code...

	return nil
}
```

### 旁注 2

你可以使用这个 [包](https://godoc.org/github.com/pkg/errors) 来整合多个不同的错误，这会非常必要因为 defer 中的 *f.Close* 可能会把之前的错误也覆盖掉，将多个错误包裹在一起能够将所有的错误信息都写入日志，在诊断问题的时候能有更多的依据。

你也可以使用这个 [包](https://github.com/kisielk/errcheck) 来查看你遗漏的本应该检查错误的地方。

## #4 —— 释放相同的资源

在第三小节中有一个小小的警告：如果你尝试使用相同的变量释放不同的资源，那么这个操作可能无法正常执行。

例子

这段看似没什么问题的代码尝试第二次关闭相同的资源。第二个 *变量 f* 会被关闭两次，因为 *f 变量* 会因第二个资源而改变它的值

```go
func do() error {
	f, err := os.Open("book.txt")
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			// log etc
		}
	}()

	// ..code...

	f, err = os.Open("another-book.txt")
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			// log etc
		}
	}()

	return nil
}
```

输出

```
closing resource #another-book.txt
closing resource #another-book.txt
```

### 发生了什么

正如我们所看到的，当延迟函数执行时，只有最后一个变量会被用到，因此，*f 变量* 会成为最后那个资源 (another-book.txt)。而且两个 *defer* 都会将这个资源作为最后的资源来关闭

### 解决方案

```go
func do() error {
	f, err := os.Open("book.txt")
	if err != nil {
		return err
	}

	defer func(f io.Closer) {
		if err := f.Close(); err != nil {
			// log etc
		}
	}(f)

	// ..code...

	f, err = os.Open("another-book.txt")
	if err != nil {
		return err
	}

	defer func(f io.Closer) {
		if err := f.Close(); err != nil {
			// log etc
		}
	}(f)

	return nil
}
```

输出

```
closing resource #another-book.txt
closing resource #book.txt
```

你也可以使用函数来避免上述问题的发生，参考我在 [这里](https://blog.learngoprogramming.com/gotchas-of-defer-in-go-1-8d070894cb01#ac69) 讲过的开闭模式。

## #5 —— panic/recover 会取得并返回任意类型

你可能认为你总是需要往 *panic* 中传 *string* 或 *error* 类型的数据

### 传入 string

```go
func errorly() {
	defer func() {
		fmt.Println(recover())
	}()

	if badHappened {
		panic("error run run")
	}
}
```

输出

```
"error run run"
```

### 传入 error

```go
func errorly() {
	defer func() {
		fmt.Println(recover())
	}()

	if badHappened {
		panic(errors.New("error run run")
	}
}
```

输出

```
"error run run"
```

### 传入任意类型

正如你所看到的 *panic* 可以接收 *string* 以及 [error类型](https://golang.org/pkg/builtin/#error) 。这意味着事实上你可以给 panic 传 "任意类型" 的数据并能够在 *defer* 中使用 *recover* 来获取这个数据。

```go
type myerror struct {}

func (myerror) String() string {
	return "myerror there!"
}

func errorly() {
	defer func() {
		fmt.Println(recover())
	}()

	if badHappened {
		panic(myerror{})
	}
}
```

### 为什么可以这么写？

这是因为 *panic* 的函数签名显示它可以接收 *interface{}* 类型，我们可以将它理解为 Go 中的 "任意类型"

这是 *panic* 的签名

```go
func panic(v interface{})
```

*recover* 的签名

```go
func recover() interface{}
```

因此，基本上它会这样运行

```
panic(value) -> recover() -> value
```

*recover* 会把传入 *panic* 的值返回出来

这一部分就先告一段落了，我们第四部分见！

---

via: https://blog.learngoprogramming.com/5-gotchas-of-defer-in-go-golang-part-iii-36a1ab3d6ef1

作者：[Inanc Gumus](https://blog.learngoprogramming.com/@inanc)
译者：[yujiahaol68](https://github.com/yujiahaol68)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
