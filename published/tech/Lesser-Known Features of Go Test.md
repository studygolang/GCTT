已发布：https://studygolang.com/articles/12587

# Go test 少为人知的特性

大多数的 Go 程序员都知道和喜欢用 `go test`，这个测试工具来自于 Go 官方的 `gc` 工具链。（想要执行测试代码）这个命令可能是最简单的了，而且还能做得很漂亮。

大家都知道，运行 `go test` 命令将执行当前目录下的包的测试代码，它会寻找 `*_test.go` 文件，并在这些文件中，寻找符合 `TestXxx(*testing.T){}` 命名的函数和参数（即，接收 `*testing.T` 参数的函数，命名为 `TestXxx`，`Xxx` 可以是任何不以小写字符开头的名字）。这个测试代码不会影响正常的编译过程，只在执行 `go test` 时被使用。

但这里还有很多隐藏的东西。

## 黑盒测试包（The black box test package）

通常情况下，在 Go 语言中，测试和要被测试的代码在同一个包中（被测系统），这样才能访问内部实现细节的代码。为了支持黑盒测试，`go test` 支持使用以 "_test" 后缀命名，并可被编译成独立的包的形式。

如：

```go
// in example.go
package example

var start int

func Add(n int) int {
	start += n
	return start
}

// in example_test.go
package example_test

import (
	"testing"

	. "bitbucket.org/splice/blog/example"
)

func TestAdd(t *testing.T) {
	got := Add(1)
	if got != 1 {
		t.Errorf("got %d, want 1", got)
	}
}
```

