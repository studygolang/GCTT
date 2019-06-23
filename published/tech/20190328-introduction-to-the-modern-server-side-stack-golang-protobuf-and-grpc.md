首发于：https://studygolang.com/articles/21374

# 现代服务端技术栈介绍 —— Golang、Protobuf 和 gRPC

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/modern-server/1_W_1gSpbn79xtYZw9_TBGPA.jpg)

身边多了一些新的服务端程序员，这一次，一切都与谷歌有关。自从谷歌开始在他们的生产系统中使用 Go 语言以来，`go` 语言的流行度快速地提升。从微服务架构创立至今，人们一直在关注像 `gRPC`、`Protobuf` 这样的现代数据通信解决方案。在这篇文章中，我将分别对它们作简要的介绍。

## Golang

`Golang`，或者说 `Go` 语言，是谷歌开发的一门开源、通用的编程语言。由于多种原因，它的流行度最近一直在提升。说起来可能令人惊讶，据谷歌宣称，这门语言已经将近 10 岁了，并且近 7 年都可用于生产环境。

`Golang` 被设计为简单的、现代的、易于理解并且可快速掌握的语言。普通程序员都能在一周内掌握如何使用。我作证，他们绝对成功了。说到 `Golang` 的开发者，他们是设计 C 语言草案的专家，因此，我们可以确定他们知道自己在做什么。

## 既然一切都很好，为什么我需要另一门编程语言

绝大多数情况下，我们确实并不需要。实际上， `Go` 并没有解决任何其他语言或工具未能解决的新问题。不过，它确实尝试着用一种高效、优雅且符合直觉的方式，去解决我们面对的一些特定问题。`Go` 优先考虑如下几点：

* 对并发的高度支持
* 一门优雅、现代、核心简要明确的语言
* 非常优秀的性能
* 对现代软件开发所需工具的基本支持

