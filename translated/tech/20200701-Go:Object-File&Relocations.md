# GO:对象文件&重定位

## 作者: Vincent Blanchon
## 时间: 2020-07-01
![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200701-Go:Object-File%26Relocations/1_HxAju6n33e9Y8AJwMuQL3w.png)

**本文章基于Go 1.14**

重定位是链接过程中的一个阶段，
重定位是链接过程中为每个外部符号分配适当地址。由于每个包都是单独编译的，因此它们不知道来自其它包的函数或者变量在哪里。 让我们从一个需要重定位的简单示例开始。

## 编译
以下程序涉及两个不同的程序包：main和fmt。
![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200701-Go:Object-File%26Relocations/1_4_DaAwHmqJbhwP8Tn10Dzg.png)

构建此程序将首先涉及编译器，该编译器分别编译每个包。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200701-Go:Object-File%26Relocations/1_4HLpept1qBXFJvL_r4qptQ.png)

通过命令
```
go tool compile -S -l main.go
```
我们可以查看指令在中间文件(译者注:目标文件)中的临时地址。

一旦我们的程序被编译，我们可以使用
```
go tool compile -S -l main.go
```
来查看程序对应的汇编代码

要查看编译器生成的指令，你有多种不同的方法:

* 重新编译，并打印汇编指令。命令是
```
go tool compile -S -l main.go
```
```
"".main STEXT size=137 args=0x0 locals=0x58
   0x0000 00000 (main.go:7)    TEXT   "".main(SB)
   [...]
   0x0058 00088 (main.go:8)    CALL   fmt.Println(SB)
```
参数 `-l` 用于避免内联，使得汇编代码更容易被阅读。

生成的汇编文件表明调用`Println`的指令相对`main`函数入口偏移88个字节。这个偏移对于链接器重新定位函数调用将会非常有用。

* 使用以下命令，反汇编已经生成的 main.o
```
go tool objdump main.o
```
```
TEXT %22%22.main(SB)
  [...]
  main.go:8       0x57e        e800000000    CALL 0x583    [1:5]R_CALL:fmt.Println
```
标识符`R_CALL`代表重定位调用

然而由于该函数属于另一个包，因此编译器不知道该函数实际位于何处。使用命令
```
go tool nm main.o
```
可以检查生成的文件 `main.o`，并列出其中包含的符号。下图是输出
![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200701-Go:Object-File%26Relocations/1__cz0Ozr4acR3Sj0GbirP2Q.png)

我们可以注意到，它需要使用go工具nm命令而不是本机nm命令。 实际上，Go生成的目标文件（.o）具有自定义格式。

符号U代表未定义，表示编译器不知道该符号在哪里。该符号必须重定位，即找到`Println`的地址，才能成功的进行调用。这就是链接器需要参与的工作。在介绍链接器的工作之前，我们分析了目标文件`main.o`, 以及它能够提供所有可用数据。链接器可以基于这些数据开展工作。

## 目标文件
这篇[文档](https://golang.org/pkg/cmd/internal/objabi/)解释了目标文件的内容和格式

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200701-Go:Object-File%26Relocations/1_WwlsAnj0J9-dUkvBYWS5sQ.png)

该文件由依赖项，调试信息(DWARF), 索引符号列表，数据段以及符号列表。符号列表中包含每个符号都需要我们进行重定向。以下是它的格式：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200701-Go:Object-File%26Relocations/1_so340hPaauZOPChu3tvSCA.png)

每个符号均以十六进制字节fe开头。可以使用十六进制编辑器打开目标文件main.o时。例如，对于Mac，可以使用xxd(译者注:xxd是mac下的一个命令)。 下面是内容的一部分，对符号(译者注:实际是对符号开头的标志"fe")进行了高亮显示。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200701-Go:Object-File%26Relocations/1_PL_o1t7dokehoO3X6rbUaw.png)

符号`main.main`是符号列表中的第一个符号。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200701-Go:Object-File%26Relocations/1_KOng-8Ed1XkprfvxSsqaXg.png)

前几个字节 `0102 00dc 0100 dc01 0a`  代表了前面定义的一系列属性：type、flag、size、data、以及重定位的次数。

字节以`zigzag-varint`格式存储。`varint`是以可变长字节的方式存储整数的值。 `zigzag`通过对最少有效位进行编码来是以减少编码后数据的大小

然后，重定位`Println`是一组字节序列`b201 0810 0008`:

* `b201`是偏移值89编码后的结果。这个偏移值是一个`int32` 类型。感谢`varint`，存储它仅耗费了2个字节.
* `08` 是需要重写的字节的数量，编码后的值是 4
* `10` 是重定位的类型，编码值8表示`R_CALL`, 即重定位函数调用
* `08` 是对索引符号的引用

装载器现在已经拥有了重定位所需的所有必要信息，可以生成可执行的二进制文件了。

## 重定位

链接器的其中一个阶段是分配虚地址给所有的段和指令。可以使用命令
```
objdump -h my-binary
```
可视化每个段的地址。下面是前面示例的输出

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200701-Go:Object-File%26Relocations/1_JGMu2mnGI-HTp35GHqx3mg.png)

函数`main`位于__text段，它也能通过命令
```
objdump -d my-binary
```
找到，这个命令显示了指令的地址。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200701-Go:Object-File%26Relocations/1_tZX5Ills5d4Dnk0Z5iZ1pA.png)

函数`main`入口地址是`109cfa0`，函数`fmt.Println`的入口地址是`1096a00`。一旦虚地址被分配，就会非常容易的重定位`fmt.Println`的入口地址。链接器将会用`fmt.Println`的入口地址依次减去`main`的入口地址、指令的偏移值、指令所占的字节大小。这样我们就能得到调用`fmt.Println`的全局偏移。对于前面的例子中，我们可以进行如下的操作：
```
1096a00 (fmt.Println) — 109cfa0 (main) — 84 (offset inside the main function) — 4 (size) = -26109
```

现在，指令知道函数`fmt.Println`的入口地址与当前内存地址的偏移是`-26109`，调用可以成功执行。

---
via: https://medium.com/a-journey-with-go/go-object-file-relocations-804438ec379b

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[vearne](https://github.com/vearne)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

