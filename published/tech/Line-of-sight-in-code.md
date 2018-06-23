已发布：https://studygolang.com/articles/12928

# 代码中的缩进线

![我在2016年伦敦Golang英国会议上谈论代码缩进线](https://raw.githubusercontent.com/studygolang/gctt-images/master/line-of-sight/1_CBjBs9EzL8q1AL6XvjjpJg.png)

在近期伦敦举行的 [Golang 英国会议](https://www.youtube.com/watch?v=yeetIgNeIkc) 上，我在[地道的Go 语言窍门](https://www.youtube.com/watch?v=yeetIgNeIkc) 交流（[幻灯片](http://go-talks.appspot.com/github.com/matryer/present/idiomatic-go-tricks/main.slide#1)）中讲到关于代码中的缩进线， 我想在这里稍微解释一下。

> 缩进线是“观察者无障碍视线的直线”

![代码中的缩进线：左图的缩进是错误处理和边缘情况的快乐路径](https://raw.githubusercontent.com/studygolang/gctt-images/master/line-of-sight/1_nXXRSHi_1kmgorkcDHyc1Q.png)

良好的代码缩进线不仅对你的功能没有任何影响，还可以帮助其他需要的人阅读你的代码。其他程序员（包括你未来的自己）可以浏览一个专栏并且理解代码的预期流程。如果他们不得不在脑子里分析 `if` 语句，若没有良好的缩进线，将会使这个任务变得非常艰难。

> 大多数人关注编写代码的代价（比如“这需要多长时间才能完成？”）但是维护代码的成本要高得多 - 特别是在成熟的项目中。 让功能明显，清晰，简单易懂才是至关重要的。

良好缩进线的建议：

* 让[快乐路径](https://en.wikipedia.org/wiki/Happy_path)居左侧对齐，这样你就可以快速扫描一列来查看预期的执行流程
* 不要隐藏缩进大括号中代码逻辑
* 尽早的退出 `function`
* 避免 `else return`，考虑翻转 `if` 语句
* 把 `return` 声明作为最后一行
* 提取 `function` 和 `method` 以保持结构小巧和可读
* 如果你需要大缩进的代码，考虑当做一个 `function` 分解出来

当然，会有很多很好的理由来打破所有这些规则 - 但是采用这种风格作为默认规则，我们发现我们的代码变得更具可读性。

## 避免 `else return`

编写具有良好视觉效果的代码的关键是保持 `else` 结构小巧，或者如果可以的话，完全避免它们。 看下这个代码：

```go
if something.OK() {
	something.Lock()
	defer something.Unlock()
	err := something.Do()
	if err == nil {
		stop := StartTimer()
		defer stop()
		log.Println("working...")
		doWork(something)
		<-something.Done() // wait for it
		log.Println("finished")
		return nil
	} else {
		return err
	}
} else {
	return errors.New("something not ok")
}
```

这代表了我们最初如何思考我们的功能在做什么（“如果某件事情没问题，那么就做，如果没有错误，那么做这些事情”等等），但是它变得很难遵循。

上面的代码很难遵循'快乐路径'（执行顺利进行的路线）。它在第二行开始缩进并从那里继续。 当我们检查来自 `something.Do()` 的错误返回时，我们进一步缩进。 事实上，语句“ `return nil` ”在代码中间完全丢失。

`else` 结构在 Go 和其他语言中作为单一行返回很常见，因为它们要处理中止或退出函数。 我认为他们不能保证缩进我们的其他代码。

## 翻转 if 语句

如果我们要翻译 `if` 语句*（如果你喜欢* ， *就把它们*翻*过来）* ，你可以看到代码变得更加可读：

```go
if !something.OK() {  // flipped
	return errors.New("something not ok")
}
something.Lock()
defer something.Unlock()
err := something.Do()
if err != nil {       // flipped
	return err
}

stop := StartTimer()
defer stop()

log.Println("working...")
doWork(something)
<-something.Done() // wait for it
log.Println("finished")
return nil
```

在此代码中，我们正在尽早退出，退出代码与正常代码不同。而且，

* 快乐路径沿着左侧向下保持，
* 我们缩进只是为了处理错误和边缘情况，
* 我们的 `retutn` 声明“ `return nil` ”在最后一行，并且，
* 我们有更少的缩进代码块。

## 促进大型条件块的功能

如果你不能避免一个笨重的 `else` 结构或臃肿的选择切换的情况（我明白了，有时候你不能），那么就考虑把每个结构分解成它自己的功能：

```go
func processValue(v interface{}) error {
	switch val := v.(type) {
	case string:
		return processString(val)
	case int:
		return processInt(val)
	case bool:
		return processBool(val)
	default:
		return fmt.Errorf("unsupported type %T", v)
	}
}
```

这比读取大量的处理代码更容易阅读。

## 分享你的经验

如果你同意我的观点，请考虑分享这篇文章 - 随着越来越多的人注册，更好的（更一致的）Go 代码将会出现。

你有一些难以阅读的代码吗？ 为什么不在 [Twitter @matryer](https://translate.googleusercontent.com/translate_c?depth=1&hl=zh-CN&prev=search&rurl=translate.google.com.hk&sl=en&sp=nmt4&u=https://twitter.com/matryer&xid=17259,15700023,15700124,15700149,15700168,15700173,15700186,15700201&usg=ALkJrhgR995EkjexZDOQl9LYu8Sl7eq3TA) 上分享它，可以看看我们是否可以找到一个更清洁，更简单的版本。

## 致谢...

评论家[戴夫切尼](https://translate.googleusercontent.com/translate_c?depth=1&hl=zh-CN&prev=search&rurl=translate.google.com.hk&sl=en&sp=nmt4&u=http://dave.cheney.net/&xid=17259,15700023,15700124,15700149,15700168,15700173,15700186,15700201&usg=ALkJrhgTC1jmfDNNabAZ1iX8dJSOjyuddw) ， [大卫埃尔南德斯](https://translate.googleusercontent.com/translate_c?depth=1&hl=zh-CN&prev=search&rurl=translate.google.com.hk&sl=en&sp=nmt4&u=http://twitter.com/dahernan&xid=17259,15700023,15700124,15700149,15700168,15700173,15700186,15700201&usg=ALkJrhhm04bLuew6j4VCw3ACIxtPeMMmxA)和[威廉肯尼迪](https://translate.googleusercontent.com/translate_c?depth=1&hl=zh-CN&prev=search&rurl=translate.google.com.hk&sl=en&sp=nmt4&u=https://twitter.com/goinggodotnet&xid=17259,15700023,15700124,15700149,15700168,15700173,15700186,15700201&usg=ALkJrhi57B943koWpS6pe4_aRslBMy-7mw) 。

---

via: https://medium.com/@matryer/line-of-sight-in-code-186dd7cdea88

作者：[Mat Ryer](https://medium.com/@matryer)
译者：[yuhanle](https://github.com/yuhanle)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
