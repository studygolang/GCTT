# goroutine 的同步（第三部分）

mutex 和 sync.Once 介绍

![image](https://cdn-images-1.medium.com/max/1600/1*CL4qTBN5ulMH7eYMDBBWHA.jpeg)

假设你的程序中有一个需要某种初始化的功能。该 bootstrap 程序成本很高，因此将其推迟到实际使用功能的那一刻是有意义的。这样，当功能未激活时，就不会浪费 CPU 周期。 这在 Go 中如何完成？

```go
package main

import "fmt"

var capitals map[string]string

func bootstrap() {
    capitals = make(map[string]string)
    capitals["France"] = "Paris"
    capitals["Germany"] = "Berlin"
    capitals["Japan"] = "Tokyo"
    capitals["Brazil"] = "Brasilia"
    capitals["China"] = "Beijing"
    capitals["USA"] = "Washington"
    ...
    capitals["Poland"] = "Warsaw"
}

func getCapitalCity(country string) string {
    if capitals == nil {
        bootstrap()
    }
    return capitals[country]
}

func main() {
    fmt.Println(getCapitalCity("Poland"))
    fmt.Println(getCapitalCity("USA"))
    fmt.Println(getCapitalCity("Japan"))
}
```

你可以想象，如果它可以处理所有的国家，其他类似数据库的结构，使用 I/O 操作等，则 *bootstrap* 函数可能非常昂贵。上述解决方案看起来简单且优雅，但不幸的是，它并不会正常工作。问题在于当 *bootstrap* 函数在运行时无法阻止其它 goroutine 做同样的事。而这些繁重的计算只做一次是很可取的。另外在 *capitals* 刚刚被初始化而其中的 key 还未设置时，其它 goroutine 会看到它不为 *nil* 从而尝试从空的 map 中获取值。

## [sync.Mutex](https://golang.org/pkg/sync/#Mutex)

![image](https://cdn-images-1.medium.com/max/1600/1*_i7pBza2IqO7mIyNRbrfzw.jpeg)

Go 有包含很多好东西的内置 [sync](https://golang.org/pkg/sync/) 包。我们可以使用 [mutex](https://en.wikipedia.org/wiki/Mutual_exclusion) (mutual exclusion) 来解决我们的同步问题。

```go
import (
    "fmt"
    "sync"
)

...

var (
    capitals map[string]string
    mutex sync.Mutex
)

...

func getCapitalCity(country string) string {
    mutex.Lock()
    if capitals == nil {
        bootstrap()
    }
    mutex.Unlock()
    return capitals[country]
}
```

bootstrap 程序可被多次运行的问题解决了。如果任何一个 goroutine 正在运行 *bootstrap* 或者甚至在判断 `capitals == nil`，那么其它 goroutine 会在 `mutex.Lock()` 处等待。*Unlock* 函数一结束另一个 goroutine 就会被“准许进入”。

但是一次只能有一个 goroutine 执行被放在 `mutex.Lock()` 和 `mutex.Unlock()` 之间的代码。因此如果存放首都城市的 map 被多个 goroutine 读取，那么一切都会在 if 语句处被处理成一个接一个的。从根本上说对于 map 的读操作（包括判断它是否为 *nil*）应该允许一次处理多个，因为它是线程安全的。

## [sync.RWMutex](https://golang.org/pkg/sync/#RWMutex)

读/写 mutex 可以同时被多个 reader 或者一个 writer 持有（writer 是指改写数据的某种东西）：

```go
mutex sync.RWMutex

...

func getCapitalCity(country string) string {
    mutex.RLock()
    if capitals != nil {
        country := capitals[country]
        mutex.RUnlock()
        return country
    }
    mutex.RUnlock()
    mutex.Lock()
    if capitals == nil {
        bootstrap()
    }
    mutex.Unlock()
    return getCapitalCity(country)
}
```

现在代码变得复杂多了。第一部分使用读锁以允许多个 reader 同时读取 *capitals*。在一开始 bootstrap 还未完成，所以调用者会拿到 `mutex.Lock()` 并且做必要的初始化。当这部分结束，函数可以被再次调用来获取期望的值。这些都在 bootstrap 函数已经返回之后。

最新的解决方法很显然维护起来比较难。幸运的是有一个内置的方法恰好帮我们应对这种情况……

## [sync.Once](https://golang.org/pkg/sync/#Once)

```go
once sync.Once

...

func getCapitalCity(country string) string {
    once.Do(bootstrap)
    return capital[country]
}
```

上面的代码比第一个简明版的解决方法（它并不正常工作）更加简单。可以保证 *bootstrap* 只会被调用一次，并且从 map 中读数据仅在 *bootstrap* 返回后才会被执行。

---

点赞以帮助别人发现这篇文章。如果你想得到新文章的更新，请关注我。

## 资源

- [Go 的内存模型 —— Go 编程语言](https://golang.org/ref/mem)

>The Go memory model specifies the conditions under which reads of a variable in one goroutine can be guaranteed to…
<br>*golang.org*

- [sync - Go 编程语言](https://golang.org/pkg/sync/)
>Package sync provides basic synchronization primitives such as mutual exclusion locks. Other than the Once and…
<br>*golang.org*

*[保留部分版权](http://creativecommons.org/licenses/by/4.0/)*

*[Golang](https://medium.com/tag/golang?source=post)*
*[Programming](https://medium.com/tag/programming?source=post)*
*[Concurrency](https://medium.com/tag/concurrency?source=post)*
*[Synchronization](https://medium.com/tag/synchronization?source=post)*
*[Software Development](https://medium.com/tag/software-development?source=post)*

**喜欢读吗？给 Michał Łowicki 一些掌声吧。**

简单鼓励下还是大喝采，根据你对这篇文章的喜欢程度鼓掌吧。

---

via: https://medium.com/golangspec/synchronized-goroutines-part-iii-c60bcfeefd2a

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[krystollia](https://github.com/krystollia)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
