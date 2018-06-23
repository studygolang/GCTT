已发布：https://studygolang.com/articles/12976

# 纹章生成器（Heraldry Generation）第二部分：Go, 背景区域和符号

发布于 2018-03-21 星期三

这次的迭代有一些激动人心的改变。 特别是我用 Go 重写的整个程序。实际上我在这块没有花太多时间。这样带来的结果就是程序有点乱，虽然如此，我还是花几分钟来介绍下。首先，我来讲讲纹章的改变。

在最进一次迭代中只用 fess 和 pale （译者注：纹章学中的专业术语，标识徽章上面的竖线和横线图案，[详看维基](https://en.wikipedia.org/wiki/Pale_(heraldry))） 两部分，之后我添加了一些其余的基本元素：bend，bend sinister（专业术语，[维基解释](https://en.wikipedia.org/wiki/Bend_(heraldry))），saltire（叉叉 [维基解释](https://en.wikipedia.org/wiki/Saltire)） ，和 chevron([维基解释](https://en.wikipedia.org/wiki/Chevron_(insignia)))

我还添加了最基本的元素 —— 这些几何图形叫做“Ordinary”（专业名词 [维基解释](https://en.wikipedia.org/wiki/Ordinary_(heraldry))）。并不是所有的普通 “Ordinary” 都包含在内，而且它看起来有点歪。因为我是手动创建并尝试错误修正，并没有进行准确的计算， 所以这还有很多问题需要我去解决。

例如，这里有一对图案来自当前的迭代程序。注意他们是否有点不对劲？
![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/heraldry-generation/01.jpg)

另一个问题是很多生成出来的设计视觉上的效果不是很好。将来很有可能进行一个迭代来改善这个问题。
说了这么多，它也产生了相当不多的效果。这是一些例子：
![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/heraldry-generation/02.jpg)

好了，现在让我们来聊一聊 Go。

我一直想把这个程序用 Go 重写；我使用了 PHP 这种快速成型的语言。然而，我遇到了阻碍，开发上没有更多的进展。我使用的 SVG 库不支持我需要的特性。我可以拆封那个库，处理原始 SVG 的 XML，但这样有些麻烦了。

然后我发现了一个优秀的 SVG Go 包，功能丰富，叫做 SVGo。它有几个我需要的功能。然后，我仅仅做了下简单改动，花了一个早上的时间来重写整个代码。

带来的结果就是程序更快且更容易编写；

不过这并不是说就没有问题了。

所有的代码都在一个大文件中。可视化的内联元素通过大块的 `switch` 或 `if` 来控制。没有单元测试（已经坑了我好几次了）。

现在 我将要添加更多复杂的符号（像动物的那种），我不能把所有的东西都放到一个文件中。我希望在下个迭代版本中去解决这个问题。

---

via:https://www.benovermyer.com/post/heraldry-pt-2/

作者：[Ben Overmyer](https://www.benovermyer.com/)
译者：[zhucheer](https://github.com/zhucheer)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