我将简要解释 `Go` 如何做到上面几点。在 [Go 语言官方网站](https://golang.org/) ，你可以详细了解更多有关 `Go` 及其特性的细节。

## 并发

并发是大多数服务端程序的主要关注点之一，考虑到现代微处理器，这也应该是 `Go` 语言的首要关注点。 `Go` 引入了 `goroutine` 的概念。一个 `goroutine` 类似一个“轻量级的用户空间线程”。事实上要复杂的多，多个 `goroutine` 会复用同一个线程，不过这个解释能让你明白个大概。`goroutine` 非常轻量，你可以同时自旋百万个 `goroutine`，因为它们启动时只需极少的栈空间。事实上，使用 `goroutine` 也是推荐的做法。`Go` 语言的任何函数或方法，都可以用来开启一个 `Goroutine`。你只需要调用 `go myAsyncTask()` 来从 `myAsyncTask` 函数中开启一个 `goroutine`。下面是一个例子：

```go
// This function performs the given task concurrently by spawing a Goroutine
// for each of those tasks.
func performAsyncTasks(task []Task) {
	for _, task := range tasks {
		// This will spawn a separate Goroutine to carry out this task.
		// This call is non-blocking
		go task.Execute()
	}
}
```

是的，就是这么简单，既然 `Go` 是一门简单的语言，它也确该如此。你应当为每个独立的异步任务开启一个 `goroutine` 而无需担心太多。如果多核可用，`Go` 的 `runtime` 会自动并发运行 `goroutine`。但是，这些 `goroutine` 间如何通信呢？答案是 `channel`。

`Channel` 也是一个语言原语，它被用于 `goroutine` 间通信。通过 `channel`，你可以向另一个 `goroutine` 传递任何数据（原生类型、结构类型甚至其他 `channel`）。本质上，`channel` 是一个阻塞的双向队列（也可以是单向的）。如果你想要 `goroutine` 等待，直到特定的条件满足才继续运行，你可以使用 `channel` 来实现 `goroutine` 间的合作阻塞。

在编写异步或并发的代码时，这两个原语提供了极大的灵活性和简洁性。使用如上原语可以非常容易地创建出类似 `goroutine 池` 等有用的库。一个基本的例子如下：

```go
package executor

import (
	"log"
	"sync/atomic"
)

// The Executor struct is the main executor for tasks.
// 'maxWorkers' represents the maximum number of simultaneous Goroutines.
// 'ActiveWorkers' tells the number of active Goroutines spawned by the Executor at given time.
// 'Tasks' is the channel on which the Executor receives the tasks.
// 'Reports' is channel on which the Executor publishes the every tasks reports.
// 'signals' is channel that can be used to control the executor. Right now, only the termination
// signal is supported which is essentially is sending '1' on this channel by the client.
type Executor struct {
	maxWorkers    int64
	ActiveWorkers int64

	Tasks   chan Task
	Reports chan Report
	signals chan int
}

// NewExecutor creates a new Executor.
// 'maxWorkers' tells the maximum number of simultaneous Goroutines.
// 'signals' channel can be used to control the Executor.
func NewExecutor(maxWorkers int, signals chan int) *Executor {
	chanSize := 1000

	if maxWorkers > chanSize {
		chanSize = maxWorkers
	}

	executor := Executor{
		maxWorkers: int64(maxWorkers),
		Tasks:      make(chan Task, chanSize),
		Reports:    make(chan Report, chanSize),
		signals:    signals,
	}

	go executor.launch()

	return &executor
}

// launch starts the main loop for polling on the all the relevant channels and handling differents
// messages.
func (executor *Executor) launch() int {
	reports := make(chan Report, executor.maxWorkers)

	for {
		select {
		case signal := <-executor.signals:
			if executor.handleSignals(signal) == 0 {
				return 0
			}

		case r := <-reports:
			executor.addReport(r)

		default:
			if executor.ActiveWorkers < executor.maxWorkers && len(executor.Tasks) > 0 {
				task := <-executor.Tasks
				atomic.AddInt64(&executor.ActiveWorkers, 1)
				go executor.launchWorker(task, reports)
			}
		}
	}
}

// handleSignals is called whenever anything is received on the 'signals' channel.
// It performs the relevant task according to the received signal(request) and then responds either
// with 0 or 1 indicating whether the request was respected(0) or rejected(1).
func (executor *Executor) handleSignals(signal int) int {
	if signal == 1 {
		log.Println("Received termination request...")

		if executor.Inactive() {
			log.Println("No active workers, exiting...")
			executor.signals <- 0
			return 0
		}

		executor.signals <- 1
		log.Println("Some tasks are still active...")
	}

	return 1
}

// launchWorker is called whenever a new Task is received and Executor can spawn more workers to spawn
// a new Worker.
// Each worker is launched on a new Goroutine. It performs the given task and publishes the report on
// the Executor's internal reports channel.
func (executor *Executor) launchWorker(task Task, reports chan<- Report) {
	report := task.Execute()

	if len(reports) < cap(reports) {
		reports <- report
	} else {
		log.Println("Executor's report channel is full...")
	}

	atomic.AddInt64(&executor.ActiveWorkers, -1)
}

// AddTask is used to submit a new task to the Executor is a non-blocking way. The Client can submit
// a new task using the Executor's tasks channel directly but that will block if the tasks channel is
// full.
// It should be considered that this method doesn't add the given task if the tasks channel is full
// and it is up to client to try again later.
func (executor *Executor) AddTask(task Task) bool {
	if len(executor.Tasks) == cap(executor.Tasks) {
		return false
	}

	executor.Tasks <- task
	return true
}

// addReport is used by the Executor to publish the reports in a non-blocking way. It client is not
// reading the reports channel or is slower that the Executor publishing the reports, the Executor's
// reports channel is Going to get full. In that case this method will not block and that report will
// not be added.
func (executor *Executor) addReport(report Report) bool {
	if len(executor.Reports) == cap(executor.Reports) {
		return false
	}

	executor.Reports <- report
	return true
}

// Inactive checks if the Executor is idle. This happens when there are no pending tasks, active
// workers and reports to publish.
func (executor *Executor) Inactive() bool {
	return executor.ActiveWorkers == 0 && len(executor.Tasks) == 0 && len(executor.Reports) == 0
}
```

## 简单的语言

与很多其他的现代语言不同，`Golang` 的特性并不多。事实上，有令人信服的案例可以证明 `Go` 在它的特性集上太过严格，这是有意为之的。它并不是围绕着像 `Java` 一样的编程范例来设计的，也不像 `Python` 一样可以支持多种编程范例。它仅仅只是简单的结构化编程。除了必要的特性外，这门语言并没有多余的东西。

了解这门语言后，你会觉得它并没有遵循特定的哲学或方向，感觉像是所有能解决特定问题的特性都被包含在内，仅此而已。例如，它有方法和接口，但是没有类；编译器生成一个静态链接的二进制文件，却仍然有一个垃圾回收器；它有严格的静态类型，却不支持泛型；它有一个瘦的 `runtime`，却不支持异常。

这些的主要意图是，开发人员应该花最少的时间用代码表达自己的想法或算法，而无需考虑“在 X 语言中，这样做的最好方式是什么？”，并且，其他人应该易于理解。它仍然并不完美，确实不时地让人感到受限制，像泛型和异常这样必要的特性正在被考虑加入 `Go 2`。

## 性能

单线程执行性能并不是评价语言的好的指标，尤其当这门语言专注于并发和并行。但是，`Golang` 仍然跑出了优秀的基准测试数据，仅仅被诸如 `C`，`C++`，`Rust` 这样的底层系统编程语言打败。它的性能仍在不断提升中。考虑到它是一门“垃圾回收型”语言，它的性能确实非常优秀，足以在任何场景下使用。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/modern-server/1_2SrMSeCcNmTExRXE-Fpr4A.png)

