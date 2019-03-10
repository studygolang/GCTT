首发于：https://studygolang.com/articles/18801

# Go GraphQL 入门指南

欢迎各位 Gophers ！在本教程中，我们将探索如何使用 Go 和 GraphQL 服务进行交互。在本教程完结之时，我们希望你可以了解到以下内容：

- GraphQL 的基础知识
- 使用 Go 构建一个简易的 GraphQL 服务
- 基于 GraphQL 执行一些基本的查询

在本篇教程中，我们会专注于 GraphQL 在数据检索方面的内容，我们将会使用内存数据源来存储其数据。同时，本篇教程的内容将会为我们之后的教程提供一个良好的基础。

## GraphQL 的基础知识

在我们深入探讨之前，我们需要真正地理解 GraphQL 相关的基础知识。换句话说，作为开发人员，我们需要知道，使用它会为我们带来哪些益处。

考虑下，如果有一个系统，它每天需要处理数十万甚至数百万的请求。一般情况下，我们会请求一个面向数据库的 API ，该 API 会返回大量的 JSON 响应体，其中包含了许多冗余信息。

此时，如果我们面对的是一个超大规模的应用程序，那么发送、接收这些冗余信息会产生很多额外的开销。并且会由于负载的关系而阻塞带宽。

事实上，GraphQL 能够减少我们因冗余信息而产生的额外心智负担，并且，它拥有能够描述从服务端返回的数据的能力，这样，我们就可以仅关注当前任务，视图，或其他任何东西所需的数据、内容。

而且，上述的功能仅仅是 GraphQL 能提供给我们的众多益处中的很小的一部分。在接下来的教程中，我们将会介绍更多关于 GraphQL 的益处。

## 为 API 而生（而不是为数据库而生）的查询语言

一个非常重要的内容就是 GraphQL 不是一个和传统 SQL 一样的查询语言。它是位于 API 之前的一层抽象，且**并不**依赖于任何特定的数据或存储引擎。

这种设计非常酷，我们可以先建立一个与现有服务交互的 GraphQL 服务，然后围绕这个 GraphQL 进行构建，而无须担心会修改现有的 RESTful API。

## REST 和 GraphQL 的区别

首先，让我们看看 RESTful 方法和 GraphQL 方法的区别。想象下我们正在构建一个能返回本站所有教程的服务，如果我们需要某些指定的教程信息，通常来说，我们会创建一个 API 端点以允许我们以一个 ID 来检索指定的教程。

```bash
# A dummy endpoint that takes in an ID path parameter
'http://api.tutorialedge.net/tutorial/:id'
```

如果给定的是一个合法的 `ID` ，它将会返回一个响应体，该响应体可能会如下所示：

```json
{
    "title": "Go GraphQL Tutorial",
    "Author": "Elliot Forbes",
    "slug": "/golang/go-graphql-beginners-tutorial/",
    "views": 1,
    "key" : "value"
}
```

现在，假设我们想创建一个控件，该控件会列出指定的作者撰写的前 5 个帖子。我们可以使用 `/author/:id` 这样的 API 来检索出所有由该作者撰写的帖子，然后再执行后续的调用获取排名前 5 的帖子。亦或者，我们可以创建一个新的 API 来返回这些数据。

上述的解决方案听起来并没有什么特别之处，因为它们创建了大量无用的请求或者返回了过多的冗余信息，这也暴露了 RESTful 方法的一些缺陷。

此时，就轮到 GraphQL 入场了。通过 GraphQL ，我们可以在查询中精确定义我们想要返回的数据。因此，如果我们需要上述的教程信息，我们可以创建一个查询，如下所示：

