## https://www.ardanlabs.com/blog/2013/09/slices-of-slices-of-slices-in-go.html

# go的多层切片

William Kennedy 2013.9.10

我是一名程序员，我的工作内容是为美国不同的地区的海洋天气预报结果做展示。这些区域数据需要存储在mongodb中，而且还需要特殊的方式来处理。如果不是因为每个地区不只有一个多边形数据结构，那也没什么难的。但是数据结构是这样的，一个大的多边形区域的数据结构，包含着0-n个多边形，并且它们之间需要维护一定的关系。

看了一会儿问题之后，我意识到我需要创建一个海洋预报区域的切片，每个区域包含一个区域数据结构切片。为了存储每个区域环，我需要一个地理坐标切片。最后，每个坐标需要存储在二维的float数组中。

一张图片胜过千言万语:

![polygons](https://www.ardanlabs.com/images/goinggo/Screen+Shot+2013-09-04+at+5.02.55+PM.png)

存储在mongodb中的数据应该是如下格式的:

![pattern](https://www.ardanlabs.com/images/goinggo/Screen+Shot+2013-09-10+at+3.46.19+PM.png)

只是看着图表和图片我就晕了。该图描述了如何将切片和对象组合在一起。

图中显示了多边形是如何在MangoDB存储的。在坐标下会有多个元素，而每个元素又都有它自己的点集。

我决定写一个测试程序来想出如何构造和存储这些数据。

越用slice我就越喜欢它们。我喜欢它可以作为函数的参数或者返回值而我不用亲自去处理应用和内存。切片是一种可以在方法中安全的传入或者返回的轻量级的数据结构。

我一直在想，我需要传递是切片的引用，这样就不会在堆栈上复制数据结构。我记得，数据结构是24字节，我不需要复制抽象层次较低一级的所有数据。

阅读下面两篇文章可以学习到更多关于slice的知识：

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

多边形区域结构由长度为2的float数组切片组成。切片中每一个数组标示多边形区域结构中的一个点。

如果你要是想通过mongodb来执行不同区域的地理空间搜索，那么在mongodb中存储多边形区域数据的结构是必须的。

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

第一行代码创建了一个存储海洋预报区域的空切片。然后我们用复合字面量的方式的方式创建了一个海洋预报区域对象。在这个复合字面量中我们需要为这个PolygonRings创建另一个复合字面量对象Polygons。在创建的PolygonRings对象中我们为Coordinates字段创建空的slice来存储Polygon对象。

若要了解复合文本的更多信息，请查看此文档：

http://golang.org/ref/spec#Composite_literals
 
现在是时候向station添加几个区域数据结构：
 
```go
 // Create the points for the second polygon ring
point1 = [2]float64{-80.4370117189999, 27.7877197270001}
point2 = [2]float64{-80.4376220699999, 27.7885131840001}
point3 = [2]float64{-80.4384155269999, 27.7885131840001}
point4 = [2]float64{-80.4370117189999, 27.7877197270001}

// Create a polygon for this ring
polygon = Polygon{point1, point2, point3, point4}

// Add the polygon to the slice of polygon coordinates
marineStation.Polygons.Coordinates = append(marineStation.Polygons.Coordinates, polygon)
 ```
 
在第二个polygon中有4个点而不是5个，剩下的最后一件事就是将polygon加入到stations切片中，并且展示出来：

```go
// Add the marine station
marineStations = append(marineStations, marineStation)

Display(marineStations)
```

展示函数使用关键字`range`来进行遍历所有的切片。

```go
func Display(marineStations []MarineStation) {
    for _, marineStation := range marineStations {
        fmt.Printf("\nStation: %s\n", marineStation.StationId)

        for index, rings := range marineStation.Polygons.Coordinates {
            fmt.Printf("Ring: %d\n", index)

            for _, coordinate := range rings {
                fmt.Printf("Point: %f,%f\n", coordinate[0], coordinate[1])
            }
        }
    }
}
```

这个方法需要传入一个MarineStation切片作为参数。记住slice结构仅仅是在栈上被拷贝，并不是传入slice对象。

当我们迭代MarineStation对象和组成它的所有切片时，我们得到以下结果：

```
Station: AMZ123
Ring: 0
Point: -79.729119,26.972940
Point: -80.079953,26.969269
Point: -80.080363,26.970533
Point: -80.081051,26.975004
Point: -79.729119,26.972940
Ring: 1
Point: -80.437012,27.787720
Point: -80.437622,27.788513
Point: -80.438416,27.788513
Point: -80.437012,27.787720
```

使用slice去解决这个问题是快速的、容易的、高效的。我将这份测试代码复制了一份放在了[The Go Playground](https://play.golang.org/)。

http://play.golang.org/p/UYO2HIKggy

通过快速构建这个测试应用让我深深地感觉到slice具有很大的优点。它可以使你开发更高效，代码更健壮。你不必担心内存管理,你可以通过切片的引用在函数数据传递时传递较大的数据。花一些时间去学习如何在代码中使用slices，学完后你将感谢你自己。

---

via: https://www.ardanlabs.com/blog/2013/09/slices-of-slices-of-slices-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[xmge](https://github.com/xmge)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
