# 使用 JSON 协议的 gRPC

JSON payload 实现简易的请求和响应的内省。

## 介绍

大家经常说 gRPC 是基于 [Google Protocol Buffers](https://developers.google.com/protocol-buffers/) payload 格式的，然而这不完全正确。gRPC payload 的*默认*格式是 Protobuf，但是 gRPC-Go 的实现中也对外暴露了 `Codec` [interface](https://godoc.org/google.golang.org/grpc/encoding#Codec) ，它支持任意的 payload 编码。我们可以使用任何一种格式，包括你自己定义的二进制格式、[flatbuffers](https://grpc.io/blog/flatbuffers)、或者使用我们今天要讨论的 JSON ，作为请求和响应。

## 服务端准备

我已经基于 JSON payload [实现](https://github.com/johanbrandhorst/grpc-json-example/blob/master/codec/json.go) 了 `grpc/encoding.Codec`，创建了[一个示例库](https://github.com/johanbrandhorst/grpc-json-example)。服务端的准备工作仅仅像引入一个包那样简单；

```go
import _ "github.com/johanbrandhorst/grpc-json-example/codec"
```

这行代码注册了一个基于 `json` 内容的子类型 JSON `Codec`，我们在后面会看到这对于方便记忆很重要。

## Request 示例

### gRPC 客户端

使用 gRPC 客户端，你只需要使用合适的内容子类型作为 `grpc.DialOption` 来初始化：

```go
import "github.com/johanbrandhorst/grpc-json-example/codec"
func main() {
    conn := grpc.Dial("localhost:1000",
        grpc.WithDefaultCallOptions(grpc.CallContentSubtype(codec.JSON{}.Name())),
    )
}
```

示例库代码包含有完整示例的[客户端](https://github.com/johanbrandhorst/grpc-json-example/blob/master/cmd/client/main.go)。

### cURL

更有趣的是，现在我们可以用 cURL 写出请求（和读取响应）！请求示例：

```bash
$ echo -en '\x00\x00\x00\x00\x17{"id":1,"role":"ADMIN"}' | curl -ss -k --http2 \
        -H "Content-Type: application/grpc+json" \
        -H "TE:trailers" \
        --data-binary @- \
        https://localhost:10000/example.UserService/AddUser | od -bc
0000000 000 000 000 000 002 173 175
         \0  \0  \0  \0 002   {   }
0000007
$ echo -en '\x00\x00\x00\x00\x17{"id":2,"role":"GUEST"}' | curl -ss -k --http2 \
        -H "Content-Type: application/grpc+json" \
        -H "TE:trailers" \
        --data-binary @- \
        https://localhost:10000/example.UserService/AddUser | od -bc
0000000 000 000 000 000 002 173 175
         \0  \0  \0  \0 002   {   }
0000007
$ echo -en '\x00\x00\x00\x00\x02{}' | curl -k --http2 \
        -H "Content-Type: application/grpc+json" \
        -H "TE:trailers" \
        --data-binary @- \
        --output - \
        https://localhost:10000/example.UserService/ListUsers
F{"id":1,"role":"ADMIN","create_date":"2018-07-21T20:18:21.961080119Z"}F{"id":2,"role":"GUEST","create_date":"2018-07-21T20:18:29.225624852Z"}
```

#### 解释

使用 `cURL` 发送请求需要手动把  [gRPC HTTP2 message payload header](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#requests) 加到 payload：

```bash
'\x00\x00\x00\x00\x17{"id":1,"role":"ADMIN"}'
#<-->----------------------------------------- Compression boolean (1 byte)
#    <-------------->------------------------- Payload size (4 bytes)
#                    <--------------------->-- JSON payload
```

请求头必须包含 `TE` 和正确的 `Content-Type`：

```bash
-H "Content-Type: application/grpc+json" -H "TE:trailers"
```

在 `Content-Type` 头中 `application/grpc+` 后的字符串需要与服务端注册的 codec 的 `Name()` 相吻合。这就是*内容子类*.

endpoint 需要与 proto 包的名字、服务和方法三者的名字都匹配：

```bash
https://localhost:10000/example.UserService/AddUser
```

响应头与请求头一致：

```bash
'\0  \0  \0  \0 002   {   }'
#<-->------------------------ Compression boolean (1 byte)
#    <------------>---------- Payload size (4 bytes)
#                     <--->-- JSON payload
```

## 总结

我们已经展示了我们可以轻易地在 gRPC 中使用 JSON payload，甚至可以用 JSON payload 直接发送 cURL 请求到我们的 gRPC 服务，没有代理，没有 grpc 网关，除了引入一个必要的包也没有其他的准备工作。

如果你对本文感兴趣，或者有任何问题和想法，请在 [@johanbrandhorst](https://twitter.com/JohanBrandhorst) 上或 在Gophers Slack `jbrandhorst`下联系我。很高兴听到你的想法。

---
via: https://jbrandhorst.com/post/grpc-json/

作者：[Johan Brandhorst](https://jbrandhorst.com/)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