```json
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

随后，它就会返回该教程的作者信息以及指定 `id` 教程下该作者所撰写的其他教程列表。这些数据都是我们所需的，但却不用通过发送额外的 REST 请求来获得！多美好，不是吗？

## 基本设置

目前为止，我们已经了解了 GraphQL 的基础知识以及使用它的益处，下面，让我们看看如何在实战中运用它。

我们将会使用 Go 创建一个简易的 GraphQL 服务，这里，我们是使用[graphql-go/graphql](https://github.com/graphql-go/graphql) 这个库实现的。

## 设置简易的 GraphQL 服务

使用 `go mod INIt` 来初始化我们的项目：

```bash
$ Go mod INIt Github.com/elliotforbes/go-graphql-tutorial
```

接下来，让我们创建一个名为 `main.go` 的文件；并从头开始创建一个简易的 GraphQL 服务，它包含一个极简的解析器。

```go
// credit - Go-graphql hello world example
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
	rJSON, _ := JSON.Marshal(r)
	fmt.Printf("%s \n", rJSON) // { “ data ” :{ “ hello ” : ” world ” }}
}
```

现在，如果我们尝试运行它，看看会发生什么？

```bash
$ go run ./...
{"data":{"hello":"world"}}
```

所以，在一切正常的情况下，我们已经配置完成一个极简的 GraphQL 服务，并创建一个真实的请求发送至该服务。

## GraphQL 模式

我们需要分解上述例子中的代码，以便于我们之后的扩展。在 `fields...` 代码行开始处，我们定义了一个 `schema` 。当我们通过 GraphQL API 执行查询时，我们实质上定义了对象中的哪些字段是我们期望得到的。所以我们必须在 `schema` 中定义这些字段。

在 `Resolve...` 代码行处，我们定义了一个解析器函数，每当这个特定 `field` 被请求时都会触发这个解析器函数。到目前为止，我们仅仅返回了一个 `"world"` 字符串，而在此之后，我们会实现查询整个数据库的能力。

## 查询

让我们继续分解 `main.go` 文件的剩下的部分。在 `query...` 代码行开始处，我们定义了一个请求 `hello` 字段的 `query` 。

接着，我们创建了一个 `params` 结构体，该结构体包含了对之前定义的 `schema` 以及 `RequestString` 请求的引用。

最后，我们开始执行请求，并将请求的结果填充到 `r` 中。然后我们处理可能出现的错误，并将响应体解析成 JSON ，并将其打印到终端上。

## 更为复杂的例子

目前为止，我们已经有一个运行中、极简的 GraphQL 服务，我们可以通过它来执行一些查询。我们需要更进一步，构建一个更为复杂的例子。

我们会创建一个 GraphQL 服务，它返回一系列存储于内存中的教程以及相应的作者、评论等信息。

首先，我们需要定义能够表示 `Tutorial` ，`Author` 和 `Comment` 的结构体 :

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

紧接着，我们创建一个极简的 `populate` 函数用于返回元素为 `Tutorial` 类型的切片。

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

上述的代码将为我们返回一个简单的教程列表，该列表会用于我们在之后进行的解析操作。

## 创建一个新的对象类型

首先，我们使用 `graphql.NewObject()` 在 GraphQL 中创建一个新对象。我们使用 GraphQL 严格的类型来定义三种不同的类型。这些类型将与我们已经定义好的三种结构体相匹配。

`Comment` 结构体无疑是最简单的，它只包含一个字符串类型的字段 `Body` ，所以我们能够很容易地将其表示为 `commentType`：

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

接着，我们需要处理 `Author` 结构体，并将其定义为新的 `graphql.NewObject()` ，这会稍微复杂一点，因为该结构体既包含 `String` 字段，也包含一个 `Int` 值列表，这些值表示该作者所编写的教程 ID 列表。

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

最后，我们定义了 `tutorialType` ，它会封装一个 `author` ，一个元素为 `comment` 的数组，`ID` 以及 `title` 。

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

## 更新模式

到目前为止，我们已经定义了一个完整的 `Type` 系统，接下来，我们需要更新 `Schema` 以映射到这些类型上。我们会定义两个不同的 `Field` ，第一个是我们的 `tutorial` 字段，该字段允许我们根据传入的 ID 参数检索单个 `tutorial` 。第二个字段则是一个 `list` ，它允许我们检索存储于内存中的完整 `tutorials` 列表。

```go
// Schema
	fields := graphql.Fields{
		"tutorial": &graphql.Field{
			Type: tutorialType,
			// it's Good form to add a description
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

到目前为止，我们创建了 `Types` 并更新了 GraphQL 模式，看起来我们似乎做的还算不错。

## 测试它是否能够工作

让我们先尝试新的 GraphQL 服务，使用我们最新提交的查询。通过更改 `main` 函数中的 `query` 来尝试 `list` 模式。

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

我们需要分析一下，在此查询中有一个特殊的 `root` 对象。此时，我们描述了我们所期望的对象的 `list` 字段。在按照 `list` 模式返回的结果列表中，我们希望能够看到 `id` ， `title` ， `comments` 和 `author` 。

当我们运行这个查询后，我们将会看到如下输出：

```bash
$ go run ./...
{"data":{"list":[{"author":{"Name":"Elliot Forbes","Tutorials":[1]},"comments":[{"body":"First Comment"}],"id":1,"title":"Go GraphQL Tutorial"}]}}
```

正如我们所看到的，我们的查询以 JSON 的格式返回了所有教程列表，这看起来和我们定义的初始查询非常相似。

现在让我们通过 `tutorial` 模式来执行另一个查询 :

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

当我们再一次运行它，我们会看到它成功地检索到了内存中唯一一个 `ID=1` 的教程。

```bash
$ go run ./...
{"data":{"tutorial":{"author":{"Name":"Elliot Forbes","Tutorials":[1]},"title":"Go GraphQL Tutorial"}}}
```

完美，从输出结果上看，我们的 `list` 和 `tutorial` 模式能够正常工作。

> 挑战：尝试在 `populate` 函数中更新教程列表，使其可以返回更多的教程。一旦我们完成了这一步，我们就可以尝试使用查询，并加深对查询的理解。

## 总结

> 注意： 本教程全部的源代码位于这里：[main.go](https://gist.github.com/elliotforbes/9b8400ef5154eb3420e409aeffe39633)

这就是我们在本次初始教程中所介绍的所有内容。我们成功地配置了一个由内存数据存储支持的极简的 GraphQL 服务。

在下篇教程中，我们将查看 GraphQL 变更的概念，并改造我们的数据源以使用 NoSQL 数据库。关于下一篇教程可以阅读[Go GraphQL Beginners Tutorial - Part2](https://tutorialedge.net/golang/go-graphql-beginners-tutorial-part-2/)

---

via: https://tutorialedge.net/golang/go-graphql-beginners-tutorial/

作者：[Elliot Forbes](https://twitter.com/elliot_f)
译者：[barryz](https://github.com/barryz)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
