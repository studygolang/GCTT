首发于：https://studygolang.com/articles/17943

# 如何发送和接收 SMS: 用 Go 语言实现 GSM 协议

当开发者出于验证或者通知的目的想要为应用程序添加 短消息服务 组件时，通常会使用像 [Twilio](https://www.twilio.com/docs/sms/api) 提供的 RESTful API，但是 API 之下到底发生了什么呢？

在这篇文章，您将了解 [通用计算机协议（UCP）](https://wiki.wireshark.org/UCP) 是什么以及如何使用 Go 语言通过这个协议直接与 [短消息服务中心（SMSC）](https://en.wikipedia.org/wiki/Short_Message_service_center) 通信来发送和接收 [SMS](https://en.wikipedia.org/wiki/SMS).

## 术语

### MT 信息

运营商发送给用户的短信息，例如天气更新信息

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-sms/mobile-terminating.png)

### MO 信息

用户发送给运营商的短消息，例如向一个指定号码发送关键字来查询余额信息

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-sms/mobile-originating.png)

### 超长 MT 消息和超长 MO 消息

起过 160 个字的 SMS 被视为 超长 SMS. 发送 超长 MT 消息 时，需要把它拆分成多个 信息片段。每个消息片段包含本片段的编号，整个消息的编号和一个引用编号。

超长 MO 消息的每个消息片段也包含本片段的编号，整个消息的编号和一个引用编号。我们需要把这些消息片段组合起来，以便解析用户发送的原始 超长 MO 短消息

## 通用计算机协议

通用计算机协议（UCP）主要用来连接 短消息服务中心（SMSC），发送和接收 SMS

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-sms/emiucp.png)

### session management operation

允许我们向 SMSC 发送登录信息

### alert operation

允许我们对 SMSC 发送 Ping

### submit short message operation

允许我们发送 MT 消息

### delivery notification operation

由 SMSC 发送给客户端，做为消息传输的状态凭证，标识之前发送的消息是否发送成功

### delivery short message operation

由 SMSC 发送给客户端，是对用户发送的 MO 消息 的响应

## 实现

