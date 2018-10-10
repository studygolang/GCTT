# 在WASM上使用GO语言做图像处理的实验

2018.6.18

GO项目主分支最近完成了对WebAssemblyde组件的工作原型。作为一名WASM爱好者，我自然先行试用一番。

在周末我用go语言做了图片处理的实验，在这篇文章中我将写下的感悟和经验。这个demo仅仅是从浏览器输入一个图片，然后对图像做了各种处理，如亮度、对比度、色调、饱和度等，然后在从浏览器返回下载。这测试了两件事-图像转化是简单的CPU处理过程，并且这个过程是通过JS和Goland传输数据完成的

## 回调

我们应该搞清楚Go与JS平台如何是进行交互的。它并不是常用的emsciptern方式。它是通过暴露一个函数并在JS中调用这个函数。在Go中通过回调来与JS进行交互操作。在你的Go代码中，你需要定义一个在js中调用的回调函数。这个回调函数就是主要的事件处理器。

就像这样-

```js
js.NewEventCallback(js.PreventDefault, func(ev js.Value) {
	// handle event
})
```

这里有一个模板-随着程序的不断增加，这个模板将变成一个回电函数列表来对应DOM的各种响应事件。我感觉它就像一个REST风格的URL处理器。

我是这样处理的，我在同一个地方将所有的回调方法都声明在全局函数。就像在不同的文件中定义处理URL请求的函数，但在同一个地方设置所有的路由。

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

然后在单独的文件中写下你们的回调函数-

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

## 实施

我首先要抱怨一下图片数据从Go平台传输到浏览器平台的方式。

在上传图片时，我将整个图片的源文件进行了base64编码，然后将这值传到go代码中,在go代码中将其转换为二进制编码的文件并进行处理，然后又将处理后的图片转为base64返回到浏览器。

这使得DOM非常重，需要从GO到JS传递一个巨大的字符串，也许SharedArrayBuffer支持WASM的话将会有所改善。我也在观察是否直接在画布上设置像素会好一些。是不是直接去掉base64转化的过程会节省一些时间呢。（各位大佬有什么好想法吗:grin:）

## 表现

对于大小为100KB的JPEG图像，转换所需的时间约为180-190ms。图片越大转换时间越长，我使用的是Chrome 65。（FF给我提示了一些错误，但是我还没有时间去检查:sweat_smile:）
![Result](https://agniva.me/assets/wasm1.png)

## 总结

完整的地址在这里-https://github.com/agnivade/shimmer。你可以在上面随便看看改改。但是提醒一下，这个是我一天内完成的。所以上面肯定会有一些不足需要改进。我将随后再研究一下。

补充一点 -  请注意图像转化的应用对最先的处理是不起作用的，如果你改变了亮度然后再改变色调，那最后的结果仅仅改变了色调。这是目前需要处理的一个点。


----------------

via: http://agniva.me/wasm/2018/06/18/shimmer-wasm.html

作者：[AGNIVA](http://agniva.me/)
译者：[xmge](https://github.com/xmge)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出


