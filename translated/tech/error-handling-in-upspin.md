# Upspin 中的错误处理

[Upspin](https://upspin.io/) 项目使用自定义的包 —— [upspin.io/errors](https://godoc.org/upspin.io/errors) —— 来表示系统内部出现的错误条件。这些错误满足标准的 Go [error](https://golang.org/pkg/builtin/#error) 接口，但是使用的是自定义类型 [upspin.io/errors.Error](https://godoc.org/upspin.io/errors#Error)，该类型具有一些已经证明对项目有用的属性。
  
这里，我们会演示这个包是如何工作的，以及如何使用这个包。这个故事为关于 Go 中的错误处理更广泛的讨论提供了经验教训。
  
**动机**  
  
在项目进行几个月后，我们清楚地知道，我们需要一致的方法来处理整个代码中的错误构建、描述和处理。我们决定实现一个自定义的 errors 包，并在某个下午将其推出。相较于初始实现，细节已经有所变化，但是，这个包背后的基本理念经久不衰。它们是：
  

  * 为了便于构建有用的错误信息。
  * 为了使用户易于理解错误。
  * 为了让错误帮助程序员进行问题诊断。

  
随着我们开发此包的经验增长，出现了一些其他的需求。下面我们会聊到这些。
  
**errors 包之旅**  
  
[upspin.io/errors](https://godoc.org/upspin.io/errors) 包是用包名“errors”导入的，所以，在 Upspin 中，它取代了 Go 标准的“errors”包。
  
我们注意到，Upspin 中的错误信息的元素都是不同类型的：用户名、路径名、错误种类（I/O、Permission 等等），诸如此类。这为 errors 包提供了起始点，它将建立在这些不同类型之上，以构建、表示和报告出现的错误。
  
这个包的中心是 [Error](https://godoc.org/upspin.io/errors#Error) 类型，这是一个 Upspin 错误的具体表示。它具有多个字段，任何一个字段都可以不做设置：

```go 
  type Error struct {  
      Path upspin.PathName  
      User upspin.UserName  
      Op  Op  
      Kind Kind  
      Err error  
  }  
```

Path 和 User 字段表示操作影响的路径和用户。注意，这些都是字符串，但是分别为 Upspin 中独有的类型，以表明其用途，并且使得类型系统可以捕获到某些类型的编程错误。
  
Op 字段表示执行的操作。它是另一种字符串类型，通常保存方法名或者报告错误的服务器函数名称：“client.Lookup”、“dir/server.Glob”等等。
  
Kind 字段把错误分类为一组标准条件（Permission、IO、NotExist，[诸如此类](https://godoc.org/upspin.io/errors#Kind)）中的一员。这使得我们很容易就可以看到出现的错误的类型的简洁描述，并且还提供了连接到其他系统的钩子。例如，[upspinfs](https://godoc.org/upspin.io/cmd/upspinfs) 把 Kind 字段当成将 Upspin 错误转换成 Unix 错误常量（例如 EPERM 和 EIO）的键来使用。
  
最后一个字段，Err，保存另一个错误值。通常是来自其他系统的错误，例如 [os](https://golang.org/pkg/os/) 包的文件系统错误，或者 [net](https://golang.org/pkg/net/) 包的网络错误。它也有可能是另一个 upspin.io/errors.Error 值，用以创建错误跟踪（稍后我们会讨论）。
  
**构建错误**  
  
为了协助错误构建，这个包提供了一个名为 [E](https://godoc.org/upspin.io/errors#E) 的函数，它简短并且便于输入。

```go 
  func E(args ...interface{}) error  
```
  
如该函数的[文档注释](https://godoc.org/upspin.io/errors#E)所述，E 根据其输入参数构建 error 值。每一个参数的类型决定了其自身的含义。思想是检查每一个参数的类型，然后将参数值赋给已构造的 Error 结构中对应类型的字段。这里有一个明显的对应点：PathName 对应 Error.Path，UserName 对应 Error.User，以此类推。
  
让我们看一个例子。通常情况下，一个方法中会出现多次对 errors.E 的调用，因此，我们定义一个常量，按惯例称其为 op，它会作为参数传给方法中所有 E 调用：

```go  
  func (s *Server) Delete(ref upspin.Reference) error {  
    const op errors.Op = "server.Delete"  
     ...  
```

然后，在整个方法中，我们都会把这个常量作为每一次 E 调用的第一个参数（虽然参数的实际顺序是不相干的，但是按惯例，op 放在第一个）：

```go   
  if err := authorize(user); err != nil {  
    return errors.E(op, user, errors.Permission, err)  
  } 
``` 
  
E 的 String 方法将会将其整洁地格式化：

```  
  server.Delete: user ann@example.com: permission denied: user not authorized  
```
  
如果错误多级嵌套，那么会抑制冗余字段，并且使用缩进来格式化嵌套：

```  
  client.Lookup: ann@example.com/file: item does not exist:  
          dir/remote("upspin.example.net:443").Lookup:  
          dir/server.Lookup  
```

注意，这条错误信息中提到了多种操作（client.Lookup，dir/remote，dir/server）。在后面的部分，我们会讨论这种多重性。
  
又如，有时，错误是特殊的，并且在调用处通过一个普通的字符串来清楚描述。为了以明显的方式使其行之有效，构造器通过类似于标准的 Go 函数 [errors.New](https://golang.org/pkg/errors/#New)[](https://www.blogger.com/) 的机制，将文字类型字符串参数转换成 Go error 类型。

因此，可以这样写：
```go   
   errors.E(op, "unexpected failure")  
```  
或者
```go   
   errors.E(op, fmt.Sprintf("could not succeed after %d tries", nTries))  
``` 
这样，会让字符串赋给结果 Err 类型的 Err 字段。这是构建特殊错误的一种自然而然的简单方式。
  
**跨网络错误**  
  
Upspin 是一个分布式系统，因此，Upspin 服务器之间的通信保留错误的结构则是至关重要的。为了做到这一点，我们使用 errors 包的 [MarshalError](https://godoc.org/upspin.io/errors#MarshalError) 和 [UnmarshalError](https://godoc.org/upspin.io/errors#UnmarshalError) 函数来在网络连接中转码错误，从而让 Upspin 的 RPC 知道这些错误类型。这些函数确保客户端将看到服务器在构造错误时提供的所有细节。
  
考虑下面的错误报告：
```  
  client.Lookup: ann@example.com/test/file: item does not exist:  
         dir/remote("dir.example.com:443").Lookup:  
         dir/server.Lookup:  
         store/remote("store.example.com:443").Get:  
         fetching https://storage.googleapis.com/bucket/C1AF...: 404 Not Found  
```  
它由四个嵌套的 errors.E 值构成。
  
从下往上看，最里面的部分来自于包 [upspin.io/store/remote](http://upspin.io/store/remotehttps://godoc.org/upspin.io/store/remote) （负责与远程存储服务器交互）。这个错误表示，在从存储获取对象时出现问题。该错误大概是这样构建的，封装了来自云储存提供者的一个底层错误：

```go   
  const op errors.Op = `store/remote("store.example.com:443").Get`  
  var resp *http.Response  
  ...  
  return errors.E(op, errors.Sprintf("fetching %s: %s", url, resp.Status))  
```  

下一个错误来自目录服务器（包 [upspin.io/dir/server](https://godoc.org/upspin.io/dir/server)，我们的目录服务器参考实现），它表示目录服务器在错误发生时正在尝试进行查找操作。这个错误是像这样构建的：

```go   
  const op errors.Op = "dir/server.Lookup"  
  ...  
  return errors.E(op, pathName, errors.NotExist, err)  
```  

这是第一层，其中，增加了一个 Kind（errors.NotExist）。
  
Lookup 错误值通过网络传递（一路上被打包和解包），接着，[upspin.io/dir/remote](https://godoc.org/upspin.io/dir/remote) 包（负责跟远程目录服务器交互）通过它自己对 errors.E 的调用来封装这个错误：

```go   
  const op errors.Op = "dir/remote.Lookup"  
  ...  
  return errors.E(op, pathName, err)  
```  

在这个调用中，没有设置任何 Kind，因此，在构建这个 Error 结构时，使用内部的 Kind（errors.NotExist）。
  
最终，[upspin.io/client](https://godoc.org/upspin.io/client) 包再一次封装这个错误：
```go 
  const op errors.Op = "client.Lookup"  
  ...  
  return errors.E(op, pathName, err)  
```  
保留服务器错误结构使得客户端能够以编程的方式知道这是一个“not exist”错误，以及问题的相关项是“ann@example.com/file”。错误的 [Error](https://godoc.org/upspin.io/errors#Error.Error) 方法可以利用这个结构来抑制冗余字段。如果服务器错误只是一个含糊不清的字符串，那么我们会在输出中多次看到路径名。
  
关键细节（PathName 和 Kind）被拉到错误的顶部，这样的话，在展示中它们会更突出。期望是，当用户看到这些错误时，错误的第一行通常就够了；当需要进一步的诊断的时候，下面的细节会更有用。

我们回过头来把错误展示作为一个整体，我们可以通过各种网络连接组件，从错误的产生一直追踪到客户端。完整的错误链也许会帮助到用户，但它是一定能帮到系统的实现者的，这能帮助他们确定问题是不是意料之外的，或者是不是非同寻常的。
  
**用户和实现者**  
  
让错误对终端用户有用并且保持用户简洁，与让错误对实现者而言信息丰富并且可供分析，二者之间存在矛盾。常常是实现者胜出，而错误变得过于冗余，达到了包含堆栈跟踪或者其他淹没式细节的程度。
  
Upspin 的错误试图让用户和实现者都满意。报告的错误适度简洁，关注于用户应该觉得有用的信息。但它们还包含内部信息，例如实现者可能用以分析的方法名，但又不会把用户淹没。在实践中，我们发现这种权衡工作良好。

相反，类似于堆栈跟踪的错误在这两方面上都更糟糕。用户没有上下文可以理解堆栈跟踪，而如果服务端错误被传给客户端的话，那么看到堆栈跟踪的实现者会被拒绝看到应该可能出现的信息。这就是为什么 Upspin 错误嵌套相当于_操作_跟踪（显示系统元素路径），而不是_执行_跟踪（显示代码执行路径）。这个区别至关重要。
  
对于那些堆栈跟踪可能会有用的场景，我们允许使用“debug”标签来构建 errors 包，这将会允许打印堆栈跟踪。这个工作良好，但是值得注意的是，我们几乎从不使用这个功能。相反，errors 包的默认行为已经够好了，避免了堆栈跟踪的开销和不堪入目。
  
**匹配错误**  
  
Upspin 的自定义错误处理的一个意想不到的好处是，易于编写错误依赖的测试以及编写测试之外的错误敏感代码。errors 包中的两个函数使得这些用法成为可能。
  
首先是一个函数，名为 [errors.Is](https://godoc.org/upspin.io/errors#Is)，它返回一个布尔值，表明参数 err 是否是 *errors.Error 类型，如果是，那么它的 Kind 字段是否为指定的值。
```go   
  func Is(kind Kind, err error) bool  
```  
这个函数使得代码可以根据错误条件直接改变行为，例如与网络错误不同，在面对权限错误时：
```go  
  if errors.Is(errors.Permission, err) { ... }  
```  
另一个函数， [Match](https://godoc.org/upspin.io/errors#Match)，对测试有用。在我们已经使用 errors 包一段时间，然后发现我们太多的测试是对错误细节敏感时，创建了它。例如，一个测试也许只需要检查打开一个特定的文件时有权限错误，但对错误信息的准确格式敏感。
  
在修复了许多像这样的脆弱的测试之后，我们编写了一个函数来报告接收到的错误 err 是否匹配一个错误模板 （template）：
```go   
  func Match(template, err error) bool  
```  
这个函数检查错误是否是 *errors.Error 类型的，如果是，那么错误中的字段是否与模板中的那些字段相等。关键是，它_只_检查模板中的那些非零字段，忽略其他字段。 
  
对于上述例子，我们可以这样写：
```go   
  if errors.Match(errors.E(errors.Permission, pathName), err) { … }  
```  
并且不会受到该错误的其他属性影响。在我们的测试中，我们无数次使用 Match；它就是一个大惊喜。
  
**经验教训**  
  
在 Go 社区中，有大量关于如何处理错误的讨论，重要的是，要意识到这个问题并没有单一的答案。没有一个包或者是一个方法可以满足所有程序的需求。正如[这里](https://blog.golang.org/errors-are-values)指出的，错误只是值，并且可以以不同的方式编程，从而满足不同的场景。
  
Upspin 的 errors 包对我们有好处。我们并非主张对于另一个系统，它就是正确的答案，或者甚至说这个方法适合其他人。但是这个包在 Upspin 中用得不错，并且教会我们一些值得记录的经验教训。
  
Upspin 的 errors 包的大小和涵盖的范围适度。其初始实现是在几个小时内完成的，而基本的设计保留了下来，并且自完成后，经历了一些改进。为另一个项目定制一个错误包应该也很容易。应该很容易适用于任何特定环境的具体需求。不要害怕尝试；只需先想一下，并且愿意尝试。当考虑你自己的项目的细节时，思考现在有什么可以改进的地方。
  
我们确保错误构造器易于使用，并且易于阅读。如果并非如此，那么编程人员可以拒绝使用。

errors 包的行为一定程度建立在底层系统内部的类型上的。这是一个很小但是很重要的点：没有哪个一般的错误包可以做到我们做到的东西。它真的是一个自定义包。
  
此外，区别参数的类型的使用使得错误构建变得通顺流畅。这个可以通过组合系统中现有的类型（PathName、UserName）和为该目的而创建的新类型（Op、Kind）来实现。帮助式类型使得错误构建干净、安全并且容易。它花费一点额外的工作量（我们必须创建这些类型，然后处处使用它们，例如通过“const op”），但结果是值得的。
  
最后，我们想要强调，缺乏堆栈跟踪是 Upspin 中的错误模型的一部分。相反，errors 包报告事件序列（通常跨网络），这样子产生的是传递给客户端的错误。通过系统中的操作小心构造错误可以比简单的堆栈跟踪更简洁、更具描述性以及更有用。

错误是给用户的，而不只是给程序员的。（Errors are for users, not just for programmers.）
  
_来自 Rob Pike 和 Andrew Gerrand_  


----------------

via: https://commandcenter.blogspot.co.uk/2017/12/error-handling-in-upspin.html

作者：[Rob Pike](https://plus.google.com/101960720994009339267)
译者：[ictar](https://github.com/ictar)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出