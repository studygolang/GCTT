# 关于Go中的panic的使用以及误用

---

Go采用明确的error值和类似异常的panic机制的方式作为独有的错误处理机制。在文章中将聚焦于panic的设计哲学。尝试去理解来自Go团队的冲突

首先，这是一个关于语言设计实用性方面的故事，并对实用主义优缺点做了一些思考。

## 介绍 错误和异常(errors and exceptions)
---

大多数编程语言支持exception作为处理error的标准方式，比如Java，Python。虽然方便，但也会带来许多问题，这就是为什么他们不喜欢其他语言或者风格。对exception的主要吐槽点是它们为控制流引入了“side channel”，当阅读代码的时候，你必须时刻记住这个exception引发的流程控制方向。这也导致某些代码阅读起来比较困难。

让我们开始具体谈谈Go中的错误处理。我假定你知道Go中错误处理的“标准”方式。下面如何打开文件的代码：
```go
f, err := os.Open("file.txt")
if err != nil {
  // handle error here
}
// do stuff with f here
```
如果文件不存在， os.Open()函数将返回一个非空error，在其他语言中这样的错误处理是完全不同的，比如 Python 中的内建open()函数将在错误发生时抛出异常
```python
try:
  with open("file.txt") as f:
    # do stuff with f here
except OSError as err:
  # handle exception err here
```
Python始终坚持通过exception来处理error。因为这种无处不在的方式错误处理方式导致经常被吐槽。甚至利用exception作为序列结束的信号。
到底exception的真正含义是什么？以下是来自Rob Pike在邮件中对此的贴切阐述，其中塑造了现有的Go panic/recover机制雏形。

这正是提案试图避免的那种事情。 Panic和recover不是通常意义的异常机制。通常的方式是将exception和一个控制结构相关联，鼓励细粒度的exception处理，导致代码往往不易阅读。
在error和调用一个panic之间确实存在差异，而且我们希望这个差异很重要。在java中打开一个文件会抛出异常。在我的经验中，打开文件失败是最平常不过的事。而且还需要我写许多代码来
处理这样的exception。

客观的讲。exception的支持者嘲笑Go的这种过于明确的error处理有多方面的原因。首先，请注意上面两个例子中代码的顺序。 在Python中，程序的主要流程紧跟在open调用之后，
并且错误处理被委托给后一阶段（更不用说在许多情况下，异常将被堆栈中更上一级的函数捕获到而不是在此函数中）。 另一方面，在Go中，立刻处理错误这种方式，可能会使主程序流程混淆。 此外，Go的错误处理非常冗长 - 这是该语言的主要吐槽点之一。 我将在后面提到一种可能的方法来解决这个问题。

除了上面Rob的引用之外，在FAQ中总结了Go的exception哲学。
我们认为将异常耦合到控制结构（如try-catch-finally惯用语）会导致代码错综复杂。 它还倾向于鼓励程序员标记太多普通错误，例如打开文件失败。

然而，在某些情况下，具有类似异常的机制实际上是有用的; 像Go这样的高级语言甚至是必不可少的。 这就是存在panic和recover的原因。

## 偶尔的panic是必要的
Go是一种安全的语言，运行时检查一些严重的编程错误。例如在你访问超出slice边界的元素时，这种行为是未定义的，然而Go会在运行时panic掉。例如下面的小程序。
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
程序将终止于一个运行时error。
```go
panic: runtime error: index out of range

goroutine 1 [running]:
main.main()
  /tmp/sandbox209906601/main.go:9 +0x40
```
其他一些会引发panic的就是通过值为nil的指针访问结构体的字段，关闭已经关闭的channel等。
怎样选择性的panic？可以通过访问slice时返回result，error两个值的方式实现。也可以将slice的元素赋值
给一个可能返回error的函数，但是这样会将代码变复杂。想象一下，写一个小片段，foo，bar，baz都只是一个字符串的
一个slice，实现片段之间的拼接。
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
这不是开玩笑，不同语言处理这样的方式是不一样的。如果slices/lists/arrays的指针i越界了，在Python和Java中就会抛出异常。
C中没有越界检查，所以你就可以尽情的蹂躏边界外的内存空间，最后将导致程序崩溃或者暴露安全漏洞。C++中将采用折中的处理方式。
性能优先的模块采用这种不安全的C模式，其他模块（比如std::vector::at）采用抛出异常的方式。

