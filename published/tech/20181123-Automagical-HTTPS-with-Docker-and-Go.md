首发于：https://studygolang.com/articles/19232

# 通过 Docker 和 Go 实现 https 访问

![title](https://raw.githubusercontent.com/studygolang/gctt-images/master/Automagical-HTTPS-with-Docker-and-Go/1_3VNZrkS-sVUaa-xJQMXCnA.jpeg)

最近一直在构建很多 Webhook，通常我需要通过 HTTPS 协议为应用程序提供服务。快速实现这一目标的一种常见方法是使用 Let's Encrypt，但设置起来可能有点繁琐。我希望能够将这个过程完全自动化，包括证书更新。我一直在使用 docker 来构建我的应用程序，并且希望保持构建过程和容器镜像尽可能轻量级。最后，我想将整个过程作为应用程序代码，这样我可以轻松地进行动态更改和重新部署。此外，不依赖运维 /shell 脚本，我的应用程序可移植并轻松部署到许多不同的环境。

幸运的是，使用 Go [acme/autocert](https://godoc.org/golang.org/x/crypto/acme/autocert) 包可以实现所有这些功能。ACME 表示*自动证书管理环境*，是一种低成本和自动化 TLS 证书生成和验证的协议。

Let's Encrypt 是利用 ACME 向任何要求认证域的人提供免费的域名证书认证。验证的一种方法是应用程序从 Let's Encrypt（通过安全连接）请求秘钥令牌，然后 Let's Encrypt 将对正在验证的域名发出 HTTP 请求。如果应用程序可以将秘钥令牌提供给 Let's Encrypt，则验证该域名具有对域的控制权，并且 Let's Encrypt 将签署证书以在域上使用。

## Let's get cooking

![let's_get_cooking](https://raw.githubusercontent.com/studygolang/gctt-images/master/Automagical-HTTPS-with-Docker-and-Go/1_ZSx82UcchxKQvn6eI4yCZA.jpeg)
首先你需要有一个域名。只要能将应用程序部署到该域所托管的服务上，任何域 / 子域都可以。您可以拥有多个域，例如，您可以在同一个应用程序中托管和验证 chat.example.com 和 www.example.com。对于此示例，我将在 kappa.serv.brendanr.net 上托管我的应用程序。理想情况下，您的应用程序不会进行负载平衡 - 虽然可以在负载平衡域上实现 ACME 验证，但它更复杂。

接下来，你需要一个可以部署应用的服务器。确保配置 DNS 服务器以将域名指向该服务器。

最后，您需要使用上文提到的 [acme/autocert](https://godoc.org/golang.org/x/crypto/acme/autocert) 包来请求和响应 ACME 请求。 [Krzysztof Kowalczyk](https://blog.kowalczyk.info/) 提供了一个很好的[示例](https://github.com/kjk/go-cookbook/blob/master/free-ssl-certificates/main.go)，您可以阅读它，但我将向您展示一个为更简单的版本以便更好地解释它是如何工作的。

## 应用代码

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello world"))
    })

    certManager := autocert.Manager{
        Prompt:     autocert.AcceptTOS,
        Cache:      autocert.DirCache("cert-cache"),
        // Put your domain here:
        HostPolicy: autocert.HostWhitelist("kappa.serv.brendanr.net"),
    }

    server := &http.Server{
        Addr:    ":443",
        Handler: mux,
        TLSConfig: &tls.Config{
            GetCertificate: certManager.GetCertificate,
        },
    }

    Go http.ListenAndServe(":80", certManager.HTTPHandler(nil))
    server.ListenAndServeTLS("", "")
}
```

## 逐步分解

```go
mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello world"))
})
```

在这里，我们设置首页请求处理程序。一旦你开始着手去做 , 你应该以同样的方式为你的应用程序去注册处理函数。

```go
certManager := autocert.Manager{
    Prompt: autocert.AcceptTOS,
    Cache:  autocert.DirCache("cert-cache"),
    HostPolicy: autocert.HostWhitelist("kappa.serv.brendanr.net"),
}
```

上述代码为 ACME 相关配置信息。`Prompt: autocert.AcceptTOS` 字段意味着您接受了 Let's Encrypt 的[服务条款](https://letsencrypt.org/repository/)。`Cache` 字段指定 autocert 包是否缓存以及缓存证书的方式。Let's Encrypt 存在请求[速率](https://letsencrypt.org/docs/rate-limits/) 限制，限制了您申请证书的频率，因此将证书存储在后续可以检索的地方非常重要。这里我们指定证书存储在 `cert-cache` 目录中。最后，`HostPolicy` 字段允许我们将期望申请证书的域名加入白名单。如果没有该设置，攻击者可能会耗尽所分配的速率限额，并可能会阻止您生成所需的证书，因此配置该字段很重要。

```go
server := &http.Server{
    Addr:    ":443",
    Handler: mux,
    TLSConfig: &tls.Config{
        GetCertificate: certManager.GetCertificate,
    },
}

