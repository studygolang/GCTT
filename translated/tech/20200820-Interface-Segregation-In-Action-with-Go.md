# 接口分离原则在Go语言中的实践

2020 年 8 月 20 日 - 标签：golang

每个人都应该写一篇关于 Golang 接口的文章！不知道我为什么等了这么久才写了这篇！

当你需要 mock 一个对象或者函数需要接受一组相关的功能从而来与对象进行交互时，Golang 的接口都可以
使这些变得更为简单。

是的！实际上接口就是被用来实现这些目的的，你或许有一个实现了很多方法的对象实例，但当你将它作为参数传递
给另外一个函数的时候，该函数可能仅仅使用了对象实例的一部分方法，为了解决这个问题，你可以通过更改函数签名的方式来解决，
你可以定义一个新的接口，函数接收实现了该接口的对象实例，该接口只包含函数所需要的功能方法。

通过上述的方法，当你对函数进行单元测试的时候，用来 mock 传参对象实例的代码也会更少，更容易处理（这种方式很好地将
不必要暴露的方法影藏了起来）。

当你将接口定义得更小，并通过组合来使用这些接口的时候，上述方式带来的好处就显得更为明显。

举个例子来说，假设你需要对一个可被增删改查的资源设计接口。这种做法在实际工作中是很有用的，因为通过这种方式，
对于那些可以被数据库持久化的资源，我们可以更好的规范它们的行为，使得对它们的操作更标准化。接下来我会具体阐述这个例子。

我在下面的例子中使用了 `interface{}`，但你在实际工作中，应该尽量避免使用，因为它实在太宽泛了。但是在例如 `Kubernetes` 的实现中使用了
[runtime.Object](https://godoc.org/k8s.io/apimachinery/pkg/runtime)，实际上是一种更好的选择。即将在Go 2.0 版本中引入
的泛型支持会使得类似场景中的实现更简单。或者你也可以用代码生成来实现。但总而言之，`Kubernetes` 中使用可序列化的对象这一思想是非常优秀的。

```golang
type Resource interface {
    Create(ctx context.Context) error
    Update(ctx context.Context, updated interface{}) error
    Delete(ctx context.Context) error
}
```

上述接口中定义的方法并不多，可以满足场景需求，但我并不是很喜欢接口的命名。我并不能通过接口名清楚地了解其意图。该接口定义了一种资源，但是
我在命名接口的时候通常更喜欢选用动词或形容词。在我们所描述的场景中，实现该接口的是一种可以被数据库持久化的资源。因此我认为更确切的接口名称应该是：
["Persistable"](https://en.wiktionary.org/wiki/persistable)，因为它使接口的意图更为明显。

我们根据动作将接口进行拆分：

```golang
type Creatable interface {
    Create(ctx context.Context) error
}

type Updatable interface {
    Update(ctx context.Context, updated interface{}) error
}

type Deletable interface {
    Delete(ctx context.Context) error
}
```

如果需要的话，你可以借助 Go 语言的特性，利用组合的方式将上述三个接口组合成一个新的接口：

```golang
type Persistable interface {
    Deletable
    Updatable
    Creatable
}
```

当函数需要两种或两种或两种以上动作的时候，上述的方式是很有用的。假设你需要定义一个接口包含 `Get` 或 `View` 操作，你可以考虑
重新定义一个 `ReadOnly` 的接口，包含 `Get`，`View` 操作，并定义一个 `Modifiable` 的接口，包含 `Update`, `Create`, `Delete` 操作。

试想一下你正在编写一组 http handlers 来实现对资源进行增删改查（CRUD）的API接口：

```golang
Create
Update
Delete
List
GetByID
```

通常来说，会像下面代码中这样，你可以对每个函数都定义一个接口，你所有的资源都需要实现这个接口里的方法，这样对于所有的实现，你都可以
统一调用 "Create" 方法来进行资源的创建：

```golang
func CreateHandle(c Creatable) func(w http.ResponseWriter, r *http.Request) {
    return http.HandleFunc("/resource", func(w http.ResponseWriter, r *http.Request) {
        if err := c.Create(r.Context); if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusCreated)
    })
}
```

如果你想要为 handler 编写测试的话，不论 resource 的实现有多复杂，你只需要确保 mock 对象实现了 `Creatable` 接口，这意味着你的 mock
对象只需要实现一个方法（因为 `Creatable` 接口只包含一个方法）。 文中描述的仅仅是一个简单的例子，假设你希望增加验证的逻辑，那么你仅需要在
`Creatable` 接口中添加方法 `func Valid() error`。

```golang
func CreateHandle(c Creatable) func(w http.ResponseWriter, r *http.Request) {
    return http.HandleFunc("/resource", func(w http.ResponseWriter, r *http.Request) {
        if err := c.Valid(); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        if err := c.Create(r.Context); if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusCreated)
    })
}
```
---
via: https://gianarb.it/blog/interface-segreation-in-action-with-go

作者：[gianarb](https://twitter.com/gianarb)
译者：[jamesxuhaozhe](https://github.com/jamesxuhaozhe)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
