首发于：https://studygolang.com/articles/13612

# 为什么 Go 语言允许百万级别的 goroutines，而 Java 只允许数千级别的 threads?

很多有过 JVM 相关语言工作经验的程序员或许都遇到过如下问题：

```
[error] (run-main-0) java.lang.OutOfMemoryError: unable to create native thread:
[error] java.lang.OutOfMemoryError: unable to create native thread:
[error] 	at java.base/java.lang.Thread.start0(Native Method)
[error] 	at java.base/java.lang.Thread.start(Thread.java:813)
...
[error] 	at java.base/java.lang.Thread.run(Thread.java:844)
```

额，超出 thread 限制导致内存溢出。在作者的笔记本的 linux 上运行，这种情况一般发生在创建了 11500 个左右的 thread 时候。

但如果你用 Go 语言来做类似的尝试，每创建一个 Goroutine ，并让它永久的 Sleep ，你会得到一个完全不同的结果。在作者的笔记本上，在作者等待的不耐烦之前，GO语言创建了大约7千万个 Goroutine 。为什么我们可以创建的 Goroutines 比 thread 多这么多呢？回答这个问题需要回到操作系统层面来进行一次愉快的探索。这不仅仅是一个学术问题---在现实世界中它也揭示了如何进行软件设计。事实上，作者碰到过很多次软件出现 JVM 的 Thread 达到上限的情况，要么是因为垃圾代码导致 Thread 泄露，要么就是因为一些开发工程师压根不知道 JVM 有 Thread 限制这回事。

## **那么到底什么是 Thread ？**

“Thread" 本身其实可以代表很多不同的意义。在这篇文章中，作者把它描述为一种逻辑上的 Thread。Thread 由如下内容组成：一系列按照线性顺序可以执行的指令（operations)；和一个逻辑上可以执行的路径。CPUs 中的每一个 Core 在同一时刻只能真正并发执行一个 logic thread<sup>[1]</sup>。这就产生了一个结论：如果你的 threads 个数大于 CPU 的 Core 个数的话，有一部分的 Threads 就必须要暂停来让其他 Threads 工作，直到这些 Threads 到达一定的时机时才会被恢复继续执行。而暂停和恢复一个线程，至少需要记录两件事情:

1. 当前执行的指令位置。亦称为：说当前线程被暂停时，线程正在执行的代码行;
2. 还需要一个栈空间。 亦可认为：这个栈空间保存了当前线程的状态。一个栈包含了 local 变量也就是一些指针指向堆内存的变量（这个是对于 Java 来说的，对于 C/C++ 可以存储非指针）。一个进程里面所有的 threads 是共享一个堆内存的<sup>[2]</sup>。

有了上面两样东西后，cpu 在调度 thread 的时候，就有了足够的信息，可以暂停一个 thread，调度其他 thread 运行，然后再将暂停的 thread 恢复，从而继续执行。这些操作对于 thread 来说通常是完全透明的。从 thread 的角度来看，它一直都在连续的运行着。thread 被取消调度这样的行为可以被观察的唯一办法就是测量后续操作的时间<sup>[3]</sup>。

让我们回到最初的问题，为什么我们可以创建那么多的 Goroutinues 呢？

## **JVM使用的是操作系统的Thread**

尽管规范没有要求所有现代的通用 JVM，在我所知道的范围内，当前市面上所有的现代通用目的的 JVM 中的 thread 都是被设计成为了操作系统的thread。下面，我将使用“用户空间 threads" 的概念来指代被语言来调度而不是被操作系统内核调度的 threads。操作系统级别实现的 threads 主要有如下两点限制：首先限制了 threads 的总数量，其次对于语言层面的 thread 和操作系统层面的 thread 进行 1:1 映射的场景，没有支持海量并发的解决方案。

### **JVM 中固定的栈大小**

**使用操作系统层面的 thread，每一个 thread 都需要耗费静态的大量的内存**

