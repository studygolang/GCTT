已发布：https://studygolang.com/articles/12354

# 使用 SwaggerUI 创建 Golang API 文档

为你的 API 提供一个文档比你想象中更加有用，即使你没有公开你的 API ，为你的前端或者移动团队提供一个文档会比你提供截图/片段或使用 Postman/Insomnia （带有同步的高级版本）等付费产品更容易。借助 SwaggerUI ，您可以自动获得所有 API 的设计良好的文档。当切换到 Go 时，由于缺少文档/教程，我在配置它的时候出现了一些问题，所以我决定写一个。

![goswagger](https://raw.githubusercontent.com/studygolang/gctt-images/master/swagger-golang/swagger-golang1.jpg)

示例程序:[链接](https://github.com/ribice/golang-swaggerui-example)

大约两年前，我曾经在开发一个 RESTful 风格的企业应用的后台时，第一次知道 [SwaggerUI](https://swagger.io/swagger-ui/) 。 SwaggerUI 的创造者 SmartBear 将其产品描述为：

> "Swagger UI 允许任何人（无论是你的开发团队还是最终用户）在没有任何实现逻辑的情况下对 API 资源进行可视化和交互。它（API文档）通过 Swagger 定义自动生成，可视化文档使得后端实现和客户端消费变得更加容易。"

简而言之，通过提供 Swagger（OpenAPI）定义，您可以获得与 API 进行交互的界面，而不必关心编程语言本身。你可以将 Swagger（OpenAPI） 视为 REST 的 WSDL 。

作为参考，Swagger Codegen 可以从这个定义中，用几十种编程语言来生成客户端和服务器代码。

回到那个时候，我使用的是 Java 和 SpringBoot ，觉得 Swagger 简单易用。你仅需创建一次 bean ，并添加一两个注解到端点上，再添加一个标题和一个项目描述。此外，我习惯将所有请求从 “/” 重定向到 “/swagger-ui” 以便在我打开 `host:port` 时自动跳转到 SwaggerUI 。在运行应用程序的时候， SwaggerUI 在同一个端口依然可用。（例如，您的应用程序运行在`[host]:[port]`， SwaggerUI 将在`[host]:[port]/swagger-ui`上访问到）。

快一年半之后，我想在我们的 Go 项目中实现 SwaggerUI 。问题是 —— 感觉太复杂了。当我在网络上搜索时，我看到不仅仅是我，其他许多用户也遇到了同样的麻烦。

我偶然发现了 Go-Swagger 项目。他们的 [GitHub Readme](https://github.com/go-swagger/go-swagger) 谈论的是 Swagger ，而不是 SwaggerUI 。你得有一个客户端，服务器，中间件和一些其他的东西，才可以考虑 SwaggerUI 。简而言之， Swagger 服务器/客户端用于从 Swagger 定义（swagger.json）中生成（后端）代码。生成服务器使您可以从规范中提供 API ，同时为这些 API 使用者生成客户端。把它看作代码生成工具。我觉得这很有用，但那不是我想要的。

在社区的帮助下（向 [casualjim](https://github.com/casualjim) 致意）和一些调查，我成功地为我们的项目生成了没有太多的样板代码的 API 文档。

另外，我准备了一个实现了 go-swagger 注释来生成有效的 swagger 文档示例，可以[在这里](https://github.com/ribice/golang-swaggerui-example)找到。

## 安装 Go-Swagger

在开始之前，您需要在本地机器上安装 go swagger 。这不是一个强制性的步骤，但会使得更容易的使用 swagger 工作。安装它可以让你在本地测试你的注释，否则，你只能依靠你的 CI 工具。

最简单的安装方式是通过运行 Homebrew / Linuxbrew ：

```console
brew tap go-swagger/go-swagger
brew install go-swagger
```

此外，你可以从[这里](https://github.com/ribice/golang-swaggerui-example)得到最新的二进制文件。

## Swagger:meta [[docs]](https://goswagger.io/generate/spec/meta.html)

![swagger-meta](https://raw.githubusercontent.com/studygolang/gctt-images/master/swagger-golang/swagger-golang2.jpg)

这是你应该添加到项目中的第一个注释。它被用来描述你的项目名称，描述，联系电子邮件，网站，许可证等等。

如果你的 API 仅提供在 HTTP 或 HTTPS 上，且只生成 JSON ，您应在此处添加它 - 允许你从每个路由中删除该注释。

安全也被添加在 swagger:meta 中，在 SwaggerUI 上添加一个授权按钮。为了实现JWT，我使用安全类型承载进行命名并将其定义为：

```go
//     Security:
//     - bearer
//
//     SecurityDefinitions:
//     bearer:
//          type: apiKey
//          name: Authorization
//          in: header
//
```

## Swagger:route [[docs]](https://goswagger.io/generate/spec/route.html)

有两种方式两个注释你的路由，swagger:operation 和swagger:route 。两者看起来都很相似，那么主要区别是什么？

把 swagger:route 看作简单 API 的短注释,它适用于没有输入参数（路径/查询参数）的 API 。那些（带有参数）的例子是 /repos/{owner} ， /user/{id} 或者 /users/search?name=ribice

如果你有一个那种类型，那么你就必须使用 swagger:operation ，除此之外，如 /user 或 /version 之类的 APIs 都可以用 swagger:route 来注释。

**swagger:route**注释包含以下内容：

```go
// swagger:route POST /repo repos users createRepoReq
// Creates a new repository for the currently authenticated user.
// If repository name is "exists", error conflict (409) will be returned.
// responses:
//  200: repoResp
//  400: badReq
//  409: conflict
//  500: internal
```

1. **swagger:route** - 注解
2. **POST** - HTTP方法
3. /**repo** - 匹配路径，端点
4. **repos** - 路由所在的空间分割标签，例如，“repos users”
5. **createRepoReq** - 用于此端点的请求（详细的稍后会解释）
6. **Creates a new repository …** - 摘要（标题）。对于swager:route注释，在第一个句号（.）前面的是标题。如果没有句号，就会没有标题并且这些文字会被用于描述。
7. **If repository name exists …** - 描述。对于swager:route类型注释，在第一个句号（.）后面的是描述。
8. **responses**: - 这个端点的响应
9. **200: repoResp** -  一个（成功的）响应HTTP状态 200，包含 repoResp（用 swagger:response 注释的模型）
10. **400: badReq, 409: conflict, 500: internal** - 此端点的错误响应（错误请求，冲突和内部错误， 定义在 cmd/api/swagger/model.go 下）

如此注释您的端点将产生以下内容：

![swagger-route-ui](https://raw.githubusercontent.com/studygolang/gctt-images/master/swagger-golang/swagger-golang3.jpg)

请记住，您还可能需要使用其他注释，具体取决于您的 API 。由于我将我的项目定义为仅使用单一模式（ https ），并且我的所有 API 都使用 https ，所以我不需要单独注释方案。如果您为端点使用多个模式，则需要以下注释：

```go
// Schemes: http, https, ws, wss
```

同样适用于 消费者/生产者 媒体类型。我所有的 API 都只消费/生成 application/json 。如果您的 API 正在 消费/生成 其他类型，则需要使用该媒体类型对其进行注释。例如：

```go
// consumes:
// - application/json
// - application/x-protobuf
//
// produces:
// - application/json
// - application/x-protobuf
```

安全性：

```go
// security:
//   api_key:
//   oauth: read, write
//   basicAuth:
//      type: basic
//   token:
//      type: apiKey
//      name: token
//      in: query
//   accessToken:
//      type: apiKey
//      name: access_token
//      in: query
```

另一方面，swagger:operation 用于更复杂的端点。三个破折号（-）下的部分被解析为 YAML ，允许更复杂的注释。确保您的缩进是一致的和正确的，否则将无法正确解析。

## Swagger:operation [docs](https://goswagger.io/generate/spec/operation.html)

使用 Swagger:operation 可以让你使用所有[OpenAPI规范](https://swagger.io/specification/)，你可以描述你的复杂的端点。如果你对细节感兴趣，你可以阅读规范文档。

简单来说 - swagger:operation 包含如下内容：

```go
// swagger:operation GET /repo/{author} repos repoList
// ---
// summary: List the repositories owned by the given author.
// description: If author length is between 6 and 8, Error Not Found (404) will be returned.
// parameters:
// - name: author
//   in: path
//   description: username of author
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/reposResp"
//   "404":
//     "$ref": "#/responses/notFound"
```

1. **swagger:operation** - 注释
2. **GET** - HTTP 方法
3. /**repo/{author}** - 匹配路径，端点
4. **repos** - 路由所在的空间分割标签，例如，“repos users”
5. **repoList** - 用于此端点的请求。这个不存在（没有定义），但参数是强制性的，所以你可以用任何东西来替换repoList（noReq，emptyReq等）
6. --- - 这个部分下面是YAML格式的swagger规范。确保您的缩进是一致的和正确的，否则将无法正确解析。注意，如果你在YAML中定义了标签，摘要，描述或操作标签，将覆盖上述常规swagger语法中的摘要，描述，标记或操作标签。
7. **summary**: - 标题
8. **description**: - 描述
9. **parameters**: - URL参数（在这个例子中是{author}）。字符串格式，强制性的（Swagger不会让你调用端点而不输入），位于路径（/{author}）中。另一种选择是参数内嵌的请求 (?name="")

定义你的路由后，你需要定义你的请求和响应。从示例中，你可以看到，我创建了一个新的包，命名为 swagger 。这不是强制性的，它把所有样板代码放在一个名为 swagger 的包中。但缺点是你必须导出你的所有 HTTP 请求和响应。

如果你创建了一个单独的 Swagger 包，确保将它导入到你的主/服务器文件中（你可以通过在导入前加一个下划线来实现）：

```go
_ "github.com/ribice/golang-swaggerui-example/cmd/swagger"
```

## Swagger:parameters [[docs]](https://goswagger.io/generate/spec/params.html)

根据您的应用程序模型，您的 HTTP 请求可能会有所不同（简单，复杂，封装等）。要生成 Swagger 规范，您需要为每个不同的请求创建一个结构，甚至包含仅包含数字（例如id）或字符串（名称）的简单请求。

一旦你有这样的结构（例如一个包含一个字符串和一个布尔值的结构），在你的Swagger包中定义如下：

```go
// Request containing string
// swagger:parameters createRepoReq
type swaggerCreateRepoReq struct {
	// in:body
	api.CreateRepoReq
}
```

- 第 1 行包含一个在 SwaggerUI 上可见的注释
- 第 2 行包含 swagger:parameters 注释，以及请求的名称（operationID）。此名称用作路由注释的最后一个参数，以定义请求。
- 第 4 行包含这个参数的位置（in:body，in:query 等）
- 第 5 行是实际的内嵌结构。正如前面所提到的，你不需要一个独立的 swagger 批注包（你可以把swagger:parameters注释放在 api.CreateRepoReq 上），但是一旦你开始创建响应注释和验证，那么在 swagger 相关批注一个单独的包会更清晰。

![swagger-parameters](https://raw.githubusercontent.com/studygolang/gctt-images/master/swagger-golang/swagger-golang4.jpg)

如果你有大的请求，比如创建或更新，你应该创建一个新类型的变量,而不是内嵌结构。例如（注意第五行的区别）:

```go
// Request containing string
// swagger:parameters createRepoReq
type swaggerCreateRepoReq struct {
	// in:body
	Body api.CreateRepoReq
}
```

这会产生以下 SwaggerUI 请求：

![swagger-patameters-ui](https://raw.githubusercontent.com/studygolang/gctt-images/master/swagger-golang/swagger-golang5.jpg)

 Swagger 有很多验证注释提供给 swagger:parameters和 swagger:response ，在注释标题旁边的文档中有详细的描述和使用方法。

## Swagger:response [[docs]](https://goswagger.io/generate/spec/response.html)

响应注释与参数注释非常相似。主要的区别在于，经常将响应包裹到更复杂的结构中，所以你必须要在 swagger 中考虑到这点。

在我的示例中，我的成功响应如下所示：

```json
{
	 "code":200, // Code containing HTTP status CODE
	 "data":{} // Data containing actual response data
}
```

虽然错误响应有点不同：

```json
{
	 "code":400, // Code containing HTTP status CODE
	 "message":"" // String containing error message
}
```

要使用常规响应，像上面错误响应那样的，我通常在 swagger 包内部创建 model.go（或swagger.go）并在里面定义它们。在示例中，下面的响应用于 OK 响应（不返回任何数据）：

```go
// Success response
// swagger:response ok
type swaggScsResp struct {
	// in:body
	Body struct {
		// HTTP status code 200 - OK
		Code int `json:"code"`
	}
}
```

对于错误响应，除了名称（和示例的情况下的 HTTP 代码注释）之外，它们中的大多数类似于彼此。尽管如此，你仍然应该为每一个错误的情况进行定义，以便把它们作为你的端点可能的响应：

```go
// Error Forbidden
// swagger:response forbidden
type swaggErrForbidden struct {
	// in:body
	Body struct {
		// HTTP status code 403 -  Forbidden
		Code int `json:"code"`
		// Detailed error message
		Message string `json:"message"`
	}
}
```

 data 中包含 model.Repository 的示例响应：

```go
// HTTP status code 200 and repository model in data
// swagger:response repoResp
type swaggRepoResp struct {
	// in:body
	Body struct {
		// HTTP status code 200/201
		Code int `json:"code"`
		// Repository model
		Data model.Repository `json:"data"`
	}
}
```

 data 中包含 model.Repository 切片的示例响应：

```go
// HTTP status code 200 and an array of repository models in data
// swagger:response reposResp
type swaggReposResp struct {
	// in:body
	Body struct {
		// HTTP status code 200 - Status OK
		Code int `json:"code"`
		// Array of repository models
		Data []model.Repository `json:"data"`
	}
}
```

总之，这将足以生成您的 API 文档。您也应该向文档添加验证，但遵循本指南将帮助您开始。由于这主要是由我自己的经验组成，并且在某种程度上参考了 Gitea 的[源代码](https://github.com/go-gitea/gitea)，我将会听取关于如何改进这部分并相应更新的反馈。

如果您有一些问题或疑问，我建议您查看[如何生成FAQ](https://goswagger.io/faq/faq_spec.html)。

## 本地运行 SwaggerUI

一旦你的注释准备就绪，你很可能会在你的本地环境中测试它。要做到这一点，你需要运行两个命令：

1. Generate spec [[docs]](https://goswagger.io/generate/spec.html)
2. Serve [[docs]](https://goswagger.io/usage/serve_ui.html)

这个命令我们用来生成 swagger.json 并使用 SwaggerUI：

```console
swagger generate spec -o ./swagger.json --scan-models
swagger serve -F=swagger swagger.json
```

或者，如果你只想使它成为一个命令：

```console
swagger generate spec -o ./swagger.json --scan-models && swagger serve -F=swagger swagger.json
```

执行该命令后，将使用 [Petstore](http://petstore.swagger.io/) 托管的 SwaggerUI 打开一个新选项卡。服务器启用了 CORS，并将标准 JSON 的 URL 作为请求字符串附加到 petstore URL。

另外，如果使用 Redoc flavor（-F = redoc），则文档将托管在您自己的计算机上（localhost:port/docs）。

## 在服务器上部署

在服务器上部署生成的 SwaggerUI 有很多种方法。一旦你生成了 swagger.json，它应该相对容易地被运行。

例如，我们的应用程序正在 Google App Engine 上运行。Swagger Spec 由我们的 CI 工具生成，并在 /docs 路径上提供。

我们将 SwaggerUI 作为 Docker 服务部署在 GKE（Google Container/Kubernates Engine）上，它从 /docs 路径中获取swagger.json。

我们的 CI（Wercker）脚本的一部分：

```yaml
build:
	steps:
		- script:
			name: workspace setup
			code: |
				mkdir -p $GOPATH/src/github.com/orga/repo
				cp -R * $GOPATH/src/github.com/orga/repo/
		- script:
			cwd: $GOPATH/src/bitbucket.org/orga/repo/cmd/api/
			name: build
			code: |
				go get -u github.com/go-swagger/go-swagger/cmd/swagger
				swagger generate spec -o ./swagger.json --scan-models
				CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o app .
				cp app *.template Dockerfile swagger.json "$WERCKER_OUTPUT_DIR"
```

路由：

```go
func (d *Doc) docHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
	r.Header.Add("Content-Type", "application/json")
	data, _ := ioutil.ReadFile("/swagger.json")
	w.Write(data)
}
```

Dockerfile：

```Dockerfile
FROM swaggerapi/swagger-ui
ENV API_URL "https://api.orga.com/swagger"
```

## 总结

SwaggerUI 是一个功能强大的 API 文档工具，可以让您轻松而漂亮地记录您的 API。在 go-swagger 项目的帮助下，您可以轻松地生成 SwaggerUI 所需的swagger规范文件（swagger.json）。

总之，我描述了为实现这一目标所采取的步骤。可能有更好的方法，我会确保根据收到的反馈更新这篇文章。

示例在 [GitHub](https://github.com/ribice/golang-swaggerui-example)上可用。从示例生成的 Swagger.json 在[LINK](https://ribice.ba/goswagg/v1/swagger)。

---

via: https://www.ribice.ba/swagger-golang/

作者：[Emir Ribic](https://github.com/ribice)
译者：[fatalc](https://github.com/fatalc)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
