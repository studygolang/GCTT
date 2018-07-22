首发于：https://studygolang.com/articles/13727

# 使用 Go 在 WASM 中进行图像处理的实验

Go 的主分支最近完成了一个 WebAssembly 的工作原型实现。作为 WASM 的爱好者，我自然要把玩一下。

这篇文章，我要记下周末我用 Go 做的处理图像实验的想法。这个演示只是从浏览器中获取图像输入，然后应用各种图像变换，如亮度，对比度，色调，饱和度等，最后将其转储回浏览器。这测试了两件事 - 简单的CPU绑定执行，这是图像转换应该做的事情，以及在 JS 和 Go 之间传递数据。

## 回调

应该明确如何在 JS 和 Go 之间进行调用，不是我们在 emscripten 中的常用的方式；它是暴露一个函数然后从 JS 调用它。在 Go 中，JS 的互操作是通过回调完成的。在您的 GO 代码中，设置可以从 JS 调用的回调。这些是您希望在 GO 代码中执行的主要事件处理程序。

它看起像这样 -

```go
js.NewEventCallback(js.PreventDefault, func(ev js.Value) {
	// handle event
})
```

这有个模式-随着你的应用增长，它成为一个 DOM 事件回调处理程序列表。我把它看作 REST 应用的 URL 处理器。

为了规范，我把所有回调作为我的主结构的方法并在一个地方关联它们。类似于你把 URL 处理器声明在不同的文件里并在一个地方设置所有路由一样。

```go
// Setup callbacks
s.setupOnImgLoadCb()
js.Global.Get("document").
	Call("getElementById", "sourceImg").
	Call("addEventListener", "load", s.onImgLoadCb)

s.setupBrightnessCb()
js.Global.Get("document").
	Call("getElementById", "brightness").
	Call("addEventListener", "change", s.brightnessCb)

s.setupContrastCb()
js.Global.Get("document").
	Call("getElementById", "contrast").
	Call("addEventListener", "change", s.contrastCb)
```

然后在一个单独文件里编写您的回调 -

```go
func (s *Shimmer) setupHueCb() {
	s.hueCb = js.NewEventCallback(js.PreventDefault, func(ev js.Value) {
		// quick return if no source image is yet uploaded
		if s.sourceImg == nil {
			return
		}
		delta := ev.Get("target").Get("value").Int()
		start := time.Now()
		res := adjust.Hue(s.sourceImg, delta)
		s.updateImage(res, start)
	})
}
```

## 执行

我吐槽的是图像数据从 Go 传给浏览器的方式。在图像上传时，我把 src 属性设置为整个图像的base64编码格式，该值传到 Go 代码中对其解码为二进制，应用转换然后再编回 base64 并设置目标图像的 src 属性。

这使得 DOM 非常沉重，需要从 Go 传递一个巨大的字符串到 JS。 如果 WASM 中 SharedArrayBuffer 有所支持可能会改善。我也在研究在画布中直接设置像素，看看有没有任何好处。即使为了消减这个 base64 转换也应该花些时间。（请不吝赐教其他方法）

## 性能

对于一个 100KB 的 JPEG图像，应用转换所需时间约为180～190毫秒。这个时间随着图像大小而增加。这是使用 Chrome 65 测试的。（FF一直报错，我也没时间调查）

![性能快照显示](https://raw.githubusercontent.com/studygolang/gctt-images/master/Experiments-with-image-manipulation-in-WASM-using-Go/wasm1.png)

性能快照显示

![堆相当大。堆快照大约1GB](https://raw.githubusercontent.com/studygolang/gctt-images/master/Experiments-with-image-manipulation-in-WASM-using-Go/wasm2.png)

堆相当大。堆快照大约1GB

## 整理想法

完整库在这 - [github.com/agnivade/shimmer](https://github.com/agnivade/shimmer)。随意使用。提醒一下这是我一天写的，所有显然会有改进的地方。我会进一步研究。

附：- 请注意，图像变换不会应用于另一个图像之上。就是说如果您更改亮度然后更改色调，则生成的图像将仅改变原始图像的色调。这是现在的待办项目。

---

via: http://agniva.me/wasm/2018/06/18/shimmer-wasm.html

作者：[AGNIVA](http://agniva.me/)
译者：[themoonbear](https://github.com/themoonbear)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
