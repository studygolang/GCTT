首发于：https://studygolang.com/articles/20543

# 在 Go 中使用服务对象

服务对象是 `Ruby on Rails` 中一个高度可用的模式，它能够保持控制器和模型简洁干净并从两者中删除域逻辑。在我看来，服务对象是单一责任原则以及通过依赖注入分配责任的一个很好的例子。

一般而言，`SOLID` 及其背后的理念允许编写可测试的代码，这对于更改非常灵活。`Robert "Uncle Bob" Martin` 推动了这些原则。`SOLID` 原理理论在 2000 年的论文 [`Design Principles and Design Patterns.`](https://fi.ort.edu.uy/innovaportal/file/2032/1/design_principles.pdf) 中有所介绍。`Dave Cheney` 有一篇很棒的关于这个原理的文章[`SOLID Go Design`](https://dave.cheney.net/2016/08/20/solid-go-design)。

`Robert Martin` 在他的书 `Clean Architecture: A Craftsman's Guide to Software Structure and Design` 中还提出了一个包含四个级别职责的架构：实体，用例，接口适配器，框架和驱动程序。这个体系结构引入了 `用例`，其原因与 `Ruby on Rails` 中的服务对象相同 - 用于封装业务逻辑。

广泛使用接口和依赖注入可以使代码独立于 UI，框架和驱动程序。此方法还提供了使用提供的 UI 和存储的模拟实现来测试业务逻辑的能力。

举个例子，让我们看看下面的代码，以及使用 SRP 和引入用例级别会有多好。

```go
// Repository is a data access layer.
type Repository interface {
    Exists(email string) (bool, error)
    Create(*Form) (*User, error)
}

// RegistrationHandler for handling registration requests.
type RegistrationHandler struct {
    Validator *validator.Validate
    Repository
}

// ServerHTTP implements http.Handler.
func (h *RegistrationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var f Form
    if err := JSON.NewDecoder(r.Body).Decode(&f); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    validations := make(map[string]string)

    err := h.Validator.Struct(f)
    if err != nil {
        if vs, ok := err.(validator.ValidationErrors); ok {
            for _, v := range vs {
                validations[v.Tag()] = fmt.Sprintf("%s is invalid", v.Tag())
            }
        }
    }

    if f.Password != f.PasswordConfirmation {
        validations["password"] = passwordMismatch
    }

    exists, err := h.Exists(f.Email)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    if exists {
        validations["email"] = emailExists
    }

    if len(validations) > 0 {
        w.WriteHeader(http.StatusUnprocessableEntity)
        JSON.NewEncoder(w).Encode(validations)
        return
    }

    u, err := h.Create(&f)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    JSON.NewEncoder(w).Encode(&u)
}
```

正如您所看到的，除了此处理程序形成传入请求的响应之外，它还包含用户注册过程的所有业务逻辑。每次上一步失败时，此代码都会写入响应并中断其他步骤。

当变化到来时，粘度症状将会上升，因为没有明显的设计可以保留，并且代码的每次更改都会成为某种黑客行为。如果需要将通知发送给注册用户以验证其电子邮件，则代码将变得更难理解和测试。

请参阅[github](https://github.com/romanyx/service_object) 上的完整代码示例。

提供的代码示例的好处是前面的开发人员为注册请求创建了集成测试，它不是遗留代码，这意味着它可以被重构。

所以让我们开始一些重构并应用用例级别。第一步是将注册过程逻辑封装到服务对象中。

```go
// Repository is a data access layer.
type Repository interface {
    Unique(email string) error
    Create(*Form) (*User, error)
}

// Validater validation abstraction.
type Validater interface {
    Validate(*Form) error
}

// ValidationErrors holds validation errors.
type ValidationErrors map[string]string

// Error implements error interface.
func (v ValidationErrors) Error() string {
    return validationMsg
}

// Service holds data required for registration.
type Service struct {
    Validater
    Repository
}

// Registrate holds registration domain logic.
func (s *Service) Registrate(f *Form) (*User, error) {
    if err := s.Validater.Validate(f); err != nil {
        return nil, errors.Wrap(err, "validater validate")
    }

    user, err := s.Repository.Create(f)
    if err != nil {
        return nil, errors.Wrap(err, "repository create")
    }

    return user, nil
}
```

该 `Registrate` 方法在系统中注册用户需要两个步骤：

1. 验证传入的表单。
2. 将模型插入存储器。

随着服务对象的引入，以前粘性的代码变得更加明显，易于理解。如果发生变化，工程师可能会理解并保留现有设计。

```go
// Registrater abstraction for registration service.
type Registrater interface {
    Registrate(*Form) (*User, error)
}

// RegistrationHandler for regisration requests.
type RegistrationHandler struct {
    Registrater
}

// ServerHTTP implements http.Handler.
func (h *RegistrationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var f Form
    if err := JSON.NewDecoder(r.Body).Decode(&f); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    u, err := h.Registrate(&f)
    if err != nil {
        switch v := errors.Cause(err).(type) {
        case ValidationErrors:
            w.WriteHeader(http.StatusUnprocessableEntity)
            JSON.NewEncoder(w).Encode(v)
        default:
            w.WriteHeader(http.StatusInternalServerError)
        }
        return
    }

    JSON.NewEncoder(w).Encode(&u)
}
```

变更的[代码](https://github.com/romanyx/service_object/pull/1/files)。

在现代 Web 开发中提出了许多要求，其中之一就是可观察性。可观察性包括日志记录，度量和跟踪，这使得能够不对任务的性能和问题的来源进行猜测，而是跟踪和修复问题。

但是，要实现该目标，需要在给定代码中通过将上下文传播到服务对象来实现，以便可以跟踪传入请求，并且可以将与其相关的日志绑定到特定 `TraceID`。

```go
// Service holds data required for registration.
type Service struct {
    Validater
    Repository
}

// Registrate holds registration domain logic.
func (s *Service) Registrate(ctx context.Context, f *Form) (*User, error) {
    if err := s.Validater.Validate(ctx, f); err != nil {
        return nil, errors.Wrap(err, "validater validate")
    }

    user, err := s.Repository.Create(ctx, f)
    if err != nil {
        return nil, errors.Wrap(err, "repository create")
    }

    return user, nil
}
```

示例中使用到的 `context` 在 `Sameer Ajmari` 的博文[Go blog](https://blog.golang.org/context) 有所介绍，其中还提及了它应该在所有的传入和传出请求的路径上的传播的好处。

变更的[代码](https://github.com/romanyx/service_object/pull/2/files)。

这样我们现在可以使用装饰器模式扩展服务对象并应用日志记录，跟踪和我们需要的所有其他扩展。您可以直接编写此类装饰器或使用某些工具生成它们。已经有一个允许装饰接口的发生器 - [gowrap](https://github.com/hexdigest/gowrap)。Max Chechel-- 作者，在 GoWayFest 2.0 的演讲 ["Code Generation to Survive"](https://www.youtube.com/watch?v=pFFfurrCEcM) 中解释了为什么你可能需要这样一个工具。

示例[代码](https://github.com/romanyx/service_object/pull/3/files)。

在服务对象之上应用装饰器模式使我们能够扩展它并达到许多目标，如度量和跟踪等，并将我们的代码移动到现代微服务时代。

尽管 Go 不是通常意义上的 OOP 语言，但用它编写的代码也应该是直观的并且具有明确的结构。这些代码可以使用 `SOLID` 原则中包含的原则来编写，该原则带有适用于许多编程语言的通用方法集。

本文的目的是表达我对 Go 中编写的代码应该是什么样的理解，并且我希望它包含比负面代码更多的积极方面。

---

via: https://itnext.io/using-service-objects-in-go-d899dc599335

作者：[Roman Budnikov](https://itnext.io/@romanyx90)
译者：[lovechuck](https://github.com/lovechuck)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