go http.ListenAndServe(":80", certManager.HTTPHandler(nil))
server.ListenAndServeTLS("", "")
```

最后，我们配置并启动 HTTP 和 HTTPS 服务。HTTPS 服务将使用我们之前编写的处理程序进行响应。在响应 HTTPS 请求时，会自动获取 HTTPS 证书（来自缓存或 Let's Encrypt）。HTTP 服务专门用于允许 Let's Encrypt 对我前面提到的秘钥令牌发出请求。也可以将 HTTP 服务重定向到 HTTPS -- 请查看之前的[示例](https://github.com/kjk/go-cookbook/blob/master/free-ssl-certificates/main.go)。

构建和部署您的应用程序，并向其发出 https 请求！虽然应用程序第一次使用 Let's Encrypt 进行 ACME 质询流程请求时会有几秒耗时，但您的应用程序仍将使用安全可靠的 HTTPS 页面进行响应。

## You promised me Docker

![You_promised_me_Docker](https://raw.githubusercontent.com/studygolang/gctt-images/master/Automagical-HTTPS-with-Docker-and-Go/1_EnBK1tCbV3p6VAdV2ivBKQ.png)
在容器化这个应用程序前，你需要避免几个问题。以下是完整的 Dockerfile。我正在使用的这个 Dockerfile 是基于 Pierre Prinetti [Go 1.11 Web service Dockerfile](https://medium.com/@pierreprinetti/the-go-1-11-dockerfile-a3218319d191) 的修改版本。

```bash
# 设置构建镜像的 Go 版本参数
# 默认版本 Go 1.11
ARG GO_VERSION=1.11

# 第一步： 构建可执行程序。
FROM Golang:${GO_VERSION}-alpine AS builder

# 获取 Git 需要的依赖项。
RUN apk add --no-cache ca-certificates Git

# 设置除 $GOPATH 以外的工作目录来确保对模块的支持。
WORKDIR /src

# 首先获取依赖包；它们在每次构建过程中变化不大，因此将会缓存以加速下一次构建。
RUN Go mod download

# 从上下文导入代码
COPY ./ ./

# 将可执行程序构建为 `/app`，将构建标记为静态链接。
RUN CGO_ENABLED=0 Go build \
    -installsuffix 'static' \
    -o /app .

# 最后一步 : 运行容器
FROM scratch AS final

# 从第一阶段导入已编译的可执行文件。
COPY --from=builder /app /app
# 导入根 ca 证书（需要 Let's Encrypt）
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 应用程序提供 443 和 80 服务端口。
EXPOSE 443
EXPOSE 80

# 将证书缓存目录挂载至本地磁盘，这样在部署新版本时仍然存在。
VOLUME ["/cert-cache"]

# 运行编译的二进制文件。
ENTRYPOINT ["/app"]
```

Dockerfile 分两个阶段，构建和最终阶段。这使我们能够发送极小的最终镜像 - 我们甚至可以使用 `scratch`（空的分层）作为基础。上面的代码和 Dockerfile 最终构建了一个只有 7 MB 的镜像！

注意事项：

* **您必须在最终镜像上安装 ca-certificates**，即使您的应用程序未进行 TLS 连接也是如此。这是因为您的应用程序将对 Let's Encrypt 发起的所有请求都是基于 HTTPS，因此您需要根证书。
* **即使您不打算通过 HTTP 提供任何服务，也必须开放 80 端口**。这是因为 Let's Encrypt 需要能够向我们的应用程序发出 HTTP 请求。
* **您应该将缓存目录缓存至本地磁盘，即便在部署过程中也会缓存证书**。如果你不这样做，你可能会超出 Let's Encrypt 速率限制。

## 总结一下

我希望这篇文章可以帮助您在设置下一个 Web 服务时减少一些时间和困惑，或者至少让您有兴趣再写一些 Go。

完整的代码可查看 https://github.com/bmon/go-web-base

最初发表在 [brendanr.net](https://brendanr.net/blog/go-docker-https/)。

---
via: https://medium.com/weareservian/automagical-https-with-docker-and-go-4953fdaf83d2

作者：[Brendan Roy](https://medium.com/@brendan.roy)
译者：[liulizhi](https://github.com/liulizhi)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
