已发布：https://studygolang.com/articles/12718

# Golang 中的回环栅栏

这篇文章中我们会研究一个基本的同步问题。并使用 Golang 中原生的 Buffered Channels 来为这个问题找到一个简洁的解决方案。

## 问题

现在假设我们我们有一堆 workers。为了充分发挥 CPU 多核的能力，我们让每个 worker 运行在单独的 goroutine 中：

```go
for i := 0; i < workers; i++ {
	go worker()
}
```

worker 需要做一系列的工作 job：

```go
func worker() {
	for i := 0; i < 3; i++ {
		job()
	}
}
```

每次 job 前都需要在所有的 worker 上同步地先进行一次准备 bootstrap 的过程。也就是说，每个 worker 在执行 job 前，需要等待所有其他 worker 都完成 bootstrap 的准备。

```go
func worker() {
	for i := 0; i < 3; i++ {
		bootstrap()
		# wait for other workers to bootstrap
		job()
	}
}
```

还有件事。如果至少有一个 worker 仍在执行 job，则所有 worker 的下一次的 bootstrap 都不能开始。换句话说，每次的 bootstrap 都是为紧接着的 job 部分做准备的，所以不能在上一次的 job 尚未结束之前就开始下一次的 bootstrap：

```go
func worker() {
	for i := 0; i < 3; i++ {
		# wait for all workers to finish previous loop
		bootstrap()
		# wait for other workers to bootstrap
		job()
	}
}
```

我们的 bootstrap 部分内容为增长一个共享的计数器。job 部分为等待一段时间并打印计数器的内容：

```go
type counter struct {
	c int
	sync.Mutex
}

func (c *counter) Incr() {
	c.Lock()
	c.c += 1
	c.Unlock()
}

func (c *counter) Get() (res int) {
	c.Lock()
	res = c.c
	c.Unlock()
	return
}

func worker(c *counter) {
	for i := 0; i < 3; i++ {
		# wait for all workers to finish previous loop
		c.Incr()
		# wait for other workers to do bootstrap
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		fmt.Println(c.Get())
	}
}
```

我们的目标是编写一个程序并以下列情形打印数字：

* 只有 n, `2*n` 和 `3*n` 的数字被打印（因为每个 worker 循环 3 次）
* 每次打印的数字不会比之前的数字小
* 每个数字会被打印 `n` 次

如果有 3 个 worker，则期望输出如下：

```
3
3
3
6
6
6
9
9
9
```

2 个 worker 的期望输出：

```
2
2
4
4
6
6
```

2 个 worker 不合法的输出可能会是这样：

```
2
4
2
4
6
6
```

想一想可能的解决办法。下面几行我故意留空并不急着给大家答案。

.

.

.

.

.

.

.

.

.

.

.

.

.

.

workers 会用一个名为回环栅栏（reusable barrier，类似 Java 中的 CyclicBarrier）的数据结构来实现同步。每个栅栏包含 2 扇门。第 1 扇门放置于增长计数器之前，一开始是关闭的。关闭的门意味着到达这扇门的 worker 会被阻塞。一旦所有的 worker 到达了第 1 扇门：

* 第 2 扇门（放置于增长计数器之后）会关闭
* 第 1 扇门开启

所有的计数器会通过并增长计数器，接着成功到达第 2 扇门。一旦所有的 worker 都到达了第 2 扇门：

* 第 1 扇门关闭
* 第 2 扇门开启

worker 此时可以开始执行 job 并接着在下一次循环中再次抵达第 1 扇门。循环再次开始。整个过程如图所示：

```

	 1st gate     2nd gate
		v             v
-w1-->  |             |
 --w2-->|
--w3--> |
 --w4-->|
-w5-->  |             |
 --w1-->|             |
 --w2-->|
 --w3-->|
 --w4-->|
 --w5-->|             |
		|      --w1-->|
			   --w2-->|
			--w3-->   |
			 --w4-->  |
		|      --w5-->|
		|      --w1-->|
			   --w2-->|
			   --w3-->|
			   --w4-->|
		|      --w5-->|
--w1--> |             |
		|              --w2-->
 --w3-->|
--w4--> |
		|             | --w5-->
```

