首发于：https://studygolang.com/articles/20747

# 60 秒搭建无服务 Golang 应用

在这篇简短的文章中，我将说明使用 [Apex Up](https://github.com/apex/up) 对任何 Golang 应用程序或 API 是如何快速的生产出无服务环境的。这篇文章假设你在你的机器上配置了 [AWS credentials](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)，因为 Up 是部署到 AWS Lambda 和 API 网关上的。

Up 适用于任何 Go 应用程序，没有必要专门为了构建无服务而去重写你的应用程序，但是如果你想跟随写，你可以将以下内容复制粘贴到你项目中的 `main.go` 文件中：

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

var pets = []string{
	"Tobi",
	"Loki",
	"Jane",
}

func main() {
	port := os.Getenv("PORT")
	http.HandleFunc("/pets", getPets)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getPets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pets)
}
```

你可能已经注意到在上一段代码中使用了 `PORT` 环境变量，这是一个由 Up 传递给你应用程序的一个端口号，这样它就知道去哪里监听请求了，除此之外它是一个完全正常的 Golang net/http 应用程序。

让我们来看看 Up 实际上是怎么工作的！

## Proxy 如何工作

当你看到术语 "serverless" 的时候，你可能想到 FaaS - 或是功能即为服务 - 但是这不是你将通过 Up 找到的，它将你的整个应用程序部署到单个 AWS Lambda 函数中，让你专注于构建你的 API 或者 应用程序，因为你可以通过 `go run main.go` 像你在本地开发一样简单。

当你的应用程序接收请求时，将由 [API Gateway](https://aws.amazon.com/cn/api-gateway/)（一个由 AWS 提供的无服务负载均衡器）处理。请求传递给你的 Lambda 函数，该函数通常类似于以下内容（在 Node.js 中），你将使用 HTTP 的 `event` 作为请求，返回对象作为响应进行交互。这在简单的场景中可以很好，但是它将你锁定在 FaaS 中，并且在本地开发可能更加困难。

```javascript
exports.handle = async function(event, context) {
  return {
    statusCode: 200,
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(event)
  }
}
```

为了允许 "vanilla" HTTP 应用程序能够工作，Up 处理 API Gateway 事件，将其转换为一个真正的 HTTP 请求。Up 发送请求到你的应用程序进程中，然后将响应转换回 API Gateway 期望的返回值。

由于 Up 充当你应用程序的反向代理，因此它可以提供中间件功能，例如 CORS，响应压缩，日志，错误页面等等。

## 部署你的 API

为了开始部署应用程序，首先安装 `up`：

```bash
$ curl -sf https://up.apex.sh/install | sh
```

在你的项目目录中运行 `up`，将提示你创建一个 up.json 文件：

![1_7Ztkz1EFWCoR5wf5R5E16w](https://raw.githubusercontent.com/studygolang/gctt-images/master/serverless-golang-apps-in-60-seconds/1_7Ztkz1EFWCoR5wf5R5E16w.png)

Up 将指导你选择应用程序的名称，要使用的 AWS 配置文件凭据和 AWS 区域。

![1_-HzvtzB3PZO318YTXDna2g](https://raw.githubusercontent.com/studygolang/gctt-images/master/serverless-golang-apps-in-60-seconds/1_-HzvtzB3PZO318YTXDna2g.png)

大约 60 秒后，你会有一个完全可以投入生产的无服务环境能够运行！当初始部署完成后，后续部署将会更快。

![1_SBH63IX2RHToaK8Qhml_Zw](https://raw.githubusercontent.com/studygolang/gctt-images/master/serverless-golang-apps-in-60-seconds/1_SBH63IX2RHToaK8Qhml_Zw.png)

列出的端点是 `staging` 环境，默认情况下还有一个生产环境，但是你也可以定义 [个性化阶段](https://up.docs.apex.sh/#configuration.stages)。

使用 `curl` 尝试你的新 API：

```bash
$ curl https://jb8mxj0cda.execute-api.eu-west-2.amazonaws.com/staging/pets
["Tobi","Loki","Jane"]
```

或者，你可以使用 `up url` 扩展到同一端点：

```bash
$ curl up url/pets
["Tobi","Loki","Jane"]
```

就这样了！如果你有兴趣了解更多 Up 提供的信息，请查看以下功能或 [文档](https://up.docs.apex.sh/)。完成后，运行以下命令去删除应用程序和资源：

```bash
$ up stack delete
```

## OSS 功能

- HTTP 中间件，如错误页面，CORS，重定向，静态文件服务等
- 基础架构即代码 - 使用行业最佳实践去管理你的配置
- 无限的可扩展性 - 无需担心扩展的机器，Up 是按需的
- 低成本 - 只为你使用的付费，每月有 1,000,000 次 AWS 免费请求
- 基于阶段部署，支持自定义阶段
- 隔离 - 不用担心多租户 PaaS 的停机时间，部署到你自己的 AWS
- 经济高效的结构化日志记录仅需 $0.5/gb
- 请求隔离 - 每个请求与它自己的 Lambda 隔离，crash 永远不会让你的整个应用程序崩溃
- 富有表现力的结构化日志查询（例如：`up logs 'error message = "login failed" region="us-west-2"'`）
- 免费的 SSL 和自定义域名支持
- [Slack chat](http://chat.apex.sh/) 社区支持

## 专业功能

- 单月固定费用为 $10 每月，团队成员无需额外费用
- 使用 Git 继承部署日志历史记录
- 即时回滚到以前的部署或者 Git tag/commit
- 加密环境变量，集中并且定义每一个阶段
- 通过 SMA，Slack 和邮件 [提醒](https://up.docs.apex.sh/#configuration.alerting) 错误和性能问题
- 支持 [区域端点](https://aws.amazon.com/about-aws/whats-new/2017/11/amazon-api-gateway-supports-regional-api-endpoints/)
- 支持 [Lambda 层](https://docs.aws.amazon.com/lambda/latest/dg/configuration-layers.html)
- 优先邮件支持
- 保持项目活力

---

## 链接

- 文档：[https://up.docs.apex.sh](https://up.docs.apex.sh/)
- GitHub 仓库：[https://github.com/apex/up](https://github.com/apex/up)

---

via: https://medium.com/@tjholowaychuk/serverless-golang-apis-in-60-seconds-46e4ac36b680

作者：[TJ Holowaychuk](https://medium.com/@tjholowaychuk)
译者：[PotoYang](https://github.com/PotoYang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
