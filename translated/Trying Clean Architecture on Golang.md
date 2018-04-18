# 在Golang中尝试简洁架构
>（独立性，可测试性的和简洁性）

在阅读了 Bob 叔叔的 Clean Architecture Concept之后，我尝试在 Golang 中实现它。我们公司也有使用相似的架构，[Kurio - App Berita Indonesia](https://kurio.co.id/)， 但是结构有点不同。并不是太不同， 相同的概念，但是文件夹结构不同。

你可以在这里找到一个示例项目[https://github.com/bxcodec/go-clean-arch](https://github.com/bxcodec/go-clean-arch)，这是一个CRUD管理示例文章
![](https://cdn-images-1.medium.com/max/1600/1*CyteJRpIHC-DFE23UtlZfQ.png)

* 免责声明：

  我不推荐这里使用的任何库或框架，你可以使用你自己的或者第三方具有相同功能的任何框架来替换。

## 基础

在设计简洁架构之前我们需要了解如下约束：

1. 独立于框架。该架构不会依赖于某些功能强大的软件库存在。这可以让你使用这样的框架作为工具，而不是让你的系统陷入到框架的限制的约束中。

2. 可测试性。业务规则可以在没有 UI， 数据库，Web 服务或其他外部元素的情况下进行测试。

3. 独立于 UI 。在无需改变系统的其他部分情况下， UI 可以轻松的改变。例如，在没有改变业务规则的情况下，Web UI 可以替换为控制台 UI。

4. 独立于数据库。你可以用 Mongo， BigTable， CouchDB 或者其他数据库来替换 Oracle 或 SQL Server，你的业务规则不要绑定到数据库。
5. 独立于外部媒介。 实际上，你的业务规则可以简单到根本不去了解外部世界。

更多详见：[ https://8thlight.com/blog/uncle-bob/2012/08/13/the-clean-architecture.html]( https://8thlight.com/blog/uncle-bob/2012/08/13/the-clean-architecture.html)

所以， 基于这个约束，每一层都必须是独立的和可测试的。

如Bob叔叔的架构有4层：

* 实体层（Entities）
* 用例层（Usecase）
* 控制层（Controller）
* 框架和驱动层（Framework & Driver）

在我的项目里，我也使用了4层架构：

* 模型层（Models）
* 仓库层（Repository)
* 用例层 (Usecase)
* 表现层（Delivery）

## 模型（Models）

与实体（Entities）一样， 模型会在每一层中使用，在这一层中将存储对象的结构和它的方法。例如： Article， Student， Book。

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

所以实体或者模型将会被存放在这里

## 仓库（Repository）

仓库将存放所有的数据库处理器，查询，创建或插入数据库的处理器将存放在这一层，该层仅对数据库执行 CRUD 操作。 该层没有业务流程。只有操作数据库的普通函数。

这层也负责选择应用中将要使用什么样的数据库。 可以是 Mysql， MongoDB， MariaDB，Postgresql，无论使用哪种数据库，都要在这层决定。

如果使用 ORM， 这层将控制输入，并与 ORM 服务对接。

如果调用微服务， 也将在这层进行处理。创建 HTTP 请求去请求其他服务并清理数据，这层必须完全充当仓库。 处理所有的数据输入，输出，并且没有特定的逻辑交互。

该仓库层（Repository）将依赖于连接 DB 或其他微服务（如果存在的话）

## 用例（Usecase）

这层将会扮演业务流程处理器的角色。任何流程都将在这里处理。该层将决定哪个仓库层被使用。并且负责提供数据给服务以便交付。处理数据进行计算或者在这里完成任何事。

用例层将接收来自传递层的所有经过处理的输入，然后将处理的输入存储到数据库中， 或者从数据库中获取数据等。

用例层将依赖于仓库层。

## 表现（Delivery）

这一层将作为表现者。决定数据如何呈现。任何交付类型都可以作为是 REST API， 或者是 HTML 文件，或者是 gRPC

这一层将接收来自用户的输入， 并清理数据然后传递给用例层。

对于我的示例项目， 我使用 REST API 作为表现方式。客户端将通过网络调用资源节点， 表现层将获取到输入或请求，然后将它传递给用例层。

该层依赖于用例层。

## 层与层之间的通信

除了模型层， 每一层都需要通过接口进行通信。例如，用例层需要仓库层，那么他们该如何通信呢？仓库层将提供一个接口作为他们沟通桥梁。

仓库层接口示例：

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
用例层将通过这个接口与仓库层进行通信，仓库层必须实现这个接口，以便用例层使用该接口。

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
