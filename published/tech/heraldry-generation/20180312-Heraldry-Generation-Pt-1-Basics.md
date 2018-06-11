已发布：https://studygolang.com/articles/12975

# 纹章生成器（Heraldry Generation）第一部分：基础

发布于 2018-3-12，星期一

不久之前，我开始在一个随机模拟环境构造器项目上工作，而纹章构造器只是该项目的最初的一部分。

它遵循构造器设计的部分原则：每一个部分只做一件事，做好，并且在整个项目的生态系统中能被共享。

所以在这个原则的指导下，徽章生成器诞生了：输出可以被编程的方式修改、能用必要的元数据给人类或机器描述内容的通用格式图像。

我选定 SVG 作为输出图片格式，主要是因为它是 XML 格式，很容易修改，相对来说体积更小，而且它是矢量图，在缩放的情况下不会失帧。

为了快速地开发，我选择使用 PHP 来完成这个生成器，如果将来需要的话，我会用更高效的语言来重写它，就目前来说，它已经满足我的需求。

最初，我使用一个已经存在的库来处理 SVG ,但当我完成初始版本的时候，我发现这个库暂时还不支持 `<msk>` 标签，这意味着我无法将文件剪切成细小的部分。

在第二个迭代版本，我移除了了这个库，直接在 XML 文件上进行处理，最后的效果比较好，支持任意属性的操作。

纹章的生成坚持必要的原则，盾型的徽章基本上如同现在的路标，目的是清楚地指明方向。于是我为这个构造器假定了一些规则：

 1. 如果盾徽的寓意物是金属，那么它的主背景将不会是金属，同样的也不会有颜色；
 2. 只考虑欧洲10世纪-16世纪的背景和寓意物元素；
 3. 徽章将包含该时代的色彩、金属、皮草等元素，英格兰的“血色”着色将不会考虑。

目前来说这是指导性的规则，后期我可能会修改它。

目前，这个构造器只会生成基本背景的盾徽，它只能是一种颜色，或者上下对称、左右对称两种颜色，金属将会暂时保留用于寓意物上。

下面是一个简单的例子：

![](https://d33wubrfki0l68.cloudfront.net/152773ef00fd7c1f177ad2de4908fa49cd8c32e3/a28dc/heraldry-basic-field.svg)

背景构造器将会被设计成更高的层次，以至于背景的设计将会用于整个构造器的设计，这样它违背了第一个原则。

然而，在颜色的分割上出现了同样的颜色（例如蓝色接着蓝色），在下个版本中我会修复这个问题。

例如下面的这个图片跟上一张看上去一样，但是实际上是有区别的，这是对称的颜色相同造成的。

![](https://d33wubrfki0l68.cloudfront.net/052c829f6fb97c432e14e3dd1fb207241e2a3b11/84fda/heraldry-duplicate-colors.svg)

到目前为止，基本的设计是合理的，下面这是一个上下对称的盾徽例子：

![](https://d33wubrfki0l68.cloudfront.net/61eef0ee0bb1590dcd0e03aa22d61f686ef6203c/19b46/heraldry-basic-division.svg)

你能在[GitHub](https://github.com/ironarachne/heraldry)上找到这个构造器的源代码。

---

via: https://www.benovermyer.com/post/heraldry-pt-1/

作者：[Ben Overmyer](https://www.benovermyer.com/page/about/)
译者：[M1seRy](https://github.com/M1seRy)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

