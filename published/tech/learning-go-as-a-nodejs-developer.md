# 一名 Node.js 工程师的 Go 学习之旅

GERGELY NEMETH 2018-02-08

近几年， Kubernetes 一直是容器编排和管理平台的首选。对于我来说，也很想搞清楚它的底层是如何工作的，所以我决定来学一学 Go 这门语言。

在这边文章中，我会基于我的经验，并从一个 Node.js 工程师的视角出发。因此我会特别关注：

- 依赖管理
- 如何处理异步操作

那就让我们开始吧 :)

## 什么是 Go ?

> Go 是一个开源的编程语言，它能让构造简单、可靠且高效的软件变得容易. - golang.org

Go 由 Google 的 Robert Griesemer, Rob Pike, 和 Ken Thompson 于 2009 年发布。它是一款静态类型的编译语言，拥有垃圾收集机制，基于 CSP 并发模型来处理异步操作。 Go 还有类 C 的语法：

```go
package main
import "fmt"
func main() {
	fmt.Println("hello world")
}
```

安装 Go: 官方指导链接： https://golang.org/doc/install.

## Go 的依赖管理

如果打算写大量的 JavaScript 代码，首要问题就是依赖管理问题，Go 是怎么处理的呢？有两个办法：
- go get
- dep

用 npm 的术语来说，你可以把他们看作是：在你需要使用 npm install -g 的时候，使用 go get，然后使用 dep 来管理不同项目的依赖。

要安装 dep，可以通过 go get 来安装，使用如下命令：

```
go get -u github.com/golang/dep/cmd/dep
```

然而，使用 go get 有一个缺点 —— go get 并不处理版本，它仅仅是获取 Github 仓库的最新版本。这就是为什么推荐大家安装并使用 dep。如果是 Mac 系统，安装 dep 也可以通过如下命令：

```go
brew install dep
brew upgrade dep
```
（如果是其他操作系统，安装请参见： https://golang.org/doc/install）

一旦安装了 dep，你就可以使用 **dep init** 来初始化项目，就好像使用 **npm init** 初始化 nodejs 项目一样。

> 开发Go项目之前，你需要花点时间设置好GOPATH环境变量。—— [官方指导链接](https://golang.org/doc/install)

dep 会像 npm 一样，创建一个 Node.js 项目中类似 package.json 的文件来描述工程 —— Gopkg.toml。类似 package-lock.json，也会有一个 Gopkg.lock 文件。不同于 nodejs 项目将依赖放入 node modules 文件夹中，dep 将依赖放入一个叫作 vendor 的文件夹中。

要添加依赖，你只需要运行 dep ensure -add github.com/pkg/errors 命令。运行结束后，这个依赖就会出现在 lock 和 toml 文件中：

```
[[constraint]]
  name = "github.com/pkg/errors"
  version = "0.8.0"
```

## Go处理异步操作

当用 JavaScript 写异步代码时，我们会用到一些库或者语言特性，比如：
- async 库
- Promises
- 或者异步函数

有了它们，我们可以很轻易的从文件系统中读取文件：

```javascript
const fs = require('fs')
const async = require('async')
const filesToRead = [
	'file1',
	'file2'
]
async.map(filesToRead, (filePath, callback) => {
	fs.readFile(filePath, 'utf-8', callback)
}, (err, results) => {
	if (err) {
	return console.log(err)
	}
	console.log(results)
})
```

再让我们来看看 Go 是怎么实现的：

```go
package main
import (
	"fmt"
	"io/ioutil"
)
func main() {
	datFile1, errFile1 := ioutil.ReadFile("file1")
	if errFile1 != nil {
		panic(errFile1)
	}
	datFile2, errFile2 := ioutil.ReadFile("file2")
	if errFile2 != nil {
		panic(errFile2)
	}
	fmt.Print(string(datFile1), string(datFile2))
}
```

我们来一行行看看上面代码是怎么工作的：

- *import* —— 有了 import 关键字，你可以引入项目依赖的包文件，就像 Node.js 中的 *require*
- func main —— 应用程序的入口
- ioutil.ReadFile —— 该函数尝试去读取文件，并由两个返回值：
  - datFile1 如果读操作成功，
  - errFile1 如果读取过程中有错
    - 你可以在这里处理错误，或者直接让程序崩溃
- fmt.Print 仅仅是打印结果到标准输出

上面的例子是可以正常运行的，但是读取文件是一个接一个读取。—— 让我们来稍加改进，将它异步化吧！

Go有一个叫作 *goroutines* 的概念来处理多线程。一个 *goroutine* 是一个轻量级的线程，它由 Go runtime 来管理。*goroutine* 使得你可以并发地跑Go的函数。

我最终使用 *errgroup* 包来管理或者说同步 *goroutines*。这个包提供同步机制，错误传播，以及对同一个由一组 *goroutines* 子任务组成的公共任务提供上下文取消机制。

有了 *errgroup*，我们可以重写读文件的代码片段，并发地执行：

```go
package main
import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"golang.org/x/sync/errgroup"
)
func readFiles(ctx context.Context, files []string) ([]string, error) {
	g, ctx := errgroup.WithContext(ctx)
	results := make([]string, len(files))
	for i, file := range files {
		i, file := i, file
		g.Go(func() error {
			data, err := ioutil.ReadFile(file)
			if err == nil {
				results[i] = string(data)
			}
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}
func main() {
	var files = []string{
		"file1",
		"file2",
	}
	results, err := readFiles(context.Background(), files)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, result := range results {
		fmt.Println(result)
	}
}
```

## 用 Go 来构建一条 REST API

在 Node.js 中，当谈及到选择一套框架来写 HTTP 服务的时候，我们有一堆的选择。Go 在这方面也不例外。Google 一下后，我选择 *Gin* 来开始。

*Gin* 的接口类似于 *Express* 或者 *Koa*，包含中间件支持，JSON 校验以及渲染：

```go
package main
import "github.com/gin-gonic/gin"
func main() {
	// 创建默认不带任何中间件的路由
	r := gin.New()
	// 默认gin的输出为标准输出
	r.Use(gin.Logger())
	// Recovery中间件从异常中恢复，并回复500
	r.Use(gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// 监听 0.0.0.0:8080
	r.Run(":8080")
}
```

这就是我目前掌握的 —— 还没有生产经验。如果你这篇博客对你有帮助，并且你想学到更多关于 Go 的东西，就跟我联系吧，我会持续分享我的 Go 学习之旅。

## 更多资源
上面文章主要参考了如下链接文章：
- https://peter.bourgon.org/go-best-practices-2016/
- https://golang.github.io/dep
- https://blog.golang.org/defer-panic-and-recover
- https://gobyexample.com
- https://www.golang-book.com/

---

via: https://nemethgergely.com/learning-go-as-a-nodejs-developer/

作者：[GERGELY NEMETH](https://nemethgergely.com)
译者：[Chandler1142](https://github.com/chandler1142)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
