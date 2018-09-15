首发于：https://studygolang.com/articles/14851

# Go 的依赖注入

过去几年里我一直使用 Java。最近，用 Go 建立了一个小项目，然而 Go 生态系统中依赖注入（DI）功能缺乏让我震惊。于是我决定尝试使用 Uber 的 [dig](https://github.com/uber-go/dig) 库来构建我的项目，期间感触颇深。

我发现 DI 帮助我解决了之前在 Go 应用程序中遇到的很多问题 - 过度使用 `init` 函数，滥用全局变量和复杂的应用程序设置等。

在这篇文章中，我将介绍 DI ，然后在使用 DI 框架（通过 `dig` 库）前后写一些例子做对比。

## DI 的简要概述

依赖注入是指你的组件（通常在 Go 中是 struct ）在创建时，就应该获取它们依赖关系的一种思想。这与那些组件在初始化过程中，就建立自身依赖关系的反关联模式不同 。我们来看一个例子。

假设你构造 `Server` 需要 `Config` 结构体。一种方法是在初始化期间 `Server` 构建 `Config` 。

```go
type Server struct {
	config *Config
}

func New() *Server {
	return &Server{
		config: buildMyConfigSomehow(),
	}
}
```

看起来很方便。调用者甚至不必知道 `Server` 需要访问 `Config` 。这些都被我们的函数隐藏起来了。

然而，这存在一些缺点。首先，如果我们想要改变我们 `Config` 的构建方式，我们不得不改变所有调用构建代码的地方。例如，假设我们的 `buildMyConfigSomehow` 函数现在需要一个参数。每个调用处都需要访问该参数并需要将其传递给构造函数。

此外，这使得实现 `Config` 函数变得十分麻烦，我们得以某种方法进入 `new` 函数的内部，并创建`Config`。

这是 DI 方式：

```go
type Server struct {
	config *Config
}

func New(config *Config) *Server {
	return &Server{
		config: config,
	}
}
```

现在我们将 `Server` 与`Config` 分离。我们可以根据自己的逻辑创造 `Config` 然后将结果传递给 `New` 函数。

此外，如果 `Config` 是一个接口，这为我们提供了一个简单的模拟途径 。只要 `New` 实现了我们的接口，就可以传递任何我们想要的东西。这使得测试实现了 `Config` 接口的 `Server` 很简单。

令人痛苦的是在创建 `server` 之前手动创建 `config` 。我们在这里创建了一个依赖关系 –  因为 `server` 依赖 `Config,` 所以需要首先创建 `Config`  。在真正的应用程序中，这些依赖会变得更加复杂，这会导致构建应用程序完成其工作所需的组件间的复杂逻辑 。

这是 DI 框架可以提供帮助的地方。 DI 框架通常提供两个功能：

1. “提供”新组件。简而言之，这告诉 DI 框架一旦你有这些组件，还需要其他什么组件（依赖关系）以及如何去构建。
2. “检索”构建组件。

DI 框架通常基于您告诉它的 “providers” 构建依赖图并确定如何构建对象。这在没有具体例子的情况下很难理解，所以让我们来看一个中等大小的例子。

## 示例程序

我们来看http服务器端的代码：客户端以 `GET` 方式请求 `/people` 路径时并返回 JSON 。我们将一步一步呈现代码，为简单起见，它们都存在于同一个包中（`main`）。请勿在真正的 Go 程序中执行此操作。可以在[此处](https://gitlab.com/drewolson/go_di_example)找到此示例的完整代码。

首先，让我们看看我们的 `Person` 。仅有一些被 JSON 标签标记的属性。

```go
type Person struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}
```

`Person` 有 `Id`，`Name` 和 `Age` 。

接下来让我们看看 `Config` 。与 `Person` 类似，它没有依赖关系。与 `Person` 不同的是，我们将提供构造函数。

```go
type Config struct {
	Enabled      bool
	DatabasePath string
	Port         string
}

func NewConfig() *Config {
	return &Config{
		Enabled:      true,
		DatabasePath: "./example.db",
		Port:         "8000",
	}
}
```

 `Enabled` 表示程序是否返回真实数据。`DatabasePath` 表示数据库的地址（使用 sqlite ）。`Port` 表示服务器运行的端口。

下方函数用来打开数据库连接。它依赖于 `Config` 并返回 `*sql.DB` 。

接下来看看 `PersonRepository`。此结构负责从数据库中提取数据并反序列化为 `Person` 。

```go
type PersonRepository struct {
	database *sql.DB
}

func (repository *PersonRepository) FindAll() []*Person {
	rows, _ := repository.database.Query(
		`SELECT id, name, age FROM people;`
	)
	defer rows.Close()

	people := []*Person{}

	for rows.Next() {
		var (
			id   int
			name string
			age  int
		)

		rows.Scan(&id, &name, &age)

		people = append(people, &Person{
			Id:   id,
			Name: name,
			Age:  age,
		})
	}

	return people
}

func NewPersonRepository(database *sql.DB) *PersonRepository {
	return &PersonRepository{database: database}
}
```

`PersonRepository` 的构建需要数据库连接。它有一个函数 `FindAll`，此函数使用数据库连接信息并返回 `Person` 列表。

要在 HTTP 服务器和 `PersonRepository` 之间提供一层，我们需要创建 `PersonService` 。

```go
type PersonService struct {
	config     *Config
	repository *PersonRepository
}

func (service *PersonService) FindAll() []*Person {
	if service.config.Enabled {
		return service.repository.FindAll()
	}

	return []*Person{}
}

func NewPersonService(config *Config, repository *PersonRepository)
*PersonService {
	return &PersonService{config: config, repository: repository}
}
```

我们的 `PersonService` 依赖于 `Config` 和 `PersonRepository` 。它有一个函数 `FindAll` ，如果启用了应用程序，则会有条件地调用 `PersonRepository` 。

最后，我们得到了 `Server` 。负责运行 HTTP 服务器并将适当的请求委托给 `PersonService` 。

```go
type Server struct {
	config        *Config
	personService *PersonService
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/people", s.people)

	return mux
}

func (s *Server) Run() {
	httpServer := &http.Server{
		Addr:    ":" + s.config.Port,
		Handler: s.Handler(),
	}

	httpServer.ListenAndServe()
}

func (s *Server) people(w http.ResponseWriter, r *http.Request) {
	people := s.personService.FindAll()
	bytes, _ := json.Marshal(people)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func NewServer(config *Config, service *PersonService) *Server {
	return &Server{
		config:        config,
		personService: service,
	}
}
```

`Server` 取决于 `PersonService` 和 `Config` 。

好的，我们了解了系统的所有组件。现在我们该如何在实际中初始化它们并启动我们的系统？

## 传统的 main()

首先，让我们用传统方式编写 `main()` 。

```go
func main() {
	config := NewConfig()

	db, err := ConnectDatabase(config)

	if err != nil {
		panic(err)
	}

	personRepository := NewPersonRepository(db)

	personService := NewPersonService(config, personRepository)

	server := NewServer(config, personService)

	server.Run()
}
```
首先，我们创建 `Config` 。然后使用 `Config` 创建数据库连接。从而创建 `PersonRepository` 和 `PersonService` 。最后，再创建 `Server` 并运行它。

这有些复杂。更糟糕的是，随着我们的应用程序的变得复杂，`main` 的复杂性也将继续增长。每次我们向任何组件添加新的依赖时，都必须通过 `main` 函数中的排序和逻辑来反映该依赖，以构建该组件。

您可能已经猜到，依赖注入框架可以帮助我们解决这个问题。一起来看看。

## 创建容器

术语“ 容器（container） ”通常用在 DI 框架中，用于描述添加“提供者（providers）”的内容，并从中请求构建对象。`dig` 库用 `Provide` 函数为我们添加 “providers”， `Invoke` 函数用于从容器中检索全部的构建对象。

首先，我们构建一个新容器。

```go
container := dig.New()
```

现在我们可以添加新的提供者。为此，我们在容器上调用 `Provide` 函数。它只需要一个参数：一个函数。此函数可以包含任意数量的参数（表示要创建的组件的依赖关系）和一个或两个返回值（表示函数提供的组件以及可选的错误）。

```go
container.Provide(func() *Config {
	return NewConfig()
})
```

上面的代码说“我为容器提供了一种 `Config` 类型。为了构建它，我不需要任何其他东西。“现在我们已经向容器展示了如何构建 `Config` 类型，继续使用它来构建其他类型。

```go
container.Provide(func(config *Config) (*sql.DB, error) {
	return ConnectDatabase(config)
})
```

这段代码说“我为容器提供了一种 `*sql.DB` 类型。为了构建它，我需要一个 `Config` 。可以选择返回错误。“

在这两种情况下，我们没必要这样写。因为我们已经有了 `NewConfig` 和 `ConnectDatabase` 函数，我们可以直接使用他们作为容器的提供者。

```go
container.Provide(NewConfig)
container.Provide(ConnectDatabase)
```

现在，我们可以从之前给容器提供的类型中创建组件。我们使用 `Invoke` 函数，函数采用单个参数 - 具有任意数量参数的函数。函数的参数是我们希望容器构建的类型。

```go
container.Invoke(func(database *sql.DB) {
	// sql.DB is ready to use here
})
```

容器做了一些非常聪明的东西，如下：

- 容器认识到我们要求的是构建 `*sql.DB`
- 它确定函数 `ConnectDatabase` 提供该类型
- 接下来它确定 `ConnectDatabase` 函数依赖 `Config`
- 它找到了 `Config` 的提供者，也就是 `NewConfig`
- `NewConfig` 没有任何依赖关系，所以它被调用
- `NewConfig` 的结果是一个 `Config` 传递给 `ConnectDatabase`
- `ConnectionDatabase` 的结果是 `*sql.DB` 被传递给 `Invoke`

这是容器为我们做的很多工作。事实上，它做的更多。容器很智能，可以构建每种类型有且仅有一个实例。这意味着如果我们在多个地方（比如多个存储库）使用它，我们永远不会意外地创建第二个数据库连接。

## 较好的 main() 写法

现在知道了 `dig` 容器是如何工作的，让我们用它来构建一个较好的 main 。

```go
func BuildContainer() *dig.Container {
	container := dig.New()

	container.Provide(NewConfig)
	container.Provide(ConnectDatabase)
	container.Provide(NewPersonRepository)
	container.Provide(NewPersonService)
	container.Provide(NewServer)

	return container
}

func main() {
	container := BuildContainer()

	err := container.Invoke(func(server *Server) {
		server.Run()
	})

	if err != nil {
		panic(err)
	}
}
```

之前唯一没见过的就是 `Invoke` 的返回值 `error` 。如果任何提供者使用 `Invoke` 返回错误，我们调用 `Invoke` 将停止并返回该错误。

虽然这个例子很小，但应该很容易看出这种方法的一些好处超过了“常规“的 main 。随着应用程序变得越来越大，这些好处变得更加明显。

最重要的好处之一是将组件的创建与其依赖的创建分离。比如说，我们  `PersonRepository` 现在需要访问 `Config` 。我们所要做的就是更改 `NewPersonRepository` 构造函数以包含 `Config` 作为参数。代码其他任何内容没有发生改变。

其他的好处是没有全局状态，没有调用 `init` （依赖关系在需要时才创建，只创建一次，不需要容易出错的 `init` 设置），并且易于测试单个组件。想象一下，在测试中创建容器并要求完整构建对象进行测试。或者，创建一个对象需要所有的依赖。使用 DI ，这些都更容易。

## 一个值得传播的想法

我相信依赖注入有助于构建更强大和可测试的应用程序。随着这些应用程序体量逐渐增大，尤为明显。 Go 非常适合构建大型应用程序，并且具有很好的 DI 工具 `dig` 。我相信 Go 社区应该接受 DI 并在更多的应用程序中使用它。

---

via: https://blog.drewolson.org/dependency-injection-in-go/

作者：[Drew Olson](https://blog.drewolson.org/author/drew/)
译者：[NoSugarCoffee](https://github.com/NoSugarCoffee)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