因为上面重写的小片段变得如此冗长是不可接受的。Go选择了panic，这是一种类似异常的机制，在代码中保留了像bugs这样最原始的异常条件。
这不只是内建代码能够这样用，自定义代码也可以在任何需要的地方调用panic。在有些可能导致可怕错误的地方还鼓励使用panic抛出error，比如bug或者一些关键因素被违反的时候。
比如在swich的某个case在当前上下文中是不可能发生的，在这种case中只有一个panic函数。这无形中等价于Python中的raise或者C++中的throw
这也强有力的证明了在捕获异常方面Go的异常处理的特殊之处。

## Go中panic恢复的限制条件
当一个panic在调用栈中的任何地方没有被 caught/recovered 时，程序终将因堆栈溢出而终止。正如上面看到的一样，这种方式对调试来说非常有用。
但是现实中却不该是这样的。如果我们要写一个server，服务于很多的client。我们不希望因为使用了一个解析数据的库的内部有个bug而导致程序崩溃。
更好的方式而是应该catch到这个error，输出到日志，并保持server能够正常的服务于其他的client。Go中的recover就是处理这种情况的。下面是 Effective Go中的
示例代码。
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
咋一看Go语言的panic/recover像其他语言的错误处理机制。其故意设置的限制使得在处理异常繁重的代码时不易出现常见的一些问题。
下面是引用的FAQ上的另一段话：

Go语言也有一对内建函数从真实的异常条件中signal和recover。recover机制仅仅只在函数块的出现一个error之后被执行。这足矣处理大灾难
但是需要额外的控制结构，如果用得好，能够写出非常清晰的错误处理的代码。

之前引用的Rob Pike的话也是很贴切的：

我们建议将处理和函数相互联系起来--一个枯燥的函数，从而故意让这种方式难用。我们希望你也考虑一下panic
不用函数去处理错误基本上是不可能的事情。如果你想保护你的代码，为整个程序使用1~2个的recover是必要的。
如果你已经觉得难以区分不同的panic，那你就没get到真正的意义。

recover调用的一个重要限制就是只能在defer代码块中，它不能将控制权交给任何一个调用点，但是可以做一些清理
工作或扭曲函数的返回值。上面Python处理打开文件错误的方式在Go中并不起作用。在不调整代码的情况下，我们不能捕获到OSError然后尝试
打开另外一个文件（或者创建一个新文件）。

这个限制遵循了一个重要的代码指南--将panic控制在包边界内。不让panic在包的公用接口中出现。在每个
对包外公开的函数和方法都应该recover到内部的panic并且将这些panic转换为错误信息，这使得panic非常的
友好，即使高可用的服务器在这种情况下可能还在外部使用了recover来防止内部panic造成程序终止。

## panic的调用艺术
每种语言的特性注定都会被滥用。这就是面对的真实的程序日常，Go语言也不例外。这不是说所有的滥用都是绝对的
错误，而是指一个特性在实际使用中和它最初被设计的目的并不一致。
看看下面一个真实的例子，是关于Go 1.10标准库fmt/scan.go中的scanInt方法。
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
里面中任何一个函数，SkipSpace(),notEOF(),getBase()都有可能出错，而错误处理在什么地方呢？事实上，这个包以
及其他的标准库，内部都是使用panic来处理内部的错误，这些panic将会在公共的API中被recover到（就像Token方法一样）
并且将其转换为error，如果我们明确的处理这些错误，代码将变成下面这样，非常的繁琐。
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
当然panic不是解决这种情况的唯一方式，就像Rob Pike说的那样，错误就是值，因此他们是可
编程的。我们能够设计一些更加巧妙的方式不使用类异常处理机制也能控制好代码流。其他语言有
一些有用的特性使得这更简单，比如Rust中的？操作符，让返回的表达式自动的传递error，所以
伪代码中我们可以这样写：
```Rust
s.SkipSpace()?
s.notEOF()?
base, digits := s.getBase(verb)?
```
但是Go中并没有？Go的核心团队选择使用panic来代替，甚至在Effective Go中还宽恕这种模式。

