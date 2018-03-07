已发布：https://studygolang.com/articles/12396

# Go 逃逸分析的缺陷

## 序

先阅读这个由四部分组成的系列文章，对理解逃逸分析和数据语义会有帮助。下面详细介绍了阅读逃逸分析报告和 pprof 输出的方法。（GCTT 已经在翻译中）

<https://www.ardanlabs.com/blog/2017/05/language-mechanics-on-stacks-and-pointers.html>

## 介绍

即使使用了 Go 4 年，我仍然会被这门语言震惊到。多亏了编译器执行的静态代码分析，编译器可以对其生成的代码进行一些有趣的优化。编码器执行的其中一种分析称为逃逸分析。这会对内存管理进行优化和简化。

在过去的两年中，（Go）语言团队一直致力于优化编译器生成的代码，以获得更好的性能，并且做了极其出色的工作。我相信，如果逃逸分析中的一些现有的缺陷得以解决，那么，Go 程序会得到更大的改进。早在 2015 年 2 月，Dmitry Vyukov 就撰写了这篇文章，概述了编译器已知的逃逸分析缺陷。

https://docs.google.com/document/d/1CxgUBPlx9iJzkz9JWkb6tIpTe5q32QDmz8l0BouG0Cw/edit

我很好奇，自这篇文章以来，当中有多少提及的缺陷被修复了，然后，我发现，迄今为止，只有少数一些缺陷得到了解决。也就是说，有五个特定的缺陷尚未被修复，而我很乐意看到在 Go 不久的将来发布的版本中能有所改善。我将这些缺陷标记为：

  * 间接赋值（Indirect Assignment）
  * 间接调用（Indirect Call）
  * 切片和 Map 赋值（Slice and Map Assignments）
  * 接口（Interfaces）
  * 未知缺陷（Unknown）

我认为，探索这些缺陷是很有趣的，所以，你可以看到它们被修复后对现有 Go 程序的积极影响。下面你所看到的所有东西都基于 1.9 编译器。

## 间接赋值

“间接赋值（Indirect Assignment）”缺陷与通过间接分配值时发生的分配有关。这是一个代码示例：

**代码清单 1**  
<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example1/example1_test.go>

```go
package flaws

import "testing"

func BenchmarkAssignmentIndirect(b *testing.B) {
	type X struct {
		p *int
	}
	for i := 0; i < b.N; i++ {
		var i1 int
		x1 := &X{
			p: &i1, // GOOD: i1 does not escape
		}
		_ = x1

		var i2 int
		x2 := &X{}
		x2.p = &i2 // BAD: Cause of i2 escape
	}
}
```

在代码清单 1 中，类型 `X` 拥有单个字段，这个字段的名字是 `p`，它是一个指向整型的指针。然后在第 11 行到第 13 行中，构造了一个类型为 `X` 的值，使用紧凑形式，用 `i1` 变量的地址来初始化 `p` 字段。`x1` 变量是作为一个指针创建的，因此，这个变量与在第 17 行创建的变量是一样的。

在第 16 行中，声明了名为 `i2` 的变量，然后在第 17 行中，构造了一个使用指针语义的类型为 `X` 的值，然后将其赋值给指针变量 `x2`。接着在第 18 行中，`i2` 变量的地址被赋给变量 `x2` 执行的值中的 `p` 字段。在这个语句中，存在通过使用指针变量的赋值，这是一种间接赋值。

以下是运行基准测试的结果，以及一份逃逸分析报告。还包括了 pprof list 命令的输出。

**基准测试输出**

```console
$ go test -gcflags "-m -m" -run none -bench . -benchmem -memprofile mem.out

BenchmarkAssignmentIndirect-8       100000000	       14.2 ns/op         8 B/op	      1 allocs/op
```

**逃逸分析报告**

```
./example2_test.go:18:10: &i2 escapes to heap
./example2_test.go:18:10: from x2.p (star-dot-equals) at ./example2_test.go:18:8
./example2_test.go:16:7: moved to heap: i2
./example2_test.go:12:7: BenchmarkAssignmentIndirect &i1 does not escape
```

**Pprof 输出**

