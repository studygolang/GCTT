首发于：https://studygolang.com/articles/24876

# Go 中记录 HTTP 请求

如果你有运行的 HTTP 服务，你可能想记录 HTTP 请求。

请求日志有助于诊断问题。（哪些请求失败了？我们一天处理多少请求？哪些请求比较慢？）

这对于分析是必需的。（哪个页面受欢迎？网页的浏览者都来自哪里？）

这篇文章介绍了在 Go Web 服务器中，记录 HTTP 请求日志相关的全部内容。

这不是关于可复用的库，而是关于实现你自己的解决方案需要知道的事情，以及关于我日志记录的选择的描述。

你可以在示例应用上查看详细内容： https://github.com/essentialbooks/books/tree/master/code/go/logging_http_requests

我在 Web 服务 [OnePage](https://onepage.nopub.io/) 中用到了这个记录系统。

[记录什么信息](https://onepage.nopub.io/p/Logging-HTTP-requests-in-Go-233de7fe59a747078b35b82a1b035d36#63fd0006-6ebd-442c-a463-d11862e8c33c)

[获取要记录的信息](https://onepage.nopub.io/p/Logging-HTTP-requests-in-Go-233de7fe59a747078b35b82a1b035d36#c8a27402-1650-402a-8679-69214078b88a)

[日志文件的格式](https://onepage.nopub.io/p/Logging-HTTP-requests-in-Go-233de7fe59a747078b35b82a1b035d36#97da9f14-289e-42f6-94fd-936a4eb88f26)

[每日滚动日志](https://onepage.nopub.io/p/Logging-HTTP-requests-in-Go-233de7fe59a747078b35b82a1b035d36#99565a90-2f57-4aab-a5e7-5eb9a9194adc)

[长期存储以及分析](https://onepage.nopub.io/p/Logging-HTTP-requests-in-Go-233de7fe59a747078b35b82a1b035d36#a099947d-2079-4d1d-a996-41e4ed1ff02a)

[更多的 Go 资源](https://onepage.nopub.io/p/Logging-HTTP-requests-in-Go-233de7fe59a747078b35b82a1b035d36#4405e240-bd60-45a8-ba47-65e175eb7f8f)

[招聘 Go 开发者](https://onepage.nopub.io/p/Logging-HTTP-requests-in-Go-233de7fe59a747078b35b82a1b035d36#5076eef2-d176-43f3-bab5-c0d3030efa23)

## 记录什么信息

为了展示通常会记录什么信息，这里有一条 Apache 的扩展日志文件格式的日志记录样本。

```
111.222.333.123 HOME - [01/Feb/1998:01:08:39 -0800] "GET /bannerad/ad.htm HTTP/1.0" 200 198 "http://www.referrer.com/bannerad/ba_intro.htm" "Mozilla/4.01 (Macintosh; I; PPC)"
```

我们能看到：

- `111.222.333.123` ：客户端发起 HTTP 请求的 IP 地址。
- `HOME` ： 域（适用单个 Web 服务器提供多个域的情况）。
- `-` ：用户认证信息（这个例子下为空）。
- `[01/Feb/1998:01:08:39 -0800]` ：请求被记录的时间。
- `"GET /bannerad/ad.htm HTTP/1.0"` ：HTTP 方法，URL 以及协议类型。
- `200`：HTTP 状态码。200 代表请求被成功处理。
- `198`：响应体的大小。
- `"http://www.referrer.com/bannerad/ba_intro.htm"` ：引荐来源（referer）。
- `"Mozilla/4.01 (Macintosh; I; PPC)"` ：应该认为用户代理标志 HTTP 客户端（极大程度上是一个 web 浏览器）

我们可以记录更多的信息，或者选择不去记录上面的某些信息。

个人而言：

- 我也会记录服务器处理单次请求的耗时，毫秒为单位。（毫秒对我而言已经足够了，用微秒来记录也可以但有点过度了）
- 我不记录协议（比如 HTTP/1.0）。
- 服务器通常只提供单一用途，所以不需要记录域。
- 如果服务有用户认证信息，我也会记录用户 ID。

## 获取记录的信息

Go 中标准 HTTP 处理函数的签名如下：

```go
func(w http.ResponseWriter, r *http.Request)
```

我们会把日志记录作为所谓的中间件，这是一种向 HTTP 服务管道中添加可复用功能的一个方法。

我们有 `logReqeustHandler` 函数，它以 `http.Handler` 接口作为参数，然后返回另一个包装了原有处理器并添加了日志记录功能的 `http.Handler`。

```go
func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		// 在我们包装的时候调用原始的 http.Handle
		h.ServeHTTP(w, r)

		// 得到请求的有关信息，并记录之
		uri := r.URL.String()
		method := r.Method
		// ... 更多信息
		logHTTPReq(uri, method, ....)
	}

	// 用 http.HandlerFunc 包装函数，这样就实现了 http.Handler 接口
	return http.HandlerFunc(fn)
}
```

我们可以把中间件处理器嵌套到每一个（HTTP 处理器）的顶部，这样所有（处理器）都会拥有这些功能。

下面介绍了我们如何使用它来把日志记录功能添加到所有的请求函数：

```go
func makeHTTPServer() *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/", handleIndex)
	// ... 可能会添加更多处理器

	var handler http.Handler = mux
	// 用我们的日志记录器包装 mux 。 this will (译者注：应当是注释没写全)
	handler = logRequestHandler(handler)
	// ... 可能会添加更多中间件处理器

	srv := &http.Server{
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second, // Go 1.8 开始引进
		Handler:      handler,
	}
	return srv
}
```

首先，我们定义一个 struct 封装所有需要记录的信息：

```go
// LogReqInfo 描述了有关 HTTP 请求的信息（译者注：此处为作者笔误，应当是 HTTPReqInfo）
type HTTPReqInfo struct {
	// GET 等方法
	method string
	uri string
	referer string
	ipaddr string
	// 响应状态码，如 200，204
	code int
	// 所发送响应的字节数
	size int64
	// 处理花了多长时间
	duration time.Duration
	userAgent string
}
```

下面是 `logRequestHandler` 的全部实现：

```go
func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ri := &HTTPReqInfo{
			method: r.Method,
			uri: r.URL.String(),
			referer: r.Header.Get("Referer"),
			userAgent: r.Header.Get("User-Agent"),
		}

		ri.ipaddr = requestGetRemoteAddress(r)

		// this runs handler h and captures information about
		// HTTP request
		// 这里运行处理器 h 并捕获有关 HTTP 请求的信息
		m := httpsnoop.CaptureMetrics(h, w, r)

		ri.code = m.Code
		ri.size = m.BytesWritten
		ri.duration = m.Duration
		logHTTPReq(ri)
	}
	return http.HandlerFunc(fn)
}
```

我们复盘下这个简单的例子：

- `r.Method` 返回 HTTP 的方法，如 "GET", "POST" 等。
- `r.URL` 是一个解析后的 url，如 `/getname?id=5`，然后 `String()返回我们需要的字符串形式。
- `r.Header` 是代表 HTTP 头部信息的结构体。头部信息包含 `Referer` 和 `User-Agent` 以及其他信息。
- 为了记录服务器处理请求的耗时，我们在开始时记录了 `timeStart`, 调用处理器候，通过 `time.Since(timeStart)` 获取时长。

其他的信息则比较难获取。

获取客户端 IP 地址的问题是有可能涉及到 HTTP 代理。客户端向代理发起请求，然后代理向我们请求。于是，我们拿到了代理的 IP 地址，而不是客户端的。

因为这样，代理通常在请求的 HTTP 头部信息中以 `X-Real-Ip` 或者 `X-Forwarded-For` 来携带客户端真正的 IP 地址。

下面展示了如何提取这个信息：

```go
// Request.RemoteAddress 包含了端口，我们需要把它删掉，比如: "[::1]:58292" => "[::1]"
func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// requestGetRemoteAddress 返回发起请求的客户端 ip 地址，这是出于存在 http 代理的考量
func requestGetRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For 可能是以","分割的地址列表
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: 应当返回第一个非本地的地址
		return parts[0]
	}
	return hdrRealIP
}
```

捕获响应写对象（ResponseWriter）的状态码以及响应的大小更为困难。

`http.ResponseWriter` 并没有给我们这些信息。但幸运的是，这是一个简单的接口:

```go
type ResponseWriter interface {
    Header() Header
    Write([]byte) (int, error)
    WriteHeader(statusCode int)
}
```

写一个包装了原始响应的接口实现，并记录我们想要了解的信息，这是可行的。幸运如我们，已经有人在包 [httpsnoop](https://github.com/felixge/httpsnoop) 中实现了。

## 日志文件的格式

Apache 的日志格式比较紧凑，虽然具备人类可读性但却难于解析。

有的时候，我们也需要阅读日志分析，然后我不赞成为这个格式的实现解析器的想法。

从实现的角度来看，一个简单的方式是用 JSON 来记录，并且换行隔开。

对于这种方法我不喜欢的是：JSON 不易于阅读。

作为一个中间层，我创建了 `siser` 库，它实现了一个可扩展，易于实现和人类可读的序列化格式。 它非常适合用于记录结构化信息，我已经在多个项目用到它了。

下面展示了一个简单请求是如何被序列化的:

```
171 1567185903788 httplog
method: GET
uri: /favicon.ico
ipaddr: 204.14.239.58
code: 404
size: 758
duration: 0
ua: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:68.0) Gecko/20100101 Firefox/68.0
```

每个记录的第一行包含了以下信息：

- `171` 是其下记录的数据的大小。提前知道数据的大小确保了安全和高效的实现。
- `1567185903788` 是时间戳的 UNIX 格式（从系统纪元（Epoch）至今的秒数）。它让我们避免在数据里记录重复的时间戳。
- `httplog` 是记录的类型。这让我们可以往同一文件写不同类型的日志。在我们的场景下，所有记录的类型都是一样的。

然后第一行之后的数据都是 `key:value` 格式。

下面展示了我们如何序列化一条记录并把它写到日志文件：

```go
var (
	muLogHTTP sync.Mutex
)

func logHTTPReq(ri *HTTPReqInfo) {
	var rec siser.Record
	rec.Name = "httplog"
	rec.Append("method", ri.method)
	rec.Append("uri", ri.uri)
	if ri.referer != "" {
		rec.Append("referer", ri.referer)
	}
	rec.Append("ipaddr", ri.ipaddr)
	rec.Append("code", strconv.Itoa(ri.code))
	rec.Append("size", strconv.FormatInt(ri.size, 10))
	durMs := ri.duration / time.Millisecond
	rec.Append("duration", strconv.FormatInt(int64(durMs), 10))
	rec.Append("ua", ri.userAgent)

	muLogHTTP.Lock()
	defer muLogHTTP.Unlock()
	_, _ = httpLogSiser.WriteRecord(&rec)
}
```

## 日志每日滚动

我通常在 Ubuntu 上部署服务器，并把日志记录到 `/data/<service-name./log` 目录。

我们不能一直往同一个日志文件里写。否则到最后会用完所有空间。

对于长时间的日志，我通常每天一个日志文件，以日期命名。如 `2019-09-23.txt`, `2019-09-24.txt` 等等。

这有时称为日志滚动 ( log rotate).

为了避免重复实现这个功能，我写了一个库 [dailyrotate](https://github.com/kjk/dailyrotate)。

它实现了 `Write`, `Close` 以及 `Flush` 方法，所以它易于接入到现有已使用 `io.Reader` 等的代码。

你要指定使用哪个目录，以及日志命名的格式。这个格式通过 Go 的时间格式化函数来实现的。我通常使用 `2006-01-02.txt` 每天生成一个唯一的时间，并根据日期来排序，`txt` 则是工具识别文本文件而不是二进制文件的标志。

接着就和写普通的文件一样，以及确保代码会每天创建文件。

你也可以提供一个通知的回调，当发生日志滚动时会通知你，这样就可以做一些动作，例如把刚刚关闭的文件上传线上存储，或者对它做分析。

下面是代码:

```go
pathFormat := filepath.Join("dir", "2006-01-02.txt")
func onClose(path string, didRotate bool) {
	fmt.Printf("we just closed a file '%s', didRotate: %v\n", path, didRotate)
	if !didRotate {
		return
	}
	// process just closed file e.g. upload to backblaze storage for backup
	go func() {
		// if processing takes a long time, do it in a background goroutine
	}()
}

w, err := dailyrotate.NewFile(pathFormat, onClose)
panicIfErr(err)

_, err = io.WriteString(w, "hello\n")
panicIfErr(err)

err = w.Close()
panicIfErr(err)
```

## 长期存储以及分析

为了长期存储我把它们压缩成 gzip 并把文件上传到线上存储。这有很多选择：S3, Google Storage, Digital Ocean Spaces, BackBlaze。

我倾向于使用 Digital Ocean Spaces 或者 BackBlaze，因为他们足够廉价（存储成本和贷款成本）。

它们均支持 S3 协议，所以我使用 [go-minio](https://github.com/minio/minio-go) 库。

为了分析，我每天都会运行代码，生成大部分有用信息的总结。

还有其他的做法，可以把数据引入到如 [BigQuery](https://cloud.google.com/bigquery/what-is-bigquery) 的系统。

## 更多的 Go 资源

- [Essential Go](https://www.programming-books.io/essential/go/) 是由我所维护关于 Go ，免费且全面的书籍。
- [siser](https://github.com/kjk/siser) 是我写的库，实现了简单的序列化格式。
- 我还写了一篇有关 `siser` 设计的[深度文章](https://blog.kowalczyk.info/article/fc9203f7c72a4532b1ae51d018fef7b3/trade-offs-in-designing-versatile-log-format.html) 。
- [dailyrotate](https://github.com/kjk/dailyrotate) 是我写的库，实现了文件每日滚动。

## 招聘 Go 程序员

如果你正在寻找程序员一起工作，[希望一起谈一下](https://blog.kowalczyk.info/goconsultantforhire.html)。

由 [Krzysztof Kowalczyk](https://blog.kowalczyk.info/) 所著。

---

via: https://onepage.nopub.io/p/Logging-HTTP-requests-in-Go-233de7fe59a747078b35b82a1b035d36

作者：[Krzysztof Kowalczyk](https://onepage.nopub.io/u/bb760e2dd6794b64b2a903005b21870a)
译者：[LSivan](https://github.com/LSivan)
校对：[JYSDeveloper](https://github.com/JYSDeveloper)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
