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

via: https://www.calhoun.io/using-named-return-variables-to-capture-panics-in-go/

作者：[Jon Calhoun](https://www.usegolang.com/)
译者：[Maple24](https://github.com/Maple24)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

