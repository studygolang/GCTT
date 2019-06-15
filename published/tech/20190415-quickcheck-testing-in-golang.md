首发于：https://studygolang.com/articles/21197

# Go 语言中的快速检查测试

本系列的前文： https://itnext.io/types-and-specifications-c4d34ade6d5c

“我们只能看到我们了解的东西。”—— [Goethe](https://en.wikipedia.org/wiki/Johann_Wolfgang_von_Goethe)

在我的上一篇关于测试的文章里，我通过使用 Clojure 这门非静态类型的语言介绍了快速检查的概念。尽管我说要优先考虑基于规范的检查而不是指望类型系统能够确保程序的正确性，但你可能会发现自己不得不使用静态类型语言。融合了过程式编程和函数式编程的 Go 是我最喜欢的静态类型语言之一。Go 原生支持快速检查，但是使用的方式却和“ spec/check ”很不一样，而且 Go 的快速检查也不是 Erlang 或 Haskell 那样的纯函数式实现。尽管如此，Go 的快速检查还是让开发者能够在工作中进行许多种模糊测试，只要有足够的时间和创造性就能够进行强大的生成测试。

## 测试框架：`goConvey`

在我开始展示代码之前我想先向你们介绍 `goConvey`，我最喜欢的 Go 测试框架之一。尽管 Go 的原生测试框架很棒，但它是过程式的，可能会导致需要在 `t.Fail` 和 `t.Fatal()` 之间编写很多条件逻辑；尽管 Go 的原生测试框架对小项目来说足够用了，但是 `goConvey` 可以进行更密集的测试，而当项目变大时这些密集的测试可以起到更好的注释的作用。除此之外，`goConvey` 可以不严格按照条件操作函数，这允许我们进行复合和数据驱动的测试。

虽然这在我的文章《Play the Wrong Game》中提到过，但我还是要再说一遍，`goConvey` 也支持 TDD 和代码覆盖，这可以通过启动一个起到持续构建集成作用（不是很确定应该怎么翻译 continuous build integration ）的本地服务器来实现。然而，我建议你不要急着进行代码覆盖，因为只需通过浏览器上对后台构建 / 测试结果的自动反馈就可以创造出持续的工作流。详情请看：http://goconvey.co/

## 工欲善其事，必先利其器

`testing`/`quick` 提供了两个基本的工具：`Check` 和 `CheckEqual`。这些函数接受一个或者两个函数，还有一个配置项。`Check` 是 `CheckEqual` 的简化版，但是它接受的第二个函数总是返回 `true`。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/quickcheck-testing/0_2qezPmGmV5CZQ2PB.png)

虽然快速检查在它能够发挥作用的情况下总是好的，但是我们使用它的原因是找到 bug，为了找到 bug，我们需要有关产生 bug 的输入的反馈。这种信息在 `CheckError` 和 `CheckEqualError` 接口中被捕获。对输出的 `Error` 进行转换可以让我们看到调用的次数、输入（输出形式为接口切片）以及在发生 `CheckEqualError` 的情况下的不匹配的输出（输出的形式同样为接口切片）。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/quickcheck-testing/0_HU2q0vgYOV4_NOG-.png)

## 享受快速检查的便利

虽然上面提到的例子都很简单，但这是参数的缘故，现在让我们做些具有实际意义却很直白的事情。通过 `CheckEqual`，一个纯函数可以很容易得和这个函数的简化版进行比较。我想继续说回日期转换的话题，不仅是因为这个话题很有趣，而且也是因为我们似乎可以有无数种方式来做这件事。下面我们将会对基于正则表达式的转换和 `time` 包里内置的转换进行对比，以此来展示快速检查如何奇妙地为我们做完所有的工作并找出正则表达式中的一个 bug：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/quickcheck-testing/0_aqOo4T0mCxx_4Ybi.png)

查看一下 `goConvey`，我知道这并不像我们希望的那样简单明了：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/quickcheck-testing/0_e78RWmkJk_tblrJO.png)

静态测试显示有 bug，但是快速检查并不能找到具体的字符串，而我们希望能它找到一个可以让我们无需额外的帮助就能发现错误的字符串。增加迭代的次数或许最终可以找到无效的字符串，但是考虑到一个可变长的字符串有无数种可能，多做几次迭代也是无用的。

