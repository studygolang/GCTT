首发于：https://studygolang.com/articles/35386

# 让 Go 的错误处理更加强大

Go 所提供的默认的 `errors` 包有很多的不足。编写多层架构应用程序并使用 API 公开功能的时候，相比于单纯的 `string` 类型的值，更需要具有上下文信息的错误处理。意识到这个缺点后，我开始实现一个更强大，更优雅的 error 包。这是一个逐渐演化的过程，随着时间推移，我需要在这个包中引入更多的功能。

在此，我们会探讨我们如何使用一个 `CustomError` 数据类型为应用中带来更多的价值，并且使错误处理更强大。

首先需要明白的是，如果实现了 `Error()` 方法，Go 允许使用任何用户定义的数据类型替代内置的 `error` 数据类型。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/implements-an-Error.png)

换言之，只要我们自定义的数据类型实现了返回一个 `string` 类型的 `Error()` 方法，我们就可以使用我们自己的数据类型代替 Go 默认提供的那个；任何应该返回 `error` 类型的函数都可以返回我们自定义的数据类型，并且运行如常。（译者注：这里所谓原生的 error “数据类型”，就是 error interface）

## 构建 'CustomError' 类型

1、我们创建一个新的数据类型，该数据类型在应用程序中被解释为 error。我们将它命名为 `CustomError`，首先，它会包含一个默认的 `error` 类型。这个 `error` 字段能够让我们在 `CustomError` 初始化的时候使用堆栈的跟踪信息对其进行注释（更多相关细节请看[这里](https://github.com/abhinav-codealchemist/custom-error-go/blob/ca21f0e42b4ed57b5390491fe25fcb16eec7cffa/CustomError.go#L23)）。记录这些堆栈信息可以让我们在诸如 NewRelic APM 之类的平台上更容易地调试错误。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/CustomError.png)

2、Go 内置的 `error` 类型将错误当做 `string` 类型的值。我一直认为这种方式是不对的，至少这样是不足的。一个错误应当有与之关联的类型——比如错误是否由于 SQL DB 写入/更新/获取操作失败而导致的？或者错误是否是由于请求中没有提供足够的数据导致的？

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/CustomError-add-code.png)

现在，更加深入地了解下 `ErrorCode` 类型的实际情况。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/ErrorCode.png)

我们依据可能在应用中捕获到的每个错误类型创建多个 `ErrorCode` 常量。这样一个 error 的解释就基于 `ErrorCode` 而不是字符串了。

获知一个 error 的类型可以让我们以不同的方式处理不同的错误，也可以基于 error 的不同类型采取业务层面的决策。不仅仅只是业务决策，快速浏览下 `ErrorCode` 甚至可以指出系统出故障的精确范围。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/precise-region.png)

举例来说，我在系统中的主要 RPC 中使用了大量的 5xx 错误在不同时间段的计数 （`5xx Count vs Time `）。通过浏览这个图可以有助于定位系统错误的准确范围。

3、出于提高可读性的目标，我们同时也希望保留 error 的 `string` 解释，这样也能让我们一眼就了解哪里出了问题。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/CustomError-add-msg.png)

`ErrorCode` 告诉你一个错误的类型，但是不会告诉你它发生在代码中的哪个位置。`ErrorMsg` 则服务于这个目的，并保持现有功能不变。

这里可能有争议，“为什么不使用在第 1 步中引入的 `error` 字段的错误信息？为什么要新加一个字段？”——这是因为我们会使用 `ErrorCode` 和 `ErrorMsg` 的组合来作为一个 `error` （代码请参考[这里](https://github.com/abhinav-codealchemist/custom-error-go/blob/ca21f0e42b4ed57b5390491fe25fcb16eec7cffa/CustomError.go#L22)）。

但是仅有这两个就足够了吗？如果我们想捕获更多信息——比如导致一个 error 的上下文数据呢？

比方说，一个顾客打开一个查询周围餐厅的页面，但是当他打开应用的时候，却没有给他展示任何东西。你也许需要捕获更多上下文信息，比如 ` 纬度 `，` 经度 ` 和其他一系列信息，方便你用来判断是该区域真的没有餐厅，还是这是一个维护性上的问题。

`LoggingParams` 正是解决办法！

4、`LoggingParams` 允许你捕获各种依赖于上下文的参数。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/LoggingParams.png)

比如，我将这些在严重错误中产生的参数追加在告警信息中，看一下这些参数，有时就可以找出错误的根本原因。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/root-cause-of-the-error.png)

如果没有找到原因，它也有助于过滤出特定事件的错误日志。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/error-log-1.png)

![这是我们记录 CustomError 时 Kibana 日志看起来的样子](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/error-log-2.png)

5、最后，我们需要一些类似于 Go 中默认 `error` 类型申明 `error != nil` 之类的东西。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/add-exists.png)

我们引入了一个名为 `exists` 的 `bool` 字段，当 `CustomError` 被创建时，该字段会被初始化为 `true`。

6、对于小部分错误，我们希望想向最终用户展示特定信息。这些信息便于他们理解并做出反应。

我们同样维护了一个 `ErrorCode` 到用户界面友好（UI-friendly）的信息的映射。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/UI-friendly-message.png)

使用：
![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20210502-Error-Handling-in-Go-made-more-Powerful/Usage.png)

看上面的例子，你会意识到每次 `Acquire()` 被调用，调用函数会期望返回 `CustomError` 的错误类型，与此同时，这会给我们带来全面提升。

> 译者注：站在 error 使用角度。上图中的做法可能并不是很好。 Acquire() 函数返回的不是 error 而是作者自定义的 CustomError 结构体，等同于要与这个 CustomError 耦合在一起。如果 Acquire() 函数仍返回 error，上层通过 errors.As 解析出 CustomError 可能会更好。
> 另外，error 实际上是 interface。将一个 struct 类型或者 struct 的指针赋值给 interface 的变量，则必定不为空。执行下面的语句，err 必定不等于 nil。
>
>```go
> var err error
> err = Acquire()
> 
> if err != nil {
>  ......
> }
>```


## 结论

我们已经看到了 `CustomError` 可以被用来使得 error 在多层次应用中更有解释力。您可以在[这里](https://github.com/abhinav-codealchemist/custom-error-go)查看具有正确接口定义和实现的完整代码。

---

via: https://medium.com/codealchemist/error-handling-in-go-made-more-powerful-ce164c2384ee

作者：[Abhinav Bhardwaj](https://codealchemist.medium.com/)
译者：[dust347](https://github.com/dust347)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
