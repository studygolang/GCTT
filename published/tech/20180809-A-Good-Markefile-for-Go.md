首发于：https://studygolang.com/articles/14919

# 一个好的 Go 语言 Makefile 是怎样的

精简的 Makefile，用于简化构建和管理用 Go 编写的 Web 服务器。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/a-good-makefile-for-go/makefile-process.gif)

我偶然调整我的 Makefiles 来加快我的开发过程，一天早上有点时间，我决定与大家分享其中诀窍。

总而言之，我使用 Go 构建 Web 服务器，我对 Makefile 的期望如下：

- 高级，简单的命令。比如：`compile` `start` `stop` `watch` 等等
- 管理具体项目环境的变量，它应该包含 `.env` 文件
- 开发模式，修改时自动编译
- 开发模式，修改时自动重启服务
- 开发模式，简洁地显示编译的错误信息
- 具体项目的 GOPATH，以使我可以在 `vendor` 目录维护依赖包
- 简化文件查看，比如 `make watch run="go test ./..."`

下面是我偏爱的文件目录布局：

```
.env
Makefile
main.go
bin/
src/
vendor/
```

在此文件结构中键入 make 命令将提供以下输出：

```shell
$ make

Choose a command run in my-web-server:

install   Install missing dependencies. Runs `go get` internally.
start     Start in development mode. Auto-starts when code changes.
stop      Stop development mode.
compile   Compile the binary.
watch     Run given command when code changes. e.g; make watch run="go test ./..."
exec      Run given command, wrapped with custom GOPATH. e.g; make exec run="go test ./..."
clean     Clean build files. Runs `go clean` internally.
```

## 1、一步一步开始

### 环境变量

首先，我们希望在 `Makefile` 中 include 我们为项目定义的环境变量，所以，第一行如下：

```
include .env
```

在具体项目的环境变量文件的头部，我们将定义这些：项目名，Go 目录/文件，进程 id 的路径...

```makefile
PROJECTNAME=$(shell basename "$(PWD)")

# Go related variables.
GOBASE=$(shell pwd)
GOPATH=$(GOBASE)/vendor:$(GOBASE)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Redirect error output to a file, so we can show it in development mode.
STDERR=/tmp/.$(PROJECTNAME)-stderr.txt

# PID file will store the server process id when it's running on development mode
PID=/tmp/.$(PROJECTNAME)-api-server.pid

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent
```

在 Makefile 文件的其余部分，我们将使用特别的 GOPATH 变量。我们所有的命令都应该包含项目特定的 GOPATH，否则它们将无法工作。这为我们的 Go 项目提供了明确的隔离，并带来了一些复杂性。为了简化操作，我们可以添加一个 `exec` 命令，该命令执行任何给定的命令，并使用上面定义的自定义 GOPATH。

```makefile
## exec: Run given command, wrapped with custom GOPATH. e.g; make exec run="go test ./..."
exec:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) $(run)
```

但这并不是很高级的做法。我们应该用简单的命令来介绍一些常见的情况，如果我们正在做 Makefile 未涵盖的事情，那么只能使用 `exec` 。

### 开发模式

开发模式应该是这样：

- 清空构建缓存
- 编译代码
- 后台执行服务
- 当代码被修改时，重复上面步骤

这听起来很简单，但很快变得复杂，我们将同时运行服务器和文件观察程序。我们需要确保在开始新流程之前正确停止服务，并且也不会破坏常见的命令行行为，例如在按下 Control-C 或 Control-D 时停止。

```makefile
start:
	bash -c "trap 'make stop' EXIT; $(MAKE) compile start-server watch run='make compile start-server'"

stop: stop-server
```

下面是代码解决的问题：

- 后台编译与执行服务
- 主进程不在后台执行，我们可以随时使用 Control-C 中断它
- 当主进程中断时，停止后台进程，为此，我们需要 `trap`
- 当代码修改时，重新编译与重启服务

在下一节中，我将详细解释这些命令

### 编译

`compile` 命令不仅仅是在后台调用 `go compile` ; 它还清理错误输出并打印简化版本。

以下是在命令行中进行重大更改的方法：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/a-good-makefile-for-go/compile-print-error.png)

```makefile
compile:
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -s go-compile 2> $(STDERR)
	@cat $(STDERR) | sed -e '1s/.*/\nError:\n/'  | sed 's/make\[.*/ /' | sed "/^/s/^/     /" 1>&2
```

### 开启/停止服务

`start-server` 基本上运行它在后台编译的二进制文件，将其 PID 保存到临时文件中。

`stop-server` 读取 PID 并在需要时终止进程。

```makefile
start-server:
	@echo "  >  $(PROJECTNAME) is available at $(ADDR)"
	@-$(GOBIN)/$(PROJECTNAME) 2>&1 & echo $$! > $(PID)
	@cat $(PID) | sed "/^/s/^/  \>  PID: /"

stop-server:
	@-touch $(PID)
	@-kill `cat $(PID)` 2> /dev/null || true
	@-rm $(PID)

restart-server: stop-server start-server
```

### 观察变化

