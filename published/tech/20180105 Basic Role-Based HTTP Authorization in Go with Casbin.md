已发布：https://studygolang.com/articles/12323

# 在 Go 语言中使用 casbin 实现基于角色的 HTTP 权限控制

身份认证和授权对 web 应用的安全至关重要。最近，我用 Go 完成了我的第一个正式的 web 应用，这篇文章是在这个过程中我所学到的部分内容。

本文中，我们的关注点在于如何在 web 应用中使用开源的 casbin 库进行 HTTP 权限控制。同时，在示例代码中我们使用了 scs 库进行 session 管理。

下面的例子十分基础，希望它尽可能的展示了如何在 Go web 应用中实现权限控制。为了更侧重于展示 casbin 的使用，我们尽量简化业务逻辑（例如：不需密码的登陆操作）。我们一起来看一下！

注意：请不要在生产环境中使用所示的用例代码，该例子侧重于描述清晰，而不是安全性。

## 建立

首先，我们创建一个 User 模型，并实现了相应方法：
```	go
type User struct {
	ID   int
	Name string
	Role string
}

type Users []User

func (u Users) Exists(id int) bool {
	...
}

func (u Users) FindByName(name string) (User, error) {
	...
}
```	

接着配置 casbin 所需文件。这里我们需要一个配置文件和一个策略文件。配置文件使用 PERM 元模型。PERM 表示策略（Policy）、效果（Effect）、请求（Request）和匹配器（Matchers）。

在 auth_model.conf 配置文件中有如下内容：

```
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*")
```
其中定义了请求和策略来表示主体，客体和动作。在本例中，主体表示用户角色，客体表示访问路径，action 表示请求方法（例：GET, POST 等）。

匹配器定义了策略是如何匹配的，可以通过直接定义主体，或者使用像 keyMatch 这样的帮助方法，它也可以匹配通配符。casbin 实际比这个简单的例子要强大得多，你可以用声明的方式定义各种自定义功能来达到轻松切换和维护鉴权配置的效果。

在安全性方面，我通常会选择最简单的解决方案，因为当系统开始变复杂和难以维护时，错误就开始发生。

在这个例子中，策略文件就是一个简单的 csv 文件，描述了哪些角色可以访问哪些路径等。

policy.csv 文件格式如下：

```
p, admin, /*, *
p, anonymous, /login, *
p, member, /logout, *
p, member, /member/*, *
```
这个配置文件十分简单。在这个例子中，我们简单的定义了 admin 角色可以访问所有内容，member 角色可以访问以 /member/ 开头的路径和 logout 路径，未认证用户可以登陆。

这种形式的好处在于即使应用具有许多规则和用户角色，它仍然是可维护的。

## 执行

让我们从 main 函数开始，将所有的东西都配置好，并启动 http 服务器：

```go
func main() {
	// setup casbin auth rules
	authEnforcer, err := casbin.NewEnforcerSafe("./auth_model.conf", "./policy.csv")
	if err != nil {
		log.Fatal(err)
	}
	// setup session store
	engine := memstore.New(30 * time.Minute)
	sessionManager := session.Manage(engine, session.IdleTimeout(30*time.Minute), session.Persist(true), session.Secure(true))
	
	// setup users
	users := createUsers()
	
	// setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler(users))
	mux.HandleFunc("/logout", logoutHandler())
	mux.HandleFunc("/member/current", currentMemberHandler())
	mux.HandleFunc("/member/role", memberRoleHandler())
	mux.HandleFunc("/admin/stuff", adminHandler())
	
	log.Print("Server started on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", sessionManager(authorization.Authorizer(authEnforcer, users)(mux))))
}
```
这里有几点需要注意的，通常，我们需要配置鉴权规则，session 管理，用户，http 处理方法，启动 http 服务器，并且用鉴权中间件和 session 管理器包装路由。

我们逐一分析上面的过程。

首先，我们用上面的 auth_model.conf 和 policy.csv 创建了一个 casbin 执行器。如果出错了，则关闭服务，因为有可能鉴权规则出错了。

第二步是设置会话管理器。我们创建了一个具有 30 分钟超时的内存 session 存储和和一个具备安全 cookie 存储的会话管理器。

CreateUsers 函数创建了三个不同的用户，其用户角色如下所示：

```go
func createUsers() model.Users {
	users := model.Users{}
	users = append(users, model.User{ID: 1, Name: "Admin", Role: "admin"})
	users = append(users, model.User{ID: 2, Name: "Sabine", Role: "member"})
	users = append(users, model.User{ID: 3, Name: "Sepp", Role: "member"})
	return users
}
```
在实际应用中，我们会使用数据库来存储用户数据，在这个例子中，为了方便起见我们使用上面的列表。

