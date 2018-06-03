已发布：https://studygolang.com/articles/12804

# 编写 Go 语言调试器 - 第三部分

到目前为止我们已经知道如何单步执行用 ptrace 暂停的进程（tracee）以及如何从二进制文件中获取一些调试信息（在[这里](https://studygolang.com/articles/12794)阅读相关内容）。接下来就是设置断点，等待程序运行到断点处，查看进程相关信息的时候了。

让我们从上一篇文章中用到的汇编代码开始

```asm
section .data
	msg db      "hello, world!", 0xA
	len equ     $ - msg
section .text
	global _start
_start:
	mov     rax, 1 ; write syscall (https://linux.die.net/man/2/write)
	mov     rdi, 1 ; stdout
	mov     rsi, msg
	mov     rdx, len
	; Passing parameters to `syscall` instruction described in
	; https://en.wikibooks.org/wiki/X86_Assembly/Interfacing_with_Linux#syscall
	syscall
	mov    rax, 60 ; exit syscall (https://linux.die.net/man/2/exit)
	mov    rdi, 0 ; exit code
	syscall
```

我们在下面这一行代码处设置了断点

```asm
mov     rdi, 1
```
	
所以在执行到这行代码的时候，程序会暂停，这时候检查 RDI 寄存器中存储的值是否为 0，然后单步执行一行代码，检查这个值是否变成了 1。

## 断点

在 x86 系统中有一条 [中断指令](https://en.wikipedia.org/wiki/INT_%28x86_instruction%29)，它可以产生一个软中断。在 Linux 系统中，这个软中断是通过 f.ex 调用 syscall 来实现的。在 x86-64 系统中引入了一个专用的系统指令来实现这个软中断，比 x86 系统中的指令更快，这就我这所以用这个专用指令的原因了。但是我们可以用普通的 INT 中断完成同样的任务。我们用 INT 3 来设置一个断点，它对应的操作码是 0xCC

INT 3 指令会生成大小为一个字节的特殊操作码（CC），通过它可以调用异常处理函数，（这个操作码非常有用，因为它可以用来替换任何一条指令的第一个字节，使之成为一个断点，然后再加入额外的一个字节，而不影响其它的代码），具体信息参见以下文档

[Intel® 64 and IA-32 系统软件使用手册](https://software.intel.com/en-us/articles/intel-sdm)

我们用 0xCC 来替换特定指令的头一个字节，使之成为一个断点，一旦这个断点被出发，我们就可以做以下的事情

1. 查看进程状态
2. 把 0xCC 操作码替换成原来的值
3. 把程序的计数器值减 1
4. 执行一条指令

我们需要处理的第一个问题是：在哪放置 0xCC，我们不知道  move rdi, 1 这条指令在内存中的具体位置。由于这是第二条指令，所以在程序的开始内存地址基础上加上第一条指令 move rax, 1 的长度，就应该是这条指令的内存地址。由于 x86 系统中指令长度不是定长的，所以让确定指令开始地址变得更加困难了。程序第一条指令的位置可以通过让程序在没有执行任何指令的时候停止的办法得到（我们之前已经做过了），第一条指令的长度可以通过 objdump 命令来获取：

```shell
> nasm -f elf64 -o hello.o src/github.com/mlowicki/hello/hello.asm && ld -o /go/bin/hello hello.o
> objdump -d -M intel /go/bin/hello
/go/bin/hello:     file format elf64-x86-64
Disassembly of section .text:
00000000004000b0 <_start>:
	4000b0:       b8 01 00 00 00          mov    eax,0x1
	4000b5:       bf 01 00 00 00          mov    edi,0x1
	4000ba:       48 be d8 00 60 00 00    movabs rsi,0x6000d8
	4000c1:       00 00 00
	4000c4:       ba 0e 00 00 00          mov    edx,0xe
	4000c9:       0f 05                   syscall
	4000cb:       b8 3c 00 00 00          mov    eax,0x3c
	4000d0:       bf 00 00 00 00          mov    edi,0x0
	4000d5:       0f 05                   syscall
```

从上面的输出我们可以发现第一条指令的长度是 5 个字节（4000b5 - 4000b0），所以我们要把 0xCC 放在第一条指令的位置加 5 个字节地方，下面是代码实现

```go
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func step(pid int) {
	err := syscall.PtraceSingleStep(pid)
	if err != nil {
		log.Fatal(err)
	}
}

func cont(pid int) {
	err := syscall.PtraceCont(pid, 0)
	if err != nil {
		log.Fatal(err)
	}
}

func setPC(pid int, pc uint64) {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		log.Fatal(err)
	}
	regs.SetPC(pc)
	err = syscall.PtraceSetRegs(pid, &regs)
	if err != nil {
		log.Fatal(err)
	}
}

func getPC(pid int) uint64 {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		log.Fatal(err)
	}
	return regs.PC()
}

func setBreakpoint(pid int, breakpoint uintptr) []byte {
	original := make([]byte, 1)
	_, err := syscall.PtracePeekData(pid, breakpoint, original)
	if err != nil {
		log.Fatal(err)
	}
	_, err = syscall.PtracePokeData(pid, breakpoint, []byte{0xCC})
	if err != nil {
		log.Fatal(err)
	}
	return original
}

func clearBreakpoint(pid int, breakpoint uintptr, original []byte) {
	_, err := syscall.PtracePokeData(pid, breakpoint, original)
	if err != nil {
		log.Fatal(err)
	}
}

func printState(pid int) {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("RAX=%d, RDI=%d\n", regs.Rax, regs.Rdi)
}

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
	pid := cmd.Process.Pid
	breakpoint := uintptr(getPC(pid) + 5)
	original := setBreakpoint(pid, breakpoint)
	cont(pid)
	var ws syscall.WaitStatus
	_, err = syscall.Wait4(pid, &ws, syscall.WALL, nil)
	clearBreakpoint(pid, breakpoint, original)  
	printState(pid)
	setPC(pid, uint64(breakpoint))
	step(pid)
	_, err = syscall.Wait4(pid, &ws, syscall.WALL, nil)
	printState(pid)
}
```
	
源文件以一些辅助函数开始，setPC 和 getPC 用来维护 [程序计数器](https://en.wikipedia.org/wiki/Program_counter)，寄存器 PC 存放的是下一条要执行的指令。如果程序在没有执行任何指令的时候被暂停，PC 中的值就是程序第一条指令的内存地址。维护断点的函数（setBreakpoint 和 clearBreakpoint）负责在指令中插入或者移除操作码 0xCC，下面是程序的输出：

```shell
> go install github.com/mlowicki/breakpoint
> breakpoint /go/bin/hello
2017/07/16 21:06:33 State: stop signal: trace/breakpoint trap
2017/07/16 21:06:33 RAX=1, RDI=0
2017/07/16 21:06:33 RAX=1, RDI=1
```
	
输出和我们预期的一样，当程序到达断点时，RDI 寄存器没有被设置（值 为 0 ），执行完下一条指令后（第二条指令），它的值变成了下面指令设置的值

```asm
mov     rdi, 1
```
	
现在我们已经完成了文章一开始列出的任务。当然我们的程序还需要一些计算指令长度的函数，不过不用担心，我们会在之后实现这些功能

## REPL

现在是时候实现调试器的基本框架了，这是一个简单的命令行程序，程序循环等待用户输入像 "set a breakpoint at " 和 "go single step "这样的命令

```go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func initTracee(path string) int {
	cmd := exec.Command(path)
	cmd.Args = []string{path}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Ptrace: true}
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	// Process should be stopped here because of trace/breakpoint trap
	if err == nil {
		log.Fatal("Program exited")
	}
	return cmd.Process.Pid
}

func main() {
	flag.Parse()
	_ = initTracee(flag.Arg(0))
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		command, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				break
			}
			log.Fatal(err)
		}
		command = command[:len(command)-1] // get rid of ending newline character
		if strings.HasPrefix(command, "register ") {
			fmt.Println("register...")
		} else if strings.HasPrefix(command, "breakpoint ") {
			fmt.Println("breakpoint...")
		} else if command == "help" {
			fmt.Println("help...")
		} else if command == "step" {
			fmt.Println("step...")
		} else if command == "continue" {
			fmt.Println("continue")
		} else {
			fmt.Println("unknown command")
		}
	}
}
```
	
这就是我们调试器的基础框架，它没有提供太多的功能（目前为上），但是它已经具备了一个 [REPL](https://en.wikipedia.org/wiki/Read%E2%80%93eval%E2%80%93print_loop) 环境必要的逻辑。我们或多或少知道了一些如何实现像下一步、继续、查看变量状态等常用调试命令，我们会在不久之后实现这些功能。在 Golang 中不太清晰的一点是断点命令。这个会在接下来的文章中详细解释，为什么在 Golang 中比想象的困难一些以及如何克服这些复杂性。

----------------

via: https://medium.com/golangspec/making-debugger-in-golang-part-iii-5aac8e49f291

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[jettyhan](https://github.com/jettyhan)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