```
$ go tool pprof -alloc_space mem.out

ROUTINE ========================
	759.51MB   759.51MB (flat, cum)   100% of Total
		.          .     11:       x1 := &X{
		.          .     12:           p: &i1, // GOOD: i1 does not escape
		.          .     13:       }
		.          .     14:       _ = x1
		.          .     15:
	759.51MB   759.51MB     16:       var i2 int
		.          .     17:       x2 := &X{}
		.          .     18:       x2.p = &i2 // BAD: Cause of i2 escape
		.          .     19:   }
		.          .     20:}	
```

在逃逸分析报告中，`i2` 逃逸给出的理由是，`(star-dot-equals)`。我想这是指编译器需要执行诸如以下的操作来完成此赋值。

**Star-Dot-Equals**

```go
(*x2).p = &i2
```

pprof 输出清晰地显示，`i2` 是在堆上分配的，而 `i1` 不是。我在 Go 语言小萌新写的 Go 代码中，大量看到 16 行到 18 行这样的代码。这个缺陷可以帮助更萌新的开发者从堆中移除一些垃圾。

## 间接调用

“间接调用（Indirect Call）”缺陷与和通过间接调用的函数共享一个值时发生的分配有关。下面是一个代码示例：

**代码清单 2.1**  

<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example2/example2_test.go>

```go
package flaws

import "testing"

func BenchmarkLiteralFunctions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var y1 int
		foo(&y1, 42) // GOOD: y1 does not escape

		var y2 int
		func(p *int, x int) {
			*p = x
		}(&y2, 42) // BAD: Cause of y2 escape

		var y3 int
		p := foo
		p(&y3, 42) // BAD: Cause of y3 escape
	}
}

func foo(p *int, x int) {
	*p = x
}
	
```

在代码清单 2.1 中，在第 21 行声明了一个名为 `foo` 的命名函数。这个函数接受一个整型的地址和一个整型值作为参数。然后，这个函数将传递的整型值赋值给 `p` 指针指向的位置。

在第 07 行，声明了一个类型为 `int`，名字为 `y1` 的变量，这个变量在第 08 行对 `foo` 的函数调用过程中发生了共享。从第 10 行到第 13 行，存在类似的情况。声明了一个类型为 `int` 的变量 `y2`，然后这个变量作为第一个参数共享给一个在第 13 行声明和执行的字面函数。这个字面函数与 `foo` 函数相同。

最后，在第 15 行到第 17 行之间，`foo`函数被赋给一个名为 `p` 的变量。通过变量 `p`，`foo` 函数被执行，其中，变量 `y3` 被共享。第 17 行的这个函数调用是通过 `p` 变量间接完成的。这与第 13 行的字面函数没有显式函数变量所执行的函数调用方式情况相同。

以下是运行基准测试的结果，以及一份逃逸分析报告。还包括了 pprof list 命令的输出。

**基准测试输出**
```
$ go test -gcflags "-m -m" -run none -bench BenchmarkLiteralFunctions -benchmem -memprofile mem.out
	
BenchmarkLiteralFunctions-8     50000000 	       30.7 ns/op        16 B/op	      2 allocs/op
```

**逃逸分析报告**

```
./example2_test.go:13:5: &y2 escapes to heap
./example2_test.go:13:5:    from (func literal)(&y2, 42) (parameter to indirect call) at ./example2_test.go:13:4
./example2_test.go:10:7: moved to heap: y2
./example2_test.go:17:5: &y3 escapes to heap
./example2_test.go:17:5:    from p(&y3, 42) (parameter to indirect call) at ./example2_test.go:17:4
./example2_test.go:15:7: moved to heap: y3
```

**Pprof 输出**

```
$ go tool pprof -alloc_space mem.out

ROUTINE ========================
 768.01MB   768.01MB (flat, cum)   100% of Total
		.          .      5:func BenchmarkLiteralFunctions(b *testing.B) {
		.          .      6:   for i := 0; i < b.N; i++ {
		.          .      7:       var y1 int
		.          .      8:       foo(&y1, 42) // GOOD: y1 does not escape
		.          .      9:
 380.51MB   380.51MB     10:       var y2 int
		.          .     11:       func(p *int, x int) {
		.          .     12:           *p = x
		.          .     13:       }(&y2, 42) // BAD: Cause of y2 escape
		.          .     14:
 387.51MB   387.51MB     15:       var y3 int
		.          .     16:       p := foo
		.          .     17:       p(&y3, 42) // BAD: Cause of y3 escape
		.          .     18:   }
		.          .     19:}

```

