# Go：使用 Ebiten 在 2D 视频游戏中进行图像渲染

![Ebiten](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/illustration.png)
插图由创作原始 Go Gopher 作品的 Renee French 为“ Go的旅程”创作。

*本文基于 Ebiten 1.10。*

[Ebiten](https://ebiten.org/) 是由 [Hajime Hosh](https://github.com/hajimehoshi) 用 Go 语言编写的成熟的 2D 游戏库。它是 Apple Store 上一些手机游戏如 [Bear's Restaurant](https://daigostudio.com/bearsrestaurant/en/) 或桌面游戏如 [OpenDiablo2](https://github.com/OpenDiablo2) 的引擎，OpenDiablo2 是 Go 版本暗黑2 的开源实现。现在，让我们深入了解电子游戏中的一些基本概念以及它们在 Ebiten 中的实现。

## 动画制作
在电子游戏世界中，Ebiten 通过分离的静态图像来渲染动画。这些图像集合被组合成一个更大的图像，通常称为“[纹理图集](https://en.wikipedia.org/wiki/Texture_atlas)”，也称为“精灵图”。这是 [网站](https://ebiten.org/examples/animation.html) 上提供的示例：

![image1](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/image1.png)

然后，经过简单的数学运算，这张图被加载到内存中，然后被一部分一部分地渲染。在前面的示例中，每个部分的宽度为 32 个像素。渲染每个图像非常简单，只需将 X 坐标移动 32 个像素。这是第一步：

![image2](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/image2.png)

![image3](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/image3.png)

渲染动画时，只需要当前图像的值即可渲染精灵图的正确部分。Ebiten 提供了所有的 API 来轻松地渲染它，下面是之前第一步中精灵图的示例：

```go
screen.DrawImage(
   // 在坐标上绘制子画面的子图像
   // x=0 to x=32 and y=32 to y=64
   runnerImg.SubImage(image.Rect(0, 32, 32, 64)).(*ebiten.Image),
   // 在该位置之前声明的变量定义了屏幕上的位置
   op,
)
```

Ebiten 的创建者 Hajime Hosh 还开发了“ [file2byteslice](https://github.com/hajimehoshi/file2byteslice)”，该工具可将任何图像转换为字符串，并允许将任何文件嵌入到 Go 中。它可以通过注释 `go:generate` 与 Go 工具轻松集成，自动将图像转储到 Go 文件中。下面是一个例子：

```go
//go:generate file2byteslice
  -package=img
  -input=./images/sprite.png
  -output=./images/sprite.go -var=Runner_png
```

现在可以从以下代码的 `img.Runner_png` 变量中获取精灵图：

![image4](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/image4.png)

然后，必须先对图像进行解码，然后再由 Ebiten 加载并在游戏中进行渲染。让我们转到另一种渲染较大背景图像的方法，该图像由循环元素组成。

## 瓷砖背景

渲染背景使用相同的技术，将主要地图集分为许多小图像，称为“瓷砖”。这是网站上提供的纹理图集的示例：

![tiles1](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/tiles1.png)

该图集可以被 16 像素的图块分割 ，这是使用 [Tiled](https://www.mapeditor.org/) 软件创建的图块集合：

![tiles2](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/tiles2.png)

每个图块将分配一个数字，从 0 开始，逐次加 1。由于每行都有相同数量的图块，因此可以通过除法和取模来获取图块的坐标。这是蓝色花朵的示例：

![tiles3](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/tiles3.png)

蓝色花朵的编号为 303，这意味着它们位于第 4 列（303 以每行的图块数为模，例如 303％25 = 3）和第 13 行（303 除以每行的图块数，例如，303/25 = 12）。

现在，我们可以使用索引数组来构建地图：

![map_array](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/map_array.png)

从这张地图上绘制图像可以得到主要装饰：

![bg1](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/bg1.png)

但是缺少背景。我们必须构建一个表示背景的类似数组。结果如下：

![bg2](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/bg2.png)

现在，这些图层已经准备好了，我们只需要迭代两个图层就可以得到最终结果：

![bg3](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/bg3.png)

此示例中的代码可在 [Ebiten 网站](https://ebiten.org/examples/tiles.html) 上找到。

生成图像后，Ebiten 必须管理屏幕更新并将指令发送到图形卡片。

## 屏幕更新

Ebiten 提供了 iOS（使用 Metal）和 Android（使用 OpenGL ES）使用的驱动程序的抽象，这使开发更加容易。它还允许您定义一个函数，该函数可以更新屏幕并绘制所有更改。但是，出于性能原因，该库会将源打包在一起，并将这些更改存储在缓冲区中，然后再发送给驱动程序：

![screen_update1](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/screen_update1.png)

该缓冲区还能够合并绘图指令，以减少对 GPU 的调用次数。在上图中，这三个指令现在可以合并为一个：

![screen_update2](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/screen_update2.png)

这项改进很重要，因为它可以减少发送指令时的开销。这是合并指令的性能：

![perf1](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/perf1.png)

通过逐个发送指令，性能会大大降低：

![perf2](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/perf2.png)

Ebiten 还提供了对屏幕刷新的控制，这有助于调整性能。

## TPS 管理

Ebiten 的默认 TPS 是 60。但是，这可以用 Ebiten 提供的 API `ebiten.SetMaxTPS()` 来轻松配置。它有助于减少机器上的压力。这是具有 25 TPS 的简单程序：

![tps1](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/tps1.png)

TPS（每秒 tick 数）与 FPS（每秒帧数）不同。[Hajime Hosh](https://github.com/hajimehoshi) 很好地描述了这些差异：

> 帧代表图形更新。这取决于用户显示屏上的刷新率。那么 FPS 可能是 60、70、120，依此类推。这个数字基本上是不可控制的。Ebiten 可以打开或关闭 vsync。如果关闭了 vsync，则 Ebiten 会尝试尽可能多地更新图形，那么 FPS 可以为 1000 左右。
> tick 表示逻辑更新。TPS 表示每秒调用更新功能的次数。默认情况下固定为 60。游戏开发人员可以通过 SetMaxTPS 配置 TPS。如果设置了 UncappedTPS，则 Ebiten 会尝试尽可能多地调用更新函数。

当窗口在后台运行时，Ebiten 为渲染带来了另一种优化。失去焦点时，游戏将进入休眠状态，直到重新获得焦点。它实际上 sleep 了 1/60 秒，然后再次检查焦点。这是保存的资源的示例：

![tps2](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200205-Go-Image-Rendering-in-2D-Video-Games-with-Ebiten/tps2.png)

尽管可以关闭此优化，但是运行的系统可能会限制游戏的运行。例如，在 [Firefox](https://hacks.mozilla.org/2018/01/firefox-58-the-quantum-era-continues/)，[Chrome](https://developers.google.com/web/updates/2017/03/background_tabs) 或其他浏览器上的后台标签中运行时，在浏览器中运行的游戏会受到限制。

Ebiten 在 [Github](https://github.com/sponsors/hajimehoshi) 上接受赞助。如果您希望看到更多 Go 编写的游戏，请随时贡献力量。

---
via: https://medium.com/a-journey-with-go/go-image-rendering-in-2d-video-games-with-ebiten-912cc2360c4f

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[alandtsang](https://github.com/alandtsang)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出