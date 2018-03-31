已发布：https://studygolang.com/articles/12725

# 反射机制介绍

反射是指一门编程语言可以在运行时( runtime )检查其数据结构的能力。利用 Go 语言的反射机制，可以获取结构体的公有字段以及私有字段的标签名，甚至一些其他比较敏感的信息。

众所周知Go标准库中有一些包利用反射机制来实现它们的功能。我们经常会以  [encoding/json](https://golang.org/pkg/encoding/json/) 包为例，该包常用来把 JSON 文档解析为结构体，同时也可以把结构体编码为JSON格式的字符串。

本文中我想给大家介绍一个略微有点不一样的例子，该例子是我最近在做的一个聊天项目的消息体，该消息体使用结构体来表示：

```go
type Message struct {
	ID         uint64    `db:"id"`
	Channel    string    `db:"channel"`
	UserName   string    `db:"user_name"`
	UserID     string    `db:"user_id"`
	UserAvatar string    `db:"user_avatar"`
	Message    string    `db:"message"`
	RawMessage string    `db:"message_raw"`
	MessageID  string    `db:"message_id"`
	Stamp      time.Time `db:"stamp"`
}
``` 

我想使用 命名 SQL 查询语句把该条消息写入到数据库，正常情况，我应该写一条类似于下面的 SQL 查询语句：

> insert into messages set id=:id, channel=:channel,...

然后使用 `jmoiron.sqlx` 包执行一次 db.NamedExec(query,message) 语句。很显然，随着结构体数量的增加，大量的时间将会浪费在写这些查询语句上面，更糟糕的是，一旦数据库表结构发生变更的话，程序很可能会报错。

试想一下，如果我们能够根据传过来的结构体，自动生成查询语句，会不会是一件很爽的事情？

的确我们可以通过引用 reflect 包来达到我们的上述需求。接下来我将会带着大家一起来领略一下整个实现过程，目前有很多 ORM 包也是使用了跟我类似的方法来达到相同的目的。

## 把结构体转为 reflect.value 类型

我们需要先创建一个 `reflect.Value` 的实例，以便于能够获取结构体的字段。同时我们也可以从该实例中获取结构体的函数。创建一个 reflect.Value 实例非常直接：

```go
message_value := reflect.ValueOf(message)
```

我们需要调用 message_value.NumField() 函数来获取结构体中字段的总数以便于迭代结构体的所有字段。如果我们试图调用 NumField() 的时候传一个 reflect.ValueOf 返回的指针值，程序会产生 panic 错误：

```go
panic:reflect:call of reflect.Value.NumField on ptr Value
```

为了解决上面这个问题，我们使用 message_value.Kind() 来检查是否是一个指针值，然后得到指针指向的实际的值：

```go
if message_value.Kind() == reflect.Ptr{
	message_value = message_value.Elem()
}
```

然后我们再调用 message_value.NumField() 就会正确的输出结构体字段的总数。接下来我们将会使用这个值通过循环迭代的方式来获取所有字段的名称和对应的字段值。

## 读取字段详情

从结构体字段中我们可以获取很多重要的信息，不过我们最感兴趣的还是想要获取字段声明中的标签信息。由于 `reflect.Value` 是用来处理结构体中每个字段实际存储的值，所以我们需要用 reflect.Type 来获取字段的名称（比如 UserName ）或者关联的标签名。

假如我们想要获取结构体所有字段的详细信息列表，详细信息包含字段名称，含有“db”的字段关联的标签名以及字段实际存储的值，代码可以这样写：

```go
message_fields := make([]struct {
	Name  string
	Tag   string
	Value interface{}
}, message_value.NumField())

for i := 0; i < len(message_fileds); i++ {
	fieldValue := message_value.Field(i)
	fieldType := message_value.Type().Field(i)
	message_fields[i].Name = fieldType.Name
	message_fields[i].Value = fieldsValue.Interface()
	message_fields[i].Tag = fieldType.Tag.Get("db")
}
```

上述代码可以看出每个字段的实际值通过 reflect.Value.Interface() 获取，字段的名称和字段的标签名通过 reflect.Type 获取。你可以在 [go playground](https://play.golang.org/p/Bu0J-jlsLB7) 上跑一下上述完整的示例。

## 组合一下功能

其实上述的功能已经完全满足我们的需求了，自动生成 sql 查询语句的关键点在于使用代码来完成字段的标签名的拼接，如下代码所示：

```go
func insert(table string, data interface{})string{
	message_value := reflect.ValueOf(data)
	if message_value.Kind() == reflect.Ptr{
		message_value = message_value.Elem()
	}
	message_fields := make([]string, message_value.NumField())
	for i := 0;i<len(message_fields);i++{
		fieldType := message_value.Type().Field(i)
		message_fields[i] = fieldType.Tag.Get("db")
	}

	sql := "insert into" + table + " set"
	for _,tagFull := range message_fields{
		if tagFull != "" && tagFull != "-"{
			tag := strings.Split(tagFull,",")
			sql = sql + " "+ tag[0]+"=:"+tag[0]+","
		}
	}
	return sql[:len(sql)-1]
}
```

最终版的代码在这里[go playground code](https://play.golang.org/p/KcuTIWa3S1F)

这里还有一些注意事项需要说明一下：

* 在我们的例子中，由于我们没有深入对结构体进行解析，所以不论结构体的字段是否是指针类型都无关紧要。同时 reflect.Type 信息的获取与字段的实际值也无关。
* 如果你需要继续对结构体进行解析，一定要注意对指针类型的值进行特殊处理，就像我们前面对 message_value 一样，通过 Elem() 来获取指针值。
* 还有一些其他非常优秀的包提供了反射机制，如果你想要阅读真实的示例，可以看一下[codegangsta/inject](https://github.com/codegangsta/inject)和[fatih/structs](https://github.com/fatih/structs)这两个包。
* 推荐大家阅读 [The Laws Of Reflection, by Rob Pike](https://blog.golang.org/laws-of-reflection)

最后提醒一下：如果你之前是一个 PHP 或者 Javascript 程序员，你应该知道在这些语言中缺乏类型安全监测。你可能希望通过反射机制来解决你遇到的一些问题。但是，如果你选择使用 Go 来解决在之前这些语言中遇到的一些使用反射机制解决的问题，我估计你会很痛苦，甚至可能会放弃。

从我以往使用 Go 案例的历史来看，使用反射机制总是磕磕绊绊的。即便是本文中的案例，也是有些牵强的。你完全可以不用使用反射机制，使用字符串切片或者查询语句本身也可以达到上述的目的。如果说你写的 API 接口完全可以满足你的需求，就尽量避免使用反射机制。

案例和观点：即便对于 JSON ，也有一个 [json-iterator/go](https://github.com/json-iterator/go) 包用来代替标准库 `encoding/json` 包。在该包中，大大减少了对反射机制的依赖，速度也得到了显著的提升。反射机制重构的可能性不太大，建议大家使用 gRPC 和 Protobuf 机制。在这个机制中，类型安全可以在代码编写的过程中得到保障。

---

via：https://scene-si.org/2017/12/21/introduction-to-reflection/

作者：[Tit Petric](https://scene-si.org)
译者：[yzhfd](https://github.com/yzhfd)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出