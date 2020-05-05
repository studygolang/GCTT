首发于：https://studygolang.com/articles/28432

# Go：跨团队协作时共享对象

![Illustration created for “A Journey With Go”, made from *the original Go Gopher, created by Renee French.*](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20190622-Go-Shared-Objects-for-Cross-Team-Collaboration/00.png)

在一个公司中，跨小组/团队协作时如果各方使用不同的语言，有时会很复杂。在[我的团队](https://www.propertyfinder.ae/)中，我们维护用 Go 语言写的项目，而我们的合作方，数据科学小组使用 Python。他们提供给我们一些公式，我们需要在项目中翻译成 Go。

当我们实现了这个公式后，我们需要对方组的验证，以确保我们实现得没有错误。让他们来测试并让他们自己提供测试数据无疑是最好的方法。幸运的是，我们可以用 Go 来实现这个过程。

## cgo 和共享对象

构建命令 `go build` 有一个参数可以[把你的 Go 包编进 C shared library](https://golang.org/pkg/cmd/go/internal/help/#pkg-variables)：

```bash
-buildmode=c-shared
   Build the listed main package, plus all packages it imports,
   into a C shared library. The only callable symbols will
   be those functions exported using a cgo //export comment.
   Requires exactly one main package to be listed.
```

用这个模式你可以编译一个 shared library，文件以 “.so” 结尾，你可以在其他语言中如 C，Java，Ruby 或 Python 中直接使用。然而，这个模式只有 [Cgo](https://golang.org/cmd/cgo/) 支持，你可以在你的 Go 包中写、调用 C 代码。基于此，你可以写一个自己的库，让别的小组用他们自己的语言来调你的库。

## 实现

Go 和共享对象间的网关的实现看起来很简单。首先你需要在你想导出的每一个函数前添加注释 `//export MyFunction`。然后你需要在强制性 `import "C"` 之前前置声明你的 C 结构体。下面是我们代码的简化版：

```go
import (
   /*
   typedef struct{
   int from_bedroom;
   int to_bedroom;
   int from_price;
   int to_price;
   int from_size;
   int to_size;
   int types[5];
   } lead;

   typedef struct{
   int bedroom;
   int price;
   int size;
   int type_id;
   } property;
   */
   "C"
)
//export NewProperty
func NewProperty(b int, p int, s int, t int) C.property {
   // business logic

   return C.property{
      bedroom:   C.int(b),
      price:     C.int(p),
      size:      C.int(s),
      type_id:   C.int(t),
   }
}
//export NewLead
func NewLead(fb int, tb int, fp int, tp int, fs int, ts int, t []int) C.lead {
   // business logic
   return C.lead{
      from_bedroom: C.int(fb),
      to_bedroom:   C.int(tb),
      from_price:   C.int(fp),
      to_price:     C.int(tp),
      from_size:    C.int(fs),
      to_size:      C.int(ts),
      types:        types,
   }
}
//export CalculateDistance
func CalculateDistance(l C.lead, p C.property) {
   // business logic here
}
```

 因为你不能导出 Go 的结构体，所以你需要把 C 的结构体作为输入/输出参数进行处理。当你写完代码后，可以使用命令 `go build -o main.so -buildmode=c-shared main.go` 来编译。为了能编译成功，你的 Go 代码中需要有 main 包和 main 函数。然后，你就可以写你的 Python 脚本了：

```python
#!/usr/bin/env python
from ctypes import *

# loading shared object
matching = cdll.LoadLibrary("main.so")

# Go type
class GoSlice(Structure):
    _fields_ = [("data", POINTER(c_void_p)), ("len", c_longlong), ("cap", c_longlong)]

class Lead(Structure):
    _fields_ = [('from_bedroom', c_int),
                ('to_bedroom', c_int),
                ('from_price', c_int),
                ('to_price', c_int),
                ('from_size', c_int),
                ('to_size', c_int),
                ('types', GoSlice)]

class Property(Structure):
    _fields_ = [('bedroom', c_int),
                ('price', c_int),
                ('size', c_int),
                ('type_id', c_int)]

#parameters definition
matching.NewLead.argtypes = [c_int, c_int, c_int, c_int, c_int, c_int, GoSlice]
matching.NewLead.restype = Lead

matching.NewProperty.argtypes = [c_int, c_int, c_int, c_int]
matching.NewProperty.restype = Property

matching.CalculateDistance.argtypes = [Lead, Property]

lead = lib.NewLead(
    # from bedroom, to bedroom
    1, 2,
    # from price, to price
    80000, 100000,
    # from size, to size
    750, 1000,
    # type
    GoSlice((c_void_p * 5)(1, 2, 3, 4, 5), 5, 5)
)
property = lib.NewProperty(2, 90000, 900, 1)

matching.CalculateDistance(lead, property)
```

你的共享对象中的所有导出的方法都应该在你的 Python 文件中有描述：类型、参数顺序及返回参数。

之后，你可以运行你的 Python 脚本 `python3 main.py`。

## 使用的方便程度

乍一看很简单，但可能需要花费很长时间才能正常运行。Python 中没有很多关于这个的文档或例子，而且很难调试。如果用 `.argtypes` 或 `.restype` 对你暴露的方法描述不当，可能会导致意想不到的结果或出现 `segmentation fault` 错误信息，且不会有足够多的信息来帮助调试。

这个 Go/Python 的通信方式很适合跨团队测试，但我不建议在大型的项目或以后在生产环境中用。因为这种开发方式很复杂，容易耗费较长时间。

---

via: https://medium.com/a-journey-with-go/go-shared-objects-for-cross-team-collaboration-b3af7d9e73af

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