## 限制随机性

我们需要做的就是帮助快速检查创建这些输入的值，就像我们使用 Clojure spec 时做的那样。要实现这一点可以有几种选择。如果输入的值不是简单类型，那么可以自己实现 `Generate` 接口，或者是直接为 `quick` 实现“ Values ”函数。在这个例子中我将选择使用后一种方法。

我想要创建一些测试用例，在这些用例中 RFC3339 规范会有一些变化。每个字符串都有相同的首部，但是会有一些时区相关的格式，并可能带有毫秒数等。我生成了随机但有效的日期格式，然而每个日期中都带有“垃圾”，以此掩饰我阅读 RFC 规范时的不仔细。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/quickcheck-testing/0_JPnHMZNm9sHCZeOw.png)

定义 `TimeValue` 函数之后我们可以将其赋给 `quick.Config` 的 `Values` 字段，并设置一个更大的迭代次数上限，这样我们就更有把握找到出错的原因：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/quickcheck-testing/0_Xk1oVRFyuZULRU1Q.png)

就如我们希望的那样，快速检查为我们找到了一个无效的格式：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/quickcheck-testing/0_pfXTE4R4w9hyuTmu.png)

当然日期部分是不相关的，但是 Zulu/GMT 时间少了“ Z ”这一问题很快就被发现了，那么我就可以根据这点来修改 bug。现在可以更进一步，或许是整合更多的 unicode 字符，或许是修改格式，但从整体而言这都显示了我们可以怎样检查一个函数。

## 输入序列

纯函数的生成测试是很有用的，它可以发现单元测试有可能忽略的边界情况。然而，如果能够用快速检查对有状态的系统进行测试，更复杂的 bug 就有可能被发现。这些类似于“组件”甚至是“整合”层次上的测试，Go 语言为我们提供了一种进行这种测试的方法。

假设我们要测试一个 API。有许多可用的结点 / 代码段，我们想要找出在执行任意的命令序列之后可能出现的问题。我们想保证不会有命令序列会导致数据库崩溃。下面的“ API ”可以演示这一点：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/quickcheck-testing/0_-YhQN1xEp-WzUmyE.png)

有一个 handler，还有一些在依次执行事件 1 到 6 之后会发生的状态，这些状态会产生 `error`。你可以把这些整数看成是存储了命令的 `map` 的键，但因为参数很简单，所以这是一个简单的直接接受整数值的 API。我们还需要一些辅助函数来进行整合：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/quickcheck-testing/0_QMvvAPRtFe3IDsDC.png)

第一个函数可以快速地求余数，并且会正确处理负数，所以我们可以使用快速检查会默认给我们的 `int64` 类型的任意切片。第二个函数接受一个 API，按照给定的顺序执行事件，并在 API 发生错误的时候返回 `false`。通过这些工具，我们可以使用快速检查找到会破坏 API 的事件序列。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/quickcheck-testing/0_DhsByh96G1pu9krC.png)

下面的测试结果向我们报告了其中一个会破坏 API 的序列：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/quickcheck-testing/0_-dUy1YwExiCmXblT.png)

这不是理想的解决方案（真正的快速检查会将可能的输入集合减到最小，但是 Go 并没有“免费”提供这个功能），不过这是在测试过程中破坏 API 的方案之一。

尽管这种对快速检查的实现方法没有尽最大努力将可能的失败情况最小化，但是它也确实能让我们定制纯函数的模糊测试，而且就像上面提到的，可以在和 API、外部数据库等系统打交道时对有状态的系统进行更多的测试。示例代码链接：

https://github.com/weberr13/Kata/blob/master/gogen/gen_test.go

仓库中的代码遵循 MIT license。

本系列的下一篇文章：https://itnext.io/gopter-property-based-testing-in-golang-b36728c7c6d7

---

via: https://itnext.io/quickcheck-testing-in-golang-772e820f0bd5

作者：[Robert Weber](https://itnext.io/@robert_70579)
译者：[maxwellhertz](https://github.com/maxwellhertz)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

