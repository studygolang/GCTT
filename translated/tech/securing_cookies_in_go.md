# Go 实现安全 cookie

我第一次学习 Go 的时候，已经有了一定的 web 开发经验，但是直接使用 cookie 的经验还比较少。我之前是用 Rails 的，在 Rails 里面如果要读/写 cookie，并不需要自己去实现所有的安全措施。

如你所见，Rails 偏向于通过默认配置就将大部分事情搞定。你不必去设置 CSRF 反制措施，或者做任何特殊操作来加密 cookie。在较新版本的 Rails 中这一切都默认帮你做好了。

用 Go 开发就完全不同了，这些事情并没有默认帮你做好。所以当你想使用 cookie 时，了解所有这些安全措施就很重要：为什么存在有这些措施，以及如何在我们自己的程序中使用它们。本文旨在帮助你了解上述知识点。

> 本文目的不在于引发关于哪条路线更好的讨论/争论。两者都有各自的优点，这里不去比较 Rails 和 Go 孰是孰非，而是关注如何确保 cookie 的安全性。

## 什么是 cookie？
在介绍 cookie 的安全措施之前，需要理解 cookie 到底是什么。本质上，cookie 就是存储在终端用户设备上的 key/value 对。这样，创建一个 cookie ，你就只需要设置[http.Cookie](https://golang.org/pkg/net/http/#Cookie)类型的`Name`和`Value`字段，然后调用[http.SetCookie](https://golang.org/pkg/net/http/#SetCookie)函数来告诉终端用户的浏览器去设置 cookie。

代码估计长这样：

```go
func someHandler(w http.ResponseWriter,r *http.Request){
	c := http.Cookie{
		Name: "theme",
		Value:"dark",
	}
	http.SetCookie(w,&c)
}
```

> `SetCookie`不会返回错误
> 
> `http.SetCookie`不会返回一个错误值，但是会默默地将非法 cookie 丢弃掉。这不是什么好的体验，但现实已经是这样了，所以调用该函数时，一定要铭记此点。

在代码中，表现出来是我们在设置一个 cookie，实际上，我们只是将想要设置的 cookie 放在一个响应体的"`Set-Cookie`"头中。我们不会把 cookie 存放到服务器上，而是依赖于终端用户的浏览器去创建、存储 cookie。

我必须要强调这些，因为对安全措施有非常大的影响。我们**不会去**控制这部分数据，终端用户设备（即终端用户）最终会去控制这部分数据。

读写终端用户最终控制的数据时，我们需要非常谨慎对待如何处理这部分数据。恶意用户可能会删除 cookie，修改 cookie 中存储的数据，或者甚至可能会遇到[中间人攻击](https://en.wikipedia.org/wiki/Man-in-the-middle_attack)，黑客可能会截取用户发往服务器的 cookie。

## cookie 可能遇到的潜在问题

以我的经验，cookie 涉及的安全性问题大致分为五个大类。本节后面会对每一类做个简述，文章后几节会详细讨论每种情况的细节以及反制措施。

1.**Cookie 盗用（theft）** — 攻击者有多种方式尝试盗用 cookie。我们会讨论如何阻止/缓和大部分的情况，但是对于物理设备入侵事实上我们没办法完全阻止。

2.**Cookie 篡改（tampering）** — 不论是否故意，cookie 中的数据可以被篡改。我们将讨论如何验证 cookie 中保存的数据就是之前写入的有效数据。

3.**数据泄露（Data leaks）** — cookie 保存在终端用户的设备上，所以保存数据时要小心，乙方数据泄露。

4.**跨站脚本攻击（Cross-site scripting（XSS））** — 尽管不是直接和 cookie 有关联，如果 XSS 攻击有访问 cookie 的权限，它会变得非常强大。我们应当考虑防止 cookie 被不需要访问它的脚本访问。

5.**跨站请求伪造（Cross-site Request Forgery（`CSRF`））** — 此类攻击通常依赖用户通过 cookie 中保存的 session 登录的情况，所以我们会讨论当以这种方式使用 cookie 时候如何防止被黑。

如上所述，本文我们会逐一解决这些问题，看完后你就可以像老手一样确保你 cookie 的安全。

## Cookie 盗用

顾名思义，Cookie 盗用就是黑客盗取了用户的 cookie，通常为了伪装成被盗用的用户。

通常有两种盗用 cookie 的方式：

1.[中间人攻击](https://en.wikipedia.org/wiki/Man-in-the-middle_attack)，或者其他类似的行为，攻击者拦截了你的 web 请求，然后从中盗取 cookie 数据。

2.获取到访问硬件的权限。

防止中间人攻击基本上归结为当你的网站用到了 cookie，那么你一定要使用 SSL。通过使用 SSL，可以保证其他人几乎不可能截获你的请求，因为他们无法破解数据。

对于有"ahh，中间人攻击可能不常见。。。"想法的人，强烈推荐去看[firesheep](http://codebutler.com/firesheep)，是一个简单工具，用于展示通过公共 wifi 盗取未加密 cookie 有多简单。

如果你想确保这种事情不发生在你用户身上，**配置 SSL!**。[Caddy 服务器](https://caddyserver.com/)通过 Let's Encrypt 让配置变得非常简单。使用它就好了。对于配置生产环境来说真的是非常的简单。例如，4行代码就可以简单了代理你的 Go 应用：

```
calhoun.io{
	gzip
	proxy / localhost:3000
}
```

Caddy 会自动处理涉及 SSL 的一切。

防止访问硬件盗用 cookie 是个更加复杂的场景。我们不太可能强制用户使用安全的系统或者在设备上使用合适的密码，所以总归是会有某人坐下来使用你的电脑、盗取 cookie 后离开的风险存在。cookie 也有可能被病毒盗取，所以如果用户点击了一个恶意附件，可能就面临类似的情况。

更有挑战的是，很难去发现这种情况。如果有人偷了你的手表，发现表不在你手腕上时候，你会意识到被偷。然而，Cookie 可能会在没有人意识到的情况下被拷贝，然后被使用。

尽管不是一个安全模式，你可以使用一些技巧检测到丢失 cookie。例如，你可以追踪用户登录的设备，标记任何新的设备，要求他们重新输入密码。你也可以追踪 IP 地址，有可以登录地点时候警告用户。

所有这些方案需要额外的一些某段工作来跟踪数据，如果你的应用处理敏感数据、金钱或者正处在上升期，这也是你的工作的方向。

也就是说，对于大多数应用，第一版采用这些措施完全足够，使用 SSL 对于发布已经可以了。

## Cookie 篡改（tampering）（即用户假数据）
我们需要面对一个现实——有些人比较”混蛋“，他们会尝试查看设置好的 cookie，让后改变其值。即便有时候只是出于好奇这么做，只要这种情况有发生的可能性，我们就必须做好准备（应对）。

有些情况我们不太关心。例如，如果我们允许用户定义主题偏好，用户做了更改的话，我们一般不会关心。如果有非法的操作，我们就恢复成默认主题就好了，如果用户改成一个有效的主题，那么我们就直接用哪个主题就好了，也不会对系统有任何损害。

而对于其他情况，会需要考虑的更多些。修改会话 cookie 并尝试冒充其他用户比改变主题性质要严重的多，我们肯定不希望 Joe 假扮成 Sally。

我们会讨论两种检测、阻止修改 cookie 的策略。

### 1.对数据数字签名

数字签名是对数据添加一个签名，以便验证其真实性。终端用户无需对数据进行加密或做掩码，但是我们需要向 cookie 添加足够的数据，这样如果用户更改了数据的话，我们能够检测出来。

通过哈希来实现这个方案——会对数据进行 hash，然后将数据和数据的哈希值都存到 cookie 中。之后当用户发送 cookie 给我们，我们会对数据再次做哈希，验证是否和之前的哈希值匹配。

我们也不希望用户创建新的哈希值，所以你通常会看到使用 HMAC 这类哈希算法，通过一个密钥对数据做哈希。防止用户同时修改数据以及数字签名（哈希值）。

> [JSON Web Tokens (JWT)](https://jwt.io)内置了数字签名功能，这种方法可能你早就比较熟悉了。

在 Go 里面的话，可以用 Gorilla 的[securecookie](http://www.gorillatoolkit.org/pkg/securecookie)包，创建`SecureCookie`的时候提供一个哈希 key，利用该对象确保 cookie 的安全性。

```go
// It is recommended to use a key with 32 or 64 bytes, but
// this key is less for simplicity.
var hashKey = []byte("very-secret")
var s = securecookie.New(hashKey,nil)

func SetCookieHandler(w http.ResponseWriter,r *http.Request){
	encoded,err:=s.Encode("cookie-name","cookie-value")
	if err == nil{
		cookie := &http.Cookie{
			Name: "cookie-name",
			Value: encoded,
			Path:"/",
		}
		http.SetCookie(w,cookie)
		fmt.Fprintln(w,encoded)
	}
}
```

你可以在另外的处理器中使用同一个 SecureCookie 对象来获取这个 cookie。

```go
func ReadCookieHandler(w http.ResponseWriter, r *http.Request) {
  if cookie, err := r.Cookie("cookie-name"); err == nil {
    var value string
    if err = s.Decode("cookie-name", cookie.Value, &value); err == nil {
      fmt.Fprintln(w, value)
    }
  }
}
```

> 例子来源于[http://www.gorillatoolkit.org/pkg/securecookie](http://www.gorillatoolkit.org/pkg/securecookie)的示例。

>这里并没有加密数据，只是编码了。在”数据泄露“部分我们会讨论如何加密数据。

这里有个非常重要的警告：对于同时往数字签名的数据中添加用户信息和过期时间的情况，如果用上述方法保证可靠性，你必须非常小心，严格遵守 JWT 的使用模式。不能单单依赖 cookie 的过期时间，因为该日期未被加密，用户可以创建一个新的没有过期时间的 cookie，然后把 cookie 签名的部分拷贝过去，基本上就是创建了一个保证他们永久在线的 cookie。

### 2.混淆数据

另外一种方式是对数据做掩码，确保用户无法伪造数据。例如，不要像如下方式一样保存 cookie：

```go
// Don't do this
http.Cookie{
  Name: "user_id",
  Value: "123",
}
```

我们可以保存一些数据值，通过这些数据值能够映射到数据库真实的数据。一般通过 session ID 或者记录 token 实现，有一个叫做`remember_tokens`的表来记录数据：

```
remember_token: LAKJFD098afj0jasdf08jad08AJFs9aj2ASfd1
user_id: 123
```

然后就可以只在 cookie 中保存记录 token，这样即便用户想要伪造，也不知道要改什么。它看起来就像乱码。

后面当用户访问我们的应用，我们会在数据库中查找其记录 token，然后判断是哪个用户登录了。

为了使该方案能够运行，你需要确保混淆数据是:

* 映射到了一个用户（或者其他资源）
* 随机的
* 熵值较高
* 可以设为失效状态（例如，删除/改变 DB 中保存的 token）

这个方法的一个缺点是，对于每个需要验证用户身份的页面请求，都需要进行数据库查询，不过这个缺点一般不会被注意到，可以通过缓存或其他类似技术解决掉。该方法相对 JWT 的优点是你可以快速废弃 session。

> 这是我知道的最常见的验证策略，尽管 JWT 最近在所有的 JS 框架得到流行。

## 数据泄露（Data leaks）

像 cookie 盗用一样，在成为真正的威胁之前，数据泄露通常需要有其他的攻击途径，不过谨慎一些总是好的。也是因为，cookie 被盗并不意味着我们想要故意告诉黑客用户密码。

无论何时往 cookie 中保存数据，都要尽可能减少存储敏感数据的量。不要存储用户的密码，确保编码过得数据中也没有密码。类似[这篇](https://hackernoon.com/your-node-js-authentication-tutorial-is-wrong-f1a3bf831a46#2491)文章指出的几个例子，开发者不知不觉地在 cookie 或者 JWT 中保存了敏感数据，采用 base64 编码，但实际上任何人都可以解码该数据。数据是被编码了，而**不是加密**了。

早前，我们讨论了如何对 cookie 数字签名，但是`securecookie`也可以用于加/解密你的 cookie 数据，故而其不太会被轻易解码、访问到。

启用该库的加密功能，你只需要在创建`SecureCookie`实例时候，简单地传入 block key。

```go
var hashKey = []byte("very-secret")
// Add this part for encryption.
var blockKey = []byte("a-lot-secret")
var s = securecookie.New(hashKey, blockKey)
```

其他和文章数字签名部分例子类似。

还是要着重强调下，**不要**在 cookie 中保存任何敏感的数据；


