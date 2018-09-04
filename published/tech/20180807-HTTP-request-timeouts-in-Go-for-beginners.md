首发于：https://studygolang.com/articles/14405

# Go 语言 HTTP 请求超时入门

在分布式系统中，超时是基本可靠性概念之一。就像这条 [tweet](https://twitter.com/copyconstruct/status/1025241837034860544) 中提到的，它可以缓和分布式系统中不可避免出现的失败所带来的影响。

## 问题

> 如何条件性地模拟 504 http.StatusGatewayTimeout 响应。

当尝试在 [zalando/skipper](https://github.com/zalando/skipper/issues/633) 中实现 OAuth token 验证时，我不得不理解并尝试使用 httptest 实现一个测试用来在服务端超时的时候 [模拟一个](https://stackoverflow.com/questions/51319726/how-to-mimic-504-timeout-error-for-http-request-inside-a-test-in-go) [504 http.StatusGatewayTimeout](https://stackoverflow.com/questions/51319726/how-to-mimic-504-timeout-error-for-http-request-inside-a-test-in-go)。但因为服务端的延迟，只产生了客户端的超时。作为这门语言的初学者，跟大多数人一样，我像下面这样创建了一个标准的带有超时的 HTTP Client。

```go
client := http.Client{Timeout: 5 * time.Second}
```

当想要创建一个 client 用以发送 http 请求的时候，上面的代码看起来非常简单和直观。但是它下面却隐藏了大量低层次的细节，包括客户端超时，服务端超时和负载均衡超时。

## 客户端超时

客户端的 http 请求超时有多种形式，具体取决于在整个请求流程中超时帧的位置。整个请求响应流程由 `Dialer`，`TLS Handshake`，`Request Header`，`Request Body`，`Response Header` 和 `Response Body` 构成。根据请求响应流程中上述不同的部件，Go 提供了如下方式来创建带有超时的请求。

* `http.client`
* `context`
* `http.Transport`

## http.client:

`http.client` 超时是超时的高层实现，包含了从 `Dial` 到 `Response Body` 的整个请求流程。`http.client` 的实现提供了一个结构体类型可以接受一个额外的 `time.Duration` 类型的 `Timeout` 属性。这个参数定义了从请求开始到响应消息体被完全接收的时间限制。

```go
client := http.Client{Timeout: 5 * time.Second}
```

## context

Go 语言的 `context` 包提供了一些有用的工具通过使用 `WithTimeout`，`WithDeadline` 和 `WithCancel` 方法来处理超时，Deadline 和可取消的请求。有了 `WithTimeout`，你就可以通过 `req.WithContext` 方法给 `http.Request` 添加一个超时时间了。

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()
req, err := http.NewRequest("GET", url, nil)
if err != nil {
    t.Error("Request error", err)
}

resp, err := http.DefaultClient.Do(req.WithContext(ctx))
```

## http.Transport:

你也可以通过使用带有 `DialContext` 的自定义 `http.Transport` 来创建 `http.client` 这种低层次的实现来指定超时时间。

```go
transport := &http.Transport{
    DialContext: (&net.Dialer{
        Timeout: timeout,
    }).DialContext,
}

client := http.Client{Transport: transport}
```

## 解决方案

根据上述的问题和各类选择，我通过 `context.WithTimeout()` 创建了一个 `http.request`。但仍然失败并得到了以下错误。

```
client_test.go:40: Response error Get http://127.0.0.1:49597: context deadline exceeded
```

## 服务端超时

使用 `context.WithTimeout()` 的问题是它仍然只是模拟的请求的客户端。万一请求的头部或者消息体超出了预定义的超时时间，请求会在客户端直接失败而不会从服务端返回 `504 http.StatusGatewayTimeout` 状态码。

创建一个每次都超时的 httptest 服务代码如下：

```go
httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request){
    w.WriteHeader(http.StatusGatewayTimeout)
}))
```

但我只希望服务端根据客户端设置的值超时。为了让服务端根据客户端超时时间返回 504 状态码，你可以使用 `http.TimeoutHandler()` 函数来包装 handler 使请求在服务端失败。如下为符合场景需求的可运行的测试代码。

```go
func TestClientTimeout(t *testing.T) {
    handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        d := map[string]interface{}{
            "id":    "12",
            "scope": "test-scope",
        }

        time.Sleep(100 * time.Millisecond) //<- Any value > 20ms
        b, err:= json.Marshal(d)
        if err != nil {
            t.Error(err)
        }
        io.WriteString(w, string(b))
        w.WriteHeader(http.StatusOK)
    })

    backend := httptest.NewServer(http.TimeoutHandler(handlerFunc, 20*time.Millisecond, "server timeout"))

    url := backend.URL
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        t.Error("Request error", err)
        return
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        t.Error("Response error", err)
        return
    }

    defer resp.Body.Close()
}
```

在项目 [zalando/skipper](https://github.com/zalando/skipper/) 的 [oauth_test/TestOAuth2TokenTimeout](https://github.com/zalando/skipper/blob/master/filters/auth/oauth_test.go#L378:6) 的实现中包含了上述问题的具体信息。

也许 Go 语言初学者对于理解 http 超时高层次的原理是比较有用的。但如果你想知道更多关于 Go 中 `http timeouts` 的细节，这篇 [Cloudflare](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) 上的文章值得一读。

---

via: https://medium.com/@addityasingh/http-request-timeouts-in-go-for-beginners-fe6445137c90

作者：[Aditya pratap singh](https://medium.com/@addityasingh)
译者：[alfred-zhong](https://github.com/alfred-zhong)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