## 开发工具

对新工具或语言的采用直接取决于它的开发人员的经验。`Go` 的使用确实代言了它的工具。我们可以发现，相同的想法和工具非常小但是很有效。这都是通过 `go` 命令和它的子命令实现的，全部是命令行。

这门语言没有像 `pip`、`npm` 一样的包管理工具。但是你可以获取任何社区的包，只需要这样做：

```go
go get gitHub.com/farkaskid/WebCrawler/blob/master/executor/executor.go
```

是的，它成功了。你可以直接从 `GitHub` 或其他任何地方拉取包。它们仅仅是源代码文件。

但是，`package.json` 文件该怎么办呢？我并没有看到任何与 `go get` 等价的命令，因为确实没有。你不需要在一个文件中列出你的所有依赖。你可以直接在你的源代码文件中使用：

```go
import "github.com/xlab/pocketsphinx-go/sphinx"
```

当你执行 `go build` 命令，它将自动为你执行 `go get`。这里是完整的源文件：

```go
package main

import (
	"encoding/binary"
	"bytes"
	"log"
	"os/exec"

	"github.com/xlab/pocketsphinx-go/sphinx"
	pulse "github.com/mesilliac/pulse-simple" // pulse-simple
)

var buffSize int

func readInt16(buf []byte) (val int16) {
	binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &val)
	return
}

func createStream() *pulse.Stream {
	ss := pulse.SampleSpec{pulse.SAMPLE_S16LE, 16000, 1}
	buffSize = int(ss.UsecToBytes(1 * 1000000))
	stream, err := pulse.Capture("pulse-simple test", "capture test", &ss)
	if err != nil {
		log.Panicln(err)
	}
	return stream
}

func listen(decoder *sphinx.Decoder) {
	stream := createStream()
	defer stream.Free()
	defer decoder.Destroy()
	buf := make([]byte, buffSize)
	var bits []int16

	log.Println("Listening...")

	for {
		_, err := stream.Read(buf)
		if err != nil {
			log.Panicln(err)
		}

		for i := 0; i < buffSize; i += 2 {
			bits = append(bits, readInt16(buf[i:i+2]))
		}

		process(decoder, bits)
		bits = nil
	}
}

func process(dec *sphinx.Decoder, bits []int16) {
	if !dec.StartUtt() {
		panic("Decoder failed to start Utt")
	}

	dec.ProcessRaw(bits, false, false)
	dec.EndUtt()
	hyp, score := dec.Hypothesis()

	if score > -2500 {
		log.Println("Predicted:", hyp, score)
		handleAction(hyp)
	}
}

func executeCommand(commands ...string) {
	cmd := exec.Command(commands[0], commands[1:]...)
	cmd.Run()
}

func handleAction(hyp string) {
	switch hyp {
		case "SLEEP":
		executeCommand("loginctl", "lock-session")

		case "WAKE UP":
		executeCommand("loginctl", "unlock-session")

		case "POWEROFF":
		executeCommand("poweroff")
	}
}

func main() {
	cfg := Sphinx.NewConfig(
		sphinx.HMMDirOption("/usr/local/share/pocketsphinx/model/en-us/en-us"),
		sphinx.DictFileOption("6129.dic"),
		sphinx.LMFileOption("6129.lm"),
		sphinx.LogFileOption("commander.log"),
	)

	dec, err := Sphinx.NewDecoder(cfg)
	if err != nil {
		panic(err)
	}

	listen(dec)
}
```

