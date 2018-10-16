已发布：https://studygolang.com/articles/13019

# Cgo 和 Python

如果你研究过 [新近的 Datadog Agent](https://github.com/DataDog/datadog-agent/) ，你可能会注意到大部分代码库是用 Go 语言编写，而我们用来收集指标的检查工具仍然是用的 Python 。这是有可能的，因为  Datadog Agent（一个标准的 Go 二进制程序），内嵌了一个 CPython 解释器，在需要运行 Python 代码的时候就会调用这个解释器。通过使用一个抽象层可以让整个过程是透明的，由此，你尽可以编写惯用的 Go 代码，即使底层同时运行着 Python 代码。

想要在 Go 应用中嵌入 Python 的原因有很多种：

- 对接口有益处；逐步把现有 Python 项目的一部分迁移到新的语言，在整个过程中不丢失功能。
- 你可以重用现有的 Python 软件或库，不必用新的语言重新实现它们。
- 通过加载和运行标准的 Python 脚本你可以动态扩展你的软件，即使是在运行时。

这个清单还可以继续下去，但对 Datadog Agent 来说最后一点最为关键： 我们希望你能够执行自定义的检查和更改而不必去重新编译 Agent ，或者更广泛点儿，编译任何一处。

内嵌的 CPython 非常简单而且文档也很完善。解释器本身使用 C 编写，并且提供了一个 C API 使得可以用编程方式执行一些比较底层的操作，例如创建对象，导入模块，以及调用函数。

在本文中我们会介绍一些简单的代码示例，并且在与 Python 交互的时候，我们会着力于保证 Go 代码仍符合语言习惯。但在开始之前我们需要先标记一处小障碍：内嵌的 API 是 C 但我们的主应用却是 Go ，这又是怎么能完成的呢？

![Introducing cgo](https://raw.githubusercontent.com/studygolang/gctt-images/master/cgo-python/cgo_python_divider_1.png)

## Cgo 简介

可能会有 [很多原因](https://dave.cheney.net/2016/01/18/cgo-is-not-go) 让你不希望引入 Cgo  到你的工作栈中，但内嵌的 CPython 却是一个能让你同意的关键砝码。Cgo 既不是语言也不是编译器。它是 [外部函数接口](https://en.wikipedia.org/wiki/Foreign_function_interface)（FFI），一种能让我们在 Go 中调用其他语言编写的函数和服务，尤其是 C 。

当我们提到 “Cgo” 时，我们实际指的是底层 Go 工具链使用的一系列工具，库，函数，以及类型，所以我们依然可以用  `go build` 获取 Go 的二进制程序。一个使用 Cgo 的极简代码示例如下所示：

 ```go
package main
// #include <float.h>
import "C"
import "fmt"

func main() {
	fmt.Println("Max float value of float is", C.FLT_MAX)
}
 ```

在 `import "C"` 上方的注释块可以预先调用，并且能包含实际的 C 代码，在这个例子里包含了一个头文件。一旦被导入，“C” 伪库就会让程序跳转到外部代码，访问 `FLT_MAX` 宏。你可以通过调用 `go build` 来 build 该示例，就跟普通的 Go 代码一样。

如果你想要了解一下底层 Cgo 所做的全部工作，就运行 `go build -x` 。你将会看到 “Cgo” 工具被调用去生成一些 C 和 Go 的模块，然后 C 和 Go 编译器会被调用以建立目标模块，最终链接器会把一切都安排好。

你可以在 [Go 博客](https://blog.golang.org/c-go-cgo) 中阅读更多 Cgo 的内容。这篇文章包含更多示例，以及一些深入细节的有用链接。

既然我们已经知道了 Cgo 可以帮助我们做哪些工作，接下来就来看看我们如何借助该机制来运行 Python 代码。

![Embedding CPython](https://raw.githubusercontent.com/studygolang/gctt-images/master/cgo-python/cgo_python_divider_2.png)

## 内嵌 CPython : 入门

一个 Go 程序，严格来说，内嵌 CPython 并不像你预计的那样复杂。事实上，在最低限度上，我们要做的所有工作就是在运行 Python 代码之前初始化解释器，以及在运行结束后回收资源。请注意，我们将会在所有的示例中都使用 Python 2.x ，但是这些都能够适用于 Python 3.x，仅仅需要很少的修改。让我们先看一个例子：

```go
package main
// #cgo pkg-config: python-2.7
// #include <Python.h>
import "C"
import "fmt"
func main() {
	C.Py_Initialize()
	fmt.Println(C.GoString(C.Py_GetVersion()))
	C.Py_Finalize()
}
```

上面的例子和下面的 Python 代码等价：

```python
import sys
print(sys.version)
```

你可以看到我们放了一个 `#cgo` 在前面；这些符号将会被传递给工具链，从而改编 build 的工作流。在这个例子，我们让 Cgo 去调用 “pkg-config” 来获取 build 需要的标志，并且链接到一个叫 “python-2.7” 的库，以及传递这些标志给 C 编译器。如果你的系统上安装了  CPython 开发库并用 pkg-config 连接，这将使得你可以继续使用普通的 `go build`  编译上面的例子。

重新再看代码，我们使用 `Py_Initialize()` 和`Py_Finalize()` 来开启和关闭解释器，以及 C 函数 `Py_GetVersion`  来获取包含内嵌解释器版本信息的字符串。

如果你有疑问，所有我们需要整合来调用 C Python API 的 Cgo 部分都是样板代码。这也是  Datadog Agent 依赖 [go-python](https://github.com/sbinet/go-python) 来执行所有内嵌操作的原因所在； go-python 库提供了 Go 风格的对 C API 的简单封装，并且隐藏了 Cgo 的细节。这是另一个简单的嵌入示例，这次使用 go-python：

```go
package main
import (
	python "github.com/sbinet/go-python"
)
func main() {
	python.Initialize()
	python.PyRun_SimpleString("print 'hello, world!'")
	python.Finalize()
}
```

这个例子更接近标准 Go 代码，没有暴漏出更多的 Cgo ， 而且我们可以在访问 Python API 的时候来回使用 Go 字符串。内嵌看起来功能强大并且对开发者友好。是时候好好使用解释器了：让我们尝试加载一个磁盘上的 Python 模块。

我们没必要使用复杂的 Python 模块，一个最简单的 “hello world” 就可以满足需求：

```python
# foo.py
def hello():
		""" Print hello world for fun and profit. """
		print "hello, world!"
```

Go 代码稍微复杂点，但依然容易阅读：

```go
// main.go
package main
import "github.com/sbinet/go-python"
func main() {
	python.Initialize()
	defer python.Finalize()
	fooModule := python.PyImport_ImportModule("foo")
	if fooModule == nil {
		panic("Error importing module")
	}
	helloFunc := fooModule.GetAttrString("hello")
	if helloFunc == nil {
		panic("Error importing function")
	}
	// The Python function takes no params but when using the C api
	// we're required to send (empty) *args and **kwargs anyways.
	helloFunc.Call(python.PyTuple_New(0), python.PyDict_New())
}
```

完成之后，我们需要设置 `PYTHONPATH` 环境变量到当前工作目录，由此 import 语句就能够找到

`foo.py` 模块。在 shell 中，命令类似下面：

```shell
$ go build main.go && PYTHONPATH=. ./main hello, world!
```

![The dreadful Global Interpreter Lock](https://raw.githubusercontent.com/studygolang/gctt-images/master/cgo-python/cgo_python_divider_3.png)

## 糟糕的全局解释器锁（ GIL ）

为了内嵌 Python 引入 Cgo 是一个妥协：build 过程会变慢，垃圾回收器不会帮我们管理外部系统使用的内存，并且交叉编译也会有难度。是否为一个特定项目引入可能会引发争论，但有一点我认为是无需讨论的： Go 并发模型。如果我们不能在一个 goroutine 里面运行 Python，这一切就毫无意义。

在用 Python 和 cgo 实现并发之前 ，有一点我们需要了解：就是全局解释器锁，简称 GIL 。GIL 是一个被语言解释器（ CPython 只是一种）广泛采用的机制，目的是防止同一时刻有超过一个以上的线程运行。这意味着被 CPython 执行的 Python 程序不可能在同一个进程中并行。并发倒是仍然有可能，锁是在速度，安全性和实现难度上的一个较好权衡。那么它为什么会在内嵌的时候引出问题？

当一个标准的，非内嵌 Python 程序启动时，不会有 GIL 卷进来，从而避免了锁操作带来的无用的间接损耗；GIL  第一次启动时一些 Python 代码会请求开启一个线程。对任何一个线程来说，解释器创建一个数据结构来储存当前状态信息和锁住 GIL 。当线程结束后，状态会被恢复而且 GIL 也会解锁，从而可以被其它线程使用。

我们在 Go 程序中运行 Python 的时候，以上这些都不会自动发生。没有 GIL，我们的 Go 程序可能会创建多个 Python 线程。这将有可能引起竞争条件而导致致命的运行时错误，而且一个模块的错误极有可能摧毁整个 Go 应用。

解决方案就是在任何时候运行 Go 中的多线程代码都要显式调用 GIL；代码不会太复杂，因为 C API 提供了我们需要的所有工具。为了更好的暴漏问题，我们需要在Python中做一些CPU绑定的事情。让我们把这些函数添加到前面示例中的 foo.py 中：

```python
import sys
def print_odds(limit=10):
	""" Print odds numbers < limit """
	for i in range(limit):
		if i%2:
			sys.stderr.write("{}\n".format(i))
def print_even(limit=10):
	""" Print even numbers < limit """
	for i in range(limit):
		if i%2 == 0:
			sys.stderr.write("{}\n".format(i))
```

我们会在 Go 中尝试并发的打印奇数和事件编号，使用两个不同的 goroutine （由此引入线程）：

```go
package main
import ( "sync"
				"github.com/sbinet/go-python"
			 )
func main() {
	//  下面代码会通过调用PyEval_InitThreads()显式调用 GIL ，
	//  无需等待解释器去执行 python.Initialize()
	var wg sync.WaitGroup
	wg.Add(2)
	fooModule := python.PyImport_ImportModule("foo")
	odds := fooModule.GetAttrString("print_odds")
	even := fooModule.GetAttrString("print_even")
	// Initialize() 已经锁定 GIL ，但这时我们并不需要它。
	// 我们保存当前状态和释放锁，从而让 goroutine 能获取它
	state := python.PyEval_SaveThread()
	go func() {
		_gstate := python.PyGILState_Ensure()
		odds.Call(python.PyTuple_New(0), python.PyDict_New())
		python.PyGILState_Release(_gstate)
		wg.Done()
	}()
	go func() {
		_gstate := python.PyGILState_Ensure()
		even.Call(python.PyTuple_New(0), python.PyDict_New())
		python.PyGILState_Release(_gstate)
		wg.Done()
	}()
	wg.Wait()
	// 在这里我们知道程序不会再需要运行 Python 代码了，
	// 我们可以恢复状态和 GIL 锁，执行退出前的最后操作。
	python.PyEval_RestoreThread(state)
	python.Finalize()
}
```

在阅读示例的时候你可能注意到了一个模式，这个模式将会是我们运行内嵌 Python 的准则

1. 保存状态并锁住 GIL。
2. 执行 Python 。
3. 恢复状态，解锁 GIL。

代码可以说得上简洁明了，但仍有一处细节需要指出：注意，即使是遵循 GIL 模式，在一个例子里面我们运行 GIL 时是通过调用`PyEval_SaveThread()`  和  `PyEval_RestoreThread()` ,在另一个例子里（请看 goroutines 里面的代码）我们是通过调用  `PyGILState_Ensure()` 和 `PyGILState_Release()` 。

我们说过，当 Python 里面运行多线程时，解释器会负责创建储存当前状态的数据结构，但如果是在 C API 里面的话，需要我们亲自动手实现。

当我们通过 go-python 初始化解释器的时候，我们是运行在 Python 上下文环境。所以当调用 `PyEval_InitThreads()`  时解释器会初始化数据结构并锁住 GIL 。我们可以使用`PyEval_SaveThread()` 和`PyEval_RestoreThread()`  对已经存在的状态进行操作。

在 goroutine 中，我们则是在一个 Go 上下文环境中运行，并且我们不需要显式的创建和删除状态， `PyGILState_Ensure()` 和 `PyGILState_Release()` 负责完成这些工作。

![Unleash the Gopher](https://raw.githubusercontent.com/studygolang/gctt-images/master/cgo-python/cgo_python_divider_4.png)

## 解放 Go 爱好者

现在我们已经知道怎么处理多线程 Go 代码在一个内嵌解释器中执行 Python 的过程了，但是在 GIL 之后，我们又面临着一个新的挑战：Go 调度器。

当一个 goroutine 启动时，它会被调度运行在 `GOMAXPROCS` 个可用线程中的其中一个线程——[点击这里](https://morsmachine.dk/go-scheduler)  可以了解更多细节。当一个 goroutine 执行系统调用或者调用 C 代码时，当前线程会把等待运行线程队列中的其它 goroutine 移交给另一个线程，从而让这些 goroutine 有更多机会执行；当前 goroutine 被挂起，直到系统调用或是 C 函数返回。如果有返回发生，线程就会试图唤醒被终止的 goroutine ，但如果没有返回的可能性，那该线程就会请求 Go 运行时去查找另一个线程来完成该 goroutine ，并且进入睡眠状态。 goroutine 最终被调度给另一个线程，然后结束。

考虑到这些，让我们来看看当一个正在运行 Python 代码的 goroutine 被移动到一个新的线程时， goroutine 都会发生什么：

1. 我们的 goroutine 启动后，执行一个 C 函数调用，然后挂起。GIL 被锁住。
2. 当 C 函数调用返回，当前线程试图唤醒该 goroutine，但它失败了。
3. 当前线程告诉 Go 运行时去查找另一个线程来唤醒我们的 goroutine。
4. Go 调度器找到一个可用的线程，并且 goroutine 也被唤醒。
5. goroutine 基本完成，并且尝试在返回前解锁 GIL。
6. 当前状态存储的线程 ID 是初始线程的ID，和当前线程的 ID 不一致。
7. Panic ！

幸运的是，我们可用强制要求 Go 运行时保证我们的 goroutine 一直运行在同一个线程上，只要通过 goroutine 调用 runtime 包里的 LockOSThread 函数就行。

```go
go func() {
	runtime.LockOSThread()
	_gstate := python.PyGILState_Ensure()
	odds.Call(python.PyTuple_New(0), python.PyDict_New())
	python.PyGILState_Release(_gstate)
	wg.Done()
}()
```

这将会干扰调度器并可能带来一些间接损耗，但为了避免随机的 panic 我们愿意付出这种代价。

## 结论

顾及到内嵌 Python ， Datadog  Agent 做出了以下取舍：

- cgo 引入的间接损耗。
- 手动操作 GIL。
- 运行期间绑定 goroutine 到同一个线程的限制。

考虑到在 Go 中运行 Python 检查的便利，我们很乐意接受这一切。但既然意识到了这些取舍，我们就能够最小化它们带来的影响。对于为支持 Python 而带来的其它限制，我们很难有对策处理可能的问题：

- build 依然是自动化的并且可配置的，因此开发者照样会类似于 `go build` 。
- 一个轻量级版本的 agent 可以被创建，而且剥离 Python 支持后也完全可以支持简单使用 Go build 标记。
- 这样的版本只依赖于 agent 本身的硬编码核心检查（绝大部分是系统和网络检查），但不受 cgo 限制而且可以被交叉编译。

我们会重新评估将来的选择，并判断是否值得继续使用 cgo ；我们甚至可能会考虑把 Python 作为一个整体是否值得，等到  [Go 插件包](https://golang.org/pkg/plugin/) 足够成熟，可以支持我们的用例。但至少现在内嵌 Python 工作的很好，而且从旧 Agent 迁移到新的上面也很简单。

你是一个掌握并且喜欢混合不同语言编程的爱好者吗？你热爱学习语言的内部工作机制来保证你的代码更加健壮吗？ [请加入 Datadog](https://www.datadoghq.com/careers/ ) !

---

via: https://www.datadoghq.com/blog/engineering/cgo-and-python/

作者：[Massimiliano Pippi](https://www.datadoghq.com/blog/engineering/cgo-and-python/)
译者：[sunzhaohao](https://github.com/sunzhaohao)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
