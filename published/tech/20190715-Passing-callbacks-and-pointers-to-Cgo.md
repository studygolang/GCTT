首发于：https://studygolang.com/articles/24447

# 传递回调函数和指针到 Cgo

`Cgo`允许 Go 程序调用 C 库或其他暴露了 C 接口的库。正是如此，这也成为 Go 程序员工具箱的重要组成部分。

使用`Cgo`可能会比较棘手，特别是在 Go 和 C 代码中传递指针和回调函数时。
这篇文章讨论了一个端到端当例子，包含了如下几方面：
* `Cgo`的基本使用，包括链接一个传统的 C 库到 Go 二进制文件中。
* 从 Go 语言中传递 struct 到 C 语言中。
* 传递 Go 函数到 C 程序中，并安排 C 程序在随后调用它们。
* 安全的传递任意的 Go 数据到 C 代码中，这些 C 代码后续会回传这些数据到它所调用的 Go 回调中。

本文并不是一个`Cgo`的使用教程-在阅读前，需要你对它对简单使用案例有所熟悉。
在本文最后列了一些有用的`Cgo`使用教程和相关的文章。这个案例的全部源代码详见[Github](https://github.com/eliben/code-for-blog/tree/master/2019/cgo-callback)。

## 问题所在-一个C库调用多个Go回调程序

如下是一个虚构的C库的头文件，该库处理（输入）数据，并基于事件调用回调函数。

```c
typedef void (*StartCallbackFn)(void* user_data, int i);
typedef void (*EndCallbackFn)(void* user_data, int a, int b);

typedef struct {
  StartCallbackFn start;
  EndCallbackFn end;
} Callbacks;
// Processes the file and invokes callbacks from cbs on events found in the
// file, each with its own relevant data. user_data is passed through to the
// callbacks.
void traverse(char* filename, Callbacks cbs, void* user_data);
```

回调标签是由几个重要的模式组成，所展示的这些模式在现实中也同样普遍:

* 每一个回调拥有自己的类型签名，这里为了简便，我们使用`int`类型的参数，这个参数可以是其他任何类型。
* 当只有较小数量的回调被调用时，它们可能作为独立的参数被传递到 traverse 中；然而，回调的数量非常大时（比如说，超过三个），后几乎总是有一个汇集它们的结构体被传递。
允许用户将某些回调参数设置为 null 很常见，以向底层库传达：对于某些特定事件并没有意义，也不应为此调用任何用户代码。
* 每个回调都获得一个不透明指针 user_data，该指针从调用者传递到 traverse （最终传递到回调函数）。它用于区分互不相同的遍历，并传递用户特定的状态。
典型的，traverse 会透传 user_data，而不尝试访问他； 由于它是`void *`，因此它对于库是完全模糊的，
并且用户代码会将其强制转换为回调中的某些具体类型。

我们对 traverse 的实现仅是一个简单的模拟：

```c
void traverse(char* filename, Callbacks cbs, void* user_data) {
  // 模拟某些遍历，调用 start 回调，之后调用 end 回调
  // callback, if they are defined.
  if (cbs.start != NULL) {
    cbs.start(user_data, 100);
  }
  if (cbs.end != NULL) {
    cbs.end(user_data, 2, 3);
  }
}
```

我们的任务是包装这个库，在 Go 代码中进行使用。我们想要在遍历中调用 Go 回调，不用再写任何多余的 C 代码。

## Go 接口

让我们从构思在 Go 代码中我们接口的样式开始，如下是一个方式：

```go
type Visitor interface {
  Start(int)
  End(int, int)
}
func GoTraverse(filename string, v Visitor) {
  // ... 实现
}
````

本文后续部分显示了使用此方法的完整实现。但是，它有一些缺点：
* 当我们需要提供大量的回调时，如果我们仅对几个回调感兴趣，编写 Visitor 的实现可能会很乏味。
可以通过提供一个结构体来实现带有某些默认操作（例如，无操作）的完整接口来减轻这种情况，然后用户结构可以匿名继承此默认结构，而不必实现每个方法。
尽管如此，带有大量方法的接口通常不是一个好的 Go 实践。
* 一个更严重的限制是，很难向 C 遍历传达我们对某些回调不感兴趣的信息。
根据定义，实现 Visitor 的对象将具有所有方法的实现，因此没有简单的方法来判断我们是否对调用其中的某些方法不感兴趣。
这可能会对性能产生严重影响。

一个可替换的方法是模仿我们在 C 语言中拥有的方式；也就是说，创建一个整合函数对象的结构体：

```go
type GoStartCallback func(int)
type GoEndCallback func(int, int)

type GoCallbacks struct {
  startCb GoStartCallback
  endCb   GoEndCallback
}
func GoTraverse(filename string, cbs *GoCallbacks) {
  // ... 实现
}
```

这立即解决了两个缺点：函数对象的默认值为`nil`，GoTraverse 可以将其解释为“对此事件不感兴趣”，其中可以将相应的 C 回调设置为`NULL`。
由于 Go 函数对象可以是闭包或绑定方法，因此在不同的回调之间保留状态没有困难。

后附的代码示例在单独的目录中提供了此替代实现，但是在其余文章中，我们将继续使用 Go 接口的更惯用的方法。
对于实现而言，选择哪种方法并不重要。

## Cgo 包装函数的实现

`Cgo`指针传递规则不允许将 Go 函数值直接传递给 C，因此要注册回调，我们需要在 C 中创建包装器函数。

而且，我们也不能直接传递 Go 程序分配的指针到 C 程序中，因为 Go 的并发垃圾回收器会移动数据。
`Cgo`的[Wiki](https://github.com/golang/go/wiki/cgo#function-variables)提供了使用间接寻址的解决方法。
在这里，我将使用 go-pointer 程序包，该程序包以稍微更方便，更通用的方式实现了相同目的。

考虑到这些，让我们之间进行实现。该代码初步看起来可能会比较晦涩，但这很快就会展现出他的意义。如下是 GoTraverse 的代码。

```go
import gopointer "github.com/mattn/go-pointer"

func GoTraverse(filename string, v Visitor) {
  cCallbacks := C.Callbacks{}

  cCallbacks.start = C.StartCallbackFn(C.startCgo)
  cCallbacks.end = C.EndCallbackFn(C.endCgo)

  var cfilename *C.char = C.CString(filename)
  defer C.free(unsafe.Pointer(cfilename))

  p := gopointer.Save(v)
  defer gopointer.Unref(p)

  C.traverse(cfilename, cCallbacks, p)
}
```

我们先在 Go 代码中创建 C 的回调结构，然后封装。因为我们不能直接将 Go 函数赋值给 C 函数指针，我们将在独立的 Go 文件[注1]中定义这些包装函数。

```c
/*
extern void goStart(void*, int);
extern void goEnd(void*, int, int);

void startCgo(void* user_data, int i) {
  goStart(user_data, i);
}

void endCgo(void* user_data, int a, int b) {
  goEnd(user_data, a, b);
}
*/
import "C"
```

这些是非常轻量的、调用 go 函数的包装器——我们不得不为每一类的回调写这样一个 C 函数。我们很快就会看到 Go 函数 goStart 和 goEnd。
在填充这个 C 回调结构体后，GoTraverse 会将文件名从 Go 字符串转换为 C 字符串（`Wiki`中有详细信息）。
之后，它创建一个代表 Go 访问者的值，我们可以使用 go-pointer 包将其传递给 C。最后，它调用 traverse。

完成这个实现，goStart 和 goEnd 代码如下：

```go
//export goStart
func goStart(user_data unsafe.Pointer, i C.int) {
  v := gopointer.Restore(user_data).(Visitor)
  v.Start(int(i))
}

//export goEnd
func goEnd(user_data unsafe.Pointer, a C.int, b C.int) {
  v := gopointer.Restore(user_data).(Visitor)
  v.End(int(a), int(b))
}
```

导出指令意味着这些功能对于 C 代码是可见的。 它们的签名应具有 C 类型或可转换为 C 类型的类型。 它们的行为类似：
1. 从 user_data 解压缩访问者对象
2. 在访问者上调用适当的方法

## 详细的调用流程

让我们研究一下“开始”事件的回调调用流程，以更好地了解各个部分是如何连接在一起的。
GoTraverse 将 startCgo 赋值给 Callbacks 结构体中的 start 指针，Callbacks 结构体将被传递给 traverse。因此，traverse 遇到 start 事件时，它将调用 startCgo。
回调的参数包括：传递给 traverse 的 user_data 指针以及事件特定的参数（该例中为一个 int 类型的参数）。

startCgo 是 goStart 的填充程序，并使用相同的参数调用它。

goStart 解压缩由 GoTraverse 打包到 user_data 中的 Visitor 实现，并从那里调用 Start 方法，并向其传递事件特定的参数。
到这一点为止，所有代码都由 Go 库包装 traverse 提供；从这里开始，我们进入由`API`用户编写的自定义代码。

## 通过C代码传递Go指针

此实现的另一个关键细节是我们用于将 Visitor 封装在`void * user_data`内在 C 回调来回传递的的技巧。

[Cgo文档](https://golang.org/cmd/cgo/#hdr-Passing_pointers)指出：

> 如果 Go 代码指向的 Go 内存不包含任何 Go 指针，则 Go 代码可以将 Go 指针传递给 C。

但是，我们当然不能保证任意的 Go 对象不包含任何指针。除了明显使用指针外，函数值，切片，字符串，接口和许多其他对象还包含隐式指针。

限制源于 Go 垃圾收集器的性质，该垃圾收集器与其他代码同时运行，并允许移动数据，从 C 角度来看，会使指针无效。

所以，我们能做些什么？如上所述，解决方案是间接的，Cgo`Wiki`提供了一个简单的示例。
我们没有直接将指针传递给 C ，而是将其保留在 Go 板块中，并找到了一种间接引用它的方法；
例如，我们可以使用一些数字索引。这保证了所有指针对于 Go 的垃圾回收仍然可见，但是我们可以在 C 板块中保留一些唯一的标识符，以便以后我们访问它们。

通过在 unsafe.Pointer（映射到 Cgo 对 C 的调用中直接`void *`）和`interface{}`之间创建一个映射，go-pointer 包便可以做到这一点，从本质上讲，我们可以存储任意的 Go 数据并提供唯一的 ID（unsafe.Pointer）以供后续引用。
为什么不像`Wiki`示例中那样使用 unsafe.Pointer 代替`int`？因为不明确的数据通常在 C 语言中用`void *`表示，所以不安全。指针是自然映射到它的东西。如果使用`int`，我们将不得不去考虑在其他几个地方进行转换。

## 如果没有 user_data 呢？

看到我们如何使用 user_data， 使其穿越 C 代码回到我们的回调函数，以传输特定于用户 Visitor 的实现，人们可能会想-如果没有可用的 user_data 怎么办？
事实证明，在大多数情况下，都存在诸如 user_data 之类的东西，因为没有它，原始的 C `API`就有缺陷。再次考虑遍历示例，但是这项没有 user_data：

```c
typedef void (*StartCallbackFn)(int i);
typedef void (*EndCallbackFn)(int a, int b);

typedef struct {
  StartCallbackFn start;
  EndCallbackFn end;
} Callbacks;

void traverse(char* filename, Callbacks cbs);
```

假设我们提供一个回调作为开始：

```c
void myStart(int i) {
    // ...
}
```

在 myStart 中，我们有些困惑了。我们不知道调用哪个遍历-可能有许多不同的遍历，不同的文件和数据结构满足不同的需求。我们也不知道在哪里记录事件的结果。这里唯一的办法是使用全局数据。这是一个不好的`API`！

有了这样的`API`，我们在 Go 板块的情况就不会差很多。我们还可以依靠全局数据来查找与此特定遍历有关的信息，并且我们可以使用相同的 Go 指针技巧在此全局数据中存储任意 Go 对象。但是，这种情况不太可能出现，因为 C `API`不太可能忽略此关键细节。

## 附属资源链接

关于使用`Cgo`的信息还有很多，其中有些是过时的（在明确定义传递指针的规则之前）。如下是我在准备这篇文章时发现的特别有用的链接集合：

* [官方的 Cgo 文档](https://golang.org/cmd/cgo/)是权威指南。
* [Cgo 的 Wiki 页面](https://github.com/golang/go/wiki/cgo)是相当有用的。
* [Go 语言并发垃圾回收的一些细节](https://blog.golang.org/go15gc)。
* Yasuhiro Matsumoto的[从C调用Go](https://dev.to/mattn/call-go-function-from-c-function-1n3)的文章。
* 指针传递规则的[详细细节](https://github.com/golang/proposal/blob/master/design/12416-cgo-pointers.md)。

[注1]由于 Cgo 生成和编译 C 代码的特殊性，它们位于单独的文件中-有关[Wiki](https://github.com/golang/go/wiki/cgo#export-and-definition-in-preamble)的更多详细信息。我没有对这些函数使用静态内联技巧的原因是我们必须获取它们的地址。

----------------

via: https://eli.thegreenplace.net/2019/passing-callbacks-and-pointers-to-cgo/

作者：[Eli Bendersky](https://eli.thegreenplace.net)
译者：[amzking](https://github.com/amzking)
校对：[DingdingZhou](https://github.com/DingdingZhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
