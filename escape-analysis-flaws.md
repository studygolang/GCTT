# Escape-Analysis Flaws

### Prelude

It will be helpful to read this four-part series first on escape analysis and
data semantics. Details on how to read an escape analysis report and pprof
output have been outlined here.

<https://www.ardanlabs.com/blog/2017/05/language-mechanics-on-stacks-and-
pointers.html>

### Introduction

Even after working with Go for 4 years, I am continually amazed by the
language. Thanks to the static code analysis the compiler performs, the
compiler can apply interesting optimizations to the code it produces. One type
of analysis the compiler performs is called escape analysis. This produces
optimizations and simplifications around memory management.

The language team has been focused for the past 2 years on optimizing the code
the compiler produces for better performance and they have done a fantastic
job. I believe Go programs could see even more dramatic improvements if some
of the current flaws in escape analysis are resolved. Back in February 2015,
Dmitry Vyukov wrote this paper outlining known escape analysis flaws in the
compiler.

<https://docs.google.com/document/d/1CxgUBPlx9iJzkz9JWkb6tIpTe5q32QDmz8l0BouG0Cw/edit#>

I was curious about how many of these flaws had been fixed since this document
was written and I found that so far a few have been resolved. That being said,
five particular flaws have not been fixed that I would love to see worked on
in a near future release of Go. I label these as:

  * 间接赋值
  * 间接调用
  * Slice 和 Map 赋值
  * 接口
  * Unknown

I thought it would be fun to explore each of these flaws so you can see the
positive impact existing Go programs will have once they are fixed. Everything
you see is based on the 1.9 compiler.

### 间接赋值

The “Indirection Assignment” flaw has to do with allocations that occur when a
value is assigned through an indirection. Here is a code example:

**代码清单 1**  
<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example1/example1_test.go>

```go

    01 package flaws
    02
    03 import "testing"
    04 
    05 func BenchmarkAssignmentIndirect(b *testing.B) {
    06     type X struct {
    07         p *int
    08     }
    09     for i := 0; i < b.N; i++ {
    10         var i1 int
    11         x1 := &X{
    12             p: &i1, // GOOD: i1 does not escape
    13         }
    14         _ = x1
    15
    16         var i2 int
    17         x2 := &X{}
    18         x2.p = &i2 // BAD: Cause of i2 escape
    19     }
    20 }
    
```

In listing 1, a type named `X` is declared with a single field named `p` as a
pointer to an integer. Then on lines 11 through 13, a value of type `X` is
constructed using the compact form to initialize the `p` field with the
address of the `i1` variable. The `x1` variable is created as a pointer so
this variable is the same as the variable created on line 17.

On line 16, a variable named `i2` is declared and on line 17, a value of type
`X` using pointer semantics is constructed and assigned to the pointer
variable `x2`. Then on line 18, the address of the `i2` variable is assigned
to the `p` field within the value that the `x2` variable points to. In this
statement, there is an assignment through the use of a pointer variable, which
is an indirection.

Here is the output from running the benchmark with an escape analysis report.
Also included is the output for the pprof list command.

**Benchmark Output**
```

    $ go test -gcflags "-m -m" -run none -bench . -benchmem -memprofile mem.out
    
    BenchmarkAssignmentIndirect-8       100000000	       14.2 ns/op         8 B/op	      1 allocs/op
    
```

**逃逸分析报告**
```

    ./example2_test.go:18:10: &i2 escapes to heap
    ./example2_test.go:18:10:   from x2.p (star-dot-equals) at ./example2_test.go:18:8
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

在逃逸分析报告中， the reason given for `i2` to escape is `(star-
dot-equals)`. I imagine this is referencing the need for the compiler to
perform an operation like this underneath to make the assignment.

**Star-Dot-Equals**
```go

    (*x2).p = &i2
    
