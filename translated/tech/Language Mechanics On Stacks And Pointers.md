Go语言结构之栈和指针

**前言**

本系列文章总共包括4篇，主要帮助大家理解Go语言中一些语法结构和其背后的设计原则，包括指针、栈、堆、指针逃逸分析和值传递/地址传递。这一篇是本系列的第一篇，主要介绍栈和指针

以下是本系列文章的索引
1) Go语言结构之栈与指针
2) Go语言结构之指针逃逸分析
3) Go语言结构之内存剖析
4) Go语言结构之数据和语法的设计哲学

**简介**


我不打算说指针的好话，它确实很难理解。如果应用不当，会产生恼人的bug，甚至是导致性能问题。当写并发和多线程的软件时更是如此。所以许多语言试着用其它方法让编程人员避免指针的使用。但如果你是在用Go语言的话，你就不得不使用它们。如果不能很好的理解指针，很难写出干净、简单并且高效的代码。

**Frame Boundaries**

Frame Boundaries(以下简称Frame)为每个函数提供了它自己独有的内存空间，函数就是在这个内存空间内执行的。每个Frame除了可以让函数在自己的上下文环境中运行还提供一些流程控制功能。函数可以通过Frame指针直接访问自己Frame中的内存，但如果想要访问自己Frame之外的内存，就需要用间接访问来实现了。要实现间接访问，被访问的内存必需和函数共享，要想弄清楚是怎么实现共享的，首先我们需要了解一下由这些Frame建立起来的内存结构和一些限制

当一个函数被调用时，两个相关的Frame之间会发生上下文切换。从调用函数切换到被调用函数，如果函数调用时需要参数，那么这些参数值也要传递到被调用函数的Frame中。Go语言中Frame间的数据传递是按值传递的。

按值传递的好处是可读性好，拷贝并被函数接收到的值就是在函数调用时传入的值 。这就是为什么我把按值传递叫做WYSIWYG(what you see is what you get 的缩写)。这样上下文环境转换发生时，我们可以很清楚的知道调用一个函数会怎样影响程序的执行

让我们看一下下面这个小程序，主程序用按值传递的方式调用了一个函数

**Listing 1**

	01 package main
	02
	03 func main() {
	04
	05    // Declare variable of type int with a value of 10.
	06    count := 10
	07
	08    // Display the "value of" and "address of" count.
	09    println("count:\tValue Of[", count, "]\tAddr Of[", &count, "]")
	10
	11    // Pass the "value of" the count.
	12    increment(count)
	13
	14    println("count:\tValue Of[", count, "]\tAddr Of[", &count, "]")
	15 }
	16
	17 //go:noinline
	18 func increment(inc int) {
	19
	20    // Increment the "value of" inc.
	21    inc++
	22    println("inc:\tValue Of[", inc, "]\tAddr Of[", &inc, "]")
	23 }

程序启动后，语言运行环境会创建主函数的goroutine来执行包括在main函数内的所有初始化代码。goroutine是放置在操作系统线程上的可执行序列，在Go语言的1.8版本中，为每一个goroutine分配了2048 byte的连续内存作为它的栈空间。这个初始化的内存大小几年来一直在变化，而且未来很有可能继续变化。

栈在Go语言中是非常重要的，因为它为分配给每个函数的Frame提供了物理内存空间。当主函数的goroutine执行 Listing 1中的函数代码时，goroutine的栈看起来像下面这个样子（在一个比较高的语言层次）

**Figure 1**

