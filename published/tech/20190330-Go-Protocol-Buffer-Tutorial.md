首发于：https://studygolang.com/articles/22294

# Go 的 Protocol Buffer 教程

欢迎程序员们！在这个教程里面，我们将学习如何在你的 Go 应用中使 protocol Buffers 数据格式。我们将详细讲述这种数据格式，以及为什么这种数据格式优于传统的数据格式，例如 XML 甚至 JSON。在我们写更多复杂的例子之前，我们将从一个简单的例子开始编写运行。

在这个教程的最后，你会对 protoco Buffe 有一个基础的了解，并且你将可以顺利的写出更好的系统。

## 视频教程

https://www.youtube.com/embed/NoDRq6Twkts?ecver=2

## Protocol Buffer 数据类型

Protocol buffers 是基础的数据格式，和 JSON、XML 非常的相似，都是用于存储结构化的数据，都支持很多不同的语言使用并进行序列化和反序列化。

这种数据格式最大的优点就是它相比 XML 甚至 JSON 的大小要小很多。这种格式是由 Google 开发的。Google 公司的规模非常巨大，以至于每一个被节省的字节空间都有巨大的价值。

假设有一个人，以三种不同的数据格式表示它：

```xml
<person>
	<name>Elliot</name>
	<age>24</age>
</person>
```

我们可以使用更小的数据格式 JSON 表示它：

```json
{
	"name": "Elliot",
	"age": 24
}
```

如果我们使用 protocol buffer 数据格式表示如下：

```
[10 6 69 108 108 105 111 116 16 24]
```

如果你仔细观察上面一行的编码输出，你可以看到从数组下标为二的位置开始，`ellio` 就拼出来了。`e`=69,`l`=108 等。后面的字节表示我现在 24 岁了。

但是这个编码内容比我们看到的要多的多。我仍然在尝试研究更多信息。如果你愿意，我建议可以查看更多 Google 关于 protocol Buffer 编码的文档：
[Protocol Buffer Encoding](https://developers.google.com/protocol-buffers/docs/encoding)

虽然 JSON 和 Protocol Buffer 的大小几乎相同。但是随着数据增大大于 "入门" 例子的数据，JSON 和 protocol buffer 使用的空间差距就变大了。

## 一个简单的例子

```bash
$ go get github.com/golang/protobuf
$ go get github.com/golang/protobuf/proto
```

上面下载一些必须的包，用于运行简单的例子。

```bash
($) export PATH=$PATH:$GOPATH/bin
```

进行上面的设置之后，你就可以在终端使用 `protoc` 这个命令了。下面我们就可以定义 protobuf 的格式了，在这个例子里，我们将尝试使用相同的 `person` 这个对象，我们用这个突出不同数据格式之间的区别。

首先我们要指出要使用的协议类型，在这个例子里面我们使用 `proto3`。然后我把它作为 `main` 包的一部分。

最后我们定义我们想要的数据结构。这个包含了 `Person` 的消息结构，其中包含了 `name` 和 `age` 两个字段。

```protobuf
syntax="proto3";

package main;

message Person {
	string name = 1;
	int32 age = 2;
}
```

然后我们使用 `protoc` 命令编译这个文件。

```bash
$ protoc --go_out=. *.proto
```

最终我们准备好写我们 Go 代码的所有东西。我们从定义一个 `Person` 开始，并将这个对象编译成 `protobuf` 对象。

为了了解它是如何存储的，我们使用 `fmt.Println(data)` 打印存储 `protobuf` 对象的编码数据。

```go
package main

import (
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
)

func main() {

	elliot := &Person{
		Name: "Elliot",
		Age:  24,
	}

	data, err := proto.Marshal(elliot)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	// printing out our raw protobuf object
	fmt.Println(data)

	// let's Go the other way and unmarshal
	// our byte array into an object we can modify
	// and use
	newElliot := &Person{}
	err = proto.Unmarshal(data, newElliot)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}

	// print out our `newElliot` object
	// for Good measure
	fmt.Println(newElliot.GetAge())
	fmt.Println(newElliot.GetName())
}
```

在运行之前，我们需要将 `test.pb.go` 编译通过以保证正常工作：

```
➜ src go run main.go test.pb.go
[10 6 69 108 108 105 111 116 16 24]
name:"Elliot" age:24
```

## 嵌套字段

好了，我们实现了一个非常简单的例子并运行了它，但是在实际中，我们经常遇到在 messag 的格式里面有嵌套的字段，并且可能会修改一些它们的值。

现在我们开始看看如何使用嵌套字段。我们继续使用 `Person` 这个消息格式，我们将添加一个社交媒体的追随者的字段。

我们用标准的字段以及自定义的 `SocialFollowers` 消息字段组成 Person 这个消息格式，像下面这样：

```protobuf
syntax="proto3";

package main;

message SocialFollowers {
	int32 YouTube = 1;
	int32 Twitter = 2;
}

message Person {
	string name = 1;
	int32 age = 2;
	SocialFollowers socialFollowers = 3;
}
```

再一次，我们使用 `protoc` 这个命令生成我们想要的东西。

```bash
($) protoc --go_out=. *.proto
```

然后我们再回到我们的 Go 程序，我们可以用 `SocialFollowers` 补充我们的 `elliot` 对象：

```go
package main

import (
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
)

func main() {

	elliot := Person{
		Name: "Elliot",
		Age:  24,
		SocialFollowers: &SocialFollowers{
			Youtube: 2500,
			Twitter: 1400,
		},
	}

	data, err := proto.Marshal(&elliot)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	// let's Go the other way and unmarshal
	// our protocol buffer into an object we can modify
	// and use
	newElliot := &Person{}
	err = proto.Unmarshal(data, newElliot)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}

	// print out our `newElliot` object
	// for Good measure
	fmt.Println(newElliot.GetName())
	fmt.Println(newElliot.GetAge())
	fmt.Println(newElliot.SocialFollowers.GetTwitter())
	fmt.Println(newElliot.SocialFollowers.GetYoutube())
}
```

我们来最后一次运行它，我们看到了所有我们希望输出的内容：

```
➜ src go run main.go test.pb.go
Elliot
24
1400
2500
```

## 总结

在这个教程里面，我们了解了如何基于 Go 应用程序使用 protocol buffer 建立数据结构并运行。

希望这个教程对您有用。

---

via: https://tutorialedge.net/golang/go-protocol-buffer-tutorial/

作者：[Elliot Forbes](https://twitter.com/Elliot_F)
译者：[amei](https://github.com/amei)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
