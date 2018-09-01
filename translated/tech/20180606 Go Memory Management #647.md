# Go Memory Management
# Go 语言的内存管理

This is a blog post version of a talk I gave at Vilnius Go Meetup. If you’re ever in Vilnius and enjoy Go come join us and consider speaking 🙂
这篇博客是我在维尔纽斯的 [Go Meetup](https://www.meetup.com/Vilnius-Golang/events/249897910/) 演讲的总结。如果你在维尔纽斯并且喜欢 Go 语言，欢迎加入我们并考虑作演讲

So, in this post we will explore Go memory management. Let’s begin with a following little program:
在这篇博文中我们将要探索 Go 语言的内存管理，首先让我们来思考以下的这个小程序：

```go
func main() {
    http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
    })

    http.ListenAndServe(":8080", nil)
}
```

Let’s compile and run it:
编译并且运行：

```bash
go build main.go
./main
```

Now let’s find the running process via ps:
接着我们通过 `ps` 命令观察这个正在运行的程序：

```bash
ps -u --pid 16609
USER PID %CPU %MEM VSZ RSS TTY STAT START TIME COMMAND
povilasv 16609 0.0 0.0 388496 5236 pts/9 Sl+ 17:21 0:00 ./main
```

We can see that this program consumes 379.39 MiB of virtual memory and resident size is 5.11 mb. Wait what? Why ~380 MiB?
我们发现，这个程序居然耗掉了 379.39M 虚拟内存，实际使用内存为 5.11M。这有点儿夸张吧，为什么会用掉 380M 虚拟内存？

一点小提示:

Virtual Memory Size(VSZ) is all memory that the process can access, including memory that is swapped out, memory that is allocated, but not used, and memory that is from shared libraries. (Edited, good explanation in stackoverflow.)
虚拟内存大小(VSZ)是进程可以访问的所有内存，包括换出的内存、分配但未使用的内存和共享库中的内存。(stackoverflow 上有很好的解释。)

Resident Set Size(RSS) is number of memory pages the process has in real memory multiplied by pagesize. This excludes swapped out memory pages.
驻留集大小（RSS）是进程在实际内存中的内存页数乘以内存页大小，这里不包括换出的内存页（译者注：包含共享库占用的内存）。

Before deep diving into this problem, let’s go thru some basics of computer architecture and memory management in computers.
在深入研究这个问题之前，让我们先介绍一些计算机架构和内存管理的基础知识。

## Memory Basics
## 内存的基本知识

[维基百科](https://en.wikipedia.org/wiki/Random-access_memory)对 RAM 的定义如下：

>Random-access memory (RAM /ræm/) is a form of computer data storage that stores data and machine code currently being used.
A random-access memory device allows data items to be read or written in almost the same amount of time irrespective of the physical location of data inside the memory.