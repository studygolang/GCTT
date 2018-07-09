首发于：https://studygolang.com/articles/13514

# 关于结构化并发的笔记 —— Go 语言中有害的声明语句

## 前言

每一个并发的 API 背后的代码都需要允许并发运行的，以下是使用不同 API 的例子：

```
go myfunc();                                // Golang

pthread_create(&thread_id, NULL, &myfunc);  /* C with POSIX threads */

spawn(modulename, myfuncname, [])           % Erlang

threading.Thread(target=myfunc).start()     # Python with threads

asyncio.create_task(myfunc())               # Python with asyncio
```

在不同的符号和术语中有许多变体，但语义是一样的。上面的例子是以并发的方式运行 `myfunc` 以便使用程序空闲的资源，并且调用后立马回到父进程（或主线程）去做其他事情。

另一种方式是使用回调:

```
QObject::connect(&emitter, SIGNAL(event()),        // C++ with Qt
                 &receiver, SLOT(myfunc()))

g_signal_connect(emitter, "event", myfunc, NULL)   /* C with GObject */

document.getElementById("myid").onclick = myfunc;  // Javascript

promise.then(myfunc, errorhandler)                 // Javascript with Promises

deferred.addCallback(myfunc)                       # Python with Twisted

future.add_done_callback(myfunc)                   # Python with asyncio
```

再说，虽然语法不一样，但是它们都完成同样的事情：它们安排好任务（arrange)，之后，直到某一事件发生了，myfunc 就会运行。注册“事件回调”成功，以上的函数就立即返回，调用者可以继续做其他事情。（有时候回调可以被巧妙地封装成 helper，例如 [promise](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/all) [combinators](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/race)，或者 [Twisted-style protocols/transports](https://twistedmatrix.com/documents/current/core/howto/servers.html) ，但核心思路是一样的）

还有其他方式吗？你使用的任何的现实上的并发的 API ，你可能会发现他们都是殊途同归的，例如 python 的 asyncio 。

但我的新开发的库 [Trio](https://trio.readthedocs.io/) 与众不同，它没有有使用其他方式。相反，如果我们希望并发执行 `myfunc` 或者其他函数，我们可以这样写：

```python
async with trio.open_nursery() as nursery:
    nursery.start_soon(myfunc)
    nursery.start_soon(anotherfunc)

```

当人们遇到 `nursery` 的构建方法时，他们想了解它的神秘之处，为什么有一个缩进？为什么我需要一个 `nursery` 对象来派发异步任务，人们开始感觉这有违以往的使用习惯，感到厌烦，这个库让人感到怪异，远远脱离原始的语法，这些都是可以理解的反应，但是请谅解我。

在这篇文章，我希望说服你 `nurseries` 并不总是怪异特殊的，而是一个新的控制流原语，它与循环或函数调用一样重要。此外，我们在上面看到的其他方法 - 线程派发和回调注册 - 应该被完全移除并替换为 `nurseries` 。

听起来难以接受？历史发生过类似的事： 语句 `goto` 曾一度认为是控制流中的王者，现在依旧沦落为 [过去式](https://xkcd.com/292/) 。一些语言依旧有类似 `goto` 的语句，相比原来的 `goto` 有所不同或者被弱化。大部分语言甚至没有它。发生了什么？时间太长久以至于人们忘记了过去的故事，但是结果令人惊讶的类似。所以我们首先会提醒自己什么是goto，然后看看，关于并发API，它可以教给我们的东西。

## 大纲目录

- 到底什么是 `goto`
- 到底什么是 `go`
- `goto` 发生了什么
    1. `goto` 抽象的毁灭者
    2. 一个惊喜，移除 `goto` 语句带来新的特性
    3. `goto` 语句：不再使用

- `go` 语句被认为有害
    1. go 语句：不再使用

- Nurseries: 一个替代 `go` 语句的构件
    1. Nurseries 支持函数抽象
    2. Nurseries 支持动态任务派发
    3. 那有一个逃逸
    4. 你可以定义一个像 nursery 的新类型
    5. 不，事实上，nurseries 总是等待里面的任务退出
    6. 自动清理资源的工作
    7. 自动传递错误信息的工作
    8. 一个令人惊讶的好处：移除 `go` 语句开启一个新的特性

- 实践 Nurseries
- 结论
- 致谢
- 脚注

## `goto` 到底是什么

让我们回顾历史：早期的计算机是使用汇编语言编程的，或者其他更原始的机器机制。有些简陋。因此，在20世纪50年代，IBM 的 John Backus 和 Remington Rand 的 Grace Hopper 等人开始开发 FORTRAN 和 FLOW-MATIC 等语言（以其直接后继 COBOL 而闻名）。

FLOW-MATIC当时非常雄心勃勃。你可以把它看作是Python的曾曾曾祖父母：第一种语言，首先为人类设计，其次是计算机。以下是一些FLOW-MATIC代码，让您体验它的外观：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/flow-matic-code-0.png)

你注意到它不像现代语言，没有 `if` 代码段，`loop` 循环语句，或者 function 函数调用，实际上根本没有块分隔符或缩进。这只是一个简单的语句（操作符）列表。并不是因为这个程序太短而无法使用更高级的控制语法，而是因为”块语法“还没有发明出来！

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/compare-sequential-to-goto.png)

反而， FLOW-MATIC 有两种控制流的方式，比较通常的是顺序（sequential），如你所想，从头到尾一句一句地、串行地执行语句，但如果你执行一句特殊的语句 `JUMP TO`，它会立马直接跳转到其他控制语句。例如，下图的语句13跳转到语句2：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/flow-matic-code-1.png)

