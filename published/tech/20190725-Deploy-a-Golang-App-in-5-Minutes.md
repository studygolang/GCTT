首发于：https://studygolang.com/articles/25297

# 在 5 分钟之内部署一个 Go 应用

在有些程序人写完了他们的 Go 应用之后，这总会成为一个大问题——“我刚写的这个 Go 应用，当它崩溃的时候我要怎么重启？”，因为你没法用 `go run main.go` 或者 `./main` 这样的命令让它持续运行，并且当程序崩溃的时候能够重启。

一个普通使用的好办法是使用 Docker。但是，设置 Docker 以及为容器配置你的应用需要花费时间，当你的程序需要和 MySQL、Redis 这样的服务器/进程交互时更是如此。对于一个大型或长期项目来说，毋庸置疑这是一个正确的选择。但是如果在你手上的是个小应用，你想要快速部署并且实时地服务器上查看状态，那么你可能需要考虑别的选择。

另一个选择就是在你的 Linux 服务器上创建一个守护进程，然后让它作为一个服务运行，但是这需要花费一些额外的工夫。而且，如果你并不具备 Linux 系统和服务相关的知识的话，这就不是一件简单的事情了。所以，这里有一个最简单的解决方案——使用 [Supervisor](http://supervisord.org/) 来部署你的 Go 应用，然后它会为你处理好其余的工作。它是一个能够帮你监控你的应用程序并在其崩溃时进行重启的工具。

## 安装

安装 Supervisor 相当简单，在 Ubuntu 上这条命令就会在你的系统上安装 Supervisor。

```bash
sudo apt install supervisor
```

然后你需要将 Supervisor 添加到系统的用户组中：

```bash
sudo addgroup --system supervisor
```

现在，在创建 Supervisor 的配置文件之前，我们先写一个简单的 Go 程序。这个程序将会读取 .env 文件中的配置项，然后和 MySQL 数据库进行交互。代码如下：

（为了方便演示，我们会让代码简单些）

```go
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var db *sql.DB

func init() {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Println("Error readin .env: ", err)
		os.Exit(1)
	}

	dbUserName := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbNAME := os.Getenv("DB_NAME")

	dsn := dbUserName + ":" + dbPassword + "@/" + dbNAME

	db, err = sql.Open("mysql", dsn)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

}

func main() {

	r := mux.NewRouter()

	r.Use(middleware)

	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/user", createUserHandler).Methods("POST")

	fmt.Println("Listening on :8070")
	if err := http.ListenAndServe(":8070", r); err != nil {
		// 退出程序
		log.Println("Failed starting server ", err)
		os.Exit(1)
	}

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is root handler")
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {

	user := &User{}

	err := json.NewDecoder(r.Body).Decode(user)

    // 对请求响应 JSON 数据
    // 在实际应用中你可能想要创建一个进行错误处理的函数
	if err != nil {
		// 我们也可以这么做
		// errREsp := `"error": "Invalid input", "status": 400`
		// w.Header().Set("Content-Type", "application/json")
		// w.WriteHeader(400)
        // w.Write([]byte(errREsp))

        // 然而我们会让服务器崩溃
		log.Fatal(err)

		return
	}

    // 在实际应用中必须对密码进行哈希，可以使用 bcrypt 算法
	_, err = db.Exec("INSERT INTO users(email, password) VALUES(?,?)", user.Email, user.Password)

	if err != nil {
		log.Println(err)
        // 简单起见，发送明文字符串
        // 创建一个有效的 JSON 响应
		errREsp := `"error": "Internal error", "status": 500` // 返回 500 状态码，因为这是我们而非用户的问题
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(errREsp))

		return
	}

}

// 一个简单的中间件，只用来记录请求的 URI
var middleware = func(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestPath := r.URL.Path
		log.Println(requestPath)
		next.ServeHTTP(w, r) // 在中间件调用链中进行处理！

	})
}
```

现在，如果我们想要用 Supervisor 来运行这个程序，我们需要构建程序的二进制文件。同时在项目的根目录下创建一个 `.env` 文件 —— 如果你想把配置文件和项目放在一起的话，在这个文件中写上 MySQL 数据库需要的变量。

将这个仓库克隆到你想要运行的服务器上。确保你遵循了 Go 目录路径的惯例：

```bash
$ go build .
```

Go 的这个命令最终会创建一个以项目根目录命名的二进制文件，所以如果项目的根目录是 `myapp`，那么文件的名称就是 `myapp`。

现在，在服务器上创建 Supervisor 的配置文件 `/etc/supervisor/conf.d`。

```bash
#/etc/supervisor/conf.d/myapp.conf

[program:myapp]
directory=/root/gocode/src/github.com/monirz/myapp
command=/root/gocode/src/github.com/monirz/myapp/myapp
autostart=true
autorestart=true
stderr_logfile=/var/log/myapp.err
stdout_logfile=/var/log/myapp.log
environment=CODENATION_ENV=prod
environment=GOPATH="/root/gocode"
```

这里的 directory 和 command 变量很重要。directory 变量应该设置为项目的根目录，因为程序将会尝试在 directory 指定的路径下读取 .env 文件或是其他需要的配置文件。`autorestart` 变量设置为 `true`，这样当程序崩溃时就会重启。

现在通过下面的命令重新加载 Supervisor：

```bash
$ sudo supervisorctl reload
```

来检查下它的状态。

```bash
$ sudo supervisorctl status
```

一切都正确配置的话，你应该会看到类似下面的输出内容：

```bash
myapp     RUNNING   pid 2023, uptime 0:00:03
```

我们名为 myapp 的 Go 服务端程序正在后台运行。

现在向我们刚写的 API 发起一些请求。首先检查 `rootHandler` 是否正在工作。然后向 `/user` 结点发送一个包含无效 JSON 格式数据的请求。这应当会让服务器崩溃。但是服务器上没有存储任何日志，不是吗？因为我们还没有实现日志功能？

等等，Supervisor 实际上已经为我们处理了日志。如果你到 `/var/log` 目录下查看 myapp.log 文件，你就会看到它记录着已经向服务器发起过的请求的 URI 路径。

```bash
$ cat /var/log/myapp.log
```

错误日志也是如此。好了，我们的服务器程序已经运行了——崩溃的话会重启，还会记录每个请求和错误信息。我觉得我们应该是在大约 5 分钟以内做完了这些事吧？（大概是吧，谁在乎呢。）但关键是，用 Supervisor 来部署和监控你的 Go 应用程序时十分简单的。

你觉得呢？毫不犹豫地回复我吧。周末愉快。

---

via: https://medium.com/@monirz/deploy-golang-app-in-5-minutes-ff354954fa8e

作者：[Monir Zaman](https://medium.com/@monirz)
译者：[maxwellhertz](https://github.com/maxwellhertz)
校对：[polaris1119](https://github.com/polaris1119)