这将依赖声明与源代码本身绑定在一起。

如你所见，它简单、简约、有效、优雅，有对单元测试和带火焰图的基准测试的直接支持。像特性集一样，它也有自己的缺点。例如，`go get` 命令不支持版本，并且你被源文件的导入 `URL` 限制了。随着其他依赖管理工具开始出现，这正在不断改善。

最初，`Golang` 被设计用来解决谷歌在他们大规模的代码基础中遇到的问题，满足他们编写高效并发程序的必要需求。它使得编写程序或库来使用现代芯片的多核性质非常容易。它从不按开发者的方式来做。它只是一门简单的现代语言，从未尝试成为其他的东西。

## Protobuf（Protocol Buffers）

`Protobuf`，或者说 `Protocol Buffers` 是谷歌提出的一种二进制通信格式。它被用来序列化结构化的数据。一种通信格式？就像 `JSON` 一样吗？是的，它诞生十多年了，谷歌使用它已经有一段时间了。

但是，我们不是有无处不在的 `JSON` 吗？

就像 `Golang` 一样，`Protobufs` 实际上并没有解决新问题。它只是用更加高效、现代的方式解决已经存在的问题。和 `Golang` 不同，它们并不需要比现存的解决方案更加优雅。下面是 `Protobuf` 的主要关注点：

* 它是二进制格式的，和 `JSON` 或 `XML` 这样的文本型不同，因此极大地节省了空间。
* 对模式的直接、精巧的支持。
* 直接支持生成多种语言的解析和客户端代码。
* 二进制格式并且快速。

