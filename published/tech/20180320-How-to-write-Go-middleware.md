已发布：https://studygolang.com/articles/12931

# 如何写 go 中间件

编写 go 中间件看起来挺简单的，但是有些情况下我们可能会遇到一些麻烦。

让我们来看一些例子。

## 读取请求

我们例子中的所有中间件都会接收一个 `http.Handler` 作为参数，并返回一个 `http.Handler` 。
这样便于将中间件链接起来。我们所有的中间件将会遵循如下基本模式：

```go
func X(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Something here...
		// 这里还有一些其他信息
		h.ServeHTTP(w, r)
	})
}
```

假设我们想要将所有请求重定向到一个末尾斜杠(例如，`/messages/`)，
这与重定向到它们的“非跟踪-斜杠”（比如 `/messages`）是等价的。我们可以这么写 ：

```go
func TrailingSlashRedirect(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path[len(r.URL.Path)-1] == '/' {
			http.Redirect(w, r, r.URL.Path[:len(r.URL.Path)-1], http.StatusMovedPermanently)
			return
		}
		h.ServeHTTP(w, r)
	})
}
```

如此简单。

## 修改请求

假设我们要向请求添加一个头部信息，或者修改它。
`http.Handler` 文档说明如下:

 >除了读取主体之外，处理程序不应修改所提供的请求。

 Go 标注库在[传递 `http.Request` 对象到响应链之前会先拷贝 `http.Request`](https://golang.org/src/net/http/server.go#L1981)，我们也应该这样做。
 假设我们要为每个请求设置一个 `Request-Id` 头部信息，用于内部跟踪。
 创建 `*Request` 的一个浅拷贝，并在代理之前修改头。

 ```go
func RequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := new(http.Request)
		*r2 = *r
		r2.Header.Set("X-Request-Id", uuid.NewV4().String())
		h.ServeHTTP(w, r2)
	})
}
```

## 写入响应头信息

如果你想设置响应头信息，你可以编写它们，然后代理请求。

```go
func Server(h http.Handler, servername string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", servername)
		h.ServeHTTP(w, r)
	})
}
```

## 使用最后写入的头部信息

上面写入响应头信息存在的问题是：如果内部处理程序也设置了服务器头部信息，你设置的响应头信息将被覆盖。
如果你不想公开内部软件的服务器头部信息，或者在向客户机发送响应之前要去掉头部，这可能会有问题。

为了应对这么问题，我必须自己实现 `ResponseWriter` 接口。
大多数情况下，我们会把它代理给潜在的 `ResponseWriter`，但是如果用户试图写一个响应，我们会偷偷添加头部信息。

```go
type serverWriter struct {
	w            http.ResponseWriter
	name         string
	wroteHeaders bool
}

func (s *serverWriter) Header() http.Header {
	return s.w.Header()
}

func (s *serverWriter) WriteHeader(code int) http.Header {
	if s.wroteHeader == false {
		s.w.Header().Set("Server", s.name)
		s.wroteHeader = true
	}
	s.w.WriteHeader(code)
}

func (s *serverWriter) Write(b []byte) (int, error) {
	if s.wroteHeader == false {
		// We hit this case if user never calls WriteHeader (default 200)
		// 如果用户从不调用 `WriteHeader`，我们就会遇到这种情况。
		s.w.Header().Set("Server", s.name)
		s.wroteHeader = true
	}
	return s.w.Write(b)
}
```

要将它用于我们的中间件，我们会这么写：

```go
func Server(h http.Handler, servername string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &serverWriter{
			w:    w,
			name: servername,
		}
		h.ServeHTTP(sw, r)
	})
}
```

## 如果用户从不调用 `Write` 或 `WriteHeader` 呢?

如果用户不调用 `Write` 或 `WriteHeader` 方法，比如一个状态码为 200 的空响应体，或者一个对可选请求的响应，
对于这些情况我们的拦截函数都不会被执行。
所以，鉴于这种情况我们应当在 `ServeHTTP` 调用之后添加至少一个检查。

```go
func Server(h http.Handler, servername string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &serverWriter{
			w:    w,
			name: servername,
		}
		h.ServeHTTP(sw, r)
		if sw.wroteHeaders == false {
			s.w.Header().Set("Server", s.name)
			s.wroteHeader = true
		}
	})
}

```

## 其他 `ResponseWriter` 接口

ResponseWriter接口只需要实现三个方法。
但实际上，它也可以响应其他接口，例如 `http.Pusher`。此外，你的中间件可能会意外禁用HTTP/2支持，这是不好的。

```go
// Push implements the http.Pusher interface.
// Push 实现 http.Pusher 接口
func (s *serverWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := s.w.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// Flush implements the http.Flusher interface.
// Flush 实现 http.Flusher 接口
func (s *serverWriter) Flush() {
	f, ok := s.w.(http.Flusher)
	if ok {
		f.Flush()
	}
}
```

## 就是这样

祝你好运！你正在写什么中间件呢，它们运行的怎样？

---

via: https://kev.inburke.com/kevin/how-to-write-go-middleware/

作者：[Kevin Burke](https://kev.inburke.com/about)
译者：[SergeyChang](https://github.com/SergeyChang)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

