首发于：https://studygolang.com/articles/11823

# Go语言中包的风格指南

Go 语言也有自己的命名与代码组织规则。漂亮的代码，布局清晰、易读易懂，就像是设计严谨的 API 一样。拿到代码，用户首先看到和接触的就是布局、命名还有包的结构。

这篇文章不是为了给大家设立硬性的规定，而是用实践引导大家形成自己的规则。因为具体问题要具体分析，通过自己的判断才能挑选出最恰当的规则。

## 包

所有的 Go 代码都是以包的形式组织起来的。Go 中的包就是目录或文件夹里面包括一个或者多个以 `.go` 结尾的文件。用目录或文件夹这样的方式来管理代码，与电脑管理目录或文件夹是一样一样的。

所有的 Go 代码都是放在包里的，且只能是通过包来进行访问。理解并且建立良好的包的习惯，可以帮助写出高效的代码。

---

## 包的组织

我们先聊聊如何组织 Go 代码，解释一下定位 Go 包的规范。

### 使用多个文件

一个包就是一个或者多个文件的目录。先把代码切分成符合逻辑且易读的多个文件。

例如，根据文件处理 HTTP 的内容不一样，一个 HTTP 包可以被切分成多个文件。在下面这个例子中，一个 HTTP 包被切成下列文件：头部类型定义和代码，cookie 类型定义加代码，还有实际 HTTP 功能的实现和包说明文档。

    - doc.go       // 包说明文档
    - headers.go   // HTTP 头部类型定义和代码
    - cookies.go   // HTTP cookie 类型定义和代码
    - http.go      // HTTP 客户端实现，请求和返回类型，等等

### 聚合类型定义

首要规则是，把类型定义尽量都聚合到他们被引用的地方。这让代码的维护者（不仅仅局限于代码的原作者）更易于找到类型的定义。比如，头结构体类型最好就是放在 `headers.go` 文件当中。

    $ cat headers.go
    package http

    // Header 表示一个 Http 头部结构体定义
    type Header struct {...}

虽然 Go 语言本身并没有严格地要求你必须在文件哪个部分定义类型，但是把核心类型的定义都放在文件的最上面是没错的。

### 根据功能进行安排

在其他语言中，通常都是把类型定义聚合到一个包里，叫作模型或者类别。在 Go 语言中，则是通过代码的功能职责来进行安排的。

```go
package models // 千万别叫这个名字

// User 代表系统中的一个用户
type User struct {...}
```

不要创建一个命名为 models 的包，然后在里面定义所有的实体类型。在这个例子中，User 类型应该定义在服务层的包中。

```go
package mngtservice

// User 代表系统中的一个客户
type User struct {...}

func UsersByQuery(ctx context.Context, q *Query) ([]*User, *Iterator, error)

func UserIDByEmail(ctx context.Context, email string) (int64, error)
```

### 优化 godoc

越早使用 godoc 越好，尤其是在初期设计包的 API 的时候。使用 godoc，你可以清楚知道自己的构思用文档表达出来是个什么。有时候，可视化也对设计有影响。因为 godoc 需要放在一个独立的包里，所以可以慢慢地来进行优化，让文档越来越容易理解。执行命令 `godoc -http=<hostport>` 来启动本地的 godoc 文档服务。

### 举几个例子说明一下

有些情况下，你可能没办法把所有相关的类型定义都写在一个包里。因为这样做可能会很繁琐。或者你只是想发布一个实现了单个包中通用接口的具体功能，或者这些类型是由第三方包提供的。下面给出几个例子来说明和理解这些情况。

    $ godoc cloud.google.com/go/datastore
    func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error)
    ...

NewClient 这个方法里有一个结构参数 option.ClientOptions，但是这个结构的定义既不在 datastore 包里，也不在提供所有 option 类型的 option 包里。

    $ godoc google.golang.org/extraoption
    func WithCustomValue(v string) option.ClientOption
    ...

