已发布：https://studygolang.com/articles/12638

# Go 实现对 HTTP 对象的查找

想象一下，在 `HTTP` 服务器上有一个巨大的 `ZIP` 文件，你想知道里面的内容。你不知道压缩包内是否有你需要的东西，而且你不想下载整个文件。是否可以像执行  `unzip -l https://example.com/giant.zip` 的操作来查看压缩包的内容呢？

这并不是一个为了用 `Go` 展示某些知识的理论问题。实际上，我也不想写一篇文章，除了我想通过那些压缩文件了解如何从 [美国专利和商标局（USPTO）](https://bulkdata.uspto.gov/data/patent/officialgazette/2017/) 下载大量专利。或者，我认为，能够从这些 `tar` 文件中获取 [1790 年发布的一些专利图像](https://bulkdata.uspto.gov/data/patent/grant/multipagepdf/1790_1999/) 有多酷？

去看看。那里有数百个巨大的 `ZIP` 和 `tarfiles` 值得探索！

在 `ZIP` 文件中最后的位置，有一个目录。因此在本地磁盘上，`“unzip -l”` 就像“寻求最终结果，找到 `TOC`，解析并打印它”一样简单。事实上，我们可以知道 `Go` 是如何处理的，因为在 [`zip.NewReader` 函数](https://godoc.org/archive/zip#NewReader) 需要传入一个文件路径。至于 `TAR` 文件，它们被设计用于磁带流式传输和内存稀少的时候，因此它们的目录在文件本身之间交错排列。

但我们不在本地，要从 `URL` 中读取内容对我们来说很有挑战。该怎么办？从哪里开始？

我们有几件事需要考虑，然后我们可以规划接下来的方向。寻找和读取 `HTTP` 文件也就是要找到和读取 `Range` 标头。那么，`USPTO` 服务器是否支持 `Range` 头呢？这很容易检查，使用 `curl` 和 `HTTP HEAD` 请求：

```shell
$ curl -I https://bulkdata.uspto.gov/data/patent/officialgazette/2017/e-OG20170103_1434-1.zip
HTTP/1.1 200 OK
Date: Mon, 11 Dec 2017 21:10:26 GMT
Server: Apache
Last-Modified: Tue, 03 Jan 2017 11:58:45 GMT
ETag: "afb8ac8-5452f63e0a82f"
Accept-Ranges: bytes
Content-Length: 184257224
X-Frame-Options: DENY
Content-Type: application/zip
```

请注意那里的 `“Accept-Ranges”` 标头，它表示我们可以向它发送字节范围。`Range` 头允许您像随机访问读取操作系统的一样操作 `HTTP`。（例如 [io.ReaderAt](https://godoc.org/io#ReaderAt) 接口）

因此理论上可以选择从 `Web` 服务器下载其中包含元数据（目录）的文件部分来决定下载哪些字节。

现在我们需要写一个处理 `ZIP` 文件格式的方法，它可以让我们使用具有 `Range` 头部的 `HTTP` 的 `GET` 请求，只读取元数据的方式，实现替换“读取下一个目录头文件”的某个部分。这就是 `Go` 的 [`archive/zip`](https://golang.org/pkg/archive/zip) 和 [`archive/tar`](https://godoc.org/archive/tar) 包的实现！

正如我们前面所说，[zip.NewReader](https://godoc.org/archive/zip#NewReader) 正在琢磨什么位置开始查找。然而，当我们看看 `TAR` 时，我们发现了一个问题。`tar.NewReader` 方法需要一个 `io.Reader`。`io.Reader` 的问题在于，它不会让我们随机访问资源，就像`io.ReaderAt` 一样。它是这样实现的，因为它使 `tar` 包更具适应性。特别是，您可以将 `Go tar` 包直接挂接到 `compress/gzip` 包并读取 `tar.gz` 文件 - 只要您按顺序读取它们，而不是像我们希望的那样跳过。

那么该怎么办？使用源码。环顾四周，找找[下一个方法](https://github.com/golang/go/blob/c007ce824d9a4fccb148f9204e04c23ed2984b71/src/archive/tar/reader.go#L88)。这就是我们期望它能够找到下一个元数据的地方。在几行代码内，对于 [`skipUnread`](https://github.com/golang/go/blob/c007ce824d9a4fccb148f9204e04c23ed2984b71/src/archive/tar/reader.go#L407) 函数， 我们发现一个有趣的调用。在那里，我们发现一些非常有趣的东西：

```go
// skipUnread skips any unread bytes in the existing file entry, as well as any alignment padding.
func (tr *Reader) skipUnread() {
  nr := tr.numBytes() + tr.pad // number of bytes to skip
  tr.curr, tr.pad = nil, 0
  if sr, ok := tr.r.(io.Seeker); ok {
    if _, err := sr.Seek(nr, os.SEEK_CUR); err == nil {
      return
    }
  }
  _, tr.err = io.CopyN(ioutil.Discard, tr.r, nr)
}

// Note: This is from Go 1.4, which had a simpler skipUnread than go 1.9 does.
```

这里表示：”如果 `io.Reader` 实际上也能够搜索，那么我们不是直接读取和丢弃，而是直接找到正确的地方。“找到了！我们只需要将 `tar` 文件传给 `io.Reader`。`NewReader` 也满足 [`io.Seeker`](https://golang.org/pkg/io/#Seeker)的功能（因此，它是一个[`io.ReadSeeker`](https://golang.org/pkg/io/#ReadSeeker)）。

所以，现在请查看包 [`github.com/jeffallen/seekinghttp`](https://godoc.org/github.com/jeffallen/seekinghttp)，就像它的名字所暗示的那样，它是一个用于在 `HTTP` 对象（[`Github` 上的源代码](https://github.com/jeffallen/seekinghttp) 中查找的软件包。

这个软件包不仅[实现](https://github.com/jeffallen/seekinghttp/blob/master/seekinghttp.go#L26)了 `io.ReadSeeker`，还实现了 `io.ReaderAt`。

为什么？因为，正如我上面提到的，读取 `ZIP` 文件需要一个 `io.ReaderAt`。它还需要传递给它的文件的长度，以便它可以查看目录文件的末尾。`HTTP HEAD` 方法可以很好地获取 `HTTP` 对象的 `Content-Length`，而不需要下载整个文件。

用于远程获取 `tar` 和 `zip` 文件目录的命令行工具位于 `remote-archive-ls` 中。打开 `“-debug”` 选项用来查看日志。**将 `Go` 的标准库作为 `TAR` 或 `ZIP` 阅读器“回调”到我们的代码中，并在这里请求几个字节，这里有几个字节是很有趣的。**

在我第一次运行这个程序后不久，我发现了一个严重的缺陷。这是一个示例运行结果：

``` shell
$ ./remote-archive-ls -debug 'https://bulkdata.uspto.gov/data/patent/grant/multipagepdf/1790_1999/grant_pdf_17900731_18641101.tar'
2017/12/12 00:07:38 got read len 512
2017/12/12 00:07:38 ReadAt len 512 off 0
2017/12/12 00:07:38 Start HTTP GET with Range: bytes=0-511
2017/12/12 00:07:39 HTTP ok.
File: 00000001-X009741H/
2017/12/12 00:07:39 got read len 512
2017/12/12 00:07:39 ReadAt len 512 off 512
2017/12/12 00:07:39 Start HTTP GET with Range: bytes=512-1023
2017/12/12 00:07:39 HTTP ok.
File: 00000001-X009741H/00/
2017/12/12 00:07:39 got read len 512
2017/12/12 00:07:39 ReadAt len 512 off 1024
2017/12/12 00:07:39 Start HTTP GET with Range: bytes=1024-1535
2017/12/12 00:07:39 HTTP ok.
File: 00000001-X009741H/00/000/
2017/12/12 00:07:39 got read len 512
2017/12/12 00:07:39 ReadAt len 512 off 1536
2017/12/12 00:07:39 Start HTTP GET with Range: bytes=1536-2047
2017/12/12 00:07:39 HTTP ok.
File: 00000001-X009741H/00/000/001/
2017/12/12 00:07:39 got read len 512
2017/12/12 00:07:39 ReadAt len 512 off 2048
2017/12/12 00:07:39 Start HTTP GET with Range: bytes=2048-2559
2017/12/12 00:07:39 HTTP ok.
File: 00000001-X009741H/00/000/001/us-patent-image.xml
2017/12/12 00:07:39 got seek 0 1
2017/12/12 00:07:39 got seek 982 1
2017/12/12 00:07:39 got read len 42
2017/12/12 00:07:39 ReadAt len 42 off 3542
2017/12/12 00:07:39 Start HTTP GET with Range: bytes=3542-3583
2017/12/12 00:07:39 HTTP ok.
2017/12/12 00:07:39 got read len 512
2017/12/12 00:07:39 ReadAt len 512 off 3584
2017/12/12 00:07:39 Start HTTP GET with Range: bytes=3584-4095
2017/12/12 00:07:39 HTTP ok.
File: 00000001-X009741H/00/000/001/00000001.pdf
2017/12/12 00:07:39 got seek 0 1
2017/12/12 00:07:39 got seek 320840 1
2017/12/12 00:07:39 got read len 184
2017/12/12 00:07:39 ReadAt len 184 off 324936
...etc...
```

你能看到问题吗？这是很多 `HTTP` 事务！ `TAR reader` 正在一次一点点地完成 `TAR` 流，发出一小串 `bit`。所有这些短的 `HTTP` 事务在服务器上都很难实现，并且对于吞吐量来说很糟糕，因为每个 `HTTP` 事务都需要多次往返服务器。

当然，解决方案是缓存。**读取TAR读取器要求的前 512 个字节，而不是读取其中的 10 倍，以便接下来的几个读取将直接从缓存中获取。**如果读取超出了缓存的范围，我们假设其他读取也将进入该区域，并删除整个当前缓存，以便用当前偏移量的 10 倍填充它。

`TAR` 阅读器发送**大量小读数**的事实指出了有关缓冲的一些非常重要的事情。将 [`os.Open`](https://godoc.org/os#Open) 的结果直接发送给 `tar`。`NewReader` 不是很聪明，尤其是如果你打算跳过文件寻找元数据。尽管 `* os.File` 实现了 `io.ReadSeeker`，我们现在知道 `TAR` 将会向内核发出大量的**小系统调用**。该解决方案与上面的解决方案非常相似，可能是使用 [`bufio`](https://godoc.org/bufio) 包来缓冲 `* os.File`，以便 `TAR` 发出的小数据将从 `RAM` 中取出，而不是转到操作系统。但请注意：它真的是解决方案吗？`bufio.Reader` 是否真的实现了 `io`？`ReadSeeker` 和 `io.ReadAt` 就像我们需要的一样？ **（破坏者：它没有;也许你们有读者想告诉我们如何使用下一个的替代品 `bufio` 加速 `Go` 的 `tar`？**

我希望你喜欢通过标准库和 `HTTP`，看看如何与标准库一起工作，以帮助它实现更多的功能，以便它可以帮助你完成你的工作这个小小的旅程。当你实现 `io.Reader` 和朋友时，你有机会走到你所调用的库的幕后，并从他们的作者从未期望的地方给他们提供数据！

---
via：https://blog.gopheracademy.com/advent-2017/seekable-http/

作者：[Jeff R. Allen](https://github.com/jeffallen)
译者：[yuhanle](https://github.com/yuhanle)
校对：[Unknwon](https://github.com/Unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
