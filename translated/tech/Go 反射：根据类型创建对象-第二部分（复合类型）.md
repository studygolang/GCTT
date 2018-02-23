# Go 反射：根据类型创建对象-第二部分（复合类型）

> 这是关于 Golang 中根据类型创建对象系列博客的第二篇，讨论的是创建复合对象。第一篇在[这里]()

在前一篇博客中，我解释了 go reflect 包 `type` 和 `kind` 的概念。这篇博客，我将深入探讨这些术语。因为相比原始类型，`type` 和 `kind` 对于复合类型来说含义更多。

## 类型和种类
“类型” 是程序员用来描述程序中数据和函数的元数据。`type` 在 Go 的运行时和编译器中有不同的含义。

可以通过一个例子来理解。例如有如下声明：

```go
var Version string
```

提到这个声明，程序员会说 `Version` 是一个 `string` 类型的变量。

考虑下面这个例子，`Version` 是一个复合类型。

```go
type VersionType string
var Version VersionType
```

在这个例子中，第一行创建了一个新类型 `VersionType`。第二行，`Version` 被定义成一个 `VersionType` 类型的变量。这个类型（VersionType）的类型是 `string`，我们称之为 `kind`。

总结来说就是，`Version` 的 `type` 是 `VersionType`，`Version` 的 `kind` 是 `string`。

> 类型是程序员定义的关于数据和函数的元数据。种类是编译器和运行时定义的关于数据和函数的元数据。


运行时和编译器根据 `Kind` 来分别给变量和函数分配内存或栈空间。


## 根据原始种类来创建复合对象
创建 `kind` 为如下值的复合对象与根据原始种类创建原始对象并没有什么不同。

```go
       Bool
       Int
       Int8
       Int16
       Int32
       Int64
       Uint
       Uint8
       Uint16
       Uint32
       Uint64
       Uintptr
       Float32
       Float64
       Complex64
       Complex128
       String
       UnsafePointer
```

下面是通过类型（`VersionType`）来创建 `Version` 的例子。因为 `Version` 有原始的 `kind`，故它可以使用零值来创建。

```go
func CreatePrimitiveObjects(t reflect.Type) reflect.Value {
  return reflect.Zero(t)
}

func extractVersionType(v reflect.Value) (VersionType, error) {
  if v.Type().String() != "VersionType" {  
    return "", errors.New("invalid input")
  }
  return v.String(), nil
}

// 译者注：上面代码似乎有错误，改成了下面所示

func extractVersionType(v reflect.Value) (string, error) {
    if v.Type().String() != "main.VersionType" {
	  return "", errors.New("invalid input")
	}
    return v.Type().String(), nil
}
```

注意到一个 `Type` 类型的变量的 `String()` 方法将返回 `Type` 的全路径名称。例如， 如果 `VersionType` 定义在 `mypkg` 包中，`String()` 返回的值将为 `mypkg.VersionType`。

## 通过复合种类创建复合对象
复合种类是包括有其他种类的种类。`Map`，`Struct`，`Array` 等，都是复合种类。下面是复合种类的列表：

```go
  Array
  Chan
  Func
  Interface
  Map
  Ptr
  Slice
  Struct
```

可以像原始对象一样使用零值来创建复合对象。但是，仅使用一个空值而不做其他额外的操作的话，它们并不能被初始化。下面一节将详细讨论如何初始化复合种类。

## 通过 type signature 来创建复合数组对象
一个空的数组对象可以通过零值来创建。一个数组的零值是一个空的数组对象。下面是通过数组的 type signature 来创建一个数组的例子：

```go
func CreateCompositeObjects(t reflect.Type) reflect.Value {
  return reflect.Zero(t)
}
```

该函数会创建一个包含一个任意空复合对象的 `reflect.Value` 类型的结构体。

reflect 包有一个 `ArrayOf(int, Type)` 函数，可以从来创建包含类型为 `Type` 的特定长度的数组，下面是一个示例：

```go
func CreateArray(t reflect.Type, length int) reflect.Value {
  var arrayType reflect.Type
  arrayType = reflect.ArrayOf(length, t)
  return reflect.Zero(arrayType)
}
```

数组中的元素类型取决于传递进来的参数 `t`。数组的长度决定于参数 `length`。要想使用 reflect 包将值提取进一个数组中，最好的办法是将数组处理为一个接口。

```go
func extractArray(v reflect.Value) (interface{}, error) {
  if v.Kind() != reflect.Array {
    return nil, errors.New("invalid input")
  }
  var array interface{}
  array = v.Interface()
  return array, nil
}
```

注意 `Slice()` 方法也可以用来提取数组的值，但是该方法需要在计算 reflect.Value 之前先将数组转换为一个可寻址的数组。为了让你的代码更简洁和可读，最好还是是使用 `Interface()` 方法。

## 通过 type signature 来创建复合信道对象
一个空的信道对象可以通过零值来创建。一个信道的零值是一个空的信道对象。下面是通过信道的 type signature 来创建一个信道的例子：

```go
func CreateCompositeObjects(t reflect.Type) reflect.Value {
  return reflect.Zero(t)
}
```
