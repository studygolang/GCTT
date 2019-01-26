首发于：https://studygolang.com/articles/17940

# 解析 Go context 通过 HTTP 的传播

Go 1.7 引入了一个内置的上下文（context）类型。在系统中，可以使用 `context` 传递请求范围的元数据，例如不同函数，线程甚至进程之间的请求 ID。Go 将 `context` 引入标准库的初衷是以统一同一进程内的 context 传播。因此整个库和框架可以使用标准 context，同时可以避免代码碎片化。在引入该包之前，每个框架都在使用它们自己的 context 类型，并且没有两个 context 彼此兼容。在这种情况下传播当前 context 就要编写丑陋的胶水代码。

尽管引入公共上下文传播机制对于统一同一进程内的 context 传递很有用，但 Go 上下文包不提供任何串联 context 的支持。如上所述，在网络系统中，上下文应该在不同进程之间在线路上传播。例如，在多服务体系结构中，请求通过多个进程（几个微服务，消息队列，数据库），直到完成用户请求。能够在进程之间传播 context 对于底层应用的的 context 协作是非常重要的。

如果要通过 HTTP 传播当前 context，则需要自己序列化 context。同样在接收端，你需要解析传入的请求并将值放入当前 context 中。假设，我们希望在上下文中传播请求 ID：

```go
package request

import "context"
// WithID 将当前 request Id 放入 context 中 .
func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextIDKey, id)
}
// IDFromContext 从 context 中取出 request Id.
func IDFromContext(ctx context.Context) string {
	v := ctx.Value(contextIDKey)
	if v == nil {
		return ""
	}
	return v.(string)
}

type contextIDType struct{}
var contextIDKey = &contextIDType{}
// ..
```

WithID 允许我们读取请求 ID，IDFromContext 允许我们将请求 ID 放在给定的 context 中。一旦我们想要跨越线程传播 context，我们就需要进行手动操作以将 context 置于同一条线上。同时，将其从传播线路上解析到接收端的 context。

在 HTTP 上，我们可以将请求 ID 转储为 header。大多数 context 元数据可以作为 header 传播。一些传输层可能不提供 header 或 header 可能不满足传播数据的要求（例如，由于大小限制和缺乏加密）。在这种情况下，由实现来决定如何传播 context。

## HTTP 传播

没有一种方法能自动将 context 放入 HTTP 请求，反之亦然。由于无法迭代 context 值，因此也无法转储整个 context。

```go
const requestIDHeader = "request-id"
// Transport 将请求 context 序列化为请求 header.
type Transport struct {
	// Base 生成 request.
	// 默认使用 http.DefaultTransport.
	Base http.RoundTripper
}
// RoundTrip 将请求 context 转换成 headers
// 并生成请求 .
func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = cloneReq(r) // per RoundTrip interface enforces
	rid := request.IDFromContext(r.Context())
	if rid != "" {
		r.Header.Add(requestIDHeader, rid)
	}
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(r)
}
```

在上面的 Transport 中，请求 ID（如果存在于请求上下文中）将作为“ request-id ” header 传播。

类似地，处理程序可以解析传入的请求以将“ request-id ”放入请求 context 中。

```go
// Handler 将请求 header 反序列化为请求 context.
type Handler struct {
	// Base is the actual handler to call once deserialization
	// 当 context 完成的时候，Base 将会调用一次反序列化过程 .
	Base http.Handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rid := r.Header.Get(requestIDHeader)
	if rid != "" {
		r = r.WithContext(request.WithID(r.Context(), rid))
	}
	h.Base.ServeHTTP(w, r)
}
```

为了继续传播 context，请确保将当前 context 传递给处理程序的传出请求。传入 context 将传播到 https：//endpoint。

```go
http.Handle("/", &Handler{
	Base: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequest("GET", "https://endpoint", nil)
		// Propagate the incoming context.
		req = req.WithContext(r.Context())
		// 生成 request.
	}),
})
```

---

via: https://medium.com/@rakyll/context-propagation-over-http-in-go-d4540996e9b0

作者：[JBD](https://medium.com/@rakyll)
译者：[hantmac](https://github.com/hantmac)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
