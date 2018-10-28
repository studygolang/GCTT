首发于：https://studygolang.com/articles/13892

# 在 Go 中使用 C 语言的动态库

我和我的儿子在上周末干了一件非常有意思的事情，我们开发了一个用 Go 编写的命令行游戏，最近我正在重写一款曾经在年轻时开发的游戏，当时用的还是 Kaypro II。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Using-C-Dynamic-Libraries-In-Go-Programs/kayproii.jpg)

我钟爱这台电脑，回想起曾经使用 BASIC 在上面日日夜夜开发游戏，它非常便携，把键盘折叠起来就可以提着走，哈哈。

额，我好像偏题了，还是回到Go上面来。我发现一种使用 VT100 控制符来显示简单屏幕的方法，并且在上面开始写一些业务逻辑。

但随后就遇到了一些艰难的问题，我要用倒叙的方式来描述一下，比如当不按回车键时，我就没办法从标准输入中获取数据，啊啊啊啊啊，为了寻找解决方案，我整个周末都在阅读资料，甚至找到两个相关的 Go 语言库，但是并没有起到什么作用。后来我意识到，如果要实现这个效果，那么要使用 C 语言来编写功能函数，链接成动态库后再由 Go 调用。

在一家爱尔兰小酒吧中我开发了四个小时终于解决了这个问题，在这里我要好好感谢一下吉尼斯黑啤酒给我带来的启发和激励。要知道我过去十年一直是在 windows 下使用 C#，十年之前的话我还是在大微软环境中用 C/C++ 开发,所以我对 Linux 和 Mac 系统下的 gcc、gco，静态和动态库都不熟悉，我到目前都是一直在学习这些东西，毕竟要学的东西很多。

经过一番探索之后，问题开始变得明朗起来，我需要使用 ncurse 动态库（注：ncurse 库提供了 API，可以允许程序员编写独立于终端的基于文本的用户界面），于是我决定先写一个简单实例程序，如果能在 C 编译器下能用，那么我觉得应该在 Go 语言下也是可以的。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Using-C-Dynamic-Libraries-In-Go-Programs/Screen%2BShot%2B2013-08-20%2Bat%2B1.57.49%2BPM.png)

在 Mac 系统下 ncurse 库的路径是 /usr/lib，这有个关于库的文档：
[https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man3/ncurses.3x.html](https://www.ardanlabs.com/blog/broken-link.html)

下面是 C 语言的测试程序头文件：

test.h

```c
int GetCharacter();
void InitKeyboard();
void CloseKeyboard();
```

然后是源文件:

test.c

```c
#include <curses.h>
#include <stdio.h>
#include "test.h"

int main() {
    InitKeyboard();

    printf("\nEnter: ");
    refresh();

    for (;;) {
        int r = GetCharacter();
        printf("%c", r);
        refresh();

        if (r == ‘q’) {
            break;
        }
    }

    CloseKeyboard();

    return 0;
}

void InitKeyboard() {
    initscr();
    noecho();
    cbreak();
    keypad(stdscr, TRUE);
    refresh();
}

int GetCharacter() {
    return getch();
}

void CloseKeyboard() {
    endwin();
}
```

接下来就是困难的部分了，该如何使用 gcc 编译器来编译这个测试程序呢? 我想确保是用同一个编译器来编译 Go 和 C 语言，而且只用到最少的编译参数和标志。

经过一个小时的探索，我写了下面这个 makefile，说实话，这是我第一次写 makefile 文件。

makefile

```makefile
build:
    rm -f test
    gcc -c test.c
    gcc -lncurses -r/usr/lib -o test test.o
    rm -f *.o
```

当运行 make 命令时，会在当前路径下搜索 makefile 文件并执行，需要注意的是每个命令右边的缩进必须使用一个 Tab 键，如果你使用空格键，那么可能会遇到问题，当然你手动运行那些命令也能工作，不过为了方便我还是用 makefile。

让我们来分析一下当 makefile 被执行时 gcc 编译器是怎么处理的：

下面的调用会使 gcc 根据源码创建一个 test.o 的文件。`-c` 参数是告诉 gcc 只需要编译源码文件并且创建 test.o 的文件。（注：如果不加 -c，那么会自动执行后面的链接流程）

```makefile
gcc -c test.c
```

接下来，gcc 会把 test.o 和 libncurses.dylib 进行链接处理，链接后会生成 test 可执行文件。命令中的 l（小写的 L）参数是让 gcc 去链接 libncurses.dylib 文件，-r（小写 R）参数指定了 gcc 去哪个路径下获取这个库文件，-o（小写的 O）参数是指定 gcc 导出可执行文件的名字，最后让gcc在链接操作中包含 test.o。

```makefile
gcc -lncurses -r/usr/lib -o test test.o
```

以上两条命令就能编译出一个能正常工作的 test 程序，你可以在命令行下输入 `./test` 来执行它：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Using-C-Dynamic-Libraries-In-Go-Programs/Screen%2BShot%2B2013-08-20%2Bat%2B3.34.39%2BPM.png)

