首发于：https://studygolang.com/articles/13876

# 使用 JWT 保护 API 访问

APIs 的一个常见用例是提供一个授权中间件，允许客户端向 APIs 发送授权请求。通常来说，客户端会执行一些授权逻辑，产生一个「会话标识」。最近比较流行的 JWT ( JSON Web Tokens ) 提供了一个带超时时间的「会话标识」，使用它不需要额外的空间来执行验证逻辑。

本文是接着上一篇文章写的，在阅读下面内容之前建议先看一下之前的那篇文章 [用 go-chi 处理 HTTP 请求](https://scene-si.org/2018/03/12/handling-http-requests-with-go-chi/)

接下来我们要用 [go-chi/jwtauth](https://github.com/go-chi/jwtauth) 在 APIs 上增加一个授权层。它是基于 [go-chi/chi](https://github.com/go-chi/chi) 实现的。授权可以是任意的（针对登录和没有登录的用户）或者有针对性的（只针对已经登录的用户）。这样就可以对两种用户实现不同的授权逻辑，根据 JWT 参数的合法性返回额外的授权验证信息。我用了 [titpetric/factory/resputil](https://github.com/titpetric/factory/tree/master/resputil) 来简化错误处理和 JSON 数据的格式化。

## JWT 到底是什么

> JSON Web Token ( JWT ) 是一个开放的标准 ( [RFC 7513](https://tools.ietf.org/html/rfc7519) )，定义如何在各部分之间安全的传输 JSON 对象，标准简洁而且自包含，另外还对其加了数字签名，所以可以对其合法性进行验证

「JWT」由三部分构成

1. 信息头：指定了使用的签名算法
2. 声明部分：其中也可以包含超时时间
3. 基于指定的算法生成的签名

通过这三部分信息，API 服务端可以根据「JWT」信息头和声明部分的信息重新生成签名。之所以可以这样做，是因为生成签名需要的秘钥存放在服务器端。

```go
jwtauth.New("HS256", []byte("K8UeMDPyb9AwFkzS"), nil)
```

如果这个签名秘钥比较简单，建议立刻换一个复杂一些的，更改以后会使所有已经产生的「JWT」 失效，强制客户端重新从服务器获取授权。

## 声明部分

通过「JWT」的声明，可以用像 "user_id": "1337"  这样的格式来标识使用了 API 服务的客户端，可以把它想象成 map[string]interface{} 这样的 Go 数据结构，加上一些适当的转换。当客户端向 API 服务发起授权请求时，会发送客户端 ID 和一些其它数据来执行登录操作。服务器端接收到请求后会产生一个相应的 「JWT」，并保存在数据库中，这样客户端随后的请求就不需要再进行授权请求了，直到这个「JWT」超时。

应用最好生成一个调试「JWT」, 并输出到日志文件中。可以通过这个合法的「JWT」来调试应用。

```go
type JWT struct {
    tokenClaim string
    tokenAuth  *jwtauth.JWTAuth
}

func (JWT) new() *JWT {
    jwt := &JWT{
        tokenClaim: "user_id",
        tokenAuth:  jwtauth.New("HS256", []byte("K8UeMDPyb9AwFkzS"), nil),
    }
    log.Println("DEBUG JWT:", jwt.Encode("1"))
    return jwt
}

func (jwt *JWT) Encode(id string) string {
    claims := jwtauth.Claims{}.
        Set(jwt.tokenClaim, id).
        SetExpiryIn(30 * time.Second).
        SetIssuedNow()
    _, tokenString, _ := jwt.tokenAuth.Encode(claims)
    return tokenString
}
```

每当通过 JWT{}.new() 生成新的「JWT」对象时，就会在日志中输出调试「JWT」的信息

```
2018/04/19 11:35:18 DEBUG JWT: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.ZEBtFVPPLaT1YxsNpIzVGSnM4Vo7ZrEvp77jKgfN66s
```

你可以通过 URL 查询参数的方式传递这个「JWT」来测试 GET 请求，或者在测试代码中测试更加复杂的 API 请求，这时可以使用「授权头」或者 Cookie 进行传递

> **Note**: 生成「JWT」时一定要记得指定过期时间，否则生成的「JWT」会一直有效，直到更换了签名秘钥。另一个方案是在服务器端使个别「JWT」失效，这需要一些代码对它们进行记录和唤醒。比如，不使用用户 ID，而是用会话 ID 来标识「JWT」，这样就可以对 过期/登出 进行额外的验证
> 上面的例子中已经设置好了必要的参数，让我们可以对带有过期时间的「JWT」进行验证。如果请求一个受保护的 API ，并且「JWT」已经超时了，服务器会返回一个错误信息，提示你在调用这些接口时需要重新请求授权。

## 使用 JWT 保护 API 访问

每一个对 API 的请求都可以包含一个「JWT 检验器」。它的工作方式和 CORS 类似 - 从「HTTP 请求参数」、cookie 或者「授权 HTTP 头」中检测「JWT」是否存在。「检验器」返回一个关于「JWT」 的上下文变量和一个可能的解析错误，即使没有发现「JWT」，「检验器」也不会中断正常的请求，它只是向「授权器」提供一些信息。

[go-chi/jwtauth](https://github.com/go-chi/jwtauth) 提供了一个默认的「检验器」，我们可以直接使用，让我们为之前的「JWT」类型添加一个辅助方法，返回这个默认的「检验器」

```go
func (jwt *JWT) Verifier() func(http.Handler) http.Handler {
    return jwtauth.Verifier(jwt.tokenAuth)
}
```

我们在每一个请求中都添加了这个「检验器」，这样做以后，即使一个 API 不需要授权，也可以收到这些标识。提取并处理其中的声明，没有发现「JWT」时仍然可以返回一个合法的响应

```go
login := JWT{}.new()

mux := chi.NewRouter()
mux.Use(cors.Handler)
mux.Use(middleware.Logger)
mux.Use(login.Verifier())
```

之前我们使用的是 mux.Route，为了把请求分成需要授权和不需要授权两个部分，我们需要使用 [mux.Group()](https://godoc.org/github.com/go-chi/chi#Mux.Group)。使用 Group() 可以给全局处理器添加新的处理器，这样我们在请求时就可以省略像 "/api/private/*" 这样的前缀

> Group 会创建一个新的内联 Mux，它有一个空的中间件栈。如果一些请求的前缀部分是相同的，并且需要执行一些相同的中间件，就特别适合使用 Group 。

```go
// Protected API endpoints
mux.Group(func(mux chi.Router) {
    // Error out on invalid/empty JWT here
    mux.Use(login.Authenticator())
    {
        mux.Get("/time", requestTime)
        mux.Route("/say", func(mux chi.Router) {
            mux.Get("/{name}", requestSay)
            mux.Get("/", requestSay)
        })
    }
})

// Public API endpoints
mux.Group(func(mux chi.Router) {
    // Print info about claim
    mux.Get("/api/info", func(w http.ResponseWriter, r *http.Request) {
        owner := login.Decode(r)
        resputil.JSON(w, owner, errors.New("Not logged in"))
    })
})
```

现在 /time 和 /say 必需有一个合法的「JWT」才能访问，/time 不直接检验「JWT」，而是把检验工作交给了「授权器」。比如，我们用一个过期了的「JWT」访问 /time ，会得到如下的信息：

```json
{
    "error": {
        "message": "Error validating JWT: jwtauth: token is expired"
    }
}
```

但是如果我们请求的是 /info，我们会收到如下的信息：

```json
{
    "response": "1"
}
```

用一个过期的「JWT」访问 /info，则会返回：

```json
{
    "error": {
        "message": "Not logged in"
    }
}
```

两个请求返回不同的信息是因为我们在 Decode 函数中实现了完整的验证逻辑。如果「JWT」 是非法或者过期的，会返回一个自定义的错误信息，而不是用来保护 /time 的 Authenticate 方法中的返回值。

```go
func (jwt *JWT) Decode(r *http.Request) string {
    val, _ := jwt.Authenticate(r)
    return val
}

func (jwt *JWT) Authenticate(r *http.Request) (string, error) {
    token, claims, err := jwtauth.FromContext(r.Context())
    if err != nil || token == nil {
        return "", errors.Wrap(err, "Empty or invalid JWT")
    }
    if !token.Valid {
        return "", errors.New("Invalid JWT")
    }
    return claims[jwt.tokenClaim].(string), nil
}
```

我们用同样的方法让授权中间件使用「JWT」来保护对私有 API 的访问 。Decode() 方法中的错误被忽略了，因为被调用的方法默认返回了一个空字符串。授权中间件返回完整的错误信息：

```go
func (jwt *JWT) Authenticator() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            _, err := jwt.Authenticate(r)
            if err != nil {
                resputil.JSON(w, err)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

以上授权微服务的完整代码可以从 [GitHub](https://github.com/titpetric/books/tree/master/api-foundations/chapter4b-jwt) 上获取，可以免费下载和体验。

---

via: https://scene-si.org/2018/05/08/protecting-api-access-with-jwt/

作者：[Tit Petric](https://scene-si.org/about)
译者：[jettyhan](https://github.com/jettyhan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
