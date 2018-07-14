首发于：https://studygolang.com/articles/13611

# 关于 Go 即将支持的 WebAssembly 的一些注意事项

这是一篇关于 webassembly 的即时记录，它的目的是给我做个备忘而不仅仅是如果使用它的教程。

即将发布的 Go 1.11 版本将支持 Wasm。@neelance 做了大部分的实施工作。对 wasm 的支持已经可以通过他在 github 上的工作分支进行测试。

看[这篇文章](https://blog.gopheracademy.com/advent-2017/go-wasm/)了解更多信息

## 工具链设置

要从 go 源码生产一个 wasm 文件，您需要从源码获取并为 go 工具集打补丁：

```
~ mkdir ~/gowasm
~ git clone https://go.googlesource.com/go ~/gowasm
~ cd ~/gowasm
~ git remote add neelance https://github.com/neelance/go
~ git fetch --all
~ git checkout wasm-wip
~ cd src
~ ./make.bash
```

然后使用这个版本的 Go，把 GOROOT 指到 ~/gowasm 并使用 ~/gowasm/bin/go 这个二进制文件。

## 第一个例子

按照惯例，让我们先来写个 “hello world”：

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello World!")
}
```

然后编译出文件并命名 **example.wasm**

```
GOROOT=~/gowasm GOARCH=wasm GOOS=js ~/gowasm/bin/go build -o example.wasm main.go
```

## 运行这个例子

这是官方文档的节选：

> 虽然有计划允许 WebAssembly 模块像 ES6 模块(...)那样加载，但目前 WebAssembly 必须由 JavaScript 加载并编译，对于基本的加载，有如下三个步骤：
> + 将 .wasm 字节转化为数组或 ArrayBuffer
> + 将字节编译为 WebAssemly.Module
> + 使用导入实例化的 WebAssembly.Module 以获取可调用的导出

幸运的是，Go 的作者已经提供了一个 Javascript 加载器（~/gowasm/misc/wasm/wasm_exec.js）来简化这个过程。它附带一个 HTML 文件，负责粘贴浏览器中所有的内容。

要实际运行我们的文件，需要将以下文件复制到一个目录中并作为 Web 服务启动：

```
~ mkdir ~/wasmtest
~ cp ~/gowasm/misc/wasm/wasm_exec.js ~/wasmtest
~ cp ~/gowasm/misc/wasm/wasm_exec.html ~/wasmtest/index.html
~ cp example.wasm ~/wasmtest
```

然后编辑这个 index.html 文件去运行正确的例子：

```javascript
// ...
WebAssembly.instantiateStreaming(fetch("example.wasm"), go.importObject).then((result) => {
	mod = result.module;
	inst = result.instance;
	document.getElementById("runButton").disabled = false;
});
// ...
```

理论上，任何 web 服务都可以运行它，但是当我们试着用 caddy 运行它时遇到一个问题。这个 javascript 加载器需要服务发送这个 wasm 文件的正确 mime 类型给它。

这有一个快速的破解方法来运行我们的测试：为我们的 wasm 文件写个带有特殊处理的 Go 服务。

```go
package main

import (
	"log"
	"net/http"
)

func wasmHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/wasm")
	http.ServeFile(w, r, "example.wasm")
}
func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("/example.wasm", wasmHandler)
	log.Fatal(http.ListenAndServe(":3000", mux))
}
```

*注意* 设置一个特殊的路由器来处理所有的 wasm 文件没什么大不了，如我所说，这是一个 POC，这篇文章只是关于它的附注。

然后使用`go run server.go`来启动服务，并打开浏览器访问 http://localhost:3000。

打开控制台看看！

## 和浏览器交互

让我们和世界互动。

### 解决 DOM 问题

*syscall/js* 包中包含允许通过 javascript API 与 DOM 交互的函数。要获取此包的文档，只需运行：
`GOROOT=~/gowasm godoc -http=:6060`
然后用浏览器访问 http://localhost:6060/pkg/syscall/js/ 。

让我们写个简单的 HTML 文件来显示一个输入框。然后从 webassembly，我们给这个元素绑定一个事件，并在监听到事件时触发一个动作。

编辑 index.html 并把代码放在 run 按钮下面：

```html
	<button onClick="run();" id="runButton" disabled>Run</button>
	<input type="number" id="myText" value="" />
</body>
```

然后修改 Go 文件：

```go
package main

import "fmt"

func main() {
	 c := make(chan struct{}, 0)
	 cb = js.NewCallback(func(args []js.Value) {
		  move := js.Global.Get("document").Call("getElementById", "myText").Get("value").Int()
		  fmt.Println(move)
	  })
	  js.Global.Get("document").Call("getElementById", "myText").Call("addEventListener", "input", cb)
	  // The goal of the channel is to wait indefinitly
	  // Otherwise, the main function ends and the wasm modules stops
	  <-c
}
```

像以前一样编译文件并刷新浏览器……打开控制台然后输入一个数字……瞧瞧

## 暴露函数

这有点辣手……我没有找到任何简单的方法将一个 Go 函数暴露给 Javascript 生态系统。我们需要做的是在 Go 文件中创建一个 Callback 对象并指定到一个 Javascript 对象。

为得到返回结果，我们不能返回一个值给 callback 而是使用 Javascript 对象代替。

这是新的 Go 代码：

```go
package main
import (
	"syscall/js"
)

func main() {
	c := make(chan struct{}, 0)
	add := func(i []js.Value) {
		js.Global.Set("output", js.ValueOf(i[0].Int()+i[1].Int()))
	}
	js.Global.Set("add", js.NewCallback(add))
	<-c
}
```

现在编译并运行代码。打开浏览器和控制台。

如果你输入 *output* 将返回 *Object not found*。现在您输入 *add(2,3)* 和 *output*...应该得到 5。

这不是很优雅的交互方式，但它按预期运行。

## 结论

Go 对 wasm 的支持刚刚开始，但正大力发展。许多功能现在都可运行。我甚至可以在浏览器运行一个完整的递归神经网络，这归功于 Gorgonia。我将稍后讲解这些。

---

via: https://blog.owulveryck.info/2018/06/08/some-notes-about-the-upcoming-webassembly-support-in-go.html

作者：[Parikshit Agnihotry](https://medium.com/@parikshit)
译者：[themoonbear](https://github.com/themoonbear)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