就好像一开始的并发原语，对于如何称呼这个“进行跳转动作”的操作，有一些意见分歧。这里是 `JUMP TO` ，但是这里叫做 `goto` (就好像 "go to")，于是我们在这里使用这个称呼。

这个小程序使用的完整的 `goto` 跳转路径:

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/flow-matic-code-2.png)

如果你感到这看起来很费解，你不是一个人！这个跳转风格的程序是 `FLOW-MATIC` 非常直接地继承汇编语言而来。这很功能强大，非常适合计算机硬件的实际工作方式，但直接使用会让人感到非常困惑。乱七八糟的箭头让人发明了“意大利面代码”这个词。显然，我们需要更好的东西。

但是... goto导致所有这些问题的关键是什么？为什么有的控制结构好，有些不是？我们如何选择好的？这时，这个问题真的很不清楚，如果你不了解问题，很难解决问题。

## `go` 到底是什么

但让我们暂停回顾历史 - 每个人都知道 `goto` 是不好的。这与并发性有什么关系？那么，考虑Go语言的出名的 `go` 语句，用于产生一个新的 `goroutine` （轻量级线程）：

```golang
// Golang
go myfunc();
```

我们可以绘制其控制流程图吗？这不同于我们上面看到，因为控制流实际上是分裂的。我们可以这样画：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/go-myfunc.png)

这里的颜色被分成两条路径。从主线（绿线）的角度来看，控制流按顺序走：它出现在顶部，然后立即出现在底部。与此同时，从支线（淡紫色线）的角度看，控制流进入顶部，然后跳转到myfunc的主体。与常规函数调用不同，这种跳转是单向的：当运行myfunc时，我们切换到一个全新的堆栈，并且运行时立即忘记了我们来自哪里。

但这不仅适用于Golang。这是我们在本文开头列出的所有基元的流程控制图：

- 线程库通常提供某种类型的句柄对象，让您稍后可以加入线程 - 但这是一种语言不知道的独立操作。实际的线程产生原语具有上面显示的控制流程。
- 注册回调在语义上等同于启动一个后台线程，该后台线程（a）阻塞，直到发生某个事件，然后（b）运行回调。（虽然显然实现是不同的）因此，就高级别控制流而言，注册回调本质上是一种 go 语句。
- `future` 和 `promise` 也是一样的：当你调用一个函数并且它返回一个 promise 时，这意味着它计划在后台发生的工作，然后给你一个句柄对象以后加入工作（如果你想的话）。就控制流语义而言，这就像产生一个线程一样。然后你在这个 promise 上注册回调，所以看到前面的要点。

同样的确切模式以许多形式出现：关键的相似之处在于，在所有这些情况下，控制流分离，一方进行单向跳转，另一方返回给调用者。一旦你知道要寻找什么，你就会开始在各地看到它 - 这是一个有趣的游戏！[1]

令人烦恼的是，这类控制流结构没有标准名称。因此，就像 `goto语句` 成为所有不同类似 goto 结构的总称一样，我将使用 `go语句` 作为这些术语的总称。为什么这么做？ 一个原因是 Go 语言给了我们一个特别纯粹的形式例子。另一个是......好吧，你可能已经猜到了我说什么了。看看这两个图。注意任何相似之处：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/compare-go-to-goto.png.png)

