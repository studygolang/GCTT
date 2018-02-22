# Go 反射：根据类型创建对象-第一部分（原始类型）  
> 这是关于在 Go 中根据类型创建对象的博客系列两部分的第一部分。这部分讨论原始类型的对象创建  

Go 中的 reflect 包提供了根据执行过程中对象的类型来改变程序控制流的 API。  

reflect 包提供了两个重要的结构 - Type 和 Value。  

Type 是一个 Go 中任何类型的代表。换句话说，它可以被用于编码任何 Go 类型（例如：int , string , bool , myCustomType 等等）。Value 是一个 Go 中任何值的代表。换句话说，它可以被用于编码、操作任何 Go 的值。

## 类型与类别(Types 和 Kinds)
Go 中有一个隐蔽的、鲜为人知的，使得 Type 和 Kind 含义有差别的公约。这种差别可以通过一个例子理解一下。
看一下这个结构：

```go
type example struct {
  field1 type1
  field2 type2
}
```

这个结构体的一个对象的 type 应该是 example。而这个对象的 kind 应该是 struct。这里的 Kind 可以被看成一个 Type 的 Type。

> 在 Go 里所有 structs 都是相同的 kind，但不是相同的 Type  

像Pointer、Array、Slice、Map 等等复杂类型，使得 type 和 kind 的含义产生了这样的差异。  

相比之下，像 int、float、string 等等原始类型，并没有产生 type 和 kind 含义上的差别。换句话说，一个 int 变量的 kind 是 int。一个 int 变量的 type 也是 int。

## 根据类型创建对象
为了根据一个类型标签（type signature）创建一个对象，这个对象的 type 和 kind 都是必要的。从这里开始，当我用到‘type signature’这个术语时，我的意思就是 Go 里的 reflect.Type 类型的对象。

### 根据原始类型创建原始对象

原始对象可以根据它们的 type signature，通过使用它们的 Zero 值被创建。

> 一个类型的 Zero 值是指一个该类型、没有初始化的对象的值

这是一个 Go 中所有原始类型的列表：

```
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

通过使用 reflect.Zero 函数，可以创建原始类型的对象

```go
func CreatePrimitiveObjects(t reflect.Type) reflect.Value {
  return reflect.Zero(t)
}
```

这里将创建一个期望的对象，并返回一个拥有相应 Zero 值的 reflect.Value 对象。为了让这个对象可以使用，需要将它的值提取一下。

对于不同的原始类型，相应对象的值可以采用相应适合的方法来提取。

### 提取整数值

在 Go 里有 5 种整数类型：

```
        Int
        Int8
        Int16
        Int32
        Int64
```

Int 类型表示平台定义的默认整型大小。另外 4 种类型分别是 8，16，32，64（bit单位）大小的整型。

为了获得每种不同的整数类型，reflect.Value 对象对应的整数需要被转换成相应的整数类型。

这里是如何提取 int32 类型：

```go
// Extract Int32
func extractInt32(v reflect.Value) (int32, error) {
  if reflect.Kind() != reflect.Int32 {
    return int32(0), errors.New("Invalid input")
  }
  var intVal int64
  intVal = v.Int()
  return int32(intVal), nil
}
```

值得注意的是：reflect.Int() 会返回 int64 类型。这是因为 int64 类型可以编码成其他所有整数类型。

这里是剩余其他整数类型的提取：

```go
// Extract Int64
func extractInt64(v reflect.Value) (int64, error) {
  if reflect.Kind() != reflect.Int64 {
    return int64(0), errors.New("Invalid input")
  }
  var intVal int64
  intVal = v.Int()
  return intVal, nil
}
// Extract Int16
func extractInt16(v reflect.Value) (int16, error) {
  if reflect.Kind() != reflect.Int16 {
    return int16(0), errors.New("Invalid input")
  }
  var intVal int64
  intVal = v.Int()
  return int16(intVal), nil
}
// Extract Int8
func extractInt8(v reflect.Value) (int8, error) {
  if reflect.Kind() != reflect.Int8 {
    return int8(0), errors.New("Invalid input")
  }
  var intVal int64
  intVal = v.Int()
  return int8(intVal), nil
}
// Extract Int
func extractInt(v reflect.Value) (int, error) {
  if reflect.Kind() != reflect.Int {
    return int(0), errors.New("Invalid input")
  }
  var intVal int64
  intVal = v.Int()
  return int(intVal), nil
}
```

### 提取布尔值

布尔值在 reflect 包中用常量 Bool 表示。

它们可以通过使用 Bool() 方法，从 reflect.Value 对象中提取出来：

```go
// Extract Bool
func extractBool(v reflect.Value) (bool, error) {
  if reflect.Kind() != reflect.Bool {
    return false, errors.New("Invalid input")
  }
  return v.Bool(), nil
}
```

### 提取无符号整数

在 Go 中有 5 种无符号整数类型：

```
        Uint
        Uint8
        Uint16
        Uint32
        Uint64
