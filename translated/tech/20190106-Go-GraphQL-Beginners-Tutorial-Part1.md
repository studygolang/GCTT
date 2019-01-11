# Go GraphQL初学者指南

欢迎各位Gophers！在本教程中，我们将探索如何使用Go程序和GraphQL服务器进行交互。在本教程完结时，我们希望可以了解以下内容：

- GraphQL的基础知识
- 使用Go构建一个简单的GraphQL服务器
- 基于GraphQL执行一些基本的查询

在本教程中，我们将专注于GraphQL的数据检索方面，并且我们将使用一个内存数据源来存储它。这将为我们在后续教程中提供一个良好的基础。

## GraphQL基础知识

好的，在我们深入之前，我们应该真正地理解GraphQL的基础知识。作为开发人员，使用它对我们来说会有哪些益处。

好吧，考虑下如果有一个系统每天需要处理数十万甚至数百万的请求。传统上，我们会遇到一个面向数据库的API，它会返回一个巨大的JSON响应体，其中包含了许多不必要的冗余信息。

如果我们正在大规模地处理应用程序，那么发送冗余的信息会产生很多额外的开销。并且由于负载的关系可能会阻塞网络带宽。

GraphQL 本质上允许我们减少相关的噪音（不必要的内容）并能够描述我们希望从服务端返回的数据，这样我们就可以 ***仅仅*** 检索当前视图/任务/其他所需的内容。

而这只是GraphQL能提供给我们的众多好处之一。希望在接下来的教程中，我们可以看到更多的好处。

## 为API而生的查询语言，而不是为数据库而生

一个非常重要的内容就是GraphQL并不和传统的数据库查询语言SQL一样。它是位于API之前的抽象，且 ***并不*** 依赖于任何特定的数据或存储引擎。

这真的非常棒，我们可以建立一个与现有服务交互的GraphQL服务器，然后围绕这个新的GraphQL进行构建，而无须担心会修改现有的RESTful API。

## REST和GraphQL的区别

让我们先看看RESTful方法和GraphQL方法的区别。现在，想象下我们在构建一个能返回本站所有教程的服务，如果我们需要某些特定的教程信息，通常我们会创建一个API端点以允许我们以一个ID来检索特定的教程。

```bash
# A dummy endpoint that takes in an ID path parameter
'http://api.tutorialedge.net/tutorial/:id'
```

这将会返回一个响应，如果给定的是一个合法的`ID`，那么该响应体可能会如下所示：

```js
{
    "title": "Go GraphQL Tutorial",
    "Author": "Elliot Forbes",
    "slug": "/golang/go-graphql-beginners-tutorial/",
    "views": 1,
    "key" : "value"
}
```

现在，假设我们想创建一个小部件，列出由该作者撰写的前5个帖子。我们可以使用`/author/:id`API端点来检索所有由该作者撰写的帖子，然后再执行之后的调用来获取前5个帖子。亦或者，我们可以创建一个新的端点来返回这些数据。

这两种解决方案听起来并没有特别吸引人的地方，因为它们创建了大量无用的请求或者返回了过多的冗余信息，这突出了RESTful方法在这一方面的缺陷。

这就是GraphQL发挥作用的地方。通过GraphQL，我们可以在查询中精确定义我们想要返回的数据。所以，如果我们需要上述的教程信息，我们可以创建一个如下所示的查询：

```js
{
    tutorial(id: 1) {
        id
        title
        author {
            name
            tutorials
        }
        comments {
            body
        }
    }
}
```

随后，它就会返回我们所需的数据，该教程的作者以及指定`id`的教程的作者所撰写的其他教程列表，而不用发送额外多的REST请求来获取信息！多美好，不是吗？

## 基本设置

好的，现在我们已经了解了一点关于GraphQL的基本信息以及使用它的好处，让我们看看如何在实战中使用它。

