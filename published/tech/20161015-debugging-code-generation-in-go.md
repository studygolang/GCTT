首发于：https://studygolang.com/articles/19815

# 调试 Go 的代码生成

2016 年 10 月 15 日

*上周，我在 [dotGo](http://www.dotgo.eu/)，参加一个最棒的 Go 会议，与大西洋彼岸的 Gopher 们相聚。我做了一个简短迅速的演讲，关于使用工具链中可用工具来进行代码生成的检查。这篇文章给没能参加会议的 Gopher 们过一遍演讲的内容。幻灯片在 [go-talks](https://go-talks.appspot.com/github.com/rakyll/talks/gcinspect/talk.slide) 也能找到。*

在这篇文章中，我们将使用以下程序：

```go
package main

import "fmt"

func main() {
    sum := 1 + 1
    fmt.Printf("sum: %v\n", sum)
}
```

## 工具链

go build 是一个对用户来说囊括了一大堆东西的命令。但是，如果你需要的话，它也提供了关于它是做什么的更详细的信息。`-x` 是一个能让 Go build 输出调用了什么的标记。如果你想看看工具链的组件是什么，它们在一个什么样的序列里以及使用了什么标记的话，使用 `-x`。

```
$ go build -x
WORK=/var/folders/00/1b8h8000h01000cxqpysvccm005d21/T/go-build190726544
mkdir -p $WORK/hello/_obj/
mkdir -p $WORK/hello/_obj/exe/
cd /Users/jbd/src/hello
/Users/jbd/go/pkg/tool/darwin_amd64/compile -o $WORK/hello.a -trimpath $WORK -p main -complete -buildid d934a5702088e0fe5c931a55ff26bec87b80cbdc -D _/Users/jbd/src/hello -I $WORK -pack ./hello.go
cd .
/Users/jbd/go/pkg/tool/darwin_amd64/link -o $WORK/hello/_obj/exe/a.out -L $WORK -extld=clang -buildmode=exe -buildid=d934a5702088e0fe5c931a55ff26bec87b80cbdc $WORK/hello.a
mv $WORK/hello/_obj/exe/a.out hello
```

## 中间汇编

在 Go 里，生成真实体系特定的汇编前，有一个中间汇编阶段。编译器拿到一些 Go 文件，生成中间指令并将其增加到 `obj` 包以生成机器码。如果你对编译器在这个阶段生成的东西感兴趣， `-S` 可以让编译器将输出转存起来。

中间汇编对理解一行 Go 代码的代价通常是个很好的参考。或者比如说当你想用一个功能相当的更优化的汇编来替换一个 Go 函数时它也是很好的参考。

在这里你将看到 main.main 的输出。

```
$ go build -gcflags="-S"
# hello
"".main t=1 size=179 args=0x0 locals=0x60
    0x0000 00000 (/Users/jbd/src/hello/hello.go:5)  TEXT    "".main(SB), $96-0
    0x0000 00000 (/Users/jbd/src/hello/hello.go:5)  MOVQ    (TLS), CX
    0x0009 00009 (/Users/jbd/src/hello/hello.go:5)  CMPQ    SP, 16(CX)
    0x000d 00013 (/Users/jbd/src/hello/hello.go:5)  JLS 169
    0x0013 00019 (/Users/jbd/src/hello/hello.go:5)  SUBQ    $96, SP
    0x0017 00023 (/Users/jbd/src/hello/hello.go:5)  MOVQ    BP, 88(SP)
    0x001c 00028 (/Users/jbd/src/hello/hello.go:5)  LEAQ    88(SP), BP
    0x0021 00033 (/Users/jbd/src/hello/hello.go:5)  FUNCDATA    $0, gclocals · 69c1753bd5f81501d95132d08af04464(SB)
    0x0021 00033 (/Users/jbd/src/hello/hello.go:5)  FUNCDATA    $1, gclocals · e226d4ae4a7cad8835311c6a4683c14f(SB)
    0x0021 00033 (/Users/jbd/src/hello/hello.go:7)  MOVQ    $2, "".autotmp_1+64(SP)
    0x002a 00042 (/Users/jbd/src/hello/hello.go:7)  MOVQ    $0, "".autotmp_0+72(SP)
    0x0033 00051 (/Users/jbd/src/hello/hello.go:7)  MOVQ    $0, "".autotmp_0+80(SP)
    0x003c 00060 (/Users/jbd/src/hello/hello.go:7)  LEAQ    type.int(SB), AX
    0x0043 00067 (/Users/jbd/src/hello/hello.go:7)  MOVQ    AX, (SP)
    0x0047 00071 (/Users/jbd/src/hello/hello.go:7)  LEAQ    "".autotmp_1+64(SP), AX
    0x004c 00076 (/Users/jbd/src/hello/hello.go:7)  MOVQ    AX, 8(SP)
    0x0051 00081 (/Users/jbd/src/hello/hello.go:7)  PCDATA  $0, $1
    0x0051 00081 (/Users/jbd/src/hello/hello.go:7)  CALL    runtime.convT2E(SB)
    0x0056 00086 (/Users/jbd/src/hello/hello.go:7)  MOVQ    16(SP), AX
    0x005b 00091 (/Users/jbd/src/hello/hello.go:7)  MOVQ    24(SP), CX
    0x0060 00096 (/Users/jbd/src/hello/hello.go:7)  MOVQ    AX, "".autotmp_0+72(SP)
    0x0065 00101 (/Users/jbd/src/hello/hello.go:7)  MOVQ    CX, "".autotmp_0+80(SP)
    0x006a 00106 (/Users/jbd/src/hello/hello.go:7)  LEAQ    Go.string."sum: %v\n"(SB), AX
    0x0071 00113 (/Users/jbd/src/hello/hello.go:7)  MOVQ    AX, (SP)
    0x0075 00117 (/Users/jbd/src/hello/hello.go:7)  MOVQ    $8, 8(SP)
    0x007e 00126 (/Users/jbd/src/hello/hello.go:7)  LEAQ    "".autotmp_0+72(SP), AX
    0x0083 00131 (/Users/jbd/src/hello/hello.go:7)  MOVQ    AX, 16(SP)
    0x0088 00136 (/Users/jbd/src/hello/hello.go:7)  MOVQ    $1, 24(SP)
    0x0091 00145 (/Users/jbd/src/hello/hello.go:7)  MOVQ    $1, 32(SP)
    0x009a 00154 (/Users/jbd/src/hello/hello.go:7)  PCDATA  $0, $1
    0x009a 00154 (/Users/jbd/src/hello/hello.go:7)  CALL    fmt.Printf(SB)
    0x009f 00159 (/Users/jbd/src/hello/hello.go:8)  MOVQ    88(SP), BP
    0x00a4 00164 (/Users/jbd/src/hello/hello.go:8)  ADDQ    $96, SP
    0x00a8 00168 (/Users/jbd/src/hello/hello.go:8)  RET
    0x00a9 00169 (/Users/jbd/src/hello/hello.go:8)  NOP
    0x00a9 00169 (/Users/jbd/src/hello/hello.go:5)  PCDATA  $0, $-1
    0x00a9 00169 (/Users/jbd/src/hello/hello.go:5)  CALL    runtime.morestack_noctxt(SB)
    0x00ae 00174 (/Users/jbd/src/hello/hello.go:5)  JMP 0
    ...
```

如果你想学习关于中间汇编的更多概念以及为何它在 Go 里很重要，我强烈推荐来自今年 GopherCon [Rob Pike 的 The Design of the Go Assembler](https://www.youtube.com/watch?v=KINIAgRpkDA)。

## 反汇编器

正如我提到的，`-S` 仅仅作用于中间汇编。真实机器上的表示在最终的工件中可用。你可以使用反汇编器去检查里面有什么。对二进制或库使用 `go tool objdump` 。你可能还想使用 `-s` 来关注符号名。在这个例子里，我将对 main.main 进行转存。这里是为 `darwin/amd64` 生成的真实汇编。

```
$ go tool objdump -s main.main hello
TEXT main.main(SB) /Users/jbd/src/hello/hello.go
    hello.go:5  0x2040  65488b0c25a0080000  GS MOVQ GS:0x8a0, CX
    hello.go:5  0x2049  483b6110            CMPQ 0x10(CX), SP
    hello.go:5  0x204d  0f8696000000        JBE 0x20e9
    hello.go:5  0x2053  4883ec60            SUBQ $0x60, SP
    hello.go:5  0x2057  48896c2458          MOVQ BP, 0x58(SP)
    hello.go:5  0x205c  488d6c2458          LEAQ 0x58(SP), BP
    hello.go:7  0x2061  48c744244002000000  MOVQ $0x2, 0x40(SP)
    hello.go:7  0x206a  48c744244800000000  MOVQ $0x0, 0x48(SP)
    hello.go:7  0x2073  48c744245000000000  MOVQ $0x0, 0x50(SP)
    hello.go:7  0x207c  488d053d4d0800      LEAQ 0x84d3d(IP), AX
    ...
```

## 符号表

有时，你需要的全部只是检查符号表而不是理解代码段或数据段。类似通用的 nm 工具，Go 分发了一个让你能列出一个工件中带注记和大小的符号表的 nm 工具。如果你想看看 Go 的一个二进制或库内部是什么，导出了什么，这是个很便利的工具。

```
$ go tool nm hello
...
f4760 B __cgo_init
f4768 B __cgo_notify_runtime_init_done
f4770 B __cgo_thread_start
4fb70 T __rt0_amd64_darwin
4e220 T _gosave
4fb90 T _main
ad1e0 R _masks
4fd00 T _nanotime
4e480 T _setg_gcc
ad2e0 R _shifts
624a0 T errors.(*errorString).Error
62400 T errors.New
52470 T fmt.(*buffer).WriteRune
...
```

## 优化

和新的 SSA 后端的贡献一起，团队贡献了一个可视化所有 SSA pass 的工具。将环境变量 GOSSAFUNC 的值设置为一个函数名称然后运行 go build 命令。将会产生一个 ssa.html 文件，显示了编译器为了优化你的代码所经过的每一步。

```
$ GOSSAFUNC=main Go build && open ssa.html
```

这里是对 main 函数应用的所有 pass 的可视化结果。

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/debug-go-code-gen/ssa.png)

Go 编译器还可以标注内联和逃逸分析。如果你将 `-m=2` 标志传给编译器，它将输出关于这两个方面的优化和标注。这里我们看到 `net/context` 包相关的内联操作和逃逸分析。

```
$ go build -gcflags="-m" golang.org/x/net/context
# golang.org/x/net/context
../golang.org/x/net/context/context.go:140: can inline Background as: func() Context { return background }
../golang.org/x/net/context/context.go:149: can inline TODO as: func() Context { return todo }
../golang.org/x/net/context/go17.go:32: cannot inline WithCancel: non-leaf function
../golang.org/x/net/context/go17.go:46: cannot inline WithDeadline: non-leaf function
../golang.org/x/net/context/go17.go:61: cannot inline WithTimeout: non-leaf function
../golang.org/x/net/context/go17.go:62: inlining call to time.Time.Add method(time.Time) func(time.Duration) time.Time { time.t · 2.sec += int64(time.d · 3 / time.Duration(1000000000)); var time.nsec · 4 int32; time.nsec · 4 = <N>; time.nsec · 4 = time.t · 2.nsec + int32(time.d · 3 % time.Duration(1000000000)); if time.nsec · 4 >= int32(1000000000) { time.t · 2.sec++; time.nsec · 4 -= int32(1000000000) } else { if time.nsec · 4 < int32(0) { time.t · 2.sec--; time.nsec · 4 += int32(1000000000) } }; time.t · 2.nsec = time.nsec · 4; return time.t · 2 }
../golang.org/x/net/context/go17.go:70: cannot inline WithValue: non-leaf function
../golang.org/x/net/context/context.go:141: background escapes to heap
../golang.org/x/net/context/context.go:141:     from ~r0 (return) at ../golang.org/x/net/context/context.go:140
../golang.org/x/net/context/context.go:150: todo escapes to heap
../golang.org/x/net/context/context.go:150:     from ~r0 (return) at ../golang.org/x/net/context/context.go:149
../golang.org/x/net/context/go17.go:33: parent escapes to heap
../golang.org/x/net/context/go17.go:33:     from parent (passed to function[unknown]) at ../golang.org/x/net/context/go17.go:33
../golang.org/x/net/context/go17.go:32: leaking param: parent
../golang.org/x/net/context/go17.go:32:     from parent (interface-converted) at ../golang.org/x/net/context/go17.go:33
../golang.org/x/net/context/go17.go:32:     from parent (passed to function[unknown]) at ../golang.org/x/net/context/go17.go:33
../golang.org/x/net/context/go17.go:47: parent escapes to heap
../golang.org/x/net/context/go17.go:47:     from parent (passed to function[unknown]) at ../golang.org/x/net/context/go17.go:47
../golang.org/x/net/context/go17.go:46: leaking param: parent
../golang.org/x/net/context/go17.go:46:     from parent (interface-converted) at ../golang.org/x/net/context/go17.go:47
../golang.org/x/net/context/go17.go:46:     from parent (passed to function[unknown]) at ../golang.org/x/net/context/go17.go:47
../golang.org/x/net/context/go17.go:46: leaking param: deadline
../golang.org/x/net/context/go17.go:46:     from deadline (passed to function[unknown]) at ../golang.org/x/net/context/go17.go:46
../golang.org/x/net/context/go17.go:48: ctx escapes to heap
../golang.org/x/net/context/go17.go:48:     from ~r2 (return) at ../golang.org/x/net/context/go17.go:46
../golang.org/x/net/context/go17.go:61: leaking param: parent
../golang.org/x/net/context/go17.go:61:     from parent (passed to function[unknown]) at ../golang.org/x/net/context/go17.go:61
../golang.org/x/net/context/go17.go:71: parent escapes to heap
../golang.org/x/net/context/go17.go:71:     from parent (passed to function[unknown]) at ../golang.org/x/net/context/go17.go:71
../golang.org/x/net/context/go17.go:70: leaking param: parent
../golang.org/x/net/context/go17.go:70:     from parent (interface-converted) at ../golang.org/x/net/context/go17.go:71
../golang.org/x/net/context/go17.go:70:     from parent (passed to function[unknown]) at ../golang.org/x/net/context/go17.go:71
../golang.org/x/net/context/go17.go:70: leaking param: key
../golang.org/x/net/context/go17.go:70:     from key (passed to function[unknown]) at ../golang.org/x/net/context/go17.go:70
../golang.org/x/net/context/go17.go:70: leaking param: val
../golang.org/x/net/context/go17.go:70:     from val (passed to function[unknown]) at ../golang.org/x/net/context/go17.go:70
../golang.org/x/net/context/go17.go:71: context.WithValue(parent, key, val) escapes to heap
../golang.org/x/net/context/go17.go:71:     from ~r3 (return) at ../golang.org/x/net/context/go17.go:70
<autogenerated>:1: leaking param: .this
<autogenerated>:1:  from .this.Deadline() (receiver in indirect call) at <autogenerated>:1
<autogenerated>:2: leaking param: .this
<autogenerated>:2:  from .this.Done() (receiver in indirect call) at <autogenerated>:2
<autogenerated>:3: leaking param: .this
<autogenerated>:3:  from .this.Err() (receiver in indirect call) at <autogenerated>:3
<autogenerated>:4: leaking param: key
<autogenerated>:4:  from .this.Value(key) (parameter to indirect call) at <autogenerated>:4
<autogenerated>:4: leaking param: .this
<autogenerated>:4:  from .this.Value(key) (receiver in indirect call) at <autogenerated>:4
```

你可以使用 `-m` 查看没附带原因，不这么冗长的输出，但是 [David Chase](https://twitter.com/Dr2chase/status/788353223522455552) 说过尽管 `-m=2` 不完美，但它经常很有用。

值得一提的是你经常需要禁用优化来得到一个关于发生了什么的更简单的视图，因为优化可能会修改操作序列，增加代码，删除代码或是对代码进行变换。开启了优化，将一行 Go 代码与优化后的输出对应起来将更难，进行性能测试也会更难，因为优化可能带来不止一处变化。可以通过 `-N` 来禁用优化，通过 `-l` 来禁用内联。

```
$ go build -gcflags="-l -N"
```

一旦优化被禁用，你调试就不会被代码变化影响，进行性能测试也不会受不止一处变化的影响。

## Lexer

如果你在 lexer 上工作，编译器提供了一个标志在检查源码时调试 lexer。

```
$ go build -gcflags="-x"
# hello
lex: PACKAGE
lex: ident main
lex: implicit semi
lex: IMPORT
lex: string literal
lex: implicit semi
lex: FUNC
lex: ident main
./hello.go:5 lex: TOKEN '('
./hello.go:5 lex: TOKEN ')'
./hello.go:5 lex: TOKEN '{'
lex: ident sum
./hello.go:6 lex: TOKEN COLAS
lex: integer literal
./hello.go:6 lex: TOKEN '+'
lex: integer literal
lex: implicit semi
lex: ident fmt
./hello.go:7 lex: TOKEN '.'
lex: ident Printf
./hello.go:7 lex: TOKEN '('
lex: string literal
./hello.go:7 lex: TOKEN ','
lex: ident sum
./hello.go:7 lex: TOKEN ')'
lex: implicit semi
./hello.go:8 lex: TOKEN '}'
lex: implicit semi
```

如果你有任何建议或评论，请 ping [@rakyll](https://twitter.com/rakyll).

---

via: https://rakyll.org/codegen/

作者：[rakyll](https://rakyll.org/about/)
译者：[krystollia](https://github.com/krystollia)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
