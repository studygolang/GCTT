# 传递指针和回调函数到Cgo

`Cgo`允许Go程序调用C库或其他暴露了C接口的库。正是如此，这也成为 Go 程序员工具箱的重要组成部分。

使用`Cgo`可能会比较棘手，特别是在 Go 和 C 代码中传递指针和回调函数时。
这篇文章讨论了一个端到端当例子，包含了如下几方面：
* `Cgo`的基本使用，包括链接一个传统的C库到Go二进制文件中。
* 从Go语言中传递struct到C语言中。
* 传递Go函数到C程序中，并安排C程序在随后调用它们。
* 安全到传递任意的Go数据到C代码中，这些C代码后续会回传这些数据到它所调用的Go回调中。

本文并不是一个`Cgo`的使用教程-在阅读前，需要你对它对简单使用案例有所熟悉。
在本文最后列了一些有用的`Cgo`使用教程和相关的文章。这个案例的全部源代码详见[Github](https://github.com/eliben/code-for-blog/tree/master/2019/cgo-callback)。

## 问题所在-一个C库调用多个Go回调程序

如下是一个虚构的C库的头文件，该库是通过一些数据和基于事件的回调来运行。

````cgo
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
````

回调标签是由几个重要的模式组成，所展示的这些模式在现实中也同样普遍:

* 每一个回调拥有自己的类型签名，这里为了简便，我们使用`int`类型的参数，这个参数可以是其他任何类型。
* 当只有较小数量的回调被调用时，它们可能作为独立的参数被传递到traverse中；然而，回调到数量会非常大（比如说，超过三个），
然后几乎总是有一个汇集它们的结构体被传递，然后几乎总是有一个收集它们的结构被传递。
通常，允许用户将某些回调设置为NULL传递到库中，这些特定事件对这种回调不感兴趣，并且不应为此调用任何用户代码。
* 每个回调通过对traverse的调用都获得一个模糊的user_data指针。它用于区分互不相同的遍历，并传递用户特定的状态。
遍历通常会传递user_data甚至不尝试访问它； 由于它是`void *`，因此它对于库是完全模糊的，
并且用户代码会将其强制转换为回调中的某些具体类型。

我们对traverse的实现仅是一个简单的模拟：

````cgo
void traverse(char* filename, Callbacks cbs, void* user_data) {
  // 模拟某些遍历，调用start回调，之后调用end回调
  // callback, if they are defined.
  if (cbs.start != NULL) {
    cbs.start(user_data, 100);
  }
  if (cbs.end != NULL) {
    cbs.end(user_data, 2, 3);
  }
}
````

我们的任务是包装这个库，在Go代码中进行使用。我们想要在遍历中调用Go回调，不用再写任何多余的C代码。

## Go 接口

让我们从构思在Go代码中我们接口的样式开始，如下是一个方式：

````
type Visitor interface {
  Start(int)
  End(int, int)
}
func GoTraverse(filename string, v Visitor) {
  // ... 实现
}
````

文章的其余部分显示了使用此方法的完整实现。但是，它有一些缺点：
* 当我们需要提供大量的回调时，如果我们仅对几个回调感兴趣，编写Visitor的实现可能会很乏味。
可以通过提供一种结构来实现带有某些默认值（例如，无操作）的完整接口来减轻这种情况，然后用户结构可以嵌入此默认结构，而不必实现每个方法。
尽管如此，使用许多方法进行接口通常不是一个好的Go实践。
* 一个更严重的限制是，很难向C遍历传达我们对某些回调不感兴趣的信息。
根据定义，实现Visitor的对象将具有所有方法的实现，因此没有简单的方法来判断我们是否对调用其中的某些方法不感兴趣。
这可能会对性能产生严重影响。

一个可替换的方法是模仿我们在C语言中拥有的东西。 也就是说，创建一个结构收集函数对象：

````
type GoStartCallback func(int)
type GoEndCallback func(int, int)

type GoCallbacks struct {
  startCb GoStartCallback
  endCb   GoEndCallback
}
func GoTraverse(filename string, cbs *GoCallbacks) {
  // ... 实现
}
````

这立即解决了两个缺点：函数对象的默认值为`nil`，GoTraverse可以将其解释为“对此事件不感兴趣”，其中可以将相应的C回调设置为`NULL`。
由于Go函数对象可以是闭包或绑定方法，因此在不同的回调之间保留状态没有困难。

后附的代码示例在单独的目录中提供了此替代实现，但是在其余文章中，我们将继续使用Go接口的更惯用的方法。
对于实现而言，选择哪种方法并不重要。

## Cgo 包装函数的实现

`Cgo`指针传递规则不允许将Go函数值直接传递给C，因此要注册回调，我们需要在C中创建包装器函数。

而且，我们也不能直接传递Go程序分配的指针到C程序中，因为Go到并发垃圾回收器会移动数据。
`Cgo`的[Wiki](https://github.com/golang/go/wiki/cgo#function-variables)提供了使用间接寻址的解决方法。
在这里，我将使用go-pointer程序包，该程序包以稍微更方便，更通用的方式实现了相同目的。

考虑到这些，让我们之间进行实现。该代码初步看起来可能会比较晦涩，但这很快就会展现出他的意义。如下是GoTraverse的代码。

````
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
````

我们先在Go代码中创建C的回调结构，然后封装。因为我们不能直接将Go函数赋值给C函数指针，我们将在独立的Go文件[注1]中定义这些包装函数。

````
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
````

这些是调用Go函数的非常薄的包装函数——我们不得不为每一类的回调写这样一个C函数。我们很快就会看到Go函数goStart和goEnd。
在封装这个C回调结构体后，GoTraverse会将文件名从Go字符串转换为C字符串（`Wiki`中有详细信息）。
之后，它创建一个代表Go访问者的值，我们可以使用go-pointer包将其传递给C。最后，它调用遍历。

完成这个实现，goStart和goEnd代码如下：

````
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
````

导出指令意味着这些功能对于C代码是可见的。 它们的签名应具有C类型或可转换为C类型的类型。 它们的行为类似：
1. 从user_data解压缩访问者对象
2. 在访问者上调用适当的方法

## 详细的调用流程

让我们研究一下“开始”事件的回调调用流程，以更好地了解各个部分是如何连接在一起的。
GoTraverse将startCgo分配给传递给traverse的Callbacks结构中的start指针。因此，traverse遇到启动事件时，它将调用startCgo。
参数是传递给traverse的user_data指针以及事件特定的参数（在这种情况下为单个`int`）。

startCgo是goStart的填充程序，并使用相同的参数调用它。

goStart解压缩由GoTraverse打包到user_data中的Visitor实现，并从那里调用Start方法，并向其传递事件特定的参数。
到这一点为止，所有代码都由Go库包装traverse提供；从这里开始，我们进入由API用户编写的自定义代码。

## 通过C代码传递Go指针

此实现的另一个关键细节是我们用于将Visitor封装在`void * user_data`内在C回调来回传递的的技巧。

[Cgo文档](https://golang.org/cmd/cgo/#hdr-Passing_pointers)指出：

> 如果Go代码指向的Go内存不包含任何Go指针，则Go代码可以将Go指针传递给C。

但是，我们当然不能保证任意的Go对象不包含任何指针。除了明显使用指针外，函数值，切片，字符串，接口和许多其他对象还包含隐式指针。

限制源于Go垃圾收集器的性质，该垃圾收集器与其他代码同时运行，并允许移动数据，从C角度来看，会使指针无效。

所以，我们能做些什么？如上所述，解决方案是间接的，Cgo`Wiki`提供了一个简单的示例。
我们没有直接将指针传递给C，而是将其保留在Go板块中，并找到了一种间接引用它的方法；
例如，我们可以使用一些数字索引。这保证了所有指针对于Go的垃圾回收仍然可见，但是我们可以在C板块中保留一些唯一的标识符，以便以后我们访问它们。

通过在unsafe.Pointer（映射到Cgo对C的调用中直接void *）和interface{}之间创建一个映射，go-pointer包便可以做到这一点，从本质上讲，我们可以存储任意的Go数据并提供唯一的ID（unsafe.Pointer）以供后续引用。
为什么不像`Wiki`示例中那样使用unsafe.Pointer代替int？因为不明确的数据通常在C语言中用`void *`表示，所以不安全。指针是自然映射到它的东西。如果使用int，我们将不得不去考虑在其他几个地方进行转换。

## 如果没有user_data呢？

看到我们如何使用user_data通过C代码将特定于用户的Visitor实现通过隧道传送回我们的通用回调，人们可能会想-如果没有可用的user_data怎么办？
事实证明，在大多数情况下，都存在诸如user_data之类的东西，因为没有它，原始的C `API`就有缺陷。再次考虑遍历示例，但是这项没有user_data：

````
typedef void (*StartCallbackFn)(int i);
typedef void (*EndCallbackFn)(int a, int b);

typedef struct {
  StartCallbackFn start;
  EndCallbackFn end;
} Callbacks;

void traverse(char* filename, Callbacks cbs);
````

假设我们提供一个回调作为开始：

````
void myStart(int i) {
    // ...
}
````

在myStart中，我们有些困惑了。我们不知道调用哪个遍历-可能有许多不同的遍历，不同的文件和数据结构满足不同的需求。我们也不知道在哪里记录事件的结果。这里唯一的办法是使用全局数据。这是一个不好的`API`！

有了这样的`API`，我们在Go板块的情况就不会差很多。我们还可以依靠全局数据来查找与此特定遍历有关的信息，并且我们可以使用相同的Go指针技巧在此全局数据中存储任意Go对象。但是，这种情况不太可能出现，因为C `API`不太可能忽略此关键细节。

## 附属资源链接

关于使用`Cgo`的信息还有很多，其中有些是过时的（在明确定义传递指针的规则之前）。如下是我在准备这篇文章时发现的特别有用的链接集合：

* [官方的Cgo文档](https://golang.org/cmd/cgo/)是权威指南。
* [Cgo的Wiki页面](https://github.com/golang/go/wiki/cgo)是相当有用的。
* [Go语言并发垃圾回收的一些细节](https://blog.golang.org/go15gc)。
* Yasuhiro Matsumoto的[从C调用Go](https://dev.to/mattn/call-go-function-from-c-function-1n3)的文章。
* 指针传递规则的[详细细节](https://github.com/golang/proposal/blob/master/design/12416-cgo-pointers.md)。

[注1]由于Cgo生成和编译C代码的特殊性，它们位于单独的文件中-有关[Wiki](https://github.com/golang/go/wiki/cgo#export-and-definition-in-preamble)的更多详细信息。我没有对这些函数使用静态内联技巧的原因是我们必须获取它们的地址。

----------------

via: https://eli.thegreenplace.net/2019/passing-callbacks-and-pointers-to-cgo/

作者：[Eli Bendersky](https://eli.thegreenplace.net)
译者：[amzking](https://github.com/amzking)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
