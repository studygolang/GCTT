# 使用ReJSON在Redis中保存Go结构体

![image](https://cdn-images-1.medium.com/max/1600/1*w3hPEpsPFtHs36dJMUdl7w.jpeg) 

图像授权 https://redislabs.com/blog/redis-go-designed-improve-performance/

大部分人可能对 Redis 都很熟悉了。对于外行人来说，Redis 是最广为人知并广泛应用的数据库/缓存产品，起码也是之一。
官方文档是这么描述 redis 的，Redis是一个开源（BSD 协议）的，内存数据结构存储，可以用作数据库，缓存或者消息分发。它支持的数据类型有，字符串，哈希，列表，集合，可以进行范围查询的有序集合，位图，基数和可以进行距离查询的地理数据索引。Redis 内置了复制，Lua 脚本语言，LRU 回收算法，事务以及不同级别的持久化到磁盘的处理，并且通过 Redis 集群的 Sentinel 系统和自动分区功能提供高可用性。

将 redis 与其它（传统）数据库区分开来的是，redis 是一个键-值 存储（并且是在内存中）。这意味着在这个数据库中所有的值都与一个 key 相关联（想想字典的情况）。不过我跑题了，这篇文章可不是讲 redis 的，让我们言归正传。

## 使用 Go 语言与Redis进行交互

当 Go 开发者使用 redis 时，有时会需要将我们的对象缓存到 redis 中。我们看看如何通过 redis 中的 HMSET 来实现这点。
一个简单的 go 结构体可能会像这样，
```go
type SimpleObject struct {
    FieldA string
    FieldB int
}
simpleObject := SimpleObject{“John Doe”,24}
```
很明显，为了将对象存到 redis 中，我们必须将它转化成一个键值对，我们将结构体的字段名作为 key，字段的值作为这个 key 对应的值。
对所有的字段取一个哈希值也是非常好的键值，将对象所有的字段与这个对象自身绑定起来。在 redis-cli 中我们可以这么做：
```
127.0.0.1:6379> HMSET simple_object fieldA “John Doe” fieldB 24
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
现在我们知道对象可以封装化后存入数据库中，让我们继续用程序的方式完成这个工作！

虽然 redis 的 go 客户端很多，但我使用 redigo，它在 github 上有一个很不错的社区，而且也是最常用的 redis 的 go 客户端之一，有超过4K个星星。

## Redigo 助手函数 —  AddFlat 和 ScanStruct 

Redigo自带了一系列很棒的助手函数，其中我们将用到 AddFlat ，在我们将结构体存入 redis 之前，用它将结构体扁平化。
```go
// 获得链接对象
conn, err := redis.Dial(“tcp”, “localhost:6379”)
if err != nil {
    return
}
// 使用Do方法调用命令
_, err = conn.Do(“HMSET”, redis.Args{“simple_object”}.AddFlat(simpleObject)…)
if err != nil {
    return
}
```
现在，如果你希望读回这个对象，我们可以使用 HGETALL 命令，
```go
value, err := redis.Values(conn.Do(“HGETALL”, key))
if err != nil {
    return
}
object := SimpleStruct{}
err = redis.ScanStruct(value, &object)
if err != nil {
    return
}
```
很简单，对吧？让我们在看看更深入的一些问题...

## Go 结构体中嵌套的对象

现在，我们来看一个更复杂的结构体，
```go
type Student struct {
    Info *StudentDetails `json:”info,omitempty”`
    Rank int `json:”rank,omitempty”`
}
type StudentDetails struct {
    FirstName string
    LastName string
    Major string
}
studentJD := Student{
    Info: &StudentDetails{
        FirstName: “John”,
        LastName: “Doe”,
        Major: “CSE”,
    },
    Rank: 1,
}
```
现在我们有一个嵌套的结构体，StudentDetails 是 Student 对象的一个成员，让我们再用 HMSET试试看，
```go
// 用Do方法调用命令
_, err = conn.Do(“HMSET”, redis.Args{“JohnDoe”}.AddFlat(studentJD)…)
if err != nil {
    return
}
```
如果我们再看看redis中存进了什么，我们可以看到的是这样的，
```go
127.0.0.1:6379> HGETALL JohnDoe
Info
&{John Doe CSE}
Rank
1
```
这就是问题点了。当我们想从 redis 中读数据并转化为对象时，ScanStruct 会报错，
```go
redigo.ScanStruct: 无法赋值字段信息:无法将redis的字符串包转化为  *main.StudentDetails
EPIC FAIL !
```
这是因为 redis 将所有的东西都存为字符串[字符串包或者大对象]。

现在该怎么办？

一个快速的搜索可以给你一些解决方案，其中之一是建议使用一个封装处理器（ JSON 封装），其他的的方案建议用 MessagePack。

以下我将展示采用 JSON 的解决方案
```go
b, err := json.Marshal(&studentJD)
if err != nil {
    return
}
_, err = conn.Do(“SET”, “JohnDoe”, string(b))
if err != nil {
    return
}
```
需要取回值时，只需要使用 GET 命令将 JSON 字符串读取回来就可以了。
```go
objStr, err = redis.String(conn.Do(“GET”, “JohnDoe”))
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
如果我们是希望将对象完整的缓存下来，这个方案工作的很好。但是，如果我们希望在对象上进行增加，修改或者读取一个字段，比如，John Doe 将他的专业从 CSE 改成了 EE,该怎么办？
唯一的办法是，先读出 JSON 字符串，转化为对象，修改对象，然后重新将对象存回 redis。这看起来可是不少的工作！

如果你在犹豫，利用 HASH 值来存储对象，并配合 HGET/HSET 命令读取存入对象只是比较繁琐，如果这个方法真的有效--糟透了。
因为每个web-post 都存在它们的文化基因。


## ReJSON
这个优秀的 RedisLabs 的团队给我们带来了一个解决方案，让我们可以对应想操作传统JSON对象那样操作 Redis 中的对象。

让我们马上来看看。我从 rejson 的文档中挑选了这个例子，
```
127.0.0.1:6379> JSON.SET amoreinterestingexample . ‘[ true, { “answer”: 42 }, null ]’
OK
127.0.0.1:6379> JSON.GET amoreinterestingexample
“[true,{\”answer\”:42},null]”
127.0.0.1:6379> JSON.GET amoreinterestingexample [1].answer
“42”
127.0.0.1:6379> JSON.DEL amoreinterestingexample [-1]
1
127.0.0.1:6379> JSON.GET amoreinterestingexample
“[true,{\”answer\”:42}]”
```
用程序来实现这个，我们绝对可以使用 Redigo 的原生形态[这意味者我们可以使用 conn.Do(...)命令调用任何 Redis 支持的命令]。

然而，我花了一些时间将所有的 REJSON 的命令转换成了 Go 包，叫做 go-rejson。回到我们之前的 Student  对象，我们可以用以下步骤使用程序将它存入 redis 中，

```go
import "github.com/nitishm/go-rejson"
_, err = rejson.JSONSet(conn, “JohnDoeJSON, “.”, studentJD, false, false)
if err != nil {
    return
}
```
在 redis-cli 中我们可以查到，
```
127.0.0.1:6379> JSON.GET JohnDoeJSON
{“info”:{“FirstName”:”John”,”LastName”:”Doe”,”Major”:”CSE”},”rank”:1}
If I wish to just read the info field from the redis entry I would perform a JSON.SET as follows,
127.0.0.1:6379> JSON.GET JohnDoeJSON .info
{“FirstName”:”John”,”LastName”:”Doe”,”Major”:”CSE”}
Similarly with the rank field, I could reference the .rank,
127.0.0.1:6379> JSON.GET JohnDoeJSON .rank
1
```
使用程序来获取 student 对象，我们可以通过 JSONGET() 方法调用 JSON.GET 命令，

```go
v, err := rejson.JSONGet(conn, “JohnDoeJSON, “”)
if err != nil {
    return
}
outStudent := &Student{}
err = json.Unmarshal(outJSON.([]byte), outStudent)
if err != nil {
    return
}
```
为了给 rank 字段赋值，我们可以在 .rank 字段上使用 JSONSET() 方法 调用 JSON.SET 命令，
```go
_, err = rejson.JSONSet(conn, “JohnDoeJSON, “.info.Major”, “EE”, false, false)
if err != nil {
    return
}
```
在redis-cli中我们查看这个条目，可以看到，
```
127.0.0.1:6379> JSON.GET JohnDoeJSON
{“info”:{“FirstName”:”John”,”LastName”:”Doe”,”Major”:”EE”},”rank”:1}
```

## 运行这个例子
用 Docker 启动带rejson的redis
```
docker run -p 6379:6379 --name redis-rejson redislabs/rejson:latest
```
从github上克隆这个例子
```
# git clone https://github.com/nitishm/rejson-struct.git
# cd rejson-struct
# go run main.go
```
想要了解更多 Go-REJSON 包，请访问 https://github.com/nitishm/go-rejson.

想要了解跟多 Rejson，访问他们的官方文档, http://rejson.io/.


via: https://medium.com/@nitishmalhotra/storing-go-structs-in-redis-using-rejson-dab7f8fc0053

作者：[Nitish Malhotra](https://medium.com/@nitishmalhotra)
译者：[MoodWu](https://github.com/MoodWu)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出