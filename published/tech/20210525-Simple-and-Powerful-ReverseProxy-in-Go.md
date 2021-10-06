首发于：https://studygolang.com/articles/35262

# Go 简单而强大的反向代理（Reverse Proxy）

在本文中，我们将了解反向代理，它的应用场景以及如何在 Golang 中实现它。

反向代理是位于 Web 服务器前面并将客户端（例如 Web 浏览器）的请求转发到 Web 服务器的服务器。它们让你可以控制来自客户端的请求和来自服务器的响应，然后我们可以利用这个特点，可以增加缓存、做一些提高网站的安全性措施等。

在我们深入了解有关反向代理之前，让我们快速看普通代理（也称为正向代理）和反向代理之间的区别。

在**正向代理**中，代理代表原始客户端从另一个网站检索数据。它位于客户端（浏览器）前面，并确保没有后端服务器直接与客户端通信。所有客户端的请求都通过代理被转发，因此服务器只与这个代理通信（服务器会认为代理是它的客户端）。在这种情况下，代理可以隐藏真正的客户端。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210525-Simple-and-Powerful-ReverseProxy-in-Go/forward-proxy.png)

另一方面，**反向代理**位于后端服务器的前面，确保没有客户端直接与服务器通信。所有客户端请求都会通过反向代理发送到服务器，因此客户端始终只与反向代理通信， 而从不会直接与实际服务器通信。在这种情况下，代理可以隐藏后端服务器。

几个常见的反向代理有 Nginx， HAProxy。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210525-Simple-and-Powerful-ReverseProxy-in-Go/reverse-proxy.png)

## 反向代理使用场景

负载均衡（Load balancing）：反向代理可以提供负载均衡解决方案，将传入的流量均匀地分布在不同的服务器之间，以防止单个服务器过载。

防止安全攻击：由于真正的后端服务器永远不需要暴露公共 IP，所以 DDoS 等攻击只能针对反向代理进行，
这能确保在网络攻击中尽量多的保护你的资源，真正的后端服务器始终是安全的。

缓存：假设你的实际服务器与用户所在的地区距离比较远，那么你可以在当地部署反向代理，它可以缓存网站内容并为当地用户提供服务。

SSL 加密：由于与每个客户端的 SSL 通信会耗费大量的计算资源，因此可以使用反向代理处理所有与 SSL 相关的内容，然后释放你真正服务器上的宝贵资源。

## Golang 实现

```go
import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewProxy takes target host and creates a reverse proxy
// NewProxy 拿到 targetHost 后，创建一个反向代理
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(url), nil
}

// ProxyRequestHandler handles the http request using proxy
// ProxyRequestHandler 使用 proxy 处理请求
func ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	// initialize a reverse proxy and pass the actual backend server url here
	// 初始化反向代理并传入真正后端服务的地址
	proxy, err := NewProxy("http://my-api-server.com")
	if err != nil {
		panic(err)
	}

	// handle all requests to your server using the proxy
	// 使用 proxy 处理所有请求到你的服务
	http.HandleFunc("/", ProxyRequestHandler(proxy))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

是的没错！这就是在 Go 中创建一个简单的反向代理所需的全部内容。我们使用标准库 `net/http/httputil` 创建了一个单主机的反向代理。到达我们代理服务器的任何请求都会被代理到位于 `http://my-api-server.com`。如果你对 Go 比较熟悉，这个代码的实现一目了然。

## 修改响应

`HttpUtil` 反向代理为我们提供了一种非常简单的机制来修改我们从服务器获得的响应，
可以根据你的应用场景来缓存或更改此响应，让我们看看应该如何实现：

```go
// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ModifyResponse = modifyResponse()
	return proxy, nil
}

func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		resp.Header.Set("X-Proxy", "Magical")
		return nil
	}
}

```

可以在 `modifyResponse` 方法中看到 ，我们设置了自定义 Header 头。同样，你也可以读取响应体正文，并对其进行更改或缓存，然后将其设置回客户端。

在 `modifyResponse` 中，可以返回一个错误（如果你在处理响应发生了错误），如果你设置了 `proxy.ErrorHandler`, `modifyResponse` 返回错误时会自动调用 `ErrorHandler` 进行错误处理。

```go
// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ModifyResponse = modifyResponse()
	proxy.ErrorHandler = errorHandler()
	return proxy, nil
}

func errorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("Got error while modifying response: %v \n", err)
		return
	}
}

func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		return errors.New("response body is invalid")
	}
}
```

## 修改请求

你也可以在将请求发送到服务器之前对其进行修改。在下面的例子中，我们将会在请求发送到服务器之前添加了一个 Header 头。同样的，你可以在请求发送之前对其进行任何更改。

```go
// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		modifyRequest(req)
	}

	proxy.ModifyResponse = modifyResponse()
	proxy.ErrorHandler = errorHandler()
	return proxy, nil
}

func modifyRequest(req *http.Request) {
	req.Header.Set("X-Proxy", "Simple-Reverse-Proxy")
}

```

## 完整代码

```go
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		modifyRequest(req)
	}

	proxy.ModifyResponse = modifyResponse()
	proxy.ErrorHandler = errorHandler()
	return proxy, nil
}

func modifyRequest(req *http.Request) {
	req.Header.Set("X-Proxy", "Simple-Reverse-Proxy")
}

func errorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("Got error while modifying response: %v \n", err)
		return
	}
}

func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		return errors.New("response body is invalid")
	}
}

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	// initialize a reverse proxy and pass the actual backend server url here
	proxy, err := NewProxy("http://my-api-server.com")
	if err != nil {
		panic(err)
	}

	// handle all requests to your server using the proxy
	http.HandleFunc("/", ProxyRequestHandler(proxy))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

反向代理非常强大，如文章之前所说，它有很多应用场景。你可以根据你的情况对其进行自定义。 如果遇到任何问题，我非常乐意为你提供帮助。如果你觉得这篇文章有趣，请分享一下，让更多 Gopher 可以阅读！ 非常感谢你的阅读。

---

via: https://blog.joshsoftware.com/2021/05/25/simple-and-powerful-reverseproxy-in-go/

作者：[Anuj Verma](https://blog.joshsoftware.com/author/devanujverma/)
译者：[h1z3y3](https://www.h1z3y3.me)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
