首发于：https://studygolang.com/articles/23462

# 闲谈 Go 1.13 的错误处理

Go 1.13 丰富了 `errors` 包。这些新增部分源自 [Go2 的错误监控提议](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md)。那么让我们看看都有些啥吧 ~

Go 的错误是任意实现 `error` 接口的值。

```go
// The error built-in interface type is the conventional interface for
// representing an error condition, with the nil value representing no error.
type error interface {
   Error() string
}
```

错误本质上是一些字符串，易于人们阅读和理解，但程序要理解它们就要难得多了。

当前有 4 种编程式地处理错误的常见方式，分别如下：

- 哨兵型错误
- 类型断言
- 临时检查
- 字符子串查找

## 哨兵型错误

一些包定义导出型错误变量，并在函数调用时返回它们。`sql.ErrNoRows` 就是这么个例子：

```go
package sql

// ErrNoRows is returned by Scan when QueryRow doesn’t return a
// row. In such a case, QueryRow returns a placeholder *Row value that
// defers this error until a Scan.
var ErrNoRows = errors.New(“sql: no rows in result set”)
```

拿到返回值后，我们将其和 `sql.ErrNoRows` 比较即可：

```go
if err == sql.ErrNoRows {
    ... handle the error ...
}
```

## 类型断言

和哨兵型错误类似，这种情况下我们想要检查返回的错误是否源自能够提供更多信息的特定错误类型。在 `os` 包可以看到一个很好的例子：

```go
type PathError struct {
    Op   string
    Path string
    Err  error
}

func (e *PathError) Error() string
func (e *PathError) Timeout() bool
```

借助类型断言，我们可以获取到 `PathError` 提供的所有额外信息

```go
if pe, ok := err.(*os.PathError); ok {
    if pe.Timeout() { ... }
    ...
}
```

## 临时检查

这是一些辅助函数，抽象了如何判断给定错误的可能类型。这种方式的显著优势之一是：一个包可以暴露这些方法，而不公开其错误处理的内部逻辑。

```go
// IsNotExist returns a boolean indicating whether the error is known to
// report that a file or directory does not exist. It is satisfied by
// ErrNotExist as well as some syscall errors.
func IsNotExist(err error) bool

if os.IsNotExist(err) {
      ...
}
```

## 字符子串搜索

名字表明了方法的操作姿势，用远古的 `strings.Contains` 来检查某些东西是否存在。在这四种方法中，是最不可取的。

```go
if strings.Contains(err.Error(), "foo bar") {
    ...
}
```

## 需要添加更多上下文或信息时：包裹

我们经常需要添加一些更加明确的信息，例如解释失败的起因。这是前面所述方法都做不到的。例如，我们可能要说明提取操作的失败原因是一个 sql 错误。

包裹本质上是创建一条错误链，允许我们添加更多信息，同时保留原始错误。基于能够存储一个错误和更多信息的任意类型很容易就可以实现。假设现有如下类型：

```go
type myError struct {
    msg string
    err error
}

func Wrap(err error, msg string, args ...interface{}) error {
   return myError{
        msg: fmt.Sprintf(msg, args...),
        err: err,
   }
}
```

我们很轻易就可以创建一些方法来遍历整条错误链，一层层地不断从外层错误剥离出底层错误。

Go 生态中有些牛逼的库对此进行了实现。在 Onefootball 这里，我们的微服务大肆使用的是 [github.com/pkg/errors](https://github.com/pkg/errors)（如果不认识的话，赶紧学习一波哟）。

但是这种方式有个缺点：由于拆包和访问其他信息都只能借助库的 API，我们被紧密绑定到包装错误的第三方库。

## Go2 错误监控提议

Go2 有一项提议是添加接口用于拆开错误：

```go
// Unwrap returns the result of calling the Unwrap method on err, if err’s
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
type Wrapper interface {
    Unwrap() error
}
```

> 作者认为将其称为 Unwrapper 要比 Wrapper 好。

这个简洁的接口使得任意 Go 程序能够拆开任意自定义错误。如果当前包装器实现了 `Unwrap` 函数，我们不用烦心于混杂的自定义错误数目即可遍历整条错误链。

更喜人的是对于*哨兵型错误*和*类型断言*，Go 的 `errors` 包给它们定义了标准方法，分别是 **Is** 和 **As**：

```go
func Is(err, target error) bool
func As(err error, target interface{}) bool
```

更多细节如下：

```go
package errors

// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
func Is(err, target error) bool

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// As will panic if target is not a non-nil pointer to either a type that implements
// error, or to any interface type. As returns false if err is nil.
func As(err error, target interface{}) bool
```

## Go 1.13 的情况如何？

Go 1.13 定义了上述的 **Unwrap**、**Is** 和 **As** 函数。

**Unwrap** 是为某些 `error` 类型的变量调用 `Unwrap()` 函数的快捷方式。由于这两个方法（@TODO: 具体所知未明。译注：会破坏接口的兼容性）都没有添加到 `error` 接口上，`errors.Unwrap` 是很好用的。

**Is** 和 **As** 则会遍历整条错误链条直至找到匹配的错误或 `nil`，从而进行类型匹配或断言和把任意错误转换成 *target*。

你可能发现了 Go 1.13 没有新添接口，以上 3 个方法都是动态地检查给定错误是否实现了它们的：

```go
u, ok := err.(interface { Unwrap() error })

x, ok := err.(interface { Is(error) bool })

x, ok := err.(interface { As(interface{}) bool })
```

那错误是怎么被包裹的呢？稍安勿躁，`fmt` 来相助：

> `Errorf` 函数有个要求操作对象是错误的新动词 `%w`，其返回错误的 `Unwrap` 方法会返回 `%w` 对应的操作对象。

```go
err := errors.New(“my error”)
err = fmt.Errorf(“1s wrapping my error with Errorf: %w”, err)
err = fmt.Errorf(“2nd wrapping my error with Errorf: %w”, err)
```

## 要迁移到 Go 1.13 吗？最后温馨提示

注意了：*modules* 现在默认会使用 Google 运行的 Go 模块镜像与校验和数据库。长话短说就是：go 命令会从 Go 的模块镜像请求模块，根据 Go 的校验和数据库验证模块校验和。因此，你需要将私有仓库排除在这个流程之外。把 `GOPRIVATE` 设置为*逗号分割的模块路径前缀的通配模式 （要符合 Go 的 `path.Match` 方法要求的语法）*。例如：

```bash
GOPRIVATE=github.com/myOrg/*,*.corp.example.com,domain.io/private
```

祝你编程快乐！

PS：我在[柏林 Golang 见面会](https://www.meetup.com/golang-users-berlin/events/259188830/)对此做了个演讲，你可以从[这里](https://github.com/AndersonQ/go1_13_errors)找到相应的幻灯片和代码。

---

via: https://medium.com/onefootball-locker-room/a-look-at-go-1-13-errors-9f6c9f6accb6

作者：[Anderson Queiroz](https://medium.com/@AndersonQ)
译者：[sammyne](https://github.com/sammyne)
校对：[DingdingZhou](https://github.com/DingdingZhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