```

pprof 输出清晰地显示，`i2` 是在堆上分配的，而 `i1` 不是。喔在 Go 语言小萌新写的 Go 代码中，大量看到 16 行到 18 行这样的代码。这个缺陷可以帮助更萌新的开发者从堆中移除一些垃圾。

### 间接调用

The “Indirect Call” flaw has to do with allocations that occur when a value is
shared with a function that is called through an indirection. Here is a code
example:

**代码清单 2.1**  
<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example2/example2_test.go>

```go

    01 package flaws
    02
    03 import "testing"
    04
    05 func BenchmarkLiteralFunctions(b *testing.B) {
    06     for i := 0; i < b.N; i++ {
    07         var y1 int
    08         foo(&y1, 42) // GOOD: y1 does not escape
    09
    10         var y2 int
    11         func(p *int, x int) {
    12             *p = x
    13         }(&y2, 42) // BAD: Cause of y2 escape
    14
    15         var y3 int
    16         p := foo
    17         p(&y3, 42) // BAD: Cause of y3 escape
    18     }
    19 }
    20
    21 func foo(p *int, x int) {
    22     *p = x
    23 }
    
```

In listing 2.1, a named function called `foo` on line 21 is declared. This
function accepts the address of an integer along with an integer value. Then
the function assigns the integer value that is passed to the location that the
`p` pointer points to.

On line 07, a variable named `y1` of type `int` is declared and shared during
the function call to `foo` on line 08. Between lines 10 through 13, a similar
situation exists. A variable named `y2` is declared of type `int` and shared
as the first parameter to a literal function that is declared and executed in
place on line 13. The literal function is identical to the `foo` function.

Finally between lines 15 through 17, the `foo` function is assigned to a
variable named `p`. Through the `p` variable, the `foo` function is executed
with the `y3` variable is shared. This function call on line 17 is done
through the indirection of the `p` variable. This is identical to how the
function call of the literal function on line 13 is performed without the
explicit function variable.

Here is the output from running the benchmark with an escape analysis report.
Also included is the output for the pprof list command.

**Benchmark Output**
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

在逃逸分析报告中， the reason given for the allocation of the `y2`
and `y3` variables is `(parameter to indirect call)`. The pprof output shows
clearly that `y2` and `y3` are allocated on the heap and `y1` is not.

Though I would consider the use of a function literal as called on line 13 to
be a code smell, the use of the `p` variable on line 16 is not. People pass
functions around in Go all the time. Especially when building web services.
Fixing this indirect call flaw could help reduce many allocations in Go web
service applications.

这里是一个你会在许多 web 服务应用中找到的例子。

**代码清单 2.2**  
<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example2/example2_http_test.go>

```go

    01 package flaws
    02
    03 import (
    04     "net/http"
    05     "testing"
    06 )
    07
    08 func BenchmarkHandler(b *testing.B) {
    09
    10     // Setup route with specific handler.
    11     h := func(w http.ResponseWriter, r *http.Request) error {
    12         // fmt.Println("Specific Request Handler")
    13         return nil
    14     }
    15     route := wrapHandler(h)
    16
    17     // Execute route.
    18     for i := 0; i < b.N; i++ {
    19         var r http.Request
    20         route(nil, &r) // BAD: Cause of r escape
    21     }
    22 }
    23
    24 type Handler func(w http.ResponseWriter, r *http.Request) error
    25
    26 func wrapHandler(h Handler) Handler {
    27     f := func(w http.ResponseWriter, r *http.Request) error {
    28         // fmt.Println("Boilerplate Code")
    29         return h(w, r)
    30     }
    31     return f
    32 }
    
```

In listing 2.2, a common handler wrapping function is declared on line 26,
which wraps a handler function within the scope of another literal function to
provide boilerplate code. Then on line 11, a handler function for a specific
route is declared and it’s passed to the `wrapHandler` function on line 15 so
it can be chained together with the boilerplate code handler function. On line
19, a `http.Request` value is created and shared with the `route` call on line
20. Calling `route` executes both the boilerplate code and specific request
handler functionality.

The `route` call on line 20 is an indirect call since the `route` variable is
a function variable. This will cause the `http.Request` variable to allocate
on the heap, which is not necessary.

Here is the output from running the test with an escape analysis report. Also
included is the output is the pprof list command.

**Benchmark Output**
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

在逃逸分析报告中，你可以看到这种分配的原因是 `(parameter to indirect call)`。pprof 报告显示，`r` variable
is allocating. As stated earlier, this is common code people write in Go when
building web services. Fixing this could reduce a large number of allocations
in programs.

### Slice 和 Map 赋值

The “Slice and Map Assignments” flaw has to do with allocations that occur
when a value is shared inside a slice or map. Here is a code example:

**代码清单 3**  
<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example3/example3_test.go>

```go

    01 package flaws
    02
    03 import "testing"
    04
    05 func BenchmarkSliceMapAssignment(b *testing.B) {
    06     for i := 0; i < b.N; i++ {
    07         m := make(map[int]*int)
    08         var x1 int
    09         m[0] = &x1 // BAD: cause of x1 escape
    10
    11         s := make([]*int, 1)
    12         var x2 int
    13         s[0] = &x2 // BAD: cause of x2 escape
    14    }
    15 }
    
