# Go 最小硬件编程（第二部分）

首发于：https://studygolang.com/articles/23085

[![STM32F030F4P6](https://raw.githubusercontent.com/studygolang/gctt-images/master/Very-Small-Hardware-Part2/1.jpeg)](https://ziutek.github.io/2018/04/14/go_on_very_small_hardware2.html)

在本文 [第一部分](https://studygolang.com/articles/22742) 的结尾，我说过要写一下关于 *接口* 的东西。我不想在这里写一篇完整或是简短的关于接口的讲稿。相反，我将会举一个简单的例子，用以说明如何定义和使用接口，同时知道如何利用通用的 *io.Writer* 接口。同时有少量关于 *reflection* 和 *semihosting* 的叙述。

接口是 Go 语言的关键部分。如果你想要更多的学习它们，我建议去阅读 [Effective Go](https://golang.org/doc/effective_go.html#interfaces) 和 [Russ Cox article](https://research.swtch.com/interfaces)。

## 并发闪烁 - 重温

当你阅读之前那个例子的代码时，可能会注意到开启或是关闭 LED 的方式有悖于常理。*Set* 方法用于关闭 LED，同时 *Clear* 方法用于打开 LED，这是由于使用了开漏配置去驱动 LED。我们可以做什么以让代码不那么让人迷惑呢？让我们使用 *On* 和 *Off* 方法来定义 LED 类型吧。

```go
type LED struct {
	pin gpio.Pin
}

func (led LED) On() {
	led.pin.Clear()
}

func (led LED) Off() {
	led.pin.Set()
}
```

现在我们可以简单地调用 `led.On()` 和 `led.Off()` 方法了，这样就不会再引起任何疑惑。

在前面所有的的例子中，我努力去使用相同的开漏配置，以让代码不会变得复杂。但是在上一个例子中，在 GND 和 PA3 引脚间连接第三个 LED，同时配置 PA3 为推拉模式，似乎更简单。下一个示例将使用这种方式连接的 LED。

但是我们的新 *LED* 类型不支持推拉模式配置。实际上，我们应该叫它 *OpenDrainLED* 同时定义另外一个 *PushPullLED* 类型。

```go
type PushPullLED struct {
	pin gpio.Pin
}

func (led PushPullLED) On() {
	led.pin.Set()
}

func (led PushPullLED) Off() {
	led.pin.Clear()
}
```

注意，这两种类型都有同样的方法进行同样的工作。如果在 LED 上运行的代码可以使用这两种类型，而不用关注它目前使用哪种类型，这将会是很棒的。*接口* 类型就能帮上忙了。

```go
package main

import (
	"delay"

	"stm32/hal/gpio"
	"stm32/hal/system"
	"stm32/hal/system/timer/systick"
)

type LED interface {
	On()
	Off()
}

type PushPullLED struct{ pin gpio.Pin }

func (led PushPullLED) On()  {
	led.pin.Set()
}

func (led PushPullLED) Off() {
	led.pin.Clear()
}

func MakePushPullLED(pin gpio.Pin) PushPullLED {
	pin.Setup(&gpio.Config{Mode: gpio.Out, Driver: gpio.PushPull})
	return PushPullLED{pin}
}

type OpenDrainLED struct{ pin gpio.Pin }

func (led OpenDrainLED) On()  {
	led.pin.Clear()
}

func (led OpenDrainLED) Off() {
	led.pin.Set()
}

func MakeOpenDrainLED(pin gpio.Pin) OpenDrainLED {
	pin.Setup(&gpio.Config{Mode: gpio.Out, Driver: gpio.OpenDrain})
	return OpenDrainLED{pin}
}

var led1, led2 LED

func init() {
	system.SetupPLL(8, 1, 48/8)
	systick.Setup(2e6)

	gpio.A.EnableClock(false)
	led1 = MakeOpenDrainLED(gpio.A.Pin(4))
	led2 = MakePushPullLED(gpio.A.Pin(3))
}

func blinky(led LED, period int) {
	for {
		led.On()
		delay.Millisec(100)
		led.Off()
		delay.Millisec(period - 100)
	}
}

func main() {
	go blinky(led1, 500)
	blinky(led2, 1000)
}
```

我们定义了 *LED* 接口，其有两个方法：*On* 和 *Off*。*PushPullLED* 和 *OpenDrainLED* 类型代表了驱动 LED 的两种方法。我们也定义了两个 *Make\*LED* 函数，其作为构造函数。两个类型都实现了 *LED* 接口，所以这些类型的值能够被赋予 *LED* 类型的变量：

```go
led1 = MakeOpenDrainLED(gpio.A.Pin(4))
led2 = MakePushPullLED(gpio.A.Pin(3))
```

在这种情况下，可赋值性在编译时期就被检查过了。赋值后，*led1* 变量包含 `OpenDrainLED{gpio.A.Pin(4)}` 和 一个指向 *OpenDrainLED* 类型方法的指针。`led1.On()` 的调用粗略对应于以下的 C 代码：

```c
led1.methods->On(led1.value)
```

正如你所看到的，如果只考虑函数调用开销，这是十分实惠的抽象方式。

但是任何对于一个接口的赋值都会造成包含大量的被赋值类型的信息。在复杂情况下可以有大量信息，其中包含许多其他类型：

```bash
$ egc
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  10356     196     212   10764    2a0c cortexm0.elf
```

如果我们不使用 [反射](https://blog.golang.org/laws-of-reflection)，我们可以通过避免引入类型名称和结构体字段名称，来节省一些字节：

```bash
$ egc -nf -nt
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  10312     196     212   10720    29e0 cortexm0.elf
```

生成的二进制码仍然包含一些必要的类型信息和所有被导出方法（和名字）的全部信息。在你把一个被存储于接口变量中的值赋值为任何其它变量的时候，这个信息在运行时主要被用于检查赋值能力。

我们也可以通过重新编译它们移除来自于被导入包的类型和域名称：

```bash
$ cd $HOME/emgo
$ ./clean.sh
$ cd $HOME/firstemgo
$ egc -nf -nt
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  10272     196     212   10680    29b8 cortexm0.elf
```

让我们来加载这个程序，看看它是否按照预期工作。这次我们将使用 [st-flash](https://github.com/texane/stlink) 指令。

```bash
$ arm-none-eabi-objcopy -O binary cortexm0.elf cortexm0.bin
$ st-flash write cortexm0.bin 0x8000000
st-flash 1.4.0-33-gd76e3c7
2018-04-10T22:04:34 INFO usb.c: -- exit_dfu_mode
2018-04-10T22:04:34 INFO common.c: Loading device parameters....
2018-04-10T22:04:34 INFO common.c: Device connected is: F0 small device, id 0x10006444
2018-04-10T22:04:34 INFO common.c: SRAM size: 0x1000 bytes (4 KiB), Flash: 0x4000 bytes (16 KiB) in pages of 1024 bytes
2018-04-10T22:04:34 INFO common.c: Attempting to write 10468 (0x28e4) bytes to stm32 address: 134217728 (0x8000000)
Flash page at addr: 0x08002800 erased
2018-04-10T22:04:34 INFO common.c: Finished erasing 11 pages of 1024 (0x400) bytes
2018-04-10T22:04:34 INFO common.c: Starting Flash write for VL/F0/F3/F1_XL core id
2018-04-10T22:04:34 INFO flash_loader.c: Successfully loaded flash loader in sram
 11/11 pages written
2018-04-10T22:04:35 INFO common.c: Starting verification of write complete
2018-04-10T22:04:35 INFO common.c: Flash written and verified! jolly good!
```

我没有连接 NRST 信号到编程器，因此 *复位* 指令不能使用，复位按钮必须被按下以启动程序。

![Interfaces](https://raw.githubusercontent.com/studygolang/gctt-images/master/Very-Small-Hardware-Part2/2.png)

似乎 *st-flash* 程序在这个开发板上工作得有那么一点不可靠（通常需要复位 ST-LINK 适配器）。除此之外，当前版本的程序没有通过 SWD 发出复位信号（只使用了 NRST 信号）。软件复位是不可靠，但是通常情况下能工作，缺乏它会带来一些不方便。对于这个开发板，编程器与 *OpenOCD* 搭配会工作得更好。

## [UART](https://ziutek.github.io/2018/04/14/go_on_very_small_hardware2.html#uart)

UART（通用异步收发传输器）仍然是现代微控制器中最重要的外设之一。以下特性的独特组合，构成了它的优势：

- 相对高的速度，
- 只有两条信号线（半双工通信下甚至只有一条），
- 角色的对称性，
- 新数据（起始 bit 位）的同步带内信号，
- 在被传输信息内准确定时。

这些优点造就了最初旨在传输 7-9 个 bit 位组成的异步信息的 UART，也能够被用于高效地实现诸多其他物理协议，比如被用于 [WS28xx LEDs](http://www.world-semi.com/solution/list-4-1.html) 或是 [1-wire](https://en.wikipedia.org/wiki/1-Wire) 设备。

然而，我们将使用 UART 的常规功能：从我们的程序打印文本信息。

```go
package main

import (
	"rtos"

	"stm32/hal/dma"
	"stm32/hal/gpio"
	"stm32/hal/irq"
	"stm32/hal/system"
	"stm32/hal/system/timer/systick"
	"stm32/hal/usart"
)

var tts *usart.Driver

func init() {
	system.SetupPLL(8, 1, 48/8)
	systick.Setup(2e6)

	gpio.A.EnableClock(true)
	tx := gpio.A.Pin(9)

	tx.Setup(&gpio.Config{Mode: gpio.Alt})
	tx.SetAltFunc(gpio.USART1_AF1)
	d := dma.DMA1
	d.EnableClock(true)
	tts = usart.NewDriver(usart.USART1, d.Channel(2, 0), nil, nil)
	tts.Periph().EnableClock(true)
	tts.Periph().SetBaudRate(115200)
	tts.Periph().Enable()
	tts.EnableTx()

	rtos.IRQ(irq.USART1).Enable()
	rtos.IRQ(irq.DMA1_Channel2_3).Enable()
}

func main() {
	tts.WriteString("Hello, World!\r\n")
}

func ttsISR() {
	tts.ISR()
}

func ttsDMAISR() {
	tts.TxDMAISR()
}

//c:__attribute__((section(".ISRs")))
var ISRs = [...]func(){
	irq.USART1:          ttsISR,
	irq.DMA1_Channel2_3: ttsDMAISR,
}
```

你会发现这个代码有一点复杂，但是到目前为止在 STM32 HAL 中还没有更简单的 UART 驱动（在某些情况下，简单的轮询驱动程序可能很有用）。*usart.Driver* 是使用 DMA 和中断来卸载 CPU 的高效驱动程序。

STM32 USART 设备提供了传统的 UART 和它的同步版本。为了利用它作为输出，我们必须把它的 Tx 信号连接到对应的 GPIO 引脚：

```go
tx.Setup(&gpio.Config{Mode: gpio.Alt})
tx.SetAltFunc(gpio.USART1_AF1)
```

*uasrt.Driver* 被配置为 Tx 模式（rxdma 和 rxbuf 被置空）：

```go
tts = usart.NewDriver(usart.USART1, d.Channel(2, 0), nil, nil)
```

我们利用它的 *WriteString* 方法去打印著名的句子。让我们清除所有的东西，编译这个程序：

```bash
$ cd $HOME/emgo
$ ./clean.sh
$ cd $HOME/firstemgo
$ egc
$ arm-none-eabi-size cortexm0.elf
  text	   data	    bss	    dec	    hex	filename
  12728	    236	    176	  13140	   3354	cortexm0.elf
```

为了能够看到一些文本，需要你的电脑有一个 UART 设备。

**不要使用 RS232 接口或者 USB 转 RS232 转换器**

STM32 都是使用 3.3V 逻辑电平，但是 RS232 能够产生从 -15V 到 +15V 的电压，这可能会烧掉你的 MCU。你需要 3.3V 逻辑电平的 USB 转 UART 转换器。流行的转换器是基于 FT232 或是 CP2102 芯片。

![UART](https://raw.githubusercontent.com/studygolang/gctt-images/master/Very-Small-Hardware-Part2/3.jpeg)

你也需要一些终端模拟器程序（我比较喜欢 [picocom](https://github.com/npat-efault/picocom)）。烧录新的镜像，运行终端模拟器，按几次复位按钮：

```bash
$ openocd -d0 -f interface/stlink.cfg -f target/stm32f0x.cfg -c 'init; program cortexm0.elf; reset run; exit'
Open On-Chip Debugger 0.10.0+dev-00319-g8f1f912a (2018-03-07-19:20)
Licensed under GNU GPL v2
For bug reports, read
        http://openocd.org/doc/doxygen/bugs.html
debug_level: 0
adapter speed: 1000 kHz
adapter_nsrst_delay: 100
none separate
adapter speed: 950 kHz
target halted due to debug-request, current mode: Thread
xPSR: 0xc1000000 pc: 0x080016f4 msp: 0x20000a20
adapter speed: 4000 kHz
** Programming Started **
auto erase enabled
target halted due to breakpoint, current mode: Thread
xPSR: 0x61000000 pc: 0x2000003a msp: 0x20000a20
wrote 13312 bytes from file cortexm0.elf in 1.020185s (12.743 KiB/s)
** Programming Finished **
adapter speed: 950 kHz
$
$ picocom -b 115200 /dev/ttyUSB0
picocom v3.1

port is        : /dev/ttyUSB0
flowcontrol    : none
baudrate is    : 115200
parity is      : none
databits are   : 8
stopbits are   : 1
escape is      : C-a
local echo is  : no
noinit is      : no
noreset is     : no
hangup is      : no
nolock is      : no
send_cmd is    : sz -vv
receive_cmd is : rz -vv -E
imap is        :
omap is        :
emap is        : crcrlf,delbs,
logfile is     : none
initstring     : none
exit_after is  : not set
exit is        : no

Type [C-a] [C-h] to see available commands
Terminal ready
Hello, World!
Hello, World!
Hello, World!
```

每按一次复位按钮，输出一行新的 "Hello, World!"。这些运行结果都符合预期。

想看这个 MCU 的双向 UART 代码，请看这个 [例子](https://github.com/ziutek/emgo/blob/master/egpath/src/stm32/examples/f030-demo-board/usart/main.go)。

## io.Writer

*io.Writer* 接口可能是 Go 中第二最常使用的接口了，紧随 *error* 接口之后。它的定义如下：

```go
type Writer interface {
	Write(p []byte) (n int, err error)
}
```

*usart.Driver* 实现了 *io.Writer*：

```go
tts.WriteString("Hello, World!\r\n")
```

因此我们可以用下面这句替换上面那句：

```go
io.WriteString(tts, "Hello, World!\r\n")
```

此外，你需要在 *import* 那里添加 *io* 包引入。

*io.WriteString* 函数声明如下所示：

```go
func WriteString(w Writer, s string) (n int, err error)
```

正如你所看到的，*io.WriteString* 允许使用实现了 *io.Writer* 接口的任何类型来写字符串。在内部检查底层类型是否具有 *WriteString* 方法并使用它而不是 *Write*。

编译修改后的程序：

```bash
$ egc
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  15456     320     248   16024    3e98 cortexm0.elf
```

正如你所看到的，*io.WriteString* 造成了明显的二进制代码大小的增加：15776 - 12964 = 2812 字节。没有太多剩余的 Flash 空间了。是什么造成了如此剧烈的增长呢？

使用下面这个命令进行分析：

```bash
arm-none-eabi-nm --print-size --size-sort --radix=d cortexm0.elf
```

我们可以打印两种情况下所有按其大小排序的符号。通过过滤和分析获得的数据（awk，diff），我们可以找到大约 80 个新符号。最大的 10 个是：

```go
> 00000062 T stm32$hal$usart$Driver$DisableRx
> 00000072 T stm32$hal$usart$Driver$RxDMAISR
> 00000076 T internal$Type$Implements
> 00000080 T stm32$hal$usart$Driver$EnableRx
> 00000084 t errors$New
> 00000096 R $8$stm32$hal$usart$Driver$$
> 00000100 T stm32$hal$usart$Error$Error
> 00000360 T io$WriteString
> 00000660 T stm32$hal$usart$Driver$Read
```

因此，尽管我们没有使用 *usart.Driver.Read* 方法，但还是被编译进来了，与 *DisableRx*、*RxDMAISR*、*EnableRx* 和其他在上面没有提及的一样。不幸的是，如果你为接口分配了一些内容，则需要其完整的方法集（包含所有依赖项）。对于要使用这些方法中大多数的大型程序来说，这不是问题，但是对于我们这个简单程序来说，就是一个巨大的负担。

我们已经逼近 MCU 的限制了，但是我们还是要尝试打印一些数字（你需要在 *import* 中使用 *strcon* 包替换 *io* 包）：

```go
func main() {
	a := 12
	b := -123

	tts.WriteString("a = ")
	strconv.WriteInt(tts, a, 10, 0, 0)
	tts.WriteString("\r\n")
	tts.WriteString("b = ")
	strconv.WriteInt(tts, b, 10, 0, 0)
	tts.WriteString("\r\n")

	tts.WriteString("hex(a) = ")
	strconv.WriteInt(tts, a, 16, 0, 0)
	tts.WriteString("\r\n")
	tts.WriteString("hex(b) = ")
	strconv.WriteInt(tts, b, 16, 0, 0)
	tts.WriteString("\r\n")
}
```

在 *io.WriteString* 函数中，*strconv.WriteInt* 的第一参数是 *io.Writer* 类型的。

```bash
$ egc
/usr/local/arm/bin/arm-none-eabi-ld: /home/michal/firstemgo/cortexm0.elf section `.rodata' will not fit in region `Flash'
/usr/local/arm/bin/arm-none-eabi-ld: region `Flash' overflowed by 692 bytes
exit status 1
```

这次我们已经用完了存储空间。让我们来尝试简化类型的信息：

```bash
$ cd $HOME/emgo
$ ./clean.sh
$ cd $HOME/firstemgo
$ egc -nf -nt
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  15876     316     320   16512    4080 cortexm0.elf
```

和之前很接近，但更为合适。让我们加载并运行此代码：

```go
a = 12
b = -123
hex(a) = c
hex(b) = -7b
```

在 Emgo 中的 *strconv* 包和它在 Go 中原始包有很大的不同。

Emgo 中的 *strconv* 包与 Go 中的原型完全不同。它旨在直接用于编写格式化数字，并且在许多情况下可以替换重型 *fmt* 包。这就是为什么函数名称以 *Write* 而不是 *Format* 开头并有另外两个参数。以下是它们的使用示例：

```go
func main() {
	b := -123
	strconv.WriteInt(tts, b, 10, 0, 0)
	tts.WriteString("\r\n")
	strconv.WriteInt(tts, b, 10, 6, ' ')
	tts.WriteString("\r\n")
	strconv.WriteInt(tts, b, 10, 6, '0')
	tts.WriteString("\r\n")
	strconv.WriteInt(tts, b, 10, 6, '.')
	tts.WriteString("\r\n")
	strconv.WriteInt(tts, b, 10, -6, ' ')
	tts.WriteString("\r\n")
	strconv.WriteInt(tts, b, 10, -6, '0')
	tts.WriteString("\r\n")
	strconv.WriteInt(tts, b, 10, -6, '.')
	tts.WriteString("\r\n")
}
```

输出：

```go
-123
  -123
-00123
..-123
-123
-123
-123..
```

## Unix 流和莫尔斯码

由于大多数编写内容的函数都使用 *io.Writer* 而不是具体类型（例如 C 中的 *FILE*），因此我们得到了类似于 *Unix* 流的功能。在 Unix 中，我们可以轻松地组合简单命令来执行更大的任务。例如，我们可以这样写文本到文件：

```bash
echo "Hello, World!" > file.txt
```

`>` 运算符将前一个命令的输出流写入文件。还有 `|` 运算符连接相邻命令的输出和输入流。

通过流，我们可以轻松转换或过滤任何命令的输出。例如，要将所有字母转换为大写，我们可以通过 *tr* 命令过滤 echo 的输出：

```bash
echo "Hello, World!" | tr a-z A-Z > file.txt
```

为了展现 *io.Writer* 和 Unix 流之间的相似之处，让我们来写我们的：

```go
io.WriteString(tts, "Hello, World!\r\n")
```

在以下类 unix 形式中：

```bash
io.WriteString "Hello, World!" | usart.Driver usart.USART1
```

下面的例子将会展现如何去做这个事：

```bash
io.WriteString "Hello, World!" | MorseWriter | usart.Driver usart.USART1
```

让我们创建一个简单的编码器，使用莫尔斯码对写入其中的文本进行编码：

```go
type MorseWriter struct {
	W io.Writer
}

func (w *MorseWriter) Write(s []byte) (int, error) {
	var buf [8]byte
	for n, c := range s {
		switch {
		case c == '\n':
			c = ' ' // Replace new lines with spaces.
		case 'a' <= c && c <= 'z':
			c -= 'a' - 'A' // Convert to upper case.
		}
		if c < ' ' || 'Z' < c {
			continue // c is outside ASCII [' ', 'Z']
		}
		var symbol morseSymbol
		if c == ' ' {
			symbol.length = 1
			buf[0] = ' '
		} else {
			symbol = morseSymbols[c-'!']
			for i := uint(0); i < uint(symbol.length); i++ {
				if (symbol.code>>i)&1 != 0 {
					buf[i] = '-'
				} else {
					buf[i] = '.'
				}
			}
		}
		buf[symbol.length] = ' '
		if _, err := w.W.Write(buf[:symbol.length+1]); err != nil {
			return n, err
		}
	}
	return len(s), nil
}

type morseSymbol struct {
	code, length byte
}

//emgo:const
var morseSymbols = [...]morseSymbol{
	{1<<0 | 1<<1 | 1<<2, 4}, // ! ---.
	{1<<1 | 1<<4, 6},        // " .-..-.
	{},                      // #
	{1<<3 | 1<<6, 7},        // $ ...-..-

	// Some code omitted...

	{1<<0 | 1<<3, 4},        // X -..-
	{1<<0 | 1<<2 | 1<<3, 4}, // Y -.--
	{1<<0 | 1<<1, 4},        // Z --..
}
```

你可以在 [这里](https://github.com/ziutek/emgo/blob/master/egpath/src/stm32/examples/f030-demo-board/morseuart/main.go) 找到所有的 *morseSymbols* 数组。`//emgo:const` 指令确保 *morseSymbols* 数组不会被复制到 RAM 中。

现在我们可以使用两种方式打印我们的句子了：

```go
func main() {
	s := "Hello, World!\r\n"
	mw := &MorseWriter{tts}

	io.WriteString(tts, s)
	io.WriteString(mw, s)
}
```

我们使用 *MorseWriter* `&MorseWriter{tts}` 指针而不是 os 中简单的 `MorseWriter{tts}` 值，因为 *MorseWriter* 太大而不适合一个接口变量。

Emgo，不像是 Go，没有为存储于接口变量中的值动态开辟内存。接口类型的大小有限，等于三个指针（适合切片）或两个 *float64*（以适应 *complex128*）的大小，这将会更大。它可以直接存储所有基本类型和小结构体或数组的值，但是对于更大的值，必须使用指针。

让我们编译这段代码并看看它的输出：

```bash
$ egc
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  15152     324     248   15724    3d6c cortexm0.elf
Hello, World!
.... . .-.. .-.. --- --..--   .-- --- .-. .-.. -.. ---.
```

### 终极闪烁

*Blinky* 与 *Hello, World!* 的程序在硬件配置上等同。一旦我们有莫尔斯编码器，我们可以很容易地将它们结合起来以获得 *终极闪烁* 程序：

```go
package main

import (
	"delay"
	"io"

	"stm32/hal/gpio"
	"stm32/hal/system"
	"stm32/hal/system/timer/systick"
)

var led gpio.Pin

func init() {
	system.SetupPLL(8, 1, 48/8)
	systick.Setup(2e6)

	gpio.A.EnableClock(false)
	led = gpio.A.Pin(4)

	cfg := gpio.Config{Mode: gpio.Out, Driver: gpio.OpenDrain, Speed: gpio.Low}
	led.Setup(&cfg)
}

type Telegraph struct {
	Pin   gpio.Pin
	Dotms int // Dot length [ms]
}

func (t Telegraph) Write(s []byte) (int, error) {
	for _, c := range s {
		switch c {
		case '.':
			t.Pin.Clear()
			delay.Millisec(t.Dotms)
			t.Pin.Set()
			delay.Millisec(t.Dotms)
		case '-':
			t.Pin.Clear()
			delay.Millisec(3 * t.Dotms)
			t.Pin.Set()
			delay.Millisec(t.Dotms)
		case ' ':
			delay.Millisec(3 * t.Dotms)
		}
	}
	return len(s), nil
}

func main() {
	telegraph := &MorseWriter{Telegraph{led, 100}}
	for {
		io.WriteString(telegraph, "Hello, World! ")
	}
}

// Some code omitted...
```

在以上的例子中，我省略了 *MorseWriter* 类型的定义，因为之前它就出现过。完整代码可以在 [这里](https://github.com/ziutek/emgo/blob/master/egpath/src/stm32/examples/f030-demo-board/morseled/main.go) 获取。让我们编译它并运行：

```bash
$ egc
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  11772     244     244   12260    2fe4 cortexm0.elf
```

![Ultimate Blinky](https://raw.githubusercontent.com/studygolang/gctt-images/master/Very-Small-Hardware-Part2/4.png)

## 反射

当然，Emgo 支持 [反射](https://blog.golang.org/laws-of-reflection)。*reflect* 包尚未完成，但完成部分就足以实现 *fmt.Print* 系列功能。让我们看看我们的小型 MCU 能做些什么。

为减少内存使用，我们将使用 [semihosting](http://infocenter.arm.com/help/topic/com.arm.doc.dui0471g/Bgbjjgij.html) 作为标准输出。为方便起见，我们还编写了简单的 *println* 函数，它在某种程度上模仿了 *fmt.Println*。

```go
package main

import (
	"debug/semihosting"
	"reflect"
	"strconv"

	"stm32/hal/system"
	"stm32/hal/system/timer/systick"
)

var stdout semihosting.File

func init() {
	system.SetupPLL(8, 1, 48/8)
	systick.Setup(2e6)

	var err error
	stdout, err = semihosting.OpenFile(":tt", semihosting.W)
	for err != nil {
	}
}

type stringer interface {
	String() string
}

func println(args ...interface{}) {
	for i, a := range args {
		if i > 0 {
			stdout.WriteString(" ")
		}
		switch v := a.(type) {
		case string:
			stdout.WriteString(v)
		case int:
			strconv.WriteInt(stdout, v, 10, 0, 0)
		case bool:
			strconv.WriteBool(stdout, v, 't', 0, 0)
		case stringer:
			stdout.WriteString(v.String())
		default:
			stdout.WriteString("%unknown")
		}
	}
	stdout.WriteString("\r\n")
}

type S struct {
	A int
	B bool
}

func main() {
	p := &S{-123, true}

	v := reflect.ValueOf(p)

	println("kind(p) =", v.Kind())
	println("kind(*p) =", v.Elem().Kind())
	println("type(*p) =", v.Elem().Type())

	v = v.Elem()

	println("*p = {")
	for i := 0; i < v.NumField(); i++ {
		ft := v.Type().Field(i)
		fv := v.Field(i)
		println("  ", ft.Name(), ":", fv.Interface())
	}
	println("}")
}
```

*semihosting.OpenFile* 函数允许在主机端打开或创建文件。特殊路径 *:tt* 对应主机的标准输出。

*println* 函数接受任意数量的参数，每个参数都是任意类型：

```go
func println(args ...interface{})
```

因为任何类型都实现了 *interface{}* 这个空接口，所有这成为了可能。*println* 使用 [switch 结构](https://golang.org/doc/effective_go.html#type_switch) 去打印 string、interger 和 boolean 类型：

```go
switch v := a.(type) {
case string:
	stdout.WriteString(v)
case int:
	strconv.WriteInt(stdout, v, 10, 0, 0)
case bool:
	strconv.WriteBool(stdout, v, 't', 0, 0)
case stringer:
	stdout.WriteString(v.String())
default:
	stdout.WriteString("%unknown")
}
```

此外，它支持任何实现 *stringer* 接口的类型，即任何具有 *String()* 方法的类型。在任何 *case* 子句中，*v* 变量具有正确的类型，与 *case* 关键字后面列出的相同。

`reflect.ValueOf(p)` 允许以编程的方式分析其类型和内容的形式返回 *p*。我们甚至可以使用 `v.Elem()` 间接引用指针，并使用其名称打印所有结构体字段。

让我们尝试编译这段代码。现在让我们看看如果没有类型和字段名称编译将会出现什么：

```bash
$ egc -nt -nf
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  16028     216     312   16556    40ac cortexm0.elf
```

Flash 只剩下 140 个空闲字节。让我们使用启用了 semihosting 的 OpenOCD 加载它：

```bash
$ openocd -d0 -f interface/stlink.cfg -f target/stm32f0x.cfg -c 'init; program cortexm0.elf; arm semihosting enable; reset run'
Open On-Chip Debugger 0.10.0+dev-00319-g8f1f912a (2018-03-07-19:20)
Licensed under GNU GPL v2
For bug reports, read
        http://openocd.org/doc/doxygen/bugs.html
debug_level: 0
adapter speed: 1000 kHz
adapter_nsrst_delay: 100
none separate
adapter speed: 950 kHz
target halted due to debug-request, current mode: Thread
xPSR: 0xc1000000 pc: 0x08002338 msp: 0x20000a20
adapter speed: 4000 kHz
** Programming Started **
auto erase enabled
target halted due to breakpoint, current mode: Thread
xPSR: 0x61000000 pc: 0x2000003a msp: 0x20000a20
wrote 16384 bytes from file cortexm0.elf in 0.700133s (22.853 KiB/s)
** Programming Finished **
semihosting is enabled
adapter speed: 950 kHz
kind(p) = ptr
kind(*p) = struct
type(*p) =
*p = {
   X. : -123
   X. : true
}
```

如果你已经真正运行了这个代码，你会注意到 semihosting 很慢，尤其是如果你一个字节一个字节的写入（使用缓冲能够改善这种情况）。

如你所见，`*p` 没有类型名称，同时所有的结构体字段有相同的 *X.* 名称。让我们来再次编译这个程序，这次不用 *-nt* *-nf* 选项：

```bash
$ egc
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  16052     216     312   16580    40c4 cortexm0.elf
```

现在已包含类型和字段名称，但仅在 *main.go* 文件主程序包中定义了这些名称。我们程序的输出如下：

```go
kind(p) = ptr
kind(*p) = struct
type(*p) = S
*p = {
   A : -123
   B : true
}
```

反射是任何易于使用的序列化库的重要组成部分，而像 [JSON](https://en.wikipedia.org/wiki/JSON) 这样的序列化算法在物联网时代变得越来越重要。

到此，我完成了本文的第二部分。我认为有可能第三部分更有趣，我们把这个开发板连接到各种有趣的设备。如果这个开发板不能负担它们，我们用更大的东西替换它。

---

via: https://ziutek.github.io/2018/04/14/go_on_very_small_hardware2.html

作者：[Michał Derkacz ](https://ziutek.github.io)
译者：[PotoYang](https://github.com/PotoYang)
校对：[zhoudingding](https://github.com/dingdinzhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出