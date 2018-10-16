已发布：https://studygolang.com/articles/12135

# Go 中的 5 种高级测试方法

只要你写过 Go 程序，肯定已经知道 Go 语言内置了一个功能完备的测试库。在这篇文章中我们将列出几种能帮助你提高编写测试能力的策略。这些策略是我们在以往的编程经历中总结出的可以很好的节省你的时间和精力的经验。

## 使用测试套件（test suites）

测试套件是本篇文章中最重要的策略。它是一种针对拥有多个实现的通用接口的测试，在下面的例子中，你们将看到我是如何将 `Thinger` 接口的不同实现传递进同一个测试函数，并且让他们测试通过。

```go
type Thinger interface {
    DoThing(input string) (Result, error)
}

// Suite tests all the functionality that Thingers should implement
func Suite(t *testing.T, impl Thinger) {
    res, _ := impl.DoThing("thing")
    if res != expected {
        t.Fail("unexpected result")
    }
}

// TestOne tests the first implementation of Thinger
func TestOne(t *testing.T) {
    one := one.NewOne()
    Suite(t, one)
}

// TestOne tests another implementation of Thinger
func TestTwo(t *testing.T) {
    two := two.NewTwo()
    Suite(t, two)
}
```

幸运的读者可能已经接触过使用该测试策略的代码库。在基于插件开发的系统经常能看到这种测试方式。针对接口的测试可以被用来验证他的所有实现是是否满足接口所需的行为。

测试套件能够让我们在面对一个接口多个实现的时候不用重复的为特定版本书写测试，这会节省我们很多的时间。并且当你切换接口的底层实现代码的时候，不用再写额外的测试，就能保证程序的稳定性。

