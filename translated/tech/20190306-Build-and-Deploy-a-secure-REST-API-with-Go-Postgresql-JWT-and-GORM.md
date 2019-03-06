# 使用 Go、Postgresql、JWT 和 GORM 搭建安全的 REST API

在本教程中，我们将要学习如何使用 Go 语言开发和部署一个安全的 REST API

## 为什么用 Go

Go 是一个非常有意思的编程语言，它是一个强类型的语言，并且它编译得非常的快，它的性能可以与 C++ 相比较，Go 还有 goroutine —— 一个更加高效版本的线程——我知道这些特性都不是什么新鲜的东西了，但我就是喜欢 Go 这个样子的。

## 我们将要做个什么东西？

我们将要写一个通讯录的 APP，我们的 API 允许用户添加以及查看他们的联系人，这样当他们手机丢失了也不会担心找不到联系人的信息了。

## 准备工作

在本教程中我假定你已经安装了以下的工具：

- Go
- Postgresql
- Goland IDE —— 可选（我将在本教程中使用它）

我还假定你已经设置好了你的 GOPATH，如果你还没有设置好，可以参考[这篇文章](https://github.com/golang/go/wiki/SettingGOPATH)。

让我们现在就开始吧！

## 什么是 REST?

REST 是 *Representational State Transfer* （表述性状态转移）的缩写，它是现代客户端程序通过 http 与数据库或服务器通讯时广泛采用的一种机制（https://en.wikipedia.org/wiki/Representational_state_transfer）。所以，如果你有个新的创业想法或者有个很酷的小项目想要做。REST 大概就是你要用到的协议。

## 创建 APP

我们先来确定要用到哪些第三方的代码包，幸运的是，Go 标准库已经足够我们搭建一个完整网站了（我希望我没说错）——你可以看看 `net.http` 代码包，但是为了让我的工作简单点，我们将需要以下几个代码包：

- gorilla/mux —— 一个强力的 URL 路由器（router）和分发器（dispatcher）。我们使用这个代码包来让 URL 匹配上它们的 handler。
- jinzhu/gorm —— 一个非常棒的 Go 语言的 ORM 库，它的宗旨是成为一个对开发者友好的库。我们使用这个 ORM（Object relational mapper，对象关系映射）代码包来使我们程序与数据库的交互更加简单。
- dgrijalva/jwt-go —— 使用它来签发和验证 JWT token。
- joho/godotenv —— 用来加载 .env 文件到项目中。

要安装这些代码包，只要打开你的终端控制台并运行：

```bash
go get github.com/{包名称}
```

这个命令会安装代码包到你的 `GOPATH`。

## 项目结构

![img](assets/1_3MJjEDEI7i29eJecxxfopA.png)

*可以在右边的面板看到项目结构*

utils.go

```go
package utils

import (
	"encoding/json"
	"net/http"
)

func Message(status bool, message string) (map[string]interface{}) {
	return map[string]interface{} {"status" : status, "message" : message}
}

func Respond(w http.ResponseWriter, data map[string] interface{})  {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
```

`utils.go` 包含了一些好用的工具函数来构建 `json` 消息并返回一个 `json` 响应包。在我们继续之前，请先关注一下 `Message()` 和 `Respond()` 这两个函数的实现。

## JWT 详解

JWT（JSON Web Tokens ）是一个开放的工业标准（[RFC 7519](https://tools.ietf.org/html/rfc7519)）方法，用于在双方之间安全的交互。通过 session，我们可以很轻易的验证一个 web 程序的用户，但是，当你的 web 程序 API 需要和安卓或 IOS 客户端交互时，session 就不能用了，因为 http 的请求有无状态的特性。使用 JWT，我们可以为每一个用户创建一个特殊的 token，这个将会被包含在后续的 API 请求头里面。这个方法能让我们验证每一个人调用我们 API 的用户身份。我们来看看下面的实现

```go
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"go-contacts/models"
	u "go-contacts/utils"
)

var JwtAuthentication = func(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		notAuth := []string{"/api/user/new", "/api/user/login"} // 不需要验证用户的 URL 列表
		requestPath := r.URL.Path                               // 当前请求的 path

		// 检查请求是否需要鉴权，如果不需要，直接接受请求
		for _, value := range notAuth {

			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		response := make(map[string]interface{})
		tokenHeader := r.Header.Get("Authorization") // 从请求头中获取 token

		if tokenHeader == "" { // 没有找到 Token，返回 403 未授权错误
			response = u.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			u.Respond(w, response)
			return
		}

		splitted := strings.Split(tokenHeader, " ") // 通常 token 的格式为 `Bearer {token-body}`, 我们看看获取到的 token 格式是否正确
		if len(splitted) != 2 {
			response = u.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			u.Respond(w, response)
			return
		}

		tokenPart := splitted[1] // 提取我们需要的 token 部分
		tk := &models.Token{}

		token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("token_password")), nil
		})

		if err != nil { // token 格式不正确，同样返回 403 错误
			response = u.Message(false, "Malformed authentication token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			u.Respond(w, response)
			return
		}

		if !token.Valid { // token 无效，可能不是本服务器签发的 token
			response = u.Message(false, "Token is not valid.")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			u.Respond(w, response)
			return
		}

		// 一切正常，继续处理请求，并且把调用者设置为 token 对应的用户
		fmt.Sprintf("User %", tk.Username) //Useful for monitoring
		ctx := context.WithValue(r.Context(), "user", tk.UserId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r) // 继续执行中间件调用链
	})
}
```

代码中的注释解释了所有需要了解的信息，基本上来说，这段代码创建了一个 `Middleware` （中间件）并拦截所有的请求，检查 token 是否存在（JWT token），验证它是否有效，然后如果有任何不对的地方，立刻返回错误给请求方。如果没有错误，则继续处理请求，待会你将会看到，我们如何处理服务器与请求方的交互。

## 编写用户注册与登录系统

在用户能够在保存它们的联系人信息在我们服务器里之前，我们希望他们可以在我们的系统中注册和登录。要做的第一件事就是连接到我们的数据库，我们使用 `.env` 文件来保存我们的数据库登录信息，我的 `.env` 文件是这样的：

```
db_name = gocontacts
db_pass = **** // windows系统默认情况下，这个是当前用户的 windows 登录密码
db_user = postgres
db_type = postgres
db_host = localhost
db_port = 5434
token_password = thisIsTheJwtSecretPassword // 不要上传到代码仓库！
```

然后，我们现在可以使用以下代码来连接数据库



---
via: https://medium.com/@adigunhammedolalekan/build-and-deploy-a-secure-rest-api-with-go-postgresql-jwt-and-gorm-6fadf3da505b

作者：[Adigun Hammed Olalekan](https://medium.com/@adigunhammedolalekan)
译者：[译者ID](https://github.com/译者ID)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出