![](https://www.ardanlabs.com/blog/images/goinggo/80_figure1.png)

在Figure 1中可以看到，一部分栈空间被框了起来，作为main函数的可用空间，这块栈区域叫做“stack frame",正是它界定了main函数在栈上的边界。这块栈空间是在函数被调用后，随着一些初始化代码执行一并被创建的。可以看到变量count的被放置到了main函数farme中的地址0x10429fa4中

在Figure 1中可以看到，在栈上为main函数框出了一个区块。这个区块叫做 "stack frame"，就是这个区块定义了main函数的在栈上的可用范围。这个区块是在函数被调用的时候建立的。能看到count变量已经被放置到了main函数地址空间中的0x10429fa4地址上

在Figure 1中也可以发现另外一点，就是在活动Frame之下的栈空间是不可用的，只在活动Frame以及它之上的栈空间是可用的。这个 可用栈空间与不可用栈空间的边界我们需要明确一下。


**地址**

变量名是为了标识一块内存，使代码更具可读性而存在的。一个好的变量名可以让编程人员清楚的知道它代表的什么。如果你已经有了一个变量，那在内存中就有一个数值与它对应，反之，如果在内存中有一个数值，你必需能一个与之对应的变量才能访问这个内存值。在第9行，主函数调用了内置函数printn来显示变量count的值和地址

**Listing 2**

    09    println("count:\tValue Of[", count, "]\tAddr Of[", &count, "]")
    
用&操作符来获取变量的地址并不新鲜，许多其它语言也同样用这个操作符取变量地址。如果你在32位机器上运行这段代码（例如playgournd)，第9行的输出应该和下面的很像

**Listing 3**

    count:  Value Of[ 10 ]  Addr Of[ 0x10429fa4 ]

**函数调用**

接下来第12行，main函数调用了increment函数

Listing 4

    12    increment(count)
    

函数调用意味着goroutine需要在栈空间中框出一个新的区块。然而，这里并没有这么简单。要成功的调用 一个函数，数据需要在上下文转换中跨越Frame传递到新建的栈区域中。特别的，对于integer值，在调用过程中需要拷贝并传递过去，我们可以在18行的increment函数声明中看到这一点

