首发于：https://studygolang.com/articles/28440

# 关于 CGO 的字符串函数的解释

[cgo](https://github.com/golang/go/wiki/cgo) 的大量文档都提到过，它提供了四个用于转换 Go 和 C 类型的字符串的函数，都是通过复制数据来实现。在 CGo 的文档中有简洁的解释，但我认为解释得太简洁了，因为文档只涉及了定义中的某些特定字符串，而忽略了两个很重要的注意事项。我曾经踩过这里的坑，现在我要详细解释一下。

四个函数分别是：

```go
func C.CString(string) *C.char
func C.GoString(*C.char) string
func C.GoStringN(*C.char, C.int) string
func C.GoBytes(unsafe.Pointer, C.int) []byte
```

`C.CString()` 等价于 C 的 `strdup()`，像文档中提到的那样，把 Go 的字符串复制为可以传递给 C 函数的 C 的 `char *`。很讨厌的一件事是，由于 Go 和 CGo 类型的定义方式，调用 `C.free` 时需要做一个转换：

```go
cs := C.CString("a string")
C.free(unsafe.Pointer(cs))
```

请留意，Go 字符串中可能嵌入了 `\0` 字符，而 C 字符串不会。如果你的 Go 字符串中有 `\0` 字符，当你调用 `C.CString()` 时，C 代码会从 `\0` 字符处截断你的字符串。这往往不会被注意到，但[有时文本并不保证不含 null 字符](https://utcc.utoronto.ca/~cks/space/blog/programming/BeSureItsACString)。

`C.GoString()` 也等价于 `strdup()`，但与 `C.CString()` 相反，是把 C 字符串转换为 Go 字符串。你可以用它定义结构体的字段，或者是声明为 C 的 `char *`（在 Go 中叫 `*C.cahr`） 的其他变量，抑或其他的一些变量（我们后面会看到）。

`C.GoStringN()` 等价于 C 的 `memmove()`，与 C 中普通的字符串函数不同。**它把整个 N 长度的 C buffer 复制为一个 Go 字符串，不单独处理 null 字符。**再详细点，它也通过复制来实现。如果你有一个定义为 `char feild[64]` 的结构体的字段，然后调用了 `C.GoStringN(&field, 64)`，那么你得到的 Go 字符串一定是 64 个字符，字符串的末尾有可能是一串 `\0` 字符。

(我认为这是 cgo 文档中的一个 bug。它宣称 GoStringN 的入参是一个 C 的字符串，但实际上很明显不是，因为 C 的字符串不能以 null 字符结束，而 GoStringN 不会在 null 字符处结束处理。)

`C.GoBytes()` 是 `C.GoStringN()` 的另一个版本，不返回 `string` 而是返回 `[]byte`。它没有宣称以 C 字符串作为入参，它仅仅是对整个 buffer 做了内存拷贝。

如果你要拷贝的东西不是以 null 字符结尾的 C 字符串，而是固定长度的 memory buffer，那么 `C.GoString()` 正好能满足需求；它避开了 C 中传统的问题[处理不是 C 字符串的 ’string‘](https://utcc.utoronto.ca/~cks/space/blog/programming/BeSureItsACString)。然而，如果你要处理定义为 `char field[N]` 的结构体字段这种限定长度的 C 字符串时，这些函数*都不能*满足需求。

传统语义的结构体中固定长度的字符串变量，定义为 `char field[N]` 的字段，以及“包含一个字符串”等描述，都表示当且仅当字符串有足够空间时以 null 字符结尾，换句话说，字符串最多有 N-1 个字符。如果字符串正好有 N 个字符，那么它不会以 null 字符结尾。这是 [C 代码中诸多 bug 的根源](https://utcc.utoronto.ca/~cks/space/blog/programming/UnixAPIMistake)，也不是一个好的 API，但我们却摆脱不了这个 API。每次我们遇到这样的字段，文档不会明确告诉你字段的内容并不一定是 null 字符结尾的，你需要自己假设你有这种 API。

`C.GoString()` 或 `C.GoStringN()` 都不能正确处理这些字段。使用 `GoStringN()` 相对来说出错更少；它仅仅返回一个末尾有一串 `\0` 字符长度为 N 的 Go 字符串（如果你仅仅是把这些字段打印出来，那么你可能不会留意到；我经常干这种事）。使用有诱惑力的 `GoString()` 更是引狼入室，因为它内部会对入参做 `strlen()`；如果字符末尾没有 null 字符，`strlen()` 会访问越界的内存地址。如果你走运，你得到的 Go 字符串末尾会有大量的垃圾。如果你不走运，你的 Go 程序出现段错误，因为 `strlen()` 访问了未映射的内存地址。

（总的来说，如果字符串末尾出现了大量垃圾，通常意味着在某处有不含结束符的 C 字符串。）

你需要的是与 C 的 `strndup()` 等价的 Go 函数，以此来确保复制不超过 N 个字符且在 null 字符处终止。下面是我写的版本，不保证无错误：

```go
func strndup(cs *C.char, len int) string {
   s := C.GoStringN(cs, C.int(len))
   i := strings.IndexByte(s, 0)
   if i == -1 {
      return s
   }
   return C.GoString(cs)
}
```

由于有 [Go 的字符串怎样占用内存](https://utcc.utoronto.ca/~cks/space/blog/programming/GoStringsMemoryHolding)的问题，这段代码做了些额外的工作来最小化额外的内存占用。你可能想用另一种方法，返回一个 `GoStringN()` 字符串的切片。你也可以写复杂的代码，根据 i 和 len 的不同来决定选用哪种方法。

更新：[Ian Lance Taylor 给我展示了份更好的代码](https://github.com/golang/go/issues/12428#issuecomment-136581154)：

```go
func strndup(cs *C.char, len int) string {
   return C.GoStringN(cs, C.int(C.strnlen(cs, C.size_t(len))))
}
```

是的，这里有大量的转换。这篇文章就是你看到的 Go 和 Gco 类型的结合。

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoCGoStringFunctions

作者：[ChrisSiebenmann](https://utcc.utoronto.ca/~cks/space/People/ChrisSiebenmann)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
