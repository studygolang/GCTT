首发于：https://studygolang.com/articles/17952

# 用 Golang 处理数据库迁移

最近在 `r/reddit` 中不断出现 ***我如何使用 Go 来完成数据库迁移？*** 对于我和大多数人这种从其他语言例如 PHP 或是 Ruby 转到 Go 的人来说，数据库迁移在这些语言上已经不是什么问题了。例如 Ruby 的 Rails 和 PHP 的 Laravel。但我如何在 Go 中复制这种功能呢？同时考虑到框架是 Go 中的反模式这一事实。

举个例子，在在 Rails 和 Laravel 中可以非常轻松的使用 `bin/rails db:migrate` 或者 `php artisan migrate` 命令作为部署流水线的一个步骤来运行。但是同样的功能如何在 Go 应用中实现呢？

已经有许多的库被创建来解决 Go 的这一问题 , 但是目前来说 [migrate library](https://github.com/golang-migrate/migrate) 是我使用效果最好的一个库。接下来我将会构建一个只有 package main 的小程序来展示你怎么能构建任何一个 Go Web 程序在它启动的时候来进行自动数据库迁移，以及你如何能解决一些部署上的疑难杂症。接下来我将会解释它在[实际](https://lanre.wtf/blog/2019/01/02/database-migration-golang/#consider) 中如何实现。

## 一个简单应用

根据每个迁移文件来说，`migrate` 库都需要一些规则。迁移文件必须命名为 `1_create_XXX.up.sql` 和 `1 _create_xxx.down.sql`。所以基本上，每个迁移都应该有一个 `up.sql` 和一个 `down.sql` 文件。在实际运行迁移时将执行 `up.sql` 文件，而在尝试回滚时将执行 `down.sql` 文件。

> 不过你也可以使用 [migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) cli 工具来创建迁移 : `migrate create -ext sql create_users_table`

```go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	var migrationDir = flag.String("migration.files", "./migrations", "Directory where the migration files are located ?")
	var MySQLDSN = flag.String("mysql.dsn", os.Getenv("MYSQL_DSN"), "Mysql DSN")

	flag.Parse()

	db, err := sql.Open("mysql", *mysqlDSN)
	if err != nil {
		log.Fatalf("could not connect to PostgreSQL database... %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("could not ping DB... %v", err)
	}

    // Run migrations
    // 开始数据迁移
	driver, err := MySQL.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("could not start sql migration... %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", *migrationDir), // file://path/to/directory
		"mysql", driver)

	if err != nil {
		log.Fatalf("migration failed... %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("An error occurred while syncing the database.. %v", err)
	}

	log.Println("Database migrated")
	// actual logic to start your application
	os.Exit(0)
}
```

以上是在 Go 中进行数据库迁移的最简单方法。你可以继续从[这个 repo](https://github.com/adelowo/migration-demo) 下载以下文件，并将它们放在 migrations 目录中或你认为合适的任何位置。之后，你需要使用以下命令运行它 :

```
$ go run main.go -mysql.dsn "root:@tcp(localhost)/xyz"
```
如果一切顺利，你应该看到在标准输出上打印了 "Database migrated" ( 数据库迁移完成 )。

## 实际部署考虑事项

虽然这非常容易设置，但是它确实对文件系统产生了依赖性——为了使迁移成为可能，必须提供迁移文件。这也很容易解决。有三种方法可以解决这个问题：

* 如果你的应用程序在容器中运行，只需将迁移文件挂载到映像中即可。下面是一个例子：

```docker
FROM Golang:1.11 as build-env

WORKDIR /go/src/github.com/adelowo/project
ADD . /go/src/github.com/adelowo/project

ENV GO111MODULE=on

RUN go mod download
RUN go mod verify
RUN go install ./cmd

## A better scratch
FROM gcr.io/distroless/base
COPY --from=build-env /go/bin/cmd /
COPY --from=build-env /go/src/github.com/adelowo/project/path/to/migrations /migrations
CMD ["/cmd"]
```

* 如果你已经有了 CI/CD 流程，那么你可以使用 `migrate` 附带的 cli 工具。只要在实际部署过程之前包含它就可以了，当在自动化测试阶段你获取了文件的源代码——那么理想情况下，至少它们是被标识版本的。详情请参考[文档](https://github.com/golang-migrate/migrate/tree/master/cli)。

> 虽然我还没实践过，但是以上方法确实是可行的方案

最后一步实际上是正在进行的工作。但是它依赖于将迁移文件嵌入到二进制文件。通过这一步，文件系统的依赖性将被破坏。现在有一个 [open Pull-Requests](https://github.com/golang-migrate/migrate/pull/144)，并且当它的状态发生改变的时候，我将会持续留意，并更新这篇文章。

---

via: https://lanre.wtf/blog/2019/01/02/database-migration-golang/

作者：[Lanre Adelowo](https://lanre.wtf/about)
译者：[wodotatop10](https://github.com/wodotatop10)
校对：[polaris1119](https://github.com/polaris)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
