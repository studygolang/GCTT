# Go中的动态作用域变量

这是一个API设计的思想实验。它从典型的Go单元测试惯用形式开始：

```go
func TestOpenFile(t *testing.T) {
        f, err := os.Open("notfound")
        if err != nil {
                t.Fatal(err)
        }

        // ...
}
```

这段代码有什么问题？断言if err != nil { ... } 是重复的，并且需要检查多个条件的情况下，如果测试的作者使用t.Error而不是t.Fatal的话会容易出错，例如：

```go
f, err := os.Open("notfound")
        if err != nil {
                t.Error(err)
        }
        f.Close() // boom!
```

有什么解决方案？当然，通过将重复的断言逻辑移到辅助函数中，来达到DRY（Don't Repeat Yourself）。

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

使用check辅助函数使得这段代码更简洁一些，并且更加清晰地检查错误，同时有望解决t.Error与t.Fatal的混淆使用。
将断言抽象为一个辅助函数的缺点是，现在你需要将一个testing.T传递到每一个调用上。更糟糕的是，为了以防万一，你需要传递*testing.T到每一个需要调用check的地方。

我猜，这并没有关系。但我会观察到只有在断言失败的时候才会用到变量t——即使在测试场景下，大部分时候，大部分的测试是通过的，因此在相对罕见的测试失败的情况下，这些变量t对阅读和书写都是固定开销。

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

先从好的地方开始，我们不需要传递一个testing.T到每一个调用check函数的地方，测试立即失败，我们从panic中获得了一条不错的信息——尽管有两次。但是哪里断言失败却不容易看到。它发生在expect_test.go:11，你不知道这一点是可以原谅的。

所以panic不是一个好的解决办法，但是在堆栈跟踪信息里面有什么——你能看到吗？这有一个提示，github.com/pkg/expect_test.TestOpenFile(0xc0000b4400)。

TestOpenFile有一个t的值，它由tRunner传递过来，所以testing.T在内存中位于地址 0xc0000b4400上。如果我们可以在check函数内部获取t会怎样？那我们可以通过它来调用t.Helper来t.Fatal。这可能吗？

### 动态作用域
我们想要的是能够访问一个变量，而该变量的申明既不是在全局范围，也不是在函数局部范围，而是在调用堆栈的更高的位置上。这被称之为*动态作用域*。Go并不支持动态作用域，但事实证明，某些情况下，我们可以模拟它。回到正题：

```go
// getT 返回由testing.tRunner传递过来的testing.T地址
// 而调用getT的函数由它（tRunner）所调用. 如果在堆栈中无法找到testing.tRunner
// 说明 getT在主测试goroutine没有被调用，
// 这时getT返回nil.
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

我们知道每个测试（Test)由testing包在自己的goroutine上调用（看上面的堆栈信息）。testing包通过一个名为tRunner的函数来启动测试，该函数需要一个*testing.T和一个func(*testing.T)来调用。因此我们抓取当前goroutine的堆栈信息，从中扫描找到已testing.tRunner开头的行——由于tRunner是私有函数，只能是testing包——并解析第一个参数的地址，该地址是一个指向testing.T的指针。有点不安全，我们将这个原始指针转换为一个 *testing.T我们就完成了。

如果搜索不到则可能是getT并不是被Test所调用。这实际上是行的通的，因为我们需要*testing.T是为了调用t.Fatal，而testing包要求t.Fatal被[主测试goroutine](https://golang.org/pkg/testing/#T.FailNow)所调用。

```go
import "github.com/pkg/expect"

func TestOpenFile(t *testing.T) {
        f, err := os.Open("notfound")
        expect.Nil(err)

        // ...
}
```

综上，在预期打开文件所产生的err为nil后，我们消除了断言样板，并且是测试看起来更加清晰易读。

### 这样好吗？
这时你应该会问，*这样好吗？*答案是，不，这不好。此时你应该会感到震惊，但是这些不好的感觉可能值得反思。除了在goroutine的调用堆栈乱窜的固有不足以外，同样存在一些严重的设计问题：
1.  expect.Nil的行为依赖于谁调用它。同样的参数，由于调用堆栈位置的原因可能导致行为的不同——这是不可预期的。
2.  采取极端的动态作用域，将传递给单个函数之前的所有函数的所有变量纳入单个函数的作用域中。这是一个在函数申明没有明确记录的情况下将数据传入和传出的辅助手段。

讽刺的是，这恰恰是我对[context.Context](https://dave.cheney.net/2017/01/26/context-is-for-cancelation)的评价。我会将这个问题留给你自己判断是否合理。

### 最后的话
这是个坏主意，这点没有异议。这不是你可以在生产模式中使用的模式。但是，这也不是生产代码。这是在测试，也许有着不同的规则适用于测试代码。毕竟，我们使用模拟（mocks）,桩（stubs），猴子补丁（monkey patching），类型断言，反射，辅助函数，构建标志以及全局变量，所有这些使得我们更加有效率得测试代码。所有这些，嗯，*黑客*是不会让它们出现在生产代码里面的，所以这真的是世界末日吗。

如果你读完本文，你也许会同意我的观点，尽管不太符合常规，并无必要将*testing.T传递到所有需要断言的函数中去，从而使测试代码更加清晰。

如果你感兴趣，我已分享了一个应用这个模式的[小的断言库](https://github.com/pkg/expect)。小心使用。

---

via: https://dave.cheney.net/2019/12/08/dynamically-scoped-variables-in-go

作者：[Dave Cheney](https://dave.cheney.net/)
译者：[dust347](https://github.com/dust347)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
