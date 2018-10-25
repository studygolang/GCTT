
https://www.ardanlabs.com/blog/2014/01/concurrency-goroutines-and-gomaxprocs.html

# 并发，协程和最大CPU核数（Concurrency, Goroutines and GOMAXPROCS）

William Kennedy 2014年1月29日

## 介绍

刚刚加入[GO-Minami](http://www.meetup.com/Go-Miami/) 组织的新人经常会说想学习更多有关 go 并发的知识。并发好像在每个语言中都是热门话题，当然我第一次听说 go 语言时也是因为这个点。而 Rob Pike 的一段 [GO Concurrency Patterns](http://www.youtube.com/watch?v=f6kdp27TYZs) 视频才让我真真意识到我需要学习这门语言。

为了了解为什么 go 语言写并发代码更容易更健壮，我们首先需要理解并发程序是什么，和并发程序会导致什么样的结果。在文章中我不就不讨论 CSP (
通信顺序过程)了，这个是 go 语言 channel 实现的基础。这边文章将关注点放在什么是并发编程，goroutines 在其中扮演什么角色、GOMAXPROCS 环境变量和 runtime 函数如何影响文章中写的 go 程序。

## 进程和线程



