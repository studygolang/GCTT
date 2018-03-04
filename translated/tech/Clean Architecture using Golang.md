# 使用 Golang 构建整洁架构

## 什么是整洁架构 ?

在《Clean Architecture: A Craftsman’s Guide to Software Structure and Design》一书中，著名作家 Robert “Uncle Bob” Martin 提出了一种具有一些重要特性的体系结构，如框架、数据库和接口的可测试性和独立性。

整洁架构的约束条件是:
* 独立的框架。该体系结构并不依赖于某些带有特性的软件库的存在。这允许您使用这些框架作为工具，而不是将您的系统束缚在有限的约束中。
* 可测试的。业务规则可以在没有 UI、数据库、Web 服务器或任何其他外部元素的情况下进行测试。
* 独立的 UI 。UI 可以很容易地更改，而不会改变系统的其他部分。例如，可以用控制台 UI 替换 Web UI，而不需要更改业务规则。
* 独立的数据库。您可以将 Oracle 或 SQL Server 替换为 Mongo、BigTable、CouchDB 或其他数据库。您的业务规则不绑定到数据库。
* 独立的任意外部代理。事实上，你的业务规则根本就不用了解外部的构成。

了解更多请查看 : https://8thlight.com/blog/uncle-bob/2012/08/13/the-clean-architecture.html

因此，基于这些约束，每一层都必须是独立的和可测试的。

从 Uncle Bob 的架构中，我们可以将代码分成 4 层:
* 实体: 封装企业范围的业务规则。Go 中的实体是一组数据结构和函数。
* 用例: 这个层中的软件包含应用程序特定的业务规则。它封装并实现了系统的所有用例。
* 控制器: 该层中的软件是一组适配器，它将数据从最方便的用例和实体转换为最方便的外部代理，例如数据库或 Web。
* 框架和驱动程序: 这个层通常由框架和工具(如数据库、Web 框架等)组成。

## 使用 Golang 构建整洁架构

让我们以 user 包为例:

```
ls -ln pkg/user
-rw-r — r — 1 501 20 5078 Feb 16 09:58 entity.go
-rw-r — r — 1 501 20 3747 Feb 16 10:03 mongodb.go
-rw-r — r — 1 501 20 509 Feb 16 09:59 repository.go
-rw-r — r — 1 501 20 2403 Feb 16 10:30 service.go
```

在 entity.go 文件中，我们有自己的实体 :

```go
//User data
type User struct {
	ID                 entity.ID    `json:"id" bson:"_id,omitempty"`
	Picture            string       `json:"picture" bson:"picture,omitempty"`
	Email              string       `json:"email" bson:"email"`
	Password           string       `json:"password" bson:"password,omitempty"`
	Type               Type         `json:"type" bson:"type"`
	Company            []*Company   `json:"company" bson:"company,omitempty"`
	CreatedAt          time.Time    `json:"created_at" bson:"created_at"`
	ValidatedAt        time.Time    `json:"validated_at" bson:"validated_at,omitempty"`
}
```

在 repository.go 文件中我们定义存储库的接口，用于保存存储实体。在这种情况下，存储库意味着 Uncle Bob 架构中的框架和驱动层。它的内容是:

```go
package user

import "github.com/thecodenation/stamp/pkg/entity"

//Repository repository interface
type Repository interface {
	Find(id entity.ID) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByChangePasswordHash(hash string) (*User, error)
	FindByValidationHash(hash string) (*User, error)
	FindAll() ([]*User, error)
	Update(user *User) error
	Store(user *User) (entity.ID, error)
	AddCompany(id entity.ID, company *Company) error
	AddInvite(userID entity.ID, companyID entity.ID) error
}
```

该接口可以在任何类型的存储层中实现，如 MongoDB、MySQL 等。在我们的例子中，我们使用 MongoDB 来实现，就像在 mongodb.go 中看到的那样:

```go
package user

import (
	"errors"
	"os"
	"github.com/juju/mgosession"
	"github.com/thecodenation/stamp/pkg/entity"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type repo struct {
	pool *mgosession.Pool
}

//NewMongoRepository create new repository
func NewMongoRepository(p *mgosession.Pool) Repository {
	return &repo{
		pool: p,
	}
}

func (r *repo) Find(id entity.ID) (*User, error) {
	result := User{}
	session := r.pool.Session(nil)
	coll := session.DB(os.Getenv("MONGODB_DATABASE")).C("user")
	err := coll.Find(bson.M{"_id": id}).One(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *repo) FindByEmail(email string) (*User, error) {
}

func (r *repo) FindByChangePasswordHash(hash string) (*User, error) {
}

func (r *repo) FindAll() ([]*User, error) {
}

func (r *repo) Update(user *User) error {
}

func (r *repo) Store(user *User) (entity.ID, error) {
}

func (r *repo) AddCompany(id entity.ID, company *Company) error {
}

func (r *repo) AddInvite(userID entity.ID, companyID entity.ID) error {
}

func (r *repo) FindByValidationHash(hash string) (*User, error) {
}
```

service.go 文件表示 Uncle Bob 定义的用例层。在文件中，我们有 Service 接口和它的实现。Service 接口是:

```go
//Service service interface
type Service interface {
	Register(user *User) (entity.ID, error)
	ForgotPassword(user *User) error
	ChangePassword(user *User, password string) error
	Validate(user *User) error
	Auth(user *User, password string) error
	IsValid(user *User) bool
	GetRepo() Repository
}
```

最后一层，我们架构中的 Controller 是在 api 的内容中实现的:

```
cd api ; tree
.
|____handler
| |____company.go
| |____user.go
| |____address.go
| |____skill.go
| |____invite.go
| |____position.go
|____rice-box.go
|____main.go
```
在以下代码中，从 api/main.go 中我们可以看看如何使用这些服务:

```go
session, err := mgo.Dial(os.Getenv("MONGODB_HOST"))
if err != nil {
	elog.Error(err)
}
mPool := mgosession.NewPool(nil, session, 1)
queueService, err := queue.NewAWSService()
if err != nil {
		elog.Error(err)
}
userRepo := user.NewMongoRepository(mPool)
userService := user.NewService(userRepo, queueService)	
```

现在我们可以轻松地创建包测试，比如:

```go
package user

import (
	"testing"
	"time"

	"github.com/thecodenation/stamp/pkg/entity"
	"github.com/thecodenation/stamp/pkg/queue"
)

func TestIsValidUser(t *testing.T) {
	u := User{
		ID:        entity.NewID(),
		FirstName: "Bill",
		LastName:  "Gates",
	}
	userRepo := NewInmemRepository()
	queueService, _ := queue.NewInmemService()
	userService := NewService(userRepo, queueService)

	if userService.IsValid(&u) == true {
		t.Errorf("got %v want %v",
			true, false)
	}

	u.ValidatedAt = time.Now()
	if userService.IsValid(&u) == false {
		t.Errorf("got %v want %v",
			false, true)
	}
}
```

使用整洁的体系结构，我们可以将数据库从 MongoDB 更改为 Neo4j ，而不会破坏应用程序的其他部分。这样，我们可以在不损失质量和速度的情况下开发我们的软件。

## 引用

https://hackernoon.com/golang-clean-archithecture-efd6d7c43047
https://8thlight.com/blog/uncle-bob/2012/08/13/the-clean-architecture.html

----------------

via: https://medium.com/@eminetto/clean-architecture-using-golang-b63587aa5e3f

作者：[Elton Minetto](https://medium.com/@eminetto)
译者：[SergeyChang](https://github.com/SergeyChang)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
