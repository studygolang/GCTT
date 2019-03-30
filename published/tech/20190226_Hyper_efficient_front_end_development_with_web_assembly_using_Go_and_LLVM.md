首发于：https://studygolang.com/articles/19400

# 使用 Go 和 LLVM 进行 Web Assembly 的高效前端开发

我今天想分享一个非常酷的项目[tinygo](https://tinygo.org/) 经历。

首先让我说 Go 中的 Web Assembly 有一个大问题，它太过依赖于完成任务而定制的 API。在我看来 `syscalls/js` 是错误的使用 Web Assembly 方式：

- Go 开发者不应该学习 JavaScript
- 随着时间的推移，Web Assembly 将获得自己的 API，可能基于[WebIDL](https://github.com/WebAssembly/design/issues/1148)
- 它对运行 Web Assembly 的主机做了太多假设，甚至导入让模块运行所需的功能也是如此
- 调用本身在 Web Assembly 堆栈机器中（不是一些抽象）的外部函数将始终是最快的方法。

当我发现 `tinygo` 非常高兴，它编译一个 Web assembly 模块并使用更少的假设来发现该模块，而且有一个非常简单的基于注释的系统用于导入函数 - 据我所知，在主线 Go 编译器中是不可能的。

`tinygo` 利用 LLVM，并能够很好地将 Web assembly 模块减少相当大的数量。这让我能做的就是编写像这个 `hello world` 这样的非常小的模块：

```go
package main
//go:export console_log
func console_log(msg string)
//go:export main
func start(){
  console_log("hello world")
}
func main() {}
```

编译下来，这个模块只有 2 个导入要求：

- console_log
- io_std_out（因为目前 tinygo 假设有一些运行时）

导出：

- main

在我迄今为止看到的用于 Web assembly 技术的所有选项中，这和我希望的一样好。

我一直在编写一些库，这些库使用基于浏览器 Web IDL（浏览器环境具有哪些功能的标准化描述）来暴露一些非常有效的 DOM 操作 API：[richardanaya/wasm-module](https://github.com/richardanaya/wasm-module)

你可以通过大量的命令式函数的调用来获取 DOM 中的资源句柄并操纵它们，而不是制作系统 `syscalls/js`。这种方法的优点是它非常简单和 C 类似，并且不需要您自己的特殊代码生成，并且完全与技术无关。只需导入您需要的功能。

这是一个画布应用程序的示例

```go
package main
//go:export global_getWindow
func GetWindow() int32
//go:export Window_get_document
func GetDocument(window int32) int32
//go:export Document_querySelector
func QuerySelector(document int32 ,query string) int32
//go:export HTMLCanvasElement_getContext
func GetContext(element int32,context string) int32
//go:export CanvasRenderingContext2D_fillRect
func FillRect(ctx ,x ,y ,w ,h int32)
//go:export CanvasRenderingContext2D_set_fillStyle
func FillStyle(ctx int32, fillStyle string)
func cstr(s string) string{
  return s+"\000"
}
//go:export main
func start(){
  win := GetWindow()
  doc := GetDocument(win)
  canvas := QuerySelector(doc,cstr("#screen"))
  ctx := GetContext(canvas,cstr("2d"))
  FillRect(ctx,0,0,50,50)
  FillStyle(ctx,cstr("red"))
  FillRect(ctx,10,10,50,50)
  FillStyle(ctx,cstr("grey"))
  FillRect(ctx,20,20,50,50)
}
func main() {}
```

你可以[在这里](https://richardanaya.github.io/wasm-module/examples/tinygo_canvas/index.html) 看到这个工作

总的来说，我很高兴在我的工具箱中再添加一个工具来更加简单方便的创建 Web assembly。也许通过一些工作，`tinygo` 可以生成的更简洁，并像 `Rust` 一样在网络的下一个技术平台上坚实可靠。

---

via: https://medium.com/@richardanaya/hyper-efficient-front-end-development-with-web-assembly-using-go-and-llvm-8e6a1ccdd2bc

作者：[Richard Anaya](https://medium.com/@richardanaya)
译者：[lovechuck](https://github.com/lovechuck)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
