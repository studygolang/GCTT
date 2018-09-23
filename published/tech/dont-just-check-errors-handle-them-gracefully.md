已发布：https://studygolang.com/articles/12484

# 优雅的处理错误，而不仅仅只是检查错误

这篇文章摘取至我在日本东京举办的 [GoCon spring conference](https://gocon.connpass.com/event/27521/) 上的演讲稿。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/error-handle/ba5a9ada.png)

## 错误只是一些值

我花了很多时间来思考如何在 Go 中处理错误是最好的。我真希望能有一种简单直接的方式来处理错误，一些我们只要让 Go 程序员记住就能使用的规则，就像教数学或字母表一样。

然而，我得到的结论是：处理错误不止有一种方式。我认为 Go 处理错误的方式可以划分为 3 种主要的策略。

## 标记错误策略

第一种错误处理策略，我称之为标记错误

```go
if err == ErrSomething { … }
```

这个名字来源于在实际编程中，使用一个指定的值来表示程序已经无法继续执行。所以在 Go 中我们使用一个指定的值来表示错误。

例如：系统包里面的 `io.EOF` 或是在 `syscall` 包中更底层些的常量错误例如 `syscall.ENOENT`。

甚至还有标记表示没有错误发生例如：`path/filepath.Walk` 中的 `go/build.NoGoError` 和 `path/filepath.SkipDir`。

使用标记值是灵活性最差的一种错误处理策略，调用者必须使用相等操作符来比较返回值和预先定义的值。当你想要提供更多的相关信息时，返回不同的错误值会破坏等式检查操作。

即使通过 `fmt.Errorf` 来提供更多的信息也会干扰调用者的等式测试，调用者必须去看 Error 方法输出的结果是否匹配某个指定的字符串。

### 永远不要检查 `error.Error` 的输出

顺便说一下，我相信你永远都不需要检查 `error.Error` 方法的返回值。 `error` 接口中的 `Error` 方法是提供给使用者查看的信息，而不是用来给代码做判断的。

这些信息应该在日志文件中或者是显示屏上出现，你不需要通过检查这些信息来改变程序行为。

我知道有时候这样很难，就像有些人在 twitter 上提到的那样，这条建议在写测试的时候不适用。尽管如此，在我看来，作为一种编码风格，你应该避免比较字符型的错误信息。

### 标记错误成为公开 API 的一部分

如果你的公开函数或方法返回了一些指定的错误值，那么这些值必须是公开的，当然也需要在文档中有所描述。这些加入到你的 API 中了。

如果你的 API 定义了一个返回指定错误的接口，那么所有该接口的实现都必须只返回这个错误，就算能提供更多的其他信息也不应该返回除了指定错误之外的信息。

我们可以在 `io.Reader` 中看到这样的处理方式。 `io.Copy` 要求 reader 实现返回 `io.EOF` 通知调用者没有更多的数据了，但是这并不是一个错误。

### 标记错误在两个包之间制造了依赖关系

最大的问题是标记错误在两个包之间制造了源码层面的依赖关系。例如：检查一个错误是否是 `io.EOF` 你的代码必须引入 `io` 包。

这个例子看起来没那么糟糕，因为这是很普通的操作。但是想象一下，项目中很多包导出错误值，而其他包必须导入对应的包才能检查错误条件，这样就违背了低耦合的设计原则。

我参与过的一个大型项目，使用的就是这种错误处理模式，我可以告诉你不好的设计所带来的循环引入问题近在咫尺。

### 结论：避免使用标记错误策略

所以，我的建议是避免在代码中使用标记错误处理策略。在标准库中有些情况使用了这种处理方式，但是这并不是你应该效仿的一种处理模式。

如果有人要求你从你的包里面暴露一个错误值，你应该礼貌的拒绝他，并提供一个替代方案，也就是下面将要提到的方法。

## 错误类型

错误类型是我想讨论的第二种 Go 错误处理模式。

```go
if err, ok := err.(SomeType); ok { … }
```

错误类型是你创建的实现 error 接口的类型。在下面的例子中， MyError 类型记录了文件，行号，相关的错误信息。

```go
type MyError struct {
	Msg string
	File string
	Line int
}

func (e *MyError) Error() string {
	return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Msg)
}

return &MyError{"Something happened", "server.go", 42}
```

因为 MyError 是一个类型，所以调用者可以使用 type assertion 从 error 中获取相关信息。

```go
err := something()
switch err := err.(type) {
case nil:
	// call succeeded, nothing to do
case *MyError:
	fmt.Println(“error occurred on line:”, err.Line)
default:
	// unknown error
}
```

错误类型比标记错误最大的改进就是通过封装底层的错误来提供更多的相关信息。

一个绝佳的例子就是 os.PathError 除了底层错误外还提供了使用哪个文件，执行哪个操作等相关信息。

```go
// PathError records an error and the operation
// and file path that caused it.
type PathError struct {
	Op   string
	Path string
	Err  error // the cause
}

func (e *PathError) Error() string
```

### 错误类型存在的问题

调用者可以使用 type assertion 或者 type switch，error 类型必须是公开的。

如果你的代码实现了一个约定指定错误类型的接口，那么所有这个接口的实现者，都要依赖于定义这个错误类型的包。

对包的错误类型的过度暴露，使调用者和包之间产生了很强的耦合性，导致了 API 的脆弱性。

### 结论：避免错误类型

虽然错误类型在发生错误时能够捕捉到更多的环境信息，比标记错误要好一些，但是错误类型也存在很多和标记错误一样的问题。

所以，在这里我的建议是避免使用错误类型，至少避免使他们成为你 API 接口的一部分。

## 封装错误

现在我们到了第三个错误处理分类。在我看来这个是灵活性最好的处理策略，在调用者和你的代码之间产生的耦合度最低。

我管这种处理方式叫做封装错误 (Opaque errors)，因为当你发现有错误发生时，你无法知道内部的错误情况。作为调用者，你只知道调用的结果成功或者失败。

封装错误处理方式只返回错误不去猜测他的内容。如果你采用了这种处理方式，那么错误处理在调试方面会变得非常有价值。

```go
import “github.com/quux/bar”

func fn() error {
	x, err := bar.Foo()
	if err != nil {
		return err
	}
	// use x
}
```

例如：Foo 的调用约定没有指定在发生错误时会返回哪些相关信息，这样 Foo 函数的开发者就可以自由的提供相关错误信息，并且不会影响到和调用者之间的约定。

### 断言行为，而不是类型

在少数情况下，这种二元错误处理方案是不够的。

例如：在和进程外交互的时候，比如网络活动，需要调用者评估错误情况来决定是否需要重试操作。

在这种情况下我们断言错误实现指定的行为要比断言指定类型或值好些。看看下面的例子：

```go
type temporary interface {
	Temporary() bool
}

// IsTemporary returns true if err is temporary.
func IsTemporary(err error) bool {
	te, ok := err.(temporary)
	return ok && te.Temporary()
}
```

我们可以传任何错误给 IsTemporary 来判断错误是否需要重试。

如果错误没有实现 temporary 接口；那么就没有 Temporary 方法，那么错误就不是 temporary。

如果错误实现了 Temporary，如果 Temporary 返回 true 那么调用者就可以考虑重试该操作。

这里的关键点是这个实现逻辑不需要导入定义错误的包或者了解任何关于错误的底层类型，我们只要简单的关注它的行为即可。

### 优雅的处理错误，而不仅仅只是检查错误

这引出了我想谈的第二个 Go 语言的格言：优雅的处理错误，而不仅仅只是检查错误。你能在下面的代码中找出错误么？

```go
func AuthenticateRequest(r *Request) error {
	err := authenticate(r.User)
	if err != nil {
		return err
	}
	return nil
}
```

一个很明显的建议是上面的代码可以简化为

```go
return authenticate(r.User)
```

但是这只是个简单的问题，任何人在代码审查的时候都应该看到。更根本的问题是这段代码看不出来原始错误是在哪里发生的。

如果 authenticate 返回错误， 那么 AuthenticateRequest 将会返回错误给调用者，调用者也一样返回。 在程序的最上一层主函数块内打印错误信息到屏幕或者日志文件，然而所有信息就是 No such file or directory

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/error-handle/53c71467.png)

