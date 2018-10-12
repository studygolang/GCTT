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

让我们看一下在mongodb中数据是如何存储和维护的：

```go
// Polygon defines a set of points that complete a ring
// around a geographic area
type Polygon [][2]float64

// PolygonRings defines a MongoDB Structure for storing multiple polygon rings
type PolygonRings struct {
    Type string           bson:&quot;type&quot;
    Coordinates []Polygon bson:&quot;coordinates&quot;
}

// Represents a marine station and its polygons
type MarineStation struct {
    StationId string      bson:&quot;station_id&quot;
    Polygons PolygonRings bson:&quot;polygons&quot;
}
```
多边形类型表示2个浮点数的一个切片。这将表示构成多边形的每个点。

如果你要是想通过mongodb来执行不同区域的地理空间搜索，那么在mongodb中存储多边形接口是必须的。

一个海洋预报区域当成一个与很多多边形有关系的站。

测试代码将创建一个带有两个多边形的站。然后它会显示一切。让我们来看看如何用slice创建一个海洋站，并创建一个单一的海洋站进行测试：

```go
// Create a nil slice to store the polygon rings
// for the different marine stations
var marineStations []MarineStation

// Create a marine station for AMZ123
marineStation := MarineStation{
    StationId: "AMZ123",
    Polygons: PolygonRings{
        Type: "Polygon",
        Coordinates: []Polygon{},
    },
}
```

第一行代码创建了一个存储海洋预报区域的空切片。然后我们用复合字面量的方式的方式创建了一个海洋预报区域对象。

 Within the composite literal we have another composite literal to create an object of type PolygonRings for the Polygons property
 在这个复合字面量中我们需要为这个PolygonRings创建另一个复合字面量对象Polygons。
 Then within the creation of the PolygonRings object we create an empty slice that can hold Polygon objects for the Coordinates property.
 






---

via: https://www.ardanlabs.com/blog/2013/09/slices-of-slices-of-slices-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[xmge](https://github.com/xmge)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
