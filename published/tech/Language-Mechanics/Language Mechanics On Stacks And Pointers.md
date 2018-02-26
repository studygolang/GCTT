已发布：https://studygolang.com/articles/12443

# Go 语言机制之栈和指针

## 前言

本系列文章总共包括 4 篇，主要帮助大家理解 Go 语言中一些语言机制和其背后的设计原则，包括指针、栈、堆、逃逸分析和值传递/地址传递。这一篇是本系列的第一篇，主要介绍栈和指针。

以下是本系列文章的索引：

1. [Go 语言机制之栈与指针](https://studygolang.com/articles/12443)
2. [Go 语言机制之逃逸分析](https://studygolang.com/articles/12444)
3. [Go 语言机制之内存剖析](https://studygolang.com/articles/12445)
4. [Go 语言机制之数据和语法的设计哲学](https://www.ardanlabs.com/blog/2017/06/design-philosophy-on-data-and-semantics.html)

## 简介

我不打算说指针的好话，它确实很难理解。如果应用不当，会产生恼人的 bug，甚至会导致性能问题。当写并发和多线程程序时更是如此。所以许多语言试着用其它方法让编程人员避免指针的使用。但如果你是在用 Go 语言的话，你就不得不使用它们。如果不能很好的理解指针，是很难写出干净、简单并且高效的代码的。

## 帧边界（Frame Boundaries）

帧边界为每个函数提供了它自己独有的内存空间，函数就是在这个内存空间内执行的。帧边界除了可以让函数在自己的上下文环境中运行外还提供一些流程控制功能。函数可以通过帧边界指针直接访问自己帧边界中的内存，但如果想要访问自己帧边界外的内存，就需要用间接访问来实现了。要实现间接访问，被访问的内存必须和函数共享，要想弄清楚共享是怎么实现的，我们就得先了解一下由这些帧边界建立起来的内存结构以及其中的一些限制。

当一个函数被调用时，会在两个相关的帧边界间进行上下文切换。从调用函数切换到被调用函数，如果函数调用时需要传递参数，那么这些参数值也要传递到被调用函数的帧边界中。Go 语言中帧边界间的数据传递是按值传递的。

按值传递的好处是可读性好，拷贝并被函数接收到的值就是在函数调用时传入的值 。这就是为什么我把按值传递叫做 WYSIWYG（what you see is what you get 的缩写）。如果发生上下文环境转换时参数是按值传递的，我们就可以很清楚的知道这个函数调用会怎样影响程序的执行

让我们看一下下面这个小程序，主程序用按值传递的方式调用了一个函数：

### 清单 1

```go
package main

func main() {

   // Declare variable of type int with a value of 10.
   count := 10

   // Display the "value of" and "address of" count.
   println("count:\tValue Of[", count, "]\tAddr Of[", &count, "]")

   // Pass the "value of" the count.
   increment(count)

   println("count:\tValue Of[", count, "]\tAddr Of[", &count, "]")
}

//go:noinline
func increment(inc int) {

   // Increment the "value of" inc.
   inc++
   println("inc:\tValue Of[", inc, "]\tAddr Of[", &inc, "]")
}
```

程序启动后，语言运行环境会创建 main goroutine 来执行包含在函数 main 内的所有初始化代码。goroutine 是被放置在操作系统线程上的可执行序列，在 Go 语言的1.8版本中，为每一个 goroutine 分配了 2048 byte 的连续内存作为它的栈空间。这个初始化的内存大小几年来一直在变化，而且未来很有可能继续变化。

栈在 Go 语言中是非常重要的，因为它为分配给每个函数的帧边界提供了物理内存空间。main goroutine 在执行表 1 中的代码时，goroutine 的栈看起来像下面这个样子（在一个比较高的语言层次）

## 图 1

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/lang-mechanics/80_figure1.png)

在图 1 中可以看到，一部分栈空间被框了起来，作为函数 main 的可用空间，这块栈区域叫做「栈帧」,正是它界定了函数 main 在栈上的边界。这块栈空间是在函数被调用后，随着一些初始化代码的执行一并被创建的。可以看到变量 count 被放置到了函数 main 的栈帧中地址为 0x10429fa4 的地方。

在图 1 中也可以发现另外一点，就是在活动栈帧之下的栈空间是不可用的，只在活动栈帧以及它之上的栈空间是可用的。这个可用栈空间与不可用栈空间的边界我们需要明确一下。


## 地址

变量名是为了标识一块内存，使代码更具可读性而存在的。一个好的变量名可以让编程人员清楚的知道它代表了什么。如果你已经有了一个变量，那在内存中就有一个值与它对应；反之，如果在内存中有一个值，就必须有一个与之对应的变量，通过这个变量来访问这个内存值。在第 9 行，主函数调用了内置函数 println 来显示变量 count 的值和地址。

### 清单 2

```
09    println("count:\tValue Of[", count, "]\tAddr Of[", &count, "]")
```

用 & 操作符来获取变量的地址并不新鲜，许多其它语言也同样用这个操作符来获取变量地址。如果你在 32 位机器上运行这段代码（例如 playgournd ），第 9 行的输出应该像下面这样。

### 清单 3

    count:  Value Of[ 10 ]  Addr Of[ 0x10429fa4 ]

## 函数调用

接下来第 12 行，函数 main 调用了函数 increment：

### 清单 4

```
12    increment(count)
```

函数调用意味着 goroutine 需要在栈空间中创建一个新的栈帧。然而，这里并没有这么简单。要成功的调用一个函数，需要将数据在上下文转换过程中跨栈帧边界传递到新建的栈帧中。特别的，对于 integer 值，在调用过程中需要拷贝并传递过去，在第 18 行对函数 increment 的声明语句中可以看到这一点：

### 清单 5

```
18 func increment(inc int) {
```

如果再看一下第 12 行对函数 increment 的调用，可以看到传递的正是变量 count 的值。这个值经过拷贝、传递并最终放置到了函数 increment 的栈帧中。因为函数 increment 只能直接访问自己栈帧里的内存，所以它用变量 inc 来接收、存储和访问从变量 count 传递过来的值。

在函数 increment 刚刚要开始执行的时候，goroutine 的栈结构看起来像下面这个样子（从一个比较高的语言层次）。

## 图 2

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/lang-mechanics/80_figure2.png)

可以看到，现在在栈里有两个栈帧, 一个是函数 main 的，它下面的是函数 increment 的。在函数 increment 栈帧里，有一个变量 inc，它的值是当函数调用时从外面拷贝并传递过来的 10，它的地址是 0x10429f98，因为栈帧是从上往下使用栈空间的，所以它的地址比上面的小，不过这只是一个实现细节，并不保证所有实现都这样。重要的是 goroutine 把函数 main 的栈帧中的变量 count 的值拷贝并传递给了函数 increment 的栈帧中的变量 inc。

函数 increment 剩下的代码显示了变量 inc 的值和地址：

### 清单 6

```
21    inc++
22    println("inc:\tValue Of[", inc, "]\tAddr Of[", &inc, "]")
```

在 playground 平台上，第 22 行的输出看起来像这样：

### 表 7

    inc:    Value Of[ 11 ]  Addr Of[ 0x10429f98 ]

当执行完了这些代码以后，栈结构变成下面这个样子

### 图 3

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/lang-mechanics/80_figure3.png)

执行完第 21 行和第 22 行以后，函数 increment 返回，控制权重新回到了函数 main 中，然后函数 main 再一次显示了变量 count 的值和地址：

### 清单 8

```
14    println("count:\tValue Of[",count, "]\tAddr Of[", &count, "]")
```

在 playgournd 平台上，程序全部的输出如下：

### 清单 9

    count:  Value Of[ 10 ]  Addr Of[ 0x10429fa4 ]
    inc:    Value Of[ 11 ]  Addr Of[ 0x10429f98 ]
    count:  Value Of[ 10 ]  Addr Of[ 0x10429fa4 ]

## 函数返回

当函数返回，控制权回到调用函数后，栈结构发生了什么变化呢？答案是什么也没有。下面就是当函数 increment 返回后，栈结构的样子：

## 图 4

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/lang-mechanics/80_figure4.png)

