首发于：https://studygolang.com/articles/17397

# 使用 Golang 构建 DNS 服务器

需求：对 DNS 查询进行转发和缓存的本地 DNS 服务器。

补充 1：提供一个记录管理的接口（HTTP handler）。

补充 2：提供一个名字（name）。

DNS 服务器的相关要点如下：

- DNS 服务器把域名转换为 IP。
- DNS 主要使用 UDP 协议，其端口为 53。
- DNS 消息的长度最多为 512 字节，若超过这个长度，则必须使用 EDNS。

需要的组成部分有：

- UDP
- DNS 消息解析器（DNS message parser）
- 转发
- 缓存
- HTTP handler

我们的解决方案是：

- UDP：标准包 `net` 支持 UDP。
- DNS 消息解析器：需要一些工作，来根据特定协议（UDP）的通信，处理报文。为了更快地实现，我们使用 `golang.org/x/net/dns/dnsmessage`。
- 转发：实现方式有很多，我们使用了 Cloudflare 公共解析器（Cloudflare public resolver）：1.1.1.1。
- 缓存：持久性存储。为了持久化写入数据，我们使用标准包 `gob` 来编码数据。
- HTTP handler：应该能够添加、查询、更新和删除 DNS 记录。不需要使用配置文件。

开启 UDP socket，监听 53 端口，可以接收 DNS 查询。需要注意的是，UDP 只需要一个 socket 来处理多条“连接”，而 TCP 对于每条连接都需要一个 socket。因此，我们在程序中，会重复使用 `conn`。

```go
conn, _ = net.ListenUDP("udp", &net.UDPAddr{Port: 53})
defer conn.Close()
for {
    buf := make([]byte, 512)
    _, addr, _ := conn.ReadFromUDP(buf)
    ...
}
```

解析报文，检查是否是 DNS 消息。

```go
var m dnsmessage.Message
err = m.Unpack(buf)
```

如果你想知道一条 DNS 消息长什么样，请查看下图：

![a DNS message](https://raw.githubusercontent.com/studygolang/gctt-images/master/build-dns-server/1.jpg)

## 转发消息到公共解析器

```go
// re-pack
packed, err = m.Pack()
resolver := net.UDPAddr{IP: net.IP{1, 1, 1, 1}, Port: 53}
_, err = conn.WriteToUDP(packed, &resolver)
```

公共解析器会返回一条 anwser，我们会抓取信息，返回给客户端。

```go
if m.Header.Response {
    packed, err = m.Pack()
    _, err = conn.WriteToUDP(packed, &addr)
}
```

当然并发使用 `conn` 很安全，所以 `WriteToUDP` 应该在 Go 协程中运行。

## 存储 answer

我们会使用 map，简单采用“ question-anwser ”的键值对，这会让查询变得很容易。同样不要忘了 `RWMutex`，对于并发操作，map 使用起来并不安全。需要提醒的是，从理论上讲，在一次 DNS 查询中，可能会有多个 question，但是大多数 DNS 服务器，都只会接收一条 question。

```go
func questionToString(q dnsmessage.Question) string {
    ...
}
type store struct {
    sync.RWMutex
    data      map[string][]dnsmessage.Resource
}
q := m.Questions[0]
var s store
s.Lock()
s.data[questionToString(q)] = m.Answers
s.Unlock()
```

## 持久化缓存（persistent cache）

我们需要把 `s.data` 写入到文件中，以便以后重新获取它。我们使用了标准包 `gob`，而无需自定义解析。

```go
f, err := os.Create(filepath.Join("path", "file"))
enc := Gob.NewEncoder(f)
err = enc.Encode(s.data)
```

需要注意，**gob** 在编码前需要知道数据类型。

```go
func INIt() {
    Gob.Register(&dnsmessage.AResource{})
    ...
}
```

## 记录管理

这个相对来说就比较简单了，Create handler 如下所示：

```go
type request struct {
    Host string
    TTL  uint32
    Type string
    Data string
}
func toResource(req request) (dnsmessage.Resource, error) {
    ...
}
// POST handler
err = JSON.NewDecoder(r.Body).Decode(&req)
// transform req to a dnsmessage.Resource
r, err := toResource(req)
// write r to the store
```

理所当然是吧？

完整的代码在[这里](https://github.com/owlwalks/rind)。我将其命名为 **rind**（REST interface name domain）。

以上就是用 Go 实现这个网络程序的简述。

欢迎任何反馈。编程愉快，Gopher ！

---

via: https://medium.com/@owlwalks/build-a-dns-server-in-golang-fec346c42889

作者：[Khoa Pham](https://medium.com/@owlwalks)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
