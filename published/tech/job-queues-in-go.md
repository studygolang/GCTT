已发布：https://studygolang.com/articles/12621

# Go 中的任务队列

在 [RapidLoop](https://www.rapidloop.com/) 中，我们几乎用 [Go](https://golang.org) 做所有事情，包括我们的服务器，应用服务和监控系统 [OpsDash](https://www.opsdash.com/)。

Go 十分擅长编写异步程序 - goroutine 和 channel 使用十分简单不容易出错并且和其他语言相比异步/等待模式，语法和功能都更加强大。请继续阅读来瞧瞧围绕任务队列的一些有趣的 Go 代码。

## 不使用任务队列

有时候你不需要任务队列。执行一个异步任务可以这样：

```go
go process(job)
```

这种方式对于一些需求确实是很好的方式，例如在处理 HTTP 请求的时候发送 email。需求的规模和复杂度决定我们是否需要更精细化的基础设施去处理任务，并将任务队列化以一种可控的方式处理它们（例如控制最大并行的任务数量）。

## 简单的任务队列

这里有一个简单的队列和一个处理队列任务 worker 函数。goroutine 和 channel 只是将其编码成优雅紧凑代码块的正确抽象。

```go
func worker(jobChan <-chan Job) {
	for job := range jobChan {
		process(job)
	}
}

// make a channel with a capacity of 100.
jobChan := make(chan Job, 100)

// start the worker
go worker(jobChan)

// enqueue a job
jobChan <- job
```

以上代码创建了一个容积为 100 的 Job 对象 channel。然后创建了一个 goruntine 执行 worker 函数。worker 从 channel 中取出任务并一次处理 1 个任务。任务可以通过推送进 channel 中进行排队。

虽然只用了几行代码，却已经做很多事情。首先它安全，正确，无竞态却不用混合复杂的锁和线程代码。

另外的功能就是生产者调节。

## 生产者调节

创建一个容积为 100 的 channel：

```go
// make a channel with a capacity of 100.
jobChan := make(chan Job, 100)
```

然后这样将任务插入队列：

```go
// enqueue a job
jobChan <- job
```

如果已经有 100 个还没处理的任务在 channel 中它就会阻塞。这通常来说是一件好事情。如果有 SLA/QoS 限制或者其他假设的条件（例如任务需要一定的时间才能完成），你肯定不想积压太多的任务。如果一个任务需要花费 1 秒钟，那么最多只需 100 秒就能完成你的工作。

如果 channel 满了，你希望你的调用者能在一段时间后返回。例如：一个 REST API 请求，你可以返回一个 503 错误码并且调用者可以稍后重试。通过这种方式可以进行压力测试来保证服务质量。

## 非阻塞入队

如果想尝试入队，在需要阻塞的时候返回 fail 怎么办？这种方式能够获取提交任务的失败状态返回 503。关键在于使用 select 的 default 语句：

```go
// TryEnqueue tries to enqueue a job to the given job channel. Returns true if
// the operation was successful, and false if enqueuing would not have been
// possible without blocking. Job is not enqueued in the latter case.
func TryEnqueue(job Job, jobChan <-chan Job) bool {
	select {
	case jobChan <- job:
		return true
	default:
		return false
	}
}
```

你能使用这种方式来返回失败状态：

```go
if !TryEnqueue(job, chan) {
	http.Error(w, "max capacity reached", 503)
	return
}
```

## 停止 worker

我们如何才能优雅的停止 worker 处理函数呢？假定我们不再向队列中插入任务并且保证所有队列中的任务可以处理完成，你只需这么做：

```go
close(jobChan)
```

没错，只需这么做。它会按照预期工作是因为在 `for ... range` 循环会弹出任务：

```go
for job := range jobChan {...}
```

并且循环会在 channel 关闭的时候退出。在关闭 channel 之前的入队的所有任务都会正常弹出并被 worker 处理。

## 等待 worker 处理

这看起来很容易，不过 `close(jobChan)` 不会等待 goroutine 完成就会退出。因此我们还需使用 sync.WaitGroup：

```go
// use a WaitGroup
var wg sync.WaitGroup

func worker(jobChan <-chan Job) {
	defer wg.Done()

	for job := range jobChan {
		process(job)
	}
}

// increment the WaitGroup before starting the worker
wg.Add(1)
go worker(jobChan)

// to stop the worker, first close the job channel
close(jobChan)

// then wait using the WaitGroup
wg.Wait()
```

这样，我们可以通过关闭 channel 给 worker 发送关闭信号并使用 wg.Wait 会等待 worker 处理完成以后才会退出。

注意：我们必须在开始 goroutine 之前递增 wait group，并且在 goroutine 结束（不管以何种方式）时递减。

## 附带超时的等待

`wg.Wait()` 会在 goroutine 退出前一直等待。但是如果我们无法无限期的等待怎么办？

如下帮助函数封装了 `wg.Wait()` 增加了超时时间：

```go
// WaitTimeout does a Wait on a sync.WaitGroup object but with a specified
// timeout. Returns true if the wait completed without timing out, false
// otherwise.
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()
	select {
	case <-ch:
			return true
	case <-time.After(timeout):
			return false
	}
}

// now use the WaitTimeout instead of wg.Wait()
WaitTimeout(&wg, 5 * time.Second)
```

现在我们在一定时限内等待 worker 退出，如果超过时限就会直接退出。

## 取消 worker

现在我们能让 worker 即使是在我们发出停止信号之后也能处理完它们的工作。可是如果我们想让 worker 抛弃当前的工作直接退出的话应该怎么做？

我们使用了 `context.Context`:

```go
// create a context that can be cancelled
ctx, cancel := context.WithCancel(context.Background())

// start the goroutine passing it the context
go worker(ctx, jobChan)

func worker(ctx context.Context, jobChan <-chan Job) {
	for {
		select {
		case <-ctx.Done():
			return

		case job := <-jobChan:
			process(job)
		}
	}
}

// Invoke cancel when the worker needs to be stopped. This *does not* wait
// for the worker to exit.
cancel()
```

总的来说，我们创建了一个"可以取消的 context"。worker 同时等待这个 context 和工作队列，而 `ctx.Done()` 会在 `cancel()` 函数调用时返回。

和关闭任务队列相似，`cancel()` 函数只会发送取消信号而不会等待取消操作完成。所以如果你需要等待 worker 退出（即使等待的时间非常短而且没有其他任务需要执行）你必须添加 wait group 代码。

但是这段代码有一点比较困难。如果你在 channel 中积压了一些工作（<-jobChan 不会阻塞），并且已经调用了 cancel() 函数（<-ctx.Done() 也不会阻塞）。因为两个 case 都没阻塞，`select` 必须在它们之间作出选择。

事实上不会出现这种情况。尽管在 `<-ctx.Done()` 没有阻塞时也会选择  `<-jobChan` 的情况看起来很合理，不过在实际使用的时候很容易让人苦恼。即使我们调用了取消函数，可 channel 依旧会弹出任务，如果我们插入更多任务，它会一直这样错误地运行。

不过我们不需要担心，但是要注意。context 的取消 case 应该比其他 case 有更高的优先级。这样做很不容易，不过使用 Go 提供的内置功能就能解决这个问题。

可以使用一个参数可以帮助我们完成目的：

```go
var flag uint64

func worker(ctx context.Context, jobChan <-chan Job) {
	for {
		select {
		case <-ctx.Done():
			return

		case job := <-jobChan:
			process(job)
			if atomic.LoadUint64(&flag) == 1 {
				return
			}
		}
	}
}

// set the flag first, before cancelling
atomic.StoreUint64(&flag, 1)
cancel()
```

也可以使用 context 的 `Err()` 函数：

```go
func worker(ctx context.Context, jobChan <-chan Job) {
	for {
		select {
		case <-ctx.Done():
			return

		case job := <-jobChan:
			process(job)
			if ctx.Err() != nil {
				return
			}
		}
	}
}

cancel()
```

我们在运行任务之前不检查 `flag/Err()` 因为我们想在任务弹出以后先把任务处理完再退出。当然如果你想让退出的优先级高于处理任务，也可以在处理之前检查。

底线就是要么在退出 worker 之前做一些额外的工作，要么仔细设计代码绕过这种缺陷。

## 不使用 context 取消 worker

`context.Context` 并不适用所有情况，有时不使用 context 可能会让代码更加整洁清晰：

```go
// create a cancel channel
cancelChan := make(chan struct{})

// start the goroutine passing it the cancel channel
go worker(jobChan, cancelChan)

func worker(jobChan <-chan Job, cancelChan <-chan struct{}) {
	for {
		select {
		case <-cancelChan:
			return

		case job := <-jobChan:
			process(job)
		}
	}
}

// to cancel the worker, close the cancel channel
close(cancelChan)
```

这其实就是 context（简单，无层级）的取消操作。不幸的是它也存在上面提到的问题。

## worker 池

最后，多个 worker 可以让任务并行处理。最简单的方式就是创建多个 worker 并在相同的任务 channel 中获取任务：

```go
for i:=0; i<workerCount; i++ {
	go worker(jobChan)
}
```

其他的代码不需要修改。这样多个 worker 会尝试从相同的 channel 中读取任务。这样既有效又安全。只有一个 worker 会获取到任务，其他的 worker 都会阻塞等待任务到来。

这也存在合理分配的问题。试想：总共有 100 个任务，4 个 worker，那么每个 worker 应该处理 25 个任务。但是事实有可能并不是这样，所以的代码不应该建立在这种假设上。

想等待 worker 的退出，可以添加 wait group：

```go
for i:=0; i<workerCount; i++ {
	wg.Add(1)
	go worker(jobChan)
}

// wait for all workers to exit
wg.Wait()
```

如果想取消操作，你可以创建一个取消 channel，关闭它会取消所有 worker。

```go
// create cancel channel
cancelChan := make(chan struct{})

// pass the channel to the workers, let them wait on it
for i:=0; i<workerCount; i++ {
	go worker(jobChan, cancelChan)
}

// close the channel to signal the workers
close(cancelChan)
```

## 一个普通的任务队列库

表面上看，任务队列很简单，可以把它封装成一个通用的，可重用的组件。而事实上，你往往需要在不同的地方对这个通用组件添加更复杂的功能。加上在 Go 中写一个队列比其他语言都要简单，所以你可以在每个需要队列的地方编写自己的队列。

## 许可证书

以上所有代码都经由 MIT 证书发布：

```
Copyright (c) 2017 RapidLoop, Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```

----------

via: https://www.opsdash.com/blog/job-queues-in-go.html

作者：[Mahadevan Ramachandran](https://twitter.com/mdevanr)
译者：[saberuster](https://github.com/saberuster)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
