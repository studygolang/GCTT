## https://www.ardanlabs.com/blog/2019/04/concurrency-trap-2-incomplete-work.html

# 并发陷阱 2:未完成的工作

Jacob Walker 2019年4月18日

## 介绍
In my first post on Goroutine Leaks, I mentioned that concurrency is a useful tool but it comes with certain traps that don’t exist in synchronous programs. To continue with this theme, I will introduce a new trap called incomplete work. Incomplete work occurs when a program terminates before outstanding Goroutines (non-main goroutines) complete. Depending on the nature of the Goroutine that is being terminated forcefully, this may be a serious problem.
在我的第一篇文章 Goroutine 泄露中，我提到并发编程是一个很有用的工具，但是使用它也会带来某些非并发编程中不存在的陷阱。为了继续这个主题，我将介绍一个新的陷阱，这个陷阱叫做未完成的工作。当进程在非主协程的协程结束前终止时这种陷阱就会发生。根据Gorotine的本性，强制关闭它将造成一个严重的问题。

Incomplete Work
To see a simple example of incomplete work, examine this program.

## 未完成的工作

为了看到一个简单的未完成任务陷阱的例子，请检查这个程序

Listing 1
https://play.golang.org/p/VORJoAD2oAh

**例1**

https://play.golang.org/p/VORJoAD2oAh

```
5 func main() {
6     fmt.Println("Hello")
7     go fmt.Println("Goodbye")
8 }
```

The program in Listing 1 prints "Hello" on line 6 and then on line 7, the program calls fmt.Println again but does so within the scope of a different Goroutine. Immediately after scheduling this new Goroutine, the program reaches the end of the main function and terminates. If you run this program you won’t see the “Goodbye” message because in the Go specification there is a rule:

在例一的程序中，第6行打印了"Hello",随后在第 7 行，这个程序再次调用了 `fmt.Println` ，但是这次是在一个不同的 Groutine 中。









---

via: https://www.ardanlabs.com/blog/2019/04/concurrency-trap-2-incomplete-work.html

作者：[Jacob Walker](https://github.com/jcbwlkr)
译者：[xmge](https://github.com/xmge)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
