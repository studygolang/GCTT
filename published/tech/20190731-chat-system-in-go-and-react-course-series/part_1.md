首发于：https://studygolang.com/articles/22423

# 使用 Go 和 ReactJS 构建聊天系统（一）：初始化项目

本节完整代码：[GitHub](https://github.com/watermelo/realtime-chat-go-react/tree/part-1-and-2)

我们将通过设置两个项目来开始这个课程。一旦我们完成了枯燥的设置，就可以开始添加新功能并构建我们的应用程序，将看到一些积极的结果！

## 目标

在这部分课程结束后，你将掌握：

- 在 `backend/` 目录创建基本的 Go 应用
- 在 `frontend/` 目录创建基本的 ReactJS 应用

通过实现这两个部分，你将能够在接下来的几节课程中为聊天系统添加一些功能。

## 准备工作

为了完成本系列教程，我们先要做以下的准备工作。

- 需要安装 `npm`
- 需要安装 `npx`。这个可以输入 `npm install -g npx` 安装。
- Go 语言版本需要满足 1.11+。
- 需要一个代码编辑器来开发这个项目，例如 VS

## 设置 Go 后端项目

如果你熟悉 Go 的话，这一步非常简单，我们首先要在项目目录中创建一个名为 `backend` 的新目录。

这个 `backend` 目录将包含该项目的所有 Go 代码。然后，我们将通过以下命令来初始化我们的项目：

```shell
$ cd backend
$ export GO111MODULE=on
$ go mod init github.com/TutorialEdge/realtime-chat-go-react
```

应该在 `backend` 目录中使用 go modules 初始化我们的项目，初始化之后我们就可以开始写项目并使其成为一个完整的 Go 应用程序。

- **go.mod** - 这个文件有点像 NodeJS 项目中的 package.json。它详细描述了我们项目所需的包和版本，以便项目的构建和运行。
- **go.sum** - 这个文件用于校验，它记录了每个依赖库的版本和哈希值。

> 注意 - 有关 Go modules 新特性的更多信息，请查看官方 Wiki 文档: [Go Modules](https://github.com/golang/go/wiki/Modules)

## 检查 Go 项目

一旦我们在 `backend/` 目录中调用了 `go mod init`，我们将检查一下一切是否按预期工作。

在 `backend/` 目录中添加一个名为 `main.go` 的新文件，并在其中添加以下 Go 代码：

```go
package main

import "fmt"

func main() {
	fmt.Println("Chat App v0.01")
}
```

将该内容保存到 `main.go` 后，运行后会得到如下内容：

```shell
$ go run main.go
Chat App v0.01
```

如果成功执行，我们可以继续设置我们的前端应用程序。

## 设置 React 前端项目

设置前端会稍微复杂一点，首先我们要在项目的根目录中创建一个 `frontend` 目录，它将容纳我们所有的 ReactJS 代码。

> 注意 - 我们将使用 [facebook/create-react-app](https://github.com/facebook/create-react-app) 来生成我们的 React 前端。

```shell
$ cd frontend
```

然后，你需要使用 `create-react-app` 包创建一个新的 ReactJS 应用程序。这可以用 `npm` 安装：

```shell
$ npm install -g create-react-app
```

安装完成后，你应该能够使用以下命令创建新的 ReactJS 应用程序：

```shell
$ npx create-react-app .
```

运行这些命令之后，你应该可以看到我们的 `frontend/` 目录生成了基本的 ReactJS 应用程序。

我们的目录结构应如下所示：

```shell
node_modules/
public/
src/
.gitignore
package.json
README.md
yarn.lock
```

## 本地运行 ReactJS 程序

现在已经成功创建了基本的 ReactJS 应用程序，我们可以测试一下是否正常。输入以下命令来运行应用程序：

```shell
$ npm start
```

如果一切正常的话，将会看到 ReactJS 应用程序编译并在本地开发服务器上运行：[http://localhost:3000](http://localhost:3000)

```plain
Compiled successfully!

You can now view frontend in the browser.

	Local:            http://localhost:3000/
	On Your Network:  http://192.168.1.234:3000/

Note that the development build is not optimized.
To create a production build, use yarn build.
```

现在已经拥有有一个基本的 ReactJS 应用程序了，我们可以在接下来的教程中进行扩展。

## 总结

太棒了，现在已经成功设置了我们项目的前端和后端部分，接下来我们可以添加一些酷炫的新功能。

> 下一节：Part 2 - [后端实现](https://studygolang.com/articles/22426)

---

via: https://tutorialedge.net/projects/chat-system-in-go-and-react/part-1-initial-setup/

作者：[Elliot Forbes](https://twitter.com/elliot_f)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
