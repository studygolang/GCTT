首发于：https://studygolang.com/articles/17369

# Go 中的对象的生命周期

尽管 Go 语言很简单，Go 的开发人员仍然发现了许多创建和使用 Go 中对象的方法。在本篇博客中，我们将介绍对象管理的三步法 - ***实例化， 初始化，以及启动***。我们还会将其与其他的创建、使用对象的方法进行对比，并审查（评估）每种方法的优缺点。

## 我们的目标

这似乎是一个愚蠢的问题，但是，我们在 Go 中创建和使用对象的目的到底是什么？为了与 Go 的风格的统一，我优先考虑了以下事项：

* 足够简单
* 足够灵活
* 文档友好

除此之外，我们也应当说明哪些事情不是我们的目标。我们应该假设知道最终使用的用户的能力水平，所以我们就不需要提供过多的障碍。
使用我们代码的用户应该可以使用[RTFM](https://www.urbandictionary.com/define.php?term=RTFM)（假设我们提供了高质量的 “FM”）。
我们同样应该假设使用我们代码的用户不是恶意的 - 例如，我们不需要保护我们的对象字段因为我们认为开发者并不会恶意使用它们。

## 过程（三步法）

### 实例化

首先我们应该为我们的对象分配内存。Go 社区通常推荐的做法是创建一个该对象的[零值](https://golang.org/ref/spec#The_zero_value)。我发现这对于像 `sync.Mutex` 或 `bytes.Buffer` 这样的原始的受其 API 限制的结构来说是个很好的建议。

```go
var mu sync.Mutex
mu.Lock()
// do Things...
mu.Unlock()
```

但是，对于大多数的应用程序和开发者来说，构造函数可以提供更高的效率并防止未来可能出现的 bug。

#### 使用构造函数

Go 中的构造函数通常采用 `New` + 类型名称的形式。我们可以看下面这个 `Client` 的例子：

```go
// DefaultClientTimeout is the default Client.Timeout.
const DefaultClientTimeout = 30 * time.Seconds

// Client represents a client to our server.
type Client struct {
    Host    string
    Timeout time.Duration
}

// NewClient returns a new instance of Client with default settings.
func NewClient(host string) *Client {
    return &Client{
        Host:    host,
        Timeout: DefaultClientTimeout,
    }
}
```

通过使用构造函数，我们能得到一些好处。首先，我们无需每次使用时都去检查 `Timeout` 的零值，以确定是否应该使用其默认值。 因为它总是会被设置为一个正确的值。

其次，如果将来需要更改字段，我们也将提供无感知的升级体验。假设我们添加了一个需要在创建时需要初始化的缓存哈希表。

```go
type Client struct {
    cache map[string]interface{}

    Host    string
    Timeout time.Duration
}
```

如果我们在未来的版本中需要添加一个构造函数来初始化缓存，那么现存的所有使用零值的客户端都将被破坏。通过从一开始我们就将构造函数包含进来，并用文档记录其用法，那么就可以避免破坏未来的版本。

#### 使用自然的命名

使用构造函数的另一个好处是，由于零值的原因，我们的配置字段名称不再需要符合一定的标准。也就是说，如果我们有一个对象在默认情况下是“可编辑的”，那么我们不需要再创建一个名为 `NotEditable` 的布尔类型的字段来匹配默认的零值（`false`）。我们可以简单地使用自然名称：`Editalbe`，因为我们的构造函数会将其设置为 `true`。

### 初始化

一个对象完成内存分配和初始默认值分配之后，你需要根据你的用例来配置对象。在这个领域中，我发现大多数的 Go 开发人员都会想的过于复杂，但是实践中，它其实非常简单。

#### 请尽量只使用字段

通常来说，你应该只使用可导出的字段进行设置。在之前我们提到的 `Client` 对象示例中，我们提供了两个字段可供配置，`Host` 和 `Timeout`。

为了避免在并发情况下出现条件竞争，这些配置字段应该只配置一次，且单独保留，因为其他的函数例如 `Open()` 或 `Start()` 可能会启动其他的 Goroutines。我们可以在结构体文档上记录这些限制。

```go
type Client struct {
    // Host and port of remote server. Must be set before Open().
    Host string

    // Time until connection is cancelled. Must be set before Open().
    Timeout time.Duration
}
```

这个规则有个例外是，如果你在开始使用对象之后需要并发地更新这些字段的话。在这种情况下，我们应该 `setter` 和 `getter` 函数。

```go
type Client struct {
    mu      sync.Mutex
    timeout time.Duration

    // Host and port of remote server. Must be set before Open().
    Host string
}

// Timeout returns the duration until connection is cancelled.
func (c *Client) Timeout() time.Duration {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.timeout
}

// SetTimeout sets the duration until connection is cancelled.
func (c *Client) SetTimeout(d time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.timeout = d
}
```

然而，我发现在使用过程中更改配置设置通常是一种[气味代码 (code smell)](https://en.wikipedia.org/wiki/Code_smell)，通常来说应该避免。我们简单地停止对象并启动一个新的对象可能更加简洁。

### 启动

现在我们的对象已经完成内存分配，并且已经配置完毕 - 让我们来做一些更有用的事情。此时，一些简单的对象可能已经准备就绪了，但是对于一些复杂的对象（例如服务器）则需要启动之类的操作。它们可能需要连接到某些资源上或者启动后台 Goroutines 来监控资源，就像 `net.Listener` 对象那样。

在 Go 中，我们通常能看到 `Open()` 或者 `Start()` 这样的函数形式。我个人更倾向于选择 `Open()`，因为它的命名和 `io.Closer` 接口中的 `Close()` 方法更加般配。

在我们的 `Client` 示例中，我们会使用 `Open()` 函数来创建一个网络连接，使用 `Close()` 函数来关闭它。

```go
type Client struct {
    conn net.Conn

    // Host and port of remote server. Must be set before Open().
    Host string
}

// Open opens the connection to the remote server.
func (c *Client) Open() error {
    conn, err := net.Dial("tcp", c.Host)
    if err != nil {
        return err
    }
    c.conn = conn

    return nil
}

// Close disconnects the underlying connection to the server.
func (c *Client) Close() error {
    if c.conn != nil {
        return c.conn.Close()
    }
    return nil
}
```

在这个简单的例子中，我们需要注意两个地方。首先，我们的 `Host` 字段仅在 `Open()` 函数中使用了一次。这避免了在打开这个对象之后，其他线程对于这个字段的修改而产生的条件竞争。其次，我们无需尝试重置对象状态以重用对象。这些一次性对象避免了在尝试重用对象时出现的 bug。

#### 一次性使用对象

实践中，很难正确地清理复杂的对象并重新使用它们。在我们的例子中，我们不需要尝试在 `Close` 中将 `conn` 的状态设置成 `nil`。这是因为客户端可能会有一个后台 Goroutine 试图监控这个连接，并且会更改这个 `conn` 的值，这就要求我们需要给这个 `conn` 加上互斥锁以保护该字段。

我们也可以使用该字段来防止重复打开连接：

```go
// Open opens the connection to the remote server.
func (c *Client) Open() error {
    if c.conn != nil {
        return errors.New("myapp.Client: cannot reopen client")
    }
    ...
}
```

但是，我们应该假设开发者有这样的基本编码能力，并且通常应该避免这些过度的设计。

### 其他的方法

现在我们已经对 ***实例化 - 初始化 - 启动*** 方法有了大概的了解，那么让我们看看 Go 社区中一些其他的方法。

#### 备选 #1： 功能性的选项

*Dave Cheney* 在他的博客中描述了一种名为 ***functional options*** 的模式， [使用功能性的选项构造友好的 API](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)。该想法是我们可以声明一个功能性的参数类型来更新我们未导出的字段。然后我们可以在同一个调用中启动我们的对象，因为它已经初始化了。

同样地，使用我们在上面举例的 `Client`，它看起来像这样：

```go
type Client struct {
    host string
}

// OpenClient returns a new, opened client.
func OpenClient(opts ...ClientOption) (*Client, error) {
    c := &Client{}
    for _, opt := range opts {
        if err := opt(c); err != nil {
            return err
        }
    }
    // open client...
    return c, nil
}

// ClientOption represents an option to INItialize the Client.
type ClientOption func(*Client) error

// Host sets the host field of the client.
func Host(host string) ClientOption {
    return func(c *Client) error {
        c.host = host
        return nil
    }
}
```

我们的用法可以写成一行：

```go
client, err := OpenClient(Host("google.com"))
```

虽然这种方式隐藏了配置字段，但是它牺牲了可读性，增加了复杂性。*godoc* API 也会因为随着选项的增加而变得臃肿，且无法复用，乍一看这些文档，很难确定哪些选项适合哪些类型。

但实际上，我们不需要隐藏字段。我们应该记录它们的用法并信任开发人员能够正确地使用它们。保留这些可导出的字段会将所有的相关配置字段组合在一起，如 `net.Request` 类型那样。

#### 备选 #2：配置的实例化

另一种常见的做法是为你的对象类型提供“ config ”类型。这种尝试是将你的配置字段与你本身的类型分开。很多时候，开发人员会将配置对象中的字段拷贝到该类型中，或者直接将配置对象类型直接嵌入到该类型中。

继续使用 `Client` 作为例子来说明：

```go
type Client struct {
    host string
}

type ClientConfig struct {
    Host string
}

func NewClient(config ClientConfig) *Client {
    return &Client{
        host: config.Host,
    }
}
```

同样，这会隐藏 `Client` 类型中的配置字段，但却没带来其他任何好处。相反，我们应该简单地公开我们的 `Client.Host` 字段，让我们的用户直接管理它。这将降低我们 API 的复杂性，并且能够提供更加简洁的文档。

##### 何时该使用配置对象

配置对象很有用，但是不应作为 API 调用者和 API 作者之间的接口。配置对象应该存在于最终用户和有的软件之间。

例如， 配置对象可以通过[YAML](http://yaml.org/) 文件的形式为你的代码提供一个可配置的接口。这些配置对象通常应该存在于 `main` 包中，因为你的二进制文件会充当最终用户和代码之间的转换层的角色。

```go
package main

func main() {
    config := NewConfig()
    if err := readConfig(path); err != nil {
        fmt.Fprintln(os.Stderr, "cannot read config file:", err)
        os.Exit(1)
    }

    client := NewClient()
    client.Host = config.Host
    if err := client.Open(); err != nil {
        fmt.Fprintln(os.Stderr, "cannot open client:", err)
        os.Exit(1)
    }

    // do stuff...
}

type Config struct {
    Host string `yaml:"host"`
}

func NewConfig() Config {
    return &Config{
        Host: "localhost:1234"
    }
}
```

## 结论

我们研究了一种管理 Go 对象生命周期的方法，该方法提供了间接、灵活以及文档友好的特性。首先，我们 ***实例化*** 对象以分配内存并设置默认值。接下来，我们通过自定义的可导出字段来 ***初始化*** 对象。最后，我们 ***启动***  启动可能会有后台 Goroutines 或连接的对象。

这种简单的三步法有助于构建开发人员能够轻松使用的代码，并且可以交由未来的开发人员维护。

---

via: https://middlemost.com/object-lifecycle/

作者：[Ben Johnson](https://twitter.com/benbjohnson)
译者：[barryz](https://github.com/barryz)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