在逃逸分析报告中，为变量 `y2` 和 `y3` 变量的分配给出的原因是 `(parameter to indirect call)`。pprof 输出很清楚的显示出，`y2` 和 `y3` 被分配在堆上，而 `y1` 不是。

虽然，我会认为在第 13 行调用的函数字面量的使用是代码异味，但是，第 16 行变量 `p` 的使用并不是。在 Go 中，人们总是会传递函数作为参数。特别是在构建 web 服务的时候。修复这个间接调用缺陷会帮助减少 Go web 服务应用中的许多分配。

这里是一个你会在许多 web 服务应用中找到的例子。

**代码清单 2.2**  

<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example2/example2_http_test.go>

```go
package flaws

import (
    "net/http"
    "testing"
)

func BenchmarkHandler(b *testing.B) {

    // Setup route with specific handler.
    h := func(w http.ResponseWriter, r *http.Request) error {
        // fmt.Println("Specific Request Handler")
        return nil
    }
    route := wrapHandler(h)

    // Execute route.
    for i := 0; i < b.N; i++ {
        var r http.Request
        route(nil, &r) // BAD: Cause of r escape
    }
}

type Handler func(w http.ResponseWriter, r *http.Request) error

func wrapHandler(h Handler) Handler {
    f := func(w http.ResponseWriter, r *http.Request) error {
        // fmt.Println("Boilerplate Code")
        return h(w, r)
    }
    return f
}
```

在代码清单 2.2 中，第 26 行声明了一个通用的处理器封装函数，该函数在另一个字面函数的范围内封装了一个处理器函数，以提供样板代码。然后在第 11 行，声明了一个用于特定路由的处理函数，然后在第 15 行，它被传给 `wrapHandler` 函数，以便可以与样板代码处理函数链接在一起。在第 19 行，创建了一个 `http.Request` 值，然后与第 20 行的 `route` 调用共享。调用 `route` 在功能上同时执行了样板代码和特定的请求处理器。

第 20 行的 `route` 调用属于间接调用，因为 `route` 变量是一个函数变量。这会导致 `http.Request` 变量分配在堆上，这是没有必要的。

以下是运行基准测试的结果，以及一份逃逸分析报告。还包括了 pprof list 命令的输出。

**基准测试输出**

```
$ go test -gcflags "-m -m" -run none -bench BenchmarkHandler -benchmem -memprofile mem.out

BenchmarkHandler-8      20000000 	       72.4 ns/op       256 B/op	      1 allocs/op
```

**逃逸分析报告**

```
./example2_http_test.go:20:14: &r escapes to heap
./example2_http_test.go:20:14:  from route(nil, &r) (parameter to indirect call) at ./example2_http_test.go:20:8
./example2_http_test.go:19:7: moved to heap: r
```

**Pprof 输出**

```
$ go tool pprof -alloc_space mem.out

ROUTINE ========================
   5.07GB     5.07GB (flat, cum)   100% of Total
		.          .     14:   }
		.          .     15:   route := wrapHandler(h)
		.          .     16:
		.          .     17:   // Execute route.
		.          .     18:   for i := 0; i < b.N; i++ {
   5.07GB     5.07GB     19:       var r http.Request
		.          .     20:       route(nil, &r) // BAD: Cause of r escape
		.          .     21:   }
		.          .     22:}
```

在逃逸分析报告中，你可以看到这种分配的原因是 `(parameter to indirect call)`。pprof 报告显示，`r` 变量正在分配。如前所述，这是人们在用 Go 构建 web 服务时编写的常见代码。修复这个缺陷会减少程序中大量的分配。

## 切片和 Map 赋值

“切片和 Map 赋值（Slice and Map Assignments）”缺陷与值在切片或者 Map 中共享时发生的分配有关。这里是一个代码示例：

**代码清单 3**  

<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example3/example3_test.go>

```go
package flaws

import "testing"

func BenchmarkSliceMapAssignment(b *testing.B) {
    for i := 0; i < b.N; i++ {
        m := make(map[int]*int)
        var x1 int
        m[0] = &x1 // BAD: cause of x1 escape

        s := make([]*int, 1)
        var x2 int
        s[0] = &x2 // BAD: cause of x2 escape
   }
}
```

在代码清单 3 中，第 07 行创建了一个 map，它保存类型 `int` 的值的地址。然后在第 08 行，创建了一个类型 `int` 的值，接着在第 09 行，在 map 中共享了这个值，map 的键为 0。在第 11 行保存 `int` 地址的切片上也发生了同样的事情。在创建切片后，索引 0 内共享了类型为 `int` 的值。

