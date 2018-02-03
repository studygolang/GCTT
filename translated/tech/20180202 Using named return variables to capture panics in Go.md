# 在Go中使用指定的返回变量捕获panic

这将是一个简短的帖子，灵感来源于Sean Kelly十一月的推特。

```
我发现了一个在golang中使用指定的返回值的原因并且我现在感到潸然泪下。
                    — Sean Kelly (@StabbyCutyou) 2017年11月15日
```
其目标是为了记录并说明一种有必要使用到指定的返回变量的情况，所以说让我们进入正题。

想象你正在编写一些用了可以panic的函数的代码，并且无论什么原因(第三方库，向后兼容，等等)你都不能改变那些函数。

```
func pressButton() {  
  fmt.Println("I'm Mr. Meeseeks, look at me!!")
  // other stuff then happens, but if Jerry asks to 
  // remove 2 strokes from his golf game...
  panic("It's gettin' weird!")
}
```
你仍然需要用到那个函数，不过如果它发生异常，而你想要捕获到这个panic并将它作为一个error返回，那么你应该写一些类似这样的代码：

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


之后你go run了你的代码然而...这是什么？你的error是`nil`值,甚至是在代码发生异常的时候。这并不是我们想要的！

## 为什么会发生这种情况？

虽然最初看起来我们的代码返回的是我们在函数开始时创建的`var err error`，但事实是我们的程序永远不会到达这一行代码。这意味着它永远不会返回特定的err变量，并将其更改为我们的defer函数，结果是毫无意义的。

在调用`pressButton`之后添加一个`Println`,但在返回之前，确实有助于说明这一点:

```
pressButton()  
// Nothing below here gets executed!
fmt.Println("we got here!")  
return err  
```

在Go Playground上执行它 - https://play.golang.org/p/Vk0DYs20eB

## 我们该如何修复它

为了修复这个问题，我们可以简单地使用指定的返回变量。

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

最终的代码看起来非常的相似，不过通过指定我们的返回变量，当我们声明这个函数时我们的程序将会立刻发挥这个`err`变量，即使我们从未触碰到`doStuff`函数的末尾的返回语句。由于这个细微的差别，我们现在可以更改我们被defer的函数中的`err`变量，并且成功地捕获到这个panic。





via: https://www.calhoun.io/using-named-return-variables-to-capture-panics-in-go/

作者：[Jon Calhoun](https://www.usegolang.com/)
译者：[Maple24](https://github.com/Maple24)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