我们可以把 UCP 看成一个传统的 客户端 - 服务器 协议。建立 TCP 连接后，我们发送包含 00 到 99 之间序列号（在[协议规范](http://documents.swisscom.com/product/1000174-Internet/Documents/Landingpage-Mobile-Mehrwertdienste/UCP_R4.7.pdf) 中称为“传输引用号”）的 UCP 请求，SMSC 会同步的返回一个 UCP 响应信息。SMSC 也可以发送 UCP 请求，比如 “ delivery notification operation ” 和 “ delivery short message operation ”。我们也需要定期的向 SMSC 发送 ping，以便它不会认为该连接过期而将其断开。

我们以 `Client` 类型开始，这个类型包含了向 SMSC 发送的登录信息。登录信息通常是由运营商提供的，但出于测试目的，我们可以使用 [SMSC 模拟器](https://github.com/jcaberio/ucp-smsc-sim)

```go
// Client represents a UCP client connection.
type Client struct {
	// IP:PORT address of the SMSC
	addr string
	// SMSC username
	user string
	// SMSC pasword
	password string
	// SMSC accesscode
	accessCode string
}

```

## 传输引用号

为了生成范围从 00 到 99 之间的合法传输引用号，我们可以使用标准库中的 [ring](https://golang.org/pkg/container/ring/) 包

```go
// Client represents a UCP client connection.
type Client struct {
	// skipped fields ...
	// ring counter for sequence numbers 00-99
	ringCounter *ring.Ring
}

const maxRefNum = 100

// INItRefNum INItializes the ringCounter counter from 00 to 99
func (c *Client) INItRefNum() {
	ringCounter := ring.New(maxRefNum)
	for i := 0; i < maxRefNum; i++ {
		ringCounter.Value = []byte(fmt.Sprintf("%02d", i))
		ringCounter = ringCounter.Next()
	}
	c.ringCounter = ringCounter
}

// nextRefNum returns the next transaction reference number
func (c *Client) nextRefNum() []byte {
	refNum := (c.ringCounter.Value).([]byte)
	c.ringCounter = c.ringCounter.Next()
	return refNum
}
```

## 建立 TCP 连接

我们可以使用 net 包与 SMSC 建立 TCP 连接。然后使用 bufio 包创建带缓冲的读写器

建立 TCP 连接后，我们就可以向 SMSC 发送一个 `session management operation` 请求。这个请求中包含发送给 SMSC 的登录信息。

```go
type Client struct {

	// skipped fields ....

	conn net.Conn
	reader *bufio.Reader
	writer *bufio.Writer

}

const etx = 3

func (c *Client) Connect() error {
	// INItialize ring counter from 00-99
	c.initRefNum()

	// establish TCP connection
	conn, _ := net.Dial("tcp", c.addr)
	c.conn = conn

	// create buffered reader and writer
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)

	// login to SMSC
	c.writer.Write(createLoginReq(c.nextRefNum(), c.user, c.password))
	c.writer.Flush()
	resp, _ := c.reader.ReadString(etx)
	err = parseSessionResp(resp)
	// ....other processing....
	return err
}
```

函数 `createLoginReq` 创建了一个包含登录信息的 `session management operation` 请求数据包。函数 `parseSessionResp` 解析 SMSC 对这个 `session management operation` 返回的响应数据包。如果我们发送的登录信息是正确的，此函数返回 nil ，否则返回 error.

## 通道和 Goroutines

我们可以为将不同的 UCP 操作视为单独的 Gorutine 和 通道 .

```go
type Client struct {
	// skipped fields ....
	// channel for handling submit short message responses from SMSC
	submitSmRespCh chan []string
	// channel for handling delivery notification requests from SMSC
	deliverNotifCh chan []string
	// channel for handling delivery message requests from SMSC
	deliverMsgCh chan []string
	// channel for handling incomplete delivery message from SMSC
	deliverMsgPartCh chan deliverMsgPart
	// channel for handling complete delivery message requests from SMSC
	deliverMsgCompleteCh chan deliverMsgPart
	// we close this channel to signal Goroutine termination
	closeChan chan struct{}
	// waitgroup for the running Goroutines
	wg *sync.WaitGroup
	// guard against closing closeChan multiple times
	once sync.Once
}

// Connect will establish a TCP connection with the SMSC
// and send a login request.
func (c *Client) Connect() error {
	// after login, spawn Goroutines
	sendAlert(/*....*/)
	readLoop(/*....*/)
	readDeliveryNotif(/*....*/)
	readDeliveryMsg(/*....*/)
	readPartialDeliveryMsg(/*....*/)
	readCompleteDeliveryMsg(/*....*/)
	return err
}

// Close will close the UCP connection.
// It's safe to call Close multiple times.
func (c *Client) Close() {
	// close closeChan to terminate the spawned Goroutines
	// we use a sync.Once to close closeChan only once.
	c.once.Do(func() {
		close(c.closeChan)
	})
	// close the underlying TCP connection
	if c.conn != nil {
		c.conn.Close()
	}
	// wait for all Goroutines to terminate
	c.wg.Wait()
}
```

## 读取 UCP 数据包

我们通过 `readLoop` 从 UCP 连接读取数据包。合法的 UCP 数据包是以 [文件结束符分隔（ETX）](https://en.wikipedia.org/wiki/End-of-Text_character) 分隔的。对应的字节码是 `03` `readLoop` 会一直读取直到发现 `etx`，然后解析读取到的信息，并将其发送到相应的通道。

```go
// readLoop reads incoming messages from the SMSC
// using the underlying bufio.Reader
func readLoop(/*.....*/) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-closeChan:
				return
			default:
				readData, _ := reader.ReadString(etx)
				opType, fields, _ := parseResp(readData)
				switch opType {
				case opSubmitShortMessage:
					submitSmRespCh <- fields
				case opDeliveryNotification:
					deliverNotifCh <- fields
				case opDeliveryShortMessage:
					deliverMsgCh <- fields
				}
			}
		}
	}()
}
```
## 发送 Keepalive

`sendAlert` 会向 SMSC 定期发送 ping，我们用 time.NewTicker 创建了一个定期触发的定时器。`createAlertReq` 创建了一个包含合法传输引用号的 `alert operation` 请求数据包

```go
// sendAlert sends a keepalive packet periodically to the SMSC
func sendAlert(/*....*/) {
	wg.Add(1)
	ticker := time.NewTicker(alertInterval)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-closeChan:
				ticker.Stop()
				return
			case <-ticker.C:
				writer.Write(createAlertReq(transRefNum, user))
				writer.Flush()
			}
		}
	}()
}
```

## 读取传递通知状态

`readDeliveryNotif` 用来读取 SMS 的传递通知状态。每读到一个 `delivery notification operation` 就会向 SMSC 发送一个确认数据包。

```go
// readDeliveryNotif reads delivery notifications from deliverNotifCh channel.
func readDeliveryNotif(/*....*/) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-closeChan:
				return
			case dr := <-deliverNotifCh:
				refNum := dr[refNumIndex]
				// msg contains the complete delivery status report from the SMSC
				msg, _ := hex.DecodeString(dr[drMsgIndex])
				// sender is the access code of the SMSC
				sender := dr[drSenderIndex]
				// recvr is the mobile number of the recipient subscriber
				recvr := dr[drRecvrIndex]
				// scts is the service center time stamp
				scts := dr[drSctsIndex]
				msgID := recvr + ":" + scts
				// send ack to SMSC
				writer.Write(createDeliveryNotifAck([]byte(refNum), msgID))
				writer.Flush()
			}
		}
	}()
}
```

## 读取传递短消息

`readDeliveryMsg` 用来读取 MO 消息。

```go
// readDeliveryMsg reads all delivery short messages
// (mobile-originating messages) from the deliverMsgCh channel.
func readDeliveryMsg(/*....*/) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-closeChan:
				return
			case mo := <-deliverMsgCh:
				xser := mo[xserIndex]
				xserData := parseXser(xser)
				msg := mo[moMsgIndex]
				refNum := mo[refNumIndex]
				sender := mo[moSenderIndex]
				recvr := mo[moRecvrIndex]
				scts := mo[moSctsIndex]
				sysmsg := recvr + ":" + scts
				msgID := sender + ":" + scts

				// send ack to SMSC with the same reference number
				writer.Write(createDeliverySmAck([]byte(refNum), sysmsg))
				writer.Flush()
				var incomingMsg deliverMsgPart
				incomingMsg.sender = sender
				incomingMsg.receiver = recvr
				incomingMsg.message = msg
				incomingMsg.msgID = msgID
				// further processing
			}
		}
	}()
}
```

类型 `deliverMsgPart` 包含了用来连接和解码收到的 超长 MO 消息片段所需要的必要信息。

```go
// deliverMsgPart represents a deliver sm message part
type deliverMsgPart struct {
	currentPart int
	totalParts  int
	refNum      int
	sender      string
	receiver    string
	message     string
	msgID       string
	dcs         string
}
```

为了处理 超长 MO 信息，我们把 每个消息片段 发送到通道 `deliverMsgPartCh` 上，把 MO 消息发送到通道 `deliverMsgCompleteCh` 上。

```go
// readDeliveryMsg reads all delivery short messages
// (mobile-originating messages) from the deliverMsgCh channel.
func readDeliveryMsg(/*....*/) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-closeChan:
				return
			case mo := <-deliverMsgCh:
				// INItial processing ......
				if xserUdh, ok := xserData[udhXserKey]; ok {
					// handle multi-part mobile originating message
					// get the total message parts in the xser data
					msgPartsLen := xserUdh[len(xserUdh)-4 : len(xserUdh)-2]
					// get the current message part in the xser data
					msgPart := xserUdh[len(xserUdh)-2:]
					// get message part reference number
					msgRefNum := xserUdh[len(xserUdh)-6 : len(xserUdh)-4]
					// convert hexstring to integer
					msgRefNumInt, _ := strconv.ParseInt(msgRefNum, 16, 0)
					msgPartsLenInt, _ := strconv.ParseInt(msgPartsLen, 16, 64)
					msgPartInt, _ := strconv.ParseInt(msgPart, 16, 64)
					incomingMsg.currentPart = int(msgPartInt)
					incomingMsg.totalParts = int(msgPartsLenInt)
					incomingMsg.refNum = int(msgRefNumInt)
					// send to partial channel
					deliverMsgPartCh <- incomingMsg
				} else {
					// handle mobile originating message with only 1 part
					// send the incoming message to the complete channel
					deliverMsgCompleteCh <- incomingMsg
				}
			}
		}
	}()
}
```

函数 `readPartialDeliveryMsg` 中启动的 Goroutine 会从通道 `deliverMsgPartCh` 中读取消息，然后把消息片段合并成完整的 超长 MO 消息。函数 `readCompleteDeliveryMsg` 中启动的 Goroutine 会从通道 `deliverMsgCompleteCh` 读取 MO 消息，并执行相应的回调函数。

## 发送 SMS

我们用 `Send` 来发 SMS.

```go
// Send will send the message to the receiver with a sender mask.
// It returns a list of message IDs from the SMSC.
func (c *Client) Send(sender, receiver, message string) ([]string, error) {
	msgType := getMessageType(message)
	msgParts := getMessageParts(message)
	refNum := rand.Intn(maxRefNum)
	ids := make([]string, len(msgParts))
	for i := 0; i < len(msgParts); i++ {
		sendPacket := encodeMessage(c.nextRefNum(), sender, receiver, msgParts[i], msgType,
			c.GetBillingID(), refNum, i+1, len(msgParts))
		c.writer.Write(sendPacket)
		c.writer.Flush()
		select {
		case fields := <-c.submitSmRespCh:
			ack := fields[ackIndex]
			if ack == negativeAck {
				errMsg := fields[len(fields)-errMsgOffset]
				errCode := fields[len(fields)-errCodeOffset]
				return ids, &UcpError{errCode, errMsg}
			}
			id := fields[submitSmIdIndex]
			ids[i] = id
		case <-time.After(c.timeout):
			return ids, &UcpError{errCodeTimeout, "Network time-out"}
		}
	}
	return ids, nil
}
```

`getMessageType` 确定消息包含的是普通 GSM-7 格式的字符还是 Unicode 字符

`getMessageParts` 把 超长 SMS 拆分成多个消息片段

`encodeMessage` 负责创建包含适当引用号的合法 `submit short message orperation` 数据包，把 unicode 格式的消息转化为 [UCS2](https://en.wikipedia.org/wiki/Universal_Coded_Character_Set) 格式，对发送者名字进行加密。

我们使用 `select`  语句从从 SMSC 获得响数据包。 它会处于阻塞状态，直到通道 `submitSmRespCh` 变成可读或者发生了超时

`Send` 返回一个消息标识符的列表，表明 SMSC 成功接收到了 `submit short message operation` 请求。数据是同步返回的。例如，如果我们发送了一个包含 5 个消息片段的 超长 MO 消息，`Send` 就会返回一个包含 5 个字符串的列表

```
[09191234567:130817221851, 09191234567:130817221852, 09191234567:130817221853, 09191234567:130817221854, 09191234567:130817221855]
```

每个标识符有如下的格式 `recipient:timestamp`。`timestamp` 部分可以使用 `020106150405` 这样的格式，用 [time.Parse](https://golang.org/pkg/time/#Parse) 来解析。如果你更熟悉 [strftime](http://strftime.org/)， `timestamp` 也可以使用 `%d%m%y%H%M%S` 这样的格式。

## 示例

我写了一个简单的项目 [CLI](https://github.com/go-gsm/ucp-cli) 来演示这个库，我们使用 [SMSC simulator](https://github.com/jcaberio/ucp-smsc-sim) 当做短消息中心，通过 [Wireshark](https://www.wireshark.org/) 查看 UCP 数据包

首先，通过 `go get` 获取 CLI 和 SMSC 模拟器，并且确保 [redis](https://redis.io/) 运行在地址 `localhost:6379` 上

```
$ go get GitHub.com/go-gsm/ucp-cli
$ go get GitHub.com/jcaberio/ucp-smsc-sim
```

导出以下环境变量

```
$ export SMSC_HOST=127.0.0.1
$ export SMSC_PORT=16004
$ export SMSC_USER=emi_client
$ export SMSC_PASSWORD=password
$ export SMSC_ACCESSCODE=2929
```

运行 SMSC 模拟器，在浏览器中访问 `localhost:16003`

```
$ ucp-smsc-sim
```

运行 CLI

```
$ ucp-cli
```

我们用 `Gopher` 向 `09191234567` 发送一条消息 `Hello, 世界`。模拟器会返回包含 `[09191234567:021218201629]` 的响应。我们还可以从模拟器中看到传递通知信息。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-sms/send-via-cli.png)

我们可以通过 Wireshark 查看具体的 UCP 数据包

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-sms/wireshark.png)

我们可以在浏览器中查看 SMS

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-sms/simulator.png)

为了模仿用户发送的 MO 信息，我们可以发送以下 `curl` 请求

```
curl -H "Content-Type: application/json" -d '{"sender":"09191234567", "receiver":"2929", "message":"This is a mobile-originating message"}' http://localhost:16003/mo
```

我们模仿的是一个号码为 `09191234567` 的用户向 `2929` 发送了以下的信息 `This is a mobile-originating message`

我们可以看到 CLI 接收到了这各 MO 信息，并且在 Wireshark 得到了验证

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-sms/recv-via-cli.png)
![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-sms/recv-wireshark.png)

## 总结

Go 语言中内置的一些特性，比如 Goroutine 和 通道 让我们可以方便的实现 UCP 协议。我们用 Go 语言的消息处理方式，以并发的方式处理不同类型的 UCP 消息。我们用不同的 Goroutine 来代表不同的 UCP 操作，并通过通道与之通信。在实现各种协议操作时我们也大量的使用的标准库。如果你在电信领域工作，并且可以访问到 SMSC，可以尝试使用 [ucp](https://github.com/go-gsm/ucp) 包，它包含额外的一些功能，比如速率限制和收费管理。欢迎提出您的宝贵建议。

谢谢

---

via: https://blog.gopheracademy.com/advent-2018/how-to-send-and-receive-sms/

作者：[Jorick Caberio](https://blog.gopheracademy.com/advent-2018/how-to-send-and-receive-sms/)
译者：[jettyhan](https://github.com/jettyhan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
