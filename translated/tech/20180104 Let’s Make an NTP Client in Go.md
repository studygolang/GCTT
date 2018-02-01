# Let's Make an NTP Client in Go

在网络编程做了一些研究之后，我邂逅了一篇题目为《Let's Make a NTP Client in C》，由David Lettier(Lettier)编写的文章。这篇文章鼓舞了我用Go去做相似的事。

```
这篇博文提到的代码都在这里 [https://github.com/vladimirvivien/go-ntp-client](https://github.com/vladimirvivien/go-ntp-client.)
```

这篇博文描述了一个（真正的）NTP客户端的结构，使用Go语言编写。它通过　encoding/binary库去封装，解封装，发送和接收来自远端NTP服务器基于UDP协议的NTP包。

你能通过[这里](http://www.ntp.org/)学到更多关于NTP协议的内容，或者阅读 RFC5905规范、研究一个实现了更多的功能，（可能）比 Go NTP客户端更好的客户端 [github.com/beevik/ntp](github.com/beevik/ntp)。

## NTP 包结构

时间同步的概念是非常复杂的，我还不能完全理解，也超过这篇博文的范围。但幸运的是，NTP使用的数据包格式很简单，对于客户端来说也小而足够了。下面的图展示了NTP v4的包格式。关于这篇博文，我们只关注前48个字节，忽略掉v4版本的扩展部分．

![NTP v4 data format (abbreviated) — https://tools.ietf.org/html/rfc5905](https://github.com/gogeof/gctt-images/blob/master/go-ntp/NTP-v4-data-format.png)
NTP v4 data format (abbreviated) — https://tools.ietf.org/html/rfc5905

## NTP 包

客户端和对应的服务端都使用上面提到的相同的包格式。下面的结构体定义了NTP包和它的属性，跟上面提到的格式一一对应。

```go
type packet struct {
    Settings       uint8  // leap yr indicator, ver number, and mode
    Stratum        uint8  // stratum of local clock
    Poll           int8   // poll exponent
    Precision      int8   // precision exponent
    RootDelay      uint32 // root delay
    RootDispersion uint32 // root dispersion
    ReferenceID    uint32 // reference id
    RefTimeSec     uint32 // reference timestamp sec
    RefTimeFrac    uint32 // reference timestamp fractional
    OrigTimeSec    uint32 // origin time secs
    OrigTimeFrac   uint32 // origin time fractional
    RxTimeSec      uint32 // receive time secs
    RxTimeFrac     uint32 // receive time frac
    TxTimeSec      uint32 // transmit time secs
    TxTimeFrac     uint32 // transmit time frac
}
```

## 启动UDP连接

接下来，我们通过UDP协议，使用　net.Dial函数去启动一个socket，与NTP服务器联系，并设定15秒的超时时间．

```
conn, err := net.Dial("udp", host)
if err != nil {
    log.Fatal("failed to connect:", err)
}
defer conn.Close()
if err := conn.SetDeadline(
  time.Now().Add(15 * time.Second)); err != nil {
    log.Fatal("failed to set deadline: ", err)
}
```

## 从服务端获取时间

在发送请求包给服务端前，第一个字节是用来设置通信的配置，我们这里用0x1B(或者二进制 00011011)，代表客户端模式为3，NTP版本为3，润年为0，如下所示：

```go
// configure request settings by specifying the first byte as
// 00 011 011 (or 0x1B)
// |  |   +-- client mode (3)
// |  + ----- version (3)
// + -------- leap year indicator, 0 no warning
req := &packet{Settings: 0x1B}
```

接下来，我们使用 binary库去自动地将packet结构体封装成字节流，并以大端格式发送出去。

```go
if err := binary.Write(conn, binary.BigEndian, req); err != nil {
    log.Fatalf("failed to send request: %v", err)
}
```

## 从服务端读取时间

接下来，我们使用binary包再次将从服务端读取的字节流自动地解封装成对应的packet结构体。

```go
rsp := &packet{}
if err := binary.Read(conn, binary.BigEndian, rsp); err != nil {
    log.Fatalf("failed to read server response: %v", err)
}
```

## 解析时间

在这个超平凡的例子里面，我们只对　Transmit Time字段(rsp.TxTimeSec和rspTxTimeFrac)感兴趣，它们是从服务端发出时的时间。但我们不能直接使用它们，必须先转成Unix时间．

Unix时间是一个开始于1970年的纪元（或者说从1970年开始的秒数）。然而NTP使用的是另外一个纪元，从1900年开始的秒数。因此，从NTP服务端获取到的值要正确地转成Unix时间必须减掉这70年间的秒数(1970-1900)，或者说 2208988800秒

```go
const ntpEpochOffset = 2208988800
...
secs := float64(rsp.TxTimeSec) - ntpEpochOffset
nanos := (int64(rsp.TxTimeFrac) * 1e9) >> 32
```

NTP值的分数部分转成纳秒。在这个平凡的例子里，这里是可选的，展示只是为了完整性．

## 显示时间

最后，函数time.Unix被用来创建一个秒数部分使用secs，分数部分使用nanos值的时间。然后这个时间会被打印到终端．

```go
fmt.Printf("%v\n", time.Unix(int64(secs), nanos))
```

## 结论

这篇博文展示了一个关于NTP客户端的普通的例子。描述了如何利用 encoding/binary库，非常容易地将一个结构体转成字节形式。相反，我们使用binary库将一个字节流转成对应的结构体值．

这个 NTP客户端还不是一个可用于生产环境的产品，毕竟它缺少了 NTP规范指定的很多功能。从服务端返回的大部分字段都被忽略了。你可以从[这里](https://github.com/beevik/ntp)获取到一个用Go写的更完整的NTP客户端。



via: https://medium.com/learning-the-go-programming-language/lets-make-an-ntp-client-in-go-287c4b9a969f

 作者：[Vladimir Vivien](https://twitter.com/VladimirVivien)
 译者：[gogeof](https://github.com/gogeof)
 校对：

 本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

