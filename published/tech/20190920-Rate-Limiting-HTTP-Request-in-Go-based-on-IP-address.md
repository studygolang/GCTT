首发于：https://studygolang.com/articles/25121

# Go 中基于 IP 地址的 HTTP 限流

如果你想限制一个正在运行的 HTTP 服务的请求量，你可以使用现有的轮子工具，比如说 [github.com/didip/tollbooth](https://github.com/didip/tollbooth) ，但是如果写一些简单的东西，你自己去实现也没有那么难。

我们可以用这个包 `x/time/rate` 。

在这篇教程中，我们将基于用户的 IP 地址构造一个简单的限流中间件。

## Pure HTTP Server

我们来开始构建一个简单的 HTTP 服务，这是一个大流量的服务，这也是我们为什么要在这里加上限制的原因。

```go
package main

import (
    "log"
    "net/http"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", okHandler)

    if err := http.ListenAndServe(":8888", mux); err != nil {
        log.Fatalf("unable to start server: %s", err.Error())
    }
}

func okHandler(w http.ResponseWriter, r *http.Request) {
    // Some very expensive database call
    w.Write([]byte("alles gut"))
}
```

在 `main.go` 文件中，我们用 `:8888` 端口启动了一个仅有单一控制器的服务。

## golang.org/x/time/rate

我们将用 `x/time/rate` 这个包，它提供了一个令牌桶限流算法。 [rate#Limiter](https://godoc.org/golang.org/x/time/rate#Limiter) 控制事件允许发生的频率，它实现了一个容量为 b 的“令牌桶”，最初是满的并以每秒 r 个令牌的速度重新填充。在足够的时间间隔里，限流器限制速度为每秒 r 个令牌，最多为桶的最大容量 b 。

因为我们想根据 IP 地址来限流，我们需要维护一个限流器字典。

```go
package main

import (
    "sync"

    "golang.org/x/time/rate"
)

// IPRateLimiter .
type IPRateLimiter struct {
    ips map[string]*rate.Limiter
    mu  *sync.RWMutex
    r   rate.Limit
    b   int
}

// NewIPRateLimiter .
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    i := &IPRateLimiter{
        ips: make(map[string]*rate.Limiter),
        mu:  &sync.RWMutex{},
        r:   r,
        b:   b,
    }

    return i
}

// AddIP creates a new rate limiter and adds it to the ips map,
// using the IP address as the key
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()

    limiter := rate.NewLimiter(i.r, i.b)

    i.ips[ip] = limiter

    return limiter
}

// GetLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise calls AddIP to add IP address to the map
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    limiter, exists := i.ips[ip]

    if !exists {
        i.mu.Unlock()
        return i.AddIP(ip)
    }

    i.mu.Unlock()

    return limiter
}
```

`NewIPRateLimiter` 创建了一个 IP 限流器的实例，HTTP 服务将调用这个实例的 `GetLimiter` 来获取指定的 IP 限流器（从字典里获取或者构造一个新的）。

## Middleware

让我们来升级我们的 HTTP Server 并在所有的控制器中添加中间件。所以如果 IP 达到限制将返回 429 表示大量请求，否则，它将继续执行请求。

在 `limitMiddleware` 方法中，每一次中间件接受一个请求，我们都会调用全局限制器的 `Allow()` 方法。如果桶里没有令牌了，`Allow()` 方法将返回 false 并且我们返回给用户 429 。否则，调用 `Allow()` 将从桶里消耗一个令牌并且我们将控制权传递给下一个处理程序。

```go
package main

import (
    "log"
    "net/http"
)

var limiter = NewIPRateLimiter(1, 5)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", okHandler)

    if err := http.ListenAndServe(":8888", limitMiddleware(mux)); err != nil {
        log.Fatalf("unable to start server: %s", err.Error())
    }
}

func limitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        limiter := limiter.GetLimiter(r.RemoteAddr)
        if !limiter.Allow() {
            http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func okHandler(w http.ResponseWriter, r *http.Request) {
    // Some very expensive database call
    w.Write([]byte("alles gut"))
}
```

## Build & Run

```go
go get golang.org/x/time/rate
go build -o server .
./server
```

## Test

[vegeta](https://github.com/tsenart/vegeta)（它是用 Go 写的）是一个很不错的工具，我喜欢用它来做 HTTP 负荷测试。

```
brew install vegeta
```

我们需要创建一个简单的配置文件说明我们想测试什么请求。

```
GET http://localhost:8888/
```

运行 10 秒钟，每个单位时间发 100 个请求 。

```go
vegeta attack -duration=10s -rate=100 -targets=vegeta.conf | vegeta report
```

结果你将会看到一些请求返回 200 ，但是大部分返回 429 。

---

via: https://dev.to/plutov/rate-limiting-http-requests-in-go-based-on-ip-address-542g

作者：[Alex Pliutau](https://dev.to/plutov)
译者：[Alihanniba](https://github.com/Alihanniba)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
