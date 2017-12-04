# Go 语言如何去解决 Web 开发人员面临的众多问题？


坦白的说,我的团队非常厌恶我对 GO 语言传道的方式,每当我们团队的代码库出现问题时,他们希望我用一种更委婉的方式提出.


![](https://ewanvalentine.io/content/images/2016/01/Screen-Shot-2016-01-29-at-11-57-56.png) 

我学会的第一种编程语言是 PHP，这是个优秀的语言，我可以用它很快地构建 Web 应用程序，这些应用程序也能够达到预期的效果。但是我注意到，为了使其可用，我会花费大量的时间来关注缓存。

我也发现自己依靠很多第三方库来做一些更复杂的任务，比如队列，Web Sockets 等等。我发现自己使用了 Pusher，RabbitMQ，Beanstalkd 等等。

这让人感觉有点不好。在使用 Ruby，Node 和 Python 的时候，会出现类似的问题。在并发性， WebSockets 和性能方面，这些语言会让人感觉到它们是不完整。

我需要完全依赖框架和大量文档，“语法糖”，DSL，坦率地说，它们经常会带来很多非常占用空间的东西。

我开始把目光转向 GO 语言。 

首先 Go 是一种的静态类型语言，我一直都喜欢这种方式。 所以我学的非常快。Go 是一种偏底层的语言，你会遇到指针和内存引用等问题。

我之前曾经涉足 C 语言，而 Go 感觉和 C 很像，但是 Go 提供的标准库非常强大且易于使用，所以我对 Go 语法的精炼扼要感到震惊。

在深入研究之后，我决定研究Go是如何解决PHP编写Web应用/ API等出现的一些问题。

如何去解决 Web Sockets？Go 有几个很出色的库文件。 下面是一个 Gin 框架的使用 Gorilla websockets 库文件的例子...

```golang
package main

import (  
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "net/http"
)

func main() {  
    r := gin.Default()
    r.LoadHTMLFiles("index.html")

    r.GET("/", func(c *gin.Context) {
        c.HTML(200, "index.html", nil)
    })

    r.GET("/ws", func(c *gin.Context) {
        wshandler(c.Writer, c.Request)
    })

    r.Run("localhost:2020")
}

var wsupgrader = websocket.Upgrader{  
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func wshandler(w http.ResponseWriter, r *http.Request) {  
    conn, err := wsupgrader.Upgrade(w, r, nil)

    if err != nil {
        fmt.Println("Failed to upgrade ws: %+v", err)
        return
    }

    for {
        t, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }
        conn.WriteMessage(t, msg)
    }
}
```

### 并发

在 PHP 中，我不得不要么使用一些黑客手法运行线程，比如使用 `shell_exec()` 将一个任务委托给一个新的线程，或者使用一个单独的服务，比如 Beanstalkd 或者 RabbitMQ。

不过，你只要在需要并发功能的函数前加 `go` 的关键字就可实现。 例如...

```golang
go func(test string) {  
    doSomething(test)
}("This is a test")
```

或者你可以使用 channel...

```golang
package main

import "fmt"

func main() {  
    ch := make(chan int, 2)
    ch <- 1
    ch <- 2
    fmt.Println(<-ch)
    fmt.Println(<-ch)
}
```

我将之前一个上传图片到 s3 的耗时任务放到 goroutine 中去实现接近即时的上传效果，没有第三方服务，完全本地。 对于大多数开发人员来说不那么令人印象深刻，但是对于 PHP 背景的开发人员来说，我对 Go 的易用性和性能提升感到震惊。

### 测试

单元测试在 PHP 或 Javascript 中可能会有点痛苦。 有无数不同的测试框架，但没有一个能够像 go built 命令去如此简单自然的进行测试。

main.go 

```golang
package main

import "fmt"

func sup(name string) string {  
    return "Sup, " + name
}

func main() {  
    fmt.Println(sup("Ewan"))
}
```

现在我们的测试, main_test.go

```golang
package main

import "testing"

func TestSup(t *testing.T) {  
    expected := "Sup, Ewan"
    outcome := sup("Ewan")
    if outcome != expected {
        t.Fatalf("Expected %s, got %s", expected, outcome)
    }
}
```

我现在把所有的通过 `go test` 进行测试，结果是... 

![](https://ewanvalentine.io/content/images/2016/02/Screen-Shot-2016-02-23-at-21-57-33.png) 

是不是非常的简单？

### 运行速度

在用 PHP 写 RESTful API 时，我有非常多的 Symfony2 和 Laravel 等框架的使用经验。

没有预先着重考虑几个级别的缓存; 如在内存缓存，操作缓存，全页缓存等。代码响应时间就会是一个蜗牛的步伐。Ruby 如此的声名狼藉就是由于缓慢。

由于 Go 的静态类型，编译性质以及对并发的原生支持。 Go 运行起来特别快。

看看[框架基准测试](https://www.techempower.com/benchmarks/) ，实践是最好的证明。 Go 最受欢迎的框架是 Gin 和 Revel，它们在大多数测试中的排名要高于 PHP 或者 Ruby。

### DevOps

关于 Go 我还注意到一些，这让我非常震惊，不需要部署成千上万的文件，或者配置 Web 服务器或者 php-fpm 等。甚至不需要在你的服务器上安装 Go。

一旦 Go 应用程序被编译（ `go build` ），你只剩下一个小小的二进制文件。你只要去运行一个单独的文件。

Go 还有一个非常稳固的内置 Http 服务器... 

```golang
package main

import (  
    "io"
    "net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {  
    io.WriteString(w, "Hello world!")
}

func main() {  
    http.HandleFunc("/", hello)
    http.ListenAndServe(":8000", nil)
}
```

### 语法

Go 的语法不像 Ruby 那样漂亮，或者像 JavaScript 一样简单。 但是它很简洁，它让人觉得便底层，但 Go 让人感觉强大，表现力强。 我们都在看传统的 PHP 代码库，并感到身体不适。 相比之下，Go 非常容易去阅读。

Go 另一个令人难以置信的好处是，在你编写 Go 代码的时候，已经有了一个很好的“最佳实践”。

当然，PHP 有 PSR 标准等，但是它们相当新，开发人员采用它们的速度很慢。 而 Go 的语言设计者从一开始就是明确的。
 
[格式化工具](https://blog.golang.org/go-fmt-your-code) 也内置在语言生态系统中，可以使用 'go fmt' 来运行。

所以我的一点陋见就是为什么我完全沉迷于 Go，现在我不能再回到 PHP 了。



via: https://ewanvalentine.io/why-go-solves-so-many-problems-for-web-developers/

作者：[Ewan Valentine](https://ewanvalentine.io/author/ewan/)
译者：[Dingo1991](https://github.com/Dingo1991)
校对：[rxcai](https://github.com/rxcai)


本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
