# Upspin 中的错误处理

[Upspin](https://upspin.io/) 项目使用自定的包 —— [upspin.io/errors](https://godoc.org/upspin.io/errors) —— 来表示系统内部出现的错误条件。这些错误满足标准的 Go [error](https://golang.org/pkg/builtin/#error) 接口，但是是使用自定义类型 [upspin.io/errors.Error](https://godoc.org/upspin.io/errors#Error)，该类型具有一些已经证明对项目有用的属性。
  
这里，我们会演示这个包是如何工作的，以及如何使用。The story
holds lessons for the larger discussion of error handling in Go.  
  
**Motivations**  
  
A few months into the project, it became clear we needed a consistent approach
to error construction, presentation, and handling throughout the code. We
decided to implement a custom errors package, and rolled one out in an
afternoon. The details have changed a bit but the basic ideas behind the
package have endured. These were:  
  

  * To make it easy to build informative error messages.
  * To make errors easy to understand for users.
  * To make errors helpful as diagnostics for programmers.

  
As we developed experience with the package, some other motivations emerged.
We'll talk about these below.  
  
**A tour of the package**  
  
The [upspin.io/errors](https://godoc.org/upspin.io/errors) package is imported
with the package name "errors", and 所以，在 Upspin 中，它取代了 Go 标准的“errors”包。
  
我们注意到，Upspin 中的错误信息的元素都是不同类型的：用户名、路径名、错误种类（I/O、Permission 等等），诸如此类。This provided the starting point for the package,
which would build on these different types to construct, represent, and report
the errors that arise.  
  
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
  
为了协助错误构建，这个包提供了一个名为 [E](https://godoc.org/upspin.io/errors#E) 的函数，它简短并且易打。

```go 
  func E(args ...interface{}) error  
```
  
As the [doc comment](https://godoc.org/upspin.io/errors#E) for the function says, E builds an error value from its arguments. The type of each argument
determines its meaning. The idea is to look at the types of the arguments and
assign each argument to the field of the corresponding type in the constructed
Error struct. There is an obvious correspondence: a PathName goes to
Error.Path, a UserName to Error.User, and so on.  
  
Let's look at an example. In typical use, calls to errors.E will arise
multiple times within a method, so we define a constant, conventionally called
op, that will be passed to all E calls within the method:  

```go  
  func (s *Server) Delete(ref upspin.Reference) error {  
    const op errors.Op = "server.Delete"  
     ...  
```
Then through the method we use the constant to prefix each call (although the
actual ordering of arguments is irrelevant, by convention op goes first):  
```go   
  if err := authorize(user); err != nil {  
    return errors.E(op, user, errors.Permission, err)  
  } 
``` 
  
The String method for E will format this neatly:  
  
  server.Delete: user ann@example.com: permission denied: user not authorized  
  
If the errors nest to multiple levels, redundant fields are suppressed and the
nesting is formatted with indentation:  
  
  client.Lookup: ann@example.com/file: item does not exist:  
          dir/remote("upspin.example.net:443").Lookup:  
          dir/server.Lookup  
  
Notice that there are multiple operations mentioned in this error message
(client.Lookup, dir/remote, dir/server). We'll discuss this multiplicity in a
later section.  
  
As another example, sometimes the error is special and is most clearly
described at the call site by a plain string. To make this work in the obvious
way, the constructor promotes arguments of literal type string to a Go error
type through a mechanism similar to the standard Go function
[errors.New](https://golang.org/pkg/errors/#New)[](https://www.blogger.com/).
Thus one can write:  
```go   
   errors.E(op, "unexpected failure")  
```  
or  
```go   
   errors.E(op, fmt.Sprintf("could not succeed after %d tries", nTries))  
``` 
and have the string be assigned to the Err field of the resulting Err type.
This is a natural and easy way to build special-case errors.  
  
**Errors across the wire**  
  
Upspin is a distributed system and so it is critical that communications
between Upspin servers preserve the structure of errors. To accomplish this we
made Upspin's RPCs aware of these error types, using the errors package's
[MarshalError](https://godoc.org/upspin.io/errors#MarshalError) and
[UnmarshalError](https://godoc.org/upspin.io/errors#UnmarshalError) functions
to transcode errors across a network connection. These functions make sure
that a client will see all the details that the server provided when it
constructed the error.  
  
Consider this error report:  
  
  client.Lookup: ann@example.com/test/file: item does not exist:  
         dir/remote("dir.example.com:443").Lookup:  
         dir/server.Lookup:  
         store/remote("store.example.com:443").Get:  
         fetching https://storage.googleapis.com/bucket/C1AF...: 404 Not Found  
  
This is represented by four nested errors.E values.  
  
Reading from the bottom up, the innermost is from the package
[upspin.io/store/remote](http://upspin.io/store/remotehttps://godoc.org/upspin.io/store/remote)
(responsible for taking to remote storage servers). The error indicates that
there was a problem fetching an object from storage. That error is constructed
with something like this, wrapping an underlying error from the cloud storage
provider:  
```go   
  const op errors.Op = `store/remote("store.example.com:443").Get`  
  var resp *http.Response  
  ...  
  return errors.E(op, errors.Sprintf("fetching %s: %s", url, resp.Status))  
```  
The next error is from the directory server (package
[upspin.io/dir/server](https://godoc.org/upspin.io/dir/server), our directory
server reference implementation), which indicates that the directory server
was trying to perform a Lookup when the error occurred. That error is
constructed like this:  
```go   
  const op errors.Op = "dir/server.Lookup"  
  ...  
  return errors.E(op, pathName, errors.NotExist, err)  
```  
This is the first layer at which a Kind (errors.NotExist) is added.  
  
The Lookup error value is passed across the network (marshaled and unmarshaled
along the way), and then the
[upspin.io/dir/remote](https://godoc.org/upspin.io/dir/remote) package
(responsible for talking to remote directory servers) wraps it with its own
call to errors.E:  
```go   
  const op errors.Op = "dir/remote.Lookup"  
  ...  
  return errors.E(op, pathName, err)  
```  
There is no Kind set in this call, so the inner Kind (errors.NotExist) is
lifted up during the construction of this Error struct.  
  
Finally, the [upspin.io/client](https://godoc.org/upspin.io/client) package
wraps the error once more:  
```go 
  const op errors.Op = "client.Lookup"  
  ...  
  return errors.E(op, pathName, err)  
```  
Preserving the structure of the server's error permits the client to know
programmatically that this is a "not exist" error and that the item in
question is "ann@example.com/file". The error's
[Error](https://godoc.org/upspin.io/errors#Error.Error) method can take
advantage of this structure to suppress redundant fields. If the server error
were merely an opaque string we would see the path name multiple times in the
output.  
  
The critical details (the PathName and Kind) are pulled to the top of the
error so they are more prominent in the display. The hope is that when seen by
a user the first line of the error is usually all that's needed; the details
below that are more useful when further diagnosis is required.  
  
Stepping back and looking at the error display as a unit, we can trace the
path the error took from its creation back through various network-connected
components to the client. The full picture might help the user but is sure to
help the system implementer if the problem is unexpected or unusual.  
  
**Users and implementers**  
  
There is a tension between making errors helpful and concise for the end user
versus making them expansive and analytic for the implementer. Too often the
implementer wins and the errors are overly verbose, to the point of including
stack traces or other overwhelming detail.  
  
Upspin's errors are an attempt to serve both the users and the implementers.
The reported errors are reasonably concise, concentrating on information the
user should find helpful. But they also contain internal details such as
method names an implementer might find diagnostic but not in a way that
overwhelms the user. In practice we find that the tradeoff has worked well.  
  
In contrast, a stack trace-like error is worse in both respects. The user does
not have the context to understand the stack trace, and an implementer shown a
stack trace is denied the information that could be presented if the server-
side error was passed to the client. This is why Upspin error nesting behaves
as an _operational_ trace, showing the path through the elements of the
system, rather than as an _execution_ trace, showing the path through the
code. The distinction is vital.  
  
For those cases where stack traces would be helpful, we allow the errors
package to be built with the "debug" tag, which enables them. This works fine,
but it's worth noting that we have almost never used this feature. Instead the
default behavior of the package serves well enough that the overhead and
ugliness of stack traces are obviated.  
  
**Matching errors**  
  
An unexpected benefit of Upspin's custom error handling was the ease with
which we could write error-dependent tests, as well as write error-sensitive
code outside of tests. Two functions in the errors package enable these uses.  
  
The first is a function, called
[errors.Is](https://godoc.org/upspin.io/errors#Is), that returns a boolean
reporting whether the argument is of type *errors.Error and, if so, that its
Kind field has the specified value.  
```go   
  func Is(kind Kind, err error) bool  
```  
This function makes it straightforward for code to change behavior depending
on the error condition, such as in the face of a permission error as opposed
to a network error:  
```go  
  if errors.Is(errors.Permission, err) { ... }  
```  
The other function, [Match](https://godoc.org/upspin.io/errors#Match), is
useful in tests. It was created after we had been using the errors package for
a while and found too many of our tests were sensitive to irrelevant details
of the errors. For instance, a test might only need to check that there was a
permission error opening a particular file, but was sensitive to the exact
formatting of the error message.  
  
After fixing a number of brittle tests like this, we responded by writing a
function to report whether the received error, err, matches an error template:  
```go   
  func Match(template, err error) bool  
```  
The function checks whether the error is of type *errors.Error, and if so,
whether the fields within equal those within the template. The key is that it
checks _only_ those fields that are non-zero in the template, ignoring the
rest.  
  
For our example described above, one can write:  
```go   
  if errors.Match(errors.E(errors.Permission, pathName), err) { … }  
```  
and be unaffected by whatever other properties the error has. We use Match
countless times throughout our tests; it has been a boon.  
  
**Lessons**  
  
There is a lot of discussion in the Go community about how to handle errors
and it's important to realize that there is no single answer. No one package
or approach can do what's needed for every program. As was pointed out
[elsewhere](https://blog.golang.org/errors-are-values), errors are just values
and can be programmed in different ways to suit different situations.  
  
The Upspin errors package has worked out well for us. We do not advocate that
it is the right answer for another system, or even that the approach is right
for anyone else. But the package worked well within Upspin and taught us some
general lessons worth recording.  
  
The Upspin errors package is modest in size and scope. The original
implementation was built in a few hours and the basic design has endured, with
a few refinements, since then. A custom error package for another project
should be just as easy to create. The specific needs of any given environment
should be easy to apply. Don't be afraid to try; just think a bit first and be
willing to experiment. What's out there now can surely be improved upon when
the details of your own project are considered.  
  
We made sure the error constructor was both easy to use and easy to read. If
it were not, programmers would resist using it.  
  
The behavior of the errors package is built in part upon the types intrinsic
to the underlying system. This is a small but important point: No general
errors package could do what ours does. It truly is a custom package.  
  
Moreover, the use of types to discriminate arguments allowed error
construction to be idiomatic and fluid. This was made possible by a
combination of the existing types in the system (PathName, UserName) and new
ones created for the purpose (Op, Kind). Helper types made error construction
clean, safe, and easy. It took a little more work—we had to create the types
and use them everywhere, such as through the "const op" idiom—but the payoff
was worthwhile.  
  
Finally, we would like to stress the lack of stack traces as part of the error
model in Upspin. Instead, the errors package reports the sequence of events,
often across the network, that resulted in a problem being delivered to the
client. Carefully constructed errors that thread through the operations in the
system can be more concise, more descriptive, and more helpful than a simple
stack trace.  
  
Errors are for users, not just for programmers.  
  
_by Rob Pike and Andrew Gerrand_  


----------------

via: https://commandcenter.blogspot.co.uk/2017/12/error-handling-in-upspin.html

作者：[Rob Pike](https://plus.google.com/101960720994009339267)
译者：[ictar](https://github.com/ictar)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出