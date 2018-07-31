# 测试驱动学习GO语言: 同步异步过程测试
本文源自WIP项目中一个名为“通过实例学习go语言”的系列中的第八篇文章，旨在围绕TDD（测试驱动开发）来接触go语言及学习相关的技术。
- 第一篇文章快速熟悉TDD
- 第二篇文章讨论数组及切片
- 第三篇文章讨论类、方法、接口还有表驱动测试
- 第四篇文章展示如何处理异常及指针为什么有用
- 第五篇文章解释为什么及如何使用依赖注入
- 第六篇文章介绍并发相关内容
- 第七篇文章讲解使用mock工具的优势及方法

本章是关于基于select 的同步异步过程的介绍

## Select
[本章节中所有源码在此处](https://github.com/quii/learn-go-with-tests/tree/master/select)

首先我们要编写一个名为 `WebsiteRacer` 的函数，该函数将通过http GET方法对两个url的的请求返回时间进行一个“赛跑”，同时如果两个请求在10秒内没有响应，即返回错误信息。

为此我们将使用：
- net/http 发起HTTP请求
- net/http/httptest 帮助我们对其测试
- goroutines
- select 同步过程

## 首先编写测试代码
让我们从最基本的开始运行
```
func TestRacer(t *testing.T) {
    slowURL := "http://www.facebook.com"
    fastURL := "http://www.quii.co.uk"

    want := fastURL
    got := Racer(slowURL, fastURL)

    if got != want{
        t.Errorf("got '%s', want '%s'", got, want)
    }
}
```

我们很清楚，上边代码并不是很完美，但足以让其运转起来。要记住，不要一开始就想着很完美的去实现它，这一点很重要。
## 接下来试着运行这个测试
`./racer_test.go:14:9: undefined: Racer`

### 编写最小量的代码进行测试并检查错误测试的输出
```
func Racer(a, b string) (winner string) {
    return
}
```

`racer_test.go:25: got '', want 'http://www.quii.co.uk'`

### 进一步编写代码进行完善
```
func Racer(a, b string) (winner string) {
    startA := time.Now()
    http.Get(a)
    aDuration := time.Since(startA)

    startB := time.Now()
    http.Get(b)
    bDuration := time.Since(startB)

    if aDuration < bDuration {
        return a
    }

    return b
}
```

对于每一个url
- 1、我们使用 `time.Now()` 在我们尝试请求url之前进行记录。
-  2、接下来，我们使用[http.Get]('https://golang.org/pkg/net/http/#Client.Get')尝试请求获取url的内容。这个方法将返回一个[http.Response]('https://golang.org/pkg/net/http/#Response"') 及一个错误值，但是目前为止我们不需要关心这个值。
- 3、`time.Since` 需要一个起始时间，并且会返回一个`time.Duration`的时间差。

一旦我们完成了这一步，我们只需比较返回的时间，看哪个最快

### 疑问
这个测试也许并不适用于你，因为我们使用的是真实的网站进行测试自己的逻辑。
对于使用HTTP请求这种很常用的测试，Go有一个标准库工具可以帮助我们测试。
在mocking和依赖注入章节中，我们介绍了在理想情况下，我们不需要依赖外部服务测试我们的代码，因为它们经常有以下问题：
- 慢
- 不规则
- 没有边缘值
在标准库中，有一个包 [net/http/httptest]('https://golang.org/pkg/net/http/httptest/') 可以更简单的创建mock HTTP服务。
让我们修改一下我们的测试代码，通过使用mock创建一个可信的服务，这样我们的测试更加可控。
```
func TestRacer(t *testing.T) {

    slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(20 * time.Millisecond)
        w.WriteHeader(http.StatusOK)
    }))

    fastServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    slowURL := slowServer.URL
    fastURL := fastServer.URL

    want := fastURL
    got := Racer(slowURL, fastURL)

    if got != want {
        t.Errorf("got '%s', want '%s'", got, want)
    }

    slowServer.Close()
    fastServer.Close()
}
```

语法可能看上去有些乱，但根据你的时间进行调整。
`httptest.NewServer` 需要一个`http.HandlerFunc` 方法，这个方法我们使用匿名函数进行发送。

`http.HandlerFunc` 是一种类型类似于: `type HandlerFunc func(ResponseWriter, * Request)`。
这一切都指明需要一个 `ResponseWriter` 和一个 `Request` 方法，这些在HTTP服务中并不稀奇。
事实证明，没有什么特别神奇的地方，这也是你如何使用go语言编写一个真正的HTTP服务。唯一的区别是我们将其包装成了`httptest.NewServer`，这也使得测试更加容易，因为它找到了一个开放的端口来监听请求，当你完成测试后可以将其关闭。

在我们的两个服务中，我们使其中一个慢的休眠一小段时间，当我们得到返回信息的时候让它比另一个相比慢一点。两个服务后边都会返回给请求者一个OK的响应，使用`w.WriteHeader(http.StatusOK)`。

当你重新运行测试的时候，肯定会更加快速的通过。使用休眠故意去打破测试。

### 重构
在实际测试和生产的代码中，我们经常会有一些重复的代码。
```
func Racer(a, b string) (winner string) {
    aDuration := measureResponseTime(a)
    bDuration := measureResponseTime(b)

    if aDuration < bDuration {
        return a
    }

    return b
}

func measureResponseTime(url string) time.Duration {
    start := time.Now()
    http.Get(url)
    return time.Since(start)
}
```

这样便使我们“赛跑”代码更加易读。
```
func TestRacer(t *testing.T) {

    slowServer := makeDelayedServer(20 * time.Millisecond)
    fastServer := makeDelayedServer(0 * time.Millisecond)

    defer slowServer.Close()
    defer fastServer.Close()

    slowURL := slowServer.URL
    fastURL := fastServer.URL

    want := fastURL
    got := Racer(slowURL, fastURL)

    if got != want {
        t.Errorf("got '%s', want '%s'", got, want)
    }
}

func makeDelayedServer(delay time.Duration) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(delay)
        w.WriteHeader(http.StatusOK)
    }))
}
```

接下来，我们重构了名为`makeDelayedServer` 的方法，用来构造我们的虚拟服务。并将一些多余的代码移除测试从而减少冗余。
### defer
通过在延迟函数前增加一个前置的函数调用， 它将包含在函数中的末尾进行调用。

有些时候，你可能需要清理资源，例如关闭文件读取或关闭某些服务以关闭端口监听。

虽然你希望在程序结束时去执行该操作，但应该尽量创建在服务的位置附近，这样代码会更加具有可读性。

我们的重构是一种改进，并且是迄今为止涵盖go属性的最合理的方案，但是我们可以使解决方案更加简单。

### 进程同步
- go是十分擅长处理并发的，那我们为什么还要一个接一个的测试网站响应速度呢？我们可以同时测试两者。

我们并不需要关心确切的响应时间，只需要知道哪一个是第一个返回的。

为了达到这个目的，我们将一个新的结构体叫 **select** ，它可以简单明了的帮助我们进行同步过程处理。

```
func Racer(a, b string) (winner string) {
    select {
    case <-ping(a):
        return a
    case <-ping(b):
        return b
    }
}

func ping(url string) chan bool {
    ch := make(chan bool)
    go func() {
        http.Get(url)
        ch <- true
    }()
    return ch
}
```
## ping
我们定义一个函数`ping` 用来创建一个`chan bool`并且返回它。

在我们这个案例中，我们并不真正关心通道中发出的信号类型，我们发送信号仅仅表示我们完成了而且状态良好。

在同样的方法里，我们开始一个go程，用来发送一个信号通道，代表我们完成了一个`http get`请求。
## select
如果你翻回“并发”那个章节中，你可以用`myVar := <-ch`发送一个信号进行等待回值。这是一个阻塞的请求，你可以用来等待返回值。

`select` 让你可以等待多个信道的消息。第一个选项得到值`成功`，该case下的代码将被执行。

我们使用  `ping` 在我们的 `select` 中为我们的 **URL** 创建两个信道。无论哪个首先写入到信道，都会执行相应的 `select`的代码，及会导致URL的返回（成为获胜的一方）。

在这些修改后，我们的意图在代码中已经非常明确，实际实现上更简单。
### Timeout
我们最后一个需求是实现当`Racer` 超过10秒未响应的时候返回一个错误提示。
### 编写第一个测试
```
t.Run("returns an error if a server doesn't respond within 10s", func(t *testing.T) {
    serverA := makeDelayedServer(11 * time.Second)
    serverB := makeDelayedServer(12 * time.Second)

    defer serverA.Close()
    defer serverB.Close()

    _, err := Racer(serverA.URL, serverB.URL)

    if err == nil {
        t.Error("expected an error but didn't get one")
    }
})
```

我们已经使我们的测试服务做了针对脚本的超过10秒的训练，并且我们预期的`Racer` 现在返回了两个值，“胜利”的URL（在此次测试忽略）和一个错误提示。

### 试着运行测试
`./racer_test.go:37:10: assignment mismatch: 2 variables but 1 values`

### 编写尽可能少的代码作测试并且检查失败的输出
```
func Racer(a, b string) (winner string, error error) {
    select {
    case <-ping(a):
        return a, nil
    case <-ping(b):
        return b, nil
    }
}
```

修改`Racer`的签名，返回“胜利者”和一个错误提示。返回`nil`作为我们乐观的情况。

编译器将抱怨我们的第一个测试中仅找寻一个值，因此，修改代码为`got, _ := Racer(slowURL, fastURL)`，了解这些，我们应该检查在乐观环境下没有得到错误提示的代码。

如果11秒后运行，将会致错。
```
-------- FAIL: TestRacer (12.00s)
    --- FAIL: TestRacer/returns_an_error_if_a_server_doesn't_respond_within_10s (12.00s)
        racer_test.go:40: expected an error but didn't get one
```
###  编写足够的代码完善测试
```
func Racer(a, b string) (winner string, error error) {
    select {
    case <-ping(a):
        return a, nil
    case <-ping(b):
        return b, nil
    case <-time.After(10 * time.Second):
        return "", fmt.Errorf("timed out waiting for %s and %s", a, b)
    }
}
```

在`select`中使用，`time.After`是一个非常便利的方法。虽然在我们这个测试中未使用，但如果你正在监听的信道永远不会返回值，则可能会编写了永久阻止的代码。`time.After`返回一个帧（类似Ping）并且会向下发送一个信号在你定义的时间后。
对我们来说这是完美的，如果a或者b返回他们的”成功“，但是如果在`time.After`十秒后发出信好，完美将返回一个错误信息。

### 缓行测试
此处的问题是我们将用10秒钟运行程序。对于这样一个简单的逻辑，并不是太好。

我们能做的是将超时做到可配置，所以在我们的测试中，我们设置一个非常短的超时时间，今后在真实的环境中，将其设置为10秒。
```
func Racer(a, b string, timeout time.Duration) (winner string, error error) {
    select {
    case <-ping(a):
        return a, nil
    case <-ping(b):
        return b, nil
    case <-time.After(timeout):
        return "", fmt.Errorf("timed out waiting for %s and %s", a, b)
    }
}
```
现在我们的测试并不进行编译，因为我们没有提供超时设置。
在我们急于将默认值添加至我们的测试中时，让我们先听听以下的内容。
- 我们是否关心超时在“乐观”的测试中？
- 有关超时的需求是明确的。
鉴于这些内容，我们做一些重构，以更贴近我们的测试和使用它的用户。
```
var tenSecondTimeout = 10 * time.Second

func Racer(a, b string) (winner string, error error) {
    return ConfigurableRacer(a, b, tenSecondTimeout)
}

func ConfigurableRacer(a, b string, timeout time.Duration) (winner string, error error) {
    select {
    case <-ping(a):
        return a, nil
    case <-ping(b):
        return b, nil
    case <-time.After(timeout):
        return "", fmt.Errorf("timed out waiting for %s and %s", a, b)
    }
}
```

我们的用户及我们第一个测试可以使用`Racer`（使用`ConfigurableRacer`），还有我们的不太友好的测试可以使用`ConfigurableRacer`.

```
func TestRacer(t *testing.T) {

    t.Run("compares speeds of servers, returning the url of the fastest one", func(t *testing.T) {
        slowServer := makeDelayedServer(20 * time.Millisecond)
        fastServer := makeDelayedServer(0 * time.Millisecond)

        defer slowServer.Close()
        defer fastServer.Close()

        slowURL := slowServer.URL
        fastURL := fastServer.URL

        want := fastURL
        got, err := Racer(slowURL, fastURL)

        if err != nil {
            t.Fatalf("did not expect an error but got one %v", err)
        }

        if got != want {
            t.Errorf("got '%s', want '%s'", got, want)
        }
    })

    t.Run("returns an error if a server doesn't respond within 10s", func(t *testing.T) {
        server := makeDelayedServer(25 * time.Millisecond)

        defer server.Close()

        _, err := ConfigurableRacer(server.URL, server.URL, 20*time.Millisecond)

        if err == nil {
            t.Error("expected an error but didn't get one")
        }
    })
}
```

我在第一次测试中添加了一个最终检查，以验证我们没有得到错误的结果。

### 总述
`select`
- 帮助你对多个通道的结果进行等待处理
- 有些时候你可以使用包括`time.After` ，在你其中一个`cases`来避免陷入挂起。
`httptest`
- 创建测试服务的便捷方式，使你可以进行可靠和可控的测试。
- 使用与真实的`net/http`服务相同一致的接口，学习成本将更小。