首发于：https://studygolang.com/articles/20172

# Go 的人脸识别教程 - 第一部分

整个人脸识别领域是我喜欢阅读的内容。自己实现一个人脸识别系统会让你听起来像 Tony Stark 并且你可以将它们用于各种不同的项目，例如自动锁门，或者为你的办公室建立一个监控系统，仅举几例。

在此教程中，我们将使用 Go 中一些内置的库去构建自己的，非常简单的人脸识别系统。我们首先在静态图像上进行简单的人脸识别，并且看看它是如何工作的，然后我们将在这个小教程的第二部分拓展到成视频输入的实时人脸识别。

---

## 视频教程

此教程以视频格式，如果您希望支持我和我的频道，请点赞并订阅我的频道！

---

## Kagami/go-face 包

在这个基础教程中，我们将使用封装了 dlib 机器学习工具包的 [kagami/go-face 包](https://github.com/Kagami/go-face)。

> 注意： Kagami 实际上写了他如何写这个包。这一定是一个有趣的文章并且你可以在这里找到它：https://hackernoon.com/face-recognition-with-go-676a555b8a7e

---

## dlib 工具包

它是用 C++ 构建的，它在面部识别与检测方面都令人惊奇。根据它的文档，在 Wild 基准测试中检测标记面部的准确度约为 99.4%，这难以置信，也是为什么许多第三方库将其作为基础的原因

> 注意：我在前面的教程中介绍了 Dlib 工具包的 Python 库 - face_recognition. 如果你想查看此教程的 Python 实现，在这里：[介绍 Python 的人脸识别](https://tutorialedge.net/python/intro-face-recognition-in-python/)

---

## 安装

我不想撒谎，让它启动并运行起来比用标准的 Go 包更痛苦一些，你需要在计算机上安装 `pkg-config` 和 `dlib`。如果你在 MacOS 上运行，命令如下：

```
$ brew install pkg-config dlib
$ sed -i '' 's/^Libs: .*/& -lblas -llapack/' /usr/local/lib/pkgconfig/dlib-1.pc
```

---

## 开始

我们首先需要下载 `kagami/go-face 包`，可以使用如下 `go get` 命令：

```
$ go get -u github.com/Kagami/go-face
```

在你的 GOPATH 目录中创建一个名为 `go-face-recognition` 的新目录。在此目录中创建一个名为 `main.go` 的文件，这是我们所有源码所在位置。

完成操作后，你需要从 TutorialEdge/ go-face-recognition-tutorial repo 的 `image/` 目录中获取文件。最简单的方法是 clone repo 到另一个目录，只需复制图片目录到当前的工作目录

```
$ git clone https://github.com/TutorialEdge/go-face-recognition-tutorial.git
```

一旦成功 clone 后，我们就有了启动人脸识别的所需的 `.dat` 文件。你应该还看到一些其他包含复仇者联盟面孔的 `.jpg` 文件列表。

```go
package main

import (
    "fmt"

    "github.com/Kagami/go-face"
)

const dataDir = "testdata"

func main() {
    fmt.Println("Facial Recognition System v0.01")

    rec, err := face.NewRecognizer(dataDir)
    if err != nil {
        fmt.Println("Cannot INItialize recognizer")
    }
    defer rec.Close()

    fmt.Println("Recognizer Initialized")
}
```

好的，如果我们尝试在这时候运行程序，应该会在我们程序的输出中看到  `Facial Recognition System v0.01` 和 `Recognizer Initialized`. 我们已经成功的设置了所需的一切，为了做一些很酷的高级人脸识别！

---

## 计算图片中的面孔

我们对这个 package 的第一个真正测试是我们是否能准确地计算照片中的面孔数量。出于本教程的目的，我们将使用此照片：
![tony-stark](https://raw.githubusercontent.com/studygolang/gctt-images/master/face-recognition/tony-stark.jpg)

如你所见，没什么特别的，只有 Tony Stark 那张孤独的脸。

因此，我们现在需要拓展已有的程序，以便能分析此图像，然后计算所述图像中的面孔数量：

```go
package main

import (
    "fmt"
    "log"
    "path/filepath"

    "github.com/Kagami/go-face"
)

const dataDir = "testdata"

func main() {
    fmt.Println("Facial Recognition System v0.01")

    rec, err := face.NewRecognizer(dataDir)
    if err != nil {
        fmt.Println("Cannot INItialize recognizer")
    }
    defer rec.Close()

    fmt.Println("Recognizer Initialized")

    // we create the path to our image with filepath.Join
    avengersImage := filepath.Join(dataDir, "tony-stark.jpg")

    // we then call RecognizeFile passing in the path
    // to our file to retrieve the number of faces and any
    // potential errors
    faces, err := rec.RecognizeFile(avengersImage)
    if err != nil {
        log.Fatalf("Can't recognize: %v", err)
    }
    // we print out the number of faces in our image
    fmt.Println("Number of Faces in Image: ", len(faces))

}
```

当我们运行它时，应该会看到以下输出：

```
$ go run main.go
Facial Recognition System v0.01
Recognizer Initialized
Number of Faces in Image:  1
```

太棒了！我们已经能够分析并确认一张图像中包含了一张人脸。让我们尝试一个包含更多复仇者的图像：![Avengers](https://raw.githubusercontent.com/studygolang/gctt-images/master/face-recognition/avengers-01.jpg)

当我们更新第 24 行时：

```go
avengersImage := filepath.Join(dataDir, "avengers-01.jpg")
```

重新运行我们的程序，应该会看到我们的程序能够确认 2 个人在这个此图像中。

---

## 人脸识别

很好，我们可以计算一张图片中有多少张面孔，现在来实际地确认这些人是谁？

为此，我们需要一些参考照片，例如，如果我们想要从照片中识别 Tony Stark，我们需要带他名字的样例照片。识别软件能分析与他相识的人脸照片，然后将其匹配。

因此，让我们用 `avengers-02.jpg` 作为 Tony Stark 的参考照片，然后看看是否能识别这张图像包含他的脸
![tony-stark](https://raw.githubusercontent.com/studygolang/gctt-images/master/face-recognition/tony-stark.jpg)

```go
avengersImage := filepath.Join(dataDir, "avengers-02.jpeg")

faces, err := rec.RecognizeFile(avengersImage)
if err != nil {
    log.Fatalf("Can't recognize: %v", err)
}
fmt.Println("Number of Faces in Image: ", len(faces))

var samples []face.Descriptor
var avengers []int32
for i, f := range faces {
    samples = append(samples, f.Descriptor)
    // Each face is unique on that image so Goes to its own category.
    avengers = append(avengers, int32(i))
}
// Name the categories, i.e. people on the image.
labels := []string{
    "Dr Strange",
    "Tony Stark",
    "Bruce Banner",
    "Wong",
}
// Pass samples to the recognizer.
rec.SetSamples(samples, avengers)
```

因此，在上面的代码中，我们已经从左到右检查了面孔并标记了合适的名字。我们的识别系统能使用这些参考样例来尝试对对后续文件进行自己的人脸识别。

让我们尝试用已有的 Tony Stark 图片测试我们的识别系统，然后看看是否能根据 avengers-02.jpeg 文件生成的人脸描述符来识别它：

```go
// Now let's try to classify some not yet known image.
testTonyStark := filepath.Join(dataDir, "tony-stark.jpg")
tonyStark, err := rec.RecognizeSingleFile(testTonyStark)
if err != nil {
    log.Fatalf("Can't recognize: %v", err)
}
if tonyStark == nil {
    log.Fatalf("Not a single face on the image")
}
avengerID := rec.Classify(tonyStark.Descriptor)
if avengerID < 0 {
    log.Fatalf("Can't classify")
}

fmt.Println(avengerID)
fmt.Println(labels[avengerID])
```

现在尝试验证这不是侥幸，并试着看看我们的图像识别系统是否适用于 Dr Strange 。

![dr-strange](https://raw.githubusercontent.com/studygolang/gctt-images/master/face-recognition/dr-strange.jpg)

```go
testDrStrange := filepath.Join(dataDir, "dr-strange.jpg")
drStrange, err := rec.RecognizeSingleFile(testDrStrange)
if err != nil {
    log.Fatalf("Can't recognize: %v", err)
}
if drStrange == nil {
    log.Fatalf("Not a single face on the image")
}
avengerID = rec.Classify(drStrange.Descriptor)
if avengerID < 0 {
    log.Fatalf("Can't classify")
}
```

最后，让我们试着用 Wong 的图片：

![wong](https://raw.githubusercontent.com/studygolang/gctt-images/master/face-recognition/wong.jpg)

```go
testWong := filepath.Join(dataDir, "wong.jpg")
wong, err := rec.RecognizeSingleFile(testWong)
if err != nil {
    log.Fatalf("Can't recognize: %v", err)
}
if wong == nil {
    log.Fatalf("Not a single face on the image")
}
avengerID = rec.Classify(wong.Descriptor)
if avengerID < 0 {
    log.Fatalf("Can't classify")
}
fmt.Println(avengerID)
fmt.Println(labels[avengerID])
```

当你一起运行时，你应该会看到以下输出：

```
$ go run main.go
Facial Recognition System v0.01
Recognizer Initialized
Number of Faces in Image:  4
1
Tony Stark
0
Dr Strange
3
Wong
```

太棒了，我们已经构建了一个真正简单的人脸识别系统，让我们能识别不同的复仇者。

> 挑战：在所有复仇者上建立一些参考文件，并尝试将人脸识别代码提取到一个可重用的函数中

---

## 完整源代码

可在 Github: Tutorialedge/go-face-recognition-tutorial 中找到本教程完整的源代码

---

## 结论

在此教程中，我们成功地构建了一个十分简单的人脸识别系统，能在静态图像上运行。这个系列教程有望成为下一部分的基础，我们将在研究如何如何在视频流的实时上下文中进行识别。

希望你喜欢本教程，如果您这样完成了，请在下面的评论部分告诉我们！

> **注意**：如果你想实时追踪新的 Go 文章发布到网站的时间，那么请在 Twitter 随时关注我所有的最新消息：[**@Elliot_F**](https://twitter.com/elliot_f).

---

via: https://tutorialedge.net/golang/go-face-recognition-tutorial-part-one/

作者：[Elliot Forbes](https://tutorialedge.net/about/)
译者：[piglig](https://github.com/piglig)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
