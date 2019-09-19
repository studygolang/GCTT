首发于：https://studygolang.com/articles/23498

# Vet 命令：超出预期的强大

!["Golang 之旅"插图，由 Go Gopher 的 Renee French 创作](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-vet-command-is-more-powerful-than-you-think/go-vet.png)

Go `vet` 命令在编写代码时非常有用。它可以帮助您检测应用程序中任何可疑、异常或无用的代码。该命令实际上由几个子分析器组成，甚至可以与您的自定义分析器一起工作。让我们首先回顾一下内置的分析器。

## 内置分析器

可以通过命令 `go tool vet help` 获取 [内置分析器](https://golang.org/cmd/vet/) 列表。让我们分析一些不太明显的例子，以便更好地理解。

### Atomic

这个分析器将防止原子函数的任何不正确使用

```go
func main() {
   var a int32 = 0

   var wg sync.WaitGroup
   for i := 0; i < 500; i++ {
      wg.Add(1)
      go func() {
         a = atomic.AddInt32(&a, 1) // 改为 atomic.AddInt32(&a, 1) 即可
         wg.Done()
      }()
   }
   wg.Wait()
}
main.go:15:4: direct assignment to atomic value
```

由于原子内存原语函数 `addInt` 是并发安全的，所以变量 `a` 会安全增加。但是，我们将结果分配给相同的变量，这不是一个并发安全的写操作。`atomic` 分析仪将发现发现这个个粗心的错误。

### copylocks

正如文档中所描述的，永远不应该复制锁。实际上，它在内部管理锁的当前状态。一旦使用了锁，此锁的副本就会复制其内部状态，使锁的副本与原始状态相同，而不是初始化的新状态。

```go
func main() {
   var lock sync.Mutex

   l := lock //直接使用 lock 即可
   l.Lock()
   l.Unlock()
}
// from vet: main.go:9:7: assignment copies lock value to l: sync.Mutex
```

使用锁的结构体应该使用指针引用，以保持内部状态一致：

```go
type Foo struct {
   lock sync.Mutex
}

func (f Foo) Lock() { // 改为：func (f *Foo) Lock()
   f.lock.Lock()
}

func main() {
   f := Foo{lock: sync.Mutex{}}
   f.Lock()
}
// from vet: main.go:9:9: Lock passes lock by value: command-line-arguments.Foo contains sync.Mutex
```

### loopclosure

当您启动一个新的 goroutine 时，主 goroutine 将继续执行。在执行时，将进行评估 goroutine 及其变量的代码将，当一个变量仍然被主 goroutine 更新时使用，这可能会导致一些常见的错误：

```go
func main() {
   var wg sync.WaitGroup
   for _, v := range []int{0,1,2,3} { // 需引入临时变量解决,或 通过传值参数解决
      wg.Add(1)
      go func() {
         print(v)
         wg.Done()
      }()
   }
   wg.Wait()
}
// output:
// 3333
// from vet: main.go:10:12: loop variable v captured by func literal
```

### lostcancel

从主上下文（main）创建一个可取消的上下文（cancellable context）将返回新上下文以及一个能够取消该上下文的函数。此函数可在任何时候用于取消与此上下文关联的所有操作，但应始终调用此函数，以避免泄漏任何上下文。

```go
func Foo(ctx context.Context) {}

func main() {
   ctx, _ := context.WithCancel(context.Background())
   Foo(ctx)
}
// from vet: main.go:8:7: the cancel function returned by context.WithCancel should be called, not discarded, to avoid a context leak
// 需改为：
    // ctx, cancleFunc := context.WithCancel(context.Background())
    // Foo(ctx)
    // cancleFunc()
```

如果需要了解关于 context 的更多细节、各种 context 的差异以及 cancel function 的功能，我建议您阅读我关于[上下文和通过传播进行取消](https://medium.com/@blanchon.vincent/go-context-and-cancellation-by-propagation-7a808bbc889c)的文章。

### stdmethods

`stdmethods` 分析器将确保你已经从标准库的接口来实现的方法是与标准库兼容：

```go
type Foo struct {}

func (f Foo) MarshalJSON() (string, error) {
   return `{a: 0}`, nil
}
// 需改为：
// func (f Foo) MarshalJSON() ([]byte, error) {
//    return []byte(`{a: 0}`), nil
// }

func main() {
   f := Foo{}
   j, _ := json.Marshal(f)
   println(string(j))
}
// {}
// from vet: main.go:7:14: method MarshalJSON() (string, error) should have signature MarshalJSON() ([]byte, error)
```

### structtag

标签是结构中的字符串，应该遵循[反射包中的约定](http://golang.org/pkg/reflect/#StructTag)。随意使用将使标签无效，并可能很难调试没有审查命令:一个多余的空格都会使 tag 失效，如果没有 `vet` 命令其将难以调试

```go
type Foo struct {
   A int `json: "foo"`// 去除 `json: "foo" 中间多余空格即可
}

func main() {
   f := Foo{}
   j, _ := json.Marshal(f)
   println(string(j))
}
// {"A":0}
// from vet: main.go:6:2: struct field tag `json: "foo"` not compatible with reflect.StructTag.Get: bad syntax for struct tag value
```

`vet` 命令还有[更多可用的分析器](https://github.com/golang/tools/blob/release-branch.go1.12/go/analysis/cmd/vet/vet.go#L51-L73),但这还不是这个命令的强大所在。它还允许我们自定义分析器。

## 自定义分析器

虽然内置分析器很有用，很强大，但是 Go 允许我们创建我们自己的分析器。

我将使用我构建的自定义分析器来检测上下文包在函数参数中的使用情况，您可以在“[构建自己的分析器](https://medium.com/@blanchon.vincent/go-how-to-build-your-own-analyzer-f6d83315586f)”一文中找到相关信息。

你的分析器一旦构建完成，就可通过 `vet` 命令直接使用。

```sh
go install github.com/blanchonvincent/ctxarg
go vet -vettool=$(which ctxarg)
```

您甚至可以构建自己的分析工具。

## 自定义分析命令

由于分析器与命令完全解耦，您可以使用您需要的分析程序创建您自己的命令。让我们来看一个自定义命令的例子，它只使用我们需要的一些分析器:

基于分析器列表的自定义命令

```go
package main

import (
    "golang.org/x/tools/go/analysis/multichecker"
    "golang.org/x/tools/go/analysis/passes/atomic"
    "golang.org/x/tools/go/analysis/passes/loopclosure"
    "github.com/blanchonvincent/ctxarg/analysis/passes/ctxarg"
)

func main() {
    multichecker.Main(
        atomic.Analyzer,
        loopclosure.Analyzer,
        ctxarg.Analyzer,
    )
}
```

构建并运行该命令将为我们提供一个基于所选分析程序的工具:

```sh
./custom-vet help
custom-vet is a tool for static analysis of Go programs.

Registered analyzers:

    atomic       check for common mistakes using the sync/atomic package
    ctxarg       check for parameters order while receiving context as parameter
    loopclosure  check references to loop variables from within nested functions
```

您还可以创建您的自定义分析程序集，并将它们与您喜欢的内置分析程序合并，得到一个适合您自己的工作流和公司编码标准的自定义命令。

---

via: <https://medium.com/a-journey-with-go/go-vet-command-is-more-powerful-than-you-think-563e9fdec2f5>

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[TomatoAres](https://github.com/TomatoAres)
校对：[DingdingZhou](https://github.com/DingdingZhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
