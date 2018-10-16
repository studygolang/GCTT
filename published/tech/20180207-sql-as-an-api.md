已发布：https://studygolang.com/articles/13066

# 以 SQL 作为 API

如果你不是在石头下住着，那么你也应该听过最近兴起一种新的对“函数作为服务”的理解。在开源社区，Alex Ellis 的 [OpenFaas](https://github.com/openfaas/faas)  项目受到了很高的关注，并且 [亚马逊Lambda宣布对Go语言的支持](https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/)。这些系统允许你按需扩容，并且通过 API 调用的方式来调用你的 CLI 程序。

## Lambda/Faas 背后的动机

让我们这么来描述 - 整个“无服务器”运动是云平台，比如  AWS，的市场营销行为，它允许你将任一个服务器管理移交给它们，比如，理想情况下，你收入系统的一小部分。在具体的条款上，这意味着 AWS 或其他类似的解决方案 托管的你的应用，运行你的应用，并且在他们的数据中心上根据需要自动维护它的硬件规模。

但是，你可能早就知道这些了。

但是，你是否知道，这种能力早就在 CGI 中存在？参考维基百科，1993，并且在 1997 正式以 RFC 规范定下来的定义。所有旧的东西又卷土重来了。 CGI(Common Gateway Interface) 的目的是:

> 在计算上，通用网关接口 (CGI)  为网站服务器提供一个标准的协议，用以在动态网页服务器上像执行控制台应用（也称为命令行接口程序）一样执行应用程序。
>
> 来源: [维基百科](https://en.wikipedia.org/wiki/Common_Gateway_Interface)
>

对 Go 语言来说，最简单的 Fass/CGI 服务程序只需要 10 行左右的代码就能实现。Go 语言的标准库中已经包含了  [net/http/cgi](https://golang.org/pkg/net/http/cgi/) 来完成所有的困难工作。实现一个 PHP CGI 只需要如下寥寥数行：

```go
func phpHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := new(cgi.Handler)
		handler.Dir = "/var/www/"
		handler.Path = handler.Dir + "api"
		args := []string{r.RequestURI}
		handler.Args = append(handler.Args, args...)
		fmt.Printf("%v", handler.Args)
		handler.ServeHTTP(w, r)
	}
}
```
调用这个也是非常简单的：

```go
http.HandleFunc("/api/", phpHandler())
```

当然，我不清楚为什么你会使用它。因为在那时黎明前的时刻， CGI  碰到了性能问题，其中最大的问题是它在系统上产生的压力。对于每个请求，就会发起一个 `os.Exec` 调用，而这绝对不是一个系统友好的调用。实际上，如果你在做任何类似实时交通这样的服务，你很可能希望这样的调用数是 0 。

这就是为什么 CGI 进化为了 FastCGI。

> FastCGI 是通用网关接口 （CGI） 早期的一种变化； FastCGI 的主要目的是减少网页服务器和 CGI 程序之间的过度的联系，让服务器能一次性处理更多的网页请求
>
> 来源：[维基百科](https://en.wikipedia.org/wiki/FastCGI)

我现在不想实现一个 FastCGI 程序（在标准库中也有一个  [net/http/fcgi](https://golang.org/pkg/net/http/fcgi/)） ,但我想要演示这种实现会带来的性能陷阱。当然，你在 AWS 上运行你的程序时，你可能不怎么关心这个，因为它们有能力按照你的访问量来扩容硬件。

## CGI 的解决办法

如果这几年我有学到点东西的话，那一定是大多数的服务是数据驱动的。这意味着，一定有某种数据库保存着至关重要的数据。根据一个 [Quora上的回答](https://www.quora.com/Which-database-system-s-does-Twitter-use) ， Twitter 使用了至少8种不同类型的数据库，从 MySQL,Cassandra 和 Redis，到其他更复杂的数据库。

实际上，我的大部分工作大概是这样的，从数据库中读取员工数据，然后把它变成 json 格式并且通过 REST 调用提供出去，这些查询通常不能仅仅用一条 SQL 语句来实现，当然在很多情况下也是可以的。那么，不如我们写一些不会有 `os.Exec` 调用成本的  SQL 脚本来实现一些功能，而不是用 CGI 程序来实现它们？

挑战接受了。

## 以 SQL 作为 API

我并不想把这个变成庞然大物，虽然 以 SQL 作为 API 是有能力成为的，但我确实想要实现一个远程可用的版本。我希望通过在磁盘上创建一个 .sql  文件来实现 API 调用，并且我还希望这个 API 可以调用 http 请求中的任意参数。这意味着我们可以通过传递给 API 的参数来过滤结果集。我选择了 MySQL 和 sqlx 来实现这个任务。

最近我为 Twitch，Slack，Yotube，Discord 写了一些聊天机器人，看起来很快我要写一个 Teleganm 的版本了。其实他们的目的是对各种通道进行相似的连接，记录消息，加总一些统计信息并且对一些命令或者问题进行反馈。对于用 Vue.js 写成的前端网站，我们需要通过 API 来向他们传递一些数据。虽然不是所有的 API 都能用 SQL 来实现，但还是有很大一部分是可以的。比如：

1. 列出所有的通道
2. 通过通道ID列出通道

这两个调用相对来说很相似并且容易实现，我特别创建了两个文件，来提供这些信息：

`api/channels.sql` (对应于 `/api/channels` )

```sql
select * from channels
```

`api/channelByID.sql` (对应于 `/api/channelByID?id=...` )

```sql
select * from channels where id=:id
```

就像你看到的这样，用 SQL 查询来实现一个新的 API 并不需要太多的工作。我尝试设计一个系统，这样一旦你创建了 `api/name.sql` 马上就可以通过 `api/name` 访问到。所有 Http 请求的参数被封装在 `map[string]interface{}` 中，并且作为绑定变量传递给 SQL 查询语句。SQL 驱动负责来处理这些参数。

我也设计了错误信息的格式化。如果你没法连接到你的数据库，或者一个 API 对应的 .sql  文件不存在，会有一个如下的错误信息返回：

```json
{
	"error": {
		"message": "open api/messages.sql: no such file or directory"
	}
}
```
在 SQL 查询中使用 URL 参数

在 go 语言中获得请求的参数，只需要在请求对象中的 `*url.URL` 结构体上调用 Query() 函数即可。这个函数返回 `url.Values` 对象，这是一个 `map[string][]string` 类型的别名。
我们需要转换这个对象并传递到 sqlx 的语句中。我们需要创建一个 `map[string]interface{}` 。因为我们需要调用的 sqlx 函数在查询时接受这种格式的参数（[sqlx.NamedStmt.Queryx](https://godoc.org/github.com/jmoiron/sqlx#NamedStmt.Queryx)）。让我们转换它们并且发起查询：

```go
params := make(map[string]interface{})
urlQuery := r.URL.Query()
for name, param := range urlQuery {
	params[name] = param[0]
}

stmt, err := db.PrepareNamed(string(query))
if err != nil {
	return err
}
rows, err := stmt.Queryx(params)
if err != nil {
	return err
}
```
我们还没有处理的是 `rows` 变量，我们可以遍历它以获得每一个行的信息。我们需要将它们加入到一个切片中，并且在 API 的最后步骤中把它们封装到 JSON 里面。

```go
for rows.Next() {
	row := make(map[string]interface{})
	err = rows.MapScan(row)
	if err != nil {
		return err
	}
```
这儿是事情变得有趣的地方。每一行中包含的值都需要转换成 JSON 编码器能理解的东西。
因为底层的类型是 `[]uint8` ，我们首先要把它们转换成字符串。如果我们不这么做，这种结构的 JSON 会自动使用 base64 编码。既然查询的反馈可以用 `map[string]string` 来表示，并且 `uint8` 是 `byte` 类型的别名，我们选择使用这种转换：

```go
rowStrings := make(map[string]string)
for name, val := range row {
	switch tval := val.(type) {
	case []uint8:
		ba := make([]byte, len(tval))
		for i, v := range tval {
			ba[i] = byte(v)
		}
		rowStrings[name] = string(ba)
	default:
		return fmt.Errorf("Unknown column type %s %#v", name, spew.Sdump(val))
	}
}
```
这里我们有一个 `rowStrings`  对象表示每个返回的 SQL 行，它可以轻松的编码到 JSON 中。我们需要做的就是把它们添加到一个返回结果中，对它编码并且返回编码后的值。完整（相对短小）的代码可以在 [titpetric/sql-as-a-service](https://github.com/titpetric/sql-as-an-api) 这里获取。

### 使用须知

虽然这种方法以数据库作为 API 层的实现有独特的好处，但是为了让它适合被更广泛的调用，还有许多用户场景需要考虑。比如：

### 对结果的排序

这种方法其实不能进行排序。因为我们没法把一个查询参数绑定到 `order by`  参数中，因为 SQL 不允许这么做。数据的清洗也是完全不可能的，你甚至都不能用函数来作为一个替代方法。这样的代码是完全的不可行的：`ORDER BY :column IF(:order = 'asc', 'asc', 'desc')`。

### 参数

为了创造某种分页规则， MySQL 提供了 `LIMIT $offset, $length` 从句。虽然你可以将这些作为查询参数，但是在这里我们没法绑定它们，或者找到一种传递它们值的办法，我们尝试做的结果是得到类似这样的错误返回信息：“未定义的变量...”

### 多条 SQL 查询

通常来说，我们会执行多条 SQL 语句来返回单一的一个结果集。然而这需要在在数据库上被配置为 enabled 状态，而这个特性通常是被禁止的，其中一个主要的原因是为了防止 SQL 注入攻击。在一个理想的世界里，类似下面的语句应该是可以执行的：

```sql
set @start=:start;
set @length=:length;
select * from channels order by id desc limit @start, @length;
```

可惜，现实中它行不通。上面的语句甚至都没法远程执行。如果你在一个 MySQL 客户端中尝试执行这些语句，它将报错。这些变量被定义了，但是它们既不能在 order by 从句也不能在 limit 从句中使用。

那么，就没法分页了么？

有一个我见过的最奇怪的曲线救国的办法， Twitch 和 Tumblr 的 API 有一个特别的特性，它们允许你传递一个 `since` 或者 `previousID` 的值。它们可以类似这样的 SQL 中起作用：

```sql
select * from messages where id < :previousID order by id desc limit 0, 20
```
这让你可以按照预先定好的分页大小来遍历一个表。这种方法要求表里面有个可排序的主键（比如 sonyflake / snowflake 算法生产的 ID）或者有另外一个可用于排序的顺序列。

### 函数

SQL 数据库并不是蠢笨的野兽，实际上它们非常强大，强大的源头之一就是可以创建函数，或者，过程。用这个你可以实现一个复杂的业务逻辑。对传统的 DBA ，或数据库程序员来说，读取或者自己创建一个 SQL 函数都是很容易的。

这是一个很好的解决在客户端无法一次执行多条 SQL 语句限制的办法，但是它要求你把所有的业务逻辑在数据库中实现。这对我所知的大部分程序员来说，都是超出了他们舒适区的事儿，基本上要与一大堆的 IF 语句打交道。

## 结论

如果你真的有一些简单的查询，通过它们就能获得你需要的结果的话，那么你可以从这种以 SQL  作为 API 的方法中大大受益，除了节约系统开销外，你还可以提供给那些不熟悉 Go 语言但是熟悉 SQL  的程序员们一个增加系统功能的方法。

随着要求一个 API 返回特定结果的需求的逐渐增加，我们需要一种脚本语言，它能达到甚至超越  [PL/SQL](https://en.wikipedia.org/wiki/PL/SQL) 的种种限制，并且在不同关系型数据库中有不同的实现。
或者，你在你的应用中一直挂接一个 JavaScript 的虚机就如同 [dop251/goja ](https://github.com/dop251/goja)  这样，然后让你们团队的前台程序员来一次他/她可能永远不会忘记的尝试。现在也有用纯粹 go 语言实现的 LUA  虚机，如果你想要某种比整个 ES5.1 小的“运行时”。

### 如果我能在这里遇到你...

如果你能买一本我的书，那就太棒了：
- [API Foundations in Go](https://leanpub.com/api-foundations)
- [12 Factor Apps with Docker and Go](https://leanpub.com/12fa-docker-golang)
- [The SaaS Handbook (创作中)](https://leanpub.com/saas-handbook)

我保证你可以从中学到更多。买一本书能支持我写更多类似的文章。对你购买我的书说声谢谢。

如果你想预约我的 顾问/作家 服务欢迎[给我发送邮件](black@scene-si.org)，我擅长 APIs，Go，Docker，VueJs 和系统扩容，[以及很多其他的事情](https://scene-si.org/about)

---

via: https://scene-si.org/2018/02/07/sql-as-an-api/

作者：[Tit Petric](https://scene-si.org/about)
译者：[MoodWu](https://github.com/MoodWu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
