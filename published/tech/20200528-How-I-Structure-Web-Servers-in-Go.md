首发于：https://studygolang.com/articles/30254

# 我是如何在 Go 中构建 Web 服务的

从用了近十年的 C# 转到 Go 是一个有趣的旅程。有时，我陶醉于 Go 的[简洁](https://www.youtube.com/watch?v=rFejpH_tAHM)；也有些时候，当熟悉的 OOP （面向对象编程）[模式](https://en.wikipedia.org/wiki/Software_design_pattern)无法在 Go 代码中使用的时候会感到沮丧。幸运的是，我已经摸索出了一些写 HTTP 服务的模式，在我的团队中应用地很好。

当在公司项目上工作时，我倾向把可发现性放在最高的优先级上。这些应用会在接下来的 20 年运行在生产环境中，必须有众多的开发人员和网站可靠性工程师（可能是指运维）来进行热补丁，维护和调整工作。因此，我不指望这些模式能适合所有人。

> [Mat Ryer 的文章](https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years.html)是我使用 Go 试验 HTTP 服务的起点之一，也是这篇文章的灵感来源。

## 代码组成

### Broker

一个 `Broker` 结构是将不同的 service 包绑定到 HTTP 逻辑的胶合结构。没有包作用域结级别的变量被使用。依赖的接口得益于了 [Go 的组合](https://www.ardanlabs.com/blog/2015/09/composition-with-go.html)的特点被嵌入了进来。

```go
type Broker struct {
    auth.Client             // 从外部仓库导入的身份验证依赖（接口）
    service.Service         // 仓库的业务逻辑包（接口）

    cfg    Config           // 该 API 服务的配置
    router *mux.Router      // 该 API 服务的路由集
}
```

broker 可以使用[阻塞](https://stackoverflow.com/questions/2407589/what-does-the-term-blocking-mean-in-programming)函数 `New()` 来初始化，该函数校验配置，并且运行所有需要的前置检查。

```go
func New(cfg Config, port int) (*Broker, error) {
    r := &Broker{
        cfg: cfg,
    }

    ...

    r.auth.Client, err = auth.New(cfg.AuthConfig)
    if err != nil {
        return nil, fmt.Errorf("Unable to create new API broker: %w", err)
    }

    ...

    return r, nil
}
```

初始化后的 `Broker` 满足了暴露在外的 `Server` 接口，这些接口定义了所有的，被 route 和 中间件（middleware）使用的功能。`service` 包接口被嵌入，这些接口与 `Broker` 上嵌入的接口相匹配。

```go
type Server interface {
    PingDependencies(bool) error
    ValidateJWT(string) error

    service.Service
}
```

web 服务通过调用 `Start()` 函数来启动。路由绑定通过一个[闭包函数](https://gobyexample.com/closures)进行绑定，这种方式保证循环依赖不会破坏导入周期规则。

```go
func (bkr *Broker) Start(binder func(s Server, r *mux.Router)) {
    ...

    bkr.router = mux.NewRouter().StrictSlash(true)
    binder(bkr, bkr.router)

    ...

    if err := http.Serve(l, bkr.router); errors.Is(err, http.ErrServerClosed) {
        log.Warn().Err(err).Msg("Web server has shut down")
    } else {
        log.Fatal().Err(err).Msg("Web server has shut down unexpectedly")
    }
}
```

那些对故障排除（比如，[Kubernetes 探针](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/）0)）或者灾难恢复方案方面有用的函数，挂在 `Broker` 上。如果被 routes/middleware 使用的话，这些仅仅被添加到 `webserver.Server` 接口上。

```go
func (bkr *Broker) SetupDatabase() { ... }
func (bkr *Broker) PingDependencies(failFast bool)) { ... }
```

### 启动引导

整个应用的入口是一个 `main` 包。默认会启动 Web 服务。我们可以通过传入一些命令行参数来调用之前提到的故障排查功能，方便使用传入 `New()` 函数的，经过验证的配置来测试代理权限以及其他网络问题。我们所要做的只是登入运行着的 pod 然后像使用其他命令行工具一样使用它们。

```go
func main() {
    subCommand := flag.String("start", "", "start the webserver")

    ...

    srv := webserver.New(cfg, 80)

    switch strings.ToLower(subCommand) {
    case "ping":
        srv.PingDependencies(false)
    case "start":
        srv.Start(BindRoutes)
    default:
        fmt.Printf("Unrecognized command %q, exiting.", subCommand)
        os.Exit(1)
    }
}
```

HTTP 管道设置在 `BindRoutes()` 函数中完成，该函数通过 `ser.Start()` 注入到服务（server）中。

```go
func BindRoutes(srv webserver.Server, r *mux.Router) {
    r.Use(middleware.Metrics(), middleware.Authentication(srv))
    r.HandleFunc("/ping", routes.Ping()).Methods(http.MethodGet)

    ...

    r.HandleFunc("/makes/{makeID}/models/{modelID}", model.get(srv)).Methods(http.MethodGet)
}
```

### 中间件

中间件（Middleware）返回一个带有 handler 的函数，handler 用来构建需要的 `http.HandlerFunc`。这使得 `webserver.Server` 接口被注入，同时所有的安静检查只在启动时执行，而不是在所有路由调用的时候。

```go
func Authentication(srv webserver.Server) func(h http.Handler) http.Handler {
    if srv == nil || !srv.Client.IsValid() {
        log.Fatal().Msg("a nil dependency was passed to authentication middleware")
    }

    // additional setup logic
    ...

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := strings.TrimSpace(r.Header.Get("Authorization"))
            if err := srv.ValidateJWT(token); err != nil {
                ...
                w.WriteHeader(401)
                w.Write([]byte("Access Denied"))

                return
            }

            next.ServeHTTP(w, r)
        }
    }
}
```

### 路由

路由有着与中间件有着类似的套路——简单的设置，但是有着同样的收益。

```go
func GetLatest(srv webserver.Server) http.HandlerFunc {
    if srv == nil {
        log.Fatal().Msg("a nil dependency was passed to the `/makes/{makeID}/models/{modelID}` route")
    }

    // additional setup logic
    ...

    return func(w http.ResponseWriter, r *http.Request) {
        ...

        makeDTO, err := srv.Get
    }
}
```

## 目录结构

代码的目录结构对可发现性进行了*高度*优化。

```
├── app/
|   └── service-api/**
├── cmd/
|   └── service-tool-x/
├── internal/
|   └── service/
|       └── mock/
├── pkg/
|   ├── client/
|   └── dtos/
├── (.editorconfig, .gitattributes, .gitignore)
└── go.mod
```

- app/ 用于项目应用——这是新来的人了解代码倾向的切入点。
dd
  - ./service-api/ 是该仓库的微服务 API；所有的 HTTP 实现细节都在这里。
- cmd/ 是存放命令行应用的地方。
- internal/ 是不可以被该仓库以外的项目引入的一个[特殊目录](https://dave.cheney.net/2019/10/06/use-internal-packages-to-reduce-your-public-api-surface)。
  - ./service/ 是所有领域逻辑（domain logic）所在的地方；可以被 `service-api`，`service-tool-x`，以及任何未来直接访问这个目录可以带来收益的应用或者包所引入。
- pkg/ 用于存放鼓励被仓库以外的项目所引入的包。
  - ./client/ 是用于访问 `service-api` 的 client 库。其他团队可以使用而不是自己写一个 client，并且我们可以借助我们在 `cmd/` 里面的 CI/CD 工具来 “[dogfood it](https://en.wikipedia.org/wiki/Eating_your_own_dog_food)” （使用自己产品的意思）。
  - ./dtos/ 是存放项目的数据传输对象，不同包之间共享的数据且以 json 形式在线路上编码或传输的结构体定义。没有从其他仓库包导出的模块化的结构体。`/internal/service` 负责 这些 DTO （数据传输对象）和自己内部模型的相互映射，避免实现细节的遗漏（如，数据库注释）并且该模型的改变不破坏下游客户端消费这些 DTO。
- .editorconfig，.gitattributes，.gitignore 因为[所有的仓库必须使用 .editorconfig，.gitattributes，.gitignore](https://www.dudley.codes/posts/2020.02.16-git-lost-in-translation/)！
- go.mod 甚至可以在[有限制的且官僚的公司环境](https://www.dudley.codes/posts/2020.04.02-golang-behind-corporate-firewall/)工作。

> 最重要的：每个包只负责意见事情，一件事情！

### HTTP 服务结构

```
└── service-api/
    ├── cfg/
    ├── middleware/
    ├── routes/
    |   ├── makes/
    |   |   └── models/**
    |   ├── create.go
    |   ├── create_test.go
    |   ├── get.go
    |   └── get_test.go
    ├── webserver/
    ├── main.go
    └── routebinds.go
```

- ./cfg/ 用于存放配置文件，通常是以 JSON 或者 YAML 形式保存的纯文本文件，它们也应该被检入到 Git 里面（除了密码，秘钥等）。
- ./middleware 用于所有的中间件。
- ./routes 采用类似应用的类 RESTFul 形式的目录对路由代码进行分组和嵌套。
- ./webserver 保存所有共享的 HTTP 结构和接口（Broker，配置，`Server` 等等）。
- main.go 启动应用程序的地方（`New()`，`Start()`）。
- routebinds.go `BindRoutes()` 函数存放的地方。

## 你觉得呢？

如果你最终采用了这种模式，或者有其他的想法我们可以讨论，我乐意听到这些想法！

---
via: https://www.dudley.codes/posts/2020.05.19-golang-structure-web-servers/

作者：[James Dudley](https://www.dudley.codes/)
译者：[dust347](https://github.com/dust347)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
