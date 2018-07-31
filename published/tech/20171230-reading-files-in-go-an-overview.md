首发于：https://studygolang.com/articles/13593

# 概述 Go 中读取文件的方式

当我开始学习 Go 时，我很难掌握各种用于读取文件的 API 和技术。我尝试编写支持多核的单词计数程序（[KGRZ/KWC](https://github.com/kgrz/kwc)），通过在一个程序中使用多种读取文件方式来展示我初始的困惑。

在今年的 [Advent of Code](http://adventofcode.com/2017) 中，有些问题需要采用不同的方式来读取输入。我最终每种技术都至少使用过一次，现在我将对这些技术的理解写在本文中。我列出的方法是按照我使用的顺序，并不一定按照难度递减的顺序。

## 一些基本的假设

* 所有的代码示例都被封装在一个 `main()` 函数中
* 大多数情况下，我会经常会交替使用“数组 `array`”和“切片 `slice`”来指代切片，但它们是不一样的。这些[博客](https://blog.golang.org/go-slices-usage-and-internals)[文章](https://blog.golang.org/slices)是了解差异的两个很好的资源。
* 我把所有的实例上传到[kgrz/reading-files-in-go](https://github.com/kgrz/reading-files-in-go)。

在 go 中像大多数低级语言和一些动态语言（例如Node）中一样，读取文件时返回一个字节流。不自动将读取内容转换为字符串有一个好处，可以避免因为昂贵的字符串分配给 GC 带来的压力。

为了让这篇文章有一个简单的概念模型，我会使用 `string(arrayOfBytes)` 将字节数组转换成字符串。不过一般来说不建议在生产环境使用这种方式。

## 按字节读取

### 读取整个文件到内存

首先，标准库提供多个函数和工具来读取文件数据。我们从 `os` 包中提供的一个基本用法开始。这意味着两个先决条件：

1. 文件大小不超过内存。
2. 我们能够提前知道文件的大小，以便实例化一个足够大的缓冲区来保存数据。

获得一个 `os.file` 对象的句柄，我们可以获取其大小并实例化一个字节类型切片。

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

### 以块读取文件

在大多数情况下，一次性读取整个一个文件是没有问题的。有时我们希望使用更节省内存的方法。比如说，按照一定大小来读取一个文件块，并处理这个文件块，然后重复直到读取完整个文件。

下面的示例使用100字节大小的缓冲区。

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

与读取整个文件内容相比，主要不同之处在于：

1. 我们持续进行读取，直到读到 `EOF` 标记，所以我们添加了一个特定的检查 `err==io.EOF`。如果你是 Go 的新手，并且对错误处理的方法感到困惑，请查看这篇由 Rob Pike 写的文章：[Errors are values](https://blog.golang.org/errors-are-values)
2. 我们定义了缓冲区大小，这样我们就可以控制我们想要的“块”大小。如果使用得当，这可以提高性能，因为操作系统的工作方式是缓存正在读取的文件。
3. 如果文件大小不是缓冲区大小的整数倍，则最后一次迭代只向缓存中添加余下的字节，因此需要切片操作 `buffer[:bytesread]`。在正常情况下，`bytesread` 和缓存大小相同。

这和下面这段 Ruby 代码类似：

```
bufsize = 100
f = File.new "_config.yml", "r"

while readstring = f.read(bufsize)
    break if readstring.nil?
    puts readstring
end
```

在每个循环中，都对内部文件指针位置进行更新。当下一次读取时，数据从文件指针偏移开始，读取并返回缓冲区大小的数据。这个指针不是由编程语言创建的，而是操作系统创建的。在 Linux 上，这个指针是操作系统创建的文件描述符。所有 `read/Read` 调用（分别在 Ruby/Go 中）被内部翻译成系统调用,并发送到内核，由内核管理这个指针。

### 并发读取文件块

如果我们想加快上面提到的对数据块的处理速度呢？一个方法就是使用多个 `goroutine`！相对于顺序读取数据块，我们需要一个额外的操作就是要知道每个 `routine` 读取数据的偏移量。注意，`ReadAt` 函数和 `Read` 函数在缓存容量大于剩余需要读取字节的时，处理方式略有不同。

还要注意的是，我并没有限制 `goroutine` 的数量，它只是由缓冲区大小决定的。事实上，这个数字可能有一个上限。

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
// Number of goroutines we need to spawn.
concurrency := filesize / BufferSize

// check for any left over bytes. Add one more goroutine if required.
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

这比以前任何方法都要复杂：

1. 我尝试创建一个特定的 `goroutine`，这取决于文件大小和缓冲区大小（在我们的例子中是100）。
2. 我们需要一种方法来确保我们“等待”所有的 `goroutine` 运行完成。在这个例子中，我是使用 `WaitGroup`。
3. 我们在 `goroutine` 运行完成时发送一个结束信号，而不是使用无限循环等待运行结束。我们使用 `defer` 调用 `wg.Done()`，当 `goroutine` 运行到 `return` 时，`wg.Done` 会被调用。

注意：总是要检查返回的字节数，并对输出缓冲区重新切片。

## 扫描文件

你可以一直使用 `Read()` 来读取文件，但有时你需要更方便的方法。在 Ruby 中有一些经常用到的 IO 函数，比如 `each_line`，`each_char`，`each_codepoint` 等。我们可以使用 `Scanner` 类型和 `bufio` 包中提供的相关函数来实现类似的功能。

`bufio.Scanner` 类型实现了一个参数为“分割”函数的函数，并基于此函数推进指针。例如，内置的 `bufio.ScanLines` 分割函数，在每次迭代中都会推进指针，直到指针推进到下一个换行符。在每个步骤中，`bufio.Scanner` 类型提供了获取在起始位置和结束位置之间的字节数组/字符串的函数。例如：

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

因此，使用这种逐行的方式读取整个文件，可以使用以下代码：

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

### 按照单词扫描

`bufio` 包包含几个基本预定义的分割函数：

1. ScanLines （默认）
2. ScanWords
3. ScanRunes （在处理 UTF-8 编码时非常有用）
4. ScanBytes

所以，读取一个文件，按照单词分割并生成一个列表，可以使用以下代码：

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

`ScanBytes` 分割函数将给出与我们之前使用 `Read()` 示例中相同的输出。两者之间的一个主要区别是每次我们都需要动态的将数据追加到 byte/string 数组。这可以通过使用预先初始化缓存的技术来规避，只在数据长度超出缓冲区时才增加缓冲区大小。使用上面的相同的例子：

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

所以我们显著减少了切片“增长”操作，但是根据缓存大小和文件大小，我们可能会在缓存末尾有空缺，这是一个折衷的方案。

### 将长字符串分割成单词

`bufio.NewScanner` 需要满足 `io.Reader` 接口的类型作为参数，这意味着它可以接受任何有 `Read` 方法的类型作为参数。标准库中字符串实用方法 `strings.NewReader` 函数返回一个 “reader” 类型。我们可以把两者结合起来，实现长字符串分割成单词：

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

### 扫描逗号分隔字符串

用基本的 `Read()` 函数或 `Scanner` 类型手动解析 CSV 文件/字符串是比较繁琐的，因为上述分割函数 `bufio.ScanWords` 将一个“单词”定义为一组由空格分割的字符。读取单个字符并记录缓冲区大小和位置（像词法分析/解析工作）需要太多的工作和操作。

我们可以通过定义新的分割函数来省去这些繁琐的操作。分割函数顺序读取每个字符直到遇到逗号，然后在 `Text()` 或 `Bytes()` 函数被调用时返回检测到的单词。`bufio.SplitFunc` 函数签名应该是这样的：

```
(data []byte, atEOF bool) -> (advance int, token []byte, err error)
```

1. `data` 是输入的字节串
2. `atEOF` 是表示输入数据是否结束的标志
3. `advance` 用于根据当前读的长度来确定指针推进值，使用这个值在循环扫描完成后更新数据指针的位置。
4. `token` 是扫描操作后得到的数据
5. `err` 用于返回错误信息

为了简单起见，我展示了一个读取字符串的例子。实现上述函数签名的简单读取器来读取 CSV 字符串：

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

## Ruby风格

我们已经按照方便性和效率的顺序看到了多种方法来读取文件。但是，如果你只想把文件读入缓冲区呢？ `ioutil` 是标准库中的一个包，其中的函数能够使用一行代码完成一些功能。

### 读取整个文件

```go
bytes, err := ioutil.ReadFile("_config.yml")
if err != nil {
  log.Fatal(err)
}

fmt.Println("Bytes read: ", len(bytes))
fmt.Println("String read: ", string(bytes))
```

这更接近我们在高级脚本语言中看到的写法。

### 读取整个目录的文件

不必多说，如果你有大文件，**不要** 运行这个脚本:D

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

### 其他有用的函数

在标准库中有更多的函数来读取文件（或者更准确的说是一个 `Reader`）。为了避免这篇文章过长，我列出了我发现的一些函数：

1. `ioutil.ReadAll()` 输入一个类似 `io` 对象，将整个数据作为字节数组返回
2. `io.ReadFull()`
3. `io.ReadAtLeast()`
4. `io.MultiReader` 组合多个类似 `io` 对象时非常有用。如果你有一个需要读取的文件列表，可以将它们视为单个连续的数据块，而无需管理复杂的前后文件之间的切换。

### 更新

为了突出显示 “read” 函数，我选择了使用错误处理函数来打印错误并关闭文件：

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

这样做，我错过了一个关键的细节：当没有发生错误并且程序运行完成时，我没有关闭文件句柄。如果程序运行多次而没有发生任何错误，则会导致文件描述符泄漏。这是由[u/shovelpost](https://www.reddit.com/r/golang/comments/7n2bee/various_ways_to_read_a_file_in_go/drzg32k/)在reddit上指出的。

我本意是避免使用 `defer`，因为 `log.Fatal` 在内部调用了不运行延迟函数的 `os.Exit`，所以我选择显式关闭文件，但忽略了成功运行的情况。

我已经更新了示例使用 `defer` 和 `return` 来代替对 `os.Exit` 的依赖。

---

via: https://kgrz.io/reading-files-in-go-an-overview.html

作者：[Kashyap Kondamudi](http://github.com/kgrz/)
译者：[alan](https://github.com/althen)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
