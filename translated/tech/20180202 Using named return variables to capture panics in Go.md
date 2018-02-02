# Using named return variables to capture panics in Go

这将是一个简短的帖子，灵感来源于Sean Kelly's十一月的推特。

```
我发现了一个在golang中使用命名返回值的原因并且我现在感到潸然泪下。
                    — Sean Kelly (@StabbyCutyou) 2017年11月15日
```
其目标是为了记录并说明在一种有必要使用命名返回变量的情况，所以说让我们进入正题。

想象你正在编写一些用了可panic的函数的代码，并且无论什么原因你都不能改变那些函数。

```
func pressButton() {  
  fmt.Println("I'm Mr. Meeseeks, look at me!!")
  // other stuff then happens, but if Jerry asks to 
  // remove 2 strokes from his golf game...
  panic("It's gettin' weird!")
}
```
你一直需要用到那个函数，不过如果它发生异常，你想要获取到这个异常并将它作为一个错误返回，那么你应该写一些类似这样的代码：

```
func doStuff() error {  
  var err error
  // If there is a panic we need to recover in a deferred func
  defer func() {
    if r := recover(); r != nil {
      err = errors.New("the meeseeks went crazy!")
    }
  }()

  pressButton()
  return err
}
```

在Go Playground上执行它 - https://play.golang.org/p/wzkjKGqFPL


之后你go run了你的代码然而...这是什么？你的error是nil甚至是在代码发生异常的时候。这不是我们想要的！

## 为什么发生这种情况？

虽然最初看起来我们的代码返回的是我们在函数开始时创建的var err错误，但事实是我们的程序永远不会到达这一行代码。这意味着它永远不会返回特定的err变量，并将其更改为我们的defer函数，结果是毫无意义的。

在调用pressButton之后添加一个Println,但在返回之前，确实有助于说明这一点:

```
pressButton()  
// Nothing below here gets executed!
fmt.Println("we got here!")  
return err  
```

在Go Playground上执行它 - https://play.golang.org/p/Vk0DYs20eB

## 我们该如何修复它

为了修复这个问题，我们可以简单地使用命名返回变量

```
func doStuff() (err error) {  
    // If there is a panic we need to recover in a deferred func
    defer func() {
        if r := recover(); r != nil {
            err = errors.New("the meeseeks went crazy!")
        }
    }()

    pressButton()
    return err
}
```
在Go Playground上执行它 - https://play.golang.org/p/bqGOroPjQJ

最终的代码看起来非常的相似，不过通过命名我们的返回变量，当我们声明这个函数时我们的程序将会立刻发挥这个错误变量，即使我们从未触碰到doStuff函数的末尾的返回语句。由于这个细微的差别，我们现在可以更改我们被defer的函数中的err变量，并且成功地捕获到这个panic。





via: https://www.calhoun.io/using-named-return-variables-to-capture-panics-in-go/

作者：[Jon Calhoun](https://www.usegolang.com/)
译者：[Maple24](https://github.com/Maple24)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