如果你设计的 API 需要引入很多非标准库的包，那么，添加 [Go 示例](https://blog.golang.org/examples) 通常会很有用，以便为用户提供一些工作代码。

样例可以提高藏地较深的包的暴光率。例如，datastore.NewCLient 这样的结构需要引用额外的 option 包。提供样例，就可以让用户知道，还有一个 option 包。

### 不要在 main 文件中导出

标识符可以被 [导出](https://golang.org/ref/spec#Exported_identifiers)，以允许从外部包来使用它。

main 包是不能被导入的，所以从 main 包中导出标记符是没有必要的。如果你要把包编译成二进制文件，就不要从 main 包中导出标记符。

这条规则也有例外，那就是 main 包被编译成了 .so 文件、.a 文件或者 Go 插件。在这种情况下，Go 代码被其他语言通过 [cgo 的导出功能](https://golang.org/cmd/cgo/#hdr-C_references_to_Go) 使用，那标识符的导出就是必要的了。

---

## 包的命名

包的名字与导入路径，都是很重要的标识，它们会告诉你这个包里有哪些内容。按规则给包命名不仅可以提高你的代码的质量，也间接地提高这个包的使用者的代码水平。

### 只用小写

包的名字应只用小写。不要用下划线式，也不要用驼峰式。Go 官方博文 [关于包命名的综合指南](https://blog.golang.org/package-names) 中有多个不同情形的样例。

### 简短而有意义

包的名字需要简短，但应该唯一且有意义。用户从包的名字中就能直接理解这个包的作用。

避免泛泛的包名，例如 "common", "util"。

```go
import "pkgs.org/common" // 可千万别这样写
```

避免重名，万一用户要同时引入并使用这两个同名包。

如果命名上确实有困难，可能是因为设计的代码结构与整体构架本身就有问题。

### 精简引入路径

避免暴露自定义的仓库结构（repository structure）给包的用户。谨遵 GOPATH 的规定。避免在引入路径中出现包含 src/, pkg/ 命名的路径。

    github.com/user/repo/src/httputil   // 可千万别这么做，不要使用 SRC ！！

    github.com/user/repo/gosrc/httputil // 可千万别这么做，不要使用 GOSRC ！！

### 使用单数

在 Go 语言中，包的名字不要使用复数。从其他语言转过来的程序员会觉得很别扭，因为在先前使用的语言中，已经形成了使用复数的习惯。给包命名的时候，用 httputil，不用 httputils！

```go
package httputils  // 用单数，不用复数
```

### 别名也应该遵守规则

如果你在引入多个相同名字的包，你可以在本地修改这些包的名字。别名也需要遵守本文提到的规则。并没有规则指明需要修改哪一个包的名字。如果你在修改标准库包的名字，最好加一个前缀来做区别，毕竟是 “Go 标准库” 中的包，比如，可以修改为 `gourl`、`goioutil`。

```go
import (
    gourl "net/url"

    "myother.com/url"
)
```

### 强制使用虚拟 URL

`go get` 支持通过另外一种 URL 来获取包，这个 URL 与包仓库的 URL 的名字不一样。这个不一样的 URL 叫做虚拟 URL，需要准备一个页面，里面包含可被 Go 工具识别的详细的元标签。你可以使用虚拟 URL 通过自定义域名和路径来提供包的服务。

例如，

    $ go get cloud.google.com/go/datastore

在后台去查看来自 `https://code.googlesource.com/gocloud` 的源码，把它加到你的工作区当中去，这个工作区是定义在 $GOPATH/src/cloud.google.com/go/datastore 下面的。

假定 code.googlesource.com/gocloud 已经在包里了，那能不能通过这个 URL 来使用这个包呢？答案是 NO，因为开启了强制使用虚拟 URL。

实际使用中，在包里添加了一个引入声明。Go 工具就不允许从任何其他路径来引入这个包，并且会给用户一个友好地错误提示。如果你没有开启强制使用虚拟 URL，那么就会有出现两个一样的包，并且因为不同的命名空间，它们不能放在一起使用。

```go
package datastore // import "cloud.google.com/go/datastore"
```

---

## 包说明文档

记得要给包写说明文档。包说明文档最可以阐明包的功能。对于非 main 包来讲，godoc 都是以 "Package {包名}" 开头，并且附上一个描述说明。对于 main 包来讲，文档就是用来说明程序的功能的。

```go
// Package ioutil 实现了一些输入或输出效用功能
package ioutil

// gops 命令会列出所有在系统中跑的进程
package main

// helloworld 样例来展示如何使用 x 功能
package main
```

### 使用 doc.go 文件

有时候，包里的文档说明内容会很多，尤其是包括详细的用法说明与指导。将包的 godoc 移到 `doc.go` 这个文件中去。 (参考样例 [doc.go](https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/datastore/doc.go).)

---

via: https://rakyll.org/style-packages/

作者：[rakyll](https://rakyll.org/about/)
译者：[chenbrooks](https://github.com/chenbrooks)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
