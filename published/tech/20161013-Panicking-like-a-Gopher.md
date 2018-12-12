首发于：https://studygolang.com/articles/16572

# 用 gopher 的方式使用 panic

Go 运行时（即成功编译后，操作系统启动该该进程）发生的错误会以 panics 的形式反馈。*panic* 可以通过这两种形式触发 :

1.直接使用内置 *panic* 函数：

```go
package main

func main(){
    panic("foo")
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
panic: foo

goroutine 1 [running]:
panic(0x56d20, 0x820142000)
        /usr/local/go/src/runtime/panic.go:481 +0x3e6
main.main()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:4 +0x65
```

> panic 接受任何实现了空接口（interface{}）的参数类型，而所有类型都实现了空接口

2.由于程序问题，产生*运行时*的 *panic* ：

```go
package main

import "fmt"

func f() int {
    return 0
}

func main() {
    fmt.Println(1 / f())
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
panic: runtime error: integer divide by zero
[signal 0x8 code=0x7 addr=0x2062 pc=0x2062]

goroutine 1 [running]:
panic(0xda560, 0x8201c80b0)
        /usr/local/go/src/runtime/panic.go:481 +0x3e6
main.main()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:10 +0x22
```

> 触发运行时的 panic 和通过接口类型 [runtime.Error](https://golang.org/pkg/runtime/#Error) 的各种值调用 panic 函数在语义上是相同的。
>
> 第二种方式的输出有一些关于信号的额外信息。*0x8 (SIGFPE)*报告了致命的算术错误。

为了更好了解 Go 的 panicking 机制，让我们先深入了解一下 Go 语言程序中的几种结构：

## goroutines

用 Go 语言实现的程序在执行时或多或少都会有一些 *goroutine*。在关于 *go statement* 的规范详述中 *goroutine* 定义如下：

> 一个 "Go" 语句在同一个地址空间内作为一个独立的并发线程控制或者 goroutine ，开始执行函数调用。

Go 是一门并发语言，这是因为它（原生）提供了并发编程的特性，比如并发运行的语句 ([Go 语句](https://golang.org/ref/spec#Go_statements)）或能够在一些并发事物中轻松交流的机制（[channels](https://golang.org/ref/spec#Channel_types)）。 不过并发是什么意思呢？和无处不在的并行又有什么关系呢？

### 组成

并发是将一些东西作为一系列独立的任务来构造的方式。它是一种结构（设计）。并发程序处理相互独立的任务，也就是说，不关心它们的执行顺序。并行是两个或多个任务的同时执行。从字面上来理解，在同一时间很多执行线程正在进行中。它需要在多核机器上执行——没有方式去「模仿」它 。

并发是比并行更为普遍的一个术语。并发程序的执行时间可以（但并不必要）更好利用多核，并且同时执行多个运算。如果只有一个核是可用的，那么就要用到，比如时间切片（将时间分成一些分散的间隔，并且将它们分配给不同的任务），它仍然是并发的，但是并行在技术上是不可能的。

> 并发是同时处理很多事情，并行是同时进行很多事情。
>
> Rob Pike

### 主 goroutine

从 *main* 包（用 Go 写的每一个程序的入口点）中运行 *main* 函数。如果这个 Goroutine 在程序执行时结束，整个程序就终止了。运行时不会等待其他 Goroutine 结束。

程序从单个（主）goroutine 开始，在其生命周期中可以创造一些新的 goroutine（创造出有成千上万个并不少见）。

### goroutines 共用同一地址

```go
package main

import "fmt"

func main() {
    i := 0
    ch := make(chan int)
    Go func() {
        i++
        ch <- i
    }()
    Go func() {
        i++
        ch <- i
    }()
    fmt.Println(<-ch)
    fmt.Println(<-ch)
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
1
2
```

## Defer 语句

defer 可以让函数在包含它的函数（使用 defer 的函数）结束后执行，方式如下：

1.在 *return* 语句：

```go
package main

import "fmt"

func f() int {
    fmt.Println("1")
    defer func() {
        fmt.Println("Inside deferred function")
    }()
    fmt.Println("2")
    return 1
}

func main() {
    f()
}
```

```

> go install github.com/mlowicki/lab && ./bin/lab
1
2
Inside deferred function
```

2.当执行到函数末尾：

```go
package main

import "fmt"

func main() {
    fmt.Println("1")
    defer func() {
        fmt.Println("Inside deferred function")
    }()
    fmt.Println("2")
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
1
2
Inside deferred function
```

3.在 *panicking* 中：

```go
package main

import "fmt"

func f() {
    fmt.Println("1")
    defer func() {
        fmt.Println("Inside deferred function")
    }()
    fmt.Println("2")
    panic("boom!")
    fmt.Println("3")
}

func main() {
    f()
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
1
2
Inside deferred function
panic: boom!

goroutine 1 [running]:
panic(0xb8260, 0x8202301f0)
        /usr/local/go/src/runtime/panic.go:481 +0x3e6
main.f()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:11 +0x20c
main.main()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:16 +0x14
```

函数值和传递的参数在 *defer* 语句处进行评定，而非在真正调用的时候发生：

```go
package main

import "fmt"

func main() {
    f := func(n int) {
        fmt.Printf("Inside f, n=%d\n", n)
    }
    n := 1
    defer f(n)
    f = func(int) {
        fmt.Println("Inside g")
    }
    n = 2
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
Inside f, n=1
```

每个函数中可以有多个 *defer* 语句。调用顺序是后进先出的（就像 defer 的调用会被放入栈中）：

```go
package main

import "fmt"

func main() {
    defer func() {
        fmt.Println("1")
    }()
    defer func() {
        fmt.Println("2")
    }()
    defer func() {
        fmt.Println("3")
    }()
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
3
2
1
```

这样的方法调用也有效：

```go
package main

import "fmt"

type T struct{}

func (t T) m() {
    fmt.Println("Inside method")
}

func main() {
    t := T{}
    defer t.m()
}
```

并且它会像你可能期待的那样输出“内部方法”

当函数值判定是 nil 时，程序会 panic。不过，当执行到 *defer* 语句时不会发生，但是在实际调用被 defer 的函数时会发生 panic：

```go
package main

import "fmt"

func main() {
    f := func() {}
    f = nil
    fmt.Println("Before defer statement")
    defer f()
    fmt.Println("After defer statement")
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
Before defer statement
After defer statement
panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xb code=0x1 addr=0x0 pc=0x549b3]

goroutine 1 [running]:
panic(0xda6c0, 0x8201c80e0)
        /usr/local/go/src/runtime/panic.go:481 +0x3e6
main.main()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:11 +0x1d5
```

内部 defer 的函数篡改已命名的返回参数是有可能的。如果没有被命名，那么通过闭包来更改返回变量值不会有任何影响：

```go
import "fmt"

func f() int {
    n := 0
    defer func() {
        n++
    }()
    return n
}

func g() (n int) {
    defer func() {
        n++
    }()
    return n
}

func main() {
    fmt.Printf("f() == %d\n", f())
    fmt.Printf("g() == %d\n", g())
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
f() == 0
g() == 1
```

正如我们将在下面看到的那样， *defer* 语句广泛地应用于处理各种 panic（当然它们也可以应用于各种其他的方面，并不一定是处理各种 error）。

## Panicking

当任意函数 f 发生 panic 时，我们在上面例子中已经看到，在 f 中调用延迟函数的函数将以后进先出的顺序调用。之后将有什么发生呢？之后对于 f 的调用者，这种过程将被重复——它的延迟的函数将被触发。如此反复直到 f 的 goroutine 中的最上面的那个函数。最后，最上面的那个函数的延迟的函数被调用，并且程序终止。就像是一个冒泡直到顶端的调用链：

```go
package main

import "fmt"

func f(ch chan int) {
    defer func() {
        fmt.Println("Deferred by f")
    }()
    g()
    ch <- 0
}

func g() {
    defer func() {
        fmt.Println("Deferred by g")
    }()
    h()
}

func h() {
    defer func() {
        fmt.Println("Deferred by h")
    }()
    panic("boom!")
}

func main() {
    ch := make(chan int)
    go f(ch)
    <-ch
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
Deferred by h
Deferred by g
Deferred by f
panic: boom!

goroutine 17 [running]:
panic(0xb83e0, 0x820220050)
        /usr/local/go/src/runtime/panic.go:481 +0x3e6
main.h()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:24 +0x86
main.g()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:17 +0x35
main.f(0x820224000)
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:9 +0x35
created by main.main
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:29 +0x53
```

值得注意的是，无论 panic 在哪个 Goroutine 开始（主 Goroutine 或者是之后创造的），整个程序都会崩溃。

### 更多关于 Panicking

在延迟的函数内部触发新的 panic 会怎样呢？

```go
import "fmt"

func f(ch chan int) {
    defer func() {
        fmt.Println("Deferred by f")
    }()
    g()
    ch <- 0
}

func g() {
    defer func() {
        fmt.Println("Deferred by g")
    }()
    h()
}

func h() {
    defer func() {
        fmt.Println("Deferred by h")
    }()
    defer func() {
        panic("2nd explosion!")
    }()
    panic("boom!")
}

func main() {
    ch := make(chan int)
    go f(ch)
    <-ch
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
Deferred by h
Deferred by g
Deferred by f
panic: boom!
        panic: 2nd explosion!

goroutine 5 [running]:
panic(0xb8480, 0x8201c82d0)
        /usr/local/go/src/runtime/panic.go:481 +0x3e6
main.h.func2()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:25 +0x65
panic(0xb8480, 0x8201c82c0)
        /usr/local/go/src/runtime/panic.go:443 +0x4e9
main.h()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:27 +0xa3
main.g()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:17 +0x35
main.f(0x820214060)
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:9 +0x35
created by main.main
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:32 +0x53
```

不论如何，事实证明结果会是，在调用直到调用链的顶部的推迟的函数，这个过程将会执行。虽然会有，第二个 panic 这样的新结果，像之前输出的那样，也会显示出来。

```
> go install github.com/mlowicki/lab && ./bin/lab
Deferred by h
Deferred by g
Deferred by f
panic: boom!
        panic: 2nd explosion!

goroutine 5 [running]:
panic(0xb8480, 0x8201c82d0)
        /usr/local/go/src/runtime/panic.go:481 +0x3e6
main.h.func2()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:25 +0x65
panic(0xb8480, 0x8201c82c0)
        /usr/local/go/src/runtime/panic.go:443 +0x4e9
main.h()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:27 +0xa3
main.g()
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:17 +0x35
main.f(0x820214060)
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:9 +0x35
created by main.main
        /Users/mlowicki/projects/golang/spec/src/github.com/mlowicki/lab/lab.go:32 +0x53
```

## Recover

内置的 *recover* 函数可以查看 panic 是否被触发并且阻止 panic 的其他影响。return 语句的返回值要么是传递给 *panic* 的参数（如果正好有 panic ），要么是 *nil* 。在调用 recover 之后，当前 panic 的序列停止，并且这个程序就像，在一个内部延迟的函数调用了 recover 函数的函数被调用时候，就从未有 panic 发生一样：

```go
package main

import "fmt"

func f() {
    fmt.Println("Start f")
    defer func() {
        fmt.Println("Deferred in f")
    }()
    g()
    fmt.Println("End f")
}

func g() {
    fmt.Println("Start g")
    defer func() {
        fmt.Println("Deferred in g")
    }()
    h()
    fmt.Println("End g")
}

func h() {
    fmt.Println("Start h")
    defer func() {
        fmt.Println("1st deferred in h")
    }()
    defer func() {
        fmt.Println("2nd deferred in h")
        if p := recover(); p != nil {
            fmt.Printf("Panic found: %v\n", p)
        }
    }()
    defer func() {
        fmt.Println("3rd deferred in h")
    }()
    panic("boom!")
}

func main() {
    f()
}
```

```
> go install github.com/mlowicki/lab && ./bin/lab
Start f
Start g
Start h
3rd deferred in h
2nd deferred in h
Panic found: boom!
1st deferred in h
End g
Deferred in g
End f
Deferred in f
```

> 允许在推迟的的函数外调用 recover 函数，但总是会返回 nil。
>
> 当没有活跃的 panic 的时候，在推迟的函数中调用 *recover* 的返回值是 *nil*。如果 *panic* 在以 *nil* 为参数的时候被调用，就无法判断 panic 是否正在进行中。

---

如果你喜欢这个帖子并且想获得一些新帖子的最新信息就请关注我吧。让别人也发现这篇文章请点击下面的小心心。

### 参考来源

- [Concurrency Is Not Parallelism](https://vimeo.com/49718712) by Rob Pike
- [Go specification](https://golang.org/ref/spec)

---

via: https://medium.com/golangspec/panicking-like-a-gopher-367a9ce04bb8

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[yixiaoer](https://github.com/yixiaoer)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
