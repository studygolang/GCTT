首发于：https://studygolang.com/articles/28438

# Go 运行时调度器处理系统调用的巧妙方式

[goroutine](https://tour.golang.org/concurrency/1) 是 Go 的一个标志性特点，是被 Go 运行时所管理的轻量线程。Go 运行时使用[一个 M:N 工作窃取调度器](https://rakyll.org/scheduler/)实现 goroutine，将 Goroutine 复用在操作系统线程上。调度器有着特殊的术语用来描述三个重要的实体；G 是 goroutine，M 是 OS 线程（一个“机器 machine”），P 是“处理器（processor）”，它的核心是有限的资源，而 M 需要这些资源来运行 Go 代码。限制 P 的供应是 Go 用来限制一次执行多少操作以避免整个系统超载的手段。通常来说，每个 OS 所报告的实际的 CPU 有一个对应的 P （P 的数量是 [GOMAXPROCS](https://golang.org/pkg/runtime/)）。

当 Goroutine 执行 网络 IO 或者任何觉得可以异步完成的系统调用操作时，Go 有一个完整的运行时子系统，[netpoller](https://morsmachine.dk/netpoller)，（使用类似 [epoll](https://medium.com/@copyconstruct/the-method-to-epolls-madness-d9d2d6378642) 的系统调用机制）将看起来像多个单独的同步操作转换为一个单独的等待。goroutine 并没有真正进行阻塞的系统调用，而是像等待一个 channel 就绪那样进入休眠状态等待其网络套接字。如果很难有效地实现，概念上讲这些都是直白的。

无论如何，网络 IO 以及类似的东西远不是 Go 程序可以处理的唯一的系统调用，因此 Go 也必须处理阻塞的系统调用。对 Goroutine 的 M 来说，处理阻塞的系统调用的直接方式是在系统调用前释放 P ，并且在系统调用恢复后尝试重新获取 P 。如果那时候没有空闲的 P ，goroutine 会随着其他等待运行的任务被停放在调度器中。

虽然理论上所有的系统调用都是阻塞的，在实践中不是所有的调用都会阻塞。例如，在现代系统中，获取当前时间的“系统调用”可能甚至没有进入内核（见 Linux 的 [vdso(7)](http://man7.org/linux/man-pages/man7/vdso.7.html)）。让 Goroutine 完成释放他们当前的 P 的全部工作再为了这些系统调用重新获取一个 P 有两个问题：首先，所有涉及到的数据结构的锁定（和释放）有着很大的开销。其次，如果可运行的 Goroutine 比 P 多，进行这类系统调用的 Goroutine 无法重新获取 P 并且不得不把自己停放；释放 P 的瞬间，其他 Goroutine 就会被调度到上面。这是额外的运行时开销，有点不公平，并且不利于进行快速系统调度的目的（尤其是那些不进入内核的调用）。

所以 Go 运行时和调度器实际上有两种处理阻塞系统调用的方法，一种悲观方式，应用于预计会很慢的系统调用；另一种乐观方式，应用于预计会很快的系统调用。悲观的系统调用路径实现了直接的方法，运行时在系统调用前主动释放 P，之后尝试将 P 找回来，如果无法获取则停放自身。乐观的系统调用路径不会释放 P，相反，会设置一个特殊的 P 的状态标识并继续进行系统调用。一个特殊的内部 goroutine，sysmon goroutine，定期执行并寻找设置了这个“进行系统调用中”状态的时间太长了的 P，并将 P 从进行系统调用的 Goroutine 那里偷走。当系统调用返回，运行时代码检查它的 P 是否被偷走，如果没有则继续执行（如果 P 被偷走了的话，运行时会尝试获取其他的 P，如果失败可能会停放 goroutine）。

如果一切顺利，乐观的系统调用路径有着非常低的开销（大多数情况下，需要几个[原子比较和交换](https://en.wikipedia.org/wiki/Compare-and-swap)操作）。如果不顺利并且可运行的 Goroutine 的数量比 P 多，一个 P 会有不必要的空闲，通常可能是数十微秒（sysmon Goroutine 最多每 20 微秒运行一次，但如果似乎没有必要的话可以减少运行频率）。可能存在着最坏的情况，但是一般来说，在 Go 运行时方面这是一个值得的抉择。

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoSchedulerAndSyscalls

作者：[ChrisSiebenmann](https://twitter.com/thatcks/)
译者：[dust347](https://github.com/dust347)
校对：[JYSDeveloper](https://github.com/JYSDeveloper)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
