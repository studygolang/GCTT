# Go GraphQL初学者指南-第二部分

___注意___：关于GraphQL的系列教程的第一部分可以在[这里：Go GraphQL Beginners Tutorial - Part 1](https://tutorialedge.net/golang/go-graphql-beginners-tutorial/)阅读。

欢迎各位Gophers！在本篇教程中，我们将在上篇教程中所做的工作的基础上进行扩展，查看GraphQL API背后的变异(Mutation)以及实现合适的数据源。

在上篇教程中，我们了解了GraphQL的一些主要的优点，以及它是如何极大地改进应用程序中特定组件的检索方式。

我们研究了为特定视图构建编排的API可能造成开销的原因，以及这种技术如何降低web组件检索数据而产生的不必要的冗余信息。

___阅读资料___： 如果你不太确信GraphQL所谓的益处，我建议你阅读*Paypal Engineering*的这篇非常有趣的文章，它突出了他们在技术实践上一些成功之处 - [GraphQL - A Success Story for Paypal Checkout](https://medium.com/paypal-engineering/graphql-a-success-story-for-paypal-checkout-3482f724fb53)。

## 变异(Mutations)

如果我们需要深入了解，我们首先需要研究的就是变异。在讨论数据源问题之前，我们将先讨论这个问题，因为学习这些概念要容易的多，因为不需要在此概念的基础上编写SQL。

GraphQL中的 _变异_ 允许我们不仅可以从GraphQL API获取数据，而且能够更新它。这种修改数据的能力使得GraphQL成为一个更加完整的数据平台，可以完全取代REST，而不仅仅是它的补充。

这些 _变异_ 遵循了与我们在上篇教程中提及的查询非常相似的结构。不仅如此，在代码中定义它们同样遵循了相似的结构。

每个 _变异_  都将会映射到一个唯一的 _解析器_ 函数上，该函数将会更新我们所要求的数据。举例来说，如果有需要，我们可以创建一个 _变异_  以允许我们更新一个特定的教程，这样它就会接受到变异中的所有信息，并解析为一个唯一的 _解析器_ 函数，这个解析器函数将会执行，并且数据库中关于此篇教程的信息将会被更新。

## 一个新的教程

让我们尝试开始创建一个非常简单的 _变异_ ，这将允许我们在已有的教程列表中添加新的教程。

我们将创建一个新的GraphQL对象，这和我们之前对`fields`所做的操作一样，在这个新的对象中我们将定义`create`字段。这个`create`字段仅仅会根据传递给它的参数填充一个新的`Tutorial`，然后将这个`Tutorial`追加到全局`Tutorials`列表中。

___一些需要微调的地方___: 我们需要将`tutorials`移动到一个新的全局变量中以便于我们使用。之后这将会被替换成数据库，所以不必担心设置或修改了全局状态。

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

然后我们需要在`main()`函数中更新`schemaConfig`来引用新的`mutationType`。

```go
schemaConfig := graphql.SchemaConfig{
        Query:    graphql.NewObject(rootQuery),
        Mutation: mutationType,
}
```

现在，我们已经创建一个简单的 _变异_ 示例，并且已经更新了`schemaConfig`，接着，我们就可以尝试使用这个新的 _变异_ 。

下面我们就更新`main()`函数中的`query`来使用这个新的 _变异_ 。

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
        rJSON, _ := json.Marshal(r)
        fmt.Printf("%s \n", rJSON)
```

然后，为了检查一切是否能够按照预期的工作，我们会迅速创建并执行第二个`query`，以便于我们在调用 _变异_ 后获取存储在内存中的所有教程列表。

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
    rJSON, _ = json.Marshal(r)
    fmt.Printf("%s \n", rJSON)
`
`
```

___所有的源代码___: 本小节所有的源代码都可以在[这里：simple-mutation.go](https://gist.github.com/elliotforbes/becdd2b6d57260e88ac0698cb6c83d0b)找到。

当我们尝试运行这些代码，我们会看到 _变异_ 已经被成功地调用，而且，返回的教程列表中已经包含了我们最新定义的教程信息。

```bash
$ go run ./...
{"data":{"create":{"title":"Hello World"}}}
{"data":{"list":[{"id":1,"title":"Go GraphQL Tutorial"},{"id":2,"title":"Go GraphQL Tutorial - Part 2"},{"id":0,"title":"Hello World"}]}}
```

这非常棒，现在我们已经成功地创建了一个变异，该变异允许我们向已存在的教程列表中插入一篇新的教程。

如果我们想更进一步，我们则可以添加更多的变异来更新现有的教程，或者从教程列表中删除某些教程。我们可以只需在变异对象上创建一个新`field`，该`field`具有针对每个操作的`resolve`函数，然后在这些`resolve`函数内部，我们可以实现针对特定教程的更新/删除的相关实现。

## 更换数据源

目前为止，我们已经了解了变异，所以是时候来了解如何将内存中的数据交换到其他数据源，比如MySQL或者MongoDB。

GraphQL有一个惊人的特性，即其不受特定数据库技术的限制。我们可以创建一个GraphQL API既可以同NoSQL交互，亦可以同SQL数据库交互。

## 极简的MySQL服务器

好的，为了达到我们的教学目的，我们会将MySQL作为GraphQL数据库背后的数据存储解决方案。毫无疑问，我在之后会为会MongoDB制作一个类似的教程。但这（MySQL）应该能够展示底层的机制。

出于本教程的目的，我们将使用SQLite3本地SQL数据库快速演示如何交换更高效的数据源。

通过下面的示例创建一个新数据库：

```bash
sqlite3 tutorials.db
```

这条命令会打开一个交互式的shell，可以让我们操作和查询SQL数据库。下面，我们从创建`tutorial`表开始：

```bash
sqlite> CREATE TABLE tutorials (id int, title string, author int);
```

接着我们想向数据库中插入一些数据，以便我们可以确认`List`查询产生的变化是否生效：

```bash
sqlite> INSERT INTO tutorials VALUES (1, "First Tutorial");
sqlite> INSERT INTO tutorials VALUES (2, "Second Tutorial");
sqlite> INSERT INTO tutorials VALUES (3, "third Tutorial");
```

现在，我们已经拥有了一个SQL数据库，并包含一些数据，接着，我们就可以更新代码以使用这个新数据库。首先，我们需要在`main.go`文件中引入一个新的包，以允许我们可以和`SQLite3`通信。

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

在我们添加了包依赖之后，我们就可以在`main`函数中更新我们的`list`字段。

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

通过这些改变，我们可以查询`list`来查看它是否正常工作。在`main`函数中更新`query`，以便查询`list`并检索`id`和`title`，如下所示：

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
rJSON, _ := json.Marshal(r)
fmt.Printf("%s \n", rJSON)
```

运行代码，我们可以看到从sqlite3数据库返回的三行数据，同时我们也可以看到从GraphQL查询中返回的JSON响应体。同样，如果我们只想返回教程的`id`，我们可以修改`query`删除`title`字段，一切也会如预期一样地工作。

```bash
$ go run ./...
2018/12/30 14:44:08 {1 First Tutorial { [] } []}
2018/12/30 14:44:08 {2 Second Tutorial { [] } []}
2018/12/30 14:44:08 {3 third Tutorial { [] } []}
{"data":{"list":[{"id":1,"title":"First Tutorial"},{"id":2,"title":"Second Tutorial"},{"id":3,"title":"third Tutorial"}]}}
```

## 检索单个教程

我们已经掌握了为GraphQL API使用外部数据源的窍门，但现在让我们看一个更加简单的例子。让我们看一下更新`tutorial`模式，以使它引用我们新的sqlite3数据源。

```bash
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

我们在这里所做的就是和数据库建立一个新的连接，然后使用指定的`ID`值通过数据库查询指定的教程。

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

运行它，我们可以看到该解析器函数已成功连接至SQLite数据库，并且可以查询到相关的教程信息：

```bash
$ go run ./...
{"data":{"tutorial":{"id":1,"title":"First Tutorial"}}}
```

## 关键点

所以，重申一下，我们不再解析并返回内存中的教程列表，而是连接到数据库，执行SQL查询并从中填充教程列表。

GraphQL在我们检索结果之后接管了所有事情。如果我们想要返回每个教程的作者和评论，我们可以在SQLite数据库中创建更多的表来存储它们。然后，我们可以需要来检索相关的结果。

## 使用ORM

如果我们想进一步地简化开发工作，在和数据库交互时，我们可以选择ORM。它将为我们处理SQL查询，并处理如果有嵌套或关联元素（例如评论和作者信息）必须进行的任何连接或其他查询。

___注意___：如果你之前还未在Go中使用过ORM，我推荐你阅读我编写的其他教程：[Go ORM Tutorial](https://tutorialedge.net/golang/golang-orm-tutorial/)。

通过使用ORM来处理数据库的检索/查询，我们不仅可以简化代码，还可以无损扩展我们的API。同样的，我也创建了一个引用参考，你可以参考[这里：elliotforbes/go-graphql-tutorial](https://github.com/elliotforbes/go-graphql-tutorial/)。

## 结论

所以，我的Go GraphQL系列教程的第二部分到此结束。在本教程中，我们介绍了 _变异_ 相关的基础知识，以及如何变更支持GraphQL API的数据源以使用各种不同的数据库技术，如SQL或NoSQL数据库。

我希望你从中学到的一点就是，在实现依赖于这些API的应用程序时，GraphQL可能会对你的开发速度产生潜在的影响。

通过使用GraphQL之类的技术，我们可以将计算如何获取和解析数据所花费的时间减少到最少，然后我们可以将重点放在改进应用程序的UI/UX上。

---

via: [Go GraphQL Beginners Tutorial - Part2](https://tutorialedge.net/golang/go-graphql-beginners-tutorial-part-2/)

作者：[Elliot Forbes](https://twitter.com/elliot_f)
译者：[barryz](https://github.com/barryz)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