```

In listing 3, a map is made on line 07 which stores addresses of values of
type `int`. Then on line 08, a value of type `int` is created and shared
inside the map on line 09, with the key of 0. The same thing happens with the
slice of `int` addresses on line 11. After the slice is made, a value of type
`int` is shared inside index 0.

Here is the output from running the benchmark with an escape analysis report.
Also included is the output for the pprof list command.

**Benchmark Output**
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

In the escape analysis report the reason given is `(value of map put)` and
`(slice-element-equals)`. What is even more interesting is the escape analysis
report says the map and slice data structures do not allocate.

**No Allocation of Map and Slice**
```

    ./example3_test.go:7:12: BenchmarkSliceMapAssignment make(map[int]*int) does not escape
    ./example3_test.go:11:12: BenchmarkSliceMapAssignment make([]*int, 1) does not escape
    
```

That further proves `x1` and `x2` in this code example have no need to
allocate on the heap.

I have always felt that data in maps and slices should be stored as values
when it is reasonable and practical to do so. Especially when these data
structures are storing the core data for a request or task. This flaw provides
a second reason for trying to avoid storing data through the use of pointers.
Fixing this flaw probably has little return on investment since maps and
slices of static size are rare.

### Interfaces

The “Interfaces” flaw is related to the “Indirect Call” flaw you saw earlier.
This is a flaw that creates a real cost to using interfaces. Here is a code
example:

**代码清单 4**  
<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example4/example4_test.go>

```go

    01 package flaws
    02
    03 import "testing"
    04
    05 type Iface interface {
    06     Method()
    07 }
    08
    09 type X struct {
    10     name string
    11 }
    12
    13 func (x X) Method() {}
    14
    15 func BenchmarkInterfaces(b *testing.B) {
    16     for i := 0; i < b.N; i++ {
    17         x1 := X{"bill"}
    18         var i1 Iface = x1
    19         var i2 Iface = &x1
    20
    21         i1.Method() // BAD: cause copy of x1 to escape
    22         i2.Method() // BAD: cause x1 to escape
    23
    24         x2 := X{"bill"}
    25         foo(x2)
    26         foo(&x2)
    27     }
    28 }
    29
    30 func foo(i Iface) {
    31     i.Method() // BAD: cause value passed in to escape
    32 }
    
