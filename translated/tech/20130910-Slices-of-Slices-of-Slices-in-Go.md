## https://www.ardanlabs.com/blog/2013/09/slices-of-slices-of-slices-in-go.html

# go的多层切片
William Kennedy 2013.9.10

我在美国用程序为不同的地区的海洋天气预报做展示。这些多边形数据需要存储在mongodb中，而且还需要特殊的方式来处理。如果不是因为每个地区不只有一个多边形，那也没什么难的。在外部有个多边形，在其内部有0-n个多边形，并且它们之间需要维护一定的关系。

看了一会儿问题之后，我意识到我需要创建一个海洋预报区域的切片，每个区域包含一个多边形切片。为了存储每个多边形环，我需要一个地理坐标切片。最后，每个坐标需要存储在二维的float数组中。

一张图片胜过千言万语:

![polygons](https://www.ardanlabs.com/images/goinggo/Screen+Shot+2013-09-04+at+5.02.55+PM.png)

存储在mongodb中的数据应该是如下格式的:

![pattern](https://www.ardanlabs.com/images/goinggo/Screen+Shot+2013-09-10+at+3.46.19+PM.png)

只是看着图表和图片我就晕了。该图描述了如何将切片和对象组合在一起。

图中显示了多边形是如何在MangoDB存储的。在坐标下会有多个元素，而每个元素又都有它自己的点集。

我决定写一个测试程序来想出如何构造和存储这些数据。

用的slice越多我就越喜欢它们。我喜欢它可以作为函数的参数或者返回值而我不用亲自去处理应用和内存。切片是一种可以在方法中安全的传入或者返回的轻量级的数据结构。

我一直在想，我需要传递是切片的引用，这样就不会在堆栈上复制数据结构。我记得，数据结构是24字节，我不需要复制抽象层次较低一级的所有数据。

阅读下面两篇文章可以学习到更多关于slice的知识

http://www.goinggo.net/2013/08/understanding-slices-in-go-programming.html<br>
http://www.goinggo.net/2013/08/collections-of-unknown-length-in-go.html


---

via: https://www.ardanlabs.com/blog/2013/09/slices-of-slices-of-slices-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[xmge](https://github.com/xmge)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
