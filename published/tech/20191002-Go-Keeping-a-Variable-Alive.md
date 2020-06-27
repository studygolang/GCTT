首发于：https://studygolang.com/articles/28437

# Go: 延长变量的生命周期

![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-Keeping-a-Variable-Alive/00.png)

本文基于 Go 1.13。

在 Go 中，我们不需要自己管理内存分配和释放。然而，有些时候我们需要对程序进行更细粒度的控制。Go 运行时提供了很多种控制运行时状态及其与内存管理器之间相互影响的方式。本文中，我们来审查让变量不被 GC 回收的能力。

## 场景

我们来看一个基于 [Go 官方文档](https://golang.org/pkg/runtime/#KeepAlive) 的例子：

```go
type File struct { d int }

func main() {
	p := openFile("t.txt")
	content := readFile(p.d)

	println("Here is the content: "+content)
}

func openFile(path string) *File {
	d, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}

	p := &File{d}
	runtime.SetFinalizer(p, func(p *File) {
		syscall.Close(p.d)
	})

	return p
}

func readFile(descriptor int) string {
	doSomeAllocation()

	var buf [1000]byte
	_, err := syscall.Read(descriptor, buf[:])
	if err != nil {
		panic(err)
	}

	return string(buf[:])
}

func doSomeAllocation() {
	var a *int

	// memory increase to force the GC
	for i:= 0; i < 10000000; i++ {
		i := 1
		a = &i
	}

	_ = a
}
```

[源码地址](https://gist.githubusercontent.com/blanchonvincent/a247b6c2af559b62f93377b5d7581b7f/raw/6488ec2a36c28c46f942b7ac8f24af4e75c19a2f/main.go)

这个程序中一个函数打开文件，另一个函数读取文件。代表文件的结构体注册了一个 finalizer，在 gc 释放结构体时自动关闭文件。运行这个程序，会出现 panic：

```bash
panic: bad file descriptor

goroutine 1 [running]:
main.readFile(0x3, 0x5, 0xc000078008)
main.go:42 +0x103
main.main()
main.go:14 +0x4b
exit status 2
```

下面是流程图：

- 打开文件，返回一个文件描述符
- 这个文件描述符被传递给读取文件的函数
- 这个函数首先做一些繁重的工作：

![图 01](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-Keeping-a-Variable-Alive/01.png)

allocate 函数触发 gc：

![02.png](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-Keeping-a-Variable-Alive/02.png)

因为文件描述符是个整型，并以副本传递，所以打开文件的函数返回的结构体 `*File*` 不再被引用。Gc 把它标记为可以被回收的。之后触发这个变量注册的 finalizer，关闭文件。

然后，主协程读取文件：

![03.png](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-Keeping-a-Variable-Alive/03.png)

因为文件已经被 finalizer 关闭，所以会出现 panic。

## 让变量不被回收

`runtime` 包暴露了一个方法，用来在 Go 程序中避免出现这种情况，并显式地声明了让变量不被回收。在运行到这个调用这个方法的地方之前，gc 不会清除指定的变量。下面是加了对这个方法的调用的新代码：

![04.png](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-Keeping-a-Variable-Alive/04.png)

在运行到 `keepAlive` 方法之前，gc 不能回收变量 `p`。如果你再运行一次程序，它会正常读取文件并成功终止。

## 追本逐源

`keepAlive` 方法本身没有做什么：

![05.png](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-Keeping-a-Variable-Alive/05.png)

运行时，Go 编译器会以很多种方式优化代码：函数内联，死码消除，等等。这个函数不会被内联，Go 编译器可以轻易地探测到哪里调用了 `keepAlive`。编译器很容易追踪到调用它的地方，它会发出一个特殊的 SSA 指令，以此来确保它不会被 gc 回收。

![06.png](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-Keeping-a-Variable-Alive/06.png)

在生成的 SSA 代码中也可以看到这个 SSA 指令：

![07.png](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20191002-Go-Keeping-a-Variable-Alive/07.png)

在我的文章 [Go 编译器概述](https://medium.com/a-journey-with-go/go-overview-of-the-compiler-4e5a153ca889) 中你可以看到更多关于 Go 编译器和 SSA 的信息。

---
via: https://medium.com/a-journey-with-go/go-keeping-a-variable-alive-c28e3633673a

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
