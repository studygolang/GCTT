已发布：https://studygolang.com/articles/12568

# 你的 pprof 暴露了

IPv4 扫描暴露出的 `net/http/pprof` 端点（endpoint）

原文发表日期: 2017/9/27

Go语言的 [net/http/pprof](https://golang.org/pkg/net/http/pprof/) 包是令人难以置信的强大的，调试正在运行的生产服务器的这个功能微不足道，而在这个调试过程，就很容易不经意间将调试信息暴露给世界。在这篇文章中，我们用 [zmap project](https://github.com/zmap) 作为例子，展示一个现实中真正的问题，并且说明你可以采取的预防措施。

> 早期版本提出，暴露的端点可能泄露源代码。[Aram Hăvărneanu 指出了这个错误](https://github.com/golang/go/issues/22085#issuecomment-333166626)，本文已修正。

## 引言

通过一个 `import _ "net/http/pprof"` ，你可以将分析端点添加到 HTTP 服务器。

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // here be dragons
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

这个服务不仅会对你说 `Hello World!`，它还会通过uri路径 `/debug/pprof` 返回诊断报告。

- `/debug/pprof/profile`: 30秒的CPU状态信息
- `/debug/pprof/heap`: 内存的堆信息
- `/debug/pprof/goroutine?debug=1`: 所有协程的堆栈踪迹
- `/debug/pprof/trace`: 执行的追踪信息

举个例子，假如我们使用 [hey](https://github.com/rakyll/hey) （负载测试工具）给这个服务增加一些负载，同时我们查看下堆栈信息，如下

```shell
$ wget -O trace.out http://localhost:8080/debug/pprof/trace
$ go tool trace trace.out
```

在几秒钟内，我们以较细的颗粒度地检查服务器

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/pprof-showing/trace.png)

此功能对于追踪仅在生产环境中出现的错误和性能问题非常重要。但是，权限越大则责任越大。

## 开启 Pprof 服务

这种机制非常简单，他只需一次 import ! 它可以在任何的地方 import，甚至在你使用的 library 中。当你兴奋地使用它来追踪你的协程泄露问题时，你或许会忘记事后移除这个调试入口。这就导致了问题：

有多少 pprof 服务暴露在网络中？

为了回答这个问题，我们可以尝试在 IPv4 扫描开启 pprof 的服务器。为了限制搜索范围，我们可以选择一些合理的端口。

- 6060 官方文档建议的
- 8080 经常在入门教程出现
- 80 标准的 HTTP 端口
- 443 HTTPS 的端口

接下来让你失望了，因为收到云服务器的警告邮件，所以作者没有完成这个搜索工作。尽管我可以用更狡猾的方法来完成这个工作，但是我已经有足够的证据说服自己这个问题在现实中是真实存在的。

[zmap project](https://github.com/zmap) 用一行命令就可以进行这些类型的扫描

```sh
$ zmap -p 6060 | zgrab --port 6060 --http="/debug/pprof/"
```

[zmap](https://github.com/zmap/zmap) 扫描 IPv4 范围中开启6060端口的服务并调用它，然后 `banner grabber` 的 [zgrab](https://github.com/zmap/zgrab) 采集 HTTP 请求的 `GET /debug pprof` 的响应结果与问题。我们可以认为任意响应为 `200 OK` 的服务器与包含 `goroutine` 的响应体即为命中。下面是我们发现的内容:

- 至少有 69 个 IP 使用 `pprof` 开启了 6060 端口
- 同上，至少有 70 个 IP 开启了 8080 端口
- 在扫描 80 端口之前， [Google Cloud](https://cloud.google.com/) 怀疑我的服务器被入侵成为挖矿机（mining cryptocurrency）而停止了我的账号.

好吧，这个"挖矿"部分有点怪异，不过打住。现在我们知道了有许多机器在公网上开放了 `pprof` 服务，这正是我强调的问题。

我根据 WHOIS 信息向服务器所有者发送了邮件报告问题。我不得不说来自 [linode](https://www.linode.com/) 的回应非常快速积极。

我很想看到更有才华的人可以完成这个全网扫描，我怀疑还有更多的服务器在 80 端口与 443 端口暴露了 pprof 服务。

## 风险

安全问题：

- 显示函数名与文件路径
- 分析数据可能揭示商业敏感信息(例如，web服务的流量)
- 分析会降低性能，为 DoS 攻击增加助攻

## 预防

Farsight Security [警告过这个问题，并且提供了建议](https://www.farsightsecurity.com/2016/10/28/cmikk-go-remote-profiling/)

> 一个简单而有效的方式是将pprof http服务器放在本地主机上的一个单独的端口上，与应用程序http服务器分开。

总之，你需要安排两台HTTP服务器。常见的设置是

- 应用程序服务将80端口暴露在公网上
- `pprof`服务监听本地6060端口并且限于本地访问

原生的写法是不使用全局的 HTTP 方法的情况下构建主应用程序(使用隐藏配置 `http.DefaultServeMux` )，而是用标准的方法启动你的 pprof 服务。

```go
// Pprof server.
go func() {
	log.Fatal(http.ListenAndServe("localhost:8081", nil))
}()

// Application server.
mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
})
log.Fatal(http.ListenAndServe(":8080", mux))

```

如果处于任何原因，你都需要将全局 `http.DefaultServeMux` 用于你的应用服务器，你可以切换它然后像往常执行。

```go
// Save pprof handlers first.
pprofMux := http.DefaultServeMux
http.DefaultServeMux = http.NewServeMux()

// Pprof server.
go func() {
	log.Fatal(http.ListenAndServe("localhost:8081", pprofMux))
}()

// Application server.
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
})
log.Fatal(http.ListenAndServe(":8080", nil))
```

我封装了一个 [professor package](https://github.com/mmcloughlin/professor)，通过它来使用 `net/http/pprof` 包，并且提供一些便利的方法。

```go
// Pprof server.
professor.Launch("localhost:8081")

// Application server.
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
})
log.Fatal(http.ListenAndServe(":8080", nil))
```

## 结论

`net/http/pprof` 是很强大，但是请不要让你的调试信息暴露给全世界，遵循以上预防措施，你会没事的。

---

via: http://mmcloughlin.com/posts/your-pprof-is-showing

作者：[mmcloughlin](http://mmcloughlin.com/)
译者：[lightfish-zhang](https://github.com/lightfish-zhang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
