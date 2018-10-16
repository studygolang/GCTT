首发于：https://studygolang.com/articles/14728

# 在 Go 语言中使用 Log 包

Linux 在许多方面相对于 Windows 来说都是独特的，在 Linux 中编写程序也不例外。标准输出，标准 err 和 null devices 的使用不仅是一个好主意，也是一个原则。如果您的程序将记录日志信息，则最好遵循目标约定。这样，您的程序将兼容所有 Mac/Linux 工具和托管环境。

Go 在标准库中有一个 log 包和 logger 类型。使用 log 包将为您提供成为优秀公民 (译注：指 log 包兼容性非常好) 所需的一切。您将能够写入所有标准设备，自定义文件或支持 io.Writer 接口的任何目标。

我提供了一个非常简单的示例，它将帮助您开始使用 logger ：

```go
package main

import (
    "io"
    "io/ioutil"
    "log"
    "os"
)

var (
    Trace   *log.Logger
    Info    *log.Logger
    Warning *log.Logger
    Error   *log.Logger
)

func Init(
    traceHandle io.Writer,
    infoHandle io.Writer,
    warningHandle io.Writer,
    errorHandle io.Writer) {

    Trace = log.New(traceHandle,
        "TRACE: ",
        log.Ldate|log.Ltime|log.Lshortfile)

    Info = log.New(infoHandle,
        "INFO: ",
        log.Ldate|log.Ltime|log.Lshortfile)

    Warning = log.New(warningHandle,
        "WARNING: ",
        log.Ldate|log.Ltime|log.Lshortfile)

    Error = log.New(errorHandle,
        "ERROR: ",
        log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
    Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

    Trace.Println("I have something standard to say")
    Info.Println("Special Information")
    Warning.Println("There is something you need to know about")
    Error.Println("Something has failed")
}
```

运行此程序时，您将获得以下输出：

```
INFO: 2013/11/05 18:11:01 main.go:44: Special Information
WARNING: 2013/11/05 18:11:01 main.go:45: There is something you need to know about
ERROR: 2013/11/05 18:11:01 main.go:46: Something has failed
```

您会注意到没有显示 Trace logging (译注：跟踪记录器)。让我们看看代码，找出原因。

查看 Trace logging 部分的代码：

```go
var Trace *log.Logger

Trace = log.New(traceHandle,
    "TRACE: ",
    log.Ldate|log.Ltime|log.Lshortfile)

Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

Trace.Println("I have something standard to say")
```

该代码创建一个名为 Trace 的包级变量，它是一个指向 log.Logger 对象的指针。然后在 Init 函数内部创建一个新的 log.Logger 对象。log.New 函数的参数如下：

```go
func New(out io.Writer, prefix string, flag int) *Logger
```

```
out:    The out variable sets the destination to which log data will be written. // 译注 out 变量设置将写入日志数据的目标
prefix: The prefix appears at the beginning of each generated log line. // 译注 前缀出现在每个生成的日志行的开头。
flags:  The flag argument defines the logging properties. // 译注 flag 参数定义日志记录属性
```

Flags:

```go
const (
    // Bits or’ed together to control what’s printed. There is no control over the
    // order they appear (the order listed here) or the format they present (as
    // described in the comments). A colon appears after these items:
    // 2009/01/23 01:23:23.123123 /a/b/c/d.go:23: message
    Ldate = 1 << iota // the date: 2009/01/23
    Ltime             // the time: 01:23:23
    Lmicroseconds     // microsecond resolution: 01:23:23.123123. assumes Ltime.
    Llongfile         // full file name and line number: /a/b/c/d.go:23
    Lshortfile        // final file name element and line number: d.go:23. overrides Llongfile
    LstdFlags = Ldate | Ltime // initial values for the standard logger
)
```

在此示例程序中，Trace 的目标是 ioutil.Discard 。这是一个 null device （译注：对应 /dev/null 相当于垃圾桶，消息直接丢弃），所有写入调用都可以成功而不做任何事情。因此，使用 Trace 写入时，终端窗口中不会显示任何内容。

再来看看 Info 的代码：

```go
var Info *log.Logger

Info = log.New(infoHandle,
    "INFO: ",
    log.Ldate|log.Ltime|log.Lshortfile)

Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

Info.Println("Special Information")

```

对于 Info (译注：消息记录器)，os.Stdout 传入到 init 函数给了 infoHandle 。这意味着当您使用 Info 写消息时，消息将通过标准输出显示在终端窗口中。

最后，看下 Error 的代码：

```go
var Error *log.Logger

Error = log.New(errorHandle,
    "ERROR: ",  // 译注： 原文是 INFO 与原始定义不同，应该是笔误，故直接修改
    log.Ldate|log.Ltime|log.Lshortfile)

Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

Error.Println("Something has failed") // 译注： 原文是 Special Information 与原始定义不同，应该是笔误，故直接修改
```

这次 os.Stderr 传入到 Init 函数给了 errorHandle 。这意味着当您使用 Error 写消息时，该消息将通过标准错误显示在终端窗口中。但是，将这些消息传递给 os.Stderr 允许运行程序的其他应用程序知道发生了错误。

由于支持 io.Writer 接口的任何目标都可以接受，因此您可以创建和使用文件：

```go
file, err := os.OpenFile("file.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatalln("Failed to open log file", output, ":", err)
}

MyFile = log.New(file,
    "PREFIX: ",
    log.Ldate|log.Ltime|log.Lshortfile)
```

在示例代码中，打开一个文件，然后将其传递给 log.New 函数。现在，当您使用 MyFile 进行写入时，数据将写到 file.txt 里。

您还可以让 logger （译注：记录器） 同时写入多个目标。

```go
file, err := os.OpenFile("file.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatalln("Failed to open log file", output, ":", err)
}

multi := io.MultiWriter(file, os.Stdout)

MyFile := log.New(multi,
    "PREFIX: ",
    log.Ldate|log.Ltime|log.Lshortfile)
```

这里数据将写入到文件和标准输出里。

注意在处理 OpenFile 的任何错误时使用 log.Fatalln 方法。 log 包提供了一个可以配置的初始 logger。以下是使用具有标准配置的日志的示例程序：

```go
package main

import (
    "log"
)

func main() {
    log.Println("Hello World")
}
```

以下是输出：

```
2013/11/05 18:42:26 Hello World
```

如果想要删除或更改输出格式，可以使用 log.SetFlags 方法：

```go
package main

import (
    "log"
)

func main() {
    log.SetFlags(0)
    log.Println("Hello World")
}
```

以下是输出：

```
Hello World
```

现在所有格式都已删除。如果要将输出发送到其他目标，请使用 log.SetOutput ：

```go
package main

import (
    "io/ioutil"
    "log"
)

func main() {
    log.SetOutput(ioutil.Discard)
    log.Println("Hello World")
}
```

现在终端窗口上不会显示任何内容。您可以使用任何支持io.Writer 接口的目标。

基于这个例子，我为我的所有程序编写了一个新的日志包：

go get github.com/goinggo/tracelog

我希望在开始编写Go程序时我就知道 log 和 loggers 。期待将来能够看到我写的更多日志包。

---

via: https://www.ardanlabs.com/blog/2013/11/using-log-package-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[chaoshong](https://github.com/chaoshong)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
