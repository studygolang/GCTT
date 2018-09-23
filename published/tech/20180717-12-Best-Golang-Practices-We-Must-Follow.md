首发于：https://studygolang.com/articles/14889

# 我们必须遵循的 12 个 Go 语言最佳实践

本文旨在提供一个切实的指导，在 Go 语言中实现最佳实践和设计模式。这些编程技巧可以帮助开发者编写出较好的代码。你一定已经读过了 [Go Tutorial](https://tour.golang.org/welcome/1) 和 [Effective Go](https://golang.org/doc/effective_go.html)。

为了让大家对这些编程技巧有更加深刻的认识，我在讨论这些最佳实践的时候会附加一些示例代码。

那些编写了许多优秀代码的大师们，一直在使用一些 Go 语言编程实践或者说是编程技巧。

下面列出了其中一些最好的编程实践，可以使写出的代码简单、易懂并且易于维护

1、使用 gofmt

2、通过首先处理错误来避免代码嵌套

3、错误字符串

4、错误处理

5、尽量避免代码重复

6、变量名声明

7、用类型选择语句来处理特例

8、在类型选择语句中声明变量

9、重要的代码要放在源文件的前面

10、点导入

11、注释代码

12、注释语句规范

## 1、使用  gofmt

对源代码执行 gofmt 命令，会自动修正大部分粗心导致的问题。几乎全世界的 Go 语言开发者都在用 gofmt。

gofmt 首先读取源代码，然后输出经过缩进、垂直对齐甚至规范注释后的代码。

### 命令和选项

gofmt 文件名 - 输出格式化后的代码

gofmt -w 文件名 - 重新格式化代码并更新文件

gofmt -r'rule' 文件名 - 格式化代码前执行指定的规则

gofmt 包所在的路径 - 格式化整个包下的源文件

### 下面是使用 gofmt 的一些小例子

文件名：demo.go

```go
package main
         import "fmt"
// this is demo to format code
           // with gofmt command
var a int=10;
            var b int=15;
                           var c string= "Welcome to Agira";
      func print(){
                  fmt.Println("Value for a,b and c is : ");
                       fmt.Println(a);
                                fmt.Println((b));
                                        fmt.Println(c);
                        }
```

输入命令：$ gofmt demo.go

输出结果：

```go
package main

import "fmt"

// this is demo to format code
// with gofmt command
var a int = 10
var b int = 15
var c string =  “Welcome to Agira”

func print() {
       fmt.Println("Value for a,b and c is : ")
       fmt.Println(a)
       fmt.Println((b))
       fmt.Println(c)
}
```

## 2、通过首先处理错误来避免代码嵌套

避免使用多重条件或者嵌套条件，当我们处理后面的代码前需要处理错误，例如下面的代码

```go
err := request()
if err != nil {
   // handling error
} else {
   // normal code
}
```

我们可以用下面的方式代替

```go
err := request()

if err != nil {
 // handling error
 return // or continue, etc.
}
// proceed to further
```

嵌套条件语句越少，读者越容易理解

如果 if 语句中包含初始化语句，例如：

```go
if x, err := f(); err != nil {
   // handling error
   return
} else {
   // use x
}
```

我们应该在代码中定义一个短变量，在之后的 if 语句中使用这个变量

```go
x, err := f()
if err != nil {
   // handling error
   return
}
// use x
```

## 3、错误字符串

错误字符串首字母不应该大写（除非是以一些特殊的名词或者缩写开头）。

例如：

fmt.Errorf("Something went wrong") 应该写成 fmt.Errorf("something went wrong")

## 4、错误处理

不要用 _ 来忽略错误。如果一个函数可能返回错误信息，检查函数的返回值 ，确认函数是否执行成功了。更好的做法是处理这个错误并返回，不然的话如果出现任何异常程序会产生一个 panic 错误

### 不要用 panic 错误

不要在正常处理流程中使用 panic, 那种情况下可以用 error 和多重返回值。

## 5、尽可能避免重复

如果你想在控制模块和数据模块使用同一个类型结构，创建一个公共文件，在那里声明这个类型

## 6、变量名声明

在 Go 编程中最好用短的变量名，尤其是那些作用域比较有限的局部变量

用 `c` 而不是 `lineCount`

用 `i` 而不是 `sliceIndex`

1、基本规则：距离声明的地方越远，变量名需要越具可读性。

2、作为一个函数接收者，1、2 个字母的变量比较高效。

3、像循环指示变量和输入流变量，用一个单字母就可以。

4、越不常用的变量和公共变量，需要用更具说明性的名字。

## 7、用类型选择语句来处理特例

如果你不确定 iterface{} 是什么类型，就可以用类型选择语句

例如：

```go
func Write(v interface{}) {
 switch v.(type) {
 case string:
   s := v.(string)
   fmt.Printf(“%T\n”,s)
 case int:
   i := v.(int)
   fmt.Printf(“%T\n”,i)
 }
}
```

## 8、在类型选择语句中声明变量

在类型选择语句中声明的变量，在每个分支中会自动转化成正确的类型

例如：

```go
func Write(v interface{}) {
 switch x := v.(type) {
 case string:
   fmt.Printf(“%T\n”,x)
 case int:
   fmt.Printf(“%T\n”,x)
 }
}
```

## 9、重要的代码要放在源文件的前面

如果你有像版权声明、构建标签、包注释这样的重要信息，尽量写在源文件的靠前位置。
我们可以用空行把导入语句分成若干个组，标准库放在最前面。

```go
import (
  "fmt"
  "io"
  "log"
  "golang.org/x/net/websocket"
)
```

在接下来的代码中，首先写重要的类型，在最后写一些辅助型的函数和类型。

## 10、点导入

点导入可以测试循环依赖。并且它不会成为被测试代码的一部分：

```go
package foo_test

import (
  "bar/testutil" // also imports "foo"
  . "foo"
)
```

这样的情况下，测试代码不能放在 foo 包中，因为它引入了 bar/testutil包，而它导入了 foo。所以我们用点导入 的形式让文件假装是包的一部分，而实际上它并不是。除了这个使用情形外，最好不要用点导入。因为它会让读者阅读代码时更加困难，因为很难确定像 Quux 这样的名字是当前包的顶层声明还是引入的包。

## 11、注释代码

在包名字之前添加包相关的注释

```go
// Package playground registers an HTTP handler at “/compile” that
// proxies requests to the golang.org playground service.

package playground
```

出现在 godoc 中的标识符，需要适当的注释

```go
// Author represents the person who wrote and/or is presenting the document.
type Author struct {
  Elem []Elem
}

// TextElem returns the first text elements of the author details.
// This is used to display the author’ name, job title, and company
// without the contact details.
func (p *Author) TextElem() (elems []Elem) {
```

## 12、注释语句规范

即使注释语句看上去有一些冗余，也需要是一个完整的句子，。这样会让它们在 godoc 中有更的格式化效果。注释需要以被注释的名字开头，以点号结尾。

```go
// Request represents a request to run a command.
type Request struct { …

// Encode writes the JSON encoding of req to w.
func Encode(w io.Writer, req *Request) { … and so on.
```

希望这些 Go 语言最佳实践可以帮助你提高代码质量。我们也列出了其它许多技术的最佳实践，可以在 [largest blog repository](http://www.agiratech.com/blog/) 找到。有其它问题可以通过 [info@agiratech.com](info@agiratech.com) 联系我们

---

via: http://www.agiratech.com/12-best-golang-agile-practices-we-must-follow/

作者：[Reddy Sai](http://www.agiratech.com/author/reddysai/)
译者：[jettyhan](https://github.com/jettyhan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
