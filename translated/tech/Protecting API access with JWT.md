# 通过 JWT 访问受保护的 API

API 的一个常见用途是提供一个授权中间件，允许客户端向 API 发送授权请求。
通常来说，客户端也会有一些授权机制，比如「会话标识」。最近比较流行的
JWT(JSON Web Tokens) 提供了一个带超时时间的「会话标识」，
它不需要额外的空间来执行验证逻辑。

本文是接着上一篇文章写的，在阅读下面内容之前建议先看一下之前的那篇文章 [用 go-chi 处理 HTTP 请求](https://scene-si.org/2018/03/12/handling-http-requests-with-go-chi/)

接下来我们要用基于 [go-chi/chi](https://github.com/go-chi/chi) 的
[go-chi/jwtauth](https://github.com/go-chi/jwtauth) 在 API 上增加一个授权层。
授权可以是可选的（用于登录或者退出的用户）或者强制的（只针对已经登录的用户）。
这样可以实现一个独立的授权逻辑，基于 JWT 的参数的合法性来丰富 API 的返回信息。
实现中我用了 [titpetric/factory/resputil](https://github.com/titpetric/factory/tree/master/resputil)
来简化错误处理和 JSON 数据的格式化。

## JWT 到底是什么

> JSON Web Token (JWT) 是一个开源标准 ([RFC 7513](https://tools.ietf.org/html/rfc7519))
，定义了一种以一个简洁并自包含的方式，在各部分之间安全传递 JSON 数据。因为它经过了数字签名，
所以传送的数据是可验证和可信的，

「JWT」由三部分构成

1、信息头：标识了签名使用的算法

2、声明部分：可能包含到期时间

3、基于头部分指定的算法生成的签名

有了这三部分信息，API 服务器端可以根据 JWT 中头部和声明部分的信息重新生成签名。之所以可以这样做，
是因为生成签名需要的秘钥只有服务器才知道。

```
jwtauth.New("HS256", []byte("K8UeMDPyb9AwFkzS"), nil)
```

如果这个秘钥比较简单，建议你立刻换一个，这会让已经产生的 JWT 失效，强制客户端重新从服务器获取授权。

## 声明

JWT 中的声明部分，可以方便的标识一个客户端使用了你的 API 服务，是像 "user_id": "1337" 这样的格式。
可以把它想象成 map[string]interface{} 这样的数据结构，再加上一些适当的转换。当客户端向你的 API
服务发起授权请求，请求中包含客户端ID和一些其它数据。服务器端会产生一个相应的 JWT，并保存在数据库中，
在这个 JWT 超时之前，客户端随后的请求就不需要再次请求请求授权了。

一个比较好的做法是产生一个调试「JWT」, 并输出到日志文件中，这样，就可以通过这个合法的「JWT」来调试你的应用。

```
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

每当通过 JWT{}.new() 生成一个新的 「JWT」 对象时，就会在日志中输出调试「JWT」

```
2018/04/19 11:35:18 DEBUG JWT: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.ZEBtFVPPLaT1YxsNpIzVGSnM4Vo7ZrEvp77jKgfN66s
```

你可以通过 URL 查询参数的方式传递这个「JWT」来测试 GET 请求，或者把它放在授权头部或者 Cookie 中进行传递，来测试更加复杂的 API 请求

> **Note**: 生成「JWT」时一定要记得指定过期时间，否则在你更改签名密钥之前，这个把这个「JWT」会一直有效。
另一个方案是在服务器端使个别「JWT」失效，这需要一些代码对他们进行记录和唤醒。
比如，产生「JWT」 时不用用户 ID，而是用会话 ID，这样就可以对 过期/登出 进行额外的验证
> 例子中已经设置好了必要的参数，让我们可以验证带有过期时间的「JWT」。如果请求一个受保护的 API 时，「JWT」已经超时了，服务器会返回一个错误信息，提示你在调用这些接口时需要重新请求授权。

## 通过 JWT 访问受保护 API

每一个对 API 的请求都可以包含一个「JWT 检验器」。它和 CORS 的工作方式很相似 -
它从 HTTP 的请求参数、cookie 或者授权 HTTP 头中检测「JWT」是否存在。「检验器」
返回一个关于「JWT」 的上下文变量和一个可能的解析错误，即使没有发现「JWT」 ，「检验器」也不会中断正常的请求，它的工作只是向授权程序提供一些信息。

[go-chi/jwtauth](https://github.com/go-chi/jwtauth) 提供了一个默认的「检验器」，
你可以直接使用，让我们为之前的「JWT」 类型添加一个辅助方法，返回这个默认的「检验器」

```
func (jwt *JWT) Verifier() func(http.Handler) http.Handler {
    return jwtauth.Verifier(jwt.tokenAuth)
}
```

我们在每一个请求中都包含了「检验器」，这样可以向一个没有显示要求只接受授权用户的 API
发关「JWT 」。这样你就可以获取并处理每一个声明，没有发现「JWT」 时仍然可以返回一个合法的响应

```
login := JWT{}.new()

mux := chi.NewRouter()
mux.Use(cors.Handler)
mux.Use(middleware.Logger)
mux.Use(login.Verifier())
```

为了把请求分为受保护的和公共的，我们需要用 [mux.Group()](https://godoc.org/github.com/go-chi/chi#Mux.Group),
而不是之前的 mux.Route。这允许我们为全局处理器添加新的处理器，这样我们在请求时就可以不用写像 "/api/private/*"
这样的前缀

> Group 会创建一个新的内联 Mux，它有一个空的中间件栈。它对拥有相同的请求路径并且需要一些额外的中间件的一组处理器来说，特别有用。

```
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

现在 /time 和 /say 必需有一个合法的「JWT」才能访问，/time
不直接检验「JWT」，而是把这个工作交给了「授权器」。比如，
我们用一个过期了的「JWT」访问 /time ，会得到如下的信息：

```
{
    "error": {
        "message": "Error validating JWT: jwtauth: token is expired"
    }
}
```

但是如果我们请求的是 /info，我们会收到如下的信息：

```
{
    "response": "1"
}
```

用一个过期的「JWT」访问 /info，则会返回：

```
{
    "error": {
        "message": "Not logged in"
    }
}
```

两个请求返回不同的信息是因为我们在 Decode 函数中实现了完整的验证。如果「JWT」 是非法或者过期的，会返回一个自定义的错误信息，而不是在 /time 请求路径上的「验证器」默认的返回值

```
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

我们用同样的方法让授权中间件使用「JWT」 JWT 来访问受保护的 API。Decode() 方法中的错误被忽略了，因为被调用的方法已经默认返回了一个空的字符串。授权中间件返回完整的错误信息：

```
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

以上授权微服务的完整代码可以从
[GitHub](https://github.com/titpetric/books/tree/master/api-foundations/chapter4b-jwt)
上获取，可以免费下载和体验

----------------

via: <https://scene-si.org/2018/05/08/protecting-api-access-with-jwt/>

作者：[Tit Petric](https://scene-si.org/about)
译者：[jettyhan](https://github.com/jettyhan)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，
[Go 中文网](https://studygolang.com/) 荣誉推出

