已发布：https://studygolang.com/articles/12909

# 在 Golang 中尝试简洁架构
>（独立性，可测试性的和简洁性）

在阅读了 Bob 叔叔的 Clean Architecture Concept 之后，我尝试在 Golang 中实现它。我们公司也有使用相似的架构，[Kurio - App Berita Indonesia](https://kurio.co.id/)， 但是结构有点不同。并不是太不同， 相同的概念，但是文件目录结构不同。

你可以在这里找到一个示例项目[https://github.com/bxcodec/go-clean-arch](https://github.com/bxcodec/go-clean-arch)，这是一个 CRUD 管理示例文章
![](https://raw.githubusercontent.com/studygolang/gctt-images/master/clean-arthitecture/1_CyteJRpIHC-DFE23UtlZfQ.png)

* 免责声明：

  我不推荐使用这里的任何库或框架，你可以使用你自己的或者第三方具有相同功能的任何框架来替换。

## 基础

在设计简洁架构之前我们需要了解如下约束：

1. 独立于框架。该架构不会依赖于某些功能强大的软件库存在。这可以让你使用这样的框架作为工具，而不是让你的系统陷入到框架的限制的约束中。

2. 可测试性。业务规则可以在没有 UI， 数据库，Web 服务或其他外部元素的情况下进行测试。

3. 独立于 UI 。在无需改变系统的其他部分情况下， UI 可以轻松的改变。例如，在没有改变业务规则的情况下，Web UI 可以替换为控制台 UI。

4. 独立于数据库。你可以用 Mongo， BigTable， CouchDB 或者其他数据库来替换 Oracle 或 SQL Server，你的业务规则不要绑定到数据库。
5. 独立于外部媒介。 实际上，你的业务规则可以简单到根本不去了解外部世界。

更多详见：[ https://8thlight.com/blog/uncle-bob/2012/08/13/the-clean-architecture.html]( https://8thlight.com/blog/uncle-bob/2012/08/13/the-clean-architecture.html)

所以， 基于这些约束，每一层都必须是独立的和可测试的。

如 Bob 叔叔的架构有 4 层：

* 实体层（ Entities ）
* 用例层（ Usecase ）
* 控制层（ Controller ）
* 框架和驱动层（ Framework & Driver ）

在我的项目里，我也使用了 4 层架构：

* 模型层（ Models ）
* 仓库层（ Repository )
* 用例层 ( Usecase )
* 表现层（ Delivery ）

## 模型层（ Models ）

与实体（ Entities ）一样， 模型会在每一层中使用，在这一层中将存储对象的结构和它的方法。例如： Article， Student， Book。

```go
import "time"

type Article struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
```

所以实体或者模型将会被存放在这一层

## 仓库层（ Repository  ）

仓库将存放所有的数据库处理器，查询，创建或插入数据库的处理器将存放在这一层，该层仅对数据库执行 CRUD 操作。 该层没有业务流程。只有操作数据库的普通函数。

这层也负责选择应用中将要使用什么样的数据库。 可以是 Mysql， MongoDB， MariaDB，Postgresql，无论使用哪种数据库，都要在这层决定。

如果使用 ORM， 这层将控制输入，并与 ORM 服务对接。

如果调用微服务， 也将在这层进行处理。创建 HTTP 请求去请求其他服务并清理数据，这层必须完全充当仓库。 处理所有的数据输入，输出，并且没有特定的逻辑交互。

该仓库层（ Repository  ）将依赖于连接数据库 或其他微服务（如果存在的话）

## 用例层（ Usecase ）

这层将会扮演业务流程处理器的角色。任何流程都将在这里处理。该层将决定哪个仓库层被使用。并且负责提供数据给服务以便交付。处理数据进行计算或者在这里完成任何事。

用例层将接收来自传递层的所有经过处理的输入，然后将处理的输入存储到数据库中， 或者从数据库中获取数据等。

用例层将依赖于仓库层。

## 表现层（ Delivery  ）

这一层将作为表现者。决定数据如何呈现。任何传递类型都可以作为是 REST API， 或者是 HTML 文件，或者是 gRPC

这一层将接收来自用户的输入， 并清理数据然后传递给用例层。

对于我的示例项目， 我使用 REST API 作为表现方式。客户端将通过网络调用资源节点， 表现层将获取到输入或请求，然后将它传递给用例层。

该层依赖于用例层。

## 层与层之间的通信

除了模型层， 每一层都需要通过接口进行通信。例如，用例（ Usecase ）层需要仓库（ Repository ）层，那么它们该如何通信呢？仓库（ Repository ）层将提供一个接口作为他们沟通桥梁。

仓库层（ Repository ）接口示例：

```go

package repository

import models "github.com/bxcodec/go-clean-arch/article"

type ArticleRepository interface {
	Fetch(cursor string, num int64) ([]*models.Article, error)
	GetByID(id int64) (*models.Article, error)
	GetByTitle(title string) (*models.Article, error)
	Update(article *models.Article) (*models.Article, error)
	Store(a *models.Article) (int64, error)
	Delete(id int64) (bool, error)
}
```
用例层（ Usecase ）将通过这个接口与仓库层进行通信，仓库层（ Repository ）必须实现这个接口，以便用例层（ Usecase ）使用该接口。

用例层接口示例：

```go
package usecase

import (
	"github.com/bxcodec/go-clean-arch/article"
)

type ArticleUsecase interface {
	Fetch(cursor string, num int64) ([]*article.Article, string, error)
	GetByID(id int64) (*article.Article, error)
	Update(ar *article.Article) (*article.Article, error)
	GetByTitle(title string) (*article.Article, error)
	Store(*article.Article) (*article.Article, error)
	Delete(id int64) (bool, error)
}
```

与用例层相同， 表现层将会使用这个约定接口。 并且用例层必须实现该接口。

## 测试

我们知道， 简洁就意味着独立。 甚至在其他层还不存在的情况下，每一层都具有可测试性。

* 模型（ Models ）层

  该层仅测试任意结构声明的函数或方法。
  这可以独立于其他层，轻松的进行测试。

* 仓库（ Repository ）层

  为了测试该层，更好的方式是进行集成测试，但你也可以为每一个测试进行模拟测试， 我使用 github.com/DATA-DOG/go-sqlmock 作为我的工具来模拟查询过程 mysql

* 用例（ Usecase ）层

  因为该层依赖于仓库层， 意味着该层需要仓库层来支持测试。所以我们根据之前定义的契约接口制作一个模拟的仓库（ Repository ）模型。

* 表现（ Delivery ）层

  与用例层相同，因为该层依赖于用例层，意味着该层需要用例层来支持测试。基于之前定义的契约接口， 也需要对用例层进行模拟。

对于模拟，我使用 vektra 的 golang的模拟库：
[https://github.com/vektra/mockery](https://github.com/vektra/mockery)

## 仓库层(Repository)测试

为了测试这层，就如我之前所说， 我使用 sql-mock 来模拟我的查询过程。

你可以像我一样使用 github.com/DATA-DOG/go-sqlmock ，或者使用其他具有相似功能的库。

```go
func TestGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()
	rows := sqlmock.NewRows([]string{
		"id", "title", "content", "updated_at", "created_at"}).
		AddRow(1, "title 1", "Content 1", time.Now(), time.Now())

	query := "SELECT id,title,content,updated_at, created_at FROM article WHERE ID = ?"

	mock.ExpectQuery(query).WillReturnRows(rows)

	a := articleRepo.NewMysqlArticleRepository(db)
	num := int64(1)

	anArticle, err := a.GetByID(num)

	assert.NoError(t, err)
	assert.NotNil(t, anArticle)
}
```

## 用例层（Usecase）测试

用于用例层的示例测试，依赖于仓库层。

```go
package usecase_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/bxcodec/faker"
	models "github.com/bxcodec/go-clean-arch/article"
	"github.com/bxcodec/go-clean-arch/article/repository/mocks"
	ucase "github.com/bxcodec/go-clean-arch/article/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetch(t *testing.T) {
	mockArticleRepo := new(mocks.ArticleRepository)
	var mockArticle models.Article
	err := faker.FakeData(&mockArticle)
	assert.NoError(t, err)

	mockListArtilce := make([]*models.Article, 0)
	mockListArtilce = append(mockListArtilce, &mockArticle)
	mockArticleRepo.On("Fetch", mock.AnythingOfType("string"), mock.AnythingOfType("int64")).Return(mockListArtilce, nil)
	u := ucase.NewArticleUsecase(mockArticleRepo)
	num := int64(1)
	cursor := "12"
	list, nextCursor, err := u.Fetch(cursor, num)
	cursorExpected := strconv.Itoa(int(mockArticle.ID))
	assert.Equal(t, cursorExpected, nextCursor)
	assert.NotEmpty(t, nextCursor)
	assert.NoError(t, err)
	assert.Len(t, list, len(mockListArtilce))

	mockArticleRepo.AssertCalled(t, "Fetch", mock.AnythingOfType("string"), mock.AnythingOfType("int64"))

}
```
Mockery 将会为我生成一个仓库层模型，我不需要先完成仓库（Repository）层， 我可以先完成用例（Usecase），即使我的仓库（Repository）层尚未实现。

## 表现层（ Delivery ）测试

表现层测试依赖于你如何传递的数据。如果使用 http REST API， 我们可以使用 golang 中的内置包 httptest。

因为该层依赖于用例( Usecase )层, 所以 我们需要模拟 Usecase，与仓库层相同，我使用 Mockery 模拟我的 Usecase 来进行表现层（ Delivery ）的测试。

```go
func TestGetByID(t *testing.T) {
	var mockArticle models.Article
	err := faker.FakeData(&mockArticle)
	assert.NoError(t, err)
	mockUCase := new(mocks.ArticleUsecase)
	num := int(mockArticle.ID)
	mockUCase.On("GetByID", int64(num)).Return(&mockArticle, nil)
	e := echo.New()
	req, err := http.NewRequest(echo.GET, "/article/"+
		strconv.Itoa(int(num)), strings.NewReader(""))

	assert.NoError(t, err)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("article/:id")
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(num))

	handler := articleHttp.ArticleHandler{
		AUsecase: mockUCase,
		Helper:   httpHelper.HttpHelper{},
	}
	handler.GetByID(c)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockUCase.AssertCalled(t, "GetByID", int64(num))
}
```

## 最终输出与合并

完成所有层的编码并通过测试之后。你应该在的根项目的 main.go 文件中将其合并成一个系统。

在这里你将会定义并创建每一个环境需求， 并将所有层合并在一起。

以我的 main.go 为示例：

```go
package main

import (
	"database/sql"
	"fmt"
	"net/url"

	httpDeliver "github.com/bxcodec/go-clean-arch/article/delivery/http"
	articleRepo "github.com/bxcodec/go-clean-arch/article/repository/mysql"
	articleUcase "github.com/bxcodec/go-clean-arch/article/usecase"
	cfg "github.com/bxcodec/go-clean-arch/config/env"
	"github.com/bxcodec/go-clean-arch/config/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
)

var config cfg.Config

func init() {
	config = cfg.NewViperConfig()

	if config.GetBool(`debug`) {
		fmt.Println("Service RUN on DEBUG mode")
	}

}

func main() {

	dbHost := config.GetString(`database.host`)
	dbPort := config.GetString(`database.port`)
	dbUser := config.GetString(`database.user`)
	dbPass := config.GetString(`database.pass`)
	dbName := config.GetString(`database.name`)
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Jakarta")
	dsn := fmt.Sprintf("%s?%s", connection, val.Encode())
	dbConn, err := sql.Open(`mysql`, dsn)
	if err != nil && config.GetBool("debug") {
		fmt.Println(err)
	}
	defer dbConn.Close()
	e := echo.New()
	middL := middleware.InitMiddleware()
	e.Use(middL.CORS)

	ar := articleRepo.NewMysqlArticleRepository(dbConn)
	au := articleUcase.NewArticleUsecase(ar)

	httpDeliver.NewArticleHttpHandler(e, au)

	e.Start(config.GetString("server.address"))
}
```

你可以看见，每一层都与它的依赖关系合并在一起了。

## 结论

总之，如果画在一张图上，就如下图所示：
![](https://raw.githubusercontent.com/studygolang/gctt-images/master/clean-arthitecture/1_GQdkAd7IwIwOWW-WLG5ikQ.png)

* 在这里使用的每一个库都可以由你自己修改。因为简洁架构的重点在于：你使用的库不重要， 关键是你的架构是简洁的，可测试的并且是独立的。

* 我项目就是这样组织的。通过评论和分享， 你可以讨论或者赞成，当然能改善它就更好了。

## 示例项目

示例项目可以在这里看见：[ https://github.com/bxcodec/go-clean-arch]( https://github.com/bxcodec/go-clean-arch)

我的项目中使用到的库：

* Glide ：包管理工具

* go-sqlmock from github.com/DATA-DOG/go-sqlmock

* Testify ： 测试库

* Echo Labstack （Golang Web 框架）用于 表现层

* Viper ：环境配置

进一步阅读简洁架构 ：

* [https://8thlight.com/blog/uncle-bob/2012/08/13/the-clean-architecture.html](https://8thlight.com/blog/uncle-bob/2012/08/13/the-clean-architecture.html)

* [http://manuel.kiessling.net/2012/09/28/applying-the-clean-architecture-to-go-applications/](http://manuel.kiessling.net/2012/09/28/applying-the-clean-architecture-to-go-applications/)。 这是Golang种另一个版本的简洁架构。

如果你任何问题，或者需要更多的解释，或者我在这里没有解释清楚的。你可以通过我的[LinkedIn](https://www.linkedin.com/in/imantumorang/)或者[email](iman.tumorang@gmail.com)联系我。谢谢。

---

链接：https://hackernoon.com/golang-clean-archithecture-efd6d7c43047

作者：[Iman Tumorang](https://hackernoon.com/@imantumorang)
译者：[fredvence](https://github.com/fredvence)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
