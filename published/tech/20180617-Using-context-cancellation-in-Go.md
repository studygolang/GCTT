首发于：https://studygolang.com/articles/13676

# 在 Go 中用 Context 取消操作

许多使用 Go 的人都会遇到 context 包。大多数时候 context 用在下游操作， 比如发送 Http 请求、查询数据库、或者开 go-routines 执行异步操作。最普通用法是通过它向下游操作传递数据。很少人知道，但是非常有用的context功能是在执行中取消或者停止操作。

这篇文章会解释我们如何使用 Context 的取消功能，还有通过一些 Context 使用方法和最佳实践让你的应用更加快速和健壮。

## 我们为什么需要取消操作?

简单来说，我们需要取消来避免系统做无用的操作。想像一下，一般的http应用，用户请求 Http Server， 然后 Http Server查询数据库并返回数据给客户端：

![http 应用](https://raw.githubusercontent.com/studygolang/gctt-images/master/using-context-cancellation-in-go/1.png)

如果每一步都很完美，耗时时图会像下面这样：

![耗时图](https://raw.githubusercontent.com/studygolang/gctt-images/master/using-context-cancellation-in-go/2.png)

但是，如果客户端中途中断请求会发生什么？会发生，比如： 请求中途，客户端关闭了浏览器。如果没有取消操作，Application Server 和 数据库 会继续他们的工作，尽管工作的结果会被浪费。

![异常耗时图](https://raw.githubusercontent.com/studygolang/gctt-images/master/using-context-cancellation-in-go/3.png)

理想条件下，如果我们知道流程（例子中的 http request）的话， 我们想要下游操作也会停止：

![理想耗时图](https://raw.githubusercontent.com/studygolang/gctt-images/master/using-context-cancellation-in-go/4.png)

## go context包的取消操作

现在我们知道为什么需要取消操作了，让我们看看在 Golang 里如何实现。因为"取消操作"高度依赖上下文，或者已执行的操作，所以它非常容易通过 context 包来实现。

有两个步骤你需要实现：
1. 监听取消事件
2. 触发取消事件

## 监听取消事件

_context_ 包提供了 _Done()_ 方法, 它返回一个当 Context 收取到 _取消_ 事件时会接收到一个 _struct{}_ 类型的 _channel_。
监听取消事件只需要简单的等待 _<- ctx.Done()_ 就好了例如： 一个 Http Server 会花2秒去处理事务，如果请求提前取消，我们想立马返回结果：

```go
func main() {
	// Create an HTTP server that listens on port 8000
	http.ListenAndServe(":8000", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// This prints to STDOUT to show that processing has started
		fmt.Fprint(os.Stdout, "processing request\n")
		// We use `select` to execute a peice of code depending on which
		// channel receives a message first
		select {
		case <-time.After(2 * time.Second):
			// If we receive a message after 2 seconds
			// that means the request has been processed
			// We then write this as the response
			w.Write([]byte("request processed"))
		case <-ctx.Done():
			// If the request gets cancelled, log it
			// to STDERR
			fmt.Fprint(os.Stderr, "request cancelled\n")
		}
	}))
}
```

> 源代码地址： https://github.com/sohamkamani/blog-example-go-context-cancellation

你可以通过执行这段代码, 用浏览器打开 [localhost:8000](http://localhost:8000)。如果你在2秒内关闭浏览器，你会看到在控制台打印了 "request canceled"。

## 触发取消事件

如果你有一个可以取消的操作，你可以通过context触发一个 _取消事件_ 。 这个你可以用 context 包 提供的 _WithCancel_ 方法， 它返回一个 context 对象，和一个没有参数的方法。这个方法不会返回任何东西，仅在你想取消这个context的时候去调用。

第二种情况是依赖。 依赖的意思是，当一个操作失败，会导致其他操作失败。 例如：我们提前知道了一个操作失败，我们会取消所有依赖操作。

```go
func operation1(ctx context.Context) error {
	// Let's assume that this operation failed for some reason
	// We use time.Sleep to simulate a resource intensive operation
	time.Sleep(100 * time.Millisecond)
	return errors.New("failed")
}

func operation2(ctx context.Context) {
	// We use a similar pattern to the HTTP server
	// that we saw in the earlier example
	select {
	case <-time.After(500 * time.Millisecond):
		fmt.Println("done")
	case <-ctx.Done():
		fmt.Println("halted operation2")
	}
}

func main() {
	// Create a new context
	ctx := context.Background()
	// Create a new context, with its cancellation function
	// from the original context
	ctx, cancel := context.WithCancel(ctx)

	// Run two operations: one in a different go routine
	go func() {
		err := operation1(ctx)
		// If this operation returns an error
		// cancel all operations using this context
		if err != nil {
			cancel()
		}
	}()

	// Run operation2 with the same context we use for operation1
	operation2(ctx)
}
```

## 基于时间的取消操作

任何程序对一个请求的最大处理时间都需要维护一个 SLA (服务级别协议)，这可以使用基于时间的取消。这个 API 基本和上一个例子相同，只是多了一点点：

```go
// The context will be cancelled after 3 seconds
// If it needs to be cancelled earlier, the `cancel` function can
// be used, like before
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)

// The context will be cancelled on 2009-11-10 23:00:00
ctx, cancel := context.WithDeadline(ctx, time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))

```

例如:HTTP API 调用内部的服务。 如果服务长时间没有响应，最好提前返回失败并取消请求。

```go
func main() {
	// Create a new context
	// With a deadline of 100 milliseconds
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 100*time.Millisecond)

	// Make a request, that will call the google homepage
	req, _ := http.NewRequest(http.MethodGet, "http://google.com", nil)
	// Associate the cancellable context we just created to the request
	req = req.WithContext(ctx)

	// Create a new HTTP client and execute the request
	client := &http.Client{}
	res, err := client.Do(req)
	// If the request failed, log to STDOUT
	if err != nil {
		fmt.Println("Request failed:", err)
		return
	}
	// Print the statuscode if the request succeeds
	fmt.Println("Response received, status code:", res.StatusCode)
}
```
根据 google 主页对您的请求的响应速度, 您将收到:

```
Response received, status code: 200
```

或者

```
Request failed: Get http://google.com: context deadline exceeded
```

你可以通过设置超时来获得以上2种结果。

## 陷阱和注意事项

尽管 Go 的 context 很好用，但是在使用之前最好记住几点。最重要的就是：context 仅可以取消一次。如果想传播多个错误的话，context 取消并不是最好的选择，最惯用的场景是你真的想取消一个操作，并通知下游操作发生了一个错误。

另一个要记住的是，一个 context 实例会贯穿所有你想使用取消操作的方法和 go-routines 。要避免使用一个已取消的 context 作为 _WithTimeout_ 或者 _WithCancel_ 的参数，这可能导致不确定的事情发生。

---

via: https://www.sohamkamani.com/blog/golang/2018-06-17-golang-using-context-cancellation/

作者：[Soham Kamani](https://github.com/sohamkamani)
译者：[Nelsonken](https://github.com/nelsonken)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
