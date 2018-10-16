已发布：https://studygolang.com/articles/12133

# 获取 Goroutine ID

## Goroutine ID 真实存在吗？

当然存在。

Go 运行时一定有某种方法来跟踪 goroutine ID。

## 那我该使用它们吗？

不该。

- 原因一：https://groups.google.com/forum/#!topic/golang-nuts/Nt0hVV_nqHE
- 原因二：https://groups.google.com/forum/#!topic/golang-nuts/0HGyCOrhuuI
- 原因三：http://stackoverflow.com/questions/19115273/looking-for-a-call-or-thread-id-to-use-for-logging

## 有没有哪些包是我可以使用的？

已有的来自 Go Team 成员的包，被评价为“[用此包者，将入地狱。](https://godoc.org/github.com/davecheney/junk/id)”

也有一些包基于 goroutine id 来建立 goroutine 本地存储，如：

- [github.com/jtolds/gls](https://github.com/jtolds/gls)
- [github.com/tylerb/gls](https://github.com/tylerb/gls)

但都有悖于 Go 语言的设计原则。

## 最简代码

如果读到这里，你仍“执迷不悟”，那么下面就将展示如何获取当前的 goroutine id ：

### Go 源码中的骇客（Hacky）代码

下列代码源于 Brad Fitzpatrick 的 [`http/2`](https://github.com/golang/net/blob/master/http2/gotrack.go) 库。它被整合进了 Go 1.6 中，仅仅被用于调试而非常规开发。

```go
package main

import (
    "bytes"
    "fmt"
    "runtime"
    "strconv"
)

func main() {
    fmt.Println(getGID())
}

func getGID() uint64 {
    b := make([]byte, 64)
    b = b[:runtime.Stack(b, false)]
    b = bytes.TrimPrefix(b, []byte("goroutine "))
    b = b[:bytes.IndexByte(b, ' ')]
    n, _ := strconv.ParseUint(string(b), 10, 64)
    return n
}
```

#### 工作原理解释

通过解析调试信息来获取 goroutine id 是可行的. `http/2` 库就使用调试性的代码来对连接进行追踪查看。但仅仅是将 goroutine id 用于调试而已。

调试信息可以通过调用 [`runtime.Stack(buf []byte, all bool) int`](https://golang.org/pkg/runtime/#Stack) 来获取，它会以文本形式打印堆栈信息到缓冲区中。堆栈信息的第一行会是如下文本： “goroutine #### […” 。这里的 #### 就是真实的 goroutine id。剩余代码不过是进行一些文本操作来提取和解析堆栈信息中的数字。

### CGo 版本对应的合法代码

C 版本的代码来自 [github.com/davecheney/junk/id](https://github.com/davecheney/junk/tree/master/id)。代码中直接获取了当前 goroutine 的 goid 属性并返回它的值。

文件 `id.c`

```c
#include "runtime.h"

int64 ·Id(void) {
	return g->goid;
}
```

文件 `id.go`

```go
package id

func Id() int64
```

## 我该怎么做？

远离 goroutine id 吧，并忘记它们的存在。从 Go 语言设计的角度来看，使用它们是危险的。因为几乎所有使用的目的都是去做一些和 goroutine-local 相关的事情。而这违反了 Go 语言编程的 “[Share Memory By Communicating](https://blog.golang.org/share-memory-by-communicating)” 原则。

---

via: http://blog.sgmansfield.com/2015/12/goroutine-ids/

作者：[Scott Mansfield](http://blog.sgmansfield.com/)
译者：[MaleicAcid](https://github.com/MaleicAcid)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
