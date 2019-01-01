首发于：https://studygolang.com/articles/17363

# 使用 Go（Golang）实现 OAuth2.0

2018 年 7 月 1 日

OAuth2 是一种身份验证协议，用于使用其他服务提供商来对应用程序中的用户进行身份验证和授权。

本文将介绍如何构建 Go 应用程序  来实现 OAuth2 协议。

> 如果您想查看代码，可以在[此处](https://github.com/sohamkamani/go-oauth-example) 查看

## OAuth2 流程

在我们开始实现之前，让我们简要介绍一下 OAuth 协议。如果您曾经见过类似这样的对话框，那么您可能对 OAuth 的含义有所了解：

![gitlab 使用 GitHub OAuth](https://raw.githubusercontent.com/studygolang/gctt-images/master/implementing_OAuth_2_with_Go/oauth_example.png)

在上图中，可以看到正在尝试使用 Github 登录 Gitlab 并进行身份验证。

在任何 OAuth 流程中都有三个参与者：

1. 客户端 - 登录的人员或用户
2. 使用者 - 客户端想要登录的应用程序（在上图中是 GitLab）
3. 服务提供者 - 用户通过其进行身份验证的外部应用程序。（上图中为 GitHub）

在这篇文章中，我们将使用 Githubs OAuth2 API 进行身份验证，并使用 Web 界面构建一个在本地端口 8080 上运行的 Go 示例应用程序。所以在我们的例子中，客户端是 Web 界面，消费者是运行的应用程序 `localhost:8080`，服务提供者则是 Github。让我们来看看这一切是如何工作的：

![oauth 流程图](https://raw.githubusercontent.com/studygolang/gctt-images/master/implementing_OAuth_2_with_Go/golang_oauth.png)

我们可以在应用程序中实现流程的每个部分。

## 登陆页面

让我们创建应用程序的第一部分，即登录页面。这将是一个简单的 HTML 页面，其中包含用户应单击以使用 Github 进行身份验证的链接。以下内容将构成文件 `public/index.html`：

```html
<!DOCTYPE HTML>
<html>

<body>
  <a href="https://github.com/login/oauth/authorize?client_id=myclientid123&redirect_uri=http://localhost:8080/oauth/redirect">
    Login with GitHub
  </a>
</body>

</html>
```

以上链接有三个关键部分：

1. `https//github.com/login/oauth/authorize` 是 Github 的 OAuth 流程的 OAuth 网关。所有 OAuth 提供商都有一个网关 URL，您必须将该网址发送给用户才能继续。
2. `client_id=myclientid123` - 这指定了应用程序的客户端 ID。此 ID 将告知 Github 有关尝试使用其 OAuth 服务的消费者的身份。OAuth 服务提供商拥有门户网站，您可以在其中注册您的消费者。在注册时，您将收到一个客户端 ID（我们在此处使用 myclientid123）和客户端密码（稍后我们将使用）。对于 Github，可以在 `https://github.com/settings/applications/new` 上进行新消费者应用的注册。
3. `redirect_uri=http://localhost:8080/oauth/redirect` - 一旦用户通过服务提供商进行身份验证，指定要获取请求令牌重定向的 URL。通常，您还必须在注册门户上设置此值，以防止任何人设置恶意回调 URL。

接下来，我们需要以服务的方式提供上面制作的文件 `index.html`。以下代码将构成一个新文件 `main.go`：

```go
func main() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	http.ListenAndServe(":8080", nil)
}
```

在当前状态下，您可以启动服务器（通过执行 `go run main.go`）并访问 `http：// localhost：8080`，您将看到我们刚刚创建的登录页面。单击“使用 GitHub 登录”链接后，您将被重定向到熟悉的 OAuth 页面以向 Github 注册。但是，一旦您进行身份验证，您将被重定向到 `http://localhost:8080/oauth/redirect`，此时，它不会执行任何操作，并将导致服务器上的 404 页面。

## 重定向路由

一旦用户使用 Github 进行身份验证，他们就会被重定向到之前指定的重定向 URL。服务提供商还会添加一个请求令牌到 URL。在当前例子中，Github 增加了为 `code` 参数，所以重定向 URL 实际上是这样的 `http://localhost:8080/oauth/redirect?code=mycode123`，在这里 `mycode123` 为请求令牌的值。我们需要这个请求令牌，以及我们的客户机密钥来获取访问令牌 (access token)，这是实际用于获取用户信息的令牌。我们通过对 `https://github.com/login/oauth/access_token` 进行 `POST` 请求调用来获取此访问令牌。

> 有关 Github 提供给重定向 URL 的信息的完整文档，以及我们提供的 POST `/login/oauth/access_token` HTTP 调用所需的信息，可以在[此处](https://developer.github.com/apps/building-oauth-apps/authorizing-oauth-apps/#2-users-are-redirected-back-to-your-site-by-github) 找到。

将下面的内容添加到 main.go 文件中，以处理 /oauth/redirect 路由：

```go
const clientID = "<your client id>"
const clientSecret = "<your client secret>"

func main() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	// We will be using `httpClient` to make external HTTP requests later in our code
	httpClient := http.Client{}

	// Create a new redirect route route
	http.HandleFunc("/oauth/redirect", func(w http.ResponseWriter, r *http.Request) {
		// First, we need to get the value of the `code` query param
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not parse query: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		code := r.FormValue("code")

		// Next, lets for the HTTP request to call the GitHub OAuth enpoint
		// to get our access token
		reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientID, clientSecret, code)
		req, err := http.NewRequest(http.MethodPost, reqURL, nil)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		// We set this header since we want the response
		// as JSON
		req.Header.Set("accept", "application/json")

		// Send out the HTTP request
		res, err := httpClient.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		defer res.Body.Close()

		// Parse the request body into the `OAuthAccessResponse` struct
		var t OAuthAccessResponse
		if err := JSON.NewDecoder(res.Body).Decode(&t); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}

		// Finally, send a response to redirect the user to the "welcome" page
		// with the access token
		w.Header().Set("Location", "/welcome.html?access_token="+t.AccessToken)
		w.WriteHeader(http.StatusFound)
	})

	http.ListenAndServe(":8080", nil)
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}
```

现在，重定向 URL（如果可用）将将用户重定向到欢迎页面并获取访问令牌 (access token)。

## 欢迎页面

欢迎页面是我们在用户登录后向用户显示的页面。现在我们拥有用户访问令牌，我们可以代表他们获得授权的 Github 用户的帐户信息。

> 有关所有可用 API 的列表，您可以查看[Github API 文档](https://developer.github.com/v3/)

我们将使用 `/user` API 获取有关用户的基本信息，并在欢迎页面上与他们打个招呼。创建一个新文件 public/welcome.html：

```html
<!DOCTYPE HTML>
<html lang="en">

<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, INItial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>Hello</title>
</head>

<body>

</body>
<script>
	// We can get the token from the "access_token" query
	// param, available in the browsers "location" global
	const query = Windows.location.search.substring(1)
	const token = query.split('access_token=')[1]

	// Call the user info API using the fetch browser library
	fetch('//api.github.com/user', {
			headers: {
				// Include the token in the Authorization header
				Authorization: 'token ' + token
			}
		})
		// Parse the response as JSON
		.then(res => res.json())
		.then(res => {
			// Once we get the response (which has many fields)
			// Documented here: https://developer.github.com/v3/users/#get-the-authenticated-user
			// Write "Welcome <user name>" to the documents body
			const nameNode = document.createTextNode(`Welcome, ${res.name}`)
			document.body.appendChild(nameNode)
		})
</script>

</html>
```

通过添加欢迎页面，我们的 OAuth 流程现已完成！应用程序启动后，可以转到 `http：//localhost:8080/`，使用 Github 进行授权，最后在欢迎页面上显示问候语。我在 GitHub 配置文件中的名字是“ Soham Kamani ”，因此登录后 "Welcome, Soham Kamani" 将显示 weclome 页面。

> 该源代码，以及如何运行的说明，可以在[这里](https://github.com/sohamkamani/go-oauth-example) 找到

## 让应用更安全

虽然这篇文章展示了 OAuth2 的基础知识，但还有很多其他方法可以使您的应用更加安全，这里没有涉及：

1. 在此示例中，我们将访问令牌 (access token) 传递给客户端，以便它可以作为授权用户发出请求。为了使您的应用更安全，访问令牌不应直接传递给用户。而是创建一个会话令牌，作为 cookie 发送给用户。

> 该应用程序在服务端将维护会话令牌及访问令牌的映射。用户不会向 GitHub 发出请求，而是向服务器发出请求（使用会话令牌），然后使用提供的会话令牌查找访问令牌并利用访问令牌在服务器端向 GitHub 发出请求。我在[这里](https://www.sohamkamani.com/blog/2017/01/08/web-security-session-cookies/) 写了更多关于会话和 cookie 的文章。

2. 在将用户发送到授权 URL 时，可以在 URL 携带一个自定义的查询参数 `state`。这个参数值应该是应用程序提供的随机不可猜测的字符串。当 GitHub 调用重定向 url 时，它会将此 `state` 变量附加到请求参数。新网址现在看起来像： `https://github.com/login/oauth/authorize?client_id=myclientid123&redirect_uri=http://localhost:8080/oauth/redirect&state=somerandomstring`

> 应用程序现在可以将此值与其最初生成的值进行比较。如果它们不相同，则意味着请求来自某个第三方，并且应该被拒绝。有关此类型的安全问题的更多信息，您可以阅读我的[其他帖子](https://www.sohamkamani.com/blog/2017/01/14/web-security-cross-site-request-forgery/)

---

via: https://www.sohamkamani.com/blog/golang/2018-06-24-oauth-with-golang/

作者：[Soham Kamani](https://github.com/sohamkamani)
译者：[lovechuck](https://github.com/lovechuck)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
