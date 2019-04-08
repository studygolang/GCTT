首发于：https://studygolang.com/articles/19397

# Exceptions 和 Return

前几天[Thiago Pontes](https://twitter.com/thiagopnts) 分享了一篇关于异常被认为是反模式的博客给他的几个朋友。我对异常有一个不同的观点。我想如果写一个关于 exceptions 的博客会非常的有趣。我认为异常是一个非常好的功能，如果缺少异常可能会引起更大的错误。

这篇博客引用了我朋友分享的帖子：[Python exceptions considered an anti-pattern.](https://sobolevn.me/2019/02/python-exceptions-considered-an-antipattern)

## 没有异常的程序

如果你曾经用过 C 语言，你就记得 -1 和 NULL 作为返回值意味着错误，或者在这些情况下你需要记得去检查全局的错误号从而查出是否哪里出了问题。

如果一门语言不支持异常的话，那么你调用了一个方法，调用者就需要检查是否执行正确并且处理所有的错误。

例如，`malloc()` 这个函数，如果不能分配空间就返回 NULL，那你就必须检查返回值：

```c
int *p;
p = malloc(sizeof(int) * 100);
if (p == NULL) {
    fprintf(stderr, "ERR: Cant allocate memory!");
    exit(1);
}
```

或者进一步演变的[例子来自于 libcurl 检查 url 是否能被访问](https://curl.haxx.se/libcurl/c/simple.html)，返回 CURLE_OK 表示没有错误。

```c
#include <stdio.h>
#include <curl/curl.h>

int main(void)
{
  CURL *curl;
  CURLcode res;

  curl = curl_easy_init();
  if(curl) {
    curl_easy_setopt(curl, CURLOPT_URL, "https://example.com");
   /* example.com is redirected, so we tell libcurl to follow redirection */
    curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1L);

    /* Perform the request, res will get the return code */
    res = curl_easy_perform(curl);
    /* Check for errors */
    if(res != CURLE_OK)
      fprintf(stderr, "curl_easy_perform() failed: %s\n",
              curl_easy_strerror(res));

    /* always cleanup */
    curl_easy_cleanup(curl);
  }
  return 0;
}
```

我之所以用 C 语言作为例子，是因为之前曾经用 C 语言检查错误。但是这个可以应用到其他不支持异常的语言上，例如 golang。

## Golang 和 `err`

Go 是没有异常的，但是当在写一个方法时，通常的处理是返回一个结果和一个 error 的值。就想 http.Get 的做法一样：

```
// func Get(url string) (resp *Response, err error)

resp, err := http.Get("http://example.com/")
```

如果调用 `Get` 方法有任何的错误，那么变量 `err` 将捕获错误信息 , 如果没有错误那么它就是 `nil`。对于每个人在 Go 和 kudos 里面这个是一个经典的写法。如果缺少异常，那么你必须了解隐含的错误信息。

让我们看看创建一个访问 URL 和读取返回数据头的方法：

```
Content-Type:
```

```go
func GetContentType(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    return resp.Header["Content-Type"][0], nil
}
```

上面的方法可以这样使用：

```go
func main() {
    contentType, err := GetContentType("http://example.com")
    if err != nil {
        fmt.Println("Found unexpected error", err)
    } else {
        fmt.Printf("Content-Type: %s\n", contentType);
    }
}
```

如果 `err` 是 `nil`, 那么一切都是好的，对吗？错，这样是不对的。如果 `Content-Type` 没有找到，那么这句代码 `resp.Header["Content-Type"][0]` 将会因为 index 超出范围而发生致命的错误（程序将被打断）。前面的错误判断是不够的，因为它不能使程序恢复，程序将崩溃。错误处理并没有完全覆盖所有的情况。 程序照样会出错，所以自以为的 `err` 检查是一种误导。

## 显示的异常处理和连续的调用

当错误作为返回值的一部分返回，当然每个方法都应该有一个错误检查，如果知道错误的意义就要对错误进行处理，或者返回给上一级的调用者（例如像上面的 `GetContentType` 返回 `err` 给调用者）。

如果你使用 lib A，那么你就要信任 lib 的作者处理了每个可能的错误。并且你也要相信 lib A 依赖的其他库也都是正确的。最后你相信所有调用的 lib 都是正确的。所以你相信所有的做法都是正确的。

如果任何一个 lib 遗漏一个检查，那么你的程序将会出现未知的结果，检查所有的异常这是一个程序应该做的。

我曾经听过一个论点，它认为如果使用书写良好的 lib 那么这个责任并不是一个大问题（但是你怎么知道这个 lib 是良好的。这仍然是一个问题。但是书写良好的 lib 将其最小化。）

我喜欢 Go 语言的处理方式，它以显示的方式检查错误，当然他们不可能像我这样在 `GetContentType` 里面这样访问下标（我承认这是一个 bug，并且 bug 发生了）。你不能假设所有的分支都实现了并且都检查了，因为如果你使用了第三方的 lib，那么你就没有办法控制了。第三方库可能某一天会出现问题，那么你的程序就会崩溃（这个和没有捕获异常是一个问题。）

## 异常捕获

有些语言，例如 java，支持异常捕获。如果一个方法可以产生异常，那么所有的调用者必须显示的声明他们可以抛出此异常。例如：

```
public void ioOperation(boolean isResourceAvailable) throws IOException {
  if (!isResourceAvailable) {
    throw new IOException();
  }
}
```

Java 的编译器需要保证所有调用了 `ioOperation` 的方法有 `try/catch IOException` 或者方法声明有显示的声明抛出 IOException。这是一个合适的检查的方法。我曾经写过一点 Jave 解码的代码，显示的声明所有的异常是非常烦人的。编译器起了很大的作用，当你的程序中有很多的静态类型和异常检测，它可以使你写代码更容易。

Python 没有检查异常的功能，这个可以看成一个缺陷，因为在程序运行前你无法发现方法中可能产生的异常。在反模式的博客中讨论了关于通过 return 传递所有的值在一个 `Result` 对象里面，在这种情况下，你需要将异常的类型作为结果的一部分，并且通过 `mypy` 对程序进行静态的分析会更安全。如果使用静态类型分析，我认为这个是一个非常有意思的观点。java 一族已经这样做了十多年了，并且运行的很好。我在 `mypy` 和 `typing` 中有两个问题，但是作者貌似并不关心。

## 安全

你的程序在每个阶段都是安全的吗？我不这么认为，即使你的程序的所有地方都使用了 `try/catch`。也许在调用过程中一个方法可以使用 `try/catch` 拦截所有的错误，但是我仍然认为绝对的安全是不可能的。但是我喜欢一些语言的扩展的习俗，例如 Go。如果你的编程语言支持异常，那么请注意对异常的使用：是应该在你的方法里捕获它还是应该将它抛出给上层的调用者。

## 不清晰的程序流程（GOTO）

不可否认异常将使你的程序流程不再是线性的。那异常会不会将是一种新的 GOTO 的写法呢？某种程度上，异常经常被用到异常的捕获，但是他也是改变程序流程的一种方式（单独的错误）。例如程序调用如下：`A() -> B() -> C()`，有了异常你可以直接从 `C()` 跳到 `A()` 从而忽略 `B()`。这个和 GOTO 很像。但是它更厉害因为当在跳转的时候你有完整的数据元（例如栈等信息）

Dijkstra 的论文["Go To Statement Considered Harmful"](https://homepages.cwi.nl/~storm/teaching/reader/Dijkstra68.pdf) 是非常有名的。很多编程人认为 GOTO 应该永远都不要被使用，这个其实有点言过其实了。这篇论文倡导结构化编程而不是废除 GOTO！认为 GOTO 是有害的观点是在有一段时间里，当时编程者还没有使用结构化流程控制程序，例如 `while/for/if`, 他们所知道的就是机器码，寄存器和跳转（这篇论文是 60 年代的）。

如果你写过大的 C 的项目，并且尽最大的努力使它安全，那么你将会使用 goto 语句或者写大量的重复代码，这个是你的选择。

## 总结

异常是一个编程语言的功能，而不是反模式。 小心处理错误，尽量使你的程序保证安全，但是别忘记安全也是有隐患的。

---

via: https://blog.hltbra.net/2019/02/26/exceptions-vs-returns.html

作者：[hltbra](https://blog.hltbra.net/)
译者：[amei](https://github.com/amei)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
