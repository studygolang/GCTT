# The Zoo of Go Functions

### An overview about: anonymous, higher-order, closures, concurrent, deferred, and other kinds of Golang funcs.



![The Zoo of Go Funcs](http://www.z4a.net/images/2017/11/26/The_zoo_of_go_funcs.png)



> *This post is a summary for the different kind of funcs in Go. I’ll go into more detail in the upcoming posts because they deserve more. This is just a start.*



---

### Named Funcs

A named func has a name and declared at the package-level—*outside of the body of another func*.



***👉 I already have explained them fully in my another post, **[**here**](https://blog.learngoprogramming.com/golang-funcs-params-named-result-values-types-pass-by-value-67f4374d9c0a)**.***

![named Func](http://www.z4a.net/images/2017/11/26/named_funcs.png)

<p align="center">This is a named func: Len func takes a string and returns and int</p>

---

### Variadic funcs

Variadic funcs can accept an optional number of input parameters



***👉 To learn more about them check out my other post, **[**here**](https://blog.learngoprogramming.com/golang-variadic-funcs-how-to-patterns-369408f19085)**.***

![Variadic Funcs](http://www.z4a.net/images/2017/11/27/variadic_funcs.png)

---

### Methods

When you attach a func to a type the func becomes a [method](https://golang.org/ref/spec#Method_declarations) of that type.So, it can be called through that type. Go will pass the type *(the receiver)* to the method when it’s called.

#### Example

Create a new counter type and attach a method to it:

```go
type Count int

func (c Count) Incr() int {
  c = c + 1
  return int(c)
}
```

The method above is similar to this func:

```go
func Incr(c Count) int
```



![Method](http://www.z4a.net/images/2017/11/27/methods.png)

<p align="center">Not exactly true but you can think of the methods as above</p>



#### **Value receiver**

The value of the Count instance is copied and passed to the method when it’s called.

```go
var c Count; c.Incr(); c.Incr()

// output: 1 1
```

<h3 align="center"><i>It doesn’t increase because “c” is a value-receiver.</i></h3>

![Value receiver](http://www.z4a.net/images/2017/11/27/value_receiver.png)

#### Pointer receiver

To increase the counter’s value you need to attach *the Incr func* to *the* *Count pointer type* — `*Count`.

```go
func (c *Count) Incr() int {
  *c = *c + 1
  return int(*c)
}

var c Count
c.Incr(); c.Incur()
// output: 1 2
```

![pointer receiver](http://www.z4a.net/images/2017/11/27/pointer_receiver.png)

[![run the code]](http://www.z4a.net/images/2017/11/27/run_the_code.png)

[run the code]:https://play.golang.org/p/hGVJWPIFZG	"receiver"

##### *There are more examples from my previous posts: [here](https://blog.learngoprogramming.com/golang-const-type-enums-iota-bc4befd096d3#c320) and [here](https://blog.learngoprogramming.com/golang-funcs-params-named-result-values-types-pass-by-value-67f4374d9c0a#638f).*

---

### Interface methods

Let’s recreate the above program using the interface methods. Let’s create a new interface named as *Counter*:

```go
type Counter interface {
  Incr() int
}
```

onApiHit func below can use any type which has an `Incr() int` method:

```go
func onApiHit(c Counter) {
  c.Incr()
}
```

Just use our dummy counter for now — *you can use a real api counter as well*:

```go
dummyCounter := Count(0)
onApiHit(&dummyCounter)
// dummyCounter = 1
```

![interface methods](http://www.z4a.net/images/2017/11/27/interface_funcs.png)

Because, the Count type has the `Incr() int` method on its method list, `onApiHit func` can use it to increase the counter — I passed a pointer of dummyCounter to onApiHit, otherwise it wouldn’t increase the counter.

[![run the code]](http://www.z4a.net/images/2017/11/27/run_the_code.png)

[run the code]: https://play.golang.org/p/w0oyZjmdMA	"interface method"

*The difference between the interface methods and the ordinary methods is that the interfaces are much more flexible and loosely-coupled. You can switch to different implementations across packages without changing any code inside onApiHit etc.*

---

### First-class funcs

First-class means that funcs are value objects just like any other values which can be stored and passed around.

![first-class funcs](http://www.z4a.net/images/2017/11/27/first-class_funcs.png)

<p align="center">Funcs can be used with other types as a value and vice-versa</p>

#### Example:

The sample program here processes a sequence of numbers by using a slice of *Crunchers* as an input param value to a func named “*crunch*”.

Declare a new “**user-defined func type**” which takes an int and returns an int.

*This means that any code that uses this type accepts a func with this exact* [**signature**](https://blog.learngoprogramming.com/golang-funcs-params-named-result-values-types-pass-by-value-67f4374d9c0a#747e):

```go
type Cruncher func(int) int
```

Declare a few cruncher funcs:

```go
func mul(n int) int {
  return n * 2
}

func add(n int) int {
  return n + 100
}

func sub(n int) int {
  return n - 1
}
```

Crunch func processes a series of ints using *a [variadic](https://blog.learngoprogramming.com/golang-variadic-funcs-how-to-patterns-369408f19085) Cruncher funcs*:

```go
func crunch(nums []int, a ...Cruncher) (rnums []int) {
  // create an identical slice
  rnums = append(rnums, nums...)
  
  for _, f := range a {
    for i, n := range rnums {
      rnums[i] = f(n)
    }
  }
  
  return
}
```

Declare an int slice with some numbers and process them:

```go
nums := []int{1, 2, 3, 4, 5}

crunch(nums, mul, add, sub)
```

#### Output:

```
[101 103 105 107 109]
```

[![run the code]](http://www.z4a.net/images/2017/11/27/run_the_code.png)

[run the code]: https://play.golang.org/p/hNSKZAo0p6	"first-class func"

---

### Anonymous funcs

A noname func is an anonymous func and it’s declared inline using *a [function literal](https://golang.org/ref/spec#Function_literals).* It becomes more useful when it’s used as a closure, higher-order func, deferred func, etc.

![annoymous funcs](http://www.z4a.net/images/2017/11/27/Anonymous_funcs.png)

#### Signature

A named func:

```go
func Bang(energy int) time.Duration
```

An anonymous func:

```go
func(energy int) time.Duration
```

They both have the same signature, so they can be used interchangeably:

```go
func(int) time.Duration
```

[![run the code]](http://www.z4a.net/images/2017/11/27/run_the_code.png)

[run the code]: https://play.golang.org/p/-az-2qBr9T	"annoymous func"

#### Example

Let’s recreate the cruncher program from the *First-Class Funcs section above* using the anonymous funcs. Declare the crunchers as anonymous funcs inside the main func.

```go
func main() {
  crunch(nums,
         func(n int) int {
           return n * 2
         },
         func(n int) int {
           return n + 100
         },
         func(n int) int {
           return n - 1
         })
}
```

*This works because, crunch func only expects the Cruncher func type, it doesn’t care that they’re named or anonymous funcs.*

To increase the readability you can also assign them to variables before passing to crunch func:

```go
mul := func(n int) int {
  return n * 2
}

add := func(n int) int {
  return n + 100
}

sub := func(n int) int {
  return n - 1
}

crunch(nums, mul, add, sub)
```

[![run the code]](http://www.z4a.net/images/2017/11/27/run_the_code.png)

[run the code]: https://play.golang.org/p/iqcumj5cka	"use annoymous func"

---

### Higher-Order funcs

An higher order func may take one or more funcs or it may return one or more funcs. Basically, it uses other funcs to do its work.

![hight-order funcs](http://www.z4a.net/images/2017/11/27/higher-order_funcs.png)

Split func in the closures section below is a higher-order func. It returns a *tokenizer func type* as a result.

---



### Closures

A closure can remember all the surrounding values where it’s defined. One of the benefits of a closure is that it can operate on the captured environment as long as you want *—beware the leaks!*

#### Example

Declare a new func type that returns the next word in a splitted string:

```go
type tokenizer func() (token string, ok bool)
```

Split func below is **an higher-order func** which splits a string by a separator and returns a **closure** which enables to walk over the words of the splitted string. *The returned closure can use the surrounding variables: “tokens” and “last”.*

![cloure](http://www.z4a.net/images/2017/11/27/closure.png)

#### **Let’s try:**

```go
const sentence = "The quick brown fox jumps over the lazy dog"

iter := split(sentence, " ")

for {
  token, ok := iter()
  if !ok { break }
  
  fmt.Println(token)
}
```

* Here, we use the split func to split the sentence into words and then get *a new iterator func as a result value* and put it into the iter variable.
* Then, we start an infinite loop which terminates only when the iter func returns false.
* Each call to the iter func returns the next word.

#### *The result:*

```
The
quick
brown
fox
jumps
over
the
lazy
dog
```

[![run the code]](http://www.z4a.net/images/2017/11/27/run_the_code.png)

[run the code]: https://play.golang.org/p/AI1_5BkO1d	"closure"

<p align="center">Again, more explanations are inside.</p>

---

### Defer funcs

A deferred func is only executed after its parent func returns. Multiple defers can be used as well, they run as a [stack](https://en.wikipedia.org/wiki/Stack_%28abstract_data_type%29), one by one.

***👉 To learn more about them check out my post about Go defers, **[**here**](https://blog.learngoprogramming.com/golang-defer-simplified-77d3b2b817ff)**.***

![defer func](http://www.z4a.net/images/2017/11/27/defer_funcs.png)

---

### Concurrent funcs

`go func()` runs the passed func concurrently with the other goroutines.

*A goroutine is a lighter thread mechanism which allows you to structure concurrent programs efficiently. The main func executes in the main-goroutine.*



#### Example

Here, “start” anonymous func becomes a concurrent func that doesn’t block its parent func’s execution when called with “go” keyword:

```go
start := func() {
  time.Sleep(2 * time.Second)
  fmt.Println("concurrent func: ends")
}

go start()

fmt.Println("main: continues...")
time.Sleep(5 * time.Second)
fmt.Println("main: ends")
```

#### Output:

```
main: continues...
concurrent func: ends
main: ends
```

![concurrent funs](http://www.z4a.net/images/2017/11/27/concurrent_funcs.png)

<h5 align="center"><i>Main func would terminate without waiting for the concurrent func to finish if there were no sleep call in the main func:</i></h5>

```
main: continues...
main: ends
```

[![run the code]](http://www.z4a.net/images/2017/11/27/run_the_code.png)

[run the code]: https://play.golang.org/p/UzbtrKxBna	"concurrent"

---

### Other Types

#### Recursive funcs

You can use recursive funcs as in any other langs, there is no real practical difference in Go. However, you must not forget that each call creates a new [call stack](https://en.wikipedia.org/wiki/Call_stack#Functions_of_the_call_stack). But, in Go, stacks are dynamic, they can shrink and grow depending on the needs of a func. If you can solve the problem at hand without a recursion prefer that instead.



#### Black hole funcs

A black hole func can be defined multiple times and they can’t be called in the usual ways. They’re sometimes useful to test a parser: see [this.](https://github.com/golang/tools/blob/master/imports/imports.go#L167)](https://github.com/golang/tools/blob/master/imports/imports.go#L167)

```go
func _() {}
func _() {}
```



#### Inlined funcs

Go linker places a func into an executable to be able to call it later at the run-time. Sometimes calling a func is an expensive operation compared to executing the code directly. So, the compiler injects func’s body into the caller. 

To learn more about them: Read [**this**](https://github.com/golang/proposal/blob/master/design/19348-midstack-inlining.md) and [**this**](http://www.agardner.me/golang/garbage/collection/gc/escape/analysis/2015/10/18/go-escape-analysis.html) and [**this**](https://medium.com/@felipedutratine/does-golang-inline-functions-b41ee2d743fa) *(by: [Felipe](https://medium.com/@felipedutratine))  and [**this**](https://github.com/golang/go/issues/17373).*



#### External funcs

If you omit the func’s body and only declare its signature, the linker will try to find it in an external func that may have written elsewhere. As an example, [*Atan func here*](https://github.com/golang/go/blob/dd8dc6f0595ffc2c4951c0ce8ff6b63228effd97/src/pkg/math/atan.go#L54) just declared with a signature and then [*implemented in here*](https://github.com/golang/go/blob/dd8dc6f0595ffc2c4951c0ce8ff6b63228effd97/src/pkg/math/atan_386.s).

---


----------------

via: https://blog.learngoprogramming.com/go-functions-overview-anonymous-closures-higher-order-deferred-concurrent-6799008dde7b

作者：[Inanc Gumus](https://blog.learngoprogramming.com/@inanc)
译者：[译者ID](https://github.com/译者ID)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
