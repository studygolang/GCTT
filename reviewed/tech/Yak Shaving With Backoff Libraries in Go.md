# 关于在 Go 代码中使用退避方法，啰嗦几句（Yak Shaving With Backoff Libraries in Go）

我相信你有过调用 API 接口需要使用退避算法的时候。在 Go 语言现有技术中，有 [github.com/cenkalti/backoff](https://github.com/cenkalti/backoff)，[github.com/jpillora/backoff](https://github.com/jpillora/backoff)，和其它库可以使用。

我曾经使用过 [github.com/cenkalti/backoff](https://github.com/cenkalti/backoff)，但是有一件让我感到困惑的事：它要求你为操作加上闭包，强制输入为 func() error 的形式。

举个例子，当你需要一个可以自动重试的函数（如下面的 myFunc 函数），返回 3 个值和一个 error，你需要和作用域斗智斗勇。

```go
var a, b, c Result
backoff.Retry(func() error {
  var err error
  a, b, c, err = myFunc(arg, ...)
  return err
}, backoffObject)
```

在性能方面，这可能可以忽略不计，但作用域全搞乱了，这个时候你必须要小心，在赋值时不要使用 :=，而是使用 = 来确保获得正确的值。

这是我写 [github.com/lestrrat-go/backoff](github.com/lestrrat-go/backoff) 库的主要原因。


使用这个库，你将不得不写更多的样板代码（使用库需要执行的函数：NewExponential()，Start() 等），但是你不需要实现一个闭包，我发现这样更加符合 Go 的风格。计算退避持续时间仍然很难，如果你有更好的算法，请告诉我。

使用 [github.com/lestrrat-go/backoff](github.com/lestrrat-go/backoff) 库，首先你要创建一个策略对象：

```go
policy := backoff.NewExponential(...)
```

传入的参数也会影响到策略，例如配置最大重试次数或者不重试。策略对象可以被多个消费者重用。

实际上 backoff 对象是通过调用策略对象的 Start() 方法创建的：

```go
b, cancel := policy.Start(ctx)
```

该方法接收一个 context 对象，因此你可以通过父作用域内的释放操作终止退避算法。

当退避对象 b 不再需要时，cancel 用于释放资源。

退避对象 b 包含两个方法，Done() 和 Next()。他们都返回一个可以通知我们事件的管道变量。

当退避算法停止时，Done() 变得可读：（停止的情况）包括父 context 被取消导致退避算法被取消或者某个条件发生（例如当我们已经重试了 MaxRetries 参数指定的次数时）。

当调用过了足够的时候后，Next() 才变为可读。在指数级回退的情况下，在调用 Next() 之后，你可能得等 1， 2， 4， 8， 16...（乘以基本区间）。

使用这些方法，你的退避方法将看起来就像这样：

```go
func MyFuncWithRetries(ctx context.Context, ... ) (Result1, Result2, Result3, error) {
  b, cancel := policy.Start(ctx)
  defer cancel()
  for {
    ret1, ret2, ret3, err := MyFunc(...)
    if err == nil { // success
      return ret1, ret2, ret3, nil
    }
    select {
    case <-b.Done():
      return nil, nil, nil, errors.New(`all attempts failed`)
    case <-b.Next():
      // continue to beginning of the for loop, execute MyFunc again
    }
  }
}
```

使用这种方法的缺点是你需要更多的样板代码。优点是你不再需要奇怪的作用域技巧，操作变得更加直接。

我认为这是一个关于口味的问题，所以你应该选择符合你口味或者习惯的库。这恰好是我想要的库，希望也对你有帮助。

编程快乐！

via: https://medium.com/@lestrrat/yak-shaving-with-backoff-libraries-in-go-80240f0aa30c

作者：[Daisuke Maki](https://github.com/lestrrat)
译者：[gogeof](https://github.com/gogeof)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

