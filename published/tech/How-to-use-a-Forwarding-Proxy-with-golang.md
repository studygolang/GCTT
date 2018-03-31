已发布：https://studygolang.com/articles/12726

# 如何用 Go 语言实现正向代理

正向代理是处理一组内网客户端发往外部机器的网络请求的一种代理方式。

实际上，正向代理是你的应用和你所要连接的服务器之间的中间人。它在 HTTP(S) 协议上起作用，并且被部署在网络设施的边缘。

你通常可以在大型组织或大学中见到正向代理，它被用来进行授权管理或网络安全方面的控制。

我发现在使用容器或者动态的云环境工作时，正向代理很有用，因为你会面临一组服务器和外部网络的通信问题。

如果你在 AWS、AZure 之类的动态环境下工作，你会拥有一批数量不定的服务器和一批数量不定的公网 IP。你把应用运行在 Kubernetes 集群上时也是一样，容器可能遍布四处。

现在假设有客户让你提供一个公网 IP 的范围，因为他需要设置防火墙。你如何提供这个特性呢？这个问题有些情况下很简单，有些情况下可能非常复杂。

2015 年 12 月 1 日，有一位用户在 [CircleCI 论坛](https://discuss.circleci.com/t/circleci-source-ip/1202)上提了这个问题，并且问题还未关闭。当然，CircleCI 很棒。我只是举个例子，并非要埋怨他们。

解决这个问题的一种可行方法是使用正向代理。你可以让一组节点以同一静态IP运转，然后把清单提供给客户即可。

几乎所有云服务提供商都是这样做的，比如 DigitalOcean 的浮动 IP（floating IP）、AWS 的弹性 IP（elastic IP）等。

你可以通过配置自己的应用来把请求转发到这个（代理）池中。这样，终点的服务所取得的IP就是正向代理节点的IP，而不是内部IP。

正向代理可以成为你的网络设施的又一安全层，因为你可以在一个中心化的地方极其方便地扫描和控制内部网络发出来的数据包。

正向代理不会带来单点故障，因为你可以运行多个正向代理服务，他们具有很好的伸缩性。

在底层，HTTP 的 `CONNECT` 方法就是一种正向代理。

> CONNECT 方法将请求连接转化为透明 TCP/IP 通道，通常用于在未加密的 HTTP 代理上进行 SSL 加密的通信（HTTPS）。

很多用各种语言写成的 HTTP 客户端已经以透明的方式支持这个功能了。在此，我以一个使用 Go 语言和 [privoxy](https://www.privoxy.org/) 的小例子来告诉大家，这很简单。

首先，我们创建一个名为 `whoyare` 的应用。它是一个 HTTP 服务器，功能是返回你的远程地址。

```go
package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	http.HandleFunc("/whoyare", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, _ := json.Marshal(map[string]string{
			"addr": r.RemoteAddr,
		})
		w.Write(body)
	})
	http.ListenAndServe(":8080", nil)
}
```

如果用 `GET` 方法访问路径 `/whoyare`，你会得到一个类似下面的 JSON 格式的响应：`{"addr": "34.35.23.54"}`，其中 `34.35.23.54` 就是你的公网地址。如果你使用的是笔记本电脑，那么在终端上发出请求后，你应该会得到 `localhost`的结果。可以用 `curl` 来试一下：

	18:36 $ curl -v http://localhost:8080/whoyare
	* TCP_NODELAY set
	> GET /whoyare HTTP/1.1
	> User-Agent: curl/7.58.0
	> Accept: */*
	>
	< HTTP/1.1 200 OK
	< Content-Type: application/json
	< Date: Sun, 18 Mar 2018 17:36:40 GMT
	< Content-Length: 31
	<
	* Connection #0 to host localhost left intact
	{"addr":"localhost:38606"}

我写了另外一个程序，它用 `http.Client` 在标准输出上打印响应。你可以在已经运行了 `whoyare` 服务的前提下运行这个程序：

```go
package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type whoiam struct {
	Addr string
}

func main() {
	url := "http://localhost:8080"
	if "" != os.Getenv("URL") {
		url = os.Getenv("URL")
	}
	log.Printf("Target %s.", url)
	resp, err := http.Get(url + "/whoyare")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	println("You are " + string(body))
}
```

这是个很简单的例子，但是你可以将其运用在很多复杂场合。

为了让例子更清楚，我在 DigitalOcean 上创建了两台虚拟机：一台运行 `privoxy`，另一台运行 `whoyare`。

* whoyare: public ip 188.166.17.88
* privoxy: public ip 167.99.41.79

Privoxy 是一个易用的正向代理。相比而言，Nginx 和 Haproxy 都不太适合在这种场景下使用，因为它们不支持`CONNECT`方法。

我在 Docker Hub 上创建了一个 docker 镜像，你可以直接运行它，默认使用端口 8118。

	core@coreos-s-1vcpu-1gb-ams3-01 ~ $ docker run -it --rm -p 8118:8118
	gianarb/privoxy:latest
	2018-03-18 17:28:05.589 7fbbf41dab88 Info: Privoxy version 3.0.26
	2018-03-18 17:28:05.589 7fbbf41dab88 Info: Program name: privoxy
	2018-03-18 17:28:05.591 7fbbf41dab88 Info: Loading filter file:
	/etc/privoxy/default.filter
	2018-03-18 17:28:05.599 7fbbf41dab88 Info: Loading filter file:
	/etc/privoxy/user.filter
	2018-03-18 17:28:05.599 7fbbf41dab88 Info: Loading actions file:
	/etc/privoxy/match-all.action
	2018-03-18 17:28:05.600 7fbbf41dab88 Info: Loading actions file:
	/etc/privoxy/default.action
	2018-03-18 17:28:05.607 7fbbf41dab88 Info: Loading actions file:
	/etc/privoxy/user.action
	2018-03-18 17:28:05.611 7fbbf41dab88 Info: Listening on port 8118 on IP address
	0.0.0.0
	
第二步，编译`whoyare`并且把可执行文件用scp传送到服务器，可使用以下命令：

	$ CGO_ENABLED=0 GOOS=linux go build -o bin/server_linux -a ./whoyare

应用运行起来之后，我们就可以用 cURL 来直接或者通过 privoxy 发送请求了。

直接发送请求如下：

	$ curl -v http://your-ip:8080/whoyare
	
cURL 使用环境变量`http_proxy`来配置代理进行请求转发：

	$ http_proxy=http://167.99.41.79:8118 curl -v http://188.166.17.88:8080/whoyare
	*   Trying 167.99.41.79...
	* TCP_NODELAY set
	* Connected to 167.99.41.79 (167.99.41.79) port 8118 (#0)
	> GET http://188.166.17.88:8080/whoyare HTTP/1.1
	> Host: 188.166.17.88:8080
	> User-Agent: curl/7.58.0
	> Accept: */*
	> Proxy-Connection: Keep-Alive
	>
	< HTTP/1.1 200 OK
	< Content-Type: application/json
	< Date: Sun, 18 Mar 2018 17:37:02 GMT
	< Content-Length: 29
	< Proxy-Connection: keep-alive
	<
	* Connection #0 to host 167.99.41.79 left intact
	{"addr":"167.99.41.79:58920"}
	
如你所见，我设置了 `http_proxy=http://167.99.41.79:8118` 之后，响应不再包含我的公网 IP 了，而是代理的 IP。

privoxy 处应该会留下如下的请求日志：

	2018-03-18 17:28:22.886 7fbbf41d5ae8 Request: 188.166.17.88:8080/whoyare
	2018-03-18 17:32:29.495 7fbbf41d5ae8 Request: 188.166.17.88:8080/whoyare 
	
你之前运行的客户端默认连接到 `localhost:8080`，但可以通过设置环境变量 `URL=http://188.166.17.88:8080` 来覆盖目标地址。运行以下命令可以直接到达 `whoyare`。

	$ URL=http://188.166.17.88:8080 ./bin/client_linux
	2018/03/18 18:37:59 Target http://188.166.17.88:8080.
	You are {"addr":"95.248.202.252:38620"}
	
Go语言的 `HTTP.Client` 包支持一组和代理相关的环境变量，设置这些环境变量可以对运行期间的服务立刻生效，十分灵活。

	export HTTP_PROXY=http://http_proxy:port/
	export HTTPS_PROXY=http://https_proxy:port/
	export NO_PROXY=127.0.0.1, localhost
	
前两个环境变量很简单，一个是 HTTP 代理，一个是 HTTPS 代理。`NO_PROXY` 排除了一组主机名，当要访问的主机在这个清单里的时候，请求不经过代理。我这里配置的是 `localhost` 和 127.0.0.1。
	
	HTT_PROXY=http://forwardproxy:8118
	     +--------------+           +----------------+         +----------------+
	     |              |           |                |         |                |
	     |   client     +----------^+ forward proxy  +--------^+    whoyare     |
	     |              |           |                |         |                |
	     +--------------+           +----------------+         +----^-----------+
	                                                                |
	                                                                |
	    +---------------+                                           |
	    |               |                                           |
	    |   client      +-------------------------------------------+
	    |               |
	    +---------------+
	   HTTP_PROXY not configured
	   
配置了环境变量的客户端将会通过代理访问，其他客户端将直接访问。

这个控制粒度很重要。你不仅可以按进程去控制是否经过代理，还可以按请求去控制，十分灵活。
	
	$ HTTP_PROXY=http://167.99.41.79:8118 URL=http://188.166.17.88:8080
	./bin/client_linux
	2018/03/18 18:39:18 Target http://188.166.17.88:8080.
	You are {"addr":"167.99.41.79:58922"}	 
	
可以看到，我们通过代理到达了 `whoyare`，响应中的 `addr` 是代理的地址。

最后一个命令有些怪异，但它只是为了展示 `NO_PROXY` 是如何工作的。我们在设置访问代理的同时，排除了 `whoyare` 的 URL。正如我们期望的那样，请求没有经过代理：

	$ HTTP_PROXY=http://167.99.41.79:8118 URL=http://188.166.17.88:8080 NO_PROXY=188.166.17.88 ./bin/client_linux
	2018/03/18 18:42:03 Target http://188.166.17.88:8080.
	You are {"addr":"95.248.202.252:38712"}
	
本文应作为 Go 语言和正向代理的实用介绍来阅读。你可以订阅我的 [rss](https://gianarb.it/atom.xml)，或者在 [twitter](https://twitter.com/gianarb)上关注我。兴许我以后还会介绍如何用 Go 替代 privoxy 以及如何在 Kubernetes 集群上部署。所以，快告诉我先写哪部分吧！

----------------

via: https://gianarb.it/blog/golang-forwarding-proxy

作者：[gianarb](https://github.com/gianarb)
译者：[vincent08](https://github.com/vincent08)
校对：[Unknwon](https://github.com/Unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出