程序运行后有个循环监听字符输入，能把我输入的字符进行显示，当我按下 `q` 键时，程序会关闭。

现在我已经有一个使用 ncurses 动态库的并且能运行的程序，但是我想在 Go 中使用它，现在我需要找到一种方法能把之前写的程序包装成动态库，然后被 Go 使用。

非常幸运的是我找到了一些非常棒的文章，里面包含了动态库给 Go 使用的方法：

http://www.adp-gmbh.ch/cpp/gcc/create_lib.html

http://stackoverflow.com/questions/3532589/how-to-build-a-dylib-from-several-o-in-mac-os-x-using-gcc

让我们在Go中实现这一切吧，先来建立一个新的工程：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Using-C-Dynamic-Libraries-In-Go-Programs/Screen%2BShot%2B2013-08-20%2Bat%2B6.48.52%2BPM.png)

我建立了一个名叫 Keyboard 的文件夹，里面有两个子文件夹，分别叫 DyLib 和 TestApp。

在 DyLib 文件夹中我们放入C的动态库源码和 makefile 文件，在 TestApp 中只有一个 main.go 文件，到时就使用这个文件来测试 Go 和 C 语言的动态库交互。

这是为动态库准备的 C 头文件，和之前 test 头文件中的内容一样：

keyboard.h

```c
int GetCharacter();
void InitKeyboard();
void CloseKeyboard();
```

然后是实现了我们所需功能的 C 源码文件，除了没有 main 函数，其他内容也和之前的 test 程序中的代码相同，因为我们要创建一个库文件，所以不需要 main 函数。

keyboard.c

```c
#include <curses.h>
#include "keyboard.h"

void InitKeyboard() {
    initscr();
    noecho();
    cbreak();
    keypad(stdscr, TRUE);
    refresh();
}

int GetCharacter() {
    return getch();
}

void CloseKeyboard() {
    endwin();
}
```

接下来是为创建动态库准备的makefile文件：

makefile

```makefile
dynamic:
    rm -f libkeyboard.dylib
    rm -f ../TestApp/libkeyboard.dylib
    gcc -c -fPIC keyboard.c
    gcc -dynamiclib -lncurses -r/usr/lib -o libkeyboard.dylib keyboard.o
    rm -f keyboard.o
    cp libkeyboard.dylib ../TestApp/libkeyboard.dylib

shared:
    rm -f libkeyboard.so
    rm -f ../TestApp/libkeyboard.so
    gcc -c -fPIC keyboard.c
    gcc -shared -W1 -lncurses -r/usr/lib -soname,libkeyboard.so -o libkeyboard.so keyboard.o
    rm -f keyboard.o
    cp libkeyboard.so ../TestApp/libkeyboard.so
```

使用这个 makefile 可以创建一个动态库或者共享库。使用 make 是如果不加任何参数，那么会执行 dynamic 标记下的那些命令行，如果加上 `shared` 参数，就会创建一个共享库文件。

要注意一个重要的 **-fPIC** 标记，有这个标记的时候 gcc 会生成共享库所需要的地址无关代码，当没有这个标记时会生成可执行程序。

让我们用 makefile 文件来创建动态库：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Using-C-Dynamic-Libraries-In-Go-Programs/Screen%2BShot%2B2013-08-20%2Bat%2B6.53.51%2BPM.png)

执行 make 命令，它会运行 makefile 文件中 dynamic 部分的命令，等运行完毕后我们就有了新的动态库:

