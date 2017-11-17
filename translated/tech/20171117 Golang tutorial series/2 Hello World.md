# Hello World
这是Golang系列教程的第2个教程。如果想要了解什么是Golang，以及如何安装Golang，请阅读Golang教程第1部分：介绍与安装。

学习一种编程语言的最好方法就是去动手实践，编写代码。让我们开始编写第一个Go程序吧。

我个人推荐使用添加了[Go扩展](https://marketplace.visualstudio.com/items?itemName=lukehoban.Go)的[Visual Studio Code](https://code.visualstudio.com/)作为IDE。它具有自动补全、编码规范(code styling)以及许多其他的特性。

## 建立Go工作区
在编写代码之前，我们首先应该建立Go的工作区(workspace)。

在**Mac或Linux**操作系统下，Go工作区应该设置在$HOME/go。所以我们要在$HOME目录下创建go目录。

而在**Windows**下，工作区应该设置在**C:\Users\YourName\go**。所以请将go目录放置在C:\Users\YourName。

其实也可以通过设置GOPATH环境变量，用其他目录来作为工作区。但为了简单起见，我们采用上面提到的放置方法。

所有Go源文件都应该放置在工作区里的**src**目录下。请在刚添加的**go**目录下面创建目录**src**。

所有Go项目都应该依次在src里面设置自己的子目录。我们在src里面创建一个目录**hello**来放置整个hello world项目。

创建上述目录之后，其目录结构如下：
```
go
  src
    hello
```

在我们刚刚创建的hello目录下，在helloworld.go文件里保存下面的程序。
```golang
package main

import "fmt"

func main() {  
    fmt.Println("Hello World")
}
```
创建该程序之后，其目录结构如下：
```
go
  src
    hello
      helloworld.go
```
## 运行Go程序
运行Go程序有多种方式，我们下面依次介绍。
1. 使用**go run**命令 - 在命令提示符旁，输出`go run workspacepath/src/hello/helloworld.go`。

上述命令中的**workspacepath**应该替换为你自己的工作区路径（Windows下的**C:/Users/YourName/go**，linux或Mac下的$HOME/go）。

在控制台上会看见`Hello World`的输出。

2. 使用**go install**命令 - 在`workspacepath/bin/hello`目录下，输入`go install hello`命令来运行程序。

上述命令中的**workspacepath**应该替换为你自己的工作区路径（Windows下的**C:/Users/YourName/go**，linux或Mac下的$HOME/go）。

当你输出**go install hello**时，go工具会在工作区中搜索hello包（hello称之为包，我们后面会更加详细地讨论包）。接下来它会在工作区的bin目录下，创建一个名为`hello`（windows下名为`hello.exe`）的二进制文件。运行go install hello后，其目录结构如下所示：
```
go
  bin
    hello
  src
    hello
      helloworld.go
```
3. 第3种运行程序的好方法是使用go playground。尽管它有自身的限制，但该方法对于运行简单的程序非常方便。我已经创建了一个hello world程序的playground。[点击这里](https://play.golang.org/p/VtXafkQHYe)在线运行程序。
你可以使用[go playground](https://play.golang.org)与其他人分享你的源代码。
### 简述hello world程序
下面就是我们刚写下的hello world程序。
```golang
package main //1

import "fmt" //2

func main() { //3  
	fmt.Println("Hello World") //4
}
```
现在简单介绍每一行大概都做了些什么。在以后的教程中还会深入探讨每个部分。

**package main - 每一个Go文件都应该在开头进行`package name`的声明**。包(packages)用于代码的封装与重用。这里的包名称是`main`。

**import "fmt"** - 我们引入了fmt包，用于在main函数里面打印文本到标准输出。

**func main()** - main是一个特殊的函数。整个程序就是从main函数开始运行的。**main函数必须放置在main包中**。{和}分别表示main函数的开始和结束部分。

**fmt.Println("Hello World")** - **fmt**包中的**Println**函数用于把文本写入标准输出。

该代码可以在[github](https://github.com/golangbot/hello)下载。

现在你可以进入**Golang系列教程第3部分：变量**中学习Golang中的变量。

请提供给我们宝贵的反馈和意见。感谢您的阅读:)

**下一教程 - 变量**

via: https://golangbot.com/hello-world/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[Noluye](https://github.com/Noluye)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
