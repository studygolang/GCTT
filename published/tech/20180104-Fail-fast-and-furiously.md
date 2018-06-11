已发布：https://studygolang.com/articles/12900

# 周五有感：失败，快速且猛烈

**摘要**：*在实现一个服务或 API 时，如果你收到一个你不太理解的请求，最友好的方式就是返回一个错误信息。*

让我们来考虑这样一个 API：

`GET /mySum?num=3&num=42`

非常简单，是不是？我可能会这样来实现它：

```go
func mySum(args url.Arguments, w http.ResponseWriter) {
	w.Write(int(args["num"][0]) + int(args["num"][1]))
}
```

等一下：如果收到一个额外的参数 `foo` 应该怎么办？一个常见的做法就是忽略它：*我们常常会让我们的 API 对请求中它所理解的部分做尽可能多的事，以此来让 API 发挥作用。* 在这个简单的例子中，忽略 `&foo=bar` 这样一个额外的参数看起来并没有什么影响。

但是，如果收到了第三个 `num` 参数该怎么办？这在更大程度上是一个灰色区域：按照前面的实现这个参数将直接被忽略。然而调用者很合理的可能预期这个参数也被包含在最终的加和中。在这个服务中我们没有对传入的参数进行校验，这带来了一个调用者不易发现的 bug，它是最糟糕的 bug 类型：***无提示故障***(Silent failures)。

无提示故障很*可怕*。它们让我夜间盗汗。他们可能会潜伏很长时间都不被侦测到，甚至是无限期的：在此期间我们认为一切都没问题。这种故障的排错是非常棘手的，因为它们是无提示的。它们带来了很高的，可能意味着真金白银的机会成本。

更糟糕的是，当 API 的复杂性增加，*特别是*当服务之间出现互相依赖，无提示故障带来的影响将更难去分辨，也会传播更广。这是一个复合的问题。相比大量的错误日志、段错误、异常，甚至是面向客户的错误，无提示故障是我们最不愿意面对的。

对此我们该做什么呢？在此我想介绍一个非常有用的模式，这个模式在构建可信赖的，避免无提示故障的服务时非常有用：

(1) 给参数以明确的表示。通常，Thrift RPC 或者 gRPC 可以很好的帮你。对于 mySum API，我们将有这样的结构：

```go
type MySumParams struct {
	Lhs, Rhs int
}
```

(2) 编写*校验器*来检验参数是否符合它的类型在语义上的预期。例如，如果 MySumParams 只接受正数，那么它应该定义一个 `Validate()` 函数，当接收到的参数不是正数，返回一个有意义的错误信息。在 go 的代码库中，对于这些检验器，我们有一个指定的函数签名：`func Validate() error`。

*认真地考虑*为每一个 Thrift RPC 或 gRPC 类型定义一个校验器。通过保证客户端和服务器不会出现不一致的预期，这增加了很大的安全性。而这些不一致的预期，往往在服务有修改或者更新时容易出现（例如，我们之后可能将 `mySum` 修改为允许参数为负）。

对于 REST API 来说，校验器应当与参数的传递和提取解耦，因为他们关心的是不同角度，分开也更加容易测试。

(3) 一旦参数被解析出来，将它们从请求的参数列表中删除。Thrift 和 gRPC 倾向于帮你管理这一切，但是例如对如下的 REST API：

```go
func extractMySumParams(args url.Arguments) MySumParams {
	var result = MySumParams{
		Lhs: args["num"][0],
		Rhs: args["num"][1],
	}
	// Strip consumed num arguments.
	args["num"] = args["num"][:2]
}
```

我可能会在使用 `extractMySumParams` 的同时，也使用一些与它没有任何关系的参数提取器，这些提取器也会从 `args` 中弹出参数。然而，当我的 API 解析并验证了所有的参数，我就可以断言 `args` 已空（如果非空，应当返回 error）。

将参数的处理分解为提取、验证、断言这几个单独的步骤被证明是非常有帮助的，可以帮助我们尽早的捕捉到问题，并快速地定位问题的根源。

— — — — — — — — — — — — — — — — — — — — — — — — — — — — — — —

*这篇文章作者是 John Graettinger, 我们纽约办公室的开发架构主管。这是由 LiveRamp 开发团队推出的关于最佳实践的周五有感系列的一部分。你喜欢一直追寻自身提升的开发团队吗？[我们一直在招聘](https://liveramp.com/careers/engineers/)*

---

via: https://medium.com/@LiveRamp/friday-thoughts-fail-fast-and-furiously-26705b55fbcc

作者：[LiveRamp](https://medium.com/@LiveRamp)
译者：[krystollia](https://github.com/krystollia)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