```

In listing 4, an interface named `Iface` is declared on line 05 and is kept
very basic for the example. Then a concrete type named `X` is declared on line
09 and the `Iface` interface is implemented using a value receiver.

On line 17, a value of type `X` is constructed and assigned to the `x1`
variable. A copy of the `x1` variable is stored inside the `i1` interface
variable on line 18 and then that same `x1` variable is shared with the `i2`
interface variable on line 19. On lines 21 and 22, `Method` is called against
both the `i1` and `i2` interface variables.

To create a more realistic example, a function named `foo` is declared on line
30 and it accepts any concrete data that implements the `Iface` interface.
Then on line 31, the same call to `Method` is made against the local interface
variable. The `foo` function represents a large number of functions people
write in Go.

On line 24, a variable named `x2` of type `X` is constructed and passed to
`foo` as a copy and shared on lines 25 and 26 respectively.

Here is the output from running the benchmark with an escape analysis report.
Also included is the output for the pprof list command.

**Benchmark Output**
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

In the benchmark report, notice there are four allocations. This is because
the code makes copies of the `x1` and `x2` variables, which allocate as well.
These copies are made on line 18 for the `x1` variable during the assignment
and on line 25 when the value of `x2` is used in the function call to `foo`.

在逃逸分析报告中， the reason given for `x1` and the copy of `x1`
to escape is `(receiver in indirect call)`. This is interesting because it is
the call to `Method` on lines 21 and 22 that is the real culprit here in this
flaw. Remember, calling a method against an interface requires an indirect
call through the iTable. As you saw earlier, indirect calls are a flaw in
escape analysis.

The reason the escape analysis report gives for the `x2` variable to escape is
`(passed to call[argument escapes])`. However in both cases, `(interface-
converted)` is another reason which describes the fact that the data is being
stored inside the interface.

What’s interesting is, if you remove the method call on line 31 inside the
`foo` function, the allocation goes away. In reality, the indirect call of
`Method` through the interface variable on lines 21, 22 and 31 inside of `foo`
is the problem.

I always teach that as of 1.9 and earlier, the use of interfaces has the cost
of indirection and allocation. This is the escape analysis flaw that if fixed,
can have the most significant impact on Go programs. This could reduce a large
number of allocations on logging packages alone. Don’t use interfaces unless
it is obvious the value they are providing.

### Unknown

这种类型的分配是某些我完全不明白的东东。Even after
looking at the output of the tooling. I am providing it here with the hope to
get some answers.

Here is a code example:

**代码清单 5**  
<https://github.com/ardanlabs/gotraining/blob/master/topics/go/language/pointers/flaws/example5/example5_test.go>

```go

    01 package flaws
    02
    03 import (
    04     "bytes"
    05     "testing"
    06 )
    07
    08 func BenchmarkUnknown(b *testing.B) {
    09     for i := 0; i < b.N; i++ {
    10         var buf bytes.Buffer
    11         buf.Write([]byte{1})
    12         _ = buf.Bytes()
    13     }
    14 }
    
```

In listing 5, a value of type `bytes.Buffer` is created on line 10 and set to
its zero value. Then the method `Write` is called against the `buf` variable
on line 11 with a slice value constructed and passed within the call. Finally,
the Bytes method is called just to prevent potential compiler optimizations
from throwing all the code away. That call is not necessary to create the
escape of the `buf` variable.

Here is the output from running the benchmark with an escape analysis report.
Also included is the output for the pprof list command.

**Benchmark Output**
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

In this code, I don’t see any reason why the method call to `Write` on line 11
is causing an escape. I was given a lead that looked interesting but I will
leave it up to you to explore further.

_Potentially it has something to do with the bootstrap array in the `Buffer`
type. It's meant to be an optimization, but from escape analysis point of view
it makes `Buffer` to point to itself, which is a circular dependency and these
are usually hard for analysis. Or perhaps it's because of `append` or maybe
it's just a combination of several factors and quite complex code in
`Buffer`._

**This issue exists which is related to the bootstrap array causing the allocation:**

[cmd/compile, bytes: bootstrap array causes bytes.Buffer to always be heap-allocated](https://github.com/golang/go/issues/7921)

**CKS on reddit posted this response:**

The `Unknown` case escapes because Go thinks the argument to
bytes.Buffer.Write() escapes. If you run escape analysis on the source to the
buffer package, it reports (for Write()):

```

    ./buffer.go:170:46: leaking param content: p
    ./buffer.go:170:46:     from *p (indirection) at ./buffer.go:170:46
    ./buffer.go:170:46:     from copy(b.buf[m:], p) (copied slice) at ./buffer.go:176:13
    (The line numbers are for the current git tip; they may be slightly off in other copies.)
    
```

Given that copy() is a language builtin, it seems like the compiler should
know that the source argument doesn't escape here. Or possibly the compiler is
doing something sufficiently interesting with the actual implementation of
copy() such that the source might escape under some circumstances.

### 总结

I have tried to point out some of the more interesting escape analysis flaws
that exist today as of 1.9. The interface flaw is probably the flaw that if
corrected, can have the largest impact on Go programs today. What I find most
interesting is that all of us can gain from fixing these flaws without any
need for personal expertise in this area. The static code analysis the
compiler performs is providing many benefits, such as the ability to optimize
the code you write over time. Maybe the biggest benefit is, removing or
reducing the cognitive load you otherwise would have to maintain.

----------------

via: https://www.ardanlabs.com/blog/2018/01/escape-analysis-flaws.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[ictar](https://github.com/ictar)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出