接下来是登陆和注销的处理方法：

```go
func loginHandler(users model.Users) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PostFormValue("name")
		user, err := users.FindByName(name)
		if err != nil {
			writeError(http.StatusBadRequest, "WRONG_CREDENTIALS", w, err)
			return
		}
		// setup session
		if err := session.RegenerateToken(r); err != nil {
			writeError(http.StatusInternalServerError, "ERROR", w, err)
			return
		}
		session.PutInt(r, "userID", user.ID)
		session.PutString(r, "role", user.Role)
		writeSuccess("SUCCESS", w)
	})
}

func logoutHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := session.Renew(r); err != nil {
			writeError(http.StatusInternalServerError, "ERROR", w, err)
			return
		}
		writeSuccess("SUCCESS", w)
	})
}
```
对于登陆，我们从请求中获取到用户名，检查该用户是否存在，若存在，则创建一个新的 session，并将用户角色和 ID 存入 session 中。

对于注销，我们创建一个新的空的 session，并从 session 存储中删除旧的 session，注销该用户。

接着，我们定义了几个处理函数，通过返回用户 ID 和角色来测试应用的实现。这些处理函数的端点由上面的 policy.csv 文件定义的 casbin 保护。

```go
func currentMemberHandler() http.HandlerFunc {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	uid, err := session.GetInt(r, "userID")
	if err != nil {
		writeError(http.StatusInternalServerError, "ERROR", w, err)
		return
	}
	writeSuccess(fmt.Sprintf("User with ID: %d", uid), w)
})
}
func memberRoleHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, err := session.GetString(r, "role")
		if err != nil {
			writeError(http.StatusInternalServerError, "ERROR", w, err)
			return
		}
		writeSuccess(fmt.Sprintf("User with Role: %s", role), w)
	})
}

func adminHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeSuccess("I'm an Admin!", w)
	})
}
```
我们可以通过 session.GetInt 和 session.GetString 来获取当前 session 中的值。

为了让鉴权机制真正的保护到处理函数，我们需要实现一个用来封装路由的鉴权中间件。

```go
func Authorizer(e *casbin.Enforcer, users model.Users) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role, err := session.GetString(r, "role")
			if err != nil {
				writeError(http.StatusInternalServerError, "ERROR", w, err)
				return
			}

			if role == "" {
				role = "anonymous"
			}

			// if it's a member, check if the user still exists
			if role == "member" {
				uid, err := session.GetInt(r, "userID")
				if err != nil {
					writeError(http.StatusInternalServerError, "ERROR", w, err)
					return
				}
				exists := users.Exists(uid)
				if !exists {
					writeError(http.StatusForbidden, "FORBIDDEN", w, errors.New("user does not exist"))
					return
				}
			}

			// casbin rule enforcing
			res, err := e.EnforceSafe(role, r.URL.Path, r.Method)
			if err != nil {
				writeError(http.StatusInternalServerError, "ERROR", w, err)
				return
			}
			if res {
				next.ServeHTTP(w, r)
			} else {
				writeError(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
				return
			}
		}

		return http.HandlerFunc(fn)
	}
}
```
	
鉴权中间件以 casbin 规则执行器和用户作为参数。首先，它从 session 中获取到请求用户的角色。若用户没有角色，设置为 anonymous 角色，否则，若用户角色为 member，我们将 session 中的 useID 和用户列表相比对，来判断用户是否合法。

在这些初步的检查之后，我们可以将用户角色，请求路径和请求方法传给 casbin 执行器，执行器决定了具有该角色（ subject ）的用户是否允许访问由该请求方法（ action ）和路径（ object ）指定的资源。若校验失败，则返回 403 ，若通过，则调用包装的 http 处理函数，允许用户访问请求资源。正如主函数中提及的，session 管理器和鉴权器对路由进行了包装，所以每个请求都需要通过这个中间件，确保了安全性。
​
我们可以通过登陆不同的用户，用 curl 或 postman 访问上述的处理函数来测试效果。

## 结论

我已经在一个中型 web 应用生产环境中使用了 casbin，并且对它的可维护性和稳定性感到十分满意。可以看看它的文档，casbin 是一个非常强大的鉴权工具，以声明的方式提供了大量的访问控制模型。

本文旨在展示 casbin 和 scs 的强大之处，并且展示 go web 应用的简洁清晰之处。

资源：

- [代码](https://github.com/zupzup/casbin-http-role-example)
- [casbin](https://github.com/casbin/casbin)
- [scs](https://github.com/alexedwards/scs)

----------------

via: https://zupzup.org/casbin-http-role-auth/

作者：[Mario](https://zupzup.org/about/)
译者：[linyy1991](https://github.com/linyy1991)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