我们需要一个文件观察器来观察变化。我尝试了很多并且感到不满意，所以最终创建了我自己的文件观察工具 [yolo](https://github.com/azer/yolo) 。通过下面命令安装在您的系统中

```shell
$  go get github.com/azer/yolo
```

一旦安装完毕，我们基本上可以开始观察项目目录中的更改，排除像 `vendor` 或者 `bin` 这样的目录，如下：

```makefile
## watch: Run given command when code changes. e.g; make watch run="echo 'hey'"
watch:
	@yolo -i . -e vendor -e bin -c $(run)
```

现在我们得到一个 `watch` 命令，它在项目目录中以递归方式监视更改，不包括 `vendor` 目录。我们可以直接传递我们想要的任何运行命令。例如，`start` 命令基本上在代码更改时调用 `make compile start-server`：

```shell
make watch run="make compile start-server"
```

我们可以用它来运行测试，或自动检查是否有任何竞争条件。将为执行设置环境变量，因此您根本不必担心 GOPATH：

```shell
make watch run="go test ./..."
```

关于 Yolo 的一个好处是它的网络界面。如果启用它，您可以立即在 Web 界面中看到命令的输出。您只需要传递 `-a` 选项来启用它：

```shell
yolo -i . -e vendor -e bin -c "go run foobar.go" -a localhost:9001
```

然后，您可以在浏览器中打开 `localhost:9001` 并立即开始在浏览器中查看结果：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/a-good-makefile-for-go/yolo-process.gif)

### 安装依赖

当我们在代码中进行更改时，我们希望在编译之前下载缺少的依赖项。`install` 命令将为我们完成这项工作;

```makefile
install: go-get
```

我们在文件改动时，编译之前自动调用 `install` , 让依赖包可以自动安装，如果您想手动安装依赖项，你可以执行:

```shell
make install get="github.com/foo/bar"
```

在内部，这个命令会转换成：

```shell
$ GOPATH=~/my-web-server GOBIN=~/my-web-server/bin go get github.com/foo/bar
```

它是如何工作的？请参阅下一节，我们实际添加了用于实现更高级别命令的 Go 命令。

### Go 命令

如果我们想设置 GOPATH 到项目的目录，简化依赖管理（在 Go 生态中还没正式解决的问题），就需要在 Makefile 中封装 Go 命令。

```makefile
go-compile: go-clean go-get go-build

go-build:
	@echo "  >  Building binary..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build -o $(GOBIN)/$(PROJECTNAME) $(GOFILES)

go-generate:
	@echo "  >  Generating dependency files..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go generate $(generate)

go-get:
	@echo "  >  Checking if there is any missing dependencies..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get $(get)

go-install:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install $(GOFILES)

go-clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean
```

### 帮助

最后，我们需要一个 help 命令来查看可用命令的概述。我们可以使用自动生成优雅的格式的命令 `sed` 与 `column` ， 如下：

```makefile
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
```

这个命令会基本地扫描 Makefile 中以 `##` 开头的文本行并输出它们。所以，你可以简单的注释你所定义的命令，这些命令会被 `help` 命令打印出来。

比如我们添加如下的注释：

```makefile
## install: Install missing dependencies. Runs `go get` internally.
install: go-get

## start: Start in development mode. Auto-starts when code changes.
start:

## stop: Stop development mode.
stop: stop-server
```

我们可以执行：

```shell
$ make help

Choose a command run in my-web-server:

install   Install missing dependencies. Runs `go get` internally.
start     Start in development mode. Auto-starts when code changes.
stop      Stop development mode.
```

## 终极版本

以下是我在上面分享的所有组合和终极版本。这是我今天早上开始的一个新项目的完美副本：

```makefile
include .env

PROJECTNAME=$(shell basename "$(PWD)")

# Go related variables.
GOBASE=$(shell pwd)
GOPATH="$(GOBASE)/vendor:$(GOBASE)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Redirect error output to a file, so we can show it in development mode.
STDERR=/tmp/.$(PROJECTNAME)-stderr.txt

# PID file will keep the process id of the server
PID=/tmp/.$(PROJECTNAME).pid

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

## install: Install missing dependencies. Runs `go get` internally. e.g; make install get=github.com/foo/bar
install: go-get

## start: Start in development mode. Auto-starts when code changes.
start:
    bash -c "trap 'make stop' EXIT; $(MAKE) compile start-server watch run='make compile start-server'"

## stop: Stop development mode.
stop: stop-server

start-server: stop-server
	@echo "  >  $(PROJECTNAME) is available at $(ADDR)"
	@-$(GOBIN)/$(PROJECTNAME) 2>&1 & echo $$! > $(PID)
	@cat $(PID) | sed "/^/s/^/  \>  PID: /"

stop-server:
	@-touch $(PID)
	@-kill `cat $(PID)` 2> /dev/null || true
	@-rm $(PID)

## watch: Run given command when code changes. e.g; make watch run="echo 'hey'"
watch:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) yolo -i . -e vendor -e bin -c "$(run)"

restart-server: stop-server start-server

## compile: Compile the binary.
compile:
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -s go-compile 2> $(STDERR)
	@cat $(STDERR) | sed -e '1s/.*/\nError:\n/'  | sed 's/make\[.*/ /' | sed "/^/s/^/     /" 1>&2

## exec: Run given command, wrapped with custom GOPATH. e.g; make exec run="go test ./..."
exec:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) $(run)

## clean: Clean build files. Runs `go clean` internally.
clean:
	@(MAKEFILE) go-clean

go-compile: go-clean go-get go-build

go-build:
	@echo "  >  Building binary..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build -o $(GOBIN)/$(PROJECTNAME) $(GOFILES)

go-generate:
	@echo "  >  Generating dependency files..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go generate $(generate)

go-get:
	@echo "  >  Checking if there is any missing dependencies..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get $(get)

go-install:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install $(GOFILES)

go-clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
```

就是这样！如果您有任何问题，想法或一些建议，以使其更好，请给我发电子邮件！

干杯！

---

via: http://azer.bike/journal/a-good-makefile-for-go/

作者：[Azer Koçulu](http://azer.bike)
译者：[lightfish-zhang](https://github.com/lightfish-zhang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