第二个使用操作系统层面的 thread 所带来的问题是，每一个 thread 都需要一个固定的栈内存。虽然这个内存大小是可以配置的，但在 64 位的 JVM 环境中，一个 thread 默认使用1MB的栈内存。虽然你可以将默认的栈内存大小改小一点，但是您会权衡内存使用情况, 从而增加堆栈溢出的风险。在你的代码中递归次数越大，越有可能触发栈溢出。如果使用1MB的栈默认值，那么创建1000个 threads ，将使用 1GB 的 RAM ，虽然 RAM 现在很便宜，但是如果要创建一亿个 threads ，就需要T级别的内存。

### **Go 语言的处理办法：动态大小的栈**

Go 语言为了避免是使用过大的栈内存（大部分都是未使用的）导致内存溢出，使用了一个非常聪明的技巧：Go 的栈大小是动态的，随着存储的数据大小增长和收缩。这不是一件简单微小的事情，这个特性经过了好几个版本的迭代开发<sup>[4]</sup>。很多其他人的关于 Go 语言的文章中都已经做了详细的说明，本文不打算在这里讨论内部的细节。结果就是新建的一个 Goroutine 实际只占用 4KB 的栈空间。一个栈只占用 4KB，1GB 的内存可以创建 250 万个 Goroutine，相对于 Java 一个栈占用 1MB 的内存，这的确是一个很大的提高。

### 在 JVM 中上下文的切换是很慢的

**使用操作系统的 threads 的最大能力一般在万级别，主要消耗是在上下文切换的延迟。**

