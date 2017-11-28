【翻译中】 by liuxinyu123
# Golang Pros and Cons for DevOps (Part 1 of 6): Goroutines, Panics, and Errors

![Golang Pros and Cons for DevOps (Part 1 of 6): Goroutines, Panics, and Errors](https://blog.bluematador.com/hubfs/Blue_Matador_Inc_October2017/Images/golang-pros-cons-for-devops-part-1-goroutines-panics-errors.png?t=1511282519832)

*Google Go can be the perfect language for your next DevOps application. As the first post in a six-part series, we delve into golang's pros and cons as they relate to building DevOps programs, starting with goroutines, panics, and errors in this one.*

We’ve lauded the merits of Google’s golang for DevOps applications here on in this blog, and we’ve also written a [mini guide on getting started with Go](https://blog.bluematador.com/posts/mini-guide-google-golang-why-its-perfect-for-devops/?utm_source=bm-blog&utm_medium=link&utm_campaign=golang-pros-cons-1).

But lest anyone think we are underground Googlers (which we aren’t) in the guise of a DevOps monitoring platform (which we are), we’d like to delve into the language’s pros _and_ cons, especially as it relates to building DevOps applications. We’ve had a lot of experience working in golang as we converted our smart agent from Python to Go, and we have some things we think are important to share with the larger DevOps community.

Starting this week, we will be posting a 6-part series on golang pros and cons, each post detailing a few of each. As we do, we’ll update this post with links to the others:

*   **Golang Pros & Cons for DevOps #1: Goroutines, Channels, Panics, and Errors [This post]**
*   [Golang Pros & Cons for DevOps #2: Auto Interface Implementation, Public/Private Variables](/blog/posts/golang-pros-cons-for-devops-part-2/)
*   [Golang Pros & Cons for DevOps #3: Speed vs Lack of Generics](/blog/posts/golang-pros-cons-devops-part-3-speed-lack-generics/)
*   [Golang Pros & Cons for DevOps #4: The Time Package and Method Overloading](/golang-pros-cons-part-4-time-package-method-overloading)
*   Golang Pros & Cons for DevOps #5: Cross-Platform Compilation, Windows, Signals, Docs, and the Compiler
*   Golang Pros & Cons for DevOps #6: Defer Statements and Package Dependency Versioning

If this is your first time reading about Go, and you already know how to program in a C-like language, you should head over to the [Go Tour](https://tour.golang.org/welcome/1), which takes about an hour, and is pretty in depth. What rantings follow are not a great way to learn how to program Go. Instead, they are the gripes and saving graces we’ve found while developing our smart agent in Go.

Ready to hear what it’s really like to program in Go? Here it goes (sorry for the [bad pun](http://knowyourmeme.com/memes/lame-pun-coon)).

## Golang Pro #1: Goroutines — Lightweight, Kernel-level Threads

Goroutines are tantamount to Go. Goroutines are execution threads, and they’re lightweight and kernel-level. _Lightweight_ because you can run a lot of them without impacting system performance, and _kernel-level_ because they run in parallel (for real, instead of just pretending, like in some other languages, like old versions of Ruby). Also, unlike Python, there is no global interpreter lock (GIL) in Go.

In other words, goroutines are a fundamental part of golang instead of being an afterthought, which is one of the reasons why golang is so fast. Java, on the other hand, has a bureaucracy of thread schedulers. They do threading, but it’s a lot of management and just slows down your application. On the other hand of the spectrum, Node.js has no actual concurrency, just the illusion of such. That’s why Go can be a great language for DevOps purposes; it doesn’t impact your system performance, actually runs in parallel, and is as reliable as possible during unexpected hardware malfunctions.

### How to Execute a Goroutine

It’s super simple to start a goroutine thread. First, some pseudo-code that you would find in our Watchdog product:

```go
import "time"
func monitorCpu() { … }
func monitorDisk() { … }
func monitorNetwork() { … }
func monitorProcesses() { … }
func monitorIdentity() { … }

func main() {
	for !shutdown {
		monitorCpu()
		monitorDisk()
		monitorNetwork()
		monitorProcesses()
		monitorIdentity()
		time.Sleep(5*time.Second)
	}
}
```

In the code above, the CPU, disk, network, processes, and identity are checked every 5 seconds. If one of them ever hangs, all monitoring stops. Also, the longer each one takes, the less frequent the checks become, because we only sleep for 5 seconds after the computation is finished.

To fix these problems, one solution (an incomplete one, but perfect for showing the value of goroutines) is to call each one of these monitor functions using a goroutine. Simply add `go` in front of any function calls that you want to spawn in threads.

```go
func main() {
	for !shutdown {
		go monitorCpu()
		go monitorDisk()
		go monitorNetwork()
		go monitorProcesses()
		go monitorIdentity()
		time.Sleep(5*time.Second)
	}
}
```

Now, if any one of them hangs, instead of blocking all the other monitoring calls, only the blocked call will be stopped. Also, because spawning a thread is so easy and quick, we’re now a lot closer to actually checking these systems every 5 seconds.

Sure, the solution above has other issues, like a panic in one goroutine can ruin the other goroutines, the sleep time is off by just a small amount, the code’s not modular like it should be, etc etc. But isn’t spawning a kernel-level thread super easy?

As you can see, Java takes 12 lines of code Java, while Go takes just two words. While we would love to say that this means your golang source is going to be lean and compact, when we introduce _panics_ and _errors_ later in this post, you’ll see that this is unfortunately not the case. (Golang actually tends to be a bit bloated of a language, believe it or not now.)

### Sync Package (and Channels) Adds Orchestration

We use Go’s sync package and channels for goroutine orchestration, signaling, and shutdown. Littered throughout our code, you’ll find references to sync.Mutex, sync.WaitGroup, and a repeated struct variable called shutdownChannel.

A critical note about sync.Mutex and sync.WaitGroup from the [Go manual](https://golang.org/pkg/sync/):

> Values containing the types defined in this package should not be copied.

Translated for people who don’t use Go, C, or C++ all day long: Structs are pass-by-value. Any time you create a Mutex or WaitGroup, use a pointer, not a straight value. This isn’t universally necessary, but if you don’t know when it’s good and when it’s bad, just always use pointers. Here’s a good, simple example:

```go
type Example struct {
	wg *sync.WaitGroup
	m *sync.Mutex
}

func main() {
	wg := &sync.WaitGroup{}
	m := &sync.Mutex{}
}
```

While the warning about these structs is right at the top of the page, it’s easy to gloss over, and will cause the app you’re building to have really strange side effects.

From the example about monitoring in the previous section, here’s how we would use a wait group to make sure we never have more than one monitor per system outstanding at any time:

```go
import "sync"
…
func main() {
	for !shutdown {
		wg := &sync.WaitGroup{}

		doCall := func(fn func()) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fn()
			}
		}

		doCall(monitorCpu)
		doCall(monitorDisk)
		doCall(monitorNetwork)
		doCall(monitorProcesses)
		doCall(monitorIdentity)

		wg.Wait()
	}
}
```

A mutex is great to protect shared resources, like the history of CPU metrics on a server, the watermark for a log file being watched, or a list of listeners interested in an update event. No surprises here, except for the [`defer`](https://tour.golang.org/flowcontrol/12) keyword, which is pretty awesome, but outside the scope of this post.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func printAndSleep(m *sync.Mutex, x int) {
	m.Lock()
	defer m.Unlock()
	fmt.Println(x)
	time.Sleep(time.Second)
}

func main() {
	m := &sync.Mutex{}
	for i := 0; i < 10; i++ {
		printAndSleep(m, i)
	}
}
```

Not only are goroutines easy to start, they’re also easy to coordinate, shut down, and wait for in aggregate. There are a couple tricky things left to tackle that will come up a bit further down: event broadcasting to multiple threads, worker pools, and distributed processing.

### Goroutines Clean Themselves Up Automatically

Goroutines keep stack and heap variable references (avoiding garbage collection) but need not be referred to themselves to stick around. They will run until the function is complete and then shut down and release all resources automatically. The one tricky part about this process is that launched goroutines will be ignored if the main thread exits.

First, a real example about launch-and-forget goroutines. In our application, we launch modules in their own subprocess, and use IPC for config updates, settings updates, and heartbeats. Both the parent process and each module process must read from the IPC channel constantly, and then send that information elsewhere. This is the kind of thread that we launch and forget about, because we don’t care if we didn’t read the whole stream on shutdown. Take this code with a grain of salt. While it comes from our codebase, some weightier lines have been removed for simplicity:

```go
package ipc

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"time"
)

type ProtocolReader struct {
	Channel chan *ProtocolMessage
	reader  *bufio.Reader
	handle  *os.File
}

func NewProtocolReader(handle *os.File) *ProtocolReader {
	return &ProtocolReader{
		make(chan *ProtocolMessage, 15),
		bufio.NewReader(handle),
		handle,
	}
}

func (this *ProtocolReader) ReadAsync() {
	go func() {
		for {
			line, err := this.reader.ReadBytes('\n')
			if err != nil {
				this.handle.Close()
				close(this.Channel)
				return nil
			}

			message := &ProtocolMessage{}
			message.Unmarshal(line)
			this.Channel <- message
		}

		return nil
	}
}
```

The second example illustrates the main thread exiting and ignoring running goroutines:

```go
package main

import (
	"fmt"
	"time"
)

func waitAndPrint() {
	time.Sleep(time.Second)
	fmt.Println("got it!")
}

func main() {
	go waitAndPrint()
}
```

It’s easy to fix this using a `sync.WaitGroup`. You’ll see a lot of examples using time.Sleep to wait in sample code like this. We will _definitely_ think less of you if you’re contributing to that madness. Just use a WaitGroup and code on.

### Channels

Golang channels are great uni-directional message passing tools.We use them in our agent code for message passing, message broadcasting, and worker queues. They don’t have to be closed, are automatically cleaned by GC, and are simple to make:

```go
numSlots := 5
make(chan int, numSlots)
```

You can send any one thing through that channel. You can make them synchronous, asynchronous, or give them multiple readers that listen to those channels and do something with it.

Unlike queues, a channel can be used to broadcast a message. The most common message we broadcast in our code is shutdown. When it’s shutdown time, we signal to all background goroutines that it’s time to clean up. There is only one way to signal to multiple listeners a single message using a channel - you must close the channel. Here’s a watered-down version of our code:

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

var shutdownChannel = make(chan struct{}, 0)
var wg = &sync.WaitGroup{}

func start() {
	wg.Add(1)
	go func() {
		ticker := time.Tick(100*time.Millisecond)

		for shutdown := false; !shutdown; {
			select {
			case <-ticker:
				fmt.Println("tick")
			case <-shutdownChannel:
				fmt.Println("tock")
				shutdown = true
			}
		}
		wg.Done()
	}()
}

func stop() {
	close(shutdownChannel)
}

func wait() {
	wg.Wait()
}

func main() {
	start()
	time.Sleep(time.Second)
	stop()
	wait()
}
```

We **love** the `select` functionality in Go. It allows us to be responsive to interruptions while still doing important work. We use it pretty liberally for managing shutdown signals and timers (like the example above), reading from multiple data streams, and working with Go’s [fsnotify package](https://github.com/fsnotify/fsnotify).

# Golang Con #1: Dealing With Both Panics & Errors

Panics and errors. These are the worst things about Golang, and worst by a longshot. First of all, let’s define what panics and errors are, because not every language deals with them.

According to a [post on Go’s official blog](https://blog.golang.org/defer-panic-and-recover),

> Panic is a built-in function that stops the ordinary flow of control and begins panicking. When the function F calls panic, execution of F stops, any deferred functions in F are executed normally, and then F returns to its caller. To the caller, F then behaves like a call to panic. The process continues up the stack until all functions in the current goroutine have returned, at which point the program crashes. Panics can be initiated by invoking panic directly. They can also be caused by runtime errors, such as out-of-bounds array accesses.

In other words, panics kill your program when you’ve got a control flow problem.

There’s several ways to trigger a panic:

*   Calling a function a panic

*   Dividing by 0

*   Closing a channel that’s already closed

*   Mapping a non-existent attribute, such as `Attribute = map["This doesn’t exist"]`

An error, on the other hand, is a built-in type that represent values that can self-declare as strings. Here is the definition from the Go source code:

```go
type error interface {
	Error() string
}
```

Having that definition, here is the summary for why we hate having both errors and panics:

**Errors were meant to avoid exception flow, panics nullify that purpose.**

It’s enough in any language to have either errors or panics. The fact that some languages choose to have both is frustrating, to say the least. Go developers unfortunately threw their lot in with the wrong crowd by choosing both.

### A Sampling of Error Handling in Popular Languages

<table>

<tbody>

<tr>

<td>Golang</td>

<td>"Panics" (really more like "errors"), exceptions & segfault</td>

</tr>

<tr>

<td>Java</td>

<td>Exceptions</td>

</tr>

<tr>

<td>Scala</td>

<td>Exceptions</td>

</tr>

<tr>

<td>Ruby</td>

<td>"Errors" (really more like exceptions)</td>

</tr>

<tr>

<td>Python</td>

<td>Exceptions</td>

</tr>

<tr>

<td>PHP</td>

<td>Errors & exceptions</td>

</tr>

<tr>

<td>Javascript</td>

<td>Exceptions</td>

</tr>

<tr>

<td>C/C++</td>

<td>Errors, exceptions & segfault</td>

</tr>

<tr>

<td>Objective-C</td>

<td>Exceptions & errors</td>

</tr>

<tr>

<td>Swift</td>

<td>Errors</td>

</tr>

</tbody>

</table>

It’s always _possible_ to return errors, but it may not be _necessary_ for the language. Go’s built-in functions for accessing map elements, reading from channels, JSON encoding, and more require the use of error handling. This is why it, and other similar languages, received the “errors” designation, and languages like Python and Scala did not.

Again, [from the Go blog](https://blog.golang.org/error-handling-and-go),

> Error handling is important. The language’s design and conventions encourage you to explicitly check for errors where they occur (as distinct from the convention in other languages of throwing exceptions and sometimes catching them). In some cases this makes Go code verbose, but fortunately there are some techniques you can use to minimize repetitive error handling.

When they say that errors can make Go code verbose, they weren’t joking.

So what does Go’s implementation of panics and errors tell us about the language?

## Errors Increase Your Golang Code Size

Before we launch into it, let’s just point out that we could rail on panics or errors. It’s not errors that we dislike; it’s having both of them together. Given the design of the language, errors could be taken out easily, so we’re going to rail against errors, not panics.

Golang applications are already bloated in terms of codebase size. The speed at which its binaries runs fast, but as source code, it’s more verbose than it needs be. Add to this verbosity having to deal with both panics and errors, and you can start to get an idea of why the language [has been called ugly](https://www.quora.com/Do-you-feel-that-golang-is-ugly) before. It takes extra effort to distinguish between panics and errors. In a typical programming language, you have a single way to manage errors. Panics and errors together? It makes us [madder than a mosquito in a mannequin factory](https://en.wikiquote.org/wiki/Larry_the_Cable_Guy)!

Here’s why.

The usual single way to manage errors is called a try/catch.

Here’s a would-be example from our agent’s code:

```go
try {
	this.downloadModule(moduleSettings)
	this.extractModule(moduleSettings)
	this.tidyManifestModule(moduleSettings)
	this.restartCommand(moduleSettings)
	this.cleanupModule(moduleSettings)
	return nil
}
catch e {
	case Exception => return e
}
```

In this example, as in our actual code, we don’t care what the error was or where it happened. All we care about was whether or not there was an error. This way both makes sense and leads to clean, concise code.

With an error, with literally every line of your code, you have to do this:

```go
if err := this.downloadModule(moduleSettings); err != nil {
	return err
}
if err := this.extractModule(moduleSettings); err != nil {
	return err
}
if err := this.tidyManifestModule(moduleSettings); err != nil {
	return err
}
if err := this.restartCommand(currentUpdate); err != nil {
	return err
}
if err := this.cleanupModule(moduleSettings); err != nil {
	return err
}
```

In the worst case, it triples the size of your codebase. Triples! No, it’s not every line - structs, interfaces, imports, and whitespace are completely unaffected. All the other lines, you know, the ones with actual code on them? All tripled.

In a minority of cases, you want to do something different for those things. It’s dumb to turn a single line of code into three lines of code when you don’t really have to. It doesn’t buy you anything to have the code duplication expanding the size of your codebase for no good reason.

In addition to errors, you have panics. If you have something that panics, it doesn’t matter what function it happened in. If you had a panic, it would do the same thing as try/catch, except now your code has to be duplicated _and_ you still have to manage the try/catch!

The solution in our agent is to use a wrapper function with retries and good logging for both panics and errors. Then, we religiously call it on the main thread and every goroutine spawned throughout our codebase. This code won’t run anywhere because it’s missing other classes, and it’s watered down anyway, but it should give you a good idea of how we manage errors and panics together.

```go
package safefunc

import (
	"common/log"
	"common/timeout"
	"runtime/debug"
	"time"
)

type RetryConfig struct {
	MaxTries           int
	BaseDelay          time.Duration
	MaxDelay           time.Duration
	SplayFraction      float64
	ShutdownChannel    <-chan struct{}
}

func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxTries:           -1,
		BaseDelay:          time.Second,
		MaxDelay:           time.Minute,
		SplayFraction:      0.25,
		ShutdownChannel:    nil,
	}
}

func Retry(name string, config *RetryConfig, callback func() error) {
	// this is stupid, but necessary.
	// when a function panics, that function's returns are zeros.
	// that's the only way to check (can't rely on a nil error during a panic)
	var noPanicSuccess int = 1
	failedAttempts := 0

	wrapped := func() (int, error) {
		defer func() {
			if err := recover(); err != nil {
				log.Warn.Println("Recovered panic inside", name, err)
				log.Debug.Println("Panic Stacktrace", string(debug.Stack()))
			}
		}()

		return noPanicSuccess, callback()
	}

retryLoop:
	for {
		wrappedReturn, err := wrapped()
		if err != nil {
			log.Warn.Println("Recovered error inside", name, err)
			log.Debug.Println("Recovered Stacktrace", string(debug.Stack()))
		} else if wrappedReturn == noPanicSuccess {
			break retryLoop
		}

		failedAttempts++
		if config.MaxTries > 0 && failedAttempts >= config.MaxTries {
			log.Trace.Println("Giving up on retrying", name, "after", failedAttempts, "attempts")
			break retryLoop
		}

		sleep := timeout.Delay(config.BaseDelay, failedAttempts, config.SplayFraction, config.MaxDelay)
		log.Trace.Println("Sleeping for", sleep, "before continuing retry loop", name)
		sleepChannel := time.After(sleep)
		select {
		case <-sleepChannel:
		case <-config.ShutdownChannel:
			log.Trace.Println("Shutting down retry loop", name)
			break retryLoop
		}
	}
}
```

You think because Go has all this error handling you’re safe, but then a panic happens at runtime. It bubbles up the call stack unchecked, crossing goroutine boundaries until it crashes your entire program.

## Not Everyone Hates Errors & Panics in Golang

Admittedly, some people may not view this as a con. Even we at Blue Matador have differing opinions (and between the co-founders no less!).

Having errors forces you to handle the problems as they happen, not in an undefined point in the future. Errors can be typechecked, so the compiler can force you to handle them, and warn you. Golang is one of the only language that makes this possible.

Some people like panics because it gets rid of a developer’s ability to return an exception. Try/catch is a better way to do that, and it goes beyond functions. It’s difficult to follow the flow when there are exceptions, and panics/errors help developers get around that.

## Our Main Gripe with Errors & Panics

It’s just that you have to simultaneously deal with both of them in Golang. You’ve already got errors catching panics anyways through your try/catch, so why make yourself have to worry about panics separately?

Panics not only interrupt all the function calls leading up to where they currently are, they will also break threads! If you have a thread that panics, and you don’t catch it on that thread, not only will that thread cease to be, the thread that called it will cease, bubbling all the way up until your program dies. A single panic in your code can break everything because it can cause the entire program to fail by cascading up.

Go forces you to have to tolerate errors and panics both, and you just have to remember that. Permanently. And if you ever forget, your program may crash unexpectedly. You could blame it on the programmer, who rightfully wrote bad code, but is it really the developer’s fault? Is it the driver’s fault that a mechanic loosened all the lugnuts on his car?

To sum up, if everything were caught by errors, great! If everything were caught by panics, great! But having to do both of them and tolerating both of them at the same time is really frustrating in Golang.

Golang Pros & Cons for DevOps #2: Auto Interface Implementation, Public/Private Variables

Every other week, we are posting a new guide like this in our six-part series on “Golang Pros & Cons for DevOps.” Next up: Auto interfaces and public/private variables.

### Author Bio

Matthew has experienced the pain of being on call with normal DevOps monitoring tools. He founded Blue Matador to fix the mess of Frankenstein installations that pervade DevOps. In his spare time, Matthew loves flying airplanes, playing boardgames, and spending time with his wife and three boys.

----------------

via: https://blog.bluematador.com/posts/golang-pros-cons-for-devops-part-1-goroutines-panics-errors/

作者：[Matthew Barlocker](https://github.com/mbarlocker)
译者：[译者ID](https://github.com/译者ID)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
