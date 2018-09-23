已发布：https://studygolang.com/articles/12350

# 如何只通过 Go 语言标准库创建 RESTful 接口

瑞安·麦丘 2017 年 12 月 9 日

Go 是一门相当新的语言，并且在最近几年得到了越来越多的关注。它的功能非常强大，而且拥有出色的工具来设计快速高效的 API 接口。虽然已经有很多库可以创建一个 API 接口，像 [Go Buffalo](https://gobuffalo.io/) 和 [Goa](https://goa.design/) 之类；但是，如果能够做到除了数据库和缓存连接器之外，仅仅使用标准库来创建，无疑将非常有趣。

在这篇博客中，我将分析如何使用 Go 语言标准库来创建一个端点（Endpoint）。整个 API（包括多个端点（Endpoint））代码在我的 GitHub  [golang-standard-lib-rest-api](https://github.com/rymccue/golang-standard-lib-rest-api) 库中。

## 入门指南

第一步，需要规划出整个目录结构，需要为控制器、请求、数据库迁移、数据库查询（存储库）和助手工具等创建目录。

项目目录结构如下：

```
controllers
database
models
repositories
requests
routes
utils
```

创建好这些目录后，让我们来创建 **main.go** 文件。这是创建的第一份源代码，因此将按照我们的构思来编写，然后再来创建周边相关的包。

我们需要 main.go 文件做以下的事情：

1. 创建数据库连接
2. 创建认证专用的缓存连接
3. 创建一个 mux
4. 加载所有的控制器
5. 使用 mux 和控制器创建路由
6. 启动服务器

```go
db, err := database.Connect(os.Getenv("PGUSER"), os.Getenv("PGPASS"), os.Getenv("PGDB"), os.Getenv("PGHOST"), os.Getenv("PGPORT"))
if err != nil {
	log.Fatal(err)
}
cache := &caching.Redis{
	Client: caching.Connect(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"), 0),
}

userController := controllers.NewUserController(db, cache)
jobController := controllers.NewJobController(db, cache)

mux := http.NewServeMux()
routes.CreateRoutes(mux, userController, jobController)

if err := http.ListenAndServe(":8000", mux); err != nil {
	log.Fatal(err)
}
```

我们希望代码看起来像上面的一样，创建了数据库和缓存连接、加载了控制器、将控制器连接到了路由，最终启动了服务器。现在，我们拥有了规划好的主入口文件，我们来创建需要使用的库。

## 连接工具

让我们从相对简单的数据库和缓存工具开始。

**database.go** 文件简单的创建了连接字符串，并且打开了一个连接。

**utils/database/database.go**

```go
func Connect(user, password, dbname, host, port string) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s",
		user, password, dbname, host, port)
	return sql.Open("postgres", connStr)
}
```

缓存包需要费一点功夫，我们使用了一个接口来抽象。这样的话，如果你决定使用另外的缓存服务时，代码修改就非常方便直接了。

**utils/caching/caching.go**

这个可以根据你的需要修改，不过在目前的使用场合，一个好的缓存接口如下所示：

```go
type Cache interface {
	Get(key string) (string, error)
	Set(key, value string, expiration time.Duration) error
}
```

现在创建一个结构体来实现这个接口：

```go
type Redis struct {
	Client *redis.Client
}

func (r *Redis) Get(key string) (string, error) {
	return r.Client.Get(key).Result()
}

func (r *Redis) Set(key, value string, expiration time.Duration) error {
	return r.Client.Set(key, value, expiration).Err()
}
```

最后，一个函数返回了 redis 客户端。

```go
func Connect(addr, password string, db int) *redis.Client {
	 return redis.NewClient(&redis.Options{
		 Addr:     addr,
		 Password: password,
		 DB:       db,
	})
}
```

## User 控制器

很好，缓存和数据库工具已经完成。现在开始做一些有趣的开发，先定义需要的路由。我们需要两个控制器，一个用户和一个工作控制器，每一个控制器都有一些端点（Endpoint）。在真正编写控制器之前，让我们把这些想法和设计落到纸面。

```
User Controller

POST /register
POST /login
```

```
Job Controller

GET /job/{id}
PUT /job/{id}
DELETE /job/{id}
POST /job
GET /feed
```

做好这些，我们现在就知道需要哪些端点（Endpoint）了。让我们来创建用户控制器和注册端点（Endpoint）。

如果你查看 **main.go**，你会发现有一个叫 **NewUserController** 的方法，让我们从那里开始起步。我们要在用户控制器中有能力访问缓存和数据库，所以需要将这两个对象传递给该函数。

该结构体需要像这样，所以我们在控制器方法内可以调用数据库和缓存，如 **userController.DB** 和 **userController.Cache**。注意缓存是一个接口, 因此我们可以随时替换它的实现。

```go
type UserController struct {
	DB    *sql.DB
	Cache caching.Cache
}
```

**NewUserController** 函数应该简单的返回带有数据库和缓存对象的 **UserController** 结构体。

```go
func NewUserController(db *sql.DB, c caching.Cache) *UserController {
	return &UserController{
		DB:    db,
		Cache: c,
	}
}
```

## 注册端点（Endpoint）

很棒，我们拥有了用户控制器，现在来增加一个方法。注册是一个很基本的功能，所以让我们来实现它。我们来一步一步地实现。

首先，我们只需要这个端点（Endpoint）的 POST 请求，所以首先检查是否是 POST (HTTP)方法，如果不是的话，返回一个 not found 状态给客户端。

```go
func (jc *UserController) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
```

检查通过后，需要解码请求的数据体来获取请求数据。如果数据体有问题，发送 bad request 状态到客户端。结构体定义将在本文的后面给出。

```go
	decoder := json.NewDecoder(r.Body)
	var rr requests.RegisterRequest
	err := decoder.Decode(&rr)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
```

现在我们已经拥有一个含有请求数据的结构体，可以用来创建用户了。因此，把该数据传递到存储库中，它会创建用户并返回 ID。如果这个阶段发生了错误，将返回一个 internal server 的错误。

```go
	id, err := repositories.CreateUser(jc.DB, rr.Email, rr.Name, rr.Password)
	if err != nil {
		log.Fatalf("Add user to database error: %s", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
```

最后，当用户被保存后，我们生成一个随机的令牌。以令牌为 key，用户的 ID 为值，将之保存到 Redis 缓存中。从而，可以在后续的请求中进行用户鉴权。

```go
	token, err := crypto.GenerateToken()
	if err != nil {
		log.Fatalf("Generate token Error: %s", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	oneMonth := time.Duration(60*60*24*30) * time.Second
	err = jc.Cache.Set(fmt.Sprintf("token_%s", token), strconv.Itoa(id), oneMonth)
	if err != nil {
		log.Fatalf("Add token to redis Error: %s", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
```

最后一步，将令牌返回给用户，并设置内容类型为 json 格式。

```go
	p := map[string]string{
		"token": token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}
```

## 最后一些事项

现在，控制器和端点（Endpoint）都完成了，但是你会注意到有一些事项被遗漏了。首先，我们需要一个请求对象来承载请求数据。

```go
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
```

最后一步是创建用户存储库。当创建好 **repositories/user_repository.go** 文件后，需要添加 **CreateUser** 函数。它将生成一个随机种子用于给密码签名，并在用户表中保存所有的数据。

```go
func CreateUser(db *sql.DB, email, name, password string) (int, error) {
	const query = `
		insert into users (
			email,
			name,
			password,
			salt
		) values (
			$1,
			$2,
			$3,
			$4
		) returning id
	`
	salt := crypto.GenerateSalt()
	hashedPassword := crypto.HashPassword(password, salt)
	var id int
	err := db.QueryRow(query, email, name, hashedPassword, salt).Scan(&id)
	return id, err
}
```

这个完整的加密代码段在整篇文章中被用到，你可以在[这里](https://github.com/rymccue/golang-standard-lib-rest-api/blob/master/utils/crypto/crypto.go)找到它的代码。在本文中，它不是最核心的组成部分，你可以自由的使用它。

## 结论

这个控制器只是我创建的 API 的一小段代码，用来展示仅用 Go 语言标准库来创建 API 是多么简单的一件事情。Go 语言是一门伟大的语言，可以用来创建 API 和微服务，并且执行效率胜过了如 javascript 和 PHP 之类的其他语言。所以，使用 Go 语言是一件非常简单的事情。虽然只使用标准库来实现了这些，但是我相信使用一些外部的库，如 [Gorilla Mux](https://github.com/gorilla/mux) 和 [go-validator](https://github.com/go-validator/validator) ，会使得开发更加简便，而且使得代码更加清晰和可维护。

---

via: https://ryanmccue.ca/how-to-create-restful-api-golang-standard-library/

作者：[Ryan McCue](https://ryanmccue.ca/author/ryan/)
译者：[arthurlee](https://github.com/arthurlee)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