没错：go语句是goto语句的一种形式。

并发程序的编写和推理是非常困难的。基于goto的程序也是如此。这可能是出于某些相同的原因吗？在现代语言中，goto引起的问题在很大程度上得到解决。如果我们从　”研究他们如何修正“　转到　”它会教我们如何制作更多可用的并发API“，让我们来找出答案。

## goto发生了什么事？

那么，为什么 goto 会导致如此多的问题？在二十世纪六十年代后期，Edsger W. Dijkstra 编写了两篇很著名的论文，帮助人们更清楚地了解这一点：[goto 语句被认为有害](https://scholar.google.com/scholar?cluster=15335993203437612903&hl=en&as_sdt=0,5) 与 [结构化编程的笔记](https://www.cs.utexas.edu/~EWD/ewd02xx/EWD249.PDF)

### goto:抽象的毁灭者

在这些论文中，Dijkstra担心如何编写非凡的软件并使其正确。我无法在这里详细说明。例如，你可能听说过这句话：

> 程序测试可以用来发现 bug 的存在，但却没法证明 bug 不存在

是的，这来自[结构化编程的笔记](https://www.cs.utexas.edu/~EWD/ewd02xx/EWD249.PDF)。但他主要关心的是抽象。他想写一些太大的项目，不能一下子把想到的都写下来。要做到这一点，您需要将程序的某些部分像黑盒子一样对待 - 就像当您看到 Python 程序时一样：

```python
print("Hello world!")
```

那么你不需要知道打印是如何实现的（字符串格式化，缓冲，跨平台差异......）的所有细节。您只需要知道它会以某种方式打印您提供的文本，然后您可以花费精力去考虑您的代码中是否希望在此时发生这种情况。 Dijkstra希望语言支持这种抽象。

至此，块语法已经被发明出来，像ALGOL这样的语言已经积累了5种不同类型的控制结构：它们仍然具有顺序流和 `goto` ：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/compare-sequential-to-goto.png)

并且还获得了if / else，循环和函数调用的变体：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/if-loop-functioncall.png)

你可以使用goto来实现这些更高层次的结构，并且在早期，人们就会想到它们：作为一个方便的简写。但是Dijkstra指出的是，如果你看这些图表，goto和其他人之间有很大的区别。对于除goto以外的所有内容，流量控制位于顶部→[有问题]→流量控制位于底部。我们可以称之为“黑盒子规则”：如果一个控制结构具有这种形状，那么在你不关心内部发生的细节的上下文中，你可以忽略[stuff happen]部分，并把整个作为规则的顺序流程。而且更好的是，任何由这些部分组成的代码也是如此。当我看这个代码时：

```python
print("Hello world!")
```

我不必去阅读 `print` 及其所有传递依赖的定义，只是想知道控制流程是如何工作的。也许在内部 `print` 有一个循环，并且在循环内部有一个 `if / else` ，并且在 `if / else` 内部还有另一个函数调用......或者也可能是其他内容。它并不重要：我知道控制将流入 `print` ，函数将完成它的事情，然后最终控制将返回到我正在阅读的代码。

看起来这很明显，但如果你有一个带有 `goto` 的语言 —— 一种语言，其功能和其他所有内容都建立在 `goto` 之上，`goto` 可以随时随地跳转 - 然后这些控制结构根本不是黑匣子！如果你有一个函数，并且在函数内部有一个循环，并且在循环内部有一个`if/else`，并且在 `if/else` 中有一个 `goto` ...那么 `goto` 可以将控制发送到任何它想要的地方。也许控制会突然从另一个你还没有调用的函数完全返回，你不知道！

这就打破了抽象：这意味着每一个函数调用都可能是一个变相的 `goto` 语句，唯一需要知道的就是将系统的整个源代码一次性保存在头脑中。只要 `goto` 使用你的语言，你就会停止对流量控制进行本地推理。这就是为什么 `goto` 会导致意大利面代码。

现在 Dijkstra 明白了这个问题，他能够解决这个问题。这是他的革命性建议：我们应该停止将 `if / loops /function call` 作为 `goto` 的简写，而应该将它们作为自己权利的基本原语 - 并且我们应该完全从我们的语言中删除 `goto`。