我们将会使用Go创建一个简单的GraphQL服务器，我们使用[graphql-go/graphql](https://github.com/graphql-go/graphql)这个库实现。

## 设置简易GraphQL服务器

使用`go mod init`来初始化我们的项目作为开始：

```bash
$ go mod init github.com/elliotforbes/go-graphql-tutorial
```

接下来，让我们创建一个名为`main.go`的文件。我们将从头开始创建一个简单的GraphQL服务器，它有一个非常简单的解析器。

```go
// credit - go-graphql hello world example
package main

import (
        "encoding/json"
        "fmt"
        "log"

        "github.com/graphql-go/graphql"
)

func main() {
        // Schema
        fields := graphql.Fields{
                "hello": &graphql.Field{
                        Type: graphql.String,
                        Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                                return "world", nil
                        },
                },
        }
        rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
        schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
        schema, err := graphql.NewSchema(schemaConfig)
        if err != nil {
                log.Fatalf("failed to create new schema, error: %v", err)
        }
        // Query
        query := `
        { hello
    }
        `
        params := graphql.Params{Schema: schema, RequestString: query}
        r := graphql.Do(params)
        if len(r.Errors) > 0 {
                log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
        }
        rJSON, _ := json.Marshal(r)
        fmt.Printf("%s \n", rJSON) // {“data”:{“hello”:”world”}}
}
```

现在，如果我们尝试运行它，让我们看看会发生什么：

```bash
$ go run ./...
{"data":{"hello":"world"}}
```

所以，如果没什么问题的话，那么我们已经设置了一个真正的简易GraphQL服务器并且创建了一个真实的请求并发送至该服务器。

## GraphQL模式(Schema)

让我们来分解下上述代码所发生的事情以便我们可以进一步扩展它。在`fields...`代码行开始处，我们定义了`模式`。当我们通过GraphQL API执行查询时，我们实际上定义了我们希望返回的某些数据，所以我们不得不在模式中定义这些字段。

在`Resolve...`代码行处，我们定义了一个每当这个特定`字段`被请求时会触发的解析器函数。现在，我们仅仅返回了一个`"workd"`字符串，但我们将从这里开始实现查询数据库的能力。

## 查询

让我们继续观察`main.go`文件的第二部分。在`query...`代码行开始处我们定义了一个请求`hello`字段的`query`。

然后，我们创建一个包含对我们定义的`schema`以及`RequestString`请求的引用的`params`结构。

最后，我们执行请求，并将请求的结果填充到`r`中。然后我们处理一些错误并将响应体解析到JSON中，将其打印到终端上。

## 一个更复杂的例子

现在我们已经拥有一个启动并运行着的极简GraphQL服务器，并且我们可以通过它来执行一些查询，让我们来更进一步，构建一个更加复杂的例子。

我们将会创建一个GraphQL服务器，它返回一系列内存中的教程以及它们的作者，以及对这些特定教程所做的任何评论。

让我们先定义能够表示`tutorial`，`Author`和`Comment`的结构:

```go
type Tutorial struct {
        Title    string
        Author   Author
        Comments []Comment
}

type Author struct {
        Name      string
        Tutorials []int
}

type Comment struct {
        Body string
}
```

接着我们可以创建一个非常简单的`populate()`函数用于返回`tutorial`类型的切片。

```go
func populate() []Tutorial {
        author := &Author{Name: "Elliot Forbes", Tutorials: []int{1}}
        tutorial := Tutorial{
                ID:     1,
                Title:  "Go GraphQL Tutorial",
                Author: *author,
                Comments: []Comment{
                Comment{Body: "First Comment"},
        },
}

        var tutorials []Tutorial
        tutorials = append(tutorials, tutorial)

        return tutorials
}
```

这将为我们提供一个简单的教程列表，以便于我们之后的解析。

## 创建一个新的对象类型

我们首先使用`graphql.NewObject()`在GraphQL中创建一个新对象。我们将会使用GraphQL严格的类型定义三种不同的类型。这些类型将与我们已经定义好的三种`structs`相匹配。

`Comment`结构无疑是最简单的，它只包含一个字符串类型的字段`Body`，所以我们能够很容易地将其表示为`commentType`：

```go
var commentType = graphql.NewObject(
        graphql.ObjectConfig{
        Name: "Comment",
        // we define the name and the fields of our
        // object. In this case, we have one solitary
        // field that is of type string
                Fields: graphql.Fields{
                        "body": &graphql.Field{
                                Type: graphql.String,
                        },
                },
        },
)
```

接下来，我们将处理`Author`结构并将其定义为新的`graphql.NewObject()`，这将会稍微复杂点，因为它既包含`String`字段，也包含一个`Int`值列表，这些值表示他们所编写的教程ID。

```go
var authorType = graphql.NewObject(
        graphql.ObjectConfig{
                Name: "Author",
                Fields: graphql.Fields{
                        "Name": &graphql.Field{
                                Type: graphql.String,
                        },
                        "Tutorials": &graphql.Field{
                // we'll use NewList to deal with an array
                // of int values
                                Type: graphql.NewList(graphql.Int),
                        },
                },
        },
)
```

最后，让我们定义`tutorialType`，它将封装一个`author`，一个元素为`comment`的切片，`ID`以及`title`。

```go
var tutorialType = graphql.NewObject(
        graphql.ObjectConfig{
                Name: "Tutorial",
                Fields: graphql.Fields{
                        "id": &graphql.Field{
                                Type: graphql.Int,
                        },
                        "title": &graphql.Field{
                                Type: graphql.String,
                        },
                        "author": &graphql.Field{
                // here, we specify type as authorType
                // which we've already defined.
                // This is how we handle nested objects
                                Type: authorType,
                        },
                        "comments": &graphql.Field{
                                Type: graphql.NewList(commentType),
                        },
                },
        },
)
```

## 更新模式(Schema)

现在我们已经定义了`Type`系统，接下来让我们开始更新我们的`Schema`以反映到这些类型上。我们将定义两个不同的`Field`，第一个是我们的`tutorial`字段，它允许我们根据传入的查询ID检索单个`tutorial`。而第二个字段将会是一个`list`，它允许我们检索我们已经定义在内存中的`tutorials`完整列表。

```go
// Schema
        fields := graphql.Fields{
                "tutorial": &graphql.Field{
                        Type: tutorialType,
                        // it's good form to add a description
                        // to each field.
                        Description: "Get Tutorial By ID",
                        // We can define arguments that allow us to
                        // pick specific tutorials. In this case
                        // we want to be able to specify the ID of the
                        // tutorial we want to retrieve
                        Args: graphql.FieldConfigArgument{
                                "id": &graphql.ArgumentConfig{
                                        Type: graphql.Int,
                                },
                        },
                        Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                                // take in the ID argument
                                id, ok := p.Args["id"].(int)
                                if ok {
                                        // Parse our tutorial array for the matching id
                                        for _, tutorial := range tutorials {
                                                if int(tutorial.ID) == id {
                                                        // return our tutorial
                                                        return tutorial, nil
                                                }
                                        }
                                }
                                return nil, nil
                        },
                },
                // this is our `list` endpoint which will return all
                // tutorials available
                "list": &graphql.Field{
                        Type:        graphql.NewList(tutorialType),
                        Description: "Get Tutorial List",
                        Resolve: func(params graphql.ResolveParams) (interface{}, error) {
                                return tutorials, nil
                        },
                },
        }
```

所以我们创建了类型并更新了GraphQL模式(Schema)，我们做的还算不错。

## 测试它是否能够工作

让我们尝试使用新的GraphQL服务器，并使用我们提交的查询。让我们通过更改`main()`函数中的`query`来尝试`list`模式。

```go
// Query
query := `
    {
        list {
            id
            title
            comments {
                body
            }
            author {
                Name
                Tutorials
            }
        }
    }
`
```

让我们来分析下，在我们的查询中我们有一个特殊的`root`对象。在这里，我们说我们想要的那个对象的`list`字段，在按`list`返回的列表中，我们希望能看到`id`，`title`，`comments` 和`author`。

而当我们运行它时，我们将会看到如下输出：

```bash
$ go run ./...
{"data":{"list":[{"author":{"Name":"Elliot Forbes","Tutorials":[1]},"comments":[{"body":"First Comment"}],"id":1,"title":"Go GraphQL Tutorial"}]}}
```

正如我们所看到的，我们的查询以JSON形式返回了所有教程，看起来和初始的查询非常相似。

现在让我们通过`tutorial`模式来执行一个查询:

```go
query := `
    {
        tutorial(id:1) {
            title
            author {
                Name
                Tutorials
            }
        }
    }
`
```

再一次，当我们运行它，我们会看到它成功地检索到了内从中`ID=1`的单独的教程。

```bash
$ go run ./...
{"data":{"tutorial":{"author":{"Name":"Elliot Forbes","Tutorials":[1]},"title":"Go GraphQL Tutorial"}}}
```

完美，看起来我们的`list`和`tutorial`模式能够正常工作。

## 结论

> 注意： 本教程完成的源代码位于[这里：main.go](https://gist.github.com/elliotforbes/9b8400ef5154eb3420e409aeffe39633)

这就是我们将在本次初始教程中所介绍的内容。我们成功地设置了一个由内存数据存储支持地简单的GraphQL服务器。

在下篇教程中，我们将查看GraphQL变异并更改我们的数据源以使用NoSQL数据库。下一篇教程可以阅读[Go GraphQL Beginners Tutorial - Part2](https://tutorialedge.net/golang/go-graphql-beginners-tutorial-part-2/)

---

via: [Go GraphQL Beginners Tutorial - Part1](https://tutorialedge.net/golang/go-graphql-beginners-tutorial/)

作者：[Elliot Forbes](https://twitter.com/elliot_f)
译者：[barryz](https://github.com/barryz)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
