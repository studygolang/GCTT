### 使用 Go 读取文件 - 概览
2017 年 12 月 30 日

2018 年 1 月 1 日：[更新](http://kgrz.io/reading-files-in-go-an-overview.html#update)


<hr>

当我开始学习 Go 的时候，我很难熟练得运用各种操作文件的 API。在我尝试写一个多核心的计数器（[kgrz/kwc](https://github.com/kgrz/kwc)）时让我感到了困惑 - 操作同一个文件的不同方法。

在今年的 [Advent of Code](http://adventofcode.com/2017/) 中遇到了一些需要多种读取输入源的方式的问题。最终我把每种方法都至少使用了一次，因此现在我对这些技术有了一个清晰的认识。我会在这篇文章中将这些记录下来。我会按照我遇到这些技术的顺序列出来，而不是按照从易到难的顺序。

* 按字节读取 
  * 将整个文件读入内存中
  * 分批读取文件
  * 并行分批读取文件

* 扫描
  * 按单词扫描
  * 将一个长字符串分割成多个单词
  * 扫描用逗号分割的字符串

* Ruby 风格
  * 读取整个文件
  * 读取目录下的所有文件

* 更多帮助方法
* 更新

#### 一些基本的假设

* 所有的代码都包裹在 `main()` 代码块内
* 大部分情况下我会使用 "array" 和 "slice" 来指代 slices，但它们的含义是不同的。[这](https://blog.golang.org/go-slices-usage-and-internals)[两](https://blog.golang.org/slices)篇文章很好得解释了两者的不同之处。
* 我会把所有的示例代码上传到 [kgrz/reading-files-in-go](https://github.com/kgrz/reading-files-in-go)。

在 Go 中 - 对于这个问题，大部分的低级语言和一些类似于 Node 的动态语言 - 会返回字节流。之所以不自动返回字符串是因为可以避免昂贵的会增加垃圾回收器的压力的字符串分配操作。

为了让这篇文章更加通俗易懂，我会使用 `string(arrayOfBytes)` 来将 `字节` 数组转化为字符串，但不建议在生产模式中使用这种方式。

#### 按字节读取

*将整个文件读入内存中*

标准库里提供了众多的函数和工具来读取文件数据。我们先从 `os` 包中提供的基本例子入手。这意味着两个先决条件：

1. 该文件需要匹配内存
2. 我们需要预先知道文件大小以便实例化一个足够装下该文件的缓冲区

当我们获得了 `os.File` 对象的句柄，我们就可以事先查询文件的大小以及实例化一个字节数组。

```go
file, err := os.Open("filetoread.txt")
if err != nil {
  fmt.Println(err)
  return
}
defer file.Close()

fileinfo, err := file.Stat()
if err != nil {
  fmt.Println(err)
  return
}

filesize := fileinfo.Size()
buffer := make([]byte, filesize)

bytesread, err := file.Read(buffer)
if err != nil {
  fmt.Println(err)
  return
}

fmt.Println("bytes read: ", bytesread)
fmt.Println("bytestream to string: ", string(buffer))
```
[basic.go](https://github.com/kgrz/reading-files-in-go/blob/master/basic.go) on Github

#### 分批读取文件

大部分情况下我们都可以将这个文件读入内存，但有时候我们希望使用更保守的内存使用策略。比如读取一定大小的文件内容，处理它们，然后循环这个过程直到结束。在下面这个例子中使用了 100 字节的缓冲区。

```go
const BufferSize = 100
file, err := os.Open("filetoread.txt")
if err != nil {
  fmt.Println(err)
  return
}
defer file.Close()

buffer := make([]byte, BufferSize)

for {
  bytesread, err := file.Read(buffer)

  if err != nil {
    if err != io.EOF {
      fmt.Println(err)
    }

    break
  }

  fmt.Println("bytes read: ", bytesread)
  fmt.Println("bytestream to string: ", string(buffer[:bytesread]))
}
```
[reading-chunkwise.go](https://github.com/kgrz/reading-files-in-go/blob/master/reading-chunkwise.go) on Github

与读取整个文件的区别在于：

1. 当读取到 EOF 标记时就停止读取，因此我们增加了一个特殊的断言 `err == io.EOF`。如果你刚开始接触 Go，你可能会对 errors 的约定感到困惑，那么阅读 Rob Pike 的这篇文章可能会对你有所帮助：[Errors are values](https://blog.golang.org/errors-are-values)
2. 我们定义了缓冲区的大小，这样我们可以控制任意的缓冲区大小。由于操作系统的这种工作方式（[caching a file that’s being read](http://www.tldp.org/LDP/sag/html/buffer-cache.html)），如果设置得当可以提高性能。
3. 如果文件的大小不是缓冲区大小的整数倍，那么最后一次迭代只会读取剩余的字节到缓冲区中，因此我们会调用 `buffer[:bytesread]`。在正常情况下，`bytesread` 和缓冲区大小相同。

这种情况和以下的 Ruby 代码非常相似：

```ruby
bufsize = 100
f = File.new "_config.yml", "r"

while readstring = f.read(bufsize)
  break if readstring.nil?

  puts readstring
end
```
在循环中的每一次迭代，内部的文件指针都会被更新。当下一次读取开始时，数据将从文件指针的偏移量处开始，直到读取了缓冲区大小的内容。这个指针不是编程语言中的概念，而是操作系统中的概念。在 linux 中，这个指针是指创建的文件描述符的属性。所有的 read/Read 函数调用（在 Ruby/Go 中）都被内部转化为系统调用并发送给内核，然后由内核管理所有的这些指针。

#### 并行分批读取文件

那怎么样才能加速分批读取文件呢？其中一种方法是用多个 go routine。相对于连续分批读取文件，我们需要知道每个 go routine 的偏移量。值得注意的是，当剩余的数据小于缓冲区时，`ReadAt` 的表现和 `Read` 有[轻微的不同](https://golang.org/pkg/io/#ReaderAt)。

另外，我在这里并没有设置 go routine 数量的上限，而是由缓冲区的大小自行决定。但在实际的应用中通常都会设定 go routine 的数量上限。

```go
const BufferSize = 100
file, err := os.Open("filetoread.txt")
if err != nil {
  fmt.Println(err)
  return
}
defer file.Close()

fileinfo, err := file.Stat()
if err != nil {
  fmt.Println(err)
  return
}

filesize := int(fileinfo.Size())
// Number of go routines we need to spawn.
concurrency := filesize / BufferSize

// check for any left over bytes. Add one more go routine if required.
if remainder := filesize % BufferSize; remainder != 0 {
  concurrency++
}

var wg sync.WaitGroup
wg.Add(concurrency)

for i := 0; i < concurrency; i++ {
  go func(chunksizes []chunk, i int) {
    defer wg.Done()

    chunk := chunksizes[i]
    buffer := make([]byte, chunk.bufsize)
    bytesread, err := file.ReadAt(buffer, chunk.offset)

    // As noted above, ReadAt differs slighly compared to Read when the
    // output buffer provided is larger than the data that's available
    // for reading. So, let's return early only if the error is
    // something other than an EOF. Returning early will run the
    // deferred function above
    if err != nil && err != io.EOF {
      fmt.Println(err)
      return
    }

    fmt.Println("bytes read, string(bytestream): ", bytesread)
    fmt.Println("bytestream to string: ", string(buffer[:bytesread]))
  }(chunksizes, i)
}

wg.Wait()
```
[reading-chunkwise-multiple.go](https://github.com/kgrz/reading-files-in-go/blob/master/reading-chunkwise-multiple.go) on Github

这比之前的方法都需要考虑得更多：

1. 我尝试创建特定数量的 Go-routines， 这个数量取决于文件大小以及缓冲区大小（在我们的例子中是 100k）。
2. 我们需要一种方法能确定等所有的 go routines 都结束。在这个例子中，我们使用 wait group。
3. 我们在每个 go routine 结束时发送信号，而不是使用 `break` 从 for 循环中跳出。由于我们在 `defer` 中调用 `wg.Done()`，每次从 go routine 中”返回“时都会调用该函数。

注意：每次都应该检查返回的字节数，并刷新（reslice）输出缓冲区。

#### 扫描