除了为函数 increment 创建的栈帧现在变为不可用外，其他和图 3 一模一样。这是因为函数 main 的栈帧变成了活动栈帧。对函数 incrment 的栈帧没有做任何处理。

函数调用完成后，没有必要立即清理被调用函数的栈帧空间，这样做只会浪费时间，因为你不知道那块内存之后是否会被再次用到。所以相应内存就原封不动的留在那里。只有当发生了函数调用，这块内存被再次用到时，才会对它进行清理。清理过程是通过拷贝过来的值在这个栈帧中的初始化完成的，因为所有的变量至少会被初始化为相应类型的零值，这就保证了发生函数调用时，栈空间一定会被合理的清理。

## 值的共享

但是如果我们想在函数 increment 中直接操作存在于函数 main 的栈帧中的变量 count，应该怎么办呢？这时候我们就要用到指针了。指针存在在目的就是为了和一个函数共享变量，从而让这个函数可以对这个共享变量进行读写，即使这个变量没有直接放置在这个函数的栈帧中。

如果当你用指针时，一下子想到的不是「共享」，那就得看看是不是真的有必要使用指针了。当我们学习指针的内容时，有一点很重要，就是要用一个明确的单词而不是操作符或者语法来对待指针。所以请记住，用指针是为了共享，在阅读代码的时候也应该把 & 操作符当做共享来看。

