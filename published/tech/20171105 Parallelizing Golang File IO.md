已发布：https://studygolang.com/articles/11915 

# 并行化 Golang 文件 IO

在这篇文章中，我们会使用一些 Go 的著名并行范例（Goroutine 和 WaitGroup），高效地遍历有大量文件的目录。所有代码都可以在 GitHub [这里](https://github.com/Tim15/golang-parallel-io)找到。

我正在开发一个项目，编写程序来将一个目录打包成一个文件。然后，我开始看 Go 的文件 IO 系统。其中貌似有几种遍历目录的方法。你可以使用 `filepath.Walk()`，或者你可以自己写一个。[有些人指出](https://github.com/golang/go/issues/16399)，与 `find` 相比，`filepath.Walk()` 真的很慢，所以我想知道，我能否写出更快的方法。我会告诉你我是怎么使用 Go 的一些很棒的功能来实现的。你可以将它们应用到其他问题上。

## 递归版本

唐纳德·克努特（Donald Knuth）曾经写道：“不成熟的优化是万恶的根源（premature optimization is the root of all evil.）”。遵循此建议，我们首先会用 Go 编写 `find` 的一个简单的递归版本，然后并行化它。

首先，打开目录：

```go
func lsFiles(dir string) {
	file, err := os.Open(dir)
	if err != nil {
		fmt.Println("error opening directory")
	}
	defer file.Close()
```

然后，获取这个文件中的子文件切片（Slice，也就是其他语言中的列表或数组）。

```go
files, err := file.Readdir(-1)
if err != nil {
	fmt.Println("error reading directory")
}
```

接着，我们将遍历这些文件，并再次调用我们的函数。

```go
	for _, f := range files {
		if f.IsDir() {
			lsFiles(dir + "/" + f.Name())
		}
		fmt.Println(dir + "/" + f.Name())
	}
}
```

可以看到，只有当文件是一个目录时，我们才会调用我们的函数，否则，只是打印出该文件的路径和名称。

## 初步测试

现在，让我们来测试一下。在一个带 SSD 的 MacBook Pro 上，使用 `time`，我获得以下结果：

```
$ find /Users/alexkreidler
	274165

real	0m2.046s
user	0m0.416s
sys	0m1.640s

$ ./recursive /Users/alexkreidler
	274165

real	0m13.127s
user	0m1.751s
sys	0m10.294s
```

并且将其与 `filepath.Walk()` 相比：

```go
func main() {
	err := filepath.Walk(os.Args[1], func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

```
./walk /Users/alexkreidler
	274165

real	0m13.287s
user	0m2.033s
sys	0m10.863s
```

## Goroutine

好了，是时候并行化了。如果我们试着将递归调用改为 goroutine，会怎样呢？

只是

```go
if f.IsDir() {
	lsFiles(dir + "/" + f.Name())
}
```

改成

```go
if f.IsDir() {
	go lsFiles(dir + "/" + f.Name())
}
```

哎呀，不好了！现在，它只是列出一些顶级文件。这个程序生成了很多 goroutine，但是随着 main 函数的结束，程序并不会等待 goroutine 完成。我们需要让程序等待所有的 goroutine 结束。

## WaitGroup

为此，我们将使用一个 `sync.WaitGroup`。基本上，它会跟踪组中的 goroutine 数目，保持阻塞状态直到没有更多的 goroutine。

首先，创建我们的 `WaitGroup`：

```go
var wg sync.WaitGroup
```

然后，我们会通过给这个 WaitGroup 加一，利用 goroutine 来启动递归函数.当 `lsFiles()` 结束，我们的 `main` 函数将会在 `wg` 为空之前都保持阻塞状态。

```go
wg.Add(1)
lsFiles(dir)
wg.Wait()
```

现在，为我们产生的每一个 goroutine 往 WaitGroup 加一：

```go
if f.IsDir() {
	wg.Add(1)
	go lsFiles(dir + "/" + f.Name())
}
```

然后，在我们的 `lsFiles` 函数尾部，调用 `wg.Done()` 来从 WaitGroup 减去一个计数。

```go
defer wg.Done()
```

好啦！现在，在它打印每一个文件之前，它应该会处于等待状态了。

## ulimits 和信号量 Channel

现在是棘手的部分。根据你的 CPU 以及 CPU 的内核数，你可能会也可能不会遇到这个问题。如果 Go 调度器有足够的内核可用，那么它可以充分加载 goroutine（[参考这里](https://stackoverflow.com/questions/8509152/max-number-of-goroutines)）。但是，多数的操作系统都会限制每个进程打开文件的数目。对于 unix 系统，这个限制是内核 `ulimits`。而在我的 Mac 上，该限制是 10,240 个文件，但是因为我只有 2 个内核，所以我不会受此影响。

在一台最近生产的有更多内核的计算机上，Go 调度器可能会同时创建超过 10,240 个 goroutine。每个 goroutine 都会打开文件，因此你会获得这样的错误：

`too many open files`

要解决这个问题，我们将使用一个信号量 channel：

```go
var semaphoreChan = make(chan struct{}, runtime.GOMAXPROCS(runtime.NumCPU()))
```

这个 channel 的大小限制为我们机器上的 CPU 或者核心数。

```go
func lsFiles(dir string) {
	// 满的时候阻塞
	semaphoreChan <- struct{}{}
	defer func() {
			// 读取以释放槽
			<-semaphoreChan
			wg.Done()
	}()
	...
```

当我们试图发送到这个 channel 时，将会被阻塞。然后当完成之后，从该 channel 读取以释放槽。详细信息，请参阅[这个 StackOverflow 帖子](https://stackoverflow.com/questions/38824899/golang-too-many-open-files-in-go-function-goroutine)。

## 测试和基准

```go
$ ./benchmark.sh
CPUs/Cores: 2
GOMAXPROCS: 2
find /Users/alexkreidler
	274165

real	0m2.046s
user	0m0.416s
sys	0m1.640s
./recursive /Users/alexkreidler
	274165

real	0m13.127s
user	0m1.751s
sys	0m10.294s
./parallel /Users/alexkreidler
	274165

real	0m9.120s
user	0m4.781s
sys	0m10.676s
./walk /Users/alexkreidler
	274165

real	0m13.287s
user	0m2.033s
sys	0m10.863s
```

## 总而言之

好啦，`find` 仍然是 IO 之王，但至少，我们的并行版本是对原始的递归版本和 `filepath.Walk()` 版本的改进。

希望这篇文章说明了如何利用 Go 中的一些强大的功能来构建并行系统。我们讨论了：

	* Goroutine
	* WaitGroup
	* Channel （信号量）

实际上，在 [github.com/golang/tools/imports/fastwalk.go](https://github.com/golang/tools/blob/master/imports/fastwalk.go) 上，Golang 有一个 `filepath.Walk` 的更快的实现，它的实现原理与本文相同。由于 `filepath` 包中的 API 保证，要在 Go 2.0 版本中才能修改它。

---

via: https://timhigins.ml/benchmarking-golang-file-io/

作者：[Timothy Higinbottom](https://timhigins.ml/about/)
译者：[ictar](https://github.com/ictar)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
