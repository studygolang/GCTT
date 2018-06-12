已发布：https://studygolang.com/articles/12338

# Go 函数 -- Go 语言新手的带图教程

简单易懂的 Go 函数带图教程

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/funcs.png)

**注意：**该教程仅介绍 Go 函数，不包括：可变参数、延迟函数、外部函数、方法、HTTP、封包编码等。

---

## 什么是函数？

函数是一个独立的，可以被重用的，可以一次又一次运行的代码块。函数可以有输入参数，也可以有返回值输出。

## 为什么我们需要函数？

- 增加可读性、可测试性和可维护性
- 使代码的不同部分可以分别执行
- 可以由小模块组成新的模块
- 可以向类型增加行为
- 便于组织代码
- 符合 [DRY 原则](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/1.png)

声明了一个函数 “Len”，输入参数为 “s”，类型为 “string”，返回值类型为 “int”。

---

### ✪ 首先：声明一个 Len 函数

```go
func Len(s string) int {
	return utf8.RuneCountInString(s)
}
```

---

### ✪ 然后：通过它的名字调用它

```go
Len("Hello world 👋")
```

[在线运行程序](https://play.golang.org/p/6c2p1yVcMY)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 输入参数和返回值类型

输入参数被用来把数据传递给函数。返回值类型被用来从函数中返回数据。从函数中返回的数据被称为“返回值”。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/2.png)

采用一个名为 “s” 的 string 类型“输入参数”，并返回一个“返回值类型”为 int 的没有名字的返回值。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/3.png)

