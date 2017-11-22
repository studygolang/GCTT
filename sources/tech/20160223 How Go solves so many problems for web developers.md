【翻译中】 by Dingo1991
# How Go solves so many problems for web developers

原文链接：https://ewanvalentine.io/why-go-solves-so-many-problems-for-web-developers/

My team at work are, quite frankly very tired of my Go evangelist ways. Every time someone mentions some problem with one of our codebases, they can usually count on me to chime in with a less than subtle hint.My first programming language was PHP, which was great, I could build web apps pretty quickly and they'd do the trick. But I noticed I'd spend a lot of time obsessing over caching just to make it usable. I also found myself relying on a lot of third party vendors in order to do some more complex tasks, such as queuing, web sockets etc. I found myself using Pusher, RabbitMQ, Beanstalkd etc, etc. This felt kind of wrong. But it seemed a similar story for Ruby, Node and Python. They started to feel incomplete when it came to concurrency, websockets, and performance.I was also completely dependent on frameworks and the mountains of documentation, 'syntactic sugar', DSL and quite frankly, bloatware they often came with. Then I started looking more into Golang. First of all Go's statically typed, which I always liked the idea of. So I got my head around that pretty quickly. Go's pretty low level, you find yourself worrying about pointers and memory references etc. I'd previously dabbled in C, and Go feels a lot like C, but the standard libraries which come shipped with Go were so powerful and easy to use, that I was stunned at how to-the-point Go's syntax is. As I delved a little deeper, I decided to look at how some of the problems I faced with PHP were solved when writing web apps/API's etc in Go. What about web sockets? Well, Go has several fantastic libraries. Here's an example using Gorilla websockets with Gin framework... package main

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
ConcurrencyIn PHP I'd have had to use either some thread execution hack, delegating a task to a new thread using shell_exec() or by using a separate service such as Beanstalkd or RabbitMQ. Go however, you simply prepend a function you want to be concurrent with the go keyword. For example... go func(test string) {  
    doSomething(test)
}("This is a test")
Or you can use channels... package main

import "fmt"

func main() {  
    ch := make(chan int, 2)
    ch <- 1
    ch <- 2
    fmt.Println(<-ch)
    fmt.Println(<-ch)
}
I made a previously slow image upload near instant by passing off the upload to S3 job to a go routine, no third party services, completely native. To most developers that might not sound that impressive, but coming from a PHP background, I was stunned at the ease of use and the performance gains. TestingUnit testing can be a bit of a ball-ache in PHP or Javascript. There's countless different testing frameworks, but none of them feel as simple, and as natural as Go's built in testing. main.go package main

import "fmt"

func sup(name string) string {  
    return "Sup, " + name
}

func main() {  
    fmt.Println(sup("Ewan"))
}
Now our tests in, main_test.go package main

import "testing"

func TestSup(t *testing.T) {  
    expected := "Sup, Ewan"
    outcome := sup("Ewan")
    if outcome != expected {
        t.Fatalf("Expected %s, got %s", expected, outcome)
    }
}
All I do now is run go test and I get... How bloody easy is that ffs? SpeedI've got a fair amount of experience writing RESTful API's in PHP using frameworks such as Symfony2 and Laravel. Without significant forethought to several levels of caching; such as in memory caching, op caching, full page caching etc. Response times can be a snail pace. Ruby is especially notorious for being sluggish. Because of Go's statically typed, compiled nature, as well as its native support for concurrency. Go's blisteringly fast. Looking at the framework benchmarks, the proof is in the pudding. Go and it's most popular frameworks, such as Gin and Revel, rank far higher in most tests than its PHP or Ruby counterparts. DevopsSoemthing else I noticed about Go which really struck me, not having to deploy thousands of files, or configure web servers, or php-fpm etc. You don't even need to install Go on your serve. Once a Go application is compiled (go build), you're left with a tiny binary. A single file you can just run. Go also has a very solid http server built in... package main

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
 SyntaxGo's syntax isn't beautiful like Ruby or as simple as Javascript's. But it's succinct, it feels low level, but at the same time it feels powerful, and expressive. We've all looked at legacy PHP codebases, and felt physically ill. In comparison Go is so easy to follow. Another fantastic benefit of Go, is that there's already a well established 'best practices' when it comes to how you write Go. Sure, PHP has PSR standards etc, but they're fairly new and developers have been slow to adopt them. Whereas Go's language designers have been explicit from the outset. Formatting tools are also built into the languages ecosystem, and can be ran using go fmt.So that's my two cents on why I'm totally obsessed with Go, and can now never go back to PHP now I've 'seen the light'.


----------------

via: https://ewanvalentine.io/why-go-solves-so-many-problems-for-web-developers/

作者：[Ewan Valentine](https://ewanvalentine.io/author/ewan/)
译者：[译者ID](https://github.com/译者ID)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推
