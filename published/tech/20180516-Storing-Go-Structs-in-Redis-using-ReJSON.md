已发布：https://studygolang.com/articles/13252

# 使用 ReJSON 在 Redis 中保存 Go 结构体

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/rejson/1_w3hPEpsPFtHs36dJMUdl7w.jpeg)

> 图像授权 https://Redislabs.com/blog/Redis-go-designed-improve-performance/

大部分人可能对 Redis 都很熟悉了。对于外行人来说，Redis 是最广为人知并广泛应用的数据库/缓存产品，起码也是之一。

官方文档是这么描述 Redis 的：

> Redis 是一个开源（BSD 许可）的，内存中的数据结构存储系统，它可以用作数据库、缓存和消息中间件。 它支持的数据结构有字符串（strings），散列（hashes），列表（lists），集合（sets），有序集合（sorted sets）与范围查询， bitmaps， hyperloglogs 和 地理空间（geospatial）索引半径查询。 Redis 内置了复制（replication），LUA脚本（Lua scripting）， LRU 驱动事件（LRU eviction），事务（transactions） 和不同级别的磁盘持久化（persistence）， 并通过 Redis 哨兵（Sentinel）和自动分区（Cluster）提供高可用性（high availability）。

将 Redis 与其它（传统）数据库区分开来的是，Redis 是一个键-值 存储（并且是在内存中）。这意味着在这个数据库中所有的值都与一个 key 相关联（想想字典的情况）。不过我跑题了，这篇文章可不是讲 Redis 的，让我们言归正传。

## 使用 Go 语言与 Redis 进行交互

当 Go 开发者使用 Redis 时，有时会需要将我们的对象缓存到 Redis 中。我们看看如何通过 Redis 中的 HMSET 来实现这点。

一个简单的 go 结构体可能会像这样，

```go
type SimpleObject struct {
    FieldA string
    FieldB int
}
simpleObject := SimpleObject{"John Doe", 24}
```

很明显，为了将对象存到 Redis 中，我们必须将它转化成一个键值对，我们将结构体的字段名作为 key，字段的值作为这个 key 对应的值。

对所有的字段取一个哈希值也是非常好的键值，将对象所有的字段与这个对象自身绑定起来。在 Redis-cli 中我们可以这么做：

```
127.0.0.1:6379> HMSET simple_object fieldA "John Doe" fieldB 24
OK
```

用 HGETALL 命令获取的结果是，

```
127.0.0.1:6379> HGETALL simple_object
fieldA
John Doe
fieldB
24
```

好吧，现在我们知道对象是怎样序列化后存入数据库中的，让我们继续用程序的方式完成这个工作！

虽然 Redis 的 Go 客户端很多，但我使用 redigo，它在 github 上有一个很不错的社区，而且也是最常用的 Redis 的 Go 客户端之一，有超过 4K 个星星。

### Redigo 助手函数 — AddFlat 和 ScanStruct

Redigo自带了一系列很棒的助手函数，其中我们将用到 AddFlat ，在我们将结构体存入 Redis 之前，用它将结构体扁平化。

```go
// 获得链接对象
conn, err := Redis.Dial("tcp", "localhost:6379")
if err != nil {
    return
}
// 使用 Do 方法调用命令
_, err = conn.Do("HMSET", Redis.Args{"simple_object"}.AddFlat(simpleObject)...)
if err != nil {
    return
}
```

现在，如果你希望读回这个对象，我们可以使用 HGETALL 命令，

```go
value, err := Redis.Values(conn.Do("HGETALL", key))
if err != nil {
    return
}
object := SimpleStruct{}
err = Redis.ScanStruct(value, &object)
if err != nil {
    return
}
```

很简单，对吧？让我们在看看更深入的一些问题...

### Go 结构体中嵌套的对象

现在，我们来看一个更复杂的结构体，

```go
type Student struct {
    Info *StudentDetails `json:"info,omitempty"`
    Rank int `json:"rank,omitempty"`
}
type StudentDetails struct {
    FirstName string
    LastName string
    Major string
}
studentJD := Student{
    Info: &StudentDetails{
        FirstName: "John",
        LastName: "Doe",
        Major: "CSE",
    },
    Rank: 1,
}
```

现在我们有一个嵌套的结构体，`StudentDetails` 是 `Student` 对象的一个成员。

让我们再用 `HMSET` 试试看，

```go
// 用 Do 方法调用命令
_, err = conn.Do("HMSET", Redis.Args{"JohnDoe"}.AddFlat(studentJD)...)
if err != nil {
    return
}
```

如果我们再看看 Redis 中存进了什么，我们可以看到的是这样的，

```
127.0.0.1:6379> HGETALL JohnDoe
Info
&{John Doe CSE}
Rank
1
```

这就是问题点了。当我们想从 Redis 中读数据并转化为对象时，**ScanStruct** 会报错，

```
redigo.ScanStruct: cannot assign field Info: cannot convert from Redis bulk string to *main.StudentDetails
```

**EPIC FAIL !**

这是因为 Redis 将所有的东西都存为**字符串**[大对象使用 bulk 字符串]。

### 现在该怎么办？

