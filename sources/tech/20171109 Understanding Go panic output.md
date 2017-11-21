My code has a bug. 😭

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x30 pc=0x751ba4]
goroutine 58 [running]:
github.com/joeshaw/example.UpdateResponse(0xad3c60, 0xc420257300, 0xc4201f4200, 0x16, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, ...)
        /go/src/github.com/joeshaw/example/resp.go:108 +0x144
github.com/joeshaw/example.PrefetchLoop(0xacfd60, 0xc420395480, 0x13a52453c000, 0xad3c60, 0xc420257300)
        /go/src/github.com/joeshaw/example/resp.go:82 +0xc00
created by main.runServer
        /go/src/github.com/joeshaw/example/cmd/server/server.go:100 +0x7e0
```

This panic is caused by dereferencing a nil pointer, as indicated by the first line of the output. These types of errors are much less common in Go than in other languages like C or Java thanks to Go’s idioms around error handling.

If a function _could_ fail, the function must return an `error` as its last return value. The caller should immediately check for errors from that function.

```go
// val is a pointer, err is an error interface value
val, err := somethingThatCouldFail()
if err != nil {
    // Deal with the error, probably pushing it up the call stack
    return err
}

// By convention, nearly all the time, val is guaranteed to not be
// nil here.
```

However, there must be a bug somewhere that is violating this implicit API contract.

Before I go any further, a caveat: this is architecture- and operating system-dependent stuff, and I am only running this on amd64 Linux and macOS systems. Other systems can and will do things differently.

Line two of the panic output gives information about the UNIX signal that triggered the panic:

```
[signal SIGSEGV: segmentation violation code=0x1 addr=0x30 pc=0x751ba4]
```

A segmentation fault (`SIGSEGV`) occurred because of the nil pointer dereference. The `code` field maps to the UNIX `siginfo.si_code` field, and a value of `0x1` is `SEGV_MAPERR` (“address not mapped to object”) in Linux’s `siginfo.h` file.

`addr` maps to `siginfo.si_addr` and is `0x30`, which isn’t a valid memory address.

`pc` is the program counter, and we could use it to figure out where the program crashed, but we conveniently don’t need to because a goroutine trace follows.

```
goroutine 58 [running]:
github.com/joeshaw/example.UpdateResponse(0xad3c60, 0xc420257300, 0xc4201f4200, 0x16, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, ...)
        /go/src/github.com/joeshaw/example/resp.go:108 +0x144
github.com/joeshaw/example.PrefetchLoop(0xacfd60, 0xc420395480, 0x13a52453c000, 0xad3c60, 0xc420257300)
        /go/src/github.com/joeshaw/example/resp.go:82 +0xc00
created by main.runServer
        /go/src/github.com/joeshaw/example/cmd/server/server.go:100 +0x7e0

The deepest stack frame, the one where the panic happened, is listed first. In this case, `resp.go` line 108.

The thing that catches my eye in this goroutine backtrace are the arguments to the `UpdateResponse` and `PrefetchLoop` functions, because the number doesn’t match up to the function signatures.

```go
func UpdateResponse(c Client, id string, version int, resp *Response, data []byte) error
func PrefetchLoop(ctx context.Context, interval time.Duration, c Client)
```

`UpdateResponse` takes 5 arguments, but the panic shows that it takes more than 10\. `PrefetchLoop` takes 3, but the panic shows 5\. What’s going on?

To understand the argument values, we have to understand a little bit about the data structures underlying Go types. Russ Cox has two great blog posts on this, one on [basic types, structs and pointers, strings, and slices](https://research.swtch.com/godata) and another on [interfaces](https://research.swtch.com/interfaces) which describe how these are laid out in memory. Both posts are essential reading for Go programmers, but to summarize:

*   Strings are two words (a pointer to string data and a length)
*   Slices are three words (a pointer to a backing array, a length, and a capacity)
*   Interfaces are two words (a pointer to the type and a pointer to the value)

When a panic happens, the arguments we see in the output include the “exploded” values of strings, slices, and interfaces. In addition, the return values of a function are added onto the end of the argument list.

To go back to our `UpdateResponse` function, the `Client` type is an interface, which is 2 values. `id` is a string, which is 2 values (4 total). `version` is an int, 1 value (5). `resp` is a pointer, 1 value (6). `data` is a slice, 3 values (9). The `error` return value is an interface, so add 2 more for a total of 11\. The panic output limits the number to 10, so the last value is truncated from the output.

Here is an annotated `UpdateResponse` stack frame:

```
github.com/joeshaw/example.UpdateResponse(
    0xad3c60,      // c Client interface, type pointer
    0xc420257300,  // c Client interface, value pointer
    0xc4201f4200,  // id string, data pointer
    0x16,          // id string, length (0x16 = 22)
    0x1,           // version int (1)
    0x0,           // resp pointer (nil!)
    0x0,           // data slice, backing array pointer (nil)
    0x0,           // data slice, length (0)
    0x0,           // data slice, capacity (0)
    0x0,           // error interface (return value), type pointer
    ...            // truncated; would have been error interface value pointer
)
```

This helps confirm what the source suggested, which is that `resp` was `nil` and being dereferenced.

Moving up one stack frame to `PrefetchLoop`: `ctx context.Context` is an interface value, `interval` is a `time.Duration` (which is just an `int64`), and `Client` again is an interface.

`PrefetchLoop` annotated:

```
github.com/joeshaw/example.PrefetchLoop(
    0xacfd60,       // ctx context.Context interface, type pointer
    0xc420395480,   // ctx context.Context interface, value pointer
    0x13a52453c000, // interval time.Duration (6h0m)
    0xad3c60,       // c Client interface, type pointer
    0xc420257300,   // c Client interface, value pointer
)
```

As I mentioned earlier, it should not have been possible for `resp` to be `nil`, because that should only happen when the returned error is not `nil`. The culprit was in code which was erroneously using the `github.com/pkg/errors` `Wrapf()` function instead of `Errorf()`.

```go
// Function returns (*Response, []byte, error)

if resp.StatusCode != http.StatusOK {
    return nil, nil, errors.Wrapf(err, "got status code %d fetching response %s", resp.StatusCode, url)
}
```

`Wrapf()` returns `nil` if the error passed into it is `nil`. This function erroneously returned `nil, nil, nil` when the HTTP status code was not `http.StatusOK`, because a non-200 status code is not an error and thus `err` was `nil`. Replacing the `errors.Wrapf()` call with `errors.Errorf()` fixed the bug.

Understanding and contextualizing panic output can make tracking down errors much easier! Hopefully this information will come in handy for you in the future.

Thanks to Peter Teichman, Damian Gryski, and Travis Bischel who all helped me decode the panic output argument lists.

----------------

via: https://joeshaw.org/understanding-go-panic-output/

作者：[Joe Shaw](https://joeshaw.org/about/)
译者：[译者ID](https://github.com/译者ID)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