Listing 5

    18 func increment(inc int) {
    
如果在再看一下第12行对函数increment的调用，可以看到传递的正是变量count的值。这个值经过拷贝、传递并最终放置到了为increment创建的Frame中。因为函数increment只能直接访问自己Frame内部的内存，所以它用变量inc来接收并存储和访问从变量count传递过来的值 

在函数increment刚刚要开始执行的时候，goroutine的栈结构看起来像下面这个样子（从一个比较高的语言层次）

Figure 2

![](https://www.ardanlabs.com/blog/images/goinggo/80_figure2.png)


可以看到，现在在栈里有两个Frame, 一个函数main的和它下的函数increment的。在函数increment的Frame内部，有一个变量inc，它的值是当函数调用时从外面拷贝并传递过来的10，它的地址是0x10429f98，因为Frame是从上往下占据栈空间的，所以它的地址比上面的小，不过这只是一个实现细节，并不保证所有实现都这样。重要的是goroutine把函数main的Frame中变量count的值拷贝并传递给了函数increment的Frame中变量inc.

函数increment余下的代码显示了变量inc的值和地址

Listing 6

    21    inc++
    22    println("inc:\tValue Of[", inc, "]\tAddr Of[", &inc, "]")

在playground平台上，第22行的输出看起来像这样

Listing 7

    inc:    Value Of[ 11 ]  Addr Of[ 0x10429f98 ]

当执行完了这些代码以后，栈结构变成下面这个样子

Figure 3

![](https://www.ardanlabs.com/blog/images/goinggo/80_figure3.png)


执行完第21行和22行后，函数increment返回，控制权重新回到了函数main中，然后main函数再一次显示了变量count的值和地址

Listing 8

    14    println("count:\tValue Of[",count, "]\tAddr Of[", &count, "]")


在playgournd平台上，程序全部的输出如下

Listing 9

    count:  Value Of[ 10 ]  Addr Of[ 0x10429fa4 ]
    inc:    Value Of[ 11 ]  Addr Of[ 0x10429f98 ]
    count:  Value Of[ 10 ]  Addr Of[ 0x10429fa4 ]


**函数返回**

当函数返回，控制权回到调用函数后，栈结构发生了什么变化呢？ 答案是什么也没有。下面就是当函数increment返回后，栈结构的样子

Figure 4

![](https://www.ardanlabs.com/blog/images/goinggo/80_figure4.png)

除了为函数increment创建的Frame现在变为不可用外，其他和Figure 3 一模一样。变是因为函数main的Frame现在变成了活动Frame。对为函数incrment创建的内存块没有做任何处理。

清理已经调用完成函数的Frame只是浪费时间，因为你不知道那块内存之后是否会被再次用到。所心相应内存就原封不动的留在那里。只有当发生函数调用，这块内存被再次用到时，才会对它进行清理。清理过程是通过拷贝过来的值在这个Frame中的初始化完成的，因为所有的变量至少会被初始化为相应类型的零值，这就保证了发生函数调用时，栈空间一定会被合理的处理

**值的共享**

但是如果我们想在函数increment中直接操作存在于函数main的Frame中的变量count，应该怎么办呢？这时候我们就要用到指针了。指针存在在目的就是为了和一个函数共享一个变量，从而让这个函数可以对这个共享变量进行读写，即使这个变量没有直接放置在这个函数的Frame中

如果当你用指针时，一下子想到的不是”共享“，那就得看看是不是有使用指针的必要了。当我们学习指针的内容时，有一点很重要，就是要用一个明确的单词而不是操作符或者语法来对待指针。所以请记住，用指针是为了共享，在阅读代码的时候也应该把&操作符当共享来看


**指针类型**

对每个已经声明的类型，不管是语言自己定义的还是用户定义的，都有一个与之对应的指针类型，用它来进行数据共享。比如Go语言中有一个内置的int类型，所以一定有一个与int对应的叫做\*int的指针类型。如果你定义了一个叫做User的类型，那么语言会自动为你生成一个与它对应的叫做\*Userr指针类型

所有的指针类型有两个共同点。一、它们以*开头。二、它们占用相同的内存大小（4个字节或者8个字节）并且表示的是一个地址。在一个32位的系统上（比如playground)，一个指针占用4个字节，在一个64位的系统上（比如你自己的电脑）占用8个字节

规范一点说，指针类型被认为是一个字面类型（type literals)，也就是说它是通过对已经有的类型组合而成的

**间接内存访问**

看下面这段程序，它同样调用 了一个函数，不过这次传递的是变量的地址。这样被调用的函数incrment就可以和函数main共享变量count了

Listing 10

    01 package main
    02
    03 func main() {
    04
    05    // Declare variable of type int with a value of 10.
    06    count := 10
    07
    08    // Display the "value of" and "address of" count.
    09    println("count:\tValue Of[", count, "]\t\tAddr Of[", &count, "]")
    10
    11    // Pass the "address of" count.
    12    increment(&count)
    13
    14    println("count:\tValue Of[", count, "]\t\tAddr Of[", &count, "]")
    15 }
    16
    17 //go:noinline
    18 func increment(inc *int) {
    19
    20    // Increment the "value of" count that the "pointer points to". (dereferencing)
    21    *inc++
    22    println("inc:\tValue Of[", inc, "]\tAddr Of[", &inc, "]\tValue Points To[", *inc, "]")
    23 }


同原来的程序比起来，新的程序存在3点不同

Listing 11

    12    increment(&count)

在程序的第12行，并没有像之前一样传递变量count的值，而传递的是变量count的地址。现在我们可以说，我将要和函数increment共享变量count，这就是&操作符想要表达的。

变量的传递方式仍然是按值传递，唯一不同的是，这次传递的是一个integer的地址。地址同样也是一个值；这就是在函数调用时跨越两个Frame被拷贝和传递的东西

鉴于有一个值正在被拷贝和传递，在函数inrement中我们就需要一个变量来接收并存储这个基于地址的integer值，所以我们在程序的第18把参数声明为了integer指针类型

Listing 12

    18 func increment(inc *int) {

如果你传递的是User类型的地址值，这里声明的类型就应该换成\*User，尽管所有的指针存储的都是地址值，传递和接收的必需是同一个类型才可以，这个是关键。我们之所心要共享一个变量，是因为在函数内我们要对那个变量进行读写操作，我们只有知道了这个类型的具体信息后才可以这样做。编译器会保证传递的是同一个指针类型的值。

下面是调用了函数increment后，栈结构的样子

Figure 5

![](https://www.ardanlabs.com/blog/images/goinggo/80_figure5.png)

在Figure 5中我们可以看到，当把一个值地址按值进行传递后，栈结构会变成什么样子。函数increment的Frame中的指针变量inc现在指向了存在于函数main的Frame中的变量count

通过这个指针变量，函数就可以以间接方式读写存在于函数main的Frame中的变量count了

Listing 13

    21    *inc++
    
这个时候，\*被用作一个操作符和指针变量一起使用，把\*用作操作符，意思是说要取获取指针变量所指向的内容，在这里也就是main函数中的count变量。指针变量允许在使用它的Frame中间接访问此Frame之外的内存空间。有时候我们把这种间接访问叫做指针的解引用。在函数incremen中仍然需要一个可以直接访问的本地指针变量来执行间接访问，在这里就是变量inc.

现在是执行了第21行后，栈结构的样子

Figure 6

![](https://www.ardanlabs.com/blog/images/goinggo/80_figure6.png)

下面是程序的全部输出

Listing 14

    count:  Value Of[ 10 ]   	   	Addr Of[ 0x10429fa4 ]
    inc:    Value Of[ 0x10429fa4 ]  	Addr Of[ 0x10429f98 ]   Value Points To[ 11 ]
    count:  Value Of[ 11 ]   	   	Addr Of[ 0x10429fa4 ]

可以看到，变量inc的值正是变量count的地址值，就是这一个关联才使得访问本Frame外的内存成为可能。一旦函数increment通过指针变量执行了写操作，当控制返回到函数main后，修改就会反应到对应的共享变量中。

**指针型变量并不特别**

指针类型和其它类型一样，一点也不特殊。它们有一块分配的内存并存放了一个值，抛开它指针的类型，指针类型总是占用同样的大小和相同的表示。唯一可能让我们感觉困惑的是字符\*，在函数increment内部，它被用作操作符，在函数声明时用来声明指针变量。如果你可以分清指针声明时和指针的解引用操作时的区别，应该就没那么困惑了


**总结**

这篇文章讨论了设计指针背后的目的，以及在Go语言中栈和指针是怎样工作的。这是理解go语言的语言结构，设计哲学的第一步，也对写出一致的、可读性好的代码有一定指导作用。

来总结一下我们学到了什么

1、函数运行在给函数分配的Frame boundaries中。它提供了函数可访问的物理内存空间

2、当调用函数时，上下文环境会在两个Frame中切换

3、按值传递的优点是可读性好

4、栈是非常重要的，因为它为分配给每个函数的Frame boundaries提供了可访问的物理内存空间

5、在活动Frame以下的栈空间是不可用的，只能活动Frame和之上的栈空间是可用的

6、函数调用意味着goroutine需要在栈上为函数创建一个新的区块

7、只有当发生函数调用 ，栈区块被分配的Frame占用后，相应栈空间才会被初始化

8、指针是为了和被调用函数共享变量，使被调用函数可以间接访问自己Frame之外的变量

9、每一个类型，不管是语言内置的还是用户定义的，都有一个与之对应的指针类型

10、使用指针变量的函数，可以通过它间接访问函数Frame外的内存

11、指针变量和其它变量一样，并不特殊，同样是有一块内存，在其中存放数值而已

via: https://hackernoon.com/the-beauty-of-go-98057e3f0a7d

作者：[Kanishk Dudeja](https://hackernoon.com/@kanishkdudeja?source=post_header_lockup)
译者：[jettyhan](https://github.com/jettyhan)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