从2018年起，这似乎显而易见。但是当你试图拿走他们的玩具时，你有没有看过程序员的反应，因为他们不够聪明，无法安全使用它们？是的，有些事情永远不会改变。 1969年，这个提议令人难以置信地引起争议。[Donald Knuth](https://en.wikipedia.org/wiki/Donald_Knuth) 为 `goto` [辩护](https://scholar.google.com/scholar?cluster=17147143327681396418&hl=en&as_sdt=0,5) 。曾经成为编写代码专家的人非常不满，他们基本上不得不基本学会如何重新编程，以便使用更新，更有约束的构造来表达自己的想法。当然，它需要建立一套全新的语言。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/goto-change.png)

最后，现代语言比 Dijkstra 的原始公式稍逊一筹。他们会让你使用 `break`，`continue` 或 `return` 等构造立即跳出多个嵌套结构。但基本上，他们都是围绕 Dijkstra 的想法而设计的；即使这些推动边界的构造只能以严格有限的方式进行。特别是，`function` - 这是在黑匣子内部封装控制流程的基本工具 - 被认为是不可侵犯的。你不能从一个函数中跳出另一个函数，并且返回可以将你从当前函数中取出，但不能再进一步。无论控制流程如何，一个函数在内部起作用，其他函数不必关心。

这甚至延伸到 goto 本身。你会发现几种语言仍然有一些他们称之为 goto 的语言，比如 C，C＃，Golang ......但是它们增加了很多限制。至少，他们不会让你跳出一个 function 并跳入另一个 function 。除非你在汇编[2]中工作，无限制的goto不见了。 Dijkstra赢了。

### 一个惊喜，移除 `goto` 语句带来新的特性

一旦 goto 消失，就会发生一些有趣的事情：语言设计者能够开始添加依赖于控制流程结构的功能。

例如，Python有一些很好的资源清理语法： with 语句。你可以写下如下内容：

```python
# Python
with open("my-file") as file_handle:
    ...
```

并保证该文件将在 `...` 代码期间打开，但随后会立即关闭。大多数现代语言都有一些等效的（RAII，使用，试用资源，推迟 ......）。他们都假设控制流程是有序的，结构化的。如果我们使用goto语句跳入我们中间有块 `...` 你会怎么办？文件是否打开？如果我们再次跳出来，而不是正常退出？文件会关闭吗？此功能只是没有任何连贯的方式工作。

错误处理有类似的问题：当出现问题时，你的代码应该做什么？答案常常是把堆栈中的堆栈传递给代码的调用者，让他们弄清楚如何处理它。现代语言具有专门的构造来使这更容易，例如异常或其他形式的[自动错误传播](https://doc.rust-lang.org/std/result/index.html#the-question-mark-operator-)。但是你的语言只能提供这种帮助，如果它有一个堆栈和一个可靠的“呼叫者”概念。再看一下我们FLOW-MATIC程序中的控制流面条，并想象在它的中间试图引发异常。它甚至会去哪里？

### goto 语句：不再使用

所以 `goto` —— 忽略函数界限的传统类型 —— 不仅仅是一种常见的坏特性，难以正确使用。如果仅仅如此，它可能会保留下来。但实际情况更糟。

即使你不使用 goto ，只要把它作为你的语言的一个选项，就会让所有的东西都难以使用。每当你开始使用第三方库时，你都不能把它当作一个黑匣子 - 你必须仔细阅读它以找出哪些函数是常规函数，哪些函数是伪装的特殊控制结构流。这是本地推断的严重障碍。你失去了强大的语言功能，如可靠的资源清理和自动错误传递。我们应该更好地完全移除goto，以支持遵循“黑匣子”规则的控制流构造。

### go 语句被认为有害

所以这就是goto的历史。现在，这多少适用于 `go`语句？　那么......基本上，所有这一切！这个比喻结果令人震惊。

Go语句打破了抽象。请记住我们如何说如果我们的语言允许跳转，那么任何功能都可能是变相跳转？在大多数并发框架中，go语句会导致完全相同的问题：无论何时调用函数，它都可能会或可能不会产生一些后台任务。该功能似乎回来了，但它仍然在后台运行？如果没有阅读所有的源代码，就没有办法知道。何时完成？很难说。如果你有 go 语句，然后功能不再相对于黑盒子来控制流量。在我的第一篇关于并发API的文章中，我称之为“违反因果关系”，并发现它是使用 asyncio 和 Twisted 的程序中的许多常见现实问题的根源，例如背压问题，正确关闭问题等等。

Go语句会中断自动资源清理。让我们再一次`with` 语句示例：

```python
# Python
with open("my-file") as file_handle:
    ...
```

之前，我们说我们是保证（guaranteed）的同时，该文件将被打开、代码运行，然后完成关闭。但是，是否有东西可以让代码生成一个后台任务？那么我们的保证就失去了：该操作看起来 `with` 块的内部操作会在 `with` 块结束之后被回收，因为文件被同时他们还在使用它关闭。再次，你不能从当地的检查中看出来; 要知道，如果发生这种情况，你必须去阅读源代码，看内部实现功能的 `...` 代码。

如果我们希望此代码正常工作，我们需要以某种方式跟踪任何后台任务，并且只有在完成后手动安排文件才能关闭。这是可行的 - 除非我们正在使用一些不提供任何方式在任务完成时得到通知的库，这是非常常见的（例如因为它没有公开任何可以加入的任务句柄）。但即使在最好的情况下，非结构化的控制流程也意味着语言无法帮助我们。我们现在回到实施资源清理手中，就像在过去的糟糕时期。

Go语句打破错误处理。就像我们上面讨论的那样，现代语言提供了强大的工具，例如异常，以帮助我们确保检测到错误并将其传播到正确的位置。但是这些工具依赖于拥有“当前代码的调用者”的可靠概念。只要您产生任务或注册回调，该概念就会被破坏。结果，我所知道的每个主流并发框架都简单地放弃了。如果后台任务发生错误，而您没有手动处理它，那么运行时只是......将它放到地板上并不再管它，这不是太重要。如果幸运的话，它可能会在控制台上打印一些东西。（我唯一使用过的认为“打印某些内容并继续前进”的软件是一个很好的错误处理策略，它是古老的Fortran库，但我们在这里。）甚至 Rust - 这门语言在高中阶段被投票选为“正确度最高的人”。如果后台线程发生混乱，Rust [抛弃错误并期望变得更好](https://doc.rust-lang.org/std/thread/) 。

当然，您可以在这些系统中正确处理错误，仔细确保加入每个线程，或者通过构建自己的错误传播机制，如 [Javascript中的Twisted](https://twistedmatrix.com/documents/current/core/howto/defer.html#visual-explanation) 或 [Promise.catch中的errbacks](https://hackernoon.com/promises-and-error-handling-4a11af37cb0e) 。但是现在你正在编写一个特殊的，脆弱的重新实现你的语言已经有的功能。你失去了诸如“回溯”和“调试器”等有用的东西。所需要的只是忘记拨打 `Promise.catch` 一次，突然间，您甚至没有意识到地板上的严重错误。即使你以某种方式解决了所有这些问题，你仍然会得到两个冗余系统来做同样的事情。

### go 语句：不再使用

就像goto是第一批实用的高级语言的明显原始代码一样，go是第一个实用并发框架的明显原语：它匹配底层调度程序实际工作的方式，并且它足够强大，可以实现任何其他并发流模式。但是，再次像goto一样，它打破了控制流抽象，所以只是将它作为您的语言的一个选项使得一切都变得更难以使用。

好消息是，这些问题都可以解决，Dijkstra 向我们展示如何做到：

- 找到具有类似功能的语句的替代品，但遵循“黑匣子规则”，
- 将这个新构造作为原语构建到我们的并发框架中，并且不包含任何形式的go语句。

这就是 Trio 所做的。

### Nurseries: 一个替代 `go` 语句的构件

以下是核心思想：每次我们的控制分裂成多个并发路径时，我们都要确保他们再次归纳起来。例如，如果我们想同时做三件事情，我们的控制流程应该如下所示：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/control-flow.png)

注意，这只有一个箭头出现在顶部，一个出现在底部，所以它遵循 Dijkstra 的黑盒子规则。现在，我们怎样才能把这个草图变成一个具体的语言结构呢？有一些现有的构造可以满足这个约束，但是（a）我的提议与我所知道的并且比它们有优势（特别是在想要使其成为独立原语的情况下）略有不同，并且（b）并发性是庞大而复杂的，试图把所有的历史和权衡分开将会使论证完全失败，所以我将把它推迟到另一篇单独的文章。在这里，我只关注解释我的解决方案。但请注意，我并不是说自己喜欢，发明了并发或某种东西，我是站在巨人的肩膀上，从很多来源吸取灵感。

无论如何，下面是我们要做的事情：首先，我们声明父任务不能启动任何子任务，除非它首先为子任务创建一个地方： Nurseries 。它通过打开一个 Nurseries 块来实现这一点; 在Trio中，我们使用 Python 的 `async with` 语法来执行此操作：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/python-async-with.png)

打开一个nursery块会自动创建一个代表这个 Nurseries 的对象，并且 nursery 语法将这个对象赋给名为 nursery 的变量。然后我们可以使用 nursery 对象的 start_soon 方法来启动并发任务：在这种情况下，一个任务调用函数 myfunc，另一个调用函数 anotherfunc。从概念上讲，这些任务在 Nurseries 区内执行。实际上，将 nursery 块内写入的代码视为创建块时自动启动的初始任务通常很方便。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/python-nursery.png)

最重要的是，在 Nurseries 区块内的所有任务都已经退出之前， Nurseries 区块不会退出 - 如果父任务在所有子任务完成之前到达该区块的末尾，那么它会在那里暂停并等待它们。 Nurseries 自动扩展以容纳子任务。

以下是控制流程：您可以看到它与我们在本节开头展示的基本模式的匹配情况：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-statement-considered-harmful/python-nursery-control-flow.png)