![](https://raw.githubusercontent.com/nicedevcn/gctt-images/master/Using-C-Dynamic-Libraries-In-Go-Programs/Screen%2BShot%2B2013-08-20%2Bat%2B6.55.09%2BPM.png)

到了这一步我们可以在 DyLib 和 TestApp 文件夹下看到生成的 libkeyboard.dylib 文件。

有个事情忘记说了，所生成的动态库或共享库名字必须以 lib 开头，如果不这么做的话就无法正常使用，同时库文件必须放置在程序运行时能够加载的工作目录下。

接下来我们来看一下 Go 下的测试程序代码：

```go
package main

/*
#cgo CFLAGS: -I../DyLib
#cgo LDFLAGS: -L. -lkeyboard
#include <keyboard.h>
*/
import "C"
import (
    "fmt"
)

func main() {
    C.InitKeyboard()

    fmt.Printf("\nEnter: ")

    for {
        r := C.GetCharacter()

        fmt.Printf("%c", r)

        if r == ‘q’ {
            break
        }
    }

    C.CloseKeyboard()
}
```

Go 开发小组提供了两篇文章来解释 Go 是如何和 C 语言的动态库进行交互的，对于理解上面的代码拥有非常重要的作用：

- http://golang.org/cmd/cgo/
- http://golang.org/doc/articles/c_go_cgo.html

如果你如何关联 C++ 的库感兴趣的话，那么 SWIG（简单封装和接口生成器）值得你去了解：

- http://www.swig.org/
- http://www.swig.org/Doc2.0/Go.html

SWIG 还是留到下次再讨论，先来分解一下上面的 Go 代码:

```go
package main

/*
#cgo CFLAGS: -I../DyLib
#cgo LDFLAGS: -L. -lkeyboard
#include <keyboard.h>
*/
import "C"
```

为了提供给编译器和链接器所需要的参数，我们需要使用特殊的 cgo 命令，这是一组内部提供的命令，必须在 `import "C"` 语句上方声明，并且和 import 命令之间不能有空行或者其他语句，否则会造成编译错误。

我们在编译和链接程序的过程中需要向 Go 提供上面的参数来生成进程标识 ，CFLAGS 标记向编译器提供参数，我们让编译器能在共享库所在文件夹中找到头文件，LDFLAGS 向链接器提供一些参数，可以看到我们使用了两个，`-L` 向链接器提供了动态库的路径，`-l` 则提供了动态库的名字。

有一点需要注意，当我们指定库名的时候不需要包含前缀（lib）和后缀名（.dylib)，程序会自动在名字前面加上 lib，在后面加上 .dylib 或者 .so 后缀名.

最后我们让 Go 导入一个特殊的包 `"C"`，它提供了 Go 语言层面访问我们库的方式，没有这个包，那我们的这一切都没法完成。

通过以下方式，我们可以调用库中的函数：

```go
C.InitKeyboard()
r := C.GetCharacter()
C.CloseKeyboard()
```

有了这个 `"C"` 包，能把头文件中的每个函数进行封装，这些封装后的函数能将输入和输出进行相应解析处理，请留意我们是如何使用原生 Go 类型和语法从键盘输入中获取字符的。

现在我们可以在命令行中构建和运行一个测试程序了：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/Using-C-Dynamic-Libraries-In-Go-Programs/Screen%2BShot%2B2013-08-20%2Bat%2B7.30.11%2BPM.png

太棒了，程序正常工作了！

有了这些能让游戏真正好玩起来的键盘事件，现在我和儿子可以继续开发我们的游戏了。

我花费了数小时处理这一切，如果想要做得更好，那么还需要学习更多的知识点，过段时间我会研究一下 SWIG 和 C++ 面向对象库的结合，不过现在能引入和使用 C 语言库已经非常好了。

如果你想浏览和获取这些代码，我已经把项目放到了 Github 仓库的 Keyboard 下，好好享用！！

阅读第二章：[Using CGO with Pkg-Config And Custom Dynamic Library Locations](https://www.ardanlabs.com/blog/2013/08/using-cgo-with-pkg-config-and-custom.html)

---

via: https://www.ardanlabs.com/blog/2013/08/using-c-dynamic-libraries-in-go-programs.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[Maxwell Hu](https://github.com/nicedevcn)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
