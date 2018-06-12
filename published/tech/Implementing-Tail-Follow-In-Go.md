已发布：https://studygolang.com/articles/12489

# 在 Go 中实现 tail 的跟踪功能

tail 是我们大多数人都熟悉的命令。我假设你也熟悉提供的 `-f` 选项。如果你不熟悉，知道它会打印出文件的最后几行即可。最近在一个项目上工作，我想知道我需要做什么来实现这个功能。这个想法来自阅读 [Feynman](http://amzn.to/2AIWVuX) 的书：

> 毫无疑问，你知道如何去做; 但是当你像小孩子一样玩这类问题，并且你没有看到答案时...试图找出如何去做是很有趣的。然后，当你进入成年时，你会培养出一定的自信，你可以去发现事物; 但是如果他们已经被发现，那你根本不应该再来打扰自己。一个傻瓜能做的事，另一个傻瓜也能做，其他一些傻瓜打你的事实不应该打扰你：你应该为将要发现的事物而快乐。

实现它可能是一件小事。但我认为这将是一系列文章中的一个良好的开端，在这篇文章中我写了如何实现一些东西。因此，让我带你看看如何在 Go 中实现 `tail -f` 的过程。

首先，让我们了解这个问题。tail 命令提供了一个标志，可以“跟踪”文件。它所做的是等待添加到文件末尾的任何更改并将其打印出来。为了简单起见，我不打算实现 tail，而只是实现跟踪功能。 所以我们在开始时打印整个文件，然后在添加它们时打印行。

我首先想到的是这样做的非常幼稚的方式。 打印文件中的所有字节，直到达到 `io.EOF`; 让这个过程睡一会儿，然后再试一次。 我们来看看这个函数：

```go
func follow(file io.Reader) error {
	r := bufio.NewReader(file)
	for {
		by, err := r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return  err
		}

		fmt.Print(string(by))
		if err == io.EOF {
			time.Sleep(time.Second)
		}
	}
}
```

当内容写入文件时，可悲的是没有及时做出反应。Linux 提供了一个 API 来监视文件系统事件：inotify AP。手册页给了你一个很好的介绍。它提供了两个我们感兴趣的函数：`inotify_init` 和 `inotify_add_watch`。`inotify_init` 函数创建一个对象，我们将使用该对象进一步与 API 进行交互。`inotify_add_watch` 函数允许你指定感兴趣的文件事件。API 提供了几个事件，但我们关心的是修改文件时发出的 `IN_MODIFY` 事件。

由于我们使用Go，不得不列出 `syscall` 包。它为前面提到的功能提供了包装器：`syscall.InotifyInit` 和 `syscall.InotifyAddWatch`。使用 syscall 让我们看看如何实现 follow 函数。为了简洁起见，我省略了错误处理，当你看到一个 `_` 变量被使用时，它是处理返回错误的好地方。

```go
func follow(filename string) error {
	file, _ := os.Open(filename)
	fd, _ := syscall.InotifyInit()
	_, _ := syscall.InotifyAddWatch(fd, filename, syscall.IN_MODIFY)
	r := bufio.NewReader(file)
	for {
		by, err := r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		fmt.Print(string(by))
		if err != io.EOF {
			continue
		}
		if err = waitForChange(fd); err != nil {
			return err
		}
	}
}

func waitForChange(fd int) error {
	for {
		var buf [syscall.SizeofInotifyEvent]byte
		_, _ := syscall.Read(fd, buf[:])
		if err != nil {
			return err
		}
		r := bytes.NewReader(buf[:])
		var ev = syscall.InotifyEvent{}
		_ = binary.Read(r, binary.LittleEndian, &ev)
		if ev.Mask&syscall.IN_MODIFY == syscall.IN_MODIFY {
			return nil
		}
	}
}
```

`InotifyInit` 函数返回一个可用于读取 `sycall.InotifyEvent` 的文件处理程序。从这个处理程序读取是一个阻塞操作。 这使我们只有在创建事件时才做出反应。

如果您要处理多个操作系统，最好更一般地处理这个操作系统。这就是 fsnotify 软件包的来源。它提供了一个针对 Linux 的 inotify，BSD 的 kqueue 等的抽象。使用 fsnotify 我们的函数看起来与前面的非常相似，但是更简单。

```go
func follow(filename string) error {
	file, _ := os.Open(filename)
	watcher, _ := fsnotify.NewWatcher()
	defer watcher.Close()
	_ = watcher.Add(filename)

	r := bufio.NewReader(file)
	for {
		by, err := r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		fmt.Print(string(by))
		if err != io.EOF {
			continue
		}
		if err = waitForChange(watcher); err != nil {
			return err
		}
	}
}

func waitForChange(w *fsnotify.Watcher) error {
	for {
		select {
		case event := <-w.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				return nil
			}
		case err := <-w.Errors:
			return err
		}
	}
}
```

我希望代码解释得比我写的文本更好。 为了简洁，我省略了代码可能失败的几种情况。 为了完善功能，我必须深入挖掘。 但这足以让我对其产生了知识的渴望：它到底是如何工作的。希望你也由此感觉。

---

via: [Implementing Tail's Follow In Go](http://satran.in/2017/11/15/Implementing_tails_follow_in_go.html)

作者：[Satyajit Ranjeev](http://satran.in/resume/)
译者：[shniu](https://github.com/shniu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
