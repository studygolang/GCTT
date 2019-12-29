首发于：https://studygolang.com/articles/25564

# Go 语言中的集成测试：第二部分 - 设计和编写测试

## 序幕

这篇文章是集成测试系列两个部分中的第二部分。你可以先读 [第一部分：使用 Docker 在有限的环境中执行测试](https://studygolang.com/articles/21759)。本文中的示例可以从 [代码仓库](https://github.com/george-e-shaw-iv/integration-tests-example) 获取。

## 简介

> “比起测试行为，设计测试行为是已知的最好的错误预防程序之一。” —— Boris Beizer

在执行集成测试之前，必须正确配置该测试相关的外部系统。否则，测试结果是无效和不可靠的。例如，数据库需要有定义好的数据，这些数据对于要测试的行为是正确的。测试期间更改的数据需要进行验证，尤其是如果要求更改的数据对于后续测试而言是准确的时侯。

Go 测试工具提供了有在执行测试函数前执行代码的能力，使用叫做 `TestMain` 的入口函数实现。它类似于 Go 应用程序的 `Main` 函数。有了 `TestMain` 函数，我们可以在执行测试之前做其他系统配置，比如数据库连接之类的。在本文中，我将分享如何使用它 `TestMain` 来配置和连接 Postgres 数据库，以及如何针对该数据库编写和运行测试。

## 填充初始数据

为了填充数据库，需要定义数据并将其放置在测试工具可以访问的位置。一种常见的方法是定义一个 SQL 文件，该文件是项目的一部分，并且包含所有需要执行的 SQL 命令。另一种方法是将 SQL 命令存储在代码内部的常量中。不同于这两种方法，我将只使用 Go 语言实现来解决此问题。

通常情况下，你已将你的数据结构定义为 Go 结构体类型，用于数据库通信。我将利用这些已存在的数据结构，已经可以控制数据从数据库中流入流出。基于已有的数据结构声明变量，构造所有填充数据，而无需 SQL 语句。

我喜欢这种解决方式，因为它简化了编写集成测试和验证数据是否能够正确用于数据库和应用程序之间的通信的。不必将数据直接与 JSON 比较，就可以将数据解编为适当的类型，然后直接与为之前数据结构定义的变量进行比较。这不仅可以最大程度地减少测试中的语法比较错误，还可以使您的测试更具可维护性、可扩展性和可读性。

## 填充数据库

> 译者注：原文为 `Seeding The Database`，下面部分相关功能函数就称为种子函数

本文提到的，所有用于填充数数据库功能函数，都在 [`testdb`](https://github.com/george-e-shaw-iv/integration-tests-example) 包中。这个包仅用于测试，不用做第三方依赖。用来辅助填充测试数据库的三个核心函数分别是：`SeedLists`, `SeedItems`, 和 `Truncate`，如下：

这是 `SeedLists` 函数：

### 代码清单 1

```go
func SeedLists(dbc *sqlx.DB) ([]list.List, error) {
    now := time.Now().Truncate(time.Microsecond)

    lists := []list.List{
        {
            Name:     "Grocery",
            Created:  now,
            Modified: now,
        },
        {
            Name:     "To-do",
            Created:  now,
            Modified: now,
        },
        {
            Name:     "Employees",
            Created:  now,
            Modified: now,
        },
    }

    for i := range lists {
        stmt, err := dbc.Prepare("INSERT INTO list (name, created, modified) VALUES ($1, $2, $3) RETURNING list_id;")
        if err != nil {
            return nil, errors.Wrap(err, "prepare list insertion")
        }

        row := stmt.QueryRow(lists[i].Name, lists[i].Created, lists[i].Modified)

        if err = row.Scan(&lists[i].ID); err != nil {
            if err := stmt.Close(); err != nil {
                return nil, errors.Wrap(err, "close psql statement")
            }

            return nil, errors.Wrap(err, "capture list id")
        }

        if err := stmt.Close(); err != nil {
            return nil, errors.Wrap(err, "close psql statement")
        }
    }

    return lists, nil
}
```

代码清单 1 展示了 `SeedLists` 函数及其如何创建测试数据。`list.List` 定义了一个用于插入的数据表。然后，将测试数据插入数据库。为了帮助将插入的数据与测试期间进行的任何数据库调用的结果进行比较，测试数据集返回给调用方。

接下来，我们看看将更多测试数据插入数据库的 `SeedItems` 函数。

### 代码清单 2

```go
func SeedItems(dbc *sqlx.DB, lists []list.List) ([]item.Item, error) {
    now := time.Now().Truncate(time.Microsecond)

    items := []item.Item{
        {
            ListID:   lists[0].ID, // Grocery
            Name:     "Chocolate Milk",
            Quantity: 1,
            Created:  now,
            Modified: now,
        },
        {
            ListID:   lists[0].ID, // Grocery
            Name:     "Mac and Cheese",
            Quantity: 2,
            Created:  now,
            Modified: now,
        },
        {
            ListID:   lists[1].ID, // To-do
            Name:     "Write Integration Tests",
            Quantity: 1,
            Created:  now,
            Modified: now,
        },
    }

    for i := range items {
        stmt, err := dbc.Prepare("INSERT INTO item (list_id, name, quantity, created, modified) VALUES ($1, $2, $3, $4, $5) RETURNING item_id;")
        if err != nil {
            return nil, errors.Wrap(err, "prepare item insertion")
        }

        row := stmt.QueryRow(items[i].ListID, items[i].Name, items[i].Quantity, items[i].Created, items[i].Modified)

        if err = row.Scan(&items[i].ID); err != nil {
            if err := stmt.Close(); err != nil {
                return nil, errors.Wrap(err, "close psql statement")
            }

            return nil, errors.Wrap(err, "capture list id")
        }

        if err := stmt.Close(); err != nil {
            return nil, errors.Wrap(err, "close psql statement")
        }
    }

    return items, nil
}
```

代码清单 2 显示了 `SeedItems` 函数如何创建测试数据。除了使用 `item.Item` 数据类型，该代码与清单 1 基本相同。`testdb` 包中还有一个未提到的函数 `Truncate`。

### 代码清单 3

```go
func Truncate(dbc *sqlx.DB) error {
    stmt := "TRUNCATE TABLE list, item;"

    if _, err := dbc.Exec(stmt); err != nil {
        return errors.Wrap(err, "truncate test database tables")
    }

    return nil
}
```

代码清单 3 展示了 `Truncate` 函数。顾名思义，它用于删除 `SeedLists` 和 `SeedItems` 函数插入的所有数据。

## 使用 testing.M 创建 TestMain

使用便于 ` 填充/清除 ` 数据库的软件包后，该集中精力配置以运行真正的集成测试了。Go 自带的测试工具可以让你在 `TestMain` 函数中定义需要的行为，在测试函数执行前执行。

### 代码清单 4

```go
func TestMain(m *testing.M) {
    os.Exit(testMain(m))
}
```

代码清单 4 是 `TestMain` 函数，它在所有集成测试之前执行。在 23 行，叫做 `testMain` 的未导出的函数被 `os.Exit` 调用。这样做是为了 `testMain` 可以执行其中的延迟函数，并且仍可以在 `os.Exit` 调用内部设置适当的整数值。以下是 `testMain` 函数的实现。

### 代码清单 5

```go
func testMain(m *testing.M) int {
    dbc, err := testdb.Open()
    if err != nil {
        log.WithError(err).Info("create test database connection")
        return 1
    }
    defer dbc.Close()

    a = handlers.NewApplication(dbc)

    return m.Run()
}
```

在代码清单 5 中，你可以看到 `testMain` 只有 8 行代码。28 行，函数调用 `testdb.Open()` 开始建立数据库连接。此调用的配置参数在 `testdb` 包中设置为常量。重要的是要注意，如果测试用的数据库未运行，调用 `Opne` 连接数据库会失败。该测试数据库是由 `docker-compose` 创建提供的，详细说明在本系列的第 1 部分中（单击 [这里](https://studygolang.com/articles/21759) 阅读第 1 部分）。

成功连接测试数据库后，连接将传递给 `handlers.NewApplication()`，并且此函数的返回值用于初始化的包级变量 `*handlers.Application` 类型。`handlers.Application` 类型是这个项目自定义的结构体，有用于 `http.Handler` 接口的字段，以简化 Web 服务的路由以及对已创建的数据库连接的引用。

现在，应用程序值已初始化，可以调用 `m.Run` 来执行所有测试函数。对 `m.Run` 的调用处于阻塞状态，直到所有确定要运行的测试函数都执行完之后，该调用才会返回。非零退出代码表示失败，0 表示成功。

## 编写 Web 服务的集成测试

集成测试将多个代码单元以及所有集成服务（例如数据库）组合在一起，并测试各个单元的功能以及各个单元之间的关系。为 Web 服务编写集成测试通常意味着每个集成测试的所有入口点都是一个路由。`http.Handler` 接口是任何 Web 服务的必需组件，它包含的 `ServeHTTP` 函数使我们能够利用应用程序中定义的路由。

在 Web 服务的集成测试中，构建初始化数据并且以 Go 类型返回初始数据，对返回的响应体的结构进行断言非常有用。在接下来的代码清单中，我将一个典型的 API 路由集成测试分解成几个不同的部分。第一步是使用代码清单 1 和代码清单 2 中定义的种子数据。

### 清单 6

```go
func Test_getItems(t *testing.T) {
    defer func() {
        if err := testdb.Truncate(a.DB); err != nil {
            t.Errorf("error truncating test database tables: %v", err)
        }
    }()

    expectedLists, err := testdb.SeedLists(a.DB)
    if err != nil {
        t.Fatalf("error seeding lists: %v", err)
    }

    expectedItems, err := testdb.SeedItems(a.DB, expectedLists)
    if err != nil {
        t.Fatalf("error seeding items: %v", err)
    }
}
```

在获取种子数据失败前，必须设置延迟函数清理数据库，这样，无论函数失败与否，测试结束后保证数据库是干净的。然后，调用 `testdb` 中的种子函数（`testdb.SeedLists` 和 `testdb.SeedItems` ）构造初始数据，并获取他们的返回值作为预期值，以便在集成测试中与实际路由请求结果（真实值）做对比。如果这两个种子函数中的任何一个失败，测试就会调用 `t.Fatalf` 。

### 清单 7

```go
// Application is the struct that contains the server handler as well as
// any references to services that the application needs.
type Application struct {
    DB      *sqlx.DB
    handler http.Handler
}

// ServeHTTP implements the http.Handler interface for the Application type.
func (a *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    a.handler.ServeHTTP(w, r)
}

```

为了调用注册的路由，`Application` 类型实现 `http.Handler` 接口。`http.Handler` 作为 `Application` 的内嵌结构体字段，因此 `Application` 可以调用 `http.Handler` 接口实现的 `ServeHTTP` 函数

### 清单 8

```go
req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/list/%d/item", test.ListID), nil)
if err != nil {
   t.Errorf("error creating request: %v", err)
}

w := httptest.NewRecorder()
a.ServeHTTP(w, req)
```

回顾一下代码清单 5，构造 `Application` 是为了在测试中使用。`ServeHTTP` 函数需要两个参数： `http.ResponseWriter`  和 `http.Request`。`http.NewRequest` 构造 `http.Request`，`httptest.NewRecorder` 构造 `http.ResponseRecorder`——即 `http.Response` 。

 `http.NewRecorder` 函数的返回 `ResponseRecorder` 值实现了 `ResponseWriter` 接口。调用路由请求后，`ResponseRecorder` 可以用来分析了。其中最关键的字段 `Code` 和 `Body`，前者是该请求的实际响应码，后者是一个指向响应内容的 `bytes.Buffer` 类型的指针。

> 译者注：这里的 `http.ResponseWriter`  和 `http.Request` 实现了 Golang 中常见的 `Writer` 和 `Reader` 接口，即 **输出** 和 **输入**，在 http 请求中即 `Response` 和 `Request`。

### 清单 9

```go
if want, got := http.StatusOK, w.Code; want != got {
    t.Errorf("expected status code: %v, got status code: %v", want, got)
}
```

清单 9 中，实际的响应码和预期的响应码做对比。如果不同，将调用 `t.Errorf`，它将输出失败原因。

### 清单 10

```go
var items []item.Item
resp := web.Response{
    Results: items,
}

if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
    t.Errorf("error decoding response body: %v", err)
}

if d := cmp.Diff(expectedItems, items); d != "" {
    t.Errorf("unexpected difference in response body:\n%v", d)
}
```

示例中使用自定义响应体 `web.Response`，使用 键为 `results` 的 JSON 字符串存储路由返回信息。代码清单 10 中声明了一个 []item.Item 类型的变量 items，用于和预期值对比。 初始化 items 变量传递给 resp 的字段 results。接下来，items 会随着解析路由响应体数据到 resp 中，从而包含响应体的数据。

Google 的 [go-cmp](https://github.com/google/go-cmp) 包可替代 `reflect.DeepEqual` ，在对比 struct,map,slice 和 array 时更安全，更易用。调用 cmp.Diff 对比清单 6 中定义的种子数据和实际响应体中返回的数据，如果不等，测试将失败，并且将差异输出到标准输出（stdout）中。

## 测试技巧

就测试而言，最好的建议是尽早测试，并且经常测试，而不是将测试放到开发之后考虑，而且测试应该推动、驱动应用程序的开发。这就是“测试驱动开发（TDD）”。通常情况下，没有随时测试代码。在编写代码时，将测试的想法抛到脑后，自己（开发人员）默认编写的代码是可测试的。代码单元（通常是一个函数）不管再小都能进行测试。你的服务进行越多测试，未知的就越少，隐藏的副作用（bug）就越少。

有了下面这些技巧，你的测试将洞察力，更易读，更快。

### 表测试

表测试是一种编写测试的方式，可以防止针对同一代码单元的不同可测试结果重复测试断言。以下面的求和函数为例：

### 清单 11

```go
// Add takes an indefinite amount of operands and adds them together, returning
// the sum of the operation.
func Add(operands ...int) int {
    var sum int

    for _, operand := range operands {
        sum += operand
    }

    return sum
}
```

在测试中，我想确保函数可以处理以下情况：

* 没有参数（operands），应返回 0。
* 一个参数，直接返回参数值。
* 两个参数，返回这两个数之和。
* 三个参数，则返回这三个数之和。

彼此独立地编写这些测试将导致重复许多相同的调用和断言。我认为，更好的方法是利用表测试。为了编写表测试，必须定义一片匿名声明的结构，其中包含我们每个测试用例的元数据。然后可以使用循环遍历不同测试用例的这些条目，并可以对用例进行测试和独立运行 `t.Run`。`t.Run` 需要两个参数，子测试函数和这个子测试函数的函数名，子测试函数必须符合这种类型：`func(*testing.T)`。

### 清单 12

```go
// TestAdd tests the Add function.
func TestAdd(t *testing.T) {
    tt := []struct {
        Name     string
        Operands []int
        Sum      int
    }{
        {
            Name:     "NoOperands",
            Operands: []int{},
            Sum:      0,
        },
        {
            Name:     "OneOperand",
            Operands: []int{10},
            Sum:      10,
        },
        {
            Name:     "TwoOperands",
            Operands: []int{10, 5},
            Sum:      15,
        },
        {
            Name:     "ThreeOperands",
            Operands: []int{10, 5, 4},
            Sum:      19,
        },
    }

    for _, test := range tt {
        fn := func(t *testing.T) {
            if e, a := test.Sum, Add(test.Operands...); e != a {
                t.Errorf("expected sum %d, got sum %d", e, a)
            }
        }

        t.Run(test.Name, fn)
    }
}
```

测试清单 12 中，使用匿名声明的结构体定义了不同的情况。遍历这些情况，执行这些测试用例。比较实际返回值和预期值，如果不等，则调用 `t.Errorf`，返回测试失败的信息。清单中，遍历调用 t.Run 执行每个测试用例。

### t.Helper() 和 t.Parallel()

标准库中的 `testing` 包提供了很多有用的程序（函数）辅助测试，而不用导入之外的第三方包。其中我最喜欢的两个函数是 `t.Helper()` 和 `t.Parallel()`，它们都定义为 `testing.T` 接收者，它是在 `_test.go` 文件中每个 `Test` 函数都必需的一个的参数。

### 清单 13

```go
// GenerateTempFile generates a temp file and returns the reference to
// the underlying os.File and an error.
func GenerateTempFile() (*os.File, error) {
    f, err := ioutil.TempFile("", "")
    if err != nil {
        return nil, err
    }

    return f, nil
}
```

在代码清单 13 中，为特定的测试包定义了一个辅助函数。这个函数返回 `os.File` 指针和 `error`。每次测试调用这个辅助函数必须判断 error 是一个 non-nil 。通常情况这也没什么，但是有一个更好的方式：使用 t.Helper() ，这种方式省略了 `error` 返回。

### 清单 14

```go
// GenerateTempFile generates a temp file and returns the reference to
// the underlying os.File.
func GenerateTempFile(t *testing.T) *os.File {
    t.Helper()

    f, err := ioutil.TempFile("", "")
    if err != nil {
        t.Fatalf("unable to generate temp file: %v", err)
    }

    return f
}
```

清单 14 和清单 13 相同，只是使用 `t.Helper()`。这个函数定义使用了 `*testing.T` 作为参数，省略了 error 的返回。函数先调用 `t.Helper()`，这在编译测试二进制文件时发出信号：如果 t 在这个函数中调用任何接收器函数，则将其报告给调用函数（Test*）。与辅助函数不同，所有行号和文件信息会都会关联到这个函数。

一些测试可以进行安全的并行进行，并且 Go testing 包原生支持并行运行测试。在所有 Test* 函数开始调用 t.Parallel(), 可以编译出可以安全并行运行的测试二进制文件。就是这么简单，就是这么强大！

## 结论

如果不配置程序运行时所需的外部系统，则无法在集成测试的上下文中完全验证程序的行为。此外，需要持续监测那些外部系统（特别是当它们包含应用程序状态数据的情况下），以确保它们包含有效和有意义的数据。Go 使开发人员不仅可以在测试过程中进行配置，还可以无需标准库之外的包就能维护外部数据。因此，我们可以编写可读性，一致性，性能和可靠性同时都能保证的集成测试。Go 的真正魅力正在于其简约而功能齐全的工具集，它为开发人员提供了无需依赖外部库或任何非常规限制的功能。

---

via: https://www.ardanlabs.com/blog/2019/10/integration-testing-in-go-set-up-and-writing-tests.html

作者：[George Shaw](https://github.com/george-e-shaw-iv/)
译者：[TomatoAres](https://github.com/TomatoAres)
校对：[lxbwolf](https://github.com/lxbwolf)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