函数签名就是一个[函数的类型](https://golang.org/ref/spec#Function_types) -- 由输入参数类型和返回值类型组成。

---

```go
func jump()
// 签名：func()

func Len(s string) int
// 签名：func(string) int

func multiply(n ...float64) []float64
// 签名：func(...float64) []float64
```

---

Go 语言中的函数是一等公民，可以被任意赋值传递。

```go
flen := Len
flen("Hello!")
```

[在线运行程序](https://play.golang.org/p/JgE1MoO-dP)

一个函数签名的示例代码。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/4.png)

当一个函数被调用时，它的主体将以提供的输入参数运行。如果函数声明了至少一个返回值类型，那么函数将会返回一个或多个返回值。

---

你可以直接从 RuneCountInString 函数返回，因为它也返回一个 int。

```go
func Len(s string) int {
	return utf8.RuneCountInString(s)
}

lettersLen := Len("Hey!")
```

---

这个函数使用[表达式](https://golang.org/ref/spec#ExpressionList)作为返回值。

```go
func returnWithExpression(a, b int) int {
	return a * b * 2 * anotherFunc(a, b)
}
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 函数块

每一组括号都会创建一个新的函数块，任何在函数块内声明的标识符只在该函数块内可见。

```go
const message = "Hello world 👋"

func HelloWorld() {
	name := "Dennis"
	message := "Hello, earthling!"
}
```

---

```go
HelloWorld()

/*

★ message 常量在这里可见。
★ 在函数内的变量 name 在这里不可见。
★ 在函数内被隐藏的变量 message 在这里不可见。

*/
```

[在线运行程序](https://play.golang.org/p/GBw0PbDw8p)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

现在，让我们看看输入参数和返回值类型不同风格的声明方式。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/5.png)

声明一个类型为 “String” 的输入参数 “s”，和一个整数返回值类型。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/6.png)

一个函数的输入参数和返回值类型就像变量一样起作用。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## Niladic 函数

Niladic 函数不接受任何输入参数。

```go
func tick() {
	fmt.Println( time.Now().Format( time.Kitchen ) )
}

tick()

// Output: 13:50pm etc.
```

[在线运行程序](https://play.golang.org/p/D6Wnt0_mLq)

如果一个函数没有返回值，你可以省略返回值类型和 return 这个关键字。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## Singular 函数

```go
func square(n int) int {
	return n * n
}

square(4)

// Output: 16
```

[在线运行程序](https://play.golang.org/p/cJ2Q02_74h)

当函数只返回一个返回值时，不要使用括号。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 多个输入参数和返回值

```go
func scale(width, height, scale int) (int, int) {
	return width * scale, height * scale
}

w, h := scale(5, 10, 2)

// Output: w is 10, h is 20
```

[在线运行程序](https://play.golang.org/p/OULN6FZa92)

多个返回值类型应该用圆括号括起来。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 自动类型分配

Go 语言会自动为前面的参数声明类型。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/7.png)

---

这些声明是一样的：

```go
func scale(width, height, scale int) (int, int)

func scale(width int, height int, scale int) (int, int)
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 错误值

一些函数[通常](https://golang.org/doc/effective_go.html#multiple-returns)会返回错误 -- 多个返回值让这使用很方便。

```go
func write(w io.Writer, str string) (int, error) {
	return w.Write([]byte(s))
}

write(os.Stdout, "hello")

// Output: hello
```

---

从 Write 函数直接返回和返回多个返回值类型是相同的。因为它也返回一个 int 和一个错误值。

```go
func write(w io.Writer, str string) (int, error) {
	n, err := w.Write([]byte(s))
	return n, err
}
```

---

如果一切正常，你可以直接返回 nil 作为结果：

```go
func div(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("divide by zero")
	}

	return a / b, nil
}

r, err := div(-1, 0)

// err: divide by zero
```

[在线运行程序](https://play.golang.org/p/7n-scmRNy5)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 丢弃返回值

你可以使用下划线来丢弃返回值。

```go
/*
假设我们有如下函数：
*/

func TempDir(dir, prefix string) (name string, err error)
```

---

丢弃错误返回值（第 2 个返回值）：

```go
name, _ := TempDir("", "test")
```

---

丢弃全部返回值：

```go
TempDir("", "test")
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 省略参数名字

你也可以在未使用的输入参数中，把下划线当作名字使用。-- 以满足一个接口为例（或者[看这里](https://blog.cloudflare.com/quick-and-dirty-annotations-for-go-stack-traces/)）。

```go
func Write(_ []byte) (n int, err error) {
	return 0, nil
}
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/8.png)

命名的返回值参数让你可以像使用变量一样使用返回值，而且它让你可以使用一个空的 return。

---

返回值 pos 的行为就像是一个变量，函数 biggest 通过一个空的 return 返回它（return 后面没有任何表达式）。

```go
// biggest 返回切片 nums 中最大的数字的下标。
func biggest(nums []int) (pos int) {

	if len(nums) == 0 {
		return -1
	}

	m := nums[0]

	for i, n := range nums {
		if n > m {
			m = n
			pos = i
		}
	}

	// returns the pos
	return
}

pos := biggest([]int{4,5,1})

// Output: 1
```

[在线运行程序](https://play.golang.org/p/B1_uRkia_I)

上面的程序没有经过优化，时间复杂度为 O(n)。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 什么时候该使用命名返回值参数？

- 命名的返回值参数主要用作返回值的提示。
- 不要为了跳过在函数内部的变量声明而使用命名返回值参数来替代。
- 如果它使你的代码更具有可读性，请使用它。

---

当你使用命名返回值参数时，也有一个有[争议](https://news.ycombinator.com/item?id=14668323)的优化技巧，但编译器很快就会修复这个问题来禁止它的使用。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 小心变量覆盖问题

```go
func incr(snum string) (rnum string, err error) {
	var i int

	// start of a new scope
	if i, err := strconv.Atoi(snum); err == nil {
		i = i + 1
	}
	// end of the new scope

	rnum = strconv.Itoa(i)

	return
}

incr("abc")

// Output: 0 and nil
```

---

变量 i 和 err 只在 if 代码块内可见。最后，错误不应该是 “nil”，因为 “abc” 不能被转化为整数，所以这是一个错误，但是我们没有发现这个错误。

[在线运行程序](https://play.golang.org/p/tx2Rmxn3nK)

点击这里查看该问题解决方案。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 值传递

函数 pass 把输入参数的值设置为了对应的零值。

```go
func pass(s string, n int) {
	s, n = "", 0
}
```

---

我们传递两个变量给 pass 函数：

```go
str, num := "knuth", 2
pass(str, num)
```

---

函数执行完，我们的两个变量的值没有任何变化。

```go
str is "knuth"
num is 2
```

---

这是因为，当我们传递参数给函数时，参数被自动的拷贝了一份新的变量。这被叫做值传递。

[在线运行程序](https://play.golang.org/p/maAz6FR-TA)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 值传递和指针

下面这个函数接受一个指向 string 变量的指针。它修改了指针 ps 指向的值。然后它尝试将指针的值设置为 nil。所以，指针将不会再指向传递进来的 string 变量的地址。

```go
func pass(ps *string) {
	*ps = "donald"
	ps = nil
}
```

---

我们定义了一个新的变量 s，然后我们通过 & 运算符来获取它的内存地址，并将它的内存地址保存在一个新的指针变量 ps 中。

```go
s := "knuth"
ps := &s
```

---

让我们把 ps 传递给 pass 函数。

```go
pass(ps)
```

在函数运行结束之后，我们会看到变量 s 的值已经改变。但是，指针 ps 仍然指向变量 s 的有效地址。

```go
// Output:
// s : "donald"
// ps: 0x1040c130
```

---

指针 ps 是按值传递给函数 pass 的，只有它指向的地址被拷贝到了函数 pass 中的一个新的指针变量（形参）。所以，在函数里面把指针变量设置为 nil 对传递给函数做参数的指针（实参）没有影响。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/9.png)

`&s` 和 `ps` 是不同的变量，但是他们都指向相同的变量 `s`。

---

[在线运行程序](https://play.golang.org/p/ymAPKVFIdg)

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

到目前为止，我们已经学完了函数的参数声明方式。现在，让我们一起来看看如何正确的命名函数、输入参数和返回值类型。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 函数命名

使用函数的好处有增加代码的可读性和可维护性等。你可能需要根据实际情况选择性的采取这些意见。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 尽可能简短

当选择尽可能简短的命名。要选择简短、自描述而且有意义的名字。

```go
// Not this:
func CheckProtocolIsFileTransferProtocol(protocolData io.Reader) bool

// This:
func Detect(in io.Reader) Name {
	return FTP
}

// Not this:
func CreateFromIncomingJSONBytes(incomingBytesSource []byte)

// This:
func NewFromJSON(src []byte)
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 使用驼峰命名法

```go
// This:
func runServer()
func RunServer()

// Not this:
func run_server()
func RUN_SERVER()
func RunSERVER()
```

---

缩略词应该全部大写：

```go
// Not this:
func ServeHttp()

// This:
func ServeHTTP()
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 选择描述性的参数名

```go
// Not this:
func encrypt(i1, a3, b2 byte) byte

// This:
func encrypt(privKey, pubKey, salt byte) byte

// Not this:
func Write(writableStream io.Writer, bytesToBeWritten []byte)

// This:
func Write(w io.Writer, s []byte)
// 类型就非常清晰了，没有必要再取名字了
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 使用动词

```go
// Not this:
func mongo(h string) error

// This:
func connectMongo(host string) error

// 如果这个函数是在包 Mongo 内，只要这样就好了：
func connect(host string) error
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 使用 is 和 are

```go
// Not this:
func pop(new bool) item

// This:
func pop(isNew bool) item
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 不需要在命名中带上类型

```go
// Not this:
func show(errorString string)

// This:
func show(err string)
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## 使用 Getters 和 Setters

在 Go 语言中没有 Getters 和 Setters。但是，你可以通过函数来模拟。

```go
// Not this:
func GetName() string

// This:
func Name() string

// Not this:
func Name() string

// This:
func SetName(name string)
```

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

## Go 函数不支持的特性：

因为我会在即将发布的文章中说明下面问题的一些解决方法，所以你不需要去 [duckduckgo](https://duckduckgo.com/?q=does+golang+support+functions&t=hg&ia=qa) 或者 [Google](https://www.google.com.tr/search?q=does+golang+support+functions) 搜索答案。

- [函数重载](https://golang.org/doc/faq#overloading) -- 它可以通过类型断言来模拟。
- [模式匹配器函数](http://learnyouahaskell.com/syntax-in-functions)。
- 函数声明中的默认参数值。
- 在声明中按任意顺序通过名字指定输入参数。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/pLine.png)

💓 希望你能把这片文章分享给你的朋友。谢谢！

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-funcs/10.png)

---

via: https://blog.learngoprogramming.com/golang-funcs-params-named-result-values-types-pass-by-value-67f4374d9c0a

作者：[Inanc Gumus](https://blog.learngoprogramming.com/@inanc)
译者：[MDGSF](https://github.com/MDGSF)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