以下是运行基准测试的结果，以及一份逃逸分析报告。还包括了 pprof list 命令的输出。

**基准测试输出**

```
$ go test -gcflags "-m -m" -run none -bench . -benchmem -memprofile mem.out

BenchmarkSliceMapAssignment-8       10000000 	      104 ns/op 	     16 B/op	      2 allocs/op
```

**逃逸分析报告**

```
./example3_test.go:9:10: &x1 escapes to heap
./example3_test.go:9:10:    from m[0] (value of map put) at ./example3_test.go:9:8
./example3_test.go:8:7: moved to heap: x1
./example3_test.go:13:10: &x2 escapes to heap
./example3_test.go:13:10:   from s[0] (slice-element-equals) at ./example3_test.go:13:8
./example3_test.go:12:7: moved to heap: x2
./example3_test.go:7:12: BenchmarkSliceMapAssignment make(map[int]*int) does not escape
./example3_test.go:11:12: BenchmarkSliceMapAssignment make([]*int, 1) does not escape
```

**Pprof 输出**

```
$ go tool pprof -alloc_space mem.out

ROUTINE ========================
 162.50MB   162.50MB (flat, cum)   100% of Total
		.          .      5:func BenchmarkSliceMapAssignment(b *testing.B) {
		.          .      6:   for i := 0; i < b.N; i++ {
		.          .      7:       m := make(map[int]*int)
 107.50MB   107.50MB      8:       var x1 int
		.          .      9:       m[0] = &x1 // BAD: cause of x1 escape
		.          .     10:
		.          .     11:       s := make([]*int, 1)
	 55MB       55MB     12:       var x2 int
		.          .     13:       s[0] = &x2 // BAD: cause of x2 escape
		.          .     14:   }
		.          .     15:}
```

逃逸分析报告中给出的原因是 `(value of map put)` 和 `(slice-element-equals)`。更有趣的是，逃逸分析报告显示，map 和切片数据结构不分配（不逃逸）。

**不分配 Map 和切片**

```
./example3_test.go:7:12: BenchmarkSliceMapAssignment make(map[int]*int) does not escape
./example3_test.go:11:12: BenchmarkSliceMapAssignment make([]*int, 1) does not escape
```

这进一步证明，代码示例中的 `x1` 和 `x2` 无需在堆上分配。

我一直认为，在合理和实际的情况下，map 和切片中的数据应该作为值存储。特别是当这些数据结构正存储着一个请求或任务的核心数据的时候。这个缺陷为尝试避免通过使用指针来存储数据提供了另一个理由。修复这个缺陷可能几乎没有什么回报，因为静态大小的 map 和切片很少见。

## 接口

“接口（Interfaces）”缺陷与之前看到的“间接调用”缺陷有关。这是一个使用接口产生实际成本的缺陷。下面是一个代码示例：

**代码清单 4**  
<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example4/example4_test.go>

```go
package flaws

import "testing"

type Iface interface {
    Method()
}

type X struct {
    name string
}

func (x X) Method() {}

func BenchmarkInterfaces(b *testing.B) {
    for i := 0; i < b.N; i++ {
        x1 := X{"bill"}
        var i1 Iface = x1
        var i2 Iface = &x1

        i1.Method() // BAD: cause copy of x1 to escape
        i2.Method() // BAD: cause x1 to escape

        x2 := X{"bill"}
        foo(x2)
        foo(&x2)
    }
}

func foo(i Iface) {
    i.Method() // BAD: cause value passed in to escape
}
```

在代码清单 4 中，在第 05 行声明了一个名为 `Iface` 的接口，并且为了示例目的，这个接口保持得非常简单。然后，在第 09 行声明了一个名为 `X` 的具体类型，并且，使用值接收器来实现 `Iface` 接口。

在第 17 行中，构建了一个类型为 `X` 的值，然后将其赋给 `x1` 变量。在第 18 行，`x1` 变量的一个拷贝存储在 `i1` 接口变量中，接着，在第 19 行，相同的 `x1` 变量与 `i2` 接口变量共享。在第 21 和 22 行，同时对 `i1` 和 `i2` 接口变量调用 `Method`。

