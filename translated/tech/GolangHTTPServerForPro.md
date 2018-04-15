#专业Golang HTTP服务器
####如何使用Go启动新的Web项目，使用路由，中间件和让我们加密认证。

Golang有一个很棒的自带http服务器软件包，不用说就是：net/http，它非常简单，但是功能非常强大。 定义处理路由的函数，端口是80。

```go

package main

import (
	"io"
	"net/http"
)
func main() {
	http.HandleFunc("/", helloWorldHandler)
	http.ListenAndServe(":80", nil)
}
func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!")
}
```

不错，但是我们可以使用一个更加强大的路由器，比如Gorilla包：`gorilla / mux` [http://www.gorillatoolkit.org/pkg/mux](http://www.gorillatoolkit.org/pkg/mux)

它实现了一个请求路由器和一个调度器。 它允许您创建具有命名参数的路由，限制在http动词和主机/域管理。

[img]
Gorilla Routing in action!
大猩猩路由在行动！

Updating the previous exemple with this package allow us to manage easily many routes with simples configurations:

在之前的例子上使用Gorilla，使我们能够使用简单配置轻松管理多条路线：

```go
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/products/{key}", ProductHandler)
    r.HandleFunc("/articles/{category}/", ArticlesCategoryHandler)
    r.HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler)
    http.Handle("/", r)
}
```



## 使用 `Alice` 来管理我们的中间件

如果您使用网络服务器软件包，[中间件模式](https://en.wikipedia.org/wiki/Middleware)非常常见。 如果您还没有看到它，您应该在2015年Golang UK Conference上观看Mat Ryer的视频，了解中间件的强大功能。([完整的博客文章在这里](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81))

[video]

另一篇关于中间件模式的文章[http://www.alexedwards.net/blog/making-and-using-middleware](http://www.alexedwards.net/blog/making-and-using-middleware)
正如作者的描述([Github](https://github.com/justinas/alice)):

> `Alice` 提供了一种便捷的方式来链接您的HTTP中间件功能和应用程序处理程序。

简单说，它把

```go
Middleware1(Middleware2(Middleware3(App)))
```

转换到

```go
alice.New(Middleware1, Middleware2, Middleware3).Then(App)
```

我们的第一个例子，加上 `Alice` 之后：

```go

func main() {
    errorChain := alice.New(loggerHandler, recoverHandler)
r := mux.NewRouter()
    r.HandleFunc("/products/{key}", ProductHandler)
    r.HandleFunc("/articles/{category}/", ArticlesCategoryHandler)
    r.HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler)
    http.Handle("/", errorChain.then(r))
}
```

你可以串联许多 `handler`，如下描述了两个：

```go
func loggerHandler(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        h.ServeHTTP(w, r)
        log.Printf("<< %s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}
```

`loggerHandler` 和 `recoverHandler`:

```go
func recoverHandler(next http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %+v", err)
                http.Error(w, http.StatusText(500), 500)
            }
        }()
next.ServeHTTP(w, r)
    }
return http.HandlerFunc(fn)
}
```

现在，我们有一个配有强大的路由包的HTTP服务器。 您还可以轻松管理中间件，以快速扩展应用程序的功能。

[img]

Alice 使中间件无处不在！

***

##HTTP服务器不错，但HTTPS服务器更好！

使用 `Let's Encrypt` 服务,简单快捷的创建一个安全的HTTP服务器 。 `Let's Encrypt` 使用 [ACME协议](https://en.wikipedia.org/wiki/Automated_Certificate_Management_Environment) 来验证您是否控制指定的域名并向您颁发证书。 这就是所谓的认证，是的，有一个自动认证软件包：[acme / autocert](https://godoc.org/golang.org/x/crypto/acme/autocert)

```go
m := autocert.Manager{
    Prompt:     autocert.AcceptTOS,
    HostPolicy: autocert.HostWhitelist("www.checknu.de"),
    Cache:      autocert.DirCache("/home/letsencrypt/"),
}
```

使用 `tls` 创建 `http.server`：

```go
server := &http.Server{
    Addr: ":443",
    TLSConfig: &tls.Config{
        GetCertificate: m.GetCertificate,
    },
}
err := server.ListenAndServeTLS("", "") if err != nil {         log.Fatal("ListenAndServe: ", err) }
```

[img]

完成了！


----------------

via: https://medium.com/@ScullWM/golang-http-server-for-pro-69034c276355

作者：[Thomas P](https://medium.com/@ScullWM)
译者：[tingtingr](https://github.com/wentingrohwer)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