没有错误发生的文件，行号等信息，也没有调用栈信息。代码的编写者必须在一堆函数中查找哪个调用路径会返回 file not found 错误。

Donovan 和 Kernighan 写的 The Go Programming Language 建议你使用 fmt.Errorf 在错误路径中增加相关信息

```go
func AuthenticateRequest(r *Request) error {
	err := authenticate(r.User)
	if err != nil {
		return fmt.Errorf("authenticate failed: %v", err)
	}
	return nil
}
```

就像我们在前面提到的，这个模式不兼容标记错误或者类型断言，因为转换错误值到字符串，再和其他的字符串合并，再使用 fmt.Errorf 转换为error 打破了对等关系，破坏了原始错误的相关信息。

### 注解错误

我在这里建议一个给错误添加相关信息的方法，要用到一个简单的包。代码在 [github.com/pkg/errors](https://godoc.org/github.com/pkg/errors)。这个包有两个主要的函数：

```go
// Wrap annotates cause with a message.
func Wrap(cause error, message string) error
```

第一个函数是封装函数 Wrap ，输入一个错误和一个信息，生成一个新的错误返回。

```go
// Cause unwraps an annotated error.
func Cause(err error) error
```

第二个函数是 Cause ，输入一个封装过的错误，解包之后得到原始的错误信息。

使用这两个函数，我们现在可以给任何错误添加相关信息，并且在我们需要查看底层错误类型的时候可以解包查看。下面的例子是读取文件内容到内存的函数。

```go
func ReadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "open failed")
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "read failed")
	}
	return buf, nil
}
```

我们使用这个函数写一个读取配置文件的函数，然后在 main 中调用。

```go
func ReadConfig() ([]byte, error) {
	home := os.Getenv("HOME")
	config, err := ReadFile(filepath.Join(home, ".settings.xml"))
	return config, errors.Wrap(err, "could not read config")
}

func main() {
	_, err := ReadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
```

如果 ReadConfig 发生错误，由于使用了 errors.Wrap，我们可以得到一个 K&D 风格的包含相关信息的错误

```
could not read config: open failed: open /Users/dfc/.settings.xml: no such file or directory
```

因为 errors.Wrap 生成了发生错误时的调用栈信息，所以我们可以查看额外的调用栈调试信息。这又是一个同样的例子，但是这次我们用 errors.Print 替换 fmt.Println

```go
func main() {
	_, err := ReadConfig()
	if err != nil {
		errors.Print(err)
		os.Exit(1)
	}
}
```

我们会得到如下的信息：

```
readfile.go:27: could not read config
readfile.go:14: open failed
open /Users/dfc/.settings.xml: no such file or directory
```

第一行来至 ReadConfig， 第二行来至 os.Open 的 ReadFile， 剩下的来至 os 包，没有携带位置信息。

现在我们介绍了关于打包错误生成栈的概念，我们需要谈谈如何解包。下面是 errors.Cause 函数的作用。

```go
// IsTemporary returns true if err is temporary.
func IsTemporary(err error) bool {
	te, ok := errors.Cause(err).(temporary)
	return ok && te.Temporary()
}
```

操作中，当你需要检查一个错误是否匹配一个指定值或类型时，你需要先使用 errors.Cause 获取原始错误信息

### 只处理一次错误

最后我想要说的是，你只需要处理一次错误。处理错误意味着检查错误值，然后作出决定。

```go
func Write(w io.Writer, buf []byte) {
	w.Write(buf)
}
```

如果你不需要做出决定，你可以忽略这个错误。在上面的例子可以看到我们忽略了 `w.Write` 返回的错误。

但是在返回一个错误时做出多个决定也是有问题的。

```go
func Write(w io.Writer, buf []byte) error {
	_, err := w.Write(buf)
	if err != nil {
		// annotated error goes to log file
		log.Println("unable to write:", err)

		// unannotated error returned to caller
		return err
	}
	return nil
}
```

在这个例子中，如果 Write 发生错误， 一行信息会写入日志，记录发送错误的文件和行号，同时把错误返回给调用者，同样的调用者也可能会写入日志，然后返回，直到程序的最顶层。

日志文件里就会出现一堆重复的信息，但是在程序最顶层获得的原始错误却没有任何相关信息。

```go
func Write(w io.Write, buf []byte) error {
	_, err := w.Write(buf)
	return errors.Wrap(err, "write failed")
}
```

使用 errors 包可以让你在 error 里面加入相关信息，并且内容是可以被人和机器所识别的。

## 结论

最后，错误是你提供的包中公开 API 的一部分，要像其他公开 API 一样小心对待。

为了获得最大的灵活性，我建议你尝试把所有的错误当做封装错误来处理，在那些无法做到的情况下，断言行为错误，而不是类型或值。

尽可能少的在你程序中使用标记错误，在错误发生时尽早的使用 errors.Wrap 打包封装为封装错误。

## 相关文章

1. [检查错误](https://dave.cheney.net/2014/12/24/inspecting-errors)
2. [常量错误](https://dave.cheney.net/2016/04/07/constant-errors)
3. [调用栈和错误包](https://dave.cheney.net/2016/06/12/stack-traces-and-the-errors-package)
4. [返回的错误和异常](https://dave.cheney.net/2016/06/12/stack-traces-and-the-errors-package)

---

via: https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully

作者：[Dave Cheney](https://dave.cheney.net/about)
译者：[tyler2018](https://github.com/tyler2018)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
