已发布：https://studygolang.com/articles/12525

# Go1.10 支持 HTTPS 代理

Go1.9 出来后 6 个多月的时间，Go1.10 就被[发布](https://blog.golang.org/go1.10)。新版本带来大大小小的变化([发行说明](https://golang.org/doc/go1.10))，但是我想谈谈有关 `net/http` 包的改变。1.10 版本支持在 HTTPS([commit](https://github.com/hyangah/go/commit/ab0372d91c17ca97a8258670beadadc6601d0da2)) 上的代理，而在原来它只能通过使用普通的（未加密）HTTP 来和代理进行沟通。接下来让我们来看看它是否真的可以工作。

## Server

为了验证这一改变，首先请用 golang 启动一个简单的 HTTP(S) 代理服务器。具体做法可以从下面文章了解。

[HTTP(S) Proxy in Golang in less than 100 lines of code](https://medium.com/@mlowicki/http-s-proxy-in-golang-in-less-than-100-lines-of-code-6a51c2f2c38c)

## Client
```go
package main

import (
	"net/url"
	"net/http"
	"crypto/tls"
	"net/http/httputil"
	"fmt"
)

func main() {
	u, err := url.Parse("https://localhost:8888")
	if err != nil {
		panic(err)
	}
	tr := &http.Transport{
		Proxy: http.ProxyURL(u),
		// disabled HTTP/2
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("https://google.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q", dump)
}
```

## 1.9 和 1.10 对比

```
>go version
go version go1.10 darwin/amd64
>go run proxyclient.go
"HTTP/1.1 200 OK\r\nTransfer-Encoding:...

>go version
go version go1.9 darwin/amd64
>go run proxyclient.go
panic:Get https://google.com:malformed HTTP response "\x15\x03\x01\x00\x02\x02\x16"

...
```
从第一个结果看到，使用 Go1.10，我们通过代理服务器[https//google.com](https//google.com)监听[https://localhost:8888](https://localhost:8888)得到正确的响应。而第二个结果显示 Go1.9 搭建的 HTTP 客户端被拒绝。

如果你想了解更多关于 Go 更新的内容，请在这里关注我或者在[Twitter](https://twitter.com/mlowicki)上。

------------
via: https://medium.com/@mlowicki/https-proxies-support-in-go-1-10-b956fb501d6b

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[zhaohj1118](https://github.com/zhaohj1118)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
