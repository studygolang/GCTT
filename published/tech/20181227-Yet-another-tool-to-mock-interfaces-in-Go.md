首发于：https://studygolang.com/articles/17946

# Go 语言中一个模拟接口的工具

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/mock-interface/1_OC_uFaDoGfZ7s1Pkg8YbGg.png)

单元测试作为一种强大的工具，可以检查代码各个方面的行为。如果对进行代码测试十分重视，那么您将会一直编写可持续、可维护的代码，并且在代码的实现过程中保持代码的完整性。依赖于抽象的、经过开发者精心设计的代码是很容易进行测试的，所以代码的可测试性也作为其质量的一个指标。

如果您已经在 Go 中尝试过测试代码，您可能知道接口的巨大作用。在 Go 的标准库中，提供了一系列接口，这些接口大多数只包含一个方法，您可以使用这些接口。

Go 还有一个补充框架，用以模拟接口。同时，还有一些社区驱动的包可以完成类似的功能。他们中的大多数都可以根据给定接口，生成实现这些接口的 `struct`。对于较大的接口，或者嵌套了其他接口，使用这种方式很有效。当接口只有一个方法时，不是更有效果吗？

关于 Go 中的接口，最令人惊讶的部分是它的默认满足性。任何类型，只需要提供其签名与接口声明中的方法匹配的实现，即可以满足该接口。这种类型甚至可以是函数，如果您熟悉 `net/http` 包，你也可能看到其中的一种可以叫做 `adapters` 的类型。

```go
// A Handler responds to an HTTP request.
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}
```

如上述代码，`adapters` 类型本身是一个函数类型，而且具有与接口方法声明相同的签名，它通过在对应方法中调用自身，实现了接口。这个适配器允许具有适当签名的任何函数来实现 `Handler` 。它作为一种模拟接口的通用工具，看起来在表驱动的测试中十分有用。例如，需要测试以下代码:

```go
package execute

import (
	"errors"
	"fmt"
	"io"
)

var (
	// ErrExpectedError return when error is expected.
	ErrExpectedError = errors.New("expected error")
)

// Doer does some job.
type Doer interface {
	Do() (jobID int, err error)
}

// Execute executes Doer interface and handles errors.
func Execute(job Doer, w io.Writer) {
	id, err := job.Do()
	if err != nil {
		switch err {
		case ErrExpectedError:
			fmt.Fprintf(w, "%v\n", err)
		default:
			fmt.Fprintf(w, "unexpected error: %v\n", err)
		}

		return
	}

	fmt.Fprintf(w, "job %d done\n", id)
}
```

使用 `adapter` 的单元测试如下：

```go
package execute

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Execute(t *testing.T) {
	tests := []struct {
		name    string
		doFunc  func() (int, error)
		wantErr bool
		expect  string
	}{
		{
			name: "expected error",
			doFunc: func() (int, error) {
				return 0, ErrExpectedError
			},
			expect: "expected error\n",
		},
		{
			name: "unexpected error",
			doFunc: func() (int, error) {
				return 0, errors.New("mocked error")
			},
			expect: "unexpected error: mocked error\n",
		},
		{
			name: "job done",
			doFunc: func() (int, error) {
				return 333, nil
			},
			expect: "job 333 done\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			Execute(doerFunc(tt.doFunc), &buf)

			assert.Equal(t, tt.expect, buf.String())
		})
	}
}

type doerFunc func() (int, error)

func (f doerFunc) Do() (int, error) {
	return f()
}
```

编写这样的 `adapters` 十分繁杂，所以我决定编写一个工具来生成代码，叫做 [adapt](https://github.com/romanyx/adapt) 。使用这个工具可以对一个指定接口生成 `adapters` ，并且打印其输出。你所需要做的工作就是，传入一个包名和接口名来生成代码。

```shell
$ adapt io Reader
type readerFunc func([]byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) {
	return f(p)
}
```

也可以在包中的文件夹内部，使用 `adapt` 工具为包中的一些接口生成适配器。

```shell
$ cd $GOPATH/src/github.com/x/execute Doer
$ adapt Doer
type doerFunc func() (int, error)

func (f doerFunc) Do() (int, error) {
	return f()
}
```

也可以和一个便捷的 [vim 插件](https://github.com/romanyx/vim-go-adapt) 配合使用，可以再 vim 中直接调用该工具。

![use in vim](https://raw.githubusercontent.com/studygolang/gctt-images/master/mock-interface/1_PCMcTGnUNvjP0hooLXYBOw.gif)

希望您会发现它很有用！

---

via: https://itnext.io/yet-another-tool-to-mock-interfaces-in-go-73de1b02c041

作者：[Roman Budnikov](https://itnext.io/@romanyx90)
译者：[Inno Jia](https://github.com/kobeHub)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
