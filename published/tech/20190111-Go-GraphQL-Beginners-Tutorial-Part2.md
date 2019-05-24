首发于：https://studygolang.com/articles/20653

# Go GraphQL 入门指南 - 第二部分

___注意___：关于 GraphQL 的系列教程的第一部分可以在这里：[Go GraphQL 入门指南 - Part 1](https://studygolang.com/articles/18801) 阅读。

首先，再次欢迎各位 Gophers ！在本篇教程中，我们将在上一篇教程中所做的工作基础上进行扩展，了解变更 (Mutation) 的概念，并在 GraphQL API 后端实现合适的数据源。

在上篇教程中，我们了解到 GraphQL 的一些主要的优点，以及它是如何在应用程序中极大地改进特定组件的检索方式。

我们研究了为特定的视图构建可编排的 API 可能会造成额外开销的原因，以及运用这种技术是如何帮助你降低 Web 组件检索数据而产生的冗余信息。

___阅读资料___： 如果你不太确信 GraphQL 所谓的益处，我建议你阅读**Paypal Engineering**的这篇文章，这篇文章很有趣，同时也强调了他们在一些技术实践上成功之处 - [GraphQL - A Success Story for Paypal Checkout](https://medium.com/paypal-engineering/graphql-a-success-story-for-paypal-checkout-3482f724fb53)。

## 变更 (Mutations)

最开始，我们需要深入研究并理解的概念就是变更。在讨论数据源问题之前，我们将先讨论这个问题，因为这样是不需要在变更的基础上编写 SQL 的 ，所以学习这些概念要相对容易很多。

GraphQL 中的**变更**不仅可以从 GraphQL API 获取数据，而且也可以更新 GraphQL API 内的数据。这种能够修改数据的能力使得 GraphQL 成为了一个功能完备的数据平台，甚至可以完全取代 REST ，而不仅仅是它的一个补充。

**变更**与我们在上篇教程中提到的**查询**有着非常相似的结构。不仅如此，如果我们在代码中定义它们，也同样遵循类似的规则。

事实上，每个**变更**都会映射到一个唯一的**解析器**函数上，该函数将会更新我们所需要变更的数据。举例来说，在必要的情况下，我们可以创建一个**变更**来更新一个特定的教程，此时，GraphQL 就会接受到变更中的所有信息，并解析为一个唯一的 _解析器_ 函数，接着这个解析器函数将会被执行，并更新数据库中关于此篇教程的信息。

## 创建新的教程

下面，让我们尝试开始创建一个简单的**变更**，该**变更**允许我们在已有的教程列表中增加新的教程。

我们会创建一个全新的 GraphQL 对象，这与我们之前对 `fields` 所做的操作一样，在这个新对象中我们会定义一个 `create` 字段。该字段只会根据传递给它的参数填充一个新的 `Tutorial` ，然后将这个 `Tutorial` 追加到全局 `Tutorials` 列表中。

___一些需要微调的地方___: 我们需要将 `tutorials` 移动到一个新的全局变量中以便于我们使用。在这之后我们会用数据库去替换它，所以不必担心设置或修改了全局状态。

```go
var mutationType = graphql.NewObject(graphql.ObjectConfig{
        Name: "Mutation",
        Fields: graphql.Fields{
                "create": &graphql.Field{
                        Type:        tutorialType,
                        Description: "Create a new Tutorial",
                        Args: graphql.FieldConfigArgument{
                                "title": &graphql.ArgumentConfig{
                                        Type: graphql.NewNonNull(graphql.String),
                                },
                        },
                        Resolve: func(params graphql.ResolveParams) (interface{}, error) {
                                tutorial := Tutorial{
                                        Title: params.Args["title"].(string),
                                }
                                tutorials = append(tutorials, tutorial)
                                return tutorial, nil
                        },
                },
        },
})
```

然后我们需要在 `main()` 函数中更新 `schemaConfig` 来引用新的 `mutationType` 。

```go
schemaConfig := graphql.SchemaConfig{
        Query:    graphql.NewObject(rootQuery),
        Mutation: mutationType,
}
```

现在，我们已经创建一个简单的 _变更_ 示例，并且已经更新了 `schemaConfig`，接着，我们就可以尝试使用这个新的 _变更_ 。

我们开始更新 `main()` 函数中的 `query` 来使用这个新的 _变更_ 。

```go
// Query
        query := `
            mutation {
            create(title: "Hello World") {
                            title
                }
                    }
                `
        params := graphql.Params{Schema: schema, RequestString: query}
        r := graphql.Do(params)
        if len(r.Errors) > 0 {
                log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
        }
        rJSON, _ := JSON.Marshal(r)
        fmt.Printf("%s \n", rJSON)
```

接着，为了检查一切是否能够按照预期地工作，我们会迅速创建并执行第二个 `query`，以便于我们在调用 _变更_ 后获取存储在内存中的所有教程列表。

```go
// Query
query = `
{
    list {
        id
        title
    }
}
`
params = graphql.Params{Schema: schema, RequestString: query}
r = graphql.Do(params)
    if len(r.Errors) > 0 {
            log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)

    }
    rJSON, _ = JSON.Marshal(r)
    fmt.Printf("%s \n", rJSON)
`
`
```

___完整源代码___: 本小节所有的源代码都可以在[simple-mutation.go](https://gist.github.com/elliotforbes/becdd2b6d57260e88ac0698cb6c83d0b) 找到。

当我们尝试运行这些代码，我们会看到 _变更_ 已经被成功地调用，并且，返回的教程列表中已经包含了我们最新定义的教程信息。

```bash
$ go run ./...
{"data":{"create":{"title":"Hello World"}}}
{"data":{"list":[{"id":1,"title":"Go GraphQL Tutorial"},{"id":2,"title":"Go GraphQL Tutorial - Part 2"},{"id":0,"title":"Hello World"}]}}
```

这非常棒，目前为止我们已经成功地创建了一个变更，该变更允许我们向已存在的教程列表中插入一篇新的教程。

如果我们想更进一步，我们则可以添加更多的变更来更新现有的教程，或从教程列表中删除某些教程。我们只需在变更对象上创建一个新 `field`，该 `field` 具有针对每个操作的 `resolve` 函数，然后在这些 `resolve` 函数内部，我们可以实现针对特定教程的更新 / 删除的相关实现。

## 更换数据源

到目前为止，我们已经了解了变更的概念，所以是时候来了解如何将内存数据源更换到其他数据源，比如 MySQL 或者 MongoDB。

GraphQL 有一个奇妙的特性，即其不受特定数据库技术的限制。我们可以创建一个 GraphQL API ，它既可以同 NoSQL 交互，也可以同 SQL 数据库交互。

## 简单的 MySQL 服务器

好的，为了达到我们的教学目的，我们会将 MySQL 作为 GraphQL 数据库背后的数据存储解决方案。毫无疑问，我在之后也会为 MongoDB 制作一个类似的教程。但 MySQL 的案例应该能够展示底层的机制。

出于本教程的最终目的，我们将使用 SQLite3 本地 SQL 数据库快速演示如何切换成更高效的数据源。

通过下面的示例创建一个新数据库：

```bash
sqlite3 tutorials.db
```

这条命令会打开一个交互式的 shell，可以让我们操作和查询 SQL 数据库。接下来，让我们先创建 `tutorial` 表：

```bash
sqlite> CREATE TABLE tutorials (id int, title string, author int);
```

接着我们会向数据库中插入一些数据，以便我们可以确认 `List` 查询产生的变化是否生效：

```bash
sqlite> INSERT INTO tutorials VALUES (1, "First Tutorial");
sqlite> INSERT INTO tutorials VALUES (2, "Second Tutorial");
sqlite> INSERT INTO tutorials VALUES (3, "third Tutorial");
```

现在，我们已经拥有了一个 SQL 数据库，并包含一些数据，紧接着，我们就可以更新代码来使用这个新数据库。首先，我们需要在 `main.go` 文件中引入一个新的包，以允许我们可以和 `SQLite3` 交互。

```go
package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"

    "github.com/graphql-go/graphql"
    _ "github.com/mattn/go-sqlite3"

)
...
```

在我们添加了包依赖之后，我们就可以在 `main` 函数中更新我们的 `list` 字段。

```go
"list": &graphql.Field{
        Type:        graphql.NewList(tutorialType),
        Description: "Get Tutorial List",
        Resolve: func(params graphql.ResolveParams) (interface{}, error) {
            db, err := sql.Open("sqlite3", "./foo.db")
            if err != nil {
                log.Fatal(err)
            }
            defer db.Close()
            // perform a db.Query insert
            var tutorials []Tutorial
            results, err := db.Query("SELECT * FROM tutorials")
            if err != nil {
                fmt.Println(err)
            }
            for results.Next() {
                var tutorial Tutorial
                err = results.Scan(&tutorial.ID, &tutorial.Title)
                if err != nil {
                    fmt.Println(err)
                }
                log.Println(tutorial)
                tutorials = append(tutorials, tutorial)
            }
            return tutorials, nil
        },
},
```

通过这些改变，我们可以查询 `list` 来观察它是否正常工作。在 `main` 函数中更新 `query`，以便查询 `list` 并检索 `id` 和 l`title`，如下所示：

```go
// Query
query := `
    {
        list {
            id
            title
        }
    }
`
params := graphql.Params{Schema: schema, RequestString: query}
r := graphql.Do(params)
if len(r.Errors) > 0 {
        log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)

}
rJSON, _ := JSON.Marshal(r)
fmt.Printf("%s \n", rJSON)
```

运行该代码，我们可以看到从 SQLite3 数据库返回的三行数据，同时我们也可以看到从 GraphQL 查询中返回的 JSON 响应体。同样地，如果我们只想返回教程的 `id`，我们可以修改 `query` 删除 `title` 字段，一切也会如预期一样地工作。

```bash
$ go run ./...
2018/12/30 14:44:08 {1 First Tutorial { [] } []}
2018/12/30 14:44:08 {2 Second Tutorial { [] } []}
2018/12/30 14:44:08 {3 third Tutorial { [] } []}
{"data":{"list":[{"id":1,"title":"First Tutorial"},{"id":2,"title":"Second Tutorial"},{"id":3,"title":"third Tutorial"}]}}
```

## 检索单个教程

我们已经掌握了为 GraphQL API 使用外部数据源的诀窍，那么现在让我们看一个更加简单的例子。试着更新 `tutorial` 模式，以使它引用我们新的 SQLite3 数据源。

```go
"tutorial": &graphql.Field{
        Type:        tutorialType,
            Description: "Get Tutorial By ID",
            Args: graphql.FieldConfigArgument{
                "id": &graphql.ArgumentConfig{
                    Type: graphql.Int,
                },
            },
        Resolve: func(p graphql.ResolveParams) (interface{}, error) {
            id, ok := p.Args["id"].(int)
            if ok {
                db, err := sql.Open("sqlite3", "./tutorials.db")
                if err != nil {
                    log.Fatal(err)
                }
                defer db.Close()
                var tutorial Tutorial
                err = db.QueryRow("SELECT ID, Title FROM tutorials where ID = ?", id).Scan(&tutorial.ID, &tutorial.Title)
                if err != nil {
                    fmt.Println(err)
                }
                return tutorial, nil
            }
            return nil, nil
         },
},
```

在这里，我们所做的事情就是和数据库建立一个新的连接，然后使用指定的 `ID` 值通过数据库查询指定的教程。

```go
// Query
query := `
{
    tutorial(id: 1) {
        id
        title
    }
}
`
```

运行代码，我们可以观察到该解析器函数已成功连接至 SQLite 数据库，并可以查询到相关的教程信息：

```bash
$ go run ./...
{"data":{"tutorial":{"id":1,"title":"First Tutorial"}}}
```

## 关键点

所以，重申一下，我们不再需要解析并返回内存中的教程列表，而是通过连接到数据库，执行 SQL 查询并从中填充教程列表。

GraphQL 在我们检索结果之后接管了所有事情。如果我们想要返回每个教程的作者和评论，我们可以在 SQLite 数据库中创建更多的表来存储它们。然后，我们可以简单地对数据库执行其他的 SQL 查询，以根据 ID 和 comment 检索作者。

## 使用 ORM

如果我们想进一步地简化开发工作，在和数据库交互时，我们可以选择 ORM。它将为我们处理 SQL 查询，并在我需要嵌入或级联元素时（例如评论和作者信息），处理任何必须进行的连接或者其他查询。

___注意___：如果你之前还未在 Go 中使用过 ORM，我推荐你阅读我编写的其他教程：[Go ORM Tutorial](https://tutorialedge.net/golang/golang-orm-tutorial/)。

通过使用 ORM 来处理数据库的检索查和查询，我们不仅可以简化代码，还可以无损扩展我们的 API。同样的，我也创建了一个引用参考，你可以参考这里：[elliotforbes/go-graphql-tutorial](https://github.com/elliotforbes/go-graphql-tutorial/)。

## 结论

到此为止，我的 Go GraphQL 系列教程的第二部分到此结束。在本教程中，我们介绍了 _变更_ 相关的基础知识，以及如何切换支持 GraphQL API 的数据源以使用各种不同的数据库技术，比如 SQL 或 NoSQL 数据库。

我希望你从中学到的一点就是，在实现依赖于这些 API 的应用程序时，GraphQL 可能会对你的开发速度产生潜在的影响。

通过使用 GraphQL 之类的技术，我们可以将花费在思考如何获取和解析数据上的时间减少到最少，然后我们可以将重点放在改进应用程序的 UI/UX 上。

---

via: [Go GraphQL Beginners Tutorial - Part2](https://tutorialedge.net/golang/go-graphql-beginners-tutorial-part-2/)

作者：[Elliot Forbes](https://twitter.com/elliot_f)
译者：[barryz](https://github.com/barryz)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