为了创建一个更实际的例子，在第 30 行声明了一个名为 `foo` 的函数，它接受任何实现 `Iface` 接口的具体数据。然后，在第 31 行，对本地接口变量同样调用 `Method`。`foo` 函数代表了大家在 Go 中写的大量函数。

在第 24 行，构造了一个类型为 `X` 名为 `x2` 的变量，然后将其作为拷贝传递给 `foo`，并分别在第 25 和 26 行中共享。

以下是运行基准测试的结果，以及一份逃逸分析报告。还包括了 pprof list 命令的输出。

**基准测试输出**

```
$ go test -gcflags "-m -m" -run none -bench . -benchmem -memprofile mem.out
	
BenchmarkInterfaces-8     10000000         126 ns/op        64 B/op        4 allocs/op
```

**逃逸分析报告**

```
./example4_test.go:18:7: x1 escapes to heap
./example4_test.go:18:7:  from i1 (assigned) at ./example4_test.go:18:7
./example4_test.go:18:7:  from i1.Method() (receiver in indirect call) at ./example4_test.go:21:12
./example4_test.go:19:7: &x1 escapes to heap
./example4_test.go:19:7:  from i2 (assigned) at ./example4_test.go:19:7
./example4_test.go:19:7:  from i2.Method() (receiver in indirect call) at ./example4_test.go:22:12
./example4_test.go:19:18: &x1 escapes to heap
./example4_test.go:19:18:   from &x1 (interface-converted) at ./example4_test.go:19:7
./example4_test.go:19:18:   from i2 (assigned) at ./example4_test.go:19:7
./example4_test.go:19:18:   from i2.Method() (receiver in indirect call) at ./example4_test.go:22:12
./example4_test.go:17:17: moved to heap: x1
./example4_test.go:25:6: x2 escapes to heap
./example4_test.go:25:6:  from x2 (passed to call[argument escapes]) at ./example4_test.go:25:6
./example4_test.go:26:7: &x2 escapes to heap
./example4_test.go:26:7:  from &x2 (passed to call[argument escapes]) at ./example4_test.go:26:6
./example4_test.go:26:7: &x2 escapes to heap
./example4_test.go:26:7:  from &x2 (interface-converted) at ./example4_test.go:26:7
./example4_test.go:26:7:  from &x2 (passed to call[argument escapes]) at ./example4_test.go:26:6
./example4_test.go:24:17: moved to heap: x2
```

**Pprof 输出**

```
$ go tool pprof -alloc_space mem.out

ROUTINE ======================== 
 658.01MB   658.01MB (flat, cum)   100% of Total
		.          .     12:
		.          .     13:func (x X) Method() {}
		.          .     14:
		.          .     15:func BenchmarkInterfaces(b *testing.B) {
		.          .     16: for i := 0; i < b.N; i++ {
 167.50MB   167.50MB     17:   x1 := X{"bill"}
 163.50MB   163.50MB     18:   var i1 Iface = x1
		.          .     19:   var i2 Iface = &x1
		.          .     20:
		.          .     21:   i1.Method() // BAD: cause copy of x1 to escape
		.          .     22:   i2.Method() // BAD: cause x1 to escape
		.          .     23:
 163.50MB   163.50MB     24:   x2 := X{"bill"}
 163.50MB   163.50MB     25:   foo(x2)
		.          .     26:   foo(&x2)
		.          .     27: }
		.          .     28:}

```

注意，在基准报告中有四个分配。这是因为代码会复制 `x1` 和 `x2` 变量，这也会产生分配。在第 18 行中使用`x1` 变量进行赋值时，以及在第 25 行中对 `foo` 进行函数调用使用 `x2` 的值时，创建了这些副本。

在逃逸分析报告中，为 `x1` 以及 `x1` 的副本逃逸提供的原因是 `(receiver in indirect call)`。这很有趣，因为第 21 和 22 行对 `Method` 的调用才是这个缺陷真正的罪魁祸首。请记住，针对接口的方法调用需要通过 iTable 进行间接调用。正如你之前看到的，间接调用是逃逸分析中的一个缺陷。

逃逸分析报告为 `x2` 变量逃逸给出的原因是 `(passed to call[argument escapes])`。但是在这两种情况下，`(interface-converted)` 是另一个原因，它描述了数据存储在接口里的事实。

有趣的是，如果你移除第 31 行中 `foo` 函数里的方法调用，那么，分配就会消失。实际上，第 21，22 和 `foo` 中的 31 行中，通过接口变量对 `Method` 的间接调用才是问题所在。

