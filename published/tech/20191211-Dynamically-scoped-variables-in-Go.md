首发于：https://studygolang.com/articles/27147

# Go 中的动态作用域变量

这是一个 API 设计的思想实验，它从典型的 Go 单元测试惯用形式开始：

```go
func TestOpenFile(t *testing.T) {
        f, err := os.Open("notfound")
        if err != nil {
                t.Fatal(err)
        }

        // ...
}
```

这段代码有什么问题？断言 `if err != nil { ... }` 是重复的，并且需要检查多个条件的情况下，如果测试的作者使用 `t.Error` 而不是 `t.Fatal` 的话会容易出错，例如：

```go
f, err := os.Open("notfound")
        if err != nil {
                t.Error(err)
        }
        f.Close() // boom!
```

有什么解决方案？当然，通过将重复的断言逻辑移到辅助函数中，来达到 DRY（Don't Repeat Yourself）。

```go
func TestOpenFile(t *testing.T) {
        f, err := os.Open("notfound")
        check(t, err)

        // ...
}

func check(t *testing.T, err error) {
       if err != nil {
                t.Helper()
                t.Fatal(err)
        }
}
```

使用 `check` 辅助函数使得这段代码更简洁一些，并且更加清晰地检查错误，同时有望解决 `t.Error` 与 `t.Fatal` 的混淆使用。
将断言抽象为一个辅助函数的缺点是，现在你需要将一个 `testing.T` 传递到每一个调用上。更糟糕的是，为了以防万一，你需要传递 `*testing.T` 到每一个需要调用 `check` 的地方。

我猜，这并没有关系。但我会观察到只有在断言失败的时候才会用到变量 t —— 即使在测试场景下，大多数时候，大部分的测试是通过的，因此在相对罕见的测试失败的情况下，会产生对这些变量 t 的固定读写开销。

如果我们这样做怎么样？

```go
func TestOpenFile(t *testing.T) {
        f, err := os.Open("notfound")
        check(err)

        // ...
}

func check(err error) {
        if err != nil {
                panic(err.Error())
        }
}
```

是的，可以，但是有一些问题。

```
% go test
--- FAIL: TestOpenFile (0.00s)
panic: open notfound: no such file or directory [recovered]
        panic: open notfound: no such file or directory

goroutine 22 [running]:
testing.tRunner.func1(0xc0000b4400)
        /Users/dfc/go/src/testing/testing.go:874 +0x3a3
panic(0x111b040, 0xc0000866f0)
        /Users/dfc/go/src/runtime/panic.go:679 +0x1b2
github.com/pkg/expect_test.check(...)
        /Users/dfc/src/github.com/pkg/expect/expect_test.go:18
github.com/pkg/expect_test.TestOpenFile(0xc0000b4400)
        /Users/dfc/src/github.com/pkg/expect/expect_test.go:10 +0xa1
testing.tRunner(0xc0000b4400, 0x115ac90)
        /Users/dfc/go/src/testing/testing.go:909 +0xc9
created by testing.(*T).Run
        /Users/dfc/go/src/testing/testing.go:960 +0x350
exit status 2
```

先从好的方面说起，我们不需要传递一个 `testing.T` 到每一个调用 `check` 函数的地方，且测试会立即失败。我们还从 panic 中获得了一条不错的信息 —— 尽管重复出现了两次。但是，哪里断言失败却不容易看到。它发生在 `expect_test.go:11`，你知道这一点是不可以原谅的。

所以 panic 不是一个好的解决办法，但是你能从堆栈跟踪信息里面看到什么有用的信息吗？这有一个提示：`github.com/pkg/expect_test.TestOpenFile(0xc0000b4400)`。

TestOpenFile 有一个 t 的值，它由 tRunner 传递过来，所以 testing.T 在内存中位于地址 0xc0000b4400 上。如果我们可以在 check 函数内部获取 t 会怎样？那我们可以通过它来调用 t.Helper 来 t.Fatal。这可能吗？

