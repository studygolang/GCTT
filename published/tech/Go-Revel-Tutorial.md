已发布：https://studygolang.com/articles/12898

# Go/Revel教程：在浏览器（使用 PaizaCloud IDE）上，构建 Go web 框架 Revel 的应用程序

![gopher](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180323134353.png)

Go 语言（golang）的特性有：

- 标准库有很多功能，如网络。
- 易于编写并发程序。
- 易于管理可执行文件（因为只有一个文件）

由于这些特点，Go 语言在 web 开发中也越发受到欢迎。

如下图所示，我们可以在 Google Trends 看到 Go 受关注的程度。

![From google trends](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180323105259.png)

来自 [google trends](https://trends.google.com/trends/explore?date=2010-02-23%202018-03-23&q=golang)

虽说 Go 本身自带着丰富的标准库，帮助我们构建 web 应用，但是使用 web 应用框架，我们能够更轻松地开发出功能齐全的 web 应用。

Go 的 web 框架有很多：Revel、Echo、Gin、Iris 等，**其中 Revel 是最受欢迎的全栈 web 应用框架之一**。

Go 框架 Revel 的 web 开发功能有：路由、MVC、生成器。按照 Revel 的规则来构建应用，你可以轻而易举地创建可读性强、易扩展的 Web 应用程序。在 Revel 中，你还可以使用 OR 映射库（如 Gorm）。

然而，要在实际中开发 Revel 应用，你需要安装和配置 Go、Revel、Gorm 和 数据库。这些安装和设置都很麻烦。仅仅根据安装说明进行，常常会出错，或者因为 OS、版本和软件依赖等原因引起各种错误。

同样，如果你发布这项服务，朋友和其他人的反馈的确会让你动力十足。但是，这项服务还需要“部署”。“部署”同样也很难搞。

所以，[PaizaCloud](https://paiza.cloud/) 这个 Cloud IDE 应运而生。这是一个基于浏览器的在线 web 和应用开发环境。

**由于 PaizaCloud 拥有 Go/Revel 应用的开发环境，因此你可以直接在你的浏览器中，开始编写你的 Go/Revel 应用程序**。

**由于你可以在云上进行开发，因此你可以在同一台机器运行并部署 Go/Revel 应用，而无需再配置一台服务器**。

现在，我们来使用 Go 和 Revel，在 PaizaCloud IDE 上开发一个任务清单（Task List）的应用。

遵循下面的指示，**只需十分钟你就能够创建并运行 Google Home 应用程序**。

## 起步：[PaizaCloud Cloud IDE](https://paiza.cloud/en/)

[这里](https://paiza.cloud/)是 [PaizaCloud Cloud IDE](https://paiza.cloud/) 的网站。

![paiza cloud](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20171214154059.png)

可以用邮箱注册，并在确认邮件中点击链接。你还可以用 GitHub 或 Google 账户来注册。

## 创建服务器

在开发工作区上，我们创建一个新的服务器。

![new server](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20171214154558.png)

点击 `New Server`，会打开一个对话框来设置服务器。

这里，你可以选择 `PHP`、`phpMyAdmin` 和 `MySQL`，并点击 `New Server` 按钮。

![server settings](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20171214154330.png)

只需 3 秒钟，你就有了一个基于浏览器的 Go/Revel 开发环境。

你将在页面中看到编辑器或浏览器窗口，目前我们可以先关闭它们。

## 设置环境

现在我们来设置环境。由于 PaizaCloud 已经安装了 Go 语言和 MySQL，你可以直接运行 `go get` 命令来安装其他的包。

在 [PaizaCloud Cloud IDE](https://paiza.cloud/)，你可以在浏览器中，使用 PaizaCloud 的应用程序 `Terminal` 来运行命令。

在页面的左边，点击 `Terminal` 按钮。

![terminal](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20171214154805.png)

现在我们启动了 `Terminal` 程序。我们现在要在终端输入 `go get [package name]` 命令。

`[package name]` 是安装包的名字。在这里，我们要为 Revel 和 Gorm 安装相应的包。

我们输入：

```bash
$ go get github.com/revel/revel
$ go get github.com/revel/cmd/revel
$ go get github.com/jinzhu/gorm
$ go get github.com/go-sql-driver/mysql
```

![bash](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180323105536.png)

现在，我们已经把包安装在了 `~/go/bin`。

## 创建一个应用

接下来，我们来创建你的 Go/Revel 应用。

你可以用 `revel new` 命令来创建 Go/Revel 应用。

在 [PaizaCloud Cloud IDE](https://paiza.cloud/)，你可以在浏览器中，使用 PaizaCloud 的应用程序 `Terminal` 来运行命令。

在页面的左边，点击 `Terminal` 按钮。

![terminal](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20171214154805.png)

现在我们启动了 `Terminal` 程序。我们现在要在终端输入 `revel new [application name]` 命令。

`[application name]` 是你创建的程序名称。可以用你喜欢的名称，如 `music-app` 或者 `game-app`。

在这里，我把程序命名为 `myapp`，用来管理任务清单。

输入：

```bash
$ revel new myapp
```

![revel new myapp](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180323111349.png)

在页面左边的文件管理器视图中，你可以看到 `go/src/myapp` 目录。点击文件夹并打开，看看里面的内容。

![file manager view](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180323110203.png)

你会看到里面有很多 Go/Revel 文件。

## 开启 Revel 服务器

现在，你可以运行程序了。我们来启动这个程序。

输入 `cd ~/go` 命令切换目录后，输入 `revel run myapp` 命令，开启服务器！

```bash
$ cd ~/go
$ revel run myapp
```

![revel run myapp](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180323111505.png)

在页面左边，会出现一个新的按钮，显示文字 `9000`。

![button 9000](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180323111533.png)

Revel 服务器会在 9000 端口上运行。[PaizaCloud Cloud IDE](https://paiza.cloud/) 监测到了这个端口号（9000），自动添加了一个按钮，用于在这个端口上打开浏览器。

点击该按钮，会出现浏览器程序（PaizaClound 中的浏览器应用程序）。现在，你可以看到 Revel 的网页了，这就是你的应用！

![your web page](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180323111629.png)

（尽管 Revel 是作为 HTTP 服务器运行的，但是 PaizaCloud 会把 HTTP 转换为 HTTPS。）

## 编辑文件

你在这个应用的页面上看到的其实是一个 HTML 文件，即 `~/go/src/myapp/app/views/App/Index.html`。我们来试着编辑这个文件，修改它的标题。

在文件管理器视图上，双击 `~/go/src/myapp/app/views/App/Index.html` 文件进行编辑。

![index.html](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180323111814.png)

编辑标题部分，把 `It works` 替换为下面内容：

`go/src/myapp/app/views/App/Index.html`：

```html
	<h1>Hello Go and Revel!</h1>
```

![edit index.html](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180326143312.png)

编辑完以后，点击 `Save` 按钮或者键入 `Command-S`（或 `Ctrl-S`），保存该文件。

如果服务器没有运行，输入命令开启服务器：

```bash
$ revel run myapp
```

接下来，在页面的左边，单击有 `9000` 文字的浏览器图标。如果你已经运行了浏览器，点击刷新按钮。

![browser](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180326143503.png)

于是你可以看到刚才修改的页面内容了！

## 创建数据库

你已经有一个运行的 MySQL 服务器，因为在服务器设置的时候你就勾选了 MySQL。但是如果没有勾选的话，你还可以这样手动地开启：

```bash
$ sudo systemctl enable mysql
$ sudo systemctl start mysql
```

在 [PaizaCloud Cloud IDE](https://paiza.cloud/) 上，你可以用 root 权限安装包。

接下来，创建这个应用的数据库。在这里，我们使用 `mysql` 命令，创建一个数据库 `mydb`。输入下面命令，可以创建数据库 `mydb`。

```bash
$ mysql -u root
create database mydb;
```

![create database mydb](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180216010049.png)

于是你创建了一个数据库。

接下来，设置 `myapp` 使用该数据库。在项目目录中，数据库配置在 `conf/app.conf` 文件中。

开发模式的配置写在 `[dev]` 部分中。数据库配置的写法是 `db.info = [DB user]:[DB password]@/[DB name]?[DB options]`。

在这里，我们把 `DB user` 设为 `root`，密码为空，`DB name` 设为 `mydb`，以及添加一些选项。我们在 `[dev]` 部分中，编写 `db.info` 的配置如下所示：

`go/src/myapp/conf/app.conf`：

```conf
[dev]
db.info = root:@/mydb?charset=utf8&parseTime=True
```

接下来我们创建一个文件 `app/controllers/gorm.go`，编写代码来使用数据库。

在文件管理器视图下，右击 `go/src/myapp/app/controllers` 目录，打开快捷菜单，选择 `New File` 菜单。

![New File menu](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180326143634.png)

输入文件名 `gorm.go`，点击 `Create` 按钮，创建该文件。

![Create a file gorm.go](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180326143715.png)

打开编辑器，编写代码，如下所示：

`go/src/myapp/app/controllers/gorm.go`：

```go
package controllers

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
	"myapp/app/models"
	"log"
)

var DB *gorm.DB

func InitDB() {
	dbInfo, _ := revel.Config.String("db.info")
	db, err := gorm.Open("mysql", dbInfo)
	if err != nil {
		log.Panicf("Failed gorm.Open: %v\n", err)
	}

	db.DB()
	db.AutoMigrate(&models.Post{})
	DB = db
}
```

编辑完以后，点击 `Save` 按钮或者键入 `Command-S`（或 `Ctrl-S`），保存该文件。

我们来看看代码。

在 `InitDB()` 函数中，会读取配置文件里 `db.info` 这一行的 DB 设置，然后使用 Gorm 库打开数据库。接下来，`db.AutoMigrate()` 创建了 `Post` 模型（model）（我们会在接下来创建它）的一个表。它会把数据库句柄（database handle）赋值给 `DB` 变量，让其他文件都可以通过 `controllers.DB` 访问数据库。

然后，编辑 `app/init.go`，调用 `InitDB()` 函数。

`go/src/myapp/app/init.go`：

```go
package app

import (
  "github.com/revel/revel"
  "myapp/app/controllers"
)

...

func init() {
  ...
  revel.OnAppStart(controllers.InitDB)
}

...
```

我们添加了两行代码。第一处变化是给 `import` 添加了 `myapp/app/controllers`，第二处变化是在 `init()` 函数的最后一行添加了 `revel.OnAppStart(controllers.InitDB)`，调用了我们前面创建的 `germ.go` 中的 `InitDB()`。

## 创建表、模型等

接下来，我们创建一个数据库表。

通过 Gorm 库，我们可使用 Go 结构体编写的模型信息，来操作数据库。

在这里，我们使用 `Post` 模型，对 `post` 表进行操作，该表存储了待办清单（Todo list）的信息。

我们在 `app/models/post.go` 文件创建 `Post` 模型。

右击 `go/src/myapp/app` 目录，打开快捷菜单，选择 `New directory` 菜单，创建 `models` 目录。然后右击 `go/src/myapp/app/models` 目录，打开快捷菜单，选择 `New File` 菜单，创建 `post.go` 文件。然后编辑创建的 `app/models/post.go` 文件，如下所示。

`go/src/myapp/app/models/post.go`：

```go
package models

type Post struct {
	Id  uint64 `gorm:"primary_key" json:"id"`
	Body string `sql:"size:255" json:"body"`
}
```

我们来看看代码吧。`Post` 结构体有一个整型的字段 `Id`，以及一个字符串类型的字段 `Body`。它们表示 `posts` 数据库表的 `id` 列和 `body` 列。

退出并重启服务器。

```bash
$ cd ~/go
$ revel run myapp
```

在启动应用的时候，我们执行数据库迁移，创建了 `posts` 表。我们可以使用 `phpMyAdmin` 看到表的数据。

在 PaizaCloud 中的浏览器上，在 URL 区域上输入 `http://localhost/phpmyadmin/`。

![phpmyadmin](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180323132256.png)

我们可以看到 `posts` 表。你可以在这里添加、编辑和删除数据记录。

## 路由设置

这个待办清单应用共有三种动作：列出待办事项、添加一条待办事项或者删除一条待办事项。我们分别设置了 3 个路由，如下所示。

Method | Path | Action
---|---|---
GET | /posts | List tasks
POST | /posts | Add a task
DELETE | /posts/{id} | Delete a task

我们在 `routes/web.php` 文件上设置这些路由。首先，移除该文件里面所有的默认路由。对于首页 `/`，设置成重定向到 `/tasks`（译注：`/posts`，原文笔误）。

如下所示，编辑 `conf/routes`。

`go/src/myapp/conf/routes`：

```conf
GET     /                                       Post.RedirectToPosts
GET     /posts                                  Post.Index
POST    /posts                                  Post.Create
POST    /posts/:id/delete                       Post.Delete
```

我们来看看代码吧。`GET /` 请求调用了 `Post` 控制器（controller）的 `RedirectToPosts` 方法。而 `GET /posts`、`POST /posts`、`POST /posts/:id/delete` 请求调用了 `Post` 控制器的 `Index`、`Create`、`Delete` 方法。控制器的方法可以把 `:id` 作为参数使用。

## 控制器设置

创建一个 `Post` 控制器，路由通过 `app/controllers/post.go` 来引用它。

右击 `app/controllers` 目录，打开快捷菜单，选择 `New file`，创建 `post.go` 文件。

定义 `Index()`、`Create()`、`Delete()` 方法，分别表示列出、添加和删除待办事项。

`go/src/myapp/app/controllers/post.go`：

```go
package controllers

import (
	"github.com/revel/revel"
	"myapp/app/models"
	"errors"
)

type Post struct {
	*revel.Controller
}

func (c Post) Index() revel.Result {
	posts := []models.Post{}

	result := DB.Order("id desc").Find(&posts);
	err := result.Error
	if err != nil {
		return c.RenderError(errors.New("Record Not Found"))
	}
	return c.Render(posts)
}
func (c Post) Create() revel.Result {
	post := models.Post{
		Body: c.Params.Form.Get("body"),
	}
	ret := DB.Create(&post)
	if ret.Error != nil {
		return c.RenderError(errors.New("Record Create failure." + ret.Error.Error()))
	}
	return c.Redirect("/posts")
}
func (c Post) Delete() revel.Result {
	id := c.Params.Route.Get("id")
	posts := []models.Post{}
	ret := DB.Delete(&posts, id)
	if ret.Error != nil {
		return c.RenderError(errors.New("Record Delete failure." + ret.Error.Error()))
	}
	return c.Redirect("/posts")
}

func (c Post) RedirectToPosts() revel.Result {
	return c.Redirect("/posts")
}
```

我们来研究一下代码。

首先我们把 `myapp/app/models` 添加到了 `import` 当中，因此我们可以访问到 `Post` 模型。接着在 `type Post struct` 创建了 `Post` 控制器。

**`func (c Post) Index() revel.Result`** 是 `Post` 控制器的 `Index()` 方法，它会返回待办清单。`posts := []models.Post{}` 会创建 `posts`，即 `Post` 模型的数组。`DB.Order("id desc").Find(&posts)` 提取了所有的 `posts` 表记录，将其存储在 `posts` 数组中。接下来，如果没有错误发生，就会调用 `Render()` 方法，创建 HTML 文件。我们后面会创建 HTML 模板，HTML 就是通过它来创建的。通过把 `Render()` 参数设置为 `posts`，模板文件就可以引用 `posts` 表了。

**`func (c Post) Create() revel.Result`** 是 `Post` 控制器的 `Create()` 方法，它会创建一个待办事项。`models.Post{...}` 会创建一个模型，使用 `c.Params.Form.Get()`，可以获取提交表单的 `body` 参数，并将其赋值给模型的 `Body` 字段。`DB.Create(&post)` 通过模型创建了一个数据库记录。然后，使用 `c.Redirect("/posts")` 重定向到待办清单页面。

**`func (c Post) Delete() revel.Result`** 是 `Post` 控制器的 `Delete()` 方法，它会删除一个待办事项。`c.Params.Route.Get("id")` 可以获取到 URL `/posts/:id/delete` 的 `:id` 部分。`DB.Delete(&posts, id)` 会删除所指定 `id` 的记录，然后 `c.Redirect("/posts")` 会重定向到待办清单的页面。

**`func (c Post) RedirectToPosts() revel.Result`** 用于从首页重定向到待办清单的页面上。

## 创建 HTML 模板

接下来，我们给 HTML 创建模板。HTML 模板是嵌入代码的 HTML 文件。

创建一个 HTML 模板文件 `app/views/Post/index.html`，它用于列出、添加和删除待办事项。

右击 `go/myapp/app/views` 目录，打开快捷菜单，选择 `New Directory` 菜单，创建 `Posts` 目录。右击 `go/myapp/app/views/Post` 目录，打开快捷菜单，选择 `New File` 目录，创建 `index.html` 文件。

如下所示，编辑创建好的 `app/views/Post/index.html` 文件。

`go/myapp/app/views/Post/index.html`：

```html
{{set . "title" "Todo list"}}
{{template "header.html" .}}

<header class="jumbotron" style="background-color:#A9F16C">
  <div class="container">
	<div class="row">
	  <h1>Todo list</h1>
	  <p></p>
	</div>
  </div>
</header>

<div class="container">
  <div class="row">
	<div class="span6">
	  {{template "flash.html" .}}
	</div>
  </div>
</div>

<div class="container">
	<form action="/posts" method="post">
		<div class="form-group">
			<div class="row">
				<label for="todo" class="col-xs-2">Todo</label>
				<input type="text" name="body" class="col-xs-8">
				<div class="col-xs-2">
					<button type="submit" class="btn btn-success">
						<i class="fa fa-plus"></i> Add Todo
					</button>
				</div>
			</div>
		</div>
	</form>

	<h2>Current Todos</h2>
	<table class="table table-striped todo-table">
		<thead>
			<th>Todos</th><th>&nbsp;</th>
		</thead>

		<tbody>
			{{ range .posts }}
				<tr>
					<td>
						<div>{{ .Body }}</div>
					</td>
					<td>
						<form action="/posts/{{.Id}}/delete" method="post">
							<button class="btn btn-danger">Delete</button>
						</form>
					</td>
				</tr>
			{{ end }}
		</tbody>
	</table>

</div>


{{template "footer.html" .}}
```

我们来看看这个模板文件。在 HTML 模板里，`{{` 和 `}}` 所包含的部分，用于描述创建 HTML 的操作。

`{{set . "title" "Todo list"}}` 会将 `title` 变量设为 `Todo list`。

`{{template "header.html" .}}` 通过模板文件 `header.html`，创建 HTML 文件。通过像这样调用其它的 HTML 模板，我们可以在一个模板文件中，共享多个模板文件的公共部分。在这里 `header.html` 有公共的 HTML 头部。

`＜form action="/posts" method="post"＞` 用于创建一个待办事项的表单。

`＜input type="text" name="body" class="col-xs-8"＞` 显示一个文本输入表单，用于输入待办事项。表单名设置成了 `body`，于是在所提交请求的参数 `body` 中，含有输入的待办事项。

`posts` 数组会读取 `{{ range .posts }}` 和 `{{ end }}` 之间的部分，并通过 HTML 模板，为每个 `post` 重复地创建 HTML。

`{{ .Body }}` 显示了每个 `post` 的 `Body` 字段。 `{{.ID}}` 显示每个 `post` 的 `Id` 字段。

最后，`{{template "footer.html" .}}` 通过 `footer.html` 模板显示了 HTTP 页脚。

## 运行应用程序

现在，我们已经编写完所有的代码。我们来看看吧。

在 PaizaCloud 中，单击浏览器图标（9000），打开浏览器。

我们可以看到 `Task List` 网页，当前还没有任务。

我们来添加和删除任务。

![Todo list](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-revel/20180326144011.png)

它奏效了！我们通过 Go/Revel，成功地创建了任务清单的应用程序！

需要指出的是，在 `PaizaCloud` 的免费方案（free plan）中，服务器将会被暂停。若要不间断地运行程序，请升级到基本方案（BASIC plan）。

## 总结

通过 PaizaCloud Cloud IDE，我们在浏览器上创建了一个 Go/Revel 应用，无需安装和设置任何开发环境。我们甚至可以直接在 PaizaCloud 上发布应用。现在，开始构建你自己的 Go/Revel 应用吧！

通过 [PaizaCloud Cloud IDE](https://paiza.cloud/)，只需在浏览器上，你就能灵活、轻松地开发和发布 web 应用或服务器应用。

---

via: http://engineering.paiza.io/entry/paizacloud_golang_revel

作者：[Tsuneo](http://twitter.com/yoshiokatsuneo)
译者：[Noluye](https://github.com/Noluye)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
