首发于：https://studygolang.com/articles/16267

# 关于 Go 中的 panic 的使用以及误用

Go 采用明确的 error 值和类似异常的 panic 机制的方式作为独有的错误处理机制。在文章中将聚焦于 panic 的设计哲学。尝试去理解来自 Go 团队的冲突

首先，这是一个关于语言设计实用性方面的故事，并对实用主义优缺点做了一些思考。

## 介绍 错误和异常 (errors and exceptions)

大多数编程语言支持 exception 作为处理 error 的标准方式，比如 Java，Python。虽然方便，但也会带来许多问题，这就是为什么他们不喜欢其他语言或者风格。对 exception 的主要吐槽点是它们为控制流引入了 "side channel"，当阅读代码的时候，你必须时刻记住这个 exception 引发的流程控制方向。这也导致某些代码阅读起来比较困难[1]。

让我们开始具体谈谈 Go 中的错误处理。我假定你知道 Go 中错误处理的“标准”方式。下面如何打开文件的代码：

```go
f, err := os.Open("file.txt")
if err != nil {
	// handle error here
}
// do stuff with f here
```

如果文件不存在， os.Open() 函数将返回一个非空 error ，在其他语言中这样的错误处理是完全不同的，比如 Python 中的内建 open() 函数将在错误发生时抛出异常。

```python
try:
	with open("file.txt") as f:
		# do stuff with f here
except OSError as err:
	# handle exception err here
```