## 指针类型

对每个已经声明的类型，不管是语言自己定义的还是用户定义的，都有一个与之对应的指针类型，用它来进行数据共享。比如 Go 语言中有一个内置的 int 类型，所以一定有一个与 int 类型对应的叫做 \*int 的指针类型。如果你定义了一个叫做 User 的类型，那么语言会自动为你生成一个与它对应的叫做 \*User 的指针类型。

所有的指针类型有两个共同点。一、它们以 \* 开头。二、它们占用相同的内存大小（4 个字节或者 8 个字节）并且表示的是一个地址。在 32 位的系统上（比如 playground )，一个指针占用 4 个字节，在 64 位的系统上（比如你自己的电脑）占用 8 个字节。

规范一点说，指针类型被认为是一个字面类型（type literals)，也就是说它是通过对已有类型进行组合而成的。

## 间接内存访问

看下面这段程序，它同样调用 了一个函数，不过这次传递的是变量的地址。这样被调用的函数 increment 就可以和函数 main 共享变量 count 了：

## 清单 10

```go
package main

func main() {

   // Declare variable of type int with a value of 10.
   count := 10

   // Display the "value of" and "address of" count.
   println("count:\tValue Of[", count, "]\t\tAddr Of[", &count, "]")

   // Pass the "address of" count.
   increment(&count)

   println("count:\tValue Of[", count, "]\t\tAddr Of[", &count, "]")
}

//go:noinline
func increment(inc *int) {

   // Increment the "value of" count that the "pointer points to". (dereferencing)
   *inc++
   println("inc:\tValue Of[", inc, "]\tAddr Of[", &inc, "]\tValue Points To[", *inc, "]")
}
```

同原来的程序比起来，新的程序存在 3 点不同

### 表 11

```
12    increment(&count)
```

在程序的第 12 行，并没有像之前一样传递变量 count 的值，而是传递的变量 count 的地址。现在我们可以说，我将要和函数 increment 共享变量 count 了，这就是 & 操作符想要表达的。

变量仍然是按值传递的，唯一不同的是，这次传递的是一个 integer 的地址。地址同样是一个值；这就是在函数调用时跨越两个帧边界被拷贝和传递的东西。

鉴于有一个值正在被拷贝和传递，在函数 inrement 中我们就需要一个变量来接收并存储这个基于地址的 integer 值，所以我们在程序的第 18 行把参数声明为了 \*int 类型。

### 表 12

```
18 func increment(inc *int) {
```

