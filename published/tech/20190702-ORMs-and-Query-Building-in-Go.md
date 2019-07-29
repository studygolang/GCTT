首发于：https://studygolang.com/articles/22296

# 在 Go 中的 ORM 和查询构建

2019.07.13 星期六 发表于 [Programming](https://andrewpillar.com/programming/)

最近，我一直在研究 Go 中与数据库交互的各种解决方案。在 Go 中与数据库交互我使用的底层库是 sqlx。你只需要写出 SQL，并使用 db tag 标记结构，之后让 sqlx 处理其余工作。但是，我遇到的主要问题是符合语法习惯的查询构建。这让我调查了这个问题，并在本文中记下了一些想法。

在 Go 中，第一类函数是进行 SQL 查询构建的惯用方法。该仓库包含我编写的一些示例代码：https://github.com/andrewpillar/query。

## GORM、分层复杂性和 Active Record 模式

在 Go 中，大多数涉足数据库工作的人更可能使用 Gorm。 Gorm 是一个功能完备的 ORM，支持迁移、关系、事务等等。对于那些使用过 ActiveRecord 或 Eloquent 的人来说，Gorm 的用法应该很熟悉。

我以前简单地使用过 Gorm，对于简单的、基于 CRUD 的应用程序，这很好。然而，当涉及更多分层复杂性时，我发现它做的并不好。假设我们正在构建一个博客应用程序，我们允许用户通过 URL 中的查询字符串搜索帖子。如果存在这种情况，我们希望用的约束条件：WHERE title LIKE。

```go
posts := make([]Post, 0)
search := r.URL.Query().Get("search")
db := Gorm.Open("postgres", "...")
if search != "" {
	db = db.Where("title LIKE ?", "%" + search + "%")
}
db.Find(&posts)
```

没有什么可争议的，我们只是检查是否有值并修改对 Gorm 本身的调用。但是，如果我们想在特定日期之后搜索帖子怎么办？我们需要添加一些检查，首先查看 URL 中是否存在关于日期的查询字符串 (after)，如果存在则修改查询条件。

```go
posts := make([]Post, 0)

search := r.URL.Query().Get("search")
after := r.URL.Query().Get("after")

db := Gorm.Open("postgres", "...")
if search != "" {
	db = db.Where("title LIKE ?", "%" + search + "%")
}
if after != "" {
	db = db.Where("created_at > ?", after)
}
db.Find(&posts)
```

如上所示，我发现 GORM 的最大缺点是处理分层的复杂逻辑时十分繁琐。但通常情况下，编写 SQL 时你会想要这样做。试想你在查询中根据用户输入添加一个 Where 条件或者决定如何排序记录。

我相信这归结为一件事，对此我前段时间在 HN 做了一个[评论](https://news.ycombinator.com/item?id=19851753)：

> 就我个人而言，我认为基于 ORM 的 Active Record 风格，对 Go 而言类似 Gorm，并不适合本身就不是 OOP 的语言。仔细翻阅 Gorm 的相关文档，Gorm 似乎严重依赖方法链，这对 Go 而言似乎也是错误的：考虑一下 Go 语言中如何处理 error。我认为，ORM 应该尽可能与语法习惯保持一致。

此评论提交在博客 [To ORM or not to ORM](https://eli.thegreenplace.net/2019/to-orm-or-not-to-orm/) 上，我强烈建议您阅读该文。该文作者在 Gorm 问题上得出了与我的一致结论。

## 在 Go 中符合语法习惯的查询构建

标准库中包 database/sql 非常适合与数据库交互。sqlx 是基于此的、处理返回数据的一个优秀扩展。但是，这仍然没有完全解决手头的问题。我们如何高效地、程序化地、符合 Go 语法习惯地构建复杂的查询 ? 假设我们使用 sqlx 进行上述相同的查询，那是什么样子呢？

```go
posts := make([]Post, 0)

search := r.URL.Query().Get("search")
after := r.URL.Query().Get("after")

db := sqlx.Open("postgres", "...")

query := "SELECT * FROM posts"
args := make([]interface{}, 0)

if search != "" {
	query += " WHERE title LIKE ?"
	args = append(args, search)
}

if after != "" {
	if search != "" {
		query += " AND "
	} else {
		query += " WHERE "
	}

	query += "created_at > ?"

	args = append(args, after)
}

err := db.Select(&posts, sqlx.Rebind(query), args...)
```

没有比我们用 Gorm 做的好多少，事实上更加丑陋。我们检查了两次 search 是否存在，以便我们可以为查询提供正确的 SQL 语法，我们将参数存储在 []interface{} 切片中，我们将 SQL 拼接到一个字符串。这些同样都是不可扩展、不易于维护的。

理想情况下，我们希望能够构建查询，并将其交给 sqlx 来处理其余的事情。那么，Go 中的符合语法习惯的查询构建器会是什么样子？嗯，在我看来，它将采用两种形式之一，第一种是利用可选的结构体，另一种利用第一类函数。

我们来看看 [squirrel](https://github.com/masterminds/squirrel)。这个库提供了构建查询的能力，并以我认为更加符合 Go 语法习惯的方式直接执行它们。当然，在此我们只关注查询构建方面。

使用 squirrel，我们可以像这样实现上面的逻辑。

```go
posts := make([]Post, 0)

search := r.URL.Query().Get("search")
after := r.URL.Query().Get("after")

eqs := make([]sq.Eq, 0)

if search != "" {
	eqs = append(eqs, sq.Like{"title", "%" + search + "%"})
}

if after != "" {
	eqs = append(eqs, sq.Gt{"created_at", after})
}

q := sq.Select("*").From("posts")

for _, eq := range eqs {
	q = q.Where(eq)
}

query, args, err := q.ToSql()

if err != nil {
	return
}

err := db.Select(&posts, query, args...)
```

这比我们使用 Gorm 时要好一些，并且比我们之前做的字符串连接要好几英里。然而，编写仍然稍显繁琐。对 SQL 查询中的某些子句，squirrel 使用可选的结构体表示。在 Go 中对于 API，可选结构体是一种常见模式，旨在实现高度可配置。

Go 中用于查询构建的 API 应满足以下两个需求：

* 符合语法习惯
* 可扩展

如何用 Go 实现这一目标？

## 查询构建：第一类函数

Dave Cheney 根据 Rob Pike 关于同一主题的帖子撰写了两篇关于第一类函数的博客文章。感兴趣可以找到如下原文：

* [Self-referential functions and the design of options](https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html)
* [Functional options for friendly APIs](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
* [Do not fear the first class functions](https://dave.cheney.net/2016/11/13/do-not-fear-first-class-functions)

我强烈建议阅读以上三篇文章，并在你下次实现高度可配置的 API 时使用他们所建议的模式。

下面是查询构建的示例：

```go
posts := make([]*Post, 0)

db := sqlx.Open("postgres", "...")

q := Select(
	Columns("*"),
	Table("posts"),
)

err := db.Select(&posts, q.Build(), q.Args()...)
```

我知道，这是一个很简单的例子。但是让我们来看看我们如何实现这样的 API，以便它可以用于查询构建。首先，我们应该实现一个查询结构来跟踪其在构建时的状态。

```go
type statement uint8

type Query struct {
	stmt  statement
	table []string
	cols  []string
	args  []interface{}
}

const (
	_select statement = iota
)
```

上述结构将跟踪我们正在构建的语句，包括 SELECT、UPDATE、INSERT、DELETE 等等，追踪我们正在操作的表，追踪我们正在使用的列，追踪将被传递到最终的查询语句中的参数。为了简单起见，让我们专注于实现 SELECT 语句的查询构建器。

接下来，我们需要定义一种类型，该类型可用于修改我们正在构建的查询，该类型作为第一类函数将被多次传递。每次调用此函数时，它都应返回新的、被修改后的查询（如果适用）。

```go
type Option func(q Query) Query
```

我们现在可以实现构建器的第一部分 : Select 函数。 这将开始为我们的 SELECT 语句构建查询。

```go
func Select(opts ...Option) Query {
	q := Query{
		stmt: select_,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

```

您现在应该能够看到一切如何慢慢地汇聚到一起，以及 UPDATE、INSERT、DELETE 等语句怎样简单的构建查询。如果没有实际实现一些 options 并传递到 Select 函数中，上面的 Select 函数是完全无用的，我们来继续实现。

```go
func Columns(cols ...string) Option {
	return func(q Query) Query {
		q.cols = cols

		return q
	}
}

func Table(table string) Option {
	return func(q Query) Query {
		q.table = table

		return q
	}
}
```

如您所见，我们以某种方式实现这些第一类函数，它们返回以后将被调用的、基础的 Option 函数。通常期望 Option 函数修改传递给它的 Query 对象，并返回 Query 的一个副本。

为了对我们构建复杂查询的用例有用，我们应该实现 WHERE 向查询添加子句的功能。这将要求必须跟踪 WHERE 查询中的各种子句。

```go
type where struct {
	col string
	op  string
	val interface{}
}

type Query struct {
	stmt   statement
	table  []string
	cols   []string
	wheres []where
	args   []interface{}
}
```

我们为 WHERE 子句定义了一种自定义类型，并向原始 Query 结构添加一个属性 wheres。让我们为我们的需求实现两种类型的 Where 子句，第一种是 WHERE LIKE，另一种是 `WHERE >`。

```go
func WhereLike(col string, val interface{}) Option {
	return func(q Query) Query {
		w := where{
			col: col,
			op:  "LIKE",
			val: fmt.Sprintf("$%d", len(q.args) + 1),
		}

		q.wheres = append(q.wheres, w)
		q.args = append(q.args, val)

		return q
	}
}

func WhereGt(col string, val interface{}) Option {
	return func(q Query) Query {
		w := where{
			col: col,
			op:  ">",
			val: fmt.Sprintf("$%d", len(q.args) + 1),
		}

		q.wheres = append(q.wheres, w)
		q.args = append(q.args, val)

		return q
	}
}
```

在处理 WHERE 向查询添加子句时，我们需适当地为底层 SQL 驱动程序处理绑定值的语法。我们的示例是 Postgres，将实际值本身存储到 Query 的 args 切片中。

因此，我们实现的如此少，并能够以符合语法习惯的方式实现我们期待的功能。

```go
posts := make([]Post, 0)

search := r.URL.Query().Get("search")
after := r.URL.Query().Get("after")

db := sqlx.Open("postgres", "...")

opts := []Option{
	Columns("*"),
	Table("posts"),
}

if search != "" {
	opts = append(opts, WhereLike("title", "%" + search + "%"))
}

if after != "" {
	opts = append(opts, WhereGt("created_at", after))
}

q := Select(opts...)

err := db.Select(&posts, q.Build(), q.Args()...)
```

稍好一点，但仍然不是很好。但是，我们可以扩展功能以获得我们想要的功能。因此，让我们实现一些函数，这些函数将返回满足我们特定需求的 Option。

```go
func Search(col, val string) Option {
	return func(q Query) Query {
		if val == "" {
			return q
		}

		return WhereLike(col, "%" + val + "%")(q)
	}
}

func After(val string) Option {
	return func(q Query) Query {
		if val == "" {
			return q
		}

		return WhereGt("created_at", val)(q)
	}
}
```

通过实现上述两个函数，我们现在可以干净地为我们的用例构建一个稍微复杂的查询。如果传递给它们的值被认为是正确的，这两个函数都只会修改查询。

```go
posts := make([]Post, 0)

search := r.URL.Query().Get("search")
after := r.URL.Query().Get("after")

db := sqlx.Open("postgres", "...")

q := Select(
	Columns("*"),
	Table("posts"),
	Search("title", search),
	After(after),
)

err := db.Select(&posts, q.Build(), q.Args()...)
```

我发现在 Go 中这是构建复杂查询时一种更加符合语法习惯的方式。现在，当然你已经在帖子中做到了这一点，并且必须想知道，“这很好但你没有实现 Build（）和 Args（）方法”。在某种程度上，这是事实。为了不再延长这个帖子 (已经足够），我没有打扰。所以，如果您对这里介绍的一些想法感兴趣，请查看我提交在 GitHub 的 [Code](https://github.com/andrewpillar/query)。它没有任何严谨性，它没有覆盖查询构建器所需的所有内容，它也缺少 Join 字句；仅仅是为示例并支持 Postgres 绑定值语法。

如果您对我在本文中所说的内容有任何不同意见，或想进一步讨论，请通过邮件 me@andrewpillar.com 与我联系。

---

via: https://andrewpillar.com/programming/2019/07/13/orms-and-query-building-in-go/

作者：[andrewpillar](https://github.com/andrewpillar)
译者：[zhoudingding](https://github.com/DingdingZhou/)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
