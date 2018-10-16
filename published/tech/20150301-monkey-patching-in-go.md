首发于：https://studygolang.com/articles/14043

# Go 语言中的 Monkey 补丁

很多人认为 monkey 补丁只能在动态语言，比如 Ruby 和 Python 中才存在。但是，这并不对。因为计算机只是很笨的机器，我们总能让它做我们想让它做的事儿！让我们看看 Go 中的函数是怎么工作的，并且，我们如何在运行时修改它们。本文会用到大量的 Intel 汇编，所以，我假设你可以读汇编代码，或者在读本文时正拿着[参考手册](https://software.intel.com/en-us/articles/introduction-to-x64-assembly).

**如果你对 monkey 补丁是怎么工作的不感兴趣，你只是想使用它的话，你可以在[这里](https://github.com/bouk/monkey)找到对应的库文件**

让我们看看一下代码产生的汇编码：

```go
package main

func a() int { return 1 }

func main() {
  print(a())
}
```

> example1.go 由 GitHub 托管   [查看源文件](https://gist.github.com/bouk/17262666fae75dd24a25/raw/712ae5ef5b1becf4f782d96ca0be0d67ccdcf061/example1.go)

上述代码应该用 go build -gcflags=-l 来编译，以避免内联。在本文中我假设你的电脑架构是 64 位，并且你使用的是一个基于unix 的操作系统比如 Mac OSX 或者某个 Linux 系统。

当代码编译后，我们用 [Hopper](http://hopperapp.com/) 来查看，可以看到如上代码会产生如下汇编代码：

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/monkey-patch/hopper-1.png)

我将引用屏幕左侧显示的各种指令的地址。

我们的代码从 main.main 过程开始，从 0x2010 到 0x2026 的指令构建堆栈。你可以在[这儿](http://dave.cheney.net/2013/06/02/why-is-a-goroutines-stack-infinite)获得更多的相关知识，本文后续的篇幅里，我将忽略这部分代码。

0x202a 行是调用 0x2000 行的 main.a 函数，这个函数只是简单的将 0x1 压入堆栈然后就返回了。0x202f 到 0x2037这几行 将这个值传递给 runtime.printint.

足够简单！现在让我们看看在 Go 语言中 函数的值是怎么实现的。

## Go 语言中的函数值是如何工作的

看看下面的代码：
```go
package main

import (
  "fmt"
  "unsafe"
)

func a() int { return 1 }

func main() {
  f := a
  fmt.Printf("0x%x\n", *(*uintptr)(unsafe.Pointer(&f)))
}
```
> funcval.go 由 GitHub 托管 [查看源文件](https://gist.github.com/bouk/c921c3627ddbaae05356/raw/4c18dbaa7cfeb06b74007b65649d85f65384841a/funcval.go)

我在第11行 将 a 赋值给 f，这意味者，执行 f() 就会调用 a。然后我使用 Go 中的 [unsafe](http://golang.org/pkg/unsafe/) 包来直接读出 f 中存储的值。如果你是有 C 语言的开发背景 ，你可以会觉得 f 就是一个简单的函数指针，并且这段代码会输出 0x2000 （我们在上面看到的 main.a 的地址）。当我在我的机器上运行时，我得到的是 0x102c38, 这个地址甚至与我们的代码都不挨着！当反编译时，这就是上面第11行所对应的：

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/monkey-patch/hopper-2.png)

这里提到了一个 main.a.f，当我们查它的地址，我们可以看到这个：
![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/monkey-patch/hopper-3.png)

啊哈！main.a.f 的地址是 0x102c38，并且保存的值是 0x2000，而这个正是 main.a 的地址。看起来 f 并不是一个函数指针，而是一个指向函数指针的指针。让我们修改一下代码，以消除其中的偏差。

```go
package main

import (
  "fmt"
  "unsafe"
)

func a() int { return 1 }

func main() {
  f := a
  fmt.Printf("0x%x\n", **(**uintptr)(unsafe.Pointer(&f)))
}
```
> funcval2.go 由GitHub托管 [查看源文件](https://gist.github.com/bouk/c470c4d80ae80d7b30af/raw/d8bd9cd2b80cad288993d5e8f67b115440c6c2a3/funcval2.go)

现在输出的正是预期中的 0x2000。我们可以在[这里](https://github.com/golang/go/blob/e9d9d0befc634f6e9f906b5ef7476fbd7ebd25e3/src/runtime/runtime2.go#L75-L78)找到一点为什么代码要这样写的线索。在 Go 语言中函数值可以包含额外的信息，闭包和绑定实例方法借此实现的。

让我们看看调用一个函数值是怎么工作的。我把上面的代码修改一下，在给 f 赋值后直接调用它。

```go
package main

func a() int { return 1 }

func main() {
	f := a
	f()
}
```
> callfuncval.go 由 GitHub 托管  [查看源文件](https://gist.github.com/bouk/58bba533fb3b742ed964/raw/41821274ea8684f7b4c59e81dcc9df6c869c5bfd/callfuncval.go)

当我们反编译后我们可以看到：
![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/monkey-patch/hopper-4.png)

main.a.f 的地址被加载到 rdx，然后无论 rdx 指向啥都会被加载到 rbx 中，然后 rbx 会被调用。函数的地址都会被首先加载到 rdx 中，然后被调用的函数可以用来加载一些额外的可能用到的信息。对绑定实例方法和匿名闭包函数来说，额外的信息就是一个指向实例的指针。如果你希望了解更多，我建议你用反编译器自己深入研究下。

让我们用刚学到的知识在 Go 中实现 monkey 补丁。

## 运行期替换一个函数

我们希望做到的是，让下面的代码输出 2：

```go
package main

func a() int { return 1 }
func b() int { return 2 }

func main() {
	replace(a, b)
	print(a())
}
```
> replace.go 由GitHub托管 [查看源文件](https://gist.github.com/bouk/713f3df2115e1b5e554d/raw/65335f4e7d9d0e11a5f72e78d617ec51249c577b/replace.go)

现在我们该怎么实现这种替换？我们需要修改函数 a 跳到 b 的代码，而不是执行它自己的函数体。本质上，我们需要这么替换，把 b 的函数值加载到 rdx 然后跳转到 rdx 所指向的地址。

```
mov rdx, main.b.f ; 48 C7 C2 ?? ?? ?? ??
jmp [rdx] ; FF 22
```
> replacement.asm 由 GitHub 托管 [查看源文件](https://gist.github.com/bouk/713f3df2115e1b5e554d/raw/65335f4e7d9d0e11a5f72e78d617ec51249c577b/replace.go)

我将上述代码编译后产生的对应的机器码列出来了（用在线编译器，比如[这个](https://defuse.ca/online-x86-assembler.htm)，你可以随意尝试编译）。很明显，我们需要写一个能产生这样机器码的函数，它应该看起来像这样：

```go
func assembleJump(f func() int) []byte {
  funcVal := *(*uintptr)(unsafe.Pointer(&f))
  return []byte{
    0x48, 0xC7, 0xC2,
    byte(funcval >> 0),
    byte(funcval >> 8),
    byte(funcval >> 16),
    byte(funcval >> 24), // MOV rdx, funcVal
    0xFF, 0x22,          // JMP [rdx]
  }
}
```
> assemble_jump.go  由 GitHub 托管 [查看源文件](https://gist.github.com/bouk/4ed563abdcd06fc45fa0/raw/fa9c65c2d5828592e846e28136871ee0bd13e5a9/assemble_jump.go)

现在万事俱备，我们已经准备好将 a 的函数体替换为从 a 跳转到 b了！下述代码尝试直接将机器码拷贝到函数体中。

```go
package main

import (
	"syscall"
	"unsafe"
)

func a() int { return 1 }
func b() int { return 2 }

func rawMemoryAccess(b uintptr) []byte {
	return (*(*[0xFF]byte)(unsafe.Pointer(b)))[:]
}

func assembleJump(f func() int) []byte {
	funcVal := *(*uintptr)(unsafe.Pointer(&f))
	return []byte{
		0x48, 0xC7, 0xC2,
		byte(funcVal >> 0),
		byte(funcVal >> 8),
		byte(funcVal >> 16),
		byte(funcVal >> 24), // MOV rdx, funcVal
		0xFF, 0x22,          // JMP [rdx]
	}
}

func replace(orig, replacement func() int) {
	bytes := assembleJump(replacement)
	functionLocation := **(**uintptr)(unsafe.Pointer(&orig))
	window := rawMemoryAccess(functionLocation)

	copy(window, bytes)
}

func main() {
	replace(a, b)
	print(a())
}
```
> patch_attempt.go  由 GitHub 托管  [查看源文件](https://gist.github.com/bouk/4ed563abdcd06fc45fa0/raw/fa9c65c2d5828592e846e28136871ee0bd13e5a9/assemble_jump.go)

然而，运行上述代码并没有达到我们的目的，实际上，它会产生一个段错误。这是因为[默认情况](https://en.wikipedia.org/wiki/Segmentation_fault#Writing_to_read-only_memory)下，已经加载的二进制代码是不可写的。我们可以用 mprotect 系统调用来取消这个保护，并且这个最终版本的代码就像我们期望的一样，把函数 a 替换成了 b，然后 '2' 被打印出来。

```go
package main

import (
	"syscall"
	"unsafe"
)

func a() int { return 1 }
func b() int { return 2 }

func getPage(p uintptr) []byte {
	return (*(*[0xFFFFFF]byte)(unsafe.Pointer(p & ^uintptr(syscall.Getpagesize()-1))))[:syscall.Getpagesize()]
}

func rawMemoryAccess(b uintptr) []byte {
	return (*(*[0xFF]byte)(unsafe.Pointer(b)))[:]
}

func assembleJump(f func() int) []byte {
	funcVal := *(*uintptr)(unsafe.Pointer(&f))
	return []byte{
		0x48, 0xC7, 0xC2,
		byte(funcVal >> 0),
		byte(funcVal >> 8),
		byte(funcVal >> 16),
		byte(funcVal >> 24), // MOV rdx, funcVal
		0xFF, 0x22,          // JMP rdx
	}
}

func replace(orig, replacement func() int) {
	bytes := assembleJump(replacement)
	functionLocation := **(**uintptr)(unsafe.Pointer(&orig))
	window := rawMemoryAccess(functionLocation)

	page := getPage(functionLocation)
	syscall.Mprotect(page, syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC)

	copy(window, bytes)
}

func main() {
	replace(a, b)
	print(a())
}
```
> patch_success.go  由 GitHub 托管 [查看源文件](https://gist.github.com/bouk/55900e1d964099368ab0/raw/976376012f9b073417a6cb68960458392b7f7952/patch_success.go)

## 包装成一个很好的库

我将上述代码包装为一个[易用的库](https://github.com/bouk/monkey)。它支持 32 位，取消补丁，以及对实例方法进行补丁，并且我在 README 中写了几个使用示例。

## 总结

有志者事竟成！让一个程序在运行期修改自身是可能的，这让我们可以实现一些很酷的事儿，比如 monkey 补丁。

我希望你从这边博客中有些收获，而我在写这篇文章时很开心！

---

via: https://bou.ke/blog/monkey-patching-in-go/

作者：[Bouke](https://github.com/bouk)
译者：[MoodWu](https://github.com/MoodWu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
