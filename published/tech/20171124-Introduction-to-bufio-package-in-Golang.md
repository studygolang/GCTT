# Go 语言 bufio 包的介绍
[原文链接](https://medium.com/golangspec/introduction-to-bufio-package-in-golang-ad7d1877f762)

[bufio](https://golang.org/pkg/bufio/) 用来帮助处理 [I/O 缓存](https://www.quora.com/In-C-what-does-buffering-I-O-or-buffered-I-O-mean/answer/Robert-Love-1)。 我们将通过一些示例来熟悉其为我们提供的：Reader, Writer and Scanner 等一系列功能

## bufio.Writer
多次进行小量的写操作会影响程序性能。每一次写操作最终都会体现为系统层调用，频繁进行该操作将有可能对 CPU 造成伤害。而且很多硬件设备更适合处理块对齐的数据，例如硬盘。为了减少进行多次写操作所需的开支，golang 提供了 [bufio.Writer](https://golang.org/pkg/bufio/#Writer)。数据将不再直接写入目的地(实现了 [io.Writer](https://golang.org/pkg/io/#Writer) 接口)，而是先写入缓存，当缓存写满后再统一写入目的地：
```
producer --> buffer --> io.Writer
```
下面具体看一下在9次写入操作中(每次写入一个字符)具有4个字符空间的缓存是如何工作的：
```
producer        buffer       destination (io.Writer)
   a    ----->    a
   b    ----->    ab
   c    ----->    abc
   d    ----->    abcd
   e    ----->    e      ----->   abcd
   f    ----->    ef
   g    ----->    efg
   h    ----->    efgh
   i    ----->    i      ----->   abcdefgh
```
`----->` 箭头代表写入操作

[`bufio.Writer`](https://golang.org/pkg/bufio/#Writer) 底层使用 `[]byte` 进行缓存
```go
type Writer int
func (*Writer) Write(p []byte) (n int, err error) {
    fmt.Println(len(p))
    return len(p), nil
}
func main() {
    fmt.Println("Unbuffered I/O")
    w := new(Writer)
    w.Write([]byte{'a'})
    w.Write([]byte{'b'})
    w.Write([]byte{'c'})
    w.Write([]byte{'d'})
    fmt.Println("Buffered I/O")
    bw := bufio.NewWriterSize(w, 3)
    bw.Write([]byte{'a'})
    bw.Write([]byte{'b'})
    bw.Write([]byte{'c'})
    bw.Write([]byte{'d'})
    err := bw.Flush()
    if err != nil {
        panic(err)
    }
}
Unbuffered I/O
1
1
1
1
Buffered I/O
3
1
```
没有被缓存的 `I/O`：意味着每一次写操作都将直接写入目的地。我们进行4次写操作，每次写操作都映射为对 `Write` 的调用，调用时传入的参数为一个长度为1的 `byte` 切片。

使用了缓存的 `I/O`：我们使用三个字节长度的缓存来存储数据，当缓存满时进行一次 `flush` 操作(将缓存中的数据进行处理)。前三次写入写满了缓存。第四次写入时检测到缓存没有剩余空间，所以将缓存中的积累的数据写出。字母 `d` 被存储了，但在此之前 `Flush` 被调用以腾出空间。当缓存被写到末尾时，缓存中未被处理的数据需要被处理。`bufio.Writer` 仅在缓存充满或者显式调用 `Flush` 方法时处理(发送)数据。

> `bufio.Writer` 默认使用 4096 长度字节的缓存，可以使用 [`NewWriterSize`](https://golang.org/pkg/bufio/#NewWriterSize) 方法来设定该值

## 实现
实现十分简单：
```go
type Writer struct {
    err error
    buf []byte
    n   int
    wr  io.Writer
}
```
字段 `buf` 用来存储数据，当缓存满或者 `Flush` 被调用时，消费者(`wr`)可以从缓存中获取到数据。如果写入过程中发生了 I/O error，此 error 将会被赋给 `err` 字段， error 发生之后，writer 将停止操作(writer is no-op)：
```go
type Writer int
func (*Writer) Write(p []byte) (n int, err error) {
    fmt.Printf("Write: %q\n", p)
    return 0, errors.New("boom!")
}
func main() {
    w := new(Writer)
    bw := bufio.NewWriterSize(w, 3)
    bw.Write([]byte{'a'})
    bw.Write([]byte{'b'})
    bw.Write([]byte{'c'})
    bw.Write([]byte{'d'})
    err := bw.Flush()
    fmt.Println(err)
}
Write: "abc"
boom!
```
这里我们可以看到 `Flush` 没有第二次调用消费者的 `write` 方法。如果发生了 error， 使用了缓存的 writer 不会尝试再次执行写操作。

字段 `n` 标识缓存内部当前操作的位置。`Buffered` 方法返回 `n` 的值：
```go
type Writer int
func (*Writer) Write(p []byte) (n int, err error) {
    return len(p), nil
}
func main() {
    w := new(Writer)
    bw := bufio.NewWriterSize(w, 3)
    fmt.Println(bw.Buffered())
    bw.Write([]byte{'a'})
    fmt.Println(bw.Buffered())
    bw.Write([]byte{'b'})
    fmt.Println(bw.Buffered())
    bw.Write([]byte{'c'})
    fmt.Println(bw.Buffered())
    bw.Write([]byte{'d'})
    fmt.Println(bw.Buffered())
}
0
1
2
3
1
```
`n` 从 0 开始，当有数据被添加到缓存中时，该数据的长度值将会被加和到 `n`中(操作位置向后移动)。当`bw.Write([] byte{'d'})`被调用时，flush会被触发，`n` 会被重设为0。

## Large writes
```go
type Writer int
func (*Writer) Write(p []byte) (n int, err error) {
    fmt.Printf("%q\n", p)
    return len(p), nil
}
func main() {
    w := new(Writer)
    bw := bufio.NewWriterSize(w, 3)
    bw.Write([]byte("abcd"))
}
```
由于使用了 `bufio`，程序打印了 `"abcd"`。如果 `Writer` 检测到 `Write` 方法被调用时传入的数据长度大于缓存的长度(示例中是三个字节)。其将直接调用 writer(目的对象)的 `Write` 方法。当数据量足够大时，其会自动跳过内部缓存代理。

## 重置
缓存是 `bufio` 的核心部分。通过使用 `Reset` 方法，`Writer` 可以用于不同的目的对象。重复使用 `Writer` 缓存减少了内存的分配。而且减少了额外的垃圾回收工作：
```go
type Writer1 int
func (*Writer1) Write(p []byte) (n int, err error) {
    fmt.Printf("writer#1: %q\n", p)
    return len(p), nil
}
type Writer2 int
func (*Writer2) Write(p []byte) (n int, err error) {
    fmt.Printf("writer#2: %q\n", p)
    return len(p), nil
}
func main() {
    w1 := new(Writer1)
    bw := bufio.NewWriterSize(w1, 2)
    bw.Write([]byte("ab"))
    bw.Write([]byte("cd"))
    w2 := new(Writer2)
    bw.Reset(w2)
    bw.Write([]byte("ef"))
    bw.Flush()
}
writer#1: "ab"
writer#2: "ef"
```
这段代码中有一个 bug。在调用 `Reset` 方法之前，我们应该使用 `Flush` flush缓存。 由于 [`Reset`](https://github.com/golang/go/blob/7b8a7f8272fd1941a199af1adb334bd9996e8909/src/bufio/bufio.go#L559) 只是简单的丢弃未被处理的数据，所以已经被写入的数据 `cd` 丢失了：
```go
func (b *Writer) Reset(w io.Writer) {
    b.err = nil
    b.n = 0
    b.wr = w
}
```

## 缓存剩余空间
为了检测缓存中还剩余多少空间, 我们可以使用方法 `Available`：
```go
w := new(Writer)
bw := bufio.NewWriterSize(w, 2)
fmt.Println(bw.Available())
bw.Write([]byte{'a'})
fmt.Println(bw.Available())
bw.Write([]byte{'b'})
fmt.Println(bw.Available())
bw.Write([]byte{'c'})
fmt.Println(bw.Available())
2
1
0
1
```

## 写`{Byte,Rune,String}`的方法
为了方便, 我们有三个用来写普通类型的实用方法：
```go
w := new(Writer)
bw := bufio.NewWriterSize(w, 10)
fmt.Println(bw.Buffered())
bw.WriteByte('a')
fmt.Println(bw.Buffered())
bw.WriteRune('ł') // 'ł' occupies 2 bytes
fmt.Println(bw.Buffered())
bw.WriteString("aa")
fmt.Println(bw.Buffered())
0
1
3
5
```

## ReadFrom
io 包中定义了 [`io.ReaderFrom`](https://golang.org/pkg/io/#ReaderFrom) 接口。 该接口通常被 writer 实现，用于从指定的 reader 中读取所有数据(直到 EOF)并对读到的数据进行底层处理：
```go
type ReaderFrom interface {
        ReadFrom(r Reader) (n int64, err error)
}
```
>比如 [`io.Copy`](https://golang.org/pkg/io/#Copy) 使用了 `io.ReaderFrom` 接口

`bufio.Writer` 实现了此接口：因此我们可以通过调用 `ReadFrom` 方法来处理从 `io.Reader` 获取到的所有数据：
```go
type Writer int
func (*Writer) Write(p []byte) (n int, err error) {
    fmt.Printf("%q\n", p)
    return len(p), nil
}
func main() {
    s := strings.NewReader("onetwothree")
    w := new(Writer)
    bw := bufio.NewWriterSize(w, 3)
    bw.ReadFrom(s)
    err := bw.Flush()
    if err != nil {
        panic(err)
    }
}
"one"
"two"
"thr"
"ee"
```
>使用 `ReadFrom` 方法的同时，调用 `Flush` 方法也很重要

## bufio.Reader
通过它，我们可以从底层的 `io.Reader` 中更大批量的读取数据。这会使读取操作变少。如果数据读取时的块数量是固定合适的，底层媒体设备将会有更好的表现，也因此会提高程序的性能：
```
io.Reader --> buffer --> consumer
```
假设消费者想要从硬盘上读取10个字符(每次读取一个字符)。在底层实现上，这将会触发10次读取操作。如果硬盘按每个数据块四个字节来读取数据，那么 `bufio.Reader` 将会起到帮助作用。底层引擎将会缓存整个数据块，然后提供一个可以挨个读取字节的 API 给消费者：
```
abcd -----> abcd -----> a
            abcd -----> b
            abcd -----> c
            abcd -----> d
efgh -----> efgh -----> e
            efgh -----> f
            efgh -----> g
            efgh -----> h
ijkl -----> ijkl -----> i
            ijkl -----> j
```
`----->` 代表读取操作<br>
这个方法仅需要从硬盘读取三次，而不是10次。

## Peek
`Peek` 方法可以帮助我们查看缓存的前 n 个字节而不会真的『吃掉』它：
- 如果缓存不满，而且缓存中缓存的数据少于 `n` 个字节，其将会尝试从 `io.Reader` 中读取
- 如果请求的数据量大于缓存的容量，将会返回 `bufio.ErrBufferFull`
- 如果 `n` 大于流的大小，将会返回 EOF

让我们来看看它是如何工作的：
```go
s1 := strings.NewReader(strings.Repeat("a", 20))
r := bufio.NewReaderSize(s1, 16)
b, err := r.Peek(3)
if err != nil {
    fmt.Println(err)
}
fmt.Printf("%q\n", b)
b, err = r.Peek(17)
if err != nil {
    fmt.Println(err)
}
s2 := strings.NewReader("aaa")
r.Reset(s2)
b, err = r.Peek(10)
if err != nil {
    fmt.Println(err)
}
"aaa"
bufio: buffer full
EOF
```
>被 `bufio.Reader` 使用的最小的缓存容器是 16。

返回的切片和被 `bufio.Reader` 使用的内部缓存底层使用相同的数组。因此引擎底层在执行任何读取操作之后内部返回的切片将会变成无效的。这是由于其将有可能被其他的缓存数据覆盖：
```go
s1 := strings.NewReader(strings.Repeat("a", 16) + strings.Repeat("b", 16))
r := bufio.NewReaderSize(s1, 16)
b, _ := r.Peek(3)
fmt.Printf("%q\n", b)
r.Read(make([]byte, 16))
r.Read(make([]byte, 15))
fmt.Printf("%q\n", b)
"aaa"
"bbb"
```

## Reset
就像 `bufio.Writer` 那样，缓存也可以用相似的方式被复用。
```go
s1 := strings.NewReader("abcd")
r := bufio.NewReader(s1)
b := make([]byte, 3)
_, err := r.Read(b)
if err != nil {
    panic(err)
}
fmt.Printf("%q\n", b)
s2 := strings.NewReader("efgh")
r.Reset(s2)
_, err = r.Read(b)
if err != nil {
    panic(err)
}
fmt.Printf("%q\n", b)
"abc"
"efg"
```
通过使用 `Reset`，我们可以避免冗余的内存分配和不必要的垃圾回收工作。

## Discard
这个方法将会丢弃 `n` 个字节的，返回时也不会返回被丢弃的 `n` 个字节。如果 `bufio.Reader` 缓存了超过或者等于 `n` 个字节的数据。那么其将不必从 `io.Reader` 中读取任何数据。其只是简单的从缓存中略去前 `n` 个字节：
```go
type R struct{}
func (r *R) Read(p []byte) (n int, err error) {
    fmt.Println("Read")
    copy(p, "abcdefghijklmnop")
    return 16, nil
}
func main() {
    r := new(R)
    br := bufio.NewReaderSize(r, 16)
    buf := make([]byte, 4)
    br.Read(buf)
    fmt.Printf("%q\n", buf)
    br.Discard(4)
    br.Read(buf)
    fmt.Printf("%q\n", buf)
}
Read
"abcd"
"ijkl"
```
调用 `Discard` 方法将不会从 reader `r` 中读取数据。另一种情况，缓存中数据量小于 `n`，那么 `bufio.Reader` 将会读取需要数量的数据来确保被丢弃的数据量不会少于 `n`：
```go
type R struct{}
func (r *R) Read(p []byte) (n int, err error) {
    fmt.Println("Read")
    copy(p, "abcdefghijklmnop")
    return 16, nil
}
func main() {
    r := new(R)
    br := bufio.NewReaderSize(r, 16)
    buf := make([]byte, 4)
    br.Read(buf)
    fmt.Printf("%q\n", buf)
    br.Discard(13)
    fmt.Println("Discard")
    br.Read(buf)
    fmt.Printf("%q\n", buf)
}
Read
"abcd"
Read
Discard
"bcde"
```
由于调用了 `Discard` 方法，所以读取方法被调用了两次。

## Read
`Read` 方法是 `bufio.Reader` 的核心。它和 [`io.Reader`](https://golang.org/pkg/io/#Reader) 的唯一方法具有相同的签名。因此 `bufio.Reader` 实现了这个普遍存在的接口：
```go
type Reader interface {
        Read(p []byte) (n int, err error)
}
```

`bufio.Reader` 的 `Read` 方法从底层的 `io.Reader` 中一次读取最大的数量:

1. 如果内部缓存具有至少一个字节的数据，那么无论传入的切片的大小(`len(p)`)是多少，`Read` 方法都将仅仅从内部缓存中获取数据，不会从底层的 reader 中读取任何数据:
```go
func (r *R) Read(p []byte) (n int, err error) {
    fmt.Println("Read")
    copy(p, "abcd")
    return 4, nil
}
func main() {
    r := new(R)
    br := bufio.NewReader(r)
    buf := make([]byte, 2)
    n, err := br.Read(buf)
    if err != nil {
        panic(err)
    }
    buf = make([]byte, 4)
    n, err = br.Read(buf)
    if err != nil {
        panic(err)
    }
    fmt.Printf("read = %q, n = %d\n", buf[:n], n)
}
Read
read = "cd", n = 2
```
我们的 `io.Reader` 实例无线返回「abcd」(不会返回 `io.EOF`)。 第二次调用 `Read`并传入长度为4的切片，但是内部缓存在第一次从 `io.Reader` 中读取数据之后已经具有数据「cd」，所以 `bufio.Reader` 返回缓存中的数据数据，而不和底层 reader 进行通信。

2. 如果内部缓存是空的，那么将会执行一次从底层 io.Reader 的读取操作。 从前面的例子中我们可以清晰的看到如果我们开启了一个空的缓存，然后调用:
```go
n, err := br.Read(buf)
```
将会触发读取操作来填充缓存。

3. 如果内部缓存是空的，但是传入的切片长度大于缓存长度，那么 `bufio.Reader` 将会跳过缓存，直接读取传入切片长度的数据到切片中:
```go
type R struct{}
func (r *R) Read(p []byte) (n int, err error) {
    fmt.Println("Read")
    copy(p, strings.Repeat("a", len(p)))
    return len(p), nil
}
func main() {
    r := new(R)
    br := bufio.NewReaderSize(r, 16)
    buf := make([]byte, 17)
    n, err := br.Read(buf)
    if err != nil {
        panic(err)
    }
    fmt.Printf("read = %q, n = %d\n", buf[:n], n)
    fmt.Printf("buffered = %d\n", br.Buffered())
}
Read
read = "aaaaaaaaaaaaaaaaa", n = 17
buffered = 0
```
从 `bufio.Reader` 读取之后，内部缓存中没有任何数据(`buffered = 0`)

## {Read, Unread}Byte
这些方法都实现了从缓存中读取单个字节或者将最后一个读取的字节返回到缓存:
```go
r := strings.NewReader("abcd")
br := bufio.NewReader(r)
byte, err := br.ReadByte()
if err != nil {
    panic(err)
}
fmt.Printf("%q\n", byte)
fmt.Printf("buffered = %d\n", br.Buffered())
err = br.UnreadByte()
if err != nil {
    panic(err)
}
fmt.Printf("buffered = %d\n", br.Buffered())
byte, err = br.ReadByte()
if err != nil {
    panic(err)
}
fmt.Printf("%q\n", byte)
fmt.Printf("buffered = %d\n", br.Buffered())
'a'
buffered = 3
buffered = 4
'a'
buffered = 3
```

## {Read, Unread}Rune
这两个方法的功能和前面方法的功能差不多, 但是用来处理 Unicode 字符(UTF-8 encoded)。

## ReadSlice
函数返回在第一次出现传入字节前的字节:
```go
func (b *Reader) ReadSlice(delim byte) (line []byte, err error)
```
示例:
```go
s := strings.NewReader("abcdef|ghij")
r := bufio.NewReader(s)
token, err := r.ReadSlice('|')
if err != nil {
    panic(err)
}
fmt.Printf("Token: %q\n", token)
Token: "abcdef|"
```
>重要：返回的切面指向内部缓冲区, 因此它可能在下一次读取操作期间被覆盖

如果找不到分隔符，而且已经读到末尾(EOF)，将会返回 `io.EOF` error。 让我们将上面程序中的一行修改为如下代码:
```go
s := strings.NewReader("abcdefghij")
```
如果数据以 `panic: EOF` 结尾。 当分隔符找不到而且没有更多的数据可以放入缓冲区时函数将返回 [`io.ErrBufferFull`](https://golang.org/pkg/bufio/#pkg-variables):
```go
s := strings.NewReader(strings.Repeat("a", 16) + "|")
r := bufio.NewReaderSize(s, 16)
token, err := r.ReadSlice('|')
if err != nil {
    panic(err)
}
fmt.Printf("Token: %q\n", token)
```
这一小段代码会出现错误:`panic: bufio: buffer full`。

## ReadBytes
```go
func (b *Reader) ReadBytes(delim byte) ([]byte, error)
```
返回出现第一次分隔符前的所有数据组成的字节切片。 它和 `ReadSlice` 具有相同的签名，但是 `ReadSlice` 是一个低级别的函数，`ReadBytes` 的实现使用了 `ReadSlice`。 那么两者之间有什么不同呢? 在分隔符找不到的情况下，`ReadBytes` 可以多次调用 `ReadSlice`，而且可以累积返回的数据。 这意味着 `ReadBytes` 将不再受到 缓存大小的限制:
```go
s := strings.NewReader(strings.Repeat("a", 40) + "|")
r := bufio.NewReaderSize(s, 16)
token, err := r.ReadBytes('|')
if err != nil {
    panic(err)
}
fmt.Printf("Token: %q\n", token)
Token: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa|"
```
另外该函数返回一个新的字节切片，所以没有数据会被将来的读取操作覆盖的风险。

## ReadString
它是我们上面讨论的 `ReadBytes` 的简单封装:
```go
func (b *Reader) ReadString(delim byte) (string, error) {
    bytes, err := b.ReadBytes(delim)
    return string(bytes), err
}
```

## ReadLine
```go
ReadLine() (line []byte, isPrefix bool, err error)
```
内部使用 `ReadSlice` (`ReadSlice('\n')`)实现，同时从返回的切片中移除掉换行符(`\n` 或者 `\r\n`)。 此方法的签名不同于 `ReadBytes` 或者 `ReadSlice`，因为它包含 `isPrefix` 标志。 由于内部缓存无法存储更多的数据，当找不到分隔符时该标志为 true:
```go
s := strings.NewReader(strings.Repeat("a", 20) + "\n" + "b")
r := bufio.NewReaderSize(s, 16)
token, isPrefix, err := r.ReadLine()
if err != nil {
    panic(err)
}
fmt.Printf("Token: %q, prefix: %t\n", token, isPrefix)
token, isPrefix, err = r.ReadLine()
if err != nil {
    panic(err)
}
fmt.Printf("Token: %q, prefix: %t\n", token, isPrefix)
token, isPrefix, err = r.ReadLine()
if err != nil {
    panic(err)
}
fmt.Printf("Token: %q, prefix: %t\n", token, isPrefix)
token, isPrefix, err = r.ReadLine()
if err != nil {
    panic(err)
}
Token: "aaaaaaaaaaaaaaaa", prefix: true
Token: "aaaa", prefix: false
Token: "b", prefix: false
panic: EOF
```
如果最后一次返回的切片以换行符结尾，此方法将不会给出任何信息:
```go
s := strings.NewReader("abc")
r := bufio.NewReaderSize(s, 16)
token, isPrefix, err := r.ReadLine()
if err != nil {
    panic(err)
}
fmt.Printf("Token: %q, prefix: %t\n", token, isPrefix)
s = strings.NewReader("abc\n")
r.Reset(s)
token, isPrefix, err = r.ReadLine()
if err != nil {
    panic(err)
}
fmt.Printf("Token: %q, prefix: %t\n", token, isPrefix)
Token: "abc", prefix: false
Token: "abc", prefix: false
```

## WriteTo
`bufio.Reader` 实现了 `io.WriterTo` 接口:
```go
type WriterTo interface {
        WriteTo(w Writer) (n int64, err error)
}
```
此方法允许我们传入一个实现了 `io.Writer` 的消费者。 从生产者读取的所有数据都将会被送到消费者。 下面通过练习来看看它是如何工作的:
```go
type R struct {
    n int
}
func (r *R) Read(p []byte) (n int, err error) {
    fmt.Printf("Read #%d\n", r.n)
    if r.n >= 10 {
         return 0, io.EOF
    }
    copy(p, "abcd")
    r.n += 1
    return 4, nil
}
func main() {
    r := bufio.NewReaderSize(new(R), 16)
    n, err := r.WriteTo(ioutil.Discard)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Written bytes: %d\n", n)
}
Read #0
Read #1
Read #2
Read #3
Read #4
Read #5
Read #6
Read #7
Read #8
Read #9
Read #10
Written bytes: 40
```

## bufio.Scanner
[go语言中对bufio.Scanner的深层分析](https://medium.com/golangspec/in-depth-introduction-to-bufio-scanner-in-golang-55483bb689b4)

## ReadBytes('\n'), ReadString('\n'), ReadLine 还是 Scanner?
就像前面说的那样，`ReadString('\n')` 只是对于 `ReadBytes(`\n`)` 的简单封装。 所以让我们来讨论一下另外三者之间的不同之处吧。

1. ReadBytes 不会自动处理 `\r\n` 序列:
```go
s := strings.NewReader("a\r\nb")
r := bufio.NewReader(s)
for {
    token, _, err := r.ReadLine()
    if len(token) > 0 {
        fmt.Printf("Token (ReadLine): %q\n", token)
    }
    if err != nil {
        break
    }
}
s.Seek(0, io.SeekStart)
r.Reset(s)
for {
    token, err := r.ReadBytes('\n')
    fmt.Printf("Token (ReadBytes): %q\n", token)
    if err != nil {
        break
    }
}
s.Seek(0, io.SeekStart)
scanner := bufio.NewScanner(s)
for scanner.Scan() {
    fmt.Printf("Token (Scanner): %q\n", scanner.Text())
}
Token (ReadLine): "a"
Token (ReadLine): "b"
Token (ReadBytes): "a\r\n"
Token (ReadBytes): "b"
Token (Scanner): "a"
Token (Scanner): "b"
```
*ReadBytes* 会将分隔符一起返回，所以需要额外的一些工作来重新处理数据(除非返回分隔符是有用的)。

2. *ReadLine* 不会处理超出内部缓存的行:
```go
s := strings.NewReader(strings.Repeat("a", 20) + "\n")
r := bufio.NewReaderSize(s, 16)
token, _, _ := r.ReadLine()
fmt.Printf("Token (ReadLine): \t%q\n", token)
s.Seek(0, io.SeekStart)
r.Reset(s)
token, _ = r.ReadBytes('\n')
fmt.Printf("Token (ReadBytes): \t%q\n", token)
s.Seek(0, io.SeekStart)
scanner := bufio.NewScanner(s)
scanner.Scan()
fmt.Printf("Token (Scanner): \t%q\n", scanner.Text())
Token (ReadLine): 	"aaaaaaaaaaaaaaaa"
Token (ReadBytes): 	"aaaaaaaaaaaaaaaaaaaa\n"
Token (Scanner): 	"aaaaaaaaaaaaaaaaaaaa"
```
为了取回流中剩余的数据，*ReadLine* 需要被调用两次。 被 Scanner 处理的最大 token 长度为 64*1024。 如果传入更长的 token，scanner 将无法工作。 当 *ReadLine* 被多次调用时可以处理任何长度的 token。 由于函数返回是否在缓存数据中找到分隔符的标志，但是这需要调用者进行处理。 *ReadBytes* 则没有任何限制:
```go
s := strings.NewReader(strings.Repeat("a", 64*1024) + "\n")
r := bufio.NewReader(s)
token, _, err := r.ReadLine()
fmt.Printf("Token (ReadLine): %d\n", len(token))
fmt.Printf("Error (ReadLine): %v\n", err)
s.Seek(0, io.SeekStart)
r.Reset(s)
token, err = r.ReadBytes('\n')
fmt.Printf("Token (ReadBytes): %d\n", len(token))
fmt.Printf("Error (ReadBytes): %v\n", err)
s.Seek(0, io.SeekStart)
scanner := bufio.NewScanner(s)
scanner.Scan()
fmt.Printf("Token (Scanner): %d\n", len(scanner.Text()))
fmt.Printf("Error (Scanner): %v\n", scanner.Err())
Token (ReadLine): 4096
Error (ReadLine): <nil>
Token (ReadBytes): 65537
Error (ReadBytes): <nil>
Token (Scanner): 0
Error (Scanner): bufio.Scanner: token too long
```
3. 就像上面那样，*Scanner* 具有非常简单的 API，对于普通的例子，它还提供了友好的抽象概念。

## bufio.ReadWriter
Go 的结构体中可以使用一种叫做内嵌的类型。 和常规的具有类型和名字的字段不同，我们可以仅仅使用类型(匿名字段)。 内嵌类型的方法或者字段如果不和其他的冲突的话，则可以使用一个简短的选择器来引用:
```go
type T1 struct {
    t1 string
}
func (t *T1) f1() {
    fmt.Println("T1.f1")
}
type T2 struct {
    t2 string
}
func (t *T2) f2() {
    fmt.Println("T1.f2")
}
type U struct {
    *T1
    *T2
}
func main() {
    u := U{T1: &T1{"foo"}, T2: &T2{"bar"}}
    u.f1()
    u.f2()
    fmt.Println(u.t1)
    fmt.Println(u.t2)
}
T1.f1
T1.f2
foo
bar
```
我们可以简单的使用 `u.t1` 来代替 `u.T1.t1`。 包 `bufio` 使用内嵌的方式来定义 *ReadWriter*。 它由 *Reader* 和 *Writer* 构成:
```go
type ReadWriter struct {
  	*Reader
  	*Writer
  }
```
让我们来看看它是如何使用的:
```go
s := strings.NewReader("abcd")
br := bufio.NewReader(s)
w := new(bytes.Buffer)
bw := bufio.NewWriter(w)
rw := bufio.NewReadWriter(br, bw)
buf := make([]byte, 2)
_, err := rw.Read(buf)
if err != nil {
    panic(err)
}
fmt.Printf("%q\n", buf)
buf = []byte("efgh")
_, err = rw.Write(buf)
if err != nil {
    panic(err)
}
err = rw.Flush()
if err != nil {
   panic(err)
}
fmt.Println(w.String())
"ab"
efgh
```
由于 reader 和 writer 都具有方法 *Buffered*，所以若想获取缓存数据的量，`rw.Buffered()` 将无法工作，编译器会报错：`ambiguous selector rw.Buffered`。 但是类似 `rw.Reader.Buffered()` 的方式是可以的。

## bufio + standard library
*bufio* 包被广泛使用在 I/O出现的标准库中，例如:
* archive/zip
* compress/*
* encoding/*
* image/*
* 类似于 net/http 的TCP连接包装。 它还结合一些类似于 sync.Pool 的缓存框架来减少垃圾回收的压力

---

via: https://medium.com/golangspec/introduction-to-bufio-package-in-golang-ad7d1877f762

作者：[Michał Łowicki](https://medium.com/@mlowicki?source=post_header_lockup)
译者：[jliu666](https://github.com/jliu666)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