这种设计有许多后果，并非全部都是显而易见的。我们来看看其中的一些。

### Nurseries 支持函数抽象

go 语句的基本问题是，当你调用一个函数时，你不知道它是否会产生一些后台任务，在完成后继续运行。使用 Nurseries ，您不必担心这一点：任何函数都可以打开 Nurseries 并运行多个并发任务，但函数只有在完成后才能返回。所以当一个函数返回时，你知道它确实完成了。

### Nurseries 支持动态的任务派发

这是一个更简单的原型，它也满足我们上面的流程控制图。它需要一个thunk的列表，并且同时运行它们：

```python
run_concurrently([myfunc, anotherfunc])
```

但问题在于你必须知道你将要运行的任务的完整列表，而这并非总是如此。例如，服务器程序通常具有接受循环，它接收传入的连接并开始一个新的任务来处理它们中的每一个。这是Trio中最小的接受循环：

```python
async with trio.open_nursery() as nursery:
    while True:
        incoming_connection = await server_socket.accept()
        nursery.start_soon(connection_handler, incoming_connection)
```

有了 Nurseries ，这是微不足道的，但使用实现它 run_concurrently 会比较尴尬。如果你愿意，可以很容易地在 nurseries 之上实现 run_concurrently - 但这并不是必须的，因为在 run_concurrently 可以处理的简单情况下 ，nursery符号就像可读的一样。

