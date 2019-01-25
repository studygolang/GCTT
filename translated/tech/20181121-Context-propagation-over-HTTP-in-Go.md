# 解析Go context 通过HTTP的传播

Go 1.7引入了一个内置的上下文（context）类型。在系统中，可以使用`context`传递请求范围的元数据，例如不同函数，线程甚至进程之间的请求ID。Go将`context`引入标准库的初衷是以统一同一进程内的context传播。因此整个库和框架可以使用标准context，同时可以避免代码碎片化。在引入该包之前，每个框架都在使用它们自己的context类型，并且没有两个context彼此兼容。在这种情况下传播当前context就要编写丑陋的胶水代码。

尽管引入公共上下文传播机制对于统一同一进程内的context传递很有用，但Go上下文包不提供任何串联context的支持。如上所述，在网络系统中，上下文应该在不同进程之间在线路上传播。例如，在多服务体系结构中，请求通过多个进程（几个微服务，消息队列，数据库），直到完成用户请求。能够在进程之间传播context对于底层应用的的context协作是非常重要的。
	如果要通过HTTP传播当前context，则需要自己序列化context。同样在接收端，你需要解析传入的请求并将值放入当前context中。假设，我们希望在上下文中传播请求ID：

```go
package request
import "context"
// WithID 将当前request Id放入context中.
func WithID(ctx context.Context, id string) context.Context {
 return context.WithValue(ctx, contextIDKey, id)
}
// IDFromContext 从context中取出request Id.
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

WithID允许我们读取请求ID，IDFromContext允许我们将请求ID放在给定的context中。一旦我们想要跨越线程传播context，我们就需要进行手动操作以将context置于同一条线上。同时，将其从传播线路上解析到接收端的context。

在HTTP上，我们可以将请求ID转储为header。大多数context元数据可以作为header传播。一些传输层可能不提供header或header可能不满足传播数据的要求（例如，由于大小限制和缺乏加密）。在这种情况下，由实现来决定如何传播context。

HTTP传播
没有一种方法能自动将context放入HTTP请求，反之亦然。由于无法迭代context值，因此也无法转储整个context。

```go
const requestIDHeader = "request-id"
// Transport 将请求context序列化为请求header.
type Transport struct {
 // Base生成request.
 // 默认使用http.DefaultTransport.
 Base http.RoundTripper
}
// RoundTrip 将请求 context转换成headers
// 并生成请求.
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

在上面的Transport中，请求ID（如果存在于请求上下文中）将作为“request-id”header传播。

类似地，处理程序可以解析传入的请求以将“request-id”放入请求context中。

```go
// Handler 将请求header反序列化为请求context.
type Handler struct {
 // Base is the actual handler to call once deserialization
 // 当context完成的时候，Base将会调用一次反序列化过程.
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

为了继续传播context，请确保将当前context传递给处理程序的传出请求。传入context将传播到https：//endpoint。

```go
http.Handle("/", &Handler{
    Base: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        req, _ := http.NewRequest("GET", "https://endpoint", nil)
        // Propagate the incoming context.
        req = req.WithContext(r.Context()) 
        // 生成request.
    }),
})
```