Python 始终坚持通过 exception 来处理 error。因为这种无处不在的错误处理方式导致经常被吐槽。甚至利用 exception 作为序列结束的信号。到底 exception 的真正含义是什么？以下是来自 Rob Pike 在[邮件中](https://groups.google.com/forum/#!topic/golang-nuts/HOXNBQu5c-Q%5B26-50%5D) 对此的贴切阐述，其中塑造了现有的 Go panic/recover 机制雏形。

> 这正是提案试图避免的那种事情。 Panic 和 recover 不是通常意义的异常机制。通常的方式是将 exception 和一个控制结构相关联，鼓励细粒度的 exception 处理，导致代码往往不易阅读。在 error 和调用一个 panic 之间确实存在差异，而且我们希望这个差异很重要。在 Java 中打开一个文件会抛出异常。在我的经验中，打开文件失败是最平常不过的事。而且还需要我写许多代码来处理这样的 exception。

客观的讲。exception 的支持者嘲笑 Go 的这种过于明确的 error 处理有多方面的原因。首先，请注意上面两个例子中代码的顺序。 在 Python 中，程序的主要流程紧跟在 open 调用之后，并且错误处理被委托给后一阶段（更不用说在许多情况下，异常将被堆栈中更上一级的函数捕获到而不是在此函数中）。 另一方面，在 Go 中，立刻处理错误这种方式，可能会使主程序流程混淆。 此外，Go 的错误处理非常冗长 - 这是该语言的主要吐槽点之一。 我将在后面提到一种可能的方法来解决这个问题。

除了上面 Rob 的引用之外，在 [FAQ](https://golang.org/doc/faq#exceptions) 中总结了 Go 的 exception 哲学。

> 我们认为将异常耦合到控制结构（如 try-catch-finally 惯用语）会导致代码错综复杂。 它还倾向于鼓励程序员标记太多普通错误，例如打开文件失败。

然而，在某些情况下，具有类似异常的机制实际上是有用的 ; 像 Go 这样的高级语言甚至是必不可少的。 这就是存在 panic 和 recover 的原因。

## 偶尔的 panic 是必要的

Go 是一种安全的语言，运行时检查一些严重的编程错误。例如在你访问超出 slice 边界的元素时，这种行为是未定义的，因此 Go 会在运行时 panic。例如下面的小程序。

```go
package main

import (
	"fmt"
)

func main() {
	s := make([]string, 3)
	fmt.Println(s[5])
}
```

程序将终止于一个运行时 error。

```
panic: runtime error: index out of range

goroutine 1 [running]:
main.main()
	/tmp/sandbox209906601/main.go:9 +0x40
```

其他一些会引发 panic 的就是通过值为 nil 的指针访问结构体的字段，关闭已经关闭的 channel 等。怎样选择性的 panic ？可以通过访问 slice 时返回 result，error 两个值的方式实现。也可以将 slice 的元素赋值给一个可能返回 error 的函数，但是这样会将代码变复杂。想象一下，写一个小片段，foo，bar，baz 都只是一个字符串的一个 slice，实现片段之间的拼接。

```go
foo[i] = bar[i] + baz[i]
```

就会变成下面这样冗长的代码：

```go
br, err := bar[i]
if err != nil {
	return err
}
bz, err := baz[i]
if err != nil {
	return err
}
err := assign_slice_element(foo, i, br + bz)
if err != nil {
	return err
 }
 ```
这不是开玩笑，不同语言处理这样的方式是不一样的。如果 slices/lists/arrays 的指针 i 越界了，在 Python 和 Java 中就会抛出异常。C 中没有越界检查，所以你就可以尽情的蹂躏边界外的内存空间，最后将导致程序崩溃或者暴露安全漏洞。C++ 中将采用折中的处理方式。性能优先的模块采用这种不安全的 C 模式，其他模块（比如 std::vector::at）采用抛出异常的方式。

因为上面重写的小片段变得如此冗长是不可接受的。Go 选择了 panic ，这是一种类似异常的机制，在代码中保留了像 bugs 这样最原始的异常条件。

这不只是内建代码能够这样用，自定义代码也可以在任何需要的地方调用 panic。在有些可能导致可怕错误的地方还鼓励使用 panic 抛出 error，比如 bug 或者一些关键因素被违反的时候。比如在 swich 的某个 case 在当前上下文中是不可能发生的，在这种 case 中只有一个 panic 函数。这无形中等价于 Python 中的 raise 或者 C++ 中的 throw。这也强有力的证明了在捕获异常方面 Go 的异常处理的特殊之处。

## Go 中 panic 恢复的限制条件

当一个 panic 在调用栈中的任何地方没有被 caught/recovered 时，程序终将因堆栈溢出而终止。正如上面看到的一样，这种方式对调试来说非常有用。但是现实中却不该是这样的。如果我们要写一个 server，服务于很多的 client。我们不希望因为使用了一个解析数据的库的内部有个 bug 而导致程序崩溃。更好的方式而是应该 catch 到这个 error，输出到日志，并保持 server 能够正常的服务于其他的 client。Go 中的 recover 就是处理这种情况的。下面是 Effective Go 中的示例代码。

```go
func server(workChan <-chan *Work) {
	for work := range workChan {
		go safelyDo(work)
	}
}

func safelyDo(work *Work) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("work failed:", err)
		}
	}()
	do(work)
}
```
咋一看 Go 语言的 panic/recover 像其他语言的错误处理机制。其故意设置的限制使得在处理异常繁重的代码时不易出现常见的一些问题。下面是引用的 FAQ 上的另一段话：

> Go 语言也有一对内建函数从真实的异常条件中 signal 和 recover。recover 机制仅仅只在函数块的出现一个 error 之后被执行。这足矣处理大灾难。但是需要额外的控制结构，如果用得好，能够写出非常清晰的错误处理的代码。

之前引用的 Rob Pike 的话也是很贴切的：

> 我们建议将处理和函数相互联系起来 -- 一个枯燥的函数，从而故意让这种方式难用。我们希望你也考虑一下 panic，不用函数去处理错误基本上是不可能的事情。如果你想保护你的代码，为整个程序使用 1~2 个的 recover 是必要的。如果你已经觉得难以区分不同的 panic，那你就没 get 到真正的意义。

recover 调用的一个重要限制就是只能在 defer 代码块中，它不能将控制权交给任何一个调用点，但是可以做一些清理工作或扭曲函数的返回值。上面 Python 处理打开文件错误的方式在 Go 中并不起作用。在不调整代码的情况下，我们不能捕获到 OSError 然后尝试打开另外一个文件（或者创建一个新文件）。

这个限制遵循了一个重要的代码指南 -- 将 panic 控制在包边界内。不让 panic 在包的公用接口中出现。在每个对包外公开的函数和方法都应该 recover 到内部的 panic 并且将这些 panic 转换为错误信息，这使得 panic 非常的友好，即使高可用的服务器在这种情况下可能还在外部使用了 recover 来防止内部 panic 造成程序终止。

## panic 的调用艺术

每种语言的特性注定都会被滥用。这就是面对的真实的程序日常，Go 语言也不例外。这不是说所有的滥用都是绝对的错误，而是指一个特性在实际使用中和它最初被设计的目的并不一致。看看下面一个真实的例子，是关于 Go 1.10 标准库 fmt/scan.go 中的 scanInt 方法。

```go
func (s *ss) scanInt(verb rune, bitSize int) int64 {
	if verb == 'c' {
		return s.scanRune(bitSize)
	}
	s.SkipSpace()
	s.notEOF()
	base, digits := s.getBase(verb)
	// ... other code
}
```
里面中任何一个函数，SkipSpace(),notEOF(),getBase() 都有可能出错，而错误处理在什么地方呢？事实上，这个包以及其他的标准库，内部都是使用 panic 来处理内部的错误，这些 panic 将会在公共的 API 中被 recover 到（就像 Token 方法一样）并且将其转换为 error，如果我们明确的处理这些错误，代码将变成下面这样，非常的繁琐[2]。

```go
if err := s.SkipSpace(); err != nil {
	return err
}
if err := s.notEOF(); err != nil {
	return err
}
base, digits, err := s.getBase(verb)
if err != nil {
	return err
}
// ... other code
```

当然 panic 不是解决这种情况的唯一方式，就像 Rob Pike 说的那样，[错误就是值](https://blog.golang.org/errors-are-values)，因此他们是可编程的。我们能够设计一些更加巧妙的方式不使用类异常处理机制也能控制好代码流。其他语言有一些有用的特性使得这更简单，比如 Rust 中的 ？操作符[3]，让返回的表达式自动的传递 error，所以伪代码中我们可以这样写：

```
s.SkipSpace()?
s.notEOF()?
base, digits := s.getBase(verb)?
```

但是 Go 中并没有 "?" Go 的核心团队选择使用 panic 来代替，甚至在 [Effective Go](https://golang.org/doc/effective_go.html) 中还宽恕这种模式。

> 有了 recover 的这种模式，通过调用 panic 使得在任意地方调用的函数都能摆脱不良的情况，我们可以使用这个方式来简化复杂软件中的错误处理。

这种 recover 和 panic 的方式在其他几个地方我也有看到：

- fmt/scan.go
- JSon/encode.go
- text/template/parse/parser.go

## 但是这不是错误的吗？

我很同情那些被诱惑引诱的人们，他们的呼声很强烈。但是也不能动摇违背语言设计的最初原则的事实。再次引用上面 Rob Pike 说过的话：

> 在我的经历中，没有比打开文件失败再普遍的异常了。

但是还有比解析过程中遇到没有预期到的字符类异常更少的吗？这不也是解析器遇到的司空见惯的错误吗？ Rob Pike 接下来还说：

> 我们希望你也考虑一下 panic，不用函数去处理错误基本上是不可能的事情。

解析错误真的很稀少吗？ fmt/scan.go 包下的很多函数使用 panic，因为这是它们来发出错误信号的方式。

> 如果你已经开始担心如何区分这些不同种类的 panic 时，那你就没 get 到真正的意义。

下面是 fmt/scan.go 中对错误的处理：

```go
func errorHandler(errp *error) {
	if e := recover(); e != nil {
		if se, ok := e.(scanError); ok { // catch local error
			*errp = se.err
		} else if eof, ok := e.(error); ok && eof == io.EOF { // out of input
			*errp = eof
		} else {
			panic(e)
		}
	}
}
```
这就不用担心如何区分不同的 panic 了吗？

## 总结 实用性 VS 简洁

我这里的目的不是攻击 Go 标准库的开发者，正如我所提到的，我清楚的知道在调用栈很深或在错误信号处理序列司空见惯的情况下为什么这么吸引人。我真心的希望 Go 将提出一些语法使得繁重的错误处理变得容易，从而使这个讨论没有实际意义。

有时候，做一个实用主义者比一个狂热者更好。如果某个语言特性对解决某个问题非常有帮助，甚至超出了经典的使用领域，使用它可能比坚持原则并最终使用复杂的代码更好。有点像我以前坚持的一个观点 --[在 C 中使用 Goto 来进行错误处理](https://eli.thegreenplace.net/2009/04/27/using-goto-for-error-handling-in-c) Go 指南很明确，并且对恢复的限制非常巧妙 - 即使用于解析器中的控制流程，也比经典异常更难以滥用。

有趣的是，当这个问题首先引起我的注意时，我正在研究 JSon/encode.go 包的源代码。 事实证明，它最近被修复使用经典的错误处理！是的，一些代码变得更加冗长，从这样：

```go
if destring {
	switch qv := d.valueQuoted().(type) {
		case nil:
			d.literalStore(nullLiteral, subv, false)
		case string:
			d.literalStore([]byte(qv), subv, true)
		// ... other code
```
变成了这样 :

```go
if destring {
	q, err := d.valueQuoted()
	if err != nil {
		return err
	}
	switch qv := q.(type) {
	case nil:
		if err := d.literalStore(nullLiteral, subv, false); err != nil {
			return err
		}
	case string:
		if err := d.literalStore([]byte(qv), subv, true); err != nil {
			return err
		}
```
但总的来说，它并不是那么糟糕，对于 Go coder 来说肯定不会陌生。 它给了我希望 :-)

[1]: C ++ 的异常安全保证集是所涉及的一些复杂性的一个很好的例子。

[2]: 如果你花一些时间去读一读[提出 recover 机制的邮件](https://groups.google.com/forum/#!topic/golang-nuts/HOXNBQu5c-Q%5B26-50%5D)
你会发现 Russ Cox 在解析二进制流时会提到类似的问题，以及如何在整个过程中传播错误。

[3]: 甚至 C ++ 也有类似的模式，你可以在一些使用标准返回类型的代码库中找到它。 通常名为 ASSIGN_OR_RETURN
的宏在 Google 发布的 C ++ 代码中很流行，并且出现在 LLVM 等其他地方。

---

via: https://eli.thegreenplace.net/2018/on-the-uses-and-misuses-of-panics-in-go/

作者：[Eli Bendersky](https://eli.thegreenplace.net/pages/about)
译者：[zouxinjiang](https://github.com/zouxinjiang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