下文有两种解决方案，实现略有不同。下面的代码可用来对两种方案进行测试：

```go
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
	// Set this import spec to point
	// to the copy of one of proposed
	// solutions.
	"path/to/package/barrier"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type counter struct {
	c int
	sync.Mutex
}

func (c *counter) Incr() {
	c.Lock()
	c.c += 1
	c.Unlock()
}

func (c *counter) Get() (res int) {
	c.Lock()
	res = c.c
	c.Unlock()
	return
}

func worker(c *counter, br *barrier.Barrier, wg *sync.WaitGroup) {
	for i := 0; i < 3; i++ {
		br.Before()
		c.Incr()
		br.After()
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		fmt.Println(c.Get())
	}
	wg.Done()
}

func main() {
	var wg sync.WaitGroup
	workers := 3
	br := barrier.New(workers)
	c := counter{}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(&c, br, &wg)
	}
	wg.Wait()
}
```

栅栏必须实现 *Before* 和 *After* 两个方法，各自对应第 1 和第 2 扇门。

## 解决方案 1

我们需要容量为 1 的 buffered channel：

```go
ch := make(chan int, 1)
```

门的逻辑可以通过先从 channel 中接收数据，然后再次向其中发送数据来实现：

```go
<-ch
ch <- 1
```

如果 channel 中包含元素数量为 1，则表示门是开的。它会让一个 worker 通过并往 channel 中放入新的元素以使另外一个 worker 通过，依此类推。

如果 channel 中没有元素了则表示门关闭了。接着 worker 从 channel 中接收元素就会被阻塞。

```go
// github.com/mlowicki/barrier
package barrier

import "sync"

type Barrier struct {
	c      int
	n      int
	m      sync.Mutex
	before chan int
	after  chan int
}

func New(n int) *Barrier {
	b := Barrier{
		n:      n,
		before: make(chan int, 1),
		after:  make(chan int, 1),
	}
	// close 1st gate
	b.after <- 1
	return &b
}

func (b *Barrier) Before() {
	b.m.Lock()
	b.c += 1
	if b.c == b.n {
		// close 2nd gate
		<-b.after
		// open 1st gate
		b.before <- 1
	}
	b.m.Unlock()
	<-b.before
	b.before <- 1
}

func (b *Barrier) After() {
	b.m.Lock()
	b.c -= 1
	if b.c == 0 {
	   // close 1st gate
	   <-b.before
	   // open 2st gate
	   b.after <- 1
	}
	b.m.Unlock()
	<-b.after
	b.after <- 1
}
```

## 解决方案 2

这个方案使用了容量和 worker 数量 *n* 相等的 buffered channel。现在我们不再让 worker 一个接一个地依次通过，而是在 channel 中放入 *n* 个元素来使所有的 worker 一次性通过：

```go
// github.com/mlowicki/barrier2
package barrier

import "sync"

type Barrier struct {
	c      int
	n      int
	m      sync.Mutex
	before chan int
	after  chan int
}

func New(n int) *Barrier {
	b := Barrier{
		n:      n,
		before: make(chan int, n),
		after:  make(chan int, n),
	}
	return &b
}

func (b *Barrier) Before() {
	b.m.Lock()
	b.c += 1
	if b.c == b.n {
		// open 2nd gate
		for i := 0; i < b.n; i++ {
			b.before <- 1
		}
	}
	b.m.Unlock()
	<-b.before
}

func (b *Barrier) After() {
	b.m.Lock()
	b.c -= 1
	if b.c == 0 {
		// open 1st gate
		for i := 0; i < b.n; i++ {
			b.after <- 1
		}
	}
	b.m.Unlock()
	<-b.after
}

```

## 参考

* "The Little Book of Semaphores" --- Allen B. Downey

---

via: https://medium.com/golangspec/reusable-barriers-in-golang-156db1f75d0b

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[alfred-zhong](https://github.com/alfred-zhong)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出