你可以在代码中看到臭名昭著的 [点导入](https://golang.org/ref/spec#Import_declarations) 。但当对一个包做黑盒测试时，在当前包的范围内导入（被导入包中）可被导出的符号来说，这是它的一个有实际意义的例子。测试代码[在通常的情况下应该尽量避免进入被测试的环境中](https://code.google.com/p/go-wiki/wiki/CodeReviewComments#Import_Dot)。

就像在点导入的链接符号章节中所解释的一样，黑盒测试模式也能被用来打破循环导入的问题（在被测试的包 “a” 被 “b” 导入，并且 “a“ 的测试也需要导入 ”b“ 时 - 测试可以被移动到 “a_test“ 包，然后可以（同时）导入 “a” 和 “b”，这样就没有循环导入的问题）。

## 跳过测试（Skipping tests）

一些测试可能要求要有特定的上下文环境。例如，一些测试可能需要调用一个外部的命令，使用一个特殊的文件，或者需要一个可以被设置的环境变量。当条件无法满足时，（如果）不想让那些测试失败，可以简单地跳过那些测试：

```go
func TestSomeProtectedResource(t *testing.T) {
	if os.Getenv("SOME_ACCESS_TOKEN") == "" {
		t.Skip("skipping test; $SOME_ACCESS_TOKEN not set")
	}
	// ... the actual test
}
```

如果 `go test -v` 被调用（注意那个冗余（”-v“）标志），输出将会提醒已跳过的测试：

```
=== RUN TestSomeProtectedResource
--- SKIP: TestSomeProtectedResource (0.00 seconds)
		example_test.go:17: skipping test; $SOME_ACCESS_TOKEN not set
```

通常是用 `-short` 命令行标志来实现这个跳过的特性，如果标志被设置的话，反映到代码中，`testing.Short()` 将简单地返回 true（就像是 `-v` 标志一样，如果它被设置，通过判断 `testing.Verbose()` ，你可以打印出额外的调试日志）。

当测试需要运行较长时间时，而你又很着急的话，你可以执行 `go test -short`，（如果）提供这个包的开发者又刚好实现了这个功能，运行时间长的测试将会被跳过。这就是从源码安装时，（通常情况下）Go 测试被执行的样子，这里有 stdlib 库中运行时间较长的测试被跳过的例子：

```go
func TestCountMallocs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	// rest of test...
}
```

跳过只是一个可选项，`-short` 标志只是一个标示，具体还依赖于开发者，他们可以选择（这种标示生效时是否）运行的测试，来避免一些运行比较慢的断言被执行。

这里还有 `-timeout` 标志，它能够被用来强制退出限定时间内没有运行完的测试。例如，运行这个命令 `go test -timeout 1s` 以执行下面的测试：

```go
func TestWillTimeout(t *testing.T) {
	time.Sleep(2 * time.Second)
	// pass if timeout > 2s
}
```

会有如下输出（截断）：

```
=== RUN TestWillTimeout
panic: test timed out after 1s
```

如果想执行特定的测试函数，而不是执行全部的测试集，只需要运行 `go test -run TestNameRegexp`。

### 并行执行测试（Parallelizing tests）

默认情况下，指定包的测试是按照顺序执行的，但也可以通过在测试的函数内部使用 `t.Parallel()` 来标志某些测试也可以被安全的并发执行（和默认的一样，假设参数名为 `t`）。在并行执行的情况下，只有当那些被标记为并行的测试才会被并行执行，所以只有一个测试函数时是没意义的。它应该在测试函数体中第一个被调用（在任何需要跳过的条件之后），因为它会重置测试时间：

```go
func TestParallel(t *testing.T) {
	t.Parallel()
	// actual test...
}
```

在并发情况下，同时运行的测试的数量默认取决于 `GOMAXPROCS`。它可以通过 `-parallel n` 被指定（`go test -parallel 4`）

另外一个可以实现并行的方法，尽管不是函数级粒度，但却是包级粒度，就是类似这样执行 `go test p1 p2 p3`（也就是说，同时调用多个测试包）。在这种情况下，包会被先编译，并同时被执行。当然，这对于总的时间来说是有好处的，但它也可能会导致错误变得具有不可预测性，比如一些资源被多个包同时使用时（例如，一些测试需要访问数据库，并删除一些行，而这些行又刚好被其他的测试包使用的话）。

为了保持可控性，`-p` 标志可以用来指定编译和测试的并发数。当仓库中有多个测试包，并且每个包在不同的子目录中，一个可以执行所有包的命令是 `go test ./...`，这包含当前目录和所有子目录。没有带 `-p` 标志执行时，总的运行时间应该接近于运行时间最长的包的时间（加上编译时间）。运行 `go test -p 1 ./...`，使编译和测试工具只能在一个包中执行时，总的时间应该接近于所有独立的包测试的时间加上编译的时间的总和。你可以自己试试，执行 `go test -p 3 ./...`，看一下对运行时间的影响。

还有，另外一个可以并行化的地方（你应该测试一下）是在包的代码里面。多亏了 Go 非常棒的并行原语，实际上，除非 GOMAXPROCS 通过环境变量或者在代码中显式设置为 GOMAXPROCS=1，否则，包中一个goroutines 都没有用是不太常见的。想要使用 2 个 CPU，可以执行 `GOMAXPROCS=2 go test`，想要使用 4 个 CPU，可以执行 `GOMAXPROCS=4 go test`，但还有更好的方法：`go test -cpu=1,2,4` 将会执行 3 次，其中 GOMAXPROCS 值分别为 1，2，和 4。

`-cpu` 标志，搭配数据竞争的探测标志 `-race`，简直进入天堂（或者下地狱，取决于它具体怎么运行）。竞争探测是一个很神奇的工具，在以高并发为主的开发中不得不使用这个工具（来防止死锁问题），但对它的讨论已经超过了本文的范围。如果你对此感兴趣，可以阅读 Go 官方博客的 [这篇文章](http://blog.golang.org/race-detector)。

## 更多的内容

`go test` 工具支持以与测试函数相似的方式运行基准测试和可断言示例（！）。`godoc` 工具甚至能够理解例子中的语法并将其包含在生成的文档中。

不得不提的还有代码覆盖率和性能测试，测试工具也支持这两个功能。对于感兴趣并想要深入了解的，可以访问 [The cover story](http://blog.golang.org/cover) 和 [Profiling Go programs](http://blog.golang.org/profiling-go-programs)，它们都在 Go 博客中。

在你写自己的测试代码前，建议看一下标准库中的 `testing/iotest`，`testing/quick` 和 `net/http/httptest` 软件包。

----------------

via: https://splice.com/blog/lesser-known-features-go-test/

作者：[MARTIN ANGERS](https://splice.com/blog/author/martin/)
译者：[gogeof](https://github.com/gogeof)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出


