首发于：https://studygolang.com/articles/25291

# 用 70 行 Go 代码击败 C 语言

Chris Penner 最近发布的一篇文章 [Beating C with 80 Lines of Haskell](https://chrispenner.ca/posts/wc) 引发了 Internet 领域内广泛的论战，进而引发了一场用不同语言实现 `wc` 的圣战：

- [Ada](http://verisimilitudes.net/2019-11-11)
- [C](https://github.com/expr-fi/fastlwc/)
- [Common Lisp](http://verisimilitudes.net/2019-11-12)
- [Dyalog APL](https://ummaycoc.github.io/wc.apl/)
- [Futhark](https://futhark-lang.org/blog/2019-10-25-beating-c-with-futhark-on-gpu.html)
- [Haskell](https://chrispenner.ca/posts/wc)
- [Rust](https://medium.com/@martinmroz/beating-c-with-120-lines-of-rust-wc-a0db679fe920)

今天我们用 Go 语言来实现 `wc` 的功能。作为有着杰出的并发基因的语言，实现与 C 语言相当的性能（原文为 [comparable performance](https://benchmarksgame-team.pages.debian.net/benchmarksgame/fastest/go-gcc.html)）应该是小菜一碟。

虽然 `wc` 也被设计为从 stdin 读取信息，处理 non-ASCII 文本编码，从命令行解析 flags（[manpage](https://ss64.com/osx/wc.html)），但我们并不去这样做。我们要做的是，像前面提到的那篇文章一样，让我们的实现尽可能简单。

本文涉及的源码可以在 [这里](https://github.com/ajeetdsouza/blog-wc-go) 找到。

```bash
$ /usr/bin/time -f "%es %MKB" wc test.txt
```

我们使用 [与原文相同版本的 `wc`](https://opensource.apple.com/source/text_cmds/text_cmds-68/wc/wc.c.auto.html) ，用 gcc 9.2.1 编译，编译优化选项为 `-O3`。在我们的实现中，使用 Go 1.13.4（我确实也试过用 gccgo，但结果不是很理想）。我们用以下配置来运行所有的基准：

- Intel Core i5-6200U @ 2.30 GHz (2 physical cores, 4 threads)
- 4+4 GB RAM @ 2133 MHz
- 240 GB M.2 SSD
- Fedora 31

公平起见，所有实现都使用一个 16 KB 的 buffer 来读取输入。输入是两个 us-ascii 编码的文本文件，大小分别是 100 MB 和 1 GB。

## 一个纯朴的方法

因为我们只需要输入文件路径，所以解析参数很容易：

```go
if len(os.Args) < 2 {
    panic("no file path specified")
}
filePath := os.Args[1]

file, err := os.Open(filePath)
if err != nil {
    panic(err)
}
defer file.Close()
```

我们会逐字节遍历文本，跟踪状态。幸运的是，在本文案例中，我们只需要引入两个状态：

- 前一个字节是 whitespace
- 前一个字节不是 whitespace

当从一个 whitespace 字符跳到一个 non-whitespace 字符时，单词计数加 1。这种方法可以直接从字节流读取信息，保持低内存消耗。

```go
const bufferSize = 16 * 1024
reader := bufio.NewReaderSize(file, bufferSize)

lineCount := 0
wordCount := 0
byteCount := 0

prevByteIsSpace := true
for {
    b, err := reader.ReadByte()
    if err != nil {
        if err == io.EOF {
            break
        } else {
            panic(err)
        }
    }

    byteCount++

    switch b {
    case '\n':
        lineCount++
        prevByteIsSpace = true
    case ' ', '\t', '\r', '\v', '\f':
        prevByteIsSpace = true
    default:
        if prevByteIsSpace {
            wordCount++
            prevByteIsSpace = false
        }
    }
}
```

为了展示结果，我们用原生的 println() 函数 — 在我的试验中，导入 fmt 包会导致运行时空间增加约 400 KB。

```go
println(lineCount, wordCount, byteCount, file.Name())
```

运行结果：

|          | input size | elapsed time | max memory |
| -------- | ---------- | ------------ | ---------- |
| wc       | 100 MB     | 0.58 s       | 2052 KB    |
| wc-naive | 100 MB     | 0.77 s       | 1416 KB    |
| wc       | 1 GB       | 5.56 s       | 2036 KB    |
| wc-naive | 1 GB       | 7.69 s       | 1416 KB    |

好消息是我们的第一次尝试在性能方面非常接近 C 语言。事实上，在内存使用方面，我们做得比 C 语言*更好*。

## 分割输入

虽然对 I/O 读取进行缓冲显著提升了性能，但是调用 ReadByte() 和在循环中检查 error 造成了一大笔不必要的开销。我们可以通过手动缓冲读请求来规避上述情况，而不再依赖 bufio.Reader。

为了实现手动缓冲，我们把输入分割成多个可以单独处理的缓冲块。幸运的是，我们只要知道前一个缓冲块（我们之前看到过）的最后一个字符是否是 whitespace 就可以处理当前的块。

我们写几个实用函数：

```go
type Chunk struct {
    PrevCharIsSpace bool
    Buffer          []byte
}

type Count struct {
    LineCount int
    WordCount int
}

func GetCount(chunk Chunk) Count {
    count := Count{}

    prevCharIsSpace := chunk.PrevCharIsSpace
    for _, b := range chunk.Buffer {
        switch b {
        case '\n':
            count.LineCount++
            prevCharIsSpace = true
        case ' ', '\t', '\r', '\v', '\f':
            prevCharIsSpace = true
        default:
            if prevCharIsSpace {
                prevCharIsSpace = false
                count.WordCount++
            }
        }
    }

    return count
}

func IsSpace(b byte) bool {
    return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\v' || b == '\f'
}
```

现在，我们可以把输入分割成多个块，然后传入 GetCount 函数。

```go
totalCount := Count{}
lastCharIsSpace := true

const bufferSize = 16 * 1024
buffer := make([]byte, bufferSize)

for {
    bytes, err := file.Read(buffer)
    if err != nil {
        if err == io.EOF {
            break
        } else {
            panic(err)
        }
    }

    count := GetCount(Chunk{lastCharIsSpace, buffer[:bytes]})
    lastCharIsSpace = IsSpace(buffer[bytes-1])

    totalCount.LineCount += count.LineCount
    totalCount.WordCount += count.WordCount
}
```

为了计数字节，我们可以进行一次系统调用来查询文件的大小：

```go
fileStat, err := file.Stat()
if err != nil {
    panic(err)
}
byteCount := fileStat.Size()
```

现在做完该做的了，来看一下表现如何：

|           | input size | elapsed time | max memory |
| --------- | ---------- | ------------ | ---------- |
| wc        | 100 MB     | 0.58 s       | 2052 KB    |
| wc-chunks | 100 MB     | 0.34 s       | 1404 KB    |
| wc        | 1 GB       | 5.56 s       | 2036 KB    |
| wc-chunks | 1 GB       | 3.31 s       | 1416 KB    |

看起来我们在两个统计维度上都超过了 `wc`，而且我们还没有开始并行化我们的程序。[`tokei`](https://github.com/XAMPPRocky/tokei) 统计结果显示这个程序一共只有 70 行代码！

## 并行化

诚然，并行化实现 `wc` 是大材小用了，但我们还是来看一下到底能达到什么程度。原文是并行地从输入文件中读的，尽管它缩短了运行时间，但作者同时也承认，并行读仅在特定几种存储的情况下对性能有提升，其他的情况下可能降低性能。

在我们的实现中，我们希望我们的代码在*所有设备*上运行都能有很好的性能，所以我们不像原文作者那样做。我们会创建两个 channel：chunks 和 counts 。每个 worker 都会从 chunks 读取和处理数据直到 channel 被 close，之后把结果写进 counts。

```go
func ChunkCounter(chunks <-chan Chunk, counts chan<- Count) {
    totalCount := Count{}
    for {
        chunk, ok := <-chunks
        if !ok {
            break
        }
        count := GetCount(chunk)
        totalCount.LineCount += count.LineCount
        totalCount.WordCount += count.WordCount
    }
    counts <- totalCount
}
```

我们在每个 CPU core 上起一个 worker：

```go
numWorkers := runtime.NumCPU()

chunks := make(chan Chunk)
counts := make(chan Count)

for i := 0; i < numWorkers; i++ {
    Go ChunkCounter(chunks, counts)
}
```

现在，我们在循环中从硬盘读取数据和给每个 worker 分配 job：

```go
const bufferSize = 16 * 1024
lastCharIsSpace := true

for {
    buffer := make([]byte, bufferSize)
    bytes, err := file.Read(buffer)
    if err != nil {
        if err == io.EOF {
            break
        } else {
            panic(err)
        }
    }
    chunks <- Chunk{lastCharIsSpace, buffer[:bytes]}
    lastCharIsSpace = IsSpace(buffer[bytes-1])
}
close(chunks)
```

这些完成后，我们可以很简单地把所有 worker 的计数相加。

```go
totalCount := Count{}
for i := 0; i < numWorkers; i++ {
    count := <-counts
    totalCount.LineCount += count.LineCount
    totalCount.WordCount += count.WordCount
}
close(counts)
```

我们运行起来然后看一下与之前结果的对比：

|            | input size | elapsed time | max memory |
| ---------- | ---------- | ------------ | ---------- |
| wc         | 100 MB     | 0.58 s       | 2052 KB    |
| wc-channel | 100 MB     | 0.27 s       | 6644 KB    |
| wc         | 1 GB       | 5.56 s       | 2036 KB    |
| wc-channel | 1 GB       | 2.22 s       | 6752 KB    |

我们实现的 `wc` 在速度方面有很大提升，但在内存使用方面与之前相比有些倒退。请特别留意我们的输入循环在每一次执行中是怎么样申请内存的！channel 是对共享内存的高度抽象，但在实际使用时，*不使用* channel 可能会大幅提升性能。

## 并行化升级版

在这部分，我们让每个 worker 都读取文件，并使用 sync.Mutex 来确保不会同时读取。我们可以创建一个 struct 来为我们处理这种情况：

```go
type FileReader struct {
    File            *os.File
    LastCharIsSpace bool
    mutex           sync.Mutex
}

func (fileReader *FileReader) ReadChunk(buffer []byte) (Chunk, error) {
    fileReader.mutex.Lock()
    defer fileReader.mutex.Unlock()

    bytes, err := fileReader.File.Read(buffer)
    if err != nil {
        return Chunk{}, err
    }

    chunk := Chunk{fileReader.LastCharIsSpace, buffer[:bytes]}
    fileReader.LastCharIsSpace = IsSpace(buffer[bytes-1])

    return chunk, nil
}
```

为了能直接读取文件，我们重写 worker 函数：

```go
func FileReaderCounter(fileReader *FileReader, counts chan Count) {
    const bufferSize = 16 * 1024
    buffer := make([]byte, bufferSize)

    totalCount := Count{}

    for {
        chunk, err := fileReader.ReadChunk(buffer)
        if err != nil {
            if err == io.EOF {
                break
            } else {
                panic(err)
            }
        }
        count := GetCount(chunk)
        totalCount.LineCount += count.LineCount
        totalCount.WordCount += count.WordCount
    }

    counts <- totalCount
}
```

像之前一样，我们还是在每个 CPU core 起一个 worker：

```go
fileReader := &FileReader{
    File:            file,
    LastCharIsSpace: true,
}
counts := make(chan Count)

for i := 0; i < numWorkers; i++ {
    Go FileReaderCounter(fileReader, counts)
}

totalCount := Count{}
for i := 0; i < numWorkers; i++ {
    count := <-counts
    totalCount.LineCount += count.LineCount
    totalCount.WordCount += count.WordCount
}
close(counts)
```

来看看表现如何：

|          | intput size | elapsed time | max memory |
| -------- | ----------- | ------------ | ---------- |
| wc       | 100 MB      | 0.58 s       | 2052 KB    |
| wc-mutex | 100 MB      | 0.12 s       | 1580 KB    |
| wc       | 1 GB        | 5.56 s       | 2036 KB    |
| wc-mutex | 1 GB        | 1.21 s       | 1576 KB    |

我们的并行化实现用更小的内存消耗比 `wc` 的运行速度快了 4.5 倍！这意义非凡，尤其是当你意识到 Go 是一种有垃圾回收机制的语言时。

## 总结

尽管本文结论并不意味着 Go > C，但我希望它能证明 Go 作为一种系统编程语言可以是 C 语言的可替代项。

如果你有任何建议、问题、意见，尽情 [给我发邮件](mailto:98ajeet@gmail.com)！

---

via: https://ajeetdsouza.github.io/blog/posts/beating-c-with-70-lines-of-go/

作者：[Ajeet D'Souza](https://ajeetdsouza.github.io/blog/)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