因为 JVM 是使用操作系统的 threads ，也就是说是由操作系统内核进行 threads 的调度。操作系统本身有一个所有正在运行的进程和线程的列表，同时操作系统给它们中的每一个都分配一个“公平”的使用 CPU 的时间片<sup>[5]</sup>。当内核从一个 thread 切换到另外一个时候，它其实有很多事情需要去做。新线程或进程的运行必须以世界的视角开始，它可以抽象出其他线程在同一 CPU 上运行的事实。本文不想在这里多说，但是如果你感兴趣的话，可以参考[这里](https://en.wikipedia.org/wiki/Context_switch)。（t 问题的关键点是上下文的切换大概需要消耗 1-100µ 秒。这个看上去好像不是很耗时，但是在现实中每次平均切换需要消耗10µ秒，如果想让在一秒钟内，所有的 threads 都能被调用到，那么 threads 在一个 core 上最多只能有 10 万个 threads，而事实上这些 threads 自身已经没有任何时间去做自己的有意义的工作了。

**Go 语言完全不同的处理：运行多个 Goroutines 在一个 OS thread 上**

Golang 语言本身有自己的调度策略，允许多个 Goroutines 运行在一个同样的 OS thread 上。既然 Golang 能像内核一样运行代码的上下文切换，这样它就能省下大量的时间来避免从用户态切换到 ring-0 的内核态再切换回来的过程。但是这只是表面上能看到的，事实上为 Go 语言支持 100 万的 goroutines，Go 语言其实还做了更多更复杂的事情。

即使 JVM 把 threads 带到了用户空间，它依然无法支持百万级别的 threads ，想象下在你的新的系统中，在 thread 间进行切换只需要耗费100 纳秒，即使只做上下文切换，有也只能使 100 万个 threads 每秒钟做 10 次上下文的切换，更重要的是，你必须要让你的 CPU 满负荷的做这样的事情。支持真正的高并发需要另外一种优化思路：当你知道这个线程能做有用的工作的时候，才去调度这个线程！如果你正在运行多线程，其实无论何时，只有少部分的线程在做有用的工作。Go 语言引入了 channel 的机制来协助这种调度机制。如果一个 goroutine 正在一个空的 channel 上等待，那么调度器就能看到这些，并不再运行这个 goroutine 。同时 Go 语言更进了一步。它把很多个大部分时间空闲的 goroutines 合并到了一个自己的操作系统线程上。这样可以通过一个线程来调度活动的 Goroutine（这个数量小得多），而是数百万大部分状态处于睡眠的 goroutines 被分离出来。这种机制也有助于降低延迟。

除非 Java 增加一些语言特性来支持调度可见的功能，否则支持智能调度是不可能实现的。但是你可以自己在“用户态”构建一个运行时的调度器，来调度何时线程可以工作。其实这就是构成Akka这种数百万 actors<sup>[6]</sup> 并发框架的基础概念。

## **结语思考**

未来，会有越来越多的从操作系统层面的 thread 模型向轻量级的用户空间级别的 threads 模型迁移发生<sup>[7]</sup>。从使用角度看，使用高级的并发特性是必须的，也是唯一的需求。这种需求其实并没有增加过多的的复杂度。如果 Go 语言改用操作系统级别的 threads 来替代目前现有的调度和栈空间自增长的机制，其实也就是在 runtime 的代码包中减少数千行的代码。但对于大多数的用户案例上考虑，这是一个更好的的模式。复杂度被语言库的作者做了很好的抽象，这样软件工程师就可以写出高并发的程序了。

---

1. 超线程技术（Hyperthreading）可以成倍的高效地使用cpu的核。指令流水线（Instruction pipelineing）也可以增加CPU的并行执行的能力，然而，到目前为止，它是 O(numCores)。
2. 这个观点在某些特殊的场景下是不成立的，如果有这种场景，麻烦告知作者。
3. 这其实是一种攻击媒介。Javascript 可以检测由键盘中断引起的时间微小差异。这可以被恶意网站用来侦听，而不是你的键盘中断，而是用于他们的时间。[https://mlq.me/download/keystroke_js.pdf](https://mlq.me/download/keystroke_js.pdf) [
4. Go语言起初使用的是“分段栈模型“，即栈空间是被分割到内存中不同的区域（译者注：在其他语言中栈空间一般是连续的），同时使用一些非常聪明的 bookkeeping 机制进行栈追踪。后来的版本实现为了提升性能，在一些特殊的场景下， 使用连续的栈来替代“分割栈模型”。就像调整hash表一样，分配一个新的大的栈空间，并通过一些复杂的指针操作，把所有内容都复制到新的更大的栈空间去。
5. 线程可以通过调用 nice（请参阅 man nice）来标记它们的优先级，以获取更多信息来控制它们被安排调度。
6. 为了能实现大规模的高并发，Actor 和 Goroutines for Scala/Java 的用户相同。就和 Goroutines 一样，actors 的调度程序可以查看哪些 actors 在他们的邮箱中有消息，并且只运行准备好做有用工作的 actors。实际上你可以有更多的 actors，而不是你可以拥有的例程，因为 actors 不需要堆栈。然而，这意味着如果一个 actor 没有快速处理消息，调度器将被阻塞（因为 Actor 不具有它自己的堆栈，所以它不能在消息中间暂停）。阻塞的调度器意味着没有消息处理，事情会迅速停止。这是一种折衷的处理方案。
7. 在 Apache 的 web 服务器上，每处理一个请求就需要一个 OS 级别的 Thread，所以一个 Apache 的 web 服务器的并发连接性能只有数千级别。Nginx 选择了另一种模型，即使用一个操作系统级别的 Thread 来处理成百甚至上千个并发连接，允许了更好程度的并发。Erlang 也使用了类似的模型，允许数百万个 actors 同时运行。Gevent 将 Python 的 greenlet（用户空间线程）带入 Python，从而实现比其他方式支持的更高程度的并发性（Python 线程是 OS 线程）。

---

via: https://rcoh.me/posts/why-you-can-have-a-million-go-routines-but-only-1000-java-threads/

作者：[Russell Cohen](https://rcoh.me/)
译者：[skyismine2010](https://github.com/skyismine2010)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
