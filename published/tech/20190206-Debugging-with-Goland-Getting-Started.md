首发于：https://studygolang.com/articles/20655

# 使用 Goland 调试 - 起步

*由 [Florin Pățan](https://blog.jetbrains.com/go/author/florin-patanjetbrains-com/) 发布于 2019 年 2 月 6 日*

调试是任何一个现代应用的生命周期中的必要部分。

对于经常使用调试器的开发者而言，调试不仅对于发现 bug 很有用，也有助于查看和理解他们即将用到的新代码库中发生了什么，或是学习一门新的语言到底是怎么回事。

一般来说，大家比较喜欢的调试风格有两种：

* 打印语句：在代码执行到各个步骤时进行记录。
* 使用类似 [Delve](https://github.com/go-delve/delve) 的调试器，或直接使用 IDE：这能让我们在程序执行过程中有更多的控制力，提供更多可查看代码的功能，可能这些功能并没有包含在原始的打印语句中，甚至可以在运行时改值，或者在运行时来回切换（单步调试）。

这一系列文章中我们将重点讨论第二种风格，使用 IDE 来调试应用。

就像上文描述的那样，您会发现这种做法提供了更多可以用于找到 bug 的方法和功能。本文将分为以下几个部分：

* [调试应用](#debugging-an-application)
* [调试测试](#debugging-tests)
* [本机调试运行中的应用](#debugging-local)
* [在远程机器上调试运行中的应用](#debugging-remote)

在我们了解完以上几个场景后，我们会知道 GoLand 是如何处理这些场景的，以至于无论您的应用在何处运行，您都拥有以下列出的同样的特性集：

* 调试的基础
  * 控制执行流
  * 评估表达式
  * 监测自定义值
  * 改变变量值
  * 使用断点

IDE 支持调试 Linux 上生成的内存转储，也支持在 Linux 上使用 Mozilla 的 rr 可逆调试器。我们将在接下来的博客中分别看到这些特性。

对于以上几点的调试，我们都将使用一个简单的 Web 服务器，但其实它们可应用于任何种类的应用，像是客户端工具、图形界面应用等等。

我们使用 Go Modules，当然，基于默认 GOPATH 的其他依赖管理方式同样也没有问题。

使用 *__Go Modules (vgo)__* 类型创建应用，务必确保您的 Go 版本是 1.11+。

如果您没有 Go 1.11 或您想要使用 *__GOPATH__* 模式，选择 *__GO__*

应用是下面这样的：

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const (
	readTimeout  = 5
	writeTimeout = 10
	idleTimeout  = 120
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	returnStatus := http.StatusOK
	w.WriteHeader(returnStatus)
	message := fmt.Sprintf("Hello %s!", r.UserAgent())
	w.Write([]byte(message))
}

func main() {
	serverAddress := ":8080"
	l := log.New(os.Stdout, "sample-srv ", log.LstdFlags|log.Lshortfile)
	m := mux.NewRouter()

	m.HandleFunc("/", indexHandler)

	srv := &http.Server{
		Addr:         serverAddress,
		ReadTimeout:  readTimeout * time.Second,
		WriteTimeout: writeTimeout * time.Second,
		IdleTimeout:  idleTimeout * time.Second,
		Handler:      m,
	}

	l.Println("server started")
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
```

我们也可以像这样创建一个测试文件：

```go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIndexHandler(t *testing.T) {
	tests := []struct {
		name           string
		r              *http.Request
		w              *httptest.ResponseRecorder
		expectedStatus int
	}{
		{
			name:           "good",
			r:              httptest.NewRequest("GET", "/", nil),
			w:              httptest.NewRecorder(),
			expectedStatus: http.StatusOK,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			indexHandler(test.w, test.r)
			if test.w.Code != test.expectedStatus {
				t.Errorf("Failed to produce expected status code %d, Got %d", test.expectedStatus, test.w.Code)
			}
		})
	}
}
```

对于支持 *__GO Modules__* 的应用，您可以使用快捷键 *__Alt+Enter__* 然后 *__Sync packages of \<my project\>__*。

对于不支持 *__GO Modules__* 的应用，您可以使用快捷键 *__Alt+Enter__* 然后 *__go get -t \<missing dependency\>__*。

最终我们会发现，我们用于编译项目的 Go 版本也会影响到我们调试的体验。随着每一个 Go 版本的发布，Go 小组成员都会添加更多的调试信息，并提升现有调试信息的质量。比如，我们将 Go 1.8 更新到 Go 1.9 就会发现这些提升的变化，如果我们从 Go  1.8 更新到 Go 1.11，提升效果就会更加明显。因此，您用越新的 Go 版本，就会有越好的体验。

好啦，我们所有的代码都已就位，开始调试吧！

## <a name="debugging-an-application"></a> 调试应用

我们可以点击绿色三角，然后选择 `Debug 'go build main.go'` 来调试程序。

或者我们也可以右击文件夹选择 *__Debug | Go build \<project name\>__*。

![1st_gif](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/1-optimized.gif)

## <a name="debugging-tests"></a> 调试测试

跟调试应用很相似，GoLand 会从标准 *__testing__* 包，*__gocheck__*，和*__testify__* 框架来识别测试，所以这些操作可以在编辑器窗口直接使用。

对于其他框架，您可能需要在 *__Run | Edit Configurations...__* 中配置自定义的测试运行器，并在 *__Go tool arguments__* 或 *__Program arguments__* 中指定额外参数。这取决于您使用的自定义库需要哪些参数。

![2nd_gif](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/2-optimized.gif)

## <a name="debugging-local"></a> 本机调试运行中的应用

以下有几个您可能会想要在 IDE 外启动调试的应用案例。

其中一个案例是在本地机器上运行的应用。

为了使用调试器运行该案例，需要在 IDE 中打开项目，并选择 *__Attach to Process …__*

如果这是您第一次使用这个特性，IDE 会让您去下载一个叫做 Gops 的小型工具程序，该程序可在 [https://github.com/google/gops](https://github.com/google/gops) 中获的。这个程序帮助 IDE 找到在您机器上运行着的 Go 进程。然后再次调用 *__Attach to Process …__* 特性。

您将会看到在您电脑上运行着的所有的 Go 项目的列表，谁知道呢，也许您甚至会发现一些新的东西呢。从列表中选择您想要调试的项目，调试器连接到该进程，您就可以开始调试了。

为了确保调试的成功，以及调试时不会出现什么问题，您要做的就是用一个特别的标识来编译您的应用。IDE 会自动为其他配置类型添加标识，因此只有在手动编译应用的时候才需要添加标识。

如果您的程序运行在 Go 1.10 及以上版本，您需要添加 `-gcflags="all=-N -l"` 到 `go build` 命令。

如果您的程序运行在 Go 1.9 及以下版本，您需要添加 `-gcflags="-N -l"` 到 `go build` 命令。

**重要提示！** 有的人也用 `-ldflags="all=-w"` 或 `-ldflags="-w"`， 这取决于其使用的 Go 版本。

这与调试应用是不兼容的，因为它丢弃了 Delve 所需的必要的 DWARF 信息。

这样的话将无法调试应用。

当在支持此特性的操作系统或文件系统上使用软链接或符号链接时，也会遇到类似的问题。由于 Go 工具链、Delve 和 IDE 之间的不兼容性，目前使用符号链接与调试应用不能兼容。

![3rd_gif](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/3-optimized.gif)

## <a name="debugging-remote"></a> 在远程机器上调试运行中的应用

最后，这个案例更加复杂，至少现在来看是这样。这中调试类型允许您连接 IDE 到远程机器来调试一个运行着的进程。通过这种方式，我们可以将运行在本地的容器作为远程目标，也可以把在内部或者云上的服务器作为远程目标。

与在本地运行相比，您要更加小心的使用编译器标识去编译应用。然后，您需要用与您的应用相同的 Go 版本和主机/目标主机来编译 Delve ，因为不同的操作系统之间可能存在一些细微的差异，这有可能导致您无法按照预期进行调试。

您还应该确保的是，如果您在使用 *__$GOPATH__* ，那么项目也是在与 *__$GOPATH__* 同一相对路径编译的。例如：如果您的项目在 *__github.com/JetBrains/go-sample__* 下是可用的，那么无论是 IDE 所在的机器上还是在应用编译的机器上，其应用所在的路径都是 **$GOPATH/src/github.com/JetBrains/go-sample** ，这两台机器上的 *__$GOPATH__* 可能是不同的。IDE 会在本地和远程机上自动映射源代码。

当你部署你应用的时候，还要部署之前被编译的 Delve 的副本，你有两种启动测试的选项：

* 让调试器运行进程：如果你选择了这个选项，你需要运行 `dlv --listen=:2345 --headless=true --api-version=2 exec ./application` 。还要注意如果你使用了防火墙或容器，你就需要将 *__2345__* 这个端口暴露给那些配置。端口号可以是你想要的任意值，不一定非得是  *__2345__*，只要是主机上空闲的就行。
* 附加到进程中：你需要运行 `dlv --listen=:2345 --headless=true --api-version=2 attach <pid>` ，*__\<pid\>__* 是你应用的进程 id 。

这些都完事了之后，最后一步是将你的 IDE 连接到远程调试器。你可以通过 ***Run | Edit Configurations … | + | Go Remote*** ，然后配置主机和你的远程调试器监听的端口号进行连接。

![4th_gif](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/4-optimized.gif)

你可以使用如下 Dockerfile 中的容器定义：

```
FROM Golang:1.11.5-alpine3.8 AS build-env

ENV CGO_ENABLED 0

# Allow Go to retreive the dependencies for the build step
RUN apk add --no-cache Git

WORKDIR /goland-debugging/
ADD . /goland-debugging/

RUN Go build -o /goland-debugging/srv .

# Get Delve from a GOPATH not from a Go Modules project
WORKDIR /go/src/
RUN Go get Github.com/go-delve/delve/cmd/dlv

# final stage
FROM alpine:3.8

WORKDIR /
COPY --from=build-env /goland-debugging/srv /
COPY --from=build-env /go/bin/dlv /

EXPOSE 8080 40000

CMD ["/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "exec", "/srv"]
```

请注意在这个 Dockerfile 中，项目被命名为 Goland-debugging，但你可以将文件夹名字改为与你创建的项目相匹配的名字。
运行 Docker 容器时，你也需要为它指定 `--security-opt="apparmor=unconfined" --cap-add=SYS_PTRACE` 参数。如果你是在命令行中执行，这些就是 *__docker run__* 命令的参数。如果你是在 IDE 中执行的话，这些选项必须被放在 *__Run__* 选项字段中。

![last_img](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-with-goland/5-optimized.png)

这就是今天的所有内容了。在这个系列的下一篇我们将学到如何在上述的调试场景中使用各种可用的特性。请让我们在下面的评论区，或 [Twitter](https://twitter.com/GoLandIDE) 上得到您的反馈，或者您可以在我们的 [issue tracker](https://youtrack.jetbrains.com/issues/Go) 上新开一个 issue 。

---

via: https://blog.jetbrains.com/go/2019/02/06/debugging-with-goland-getting-started/

作者：[Florin Pățan](https://blog.jetbrains.com/go/author/florin-patanjetbrains-com/)
译者：[130-133](https://github.com/130-133)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
