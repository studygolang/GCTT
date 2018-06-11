已发布：https://studygolang.com/articles/12526

# Go 反射：根据类型创建对象-第二部分（复合类型）

> 这是关于 Golang 中根据类型创建对象系列博客的第二篇，讨论的是创建复合对象。第一篇在[这里](https://studygolang.com/articles/12434)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-reflect/cover1.png)

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

// 译者注：上面代码似乎有错误，改成了下面所示（亦或是故意省略包名前缀）

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

数组中的元素类型取决于传递进来的参数 `t`。数组的长度决定于参数 `length`。要想使用 reflect 包将值提取进一个数组中，最好的办法是将数组作为接口处理。

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

该函数会创建一个包含一个任意空复合对象的 `reflect.Value` 类型的结构体。

reflect 包有两个方法来创建一个信道。`ChanOf` 是用来创建信道的 type signature，`MakeChan(Type, int)` 方法可以用来给信道分配内存。下面是一个例子：

```go
func CreateChan(t reflect.Type, buffer int) reflect.Value {
	chanType := reflect.ChanOf(reflect.BothDir, t)
	return reflect.MakeChan(chanType, buffer)
}
```

信道中元素的类型取决于传递进来的参数 `t`。信道的容量取决于参数 `buffer`。要想使用 reflect 包将值提取进一个信道中，最好的办法是将信道作为接口处理。信道的方向通过传入 `ChanOf` 的第一个参数来控制，可能的取值有：

```go
SendDir
RecvDir
BothDir
```

`BothDir` 表明信道既可读又可写。`SendDir` 表明信道只能写。`RecvDir` 表明信道只能读。

```go
func extractChan(v reflect.Value) (interface{}, error) {
	if v.Kind() != reflect.Chan {
		return nil, errors.New("invalid input")
	}
	var ch interface{}
	ch = v.Interface()
	return ch, nil
}
```

## 通过 type signature 来创建复合函数对象

函数对象不能使用零值来创建。

reflect 包有两个方法来创建函数。`FuncOf` 方法是用于创建函数的 type signature，`MakeFunc(Type, func(args []Value) (results []Value)) Value` 方法可以用来给函数分配内存。下面是一个例子：

```go
func CreateFunc(
	fType reflect.Type,
	f func(args []reflect.Value) (results []reflect.Value)
)(reflect.Value, error) {
	 if fType.Kind() != reflect.Func {
		return reflect.Value{}, errors.New("invalid input")
	 }

	var ins, outs *[]reflect.Type

	ins = new([]reflect.Type)
	outs = new([]reflect.Type)

	for i := 0; i < fType.NumIn(); i++ {
		*ins = append(*ins, fType.In(i))
	}

	for i := 0; i < fType.NumOut(); i++ {
		*outs = append(*outs, fType.Out(i))
	}
	var variadic bool
	variadic = fType.IsVariadic()
	return AllocateStackFrame(*ins, *outs, variadic, f), nil
}

func AllocateStackFrame(
	ins []reflect.Type,
	outs []reflect.Type,
	variadic bool,
	f func(args []reflect.Value) (results []reflect.Value)
) reflect.Value {
	 var funcType reflect.Type
	 funcType = reflect.FuncOf(ins, outs, variadic)
	 return reflect.MakeFunc(funcType, f)
}
```

`CreateFunc` 有两个参数。第一个参数是你想创建的函数的 `type`，第二个是一个具体的函数，该函数实现了第一个参数的类型，但是其输入和输出都转换为了 `reflect.Value` 对象。我已经创建好了这个函数，所以关于如何动态创建函数你就不用再重新造轮子了。

为了使用它来创建函数，函数的 type signature 需要事先定义。比如：

```go
type fn func(int) (int)
```

描述函数 signature 的 Type 结构体可以通过下面方法来创建：

```go
var funcVar fn
var funcType reflect.Type
funcType = reflect.TypeOf(funcVar)
```

`funcType` 可以作为 `CreateFunc` 的第一个参数传入。满足 `funcType` 的 type signature 的函数可以作为第二个参数传入。比如：

```go
func doubler (input int) (int) {
	return input * 2
}
```

函数 `doubler` 将输入值乘以 2，它接收一个 `int` 类型的参数并返回一个 `int` 类型的值。该函数满足 `fn` 类型，但是不满足通用的类型：

```go
f func (args []reflect.Value) (results []reflect.Value)
```

为了满足通用的类型，需要一个等价版的 `doubler`，如下所示：

```go
func doublerReflect(args []reflect.Value) (result []reflect.Value) {
	if len(args) != 1 {
		panic(fmt.Sprintf("expected 1 arg, found %d", len(args)))
	}
	if args[0].Kind() != reflect.Int {
		panic(fmt.Sprintf("expected 1 arg of kind int, found 1 args of kind", args[0].Kind()))
	}

	var intVal int64
	intVal = args[0].Int()

	var doubleIntVal int
	doubleIntVal = doubler(int(intVal))

	var returnValue reflect.Value
	returnValue = reflect.ValueOf(doubleIntVal)

	return []reflect.Value{returnValue}
}
```

`doublerReflect` 在功能上相当于 `doubler`，它满足通用函数的 type signature。也就是说，它需要 1 个 `reflect.Value` 切片作为参数。并返回 1 个 `reflect.Value` 切片的值。输入表示函数的输入，返回值表示正在生成的新函数的返回值。

调用 `CreateFunc` 可以用 `funcType` 和 `doublerReflect` 作为参数。下面的示例显示了对新创建的函数的调用。

```go
func main() {
	var funcVar fn
	var funcType reflect.Type
	funcType = reflect.TypeOf(funcVar)
	v, err := CreateFunc(funcType, doublerReflect)
	if err != nil {
		fmt.Printf("Error creating function %v\n", err)
	}

	input := 42

	ins := []reflect.Value([]reflect.Value{reflect.ValueOf(input)})

	outs := v.Call(ins)
	for i := range outs {
		fmt.Printf("%+v\n", outs[i].Interface())
	}
}
// Output: 84
```

## 通过 type signature 来创建复合映射对象

一个空的映射对象可以通过零值来创建。一个映射的零值是一个空的映射对象。下面是通过映射的 type signature 来创建一个映射的例子：

```go
func CreateCompositeObjects(t reflect.Type) reflect.Value {
	return reflect.Zero(t)
}
```

该函数会创建一个包含一个任意空复合对象的 `reflect.Value` 类型的结构体。

reflect 包有一个 `MapOf(Type, Type)` 方法，该方法可以被用于创建映射，映射键的类型是传入的第一个 `type` ，值的类型为第二个 `type`。下面是一个示例：

```go
func CreateMap(key , elem reflect.Type) reflect.Value {
	var mapType reflect.Type
	mapType = reflect.MapOf(key, elem)
	return reflect.MakeMap(mapType)
}
```

要想使用 reflect 包将值提取进一个映射对象中，最好的办法是将映射作为接口处理。

```go
func extractMap(v reflect.Value) (interface{}, error) {
	if v.Kind() != reflect.Map {
		return nil, errors.New("invalid input")
	}
	var mapVal interface{}
	mapVal = v.Interface()
	return mapVal, nil
}
```

注意到映射同样可以使用 `MakeMapWithSize` 来分配空间。使用该方法的步骤同上面一样，只是 `MakeMap` 可以用 `MakeMapWithSize` 来代替并且还需要传入一个大小的参数。

## 通过 type signature 来创建复合指针对象

一个空的指针对象可以通过零值来创建。一个指针的零值是一个 `nil` 的指针对象。下面是通过指针的 type signature 来创建一个指针的例子：

```go
func CreateCompositeObjects(t reflect.Type) reflect.Value {
	return reflect.Zero(t)
}
```

该函数会创建一个包含一个任意空复合对象的 `reflect.Value` 类型的结构体。

reflect 包有一个 `PtrTo(Type)` 方法，可以用于创建指向 `Type` 类型的指针。比如：

```go
func CreatePtr(t reflect.Type) reflect.Value {
	var ptrType reflect.Type
	ptrType = reflect.PtrTo(t)
	return reflect.Zero(ptrType)
}
```

注意到上面的功能同样可以使用 `reflect.New(t)` 来实现。

指针所指向的元素的类型决定于传入参数 `t`。要想使用 reflect 包将值提取进一个对象中，最好的办法是首先使用 `Indirect()` 或者 `Elem()` 方法来得到指针所指向对象的值，然后将该值作为接口处理。

```go
func extractElement(v reflect.Value) (interface{}, error) {
	if v.Kind() != reflect.Ptr {
		return nil, errors.New("invalid input")
	}
	var elem reflect.Value
	elem = v.Elem()

	var elem interface{}
	elem = v.Interface()
	return elem, nil
}
```

## 通过 type signature 来创建复合切片对象

一个空的切片对象可以通过零值来创建。一个切片的零值是一个空的切片对象。下面是通过切片的 type signature 来创建一个切片对象的例子：

```go
func CreateCompositeObjects(t reflect.Type) reflect.Value {
	return reflect.Zero(t)
}
```

该函数会创建一个包含一个任意空复合对象的 `reflect.Value` 类型的结构体。

reflect 包有一个 `SliceOf(Type)` 的方法可以用于创建包含 `Type` 类型的切片。下面是使用方法：

```go
func CreateSlice(t reflect.Type) reflect.Value {
	var sliceType reflect.Type
	sliceType = reflect.SliceOf(length, t)
	return reflect.Zero(sliceType)
}
```

切片中元素的类型由传入的参数 `t` 决定。要想使用 reflect 包将值提取进一个切片对象中，最好的办法是将该切片作为接口处理。

```go
func extractSlice(v reflect.Value) (interface{}, error) {
	if v.Kind() != reflect.Slice {
		return nil, errors.New("invalid input")
	}
	var slice interface{}
	slice = v.Interface()
	return slice, nil
}
```

## 通过 type signature 来创建复合结构体对象

一个空的结构体对象可以通过零值来创建。一个结构体的零值是一个空的结构体对象。下面是通过结构体的 type signature 来创建一个结构体对象的例子：

```go
func CreateCompositeObjects(t reflect.Type) reflect.Value {
 return reflect.Zero(t)
}
```

该函数会创建一个包含一个任意空复合对象的 `reflect.Value` 类型的结构体。

reflect 包有一个 `StructOf([]reflect.StructFields)` 的方法可以用于创建包含 `StructField` 中定义的字段类型的结构体。下面是使用方法：

```go
func CreateStruct(fields []reflect.StructField) reflect.Value {
	var structType reflect.Type
	structType = reflect.StructOf(fields)
	return reflect.Zero(structType)
}
```

要想使用 reflect 包将值提取进一个结构体对象中，最好的办法是将该结构体作为接口处理。

```go
func extractStruct(v reflect.Value) (interface{}, error) {
	if v.Kind() != reflect.Struct {
		return nil, errors.New("invalid input")
	}
	var st interface{}
	st = v.Interface()
	return st, nil
}
```

## 结论

这是一个关于使用 reflect 包来动态创建任意 Go 类型对象的完整教程。我提供了创建 `Func` 类型的便利方法，因为它比其他类型更复杂，如果不仔细设计，很容易污染您的代码库。
请继续关注我的下一篇博文，将任何类型转换为 Golang 中的其他类型！接下来，我将解释如何在 Golang 中编写 JIT （即时），然后介绍使用 reflect 来生成代码。

---
via：https://medium.com/kokster/go-reflection-creating-objects-from-types-part-ii-composite-types-69a0e8134f20

作者：[Sidhartha Mani](https://medium.com/@utter_babbage)
译者：[ParadeTo](https://github.com/ParadeTo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