`Protobuf` 确实那么快吗？简明扼要：是的。根据 [谷歌开发者网站](https://developers.google.com/protocol-buffers/docs/overview#whynotxml)，它们比 `XML` 小 3-10 倍，并且快 20-100 倍。由于它是二进制格式，所以这并不令人惊讶，序列化后的数据是人类无法阅读的。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/modern-server/1_FI7JuupJ02r5cVSMqKAuAA.png)

`Protobufs` 采取更有计划的步骤。你定义 `.proto` 文件，这些文件有点像模式文件，但作用更大。本质上，你定义的是你期望的消息被格式化的方式，可选的字段，必需的字段，它们的数据类型等。在这之后，`Protobuf` 编译器将会为你生成类来使用数据。你可以在你的业务逻辑中使用这些类来进行通信。

阅读一个关联到服务的 `.proto` 文件也会让你清楚地了解通信的具体细节和它暴露的特性。一个典型的 `.proto` 文件如下：

```protobuf
message Person {
  required string name = 1;
  required int32 id = 2;
  optional string email = 3;

  enum PhoneType {
	MOBILE = 0;
	HOME = 1;
	WORK = 2;
  }

  message PhoneNumber {
	required string number = 1;
	optional PhoneType type = 2 [default = HOME];
  }

  repeated PhoneNumber phone = 4;
}
```

趣闻：[Jon Skeet](https://stackoverflow.com/users/22656/jon-skeet)，`Stack Overflow` 的王牌，也是这个项目的主要贡献者之一。

## gRPC

`gRPC`，正如你所猜想的，是一个现代的 `RPC`（远程过程调用）框架。它包含了系列框架，内置了对负载均衡，追踪，健康检查和认证的支持。谷歌于 2015 年将它开源，从那时起，它日益流行。

## 一个 RPC 框架？ REST 呢？

作为面向服务架构中不同系统之间的通信方式，使用 `WSDL`（译注：网络服务描述语言）的 `SOAP`( 简单对象访问协议 ) 已经被使用了很长时间。在那时，协议通常被制定得非常严格，系统庞大，耦合严重，暴露了非常多的接口。

后来出现了“浏览”的概念，服务端和客户端不需要紧密耦合。即使服务端提供的服务是独立编码的，客户端也应该能够浏览服务。如果客户端请求一本书的信息，该服务通过请求的内容也可以提供相关书籍的列表以供客户端浏览。`REST` 范例对此至关重要，因为它允许服务端和客户端在没有使用某些原语动词的情况下自由通信。

正如你之前看到的，服务的行为类似一个统一的系统，这个系统除了必需的，还做了其他事情来提供给客户端所预期的“浏览”体验。但用法并不总是这样，不是吗？

## 进入微服务

有足够多的理由去使用微服务架构，一个显著的事实是，扩展一个集成系统是非常困难的。在设计一个使用微服务架构的大型系统时，每个业务或技术要求都被设计作为相互协作组件的多个“微”服务之一。

这些服务的响应不需要全面。它们应该完成特定的职责，提供预期的响应。理论上，它们应该像纯函数一样无缝组合。

如今，在这些服务中使用 `REST` 作为通信范例并没有给我们带来很多好处。但是，为服务暴露 `REST API` 确实提升了服务的表达能力，如果我们既不需要也不打算使用这种表达能力，我们可以使用更加关注其他因素的范例。

`gRPC` 旨在改进传统 `HTTP` 请求的以下技术方面：

* 默认使用 `HTTP/2` 的全部优点
* 使用 `Protobuf` 在机器间通信
* 使用 `HTTP/2` 专注于流通信的支持
* 可插拔的认证，追踪，负载均衡和健康检查，你总会需要这些

由于它是一个 `RPC` 框架，我有依然有像服务定义、接口描述语言这样的概念，这可能让未使用过 `REST` 的人感到陌生，但这次感觉不会那么难用，因为 `gRPC` 使用 `Protobuf` 来实现这两者。

`Protobuf` 的设计使其既可以用作通信格式，也可以用作协议规范工具，而无需引入任何新内容。典型的 `gRPC` 服务定义如下所示：

```protobuf
service HelloService {
  rpc SayHello (HelloRequest) returns (HelloResponse);
}

message HelloRequest {
  string greeting = 1;
}

message HelloResponse {
  string reply = 1;
}
```

你只需为你的服务编写一个 `.proto` 文件来描述接口名称，服务期待的入参和作为返回值的 `Protobuf` 消息。`Protobuf` 编译器将会生成客户端和服务端代码，客户端你可以直接调用它，服务端可以用业务逻辑来实现这些 `API`。

## 结论

`Golang` 和使用 `Protobuf` 的 `gRPC` 是现代服务端编程的新兴技术栈。`Golang` 简化了并发、并行程序，使用 `Protobuf` 的 `gRPC` 为开发者实现高效通信提供了愉悦的体验。

---

via: https://medium.com/velotio-perspectives/introduction-to-the-modern-server-side-stack-golang-protobuf-and-grpc-40407486568

作者：[Velotio Technologies](https://medium.com/@velotio)
译者：[DoubleLuck](https://github.com/DoubleLuck)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出