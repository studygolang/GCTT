已发布：https://studygolang.com/articles/12619

# Go 语言神奇的 JSON

今天我想和大家分享 Go 语言一些非常实用的技巧，用于编码和解码 JSON 文档。Go 语言的  `encoding/json` 包有一些有趣的特性，帮助我们轻松地解析 JSON 文档。你可以轻松地将大多数实际应用中的 JSON 转换为带有 Go 语言结构体标签的接口或者是 `Marshaler` 和 `Unmarshaler` 接口。

但有一个案例比较棘手：包含转义 JSON 元素的 JSON 文档。如下所示：

```json
{
	"id": 12345,
	"name": "Test Document",
	"payload": "{\"message\":\"hello!\"}"
}
```

我不建议构建像这样创建文档的应用程序，但有时候这样的情况是难以避免的，你希望像平常的 JSON 那样，一步就能解析这个文档。也许你从如下两种类型开始：

```go
type LogEntry struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Payload string `json:"payload"`
}
type LogPayload struct {
	Message string `json:"message"`
}
```

Matt Holt 的 [*json-to-go*](https://mholt.github.io/json-to-go/) 能够帮助你从 JSON 示例中生成初始结构体，不妨试一下！

首先要将 `LogEntry.Payload` 的类型从 `string` 类型改为 `LogPayload` 类型。这点很重要，因为这是你最终想要得到的，这就是 `encoding/json` 包处理该元素的方式。现在的问题是 `payload` 元素的实际入站类型是一个 JSON 字符串。你需要在 `LogPayload` 类型上实现 `Unmarshaler` 接口，并将其解码为字符串，然后再解码为 `LogPayload` 类型。

```go
func (lp *LogPayload) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(s), lp); err != nil {
		return err
	}
 
	return nil
}
```

看起来很棒，然而不幸的是第二个 `json.Unmarshal` 调用将会导致调用堆栈的递归。你需要将它解码成一个中间类型，你可以通过定义一个带有 `LogPayload` 基础类型的新类型来实现，例如这样：

```go
type fauxLogPayload LogPayload
```

你可以将上面的代码调整一下，将其解码为 `fauxLogPayload` 类型，然后将结果转换为 `LogPayload` 类型。


```go
func (lp *LogPayload) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	var f fauxLogPayload
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
 
	*lp = LogPayload(f)
 
	return nil
}
```


现在，要解析整个文档的调用站点变得更好了，也简洁了:

```go
func main() {
	doc := []byte(`{
		"id": 12345,
		"name": "Test Document",
		"payload": "{\"message\":\"test\"}"
	}`)
	var entry LogEntry
	if err := json.Unmarshal(doc, &entry); err != nil {
		fmt.Println("Error!", err)
	}
	fmt.Printf("%v", entry)
}
```

你可以在 [*Go Playground*](https://play.golang.org/p/8l4K4GCF--U) 找到这些代码。

我希望这个例子说明了 Go 语言可以多么容易地将对  `encoding/decoding` 的关注点从业务逻辑中分离出来。你可以在任何时候使用此方法将基本 JSON 类型转换为更复杂的用户定义类型。

Cheers！

感谢 [Redditors BubuX](https://www.reddit.com/r/golang/comments/801c4i/json_in_go_is_magical/dusgzny/) 和 [quiI](https://www.reddit.com/r/golang/comments/801c4i/json_in_go_is_magical/duso6pc/) ，他们建议链接到 *JSON -to- Go* ，并在 `main.go` 中为我的 JSON 使用 Go 语言的字符串文字。

----------------

via: https://medium.com/@turgon/json-in-go-is-magical-c5b71505a937

作者：[turgon](https://medium.com/@turgon)
译者：[SergeyChang](https://github.com/SergeyChang)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出


