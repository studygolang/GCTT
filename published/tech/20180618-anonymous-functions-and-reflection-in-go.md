首发于：https://studygolang.com/articles/14044

# Go 中的匿名函数和反射

我最近在浏览 Hacker News 时看到一篇吸引我眼球的文章《[Python中的Lambdas和函数](http://www.thepythoncorner.com/2018/05/lambdas-and-functions-in-python.html?m=1)》，这篇文章 —— 我推荐你自己阅读一下 —— 详细讲解了如何运用 Python 的 lambda 函数，并举了一个例子展示如何使用 Lambda 函数实现干净，[DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself) 风格的代码。

读这篇文章，我大脑中喜欢设计模式的部分对文章里精巧的设计模式兴奋不已，然而同时，我大脑中讨厌动态语言的部分说，“呃~”。一点简短的题外话来表达一下我对动态语言的厌恶（如果你没有同感，请略过）：

我曾经是一个动态语言的狂热粉丝（对某些任务我仍然喜欢动态语言并且几乎每天都会使用到它）。Python 是我大学一直选择的语言，我用它做科学计算并且做小的，概念验证的项目（我的个人网站曾经使用Flask）。但是当我在现实世界（[Qadium](https://www.qadium.com/)）中开始为我的第一个大型 Python 项目做贡献时，一切都变了。这些项目包含了收集，处理并且增强各种定义好的数据类型的系统职责。

最开始我们选择 Python 基于两个原因：1）早期的员工都习惯使用 2）它是一门快速开发语言。

当我开始项目时，我们刚启动了我们最早的 B2B 产品，并且有几个很早就使用 Python 的开发者参与进来。留下来的代码有几个问题：1）代码很难读 2）代码很难调试 3）代码在不破坏些东西的前提下几乎无法改变/重构。而且，代码只经过了非常少的测试。这些就是快速搭建原型系统来验证我们第一个产品价值的代价了。上述提到的问题太严重了，以至于后来大部分的开发时间都用来定位解决问题，很少有宝贵的时间来开发新功能或者修改系统来满足我们不断增长的收集和处理数据的欲望和需要。

为了解决这些问题，我和另外一些工程师开始缓慢的，用一个静态类型语言重新架构和重写系统(对我当时来说，整体的体验就像是一边开着一辆着火的车，一边还要再建造另一辆新车)。对处理系统，我们选择了 Java 语言，对数据收集系统，我们选择了 Go 语言。两年后，我可以诚实的说，使用静态语言，比如 Go（ Go 依然保留了很多动态语言的感觉，比如 Python )。

现在，我想肯定有不少读者在嘟囔着“像 Python 这样的动态语言是好的，只要把代码组织好并且测试好”这类的话了吧，我并不想较真这个观点，不过我要说的是，静态语言在解决我们系统的问题中帮了大忙，并且更适合我们的系统。当支持和修复好这边的麻烦事后，我们自己的基于 Python 的生成系统也做好了，我可以说我短期内都不想用动态语言来做任何的大项目了。

那么言归正传，当初我看到这篇文章时，
我看到里面有一些很棒的设计模式，我想试试看能否将它轻松的复制到 Go 中。如果你还没有读完[上述提及的文章](http://www.thepythoncorner.com/2018/05/lambdas-and-functions-in-python.html?m=1)，我将它用 lambda /匿名函数解决的问题引述如下：

> 假设你的一个客户要求你写一个程序来模拟“逆波兰表达式计算器”，他们会将这个程序安装到他们全体员工的电脑上。你接受了这个任务，并且获得了这个程序的需求说明：
>
> 程序能做所有的基础运算（加减乘除），能求平方根和平方运算。很明显，你应该能清空计算器的所有堆栈或者只删除最后一个入栈的数值。

如果你对逆波兰表达式（RPN）不是很熟悉，可以在维基上或者找找它最开始的论文。

现在开始解决问题，之前的文章作者提供一个可用但是极度冗余的代码。把它移植到 Go 中，就是这样的

```go
package main

import (
	"fmt"
	"math"
)

func main() {
	engine := NewRPMEngine()
	engine.Push(2)
	engine.Push(3)
	engine.Compute("+")
	engine.Compute("^2")
	engine.Compute("SQRT")
	fmt.Println("Result", engine.Pop())
}

// RPMEngine 是一个 RPN 计算引擎
type RPMEngine struct {
	stack stack
}

// NewRPMEngine 返回一个 RPMEngine
func NewRPMEngine() *RPMEngine {
	return &RPMEngine{
		stack: make(stack, 0),
	}
}

// 把一个值压入内部堆栈
func (e *RPMEngine) Push(v int) {
	e.stack = e.stack.Push(v)
}

// 把一个值从内部堆栈中取出
func (e *RPMEngine) Pop() int {
	var v int
	e.stack, v = e.stack.Pop()
	return v
}

// 计算一个运算
// 如果这个运算返回一个值，把这个值压栈
func (e *RPMEngine) Compute(operation string) error {
	switch operation {
	case "+":
		e.addTwoNumbers()
	case "-":
		e.subtractTwoNumbers()
	case "*":
		e.multiplyTwoNumbers()
	case "/":
		e.divideTwoNumbers()
	case "^2":
		e.pow2ANumber()
	case "SQRT":
		e.sqrtANumber()
	case "C":
		e.Pop()
	case "AC":
		e.stack = make(stack, 0)
	default:
		return fmt.Errorf("Operation %s not supported", operation)
	}
	return nil
}

func (e *RPMEngine) addTwoNumbers() {
	op2 := e.Pop()
	op1 := e.Pop()
	e.Push(op1 + op2)
}

func (e *RPMEngine) subtractTwoNumbers() {
	op2 := e.Pop()
	op1 := e.Pop()
	e.Push(op1 - op2)
}

func (e *RPMEngine) multiplyTwoNumbers() {
	op2 := e.Pop()
	op1 := e.Pop()
	e.Push(op1 * op2)
}

func (e *RPMEngine) divideTwoNumbers() {
	op2 := e.Pop()
	op1 := e.Pop()
	e.Push(op1 * op2)
}

func (e *RPMEngine) pow2ANumber() {
	op1 := e.Pop()
	e.Push(op1 * op1)
}

func (e *RPMEngine) sqrtANumber() {
	op1 := e.Pop()
	e.Push(int(math.Sqrt(float64(op1))))
}
```
> rpn_calc_solution1.go 由 GitHub  托管，[查看源文件](https://gist.github.com/jholliman/3f2461466ca1bc8e6b2d5c497de6c198/raw/179966bf7309e625ae304151937eedc9d3f2d067/rpn_calc_solution1.go)

注：Go 并没有一个自带的堆栈，所以，我自己创建了一个。

```go
package main

type stack []int

func (s stack) Push(v int) stack {
	return append(s, v)
}

func (s stack) Pop() (stack, int) {
	l := len(s)
	return s[:l-1], s[l-1]
}
```
> simple_stack.go 由 GitHub 托管，[查看源文件](https://gist.github.com/jholliman/f1c8ce62ce2fbeb5ec4bc48f9326266b/raw/08e0c5df28eb0c527e0d4f0b2e85c1d381f7bc7c/simple_stack.go)

（另外，这个堆栈不是线程安全的，并且对空堆栈进行 `Pop` 操作会引发 panic，除此之外，这个堆栈工作的很好）

以上的方案是可以工作的，但是有大堆的代码重复 —— 特别是获取提供给运算符的参数/操作的代码。

Python-lambda 文章对这个方案做了一个改进，将运算函数写为 lambda 表达式并且放入一个字典中，这样它们可以通过名称来引用，在运行期查找一个运算所需要操作的数值，并用普通的代码将这些操作数提供给运算函数。最终的python代码如下：

```python
"""
Engine class of the RPN Calculator
"""

import math
from inspect import signature

class rpn_engine:
    def __init__(self):
        """ Constructor """
        self.stack = []
        self.catalog = self.get_functions_catalog()

    def get_functions_catalog(self):
        """ Returns the catalog of all the functions supported by the calculator """
        return {"+": lambda x, y: x + y,
                "-": lambda x, y: x - y,
                "*": lambda x, y: x * y,
                "/": lambda x, y: x / y,
                "^2": lambda x: x * x,
                "SQRT": lambda x: math.sqrt(x),
                "C": lambda: self.stack.pop(),
                "AC": lambda: self.stack.clear()}

    def push(self, number):
        """ push a value to the internal stack """
        self.stack.append(number)

    def pop(self):
        """ pop a value from the stack """
        try:
            return self.stack.pop()
        except IndexError:
            pass # do not notify any error if the stack is empty...

    def compute(self, operation):
        """ compute an operation """

        function_requested = self.catalog[operation]
        number_of_operands = 0
        function_signature = signature(function_requested)
        number_of_operands = len(function_signature.parameters)

        if number_of_operands == 2:
            self.compute_operation_with_two_operands(self.catalog[operation])

        if number_of_operands == 1:
            self.compute_operation_with_one_operand(self.catalog[operation])

        if number_of_operands == 0:
            self.compute_operation_with_no_operands(self.catalog[operation])

    def compute_operation_with_two_operands(self, operation):
        """ exec operations with two operands """
        try:
            if len(self.stack) < 2:
                raise BaseException("Not enough operands on the stack")

            op2 = self.stack.pop()
            op1 = self.stack.pop()
            result = operation(op1, op2)
            self.push(result)
        except BaseException as error:
            print(error)

    def compute_operation_with_one_operand(self, operation):
        """ exec operations with one operand """
        try:
            op1 = self.stack.pop()
            result = operation(op1)
            self.push(result)
        except BaseException as error:
            print(error)

    def compute_operation_with_no_operands(self, operation):
        """ exec operations with no operands """
        try:
            operation()
        except BaseException as error:
            print(error)

```
> engine_peter_rel5.py 由GitHub托管 [查看源文件](https://gist.github.com/mastro35/66044197fa886bf842213ace58457687/raw/a84776f1a93919fe83ec6719954384a4965f3788/engine_peter_rel5.py)

这个方案比原来的方案只增加了一点点复杂度，但是现在增加一个新的运算符简直就像增加一条线一样简单！我看到这个的第一个想法就是：我怎么在 Go 中实现？

我知道在 Go 中有[函数字面量](https://golang.org/ref/spec#Function_literals)，它是一个很简单的东西，就像在 Python 的方案中，创建一个运算符的名字与运算符操作的 map。它可以这么被实现：

```go
package main

import "math"

func main() {
	catalog := map[string]interface{}{
		"+":    func(x, y int) int { return x + y },
		"-":    func(x, y int) int { return x - y },
		"*":    func(x, y int) int { return x * y },
		"/":    func(x, y int) int { return x / y },
		"^2":   func(x int) int { return x * x },
		"SQRT": func(x int) int { return int(math.Sqrt(float64(x))) },
		"C":    func() { /* TODO: need engine object */ },
		"AC":   func() { /* TODO: need engine object */ },
	}
}
view rawrpn_operations_map.go hosted with ❤ by GitHub
```
> rpn_operations_map.go 由gitHub托管 [查看源文件](https://gist.github.com/jholliman/9108b105be6ab136c2f163834b9e5e32/raw/bcd7634a1336b464a9957ea5f32a4868071bef6e/rpn_operations_map.go)

注意：在 Go 语言中，为了将我们所有的匿名函数保存在同一个 map 中，我们需要使用空接口类型，`interfa{}`。在 Go 中所有类型都实现了空接口（它是一个没有任何方法的接口；所有类型都至少有 0 个函数）。在底层，Go 用两个指针来表示一个接口：一个指向值，另一个指向类型。

识别接口实际保存的类型的一个方法是用 `.(type)` 来做断言，比如：

```go
package main

import (
	"fmt"
	"math"
)

func main() {
	catalog := map[string]interface{} {
		"+":    func(x, y int) int { return x + y },
		"-":    func(x, y int) int { return x - y },
		"*":    func(x, y int) int { return x * y },
		"/":    func(x, y int) int { return x / y },
		"^2":   func(x int) int { return x * x },
		"SQRT": func(x int) int { return int(math.Sqrt(float64(x))) },
		"C":    func() { /* TODO: need engine object */ },
		"AC":   func() { /* TODO: need engine object */ },
	}

	for k, v := range catalog {
		switch v.(type) {
		case func(int, int) int:
			fmt.Printf("%s takes two operands\n", k)
		case func(int) int:
			fmt.Printf("%s takes one operands\n", k)
		case func():
			fmt.Printf("%s takes zero operands\n", k)
		}
	}
}
```
> rpn_operations_map2.go 由GitHub托管    [查看源文件](https://gist.github.com/jholliman/497733a937fa5949148a7160473b7742/raw/7ec842fe18026bfc1c4d801eca566b4c0541008b/rpn_operations_map2.go)

这段代码会产生如下输出（请原谅语法上的瑕疵）：

```
SQRT takes one operands
AC takes zero operands
+ takes two operands
/ takes two operands
^2 takes one operands
- takes two operands
* takes two operands
C takes zero operands
```

这就揭示了一种方法，可以获得一个运算符需要多少个操作数，以及如何复制 Python 的解决方案。但是我们如何能做到更好？我们能否为提取运算符所需参数抽象出一个更通用的逻辑？我们能否在不用 `if` 或者 `switch` 语句的情况下，查找一个函数所需要的操作数的个数并且调用它？实际上通过 Go 中的 `relect` 包提供的反射功能，我们是可以做到的。

对于 Go 中的反射，一个简要的说明如下：

在 Go 中，通常来讲，如果你需要一个变量，类型或者函数，你可以定义它然后使用它。然而，如果你发现你是在运行时需要它们，或者你在设计一个系统需要使用多种不同类型（比如，实现运算符的函数 —— 它们接受不同数量的变量，因此是不同的类型），那么你可以就需要使用反射。反射给你在运行时检查，创建和修改不同类型的能力。如果需要更详尽的 Go 的反射说明以及一些使用 `reflect` 包的基础知识，请参阅[反射的规则](https://blog.golang.org/laws-of-reflection)这篇博客。

下列代码演示了另一种解决方法，通过反射来实现查找我们匿名函数需要的操作数的个数：

```go
package main

import (
	"fmt"
	"math"
	"reflect"
)

func main() {
	catalog := map[string]interface{}{
		"+":    func(x, y int) int { return x + y },
		"-":    func(x, y int) int { return x - y },
		"*":    func(x, y int) int { return x * y },
		"/":    func(x, y int) int { return x / y },
		"^2":   func(x int) int { return x * x },
		"SQRT": func(x int) int { return int(math.Sqrt(float64(x))) },
		"C":    func() { /* TODO: need engine object */ },
		"AC":   func() { /* TODO: need engine object */ },
	}

	for k, v := range catalog {
		method := reflect.ValueOf(v)
		numOperands := method.Type().NumIn()
		fmt.Printf("%s has %d operands\n", k, numOperands)
	}
}
```
> rpn_operations_map3.go 由GitHub托管 [查看源文件](https://gist.github.com/jholliman/e3d7abd71b9bf6cb71eb55d49c40b145/raw/ed64e1d3f718e4219c68d189a8149d8179cdc90c/rpn_operations_map3.go)

类似与用 `.(type)` 来切换的方法，代码输出如下：

```
^2 has 1 operands
SQRT has 1 operands
AC has 0 operands
* has 2 operands
/ has 2 operands
C has 0 operands
+ has 2 operands
- has 2 operands
```

现在我不再需要根据函数的签名来硬编码参数的数量了！

注意：如果值的种类（[种类（Kind）](https://golang.org/pkg/reflect/#Kind) 不要与类型弄混了））不是 `Func`，调用 `toNumIn` 会触发 `panic`，所以小心使用，因为 panic 只有在运行时才会发生。

通过检查 Go 的 `reflect` 包，我们知道，如果一个值的种类(Kind)是 Func 的话，我们是可以通过调用 [`Call`](https://golang.org/pkg/reflect/#Value.Call) 方法，并且传给它一个 值对象的切片来调用这个函数。比如，我们可以这么做：

```go
package main

import (
	"fmt"
	"math"
	"reflect"
)

func main() {
	catalog := map[string]interface{}{
		"+":    func(x, y int) int { return x + y },
		"-":    func(x, y int) int { return x - y },
		"*":    func(x, y int) int { return x * y },
		"/":    func(x, y int) int { return x / y },
		"^2":   func(x int) int { return x * x },
		"SQRT": func(x int) int { return int(math.Sqrt(float64(x))) },
		"C":    func() { /* TODO: need engine object */ },
		"AC":   func() { /* TODO: need engine object */ },
	}

	method := reflect.ValueOf(catalog["+"])
	operands := []reflect.Value{
		reflect.ValueOf(3),
		reflect.ValueOf(2),
	}

	results := method.Call(operands)
	fmt.Println("The result is ", int(results[0].Int()))
}
```
> rpn_operations_map4.go 由GitHub托管 [查看源文件](https://gist.github.com/jholliman/96bccb0c73c75a165211892da87cd676/raw/e4b420f77350e3f7e081dddb20c0d2a7232cc071/rpn_operations_map4.go)

就像我们期待的那样，这段代码会输出：

```
The result is  5
```

酷！

现在我们可以写出我们的终极解决方案了：

```go
package main

import (
	"fmt"
	"math"
	"reflect"
)

func main() {
	engine := NewRPMEngine()
	engine.Push(2)
	engine.Push(3)
	engine.Compute("+")
	engine.Compute("^2")
	engine.Compute("SQRT")
	fmt.Println("Result", engine.Pop())
}

// RPMEngine 是一个 RPN 计算引擎
type RPMEngine struct {
	stack   stack
	catalog map[string]interface{}
}

// NewRPMEngine 返回 一个 带有 缺省功能目录的 RPMEngine
func NewRPMEngine() *RPMEngine {
	engine := &RPMEngine{
		stack: make(stack, 0),
	}
	engine.catalog = map[string]interface{}{
		"+":    func(x, y int) int { return x + y },
		"-":    func(x, y int) int { return x - y },
		"*":    func(x, y int) int { return x * y },
		"/":    func(x, y int) int { return x / y },
		"^2":   func(x int) int { return x * x },
		"SQRT": func(x int) int { return int(math.Sqrt(float64(x))) },
		"C":    func() { _ = engine.Pop() },
		"AC":   func() { engine.stack = make(stack, 0) },
	}
	return engine
}

// 将一个值压入内部堆栈
func (e *RPMEngine) Push(v int) {
	e.stack = e.stack.Push(v)
}

// 从内部堆栈取出一个值
func (e *RPMEngine) Pop() int {
	var v int
	e.stack, v = e.stack.Pop()
	return v
}

// 计算一个运算
// 如果这个运算返回一个值，把这个值压栈
func (e *RPMEngine) Compute(operation string) error {
	opFunc, ok := e.catalog[operation]
	if !ok {
		return fmt.Errorf("Operation %s not supported", operation)
	}

	method := reflect.ValueOf(opFunc)
	numOperands := method.Type().NumIn()
	if len(e.stack) < numOperands {
		return fmt.Errorf("Too few operands for requested operation %s", operation)
	}

	operands := make([]reflect.Value, numOperands)
	for i := 0; i < numOperands; i++ {
		operands[numOperands-i-1] = reflect.ValueOf(e.Pop())
	}

	results := method.Call(operands)
	// If the operation returned a result, put it on the stack
	if len(results) == 1 {
		result := results[0].Int()
		e.Push(int(result))
	}

	return nil
}
```
> rpn_calc_solution2.go 由 GitHub托管 [查看源文件](https://gist.github.com/jholliman/c636340cac9253da98efdbfcfde56282/raw/b5f80bb79164b5250b088c6df094ca4ec9840009/rpn_calc_solution2.go)

我们确定操作数的个数（第 64 行），从堆栈中获得操作数（第 69-72 行），然后调用需要的运算函数，而且对不同参数个数的运算函数的调用都是一样的（第 74 行）。而且与 Python 的解决方案一样，增加新的运算函数，只需要往 map 中增加一个匿名函数的条目就可以了（第 30 行）。

总结一下，我们已经知道如果使用匿名函数和反射将一个 Python 中有趣的设计模式复制到 Go 语言中来。对反射的过度使用我会保持警惕，反射增加了代码的复杂度，而且我经常看到反射用于绕过一些坏的设计。另外，它会将一些本该在编译期发生的错误变成运行期错误，而且它还会显著拖慢程序（即将推出：两种方案的基准检查） —— 通过[检查源码](https://golang.org/src/reflect/value.go?s=9676:9715#L295)，看起来 `reflect.Value.Call` 执行了很多的准备工作，并且对每一次调用 都为它的 `[]reflect.Value` 返回结果参数分配了新的切片。也就是说，如果性能不需要关注 —— 在 Python 中通常都不怎么关注性能;)，有足够的测试，而且我们的目标是优化代码的长度，并且想让它易于加入新的运算，那么反射是一个值得推荐的方法。

---

via: https://medium.com/@jhh3/anonymous-functions-and-reflection-in-go-71274dd9e83a

作者：[John Holliman](https://medium.com/@jhh3)
译者：[MoodWu](https://github.com/MoodWu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
