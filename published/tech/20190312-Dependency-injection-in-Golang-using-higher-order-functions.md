首发于：https://studygolang.com/articles/19578

# Golang 中的依赖注入之使用更高阶的函数

你可以找到一个完整的代码示例在[github.com/steinfletcher/func-dependency-injection-go](https://github.com/steinfletcher/func-dependency-injection-go)。例子包含了一个暴露 REST 接口的 http 服务器。

## 简介

在这篇博文我们介绍一种 Go 中依赖注入的方式 -- 使用更高阶的函数和闭包。

考虑下以下返回用户资料的 domain 层函数。

```go

func GetUserProfile(id string) UserProfile {
    rows, err := db.Query("SELECT ...")
    ...
    return profileText
}
```

我们想要将操作用户数据和接入数据库的代码分离开。在这个例子中，我们想要对 domain 层和任意的业务逻辑进行单元测试，同时为数据库接入函数提供 mock。让我们把这些关系分离，使得每个函数拥有单一职责。

```go

// 包含任意业务逻辑或者映射代码的 domain 层函数
func GetUserProfile(id string) User {
    ...
}

// 数据库接入层函数
func SelectUserByID(id string) UserProfile {
    ...
}
```

我们也可以在其他的 domain 函数复用 `SelectUserByID`。我们需要一种方式把 `SelectUserByID` 注入到 `GetUserProfile` 中，从而可以在测试中对 `GetUserProfile` 进行单元测试以及数据接入层提供 mock。`go` 中能做到这种效果的一种方式是对函数定义使用类型别名。

## 类型别名

让 `GetUserProfile` 依赖于一个抽象实现，这意味着我们可以在测试里注入一个数据接入层的 mock。`go` 里两种通常的做法是使用接口或者类型别名。类型别名很简便，不需要生成 struct，所以我们在这就用它吧。我们会为两个函数都定义类型别名。

```go
type SelectUserByID func(id string) User

type GetUserProfile func(id string) UserProfile

func NewGetUserProfile(selectUser SelectUserByID) GetUserProfile {
    return func(id string) string {
        user := selectUser(id)
        return user.ProfileText
    }
}

func selectUser(id string) User {
    ...
    return User{ProfileText: userRow.ProfileText}
}

```

`SelectUserByID` 是提供用户 ID 返回一个用户的函数，我们不定义它的实现。`NewGetUserProfile` 是一个从参数 `selectUser` 中得到依赖的工厂方法，然后返回一个能被调用者使用的函数。这个策略使用闭包让内部函数访问到外层函数的依赖。闭包在变量和常量定义的地方捕获它们的上下文，这被称为 ` 关闭 ` 那些变量和常量。( 译者注：想表达的意思应该是，把变量和常量的上下文给 ` 关闭 ` 起来，因而称为闭包。一个形象化的比喻 )

我们可以像这样调用 domain 函数。

```go
// 应用中某一处的连接依赖项
getUser := NewGetUserProfile(selectUser)

user := getUser("1234")
```

## 另一种看法

如果你对类似于 Java 这类的语言比较熟悉的话，这类似于创建了一个类，注入类依赖到构造器，然后在某个方法中访问依赖。其实和这个途径并没有功能性的区别，你可以认为函数的类型别名是一个简单抽象方法 (SAM) 的接口。在 Java 里我们可能会使用构造器注入依赖。

```go
interface DB {
    User SelectUser(String id)
}

public class UserService {
    private final DB db;

    public UserService(DB db) { // 注入依赖到构造器中
        this.DB = db;
    }

    public UserProfile getUserProfile(String id) { // 访问 ( 依赖 ) 的方法
        User user = this.DB.SelectUser(id);
        ...
        return userProfile;
    }
}
```

而 `go` 使用更高阶的函数也有等同的效果

```go
type SelectUser func(id string) User

type GetUserProfile func(id string) UserProfile

func NewGetUserProfile(selectUser SelectUser) { // 注入依赖的工厂方法
    return func(id string) UserProfile { // 访问 ( 依赖 ) 的方法
        user := selectUser(id)
        ...
        return userProfile
    }
}
```

## 测试

现在可以对我们的 domain 层功能进行单元测试以及为数据库接入层提供 mock。

```go
func TestGetUserProfile(t *testing.T) {
    selectUserMock := func(id string) User {
        return User{name: "jan"}
    }
    getUser := NewGetUserProfile(selectUserMock)

    user := getUser("12345")

    assert.Equal(t, UserProfile{ID: "12345", Name: "jan"}, user)
}
```

你可以找到一个完整的代码示例在[github.com/steinfletcher/func-dependency-injection-go](https://github.com/steinfletcher/func-dependency-injection-go)。例子包含了一个暴露 REST 接口的 http 服务器。

---

via: https://stein.wtf/posts/2019-03-12/inject/

作者：[Stein Fletcher](https://github.com/steinfletcher)
译者：[LSivan](https://github.com/LSivan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