```

Uint 类型表示平台定义的默认无符号整型大小。另外 4 种类型分别是 8，16，32，64（bit单位）大小的无符号整型。

为了获得每种不同的无符号整数类型，reflect.Value 对象对应的无符号整数需要被转换成相应的无符号整数类型。

这里是如何提取 Uint32 类型：

```go
// Extract Uint32
func extractUint32(v reflect.Value) (uint32, error) {
  if reflect.Kind() != reflect.Uint32 {
    return uint32(0), errors.New("Invalid input")
  }
  var uintVal uint64
  uintVal = v.Uint()
  return uint32(uintVal), nil
}
```

值得注意的是：reflect.Uint() 会返回 uint64 类型。这是因为 uint64 类型可以编码成其他所有整数类型。

这里是剩余其他无符号整数类型的提取：

```go
// Extract Uint64
func extractUint64(v reflect.Value) (uint64, error) {
  if reflect.Kind() != reflect.Uint64 {
    return uint64(0), errors.New("Invalid input")
  }
  var uintVal uint64
  uintVal = v.Uint()
  return uintVal, nil
}
// Extract Uint16
func extractUint16(v reflect.Value) (uint16, error) {
  if reflect.Kind() != reflect.Uint16 {
    return uint16(0), errors.New("Invalid input")
  }
  var uintVal uint64
  uintVal = v.Uint()
  return uint16(uintVal), nil
}
// Extract Uint8
func extractUint8(v reflect.Value) (uint8, error) {
  if reflect.Kind() != reflect.Uint8 {
    return uint8(0), errors.New("Invalid input")
  }
  var uintVal uint64
  uintVal = v.Uint()
  return uint8(uintVal), nil
}
// Extract Uint
func extractUint(v reflect.Value) (uint, error) {
  if reflect.Kind() != reflect.Uint {
    return uint(0), errors.New("Invalid input")
  }
  var uintVal uint64
  uintVal = v.Uint()
  return uint(uintVal), nil
}
```

### 提取浮点数

在 Go 中有 2 种浮点数类型：

```
        Float32
        Float64
```

Float32 类型表示 32bit 大小的浮点数。  
Float64 类型表示 64bit 大小的浮点数。

为了获得每种不同的浮点数类型，reflect.Value 对象对应的浮点数需要被转换成相应的浮点数类型。

这里是如何提取 Float32 类型：

```go
// Extract Float32
func extractFloat32(v reflect.Value) (float32, error) {
  if reflect.Kind() != reflect.Float32 {
    return float32(0), errors.New("Invalid input")
  }
  var floatVal float64
  floatVal = v.Float()
  return float32(floatVal), nil
}
```

值得注意的是：reflect.Float() 会返回 float64 类型。这是因为 float64 类型可以编码成其他所有浮点数类型。

这里是如何提取 64-bit 的浮点数值：

```go
// Extract Float64
func extractFloat64(v reflect.Value) (float64, error) {
  if reflect.Kind() != reflect.Float64 {
    return float64(0), errors.New("Invalid input")
  }
  var floatVal float64
  floatVal = v.Float()
  return floatVal, nil
}
```

### 提取复数值

在 Go 中有 2 种复数类型：

```
        Complex64
        Complex128