如果你传递的是 User 类型的地址值，这里声明的类型就应该换成 *User，尽管所有的指针存储的都是地址值，但是传递和接收的必须是同一个类型才可以，这个是关键。我们之所以要共享一个变量，是因为在函数内我们要对那个变量进行读写操作，而我们只有知道了这个类型的具体信息后才可以这样做。编译器会保证传递的是同一个指针类型的值。

下面是调用了函数 increment 后，栈结构的样子。

### 图 5

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/lang-mechanics/80_figure5.png)

在图 5 中我们可以看到，当把一个地址按值进行传递后，栈结构会变成什么样子。函数 increment 的栈帧中的指针变量 inc 指向了存在于函数 main 的栈帧中的变量 count。

通过这个指针变量，函数就可以以间接方式读写存在于函数 main 的栈帧中的变量 count 了。

### 清单 13

```
21    *inc++
```

这个时候，\* 被用作一个操作符和指针变量一起使用，把 * 用作操作符，意思是说要得到指针变量所指向的内容，在这里也就是函数 main 中的 count 变量。指针变量允许在使用它的栈帧中间接访问此栈帧之外的内存空间。有时候我们把这种间接访问叫做指针的解引用。在函数 increment 中仍然需要一个可以直接访问的本地指针变量来执行间接访问，在这里就是变量 inc。

当执行了第 21 行后，栈结构的变成下面这个样子。

### 图 6

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/lang-mechanics/80_figure6.png)

下面是程序的全部输出：

### 表 14

    count:  Value Of[ 10 ]   	   	Addr Of[ 0x10429fa4 ]
    inc:    Value Of[ 0x10429fa4 ]  	Addr Of[ 0x10429f98 ]   Value Points To[ 11 ]
    count:  Value Of[ 11 ]   	   	Addr Of[ 0x10429fa4 ]

可以看到，变量 inc 的值正是变量 count 的地址，就是这一个联系才使得访问本栈帧外的内存成为可能。一旦函数 increment 通过指针变量执行了写操作，当控制返回到函数 main 后，修改就会反应到对应的共享变量中。

## 指针型变量并不特别

指针类型和其它类型一样，一点也不特殊。它们有一块分配的内存并存放了一个值，抛开它指向的类型，指针类型总是占用同样的大小并且有相同的表示。唯一可能让我们感到困惑的是字符 \*，在函数 increment 内部，它被用作操作符，在函数声明时用来声明指针变量。如果你可以分清指针声明时和指针的解引用操作时的区别，应该就没那么困惑了。

## 总结

这篇文章讨论了设计指针背后的目的，以及在 Go 语言中栈和指针是怎样工作的。这是理解 Go 语言的语言机制、设计哲学的第一步，也对写出一致的、可读性好的代码有一定的指导作用。

下面来总结一下我们学到了什么：

1. 帧边界为每个函数提供了独立的内存空间，函数就是在自己的帧边界内执行的
2. 当调用函数时，上下文环境会在两个帧边界间切换
3. 按值传递的优点是可读性好
4. 栈是非常重要的，因为它为分配给每个函数的帧边界提供了可访问的物理内存空间
5. 在活动栈帧以下的栈空间是不可用的，只有活动栈帧和它之上的栈空间是可用的
6. 函数调用意味着 goroutine 需要在栈上为函数创建一个新的栈帧
7. 只有当发生了函数调用 ，栈区块被分配的栈帧占用后，相应栈空间才会被初始化
8. 使用指针是为了和被调用函数共享变量，使被调用函数可以用间接方式访问自己栈帧之外的变量
9. 每一个类型，不管是语言内置的还是用户定义的，都有一个与之对应的指针类型
10. 使用指针变量的函数，可以通过它间接访问函数栈帧之外的内存
11. 指针变量和其它变量一样，并不特殊，同样是有一块内存，在其中存放值而已

---
via: https://www.ardanlabs.com/blog/2017/05/language-mechanics-on-stacks-and-pointers.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[jettyhan](https://github.com/jettyhan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