有了recover的这种模式，通过调用panic使得在任意地方调用的函数都能拜托不良的情况，
我们可以使用这个方式来简化复杂软件中的错误处理。

这种recover和panic的方式在其他几个地方我也有看到：
- fmt/scan.go
- json/encode.go
- text/template/parse/parser.go

## 但是这不是错误的吗？
我很同情那些被诱惑引诱的人们，他们的呼声很强烈。但是也不能动摇违背语言设计的最初原则的事实。
再次引用上面Rob Pike说过的话：
在我的经历中，没有比打开文件失败再普遍的异常了。
但是还有比解析过程中遇到没有预期到的字符类异常更少的吗？这不也是解析器遇到的司空见惯的错误吗？
Rob Pike接下来还说：
我们希望你也考虑一下panic，不用函数去处理错误基本上是不可能的事情。

解析错误真的很稀少吗？fmt/scan.go包下的很多函数采用panic因为通过这种方式来发出错误信号。

如果你已经开始担心如何区分这些不同种类的panic时，那你就没get到真正的意义。

下面是fmt/scan.go中对错误的处理：
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
这就不用担心如何区分不同的panic了吗？

## 总结 实用性VS简洁
我这里的目的不是攻击Go标准库的开发者，正如我所提到的，我清楚的知道在调用栈很深或在
错误信号处理序列司空见惯的情况下为什么这么吸引人。我真心的希望Go将提出一些语法使得繁重的
错误处理变得容易，从而使这个讨论没有实际意义。

有时候，做一个实用主义者比一个狂热者更好。如果某个语言特性对解决某个问题非常有帮助，甚至超出
了经典的使用领域，使用它可能比坚持原则并最终使用复杂的代码更好。有点像我以前坚持的一个观点
--[在C中使用goto来进行错误处理](https://eli.thegreenplace.net/2009/04/27/using-goto-for-error-handling-in-c) 
Go指南很明确，并且对恢复的限制非常巧妙 - 即使用于解析器中的控制流程，也比经典异常更难以滥用。
有趣的是，当这个问题首先引起我的注意时，我正在研究json / encode.go包的源代码。 事实证明，它最近被修复使用经典的错误处理！
是的，一些代码变得更加冗长，从这样：
```go
if destring {
  switch qv := d.valueQuoted().(type) {
    case nil:
      d.literalStore(nullLiteral, subv, false)
    case string:
      d.literalStore([]byte(qv), subv, true)
    // ... other code
```
变成了这样:
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
但总的来说，它并不是那么糟糕，对于Go coder来说肯定不会陌生。 它给了我希望:-)

【1】:C ++的异常安全保证集是所涉及的一些复杂性的一个很好的例子。

【2】: 如果你花一些时间去读一读[提出recover机制的邮件](https://groups.google.com/forum/#!topic/golang-nuts/HOXNBQu5c-Q%5B26-50%5D)
你会发现Russ Cox在解析二进制流时会提到类似的问题，以及如何在整个过程中传播错误。

【3】:甚至C ++也有类似的模式，你可以在一些使用标准返回类型的代码库中找到它。 通常名为ASSIGN_OR_RETURN
的宏在Google发布的C ++代码中很流行，并且出现在LLVM等其他地方。


---

via: https://eli.thegreenplace.net/2018/on-the-uses-and-misuses-of-panics-in-go/

作者：[Eli Bendersky](https://eli.thegreenplace.net/pages/about)
译者：[zouxinjiang](https://github.com/zouxinjiang)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出