我总是在说，从 1.9 甚至更早的版本开始，使用接口会产生间接和分配的开销。这是逃逸分析的缺陷，如果修正这一缺陷，会给 Go 程序带来最大的影响。这可以减少单独日志包的大量分配。不要使用接口，除非它们（指接口）提供的价值是显著的。

## 未知

这种类型的分配是某些我完全不明白的东东。即使在看了工具的输出，还是没搞明白。这里，我把它们供出，期望能得到一些答案。

下面是代码示例。

**代码清单 5**  
<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example5/example5_test.go>

```go
package flaws

import (
    "bytes"
    "testing"
)

func BenchmarkUnknown(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var buf bytes.Buffer
        buf.Write([]byte{1})
        _ = buf.Bytes()
    }
}
```

在代码清单 5 中，第 10 行创建了一个类型为 `bytes.Buffer` 的值，并将其设置为零值。然后，在第 11 行构造了一个切片值，并将其传递给 `buf` 变量上的 `Write` 方法调用。最后，为了防止潜在的编译器优化抛出所有的代码，调用 Bytes 方法。该调用不是创造 `buf` 变量逃逸的必要条件。

以下是运行基准测试的结果，以及一份逃逸分析报告。还包括了 pprof list 命令的输出。

**基准测试输出**

```
$ go test -gcflags "-m -m" -run none -bench . -benchmem -memprofile mem.out

Benchmark-8     20000000 	       50.8 ns/op       112 B/op	      1 allocs/op
```

**逃逸分析报告**

```
./example5_test.go:11:6: buf escapes to heap
./example5_test.go:11:6:    from buf (passed to call[argument escapes]) at ./example5_test.go:11:12
```

**Pprof 输出**

```
$ go tool pprof -alloc_space mem.out

ROUTINE ======================== 
   2.19GB     2.19GB (flat, cum)   100% of Total
		.          .      8:func BenchmarkUnknown(b *testing.B) {
		.          .      9:   for i := 0; i < b.N; i++ {
   2.19GB     2.19GB     10:       var buf bytes.Buffer
		.          .     11:       buf.Write([]byte{1})
		.          .     12:       _ = buf.Bytes()
		.          .     13:   }
		.          .     14:}

```

在这个代码中，我没有看到第 11 行对 `Write` 的方法调用引起逃逸的任何原因。我得到了一个看起来很有意思的指引，但我会留给你去进一步探索。

_这可能与 `Buffer` 类型的引导数组有关。它意味着一种优化，但是从逃逸分析的角度来说，它让 `Buffer` 指向自身，这是一种循环依赖，通常难以分析。或者也许是因为 `append`，又或者也许只是几个因素和 `Buffer` 中非常复杂的代码的组合。_

**有这个问题，它与导致这种分配的引导数组有关：**

[cmd/compile, bytes: bootstrap array causes bytes.Buffer to always be heap-allocated](https://github.com/golang/go/issues/7921)

**CKS 在 reddit 上发布了此回复：**

`Unknown` 情况下的逃逸是因为 Go 认为给 bytes.Buffer.Write() 的参数逃逸。如果你在 buffer 包的源代码上运行逃逸分析，那么它会输出（对于 Write()）：

```
./buffer.go:170:46: leaking param content: p
./buffer.go:170:46:     from *p (indirection) at ./buffer.go:170:46
./buffer.go:170:46:     from copy(b.buf[m:], p) (copied slice) at ./buffer.go:176:13
(The line numbers are for the current git tip; they may be slightly off in other copies.)
```

考虑到 copy() 是语言内置函数，似乎编译器应该知道这里，源参数不逃逸。或者有可能编译器在对 copy() 的实际实现做一些十分有趣的事情，以至于源在某些情况下会逃逸。

## 总结

我试图指出 1.9 版本至今存在的一些更有趣的逃逸分析缺陷。接口缺陷可能是一旦修复就会对当今的 Go 程序产生最大影响的缺陷。我觉得最有意思的是，我们所有人都能从这些缺陷的修复中获益，而不需要有这方面的个人专长。编译器执行的静态代码分析提供了诸多好处，例如优化你随时写入的代码的能力。也许最大的好处是，消除或减少你不得不维持的认知负担。

----------------

via: https://www.ardanlabs.com/blog/2018/01/escape-analysis-flaws.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[ictar](https://github.com/ictar)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
