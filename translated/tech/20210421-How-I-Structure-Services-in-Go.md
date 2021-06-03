# 如何在 Go 中组织项目结构

译者注：在翻译这篇文章之前，我自己其实对 Bob 大叔的 Clean Architecture 也做过一些研究，在项目中实践之后，
也确确实实体验到了分层的魅力。在层与层之间将依赖进行隔离，各个层只关注自己本身的逻辑，
所以能让开发者只关注本层的业务逻辑， 也更容易进行单元测试，无形中就提高了你代码的质量和可阅读性。
我觉得如果你对自己的代码有追求，就一定要去学习一下 Clean Architecture。
当然另一方面，Clean Architecture 也不是银弹，在复杂的项目中确实能帮助我们解藕，但是如果你的项目非常简单，
那传统的 MVC 就足够了，就像本文作者最后说的，千万不要让简单的事情变复杂。
另外，其实对于 Golang 的项目组织方式，github 上面火许多 star 非常多的项目，大多开箱即用，
比如：
[go-gin-api](https://github.com/xinliangnote/go-gin-api) （国人开源的）、
[go-clean-arch](https://github.com/bxcodec/go-clean-arch) ，这里分享给大家，也是给大家提供更多的选择。

以下是原文：

一个 main.go 文件，几个 HTTP handler 就可以构建一个新的 HTTP 服务。然而，当你开始添加更多的路由规则，
开始将不同的功能拆分到不同的文件中，可能会随处创建好多 packages ，但是你不确定长远来看将会它们怎样发展，
同时你也希望它们能够随着服务增长而有意义。

这几年我经历过几次这样的场景，后来读了一些文章、博客还有 Robert Martin 的
[Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) (干净架构)，
我找到了一种适合我的通用的代码结构，所以我想我应该分享出来。需要注意的是它可能并不能刚好适合你的应用场景，特别是一些特别简单的服务，
比如说一个 main.go 文件，几个 packages 就已经足够的服务。

让我们直接开始！

## 概览

```goalng
Go-Service
- cmd/
  - api/
- pkg/
  - api/
  - db/
  - services/
    - serviceA/
    - serviceB/
    - ...
  - utils/
- docker-compose.yml
- Dockerfile
- Makefile
- go.mod
- go.sum
- <environment>.env
- README.md
- ...
```

整个结构有 3 个重要的部分：root（译者注：根目录）、cmd 和 pkg。我将逐个解释各个文件夹的职责，
然后我们再来仔细看看每个 service (`pkg/service/...`) 如何组织。

### Root（根目录）

我喜欢将一些启动和运行的代码放到根目录，比如：构建工具、配置文件、依赖管理等等。
它也提供给阅读代码的人或开发代码的人一个很好的切入点，他们启动服务所需的所有配置都在项目根目录下。

### Cmd

这里会被分成几个目录，每个目录都是我们整个服务一部分，比如 API 服务，定时脚本任务等等。
实际上这里会有各子服务的 main package，所以我们在这里初始化配置和我们需要的依赖包，最后子服务会被编译成对应的二进制来提供服务。

### Pkg

这里包含了我们项目的主要部分：定义我们服务业务逻辑的一些 package。

* api/

  在这里，我定义了如何通过初始化数据库，服务，HTTP路由器+中间件来连接API，并定义了运行API所需的配置。
  我一般会加一个 `Start(cfg *Config)` 函数，提供给 `cmd/api/main.go` 调用。

* db/

  顾名思义，这里是连接、迁移数据库逻辑，我也倾向于将任何关于迁移的文件夹或文件都放在这里。

* utils/

  我会将任何对请求、日志、自定义中间件等提供辅助功能的 pakcage 放在这里。我虽然不太喜欢这个名字，但是我也没找到更适合它的名字了。

* services/

  这个需要详细解释一下，因为我用特定的方式去组织所有的 service。通常来说，每个 package
  都定义了各自服务的功能（基于功能而不是函数进行组织结构）。

## Services

让我们通过一个例子来看看他们是如何组织的。我们要创建一个服务，可以让我们保存并创建文章，他看起来是下面这样：

```
...
- Services/
  - Article/
    - store/
      - repo.go
    - transport/
      - http.go
    - article.go
    - errors.go
    - models.go
```

我们将数据的存储和传输逻辑分到了不同的 package，这帮助我们专注于我们的业务逻辑而不需要关心我们应该如何保存数据或者其如何传递给调用方。
此外，当我们想要改变我们的底层存储时，我们只需要定义好存储的 interface，就可以轻松地更换底层存储，
而不需要修改其余的逻辑（ 一个简单的
[依赖反转原则](https://en.wikipedia.org/wiki/Dependency_inversion_principle) 的例子 ）。

`error.go` 和 `models.go` 比较简单，就不赘述了，让我们看看 `article.go` 都有什么功能：

```golang
package articles

import (
    "context"
)

// Repo defines the DB level interaction of articles
type Repo interface {
    Get(ctx context.Context, id string) (Article, error)
    Create(ctx context.Context, ar ArticleCreateUpdate) (string, error)
}

// Service defines the service level contract that other services
// outside this package can use to interact with Article resources
type Service interface {
    Get(ctx context.Context, id string) (Article, error)
    Create(ctx context.Context, ar ArticleCreateUpdate) (Article, error)
}

type article struct {
    repo Repo
}

// New Service instance
func New(repo Repo) Service {
    return &article{repo}
}

// Get sends the request straight to the repo
func (s *article) Get(ctx context.Context, id string) (Article, error) {
    return s.repo.Get(ctx, id)
}

// Create passes of the created to the repo and retrieves the newly created record
func (s *article) Create(ctx context.Context, ar ArticleCreateUpdate) (Article, error) {
    id, err := s.repo.Create(ctx, ar)
    if err != nil {
        return Article{}, err
    }
    return s.repo.Get(ctx, id)
}

```

这里需要注意的是，在调用 `New()` 创建我们 `Article` 服务实例的时候，传递了一个 `Repo` interface。
这个是我们刚刚说的解耦的好处，而且也能帮助我们更好地去做单元测试。
我们可以通过创建一个实现了 `Repo` interface 的 mock 实例，然后作为 `New()` 的参数传递给 `Article`，
这样我们就可以绕过我们的数据库去对我们的逻辑进行单元测试。

### 如何暴露服务

设置方法不需要知道每个服务的接入点、如何初始化存储层或者其他的一些事项。
只需要将数据库连接和路由实例传递给 `Activate()` 方法，
然后 `transport` package 中的路由注册程序将其路由进行注册，就可以对外提供服务了：

```golang
package transport

import (
    "database/sql"
    "net/http"

    "github.com/gin-gonic/gin"

    "github.com/kott/go-service-example/pkg/services/articles"
    "github.com/kott/go-service-example/pkg/services/articles/store"
)

type handler struct {
    ArticleService articles.Service
}

// Activate sets all the services required for articles and registers all the endpoints with the engine.
func Activate(router *gin.Engine, db *sql.DB) {
    articleService := articles.New(store.New(db))
    newHandler(router, articleService)
}

func newHandler(router *gin.Engine, as articles.Service) {
    h := handler{
        ArticleService: as,
    }
    router.GET("/articles/:id", h.Get)
    router.POST("/articles/", h.Create)
}

func (h *handler) Get(c *gin.Context) {...}

func (h *handler) Create(c *gin.Context) {...}
```

还记得我之前说的 `Start()` 方法（在 `pkg/api` 中）吗？
它是我们启动我们服务和配置的入口：

```golang
package api

import (
    "context"
    "fmt"

    "github.com/gin-gonic/gin"

    "github.com/kott/go-service-example/pkg/db"
    articles "github.com/kott/go-service-example/pkg/services/articles/transport"
    "github.com/kott/go-service-example/pkg/utils/log"
    "github.com/kott/go-service-example/pkg/utils/middleware"
)

// Config defines what the API requires to run
type Config struct {
    DBHost       string
    DBPort       int
    DBUser       string
    DBPassword   string
    DBName       string
    AppHost string
    AppPort int
}

// Start initializes the API server, adding the required middleware and dependent services
func Start(cfg *Config) {
    conn, err := db.GetConnection(
        cfg.DBHost,
        cfg.DBPort,
        cfg.DBUser,
        cfg.DBPassword,
        cfg.DBName)
    if err != nil {
        log.Error(ctx, "unable to establish a database connection: %s", err.Error())
    }
    defer func() {
        if conn != nil {
            conn.Close()
        }
    }()

    router := gin.New()
    router.Use(/* some middleware */)
    articles.Activate(router, conn)

    if err := router.Run(fmt.Sprintf("%s:%d", cfg.AppHost, cfg.AppPort)); err != nil {
        log.Fatal(context.Background(), err.Error())
    }
}
```

以上就是全部内容，实际上只是试图去找到一种抽象的方法，以便于让你的程序更易读，而不需要增加你项目的复杂度。

## 总结

有许许多多可以组织项目的方式，这个方式是我认为最好的。
根据功能将 service 分开有助于之后进行修改时定义上下文边界和代码导航。
将路由注册、业务逻辑、存储等放到同一个 `service` 层中，也让我们更关注业务逻辑本身和更容易地进行测试。
不要让简单的事情变复杂！如果您想节约时间，那就一定要这样做。

我希望能帮助到你，如果你想阅读所有源码，你可以在 [Github](https://github.com/kott/go-service-example) 下载。

---
via: https://medium.com/@ott.kristian/how-i-structure-services-in-go-19147ad0e6bd

作者：[Kristian Ott](https://medium.com/@ott.kristian)
译者：[h1z3y3](https://h1z3y3.me)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出