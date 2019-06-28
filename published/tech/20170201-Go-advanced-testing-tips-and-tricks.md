首发于：https://studygolang.com/articles/21231

# Go 测试高级窍门和技巧

2017 年 2 月 1 日 · 5 min 阅读

这篇文章基于我在 [维尔纽斯的 Golang 交流会](https://www.meetup.com/Vilnius-Golang/) 上的演讲。

我读了很多博客，看了很多演讲并把所有这些窍门和技巧都集中在一个地方。首先我想感谢那些提出这些想法并把它们分享到社区的人。我从下面的这些工作中借鉴了资料和示例：

- [Andrew Gerrand - Testing Techniques](https://talks.golang.org/2014/testing.slide)
- [Mitchell Hashimoto - Advanced Testing with Go](https://www.youtube.com/watch?v=yszygk1cpEc)
- [Ben Johnson - Structuring Tests in Go](https://medium.com/@benbjohnson/structuring-tests-in-go-46ddee7a25c#.q88391hne)
- [Dave Cheney - Test Fixtures in Go](https://dave.cheney.net/2016/05/10/test-fixtures-in-go)
- [Peter Bourgon - Go: Best Practices for Production Environments](https://peter.bourgon.org/go-in-production/)

在阅读这篇之前，我希望你已经知道如何做表格驱动的测试以及使用 interface 进行 模拟 (mock)/ 桩 (stub) 注入。这里是一些窍门：

## 窍门 1. 不使用框架

来自 Ben Johnson 的窍门。Go 有一个着实很棒的测试框架，它让你能够使用同样的语言编写测试代码，无需学习任何的库或测试引擎，直接用！ 也可以参考 Ben Johnson 的[帮助函数](https://github.com/benbjohnson/testing)，可以帮你省一些代码行数。

## 窍门 2. 使用 "_test" 包名

> Ben Johnson's tip. Using `*_test` package doesn't allow you to enter unexported identifiers. This puts you into position of a package's user, allowing you to check whether package's public API is useful.

来自 Ben Johnson 的提示。使用 `*_test` 包名，你无法使用包中未导出的标识符。这样在测试中，你就是包的使用者，让你更好地检查包的公有 API 是否可用。

## 窍门 3. 避免全局常量

来自 Mitchell Hashimoto 的窍门。如果你使用了全局的常量标识符，测试时将无法改变它们的行为。这个窍门的例外是全局常量对默认值是有用的。查看下面的例子：

```go
// Bad, tests cannot change value!
const port = 8080
// Better, tests can change the value.
var port = 8080
// Even better, tests can configure Port via struct.
const defaultPort = 8080
type AppConfig {
	Port int // set it to defaultPort using constructor.
}
```

这里是一些技巧，希望能够让你的测试代码更好：

## 技巧 1. Test fixtures

这个技巧在[标准库](https://golang.org/src/cmd/gofmt/testdata/) 中用到。这是我从 Mitchell Hashimoto 和 Dave Cheney 的作品中学到的。go test 很好地支持从文件中加载测试数据。第一，go build 忽略名为 **testdata** 的文件夹。第二，当 ge test 运行时，它将当前目录设置为包目录。这使得你可以使用相对路径 **testdata** 目录作为存储和加载数据的地方。这是一个例子：

```go
func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}
```

## 技巧 2. Golden 文件

这个技巧也在[标准库](https://golang.org/src/cmd/gofmt/gofmt_test.go) 中被用到，但我是从 Mitchell Hashimoto 的演讲中学到的。这里的思想是将期望输出存储在一个名为 **.golden** 的文件中并提供一个 flag 来更新它。这里是例子：

```go
var update = flag.Bool("update", false, "update .golden files")
func TestSomething(t *testing.T) {
	actual := doSomething()
	Golden := filepath.Join("testdata", tc.Name+ ".golden" )
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	expected, _ := ioutil.ReadFile(golden)

	if !bytes.Equal(actual, expected) {
		// FAIL!
	}
}
```

这个技巧使你得以测试复杂的输出而无需硬编码。

## 技巧 3. 测试帮助函数

Mitchell Hashimoto 的技巧。有时候测试代码变得有点复杂。当你需要为测试案例做合适的配置，经常包含很多无关的 err 检查，例如检查测试文件是否正确加载，检查数据是否可以解析成 JSON，等等。这些代码很快就变得十分丑陋！

为了解决这个问题，你需要将无关的代码分离到帮助函数里。这些函数永远不应该返回 error，如果有一些操作失败，应该将 `*testing.T` 作为参数并让测试失败。

还有，如果你的帮助函数需要在后面做一些清理，你应该返回一个做清理工作的函数。看看下面的例子：

```go
func testChdir(t *testing.T, dir string) func() {
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("err: %s", err)
	}
	return func() {
		if err := os.Chdir(old); err != nil {
			 t.Fatalf("err: %s", err)
		}
	}
}
func TestThing(t *testing.T) {
	defer testChdir(t, "/other")()
	// ...
}
```

( 注：这个例子摘自 [Mitchell Hashimoto - Advanced Testing with Go](https://www.youtube.com/watch?v=yszygk1cpEc) 的演讲 )。这个例子里用到了另一个很酷的技巧就是 `defer`。在这段代码中 `defer testChdir(t, “ /other")()` 执行 testChdir 函数并将其返回的清理函数延迟执行。

## 技巧 4. 子进程：真实调用

有时候你需要测试依赖于可执行程序的代码。例如，你的程序使用 Git。测试这段代码的一个办法是模拟 Git 的行为，但那真的很难！另一个办法是真正调用 Git 可执行程序。但如果执行测试的用户没有安装 Git 怎么办？

这个技巧解决了检查系统是否有安装 Git 的问题，如果没有安装则跳过测试。例子如下：

```go
var testHasGit bool
func INIt() {
	if _, err := exec.LookPath("git"); err == nil {
		testHasGit = true
	}
}
func TestGitGetter(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}
	// ...
}
```

( 注：这个例子摘自 [Mitchell Hashimoto - Advanced Testing with Go](https://www.youtube.com/watch?v=yszygk1cpEc) 的演讲。)

## 技巧 5. 子进程：模拟

Andrew Gerrand / Mitchell Hashimoto 的技巧。下面的技巧让我们模拟一个子进程，无需跳过测试代码。这个例子也可以在 [标准库测试](https://golang.org/src/os/exec/exec_test.go) 中看到。假设我们要测试 Git 失败的场景。看看这个例子：

```go
func CrashingGit() {
	os.Exit(1)
}
func TestFailingGit(t *testing.T) {
	if os.Getenv("BE_CRASHING_GIT") == "1" {
		CrashingGit()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFailingGit")
	cmd.Env = append(os.Environ(), "BE_CRASHING_GIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("Process ran with err %v, want os.Exit(1)", err)
}
```

这里的思想把 Go 测试框架作为稍做修改（`os.Args[0]`- 是生成的 Go test 二进制）的子进程运行。稍做修改是为了在环境变量为 `BE_CRASHING_GIT=1` 时运行同样的测试（`-test.run=TestFailingGit` 的部分），这样你可以区分何时作为子进程运行，何时正常执行。

## 技巧 6. 将模拟、帮助函数放在 testing.go 文件中

Hashimoto 提的一个有趣的建议是将帮助函数，fixtures，桩都导出并放在 **testing.go** 文件中。（注意 **testing.go** 文件被当成正常的代码对待，而不是测试代码。）这使你可以在不同的包中使用模拟和帮助函数，包的使用者在他们的代码中也可以使用它们。

## 技巧 7. 处理那些运行慢的测试

Peter Bourgon 的技巧。当你有一些运行很慢的测试时，等待所有测试完成会变得很烦人，特别是当你想立刻知道编译是否成功时。这个问题的解决办法是将那些运行慢的测试移到 `*_integration_test.go` 文件中并在文件的开头添加编译选项。例如：

```go
// +build integration
```

这样 `go test` 就不会包含有编译选项的那些测试。为了执行它们，你需要在 `go test` 命令中指定编译选项。

```
go test -tags=integration
```

我个人使用 alias，用于运行当前包以及子包里除 vendor 目录以外的所有测试。

```
alias gtest="go test \$(go list ./ … | grep -v /vendor/)
-tags=integration"
```

这个 alias 兼容 -v 选项。

```
$ gtest
…
$ gtest -v
…
```

感谢阅读！如果你有任何问题或想要提供反馈，可以在我的 blog https://povilasv.me 找我或者通过 Twitter [@PofkeVe](https://twitter.com/Pofkeve) 跟我联系。

[Golang](https://medium.com/tag/golang?source=post)
[Programming](https://medium.com/tag/programming?source=post)
[Unit Testing](https://medium.com/tag/unit-testing?source=post)
[Testing](https://medium.com/tag/testing?source=post)

---

via: https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859

作者：[Povilas Versockas](https://medium.com/@povilasve)
译者：[krystollia](https://github.com/krystollia)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