### 有一个逃逸

 Nurseries 对象也给我们一个逃逸的出口。如果你确实需要编写一个产生背景任务的函数，那么背景任务会超出函数本身呢？这也很容易：通过功能一个 Nurseries 对象。直接在open_nursery（）块内异步的代码调用 nursery.start_soon - 只要nursery块保持打开 [4]，那么任何获得对该 Nurseries 对象的引用的人都可以获得派发任务的能力进入那个 Nurseries 。你可以将它作为函数参数传递，通过队列发送。

在实践中，这意味着你可以编写“违反规则”的函数，但在以下限制内：

- 由于 Nurseries 对象必须明确地通过，您可以立即通过查看其呼叫站点来识别哪些功能违反正常的控制流，因此本地推理仍然是可能的。
- 函数产生的任何任务仍然受到传入的 Nurseries 的生命周期的约束。
- 调用代码只能通过它自己有权访问的 Nurseries 对象。

所以这与传统模式仍然有很大的不同，任何时候任何代码都可以在无限的生命周期内产生背景任务。

有一个地方很有用就是证明 Nurseries 具有相当的表达能力去发表声明，但是这篇文章已经足够长了，所以我会再留到下一篇。

### 你可以定义一个像 nursery 的新类型

标准的 Nurseries 语义提供了一个坚实的基础，但有时候你想要一些不同的东西。也许你会羡慕 Erlang ，并且想要定义一个类似 Nurseries 的类，通过重新启动子任务来处理异常。这完全有可能，对于你的用户来说，它看起来就像一个普通的 Nurseries ：

```python
async with my_supervisor_library.open_supervisor() as nursery_alike:
    nursery_alike.start_soon(...)
```

