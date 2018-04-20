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

你可以在各种场景下使用 `Read()` 方法来读取文件，但有时候你需要一些更加方便的方法。就像在 Ruby 中经常使用的类似于 `each_line`，`each_char`，`each_codepoint` 等 IO 函数。我们可以使用 `Scanner` 类型以及 `bufio` 包中的相关函数来达到类似的效果。

`buifo.Scanner` 类型实现了具有 “分割” 功能的函数，并基于该函数更新指针位置。比如内建的 `bufio.ScanLines` 分割函数，在每次迭代中将指针指向下一行第一个字符。在每一步中，该类型同时暴露一些方法来获得从起始位置到结束位置之间的字节数组/字符串。比如：

```go
file, err := os.Open("filetoread.txt")
if err != nil {
  fmt.Println(err)
  return
}
defer file.Close()

scanner := bufio.NewScanner(file)
scanner.Split(bufio.ScanLines)

// Returns a boolean based on whether there's a next instance of `\n`
// character in the IO stream. This step also advances the internal pointer
// to the next position (after '\n') if it did find that token.
read := scanner.Scan()

if read {
  fmt.Println("read byte array: ", scanner.Bytes())
  fmt.Println("read string: ", scanner.Text())
}

// goto Scan() line, and repeat
```
[scanner-example.go](https://github.com/kgrz/reading-files-in-go/blob/master/scanner-example.go) on Github

因此，若想按行读取整个文件，可以这么做：

```go
file, err := os.Open("filetoread.txt")
if err != nil {
  fmt.Println(err)
  return
}
defer file.Close()

scanner := bufio.NewScanner(file)
scanner.Split(bufio.ScanLines)

// This is our buffer now
var lines []string

for scanner.Scan() {
  lines = append(lines, scanner.Text())
}

fmt.Println("read lines:")
for _, line := range lines {
  fmt.Println(line)
}
```
[scanner.go](https://github.com/kgrz/reading-files-in-go/blob/master/scanner.go) on Github

#### 按单词扫描

`bufio` 包包含了一些基本的预定义的分割函数：

1. ScanLines（默认）
2. ScanWords
3. ScanRunes（在遍历 UTF-8 字符串而不是字节时将会非常有用）
4. ScanBytes

若想从文件中得到单词数组，则可以这么做：

```go
file, err := os.Open("filetoread.txt")
if err != nil {
  fmt.Println(err)
  return
}
defer file.Close()

scanner := bufio.NewScanner(file)
scanner.Split(bufio.ScanWords)

var words []string

for scanner.Scan() {
	words = append(words, scanner.Text())
}

fmt.Println("word list:")
for _, word := range words {
	fmt.Println(word)
}
```

`ScanBytes` 分割函数会返回和之前所说的 `Read()` 示例一样的结果。两者的主要区别在于在扫描器中每次我们需要添加到字节/字符串数组时存在的动态分配问题。我们可以用预先定义缓冲区大小并在达到大小限制后才增加其长度的技术来规避这种问题。示例如下：

```go
file, err := os.Open("filetoread.txt")
if err != nil {
  fmt.Println(err)
  return
}
defer file.Close()

scanner := bufio.NewScanner(file)
scanner.Split(bufio.ScanWords)

// initial size of our wordlist
bufferSize := 50
words := make([]string, bufferSize)
pos := 0

for scanner.Scan() {
  if err := scanner.Err(); err != nil {
    // This error is a non-EOF error. End the iteration if we encounter
    // an error
    fmt.Println(err)
    break
  }

  words[pos] = scanner.Text()
  pos++

  if pos >= len(words) {
    // expand the buffer by 100 again
    newbuf := make([]string, bufferSize)
    words = append(words, newbuf...)
  }
}

fmt.Println("word list:")
// we are iterating only until the value of "pos" because our buffer size
// might be more than the number of words because we increase the length by
// a constant value. Or the scanner loop might've terminated due to an
// error prematurely. In this case the "pos" contains the index of the last
// successful update.
for _, word := range words[:pos] {
  fmt.Println(word)
}
```
[scanner-word-list-grow.go](https://github.com/kgrz/reading-files-in-go/blob/master/scanner-word-list-grow.go) on Github

最终我们可以实现更少的 “扩增” 操作，但同时根据 `bufferSize` 我们可能会在末尾存在一些空的插槽，这算是一种折中的方法。

#### 将一个长字符串分割成多个单词

`bufio.Scanner` 有一个参数，这个参数是实现了 `io.Reader` 接口的类型，这意味着该类型可以是任何拥有 `Read` 方法的类型。在标准库中 `strings.NewReader` 函数是一个返回 “reader” 类型的字符串实用方法。
我们可以把两者结合起来使用：

```go
file, err := os.Open("_config.yml")
longstring := "This is a very long string. Not."
handle(err)

var words []string

scanner := bufio.NewScanner(strings.NewReader(longstring))
scanner.Split(bufio.ScanWords)

for scanner.Scan() {
	words = append(words, scanner.Text())
}

fmt.Println("word list:")
for _, word := range words {
	fmt.Println(word)
}
```

#### 读取逗号分隔的字符串

用基本的 `file.Read()` 或者 `Scanner` 类型去解析 CSV 文件/字符串显得过于笨重，因为在 `bufio.ScanWords` 函数中“单词”是指被 unicode 空格分隔的符号（runes）。读取单个符号（runes)，并持续跟踪缓冲区大小以及位置（就像 lexing/parsing 所做的）需要做太多的工作和操作。

当然，这是可以避免的。我们可以定义一个新的分割函数，这个函数读取字符知道遇到逗号，然后调用 `Text()` 或者 `Bytes()` 返回该数据块。`bufio.SplitFunc` 的函数签名如下所示：

```go
(data []byte, atEOF bool) -> (advance int, token []byte, err error)
```

1. `data` 是输入的字节字符串
2. `atEOF` 是传递给函数的结束符标志
3. `advance` 使用它，我们可以指定处理当前读取长度的位置数。此值用于在扫描循环完成后更新游标位置
4. `token` 是指扫描操作的实际数据
5. `err` 你可能想返回发现的问题

简单起见，我将演示读取一个字符串而不是一个文件。一个使用上述签名的简单的 CSV 读取器如下所示：

```go
csvstring := "name, age, occupation"

// An anonymous function declaration to avoid repeating main()
ScanCSV := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
  commaidx := bytes.IndexByte(data, ',')
  if commaidx > 0 {
    // we need to return the next position
    buffer := data[:commaidx]
    return commaidx + 1, bytes.TrimSpace(buffer), nil
  }

  // if we are at the end of the string, just return the entire buffer
  if atEOF {
    // but only do that when there is some data. If not, this might mean
    // that we've reached the end of our input CSV string
    if len(data) > 0 {
      return len(data), bytes.TrimSpace(data), nil
    }
  }

  // when 0, nil, nil is returned, this is a signal to the interface to read
  // more data in from the input reader. In this case, this input is our
  // string reader and this pretty much will never occur.
  return 0, nil, nil
}

scanner := bufio.NewScanner(strings.NewReader(csvstring))
scanner.Split(ScanCSV)

for scanner.Scan() {
  fmt.Println(scanner.Text())
}
```
[comma-separated-string.go](https://github.com/kgrz/reading-files-in-go/blob/master/comma-separated-string.go#L10) on Github

#### Ruby 风格

我们已经按照便利程度和功能一次增加的顺序列举了许多读取文件的方法。如果我们仅仅是想将一个文件读入缓冲区呢？标准库中的 `ioutil` 包包含了一些更简便的函数。

#### 读取整个文件

```go
bytes, err := ioutil.ReadFile("_config.yml")
if err != nil {
  log.Fatal(err)
}

fmt.Println("Bytes read: ", len(bytes))
fmt.Println("String read: ", string(bytes))
```

这种方式看起来更像一些高级脚本语言的写法。

#### 读取这个文件夹下的所有文件

无需多言， 如果你有很大的文件，*不要*运行这个脚本 :D

```go
filelist, err := ioutil.ReadDir(".")
if err != nil {
  log.Fatal(err)
}

for _, fileinfo := range filelist {
  if fileinfo.Mode().IsRegular() {
    bytes, err := ioutil.ReadFile(fileinfo.Name())
    if err != nil {
      log.Fatal(err)
    }
    fmt.Println("Bytes read: ", len(bytes))
    fmt.Println("String read: ", string(bytes))
  }
}
```

#### 更多的帮助方法

在标准库中还有很多读取文件的函数（确切得说，读取器）。为了防止这篇本已冗长的文章变得更加冗长，我列举了一些我发现的函数：

1. `ioutil.ReadAll()` -> 使用一个类似 io 的对象，返回字节数组
2. `io.ReadFull()
3. `io.ReadAtLeast()`
4. `io.MultiReader` -> 一个非常有用的合并多个类 io 对象的基本工具（primitive）。你可以把多个文件当成是一个连续的数据块来处理，而无需处理
在上一个文件结束后切换至另一个文件对象的复杂操作。

#### 更新

我尝试强调 “读取” 函数，我选择使用 error 函数来打印以及关闭文件：

```go
func handleFn(file *os.File) func(error) {
  return func(err error) {
    if err != nil {
      file.Close()
      log.Fatal(err)
    }
  }
}

// inside the main function:
file, err := os.Open("filetoread.txt")
handle := handleFn(file)
handle(err)
```

这样操作，我忽略了一个重要的细节：如果没有错误发生且程序运行结束，那文件就不会被关闭。如果程序多次运行且没有发生错误，则会导致文件描述符泄露。这个问题已经在 [on reddit by u/shovelpost](https://www.reddit.com/r/golang/comments/7n2bee/various_ways_to_read_a_file_in_go/drzg32k/) 中指出。

我之所以不想用 `defer` 是因为 `log.Fatal` 内部会调用 `os.Exit` 函数，而该函数不会运行 deferred 函数，所以我选择了手动关闭文件，然而却忽略了正常运行的情况。

我已经在更新了的例子中使用了 `defer`，并用 `return` 取代了 `os.Exit()`。


----------------

via: [原文链接](http://kgrz.io/reading-files-in-go-an-overview.html#update)

作者：[作者](作者链接)
译者：[Killernova](https://github.com/killernova)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出