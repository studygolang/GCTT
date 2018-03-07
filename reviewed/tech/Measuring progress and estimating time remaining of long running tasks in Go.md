## 对Go中长时间运行io.Reader和io.Writer的操作测算进度和估算剩余时间

​                                                                   				 		 Mat Ryer



![img](https://cdn-images-1.medium.com/max/900/1*YfQ0FQIK4l6NMW3wsl9NNw.jpeg)

每当我们在使用类似 io.Copy 和  ioutil.ReadAll 的工具时，比如我们正在从 http.Response 主体读入或者上传一个文件，我们会发现这些方法将一直堵塞，直到整个过程完成，哪怕耗时数十分钟甚至是小时——而且我们没有办法来查看进度，以及计算出完成所需剩余时间的估测值。

本文很长，不想深究瞅这里：这篇文章最终导向 progress 包，你可以在自己的项目中自由使用——https://github.com/machinebox/progress

考虑到 io.Reader 和 io.Writer 都是接口，我们可以封装它们并且拦截 Read 和 Write 方法，捕获实际已经通过它们的字节数。通过一些简单的数学计算，我们可以计算出已完成部分所占的比例。再多上一点数学计算，我们甚至可以估测整个过程还剩余多少时间，假设传输流是相对一致的话。

## 封装 Reader

一个新的 Reader 类型只需要包含另一个 io.Reader , 并且调用它的 Read 方法来获取返回前读到的字节数。为了保证 reader 可以在并发环境中安全使用（在这个例子中至关重要），我们可以使用 atomic.AddInt64 作为安全的计数器。

```go
// Reader ：计数通过它读取的字节数。
type Reader struct {
 r io.Reader
 n int64
}
// NewReader 返回一个可以计数通过它读取到字节数的
// Reader
func NewReader(r io.Reader) *Reader {
 return &Reader{
 r: r,
 }
}
func (r *Reader) Read(p []byte) (n int, err error) {
 n, err = r.r.Read(p)
 atomic.AddInt64(&r.n, int64(n))
 return
}
// N 表示目前为止读取到的字节数
func (r *Reader) N() int64 {
 return atomic.LoadInt64(&r.n)
}

```

试试看你能不在自己写出 Writer 的计数部分，两者很类似。

由于方法 N 返回（ 基于 atomic.LoadInt64 的安全调用）读取到的字节数，我们能在任意时刻使用另一个 goroutine 调用它，从而获取当前状况。

## 获取总共的字节数

为了计算百分比，我们需要知道总数是多少——我们预期读取多少字节？

上传文件时，我们能够利用操作系统获取文件大小。

```go
info, err := os.Stat(filename)
if err != nil {
 return errors.Wrap(err, "cannot get file info")
}
size := info.Size(
```

在 HTTP 环境中，你可以借助下面这些代码来获取 Content-Length 报头值。

```go
contentLengthHeader := resp.Header.Get("ContentLength")
size, err := strconv.ParseInt(contentLengthHeader, 10,
64)
if err != nil {
 return err
}
```

如果 Content-Length 报头是空的（这有可能），那么就无法判断进度或者估计剩余时间。

在其他状况下，你也会需要弄清楚如何获取字节总数。

## 计算百分比

现在我们可以计算已经被处理的字节数所占百分比：

```go
func percent(n, size float64) float64 {
 if n == 0 {
 return 0
 }
 if n >= size {
 return 100
 }
 return 100.0 / (size / n ）
}
```

我们需要把值转换为 float64 从而避免早期的向下取整。如果需要整数级精度的话我们依然可以把结果向下取整。

## 估算剩余时间

有一个非常简单的方法：求出读取 X 字节所需时间，然后乘以剩余的字节数。

举个例子，如果耗时 10 秒完成了 50% 的操作，那么就可以假设仍需要 10 秒来完成整个任务；总耗时 20 秒。

这并不绝对精确，但大多时候都可以给出一个可采用的倒计时。

代码就在下面，但不需要担心你可能理解不了 —— 阅读我们的 package 下面的详细信息可以帮到你。

```go
// 开始时...
started := time.Now()
// 每次我们想查看时...
ratio := n / size
past := float64(time.Now().Sub(started))
total := time.Duration(past / ratio)
estimated := started.Add(total)
duration := estimated.Sub(time.Now())
```

 

- `ratio`  — 已经完成字节数所占的百分比
- `past`  — 从开始到现在的耗时
- `total` — 基于已完成的百分比 ratio 和相应耗费的时间，从而得出的预计总耗时
- `estimated`  — 预测的结束时间点
- `duration` — 预测距离完成还需要耗费的时间

## 浏览 progess 包

![img](https://cdn-images-1.medium.com/max/1200/1*zjDaQfSU9YYY4WIz0K5CxA.png)

我们热爱开源，所以我们封装了所有代码到一个  [package](https://github.com/machinebox/progress)  中以方便您的使用。

它也支持 io.EOF 和其他你知道的可能会在操作时发生的错误。

### 小助手

我们还添加了一个小助手，它可以给你一个进度上的 go channel 来周期性报告。  你可以开启一个新的 goroutine 并打印进度，或更新进度，这取决于您的用例。

```go
ctx := context.Background()

// 得到一个 reader 和字节总数
s := `Now that's what I call progress`
size := len(s)
r := progress.NewReader(strings.NewReader(s))

// 开启一个 goroutine 打印进度
go func() {
    progressChan := progress.NewTicker(ctx, r, size, 1*time.Second)
    for p := range <-progressChan {
        fmt.Printf("\r%v remaining...", 
                   p.Remaining().Round(time.Second))
    }
    fmt.Println("\rdownload is completed")
}()

// 使用 reader
if _, err := io.Copy(dest, r); err != nil {
	log.Fatalln(err)
}
```

该 channel 会周期性的返回一个  [Progress](https://godoc.org/github.com/machinebox/progress#Progress)  结构体，该结构体有下列几个方法帮助你了解细节。

- `Percent` — 获取操作完成的百分比
- `Estimated` —  `time.Time` 表示预期操作结束的时间点
- `Remaining` — 一个 `time.Duration` 变量标识剩余时间

 channel 会在几种情况下被关闭，例如操作已完成，或者操作被取消。

[点击文档](https://godoc.org/github.com/machinebox/progress) 可以获取 API 的最新详细目录

### 示例

我们创建了一个  [example file downloader](https://github.com/machinebox/progress/blob/master/example/download/main.go) 来演示该 package 如何使用。

## 还有什么？

请尝试我们的开源项目，提出问题，报告议题，提交重要的 PR 。

## 什么是 Machine Box ？

![img](https://cdn-images-1.medium.com/max/1200/1*GPdHUaxzqp2dJYd0l_hwcA.jpeg)

[Machine Box](https://machinebox.io/?utm_source=blog&utm_medium=medium&utm_campaign=matblog) 把先进的机器学习技术放到 Docker 容器中，以便让开发人员可以更轻松的集成

自然语言处理，面部检测，对象识别等技术到你自己的应用中。

该技术是按比例构建，所以当你的应用扩大时只需要添加更多同级的 box 。噢，而且它比云服务廉价的多（可能还会更好）……而且你的数据也不会离开你自己的基础设备。

[玩一玩](https://machinebox.io/docs/facebox/teaching-facebox) , 并且请告知我们您宝贵的意见。



via: <https://blog.machinebox.io/measuring-the-progress-of-long-running-io-reader-and-io-writer-operations-in-go-ba26b204a507>

作者：[Mat Ryer](https://blog.machinebox.io/@matryer)
译者：[sunzhaohao](https://github.com/sunzhaohao)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出