---
layout:     post
title:      My Personal Complaints about Programming in Go
subtitle:   
date:       2019-03-30
author:     junpeng
header-img: img/post-bg-kuaidi.jpg
catalog: true
tags:
    - GO
---

#<center>我个人关于用go编程的一些不满</center>

GO是相当棒的语言。但是经常在公司的slack编程交流中出现我的问题，我得把他们写下来，当人们问我对GO有什么不满的时候我可以给他们个链接。

[图片](/img/2019-03-30/gloang.png)

去年我大量使用GO编程，写过命令行应用程序，[scc](https://github.com/boyter/scc/)(统计代码行数的工具)[lc](https://github.com/boyter/lc/)(语法检查)和API。其中包括一个提供了大量API的语法高亮显示器https://searchcode.com/

我在这里仅对GO提出一些意见，其实我对每一种使用过的语言都有抱怨。实际上，下面这段引用非常合适。
>有两种语言，一种人们经常抱怨，另一种没人用。 -Bjarne Stroustrup


##1 缺乏函数式编程

我不热衷于函数式编程，当我看到[Lisp](https://en.wikipedia.org/wiki/Lisp_(programming_language))时，首先出现在我脑中的是语言障碍这个单词。

这可能是我认为的GO语言最大的痛点。与大多数人相反，我不想要generics(泛型)，我认为(泛型)只会给大多数的GO项目带来不必要的复杂性。我想要的是一些适用于数组和集合的GO语言的内置方法。这两种类型在某种意义上都是神奇的，他们能支持任何类型并且是通用的，你不能直接在GO语言中使用，除非你写接口并且还会损失安全性和速度。

看看下面的代码。

给出两个字符串找出公共部分并放入一个数组中，我们接下来会处理它。

```<long>
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

以上是在GO中解决这个问题的一个简单方法。还有别的方式去解决，例如使用集合可以减少运行时，我们假设一下我们受到内存限制，或者是我们不去处理很大的数组，并且不会有其余复杂的耗时操，我们将相同的逻辑操作与java中streams和函数式编程做对比。

```jave
var existsBoth = firstList.stream()
                .filter(x -> secondList.contains(x))
                .collect(Collectors.toList());
```

现在上面代码隐藏了算法的复杂性，也更清晰的明白实际上做了什么。

意图比GO实现的代码更清晰。并且添加额外的过滤器也很简单。同时代码也可以保持整洁，向GO代码添加其他过滤器，则需要在已有循环中添加更多if条件。

```golong
var existsBoth = firstList.stream()
                .filter(x -> secondList.contains(x))
                .filter(x -> x.startsWith(needle))
                .filter(x -> x.length() >= 5)
                .collect(Collectors.toList());
```

有一些使用`go generate`的项目也可以实现上面的一些功能，但是如果没有好的IDE支持，就会显得很笨拙，而且在将上面的循环提取到它自己的方法中时会遇到更多的麻烦(存疑）

##2 Channel/Parallel Slice Processing

GO的`channel`是非常整洁，你可能在出现问题的地方永远阻塞，但是`channel`不提供不安全的并发，通过竞争检测，你可以轻松的解决这些问题。对比较`streaming values`来说，你不知道有多少问题存在，并且他们什么时候会结束，除非处理`values`的方法不受CPU的限制。

它们不太适合处理预先知道大小并希望能并行处理的数组操作（这句话我自己也有疑问）
> 多线程编程，理论与实践

[图片](/img/2019-03-30/mulitithreaded.png)

在处理一个大的列表或者数组时，大多数语言都会使用`parallel stream` ，`parallel linq`，`rayon`，    `多线程`，或者其他办法让所有可用的CPU迭代这个列表。你提供一个列表，并且拿到处理后元素的列表集合，如果你有足够的元素并且正在使用的功能足够复杂，在多核系统中可能处理的更快。

然而在go语言中，为了实现这一目的你需要做的的事情并不明显。

一种可能的解决方案是为数组中的每个元素生创建一个`go-routines`。由于`go-routines`的开销很低，所以这是一个有效的策略。

```golong
toProcess := []int{1,2,3,4,5,6,7,8,9}
var wg sync.WaitGroup

for i, _ := range toProcess {
	wg.Add(1)
	go func(j int) {
		toProcess[j] = someSlowCalculation(toProcess[j])
		wg.Done()
	}(i)
}

wg.Wait()
fmt.Println(toProcess)
```

以上的操作会保证元素在数组过程中的顺序，但是我们在这个案例中不需要保证顺序。
以上代码存在的首要问题是增加了`waitgroup`并且还需要记录下`wg.Add(1)`和`wg.Done()`。这是给开发人来带来的额外操作。遇到错误或者是没能拿到正确的输出，也可能是其他不确定的结果或者程序不能结束。另外，如果你的列表很长并且准备给每一个元素都分配`go-routine`的话。就像我之前说的一样，这种做法不麻烦，因为Go可以很轻松的实现。真正会成为问题的是每一个 `go-routines`都争夺CPU资源，这将不是最高效的执行任务的方式。

你可能想到的是为每个CPU创建一个`go-routine`，轮流从列表中取得数据然后处理。`go-routine`带来的额外开销是很小的。并且对一个短的循环是微不足道的。自从每个核限制一个`go-routine`后，我遇到了一些问题，当我用[scc](https://github.com/boyter/scc/)工作的时候。用Go实现的话，你需要创建一个管道，然后循环数组中的每个元素，然后从这个管道中读取(存疑)。看一下下面的代码。

```golong
toProcess := []int{1,2,3,4,5,6,7,8,9}
var input = make(chan int, len(toProcess))

for i, _ := range toProcess {
	input <- i
}
close(input)

var wg sync.WaitGroup
for i := 0; i < runtime.NumCPU(); i++ {
	wg.Add(1)
	go func(input chan int, output []int) {
		for j := range input {
			toProcess[j] = someSlowCalculation(toProcess[j])
		}
		wg.Done()
	}(input, toProcess)
}

wg.Wait()
fmt.Println(toProcess)
```

上面的代码创建一个管道，循环我们的数组然后将值放入管道中，然后我们给每一个CPU创建一个`go-routine`，接下来处理管道的元素，然后我们等待直到结束，很多代码需要理解。

如果你的数组很大，你可能不会想着去创建一个同样长度的channel，所以你真正该做的是再创建`go-routine`去循环数组，然后将数组的值放入管道中，完成后关闭管道，我放弃了这段代码，虽然和我的主要思想相似，但是代码太长了。

来看看java版本近似的代码。

```java

var firstList = List.of(1,2,3,4,5,6,7,8,9);

firstList = firstList.parallelStream()
        .map(this::someSlowCalculation)
        .collect(Collectors.toList());
```

`channel`和streams是不一样的，你可以使用`queue`实现的更接近c`channel`，但是并不能做到完全一样，尽管我们同样都在使用所有的CPU都处理`list/slice`。

如果`someSlowCalucation`是一种调用网络或其他非CPU密集型任务的不需要关注运行耗时的操作，这样写也是没什么问题的。在这种情况下`channel`和`go-routines`也是很合适的。

这个问题`＃1`有关。如果GO在`slice/map`对象上添加功能是可行的话。Go是否有人能写出想库一样每个人都可以获益的`generics`泛型也是个烦人的问题。顺便说一句，我认为这阻碍了GO在数据科学领域取得的任何成功，因此Python仍然是其中的佼佼者。GO在数值操作中缺乏表现力和力量，这就是产生现在的现象的原因。

## 3垃圾回收

go的垃圾回收器是非常可靠的。我参加的每一个发布了的应用速度也越来越快，通常是因为垃圾回收机制的提升。`prioritizes latency`优先低延迟性超出了所有需求，对于`API's`(应用程序编程接口)，`  UI's`(用户界面接口)来说都是非常好的。并且也适用于任何网络调用将要成为瓶颈的应用。

问题是Go对UI处理上(存疑，用户图形界面，我也想不到合适的解释)也不友好，当你想要高性能时。我在使用[scc](https://github.com/boyter/scc/)时遇到了这个问题，[scc](https://github.com/boyter/scc/)是一个非常受CPU限制的命令行应用程序。这是个问题，我在其中添加了逻辑来关闭GC，直到它达到阈值。实际上我并不能禁用GC，因为在某些场景下内存很快会被耗尽。

缺乏对GC的控制有时很令人丧气。你需要学着适应它。有些时候你要微笑的对它说：“嗨，这片代码需要尽快运行，你只需要高性能工作一会就好了”。

[图片](/img/2019-03-30/throughput.png)

我认为GC在Go1.12版本之后不会再有什么提升了。当我打开或者是关闭GC的时候我就不能再控制GC了。我有空的时候应该再去研究一下。

##4 错误处理

这不是我一个人抱怨的，但是又有必要写一下。

```long
value, err := someFunc()
if err != nil {
	// Do something here
}

err = someOtherFunc(value)
if err != nil {
	// Do something here
}   
```

Is pretty tedious. Go does not even force you to handle the error either which some people suggest. You can explicitly ignore it (does this count as handling it?) with _ but you can also just ignore it totally. For example I could rewrite the above like,


很无聊，GO甚至没有强制你使用像一些人建议的方式处理这个错误。您可以使用`_`显式忽略它（这是否算作处理它？），其实也可以完全忽略它。例如，我可以重写上面的内容.

```golong
value, _ := someFunc()

someOtherFunc(value)

```

很容易的看出我忽略了`someFunc`返回的一些值，并且`someOtherFunc(value)`也能返回错误。而我完全无视它。根本没有处理这种情况

坦诚地说，我不知道解决的办法，我喜欢这个 `RUST`中`?` 运算符来帮助避免这种情况。`V-Lang https://vlang.io/看起来也可能有一些有趣的解决方案。

另一个想法是可选类型和删除`nil`，尽管即使使用Go2.0也不会发生这种情况，因为它会破坏向后兼容性。

## 总结

Go仍然是一种非常不错的语言。如果你告诉我要写一个API，或者某个需要大量磁盘操作或者网络调用的任务，GO仍然是我的第一个选择。实际上我是用Python写许多一次性任的任务，除了数据合并之外，缺乏函数式编程仍足以让受到速度的影响。像字`stringA==stringB`和编译错误之间的合理比较。如果你尝试对`slice`做同样的操作是很好的事情如果能保持最小的意外的原则，不像我在上述比较中使用的Java代码哪样(存疑)。

是的，使用二进制之后还能更小（一些[编译标志](https://boyter.org/posts/trimming-golang-binary-fat/)和`upx`可以实现），我希望它在某些方面更快，GOPATH不是很好但也没有每个人做出的那么糟糕，默认的单元测试框架缺少很多功能，一个被嘲笑的痛点......

它仍然是我使用过的高效语言之一。我会继续使用它，虽然我希望https://vlang.io/最终会发布并解决我的很多抱怨。无论是vlong发布还是Go 2.0，Nim还是Rust。这些天来，有很多很酷的新语言可以玩。我们开发人员真的太幸福了。

---

via: https://boyter.org/posts/my-personal-complaints-about-golang/

作者：[Ewan Valentine](http://ewanvalentine.io/author/ewan)
译者：[junpengxu](https://github.com/junpengxu)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出