已发布：https://studygolang.com/articles/12172

# Go 语言中的动态 JSON

Go 语言是静态类型语言，虽然它也可以表现出动态类型，但是使用一个嵌套的 `map[string]interface{}` 在那里乱叫会让代码变得特别丑。通过掌握语言的静态特性，我们可以做的更好。

通过同一通道交换多种信息的时候，我们经常需要 JSON 具有动态的，或者更合适的参数内容。首先，让我们来讨论一下消息封装(message envelopes)，JSON 在这里看起来就像这样：

```json
{
	"type": "this part tells you how to interpret the message",
	"msg": ...the actual message is here, in some kind of json...
}
```

## 通过不同的消息类型生成 JSON

通过 `interface{}`，我们可以很容易的将数据结构编码成为独立封装的，具有多种类型的消息体的 JSON 数据。为了生成下面的 JSON ：

```json
{
	"type": "sound",
	"msg": {
		"description": "dynamite",
		"authority": "the Bruce Dickinson"
	}
}
```

```json
{
	"type": "cowbell",
	"msg": {
		"more": true
	}
}
```
我们可以使用这些 Go 类型：

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type Envelope struct {
	Type string
	Msg  interface{}
}

type Sound struct {
	Description string
	Authority   string
}

type Cowbell struct {
	More bool
}

func main() {
	s := Envelope{
		Type: "sound",
		Msg: Sound{
			Description: "dynamite",
			Authority:   "the Bruce Dickinson",
		},
	}
	buf, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf)

	c := Envelope{
		Type: "cowbell",
		Msg: Cowbell{
			More: true,
		},
	}
	buf, err = json.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf)
}
```
输出的结果是：

```json
{"Type":"sound","Msg":{"Description":"dynamite","Authority":"the Bruce Dickinson"}}
{"Type":"cowbell","Msg":{"More":true}}
```
这些并没有什么特殊的。

## 解析 JSON 到动态类型

如果你想将上面的 JSON 对象解析成为一个 `Envelope` 类型的对象，最终你会将 `Msg` 字段解析成为一个 `map[string]interface{}`。 这种方式不是很好用，会使你后悔你的选择。

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
)

const input = `
{
	"type": "sound",
	"msg": {
		"description": "dynamite",
		"authority": "the Bruce Dickinson"
	}
}
`

type Envelope struct {
	Type string
	Msg  interface{}
}

func main() {
	var env Envelope
	if err := json.Unmarshal([]byte(input), &env); err != nil {
		log.Fatal(err)
	}
	// for the love of Gopher DO NOT DO THIS
	var desc string = env.Msg.(map[string]interface{})["description"].(string)
	fmt.Println(desc)
}
```
输出：

```
dynamite
```

## 明确的解析方式

就像前面说的，我推荐修改 `Envelope` 类型，就像这样：

```go
type Envelope {
	Type string
	Msg  *json.RawMessage
}
```
[`json.RawMessage`](http://golang.org/pkg/encoding/json/#RawMessage) 非常有用，它可以让你延迟解析相应的 JSON 数据。它会将未处理的数据存储为 `[]byte`。

这种方式可以让你显式控制 `Msg` 的解析。从而延迟到获取到 `Type` 的值之后，依据 `Type` 的值进行解析。这种方式不好的地方在于你需要先明确解析 `Msg`，或者你需要单独分为 `EnvelopeIn` 和 `EnvelopeOut` 两种类型，其中 `EnvelopeOut` 仍然有 `Msg interface{}`。

## 结合 `*json.RawMessage` 和 `interface{}` 的优点

那么如何将上述两者好的一面结合起来呢？通过在 `interface{}` 字段中放入 `*json.RawMessage`!

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
)

const input = `
{
	"type": "sound",
	"msg": {
		"description": "dynamite",
		"authority": "the Bruce Dickinson"
	}
}
`

type Envelope struct {
	Type string
	Msg  interface{}
}

type Sound struct {
	Description string
	Authority   string
}

func main() {
	var msg json.RawMessage
	env := Envelope{
		Msg: &msg,
	}
	if err := json.Unmarshal([]byte(input), &env); err != nil {
		log.Fatal(err)
	}
	switch env.Type {
	case "sound":
		var s Sound
		if err := json.Unmarshal(msg, &s); err != nil {
			log.Fatal(err)
		}
		var desc string = s.Description
		fmt.Println(desc)
	default:
		log.Fatalf("unknown message type: %q", env.Type)
	}
}
```
输出：

```
dynamite
```

## 如何把所有数据都放在最外层（顶层）

虽然我极其推荐你将动态可变的部分放在一个单独的 key 下面，但是有时你可能需要处理一些预先存在的数据，它们并没有用这样的方式进行格式化。

如果可以的话，请使用文章前面提到的风格。

```json
{
	"type": "this part tells you how to interpret the message",
	...the actual message is here, as multiple keys...
}
```
我们可以通过解析两次数据的方式来解决。

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
)

const input = `
{
	"type": "sound",
	"description": "dynamite",
	"authority": "the Bruce Dickinson"
}
`

type Envelope struct {
	Type string
}

type Sound struct {
	Description string
	Authority   string
}

func main() {
	var env Envelope
	buf := []byte(input)
	if err := json.Unmarshal(buf, &env); err != nil {
		log.Fatal(err)
	}
	switch env.Type {
	case "sound":
		var s struct {
			Envelope
			Sound
		}
		if err := json.Unmarshal(buf, &s); err != nil {
			log.Fatal(err)
		}
		var desc string = s.Description
		fmt.Println(desc)
	default:
		log.Fatalf("unknown message type: %q", env.Type)
	}
}
```

```
dynamite
```

---

via: http://eagain.net/articles/go-dynamic-json/

作者：[Tommi Virtanen](http://eagain.net/about/)
译者：[jliu666](https://github.com/jliu666)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出