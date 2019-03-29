# 原文连接 [https://boyter.org/posts/my-personal-complaints-about-golang]
我个人关于用go编程的一些不满
go 作为一个相当棒的语言，

 However because questions about why I have issues with tends to come up often enough on the company slack programming channel (see what I did there?) I figured I would write them down and put it here so I can point people at a link when they ask what my complaints are.
 
 然而，（省略翻译）为什么我遇到的问题往往会在公司slack（通讯软件）编程频道上上经常出现
 
我想把他们写在这里，方便我能给那些问我不满什么的人一个连接


去年我大量使用GO，编写过命令行应用程序，ssc(统计代码行数的工具) lc(语法检查) 和API。


一个提供了大量接口的语法高亮显示器 不久之后将会被用于https://searchcode.com/

我在这里仅对go提出一些批判的意见，但是我对每一种我使用过的语言都有抱怨。实际上，以下的引用非常合适。


'有两种语言，一种人们经常抱怨，另一种没人用 -Bjarne Stroustrup'


#1 缺乏函数式编程
我不热衷于函数式编程，当我想到Lisp(https://en.wikipedia.org/wiki/Lisp_(programming_language))时，首先想到的是语言障碍这个单词。
这可能是我认为的go语言的最大的痛点。与大多数人相反，我不支持泛型，我认为只会给大多数的go项目带来不必要的复杂性。我想要的是一些适用于切片和集合的GO语言的内置方法。这两种类型在某种意义上都是神奇的，他们能支持任何类型并且是通用的，
你不能在不使用接口和失去安全性和速度的情况下执行你的GO代码

考虑下面的代码
给出两个字符串找出公共部分.
```golong
existsBoth := []string{}
for _, first := range firstSlice {
	for _, second := range secondSlice {
		if first == second {
			existsBoth = append(existsBoth, proxy)
			break
		}
	}
}
```

以上是在Go中解决这个问题的一个简单方法。还有别的方式去解决，例如使用集合，这种做法将减少运行时间。我们假设一下我们只有少量的内存，或者是我们不去处理很大的切片，并且不会有其余复杂的耗时操作。我们将相同的逻辑操作与java streams 方法和函数式编程。做对比
```jave
var existsBoth = firstList.stream()
                .filter(x -> secondList.contains(x))
                .collect(Collectors.toList());
```
现在，这种代码隐藏了所发生的算法的复杂性，但它更容易看到它实际上在做什么。

与go代码相比，可以显而易见的看出来代码的意图。并且添加额外的过滤器也很简单，同时代码也可以保持整洁



要向Go示例添加其他过滤器（如下面的示例），我们需要在已嵌套的循环中添加更多if条件。

There are projects which using go generate can achieve some of the above for you, but without nice IDE support its clunky and more of a hassle over pulling out the loop above into its own method.
有些项目使用go generate可以为你实现上述一些功能,但没有很好的IDE支持它笨重而且更麻烦的是将上面的循环拉出到它自己的方法中。




#2 channel/并行处理

go的channel是非常整洁的，你可以阻塞在出现问题的地方，并且channel不提供没有意义的并发，而是通过竞争检测来


当出现了一些问题的时候，你可以阻塞在有问题的位置。但channel并不是提供无畏的并发性，而且通过竞争检测器，你可以很容易地解决这些问题。




neat。有趣。 what is really neat about it is that 

slack。一个通讯工具。 company slack programming channel

让我们对比一下同样的逻辑使用Java streams 和函数式编程，以上的代码掩饰了算法的复杂性，但是又很清晰表现出实际上做了什么。与复制的Go代码相比，代码的意图很明显，真正有意思的是，添加额外的过滤条件也很简单。要像下面的示例代码一样添加过滤条件，我们需要向已有的循环条件中添加更多的if条件判断。有一些go写的project可以帮你实现上面的一些功能，但是如果没有好的ide支持，将上面的循环拉出到自己的方法中是很笨拙麻烦的。

go的channel也很简洁，你可能在出现问题的地方永远阻塞，但是channel不提供不安全的并发，通过竞争检测，你可以轻松的解决这些问题。对streaming values来说，你不知道有多少values，并且他们什么时候会结束，除非处理的值不受CPU的限制。

他们不是那么适合分片处理。，前面有多少需要处理并且想要并行的处理它们。

几乎是所有语言中都很常见，当你有一个大的列表或者切片，你使用并行stream ，并行linq，rayon，或者其他办法让所有可用的cpu迭代这个列表。你提供给他们一个列表，并且拿到处理后元素的列表集合，如果你有足够的元素并且正在使用的功能足够复杂，在多核系统中，可能处理的更快。

然而在go语言中，为了实现这一目的你需要做的的事情并不明显
一种可能的解决方案是为切片中的每个元素生成一个go-routines，由于go-routines的开销很低，这也是一个有效的策略
以上的操作会保证元素在切片过程中的顺序，但是让我们在这个案例中不需要保证顺序，



# 3垃圾回收

# 总结
