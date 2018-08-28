首发于：https://studygolang.com/articles/14480

# Go 程序在 macOS 上的打包功能

这篇文章简单地说明了如何在 macOS 上打包一个 Go 的程序，包括引用或者不引用外部资源。作为一个原生的 Cocoa 程序它是可以下载、安装和运行的。开发过程并不需要 XCode，cgo 或者其他特殊的库文件。

在此文章结尾处，你的 Go 程序将会被打包进一个 `.app` 文件，成为一个拖放安装的只读 DMG 文件。这也适用于其他非 Go 语言的程序。

关于这个端到端的解决方案的主题还有其他很多的指导文章，有的讲的比我想要做的内容多很多，有的又是采用不同的方式，而我想呈现给你的是如何将文件打包的过程。

**要求**：你必须要有一个 macOS。此方法不会在 Windows 或者 Linux 上生效。事实上，它也可能生效，但是你必须要手动或者借助某些疯狂的工具来制作和 macOS 相关的东西，比如 DMG 或者 `.DS_Store` 文件。

## 为你的 app 制作一个图标

你需要一个高分辨率的图标；最好漂亮点儿。需要在浅色背景和深色都表现优秀。[参见 Apple 的官方文档](https://developer.apple.com/library/content/documentation/GraphicsAnimation/Conceptual/HighResolutionOSX/Optimizing/Optimizing.html)。

为了最好的效果，将图标保存为 1024x1024 分辨率的透明 PNG 文件。

如下为一些示例：
![avatar](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-for-mac/1.png)

### 编译图标集

当准备好大图标之后，你需要将它保存成不同的尺寸和分辨率。为了获得更好的兼容性，推荐的尺寸列表为：16，32，64，128，256，512 和 1024；每一个尺寸都需要为高分辨率的场景（除了 1024）准备 `@2x` 的图标。如果没有一个图形化的程序帮你做这件事情，将会非常乏味和冗长；或者你可以利用如下所示的命令 (sips) 来做分辨率调整的工作：

```shell
$ sips -z $SIZE $SIZE myicon.png \
       --out myicon_${SIZE}x${SIZE}.png
```

以此类推。不要忘了 `@2x` 的批次，它们的尺寸的级别更高，但是会标注为当前的名称。这的确是很繁琐的工作，但这是可以自动化的，而且你只需要做一次（本文末尾链接了一个示例程序）。

接着，假设你的图标文件位于一个叫做 myicons 的文件夹呢，使用[iconutil](https://developer.apple.com/library/content/documentation/GraphicsAnimation/Conceptual/HighResolutionOSX/Optimizing/Optimizing.html#//apple_ref/doc/uid/TP40012302-CH7-SW2)来生成图标：

```shell
$ iconutil -c icns -o \
        icon.icns myicons.iconset
```

### 制作.app bundle

macOS 上可装载可运行的的程序仅仅是一个 .app 的文件夹，里面包含了一个二进制文件和清单文件。你甚至可以手动创建。它的目录结构类似于：

```shell
$ tree Caddy.app
Caddy.app/
└── Contents
    ├── Info.plist
    ├── MacOS
    │   └── caddy
    └── Resources
        └── icon.icns
3 directories, 3 files
```

关键文件包含：

+ Info.plist: 清单文件。可以从其他地方拷贝过来进行定制化
+ caddy: 你要打包的二进制文件
+ icon.icns: 打包好的不同尺寸的图标文件

（右键或者按住 Ctrl 键单击程序图标，选择"显示包内容"可以浏览 .app 的文件夹）

### 添加 Info.plist 文件

这里展示了一个清单文件，你可以根据自己的内容进行修改（[来自 Dmitri Shurylov](https://github.com/shurcooL/trayhost)，感谢原作者）

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>caddy</string>
    <key>CFBundleIconFile</key>
    <string>icon.icns</string>
    <key>CFBundleIdentifier</key>
    <string>com.example.yours</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <true/>
</dict>
</plist>
```

请确保替换了 [CFBundleExecutable](https://developer.apple.com/library/content/documentation/General/Reference/InfoPlistKeyReference/Articles/CoreFoundationKeys.html) 和 [CFBundleIdentifier](https://developer.apple.com/library/content/documentation/General/Reference/InfoPlistKeyReference/Articles/CoreFoundationKeys.html)。

注意：[LSUIElement](https://developer.apple.com/library/content/documentation/General/Reference/InfoPlistKeyReference/Articles/LaunchServicesKeys.html) 告诉操作系统你的程序是一个代理程序，所以他不会出现在任务栏。这非常重要，否则它将一直反复出现在任务栏中，除非你的 Go 程序设计成像 Cocoa 程序一样响应 Mac 事件循环中的事件。

恭喜你！你的 Go 程序已经有了一个打好包 `.app`。双击图标后它将会运行。你可以将它拖到应用程序文件夹，它会像一个 Cocoa 程序一样出现在你的启动器中。

## 制作 DMG 文件

DMG 文件是你分发程序的文件。它压缩了整个 `.app` 文件，可以很轻易的拖到应用程序文件夹并且安装。

此教程展示了手动的制作过程，后面的要点是告诉你如何自动化的做这件事情。

### 制作 DMB 模板

模板只需要制作一次。

打开磁盘工具。按 ⌘N 创建一个新的磁盘镜像。给它取个名字，配置好足以容纳你的程序包的空间大小。

![avatar](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-for-mac/2.png)

在 Finder 中找到挂载的镜像。定制化此文件夹的视图设置，使得它成为你想要用户实际安装时看到的样子。可以考虑设置一个背景图，隐藏工具栏和边栏，增加图标尺寸等。

注意：背景图必须包含在 DMG 中。通常会放在 `.background` 文件夹中。文件夹名字以点号开头，它会显示为隐藏文件夹。把背景图片放在这里，并拖放到视图选项中进行配置。

![avatar](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-for-mac/3.png)

为了方便起见，你可能需要为 DMG 的 /Application 文件夹创建一个快捷方式（别名）。你可以右键点击 /Application 文件夹，选择“创建别名”，并移动到挂载好的镜像中。

![avatar](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-for-mac/4.png)

现在我们可以把打包好的 app 加入到 DMG 中来完成视图的定制化：

![avatar](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-for-mac/5.png)

干得好！我们的模板 DMG 已经全部建好了。现在它已经可以用于分发了。

## 转换 DMG 文件以便于分发

当前的 DMG 没有经过压缩并且是可写的。这对于发布程序是不够的，所以我们必须做一下转换来修复。

打开磁盘工具，选择镜像->转换。给文件起一个有意义的名称，其他设置保持不变。（镜像格式应为“压缩”）

![avatar](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-for-mac/6.png)

瞧！你现在拥有了一个可以发布的压缩包。当你打开这个文件，把它拖动到应用程序创建快捷方式后，它会出现在启动器中：

![avatar](https://raw.githubusercontent.com/studygolang/gctt-images/master/package-for-mac/7.png)

“哦，很漂亮！”...除了有点儿单调。:)

## 未来的镜像制作

以后，你可以重用这个模板 DMG 和图标。所有痛苦的工作只需要做一次。为了再次发布，你只需要重新打包 .app （可以自动化），替换到模板 DMG 里（可以自动化）和重新转换 DMG 来分发（同样可以自动化）。

## 将过程自动化

上面介绍的一些一次性的工作可以自动化（比如创建图标库）。但是有些很难用自动化取得很好的结果（比如定制化 DMG 的视图）。幸运的是，除此之外的部分是很容易做到自动化的，包括每次重新打包一个新的发布版本所需要的步骤。hdiutil 这个命令可以帮助你来创建、挂在和转换镜像。

[这里](https://gist.github.com/mholt/11008646c95d787c30806d3f24b2c844)是我写的自动化的示例。它会创建一个打包的 .app，如果你给它模板 .dmg 它可以创建最终的 .dmg。它甚至封装了上文所述的创建图标的命令，可以帮你创建全部的不同尺寸的图标。它做了所有的拷贝、挂载、卸载和镜像转换。你可以对你不喜欢的部分或者它不能工作的地方（非常有可能发生，哈哈）做改动。它有很好的注释，很容易阅读。你只需要告诉它二进制文件在哪里、资源文件在哪里、二进制文件的名字、1024像素的图标文件和有意义的程序名字。

这不是一个会继续维护的开源项目。如果要使用它，你可以将它整合进实际的应用场景中。

我认为让更多的 Go 项目以 app 的形式运行是非常酷的事情。Go 代码非常美丽，为什么不能让编译出来的二进制文件也同样美丽呢？

---

via: https://medium.com/@mattholt/packaging-a-go-application-for-macos-f7084b00f6b5

作者：[Matt Holt](https://medium.com/@mattholt)
译者：[lebai03](https://github.com/lebai03)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
