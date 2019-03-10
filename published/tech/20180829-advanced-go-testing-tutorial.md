首发于：https://studygolang.com/articles/18797

# Go 语言测试进阶教程

你好啊各位码农们！在这个教程中，我们将介绍一些进阶的测试实践，很多 Go 语言的核心开发人员以及流行的生产级工具都使用到了它们。

我希望这种通过生产上真实使用的案例来讲解的方法。能够给你一些启示，让你深入了解怎么去测试你自己的生产级别的 Go 程序。

> 注意：如果你对如何测试 Go 语言的程序完全不了解的话，我建议你先看看之前的教程：
> [an introduction to testing in Go](https://tutorialedge.net/golang/intro-testing-in-go/)
> （译注：Go 语言中文网译文：[Go 测试介绍](https://studygolang.com/articles/16772)）

## 通过表格驱动的测试来实现良好的测试覆盖率

我们先来看看 `strings` 代码包，如果你看一眼 `src/strings/` 目录里面的 `strings_test.go` 文件，你会发现文件开头定义了一些数组。

举个例子，我们来看看 `lastIndexTests`，它是一个 `IndexTest` 类型的数组：

```go
var lastIndexTests = []IndexTest{
    {"", "", 0},
    {"", "a", -1},
    {"", "foo", -1},
    {"fo", "foo", -1},
    {"foo", "foo", 0},
    {"foo", "f", 0},
    {"oofofoofooo", "f", 7},
    {"oofofoofooo", "foo", 7},
    {"barfoobarfoo", "foo", 9},
    {"foo", "", 3},
    {"foo", "o", 2},
    {"abcABCabc", "A", 3},
    {"abcABCabc", "a", 6},
}
```

这个数组用来测试 `strings.go` 文件里面的 `LastIndex` 函数，里面有一系列正确的和错误的测试用例。
数组的每个元素由一个字符串、分隔符和一个 `out` 整数组成，它的结构如下：

```go
type IndexTest struct {
    s   string
    sep string
    out int
}
```

这些测试由 `TestLastIndex()` 函数触发，它会把所有这些测试用例都遍历一遍，检查 `lastIndex` 函数的返回值与数组中事先定义的期待的目标值是否一致。

同样的实践在无数的函数中使用过。这个实践能够帮助我们确保函数的代码发生改动时，函数的预期行为不会发生变化。

## 使用 testdata 目录

在某些情况下，你没办法像上述的例子一样，用数组的形式来指定你期待的测试输入与输出。比如说你想要测试在文件系统上读写文件，或想要测试解析某些特定格式的数据文件等等。

这个时候，你可以选择创建一个 `testdata` 目录，然后把你要用于测试的文件保存的那个目录中。

在标准库的 `src/archive/tar/` 目录里面有一个 `testdata` 目录。它包含了一些 `.tar` 文件，用来进行测试。

你可以在 `reader_test.go` 文件中看到一些较为复杂的例子：

```go
func TestReader(t *testing.T) {
    vectors := []struct {
        file    string    // Test input file
        headers []*Header // Expected output headers
        chksums []string  // MD5 checksum of files, leave as nil if not checked
        err     error     // Expected error to occur
    }{{
        file: "testdata/gnu.tar",
        headers: []*Header{{
            Name:     "small.txt",
            Mode:     0640,
            Uid:      73025,
            Gid:      5000,
            Size:     5,
            ModTime:  time.Unix(1244428340, 0),
            Typeflag: '0',
            Uname:    "dsymonds",
            Gname:    "eng",
            Format:   FormatGNU,
        }, {
            Name:     "small2.txt",
            Mode:     0640,
            Uid:      73025,
            Gid:      5000,
            Size:     11,
            ModTime:  time.Unix(1244436044, 0),
            Typeflag: '0',
            Uname:    "dsymonds",
            Gname:    "eng",
            Format:   FormatGNU,
        }},
        chksums: []string{
            "e38b27eaccb4391bdec553a7f3ae6b2f",
            "c65bd2e50a56a2138bf1716f2fd56fe9",
        },
  },
  // 更多的测试用例
```

上面的函数中你可以看到 Go 语言的核心开发者使用了我们一开始提到的表格驱动的测试方法以及我们在本节提到的方法。
他们把用于测试的 `.tar` 文件放在 `testdata` 目录里面，然后编写测试代码，确保这些 `.tar` 文件被解压出来以后，里面的文件和文件的校验和与预期的结果一致。

## Mock HTTP 请求

当你开始编写生产级别的 API 和服务的时候，你可能会需要与其他的服务进行交互。能按照你与这些服务的交互方式来进行测试，跟测试你本地的代码一样有必要。

但是，你的交互可能是一个会对数据库进行 CRUD（译注：指创建 - 查询 - 更新 - 删除）操作的 REST API，当你只是想测试一下这些操作能不够使用的时候，你肯定不会喜欢这些测试会真正地改动到你的数据库里面的数据。

所以，为了解决这个问题，我们可以使用 `net/http/httptest` 来 mock HTTP 的应答：

```go
package main_test

import (
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHttp(t *testing.T) {
  //
    handler := func(w http.ResponseWriter, r *http.Request) {
    // 我们在这里面编写预期的应答，通常情况下 REST API 会
    // 返回一个 JSON 字符串
        io.WriteString(w, "{ \"status\": \"expected service response\"}")
    }

    req := httptest.NewRequest("GET", "https://tutorialedge.net", nil)
    w := httptest.NewRecorder()
    handler(w, req)

    resp := w.Result()
    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Println(resp.StatusCode)
    fmt.Println(resp.Header.Get("Content-Type"))
    fmt.Println(string(body))
}
```

上面的测试用例中，我们用我们指定的响应内容，覆盖了我们请求的 URL 的原本的回复，然后根据我们自己 mock 的响应内容来继续测试我们程序的其它部分。

## 测试代码使用独立的代码包

查看 `strings_test.go` 文件的前面，你会发现它跟 `strings.go` 文件并不属于同一个代码包。

为什么要这样呢？它可以帮助你避免循环导入。在某些情况下，你需要在你的 `*_test.go` 文件里面导入一些代码包来方便你编写你的测试代码，但是如果你导入的这些包，在你准备测试的包中已经导入过了，就有可能会产生循环依赖。

## 把单元测试和集成测试区分开

> 注意：我是从这个文章学习到的这个技巧：[Go Advanced Tips Tricks](https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859)

如果你在为一个大型的企业级 Go 系统编写测试，那么你很有可能会有一系列的**单元**测试和**集成**测试来保证你系统的有效性。

但在通常情况下，你会发现集成测试运行的时间要比单元测试长很多。因为它们可能会接触到其它的系统。

在这个情况下，我们应该把集成测试的代码放到 `*_integration_test.go` 文件中，
并且这个文件的顶端添加 `// +build integration` 指令：

```go
// +build integration

package main_test

import (
    "fmt"
    "testing"
)

func TestMainIntegration(t *testing.T) {
    fmt.Println("My Integration Test")
}
```

这时如果你想要运行这个集成测试，你可以这样的使用 `go test` ：

```bash
$ Go test -tags=integration
My Integration Test
PASS
ok      _/Users/elliot/Documents/Projects/tutorials/golang/advanced-go-testing-tutorial 0.006s
```

## 结论

在这个教程中，我们讨论了一些被 Go 语言开发者所使用的进阶的测试技巧。

希望你能从中获得一些收获，并且对编写自己的 Go 测试代码能够有些深入的理解，如果你觉得它有用，或者有任何疑问，请不要犹豫，在评论区给我留言。

> 注意：如果你想了解更多新文章的咨询，请关注我的 Twitter [@Elliot_F](https://twitter.com/elliot_f)

## 延伸阅读

如果你觉得这个文章有点意思，你可以也会喜欢我讲解 Go 测试的另外一篇文章：

- [Improving Your Tests with Testify in Go](https://tutorialedge.net/golang/improving-your-tests-with-testify-go/)（译注：Go 语言中文网译文：[用 Testify 来改善 GO 测试和模拟](https://studygolang.com/articles/16799)）

---

via: https://tutorialedge.net/golang/advanced-go-testing-tutorial/

作者：[Elliot Forbes](https://tutorialedge.net/about/)
译者：[Alex-liutao](https://github.com/Alex-liutao)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
