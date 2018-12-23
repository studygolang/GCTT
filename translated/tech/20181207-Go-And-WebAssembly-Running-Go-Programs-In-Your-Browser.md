# Go和WebAssembly:在浏览器中运行Go程序

很长一段时间以来，Javascript 是在 web 开发者中的通用语言。如果你想写出一个稳定、成熟的 web 应用程序，那么使用 Javascript 几乎是唯一的方法。

WebAssembly（也叫 wasm ）即将改变这种情况。使用 WebAssembly，现在可以用*任何*语言来编写 web 应用程序。在这篇文章当中，我们将明白怎样编写 Go 程序并使用 wasm 在浏览器中运行它们。

## 首先，什么是WebAssembly

WebAssembly 官方网站 [webassembly.org](https://webassembly.org/) 对它的定义是“一个基于堆栈的二进制指令格式的虚拟机”。这是一个很好的定义，但是让我们来将它分解为我们能够简单理解的内容。

从本质上讲，wasm 是一种二进制格式，就像 ELF、Mach 和 PE 一样。唯一的区别是它适用于虚拟编译目标，而不是真正的物理机。为什么是虚拟机？因为与 C/C++ 的二进制不一样，wasm二进制不针对于特定平台。因此，你可以不用改变任何东西而在 Linux、Windows 和 Mac 上使用同一份二进制文件。但是，我们需要额外的“代理”，它将 wasm 指令中的二进制文件转换为特定平台的指令并运行它们。通常，这个“代理”就是一个 web 浏览器，但是理论上讲，它可以是任何其它东西。

这为我们提供了一个通用的编译目标，我们可以使用自己选择的任何编程语言来构建 web 应用程序。只要我们将程序编译为wasm 格式，我们就不用担心目标平台。就像我们编写一个 web 应用程序，但现在我们可以使用我们选择的任何语言编写它。

## Hello WASM

让我们尝试从编写一个简单的“hello world”程序开始。确保你的 Go 版本至少为 1.11。我们可以这样写：

```go
package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello wasm")
}
```

保存为一个 `test.go` 文件。这看上去就是一个常规的 Go 程序。现在让我们来编译这个文件到 wasm 平台。我们需要像下面这样设置 `GOOS` 和 `GOARCH` 来编译它。

```
$GOOS=js GOARCH=wasm go build -o test.wasm test.go
```

我们现在就生成了 wasm 二进制文件。但是与本机系统不同，我们需要在浏览器中运行它。为此，我们需要投入一些额外的东西来实现这一个目标：

* 一个将为我们 web 应用程序提供服务的 webserver
* 一个 index.html 文件，包含一些用于加载 wasm 二进制文件的js代码
* 一个 js 文件，用于作为浏览器和我们的 wasm 二进制文件之间的通信接口

我喜欢把它想象成制作飞天小女警所需要的东西。

![wasmrequirements](https://blog.gopheracademy.com/postimages/advent-2018/go-in-the-browser/powerpuff.jpg)

然后 **BOOM**，我们就有了一个 WebAssembly 应用程序！

我们已经在 Go 发行版本中提供了 html 和 js 文件，在此我们将它们复制下来。

```c
$cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
$cp "$(go env GOROOT)/misc/wasm/wasm_exec.html" .
$# we rename the html file to index.html for convenience.
$mv wasm_exec.html index.html
$ls -l
total 8960
-rw-r--r-- 1 agniva agniva    1258 Dec  6 12:16 index.html
-rwxrwxr-x 1 agniva agniva 6721905 Sep 24 12:28 serve
-rw-rw-r-- 1 agniva agniva      76 Dec  6 12:08 test.go
-rwxrwxr-x 1 agniva agniva 2425246 Dec  6 12:09 test.wasm
-rw-r--r-- 1 agniva agniva   11905 Dec  6 12:16 wasm_exec.js
```

`serve` 是一个简单的 Go 二进制文件，它为当前目录中的所有文件提供服务。但是几乎所有的 web 服务器都会这样做。

一旦我们运行它，并打开我们的浏览器。我们看到一个 `Run` 按钮，点击它，将执行我们的应用程序。然后我们点击它并检查控制台：

![hellowasm](https://blog.gopheracademy.com/postimages/advent-2018/go-in-the-browser/hellowasm.png)

优美！我们刚刚使用 Go 编写了一个程序并在浏览器中运行了它。

到现在为止还挺好。但这是一个简单的“hello world”程序。一个现实世界中的 web 应用程序需要与 DOM 进行交互。我们需要对按钮点击事件进行响应，从文本框中获取数据，并将数据发送回 DOM。现在我们将构建一个最小的图片编辑器，这个示例将用到所有的这些功能。

## DOM API

首先，为了让 Go 代码与浏览器进行交互，我们需要一个 DOM API。我们需要 `syscall/js` 库来帮助我们解决这个问题。它是一个非常基础但却强大的 DOM API，我们在其上构建我们的应用程序。在我们转向制作我们的应用程序之前，让我们快速了解它的一些功能。

**回调**

为了响应 DOM 事件，我们声明了回调并用这样的事件将它们连接起来：

```go
import "syscall/js"

// Declare callback
cb := js.NewEventCallback(js.PreventDefault, func(ev js.Value) {
	// handle event
})

// Hook it up with a DOM event
js.Global().Get("document").
	Call("getElementById", "myBtn").
	Call("addEventListener", "click", cb)

// Call cb.Release() on your way out.
```

**更新 DOM**

要从 Go 内部更新 DOM，我们可以这样做-

```go
import "syscall/js"

js.Global().Get("document").
		Call("getElementById", "myTextBox").
		Set("value", "hello wasm")
```

你甚至可以调用 js 函数和操作本地原生 js 对象，就像 `FileReader` 或 `Canvas` 一样。请随时查看 [syscall/js](https://golang.org/pkg/syscall/js/) 文档以获取更多详细的信息。

好了，现在开始构建我们的应用程序！

## 一个像样的web应用

我们将构建一个小的应用程序，它将获取一个图片输入，然后对图片进行一个操作，如亮度、对比度、色调、饱和度，最后将图片输出回浏览器中。每一个效果都将会用滑块，用户可以更改这些效果并实时查看目标图像的变化。

首先，我们需要从浏览器获取输入图片到我们的 Go 代码中，以便我们可以处理它。为了有效的做到这一点，我们需要采取一些 `unsafe` 技巧，具体细节跳过。一旦我们获取到了图片，它就完全在我们的掌控之下了，我们就可以随心所欲的做任何事情。下面是图片加载器回调的简短片段，为简洁起见略有优化：

```go
onImgLoadCb = js.NewCallback(func(args []js.Value) {
	reader := bytes.NewReader(inBuf) // inBuf is a []uint8 slice where our image is loaded
	sourceImg, _, err := image.Decode(reader)
	if err != nil {
		// handle error
	}
	// Now the sourceImg is an image.Image with which we are free to do anything!
})

js.Global().Set("loadImage", onImgLoadCb)
```

然后我们从任何效果滑块中获取用户的值，并操纵图片。我们使用很棒的 [bild](https://github.com/anthonynsimon/bild) 库。这是操作对比度回调的一小部分片段：

```go
import "github.com/anthonynsimon/bild/adjust"

contrastCb = js.NewEventCallback(js.PreventDefault, func(ev js.Value) {
	delta := ev.Get("target").Get("valueAsNumber").Float()
	res := adjust.Contrast(sourceImg, delta)
})

js.Global().Get("document").
		Call("getElementById", "contrast").
		Call("addEventListener", "change", contrastCb)
```

在此之后，我们将图片编码为 jpeg 格式并将其发送回浏览器。这是完整的应用程序操作：

我们加载图片：

![initial](https://blog.gopheracademy.com/postimages/advent-2018/go-in-the-browser/initial.png)

改变对比度：

![contrast](https://blog.gopheracademy.com/postimages/advent-2018/go-in-the-browser/contrast.png)

改变色调：

![hue](https://blog.gopheracademy.com/postimages/advent-2018/go-in-the-browser/hue.png)

太棒了，我们可以在浏览器中本地操作图片而不需要编写一行 Javascript 代码！可以在 [这里](https://github.com/agnivade/shimmer) 找到源码。

请注意，所有这些都是在浏览器本身中完成的。这里没有 Flash 插件、JavaApplets 或 Silverlight。开箱即用的浏览器本身对 WebAssembly 提供了支持。

## 最后说两句

我的一些结束语：

* 由于 Go 是一门垃圾回收的语言，因此整个运行周期都是在 wasm 二进制文件当中。正因为如此，这些二进制文件通常是MB级别的大小。与 C/Rust 语言相比，这是一个痛点；因为向浏览器发送 MB 级别的数据是不理想的。然而，要是 wasm 规范本身支持 GC，那么这可能会改变。
* 在 Go 中对 wasm 进行支持官方在进行试验。`syscall/js` API 本身是在不断变化，将在可能还会变。如果你发现一个 bug，请在我们的 [issue tracker](https://github.com/golang/go/issues) 上提出 issue。
* 与所有集数一样，WebAssembly 也不是一颗银弹。有时，简单的 js 更快更容易编写。然而，wasm 本身正在开发中，并将推出更多的功能。线程支持就是这样一个特性。

希望这篇文章展示了 WebAssembly 一些很酷的方面，以及如何使用 Go 编写一个功能齐全的 web 应用程序。如果你发现一个 bug，请尝试解决一下，并提出 issue。如果您需要任何帮助，请随时访问 [#webassembly](https://gophers.slack.com/) 频道。

comments powered by Disqus

-----

via: https://blog.gopheracademy.com/advent-2018/go-in-the-browser/

作者：[Agniva De Sarker](https://blog.gopheracademy.com/advent-2018/go-in-the-browser/)
译者：[PotoYang](https://github.com/PotoYang)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出