如果您有一个将 nursery 作为参数的函数，那么您可以将其中的一个传递给它，以便为其生成的任务控制错误处理策略。相当漂亮。但是这里有一个微妙之处，它推动 Trio 走向不同的约定，而不是 asyncio 或其他库：这意味着 start_soon 必须采用一个函数，而不是协程对象或未来。（你可以多次调用一个函数，但是没有办法重启一个协程对象或 Future。）我认为这是一个更好的约定，无论如何由于多种原因（特别是因为 Trio 甚至没有 Future！），但仍值得一提。

### 不，实际上， Nurseries 总是等待里面的任务退出。

关于任务取消和任务加入如何相互作用也值得讨论，因为这里有一些微妙之处 - 如果处理不正确 - 就打破了 Nurseries 不变式。

在Trio中，代码可能随时收到取消请求。请求取消之后，下一次代码执行“检查点”操作（[详细信息](https://trio.readthedocs.io/en/latest/reference-core.html#checkpoints)）时，会引发取消异常。这意味着请求取消与实际 发生时间之间存在差距- 任务执行检查点之前可能需要一段时间，然后异常必须展开堆栈，运行清理处理程序等。发生时， Nurseries 总是等待全面清理。我们永远不会终止任务，也不会让它有机会运行清理处理程序，而我们永远不会 即使在 Nurseries 正在取消的过程中，也可以让任务在 Nurseries 外无人监管。

### 自动清理资源的工作

由于 Nurseries 按照黑匣子规则，它让 `with` 块重新工作。禁止这样的行为：代码块结束时依旧有后台任务在执行。

### 自动传递错误信息的工作

如上所述，在大多数并发系统中，后台任务中未处理的错误只是被丢弃，来不及做别的事。

在 Trio 中，由于每项任务都在 Nurseries 内，每个 Nurseries 都是父任务的一部分，父任务需要等待 Nurseries 内的任务......我们确实有一些事情可以通过未处理的错误来完成。如果后台任务以异常终止，我们可以在父任务中重新抛出它。这里的直觉是，一个 Nurseries 就像一个“并发呼叫”原语：我们可以将上面的例子看作是同时调用 myfunc 和 anotherfunc ，所以我们的调用栈已经变成了一棵树。异常将这个调用树传播给根，就像它们传播一个普通的调用栈一样。

这里有一个细微之处：当我们在父任务中重新引发异常时，它将开始在父任务中传播。通常，这意味着父任务将退出 Nurseries 区块。但是我们已经说过，在任务仍在运行的情况下，父任务不能离开 Nurseries 区块。那么我们该怎么办？

答案是，当子任务发生未处理的异常时，Trio 立即取消同一 Nurseries 中的所有其他任务，然后等待它们完成后再重新引发异常。这异常直接导致导致堆栈释放，如果我们想要释放我们的堆栈树中的一个分支点，我们需要展开其他分支，取消它们。

这确实意味着如果你想用你的语言来实施 Nurseries，你可能需要在 Nurseries 代码和你的取消系统之间进行某种整合。如果你使用像 C＃ 或 Golang 这样的语言，这可能会非常棘手，通常通过手动对象传递和约定来管理取消，或者（更糟的是）没有通用的取消机制的取消。

### 一个令人惊讶的好处：移除 `go` 语句开启一个新的特性

消除 goto 使得以前的语言设计师能够对程序结构做出更强的假设，从而实现了块和例外等新功能; 消除 go 语句也有类似的效果。例如：

- Trio 的注销系统（cancellation system）比竞争对手更容易使用和更可靠，因为它可以假定任务嵌套在常规树形结构中; 请参阅 [完超时和人为取消](https://vorpus.org/blog/timeouts-and-cancellation-for-humans/) 。
- Trio 是唯一的 Python 并发库，其中 control-C 以 Python 开发人员期望的方式工作（[详细信息](https://vorpus.org/blog/control-c-handling-in-python-and-trio/)）。Nurseries 提供可靠的机制来处理异常。

## 实践 Nurseries

这就是理论。它在实践中如何工作？

这是一个实践问题：你应该尝试一下并找出答案！但是，严重的是，我们经历过问题才明白过来。在这一点上，我非常确信基础是健全的，但是也许我们会意识到我们需要做一些调整，比如早期的结构化编程倡导者最终如何从消除 `break` 和 `continue` 中得到回应。

如果你是一位经验丰富的并发程序员，他们只是学习Trio，那么你应该预料到这需要习惯它。你将不得不学习新的方法来做事情 - 就像在20世纪70年代的程序员一样，学习如何在没有 `goto` 下编写代码是一个挑战。

但当然，这是关键。正如Knuth所写的（Knuth，1974，p.275）：

> 也许最糟糕的错误任何一个可以相对于标的做出去报表是假设“结构化编程”是通过编写程序来实现，因为我们总是有，然后消除去的。大部分去的不应该在那里！我们真正需要的是这样一种方式，我们很少甚至设想我们的计划想约去 陈述，因为他们真正需要的几乎没有出现。我们表达思想的语言对我们的思维过程有着强烈的影响。因此，迪克斯特拉要求更多新的语言特征 - 鼓励清晰思考的结构 - 以避免这样做对并发症的诱惑。

到目前为止，这是我使用 Nurseries 的经验：它鼓励清晰的思维。它导致设计更加健壮，更易于使用，而且更好。而这些限制实际上使解决问题变得更容易，因为您花费更少的时间去尝试不必要的复杂问题。在一个非常真实的意义上，使用 Trio 已经教会我成为一个更好的程序员。

例如，考虑 Happy Eyeballs 算法（[RFC 8305](https://tools.ietf.org/html/rfc8305)），这是一种简单的并发算法，用于加快建立TCP连接。从概念上来说，算法并不复杂 - 您尝试了多次连接尝试，并且为了避免网络过载而采用错开的方式。但是如果你看看[Twisted的最佳实现](https://github.com/twisted/twisted/compare/trunk...glyph:statemachine-hostnameendpoint)，它几乎有600行Python，并且仍然[至少有一个逻辑错误](https://twistedmatrix.com/trac/ticket/9345)。 比起 Trio 的同类型项目缩短了15倍以上。更重要的是，使用Trio，我可以在几分钟内写出它，而不是几个月，而且我在第一次尝试时就得到了正确的逻辑。我从来不可能在其他任何框架中做到这一点，即使我有更多的经验。有关更多详细信息，您可以观看 [我上个月在Pyninsula的演讲](https://www.youtube.com/watch?v=i-R704I8ySE)。这只是个例吗？时间会证明一切，它充满着希望。

## 结论

流行的并发原语 - go 语句， thread spawning functions, callbacks, futures, promises 等等，它们都是 goto 的变体，理论上和实践上都是如此。即使是现代化的驯化过的 `goto`，但旧的 `goto` 是可以跳出函数边界的。即使我们不直接使用它们，这些原语也是危险的，因为它们破坏了我们推理控制流的能力，并且从抽象的模块化部分组成复杂的系统，并干扰了自动资源清理和错误传播等有用的语言功能。因此，像 goto 一样，他们在现代高级语言中没有地位。

 Nurseries 提供了一种安全方便的替代方案，保留了您语言的全部功能，实现了强大的新功能（如 Trio 的取消范围和控制 C 处理所证明的），并且可以显着提高可读性，生产力和正确性。

不幸的是，要完全享受这些好处，我们需要完全删除旧的基元，这可能需要从头开始构建新的并发框架 - 就像消除 设计新语言所需的 goto 一样。但与 FLOW-MATIC 相比，它的表现令人印象深刻，我们大多数人都很高兴我们已经升级到了更好的产品。我不认为我们会后悔切换到 Nurseries，Trio 也证明这是一个实用的通用并发框架的可行设计。

## 致谢

非常感谢 Graydon Hoare，Quentin Pradet 和 Hynek Schlawack 对本文的草稿提出的意见。任何剩余的错误当然都是我的错。

## 参考文献

- FLOW-MATIC 示例代码来自 [this brochure(pdf)](http://archive.computerhistory.org/resources/text/Remington_Rand/Univac.Flowmatic.1957.102646140.pdf) 存储在 [Computer History Museum](http://www.computerhistory.org/collections/catalog/102646140)

- [Wolves in Action](https://www.flickr.com/photos/iam_photo/478178221) by i:am. photography / Martin Pannier, licensed under [CC-BY-SA 2.0](https://creativecommons.org/licenses/by-nc-sa/2.0/)

- [French Bulldog Pet Dog](https://pixabay.com/en/french-bulldog-pet-dog-funny-2427629/) by Daniel Borker, 登载在 [CC0 public domain dedication](https://creativecommons.org/publicdomain/zero/1.0/)

---

via: https://vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/

作者：[Nathaniel J. Smith](https://vorpus.org)
译者：[lightfish-zhang](https://github.com/lightfish-zhang)
校对：[polaris1119](https://github.com/polaris1119)，[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