[这里有一个完整的例子](https://github.com/segmentio/testdemo)。虽然这个例子都是相同的设计。你可以把它想象成一个是远程数据库（Mysql），一个是内存数据库（sqlite）。

另外一个比较棒的例子就是 `golang.org/x/net/nettest` 包。 当我们实现了自定义的 `net.Conn` 接口的时候，可以使用这个包（`golang.org/x/net/nettest`）来直接验证我们的实现是否满足接口的要求，而不用自己重新设计测试代码。

## 避免接口污染（interface pollution）

我们不能撇开接口谈 Go 的测试。

接口在测试的上下文中十分重要，因为对于测试来讲，接口是一种十分有力的工具，所以正确的使用他变得尤为重要。一个包经常会导出接口给用户，用户可以使用包中预定义的接口实现，也可以自己为该接口定义实现。

> The bigger the interface, the weaker the abstraction.
>
> Rob Pike, Go Proverbs

在导出一个接口的时候我们需要很谨慎的考虑是否应该导出它。开发者为了能让用户可以自定义接口的实现往往选择导出这个接口。然而我们并不需要这样，你只需要在你的结构体中实现了这个接口的行为，就可以在需要该接口的地方使用这个结构体。这样，你的代码包和用户的代码包就不会有强制的依赖关系。一个很好的例子就是 [errors package](https://godoc.org/github.com/pkg/errors) ( `error` 接口没有被导出，但是我们可以在任何包中都可以定义自己的实现)。

如果你不想导出一个接口，那么可以使用 [internal/package subtree](https://golang.org/doc/go1.4#internalpackages) 来保证接口只有在包内可见。这样我们就不用担心用户会依赖这个接口，也可以在新的需求出现的时候，十分灵活的修改这个接口。我们经常在使用一个外部依赖的时候创建接口，并使用依赖注入的方式把这个外部依赖作为这个接口的实现，这样我们就可以排除外部依赖的因素而只测试自己的代码。这让用户只需要封装代码库中自己使用的那一小部分。

更多详情，[https://rakyll.org/interface-pollution/](https://rakyll.org/interface-pollution/)

## 不要导出并发原语

Go 提供了非常易于使用的并发原语，这也导致了它被过度的使用。我们主要担心的是 `channel` 和 `sync` package 。有的时候我们会导出一个 `channel` 给用户使用。另外一个常见的错误就是在使用 `sync.Mutex` 作为结构体字段的时候没有把它设置成私有。这并不总是很糟糕，不过在写测试的时候却需要考虑的更加全面。

当我们导出 `channel` 的时候我们就为这个包的用户带来了测试上的一些不必要的麻烦。你每导出一个 `channel` 就是在提高用户在测试时候的难度。为了写出正确的测试，用户必须考虑这些：

* 什么时候数据发送完成。
* 在接受数据的时候是否会发生错误。
* 如果需要在包中清理使用过的channel的时候该怎么做。
* 如何将 API 封装成一个接口，使我们不用直接去调用它。

请看下面这个在队列中读取数据的例子。这个库导出了一个 `channel` 用来让用户读取他的数据。

```go
type Reader struct {...}
func (r *Reader) ReadChan() <-chan Msg {...}
```

现在有一个使用你的库的用户想写一个测试程序。

```go
func TestConsumer(t testing.T) {
    cons := &Consumer{
        r: libqueue.NewReader(),
    }
    for msg := range cons.r.ReadChan() {
        // Test thing.
    }
}
```

用户可能会认为使用依赖注入是一种好的方式。并且用下面的方式实现了自己的队列：

```go
func TestConsumer(t testing.T, q queueIface) {
    cons := &Consumer{
        r: q,
    }
    for msg := range cons.r.ReadChan() {
        // Test thing.
    }
}
```

这时有一个潜在的问题。

```go
func TestConsumer(t testing.T, q queueIface) {
    cons := &Consumer{
        r: q,
    }
    for {
        select {
        case msg := <-cons.r.ReadChan():
            // Test thing.
        case err := <-cons.r.ErrChan():
            // What caused this again?
        }
    }
}
```

现在我们不知道如何来向这个 `channel` 中插入数据，来模拟使用时这个代码库的真实运行情况，如果这个库提供了一个同步的API接口，那么我们可以并发的调用它，这样测试就会非常的简单。

```go
func TestConsumer(t testing.T, q queueIface) {
    cons := &Consumer{
        r: q,
    }
    msg, err := cons.r.ReadMsg()
    // handle err, test thing
}
```

当你有疑问的时候，一定要记住在用户的包中使用 goroutine 是很简单的事情，但是你的包一旦导出了就很难被移除。所以一定别忘了在 package 的文档中注明这个包是不是多 goroutine 并发安全的。

有时，我们不可避免的需要导出一个 `channel`。为了减少这样带来的问题，你可以通过导出只读的 `channel(<-chan)` 或者只写的 `channel(chan<-)` 来替代直接导出一个 `channel`。

## 使用 `net/http/httptest`

`httptest` 包可以让你在不用绑定端口或启动一个服务器的情况下测试你的  `http.Handle` 函数。这样可以加速测试的速度，并且以一种很小的成本让这些测试并行的运行。

下面是一个用 2 种方法实现同一个测试用例的例子。它们并不相似但是可以在测试时节省你很多的代码和系统资源。

```go
func TestServe(t *testing.T) {
    // The method to use if you want to practice typing
    s := &http.Server{
        Handler: http.HandlerFunc(ServeHTTP),
    }
    // Pick port automatically for parallel tests and to avoid conflicts
    l, err := net.Listen("tcp", ":0")
    if err != nil {
        t.Fatal(err)
    }
    defer l.Close()
    go s.Serve(l)

    res, err := http.Get("http://" + l.Addr().String() + "/?sloths=arecool")
    if err != nil {
        log.Fatal(err)
    }
    greeting, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(greeting))
}

func TestServeMemory(t *testing.T) {
    // Less verbose and more flexible way
    req := httptest.NewRequest("GET", "http://example.com/?sloths=arecool", nil)
    w := httptest.NewRecorder()

    ServeHTTP(w, req)
    greeting, err := ioutil.ReadAll(w.Body)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(greeting))
}
```

可能这种方式给你带来的最大好处就是能让你单独测试你想测试的函数。没有路由、中间件等等其他的影响因素。

想要了解更多，请看 [post by Mark Berger](http://markjberger.com/testing-web-apps-in-golang/)。

##  使用单独的 `_test` 测试包

大多数情况下，测试都是写在和 package 相同的包中并以 `pkg_test.go` 命名。而一个单独的测试包就是将测试代码和正式代码分割在不同的包中。一般单独的测试包都以包名+`_test` 命名（例如:`foo` 包的测试包为 `foo_test`）。在测试包中你可以把需要测试的包和其他测试依赖的包一起导入进去。这种方式能让测试更加的灵活。当遇到包中循环引用的情况，我们推荐这种变通的方式。他能防止你对代码中易变部分进行测试。并且能让开发者站在包的使用者的角度上来使用自己开发的包。如果你开发的包很难被使用，那么他也肯定很难被测试。

这种测试方法通过限制易变的私有变量来避免容易发生改变的测试。如果你的代码不能通过这种测试，那么在使用的过程中肯定也会有问题。

这种测试方法也有助于避免循环的引用。大多数包都会依赖你在测试中所要用到的包，所以很容易发生循环依赖的情况。而这种单独的测试包在原包，和被依赖包的层次之外，就不会出现循环依赖的问题。一个例子就是 `net/url` 包中实现了一个URL的解析器，这个解析器被 `net/http` 包所使用。但是当对 `net/url` 包进行测试的时候，就需要导入 `net/http` 包，因此 `net/url_test` 包产生了。

现在当你使用一个单独的测试包的时候，包中的一些结构体或者函数由于包的可见性的原因在单独的测试包中不能被访问到。大部分人在基于时间的测试的时候都会遇到这种问题。针对这种问题，我们可以在包中 `xx_test.go` 文件中将他们导出，这样我们就可以正常的使用了。

## 记住这些事情

上面的这些方法并不是银弹，但是在实际使用的过程中需要我们仔细的分析问题，并找到最适合的解决方案。

想要了解更多的测试方法？

你可以看看这些文章：

* [Writing Table Driven Tests in Go](https://dave.cheney.net/2013/06/09/writing-table-driven-tests-in-go) by Dave Cheney
* [The Go Programming Language chapter on Testing](http://www.gopl.io/)

或者这些视频：

* [Hashimoto’s Advanced Testing With Go talk from Gophercon 2017](https://www.youtube.com/watch?v=yszygk1cpEc)
* [Andrew Gerrand's Testing Techniques talk from 2014](https://talks.golang.org/2014/testing.slide#1)

---

via: https://segment.com/blog/5-advanced-testing-techniques-in-go/

作者：[Alan Braithwaite](https://segment.com/blog/authors/alan-braithwaite/)
译者：[saberuster](https://github.com/saberuster)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
