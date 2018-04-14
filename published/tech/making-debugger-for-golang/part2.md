已发布：https://studygolang.com/articles/12794

# 实现一个 Golang 调试器（第二部分）

在[第一部分](https://studygolang.com/articles/12553)里，我们首先介绍了开发环境并且实现了一个简单的调试器（tracer），它可以使子进程（tracee）在最开始处停止运行，然后继续执行，并显示它的标准输出。现在是扩展这个程序的时候了。

通常，调试器允许单步执行被调试的代码，这个可以通过 [ptrace](http://man7.org/linux/man-pages/man2/ptrace.2.html) 的 PTRACE_SINGLESTEP 命令实现，它告诉 tracee　执行完一条指令后停止运行。

```go
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	flag.Parse()
	input := flag.Arg(0)
	cmd := exec.Command(input)
	cmd.Args = []string{input}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Ptrace: true}
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	log.Printf("State: %v\n", err)
	wpid := cmd.Process.Pid
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		log.Panic(err)
	}
	err = syscall.PtraceSetOptions(cmd.Process.Pid, syscall.PTRACE_O_TRACECLONE)
	if err != nil {
		log.Fatal(err)
	}
	err = syscall.PtraceSingleStep(wpid)
	if err != nil {
		log.Fatal(err)
	}
	steps := 1
	
	for {
		var ws syscall.WaitStatus
		wpid, err = syscall.Wait4(-1*pgid, &ws, syscall.WALL, nil)
		if wpid == -1 {
			log.Fatal(err)
		}
		if wpid == cmd.Process.Pid && ws.Exited() {
			break
		}
		if !ws.Exited() {
			err := syscall.PtraceSingleStep(wpid)
			if err != nil {
				log.Fatal(err)
			}
			steps += 1
		}
	}
	log.Printf("Steps: %d\n", steps)
}
```
	
构建并运行这个段代码，输出应该像下面这样（每次调用显示的步数可能不一样）

```
> go install -gcflags="-N -l" github.com/mlowicki/hello
> go install github.com/mlowicki/debugger
> debugger /go/bin/hello
2017/06/09 19:54:42 State: stop signal: trace/breakpoint trap
hello world
2017/06/09 19:54:49 Steps: 297583
```
	
程序的前半部分和上一篇文章里的一样，新加的地方是对 [ syscall.PtraceSingleStep](https://golang.org/pkg/syscall/#PtraceSingleStep) 的调用，它使被调试的程序（在这里是 hello )执行完一条指令后停止。

PTRACE_O_TRACECLONE 选项也被设定了

> PTRACE_O_TRACECLONE (since Linux 2.5.46)
> Stop the tracee at the next clone(2) and automatically start tracing the newly cloned process...
	
(http://man7.org/linux/man-pages/man2/ptrace.2.html)

由于我们的调试器知道新线程什么时间开始并且可以跳过它，所以最后显示的步数是通过所有进程执行的指令总数

被执行的指令数量可能相当多，但是里面包含了 Go 运行时中其它一些初始化代码（有 C 语言开始经验的人应该了解 [libc](https://www.gnu.org/software/libc/) 的初始化过程）。我们可以写一个非常简单的程序来验证我们的调试器工作是正常的。让我们创建一个汇编文件 src/github.com/mlowicki/hello/hello.asm:

```asm
section .data
	msg db "hello, world!", 0xA
	len equ $ — msg
section .text
	global _start
_start:
	mov rax, 1 ; write syscall (https://linux.die.net/man/2/write)
	mov rdi, 1 ; stdout
	mov rsi, msg
	mov rdx, len
	; Passing parameters to `syscall` instruction described in
	; https://en.wikibooks.org/wiki/X86_Assembly/Interfacing_with_Linux#syscall
	syscall
	mov rax, 60 ; exit syscall (https://linux.die.net/man/2/exit)
	mov rdi, 0 ; exit code
	syscall
```	
	
在容器中构建我们的 "hello world" 程序，看一下执行了多少条指令

```shell
> pwd
/go
> apt-get install nasm
> nasm -f elf64 -o hello.o src/github.com/mlowicki/hello/hello.asm && ld -o hello hello.o
> ./hello
hello, world!
> debugger ./hello
2017/06/17 17:58:43 State: stop signal: trace/breakpoint trap
hello, world!
2017/06/17 17:58:43 Steps: 8
```
	
输出结果很好，正好等于 hello.asm 中指令的数量

到目前为止，我们已经知道怎样让程序在一开始停止，如何一步一步的执行代码并查看 进程/线程 的状态，现在是在需要的地方设置断点，监视像变量值这样的进程状态的时候了。

让我们从一个简单的例子开始，hello.go 中有一个 main 函数

```go
package main

import "fmt"

func main() {
	fmt.Println("hello world")
}
```
	
怎样在这个函数的一开始设置断点呢？我们的程序经过编译链接后，最终生成的是一系列机器指令。怎样在只包含了一些二进制代码（只有 CPU 能理解的格式）的源文件里表示我们要设置一个断点呢？

## lineTable

Golang 内置有一些功能，可以访问编译生成的二进制文件中的调试信息。 维护 指令计数器 ( [PC](https://en.wikipedia.org/wiki/Program_counter) ) 和程序代码行的映射关系的结构叫做[行表](https://golang.org/pkg/debug/gosym/#LineTable)，让我们通过一个例子来看一下

```go
package main

import (
	"debug/elf"
	"debug/gosym"
	"flag"
	"log"
)

func main() {
	flag.Parse()
	path := flag.Arg(0)
	exe, err := elf.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	var pclndat []byte
	if sec := exe.Section(".gopclntab"); sec != nil {
		pclndat, err = sec.Data()
		if err != nil {
			log.Fatalf("Cannot read .gopclntab section: %v", err)
		}
	}
	sec := exe.Section(".gosymtab")
	symTabRaw, err := sec.Data()
	pcln := gosym.NewLineTable(pclndat, exe.Section(".text").Addr)
	symTab, err := gosym.NewTable(symTabRaw, pcln)
	if err != nil {
		log.Fatal("Cannot create symbol table: %v", err)
	}
	sym := symTab.LookupFunc("main.main")
	filename, lineno, _ := symTab.PCToLine(sym.Entry)
	log.Printf("filename: %v\n", filename)
	log.Printf("lineno: %v\n", lineno)
}
```
	
如果传递给上面程序的文件中包含以下代码

```go
package main

import "fmt"

func main() {
	fmt.Println("hello world")
}
```
	
那么输出应该是这样的

```shell
> go install github.com/mlowicki/linetable
> go install — gcflags=”-N -l” github.com/mlowicki/hello
> linetable /go/bin/hello
2017/06/30 18:47:38 filename: /go/src/github.com/mlowicki/hello/hello.go
2017/06/30 18:47:38 lineno: 5
```
	
ELF 是 [Executable and Linkable Format](https://en.wikipedia.org/wiki/Executable_and_Linkable_Format) 的缩写，是一种可执行文件的格式

```shell
> apt-get install file
> file /go/bin/hello
/go/bin/hello: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, not stripped
```

ELF 中包含许多段，我们用到了其中三个：.text、.gopclntab 和 .gosymtab。第一个包含了机器指令，第二个实现了指令计数器到源码行的映射，最后一个是一个[符号表](https://en.wikipedia.org/wiki/Symbol_table)

在接下来的文章中，我们会学习怎样用「行表」在一个需要的地方设置断点以及怎样监视程序状态

---

via: https://medium.com/golangspec/making-debugger-in-golang-part-ii-d2b8eb2f19e0

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[jettyhan](https://github.com/jettyhan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出