## 动态作用域
我们想要的是能够访问一个变量，而该变量的申明既不是在全局范围，也不是在函数局部范围，而是在调用堆栈的更高的位置上。这被称之为*动态作用域*。Go 并不支持动态作用域，但事实证明，某些情况下，我们可以模拟它。回到正题：

```go
// getT 返回由 testing.tRunner 传递过来的 testing.T 地址
// 而调用 getT 的函数由它（tRunner）所调用. 如果在堆栈中无法找到 testing.tRunner
// 说明 getT 在主测试 goroutine 没有被调用，
// 这时 getT 返回 nil.
func getT() *testing.T {
        var buf [8192]byte
        n := runtime.Stack(buf[:], false)
        sc := bufio.NewScanner(bytes.NewReader(buf[:n]))
        for sc.Scan() {
                var p uintptr
                n, _ := fmt.Sscanf(sc.Text(), "testing.tRunner(%v", &p)
                if n != 1 {
                        continue
                }
                return (*testing.T)(unsafe.Pointer(p))
        }
        return nil
}
```

我们知道每个测试（Test)由 testing 包在自己的 goroutine 上调用（看上面的堆栈信息）。testing 包通过一个名为 tRunner 的函数来启动测试，该函数需要一个*testing.T 和一个 func(*testing.T)来调用。因此我们抓取当前 goroutine 的堆栈信息，从中扫描找到已 testing.tRunner 开头的行——由于 tRunner 是私有函数，只能是 testing 包——并解析第一个参数的地址，该地址是一个指向 testing.T 的指针。有点不安全，我们将这个原始指针转换为一个 *testing.T 我们就完成了。

如果搜索不到则可能是 getT 并不是被 Test 所调用。这实际上是行的通的，因为我们需要*testing.T 是为了调用 t.Fatal，而 testing 包要求 t.Fatal 被[主测试 goroutine](https://golang.org/pkg/testing/#T.FailNow)所调用。

```go
import "github.com/pkg/expect"

func TestOpenFile(t *testing.T) {
        f, err := os.Open("notfound")
        expect.Nil(err)

        // ...
}
```

综上，在预期打开文件所产生的 err 为 nil 后，我们消除了断言样板，并且是测试看起来更加清晰易读。

## 这样好吗？

这时你应该会问，*这样好吗？*答案是，不，这不好。此时你应该会感到震惊，但是这些不好的感觉可能值得反思。除了在 goroutine 的调用堆栈乱窜的固有不足以外，同样存在一些严重的设计问题：
1.  expect.Nil 的行为依赖于谁调用它。同样的参数，由于调用堆栈位置的原因可能导致行为的不同——这是不可预期的。
2.  采取极端的动态作用域，将传递给单个函数之前的所有函数的所有变量纳入单个函数的作用域中。这是一个在函数申明没有明确记录的情况下将数据传入和传出的辅助手段。

讽刺的是，这恰恰是我对[context.Context](https://dave.cheney.net/2017/01/26/context-is-for-cancelation)的评价。我会将这个问题留给你自己判断是否合理。

## 最后的话

这是个坏主意，这点没有异议。这不是你可以在生产模式中使用的模式。但是，这也不是生产代码。这是在测试，也许有着不同的规则适用于测试代码。毕竟，我们使用模拟（mocks）、桩（stubs）、猴子补丁（monkey patching）、类型断言、反射、辅助函数、构建标志以及全局变量，所有这些使得我们更加有效率得测试代码。所有这些，*奇技淫巧*是不会让它们出现在生产代码里面的，所以这真的是世界末日吗？

如果你读完本文，你也许会同意我的观点，尽管不太符合常规，并无必要将*testing.T 传递到所有需要断言的函数中去，从而使测试代码更加清晰。

如果你感兴趣，我已分享了一个应用这个模式的[小的断言库](https://github.com/pkg/expect)。小心使用。

---

via: https://dave.cheney.net/2019/12/08/dynamically-scoped-variables-in-go

作者：[Dave Cheney](https://dave.cheney.net/)
译者：[dust347](https://github.com/dust347)
校对：[@unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