快速的搜索可以给你一些解决方案，其中之一是建议使用一个封装处理器（Marshaler）（`JSON` marshal），其他的的方案建议用 `MessagePack`。

以下我将展示采用 `JSON` 的解决方案

```go
b, err := json.Marshal(&studentJD)
if err != nil {
    return
}
_, err = conn.Do("SET", "JohnDoe", string(b))
if err != nil {
    return
}
```

需要取回值时，只需要使用 `GET` 命令将 `JSON` 字符串读取回来就可以了。

```go
objStr, err = Redis.String(conn.Do("GET", "JohnDoe"))
if err != nil {
    return
}
b := []byte(objStr)
student := &Student{}
err = json.Unmarshal(b, student)
if err != nil {
    return
}
```

如果我们是希望**将对象完整的缓存下来**，这个方案工作的很好。但是，如果我们希望在对象上进行增加、修改或者读取一个字段，比如，**John Doe** 将他的专业从 **CSE** 改成了 **EE**，该怎么办？

唯一的办法是，先读出 JSON 字符串，转化为对象，修改对象，然后重新将对象存回 Redis。这看起来可是不少的工作！

> 如果你发现，使用 Hash，通过 `HGET`/`HSET` 命令来实现这一点很简单。如果只是这样，那就这么做吧。（原文：If you are wondering, doing this with the Hash is trivial by using the HGET/HSET commands. If only, that worked — bummer!）

### ReJSON

优秀的 [RedisLabs](https://redislabs.com/) 团队给我们带来了一个解决方案，让我们可以对应像操作传统 JSON 对象那样操作 Redis 中的对象。

让我们马上来看看。我从 [rejson](http://rejson.io/) 的文档中挑选了这个例子，

```
127.0.0.1:6379> JSON.SET amoreinterestingexample . '[ true, { "answer": 42 }, null ]'
OK
127.0.0.1:6379> JSON.GET amoreinterestingexample
"[true,{\"answer\":42},null]"
127.0.0.1:6379> JSON.GET amoreinterestingexample [1].answer
“42”
127.0.0.1:6379> JSON.DEL amoreinterestingexample [-1]
1
127.0.0.1:6379> JSON.GET amoreinterestingexample
"[true,{\"answer\":42}]"
```

用程序来实现这个，我们绝对可以使用 `Redigo` 的原生形态[这意味者我们可以使用 `conn.Do(...)` 命令调用任何 Redis 支持的命令]。

然而，我花了一些时间将所有的 `ReJSON` 的命令转换成了 Go 包，叫做 [go-rejson](https://github.com/nitishm/go-rejson)。回到我们之前的 `Student` 对象，我们可以用以下步骤使用程序将它存入 Redis 中，

```go
import "github.com/nitishm/go-rejson"
_, err = rejson.JSONSet(conn, "JohnDoeJSON", ".", studentJD, false, false)
if err != nil {
    return
}
```

在 `redis-cli` 中我们可以查到，

```
127.0.0.1:6379> JSON.GET JohnDoeJSON
{"info":{"FirstName":"John","LastName":"Doe","Major":"CSE"},"rank":1}
```

如果我只想从 Redis 条目（entry）读取 `info` 字段，我会执行 `JSON.GET`，如下所示，

```
127.0.0.1:6379> JSON.GET JohnDoeJSON .info
{"FirstName":"John","LastName":”Doe","Major":"CSE"}
```

类似地，对于 `rank` 字段，我通过 `.rank` 引用，

```
127.0.0.1:6379> JSON.GET JohnDoeJSON .rank
1
```

使用程序来获取 student 对象，我们可以通过 `JSONGet()` 方法调用 `JSON.GET` 命令，

```go
v, err := rejson.JSONGet(conn, "JohnDoeJSON", "")
if err != nil {
    return
}
outStudent := &Student{}
err = json.Unmarshal(outJSON.([]byte), outStudent)
if err != nil {
    return
}
```

为了给 `rank` 字段赋值，我们可以在 `.rank` 字段上使用 `JSONSet()` 方法来调用 `JSON.SET` 命令，

```go
_, err = rejson.JSONSet(conn, "JohnDoeJSON", ".info.Major", "EE", false, false)
if err != nil {
    return
}
```

在 `redis-cli` 中我们查看这个条目，可以看到，

```
127.0.0.1:6379> JSON.GET JohnDoeJSON
{"info":{"FirstName":"John","LastName":"Doe","Major":"EE"},"rank":1}
```

## 运行这个例子

### 用 Docker 启动带 rejson 模块的 Redis

```
docker run -p 6379:6379 --name Redis-rejson Redislabs/rejson:latest
```

### 从 github 上克隆这个例子

```
# git clone https://github.com/nitishm/rejson-struct.git
# cd rejson-struct
# go run main.go
```

想要了解更多 **Go-REJSON** 包，请访问 https://github.com/nitishm/go-rejson.

想要了解跟多 **ReJSON**，访问它们的官方文档, http://rejson.io/.

---

via: https://medium.com/@nitishmalhotra/storing-go-structs-in-Redis-using-rejson-dab7f8fc0053

作者：[Nitish Malhotra](https://medium.com/@nitishmalhotra)
译者：[MoodWu](https://github.com/MoodWu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