```

Complex64 类型表示 64bit 大小的复数。Complex128 类型表示 128bit 大小的复数。

为了获得每种不同的复数类型，reflect.Value 对象对应的复数需要被转换成相应的复数类型。

这里是如何提取 Complex64 类型：

```go
// Extract Complex64
func extractComplex64(v reflect.Value) (complex64, error) {
  if reflect.Kind() != reflect.Complex64 {
    return complex64(0), errors.New("Invalid input")
  }
  var complexVal complex128
  complexVal = v.Complex()
  return complex64(complexVal), nil
}
```

值得注意的是：reflect.Complex() 会返回 complex128 类型。这是因为 complex128 类型可以编码成其他所有复数类型。

这里是如何提取 128-bit 的复数值：

```go
// Extract Complex128
func extractComplex128(v reflect.Value) (complex128, error) {
  if reflect.Kind() != reflect.Complex128 {
    return complex128(0), errors.New("Invalid input")
  }
  var complexVal complex128
  complexVal = v.Complex()
  return complexVal, nil
}
```

### 提取字符串值

在 reflect 包中，字符串值用常量 String 表示。

它们可以通过使用 String() 方法，从 reflect.Value 对象中提取出来：

这里是如何提取 String 类型：

```go
// Extract String
func extractString(v reflect.Value) (string, error) {
  if reflect.Kind() != reflect.String {
    return "", errors.New("Invalid input")
  }
  return v.String(), nil
}
```

### 提取指针值

在 Go 中，有 2 种指针类型：

```
        Uintptr
        UnsafePointer
```

Uintptr 和 UnsafePointer 其实是程序内存中代表一个虚拟地址的 uint 值。它可以表示一个变量或函数的位置。

Uintptr 和 UnsafePointer 两者间的不同在于：Uintptr 会在 Go 运行时进行类型校验，而 UnsafePointer 不会。UnsafePointer 可以被用于 Go 中任意类型向 Go 中其他任何拥有相同内存结构的类型转换。如果这是你想要探索的，请在下面评论，我会写更多关于它的东西。

Uintptr 和 UnsafePointer 可以分别通过 Addr() 和 UnsafeAddr() 方法，从 reflect.Value 的对象中提取出来。这是一个展示 Uintptr 提取的例子:

```go
// Extract Uintptr
func extractUintptr(v reflect.Value) (uintptr, error) {
  if reflect.Kind() != reflect.Uintptr {
    return uintptr(0), errors.New("Invalid input")
  }
  var ptrVal uintptr
  ptrVal = v.Addr()
  return ptrVal, nil
}
```

这是如何提取 UnsafePointer 类型：

```go
// Extract UnsafePointer
func extractUnsafePointer(v reflect.Value) (unsafe.Pointer, error) {
  if reflect.Kind() != reflect.UnsafePointer {
    return unsafe.Pointer(0), errors.New("Invalid input")
  }
  var unsafeVal unsafe.Pointer
  unsafeVal = unsafe.Pointer(v.UnsafeAddr())
  return unsafeVal, nil
}
```

值得注意的是：上面 v.UnsafeAddr() 会返回 uintptr 值。它应该在同一行进行类型转换，否则这个 unsafe.Pointer 的值不一定指向预期的位置。

## 接下来是什么
请注意：reflect.Value 结构的所有方法在使用时都需要检验它们的 kind，否则很容易引发 panic。

在下一篇博客中，我会写更多像 struct、pointer、chan、map、slice、array 等复杂类型对象的创建。敬请期待！


----------------

via: https://medium.com/kokster/go-reflection-creating-objects-from-types-part-i-primitive-types-6119e3737f5d

作者：[Sidhartha Mani](https://medium.com/@utter_babbage)  
译者：[yiyulantian](https://github.com/yiyulantian)  
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出